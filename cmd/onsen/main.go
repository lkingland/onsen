// onsen manages an Onsen machine: Omarchy Quattro plus the commits in this
// fork. Install it from the fork's checkout:
//
//	go install ./cmd/onsen
//
// After that, onsen update keeps it current with the fork on its own.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/spf13/cobra"

	"github.com/lkingland/onsen/internal/run"
	"github.com/lkingland/onsen/internal/update"
)

// system is the runner for the real machine, shared with the signal
// handler.
var system = run.System()

func main() {
	// Ctrl-C or SIGTERM stops the run once the program running now has
	// finished, cleanly or not; onsen itself does not die mid-program.
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			cancel()
			if sig == syscall.SIGTERM {
				system.Signal(sig)
			}
		}
	}()
	code := execute(ctx, os.Args[1:], os.Stdout, os.Stderr)
	signal.Stop(sigs)
	cancel()
	os.Exit(code)
}

// execute runs onsen with args and returns its exit status.
func execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := rootCmd(stdout)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(ctx)
	var code update.ExitCode
	switch {
	case err == nil:
		return 0
	case errors.As(err, &code):
		return int(code)
	default:
		fmt.Fprintln(stderr, "==> "+sentence(err.Error()))
		return 1
	}
}

// sentence is msg as onsen prints it: a capital first letter and a full stop.
func sentence(msg string) string {
	r, n := utf8.DecodeRuneInString(msg)
	msg = string(unicode.ToUpper(r)) + msg[n:]
	if !strings.HasSuffix(msg, ".") {
		msg += "."
	}
	return msg
}

func rootCmd(stdout io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   "onsen",
		Short: "Manage this Onsen machine",
		Long: `onsen manages an Onsen machine: Omarchy Quattro plus the commits in the
Onsen fork.

With no command it shows the machine's status.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(stdout, "onsen: the machine's status is not shown yet.")
			fmt.Fprintln(stdout, "  onsen update --check   check this machine against the edition")
			fmt.Fprintln(stdout, "  onsen update           bring it fully up to date")
			fmt.Fprintln(stdout, "  onsen identity         compare the machine's identity files with their copies")
			return nil
		},
	}
	root.AddCommand(updateCmd(stdout))
	root.AddCommand(identityCmd(stdout))
	root.AddCommand(diffCmd(stdout))
	return root
}

// defaults is the configuration for this machine, from flags.
func defaults(fork, pkgbuilds string) (update.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return update.Config{}, err
	}
	me, err := user.Current()
	if err != nil {
		return update.Config{}, err
	}
	cfg := update.Defaults(home, me.Username)
	if fork != "" {
		cfg.Fork = fork
	}
	if pkgbuilds != "" {
		cfg.PKGBUILDs = pkgbuilds
	}
	return cfg, nil
}

func identityCmd(stdout io.Writer) *cobra.Command {
	var capture bool
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Compare the machine's identity files with their copies",
		Long: `Compare the files that make this machine itself (the tailnet node key,
the SSH host keys, the Syncthing device, the Bluetooth pairings, the
touch-to-sudo enrolment) with their copies under ~/state/identity, and
list each that differs or is on one side only.

onsen update keeps the two in step on every run, with the live file
winning. --capture does the live-to-copy half now, for a backup before a
rebuild.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := defaults("", "")
			if err != nil {
				return err
			}
			u := &update.Updater{Config: cfg, Run: system, Out: stdout, Err: cmd.ErrOrStderr()}
			st := u.Identity()
			files, err := st.Scan(cmd.Context())
			if err != nil {
				return err
			}
			same := 0
			for _, f := range files {
				switch {
				case f.Same:
					same++
				case !f.Live:
					fmt.Fprintf(stdout, "copy only  %s\n", f.Path)
				case !f.Copy:
					fmt.Fprintf(stdout, "live only  %s\n", f.Path)
				default:
					fmt.Fprintf(stdout, "differ     %s\n", f.Path)
				}
				if capture && f.Live && !f.Same {
					if err := st.Capture(cmd.Context(), f); err != nil {
						return err
					}
					fmt.Fprintf(stdout, "captured   %s\n", f.Path)
				}
			}
			fmt.Fprintf(stdout, "%d of %d the same\n", same, len(files))
			return nil
		},
	}
	cmd.Flags().BoolVar(&capture, "capture", false, "refresh the copies from the live files")
	return cmd
}

func diffCmd(stdout io.Writer) *cobra.Command {
	var ignoredOnly, names bool
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show how the live config differs from what the edition ships",
		Long: `Show, as unified diffs, every file the edition ships into ~/.config,
~/.gnupg, and ~/.ssh whose live copy differs from the edition's, and name
every local file below a path on default/onsen/drift-ignore.

The drift check (onsen update --check) stops on a managed file that was
edited, but never looks at an ignored one: those are files an app rewrites
itself, so they drift silently and are lost at a rebuild. This shows that
drift too, so a change worth keeping can be carried into the fork by hand.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := defaults("", "")
			if err != nil {
				return err
			}
			u := &update.Updater{Config: cfg, Run: system, Out: stdout, Err: cmd.ErrOrStderr()}
			entries, err := u.Diff()
			if err != nil {
				return err
			}
			shown := 0
			for _, e := range entries {
				if ignoredOnly && !e.Ignored {
					continue
				}
				shown++
				var kind string
				switch {
				case e.Skel == "":
					kind = "local, under an ignored path; the edition ships nothing here"
				case e.Live == "":
					kind = "missing; the edition ships it"
				case e.Ignored:
					kind = "ignored; its app rewrites it, so onsen update never stops on it"
				default:
					kind = "managed; onsen update stops on it until the fork carries or reverts it"
				}
				fmt.Fprintf(stdout, "==> ~/%s (%s)\n", e.Rel, kind)
				if names || e.Skel == "" || e.Live == "" {
					continue
				}
				c := exec.CommandContext(cmd.Context(), "diff", "-u", "--label", "edition/"+e.Rel, "--label", "live/"+e.Rel, e.Skel, e.Live)
				c.Stdout = stdout
				c.Stderr = cmd.ErrOrStderr()
				if err := c.Run(); err != nil {
					var exit *exec.ExitError
					if !errors.As(err, &exit) || exit.ExitCode() > 1 {
						return fmt.Errorf("diff ~/%s: %w", e.Rel, err)
					}
				}
			}
			if shown == 0 {
				fmt.Fprintln(stdout, "The live config is as the edition ships it.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&ignoredOnly, "ignored", false, "only the files drift-ignore covers")
	cmd.Flags().BoolVar(&names, "names", false, "name the files without the diffs")
	return cmd
}

func updateCmd(stdout io.Writer) *cobra.Command {
	var check bool
	var fork, pkgbuilds string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Bring this machine fully up to date",
		Long: `Bring this machine fully up to date from every source: the fork's own
commits, upstream Omarchy Quattro, and every package.

  1. Drift: stop, changing nothing, if config the edition ships was edited
     by hand or a package was installed that the edition does not declare.
  2. Rebase the fork onto upstream Quattro; a conflict stops the run.
     If that changed onsen itself, install onsen again and carry on as the
     new build. Pull upstream's PKGBUILDs and show what changed under
     pkgbuilds/omarchy-nvim, the Neovim config this edition carries a
     copy of.
  3. Build, sign, and publish the edition's packages to [onsen].
  4. Point pacman at [onsen] first; remove what the edition takes away.
  5. Run omarchy update unattended, then install the declared packages,
     bring the machine's identity files and their copies under ~/state
     into step (onsen identity --help).
  6. Copy the edition's config into ~/.config, ~/.gnupg, and ~/.ssh, then
     enable the declared services, so a unit the config ships is in place.
Last, offer the reboot the update needs, if any. sudo is asked for once,
up front, and kept fresh for the whole run.

--check runs step 1 only and reports the fork's state.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := defaults(fork, pkgbuilds)
			if err != nil {
				return err
			}
			gpgTTY(0)
			u := &update.Updater{
				Config: cfg,
				Run:    system,
				Out:    stdout,
				Err:    cmd.ErrOrStderr(),
				Now:    time.Now,
				Getenv: os.Getenv,
				Self: &update.Self{
					Current: update.Source,
					Reexec: func(path, extra string) error {
						return syscall.Exec(path, os.Args, append(os.Environ(), extra))
					},
				},
			}
			return u.Update(cmd.Context(), check)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "check for drift and report the fork's state; change nothing")
	cmd.Flags().StringVar(&fork, "fork", "", "the fork's checkout (default ~/src/kingland.io/onsen)")
	cmd.Flags().StringVar(&pkgbuilds, "pkgbuilds", "", "upstream's omarchy-pkgs checkout (default ~/src/kingland.io/omarchy-pkgs)")
	return cmd
}

// gpgTTY gives pinentry the terminal on fd when it is one: jj and makepkg
// sign with the card.
func gpgTTY(fd uintptr) {
	if os.Getenv("GPG_TTY") != "" {
		return
	}
	// A terminal, not just any character device: /dev/null is one too.
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return
	}
	if tty, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", fd)); err == nil {
		os.Setenv("GPG_TTY", tty)
	}
}
