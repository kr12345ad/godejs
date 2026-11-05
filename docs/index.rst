Getting Started
===============

GodeJS is a Node.js 24-compatible runtime built in Go on JavaScriptCore.

Download
--------

.. list-table::
   :header-rows: 1
   :widths: 26 40 34

   * - Platform
     - Download
     - Requirements
   * - Linux x86-64
     - `gode-linux-amd64 <https://github.com/kr12345ad/godejs/releases/latest/download/gode-linux-amd64>`_
     - JavaScriptCore GTK 4.0
   * - macOS Apple silicon
     - `gode-darwin-arm64 <https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-arm64>`_
     - macOS 12 or later
   * - macOS Intel
     - `gode-darwin-amd64 <https://github.com/kr12345ad/godejs/releases/latest/download/gode-darwin-amd64>`_
     - macOS 12 or later

Each release also publishes a ``.sha256`` file. All releases are on the
`GitHub releases page <https://github.com/kr12345ad/godejs/releases>`_.

Install on Linux or macOS
-------------------------

.. code-block:: sh

   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh

The installer downloads the release for the current machine into
``~/.local/bin``. To choose another prefix:

.. code-block:: sh

   curl -fsSL https://kr12345ad.github.io/godejs/install.sh | \
     GODE_PREFIX=/usr/local sh

The macOS build links the system ``JavaScriptCore.framework``, so no runtime
package is needed. On Linux, install the JavaScriptCore GTK 4.0 runtime first:

.. code-block:: sh

   sudo apt install libjavascriptcoregtk-4.0-18   # Debian 12, Ubuntu 22.04
   sudo apk add webkit2gtk-4.0                    # Alpine
   sudo dnf install webkit2gtk4.0                 # Fedora

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
:doc:`platforms` for package names.

First program
-------------

.. code-block:: sh

   gode -e "console.log(process.version)"
   gode app.js

Both print ``v24.22.2`` for the tracked runtime version.

Documentation
-------------

.. toctree::
   :maxdepth: 2

   usage/index
   platforms
   javascriptcore
   http3
   module-guide
   usage/api/index
