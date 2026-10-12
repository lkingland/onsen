package update

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lkingland/onsen/internal/channel"
	"github.com/lkingland/onsen/internal/identity"
	"github.com/lkingland/onsen/internal/run"
)

const (
	commit    = "c0ffee0123456789c0ffee0123456789c0ffee01"
	version   = "4.0.0.202610130000.r100.gc0ffee0"
	stockConf = "[options]\nHoldPkg = pacman\n\n[omarchy]\nServer = edge\n\n[onsen]\nServer = file:///repo\n\n[core]\nInclude = x\n"
)

// machine is a fake Onsen box: files on disk under a temp root, and
// pacman, systemd, jj and the rest answered from the fields below.
type machine struct {
	t   *testing.T
	cfg Config
	f   *run.Fake

	pkgs       map[string]bool   // installed package -> explicit
	versions   map[string]string // installed version of ours
	enabled    map[string]bool   // enabled units ("user x" / "system x")
	dirty      string            // git status --porcelain
	offBook    bool              // the working copy is not on master
	behind     int               // new upstream commits
	conflict   bool              // the rebase conflicts
	failQqe    bool              // pacman -Qqe crashes
	starter    []string          // pacman -Qlq omarchy-nvim
	pkgsHead   string            // the PKGBUILDs' HEAD; the pull moves it to pkgsPulled
	pkgsPulled string
	nvimLog    string // git log of the pull under pkgbuilds/omarchy-nvim
	sudoAsks   bool   // sudo -n fails until sudo -v has run
	slowUpdate bool   // omarchy update takes a moment (the sudo refresh ticks)

	// What omarchy update does, besides running: upstream migration
	// 1791403252 putting [omarchy] first, a migration editing a file, and
	// (when stockOurs) stock packages replacing ours.
	moveOmarchy bool
	migrate     func()
	stockOurs   bool
	sticky      bool // reinstalling ours does not take

	// What installing and removing packages does to /etc/skel: removing
	// omarchy-nvim takes its starter files, and installing our settings
	// package ships the edition's.
	editionSkel func()

	// Observations.
	pendingDuringUpdate bool
	reboots             int
	atReboot            struct{ pending, gold bool }
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

func newMachine(t *testing.T) *machine {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	cfg := Defaults(home, "drekar")
	cfg.Fork = filepath.Join(root, "fork")
	cfg.PKGBUILDs = filepath.Join(root, "pkgs")
	cfg.RepoDir = filepath.Join(root, "repo")
	cfg.Skel = filepath.Join(root, "skel")
	cfg.PacmanConf = filepath.Join(root, "pacman.conf")
	cfg.Mirrorlist = filepath.Join(root, "mirrorlist")
	cfg.QuattroInstall = filepath.Join(root, "usr/install")
	cfg.Identity = filepath.Join(home, "state/identity")
	// The machine's identity, as root would see it, under the temp root.
	cfg.Roots = []identity.Root{
		{Path: filepath.Join(root, "var/lib/tailscale/tailscaled.state"), Mode: 0o600},
		{Path: filepath.Join(root, "etc/ssh"), Match: "ssh_host_*", Mode: 0o600},
	}

	m := &machine{t: t, cfg: cfg,
		pkgs: map[string]bool{
			"omarchy-dev": true, "omarchy-settings-dev": true, "hyprland": true,
			"neovim": true, "nnn": true, "bat-extras": true, "linux-omarchy": true,
		},
		versions: map[string]string{"omarchy-dev": version + "-1", "omarchy-settings-dev": version + "-1"},
		enabled:  map[string]bool{"system pcscd.socket": true, "user syncthing.service": true},
	}

	// The fork.
	write(t, filepath.Join(cfg.Fork, ".jj/repo"), "")
	write(t, cfg.onsen("packages"), "neovim\nnnn\naur:bat-extras\n-omarchy-nvim\n")
	write(t, cfg.onsen("services"), "system pcscd.socket\nuser syncthing.service\n")
	write(t, cfg.onsen("drift-ignore"), ".config/tmux/session-order\n")
	write(t, filepath.Join(cfg.Fork, "install/omarchy-base.packages"), "hyprland\nomarchy-nvim\n")
	// The installed copy of Quattro's lists, which can be newer or older
	// than the checkout's (review finding #7 on Hamr#762-3).
	write(t, filepath.Join(cfg.QuattroInstall, "omarchy-other.packages"), "linux-omarchy\n")
	write(t, filepath.Join(cfg.Fork, "default/pacman/pacman-edge.conf"), stockConf)
	write(t, filepath.Join(cfg.Fork, "default/pacman/mirrorlist-edge"), "Server = edge\n")
	write(t, filepath.Join(cfg.Fork, "config/nvim/init.lua"), "-- ours\n")
	write(t, filepath.Join(cfg.Fork, "config/nvim/lua/config/options.lua"), "-- ours\n")
	write(t, filepath.Join(cfg.PKGBUILDs, "pkgbuilds/omarchy-dev/PKGBUILD"), "pkgver=4.0.0.r1.gabc\n")

	// pacman, with [onsen] first, and ours published already.
	write(t, cfg.PacmanConf, channel.Reorder(stockConf, "onsen"))
	write(t, cfg.Mirrorlist, "Server = edge\n")
	write(t, filepath.Join(cfg.RepoDir, "onsen.db"), "db")
	write(t, cfg.record(), "tree+pkgtree c0ffee012345 "+version+"\n")

	// The edition's config, synced once.
	write(t, filepath.Join(cfg.Skel, ".config/tmux/tmux.conf"), "edition\n")
	write(t, filepath.Join(cfg.Skel, ".config/nvim/init.lua"), "-- ours\n")
	write(t, filepath.Join(cfg.Skel, ".config/nvim/lua/config/options.lua"), "-- ours\n")
	write(t, filepath.Join(cfg.Skel, ".gnupg/common.conf"), "use-keyboxd\n")
	for _, rel := range []string{".config/tmux/tmux.conf", ".config/nvim/init.lua", ".config/nvim/lua/config/options.lua", ".gnupg/common.conf"} {
		data, _ := os.ReadFile(filepath.Join(cfg.Skel, rel))
		write(t, filepath.Join(home, rel), string(data))
	}
	m.f = &run.Fake{Handle: m.handle, Programs: map[string]bool{}}
	u := m.updater()
	if err := u.applyConfig(context.Background()); err != nil {
		t.Fatal(err)
	}
	m.f.Calls = nil
	return m
}

func (m *machine) updater() *Updater {
	return &Updater{
		Config: m.cfg, Run: m.f, Out: &bytes.Buffer{}, Err: &bytes.Buffer{},
		Now:    func() time.Time { return time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC) },
		Getenv: func(string) string { return "" },
	}
}

func (m *machine) exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func (m *machine) handle(c run.Cmd) (string, error) {
	line := c.String()
	a := c.Args
	switch {
	// The fork.
	case line == "git -C "+m.cfg.Fork+" status --porcelain":
		return m.dirty, nil
	// sudo.
	case line == "sudo -n true":
		if m.sudoAsks {
			return "", errors.New("a password is required")
		}
		return "", nil
	case line == "sudo -v":
		m.sudoAsks = false
		return "", nil
	// Upstream's PKGBUILDs.
	case line == "git -C "+m.cfg.PKGBUILDs+" rev-parse HEAD":
		return m.pkgsHead + "\n", nil
	case line == "git -C "+m.cfg.PKGBUILDs+" pull --ff-only --quiet":
		if m.pkgsPulled != "" {
			m.pkgsHead = m.pkgsPulled
		}
		return "", nil
	case strings.HasPrefix(line, "git -C "+m.cfg.PKGBUILDs+" log ") && strings.HasSuffix(line, " -- pkgbuilds/omarchy-nvim"):
		return m.nvimLog, nil
	case strings.Contains(line, "log -r @ & master::"):
		if m.offBook {
			return "", nil
		}
		return commit, nil
	case strings.Contains(line, "log -r conflicts() & ::master"):
		if m.conflict && m.behind == 0 {
			return "zzz ours\n", nil
		}
		return "", nil
	case strings.Contains(line, "log -r ::quattro@upstream ~ ::master"):
		return strings.Repeat("x\n", m.behind), nil
	case strings.Contains(line, " rebase -b master -d quattro@upstream"):
		m.behind = 0
		return "", nil
	case strings.Contains(line, "log -r master -n 1"):
		return commit, nil
	case line == "git -C "+m.cfg.Fork+" rev-parse "+commit+"^{tree}":
		return "tree\n", nil
	case strings.HasSuffix(line, "rev-parse HEAD^{tree}"):
		return "pkgtree\n", nil

	// The signing key.
	case c.Name == "gpg":
		return "fpr:::::::::FPR:\n", nil

	// Files written as root, and the identity files read as root.
	case c.Sudo && c.Name == "mv":
		return "", os.Rename(a[1], a[2])
	case c.Name == "install" && a[0] == "-d":
		return "", os.MkdirAll(a[len(a)-1], 0o700)
	case c.Name == "install" && len(a) > 2:
		data, err := os.ReadFile(a[len(a)-2])
		if err != nil {
			return "", err
		}
		return "", os.WriteFile(a[len(a)-1], data, 0o644)
	case c.Name == "find":
		var files []string
		filepath.WalkDir(a[0], func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				if len(a) < 5 || func() bool { ok, _ := filepath.Match(a[4], d.Name()); return ok }() {
					files = append(files, p)
				}
			}
			return nil
		})
		if len(files) == 0 {
			return "", errors.New("no such file")
		}
		return strings.Join(files, "\n") + "\n", nil
	case c.Name == "cmp":
		x, _ := os.ReadFile(a[1])
		y, _ := os.ReadFile(a[2])
		if !bytes.Equal(x, y) {
			return "", errors.New("differ")
		}
	case c.Name == "test":
		if !m.exists(a[1]) {
			return "", errors.New("missing")
		}

	// pacman.
	case c.Name == "pacman" && !c.Sudo:
		return m.pacmanQuery(a)
	case c.Sudo && c.Name == "pacman":
		return "", m.pacmanChange(a)
	case c.Name == "yay":
		for _, p := range a[3:] {
			m.pkgs[p] = true
		}

	// systemd.
	case c.Name == "systemctl":
		scope := "system"
		if a[0] == "--user" {
			scope, a = "user", a[1:]
		}
		unit := a[len(a)-1]
		if a[0] == "is-enabled" && !m.enabled[scope+" "+unit] {
			return "", errors.New("disabled")
		}
		if a[0] == "enable" {
			m.enabled[scope+" "+unit] = true
		}

	// Omarchy.
	case line == "omarchy-update -y":
		if m.slowUpdate {
			time.Sleep(30 * time.Millisecond)
		}
		m.pendingDuringUpdate = m.exists(m.cfg.pending())
		if m.moveOmarchy {
			write(m.t, m.cfg.PacmanConf, stockConf)
		}
		if m.migrate != nil {
			m.migrate()
		}
		if m.stockOurs {
			m.versions["omarchy-dev"] = "4.0.0.r9999.gstock-1"
		}
	case line == "omarchy-update-restart --reboot-only":
		if m.reboots++; m.reboots == 1 {
			m.atReboot.pending = m.exists(m.cfg.pending())
			m.atReboot.gold = m.exists(m.cfg.gold())
		}
		return "", errors.New("declined")
	case line == "tmux info":
		return "", errors.New("no server")
	}
	return "", nil
}

func (m *machine) pacmanQuery(a []string) (string, error) {
	var out []string
	switch {
	case a[0] == "-Qq":
		for p := range m.pkgs {
			out = append(out, p)
		}
	case a[0] == "-Qqe":
		if m.failQqe {
			return "", errors.New("pacman crashed")
		}
		for p, explicit := range m.pkgs {
			if explicit {
				out = append(out, p)
			}
		}
	case a[0] == "-Qlq":
		out = m.starter
	case a[0] == "-Q":
		for _, p := range a[1:] {
			if _, ok := m.pkgs[p]; !ok {
				return "", errors.New("not installed: " + p)
			}
			v := m.versions[p]
			if v == "" {
				v = "1.0-1"
			}
			out = append(out, p+" "+v)
		}
	}
	return strings.Join(out, "\n") + "\n", nil
}

func (m *machine) pacmanChange(a []string) error {
	switch a[0] {
	case "-R":
		for _, p := range a[2:] {
			delete(m.pkgs, p)
			if p == "omarchy-nvim" {
				for _, f := range m.starter {
					if !strings.HasSuffix(f, "/") {
						os.Remove(f)
					}
				}
			}
		}
	case "-D":
		for _, p := range a[2:] {
			m.pkgs[p] = true
		}
	case "-S":
		for _, p := range a[1:] {
			if strings.HasPrefix(p, "-") || p == "4" {
				continue
			}
			if p == "onsen/omarchy-settings-dev" && m.editionSkel != nil {
				m.editionSkel()
				m.editionSkel = nil
			}
			if name, ok := strings.CutPrefix(p, "onsen/"); ok {
				delete(m.pkgs, strings.TrimSuffix(name, "-dev"))
				if !m.sticky {
					m.versions[name] = version + "-1"
				}
				p = name
			}
			m.pkgs[p] = true
		}
	}
	return nil
}

func (m *machine) update(check bool) (*Updater, error) {
	u := m.updater()
	return u, u.Update(context.Background(), check)
}

func out(u *Updater) string { return u.Out.(*bytes.Buffer).String() }
func errs(u *Updater) string {
	return u.Err.(*bytes.Buffer).String()
}

// before asserts that the commands ran in this order: each prefix matches a
// command after the one the previous prefix matched.
func before(t *testing.T, f *run.Fake, prefixes ...string) {
	t.Helper()
	lines := f.Lines()
	next := 0
	for _, p := range prefixes {
		i := next
		for i < len(lines) && !strings.HasPrefix(lines[i], p) {
			i++
		}
		if i == len(lines) {
			t.Errorf("%q did not run after the commands before it in the list\nran: %q", p, lines)
			return
		}
		next = i + 1
	}
}

func (m *machine) changes() []string {
	var out []string
	for _, line := range m.f.Lines() {
		if line == "sudo -n true" || line == "sudo -v" { // asking is not changing
			continue
		}
		if strings.HasPrefix(line, "sudo ") || strings.HasPrefix(line, "yay ") ||
			strings.HasPrefix(line, "omarchy-update") || strings.Contains(line, " git fetch") ||
			strings.Contains(line, " rebase ") || strings.HasPrefix(line, "systemctl --user enable") {
			out = append(out, line)
		}
	}
	return out
}

// A fresh install of stock Quattro: the stock packages, the LazyVim
// starter seeded into ~/.config/nvim, no record yet, and a declared package
// and service missing. The run takes the machine to the edition in v1's
// order, and a second run finds nothing to do.
func TestFreshInstallRunsEveryStepInOrder(t *testing.T) {
	m := newMachine(t)
	skel, home := m.cfg.Skel, m.cfg.Home
	if err := os.Remove(m.cfg.gold()); err != nil {
		t.Fatal(err)
	}
	delete(m.pkgs, "omarchy-dev")
	delete(m.pkgs, "omarchy-settings-dev")
	m.pkgs["omarchy"], m.pkgs["omarchy-settings"] = true, true
	m.pkgs["omarchy-nvim"] = true
	m.pkgs["neovim"] = false // brought by omarchy-nvim
	delete(m.pkgs, "nnn")
	delete(m.pkgs, "bat-extras")
	delete(m.enabled, "user syncthing.service")
	m.behind = 3
	m.moveOmarchy = true
	// ~/state came back from cumulus with the node key; the machine has
	// none yet.
	tailscale := m.cfg.Roots[0].Path
	write(t, m.cfg.Identity+tailscale, "node\n")
	// Stock /etc/skel, and the home the installer seeded from it. The
	// starter has a directory the edition does not ship (lua/extra).
	for _, rel := range []string{".config/nvim/lua/config/options.lua"} {
		os.Remove(filepath.Join(skel, rel))
		os.Remove(filepath.Join(home, rel))
	}
	stock := map[string]string{
		".config/nvim/lua/extra/x.lua":     "starter\n",
		".config/tmux/tmux.conf":           "stock\n",
		".config/nvim/init.lua":            "starter\n",
		".config/nvim/lua/config/lazy.lua": "starter\n",
	}
	for rel, body := range stock {
		write(t, filepath.Join(skel, rel), body)
		write(t, filepath.Join(home, rel), body)
	}
	starter := filepath.Join(skel, ".config/nvim")
	m.starter = []string{starter + "/", starter + "/lua/", starter + "/init.lua", starter + "/lua/config/lazy.lua", starter + "/lua/extra/x.lua"}
	m.editionSkel = func() {
		write(t, filepath.Join(skel, ".config/tmux/tmux.conf"), "edition\n")
		write(t, filepath.Join(skel, ".config/nvim/init.lua"), "-- ours\n")
		write(t, filepath.Join(skel, ".config/nvim/lua/config/options.lua"), "-- ours\n")
	}

	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	f := m.f
	before(t, f,
		"jj -R "+m.cfg.Fork+" git fetch --remote upstream",
		"jj -R "+m.cfg.Fork+" --ignore-immutable rebase -b master -d quattro@upstream",
		"git -C "+m.cfg.PKGBUILDs+" pull --ff-only --quiet",
		"sudo pacman-key --lsign-key FPR",
		"sudo pacman -R --noconfirm omarchy-nvim",
		"sudo pacman -S --needed --noconfirm --ask 4 onsen/omarchy-settings-dev onsen/omarchy-dev",
		"sudo pacman -D --asexplicit neovim",
		"omarchy-update -y",
		"sudo pacman -Sy",
		"sudo pacman -S --needed --noconfirm nnn",
		"yay -S --needed --noconfirm bat-extras",
		"install -m 600 "+m.cfg.Identity+tailscale+" "+tailscale,
		"systemctl --user enable --now syncthing.service",
		"omarchy-update-restart --reboot-only",
	)
	// The copied identity is live, and the copy still there.
	if body, _ := os.ReadFile(tailscale); string(body) != "node\n" {
		t.Errorf("tailscaled.state is %q after the restore", body)
	}
	if !m.exists(m.cfg.Identity + tailscale) {
		t.Error("the copy of tailscaled.state is gone")
	}
	// The [onsen] that the migration demoted is first again, and pacman
	// re-read it, after the update.
	if conf, _ := os.ReadFile(m.cfg.PacmanConf); strings.Index(string(conf), "[onsen]") > strings.Index(string(conf), "[omarchy]") {
		t.Errorf("[onsen] is not first after the update:\n%s", conf)
	}
	if sy := f.Matching("sudo pacman -Sy"); len(sy) != 2 {
		t.Errorf("pacman -Sy ran %d times; want once for our install, once after the reorder", len(sy))
	}
	if !m.pendingDuringUpdate {
		t.Error("no pending marker while omarchy update ran")
	}
	if m.reboots != 1 {
		t.Errorf("the reboot was offered %d times, want once", m.reboots)
	}
	if m.atReboot.pending || !m.atReboot.gold {
		t.Errorf("at the reboot prompt: pending %v, gold %v; want the run finished first", m.atReboot.pending, m.atReboot.gold)
	}
	if last := f.Lines()[len(f.Lines())-1]; last != "omarchy-update-restart --reboot-only" {
		t.Errorf("the reboot prompt is not last: %q", last)
	}
	for _, line := range f.Lines() {
		if line == "omarchy-update" {
			t.Error("omarchy update ran attended: its reboot prompt would race the rest")
		}
	}
	for rel, want := range map[string]string{".config/tmux/tmux.conf": "edition\n", ".config/nvim/init.lua": "-- ours\n"} {
		if got, _ := os.ReadFile(filepath.Join(home, rel)); string(got) != want {
			t.Errorf("~/%s is %q, want the edition's", rel, got)
		}
	}
	for _, rel := range []string{".config/nvim/lua/config/lazy.lua", ".config/nvim/lua/extra"} {
		if m.exists(filepath.Join(home, rel)) {
			t.Errorf("the starter's leftover ~/%s is still there", rel)
		}
	}
	if !strings.Contains(out(u), "==> Update complete: "+version+".") {
		t.Errorf("output:\n%s", out(u))
	}

	// The second run, as the README's recipe has it, finds no drift.
	m.f.Calls = nil
	if u, err := m.update(true); err != nil {
		t.Errorf("second run: %v\n%s%s", err, out(u), errs(u))
	}
}

// starterLeftovers removes a starter file only when it is unchanged and the
// edition does not ship that path.
func TestStarterLeftovers(t *testing.T) {
	m := newMachine(t)
	starter := filepath.Join(m.cfg.Skel, ".config/nvim")
	live := filepath.Join(m.cfg.Home, ".config/nvim")
	for rel, bodies := range map[string][2]string{
		"lua/config/lazy.lua":   {"starter\n", "starter\n"}, // leftover: goes
		"lua/config/edited.lua": {"starter\n", "mine\n"},    // edited: stays
		"init.lua":              {"starter\n", "starter\n"}, // the edition ships it: step 6 writes ours
	} {
		write(t, filepath.Join(starter, rel), bodies[0])
		write(t, filepath.Join(live, rel), bodies[1])
		m.starter = append(m.starter, filepath.Join(starter, rel))
	}
	if err := os.Symlink("../../theme.lua", filepath.Join(starter, "lua/config/theme.lua")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../theme.lua", filepath.Join(live, "lua/config/theme.lua")); err != nil {
		t.Fatal(err)
	}
	m.starter = append(m.starter, filepath.Join(starter, "lua/config/theme.lua"), starter+"/lua/")
	m.pkgs["omarchy-nvim"] = true
	u := m.updater()
	set, err := u.declared()
	if err != nil {
		t.Fatal(err)
	}
	if err := u.starterLeftovers(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	for rel, want := range map[string]bool{"lua/config/lazy.lua": false, "lua/config/theme.lua": false, "lua/config/edited.lua": true, "init.lua": true} {
		if got := m.exists(filepath.Join(live, rel)); got != want {
			t.Errorf("%s: exists %v, want %v", rel, got, want)
		}
	}
}

func TestDriftStopsBeforeAnythingChanges(t *testing.T) {
	m := newMachine(t)
	write(t, filepath.Join(m.cfg.Home, ".config/tmux/tmux.conf"), "by hand\n")
	m.pkgs["steam"] = true
	u, err := m.update(false)
	var code ExitCode
	if !errors.As(err, &code) || code != 1 {
		t.Fatalf("err %v, want exit status 1", err)
	}
	for _, want := range []string{"This machine differs from the edition. Nothing was changed.", "    changed ~/.config/tmux/tmux.conf", "    undeclared steam"} {
		if !strings.Contains(errs(u), want) {
			t.Errorf("missing %q in:\n%s", want, errs(u))
		}
	}
	if c := m.changes(); c != nil {
		t.Errorf("changed the machine: %q", c)
	}
}

// Review finding #5 on Hamr#762-3: a check that crashes stops the run with
// an error; it never reads as "no drift".
func TestCrashedCheckFails(t *testing.T) {
	m := newMachine(t)
	m.failQqe = true
	u, err := m.update(false)
	var code ExitCode
	if err == nil || errors.As(err, &code) || !strings.Contains(err.Error(), "the package check failed") {
		t.Fatalf("err %v", err)
	}
	if strings.Contains(out(u), "No drift") {
		t.Error("said no drift")
	}
	if c := m.changes(); c != nil {
		t.Errorf("changed the machine: %q", c)
	}

	m = newMachine(t)
	write(t, m.cfg.gold(), "corrupt\n")
	if _, err := m.update(false); err == nil || !strings.Contains(err.Error(), "the config check failed") {
		t.Fatalf("corrupt record: %v", err)
	}
}

// A run that died after omarchy update left a pending marker: the next run
// finishes its packages, services, and config before checking drift, so the
// file a migration changed is not taken for a hand edit.
func TestPendingRunIsFinishedFirst(t *testing.T) {
	m := newMachine(t)
	write(t, m.cfg.pending(), "")
	write(t, filepath.Join(m.cfg.Skel, ".config/tmux/tmux.conf"), "edition v2\n")
	write(t, filepath.Join(m.cfg.Home, ".config/tmux/tmux.conf"), "a migration wrote this\n")
	delete(m.pkgs, "nnn")
	delete(m.enabled, "user syncthing.service")
	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	o := out(u)
	finish := strings.Index(o, "Finishing its packages, config, and services first.")
	installs := strings.Index(o, "Installing the declared packages.")
	services := strings.Index(o, "Enabling the declared services.")
	copies := strings.Index(o, "Copying the edition's config")
	check := strings.Index(o, "Checking ~/.config")
	if finish < 0 || !(finish < installs && installs < copies && copies < services && services < check) {
		t.Errorf("not finished first:\n%s", o)
	}
	if f := m.f.Index("sudo pacman -S --needed --noconfirm nnn"); f < 0 || f > m.f.Index("jj -R") {
		t.Error("the declared package was not installed before the run went on")
	}
	if got, _ := os.ReadFile(filepath.Join(m.cfg.Home, ".config/tmux/tmux.conf")); string(got) != "edition v2\n" {
		t.Errorf("tmux.conf %q", got)
	}
}

func TestCheckChangesNothing(t *testing.T) {
	m := newMachine(t)
	write(t, m.cfg.pending(), "")
	m.behind = 2
	u, err := m.update(true)
	if err != nil {
		t.Fatalf("%v\n%s", err, out(u))
	}
	for _, want := range []string{
		"The last run stopped after omarchy update. Drift below may be its unfinished config copy.",
		"No drift.", "Fork working copy is clean.", "Working copy rides master.", "Published: " + version + ".",
	} {
		if !strings.Contains(out(u), want) {
			t.Errorf("missing %q in:\n%s", want, out(u))
		}
	}
	if strings.Contains(out(u), "Finishing") {
		t.Error("a check finished the pending run")
	}
	if c := m.changes(); c != nil {
		t.Errorf("changed the machine: %q", c)
	}
	if !m.exists(m.cfg.pending()) {
		t.Error("a check cleared the pending marker")
	}
}

func TestCheckFailsWhenUpdateWouldRefuse(t *testing.T) {
	for name, set := range map[string]func(*machine){
		"dirty":      func(m *machine) { m.dirty = " M README.md\n" },
		"off master": func(m *machine) { m.offBook = true },
	} {
		m := newMachine(t)
		set(m)
		u, err := m.update(true)
		var code ExitCode
		if !errors.As(err, &code) || code != 1 {
			t.Errorf("%s: err %v\n%s", name, err, out(u))
		}

		m = newMachine(t)
		set(m)
		if _, err := m.update(false); err == nil || errors.As(err, &code) {
			t.Errorf("%s: update ran: %v", name, err)
		}
		if c := m.changes(); c != nil {
			t.Errorf("%s: changed the machine: %q", name, c)
		}
	}
}

func TestRebaseConflictStopsBeforeBuilding(t *testing.T) {
	m := newMachine(t)
	m.behind = 1
	m.conflict = true
	if _, err := m.update(false); err == nil || !strings.Contains(err.Error(), "That is the review") {
		t.Fatalf("err %v", err)
	}
	for _, line := range m.f.Lines() {
		if (strings.HasPrefix(line, "sudo") && line != "sudo -n true" && line != "sudo -v") || strings.HasPrefix(line, "git -C "+m.cfg.PKGBUILDs) || strings.HasPrefix(line, "omarchy-update") {
			t.Errorf("went on after the conflict: %q", line)
		}
	}
}

// Review finding #2 on Hamr#762-3: if the update installed the stock
// packages over ours (from [omarchy], once a migration put it first), ours
// go back in from [onsen].
func TestOursComeBackAfterTheUpdate(t *testing.T) {
	m := newMachine(t)
	m.moveOmarchy, m.stockOurs = true, true
	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s", err, out(u))
	}
	before(t, m.f, "omarchy-update -y", "sudo pacman -Sy", "sudo pacman -S --noconfirm --ask 4 onsen/omarchy-settings-dev onsen/omarchy-dev")
}

func TestWrongVersionStopsBeforeConfig(t *testing.T) {
	m := newMachine(t)
	m.stockOurs, m.sticky = true, true
	gold, _ := os.ReadFile(m.cfg.gold())
	_, err := m.update(false)
	if err == nil || !strings.Contains(err.Error(), "not the published "+version+"-1") {
		t.Fatalf("err %v", err)
	}
	if now, _ := os.ReadFile(m.cfg.gold()); !bytes.Equal(now, gold) {
		t.Error("the config was copied anyway")
	}
	if !m.exists(m.cfg.pending()) {
		t.Error("the pending marker is gone, so the next run would not finish this one")
	}
	if m.f.Index("omarchy-update-restart") >= 0 {
		t.Error("offered the reboot after a failed run")
	}
}

func TestStarterLeftoversNeverThroughALink(t *testing.T) {
	m := newMachine(t)
	m.pkgs["omarchy-nvim"] = true
	starter := filepath.Join(m.cfg.Skel, ".config/nvim")
	write(t, filepath.Join(starter, "lua/config/lazy.lua"), "starter\n")
	checkout := filepath.Join(t.TempDir(), "nvim")
	write(t, filepath.Join(checkout, "lua/config/lazy.lua"), "starter\n")
	if err := os.RemoveAll(filepath.Join(m.cfg.Home, ".config/nvim")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(checkout, filepath.Join(m.cfg.Home, ".config/nvim")); err != nil {
		t.Fatal(err)
	}
	m.starter = []string{starter + "/lua/config/lazy.lua"}
	u := m.updater()
	set, err := u.declared()
	if err != nil {
		t.Fatal(err)
	}
	if err := u.starterLeftovers(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	if !m.exists(filepath.Join(checkout, "lua/config/lazy.lua")) {
		t.Error("removed a file through the link")
	}
}

func TestStarterLeftoversOnlyWhenRemovalDeclared(t *testing.T) {
	m := newMachine(t)
	write(t, m.cfg.onsen("packages"), "neovim\nnnn\naur:bat-extras\n")
	m.pkgs["omarchy-nvim"] = true
	starter := filepath.Join(m.cfg.Skel, ".config/nvim")
	write(t, filepath.Join(starter, "lua/config/lazy.lua"), "starter\n")
	write(t, filepath.Join(m.cfg.Home, ".config/nvim/lua/config/lazy.lua"), "starter\n")
	m.starter = []string{starter + "/lua/config/lazy.lua"}
	u := m.updater()
	set, err := u.declared()
	if err != nil {
		t.Fatal(err)
	}
	if err := u.starterLeftovers(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	if !m.exists(filepath.Join(m.cfg.Home, ".config/nvim/lua/config/lazy.lua")) {
		t.Error("removed the starter's files while the edition keeps the starter")
	}
}

func TestSelfUpdate(t *testing.T) {
	for _, c := range []struct {
		name, current, marker  string
		breakBuild, breakStart bool
		reexec, installs       bool
	}{
		{name: "hand-built", current: "", reexec: true, installs: true},
		{name: "current", current: "SAME"},
		{name: "already re-executed", current: "old", marker: "x"},
		{name: "build fails", current: "old", breakBuild: true},
		{name: "build does not start", current: "old", breakStart: true},
	} {
		m := newMachine(t)
		listing := "100644 blob aaa\tgo.mod\n040000 tree bbb\tcmd\n040000 tree ccc\tinternal\n"
		target := filepath.Join(t.TempDir(), "go/bin/onsen")
		write(t, target, "old build")
		var reexec []string
		self := &Self{Current: c.current, Reexec: func(path, extra string) error {
			reexec = []string{path, extra}
			return errors.New("stop here")
		}}
		if c.current == "SAME" {
			self.Current = sourceOf(listing)
		}
		inner := m.f.Handle
		m.f.Handle = func(cmd run.Cmd) (string, error) {
			switch {
			case strings.Contains(cmd.String(), "ls-tree "+commit+" -- go.mod go.sum cmd internal"):
				return listing, nil
			case cmd.Name == "go" && cmd.Args[0] == "build":
				if c.breakBuild {
					return "", errors.New("go: no network")
				}
				return "", os.WriteFile(cmd.Args[2], []byte("new build"), 0o755)
			case strings.HasSuffix(cmd.Name, "/onsen") && cmd.Args[0] == "--help":
				if c.breakStart {
					return "", errors.New("panic at startup")
				}
				return "usage", nil
			case cmd.String() == "go list -f {{.Target}} ./cmd/onsen":
				return target + "\n", nil
			}
			return inner(cmd)
		}
		u := m.updater()
		u.Self = self
		u.Getenv = func(k string) string {
			if k == reexecMarker {
				return c.marker
			}
			return ""
		}
		err := u.Update(context.Background(), false)
		got, _ := os.ReadFile(target)
		if c.installs != (string(got) == "new build") {
			t.Errorf("%s: installed build is %q", c.name, got)
		}
		if !c.reexec {
			if reexec != nil {
				t.Errorf("%s: re-executed %q", c.name, reexec)
			}
			broken := c.breakBuild || c.breakStart
			var code ExitCode
			switch {
			case broken && (!errors.As(err, &code) || !strings.Contains(out(u), "onsen could not install itself")):
				t.Errorf("%s: err %v; want the run to finish and report the problem\n%s", c.name, err, out(u))
			case broken && m.f.Index("omarchy-update -y") < 0:
				t.Errorf("%s: a failed self-install stopped the update", c.name)
			case !broken && err != nil:
				t.Errorf("%s: err %v", c.name, err)
			}
			continue
		}
		if err == nil || reexec == nil {
			t.Fatalf("%s: no re-exec (err %v)", c.name, err)
		}
		if reexec[0] != target || reexec[1] != reexecMarker+"="+sourceOf(listing) {
			t.Errorf("%s: re-exec %q", c.name, reexec)
		}
		before(t, m.f,
			"git -C "+m.cfg.Fork+" archive -o ",
			"tar -x -f ",
			"go build -o ",
			"go list -f {{.Target}} ./cmd/onsen")
		if b := m.f.Matching("go build"); len(b) != 1 || !strings.HasSuffix(b[0], "-ldflags -X github.com/lkingland/onsen/internal/update.Source="+sourceOf(listing)+" ./cmd/onsen") {
			t.Errorf("%s: build %q", c.name, b)
		}
		if arch := m.f.Matching("git -C " + m.cfg.Fork + " archive"); len(arch) != 1 || !strings.HasSuffix(arch[0], commit+" go.mod cmd internal") {
			t.Errorf("%s: archive %q", c.name, arch)
		}
		for _, line := range m.f.Matching("sudo") {
			if line != "sudo -n true" && line != "sudo -v" {
				t.Errorf("%s: changed the machine before re-executing: %q", c.name, line)
			}
		}
	}
}

// Hamr#762-8 review: a declared package that will not install, or a unit
// that will not start, must not hold every later update behind the pending
// marker. The run copies the config, clears the marker, offers the reboot,
// and exits non-zero naming the problem.
func TestBrokenDeclaredPackageDoesNotWedge(t *testing.T) {
	m := newMachine(t)
	delete(m.pkgs, "nnn")
	inner := m.f.Handle
	m.f.Handle = func(c run.Cmd) (string, error) {
		if c.String() == "sudo pacman -S --needed --noconfirm nnn" {
			return "", errors.New("target not found: nnn")
		}
		return inner(c)
	}
	u, err := m.update(false)
	var code ExitCode
	if !errors.As(err, &code) || code != 1 {
		t.Fatalf("err %v", err)
	}
	if !strings.Contains(out(u), "finished, with problems:") || !strings.Contains(out(u), "target not found: nnn") {
		t.Errorf("output:\n%s", out(u))
	}
	if m.exists(m.cfg.pending()) || m.reboots != 1 {
		t.Errorf("pending %v, reboots %d", m.exists(m.cfg.pending()), m.reboots)
	}
}

// A resumed run puts our packages back first: the config it copies must
// come from the build it published, not from whatever /etc/skel holds.
func TestResumeConfirmsOursBeforeCopyingConfig(t *testing.T) {
	m := newMachine(t)
	write(t, m.cfg.pending(), "")
	m.versions["omarchy-dev"] = "4.0.0.r9999.gstock-1"
	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s", err, out(u))
	}
	if r, f := m.f.Index("sudo pacman -S --noconfirm --ask 4 onsen/"), m.f.Index("jj -R"); r < 0 || r > f {
		t.Errorf("ours were not reinstalled before the run went on: %q", m.f.Lines())
	}

	m = newMachine(t)
	write(t, m.cfg.pending(), "")
	m.versions["omarchy-dev"] = "4.0.0.r9999.gstock-1"
	m.sticky = true
	gold, _ := os.ReadFile(m.cfg.gold())
	if _, err := m.update(false); err == nil {
		t.Fatal("copied config over the wrong packages")
	}
	if now, _ := os.ReadFile(m.cfg.gold()); !bytes.Equal(now, gold) || !m.exists(m.cfg.pending()) {
		t.Error("the config was copied, or the marker cleared, without our packages")
	}

	m = newMachine(t)
	write(t, m.cfg.pending(), "")
	if err := os.Remove(m.cfg.record()); err != nil {
		t.Fatal(err)
	}
	if _, err := m.update(false); err == nil || !strings.Contains(err.Error(), "nothing was published") {
		t.Errorf("err %v", err)
	}
}

func TestOneUpdateAtATime(t *testing.T) {
	m := newMachine(t)
	held := m.updater()
	unlock, err := held.lock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.update(false); err == nil || !strings.Contains(err.Error(), "another onsen update is running") {
		t.Errorf("err %v", err)
	}
	if c := m.changes(); c != nil {
		t.Errorf("changed the machine: %q", c)
	}
	if _, err := m.update(true); err != nil {
		t.Errorf("a check waits for no lock: %v", err)
	}
	unlock()
	if _, err := m.update(false); err != nil {
		t.Errorf("after unlock: %v", err)
	}
}

// On a fresh install the copy fills what is missing, and never writes over
// what is there: a host key sshd already made stops the run before the
// services and before the gold record, so the next run is still a first
// run and does not capture the fresh key over the copy (tex's kickback on
// Hamr#762-7).
func TestIdentityFirstRunRestoresOnlyWhatIsMissing(t *testing.T) {
	m := newMachine(t)
	if err := os.Remove(m.cfg.gold()); err != nil {
		t.Fatal(err)
	}
	ssh := m.cfg.Roots[1].Path
	write(t, ssh+"/ssh_host_ed25519_key", "new key\n")
	write(t, m.cfg.Identity+ssh+"/ssh_host_ed25519_key", "old key\n")
	write(t, m.cfg.Identity+ssh+"/ssh_host_rsa_key", "rsa\n")

	u, err := m.update(false)
	if err == nil || errors.Is(err, ExitCode(1)) || !strings.Contains(err.Error(), "already on the machine and differ") {
		t.Fatalf("err %v, want the host key conflict to stop the run\n%s", err, out(u))
	}
	if body, _ := os.ReadFile(ssh + "/ssh_host_ed25519_key"); string(body) != "new key\n" {
		t.Errorf("the live host key was replaced: %q", body)
	}
	if body, _ := os.ReadFile(m.cfg.Identity + ssh + "/ssh_host_ed25519_key"); string(body) != "old key\n" {
		t.Errorf("the copy was captured over: %q", body)
	}
	if body, _ := os.ReadFile(ssh + "/ssh_host_rsa_key"); string(body) != "rsa\n" {
		t.Errorf("the missing host key was not restored: %q", body)
	}
	if !strings.Contains(errs(u), ssh+"/ssh_host_ed25519_key") {
		t.Errorf("stderr does not name the file:\n%s", errs(u))
	}
	// The run stopped there: no services, no gold record, so the next run
	// is still a first run.
	if m.exists(m.cfg.gold()) {
		t.Error("the gold record was written; the next run would capture the fresh key over the copy")
	}
	if strings.Contains(out(u), "Enabling the declared services.") {
		t.Error("the services were enabled on the fresh files")
	}
	if m.reboots != 0 {
		t.Error("the reboot was offered")
	}

	// The operator removes the fresh key; the run goes through as a first
	// run and restores the copy.
	os.Remove(ssh + "/ssh_host_ed25519_key")
	m.f.Calls = nil
	if u, err := m.update(false); err != nil {
		t.Fatalf("second run: %v\n%s%s", err, out(u), errs(u))
	}
	if body, _ := os.ReadFile(ssh + "/ssh_host_ed25519_key"); string(body) != "old key\n" {
		t.Errorf("the host key after the re-run is %q, want the copy", body)
	}
	if !m.exists(m.cfg.gold()) {
		t.Error("the re-run did not finish")
	}
}

// On a machine already on the edition, live wins: a new pairing or a
// changed Syncthing config refreshes the copy, and a copy whose live file
// is gone is only reported.
func TestIdentityLaterRunCapturesLive(t *testing.T) {
	m := newMachine(t)
	tailscale := m.cfg.Roots[0].Path
	ssh := m.cfg.Roots[1].Path
	write(t, tailscale, "node\n")
	write(t, ssh+"/ssh_host_ed25519_key", "key v2\n")
	write(t, m.cfg.Identity+ssh+"/ssh_host_ed25519_key", "key v1\n")
	write(t, m.cfg.Identity+ssh+"/ssh_host_rsa_key", "gone live\n")

	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	for path, want := range map[string]string{
		m.cfg.Identity + tailscale:                     "node\n",
		m.cfg.Identity + ssh + "/ssh_host_ed25519_key": "key v2\n",
		m.cfg.Identity + ssh + "/ssh_host_rsa_key":     "gone live\n",
	} {
		if body, _ := os.ReadFile(path); string(body) != want {
			t.Errorf("%s is %q, want %q", path, body, want)
		}
	}
	if m.exists(ssh + "/ssh_host_rsa_key") {
		t.Error("a copy with no live file was restored onto a running machine")
	}
	if !strings.Contains(out(u), ssh+"/ssh_host_rsa_key has a copy under "+m.cfg.Identity+" but is not on the machine") {
		t.Errorf("output:\n%s", out(u))
	}
	if sudo := m.f.Matching("sudo install"); len(sudo) != 0 {
		t.Errorf("identity ran through sudo for roots that carry none: %q", sudo)
	}
}

// A pull of the PKGBUILDs that changes the Neovim overlay is shown, patch
// and all; one that does not is silent.
func TestUpstreamNeovimChangesShown(t *testing.T) {
	m := newMachine(t)
	m.pkgsHead, m.pkgsPulled = "old", "new"
	m.nvimLog = "commit new\n\n    Let the status line show the terminal background\n\n+-- nugget\n"
	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	if o := out(u); !strings.Contains(o, "Upstream changed pkgbuilds/omarchy-nvim") || !strings.Contains(o, "+-- nugget") {
		t.Errorf("the overlay change was not shown:\n%s", o)
	}
	if got := m.f.Matching("git -C " + m.cfg.PKGBUILDs + " log"); len(got) != 1 || !strings.Contains(got[0], "old..new -- pkgbuilds/omarchy-nvim") {
		t.Errorf("log %q", got)
	}

	m = newMachine(t)
	m.pkgsHead, m.nvimLog = "same", "+-- nugget\n"
	u, err = m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	if strings.Contains(out(u), "nugget") || len(m.f.Matching("git -C "+m.cfg.PKGBUILDs+" log")) != 0 {
		t.Errorf("an unchanged PKGBUILDs tree was reported:\n%s", out(u))
	}
}

// sudo is asked for once, up front, only when it would ask, and never by
// --check; the timestamp is refreshed while the run goes on.
func TestSudoAskedOnceUpFront(t *testing.T) {
	m := newMachine(t)
	m.sudoAsks = true
	m.slowUpdate = true
	u, err := m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	before(t, m.f, "sudo -n true", "sudo -v", "sudo pacman-key")
	if !strings.Contains(out(u), "sudo needs your password once") {
		t.Errorf("did not say it asks:\n%s", out(u))
	}

	m = newMachine(t)
	u, err = m.update(false)
	if err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	if len(m.f.Matching("sudo -v")) != 0 {
		t.Error("asked for a password sudo did not need")
	}

	m = newMachine(t)
	m.sudoAsks = true
	if _, err := m.update(true); err != nil {
		t.Fatal(err)
	}
	if len(m.f.Matching("sudo")) != 0 {
		t.Errorf("--check touched sudo: %q", m.f.Matching("sudo"))
	}
}

func TestSudoKeptFresh(t *testing.T) {
	m := newMachine(t)
	m.slowUpdate = true
	u := m.updater()
	u.Keepalive = 5 * time.Millisecond
	if err := u.Update(context.Background(), false); err != nil {
		t.Fatalf("%v\n%s%s", err, out(u), errs(u))
	}
	if n := len(m.f.Matching("sudo -n true")); n < 2 {
		t.Errorf("sudo refreshed %d time(s) during a slow run", n)
	}
}
