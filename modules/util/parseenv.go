package util

import (
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"godejs/internal/js"
)

// describeArg renders the received argument for ERR_INVALID_ARG_TYPE wording.
func describeArg(args []js.Value) string {
	if len(args) == 0 {
		return "undefined"
	}
	return describe(args[0])
}

// systemErrorMessage returns the libc strerror text for an errno, matching
// util.getSystemErrorMessage on Linux.
func systemErrorMessage(n int) string {
	if n < 0 {
		n = -n
	}
	if msg := syscall.Errno(n).Error(); msg != "" && msg != "errno 0" {
		return msg
	}
	return "Unknown system error " + strconv.Itoa(n)
}

// ParseEnv implements util.parseEnv(content): the `--env-file` parser Node
// applies to dotenv-style files. It returns the parsed key/value pairs in file
// order (later keys win) and follows Node's quoting and escaping rules:
// single-quoted values are literal, double-quoted values expand \n, \r, \t,
// \\ and \" (and may span lines), and unquoted values are trimmed of
// surrounding whitespace.
func ParseEnv(content string) (map[string]string, []string, error) {
	if !utf8.ValidString(content) {
		return nil, nil, &js.Error{
			Name:    "TypeError",
			Code:    "ERR_INVALID_ARG_VALUE",
			Message: "The argument 'content' is invalid. Received invalid UTF-8",
		}
	}
	out := map[string]string{}
	order := make([]string, 0)
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(strings.TrimSuffix(lines[i], "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, rest, ok := strings.Cut(line, "=")
		if !ok {
			// Node's env-file parser ignores lines it cannot understand rather
			// than failing the whole file.
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || !validEnvKey(key) {
			continue
		}
		rest = strings.TrimSpace(rest)
		value, consumed, err := parseEnvValue(rest, lines[i+1:])
		if err != nil {
			return nil, nil, err
		}
		i += consumed
		if _, seen := out[key]; !seen {
			order = append(order, key)
		}
		out[key] = value
	}
	return out, order, nil
}

// validEnvKey reports whether name is an identifier the parser accepts:
// [A-Za-z_][A-Za-z0-9_]*.
func validEnvKey(name string) bool {
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return name != ""
}

func invalidEnvLine(line string) error {
	return &js.Error{
		Name:    "TypeError",
		Code:    "ERR_INVALID_ARG_VALUE",
		Message: "The argument 'content' is invalid. Received " + quoteEnv(line),
	}
}

func describe(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsBool():
		if v.Bool() {
			return "type boolean (true)"
		}
		return "type boolean (false)"
	case v.IsObject():
		return "an instance of " + v.Tag()
	case v.Tag() == "Symbol":
		return "type symbol (" + jsToString(v) + ")"
	}
	return "an unsupported value"
}

func quoteEnv(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "\\'") + "'"
}

// parseEnvValue decodes one value that starts at rest, possibly consuming
// continuation lines from following for multi-line double-quoted values.
func parseEnvValue(rest string, following []string) (string, int, error) {
	raw := rest
	if rest == "" {
		return "", 0, nil
	}
	quote := rest[0]
	if quote == '\'' || quote == '"' || quote == '`' {
		var b strings.Builder
		consumed := 0
		i := 1
		for {
			if i >= len(rest) {
				if len(following) == 0 {
					// An unterminated quote is not fatal in Node: the value is the
					// literal text after "=" and parsing continues on the next
					// line.
					return raw, 0, nil
				}
				b.WriteByte('\n')
				consumed++
				rest = strings.TrimSuffix(following[0], "\r")
				following = following[1:]
				i = 0
				continue
			}
			c := rest[i]
			if c == quote {
				i++
				break
			}
			if quote == '\'' && c == '\\' {
				// Single quotes are literal in Node's parser.
				b.WriteByte(c)
				i++
				continue
			}
			if c == '\\' && quote != '\'' && i+1 < len(rest) {
				esc := rest[i+1]
				switch esc {
				case 'n':
					b.WriteByte('\n')
				case 'r':
					b.WriteByte('\r')
				case 't':
					b.WriteByte('\t')
				case '\\':
					b.WriteByte('\\')
				case '"':
					b.WriteByte('"')
				case '`':
					b.WriteByte('`')
				default:
					b.WriteByte('\\')
					b.WriteByte(esc)
				}
				i += 2
				continue
			}
			b.WriteByte(c)
			i++
		}
		// Anything after the closing quote must be whitespace or a comment.
		if tail := strings.TrimSpace(rest[i:]); tail != "" {
			if !strings.HasPrefix(tail, "#") {
				return "", 0, invalidEnvLine(tail)
			}
		}
		return b.String(), consumed, nil
	}
	// Unquoted: a `#` starts an inline comment even without preceding space
	// (Node's env-file parser stops the value at the first hash).
	value := rest
	if idx := strings.Index(value, "#"); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value), 0, nil
}
