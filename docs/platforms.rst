Platforms
=========

Linux
-----

Release binaries are built on Ubuntu 22.04 for x86-64 and link
``libjavascriptcoregtk-4.0.so.18``. Debian 12 and Ubuntu 22.04 package the
runtime and development headers. Ubuntu 24.04 removed the 4.0 packages; use a
22.04 environment or build JavaScriptCore 4.0 from source.

Install the runtime before running a downloaded binary:

.. code-block:: sh

   sudo apt install libjavascriptcoregtk-4.0-18   # Debian 12, Ubuntu 22.04
   sudo apk add webkit2gtk-4.0                    # Alpine
   sudo dnf install webkit2gtk4.0                 # Fedora

macOS
-----

The macOS binary links the system ``JavaScriptCore.framework``, so no separate
runtime package is required. Releases are built for both architectures:

* ``gode-darwin-arm64`` for Apple silicon
* ``gode-darwin-amd64`` for Intel Macs

macOS 12 or later is required. The release binary is not code-signed; if
Gatekeeper blocks it, allow it under System Settings > Privacy & Security, or
remove the quarantine attribute:

.. code-block:: sh

   xattr -d com.apple.quarantine ./gode

Xcode Command Line Tools provide the compiler and framework for a source build:

.. code-block:: sh

   xcode-select --install
   git clone https://github.com/kr12345ad/godejs
   cd godejs
   CGO_ENABLED=1 go build -o gode ./cmd/gode

The Node-API ``.node`` loader is currently Linux-only. Core modules, CommonJS,
ESM, the REPL, HTTP/1.1, HTTP/2, and HTTP/3 work on macOS.

Source builds
-------------

GodeJS requires Go 1.26 or later and cgo enabled. The JavaScriptCore link flags
are selected by platform in ``internal/js``:

* Linux: ``libjavascriptcoregtk-4.0.so.18``
* macOS: ``-framework JavaScriptCore``

On Linux the development headers are also required:

.. code-block:: sh

   sudo apt install libjavascriptcoregtk-4.0-dev golang gcc
