GodeJS
======

Node.js 24 · JavaScriptCore · Go

GodeJS is a Node.js 24-compatible runtime built in Go on JavaScriptCore. Bring
your CommonJS and ESM apps, the standard library, and the command line you
already know.

:ref:`Install GodeJS <install>` · :doc:`Browse the API <usage/api/index>` ·
`View on GitHub <https://github.com/kr12345ad/godejs>`_

macOS ``.pkg`` · Linux binary · Node 24.22.2

.. _install:

Download
--------

.. list-table::
   :header-rows: 1
   :widths: 26 40 34

   * - Platform
     - Download
     - Requirements
   * - Linux x86-64
     - `One-line installer <https://kr12345ad.github.io/godejs/install.sh>`_
     - JavaScriptCore GTK 4.0
   * - macOS Apple silicon
     - `gode-darwin-arm64.pkg <https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-arm64.pkg>`_
     - macOS 12 or later
   * - macOS Intel
     - `gode-darwin-amd64.pkg <https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-amd64.pkg>`_
     - macOS 12 or later

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

* **Standard library** — Files, streams, HTTP, TLS, crypto, workers, sqlite, and
  the rest of the core modules.
* **CommonJS + ESM** — Both module systems work, including package exports,
  imports maps, and dynamic import.
* **HTTP/1.1, HTTP/2, HTTP/3** — Compatible request and response objects, with a
  native QUIC server in ``gode:http3``.
* **CLI + REPL + debugger** — Node-style flags, an interactive REPL, and a line
  debugger built around real lexical scope.

Documentation
-------------

* :doc:`Command line <usage/cli>` — Run files, eval snippets, read stdin, and
  use watch mode.
* :doc:`Modules and packages <usage/modules>` — How CommonJS, ESM, package
  exports, and extension resolution behave.
* :doc:`API index <usage/api/index>` — Every module and global, with
  signatures, arguments, returns, and examples.
* :doc:`HTTP/3 <http3>` — Build a QUIC server with the same request and
  response objects as ``node:http``.
* :doc:`Platforms <platforms>` — macOS packages, Linux runtime dependencies,
  and source builds.
* :doc:`JavaScriptCore <javascriptcore>` — What changes when the engine is
  JavaScriptCore instead of V8.

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
