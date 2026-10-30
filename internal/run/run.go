// Package run starts the programs onsen drives (pacman, jj, systemctl and
// the rest) behind one small interface, so the tests can stand in for the
// machine.
package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Cmd is one program to start.
type Cmd struct {
	Name string
	Args []string
	// Dir is the working directory; empty means the current one.
	Dir string
	// Env is added to the inherited environment.
	Env []string
	// Sudo runs the program as root, through sudo.
	Sudo bool
	// Quiet discards the program's stderr (the shell's 2>/dev/null).
	Quiet bool
}

// C is shorthand for a Cmd with no options.
func C(name string, args ...string) Cmd { return Cmd{Name: name, Args: args} }

// Root is shorthand for a Cmd run through sudo.
func Root(name string, args ...string) Cmd { return Cmd{Name: name, Args: args, Sudo: true} }

// PacmanTx is a pacman transaction through omarchy-update-pacman, which runs
// pacman under a systemd scope so a mid-transaction systemd upgrade cannot
// SIGKILL it (Hamr#768). Queries (-Q*) stay plain pacman.
func PacmanTx(args ...string) Cmd { return C("omarchy-update-pacman", args...) }

// String is the command line as a shell would show it, with "sudo" in front
// when it runs as root. Tests match on it.
func (c Cmd) String() string {
	parts := make([]string, 0, len(c.Args)+2)
	if c.Sudo {
		parts = append(parts, "sudo")
	}
	parts = append(parts, c.Name)
	parts = append(parts, c.Args...)
	return strings.Join(parts, " ")
}

// Runner starts programs.
type Runner interface {
	// Run starts c on the terminal: stdin, stdout and stderr are this
	// process's own, so prompts (sudo, pinentry, pacman) reach the operator.
	Run(ctx context.Context, c Cmd) error
	// Output starts c and returns what it writes to stdout. Its stderr goes
	// to the terminal unless c is Quiet.
	Output(ctx context.Context, c Cmd) (string, error)
	// Has reports whether a program is on the PATH.
	Has(name string) bool
}

// Exec is the Runner for the real machine.
//
// A cancelled context stops the run between programs, never inside one:
// Ctrl-C reaches the running program from the terminal on its own, and it
// must be left to clean up (repo-add removes its lock, pacman finishes or
// rolls back its transaction). Killing it would leave that state behind.
type Exec struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer

	mu      sync.Mutex
	current *os.Process
}

// System returns the Runner for the real machine, on this process's own
// terminal.
func System() *Exec { return &Exec{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr} }

func (e *Exec) command(_ context.Context, c Cmd) *exec.Cmd {
	name, args := c.Name, c.Args
	if c.Sudo {
		name, args = "sudo", append([]string{c.Name}, c.Args...)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = c.Dir
	if len(c.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}
	cmd.Stdin = e.Stdin
	cmd.Stderr = e.Stderr
	if c.Quiet {
		cmd.Stderr = nil
	}
	return cmd
}

// start runs cmd to completion unless ctx is already done, keeping it as
// the program Signal reaches while it runs.
func (e *Exec) start(ctx context.Context, c Cmd, cmd *exec.Cmd) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("stopped before %s: %w", c, err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: %w", c, err)
	}
	e.mu.Lock()
	e.current = cmd.Process
	e.mu.Unlock()
	err := cmd.Wait()
	e.mu.Lock()
	e.current = nil
	e.mu.Unlock()
	if err != nil {
		return fmt.Errorf("%s: %w", c, err)
	}
	return nil
}

// Signal passes sig on to the program running now, if any. onsen uses it
// for a SIGTERM sent to onsen alone; a terminal's Ctrl-C already reaches the
// whole foreground process group.
func (e *Exec) Signal(sig os.Signal) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current != nil {
		_ = e.current.Signal(sig)
	}
}

// Run implements Runner.
func (e *Exec) Run(ctx context.Context, c Cmd) error {
	cmd := e.command(ctx, c)
	cmd.Stdout = e.Stdout
	return e.start(ctx, c, cmd)
}

// Output implements Runner.
func (e *Exec) Output(ctx context.Context, c Cmd) (string, error) {
	cmd := e.command(ctx, c)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := e.start(ctx, c, cmd)
	return out.String(), err
}

// Has implements Runner.
func (e *Exec) Has(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Lines splits a program's output into its non-empty lines.
func Lines(out string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
