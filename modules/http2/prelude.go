package http2

// The prelude is the Node-facing half of the bridge. Go owns the wire protocol
// and calls dispatch() from its pump; every host response method forwards one
// response command back to the corresponding Go handler.
const prelude = `(function (host, EventEmitter) {
  'use strict';

  var STATUS_CODES = { 200: 'OK', 201: 'Created', 204: 'No Content', 400: 'Bad Request', 404: 'Not Found', 500: 'Internal Server Error' };

  function schedule(fn) {
    if (typeof process !== 'undefined' && process && typeof process.nextTick === 'function') process.nextTick(fn);
    else Promise.resolve().then(fn);
  }
  function invalidArg(name, expected, actual) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + String(actual));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
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
      } else if (headers[key] !== undefined) {
        headers[key] += ', ' + value;
      } else {
        headers[key] = value;
      }
    }
    return { headers: headers, distinct: distinct };
  }
  function makeSocket() {
    var s = new EventEmitter();
    s.remoteAddress = '127.0.0.1';
    s.remotePort = 0;
    s.localAddress = '127.0.0.1';
    s.localPort = 0;
    s.destroyed = false;
    s.destroy = function () { this.destroyed = true; };
    s.setTimeout = function () { return this; };
    s.setNoDelay = function () { return this; };
    s.setKeepAlive = function () { return this; };
    s.ref = function () { return this; };
    s.unref = function () { return this; };
    s.address = function () { return { address: this.localAddress, port: this.localPort }; };
    return s;
  }

  class Http2ServerRequest extends EventEmitter {
    constructor(id, server, method, url, scheme, authority, rawHeaders, body, remoteAddress, remotePort, localAddress, localPort) {
      super();
      this._id = id;
      this._server = server;
      this._body = body || Buffer.alloc(0);
      this._bodyEmitted = false;
      this._paused = false;
      this._encoding = null;
      this.method = method;
      this.url = url;
      this.scheme = scheme;
      this.authority = authority;
      this.httpVersion = '2.0';
      this.httpVersionMajor = 2;
      this.httpVersionMinor = 0;
      this.rawHeaders = rawHeaders;
      var parsed = headersFromRaw(rawHeaders);
      this.headers = parsed.headers;
      this.headersDistinct = parsed.distinct;
      this.trailers = Object.create(null);
      this.trailersDistinct = Object.create(null);
      this.aborted = false;
      this.complete = false;
      this.readable = true;
      this.socket = makeSocket();
      this.socket.remoteAddress = remoteAddress;
      this.socket.remotePort = remotePort;
      this.socket.localAddress = localAddress;
      this.socket.localPort = localPort;
      this.connection = this.socket;
    }
    _emitBody() {
      if (this._bodyEmitted || this._paused) return;
      this._bodyEmitted = true;
      if (this._body.length) this.emit('data', this._encoding ? this._body.toString(this._encoding) : this._body);
      this.complete = true;
      this.readable = false;
      this.emit('end');
      var self = this;
      schedule(function () { self.emit('close'); });
    }
    pause() { this._paused = true; return this; }
    resume() { this._paused = false; this._emitBody(); return this; }
    isPaused() { return this._paused; }
    setEncoding(encoding) { this._encoding = encoding; return this; }
    setTimeout() { return this; }
    read() { return null; }
    destroy(err) {
      if (this.aborted) return this;
      this.aborted = true;
      this.emit('aborted');
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
  }

  class Http2ServerResponse extends EventEmitter {
    constructor(id, req, stream) {
      super();
      this._id = id;
      this.req = req;
      this.stream = stream;
      this.socket = req.socket;
      this.connection = req.socket;
      this.statusCode = 200;
      this.statusMessage = undefined;
      this.headersSent = false;
      this.sendDate = false;
      this.finished = false;
      this.writableEnded = false;
      this.writableFinished = false;
      this.writable = true;
      this.destroyed = false;
      this._headers = Object.create(null);
      this._names = [];
      this._trailers = null;
    }
    setHeader(name, value) {
      if (this.headersSent) throw new Error('Cannot set headers after they are sent to the client');
      var key = String(name).toLowerCase();
      if (!this._headers[key]) this._names.push(String(name));
      this._headers[key] = { name: String(name), value: value };
      return this;
    }
    appendHeader(name, value) {
      if (this.headersSent) throw new Error('Cannot append headers after they are sent to the client');
      var key = String(name).toLowerCase();
      if (!this._headers[key]) this._names.push(String(name));
      var old = this._headers[key];
      if (!old) this._headers[key] = { name: String(name), value: value };
      else if (Array.isArray(old.value)) old.value = old.value.concat(value);
      else old.value = [old.value].concat(value);
      return this;
    }
    getHeader(name) { var e = this._headers[String(name).toLowerCase()]; return e ? e.value : undefined; }
    getHeaders() { var o = Object.create(null); for (var k in this._headers) o[k] = this._headers[k].value; return o; }
    getHeaderNames() { return this._names.slice(); }
    hasHeader(name) { return !!this._headers[String(name).toLowerCase()]; }
    removeHeader(name) { var k = String(name).toLowerCase(); delete this._headers[k]; return this; }
    _headersObject() {
      var o = Object.create(null);
      for (var k in this._headers) o[this._headers[k].name] = this._headers[k].value;
      return o;
    }
    _sendHead() {
      if (this.headersSent) return;
      this.headersSent = true;
      host.respHead.call(this, this.statusCode, this._headersObject());
    }
    writeHead(statusCode, statusMessage, headers) {
      if (this.headersSent) throw new Error('Cannot write headers after they are sent to the client');
      if (typeof statusMessage === 'object' && statusMessage !== null) { headers = statusMessage; statusMessage = undefined; }
      this.statusCode = Number(statusCode);
      if (headers) for (var k in headers) if (Object.prototype.hasOwnProperty.call(headers, k)) this.setHeader(k, headers[k]);
      this._sendHead();
      return this;
    }
    write(chunk, encoding, cb) {
      if (this.writableEnded) { var e = new Error('write after end'); e.code = 'ERR_STREAM_WRITE_AFTER_END'; throw e; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var b = toBuffer(chunk, encoding);
      if (!this.headersSent) this._sendHead();
      host.respWrite.call(this, b);
      if (typeof cb === 'function') schedule(function () { cb(null); });
      return true;
    }
    end(chunk, encoding, cb) {
      if (this.writableEnded) return this;
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; encoding = undefined; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var b = chunk === undefined || chunk === null ? Buffer.alloc(0) : toBuffer(chunk, encoding);
      if (!this.headersSent) {
        if (!this.hasHeader('content-length') && !this.hasHeader('transfer-encoding')) this.setHeader('content-length', String(b.length));
        this._sendHead();
      }
      this._emitTrailers();
      host.respEnd.call(this, b);
      this.writableEnded = true;
      this.finished = true;
      var self = this;
      schedule(function () { self.writableFinished = true; self.emit('finish'); self.emit('close'); });
      if (typeof cb === 'function') schedule(cb);
      return this;
    }
    flushHeaders() { if (!this.headersSent) this._sendHead(); }
    writeContinue() { return this; }
    writeProcessing() { return this; }
    addTrailers(t) {
      if (typeof t !== 'object' || t === null) throw invalidArg('trailers', 'an object', t);
      this._trailers = t;
      return this;
    }
    _emitTrailers() {
      if (!this._trailers) return;
      host.respTrailers.call(this, this._trailers);
      this._trailers = null;
    }
    setTimeout() { return this; }
    destroy(err) {
      if (this.destroyed) return this;
      this.destroyed = true;
      host.respAbort.call(this, err || null);
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
  }

  class Http2Stream extends EventEmitter {
    constructor(id, server, req, res) {
      super();
      this._id = id;
      this._server = server;
      this._req = req;
      this._res = res;
      this.id = id;
      this.session = server;
      this.closed = false;
      this.destroyed = false;
      this.headersSent = false;
      this.rstCode = 0;
      this.aborted = false;
      this.sentHeaders = false;
    }
    respond(headers, options) {
      if (this.headersSent) return;
      headers = headers || {};
      var status = headers[':status'];
      delete headers[':status'];
      this._res.statusCode = status === undefined ? 200 : Number(status);
      for (var k in headers) if (Object.prototype.hasOwnProperty.call(headers, k)) this._res.setHeader(k, headers[k]);
      this._res._sendHead();
      this.headersSent = true;
      this.sentHeaders = true;
      this.emit('headers', headers);
    }
    additionalHeaders(headers) {
      if (!this.headersSent) throw new Error('Headers have not been sent yet');
      for (var k in headers) this._res.setHeader(k, headers[k]);
      return this;
    }
    write(chunk, encoding, cb) { return this._res.write(chunk, encoding, cb); }
    setEncoding(encoding) { this._req.setEncoding(encoding); return this; }
    pause() { this._req.pause(); return this; }
    resume() { this._req.resume(); return this; }
    isPaused() { return this._req.isPaused(); }
    read() { return null; }
    addTrailers(trailers) { this._res.addTrailers(trailers); return this; }
    end(chunk, encoding, cb) {
      if (this._res.writableEnded) return this;
      this._res.end(chunk, encoding, cb);
      this.closed = true;
      this.emit('close');
      return this;
    }
    close(code, cb) {
      if (!this._res.writableEnded) this._res.end();
      this.closed = true;
      if (typeof cb === 'function') schedule(cb);
    }
    destroy(err) {
      this.destroyed = true;
      this._res.destroy(err);
      this.emit('close');
    }
    setTimeout() { return this; }
    pushStream(headers, options, cb) {
      if (typeof options === 'function') { cb = options; options = {}; }
      if (typeof headers !== 'object' || headers === null) throw invalidArg('headers', 'an object', headers);
      if (typeof headers[':path'] !== 'string' || headers[':path'] === '') {
        var pe = new TypeError('The ":path" pseudo-header is required');
        pe.code = 'ERR_INVALID_ARG_VALUE';
        throw pe;
      }
      var err = null;
      try { host.respPush.call(this, headers, options || {}); } catch (e) { err = e; }
      var pushed = new EventEmitter();
      pushed.id = 0;
      pushed.session = this.session;
      pushed.closed = false;
      pushed.destroyed = false;
      pushed.headersSent = false;
      pushed.respond = function () { this.headersSent = true; };
      pushed.write = function () { return true; };
      pushed.end = function () { this.closed = true; this.emit('close'); return this; };
      pushed.close = function () { this.end(); };
      pushed.destroy = function () { this.end(); };
      if (typeof cb === 'function') {
        schedule(function () { cb(err, pushed, headers); });
      }
      return pushed;
    }
    respondWithFile(path, options, cb) {
      if (typeof options === 'function') { cb = options; options = {}; }
      if (typeof path !== 'string') throw invalidArg('path', 'a string', path);
      var err = null;
      try { host.respFile.call(this, path, options || {}); } catch (e) { err = e; }
      if (typeof cb === 'function') schedule(function () { cb(err || null); });
      else if (err) throw err;
      return this;
    }
    respondWithFD(fd, options, cb) {
      if (typeof options === 'function') { cb = options; options = {}; }
      if (typeof fd !== 'number' || fd < 0) throw invalidArg('fd', 'a non-negative number', fd);
      var err = null;
      try { host.respFD.call(this, fd, options || {}); } catch (e) { err = e; }
      if (typeof cb === 'function') schedule(function () { cb(err || null); });
      else if (err) throw err;
      return this;
    }
  }

  class Http2Server extends EventEmitter {
    constructor(options, listener, secure) {
      super();
      if (typeof options === 'function') { listener = options; options = {}; }
      this._options = options || {};
      this._secure = !!secure;
      this.listening = false;
      this.maxHeadersCount = 2000;
      this.timeout = this._options.timeout || 0;
      host.srvInit.call(this, this._secure);
      if (typeof listener === 'function') this.on('request', listener);
    }
    listen() { return host.srvListen.call(this, Array.prototype.slice.call(arguments)); }
    close(cb) { return host.srvClose.call(this, cb); }
    address() { return host.srvAddress.call(this); }
    getConnections(cb) { if (typeof cb === 'function') schedule(function () { cb(null, 0); }); }
    setTimeout(ms, cb) { if (typeof cb === 'function') this.once('timeout', cb); return this; }
    ref() { return this; }
    unref() { return this; }
  }

  class ClientHttp2Stream extends EventEmitter {
    constructor(session, headers) {
      super();
      this._session = session;
      this._headers = headers || {};
      this._id = 0;
      this._chunks = [];
      this._ended = false;
      this._encoding = null;
      this.id = 0;
      this.session = session;
      this.aborted = false;
      this.closed = false;
      this.destroyed = false;
      this.rstCode = 0;
      this.sentHeaders = false;
      this.headersSent = false;
    }
    setEncoding(encoding) { this._encoding = encoding; return this; }
    write(chunk, encoding, cb) {
      if (this._ended) throw new Error('write after end');
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var b = toBuffer(chunk, encoding);
      host.streamWrite.call(this, b);
      if (typeof cb === 'function') schedule(function () { cb(null); });
      return true;
    }
    end(chunk, encoding, cb) {
      if (this._ended) return this;
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; encoding = undefined; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var b = chunk === undefined || chunk === null ? Buffer.alloc(0) : toBuffer(chunk, encoding);
      this._ended = true;
      host.streamEnd.call(this, b);
      if (typeof cb === 'function') schedule(cb);
      return this;
    }
    close(code, cb) { this.closed = true; host.respAbort.call(this, null); if (typeof cb === 'function') schedule(cb); }
    destroy(err) { this.destroyed = true; this.emit('close'); if (err) this.emit('error', err); }
    setTimeout() { return this; }
    pause() { return this; }
    resume() { return this; }
    isPaused() { return false; }
    priority() { return this; }
  }

  class ClientHttp2Session extends EventEmitter {
    constructor(authority, options) {
      super();
      if (typeof authority === 'string' && authority.indexOf('http') !== 0) {
        var e = new TypeError('The URL must use the http: or https: protocol');
        e.code = 'ERR_INVALID_ARG_VALUE';
        throw e;
      }
      this._options = options || {};
      this._id = 0;
      this.closed = false;
      this.destroyed = false;
      this.connecting = true;
      this.socket = makeSocket();
      this.remoteSettings = {};
      host.connect.call(this, String(authority), this._options);
      this.connecting = false;
      var self = this;
      schedule(function () { self.emit('connect', self, 1); });
    }
    request(headers, options) {
      headers = headers || {};
      var stream = new ClientHttp2Stream(this, headers);
      host.request.call(this, stream, headers, options || {});
      return stream;
    }
    close(cb) { this.closed = true; host.sessionClose.call(this); if (typeof cb === 'function') schedule(cb); }
    destroy() { this.close(); this.destroyed = true; }
    ping(cb) { if (typeof cb === 'function') schedule(function () { cb(null, 1); }); return true; }
    settings(settings, cb) {
      if (typeof settings !== 'object' || settings === null) throw invalidArg('settings', 'an object', settings);
      host.sessionSettings.call(this, settings);
      if (typeof cb === 'function') schedule(function () { cb(null); });
      return this;
    }
    goaway(code, lastStreamID, opaqueData) { this.closed = true; }
    ref() { return this; }
    unref() { return this; }
    setTimeout() { return this; }
  }

  function createServer(options, listener) {
    if (typeof options === 'function') { listener = options; options = {}; }
    return new Http2Server(options || {}, listener, false);
  }
  function createSecureServer(options, listener) {
    if (typeof options === 'function') { listener = options; options = {}; }
    return new Http2Server(options || {}, listener, true);
  }
  function connect(authority, options) {
    return new ClientHttp2Session(authority, options || {});
  }

  function dispatch(server, id, method, url, scheme, authority, rawHeaders, body, remoteAddress, remotePort, localAddress, localPort) {
    var req = new Http2ServerRequest(id, server, method, url, scheme, authority, rawHeaders, body, remoteAddress, remotePort, localAddress, localPort);
    var stream = null;
    var res = new Http2ServerResponse(id, req, null);
    stream = new Http2Stream(id, server, req, res);
    res.stream = stream;
    req.stream = stream;
    req.on('data', function (d) { stream.emit('data', d); });
    req.on('end', function () { stream.emit('end'); });
    schedule(function () { req._emitBody(); });
    if (server.listenerCount('stream') > 0) server.emit('stream', stream, req.headers);
    if (server.listenerCount('request') > 0) server.emit('request', req, res);
    else if (server.listenerCount('stream') === 0) res.destroy(new Error('No request or stream listener'));
  }

  var constants = {
    NGHTTP2_NO_ERROR: 0, NGHTTP2_PROTOCOL_ERROR: 1, NGHTTP2_INTERNAL_ERROR: 2,
    NGHTTP2_FLOW_CONTROL_ERROR: 3, NGHTTP2_SETTINGS_TIMEOUT: 4,
    NGHTTP2_STREAM_CLOSED: 5, NGHTTP2_FRAME_SIZE_ERROR: 6, NGHTTP2_REFUSED_STREAM: 7,
    NGHTTP2_CANCEL: 8, NGHTTP2_COMPRESSION_ERROR: 9, NGHTTP2_CONNECT_ERROR: 10,
    NGHTTP2_ENHANCE_YOUR_CALM: 11, NGHTTP2_INADEQUATE_SECURITY: 12,
    NGHTTP2_HTTP_1_1_REQUIRED: 13,
    HTTP2_HEADER_STATUS: ':status', HTTP2_HEADER_METHOD: ':method',
    HTTP2_HEADER_PATH: ':path', HTTP2_HEADER_AUTHORITY: ':authority',
    HTTP2_HEADER_SCHEME: ':scheme', HTTP2_HEADER_CONTENT_TYPE: 'content-type',
    HTTP2_HEADER_CONTENT_LENGTH: 'content-length', HTTP2_HEADER_USER_AGENT: 'user-agent'
  };

  return {
    Http2Server: Http2Server,
    Http2ServerRequest: Http2ServerRequest,
    Http2ServerResponse: Http2ServerResponse,
    Http2Stream: Http2Stream,
    ClientHttp2Session: ClientHttp2Session,
    ClientHttp2Stream: ClientHttp2Stream,
    createServer: createServer,
    createSecureServer: createSecureServer,
    connect: connect,
    constants: constants,
    dispatch: dispatch
  };
})`
