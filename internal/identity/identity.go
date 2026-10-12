// Package identity keeps the files that make this machine itself: the
// tailnet node key, the SSH host keys, the Syncthing device, the Bluetooth
// pairings, and the touch-to-sudo enrolment. None lives in a repository or
// a synced folder, so a rebuild would otherwise make each anew: a new
// tailnet address, a new host key on every peer, a re-pairing with every
// device.
//
// A copy of each file sits under ~/state/identity at its live path, and
// syncs with the rest of ~/state. The live file wins: a run refreshes the
// copy from it, and writes live from the copy only where live is absent,
// which is what a fresh install is. A pod set up new has nothing to restore
// and captures its own identity once it runs.
package identity

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Root is one place identity lives: a file, or a directory whose files
// (those matching Match, when set) are all kept.
type Root struct {
	Path  string
	Match string // a glob on the file name; empty keeps every file
	// Sudo reads and writes the live files as root.
	Sudo bool
	// Owner owns a restored file; empty is the user.
	Owner string
	// Mode is a restored file's mode. A *.pub file is 0644 regardless.
	Mode os.FileMode
}

// Roots is the set for this machine.
func Roots(home string) []Root {
	return []Root{
		{Path: "/var/lib/tailscale/tailscaled.state", Sudo: true, Owner: "root", Mode: 0o600},
		{Path: "/etc/ssh", Match: "ssh_host_*", Sudo: true, Owner: "root", Mode: 0o600},
		{Path: "/var/lib/bluetooth", Sudo: true, Owner: "root", Mode: 0o600},
		{Path: "/etc/fido2/fido2", Sudo: true, Owner: "root", Mode: 0o644},
		{Path: filepath.Join(home, ".local/state/syncthing/cert.pem"), Mode: 0o600},
		{Path: filepath.Join(home, ".local/state/syncthing/key.pem"), Mode: 0o600},
		{Path: filepath.Join(home, ".local/state/syncthing/config.xml"), Mode: 0o600},
	}
}

// Store is the copies under Dir.
type Store struct {
	Dir   string // ~/state/identity
	User  string
	Roots []Root
	Run   run.Runner
}

// File is one identity file, live and in the store.
type File struct {
	Root Root
	Path string // the live path
	// Live and Copy say which of the two exist; Same that both do and match.
	Live, Copy, Same bool
}

// CopyPath is where f's copy sits in the store.
func (s Store) CopyPath(f File) string { return filepath.Join(s.Dir, f.Path) }

// Scan lists every identity file, live or copied, in path order.
func (s Store) Scan(ctx context.Context) ([]File, error) {
	var out []File
	for _, r := range s.Roots {
		live, err := s.live(ctx, r)
		if err != nil {
			return nil, err
		}
		copies, err := s.copies(r)
		if err != nil {
			return nil, err
		}
		paths := map[string]bool{}
		for _, p := range live {
			paths[p] = true
		}
		for _, p := range copies {
			paths[p] = true
		}
		var sorted []string
		for p := range paths {
			sorted = append(sorted, p)
		}
		sort.Strings(sorted)
		for _, p := range sorted {
			f := File{Root: r, Path: p}
			for _, l := range live {
				if l == p {
					f.Live = true
				}
			}
			for _, c := range copies {
				if c == p {
					f.Copy = true
				}
			}
			if f.Live && f.Copy {
				f.Same = s.Run.Run(ctx, s.cmd(r, "cmp", "-s", p, s.CopyPath(f))) == nil
			}
			out = append(out, f)
		}
	}
	return out, nil
}

// live is the files under r on the machine. A root that is not there has
// none; a root that cannot be read is an error, never an empty list that a
// first run would then restore over.
func (s Store) live(ctx context.Context, r Root) ([]string, error) {
	if s.Run.Run(ctx, s.cmd(r, "test", "-e", r.Path)) != nil {
		return nil, nil
	}
	args := []string{r.Path, "-type", "f"}
	if r.Match != "" {
		args = append(args, "-name", r.Match)
	}
	out, err := s.Run.Output(ctx, s.cmd(r, "find", args...))
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", r.Path, err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// copies is the live paths of the files kept under r in the store.
func (s Store) copies(r Root) ([]string, error) {
	var files []string
	root := filepath.Join(s.Dir, r.Path)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && p == root {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if r.Match != "" {
			if ok, _ := filepath.Match(r.Match, d.Name()); !ok {
				return nil
			}
		}
		files = append(files, strings.TrimPrefix(p, s.Dir))
		return nil
	})
	return files, err
}

// Restore writes f's copy to its live path, with the root's owner and
// mode. A directory above it that is missing is made 0700; one that exists
// is left as it is (install -d would change its mode, and /etc/ssh is
// 0755 for a reason). It refuses when live exists.
func (s Store) Restore(ctx context.Context, f File) error {
	if f.Live {
		return fmt.Errorf("%s exists; not restoring over it", f.Path)
	}
	r := f.Root
	owner := []string{}
	if r.Owner != "" {
		owner = []string{"-o", r.Owner, "-g", r.Owner}
	}
	dir := filepath.Dir(f.Path)
	if s.Run.Run(ctx, s.cmd(r, "test", "-d", dir)) != nil {
		args := append([]string{"-d", "-m", "700"}, owner...)
		if err := s.Run.Run(ctx, s.cmd(r, "install", append(args, dir)...)); err != nil {
			return err
		}
	}
	mode := r.Mode
	if strings.HasSuffix(f.Path, ".pub") {
		mode = 0o644
	}
	args := append([]string{"-m", fmt.Sprintf("%o", mode)}, owner...)
	return s.Run.Run(ctx, s.cmd(r, "install", append(args, s.CopyPath(f), f.Path)...))
}

// Capture writes f's live file into the store, owned by the user, 0600.
func (s Store) Capture(ctx context.Context, f File) error {
	if !f.Live {
		return fmt.Errorf("%s is not there to capture", f.Path)
	}
	copy := s.CopyPath(f)
	// The directories are made here, never through sudo: root-owned
	// directories under ~/state would stop Syncthing writing there.
	if err := os.MkdirAll(filepath.Dir(copy), 0o700); err != nil {
		return err
	}
	args := []string{"-m", "600"}
	if f.Root.Sudo {
		args = append(args, "-o", s.User, "-g", s.User)
	}
	return s.Run.Run(ctx, s.cmd(f.Root, "install", append(args, f.Path, copy)...))
}

func (s Store) cmd(r Root, name string, args ...string) run.Cmd {
	return run.Cmd{Name: name, Args: args, Sudo: r.Sudo}
}
