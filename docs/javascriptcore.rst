JavaScriptCore
==============

GodeJS does not use V8, goja, or a JavaScript shim. It embeds JavaScriptCore
directly, and every host object exposed to JavaScript is built from the C API
primitives in ``internal/js``.

Binding
-------

The binding is deliberately small and auditable. JavaScriptCore
``JSValueRef`` values are NaN-boxed, so ``undefined`` can be the machine word
``0xa``. Storing a value in a Go pointer slot makes the garbage collector abort
with ``invalid pointer found on stack``. Handles are therefore kept as tagged
``uintptr`` values, and the C prototypes in ``godejsc.h`` declare pointers as
``uintptr_t``.

Relevant files:

.. list-table::
   :header-rows: 1

   * - File
     - Purpose
   * - ``internal/js/jsc.go``
     - Contexts, evaluation, promises, errors, GC roots
   * - ``internal/js/value.go``
     - Objects, arrays, typed arrays, property access
   * - ``internal/js/callback.go``
     - C callback bridge into Go
   * - ``internal/js/godejsc.h``
     - Minimal C declarations used by the binding
   * - ``internal/env/stack.go``
     - V8-style ``CallSite`` layer over JSC stack strings

Threading
---------

JavaScriptCore keeps thread-affine VM state. ``cmd/gode`` pins the goroutine
that creates the context to one OS thread for the life of the process. Socket,
file, and child-process I/O run on other goroutines and deliver results back to
the engine thread.

The process uses one context per runtime. Worker threads create independent
contexts; JavaScriptCore cannot share ordinary values between them.

Linking
-------

Linux links the GTK 4.0 JavaScriptCore library:

.. code-block:: c

   #cgo LDFLAGS: -l:libjavascriptcoregtk-4.0.so.18

macOS links the system framework:

.. code-block:: c

   #cgo LDFLAGS: -framework JavaScriptCore

The link flags live in platform files so one package can be built on both
systems. The Node-API loader is currently limited to Linux.

Dependencies
------------

Linux runtime packages:

.. code-block:: sh

   # Debian 12 / Ubuntu 22.04
   sudo apt install libjavascriptcoregtk-4.0-18

   # Alpine
   sudo apk add webkit2gtk-4.0

   # Fedora
   sudo dnf install webkit2gtk4.0

Linux source builds also need the development headers:

.. code-block:: sh

   # Debian 12 / Ubuntu 22.04
   sudo apt install libjavascriptcoregtk-4.0-dev golang gcc

   # Alpine
   sudo apk add webkit2gtk-4.0-dev go gcc musl-dev

`Ubuntu 24.04 removed the JavaScriptCore GTK 4.0 packages
<https://packages.ubuntu.com/>`_. GodeJS links the 4.0 soname, so installing
4.1 does not provide the required library. Use Ubuntu 22.04, Debian 12, or
build JSC 4.0 for 24.04.

Stack traces
------------

JavaScriptCore exposes a formatted ``name@file:line:column`` stack string
instead of V8's structured frames. GodeJS preserves that default output and
activates structured ``CallSite`` objects only when
``Error.prepareStackTrace`` is set. This keeps normal errors readable while
supporting the Node compatibility layer.
