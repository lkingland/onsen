package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Source names the source the running onsen was built from. onsen sets it
// when it installs itself (go install -ldflags -X); a build made by hand
// leaves it empty, and the first update installs onsen again.
var Source string

// selfPaths are the fork's paths onsen is built from.
var selfPaths = []string{"go.mod", "go.sum", "cmd", "internal"}

// Self keeps the installed onsen built from the fork's bookmark.
type Self struct {
	// Current is the running build's Source.
	Current string
	// Reexec replaces this process with the program at path, with extra
	// added to the environment. It returns only on failure.
	Reexec func(path string, extra string) error
}

// reexecMarker is set in the environment of the onsen that Update starts,
// so a build that still reports the wrong source cannot loop.
const reexecMarker = "ONSEN_REEXEC"

// Update installs onsen from commit when commit's onsen source differs from
// the running build's, then continues the run as the new build: the new
// build repeats steps 1 and 2 (now with nothing to fetch) and carries on.
func (s *Self) Update(ctx context.Context, u *Updater, commit string) error {
	f := u.fork()
	listing, err := f.Ident(ctx, commit, selfPaths...)
	if err != nil {
		return err
	}
	source := sourceOf(listing)
	if source == s.Current {
		return nil
	}
	if u.Getenv(reexecMarker) != "" {
		u.say("onsen was installed again but reports an older source (%s, not %s). Continuing.", s.Current, source)
		return nil
	}
	var paths []string
	for _, line := range run.Lines(listing) {
		if _, path, ok := strings.Cut(line, "\t"); ok {
			paths = append(paths, path)
		}
	}
	u.say("onsen's own source changed in %s. Installing it again, then continuing as the new build.", u.Bookmark)
	target, err := s.install(ctx, u, commit, source, paths)
	if err != nil {
		// The installed onsen stays as it was, and so does this run: a
		// build that cannot be made (offline, no go) must not hold every
		// update behind it.
		u.problem(fmt.Errorf("onsen could not install itself from %s, so this run carries on as the installed build: %w", u.Bookmark, err))
		return nil
	}
	return s.Reexec(target, reexecMarker+"="+source)
}

// install builds onsen from commit, checks that the build starts, and only
// then puts it where go install would. It returns that path.
func (s *Self) install(ctx context.Context, u *Updater, commit, source string, paths []string) (string, error) {
	work, err := os.MkdirTemp("", "onsen-self.")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)
	src := filepath.Join(work, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		return "", err
	}
	tarFile := filepath.Join(work, "src.tar")
	if err := u.fork().Archive(ctx, commit, tarFile, paths...); err != nil {
		return "", err
	}
	if err := u.Run.Run(ctx, run.C("tar", "-x", "-f", tarFile, "-C", src)); err != nil {
		return "", err
	}
	built := filepath.Join(work, "onsen")
	ldflags := "-X github.com/lkingland/onsen/internal/update.Source=" + source
	if err := u.Run.Run(ctx, run.Cmd{Name: "go", Args: []string{"build", "-o", built, "-ldflags", ldflags, "./cmd/onsen"}, Dir: src}); err != nil {
		return "", err
	}
	if _, err := u.Run.Output(ctx, run.C(built, "--help")); err != nil {
		return "", fmt.Errorf("the new build does not start: %w", err)
	}
	out, err := u.Run.Output(ctx, run.Cmd{Name: "go", Args: []string{"list", "-f", "{{.Target}}", "./cmd/onsen"}, Dir: src})
	if err != nil {
		return "", err
	}
	target := strings.TrimSpace(out)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	// Beside the target, then renamed over it: the old build is never half
	// replaced.
	next := target + ".new"
	if err := copyFile(built, next, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(next, target); err != nil {
		os.Remove(next)
		return "", err
	}
	return target, nil
}

func copyFile(src, dest string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dest, data, mode); err != nil {
		return err
	}
	return os.Chmod(dest, mode)
}

// sourceOf names onsen's source by git's listing of its paths, which
// carries each one's object id.
func sourceOf(listing string) string {
	sum := sha256.Sum256([]byte(listing))
	return hex.EncodeToString(sum[:8])
}
