package fork

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lkingland/onsen/internal/run"
)

// These run real jj and git: the revsets are where a mistake would hide.

type repos struct {
	t        *testing.T
	up, fork string
	f        Fork
}

func sh(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newRepos is an upstream git repository with one commit on quattro, and a
// colocated jj fork with one commit of ours on master on top of it.
func newRepos(t *testing.T) *repos {
	for _, tool := range []string{"jj", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip("no " + tool)
		}
	}
	root := t.TempDir()
	config := filepath.Join(root, "jj.toml")
	writeFile(t, config, "user.name = \"Test\"\nuser.email = \"test@example.com\"\nsigning.behavior = \"drop\"\n")
	t.Setenv("JJ_CONFIG", config)
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")

	r := &repos{t: t, up: filepath.Join(root, "up"), fork: filepath.Join(root, "fork")}
	sh(t, root, "git", "init", "-q", "-b", "quattro", r.up)
	writeFile(t, filepath.Join(r.up, "README.md"), "upstream\n")
	sh(t, r.up, "git", "add", ".")
	sh(t, r.up, "git", "commit", "-q", "-m", "upstream one")

	sh(t, root, "jj", "git", "init", "--colocate", r.fork)
	sh(t, r.fork, "jj", "git", "remote", "add", "upstream", r.up)
	sh(t, r.fork, "jj", "git", "fetch", "--remote", "upstream")
	sh(t, r.fork, "jj", "bookmark", "track", "quattro@upstream")
	sh(t, r.fork, "jj", "new", "quattro@upstream")
	writeFile(t, filepath.Join(r.fork, "ours.txt"), "ours\n")
	sh(t, r.fork, "jj", "commit", "-m", "ours")
	sh(t, r.fork, "jj", "bookmark", "create", "master", "-r", "@-")

	exe := &run.Exec{Stdin: strings.NewReader(""), Stdout: io.Discard, Stderr: io.Discard}
	r.f = Fork{Run: exe, Dir: r.fork, Bookmark: "master", Remote: "upstream", Upstream: "quattro@upstream"}
	return r
}

// upstream commits one more change to upstream.
func (r *repos) upstream(file, body string) {
	writeFile(r.t, filepath.Join(r.up, file), body)
	sh(r.t, r.up, "git", "add", ".")
	sh(r.t, r.up, "git", "commit", "-q", "-m", "upstream "+file)
}

func TestFetchBehindRebase(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	if !r.f.OnBookmark(ctx) {
		t.Fatal("the working copy rides master but OnBookmark says no")
	}
	if dirty, err := r.f.Dirty(ctx); err != nil || len(dirty) != 0 {
		t.Fatalf("dirty %q, %v", dirty, err)
	}
	before, err := r.f.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	r.upstream("new.txt", "new\n")
	r.upstream("newer.txt", "newer\n")
	if err := r.f.Fetch(ctx); err != nil {
		t.Fatal(err)
	}
	if n, err := r.f.Behind(ctx); err != nil || n != 2 {
		t.Fatalf("behind %d, %v; want 2", n, err)
	}
	if err := r.f.Rebase(ctx); err != nil {
		t.Fatal(err)
	}
	if n, err := r.f.Behind(ctx); err != nil || n != 0 {
		t.Fatalf("behind %d after the rebase, %v", n, err)
	}
	if c, err := r.f.Conflicts(ctx); err != nil || len(c) != 0 {
		t.Fatalf("conflicts %q, %v", c, err)
	}
	after, err := r.f.Commit(ctx)
	if err != nil || after == before {
		t.Fatalf("master did not move: %s, %v", after, err)
	}
	if n, err := r.f.Count(ctx, after); err != nil || n != 4 {
		t.Errorf("count %d, %v; want 4", n, err)
	}
	if tree, err := r.f.Tree(ctx, after); err != nil || len(tree) != 40 {
		t.Errorf("tree %q, %v", tree, err)
	}
	if !r.f.OnBookmark(ctx) {
		t.Error("the working copy no longer rides master after the rebase")
	}
}

// Our commits are pushed to origin, which puts them in jj's default immutable
// set (trunk() is master@origin). The rebase rewrites them all the same.
func TestRebaseRewritesPushedCommits(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	origin := filepath.Join(filepath.Dir(r.fork), "origin.git")
	sh(t, filepath.Dir(r.fork), "git", "init", "-q", "--bare", origin)
	sh(t, r.fork, "jj", "git", "remote", "add", "origin", origin)
	sh(t, r.fork, "jj", "git", "push", "--remote", "origin", "--all")
	if out := sh(t, r.fork, "jj", "log", "-r", "master & immutable()", "--no-graph", "-T", "commit_id"); strings.TrimSpace(out) == "" {
		t.Fatal("master is not immutable after the push, so this proves nothing")
	}
	r.upstream("two.txt", "two\n")
	if err := r.f.Fetch(ctx); err != nil {
		t.Fatal(err)
	}
	if err := r.f.Rebase(ctx); err != nil {
		t.Fatal(err)
	}
	if n, err := r.f.Behind(ctx); err != nil || n != 0 {
		t.Fatalf("behind %d after the rebase, %v", n, err)
	}
}

func TestRebaseConflictIsReported(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	r.upstream("ours.txt", "theirs\n")
	if err := r.f.Fetch(ctx); err != nil {
		t.Fatal(err)
	}
	if err := r.f.Rebase(ctx); err != nil {
		t.Fatal(err)
	}
	c, err := r.f.Conflicts(ctx)
	if err != nil || len(c) != 1 || !strings.HasSuffix(c[0], " ours") {
		t.Fatalf("conflicts %q, %v; want our commit", c, err)
	}
}

func TestFetchUnknownUpstreamFails(t *testing.T) {
	r := newRepos(t)
	r.f.Upstream = "nothing@upstream"
	if err := r.f.Fetch(context.Background()); err == nil {
		t.Fatal("a missing upstream revision passed")
	}
}

func TestDirtyAndOffBookmark(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	writeFile(t, filepath.Join(r.fork, "scratch.txt"), "x\n")
	if dirty, err := r.f.Dirty(ctx); err != nil || len(dirty) != 1 {
		t.Errorf("dirty %q, %v; want the scratch file", dirty, err)
	}
	sh(t, r.fork, "jj", "new", "quattro@upstream")
	if r.f.OnBookmark(ctx) {
		t.Error("OnBookmark on a working copy beside master")
	}
}

func TestArchiveAndIdent(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	commit, err := r.f.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tarFile := filepath.Join(t.TempDir(), "src.tar")
	if err := r.f.Archive(ctx, commit, tarFile, "ours.txt"); err != nil {
		t.Fatal(err)
	}
	if list := sh(t, ".", "tar", "-tf", tarFile); strings.TrimSpace(list) != "ours.txt" {
		t.Errorf("archive holds %q", list)
	}
	a, err := r.f.Ident(ctx, commit, "ours.txt", "missing")
	if err != nil || !strings.Contains(a, "ours.txt") {
		t.Fatalf("ident %q, %v", a, err)
	}
	b, _ := r.f.Ident(ctx, commit, "README.md")
	if a == b {
		t.Error("different paths, same ident")
	}
}
