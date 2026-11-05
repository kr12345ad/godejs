wasi
====

<strong class="critical">The ``node:wasi`` module does not currently provide the comprehensive file system security properties provided by some WASI runtimes. Full support for secure file system sandboxing may or may not be implemented in future. In the mean time, do not rely on it to run untrusted code. </strong>

The WASI API provides an implementation of the WebAssembly System Interface specification. WASI gives WebAssembly applications access to the underlying operating system via a collection of POSIX-like functions.

Loading
-------

.. code-block:: javascript

   const wasi = require('node:wasi');
   // ESM
   import * as wasi from 'node:wasi';

Quick start
-----------

.. code-block:: javascript

   const { readFile } = require('node:fs/promises');
   const { WASI } = require('node:wasi');
   const { argv, env } = require('node:process');
   const { join } = require('node:path');

   const wasi = new WASI({
     version: 'preview1',
     args: argv,
     env,
     preopens: {
       '/local': '/some/real/path/that/wasm/can/access',
     },
   });

   (async () => {
     const wasm = await WebAssembly.compile(
       await readFile(join(__dirname, 'demo.wasm')),
     );
     const instance = await WebAssembly.instantiate(wasm, wasi.getImportObject());

     wasi.start(instance);
   })();

API
---

WASI
~~~~

.. code-block:: text

   WASI

*Kind:* class · *Arguments:* 1

The ``WASI`` class provides the WASI system call API and additional convenience methods for working with WASI-based applications. Each ``WASI`` instance represents a distinct environment.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/wasi.html#class-wasi>`__

WASI class
----------

**WASI** — The ``WASI`` class provides the WASI system call API and additional convenience methods for working with WASI-based applications. Each ``WASI`` instance represents a distinct environment.

WASI.finalizeBindings
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   wasi.finalizeBindings(instance[, options])

*Kind:* function · *Arguments:* 2

Set up WASI host bindings to ``instance`` without calling ``initialize()`` or ``start()``. This method is useful when the WASI module is instantiated in child threads for sharing the memory across threads.

``finalizeBindings()`` requires that either ``instance`` exports a ``WebAssembly.Memory`` named ``memory`` or user specify a ``WebAssembly.Memory`` object in ``options.memory``. If the ``memory`` is invalid an exception is thrown.

**Arguments**

* ``instance`` (*WebAssembly.Instance*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/wasi.html#wasifinalizebindingsinstance-options>`__

WASI.getImportObject
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   wasi.getImportObject()

*Kind:* function · *Arguments:* 0

Return an import object that can be passed to ``WebAssembly.instantiate()`` if no other WASM imports are needed beyond those provided by WASI.

If version ``unstable`` was passed into the constructor it will return:

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/wasi.html#wasigetimportobject>`__

WASI.initialize
~~~~~~~~~~~~~~~

.. code-block:: text

   wasi.initialize(instance)

*Kind:* function · *Arguments:* 1

Attempt to initialize ``instance`` as a WASI reactor by invoking its ``_initialize()`` export, if it is present. If ``instance`` contains a ``_start()`` export, then an exception is thrown.

``initialize()`` requires that ``instance`` exports a ``WebAssembly.Memory`` named ``memory``. If ``instance`` does not have a ``memory`` export an exception is thrown.

**Arguments**

* ``instance`` (*WebAssembly.Instance*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/wasi.html#wasiinitializeinstance>`__

WASI.start
~~~~~~~~~~

.. code-block:: text

   wasi.start(instance)

*Kind:* function · *Arguments:* 1

Attempt to begin execution of ``instance`` as a WASI command by invoking its ``_start()`` export. If ``instance`` does not contain a ``_start()`` export, or if ``instance`` contains an ``_initialize()`` export, then an exception is thrown.

``start()`` requires that ``instance`` exports a ``WebAssembly.Memory`` named ``memory``. If ``instance`` does not have a ``memory`` export an exception is thrown.

**Arguments**

* ``instance`` (*WebAssembly.Instance*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/wasi.html#wasistartinstance>`__

Properties and constants
------------------------

* ``errno`` — object
* ``errno.badf`` — number
* ``errno.inval`` — number
* ``errno.io`` — number
* ``errno.noent`` — number
* ``errno.nosys`` — number
* ``errno.notdir`` — number
* ``errno.notempty`` — number
* ``errno.success`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
