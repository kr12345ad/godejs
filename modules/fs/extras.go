package fs

import (
	"godejs/internal/js"
)

// installExtras adds the Node 24 APIs that sit above the core descriptor and
// path calls: fs.Utf8Stream, fs.glob/globSync, fs.openAsBlob, the disposable
// mkdtemp helpers, fs.StatWatcher, the FileHandle stream helpers, and the
// Symbol.dispose/Symbol.asyncDispose hooks on Dir, FileHandle and Utf8Stream.
func (m *module) installExtras(exports js.Object) {
	deps := m.e.JS.Obj()
	if ev := m.e.Exports("events"); ev.Valid() {
		if v := ev.Get("EventEmitter"); v.Valid() {
			deps.Set("EventEmitter", v)
		}
	}
	factory, err := m.e.JS.Eval(extrasSource, "[fs:extras]")
	if err != nil {
		panic("godejs fs extras: " + err.Error())
	}
	if _, err := factory.Object().Call(js.Object{}, exports.V(), deps.V()); err != nil {
		panic("godejs fs extras: " + err.Error())
	}
}

// extrasSource is evaluated once from Install. It only uses the public fs
// surface installed before it, so the JS stays independent of the Go internals.
const extrasSource = `(function (fs, deps) {
  'use strict';
  var EventEmitter = deps.EventEmitter || function () {};
  var SymDispose = typeof Symbol === 'function' ? Symbol.dispose : undefined;
  var SymAsyncDispose = typeof Symbol === 'function' ? Symbol.asyncDispose : undefined;

  function define(obj, name, value) {
    if (!obj || name === undefined) return;
    if (Object.prototype.hasOwnProperty.call(obj, name)) return;
    Object.defineProperty(obj, name, {
      value: value, writable: true, configurable: true, enumerable: false
    });
  }

  function eventEmitterCtor() {
    return EventEmitter;
  }

  // ------------------------------------------------------------ Utf8Stream --
  function invalidState(msg) {
    var err = new Error(msg);
    err.code = 'ERR_INVALID_STATE';
    return err;
  }

  function Utf8Stream(options) {
    if (!(this instanceof Utf8Stream)) return new Utf8Stream(options);
    var EE = eventEmitterCtor();
    if (typeof EE === 'function') EE.call(this);
    options = options || {};
    this.file = options.destination !== undefined ? options.destination
      : (options.file !== undefined ? options.file : options.path);
    this.fd = options.fd === undefined ? null : options.fd;
    this.append = options.append !== false;
    this.mode = options.mode === undefined ? 438 : options.mode;
    this.fsync = options.fsync === true;
    this.sync = options.sync !== false;
    this.mkdir = options.mkdir === true;
    this.minLength = options.minLength === undefined ? 0 : options.minLength;
    this.maxLength = options.maxLength === undefined ? Infinity : options.maxLength;
    this.periodicFlush = options.periodicFlush === undefined ? 0 : options.periodicFlush;
    this.contentMode = options.contentMode === undefined ? 'utf8' : options.contentMode;
    this.writing = false;
    this.destroyed = false;
    this._buffer = [];
    this._length = 0;
    if (typeof this.file === 'number') {
      this.fd = this.file;
      this.file = null;
    }
    if (this.fd === null && this.file != null) this._open();
    if (this.periodicFlush > 0) {
      var self = this;
      this._timer = setInterval(function () { self.flush(); }, this.periodicFlush);
      if (this._timer && typeof this._timer.unref === 'function') this._timer.unref();
    }
  }
  Utf8Stream.prototype = Object.create(EventEmitter.prototype);
  Utf8Stream.prototype.constructor = Utf8Stream;

  Utf8Stream.prototype._open = function () {
    if (this.mkdir) {
      var idx = String(this.file).lastIndexOf('/');
      if (idx > 0) {
        try { fs.mkdirSync(String(this.file).slice(0, idx), { recursive: true }); } catch (e) {}
      }
    }
    this.fd = fs.openSync(this.file, this.append ? 'a' : 'w', this.mode);
  };

  Utf8Stream.prototype.write = function (data) {
    if (this.destroyed) throw invalidState('Utf8Stream is destroyed');
    var piece = typeof data === 'string' ? data : String(data);
    this._buffer.push(piece);
    this._length += piece.length;
    if (this._length >= this.maxLength) this.flushSync();
    else if (this._length >= this.minLength) this.flushSync();
    return true;
  };

  Utf8Stream.prototype._drain = function () {
    if (this._length === 0) return;
    var text = this._buffer.join('');
    this._buffer = [];
    this._length = 0;
    var bytes = Buffer.from(text, 'utf8');
    var offset = 0;
    while (offset < bytes.length) {
      offset += fs.writeSync(this.fd, bytes, offset, bytes.length - offset);
    }
    if (this.fsync) fs.fsyncSync(this.fd);
  };

  Utf8Stream.prototype.flushSync = function () {
    if (this.destroyed) throw invalidState('Utf8Stream is destroyed');
    this.writing = true;
    try { this._drain(); } finally { this.writing = false; }
    return true;
  };

  Utf8Stream.prototype.flush = function (cb) {
    var err = null;
    try { this.flushSync(); } catch (e) { err = e; }
    if (typeof cb === 'function') {
      var self = this;
      setTimeout(function () { cb.call(self, err); }, 0);
      return undefined;
    }
    return err ? Promise.reject(err) : Promise.resolve(true);
  };

  Utf8Stream.prototype.end = function (cb) {
    var err = null;
    try {
      if (!this.destroyed) {
        this.flushSync();
        if (this.fd !== null && typeof this.fd === 'number') fs.closeSync(this.fd);
        this.destroyed = true;
        this.fd = null;
        if (this._timer) clearInterval(this._timer);
      }
    } catch (e) { err = e; }
    if (typeof cb === 'function') { var self = this; setTimeout(function () { cb.call(self, err); }, 0); return undefined; }
    return err ? Promise.reject(err) : Promise.resolve(true);
  };

  Utf8Stream.prototype.destroy = function () {
    try {
      if (this.fd !== null && typeof this.fd === 'number') fs.closeSync(this.fd);
    } catch (e) {}
    this.fd = null;
    this.destroyed = true;
    if (this._timer) clearInterval(this._timer);
    return this;
  };

  Utf8Stream.prototype.reopen = function () {
    if (this.destroyed) throw invalidState('Utf8Stream is destroyed');
    if (this.fd !== null && typeof this.fd === 'number') fs.closeSync(this.fd);
    this.fd = null;
    this._open();
    return this;
  };

  if (SymDispose) define(Utf8Stream.prototype, SymDispose, function () { this.destroy(); });
  define(fs, 'Utf8Stream', Utf8Stream);

  // ------------------------------------------------------------------ glob --
  function escapeRegExp(ch) { return '\\.^$+{}()|'.indexOf(ch) >= 0 ? '\\' + ch : ch; }

  function patternToRegExp(pattern) {
    var out = '^';
    var i = 0;
    while (i < pattern.length) {
      var ch = pattern.charAt(i);
      if (ch === '*') {
        if (pattern.charAt(i + 1) === '*') {
          i += 2;
          if (pattern.charAt(i) === '/') { out += '(?:.*/)?'; i += 1; }
          else out += '.*';
          continue;
        }
        out += '[^/]*';
      } else if (ch === '?') {
        out += '[^/]';
      } else if (ch === '[') {
        var end = pattern.indexOf(']', i + 1);
        if (end > i) { out += pattern.slice(i, end + 1); i = end + 1; continue; }
        out += '\\[';
      } else {
        out += escapeRegExp(ch);
      }
      i += 1;
    }
    return new RegExp(out + '$');
  }

  function globBase(pattern) {
    var idx = pattern.search(/[*?[]/);
    if (idx < 0) {
      var slash = pattern.lastIndexOf('/');
      return slash < 0 ? '.' : pattern.slice(0, slash + 1);
    }
    var slash = pattern.slice(0, idx).lastIndexOf('/');
    return slash < 0 ? '.' : pattern.slice(0, slash + 1);
  }

  function walkTree(root, entries) {
    var stack = [root === '' ? '.' : root];
    while (stack.length) {
      var dir = stack.pop();
      var list;
      try { list = fs.readdirSync(dir, { withFileTypes: true }); } catch (e) { continue; }
      for (var i = 0; i < list.length; i++) {
        var ent = list[i];
        var full = dir === '.' ? ent.name : String(dir).replace(/\/$/, '') + '/' + ent.name;
        entries.push({ path: full, dirent: ent });
        if (ent.isDirectory && ent.isDirectory()) stack.push(full);
      }
    }
  }

  function globEntries(pattern, options) {
    options = options || {};
    var cwd = options.cwd === undefined ? process.cwd() : String(options.cwd);
    var absolute = String(pattern).charAt(0) === '/';
    var pat = absolute ? String(pattern).slice(1) : String(pattern);
    var root = absolute ? '/' : cwd;
    var entries = [];
    walkTree(root, entries);
    var re = patternToRegExp(pat);
    var excludes = options.exclude === undefined ? []
      : (Array.isArray(options.exclude) ? options.exclude : [options.exclude]);
    var excludeRes = excludes.map(function (p) { return patternToRegExp(String(p)); });
    var out = [];
    for (var i = 0; i < entries.length; i++) {
      var full = entries[i].path;
      var rel = absolute ? full.replace(/^\//, '') : (full.indexOf(cwd + '/') === 0 ? full.slice(cwd.length + 1) : full);
      if (!re.test(rel)) continue;
      var excluded = false;
      for (var j = 0; j < excludeRes.length; j++) {
        if (excludeRes[j].test(rel)) { excluded = true; break; }
      }
      if (excluded) continue;
      if (options.withFileTypes) {
        try { if (entries[i].dirent && entries[i].dirent.parentPath === undefined && entries[i].dirent.path) entries[i].dirent.parentPath = entries[i].dirent.path; } catch (e) {}
        out.push(entries[i].dirent);
      } else {
        out.push(absolute ? full : rel);
      }
    }
    out.sort();
    return out;
  }

  function globSync(pattern, options) {
    return globEntries(pattern, options);
  }

  function makeAsyncIterator(list) {
    var i = 0;
    var it = {
      next: function () {
        return Promise.resolve(i < list.length ? { value: list[i++], done: false } : { value: undefined, done: true });
      },
      return: function () { return Promise.resolve({ value: undefined, done: true }); }
    };
    if (typeof Symbol === 'function' && Symbol.asyncIterator) it[Symbol.asyncIterator] = function () { return this; };
    return it;
  }

  function glob(pattern, options) {
    return makeAsyncIterator(globEntries(pattern, options));
  }

  define(fs, 'globSync', globSync);
  define(fs, 'glob', glob);
  if (fs.promises) define(fs.promises, 'glob', glob);

  // ----------------------------------------------------------- openAsBlob --
  function blobConstructor() {
    if (typeof Blob === 'function') return Blob;
    try {
      if (typeof require === 'function') {
        var buf = require('buffer');
        if (buf && typeof buf.Blob === 'function') return buf.Blob;
      }
    } catch (e) {}
    return null;
  }

  define(fs, 'openAsBlob', function (path, options) {
    return new Promise(function (resolve, reject) {
      var data, Ctor;
      try {
        data = fs.readFileSync(path);
        Ctor = blobConstructor();
        if (!Ctor) {
          var err = new Error('Blob is not available in this runtime');
          err.code = 'ERR_METHOD_NOT_IMPLEMENTED';
          throw err;
        }
      } catch (e) { reject(e); return; }
      var type = options && options.type ? { type: options.type } : undefined;
      resolve(new Ctor([data], type));
    });
  });

  // ------------------------------------------------- disposable temp dirs --
  function disposableDir(dir) {
    var handle = {
      path: dir,
      remove: function () { fs.rmSync(dir, { recursive: true, force: true }); }
    };
    if (SymDispose) define(handle, SymDispose, function () { handle.remove(); });
    if (SymAsyncDispose) define(handle, SymAsyncDispose, function () {
      return Promise.resolve().then(function () { handle.remove(); });
    });
    return handle;
  }

  define(fs, 'mkdtempDisposableSync', function (prefix, options) {
    return disposableDir(fs.mkdtempSync(prefix, options));
  });
  if (fs.promises) define(fs.promises, 'mkdtempDisposable', function (prefix, options) {
    return fs.promises.mkdtemp(prefix, options).then(disposableDir);
  });

  // -------------------------------------------------- lchmod and lutimes ---
  var O_SYMLINK = fs.constants ? fs.constants.O_SYMLINK : undefined;
  function notImplemented(name) {
    var err = new Error('The ' + name + ' method is not implemented on this platform');
    err.code = 'ERR_METHOD_NOT_IMPLEMENTED';
    return err;
  }
  if (O_SYMLINK !== undefined) {
    define(fs, 'lchmod', function (path, mode, cb) { return fs.chmod(path, mode, cb); });
    define(fs, 'lchmodSync', function (path, mode) { return fs.chmodSync(path, mode); });
  } else {
    // Linux has no O_SYMLINK, so Node does not export lchmod at all; tests
    // guard on 'if (fs.lchmod)' and skip it.
    delete fs.lchmod;
    delete fs.lchmodSync;
    if (fs.promises) delete fs.promises.lchmod;
  }
  // lutimes follows the link (Linux utimensat without AT_SYMLINK_NOFOLLOW is
  // what this runtime can reach); the signature and option handling match.
  define(fs, 'lutimes', function (path, atime, mtime, cb) { return fs.utimes(path, atime, mtime, cb); });
  define(fs, 'lutimesSync', function (path, atime, mtime) { return fs.utimesSync(path, atime, mtime); });
  if (fs.promises) define(fs.promises, 'lutimes', function (path, atime, mtime) {
    return fs.promises.utimes(path, atime, mtime);
  });

  // -------------------------------------------------------- StatWatcher ----
  function StatWatcher() {
    var EE = eventEmitterCtor();
    if (typeof EE === 'function') EE.call(this);
  }
  StatWatcher.prototype = Object.create(EventEmitter.prototype);
  StatWatcher.prototype.constructor = StatWatcher;
  StatWatcher.prototype.ref = function () { return this; };
  StatWatcher.prototype.unref = function () { return this; };

  var nativeWatchFile = fs.watchFile;
  if (typeof nativeWatchFile === 'function') {
    fs.watchFile = function (filename, options, listener) {
      var watcher = nativeWatchFile.apply(fs, arguments);
      if (watcher && !(watcher instanceof StatWatcher)) {
        try { Object.setPrototypeOf(watcher, StatWatcher.prototype); } catch (e) {}
      }
      return watcher;
    };
  }
  define(fs, 'StatWatcher', StatWatcher);

  // ------------------------------------------------- FSWatcher ref/unref ---
  if (typeof fs.FSWatcher === 'function') {
    var fwp = fs.FSWatcher.prototype;
    define(fwp, 'ref', function () {
      if (this._timer && typeof this._timer.ref === 'function') this._timer.ref();
      return this;
    });
    define(fwp, 'unref', function () {
      if (this._timer && typeof this._timer.unref === 'function') this._timer.unref();
      return this;
    });
    define(fwp, 'close', function () {
      if (this._closed) return;
      this._closed = true;
      if (this._timer && typeof clearInterval === 'function') clearInterval(this._timer);
      var self = this;
      setTimeout(function () { self.emit('close'); }, 0);
    });
  }

  // --------------------------------------------------------- Dir disposal --
  if (typeof fs.Dir === 'function') {
    define(fs.Dir.prototype, SymDispose, function () { this.closeSync(); });
    define(fs.Dir.prototype, SymAsyncDispose, function () { return this.close(); });
  }

  // ----------------------------------------------- FileHandle extras ------
  if (fs.promises && typeof fs.promises.FileHandle === 'function') {
    var fhp = fs.promises.FileHandle.prototype;

    define(fhp, 'createReadStream', function (options) {
      var opts = {};
      if (options) for (var k in options) opts[k] = options[k];
      opts.fd = this.fd;
      opts.autoClose = false;
      return fs.createReadStream('', opts);
    });

    define(fhp, 'createWriteStream', function (options) {
      var opts = {};
      if (options) for (var k in options) opts[k] = options[k];
      opts.fd = this.fd;
      opts.autoClose = false;
      return fs.createWriteStream('', opts);
    });

    define(fhp, 'readLines', function (options) {
      var self = this;
      var encoding = options && options.encoding ? options.encoding : 'utf8';
      return (async function* () {
        var text = await self.readFile(encoding);
        var lines = String(text).split(/\r?\n/);
        if (lines.length && lines[lines.length - 1] === '') lines.pop();
        for (var i = 0; i < lines.length; i++) yield lines[i];
      })();
    });

    define(fhp, 'readableWebStream', function () {
      var self = this;
      var pos = 0;
      if (typeof ReadableStream !== 'function') {
        return {
          getReader: function () {
            return {
              read: function () { return Promise.resolve({ value: undefined, done: true }); },
              cancel: function () { return Promise.resolve(); }
            };
          }
        };
      }
      return new ReadableStream({
        pull: function (controller) {
          var buf = Buffer.allocUnsafe(65536);
          return self.read(buf, 0, 65536, pos).then(function (res) {
            if (!res || res.bytesRead === 0) { controller.close(); return; }
            pos += res.bytesRead;
            controller.enqueue(new Uint8Array(buf.buffer, buf.byteOffset, res.bytesRead));
          }, function (err) { controller.error(err); });
        }
      });
    });

    define(fhp, 'writer', function () {
      var self = this;
      if (typeof WritableStream !== 'function') {
        return {
          write: function (chunk) { return self.write(chunk); },
          close: function () { return Promise.resolve(); }
        };
      }
      return new WritableStream({
        write: function (chunk) { return self.write(chunk); }
      });
    });

    define(fhp, 'pull', function (buffer, offset, length, position) {
      return this.read(buffer, offset === undefined ? 0 : offset,
        length === undefined ? buffer.byteLength - (offset || 0) : length, position);
    });

    define(fhp, 'pullSync', function (buffer, offset, length, position) {
      return fs.readSync(this.fd, buffer, offset === undefined ? 0 : offset,
        length === undefined ? buffer.byteLength - (offset || 0) : length, position);
    });

    define(fhp, SymAsyncDispose, function () { return this.close(); });
    define(fhp, SymDispose, function () { this.close(); });
  }

  // ------------------------------------------------------------- fs.native --
  if (!fs.native) {
    define(fs, 'native', {
      statfs: fs.statfs,
      statfsSync: fs.statfsSync,
      open: fs.open,
      close: fs.close,
      read: fs.read,
      write: fs.write,
      fstat: fs.fstat
    });
  }

  return fs;
})`
