package web

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

// installFetch installs the undici-shaped fetch surface on top of GodeJS's
// own HTTP client. Everything is implemented in JavaScript so the client talks
// to modules/http (and modules/https) exactly the way user code would.
func installFetch(e *env.Env) {
	fn, err := e.JS.Eval(fetchPrelude, "[gode:web fetch]")
	if err != nil || !fn.Valid() || !fn.IsFunction() {
		return
	}
	_, _ = fn.Object().Call(js.Object{})
}

const fetchPrelude = `(function () {
  'use strict';

  function load(name) {
    if (typeof require === 'function') return require(name);
    if (typeof globalThis.require === 'function') return globalThis.require(name);
    throw new Error('require is unavailable');
  }

  function isBlob(value) {
    return typeof Blob === 'function' && value instanceof Blob;
  }

  function toBytes(value) {
    if (value instanceof Uint8Array) return value;
    if (typeof ArrayBuffer !== 'undefined' && value instanceof ArrayBuffer) {
      return new Uint8Array(value);
    }
    if (typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(value)) {
      return new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
    }
    return null;
  }

  function concat(chunks) {
    var length = 0;
    for (var i = 0; i < chunks.length; i++) length += chunks[i].byteLength;
    var out = new Uint8Array(length);
    var offset = 0;
    for (var j = 0; j < chunks.length; j++) {
      out.set(chunks[j], offset);
      offset += chunks[j].byteLength;
    }
    return out;
  }

  function abortError() {
    if (typeof DOMException === 'function') {
      return new DOMException('This operation was aborted', 'AbortError');
    }
    var error = new Error('This operation was aborted');
    error.name = 'AbortError';
    return error;
  }

  function fetchFailed(cause) {
    var error = new TypeError('fetch failed');
    error.cause = cause;
    return error;
  }

  // --- Headers ------------------------------------------------------------
  function normalizeName(name) {
    var text = String(name);
    if (!/^[!#$%&'*+\-.^_\x60|~0-9A-Za-z]+$/.test(text)) {
      throw new TypeError("Invalid header name: '" + text + "'");
    }
    return text.toLowerCase();
  }

  function normalizeValue(value) {
    var text = String(value).replace(/^[\t\n\r ]+|[\t\n\r ]+$/g, '');
    if (/[\u0000\r\n]/.test(text)) {
      throw new TypeError('Invalid header value');
    }
    return text;
  }

  function Headers(init) {
    if (!(this instanceof Headers)) return new Headers(init);
    this._entries = new Map();
    if (init === undefined || init === null) return;
    if (init instanceof Headers) {
      var self = this;
      init.forEach(function (value, name) { self.append(name, value); });
      return;
    }
    if (typeof init === 'object' && typeof init[Symbol.iterator] === 'function') {
      var pairs = Array.from(init);
      for (var i = 0; i < pairs.length; i++) {
        var pair = pairs[i];
        if (!pair || pair.length !== 2) {
          throw new TypeError('Headers init must be a sequence of [name, value] pairs');
        }
        this.append(pair[0], pair[1]);
      }
      return;
    }
    if (typeof init === 'object') {
      var keys = Object.keys(init);
      for (var j = 0; j < keys.length; j++) this.append(keys[j], init[keys[j]]);
      return;
    }
    throw new TypeError('Failed to construct Headers: invalid initializer');
  }

  Headers.prototype.append = function (name, value) {
    var key = normalizeName(name);
    var text = normalizeValue(value);
    var list = this._entries.get(key);
    if (list) list.push(text);
    else this._entries.set(key, [text]);
  };

  Headers.prototype.set = function (name, value) {
    this._entries.set(normalizeName(name), [normalizeValue(value)]);
  };

  Headers.prototype.get = function (name) {
    var list = this._entries.get(normalizeName(name));
    if (!list || list.length === 0) return null;
    return list.join(', ');
  };

  Headers.prototype.getSetCookie = function () {
    var list = this._entries.get('set-cookie');
    return list ? list.slice() : [];
  };

  Headers.prototype.has = function (name) {
    return this._entries.has(normalizeName(name));
  };

  Headers.prototype.delete = function (name) {
    this._entries.delete(normalizeName(name));
  };

  Headers.prototype.forEach = function (callback, thisArg) {
    var names = Array.from(this._entries.keys()).sort();
    for (var i = 0; i < names.length; i++) {
      var list = this._entries.get(names[i]);
      for (var j = 0; j < list.length; j++) {
        callback.call(thisArg, list[j], names[i], this);
      }
    }
  };

  function makeIterator(items) {
    var index = 0;
    var iterator = {
      next: function () {
        if (index >= items.length) return { value: undefined, done: true };
        return { value: items[index++], done: false };
      }
    };
    iterator[Symbol.iterator] = function () { return iterator; };
    return iterator;
  }

  Headers.prototype.entries = function () {
    var out = [];
    this.forEach(function (value, name) { out.push([name, value]); });
    return makeIterator(out);
  };
  Headers.prototype.keys = function () {
    var out = [];
    this.forEach(function (value, name) { out.push(name); });
    return makeIterator(out);
  };
  Headers.prototype.values = function () {
    var out = [];
    this.forEach(function (value) { out.push(value); });
    return makeIterator(out);
  };
  Headers.prototype[Symbol.iterator] = Headers.prototype.entries;
  Object.defineProperty(Headers.prototype, Symbol.toStringTag,
    { value: 'Headers', configurable: true });

  // --- FormData -----------------------------------------------------------
  function FormData() {
    if (!(this instanceof FormData)) return new FormData();
    this._entries = [];
  }

  FormData.prototype.append = function (name, value, filename) {
    var entry = {
      name: String(name),
      value: isBlob(value) ? value : String(value)
    };
    if (filename !== undefined) entry.filename = String(filename);
    this._entries.push(entry);
  };

  FormData.prototype.set = function (name, value, filename) {
    this.delete(name);
    this.append(name, value, filename);
  };

  FormData.prototype.delete = function (name) {
    var key = String(name);
    this._entries = this._entries.filter(function (entry) { return entry.name !== key; });
  };

  FormData.prototype.get = function (name) {
    var key = String(name);
    for (var i = 0; i < this._entries.length; i++) {
      if (this._entries[i].name === key) return this._entries[i].value;
    }
    return null;
  };

  FormData.prototype.getAll = function (name) {
    var key = String(name);
    var out = [];
    for (var i = 0; i < this._entries.length; i++) {
      if (this._entries[i].name === key) out.push(this._entries[i].value);
    }
    return out;
  };

  FormData.prototype.has = function (name) {
    var key = String(name);
    for (var i = 0; i < this._entries.length; i++) {
      if (this._entries[i].name === key) return true;
    }
    return false;
  };

  FormData.prototype.forEach = function (callback, thisArg) {
    for (var i = 0; i < this._entries.length; i++) {
      callback.call(thisArg, this._entries[i].value, this._entries[i].name, this);
    }
  };

  FormData.prototype.entries = function () {
    var out = [];
    for (var i = 0; i < this._entries.length; i++) {
      out.push([this._entries[i].name, this._entries[i].value]);
    }
    return makeIterator(out);
  };
  FormData.prototype.keys = function () {
    var out = [];
    for (var i = 0; i < this._entries.length; i++) out.push(this._entries[i].name);
    return makeIterator(out);
  };
  FormData.prototype.values = function () {
    var out = [];
    for (var i = 0; i < this._entries.length; i++) out.push(this._entries[i].value);
    return makeIterator(out);
  };
  FormData.prototype[Symbol.iterator] = FormData.prototype.entries;
  Object.defineProperty(FormData.prototype, Symbol.toStringTag,
    { value: 'FormData', configurable: true });

  function quotePart(value) {
    return String(value).replace(/\r/g, '%0D').replace(/\n/g, '%0A').replace(/"/g, '%22');
  }

  async function encodeForm(form) {
    var boundary = '----GodeFormBoundary' +
      (globalThis.crypto && typeof crypto.randomUUID === 'function'
        ? crypto.randomUUID() : String(Math.random()).slice(2));
    var parts = [];
    for (var i = 0; i < form._entries.length; i++) {
      var entry = form._entries[i];
      var head = '--' + boundary +
        '\r\nContent-Disposition: form-data; name="' + quotePart(entry.name) + '"';
      if (isBlob(entry.value)) {
        var filename = entry.filename === undefined ? 'blob' : entry.filename;
        head += '; filename="' + quotePart(filename) + '"\r\nContent-Type: ' +
          (entry.value.type || 'application/octet-stream') + '\r\n\r\n';
        parts.push(new TextEncoder().encode(head));
        parts.push(new Uint8Array(await entry.value.arrayBuffer()));
      } else {
        head += '\r\n\r\n' + entry.value;
        parts.push(new TextEncoder().encode(head));
      }
      parts.push(new TextEncoder().encode('\r\n'));
    }
    parts.push(new TextEncoder().encode('--' + boundary + '--\r\n'));
    return {
      bytes: concat(parts),
      contentType: 'multipart/form-data; boundary=' + boundary
    };
  }

  // --- body plumbing ------------------------------------------------------
  function readStream(stream) {
    var reader = stream.getReader();
    var chunks = [];
    return (async function () {
      for (;;) {
        var result = await reader.read();
        if (result.done) break;
        chunks.push(result.value instanceof Uint8Array
          ? result.value : new TextEncoder().encode(String(result.value)));
      }
      return concat(chunks);
    })();
  }

  async function readSource(source) {
    if (source === null) return new Uint8Array(0);
    if (source.kind === 'bytes') return source.bytes;
    if (source.kind === 'blob') return new Uint8Array(await source.blob.arrayBuffer());
    if (source.kind === 'form') return (await encodeForm(source.form)).bytes;
    if (source.kind === 'stream') return readStream(source.stream);
    return new Uint8Array(0);
  }

  function bodySource(value) {
    if (value === undefined || value === null) return null;
    if (typeof value === 'object' && typeof value.kind === 'string') return value;
    if (typeof value === 'string') {
      return {
        kind: 'bytes', bytes: new TextEncoder().encode(value),
        type: 'text/plain;charset=UTF-8'
      };
    }
    if (typeof URLSearchParams === 'function' && value instanceof URLSearchParams) {
      return {
        kind: 'bytes', bytes: new TextEncoder().encode(value.toString()),
        type: 'application/x-www-form-urlencoded;charset=UTF-8'
      };
    }
    if (isBlob(value)) return { kind: 'blob', blob: value };
    if (value instanceof FormData) return { kind: 'form', form: value };
    if (typeof ReadableStream === 'function' && value instanceof ReadableStream) {
      return { kind: 'stream', stream: value };
    }
    var bytes = toBytes(value);
    if (bytes !== null) return { kind: 'bytes', bytes: bytes, type: '' };
    throw new TypeError('Unsupported body type');
  }

  function streamFromSource(source) {
    if (source === null) return null;
    if (source.kind === 'stream') return source.stream;
    var started = false;
    return new ReadableStream({
      pull: function (controller) {
        if (started) return undefined;
        started = true;
        return readSource(source).then(function (bytes) {
          if (bytes.byteLength > 0) controller.enqueue(bytes);
          controller.close();
        }, function (error) { controller.error(error); });
      }
    });
  }

  function consumeBytes(object) {
    if (object.bodyUsed) {
      return Promise.reject(new TypeError('Body is unusable: Body has already been read'));
    }
    object.bodyUsed = true;
    return readSource(object._source);
  }

  function parseFormData(response) {
    return consumeBytes(response).then(function (bytes) {
      var contentType = response.headers.get('content-type') || '';
      var text = new TextDecoder().decode(bytes);
      var form = new FormData();
      var lower = contentType.toLowerCase();
      if (lower.indexOf('multipart/form-data') >= 0) {
        var match = /boundary=(?:"([^"]+)"|([^;]+))/i.exec(contentType);
        if (!match) throw new TypeError('Missing multipart boundary');
        var boundary = '--' + (match[1] || match[2]);
        var parts = text.split(boundary);
        for (var i = 1; i < parts.length; i++) {
          var part = parts[i];
          if (part.indexOf('--') === 0) break;
          if (part.indexOf('\r\n') === 0) part = part.slice(2);
          var split = part.indexOf('\r\n\r\n');
          if (split < 0) continue;
          var headers = part.slice(0, split);
          var value = part.slice(split + 4);
          if (value.slice(-2) === '\r\n') value = value.slice(0, -2);
          var nameMatch = /name="([^"]*)"/i.exec(headers);
          if (!nameMatch) continue;
          var filenameMatch = /filename="([^"]*)"/i.exec(headers);
          if (filenameMatch) {
            var typeMatch = /content-type:\s*([^\r\n]+)/i.exec(headers);
            var fileBytes = new TextEncoder().encode(value);
            var file = typeof File === 'function'
              ? new File([fileBytes], filenameMatch[1],
                { type: typeMatch ? typeMatch[1].trim() : '' })
              : new Blob([fileBytes], { type: typeMatch ? typeMatch[1].trim() : '' });
            form.append(nameMatch[1], file, filenameMatch[1]);
          } else {
            form.append(nameMatch[1], value);
          }
        }
        return form;
      }
      if (lower.indexOf('application/x-www-form-urlencoded') >= 0) {
        var params = new URLSearchParams(text);
        params.forEach(function (value, name) { form.append(name, value); });
        return form;
      }
      throw new TypeError('Unsupported content type for formData(): ' + contentType);
    });
  }

  function defineBody(proto) {
    proto.text = function () {
      return consumeBytes(this).then(function (bytes) {
        return new TextDecoder().decode(bytes);
      });
    };
    proto.json = function () {
      return this.text().then(function (text) { return JSON.parse(text); });
    };
    proto.arrayBuffer = function () {
      return consumeBytes(this).then(function (bytes) {
        return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
      });
    };
    proto.bytes = function () { return consumeBytes(this); };
    proto.blob = function () {
      var self = this;
      return consumeBytes(this).then(function (bytes) {
        return new Blob([bytes], { type: self.headers.get('content-type') || '' });
      });
    };
    proto.formData = function () { return parseFormData(this); };
    proto.clone = function () {
      if (this.bodyUsed) {
        throw new TypeError('Body is unusable: Body has already been read');
      }
      var copy = Object.create(Object.getPrototypeOf(this));
      copy._source = this._source;
      copy.body = this.body;
      copy.bodyUsed = false;
      if (this.body !== null && this.body !== undefined) {
        var pair = this.body.tee();
        this.body = pair[0];
        this._source = { kind: 'stream', stream: pair[0] };
        copy.body = pair[1];
        copy._source = { kind: 'stream', stream: pair[1] };
      }
      copy.headers = new Headers(this.headers);
      copy.url = this.url;
      copy.redirected = this.redirected;
      copy.type = this.type;
      copy.status = this.status;
      copy.statusText = this.statusText;
      copy.ok = this.ok;
      return copy;
    };
  }

  // --- Response -----------------------------------------------------------
  function Response(body, init) {
    if (!(this instanceof Response)) {
      throw new TypeError("Class constructor Response cannot be invoked without 'new'");
    }
    init = init || {};
    var status = init.status === undefined ? 200 : Number(init.status);
    if (status < 200 || status > 599) {
      throw new RangeError('init.status must be in the range of 200 to 599, inclusive');
    }
    this.status = status;
    this.statusText = init.statusText === undefined ? '' : String(init.statusText);
    this.headers = new Headers(init.headers);
    this.ok = status >= 200 && status < 300;
    this.redirected = false;
    this.type = 'default';
    this.url = '';
    this.bodyUsed = false;
    this._source = bodySource(body);
    this.body = streamFromSource(this._source);
  }

  Response.prototype[Symbol.toStringTag] = 'Response';

  Response.error = function () {
    var response = Object.create(Response.prototype);
    response.status = 0;
    response.statusText = '';
    response.headers = new Headers();
    response.ok = false;
    response.redirected = false;
    response.type = 'error';
    response.url = '';
    response.bodyUsed = false;
    response._source = null;
    response.body = null;
    return response;
  };

  Response.redirect = function (url, status) {
    if (status === undefined) status = 302;
    if ([301, 302, 303, 307, 308].indexOf(Number(status)) < 0) {
      throw new RangeError('Invalid redirect status code');
    }
    var headers = new Headers({ location: String(url) });
    var response = new Response(null, { status: Number(status), headers: headers });
    return response;
  };

  Response.json = function (data, init) {
    init = init || {};
    var headers = new Headers(init.headers);
    if (!headers.has('content-type')) {
      headers.set('content-type', 'application/json');
    }
    var body = JSON.stringify(data);
    var response = new Response(body, {
      status: init.status, statusText: init.statusText, headers: headers
    });
    response._source = { kind: 'bytes', bytes: new TextEncoder().encode(body), type: 'application/json' };
    return response;
  };

  defineBody(Response.prototype);

  // --- Request ------------------------------------------------------------
  var METHOD_TOKEN = /^[!#$%&'*+\-.^_\x60|~0-9A-Za-z]+$/;

  function Request(input, init) {
    if (!(this instanceof Request)) {
      throw new TypeError("Class constructor Request cannot be invoked without 'new'");
    }
    init = init || {};
    var base = input instanceof Request ? input : null;
    var target = base ? base.url : String(input);
    this.url = new URL(target).href;
    var method = init.method !== undefined
      ? String(init.method) : (base ? base.method : 'GET');
    this.method = method.toUpperCase();
    if (!METHOD_TOKEN.test(this.method)) {
      throw new TypeError("Invalid request method: '" + method + "'");
    }
    this.headers = new Headers(init.headers !== undefined ? init.headers
      : (base ? base.headers : undefined));
    this.redirect = init.redirect !== undefined ? String(init.redirect)
      : (base ? base.redirect : 'follow');
    this.signal = init.signal !== undefined ? init.signal
      : (base ? base.signal : null);
    this.credentials = 'same-origin';
    this.mode = 'cors';
    this.referrer = 'about:client';
    this.duplex = init.duplex;
    this.bodyUsed = false;
    var source = init.body !== undefined ? bodySource(init.body)
      : (base ? base._source : null);
    if (source !== null && (this.method === 'GET' || this.method === 'HEAD')) {
      throw new TypeError('Request with GET/HEAD method cannot have body.');
    }
    this._source = source;
    this.body = streamFromSource(source);
  }

  Request.prototype[Symbol.toStringTag] = 'Request';
  defineBody(Request.prototype);

  // --- fetch --------------------------------------------------------------
  function headersFromObject(source) {
    var headers = new Headers();
    if (!source) return headers;
    var keys = Object.keys(source);
    for (var i = 0; i < keys.length; i++) {
      var value = source[keys[i]];
      if (value === undefined) continue;
      if (Array.isArray(value)) {
        for (var j = 0; j < value.length; j++) headers.append(keys[i], value[j]);
      } else {
        headers.append(keys[i], value);
      }
    }
    return headers;
  }

  function isRedirect(status) {
    return status === 301 || status === 302 || status === 303 ||
      status === 307 || status === 308;
  }

  function fetch(input, init) {
    return new Promise(function (resolve, reject) {
      var request;
      try {
        request = new Request(input, init);
      } catch (error) {
        reject(error);
        return;
      }
      var signal = request.signal;
      if (signal && signal.aborted) {
        reject(abortError());
        return;
      }

      var redirectCount = 0;

      function send(current, url) {
        var http;
        try {
          http = load(url.protocol === 'https:' ? 'https' : 'http');
        } catch (error) {
          reject(error);
          return;
        }
        var bodyPromise;
        if (current._source === null) {
          bodyPromise = Promise.resolve(null);
        } else if (current._source.kind === 'form') {
          bodyPromise = encodeForm(current._source.form).then(function (encoded) {
            return { bytes: encoded.bytes, type: encoded.contentType };
          });
        } else {
          bodyPromise = readSource(current._source).then(function (bytes) {
            return { bytes: bytes, type: current._source.type || '' };
          });
        }
        bodyPromise
          .then(function (body) {
            var headers = {};
            current.headers.forEach(function (value, name) { headers[name] = value; });
            if (body && body.bytes.byteLength > 0) {
              headers['content-length'] = String(body.bytes.byteLength);
              if (!headers['content-type'] && body.type) {
                headers['content-type'] = body.type;
              }
            }
            var options = {
              protocol: url.protocol,
              hostname: url.hostname,
              port: url.port || undefined,
              path: url.pathname + url.search,
              method: current.method,
              headers: headers
            };
            var client;
            try {
              client = http.request(options);
            } catch (error) {
              reject(error);
              return;
            }
            var settled = false;
            function onAbort() {
              if (settled) return;
              settled = true;
              try { if (typeof client.destroy === 'function') client.destroy(); } catch (ignored) {}
              reject(abortError());
            }
            if (signal && typeof signal.addEventListener === 'function') {
              signal.addEventListener('abort', onAbort);
            }
            client.on('error', function (error) {
              if (settled) return;
              settled = true;
              if (signal && typeof signal.removeEventListener === 'function') {
                signal.removeEventListener('abort', onAbort);
              }
              reject(fetchFailed(error));
            });
            client.on('response', function (response) {
              if (settled) return;
              var location = response.headers && (response.headers.location ||
                response.headers.Location);
              if (current.redirect === 'follow' && location &&
                  isRedirect(response.statusCode) && redirectCount < 20) {
                redirectCount++;
                var nextUrl;
                try {
                  nextUrl = new URL(String(location), url.href);
                } catch (error) {
                  settled = true;
                  reject(fetchFailed(error));
                  return;
                }
                var nextMethod = current.method;
                if ((response.statusCode === 303 && nextMethod !== 'HEAD') ||
                    ((response.statusCode === 301 || response.statusCode === 302) &&
                     nextMethod === 'POST')) {
                  nextMethod = 'GET';
                }
                var nextInit = {
                  method: nextMethod,
                  headers: current.headers,
                  redirect: 'follow',
                  signal: signal
                };
                if (nextMethod !== 'GET' && nextMethod !== 'HEAD') {
                  nextInit.body = current._source;
                }
                if (typeof response.destroy === 'function') response.destroy();
                var nextRequest;
                try {
                  nextRequest = new Request(nextUrl.href, nextInit);
                } catch (error) {
                  settled = true;
                  reject(error);
                  return;
                }
                send(nextRequest, nextUrl);
                return;
              }
              settled = true;
              if (signal && typeof signal.removeEventListener === 'function') {
                signal.removeEventListener('abort', onAbort);
              }
              var status = response.statusCode;
              var headersOut = headersFromObject(response.headers);
              var result = new Response(null, {
                status: status,
                statusText: response.statusMessage || '',
                headers: headersOut
              });
              result.url = url.href;
              result.redirected = redirectCount > 0;
              var noBody = current.method === 'HEAD' ||
                status === 204 || status === 205 || status === 304;
              if (noBody) {
                result.body = null;
                result._source = null;
                if (typeof response.resume === 'function') response.resume();
                resolve(result);
                return;
              }
              var stream = new ReadableStream({
                start: function (controller) {
                  response.on('data', function (chunk) {
                    controller.enqueue(chunk instanceof Uint8Array
                      ? chunk : new Uint8Array(chunk));
                  });
                  response.on('end', function () { controller.close(); });
                  response.on('error', function (error) { controller.error(error); });
                }
              });
              result.body = stream;
              result._source = { kind: 'stream', stream: stream };
              resolve(result);
            });
            if (body && body.bytes.byteLength > 0) client.write(body.bytes);
            client.end();
          }, reject);
      }

      var startUrl;
      try {
        startUrl = new URL(request.url);
      } catch (error) {
        reject(error);
        return;
      }
      send(request, startUrl);
    });
  }

  globalThis.Headers = Headers;
  globalThis.Request = Request;
  globalThis.Response = Response;
  globalThis.FormData = FormData;
  globalThis.fetch = fetch;
})`
