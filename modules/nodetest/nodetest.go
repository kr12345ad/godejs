// Package nodetest implements the node:test builtin: a small but faithful
// test runner with TAP output, describe/it suites, hooks, subtests, assertion
// planning and the mock API. The scheduling and reporting live in JavaScript;
// the host only supplies timers, stdout and the exit code.
package nodetest

import (
	"fmt"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("test", Install) }

type module struct {
	e     *env.Env
	start time.Time
	fails int
}

// Install builds the node:test exports. The exports object is a function (the
// `test` entry point) carrying every named API as a property, because Node's
// module is callable and destructured in different ways.
func Install(e *env.Env) js.Object {
	m := &module{e: e, start: time.Now()}

	host := e.JS.Obj()
	host.SetFunc("assert", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.Exports("assert").V(), nil
	})
	host.SetFunc("now", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(time.Since(m.start)) / float64(time.Millisecond)), nil
	})
	host.SetFunc("immediate", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			return e.JS.Undefined(), nil
		}
		e.SetTimeout(args[0], 0, false, nil)
		return e.JS.Undefined(), nil
	})
	host.SetFunc("sleep", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			return e.JS.Undefined(), nil
		}
		delay := 0.0
		if len(args) > 1 {
			delay = args[1].Float()
		}
		if delay < 0 || delay != delay {
			delay = 0
		}
		e.SetTimeout(args[0], delay, false, nil)
		return e.JS.Undefined(), nil
	})
	host.SetFunc("write", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			fmt.Fprint(e.Out(), args[0].Str())
		}
		return e.JS.Undefined(), nil
	})
	host.SetFunc("fail", func(this js.Object, args []js.Value) (js.Value, error) {
		m.fails++
		e.ExitCode = 1
		return e.JS.Undefined(), nil
	})
	host.SetFunc("filePath", func(this js.Object, args []js.Value) (js.Value, error) {
		argv := e.Argv()
		if len(argv) > 1 {
			return e.JS.Str(argv[1]), nil
		}
		return e.JS.Str(""), nil
	})

	factory, err := e.JS.Eval(testPrelude, "[gode:test]")
	if err != nil {
		panic("godejs node:test: " + err.Error())
	}
	built, err := factory.Object().Call(js.Object{}, host.V())
	if err != nil {
		panic("godejs node:test: " + err.Error())
	}
	return built.Object()
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"test",
		"test.test",
		"test.describe",
		"test.it",
		"test.suite",
		"test.skip",
		"test.todo",
		"test.only",
		"test.before",
		"test.after",
		"test.beforeEach",
		"test.afterEach",
		"test.mock",
		"test.mock.fn",
		"test.mock.method",
		"test.mock.getter",
		"test.mock.setter",
		"test.mock.property",
		"test.mock.reset",
		"test.mock.restoreAll",
		"test.mock.timers",
	}
}

// testPrelude is the whole runner. It takes the host object and returns the
// callable `test` function with every API attached.
const testPrelude = `(function (host) {
  'use strict';

  var assertModule = host.assert();

  // ------------------------------------------------------------- reporting --
  var tapIndex = 0;
  var started = false;
  var stats = { tests: 0, pass: 0, fail: 0, skip: 0, todo: 0 };

  function writeLine(s) { host.write(s + '\n'); }

  function tapName(node) {
    var parts = [];
    for (var n = node; n; n = n.parent) {
      if (n.name) parts.unshift(n.name);
    }
    return parts.join(' > ').replace(/#/g, '\\#').replace(/\r?\n/g, ' ');
  }

  function describeError(err) {
    if (err && typeof err === 'object') {
      var name = err.name || 'Error';
      var message = err.message === undefined || err.message === null ? '' : String(err.message);
      var head = name + (err.code ? ' [' + err.code + ']' : '') + (message ? ': ' + message : '');
      var stack = typeof err.stack === 'string' ? err.stack : '';
      if (!stack) return head;
      // JavaScriptCore stacks are frame lists; Node's include the message.
      if (stack.indexOf(head) === 0 || stack.indexOf(name + ':') === 0) return stack;
      return head + '\n' + stack;
    }
    return 'Error: ' + String(err);
  }

  function ensureTAP() {
    if (started) return;
    started = true;
    writeLine('TAP version 13');
  }

  function reportOk(node) {
    ensureTAP();
    writeLine('ok ' + (++tapIndex) + ' - ' + tapName(node));
  }

  function reportSkip(node, reason) {
    ensureTAP();
    var suffix = ' # SKIP' + (reason ? ' ' + String(reason).replace(/\r?\n/g, ' ') : '');
    writeLine('ok ' + (++tapIndex) + ' - ' + tapName(node) + suffix);
  }

  function reportTodo(node, reason) {
    ensureTAP();
    var suffix = ' # TODO' + (reason ? ' ' + String(reason).replace(/\r?\n/g, ' ') : '');
    writeLine('ok ' + (++tapIndex) + ' - ' + tapName(node) + suffix);
  }

  function reportFail(node, err) {
    ensureTAP();
    writeLine('not ok ' + (++tapIndex) + ' - ' + tapName(node));
    writeLine('  ---');
    var lines = describeError(err).split('\n');
    writeLine('  error: |-');
    for (var i = 0; i < lines.length; i++) {
      writeLine('    ' + lines[i]);
    }
    var code = err && err.code ? String(err.code) : 'ERR_TEST_FAILURE';
    writeLine('  code: ' + code);
    writeLine('  ...');
    host.fail();
  }

  // ---------------------------------------------------------------- tree ----
  function makeSuite(name, options, parent) {
    return {
      kind: 'suite',
      name: name,
      options: options || {},
      parent: parent || null,
      children: [],
      before: [],
      after: [],
      beforeEach: [],
      afterEach: []
    };
  }

  var root = makeSuite('', {}, null);
  var currentSuite = root;
  var scheduled = false;
  var running = false;
  var finished = false;

  function schedule() {
    if (scheduled || finished) return;
    scheduled = true;
    host.immediate(pump);
  }

  function pump() {
    scheduled = false;
    if (running || finished) return;
    if (root.children.length === 0) {
      finished = true;
      finishRun();
      return;
    }
    running = true;
    runSuite(root, function () {
      running = false;
      finished = true;
      finishRun();
    });
  }

  function finishRun() {
    ensureTAP();
    writeLine('1..' + tapIndex);
    writeLine('# tests ' + stats.tests);
    writeLine('# pass ' + stats.pass);
    if (stats.fail > 0) writeLine('# fail ' + stats.fail);
    if (stats.skip > 0) writeLine('# skipped ' + stats.skip);
    if (stats.todo > 0) writeLine('# todo ' + stats.todo);
  }

  function runHooks(list, node, cb) {
    var i = 0;
    function next() {
      if (i >= list.length) return cb();
      var hook = list[i++];
      var result;
      try {
        result = hook.call(null, makeContext(node, null, null));
      } catch (err) {
        stats.fail++;
        reportFail({ name: hook.name || 'hook', parent: node }, err);
        return next();
      }
      if (result && typeof result.then === 'function') {
        result.then(function () { next(); }, function (err) {
          stats.fail++;
          reportFail({ name: hook.name || 'hook', parent: node }, err);
          next();
        });
        return;
      }
      next();
    }
    next();
  }

  function runSuite(node, cb) {
    runHooks(node.before, node, function () {
      var index = 0;
      function step() {
        if (index >= node.children.length) {
          return runHooks(node.after, node, cb);
        }
        var child = node.children[index++];
        if (child.kind === 'suite') {
          runSuite(child, step);
          return;
        }
        runHooks(node.beforeEach, node, function () {
          runTest(child, function () {
            runHooks(node.afterEach, node, function () { step(); });
          });
        });
      }
      step();
    });
  }

  // ------------------------------------------------------------ execution ---
  function runTest(node, cb) {
    var options = node.options || {};
    stats.tests++;

    if (options.skip) {
      stats.skip++;
      reportSkip(node, typeof options.skip === 'string' ? options.skip : '');
      return cb();
    }
    if (options.todo) {
      stats.todo++;
      reportTodo(node, typeof options.todo === 'string' ? options.todo : '');
      return cb();
    }
    if (typeof node.fn !== 'function') {
      stats.pass++;
      reportOk(node);
      return cb();
    }

    var done = false;
    var planned = 0;
    var count = 0;
    var planDeadline = 0;

    var ctx = makeContext(node, function () {
      count++;
      if (planned && count >= planned) settle();
    }, function (n) { planned = n; });

    function finishOk() {
      if (done) return;
      done = true;
      stats.pass++;
      reportOk(node);
      cb();
    }

    function finishErr(err) {
      if (done) return;
      done = true;
      stats.fail++;
      reportFail(node, err);
      cb();
    }

    // A test that declared t.plan(n) stays open until the assertions arrive,
    // the way Node keeps an awaiting test alive. The wait is bounded so a
    // genuinely unsatisfied plan fails instead of hanging the run.
    function settle() {
      if (done) return;
      if (planned && count < planned) {
        if (planDeadline === 0) planDeadline = host.now() + 1000;
        if (host.now() > planDeadline) {
          finishErr(new Error('test planned ' + planned + ' assertion(s) but received ' + count));
          return;
        }
        host.sleep(settle, 1);
        return;
      }
      finishOk();
    }

    // node:test also accepts the callback form, test(name, (t, done) => ...).
    // Such a test stays open until done() runs, so its arity decides whether the
    // synchronous return settles it.
    var awaitDone = node.fn.length >= 2;
    var doneArg = function (err) {
      if (done) return;
      if (err) { finishErr(err); return; }
      settle();
    };

    var result;
    try {
      result = awaitDone ? node.fn.call(ctx, ctx, doneArg) : node.fn.call(ctx, ctx);
    } catch (err) {
      finishErr(err);
      return;
    }
    if (result && typeof result.then === 'function') {
      result.then(settle, finishErr);
      return;
    }
    if (awaitDone) return;
    settle();
  }

  var testSymbol = typeof Symbol === 'function' ? Symbol('test') : 'test';

  function makeContext(node, onAssert, onPlan) {
    var signal = null;
    if (typeof AbortController === 'function') {
      signal = new AbortController().signal;
    } else {
      signal = { aborted: false, addEventListener: function () {}, removeEventListener: function () {} };
    }

    function wrapAssert() {
      var wrapped = {};
      var names = Object.getOwnPropertyNames(assertModule);
      for (var i = 0; i < names.length; i++) {
        var name = names[i];
        if (name === 'length' || name === 'name' || name === 'prototype' ||
            name === 'arguments' || name === 'caller') {
          continue;
        }
        var value = assertModule[name];
        if (typeof value !== 'function') {
          wrapped[name] = value;
          continue;
        }
        wrapped[name] = (function (fn) {
          return function () {
            if (onAssert) onAssert();
            return fn.apply(this, arguments);
          };
        })(value);
      }
      return wrapped;
    }

    var ctx = {
      name: node.name,
      fullName: tapName(node),
      filePath: host.filePath(),
      signal: signal,
      assert: wrapAssert(),
      mock: null,
      diagnostic: function (message) { writeLine('# ' + String(message)); },
      comment: function (message) { writeLine('# ' + String(message)); },
      skip: function (message) {
        var err = new Error(message === undefined ? 'test skipped' : String(message));
        err.code = 'ERR_TEST_SKIP';
        throw err;
      },
      todo: function (message) {
        var err = new Error(message === undefined ? 'test todo' : String(message));
        err.code = 'ERR_TEST_TODO';
        throw err;
      },
      test: function (name, opts, fn) {
        if (typeof opts === 'function') {
          fn = opts;
          opts = {};
        }
        var child = {
          kind: 'test',
          name: name,
          options: opts || {},
          fn: fn,
          parent: node
        };
        if (onAssert) onAssert();
        return new Promise(function (resolve) {
          runTest(child, function () { resolve(); });
        });
      },
      after: function (fn) { if (typeof fn === 'function') node.after = (node.after || []).concat([fn]); },
      before: function (fn) { if (typeof fn === 'function') node.before = (node.before || []).concat([fn]); },
      beforeEach: function (fn) { if (typeof fn === 'function') node.beforeEach = (node.beforeEach || []).concat([fn]); },
      afterEach: function (fn) { if (typeof fn === 'function') node.afterEach = (node.afterEach || []).concat([fn]); },
      error: function (err) { throw err; },
      waitFor: function (condition, opts) {
        opts = opts || {};
        var interval = opts.interval === undefined ? 50 : opts.interval;
        var timeout = opts.timeout === undefined ? 1000 : opts.timeout;
        return new Promise(function (resolve, reject) {
          var start = host.now();
          function poll() {
            var value;
            try {
              value = condition();
            } catch (err) {
              reject(err);
              return;
            }
            if (value) {
              resolve(value);
              return;
            }
            if (host.now() - start >= timeout) {
              reject(new Error('waitFor timed out after ' + timeout + 'ms'));
              return;
            }
            host.sleep(poll, interval);
          }
          poll();
        });
      }
    };
    // t.plan() feeds the runner's own counter through onPlan.
    ctx.plan = function (n) {
      if (onPlan) onPlan(n);
    };
    ctx.mock = mock;
    return ctx;
  }

  // ---------------------------------------------------------------- mock ----
  var mockRegistry = [];
  var activeTimers = null;

  function makeMockFunction(impl, originalName) {
    var calls = [];
    var results = [];
    function mocked() {
      var args = Array.prototype.slice.call(arguments);
      var record = { arguments: args, this: this };
      calls.push(record);
      var value;
      try {
        value = (impl || function () {}).apply(this, args);
      } catch (err) {
        results.push({ type: 'throw', value: err });
        throw err;
      }
      results.push({ type: 'return', value: value });
      return value;
    }
    if (originalName) {
      try {
        Object.defineProperty(mocked, 'name', { value: originalName, configurable: true });
      } catch (e) { /* name is only cosmetic */ }
    }
    mocked.mock = {
      calls: calls,
      results: results,
      callCount: function () { return calls.length; },
      resetCalls: function () {
        calls.length = 0;
        results.length = 0;
      }
    };
    mockRegistry.push(mocked);
    return mocked;
  }

  function mockMethod(object, methodName, impl) {
    var original = object[methodName];
    var mocked = makeMockFunction(impl === undefined ? original : impl, methodName);
    mocked.mock.object = object;
    mocked.mock.key = methodName;
    mocked.mock.original = original;
    mocked.mock.restore = function () { object[methodName] = original; };
    object[methodName] = mocked;
    return mocked;
  }

  function mockAccessor(object, key, getter, setter) {
    var descriptor = Object.getOwnPropertyDescriptor(object, key);
    var mocked = getter ? makeMockFunction(getter, key) : null;
    var mockedSetter = setter ? makeMockFunction(setter, key) : null;
    var replacement = {
      configurable: true,
      enumerable: descriptor ? descriptor.enumerable : true
    };
    if (mocked) replacement.get = mocked;
    else if (descriptor && descriptor.get) replacement.get = descriptor.get;
    if (mockedSetter) replacement.set = mockedSetter;
    else if (descriptor && descriptor.set) replacement.set = descriptor.set;
    Object.defineProperty(object, key, replacement);
    var restore = function () {
      if (descriptor) Object.defineProperty(object, key, descriptor);
      else delete object[key];
    };
    if (mocked) mocked.mock.restore = restore;
    if (mockedSetter) mockedSetter.mock.restore = restore;
    return mocked || mockedSetter;
  }

  var mock = {
    fn: function (impl) { return makeMockFunction(impl); },
    method: function (object, methodName, impl) { return mockMethod(object, methodName, impl); },
    getter: function (object, key, getter) { return mockAccessor(object, key, getter, null); },
    setter: function (object, key, setter) { return mockAccessor(object, key, null, setter); },
    property: function (object, key, value) {
      var descriptor = Object.getOwnPropertyDescriptor(object, key);
      Object.defineProperty(object, key, {
        value: value,
        configurable: true,
        writable: true,
        enumerable: descriptor ? descriptor.enumerable : true
      });
      var mocked = makeMockFunction(function () { return value; }, key);
      mocked.mock.restore = function () {
        if (descriptor) Object.defineProperty(object, key, descriptor);
        else delete object[key];
      };
      return mocked;
    },
    reset: function () {
      for (var i = 0; i < mockRegistry.length; i++) mockRegistry[i].mock.resetCalls();
    },
    restoreAll: function () {
      for (var i = 0; i < mockRegistry.length; i++) {
        var restore = mockRegistry[i].mock.restore;
        if (typeof restore === 'function') restore();
      }
      mockRegistry = [];
    },
    timers: {
      enable: function (options) { enableFakeTimers(options); },
      enableFor: function (target, options) { enableFakeTimers(options); },
      reset: function () { if (activeTimers) activeTimers.reset(); },
      tick: function (ms) { if (activeTimers) activeTimers.tick(ms); },
      runAll: function () { if (activeTimers) activeTimers.tick(Infinity); },
      setTime: function (ms) { if (activeTimers) activeTimers.now = ms; },
      get: function () { return activeTimers; }
    }
  };

  // --------------------------------------------------------- fake timers ----
  function enableFakeTimers(options) {
    if (activeTimers) return;
    var apis = (options && options.apis) || ['setTimeout', 'setInterval', 'setImmediate', 'Date'];
    var real = {
      setTimeout: globalThis.setTimeout,
      clearTimeout: globalThis.clearTimeout,
      setInterval: globalThis.setInterval,
      clearInterval: globalThis.clearInterval,
      setImmediate: globalThis.setImmediate,
      clearImmediate: globalThis.clearImmediate,
      Date: globalThis.Date
    };
    var clock = {
      now: (options && typeof options.now === 'number') ? options.now : 0,
      seq: 0,
      entries: [],
      real: real,
      reset: function () { clock.entries = []; },
      tick: function (ms) {
        var target = ms === Infinity ? Infinity : clock.now + ms;
        for (;;) {
          var due = null;
          for (var i = 0; i < clock.entries.length; i++) {
            var e = clock.entries[i];
            if (e.due <= target && (!due || e.due < due.due || (e.due === due.due && e.seq < due.seq))) {
              due = e;
            }
          }
          if (!due) break;
          clock.now = due.due;
          if (due.repeat) {
            due.due = clock.now + due.delay;
            due.seq = ++clock.seq;
          } else {
            clock.entries.splice(clock.entries.indexOf(due), 1);
          }
          if (typeof due.fn === 'function') {
            try {
              due.fn.apply(null, due.args);
            } catch (err) {
              throw err;
            }
          }
          if (target === Infinity && clock.entries.length === 0) break;
        }
        clock.now = target === Infinity ? clock.now : target;
      }
    };
    var handles = new Map();
    var nextId = 1;
    function add(kind, fn, delay, args, repeat) {
      var id = nextId++;
      var entry = {
        id: id, kind: kind, fn: fn, delay: delay, args: args, repeat: repeat,
        due: clock.now + (kind === 'immediate' ? 0 : delay), seq: ++clock.seq
      };
      clock.entries.push(entry);
      handles.set(id, entry);
      return id;
    }
    function clear(id) {
      var entry = handles.get(id);
      if (!entry) return;
      handles.delete(id);
      var index = clock.entries.indexOf(entry);
      if (index >= 0) clock.entries.splice(index, 1);
    }
    function restArgs(args) {
      return Array.prototype.slice.call(args, 2);
    }
    globalThis.setTimeout = function (fn, delay) {
      return add('timeout', fn, Number(delay) || 0, restArgs(arguments), false);
    };
    globalThis.setInterval = function (fn, delay) {
      return add('interval', fn, Number(delay) || 0, restArgs(arguments), true);
    };
    globalThis.setImmediate = function (fn) {
      return add('immediate', fn, 0, Array.prototype.slice.call(arguments, 1), false);
    };
    globalThis.clearTimeout = function (id) { clear(id); };
    globalThis.clearInterval = function (id) { clear(id); };
    globalThis.clearImmediate = function (id) { clear(id); };
    if (apis.indexOf('Date') !== -1) {
      var RealDate = real.Date;
      function FakeDate() {
        if (!(this instanceof FakeDate)) return new RealDate(clock.now).toString();
        if (arguments.length === 0) return new RealDate(clock.now);
        return new (Function.prototype.bind.apply(RealDate, [null].concat(Array.prototype.slice.call(arguments))))();
      }
      FakeDate.prototype = RealDate.prototype;
      FakeDate.now = function () { return clock.now; };
      FakeDate.parse = RealDate.parse;
      FakeDate.UTC = RealDate.UTC;
      globalThis.Date = FakeDate;
    }
    clock.disable = function () {
      globalThis.setTimeout = real.setTimeout;
      globalThis.setInterval = real.setInterval;
      globalThis.setImmediate = real.setImmediate;
      globalThis.clearTimeout = real.clearTimeout;
      globalThis.clearInterval = real.clearInterval;
      globalThis.clearImmediate = real.clearImmediate;
      globalThis.Date = real.Date;
      activeTimers = null;
    };
    activeTimers = clock;
  }

  // ---------------------------------------------------------- registration --
  function normalise(options) {
    if (typeof options === 'function') return { options: {}, fn: options };
    if (options === undefined || options === null) return { options: {}, fn: undefined };
    return { options: options, fn: undefined };
  }

  function addTest(name, options, fn) {
    var normalised = normalise(options);
    var opts = normalised.options;
    var callback = fn === undefined ? normalised.fn : fn;
    var node = {
      kind: 'test',
      name: String(name),
      options: opts || {},
      fn: callback,
      parent: currentSuite
    };
    currentSuite.children.push(node);
    schedule();
    return node;
  }

  function test(name, options, fn) {
    addTest(name, options, fn);
    return Promise.resolve();
  }

  test.skip = function (name, options, fn) {
    if (typeof options === 'function' || options === undefined) {
      addTest(name, { skip: true }, typeof options === 'function' ? options : undefined);
      return Promise.resolve();
    }
    var opts = {};
    for (var key in options) opts[key] = options[key];
    opts.skip = opts.skip === undefined ? true : opts.skip;
    addTest(name, opts, fn);
    return Promise.resolve();
  };

  test.todo = function (name, options, fn) {
    var opts = {};
    if (options && typeof options === 'object') {
      for (var key in options) opts[key] = options[key];
    }
    if (opts.todo === undefined) opts.todo = true;
    addTest(name, opts, typeof options === 'function' ? options : fn);
    return Promise.resolve();
  };

  test.only = function (name, options, fn) {
    addTest(name, options, fn);
    return Promise.resolve();
  };

  function describe(name, options, fn) {
    if (typeof options === 'function') {
      fn = options;
      options = {};
    }
    var suite = makeSuite(String(name), options || {}, currentSuite);
    currentSuite.children.push(suite);
    var previous = currentSuite;
    currentSuite = suite;
    try {
      if (typeof fn === 'function') fn.call(suite);
    } catch (err) {
      suite.children.push({
        kind: 'test',
        name: 'describe block',
        options: {},
        fn: function () { throw err; },
        parent: suite
      });
    } finally {
      currentSuite = previous;
    }
    schedule();
    return suite;
  }

  // Hooks attach to the suite being described, so a before hook inside a
  // describe block belongs to that block and a top-level one to the file.
  function addHook(kind) {
    return function (nameOrFn, maybeFn) {
      var fn = typeof nameOrFn === 'function' ? nameOrFn : maybeFn;
      if (typeof fn === 'function') currentSuite[kind].push(fn);
    };
  }

  test.test = test;
  test.it = test;
  test.describe = describe;
  test.suite = describe;
  test.before = addHook('before');
  test.after = addHook('after');
  test.beforeEach = addHook('beforeEach');
  test.afterEach = addHook('afterEach');
  test.mock = mock;
  test.assert = assertModule;

  Object.defineProperty(test, testSymbol, { value: true, enumerable: false });

  return test;
})`
