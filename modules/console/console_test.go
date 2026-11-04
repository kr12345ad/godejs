package console

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newConsoleEnv(t *testing.T) (*env.Env, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var out, errOut bytes.Buffer
	e := env.New(env.Options{
		Args:   []string{"gode", "test.js"},
		Cwd:    t.TempDir(),
		Stdout: &out,
		Stderr: &errOut,
		Stdin:  strings.NewReader(""),
	})
	t.Cleanup(func() { e.JS.Close() })
	Install(e)
	return e, &out, &errOut
}

func TestLogStreams(t *testing.T) {
	e, out, errOut := newConsoleEnv(t)

	jstest.Eval(t, e, `console.log("a", 1)`)
	jstest.Equal(t, out.String(), "a 1\n")
	jstest.Equal(t, errOut.String(), "")

	out.Reset()
	jstest.Eval(t, e, `console.log({a: 1})`)
	jstest.Equal(t, out.String(), "{ a: 1 }\n")

	out.Reset()
	jstest.Eval(t, e, `console.info("i"); console.debug("d")`)
	jstest.Equal(t, out.String(), "i\nd\n")

	out.Reset()
	jstest.Eval(t, e, `console.log("a\nb")`)
	jstest.Equal(t, out.String(), "a\nb\n")

	out.Reset()
	jstest.Eval(t, e, `console.warn("w"); console.error("x")`)
	jstest.Equal(t, errOut.String(), "w\nx\n")
	jstest.Equal(t, out.String(), "")
}

func TestDirAndDirxml(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `console.dir({a: 1})`)
	jstest.Equal(t, out.String(), "{ a: 1 }\n")

	out.Reset()
	jstest.Eval(t, e, `console.dir("x")`)
	jstest.Equal(t, out.String(), "'x'\n")

	out.Reset()
	jstest.Eval(t, e, `console.dirxml("hi", 2)`)
	jstest.Equal(t, out.String(), "hi 2\n")
}

func TestCounters(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `console.count(); console.count(); console.count("x"); console.countReset(); console.count()`)
	want := "default: 1\ndefault: 2\nx: 1\ndefault: 1\n"
	jstest.Equal(t, out.String(), want)
}

func TestTimers(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `console.time("t"); console.timeEnd("t")`)
	if !regexp.MustCompile(`^t: \d+\.\d{3}ms\n$`).MatchString(out.String()) {
		t.Fatalf("timeEnd output = %q", out.String())
	}

	out.Reset()
	jstest.Eval(t, e, `console.time("t"); console.timeLog("t", "step"); console.timeEnd("t")`)
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("timeLog/timeEnd lines = %#v", lines)
	}
	if !regexp.MustCompile(`^t: \d+\.\d{3}ms step$`).MatchString(lines[0]) {
		t.Fatalf("timeLog = %q", lines[0])
	}
	if !regexp.MustCompile(`^t: \d+\.\d{3}ms$`).MatchString(lines[1]) {
		t.Fatalf("timeEnd = %q", lines[1])
	}
}

func TestGroups(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `console.group("G"); console.log("a"); console.log("b\nc"); console.groupEnd(); console.log("z")`)
	jstest.Equal(t, out.String(), "G\n  a\n  b\n  c\nz\n")

	out.Reset()
	jstest.Eval(t, e, `console.groupCollapsed("C"); console.log("x"); console.groupEnd()`)
	jstest.Equal(t, out.String(), "C\n  x\n")
}

func TestAssertAndTrace(t *testing.T) {
	e, out, errOut := newConsoleEnv(t)

	jstest.Eval(t, e, `console.assert(true, "no")`)
	jstest.Equal(t, errOut.String(), "")
	jstest.Equal(t, out.String(), "")

	jstest.Eval(t, e, `console.assert(false, "bad", 2)`)
	jstest.Equal(t, errOut.String(), "Assertion failed: bad 2\n")

	errOut.Reset()
	jstest.Eval(t, e, `console.trace("boom")`)
	if !strings.HasPrefix(errOut.String(), "Trace: boom\n") {
		t.Fatalf("trace output = %q", errOut.String())
	}
}

func TestConsoleConstructor(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `
		globalThis.buf = "";
		const c = new console.Console(
			{ write(s) { globalThis.buf += "O:" + s; } },
			{ write(s) { globalThis.buf += "E:" + s; } }
		);
		c.log("hi");
		c.error("oops");
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.buf`), "O:hi\nE:oops\n")

	jstest.Eval(t, e, `
		globalThis.buf2 = "";
		const c2 = new console.Console({
			stdout: { write(s) { globalThis.buf2 += "O:" + s; } },
			stderr: { write(s) { globalThis.buf2 += "E:" + s; } }
		});
		c2.log("a"); c2.warn("b");
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.buf2`), "O:a\nE:b\n")

	out.Reset()
	jstest.Eval(t, e, `new console.Console().log("z")`)
	jstest.Equal(t, out.String(), "z\n")
}

func TestGlobalConsole(t *testing.T) {
	e, _, _ := newConsoleEnv(t)

	if got := jstest.Str(t, e, `String(globalThis.console === console)`); got != "true" {
		t.Fatalf("global console is not the module instance: %s", got)
	}
	if got := jstest.Str(t, e, `typeof console.log`); got != "function" {
		t.Fatalf("console.log type = %s", got)
	}
}

func TestTable(t *testing.T) {
	e, out, _ := newConsoleEnv(t)

	jstest.Eval(t, e, `console.table(["a", "b"])`)
	want := "" +
		"┌─────────┬────────┐\n" +
		"│ (index) │ Values │\n" +
		"├─────────┼────────┤\n" +
		"│ 0       │ 'a'    │\n" +
		"│ 1       │ 'b'    │\n" +
		"└─────────┴────────┘\n"
	jstest.Equal(t, out.String(), want)
}

func TestClearNonTTY(t *testing.T) {
	e, out, _ := newConsoleEnv(t)
	jstest.Eval(t, e, `console.clear()`)
	jstest.Equal(t, out.String(), "")
}

func TestImplementedMatchesBaseline(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "parity", "baseline", "console.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, k := range baseline {
		have[k] = true
	}
	seen := map[string]bool{}
	for _, k := range Implemented() {
		if !have[k] {
			t.Fatalf("Implemented() key %q is not in the baseline", k)
		}
		if seen[k] {
			t.Fatalf("duplicate Implemented() key %q", k)
		}
		seen[k] = true
	}
	if len(seen) != len(baseline) {
		t.Fatalf("implemented %d of %d console APIs", len(seen), len(baseline))
	}
}
