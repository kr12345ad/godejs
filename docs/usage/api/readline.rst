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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlineclearlinestream-dir-callback>`__

clearScreenDown
~~~~~~~~~~~~~~~

.. code-block:: text

   readline.clearScreenDown(stream[, callback])

*Kind:* function · *Arguments:* 2

The ``readline.clearScreenDown()`` method clears the given TTY stream from the current position of the cursor down.

**Arguments**

* ``stream`` (*stream.Writable*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlineclearscreendownstream-callback>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlinecreateinterfaceoptions>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlinecursortostream-x-y-callback>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlineemitkeypresseventsstream-interface>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#class-readlineinterface>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#readlinemovecursorstream-dx-dy-callback>`__

Interface class
---------------

**Interface** — Instances of the ``readline.Interface`` class are constructed using the ``readline.createInterface()`` method. Every instance is associated with a single ``input`` Readable stream and a single ``output`` Writable stream. The ``output`` stream is used to print prompts for user input that arrives on, and is read from, the ``input`` stream.

Interface._ask
~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._ask(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Interface._handleData
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._handleData(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Interface._handleEnd
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._handleEnd(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Interface._line
~~~~~~~~~~~~~~~

.. code-block:: text

   readline.Interface._line(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Interface.close
~~~~~~~~~~~~~~~

.. code-block:: text

   rl.close()

*Kind:* function · *Arguments:* 0

The ``rl.close()`` method closes the ``InterfaceConstructor`` instance and relinquishes control over the ``input`` and ``output`` streams. When called, the ``'close'`` event will be emitted.

Calling ``rl.close()`` does not immediately stop other events (including ``'line'``) from being emitted by the ``InterfaceConstructor`` instance.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlclose>`__

Interface.getCursorPos
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.getCursorPos()

*Kind:* function · *Arguments:* 0

Returns the real position of the cursor in relation to the input prompt + string. Long input (wrapping) strings, as well as multiple line prompts are included in the calculations.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlgetcursorpos>`__

Interface.getPrompt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.getPrompt()

*Kind:* function · *Arguments:* 0

The ``rl.getPrompt()`` method returns the current prompt used by ``rl.prompt()``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlgetprompt>`__

Interface.pause
~~~~~~~~~~~~~~~

.. code-block:: text

   rl.pause()

*Kind:* function · *Arguments:* 0

The ``rl.pause()`` method pauses the ``input`` stream, allowing it to be resumed later if necessary.

Calling ``rl.pause()`` does not immediately pause other events (including ``'line'``) from being emitted by the ``InterfaceConstructor`` instance.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlpause>`__

Interface.prompt
~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.prompt([preserveCursor])

*Kind:* function · *Arguments:* 1

The ``rl.prompt()`` method writes the ``InterfaceConstructor`` instances configured ``prompt`` to a new line in ``output`` in order to provide a user with a new location at which to provide input.

When called, ``rl.prompt()`` will resume the ``input`` stream if it has been paused.

**Arguments**

* ``preserveCursor`` (*boolean*) — optional: If ``true``, prevents the cursor placement from

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlpromptpreservecursor>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlquestionquery-options-callback>`__

Interface.resume
~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.resume()

*Kind:* function · *Arguments:* 0

The ``rl.resume()`` method resumes the ``input`` stream if it has been paused.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlresume>`__

Interface.setPrompt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   rl.setPrompt(prompt)

*Kind:* function · *Arguments:* 1

The ``rl.setPrompt()`` method sets the prompt that will be written to ``output`` whenever ``rl.prompt()`` is called.

**Arguments**

* ``prompt`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlsetpromptprompt>`__

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

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/readline.html#rlwritedata-key>`__

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
