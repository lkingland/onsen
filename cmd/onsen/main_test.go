package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// /dev/null is a character device but not a terminal: pinentry must not be
// pointed at it (an agent's shell has stdin on /dev/null).
func TestGPGTTYNeedsATerminal(t *testing.T) {
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer null.Close()
	t.Setenv("GPG_TTY", "")
	gpgTTY(null.Fd())
	if got := os.Getenv("GPG_TTY"); got != "" {
		t.Errorf("GPG_TTY=%q for /dev/null", got)
	}
}

func TestStatusPlaceholder(t *testing.T) {
	var out, errs bytes.Buffer
	if code := execute(context.Background(), nil, &out, &errs); code != 0 {
		t.Fatalf("exit %d: %s", code, errs.String())
	}
	if !strings.Contains(out.String(), "onsen update --check") {
		t.Errorf("output %q", out.String())
	}
	if code := execute(context.Background(), []string{"bogus"}, &out, &errs); code == 0 {
		t.Error("an unknown command succeeded")
	}
}

func TestSentence(t *testing.T) {
	if got := sentence("fork working copy is dirty"); got != "Fork working copy is dirty." {
		t.Errorf("got %q", got)
	}
}
