// Package skel keeps the config trees /etc/skel ships (~/.config, ~/.gnupg,
// and ~/.ssh) as the edition ships them.
//
// Check reports drift: a file changed or missing since the last sync, or a
// file or directory that appeared beside the ones the edition manages.
// Apply copies the edition's files over the home and records them (the gold
// record) as the copy the next Check compares against.
//
// Paths are relative to the home (.config/tmux/tmux.conf,
// .gnupg/common.conf) everywhere: in Check's output, in the gold record, and
// in the ignore list. An ignored path is never drift; Apply seeds it when it
// is missing and otherwise leaves it to the app that owns it.
package skel

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Trees is one home and the skeleton it follows.
type Trees struct {
	Skel string   // /etc/skel
	Home string   // the home directory
	Gold string   // the gold record: each managed file as of the last sync
	Dirs []string // the trees under Skel and Home: .config, .gnupg, and .ssh
	// Fork is the fork's checkout, where a change is committed before any
	// package ships it: its copy of a file is as clean as the skeleton's, so
	// committing drift clears it before the run that installs it. "" to skip.
	Fork string
	// Ignore is the drift-ignore list, one glob per entry (see Match).
	Ignore []string
}

// The first line of the gold record.
const goldHeader = "# written by onsen update: ~/.config, ~/.gnupg, and ~/.ssh as of the last sync"

// ReadIgnore reads a drift-ignore file: one glob per line; blank lines and
// lines starting with "#" are skipped. A missing file is an empty list.
func ReadIgnore(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var globs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			globs = append(globs, line)
		}
	}
	return globs, nil
}

// state is what Check and Apply both read before they act.
type state struct {
	t       Trees
	ignore  []*regexp.Regexp
	shipped map[string]string // rel -> path under Skel
	gold    map[string]string // rel -> ident
}

func load(t Trees) (*state, error) {
	s := &state{t: t, shipped: map[string]string{}, gold: map[string]string{}}
	for _, glob := range t.Ignore {
		re, err := compileGlob(glob)
		if err != nil {
			return nil, fmt.Errorf("drift-ignore %q: %w", glob, err)
		}
		s.ignore = append(s.ignore, re)
	}
	for _, dir := range t.Dirs {
		if err := s.walk(dir); err != nil {
			return nil, err
		}
	}
	// An empty skeleton is breakage, not a clean machine: Check would pass
	// everything and Apply would remove every recorded file. The first tree
	// (.config) is always shipped; .gnupg only by the edition.
	if len(t.Dirs) > 0 && !s.ships(t.Dirs[0]) {
		return nil, fmt.Errorf("%s ships no files under %s", t.Skel, t.Dirs[0])
	}
	if err := s.readGold(); err != nil {
		return nil, err
	}
	return s, nil
}

// walk lists the files under Skel/dir. Like the os.walk it replaces, it
// follows the tree itself if that is a link, but does not descend into a
// link to a directory below it; any other link, broken ones included, is a
// file. A missing tree ships nothing.
func (s *state) walk(dir string) error {
	root, err := filepath.EvalSymlinks(filepath.Join(s.t.Skel, dir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return nil
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		s.shipped[filepath.Join(dir, rel)] = path
		return nil
	})
}

func (s *state) readGold() error {
	f, err := os.Open(s.t.Gold)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	for n := 1; scan.Scan(); n++ {
		line := scan.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		digest, rel, ok := strings.Cut(line, "  ")
		if !ok {
			return fmt.Errorf("%s:%d: not \"<ident>  <path>\"", s.t.Gold, n)
		}
		s.gold[rel] = digest
	}
	return scan.Err()
}

// ships reports whether the skeleton ships any file under the tree dir.
func (s *state) ships(dir string) bool {
	for rel := range s.shipped {
		if strings.HasPrefix(rel, dir+"/") {
			return true
		}
	}
	return false
}

// Linked is the first directory on the way from home to rel that is a link
// (~/.config/nvim pointing at a checkout, say), or "" when there is none.
// rel itself is not looked at: a managed file may be a link.
func Linked(home, rel string) string {
	parts := strings.Split(filepath.Dir(rel), string(filepath.Separator))
	for i := range parts {
		if parts[0] == "." {
			break
		}
		dir := filepath.Join(parts[:i+1]...)
		if info, err := os.Lstat(filepath.Join(home, dir)); err == nil && info.Mode()&fs.ModeSymlink != 0 {
			return dir
		}
	}
	return ""
}

// forkPath is where the fork's checkout keeps the file it ships at rel:
// config/ for ~/.config, etc/skel/ for the rest.
func (s *state) forkPath(rel string) string {
	if s.t.Fork == "" {
		return ""
	}
	if rest, ok := strings.CutPrefix(rel, ".config/"); ok {
		return filepath.Join(s.t.Fork, "config", rest)
	}
	return filepath.Join(s.t.Fork, "etc/skel", rel)
}

func (s *state) ignored(rel string) bool {
	for _, re := range s.ignore {
		if re.MatchString(rel) {
			return true
		}
	}
	return false
}

// ident names a file's content: "sym:<target>" for a link, "file:<sha256>"
// for a regular file, and "" for anything else, a missing path included.
func ident(path string) (string, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		return "sym:" + target, nil
	}
	if !info.Mode().IsRegular() {
		return "", nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "file:" + hex.EncodeToString(h.Sum(nil)), nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// privateTree is the tree rel sits in when that tree must stay the user's
// alone (.gnupg, .ssh), else "".
func privateTree(rel string) string {
	for _, dir := range []string{".gnupg", ".ssh"} {
		if strings.HasPrefix(rel, dir+"/") {
			return dir
		}
	}
	return ""
}

// parents is every ancestor directory of rel, nearest first, without ".".
func parents(rel string) []string {
	var out []string
	for d := filepath.Dir(rel); d != "." && d != "/"; d = filepath.Dir(d) {
		out = append(out, d)
	}
	return out
}

// Check reports drift, one line each: the managed files in path order, then
// what is new, by directory:
//
//	missing ~/<path>   a managed file is gone
//	changed ~/<path>   a managed file matches neither the record nor /etc/skel
//	new ~/<path>       a file beside the managed ones
//	new ~/<path>/      a directory beside them
//
// Clean means as recorded at the last sync, as /etc/skel ships it now, or
// as the fork's checkout has it (a change committed and not yet shipped).
// With no record yet, /etc/skel is the record.
func Check(t Trees) ([]string, error) {
	s, err := load(t)
	if err != nil {
		return nil, err
	}
	base := s.gold
	if len(base) == 0 {
		base = map[string]string{}
		for rel, path := range s.shipped {
			id, err := ident(path)
			if err != nil {
				return nil, err
			}
			base[rel] = id
		}
	}

	// A managed path below a linked directory is not the edition's to read
	// or write: the link is the drift, named once.
	var drift []string
	linked := map[string]bool{}
	for _, rel := range sortedKeys(base) {
		if s.ignored(rel) {
			continue
		}
		if l := Linked(t.Home, rel); l != "" {
			if !linked[l] {
				linked[l] = true
				drift = append(drift, "linked ~/"+l)
			}
			continue
		}
		live, err := ident(filepath.Join(t.Home, rel))
		if err != nil {
			return nil, err
		}
		if live == "" {
			drift = append(drift, "missing ~/"+rel)
			continue
		}
		if live == base[rel] {
			continue
		}
		if src, ok := s.shipped[rel]; ok {
			now, err := ident(src)
			if err != nil {
				return nil, err
			}
			if live == now {
				continue
			}
		}
		if fork := s.forkPath(rel); fork != "" {
			committed, err := ident(fork)
			if err != nil {
				return nil, err
			}
			if committed != "" && live == committed {
				continue
			}
		}
		drift = append(drift, "changed ~/"+rel)
	}

	// The directories that hold a managed file, below the top of each tree,
	// are watched for anything new. Every ancestor of a managed file is known.
	top := map[string]bool{}
	for _, dir := range t.Dirs {
		top[dir] = true
	}
	watched := map[string]bool{}
	known := map[string]bool{}
	for rel := range base {
		if d := filepath.Dir(rel); d != "." && !top[d] {
			watched[d] = true
		}
		for _, p := range parents(rel) {
			known[p] = true
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
			if e.IsDir() {
				if !known[rel] && !s.ignored(rel+"/") {
					drift = append(drift, "new ~/"+rel+"/")
				}
				continue
			}
			if _, ok := base[rel]; !ok && !s.ignored(rel) {
				drift = append(drift, "new ~/"+rel)
			}
		}
	}
	return drift, nil
}

// Apply copies every file /etc/skel ships into the home where the home's
// copy differs, removes a file the edition stopped shipping if it is as the
// last sync left it, and writes the gold record. It returns one line per
// file it touched.
//
// Apply runs after the update, and Check passed before it, so a file that
// differs from the last record now was changed during the run: by a
// migration, or by hand. The edition's copy wins, the line says so, and the
// old copy is kept under backup first. Without a record, every file
// replaced is kept.
func Apply(t Trees, backup string) ([]string, error) {
	s, err := load(t)
	if err != nil {
		return nil, err
	}
	// Check stops a run on a linked directory; this keeps Apply from ever
	// writing or removing through one, before it changes anything.
	for _, m := range []map[string]string{s.shipped, s.gold} {
		for rel := range m {
			if l := Linked(t.Home, rel); l != "" && !s.ignored(rel) {
				return nil, fmt.Errorf("~/%s is a link; onsen writes nothing through it", l)
			}
		}
	}
	var report, rows []string
	for _, rel := range sortedKeys(s.shipped) {
		src := s.shipped[rel]
		dest := filepath.Join(t.Home, rel)
		appOwned := s.ignored(rel)
		_, lerr := os.Lstat(dest)
		exists := lerr == nil
		if appOwned && exists {
			continue
		}
		live, err := ident(dest)
		if err != nil {
			return report, err
		}
		want, err := ident(src)
		if err != nil {
			return report, err
		}
		if live != want {
			note := ""
			if exists && live != s.gold[rel] {
				// Neither the edition's copy nor the last record: keep it.
				// With no record yet (a machine's first run) that is every
				// file replaced, since Check never saw the paths only the
				// new edition ships.
				kept := filepath.Join(backup, rel)
				if err := copyEntry(dest, kept); err != nil {
					return report, err
				}
				note = "  (old copy in " + kept + ")"
				if len(s.gold) > 0 {
					note = "  (changed during this run) (old copy in " + kept + ")"
				}
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return report, err
			}
			if exists {
				if err := os.Remove(dest); err != nil {
					return report, err
				}
			}
			if err := copyEntry(src, dest); err != nil {
				return report, err
			}
			// gpg and ssh refuse a directory or file others can read.
			if private := privateTree(rel); private != "" {
				if err := os.Chmod(filepath.Join(t.Home, private), 0o700); err != nil {
					return report, err
				}
				if !strings.HasPrefix(want, "sym:") {
					if err := os.Chmod(dest, 0o600); err != nil {
						return report, err
					}
				}
			}
			report = append(report, "~/"+rel+note)
		}
		if !appOwned {
			rows = append(rows, want+"  "+rel)
		}
	}

	// A file the edition stopped shipping goes if it is as the last sync left
	// it; an edited one stays, and the next Check names it.
	for _, rel := range sortedKeys(s.gold) {
		if _, ok := s.shipped[rel]; ok || s.ignored(rel) {
			continue
		}
		dest := filepath.Join(t.Home, rel)
		live, err := ident(dest)
		if err != nil {
			return report, err
		}
		if live != "" && live == s.gold[rel] {
			if err := os.Remove(dest); err != nil {
				return report, err
			}
			report = append(report, "~/"+rel+"  (removed: no longer shipped)")
		}
	}

	return report, writeGold(t.Gold, rows)
}

func writeGold(path string, rows []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".onsen-gold.")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	body := goldHeader + "\n" + strings.Join(rows, "\n") + "\n"
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// copyEntry copies a file or a link (as a link) to dest, creating dest's
// directory. A file keeps its mode (setuid, setgid and sticky included) and
// its modification time.
func copyEntry(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dest)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	mode := info.Mode() & (fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky)
	if err := os.Chmod(dest, mode); err != nil {
		return err
	}
	return os.Chtimes(dest, info.ModTime(), info.ModTime())
}
