API Reference
=============

This reference is generated from the running runtime, not hand-maintained. Every callable the runtime exposes has its own entry with a signature, typed arguments, return value, and an example where the upstream Node.js documentation provides one. Private helpers and GodeJS extensions are labelled explicitly.

Signatures and prose are derived from the Node.js v24.22.2 documentation, licensed under the MIT License. The runtime surface is the source of truth for what is available here.

Global objects
--------------

.. toctree::
   :maxdepth: 1

   globals

Core runtime
------------

.. toctree::
   :maxdepth: 1

   assert
   buffer
   console
   events
   fs
   path
   process
   querystring
   stream
   string_decoder
   timers
   tty
   url
   util

Networking
----------

.. toctree::
   :maxdepth: 1

   dgram
   dns
   http
   http2
   http3
   https
   net
   tls

Concurrency and loading
-----------------------

.. toctree::
   :maxdepth: 1

   async_hooks
   child_process
   cluster
   module
   os
   readline
   repl
   sqlite
   test
   vm
   worker_threads

Platform and diagnostics
------------------------

.. toctree::
   :maxdepth: 1

   constants
   crypto
   diagnostics_channel
   domain
   inspector
   perf_hooks
   punycode
   sea
   v8
   wasi
   zlib
