package channel

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lkingland/onsen/internal/run"
)

const edge = `# See the pacman.conf(5) manpage
[options]
HoldPkg = pacman glibc
SigLevel = Required DatabaseOptional

# Onsen edition: this machine's own build. It comes first.
[onsen]
SigLevel = Required
Server = file:///var/lib/omarchy-onsen

[omarchy]
Server = https://pkgs.omarchy.org/edge/$arch

[core]
Include = /etc/pacman.d/mirrorlist

# [core-debug]
# Server = https://geo.mirror.pkgbuild.com/$repo/os/$arch
`

// What upstream migration 1791403252 leaves: [omarchy] moved to the front.
const moved = `# See the pacman.conf(5) manpage
[options]
HoldPkg = pacman glibc
SigLevel = Required DatabaseOptional

# Onsen edition: this machine's own build. It comes first.
[omarchy]
Server = https://pkgs.omarchy.org/edge/$arch

[onsen]
SigLevel = Required
Server = file:///var/lib/omarchy-onsen

[core]
Include = /etc/pacman.d/mirrorlist

# [core-debug]
# Server = https://geo.mirror.pkgbuild.com/$repo/os/$arch
`

func TestReorder(t *testing.T) {
	if got := Reorder(edge, "onsen"); got != edge {
		t.Errorf("reordered a conf that was already in order:\n%s", got)
	}
	got := Reorder(moved, "onsen")
	if !strings.Contains(got, "[options]\nHoldPkg") || strings.Index(got, "[onsen]") > strings.Index(got, "[omarchy]") {
		t.Errorf("[onsen] is not first:\n%s", got)
	}
	if again := Reorder(got, "onsen"); again != got {
		t.Error("Reorder is not idempotent")
	}
	if Reorder("", "onsen") != "" {
		t.Error("an empty conf grew")
	}
}

// v1 (the shell script onsen replaced) did this with awk. Same input, same
// output, including the corners: indented headers, comments after a header,
// no trailing newline, a conf without [onsen].
func TestReorderMatchesV1(t *testing.T) {
	if _, err := exec.LookPath("awk"); err != nil {
		t.Skip("no awk")
	}
	const v1 = `
    /^[[:space:]]*\[[^]]+\]/ {
      section = $0
      sub(/^[[:space:]]*\[/, "", section)
      sub(/\].*$/, "", section)
      if (section != "options") repositories = 1
    }
    !repositories { prefix = prefix $0 "\n"; next }
    section == repo { ours = ours $0 "\n"; next }
    { others = others $0 "\n" }
    END { printf "%s%s%s", prefix, ours, others }`
	for _, conf := range []string{
		edge, moved,
		strings.TrimSuffix(moved, "\n"),
		"[options]\n[core]\nInclude = x\n  [onsen]  # ours\nServer = y\n\n\n",
		"[core]\n[extra]\n",
		"# nothing\n",
		"[options]\nX = 1\n[omarchy]\n[onsen]\n[onsen-other]\n[onsen]\nServer = z\n",
	} {
		cmd := exec.Command("awk", "-v", "repo=onsen", v1)
		cmd.Stdin = strings.NewReader(conf)
		want, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if got := Reorder(conf, "onsen"); got != string(want) {
			t.Errorf("for %q:\n got %q\nwant %q", conf, got, want)
		}
	}
}

// fakeRoot runs "sudo install ... src dest" and "sudo cp -f src dest" as
// copies, so the files the fake machine writes are real.
func fakeRoot(t *testing.T, handle func(run.Cmd) (string, error)) *run.Fake {
	return &run.Fake{Handle: func(c run.Cmd) (string, error) {
		if c.Sudo && c.Name == "mv" {
			return "", os.Rename(c.Args[1], c.Args[2])
		}
		if c.Sudo && (c.Name == "install" || c.Name == "cp") {
			src, dest := c.Args[len(c.Args)-2], c.Args[len(c.Args)-1]
			data, err := os.ReadFile(src)
			if err != nil {
				return "", err
			}
			return "", os.WriteFile(dest, data, 0o644)
		}
		if handle != nil {
			return handle(c)
		}
		return "", nil
	}}
}

func newChannel(t *testing.T, conf string, f *run.Fake) (Channel, *[]string) {
	dir := t.TempDir()
	var said []string
	c := Channel{
		Run: f, Repo: "onsen",
		Conf:           filepath.Join(dir, "pacman.conf"),
		Mirrorlist:     filepath.Join(dir, "mirrorlist"),
		EdgeConf:       filepath.Join(dir, "pacman-edge.conf"),
		EdgeMirrorlist: filepath.Join(dir, "mirrorlist-edge"),
		Say:            func(s string) { said = append(said, s) },
	}
	for path, body := range map[string]string{c.Conf: conf, c.Mirrorlist: "Server = stable\n", c.EdgeConf: edge, c.EdgeMirrorlist: "Server = edge\n"} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return c, &said
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestKeepFirst(t *testing.T) {
	f := fakeRoot(t, nil)
	c, said := newChannel(t, moved, f)
	changed, err := c.KeepFirst(context.Background())
	if err != nil || !changed {
		t.Fatalf("changed %v, err %v", changed, err)
	}
	if got := read(t, c.Conf); got != Reorder(moved, "onsen") {
		t.Errorf("wrote:\n%s", got)
	}
	if f.Index("sudo install -m 0644 -o root -g root ") < 0 || f.Index("sudo mv -f "+c.Conf+".onsen-new "+c.Conf) < 0 {
		t.Errorf("not written as root beside the file, then renamed: %q", f.Lines())
	}
	if _, err := os.Lstat(c.Conf + ".onsen-new"); err == nil {
		t.Error("the new file was left beside pacman.conf")
	}
	if len(*said) != 1 {
		t.Errorf("said %q", *said)
	}

	f = fakeRoot(t, nil)
	c, _ = newChannel(t, edge, f)
	if changed, err := c.KeepFirst(context.Background()); err != nil || changed || len(f.Calls) != 0 {
		t.Errorf("in order already: changed %v, err %v, ran %q", changed, err, f.Lines())
	}
}

func TestSwitchOncePerMachine(t *testing.T) {
	f := fakeRoot(t, func(c run.Cmd) (string, error) { return "x86\n", nil })
	c, _ := newChannel(t, "[options]\n[core]\n", f)
	if err := c.Switch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if read(t, c.Conf) != edge || read(t, c.Mirrorlist) != "Server = edge\n" {
		t.Error("the edge templates were not installed")
	}
	if read(t, c.Conf+".bak") != "[options]\n[core]\n" || read(t, c.Mirrorlist+".bak") != "Server = stable\n" {
		t.Error("the old files were not kept")
	}
	n := len(f.Calls)
	if err := c.Switch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(f.Calls) != n {
		t.Errorf("switched again: %q", f.Lines()[n:])
	}

	// Something dropped [onsen]: switch again, but keep the first backup.
	if err := os.WriteFile(c.Conf, []byte("[options]\n[extra]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Switch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if read(t, c.Conf) != edge || read(t, c.Conf+".bak") != "[options]\n[core]\n" {
		t.Error("the pre-Onsen backup was overwritten")
	}
}

func TestSwitchRefusesOtherPlatforms(t *testing.T) {
	f := fakeRoot(t, func(c run.Cmd) (string, error) { return "aarch64-apple\n", nil })
	c, _ := newChannel(t, "[options]\n[core]\n", f)
	if err := c.Switch(context.Background()); err == nil {
		t.Fatal("switched on aarch64")
	}
	if read(t, c.Conf) != "[options]\n[core]\n" {
		t.Error("pacman.conf changed")
	}
}

func TestTrustKey(t *testing.T) {
	const colons = "pub:u:255:22:7DB461E762F282F9:::::\nfpr:::::::::ABCDEF0123456789ABCDEF0123456789ABCDEF01:\nfpr:::::::::SUBKEY:\n"
	for _, known := range []bool{true, false} {
		var added string
		f := &run.Fake{Handle: func(c run.Cmd) (string, error) {
			switch c.String() {
			case "gpg --with-colons --list-keys 7DB461E762F282F9":
				return colons, nil
			case "sudo pacman-key --list-keys ABCDEF0123456789ABCDEF0123456789ABCDEF01":
				if !known {
					return "", errors.New("not found")
				}
			case "gpg --export --armor ABCDEF0123456789ABCDEF0123456789ABCDEF01":
				return "-----BEGIN PGP PUBLIC KEY BLOCK-----\n", nil
			}
			if c.Name == "pacman-key" && c.Args[0] == "--add" {
				data, _ := os.ReadFile(c.Args[1])
				added = string(data)
			}
			return "", nil
		}}
		c, _ := newChannel(t, edge, f)
		if err := c.TrustKey(context.Background(), "7DB461E762F282F9!"); err != nil {
			t.Fatal(err)
		}
		lsign := f.Index("sudo pacman-key --lsign-key ABCDEF0123456789ABCDEF0123456789ABCDEF01")
		add := f.Index("sudo pacman-key --add ")
		if known {
			if add >= 0 || lsign >= 0 {
				t.Errorf("known key must be a no-op; got %q", f.Lines())
			}
			continue
		}
		if lsign < 0 {
			t.Errorf("unknown key: never signed locally: %q", f.Lines())
		}
		if add < 0 || add > lsign || !strings.HasPrefix(added, "-----BEGIN") {
			t.Errorf("unknown key: added %q at %d, lsign at %d", added, add, lsign)
		}
	}
}

func TestTrustKeyWithoutPublicKey(t *testing.T) {
	f := &run.Fake{Handle: func(c run.Cmd) (string, error) { return "pub:u:255:22:X:::::\n", nil }}
	c, _ := newChannel(t, edge, f)
	if err := c.TrustKey(context.Background(), "X"); err == nil {
		t.Fatal("trusted a key with no fingerprint")
	}
	if got := f.Matching("sudo"); got != nil {
		t.Errorf("ran %q", got)
	}
}

func TestListed(t *testing.T) {
	for conf, want := range map[string]bool{edge: true, "[options]\n[onsen-x]\n": false, "  [onsen]\n": false} {
		c, _ := newChannel(t, conf, &run.Fake{})
		if got, err := c.Listed(); err != nil || got != want {
			t.Errorf("Listed(%q) = %v, %v", conf, got, err)
		}
	}
}
