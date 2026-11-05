API Reference
=============

Every callable in the runtime is listed here. The reference is generated from the running GodeJS binary, so signatures, argument counts, and return values describe the API you actually have, not a hoped-for one.

Some prose and examples are adapted from MIT-licensed Node.js documentation; see the repository `NOTICE <https://github.com/kr12345ad/godejs/blob/main/NOTICE>`_ file.

.. raw:: html

   <section class="api-toolbar" aria-label="Filter API modules">
     <div><span class="api-toolbar-kicker">Browse by module</span>
     <h2>Find an API surface</h2></div>
     <div class="api-filter-wrap">
       <label class="sr-only" for="api-filter">Filter modules</label>
       <input id="api-filter" type="search" placeholder="Try fs, http, crypto…" autocomplete="off">
       <span id="api-count" class="api-count">45 modules</span>
     </div>
   </section>
   <section class="api-section" data-api-section>
     <div class="api-section-heading"><div><span class="api-section-kicker">Everywhere</span><h2>Global objects</h2></div><p>Host globals, Web APIs, and CommonJS globals.</p></div>
     <div class="api-grid"><a class="api-card" href="globals.html" data-search="Global objects __globals__ Host globals, Web APIs, and the CommonJS globals GodeJS exposes."><span class="api-card-title">Global objects</span><span class="api-card-blurb">Host globals, Web APIs, and the CommonJS globals GodeJS exposes.</span><span class="api-card-meta">300 callables</span></a></div>
   </section>
   <section class="api-section" data-api-section>
     <div class="api-section-heading"><div><span class="api-section-kicker">Collection</span><h2>Core runtime</h2></div><p>Everyday modules for files, paths, processes, streams, and utilities.</p></div>
     <div class="api-grid">
       <a class="api-card" href="assert.html" data-search="assert assert Assertions for tests and invariants."><span class="api-card-title">assert</span><span class="api-card-blurb">Assertions for tests and invariants.</span><span class="api-card-meta">47 callables</span></a>
       <a class="api-card" href="buffer.html" data-search="buffer buffer Binary data with Buffer and typed-array views."><span class="api-card-title">buffer</span><span class="api-card-blurb">Binary data with Buffer and typed-array views.</span><span class="api-card-meta">109 callables</span></a>
       <a class="api-card" href="console.html" data-search="console console Formatted output, timers, and counters."><span class="api-card-title">console</span><span class="api-card-blurb">Formatted output, timers, and counters.</span><span class="api-card-meta">23 callables</span></a>
       <a class="api-card" href="events.html" data-search="events events EventEmitter and event-driven APIs."><span class="api-card-title">events</span><span class="api-card-blurb">EventEmitter and event-driven APIs.</span><span class="api-card-meta">54 callables</span></a>
       <a class="api-card" href="fs.html" data-search="fs fs Files, directories, streams, and file descriptors."><span class="api-card-title">fs</span><span class="api-card-blurb">Files, directories, streams, and file descriptors.</span><span class="api-card-meta">207 callables</span></a>
       <a class="api-card" href="path.html" data-search="path path Cross-platform filesystem paths."><span class="api-card-title">path</span><span class="api-card-blurb">Cross-platform filesystem paths.</span><span class="api-card-meta">36 callables</span></a>
       <a class="api-card" href="process.html" data-search="process process The current process, environment, and streams."><span class="api-card-title">process</span><span class="api-card-blurb">The current process, environment, and streams.</span><span class="api-card-meta">97 callables</span></a>
       <a class="api-card" href="querystring.html" data-search="querystring querystring Parse and serialize URL query strings."><span class="api-card-title">querystring</span><span class="api-card-blurb">Parse and serialize URL query strings.</span><span class="api-card-meta">7 callables</span></a>
       <a class="api-card" href="stream.html" data-search="stream stream Readable, writable, transform, and duplex streams."><span class="api-card-title">stream</span><span class="api-card-blurb">Readable, writable, transform, and duplex streams.</span><span class="api-card-meta">86 callables</span></a>
       <a class="api-card" href="string_decoder.html" data-search="string_decoder string_decoder Decode Buffers into strings across chunks."><span class="api-card-title">string_decoder</span><span class="api-card-blurb">Decode Buffers into strings across chunks.</span><span class="api-card-meta">4 callables</span></a>
       <a class="api-card" href="timers.html" data-search="timers timers Timeouts, intervals, and immediate callbacks."><span class="api-card-title">timers</span><span class="api-card-blurb">Timeouts, intervals, and immediate callbacks.</span><span class="api-card-meta">21 callables</span></a>
       <a class="api-card" href="tty.html" data-search="tty tty Terminal detection and raw-mode control."><span class="api-card-title">tty</span><span class="api-card-blurb">Terminal detection and raw-mode control.</span><span class="api-card-meta">18 callables</span></a>
       <a class="api-card" href="url.html" data-search="url url Parse, build, and resolve URLs."><span class="api-card-title">url</span><span class="api-card-blurb">Parse, build, and resolve URLs.</span><span class="api-card-meta">30 callables</span></a>
       <a class="api-card" href="util.html" data-search="util util Formatting, inspection, promises, and type helpers."><span class="api-card-title">util</span><span class="api-card-blurb">Formatting, inspection, promises, and type helpers.</span><span class="api-card-meta">108 callables</span></a>
     </div>
   </section>
   <section class="api-section" data-api-section>
     <div class="api-section-heading"><div><span class="api-section-kicker">Collection</span><h2>Networking</h2></div><p>Clients and servers for HTTP, TLS, TCP, UDP, DNS, and QUIC.</p></div>
     <div class="api-grid">
       <a class="api-card" href="dgram.html" data-search="dgram dgram UDP sockets and datagrams."><span class="api-card-title">dgram</span><span class="api-card-blurb">UDP sockets and datagrams.</span><span class="api-card-meta">22 callables</span></a>
       <a class="api-card" href="dns.html" data-search="dns dns Resolve hostnames and IP addresses."><span class="api-card-title">dns</span><span class="api-card-blurb">Resolve hostnames and IP addresses.</span><span class="api-card-meta">53 callables</span></a>
       <a class="api-card" href="http.html" data-search="http http HTTP/1.1 clients, servers, and messages."><span class="api-card-title">http</span><span class="api-card-blurb">HTTP/1.1 clients, servers, and messages.</span><span class="api-card-meta">73 callables</span></a>
       <a class="api-card" href="http2.html" data-search="http2 http2 HTTP/2 sessions, streams, and servers."><span class="api-card-title">http2</span><span class="api-card-blurb">HTTP/2 sessions, streams, and servers.</span><span class="api-card-meta">78 callables</span></a>
       <a class="api-card" href="http3.html" data-search="http3 http3 HTTP/3 servers built on QUIC."><span class="api-card-title">http3</span><span class="api-card-blurb">HTTP/3 servers built on QUIC.</span><span class="api-card-meta">53 callables</span></a>
       <a class="api-card" href="https.html" data-search="https https TLS-backed HTTP clients and servers."><span class="api-card-title">https</span><span class="api-card-blurb">TLS-backed HTTP clients and servers.</span><span class="api-card-meta">73 callables</span></a>
       <a class="api-card" href="net.html" data-search="net net TCP and IPC sockets and servers."><span class="api-card-title">net</span><span class="api-card-blurb">TCP and IPC sockets and servers.</span><span class="api-card-meta">41 callables</span></a>
       <a class="api-card" href="tls.html" data-search="tls tls TLS sockets, servers, and secure contexts."><span class="api-card-title">tls</span><span class="api-card-blurb">TLS sockets, servers, and secure contexts.</span><span class="api-card-meta">29 callables</span></a>
     </div>
   </section>
   <section class="api-section" data-api-section>
     <div class="api-section-heading"><div><span class="api-section-kicker">Collection</span><h2>Concurrency and loading</h2></div><p>Workers, child processes, module loading, testing, and storage.</p></div>
     <div class="api-grid">
       <a class="api-card" href="async_hooks.html" data-search="async_hooks async_hooks Track asynchronous resources and their lifetimes."><span class="api-card-title">async_hooks</span><span class="api-card-blurb">Track asynchronous resources and their lifetimes.</span><span class="api-card-meta">17 callables</span></a>
       <a class="api-card" href="child_process.html" data-search="child_process child_process Run subprocesses and connect to their pipes."><span class="api-card-title">child_process</span><span class="api-card-blurb">Run subprocesses and connect to their pipes.</span><span class="api-card-meta">24 callables</span></a>
       <a class="api-card" href="cluster.html" data-search="cluster cluster Fork workers that share a listening server."><span class="api-card-title">cluster</span><span class="api-card-blurb">Fork workers that share a listening server.</span><span class="api-card-meta">8 callables</span></a>
       <a class="api-card" href="module.html" data-search="module module CommonJS and ES module loading and metadata."><span class="api-card-title">module</span><span class="api-card-blurb">CommonJS and ES module loading and metadata.</span><span class="api-card-meta">34 callables</span></a>
       <a class="api-card" href="os.html" data-search="os os Host CPU, memory, network, and platform details."><span class="api-card-title">os</span><span class="api-card-blurb">Host CPU, memory, network, and platform details.</span><span class="api-card-meta">20 callables</span></a>
       <a class="api-card" href="readline.html" data-search="readline readline Line-oriented terminal input and prompts."><span class="api-card-title">readline</span><span class="api-card-blurb">Line-oriented terminal input and prompts.</span><span class="api-card-meta">20 callables</span></a>
       <a class="api-card" href="repl.html" data-search="repl repl Interactive JavaScript evaluation."><span class="api-card-title">repl</span><span class="api-card-blurb">Interactive JavaScript evaluation.</span><span class="api-card-meta">11 callables</span></a>
       <a class="api-card" href="sqlite.html" data-search="sqlite sqlite Embedded SQLite databases."><span class="api-card-title">sqlite</span><span class="api-card-blurb">Embedded SQLite databases.</span><span class="api-card-meta">31 callables</span></a>
       <a class="api-card" href="test.html" data-search="test (node:test) test Write and run tests with the node:test harness."><span class="api-card-title">test (node:test)</span><span class="api-card-blurb">Write and run tests with the node:test harness.</span><span class="api-card-meta">26 callables</span></a>
       <a class="api-card" href="vm.html" data-search="vm vm Run JavaScript in isolated contexts."><span class="api-card-title">vm</span><span class="api-card-blurb">Run JavaScript in isolated contexts.</span><span class="api-card-meta">27 callables</span></a>
       <a class="api-card" href="worker_threads.html" data-search="worker_threads worker_threads Run JavaScript on additional threads."><span class="api-card-title">worker_threads</span><span class="api-card-blurb">Run JavaScript on additional threads.</span><span class="api-card-meta">28 callables</span></a>
     </div>
   </section>
   <section class="api-section" data-api-section>
     <div class="api-section-heading"><div><span class="api-section-kicker">Collection</span><h2>Platform and diagnostics</h2></div><p>Crypto, performance, tracing, terminal, and host integration.</p></div>
     <div class="api-grid">
       <a class="api-card" href="constants.html" data-search="constants constants Operating-system and signal constants."><span class="api-card-title">constants</span><span class="api-card-blurb">Operating-system and signal constants.</span><span class="api-card-meta">0 callables</span></a>
       <a class="api-card" href="crypto.html" data-search="crypto crypto Hashes, ciphers, signatures, and random bytes."><span class="api-card-title">crypto</span><span class="api-card-blurb">Hashes, ciphers, signatures, and random bytes.</span><span class="api-card-meta">156 callables</span></a>
       <a class="api-card" href="diagnostics_channel.html" data-search="diagnostics_channel diagnostics_channel Named channels for runtime diagnostics."><span class="api-card-title">diagnostics_channel</span><span class="api-card-blurb">Named channels for runtime diagnostics.</span><span class="api-card-meta">5 callables</span></a>
       <a class="api-card" href="domain.html" data-search="domain domain Legacy error-handling domains."><span class="api-card-title">domain</span><span class="api-card-blurb">Legacy error-handling domains.</span><span class="api-card-meta">12 callables</span></a>
       <a class="api-card" href="inspector.html" data-search="inspector inspector Debugger and inspector session APIs."><span class="api-card-title">inspector</span><span class="api-card-blurb">Debugger and inspector session APIs.</span><span class="api-card-meta">10 callables</span></a>
       <a class="api-card" href="perf_hooks.html" data-search="perf_hooks perf_hooks Performance marks, measures, and observers."><span class="api-card-title">perf_hooks</span><span class="api-card-blurb">Performance marks, measures, and observers.</span><span class="api-card-meta">34 callables</span></a>
       <a class="api-card" href="punycode.html" data-search="punycode punycode Unicode and ASCII domain-name conversion."><span class="api-card-title">punycode</span><span class="api-card-blurb">Unicode and ASCII domain-name conversion.</span><span class="api-card-meta">6 callables</span></a>
       <a class="api-card" href="sea.html" data-search="sea sea Single executable application helpers."><span class="api-card-title">sea</span><span class="api-card-blurb">Single executable application helpers.</span><span class="api-card-meta">4 callables</span></a>
       <a class="api-card" href="v8.html" data-search="v8 v8 V8-compatible inspection and serialization helpers."><span class="api-card-title">v8</span><span class="api-card-blurb">V8-compatible inspection and serialization helpers.</span><span class="api-card-meta">67 callables</span></a>
       <a class="api-card" href="wasi.html" data-search="wasi wasi WASI preview 1 applications."><span class="api-card-title">wasi</span><span class="api-card-blurb">WASI preview 1 applications.</span><span class="api-card-meta">5 callables</span></a>
       <a class="api-card" href="zlib.html" data-search="zlib zlib Compress and decompress streams and buffers."><span class="api-card-title">zlib</span><span class="api-card-blurb">Compress and decompress streams and buffers.</span><span class="api-card-meta">81 callables</span></a>
     </div>
   </section>
   <p id="api-empty" class="api-empty" hidden>No modules match that filter. Try a shorter term.</p>

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

.. note::

   Entries with a runtime-only note are still callable, but their argument names and descriptions may be incomplete. The snapshot is the source of truth for what exists.
