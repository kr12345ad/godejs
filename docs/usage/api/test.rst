test (node:test)
================

The ``node:test`` module facilitates the creation of JavaScript tests. To access it:

This module is only available under the ``node:`` scheme.

Loading
-------

.. code-block:: javascript

   const test = require('node:test');
   // ESM
   import * as test from 'node:test';

Quick start
-----------

.. code-block:: javascript

   const test = require('node:test');

API
---

after
~~~~~

.. code-block:: text

   after([fn][, options])

*Kind:* function · *Arguments:* 2 · *Node test harness*

This function creates a hook that runs after executing a suite.

**Note:** The ``after`` hook is guaranteed to run, even if tests within the suite fail.

**Arguments**

* ``fn`` (*Function|AsyncFunction*) — optional: The hook function.
* ``options`` (*Object*) — optional: Configuration options for the hook. The following

**Example**

.. code-block:: javascript

   describe('tests', async () => {
     after(() => console.log('finished running tests'));
     it('is a subtest', () => {
       // Some relevant assertion here
     });
   });


afterEach
~~~~~~~~~

.. code-block:: text

   afterEach([fn][, options])

*Kind:* function · *Arguments:* 2 · *Node test harness*

This function creates a hook that runs after each test in the current suite. The ``afterEach()`` hook is run even if the test fails.

**Arguments**

* ``fn`` (*Function|AsyncFunction*) — optional: The hook function.
* ``options`` (*Object*) — optional: Configuration options for the hook. The following

**Example**

.. code-block:: javascript

   describe('tests', async () => {
     afterEach(() => console.log('finished running a test'));
     it('is a subtest', () => {
       // Some relevant assertion here
     });
   });


assert
~~~~~~

.. code-block:: text

   assert

*Kind:* function · *Arguments:* 2 · *Node test harness*

An object whose methods are used to configure available assertions on the ``TestContext`` objects in the current process. The methods from ``node:assert`` and snapshot testing functions are available by default.

It is possible to apply the same configuration to all files by placing common configuration code in a module preloaded with ``--require`` or ``--import``.

This callable takes no arguments.


before
~~~~~~

.. code-block:: text

   before([fn][, options])

*Kind:* function · *Arguments:* 2 · *Node test harness*

This function creates a hook that runs before executing a suite.

**Arguments**

* ``fn`` (*Function|AsyncFunction*) — optional: The hook function.
* ``options`` (*Object*) — optional: Configuration options for the hook. The following

**Example**

.. code-block:: javascript

   describe('tests', async () => {
     before(() => console.log('about to run some test'));
     it('is a subtest', () => {
       // Some relevant assertions here
     });
   });


beforeEach
~~~~~~~~~~

.. code-block:: text

   beforeEach([fn][, options])

*Kind:* function · *Arguments:* 2 · *Node test harness*

This function creates a hook that runs before each test in the current suite.

**Arguments**

* ``fn`` (*Function|AsyncFunction*) — optional: The hook function.
* ``options`` (*Object*) — optional: Configuration options for the hook. The following

**Example**

.. code-block:: javascript

   describe('tests', async () => {
     beforeEach(() => console.log('about to run a test'));
     it('is a subtest', () => {
       // Some relevant assertion here
     });
   });


describe
~~~~~~~~

.. code-block:: text

   describe()

*Kind:* function · *Arguments:* 3 · *Node test harness*

Suites and tests can also be written using the ``describe()`` and ``it()`` functions. ``describe()`` is an alias for ``suite()``, and ``it()`` is an alias for ``test()``.

``describe()`` and ``it()`` are imported from the ``node:test`` module.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { describe, it } = require('node:test');


it
~~~

.. code-block:: text

   it([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

Alias for ``test()``.

The ``it()`` function is imported from the ``node:test`` module.

**Arguments**

* ``name`` — optional
* ``options`` — optional
* ``fn`` — optional


only
~~~~

.. code-block:: text

   test.only([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

Shorthand for marking a test as ``only``, same as ``test([name], { only: true }[, fn])``.

**Arguments**

* ``name`` — optional
* ``options`` — optional
* ``fn`` — optional


skip
~~~~

.. code-block:: text

   test.skip([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

Shorthand for skipping a test, same as ``test([name], { skip: true }[, fn])``.

**Arguments**

* ``name`` — optional
* ``options`` — optional
* ``fn`` — optional


suite
~~~~~

.. code-block:: text

   suite([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

The ``suite()`` function is imported from the ``node:test`` module.

**Arguments**

* ``name`` (*string*) — optional: The name of the suite, which is displayed when reporting test
* ``options`` (*Object*) — optional: Optional configuration options for the suite.
* ``fn`` (*Function|AsyncFunction*) — optional: The suite function declaring nested tests and


test
~~~~

.. code-block:: text

   test([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

The ``test()`` function is the value imported from the ``test`` module. Each invocation of this function results in reporting the test to the ``TestsStream``.

The ``TestContext`` object passed to the ``fn`` argument can be used to perform actions related to the current test. Examples include skipping the test, adding additional diagnostic information, or creating subtests.

**Arguments**

* ``name`` (*string*) — optional: The name of the test, which is displayed when reporting test
* ``options`` (*Object*) — optional: Configuration options for the test. The following
* ``fn`` (*Function|AsyncFunction*) — optional: The function under test. The first argument

**Example**

.. code-block:: javascript

   test('top level test', async (t) => {
     // The setTimeout() in the following subtest would cause it to outlive its
     // parent test if 'await' is removed on the next line. Once the parent test
     // completes, it will cancel any outstanding subtests.
     await t.test('longer running subtest', async (t) => {
       return new Promise((resolve, reject) => {
         setTimeout(resolve, 1000);
       });
     });
   });


todo
~~~~

.. code-block:: text

   test.todo([name][, options][, fn])

*Kind:* function · *Arguments:* 3 · *Node test harness*

Shorthand for marking a test as ``TODO``, same as ``test([name], { todo: true }[, fn])``.

**Arguments**

* ``name`` — optional
* ``options`` — optional
* ``fn`` — optional


mock
----

mock.fn
~~~~~~~

.. code-block:: text

   mock.fn([original[, implementation]][, options])

*Kind:* function · *Arguments:* 3 · *Node test harness*

This function is used to create a mock function.

The following example creates a mock function that increments a counter by one on each invocation. The ``times`` option is used to modify the mock behavior such that the first two invocations add two to the counter instead of one.

**Arguments**

* ``original`` (*Function|AsyncFunction*) — optional: An optional function to create a mock on.
* ``implementation`` (*Function|AsyncFunction*) — optional: An optional function used as the
* ``options`` (*Object*) — optional: Optional configuration options for the mock function. The

**Example**

.. code-block:: javascript

   test('mocks a counting function', (t) => {
     let cnt = 0;

     function addOne() {
       cnt++;
       return cnt;
     }

     function addTwo() {
       cnt += 2;
       return cnt;
     }

     const fn = t.mock.fn(addOne, addTwo, { times: 2 });

     assert.strictEqual(fn(), 2);
     assert.strictEqual(fn(), 4);
     assert.strictEqual(fn(), 5);
     assert.strictEqual(fn(), 6);
   });


mock.getter
~~~~~~~~~~~

.. code-block:: text

   mock.getter(object, methodName[, implementation][, options])

*Kind:* function · *Arguments:* 4 · *Node test harness*

This function is syntax sugar for ``MockTracker.method`` with ``options.getter`` set to ``true``.

**Arguments**

* ``object``
* ``methodName``
* ``implementation`` — optional
* ``options`` — optional


mock.method
~~~~~~~~~~~

.. code-block:: text

   mock.method(object, methodName[, implementation][, options])

*Kind:* function · *Arguments:* 4 · *Node test harness*

This function is used to create a mock on an existing object method. The following example demonstrates how a mock is created on an existing object method.

**Arguments**

* ``object`` (*Object*): The object whose method is being mocked.
* ``methodName`` (*string|symbol*): The identifier of the method on ``object`` to mock.
* ``implementation`` (*Function|AsyncFunction*) — optional: An optional function used as the
* ``options`` (*Object*) — optional: Optional configuration options for the mock method. The

**Example**

.. code-block:: javascript

   test('spies on an object method', (t) => {
     const number = {
       value: 5,
       subtract(a) {
         return this.value - a;
       },
     };

     t.mock.method(number, 'subtract');
     assert.strictEqual(number.subtract.mock.callCount(), 0);
     assert.strictEqual(number.subtract(3), 2);
     assert.strictEqual(number.subtract.mock.callCount(), 1);

     const call = number.subtract.mock.calls[0];

     assert.deepStrictEqual(call.arguments, [3]);
     assert.strictEqual(call.result, 2);
     assert.strictEqual(call.error, undefined);
     assert.strictEqual(call.target, undefined);
     assert.strictEqual(call.this, number);
   });


mock.property
~~~~~~~~~~~~~

.. code-block:: text

   mock.property(object, propertyName[, value])

*Kind:* function · *Arguments:* 3 · *Node test harness*

Creates a mock for a property value on an object. This allows you to track and control access to a specific property, including how many times it is read (getter) or written (setter), and to restore the original value after mocking.

**Arguments**

* ``object`` (*Object*): The object whose value is being mocked.
* ``propertyName`` (*string|symbol*): The identifier of the property on ``object`` to mock.
* ``value`` (*any*) — optional: An optional value used as the mock value

**Example**

.. code-block:: javascript

   test('mocks a property value', (t) => {
     const obj = { foo: 42 };
     const prop = t.mock.property(obj, 'foo', 100);

     assert.strictEqual(obj.foo, 100);
     assert.strictEqual(prop.mock.accessCount(), 1);
     assert.strictEqual(prop.mock.accesses[0].type, 'get');
     assert.strictEqual(prop.mock.accesses[0].value, 100);

     obj.foo = 200;
     assert.strictEqual(prop.mock.accessCount(), 2);
     assert.strictEqual(prop.mock.accesses[1].type, 'set');
     assert.strictEqual(prop.mock.accesses[1].value, 200);

     prop.mock.restore();
     assert.strictEqual(obj.foo, 42);
   });


mock.reset
~~~~~~~~~~

.. code-block:: text

   mock.reset()

*Kind:* function · *Arguments:* 0 · *Node test harness*

This function restores the default behavior of all mocks that were previously created by this ``MockTracker`` and disassociates the mocks from the ``MockTracker`` instance. Once disassociated, the mocks can still be used, but the ``MockTracker`` instance can no longer be used to reset their behavior or otherwise interact with them.

After each test completes, this function is called on the test context's ``MockTracker``. If the global ``MockTracker`` is used extensively, calling this function manually is recommended.

This callable takes no arguments.


mock.restoreAll
~~~~~~~~~~~~~~~

.. code-block:: text

   mock.restoreAll()

*Kind:* function · *Arguments:* 0 · *Node test harness*

This function restores the default behavior of all mocks that were previously created by this ``MockTracker``. Unlike ``mock.reset()``, ``mock.restoreAll()`` does not disassociate the mocks from the ``MockTracker`` instance.

This callable takes no arguments.


mock.setter
~~~~~~~~~~~

.. code-block:: text

   mock.setter(object, methodName[, implementation][, options])

*Kind:* function · *Arguments:* 4 · *Node test harness*

This function is syntax sugar for ``MockTracker.method`` with ``options.setter`` set to ``true``.

**Arguments**

* ``object``
* ``methodName``
* ``implementation`` — optional
* ``options`` — optional


mock.timers.enable
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   timers.enable([enableOptions])

*Kind:* function · *Arguments:* 1 · *Node test harness*

Enables timer mocking for the specified timers.

**Note:** When you enable mocking for a specific timer, its associated clear function will also be implicitly mocked.

**Arguments**

* ``enableOptions`` (*Object*) — optional: Optional configuration options for enabling timer

**Example**

.. code-block:: javascript

   const { mock } = require('node:test');
   mock.timers.enable({ apis: ['setInterval'] });


mock.timers.enableFor
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   test.mock.timers.enableFor(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *Node test harness*

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


mock.timers.get
~~~~~~~~~~~~~~~

.. code-block:: text

   test.mock.timers.get(...args)

*Kind:* function · *Arguments:* 0 · *Node test harness*

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


mock.timers.reset
~~~~~~~~~~~~~~~~~

.. code-block:: text

   timers.reset()

*Kind:* function · *Arguments:* 0 · *Node test harness*

This function restores the default behavior of all mocks that were previously created by this  ``MockTimers`` instance and disassociates the mocks from the  ``MockTracker`` instance.

**Note:** After each test completes, this function is called on the test context's  ``MockTracker``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { mock } = require('node:test');
   mock.timers.reset();


mock.timers.runAll
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   timers.runAll()

*Kind:* function · *Arguments:* 0 · *Node test harness*

Triggers all pending mocked timers immediately. If the ``Date`` object is also mocked, it will also advance the ``Date`` object to the furthest timer's time.

The example below triggers all pending timers immediately, causing them to execute without any delay.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');
   const { test } = require('node:test');

   test('runAll functions following the given order', (context) => {
     context.mock.timers.enable({ apis: ['setTimeout', 'Date'] });
     const results = [];
     setTimeout(() => results.push(1), 9999);

     // Notice that if both timers have the same timeout,
     // the order of execution is guaranteed
     setTimeout(() => results.push(3), 8888);
     setTimeout(() => results.push(2), 8888);

     assert.deepStrictEqual(results, []);

     context.mock.timers.runAll();
     assert.deepStrictEqual(results, [3, 2, 1]);
     // The Date object is also advanced to the furthest timer's time
     assert.strictEqual(Date.now(), 9999);
   });


mock.timers.setTime
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   timers.setTime(milliseconds)

*Kind:* function · *Arguments:* 1 · *Node test harness*

Sets the current Unix timestamp that will be used as reference for any mocked ``Date`` objects.

**Arguments**

* ``milliseconds``

**Example**

.. code-block:: javascript

   const assert = require('node:assert');
   const { test } = require('node:test');

   test('setTime replaces current time', (context) => {
     const now = Date.now();
     const setTime = 1000;
     // Date.now is not mocked
     assert.deepStrictEqual(Date.now(), now);

     context.mock.timers.enable({ apis: ['Date'] });
     context.mock.timers.setTime(setTime);
     // Date.now is now 1000
     assert.strictEqual(Date.now(), setTime);
   });


mock.timers.tick
~~~~~~~~~~~~~~~~

.. code-block:: text

   timers.tick([milliseconds])

*Kind:* function · *Arguments:* 1 · *Node test harness*

Advances time for all mocked timers.

**Note:** This diverges from how ``setTimeout`` in Node.js behaves and accepts only positive numbers. In Node.js, ``setTimeout`` with negative numbers is only supported for web compatibility reasons.

**Arguments**

* ``milliseconds`` (*number*) — optional: The amount of time, in milliseconds,

**Example**

.. code-block:: javascript

   const assert = require('node:assert');
   const { test } = require('node:test');

   test('mocks setTimeout to be executed synchronously without having to actually wait for it', (context) => {
     const fn = context.mock.fn();
     context.mock.timers.enable({ apis: ['setTimeout'] });

     setTimeout(fn, 9999);
     assert.strictEqual(fn.mock.callCount(), 0);

     // Advance in time
     context.mock.timers.tick(9999);

     assert.strictEqual(fn.mock.callCount(), 1);
   });


Properties and constants
------------------------

* ``length`` — number
* ``mock`` — object
* ``mock.timers`` — object
* ``name`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
