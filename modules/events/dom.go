package events

import (
	"godejs/internal/js"
)

// domPrelude builds the web-style half of the events module. Node exposes
// Event, EventTarget and CustomEvent even on the server; NodeEventTarget is
// an EventTarget with the EventEmitter methods mixed in.
//
// Everything here is deliberately plain JavaScript: the listener registries,
// event dispatch and async iterators are closures, so a host round trip is
// only needed for EventEmitter's own emit/on implementation.
const domPrelude = `(function (EventEmitter) {
  'use strict';

  var kDispose = (typeof Symbol.dispose === 'symbol') ? Symbol.dispose : Symbol.for('nodejs.dispose');
  var kInspect = Symbol.for('nodejs.util.inspect.custom');

  function disposable(fn) {
    var d = {};
    d[kDispose] = fn;
    return d;
  }

  function abortedError(reason) {
    var err = new Error('The operation was aborted');
    err.name = 'AbortError';
    err.code = 'ABORT_ERR';
    if (reason !== undefined) err.cause = reason;
    return err;
  }

  function describeValue(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    var t = typeof value;
    if (t === 'number' || t === 'boolean' || t === 'bigint') return 'type ' + t + ' (' + String(value) + ')';
    if (t === 'string') return 'type string (\'' + value + '\')';
    if (t === 'symbol') return 'type symbol';
    if (t === 'function') return 'function';
    return 'an instance of ' + (value.constructor && value.constructor.name ? value.constructor.name : 'Object');
  }

  function invalidArgType(name, expected, actual) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + describeValue(actual));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function isAbortSignal(value) {
    return value !== null && typeof value === 'object' &&
      typeof value.aborted === 'boolean' &&
      typeof value.addEventListener === 'function' &&
      typeof value.removeEventListener === 'function';
  }

  class Event {
    constructor(type, init) {
      // Node rejects a missing type and symbols (String(Symbol()) would not
      // throw but the spec requires a TypeError).
      if (arguments.length === 0 || type === undefined || typeof type === 'symbol') {
        throw new TypeError('The "type" argument must be specified and must not be a symbol');
      }
      if (init !== undefined && init !== null && typeof init !== 'object') {
        throw invalidArgType('options', 'of type object', init);
      }
      init = init || {};
      this._type = String(type);
      this._bubbles = !!init.bubbles;
      this._cancelable = !!init.cancelable;
      this._composed = !!init.composed;
      this._defaultPrevented = false;
      this._target = null;
      this._currentTarget = null;
      this._cancelBubble = false;
      this._stopImmediate = false;
      this._phase = 0;
      this._timeStamp = Date.now();
      this._isTrusted = false;
      this._path = [];
    }
    get type() { return this._type; }
    get bubbles() { return this._bubbles; }
    get cancelable() { return this._cancelable; }
    get composed() { return this._composed; }
    get defaultPrevented() { return this._defaultPrevented; }
    get target() { return this._target; }
    get currentTarget() { return this._currentTarget; }
    get srcElement() { return this._target; }
    get cancelBubble() { return this._cancelBubble; }
    set cancelBubble(v) { if (v) this._cancelBubble = true; }
    get returnValue() { return !this._defaultPrevented; }
    set returnValue(v) { if (v === false) this.preventDefault(); }
    get eventPhase() { return this._phase; }
    get isTrusted() { return this._isTrusted; }
    get timeStamp() { return this._timeStamp; }
    composedPath() { return this._path.slice(); }
    preventDefault() {
      if (this._cancelable) this._defaultPrevented = true;
    }
    stopPropagation() { this._cancelBubble = true; }
    stopImmediatePropagation() {
      this._cancelBubble = true;
      this._stopImmediate = true;
    }
    initEvent(type, bubbles, cancelable) {
      this._type = String(type);
      this._bubbles = !!bubbles;
      this._cancelable = !!cancelable;
      this._defaultPrevented = false;
      this._cancelBubble = false;
      this._stopImmediate = false;
    }
    [kInspect](depth, options, inspect) {
      // Node's event_target custom inspector: an exhausted depth renders just
      // the class name, otherwise the public event fields are shown.
      if (typeof depth === 'number' && depth < 0) return this.constructor.name;
      if (typeof inspect !== 'function') return this.constructor.name;
      var opts = Object.assign({}, options);
      opts.depth = (options && options.depth !== null && options.depth !== undefined) ?
        options.depth - 1 : null;
      return this.constructor.name + ' ' + inspect({
        type: this.type,
        defaultPrevented: this.defaultPrevented,
        cancelable: this.cancelable,
        timeStamp: this.timeStamp
      }, opts);
    }
  }

  class CustomEvent extends Event {
    constructor(type, init = undefined) {
      super(type, init);
      this._detail = init && init.detail !== undefined ? init.detail : null;
    }
    get detail() { return this._detail; }
  }

  // DOM event phase constants are static on both constructors (CustomEvent
  // inherits them through its prototype chain).
  Event.NONE = 0;
  Event.CAPTURING_PHASE = 1;
  Event.AT_TARGET = 2;
  Event.BUBBLING_PHASE = 3;

  var kListeners = Symbol('gode.eventtarget.listeners');

  class EventTarget {
    constructor() {
      Object.defineProperty(this, kListeners, {
        value: new Map(), writable: true, configurable: true
      });
    }
    addEventListener(type, listener, options) {
      if (listener === null || listener === undefined) return;
      if (typeof listener !== 'function' && typeof listener !== 'object') {
        throw new TypeError('The "listener" argument must be of type function or an object');
      }
      options = options || {};
      var key = String(type);
      var list = this[kListeners].get(key);
      if (!list) { list = []; this[kListeners].set(key, list); }
      for (var i = 0; i < list.length; i++) {
        if (list[i].callback === listener && list[i].capture === !!options.capture) return;
      }
      var entry = {
        callback: listener,
        capture: !!options.capture,
        once: !!options.once,
        resistStopPropagation: !!options.resistStopPropagation,
        signal: options.signal
      };
      list.push(entry);
      if (options.signal && typeof options.signal.addEventListener === 'function') {
        var self = this;
        entry.abort = function () {
          self.removeEventListener(key, listener, { capture: entry.capture });
        };
        if (options.signal.aborted) {
          entry.abort();
        } else {
          options.signal.addEventListener('abort', entry.abort, { once: true });
        }
      }
    }
    removeEventListener(type, listener, options) {
      options = options || {};
      var list = this[kListeners].get(String(type));
      if (!list) return;
      for (var i = 0; i < list.length; i++) {
        if (list[i].callback === listener && list[i].capture === !!options.capture) {
          if (list[i].signal && list[i].abort && typeof list[i].signal.removeEventListener === 'function') {
            list[i].signal.removeEventListener('abort', list[i].abort);
          }
          list.splice(i, 1);
          return;
        }
      }
    }
    dispatchEvent(event) {
      if (!(event instanceof Event)) {
        throw new TypeError('The "event" argument must be an instance of Event');
      }
      var list = (this[kListeners].get(event.type) || []).slice();
      event._target = this;
      event._currentTarget = this;
      event._phase = 2;
      event._path = [this];
      event._stopImmediate = false;
      for (var i = 0; i < list.length; i++) {
        var entry = list[i];
        if (event._stopImmediate && !entry.resistStopPropagation) continue;
        if (entry.once) {
          this.removeEventListener(event.type, entry.callback, { capture: entry.capture });
        }
        try {
          if (typeof entry.callback === 'function') entry.callback.call(this, event);
          else if (typeof entry.callback.handleEvent === 'function') entry.callback.handleEvent(event);
        } catch (err) {
          if (typeof queueMicrotask === 'function') queueMicrotask(function () { throw err; });
          else Promise.resolve().then(function () { throw err; });
        }
      }
      event._currentTarget = null;
      event._phase = 0;
      // composedPath() is only populated during dispatch, as in the DOM.
      event._path = [];
      return !event._defaultPrevented;
    }
  }

  // Node brands each class so Object.prototype.toString reports the same
  // tags as the platform ([object Event], [object CustomEvent], ...).
  Object.defineProperty(Event.prototype, Symbol.toStringTag, { value: 'Event', configurable: true });
  Object.defineProperty(CustomEvent.prototype, Symbol.toStringTag, { value: 'CustomEvent', configurable: true });
  Object.defineProperty(EventTarget.prototype, Symbol.toStringTag, { value: 'EventTarget', configurable: true });

  class NodeEventTarget extends EventTarget {}
  Object.defineProperty(NodeEventTarget.prototype, Symbol.toStringTag, { value: 'NodeEventTarget', configurable: true });
  // Mix EventEmitter's methods into NodeEventTarget.prototype. EventEmitter
  // state is created lazily by every method, so no constructor call is needed.
  Object.getOwnPropertyNames(EventEmitter.prototype).forEach(function (name) {
    if (name === 'constructor') return;
    Object.defineProperty(NodeEventTarget.prototype, name,
      Object.getOwnPropertyDescriptor(EventEmitter.prototype, name));
  });
  Object.getOwnPropertySymbols(EventEmitter.prototype).forEach(function (sym) {
    Object.defineProperty(NodeEventTarget.prototype, sym,
      Object.getOwnPropertyDescriptor(EventEmitter.prototype, sym));
  });

  // Node's default rejection handler emits the error on the emitter.
  if (!EventEmitter.prototype[Symbol.for('nodejs.rejection')]) {
    Object.defineProperty(EventEmitter.prototype, Symbol.for('nodejs.rejection'), {
      value: function (err) {
        if (typeof this.emit === 'function') this.emit('error', err);
      },
      writable: true,
      configurable: true
    });
  }

  function agnosticAddListener(emitter, name, listener, flags) {
    if (typeof emitter.on === 'function') {
      if (flags && flags.once) emitter.once(name, listener);
      else emitter.on(name, listener);
    } else if (typeof emitter.addEventListener === 'function') {
      emitter.addEventListener(name, listener, flags);
    } else {
      throw invalidArgType('emitter', 'an instance of EventEmitter or EventTarget', emitter);
    }
  }

  function agnosticRemoveListener(emitter, name, listener, flags) {
    if (typeof emitter.removeListener === 'function') {
      emitter.removeListener(name, listener);
    } else if (typeof emitter.removeEventListener === 'function') {
      emitter.removeEventListener(name, listener, flags);
    } else {
      throw invalidArgType('emitter', 'an instance of EventEmitter or EventTarget', emitter);
    }
  }

  // Node's once() validates before it subscribes and rejects, rather than
  // throws, so callers can hand the return value straight to await.
  function once(emitter, name, options) {
    if (options === undefined) options = {};
    var signal;
    try {
      if (options === null || typeof options !== 'object') {
        throw invalidArgType('options', 'of type object', options);
      }
      signal = options.signal;
      if (signal !== undefined && !isAbortSignal(signal)) {
        throw invalidArgType('options.signal', 'an instance of AbortSignal', signal);
      }
      if (emitter === null || typeof emitter !== 'object') {
        throw invalidArgType('emitter', 'an instance of EventEmitter or EventTarget', emitter);
      }
      if (typeof emitter.on !== 'function' && typeof emitter.addEventListener !== 'function') {
        throw invalidArgType('emitter', 'an instance of EventEmitter or EventTarget', emitter);
      }
      if (signal && signal.aborted) throw abortedError(signal.reason);
    } catch (err) {
      return Promise.reject(err);
    }

    return new Promise(function (resolve, reject) {
      var settled = false;
      var listenFlags = { once: true, capture: false, resistStopPropagation: true };

      function cleanup() {
        agnosticRemoveListener(emitter, name, resolver);
        if (name !== 'error' && typeof emitter.removeListener === 'function') {
          emitter.removeListener('error', errorListener);
        }
        if (signal && typeof signal.removeEventListener === 'function') {
          signal.removeEventListener('abort', abortListener);
        }
      }
      function resolver() {
        if (settled) return;
        settled = true;
        var args = Array.prototype.slice.call(arguments);
        cleanup();
        resolve(args);
      }
      function errorListener(err) {
        if (settled) return;
        settled = true;
        cleanup();
        reject(err);
      }
      function abortListener() {
        if (settled) return;
        settled = true;
        cleanup();
        reject(abortedError(signal.reason));
      }

      try {
        agnosticAddListener(emitter, name, resolver, listenFlags);
        if (name !== 'error' && typeof emitter.once === 'function') {
          emitter.once('error', errorListener);
        }
        if (signal) {
          signal.addEventListener('abort', abortListener, { once: true, resistStopPropagation: true });
        }
      } catch (err) {
        settled = true;
        reject(err);
      }
    });
  }

  function listenerCallbacks(target, type) {
    if (target === null || typeof target !== 'object') return null;
    if (String(type) === 'abort' && Array.isArray(target._listeners)) {
      return target._listeners.map(function (entry) { return entry.fn; });
    }
    var map = target[kListeners];
    if (map && typeof map.get === 'function') {
      var list = map.get(String(type)) || [];
      return list.map(function (entry) { return entry.callback; });
    }
    return null;
  }

  Object.defineProperty(globalThis, '__godeListenerList', {
    value: listenerCallbacks, writable: true, configurable: true, enumerable: false
  });

  function on(emitter, name, options) {
    options = options || {};
    var queue = [];
    var waiters = [];
    var failure = null;
    var finished = false;
    var signal = options.signal;

    function cleanup() {
      emitter.removeListener(name, handle);
      if (name !== 'error') emitter.removeListener('error', handleError);
      if (signal && typeof signal.removeEventListener === 'function') {
        signal.removeEventListener('abort', onAbort);
      }
    }
    function settle() {
      while (waiters.length) {
        var w = waiters.shift();
        if (failure) w.reject(failure);
        else if (queue.length) w.resolve({ value: queue.shift(), done: false });
        else w.resolve({ value: undefined, done: true });
      }
    }
    function handle() {
      queue.push(Array.prototype.slice.call(arguments));
      settle();
    }
    function handleError(err) {
      failure = err;
      cleanup();
      settle();
    }
    function onAbort() {
      failure = abortedError();
      cleanup();
      settle();
    }
    if (signal && signal.aborted) {
      failure = abortedError();
      finished = true;
    } else {
      emitter.on(name, handle);
      if (name !== 'error') emitter.on('error', handleError);
      if (signal && typeof signal.addEventListener === 'function') {
        signal.addEventListener('abort', onAbort, { once: true });
      }
    }

    var iterator = {
      next: function () {
        if (queue.length) return Promise.resolve({ value: queue.shift(), done: false });
        if (failure) return Promise.reject(failure);
        if (finished) return Promise.resolve({ value: undefined, done: true });
        var resolve, reject;
        var promise = new Promise(function (res, rej) { resolve = res; reject = rej; });
        waiters.push({ resolve: resolve, reject: reject });
        return promise;
      },
      return: function () {
        finished = true;
        cleanup();
        while (waiters.length) waiters.shift().resolve({ value: undefined, done: true });
        return Promise.resolve({ value: undefined, done: true });
      },
      throw: function (err) {
        finished = true;
        cleanup();
        return Promise.reject(err);
      }
    };
    iterator[Symbol.asyncIterator] = function () { return iterator; };
    return iterator;
  }

  function addAbortListener(signal, listener) {
    if (signal === undefined || signal === null) {
      throw new TypeError('The "signal" argument must be an instance of AbortSignal. Received undefined');
    }
    if (typeof listener !== 'function') {
      throw new TypeError('The "listener" argument must be of type function. Received undefined');
    }
    var active = true;
    function wrapped() {
      if (!active) return;
      listener.call(undefined, new Event('abort'));
    }
    if (signal.aborted) {
      if (typeof queueMicrotask === 'function') queueMicrotask(wrapped);
      else Promise.resolve().then(wrapped);
      return disposable(function () { active = false; });
    }
    if (typeof signal.addEventListener === 'function') {
      signal.addEventListener('abort', wrapped, { once: true });
    }
    return disposable(function () {
      active = false;
      if (typeof signal.removeEventListener === 'function') {
        signal.removeEventListener('abort', wrapped);
      }
    });
  }

  return {
    Event: Event,
    CustomEvent: CustomEvent,
    EventTarget: EventTarget,
    NodeEventTarget: NodeEventTarget,
    on: on,
    once: once,
    addAbortListener: addAbortListener
  };
})`

// installDOM installs the DOM-style classes and the Promise/async-iterator
// helpers built on top of the EventEmitter constructor.
func (m *module) installDOM() {
	fn := m.mustEval(domPrelude, "events.dom")
	exports, err := fn.Call(js.Object{}, m.ctor.V())
	if err != nil {
		panic("godejs events: " + err.Error())
	}
	o := exports.Object()
	m.exports.Set("Event", o.Get("Event"))
	m.exports.Set("CustomEvent", o.Get("CustomEvent"))
	m.exports.Set("EventTarget", o.Get("EventTarget"))
	m.exports.Set("NodeEventTarget", o.Get("NodeEventTarget"))
	m.exports.Set("on", o.Get("on"))
	m.exports.Set("once", o.Get("once"))
	m.exports.Set("addAbortListener", o.Get("addAbortListener"))
	// Node installs Event, EventTarget and CustomEvent as globals, not only
	// as events exports. Keep them non-enumerable, like the real runtime.
	m.e.Global().SetHidden("Event", o.Get("Event"))
	m.e.Global().SetHidden("EventTarget", o.Get("EventTarget"))
	m.e.Global().SetHidden("CustomEvent", o.Get("CustomEvent"))
}
