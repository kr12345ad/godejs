repl
====

The ``node:repl`` module provides a Read-Eval-Print-Loop (REPL) implementation that is available both as a standalone program or includible in other applications. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const repl = require('node:repl');
   // ESM
   import * as repl from 'node:repl';

Quick start
-----------

.. code-block:: javascript

   const repl = require('node:repl');

API
---

Recoverable
~~~~~~~~~~~

.. code-block:: text

   repl.Recoverable(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


REPLServer
~~~~~~~~~~

.. code-block:: text

   REPLServer

*Kind:* class · *Arguments:* 1

Instances of ``repl.REPLServer`` are created using the ``repl.start()`` method or directly using the JavaScript ``new`` keyword.

Calling ``repl.REPLServer()`` without the ``new`` keyword throws a ``TypeError`` (see DEP0185).

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const options = { useColors: true };

   const firstInstance = repl.start(options);
   const secondInstance = new repl.REPLServer(options);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/repl.html#class-replserver>`__

start
~~~~~

.. code-block:: text

   repl.start([options])

*Kind:* function · *Arguments:* 1

The ``repl.start()`` method creates and starts a ``repl.REPLServer`` instance.

If ``options`` is a string, then it specifies the input prompt:

**Arguments**

* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   // a Unix style prompt
   repl.start('$ ');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/repl.html#replstartoptions>`__

REPLServer class
----------------

**REPLServer** — Instances of ``repl.REPLServer`` are created using the ``repl.start()`` method or directly using the JavaScript ``new`` keyword.

Calling ``repl.REPLServer()`` without the ``new`` keyword throws a ``TypeError`` (see DEP0185).

REPLServer._help
~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer._help(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


REPLServer._onLine
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer._onLine(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


REPLServer.clearBufferedCommand
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   replServer.clearBufferedCommand()

*Kind:* function · *Arguments:* 0

The ``replServer.clearBufferedCommand()`` method clears any command that has been buffered but not yet executed. This method is primarily intended to be called from within the action function for commands registered using the ``replServer.defineCommand()`` method.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/repl.html#replserverclearbufferedcommand>`__

REPLServer.close
~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.close(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


REPLServer.defaultEval
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.defaultEval(arg0, arg1, arg2, arg3)

*Kind:* function · *Arguments:* 4

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


REPLServer.defineCommand
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   replServer.defineCommand(keyword, cmd)

*Kind:* function · *Arguments:* 2

The ``replServer.defineCommand()`` method is used to add new ``.``-prefixed commands to the REPL instance. Such commands are invoked by typing a ``.`` followed by the ``keyword``. The ``cmd`` is either a ``Function`` or an ``Object`` with the following properties:

The following example shows two new commands added to the REPL instance:

**Arguments**

* ``keyword`` (*string*): The command keyword (*without* a leading ``.`` character).
* ``cmd`` (*Object|Function*): The function to invoke when the command is processed.

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const replServer = repl.start({ prompt: '> ' });
   replServer.defineCommand('sayhello', {
     help: 'Say hello',
     action(name) {
       this.clearBufferedCommand();
       console.log(`Hello, ${name}!`);
       this.displayPrompt();
     },
   });
   replServer.defineCommand('saybye', function saybye() {
     console.log('Goodbye!');
     this.close();
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/repl.html#replserverdefinecommandkeyword-cmd>`__

REPLServer.displayPrompt
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   replServer.displayPrompt([preserveCursor])

*Kind:* function · *Arguments:* 1

The ``replServer.displayPrompt()`` method readies the REPL instance for input from the user, printing the configured ``prompt`` to a new line in the ``output`` and resuming the ``input`` to accept new input.

When multi-line input is being entered, a pipe ``'|'`` is printed rather than the 'prompt'.

**Arguments**

* ``preserveCursor`` (*boolean*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/repl.html#replserverdisplaypromptpreservecursor>`__

REPLServer.setupContext
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.setupContext(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Properties and constants
------------------------

* ``builtinModules`` — array
* ``REPL_MODE_SLOPPY`` — number
* ``REPL_MODE_STRICT`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
