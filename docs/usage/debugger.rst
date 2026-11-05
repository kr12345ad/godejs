Debugger
========

GodeJS includes a source-level debugger for the entry script. Start it with
``--debug`` or ``-d``:

.. code-block:: sh

   gode --debug app.js

The debugger prints every instrumented breakpoint site, pauses at entry, and
shows the current line:

.. code-block:: text

   gode debugger: /work/app.js (12 breakpoint sites)
   Paused at entry. Type 'help' for commands, 'continue' to run.

   --- paused at /work/app.js:1
       1 | const fs = require('node:fs');
   (gode) _

The prompt is ``(gode)``. Commands are line-oriented.

Commands
--------

.. list-table::
   :header-rows: 1
   :widths: 30 70

   * - Command
     - Effect
   * - ``break <line>``, ``b <line>``
     - Set a breakpoint at a line in the current file.
   * - ``break <file:line>``, ``b <file:line>``
     - Set a breakpoint in a specific file.
   * - ``delete <line>``, ``d <line>``
     - Remove a breakpoint at the current line or the given line.
   * - ``delete <file:line>``
     - Remove a breakpoint by file and line.
   * - ``info breakpoints``, ``i``
     - List active breakpoints.
   * - ``continue``, ``c``, ``run``, ``r``
     - Run until the next breakpoint or the program exits.
   * - ``step``, ``s``, ``next``, ``n``
     - Pause at the next statement.
   * - ``locals``, ``loc``
     - Print declarations visible at the current line.
   * - ``eval <expr>``, ``p <expr>``, ``print <expr>``
     - Evaluate an expression in the paused scope.
   * - ``watch <expr>``
     - Evaluate and print an expression at every pause.
   * - ``unwatch``
     - Clear all watches.
   * - ``bt``, ``where``, ``backtrace``
     - Print the JavaScript call stack.
   * - ``list [line]``, ``l [line]``
     - Show source around the current line or the requested line.
   * - ``clear``
     - Clear every breakpoint.
   * - ``quit``, ``q``, ``exit``
     - Exit the process with status 0.
   * - ``help``, ``h``, ``?``
     - Print the same command list.

Example
-------

.. code-block:: javascript

   function total(items) {
     let sum = 0;
     for (const item of items) sum += item.price;
     return sum;
   }

   console.log(total([{ price: 2 }, { price: 3 }]));

.. code-block:: text

   $ gode --debug app.js
   Paused at entry. Type 'help' for commands, 'continue' to run.
   (gode) break 3
   breakpoint at /work/app.js:3
   (gode) continue

   --- paused at /work/app.js:3
       3 |   for (const item of items) sum += item.price;
   (gode) eval item
   { price: 2 }
   (gode) p sum
   0
   (gode) next
   (gode) c
   5

Scope and limitations
---------------------

The debugger instruments statement boundaries in the entry script and passes a
closure created in the script's lexical scope to the pause hook. ``eval`` and
``locals`` therefore see real local variables rather than a detached sandbox.
The stack is rewritten from instrumented line numbers back to the original
source.

It is not the V8 inspector: there is no remote protocol, Chrome DevTools
attach, ``--inspect``, or network debugging. Code generated at runtime with
``eval`` or ``new Function`` is not instrumented. ``--debug`` requires a script
file; it cannot be combined with ``--eval``.
