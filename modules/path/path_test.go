package path

import (
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "path", Install)
	return e
}

func expect(t *testing.T, e *env.Env, expr, want string) {
	t.Helper()
	if got := jstest.Str(t, e, expr); got != want {
		t.Errorf("%s = %q, want %q", expr, got, want)
	}
}

func TestJoin(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.join()`, ".")
	expect(t, e, `path.join("")`, ".")
	expect(t, e, `path.join("a", "b")`, "a/b")
	expect(t, e, `path.join("a", "", "b")`, "a/b")
	expect(t, e, `path.join("a", "..", "b")`, "b")
	expect(t, e, `path.join("/a", "/b")`, "/a/b")
	expect(t, e, `path.join("a", "b", "..", "..", "c")`, "c")
}

func TestNormalize(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.normalize("")`, ".")
	expect(t, e, `path.normalize("a//b")`, "a/b")
	expect(t, e, `path.normalize("/../a")`, "/a")
	expect(t, e, `path.normalize("a/../../b")`, "../b")
	expect(t, e, `path.normalize("/")`, "/")
	// Node preserves the trailing separator: path.posix.normalize('./') is './'.
	expect(t, e, `path.normalize("./")`, "./")
	expect(t, e, `path.normalize("foo/../../bar")`, "../bar")
}

func TestBasename(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.basename("/foo/bar/baz.js")`, "baz.js")
	expect(t, e, `path.basename("/foo/bar/baz.js", ".js")`, "baz")
	expect(t, e, `path.basename(".bashrc")`, ".bashrc")
	expect(t, e, `path.basename("/foo/bar/")`, "bar")
	expect(t, e, `path.basename("")`, "")
	// A suffix equal to the whole path strips to the empty string in Node.
	expect(t, e, `path.basename(".js", ".js")`, "")
}

func TestDirname(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.dirname("/foo/bar/baz")`, "/foo/bar")
	expect(t, e, `path.dirname("/foo")`, "/")
	expect(t, e, `path.dirname("foo")`, ".")
	expect(t, e, `path.dirname("")`, ".")
	expect(t, e, `path.dirname("/")`, "/")
	expect(t, e, `path.dirname("a//b")`, "a/")
}

func TestExtname(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.extname("index.js")`, ".js")
	expect(t, e, `path.extname(".bashrc")`, "")
	expect(t, e, `path.extname("a.")`, ".")
	expect(t, e, `path.extname("a.b.c")`, ".c")
	expect(t, e, `path.extname("/a/b/")`, "")
	expect(t, e, `path.extname(".a.b")`, ".b")
}

func TestParse(t *testing.T) {
	e := setup(t)
	expect(t, e, `JSON.stringify(path.parse("/home/user/dir/file.txt"))`,
		`{"root":"/","dir":"/home/user/dir","base":"file.txt","ext":".txt","name":"file"}`)
	expect(t, e, `path.parse("file").dir`, "")
	expect(t, e, `path.parse("file").base`, "file")
	expect(t, e, `path.parse("/").base`, "")
	expect(t, e, `path.parse("").root`, "")
	expect(t, e, `JSON.stringify(path.parse("/x/y.js"))`,
		`{"root":"/","dir":"/x","base":"y.js","ext":".js","name":"y"}`)
}

func TestFormat(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.format({dir: "/x", base: "y.js"})`, "/x/y.js")
	expect(t, e, `path.format({root: "/", name: "a", ext: ".js"})`, "/a.js")
	expect(t, e, `path.format({dir: "/x", name: "a", ext: ".js"})`, "/x/a.js")
	expect(t, e, `path.format({base: "y.js"})`, "y.js")
}

func TestResolveAndRelative(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.resolve("/a/b", "c")`, "/a/b/c")
	expect(t, e, `path.resolve("/a", "/b")`, "/b")
	expect(t, e, `path.relative("/a/b/c", "/a/b")`, "..")
	expect(t, e, `path.relative("/a/b", "/a/b/c")`, "c")
	expect(t, e, `path.relative("/a", "/b")`, "../b")
	expect(t, e, `path.relative("/a", "/a")`, "")
}

func TestPosixProperties(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.sep`, "/")
	expect(t, e, `path.delimiter`, ":")
	expect(t, e, `path.posix.sep`, "/")
	expect(t, e, `typeof path.posix.normalize`, "function")
	expect(t, e, `typeof path.win32.join`, "function")
	expect(t, e, `path.isAbsolute("/x")`, "true")
	expect(t, e, `path.isAbsolute("x")`, "false")
	expect(t, e, `path.toNamespacedPath("/x")`, "/x")
	expect(t, e, `path.toNamespacedPath(123)`, "123")
}

func TestMatchesGlob(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.matchesGlob("/foo/bar", "/foo/*")`, "true")
	expect(t, e, `path.matchesGlob("/foo/bar/baz", "/foo/**")`, "true")
	expect(t, e, `path.matchesGlob("/foo/bar", "/*")`, "false")
	expect(t, e, `path.matchesGlob("/foo/bar.js", "/foo/*.js")`, "true")
	expect(t, e, `path.matchesGlob("/foo/bar.js", "/foo/*.txt")`, "false")
	expect(t, e, `path.matchesGlob("/a/b", "/?/b")`, "true")
	expect(t, e, `path.matchesGlob("/a/b", "/[a-z]/b")`, "true")
	expect(t, e, `path.matchesGlob("/a/b", "/[!a-z]/b")`, "false")
}

func TestWin32(t *testing.T) {
	e := setup(t)
	expect(t, e, `path.win32.sep`, `\`)
	expect(t, e, `path.win32.delimiter`, ";")
	expect(t, e, `path.win32.join("C:\\foo", "bar", "..", "baz")`, `C:\foo\baz`)
	expect(t, e, `path.win32.normalize("C:\\foo\\\\bar\\..\\baz")`, `C:\foo\baz`)
	expect(t, e, `path.win32.basename("C:\\foo\\bar\\")`, "bar")
	expect(t, e, `path.win32.dirname("C:\\foo\\bar")`, `C:\foo`)
	expect(t, e, `path.win32.extname("C:\\foo\\bar.TXT")`, ".TXT")
	expect(t, e, `path.win32.parse("C:\\tmp\\f.txt").root`, `C:\`)
	expect(t, e, `path.win32.parse("C:\\tmp\\f.txt").dir`, `C:\tmp`)
	expect(t, e, `path.win32.parse("C:\\tmp\\f.txt").name`, "f")
	expect(t, e, `path.win32.isAbsolute("C:\\x")`, "true")
	expect(t, e, `path.win32.isAbsolute("C:x")`, "false")
	expect(t, e, `path.win32.isAbsolute("\\\\server\\share")`, "true")
	expect(t, e, `path.win32.relative("C:\\a\\b", "C:\\a\\c")`, `..\c`)
	expect(t, e, `path.win32.resolve("C:\\a", "b")`, `C:\a\b`)
	expect(t, e, `path.win32.toNamespacedPath("C:\\x")`, `\\?\C:\x`)
	expect(t, e, `path.win32.matchesGlob("C:\\Foo\\Bar.TXT", "c:/foo/*.txt")`, "true")
}

func TestArgumentErrors(t *testing.T) {
	e := setup(t)

	err := jstest.Throws(t, e, `path.join(1)`)
	if !strings.Contains(err.Message, `The "path" argument must be of type string. Received type number (1)`) {
		t.Fatalf("join error = %q", err.Message)
	}

	err = jstest.Throws(t, e, `path.basename("/x", 2)`)
	if !strings.Contains(err.Message, `The "ext" argument must be of type string`) {
		t.Fatalf("basename ext error = %q", err.Message)
	}

	err = jstest.Throws(t, e, `path.format(null)`)
	if !strings.Contains(err.Message, `The "pathObject" argument must be of type object. Received null`) {
		t.Fatalf("format error = %q", err.Message)
	}

	err = jstest.Throws(t, e, `path.relative()`)
	if !strings.Contains(err.Message, `The "from" argument must be of type string. Received undefined`) {
		t.Fatalf("relative error = %q", err.Message)
	}
}
