Using GodeJS
============

GodeJS runs JavaScript on JavaScriptCore and exposes a Node.js 24-compatible
standard library. These guides cover the runtime itself; the
:doc:`generated API reference <api/index>` covers every callable and global.

.. toctree::
   :maxdepth: 2

   cli
   repl
   modules
   debugger

Start with the command line if you are new:

.. code-block:: sh

   gode app.js
   gode -e "console.log(process.version)"
   gode

The reference is generated from the running runtime. Each callable has a
signature, argument list, return value, and, when one exists, an example from
the Node.js 24 documentation. Private helpers and GodeJS extensions are marked
instead of being described as if they were public Node APIs.

.. note::

   GodeJS targets Node's JavaScript-visible APIs, not its engine internals.
   V8 flags, the inspector protocol, OpenSSL configuration, snapshots, SEA,
   permissions, VFS, FFI, the benchmark runner, and the Node test runner are
   not available. See :doc:`cli` for the exact refusal list.
