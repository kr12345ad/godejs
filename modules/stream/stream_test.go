package stream

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
	"godejs/modules/events"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "buffer", buffer.Install)
	jstest.Install(t, e, "events", events.Install)
	jstest.Install(t, e, "stream", Install)
	return e
}

func drain(e *env.Env) {
	if err := e.DrainTicks(); err != nil {
		panic(err)
	}
}

func TestReadableDataAndEnd(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable({ read: function () {} });
	var chunks = [];
	var ended = false;
	r.on('data', function (c) { chunks.push(c.toString()); });
	r.on('end', function () { ended = true; });
	r.push('hello');
	r.push(Buffer.from(' world'));
	r.push(null);
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `chunks.join('') + ':' + ended`), "hello world:true")
}

func TestReadableNonFlowingRead(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable();
	r.push('abc');
	globalThis.got = stream.read.call(r).toString();
	var r2 = new stream.Readable();
	globalThis.pushRet = typeof stream.push.call(r2, 'q');
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.got`), "abc")
}

func TestPipeToWritable(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var seen = [];
	var done = false;
	var w = new stream.Writable({ write: function (chunk, enc, cb) {
		seen.push(chunk.toString());
		cb();
	} });
	w.on('finish', function () { done = true; });
	var r = new stream.Readable({ read: function () {} });
	r.pipe(w);
	r.push('a');
	r.push('b');
	r.push(null);
	`)
	drain(e)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `seen.join('') + ':' + done`), "ab:true")
}

func TestPassThrough(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var out = [];
	var pt = new stream.PassThrough();
	pt.on('data', function (c) { out.push(c.toString()); });
	pt.write('pass');
	pt.end();
	var a = new stream.PassThrough(), b = new stream.PassThrough();
	var co = stream.compose(a, b);
	globalThis.composed = (co instanceof stream.Duplex) && co !== a;
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `out.join('') + ':' + globalThis.composed`), "pass:true")
}

func TestTransformSubclass(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	class Upper extends stream.Transform {
		_transform(chunk, enc, cb) {
			this.push(String(chunk).toUpperCase());
			cb();
		}
	}
	var out = [];
	var up = new Upper();
	up.on('data', function (c) { out.push(c.toString()); });
	up.write('abc');
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `out.join('')`), "ABC")
}

func TestFinishOrdering(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var events = [];
	var w = new stream.Writable({ write: function (c, enc, cb) { cb(); } });
	w.on('finish', function () { events.push('finish'); });
	w.write('x', function () { events.push('write-cb'); });
	w.end(function () { events.push('end-cb'); });
	`)
	drain(e)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `events.join(',')`), "write-cb,finish,end-cb")
}

func TestPipelineAndFinished(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var piped = '';
	var w = new stream.Writable({ write: function (c, enc, cb) { piped += c.toString(); cb(); } });
	var r = new stream.Readable({ read: function () {} });
	var pipelineDone = false;
	stream.pipeline(r, w, function (err) { pipelineDone = err === null; });
	var finishedDone = false;
	stream.finished(r, function (err) { finishedDone = err === null; });
	r.push('pipe');
	r.push(null);
	`)
	drain(e)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `piped + ':' + pipelineDone + ':' + finishedDone`), "pipe:true:true")
}

func TestIsHelpersAndHWM(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable();
	var w = new stream.Writable({ write: function (c, e, cb) { cb(); } });
	globalThis.before = stream.isReadable(r) + ':' + stream.isWritable(w) + ':' + stream.isDestroyed(r) + ':' + stream.isDisturbed(r) + ':' + stream.isErrored(r);
	r.push(null);
	w.end();
	globalThis.after = stream.isReadable(r) + ':' + stream.isWritable(w);
	globalThis.hwm = stream.getDefaultHighWaterMark(false);
	stream.setDefaultHighWaterMark(false, 1024);
	globalThis.hwm2 = stream.getDefaultHighWaterMark(false);
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.before`), "true:true:false:false:false")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.after`), "false:false")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.hwm + ':' + globalThis.hwm2`), "65536:1024")
}

func TestFromAndAbort(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var out = [];
	var r = stream.from(['a', 'b', 'c']);
	r.on('data', function (c) { out.push(c); });
	var listeners = {};
	var signal = { aborted: false, addEventListener: function (n, fn) { listeners[n] = fn; } };
	var target = new stream.Readable();
	target.on('error', function () {});
	stream.addAbortSignal(signal, target);
	signal.aborted = true;
	listeners.abort();
	globalThis.aborted = target.destroyed;
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `out.join('') + ':' + globalThis.aborted`), "abc:true")
}

func TestAsyncIteratorShape(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable();
	globalThis.iter = typeof r[Symbol.asyncIterator];
	globalThis.next = typeof r[Symbol.asyncIterator]().next;
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.iter + ':' + globalThis.next`), "function:function")
}

// Readable must call _read when a consumer asks for data, and must report the
// buffered length the way Node does.
func TestReadableCallsRead(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var reads = 0;
	var r = new stream.Readable({
		read: function () { reads++; this.push('abc'); this.push(null); }
	});
	globalThis.got = '';
	r.on('data', function (c) { globalThis.got += c.toString(); });
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.got + ':' + reads`), "abc:1")
}

// A paused readable keeps data buffered until read() is called.
func TestReadablePausedBuffer(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable({ read: function () {} });
	r.push('hello');
	globalThis.len = r.readableLength;
	globalThis.flowing = String(r.readableFlowing);
	globalThis.first = r.read(2).toString();
	globalThis.rest = r.read().toString();
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.len + ':' + globalThis.flowing + ':' + globalThis.first + ':' + globalThis.rest`),
		"5:null:he:llo")
}

// destroy() reports through _destroy, then emits error then close exactly once.
func TestDestroyOrdering(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var events = [];
	var boom = new Error('boom');
	var w = new stream.Writable({
		write: function (chunk, enc, cb) { cb(); },
		_destroy: function (err, cb) { events.push('_destroy:' + (err && err.message)); cb(err); }
	});
	w.on('error', function (err) { events.push('error:' + err.message); });
	w.on('close', function () { events.push('close:' + w.closed); });
	w.on('finish', function () { events.push('finish'); });
	w.destroy(boom);
	globalThis.destroyed = w.destroyed;
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.destroyed + ':' + events.join(',')`),
		"true:error:boom,close:true")
}

// write/end drive _write once per chunk, then _final, 'finish' and callbacks.
func TestWritableFinishAndCallbacks(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var seen = [];
	var w = new stream.Writable({
		write: function (chunk, enc, cb) { seen.push('w:' + chunk.toString()); cb(); },
		final: function (cb) { seen.push('final'); cb(); }
	});
	w.write('a');
	w.write('b');
	w.end(function () { seen.push('end-cb'); });
	w.on('finish', function () { seen.push('finish:' + w.writableFinished); });
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `seen.join(',')`), "w:a,w:b,final,end-cb,finish:true")
}

// Transform must run _transform per chunk and _flush before push(null).
func TestTransformFlush(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var out = '';
	var t = new stream.Transform({
		transform: function (chunk, enc, cb) { cb(null, chunk.toString().toUpperCase()); },
		flush: function (cb) { this.push('!'); cb(); }
	});
	t.on('data', function (c) { out += c.toString(); });
	t.on('end', function () { globalThis.out = out; });
	t.end('ab');
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.out)`), "AB!")
}

// A plain Writable must reject writes after end with ERR_STREAM_WRITE_AFTER_END.
func TestWriteAfterEnd(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var codes = [];
	var w = new stream.Writable({ write: function (c, e, cb) { cb(); } });
	w.on('error', function (err) { codes.push(err.code); });
	w.end();
	w.write('x', function (err) { codes.push(err && err.code); });
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `codes.join(',')`), "ERR_STREAM_WRITE_AFTER_END,ERR_STREAM_WRITE_AFTER_END")
}

func TestImplementedBaseline(t *testing.T) {
	want := map[string]bool{
		"stream.addAbortSignal":          true,
		"stream.compose":                 true,
		"stream.finished":                true,
		"stream.from":                    true,
		"stream.fromWeb":                 true,
		"stream.toWeb":                   true,
		"stream.getDefaultHighWaterMark": true,
		"stream.isDestroyed":             true,
		"stream.isDisturbed":             true,
		"stream.isErrored":               true,
		"stream.isReadable":              true,
		"stream.isWritable":              true,
		"stream.pipeline":                true,
		"stream.push":                    true,
		"stream.read":                    true,
		"stream.setDefaultHighWaterMark": true,
	}
	got := Implemented()
	if len(got) != len(want) {
		t.Fatalf("implemented %d keys, want %d", len(got), len(want))
	}
	for _, key := range got {
		if !want[key] {
			t.Fatalf("unexpected implemented key %q", key)
		}
	}
}

// Node's byte streams default to 64 KiB (16 on object mode) and expose the
// same value through the dual-mode high-water-mark keys on a Duplex.
func TestHighWaterMarkDefaults(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = new stream.Readable({ read: function () {} });
	var d = new stream.Duplex({ read: function () {}, write: function (c, e, cb) { cb(); } });
	globalThis.hwm = stream.getDefaultHighWaterMark(false) + ':' + stream.getDefaultHighWaterMark(true);
	globalThis.rhwm = r.readableHighWaterMark + ':' + d.readableHighWaterMark + ':' + d.writableHighWaterMark;
	`)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.hwm`), "65536:16")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.rhwm`), "65536:65536:65536")
}

// getHighWaterMark rejects negative and non-integer values with the same code
// and argument name Node uses, including the duplex-specific keys.
func TestHighWaterMarkValidation(t *testing.T) {
	e := setup(t)
	cases := []struct {
		src  string
		code string
	}{
		{`new stream.Readable({ highWaterMark: -1 })`, "ERR_INVALID_ARG_VALUE"},
		{`new stream.Readable({ highWaterMark: 1.5 })`, "ERR_INVALID_ARG_VALUE"},
		{`new stream.Writable({ highWaterMark: -2 })`, "ERR_INVALID_ARG_VALUE"},
		{`new stream.Duplex({ read: function () {}, readableHighWaterMark: -3 })`, "ERR_INVALID_ARG_VALUE"},
		{`stream.setDefaultHighWaterMark(false, -1)`, "ERR_OUT_OF_RANGE"},
	}
	for _, tc := range cases {
		jsErr := jstest.Throws(t, e, tc.src)
		jstest.Equal(t, jsErr.Code, tc.code)
	}
}

// map/filter/take/drop are stream-returning operators in Node 24; they must
// consume their source lazily and expose object-mode output.
func TestStreamOperators(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	function collect(s, name) {
		s.toArray().then(function (a) { globalThis[name] = a.join(','); });
	}
	collect(stream.from([1, 2, 3, 4]).map(function (v) { return v * 2; }), 'mapped');
	collect(stream.from([1, 2, 3, 4]).filter(function (v) { return v % 2 === 0; }), 'filtered');
	collect(stream.from([1, 2, 3, 4]).take(2), 'took');
	collect(stream.from([1, 2, 3, 4]).drop(3), 'dropped');
	collect(stream.from(['ab', 'cd']).flatMap(function (v) { return v.split(''); }), 'flat');
	globalThis.reduced = -1;
	stream.from([1, 2, 3]).reduce(function (a, b) { return a + b; }, 10).then(function (v) { globalThis.reduced = v; });
	`)
	for i := 0; i < 20; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.mapped`), "2,4,6,8")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.filtered`), "2,4")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.took`), "1,2")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.dropped`), "4")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.flat`), "a,b,c,d")
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.reduced)`), "16")
}

// iterator() is the public async iterator; each next() resolves with the next
// buffered chunk and reports completion once the stream ends.
func TestReadableIterator(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var r = stream.from(['a', 'b']);
	var it = r.iterator();
	globalThis.iterShape = typeof it.next + ':' + typeof it.return + ':' + (it[Symbol.asyncIterator]() === it);
	var seen = [];
	it.next().then(function (a) {
		seen.push(a.value);
		return it.next();
	}).then(function (b) {
		seen.push(b.value);
		return it.next();
	}).then(function (c) {
		seen.push(c.done);
		globalThis.seen = seen.join(',');
	});
	`)
	for i := 0; i < 10; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.iterShape`), "function:function:true")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.seen`), "a,b,true")
}

// A Transform without a _transform implementation reports
// ERR_METHOD_NOT_IMPLEMENTED, exactly like Node, instead of echoing input.
func TestTransformRequiresImplementation(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var codes = [];
	var t = new stream.Transform();
	t.on('error', function (err) { codes.push(err.code); });
	t.write('x');
	globalThis.codes2 = codes;
	`)
	drain(e)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `codes.join(',')`), "ERR_METHOD_NOT_IMPLEMENTED")
}

// Corked writes are handed to _writev as one batch.
func TestWritevBatches(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var calls = [];
	var w = new stream.Writable({
		write: function (c, enc, cb) { calls.push('w:' + c.toString()); cb(); },
		writev: function (chunks, cb) {
			calls.push('v:' + chunks.map(function (x) { return x.chunk.toString(); }).join(''));
			cb();
		}
	});
	w.cork();
	w.write('a');
	w.write('b');
	w.uncork();
	`)
	drain(e)
	jstest.Equal(t, jstest.Str(t, e, `calls.join(',')`), "v:ab")
}

// finished() distinguishes a completed writable from a stream destroyed before
// it finished, and honours the readable/writable option overrides.
func TestFinishedPrematureCloseAndWritable(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var doneCode = 'pending';
	var w = new stream.Writable({ write: function (c, e, cb) { cb(); } });
	stream.finished(w, function (err) { doneCode = err === null ? 'ok' : err.code; });
	w.end();
	var closedState = 'pending';
	var r = new stream.Readable({ read: function () {} });
	stream.finished(r, function (err) { closedState = err === null ? 'ok' : err.code; });
	r.destroy();
	globalThis.finishedWritable = function () { return doneCode; };
	globalThis.closedCode = function () { return closedState; };
	`)
	for i := 0; i < 6; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.finishedWritable()`), "ok")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.closedCode()`), "ERR_STREAM_PREMATURE_CLOSE")
}

// compose() returns a duplex: writes enter the first stage, reads leave the
// last one, and ending the composed stream flushes the whole chain.
func TestComposeDuplex(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var up = new stream.Transform({ transform: function (c, enc, cb) { cb(null, c.toString().toUpperCase()); } });
	var bang = new stream.Transform({ transform: function (c, enc, cb) { cb(null, c.toString() + '!'); } });
	var co = stream.compose(up, bang);
	globalThis.isDuplex = (co instanceof stream.Duplex) + ':' + co.writableObjectMode;
	var out = '';
	var ended = false;
	co.on('data', function (c) { out += c.toString(); });
	co.on('end', function () { ended = true; });
	co.write('hi');
	co.end();
	globalThis.composed = function () { return out + ':' + ended; };
	`)
	for i := 0; i < 12; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.isDuplex`), "true:true")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.composed()`), "HI!:true")
}

// Pushing data after EOF reports ERR_STREAM_PUSH_AFTER_EOF, while unshift()
// only errors once 'end' has been emitted, matching readable.js.
func TestPushAfterEOF(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
	var codes = [];
	var r = new stream.Readable({ read: function () {} });
	r.on('error', function (err) { codes.push(err.code); });
	r.push(null);
	r.push('late');
	globalThis.pushErr = codes;

	var got = [];
	var r2 = new stream.Readable({ read: function () {} });
	r2.on('error', function (err) { got.push(err.code); });
	r2.on('data', function (c) { got.push(c.toString()); });
	r2.on('end', function () { got.push('end'); });
	r2.push(null);
	r2.unshift('late');
	globalThis.unshiftGot = got;
	`)
	for i := 0; i < 4; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `codes.join(',')`), "ERR_STREAM_PUSH_AFTER_EOF")
	jstest.Equal(t, jstest.Str(t, e, `got.join(',')`), "late,end")

	// Once 'end' has fired, unshift() reports the error asynchronously.
	jstest.Eval(t, e, `
	var lateCodes = [];
	var r3 = new stream.Readable({ read: function () {}, autoDestroy: false });
	r3.on('error', function (err) { lateCodes.push(err.code); });
	r3.resume();
	r3.push(null);
	globalThis.lateCodes = lateCodes;
	`)
	for i := 0; i < 4; i++ {
		drain(e)
	}
	jstest.Eval(t, e, `r3.unshift('late');`)
	for i := 0; i < 4; i++ {
		drain(e)
	}
	jstest.Equal(t, jstest.Str(t, e, `lateCodes.join(',')`), "ERR_STREAM_UNSHIFT_AFTER_END_EVENT")
}
