package skel

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type box struct {
	t      *testing.T
	trees  Trees
	backup string
}

// newBox is a skeleton and a home that match it, recorded by a first Apply.
func newBox(t *testing.T) *box {
	t.Helper()
	root := t.TempDir()
	b := &box{t: t, trees: Trees{
		Skel: filepath.Join(root, "skel"),
		Home: filepath.Join(root, "home"),
		Gold: filepath.Join(root, "home/.local/state/omarchy/onsen-gold.sha256"),
		Dirs: []string{".config", ".gnupg", ".ssh"},
	}, backup: filepath.Join(root, "backup")}
	b.ship(".config/tmux/tmux.conf", "set -g prefix C-a\n")
	b.ship(".config/hypr/monitors.lua", "scale 1.25\n")
	b.ship(".gnupg/common.conf", "use-keyboxd\n")
	b.link(b.trees.Skel, ".config/nvim/lua/plugins/theme.lua", "../../../../.local/state/omarchy/current/theme/neovim.lua")
	if _, err := Apply(b.trees, b.backup); err != nil {
		t.Fatal(err)
	}
	return b
}

func (b *box) write(root, rel, body string) {
	b.t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		b.t.Fatal(err)
	}
}

func (b *box) link(root, rel, target string) {
	b.t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		b.t.Fatal(err)
	}
}

func (b *box) ship(rel, body string) { b.write(b.trees.Skel, rel, body) }
func (b *box) edit(rel, body string) { b.write(b.trees.Home, rel, body) }

func (b *box) read(rel string) string {
	b.t.Helper()
	data, err := os.ReadFile(filepath.Join(b.trees.Home, rel))
	if err != nil {
		b.t.Fatal(err)
	}
	return string(data)
}

func (b *box) check() []string {
	b.t.Helper()
	drift, err := Check(b.trees)
	if err != nil {
		b.t.Fatal(err)
	}
	return drift
}

func (b *box) apply() []string {
	b.t.Helper()
	report, err := Apply(b.trees, b.backup)
	if err != nil {
		b.t.Fatal(err)
	}
	return report
}

func same(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s:\n got %q\nwant %q", what, got, want)
	}
}

func TestFreshSyncIsClean(t *testing.T) {
	b := newBox(t)
	same(t, "drift", b.check(), nil)
	if got, _ := os.Readlink(filepath.Join(b.trees.Home, ".config/nvim/lua/plugins/theme.lua")); got != "../../../../.local/state/omarchy/current/theme/neovim.lua" {
		t.Errorf("theme.lua links to %q", got)
	}
}

func TestCheckNamesEachKindOfDrift(t *testing.T) {
	b := newBox(t)
	b.edit(".config/tmux/tmux.conf", "by hand\n")
	if err := os.Remove(filepath.Join(b.trees.Home, ".config/hypr/monitors.lua")); err != nil {
		t.Fatal(err)
	}
	b.edit(".config/tmux/extra.conf", "new\n")
	b.edit(".config/hypr/plugins/x.lua", "new dir\n")
	same(t, "drift", b.check(), []string{
		"missing ~/.config/hypr/monitors.lua",
		"changed ~/.config/tmux/tmux.conf",
		"new ~/.config/hypr/plugins/",
		"new ~/.config/tmux/extra.conf",
	})
}

// Files beside the trees' top (an app's own directory under ~/.config) are
// not watched: only directories that hold a managed file are.
func TestCheckWatchesOnlyManagedDirectories(t *testing.T) {
	b := newBox(t)
	b.edit(".config/someapp/settings.json", "{}\n")
	b.edit(".config/loose.conf", "x\n")
	same(t, "drift", b.check(), nil)
}

func TestCheckAcceptsWhatSkelShipsNow(t *testing.T) {
	b := newBox(t)
	// The update shipped a new tmux.conf and a migration copied it: not drift.
	b.ship(".config/tmux/tmux.conf", "upstream changed\n")
	b.edit(".config/tmux/tmux.conf", "upstream changed\n")
	same(t, "drift", b.check(), nil)
}

func TestCheckWithoutRecordComparesToSkel(t *testing.T) {
	b := newBox(t)
	if err := os.Remove(b.trees.Gold); err != nil {
		t.Fatal(err)
	}
	same(t, "clean", b.check(), nil)
	b.edit(".gnupg/common.conf", "by hand\n")
	same(t, "drift", b.check(), []string{"changed ~/.gnupg/common.conf"})
}

func TestIgnoredPathsAreNeverDrift(t *testing.T) {
	b := newBox(t)
	b.trees.Ignore = []string{".config/tmux/session-order", ".config/tmux/plugins/*", ".config/hypr/monitors.lua"}
	b.edit(".config/tmux/session-order", "pins\n")
	b.edit(".config/tmux/plugins/tpm/tpm", "x\n")
	b.edit(".config/hypr/monitors.lua", "app wrote this\n")
	same(t, "drift", b.check(), nil)
}

// "dir/" is how a directory is matched, and "*" crosses "/".
func TestIgnoreMatchesDirectoriesWithSlash(t *testing.T) {
	b := newBox(t)
	b.edit(".config/tmux/plugins/tpm/tpm", "x\n")
	same(t, "unignored", b.check(), []string{"new ~/.config/tmux/plugins/"})
	b.trees.Ignore = []string{".config/tmux/plugins/"}
	same(t, "ignored", b.check(), nil)
	b.trees.Ignore = []string{".config/*/plugins/"}
	same(t, "star", b.check(), nil)
}

func TestApplySeedsIgnoredFileOnlyWhenMissing(t *testing.T) {
	b := newBox(t)
	b.trees.Ignore = []string{".config/hypr/monitors.lua"}
	b.ship(".config/hypr/monitors.lua", "edition v2\n")
	b.edit(".config/hypr/monitors.lua", "the app's own\n")
	same(t, "report", b.apply(), nil)
	if got := b.read(".config/hypr/monitors.lua"); got != "the app's own\n" {
		t.Errorf("app-owned file overwritten: %q", got)
	}
	if err := os.Remove(filepath.Join(b.trees.Home, ".config/hypr/monitors.lua")); err != nil {
		t.Fatal(err)
	}
	same(t, "report", b.apply(), []string{"~/.config/hypr/monitors.lua"})
	if got := b.read(".config/hypr/monitors.lua"); got != "edition v2\n" {
		t.Errorf("seeded %q", got)
	}
	gold, _ := os.ReadFile(b.trees.Gold)
	if strings.Contains(string(gold), "monitors.lua") {
		t.Errorf("an ignored file is in the record:\n%s", gold)
	}
}

func TestApplyCopiesNewEditionAndRecordsIt(t *testing.T) {
	b := newBox(t)
	b.ship(".config/tmux/tmux.conf", "edition v2\n")
	same(t, "report", b.apply(), []string{"~/.config/tmux/tmux.conf"})
	if got := b.read(".config/tmux/tmux.conf"); got != "edition v2\n" {
		t.Errorf("copied %q", got)
	}
	same(t, "drift", b.check(), nil)
	if entries, err := os.ReadDir(b.backup); err == nil && len(entries) > 0 {
		t.Errorf("backed up a file that matched the record")
	}
}

// Review finding #4 on Hamr#762-3: a file that matches neither the record
// nor the edition is kept before it is replaced.
func TestApplyBacksUpFileChangedDuringTheRun(t *testing.T) {
	b := newBox(t)
	b.edit(".config/tmux/tmux.conf", "a migration wrote this\n")
	kept := filepath.Join(b.backup, ".config/tmux/tmux.conf")
	same(t, "report", b.apply(), []string{
		"~/.config/tmux/tmux.conf  (changed during this run) (old copy in " + kept + ")",
	})
	if got := b.read(".config/tmux/tmux.conf"); got != "set -g prefix C-a\n" {
		t.Errorf("the edition's copy did not win: %q", got)
	}
	data, err := os.ReadFile(kept)
	if err != nil || string(data) != "a migration wrote this\n" {
		t.Errorf("backup %q, %v", data, err)
	}
}

// Review finding #6 on Hamr#762-3: a file the edition stopped shipping goes
// when it is unchanged, and stays (as drift) when it was edited.
func TestApplyRemovesDroppedFilesOnlyWhenUnchanged(t *testing.T) {
	b := newBox(t)
	b.ship(".config/tmux/old.conf", "old\n")
	b.ship(".config/tmux/edited.conf", "old\n")
	b.apply()
	for _, rel := range []string{".config/tmux/old.conf", ".config/tmux/edited.conf"} {
		if err := os.Remove(filepath.Join(b.trees.Skel, rel)); err != nil {
			t.Fatal(err)
		}
	}
	b.edit(".config/tmux/edited.conf", "by hand\n")
	same(t, "report", b.apply(), []string{"~/.config/tmux/old.conf  (removed: no longer shipped)"})
	if _, err := os.Lstat(filepath.Join(b.trees.Home, ".config/tmux/old.conf")); err == nil {
		t.Error("unchanged dropped file is still there")
	}
	if got := b.read(".config/tmux/edited.conf"); got != "by hand\n" {
		t.Errorf("edited dropped file: %q", got)
	}
	same(t, "drift", b.check(), []string{"new ~/.config/tmux/edited.conf"})
}

func TestApplyKeepsGnupgAndSSHPrivate(t *testing.T) {
	b := newBox(t)
	b.ship(".gnupg/gpg-agent.conf", "pinentry-program x\n")
	b.ship(".ssh/config", "Host *\n")
	for _, dir := range []string{".gnupg", ".ssh"} {
		if err := os.MkdirAll(filepath.Join(b.trees.Home, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(b.trees.Home, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	b.apply()
	for rel, want := range map[string]os.FileMode{".gnupg": 0o700, ".gnupg/gpg-agent.conf": 0o600, ".ssh": 0o700, ".ssh/config": 0o600} {
		info, err := os.Stat(filepath.Join(b.trees.Home, rel))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s mode %o, want %o", rel, got, want)
		}
	}
}

func TestApplyReplacesLinkWithLink(t *testing.T) {
	b := newBox(t)
	rel := ".config/nvim/lua/plugins/theme.lua"
	if err := os.Remove(filepath.Join(b.trees.Home, rel)); err != nil {
		t.Fatal(err)
	}
	b.edit(rel, "a file, not the link\n")
	same(t, "drift", b.check(), []string{"changed ~/" + rel})
	b.apply()
	if got, err := os.Readlink(filepath.Join(b.trees.Home, rel)); err != nil || !strings.HasSuffix(got, "neovim.lua") {
		t.Errorf("theme.lua: %q, %v", got, err)
	}
}

func TestCorruptRecordFails(t *testing.T) {
	b := newBox(t)
	b.write(filepath.Dir(b.trees.Gold), filepath.Base(b.trees.Gold), "no separator here\n")
	if _, err := Check(b.trees); err == nil {
		t.Error("Check read a corrupt record")
	}
}

func TestReadIgnore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drift-ignore")
	if err := os.WriteFile(path, []byte("# why\n\n  .config/a/*  \n.config/b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadIgnore(path)
	if err != nil {
		t.Fatal(err)
	}
	same(t, "globs", got, []string{".config/a/*", ".config/b"})
	if got, err := ReadIgnore(path + ".missing"); err != nil || got != nil {
		t.Errorf("missing file: %q, %v", got, err)
	}
}

func TestMatch(t *testing.T) {
	for _, c := range []struct {
		pattern, name string
		want          bool
	}{
		{".config/nushell/history.*", ".config/nushell/history.sqlite3", true},
		{".config/obsidian/*", ".config/obsidian/a/b/c", true},
		{".config/obsidian/*", ".config/obsidian", false},
		{".config/tmux/plugins/*", ".config/tmux/plugins/", true},
		{".config/?mux/x", ".config/tmux/x", true},
		{".config/[!t]mux", ".config/tmux", false},
		{".config/[t]mux", ".config/tmux", true},
		{"a.b", "axb", false},
		{"[", "[", true},
	} {
		got, err := Match(c.pattern, c.name)
		if err != nil || got != c.want {
			t.Errorf("Match(%q, %q) = %v, %v; want %v", c.pattern, c.name, got, err, c.want)
		}
	}
}

// An empty skeleton (a missing package, a bad mount) must not read as a
// clean machine, nor let Apply remove every recorded file.
func TestEmptySkeletonFails(t *testing.T) {
	b := newBox(t)
	if err := os.RemoveAll(b.trees.Skel); err != nil {
		t.Fatal(err)
	}
	if _, err := Check(b.trees); err == nil {
		t.Error("Check passed an empty skeleton")
	}
	if _, err := Apply(b.trees, b.backup); err == nil {
		t.Error("Apply ran on an empty skeleton")
	}
	if got := b.read(".config/tmux/tmux.conf"); got != "set -g prefix C-a\n" {
		t.Errorf("tmux.conf %q", got)
	}
}

// A tree that is itself a link is followed, as os.walk followed it.
func TestLinkedTreeIsFollowed(t *testing.T) {
	b := newBox(t)
	real := filepath.Join(filepath.Dir(b.trees.Skel), "real-config")
	if err := os.Rename(filepath.Join(b.trees.Skel, ".config"), real); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(b.trees.Skel, ".config")); err != nil {
		t.Fatal(err)
	}
	same(t, "drift", b.check(), nil)
	same(t, "report", b.apply(), nil)
	if got := b.read(".config/tmux/tmux.conf"); got != "set -g prefix C-a\n" {
		t.Errorf("tmux.conf %q", got)
	}
}

func TestApplyKeepsModeBits(t *testing.T) {
	b := newBox(t)
	path := filepath.Join(b.trees.Skel, ".config/tmux/sessions-menu")
	b.ship(".config/tmux/sessions-menu", "#!/bin/sh\n")
	if err := os.Chmod(path, 0o755|os.ModeSetgid); err != nil {
		t.Fatal(err)
	}
	b.apply()
	info, err := os.Stat(filepath.Join(b.trees.Home, ".config/tmux/sessions-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode() & (os.ModePerm | os.ModeSetgid); got != 0o755|os.ModeSetgid {
		t.Errorf("mode %v", got)
	}
}

// Hamr#762-8 review: with no record yet, Check never looks at a path only
// the new edition ships, so Apply must keep whatever it replaces there.
func TestFirstRunKeepsWhatItReplaces(t *testing.T) {
	b := newBox(t)
	if err := os.Remove(b.trees.Gold); err != nil {
		t.Fatal(err)
	}
	b.edit(".config/jj/config.toml", "the operator's own\n")
	same(t, "drift", b.check(), nil) // the stock skeleton does not ship jj
	b.ship(".config/jj/config.toml", "edition\n")
	kept := filepath.Join(b.backup, ".config/jj/config.toml")
	same(t, "report", b.apply(), []string{"~/.config/jj/config.toml  (old copy in " + kept + ")"})
	data, err := os.ReadFile(kept)
	if err != nil || string(data) != "the operator's own\n" {
		t.Errorf("backup %q, %v", data, err)
	}
}

// A managed directory that is a link (~/.config/nvim pointing at a
// checkout) is drift, and nothing is read or written through it.
func TestLinkedDirectoryIsDriftAndNeverWritten(t *testing.T) {
	b := newBox(t)
	elsewhere := filepath.Join(filepath.Dir(b.trees.Home), "checkout")
	if err := os.Rename(filepath.Join(b.trees.Home, ".config/nvim"), elsewhere); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, filepath.Join(b.trees.Home, ".config/nvim")); err != nil {
		t.Fatal(err)
	}
	b.write(elsewhere, "lua/plugins/mine.lua", "x\n")
	same(t, "drift", b.check(), []string{"linked ~/.config/nvim"})
	b.ship(".config/nvim/init.lua", "edition\n")
	if _, err := Apply(b.trees, b.backup); err == nil {
		t.Fatal("Apply wrote through a linked directory")
	}
	if _, err := os.Lstat(filepath.Join(elsewhere, "init.lua")); err == nil {
		t.Error("wrote into the link's target")
	}
}

func TestSkeletonWithoutConfigFails(t *testing.T) {
	b := newBox(t)
	if err := os.RemoveAll(filepath.Join(b.trees.Skel, ".config")); err != nil {
		t.Fatal(err)
	}
	if _, err := Check(b.trees); err == nil {
		t.Error("Check passed a skeleton with .gnupg only")
	}
}

// A managed file edited to match what the fork's checkout ships is a change
// on its way in, not drift: committing drift clears it before the run that
// installs it. The same edit without the fork's copy is drift.
func TestCheckAcceptsForkCopy(t *testing.T) {
	b := newBox(t)
	b.edit(".config/tmux/tmux.conf", "set -g prefix C-b\n")
	b.edit(".gnupg/common.conf", "use-keyboxd\nno-autostart\n")
	fork := filepath.Join(filepath.Dir(b.trees.Home), "fork")
	b.write(fork, "config/tmux/tmux.conf", "set -g prefix C-b\n")
	b.write(fork, "etc/skel/.gnupg/common.conf", "use-keyboxd\nno-autostart\n")

	b.trees.Fork = fork
	got, err := Check(b.trees)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("with the fork's copies: got %q, want none", got)
	}

	b.trees.Fork = ""
	got, err = Check(b.trees)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"changed ~/.config/tmux/tmux.conf", "changed ~/.gnupg/common.conf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("without: got %q, want %q", got, want)
	}
}
