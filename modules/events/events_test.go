package events

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newEnv(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "events", Install)
	return e
}

func TestOnEmitThisAndArgs(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const seen = [];
		function h(a, b) { seen.push(this === em, a, b); }
		em.on("x", h);
		const ret = em.emit("x", 1, 2);
		return [seen.join("|"), ret].join(";");
	})()`)
	jstest.Equal(t, got, "true|1|2;true")
}

func TestDOMBrandsAndEmptyListenerList(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const list = events.getEventListeners(em, 'x');
		const fn = function () {};
		em.on('x', fn);
		return [
			Array.isArray(list), list.length,
			Object.prototype.toString.call(new Event('e')),
			Object.prototype.toString.call(new CustomEvent('c')),
			Object.prototype.toString.call(new EventTarget()),
			events.getEventListeners(em, 'x')[0] === fn,
		].map(String).join(',');
	})()`)
	jstest.Equal(t, got, "true,0,[object Event],[object CustomEvent],[object EventTarget],true")
}

func TestEmitWithoutListeners(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `new events.EventEmitter().emit("nobody")`)
	jstest.Equal(t, got, "false")
}

func TestOnceFiresOnce(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		let n = 0;
		em.once("x", function () { n++; });
		em.emit("x");
		em.emit("x");
		return n;
	})()`)
	jstest.Equal(t, got, "1")
}

func TestPrependOrder(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const order = [];
		em.on("x", () => order.push("a"));
		em.prependListener("x", () => order.push("b"));
		em.prependOnceListener("x", () => order.push("c"));
		em.emit("x");
		em.emit("x");
		return order.join("");
	})()`)
	jstest.Equal(t, got, "cbaba")
}

func TestOffAndRemoveListener(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		let n = 0;
		function f() { n++; }
		em.on("x", f);
		const ret = em.off("x", f);
		em.emit("x");
		em.on("x", f);
		em.removeListener("x", f);
		em.emit("x");
		return [n, ret === em].join(",");
	})()`)
	jstest.Equal(t, got, "0,true")
}

func TestListenerCountAndEventNames(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		function f() {}
		em.addListener("a", f);
		em.on("a", f);
		em.on("b", f);
		return [em.listenerCount("a"), em.listenerCount("a", f), em.eventNames().join("|")].join(",");
	})()`)
	jstest.Equal(t, got, "2,2,a|b")
}

func TestRemoveAllListeners(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		function f() {}
		em.on("a", f);
		em.on("b", f);
		em.removeAllListeners("a");
		const one = [em.listenerCount("a"), em.listenerCount("b")].join(",");
		em.removeAllListeners();
		return one + ";" + em.listenerCount("b");
	})()`)
	jstest.Equal(t, got, "0,1;0")
}

func TestErrorEventRethrows(t *testing.T) {
	e := newEnv(t)
	err := jstest.Throws(t, e, `(function () {
		const em = new events.EventEmitter();
		em.emit("error", new Error("boom"));
	})()`)
	if err.Message != "boom" {
		t.Fatalf("error message = %q, want %q", err.Message, "boom")
	}

	err = jstest.Throws(t, e, `new events.EventEmitter().emit("error", "kaput")`)
	if err.Message != "Unhandled error. ('kaput')" {
		t.Fatalf("unhandled message = %q", err.Message)
	}
}

func TestMaxListeners(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		em.setMaxListeners(3);
		events.defaultMaxListeners = 20;
		const em2 = new events.EventEmitter();
		events.setMaxListeners(4, em2);
		return [em.getMaxListeners(), em2.getMaxListeners(), events.getMaxListeners(em)].join(",");
	})()`)
	jstest.Equal(t, got, "3,4,3")

	if err := jstest.Throws(t, e, `new events.EventEmitter().setMaxListeners(-1)`); err.Name != "RangeError" {
		t.Fatalf("setMaxListeners(-1) error name = %q, want RangeError", err.Name)
	}
}

func TestModuleHelpers(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		function f() {}
		em.on("x", f);
		return [
			events.listenerCount(em, "x"),
			events.getEventListeners(em, "x").length,
			typeof events.getEventListeners(em, "none")
		].join(",");
	})()`)
	jstest.Equal(t, got, "1,1,object")
}

func TestOnceRawAndUnwrapped(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		function f() {}
		const em = new events.EventEmitter();
		em.once("x", f);
		const raw = em.rawListeners("x");
		return [em.listeners("x")[0] === f, typeof raw[0], raw[0].listener === f].join(",");
	})()`)
	jstest.Equal(t, got, "true,function,true")
}

func TestNewAndRemoveListenerEvents(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const added = [];
		const removed = [];
		em.on("removeListener", (n) => removed.push(n));
		em.on("newListener", (n) => added.push(n));
		function f() {}
		em.on("x", f);
		em.off("x", f);
		return added.join("|") + ";" + removed.join("|");
	})()`)
	jstest.Equal(t, got, "x;x")
}

func TestOnceReturnsPromise(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `events.once(new events.EventEmitter(), "x") instanceof Promise`)
	jstest.Equal(t, got, "true")
}

func TestErrorMonitorSymbol(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `[typeof events.errorMonitor, events.errorMonitor === Symbol.for("events.errorMonitor")].join(",")`)
	jstest.Equal(t, got, "symbol,true")
}

func TestEventClass(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const ev = new events.Event('go', {bubbles: true, cancelable: true, composed: true});
		const before = [ev.type, ev.bubbles, ev.cancelable, ev.composed, ev.defaultPrevented, ev.cancelBubble, ev.returnValue, ev.eventPhase, ev.isTrusted, typeof ev.timeStamp].join(',');
		ev.preventDefault();
		ev.cancelBubble = true;
		ev.initEvent('again', false, false);
		return before + ';' + [ev.defaultPrevented, ev.cancelBubble, ev.type].join(',');
	})()`)
	jstest.Equal(t, got, "go,true,true,true,false,false,true,0,false,number;false,false,again")
}

func TestEventTargetDispatch(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const target = new events.EventTarget();
		const seen = [];
		function h(ev) { seen.push(ev.type, ev.target === target, ev.currentTarget === target, ev.eventPhase); }
		const obj = { handleEvent(ev) { seen.push('obj:' + ev.type); } };
		target.addEventListener('go', h, { once: true });
		target.addEventListener('go', obj);
		const ev = new events.Event('go', { cancelable: true });
		const ok = target.dispatchEvent(ev);
		target.dispatchEvent(new events.Event('go'));
		target.removeEventListener('go', obj);
		target.dispatchEvent(new events.Event('go'));
		return [seen.join('|'), ok, ev.target === target, ev.defaultPrevented].join(';');
	})()`)
	jstest.Equal(t, got, "go|true|true|2|obj:go|obj:go;true;true;false")
}

func TestEventTargetStopAndPrevent(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const target = new events.EventTarget();
		const seen = [];
		target.addEventListener('x', (ev) => { seen.push('a'); ev.preventDefault(); ev.stopImmediatePropagation(); });
		target.addEventListener('x', () => seen.push('b'));
		const ev = new events.Event('x', { cancelable: true });
		const ok = target.dispatchEvent(ev);
		return [seen.join(''), ok, ev.defaultPrevented, ev.cancelBubble].join(',');
	})()`)
	jstest.Equal(t, got, "a,false,true,true")
}

func TestCustomEvent(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const ev = new events.CustomEvent('x', { detail: { a: 1 } });
		return [ev.type, JSON.stringify(ev.detail), ev instanceof events.Event].join(',');
	})()`)
	jstest.Equal(t, got, "x,{\"a\":1},true")
}

func TestNodeEventTarget(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const nt = new events.NodeEventTarget();
		let total = 0;
		function f(v) { total += v; }
		nt.addListener('a', f);
		nt.on('b', () => {});
		const emitted = nt.emit('a', 2);
		nt.off('a', f);
		nt.emit('a', 3);
		nt.setMaxListeners(7);
		return [total, emitted, nt.eventNames().join('|'), nt.listenerCount('a'), nt.getMaxListeners(), typeof nt.removeAllListeners].join(',');
	})()`)
	jstest.Equal(t, got, "2,true,b,0,7,function")
}

func TestOnAsyncIterator(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const it = events.on(em, 'x');
		const structural = [typeof it.next, typeof it.return, typeof it.throw, it[Symbol.asyncIterator]() === it].join(',');
		const before = em.listenerCount('x');
		it.return();
		return structural + ';' + before + ',' + em.listenerCount('x');
	})()`)
	jstest.Equal(t, got, "function,function,function,true;1,0")
}

func TestOncePromiseSettles(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const p = events.once(em, 'x');
		em.emit('x', 1, 2);
		return [p instanceof Promise, em.listenerCount('x'), em.listenerCount('error')].join(',');
	})()`)
	jstest.Equal(t, got, "true,0,0")
}

func TestOnceRejectsOnError(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const p = events.once(em, 'x');
		p.catch(function () {});
		em.emit('error', new Error('nope'));
		return [p instanceof Promise, em.listenerCount('x'), em.listenerCount('error')].join(',');
	})()`)
	jstest.Equal(t, got, "true,0,0")
}

func TestOnceEventTarget(t *testing.T) {
	e := newEnv(t)
	if _, err := e.JS.Eval(`globalThis.__onceTarget = 'pending';
		(function () {
			const et = new EventTarget();
			const ev = new Event('go');
			events.once(et, 'go').then(function (args) {
				globalThis.__onceTarget = args.length + ':' + (args[0] === ev);
			});
			et.dispatchEvent(ev);
		})()`, "[test]"); err != nil {
		t.Fatal(err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__onceTarget`), "1:true")
}

func TestOnceInvalidOptionsReject(t *testing.T) {
	e := newEnv(t)
	if _, err := e.JS.Eval(`globalThis.__onceCodes = '';
		(function () {
			const em = new events.EventEmitter();
			Promise.all([1, 'hi', null, false, function () {}, Symbol(), 1n].map(function (options) {
				return events.once(em, 'x', options).then(
					function () { return 'resolved'; },
					function (err) { return err && err.code; });
			})).then(function (codes) { globalThis.__onceCodes = codes.join(','); });
		})()`, "[test]"); err != nil {
		t.Fatal(err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__onceCodes`),
		"ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE")
}

func TestOnceInvalidEmitterRejects(t *testing.T) {
	e := newEnv(t)
	if _, err := e.JS.Eval(`globalThis.__onceEmitter = '';
		events.once({ signal: null }, 'x').then(
			function () { globalThis.__onceEmitter = 'resolved'; },
			function (err) { globalThis.__onceEmitter = err && err.code; })`, "[test]"); err != nil {
		t.Fatal(err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__onceEmitter`), "ERR_INVALID_ARG_TYPE")
}

func TestOncePreAbortedSignal(t *testing.T) {
	e := newEnv(t)
	if _, err := e.JS.Eval(`globalThis.__onceAborted = '';
		(function () {
			const signal = {
				aborted: true, reason: 'why',
				addEventListener: function () {},
				removeEventListener: function () {}
			};
			const em = new events.EventEmitter();
			events.once(em, 'x', { signal: signal }).catch(function (err) {
				globalThis.__onceAborted = [err.name, err.code, err.cause, em.listenerCount('x')].join(',');
			});
		})()`, "[test]"); err != nil {
		t.Fatal(err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__onceAborted`), "AbortError,ABORT_ERR,why,0")
}

func TestOnceAbortCleansUp(t *testing.T) {
	e := newEnv(t)
	if _, err := e.JS.Eval(`globalThis.__onceAbortCleanup = '';
		(function () {
			const signal = {
				aborted: false, reason: undefined, _listeners: [],
				addEventListener: function (type, fn, opts) {
					if (type === 'abort') this._listeners.push({ fn: fn, once: !!(opts && opts.once) });
				},
				removeEventListener: function (type, fn) {
					if (type === 'abort') this._listeners = this._listeners.filter(function (l) { return l.fn !== fn; });
				}
			};
			const et = new EventTarget();
			const p = events.once(et, 'foo', { signal: signal });
			const before = events.listenerCount(signal, 'abort') + ',' + events.listenerCount(et, 'foo');
			p.catch(function (err) {
				globalThis.__onceAbortCleanup = before + ';' + [err.name, err.code,
					events.listenerCount(signal, 'abort'), events.listenerCount(et, 'foo')].join(',');
			});
			signal._listeners.slice().forEach(function (entry) { entry.fn(); });
		})()`, "[test]"); err != nil {
		t.Fatal(err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__onceAbortCleanup`), "1,1;AbortError,ABORT_ERR,0,0")
}

func TestListenerCountEventTarget(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const et = new EventTarget();
		function h() {}
		et.addEventListener('x', h);
		et.addEventListener('x', function () {}, { once: true });
		return [events.listenerCount(et, 'x'), events.getEventListeners(et, 'x').length,
			events.getEventListeners(et, 'none').length].join(',');
	})()`)
	jstest.Equal(t, got, "2,2,0")
}

func TestDOMGlobals(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `[typeof Event, typeof EventTarget, typeof CustomEvent,
		new EventTarget() instanceof events.EventTarget].join(',')`)
	jstest.Equal(t, got, "function,function,function,true")
}

func TestOnceAbortedSignal(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter();
		const signal = { aborted: true, addEventListener() {}, removeEventListener() {} };
		const p = events.once(em, 'x', { signal });
		p.catch(function () {});
		return [p instanceof Promise, em.listenerCount('x')].join(',');
	})()`)
	jstest.Equal(t, got, "true,0")
}

func TestAddAbortListener(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const handlers = [];
		const signal = {
			aborted: false,
			addEventListener(type, fn) { if (type === 'abort') handlers.push(fn); },
			removeEventListener(type, fn) { const i = handlers.indexOf(fn); if (i >= 0) handlers.splice(i, 1); }
		};
		let calls = 0;
		const d = events.addAbortListener(signal, () => calls++);
		const sym = (typeof Symbol.dispose === 'symbol') ? Symbol.dispose : Symbol.for('nodejs.dispose');
		const before = handlers.length;
		handlers[0]();
		const afterFire = calls;
		d[sym]();
		return [before, afterFire, calls, handlers.length, typeof d[sym]].join(',');
	})()`)
	jstest.Equal(t, got, "1,1,1,0,function")
}

func TestCaptureRejectionsSurface(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		const em = new events.EventEmitter({ captureRejections: true });
		const handler = function (err, name) { globalThis.__rejected = err.message + ':' + name; };
		em[events.captureRejectionSymbol] = handler;
		em.on('x', () => Promise.reject(new Error('boom')));
		em.emit('x');
		return [em.captureRejections, typeof em[events.captureRejectionSymbol], events.captureRejections].join(',');
	})()`)
	jstest.Equal(t, got, "true,function,false")
}

func TestCaptureRejectionSymbolIdentity(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `[typeof events.captureRejectionSymbol, events.captureRejectionSymbol === Symbol.for('nodejs.rejection')].join(',')`)
	jstest.Equal(t, got, "symbol,true")
}

func TestClassAndSubclass(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `(function () {
		class Sub extends events.EventEmitter {}
		const s = new Sub();
		function Other() { events.EventEmitter.call(this); }
		Other.prototype = Object.create(events.EventEmitter.prototype);
		const o = new Other();
		o.on("x", () => {});
		return [
			s instanceof events.EventEmitter,
			s instanceof Sub,
			typeof s.on,
			o.listenerCount("x")
		].join(",");
	})()`)
	jstest.Equal(t, got, "true,true,function,1")
}
