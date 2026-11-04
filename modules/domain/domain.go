// Package domain implements Node's deprecated domain module on top of the
// events module.
//
// The runtime has no uncaughtException capture callback, so Domain.prototype.run
// routes a synchronous throw to the domain's 'error' handler itself (and
// rethrows when the domain has no such handler). Everything else — the entered
// domain stack, process.domain, add/remove/bind/intercept — follows
// lib/domain.js.
package domain

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("domain", Install) }

// Implemented lists the tracked domain API entries.
func Implemented() []string {
	return []string{
		"domain.Domain",
		"domain.active",
		"domain.create",
		"domain.createDomain",
		"domain.Domain.prototype.add",
		"domain.Domain.prototype.bind",
		"domain.Domain.prototype.enter",
		"domain.Domain.prototype.exit",
		"domain.Domain.prototype.intercept",
		"domain.Domain.prototype.remove",
		"domain.Domain.prototype.run",
	}
}

// Install builds the domain exports object.
func Install(e *env.Env) js.Object {
	deps := e.JS.Obj()
	deps.Set("events", e.Exports("events").V())
	if env.Registered("process") {
		deps.Set("process", e.Exports("process").V())
	}
	fn, err := e.JS.Eval(prelude, "[gode:domain]")
	if err != nil {
		panic("godejs domain: " + err.Error())
	}
	v, err := fn.Object().Call(js.Object{}, deps.V())
	if err != nil {
		panic("godejs domain: " + err.Error())
	}
	return v.Object()
}

const prelude = `(function (deps) {
  'use strict';

  var EventEmitter = deps.events.EventEmitter;
  var processModule = deps.process || null;

  var exports = {};
  var stack = [];
  var current = null;

  function setActive(domain) {
    current = domain;
    if (processModule) {
      try { processModule.domain = domain; } catch (err) { /* frozen process */ }
    }
  }

  function domainUncaughtExceptionClear() {
    stack.length = 0;
    setActive(null);
  }

  class Domain extends EventEmitter {
    constructor() {
      super();
      this.members = [];
    }

    // Called by process._fatalException in Node; here it is also the target of
    // a routed emitter error and of a synchronous throw inside run().
    _errorHandler(er) {
      var caught = false;

      if ((typeof er === 'object' && er !== null) || typeof er === 'function') {
        try {
          Object.defineProperty(er, 'domain', {
            configurable: true, enumerable: false, value: this, writable: true
          });
        } catch (defineError) { /* primitives and frozen errors */ }
        try { er.domainThrown = true; } catch (flagError) { /* frozen error */ }
      }

      // Pop adjacent duplicates so a failing handler cannot re-enter itself.
      while (current === this) this.exit();

      if (this.listenerCount('error') > 0) {
        caught = this.emit('error', er);
      } else if (stack.length === 0) {
        // Without a handler there is nothing to route to: behave like an
        // uncaught exception and let the host report it.
        throw er;
      }

      domainUncaughtExceptionClear();
      return caught;
    }

    enter() {
      setActive(this);
      stack.push(this);
      return this;
    }

    exit() {
      var index = stack.lastIndexOf(this);
      if (index === -1) return this;
      stack.splice(index);
      setActive(stack.length === 0 ? null : stack[stack.length - 1]);
      return this;
    }

    add(ee) {
      if (ee === null || (typeof ee !== 'object' && typeof ee !== 'function')) {
        var err = new TypeError('The "emitter" argument must be an instance of EventEmitter');
        err.code = 'ERR_INVALID_ARG_TYPE';
        throw err;
      }
      if (ee.domain === this) return;
      if (ee.domain) ee.domain.remove(ee);

      // A domain added to itself (indirectly) would recurse forever on error.
      if (this.domain && ee instanceof Domain) {
        for (var d = this.domain; d; d = d.domain) {
          if (ee === d) return;
        }
      }

      Object.defineProperty(ee, 'domain', {
        configurable: true, enumerable: false, value: this, writable: true
      });

      // Route an unhandled 'error' emission of the member into the domain
      // instead of letting it throw. Node does this inside EventEmitter; here
      // the member's emit is shadowed for the lifetime of the membership.
      if (typeof ee.emit === 'function') {
        var originalEmit = ee.emit;
        ee.domainEmitOriginal = originalEmit;
        ee.emit = function (name) {
          if (name === 'error' && this.listenerCount('error') === 0) {
            return this.domain._errorHandler(arguments[1]);
          }
          return originalEmit.apply(this, arguments);
        };
      }

      this.members.push(ee);
    }

    remove(ee) {
      ee.domain = null;
      if (ee.domainEmitOriginal) {
        ee.emit = ee.domainEmitOriginal;
        delete ee.domainEmitOriginal;
      }
      var index = this.members.indexOf(ee);
      if (index !== -1) this.members.splice(index, 1);
    }

    run(fn) {
      var args = Array.prototype.slice.call(arguments, 1);
      this.enter();
      try {
        return fn.apply(this, args);
      } catch (err) {
        return this._errorHandler(err);
      } finally {
        this.exit();
      }
    }

    intercept(cb) {
      var self = this;
      return function () {
        var args = arguments;
        if (args[0] instanceof Error) {
          var er = args[0];
          er.domainBound = cb;
          er.domainThrown = false;
          try {
            Object.defineProperty(er, 'domain', {
              configurable: true, enumerable: false, value: self, writable: true
            });
          } catch (defineError) { /* frozen error */ }
          self.emit('error', er);
          return undefined;
        }
        self.enter();
        try {
          return cb.apply(this, Array.prototype.slice.call(args, 1));
        } finally {
          self.exit();
        }
      };
    }

    bind(cb) {
      var self = this;
      var bound = function () {
        self.enter();
        try {
          return cb.apply(this, arguments);
        } finally {
          self.exit();
        }
      };
      Object.defineProperty(bound, 'domain', {
        configurable: true, enumerable: false, value: this, writable: true
      });
      return bound;
    }

    dispose() {
      this.removeAllListeners();
      this.members.length = 0;
      return this;
    }
  }

  function createDomain() { return new Domain(); }

  exports.Domain = Domain;
  exports.create = createDomain;
  exports.createDomain = createDomain;
  exports._stack = stack;

  Object.defineProperty(exports, 'active', {
    enumerable: true,
    configurable: true,
    get: function () { return current; },
    set: function (value) { setActive(value === undefined ? null : value); }
  });

  // Node exposes process.domain for the same stack.
  setActive(null);

  return exports;
})`
