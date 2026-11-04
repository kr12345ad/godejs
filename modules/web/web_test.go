package web

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	Install(e)
	return e
}

func TestGlobalsInstalled(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `[
	  typeof TextEncoder, typeof TextDecoder, typeof atob, typeof btoa,
	  typeof queueMicrotask, typeof structuredClone, typeof performance,
	  typeof AbortController, typeof AbortSignal
	].join(',')`)
	want := "function,function,function,function,function,function,object,function,function"
	jstest.Equal(t, got, want)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__godeWeb === undefined ? 'clean' : 'leaked'`), "clean")
}

func TestTextEncoder(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `new TextEncoder().encoding`), "utf-8")
	jstest.Equal(t, jstest.Str(t, e, `Array.from(new TextEncoder().encode('é☃')).join(',')`), "195,169,226,152,131")
	jstest.Equal(t, jstest.Str(t, e, `Array.from(new TextEncoder().encode('\uD800')).join(',')`), "239,191,189")
	jstest.Equal(t, jstest.Str(t, e, `new TextEncoder().encode().length`), "0")
	jstest.Equal(t, jstest.Str(t, e, `Array.from(new TextEncoder().encode(null)).join(',')`), "110,117,108,108")
}

func TestTextEncoderEncodeInto(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `
	  var d = new Uint8Array(4);
	  var r = new TextEncoder().encodeInto('é☃', d);
	  JSON.stringify([r.read, r.written, Array.from(d)])
	`)
	jstest.Equal(t, got, `[1,2,[195,169,0,0]]`)

	got = jstest.Str(t, e, `
	  var d = new Uint8Array(2);
	  var r = new TextEncoder().encodeInto('☃', d);
	  JSON.stringify([r.read, r.written, Array.from(d)])
	`)
	jstest.Equal(t, got, `[0,0,[0,0]]`)

	jsErr := jstest.Throws(t, e, `new TextEncoder().encodeInto('x', {})`)
	jstest.Equal(t, jsErr.Name, "TypeError")
}

func TestTextDecoderUTF8(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().encoding`), "utf-8")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('UTF8').encoding`), "utf-8")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new Uint8Array([195,169]))`), "é")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode()`), "")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new DataView(new Uint8Array([104,105]).buffer))`), "hi")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new Uint8Array([]).buffer)`), "")
	// An incomplete trailing sequence is one replacement character.
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new Uint8Array([0xE2,0x82]))`), "\uFFFD")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new Uint8Array([0xC0,0x80]))`), "\uFFFD\uFFFD")
}

func TestTextDecoderStreaming(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.dec = new TextDecoder();
	  globalThis.p1 = dec.decode(new Uint8Array([0xE2]), { stream: true });
	  globalThis.p2 = dec.decode(new Uint8Array([0x82, 0xAC]));
	  p1 + '|' + p2
	`), "|€")

	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.u = new TextDecoder('utf-16le');
	  globalThis.q1 = u.decode(new Uint8Array([0x61]), { stream: true });
	  globalThis.q2 = u.decode(new Uint8Array([0x00, 0x62, 0x00]));
	  q1 + '|' + q2
	`), "|ab")
}

func TestTextDecoderBOM(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder().decode(new Uint8Array([0xEF,0xBB,0xBF,0x61]))`), "a")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('utf-8', { ignoreBOM: true }).decode(new Uint8Array([0xEF,0xBB,0xBF,0x61])).charCodeAt(0)`), "65279")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('utf-16le').decode(new Uint8Array([0xFF,0xFE,0x61,0x00]))`), "a")
	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.bomDec = new TextDecoder();
	  globalThis.r1 = bomDec.decode(new Uint8Array([0xEF]), { stream: true });
	  globalThis.r2 = bomDec.decode(new Uint8Array([0xBB, 0xBF, 0x61]));
	  r1 + '|' + r2
	`), "|a")
}

func TestTextDecoderFatal(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `
	  try {
	    new TextDecoder('utf-8', { fatal: true }).decode(new Uint8Array([0xFF]));
	    'no error';
	  } catch (err) {
	    err.name + ':' + err.code;
	  }
	`)
	jstest.Equal(t, got, "TypeError:ERR_ENCODING_INVALID_ENCODED_DATA")

	got = jstest.Str(t, e, `
	  try {
	    new TextDecoder('utf-8', { fatal: true }).decode(new Uint8Array([0xE2, 0x82]));
	    'no error';
	  } catch (err) {
	    err.name + ':' + err.code;
	  }
	`)
	jstest.Equal(t, got, "TypeError:ERR_ENCODING_INVALID_ENCODED_DATA")

	got = jstest.Str(t, e, `
	  try {
	    new TextDecoder('nope');
	    'no error';
	  } catch (err) {
	    err.name + ':' + err.code;
	  }
	`)
	jstest.Equal(t, got, "RangeError:ERR_ENCODING_NOT_SUPPORTED")
}

func TestTextDecoderOtherEncodings(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('latin1').encoding`), "windows-1252")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('ascii').encoding`), "windows-1252")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('ucs-2').encoding`), "utf-16le")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('latin1').decode(new Uint8Array([0x80]))`), "€")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('latin1').decode(new Uint8Array([0xE9]))`), "é")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('utf-16le').decode(new Uint8Array([0x61,0x00,0x62,0x00]))`), "ab")
	jstest.Equal(t, jstest.Str(t, e, `new TextDecoder('utf-16be').decode(new Uint8Array([0x00,0x61,0x00,0x62]))`), "ab")
}

func TestAtobBtoa(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `btoa('hello')`), "aGVsbG8=")
	jstest.Equal(t, jstest.Str(t, e, `atob('aGVsbG8=')`), "hello")
	jstest.Equal(t, jstest.Str(t, e, `atob(' a G V s b G 8 = ')`), "hello")
	jstest.Equal(t, jstest.Str(t, e, `atob('aGVsbG8')`), "hello")
	jstest.Equal(t, jstest.Str(t, e, `btoa(123)`), "MTIz")
	jstest.Equal(t, jstest.Str(t, e, `atob(btoa(String.fromCharCode(255))).charCodeAt(0)`), "255")

	// Latin-1 characters are allowed; only code units above 0xFF are rejected.
	jstest.Equal(t, jstest.Str(t, e, `btoa('é')`), "6Q==")
	for _, src := range []string{`atob('!')`, `btoa('€')`, `atob('abcde')`} {
		jsErr := jstest.Throws(t, e, src)
		jstest.Equal(t, jsErr.Name, "InvalidCharacterError")
	}
}

func TestQueueMicrotask(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	  globalThis.order = [];
	  queueMicrotask(function () { order.push('micro'); });
	  order.push('sync');
	`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("RunMain: %v", err)
	}
	// Give the engine one more entry point so a pending microtask is drained
	// even on builds that defer it past the first evaluation.
	jstest.Eval(t, e, `0`)
	jstest.Equal(t, jstest.Str(t, e, `order.join(',')`), "sync,micro")

	jsErr := jstest.Throws(t, e, `queueMicrotask(1)`)
	jstest.Equal(t, jsErr.Name, "TypeError")
}

func TestStructuredClone(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
	  var o = { a: 1, b: { c: 2 }, d: [1, 2] };
	  var c = structuredClone(o);
	  [c !== o, c.b !== o.b, JSON.stringify(c) === JSON.stringify(o)].join(',')
	`), "true,true,true")

	jstest.Equal(t, jstest.Str(t, e, `
	  var o = {}; o.self = o;
	  var c = structuredClone(o);
	  c.self === c
	`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
	  var d = new Date(0);
	  var c = structuredClone(d);
	  [c !== d, c.getTime()].join(',')
	`), "true,0")

	jstest.Equal(t, jstest.Str(t, e, `
	  var m = new Map([['a', { x: 1 }]]);
	  var cm = structuredClone(m);
	  [cm !== m, cm.get('a') !== m.get('a'), cm.get('a').x].join(',')
	`), "true,true,1")

	jstest.Equal(t, jstest.Str(t, e, `
	  var s = new Set([1, 2]);
	  var cs = structuredClone(s);
	  [cs !== s, cs.has(1), cs.has(2)].join(',')
	`), "true,true,true")

	jstest.Equal(t, jstest.Str(t, e, `
	  var a = new Uint8Array([1, 2, 3]);
	  var ca = structuredClone(a);
	  [ca !== a, Array.from(ca).join(',')].join('|')
	`), "true|1,2,3")

	jstest.Equal(t, jstest.Str(t, e, `
	  var r = /ab+c/gi;
	  var cr = structuredClone(r);
	  [cr !== r, cr.source, cr.flags].join('|')
	`), "true|ab+c|gi")

	if jsErr := jstest.Throws(t, e, `structuredClone(function f() {})`); jsErr.Name != "DataCloneError" {
		t.Fatalf("function clone error = %q, want DataCloneError", jsErr.Name)
	}
	if jsErr := jstest.Throws(t, e, `structuredClone(Symbol('x'))`); jsErr.Name != "DataCloneError" {
		t.Fatalf("symbol clone error = %q, want DataCloneError", jsErr.Name)
	}
}

func TestPerformance(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof performance.now()`), "number")
	jstest.Equal(t, jstest.Str(t, e, `typeof performance.timeOrigin`), "number")
	jstest.Equal(t, jstest.Str(t, e, `
	  var a = performance.now();
	  var b = performance.now();
	  b >= a
	`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
	  performance.mark('start');
	  performance.measure('span', 'start');
	  var entry = performance.getEntriesByName('span')[0];
	  [entry.entryType, entry.duration >= 0, performance.getEntriesByType('mark').length >= 1].join(',')
	`), "measure,true,true")

	jstest.Equal(t, jstest.Str(t, e, `
	  performance.clearMarks('start');
	  performance.clearMeasures('span');
	  [performance.getEntriesByName('start').length, performance.getEntriesByName('span').length].join(',')
	`), "0,0")
}

func TestAbortController(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.c = new AbortController();
	  globalThis.fired = 0;
	  c.signal.addEventListener('abort', function () { fired++; });
	  var before = c.signal.aborted;
	  c.abort();
	  c.abort();
	  [before, c.signal.aborted, fired, c.signal.reason.name].join(',')
	`), "false,true,1,AbortError")

	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.reason = { why: 'stop' };
	  globalThis.c2 = new AbortController();
	  c2.abort(reason);
	  c2.signal.reason === reason
	`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.c3 = new AbortController();
	  var caught;
	  try { c3.signal.throwIfAborted(); } catch (err) { caught = err; }
	  var before = caught === undefined;
	  c3.abort('nope');
	  try { c3.signal.throwIfAborted(); } catch (err) { caught = err; }
	  [before, caught].join(',')
	`), "true,nope")

	jstest.Equal(t, jstest.Str(t, e, `
	  var s = AbortSignal.abort();
	  [s.aborted, s.reason.name].join(',')
	`), "true,AbortError")
	jstest.Equal(t, jstest.Str(t, e, `AbortSignal.abort('why').reason`), "why")
}

func TestAbortSignalAnyAndTimeout(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.a = new AbortController();
	  globalThis.b = new AbortController();
	  globalThis.any = AbortSignal.any([a.signal, b.signal]);
	  var before = any.aborted;
	  b.abort('b-reason');
	  [before, any.aborted, any.reason].join(',')
	`), "false,true,b-reason")

	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.already = AbortSignal.abort('early');
	  var immediate = AbortSignal.any([new AbortController().signal, already]);
	  [immediate.aborted, immediate.reason].join(',')
	`), "true,early")

	jstest.Equal(t, jstest.Str(t, e, `
	  globalThis.to = AbortSignal.timeout(1);
	  globalThis.to.aborted
	`), "false")
	if err := e.RunMain(); err != nil {
		t.Fatalf("RunMain: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `[to.aborted, to.reason.name].join(',')`), "true,TimeoutError")

	jsErr := jstest.Throws(t, e, `AbortSignal.timeout(-1)`)
	jstest.Equal(t, jsErr.Name, "RangeError")
}
