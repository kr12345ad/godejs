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

The macOS build links the system ``JavaScriptCore.framework``, so no separate
runtime package is required. Releases are built for both architectures:

* ``gode-darwin-arm64.pkg`` for Apple silicon
* ``gode-darwin-amd64.pkg`` for Intel Macs

Double-click the package, or install it from a terminal:

.. code-block:: sh

   sudo installer -pkg gode-darwin-arm64.pkg -target /
   gode --version

The package installs ``/usr/local/bin/gode`` and the project licenses under
``/usr/local/share/doc/godejs``. It is not notarized by default; if Gatekeeper
blocks it, right-click the package and choose **Open**, or allow it in
**System Settings > Privacy & Security**. A Homebrew-style prefix is not used.

Standalone binaries remain available as ``gode-darwin-arm64`` and
``gode-darwin-amd64`` for automation. If Gatekeeper blocks a raw binary,
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
