package path

import "strings"

func posixIsAbsolute(p string) bool { return len(p) > 0 && p[0] == '/' }

// posixNormalizeString resolves . and .. segments the way Node's
// normalizeString does. It never adds a leading or trailing separator.
func posixNormalizeString(p string, allowAboveRoot bool) string {
	segments := make([]string, 0, 8)
	for i := 0; i < len(p); {
		for i < len(p) && p[i] == '/' {
			i++
		}
		if i >= len(p) {
			break
		}
		j := i
		for j < len(p) && p[j] != '/' {
			j++
		}
		segment := p[i:j]
		i = j
		switch segment {
		case ".":
			// skip
		case "..":
			if len(segments) > 0 && segments[len(segments)-1] != ".." {
				segments = segments[:len(segments)-1]
			} else if allowAboveRoot {
				segments = append(segments, "..")
			}
		default:
			segments = append(segments, segment)
		}
	}
	return strings.Join(segments, "/")
}

func posixNormalize(p string) string {
	if p == "" {
		return "."
	}
	isAbs := p[0] == '/'
	trailing := p[len(p)-1] == '/'
	out := posixNormalizeString(p, !isAbs)
	if out == "" {
		if isAbs {
			return "/"
		}
		if trailing {
			return "./"
		}
		return "."
	}
	if trailing {
		out += "/"
	}
	if isAbs {
		return "/" + out
	}
	return out
}

func posixJoin(parts []string) string {
	if len(parts) == 0 {
		return "."
	}
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return "."
	}
	return posixNormalize(strings.Join(nonEmpty, "/"))
}

func posixResolve(cwd string, parts []string) string {
	if len(parts) == 0 || (len(parts) == 1 && (parts[0] == "" || parts[0] == ".")) {
		// Node's resolve falls back to '.', including when process.cwd()
		// itself returned an empty string.
		if cwd == "" {
			return "."
		}
		return cwd
	}
	resolved := ""
	absolute := false
	for i := len(parts) - 1; i >= 0 && !absolute; i-- {
		p := parts[i]
		if p == "" {
			continue
		}
		resolved = p + "/" + resolved
		absolute = p[0] == '/'
	}
	if !absolute {
		resolved = cwd + "/" + resolved
		absolute = len(cwd) > 0 && cwd[0] == '/'
	}
	out := posixNormalizeString(resolved, !absolute)
	if absolute {
		return "/" + out
	}
	if out == "" {
		return "."
	}
	return out
}

func posixDirname(p string) string {
	if p == "" {
		return "."
	}
	hasRoot := p[0] == '/'
	end := -1
	matchedSlash := true
	for i := len(p) - 1; i >= 1; i-- {
		if p[i] == '/' {
			if !matchedSlash {
				end = i
				break
			}
		} else {
			matchedSlash = false
		}
	}
	if end == -1 {
		if hasRoot {
			return "/"
		}
		return "."
	}
	if hasRoot && end == 1 {
		return "//"
	}
	return p[:end]
}

func posixBasename(p string, suffix *string) string {
	start := 0
	if suffix != nil && len(*suffix) > 0 && len(*suffix) <= len(p) {
		s := *suffix
		if s == p {
			return ""
		}
		end := -1
		matchedSlash := true
		extIdx := len(s) - 1
		firstNonSlashEnd := -1
		for i := len(p) - 1; i >= start; i-- {
			code := p[i]
			if code == '/' {
				if !matchedSlash {
					start = i + 1
					break
				}
			} else {
				if firstNonSlashEnd == -1 {
					matchedSlash = false
					firstNonSlashEnd = i + 1
				}
				if extIdx >= 0 {
					if code == s[extIdx] {
						extIdx--
						if extIdx == -1 {
							end = i
						}
					} else {
						extIdx = -1
						end = firstNonSlashEnd
					}
				}
			}
		}
		if start == end {
			end = firstNonSlashEnd
		} else if end == -1 {
			end = len(p)
		}
		if start < 0 || end < start || end > len(p) {
			return ""
		}
		return p[start:end]
	}
	end := -1
	matchedSlash := true
	for i := len(p) - 1; i >= start; i-- {
		if p[i] == '/' {
			if !matchedSlash {
				start = i + 1
				break
			}
		} else if end == -1 {
			matchedSlash = false
			end = i + 1
		}
	}
	if end == -1 {
		return ""
	}
	return p[start:end]
}

func posixExtname(p string) string {
	startDot := -1
	startPart := 0
	end := -1
	matchedSlash := true
	preDotState := 0
	for i := len(p) - 1; i >= 0; i-- {
		ch := p[i]
		if ch == '/' {
			if !matchedSlash {
				startPart = i + 1
				break
			}
			continue
		}
		if end == -1 {
			matchedSlash = false
			end = i + 1
		}
		if ch == '.' {
			if startDot == -1 {
				startDot = i
			} else if preDotState != 1 {
				preDotState = 1
			}
		} else if startDot != -1 {
			preDotState = -1
		}
	}
	if startDot == -1 || end == -1 || preDotState == 0 ||
		(preDotState == 1 && startDot == end-1 && startDot == startPart+1) {
		return ""
	}
	return p[startDot:end]
}

func posixParse(p string) (root, dir, base, ext, name string) {
	if p == "" {
		return
	}
	start := 0
	if p[0] == '/' {
		root = "/"
		start = 1
	}
	end := -1
	startDot := -1
	// The basename slice begins after the root, so an absolute path never
	// includes its leading slash in base/name.
	startPart := start
	matchedSlash := true
	preDotState := 0
	for i := len(p) - 1; i >= start; i-- {
		ch := p[i]
		if ch == '/' {
			if !matchedSlash {
				startPart = i + 1
				break
			}
			continue
		}
		if end == -1 {
			matchedSlash = false
			end = i + 1
		}
		if ch == '.' {
			if startDot == -1 {
				startDot = i
			} else if preDotState != 1 {
				preDotState = 1
			}
		} else if startDot != -1 {
			preDotState = -1
		}
	}
	if end != -1 {
		if startDot == -1 || preDotState == 0 ||
			(preDotState == 1 && startDot == end-1 && startDot == startPart+1) {
			base = p[startPart:end]
			name = base
		} else {
			name = p[startPart:startDot]
			base = p[startPart:end]
			ext = p[startDot:end]
		}
	}
	if startPart > 0 && startPart != start {
		dir = p[:startPart-1]
	} else {
		dir = root
	}
	return
}

func posixRelative(cwd, from, to string) string {
	if from == to {
		return ""
	}
	fromAbs := posixResolve(cwd, []string{from})
	toAbs := posixResolve(cwd, []string{to})
	if fromAbs == toAbs {
		return ""
	}
	fromStart := 1
	fromEnd := len(fromAbs)
	toStart := 1
	toEnd := len(toAbs)
	fromLen := fromEnd - fromStart
	toLen := toEnd - toStart
	length := fromLen
	if toLen < length {
		length = toLen
	}
	lastCommonSep := -1
	i := 0
	for ; i < length; i++ {
		fromCode := fromAbs[fromStart+i]
		if fromCode != toAbs[toStart+i] {
			break
		} else if fromCode == '/' {
			lastCommonSep = i
		}
	}
	if i == length {
		if toLen > length {
			if toAbs[toStart+i] == '/' {
				return toAbs[toStart+i+1 : toEnd]
			}
			if i == 0 {
				return toAbs[toStart+i : toEnd]
			}
		} else if fromLen > length {
			if fromAbs[fromStart+i] == '/' {
				lastCommonSep = i
			} else if i == 0 {
				lastCommonSep = 0
			}
		}
	}
	var out strings.Builder
	for i = fromStart + lastCommonSep + 1; i <= fromEnd; i++ {
		if i == fromEnd || fromAbs[i] == '/' {
			if out.Len() == 0 {
				out.WriteString("..")
			} else {
				out.WriteString("/..")
			}
		}
	}
	return out.String() + toAbs[toStart+lastCommonSep:toEnd]
}
