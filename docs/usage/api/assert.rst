assert
======

The ``node:assert`` module provides a set of assertion functions for verifying invariants.

Loading
-------

.. code-block:: javascript

   const assert = require('node:assert');
   // ESM
   import * as assert from 'node:assert';

API
---

Assert
~~~~~~

.. code-block:: text

   assert.Assert

*Kind:* class · *Arguments:* 1

The ``Assert`` class allows creating independent assertion instances with custom options.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Assert } = require('node:assert');
   const assertInstance = new Assert({ diff: 'full' });
   assertInstance.deepStrictEqual({ a: 1 }, { a: 2 });
   // Shows a full diff in the error message.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#class-assertassert>`__

AssertionError
~~~~~~~~~~~~~~

.. code-block:: text

   assert.AssertionError

*Kind:* class · *Arguments:* 1

Indicates the failure of an assertion. All errors thrown by the ``node:assert`` module will be instances of the ``AssertionError`` class.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   // Generate an AssertionError to compare the error message later:
   const { message } = new assert.AssertionError({
     actual: 1,
     expected: 2,
     operator: 'strictEqual',
   });

   // Verify error output:
   try {
     assert.strictEqual(1, 2);
   } catch (err) {
     assert(err instanceof assert.AssertionError);
     assert.strictEqual(err.message, message);
     assert.strictEqual(err.name, 'AssertionError');
     assert.strictEqual(err.actual, 1);
     assert.strictEqual(err.expected, 2);
     assert.strictEqual(err.code, 'ERR_ASSERTION');
     assert.strictEqual(err.operator, 'strictEqual');
     assert.strictEqual(err.generatedMessage, true);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#class-assertassertionerror>`__

CallTracker
~~~~~~~~~~~

.. code-block:: text

   assert.CallTracker

*Kind:* class · *Arguments:* 0

This feature is deprecated and will be removed in a future version. Please consider using alternatives such as the ``mock`` helper function.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const tracker = new assert.CallTracker();

   function func() {}

   // callsfunc() must be called exactly 1 time before tracker.verify().
   const callsfunc = tracker.calls(func, 1);

   callsfunc();

   // Calls tracker.verify() and verifies if all tracker.calls() functions have
   // been called exact times.
   process.on('exit', () => {
     tracker.verify();
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#class-assertcalltracker>`__

deepEqual
~~~~~~~~~

.. code-block:: text

   assert.deepEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.deepStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');
   // WARNING: This does not throw an AssertionError!

   assert.deepEqual('+00000000', false);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdeepequalactual-expected-message>`__

deepStrictEqual
~~~~~~~~~~~~~~~

.. code-block:: text

   assert.deepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for deep equality between the ``actual`` and ``expected`` parameters. "Deep" equality means that the enumerable "own" properties of child objects are recursively evaluated also by the following rules.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   // This fails because 1 !== '1'.
   assert.deepStrictEqual({ a: 1 }, { a: '1' });
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   //   {
   // +   a: 1
   // -   a: '1'
   //   }

   // The following objects don't have own properties
   const date = new Date();
   const object = {};
   const fakeDate = {};
   Object.setPrototypeOf(fakeDate, Date.prototype);

   // Different [[Prototype]]:
   assert.deepStrictEqual(object, fakeDate);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + {}
   // - Date {}

   // Different type tags:
   assert.deepStrictEqual(date, fakeDate);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + 2018-04-26T00:49:08.604Z
   // - Date {}

   assert.deepStrictEqual(NaN, NaN);
   // OK because Object.is(NaN, NaN) is true.

   // Different unwrapped numbers:
   assert.deepStrictEqual(new Number(1), new Number(2));
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + [Number: 1]
   // - [Number: 2]

   assert.deepStrictEqual(new String('foo'), Object('foo'));
   // OK because the object and the string are identical when unwrapped.

   assert.deepStrictEqual(-0, -0);
   // OK

   // Different zeros:
   assert.deepStrictEqual(0, -0);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + 0
   // - -0

   const symbol1 = Symbol();
   const symbol2 = Symbol();
   assert.deepStrictEqual({ [symbol1]: 1 }, { [symbol1]: 1 });
   // OK, because it is the same symbol on both objects.

   assert.deepStrictEqual({ [symbol1]: 1 }, { [symbol2]: 1 });
   // AssertionError [ERR_ASSERTION]: Inputs identical but not reference equal:
   //
   // {
   //   Symbol(): 1
   // }

   const weakMap1 = new WeakMap();
   const weakMap2 = new WeakMap();
   const obj = {};

   weakMap1.set(obj, 'value');
   weakMap2.set(obj, 'value');

   // Comparing different instances fails, even with same contents
   assert.deepStrictEqual(weakMap1, weakMap2);
   // AssertionError: Values have same structure but are not reference-equal:
   //
   // WeakMap {
   //   <items unknown>
   // }

   // Comparing the same instance to itself succeeds
   assert.deepStrictEqual(weakMap1, weakMap1);
   // OK

   const weakSet1 = new WeakSet();
   const weakSet2 = new WeakSet();
   weakSet1.add(obj);
   weakSet2.add(obj);

   // Comparing different instances fails, even with same contents
   assert.deepStrictEqual(weakSet1, weakSet2);
   // AssertionError: Values have same structure but are not reference-equal:
   // + actual - expected
   //
   // WeakSet {
   //   <items unknown>
   // }

   // Comparing the same instance to itself succeeds
   assert.deepStrictEqual(weakSet1, weakSet1);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdeepstrictequalactual-expected-message>`__

doesNotMatch
~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotMatch(string, regexp[, message])

*Kind:* function · *Arguments:* 3

Expects the ``string`` input not to match the regular expression.

If the values do match, or if the ``string`` argument is of another type than ``string``, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``string`` (*string*)
* ``regexp`` (*RegExp*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.doesNotMatch('I will fail', /fail/);
   // AssertionError [ERR_ASSERTION]: The input was expected to not match the ...

   assert.doesNotMatch(123, /pass/);
   // AssertionError [ERR_ASSERTION]: The "string" argument must be of type string.

   assert.doesNotMatch('I will pass', /different/);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotmatchstring-regexp-message>`__

doesNotReject
~~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotReject(asyncFn[, error][, message])

*Kind:* function · *Arguments:* 3

Awaits the ``asyncFn`` promise or, if ``asyncFn`` is a function, immediately calls the function and awaits the returned promise to complete. It will then check that the promise is not rejected.

If ``asyncFn`` is a function and it throws an error synchronously, ``assert.doesNotReject()`` will return a rejected ``Promise`` with that error. If the function does not return a promise, ``assert.doesNotReject()`` will return a rejected ``Promise`` with an ``ERR_INVALID_RETURN_VALUE`` error. In both cases the error handler is skipped.

**Arguments**

* ``asyncFn`` (*Function|Promise*)
* ``error`` (*RegExp|Function*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   (async () => {
     await assert.doesNotReject(
       async () => {
         throw new TypeError('Wrong value');
       },
       SyntaxError,
     );
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotrejectasyncfn-error-message>`__

doesNotThrow
~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotThrow(fn[, error][, message])

*Kind:* function · *Arguments:* 3

Asserts that the function ``fn`` does not throw an error.

Using ``assert.doesNotThrow()`` is actually not useful because there is no benefit in catching an error and then rethrowing it. Instead, consider adding a comment next to the specific code path that should not throw and keep error messages as expressive as possible.

**Arguments**

* ``fn`` (*Function*)
* ``error`` (*RegExp|Function*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.doesNotThrow(
     () => {
       throw new TypeError('Wrong value');
     },
     SyntaxError,
   );

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotthrowfn-error-message>`__

equal
~~~~~

.. code-block:: text

   assert.equal(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.strictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.equal(1, 1);
   // OK, 1 == 1
   assert.equal(1, '1');
   // OK, 1 == '1'
   assert.equal(NaN, NaN);
   // OK

   assert.equal(1, 2);
   // AssertionError: 1 == 2
   assert.equal({ a: { b: 1 } }, { a: { b: 1 } });
   // AssertionError: { a: { b: 1 } } == { a: { b: 1 } }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertequalactual-expected-message>`__

fail
~~~~

.. code-block:: text

   assert.fail([message])

*Kind:* function · *Arguments:* 1

Throws an ``AssertionError`` with the provided error message or a default error message. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

Using ``assert.fail()`` with more than two arguments is possible but deprecated. See below for further details.

**Arguments**

* ``message`` (*string|Error*) — optional: **Default:** ``'Failed'``

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.fail();
   // AssertionError [ERR_ASSERTION]: Failed

   assert.fail('boom');
   // AssertionError [ERR_ASSERTION]: boom

   assert.fail(new TypeError('need array'));
   // TypeError: need array

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertfailmessage>`__

ifError
~~~~~~~

.. code-block:: text

   assert.ifError(value)

*Kind:* function · *Arguments:* 1

Throws ``value`` if ``value`` is not ``undefined`` or ``null``. This is useful when testing the ``error`` argument in callbacks. The stack trace contains all frames from the error passed to ``ifError()`` including the potential new frames for ``ifError()`` itself.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.ifError(null);
   // OK
   assert.ifError(0);
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: 0
   assert.ifError('error');
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: 'error'
   assert.ifError(new Error());
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: Error

   // Create some random error frames.
   let err;
   (function errorFrame() {
     err = new Error('test error');
   })();

   (function ifErrorFrame() {
     assert.ifError(err);
   })();
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: test error
   //     at ifErrorFrame
   //     at errorFrame

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertiferrorvalue>`__

match
~~~~~

.. code-block:: text

   assert.match(string, regexp[, message])

*Kind:* function · *Arguments:* 3

Expects the ``string`` input to match the regular expression.

If the values do not match, or if the ``string`` argument is of another type than ``string``, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``string`` (*string*)
* ``regexp`` (*RegExp*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.match('I will fail', /pass/);
   // AssertionError [ERR_ASSERTION]: The input did not match the regular ...

   assert.match(123, /pass/);
   // AssertionError [ERR_ASSERTION]: The "string" argument must be of type string.

   assert.match('I will pass', /pass/);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertmatchstring-regexp-message>`__

notDeepEqual
~~~~~~~~~~~~

.. code-block:: text

   assert.notDeepEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.notDeepStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const obj1 = {
     a: {
       b: 1,
     },
   };
   const obj2 = {
     a: {
       b: 2,
     },
   };
   const obj3 = {
     a: {
       b: 1,
     },
   };
   const obj4 = { __proto__: obj1 };

   assert.notDeepEqual(obj1, obj1);
   // AssertionError: { a: { b: 1 } } notDeepEqual { a: { b: 1 } }

   assert.notDeepEqual(obj1, obj2);
   // OK

   assert.notDeepEqual(obj1, obj3);
   // AssertionError: { a: { b: 1 } } notDeepEqual { a: { b: 1 } }

   assert.notDeepEqual(obj1, obj4);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotdeepequalactual-expected-message>`__

notDeepStrictEqual
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.notDeepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for deep strict inequality. Opposite of ``assert.deepStrictEqual()``.

If the values are deeply and strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.notDeepStrictEqual({ a: 1 }, { a: '1' });
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotdeepstrictequalactual-expected-message>`__

notEqual
~~~~~~~~

.. code-block:: text

   assert.notEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.notStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.notEqual(1, 2);
   // OK

   assert.notEqual(1, 1);
   // AssertionError: 1 != 1

   assert.notEqual(1, '1');
   // AssertionError: 1 != '1'

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotequalactual-expected-message>`__

notStrictEqual
~~~~~~~~~~~~~~

.. code-block:: text

   assert.notStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests strict inequality between the ``actual`` and ``expected`` parameters as determined by ``Object.is()``.

If the values are strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.notStrictEqual(1, 2);
   // OK

   assert.notStrictEqual(1, 1);
   // AssertionError [ERR_ASSERTION]: Expected "actual" to be strictly unequal to:
   //
   // 1

   assert.notStrictEqual(1, '1');
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotstrictequalactual-expected-message>`__

ok
~~~

.. code-block:: text

   assert.ok(value[, message])

*Kind:* function · *Arguments:* 2

Tests if ``value`` is truthy. It is equivalent to ``assert.equal(!!value, true, message)``.

If ``value`` is not truthy, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is ``undefined``, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``. If no arguments are passed in at all ``message`` will be set to the string: 'No value argument passed to ``assert.ok()``'.

**Arguments**

* ``value`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.ok(true);
   // OK
   assert.ok(1);
   // OK

   assert.ok();
   // AssertionError: No value argument passed to `assert.ok()`

   assert.ok(false, 'it\'s false');
   // AssertionError: it's false

   // In the repl:
   assert.ok(typeof 123 === 'string');
   // AssertionError: false == true

   // In a file (e.g. test.js):
   assert.ok(typeof 123 === 'string');
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(typeof 123 === 'string')

   assert.ok(false);
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(false)

   assert.ok(0);
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(0)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertokvalue-message>`__

partialDeepStrictEqual
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.partialDeepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for partial deep equality between the ``actual`` and ``expected`` parameters. "Deep" equality means that the enumerable "own" properties of child objects are recursively evaluated also by the following rules. "Partial" equality means that only properties that exist on the ``expected`` parameter are going to be compared.

This method always passes the same test cases as ``assert.deepStrictEqual()``, behaving as a super set of it.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.partialDeepStrictEqual(
     { a: { b: { c: 1 } } },
     { a: { b: { c: 1 } } },
   );
   // OK

   assert.partialDeepStrictEqual(
     { a: 1, b: 2, c: 3 },
     { b: 2 },
   );
   // OK

   assert.partialDeepStrictEqual(
     [1, 2, 3, 4, 5, 6, 7, 8, 9],
     [4, 5, 8],
   );
   // OK

   assert.partialDeepStrictEqual(
     new Set([{ a: 1 }, { b: 1 }]),
     new Set([{ a: 1 }]),
   );
   // OK

   assert.partialDeepStrictEqual(
     new Map([['key1', 'value1'], ['key2', 'value2']]),
     new Map([['key2', 'value2']]),
   );
   // OK

   assert.partialDeepStrictEqual(123n, 123n);
   // OK

   assert.partialDeepStrictEqual(
     [1, 2, 3, 4, 5, 6, 7, 8, 9],
     [5, 4, 8],
   );
   // AssertionError

   assert.partialDeepStrictEqual(
     { a: 1 },
     { a: 1, b: 2 },
   );
   // AssertionError

   assert.partialDeepStrictEqual(
     { a: { b: 2 } },
     { a: { b: '2' } },
   );
   // AssertionError

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertpartialdeepstrictequalactual-expected-message>`__

rejects
~~~~~~~

.. code-block:: text

   assert.rejects(asyncFn[, error][, message])

*Kind:* function · *Arguments:* 3

Awaits the ``asyncFn`` promise or, if ``asyncFn`` is a function, immediately calls the function and awaits the returned promise to complete. It will then check that the promise is rejected.

If ``asyncFn`` is a function and it throws an error synchronously, ``assert.rejects()`` will return a rejected ``Promise`` with that error. If the function does not return a promise, ``assert.rejects()`` will return a rejected ``Promise`` with an ``ERR_INVALID_RETURN_VALUE`` error. In both cases the error handler is skipped.

**Arguments**

* ``asyncFn`` (*Function|Promise*)
* ``error`` (*RegExp|Function|Object|Error*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   (async () => {
     await assert.rejects(
       async () => {
         throw new TypeError('Wrong value');
       },
       {
         name: 'TypeError',
         message: 'Wrong value',
       },
     );
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertrejectsasyncfn-error-message>`__

strict
~~~~~~

.. code-block:: text

   assert.strict(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


strictEqual
~~~~~~~~~~~

.. code-block:: text

   assert.strictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests strict equality between the ``actual`` and ``expected`` parameters as determined by ``Object.is()``.

If the values are not strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.strictEqual(1, 2);
   // AssertionError [ERR_ASSERTION]: Expected inputs to be strictly equal:
   //
   // 1 !== 2

   assert.strictEqual(1, 1);
   // OK

   assert.strictEqual('Hello foobar', 'Hello World!');
   // AssertionError [ERR_ASSERTION]: Expected inputs to be strictly equal:
   // + actual - expected
   //
   // + 'Hello foobar'
   // - 'Hello World!'
   //          ^

   const apples = 1;
   const oranges = 2;
   assert.strictEqual(apples, oranges, `apples ${apples} !== oranges ${oranges}`);
   // AssertionError [ERR_ASSERTION]: apples 1 !== oranges 2

   assert.strictEqual(1, '1', new TypeError('Inputs are not identical'));
   // TypeError: Inputs are not identical

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertstrictequalactual-expected-message>`__

throws
~~~~~~

.. code-block:: text

   assert.throws(fn[, error][, message])

*Kind:* function · *Arguments:* 3

Expects the function ``fn`` to throw an error.

If specified, ``error`` can be a ``Class``, ``RegExp``, a validation function, a validation object where each property will be tested for strict deep equality, or an instance of error where each property will be tested for strict deep equality including the non-enumerable ``message`` and ``name`` properties. When using an object, it is also possible to use a regular expression, when validating against a string property. See below for examples.

**Arguments**

* ``fn`` (*Function*)
* ``error`` (*RegExp|Function|Object|Error*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   const err = new TypeError('Wrong value');
   err.code = 404;
   err.foo = 'bar';
   err.info = {
     nested: true,
     baz: 'text',
   };
   err.reg = /abc/i;

   assert.throws(
     () => {
       throw err;
     },
     {
       name: 'TypeError',
       message: 'Wrong value',
       info: {
         nested: true,
         baz: 'text',
       },
       // Only properties on the validation object will be tested for.
       // Using nested objects requires all properties to be present. Otherwise
       // the validation is going to fail.
     },
   );

   // Using regular expressions to validate error properties:
   assert.throws(
     () => {
       throw err;
     },
     {
       // The `name` and `message` properties are strings and using regular
       // expressions on those will match against the string. If they fail, an
       // error is thrown.
       name: /^TypeError$/,
       message: /Wrong/,
       foo: 'bar',
       info: {
         nested: true,
         // It is not possible to use regular expressions for nested properties!
         baz: 'text',
       },
       // The `reg` property contains a regular expression and only if the
       // validation object contains an identical regular expression, it is going
       // to pass.
       reg: /abc/i,
     },
   );

   // Fails due to the different `message` and `name` properties:
   assert.throws(
     () => {
       const otherErr = new Error('Not found');
       // Copy all enumerable properties from `err` to `otherErr`.
       for (const [key, value] of Object.entries(err)) {
         otherErr[key] = value;
       }
       throw otherErr;
     },
     // The error's `message` and `name` properties will also be checked when using
     // an error as validation object.
     err,
   );

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertthrowsfn-error-message>`__

Assert class
------------

**Assert** — The ``Assert`` class allows creating independent assertion instances with custom options.

Assert.deepEqual
~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.deepEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.deepStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');
   // WARNING: This does not throw an AssertionError!

   assert.deepEqual('+00000000', false);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdeepequalactual-expected-message>`__

Assert.deepStrictEqual
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.deepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for deep equality between the ``actual`` and ``expected`` parameters. "Deep" equality means that the enumerable "own" properties of child objects are recursively evaluated also by the following rules.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   // This fails because 1 !== '1'.
   assert.deepStrictEqual({ a: 1 }, { a: '1' });
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   //   {
   // +   a: 1
   // -   a: '1'
   //   }

   // The following objects don't have own properties
   const date = new Date();
   const object = {};
   const fakeDate = {};
   Object.setPrototypeOf(fakeDate, Date.prototype);

   // Different [[Prototype]]:
   assert.deepStrictEqual(object, fakeDate);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + {}
   // - Date {}

   // Different type tags:
   assert.deepStrictEqual(date, fakeDate);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + 2018-04-26T00:49:08.604Z
   // - Date {}

   assert.deepStrictEqual(NaN, NaN);
   // OK because Object.is(NaN, NaN) is true.

   // Different unwrapped numbers:
   assert.deepStrictEqual(new Number(1), new Number(2));
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + [Number: 1]
   // - [Number: 2]

   assert.deepStrictEqual(new String('foo'), Object('foo'));
   // OK because the object and the string are identical when unwrapped.

   assert.deepStrictEqual(-0, -0);
   // OK

   // Different zeros:
   assert.deepStrictEqual(0, -0);
   // AssertionError: Expected inputs to be strictly deep-equal:
   // + actual - expected
   //
   // + 0
   // - -0

   const symbol1 = Symbol();
   const symbol2 = Symbol();
   assert.deepStrictEqual({ [symbol1]: 1 }, { [symbol1]: 1 });
   // OK, because it is the same symbol on both objects.

   assert.deepStrictEqual({ [symbol1]: 1 }, { [symbol2]: 1 });
   // AssertionError [ERR_ASSERTION]: Inputs identical but not reference equal:
   //
   // {
   //   Symbol(): 1
   // }

   const weakMap1 = new WeakMap();
   const weakMap2 = new WeakMap();
   const obj = {};

   weakMap1.set(obj, 'value');
   weakMap2.set(obj, 'value');

   // Comparing different instances fails, even with same contents
   assert.deepStrictEqual(weakMap1, weakMap2);
   // AssertionError: Values have same structure but are not reference-equal:
   //
   // WeakMap {
   //   <items unknown>
   // }

   // Comparing the same instance to itself succeeds
   assert.deepStrictEqual(weakMap1, weakMap1);
   // OK

   const weakSet1 = new WeakSet();
   const weakSet2 = new WeakSet();
   weakSet1.add(obj);
   weakSet2.add(obj);

   // Comparing different instances fails, even with same contents
   assert.deepStrictEqual(weakSet1, weakSet2);
   // AssertionError: Values have same structure but are not reference-equal:
   // + actual - expected
   //
   // WeakSet {
   //   <items unknown>
   // }

   // Comparing the same instance to itself succeeds
   assert.deepStrictEqual(weakSet1, weakSet1);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdeepstrictequalactual-expected-message>`__

Assert.doesNotMatch
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotMatch(string, regexp[, message])

*Kind:* function · *Arguments:* 3

Expects the ``string`` input not to match the regular expression.

If the values do match, or if the ``string`` argument is of another type than ``string``, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``string`` (*string*)
* ``regexp`` (*RegExp*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.doesNotMatch('I will fail', /fail/);
   // AssertionError [ERR_ASSERTION]: The input was expected to not match the ...

   assert.doesNotMatch(123, /pass/);
   // AssertionError [ERR_ASSERTION]: The "string" argument must be of type string.

   assert.doesNotMatch('I will pass', /different/);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotmatchstring-regexp-message>`__

Assert.doesNotReject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotReject(asyncFn[, error][, message])

*Kind:* function · *Arguments:* 3

Awaits the ``asyncFn`` promise or, if ``asyncFn`` is a function, immediately calls the function and awaits the returned promise to complete. It will then check that the promise is not rejected.

If ``asyncFn`` is a function and it throws an error synchronously, ``assert.doesNotReject()`` will return a rejected ``Promise`` with that error. If the function does not return a promise, ``assert.doesNotReject()`` will return a rejected ``Promise`` with an ``ERR_INVALID_RETURN_VALUE`` error. In both cases the error handler is skipped.

**Arguments**

* ``asyncFn`` (*Function|Promise*)
* ``error`` (*RegExp|Function*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   (async () => {
     await assert.doesNotReject(
       async () => {
         throw new TypeError('Wrong value');
       },
       SyntaxError,
     );
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotrejectasyncfn-error-message>`__

Assert.doesNotThrow
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.doesNotThrow(fn[, error][, message])

*Kind:* function · *Arguments:* 3

Asserts that the function ``fn`` does not throw an error.

Using ``assert.doesNotThrow()`` is actually not useful because there is no benefit in catching an error and then rethrowing it. Instead, consider adding a comment next to the specific code path that should not throw and keep error messages as expressive as possible.

**Arguments**

* ``fn`` (*Function*)
* ``error`` (*RegExp|Function*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.doesNotThrow(
     () => {
       throw new TypeError('Wrong value');
     },
     SyntaxError,
   );

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertdoesnotthrowfn-error-message>`__

Assert.equal
~~~~~~~~~~~~

.. code-block:: text

   assert.equal(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.strictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.equal(1, 1);
   // OK, 1 == 1
   assert.equal(1, '1');
   // OK, 1 == '1'
   assert.equal(NaN, NaN);
   // OK

   assert.equal(1, 2);
   // AssertionError: 1 == 2
   assert.equal({ a: { b: 1 } }, { a: { b: 1 } });
   // AssertionError: { a: { b: 1 } } == { a: { b: 1 } }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertequalactual-expected-message>`__

Assert.fail
~~~~~~~~~~~

.. code-block:: text

   assert.fail(actual, expected[, message[, operator[, stackStartFn]]])

*Kind:* function · *Arguments:* 5

If ``message`` is falsy, the error message is set as the values of ``actual`` and ``expected`` separated by the provided ``operator``. If just the two ``actual`` and ``expected`` arguments are provided, ``operator`` will default to ``'!='``. If ``message`` is provided as third argument it will be used as the error message and the other arguments will be stored as properties on the thrown object. If ``stackStartFn`` is provided, all stack frames above that function will be removed from stacktrace (see ``Error.captureStackTrace``). If no arguments are given, the default message ``Failed`` will be used.

In the last three cases ``actual``, ``expected``, and ``operator`` have no influence on the error message.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional
* ``operator`` (*string*) — optional: **Default:** ``'!='``
* ``stackStartFn`` (*Function*) — optional: **Default:** ``assert.fail``

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.fail('a', 'b');
   // AssertionError [ERR_ASSERTION]: 'a' != 'b'

   assert.fail(1, 2, undefined, '>');
   // AssertionError [ERR_ASSERTION]: 1 > 2

   assert.fail(1, 2, 'fail');
   // AssertionError [ERR_ASSERTION]: fail

   assert.fail(1, 2, 'whoops', '>');
   // AssertionError [ERR_ASSERTION]: whoops

   assert.fail(1, 2, new TypeError('need array'));
   // TypeError: need array

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertfailactual-expected-message-operator-stackstartfn>`__

Assert.ifError
~~~~~~~~~~~~~~

.. code-block:: text

   assert.ifError(value)

*Kind:* function · *Arguments:* 1

Throws ``value`` if ``value`` is not ``undefined`` or ``null``. This is useful when testing the ``error`` argument in callbacks. The stack trace contains all frames from the error passed to ``ifError()`` including the potential new frames for ``ifError()`` itself.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.ifError(null);
   // OK
   assert.ifError(0);
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: 0
   assert.ifError('error');
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: 'error'
   assert.ifError(new Error());
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: Error

   // Create some random error frames.
   let err;
   (function errorFrame() {
     err = new Error('test error');
   })();

   (function ifErrorFrame() {
     assert.ifError(err);
   })();
   // AssertionError [ERR_ASSERTION]: ifError got unwanted exception: test error
   //     at ifErrorFrame
   //     at errorFrame

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertiferrorvalue>`__

Assert.match
~~~~~~~~~~~~

.. code-block:: text

   assert.match(string, regexp[, message])

*Kind:* function · *Arguments:* 3

Expects the ``string`` input to match the regular expression.

If the values do not match, or if the ``string`` argument is of another type than ``string``, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``string`` (*string*)
* ``regexp`` (*RegExp*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.match('I will fail', /pass/);
   // AssertionError [ERR_ASSERTION]: The input did not match the regular ...

   assert.match(123, /pass/);
   // AssertionError [ERR_ASSERTION]: The "string" argument must be of type string.

   assert.match('I will pass', /pass/);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertmatchstring-regexp-message>`__

Assert.notDeepEqual
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.notDeepEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.notDeepStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const obj1 = {
     a: {
       b: 1,
     },
   };
   const obj2 = {
     a: {
       b: 2,
     },
   };
   const obj3 = {
     a: {
       b: 1,
     },
   };
   const obj4 = { __proto__: obj1 };

   assert.notDeepEqual(obj1, obj1);
   // AssertionError: { a: { b: 1 } } notDeepEqual { a: { b: 1 } }

   assert.notDeepEqual(obj1, obj2);
   // OK

   assert.notDeepEqual(obj1, obj3);
   // AssertionError: { a: { b: 1 } } notDeepEqual { a: { b: 1 } }

   assert.notDeepEqual(obj1, obj4);
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotdeepequalactual-expected-message>`__

Assert.notDeepStrictEqual
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.notDeepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for deep strict inequality. Opposite of ``assert.deepStrictEqual()``.

If the values are deeply and strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.notDeepStrictEqual({ a: 1 }, { a: '1' });
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotdeepstrictequalactual-expected-message>`__

Assert.notEqual
~~~~~~~~~~~~~~~

.. code-block:: text

   assert.notEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

**Strict assertion mode**

An alias of ``assert.notStrictEqual()``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.notEqual(1, 2);
   // OK

   assert.notEqual(1, 1);
   // AssertionError: 1 != 1

   assert.notEqual(1, '1');
   // AssertionError: 1 != '1'

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotequalactual-expected-message>`__

Assert.notStrictEqual
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.notStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests strict inequality between the ``actual`` and ``expected`` parameters as determined by ``Object.is()``.

If the values are strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.notStrictEqual(1, 2);
   // OK

   assert.notStrictEqual(1, 1);
   // AssertionError [ERR_ASSERTION]: Expected "actual" to be strictly unequal to:
   //
   // 1

   assert.notStrictEqual(1, '1');
   // OK

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertnotstrictequalactual-expected-message>`__

Assert.ok
~~~~~~~~~

.. code-block:: text

   assert.ok(value[, message])

*Kind:* function · *Arguments:* 2

Tests if ``value`` is truthy. It is equivalent to ``assert.equal(!!value, true, message)``.

If ``value`` is not truthy, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is ``undefined``, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``. If no arguments are passed in at all ``message`` will be set to the string: 'No value argument passed to ``assert.ok()``'.

**Arguments**

* ``value`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.ok(true);
   // OK
   assert.ok(1);
   // OK

   assert.ok();
   // AssertionError: No value argument passed to `assert.ok()`

   assert.ok(false, 'it\'s false');
   // AssertionError: it's false

   // In the repl:
   assert.ok(typeof 123 === 'string');
   // AssertionError: false == true

   // In a file (e.g. test.js):
   assert.ok(typeof 123 === 'string');
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(typeof 123 === 'string')

   assert.ok(false);
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(false)

   assert.ok(0);
   // AssertionError: The expression evaluated to a falsy value:
   //
   //   assert.ok(0)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertokvalue-message>`__

Assert.partialDeepStrictEqual
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.partialDeepStrictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests for partial deep equality between the ``actual`` and ``expected`` parameters. "Deep" equality means that the enumerable "own" properties of child objects are recursively evaluated also by the following rules. "Partial" equality means that only properties that exist on the ``expected`` parameter are going to be compared.

This method always passes the same test cases as ``assert.deepStrictEqual()``, behaving as a super set of it.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   assert.partialDeepStrictEqual(
     { a: { b: { c: 1 } } },
     { a: { b: { c: 1 } } },
   );
   // OK

   assert.partialDeepStrictEqual(
     { a: 1, b: 2, c: 3 },
     { b: 2 },
   );
   // OK

   assert.partialDeepStrictEqual(
     [1, 2, 3, 4, 5, 6, 7, 8, 9],
     [4, 5, 8],
   );
   // OK

   assert.partialDeepStrictEqual(
     new Set([{ a: 1 }, { b: 1 }]),
     new Set([{ a: 1 }]),
   );
   // OK

   assert.partialDeepStrictEqual(
     new Map([['key1', 'value1'], ['key2', 'value2']]),
     new Map([['key2', 'value2']]),
   );
   // OK

   assert.partialDeepStrictEqual(123n, 123n);
   // OK

   assert.partialDeepStrictEqual(
     [1, 2, 3, 4, 5, 6, 7, 8, 9],
     [5, 4, 8],
   );
   // AssertionError

   assert.partialDeepStrictEqual(
     { a: 1 },
     { a: 1, b: 2 },
   );
   // AssertionError

   assert.partialDeepStrictEqual(
     { a: { b: 2 } },
     { a: { b: '2' } },
   );
   // AssertionError

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertpartialdeepstrictequalactual-expected-message>`__

Assert.rejects
~~~~~~~~~~~~~~

.. code-block:: text

   assert.rejects(asyncFn[, error][, message])

*Kind:* function · *Arguments:* 3

Awaits the ``asyncFn`` promise or, if ``asyncFn`` is a function, immediately calls the function and awaits the returned promise to complete. It will then check that the promise is rejected.

If ``asyncFn`` is a function and it throws an error synchronously, ``assert.rejects()`` will return a rejected ``Promise`` with that error. If the function does not return a promise, ``assert.rejects()`` will return a rejected ``Promise`` with an ``ERR_INVALID_RETURN_VALUE`` error. In both cases the error handler is skipped.

**Arguments**

* ``asyncFn`` (*Function|Promise*)
* ``error`` (*RegExp|Function|Object|Error*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   (async () => {
     await assert.rejects(
       async () => {
         throw new TypeError('Wrong value');
       },
       {
         name: 'TypeError',
         message: 'Wrong value',
       },
     );
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertrejectsasyncfn-error-message>`__

Assert.strictEqual
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.strictEqual(actual, expected[, message])

*Kind:* function · *Arguments:* 3

Tests strict equality between the ``actual`` and ``expected`` parameters as determined by ``Object.is()``.

If the values are not strictly equal, an ``AssertionError`` is thrown with a ``message`` property set equal to the value of the ``message`` parameter. If the ``message`` parameter is undefined, a default error message is assigned. If the ``message`` parameter is an instance of ``Error`` then it will be thrown instead of the ``AssertionError``.

**Arguments**

* ``actual`` (*any*)
* ``expected`` (*any*)
* ``message`` (*string|Error*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   assert.strictEqual(1, 2);
   // AssertionError [ERR_ASSERTION]: Expected inputs to be strictly equal:
   //
   // 1 !== 2

   assert.strictEqual(1, 1);
   // OK

   assert.strictEqual('Hello foobar', 'Hello World!');
   // AssertionError [ERR_ASSERTION]: Expected inputs to be strictly equal:
   // + actual - expected
   //
   // + 'Hello foobar'
   // - 'Hello World!'
   //          ^

   const apples = 1;
   const oranges = 2;
   assert.strictEqual(apples, oranges, `apples ${apples} !== oranges ${oranges}`);
   // AssertionError [ERR_ASSERTION]: apples 1 !== oranges 2

   assert.strictEqual(1, '1', new TypeError('Inputs are not identical'));
   // TypeError: Inputs are not identical

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertstrictequalactual-expected-message>`__

Assert.throws
~~~~~~~~~~~~~

.. code-block:: text

   assert.throws(fn[, error][, message])

*Kind:* function · *Arguments:* 3

Expects the function ``fn`` to throw an error.

If specified, ``error`` can be a ``Class``, ``RegExp``, a validation function, a validation object where each property will be tested for strict deep equality, or an instance of error where each property will be tested for strict deep equality including the non-enumerable ``message`` and ``name`` properties. When using an object, it is also possible to use a regular expression, when validating against a string property. See below for examples.

**Arguments**

* ``fn`` (*Function*)
* ``error`` (*RegExp|Function|Object|Error*) — optional
* ``message`` (*string*) — optional

**Example**

.. code-block:: javascript

   const assert = require('node:assert/strict');

   const err = new TypeError('Wrong value');
   err.code = 404;
   err.foo = 'bar';
   err.info = {
     nested: true,
     baz: 'text',
   };
   err.reg = /abc/i;

   assert.throws(
     () => {
       throw err;
     },
     {
       name: 'TypeError',
       message: 'Wrong value',
       info: {
         nested: true,
         baz: 'text',
       },
       // Only properties on the validation object will be tested for.
       // Using nested objects requires all properties to be present. Otherwise
       // the validation is going to fail.
     },
   );

   // Using regular expressions to validate error properties:
   assert.throws(
     () => {
       throw err;
     },
     {
       // The `name` and `message` properties are strings and using regular
       // expressions on those will match against the string. If they fail, an
       // error is thrown.
       name: /^TypeError$/,
       message: /Wrong/,
       foo: 'bar',
       info: {
         nested: true,
         // It is not possible to use regular expressions for nested properties!
         baz: 'text',
       },
       // The `reg` property contains a regular expression and only if the
       // validation object contains an identical regular expression, it is going
       // to pass.
       reg: /abc/i,
     },
   );

   // Fails due to the different `message` and `name` properties:
   assert.throws(
     () => {
       const otherErr = new Error('Not found');
       // Copy all enumerable properties from `err` to `otherErr`.
       for (const [key, value] of Object.entries(err)) {
         otherErr[key] = value;
       }
       throw otherErr;
     },
     // The error's `message` and `name` properties will also be checked when using
     // an error as validation object.
     err,
   );

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#assertthrowsfn-error-message>`__

AssertionError class
--------------------

**AssertionError** — Indicates the failure of an assertion. All errors thrown by the ``node:assert`` module will be instances of the ``AssertionError`` class.

AssertionError.toString
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.AssertionError.toString(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


CallTracker class
-----------------

**CallTracker** — This feature is deprecated and will be removed in a future version. Please consider using alternatives such as the ``mock`` helper function.

CallTracker._getTracked
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   assert.CallTracker._getTracked(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


CallTracker.calls
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tracker.calls([fn][, exact])

*Kind:* function · *Arguments:* 2

The wrapper function is expected to be called exactly ``exact`` times. If the function has not been called exactly ``exact`` times when ``tracker.verify()`` is called, then ``tracker.verify()`` will throw an error.

**Arguments**

* ``fn`` (*Function*) — optional: **Default:** A no-op function.
* ``exact`` (*number*) — optional: **Default:** ``1``.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   // Creates call tracker.
   const tracker = new assert.CallTracker();

   function func() {}

   // Returns a function that wraps func() that must be called exact times
   // before tracker.verify().
   const callsfunc = tracker.calls(func);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#trackercallsfn-exact>`__

CallTracker.getCalls
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tracker.getCalls(fn)

*Kind:* function · *Arguments:* 1

**Arguments**

* ``fn`` (*Function*)

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   // Creates call tracker.
   const tracker = new assert.CallTracker();

   function func() {}
   const callsfunc = tracker.calls(func);
   callsfunc(1, 2, 3);

   assert.deepStrictEqual(tracker.getCalls(callsfunc),
                          [{ thisArg: undefined, arguments: [1, 2, 3] }]);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#trackergetcallsfn>`__

CallTracker.report
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tracker.report()

*Kind:* function · *Arguments:* 0

The arrays contains information about the expected and actual number of calls of the functions that have not been called the expected number of times.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   // Creates call tracker.
   const tracker = new assert.CallTracker();

   function func() {}

   // Returns a function that wraps func() that must be called exact times
   // before tracker.verify().
   const callsfunc = tracker.calls(func, 2);

   // Returns an array containing information on callsfunc()
   console.log(tracker.report());
   // [
   //  {
   //    message: 'Expected the func function to be executed 2 time(s) but was
   //    executed 0 time(s).',
   //    actual: 0,
   //    expected: 2,
   //    operator: 'func',
   //    stack: stack trace
   //  }
   // ]

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#trackerreport>`__

CallTracker.reset
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tracker.reset([fn])

*Kind:* function · *Arguments:* 1

Reset calls of the call tracker. If a tracked function is passed as an argument, the calls will be reset for it. If no arguments are passed, all tracked functions will be reset.

**Arguments**

* ``fn`` (*Function*) — optional: a tracked function to reset.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const tracker = new assert.CallTracker();

   function func() {}
   const callsfunc = tracker.calls(func);

   callsfunc();
   // Tracker was called once
   assert.strictEqual(tracker.getCalls(callsfunc).length, 1);

   tracker.reset(callsfunc);
   assert.strictEqual(tracker.getCalls(callsfunc).length, 0);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#trackerresetfn>`__

CallTracker.verify
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tracker.verify()

*Kind:* function · *Arguments:* 0

Iterates through the list of functions passed to ``tracker.calls()`` and will throw an error for functions that have not been called the expected number of times.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   // Creates call tracker.
   const tracker = new assert.CallTracker();

   function func() {}

   // Returns a function that wraps func() that must be called exact times
   // before tracker.verify().
   const callsfunc = tracker.calls(func, 2);

   callsfunc();

   // Will throw an error since callsfunc() was only called once.
   tracker.verify();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/assert.html#trackerverify>`__

Properties and constants
------------------------

* ``length`` — number
* ``name`` — string

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
