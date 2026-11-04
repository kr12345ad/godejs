// Package jstest holds the helpers module tests share. It intentionally does
// not import any builtin module, so a module's own test binary stays acyclic.
package jstest

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/js"
)

// New returns an environment wired to test-owned streams.
func New(t *testing.T) *env.Env {
	t.Helper()
	e := env.New(env.Options{
		Args:   []string{"gode", "test.js"},
		Cwd:    t.TempDir(),
		Env:    []string{"GODE_TEST=1"},
		Stdout: &bytes.Buffer{},
		Stderr: io.Discard,
		Stdin:  strings.NewReader(""),
	})
	t.Cleanup(func() { e.JS.Close() })
	return e
}

// Install installs a module and exposes its exports under name, returning the
// exports object.
func Install(t *testing.T, e *env.Env, name string, install func(*env.Env) js.Object) js.Object {
	t.Helper()
	exports := install(e)
	e.Global().Set(name, exports.V())
	return exports
}

// Eval evaluates JavaScript and fails the test on error.
func Eval(t *testing.T, e *env.Env, src string) js.Value {
	t.Helper()
	v, err := e.JS.Eval(src, "test.js")
	if err != nil {
		t.Fatalf("eval %q: %v", src, err)
	}
	return v
}

// Str evaluates JavaScript and returns the result coerced to a string.
func Str(t *testing.T, e *env.Env, src string) string {
	t.Helper()
	return Eval(t, e, src).Str()
}

// Equal fails unless got == want.
func Equal(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// EqualInt fails unless got == want.
func EqualInt(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

// Throws runs src and asserts it throws, returning the thrown error.
func Throws(t *testing.T, e *env.Env, src string) *js.Error {
	t.Helper()
	_, err := e.JS.Eval(src, "test.js")
	if err == nil {
		t.Fatalf("expected %q to throw", src)
	}
	jsErr, ok := err.(*js.Error)
	if !ok {
		t.Fatalf("expected a JS error from %q, got %T: %v", src, err, err)
	}
	return jsErr
}
