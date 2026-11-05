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

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.Recoverable(undefined);
   console.log(instance);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


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


REPLServer class
----------------

**REPLServer** — Instances of ``repl.REPLServer`` are created using the ``repl.start()`` method or directly using the JavaScript ``new`` keyword.

Calling ``repl.REPLServer()`` without the ``new`` keyword throws a ``TypeError`` (see DEP0185).

REPLServer._help
~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer._help(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   console.log(typeof repl.REPLServer.prototype._help()); // 'function'
   console.log(repl.REPLServer.prototype._help().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


REPLServer._onLine
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer._onLine(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   console.log(typeof repl.REPLServer.prototype._onLine()); // 'function'
   console.log(repl.REPLServer.prototype._onLine().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


REPLServer.clearBufferedCommand
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   replServer.clearBufferedCommand()

*Kind:* function · *Arguments:* 0

The ``replServer.clearBufferedCommand()`` method clears any command that has been buffered but not yet executed. This method is primarily intended to be called from within the action function for commands registered using the ``replServer.defineCommand()`` method.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.REPLServer();
   const result = instance.clearBufferedCommand();
   console.log(result);


REPLServer.close
~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.close(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.REPLServer();
   const result = instance.close();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


REPLServer.defaultEval
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.defaultEval(arg0, arg1, arg2, arg3)

*Kind:* function · *Arguments:* 4

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.REPLServer();
   const result = instance.defaultEval(undefined, undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


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


REPLServer.displayPrompt
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   replServer.displayPrompt([preserveCursor])

*Kind:* function · *Arguments:* 1

The ``replServer.displayPrompt()`` method readies the REPL instance for input from the user, printing the configured ``prompt`` to a new line in the ``output`` and resuming the ``input`` to accept new input.

When multi-line input is being entered, a pipe ``'|'`` is printed rather than the 'prompt'.

**Arguments**

* ``preserveCursor`` (*boolean*) — optional

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.REPLServer();
   const result = instance.displayPrompt(true);
   console.log(result);


REPLServer.setupContext
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   repl.REPLServer.setupContext(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const repl = require('node:repl');

   const instance = new repl.REPLServer();
   const result = instance.setupContext(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``builtinModules`` — array
* ``REPL_MODE_SLOPPY`` — number
* ``REPL_MODE_STRICT`` — number

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
