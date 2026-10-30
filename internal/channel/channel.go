// Package channel points pacman at the Onsen channel: the local [onsen]
// repository first, the fork's edge pacman.conf, and the signing key
// trusted.
package channel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Channel is pacman's configuration on this machine.
type Channel struct {
	Run        run.Runner
	Repo       string // the repository name: onsen
	Conf       string // /etc/pacman.conf
	Mirrorlist string // /etc/pacman.d/mirrorlist
	// EdgeConf and EdgeMirrorlist are the fork's edge templates
	// (default/pacman/pacman-edge.conf and mirrorlist-edge).
	EdgeConf, EdgeMirrorlist string
	Say                      func(string)
}

var section = regexp.MustCompile(`^[[:space:]]*\[([^]]+)\]`)

// Reorder returns conf with the repository section named repo moved in
// front of every other repository. Everything up to the first repository
// ([options] and the comments around it) stays where it is, and so does the
// order of the rest. A comment line belongs to the section above it.
func Reorder(conf, repo string) string {
	var prefix, ours, others strings.Builder
	inRepos := false
	name := ""
	lines := strings.Split(strings.TrimSuffix(conf, "\n"), "\n")
	if conf == "" {
		lines = nil
	}
	for _, line := range lines {
		if m := section.FindStringSubmatch(line); m != nil {
			name = m[1]
			if name != "options" {
				inRepos = true
			}
		}
		switch {
		case !inRepos:
			prefix.WriteString(line + "\n")
		case name == repo:
			ours.WriteString(line + "\n")
		default:
			others.WriteString(line + "\n")
		}
	}
	return prefix.String() + ours.String() + others.String()
}

// install puts content at dest as root, mode 0644: written beside dest, then
// renamed over it, so pacman.conf is never missing or half written.
func (c Channel) install(ctx context.Context, content, dest string) error {
	f, err := os.CreateTemp("", "onsen-pacman.")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	next := dest + ".onsen-new"
	if err := c.Run.Run(ctx, run.Root("install", "-m", "0644", "-o", "root", "-g", "root", f.Name(), next)); err != nil {
		return err
	}
	return c.Run.Run(ctx, run.Root("mv", "-f", next, dest))
}

// KeepFirst makes [Repo] the first repository in pacman.conf. pacman takes a
// package from the first repository that carries it, and [omarchy] carries
// packages with our two names. Upstream migration 1791403252 moves
// [omarchy] to the front, so the update calls this again afterwards. It
// reports whether it changed the file.
func (c Channel) KeepFirst(ctx context.Context) (bool, error) {
	data, err := os.ReadFile(c.Conf)
	if err != nil {
		return false, err
	}
	want := Reorder(string(data), c.Repo)
	if want == string(data) {
		return false, nil
	}
	c.Say(fmt.Sprintf("Putting [%s] back in front of the other repositories.", c.Repo))
	return true, c.install(ctx, want, c.Conf)
}

// Listed reports whether pacman.conf has the [Repo] section.
func (c Channel) Listed() (bool, error) {
	data, err := os.ReadFile(c.Conf)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "["+c.Repo+"]" {
			return true, nil
		}
	}
	return false, nil
}

// Switch writes the fork's edge pacman.conf and mirrorlist, once per
// machine: only when pacman.conf does not list [Repo] yet. The old files
// are kept as .bak, once: a later switch (after something dropped [Repo])
// keeps the first, pre-Onsen copy. The channel is built for x86 only.
func (c Channel) Switch(ctx context.Context) error {
	listed, err := c.Listed()
	if err != nil || listed {
		return err
	}
	platform, err := c.Run.Output(ctx, run.C("omarchy-hw-platform"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(platform) != "x86" {
		return fmt.Errorf("the Onsen channel is built for x86 only (this is %s)", strings.TrimSpace(platform))
	}
	c.Say("Switching pacman to the Onsen channel (the fork's edge template).")
	var cmds []run.Cmd
	for _, path := range []string{c.Conf, c.Mirrorlist} {
		if _, err := os.Lstat(path + ".bak"); err != nil {
			cmds = append(cmds, run.Root("cp", "-f", path, path+".bak"))
		}
	}
	cmds = append(cmds,
		run.Root("install", "-m", "0644", c.EdgeConf, c.Conf),
		run.Root("install", "-m", "0644", c.EdgeMirrorlist, c.Mirrorlist),
	)
	for _, cmd := range cmds {
		if err := c.Run.Run(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

// fingerprint looks up the primary fingerprint for key in the user keyring.
func (c Channel) fingerprint(ctx context.Context, key string) (string, error) {
	out, err := c.Run.Output(ctx, run.C("gpg", "--with-colons", "--list-keys", strings.TrimSuffix(key, "!")))
	if err != nil {
		return "", err
	}
	for _, line := range run.Lines(out) {
		if f := strings.Split(line, ":"); f[0] == "fpr" && len(f) > 9 {
			return f[9], nil
		}
	}
	return "", fmt.Errorf("no public key for %s in this keyring", key)
}

// pacmanKeyringGPG lists a key in pacman's keyring (--homedir /etc/pacman.d/gnupg).
func (c Channel) pacmanKeyringGPG(ctx context.Context, fpr string) run.Cmd {
	return run.Cmd{
		Name: "gpg", Sudo: true, Quiet: true,
		Args: []string{"--homedir", "/etc/pacman.d/gnupg", "--with-colons", "--list-keys", fpr},
	}
}

// KeyTrusted reports whether pacman's keyring carries key and has locally
// signed it (pub validity f or u). Listed-but-unsigned is not trusted
// (Hamr#768 kickback — an interrupted first run can leave the key listed
// without --lsign-key).
func (c Channel) KeyTrusted(ctx context.Context, key string) (bool, error) {
	fpr, err := c.fingerprint(ctx, key)
	if err != nil {
		return false, err
	}
	out, err := c.Run.Output(ctx, c.pacmanKeyringGPG(ctx, fpr))
	if err != nil {
		return false, nil // not in the keyring
	}
	for _, line := range run.Lines(out) {
		f := strings.Split(line, ":")
		if len(f) > 1 && f[0] == "pub" {
			return f[1] == "f" || f[1] == "u", nil
		}
	}
	return false, nil
}

// TrustKey adds the signing key to pacman's keyring when it is missing and
// signs it locally. listed+trusted → no-op (settled quiet); listed+untrusted
// → --lsign-key only; unlisted → add then lsign (Hamr#768).
func (c Channel) TrustKey(ctx context.Context, key string) error {
	fpr, err := c.fingerprint(ctx, key)
	if err != nil {
		return err
	}
	trusted, err := c.KeyTrusted(ctx, key)
	if err != nil {
		return err
	}
	if trusted {
		return nil
	}
	_, listErr := c.Run.Output(ctx, run.Cmd{Name: "pacman-key", Args: []string{"--list-keys", fpr}, Sudo: true, Quiet: true})
	if listErr != nil {
		c.Say("Adding the signing key to pacman's keyring.")
		armored, err := c.Run.Output(ctx, run.C("gpg", "--export", "--armor", fpr))
		if err != nil {
			return err
		}
		dir, err := os.MkdirTemp("", "onsen-key.")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		path := filepath.Join(dir, "key.asc")
		if err := os.WriteFile(path, []byte(armored), 0o644); err != nil {
			return err
		}
		if err := c.Run.Run(ctx, run.Root("pacman-key", "--add", path)); err != nil {
			return err
		}
	}
	_, err = c.Run.Output(ctx, run.Root("pacman-key", "--lsign-key", fpr))
	return err
}
