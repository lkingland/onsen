package skel

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is one file whose live copy is not what the edition ships: a
// managed file edited (Check names it), an ignored file the app that owns
// it has rewritten (Check never names it), or a local file under an ignored
// path the edition ships nothing for.
type Entry struct {
	Rel     string // the path below home
	Skel    string // the edition's copy; "" when it ships none
	Live    string // the live file; "" when it is missing
	Ignored bool   // on the drift-ignore list, so never drift
}

// Diff lists every shipped file whose live copy differs from the edition's,
// and every local file below an ignored path beside the managed ones. It is
// the drift the ignore list hides, so a hand can carry it into the fork
// instead of losing it at the next rebuild.
func Diff(t Trees) ([]Entry, error) {
	s, err := load(t)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, rel := range sortedKeys(s.shipped) {
		if Linked(t.Home, rel) != "" {
			continue
		}
		live := filepath.Join(t.Home, rel)
		shipped, err := ident(s.shipped[rel])
		if err != nil {
			return nil, err
		}
		now, err := ident(live)
		if err != nil {
			return nil, err
		}
		if shipped == now {
			continue
		}
		e := Entry{Rel: rel, Skel: s.shipped[rel], Ignored: s.ignored(rel)}
		if now != "" {
			e.Live = live
		}
		out = append(out, e)
	}

	// Local files beside the managed ones that the ignore list covers: a file
	// in a directory the edition ships into, or anything below an ignored
	// directory there.
	top := map[string]bool{}
	for _, dir := range t.Dirs {
		top[dir] = true
	}
	watched := map[string]bool{}
	for rel := range s.shipped {
		if d := filepath.Dir(rel); d != "." && !top[d] {
			watched[d] = true
		}
	}
	dirs := make([]string, 0, len(watched))
	for d := range watched {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, d := range dirs {
		live := filepath.Join(t.Home, d)
		if Linked(t.Home, filepath.Join(d, "x")) != "" {
			continue
		}
		if info, err := os.Stat(live); err != nil || !info.IsDir() {
			continue
		}
		entries, err := os.ReadDir(live)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			rel := filepath.Join(d, e.Name())
			if _, ok := s.shipped[rel]; ok {
				continue
			}
			if e.IsDir() {
				if !s.ignored(rel + "/") {
					continue
				}
				err := filepath.WalkDir(filepath.Join(t.Home, rel), func(path string, de fs.DirEntry, err error) error {
					if err != nil || de.IsDir() {
						return err
					}
					sub, err := filepath.Rel(t.Home, path)
					if err != nil {
						return err
					}
					if _, ok := s.shipped[sub]; !ok {
						out = append(out, Entry{Rel: sub, Live: path, Ignored: true})
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
				continue
			}
			if s.ignored(rel) {
				out = append(out, Entry{Rel: rel, Live: filepath.Join(t.Home, rel), Ignored: true})
			}
		}
	}
	// A directory both watched and below an ignored one is reached twice.
	seen := map[string]bool{}
	uniq := out[:0]
	for _, e := range out {
		if seen[e.Rel] {
			continue
		}
		seen[e.Rel] = true
		uniq = append(uniq, e)
	}
	out = uniq
	sort.Slice(out, func(i, j int) bool { return strings.Compare(out[i].Rel, out[j].Rel) < 0 })
	return out, nil
}
