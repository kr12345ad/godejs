package diag

import "testing"

func TestRender(t *testing.T) {
	got := Render(Diagnostic{
		Program: "gode",
		Message: "ReferenceError: foo is not defined",
		File:    "script.js",
		Line:    3,
		Column:  5,
		Source:  "    foo()",
		Label:   "foo is not defined",
		Help:    "check for a typo, or declare the identifier before using it",
	})
	want := "gode: ReferenceError: foo is not defined\n" +
		"   ╭─[ script.js:3:5 ]\n" +
		"   │\n" +
		" 3 │     foo()\n" +
		"   │     ┬\n" +
		"   │     ╰── foo is not defined\n" +
		"   │\n" +
		"   │ Help: check for a typo, or declare the identifier before using it\n" +
		"───╯\n"
	if got != want {
		t.Fatalf("render mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderWithoutSource(t *testing.T) {
	got := Render(Diagnostic{Program: "gode", Message: "Cannot find module 'x'", Help: "check the path"})
	want := "gode: Cannot find module 'x'\n\nHelp: check the path\n"
	if got != want {
		t.Fatalf("render mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestSourceLine(t *testing.T) {
	src := "one\r\ntwo\nthree"
	if got := SourceLine(src, 2); got != "two" {
		t.Fatalf("got %q", got)
	}
	if got := SourceLine(src, 9); got != "" {
		t.Fatalf("got %q", got)
	}
}
