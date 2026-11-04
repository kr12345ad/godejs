package env

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// transformESM rewrites the common ES module syntax into the CommonJS wrapper
// loadFile evaluates. JavaScriptCore has no module loader, so imports become
// require() calls and exports become assignments on module.exports.
//
// The transform is deliberately syntactic: it understands string, template,
// comment and regex literals so code inside them is never rewritten, and it
// only treats import/export declarations at the top level as declarations.
// Unsupported forms are left untouched and produce JavaScriptCore's own
// SyntaxError instead of a wrong module.
// ESM wrapper bindings use names user modules are very unlikely to declare,
// so `const require = createRequire(...)` and `const __dirname = ...` remain
// legal exactly as they are under Node's ESM loader.
const (
	esmExports = "__godeExports"
	esmRequire = "__godeRequire"
	esmModule  = "__godeModule"
)

func transformESM(src, filename string) string {
	t := &esmTransform{src: src}
	body := t.scan()

	var head strings.Builder
	head.WriteString("'use strict';\n")
	head.WriteString("Object.defineProperty(" + esmExports + ", '__esModule', { value: true });\n")
	head.WriteString("const __godeImportMeta = { url: " + jsString(fileURL(filename)) +
		", filename: " + jsString(filename) +
		", dirname: " + jsString(filepath.Dir(filename)) +
		", resolve: (s) => " + esmRequire + ".resolve(s) };\n")
	head.WriteString("const __godeDefault = (m) => (m && m.__esModule ? m.default : m);\n")
	head.WriteString("const __godeImport = (s) => Promise.resolve().then(() => " + esmRequire + "(s));\n")
	head.WriteString("const __godeStar = (e, m) => { for (const k of Object.keys(m || {})) { if (k !== 'default' && !(k in e)) e[k] = m[k]; } };\n")
	for _, line := range t.requires {
		head.WriteString(line)
		head.WriteByte('\n')
	}
	if len(t.exports) > 0 {
		body += "\n" + strings.Join(t.exports, "\n") + "\n"
	}
	return head.String() + body
}

func fileURL(path string) string {
	u := url.URL{Scheme: "file", Path: path}
	return u.String()
}

func jsString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

type esmTransform struct {
	src         string
	requires    []string
	exports     []string
	seq         int
	dynamicOnly bool
}

func (t *esmTransform) nextID(prefix string) string {
	t.seq++
	return fmt.Sprintf("__gode%s%d", prefix, t.seq)
}

func (t *esmTransform) scan() string {
	src := t.src
	var out strings.Builder
	depth := 0
	lastSig := byte(0)
	prevSig := byte(0)

	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == '\'' || c == '"':
			j := scanQuoted(src, i, c)
			out.WriteString(src[i:j])
			i = j
			lastSig = 'x'
			continue
		case c == '`':
			j := scanTemplate(src, i)
			out.WriteString(src[i:j])
			i = j
			lastSig = 'x'
			continue
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			j := i
			for j < len(src) && src[j] != '\n' {
				j++
			}
			out.WriteString(src[i:j])
			i = j
			continue
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			j := i + 2
			for j+1 < len(src) && !(src[j] == '*' && src[j+1] == '/') {
				j++
			}
			if j+1 < len(src) {
				j += 2
			} else {
				j = len(src)
			}
			out.WriteString(src[i:j])
			i = j
			continue
		case c == '/' && canStartRegex(lastSig):
			j := scanRegex(src, i)
			out.WriteString(src[i:j])
			i = j
			lastSig = 'x'
			continue
		}

		if isIdentStart(c) {
			j := i + 1
			for j < len(src) && isIdentPart(src[j]) {
				j++
			}
			word := src[i:j]
			switch {
			case word == "await" && depth == 0:
				// Top-level `await import(x)` becomes a synchronous require;
				// other top-level awaits are not supported by the CJS wrapper.
				k := skipSpace(src, j)
				if strings.HasPrefix(src[k:], "import") && (k+6 == len(src) || !isIdentPart(src[k+6])) {
					p := skipSpace(src, k+6)
					if p < len(src) && src[p] == '(' {
						out.WriteString(esmRequire + "(")
						prevSig = '('
						lastSig = '('
						i = p + 1
						continue
					}
				}
			case word == "import" && prevSig != '.':
				if k, ok := t.rewriteImport(src, i, j, depth); ok {
					out.WriteString(k.text)
					i = k.end
					lastSig = ';'
					prevSig = ';'
					continue
				}
			case word == "export" && depth == 0 && prevSig != '.':
				if k, ok := t.rewriteExport(src, i, j); ok {
					out.WriteString(k.text)
					i = k.end
					lastSig = ';'
					prevSig = ';'
					continue
				}
			}
			out.WriteString(word)
			prevSig = lastSig
			lastSig = src[j-1]
			i = j
			continue
		}

		switch c {
		case '{', '(', '[':
			depth++
		case '}', ')', ']':
			if depth > 0 {
				depth--
			}
		}
		out.WriteByte(c)
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			prevSig = lastSig
			lastSig = c
		}
		i++
	}
	return out.String()
}

type rewrite struct {
	text string
	end  int
}

// rewriteDynamicImports converts import(spec) in CommonJS source into a
// promise-returning helper backed by the module-local require. JavaScriptCore
// has no fetch hook for bare or absolute specifiers, so leaving import() to the
// engine makes packages such as chalk unusable from CJS code.
func rewriteDynamicImports(src string) string {
	if !strings.Contains(src, "import") {
		return src
	}
	t := &esmTransform{src: src, dynamicOnly: true}
	return t.scan()
}

// dynamicImportHelper is prepended to a CommonJS wrapper only when the module
// actually contains a rewritten dynamic import. It mirrors the namespace shape
// Node gives import(): ESM exports are already namespaced, while a CJS
// module.exports becomes the default export alongside its named properties.
const dynamicImportHelper = `const __godeImport = (s) => Promise.resolve().then(() => {
  const m = require(s);
  if (m && m.__esModule) return m;
  const ns = Object.create(null);
  if (m !== null && (typeof m === 'object' || typeof m === 'function')) {
    for (const k of Object.keys(m)) ns[k] = m[k];
  }
  ns.default = m;
  return ns;
});`

// rewriteImport handles import declarations, import.meta and dynamic import().
func (t *esmTransform) rewriteImport(src string, start, wordEnd, depth int) (rewrite, bool) {
	j := skipSpace(src, wordEnd)
	if j < len(src) && src[j] == '.' {
		if t.dynamicOnly {
			return rewrite{}, false
		}
		// import.meta
		if strings.HasPrefix(src[j:], ".meta") {
			return rewrite{text: "__godeImportMeta", end: j + len(".meta")}, true
		}
		return rewrite{}, false
	}
	if j < len(src) && src[j] == '(' {
		return rewrite{text: "__godeImport", end: wordEnd}, true
	}
	if t.dynamicOnly {
		return rewrite{}, false
	}
	if depth != 0 {
		return rewrite{}, false
	}

	blank := func(end int) string {
		return blankRange(src[start:end])
	}

	// import 'specifier';
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		spec, end := scanStringAt(src, j)
		if spec == "" {
			return rewrite{}, false
		}
		end = skipImportAttributes(src, end)
		t.requires = append(t.requires, esmRequire+"("+jsString(spec)+");")
		return rewrite{text: blank(end), end: end}, true
	}

	// Parse the clause up to the `from` keyword.
	fromAt := -1
	braceDepth := 0
	for k := j; k < len(src); {
		c := src[k]
		if c == '\'' || c == '"' {
			k = scanQuoted(src, k, c)
			continue
		}
		if c == '`' {
			k = scanTemplate(src, k)
			continue
		}
		if c == '{' {
			braceDepth++
			k++
			continue
		}
		if c == '}' {
			braceDepth--
			k++
			continue
		}
		if braceDepth == 0 && isIdentStart(c) {
			end := k + 1
			for end < len(src) && isIdentPart(src[end]) {
				end++
			}
			if src[k:end] == "from" {
				fromAt = k
				break
			}
			k = end
			continue
		}
		k++
	}
	if fromAt < 0 {
		return rewrite{}, false
	}
	specAt := skipSpace(src, fromAt+4)
	if specAt >= len(src) || (src[specAt] != '\'' && src[specAt] != '"') {
		return rewrite{}, false
	}
	spec, end := scanStringAt(src, specAt)
	if spec == "" {
		return rewrite{}, false
	}
	end = skipImportAttributes(src, end)

	clause := strings.TrimSpace(src[j:fromAt])
	clause = strings.TrimSuffix(clause, ",")
	t.requires = append(t.requires, importLines(clause, spec, t)...)
	return rewrite{text: blank(end), end: end}, true
}

// importLines turns an import clause into require-based declarations.
func importLines(clause, spec string, t *esmTransform) []string {
	tmp := t.nextID("M")
	req := esmRequire + "(" + jsString(spec) + ")"
	if clause == "" {
		return []string{req + ";"}
	}
	var defaultName, namespace string
	type pair struct{ imported, local string }
	var pairs []pair

	if i := strings.IndexByte(clause, '{'); i >= 0 {
		if j := strings.LastIndexByte(clause, '}'); j > i {
			for _, entry := range strings.Split(stripListComments(clause[i+1:j]), ",") {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				parts := strings.Fields(entry)
				p := pair{imported: parts[0], local: parts[0]}
				if len(parts) >= 3 && parts[1] == "as" {
					p.local = parts[2]
				}
				pairs = append(pairs, p)
			}
		}
		clause = strings.TrimSpace(clause[:i])
	}
	if i := strings.Index(clause, "*"); i >= 0 {
		rest := strings.TrimSpace(clause[i+1:])
		rest = strings.TrimPrefix(rest, "as")
		namespace = strings.TrimSpace(rest)
		clause = strings.TrimSpace(clause[:i])
	}
	clause = strings.TrimSuffix(strings.TrimSpace(clause), ",")
	if clause != "" && clause != "as" {
		defaultName = clause
	}

	var defaults []string
	var named []string
	for _, p := range pairs {
		if p.imported == "default" {
			defaults = append(defaults, p.local)
			continue
		}
		named = append(named, p.imported+": "+p.local)
	}

	if defaultName == "" && namespace == "" && len(defaults) == 0 && len(named) == 0 {
		return []string{req + ";"}
	}
	if defaultName != "" && namespace == "" && len(defaults) == 0 && len(named) == 0 {
		return []string{"const " + defaultName + " = __godeDefault(" + req + ");"}
	}
	if defaultName == "" && namespace != "" && len(defaults) == 0 && len(named) == 0 {
		return []string{"const " + namespace + " = " + req + ";"}
	}
	if defaultName == "" && namespace == "" && len(defaults) == 0 {
		return []string{"const { " + strings.Join(named, ", ") + " } = " + req + ";"}
	}

	lines := []string{"const " + tmp + " = " + req + ";"}
	if defaultName != "" {
		lines = append(lines, "const "+defaultName+" = __godeDefault("+tmp+");")
	}
	for _, name := range defaults {
		lines = append(lines, "const "+name+" = __godeDefault("+tmp+");")
	}
	if namespace != "" {
		lines = append(lines, "const "+namespace+" = "+tmp+";")
	}
	if len(named) > 0 {
		lines = append(lines, "const { "+strings.Join(named, ", ")+" } = "+tmp+";")
	}
	return lines
}

// stripListComments removes // and /* */ comments from an import/export
// clause. Packages commonly put explanatory comments inside multi-line
// export lists, which would otherwise be parsed as specifier names.
func stripListComments(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '/' && i+1 < len(s) && s[i+1] == '/' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		if s[i] == '/' && i+1 < len(s) && s[i+1] == '*' {
			i += 2
			for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
				i++
			}
			if i+1 < len(s) {
				i += 2
			}
			b.WriteByte(' ')
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// parseNamedList converts "a, b as c, default as d" to destructuring syntax.
func parseNamedList(list string) []string {
	var out []string
	for _, raw := range strings.Split(stripListComments(list), ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		parts := strings.Fields(entry)
		if len(parts) == 1 {
			out = append(out, parts[0])
			continue
		}
		if len(parts) >= 3 && parts[1] == "as" {
			out = append(out, parts[2]+": "+parts[0])
			continue
		}
		out = append(out, entry)
	}
	return out
}

// rewriteExport handles every export form at the top level.
func (t *esmTransform) rewriteExport(src string, start, wordEnd int) (rewrite, bool) {
	j := skipSpace(src, wordEnd)
	blank := func(end int) string { return blankRange(src[start:end]) }

	// export default ...
	if strings.HasPrefix(src[j:], "default") && (j+7 == len(src) || !isIdentPart(src[j+7])) {
		k := skipSpace(src, j+7)
		end := expressionEnd(src, k)
		decl := strings.TrimSpace(src[k:end])
		// A named default function/class is a declaration in JavaScript, so
		// the name must stay visible outside the exported value. This matters
		// for packages that both `export default function ui` and re-export
		// `ui` by name.
		if name := declarationName(decl); name != "" && (strings.HasPrefix(decl, "function") ||
			strings.HasPrefix(decl, "class") || strings.HasPrefix(decl, "async function")) {
			decl = rewriteNestedESM(decl)
			text := strings.TrimRight(decl, ";") + "\n" + esmExports + ".default = " + name + ";"
			return rewrite{text: text, end: end}, true
		}
		text := esmExports + ".default = " + src[k:end]
		text = strings.TrimRight(text, ";") + ";"
		return rewrite{text: text, end: end}, true
	}

	// export * from 'm'; / export * as ns from 'm';
	if j < len(src) && src[j] == '*' {
		k := j + 1
		ns := ""
		if p := skipSpace(src, k); strings.HasPrefix(src[p:], "as") && (p+2 == len(src) || !isIdentPart(src[p+2])) {
			nameAt := skipSpace(src, p+2)
			e := nameAt
			for e < len(src) && isIdentPart(src[e]) {
				e++
			}
			ns = src[nameAt:e]
			k = e
		}
		fromAt := skipSpace(src, k)
		if !strings.HasPrefix(src[fromAt:], "from") {
			return rewrite{}, false
		}
		specAt := skipSpace(src, fromAt+4)
		spec, end := scanStringAt(src, specAt)
		if spec == "" {
			return rewrite{}, false
		}
		end = consumeSemicolon(src, end)
		if ns != "" {
			t.requires = append(t.requires, esmExports+"."+ns+" = "+esmRequire+"("+jsString(spec)+");")
		} else {
			t.requires = append(t.requires, "__godeStar("+esmExports+", "+esmRequire+"("+jsString(spec)+"));")
		}
		return rewrite{text: blank(end), end: end}, true
	}

	// export { a, b as c } [from 'm'];
	if j < len(src) && src[j] == '{' {
		closeAt := strings.IndexByte(src[j:], '}')
		if closeAt < 0 {
			return rewrite{}, false
		}
		list := src[j+1 : j+closeAt]
		end := j + closeAt + 1
		fromAt := skipSpace(src, end)
		if strings.HasPrefix(src[fromAt:], "from") {
			specAt := skipSpace(src, fromAt+4)
			spec, e := scanStringAt(src, specAt)
			if spec == "" {
				return rewrite{}, false
			}
			e = consumeSemicolon(src, e)
			id := t.nextID("R")
			lines := []string{"const " + id + " = " + esmRequire + "(" + jsString(spec) + ");"}
			for _, pair := range parseNamedList(list) {
				local, exported := splitAs(pair)
				if local == "default" {
					lines = append(lines, exportAssignment(exported, "__godeDefault("+id+")"))
				} else {
					lines = append(lines, exportAssignment(exported, id+"."+local))
				}
			}
			t.requires = append(t.requires, lines...)
			return rewrite{text: blank(e), end: e}, true
		}
		end = consumeSemicolon(src, end)
		for _, pair := range parseNamedList(list) {
			local, exported := splitAs(pair)
			t.exports = append(t.exports, exportAssignment(exported, local))
		}
		return rewrite{text: blank(end), end: end}, true
	}

	// export const/let/var/function/class/async function
	kwAt := j
	for kwAt < len(src) && isIdentPart(src[kwAt]) {
		kwAt++
	}
	keyword := src[j:kwAt]
	switch keyword {
	case "const", "let", "var":
		end := expressionEnd(src, j)
		decl := rewriteNestedESM(src[j:end])
		names := declaredNames(decl)
		for _, name := range names {
			t.exports = append(t.exports, esmExports+"."+name+" = "+name+";")
		}
		return rewrite{text: blankRange(src[start:j]) + decl, end: end}, true
	case "function", "class", "async":
		end := expressionEnd(src, j)
		decl := rewriteNestedESM(src[j:end])
		name := declarationName(decl)
		if name != "" {
			t.exports = append(t.exports, esmExports+"."+name+" = "+name+";")
		}
		return rewrite{text: blankRange(src[start:j]) + decl, end: end}, true
	}
	return rewrite{}, false
}

// rewriteNestedESM rewrites import() and import.meta inside a declaration that
// the main scanner emits verbatim. Without this, exported function bodies would
// keep syntax that only the ESM wrapper makes legal.
func rewriteNestedESM(src string) string {
	if !strings.Contains(src, "import") {
		return src
	}
	t := &esmTransform{src: src}
	return t.scan()
}

// exportAssignment emits a valid assignment for identifier names and for the
// string-literal names Node allows in export clauses (`as 'module.exports'`).
func exportAssignment(name, value string) string {
	if len(name) >= 2 && (name[0] == '\'' || name[0] == '"') && name[len(name)-1] == name[0] {
		return esmExports + "[" + name + "] = " + value + ";"
	}
	return esmExports + "." + name + " = " + value + ";"
}

// splitAs reads a named-export entry. parseNamedList returns "exported:
// imported" for aliased entries and a bare name otherwise; callers need the
// imported binding on the left and the exported name on the right.
func splitAs(entry string) (local, exported string) {
	entry = strings.TrimSpace(entry)
	if i := strings.IndexByte(entry, ':'); i >= 0 {
		exported = strings.TrimSpace(entry[:i])
		local = strings.TrimSpace(entry[i+1:])
		return local, exported
	}
	return entry, entry
}

func declarationName(decl string) string {
	rest := decl
	for {
		rest = strings.TrimSpace(rest)
		switch {
		case strings.HasPrefix(rest, "async "):
			rest = rest[6:]
			continue
		case strings.HasPrefix(rest, "function"):
			rest = strings.TrimPrefix(rest, "function")
			rest = strings.TrimPrefix(rest, "*")
		case strings.HasPrefix(rest, "class"):
			rest = strings.TrimPrefix(rest, "class")
		}
		break
	}
	rest = strings.TrimSpace(rest)
	end := 0
	for end < len(rest) && isIdentPart(rest[end]) {
		end++
	}
	return rest[:end]
}

// declaredNames extracts the bindings of a const/let/var declaration.
func declaredNames(decl string) []string {
	decl = strings.TrimPrefix(strings.TrimSpace(decl), "const")
	decl = strings.TrimPrefix(decl, "let")
	decl = strings.TrimPrefix(decl, "var")
	var out []string
	depth := 0
	start := 0
	flush := func(seg string) {
		seg = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(seg), ";"))
		if seg == "" {
			return
		}
		if seg[0] == '{' || seg[0] == '[' {
			out = append(out, bindingNames(seg)...)
			return
		}
		end := 0
		for end < len(seg) && isIdentPart(seg[end]) {
			end++
		}
		if end > 0 {
			out = append(out, seg[:end])
		}
	}
	for i := 0; i < len(decl); i++ {
		switch decl[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				flush(decl[start:i])
				start = i + 1
			}
		}
	}
	flush(decl[start:])
	return out
}

func bindingNames(pattern string) []string {
	pattern = strings.TrimSpace(pattern)
	if i := strings.IndexByte(pattern, '='); i >= 0 {
		pattern = pattern[:i]
	}
	var out []string
	for i := 0; i < len(pattern); {
		if isIdentStart(pattern[i]) {
			j := i + 1
			for j < len(pattern) && isIdentPart(pattern[j]) {
				j++
			}
			out = append(out, pattern[i:j])
			i = j
			continue
		}
		i++
	}
	return out
}

// scanQuoted returns the index just past the closing quote.
func scanQuoted(src string, start int, quote byte) int {
	i := start + 1
	for i < len(src) {
		switch src[i] {
		case '\\':
			i += 2
		case quote:
			return i + 1
		default:
			i++
		}
	}
	return len(src)
}

func scanStringAt(src string, start int) (string, int) {
	if start >= len(src) || (src[start] != '\'' && src[start] != '"') {
		return "", start
	}
	end := scanQuoted(src, start, src[start])
	raw := src[start:end]
	if len(raw) >= 2 {
		raw = raw[1 : len(raw)-1]
	}
	raw = strings.ReplaceAll(raw, `\'`, `'`)
	raw = strings.ReplaceAll(raw, `\"`, `"`)
	raw = strings.ReplaceAll(raw, `\\`, `\`)
	return raw, end
}

// scanTemplate returns the index just past the closing backtick, including
// nested ${...} expressions.
func scanTemplate(src string, start int) int {
	i := start + 1
	for i < len(src) {
		switch src[i] {
		case '\\':
			i += 2
		case '`':
			return i + 1
		case '$':
			if i+1 < len(src) && src[i+1] == '{' {
				i = scanTemplateExpr(src, i+2)
				continue
			}
			i++
		default:
			i++
		}
	}
	return len(src)
}

func scanTemplateExpr(src string, start int) int {
	depth := 1
	i := start
	for i < len(src) {
		switch src[i] {
		case '\'', '"':
			i = scanQuoted(src, i, src[i])
		case '`':
			i = scanTemplate(src, i)
		case '{':
			depth++
			i++
		case '}':
			depth--
			i++
			if depth == 0 {
				return i
			}
		default:
			i++
		}
	}
	return len(src)
}

func scanRegex(src string, start int) int {
	i := start + 1
	inClass := false
	for i < len(src) {
		switch src[i] {
		case '\\':
			i += 2
		case '[':
			inClass = true
			i++
		case ']':
			inClass = false
			i++
		case '/':
			if !inClass {
				i++
				for i < len(src) && isIdentPart(src[i]) {
					i++
				}
				return i
			}
			i++
		case '\n':
			return i
		default:
			i++
		}
	}
	return len(src)
}

func canStartRegex(prev byte) bool {
	if prev == 0 {
		return true
	}
	switch prev {
	case '=', '(', '[', '{', ',', ':', ';', '!', '&', '|', '?', '+', '-', '*', '/', '%', '<', '>', '~', '^':
		return true
	}
	return false
}

func skipSpace(src string, i int) int {
	for i < len(src) {
		switch src[i] {
		case ' ', '\t', '\r', '\n':
			i++
		default:
			return i
		}
	}
	return i
}

func skipImportAttributes(src string, i int) int {
	j := skipSpace(src, i)
	for _, kw := range []string{"with", "assert"} {
		if strings.HasPrefix(src[j:], kw) && (j+len(kw) == len(src) || !isIdentPart(src[j+len(kw)])) {
			k := skipSpace(src, j+len(kw))
			if k < len(src) && src[k] == '{' {
				depth := 1
				k++
				for k < len(src) && depth > 0 {
					switch src[k] {
					case '\'', '"':
						k = scanQuoted(src, k, src[k])
						continue
					case '{':
						depth++
					case '}':
						depth--
					}
					k++
				}
				j = k
			}
		}
	}
	return consumeSemicolon(src, j)
}

func consumeSemicolon(src string, i int) int {
	j := skipSpace(src, i)
	if j < len(src) && src[j] == ';' {
		return j + 1
	}
	return i
}

// expressionEnd finds the end of an expression or declaration starting at j.
func expressionEnd(src string, j int) int {
	depth := 0
	prev := byte(0) // last significant byte, for continuation decisions
	for i := j; i < len(src); {
		switch src[i] {
		case '\'', '"':
			i = scanQuoted(src, i, src[i])
			continue
		case '`':
			i = scanTemplate(src, i)
			continue
		case '/':
			if i+1 < len(src) && src[i+1] == '/' {
				for i < len(src) && src[i] != '\n' {
					i++
				}
				continue
			}
			if i+1 < len(src) && src[i+1] == '*' {
				i += 2
				for i+1 < len(src) && !(src[i] == '*' && src[i+1] == '/') {
					i++
				}
				i += 2
				continue
			}
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		case ';':
			if depth == 0 {
				return i + 1
			}
		case '\n':
			if depth == 0 && !continuesExpression(prev) && !startsWithContinuation(src, i+1) {
				return i
			}
		}
		if src[i] != ' ' && src[i] != '\t' && src[i] != '\r' && src[i] != '\n' {
			prev = src[i]
		}
		i++
	}
	return len(src)
}

// continuesExpression reports whether a line ending in prev cannot end a
// statement, so an export declaration must keep scanning onto the next line.
func continuesExpression(prev byte) bool {
	switch prev {
	case '=', '+', '-', '*', '/', '%', '&', '|', '^', ',', ':', '?', '(', '[', '{', '<', '>', '!', '~':
		return true
	}
	return false
}

func startsWithContinuation(src string, i int) bool {
	j := skipSpace(src, i)
	if j >= len(src) {
		return false
	}
	switch src[j] {
	case '.', '?', '+', '-', '*', '/', '%', '&', '|', '^', ',', ')', ']', '}', '=', ':', '<', '>':
		return true
	}
	return false
}

func blankRange(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			b.WriteByte('\n')
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}
