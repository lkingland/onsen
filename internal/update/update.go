// Package update brings an Onsen machine fully up to date: the fork's own
// commits, upstream Quattro, and every package, in one run.
//
//  1. Drift. Every config file /etc/skel ships into ~/.config, ~/.gnupg,
//     and ~/.ssh
//     must still match the copy recorded at the last run (or the copy the
//     installed /etc/skel ships now), and no file or directory may have
//     appeared beside them. Every explicitly installed package must be
//     declared, by Quattro's install lists or default/onsen/packages. A hand
//     edit or a hand install stops the run here, naming each, before
//     anything changes. Move it into the fork, or undo it, and run again.
//  2. Rebase the fork onto upstream Quattro. A conflict stops the run; the
//     conflict is the review. Nothing is built or applied. If the rebase
//     changed onsen's own source, onsen installs itself again and finishes
//     the run as the new build. Pull upstream's PKGBUILDs, and show what
//     changed under the paths this edition carries its own copy of
//     (Watched: the Neovim overlay), since nothing merges those.
//  3. Build omarchy-dev and omarchy-settings-dev from the fork with
//     upstream's PKGBUILDs, sign them with the YubiKey, and publish them to
//     the local [onsen] repository. Skipped when neither the fork's tree nor
//     the PKGBUILDs changed since the last publish.
//  4. Channel. Trust the signing key in pacman and, once per machine, write
//     the fork's edge pacman.conf. Keep [onsen] the first repository. Remove
//     the packages the edition takes away, swap in our packages, and mark
//     the declared ones explicit so the update's orphan sweep keeps them.
//  5. The stock `omarchy update`, unattended: snapshot, packages,
//     migrations, boot check. Then put [onsen] first again (an upstream
//     migration moves [omarchy] ahead), confirm the installed packages are
//     the build just published, install the declared packages that are
//     missing.
//  6. Copy the config files the new /etc/skel ships into ~/.config,
//     ~/.gnupg, and ~/.ssh, remove the ones the edition stopped shipping (if unchanged),
//     and record them as the copy step 1 checks next time. A file replaced
//     that matched neither the record nor the edition is kept under
//     ~/.local/state/omarchy/onsen-replaced/ first.
//     Then enable the declared services (default/onsen/services): after
//     the copy, so a unit the config ships is in place first.
//
// sudo is asked for once, up front, and kept fresh for the whole run, so a
// prompt never lands mid-step; with the edition's wheel rule in, never.
//
// Last, the stock reboot prompt, if the update needs one. omarchy update
// runs with -y, so it says "reboot when ready" instead of rebooting in the
// middle of the run. A pending marker written before step 5 makes the next
// run finish the rest of step 5 and step 6 first if a run dies after the
// update, so the files migrations touched are not mistaken for hand edits.
//
// Never omarchy-reinstall-configs: /etc/skel also carries a migration
// marker for every shipped migration and the Hyprland toggles. Replaying it
// marks migrations done without running them.
package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/lkingland/onsen/internal/build"
	"github.com/lkingland/onsen/internal/channel"
	"github.com/lkingland/onsen/internal/fork"
	"github.com/lkingland/onsen/internal/identity"
	"github.com/lkingland/onsen/internal/packages"
	"github.com/lkingland/onsen/internal/run"
	"github.com/lkingland/onsen/internal/services"
	"github.com/lkingland/onsen/internal/skel"
)

// Config is where everything lives.
type Config struct {
	Fork      string // the fork's checkout
	PKGBUILDs string // omacom-io/omarchy-pkgs
	Bookmark  string // master
	Remote    string // upstream
	Upstream  string // quattro@upstream
	Key       string // the signing key: 7DB461E762F282F9!
	RepoDir   string // /var/lib/omarchy-onsen
	RepoName  string // onsen
	Home      string
	User      string
	State     string // ~/.local/state/omarchy
	Skel      string // /etc/skel
	// PacmanConf and Mirrorlist are /etc/pacman.conf and
	// /etc/pacman.d/mirrorlist.
	PacmanConf, Mirrorlist string
	// QuattroInstall is the installed copy of Quattro's install lists:
	// /usr/share/omarchy/install.
	QuattroInstall string
	// Identity is where the machine's identity files are copied:
	// ~/state/identity. Roots is the set; nil means identity.Roots.
	Identity string
	Roots    []identity.Root
}

// Defaults is the configuration for this machine.
func Defaults(home, user string) Config {
	return Config{
		Fork:           filepath.Join(home, "src/kingland.io/onsen"),
		PKGBUILDs:      filepath.Join(home, "src/kingland.io/omarchy-pkgs"),
		Bookmark:       "master",
		Remote:         "upstream",
		Upstream:       "quattro@upstream",
		Key:            "7DB461E762F282F9!",
		RepoDir:        "/var/lib/omarchy-onsen",
		RepoName:       "onsen",
		Home:           home,
		User:           user,
		State:          filepath.Join(home, ".local/state/omarchy"),
		Skel:           "/etc/skel",
		PacmanConf:     "/etc/pacman.conf",
		Mirrorlist:     "/etc/pacman.d/mirrorlist",
		QuattroInstall: "/usr/share/omarchy/install",
		Identity:       filepath.Join(home, "state/identity"),
	}
}

func (c Config) pending() string { return filepath.Join(c.State, "onsen-pending") }
func (c Config) gold() string    { return filepath.Join(c.State, "onsen-gold.sha256") }
func (c Config) record() string  { return filepath.Join(c.State, "onsen-built") }
func (c Config) onsen(name string) string {
	return filepath.Join(c.Fork, "default/onsen", name)
}

// Updater runs an update.
type Updater struct {
	Config
	Run      run.Runner
	Out, Err io.Writer
	Now      func() time.Time
	Getenv   func(string) string
	// Self reinstalls onsen when the rebase changed its source; nil skips.
	Self *Self
	// Keepalive is how often the run refreshes sudo's timestamp; zero means
	// a minute.
	Keepalive time.Duration

	problems []string
}

// ExitCode ends the run with a status and no further message: what had to
// be said was said.
type ExitCode int

func (e ExitCode) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

func (u *Updater) say(format string, args ...any) {
	fmt.Fprintf(u.Out, "==> "+format+"\n", args...)
}

// list writes lines indented, as v1 did.
func list(w io.Writer, lines []string) {
	for _, l := range lines {
		fmt.Fprintln(w, "    "+l)
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func (u *Updater) trees() (skel.Trees, error) {
	ignore, err := skel.ReadIgnore(u.onsen("drift-ignore"))
	if err != nil {
		return skel.Trees{}, err
	}
	return skel.Trees{Skel: u.Skel, Home: u.Home, Gold: u.gold(), Dirs: []string{".config", ".gnupg", ".ssh"}, Ignore: ignore, Fork: u.Fork}, nil
}

func (u *Updater) declared() (*packages.Set, error) {
	return packages.Read(u.onsen("packages"), filepath.Join(u.Fork, "install"), u.QuattroInstall)
}

func (u *Updater) pacman() packages.Pacman { return packages.Pacman{Run: u.Run} }

func (u *Updater) fork() fork.Fork {
	return fork.Fork{Run: u.Run, Dir: u.Fork, Bookmark: u.Bookmark, Remote: u.Remote, Upstream: u.Upstream}
}

func (u *Updater) channel() channel.Channel {
	return channel.Channel{
		Run: u.Run, Repo: u.RepoName, Conf: u.PacmanConf, Mirrorlist: u.Mirrorlist,
		EdgeConf:       filepath.Join(u.Fork, "default/pacman/pacman-edge.conf"),
		EdgeMirrorlist: filepath.Join(u.Fork, "default/pacman/mirrorlist-edge"),
		Say:            func(s string) { u.say("%s", s) },
	}
}

func (u *Updater) build() build.Build {
	return build.Build{
		Run: u.Run, Fork: u.fork(), PKGBUILDs: u.PKGBUILDs, Packages: []string{"omarchy-settings-dev", "omarchy-dev"},
		Key: u.Key, RepoDir: u.RepoDir, RepoName: u.RepoName, Record: u.record(), User: u.User, Now: u.Now,
		Say: func(s string) { u.say("%s", s) },
	}
}

// ours is the edition's own packages, as pacman installs them from [onsen].
func (u *Updater) ours() []string {
	b := u.build()
	out := make([]string, len(b.Packages))
	for i, p := range b.Packages {
		out[i] = u.RepoName + "/" + p
	}
	return out
}

// Update runs the whole update, or with check only step 1 and the fork's
// state.
func (u *Updater) Update(ctx context.Context, check bool) error {
	if !exists(filepath.Join(u.Fork, ".jj")) {
		return fmt.Errorf("no jj repo at %s", u.Fork)
	}
	if !exists(filepath.Join(u.PKGBUILDs, "pkgbuilds/omarchy-dev/PKGBUILD")) {
		return fmt.Errorf("no upstream PKGBUILDs at %s (clone omacom-io/omarchy-pkgs there)", u.PKGBUILDs)
	}

	if !check {
		unlock, err := u.lock()
		if err != nil {
			return err
		}
		defer unlock()
		defer u.keepSudo(ctx)()
	}

	// --- Unfinished run ---
	// The last run died after omarchy update. Its config copy comes first,
	// or the files its migrations touched would read as hand edits; and
	// before that, our packages must be the build it published, or the
	// copy would record whatever /etc/skel holds now.
	if exists(u.pending()) {
		if check {
			u.say("The last run stopped after omarchy update. Drift below may be its unfinished config copy.")
		} else {
			u.say("The last run stopped after omarchy update. Finishing its packages, config, and services first.")
			_, published, err := u.build().Published()
			if err != nil {
				return err
			}
			if published == "" {
				return errors.New("a run stopped after omarchy update, but nothing was published. Remove " + u.pending() + " once the machine is checked")
			}
			if err := u.confirmOurs(ctx, published); err != nil {
				return err
			}
			if err := u.finish(ctx); err != nil {
				return err
			}
		}
	}

	// --- 1. Drift ---
	if err := u.drift(ctx); err != nil {
		return err
	}

	// --- Fork state ---
	f := u.fork()
	dirty, err := f.Dirty(ctx)
	if err != nil {
		return err
	}
	onBookmark := f.OnBookmark(ctx)
	if check {
		return u.report(dirty, onBookmark)
	}
	if len(dirty) > 0 {
		list(u.Err, dirty)
		return errors.New("fork working copy is dirty. Commit or park it first; a rebase would snapshot it")
	}
	if !onBookmark {
		return fmt.Errorf("fork working copy is not %s. jj new %s, then run again", u.Bookmark, u.Bookmark)
	}

	// --- 2. Rebase ---
	commit, err := u.rebase(ctx)
	if err != nil {
		return err
	}
	if u.Self != nil {
		if err := u.Self.Update(ctx, u, commit); err != nil {
			return err
		}
	}

	// --- 3. Build and publish ---
	version, err := u.build().Publish(ctx, commit)
	if err != nil {
		return err
	}

	// --- 4. Channel ---
	if err := u.channelStep(ctx); err != nil {
		return err
	}

	// --- 5. Stock update ---
	if err := os.MkdirAll(u.State, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(u.pending(), nil, 0o644); err != nil {
		return err
	}
	u.say("Running omarchy update.")
	if err := u.Run.Run(ctx, run.C("omarchy-update", "-y")); err != nil {
		return err
	}
	if err := u.confirmOurs(ctx, version); err != nil {
		return err
	}

	// --- 5, the rest, and 6. Config ---
	if err := u.finish(ctx); err != nil {
		return err
	}
	if len(u.problems) == 0 {
		u.say("Update complete: %s.", version)
	} else {
		u.say("Update to %s finished, with problems:", version)
		list(u.Out, u.problems)
	}

	// The update ran unattended, so a reboot it needs is offered here, last.
	_ = u.Run.Run(ctx, run.C("omarchy-update-restart", "--reboot-only"))
	if len(u.problems) > 0 {
		return ExitCode(1)
	}
	return nil
}

// problem records a failure that must not stop the run: the run finishes,
// says what failed, and exits non-zero.
func (u *Updater) problem(err error) {
	u.say("%v", err)
	u.problems = append(u.problems, err.Error())
}

// keepSudo asks for the password once, up front, when sudo would ask at
// all, and then refreshes sudo's timestamp every Keepalive until the run
// ends, so no prompt lands in the middle of a step and times out there.
// Once the edition's wheel rule is in, there is nothing to ask and the
// refreshes are free. It returns the function that stops the refreshes.
func (u *Updater) keepSudo(ctx context.Context) func() {
	fresh := func() error {
		_, err := u.Run.Output(ctx, run.Cmd{Name: "sudo", Args: []string{"-n", "true"}, Quiet: true})
		return err
	}
	if fresh() != nil {
		u.say("sudo needs your password once; the run keeps it fresh from here.")
		if err := u.Run.Run(ctx, run.C("sudo", "-v")); err != nil {
			u.say("Could not get sudo up front: %v. Each step will ask as it needs it.", err)
			return func() {}
		}
	}
	interval := u.Keepalive
	if interval == 0 {
		interval = time.Minute
	}
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = fresh()
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

// lock keeps two updates from running at once (one forgotten at a prompt,
// another started): both would rebase the fork and share the pending
// marker.
func (u *Updater) lock() (func(), error) {
	if err := os.MkdirAll(u.State, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(u.State, "onsen.lock"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, errors.New("another onsen update is running")
	}
	return func() { f.Close() }, nil
}

// drift is step 1. Each check is its own: a check that fails stops the run;
// it never reads as "no drift".
func (u *Updater) drift(ctx context.Context) error {
	u.say("Checking ~/.config, ~/.gnupg, ~/.ssh, and the installed packages against the edition.")
	trees, err := u.trees()
	if err != nil {
		return fmt.Errorf("the config check failed: %w. Nothing was changed", err)
	}
	drift, err := skel.Check(trees)
	if err != nil {
		return fmt.Errorf("the config check failed: %w. Nothing was changed", err)
	}
	set, err := u.declared()
	if err != nil {
		return fmt.Errorf("the package check failed: %w. Nothing was changed", err)
	}
	undeclared, err := u.pacman().Undeclared(ctx, set)
	if err != nil {
		return fmt.Errorf("the package check failed: %w. Nothing was changed", err)
	}
	for _, p := range undeclared {
		drift = append(drift, "undeclared "+p)
	}
	if len(drift) > 0 {
		fmt.Fprintln(u.Err, "==> This machine differs from the edition. Nothing was changed.")
		list(u.Err, drift)
		list(u.Err, []string{
			"Commit each change to the fork (a package goes in default/onsen/packages),",
			"or undo it, then run again.",
		})
		return ExitCode(1)
	}
	u.say("No drift.")
	return nil
}

// report is the fork's state for a check run; it fails when an update
// would refuse to start.
func (u *Updater) report(dirty []string, onBookmark bool) error {
	if len(dirty) == 0 {
		u.say("Fork working copy is clean.")
	} else {
		u.say("Fork working copy is dirty:")
		list(u.Out, dirty)
	}
	if onBookmark {
		u.say("Working copy rides %s.", u.Bookmark)
	} else {
		u.say("Working copy is not on %s.", u.Bookmark)
	}
	_, version, err := u.build().Published()
	switch {
	case err != nil:
		return err
	case version == "":
		u.say("Nothing published yet.")
	default:
		u.say("Published: %s.", version)
	}
	if len(dirty) > 0 || !onBookmark {
		return ExitCode(1)
	}
	return nil
}

// rebase is step 2. It returns the bookmark's commit.
func (u *Updater) rebase(ctx context.Context) (string, error) {
	f := u.fork()
	u.say("Fetching %s.", u.Remote)
	if err := f.Fetch(ctx); err != nil {
		return "", err
	}
	conflicts, err := f.Conflicts(ctx)
	if err != nil {
		return "", err
	}
	if len(conflicts) > 0 {
		list(u.Err, conflicts)
		return "", fmt.Errorf("%s has conflicts. Resolve them, then run again", u.Bookmark)
	}
	behind, err := f.Behind(ctx)
	if err != nil {
		return "", err
	}
	if behind > 0 {
		u.say("%d new upstream commit(s). Rebasing; the YubiKey re-signs each of ours.", behind)
		if err := f.Rebase(ctx); err != nil {
			return "", err
		}
		if conflicts, err = f.Conflicts(ctx); err != nil {
			return "", err
		}
		if len(conflicts) > 0 {
			list(u.Err, conflicts)
			return "", errors.New("the rebase conflicts. That is the review: resolve, then run again. Nothing was built or applied")
		}
	} else {
		u.say("Already on %s.", u.Upstream)
	}
	u.say("Updating upstream's PKGBUILDs.")
	before := u.pkgsHead(ctx)
	if err := u.Run.Run(ctx, run.C("git", "-C", u.PKGBUILDs, "pull", "--ff-only", "--quiet")); err != nil {
		return "", err
	}
	u.upstreamChanges(ctx, before)
	return f.Commit(ctx)
}

// Watched is what upstream keeps in its PKGBUILDs that this edition carries
// a copy of in its own tree. A change there never conflicts, so it is shown
// when a pull brings one, for the operator to copy over or leave. Today the
// Neovim overlay the stock omarchy-nvim package is made of (config/nvim here).
var Watched = []string{"pkgbuilds/omarchy-nvim"}

// pkgsHead is the PKGBUILDs' checked-out commit, or "" when that cannot be
// read.
func (u *Updater) pkgsHead(ctx context.Context) string {
	out, err := u.Run.Output(ctx, run.Cmd{Name: "git", Args: []string{"-C", u.PKGBUILDs, "rev-parse", "HEAD"}, Quiet: true})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// upstreamChanges prints each commit the pull brought that touches a
// watched path, patch included: the glance at upstream's ideas, with nothing
// to merge.
func (u *Updater) upstreamChanges(ctx context.Context, before string) {
	after := u.pkgsHead(ctx)
	if before == "" || after == "" || before == after {
		return
	}
	for _, path := range Watched {
		out, err := u.Run.Output(ctx, run.Cmd{Name: "git", Args: []string{"-C", u.PKGBUILDs, "log", "--no-color", "--reverse", "-p", before + ".." + after, "--", path}, Quiet: true})
		if err != nil || strings.TrimSpace(out) == "" {
			continue
		}
		u.say("Upstream changed %s since the last run. This edition carries its own copy; copy over what is worth it.", path)
		fmt.Fprintln(u.Out, strings.TrimRight(out, "\n"))
	}
}

// channelStep is step 4.
func (u *Updater) channelStep(ctx context.Context) error {
	change, err := u.wouldChange(ctx)
	if err != nil {
		return err
	}
	if change {
		// Snapshot before any root write so a pre-update snapper boot
		// undoes channel/package changes too (Hamr#768).
		u.say("Taking a snapshot before changing the channel and packages.")
		if err := u.Run.Run(ctx, run.Root("omarchy-snapshot", "create")); err != nil {
			return fmt.Errorf("snapshot before channel step: %w", err)
		}
	}
	ch := u.channel()
	if err := ch.TrustKey(ctx, u.Key); err != nil {
		return err
	}
	if err := ch.Switch(ctx); err != nil {
		return err
	}
	if _, err := ch.KeepFirst(ctx); err != nil {
		return err
	}
	set, err := u.declared()
	if err != nil {
		return err
	}
	if err := u.starterLeftovers(ctx, set); err != nil {
		return err
	}
	// The "-" lines go before our packages are installed, because one of
	// them may own a path ours ships.
	if err := u.pacman().Remove(ctx, set); err != nil {
		return err
	}
	b := u.build()
	if _, err := u.Run.Output(ctx, run.Cmd{Name: "pacman", Args: append([]string{"-Q"}, b.Packages...), Quiet: true}); err != nil {
		// A refresh and a targeted install is a partial upgrade for a
		// moment; the full omarchy update runs straight after. --ask 4
		// removes the stock omarchy and omarchy-settings these conflict with.
		u.say("Installing our packages in place of the stock ones.")
		if err := u.Run.Run(ctx, run.PacmanTx("-Sy")); err != nil {
			return err
		}
		args := append([]string{"-S", "--needed", "--noconfirm", "--ask", "4"}, u.ours()...)
		if err := u.Run.Run(ctx, run.PacmanTx(args...)); err != nil {
			return err
		}
	}
	// Before the update: its orphan sweep would otherwise take a declared
	// package that is only a dependency (neovim, once omarchy-nvim is gone).
	return u.pacman().Mark(ctx, set)
}

// wouldChange reports whether channelStep would write root state (Hamr#768).
// Read-only aside from the sudo pacman-key --list-keys probe inside KeyTrusted.
func (u *Updater) wouldChange(ctx context.Context) (bool, error) {
	ch := u.channel()
	trusted, err := ch.KeyTrusted(ctx, u.Key)
	if err != nil {
		return false, err
	}
	if !trusted {
		return true, nil
	}
	listed, err := ch.Listed()
	if err != nil {
		return false, err
	}
	if !listed {
		return true, nil
	}
	data, err := os.ReadFile(ch.Conf)
	if err != nil {
		return false, err
	}
	if channel.Reorder(string(data), ch.Repo) != string(data) {
		return true, nil
	}
	set, err := u.declared()
	if err != nil {
		return false, err
	}
	installed, err := u.pacman().Installed(ctx)
	if err != nil {
		return false, err
	}
	for _, n := range set.Remove {
		if installed[n] {
			return true, nil
		}
	}
	b := u.build()
	for _, pkg := range b.Packages {
		if !installed[pkg] {
			return true, nil
		}
	}
	explicit, err := u.pacman().Explicit(ctx)
	if err != nil {
		return false, err
	}
	for _, n := range append(append(append([]string{}, set.Repo...), set.AUR...), set.Local...) {
		if installed[n] && !explicit[n] {
			return true, nil
		}
	}
	return false, nil
}

// starterLeftovers clears the way for removing stock omarchy-nvim, the
// LazyVim starter, when the edition declares its removal: the edition ships
// its own ~/.config/nvim, and a fresh install seeded that directory from the
// starter. A starter file the edition does not ship, still unchanged, is the
// starter's leftover and goes before the package does, and so does a
// directory that leaves empty; step 6 writes ours over the rest. One that
// was edited stays, and the next drift check names it.
func (u *Updater) starterLeftovers(ctx context.Context, set *packages.Set) error {
	removing := false
	for _, p := range set.Remove {
		removing = removing || p == "omarchy-nvim"
	}
	// Nothing is removed through a link (~/.config/nvim pointing at a
	// checkout): the drift check names it instead.
	if !removing || skel.Linked(u.Home, ".config/nvim/x") != "" {
		return nil
	}
	if _, err := u.Run.Output(ctx, run.Cmd{Name: "pacman", Args: []string{"-Q", "omarchy-nvim"}, Quiet: true}); err != nil {
		return nil
	}
	out, err := u.Run.Output(ctx, run.C("pacman", "-Qlq", "omarchy-nvim"))
	if err != nil {
		return err
	}
	starter := filepath.Join(u.Skel, ".config/nvim") + "/"
	live := filepath.Join(u.Home, ".config/nvim")
	emptied := map[string]bool{}
	for _, src := range run.Lines(out) {
		rel, ok := strings.CutPrefix(src, starter)
		if !ok || rel == "" || strings.HasSuffix(rel, "/") {
			continue
		}
		if exists(filepath.Join(u.Fork, "config/nvim", rel)) {
			continue
		}
		path := filepath.Join(live, rel)
		if same, err := sameEntry(src, path); err != nil {
			return err
		} else if same {
			if err := os.Remove(path); err != nil {
				return err
			}
			for d := filepath.Dir(rel); d != "."; d = filepath.Dir(d) {
				emptied[d] = true
			}
		}
	}
	// Deepest first, so a parent is empty once its children are gone. A
	// directory that still holds something, or that the edition ships, stays.
	dirs := make([]string, 0, len(emptied))
	for d := range emptied {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, d := range dirs {
		if exists(filepath.Join(u.Fork, "config/nvim", d)) {
			continue
		}
		if entries, err := os.ReadDir(filepath.Join(live, d)); err == nil && len(entries) == 0 {
			if err := os.Remove(filepath.Join(live, d)); err != nil {
				return err
			}
		}
	}
	return nil
}

// sameEntry reports whether live is the same link as src, or a regular
// file with the same bytes as src.
func sameEntry(src, live string) (bool, error) {
	si, err := os.Lstat(src)
	if err != nil {
		return false, nil
	}
	li, err := os.Lstat(live)
	if err != nil {
		return false, nil
	}
	if si.Mode()&os.ModeSymlink != 0 {
		if li.Mode()&os.ModeSymlink == 0 {
			return false, nil
		}
		a, err1 := os.Readlink(src)
		b, err2 := os.Readlink(live)
		return err1 == nil && err2 == nil && a == b, nil
	}
	if !li.Mode().IsRegular() || !si.Mode().IsRegular() {
		return false, nil
	}
	a, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(live)
	if err != nil {
		return false, err
	}
	return string(a) == string(b), nil
}

// confirmOurs is the rest of step 5's package work: [onsen] first again,
// and our packages at the version just published.
func (u *Updater) confirmOurs(ctx context.Context, version string) error {
	changed, err := u.channel().KeepFirst(ctx)
	if err != nil {
		return err
	}
	if changed {
		if err := u.Run.Run(ctx, run.PacmanTx("-Sy")); err != nil {
			return err
		}
	}
	want := version + "-1"
	installed := func(pkg string) string {
		out, err := u.Run.Output(ctx, run.C("pacman", "-Q", pkg))
		if f := strings.Fields(out); err == nil && len(f) == 2 {
			return f[1]
		}
		return ""
	}
	b := u.build()
	for _, pkg := range b.Packages {
		if installed(pkg) != want {
			u.say("Reinstalling our packages from [%s].", u.RepoName)
			args := append([]string{"-S", "--noconfirm", "--ask", "4"}, u.ours()...)
			if err := u.Run.Run(ctx, run.PacmanTx(args...)); err != nil {
				return err
			}
			break
		}
	}
	for _, pkg := range b.Packages {
		if got := installed(pkg); got != want {
			return fmt.Errorf("%s is %q, not the published %s. Check pacman's output above; ~/.config was not copied", pkg, got, want)
		}
	}
	return nil
}

// finish is what follows omarchy update, and what a run that died after it
// left undone: the declared packages, the declared services, and the config
// copy.
//
// A declared package that will not install (an AUR build that broke, a
// package Arch dropped) or a unit that will not start is a problem, not a
// stop: the config copy still runs and clears the pending marker, so one
// broken package cannot hold every later update behind it.
func (u *Updater) finish(ctx context.Context) error {
	set, err := u.declared()
	if err != nil {
		return err
	}
	u.say("Installing the declared packages.")
	if err := u.pacman().Install(ctx, set); err != nil {
		u.problem(fmt.Errorf("declared packages: %w", err))
	}
	svcs, err := services.Read(u.onsen("services"))
	if err != nil {
		return err
	}
	if err := u.identity(ctx); err != nil {
		return err
	}
	// The config first: a unit the edition ships through ~/.config has to be
	// on disk before it can be enabled.
	if err := u.applyConfig(ctx); err != nil {
		return err
	}
	u.say("Enabling the declared services.")
	if err := services.Enable(ctx, u.Run, svcs); err != nil {
		u.problem(fmt.Errorf("declared services: %w", err))
	}
	return nil
}

// Identity is the Store for this machine.
// Diff is what onsen diff shows: every file the edition ships whose live
// copy differs, the drift-ignore list included, and the local files below
// an ignored path.
func (u *Updater) Diff() ([]skel.Entry, error) {
	trees, err := u.trees()
	if err != nil {
		return nil, err
	}
	return skel.Diff(trees)
}

func (u *Updater) Identity() identity.Store {
	roots := u.Roots
	if roots == nil {
		roots = identity.Roots(u.Home)
	}
	return identity.Store{Dir: u.Config.Identity, User: u.User, Roots: roots, Run: u.Run}
}

// identity keeps the machine's identity files (package identity) in step
// with their copies under ~/state, before the services that use them
// start. The live file wins. On a machine's first run (no gold record
// yet: a fresh install) a copy with no live file is restored, and a live
// file that differs from its copy stops the run, since one of the two is
// not this machine's and the next run would otherwise capture the fresh
// one over the copy. The stop comes before the services and before the
// gold record, so the next run is still a first run. On every later run
// the copies are refreshed from live, and a copy with no live file is only
// reported.
func (u *Updater) identity(ctx context.Context) error {
	st := u.Identity()
	files, err := st.Scan(ctx)
	if err != nil {
		return err
	}
	first := !exists(u.gold())
	var conflicts []string
	for _, f := range files {
		switch {
		case f.Live && f.Copy && f.Same:
		case !f.Live:
			if !first {
				u.say("%s has a copy under %s but is not on the machine; restore it by hand or remove the copy.", f.Path, u.Config.Identity)
				continue
			}
			u.say("Restoring %s.", f.Path)
			if err := st.Restore(ctx, f); err != nil {
				u.problem(fmt.Errorf("identity: %w", err))
			}
		case first && f.Copy:
			conflicts = append(conflicts, f.Path)
		case !first:
			u.say("Capturing %s.", f.Path)
			if err := st.Capture(ctx, f); err != nil {
				u.problem(fmt.Errorf("identity: %w", err))
			}
		}
	}
	if len(conflicts) > 0 {
		list(u.Err, conflicts)
		return fmt.Errorf("identity: these are already on the machine and differ from their copies under %s. Remove the one that is not this machine's and run again", u.Config.Identity)
	}
	return nil
}

// applyConfig is step 6, then the reloads that make the new config live.
// The reloads only warn: by then the config is copied and recorded.
func (u *Updater) applyConfig(ctx context.Context) error {
	u.say("Copying the edition's config into ~/.config, ~/.gnupg, and ~/.ssh.")
	trees, err := u.trees()
	if err != nil {
		return err
	}
	backup := filepath.Join(u.State, "onsen-replaced", u.Now().UTC().Format("20060102T150405Z"))
	report, err := skel.Apply(trees, backup)
	list(u.Out, report)
	if err != nil {
		return err
	}
	if err := os.Remove(u.pending()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	warn := func(what string, err error) {
		if err != nil {
			u.say("Could not %s: %v", what, err)
		}
	}
	if u.Run.Has("hyprctl") && u.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		_, err := u.Run.Output(ctx, run.C("hyprctl", "reload"))
		warn("reload Hyprland", err)
	}
	if _, err := u.Run.Output(ctx, run.Cmd{Name: "tmux", Args: []string{"info"}, Quiet: true}); err == nil {
		warn("reload tmux", u.Run.Run(ctx, run.C("tmux", "source-file", filepath.Join(u.Home, ".config/tmux/tmux.conf"))))
	}
	return nil
}
