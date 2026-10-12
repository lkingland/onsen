// Package fork reads and moves the Onsen fork: a colocated jj repository
// whose bookmark carries our commits on top of upstream Quattro.
package fork

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Fork is the fork's checkout.
type Fork struct {
	Run      run.Runner
	Dir      string // the checkout
	Bookmark string // our commits: master
	Remote   string // upstream's remote: upstream
	Upstream string // the revision we sit on: quattro@upstream
}

func (f Fork) jj(args ...string) run.Cmd {
	return run.C("jj", append([]string{"-R", f.Dir}, args...)...)
}

// jjStill is jj without snapshotting the working copy: a snapshot would sign
// a commit with the card.
func (f Fork) jjStill(args ...string) run.Cmd {
	return f.jj(append([]string{"--ignore-working-copy"}, args...)...)
}

func (f Fork) git(args ...string) run.Cmd {
	return run.C("git", append([]string{"-C", f.Dir}, args...)...)
}

// Dirty is git's porcelain status of the checkout, one line per path. git,
// not jj: jj status snapshots the working copy and signs it.
func (f Fork) Dirty(ctx context.Context) ([]string, error) {
	out, err := f.Run.Output(ctx, f.git("status", "--porcelain"))
	if err != nil {
		return nil, err
	}
	return run.Lines(out), nil
}

// OnBookmark reports whether the working copy is the bookmark or a
// descendant of it.
func (f Fork) OnBookmark(ctx context.Context) bool {
	out, err := f.Run.Output(ctx, run.Cmd{Name: "jj", Args: []string{"--ignore-working-copy", "-R", f.Dir,
		"log", "-r", "@ & " + f.Bookmark + "::", "-n", "1", "--no-graph", "-T", "commit_id"}, Quiet: true})
	return err == nil && strings.TrimSpace(out) != ""
}

// Fetch fetches upstream and checks that the upstream revision exists.
func (f Fork) Fetch(ctx context.Context) error {
	if err := f.Run.Run(ctx, f.jj("git", "fetch", "--remote", f.Remote)); err != nil {
		return err
	}
	// Tracking is a no-op once done; a failure here shows up in the check below.
	track := f.jj("bookmark", "track", f.Upstream)
	track.Quiet = true
	_, _ = f.Run.Output(ctx, track)
	if _, err := f.Run.Output(ctx, f.jj("log", "-r", f.Upstream, "-n", "1", "--no-graph", "-T", `""`)); err != nil {
		return fmt.Errorf("%s does not exist after the fetch: %w", f.Upstream, err)
	}
	return nil
}

// Conflicts is each conflicted commit on the bookmark, as
// "<change id> <subject>".
func (f Fork) Conflicts(ctx context.Context) ([]string, error) {
	out, err := f.Run.Output(ctx, f.jj("log", "-r", "conflicts() & ::"+f.Bookmark, "--no-graph",
		"-T", `change_id.short() ++ " " ++ description.first_line() ++ "\n"`))
	if err != nil {
		return nil, err
	}
	return run.Lines(out), nil
}

// Behind is how many upstream commits the bookmark does not have.
func (f Fork) Behind(ctx context.Context) (int, error) {
	out, err := f.Run.Output(ctx, f.jj("log", "-r", "::"+f.Upstream+" ~ ::"+f.Bookmark, "--no-graph",
		"-T", `commit_id ++ "\n"`))
	if err != nil {
		return 0, err
	}
	return len(run.Lines(out)), nil
}

// Rebase moves our commits onto the upstream revision. The card re-signs
// each one. Our commits are pushed to origin, so jj's default immutable set
// (trunk(), the bookmark on origin) covers them; rewriting them is what the
// fork is, and this is the one place it happens.
func (f Fork) Rebase(ctx context.Context) error {
	return f.Run.Run(ctx, f.jj("--ignore-immutable", "rebase", "-b", f.Bookmark, "-d", f.Upstream))
}

// Commit is the bookmark's commit id.
func (f Fork) Commit(ctx context.Context) (string, error) {
	out, err := f.Run.Output(ctx, f.jjStill("log", "-r", f.Bookmark, "-n", "1", "--no-graph", "-T", "commit_id"))
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(out)
	if id == "" {
		return "", fmt.Errorf("%s has no commit", f.Bookmark)
	}
	return id, nil
}

// Tree is git's tree id for a commit.
func (f Fork) Tree(ctx context.Context, commit string) (string, error) {
	out, err := f.Run.Output(ctx, f.git("rev-parse", commit+"^{tree}"))
	return strings.TrimSpace(out), err
}

// Count is the number of commits reachable from commit.
func (f Fork) Count(ctx context.Context, commit string) (int, error) {
	out, err := f.Run.Output(ctx, f.git("rev-list", "--count", commit))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

// Archive writes commit's tree, or only the named paths in it, to a tar
// file.
func (f Fork) Archive(ctx context.Context, commit, tarFile string, paths ...string) error {
	args := append([]string{"archive", "-o", tarFile, commit}, paths...)
	return f.Run.Run(ctx, f.git(args...))
}

// Ident names the content of the given paths at commit: git's listing of
// them, which carries each one's object id. A path that does not exist is
// left out.
func (f Fork) Ident(ctx context.Context, commit string, paths ...string) (string, error) {
	out, err := f.Run.Output(ctx, f.git(append([]string{"ls-tree", commit, "--"}, paths...)...))
	return out, err
}
