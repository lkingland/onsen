package skel

import (
	"regexp"
	"strings"
)

// Match reports whether name matches the shell glob pattern the way
// default/onsen/drift-ignore has always been read (Python's fnmatch): "*"
// and "?" also cross "/", "[...]" is a set ("[!...]" negated), and the whole
// name must match. Go's path.Match stops "*" at "/", so it cannot stand in.
func Match(pattern, name string) (bool, error) {
	re, err := compileGlob(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(name), nil
}

func compileGlob(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString(`(?s)\A`)
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			b.WriteString(".*")
			for i+1 < len(pattern) && pattern[i+1] == '*' {
				i++
			}
		case '?':
			b.WriteString(".")
		case '[':
			j := i + 1
			if j < len(pattern) && pattern[j] == '!' {
				j++
			}
			if j < len(pattern) && pattern[j] == ']' {
				j++
			}
			for j < len(pattern) && pattern[j] != ']' {
				j++
			}
			if j >= len(pattern) {
				b.WriteString(`\[`)
				continue
			}
			set := pattern[i+1 : j]
			i = j
			b.WriteByte('[')
			if strings.HasPrefix(set, "!") {
				b.WriteByte('^')
				set = set[1:]
			}
			for k := 0; k < len(set); k++ {
				switch set[k] {
				case '\\', '[', ']', '^':
					b.WriteByte('\\')
				}
				b.WriteByte(set[k])
			}
			b.WriteByte(']')
		default:
			b.WriteString(regexp.QuoteMeta(pattern[i : i+1]))
		}
	}
	b.WriteString(`\z`)
	return regexp.Compile(b.String())
}
