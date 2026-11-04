package path

import "strings"

// globMatch matches a whole path against a glob pattern. `*` never crosses a
// path separator, `**` does. win32 matching is case-insensitive and treats
// both separators as '/'.
func globMatch(p, pattern string, sep byte, nocase bool) bool {
	if nocase {
		p = strings.ToLower(p)
		pattern = strings.ToLower(pattern)
		p = strings.ReplaceAll(p, `\`, "/")
		pattern = strings.ReplaceAll(pattern, `\`, "/")
	}
	p = trimGlobSeparator(p)
	pattern = trimGlobSeparator(pattern)
	return matchGlobRunes([]rune(pattern), []rune(p), rune(sep))
}

func trimGlobSeparator(s string) string {
	for len(s) > 1 && (s[len(s)-1] == '/' || s[len(s)-1] == '\\') {
		s = s[:len(s)-1]
	}
	return s
}

func matchGlobRunes(pattern, name []rune, sep rune) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '*':
			if len(pattern) > 1 && pattern[1] == '*' {
				rest := pattern[2:]
				// "**/" also matches zero directories.
				if len(rest) > 0 && rest[0] == sep {
					if matchGlobRunes(rest[1:], name, sep) {
						return true
					}
				}
				for i := 0; i <= len(name); i++ {
					if matchGlobRunes(rest, name[i:], sep) {
						return true
					}
				}
				return false
			}
			rest := pattern[1:]
			for i := 0; i <= len(name); i++ {
				if matchGlobRunes(rest, name[i:], sep) {
					return true
				}
				if i < len(name) && name[i] == sep {
					break
				}
			}
			return false
		case '?':
			if len(name) == 0 || name[0] == sep {
				return false
			}
			pattern, name = pattern[1:], name[1:]
		case '[':
			matched, rest, valid := matchGlobClass(pattern, name, sep)
			if !valid || !matched {
				return false
			}
			pattern, name = rest, name[1:]
		case '\\':
			if len(pattern) > 1 {
				pattern = pattern[1:]
			}
			if len(name) == 0 || name[0] != pattern[0] {
				return false
			}
			pattern, name = pattern[1:], name[1:]
		default:
			if len(name) == 0 || pattern[0] != name[0] {
				return false
			}
			pattern, name = pattern[1:], name[1:]
		}
	}
	return len(name) == 0
}

func matchGlobClass(pattern, name []rune, sep rune) (matched bool, rest []rune, valid bool) {
	if len(name) == 0 || name[0] == sep {
		return false, nil, true
	}
	i := 1
	negated := false
	if i < len(pattern) && (pattern[i] == '!' || pattern[i] == '^') {
		negated = true
		i++
	}
	first := true
	for i < len(pattern) && (pattern[i] != ']' || first) {
		first = false
		low := pattern[i]
		if low == '\\' && i+1 < len(pattern) {
			i++
			low = pattern[i]
		}
		if i+2 < len(pattern) && pattern[i+1] == '-' && pattern[i+2] != ']' {
			high := pattern[i+2]
			if name[0] >= low && name[0] <= high {
				matched = true
			}
			i += 3
			continue
		}
		if name[0] == low {
			matched = true
		}
		i++
	}
	if i >= len(pattern) {
		return false, nil, false
	}
	if negated {
		matched = !matched
	}
	return matched, pattern[i+1:], true
}
