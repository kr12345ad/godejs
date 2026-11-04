package url

import (
	"strconv"
	"strings"
	"unicode/utf16"
)

// urlParts is the serialized state of one WHATWG URL. Every field except the
// booleans is already percent-encoded; username and password keep their
// encoded form because that is what Node's URL getters return.
type urlParts struct {
	protocol  string // includes the trailing colon, e.g. "https:"
	username  string
	password  string
	host      string
	port      string
	pathname  string
	search    string // without the leading '?'
	hash      string // without the leading '#'
	hasSearch bool
	hasHash   bool
	hasAuth   bool
	opaque    bool // non-special scheme without an authority, e.g. mailto:
}

var defaultPorts = map[string]string{
	"ftp:":   "21",
	"http:":  "80",
	"https:": "443",
	"ws:":    "80",
	"wss:":   "443",
}

func isSpecial(protocol string) bool {
	_, ok := defaultPorts[protocol]
	return ok || protocol == "file:"
}

// href serializes the URL exactly as URL.prototype.href would.
func (p *urlParts) href() string {
	var b strings.Builder
	b.WriteString(p.protocol)
	if p.hasAuth {
		b.WriteString("//")
		if p.username != "" || p.password != "" {
			b.WriteString(p.username)
			if p.password != "" {
				b.WriteString(":")
				b.WriteString(p.password)
			}
			b.WriteString("@")
		}
		b.WriteString(p.host)
		if p.port != "" {
			b.WriteString(":")
			b.WriteString(p.port)
		}
	}
	b.WriteString(p.pathname)
	if p.hasSearch {
		b.WriteString("?")
		b.WriteString(p.search)
	}
	if p.hasHash {
		b.WriteString("#")
		b.WriteString(p.hash)
	}
	return b.String()
}

func (p *urlParts) origin() string {
	switch p.protocol {
	case "http:", "https:", "ws:", "wss:", "ftp:":
		host := p.host
		if p.port != "" {
			host += ":" + p.port
		}
		return p.protocol + "//" + host
	}
	return "null"
}

func (p *urlParts) hostport() string {
	if p.port == "" {
		return p.host
	}
	return p.host + ":" + p.port
}

type urlError struct{}

func (urlError) Error() string { return "Invalid URL" }

// parseURL parses input, resolving it against base when input is relative.
func parseURL(input string, base *urlParts) (*urlParts, error) {
	input = trimURL(input)
	if input == "" && base != nil {
		p := *base
		p.hash = ""
		p.hasHash = false
		return &p, nil
	}

	scheme, rest, hasScheme := splitScheme(input)

	if !hasScheme {
		if base == nil {
			return nil, urlError{}
		}
		return resolveRelative(input, base)
	}

	protocol := strings.ToLower(scheme) + ":"
	if protocol == "file:" {
		return parseFile(rest)
	}
	special := isSpecial(protocol)

	if special {
		rest = strings.TrimLeft(rest, "/\\")
		auth, tail := cutAuthority(rest)
		user, pass, host, port, err := parseAuthority(auth, true)
		if err != nil || host == "" {
			return nil, urlError{}
		}
		path, search, hash, hasSearch, hasHash := splitTail(tail)
		path = normalizePath(encodePath(path, true))
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if path == "" {
			path = "/"
		}
		if defaultPorts[protocol] == port {
			port = ""
		}
		return &urlParts{
			protocol: protocol,
			username: user, password: pass, host: host, port: port,
			pathname: path, search: encodeQuery(search, true), hash: encodeFragment(hash),
			hasSearch: hasSearch, hasHash: hasHash, hasAuth: true,
		}, nil
	}

	// Non-special scheme.
	if strings.HasPrefix(rest, "//") {
		rest = rest[2:]
		auth, tail := cutAuthority(rest)
		user, pass, host, port, err := parseAuthority(auth, false)
		if err != nil {
			return nil, urlError{}
		}
		path, search, hash, hasSearch, hasHash := splitTail(tail)
		path = normalizePath(encodePath(path, false))
		if path != "" && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return &urlParts{
			protocol: protocol,
			username: user, password: pass, host: host, port: port,
			pathname: path, search: encodeQuery(search, false), hash: encodeFragment(hash),
			hasSearch: hasSearch, hasHash: hasHash, hasAuth: true,
		}, nil
	}

	path, search, hash, hasSearch, hasHash := splitTail(rest)
	p := &urlParts{
		protocol:  protocol,
		search:    encodeQuery(search, false),
		hash:      encodeFragment(hash),
		hasSearch: hasSearch,
		hasHash:   hasHash,
	}
	if strings.HasPrefix(path, "/") {
		p.pathname = normalizePath(encodePath(path, false))
	} else {
		p.opaque = true
		p.pathname = encodeOpaque(path)
	}
	return p, nil
}

func parseFile(rest string) (*urlParts, error) {
	p := &urlParts{protocol: "file:", hasAuth: true}
	rest = strings.ReplaceAll(rest, "\\", "/")
	if strings.HasPrefix(rest, "//") {
		rest = rest[2:]
		auth, tail := cutAuthority(rest)
		if auth != "" {
			host := strings.ToLower(auth)
			if host == "localhost" {
				host = ""
			}
			if strings.ContainsAny(host, " :[]") {
				return nil, urlError{}
			}
			p.host = host
		}
		path, search, hash, hasSearch, hasHash := splitTail(tail)
		p.pathname = ensureLeadingSlash(normalizePath(encodePath(path, false)))
		p.search = encodeQuery(search, false)
		p.hash = encodeFragment(hash)
		p.hasSearch, p.hasHash = hasSearch, hasHash
		return p, nil
	}
	path, search, hash, hasSearch, hasHash := splitTail(rest)
	p.pathname = ensureLeadingSlash(normalizePath(encodePath(path, false)))
	p.search = encodeQuery(search, false)
	p.hash = encodeFragment(hash)
	p.hasSearch, p.hasHash = hasSearch, hasHash
	return p, nil
}

func resolveRelative(input string, base *urlParts) (*urlParts, error) {
	switch {
	case strings.HasPrefix(input, "#"):
		p := *base
		p.hash = encodeFragment(input[1:])
		p.hasHash = true
		return &p, nil
	case strings.HasPrefix(input, "?"):
		p := *base
		p.search = encodeQuery(input[1:], isSpecial(base.protocol))
		p.hasSearch = true
		p.hash = ""
		p.hasHash = false
		return &p, nil
	case base.opaque:
		// Opaque paths cannot be joined with a path; only queries and
		// fragments apply.
		p := *base
		path, search, hash, hasSearch, hasHash := splitTail(input)
		if path != "" {
			return nil, urlError{}
		}
		p.search = encodeQuery(search, false)
		p.hasSearch = hasSearch
		p.hash = encodeFragment(hash)
		p.hasHash = hasHash
		return &p, nil
	}

	if strings.HasPrefix(input, "//") {
		auth, tail := cutAuthority(input[2:])
		user, pass, host, port, err := parseAuthority(auth, isSpecial(base.protocol))
		if err != nil {
			return nil, err
		}
		path, search, hash, hasSearch, hasHash := splitTail(tail)
		path = ensureLeadingSlash(normalizePath(encodePath(path, isSpecial(base.protocol))))
		if defaultPorts[base.protocol] == port {
			port = ""
		}
		return &urlParts{
			protocol: base.protocol,
			username: user, password: pass, host: host, port: port,
			pathname: path, search: encodeQuery(search, isSpecial(base.protocol)), hash: encodeFragment(hash),
			hasSearch: hasSearch, hasHash: hasHash, hasAuth: true,
		}, nil
	}

	path, search, hash, hasSearch, hasHash := splitTail(input)
	var joined string
	if strings.HasPrefix(path, "/") {
		joined = path
	} else {
		dir := base.pathname
		if i := strings.LastIndex(dir, "/"); i >= 0 {
			dir = dir[:i+1]
		} else {
			dir = "/"
		}
		joined = dir + path
	}
	p := *base
	p.pathname = ensureLeadingSlash(normalizePath(encodePath(joined, isSpecial(base.protocol))))
	p.search = encodeQuery(search, isSpecial(base.protocol))
	p.hasSearch = hasSearch
	p.hash = encodeFragment(hash)
	p.hasHash = hasHash
	return &p, nil
}

func splitScheme(s string) (scheme, rest string, ok bool) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' {
			if i == 0 {
				return "", s, false
			}
			return s[:i], s[i+1:], true
		}
		if i == 0 {
			if !isAlpha(c) {
				return "", s, false
			}
			continue
		}
		if !isAlpha(c) && !isDigit(c) && c != '+' && c != '-' && c != '.' {
			return "", s, false
		}
	}
	return "", s, false
}

func cutAuthority(s string) (authority, tail string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '/', '?', '#':
			return s[:i], s[i:]
		}
	}
	return s, ""
}

func splitTail(s string) (path, search, hash string, hasSearch, hasHash bool) {
	if i := strings.IndexByte(s, '#'); i >= 0 {
		hash = s[i+1:]
		hasHash = true
		s = s[:i]
	}
	if i := strings.IndexByte(s, '?'); i >= 0 {
		search = s[i+1:]
		hasSearch = true
		s = s[:i]
	}
	return s, search, hash, hasSearch, hasHash
}

func parseAuthority(auth string, special bool) (user, pass, host, port string, err error) {
	if i := strings.LastIndexByte(auth, '@'); i >= 0 {
		userinfo := auth[:i]
		auth = auth[i+1:]
		if j := strings.IndexByte(userinfo, ':'); j >= 0 {
			user = encodeUserinfo(userinfo[:j])
			pass = encodeUserinfo(userinfo[j+1:])
		} else {
			user = encodeUserinfo(userinfo)
		}
	}
	if strings.HasPrefix(auth, "[") {
		if i := strings.IndexByte(auth, ']'); i >= 0 {
			host = strings.ToLower(auth[:i+1])
			rest := auth[i+1:]
			if strings.HasPrefix(rest, ":") {
				port = rest[1:]
			} else if rest != "" {
				return "", "", "", "", urlError{}
			}
		} else {
			return "", "", "", "", urlError{}
		}
	} else {
		if i := strings.LastIndexByte(auth, ':'); i >= 0 {
			host = auth[:i]
			port = auth[i+1:]
		} else {
			host = auth
		}
		host = strings.ToLower(host)
	}
	host = idnaASCII(host)
	if host == "" && special {
		return "", "", "", "", urlError{}
	}
	if strings.ContainsAny(host, " \t\r\n/?#@") {
		return "", "", "", "", urlError{}
	}
	if port != "" {
		n, convErr := strconv.Atoi(port)
		if convErr != nil || n < 0 || n > 65535 || !allDigits(port) {
			return "", "", "", "", urlError{}
		}
		port = strconv.Itoa(n)
	}
	return user, pass, host, port, nil
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func trimURL(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
	return strings.TrimFunc(s, func(r rune) bool { return r <= 0x20 })
}

func ensureLeadingSlash(s string) string {
	if s == "" {
		return "/"
	}
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

// normalizePath removes single- and double-dot segments the way the URL
// parser does, including their percent-encoded spellings.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	leading := strings.HasPrefix(path, "/")
	trailing := strings.HasSuffix(path, "/") || strings.HasSuffix(path, "/.") || strings.HasSuffix(path, "/..")
	segs := strings.Split(path, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		lower := strings.ToLower(seg)
		switch lower {
		case ".", "%2e":
			continue
		case "..", ".%2e", "%2e.", "%2e%2e":
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
			continue
		}
		out = append(out, seg)
	}
	res := strings.Join(out, "/")
	if leading && !strings.HasPrefix(res, "/") {
		res = "/" + res
	}
	if trailing && !strings.HasSuffix(res, "/") {
		res += "/"
	}
	if res == "" && leading {
		return "/"
	}
	return res
}

// normalizeRelativePath normalizes a legacy relative path, preserving leading
// ".." segments (foo/../../../baz becomes ../../baz).
func normalizeRelativePath(path string) string {
	segs := strings.Split(path, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		switch seg {
		case "", ".":
			continue
		case "..":
			if len(out) > 0 && out[len(out)-1] != ".." {
				out = out[:len(out)-1]
			} else {
				out = append(out, "..")
			}
		default:
			out = append(out, seg)
		}
	}
	return strings.Join(out, "/")
}

func isAlpha(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlnum(c byte) bool { return isAlpha(c) || isDigit(c) }

func encodeAuthComponent(s string) string {
	return encodeWith(s, func(c byte) bool {
		return isAlnum(c) || strings.IndexByte("-._~!$&'()*+,;=", c) >= 0
	})
}

func encodeWith(s string, keep func(byte) bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' || keep(c) {
			b.WriteByte(c)
			continue
		}
		const hex = "0123456789ABCDEF"
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0xF])
	}
	return b.String()
}

func encodePath(s string, special bool) string {
	return encodeWith(s, func(c byte) bool {
		if isAlnum(c) || strings.IndexByte("-._~!$&'()*+,;=:@/[]", c) >= 0 {
			return true
		}
		return false
	})
}

func encodeOpaque(s string) string {
	return encodeWith(s, func(c byte) bool {
		if c < 0x20 || c == 0x7F {
			return false
		}
		switch c {
		case ' ', '"', '<', '>', '`', '#':
			return false
		}
		return true
	})
}

func encodeQuery(s string, special bool) string {
	return encodeWith(s, func(c byte) bool {
		if isAlnum(c) || strings.IndexByte("-._~!$&()*+,;=:@/?", c) >= 0 {
			return true
		}
		if c == '\'' {
			return !special
		}
		return false
	})
}

func encodeFragment(s string) string {
	return encodeWith(s, func(c byte) bool {
		if isAlnum(c) || strings.IndexByte("-._~!$&'()*+,;=:@/?", c) >= 0 {
			return true
		}
		return false
	})
}

func encodeUserinfo(s string) string {
	return encodeWith(s, func(c byte) bool {
		if isAlnum(c) || strings.IndexByte("-._~!$&'()*+,", c) >= 0 {
			return true
		}
		return false
	})
}

// encodeFilePath escapes a filesystem path for the pathname of a file URL.
// Unlike URL parsing, a literal '%' in a filesystem path is data.
func encodeFilePath(s string) string {
	var b strings.Builder
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Node's file-path escape table leaves these unescaped but escapes
		// '~' (unlike encodePath/encodeURIComponent).
		if isAlnum(c) || strings.IndexByte("-._!$&'()*+,;=:@/", c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0xF])
	}
	return b.String()
}

func decodePercent(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			b.WriteByte(hexVal(s[i+1])<<4 | hexVal(s[i+2]))
			i += 2
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return isDigit(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func hexVal(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

// --- application/x-www-form-urlencoded -------------------------------------

func formEncode(s string) string {
	var b strings.Builder
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case isAlnum(c) || c == '*' || c == '-' || c == '.' || c == '_':
			b.WriteByte(c)
		case c == ' ':
			b.WriteByte('+')
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0xF])
		}
	}
	return b.String()
}

func formDecode(s string) string {
	s = strings.ReplaceAll(s, "+", " ")
	return decodePercent(s)
}

func parseForm(s string) [][2]string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "&")
	out := make([][2]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		name, value := part, ""
		if i := strings.IndexByte(part, '='); i >= 0 {
			name, value = part[:i], part[i+1:]
		}
		out = append(out, [2]string{formDecode(name), formDecode(value)})
	}
	return out
}

func serializeForm(pairs [][2]string) string {
	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(formEncode(p[0]))
		b.WriteByte('=')
		b.WriteString(formEncode(p[1]))
	}
	return b.String()
}

func compareUTF16(a, b string) int {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			if ua[i] < ub[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(ua) < len(ub):
		return -1
	case len(ua) > len(ub):
		return 1
	}
	return 0
}

// --- IDNA / Punycode (RFC 3492) -------------------------------------------

func idnaASCII(host string) string {
	if host == "" {
		return ""
	}
	labels := strings.Split(host, ".")
	for i, label := range labels {
		if isASCII(label) {
			labels[i] = strings.ToLower(label)
			continue
		}
		encoded, ok := punycodeEncode(label)
		if !ok {
			return host
		}
		labels[i] = "xn--" + encoded
	}
	return strings.Join(labels, ".")
}

func idnaUnicode(host string) string {
	labels := strings.Split(host, ".")
	for i, label := range labels {
		lower := strings.ToLower(label)
		if strings.HasPrefix(lower, "xn--") {
			decoded, ok := punycodeDecode(lower[4:])
			if !ok {
				return ""
			}
			labels[i] = decoded
		}
	}
	return strings.Join(labels, ".")
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

const (
	pcBase        = 36
	pcTMin        = 1
	pcTMax        = 26
	pcSkew        = 38
	pcDamp        = 700
	pcInitialBias = 72
	pcInitialN    = 128
)

func pcAdapt(delta, numPoints int, first bool) int {
	if first {
		delta /= pcDamp
	} else {
		delta /= 2
	}
	delta += delta / numPoints
	k := 0
	for delta > ((pcBase-pcTMin)*pcTMax)/2 {
		delta /= pcBase - pcTMin
		k += pcBase
	}
	return k + (pcBase-pcTMin+1)*delta/(delta+pcSkew)
}

func pcDigit(d int) byte {
	if d < 26 {
		return byte('a' + d)
	}
	return byte('0' + d - 26)
}

func punycodeEncode(s string) (string, bool) {
	runes := []rune(s)
	var out []byte
	basic := 0
	for _, r := range runes {
		if r < 0x80 {
			out = append(out, byte(r))
			basic++
		}
	}
	h := basic
	if basic > 0 && basic < len(runes) {
		out = append(out, '-')
	}
	n, delta, bias := pcInitialN, 0, pcInitialBias
	for h < len(runes) {
		m := int(^uint(0) >> 1)
		for _, r := range runes {
			if int(r) >= n && int(r) < m {
				m = int(r)
			}
		}
		if m-n > (1<<30)/(h+1) {
			return "", false
		}
		delta += (m - n) * (h + 1)
		n = m
		for _, r := range runes {
			c := int(r)
			if c < n {
				delta++
			}
			if c != n {
				continue
			}
			q := delta
			for k := pcBase; ; k += pcBase {
				t := k - bias
				if t < pcTMin {
					t = pcTMin
				} else if t > pcTMax {
					t = pcTMax
				}
				if q < t {
					break
				}
				out = append(out, pcDigit(t+(q-t)%(pcBase-t)))
				q = (q - t) / (pcBase - t)
			}
			out = append(out, pcDigit(q))
			bias = pcAdapt(delta, h+1, h == basic)
			delta = 0
			h++
		}
		delta++
		n++
	}
	return string(out), true
}

func punycodeDecode(s string) (string, bool) {
	var output []rune
	bias, n, i := pcInitialBias, pcInitialN, 0
	pos := strings.LastIndexByte(s, '-')
	if pos > 0 {
		for _, r := range s[:pos] {
			if r >= 0x80 {
				return "", false
			}
			output = append(output, r)
		}
		s = s[pos+1:]
	}
	// With no delimiter the entire label is the encoded section, e.g. the
	// 'mhb8f' of 'xn--mhb8f'.
	for len(s) > 0 {
		oldi, w := i, 1
		for k := pcBase; ; k += pcBase {
			if len(s) == 0 {
				return "", false
			}
			digit, ok := pcDecodeDigit(s[0])
			if !ok {
				return "", false
			}
			s = s[1:]
			i += digit * w
			if i < 0 {
				return "", false
			}
			t := k - bias
			if t < pcTMin {
				t = pcTMin
			} else if t > pcTMax {
				t = pcTMax
			}
			if digit < t {
				break
			}
			w *= pcBase - t
		}
		bias = pcAdapt(i-oldi, len(output)+1, oldi == 0)
		n += i / (len(output) + 1)
		i %= len(output) + 1
		if n > 0x10FFFF {
			return "", false
		}
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = rune(n)
		i++
	}
	return string(output), true
}

func pcDecodeDigit(c byte) (int, bool) {
	switch {
	case c >= 'a' && c <= 'z':
		return int(c - 'a'), true
	case c >= 'A' && c <= 'Z':
		return int(c - 'A'), true
	case c >= '0' && c <= '9':
		return int(c-'0') + 26, true
	}
	return 0, false
}
