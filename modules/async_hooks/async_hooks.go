// Package async_hooks implements Node's async_hooks module: async ids,
// AsyncResource, createHook and AsyncLocalStorage.
//
// The runtime has no engine-level promise hooks (JavaScriptCore does not expose
// V8's setPromiseHooks), so ids and stores are propagated by wrapping the
// scheduling primitives that are observable from JavaScript: timers, immediates,
// process.nextTick and queueMicrotask. Promise chains created with await keep
// their store because the continuation runs inside the same microtask checkpoint
// as the wrapping callback; a promise resolved from an unrelated turn does not
// carry an id, which matches the documented limitation on this module.
package async_hooks

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("async_hooks", Install) }

// Implemented lists the tracked async_hooks API entries.
func Implemented() []string {
	return []string{
		"async_hooks.AsyncLocalStorage",
		"async_hooks.AsyncResource",
		"async_hooks.asyncWrapProviders",
		"async_hooks.createHook",
		"async_hooks.executionAsyncId",
		"async_hooks.executionAsyncResource",
		"async_hooks.triggerAsyncId",
	}
}

// Install builds the async_hooks exports object. Installing it also wraps the
// timer/nextTick globals so ids and AsyncLocalStorage stores survive a
// scheduling boundary.
func Install(e *env.Env) js.Object {
	deps := e.JS.Obj()
	if env.Registered("timers") {
		deps.Set("timers", e.Exports("timers").V())
	}
	if env.Registered("process") {
		deps.Set("process", e.Exports("process").V())
	}
	fn, err := e.JS.Eval(prelude, "[gode:async_hooks]")
	if err != nil {
		panic("godejs async_hooks: " + err.Error())
	}
	v, err := fn.Object().Call(js.Object{}, deps.V())
	if err != nil {
		panic("godejs async_hooks: " + err.Error())
	}
	return v.Object()
}

const prelude = `(function (deps) {
  'use strict';

  var timers = deps.timers || {};
  var processModule = deps.process || {};

  var ROOT_ID = 1;
  var nextId = 2;
  var currentId = ROOT_ID;
  var currentTrigger = 0;
  var rootResource = { type: 'ROOT', asyncId: ROOT_ID };
  var currentResource = rootResource;
  var hooks = [];
  var storages = [];
  var destroyed = Object.create(null);
  var timerRecords = new Map();
  var primitiveRecords = new Map();

  // ------------------------------------------------------------ ids/hooks ----

  function allocate() { return nextId++; }

  function emit(name, a, b, c, d) {
    // A throwing hook must not corrupt the scheduler, so hook failures are
    // reported as uncaught exceptions on the next tick instead of aborting the
    // callback that triggered them.
    for (var i = 0; i < hooks.length; i++) {
      var fn = hooks[i][name];
      if (typeof fn !== 'function') continue;
      try {
        fn.call(hooks[i], a, b, c, d);
      } catch (err) {
        if (typeof processModule.nextTick === 'function') {
          processModule.nextTick(function () { throw err; });
        }
      }
    }
  }

  function init(id, type, trigger, resource) {
    emit('init', id, type, trigger, resource);
  }
  function before(id) { emit('before', id); }
  function after(id) { emit('after', id); }
  function destroy(id) {
    if (destroyed[id]) return;
    destroyed[id] = true;
    emit('destroy', id);
  }

  // Kept as a small helper so hook callbacks never see a stale resource id
  // after a timer has been cleared or has already fired.
  function isDestroyed(id) { return !!destroyed[id]; }

  function createHook(callbacks) {
    if (callbacks === null || typeof callbacks !== 'object') {
      var err = new TypeError('The "callbacks" argument must be an object');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var hook = {
      init: callbacks.init,
      before: callbacks.before,
      after: callbacks.after,
      destroy: callbacks.destroy,
      promiseResolve: callbacks.promiseResolve,
      enable: function () {
        if (hooks.indexOf(hook) === -1) hooks.push(hook);
        return hook;
      },
      disable: function () {
        var at = hooks.indexOf(hook);
        if (at !== -1) hooks.splice(at, 1);
        return hook;
      }
    };
    ['init', 'before', 'after', 'destroy', 'promiseResolve'].forEach(function (name) {
      var fn = hook[name];
      if (fn !== undefined && typeof fn !== 'function') {
        var bad = new TypeError('hook.' + name + ' must be a function');
        bad.code = 'ERR_ASYNC_CALLBACK';
        throw bad;
      }
    });
    return hook;
  }

  function executionAsyncId() { return currentId; }
  function triggerAsyncId() { return currentTrigger; }
  function executionAsyncResource() { return currentResource; }

  // ------------------------------------------------- store propagation ------

  function captureScope(id, trigger) {
    var stores = new Array(storages.length);
    for (var i = 0; i < storages.length; i++) stores[i] = storages[i]._store;
    return { id: id, trigger: trigger, stores: stores };
  }

  function runInScope(scope, fn, thisArg, args) {
    var prevId = currentId;
    var prevTrigger = currentTrigger;
    var prevResource = currentResource;
    var prevStores = new Array(storages.length);
    for (var i = 0; i < storages.length; i++) prevStores[i] = storages[i]._store;

    currentId = scope.id;
    currentTrigger = scope.trigger;
    currentResource = scope.resource || rootResource;
    if (scope.stores) {
      for (var j = 0; j < storages.length; j++) {
        storages[j]._store = j < scope.stores.length ? scope.stores[j] : undefined;
      }
    }
    before(currentId);
    try {
      return fn.apply(thisArg, args);
    } finally {
      after(currentId);
      for (var k = 0; k < storages.length; k++) storages[k]._store = prevStores[k];
      currentId = prevId;
      currentTrigger = prevTrigger;
      currentResource = prevResource;
    }
  }

  function wrapCallback(type, fn, thisArg, args) {
    var id = allocate();
    var trigger = currentId;
    var resource = { type: type, asyncId: id, triggerAsyncId: trigger };
    var scope = captureScope(id, trigger);
    scope.resource = resource;
    init(id, type, trigger, resource);
    var live = true;
    return function () {
      if (!live) return undefined;
      try {
        return runInScope(scope, fn, thisArg, args);
      } finally {
        live = false;
        destroy(id);
      }
    };
  }

  function destroyTimer(record) {
    if (!record || record.destroyed) return;
    record.destroyed = true;
    if (record.handle !== undefined && timerRecords.get(record.handle) === record) {
      timerRecords.delete(record.handle);
    }
    if (record.primitive !== undefined && primitiveRecords.get(record.primitive) === record) {
      primitiveRecords.delete(record.primitive);
    }
    destroy(record.id);
  }

  function runTimer(record, thisArg) {
    if (!record || record.destroyed) return undefined;
    try {
      return runInScope(record.scope, record.fn, thisArg, record.args);
    } finally {
      // A Timeout/Immediate owns its resource until it fires or is cleared.
      // An Interval keeps the same resource (and id) across every tick.
      if (!record.repeat) destroyTimer(record);
    }
  }

  function timerRecordFor(value) {
    if (value === null || value === undefined) return undefined;
    if (typeof value === 'object' || typeof value === 'function') {
      var direct = timerRecords.get(value);
      if (direct) return direct;
      if (value.$godeId !== undefined) return primitiveRecords.get(value.$godeId);
      return undefined;
    }
    if (typeof value === 'number' || typeof value === 'string') {
      return primitiveRecords.get(Number(value));
    }
    return undefined;
  }

  function patchClear(owner, name) {
    if (!owner || typeof owner[name] !== 'function') return;
    var original = owner[name];
    define(owner, name, function (value) {
      var record = timerRecordFor(value);
      if (record) destroyTimer(record);
      return original.apply(this, arguments);
    });
  }

  // ------------------------------------------------------ AsyncResource -----

  function AsyncResource(type, options) {
    if (!(this instanceof AsyncResource)) {
      throw new TypeError("Class constructor AsyncResource cannot be invoked without 'new'");
    }
    if (typeof type !== 'string') {
      var typeErr = new TypeError('The "type" argument must be of type string. Received ' +
        (type === undefined ? 'undefined' : 'type ' + typeof type));
      typeErr.code = 'ERR_INVALID_ARG_TYPE';
      throw typeErr;
    }
    if (type === '') {
      var emptyErr = new TypeError('type must be a string with a non-zero length');
      emptyErr.code = 'ERR_ASYNC_TYPE';
      throw emptyErr;
    }
    var trigger = currentId;
    if (typeof options === 'number') {
      trigger = options;
    } else if (options !== undefined && options !== null) {
      if (typeof options !== 'object') {
        var optErr = new TypeError('The "options" argument must be of type object');
        optErr.code = 'ERR_INVALID_ARG_TYPE';
        throw optErr;
      }
      if (options.triggerAsyncId !== undefined) trigger = options.triggerAsyncId;
    }
    if (!Number.isInteger(trigger) || trigger < 0 || trigger > 9007199254740991) {
      var idErr = new RangeError('Invalid triggerAsyncId value: ' + String(trigger));
      idErr.code = 'ERR_INVALID_ASYNC_ID';
      throw idErr;
    }
    this.type = type;
    this._id = allocate();
    this._trigger = trigger;
    this._manual = !!(options && typeof options === 'object' && options.requireManualDestroy);
    this._destroyed = false;
    this._resource = { type: type, asyncId: this._id, triggerAsyncId: this._trigger };
    this._scope = captureScope(this._id, this._trigger);
    this._scope.resource = this._resource;
    init(this._id, type, this._trigger, this._resource);
  }

  AsyncResource.prototype.runInAsyncScope = function (fn, thisArg) {
    if (typeof fn !== 'function') {
      var err = new TypeError('The "fn" argument must be of type function');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var args = Array.prototype.slice.call(arguments, 2);
    return runInScope(this._scope, fn, thisArg, args);
  };

  AsyncResource.prototype.bind = function (fn, thisArg) {
    var self = this;
    return function () {
      var callArgs = Array.prototype.slice.call(arguments);
      var target = thisArg === undefined ? this : thisArg;
      return self.runInAsyncScope.apply(self, [fn, target].concat(callArgs));
    };
  };

  AsyncResource.prototype.emitDestroy = function () {
    if (!this._destroyed) {
      this._destroyed = true;
      destroy(this._id);
    }
    return this;
  };
  AsyncResource.prototype.asyncId = function () { return this._id; };
  AsyncResource.prototype.triggerAsyncId = function () { return this._trigger; };

  AsyncResource.bind = function (fn, type, thisArg) {
    if (thisArg === undefined && type !== undefined && typeof type !== 'string') {
      thisArg = type;
      type = undefined;
    }
    return new AsyncResource(type || 'bound-anonymous-fn').bind(fn, thisArg);
  };

  // --------------------------------------------------- AsyncLocalStorage ----

  function AsyncLocalStorage() {
    if (!(this instanceof AsyncLocalStorage)) return new AsyncLocalStorage();
    this._store = undefined;
    this._enabled = true;
    storages.push(this);
  }

  AsyncLocalStorage.prototype.getStore = function () {
    return this._enabled ? this._store : undefined;
  };

  AsyncLocalStorage.prototype.run = function (store, callback) {
    if (typeof callback !== 'function') {
      var err = new TypeError('The "callback" argument must be of type function');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var args = Array.prototype.slice.call(arguments, 2);
    var prev = this._store;
    this._store = store;
    try {
      return callback.apply(undefined, args);
    } finally {
      this._store = prev;
    }
  };

  AsyncLocalStorage.prototype.enterWith = function (store) {
    this._store = store;
  };

  AsyncLocalStorage.prototype.exit = function (callback) {
    if (typeof callback !== 'function') {
      var err = new TypeError('The "callback" argument must be of type function');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var args = Array.prototype.slice.call(arguments, 1);
    var prev = this._store;
    this._store = undefined;
    try {
      return callback.apply(undefined, args);
    } finally {
      this._store = prev;
    }
  };

  AsyncLocalStorage.prototype.disable = function () {
    this._enabled = false;
    this._store = undefined;
    var at = storages.indexOf(this);
    if (at !== -1) storages.splice(at, 1);
  };

  AsyncLocalStorage.prototype.bind = function (fn, thisArg) {
    if (typeof fn !== 'function') {
      var err = new TypeError('The "fn" argument must be of type function');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var self = this;
    var store = this.getStore();
    return function () {
      var args = Array.prototype.slice.call(arguments);
      var target = thisArg === undefined ? this : thisArg;
      return self.run.apply(self, [store, fn, target].concat(args));
    };
  };

  AsyncLocalStorage.snapshot = function () {
    var scope = captureScope(currentId, currentTrigger);
    scope.resource = currentResource;
    return function (fn, thisArg) {
      var args = Array.prototype.slice.call(arguments, 2);
      return runInScope(scope, fn, thisArg, args);
    };
  };

  // ---------------------------------------------- scheduling integration ----

  function isFn(fn) { return typeof fn === 'function'; }

  // The runtime installs some globals as read-only host properties, so a plain
  // assignment throws. Redefine through defineProperty and never let a failed
  // patch take the process down: async ids are an enhancement, not a
  // precondition for running JavaScript.
  function define(owner, name, value) {
    try {
      Object.defineProperty(owner, name, {
        value: value, writable: true, configurable: true, enumerable: false
      });
      return true;
    } catch (e) {
      try { owner[name] = value; return true; } catch (e2) { return false; }
    }
  }

  // The timer callback must be installed before scheduling, otherwise the
  // first invocation would run the original function without a hook scope.
  // This implementation schedules with the wrapped function and attaches the
  // record to the returned handle, so clearTimeout/clearImmediate can destroy
  // the same id.
  function patchTimer(owner, name, type) {
    if (!owner || typeof owner[name] !== 'function') return;
    var original = owner[name];
    var immediate = name === 'setImmediate';
    var repeat = name === 'setInterval';
    var wrapped = function (fn) {
      if (!isFn(fn)) return original.apply(this, arguments);
      var args = Array.prototype.slice.call(arguments);
      var extra = Array.prototype.slice.call(arguments, immediate ? 1 : 2);
      var id = allocate();
      var trigger = currentId;
      var scope = captureScope(id, trigger);
      var record = null;
      var callback = function () {
        return runTimer(record, this);
      };
      args[0] = callback;
      if (immediate) {
        args.length = 1;
      } else if (args.length < 2) {
        args[1] = 0;
      }
      var handle = original.apply(this, args);
      if (!handle || (typeof handle !== 'object' && typeof handle !== 'function')) {
        return handle;
      }
      scope.resource = handle;
      record = {
        id: id, type: type, handle: handle, repeat: !!repeat,
        fn: fn, args: extra, scope: scope, destroyed: false,
        primitive: handle.$godeId
      };
      timerRecords.set(handle, record);
      if (record.primitive !== undefined) primitiveRecords.set(record.primitive, record);
      init(id, type, trigger, handle);
      return handle;
    };
    try { wrapped.name = original.name; } catch (e) {}
    define(owner, name, wrapped);
  }

  patchTimer(globalThis, 'setTimeout', 'Timeout');
  patchTimer(globalThis, 'setInterval', 'Interval');
  patchTimer(globalThis, 'setImmediate', 'Immediate');
  if (timers && typeof timers === 'object') {
    patchTimer(timers, 'setTimeout', 'Timeout');
    patchTimer(timers, 'setInterval', 'Interval');
    patchTimer(timers, 'setImmediate', 'Immediate');
  }
  patchClear(globalThis, 'clearTimeout');
  patchClear(globalThis, 'clearInterval');
  patchClear(globalThis, 'clearImmediate');
  if (timers && typeof timers === 'object') {
    patchClear(timers, 'clearTimeout');
    patchClear(timers, 'clearInterval');
    patchClear(timers, 'clearImmediate');
  }

  if (isFn(processModule.nextTick)) {
    var originalNextTick = processModule.nextTick;
    var wrappedNextTick = function (fn) {
      if (!isFn(fn)) return originalNextTick.apply(this, arguments);
      var args = Array.prototype.slice.call(arguments, 1);
      return originalNextTick.call(this, wrapCallback('TickObject', fn, undefined, args));
    };
    define(processModule, 'nextTick', wrappedNextTick);
    if (typeof globalThis.process === 'object' && globalThis.process.nextTick) {
      define(globalThis.process, 'nextTick', wrappedNextTick);
    }
  }

  if (isFn(globalThis.queueMicrotask)) {
    var originalMicrotask = globalThis.queueMicrotask;
    define(globalThis, 'queueMicrotask', function (fn) {
      if (!isFn(fn)) return originalMicrotask.call(this, fn);
      return originalMicrotask.call(this, wrapCallback('Microtask', fn, undefined, []));
    });
  }

  var asyncWrapProviders = {
    NONE: 0, DIRHANDLE: 1, DNSCHANNEL: 2, ELDHISTOGRAM: 3, FILEHANDLE: 4,
    FILEHANDLECLOSEREQ: 5, FIXEDSIZEBLOBCOPY: 6, FSEVENTWRAP: 7, FSREQCALLBACK: 8,
    FSREQPROMISE: 9, GETADDRINFOREQWRAP: 10, GETNAMEINFOREQWRAP: 11, HEAPSNAPSHOT: 12,
    HTTP2SESSION: 13, HTTP2STREAM: 14, HTTP2PING: 15, HTTP2SETTINGS: 16,
    HTTPINCOMINGMESSAGE: 17, HTTPCLIENTREQUEST: 18, JSSTREAM: 19, JSUDPWRAP: 20,
    MESSAGEPORT: 21, PIPECONNECTWRAP: 22, PIPESERVERWRAP: 23, PIPEWRAP: 24,
    PROCESSWRAP: 25, PROMISE: 26, QWRAP: 27, SHUTDOWNWRAP: 28, SIGNALWRAP: 29,
    STATWATCHER: 30, STREAMPIPE: 31, TCPCONNECTWRAP: 32, TCPSERVERWRAP: 33,
    TCPWRAP: 34, TTYWRAP: 35, UDPSENDWRAP: 36, UDPWRAP: 37, SIGINTWATCHDOG: 38,
    WORKER: 39, WORKERHEAPSNAPSHOT: 40, WRITEWRAP: 41, ZLIB: 42, CHECKPRIMEREQUEST: 43,
    PBKDF2REQUEST: 44, KEYPAIRGENREQUEST: 45, KEYGENREQUEST: 46, KEYEXPORTREQUEST: 47,
    CIPHERREQUEST: 48, DERIVEBITSREQUEST: 49, HASHREQUEST: 50, RANDOMBYTESREQUEST: 51,
    RANDOMPRIMEREQUEST: 52, SCRYPTREQUEST: 53, SIGNREQUEST: 54, TLSWRAP: 55,
    VERIFYREQUEST: 56, INSPECTORJSBINDING: 57
  };

  return {
    AsyncLocalStorage: AsyncLocalStorage,
    AsyncResource: AsyncResource,
    asyncWrapProviders: asyncWrapProviders,
    createHook: createHook,
    executionAsyncId: executionAsyncId,
    executionAsyncResource: executionAsyncResource,
    triggerAsyncId: triggerAsyncId
  };
})`
