REPL
====

Run ``gode`` with no script and a terminal on stdin:

.. code-block:: sh

   $ gode
   Welcome to GodeJS v24.22.2!
   Type ".help" for more information.
   > 1 + 1
   2
   > const path = require('node:path')
   undefined
   > path.join('a', 'b')
   'a/b'
   > _

The last expression value is stored in ``_``. ``undefined`` results are not
printed.

Multiline input
---------------

The REPL continues on the ``...`` prompt while brackets, braces, strings,
template literals, block comments, or a trailing backslash are open:

.. code-block:: text

   > function add(a, b) {
   ...   return a + b;
   ... }
   undefined
   > add(2, 3)
   5

A syntax error whose message says the input ended unexpectedly also keeps the
buffer open, so a missing ``}`` can be completed on the next line.

Commands
--------

.. list-table::
   :header-rows: 1
   :widths: 20 80

   * - Command
     - Effect
   * - ``.help``
     - Print the command list.
   * - ``.clear``
     - Discard the current multiline buffer.
   * - ``.exit`` or ``exit``
     - Leave the REPL with status 0.
   * - ``Ctrl-D``
     - Leave the REPL when the current line is empty.

``Ctrl-C`` is handled by the terminal; the REPL itself does not install a
signal handler while reading a line.

Colors and prompts
------------------

Prompts are green, continuation prompts are yellow, and errors are red when
stdout is a terminal. Color is disabled when ``NO_COLOR`` is set, when
``TERM=dumb``, or when output is piped:

.. code-block:: sh

   echo '1 + 1' | gode

That prints the result without prompts or color.

Interactive use after a script
------------------------------

``-i`` runs the script and then enters the REPL with the script's globals
still installed:

.. code-block:: sh

   gode -i app.js

``--eval`` and ``--input-type`` work with the REPL as well:

.. code-block:: sh

   gode --input-type=module -e 'import os from "node:os"; console.log(os.platform())'

Limitations
-----------

The REPL is deliberately small. It does not provide persistent history,
reverse search, tab completion, or top-level ``await``. Feed it a script or
use ``--eval`` for automation. Custom REPL servers can be built with the
``node:repl`` module documented in the API reference.
