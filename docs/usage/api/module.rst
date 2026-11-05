module
======

The ``module`` builtin.

Loading
-------

.. code-block:: javascript

   const module = require('node:module');
   // ESM
   import * as module from 'node:module';

API
---

_findPath
~~~~~~~~~

.. code-block:: text

   module._findPath(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._findPath); // 'function'
   console.log(module._findPath.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_initPaths
~~~~~~~~~~

.. code-block:: text

   module._initPaths(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._initPaths); // 'function'
   console.log(module._initPaths.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_nodeModulePaths
~~~~~~~~~~~~~~~~

.. code-block:: text

   module._nodeModulePaths(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._nodeModulePaths); // 'function'
   console.log(module._nodeModulePaths.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_resolveFilename
~~~~~~~~~~~~~~~~

.. code-block:: text

   module._resolveFilename(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._resolveFilename); // 'function'
   console.log(module._resolveFilename.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_resolveLookupPaths
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module._resolveLookupPaths(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._resolveLookupPaths); // 'function'
   console.log(module._resolveLookupPaths.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_stat
~~~~~

.. code-block:: text

   module._stat(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   console.log(typeof module._stat); // 'function'
   console.log(module._stat.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


createRequire
~~~~~~~~~~~~~

.. code-block:: text

   module.createRequire(filename)

*Kind:* function · *Arguments:* 1

**Arguments**

* ``filename`` (*string|URL*): Filename to be used to construct the require

**Example**

.. code-block:: javascript

   import { createRequire } from 'node:module';
   const require = createRequire(import.meta.url);

   // sibling-module.js is a CommonJS module.
   const siblingModule = require('./sibling-module');


createRequireFromPath
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Module.createRequireFromPath()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

Use ``module.createRequire()`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.createRequireFromPath();
   console.log(result);


enableCompileCache
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.enableCompileCache([options])

*Kind:* function · *Arguments:* 1

Enable module compile cache in the current Node.js instance.

For general use cases, it's recommended to call ``module.enableCompileCache()`` without specifying the ``options.directory``, so that the directory can be overridden by the ``NODE_COMPILE_CACHE`` environment variable when necessary.

**Arguments**

* ``options`` (*string|Object*) — optional: Optional. If a string is passed, it is considered to be ``options.directory``.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.enableCompileCache({});
   console.log(result);


findPackageJSON
~~~~~~~~~~~~~~~

.. code-block:: text

   module.findPackageJSON(specifier[, base])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``specifier`` (*string|URL*): The specifier for the module whose ``package.json`` to
* ``base`` (*string|URL*) — optional: The absolute location (``file:`` URL string or FS path) of the

**Example**

.. code-block:: javascript

   // /path/to/project/packages/bar/bar.js
   const { findPackageJSON } = require('node:module');
   const { pathToFileURL } = require('node:url');
   const path = require('node:path');

   findPackageJSON('..', __filename);
   // '/path/to/project/package.json'
   // Same result when passing an absolute specifier instead:
   findPackageJSON(pathToFileURL(path.join(__dirname, '..')));

   findPackageJSON('some-package', __filename);
   // '/path/to/project/packages/bar/node_modules/some-package/package.json'
   // When passing an absolute specifier, you might get a different result if the
   // resolved module is inside a subfolder that has nested `package.json`.
   findPackageJSON(pathToFileURL(require.resolve('some-package')));
   // '/path/to/project/packages/bar/node_modules/some-package/some-subfolder/package.json'

   findPackageJSON('@foo/qux', __filename);
   // '/path/to/project/packages/qux/package.json'


findSourceMap
~~~~~~~~~~~~~

.. code-block:: text

   module.findSourceMap(path)

*Kind:* function · *Arguments:* 1

``path`` is the resolved path for the file for which a corresponding source map should be fetched.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.findSourceMap('./file.txt');
   console.log(result);


flushCompileCache
~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.flushCompileCache()

*Kind:* function · *Arguments:* 0

Flush the module compile cache accumulated from modules already loaded in the current Node.js instance to disk. This returns after all the flushing file system operations come to an end, no matter they succeed or not. If there are any errors, this will fail silently, since compile cache misses should not interfere with the actual operation of the application.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.flushCompileCache();
   console.log(result);


getCompileCacheDir
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.getCompileCacheDir()

*Kind:* function · *Arguments:* 0

<i id="module\_customization\_hooks"></i>

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.getCompileCacheDir();
   console.log(result);


getSourceMapsSupport
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.getSourceMapsSupport()

*Kind:* function · *Arguments:* 0

This method returns whether the Source Map v3 support for stack traces is enabled.

<a id="module\_module\_findsourcemap\_path\_error"></a>

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.getSourceMapsSupport();
   console.log(result);


isBuiltin
~~~~~~~~~

.. code-block:: text

   module.isBuiltin(moduleName)

*Kind:* function · *Arguments:* 1

**Arguments**

* ``moduleName`` (*string*): name of the module

**Example**

.. code-block:: javascript

   import { isBuiltin } from 'node:module';
   isBuiltin('node:fs'); // true
   isBuiltin('fs'); // true
   isBuiltin('wss'); // false


Module
~~~~~~

.. code-block:: text

   Module

*Kind:* class · *Arguments:* 0

Provides general utility methods when interacting with instances of ``Module``, the ``module`` variable often seen in CommonJS modules. Accessed via ``import 'node:module'`` or ``require('node:module')``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   // module.cjs
   // In a CommonJS module
   const builtin = require('node:module').builtinModules;


register
~~~~~~~~

.. code-block:: text

   module.register()

*Kind:* function · *Arguments:* 3

Type: Documentation-only

``module.register()`` is deprecated. Use ``module.registerHooks()`` instead.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.register(undefined, undefined, undefined);
   console.log(result);


registerHooks
~~~~~~~~~~~~~

.. code-block:: text

   module.registerHooks(options)

*Kind:* function · *Arguments:* 1

Register hooks that customize Node.js module resolution and loading behavior. See Customization hooks. The returned object can be used to deregister the hooks.

**Arguments**

* ``options`` (*Object*)

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.registerHooks({});
   console.log(result);


runMain
~~~~~~~

.. code-block:: text

   module.runMain(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.runMain();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


setSourceMapsSupport
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.setSourceMapsSupport(enabled[, options])

*Kind:* function · *Arguments:* 2

This function enables or disables the Source Map v3 support for stack traces.

It provides same features as launching Node.js process with commandline options ``--enable-source-maps``, with additional options to alter the support for files in ``node_modules`` or generated codes.

**Arguments**

* ``enabled`` (*boolean*): Enable the source map support.
* ``options`` (*Object*) — optional: Optional

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.setSourceMapsSupport(true, {});
   console.log(result);


SourceMap
~~~~~~~~~

.. code-block:: text

   module.SourceMap

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const instance = new module.SourceMap(undefined);
   console.log(instance);


stripTypeScriptTypes
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.stripTypeScriptTypes(code[, options])

*Kind:* function · *Arguments:* 2

``module.stripTypeScriptTypes()`` removes type annotations from TypeScript code. It can be used to strip type annotations from TypeScript code before running it with ``vm.runInContext()`` or ``vm.compileFunction()``.

By default, it will throw an error if the code contains TypeScript features that require transformation such as ``Enums``, see type-stripping for more information.

**Arguments**

* ``code`` (*string*): The code to strip type annotations from.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { stripTypeScriptTypes } = require('node:module');
   const code = 'const a: number = 1;';
   const strippedCode = stripTypeScriptTypes(code);
   console.log(strippedCode);
   // Prints: const a         = 1;


syncBuiltinESMExports
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   module.syncBuiltinESMExports()

*Kind:* function · *Arguments:* 0

The ``module.syncBuiltinESMExports()`` method updates all the live bindings for builtin ES Modules to match the properties of the CommonJS exports. It does not add or remove exported names from the ES Modules.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const fs = require('node:fs');
   const assert = require('node:assert');
   const { syncBuiltinESMExports } = require('node:module');

   fs.readFile = newAPI;

   delete fs.readFileSync;

   function newAPI() {
     // ...
   }

   fs.newAPI = newAPI;

   syncBuiltinESMExports();

   import('node:fs').then((esmFS) => {
     // It syncs the existing readFile property with the new value
     assert.strictEqual(esmFS.readFile, newAPI);
     // readFileSync has been deleted from the required fs
     assert.strictEqual('readFileSync' in fs, false);
     // syncBuiltinESMExports() does not remove readFileSync from esmFS
     assert.strictEqual('readFileSync' in esmFS, true);
     // syncBuiltinESMExports() does not add names
     assert.strictEqual(esmFS.newAPI, undefined);
   });


wrap
~~~~

.. code-block:: text

   module.wrap(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.wrap();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


wrapSafe
~~~~~~~~

.. code-block:: text

   module.wrapSafe(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module.wrapSafe();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_extensions
-----------

_extensions..js
~~~~~~~~~~~~~~~

.. code-block:: text

   module._extensions..js(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module._extensions..js();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_extensions..json
~~~~~~~~~~~~~~~~~

.. code-block:: text

   module._extensions..json(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module._extensions..json();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_extensions..node
~~~~~~~~~~~~~~~~~

.. code-block:: text

   module._extensions..node(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const result = module._extensions..node();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Module class
------------

**Module** — Provides general utility methods when interacting with instances of ``Module``, the ``module`` variable often seen in CommonJS modules. Accessed via ``import 'node:module'`` or ``require('node:module')``.

Module.load
~~~~~~~~~~~

.. code-block:: text

   load(url, context, nextLoad)

*Kind:* function · *Arguments:* 3

The ``load`` hook provides a way to define a custom method for retrieving the source code of a resolved URL. This would allow a loader to potentially avoid reading files from disk. It could also be used to map an unrecognized format to a supported one, for example ``yaml`` to ``module``.

In a more advanced scenario, this can also be used to transform an unsupported source to a supported one (see \`Examples <#examples>\`_ below).

**Arguments**

* ``url`` (*string*): The URL returned by the ``resolve`` chain
* ``context`` (*Object*)
* ``nextLoad`` (*Function*): The subsequent ``load`` hook in the chain, or the

**Example**

.. code-block:: javascript

   import { registerHooks } from 'node:module';
   import { Buffer } from 'node:buffer';

   function load(url, context, nextLoad) {
     // The hook can skip default loading and provide a custom source code.
     if (url === 'special-module') {
       return {
         source: 'export const special = 42;',
         format: 'module',
         shortCircuit: true,  // This is mandatory if nextLoad() is not called.
       };
     }

     // It's possible to modify the source code loaded by the next - possibly default - step,
     // for example, replacing 'foo' with 'bar' in the source code of the module.
     const result = nextLoad(url, context);
     const source = typeof result.source === 'string' ?
       result.source : Buffer.from(result.source).toString('utf8');
     return {
       source: source.replace(/foo/g, 'bar'),
       ...result,
     };
   }

   registerHooks({ load });


Module.require
~~~~~~~~~~~~~~

.. code-block:: text

   module.require(id)

*Kind:* function · *Arguments:* 1

The ``module.require()`` method provides a way to load a module as if ``require()`` was called from the original module.

In order to do this, it is necessary to get a reference to the ``module`` object. Since ``require()`` returns the ``module.exports``, and the ``module`` is typically *only* available within a specific module's code, it must be explicitly exported in order to be used.

**Arguments**

* ``id`` (*string*)

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const instance = new module.Module();
   const result = instance.require('example');
   console.log(result);


Module.resolve
~~~~~~~~~~~~~~

.. code-block:: text

   resolve(specifier, context, nextResolve)

*Kind:* function · *Arguments:* 3

The ``resolve`` hook chain is responsible for telling Node.js where to find and how to cache a given ``import`` statement or expression, or ``require`` call. It can optionally return a format (such as ``'module'``) as a hint to the ``load`` hook. If a format is specified, the ``load`` hook is ultimately responsible for providing the final ``format`` value (and it is free to ignore the hint provided by ``resolve``); if ``resolve`` provides a ``format``, a custom ``load`` hook is required even if only to pass the value to the Node.js default ``load`` hook.

Import type attributes are part of the cache key for saving loaded modules into the internal module cache. The ``resolve`` hook is responsible for returning an ``importAttributes`` object if the module should be cached with different attributes than were present in the source code.

**Arguments**

* ``specifier`` (*string*)
* ``context`` (*Object*)
* ``nextResolve`` (*Function*): The subsequent ``resolve`` hook in the chain, or the

**Example**

.. code-block:: javascript

   import { registerHooks } from 'node:module';

   function resolve(specifier, context, nextResolve) {
     // When calling `defaultResolve`, the arguments can be modified. For example,
     // to change the specifier or to add applicable export conditions.
     if (specifier.includes('foo')) {
       specifier = specifier.replace('foo', 'bar');
       return nextResolve(specifier, {
         ...context,
         conditions: [...context.conditions, 'another-condition'],
       });
     }

     // The hook can also skip default resolution and provide a custom URL.
     if (specifier === 'special-module') {
       return {
         url: 'file:///path/to/special-module.mjs',
         format: 'module',
         shortCircuit: true,  // This is mandatory if nextResolve() is not called.
       };
     }

     // If no customization is needed, defer to the next hook in the chain which would be the
     // Node.js default resolve if this is the last user-specified loader.
     return nextResolve(specifier);
   }

   registerHooks({ resolve });


SourceMap class
---------------

SourceMap.findEntry
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceMap.findEntry(lineOffset, columnOffset)

*Kind:* function · *Arguments:* 2

Given a line offset and column offset in the generated source file, returns an object representing the SourceMap range in the original file if found, or an empty object if not.

The object returned contains the following keys:

**Arguments**

* ``lineOffset`` (*number*): The zero-indexed line number offset in
* ``columnOffset`` (*number*): The zero-indexed column number offset

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const instance = new module.SourceMap();
   const result = instance.findEntry(1, 1);
   console.log(result);


SourceMap.findOrigin
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceMap.findOrigin(lineNumber, columnNumber)

*Kind:* function · *Arguments:* 2

Given a 1-indexed ``lineNumber`` and ``columnNumber`` from a call site in the generated source, find the corresponding call site location in the original source.

If the ``lineNumber`` and ``columnNumber`` provided are not found in any source map, then an empty object is returned. Otherwise, the returned object contains the following keys:

**Arguments**

* ``lineNumber`` (*number*): The 1-indexed lineNumber of the
* ``columnNumber`` (*number*): The 1-indexed columnNumber of the

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const instance = new module.SourceMap();
   const result = instance.findOrigin(1, 1);
   console.log(result);


SourceMap.return
~~~~~~~~~~~~~~~~

.. code-block:: text

   module.SourceMap.return(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const module = require('node:module');

   const instance = new module.SourceMap();
   const result = instance.return();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``_cache`` — object
* ``_extensions`` — object
* ``builtinModules`` — array
* ``compileCacheStatus`` — object
* ``compileCacheStatus.ALREADY_ENABLED`` — number
* ``compileCacheStatus.DISABLED`` — number
* ``compileCacheStatus.ENABLED`` — number
* ``compileCacheStatus.FAILED`` — number
* ``constants`` — object
* ``constants.compileCacheStatus`` — object
* ``globalPaths`` — array
* ``length`` — number
* ``name`` — string
* ``wrapper`` — array

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
