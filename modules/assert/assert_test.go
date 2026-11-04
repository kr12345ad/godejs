package assert

import (
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newEnv(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "assert", Install)
	return e
}

func TestModuleIsCallable(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `(() => { assert(true); assert.ok(1); return "ok"; })()`), "ok")
}

func TestOkAndAssertionErrorShape(t *testing.T) {
	e := newEnv(t)

	err := jstest.Throws(t, e, `assert(false)`)
	if err.Name != "AssertionError" {
		t.Fatalf("name = %q, want AssertionError", err.Name)
	}

	got := jstest.Str(t, e, `(() => {
		try { assert.strictEqual(1, 2); } catch (e) {
			return [e.name, e.code, e.actual, e.expected, e.operator].join(",");
		}
	})()`)
	jstest.Equal(t, got, "AssertionError,ERR_ASSERTION,1,2,strictEqual")
}

func TestStrictAndLooseEquality(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => { assert.strictEqual("a", "a"); assert.notStrictEqual(1, "1"); return "ok"; })()`), "ok")
	jstest.Equal(t, jstest.Str(t, e, `(() => { assert.equal(1, "1"); assert.notEqual(1, 2); return "ok"; })()`), "ok")

	// Object.is(NaN, NaN) is true, so strictEqual must accept it.
	jstest.Equal(t, jstest.Str(t, e, `(() => { assert.strictEqual(NaN, NaN); return "ok"; })()`), "ok")
}

func TestDeepStrictEqual(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.deepStrictEqual({a: [1, {b: NaN}]}, {a: [1, {b: NaN}]});
		assert.deepStrictEqual(new Date(0), new Date(0));
		assert.deepStrictEqual(/x/gi, /x/gi);
		return "ok";
	})()`), "ok")

	jstest.Throws(t, e, `assert.deepStrictEqual({a: 1}, {a: 2})`)
	jstest.Throws(t, e, `assert.deepStrictEqual([1, 2], [1, 2, 3])`)
	jstest.Throws(t, e, `assert.deepStrictEqual(0, -0)`)
	jstest.Throws(t, e, `assert.deepStrictEqual({a: 1}, {b: 1})`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Date(0), new Date(1))`)
}

func TestLooseAndPartialDeepEquality(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => { assert.deepEqual(1, "1"); assert.notDeepEqual({a: 1}, {a: 2}); return "ok"; })()`), "ok")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.partialDeepStrictEqual({a: 1, b: 2}, {a: 1});
		assert.partialDeepStrictEqual({a: {b: 1, c: 2}}, {a: {b: 1}});
		return "ok";
	})()`), "ok")

	jstest.Throws(t, e, `assert.partialDeepStrictEqual({a: 1}, {a: 2})`)
}

func TestAssertionErrorMessages(t *testing.T) {
	e := newEnv(t)

	// Node formats a simple strict-equality failure without a diff header.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { assert.strictEqual(1, 2); } catch (e) { return e.message; }
	})()`), "Expected values to be strictly equal:\n\n1 !== 2\n")

	// err.toString() carries the code, and name stays 'AssertionError'.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { assert.strictEqual(1, 2); } catch (e) {
			return e.name + '|' + e.code + '|' + e.toString();
		}
	})()`), "AssertionError|ERR_ASSERTION|AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:\n\n1 !== 2\n")

	// A falsy assertion reports the failing source line, like Node.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { assert.ok(false); } catch (e) { return e.message.split('\n')[0]; }
	})()`), "The expression evaluated to a falsy value:")

	// Missing or mismatched properties on a thrown error render Node's
	// `Comparison` placeholder objects.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try {
			assert.throws(function () { var err = new Error('x'); err.code = 'E1'; throw err; }, { code: 'E2' });
			return 'no-throw';
		} catch (e) {
			return e.message.indexOf('Expected values to be strictly deep-equal:') === 0 ? 'formatted' : e.message;
		}
	})()`), "formatted")

	// Unknown operators render `actual OP expected` without a diff header.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { assert.equal(1, 2); } catch (e) { return e.message; }
	})()`), "1 == 2")
}

func TestThrowsAndDoesNotThrow(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.throws(function () { throw new TypeError("x"); }, TypeError);
		assert.throws(function () { throw new Error("boom"); }, /boom/);
		assert.throws(function () { var e = new Error("x"); e.code = "E1"; throw e; }, {code: "E1"});
		assert.throws(function () { throw new Error("x"); }, function (err) { return err.message === "x"; });
		assert.doesNotThrow(function () { return 1; });
		return "ok";
	})()`), "ok")

	// Node semantics: a plain function is a validation function that must
	// return exactly true, and assert.throws returns undefined.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		var seen;
		assert.throws(function () { throw new Error("x"); }, function (err) { seen = err.message; return true; });
		return String(seen);
	})()`), "x")
	jstest.Equal(t, jstest.Str(t, e, `String(assert.throws(function () { throw new Error("x"); }))`), "undefined")
	jstest.Throws(t, e, `assert.throws(function () { throw new Error("x"); }, function () { return "no"; })`)

	jstest.Throws(t, e, `assert.throws(function () { return 1; })`)
	jstest.Throws(t, e, `assert.throws(function () { throw new Error("boom"); }, TypeError)`)
	jstest.Throws(t, e, `assert.doesNotThrow(function () { throw new Error("boom"); })`)
}

func TestFailIfErrorMatch(t *testing.T) {
	e := newEnv(t)

	err := jstest.Throws(t, e, `assert.fail("custom failure")`)
	if err.Name != "AssertionError" || err.Message != "custom failure" {
		t.Fatalf("fail error = %#v", err)
	}
	jstest.Throws(t, e, `assert.ifError(new Error("bad"))`)
	jstest.Equal(t, jstest.Str(t, e, `(() => { assert.ifError(null); assert.ifError(undefined); return "ok"; })()`), "ok")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.match("hello world", /world/);
		assert.doesNotMatch("hello world", /nope/);
		return "ok";
	})()`), "ok")
	jstest.Throws(t, e, `assert.match("hello", /nope/)`)
	jstest.Throws(t, e, `assert.doesNotMatch("hello", /ell/)`)
}

func TestAssertionErrorClass(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		var e = new assert.AssertionError({message: "m"});
		return [e instanceof Error, e.name, e.code, e.message].join(",");
	})()`)
	jstest.Equal(t, got, "true,AssertionError,ERR_ASSERTION,m")
}

func TestRejectsReturnsThenable(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		var p = assert.rejects(function () { return Promise.reject(new Error("x")); }, /x/);
		var q = assert.doesNotReject(function () { return Promise.resolve(1); });
		return typeof p.then + "," + typeof q.then;
	})()`)
	jstest.Equal(t, got, "function,function")

	// Non-promise functions produce a rejected promise, like Node.
	got = jstest.Str(t, e, `(() => {
		var p = assert.rejects(function () { return 1; });
		p.catch(function () {});
		return typeof p.then;
	})()`)
	jstest.Equal(t, got, "function")
}

func TestCallTracker(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		const tracker = new assert.CallTracker();
		function foo(a) { return a * 2; }
		const calls = tracker.calls(foo, 2);
		const first = calls(2);
		const second = calls.call({x: 1}, 3);
		const list = tracker.getCalls(calls);
		const snapshot = [first, second, list.length,
			list[0].thisArg === undefined, list[1].thisArg.x,
			list[0].arguments[0], list[1].arguments[0],
			Object.isFrozen(list), Object.isFrozen(list[0]),
			Object.isFrozen(list[0].arguments)].join(',');
		tracker.reset(calls);
		return snapshot + '|' + tracker.getCalls(calls).length;
	})()`)
	jstest.Equal(t, got, "4,6,2,true,1,2,3,true,true,true|0")

	codes := jstest.Str(t, e, `(() => {
		const tracker = new assert.CallTracker();
		const out = [];
		for (const value of ['1', 0.1, true, null]) {
			try { tracker.calls(function() {}, value); } catch (e) { out.push(e.code); }
		}
		try { tracker.getCalls(function() {}); } catch (e) { out.push(e.code); }
		return out.join(',');
	})()`)
	jstest.Equal(t, codes, "ERR_INVALID_ARG_TYPE,ERR_OUT_OF_RANGE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_TYPE,ERR_INVALID_ARG_VALUE")
}

func TestCallTrackerReportAndVerify(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		const tracker = new assert.CallTracker();
		const foo = tracker.calls(function foo() {}, 1);
		const bar = tracker.calls(function bar() {}, 1);
		const before = tracker.report().map((item) => item.operator).join(',');
		foo();
		const afterOne = tracker.report().length;
		try {
			tracker.verify();
		} catch (e) {
			return [before, afterOne, e.name, e.code, e.message].join('|');
		}
	})()`)
	jstest.Equal(t, got, "foo,bar|1|AssertionError|ERR_ASSERTION|Expected the bar function to be executed 1 time(s) but was executed 0 time(s).")
}

func TestAssertClass(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		const strict = new assert.Assert();
		const loose = new assert.Assert({ strict: false });
		const results = [
			strict.equal === strict.strictEqual,
			strict.deepEqual === strict.deepStrictEqual,
			strict.notEqual === strict.notStrictEqual,
			typeof strict.ok, typeof loose.ok
		];
		try { assert.Assert(); } catch (e) { results.push(e.code, e.name); }
		loose.equal(1, '1');
		try { loose.deepEqual(/a/); } catch (e) { results.push(e.code); }
		function A() { this.x = 1; }
		function B() { this.x = 1; }
		new assert.Assert({ skipPrototype: true }).deepStrictEqual(new A(), new B());
		return results.join(',');
	})()`)
	jstest.Equal(t, got, "true,true,true,function,function,ERR_CONSTRUCT_CALL_REQUIRED,TypeError,ERR_MISSING_ARGS")

	failed := jstest.Str(t, e, `(() => {
		try { new assert.Assert().fail(); } catch (e) {
			return [e.message, e.operator, e.generatedMessage].join(',');
		}
	})()`)
	jstest.Equal(t, failed, "Failed,fail,true")

	code := jstest.Str(t, e, `(() => {
		try { new assert.Assert({ diff: 'invalid' }); } catch (e) { return e.code; }
	})()`)
	jstest.Equal(t, code, "ERR_INVALID_ARG_VALUE")
}

func TestThrowsMissingExceptionMessages(t *testing.T) {
	e := newEnv(t)

	cases := map[string]string{
		`(() => { try { assert.throws(function () {}); } catch (e) { return e.message; } })()`:                    "Missing expected exception.",
		`(() => { try { assert.throws(function () {}, TypeError); } catch (e) { return e.message; } })()`:         "Missing expected exception (TypeError).",
		`(() => { try { assert.throws(function () {}, 'boom'); } catch (e) { return e.message; } })()`:            "Missing expected exception: boom",
		`(() => { try { assert.throws(function () {}, TypeError, 'boom'); } catch (e) { return e.message; } })()`: "Missing expected exception (TypeError): boom",
	}
	for src, want := range cases {
		jstest.Equal(t, jstest.Str(t, e, src), want)
	}

	// A string in the error slot is ambiguous when it equals the thrown message.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { assert.throws(function () { throw new Error('boom'); }, 'boom'); }
		catch (e) { return e.code + '|' + e.message; }
	})()`), "ERR_AMBIGUOUS_ARGUMENT|The \"error/message\" argument is ambiguous. "+
		"The error message \"boom\" is identical to the message.")
}

func TestErrorCauseComparison(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.deepStrictEqual(new Error('a', { cause: new Error('x') }), new Error('a', { cause: new Error('x') }));
		assert.deepStrictEqual(new Error('a', { cause: undefined }), new Error('a', { cause: undefined }));
		return 'ok';
	})()`), "ok")

	// Different causes, and a missing versus an explicitly undefined cause, differ.
	jstest.Throws(t, e, `assert.deepStrictEqual(new Error('a', { cause: new Error('x') }), new Error('a', { cause: new Error('y') }))`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Error('a'), new Error('a', { cause: undefined }))`)

	// Errors are never deep-equal to plain objects.
	jstest.Throws(t, e, `assert.deepStrictEqual(new Error('a'), { message: 'a' })`)
}

func TestDeepEqualityBuiltins(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.deepStrictEqual(new Map([[1, { a: 1 }]]), new Map([[1, { a: 1 }]]));
		assert.deepStrictEqual(new Set([{ a: 1 }, 2]), new Set([2, { a: 1 }]));
		assert.deepStrictEqual(new Uint8Array([1, 2]), new Uint8Array([1, 2]));
		assert.deepStrictEqual(new Number(3), new Number(3));
		assert.deepStrictEqual(Object.create(null), Object.create(null));
		return 'ok';
	})()`), "ok")

	jstest.Throws(t, e, `assert.deepStrictEqual(new Map([[1, 2]]), new Map([[1, 3]]))`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Set([1]), new Set([2]))`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Int16Array(1), new Uint16Array(1))`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Error('a'), new TypeError('a'))`)
	jstest.Throws(t, e, `assert.deepStrictEqual(new Number(1), new String('1'))`)

	// Circular structures terminate and compare structurally.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		var a = { name: 'a' }, b = { name: 'a' };
		a.self = a; b.self = b;
		assert.deepStrictEqual(a, b);
		b.self = { name: 'a' };
		try { assert.deepStrictEqual(a, b); return 'equal'; } catch (e) { return 'different'; }
	})()`), "different")

	// A repeated unequal pair must not be mistaken for a cycle.
	jstest.Throws(t, e, `(() => {
		var shared = { v: 1 };
		assert.deepStrictEqual({ a: shared, b: shared }, { a: { v: 2 }, b: { v: 2 } });
	})()`)

	// Loose mode compares float elements numerically but keeps element types apart.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		assert.deepEqual(new Float64Array([0]), new Float64Array([-0]));
		return 'ok';
	})()`), "ok")
	jstest.Throws(t, e, `assert.deepStrictEqual(new Float64Array([0]), new Float64Array([-0]))`)
	jstest.Throws(t, e, `assert.deepEqual(new Float64Array([0]), new Float32Array([0]))`)

	// A boxed primitive never equals its primitive counterpart.
	jstest.Throws(t, e, `assert.deepEqual(new String('1'), '1')`)
}

func TestAssertOkIdentityAndGeneratedMessages(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(assert.ok === assert) + ',' + (assert.strict.ok === assert)`), "true,true")

	got := jstest.Str(t, e, `(() => {
		try { assert.ok(); } catch (e) { return e.message + '|' + e.generatedMessage; }
	})()`)
	jstest.Equal(t, got, "No value argument passed to `assert.ok()`|true")

	got = jstest.Str(t, e, `(() => {
		try { assert.strictEqual(1, 2, 'oh no'); } catch (e) { return e.message + '|' + e.generatedMessage; }
	})()`)
	jstest.Equal(t, got, "oh no\n\n1 !== 2\n|false")
}

func TestCallTrackerVerifyDetails(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		const tracker = new assert.CallTracker();
		const a = tracker.calls(function a() {}, 1);
		const b = tracker.calls(function b() {}, 1);
		a();
		try { tracker.verify(); } catch (e) {
			return [e.message, e['message 0'], e.actual, e.expected, e.operator].join('|');
		}
	})()`)
	// Node sets actual/expected/operator to undefined when it builds the error
	// from `details`, and Array.prototype.join renders those as empty strings.
	jstest.Equal(t, got, "Expected the b function to be executed 1 time(s) but was executed 0 time(s)."+
		"|Expected the b function to be executed 1 time(s) but was executed 0 time(s).|||")

	// validateUint32 argument handling matches Node's validator.
	got = jstest.Str(t, e, `(() => {
		const tracker = new assert.CallTracker();
		const out = [];
		for (const value of [0, 1.5, 4294967296]) {
			try { tracker.calls(function () {}, value); } catch (e) { out.push(e.code + ':' + e.message); }
		}
		return out.join('|');
	})()`)
	jstest.Equal(t, got, "ERR_OUT_OF_RANGE:The value of \"expected\" is out of range. It must be >= 1 && <= 4294967295. Received 0"+
		"|ERR_OUT_OF_RANGE:The value of \"expected\" is out of range. It must be an integer. Received 1.5"+
		"|ERR_OUT_OF_RANGE:The value of \"expected\" is out of range. It must be >= 1 && <= 4294967295. Received 4294967296")
}

func TestAssertClassDiffOption(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		const full = new assert.Assert({ diff: 'full' });
		const simple = new assert.Assert({ diff: 'simple' });
		const out = [];
		try { full.strictEqual({ a: 1 }, { a: 2 }); } catch (e) { out.push(e.diff); }
		try { simple.strictEqual({ a: 1 }, { a: 2 }); } catch (e) { out.push(e.diff); }
		const { strictEqual } = full;
		try { strictEqual({ a: 1 }, { a: 2 }); } catch (e) { out.push(e.diff); }
		return out.join(',');
	})()`)
	jstest.Equal(t, got, "full,simple,simple")

	got = jstest.Str(t, e, `(() => {
		try { new assert.Assert({ diff: 'invalid' }); } catch (e) { return e.name + '|' + e.code + '|' + e.message; }
	})()`)
	jstest.Equal(t, got, "TypeError|ERR_INVALID_ARG_VALUE|The property 'options.diff' must be one of: 'simple', 'full'. Received 'invalid'")

	// skipPrototype drops only the prototype check, still comparing values.
	got = jstest.Str(t, e, `(() => {
		function A() { this.x = 1; }
		function B() { this.y = 2; }
		const loose = new assert.Assert({ skipPrototype: true });
		loose.deepStrictEqual(new A(), new A());
		try { loose.deepStrictEqual(new A(), new B()); return 'equal'; } catch (e) { return 'different'; }
	})()`)
	jstest.Equal(t, got, "different")
}

func TestMatchTypeErrorMessage(t *testing.T) {
	e := newEnv(t)

	got := jstest.Str(t, e, `(() => {
		try { assert.match({ abc: 123 }, /abc/); } catch (e) {
			return e.code + '|' + e.operator + '|' + e.message;
		}
	})()`)
	jstest.Equal(t, got, "ERR_ASSERTION|match|The \"string\" argument must be of type string. Received type object ({ abc: 123 })")

	// A RegExp argument that is not a RegExp fails validation first.
	got = jstest.Str(t, e, `(() => {
		try { assert.match(/abc/, 'string'); } catch (e) { return e.code + '|' + e.message; }
	})()`)
	jstest.Equal(t, got, "ERR_INVALID_ARG_TYPE|The \"regexp\" argument must be an instance of RegExp. Received type string ('string')")
}

func TestImplementedKeysAreTracked(t *testing.T) {
	keys := Implemented()
	if len(keys) != 27 {
		t.Fatalf("Implemented returned %d keys", len(keys))
	}
	for _, key := range keys {
		if !strings.HasPrefix(key, "assert.") {
			t.Fatalf("bad key %q", key)
		}
	}
}
