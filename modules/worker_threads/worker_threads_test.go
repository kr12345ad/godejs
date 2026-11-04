package worker_threads

import (
	"os"
	"path/filepath"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "worker_threads", Install)
	return e
}

// writeWorker drops a worker script into the test's working directory and
// returns its absolute path.
func writeWorker(t *testing.T, e *env.Env, src string) string {
	t.Helper()
	path := filepath.Join(e.Cwd(), "worker.js")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write worker: %v", err)
	}
	return path
}

// jsLiteral quotes s as a JavaScript string literal.
func jsLiteral(s string) string {
	out := []byte{'"'}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			out = append(out, '\\', '\\')
		case '"':
			out = append(out, '\\', '"')
		case '\n':
			out = append(out, '\\', 'n')
		default:
			out = append(out, s[i])
		}
	}
	return string(append(out, '"'))
}

func TestIsMainThreadAndSurface(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `String(worker_threads.isMainThread)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(worker_threads.threadId)`), "0")
	jstest.Equal(t, jstest.Str(t, e, `String(worker_threads.parentPort)`), "null")
	jstest.Equal(t, jstest.Str(t, e, `String(worker_threads.workerData)`), "null")
	jstest.Equal(t, jstest.Str(t, e, `typeof worker_threads.Worker`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof worker_threads.MessageChannel`), "function")
	jstest.Equal(t, jstest.Str(t, e, `String(worker_threads.SHARE_ENV !== undefined)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof worker_threads.getEnvironmentData`), "function")
}

// MessageChannel is entangled in-process and delivers on the next tick.
func TestMessageChannelRoundTrip(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, "(() => {"+
		"const { MessageChannel } = worker_threads;"+
		"const { port1, port2 } = new MessageChannel();"+
		"port2.on('message', (value) => { globalThis.__channel = value.text; });"+
		"port1.postMessage({ text: 'hello' });"+
		"return 'queued';"+
		"})()"), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__channel`), "hello")
}

// A real worker thread receives a parent message and answers back.
func TestWorkerMessageRoundTrip(t *testing.T) {
	e := setup(t)
	path := writeWorker(t, e, `
const { parentPort } = require('worker_threads');
parentPort.on('message', (value) => {
  parentPort.postMessage({ doubled: value.n * 2 });
});
`)

	jstest.Equal(t, jstest.Str(t, e, "(() => {"+
		"const { Worker } = worker_threads;"+
		"const w = new Worker("+jsLiteral(path)+");"+
		"w.on('online', () => { globalThis.__online = true; });"+
		"w.on('message', (value) => { globalThis.__answer = value.doubled; w.terminate(); });"+
		"w.postMessage({ n: 21 });"+
		"return 'started';"+
		"})()"), "started")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__online)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__answer)`), "42")
}

// workerData crosses the thread boundary and the worker exits on its own.
func TestWorkerDataAndNaturalExit(t *testing.T) {
	e := setup(t)
	path := writeWorker(t, e, `
const { parentPort, workerData } = require('worker_threads');
parentPort.postMessage({ got: workerData.label, id: workerData.nested[1] });
`)

	jstest.Equal(t, jstest.Str(t, e, "(() => {"+
		"const { Worker } = worker_threads;"+
		"const w = new Worker("+jsLiteral(path)+", { workerData: { label: 'x', nested: [1, 2] } });"+
		"w.on('exit', (code) => { globalThis.__exit = code; });"+
		"w.on('message', (value) => { globalThis.__data = value.got + ':' + value.id; });"+
		"return 'started';"+
		"})()"), "started")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__data`), "x:2")
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__exit)`), "0")
}
