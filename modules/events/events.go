package events

import (
	"errors"
	"math"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("events", Install) }

// stateKey is the hidden property holding an emitter's listener state.
const stateKey = "$gode$events"

type module struct {
	e       *env.Env
	exports js.Object
	ctor    js.Object
	proto   js.Object

	seq int

	eq      js.Object // (a, b) => a === b
	thrower js.Object // (name, msg) => { throw new globalThis[name](msg) }
	rethrow js.Object // (v) => { throw v }
	promise js.Object // (executor) => new Promise(executor)
	capture js.Object // attaches a rejection handler to a listener's promise
	symKeys js.Object // symbol event-name registry: key(v) / name(k)
}

// Install builds the events module's exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e}
	m.exports = e.JS.Obj().Keep()
	m.exports.Set("defaultMaxListeners", e.JS.Num(10))

	m.proto = e.JS.Obj().Keep()
	m.eq = m.mustEval(`(function (a, b) { return a === b; })`, "events.eq").Keep()
	m.thrower = m.mustEval(`(function (name, msg) { throw new globalThis[name](msg); })`, "events.throw").Keep()
	m.rethrow = m.mustEval(`(function (v) { throw v; })`, "events.rethrow").Keep()
	m.promise = m.mustEval(`(function (exec) { return new Promise(exec); })`, "events.promise").Keep()
	m.capture = m.mustEval(`(function (p, em, name, args) {
		if (p && typeof p.then === 'function') {
			p.then(undefined, function (err) {
				var h = em[Symbol.for('nodejs.rejection')];
				if (typeof h === 'function') { h.call(em, err, name, ...args); }
				else { em.emit('error', err); }
			});
		}
	})`, "events.capture").Keep()
	m.exports.Set("captureRejections", e.JS.Bool(false))
	// Event names may be symbols; the listener table is keyed by strings, so
	// symbols get a stable synthetic key and the original value is kept for
	// eventNames()/listeners().
	m.symKeys = m.mustEval(`(function () {
		var ids = new Map(), values = new Map(), next = 1;
		return {
			key: function (v) {
				if (typeof v !== 'symbol') return null;
				var k = ids.get(v);
				if (!k) { k = '$sym$' + (next++); ids.set(v, k); values.set(k, v); }
				return k;
			},
			name: function (k) {
				if (typeof k !== 'string' || k.indexOf('$sym$') !== 0) return null;
				return values.get(k);
			}
		};
	})()`, "events.symKeys").Keep()

	m.proto.SetFunc("addListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.add(this, args, false, false)
	})
	m.proto.SetFunc("on", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.add(this, args, false, false)
	})
	m.proto.SetFunc("once", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.add(this, args, false, true)
	})
	m.proto.SetFunc("prependListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.add(this, args, true, false)
	})
	m.proto.SetFunc("prependOnceListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.add(this, args, true, true)
	})
	m.proto.SetFunc("off", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	m.proto.SetFunc("removeListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	m.proto.SetFunc("removeAllListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeAll(this, args)
	})
	m.proto.SetFunc("emit", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.emit(this, args)
	})
	m.proto.SetFunc("listeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.listeners(this, args, false)
	})
	m.proto.SetFunc("rawListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.listeners(this, args, true)
	})
	m.proto.SetFunc("listenerCount", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.listenerCount(this, args)
	})
	// _godeHasAnyListeners checks several event names in one host call. The
	// HTTP prelude asks about four body events per request, and one bridge
	// round trip per name was the hottest path in the server.
	m.proto.SetFunc("_godeHasAnyListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.hasAny(this, args), nil
	})
	m.proto.SetFunc("eventNames", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.eventNames(this), nil
	})
	m.proto.SetFunc("getMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Num(m.max(this)), nil
	})
	m.proto.SetFunc("setMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		if err := m.setMax(this, args); err != nil {
			return js.Value{}, err
		}
		return this.V(), nil
	})

	// A tiny constructor wrapper gives the host object real class semantics:
	// new, instanceof, subclassing via extends, and EventEmitter.call(this).
	// The hidden listener table is created lazily by state(), so constructing
	// an emitter stays entirely in JavaScript; the HTTP server builds two per
	// request and a host call here dominated small responses.
	maker := m.mustEval(`(function (proto) {
		function EventEmitter(options) {
			if (!(this instanceof EventEmitter)) {
				throw new TypeError("Class constructor EventEmitter cannot be invoked without 'new'");
			}
			if (options !== undefined && options !== null && typeof options === 'object') {
				if (typeof options.captureRejections === 'boolean') {
					this.captureRejections = options.captureRejections;
				}
			}
		}
		EventEmitter.prototype = proto;
		Object.defineProperty(proto, "constructor", {
			value: EventEmitter, writable: true, configurable: true
		});
		return EventEmitter;
	})`, "events.ctor")
	ctorVal, err := maker.Object().Call(js.Object{}, m.proto.V())
	if err != nil {
		panic("godejs events: " + err.Error())
	}
	m.ctor = ctorVal.Object().Keep()

	m.exports.Set("EventEmitter", m.ctor.V())
	if sym, err := e.JS.Eval(`Symbol.for("events.errorMonitor")`, "events.symbol"); err == nil {
		m.exports.Set("errorMonitor", sym)
	}
	if sym, err := e.JS.Eval(`Symbol.for("nodejs.rejection")`, "events.symbol"); err == nil {
		m.exports.Set("captureRejectionSymbol", sym)
	}

	// EventEmitterAsyncResource wraps every emitter in an async_hooks resource,
	// so destroy tracking and async ids work exactly like Node's subclass.
	arHost := e.JS.Obj()
	arHost.SetFunc("asyncResource", func(_ js.Object, _ []js.Value) (js.Value, error) {
		hooks := e.Exports("async_hooks")
		if hooks.Valid() {
			if v := hooks.Get("AsyncResource"); v.IsFunction() {
				return v, nil
			}
		}
		return e.JS.Undefined(), nil
	})
	arMaker, arErr := e.JS.Eval(`(function (EventEmitter, host) {
  function EventEmitterAsyncResource(options) {
    if (!(this instanceof EventEmitterAsyncResource)) {
      return new EventEmitterAsyncResource(options);
    }
    EventEmitter.call(this, typeof options === 'object' && options !== null ? options : undefined);
    var name = typeof options === 'string' ? options : (options && options.name);
    if (name !== undefined && typeof name !== 'string') {
      throw new TypeError('The "options.name" property must be of type string. Received ' + typeof name);
    }
    var ASR = host.asyncResource();
    if (typeof ASR !== 'function') {
      this.asyncResource = undefined;
      this.asyncId = 0;
      this.triggerAsyncId = 0;
      return;
    }
    this.asyncResource = new ASR(name || 'EventEmitterAsyncResource', options);
    this.asyncId = this.asyncResource.asyncId();
    this.triggerAsyncId = this.asyncResource.triggerAsyncId();
  }
  Object.setPrototypeOf(EventEmitterAsyncResource.prototype, EventEmitter.prototype);
  Object.setPrototypeOf(EventEmitterAsyncResource, EventEmitter);
  Object.defineProperty(EventEmitterAsyncResource.prototype, 'constructor', {
    value: EventEmitterAsyncResource, writable: true, configurable: true
  });
  Object.defineProperty(EventEmitterAsyncResource.prototype, 'emitDestroy', {
    value: function emitDestroy() {
      if (this.asyncResource && typeof this.asyncResource.emitDestroy === 'function') {
        this.asyncResource.emitDestroy();
      }
      return this;
    }, writable: true, configurable: true
  });
  return EventEmitterAsyncResource;
})`, "events.asyncResource")
	if arErr == nil {
		if arCtor, callErr := arMaker.Object().Call(js.Object{}, m.ctor.V(), arHost.V()); callErr == nil && arCtor.Valid() {
			m.exports.Set("EventEmitterAsyncResource", arCtor)
		}
	}
	m.installDOM()

	m.exports.SetFunc("getEventListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 1 || !args[0].IsObject() {
			got := "undefined"
			if len(args) > 0 {
				got = describe(args[0])
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "emitter" argument must be of type EventEmitter. Received `+got)
		}
		em := args[0].Object()
		name := ""
		if len(args) > 1 {
			name = m.nameKey(args[1])
		}
		if !em.Has(stateKey) {
			if vals, ok := m.targetListeners(em, name); ok {
				return e.JS.Arr(vals).V(), nil
			}
		}
		recs := m.records(em, name)
		if len(recs) == 0 {
			// Node returns an empty array (not an empty object).
			return e.JS.Arr(nil).V(), nil
		}
		out := e.JS.Arr(nil)
		for i, rec := range recs {
			out.SetAt(i, rec.Get("original"))
		}
		return out.V(), nil
	})
	m.exports.SetFunc("getMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 1 || !args[0].IsObject() {
			return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
		}
		return e.JS.Num(m.max(args[0].Object())), nil
	})
	m.exports.SetFunc("setMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 1 {
			return js.Value{}, errors.New(`The "n" argument must be of type number. Received undefined`)
		}
		n, err := m.checkMax(args[0])
		if err != nil {
			return js.Value{}, err
		}
		for _, a := range args[1:] {
			if a.IsObject() {
				m.state(a.Object()).Set("max", e.JS.Num(n))
			}
		}
		return e.JS.Undefined(), nil
	})
	m.exports.SetFunc("listenerCount", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[0].IsObject() {
			return e.JS.Num(0), nil
		}
		return m.countOf(args[0].Object(), m.nameKey(args[1]), args[2:]), nil
	})
	// events.once/on are the promise helpers. They are implemented in JS over
	// the public emitter/EventTarget API, as in lib/events.js.
	if onceFn, onceErr := e.JS.Eval(`(function () {
	  'use strict';
	  function describe(value) {
	    if (value === null) return 'null';
	    if (value === undefined) return 'undefined';
	    if (Array.isArray(value)) return 'an instance of Array';
	    var t = typeof value;
	    if (t === 'string') return 'type string (' + JSON.stringify(value) + ')';
	    if (t === 'number') return 'type number (' + String(value) + ')';
	    if (t === 'boolean') return 'type boolean (' + String(value) + ')';
	    if (t === 'symbol') return 'type symbol (' + String(value) + ')';
	    if (t === 'bigint') return 'type bigint (' + String(value) + 'n)';
	    if (t === 'function') return 'function ' + (value.name || '(anonymous)');
	    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
	  }
	  function invalidArg(name, expected, value) {
	    var err = new TypeError('The "' + name + '" argument must be ' + expected +
	      '. Received ' + describe(value));
	    err.code = 'ERR_INVALID_ARG_TYPE';
	    return err;
	  }
	  function isAbortSignal(s) {
	    return !!s && typeof s === 'object' && typeof s.aborted === 'boolean' &&
	      typeof s.addEventListener === 'function';
	  }
	  function abortError(signal) {
	    var reason = signal ? signal.reason : undefined;
	    var err;
	    if (typeof DOMException === 'function') {
	      err = new DOMException('The operation was aborted', 'AbortError');
	    } else {
	      err = new Error('The operation was aborted');
	      err.name = 'AbortError';
	    }
	    try {
	      Object.defineProperty(err, 'code', {
	        __proto__: null, value: 'ABORT_ERR',
	        writable: true, configurable: true, enumerable: false
	      });
	    } catch (e) { try { err.code = 'ABORT_ERR'; } catch (e2) {} }
	    try {
	      Object.defineProperty(err, 'cause', {
	        __proto__: null, value: reason,
	        writable: true, configurable: true, enumerable: false
	      });
	    } catch (e) {}
	    return err;
	  }
	  function once(emitter, name, options) {
	    return new Promise(function (resolve, reject) {
	      if (options !== undefined) {
	        if (options === null || typeof options !== 'object') {
	          reject(invalidArg('options', 'of type object', options));
	          return;
	        }
	        if (options.signal !== undefined && !isAbortSignal(options.signal)) {
	          reject(invalidArg('options.signal', 'an instance of AbortSignal', options.signal));
	          return;
	        }
	      }
	      var signal = options && options.signal;
	      if (signal && signal.aborted) { reject(abortError(signal)); return; }
	      var isEmitter = !!emitter && typeof emitter.once === 'function' &&
	        typeof emitter.removeListener === 'function';
	      var isTarget = !!emitter && typeof emitter.addEventListener === 'function';
	      if (!isEmitter && !isTarget) {
	        reject(invalidArg('emitter', 'an instance of EventEmitter or EventTarget', emitter));
	        return;
	      }
	      var done = false;
	      function cleanup() {
	        if (done) return;
	        done = true;
	        if (signal) { try { signal.removeEventListener('abort', onAbort); } catch (e) {} }
	        if (isEmitter) {
	          emitter.removeListener(name, onEvent);
	          if (name !== 'error') emitter.removeListener('error', onError);
	        } else {
	          try { emitter.removeEventListener(name, onEvent); } catch (e) {}
	        }
	      }
	      function onAbort() { cleanup(); reject(abortError(signal)); }
	      function onEvent() {
	        var args = Array.prototype.slice.call(arguments);
	        cleanup();
	        resolve(args);
	      }
	      function onError(err) { cleanup(); reject(err); }
	      // resistStopPropagation makes the abort listener run even when another
      // signal listener calls stopImmediatePropagation(), as Node guarantees.
      if (signal) signal.addEventListener('abort', onAbort, { once: true, resistStopPropagation: true });
	      if (isEmitter) {
	        emitter.once(name, onEvent);
	        if (name !== 'error') emitter.once('error', onError);
	      } else {
	        emitter.addEventListener(name, onEvent, { once: true });
	      }
	    });
	  }
	  return once;
	})()`, "events.once"); onceErr == nil && onceFn.Valid() {
		m.exports.Set("once", onceFn)
	}
	// Node's events module *is* the EventEmitter class: require('events') is
	// constructible and also carries every helper as a property, and
	// EventEmitter.EventEmitter points back at itself.
	merge, err := e.JS.Eval(`(function (ctor, exports) {
  if (!ctor) return exports;
  Object.getOwnPropertyNames(exports).forEach(function (name) {
    if (name === 'EventEmitter') return;
    try {
      Object.defineProperty(ctor, name,
        Object.getOwnPropertyDescriptor(exports, name));
    } catch (e) {}
  });
  Object.defineProperty(ctor, 'EventEmitter', {
    value: ctor, writable: true, configurable: true, enumerable: false
  });
  return ctor;
})`, "events.merge")
	if err == nil && merge.Valid() && merge.IsFunction() {
		if merged, callErr := merge.Object().Call(js.Object{}, m.ctor.V(), m.exports.V()); callErr == nil && merged.Valid() {
			return merged.Object()
		}
	}
	return m.exports
}

func (m *module) mustEval(src, name string) js.Object {
	v, err := m.e.JS.Eval(src, name)
	if err != nil {
		panic("godejs events: " + err.Error())
	}
	return v.Object()
}

func (m *module) fnValue(name string, fn js.Func) js.Value {
	holder := m.e.JS.Obj()
	holder.SetFunc(name, fn)
	return holder.Get(name)
}

func (m *module) same(a, b js.Value) bool {
	v, err := m.eq.Call(js.Object{}, a, b)
	return err == nil && v.Bool()
}

// nameKey returns the listener-table key for an event name, preserving symbol
// identity through the synthetic symbol registry.
func (m *module) nameKey(v js.Value) string {
	if !v.Valid() {
		return ""
	}
	// Ordinary string event names need no help from the symbol registry.
	if v.IsString() {
		return v.Str()
	}
	fn := m.symKeys.Get("key")
	if fn.Valid() && fn.IsFunction() {
		if res, err := fn.Object().Call(m.symKeys, v); err == nil && res.Valid() && res.IsString() {
			return res.Str()
		}
	}
	return v.Str()
}

// nameValue resolves a synthetic symbol key back to its symbol; invalid for
// ordinary string event names.
func (m *module) nameValue(key string) js.Value {
	fn := m.symKeys.Get("name")
	if fn.Valid() && fn.IsFunction() {
		if res, err := fn.Object().Call(m.symKeys, m.e.JS.Str(key)); err == nil && res.Valid() && !res.IsUndefined() && !res.IsNull() {
			return res
		}
	}
	return js.Value{}
}

func (m *module) throwNamed(name, msg string) error {
	return js.Throw(name, msg)
}

// state returns the emitter's hidden state object, creating it on first use.
func (m *module) state(em js.Object) js.Object {
	if em.Has(stateKey) {
		v := em.Get(stateKey)
		if v.IsObject() {
			return v.Object()
		}
	}
	st := m.e.JS.Obj()
	st.SetHidden("list", m.e.JS.Obj().V())
	max := m.exports.Get("defaultMaxListeners").Float()
	if math.IsNaN(max) || max < 0 {
		max = 10
	}
	st.Set("max", m.e.JS.Num(max))
	em.SetHidden(stateKey, st.V())
	return st
}

func (m *module) lists(em js.Object) js.Object {
	return m.state(em).Get("list").Object()
}

func (m *module) max(em js.Object) float64 {
	// AbortSignal starts at zero rather than defaultMaxListeners; only an
	// explicit setMaxListeners changes it.
	if !em.Has(stateKey) {
		if ctor := em.Get("constructor"); ctor.Valid() && ctor.IsFunction() {
			if name := ctor.Object().Get("name"); name.Valid() && name.Str() == "AbortSignal" {
				return 0
			}
		}
	}
	n := m.state(em).Get("max").Float()
	if math.IsNaN(n) || n < 0 {
		return 0
	}
	return n
}

func (m *module) records(em js.Object, name string) []js.Object {
	v := m.lists(em).Get(name)
	if !v.IsArray() {
		return nil
	}
	arr := v.Object()
	n := arr.Length()
	out := make([]js.Object, 0, n)
	for i := 0; i < n; i++ {
		item := arr.At(i)
		if item.IsObject() {
			out = append(out, item.Object())
		}
	}
	return out
}

func (m *module) write(em js.Object, name string, recs []js.Object) {
	lists := m.lists(em)
	if len(recs) == 0 {
		lists.Delete(name)
		return
	}
	arr := m.e.JS.Arr(nil)
	for i, rec := range recs {
		arr.SetAt(i, rec.V())
	}
	lists.Set(name, arr.V())
}

func (m *module) nextID() int {
	m.seq++
	return m.seq
}

func (m *module) makeRecord(em js.Object, name string, listener js.Value, once bool) js.Object {
	id := m.nextID()
	rec := m.e.JS.Obj()
	rec.SetHidden("id", m.e.JS.Num(float64(id)))
	rec.SetHidden("once", m.e.JS.Bool(once))
	rec.SetHidden("original", listener)

	if !once {
		rec.SetHidden("fn", listener)
		return rec
	}

	wrapper := m.fnValue("onceWrapper", func(this js.Object, args []js.Value) (js.Value, error) {
		m.removeByID(em, name, id)
		return listener.Object().Call(em, args...)
	})
	wrapper.Object().Set("listener", listener)
	rec.SetHidden("fn", wrapper)
	return rec
}

func (m *module) add(this js.Object, args []js.Value, prepend, once bool) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	nameArg := js.Value{}
	listener := js.Value{}
	if len(args) > 0 {
		nameArg = args[0]
	}
	if len(args) > 1 {
		listener = args[1]
	}
	if !listener.Valid() || !listener.IsFunction() {
		got := "undefined"
		if listener.Valid() {
			got = listener.Str()
		}
		return js.Value{}, errors.New(`The "listener" argument must be of type function. Received ` + got)
	}
	name := m.nameKey(nameArg)

	if name != "newListener" {
		if recs := m.records(this, "newListener"); len(recs) > 0 {
			payload := []js.Value{nameArg, listener}
			if _, err := m.emitRaw(this, "newListener", payload); err != nil {
				return js.Value{}, err
			}
		}
	}

	rec := m.makeRecord(this, name, listener, once)
	recs := m.records(this, name)
	if prepend {
		recs = append([]js.Object{rec}, recs...)
	} else {
		recs = append(recs, rec)
	}
	m.write(this, name, recs)
	return this.V(), nil
}

func (m *module) removeListener(this js.Object, args []js.Value) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	if len(args) < 2 {
		return this.V(), nil
	}
	name := m.nameKey(args[0])
	target := args[1]
	recs := m.records(this, name)
	for i := len(recs) - 1; i >= 0; i-- {
		rec := recs[i]
		if m.same(rec.Get("original"), target) || m.same(rec.Get("fn"), target) {
			removed := rec
			recs = append(recs[:i], recs[i+1:]...)
			m.write(this, name, recs)
			if name != "removeListener" {
				payload := []js.Value{args[0], removed.Get("original")}
				if _, err := m.emitRaw(this, "removeListener", payload); err != nil {
					return js.Value{}, err
				}
			}
			break
		}
	}
	return this.V(), nil
}

func (m *module) removeByID(em js.Object, name string, id int) (js.Object, bool) {
	recs := m.records(em, name)
	for i, rec := range recs {
		if rec.Get("id").Int() == id {
			recs = append(recs[:i], recs[i+1:]...)
			m.write(em, name, recs)
			if name != "removeListener" {
				payload := []js.Value{m.e.JS.Str(name), rec.Get("original")}
				_, _ = m.emitRaw(em, "removeListener", payload)
			}
			return rec, true
		}
	}
	return js.Object{}, false
}

func (m *module) removeAll(this js.Object, args []js.Value) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	if len(args) > 0 {
		m.removeAllFor(this, m.nameKey(args[0]))
		return this.V(), nil
	}
	lists := m.lists(this)
	for _, name := range lists.OwnKeys() {
		if name == "removeListener" {
			continue
		}
		m.removeAllFor(this, name)
	}
	m.removeAllFor(this, "removeListener")
	return this.V(), nil
}

func (m *module) removeAllFor(em js.Object, name string) {
	recs := m.records(em, name)
	if len(recs) == 0 {
		return
	}
	m.write(em, name, nil)
	for i := len(recs) - 1; i >= 0; i-- {
		if name == "removeListener" {
			continue
		}
		payload := []js.Value{m.e.JS.Str(name), recs[i].Get("original")}
		_, _ = m.emitRaw(em, "removeListener", payload)
	}
}

func (m *module) emit(this js.Object, args []js.Value) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	if len(args) < 1 {
		return js.Value{}, errors.New(`The "eventName" argument must be of type string or symbol. Received undefined`)
	}
	name := m.nameKey(args[0])
	payload := args[1:]
	recs := m.records(this, name)

	if name == "error" && len(recs) == 0 {
		if len(payload) == 0 {
			return js.Value{}, errors.New("Unhandled error. (undefined)")
		}
		if payload[0].IsError() {
			// Rethrow the original error object so name/message identity survive.
			_, err := m.rethrow.Call(js.Object{}, payload[0])
			if err != nil {
				return js.Value{}, err
			}
			return js.Value{}, errors.New("Unhandled error.")
		}
		return js.Value{}, errors.New("Unhandled error. (" + describe(payload[0]) + ")")
	}
	return m.emitRaw(this, name, payload)
}

func (m *module) emitRaw(em js.Object, name string, payload []js.Value) (js.Value, error) {
	recs := m.records(em, name)
	for _, rec := range recs {
		fn := rec.Get("fn")
		if !fn.Valid() || !fn.IsFunction() {
			continue
		}
		res, err := fn.Object().Call(em, payload...)
		if err != nil {
			return js.Value{}, err
		}
		if res.Valid() && res.IsObject() && m.captureEnabled(em) {
			_, _ = m.capture.Call(js.Object{}, res, em.V(), m.e.JS.Str(name), m.e.JS.Arr(payload).V())
		}
	}
	return m.e.JS.Bool(len(recs) > 0), nil
}

// captureEnabled reports whether rejected promises returned by listeners should
// be forwarded to the rejection handler (EventEmitter captureRejections).
func (m *module) captureEnabled(em js.Object) bool {
	if v := em.Get("captureRejections"); v.Valid() && v.IsBool() {
		return v.Bool()
	}
	v := m.exports.Get("captureRejections")
	return v.Valid() && v.IsBool() && v.Bool()
}

func (m *module) listeners(this js.Object, args []js.Value, raw bool) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	name := ""
	if len(args) > 0 {
		name = m.nameKey(args[0])
	}
	recs := m.records(this, name)
	out := m.e.JS.Arr(nil)
	for i, rec := range recs {
		if raw {
			out.SetAt(i, rec.Get("fn"))
		} else {
			out.SetAt(i, rec.Get("original"))
		}
	}
	return out.V(), nil
}

func (m *module) countOf(em js.Object, name string, filter []js.Value) js.Value {
	var want js.Value
	if len(filter) > 0 && filter[0].IsFunction() {
		want = filter[0]
	}
	// EventTarget and AbortSignal keep listeners in JavaScript, so ask the
	// DOM helper when the object is not an EventEmitter (which has state).
	if !em.Has(stateKey) {
		if vals, ok := m.targetListeners(em, name); ok {
			n := 0
			for _, v := range vals {
				if !want.Valid() || m.same(v, want) {
					n++
				}
			}
			return m.e.JS.Num(float64(n))
		}
	}
	recs := m.lists(em).Get(name)
	if !recs.IsArray() {
		return m.e.JS.Num(0)
	}
	arr := recs.Object()
	n := arr.Length()
	if !want.Valid() {
		return m.e.JS.Num(float64(n))
	}
	count := 0
	for i := 0; i < n; i++ {
		item := arr.At(i)
		if !item.IsObject() {
			continue
		}
		rec := item.Object()
		if m.same(rec.Get("original"), want) || m.same(rec.Get("fn"), want) {
			count++
		}
	}
	return m.e.JS.Num(float64(count))
}

// hasAny reports whether any of the named events has a listener. Unlike
// countOf it never creates an emitter's state, so the common no-listener case
// allocates nothing.
func (m *module) hasAny(em js.Object, args []js.Value) js.Value {
	if len(args) == 0 || !args[0].IsArray() || !em.Has(stateKey) {
		return m.e.JS.Bool(false)
	}
	st := em.Get(stateKey)
	if !st.IsObject() {
		return m.e.JS.Bool(false)
	}
	lists := st.Object().Get("list")
	if !lists.IsObject() {
		return m.e.JS.Bool(false)
	}
	table := lists.Object()
	names := args[0].Object()
	for i, n := 0, names.Length(); i < n; i++ {
		name := names.At(i)
		if !name.IsString() {
			continue
		}
		recs := table.Get(name.Str())
		if recs.IsArray() && recs.Object().Length() > 0 {
			return m.e.JS.Bool(true)
		}
	}
	return m.e.JS.Bool(false)
}

// targetListeners returns the callbacks registered on a DOM-style target.
// The helper is installed by the events prelude and returns null for objects
// that are not EventTargets, so ok distinguishes "no listeners" from "not a
// target".
func (m *module) targetListeners(em js.Object, name string) ([]js.Value, bool) {
	fn := m.e.Global().Get("__godeListenerList")
	if !fn.Valid() || !fn.IsFunction() {
		return nil, false
	}
	res, err := fn.Object().Call(js.Object{}, em.V(), m.e.JS.Str(name))
	if err != nil || !res.Valid() || !res.IsArray() {
		return nil, false
	}
	arr := res.Object()
	out := make([]js.Value, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		v := arr.At(i)
		dup := false
		for _, seen := range out {
			if m.same(seen, v) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, v)
		}
	}
	return out, true
}

func (m *module) listenerCount(this js.Object, args []js.Value) (js.Value, error) {
	if !this.IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	name := ""
	if len(args) > 0 {
		name = m.nameKey(args[0])
	}
	return m.countOf(this, name, args[1:]), nil
}

func (m *module) eventNames(this js.Object) js.Value {
	lists := m.lists(this)
	keys := lists.OwnKeys()
	stringsOut := m.e.JS.Arr(nil)
	symbolsOut := m.e.JS.Arr(nil)
	si, yi := 0, 0
	for _, name := range keys {
		v := lists.Get(name)
		if v.IsArray() && v.Object().Length() > 0 {
			if sym := m.nameValue(name); sym.Valid() {
				symbolsOut.SetAt(yi, sym)
				yi++
			} else {
				stringsOut.SetAt(si, m.e.JS.Str(name))
				si++
			}
		}
	}
	// eventNames() lists string events first, then symbols, in insertion order.
	out := m.e.JS.Arr(nil)
	for i := 0; i < si; i++ {
		out.SetAt(i, stringsOut.At(i))
	}
	for i := 0; i < yi; i++ {
		out.SetAt(si+i, symbolsOut.At(i))
	}
	return out.V()
}

func (m *module) checkMax(v js.Value) (float64, error) {
	if !v.IsNumber() {
		return 0, m.throwNamed("TypeError", `The "n" argument must be of type number. Received `+describe(v))
	}
	n := v.Float()
	if math.IsNaN(n) || n < 0 {
		return 0, m.throwNamed("RangeError",
			`The value of "n" is out of range. It must be a non-negative number. Received `+describe(v))
	}
	return n, nil
}

func (m *module) setMax(em js.Object, args []js.Value) error {
	if len(args) < 1 {
		return m.throwNamed("TypeError", `The "n" argument must be of type number. Received undefined`)
	}
	n, err := m.checkMax(args[0])
	if err != nil {
		return err
	}
	m.state(em).Set("max", m.e.JS.Num(n))
	return nil
}

func (m *module) oncePromise(args []js.Value) (js.Value, error) {
	if len(args) < 2 || !args[0].IsObject() {
		return js.Value{}, errors.New(`The "emitter" argument must be of type EventEmitter. Received undefined`)
	}
	em := args[0].Object()
	name := m.nameKey(args[1])

	var resolve, reject js.Value
	holder := m.e.JS.Obj()
	holder.SetFunc("executor", func(this js.Object, a []js.Value) (js.Value, error) {
		if len(a) > 0 {
			resolve = a[0]
			resolve.Protect()
		}
		if len(a) > 1 {
			reject = a[1]
			reject.Protect()
		}
		return m.e.JS.Undefined(), nil
	})
	p, err := m.promise.Call(js.Object{}, holder.Get("executor"))
	if err != nil {
		return js.Value{}, err
	}

	id := m.nextID()
	handler := m.fnValue("onceHandler", func(this js.Object, hargs []js.Value) (js.Value, error) {
		m.removeByID(em, name, id)
		if name == "error" {
			if reject.Valid() {
				arg := m.e.JS.Undefined()
				if len(hargs) > 0 {
					arg = hargs[0]
				}
				_, _ = reject.Object().Call(js.Object{}, arg)
			}
			return m.e.JS.Undefined(), nil
		}
		arr := m.e.JS.Arr(hargs)
		if resolve.Valid() {
			_, _ = resolve.Object().Call(js.Object{}, arr.V())
		}
		return m.e.JS.Undefined(), nil
	})

	rec := m.e.JS.Obj()
	rec.SetHidden("id", m.e.JS.Num(float64(id)))
	rec.SetHidden("once", m.e.JS.Bool(true))
	rec.SetHidden("original", handler)
	rec.SetHidden("fn", handler)
	recs := m.records(em, name)
	recs = append(recs, rec)
	m.write(em, name, recs)
	return p, nil
}

func describe(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return "'" + v.Str() + "'"
	case v.IsBool():
		if v.Bool() {
			return "true"
		}
		return "false"
	case v.IsNumber():
		s, _ := v.JSON(0)
		return s
	}
	return "[object Object]"
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"events.CustomEvent",
		"events.CustomEvent.detail",
		"events.Event",
		"events.Event#composedPath",
		"events.Event#initEvent",
		"events.Event#preventDefault",
		"events.Event#stopImmediatePropagation",
		"events.Event#stopPropagation",
		"events.Event.bubbles",
		"events.Event.cancelBubble",
		"events.Event.cancelable",
		"events.Event.composed",
		"events.Event.currentTarget",
		"events.Event.defaultPrevented",
		"events.Event.eventPhase",
		"events.Event.isTrusted",
		"events.Event.returnValue",
		"events.Event.srcElement",
		"events.Event.target",
		"events.Event.timeStamp",
		"events.Event.type",
		"events.EventEmitter",
		"events.EventEmitter#addListener",
		"events.EventEmitter#emit",
		"events.EventEmitter#eventNames",
		"events.EventEmitter#getMaxListeners",
		"events.EventEmitter#listenerCount",
		"events.EventEmitter#listeners",
		"events.EventEmitter#off",
		"events.EventEmitter#on",
		"events.EventEmitter#once",
		"events.EventEmitter#prependListener",
		"events.EventEmitter#prependOnceListener",
		"events.EventEmitter#rawListeners",
		"events.EventEmitter#removeAllListeners",
		"events.EventEmitter#removeListener",
		"events.EventEmitter#setMaxListeners",
		"events.EventEmitterAsyncResource",
		"events.EventEmitterAsyncResource#emitDestroy",
		"events.EventEmitterAsyncResource.asyncId",
		"events.EventEmitterAsyncResource.asyncResource",
		"events.EventEmitterAsyncResource.triggerAsyncId",
		"events.EventTarget",
		"events.EventTarget#addEventListener",
		"events.EventTarget#dispatchEvent",
		"events.EventTarget#removeEventListener",
		"events.NodeEventTarget",
		"events.NodeEventTarget#addListener",
		"events.NodeEventTarget#emit",
		"events.NodeEventTarget#eventNames",
		"events.NodeEventTarget#getMaxListeners",
		"events.NodeEventTarget#listenerCount",
		"events.NodeEventTarget#off",
		"events.NodeEventTarget#on",
		"events.NodeEventTarget#once",
		"events.NodeEventTarget#removeAllListeners",
		"events.NodeEventTarget#removeListener",
		"events.NodeEventTarget#setMaxListeners",
		"events.addAbortListener",
		"events.captureRejectionSymbol",
		"events.captureRejections",
		"events.defaultMaxListeners",
		"events.errorMonitor",
		"events.getEventListeners",
		"events.getMaxListeners",
		"events.listenerCount",
		"events.on",
		"events.once",
		"events.setMaxListeners",
	}
}
