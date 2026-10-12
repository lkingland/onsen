package build

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lkingland/onsen/internal/fork"
	"github.com/lkingland/onsen/internal/run"
)

const commit = "0123456789abcdef0123456789abcdef01234567"

type bench struct {
	t         *testing.T
	b         Build
	f         *run.Fake
	tree      string            // the fork's tree id
	installed string            // pacman -Q omarchy-dev
	noSig     bool              // makepkg forgets the signature
	pkgvers   map[string]string // pkg -> the pkgver line makepkg saw
	srcSeen   string            // what OMARCHY_SRC held
	said      []string
}

func newBench(t *testing.T) *bench {
	if _, err := exec.LookPath("vercmp"); err != nil {
		t.Skip("no vercmp (pacman's)")
	}
	root := t.TempDir()
	for _, pkg := range []string{"omarchy-settings-dev", "omarchy-dev"} {
		dir := filepath.Join(root, "pkgs/pkgbuilds", pkg)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "pkgname=" + pkg + "\npkgver=4.0.0.r6897.g4df5459\npkgrel=3\n"
		if err := os.WriteFile(filepath.Join(dir, "PKGBUILD"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w := &bench{t: t, tree: "forktree", pkgvers: map[string]string{}}
	w.f = &run.Fake{Handle: w.handle}
	w.b = Build{
		Run:       w.f,
		Fork:      fork.Fork{Run: w.f, Dir: "/fork", Bookmark: "master"},
		PKGBUILDs: filepath.Join(root, "pkgs"),
		Packages:  []string{"omarchy-settings-dev", "omarchy-dev"},
		Key:       "KEY!",
		RepoDir:   filepath.Join(root, "repo"),
		RepoName:  "onsen",
		Record:    filepath.Join(root, "state/onsen-built"),
		User:      "drekar",
		Now:       func() time.Time { return time.Date(2026, 10, 13, 14, 5, 0, 0, time.FixedZone("x", 3600)) },
		Say:       func(s string) { w.said = append(w.said, s) },
	}
	return w
}

func (w *bench) handle(c run.Cmd) (string, error) {
	switch {
	case c.String() == "git -C /fork rev-parse "+commit+"^{tree}":
		return w.tree + "\n", nil
	case strings.HasSuffix(c.String(), "rev-parse HEAD^{tree}"):
		return "pkgtree\n", nil
	case c.String() == "git -C /fork rev-list --count "+commit:
		return "6917\n", nil
	case c.Name == "git" && c.Args[2] == "archive":
		// A tree with one file, as git archive would write it.
		dir := w.t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "version"), []byte("4.0.0\n"), 0o644); err != nil {
			return "", err
		}
		return "", exec.Command("tar", "-c", "-f", c.Args[4], "-C", dir, "version").Run()
	case c.Name == "tar" || c.Name == "cp":
		return "", exec.Command(c.Name, c.Args...).Run()
	case c.Name == "install" && c.Sudo:
		return "", os.MkdirAll(c.Args[len(c.Args)-1], 0o755)
	case c.String() == "pacman -Q omarchy-dev":
		if w.installed == "" {
			return "", errors.New("not installed")
		}
		return "omarchy-dev " + w.installed + "\n", nil
	case c.Name == "vercmp":
		out, err := exec.Command("vercmp", c.Args...).Output()
		return string(out), err
	case c.Name == "makepkg":
		data, err := os.ReadFile(filepath.Join(c.Dir, "PKGBUILD"))
		if err != nil {
			return "", err
		}
		pkg := filepath.Base(c.Dir)
		var ver string
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "pkgver=") {
				w.pkgvers[pkg] = line
				ver = strings.TrimPrefix(line, "pkgver=")
			}
			if strings.HasPrefix(line, "pkgrel=") && line != "pkgrel=1" {
				return "", errors.New("pkgrel not reset: " + line)
			}
		}
		for _, e := range c.Env {
			if src, ok := strings.CutPrefix(e, "OMARCHY_SRC="); ok {
				got, _ := os.ReadFile(filepath.Join(src, "version"))
				w.srcSeen = string(got)
			}
		}
		file := filepath.Join(c.Dir, pkg+"-"+ver+"-1-x86_64.pkg.tar.zst")
		if err := os.WriteFile(file, []byte(pkg), 0o644); err != nil {
			return "", err
		}
		if w.noSig {
			return "", nil
		}
		return "", os.WriteFile(file+".sig", []byte("sig"), 0o644)
	case c.Name == "repo-add":
		return "", os.WriteFile(filepath.Join(c.Dir, "onsen.db"), []byte("db"), 0o644)
	}
	return "", nil
}

func (w *bench) publish() (string, error) { return w.b.Publish(context.Background(), commit) }

func TestPublishBuildsSignsAndRecords(t *testing.T) {
	w := newBench(t)
	version, err := w.publish()
	if err != nil {
		t.Fatal(err)
	}
	const want = "4.0.0.202610131305.r6917.g0123456"
	if version != want {
		t.Errorf("version %s, want %s (UTC stamp, commit count, short sha)", version, want)
	}
	for _, pkg := range w.b.Packages {
		if w.pkgvers[pkg] != "pkgver="+want {
			t.Errorf("%s built with %q", pkg, w.pkgvers[pkg])
		}
	}
	if w.srcSeen != "4.0.0\n" {
		t.Errorf("OMARCHY_SRC held %q, want the fork's archived tree", w.srcSeen)
	}
	makepkg := w.f.Matching("makepkg")
	if !reflect.DeepEqual(makepkg, []string{"makepkg --force --noconfirm --nodeps --sign --key KEY!", "makepkg --force --noconfirm --nodeps --sign --key KEY!"}) {
		t.Errorf("makepkg %q", makepkg)
	}
	if order := []int{w.f.Index("makepkg"), w.f.Index("repo-add")}; order[0] > order[1] {
		t.Error("published before building")
	}
	repoAdd := w.f.Matching("repo-add")
	if len(repoAdd) != 1 || repoAdd[0] != "repo-add --sign --key KEY! --remove onsen.db.tar.zst omarchy-settings-dev-"+want+"-1-x86_64.pkg.tar.zst omarchy-dev-"+want+"-1-x86_64.pkg.tar.zst" {
		t.Errorf("repo-add %q", repoAdd)
	}
	for _, pkg := range w.b.Packages {
		for _, suffix := range []string{"", ".sig"} {
			if _, err := os.Stat(filepath.Join(w.b.RepoDir, pkg+"-"+want+"-1-x86_64.pkg.tar.zst"+suffix)); err != nil {
				t.Errorf("not in the repository: %v", err)
			}
		}
	}
	rec, _ := os.ReadFile(w.b.Record)
	if string(rec) != "forktree+pkgtree 0123456789ab "+want+"\n" {
		t.Errorf("record %q", rec)
	}
	if w.f.Index("sudo install -d -m 0755 -o drekar -g drekar "+w.b.RepoDir) < 0 {
		t.Errorf("repository not created as the user: %q", w.f.Lines())
	}
}

func TestPublishSkipsWhatIsPublished(t *testing.T) {
	w := newBench(t)
	first, err := w.publish()
	if err != nil {
		t.Fatal(err)
	}
	n := len(w.f.Matching("makepkg"))
	w.b.Now = func() time.Time { return time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC) }
	again, err := w.publish()
	if err != nil || again != first {
		t.Fatalf("again: %s, %v; want %s", again, err, first)
	}
	if len(w.f.Matching("makepkg")) != n {
		t.Error("rebuilt an unchanged source")
	}

	// A changed fork tree builds again; so does a missing repository database.
	w.tree = "othertree"
	if v, err := w.publish(); err != nil || v == first || len(w.f.Matching("makepkg")) != n+2 {
		t.Errorf("changed source: %s, %v", v, err)
	}
	if err := os.Remove(filepath.Join(w.b.RepoDir, "onsen.db")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.publish(); err != nil || len(w.f.Matching("makepkg")) != n+4 {
		t.Errorf("missing database: %v, %d builds", err, len(w.f.Matching("makepkg")))
	}
}

func TestPublishRefusesAVersionThatDoesNotSortAbove(t *testing.T) {
	w := newBench(t)
	w.installed = "4.0.0.209912312359.r1.gabcdef0-1"
	if _, err := w.publish(); err == nil || !strings.Contains(err.Error(), "would not sort above 4.0.0.209912312359.r1.gabcdef0") {
		t.Fatalf("err %v", err)
	}
	if w.f.Index("makepkg") >= 0 {
		t.Error("built anyway")
	}
}

func TestPublishNeedsSignatures(t *testing.T) {
	w := newBench(t)
	w.noSig = true
	if _, err := w.publish(); err == nil {
		t.Fatal("published an unsigned package")
	}
	if w.f.Index("repo-add") >= 0 || w.f.Index("cp -f") >= 0 {
		t.Error("published anyway")
	}
	if _, err := os.Stat(w.b.Record); err == nil {
		t.Error("recorded a failed build")
	}
}

func TestVersionNeedsUpstreamPrefix(t *testing.T) {
	w := newBench(t)
	path := w.b.pkgbuild("omarchy-dev")
	if err := os.WriteFile(path, []byte("pkgver=4.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := w.publish(); err == nil {
		t.Fatal("built without an upstream prefix")
	}
}

func TestPublishedReadsRecord(t *testing.T) {
	w := newBench(t)
	if id, v, err := w.b.Published(); err != nil || id != "" || v != "" {
		t.Errorf("no record: %q %q %v", id, v, err)
	}
	if err := os.MkdirAll(filepath.Dir(w.b.Record), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(w.b.Record, []byte("a+b 0123 4.0.0.x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if id, v, err := w.b.Published(); err != nil || id != "a+b" || v != "4.0.0.x" {
		t.Errorf("record: %q %q %v", id, v, err)
	}
}
