package web

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

// installWebSocket installs the WhatWG WebSocket client. The handshake and
// RFC 6455 framing live in JavaScript and run over modules/net or modules/tls,
// so no separate socket stack is duplicated in Go.
func installWebSocket(e *env.Env) {
	fn, err := e.JS.Eval(websocketPrelude, "[gode:web websocket]")
	if err != nil || !fn.Valid() || !fn.IsFunction() {
		return
	}
	// Load transports on first connection instead of paying for net/tls at
	// startup; the prelude calls these thunks only from the constructor.
	deps := e.JS.Obj()
	for _, name := range []string{"net", "tls", "crypto"} {
		module := name
		deps.SetFunc(module, func(this js.Object, args []js.Value) (js.Value, error) {
			return e.Exports(module).V(), nil
		})
	}
	_, _ = fn.Object().Call(js.Object{}, deps.V())
}

const websocketPrelude = `(function (deps) {
  'use strict';

  function load(name) {
    if (deps && typeof deps[name] === 'function') return deps[name]();
    if (typeof require === 'function') return require(name);
    if (typeof globalThis.require === 'function') return globalThis.require(name);
    throw new Error('require is unavailable');
  }

  var GUID = '258EAFA5-E914-47DA-95CA-C5AB0DC85B11';
  var CONNECTING = 0;
  var OPEN = 1;
  var CLOSING = 2;
  var CLOSED = 3;

  var BaseEventTarget = typeof EventTarget === 'function'
    ? EventTarget
    : class { addEventListener() {} removeEventListener() {} dispatchEvent() {} };

  function randomBytes(count) {
    var out = new Uint8Array(count);
    if (globalThis.crypto && typeof crypto.getRandomValues === 'function') {
      crypto.getRandomValues(out);
    } else {
      for (var i = 0; i < count; i++) out[i] = Math.floor(Math.random() * 256);
    }
    return out;
  }

  function base64(bytes) {
    var text = '';
    for (var i = 0; i < bytes.length; i++) text += String.fromCharCode(bytes[i]);
    return btoa(text);
  }

  function utf8(text) { return new TextEncoder().encode(text); }
  function fromUtf8(bytes) { return new TextDecoder().decode(bytes); }

  function concat(first, second) {
    if (first.length === 0) return second;
    var out = new Uint8Array(first.length + second.length);
    out.set(first, 0);
    out.set(second, first.length);
    return out;
  }

  // Timers may be absent in embedders and unit-test environments; fall back to
  // running immediately (connect errors) or not scheduling (close watchdog).
  function later(callback, delay) {
    if (typeof setTimeout === 'function') return setTimeout(callback, delay);
    return null;
  }

  function cancelLater(handle) {
    if (handle !== null && handle !== undefined && typeof clearTimeout === 'function') {
      clearTimeout(handle);
    }
  }

  function frame(opcode, payload) {
    var length = payload.length;
    var headerLength = length < 126 ? 2 : (length < 65536 ? 4 : 10);
    var out = new Uint8Array(headerLength + 4 + length);
    out[0] = 0x80 | opcode;
    if (length < 126) {
      out[1] = 0x80 | length;
    } else if (length < 65536) {
      out[1] = 0x80 | 126;
      out[2] = (length >> 8) & 255;
      out[3] = length & 255;
    } else {
      out[1] = 0x80 | 127;
      out[6] = (length >>> 24) & 255;
      out[7] = (length >>> 16) & 255;
      out[8] = (length >>> 8) & 255;
      out[9] = length & 255;
    }
    var mask = randomBytes(4);
    out.set(mask, headerLength);
    var offset = headerLength + 4;
    for (var i = 0; i < length; i++) {
      out[offset + i] = payload[i] ^ mask[i & 3];
    }
    return out;
  }

  function makeEvent(type, props) {
    var event;
    if (typeof Event === 'function') {
      event = new Event(type);
    } else {
      event = { type: type };
    }
    if (props) {
      var keys = Object.keys(props);
      for (var i = 0; i < keys.length; i++) {
        try {
          Object.defineProperty(event, keys[i], {
            value: props[keys[i]], enumerable: true, configurable: true
          });
        } catch (ignored) {}
      }
    }
    return event;
  }

  function messageEvent(data, origin) {
    if (typeof MessageEvent === 'function') {
      return new MessageEvent('message', { data: data, origin: origin });
    }
    return makeEvent('message', { data: data, origin: origin });
  }

  // GodeJS's URL parser accepts the special web schemes; parse ws/wss here so
  // the client owns the authority rules and does not depend on that surface.
  function parseWebSocketUrl(value) {
    var match = /^(ws|wss):\/\/([^\/?#]+)([^?#]*)(\?[^#]*)?/i.exec(String(value));
    if (!match) throw new DOMException('Invalid URL', 'SyntaxError');
    var scheme = match[1].toLowerCase();
    var authority = match[2];
    if (authority.indexOf('@') >= 0) {
      throw new DOMException('Invalid URL', 'SyntaxError');
    }
    var hostname = authority;
    var portText = '';
    if (authority.charAt(0) === '[') {
      var close = authority.indexOf(']');
      if (close < 0) throw new DOMException('Invalid URL', 'SyntaxError');
      hostname = authority.slice(0, close + 1);
      var tail = authority.slice(close + 1);
      if (tail) {
        if (tail.charAt(0) !== ':') throw new DOMException('Invalid URL', 'SyntaxError');
        portText = tail.slice(1);
      }
    } else if (authority.lastIndexOf(':') >= 0) {
      var colon = authority.lastIndexOf(':');
      hostname = authority.slice(0, colon);
      portText = authority.slice(colon + 1);
    }
    if (!hostname) throw new DOMException('Invalid URL', 'SyntaxError');
    if (portText !== '' && !/^\d+$/.test(portText)) {
      throw new DOMException('Invalid URL', 'SyntaxError');
    }
    var secure = scheme === 'wss';
    var defaultPort = secure ? 443 : 80;
    var port = portText === '' ? defaultPort : Number(portText);
    if (port < 1 || port > 65535) throw new DOMException('Invalid URL', 'SyntaxError');
    var path = match[3] || '/';
    if (path.charAt(0) !== '/') path = '/' + path;
    var search = match[4] || '';
    var host = hostname + (port === defaultPort ? '' : ':' + port);
    return {
      protocol: scheme + ':',
      hostname: hostname,
      host: host,
      port: port,
      secure: secure,
      pathname: path,
      search: search,
      href: scheme + '://' + host + path + search
    };
  }

  class WebSocket extends BaseEventTarget {
    constructor(url, protocols) {
      super();
      var parsed = parseWebSocketUrl(url);
      var list = [];
      if (protocols !== undefined) {
        if (typeof protocols === 'string') {
          list = [protocols];
        } else if (protocols !== null && typeof protocols[Symbol.iterator] === 'function') {
          list = Array.from(protocols);
        } else {
          throw new TypeError(
            'The protocols argument must be a string or a sequence of strings');
        }
      }
      var seen = Object.create(null);
      for (var i = 0; i < list.length; i++) {
        var protocol = String(list[i]);
        if (!/^[!#$%&'*+\-.^_\x60|~0-9A-Za-z]+$/.test(protocol)) {
          throw new DOMException('Invalid subprotocol', 'SyntaxError');
        }
        if (seen[protocol]) {
          throw new DOMException('Duplicate subprotocol', 'SyntaxError');
        }
        seen[protocol] = true;
        list[i] = protocol;
      }
      this.url = parsed.href;
      this.readyState = CONNECTING;
      this.protocol = '';
      this.extensions = '';
      this.binaryType = 'arraybuffer';
      this.bufferedAmount = 0;
      this._protocols = list;
      this._socket = null;
      this._buffer = new Uint8Array(0);
      this._handshake = new Uint8Array(0);
      this._handshakeDone = false;
      this._fragments = [];
      this._fragmentOpcode = 0;
      this._closeSent = false;
      this._closeReceived = false;
      this._closed = false;
      this._closeTimer = null;
      this._origin = parsed.protocol + '//' + parsed.host;
      this._key = base64(randomBytes(16));

      var secure = parsed.secure;
      var transport = load(secure ? 'tls' : 'net');
      var port = parsed.port;
      var connectHost = parsed.hostname.charAt(0) === '['
        ? parsed.hostname.slice(1, -1) : parsed.hostname;
      var options = { host: connectHost, port: port };
      if (secure) options.servername = connectHost;
      var self = this;
      var socket;
      try {
        socket = transport.connect(options);
      } catch (error) {
        var scheduled = later(function () { self._fail(1006, '', false); }, 0);
        if (scheduled === null) self._fail(1006, '', false);
        return;
      }
      this._socket = socket;
      var started = false;
      function start() {
        if (started) return;
        started = true;
        self._handshakeRequest(parsed);
      }
      socket.on('connect', start);
      if (secure) socket.on('secureConnect', start);
      socket.on('data', function (chunk) { self._onData(chunk); });
      socket.on('error', function () { self._fail(1006, '', false); });
      socket.on('close', function () {
        if (!self._closed) {
          self._finishClose(self._closeReceived ? 1000 : 1006, '', self._closeReceived);
        }
      });
    }

    _handshakeRequest(parsed) {
      var lines = [
        'GET ' + parsed.pathname + parsed.search + ' HTTP/1.1',
        'Host: ' + parsed.host,
        'Upgrade: websocket',
        'Connection: Upgrade',
        'Sec-WebSocket-Key: ' + this._key,
        'Sec-WebSocket-Version: 13'
      ];
      if (this._protocols.length > 0) {
        lines.push('Sec-WebSocket-Protocol: ' + this._protocols.join(', '));
      }
      this._socket.write(lines.join('\r\n') + '\r\n\r\n');
    }

    _onData(chunk) {
      var bytes = chunk instanceof Uint8Array
        ? chunk : new Uint8Array(chunk);
      if (!this._handshakeDone) {
        this._handshake = concat(this._handshake, bytes);
        var text = '';
        for (var i = 0; i < this._handshake.length; i++) {
          text += String.fromCharCode(this._handshake[i]);
        }
        var end = text.indexOf('\r\n\r\n');
        if (end < 0) return;
        var head = text.slice(0, end);
        var rest = this._handshake.slice(end + 4);
        this._handshake = new Uint8Array(0);
        if (!this._validateHandshake(head)) return;
        this._handshakeDone = true;
        this.readyState = OPEN;
        this._dispatch('open', makeEvent('open'));
        if (rest.length > 0) this._onData(rest);
        return;
      }
      this._buffer = concat(this._buffer, bytes);
      this._parseFrames();
    }

    _validateHandshake(head) {
      var lines = head.split('\r\n');
      if ((lines[0] || '').indexOf(' 101 ') < 0) {
        this._fail(1006, '', false);
        return false;
      }
      var headers = Object.create(null);
      for (var i = 1; i < lines.length; i++) {
        var at = lines[i].indexOf(':');
        if (at < 0) continue;
        headers[lines[i].slice(0, at).trim().toLowerCase()] =
          lines[i].slice(at + 1).trim();
      }
      var expected = load('crypto').createHash('sha1')
        .update(this._key + GUID).digest('base64');
      if (headers['sec-websocket-accept'] !== expected) {
        this._fail(1006, '', false);
        return false;
      }
      if (headers['sec-websocket-protocol']) {
        this.protocol = headers['sec-websocket-protocol'];
      }
      return true;
    }

    _parseFrames() {
      while (this._buffer.length >= 2) {
        var first = this._buffer[0];
        var second = this._buffer[1];
        var fin = (first & 0x80) !== 0;
        var opcode = first & 0x0f;
        var masked = (second & 0x80) !== 0;
        var length = second & 0x7f;
        var offset = 2;
        if (length === 126) {
          if (this._buffer.length < 4) return;
          length = (this._buffer[2] << 8) | this._buffer[3];
          offset = 4;
        } else if (length === 127) {
          if (this._buffer.length < 10) return;
          length = ((this._buffer[6] << 24) >>> 0) +
            (this._buffer[7] << 16) + (this._buffer[8] << 8) + this._buffer[9];
          offset = 10;
        }
        var maskOffset = offset;
        if (masked) offset += 4;
        if (this._buffer.length < offset + length) return;
        var payload = this._buffer.slice(offset, offset + length);
        if (masked) {
          for (var i = 0; i < payload.length; i++) {
            payload[i] = payload[i] ^ this._buffer[maskOffset + (i & 3)];
          }
        }
        this._buffer = this._buffer.slice(offset + length);
        this._handleFrame(fin, opcode, payload);
      }
    }

    _handleFrame(fin, opcode, payload) {
      if (opcode === 0x8) {
        var code = payload.length >= 2 ? (payload[0] << 8) | payload[1] : 1005;
        var reason = payload.length > 2 ? fromUtf8(payload.slice(2)) : '';
        this._closeReceived = true;
        if (!this._closeSent) this._sendClose(code === 1005 ? 1000 : code, '');
        this._finishClose(code, reason, true);
        return;
      }
      if (opcode === 0x9) {
        if (this.readyState === OPEN) this._socket.write(frame(0xA, payload));
        return;
      }
      if (opcode === 0xA) return;
      if (opcode === 0x0) {
        this._fragments.push(payload);
        if (fin) {
          var combined = new Uint8Array(0);
          for (var i = 0; i < this._fragments.length; i++) {
            combined = concat(combined, this._fragments[i]);
          }
          this._fragments = [];
          this._deliver(this._fragmentOpcode, combined);
        }
        return;
      }
      if (opcode === 0x1 || opcode === 0x2) {
        if (!fin) {
          this._fragments = [payload];
          this._fragmentOpcode = opcode;
          return;
        }
        this._deliver(opcode, payload);
      }
    }

    _deliver(opcode, payload) {
      var data;
      if (opcode === 0x1) {
        data = fromUtf8(payload);
      } else {
        data = payload.buffer.slice(payload.byteOffset, payload.byteOffset + payload.byteLength);
        if (this.binaryType === 'blob' && typeof Blob === 'function') {
          data = new Blob([payload]);
        }
      }
      this._dispatch('message', messageEvent(data, this._origin));
    }

    _sendClose(code, reason) {
      if (this._closeSent) return;
      this._closeSent = true;
      var payload = new Uint8Array(0);
      if (code !== undefined && code !== 1005) {
        var reasonBytes = utf8(reason || '');
        payload = new Uint8Array(2 + reasonBytes.length);
        payload[0] = (code >> 8) & 255;
        payload[1] = code & 255;
        payload.set(reasonBytes, 2);
      }
      try { this._socket.write(frame(0x8, payload)); } catch (ignored) {}
    }

    _finishClose(code, reason, clean) {
      if (this._closed) return;
      this._closed = true;
      this.readyState = CLOSED;
      if (this._closeTimer !== null) {
        cancelLater(this._closeTimer);
        this._closeTimer = null;
      }
      try { if (this._socket) this._socket.end(); } catch (ignored) {}
      this._dispatch('close', makeEvent('close', {
        code: code, reason: reason, wasClean: clean
      }));
    }

    _fail(code, reason, clean) {
      if (this._closed) return;
      this._dispatch('error', makeEvent('error'));
      this._finishClose(code, reason, clean);
    }

    _dispatch(type, event) {
      var handler = this['on' + type];
      if (typeof handler === 'function') {
        try { handler.call(this, event);        } catch (error) {
          later(function () { throw error; }, 0);
        }
      }
      this.dispatchEvent(event);
    }

    send(data) {
      if (this.readyState !== OPEN) {
        throw new DOMException('WebSocket is not open', 'InvalidStateError');
      }
      var opcode;
      var payload;
      if (typeof data === 'string') {
        opcode = 0x1;
        payload = utf8(data);
      } else if (typeof ArrayBuffer !== 'undefined' && data instanceof ArrayBuffer) {
        opcode = 0x2;
        payload = new Uint8Array(data);
      } else if (typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(data)) {
        opcode = 0x2;
        payload = new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
      } else if (typeof Blob === 'function' && data instanceof Blob) {
        throw new TypeError('Blob send is not supported; use arrayBuffer()');
      } else {
        throw new TypeError('WebSocket send expects a string, ArrayBuffer or view');
      }
      this._socket.write(frame(opcode, payload));
    }

    close(code, reason) {
      if (this.readyState === CLOSED || this.readyState === CLOSING) return;
      if (code !== undefined && code !== 1000 && (code < 3000 || code > 4999)) {
        throw new DOMException('Invalid close code', 'InvalidAccessError');
      }
      var text = reason === undefined ? '' : String(reason);
      if (utf8(text).length > 123) {
        throw new DOMException('Close reason must not exceed 123 bytes', 'SyntaxError');
      }
      this.readyState = CLOSING;
      this._sendClose(code === undefined ? 1000 : code, text);
      var self = this;
      this._closeTimer = later(function () {
        self._closeTimer = null;
        if (self._closed) return;
        try { if (self._socket) self._socket.destroy(); } catch (ignored) {}
        self._finishClose(1006, '', false);
      }, 1000);
    }
  }

  ['open', 'message', 'error', 'close'].forEach(function (type) {
    Object.defineProperty(WebSocket.prototype, 'on' + type, {
      get: function () { return this['_on' + type] || null; },
      set: function (value) { this['_on' + type] = value; },
      configurable: true
    });
  });

  WebSocket.CONNECTING = 0;
  WebSocket.OPEN = 1;
  WebSocket.CLOSING = 2;
  WebSocket.CLOSED = 3;
  WebSocket.prototype.CONNECTING = 0;
  WebSocket.prototype.OPEN = 1;
  WebSocket.prototype.CLOSING = 2;
  WebSocket.prototype.CLOSED = 3;

  Object.defineProperty(WebSocket.prototype, Symbol.toStringTag,
    { value: 'WebSocket', configurable: true });

  globalThis.WebSocket = WebSocket;
})`
