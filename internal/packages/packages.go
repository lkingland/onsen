// Package packages keeps this machine's packages to the declared set in
// default/onsen/packages.
//
// The declared set is that list plus the edition's own two packages. A line
// is a repository package, "aur:<name>" for one built from the AUR with yay,
// "local:<name>" for one built and installed by hand from its own checkout
// (declared, so it is not drift; never installed from here, and named when
// it is missing), or "-<name>" for one the edition takes away. Quattro's install lists
// (omarchy-base, omarchy-x86_64-only, and omarchy-other, which carries the
// installer's base system and the drivers it picks per machine) are what the
// installer chose from: allowed, never installed from here, because most of
// their driver lines are for other hardware.
//
// Quattro's lists are read from the fork's checkout and from the installed
// copy (/usr/share/omarchy/install): between a rebase and the update that
// installs it the two can differ, and a package either allows is not drift.
package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lkingland/onsen/internal/run"
)

// Ours are the edition's own packages, always declared.
var Ours = []string{"omarchy-dev", "omarchy-settings-dev"}

// Replaced are the stock packages ours conflict with. A fresh install has
// them; the run that installs ours removes them, so they are not drift.
var Replaced = []string{"omarchy", "omarchy-settings"}

// quattroLists are Quattro's install lists, by file name without .packages.
var quattroLists = []string{"omarchy-base", "omarchy-x86_64-only", "omarchy-other"}

// Set is the declared set, read from the files.
type Set struct {
	Repo    []string // repository packages
	AUR     []string // aur: packages, without the prefix
	Local   []string // local: packages, without the prefix
	Remove  []string // "-" packages, without the prefix
	Quattro []string // allowed by Quattro's install lists
}

// Names is the first word of each line of a list file, with "#" comments
// and blank lines dropped.
func Names(data string) []string {
	var names []string
	for _, line := range strings.Split(data, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if fields := strings.Fields(line); len(fields) > 0 {
			names = append(names, fields[0])
		}
	}
	return names
}

// Read reads the declared list at list and Quattro's lists from each of
// installDirs that has them.
func Read(list string, installDirs ...string) (*Set, error) {
	data, err := os.ReadFile(list)
	if err != nil {
		return nil, err
	}
	s := &Set{}
	for _, name := range Names(string(data)) {
		switch {
		case strings.HasPrefix(name, "-"):
			s.Remove = append(s.Remove, strings.TrimPrefix(name, "-"))
		case strings.HasPrefix(name, "aur:"):
			s.AUR = append(s.AUR, strings.TrimPrefix(name, "aur:"))
		case strings.HasPrefix(name, "local:"):
			s.Local = append(s.Local, strings.TrimPrefix(name, "local:"))
		default:
			s.Repo = append(s.Repo, name)
		}
	}
	for _, dir := range installDirs {
		for _, l := range quattroLists {
			data, err := os.ReadFile(filepath.Join(dir, l+".packages"))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			s.Quattro = append(s.Quattro, Names(string(data))...)
		}
	}
	s.Repo, s.AUR, s.Local, s.Remove, s.Quattro = uniq(s.Repo), uniq(s.AUR), uniq(s.Local), uniq(s.Remove), uniq(s.Quattro)
	return s, nil
}

// Declared is every package the edition declares: Repo, AUR, Local, and Ours.
func (s *Set) Declared() []string {
	return uniq(append(append(append(append([]string{}, s.Repo...), s.AUR...), s.Local...), Ours...))
}

func uniq(names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range names {
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// Pacman queries and changes the installed packages.
type Pacman struct {
	Run run.Runner
}

func (p Pacman) query(ctx context.Context, flags string) (map[string]bool, error) {
	out, err := p.Run.Output(ctx, run.Cmd{Name: "pacman", Args: []string{flags}, Quiet: true})
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, line := range run.Lines(out) {
		set[strings.TrimSpace(line)] = true
	}
	return set, nil
}

// Installed is every installed package.
func (p Pacman) Installed(ctx context.Context) (map[string]bool, error) {
	return p.query(ctx, "-Qq")
}

// Explicit is every explicitly installed package.
func (p Pacman) Explicit(ctx context.Context) (map[string]bool, error) {
	return p.query(ctx, "-Qqe")
}

// Undeclared is each explicitly installed package that neither the declared
// set nor Quattro's lists allow. A package the edition removes, or one ours
// replace, is not drift: the same run removes it.
func (p Pacman) Undeclared(ctx context.Context, s *Set) ([]string, error) {
	explicit, err := p.Explicit(ctx)
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{}
	for _, list := range [][]string{s.Quattro, s.Declared(), s.Remove, Replaced} {
		for _, n := range list {
			allowed[n] = true
		}
	}
	var out []string
	for n := range explicit {
		if !allowed[n] {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out, nil
}

func filter(names []string, keep func(string) bool) []string {
	var out []string
	for _, n := range names {
		if keep(n) {
			out = append(out, n)
		}
	}
	return out
}

// Remove uninstalls each "-" package that is installed. It runs before our
// packages go in, because one of these may own a path ours ships.
func (p Pacman) Remove(ctx context.Context, s *Set) error {
	installed, err := p.Installed(ctx)
	if err != nil {
		return err
	}
	gone := filter(s.Remove, func(n string) bool { return installed[n] })
	if len(gone) == 0 {
		return nil
	}
	return p.Run.Run(ctx, run.Root("pacman", append([]string{"-R", "--noconfirm"}, gone...)...))
}

// Mark marks the declared packages that are installed only as dependencies
// explicit, so the update's orphan sweep keeps them (neovim, once
// omarchy-nvim is gone).
func (p Pacman) Mark(ctx context.Context, s *Set) error {
	installed, err := p.Installed(ctx)
	if err != nil {
		return err
	}
	explicit, err := p.Explicit(ctx)
	if err != nil {
		return err
	}
	deps := filter(append(append(append([]string{}, s.Repo...), s.AUR...), s.Local...), func(n string) bool {
		return installed[n] && !explicit[n]
	})
	if len(deps) == 0 {
		return nil
	}
	sort.Strings(deps)
	_, err = p.Run.Output(ctx, run.Root("pacman", append([]string{"-D", "--asexplicit"}, deps...)...))
	return err
}

// Install installs each declared package that is missing, repository
// packages with pacman and aur: packages with yay, then marks as Mark does.
// A local: package it cannot install; one that is missing is named, after
// the rest are in.
func (p Pacman) Install(ctx context.Context, s *Set) error {
	installed, err := p.Installed(ctx)
	if err != nil {
		return err
	}
	missing := func(n string) bool { return !installed[n] }
	if need := filter(s.Repo, missing); len(need) > 0 {
		if err := p.Run.Run(ctx, run.Root("pacman", append([]string{"-S", "--needed", "--noconfirm"}, need...)...)); err != nil {
			return err
		}
	}
	if need := filter(s.AUR, missing); len(need) > 0 {
		if err := p.Run.Run(ctx, run.C("yay", append([]string{"-S", "--needed", "--noconfirm"}, need...)...)); err != nil {
			return err
		}
	}
	if err := p.Mark(ctx, s); err != nil {
		return err
	}
	if absent := filter(s.Local, missing); len(absent) > 0 {
		return fmt.Errorf("local packages not installed, built by hand from their own checkouts: %s", strings.Join(absent, ", "))
	}
	return nil
}
