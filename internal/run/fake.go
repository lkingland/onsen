package run

import (
	"context"
	"strings"
	"sync"
)

// Fake is a Runner for tests. It records every command it is asked to start
// and answers each through Handle.
type Fake struct {
	// Calls is every command started, in order (Run and Output alike).
	Calls []Cmd
	// Handle answers a command with its stdout and error. Without one,
	// every command succeeds and prints nothing.
	Handle func(c Cmd) (string, error)
	// Programs are the programs Has finds.
	Programs map[string]bool

	mu sync.Mutex // Calls: the updater's sudo refresh runs on its own goroutine
}

// Run implements Runner.
func (f *Fake) Run(_ context.Context, c Cmd) error {
	_, err := f.answer(c)
	return err
}

// Output implements Runner.
func (f *Fake) Output(_ context.Context, c Cmd) (string, error) {
	return f.answer(c)
}

// Has implements Runner.
func (f *Fake) Has(name string) bool { return f.Programs[name] }

func (f *Fake) answer(c Cmd) (string, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, c)
	f.mu.Unlock()
	if f.Handle == nil {
		return "", nil
	}
	return f.Handle(c)
}

// Lines is every command started, as command lines.
func (f *Fake) Lines() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.Calls))
	for i, c := range f.Calls {
		out[i] = c.String()
	}
	return out
}

// Index is the position of the first command line that starts with prefix,
// or -1.
func (f *Fake) Index(prefix string) int {
	for i, line := range f.Lines() {
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}
	return -1
}

// Matching is every command line that starts with prefix.
func (f *Fake) Matching(prefix string) []string {
	var out []string
	for _, line := range f.Lines() {
		if strings.HasPrefix(line, prefix) {
			out = append(out, line)
		}
	}
	return out
}
