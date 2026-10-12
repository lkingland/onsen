// Package services enables the declared services in default/onsen/services.
//
// Each line is "user <unit>" or "system <unit>"; "#" starts a comment.
// Services come after packages: a unit can only be enabled once the package
// that ships it is installed, which a migration (run inside omarchy update)
// cannot count on. They are enabled on every run, so a unit disabled by hand
// comes back.
package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Service is one declared unit.
type Service struct {
	User bool // a user unit (systemctl --user); otherwise a system unit
	Unit string
}

// Read reads a services file. Every line is checked before any is used, so
// a bad line stops the run before a unit is enabled.
func Read(path string) ([]Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []Service
	for n, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || (fields[0] != "user" && fields[0] != "system") {
			return nil, fmt.Errorf("%s:%d: %q is not \"user <unit>\" or \"system <unit>\"", path, n+1, strings.TrimSpace(line))
		}
		out = append(out, Service{User: fields[0] == "user", Unit: fields[1]})
	}
	return out, nil
}

// Enable enables and starts each service that is not enabled, and starts
// each enabled one that is not running (a unit whose start condition failed
// at login, or whose file the config copy just replaced, is enabled and
// dead). One that fails does not stop the rest; the error names every one
// that failed.
func Enable(ctx context.Context, r run.Runner, svcs []Service) error {
	var errs []error
	// A user unit the edition ships through ~/.config is new to the manager
	// until it reloads.
	for _, s := range svcs {
		if s.User {
			_ = r.Run(ctx, run.Cmd{Name: "systemctl", Args: []string{"--user", "daemon-reload"}, Quiet: true})
			break
		}
	}
	for _, s := range svcs {
		var isEnabled, isActive, enable, start run.Cmd
		if s.User {
			isEnabled = run.Cmd{Name: "systemctl", Args: []string{"--user", "is-enabled", "--quiet", s.Unit}, Quiet: true}
			isActive = run.Cmd{Name: "systemctl", Args: []string{"--user", "is-active", "--quiet", s.Unit}, Quiet: true}
			enable = run.C("systemctl", "--user", "enable", "--now", s.Unit)
			start = run.C("systemctl", "--user", "start", s.Unit)
		} else {
			isEnabled = run.Cmd{Name: "systemctl", Args: []string{"is-enabled", "--quiet", s.Unit}, Quiet: true}
			isActive = run.Cmd{Name: "systemctl", Args: []string{"is-active", "--quiet", s.Unit}, Quiet: true}
			enable = run.Root("systemctl", "enable", "--now", s.Unit)
			start = run.Root("systemctl", "start", s.Unit)
		}
		if err := r.Run(ctx, isEnabled); err == nil {
			if err := r.Run(ctx, isActive); err == nil {
				continue
			}
			if err := r.Run(ctx, start); err != nil {
				errs = append(errs, err)
			}
			continue
		}
		if err := r.Run(ctx, enable); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
