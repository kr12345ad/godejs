package compat

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

// evalFactory evaluates a factory prelude and calls it with args. Keeping the
// classes and state machines in JavaScript makes the exported values behave
// like realm-native Node objects instead of generic host wrappers.
func evalFactory(e *env.Env, source, filename string, args ...js.Value) js.Object {
	factory, err := e.JS.Eval(source, filename)
	if err != nil {
		panic("godejs " + filename + ": " + err.Error())
	}
	built, err := factory.Object().Call(js.Object{}, args...)
	if err != nil {
		panic("godejs " + filename + ": " + err.Error())
	}
	return built.Object()
}

func installPunycode(e *env.Env) js.Object {
	return evalFactory(e, punycodePrelude, "punycode.js")
}

func installPerfHooks(e *env.Env) js.Object {
	performance := e.Global().Get("performance")
	if !performance.Valid() || !performance.IsObject() {
		performance = js.Value{}
	}
	return evalFactory(e, perfHooksPrelude, "perf_hooks.js", performance)
}

func installTraceEvents(e *env.Env) js.Object {
	return evalFactory(e, traceEventsPrelude, "trace_events.js")
}

func installInspector(e *env.Env) js.Object {
	return evalFactory(e, inspectorPrelude, "inspector.js")
}

func installRepl(e *env.Env) js.Object {
	events := e.Exports("events")
	readline := e.Exports("readline")
	vm := e.Exports("vm")
	util := e.Exports("util")

	emitter := events.Get("EventEmitter")
	iface := readline.Get("Interface")
	if !iface.Valid() || !iface.IsFunction() {
		iface = js.Value{}
	}
	var proc js.Value
	if p := e.Global().Get("process"); p.Valid() {
		proc = p
	}
	return evalFactory(e, replPrelude, "repl.js",
		emitter, iface, vm.V(), util.V(), proc, e.Global().V())
}

func installSea(e *env.Env) js.Object {
	return evalFactory(e, seaPrelude, "sea.js")
}

// punycodePrelude is the RFC 3492 codec used by Node's deprecated punycode
// builtin. It deliberately does not use URL.domainToASCII because punycode
// also accepts standalone labels and email-like strings.
const punycodePrelude = `(function () {
  'use strict';
  var maxInt = 2147483647;
  var base = 36;
  var tMin = 1;
  var tMax = 26;
  var skew = 38;
  var damp = 700;
  var initialBias = 72;
  var initialN = 128;
  var delimiter = '-';

  function error(message) { throw new RangeError(message); }

  function ucs2decode(string) {
    var output = [];
    var counter = 0;
    var length = string.length;
    while (counter < length) {
      var value = string.charCodeAt(counter++);
      if (value >= 0xD800 && value <= 0xDBFF && counter < length) {
        var extra = string.charCodeAt(counter++);
        if ((extra & 0xFC00) === 0xDC00) {
          output.push(((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000);
        } else {
          output.push(value);
          counter--;
        }
      } else {
        output.push(value);
      }
    }
    return output;
  }

  function ucs2encode(array) {
    var output = '';
    for (var i = 0; i < array.length; i++) {
      var value = array[i];
      if (value > 0xFFFF) {
        value -= 0x10000;
        output += String.fromCharCode((value >>> 10) + 0xD800);
        output += String.fromCharCode((value & 0x3FF) + 0xDC00);
      } else {
        output += String.fromCharCode(value);
      }
    }
    return output;
  }

  function adapt(delta, numPoints, firstTime) {
    var k = 0;
    delta = firstTime ? Math.floor(delta / damp) : delta >> 1;
    delta += Math.floor(delta / numPoints);
    for (; delta > (base - tMin) * tMax >> 1; k += base) {
      delta = Math.floor(delta / (base - tMin));
    }
    return Math.floor(k + (base - tMin + 1) * delta / (delta + skew));
  }

  function basicToDigit(codePoint) {
    if (codePoint - 48 < 10) return codePoint - 22;
    if (codePoint - 65 < 26) return codePoint - 65;
    if (codePoint - 97 < 26) return codePoint - 97;
    return base;
  }

  function digitToBasic(digit) {
    return digit + 22 + 75 * (digit < 26);
  }

  function decode(input) {
    var output = [];
    var inputLength = input.length;
    var i = 0;
    var n = initialN;
    var bias = initialBias;
    var basic = input.lastIndexOf(delimiter);
    if (basic < 0) basic = 0;
    for (var j = 0; j < basic; j++) {
      if (input.charCodeAt(j) >= 0x80) error('Illegal input');
      output.push(input.charCodeAt(j));
    }
    for (var index = basic > 0 ? basic + 1 : 0; index < inputLength;) {
      var oldi = i;
      var w = 1;
      for (var k = base;; k += base) {
        if (index >= inputLength) error('Invalid input');
        var digit = basicToDigit(input.charCodeAt(index++));
        if (digit >= base) error('Invalid input');
        if (digit > Math.floor((maxInt - i) / w)) error('Overflow');
        i += digit * w;
        var t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias);
        if (digit < t) break;
        if (w > Math.floor(maxInt / (base - t))) error('Overflow');
        w *= base - t;
      }
      var out = output.length + 1;
      bias = adapt(i - oldi, out, oldi === 0);
      if (Math.floor(i / out) > maxInt - n) error('Overflow');
      n += Math.floor(i / out);
      i %= out;
      output.splice(i++, 0, n);
    }
    return ucs2encode(output);
  }

  function encode(input) {
    var codePoints = ucs2decode(String(input));
    var output = [];
    var n = initialN;
    var delta = 0;
    var bias = initialBias;
    var basicLength = 0;
    for (var i = 0; i < codePoints.length; i++) {
      var c = codePoints[i];
      if (c < 0x80) {
        output.push(String.fromCharCode(c));
        basicLength++;
      }
    }
    var handled = basicLength;
    if (basicLength) output.push(delimiter);
    while (handled < codePoints.length) {
      var m = maxInt;
      for (i = 0; i < codePoints.length; i++) {
        if (codePoints[i] >= n && codePoints[i] < m) m = codePoints[i];
      }
      if (m - n > Math.floor((maxInt - delta) / (handled + 1))) error('Overflow');
      delta += (m - n) * (handled + 1);
      n = m;
      for (i = 0; i < codePoints.length; i++) {
        if (codePoints[i] < n && ++delta > maxInt) error('Overflow');
        if (codePoints[i] === n) {
          var q = delta;
          for (var k = base;; k += base) {
            var t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias);
            if (q < t) break;
            output.push(String.fromCharCode(digitToBasic(t + (q - t) % (base - t))));
            q = Math.floor((q - t) / (base - t));
          }
          output.push(String.fromCharCode(digitToBasic(q)));
          bias = adapt(delta, handled + 1, handled === basicLength);
          delta = 0;
          handled++;
        }
      }
      delta++;
      n++;
    }
    return output.join('');
  }

  function mapLabels(input, fn) {
    return String(input).split('@').map(function (part) {
      return part.split('.').map(fn).join('.');
    }).join('@');
  }

  function toASCII(input) {
    return mapLabels(input, function (label) {
      return /[^\x00-\x7F]/.test(label) ? 'xn--' + encode(label) : label;
    });
  }

  function toUnicode(input) {
    return mapLabels(input, function (label) {
      return /^xn--/i.test(label) ? decode(label.slice(4)) : label;
    });
  }

  var api = {
    decode: decode,
    encode: encode,
    toASCII: toASCII,
    toUnicode: toUnicode,
    ucs2: { decode: ucs2decode, encode: ucs2encode },
    version: '2.1.0'
  };
  return api;
})`

// perfHooksPrelude adds the Node performance observer/histogram surface to the
// web performance object already installed by the runtime. Marks and measures
// are bridged into observers without replacing the web implementation.
const perfHooksPrelude = `(function (existing) {
  'use strict';
  var now = (existing && typeof existing.now === 'function')
    ? function () { return existing.now(); }
    : function () { return Date.now() - timeOrigin; };
  var timeOrigin = (existing && typeof existing.timeOrigin === 'number')
    ? existing.timeOrigin : Date.now();
  var performance = existing && typeof existing === 'object' ? existing : {};
  if (typeof performance.now !== 'function') performance.now = now;
  if (typeof performance.timeOrigin !== 'number') performance.timeOrigin = timeOrigin;

  function schedule(fn) {
    if (typeof queueMicrotask === 'function') queueMicrotask(fn);
    else Promise.resolve().then(fn);
  }

  function percentile(values, p) {
    if (!values.length) return 0;
    var sorted = values.slice().sort(function (a, b) { return a - b; });
    var index = (sorted.length - 1) * Math.min(100, Math.max(0, Number(p))) / 100;
    var lower = Math.floor(index);
    var upper = Math.ceil(index);
    if (lower === upper) return sorted[lower];
    return sorted[lower] + (sorted[upper] - sorted[lower]) * (index - lower);
  }

  class Histogram {
    constructor(options) {
      options = options || {};
      this.lowest = options.lowest === undefined ? 1 : options.lowest;
      this.highest = options.highest === undefined ? Number.MAX_SAFE_INTEGER : options.highest;
      this.figures = options.figures === undefined ? 3 : options.figures;
      this._values = [];
    }
    record(value) {
      value = Number(value);
      if (!Number.isFinite(value)) throw new TypeError('value must be a finite number');
      this._values.push(value);
    }
    reset() { this._values.length = 0; }
    get count() { return this._values.length; }
    get min() { return this._values.length ? Math.min.apply(null, this._values) : 0; }
    get max() { return this._values.length ? Math.max.apply(null, this._values) : 0; }
    get mean() { return this._values.length ? this._values.reduce(function (a, b) { return a + b; }, 0) / this._values.length : 0; }
    get stddev() {
      if (this._values.length < 2) return 0;
      var mean = this.mean;
      var spans = this._values.map(function (v) { return Math.pow(v - mean, 2); });
      return Math.sqrt(spans.reduce(function (a, b) { return a + b; }, 0) / this._values.length);
    }
    percentile(p) { return percentile(this._values, p); }
    get exceeds() { return this.count; }
    toJSON() {
      return { lowest: this.lowest, highest: this.highest, figures: this.figures,
        count: this.count, min: this.min, max: this.max, mean: this.mean,
        stddev: this.stddev, percentiles: {} };
    }
  }

  class PerformanceEntry {
    constructor(name, entryType, startTime, duration) {
      this.name = String(name);
      this.entryType = String(entryType);
      this.startTime = Number(startTime) || 0;
      this.duration = Number(duration) || 0;
    }
    toJSON() {
      return { name: this.name, entryType: this.entryType,
        startTime: this.startTime, duration: this.duration };
    }
  }

  var observers = [];
  function notify(entry) {
    if (!entry) return;
    for (var i = 0; i < observers.length; i++) {
      if (observers[i]._matches(entry)) observers[i]._enqueue(entry);
    }
  }

  var originalMark = typeof performance.mark === 'function' ? performance.mark : null;
  var originalMeasure = typeof performance.measure === 'function' ? performance.measure : null;
  if (originalMark) {
    performance.mark = function () {
      var entry = originalMark.apply(this, arguments);
      notify(entry);
      return entry;
    };
  }
  if (originalMeasure) {
    performance.measure = function () {
      var entry = originalMeasure.apply(this, arguments);
      notify(entry);
      return entry;
    };
  }

  class PerformanceObserverEntryList {
    constructor(entries) { this._entries = entries || []; }
    getEntries() { return this._entries.slice(); }
    getEntriesByType(type) { return this._entries.filter(function (e) { return e.entryType === type; }); }
    getEntriesByName(name, type) {
      return this._entries.filter(function (e) {
        return e.name === name && (type === undefined || e.entryType === type);
      });
    }
  }

  class PerformanceObserver {
    constructor(callback) {
      if (typeof callback !== 'function') throw new TypeError('callback must be a function');
      this._callback = callback;
      this._entryTypes = [];
      this._buffer = [];
      this._connected = false;
      observers.push(this);
    }
    static get supportedEntryTypes() {
      return ['draft', 'gc', 'http', 'http2', 'mark', 'measure', 'net', 'resource', 'function'];
    }
    _matches(entry) {
      return this._entryTypes.indexOf(entry.entryType) !== -1;
    }
    _enqueue(entry) {
      this._buffer.push(entry);
      var self = this;
      schedule(function () { self._flush(); });
    }
    _flush() {
      if (!this._buffer.length || !this._connected) return;
      var entries = this._buffer.splice(0);
      this._callback(new PerformanceObserverEntryList(entries), this);
    }
    observe(options) {
      options = options || {};
      if (options.entryTypes) {
        if (!Array.isArray(options.entryTypes)) throw new TypeError('entryTypes must be an array');
        this._entryTypes = options.entryTypes.slice();
      } else if (options.type) {
        this._entryTypes = [String(options.type)];
      } else {
        throw new TypeError('options must contain entryTypes or type');
      }
      this._connected = true;
      if (options.buffered && typeof performance.getEntries === 'function') {
        var entries = performance.getEntries();
        for (var i = 0; i < entries.length; i++) this._enqueue(entries[i]);
      }
      return this;
    }
    disconnect() {
      this._connected = false;
      this._buffer.length = 0;
      var index = observers.indexOf(this);
      if (index !== -1) observers.splice(index, 1);
    }
    takeRecords() { return this._buffer.splice(0); }
  }

  function monitorEventLoopDelay(options) {
    options = options || {};
    var resolution = Number(options.resolution) || 10;
    var hist = new Histogram(options);
    var timer = null;
    var last = now();
    function sample() {
      var current = now();
      var elapsed = (current - last) * 1e6;
      last = current;
      if (elapsed >= 0) hist.record(elapsed);
    }
    return {
      enable: function () { if (timer === null) { last = now(); timer = setInterval(sample, resolution); if (timer && timer.unref) timer.unref(); } },
      disable: function () { if (timer !== null) { clearInterval(timer); timer = null; } },
      reset: function () { hist.reset(); last = now(); },
      get count() { return hist.count; }, get min() { return hist.min; },
      get max() { return hist.max; }, get mean() { return hist.mean; },
      get stddev() { return hist.stddev; },
      percentile: function (p) { return hist.percentile(p); }
    };
  }

  function createHistogram(options) { return new Histogram(options); }

  if (typeof performance.eventLoopUtilization !== 'function') {
    performance.eventLoopUtilization = function () {
      return { idle: 0, active: 0, utilization: 0 };
    };
  }
  if (typeof performance.timerify !== 'function') {
    performance.timerify = function (fn) {
      if (typeof fn !== 'function') throw new TypeError('fn must be a function');
      return function () {
        var start = now();
        try { return fn.apply(this, arguments); }
        finally {
          if (typeof performance.measure === 'function') {
            try { performance.measure(fn.name || 'timerified', { start: start, end: now() }); } catch (e) {}
          }
        }
      };
    };
  }

  class PerformanceNodeTiming {
    constructor() {
      this.name = 'node';
      this.entryType = 'node';
      this.startTime = 0;
      this.duration = 0;
      this.nodeStart = 0;
      this.v8Start = 0;
      this.bootstrapComplete = 0;
      this.environment = 0;
      this.loopStart = 0;
      this.loopExit = -1;
      this.idleTime = 0;
    }
    toJSON() { return Object.assign({}, this); }
  }

  return {
    performance: performance,
    performance_node: performance,
    PerformanceObserver: PerformanceObserver,
    PerformanceObserverEntryList: PerformanceObserverEntryList,
    PerformanceEntry: PerformanceEntry,
    PerformanceNodeTiming: PerformanceNodeTiming,
    monitorEventLoopDelay: monitorEventLoopDelay,
    histogram: createHistogram,
    createHistogram: createHistogram,
    Histogram: Histogram
  };
})`

const traceEventsPrelude = `(function () {
  'use strict';
  var enabled = new Set();
  function validate(options) {
    if (options === null || typeof options !== 'object') {
      throw new TypeError('The "options" argument must be of type object. Received ' + typeof options);
    }
    var categories = options.categories;
    if (categories === undefined) throw new TypeError('options.categories is required');
    if (typeof categories === 'string') categories = categories.split(',').map(function (s) { return s.trim(); }).filter(Boolean);
    if (!Array.isArray(categories)) throw new TypeError('options.categories must be an array or string');
    for (var i = 0; i < categories.length; i++) {
      if (typeof categories[i] !== 'string') throw new TypeError('options.categories must contain strings');
    }
    return categories;
  }
  class Tracing {
    constructor(categories, initiallyEnabled) {
      this.categories = Object.freeze(categories.slice());
      this._enabled = !!initiallyEnabled;
    }
    get enabled() { return this._enabled; }
    enable() { if (!this._enabled) { this._enabled = true; this.categories.forEach(function (c) { enabled.add(c); }); } }
    disable() { if (this._enabled) { this._enabled = false; this.categories.forEach(function (c) { enabled.delete(c); }); } }
    [Symbol.dispose]() { this.disable(); }
  }
  function createTracing(options) {
    var categories = validate(options);
    return new Tracing(categories, options.enabled === true);
  }
  function getEnabledCategories() {
    return Array.from(enabled).sort().join(',');
  }
  return { createTracing: createTracing, getEnabledCategories: getEnabledCategories, Tracing: Tracing };
})`

const inspectorPrelude = `(function () {
  'use strict';
  function Emitter() { this._events = Object.create(null); }
  Emitter.prototype.on = function (name, fn) { (this._events[name] || (this._events[name] = [])).push(fn); return this; };
  Emitter.prototype.addListener = Emitter.prototype.on;
  Emitter.prototype.once = function (name, fn) { var self = this; function once() { self.removeListener(name, once); return fn.apply(self, arguments); } once.listener = fn; return this.on(name, once); };
  Emitter.prototype.removeListener = function (name, fn) { var list = this._events[name]; if (!list) return this; for (var i = 0; i < list.length; i++) if (list[i] === fn || list[i].listener === fn) { list.splice(i, 1); break; } return this; };
  Emitter.prototype.off = Emitter.prototype.removeListener;
  Emitter.prototype.emit = function (name) { var list = this._events[name]; if (!list || !list.length) { if (name === 'error') throw arguments[1]; return false; } var args = Array.prototype.slice.call(arguments, 1); list = list.slice(); for (var i = 0; i < list.length; i++) list[i].apply(this, args); return true; };
  class Session extends Emitter {
    constructor() { super(); this.connected = false; }
    connect() { this.connected = true; this.emit('connect'); return this; }
    connectToMainThread() { return this.connect(); }
    disconnect() { if (!this.connected) return; this.connected = false; this.emit('disconnect'); }
    post(method, params, callback) {
      if (typeof params === 'function') { callback = params; params = {}; }
      if (!this.connected) {
        var err = new Error('Inspector session is not connected');
        err.code = 'ERR_INSPECTOR_NOT_CONNECTED';
        if (typeof callback === 'function') { process.nextTick(callback, err); return; }
        throw err;
      }
      if (typeof callback !== 'function') return;
      var err = new Error('Inspector command is not available in GodeJS');
      err.code = 'ERR_INSPECTOR_NOT_AVAILABLE';
      process.nextTick(callback, err);
    }
  }

  var opened = false;
  var api = {
    Session: Session,
    open: function () { opened = true; },
    close: function () { opened = false; },
    url: function () { return undefined; },
    waitForDebugger: function () {},
    isOpen: function () { return opened; }
  };
  return api;
})`

const replPrelude = `(function (EventEmitter, Interface, vm, util, processObject, globalObject) {
  'use strict';
  var proc = processObject || globalObject.process || {};
  function schedule(fn) { if (proc && typeof proc.nextTick === 'function') proc.nextTick(fn); else Promise.resolve().then(fn); }
  function inspect(value) {
    if (util && typeof util.inspect === 'function') return util.inspect(value, { colors: false, depth: 2 });
    try { return JSON.stringify(value); } catch (e) { return String(value); }
  }
  function write(stream, text) {
    if (stream && typeof stream.write === 'function') stream.write(text);
    else if (proc && proc.stdout && typeof proc.stdout.write === 'function') proc.stdout.write(text);
  }

  class Recoverable extends Error {
    constructor(err) { super(err && err.message ? err.message : String(err)); this.err = err; this.name = 'Recoverable'; }
  }

  class REPLServer extends (Interface || BaseEmitter()) {
    constructor(options) {
      options = options || {};
      var input = options.input || (proc && proc.stdin);
      var output = options.output || (proc && proc.stdout);
      if (!input) throw new TypeError('The "input" option must be a stream');
      super(input, output, {
        terminal: options.terminal === undefined ? !!(output && output.isTTY) : !!options.terminal,
        prompt: options.prompt === undefined ? '> ' : String(options.prompt),
        historySize: options.historySize === undefined ? 1000 : options.historySize
      });
      this.options = options;
      this.useGlobal = !!options.useGlobal;
      this.ignoreUndefined = !!options.ignoreUndefined;
      this.context = this.useGlobal ? globalObject : (options.context || vm.createContext({}));
      this.writer = typeof options.writer === 'function' ? options.writer : inspect;
      this.eval = typeof options.eval === 'function' ? options.eval : this.defaultEval.bind(this);
      this.commands = Object.create(null);
      this._buffered = '';
      this._recoverable = false;
      this._closed = false;
      var self = this;
      this.on('line', function (line) { self._onLine(line); });
      this.on('close', function () { self._closed = true; });
      this.defineCommand('exit', { help: 'Exit the REPL', action: function () { self.close(); } });
      this.defineCommand('help', { help: 'Print this help message', action: function () { self._help(); } });
      this.defineCommand('clear', { help: 'Clear the terminal screen', action: function () { self.clearBufferedCommand(); write(output, '\\u001b[2J\\u001b[0;0H'); self.displayPrompt(); } });
      this.defineCommand('break', { help: 'Break out of a multiline expression', action: function () { self.clearBufferedCommand(); self.displayPrompt(); } });
    }
    defaultEval(code, context, filename, callback) {
      var result;
      try {
        if (this.useGlobal) result = vm.runInThisContext(code, { filename: filename });
        else result = vm.runInContext(code, context, { filename: filename });
        callback(null, result);
      } catch (err) {
        if (err && err.name === 'SyntaxError' && /Unexpected end of input|Unexpected token.*end/.test(err.message)) {
          callback(new Recoverable(err));
        } else {
          callback(err);
        }
      }
    }
    defineCommand(name, command) {
      if (typeof name !== 'string' || !name) throw new TypeError('name must be a non-empty string');
      if (!command || typeof command.action !== 'function') throw new TypeError('command.action must be a function');
      this.commands[name] = { help: command.help || '', action: command.action };
      return this;
    }
    _help() {
      var out = 'Commands:\\n';
      Object.keys(this.commands).forEach(function (name) {
        out += '.' + name + '\\t' + (this.commands[name].help || '') + '\\n';
      }, this);
      write(this.output || proc.stdout, out);
      this.displayPrompt();
    }
    _onLine(line) {
      if (this._closed) return;
      var trimmed = line.trim();
      if (!this._recoverable && trimmed.charAt(0) === '.') {
        var command = this.commands[trimmed.slice(1).split(/\\s+/)[0]];
        if (command) { command.action.call(this, trimmed); return; }
      }
      var source = this._recoverable ? this._buffered + '\\n' + line : line;
      var self = this;
      this.eval(source, this.context, 'repl', function (err, result) {
        if (err instanceof Recoverable) {
          self._buffered = source;
          self._recoverable = true;
          if (typeof self.setPrompt === 'function') self.setPrompt('... ');
          self.displayPrompt();
          return;
        }
        self._buffered = '';
        self._recoverable = false;
        if (err) {
          write(self.output || proc.stdout, (err.stack || String(err)) + '\\n');
        } else if (result !== undefined || !self.ignoreUndefined) {
          write(self.output || proc.stdout, self.writer(result) + '\\n');
        }
        if (typeof self.setPrompt === 'function') self.setPrompt(self.options.prompt === undefined ? '> ' : String(self.options.prompt));
        self.displayPrompt();
      });
    }
    displayPrompt() {
      if (this._closed) return;
      if (typeof this.prompt === 'function') this.prompt();
    }
    clearBufferedCommand() { this._buffered = ''; this._recoverable = false; if (typeof this.setPrompt === 'function') this.setPrompt(this.options.prompt === undefined ? '> ' : String(this.options.prompt)); }
    setupContext(context) { this.context = context || vm.createContext({}); return this.context; }
    close() { if (typeof super.close === 'function') return super.close(); this._closed = true; }
  }

  function BaseEmitter() {
    class Fallback { constructor() { this._events = Object.create(null); } on(n, f) { (this._events[n] || (this._events[n] = [])).push(f); return this; } once(n, f) { var self = this; function once() { self.removeListener(n, once); return f.apply(self, arguments); } return this.on(n, once); } removeListener(n, f) { var l = this._events[n]; if (l) { var i = l.indexOf(f); if (i >= 0) l.splice(i, 1); } return this; } emit(n) { var l = this._events[n]; if (!l) return false; var a = Array.prototype.slice.call(arguments, 1); l.slice().forEach(function (f) { f.apply(this, a); }, this); return true; } }
    return Fallback;
  }

  var REPL_MODE_SLOPPY = 0;
  var REPL_MODE_STRICT = 1;
  function start(options) {
    options = options || {};
    if (options.input === undefined && proc) options.input = proc.stdin;
    if (options.output === undefined && proc) options.output = proc.stdout;
    var server = new REPLServer(options);
    server.displayPrompt();
    return server;
  }

  return {
    start: start,
    REPLServer: REPLServer,
    Recoverable: Recoverable,
    REPL_MODE_SLOPPY: REPL_MODE_SLOPPY,
    REPL_MODE_STRICT: REPL_MODE_STRICT,
    builtinModules: []
  };
})`

const seaPrelude = `(function () {
  'use strict';
  function missing(name) {
    var err = new Error('Asset not found: ' + name);
    err.code = 'ERR_NOT_FOUND';
    return err;
  }
  return {
    isSea: function () { return false; },
    getAsset: function (name) { throw missing(name); },
    getAssetAsBlob: function (name) { throw missing(name); },
    getRawAsset: function (name) { throw missing(name); }
  };
})`
