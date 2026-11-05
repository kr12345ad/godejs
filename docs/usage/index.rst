Documentation
=============

.. raw:: html

   <section class="hero">
     <p class="eyebrow">Guides · reference · runtime</p>
     <p class="hero-lede">Start with the command line, learn how GodeJS loads modules, then jump into the generated API reference or the runtime internals.</p>
   </section>

.. raw:: html

   <div class="section-grid">
     <a class="doc-card" href="cli.html">
       <span class="card-kicker">Start here</span>
       <h3>Command line</h3>
       <p>Files, eval, stdin, watch mode, reports, and the exact refusal list.</p>
       <span class="card-link">Run your first script →</span>
     </a>
     <a class="doc-card" href="repl.html">
       <span class="card-kicker">Interactive</span>
       <h3>REPL</h3>
       <p>Evaluate expressions, inspect values, and control the session.</p>
       <span class="card-link">Open the REPL guide →</span>
     </a>
     <a class="doc-card" href="modules.html">
       <span class="card-kicker">Core concepts</span>
       <h3>Modules</h3>
       <p>CommonJS, ESM, package exports, imports maps, and symlinks.</p>
       <span class="card-link">Understand module loading →</span>
     </a>
     <a class="doc-card" href="debugger.html">
       <span class="card-kicker">Tooling</span>
       <h3>Debugger</h3>
       <p>Breakpoints, stepping, locals, watches, and real-scope evaluation.</p>
       <span class="card-link">Debug a script →</span>
     </a>
   </div>

Reference
---------

.. raw:: html

   <div class="section-grid">
     <a class="doc-card" href="api/index.html">
       <span class="card-kicker">Everything callable</span>
       <h3>API reference</h3>
       <p>Every module and global in the runtime, with signatures and examples.</p>
       <span class="card-link">Browse the API →</span>
     </a>
     <a class="doc-card" href="api/globals.html">
       <span class="card-kicker">Host surface</span>
       <h3>Global objects</h3>
       <p>Buffer, fetch, Web Storage, streams, URL, and the CommonJS globals.</p>
       <span class="card-link">Inspect globals →</span>
     </a>
   </div>

Runtime and platforms
---------------------

.. raw:: html

   <div class="section-grid">
     <a class="doc-card" href="../platforms.html">
       <span class="card-kicker">Deployment</span>
       <h3>Platforms</h3>
       <p>macOS installers, Linux dependencies, and source builds.</p>
       <span class="card-link">Check your platform →</span>
     </a>
     <a class="doc-card" href="../javascriptcore.html">
       <span class="card-kicker">Engine</span>
       <h3>JavaScriptCore</h3>
       <p>How the runtime maps Node APIs onto JavaScriptCore.</p>
       <span class="card-link">Read the engine notes →</span>
     </a>
     <a class="doc-card" href="../http3.html">
       <span class="card-kicker">Extension</span>
       <h3>HTTP/3</h3>
       <p>Run a QUIC server with Node-compatible request handling.</p>
       <span class="card-link">Open the HTTP/3 guide →</span>
     </a>
     <a class="doc-card" href="../module-guide.html">
       <span class="card-kicker">Contributing</span>
       <h3>Writing a builtin</h3>
       <p>The package contract and test helpers for new core modules.</p>
       <span class="card-link">See the module guide →</span>
     </a>
   </div>

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
