Documentation
=============

Guides · reference · runtime

Start with the command line, learn how GodeJS loads modules, then jump into the
generated API reference or the runtime internals.

Guides
------

* :doc:`Command line <cli>` — Files, eval, stdin, watch mode, reports, and the
  exact refusal list.
* :doc:`REPL <repl>` — Evaluate expressions, inspect values, and control the
  session.
* :doc:`Modules <modules>` — CommonJS, ESM, package exports, imports maps, and
  symlinks.
* :doc:`Debugger <debugger>` — Breakpoints, stepping, locals, watches, and
  real-scope evaluation.

Reference
---------

* :doc:`API reference <api/index>` — Every module and global in the runtime,
  with signatures and examples.
* :doc:`Global objects <api/globals>` — Buffer, fetch, Web Storage, streams,
  URL, and the CommonJS globals.

Runtime and platforms
---------------------

* :doc:`Platforms <../platforms>` — macOS installers, Linux dependencies, and
  source builds.
* :doc:`JavaScriptCore <../javascriptcore>` — How the runtime maps Node APIs
  onto JavaScriptCore.
* :doc:`HTTP/3 <../http3>` — Run a QUIC server with Node-compatible request
  handling.
* :doc:`Writing a builtin <../module-guide>` — The package contract and test
  helpers for new core modules.

.. note::

   GodeJS targets Node's JavaScript-visible APIs, not its engine internals.
   V8 flags, the inspector protocol, OpenSSL configuration, snapshots, SEA,
   permissions, VFS, FFI, the benchmark runner, and the Node test runner are
   not available. See :doc:`cli` for the exact refusal list.

.. toctree::
   :hidden:
   :maxdepth: 2

   cli
   repl
   modules
   debugger
