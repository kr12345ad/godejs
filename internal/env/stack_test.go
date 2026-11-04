package env_test

import (
	"strings"
	"testing"
)

func TestCallSiteV8API(t *testing.T) {
	got := mustRun(t, "main.js", map[string]string{
		"main.js": `'use strict';
var stack;
Error.prepareStackTrace = function (e, sites) { return sites; };
function callee() { return new Error('x'); }
class Widget { constructor() { this.err = new Error('ctor'); } }
var obj = {
  method: function () { return new Error('m'); },
  get prop() { return new Error('g'); }
};
globalThis.obj = obj;
function globalFn() { return new Error('gf'); }
globalThis.globalFn = globalFn;
var plain = callee().stack;
var ctor = new Widget().err.stack;
var method = globalThis.obj.method().stack;
var getter = globalThis.obj.prop.stack;
var globalStack = globalFn().stack;
var required = ['getThis', 'getTypeName', 'getFunction', 'getFunctionName',
  'getMethodName', 'getFileName', 'getLineNumber', 'getColumnNumber',
  'getEvalOrigin', 'isToplevel', 'isEval', 'isNative', 'isConstructor'];
var missing = required.filter(function (key) {
  return typeof plain[0][key] !== 'function';
});
function capture() {
  var target = {};
  Error.captureStackTrace(target);
  return target.stack;
}
var captured = capture();
var evalStack = eval('new Error("e")').stack;
globalThis.__out = [
  missing.length,
  plain[0].getFunctionName() === 'callee',
  plain[0].getFileName() === __filename,
  plain[0].getLineNumber(),
  plain[0].getColumnNumber() > 0,
  plain[0].getScriptHash().length,
  typeof plain[0].getThis(),
  String(plain[0].getTypeName()),
  String(plain[0].getMethodName()),
  String(plain[0].getEvalOrigin()),
  plain[0].isToplevel(),
  plain[0].isEval(),
  plain[0].isNative(),
  plain[0].isConstructor(),
  globalStack[0].getFunction() === globalFn,
  method[0].getFunctionName(),
  method[0].getMethodName(),
  method[0].getTypeName(),
  method[0].getThis() === obj,
  typeof method[0].getFunction(),
  ctor[0].isConstructor(),
  String(ctor[0].getThis()),
  getter[0].getFunctionName(),
  typeof getter[0].getThis(),
  captured[0].getFunctionName() === 'capture',
  evalStack[0].isEval(),
  evalStack[0].getFileName() === '',
  String(plain[0]).indexOf('callee (') === 0
].join('|');
`,
	})
	want := "0|true|true|4|true|64|undefined|null|null|undefined|true|false|false|false|true|method|method|Object|true|function|true|undefined|get prop|object|true|true|true|true"
	if got != want {
		t.Fatalf("CallSite output:\n got %q\nwant %q", got, want)
	}
}

func TestDefaultStackFormatPreserved(t *testing.T) {
	got := mustRun(t, "main.js", map[string]string{
		"main.js": `function f() { return new Error('d').stack; }
globalThis.__out = f();`,
	})
	if !strings.HasPrefix(got, "f@") {
		t.Fatalf("default stack did not start with JSC frame %q", got)
	}
	if !strings.Contains(got, "main.js:") {
		t.Fatalf("default stack lost the source location: %q", got)
	}
}

func TestRequireDottedBasenameFallsBackToJS(t *testing.T) {
	got := mustRun(t, "main.js", map[string]string{
		"util.inspect.js": `module.exports = 'dotted';`,
		"main.js":         `globalThis.__out = require('./util.inspect');`,
	})
	if got != "dotted" {
		t.Fatalf("require('./util.inspect') = %q, want dotted", got)
	}
}
