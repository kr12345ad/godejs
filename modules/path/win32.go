package path

import (
	"strings"
	"unicode"
)

func winIsSep(c byte) bool { return c == '\\' || c == '/' }

func winIsDeviceRoot(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

var winReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	"COM\u00b9": true, "COM\u00b2": true, "COM\u00b3": true,
	"LPT\u00b9": true, "LPT\u00b2": true, "LPT\u00b3": true,
}

// isWindowsReservedName mirrors Node: the device part is the text before the
// first colon, or everything except the last byte when there is no colon.
func isWindowsReservedName(p string, colonIndex int) bool {
	device := ""
	if colonIndex >= 0 {
		if colonIndex <= len(p) {
			device = p[:colonIndex]
		}
	} else if len(p) > 0 {
		device = p[:len(p)-1]
	}
	return winReservedNames[strings.ToUpper(device)]
}

// winNormalizeString resolves dot segments with both separators accepted.
func winNormalizeString(p string, allowAboveRoot bool) string {
	segments := make([]string, 0, 8)
	for i := 0; i < len(p); {
		for i < len(p) && winIsSep(p[i]) {
			i++
		}
		if i >= len(p) {
			break
		}
		j := i
		for j < len(p) && !winIsSep(p[j]) {
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
	return strings.Join(segments, `\`)
}

func winNormalize(p string) string {
	if p == "" {
		return "."
	}
	n := len(p)
	if n == 1 {
		if winIsSep(p[0]) {
			return `\`
		}
		return p
	}

	device := ""
	rootEnd := 0
	absolute := false
	code := p[0]

	if winIsSep(code) {
		absolute = true
		if winIsSep(p[1]) {
			j := 2
			last := j
			for j < n && !winIsSep(p[j]) {
				j++
			}
			if j < n && j != last {
				firstPart := p[last:j]
				last = j
				for j < n && winIsSep(p[j]) {
					j++
				}
				if j < n && j != last {
					last = j
					for j < n && !winIsSep(p[j]) {
						j++
					}
					if j == n || j != last {
						if firstPart == "." || firstPart == "?" {
							device = `\\` + firstPart
							rootEnd = 4
							colonIndex := strings.IndexByte(p, ':')
							possibleDevice := ""
							if colonIndex >= 0 && colonIndex+1 >= 4 {
								possibleDevice = p[4 : colonIndex+1]
							}
							if isWindowsReservedName(possibleDevice, len(possibleDevice)-1) {
								device = `\\?\` + possibleDevice
								rootEnd = 4 + len(possibleDevice)
							}
						} else if j == n {
							return `\\` + firstPart + `\` + p[last:] + `\`
						} else {
							device = `\\` + firstPart + `\` + p[last:j]
							rootEnd = j
						}
					}
				}
			}
		} else {
			rootEnd = 1
		}
	} else {
		colonIndex := strings.IndexByte(p, ':')
		if colonIndex > 0 {
			if winIsDeviceRoot(code) && colonIndex == 1 {
				device = p[:2]
				rootEnd = 2
				if n > 2 && winIsSep(p[2]) {
					absolute = true
					rootEnd = 3
				}
			} else if isWindowsReservedName(p, colonIndex) {
				device = p[:colonIndex+1]
				rootEnd = colonIndex + 1
			}
		}
	}

	tail := ""
	if rootEnd < n {
		tail = winNormalizeString(p[rootEnd:], !absolute)
	}
	if tail == "" && !absolute {
		tail = "."
	}
	if len(tail) > 0 && winIsSep(p[n-1]) {
		tail += `\`
	}

	if !absolute && device == "" && strings.IndexByte(p, ':') >= 0 {
		if len(tail) >= 2 && winIsDeviceRoot(tail[0]) && tail[1] == ':' {
			return `.\` + tail
		}
		index := strings.IndexByte(p, ':')
		for index != -1 {
			if index == n-1 || winIsSep(p[index+1]) {
				return `.\` + tail
			}
			next := strings.IndexByte(p[index+1:], ':')
			if next == -1 {
				index = -1
			} else {
				index = index + 1 + next
			}
		}
	}

	colonIndex := strings.IndexByte(p, ':')
	if isWindowsReservedName(p, colonIndex) {
		return `.\` + device + tail
	}
	if device == "" {
		if absolute {
			return `\` + tail
		}
		return tail
	}
	if absolute {
		return device + `\` + tail
	}
	return device + tail
}

func winJoin(parts []string) string {
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
	firstPart := nonEmpty[0]
	joined := strings.Join(nonEmpty, `\`)
	needsReplace := true
	slashCount := 0
	if winIsSep(firstPart[0]) {
		slashCount++
		if len(firstPart) > 1 && winIsSep(firstPart[1]) {
			slashCount++
			if len(firstPart) > 2 {
				if winIsSep(firstPart[2]) {
					slashCount++
				} else {
					needsReplace = false
				}
			}
		}
	}
	if needsReplace {
		for slashCount < len(joined) && winIsSep(joined[slashCount]) {
			slashCount++
		}
		if slashCount >= 2 {
			joined = `\` + joined[slashCount:]
		}
	}
	// Reserved device names turn off dot-segment collapsing, as in Node.
	for _, part := range strings.Split(joined, `\`) {
		if part == "" {
			continue
		}
		if colonIndex := strings.IndexByte(part, ':'); colonIndex != -1 && isWindowsReservedName(part, colonIndex) {
			return strings.ReplaceAll(joined, "/", `\`)
		}
	}
	return winNormalize(joined)
}

// winResolveRoot matches the root/device scanner in Node's win32.resolve.
func winResolveRoot(p string) (device string, rootEnd int, isAbsolute bool) {
	n := len(p)
	code := p[0]
	if n == 1 {
		if winIsSep(code) {
			rootEnd = 1
			isAbsolute = true
		}
		return
	}
	if winIsSep(code) {
		isAbsolute = true
		if winIsSep(p[1]) {
			j := 2
			last := j
			for j < n && !winIsSep(p[j]) {
				j++
			}
			if j < n && j != last {
				firstPart := p[last:j]
				last = j
				for j < n && winIsSep(p[j]) {
					j++
				}
				if j < n && j != last {
					last = j
					for j < n && !winIsSep(p[j]) {
						j++
					}
					if j == n || j != last {
						if firstPart != "." && firstPart != "?" {
							device = `\\` + firstPart + `\` + p[last:j]
							rootEnd = j
						} else {
							device = `\\` + firstPart
							rootEnd = 4
						}
					}
				}
			}
		} else {
			rootEnd = 1
		}
		return
	}
	if winIsDeviceRoot(code) && p[1] == ':' {
		device = p[:2]
		rootEnd = 2
		if n > 2 && winIsSep(p[2]) {
			isAbsolute = true
			rootEnd = 3
		}
	}
	return
}

func winResolve(cwd string, parts []string) string {
	if len(parts) == 0 ||
		(len(parts) == 1 && (parts[0] == "" || parts[0] == ".") && len(cwd) > 0 && winIsSep(cwd[0])) {
		// Node's fast path still normalizes the separators of cwd on a POSIX
		// host, where process.cwd() uses forward slashes.
		return strings.ReplaceAll(cwd, "/", `\`)
	}

	resolvedDevice := ""
	resolvedTail := ""
	resolvedAbsolute := false

	for i := len(parts) - 1; i >= -1; i-- {
		var p string
		if i >= 0 {
			p = parts[i]
			if p == "" {
				continue
			}
		} else if resolvedDevice == "" {
			p = cwd
			if len(parts) == 0 || (len(parts) == 1 && (parts[0] == "" || parts[0] == ".") && len(cwd) > 0 && winIsSep(cwd[0])) {
				return strings.ReplaceAll(cwd, "/", `\`)
			}
		} else {
			p = cwd
			if len(p) < 2 || !strings.EqualFold(p[:2], resolvedDevice) {
				p = resolvedDevice + `\`
			}
		}

		device, rootEnd, isAbsolute := winResolveRoot(p)
		if device != "" {
			if resolvedDevice != "" {
				if !strings.EqualFold(device, resolvedDevice) {
					continue
				}
			} else {
				resolvedDevice = device
			}
		}
		if resolvedAbsolute {
			if resolvedDevice != "" {
				break
			}
		} else {
			resolvedTail = p[rootEnd:] + `\` + resolvedTail
			resolvedAbsolute = isAbsolute
			if isAbsolute && resolvedDevice != "" {
				break
			}
		}
	}

	tail := winNormalizeString(resolvedTail, !resolvedAbsolute)
	if resolvedAbsolute {
		return resolvedDevice + `\` + tail
	}
	out := resolvedDevice + tail
	if out == "" {
		return "."
	}
	return out
}

func winSplitRoot(p string) (root, rest string) {
	if p == "" {
		return "", ""
	}
	device, rootEnd, _ := winResolveRoot(p)
	if rootEnd == 0 {
		return "", p
	}
	return winRoot(deviceOr(p, device, rootEnd)), p[rootEnd:]
}

func deviceOr(p, device string, rootEnd int) string {
	if device != "" {
		return device
	}
	return p[:rootEnd]
}

func winRoot(root string) string { return strings.ReplaceAll(root, "/", `\`) }

func winIsAbsolute(p string) bool {
	if p == "" {
		return false
	}
	if winIsSep(p[0]) {
		return true
	}
	return len(p) >= 3 && winIsDeviceRoot(p[0]) && p[1] == ':' && winIsSep(p[2])
}

func winDirname(p string) string {
	if p == "" {
		return "."
	}
	n := len(p)
	if n == 1 {
		if winIsSep(p[0]) {
			return p
		}
		return "."
	}
	rootEnd := -1
	offset := 0
	code := p[0]
	if winIsSep(code) {
		rootEnd = 1
		offset = 1
		if winIsSep(p[1]) {
			j := 2
			last := j
			for j < n && !winIsSep(p[j]) {
				j++
			}
			if j < n && j != last {
				last = j
				for j < n && winIsSep(p[j]) {
					j++
				}
				if j < n && j != last {
					last = j
					for j < n && !winIsSep(p[j]) {
						j++
					}
					if j == n {
						return p
					}
					if j != last {
						rootEnd = j + 1
						offset = j + 1
					}
				}
			}
		}
	} else if n >= 2 && winIsDeviceRoot(code) && p[1] == ':' {
		if n > 2 && winIsSep(p[2]) {
			rootEnd = 3
		} else {
			rootEnd = 2
		}
		offset = rootEnd
	}
	end := -1
	matchedSlash := true
	for i := n - 1; i >= offset; i-- {
		if winIsSep(p[i]) {
			if !matchedSlash {
				end = i
				break
			}
		} else {
			matchedSlash = false
		}
	}
	if end == -1 {
		if rootEnd == -1 {
			return "."
		}
		end = rootEnd
	}
	return p[:end]
}

func winBasename(p string, suffix *string) string {
	start := 0
	if len(p) >= 2 && winIsDeviceRoot(p[0]) && p[1] == ':' {
		start = 2
	}
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
			if winIsSep(code) {
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
		if winIsSep(p[i]) {
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

func winExtname(p string) string {
	start := 0
	if len(p) >= 2 && p[1] == ':' && winIsDeviceRoot(p[0]) {
		start = 2
	}
	startDot := -1
	startPart := start
	end := -1
	matchedSlash := true
	preDotState := 0
	for i := len(p) - 1; i >= start; i-- {
		code := p[i]
		if winIsSep(code) {
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
		if code == '.' {
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

func winParse(p string) (root, dir, base, ext, name string) {
	if p == "" {
		return
	}
	n := len(p)
	rootEnd := 0
	code := p[0]
	if n == 1 {
		if winIsSep(code) {
			root = p
			dir = p
			return
		}
		base = p
		name = p
		return
	}
	if winIsSep(code) {
		rootEnd = 1
		if winIsSep(p[1]) {
			j := 2
			last := j
			for j < n && !winIsSep(p[j]) {
				j++
			}
			if j < n && j != last {
				last = j
				for j < n && winIsSep(p[j]) {
					j++
				}
				if j < n && j != last {
					last = j
					for j < n && !winIsSep(p[j]) {
						j++
					}
					if j == n {
						rootEnd = j
					} else if j != last {
						rootEnd = j + 1
					}
				}
			}
		}
	} else if n >= 2 && winIsDeviceRoot(code) && p[1] == ':' {
		if n <= 2 {
			root = p
			dir = p
			return
		}
		rootEnd = 2
		if winIsSep(p[2]) {
			if n == 3 {
				root = p
				dir = p
				return
			}
			rootEnd = 3
		}
	}
	if rootEnd > 0 {
		root = p[:rootEnd]
	}
	startDot := -1
	startPart := rootEnd
	end := -1
	matchedSlash := true
	preDotState := 0
	for i := n - 1; i >= rootEnd; i-- {
		code = p[i]
		if winIsSep(code) {
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
		if code == '.' {
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
	if startPart > 0 && startPart != rootEnd {
		dir = p[:startPart-1]
	} else {
		dir = root
	}
	return
}

// winFold applies JavaScript's full lowercase mapping closely enough for path
// comparison, including U+0130 (LATIN CAPITAL I WITH DOT ABOVE), which expands
// to "i" plus a combining dot.
func winFold(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\u0130' {
			b.WriteString("i\u0307")
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func splitWinParts(s string) []string {
	if s == "" {
		return nil
	}
	raw := strings.Split(s, `\`)
	out := raw[:0]
	for _, part := range raw {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func winRelative(cwd, from, to string) string {
	if from == to {
		return ""
	}
	fromOrig := winResolve(cwd, []string{from})
	toOrig := winResolve(cwd, []string{to})
	if fromOrig == toOrig {
		return ""
	}
	// Node compares the lowercased resolved paths on UTF-16 code units; a
	// lowercasing that changes length switches to its split-based branch.
	f := strings.ToLower(fromOrig)
	t := strings.ToLower(toOrig)
	if f == t {
		return ""
	}
	if len(fromOrig) != len(f) || len(toOrig) != len(t) {
		fromSplit := strings.Split(fromOrig, `\`)
		toSplit := strings.Split(toOrig, `\`)
		if n := len(fromSplit); n > 0 && fromSplit[n-1] == "" {
			fromSplit = fromSplit[:n-1]
		}
		if n := len(toSplit); n > 0 && toSplit[n-1] == "" {
			toSplit = toSplit[:n-1]
		}
		fromLen, toLen := len(fromSplit), len(toSplit)
		length := fromLen
		if toLen < length {
			length = toLen
		}
		i := 0
		for ; i < length; i++ {
			if winFold(fromSplit[i]) != winFold(toSplit[i]) {
				break
			}
		}
		switch {
		case i == 0:
			return toOrig
		case i == length:
			if toLen > length {
				return strings.Join(toSplit[i:], `\`)
			}
			if fromLen > length {
				return strings.Repeat(`..\`, fromLen-1-i) + ".."
			}
			return ""
		}
		return strings.Repeat(`..\`, fromLen-i) + strings.Join(toSplit[i:], `\`)
	}

	// Trim leading and trailing backslashes (the latter only matters for UNC).
	fromStart := 0
	for fromStart < len(f) && f[fromStart] == '\\' {
		fromStart++
	}
	fromEnd := len(f)
	for fromEnd-1 > fromStart && f[fromEnd-1] == '\\' {
		fromEnd--
	}
	fromLen := fromEnd - fromStart
	toStart := 0
	for toStart < len(t) && t[toStart] == '\\' {
		toStart++
	}
	toEnd := len(t)
	for toEnd-1 > toStart && t[toEnd-1] == '\\' {
		toEnd--
	}
	toLen := toEnd - toStart

	length := fromLen
	if toLen < length {
		length = toLen
	}
	lastCommonSep := -1
	i := 0
	for ; i < length; i++ {
		c := f[fromStart+i]
		if c != t[toStart+i] {
			break
		}
		if c == '\\' {
			lastCommonSep = i
		}
	}
	if i != length {
		if lastCommonSep == -1 {
			return toOrig
		}
	} else {
		if toLen > length {
			if t[toStart+i] == '\\' {
				return toOrig[toStart+i+1:]
			}
			if i == 2 {
				return toOrig[toStart+i:]
			}
		}
		if fromLen > length {
			if f[fromStart+i] == '\\' {
				lastCommonSep = i
			} else if i == 2 {
				lastCommonSep = 3
			}
		}
		if lastCommonSep == -1 {
			lastCommonSep = 0
		}
	}

	var b strings.Builder
	for i = fromStart + lastCommonSep + 1; i <= fromEnd; i++ {
		if i == fromEnd || f[i] == '\\' {
			if b.Len() == 0 {
				b.WriteString("..")
			} else {
				b.WriteString(`\..`)
			}
		}
	}
	out := b.String()
	toStart += lastCommonSep
	if len(out) > 0 {
		return out + toOrig[toStart:toEnd]
	}
	if toStart < len(toOrig) && toOrig[toStart] == '\\' {
		toStart++
	}
	return toOrig[toStart:toEnd]
}

func winToNamespacedPath(cwd, p string) string {
	if p == "" {
		return ""
	}
	resolvedPath := winResolve(cwd, []string{p})
	if len(resolvedPath) <= 2 {
		return p
	}
	if winIsSep(resolvedPath[0]) && len(resolvedPath) > 2 && winIsSep(resolvedPath[1]) {
		code := resolvedPath[2]
		if code != '?' && code != '.' {
			return `\\?\UNC\` + resolvedPath[2:]
		}
	} else if len(resolvedPath) >= 3 && winIsDeviceRoot(resolvedPath[0]) &&
		resolvedPath[1] == ':' && winIsSep(resolvedPath[2]) {
		return `\\?\` + resolvedPath
	}
	return resolvedPath
}
