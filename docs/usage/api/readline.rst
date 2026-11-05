readline
========

The ``node:readline`` module provides an interface for reading data from a Readable stream (such as ``process.stdin``) one line at a time.

To use the promise-based APIs:

Loading
-------

.. code-block:: javascript

   const readline = require('node:readline');
   // ESM
   import * as readline from 'node:readline';

Quick start
-----------

.. code-block:: javascript

   const readline = require('node:readline/promises');

API
---

clearLine
~~~~~~~~~

.. code-block:: text

   readline.clearLine(stream, dir[, callback])

*Kind:* function · *Arguments:* 3

The ``readline.clearLine()`` method clears current line of given TTY stream in a specified direction identified by ``dir``.

**Arguments**

* ``stream`` (*stream.Writable*)
* ``dir`` (*number*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const result = readline.clearLine(process.stdout, './file.txt', () => {});
   console.log(result);


clearScreenDown
~~~~~~~~~~~~~~~

.. code-block:: text

   readline.clearScreenDown(stream[, callback])

*Kind:* function · *Arguments:* 2

The ``readline.clearScreenDown()`` method clears the given TTY stream from the current position of the cursor down.

**Arguments**

* ``stream`` (*stream.Writable*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const result = readline.clearScreenDown(process.stdout, () => {});
   console.log(result);


createInterface
~~~~~~~~~~~~~~~

.. code-block:: text

   readline.createInterface(options)

*Kind:* function · *Arguments:* 1

The ``readline.createInterface()`` method creates a new ``readline.Interface`` instance.

Once the ``readline.Interface`` instance is created, the most common case is to listen for the ``'line'`` event:

**Arguments**

* ``options`` (*Object*)

**Example**

.. code-block:: javascript

   const { createInterface } = require('node:readline');
   const rl = createInterface({
     input: process.stdin,
     output: process.stdout,
   });


cursorTo
~~~~~~~~

.. code-block:: text

   readline.cursorTo(stream, x[, y][, callback])

*Kind:* function · *Arguments:* 4

The ``readline.cursorTo()`` method moves cursor to the specified position in a given TTY ``stream``.

**Arguments**

* ``stream`` (*stream.Writable*)
* ``x`` (*number*)
* ``y`` (*number*) — optional
* ``callback`` (*Function*) — optional: Invoked once the operation completes.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const result = readline.cursorTo(process.stdout, 1, 1, () => {});
   console.log(result);


emitKeypressEvents
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readline.emitKeypressEvents(stream[, interface])

*Kind:* function · *Arguments:* 2

The ``readline.emitKeypressEvents()`` method causes the given Readable stream to begin emitting ``'keypress'`` events corresponding to received input.

Optionally, ``interface`` specifies a ``readline.Interface`` instance for which autocompletion is disabled when copy-pasted input is detected.

**Arguments**

* ``stream`` (*stream.Readable*)
* ``interface`` (*readline.InterfaceConstructor*) — optional

**Example**

.. code-block:: javascript

   readline.emitKeypressEvents(process.stdin);
   if (process.stdin.isTTY)
     process.stdin.setRawMode(true);


Interface
~~~~~~~~~

.. code-block:: text

   readline.Interface

*Kind:* class · *Arguments:* 3

Instances of the ``readline.Interface`` class are constructed using the ``readline.createInterface()`` method. Every instance is associated with a single ``input`` Readable stream and a single ``output`` Writable stream. The ``output`` stream is used to print prompts for user input that arrives on, and is read from, the ``input`` stream.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   rl.question('What is your favorite food? ', (answer) => {
     console.log(`Oh, so your favorite food is ${answer}`);
   });


moveCursor
~~~~~~~~~~

.. code-block:: text

   readline.moveCursor(stream, dx, dy[, callback])

*Kind:* function · *Arguments:* 4

The ``readline.moveCursor()`` method moves the cursor *relative* to its current position in a given TTY ``stream``.

**Arguments**

* ``stream`` (*stream.Writable*)
* ``dx`` (*number*)
* ``dy`` (*number*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const result = readline.moveCursor(process.stdout, 1, 1, () => {});
   console.log(result);


Interface class
---------------

**Interface** — Instances of the ``readline.Interface`` class are constructed using the ``readline.createInterface()`` method. Every instance is associated with a single ``input`` Readable stream and a single ``output`` Writable stream. The ``output`` stream is used to print prompts for user input that arrives on, and is read from, the ``input`` stream.

Interface._ask
~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._ask(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   console.log(typeof readline.Interface.prototype._ask()); // 'function'
   console.log(readline.Interface.prototype._ask().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Interface._handleData
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._handleData(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   console.log(typeof readline.Interface.prototype._handleData()); // 'function'
   console.log(readline.Interface.prototype._handleData().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Interface._handleEnd
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._handleEnd(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   console.log(typeof readline.Interface.prototype._handleEnd()); // 'function'
   console.log(readline.Interface.prototype._handleEnd().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Interface._line
~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._line(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   console.log(typeof readline.Interface.prototype._line()); // 'function'
   console.log(readline.Interface.prototype._line().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Interface.close
~~~~~~~~~~~~~~~

.. code-block:: text

   rl.close()

*Kind:* function · *Arguments:* 0

The ``rl.close()`` method closes the ``InterfaceConstructor`` instance and relinquishes control over the ``input`` and ``output`` streams. When called, the ``'close'`` event will be emitted.

Calling ``rl.close()`` does not immediately stop other events (including ``'line'``) from being emitted by the ``InterfaceConstructor`` instance.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.close();
   console.log(result);


Interface.getCursorPos
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.getCursorPos()

*Kind:* function · *Arguments:* 0

Returns the real position of the cursor in relation to the input prompt + string. Long input (wrapping) strings, as well as multiple line prompts are included in the calculations.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.getCursorPos();
   console.log(result);


Interface.getPrompt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.getPrompt()

*Kind:* function · *Arguments:* 0

The ``rl.getPrompt()`` method returns the current prompt used by ``rl.prompt()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.getPrompt();
   console.log(result);


Interface.pause
~~~~~~~~~~~~~~~

.. code-block:: text

   rl.pause()

*Kind:* function · *Arguments:* 0

The ``rl.pause()`` method pauses the ``input`` stream, allowing it to be resumed later if necessary.

Calling ``rl.pause()`` does not immediately pause other events (including ``'line'``) from being emitted by the ``InterfaceConstructor`` instance.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.pause();
   console.log(result);


Interface.prompt
~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.prompt([preserveCursor])

*Kind:* function · *Arguments:* 1

The ``rl.prompt()`` method writes the ``InterfaceConstructor`` instances configured ``prompt`` to a new line in ``output`` in order to provide a user with a new location at which to provide input.

When called, ``rl.prompt()`` will resume the ``input`` stream if it has been paused.

**Arguments**

* ``preserveCursor`` (*boolean*) — optional: If ``true``, prevents the cursor placement from

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.prompt(true);
   console.log(result);


Interface.question
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.question(query[, options], callback)

*Kind:* function · *Arguments:* 3

The ``rl.question()`` method displays the ``query`` by writing it to the ``output``, waits for user input to be provided on ``input``, then invokes the ``callback`` function passing the provided input as the first argument.

When called, ``rl.question()`` will resume the ``input`` stream if it has been paused.

**Arguments**

* ``query`` (*string*): A statement or query to write to ``output``, prepended to the
* ``options`` (*Object*) — optional
* ``callback`` (*Function*): A callback function that is invoked with the user's

**Example**

.. code-block:: javascript

   rl.question('What is your favorite food? ', (answer) => {
     console.log(`Oh, so your favorite food is ${answer}`);
   });


Interface.resume
~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.resume()

*Kind:* function · *Arguments:* 0

The ``rl.resume()`` method resumes the ``input`` stream if it has been paused.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.resume();
   console.log(result);


Interface.setPrompt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.setPrompt(prompt)

*Kind:* function · *Arguments:* 1

The ``rl.setPrompt()`` method sets the prompt that will be written to ``output`` whenever ``rl.prompt()`` is called.

**Arguments**

* ``prompt`` (*string*)

**Example**

.. code-block:: javascript

   const readline = require('node:readline');

   const instance = new readline.Interface();
   const result = instance.setPrompt('value');
   console.log(result);


Interface.write
~~~~~~~~~~~~~~~

.. code-block:: text

   rl.write(data[, key])

*Kind:* function · *Arguments:* 2

The ``rl.write()`` method will write either ``data`` or a key sequence identified by ``key`` to the ``output``. The ``key`` argument is supported only if ``output`` is a TTY text terminal. See TTY keybindings for a list of key combinations.

If ``key`` is specified, ``data`` is ignored.

**Arguments**

* ``data`` (*string*)
* ``key`` (*Object*) — optional

**Example**

.. code-block:: javascript

   rl.write('Delete this!');
   // Simulate Ctrl+U to delete the line written previously
   rl.write(null, { ctrl: true, name: 'u' });


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
