package http

// The HTTP protocol lives in Go's net/http server. This prelude supplies the
// Node-facing classes and is deliberately free of template literals so it can
// live in a Go raw string.
const prelude = `(function (host, EventEmitter) {
  'use strict';

  var STATUS_CODES = {
    100: 'Continue', 101: 'Switching Protocols', 102: 'Processing', 103: 'Early Hints',
    200: 'OK', 201: 'Created', 202: 'Accepted', 203: 'Non-Authoritative Information',
    204: 'No Content', 205: 'Reset Content', 206: 'Partial Content',
    300: 'Multiple Choices', 301: 'Moved Permanently', 302: 'Found', 303: 'See Other',
    304: 'Not Modified', 305: 'Use Proxy', 307: 'Temporary Redirect', 308: 'Permanent Redirect',
    400: 'Bad Request', 401: 'Unauthorized', 402: 'Payment Required', 403: 'Forbidden',
    404: 'Not Found', 405: 'Method Not Allowed', 406: 'Not Acceptable',
    407: 'Proxy Authentication Required', 408: 'Request Timeout', 409: 'Conflict',
    410: 'Gone', 411: 'Length Required', 412: 'Precondition Failed',
    413: 'Payload Too Large', 414: 'URI Too Long', 415: 'Unsupported Media Type',
    416: 'Range Not Satisfiable', 417: 'Expectation Failed',     418: "I'm a Teapot",
    421: 'Misdirected Request', 422: 'Unprocessable Entity', 423: 'Locked',
    424: 'Failed Dependency', 425: 'Too Early', 426: 'Upgrade Required',
    428: 'Precondition Required', 429: 'Too Many Requests',
    431: 'Request Header Fields Too Large', 451: 'Unavailable For Legal Reasons',
    500: 'Internal Server Error', 501: 'Not Implemented', 502: 'Bad Gateway',
    503: 'Service Unavailable', 504: 'Gateway Timeout',
    505: 'HTTP Version Not Supported', 506: 'Variant Also Negotiates',
    507: 'Insufficient Storage', 508: 'Loop Detected', 510: 'Not Extended',
    511: 'Network Authentication Required'
  };

  var METHODS = [
    'ACL', 'BIND', 'CHECKOUT', 'CONNECT', 'COPY', 'DELETE', 'GET', 'HEAD',
    'LINK', 'LOCK', 'M-SEARCH', 'MERGE', 'MKACTIVITY', 'MKCALENDAR', 'MKCOL',
    'MOVE', 'NOTIFY', 'OPTIONS', 'PATCH', 'POST', 'PROPFIND', 'PROPPATCH',
    'PURGE', 'PUT', 'REBIND', 'REPORT', 'SEARCH', 'SOURCE', 'SUBSCRIBE',
    'TRACE', 'UNBIND', 'UNLINK', 'UNLOCK', 'UNSUBSCRIBE'
  ];

  // Header and connection metadata crosses the Go bridge packed into one
  // string per direction. RS separates records, US separates a name from its
  // value. Both are control characters that valid HTTP field values never
  // contain, so the encoding is unambiguous.
  var RS = '\x1e';
  var US = '\x1f';

  function describe(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (typeof value === 'string') return "type string ('" + value + "')";
    if (typeof value === 'number') return 'type number (' + String(value) + ')';
    if (typeof value === 'boolean') return 'type boolean (' + String(value) + ')';
    if (typeof value === 'function') return 'function ' + (value.name || '(anonymous)');
    if (Array.isArray(value)) return 'an instance of Array';
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }

  function schedule(fn) {
    if (typeof process !== 'undefined' && process && typeof process.nextTick === 'function') {
      process.nextTick(fn);
    } else {
      Promise.resolve().then(fn);
    }
  }

  function invalidArg(name, expected, actual) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(actual));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function invalidHeaderName(name) {
    var err = new TypeError('Header name must be a valid HTTP token ["' + name + '"]');
    err.code = 'ERR_INVALID_HTTP_TOKEN';
    return err;
  }

  function invalidHeaderValue(name, value) {
    var err = new TypeError('Invalid character in header content ["' + name + '"]');
    err.code = 'ERR_INVALID_CHAR';
    return err;
  }

  function validHeaderName(name) {
    return typeof name === 'string' && /^[!#$%&'*+\-.^_\x60|~0-9A-Za-z]+$/.test(name);
  }

  function validHeaderValue(value) {
    return typeof value === 'string' && !/[\r\n]/.test(value);
  }

  function headerValue(value) {
    if (Array.isArray(value)) {
      return value.map(function (v) { return String(v); });
    }
    return [String(value)];
  }

  function toBuffer(chunk, encoding) {
    if (typeof chunk === 'string') return Buffer.from(chunk, encoding || 'utf8');
    if (Buffer.isBuffer(chunk)) return chunk;
    if (chunk instanceof Uint8Array) return Buffer.from(chunk.buffer, chunk.byteOffset, chunk.byteLength);
    throw invalidArg('chunk', 'of type string or an instance of Buffer or Uint8Array', chunk);
  }

  function headersFromRaw(raw) {
    var headers = Object.create(null);
    var distinct = Object.create(null);
    for (var i = 0; i + 1 < raw.length; i += 2) {
      var name = String(raw[i]);
      var key = name.toLowerCase();
      var value = String(raw[i + 1]);
      if (!distinct[key]) distinct[key] = [];
      distinct[key].push(value);
      if (key === 'set-cookie') {
        if (!headers[key]) headers[key] = [];
        headers[key].push(value);
      } else if (key === 'cookie' && headers[key] !== undefined) {
        headers[key] += '; ' + value;
      } else if (headers[key] !== undefined) {
        headers[key] += ', ' + value;
      } else {
        headers[key] = value;
      }
    }
    return { headers: headers, distinct: distinct };
  }

  var BODY_EVENT_NAMES = ['data', 'end', 'close', 'aborted'];
  var FINISH_EVENT_NAMES = ['finish', 'close'];
  var EMPTY_BUFFER = null;

  // Node reports the remote and local address on req.socket, which most
  // handlers never touch. Keeping the pair packed keeps the request free of
  // two address parses; the socket splits it on first access.
  function splitAddress(addr) {
    var at = addr.lastIndexOf(':');
    if (at < 0) return [addr, 0];
    var port = Number(addr.slice(at + 1));
    return [addr.slice(0, at), port > 0 ? port : 0];
  }

  function makeSocket(remoteAddress, remotePort, localAddress, localPort) {
    var socket = new EventEmitter();
    socket.remoteAddress = remoteAddress;
    socket.remotePort = remotePort;
    socket.remoteFamily = remoteAddress && remoteAddress.indexOf(':') !== -1 ? 'IPv6' : 'IPv4';
    socket.localAddress = localAddress;
    socket.localPort = localPort;
    socket.readable = true;
    socket.writable = true;
    socket.destroyed = false;
    socket.connecting = false;
    socket.destroy = function () { this.destroyed = true; this.readable = false; this.writable = false; };
    socket.setTimeout = function () { return this; };
    socket.setNoDelay = function () { return this; };
    socket.setKeepAlive = function () { return this; };
    socket.ref = function () { return this; };
    socket.unref = function () { return this; };
    socket.address = function () {
      return { address: this.localAddress || '127.0.0.1', port: this.localPort || 0 };
    };
    return socket;
  }

  class IncomingMessage extends EventEmitter {
    constructor(id, server, method, url, version, rawHeaders, body, connInfo) {
      super();
      this._id = id;
      this._server = server;
      this._body = body || Buffer.alloc(0);
      this._bodyEmitted = false;
      this._paused = false;
      this._encoding = null;
      this.method = method;
      this.url = url;
      this.httpVersion = version;
      // Parsing in place avoids two array allocations per request.
      var slash = version.indexOf('/');
      var major = 1;
      var minor = 1;
      if (slash >= 0) {
        var protocol = version.slice(slash + 1);
        var dot = protocol.indexOf('.');
        major = Number(dot < 0 ? protocol : protocol.slice(0, dot)) || 1;
        minor = dot < 0 ? 0 : (Number(protocol.slice(dot + 1)) || 0);
      }
      this.httpVersionMajor = major;
      this.httpVersionMinor = minor;
      // Node keeps rawHeaders as a flat name/value array, but building that
      // array across the Go bridge costs a host call per entry. Go packs the
      // pairs into one string and this class decodes them only when the
      // handler actually looks at headers or rawHeaders.
      this._rawHeaders = Array.isArray(rawHeaders) ? rawHeaders : null;
      this._rawHeadersPacked = typeof rawHeaders === 'string' ? rawHeaders : '';
      this._connInfo = connInfo || '';
      this._headers = null;
      this._headersDistinct = null;
      this._socket = null;
      this._trailers = null;
      this._trailersDistinct = null;
      this.aborted = false;
      this.complete = false;
      this.readable = true;
      this.httpVersion = version;
      this.statusCode = undefined;
      this.statusMessage = undefined;
    }
    _parseHeaders() {
      if (this._rawHeaders === null) {
        var packed = this._rawHeadersPacked;
        var raw = [];
        if (packed) {
          var records = packed.split(RS);
          for (var i = 0; i < records.length; i++) {
            var record = records[i];
            if (!record) continue;
            var at = record.indexOf(US);
            if (at < 0) continue;
            raw.push(record.slice(0, at), record.slice(at + 1));
          }
        }
        this._rawHeaders = raw;
      }
      var parsed = headersFromRaw(this._rawHeaders);
      this._headers = parsed.headers;
      this._headersDistinct = parsed.distinct;
    }
    get rawHeaders() {
      if (this._rawHeaders === null) this._parseHeaders();
      return this._rawHeaders;
    }
    set rawHeaders(value) {
      this._rawHeaders = value;
      this._rawHeadersPacked = '';
      this._headers = null;
      this._headersDistinct = null;
    }
    get headers() {
      if (this._headers === null) this._parseHeaders();
      return this._headers;
    }
    set headers(value) { this._headers = value; }
    get headersDistinct() {
      if (this._headersDistinct === null) this._parseHeaders();
      return this._headersDistinct;
    }
    set headersDistinct(value) { this._headersDistinct = value; }
    get trailers() {
      if (this._trailers === null) this._trailers = Object.create(null);
      return this._trailers;
    }
    set trailers(value) { this._trailers = value; }
    get trailersDistinct() {
      if (this._trailersDistinct === null) this._trailersDistinct = Object.create(null);
      return this._trailersDistinct;
    }
    set trailersDistinct(value) { this._trailersDistinct = value; }
    get socket() {
      if (this._socket === null) {
        var ci = this._connInfo ? this._connInfo.split(US) : ['', ''];
        var remote = splitAddress(ci[0] || '');
        var local = splitAddress(ci[1] || '');
        this._socket = makeSocket(remote[0], remote[1], local[0], local[1]);
      }
      return this._socket;
    }
    set socket(value) { this._socket = value; }
    get connection() { return this.socket; }
    set connection(value) { this._socket = value; }
    _hasBodyListeners() {
      // The events module installs its hidden state only once a listener is
      // added, so the common no-listener request never crosses the bridge.
      if (this['$gode$events'] === undefined) return false;
      var any = this._godeHasAnyListeners;
      if (typeof any === 'function') return any.call(this, BODY_EVENT_NAMES);
      return this.listenerCount('data') > 0 || this.listenerCount('end') > 0 ||
        this.listenerCount('close') > 0 || this.listenerCount('aborted') > 0;
    }
    _completeWithoutListeners() {
      this.complete = true;
      this.readable = false;
    }
    _emitBody() {
      if (this._bodyEmitted || this._paused) return;
      this._bodyEmitted = true;
      if (this._body.length > 0) {
        var data = this._encoding ? this._body.toString(this._encoding) : this._body;
        this.emit('data', data);
      }
      this.complete = true;
      this.readable = false;
      this.emit('end');
      if (this.listenerCount('close') > 0) {
        var self = this;
        schedule(function () { self.emit('close'); });
      }
    }
    pause() { this._paused = true; return this; }
    resume() { this._paused = false; this._emitBody(); return this; }
    isPaused() { return this._paused; }
    setEncoding(encoding) {
      if (!Buffer.isEncoding || !Buffer.isEncoding(encoding)) {
        throw new TypeError('Unknown encoding: ' + encoding);
      }
      this._encoding = encoding;
      return this;
    }
    setTimeout(ms, cb) {
      if (typeof cb === 'function') this.once('timeout', cb);
      return this;
    }
    destroy(err) {
      if (this.aborted) return this;
      this.aborted = true;
      this.readable = false;
      this.emit('aborted');
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
    read() { return null; }
  }

  class ServerResponse extends EventEmitter {
    constructor(id, req) {
      super();
      this._id = id;
      this.req = req;
      this._socket = null;
      this.statusCode = 200;
      this.statusMessage = undefined;
      this.headersSent = false;
      this.sendDate = true;
      this.finished = false;
      this.writableEnded = false;
      this.writableFinished = false;
      this.writable = true;
      this._headers = Object.create(null);
      this._headerNames = [];
      this._trailers = null;
      this._wrote = false;
      this._headPending = false;
      this._inDispatch = false;
      this._syncChunks = null;
      this._syncStatus = 0;
      this._syncHeaders = '';
      this._syncBody = null;
      this._syncDone = false;
    }
    get socket() { return this._socket === null && this.req ? this.req.socket : this._socket; }
    set socket(value) { this._socket = value; }
    get connection() { return this.socket; }
    set connection(value) { this._socket = value; }
    setHeader(name, value) {
      if (this.headersSent) {
        var err = new Error('Cannot set headers after they are sent to the client');
        err.code = 'ERR_HTTP_HEADERS_SENT';
        throw err;
      }
      if (!validHeaderName(name)) throw invalidHeaderName(name);
      var values = headerValue(value);
      for (var i = 0; i < values.length; i++) {
        if (!validHeaderValue(values[i])) throw invalidHeaderValue(name, values[i]);
      }
      var key = String(name).toLowerCase();
      if (!this._headers[key]) this._headerNames.push(String(name));
      this._headers[key] = { name: String(name), value: values };
      return this;
    }
    appendHeader(name, value) {
      if (this.headersSent) {
        var err = new Error('Cannot set headers after they are sent to the client');
        err.code = 'ERR_HTTP_HEADERS_SENT';
        throw err;
      }
      if (!validHeaderName(name)) throw invalidHeaderName(name);
      var values = headerValue(value);
      var key = String(name).toLowerCase();
      if (!this._headers[key]) {
        this._headerNames.push(String(name));
        this._headers[key] = { name: String(name), value: [] };
      }
      for (var i = 0; i < values.length; i++) {
        if (!validHeaderValue(values[i])) throw invalidHeaderValue(name, values[i]);
        this._headers[key].value.push(values[i]);
      }
      return this;
    }
    getHeader(name) {
      var entry = this._headers[String(name).toLowerCase()];
      if (!entry) return undefined;
      return entry.value.length === 1 ? entry.value[0] : entry.value.slice();
    }
    getHeaders() {
      var out = Object.create(null);
      for (var key in this._headers) {
        var entry = this._headers[key];
        out[key] = entry.value.length === 1 ? entry.value[0] : entry.value.slice();
      }
      return out;
    }
    getHeaderNames() {
      var out = [];
      for (var key in this._headers) out.push(key);
      return out;
    }
    getRawHeaderNames() { return this._headerNames.slice(); }
    hasHeader(name) { return !!this._headers[String(name).toLowerCase()]; }
    removeHeader(name) {
      if (this.headersSent) {
        var err = new Error('Cannot remove headers after they are sent to the client');
        err.code = 'ERR_HTTP_HEADERS_SENT';
        throw err;
      }
      var key = String(name).toLowerCase();
      if (!this._headers[key]) return this;
      delete this._headers[key];
      for (var i = 0; i < this._headerNames.length; i++) {
        if (this._headerNames[i].toLowerCase() === key) {
          this._headerNames.splice(i, 1);
          break;
        }
      }
      return this;
    }
    _headersObject() {
      var out = Object.create(null);
      for (var key in this._headers) {
        var entry = this._headers[key];
        out[entry.name] = entry.value.length === 1 ? entry.value[0] : entry.value.slice();
      }
      return out;
    }
    _packedHeaders() {
      var out = '';
      for (var key in this._headers) {
        var entry = this._headers[key];
        for (var i = 0; i < entry.value.length; i++) {
          out += RS + entry.name + US + entry.value[i];
        }
      }
      return out;
    }
    _sendHead() {
      if (this.headersSent) return;
      this.headersSent = true;
      if (this.statusMessage === undefined) {
        this.statusMessage = STATUS_CODES[this.statusCode] || '';
      }
      // Keep the headers in the JS object until the first body write. A
      // writeHead followed by end can then cross the Go bridge once instead
      // of once for the head and once for the body, and a synchronous
      // response never builds the head object at all.
      this._headPending = true;
    }
    _flushHead() {
      if (!this._headPending) return;
      this._headPending = false;
      host.respHead.call(this, this.statusCode, this.statusMessage, this._headersObject());
    }
    writeHead(statusCode, statusMessage, headers) {
      if (this.headersSent) {
        var err = new Error('Cannot write headers after they are sent to the client');
        err.code = 'ERR_HTTP_HEADERS_SENT';
        throw err;
      }
      if (typeof statusMessage === 'object' && statusMessage !== null) {
        headers = statusMessage;
        statusMessage = undefined;
      }
      if (statusMessage !== undefined && typeof statusMessage !== 'string') {
        throw invalidArg('statusMessage', 'of type string', statusMessage);
      }
      this.statusCode = Number(statusCode);
      if (statusMessage !== undefined) this.statusMessage = statusMessage;
      if (headers) {
        for (var key in headers) {
          if (Object.prototype.hasOwnProperty.call(headers, key)) this.setHeader(key, headers[key]);
        }
      }
      this._sendHead();
      return this;
    }
    write(chunk, encoding, cb) {
      if (this.writableEnded) {
        var err = new Error('write after end');
        err.code = 'ERR_STREAM_WRITE_AFTER_END';
        throw err;
      }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var data = toBuffer(chunk, encoding);
      if (!this.headersSent) this._sendHead();
      if (this._inDispatch) {
        if (this._syncChunks === null) this._syncChunks = [];
        this._syncChunks.push(data);
      } else if (this._headPending) {
        this._headPending = false;
        host.respWriteAll.call(this, this.statusCode, this._headersObject(), data);
      } else {
        host.respWrite.call(this, data);
      }
      this._wrote = true;
      if (typeof cb === 'function') schedule(function () { cb(null); });
      return true;
    }
    end(chunk, encoding, cb) {
      if (this.writableEnded) {
        if (typeof cb === 'function') process.nextTick(cb);
        return this;
      }
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; encoding = undefined; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var data = null;
      var strBody = null;
      if (chunk === undefined || chunk === null) {
        data = EMPTY_BUFFER !== null ? EMPTY_BUFFER : (EMPTY_BUFFER = Buffer.alloc(0));
      } else if (this._inDispatch && this._syncChunks === null && typeof chunk === 'string' &&
                 (encoding === undefined || encoding === null || encoding === 'utf8' || encoding === 'utf-8')) {
        // A synchronous handler that ends with a UTF-8 string needs no Buffer
        // at all: Go converts the string straight to bytes. Non-UTF-8 encodings
        // still have to be decoded through Buffer.
        strBody = chunk;
      } else {
        data = toBuffer(chunk, encoding);
      }
      if (!this.headersSent) {
        // The generated Content-Length is already a valid token and value;
        // skip the user-input checks and lowercasing on this hot path.
        if (this._headers['content-length'] === undefined && this._headers['transfer-encoding'] === undefined) {
          var length = strBody !== null ? Buffer.byteLength(strBody, encoding || 'utf8') : data.length;
          this._headers['content-length'] = { name: 'Content-Length', value: [String(length)] };
          this._headerNames.push('Content-Length');
        }
        this._sendHead();
      }
      if (this._inDispatch) {
        // A handler that ends synchronously never needs a second Go bridge
        // call. Hand the complete response back to dispatch(), which writes
        // it to the waiting Go request goroutine directly.
        var syncBody = strBody !== null ? strBody : data;
        if (this._syncChunks !== null && this._syncChunks.length > 0) {
          this._syncChunks.push(data);
          syncBody = Buffer.concat(this._syncChunks);
        }
        this._syncStatus = this.statusCode;
        this._syncHeaders = this._packedHeaders();
        this._syncBody = syncBody;
        this._syncDone = true;
        this._headPending = false;
      } else if (this._headPending) {
        this._headPending = false;
        host.respEndAll.call(this, this.statusCode, this._headersObject(), data);
      } else {
        host.respEnd.call(this, data);
      }
      this.writableEnded = true;
      this.finished = true;
      var any = this._godeHasAnyListeners;
      var needsEvents = this['$gode$events'] !== undefined && (typeof any === 'function'
        ? any.call(this, FINISH_EVENT_NAMES)
        : (this.listenerCount('finish') > 0 || this.listenerCount('close') > 0));
      if (needsEvents) {
        var self = this;
        schedule(function () {
          self.writableFinished = true;
          self.emit('finish');
          self.emit('close');
        });
      } else {
        this.writableFinished = true;
      }
      if (typeof cb === 'function') schedule(cb);
      return this;
    }
    flushHeaders() {
      if (!this.headersSent) this._sendHead();
      this._flushHead();
    }
    writeContinue() { return this; }
    writeProcessing() { return this; }
    addTrailers(trailers) { this._trailers = trailers; return this; }
    setTimeout(ms, cb) {
      if (typeof cb === 'function') this.once('timeout', cb);
      return this;
    }
    destroy(err) {
      if (this.destroyed) return this;
      this.destroyed = true;
      this.writable = false;
      host.respAbort.call(this, err || null);
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
  }

  class ClientRequest extends EventEmitter {
    constructor(options, callback) {
      super();
      this.method = String(options.method || 'GET').toUpperCase();
      this.path = options.path || '/';
      this.aborted = false;
      this._url = options._url;
      this._options = options._options || options;
      this._headers = Object.create(null);
      this._headerNames = [];
      this._chunks = [];
      this._ended = false;
      if (typeof callback === 'function') this.once('response', callback);
    }
    setHeader(name, value) {
      if (!validHeaderName(name)) throw invalidHeaderName(name);
      var values = headerValue(value);
      for (var i = 0; i < values.length; i++) {
        if (!validHeaderValue(values[i])) throw invalidHeaderValue(name, values[i]);
      }
      var key = String(name).toLowerCase();
      if (!this._headers[key]) this._headerNames.push(String(name));
      this._headers[key] = { name: String(name), value: values };
      return this;
    }
    getHeader(name) {
      var entry = this._headers[String(name).toLowerCase()];
      if (!entry) return undefined;
      return entry.value.length === 1 ? entry.value[0] : entry.value.slice();
    }
    getHeaders() {
      var out = Object.create(null);
      for (var key in this._headers) {
        var entry = this._headers[key];
        out[key] = entry.value.length === 1 ? entry.value[0] : entry.value.slice();
      }
      return out;
    }
    removeHeader(name) {
      var key = String(name).toLowerCase();
      delete this._headers[key];
      for (var i = 0; i < this._headerNames.length; i++) {
        if (this._headerNames[i].toLowerCase() === key) {
          this._headerNames.splice(i, 1);
          break;
        }
      }
      return this;
    }
    _headersObject() {
      var out = Object.create(null);
      for (var key in this._headers) {
        var entry = this._headers[key];
        out[entry.name] = entry.value.length === 1 ? entry.value[0] : entry.value.slice();
      }
      return out;
    }
    write(chunk, encoding, cb) {
      if (this._ended) {
        var err = new Error('write after end');
        err.code = 'ERR_STREAM_WRITE_AFTER_END';
        throw err;
      }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      this._chunks.push(toBuffer(chunk, encoding));
      if (typeof cb === 'function') schedule(function () { cb(null); });
      return true;
    }
    end(chunk, encoding, cb) {
      if (this._ended) return this;
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; encoding = undefined; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      if (chunk !== undefined && chunk !== null) this._chunks.push(toBuffer(chunk, encoding));
      this._ended = true;
      var body = this._chunks.length ? Buffer.concat(this._chunks) : Buffer.alloc(0);
      host.clientRequest.call(this, this.method, this._url, this._headersObject(), body, this._options);
      if (typeof cb === 'function') schedule(cb);
      return this;
    }
    abort() {
      if (!this.aborted) {
        this.aborted = true;
        this.emit('abort');
      }
      return this;
    }
    setTimeout() { return this; }
    flushHeaders() { return this; }
  }

  class Server extends EventEmitter {
    constructor(options, listener) {
      super();
      if (typeof options === 'function') {
        listener = options;
        options = {};
      }
      this._options = options || {};
      this.listening = false;
      this.maxHeadersCount = this._options.maxHeadersCount === undefined ? 2000 : this._options.maxHeadersCount;
      this.headersTimeout = this._options.headersTimeout === undefined ? 60000 : this._options.headersTimeout;
      this.requestTimeout = this._options.requestTimeout === undefined ? 300000 : this._options.requestTimeout;
      this.keepAliveTimeout = this._options.keepAliveTimeout === undefined ? 5000 : this._options.keepAliveTimeout;
      this.timeout = this._options.timeout === undefined ? 0 : this._options.timeout;
      host.initServer.call(this, this._options);
      this._fastRequestListener = null;
      if (typeof listener === 'function') {
        // Register normally for listeners()/emit(), but remember the common
        // single-listener case so dispatch can call it without a Go event
        // emitter round trip. Any later mutation falls back to emit().
        super.on('request', listener);
        this._fastRequestListener = listener;
      }
    }
    on(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.on(name, fn);
    }
    addListener(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.addListener(name, fn);
    }
    once(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.once(name, fn);
    }
    prependListener(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.prependListener(name, fn);
    }
    prependOnceListener(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.prependOnceListener(name, fn);
    }
    removeListener(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.removeListener(name, fn);
    }
    off(name, fn) {
      if (name === 'request') this._fastRequestListener = null;
      return super.off(name, fn);
    }
    removeAllListeners(name) {
      if (name === undefined || name === 'request') this._fastRequestListener = null;
      return name === undefined ? super.removeAllListeners() : super.removeAllListeners(name);
    }
    listen() {
      return host.srvListen.call(this, Array.prototype.slice.call(arguments));
    }
    close(cb) {
      return host.srvClose.call(this, cb);
    }
    address() {
      return host.srvAddress.call(this);
    }
    getConnections(cb) {
      return host.srvGetConnections.call(this, cb);
    }
    ref() { return this; }
    unref() { return this; }
    closeAllConnections() { return this; }
    closeIdleConnections() { return this; }
  }

  class Agent extends EventEmitter {
    constructor(options) {
      super();
      this.options = options || {};
      this.maxSockets = this.options.maxSockets === undefined ? Infinity : this.options.maxSockets;
      this.keepAlive = !!this.options.keepAlive;
      this.sockets = Object.create(null);
      this.requests = Object.create(null);
    }
    destroy() {}
  }

  function createServer(options, listener) {
    if (typeof options === 'function') {
      listener = options;
      options = {};
    }
    return new Server(options || {}, listener);
  }

  function request(options, callback) { return makeRequest(options, callback, false); }

  function makeRequest(options, callback, secure) {
    var method = 'GET';
    var headers = null;
    var url = null;
    var expectedProtocol = secure ? 'https:' : 'http:';
    if (typeof options === 'string') {
      url = options;
      if (url.indexOf(expectedProtocol + '//') !== 0) {
        var stringErr = new TypeError('Protocol "' + url.split(':')[0] + ':" not supported. Expected "' + expectedProtocol + '"');
        stringErr.code = 'ERR_INVALID_PROTOCOL';
        throw stringErr;
      }
    } else if (options && typeof options === 'object') {
      if (options.href) {
        url = options.href;
      } else {
        var protocol = options.protocol || expectedProtocol;
        if (protocol !== expectedProtocol) {
          var protoErr = new TypeError('Protocol "' + protocol + '" not supported. Expected "' + expectedProtocol + '"');
          protoErr.code = 'ERR_INVALID_PROTOCOL';
          throw protoErr;
        }
        var host = options.hostname || options.host || 'localhost';
        var port = options.port !== undefined ? ':' + options.port : '';
        var path = options.path || '/';
        url = expectedProtocol + '//' + host + port + path;
      }
      method = options.method || method;
      headers = options.headers;
    }
    if (!url) throw invalidArg('options', 'of type string or an instance of URL', options);
    if (secure && String(url).indexOf('https://') !== 0) {
      var httpsErr = new TypeError('Protocol "http:" not supported. Expected "https:"');
      httpsErr.code = 'ERR_INVALID_PROTOCOL';
      throw httpsErr;
    }
    var req = new ClientRequest({ method: String(method).toUpperCase(), _url: String(url), _options: options || {} }, callback);
    if (headers) {
      for (var key in headers) {
        if (Object.prototype.hasOwnProperty.call(headers, key)) req.setHeader(key, headers[key]);
      }
    }
    return req;
  }

  function get(options, callback) {
    var req = request(options, callback);
    req.end();
    return req;
  }

  function requestSecure(options, callback) { return makeRequest(options, callback, true); }

  function getSecure(options, callback) {
    var req = requestSecure(options, callback);
    req.end();
    return req;
  }

  function createSecureServer(options, listener) {
    if (typeof options === 'function') {
      listener = options;
      options = {};
    }
    var secureOptions = {};
    if (options && typeof options === 'object') {
      for (var key in options) {
        if (Object.prototype.hasOwnProperty.call(options, key)) secureOptions[key] = options[key];
      }
    }
    secureOptions._tls = true;
    return new Server(secureOptions, listener);
  }

  function clientDispatch(req, statusCode, statusMessage, version, rawHeaders, body, method, url) {
    var res = new IncomingMessage(0, null, method, url, version, rawHeaders, body, '');
    res.statusCode = statusCode;
    res.statusMessage = statusMessage;
    res.req = req;
    req.emit('response', res);
    schedule(function () { res._emitBody(); });
  }

  // Reused across requests: a synchronous response is returned to Go before
  // the next dispatch can start, so one array serves every request and saves
  // an object allocation per response.
  var syncOut = [0, '', null];

  function dispatch(server, id, head, packedHeaders, body) {
    // head is method US url US version US remoteAddress US localAddress.
    var at = head.indexOf(US);
    var method = at < 0 ? head : head.slice(0, at);
    var rest = at < 0 ? '' : head.slice(at + 1);
    at = rest.indexOf(US);
    var url = at < 0 ? rest : rest.slice(0, at);
    rest = at < 0 ? '' : rest.slice(at + 1);
    at = rest.indexOf(US);
    var version = at < 0 ? rest : rest.slice(0, at);
    rest = at < 0 ? '' : rest.slice(at + 1);
    at = rest.indexOf(US);
    var remoteAddress = at < 0 ? rest : rest.slice(0, at);
    var localAddress = at < 0 ? '' : rest.slice(at + 1);
    var req = new IncomingMessage(id, server, method, url, version, packedHeaders, body,
      remoteAddress + US + localAddress);
    var res = new ServerResponse(id, req);
    var fast = server._fastRequestListener;
    res._inDispatch = true;
    try {
      if (typeof fast === 'function') fast.call(server, req, res);
      else server.emit('request', req, res);
    } finally {
      res._inDispatch = false;
    }
    // EventEmitter bodies are lazy now. A handler that never registered a
    // body listener does not need a nextTick closure just to emit nothing.
    if (req._hasBodyListeners()) schedule(function () { req._emitBody(); });
    else req._completeWithoutListeners();
    if (res._syncDone) {
      syncOut[0] = res._syncStatus;
      syncOut[1] = res._syncHeaders;
      syncOut[2] = res._syncBody;
      return syncOut;
    }
    return undefined;
  }

  return {
    Server: Server,
    IncomingMessage: IncomingMessage,
    ServerResponse: ServerResponse,
    ClientRequest: ClientRequest,
    ClientResponse: IncomingMessage,
    Agent: Agent,
    createServer: createServer,
    request: request,
    get: get,
    _createSecureServer: createSecureServer,
    _requestSecure: requestSecure,
    _getSecure: getSecure,
    dispatch: dispatch,
    clientDispatch: clientDispatch,
    STATUS_CODES: STATUS_CODES,
    METHODS: METHODS,
    maxHeaderSize: 16384,
    globalAgent: new Agent()
  };
})`
