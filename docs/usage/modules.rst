Modules
=======

GodeJS supports CommonJS ``require()`` and the commonly used ES module
syntax. Builtins work with both the bare name and the ``node:`` prefix:

.. code-block:: javascript

   const fs = require('fs');
   const path = require('node:path');

   import * as os from 'node:os';
   import { readFile } from 'node:fs/promises';

File extensions
---------------

.. list-table::
   :header-rows: 1
   :widths: 22 78

   * - Extension
     - Behavior
   * - ``.js``
     - CommonJS unless the nearest ``package.json`` has ``"type": "module"``.
   * - ``.cjs``
     - Always CommonJS.
   * - ``.mjs``
     - ES module syntax, transformed into the CommonJS wrapper.
   * - ``.json``
     - Parsed and returned as ``module.exports``.
   * - ``.node``
     - Loaded as a native Node-API addon when the runtime supports it.

There is no source-level TypeScript loader. ``--experimental-strip-types`` is
refused rather than silently running the wrong file.

Package resolution
------------------

Relative and absolute paths follow Node's algorithm:

1. The exact file.
2. The file with ``.js``, ``.json``, ``.cjs``, or ``.node`` appended.
3. A directory's ``package.json`` ``exports`` map.
4. ``package.json`` ``main``.
5. ``index.js``, ``index.json``, ``index.cjs``, or ``index.node``.

Bare specifiers search ``node_modules`` from the importing directory upward.
Package ``exports`` and ``imports`` maps are honored, including wildcard
subpaths, arrays, and the ``require``, ``node``, ``default``, and ``import``
conditions. ``#internal`` imports declared in ``package.json`` also resolve.

Custom export conditions (``--conditions``) and custom loaders are not
implemented and fail explicitly.

ES modules
----------

Static ``import``/``export`` syntax in ``.mjs`` files (and ``.js`` files in a
module package) is rewritten into GodeJS's CommonJS wrapper. That means:

* Default, namespace, and named imports work.
* ``export``, ``export default``, and re-exports work.
* ``import.meta.url``, ``import.meta.filename``, ``import.meta.dirname``, and
  ``import.meta.resolve()`` are available.
* Dynamic ``import()`` returns a promise and loads through ``require``.
* Top-level ``await import(...)`` is folded to a synchronous require. General
  top-level ``await`` is not supported.

Because JavaScriptCore has no native ESM loader, there is no separate module
graph or live-binding semantics. If a package depends on true ESM live
bindings, test it before relying on it.

``require()`` cannot load an ``.mjs`` file directly. Use a static import,
``import()``, or point the entry script at the file.

Preloading
----------

``-r``/``--require`` loads a CommonJS module before the entry script.
``--import`` loads another module in the same preload pass:

.. code-block:: sh

   gode -r ./setup.js --import ./telemetry.mjs app.js

Preloads run after the runtime globals are installed and before the entry
script. A preload error stops startup with status 1.

Symlinks
--------

By default, resolved files are real paths. Pass ``--preserve-symlinks`` to keep
the symlinked path for all modules, or ``--preserve-symlinks-main`` to keep it
for the entry script only.

Native addons
-------------

``.node`` files are loaded through the host's Node-API loader. ``--no-addons``
is refused because addon loading cannot currently be disabled safely.

Writing a builtin
-----------------

If you are adding a new core module in Go rather than using the runtime, see
:doc:`../module-guide` for the package contract and test helpers.
