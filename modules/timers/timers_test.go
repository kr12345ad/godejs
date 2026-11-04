package timers

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/jstest"
)

func newTimersEnv(t *testing.T) (*env.Env, *[]string) {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "timers", Install)
	var log []string
	e.Global().SetFunc("report", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			log = append(log, args[0].Str())
		}
		return e.JS.Undefined(), nil
	})
	return e, &log
}

func TestModuleSurfaceAndGlobals(t *testing.T) {
	e, _ := newTimersEnv(t)
	got := jstest.Str(t, e, `[
		typeof timers.setTimeout,
		typeof timers.setInterval,
		typeof timers.setImmediate,
		typeof timers.clearTimeout,
		typeof timers.clearInterval,
		typeof timers.clearImmediate,
		typeof timers.Timeout,
		typeof timers.Immediate,
		typeof setTimeout,
		typeof setInterval,
		typeof setImmediate,
		typeof clearTimeout,
		typeof clearInterval,
		typeof clearImmediate
	].join(",")`)
	want := "function,function,function,function,function,function,function,function," +
		"function,function,function,function,function,function"
	if got != want {
		t.Fatalf("surface = %s, want %s", got, want)
	}
}

func TestSetTimeoutForwardsArgsAndRunsBeforeImmediate(t *testing.T) {
	e, log := newTimersEnv(t)
	jstest.Eval(t, e, `
		setTimeout((x) => report("t" + x), 0, "1")
		setImmediate((x) => report("i" + x), "2")
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if want := []string{"t1", "i2"}; !reflect.DeepEqual(*log, want) {
		t.Fatalf("log = %v, want %v", *log, want)
	}
}

func TestClearTimeoutByObjectAndByNumber(t *testing.T) {
	e, log := newTimersEnv(t)
	jstest.Eval(t, e, `
		const a = setTimeout(() => report("a"), 1000)
		clearTimeout(a)
		const b = setTimeout(() => report("b"), 1000)
		clearTimeout(Number(b))
		const c = setImmediate(() => report("c"))
		clearImmediate(c)
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(*log) != 0 {
		t.Fatalf("cleared timers fired: %v", *log)
	}
}

func TestIntervalRepeatsUntilClosed(t *testing.T) {
	e, log := newTimersEnv(t)
	jstest.Eval(t, e, `
		let n = 0
		const iv = setInterval(() => {
			n++
			report("tick")
			if (n === 3) iv.close()
		}, 1)
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := jstest.Str(t, e, `String(n)`); got != "3" {
		t.Fatalf("n = %s, want 3", got)
	}
	if want := []string{"tick", "tick", "tick"}; !reflect.DeepEqual(*log, want) {
		t.Fatalf("log = %v, want %v", *log, want)
	}
}

func TestTimeoutObjectMethods(t *testing.T) {
	e, _ := newTimersEnv(t)
	got := jstest.Str(t, e, `
		const t = setTimeout(() => {}, 1000)
		const result = [
			t.hasRef(),
			t.ref() === t,
			t.unref() === t,
			t.refresh() === t,
			t instanceof timers.Timeout,
			typeof t[Symbol.toPrimitive] === "function" && Number(t) > 0
		]
		t.close()
		result.join(",")
	`)
	jstest.Equal(t, got, "true,true,true,true,true,true")
}

func TestImmediateObjectMethods(t *testing.T) {
	e, _ := newTimersEnv(t)
	got := jstest.Str(t, e, `
		const i = setImmediate(() => {})
		const result = [
			i instanceof timers.Immediate,
			i.hasRef(),
			i.ref() === i,
			i.unref() === i
		]
		clearImmediate(i)
		result.join(",")
	`)
	jstest.Equal(t, got, "true,true,true,true")
}

func TestNonFunctionCallbackThrowsTypeError(t *testing.T) {
	e, _ := newTimersEnv(t)
	err := jstest.Throws(t, e, `setTimeout("nope", 0)`)
	if err.Name != "TypeError" {
		t.Fatalf("name = %q, want TypeError", err.Name)
	}
	// Node appends the received type to the validator message.
	if want := `The "callback" argument must be of type function. Received type string ('nope')`; err.Message != want {
		t.Fatalf("message = %q, want %q", err.Message, want)
	}
	if err.Code != "ERR_INVALID_ARG_TYPE" {
		t.Fatalf("code = %q, want ERR_INVALID_ARG_TYPE", err.Code)
	}
}

// Node keeps the requested duration on _idleTimeout (clamped to at least 1)
// but schedules on the truncated millisecond value, so 1.8 and 1.1 share the
// 1 ms bucket and keep insertion order.
func TestTimeoutBookkeeping(t *testing.T) {
	e, _ := newTimersEnv(t)
	got := jstest.Str(t, e, `
		const a = setTimeout(() => {}, 1.8)
		const b = setTimeout(() => {}, 0)
		const c = setInterval(() => {}, undefined)
		const out = [
			a._idleTimeout,
			b._idleTimeout,
			a._repeat,
			c._repeat,
			a._destroyed,
			a._onTimeout === null ? "null" : "fn",
			a.unref().hasRef(),
			a.ref().hasRef()
		]
		a.close()
		out.push(String(a._destroyed), String(a._idleTimeout))
		b.close(); c.close()
		out.join("|")
	`)
	// Array.prototype.join renders both null and undefined as an empty
	// string, so the null _repeat for a one-shot timer shows up as nothing.
	jstest.Equal(t, got, "1.8|1||1|false|fn|false|true|true|-1")
}

// clearTimeout accepts the primitive id and its string form, and clearing
// null or nothing is a no-op.
func TestClearByPrimitiveId(t *testing.T) {
	e, log := newTimersEnv(t)
	jstest.Eval(t, e, `
		const a = setTimeout(() => report("a"), 1)
		clearTimeout(Number(a))
		const b = setTimeout(() => report("b"), 1)
		clearTimeout(String(b))
		clearTimeout(null)
		clearTimeout()
		clearImmediate(null)
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(*log) != 0 {
		t.Fatalf("cleared timers fired: %v", *log)
	}
}

func TestNegativeAndUndefinedDelayAreZero(t *testing.T) {
	e, log := newTimersEnv(t)
	jstest.Eval(t, e, `
		setTimeout(() => report("negative"), -10)
		setTimeout(() => report("undefined"))
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(*log) != 2 {
		t.Fatalf("log = %v, want both timers", *log)
	}
}

func TestImplementedKeysAreTracked(t *testing.T) {
	data, err := os.ReadFile("../../parity/baseline/timers.json")
	if err != nil {
		t.Skipf("baseline unavailable: %v", err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("baseline: %v", err)
	}
	tracked := map[string]bool{}
	for _, k := range baseline {
		tracked[k] = true
	}
	if len(Implemented()) == 0 {
		t.Fatal("no implemented keys")
	}
	for _, k := range Implemented() {
		if !tracked[k] {
			t.Fatalf("implemented key %q is not in parity/baseline/timers.json", k)
		}
	}
}
