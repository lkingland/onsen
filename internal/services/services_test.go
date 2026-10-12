package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lkingland/onsen/internal/run"
)

func file(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "services")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRead(t *testing.T) {
	got, err := Read(file(t, "# declared\nsystem pcscd.socket   # the card\n\nuser syncthing.service\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Service{{Unit: "pcscd.socket"}, {User: true, Unit: "syncthing.service"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadRejectsBadLines(t *testing.T) {
	for _, body := range []string{"pcscd.socket\n", "global x.service\n", "user a.service b.service\n", "system\n"} {
		if _, err := Read(file(t, "user ok.service\n"+body)); err == nil {
			t.Errorf("accepted %q", body)
		}
	}
}

// Only units that are not enabled are enabled; a user unit as the user, a
// system unit through sudo. An enabled unit that is not running is started.
func TestEnable(t *testing.T) {
	enabled := map[string]bool{"syncthing.service": true, "ollama.service": true}
	active := map[string]bool{"syncthing.service": true}
	f := &run.Fake{Handle: func(c run.Cmd) (string, error) {
		if len(c.Args) >= 2 && c.Args[len(c.Args)-2] == "--quiet" {
			unit := c.Args[len(c.Args)-1]
			state := enabled
			if c.Args[len(c.Args)-3] == "is-active" {
				state = active
			}
			if !state[unit] {
				return "", errors.New("no")
			}
		}
		return "", nil
	}}
	svcs := []Service{{Unit: "pcscd.socket"}, {User: true, Unit: "syncthing.service"}, {User: true, Unit: "podman.socket"}, {Unit: "ollama.service"}}
	if err := Enable(context.Background(), f, svcs); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"systemctl --user daemon-reload",
		"systemctl is-enabled --quiet pcscd.socket",
		"sudo systemctl enable --now pcscd.socket",
		"systemctl --user is-enabled --quiet syncthing.service",
		"systemctl --user is-active --quiet syncthing.service",
		"systemctl --user is-enabled --quiet podman.socket",
		"systemctl --user enable --now podman.socket",
		"systemctl is-enabled --quiet ollama.service",
		"systemctl is-active --quiet ollama.service",
		"sudo systemctl start ollama.service",
	}
	if got := f.Lines(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// One unit that will not start does not keep the rest off; the error names
// each that failed.
func TestEnableFailureDoesNotStopTheRest(t *testing.T) {
	f := &run.Fake{Handle: func(c run.Cmd) (string, error) { return "", errors.New("no such unit " + c.Args[len(c.Args)-1]) }}
	err := Enable(context.Background(), f, []Service{{Unit: "a.service"}, {Unit: "b.service"}})
	if err == nil || !strings.Contains(err.Error(), "a.service") || !strings.Contains(err.Error(), "b.service") {
		t.Fatalf("err %v", err)
	}
	if got := f.Matching("sudo systemctl enable"); len(got) != 2 {
		t.Errorf("enabled %q, want both tried", got)
	}
}
