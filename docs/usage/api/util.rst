util
====

The ``node:util`` module supports the needs of Node.js internal APIs. Many of the utilities are useful for application and module developers as well. To access it:

Loading
-------

.. code-block:: javascript

   const util = require('node:util');
   // ESM
   import * as util from 'node:util';

Quick start
-----------

.. code-block:: javascript

   const util = require('node:util');

API
---

__rawStack
~~~~~~~~~~

.. code-block:: text

   util.__rawStack(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const util = require('node:util');

   console.log(typeof util.__rawStack); // 'function'
   console.log(util.__rawStack.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_exceptionWithHostPort
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util._exceptionWithHostPort(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const util = require('node:util');

   console.log(typeof util._exceptionWithHostPort); // 'function'
   console.log(util._exceptionWithHostPort.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_extend
~~~~~~~

.. code-block:: text

   util._extend()

*Kind:* function · *Arguments:* 0

Type: Runtime

The ``util._extend()`` API is deprecated because it's an unmaintained legacy API that was exposed to user land by accident. Please use ``target = Object.assign(target, source)`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util._extend();
   console.log(result);


aborted
~~~~~~~

.. code-block:: text

   util.aborted(signal, resource)

*Kind:* function · *Arguments:* 2

Listens to abort event on the provided ``signal`` and returns a promise that resolves when the ``signal`` is aborted. If ``resource`` is provided, it weakly references the operation's associated object, so if ``resource`` is garbage collected before the ``signal`` aborts, then returned promise shall remain pending. This prevents memory leaks in long-running or non-cancelable operations.

**Arguments**

* ``signal`` (*AbortSignal*)
* ``resource`` (*Object*): Any non-null object tied to the abortable operation and held weakly.

**Example**

.. code-block:: javascript

   const { aborted } = require('node:util');

   // Obtain an object with an abortable signal, like a custom resource or operation.
   const dependent = obtainSomethingAbortable();

   // Pass `dependent` as the resource, indicating the promise should only resolve
   // if `dependent` is still in memory when the signal is aborted.
   aborted(dependent.signal, dependent).then(() => {

     // This code runs when `dependent` is aborted.
     console.log('Dependent resource was aborted.');
   });

   // Simulate an event that triggers the abort.
   dependent.on('event', () => {
     dependent.abort(); // This will cause the `aborted` promise to resolve.
   });


callbackify
~~~~~~~~~~~

.. code-block:: text

   util.callbackify(original)

*Kind:* function · *Arguments:* 1

Takes an ``async`` function (or a function that returns a ``Promise``) and returns a function following the error-first callback style, i.e. taking an ``(err, value) => ...`` callback as the last argument. In the callback, the first argument will be the rejection reason (or ``null`` if the ``Promise`` resolved), and the second argument will be the resolved value.

Will print:

**Arguments**

* ``original`` (*Function*): An ``async`` function

**Example**

.. code-block:: javascript

   const { callbackify } = require('node:util');

   async function fn() {
     return 'hello world';
   }
   const callbackFunction = callbackify(fn);

   callbackFunction((err, ret) => {
     if (err) throw err;
     console.log(ret);
   });


convertProcessSignalToExitCode
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.convertProcessSignalToExitCode(signal)

*Kind:* function · *Arguments:* 1

The ``util.convertProcessSignalToExitCode()`` method converts a signal name to its corresponding POSIX exit code. Following the POSIX standard, the exit code for a process terminated by a signal is calculated as ``128 + signal number``.

If ``signal`` is not a valid signal name, then an error will be thrown. See ``signal(7)`` for a list of valid signals.

**Arguments**

* ``signal`` (*string*): A signal name (e.g. ``'SIGTERM'``)

**Example**

.. code-block:: javascript

   const { convertProcessSignalToExitCode } = require('node:util');

   console.log(convertProcessSignalToExitCode('SIGTERM')); // 143 (128 + 15)
   console.log(convertProcessSignalToExitCode('SIGKILL')); // 137 (128 + 9)


debounce
~~~~~~~~

.. code-block:: text

   util.debounce(fn, wait[, options])

*Kind:* function · *Arguments:* 3

Creates a function that delays calling ``fn`` until ``wait`` milliseconds have elapsed since the most recent invocation. The debounced function returns a ``Promise`` for the value returned by ``fn``. If ``fn`` throws or returns a rejected promise, the returned promise is rejected with the same reason.

When the debounced function is called more than once before the delay expires, ``fn`` receives the arguments from the most recent call. By default, the promises from all calls resolve or reject with the result of that invocation. If ``options.rejectOnCancel`` is ``true``, the promises from superseded calls reject with an ``AbortError`` instead.

**Arguments**

* ``fn`` (*Function*): The function to debounce.
* ``wait`` (*integer*): The number of milliseconds to delay ``fn``.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { debounce } = require('node:util');

   const onInactivity = debounce(() => {
     console.log('No activity for 5 seconds');
   }, 5_000).unref();

   process.stdin.on('data', (data) => {
     console.log(`Received ${data.length} bytes`);
     onInactivity();
   });

   // Start the initial inactivity timeout.
   onInactivity();


debug
~~~~~

.. code-block:: text

   util.debug()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

``util.debug()`` has been removed. Please use ``console.error()`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.debug();
   console.log(result);


debuglog
~~~~~~~~

.. code-block:: text

   util.debuglog(section[, callback])

*Kind:* function · *Arguments:* 2

The ``util.debuglog()`` method is used to create a function that conditionally writes debug messages to ``stderr`` based on the existence of the ``NODE_DEBUG`` environment variable. If the ``section`` name appears within the value of that environment variable, then the returned function operates similar to ``console.error()``. If not, then the returned function is a no-op.

If this program is run with ``NODE_DEBUG=foo`` in the environment, then it will output something like:

**Arguments**

* ``section`` (*string*): A string identifying the portion of the application for
* ``callback`` (*Function*) — optional: A callback invoked the first time the logging function

**Example**

.. code-block:: javascript

   const { debuglog } = require('node:util');
   const log = debuglog('foo');

   log('hello from foo [%d]', 123);


deprecate
~~~~~~~~~

.. code-block:: text

   util.deprecate(fn, msg[, code[, options]])

*Kind:* function · *Arguments:* 4

The ``util.deprecate()`` method wraps ``fn`` (which may be a function or class) in such a way that it is marked as deprecated.

When called, ``util.deprecate()`` will return a function that will emit a ``DeprecationWarning`` using the ``'warning'`` event. The warning will be emitted and printed to ``stderr`` the first time the returned function is called. After the warning is emitted, the wrapped function is called without emitting a warning.

**Arguments**

* ``fn`` (*Function*): The function that is being deprecated.
* ``msg`` (*string*): A warning message to display when the deprecated function is
* ``code`` (*string*) — optional: A deprecation code. See the list of deprecated APIs for a
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { deprecate } = require('node:util');

   exports.obsoleteFunction = deprecate(() => {
     // Do something here.
   }, 'obsoleteFunction() is deprecated. Use newShinyFunction() instead.');


diff
~~~~

.. code-block:: text

   util.diff(actual, expected)

*Kind:* function · *Arguments:* 2

``util.diff()`` compares two string or array values and returns an array of difference entries. It uses the Myers diff algorithm to compute minimal differences, which is the same algorithm used internally by assertion error messages.

If the values are equal, an empty array is returned.

**Arguments**

* ``actual`` (*Array|string*): The first value to compare
* ``expected`` (*Array|string*): The second value to compare

**Example**

.. code-block:: javascript

   const { diff } = require('node:util');

   // Comparing strings
   const actualString = '12345678';
   const expectedString = '12!!5!7!';
   console.log(diff(actualString, expectedString));
   // [
   //   [0, '1'],
   //   [0, '2'],
   //   [1, '3'],
   //   [1, '4'],
   //   [-1, '!'],
   //   [-1, '!'],
   //   [0, '5'],
   //   [1, '6'],
   //   [-1, '!'],
   //   [0, '7'],
   //   [1, '8'],
   //   [-1, '!'],
   // ]
   // Comparing arrays
   const actualArray = ['1', '2', '3'];
   const expectedArray = ['1', '3', '4'];
   console.log(diff(actualArray, expectedArray));
   // [
   //   [0, '1'],
   //   [1, '2'],
   //   [0, '3'],
   //   [-1, '4'],
   // ]
   // Equal values return empty array
   console.log(diff('same', 'same'));
   // []


format
~~~~~~

.. code-block:: text

   util.format(format[, ...args])

*Kind:* function · *Arguments:* 2

The ``util.format()`` method returns a formatted string using the first argument as a ``printf``-like format string which can contain zero or more format specifiers. Each specifier is replaced with the converted value from the corresponding argument. Supported specifiers are:

* ``%s``: ``String`` will be used to convert all values except ``BigInt``, ``Object`` and ``-0``. ``BigInt`` values will be represented with an ``n`` and Objects that have neither a user defined ``toString`` function nor ``Symbol.toPrimitive`` function are inspected using ``util.inspect()`` with options ``{ depth: 0, colors: false, compact: 3 }``. * ``%d``: ``Number`` will be used to convert all values except ``BigInt`` and ``Symbol``. * ``%i``: ``parseInt(value, 10)`` is used for all values except ``BigInt`` and ``Symbol``. * ``%f``: ``parseFloat(value)`` is used for all values except ``Symbol``. * ``%j``: JSON. Replaced with the string ``'[Circular]'`` if the argument contains circular references. * ``%o``: ``Object``. A string representation of an object with generic JavaScript object formatting. Similar to ``util.inspect()`` with options ``{ showHidden: true, showProxy: true }``. This will show the full object including non-enumerable properties and proxies. * ``%O``: ``Object``. A string representation of an object with generic JavaScript object formatting. Similar to ``util.inspect()`` without options. This will show the full object not including non-enumerable properties and proxies. * ``%c``: ``CSS``. This specifier is ignored and will skip any CSS passed in. * ``%%``: single percent sign (``'%'``). This does not consume an argument. * Returns: ``string`` The formatted string

**Arguments**

* ``format`` (*string*): A ``printf``-like format string.
* ``...args`` — optional

**Example**

.. code-block:: javascript

   util.format('%s:%s', 'foo');
   // Returns: 'foo:%s'


formatWithOptions
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.formatWithOptions(inspectOptions, format[, ...args])

*Kind:* function · *Arguments:* 3

This function is identical to ``util.format()``, except in that it takes an ``inspectOptions`` argument which specifies options that are passed along to ``util.inspect()``.

**Arguments**

* ``inspectOptions`` (*Object*)
* ``format`` (*string*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   util.formatWithOptions({ colors: true }, 'See object %O', { foo: 42 });
   // Returns 'See object { foo: 42 }', where `42` is colored as a number
   // when printed to a terminal.


getCallSites
~~~~~~~~~~~~

.. code-block:: text

   util.getCallSites([frameCount][, options])

*Kind:* function · *Arguments:* 2

Returns an array of call site objects containing the stack of the caller function.

Unlike accessing an ``error.stack``, the result returned from this API is not interfered with ``Error.prepareStackTrace``.

**Arguments**

* ``frameCount`` (*integer*) — optional: Optional number of frames to capture as call site objects.
* ``options`` (*Object*) — optional: Optional

**Example**

.. code-block:: javascript

   const { getCallSites } = require('node:util');

   function exampleFunction() {
     const callSites = getCallSites();

     console.log('Call Sites:');
     callSites.forEach((callSite, index) => {
       console.log(`CallSite ${index + 1}:`);
       console.log(`Function Name: ${callSite.functionName}`);
       console.log(`Script Name: ${callSite.scriptName}`);
       console.log(`Line Number: ${callSite.lineNumber}`);
       console.log(`Column Number: ${callSite.columnNumber}`);
     });
     // CallSite 1:
     // Function Name: exampleFunction
     // Script Name: /home/example.js
     // Line Number: 5
     // Column Number: 26

     // CallSite 2:
     // Function Name: anotherFunction
     // Script Name: /home/example.js
     // Line Number: 22
     // Column Number: 3

     // ...
   }

   // A function to simulate another stack layer
   function anotherFunction() {
     exampleFunction();
   }

   anotherFunction();


getSystemErrorMap
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.getSystemErrorMap()

*Kind:* function · *Arguments:* 0

Returns a Map of all system error codes available from the Node.js API. The mapping between error codes and error names is platform-dependent. See Common System Errors for the names of common errors.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   fs.access('file/that/does/not/exist', (err) => {
     const errorMap = util.getSystemErrorMap();
     const name = errorMap.get(err.errno);
     console.error(name);  // ENOENT
   });


getSystemErrorMessage
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.getSystemErrorMessage(err)

*Kind:* function · *Arguments:* 1

Returns the string message for a numeric error code that comes from a Node.js API. The mapping between error codes and string messages is platform-dependent.

**Arguments**

* ``err`` (*number*)

**Example**

.. code-block:: javascript

   fs.access('file/that/does/not/exist', (err) => {
     const message = util.getSystemErrorMessage(err.errno);
     console.error(message);  // No such file or directory
   });


getSystemErrorName
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.getSystemErrorName(err)

*Kind:* function · *Arguments:* 1

Returns the string name for a numeric error code that comes from a Node.js API. The mapping between error codes and error names is platform-dependent. See Common System Errors for the names of common errors.

**Arguments**

* ``err`` (*number*)

**Example**

.. code-block:: javascript

   fs.access('file/that/does/not/exist', (err) => {
     const name = util.getSystemErrorName(err.errno);
     console.error(name);  // ENOENT
   });


inherits
~~~~~~~~

.. code-block:: text

   util.inherits(constructor, superConstructor)

*Kind:* function · *Arguments:* 2

Usage of ``util.inherits()`` is discouraged. Please use the ES6 ``class`` and ``extends`` keywords to get language level inheritance support. Also note that the two styles are semantically incompatible.

Inherit the prototype methods from one constructor into another. The prototype of ``constructor`` will be set to a new object created from ``superConstructor``.

**Arguments**

* ``constructor`` (*Function*)
* ``superConstructor`` (*Function*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');

   class MyStream extends EventEmitter {
     write(data) {
       this.emit('data', data);
     }
   }

   const stream = new MyStream();

   stream.on('data', (data) => {
     console.log(`Received data: "${data}"`);
   });
   stream.write('With ES6');


inspect
~~~~~~~

.. code-block:: text

   util.inspect(object[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``object``
* ``options`` — optional

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.inspect(undefined, {});
   console.log(result);


isArray
~~~~~~~

.. code-block:: text

   util.isArray()

*Kind:* function · *Arguments:* 0

Type: Runtime

The ``util.isArray()`` API is deprecated. Please use ``Array.isArray()`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isArray();
   console.log(result);


isBoolean
~~~~~~~~~

.. code-block:: text

   util.isBoolean()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isBoolean()`` API has been removed. Please use ``typeof arg === 'boolean'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isBoolean();
   console.log(result);


isBuffer
~~~~~~~~

.. code-block:: text

   util.isBuffer()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isBuffer()`` API has been removed. Please use ``Buffer.isBuffer()`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isBuffer();
   console.log(result);


isDate
~~~~~~

.. code-block:: text

   util.isDate()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isDate()`` API has been removed. Please use ``arg instanceof Date`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isDate();
   console.log(result);


isDeepStrictEqual
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.isDeepStrictEqual(val1, val2[, options])

*Kind:* function · *Arguments:* 3

Returns ``true`` if there is deep strict equality between ``val1`` and ``val2``. Otherwise, returns ``false``.

By default, deep strict equality includes comparison of object prototypes and constructors. When ``skipPrototype`` is ``true``, objects with different prototypes or constructors can still be considered equal if their enumerable properties are deeply strictly equal.

**Arguments**

* ``val1`` (*any*)
* ``val2`` (*any*)
* ``options`` — optional

**Example**

.. code-block:: javascript

   const util = require('node:util');

   class Foo {
     constructor(a) {
       this.a = a;
     }
   }

   class Bar {
     constructor(a) {
       this.a = a;
     }
   }

   const foo = new Foo(1);
   const bar = new Bar(1);

   // Different constructors, same properties
   console.log(util.isDeepStrictEqual(foo, bar));
   // false

   console.log(util.isDeepStrictEqual(foo, bar, true));
   // true


isError
~~~~~~~

.. code-block:: text

   util.isError()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isError()`` API has been removed. Please use ``Error.isError(arg)``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isError();
   console.log(result);


isFunction
~~~~~~~~~~

.. code-block:: text

   util.isFunction()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isFunction()`` API has been removed. Please use ``typeof arg === 'function'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isFunction();
   console.log(result);


isNull
~~~~~~

.. code-block:: text

   util.isNull()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isNull()`` API has been removed. Please use ``arg === null`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isNull();
   console.log(result);


isNullOrUndefined
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.isNullOrUndefined()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isNullOrUndefined()`` API has been removed. Please use ``arg === null || arg === undefined`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isNullOrUndefined();
   console.log(result);


isNumber
~~~~~~~~

.. code-block:: text

   util.isNumber()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isNumber()`` API has been removed. Please use ``typeof arg === 'number'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isNumber();
   console.log(result);


isObject
~~~~~~~~

.. code-block:: text

   util.isObject()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isObject()`` API has been removed. Please use ``arg && typeof arg === 'object'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isObject();
   console.log(result);


isPrimitive
~~~~~~~~~~~

.. code-block:: text

   util.isPrimitive()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isPrimitive()`` API has been removed. Please use ``Object(arg) !== arg`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isPrimitive();
   console.log(result);


isRegExp
~~~~~~~~

.. code-block:: text

   util.isRegExp()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isRegExp()`` API has been removed. Please use ``arg instanceof RegExp`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isRegExp();
   console.log(result);


isString
~~~~~~~~

.. code-block:: text

   util.isString()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isString()`` API has been removed. Please use ``typeof arg === 'string'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isString();
   console.log(result);


isSymbol
~~~~~~~~

.. code-block:: text

   util.isSymbol()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isSymbol()`` API has been removed. Please use ``typeof arg === 'symbol'`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isSymbol();
   console.log(result);


isUndefined
~~~~~~~~~~~

.. code-block:: text

   util.isUndefined()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``util.isUndefined()`` API has been removed. Please use ``arg === undefined`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.isUndefined();
   console.log(result);


markPromiseAsHandled
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.markPromiseAsHandled(promise)

*Kind:* function · *Arguments:* 1

Marks a promise as handled so that unhandled rejections are ignored and are not reported to the ``'unhandledrejection'`` event.

**Arguments**

* ``promise`` (*Promise*): The promise to mark as handled

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.markPromiseAsHandled(undefined);
   console.log(result);


MIMEParams
~~~~~~~~~~

.. code-block:: text

   util.MIMEParams

*Kind:* class · *Arguments:* 0

The ``MIMEParams`` API provides read and write access to the parameters of a ``MIMEType``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { MIMEParams } = require('node:util');

   const myParams = new MIMEParams();


MIMEType
~~~~~~~~

.. code-block:: text

   util.MIMEType

*Kind:* class · *Arguments:* 2

An implementation of \`the MIMEType class <https://bmeck.github.io/node-proposal-mime-api/>\`_.

In accordance with browser conventions, all properties of ``MIMEType`` objects are implemented as getters and setters on the class prototype, rather than as data properties on the object itself.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { MIMEType } = require('node:util');

   const myMIME = new MIMEType('text/plain');


parseArgs
~~~~~~~~~

.. code-block:: text

   util.parseArgs([config])

*Kind:* function · *Arguments:* 1

Provides a higher level API for command-line argument parsing than interacting with ``process.argv`` directly. Takes a specification for the expected arguments and returns a structured object with the parsed options and positionals.

**Arguments**

* ``config`` (*Object*) — optional: Used to provide arguments for parsing and to configure

**Example**

.. code-block:: javascript

   const { parseArgs } = require('node:util');
   const args = ['-f', '--bar', 'b'];
   const options = {
     foo: {
       type: 'boolean',
       short: 'f',
     },
     bar: {
       type: 'string',
     },
   };
   const {
     values,
     positionals,
   } = parseArgs({ args, options });
   console.log(values, positionals);
   // Prints: [Object: null prototype] { foo: true, bar: 'b' } []


parseEnv
~~~~~~~~

.. code-block:: text

   util.parseEnv(content)

*Kind:* function · *Arguments:* 1

The raw contents of a ``.env`` file.

* Returns: ``Object``

**Arguments**

* ``content`` (*string*)

**Example**

.. code-block:: javascript

   const { parseEnv } = require('node:util');

   parseEnv('HELLO=world\nHELLO=oh my\n');
   // Returns: { HELLO: 'oh my' }


promisify
~~~~~~~~~

.. code-block:: text

   util.promisify(original)

*Kind:* function · *Arguments:* 1

Takes a function following the common error-first callback style, i.e. taking an ``(err, value) => ...`` callback as the last argument, and returns a version that returns promises.

Or, equivalently using ``async function``\ s:

**Arguments**

* ``original`` (*Function*)

**Example**

.. code-block:: javascript

   const { promisify } = require('node:util');
   const { stat } = require('node:fs');

   const promisifiedStat = promisify(stat);
   promisifiedStat('.').then((stats) => {
     // Do something with `stats`
   }).catch((error) => {
     // Handle the error.
   });


setTraceSigInt
~~~~~~~~~~~~~~

.. code-block:: text

   util.setTraceSigInt(enable)

*Kind:* function · *Arguments:* 1

Enable or disable printing a stack trace on ``SIGINT``. The API is only available on the main thread.

**Arguments**

* ``enable`` (*boolean*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.setTraceSigInt(true);
   console.log(result);


sleep
~~~~~

.. code-block:: text

   util.sleep(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.sleep(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stripVTControlCharacters
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.stripVTControlCharacters(str)

*Kind:* function · *Arguments:* 1

Returns ``str`` with any ANSI escape codes removed.

**Arguments**

* ``str`` (*string*)

**Example**

.. code-block:: javascript

   console.log(util.stripVTControlCharacters('\u001B[4mvalue\u001B[0m'));
   // Prints "value"


styleText
~~~~~~~~~

.. code-block:: text

   util.styleText(format, text[, options])

*Kind:* function · *Arguments:* 3

This function returns a formatted text considering the ``format`` passed for printing in a terminal. It is aware of the terminal's capabilities and acts according to the configuration set via ``NO_COLOR``, ``NODE_DISABLE_COLORS`` and ``FORCE_COLOR`` environment variables.

``util.inspect.colors`` also provides text formats such as ``italic``, and ``underline`` and you can combine both:

**Arguments**

* ``format`` (*string | Array*): A text format or an Array
* ``text`` (*string*): The text to be formatted.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { styleText } = require('node:util');
   const { stderr } = require('node:process');

   const successMessage = styleText('green', 'Success!');
   console.log(successMessage);

   const errorMessage = styleText(
     'red',
     'Error! Error!',
     // Validate if process.stderr has TTY
     { stream: stderr },
   );
   console.error(errorMessage);


TextDecoder
~~~~~~~~~~~

.. code-block:: text

   util.TextDecoder

*Kind:* class · *Arguments:* 2

An implementation of the WHATWG Encoding Standard ``TextDecoder`` API.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const decoder = new TextDecoder();
   const u8arr = new Uint8Array([72, 101, 108, 108, 111]);
   console.log(decoder.decode(u8arr)); // Hello


TextEncoder
~~~~~~~~~~~

.. code-block:: text

   util.TextEncoder

*Kind:* class · *Arguments:* 0

An implementation of the WHATWG Encoding Standard ``TextEncoder`` API. All instances of ``TextEncoder`` only support UTF-8 encoding.

The ``TextEncoder`` class is also available on the global object.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const encoder = new TextEncoder();
   const uint8array = encoder.encode('this is some data');


throttle
~~~~~~~~

.. code-block:: text

   util.throttle(fn, limit, interval[, options])

*Kind:* function · *Arguments:* 4

Creates a function that limits how often ``fn`` is invoked. By default, calls that exceed the limit are queued in the order received rather than discarded. The throttled function returns a ``Promise`` for the value returned by ``fn``. If ``fn`` throws or returns a rejected promise, the returned promise is rejected with the same reason.

An invocation starts only when both rate and concurrency capacity are available. Rate capacity is consumed when ``fn`` starts, not when a call enters the queue. Concurrency capacity is released when the value returned by ``fn`` settles. Non-promise values settle during the next microtask.

**Arguments**

* ``fn`` (*Function*): The function to throttle.
* ``limit`` (*integer*): The maximum number of times to invoke ``fn`` during an
* ``interval`` (*integer*): The length of each interval in milliseconds.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { throttle } from 'node:util';

   const request = throttle(async (id) => {
     const response = await fetch(`https://example.com/items/${id}`);
     return response.json();
   }, 2, 1_000);

   // At most two requests begin during each one-second interval. All other calls
   // remain queued and retain their original arguments.
   const results = await Promise.all([
     request(1),
     request(2),
     request(3),
     request(4),
   ]);


toUSVString
~~~~~~~~~~~

.. code-block:: text

   util.toUSVString

*Kind:* function · *Arguments:* 0

Type: Documentation-only

The ``util.toUSVString()`` API is deprecated. Please use ``String.prototype.toWellFormed`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.toUSVString();
   console.log(result);


transferableAbortController
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.transferableAbortController()

*Kind:* function · *Arguments:* 0

Creates and returns an ``AbortController`` instance whose ``AbortSignal`` is marked as transferable and can be used with ``structuredClone()`` or ``postMessage()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.transferableAbortController();
   console.log(result);


transferableAbortSignal
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.transferableAbortSignal(signal)

*Kind:* function · *Arguments:* 1

Marks the given ``AbortSignal`` as transferable so that it can be used with ``structuredClone()`` and ``postMessage()``.

**Arguments**

* ``signal`` (*AbortSignal*)

**Example**

.. code-block:: javascript

   const signal = transferableAbortSignal(AbortSignal.timeout(100));
   const channel = new MessageChannel();
   channel.port2.postMessage(signal, [signal]);


MIMEParams class
----------------

**MIMEParams** — The ``MIMEParams`` API provides read and write access to the parameters of a ``MIMEType``.

MIMEParams.delete
~~~~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.delete(name)

*Kind:* function · *Arguments:* 1

Remove all name-value pairs whose name is ``name``.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.delete('example');
   console.log(result);


MIMEParams.entries
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.entries()

*Kind:* function · *Arguments:* 0

Returns an iterator over each of the name-value pairs in the parameters. Each item of the iterator is a JavaScript ``Array``. The first item of the array is the ``name``, the second item of the array is the ``value``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.entries();
   console.log(result);


MIMEParams.get
~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.get(name)

*Kind:* function · *Arguments:* 1

Returns the value of the first name-value pair whose name is ``name``. If there are no such pairs, ``null`` is returned.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.get('example');
   console.log(result);


MIMEParams.has
~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.has(name)

*Kind:* function · *Arguments:* 1

Returns ``true`` if there is at least one name-value pair whose name is ``name``.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.has('example');
   console.log(result);


MIMEParams.keys
~~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.keys()

*Kind:* function · *Arguments:* 0

Returns an iterator over the names of each name-value pair.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { MIMEType } = require('node:util');

   const { params } = new MIMEType('text/plain;foo=0;bar=1');
   for (const name of params.keys()) {
     console.log(name);
   }
   // Prints:
   //   foo
   //   bar


MIMEParams.set
~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.set(name, value)

*Kind:* function · *Arguments:* 2

Sets the value in the ``MIMEParams`` object associated with ``name`` to ``value``. If there are any pre-existing name-value pairs whose names are ``name``, set the first such pair's value to ``value``.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*)

**Example**

.. code-block:: javascript

   const { MIMEType } = require('node:util');

   const { params } = new MIMEType('text/plain;foo=0;bar=1');
   params.set('foo', 'def');
   params.set('baz', 'xyz');
   console.log(params.toString());
   // Prints: foo=def;bar=1;baz=xyz


MIMEParams.toJSON
~~~~~~~~~~~~~~~~~

.. code-block:: text

   mime.toJSON()

*Kind:* function · *Arguments:* 0

Alias for ``mime.toString()``.

This method is automatically called when an ``MIMEType`` object is serialized with ``JSON.stringify()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { MIMEType } = require('node:util');

   const myMIMES = [
     new MIMEType('image/png'),
     new MIMEType('image/gif'),
   ];
   console.log(JSON.stringify(myMIMES));
   // Prints: ["image/png", "image/gif"]


MIMEParams.toString
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   mime.toString()

*Kind:* function · *Arguments:* 0

The ``toString()`` method on the ``MIMEType`` object returns the serialized MIME.

Because of the need for standard compliance, this method does not allow users to customize the serialization process of the MIME.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.toString();
   console.log(result);


MIMEParams.values
~~~~~~~~~~~~~~~~~

.. code-block:: text

   mimeParams.values()

*Kind:* function · *Arguments:* 0

Returns an iterator over the values of each name-value pair.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEParams();
   const result = instance.values();
   console.log(result);


MIMEType class
--------------

**MIMEType** — An implementation of \`the MIMEType class <https://bmeck.github.io/node-proposal-mime-api/>\`_.

In accordance with browser conventions, all properties of ``MIMEType`` objects are implemented as getters and setters on the class prototype, rather than as data properties on the object itself.

MIMEType.toJSON
~~~~~~~~~~~~~~~

.. code-block:: text

   mime.toJSON()

*Kind:* function · *Arguments:* 0

Alias for ``mime.toString()``.

This method is automatically called when an ``MIMEType`` object is serialized with ``JSON.stringify()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { MIMEType } = require('node:util');

   const myMIMES = [
     new MIMEType('image/png'),
     new MIMEType('image/gif'),
   ];
   console.log(JSON.stringify(myMIMES));
   // Prints: ["image/png", "image/gif"]


MIMEType.toString
~~~~~~~~~~~~~~~~~

.. code-block:: text

   mime.toString()

*Kind:* function · *Arguments:* 0

The ``toString()`` method on the ``MIMEType`` object returns the serialized MIME.

Because of the need for standard compliance, this method does not allow users to customize the serialization process of the MIME.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.MIMEType();
   const result = instance.toString();
   console.log(result);


TextDecoder class
-----------------

**TextDecoder** — An implementation of the WHATWG Encoding Standard ``TextDecoder`` API.

TextDecoder.decode
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textDecoder.decode([input[, options]])

*Kind:* function · *Arguments:* 2

Decodes the ``input`` and returns a string. If ``options.stream`` is ``true``, any incomplete byte sequences occurring at the end of the ``input`` are buffered internally and emitted after the next call to ``textDecoder.decode()``.

If ``textDecoder.fatal`` is ``true``, decoding errors that occur will result in a ``TypeError`` being thrown.

**Arguments**

* ``input`` (*ArrayBuffer|DataView|TypedArray*) — optional: An ``ArrayBuffer``, ``DataView``, or
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.TextDecoder();
   const result = instance.decode([], {});
   console.log(result);


TextEncoder class
-----------------

**TextEncoder** — An implementation of the WHATWG Encoding Standard ``TextEncoder`` API. All instances of ``TextEncoder`` only support UTF-8 encoding.

The ``TextEncoder`` class is also available on the global object.

TextEncoder.encode
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textEncoder.encode([input])

*Kind:* function · *Arguments:* 1

UTF-8 encodes the ``input`` string and returns a ``Uint8Array`` containing the encoded bytes.

**Arguments**

* ``input`` (*string*) — optional: The text to encode. **Default:** an empty string.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const instance = new util.TextEncoder();
   const result = instance.encode('value');
   console.log(result);


TextEncoder.encodeInto
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textEncoder.encodeInto(src, dest)

*Kind:* function · *Arguments:* 2

UTF-8 encodes the ``src`` string to the ``dest`` Uint8Array and returns an object containing the read Unicode code units and written UTF-8 bytes.

**Arguments**

* ``src`` (*string*): The text to encode.
* ``dest`` (*Uint8Array*): The array to hold the encode result.

**Example**

.. code-block:: javascript

   const encoder = new TextEncoder();
   const src = 'this is some data';
   const dest = new Uint8Array(10);
   const { read, written } = encoder.encodeInto(src, dest);


types
-----

**types** — ``util.types`` provides type checks for different kinds of built-in objects. Unlike ``instanceof`` or ``Object.prototype.toString.call(value)``, these checks do not inspect properties of the object that are accessible from JavaScript (like their prototype), and usually have the overhead of calling into C++.

The result generally does not make any guarantees about what kinds of properties or behavior a value exposes in JavaScript. They are primarily useful for addon developers who prefer to do type checking in JavaScript.

types.isAnyArrayBuffer
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isAnyArrayBuffer(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``ArrayBuffer`` or ``SharedArrayBuffer`` instance.

See also ``util.types.isArrayBuffer()`` and ``util.types.isSharedArrayBuffer()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isAnyArrayBuffer(new ArrayBuffer());  // Returns true
   util.types.isAnyArrayBuffer(new SharedArrayBuffer());  // Returns true


types.isArgumentsObject
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isArgumentsObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an ``arguments`` object.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   function foo() {
     util.types.isArgumentsObject(arguments);  // Returns true
   }


types.isArrayBuffer
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isArrayBuffer(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``ArrayBuffer`` instance. This does *not* include ``SharedArrayBuffer`` instances. Usually, it is desirable to test for both; See ``util.types.isAnyArrayBuffer()`` for that.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isArrayBuffer(new ArrayBuffer());  // Returns true
   util.types.isArrayBuffer(new SharedArrayBuffer());  // Returns false


types.isArrayBufferView
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isArrayBufferView(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an instance of one of the ``ArrayBuffer`` views, such as typed array objects or ``DataView``. Equivalent to ``ArrayBuffer.isView()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isArrayBufferView(new Int8Array());  // true
   util.types.isArrayBufferView(Buffer.from('hello world')); // true
   util.types.isArrayBufferView(new DataView(new ArrayBuffer(16)));  // true
   util.types.isArrayBufferView(new ArrayBuffer());  // false


types.isAsyncFunction
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isAsyncFunction(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an async function. This only reports back what the JavaScript engine is seeing; in particular, the return value may not match the original source code if a transpilation tool was used.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isAsyncFunction(function foo() {});  // Returns false
   util.types.isAsyncFunction(async function foo() {});  // Returns true


types.isBigInt64Array
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isBigInt64Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a ``BigInt64Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isBigInt64Array(new BigInt64Array());   // Returns true
   util.types.isBigInt64Array(new BigUint64Array());  // Returns false


types.isBigIntObject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isBigIntObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a BigInt object, e.g. created by ``Object(BigInt(123))``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isBigIntObject(Object(BigInt(123)));   // Returns true
   util.types.isBigIntObject(BigInt(123));   // Returns false
   util.types.isBigIntObject(123);  // Returns false


types.isBigUint64Array
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isBigUint64Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a ``BigUint64Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isBigUint64Array(new BigInt64Array());   // Returns false
   util.types.isBigUint64Array(new BigUint64Array());  // Returns true


types.isBooleanObject
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isBooleanObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a boolean object, e.g. created by ``new Boolean()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isBooleanObject(false);  // Returns false
   util.types.isBooleanObject(true);   // Returns false
   util.types.isBooleanObject(new Boolean(false)); // Returns true
   util.types.isBooleanObject(new Boolean(true));  // Returns true
   util.types.isBooleanObject(Boolean(false)); // Returns false
   util.types.isBooleanObject(Boolean(true));  // Returns false


types.isBoxedPrimitive
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isBoxedPrimitive(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is any boxed primitive object, e.g. created by ``new Boolean()``, ``new String()`` or ``Object(Symbol())``.

For example:

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isBoxedPrimitive(false); // Returns false
   util.types.isBoxedPrimitive(new Boolean(false)); // Returns true
   util.types.isBoxedPrimitive(Symbol('foo')); // Returns false
   util.types.isBoxedPrimitive(Object(Symbol('foo'))); // Returns true
   util.types.isBoxedPrimitive(Object(BigInt(5))); // Returns true


types.isCryptoKey
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isCryptoKey(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``value`` is a ``CryptoKey``, ``false`` otherwise.

**Arguments**

* ``value`` (*Object*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.types.isCryptoKey({});
   console.log(result);


types.isDataView
~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isDataView(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``DataView`` instance.

See also ``ArrayBuffer.isView()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const ab = new ArrayBuffer(20);
   util.types.isDataView(new DataView(ab));  // Returns true
   util.types.isDataView(new Float64Array());  // Returns false


types.isDate
~~~~~~~~~~~~

.. code-block:: text

   util.types.isDate(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Date`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isDate(new Date());  // Returns true


types.isExternal
~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isExternal(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a native ``External`` value.

A native ``External`` value is a special type of object that contains a raw C++ pointer (``void*``) for access from native code, and has no other properties. Such objects are created either by Node.js internals or native addons. In JavaScript, they are frozen objects with a ``null`` prototype.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const native = require('napi_addon.node');
   const { types } = require('node:util');

   const data = native.myNapi();
   types.isExternal(data); // returns true
   types.isExternal(0); // returns false
   types.isExternal(new String('foo')); // returns false


types.isFloat16Array
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isFloat16Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Float16Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isFloat16Array(new ArrayBuffer());  // Returns false
   util.types.isFloat16Array(new Float16Array());  // Returns true
   util.types.isFloat16Array(new Float32Array());  // Returns false


types.isFloat32Array
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isFloat32Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Float32Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isFloat32Array(new ArrayBuffer());  // Returns false
   util.types.isFloat32Array(new Float32Array());  // Returns true
   util.types.isFloat32Array(new Float64Array());  // Returns false


types.isFloat64Array
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isFloat64Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Float64Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isFloat64Array(new ArrayBuffer());  // Returns false
   util.types.isFloat64Array(new Uint8Array());  // Returns false
   util.types.isFloat64Array(new Float64Array());  // Returns true


types.isGeneratorFunction
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isGeneratorFunction(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a generator function. This only reports back what the JavaScript engine is seeing; in particular, the return value may not match the original source code if a transpilation tool was used.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isGeneratorFunction(function foo() {});  // Returns false
   util.types.isGeneratorFunction(function* foo() {});  // Returns true


types.isGeneratorObject
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isGeneratorObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a generator object as returned from a built-in generator function. This only reports back what the JavaScript engine is seeing; in particular, the return value may not match the original source code if a transpilation tool was used.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   function* foo() {}
   const generator = foo();
   util.types.isGeneratorObject(generator);  // Returns true


types.isInt16Array
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isInt16Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Int16Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isInt16Array(new ArrayBuffer());  // Returns false
   util.types.isInt16Array(new Int16Array());  // Returns true
   util.types.isInt16Array(new Float64Array());  // Returns false


types.isInt32Array
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isInt32Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Int32Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isInt32Array(new ArrayBuffer());  // Returns false
   util.types.isInt32Array(new Int32Array());  // Returns true
   util.types.isInt32Array(new Float64Array());  // Returns false


types.isInt8Array
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isInt8Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Int8Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isInt8Array(new ArrayBuffer());  // Returns false
   util.types.isInt8Array(new Int8Array());  // Returns true
   util.types.isInt8Array(new Float64Array());  // Returns false


types.isKeyObject
~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isKeyObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``value`` is a ``KeyObject``, ``false`` otherwise.

**Arguments**

* ``value`` (*Object*)

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.types.isKeyObject({});
   console.log(result);


types.isMap
~~~~~~~~~~~

.. code-block:: text

   util.types.isMap(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Map`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isMap(new Map());  // Returns true


types.isMapIterator
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isMapIterator(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an iterator returned for a built-in ``Map`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const map = new Map();
   util.types.isMapIterator(map.keys());  // Returns true
   util.types.isMapIterator(map.values());  // Returns true
   util.types.isMapIterator(map.entries());  // Returns true
   util.types.isMapIterator(map[Symbol.iterator]());  // Returns true


types.isModuleNamespaceObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isModuleNamespaceObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an instance of a Module Namespace Object.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   import * as ns from './a.js';

   util.types.isModuleNamespaceObject(ns);  // Returns true


types.isNativeError
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isNativeError()

*Kind:* function · *Arguments:* 0

Type: Documentation-only

The ``util.types.isNativeError`` API is deprecated. Please use ``Error.isError`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const util = require('node:util');

   const result = util.types.isNativeError();
   console.log(result);


types.isNumberObject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isNumberObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a number object, e.g. created by ``new Number()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isNumberObject(0);  // Returns false
   util.types.isNumberObject(new Number(0));   // Returns true


types.isPromise
~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isPromise(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Promise``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isPromise(Promise.resolve(42));  // Returns true


types.isProxy
~~~~~~~~~~~~~

.. code-block:: text

   util.types.isProxy(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a ``Proxy`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const target = {};
   const proxy = new Proxy(target, {});
   util.types.isProxy(target);  // Returns false
   util.types.isProxy(proxy);  // Returns true


types.isRegExp
~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isRegExp(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a regular expression object.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isRegExp(/abc/);  // Returns true
   util.types.isRegExp(new RegExp('abc'));  // Returns true


types.isSet
~~~~~~~~~~~

.. code-block:: text

   util.types.isSet(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Set`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isSet(new Set());  // Returns true


types.isSetIterator
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isSetIterator(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is an iterator returned for a built-in ``Set`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const set = new Set();
   util.types.isSetIterator(set.keys());  // Returns true
   util.types.isSetIterator(set.values());  // Returns true
   util.types.isSetIterator(set.entries());  // Returns true
   util.types.isSetIterator(set[Symbol.iterator]());  // Returns true


types.isSharedArrayBuffer
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isSharedArrayBuffer(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``SharedArrayBuffer`` instance. This does *not* include ``ArrayBuffer`` instances. Usually, it is desirable to test for both; See ``util.types.isAnyArrayBuffer()`` for that.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isSharedArrayBuffer(new ArrayBuffer());  // Returns false
   util.types.isSharedArrayBuffer(new SharedArrayBuffer());  // Returns true


types.isStringObject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isStringObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a string object, e.g. created by ``new String()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isStringObject('foo');  // Returns false
   util.types.isStringObject(new String('foo'));   // Returns true


types.isSymbolObject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isSymbolObject(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a symbol object, created by calling ``Object()`` on a ``Symbol`` primitive.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const symbol = Symbol('foo');
   util.types.isSymbolObject(symbol);  // Returns false
   util.types.isSymbolObject(Object(symbol));   // Returns true


types.isTypedArray
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isTypedArray(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``TypedArray`` instance.

See also ``ArrayBuffer.isView()``.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isTypedArray(new ArrayBuffer());  // Returns false
   util.types.isTypedArray(new Uint8Array());  // Returns true
   util.types.isTypedArray(new Float64Array());  // Returns true


types.isUint16Array
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isUint16Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Uint16Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isUint16Array(new ArrayBuffer());  // Returns false
   util.types.isUint16Array(new Uint16Array());  // Returns true
   util.types.isUint16Array(new Float64Array());  // Returns false


types.isUint32Array
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isUint32Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Uint32Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isUint32Array(new ArrayBuffer());  // Returns false
   util.types.isUint32Array(new Uint32Array());  // Returns true
   util.types.isUint32Array(new Float64Array());  // Returns false


types.isUint8Array
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isUint8Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Uint8Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isUint8Array(new ArrayBuffer());  // Returns false
   util.types.isUint8Array(new Uint8Array());  // Returns true
   util.types.isUint8Array(new Float64Array());  // Returns false


types.isUint8ClampedArray
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isUint8ClampedArray(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Uint8ClampedArray`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isUint8ClampedArray(new ArrayBuffer());  // Returns false
   util.types.isUint8ClampedArray(new Uint8ClampedArray());  // Returns true
   util.types.isUint8ClampedArray(new Float64Array());  // Returns false


types.isWeakMap
~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isWeakMap(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``WeakMap`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isWeakMap(new WeakMap());  // Returns true


types.isWeakSet
~~~~~~~~~~~~~~~

.. code-block:: text

   util.types.isWeakSet(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``WeakSet`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isWeakSet(new WeakSet());  // Returns true


Properties and constants
------------------------

* ``custom`` — symbol
* ``MIMEType.essence`` — string
* ``MIMEType.params`` — undefined
* ``MIMEType.subtype`` — undefined
* ``MIMEType.type`` — undefined
* ``TextEncoder.encoding`` — string
* ``types`` — object

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
