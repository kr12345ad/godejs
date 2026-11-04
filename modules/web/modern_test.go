package web

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	_ "godejs/modules/events"
)

func modernSetup(t *testing.T) *env.Env {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := jstest.New(t)
	// The events module installs the global Event class; MessageEvent extends
	// it, exactly as in Node where Event is present before user code runs.
	e.Exports("events")
	Install(e)
	return e
}

func TestModernGlobals(t *testing.T) {
	e := modernSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `[
	  typeof MessageEvent,
	  typeof DisposableStack, typeof AsyncDisposableStack,
	  typeof TextEncoderStream, typeof TextDecoderStream,
	  typeof CompressionStream, typeof DecompressionStream,
	  performance.constructor.name,
	  Object.prototype.toString.call(performance)
	].join(',')`),
		"function,function,function,function,function,function,function,Performance,[object Performance]")
}

func TestMessageEvent(t *testing.T) {
	e := modernSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
	  var event = new MessageEvent('greeting', { data: 42, origin: 'https://a.test' });
	  return [
	    event.type, event.data, event.origin, event.lastEventId,
	    event.source === null, event.ports.length, event instanceof Event
	  ].join(',');
	})()`), "greeting,42,https://a.test,,true,0,true")
}

func TestDisposableStack(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
	  var order = [];
	  var stack = new DisposableStack();
	  stack.defer(function () { order.push('a'); });
	  stack.adopt('v', function (value) { order.push(value); });
	  stack.dispose();
	  var afterDispose = stack.disposed;
	  var moved = new DisposableStack();
	  moved.defer(function () { order.push('m'); });
	  var target = moved.move();
	  var sourceDisposed = moved.disposed;
	  target.dispose();
	  return JSON.stringify([order.join(''), afterDispose, sourceDisposed, target.disposed]);
	})()`), `["vam",true,true,true]`)

	// A throwing disposer is surfaced, and the rest of the stack still runs.
	jstest.Equal(t, jstest.Str(t, e, `(function () {
	  var ran = false;
	  var stack = new DisposableStack();
	  stack.defer(function () { ran = true; });
	  stack.defer(function () { throw new Error('boom'); });
	  try { stack.dispose(); } catch (error) { return error.message + ':' + ran; }
	  return 'no-throw';
	})()`), "boom:true")
}

func TestAsyncDisposableStack(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	jstest.Str(t, e, `(function () {
	  globalThis.__asyncDispose = 'pending';
	  (async function () {
	    var order = [];
	    var stack = new AsyncDisposableStack();
	    stack.defer(async function () { order.push('a'); });
	    stack.defer(function () { order.push('s'); });
	    await stack.disposeAsync();
	    globalThis.__asyncDispose = order.join('') + ':' + stack.disposed;
	  })();
	  return 'scheduled';
	})()`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__asyncDispose`), "sa:true")
}
