console
=======

The ``node:console`` module provides a simple debugging console that is similar to the JavaScript console mechanism provided by web browsers.

The module exports two specific components:

Loading
-------

.. code-block:: javascript

   const console = require('node:console');
   // ESM
   import * as console from 'node:console';

Quick start
-----------

.. code-block:: javascript

   console.log('hello world');
   // Prints: hello world, to stdout
   console.log('hello %s', 'world');
   // Prints: hello world, to stdout
   console.error(new Error('Whoops, something bad happened'));
   // Prints error message and stack trace to stderr:
   //   Error: Whoops, something bad happened
   //     at [eval]:5:15
   //     at Script.runInThisContext (node:vm:132:18)
   //     at Object.runInThisContext (node:vm:309:38)
   //     at node:internal/process/execution:77:19
   //     at [eval]-wrapper:6:22
   //     at evalScript (node:internal/process/execution:76:60)
   //     at node:internal/main/eval_string:23:3

   const name = 'Will Robinson';
   console.warn(`Danger ${name}! Danger!`);
   // Prints: Danger Will Robinson! Danger!, to stderr

API
---

assert
~~~~~~

.. code-block:: text

   console.assert(value[, ...message])

*Kind:* function · *Arguments:* 2

``console.assert()`` writes a message if ``value`` is falsy or omitted. It only writes a message and does not otherwise affect execution. The output always starts with ``"Assertion failed"``. If provided, ``message`` is formatted using ``util.format()``.

If ``value`` is truthy, nothing happens.

**Arguments**

* ``value`` (*any*): The value tested for being truthy.
* ``...message`` — optional

**Example**

.. code-block:: javascript

   console.assert(true, 'does nothing');

   console.assert(false, 'Whoops %s work', 'didn\'t');
   // Assertion failed: Whoops didn't work

   console.assert();
   // Assertion failed

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleassertvalue-message>`__

clear
~~~~~

.. code-block:: text

   console.clear()

*Kind:* function · *Arguments:* 0

When ``stdout`` is a TTY, calling ``console.clear()`` will attempt to clear the TTY. When ``stdout`` is not a TTY, this method does nothing.

The specific operation of ``console.clear()`` can vary across operating systems and terminal types. For most Linux operating systems, ``console.clear()`` operates similarly to the ``clear`` shell command. On Windows, ``console.clear()`` will clear only the output in the current terminal viewport for the Node.js binary.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleclear>`__

Console
~~~~~~~

.. code-block:: text

   Console

*Kind:* class · *Arguments:* 2

The ``Console`` class can be used to create a simple logger with configurable output streams and can be accessed using either ``require('node:console').Console`` or ``console.Console`` (or their destructured counterparts):

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Console } = require('node:console');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#class-console>`__

count
~~~~~

.. code-block:: text

   console.count([label])

*Kind:* function · *Arguments:* 1

Maintains an internal counter specific to ``label`` and outputs to ``stdout`` the number of times ``console.count()`` has been called with the given ``label``.

**Arguments**

* ``label`` (*string*) — optional: The display label for the counter. **Default:** ``'default'``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolecountlabel>`__

countReset
~~~~~~~~~~

.. code-block:: text

   console.countReset([label])

*Kind:* function · *Arguments:* 1

Resets the internal counter specific to ``label``.

**Arguments**

* ``label`` (*string*) — optional: The display label for the counter. **Default:** ``'default'``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolecountresetlabel>`__

debug
~~~~~

.. code-block:: text

   console.debug(data[, ...args])

*Kind:* function · *Arguments:* 2

The ``console.debug()`` function is an alias for ``console.log()``.

**Arguments**

* ``data`` (*any*)
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoledebugdata-args>`__

dir
~~~

.. code-block:: text

   console.dir(obj[, options])

*Kind:* function · *Arguments:* 2

Uses ``util.inspect()`` on ``obj`` and prints the resulting string to ``stdout``. This function bypasses any custom ``inspect()`` function defined on ``obj``.

**Arguments**

* ``obj`` (*any*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoledirobj-options>`__

dirxml
~~~~~~

.. code-block:: text

   console.dirxml(...data)

*Kind:* function · *Arguments:* 1

This method calls ``console.log()`` passing it the arguments received. This method does not produce any XML formatting.

**Arguments**

* ``...data``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoledirxmldata>`__

error
~~~~~

.. code-block:: text

   console.error([data][, ...args])

*Kind:* function · *Arguments:* 2

Prints to ``stderr`` with newline. Multiple arguments can be passed, with the first used as the primary message and all additional used as substitution values similar to printf(3) (the arguments are all passed to ``util.format()``).

If formatting elements (e.g. ``%d``) are not found in the first string then ``util.inspect()`` is called on each argument and the resulting string values are concatenated. See ``util.format()`` for more information.

**Arguments**

* ``data`` (*any*) — optional
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const code = 5;
   console.error('error #%d', code);
   // Prints: error #5, to stderr
   console.error('error', code);
   // Prints: error 5, to stderr

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleerrordata-args>`__

group
~~~~~

.. code-block:: text

   console.group([...label])

*Kind:* function · *Arguments:* 1

Increases indentation of subsequent lines by spaces for ``groupIndentation`` length.

If one or more ``label``\ s are provided, those are printed first without the additional indentation.

**Arguments**

* ``...label`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolegrouplabel>`__

groupCollapsed
~~~~~~~~~~~~~~

.. code-block:: text

   console.groupCollapsed()

*Kind:* function · *Arguments:* 0

An alias for ``console.group()``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolegroupcollapsed>`__

groupEnd
~~~~~~~~

.. code-block:: text

   console.groupEnd()

*Kind:* function · *Arguments:* 0

Decreases indentation of subsequent lines by spaces for ``groupIndentation`` length.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolegroupend>`__

info
~~~~

.. code-block:: text

   console.info([data][, ...args])

*Kind:* function · *Arguments:* 2

The ``console.info()`` function is an alias for ``console.log()``.

**Arguments**

* ``data`` (*any*) — optional
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleinfodata-args>`__

log
~~~

.. code-block:: text

   console.log([data][, ...args])

*Kind:* function · *Arguments:* 2

Prints to ``stdout`` with newline. Multiple arguments can be passed, with the first used as the primary message and all additional used as substitution values similar to printf(3) (the arguments are all passed to ``util.format()``).

See ``util.format()`` for more information.

**Arguments**

* ``data`` (*any*) — optional
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const count = 5;
   console.log('count: %d', count);
   // Prints: count: 5, to stdout
   console.log('count:', count);
   // Prints: count: 5, to stdout

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolelogdata-args>`__

profile
~~~~~~~

.. code-block:: text

   console.profile([label])

*Kind:* function · *Arguments:* 1

This method does not display anything unless used in the inspector. The ``console.profile()`` method starts a JavaScript CPU profile with an optional label until ``console.profileEnd()`` is called. The profile is then added to the **Profile** panel of the inspector.

**Arguments**

* ``label`` (*string*) — optional

**Example**

.. code-block:: javascript

   console.profile('MyLabel');
   // Some code
   console.profileEnd('MyLabel');
   // Adds the profile 'MyLabel' to the Profiles panel of the inspector.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleprofilelabel>`__

profileEnd
~~~~~~~~~~

.. code-block:: text

   console.profileEnd([label])

*Kind:* function · *Arguments:* 1

This method does not display anything unless used in the inspector. Stops the current JavaScript CPU profiling session if one has been started and prints the report to the **Profiles** panel of the inspector. See ``console.profile()`` for an example.

If this method is called without a label, the most recently started profile is stopped.

**Arguments**

* ``label`` (*string*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoleprofileendlabel>`__

table
~~~~~

.. code-block:: text

   console.table(tabularData[, properties])

*Kind:* function · *Arguments:* 2

Try to construct a table with the columns of the properties of ``tabularData`` (or use ``properties``) and rows of ``tabularData`` and log it. Falls back to just logging the argument if it can't be parsed as tabular.

**Arguments**

* ``tabularData`` (*any*)
* ``properties`` (*string\[]*) — optional: Alternate properties for constructing the table.

**Example**

.. code-block:: javascript

   // These can't be parsed as tabular data
   console.table(Symbol());
   // Symbol()

   console.table(undefined);
   // undefined

   console.table([{ a: 1, b: 'Y' }, { a: 'Z', b: 2 }]);
   // ┌─────────┬─────┬─────┐
   // │ (index) │ a   │ b   │
   // ├─────────┼─────┼─────┤
   // │ 0       │ 1   │ 'Y' │
   // │ 1       │ 'Z' │ 2   │
   // └─────────┴─────┴─────┘

   console.table([{ a: 1, b: 'Y' }, { a: 'Z', b: 2 }], ['a']);
   // ┌─────────┬─────┐
   // │ (index) │ a   │
   // ├─────────┼─────┤
   // │ 0       │ 1   │
   // │ 1       │ 'Z' │
   // └─────────┴─────┘

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletabletabulardata-properties>`__

time
~~~~

.. code-block:: text

   console.time([label])

*Kind:* function · *Arguments:* 1

Starts a timer that can be used to compute the duration of an operation. Timers are identified by a unique ``label``. Use the same ``label`` when calling ``console.timeEnd()`` to stop the timer and output the elapsed time in suitable time units to ``stdout``. For example, if the elapsed time is 3869ms, ``console.timeEnd()`` displays "3.869s".

**Arguments**

* ``label`` (*string*) — optional: **Default:** ``'default'``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletimelabel>`__

timeEnd
~~~~~~~

.. code-block:: text

   console.timeEnd([label])

*Kind:* function · *Arguments:* 1

Stops a timer that was previously started by calling ``console.time()`` and prints the result to ``stdout``:

**Arguments**

* ``label`` (*string*) — optional: **Default:** ``'default'``

**Example**

.. code-block:: javascript

   console.time('bunch-of-stuff');
   // Do a bunch of stuff.
   console.timeEnd('bunch-of-stuff');
   // Prints: bunch-of-stuff: 225.438ms

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletimeendlabel>`__

timeLog
~~~~~~~

.. code-block:: text

   console.timeLog([label][, ...data])

*Kind:* function · *Arguments:* 2

For a timer that was previously started by calling ``console.time()``, prints the elapsed time and other ``data`` arguments to ``stdout``:

**Arguments**

* ``label`` (*string*) — optional: **Default:** ``'default'``
* ``...data`` — optional

**Example**

.. code-block:: javascript

   console.time('process');
   const value = expensiveProcess1(); // Returns 42
   console.timeLog('process', value);
   // Prints "process: 365.227ms 42".
   doExpensiveProcess2(value);
   console.timeEnd('process');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletimeloglabel-data>`__

timeStamp
~~~~~~~~~

.. code-block:: text

   console.timeStamp([label])

*Kind:* function · *Arguments:* 1

This method does not display anything unless used in the inspector. The ``console.timeStamp()`` method adds an event with the label ``'label'`` to the **Timeline** panel of the inspector.

**Arguments**

* ``label`` (*string*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletimestamplabel>`__

trace
~~~~~

.. code-block:: text

   console.trace([message][, ...args])

*Kind:* function · *Arguments:* 2

Prints to ``stderr`` the string ``'Trace: '``, followed by the ``util.format()`` formatted message and stack trace to the current position in the code.

**Arguments**

* ``message`` (*any*) — optional
* ``...args`` — optional

**Example**

.. code-block:: javascript

   console.trace('Show me');
   // Prints: (stack trace will vary based on where trace is called)
   //  Trace: Show me
   //    at repl:2:9
   //    at REPLServer.defaultEval (repl.js:248:27)
   //    at bound (domain.js:287:14)
   //    at REPLServer.runBound [as eval] (domain.js:300:12)
   //    at REPLServer.<anonymous> (repl.js:412:12)
   //    at emitOne (events.js:82:20)
   //    at REPLServer.emit (events.js:169:7)
   //    at REPLServer.Interface._onLine (readline.js:210:10)
   //    at REPLServer.Interface._line (readline.js:549:8)
   //    at REPLServer.Interface._ttyWrite (readline.js:826:14)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consoletracemessage-args>`__

warn
~~~~

.. code-block:: text

   console.warn([data][, ...args])

*Kind:* function · *Arguments:* 2

The ``console.warn()`` function is an alias for ``console.error()``.

**Arguments**

* ``data`` (*any*) — optional
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/console.html#consolewarndata-args>`__

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
