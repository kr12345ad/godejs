API Reference
=============

Every callable in the runtime is listed here. The reference is generated from
the running GodeJS binary, so signatures, argument counts, and return values
describe the API you actually have, not a hoped-for one.

Some prose and examples are adapted from MIT-licensed Node.js documentation;
see the repository `NOTICE <https://github.com/kr12345ad/godejs/blob/main/NOTICE>`_
file.

Global objects
--------------

Host globals, Web APIs, and CommonJS globals.

* :doc:`Global objects <globals>` — Host globals, Web APIs, and the CommonJS
  globals GodeJS exposes. (300 callables)

Core runtime
------------

Everyday modules for files, paths, processes, streams, and utilities.

* :doc:`assert <assert>` — Assertions for tests and invariants. (47 callables)
* :doc:`buffer <buffer>` — Binary data with Buffer and typed-array views. (109 callables)
* :doc:`console <console>` — Formatted output, timers, and counters. (23 callables)
* :doc:`events <events>` — EventEmitter and event-driven APIs. (54 callables)
* :doc:`fs <fs>` — Files, directories, streams, and file descriptors. (207 callables)
* :doc:`path <path>` — Cross-platform filesystem paths. (36 callables)
* :doc:`process <process>` — The current process, environment, and streams. (97 callables)
* :doc:`querystring <querystring>` — Parse and serialize URL query strings. (7 callables)
* :doc:`stream <stream>` — Readable, writable, transform, and duplex streams. (86 callables)
* :doc:`string_decoder <string_decoder>` — Decode Buffers into strings across
  chunks. (4 callables)
* :doc:`timers <timers>` — Timeouts, intervals, and immediate callbacks. (21 callables)
* :doc:`tty <tty>` — Terminal detection and raw-mode control. (18 callables)
* :doc:`url <url>` — Parse, build, and resolve URLs. (30 callables)
* :doc:`util <util>` — Formatting, inspection, promises, and type helpers. (108 callables)

Networking
----------

Clients and servers for HTTP, TLS, TCP, UDP, DNS, and QUIC.

* :doc:`dgram <dgram>` — UDP sockets and datagrams. (22 callables)
* :doc:`dns <dns>` — Resolve hostnames and IP addresses. (53 callables)
* :doc:`http <http>` — HTTP/1.1 clients, servers, and messages. (73 callables)
* :doc:`http2 <http2>` — HTTP/2 sessions, streams, and servers. (78 callables)
* :doc:`http3 <http3>` — HTTP/3 servers built on QUIC. (53 callables)
* :doc:`https <https>` — TLS-backed HTTP clients and servers. (73 callables)
* :doc:`net <net>` — TCP and IPC sockets and servers. (41 callables)
* :doc:`tls <tls>` — TLS sockets, servers, and secure contexts. (29 callables)

Concurrency and loading
-----------------------

Workers, child processes, module loading, testing, and storage.

* :doc:`async_hooks <async_hooks>` — Track asynchronous resources and their
  lifetimes. (17 callables)
* :doc:`child_process <child_process>` — Run subprocesses and connect to their
  pipes. (24 callables)
* :doc:`cluster <cluster>` — Fork workers that share a listening server. (8 callables)
* :doc:`module <module>` — CommonJS and ES module loading and metadata. (34 callables)
* :doc:`os <os>` — Host CPU, memory, network, and platform details. (20 callables)
* :doc:`readline <readline>` — Line-oriented terminal input and prompts. (20 callables)
* :doc:`repl <repl>` — Interactive JavaScript evaluation. (11 callables)
* :doc:`sqlite <sqlite>` — Embedded SQLite databases. (31 callables)
* :doc:`test <test>` — Write and run tests with the ``node:test`` harness. (26 callables)
* :doc:`vm <vm>` — Run JavaScript in isolated contexts. (27 callables)
* :doc:`worker_threads <worker_threads>` — Run JavaScript on additional threads.
  (28 callables)

Platform and diagnostics
------------------------

Crypto, performance, tracing, terminal, and host integration.

* :doc:`constants <constants>` — Operating-system and signal constants. (0 callables)
* :doc:`crypto <crypto>` — Hashes, ciphers, signatures, and random bytes. (156 callables)
* :doc:`diagnostics_channel <diagnostics_channel>` — Named channels for runtime
  diagnostics. (5 callables)
* :doc:`domain <domain>` — Legacy error-handling domains. (12 callables)
* :doc:`inspector <inspector>` — Debugger and inspector session APIs. (10 callables)
* :doc:`perf_hooks <perf_hooks>` — Performance marks, measures, and observers.
  (34 callables)
* :doc:`punycode <punycode>` — Unicode and ASCII domain-name conversion. (6 callables)
* :doc:`sea <sea>` — Single executable application helpers. (4 callables)
* :doc:`v8 <v8>` — V8-compatible inspection and serialization helpers. (67 callables)
* :doc:`wasi <wasi>` — WASI preview 1 applications. (5 callables)
* :doc:`zlib <zlib>` — Compress and decompress streams and buffers. (81 callables)

.. note::

   Entries with a runtime-only note are still callable, but their argument names
   and descriptions may be incomplete. The snapshot is the source of truth for
   what exists.

.. toctree::
   :hidden:
   :maxdepth: 1

   globals
   assert
   async_hooks
   buffer
   child_process
   cluster
   console
   constants
   crypto
   dgram
   diagnostics_channel
   dns
   domain
   events
   fs
   http
   http2
   http3
   https
   inspector
   module
   net
   os
   path
   perf_hooks
   process
   punycode
   querystring
   readline
   repl
   sea
   sqlite
   stream
   string_decoder
   test
   timers
   tls
   tty
   url
   util
   v8
   vm
   wasi
   worker_threads
   zlib
