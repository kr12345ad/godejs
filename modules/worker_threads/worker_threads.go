// Package worker_threads implements Node's worker_threads builtin.
//
// GodeJS binds one JavaScriptCore context to the goroutine that created it, so
// a Worker is a real OS thread with its own env.Env. Nothing crosses threads
// except immutable strings: every cross-thread action is queued as a closure
// on the receiving endpoint and drained by a 1ms pump timer that runs on the
// receiving goroutine (the same pattern package net uses for sockets).
//
// A worker arms its pump while its main script is still running or while
// parentPort has listeners, so a script that simply falls off the end lets
// RunMain return and the thread exit with code 0, exactly like Node.
package worker_threads

import (
	"encoding/json"
	"strings"
	"sync"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/modules/web"

	// A worker evaluates the same builtins as the main thread, so the events
	// and buffer modules must be registered in this package's test binary too.
	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("worker_threads", Install) }

// pumpEveryMS is how often a live endpoint wakes to drain its queue.
const pumpEveryMS = 1.0

// endpoint owns a queue of closures that must run on one environment's
// goroutine.
type endpoint struct {
	env     *env.Env
	mu      sync.Mutex
	queue   []func()
	pumpID  int
	started bool
	stopped bool

	// msgs carries worker->parent payloads. A payload is produced on the
	// worker's goroutine but must be delivered by the parent's goroutine,
	// because the Worker object it is emitted on lives in the parent's JS
	// context: running it on the worker's pump would cross contexts.
	msgs []workerMsg
}

type workerMsg struct {
	id      float64
	payload string
}

// push queues fn for the endpoint's own goroutine. Safe from any goroutine.
func (ep *endpoint) push(fn func()) {
	ep.mu.Lock()
	ep.queue = append(ep.queue, fn)
	ep.mu.Unlock()
}

// take removes and returns the queued closures.
func (ep *endpoint) take() []func() {
	ep.mu.Lock()
	q := ep.queue
	ep.queue = nil
	ep.mu.Unlock()
	return q
}

// pushMsg queues a worker payload for the parent's pump.
func (ep *endpoint) pushMsg(id float64, payload string) {
	ep.mu.Lock()
	ep.msgs = append(ep.msgs, workerMsg{id: id, payload: payload})
	ep.mu.Unlock()
}

// takeMsgs removes and returns the queued payloads.
func (ep *endpoint) takeMsgs() []workerMsg {
	ep.mu.Lock()
	m := ep.msgs
	ep.msgs = nil
	ep.mu.Unlock()
	return m
}

// maxAbandonedTicks bounds how long a finished worker keeps its pump alive
// for messages nobody is listening to.
const maxAbandonedTicks = 2000

// pendingMsgs reports whether undelivered payloads are queued.
func (ep *endpoint) pendingMsgs() bool {
	ep.mu.Lock()
	n := len(ep.msgs)
	ep.mu.Unlock()
	return n > 0
}

type module struct {
	e       *env.Env
	host    js.Object
	exports js.Object

	ep   *endpoint
	info *childInfo

	mu      sync.Mutex
	workers map[float64]*workerState
	active  int
	seq     float64

	nextThreadID float64

	// JS helpers installed by the prelude.
	decode       js.Value
	deliver      js.Value // child: deliver a payload to parentPort
	deliverToWkr js.Value // main: deliver a payload to a Worker object
	hasListeners js.Value
	flushPorts   js.Value // main: replay messages queued before a listener
	hasPendingJS js.Value // main: any port or Worker still holding messages

	envMu   sync.Mutex
	envData map[string]string
}

type workerState struct {
	id     float64
	obj    js.Object
	name   string
	ep     *endpoint
	info   *childInfo
	refed  bool
	done   bool
	exit   int
	killed bool
}

type childInfo struct {
	ep *endpoint

	threadID   float64
	workerData string
	name       string

	mu         sync.Mutex
	mainDone   bool
	terminated bool

	worker *workerState

	// JS functions owned by the child environment.
	deliver      js.Value
	hasListeners js.Value

	// parentEP is the main thread's endpoint: payloads coming back from the
	// worker are queued there so the parent's pump delivers them.
	parentEP *endpoint

	// idleTicks counts consecutive pump ticks where the script has finished
	// and nothing is listening, while messages are still queued.
	idleTicks int

	// pumpStopped records that the child's drain timer was cleared.
	pumpStopped bool
}

func (ci *childInfo) bumpIdle() int {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.idleTicks++
	return ci.idleTicks
}

func (ci *childInfo) resetIdle() {
	ci.mu.Lock()
	ci.idleTicks = 0
	ci.mu.Unlock()
}

// setPumpStopped records that the child's drain timer has been cleared, so a
// delivery waiting for the child stops retrying once the thread is gone.
func (ci *childInfo) setPumpStopped() {
	ci.mu.Lock()
	ci.pumpStopped = true
	ci.mu.Unlock()
}

func (ci *childInfo) isPumpStopped() bool {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.pumpStopped
}

func (ci *childInfo) setMainDone() {
	ci.mu.Lock()
	ci.mainDone = true
	ci.mu.Unlock()
}

func (ci *childInfo) isMainDone() bool {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.mainDone
}

func (ci *childInfo) isTerminated() bool {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return ci.terminated
}

// pendingByEnv carries the child identity into Install: env.New has no field
// for "this environment is a worker", and process.env must stay clean.
var (
	pendingMu    sync.Mutex
	pendingByEnv = map[*env.Env]*childInfo{}
)

func registerPending(e *env.Env, ci *childInfo) {
	pendingMu.Lock()
	pendingByEnv[e] = ci
	pendingMu.Unlock()
}

func takePending(e *env.Env) *childInfo {
	pendingMu.Lock()
	ci := pendingByEnv[e]
	delete(pendingByEnv, e)
	pendingMu.Unlock()
	return ci
}

// Install builds the worker_threads exports object for e.
func Install(e *env.Env) js.Object {
	m := &module{
		e:            e,
		workers:      map[float64]*workerState{},
		nextThreadID: 1,
		envData:      map[string]string{},
		ep:           &endpoint{env: e},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()

	m.info = takePending(e)
	if m.info != nil {
		m.ep = m.info.ep
		m.nextThreadID = m.info.threadID + 1
	}

	// --- host surface -----------------------------------------------------
	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("createWorker", m.hostCreateWorker)
	m.host.SetFuncHidden("postMessage", m.hostPostMessage)
	m.host.SetFuncHidden("postToParent", m.hostPostToParent)
	m.host.SetFuncHidden("terminate", m.hostTerminate)
	m.host.SetFuncHidden("refWorker", m.hostRefWorker)
	m.host.SetFuncHidden("hasRefWorker", m.hostHasRefWorker)
	m.host.SetFuncHidden("nextTick", func(this js.Object, args []js.Value) (js.Value, error) {
		fn := arg(args, 0)
		if fn.Valid() && fn.IsFunction() {
			e.NextTick(fn)
		}
		return e.JS.Undefined(), nil
	})
	m.host.SetFuncHidden("nextThreadId", func(this js.Object, args []js.Value) (js.Value, error) {
		m.mu.Lock()
		id := m.nextThreadID
		m.nextThreadID++
		m.mu.Unlock()
		return e.JS.Num(id), nil
	})
	m.host.SetFuncHidden("setEnvData", func(this js.Object, args []js.Value) (js.Value, error) {
		key := argStr(args, 0)
		val := argStr(args, 1)
		m.envMu.Lock()
		if m.envData == nil {
			m.envData = map[string]string{}
		}
		m.envData[key] = val
		m.envMu.Unlock()
		return e.JS.Undefined(), nil
	})
	m.host.SetFuncHidden("getEnvData", func(this js.Object, args []js.Value) (js.Value, error) {
		key := argStr(args, 0)
		m.envMu.Lock()
		val, ok := m.envData[key]
		m.envMu.Unlock()
		if !ok {
			return e.JS.Undefined(), nil
		}
		return e.JS.Str(val), nil
	})

	var ee js.Value
	if env.Registered("events") {
		ee = e.Exports("events").Get("EventEmitter")
	}
	var buf js.Value
	if env.Registered("buffer") {
		buf = e.Exports("buffer").Get("Buffer")
	}

	if !ee.Valid() {
		ee = e.JS.Undefined()
	}
	if !buf.Valid() {
		buf = e.JS.Undefined()
	}
	cfg := e.JS.Obj()
	cfg.SetBool("isMainThread", m.info == nil)
	cfg.Set("EventEmitter", ee)
	cfg.Set("Buffer", buf)
	if m.info != nil {
		cfg.SetNum("threadId", m.info.threadID)
		cfg.SetStr("workerData", m.info.workerData)
		cfg.SetStr("threadName", m.info.name)
	} else {
		cfg.SetNum("threadId", 0)
		cfg.SetStr("workerData", "")
		cfg.SetStr("threadName", "MainThread")
	}

	api, err := e.JS.Eval(prelude, "[gode:worker_threads]")
	if err != nil {
		panic("godejs worker_threads: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V(), cfg.V())
	if err != nil {
		panic("godejs worker_threads: " + err.Error())
	}
	obj := built.Object()
	for _, name := range obj.OwnKeys() {
		if strings.HasPrefix(name, "__") {
			continue
		}
		m.exports.Set(name, obj.Get(name))
	}
	m.decode = obj.Get("__decode").Protect().Keep()
	if m.info != nil {
		m.info.deliver = obj.Get("__deliver").Protect()
		m.info.hasListeners = obj.Get("__hasParentListeners").Protect()
		// Keep the exports object alive for the lifetime of the worker.
		m.exports.V().Protect()
	}
	m.deliverToWkr = obj.Get("__deliverWorker").Protect().Keep()
	m.hasListeners = obj.Get("__hasParentListeners").Protect().Keep()
	m.flushPorts = obj.Get("__flush").Protect().Keep()
	m.hasPendingJS = obj.Get("__hasPending").Protect().Keep()
	obj.Delete("__decode")
	obj.Delete("__deliver")
	obj.Delete("__deliverWorker")
	obj.Delete("__hasParentListeners")
	obj.Delete("__flush")
	obj.Delete("__hasPending")

	if m.info != nil {
		// A worker keeps its pump armed while the main script runs so that
		// parent messages can arrive at any point.
		m.startPump()
	}

	return m.exports
}

// --- endpoint plumbing ----------------------------------------------------

// startPump arms the repeating drain timer on the owning goroutine.
func (m *module) startPump() {
	m.ep.mu.Lock()
	if m.ep.started || m.ep.env == nil {
		m.ep.mu.Unlock()
		return
	}
	// A later Worker can revive the parent pump after the previous one
	// stopped it, so a previous stop is not sticky.
	m.ep.started = true
	m.ep.stopped = false
	m.ep.mu.Unlock()
	fn := m.host.Get("pump")
	id := m.e.SetTimeout(fn, pumpEveryMS, true, nil)
	m.ep.mu.Lock()
	m.ep.pumpID = id
	m.ep.mu.Unlock()
}

// stopPump clears the drain timer on the owning goroutine.
func (m *module) stopPump() {
	m.ep.mu.Lock()
	id := m.ep.pumpID
	m.ep.pumpID = 0
	m.ep.started = false
	m.ep.stopped = true
	m.ep.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
}

// pump is the timer callback: it runs queued closures on this environment's
// goroutine and, inside a worker, stops draining once the script is finished
// and parentPort has no listeners so the thread can exit.
func (m *module) pump(this js.Object, _ []js.Value) (js.Value, error) {
	q := m.ep.take()
	for _, fn := range q {
		fn()
	}
	// Payloads a worker posted to its parent are delivered here, on the
	// parent's goroutine and in the parent's JS context.
	for _, mm := range m.ep.takeMsgs() {
		m.deliverWorkerMsg(mm)
	}
	m.flushQueued()
	if m.info != nil {
		if m.info.isMainDone() && !m.hasParentListeners() {
			if !m.hasQueued() {
				m.info.setPumpStopped()
				m.stopPump()
			} else if m.info.bumpIdle() >= maxAbandonedTicks {
				// Messages are queued but nothing is listening and nothing
				// drained them: let the thread exit instead of spinning.
				m.info.setPumpStopped()
				m.stopPump()
			}
		} else {
			m.info.resetIdle()
		}
	}
	return m.e.JS.Undefined(), nil
}

// flushQueued replays messages that arrived before a listener existed.
func (m *module) flushQueued() {
	if !m.flushPorts.Valid() || !m.flushPorts.IsFunction() {
		return
	}
	_, _ = m.flushPorts.Object().Call(js.Object{})
}

// hasQueued reports whether any port or Worker still holds undelivered
// messages.
func (m *module) hasQueued() bool {
	if !m.hasPendingJS.Valid() || !m.hasPendingJS.IsFunction() {
		return false
	}
	v, err := m.hasPendingJS.Object().Call(js.Object{})
	if err != nil {
		return false
	}
	return v.Valid() && v.IsBool() && v.Bool()
}

// deliverWorkerMsg emits a 'message' event on the Worker object identified by
// id. Must run on the parent's goroutine.
func (m *module) deliverWorkerMsg(mm workerMsg) {
	if !m.deliverToWkr.Valid() || !m.deliverToWkr.IsFunction() {
		return
	}
	_, _ = m.deliverToWkr.Object().Call(js.Object{},
		m.e.JS.Num(mm.id), m.e.JS.Str(mm.payload))
}

func (m *module) hasParentListeners() bool {
	if !m.info.hasListeners.Valid() || !m.info.hasListeners.IsFunction() {
		return false
	}
	v, err := m.info.hasListeners.Object().Call(js.Object{})
	if err != nil {
		return false
	}
	return v.Valid() && v.IsBool() && v.Bool()
}

// push queues fn for the parent environment's goroutine.
func (m *module) push(fn func()) { m.ep.push(fn) }

// --- host functions -------------------------------------------------------

func (m *module) hostCreateWorker(this js.Object, args []js.Value) (js.Value, error) {
	id := arg(args, 0).Float()
	filename := argStr(args, 1)
	evalMode := arg(args, 2).Bool()
	workerData := argStr(args, 3)
	argv := jsonStrings(argStr(args, 4))
	envExtra := jsonStrings(argStr(args, 5))
	name := argStr(args, 6)

	// The child identity is published before the goroutine starts so that a
	// postMessage() immediately after construction cannot observe a nil port.
	ep := &endpoint{}
	ci := &childInfo{ep: ep, threadID: id, workerData: workerData, name: name, parentEP: m.ep}
	w := &workerState{id: id, obj: this, name: name, refed: true, ep: ep, info: ci}
	ci.worker = w

	m.mu.Lock()
	m.workers[id] = w
	m.active++
	m.mu.Unlock()
	m.startPump()

	go m.runWorker(w, filename, evalMode, argv, envExtra)

	// 'online' fires on the next parent tick, after the constructor returns.
	m.push(func() { m.emit(w.obj, "online") })
	return m.e.JS.Undefined(), nil
}

func (m *module) runWorker(w *workerState, filename string, evalMode bool, argv, envExtra []string) {
	ci := w.info
	ep := ci.ep
	defer func() {
		if r := recover(); r != nil {
			errMsg := "worker failed to start"
			if e, ok := r.(error); ok {
				errMsg = e.Error()
			} else if s, ok := r.(string); ok {
				errMsg = s
			}
			ci.setMainDone()
			m.push(func() {
				m.emit(w.obj, "error", m.e.JS.Error(errMsg).V())
				m.onWorkerExit(w, 1)
			})
		}
	}()

	args := append([]string{"gode", filename}, argv...)
	child := env.New(env.Options{
		Args:     args,
		Env:      mergeEnv(m.e.Environ(), envExtra),
		Cwd:      m.e.Cwd(),
		ExecPath: m.e.Opts.ExecPath,
		Stdout:   m.e.Out(),
		Stderr:   m.e.Err(),
		Stdin:    m.e.In(),
	})
	ep.env = child
	registerPending(child, ci)

	child.InstallAll()
	web.Install(child)

	var runErr error
	if evalMode {
		_, runErr = child.RunScript(filename, "[worker eval]")
	} else {
		_, runErr = child.RunFile(filename)
	}
	ci.setMainDone()
	if runErr == nil {
		runErr = child.RunMain()
	}
	child.FireExitHooks()
	ci.setPumpStopped()

	code := child.ExitCode
	if runErr != nil && code == 0 {
		code = 1
	}
	child.JS.Close()
	registerPendingRemove(child)

	m.push(func() {
		if runErr != nil {
			errObj := m.e.JS.Error(runErr.Error())
			m.emit(w.obj, "error", errObj.V())
		}
		m.onWorkerExit(w, code)
	})
}

func registerPendingRemove(e *env.Env) {
	pendingMu.Lock()
	delete(pendingByEnv, e)
	pendingMu.Unlock()
}

func (m *module) onWorkerExit(w *workerState, code int) {
	if w.done {
		return
	}
	w.done = true
	w.exit = code
	// Node guarantees a worker's final messages are delivered before its
	// 'exit' event. Drain first; otherwise the exit listener can stop the
	// event loop before __data/message handlers run.
	for _, mm := range m.ep.takeMsgs() {
		m.deliverWorkerMsg(mm)
	}
	m.emit(w.obj, "exit", m.e.JS.Num(float64(code)))
	m.mu.Lock()
	delete(m.workers, w.id)
	m.active--
	idle := m.active <= 0
	m.mu.Unlock()
	if idle {
		m.stopPump()
	}
}

func (m *module) hostPostMessage(this js.Object, args []js.Value) (js.Value, error) {
	id := arg(args, 0).Float()
	payload := argStr(args, 1)
	m.mu.Lock()
	w := m.workers[id]
	m.mu.Unlock()
	if w == nil || w.ep == nil || w.info == nil || w.done {
		return m.e.JS.Undefined(), nil
	}
	ci := w.info
	bytes := payload
	// The child may not have finished Install yet; retry on a later pump tick
	// instead of dropping the message.
	var deliver func()
	deliver = func() {
		if ci.isPumpStopped() {
			// The worker thread is gone; nothing will ever drain this.
			return
		}
		if !ci.deliver.Valid() || !ci.deliver.IsFunction() {
			ci.ep.push(deliver)
			return
		}
		_, _ = ci.deliver.Object().Call(js.Object{}, ci.ep.env.JS.Str(bytes))
	}
	ci.ep.push(deliver)
	return m.e.JS.Undefined(), nil
}

func (m *module) hostPostToParent(this js.Object, args []js.Value) (js.Value, error) {
	if m.info == nil || m.info.worker == nil || m.info.parentEP == nil {
		return m.e.JS.Undefined(), nil
	}
	w := m.info.worker
	// Queue the payload on the parent's endpoint: the parent's pump decodes
	// and emits it inside the parent's JS context.
	m.info.parentEP.pushMsg(w.id, argStr(args, 0))
	return m.e.JS.Undefined(), nil
}

// decodePayload turns a serialized payload back into a JS value. It must run
// on the goroutine that owns m.e.
func (m *module) decodePayload(payload string) (js.Value, error) {
	if !m.decode.Valid() || !m.decode.IsFunction() {
		return m.e.JS.Undefined(), nil
	}
	return m.decode.Object().Call(js.Object{}, m.e.JS.Str(payload))
}

func (m *module) hostTerminate(this js.Object, args []js.Value) (js.Value, error) {
	id := arg(args, 0).Float()
	m.mu.Lock()
	w := m.workers[id]
	m.mu.Unlock()
	if w == nil || w.info == nil || w.done {
		return m.e.JS.Bool(true), nil
	}
	ci := w.info
	ci.mu.Lock()
	ci.terminated = true
	ci.mu.Unlock()
	ci.ep.push(func() {
		if ci.ep.env != nil {
			ci.ep.env.Exit(1)
		}
	})
	return m.e.JS.Bool(true), nil
}

func (m *module) hostRefWorker(this js.Object, args []js.Value) (js.Value, error) {
	id := arg(args, 0).Float()
	refed := arg(args, 1).Bool()
	m.mu.Lock()
	w := m.workers[id]
	m.mu.Unlock()
	if w != nil {
		w.refed = refed
	}
	// The pump timer is the only handle this module owns, and it is shared by
	// every worker, so ref()/unref() cannot change it without stranding the
	// others. hasRef() still reports the per-worker state.
	return m.e.JS.Undefined(), nil
}

func (m *module) hostHasRefWorker(this js.Object, args []js.Value) (js.Value, error) {
	id := arg(args, 0).Float()
	m.mu.Lock()
	w := m.workers[id]
	m.mu.Unlock()
	if w == nil {
		return m.e.JS.Bool(false), nil
	}
	return m.e.JS.Bool(w.refed), nil
}

// emit calls obj.emit(event, args...) if the object has an emitter on it.
func (m *module) emit(obj js.Object, event string, args ...js.Value) {
	fn := obj.Get("emit")
	if !fn.Valid() || !fn.IsFunction() {
		return
	}
	all := make([]js.Value, 0, len(args)+1)
	all = append(all, m.e.JS.Str(event))
	all = append(all, args...)
	_, _ = fn.Object().Call(obj, all...)
}

// Implemented lists the tracked worker_threads surface.
func Implemented() []string {
	return []string{
		"worker_threads.BroadcastChannel",
		"worker_threads.MessageChannel",
		"worker_threads.MessagePort",
		"worker_threads.SHARE_ENV",
		"worker_threads.Worker",
		"worker_threads.getEnvironmentData",
		"worker_threads.isMainThread",
		"worker_threads.isMarkedAsUntransferable",
		"worker_threads.markAsUntransferable",
		"worker_threads.parentPort",
		"worker_threads.receiveMessageOnPort",
		"worker_threads.resourceLimits",
		"worker_threads.setEnvironmentData",
		"worker_threads.threadId",
		"worker_threads.workerData",
	}
}

// --- helpers --------------------------------------------------------------

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func argStr(args []js.Value, i int) string {
	v := arg(args, i)
	if v.Valid() && v.IsString() {
		return v.Str()
	}
	return ""
}

// jsonStrings decodes a JSON string array; invalid or empty input is nil.
func jsonStrings(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil
	}
	return out
}

// prelude is the JavaScript half of worker_threads: the structured-clone
// codec, MessagePort/MessageChannel/BroadcastChannel, the Worker class and the
// parent side of the message channel. It never talks to JavaScriptCore from a
// foreign goroutine; every cross-thread action goes through host.
const prelude = `(function (host, cfg) {
  'use strict';
  var EventEmitter = cfg.EventEmitter;
  var BufferCtor = cfg.Buffer;
  var isMainThread = cfg.isMainThread === true;
  var SHARE_ENV = Symbol.for('nodejs.worker_threads.SHARE_ENV');

  var allPorts = [];

  // Node buffers a message that arrives before anyone is listening and
  // replays it when the first listener is attached. Without this a worker
  // that registers its handler after the first posted message silently
  // loses it, which is exactly what a startup race looks like.
  function queueOrEmit(target, event, value) {
    if (!target || typeof target.emit !== 'function') return;
    if (typeof target.listenerCount === 'function' && target.listenerCount(event) === 0) {
      if (!target._pending) target._pending = [];
      target._pending.push({ e: event, v: value });
      return;
    }
    target.emit(event, value);
  }

  function flushPending(target) {
    if (!target || !target._pending || target._pending.length === 0) return;
    if (typeof target.listenerCount === 'function' &&
        target.listenerCount(target._pending[0].e) === 0) return;
    var q = target._pending;
    target._pending = [];
    for (var i = 0; i < q.length; i++) target.emit(q[i].e, q[i].v);
  }

  if (typeof EventEmitter !== 'function') {
    // A tiny emitter keeps the module usable when events is not installed.
    EventEmitter = class {
      constructor() { this._events = Object.create(null); }
      on(name, fn) { (this._events[name] || (this._events[name] = [])).push(fn); return this; }
      addListener(name, fn) { return this.on(name, fn); }
      prependListener(name, fn) {
        var list = this._events[name] || (this._events[name] = []);
        list.unshift(fn);
        return this;
      }
      once(name, fn) {
        var self = this;
        function wrapped() {
          self.removeListener(name, wrapped);
          return fn.apply(this, arguments);
        }
        wrapped.listener = fn;
        return this.on(name, wrapped);
      }
      removeListener(name, fn) {
        var list = this._events[name] || [];
        for (var i = 0; i < list.length; i++) {
          if (list[i] === fn || list[i].listener === fn) { list.splice(i, 1); break; }
        }
        return this;
      }
      off(name, fn) { return this.removeListener(name, fn); }
      removeAllListeners(name) {
        if (name === undefined) this._events = Object.create(null);
        else delete this._events[name];
        return this;
      }
      emit(name) {
        var list = this._events[name];
        if (!list || list.length === 0) return false;
        var args = Array.prototype.slice.call(arguments, 1);
        list = list.slice();
        for (var i = 0; i < list.length; i++) list[i].apply(this, args);
        return true;
      }
      listenerCount(name) { var l = this._events[name]; return l ? l.length : 0; }
      listeners(name) { return (this._events[name] || []).slice(); }
      eventNames() { return Object.keys(this._events); }
    };
  }

  function describe(v) {
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    if (typeof v === 'string') return "type string ('" + v + "')";
    if (typeof v === 'number') return 'type number (' + v + ')';
    if (typeof v === 'boolean') return 'type boolean (' + v + ')';
    return 'type ' + typeof v;
  }

  function invalidArgType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }

  // ---------------------------------------------------- structured clone ---
  function encode(value) {
    var out = JSON.stringify({ v: value }, function (key, val) {
      if (val === undefined) return { __g: 'undefined' };
      if (typeof val === 'bigint') return { __g: 'bigint', v: String(val) };
      if (typeof val === 'symbol') {
        throw new TypeError('DataCloneError: Symbol() could not be cloned.');
      }
      if (typeof val === 'function') {
        throw new TypeError('DataCloneError: function could not be cloned.');
      }
      if (ArrayBuffer.isView(val)) {
        var view = new Uint8Array(val.buffer, val.byteOffset, val.byteLength);
        return { __g: 'bytes', d: Array.prototype.slice.call(view) };
      }
      if (val instanceof Map) {
        var pairs = [];
        val.forEach(function (v, k) { pairs.push([k, v]); });
        return { __g: 'map', d: pairs };
      }
      if (val instanceof Set) {
        var items = [];
        val.forEach(function (v) { items.push(v); });
        return { __g: 'set', d: items };
      }
      if (val instanceof Error) {
        return { __g: 'error', name: val.name, message: val.message, code: val.code, stack: val.stack };
      }
      return val;
    });
    return out === undefined ? '{"v":{"__g":"undefined"}}' : out;
  }

  function revive(node) {
    if (Array.isArray(node)) {
      var arr = [];
      for (var i = 0; i < node.length; i++) arr.push(revive(node[i]));
      return arr;
    }
    if (node !== null && typeof node === 'object') {
      var g = node.__g;
      if (g === 'undefined') return undefined;
      if (g === 'bigint') return BigInt(node.v);
      if (g === 'bytes') {
        if (BufferCtor && typeof BufferCtor.from === 'function') return BufferCtor.from(node.d);
        return new Uint8Array(node.d);
      }
      if (g === 'map') {
        var map = new Map();
        for (var j = 0; j < node.d.length; j++) map.set(revive(node.d[j][0]), revive(node.d[j][1]));
        return map;
      }
      if (g === 'set') {
        var set = new Set();
        for (var k = 0; k < node.d.length; k++) set.add(revive(node.d[k]));
        return set;
      }
      if (g === 'error') {
        var err = new Error(node.message);
        if (node.name) err.name = node.name;
        if (node.code !== undefined) err.code = node.code;
        if (node.stack) err.stack = node.stack;
        return err;
      }
      var copy = {};
      for (var key in node) copy[key] = revive(node[key]);
      return copy;
    }
    return node;
  }

  function decode(payload) {
    var parsed = JSON.parse(payload);
    return revive(parsed.v);
  }

  // ----------------------------------------------------------- ports -------
  var portSeq = 0;

  class MessagePort extends EventEmitter {
    constructor() {
      super();
      this._portId = ++portSeq;
      this._other = null;
      this._closed = false;
      this._started = false;
      this._refed = true;
      allPorts.push(this);
    }
    postMessage(value) {
      var payload = encode(value);
      var self = this;
      host.nextTick(function () {
        var other = self._other;
        if (!other || other._closed) return;
        other._started = true;
        queueOrEmit(other, 'message', decode(payload));
      });
    }
    start() { this._started = true; return this; }
    close() { this._closed = true; return this; }
    ref() { this._refed = true; return this; }
    unref() { this._refed = false; return this; }
    hasRef() { return this._refed; }
  }

  class MessageChannel {
    constructor() {
      this.port1 = new MessagePort();
      this.port2 = new MessagePort();
      this.port1._other = this.port2;
      this.port2._other = this.port1;
    }
  }

  function receiveMessageOnPort(port) {
    if (!(port instanceof MessagePort)) {
      throw invalidArgType('port', 'an instance of MessagePort', port);
    }
    if (!port._pending || port._pending.length === 0) return undefined;
    return { message: port._pending.shift().v };
  }

  // BroadcastChannel is process-local: enough for same-thread fan-out, which
  // is the subset the module can honour without a shared broker.
  var channelsByName = Object.create(null);

  class BroadcastChannel extends EventEmitter {
    constructor(name) {
      super();
      if (typeof name !== 'string') throw invalidArgType('name', 'of type string', name);
      this.name = name;
      this._closed = false;
      var list = channelsByName[name] || (channelsByName[name] = []);
      list.push(this);
    }
    postMessage(value) {
      var payload = encode(value);
      var self = this;
      host.nextTick(function () {
        var list = channelsByName[self.name] || [];
        for (var i = 0; i < list.length; i++) {
          if (list[i] === self || list[i]._closed) continue;
          queueOrEmit(list[i], 'message', { data: decode(payload) });
        }
      });
    }
    close() {
      this._closed = true;
      var list = channelsByName[this.name] || [];
      var at = list.indexOf(this);
      if (at >= 0) list.splice(at, 1);
      this.emit('close');
    }
    ref() { return this; }
    unref() { return this; }
  }

  // ---------------------------------------------------------- parent port --
  var parentPort = null;
  if (!isMainThread) {
    parentPort = new MessagePort();
    parentPort.postMessage = function (value) {
      host.postToParent(encode(value));
    };
    parentPort._isParentPort = true;
  }

  // ------------------------------------------------------------ workers ----
  var workersById = Object.create(null);

  class Worker extends EventEmitter {
    constructor(filename, options) {
      super();
      if (typeof filename !== 'string') {
        throw invalidArgType('filename', 'of type string or an instance of URL', filename);
      }
      var opts = options === undefined ? {} : options;
      if (opts === null || typeof opts !== 'object') {
        throw invalidArgType('options', 'of type object', opts);
      }
      var self = this;
      this.threadId = host.nextThreadId();
      this.threadName = typeof opts.name === 'string' ? opts.name : '';
      this.resourceLimits = opts.resourceLimits || Object.create(null);
      this.stdin = null;
      this.stdout = null;
      this.stderr = null;
      this._refed = true;
      this._exited = false;
      this._exitCode = 0;

      var argv = [];
      if (Array.isArray(opts.argv)) argv = opts.argv.map(function (a) { return String(a); });
      var envList = null;
      if (opts.env !== undefined && opts.env !== SHARE_ENV && opts.env !== null) {
        if (typeof opts.env !== 'object') {
          throw invalidArgType('options.env', 'of type object', opts.env);
        }
        envList = Object.keys(opts.env).map(function (key) {
          return key + '=' + String(opts.env[key]);
        });
      }

      host.createWorker.call(this, this.threadId, filename, opts.eval === true,
        encode(opts.workerData), JSON.stringify(argv),
        envList === null ? '' : JSON.stringify(envList), this.threadName);
      workersById[this.threadId] = this;
      this.on('exit', function (code) {
        self._exited = true;
        self._exitCode = code;
        delete workersById[self.threadId];
      });
    }
    postMessage(value, transferList) {
      if (this._exited) return;
      host.postMessage(this.threadId, encode(value));
    }
    terminate() {
      var self = this;
      return new Promise(function (resolve) {
        if (self._exited) { resolve(self._exitCode); return; }
        self.once('exit', function (code) { resolve(code); });
        host.terminate(self.threadId);
      });
    }
    ref() { this._refed = true; host.refWorker(this.threadId, true); return this; }
    unref() { this._refed = false; host.refWorker(this.threadId, false); return this; }
    hasRef() { return host.hasRefWorker(this.threadId); }
    getHeapStatistics() {
      return {
        total_heap_size: 0, total_heap_size_executable: 0, total_physical_size: 0,
        total_available_size: 0, used_heap_size: 0, heap_size_limit: 0,
        malloced_memory: 0, peak_malloced_memory: 0, does_zap_garbage: 0,
        number_of_native_contexts: 0, number_of_detached_contexts: 0,
        total_global_handles_size: 0, external_memory: 0,
      };
    }
  }

  // ------------------------------------------------------- environment -----
  var markedUntransferable = new WeakSet();

  function markAsUntransferable(object) {
    if (object === null || (typeof object !== 'object' && typeof object !== 'function')) {
      throw invalidArgType('object', 'of type object', object);
    }
    markedUntransferable.add(object);
  }

  function isMarkedAsUntransferable(object) {
    return markedUntransferable.has(object);
  }

  function setEnvironmentData(key, value) {
    host.setEnvData(String(key), encode(value));
  }

  function getEnvironmentData(key) {
    var payload = host.getEnvData(String(key));
    if (payload === undefined) return undefined;
    return decode(payload);
  }

  function moveMessagePortToContext() {
    var e = new Error('moveMessagePortToContext() is not supported');
    e.code = 'ERR_NOT_SUPPORTED';
    throw e;
  }

  function postMessageToThread() {
    return Promise.reject(new Error('postMessageToThread() is not supported'));
  }

  // ----------------------------------------------------------- exports ----
  var exports = {};
  exports.isMainThread = isMainThread;
  exports.isInternalThread = false;
  exports.threadId = cfg.threadId;
  exports.threadName = cfg.threadName;
  exports.workerData = !isMainThread && cfg.workerData ? decode(cfg.workerData) : null;
  exports.parentPort = parentPort;
  exports.Worker = Worker;
  exports.MessagePort = MessagePort;
  exports.MessageChannel = MessageChannel;
  exports.BroadcastChannel = BroadcastChannel;
  exports.receiveMessageOnPort = receiveMessageOnPort;
  exports.markAsUntransferable = markAsUntransferable;
  exports.markAsUncloneable = markAsUntransferable;
  exports.isMarkedAsUntransferable = isMarkedAsUntransferable;
  exports.moveMessagePortToContext = moveMessagePortToContext;
  exports.setEnvironmentData = setEnvironmentData;
  exports.getEnvironmentData = getEnvironmentData;
  exports.postMessageToThread = postMessageToThread;
  exports.resourceLimits = Object.create(null);
  exports.SHARE_ENV = SHARE_ENV;
  exports.locks = {};

  exports.__decode = decode;
  exports.__deliver = function (payload) {
    if (!parentPort) return;
    parentPort._started = true;
    queueOrEmit(parentPort, 'message', decode(payload));
  };
  exports.__deliverWorker = function (id, payload) {
    var w = workersById[id];
    if (!w) return;
    queueOrEmit(w, 'message', decode(payload));
  };
  exports.__flush = function () {
    for (var i = 0; i < allPorts.length; i++) flushPending(allPorts[i]);
    if (parentPort) flushPending(parentPort);
    for (var id in workersById) flushPending(workersById[id]);
  };
  exports.__hasPending = function () {
    if (parentPort && parentPort._pending && parentPort._pending.length) return true;
    for (var i = 0; i < allPorts.length; i++) {
      if (allPorts[i]._pending && allPorts[i]._pending.length) return true;
    }
    for (var id in workersById) {
      var w = workersById[id];
      if (w._pending && w._pending.length) return true;
    }
    return false;
  };
  exports.__hasParentListeners = function () {
    if (!parentPort) return false;
    try {
      return parentPort.listenerCount('message') > 0 ||
        parentPort.listenerCount('messageerror') > 0 ||
        parentPort.listenerCount('close') > 0;
    } catch (err) {
      return false;
    }
  };

  return exports;
})`

// mergeEnv returns base with extra KEY=VALUE entries applied on top.
func mergeEnv(base []string, extra []string) []string {
	if len(extra) == 0 {
		return append([]string(nil), base...)
	}
	out := make([]string, 0, len(base)+len(extra))
	seen := map[string]int{}
	for _, kv := range base {
		if i := strings.IndexByte(kv, '='); i > 0 {
			seen[kv[:i]] = len(out)
		}
		out = append(out, kv)
	}
	for _, kv := range extra {
		i := strings.IndexByte(kv, '=')
		if i <= 0 {
			continue
		}
		if at, ok := seen[kv[:i]]; ok {
			out[at] = kv
			continue
		}
		seen[kv[:i]] = len(out)
		out = append(out, kv)
	}
	return out
}
