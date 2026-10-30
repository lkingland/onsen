package packages

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lkingland/onsen/internal/run"
)

// machine is a fake pacman: installed packages, the explicit ones marked.
type machine struct {
	installed map[string]bool // name -> explicit
	fail      string          // a pacman flag that fails, e.g. "-Qqe"
}

func (m *machine) runner() *run.Fake {
	return &run.Fake{Handle: func(c run.Cmd) (string, error) {
		if c.Name == "pacman" && len(c.Args) > 0 && c.Args[0] == m.fail {
			return "", errors.New("pacman: database is locked")
		}
		if c.Name != "pacman" || len(c.Args) == 0 {
			return "", nil
		}
		var out []string
		switch c.Args[0] {
		case "-Qq":
			for n := range m.installed {
				out = append(out, n)
			}
		case "-Qqe":
			for n, explicit := range m.installed {
				if explicit {
					out = append(out, n)
				}
			}
		}
		return strings.Join(out, "\n") + "\n", nil
	}}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// set is a declared list and Quattro's lists in two install trees: the
// fork's checkout and the installed copy.
func set(t *testing.T) *Set {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "packages"), `# The edition's packages
neovim            # the editor
nnn
aur:bat-extras      # from the AUR
local:onsen-matrix  # built by hand from its checkout
-omarchy-nvim     # the starter config, replaced by ours

`)
	write(t, filepath.Join(root, "fork/install/omarchy-base.packages"), "hyprland\nbtop # monitor\n")
	write(t, filepath.Join(root, "fork/install/omarchy-other.packages"), "linux-omarchy\n")
	write(t, filepath.Join(root, "usr/install/omarchy-x86_64-only.packages"), "intel-ucode\n")
	s, err := Read(filepath.Join(root, "packages"), filepath.Join(root, "fork/install"), filepath.Join(root, "usr/install"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRead(t *testing.T) {
	s := set(t)
	for what, c := range map[string][2][]string{
		"repo":     {s.Repo, {"neovim", "nnn"}},
		"aur":      {s.AUR, {"bat-extras"}},
		"remove":   {s.Remove, {"omarchy-nvim"}},
		"quattro":  {s.Quattro, {"btop", "hyprland", "intel-ucode", "linux-omarchy"}},
		"declared": {s.Declared(), {"bat-extras", "neovim", "nnn", "omarchy-dev", "omarchy-settings-dev", "onsen-matrix"}},
	} {
		if !reflect.DeepEqual(c[0], c[1]) {
			t.Errorf("%s: got %q, want %q", what, c[0], c[1])
		}
	}
}

// Review finding #7 on Hamr#762-3: the installed copy's Quattro lists allow
// a package too, not only the checkout's.
func TestUndeclared(t *testing.T) {
	m := &machine{installed: map[string]bool{
		"neovim": true, "bat-extras": true, "omarchy-dev": true, // declared
		"hyprland": true, "intel-ucode": true, // Quattro's, in either tree
		"omarchy-nvim":     true, // removed by the same run: not drift
		"omarchy":          true, // replaced by ours in the same run
		"omarchy-settings": true,
		"steam":            true,  // a hand install
		"glibc":            false, // a dependency is never drift
	}}
	got, err := Pacman{Run: m.runner()}.Undeclared(context.Background(), set(t))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"steam"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A crashed query is an error, never an empty answer: an empty answer reads
// as "no drift" (review finding #5 on Hamr#762-3).
func TestQueryFailureIsAnError(t *testing.T) {
	for _, flag := range []string{"-Qqe", "-Qq"} {
		m := &machine{installed: map[string]bool{"steam": true}, fail: flag}
		p := Pacman{Run: m.runner()}
		ctx := context.Background()
		if flag == "-Qqe" {
			if _, err := p.Undeclared(ctx, set(t)); err == nil {
				t.Errorf("%s: Undeclared ignored the failure", flag)
			}
		}
		for name, step := range map[string]func(context.Context, *Set) error{"Mark": p.Mark, "Install": p.Install, "Remove": p.Remove} {
			if name == "Remove" && flag == "-Qqe" {
				continue // Remove asks only for -Qq
			}
			if err := step(ctx, set(t)); err == nil {
				t.Errorf("%s: %s ignored the failure", flag, name)
			}
		}
	}
}

func TestRemoveOnlyWhatIsInstalled(t *testing.T) {
	m := &machine{installed: map[string]bool{"omarchy-nvim": true}}
	f := m.runner()
	if err := (Pacman{Run: f}).Remove(context.Background(), set(t)); err != nil {
		t.Fatal(err)
	}
	if got := f.Matching("omarchy-update-pacman -R"); !reflect.DeepEqual(got, []string{"omarchy-update-pacman -R --noconfirm omarchy-nvim"}) {
		t.Errorf("got %q", got)
	}

	f = (&machine{installed: map[string]bool{}}).runner()
	if err := (Pacman{Run: f}).Remove(context.Background(), set(t)); err != nil {
		t.Fatal(err)
	}
	if got := f.Matching("omarchy-update-pacman -R"); got != nil {
		t.Errorf("removed a package that is not installed: %q", got)
	}
}

func TestMarkOnlyDeclaredDependencies(t *testing.T) {
	m := &machine{installed: map[string]bool{
		"neovim":     false, // declared, a dependency: mark it
		"nnn":        true,  // declared, explicit already
		"bat-extras": false, // declared aur:, a dependency: mark it
		"glibc":      false, // not declared: leave it
	}}
	f := m.runner()
	if err := (Pacman{Run: f}).Mark(context.Background(), set(t)); err != nil {
		t.Fatal(err)
	}
	if got := f.Matching("omarchy-update-pacman -D"); !reflect.DeepEqual(got, []string{"omarchy-update-pacman -D --asexplicit bat-extras neovim"}) {
		t.Errorf("got %q", got)
	}
}

func TestInstallMissingThenMark(t *testing.T) {
	m := &machine{installed: map[string]bool{"nnn": true, "onsen-matrix": true}}
	f := m.runner()
	if err := (Pacman{Run: f}).Install(context.Background(), set(t)); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"omarchy-update-pacman -S --needed --noconfirm neovim",
		"yay -S --needed --noconfirm bat-extras",
	}
	var got []string
	for _, line := range f.Lines() {
		if strings.Contains(line, " -S ") {
			got = append(got, line)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	// Mark queries again after the installs.
	if last := f.Lines()[len(f.Lines())-1]; last != "pacman -Qqe" {
		t.Errorf("last command %q, want Mark's queries after the installs", last)
	}
	if yay := f.Calls[f.Index("yay")]; yay.Sudo {
		t.Error("yay runs as root")
	}
}

func TestLocalIsDeclaredNeverInstalledAndNamedWhenMissing(t *testing.T) {
	s := set(t)
	if !reflect.DeepEqual(s.Local, []string{"onsen-matrix"}) {
		t.Errorf("Local = %q", s.Local)
	}
	declared := strings.Join(s.Declared(), " ")
	if !strings.Contains(declared, "onsen-matrix") {
		t.Errorf("Declared() = %q, lacks the local package", declared)
	}
	// Everything but the local package is installed: the run installs
	// nothing, and names the local package as missing.
	m := &machine{installed: map[string]bool{"neovim": true, "nnn": true, "bat-extras": true}}
	f := m.runner()
	err := (Pacman{Run: f}).Install(context.Background(), s)
	if err == nil || !strings.Contains(err.Error(), "onsen-matrix") {
		t.Errorf("err = %v, want the missing local package named", err)
	}
	for _, line := range f.Lines() {
		if strings.Contains(line, " -S ") {
			t.Errorf("tried to install something: %q", line)
		}
	}
	// Installed as a dependency, it is marked explicit like the rest.
	m = &machine{installed: map[string]bool{"neovim": true, "nnn": true, "bat-extras": true, "onsen-matrix": false}}
	f = m.runner()
	if err := (Pacman{Run: f}).Install(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(f.Lines(), "\n"), "--asexplicit onsen-matrix") {
		t.Errorf("local package not marked explicit: %q", f.Lines())
	}
}

func TestInstallNothingMissing(t *testing.T) {
	m := &machine{installed: map[string]bool{"neovim": true, "nnn": true, "bat-extras": true, "onsen-matrix": true}}
	f := m.runner()
	if err := (Pacman{Run: f}).Install(context.Background(), set(t)); err != nil {
		t.Fatal(err)
	}
	for _, line := range f.Lines() {
		if strings.HasPrefix(line, "sudo") || strings.HasPrefix(line, "yay") {
			t.Errorf("changed something: %q", line)
		}
	}
}
