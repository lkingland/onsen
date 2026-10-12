package skel

import (
	"testing"
)

// Diff names a managed file that was edited, an ignored file the app
// rewrote, and a local file below an ignored directory, each with the paths
// a diff needs; a shipped file the home still matches is left out.
func TestDiff(t *testing.T) {
	b := newBox(t)
	b.ship(".config/omarchy/shell.json", `{"bar":{"position":"bottom"}}`+"\n")
	b.edit(".config/omarchy/shell.json", `{"bar":{"position":"top"}}`+"\n")
	b.edit(".config/omarchy/backgrounds/tokyo-night/0-black.png", "png")
	b.edit(".config/tmux/tmux.conf", "set -g prefix C-b\n")
	b.edit(".config/tmux/session-order", "work\n")
	b.trees.Ignore = []string{".config/omarchy/shell.json", ".config/omarchy/backgrounds/*", ".config/tmux/session-order"}

	got, err := Diff(b.trees)
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{
		{Rel: ".config/omarchy/backgrounds/tokyo-night/0-black.png", Live: b.trees.Home + "/.config/omarchy/backgrounds/tokyo-night/0-black.png", Ignored: true},
		{Rel: ".config/omarchy/shell.json", Skel: b.trees.Skel + "/.config/omarchy/shell.json", Live: b.trees.Home + "/.config/omarchy/shell.json", Ignored: true},
		{Rel: ".config/tmux/session-order", Live: b.trees.Home + "/.config/tmux/session-order", Ignored: true},
		{Rel: ".config/tmux/tmux.conf", Skel: b.trees.Skel + "/.config/tmux/tmux.conf", Live: b.trees.Home + "/.config/tmux/tmux.conf"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries %+v\nwant %d %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d:\n got %+v\nwant %+v", i, got[i], want[i])
		}
	}
}

// A file below an ignored directory that is itself watched (the edition
// ships a sibling there) is named once.
func TestDiffNamesOnce(t *testing.T) {
	b := newBox(t)
	b.ship(".config/fcitx5/conf/seed.conf", "seed\n")
	b.edit(".config/fcitx5/conf/seed.conf", "seed\n")
	b.edit(".config/fcitx5/conf/classicui.conf", "ui\n")
	b.trees.Ignore = []string{".config/fcitx5/*"}
	got, err := Diff(b.trees)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rel != ".config/fcitx5/conf/classicui.conf" {
		t.Fatalf("got %+v", got)
	}
}

// A shipped file missing from the home is an entry with no live path.
func TestDiffMissing(t *testing.T) {
	b := newBox(t)
	b.ship(".config/foot/foot.ini", "[main]\n")
	got, err := Diff(b.trees)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rel != ".config/foot/foot.ini" || got[0].Live != "" || got[0].Ignored {
		t.Fatalf("got %+v", got)
	}
}
