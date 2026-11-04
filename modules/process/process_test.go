package process

import (
	"errors"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/timers"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "process", Install)
	// beforeExit handlers schedule timers, so the event loop needs them.
	timers.Install(e)
	return e
}

func expect(t *testing.T, e *env.Env, expr, want string) {
	t.Helper()
	if got := jstest.Str(t, e, expr); got != want {
		t.Errorf("%s = %q, want %q", expr, got, want)
	}
}

func TestEnvCoercion(t *testing.T) {
	e := setup(t)

	expect(t, e, `process.env.GODE_TEST_ENV = 42; process.env.GODE_TEST_ENV`, "42")
	expect(t, e, `process.env.GODE_TEST_ENV = undefined; process.env.GODE_TEST_ENV`, "undefined")
	expect(t, e, `process.env.GODE_TEST_ENV = false; process.env.GODE_TEST_ENV`, "false")
	expect(t, e, `delete process.env.GODE_TEST_ENV; String(process.env.GODE_TEST_ENV)`, "undefined")
	expect(t, e, `'GODE_TEST_ENV' in process.env`, "false")
	expect(t, e, `delete process.env.NOT_THERE`, "true")
	expect(t, e, `'GODE_TEST' in process.env`, "true")
}

func TestEnvSymbols(t *testing.T) {
	e := setup(t)

	expect(t, e, `(function(){ var s = Symbol('s'); return String(process.env[s]); })()`, "undefined")
	expect(t, e, `(function(){ var s = Symbol('s'); return String(s in process.env); })()`, "false")
	expect(t, e, `(function(){
		var s = Symbol('s');
		try { process.env[s] = 42; return 'no'; } catch (err) { return err.name; }
	})()`, "TypeError")
	expect(t, e, `(function(){
		try { process.env.GODE_SYM = Symbol('s'); return 'no'; } catch (err) { return err.name; }
	})()`, "TypeError")
}

func TestExitCodeValidation(t *testing.T) {
	e := setup(t)

	expect(t, e, `process.exitCode = '2'; String(process.exitCode)`, "2")
	expect(t, e, `process.exitCode = undefined; String(process.exitCode)`, "0")
	expect(t, e, `(function(){ try { process.exitCode = 2.1; } catch (err) { return err.code; } })()`, "ERR_OUT_OF_RANGE")
	expect(t, e, `(function(){ try { process.exitCode = 'two'; } catch (err) { return err.code; } })()`, "ERR_INVALID_ARG_TYPE")
	expect(t, e, `(function(){ try { process.exitCode = true; } catch (err) { return err.code; } })()`, "ERR_INVALID_ARG_TYPE")
	expect(t, e, `(function(){
		'use strict';
		try { delete process.exitCode; return 'no'; } catch (err) { return err.name; }
	})()`, "TypeError")
}

func TestHrtimeValidation(t *testing.T) {
	e := setup(t)

	expect(t, e, `Array.isArray(process.hrtime()) + ':' + process.hrtime().length`, "true:2")
	expect(t, e, `(function(){ try { process.hrtime(1); } catch (err) { return err.code; } })()`, "ERR_INVALID_ARG_TYPE")
	expect(t, e, `(function(){ try { process.hrtime([]); } catch (err) { return err.code; } })()`, "ERR_OUT_OF_RANGE")
	expect(t, e, `(function(){ try { process.hrtime([1, 2, 3]); } catch (err) { return err.code; } })()`, "ERR_OUT_OF_RANGE")
	expect(t, e, `typeof process.hrtime.bigint()`, "bigint")
}

func TestNextTickValidation(t *testing.T) {
	e := setup(t)

	expect(t, e, `(function(){ try { process.nextTick(null); } catch (err) { return err.code + ':' + err.name; } })()`, "ERR_INVALID_ARG_TYPE:TypeError")
}

func TestUmaskValidation(t *testing.T) {
	e := setup(t)

	expect(t, e, `typeof process.umask()`, "number")
	expect(t, e, `(function(){
		var old = process.umask('0664');
		var set = process.umask(old);
		return typeof set;
	})()`, "number")
	expect(t, e, `(function(){ try { process.umask({}); } catch (err) { return err.code; } })()`, "ERR_INVALID_ARG_TYPE")
	expect(t, e, `(function(){ try { process.umask('abc'); } catch (err) { return err.code; } })()`, "ERR_INVALID_ARG_VALUE")
}

func TestHrtimeMessages(t *testing.T) {
	e := setup(t)

	expect(t, e, `(function(){ try { process.hrtime(1); } catch (err) { return err.message; } })()`,
		`The "time" argument must be an instance of Array. Received type number (1)`)
	expect(t, e, `(function(){ try { process.hrtime([]); } catch (err) { return err.message; } })()`,
		`The value of "time" is out of range. It must be 2. Received 0`)
}

func TestActiveResourcesAndFlags(t *testing.T) {
	e := setup(t)

	// Without the timers builtin there is nothing to track, but the shape
	// must still match Node.
	expect(t, e, `Array.isArray(process.getActiveResourcesInfo())`, "true")
	expect(t, e, `process.getActiveRequests === process.getActiveResourcesInfo`, "true")
	expect(t, e, `process.allowedNodeEnvironmentFlags.has('--inspect')`, "true")
	expect(t, e, `process.allowedNodeEnvironmentFlags.has('--not-a-flag')`, "false")
	expect(t, e, `typeof process.debugPort`, "number")
}

func TestUmaskIgnoresHighBits(t *testing.T) {
	e := setup(t)

	// libuv masks the mode with 0o777, so 0o10664 reads back as 0o664.
	expect(t, e, `(function(){
		var old = process.umask();
		process.umask(0o10664);
		var now = process.umask();
		process.umask(old);
		return String(now);
	})()`, "436")
}

func TestUncaughtCaptureBookkeeping(t *testing.T) {
	e := setup(t)

	expect(t, e, `process.hasUncaughtExceptionCaptureCallback()`, "false")
	expect(t, e, `(function(){
		process.setUncaughtExceptionCaptureCallback(function () {});
		var has = process.hasUncaughtExceptionCaptureCallback();
		var again = 'threw';
		try { process.setUncaughtExceptionCaptureCallback(function () {}); } catch (err) {
			again = err.code;
		}
		process.setUncaughtExceptionCaptureCallback(null);
		return has + ':' + again + ':' + process.hasUncaughtExceptionCaptureCallback();
	})()`, "true:ERR_UNCAUGHT_EXCEPTION_CAPTURE_ALREADY_SET:false")
}

// An 'uncaughtException' listener absorbs an exception that escaped the loop.
func TestUncaughtExceptionDelivery(t *testing.T) {
	e := setup(t)

	jstest.Eval(t, e, `globalThis.seen = '';
		process.on('uncaughtException', function (err, origin) {
			globalThis.seen = err.message + ':' + origin;
		});`)
	if !e.DeliverUncaught(errors.New("boom"), "uncaughtException") {
		t.Fatal("DeliverUncaught reported the exception unhandled")
	}
	if got := jstest.Str(t, e, `globalThis.seen`); got != "boom:uncaughtException" {
		t.Fatalf("handler saw %q", got)
	}
	if e.ExitCode != 0 {
		t.Fatalf("handled exception changed the exit code to %d", e.ExitCode)
	}

	// A throwing listener is itself unhandled and ends the process with 7.
	jstest.Eval(t, e, `process.on('uncaughtException', function () {
		throw new Error('again');
	});`)
	if e.DeliverUncaught(errors.New("second"), "uncaughtException") {
		t.Fatal("a throwing listener reported the exception handled")
	}
	if e.ExitCode != 7 {
		t.Fatalf("exit code = %d, want 7", e.ExitCode)
	}
}

// With no listener installed the CLI still sees the error.
func TestUncaughtExceptionUnhandled(t *testing.T) {
	e := setup(t)
	if e.DeliverUncaught(errors.New("boom"), "uncaughtException") {
		t.Fatal("DeliverUncaught reported a handled exception without a listener")
	}
}

// beforeExit fires when the loop empties and can schedule more work.
func TestBeforeExitRuns(t *testing.T) {
	e := setup(t)

	jstest.Eval(t, e, `globalThis.passes = 0;
		globalThis.scheduled = false;
		process.on('beforeExit', function () {
			globalThis.passes++;
			if (!globalThis.scheduled) {
				globalThis.scheduled = true;
				setTimeout(function () {}, 1);
			}
		});`)
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	if got := jstest.Str(t, e, `String(globalThis.passes)`); got != "2" {
		t.Fatalf("beforeExit passes = %s, want 2", got)
	}
}

// Warnings are delivered on the next tick, so a listener registered right
// after emitWarning still sees them.
func TestEmitWarningDefers(t *testing.T) {
	e := setup(t)

	jstest.Eval(t, e, `globalThis.w = 'none';
		process.emitWarning('deferred');
		process.on('warning', function (w) { globalThis.w = w.name + ':' + w.message; });`)
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	if got := jstest.Str(t, e, `globalThis.w`); got != "Warning:deferred" {
		t.Fatalf("warning listener saw %q", got)
	}
}

func TestFeatures(t *testing.T) {
	e := setup(t)

	expect(t, e, `typeof process.features.inspector + ':' + typeof process.features.tls + ':' + typeof process.features.typescript`, "boolean:boolean:string")
	expect(t, e, `Object.getOwnPropertyDescriptor(process.versions, 'node').writable`, "false")
	expect(t, e, `process.config.variables.napi_build_version === process.versions.napi`, "true")
}
