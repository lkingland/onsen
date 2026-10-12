package identity

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lkingland/onsen/internal/run"
)

// The tests run the real find, cmp, test and install against a temporary
// machine, as the user: the roots carry no Sudo and no Owner.
func store(t *testing.T) (Store, string) {
	t.Helper()
	root := t.TempDir()
	machine := filepath.Join(root, "machine")
	s := Store{
		Dir:  filepath.Join(root, "state/identity"),
		User: "me",
		Roots: []Root{
			{Path: filepath.Join(machine, "var/lib/tailscale/tailscaled.state"), Mode: 0o600},
			{Path: filepath.Join(machine, "etc/ssh"), Match: "ssh_host_*", Mode: 0o600},
			{Path: filepath.Join(machine, "var/lib/bluetooth"), Mode: 0o600},
		},
		Run: &run.Exec{Stdout: io.Discard, Stderr: io.Discard},
	}
	return s, machine
}

func write(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func TestScanClassifiesEveryFile(t *testing.T) {
	s, m := store(t)
	ctx := context.Background()
	// Live only; both and same; both and different; copy only; a file in
	// /etc/ssh the match leaves alone; nothing at all under bluetooth.
	write(t, m+"/var/lib/tailscale/tailscaled.state", "node\n", 0o600)
	write(t, m+"/etc/ssh/ssh_host_ed25519_key", "key\n", 0o600)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_ed25519_key", "key\n", 0o600)
	write(t, m+"/etc/ssh/ssh_host_ed25519_key.pub", "new pub\n", 0o644)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_ed25519_key.pub", "old pub\n", 0o600)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_rsa_key", "rsa\n", 0o600)
	write(t, m+"/etc/ssh/sshd_config", "cfg\n", 0o644)

	files, err := s.Scan(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[strings.TrimPrefix(f.Path, m)] = state(f)
	}
	want := map[string]string{
		"/var/lib/tailscale/tailscaled.state": "live",
		"/etc/ssh/ssh_host_ed25519_key":       "same",
		"/etc/ssh/ssh_host_ed25519_key.pub":   "differ",
		"/etc/ssh/ssh_host_rsa_key":           "copy",
	}
	if len(got) != len(want) {
		t.Errorf("scanned %v, want %v", got, want)
	}
	for p, w := range want {
		if got[p] != w {
			t.Errorf("%s: %s, want %s", p, got[p], w)
		}
	}
}

func state(f File) string {
	switch {
	case f.Live && f.Copy && f.Same:
		return "same"
	case f.Live && f.Copy:
		return "differ"
	case f.Live:
		return "live"
	default:
		return "copy"
	}
}

func TestRestoreMakesMissingDirectoriesPrivateAndKeepsExistingOnes(t *testing.T) {
	s, m := store(t)
	ctx := context.Background()
	write(t, s.Dir+m+"/var/lib/tailscale/tailscaled.state", "node\n", 0o600)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_ed25519_key", "key\n", 0o600)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_ed25519_key.pub", "pub\n", 0o600)
	write(t, s.Dir+m+"/var/lib/bluetooth/AA/BB/info", "paired\n", 0o600)
	if err := os.MkdirAll(m+"/etc/ssh", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(m+"/var/lib/bluetooth/AA", 0o700); err != nil {
		t.Fatal(err)
	}

	files, err := s.Scan(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := s.Restore(ctx, f); err != nil {
			t.Fatal(err)
		}
	}
	for path, want := range map[string]os.FileMode{
		"/var/lib/tailscale":                  0o700 | os.ModeDir,
		"/var/lib/tailscale/tailscaled.state": 0o600,
		"/etc/ssh":                            0o755 | os.ModeDir,
		"/etc/ssh/ssh_host_ed25519_key":       0o600,
		"/etc/ssh/ssh_host_ed25519_key.pub":   0o644,
		"/var/lib/bluetooth/AA/BB":            0o700 | os.ModeDir,
		"/var/lib/bluetooth/AA/BB/info":       0o600,
	} {
		fi, err := os.Stat(m + path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		if fi.Mode() != want {
			t.Errorf("%s is %v, want %v", path, fi.Mode(), want)
		}
	}
	if body, _ := os.ReadFile(m + "/var/lib/bluetooth/AA/BB/info"); string(body) != "paired\n" {
		t.Errorf("info is %q", body)
	}

	// A live file is never written over.
	write(t, m+"/var/lib/tailscale/tailscaled.state", "live\n", 0o600)
	files, _ = s.Scan(ctx)
	if err := s.Restore(ctx, files[0]); err == nil {
		t.Error("restored over a live file")
	}
	if body, _ := os.ReadFile(m + "/var/lib/tailscale/tailscaled.state"); string(body) != "live\n" {
		t.Errorf("live file is %q after the refused restore", body)
	}
}

func TestCaptureCopiesLiveIntoThePrivateStore(t *testing.T) {
	s, m := store(t)
	ctx := context.Background()
	write(t, m+"/etc/ssh/ssh_host_ed25519_key.pub", "pub\n", 0o644)
	write(t, s.Dir+m+"/etc/ssh/ssh_host_ed25519_key.pub", "old\n", 0o600)
	write(t, m+"/var/lib/bluetooth/AA/BB/info", "paired\n", 0o600)

	files, err := s.Scan(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := s.Capture(ctx, f); err != nil {
			t.Fatal(err)
		}
	}
	for path, want := range map[string]string{
		"/etc/ssh/ssh_host_ed25519_key.pub": "pub\n",
		"/var/lib/bluetooth/AA/BB/info":     "paired\n",
	} {
		body, err := os.ReadFile(s.Dir + m + path)
		if err != nil || string(body) != want {
			t.Errorf("copy of %s: %q, %v", path, body, err)
		}
		fi, _ := os.Stat(s.Dir + m + path)
		if fi.Mode() != 0o600 {
			t.Errorf("copy of %s is %v, want 0600", path, fi.Mode())
		}
	}
	fi, _ := os.Stat(s.Dir + m + "/var/lib/bluetooth/AA/BB")
	if fi.Mode() != 0o700|os.ModeDir {
		t.Errorf("store directory is %v, want 0700", fi.Mode())
	}
	files, _ = s.Scan(ctx)
	for _, f := range files {
		if !f.Same {
			t.Errorf("%s is not the same after capture", f.Path)
		}
	}
}

// A root that is there but cannot be listed is an error: an empty list
// would read as "nothing live", and a first run would restore over it.
func TestScanFailsOnARootItCannotRead(t *testing.T) {
	s, m := store(t)
	ctx := context.Background()
	write(t, m+"/var/lib/bluetooth/AA/settings", "x\n", 0o600)
	if err := os.Chmod(m+"/var/lib/bluetooth/AA", 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(m+"/var/lib/bluetooth/AA", 0o700) })
	if _, err := s.Scan(ctx); err == nil {
		t.Error("Scan read an unreadable root as empty")
	}
	// Absent roots are still simply absent.
	os.Chmod(m+"/var/lib/bluetooth/AA", 0o700)
	os.RemoveAll(m + "/var/lib/bluetooth")
	files, err := s.Scan(ctx)
	if err != nil || len(files) != 0 {
		t.Errorf("absent roots: %v, %v", files, err)
	}
}
