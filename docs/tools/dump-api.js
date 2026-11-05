// Dump every own property of every builtin module and every host global,
// including class statics and prototype members, as JSON. Used by
// docs/tools/gen_api.py.
//
//   ./gode docs/tools/dump-api.js > docs/tools/api.json
'use strict';

const mods = [
  'assert', 'async_hooks', 'buffer', 'child_process', 'cluster', 'console',
  'constants', 'crypto', 'dgram', 'diagnostics_channel', 'dns', 'domain',
  'events', 'fs', 'fs/promises', 'http', 'http2', 'http3', 'https',
  'inspector', 'module', 'net', 'os', 'path', 'perf_hooks', 'process',
  'punycode', 'querystring', 'readline', 'readline/promises', 'repl', 'sea',
  'sqlite', 'stream', 'stream/promises', 'stream/consumers', 'stream/web',
  'string_decoder', 'test', 'timers', 'timers/promises', 'tls', 'tty', 'url',
  'util', 'v8', 'vm', 'wasi', 'worker_threads', 'zlib',
];

// ECMAScript built-ins are standard JavaScript, not GodeJS host APIs. The
// generated global page links to MDN for those and documents everything else.
const STANDARD = new Set([
  'Object', 'Function', 'Array', 'Number', 'parseFloat', 'parseInt',
  'Infinity', 'NaN', 'undefined', 'Boolean', 'String', 'Symbol', 'Date',
  'Promise', 'RegExp', 'Error', 'EvalError', 'RangeError', 'ReferenceError',
  'SyntaxError', 'TypeError', 'URIError', 'AggregateError', 'globalThis',
  'JSON', 'Math', 'Intl', 'ArrayBuffer', 'SharedArrayBuffer', 'DataView',
  'Atomics', 'WeakRef', 'FinalizationRegistry', 'Map', 'Set', 'WeakMap',
  'WeakSet', 'BigInt', 'BigInt64Array', 'BigUint64Array', 'eval', 'isFinite',
  'isNaN', 'escape', 'unescape', 'decodeURI', 'decodeURIComponent',
  'encodeURI', 'encodeURIComponent', 'Reflect', 'Proxy', 'Int8Array',
  'Uint8Array', 'Uint8ClampedArray', 'Int16Array', 'Uint16Array',
  'Int32Array', 'Uint32Array', 'Float16Array', 'Float32Array',
  'Float64Array', 'Iterator', 'Temporal', 'DisposableStack',
  'AsyncDisposableStack', 'SuppressedError', 'WebAssembly',
]);

// These globals have a full reference page of their own. Listing their whole
// surface again on the globals page would duplicate thousands of entries.
const MODULE_PAGE = new Set(['Buffer', 'console', 'os', 'process', 'global']);

const FUNCTION_META = new Set([
  'length', 'name', 'arguments', 'caller', 'prototype', 'constructor',
  'toString', 'toLocaleString', 'valueOf', 'apply', 'call', 'bind',
  'hasOwnProperty', 'isPrototypeOf', 'propertyIsEnumerable', '__proto__',
]);

const out = {};

function kind(v) {
  if (typeof v === 'function') return 'function';
  if (v === null) return 'null';
  if (Array.isArray(v)) return 'array';
  return typeof v;
}

function addProto(value, prefix, rows, seen) {
  let proto;
  try {
    proto = Object.getPrototypeOf(value);
  } catch (err) {
    return;
  }
  if (!proto || proto === Object.prototype || seen.indexOf(proto) >= 0) return;
  seen.push(proto);
  let members;
  try {
    members = Object.getOwnPropertyNames(proto);
  } catch (err) {
    return;
  }
  for (const member of members) {
    if (member === 'constructor') continue;
    if (typeof value === 'function' && FUNCTION_META.has(member)) continue;
    let memberValue;
    try {
      memberValue = proto[member];
    } catch (err) {
      continue;
    }
    const memberType = kind(memberValue);
    rows.push({
      p: prefix + '.' + member + (memberType === 'function' ? '()' : ''),
      k: memberType,
      a: typeof memberValue === 'function' ? memberValue.length : 0,
    });
  }
}

function walk(obj, prefix, depth, seen) {
  const rows = [];
  let keys;
  try {
    keys = Object.getOwnPropertyNames(obj);
  } catch (err) {
    return rows;
  }
  for (const key of keys) {
    if (key === 'prototype' || key === 'constructor' || key === '__proto__') {
      continue;
    }
    let value;
    try {
      value = obj[key];
    } catch (err) {
      rows.push({ p: prefix + '.' + key, k: 'getter-error', a: 0 });
      continue;
    }
    const type = kind(value);
    rows.push({ p: prefix + '.' + key, k: type, a: typeof value === 'function' ? value.length : 0 });

    if (depth > 0 && type === 'object' && value && seen.indexOf(value) < 0) {
      seen.push(value);
      rows.push.apply(rows, walk(value, prefix + '.' + key, depth - 1, seen));
      addProto(value, prefix + '.' + key, rows, seen);
    }

    if (depth > 0 && type === 'function' && value.prototype && seen.indexOf(value.prototype) < 0) {
      const proto = value.prototype;
      seen.push(proto);
      for (const member of Object.getOwnPropertyNames(proto)) {
        if (member === 'constructor') continue;
        let memberValue;
        try {
          memberValue = proto[member];
        } catch (err) {
          continue;
        }
        const memberType = kind(memberValue);
        rows.push({
          p: prefix + '.' + key + '.prototype.' + member + (memberType === 'function' ? '()' : ''),
          k: memberType,
          a: typeof memberValue === 'function' ? memberValue.length : 0,
        });
      }
    }
  }
  return rows;
}

function dumpGlobals() {
  const rows = [];
  let names;
  try {
    names = Object.getOwnPropertyNames(globalThis);
  } catch (err) {
    return rows;
  }
  for (const key of names) {
    if (key === 'globalThis' || key.startsWith('__gode') || STANDARD.has(key)) {
      continue;
    }
    let value;
    try {
      value = globalThis[key];
    } catch (err) {
      rows.push({ p: 'globalThis.' + key, k: 'getter-error', a: 0 });
      continue;
    }
    const type = kind(value);
    rows.push({ p: 'globalThis.' + key, k: type, a: typeof value === 'function' ? value.length : 0 });
    if (type === 'function' && value.prototype) {
      for (const member of Object.getOwnPropertyNames(value)) {
        if (FUNCTION_META.has(member)) {
          continue;
        }
        let staticValue;
        try {
          staticValue = value[member];
        } catch (err) {
          continue;
        }
        const staticType = kind(staticValue);
        rows.push({
          p: 'globalThis.' + key + '.' + member + (staticType === 'function' ? '()' : ''),
          k: staticType,
          a: typeof staticValue === 'function' ? staticValue.length : 0,
        });
      }
      const proto = value.prototype;
      for (const member of Object.getOwnPropertyNames(proto)) {
        if (member === 'constructor') continue;
        let memberValue;
        try {
          memberValue = proto[member];
        } catch (err) {
          continue;
        }
        const memberType = kind(memberValue);
        rows.push({
          p: 'globalThis.' + key + '.prototype.' + member + (memberType === 'function' ? '()' : ''),
          k: memberType,
          a: typeof memberValue === 'function' ? memberValue.length : 0,
        });
      }
    }
    if (type === 'object' && value && !MODULE_PAGE.has(key)) {
      addProto(value, 'globalThis.' + key, rows, []);
      for (const member of Object.getOwnPropertyNames(value)) {
        if (FUNCTION_META.has(member)) continue;
        let memberValue;
        try {
          memberValue = value[member];
        } catch (err) {
          continue;
        }
        const memberType = kind(memberValue);
        rows.push({
          p: 'globalThis.' + key + '.' + member + (memberType === 'function' ? '()' : ''),
          k: memberType,
          a: typeof memberValue === 'function' ? memberValue.length : 0,
        });
      }
    }
  }
  return rows;
}

for (const name of mods) {
  try {
    const exports_ = require(name);
    out[name] = walk(exports_, name, 2, [exports_]);
  } catch (err) {
    out[name] = [{ p: name, k: 'ERROR', a: 0, err: String((err && err.message) || err) }];
  }
}
out.__globals__ = dumpGlobals();

console.log(JSON.stringify(out));
