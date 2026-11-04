package timers

import (
	"strconv"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("timers", Install)
	env.Register("timers/promises", installPromises)
}

type module struct {
	e *env.Env
	// ids maps a timer's primitive id to the Timeout that produced it, the
	// way Node's knownTimersById does for clearTimeout(number|string).
	ids map[int]js.Value
}

// Install builds Node's timers module and installs the timer globals.
//
// The timer objects are created in JavaScript because Node exposes their
// internals (_idleTimeout, _onTimeout, _destroyed, the prototype methods and
// Symbol.toPrimitive) as ordinary properties; Go only schedules, cancels and
// re-arms them.
func Install(e *env.Env) js.Object {
	m := &module{e: e, ids: map[int]js.Value{}}

	host := e.JS.Obj()
	host.SetFunc("schedule", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.schedule(args, false)
	})
	host.SetFunc("immediate", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.schedule(args, true)
	})
	host.SetFunc("clear", func(this js.Object, args []js.Value) (js.Value, error) {
		m.e.ClearTimer(int(argAt(args, 0).Float()))
		return e.JS.Undefined(), nil
	})
	host.SetFunc("refresh", func(this js.Object, args []js.Value) (js.Value, error) {
		delay := argAt(args, 1).Float()
		// A refreshed timer behaves like a freshly scheduled one: Node inserts
		// it again, so a sub-millisecond duration is truncated for the list.
		m.e.RefreshTimer(int(argAt(args, 0).Float()), delay)
		return e.JS.Undefined(), nil
	})
	// remember/lookup back Symbol.toPrimitive's number->Timeout map, which is
	// what makes clearTimeout(number) and clearTimeout("5") work in Node.
	host.SetFunc("remember", func(this js.Object, args []js.Value) (js.Value, error) {
		id := int(argAt(args, 0).Float())
		val := argAt(args, 1)
		if val.IsObject() {
			m.ids[id] = val.Protect()
		}
		return e.JS.Undefined(), nil
	})
	host.SetFunc("forget", func(this js.Object, args []js.Value) (js.Value, error) {
		delete(m.ids, int(argAt(args, 0).Float()))
		return e.JS.Undefined(), nil
	})
	host.SetFunc("lookup", func(this js.Object, args []js.Value) (js.Value, error) {
		if v, ok := m.ids[m.timerKey(argAt(args, 0))]; ok {
			return v, nil
		}
		return e.JS.Undefined(), nil
	})
	host.SetFunc("promises", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.Exports("timers/promises").V(), nil
	})
	host.SetFunc("warn", func(this js.Object, args []js.Value) (js.Value, error) {
		m.warn(argAt(args, 0).Str(), argAt(args, 1).Str())
		return e.JS.Undefined(), nil
	})

	built, err := e.JS.Eval(prelude, "node:timers")
	if err != nil {
		panic("godejs timers: " + err.Error())
	}
	res, err := built.Object().Call(js.Object{}, host.V())
	if err != nil {
		panic("godejs timers: " + err.Error())
	}
	exports := res.Object()
	for _, name := range []string{
		"setTimeout", "clearTimeout", "setImmediate",
		"clearImmediate", "setInterval", "clearInterval",
	} {
		e.Global().Set(name, exports.Get(name))
	}
	return exports
}

// schedule creates a timer from the JS-side Timeout/Immediate constructor.
// Arguments are [callback, delayMs, repeat, argsArray].
func (m *module) schedule(args []js.Value, immediate bool) (js.Value, error) {
	cb := argAt(args, 0)
	delay := 0.0
	repeat := false
	if !immediate {
		if v := argAt(args, 1); v.IsNumber() {
			delay = v.Float()
		}
		repeat = argAt(args, 2).Bool()
	}
	var rest []js.Value
	if a := argAt(args, 3); a.IsArray() {
		arr := a.Object()
		n := arr.Length()
		if n > 0 {
			rest = make([]js.Value, n)
			for i := 0; i < n; i++ {
				rest[i] = arr.At(i)
			}
		}
	}
	id := m.e.SetTimeout(cb, delay, repeat, rest)
	return m.e.JS.Num(float64(id)), nil
}

// timerKey normalises the id forms Node accepts for clearTimeout(number) and
// clearTimeout("5") to one map key.
func (m *module) timerKey(v js.Value) int {
	switch {
	case v.IsNumber():
		return int(v.Float())
	case v.IsString():
		n, err := strconv.Atoi(strings.TrimSpace(v.Str()))
		if err == nil {
			return n
		}
	}
	return 0
}

// warn mirrors process.emitWarning for the timer duration warnings.
func (m *module) warn(kind, message string) {
	proc := m.e.Global().Get("process")
	if !proc.IsObject() {
		return
	}
	emit := proc.Object().Get("emitWarning")
	if !emit.IsFunction() {
		return
	}
	_, _ = emit.Object().Call(proc.Object(), m.e.JS.Str(message), m.e.JS.Str(kind))
}

// installPromises builds the Promise-based timer API on top of the timers
// module, so `promisify(timers.setTimeout)` and `timers/promises.setTimeout`
// are the same function.
func installPromises(e *env.Env) js.Object {
	base := e.Exports("timers")
	v, err := e.JS.Eval(promisesPrelude, "node:timers/promises")
	if err != nil {
		panic("timers/promises: " + err.Error())
	}
	res, err := v.Object().Call(js.Object{}, base.V())
	if err != nil {
		panic("timers/promises: " + err.Error())
	}
	return res.Object()
}

func argAt(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"timers.setTimeout",
		"timers.clearTimeout",
		"timers.setInterval",
		"timers.clearInterval",
		"timers.setImmediate",
		"timers.clearImmediate",
		"timers.Timeout",
		"timers.Timeout#ref",
		"timers.Timeout#unref",
		"timers.Timeout#hasRef",
		"timers.Timeout#refresh",
		"timers.Timeout#close",
		"timers.Timeout#[Symbol.toPrimitive]",
		"timers.Immediate",
		"timers.Immediate#ref",
		"timers.Immediate#unref",
		"timers.Immediate#hasRef",
		"timers.Timeout#[Symbol.dispose]",
		"timers.Immediate#[Symbol.dispose]",
		"timers.scheduler",
		"timers.wait",
		"timers.yield",
	}
}

// prelude implements lib/timers.js and lib/internal/timers.js's Timeout and
// Immediate on top of four host callbacks: schedule, immediate, clear and
// refresh (plus the primitive map and promise helpers).
const prelude = `(function (host) {
  'use strict';

  var TIMEOUT_MAX = 2147483647;
  var warnedOverflow = false;
  var warnedNegative = false;
  var warnedNaN = false;

  function describe(value) {
    if (value === null) return 'null';
    var type = typeof value;
    if (type === 'undefined') return 'undefined';
    if (type === 'string') {
      return 'type string (' + "'" + value.replace(/\\/g, '\\\\').replace(/'/g, "\\'") + "'" + ')';
    }
    if (type === 'number') return 'type number (' + (Object.is(value, -0) ? '-0' : String(value)) + ')';
    if (type === 'boolean') return 'type boolean (' + value + ')';
    if (type === 'bigint') return 'type bigint (' + value + 'n)';
    if (type === 'symbol') return 'type symbol (' + String(value) + ')';
    if (type === 'function') return 'function ' + (value.name || '<anonymous>');
    var name = value.constructor && value.constructor.name;
    return 'an instance of ' + (name || 'Object');
  }

  function invalidArgType(name, expected, value) {
    var err = new TypeError(
      'The "' + name + '" argument must be of type ' + expected +
      '. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function validateFunction(value, name) {
    if (typeof value !== 'function') throw invalidArgType(name, 'function', value);
  }

  function hide(target, key, value) {
    Object.defineProperty(target, key, {
      value: value, writable: true, enumerable: false, configurable: true
    });
  }

  function defineMethod(proto, name, fn) {
    Object.defineProperty(proto, name, {
      value: fn, writable: true, enumerable: false, configurable: true
    });
  }

  function Timeout(callback, after, args, isRepeat, isRefed) {
    if (after === undefined) {
      after = 1;
    } else {
      after *= 1;
    }
    var delay;
    if (!(after >= 1 && after <= TIMEOUT_MAX)) {
      if (after > TIMEOUT_MAX && !warnedOverflow) {
        warnedOverflow = true;
        host.warn('TimeoutOverflowWarning', after +
          ' does not fit into a 32-bit signed integer.' +
          '\nTimeout duration was set to 1.');
      } else if (after < 0 && !warnedNegative) {
        warnedNegative = true;
        host.warn('TimeoutNegativeWarning', after +
          ' is a negative number.\nTimeout duration was set to 1.');
      } else if (Number.isNaN(after) && !warnedNaN) {
        warnedNaN = true;
        host.warn('TimeoutNaNWarning', after +
          ' is not a number.\nTimeout duration was set to 1.');
      }
      after = 1;
      // A clamped duration fires on the next turn rather than a full
      // millisecond later, which is what the event loop does with 0.
      delay = 0;
    } else {
      delay = Math.trunc(after);
    }
    this._idleTimeout = after;
    this._idlePrev = this;
    this._idleNext = this;
    this._idleStart = null;
    this._onTimeout = callback;
    this._timerArgs = args;
    this._repeat = isRepeat ? after : null;
    this._destroyed = false;
    // Node calls timer._onTimeout(...timer._timerArgs) as a method on the
    // Timeout, so user callbacks observe 'this' as the timer object.
    var self = this;
    hide(this, '$godeId', host.schedule(function () {
      return callback.apply(self, args || []);
    }, delay, !!isRepeat, args));
    hide(this, '$godeRefed', !!isRefed);
    hide(this, '$godePrimitive', false);
  }

  defineMethod(Timeout.prototype, 'refresh', function refresh() {
    if (this._destroyed) return this;
    var delay = Math.trunc(this._idleTimeout);
    host.refresh(this.$godeId, delay >= 1 ? delay : 1);
    this._idleStart = Date.now();
    return this;
  });

  defineMethod(Timeout.prototype, 'unref', function unref() {
    this.$godeRefed = false;
    return this;
  });

  defineMethod(Timeout.prototype, 'ref', function ref() {
    this.$godeRefed = true;
    return this;
  });

  defineMethod(Timeout.prototype, 'hasRef', function hasRef() {
    return !!this.$godeRefed;
  });

  defineMethod(Timeout.prototype, 'close', function close() {
    clearTimeout(this);
    return this;
  });

  var toPrimitive = Symbol.toPrimitive;
  if (toPrimitive) {
    defineMethod(Timeout.prototype, toPrimitive, function () {
      if (!this.$godePrimitive) {
        this.$godePrimitive = true;
        host.remember(this.$godeId, this);
      }
      return this.$godeId;
    });
  }

  var dispose = Symbol.dispose;
  if (dispose) {
    defineMethod(Timeout.prototype, dispose, function () {
      clearTimeout(this);
    });
  }

  function Immediate(callback, args) {
    this._idleNext = null;
    this._idlePrev = null;
    this._onImmediate = callback;
    this._argv = args;
    this._destroyed = false;
    var self = this;
    hide(this, '$godeId', host.immediate(function () {
      return callback.apply(self, args || []);
    }, 0, false, args));
    hide(this, '$godeRefed', true);
    hide(this, '$godePrimitive', false);
  }

  defineMethod(Immediate.prototype, 'ref', function ref() {
    this.$godeRefed = true;
    return this;
  });

  defineMethod(Immediate.prototype, 'unref', function unref() {
    this.$godeRefed = false;
    return this;
  });

  defineMethod(Immediate.prototype, 'hasRef', function hasRef() {
    return !!this.$godeRefed;
  });

  if (dispose) {
    defineMethod(Immediate.prototype, dispose, function () {
      clearImmediate(this);
    });
  }

  function unenroll(item) {
    if (item._destroyed) return;
    item._destroyed = true;
    host.forget(item.$godeId);
    host.clear(item.$godeId);
    item._idleTimeout = -1;
  }

  function setTimeout(callback, after) {
    validateFunction(callback, 'callback');
    var args = arguments.length > 2 ?
      Array.prototype.slice.call(arguments, 2) : undefined;
    return new Timeout(callback, after, args, false, true);
  }

  function setInterval(callback, repeat) {
    validateFunction(callback, 'callback');
    var args = arguments.length > 2 ?
      Array.prototype.slice.call(arguments, 2) : undefined;
    return new Timeout(callback, repeat, args, true, true);
  }

  function setImmediate(callback) {
    validateFunction(callback, 'callback');
    var args = arguments.length > 1 ?
      Array.prototype.slice.call(arguments, 1) : undefined;
    return new Immediate(callback, args);
  }

  function clearTimeout(timer) {
    if (timer !== null && timer !== undefined && timer._onTimeout) {
      timer._onTimeout = null;
      unenroll(timer);
      return;
    }
    if (typeof timer === 'number' || typeof timer === 'string') {
      var instance = host.lookup(timer);
      if (instance !== undefined) {
        instance._onTimeout = null;
        unenroll(instance);
      }
    }
  }

  function clearInterval(timer) {
    clearTimeout(timer);
  }

  function clearImmediate(immediate) {
    if (immediate === null || immediate === undefined ||
        !immediate._onImmediate || immediate._destroyed) {
      return;
    }
    immediate._destroyed = true;
    immediate._onImmediate = undefined;
    host.forget(immediate.$godeId);
    host.clear(immediate.$godeId);
  }

  var timers = {
    setTimeout: setTimeout,
    clearTimeout: clearTimeout,
    setImmediate: setImmediate,
    clearImmediate: clearImmediate,
    setInterval: setInterval,
    clearInterval: clearInterval
  };

  Object.defineProperty(timers, 'promises', {
    configurable: true,
    enumerable: true,
    get: function () { return host.promises(); }
  });

  // Node's timers module does not enumerate the classes, but they are on the
  // prototypes so keep them reachable without changing Object.keys(timers).
  Object.defineProperty(timers, 'Timeout', {
    value: Timeout, writable: true, configurable: true
  });
  Object.defineProperty(timers, 'Immediate', {
    value: Immediate, writable: true, configurable: true
  });

  // timers/promises also hangs off the timers module (the docs paths the
  // parity baseline tracks are timers.scheduler/wait/yield).
  ['scheduler', 'wait', 'yield'].forEach(function (name) {
    Object.defineProperty(timers, name, {
      configurable: true,
      enumerable: true,
      get: function () { return host.promises()[name]; }
    });
  });

  var promisifyCustom = Symbol.for('nodejs.util.promisify.custom');
  if (promisifyCustom) {
    [[setTimeout, 'setTimeout'], [setImmediate, 'setImmediate'],
     [setInterval, 'setInterval']].forEach(function (pair) {
      Object.defineProperty(pair[0], promisifyCustom, {
        configurable: true,
        enumerable: true,
        get: function () { return host.promises()[pair[1]]; }
      });
    });
  }

  return timers;
})`

// promisesPrelude implements lib/timers/promises.js. Every entry point is an
// async function, so argument validation surfaces as a rejected promise with
// Node's ERR_INVALID_ARG_TYPE code, exactly like upstream.
const promisesPrelude = `(function (timers) {
  'use strict';

  function describe(value) {
    if (value === null) return 'null';
    var type = typeof value;
    if (type === 'undefined') return 'undefined';
    if (type === 'string') {
      return 'type string (' + "'" + value.replace(/\\/g, '\\\\').replace(/'/g, "\\'") + "'" + ')';
    }
    if (type === 'number') return 'type number (' + (Object.is(value, -0) ? '-0' : String(value)) + ')';
    if (type === 'boolean') return 'type boolean (' + value + ')';
    if (type === 'bigint') return 'type bigint (' + value + 'n)';
    if (type === 'symbol') return 'type symbol (' + String(value) + ')';
    if (type === 'function') return 'function ' + (value.name || '<anonymous>');
    var name = value.constructor && value.constructor.name;
    return 'an instance of ' + (name || 'Object');
  }

  function invalidArgType(name, expected, value) {
    var err = new TypeError(
      'The "' + name + '" argument must be of type ' + expected +
      '. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function validateNumber(value, name) {
    if (typeof value !== 'number' || Number.isNaN(value)) {
      throw invalidArgType(name, 'number', value);
    }
  }

  function validateBoolean(value, name) {
    if (typeof value !== 'boolean') throw invalidArgType(name, 'boolean', value);
  }

  function validateObject(value, name) {
    if (value === null || typeof value !== 'object') {
      throw invalidArgType(name, 'Object', value);
    }
  }

  function validateAbortSignal(signal, name) {
    if (signal !== undefined &&
        (signal === null || typeof signal !== 'object' || !('aborted' in signal))) {
      throw invalidArgType(name, 'AbortSignal', signal);
    }
  }

  function validateOptions(options) {
    if (options === undefined) return;
    validateObject(options, 'options');
    if (options.signal !== undefined) validateAbortSignal(options.signal, 'options.signal');
    if (options.ref !== undefined) validateBoolean(options.ref, 'options.ref');
  }

  function abortError(signal) {
    var err = new Error('The operation was aborted');
    err.name = 'AbortError';
    err.code = 'ABORT_ERR';
    if (signal) err.cause = signal.reason;
    return err;
  }

  function withTimer(start, options) {
    return new Promise(function (resolve, reject) {
      var signal = options && options.signal;
      var timer;
      var settled = false;
      function cleanup() {
        if (signal && typeof signal.removeEventListener === 'function') {
          signal.removeEventListener('abort', onAbort);
        }
      }
      function onAbort() {
        if (settled) return;
        settled = true;
        if (timer !== undefined && typeof timer.close === 'function') timer.close();
        cleanup();
        reject(abortError(signal));
      }
      if (signal) {
        if (signal.aborted) return reject(abortError(signal));
        if (typeof signal.addEventListener === 'function') {
          signal.addEventListener('abort', onAbort);
        }
      }
      timer = start(function (value) {
        if (settled) return;
        settled = true;
        cleanup();
        resolve(value);
      });
      if (options && options.ref === false && timer && typeof timer.unref === 'function') {
        timer.unref();
      }
    });
  }

  async function setTimeoutPromise(delay, value, options) {
    validateNumber(delay, 'delay');
    validateOptions(options);
    return withTimer(function (done) {
      return timers.setTimeout(done, delay, value);
    }, options);
  }

  async function setImmediatePromise(value, options) {
    validateOptions(options);
    return withTimer(function (done) {
      return timers.setImmediate(done, value);
    }, options);
  }

  async function* setIntervalPromise(delay, value, options) {
    validateNumber(delay, 'delay');
    validateOptions(options);
    var signal = options && options.signal;
    if (signal && signal.aborted) {
      throw abortError(signal);
    }
    var queue = [];
    var waiting = [];
    var done = false;
    var timer;
    function push(item) {
      var pending = waiting;
      waiting = [];
      for (var i = 0; i < pending.length; i++) pending[i](item);
      queue.push(item);
    }
    function stop() {
      done = true;
      if (timer) timer.close();
      if (signal && typeof signal.removeEventListener === 'function') {
        signal.removeEventListener('abort', onAbort);
      }
    }
    function onAbort() {
      stop();
      push({ value: undefined, done: true });
    }
    if (signal && typeof signal.addEventListener === 'function') {
      signal.addEventListener('abort', onAbort);
    }
    timer = timers.setInterval(function () { push({ value: value, done: false }); }, delay);
    if (options && options.ref === false && timer && typeof timer.unref === 'function') {
      timer.unref();
    }
    try {
      while (true) {
        if (queue.length) {
          var item = queue.shift();
          if (item.done) return item.value;
          yield item.value;
          continue;
        }
        if (done) return undefined;
        var next = await new Promise(function (resolve) { waiting.push(resolve); });
        if (next.done) return next.value;
        yield next.value;
      }
    } finally {
      if (!done) stop();
    }
  }

  async function wait(delay, options) {
    return setTimeoutPromise(delay, undefined, options);
  }

  function yieldToLoop() {
    return new Promise(function (resolve) {
      timers.setImmediate(resolve);
    });
  }

  // scheduler is a class Node exposes as an instance: its methods reject a
  // foreign receiver and its constructor cannot be invoked by users.
  function invalidThis() {
    var err = new TypeError(
      'Illegal invocation: Scheduler method called with an incorrect this value');
    err.code = 'ERR_INVALID_THIS';
    return err;
  }
  function Scheduler() {
    var err = new Error('Illegal constructor');
    err.code = 'ERR_ILLEGAL_CONSTRUCTOR';
    throw err;
  }
  Scheduler.prototype.wait = function (delay, options) {
    if (!(this instanceof Scheduler)) throw invalidThis();
    return wait(delay, options);
  };
  Scheduler.prototype.yield = function () {
    if (!(this instanceof Scheduler)) throw invalidThis();
    return yieldToLoop();
  };
  var scheduler = Object.create(Scheduler.prototype);

  return {
    setTimeout: setTimeoutPromise,
    setImmediate: setImmediatePromise,
    setInterval: setIntervalPromise,
    scheduler: scheduler
  };
})`
