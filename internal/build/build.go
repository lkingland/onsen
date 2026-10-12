// Package build builds the edition's packages from the fork with upstream's
// PKGBUILDs, signs them with the card, and publishes them to the local
// [onsen] repository.
package build

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lkingland/onsen/internal/fork"
	"github.com/lkingland/onsen/internal/run"
)

// Build is one fork, one set of PKGBUILDs, and one repository.
type Build struct {
	Run       run.Runner
	Fork      fork.Fork
	PKGBUILDs string   // a checkout of omacom-io/omarchy-pkgs
	Packages  []string // ours, in build order
	Key       string   // the signing key, as makepkg and repo-add take it
	RepoDir   string   // /var/lib/omarchy-onsen
	RepoName  string   // onsen
	Record    string   // the built record: "<source id> <commit12> <version>"
	User      string   // owns RepoDir
	Now       func() time.Time
	Say       func(string)
}

// pkgbuild is the PKGBUILD for one of our packages.
func (b Build) pkgbuild(pkg string) string {
	return filepath.Join(b.PKGBUILDs, "pkgbuilds", pkg, "PKGBUILD")
}

// SourceID names what a build is made from: the fork's tree at commit and
// the PKGBUILDs' tree.
func (b Build) SourceID(ctx context.Context, commit string) (string, error) {
	fork, err := b.Fork.Tree(ctx, commit)
	if err != nil {
		return "", err
	}
	out, err := b.Run.Output(ctx, run.C("git", "-C", b.PKGBUILDs, "rev-parse", "HEAD^{tree}"))
	if err != nil {
		return "", err
	}
	return fork + "+" + strings.TrimSpace(out), nil
}

// Published is the version the record says was last published, and the
// record's source id; empty when nothing has been.
func (b Build) Published() (sourceID, version string, err error) {
	data, err := os.ReadFile(b.Record)
	if errors.Is(err, fs.ErrNotExist) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	f := strings.Fields(string(data))
	if len(f) != 3 {
		return "", "", fmt.Errorf("%s: not \"<source id> <commit> <version>\"", b.Record)
	}
	return f[0], f[2], nil
}

var pkgver = regexp.MustCompile(`(?m)^pkgver=(.*)$`)

// Version is the version a build of commit gets:
// <upstream prefix>.<UTC stamp>.r<commits>.g<sha>. The stamp leads, so a
// build always sorts above every earlier one (even after a squash lowers the
// count) and above upstream's edge build, 4.0.0.r<commits>.g<sha>.
func (b Build) Version(ctx context.Context, commit string) (version, upstream string, err error) {
	data, err := os.ReadFile(b.pkgbuild("omarchy-dev"))
	if err != nil {
		return "", "", err
	}
	m := pkgver.FindStringSubmatch(string(data))
	if m == nil {
		return "", "", fmt.Errorf("no pkgver in omarchy-dev's PKGBUILD")
	}
	upstream = m[1]
	prefix, _, found := strings.Cut(upstream, ".r")
	if !found || prefix == "" {
		return "", "", fmt.Errorf("cannot read upstream's version prefix from omarchy-dev's PKGBUILD (%s)", upstream)
	}
	count, err := b.Fork.Count(ctx, commit)
	if err != nil {
		return "", "", err
	}
	stamp := b.Now().UTC().Format("200601021504")
	return fmt.Sprintf("%s.%s.r%d.g%s", prefix, stamp, count, commit[:7]), upstream, nil
}

// vercmp compares two versions the way pacman does.
func (b Build) vercmp(ctx context.Context, a, c string) (int, error) {
	out, err := b.Run.Output(ctx, run.C("vercmp", a, c))
	if err != nil {
		return 0, err
	}
	var n int
	if _, err := fmt.Sscan(out, &n); err != nil {
		return 0, fmt.Errorf("vercmp %s %s: %q", a, c, out)
	}
	return n, nil
}

// installed is the installed version of a package without its pkgrel, or
// "" when it is not installed.
func (b Build) installed(ctx context.Context, pkg string) string {
	out, err := b.Run.Output(ctx, run.Cmd{Name: "pacman", Args: []string{"-Q", pkg}, Quiet: true})
	if err != nil {
		return ""
	}
	f := strings.Fields(out)
	if len(f) < 2 {
		return ""
	}
	v := f[1]
	if i := strings.LastIndexByte(v, '-'); i >= 0 {
		v = v[:i]
	}
	return v
}

// Publish builds and publishes commit, unless the record says this source
// was already published and the repository is there. It returns the
// version now published.
func (b Build) Publish(ctx context.Context, commit string) (string, error) {
	sourceID, err := b.SourceID(ctx, commit)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(b.RepoDir); errors.Is(err, fs.ErrNotExist) {
		b.Say(fmt.Sprintf("Creating the [%s] repository at %s.", b.RepoName, b.RepoDir))
		if err := b.Run.Run(ctx, run.Root("install", "-d", "-m", "0755", "-o", b.User, "-g", b.User, b.RepoDir)); err != nil {
			return "", err
		}
	}
	recorded, version, err := b.Published()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(b.RepoDir, b.RepoName+".db")); err == nil && recorded == sourceID {
		b.Say("Already published: " + version + ".")
		return version, nil
	}

	version, upstream, err := b.Version(ctx, commit)
	if err != nil {
		return "", err
	}
	for _, older := range []string{upstream, b.installed(ctx, "omarchy-dev")} {
		if older == "" {
			continue
		}
		n, err := b.vercmp(ctx, version, older)
		if err != nil {
			return "", err
		}
		if n <= 0 {
			return "", fmt.Errorf("version %s would not sort above %s", version, older)
		}
	}

	work, err := os.MkdirTemp("", "omarchy-onsen.")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)
	src := filepath.Join(work, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		return "", err
	}
	tarFile := filepath.Join(work, "src.tar")
	if err := b.Fork.Archive(ctx, commit, tarFile); err != nil {
		return "", err
	}
	if err := b.Run.Run(ctx, run.C("tar", "-x", "-f", tarFile, "-C", src)); err != nil {
		return "", err
	}

	var built []string
	for _, pkg := range b.Packages {
		b.Say(fmt.Sprintf("Building %s %s.", pkg, version))
		dir := filepath.Join(work, pkg)
		if err := b.Run.Run(ctx, run.C("cp", "-a", filepath.Dir(b.pkgbuild(pkg)), dir)); err != nil {
			return "", err
		}
		if err := setVersion(filepath.Join(dir, "PKGBUILD"), version); err != nil {
			return "", err
		}
		// --nodeps: omarchy-dev depends on omarchy-settings-dev, which is built
		// just before it and installed only in the channel step, so makepkg's
		// own check of the runtime dependencies cannot pass here. pacman
		// enforces them when the packages install.
		makepkg := run.Cmd{Name: "makepkg", Args: []string{"--force", "--noconfirm", "--nodeps", "--sign", "--key", b.Key},
			Dir: dir, Env: []string{"OMARCHY_SRC=" + src}}
		if err := b.Run.Run(ctx, makepkg); err != nil {
			return "", err
		}
		file, err := signedPackage(dir, pkg, version)
		if err != nil {
			return "", err
		}
		built = append(built, file)
	}

	b.Say(fmt.Sprintf("Publishing to [%s].", b.RepoName))
	names := make([]string, 0, len(built))
	for _, file := range built {
		if err := b.Run.Run(ctx, run.C("cp", "-f", file, file+".sig", b.RepoDir+"/")); err != nil {
			return "", err
		}
		names = append(names, filepath.Base(file))
	}
	repoAdd := run.Cmd{Name: "repo-add", Args: append([]string{"--sign", "--key", b.Key, "--remove", b.RepoName + ".db.tar.zst"}, names...), Dir: b.RepoDir}
	if err := b.Run.Run(ctx, repoAdd); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(b.Record), 0o755); err != nil {
		return "", err
	}
	record := fmt.Sprintf("%s %s %s\n", sourceID, commit[:12], version)
	if err := os.WriteFile(b.Record, []byte(record), 0o644); err != nil {
		return "", err
	}
	return version, nil
}

var pkgrel = regexp.MustCompile(`(?m)^pkgrel=.*$`)

// setVersion sets a PKGBUILD's pkgver to version and its pkgrel to 1.
func setVersion(path, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := pkgver.ReplaceAllLiteralString(string(data), "pkgver="+version)
	s = pkgrel.ReplaceAllLiteralString(s, "pkgrel=1")
	return os.WriteFile(path, []byte(s), 0o644)
}

// signedPackage is the package makepkg left in dir for pkg at version; it
// must have its signature beside it.
func signedPackage(dir, pkg, version string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pkg+"-"+version+"-*.pkg.tar.*"))
	if err != nil {
		return "", err
	}
	for _, m := range matches {
		if strings.HasSuffix(m, ".sig") {
			continue
		}
		if _, err := os.Stat(m + ".sig"); err == nil {
			return m, nil
		}
	}
	return "", fmt.Errorf("makepkg left no signed package for %s", pkg)
}
