GodeJS
======

.. raw:: html

   <section class="hero">
     <p class="eyebrow">Node.js 24 · JavaScriptCore · Go</p>
     <p class="hero-lede">GodeJS is a Node.js 24-compatible runtime built in Go on JavaScriptCore. Bring your CommonJS and ESM apps, the standard library, and the command line you already know.</p>
     <div class="hero-actions">
       <a class="btn btn-primary" href="#install">Install GodeJS</a>
       <a class="btn btn-secondary" href="usage/api/index.html">Browse the API</a>
       <a class="btn btn-quiet" href="https://github.com/kr12345ad/godejs">View on GitHub <span aria-hidden="true">↗</span></a>
     </div>
     <div class="hero-meta">
       <span class="pill">macOS .pkg</span>
       <span class="pill">Linux binary</span>
       <span class="pill">Node 24.22.2</span>
     </div>
   </section>

.. _install:

Install
-------

.. raw:: html

   <div class="platform-grid">
     <a class="platform-card" href="https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-arm64.pkg">
       <span class="platform-kicker">macOS · Apple silicon</span>
       <strong>gode-darwin-arm64.pkg</strong>
       <span class="platform-note">macOS 12 or later · installer package</span>
       <span class="card-link">Download →</span>
     </a>
     <a class="platform-card" href="https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-amd64.pkg">
       <span class="platform-kicker">macOS · Intel</span>
       <strong>gode-darwin-amd64.pkg</strong>
       <span class="platform-note">macOS 12 or later · installer package</span>
       <span class="card-link">Download →</span>
     </a>
     <a class="platform-card" href="https://kr12345ad.github.io/godejs/install.sh">
       <span class="platform-kicker">Linux · x86-64</span>
       <strong>One-line installer</strong>
       <span class="platform-note">Requires JavaScriptCore GTK 4.0</span>
       <span class="card-link">Install →</span>
     </a>
   </div>

On macOS, open the package or install it from a terminal:

.. code-block:: sh

   sudo installer -pkg gode-darwin-arm64.pkg -target /
   gode --version

The package installs ``gode`` in ``/usr/local/bin`` and the project licenses
under ``/usr/local/share/doc/godejs``. Release packages are not notarized by
default; if Gatekeeper blocks one, right-click it and choose **Open**, or allow
it in **System Settings > Privacy & Security**.

On Linux, install JavaScriptCore first, then run the installer:

.. code-block:: sh

   sudo apt install libjavascriptcoregtk-4.0-18   # Debian 12, Ubuntu 22.04
   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh

Each release also publishes a ``.sha256`` file. All builds are on the
`GitHub releases page <https://github.com/kr12345ad/godejs/releases>`_.

Re-running the installer keeps an existing ``gode`` installation: it upgrades
when the requested release is newer, skips the install when the version matches,
and warns instead of downgrading. Pass ``--force`` to reinstall or downgrade:

.. code-block:: sh

   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh -s -- --force

First program
-------------

.. code-block:: sh

   gode -e "console.log(process.version)"
   gode app.js

Both print ``v24.22.2`` for the tracked runtime version. The
:doc:`command-line guide <usage/cli>` covers eval, stdin, watch mode, and every
supported flag.

What works
----------

.. raw:: html

   <div class="feature-grid">
     <article class="feature-card">
       <span class="card-kicker">Node 24 surface</span>
       <h3>Standard library</h3>
       <p>Files, streams, HTTP, TLS, crypto, workers, sqlite, and the rest of the core modules.</p>
     </article>
     <article class="feature-card">
       <span class="card-kicker">Loading</span>
       <h3>CommonJS + ESM</h3>
       <p>Both module systems work, including package exports, imports maps, and dynamic import.</p>
     </article>
     <article class="feature-card">
       <span class="card-kicker">Network</span>
       <h3>HTTP/1.1, HTTP/2, HTTP/3</h3>
       <p>Compatible request and response objects, with a native QUIC server in <code>gode:http3</code>.</p>
     </article>
     <article class="feature-card">
       <span class="card-kicker">Tooling</span>
       <h3>CLI + REPL + debugger</h3>
       <p>Node-style flags, an interactive REPL, and a line debugger built around real lexical scope.</p>
     </article>
   </div>

Documentation
-------------

.. raw:: html

   <div class="section-grid">
     <a class="doc-card" href="usage/cli.html">
       <span class="card-kicker">Start here</span>
       <h3>Command line</h3>
       <p>Run files, eval snippets, read stdin, and use watch mode.</p>
       <span class="card-link">Read the guide →</span>
     </a>
     <a class="doc-card" href="usage/modules.html">
       <span class="card-kicker">Core concepts</span>
       <h3>Modules and packages</h3>
       <p>How CommonJS, ESM, package exports, and extension resolution behave.</p>
       <span class="card-link">Understand loading →</span>
     </a>
     <a class="doc-card" href="usage/api/index.html">
       <span class="card-kicker">Reference</span>
       <h3>API index</h3>
       <p>Every module and global, with signatures, arguments, returns, and examples.</p>
       <span class="card-link">Browse 45 surfaces →</span>
     </a>
     <a class="doc-card" href="http3.html">
       <span class="card-kicker">Extension</span>
       <h3>HTTP/3</h3>
       <p>Build a QUIC server with the same request and response objects as <code>node:http</code>.</p>
       <span class="card-link">Open the guide →</span>
     </a>
     <a class="doc-card" href="platforms.html">
       <span class="card-kicker">Deployment</span>
       <h3>Platforms</h3>
       <p>macOS packages, Linux runtime dependencies, and source builds.</p>
       <span class="card-link">Check your platform →</span>
     </a>
     <a class="doc-card" href="javascriptcore.html">
       <span class="card-kicker">Internals</span>
       <h3>JavaScriptCore</h3>
       <p>What changes when the engine is JavaScriptCore instead of V8.</p>
       <span class="card-link">See the differences →</span>
     </a>
   </div>

Build from source
-----------------

GodeJS needs Go 1.26 or later and a C compiler. On macOS, install the Xcode
Command Line Tools and build:

.. code-block:: sh

   xcode-select --install
   git clone https://github.com/kr12345ad/godejs
   cd godejs
   CGO_ENABLED=1 go build -o gode ./cmd/gode

On Linux, install the JavaScriptCore development headers first. See
:doc:`platforms` for package names and link flags.

.. toctree::
   :hidden:
   :maxdepth: 2

   usage/index
   platforms
   javascriptcore
   http3
   module-guide
   usage/api/index
