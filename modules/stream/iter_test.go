package stream

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
	"godejs/modules/web"
)

// iterSetup installs the classic stream module (for the interop constructors)
// plus the three iterable-stream modules under test.
func iterSetup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	web.Install(e)
	jstest.Install(t, e, "buffer", buffer.Install)
	jstest.Install(t, e, "stream", Install)
	jstest.Install(t, e, "iter", InstallIter)
	jstest.Install(t, e, "consumers", InstallConsumers)
	jstest.Install(t, e, "web", InstallWeb)
	return e
}

func TestIterFromSyncAndSyncConsumers(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `iter.textSync(iter.fromSync('hello'))`), "hello")
	jstest.Equal(t, jstest.Str(t, e,
		`Buffer.from(iter.bytesSync(iter.fromSync([Buffer.from('ab'), Buffer.from('cd')]))).toString()`), "abcd")
	jstest.Equal(t, jstest.Str(t, e, `String(iter.arraySync(iter.fromSync('abc')).length)`), "1")
	jstest.Equal(t, jstest.Str(t, e,
		`Buffer.from(iter.arraySync(iter.fromSync('abc'))[0]).toString()`), "abc")
	jstest.Equal(t, jstest.Str(t, e,
		`Buffer.from(iter.arrayBufferSync(iter.fromSync('hi'))).toString()`), "hi")
	jstest.Equal(t, jstest.Str(t, e, `String(iter.bytesSync(iter.fromSync('hello'), { limit: 5 }).length)`), "5")
}

// A source of many small chunks is coalesced into bounded batches.
func TestIterSyncBatching(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var chunks = [];
		for (var i = 0; i < 300; i++) chunks.push(Buffer.from('a'));
		var batches = 0;
		for (var batch of iter.fromSync(chunks)) batches++;
		return String(batches);
	})()`), "3")
}

func TestIterPullSync(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e,
		`iter.textSync(iter.pullSync('hello', function (chunks) { return chunks; }))`), "hello")

	// A stateful transform object receives the source iterable (with a null
	// flush sentinel) and yields batches of its own.
	jstest.Equal(t, jstest.Str(t, e, `iter.textSync(iter.pullSync('hello', {
		transform: function* (source) {
			for (var chunks of source) {
				if (chunks === null) continue;
				var out = [];
				for (var i = 0; i < chunks.length; i++) {
					out.push(Buffer.from(Buffer.from(chunks[i]).toString().toUpperCase()));
				}
				yield out;
			}
		}
	}))`), "HELLO")
}

func TestIterPipeToSync(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var parts = [];
		var writer = {
			writeSync: function (chunk) {
				parts.push(Buffer.from(chunk).toString());
				return true;
			},
			endSync: function () { return 0; }
		};
		var written = iter.pipeToSync('hello', writer);
		return parts.join('') + ':' + written;
	})()`), "hello:5")
}

func TestIterPushDeliversBatches(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var pair = iter.push();
		var seen = [];
		var iterator = pair.readable[Symbol.asyncIterator]();
		pair.writer.write('ab');
		pair.writer.end('cd');
		iterator.next().then(function (first) {
			seen.push(Buffer.from(first.value[0]).toString());
			return iterator.next();
		}).then(function (second) {
			seen.push(Buffer.from(second.value[0]).toString());
			return iterator.next();
		}).then(function (done) {
			seen.push(String(done.done));
			globalThis.__iterPush = seen.join(':');
		});
		return 'queued';
	})()`), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__iterPush`), "ab:cd:true")
}

func TestIterDuplexChannels(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var channels = iter.duplex();
		return channels.length + ':' + (typeof channels[0].writer.write) + ':' +
			(typeof channels[1].readable[Symbol.asyncIterator]);
	})()`), "2:function:function")
}

func TestIterShareAndBroadcast(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var shared = iter.share('hello');
		var iterator = shared[Symbol.asyncIterator]();
		return typeof iterator.next;
	})()`), "function")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var cast = iter.broadcast('hello');
		return typeof cast[Symbol.asyncIterator];
	})()`), "function")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var syncShared = iter.shareSync('hello');
		var out = '';
		for (var batch of syncShared) {
			for (var i = 0; i < batch.length; i++) out += Buffer.from(batch[i]).toString();
		}
		return out;
	})()`), "hello")
}

func TestIterValidation(t *testing.T) {
	e := iterSetup(t)

	if err := jstest.Throws(t, e, `iter.from(42)`); err.Code != "ERR_INVALID_ARG_TYPE" {
		t.Errorf("iter.from(42) code = %q", err.Code)
	}
	if err := jstest.Throws(t, e, `iter.fromSync({})`); err.Code != "ERR_INVALID_ARG_TYPE" {
		t.Errorf("iter.fromSync({}) code = %q", err.Code)
	}
	if err := jstest.Throws(t, e, `iter.pullSync('x', 7)`); err.Code != "ERR_INVALID_ARG_TYPE" {
		t.Errorf("iter.pullSync('x', 7) code = %q", err.Code)
	}
	if err := jstest.Throws(t, e, `iter.pipeToSync('x')`); err.Code == "" {
		t.Error("iter.pipeToSync('x') should have a Node error code")
	}
}

func TestIterExportSurface(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `Object.isFrozen(iter.Stream) ? 'frozen' : 'open'`), "frozen")
	jstest.Equal(t, jstest.Str(t, e,
		`iter.toStreamable === Symbol.for('Stream.toStreamable') ? 'same' : 'different'`), "same")

	names := []string{
		"from", "fromSync", "pull", "pullSync", "pipeTo", "pipeToSync",
		"bytes", "text", "arrayBuffer", "array",
		"bytesSync", "textSync", "arrayBufferSync", "arraySync",
		"merge", "broadcast", "share", "shareSync", "tap", "tapSync", "ondrain",
		"fromReadable", "fromWritable", "toReadable", "toReadableSync", "toWritable",
		"duplex", "push",
	}
	for _, name := range names {
		if got := jstest.Str(t, e, `typeof iter.`+name); got != "function" {
			t.Errorf("iter.%s = %s, want function", name, got)
		}
	}
	for _, name := range []string{"pull", "push", "duplex", "broadcast", "tap"} {
		if got := jstest.Str(t, e, `typeof iter.Stream.`+name); got != "function" {
			t.Errorf("iter.Stream.%s = %s, want function", name, got)
		}
	}
}

func TestIterConsumersAndWebExports(t *testing.T) {
	e := iterSetup(t)

	for _, name := range []string{"arrayBuffer", "blob", "buffer", "bytes", "text", "json"} {
		if got := jstest.Str(t, e, `typeof consumers.`+name); got != "function" {
			t.Errorf("consumers.%s = %s, want function", name, got)
		}
	}
	for _, name := range []string{"ReadableStream", "WritableStream", "TransformStream",
		"ByteLengthQueuingStrategy", "CountQueuingStrategy"} {
		if got := jstest.Str(t, e, `typeof web.`+name); got != "function" {
			t.Errorf("web.%s = %s, want function", name, got)
		}
	}
}

// stream/consumers works on any async iterable of chunks, not just Readables.
func TestConsumersOverAsyncIterable(t *testing.T) {
	e := iterSetup(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var source = {};
		source[Symbol.asyncIterator] = async function* () {
			yield 'ab';
			yield Buffer.from('cd');
		};
		var result = '';
		consumers.text(source).then(function (value) { globalThis.__consumerText = value; });
		return 'pending';
	})()`), "pending")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__consumerText`), "abcd")
}
