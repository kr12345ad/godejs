API Reference
=============

This reference is generated from the running GodeJS runtime, not hand-maintained. Every callable the runtime exposes has its own entry with a signature, typed arguments, return value, and an example where one is available. Private helpers and GodeJS extensions are labelled explicitly.

Signatures, argument names, return values, and examples reflect the GodeJS runtime surface. The runtime is the source of truth for what is available here.

Some prose and examples are adapted from MIT-licensed material; see the repository `NOTICE <https://github.com/kr12345ad/godejs/blob/main/NOTICE>`_ file for the applicable notice.

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
