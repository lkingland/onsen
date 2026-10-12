package run

import (
	"bytes"
	"context"
	"io"
	"strings"
	"syscall"
	"testing"
	"time"
)

func quiet() *Exec {
	return &Exec{Stdin: strings.NewReader(""), Stdout: io.Discard, Stderr: io.Discard}
}

// Ctrl-C reaches the running program from the terminal; onsen must let it
// finish its cleanup rather than kill it (repo-add's lock, pacman's
// transaction).
func TestCancelLetsTheRunningProgramFinish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	out, err := quiet().Output(ctx, C("sh", "-c", "sleep 0.3; echo finished"))
	if err != nil || strings.TrimSpace(out) != "finished" {
		t.Fatalf("out %q, err %v; the program was cut off", out, err)
	}
}

func TestCancelStartsNothingMore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout bytes.Buffer
	e := quiet()
	e.Stdout = &stdout
	if err := e.Run(ctx, C("echo", "ran")); err == nil || stdout.Len() != 0 {
		t.Fatalf("started after cancel: err %v, out %q", err, stdout.String())
	}
}

func TestSignalReachesTheRunningProgram(t *testing.T) {
	e := quiet()
	time.AfterFunc(100*time.Millisecond, func() { e.Signal(syscall.SIGTERM) })
	out, err := e.Output(context.Background(), C("sh", "-c", "trap 'echo cleaned; exit 3' TERM; sleep 5 >/dev/null 2>&1 & wait"))
	if err == nil || strings.TrimSpace(out) != "cleaned" {
		t.Fatalf("out %q, err %v", out, err)
	}
	e.Signal(syscall.SIGTERM) // nothing running: a no-op
}

func TestCmdString(t *testing.T) {
	if got := Root("pacman", "-Sy").String(); got != "sudo pacman -Sy" {
		t.Errorf("got %q", got)
	}
}
