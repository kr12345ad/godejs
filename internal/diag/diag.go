// Package diag renders GodeJS errors in the boxed, source-anchored style used
// by uutils. stdout stays Node-identical; this is only for diagnostics on
// stderr.
package diag

import (
	"fmt"
	"strconv"
	"strings"

	"godejs/internal/js"
)

// Diagnostic is one rendered error.
type Diagnostic struct {
	// Program is the executable name in the headline ("gode").
	Program string
	// Message is the full headline text, e.g. "ReferenceError: x is not defined".
	Message string
	// File, Line and Column anchor the carets. File may be empty.
	File   string
	Line   int
	Column int
	// Source is the text of the anchored line.
	Source string
	// Label is the short explanation under the caret.
	Label string
	// Help is an optional hint line.
	Help string
}

// Render returns the diagnostic exactly as it should be written to stderr.
//
//	gode: bad option '--foo'
//	   ╭─[ script.js:3:5 ]
//	   │
//	 3 │     foo()
//	   │     ┬
//	   │     ╰── 'foo' is not defined
//	   │
//	   │ Help: check for a typo, or declare 'foo' before using it
//	───╯
func Render(d Diagnostic) string {
	if d.Program == "" {
		d.Program = "gode"
	}
	if d.Message == "" {
		d.Message = "error"
	}
	var b strings.Builder
	b.WriteString(d.Program)
	b.WriteString(": ")
	b.WriteString(d.Message)
	b.WriteByte('\n')

	if d.Source == "" && d.File == "" {
		if d.Help != "" {
			b.WriteString("\nHelp: ")
			b.WriteString(d.Help)
			b.WriteByte('\n')
		}
		return b.String()
	}

	line := d.Line
	if line < 1 {
		line = 1
	}
	col := d.Column
	if col < 1 {
		col = 1
	}
	if col > len(d.Source)+1 {
		col = len(d.Source) + 1
	}
	numWidth := len(strconv.Itoa(line))
	gutter := strings.Repeat(" ", numWidth+2)
	loc := d.File
	if loc == "" {
		loc = "<anonymous>"
	}

	b.WriteString(gutter)
	b.WriteString("╭─[ ")
	b.WriteString(fmt.Sprintf("%s:%d:%d", loc, line, col))
	b.WriteString(" ]\n")
	b.WriteString(gutter)
	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("%*d │ %s\n", numWidth+1, line, d.Source))
	b.WriteString(gutter)
	b.WriteString("│ ")
	b.WriteString(strings.Repeat(" ", col-1))
	b.WriteString("┬\n")
	if d.Label != "" {
		b.WriteString(gutter)
		b.WriteString("│ ")
		b.WriteString(strings.Repeat(" ", col-1))
		b.WriteString("╰── ")
		b.WriteString(d.Label)
		b.WriteByte('\n')
	}
	if d.Help != "" {
		b.WriteString(gutter)
		b.WriteString("│\n")
		b.WriteString(gutter)
		b.WriteString("│ Help: ")
		b.WriteString(d.Help)
		b.WriteByte('\n')
	}
	b.WriteString(strings.Repeat("─", numWidth+2))
	b.WriteString("╯\n")
	return b.String()
}

// FromError turns a JS error into a diagnostic. sourceLines may be the whole
// file; the line at the error position is selected.
func FromError(program string, err error, file, source string) Diagnostic {
	d := Diagnostic{Program: program, Message: "error"}
	switch e := err.(type) {
	case *js.Error:
		d.Message = e.Error()
		d.File = e.File
		if d.File == "" {
			d.File = file
		}
		d.Line = e.Line
		d.Column = e.Column
		d.Label = e.Message
		d.Help = HelpFor(e.Name)
	default:
		d.Message = err.Error()
		d.File = file
		d.Label = err.Error()
	}
	// Node's stack exposes failed requires through err.code. The headline is
	// the only place the boxed format can carry it, and tooling (and the official
	// module tests) grep stderr for MODULE_NOT_FOUND.
	if coder, ok := err.(interface{ JSErrorCode() (string, string) }); ok {
		if _, code := coder.JSErrorCode(); code == "MODULE_NOT_FOUND" {
			if len(d.Message) < 7 || d.Message[:7] != "Error: " {
				d.Message = "Error: " + d.Message
			}
			d.Message += " (MODULE_NOT_FOUND)"
		}
	}
	if strings.Contains(d.Message, "Cannot find module") && !strings.Contains(d.Message, "MODULE_NOT_FOUND") {
		d.Message += " (MODULE_NOT_FOUND)"
	}
	if d.Line < 1 {
		d.Line = 1
	}
	if d.Column < 1 {
		d.Column = 1
	}
	d.Source = SourceLine(source, d.Line)
	d.Column = refineColumn(d.Source, d.Column)
	return d
}

// HelpFor returns the hint shown for a JS error class.
func HelpFor(name string) string {
	switch name {
	case "ReferenceError":
		return "check for a typo, or declare the identifier before using it"
	case "TypeError":
		return "the value does not have the type this operation requires"
	case "SyntaxError":
		return "the file could not be parsed; check the highlighted line"
	case "RangeError":
		return "a value is outside the range this API accepts"
	case "ModuleError":
		return "check the path and that the package is installed"
	default:
		return ""
	}
}

// refineColumn keeps the caret on a meaningful token. JavaScriptCore reports
// the position just past an expression for thrown errors, which would point at
// the closing paren or off the end of the line.
func refineColumn(src string, col int) int {
	if src == "" {
		return 1
	}
	if col < 1 {
		col = 1
	}
	if col > len(src) {
		col = len(src)
	}
	for col > 1 && (src[col-1] == ' ' || src[col-1] == '\t') {
		col--
	}
	if isIdentChar(src[col-1]) {
		for col > 1 && isIdentChar(src[col-2]) {
			col--
		}
		return col
	}
	for i := 0; i < len(src); i++ {
		if src[i] != ' ' && src[i] != '\t' {
			return i + 1
		}
	}
	return 1
}

func isIdentChar(b byte) bool {
	return b == '_' || b == '$' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

// SourceLine returns 1-based line n of source.
func SourceLine(source string, n int) string {
	if n < 1 {
		return ""
	}
	lines := strings.Split(source, "\n")
	if n > len(lines) {
		return ""
	}
	return strings.TrimSuffix(lines[n-1], "\r")
}
