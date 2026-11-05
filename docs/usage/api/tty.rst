tty
===

The ``node:tty`` module provides the ``tty.ReadStream`` and ``tty.WriteStream`` classes. In most cases, it will not be necessary or possible to use this module directly. However, it can be accessed using:

When Node.js detects that it is being run with a text terminal ("TTY") attached, ``process.stdin`` will, by default, be initialized as an instance of ``tty.ReadStream`` and both ``process.stdout`` and ``process.stderr`` will, by default, be instances of ``tty.WriteStream``. The preferred method of determining whether Node.js is being run within a TTY context is to check that the value of the ``process.stdout.isTTY`` property is ``true``:

Loading
-------

.. code-block:: javascript

   const tty = require('node:tty');
   // ESM
   import * as tty from 'node:tty';

Quick start
-----------

.. code-block:: javascript

   const tty = require('node:tty');

API
---

isatty
~~~~~~

.. code-block:: text

   tty.isatty(fd)

*Kind:* function · *Arguments:* 1

The ``tty.isatty()`` method returns ``true`` if the given ``fd`` is associated with a TTY and ``false`` if it is not, including whenever ``fd`` is not a non-negative integer.

**Arguments**

* ``fd`` (*number*): A numeric file descriptor


ReadStream
~~~~~~~~~~

.. code-block:: text

   tty.ReadStream

*Kind:* class · *Arguments:* 2

Represents the readable side of a TTY. In normal circumstances ``process.stdin`` will be the only ``tty.ReadStream`` instance in a Node.js process and there should be no reason to create additional instances.

This is a class; construct instances with ``new``.


WriteStream
~~~~~~~~~~~

.. code-block:: text

   tty.WriteStream

*Kind:* class · *Arguments:* 2

Represents the writable side of a TTY. In normal circumstances, ``process.stdout`` and ``process.stderr`` will be the only ``tty.WriteStream`` instances created for a Node.js process and there should be no reason to create additional instances.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   process.stdout.on('resize', () => {
     console.log('screen size has changed!');
     console.log(`${process.stdout.columns}x${process.stdout.rows}`);
   });


ReadStream class
----------------

**ReadStream** — Represents the readable side of a TTY. In normal circumstances ``process.stdin`` will be the only ``tty.ReadStream`` instance in a Node.js process and there should be no reason to create additional instances.

ReadStream.destroy
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.ReadStream.destroy(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ReadStream.pause
~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.ReadStream.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ReadStream.resume
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.ReadStream.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ReadStream.setRawMode
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readStream.setRawMode(mode)

*Kind:* function · *Arguments:* 1

Allows configuration of ``tty.ReadStream`` so that it operates as a raw device.

When in raw mode, input is always available character-by-character, not including modifiers. Additionally, all special processing of input characters by the terminal is disabled, including echoing input characters. <kbd>Ctrl</kbd>+<kbd>C</kbd> will no longer cause a ``SIGINT`` when in this mode. This mode does not affect terminal output processing, such as newline translation on Unix terminals.

**Arguments**

* ``mode`` (*boolean|string*): If ``true`` or ``'raw'``, configures the


WriteStream class
-----------------

**WriteStream** — Represents the writable side of a TTY. In normal circumstances, ``process.stdout`` and ``process.stderr`` will be the only ``tty.WriteStream`` instances created for a Node.js process and there should be no reason to create additional instances.

WriteStream._refreshSize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.WriteStream._refreshSize(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WriteStream.clearLine
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.clearLine(dir[, callback])

*Kind:* function · *Arguments:* 2

``writeStream.clearLine()`` clears the current line of this ``WriteStream`` in a direction identified by ``dir``.

**Arguments**

* ``dir`` (*number*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.


WriteStream.clearScreenDown
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.clearScreenDown([callback])

*Kind:* function · *Arguments:* 1

``writeStream.clearScreenDown()`` clears this ``WriteStream`` from the current cursor down.

**Arguments**

* ``callback`` (*Function*) — optional: Invoked once the operation completes.


WriteStream.cursorTo
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.cursorTo(x[, y][, callback])

*Kind:* function · *Arguments:* 3

``writeStream.cursorTo()`` moves this ``WriteStream``'s cursor to the specified position.

**Arguments**

* ``x`` (*number*)
* ``y`` (*number*) — optional
* ``callback`` (*Function*) — optional: Invoked once the operation completes.


WriteStream.destroy
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.WriteStream.destroy(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WriteStream.end
~~~~~~~~~~~~~~~

.. code-block:: text

   tty.WriteStream.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WriteStream.getColorDepth
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.getColorDepth([env])

*Kind:* function · *Arguments:* 1

Returns:

* ``1`` for 2, * ``4`` for 16, * ``8`` for 256, * ``24`` for 16,777,216 colors supported.

**Arguments**

* ``env`` (*Object*) — optional: An object containing the environment variables to check. This


WriteStream.getWindowSize
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.getWindowSize()

*Kind:* function · *Arguments:* 0

``writeStream.getWindowSize()`` returns the size of the TTY corresponding to this ``WriteStream``. The array is of the type ``[numColumns, numRows]`` where ``numColumns`` and ``numRows`` represent the number of columns and rows in the corresponding TTY.

This callable takes no arguments.


WriteStream.hasColors
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.hasColors([count][, env])

*Kind:* function · *Arguments:* 2

Returns ``true`` if the ``writeStream`` supports at least as many colors as provided in ``count``. Minimum support is 2 (black and white).

This has the same false positives and negatives as described in ``writeStream.getColorDepth()``.

**Arguments**

* ``count`` (*integer*) — optional: The number of colors that are requested (minimum 2).
* ``env`` (*Object*) — optional: An object containing the environment variables to check. This

**Example**

.. code-block:: javascript

   process.stdout.hasColors();
   // Returns true or false depending on if `stdout` supports at least 16 colors.
   process.stdout.hasColors(256);
   // Returns true or false depending on if `stdout` supports at least 256 colors.
   process.stdout.hasColors({ TMUX: '1' });
   // Returns true.
   process.stdout.hasColors(2 ** 24, { TMUX: '1' });
   // Returns false (the environment setting pretends to support 2 ** 8 colors).


WriteStream.moveCursor
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.moveCursor(dx, dy[, callback])

*Kind:* function · *Arguments:* 3

``writeStream.moveCursor()`` moves this ``WriteStream``'s cursor *relative* to its current position.

**Arguments**

* ``dx`` (*number*)
* ``dy`` (*number*)
* ``callback`` (*Function*) — optional: Invoked once the operation completes.


WriteStream.write
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tty.WriteStream.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``ReadStream.isTTY`` — boolean
* ``WriteStream.isTTY`` — boolean

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
