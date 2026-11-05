vm
===

The ``node:vm`` module enables compiling and running code within V8 Virtual Machine contexts.

<strong class="critical">The ``node:vm`` module is not a security mechanism. Do not use it to run untrusted code.</strong>

Loading
-------

.. code-block:: javascript

   const vm = require('node:vm');
   // ESM
   import * as vm from 'node:vm';

Quick start
-----------

.. code-block:: javascript

   const { createContext, runInContext } = require('node:vm');

   const x = 1;

   const context = { x: 2 };
   createContext(context); // Contextify the object.

   const code = 'x += 40; var y = 17;';
   // `x` and `y` are global variables in the context.
   // Initially, x has the value 2 because that is the value of context.x.
   runInContext(code, context);

   console.log(context.x); // 42
   console.log(context.y); // 17

   console.log(x); // 1; y is not defined

API
---

compileFunction
~~~~~~~~~~~~~~~

.. code-block:: text

   vm.compileFunction(code[, params[, options]])

*Kind:* function · *Arguments:* 3

Compiles the given code into the provided context (if no context is supplied, the current context is used), and returns it wrapped inside a function with the given ``params``.

**Arguments**

* ``code`` (*string*): The body of the function to compile.
* ``params`` (*string\[]*) — optional: An array of strings containing all parameters for the
* ``options`` (*Object*) — optional


createContext
~~~~~~~~~~~~~

.. code-block:: text

   vm.createContext([contextObject[, options]])

*Kind:* function · *Arguments:* 2

If the given ``contextObject`` is an object, the ``vm.createContext()`` method will prepare that object and return a reference to it so that it can be used in calls to ``vm.runInContext()`` or ``script.runInContext()``. Inside such scripts, the global object will be wrapped by the ``contextObject``, retaining all of its existing properties but also having the built-in objects and functions any standard global object has. Outside of scripts run by the vm module, global variables will remain unchanged.

If ``contextObject`` is omitted (or passed explicitly as ``undefined``), a new, empty contextified object will be returned.

**Arguments**

* ``contextObject`` (*Object|vm.constants.DONT\_CONTEXTIFY|undefined*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { createContext, runInContext } = require('node:vm');

   global.globalVar = 3;

   const context = { globalVar: 1 };
   createContext(context);

   runInContext('globalVar *= 2;', context);

   console.log(context);
   // Prints: { globalVar: 2 }

   console.log(global.globalVar);
   // Prints: 3


createScript
~~~~~~~~~~~~

.. code-block:: text

   vm.createScript(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


isContext
~~~~~~~~~

.. code-block:: text

   vm.isContext(object)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the given ``object`` object has been contextified using ``vm.createContext()``, or if it's the global object of a context created using ``vm.constants.DONT_CONTEXTIFY``.

**Arguments**

* ``object`` (*Object*)


measureMemory
~~~~~~~~~~~~~

.. code-block:: text

   vm.measureMemory([options])

*Kind:* function · *Arguments:* 1

Measure the memory known to V8 and used by all contexts known to the current V8 isolate, or the main context.

* Returns: ``Promise`` If the memory is successfully measured, the promise will resolve with an object containing information about the memory usage. Otherwise it will be rejected with an ``ERR_CONTEXT_NOT_INITIALIZED`` error.

**Arguments**

* ``options`` (*Object*) — optional: Optional.

**Example**

.. code-block:: javascript

   const { createContext, measureMemory } = require('node:vm');
   // Measure the memory used by the main context.
   measureMemory({ mode: 'summary' })
     // This is the same as vm.measureMemory()
     .then((result) => {
       // The current format is:
       // {
       //   total: { jsMemoryEstimate: 1601828, jsMemoryRange: [1601828, 5275288] },
       //   WebAssembly: { code: 0, metadata: 33962 },
       // }
       console.log(result);
     });

   const context = createContext({ a: 1 });
   measureMemory({ mode: 'detailed', execution: 'eager' }).then((result) => {
     // Reference the context here so that it won't be GC'ed
     // until the measurement is complete.
     console.log('Context:', context.a);
     // {
     //   total: { jsMemoryEstimate: 1767100, jsMemoryRange: [1767100, 5440560] },
     //   WebAssembly: { code: 0, metadata: 33962 },
     //   current: { jsMemoryEstimate: 1601828, jsMemoryRange: [1601828, 5275288] },
     //   other: [{ jsMemoryEstimate: 165272, jsMemoryRange: [Array] }],
     // }
     console.log(result);
   });


Module
~~~~~~

.. code-block:: text

   vm.Module

*Kind:* class · *Arguments:* 1

This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

The ``vm.Module`` class provides a low-level interface for using ECMAScript modules in VM contexts. It is the counterpart of the ``vm.Script`` class that closely mirrors Module Records as defined in the ECMAScript specification.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const vm = require('node:vm');

   const contextifiedObject = vm.createContext({
     secret: 42,
     print: console.log,
   });

   (async () => {
     // Step 1
     //
     // Create a Module by constructing a new `vm.SourceTextModule` object. This
     // parses the provided source text, throwing a `SyntaxError` if anything goes
     // wrong. By default, a Module is created in the top context. But here, we
     // specify `contextifiedObject` as the context this Module belongs to.
     //
     // Here, we attempt to obtain the default export from the module "foo", and
     // put it into local binding "secret".

     const rootModule = new vm.SourceTextModule(`
       import s from 'foo';
       s;
       print(s);
     `, { context: contextifiedObject });

     // Step 2
     //
     // "Link" the imported dependencies of this Module to it.
     //
     // Obtain the requested dependencies of a SourceTextModule by
     // `sourceTextModule.moduleRequests` and resolve them.
     //
     // Even top-level Modules without dependencies must be explicitly linked. The
     // array passed to `sourceTextModule.linkRequests(modules)` can be
     // empty, however.
     //
     // Note: This is a contrived example in that the resolveAndLinkDependencies
     // creates a new "foo" module every time it is called. In a full-fledged
     // module system, a cache would probably be used to avoid duplicated modules.

     const moduleMap = new Map([
       ['root', rootModule],
     ]);

     function resolveAndLinkDependencies(module) {
       const requestedModules = module.moduleRequests.map((request) => {
         // In a full-fledged module system, the resolveAndLinkDependencies would
         // resolve the module with the module cache key `[specifier, attributes]`.
         // In this example, we just use the specifier as the key.
         const specifier = request.specifier;

         let requestedModule = moduleMap.get(specifier);
         if (requestedModule === undefined) {
           requestedModule = new vm.SourceTextModule(`
             // The "secret" variable refers to the global variable we added to
             // "contextifiedObject" when creating the context.
             export default secret;
           `, { context: module.context });
           moduleMap.set(specifier, requestedModule);
           // Resolve the dependencies of the new module as well.
           resolveAndLinkDependencies(requestedModule);
         }

         return requestedModule;
       });

       module.linkRequests(requestedModules);
     }

     resolveAndLinkDependencies(rootModule);
     rootModule.instantiate();

     // Step 3
     //
     // Evaluate the Module. The evaluate() method returns a promise which will
     // resolve after the module has finished evaluating.

     // Prints 42.
     await rootModule.evaluate();
   })();


runInContext
~~~~~~~~~~~~

.. code-block:: text

   vm.runInContext(code, contextifiedObject[, options])

*Kind:* function · *Arguments:* 3

The ``vm.runInContext()`` method compiles ``code``, runs it within the context of the ``contextifiedObject``, then returns the result. Running code does not have access to the local scope. The ``contextifiedObject`` object *must* have been previously contextified using the ``vm.createContext()`` method.

If ``options`` is a string, then it specifies the filename.

**Arguments**

* ``code`` (*string*): The JavaScript code to compile and run.
* ``contextifiedObject`` (*Object*): The contextified object that will be used
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const { createContext, runInContext } = require('node:vm');

   const contextObject = { globalVar: 1 };
   createContext(contextObject);

   for (let i = 0; i < 10; ++i) {
     runInContext('globalVar *= 2;', contextObject);
   }
   console.log(contextObject);
   // Prints: { globalVar: 1024 }


runInNewContext
~~~~~~~~~~~~~~~

.. code-block:: text

   vm.runInNewContext(code[, contextObject[, options]])

*Kind:* function · *Arguments:* 3

This method is a shortcut to ``(new vm.Script(code, options)).runInContext(vm.createContext(options), options)``. If ``options`` is a string, then it specifies the filename.

It does several things at once:

**Arguments**

* ``code`` (*string*): The JavaScript code to compile and run.
* ``contextObject`` (*Object|vm.constants.DONT\_CONTEXTIFY|undefined*) — optional
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const { runInNewContext, constants } = require('node:vm');

   const contextObject = {
     animal: 'cat',
     count: 2,
   };

   runInNewContext('count += 1; name = "kitty"', contextObject);
   console.log(contextObject);
   // Prints: { animal: 'cat', count: 3, name: 'kitty' }

   // This would throw if the context is created from a contextified object.
   // vm.constants.DONT_CONTEXTIFY allows creating contexts with ordinary global objects that
   // can be frozen.
   const frozenContext = runInNewContext(
     'Object.freeze(globalThis); globalThis;',
     constants.DONT_CONTEXTIFY,
   );


runInThisContext
~~~~~~~~~~~~~~~~

.. code-block:: text

   vm.runInThisContext(code[, options])

*Kind:* function · *Arguments:* 2

``vm.runInThisContext()`` compiles ``code``, runs it within the context of the current ``global`` and returns the result. Running code does not have access to local scope, but does have access to the current ``global`` object.

If ``options`` is a string, then it specifies the filename.

**Arguments**

* ``code`` (*string*): The JavaScript code to compile and run.
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const { runInThisContext } = require('node:vm');
   let localVar = 'initial value';

   const vmResult = runInThisContext('localVar = "vm";');
   console.log(`vmResult: '${vmResult}', localVar: '${localVar}'`);
   // Prints: vmResult: 'vm', localVar: 'initial value'

   const evalResult = eval('localVar = "eval";');
   console.log(`evalResult: '${evalResult}', localVar: '${localVar}'`);
   // Prints: evalResult: 'eval', localVar: 'eval'


Script
~~~~~~

.. code-block:: text

   vm.Script

*Kind:* function · *Arguments:* 2

Type: Documentation-only

The ``produceCachedData`` option is deprecated. Use ``script.createCachedData()`` instead.

This callable takes no arguments.


SourceTextModule
~~~~~~~~~~~~~~~~

.. code-block:: text

   vm.SourceTextModule

*Kind:* class · *Arguments:* 2

This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

* Extends: ``vm.Module``

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const vm = require('node:vm');
   const contextifiedObject = vm.createContext({ secret: 42 });
   (async () => {
     const module = new vm.SourceTextModule(
       'Object.getPrototypeOf(import.meta.prop).secret = secret;',
       {
         context: contextifiedObject,
         initializeImportMeta(meta) {
           // Note: this object is created in the top context. As such,
           // Object.getPrototypeOf(import.meta.prop) points to the
           // Object.prototype in the top context rather than that in
           // the contextified object.
           meta.prop = {};
         },
       });
     // The module has an empty `moduleRequests` array.
     module.linkRequests([]);
     module.instantiate();
     await module.evaluate();
     // Now, Object.prototype.secret will be equal to 42.
     //
     // To fix this problem, replace
     //     meta.prop = {};
     // above with
     //     meta.prop = vm.runInContext('({})', contextifiedObject);
   })();


SyntheticModule
~~~~~~~~~~~~~~~

.. code-block:: text

   vm.SyntheticModule

*Kind:* class · *Arguments:* 3

This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

* Extends: ``vm.Module``

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { SyntheticModule } = require('node:vm');

   const source = '{ "a": 1 }';
   const syntheticModule = new SyntheticModule(['default'], function() {
     const obj = JSON.parse(source);
     this.setExport('default', obj);
   });

   // Use `syntheticModule` in linking
   (async () => {
     await syntheticModule.link(() => {});
     await syntheticModule.evaluate();

     console.log('Default export:', syntheticModule.namespace.default);
   })();


Module class
------------

**Module** — This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

The ``vm.Module`` class provides a low-level interface for using ECMAScript modules in VM contexts. It is the counterpart of the ``vm.Script`` class that closely mirrors Module Records as defined in the ECMAScript specification.

Module._evaluateNow
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   vm.Module._evaluateNow(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Module.evaluate
~~~~~~~~~~~~~~~

.. code-block:: text

   module.evaluate([options])

*Kind:* function · *Arguments:* 1

Evaluate the module and its dependencies. Corresponds to the Evaluate() concrete method field of Cyclic Module Records in the ECMAScript specification.

If the module is a ``vm.SourceTextModule``, ``evaluate()`` must be called after the module has been instantiated; otherwise ``evaluate()`` will return a rejected promise.

**Arguments**

* ``options`` (*Object*) — optional


Module.instantiate
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceTextModule.instantiate()

*Kind:* function · *Arguments:* 0

Instantiate the module with the linked requested modules.

This resolves the imported bindings of the module, including re-exported binding names. When there are any bindings that cannot be resolved, an error would be thrown synchronously.

This callable takes no arguments.


Module.link
~~~~~~~~~~~

.. code-block:: text

   module.link(linker)

*Kind:* function · *Arguments:* 1

Link module dependencies. This method must be called before evaluation, and can only be called once per module.

Use ``sourceTextModule.linkRequests(modules)`` and ``sourceTextModule.instantiate()`` to link modules either synchronously or asynchronously.

**Arguments**

* ``linker`` (*Function*)

**Example**

.. code-block:: javascript

       import foo from 'foo';
       //              ^^^^^ the module specifier
       


Script class
------------

**Script** — Type: Documentation-only

The ``produceCachedData`` option is deprecated. Use ``script.createCachedData()`` instead.

Script.createCachedData
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   script.createCachedData()

*Kind:* function · *Arguments:* 0

Creates a code cache that can be used with the ``Script`` constructor's ``cachedData`` option. Returns a ``Buffer``. This method may be called at any time and any number of times.

The code cache of the ``Script`` doesn't contain any JavaScript observable states. The code cache is safe to be saved along side the script source and used to construct new ``Script`` instances multiple times.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const script = new vm.Script(`
   function add(a, b) {
     return a + b;
   }

   const x = add(1, 2);
   `);

   const cacheWithoutAdd = script.createCachedData();
   // In `cacheWithoutAdd` the function `add()` is marked for full compilation
   // upon invocation.

   script.runInThisContext();

   const cacheWithAdd = script.createCachedData();
   // `cacheWithAdd` contains fully compiled function `add()`.


Script.runInContext
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   script.runInContext(contextifiedObject[, options])

*Kind:* function · *Arguments:* 2

Runs the compiled code contained by the ``vm.Script`` object within the given ``contextifiedObject`` and returns the result. Running code does not have access to local scope.

The following example compiles code that increments a global variable, sets the value of another global variable, then execute the code multiple times. The globals are contained in the ``context`` object.

**Arguments**

* ``contextifiedObject`` (*Object*): A contextified object as returned by the
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { createContext, Script } = require('node:vm');

   const context = {
     animal: 'cat',
     count: 2,
   };

   const script = new Script('count += 1; name = "kitty";');

   createContext(context);
   for (let i = 0; i < 10; ++i) {
     script.runInContext(context);
   }

   console.log(context);
   // Prints: { animal: 'cat', count: 12, name: 'kitty' }


Script.runInNewContext
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   script.runInNewContext([contextObject[, options]])

*Kind:* function · *Arguments:* 2

This method is a shortcut to ``script.runInContext(vm.createContext(options), options)``. It does several things at once:

1. Creates a new context. 2. If ``contextObject`` is an object, contextifies it with the new context. If  ``contextObject`` is undefined, creates a new object and contextifies it. If ``contextObject`` is ``vm.constants.DONT_CONTEXTIFY``, don't contextify anything. 3. Runs the compiled code contained by the ``vm.Script`` object within the created context. The code does not have access to the scope in which this method is called. 4. Returns the result.

**Arguments**

* ``contextObject`` (*Object|vm.constants.DONT\_CONTEXTIFY|undefined*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { constants, Script } = require('node:vm');

   const script = new Script('globalVar = "set"');

   const contexts = [{}, {}, {}];
   contexts.forEach((context) => {
     script.runInNewContext(context);
   });

   console.log(contexts);
   // Prints: [{ globalVar: 'set' }, { globalVar: 'set' }, { globalVar: 'set' }]

   // This would throw if the context is created from a contextified object.
   // constants.DONT_CONTEXTIFY allows creating contexts with ordinary
   // global objects that can be frozen.
   const freezeScript = new Script('Object.freeze(globalThis); globalThis;');
   const frozenContext = freezeScript.runInNewContext(constants.DONT_CONTEXTIFY);


Script.runInThisContext
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   script.runInThisContext([options])

*Kind:* function · *Arguments:* 1

Runs the compiled code contained by the ``vm.Script`` within the context of the current ``global`` object. Running code does not have access to local scope, but *does* have access to the current ``global`` object.

The following example compiles code that increments a ``global`` variable then executes that code multiple times:

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { Script } = require('node:vm');

   global.globalVar = 0;

   const script = new Script('globalVar += 1', { filename: 'myfile.vm' });

   for (let i = 0; i < 1000; ++i) {
     script.runInThisContext();
   }

   console.log(globalVar);

   // 1000


SourceTextModule class
----------------------

**SourceTextModule** — This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

* Extends: ``vm.Module``

SourceTextModule._evaluateNow
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   vm.SourceTextModule._evaluateNow(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


SourceTextModule.createCachedData
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceTextModule.createCachedData()

*Kind:* function · *Arguments:* 0

Creates a code cache that can be used with the ``SourceTextModule`` constructor's ``cachedData`` option. Returns a ``Buffer``. This method may be called any number of times before the module has been evaluated.

The code cache of the ``SourceTextModule`` doesn't contain any JavaScript observable states. The code cache is safe to be saved along side the script source and used to construct new ``SourceTextModule`` instances multiple times.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   // Create an initial module
   const module = new vm.SourceTextModule('const a = 1;');

   // Create cached data from this module
   const cachedData = module.createCachedData();

   // Create a new module using the cached data. The code must be the same.
   const module2 = new vm.SourceTextModule('const a = 1;', { cachedData });


SourceTextModule.hasAsyncGraph
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceTextModule.hasAsyncGraph()

*Kind:* function · *Arguments:* 0

Iterates over the dependency graph and returns ``true`` if any module in its dependencies or this module itself contains top-level ``await`` expressions, otherwise returns ``false``.

The search may be slow if the graph is big enough.

This callable takes no arguments.


SourceTextModule.hasTopLevelAwait
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceTextModule.hasTopLevelAwait()

*Kind:* function · *Arguments:* 0

Returns whether the module itself contains any top-level ``await`` expressions.

This corresponds to the field ``[[HasTLA]]`` in Cyclic Module Record in the ECMAScript specification.

This callable takes no arguments.


SourceTextModule.linkRequests
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sourceTextModule.linkRequests(modules)

*Kind:* function · *Arguments:* 1

Link module dependencies. This method must be called before evaluation, and can only be called once per module.

The order of the module instances in the ``modules`` array should correspond to the order of ``sourceTextModule.moduleRequests`` being resolved. If two module requests have the same specifier and import attributes, they must be resolved with the same module instance or an ``ERR_MODULE_LINK_MISMATCH`` would be thrown. For example, when linking requests for this module:

**Arguments**

* ``modules`` (*vm.Module\[]*): Array of ``vm.Module`` objects that this module depends on.

**Example**

.. code-block:: javascript

   import foo from 'foo';
   import source Foo from 'foo';


SyntheticModule class
---------------------

**SyntheticModule** — This feature is only available with the ``--experimental-vm-modules`` command flag enabled.

* Extends: ``vm.Module``

SyntheticModule._evaluateNow
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   vm.SyntheticModule._evaluateNow(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


SyntheticModule.setExport
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   syntheticModule.setExport(name, value)

*Kind:* function · *Arguments:* 2

This method sets the module export binding slots with the given value.

**Arguments**

* ``name`` (*string*): Name of the export to set.
* ``value`` (*any*): The value to set the export to.

**Example**

.. code-block:: javascript

   const vm = require('node:vm');
   (async () => {
     const m = new vm.SyntheticModule(['x'], () => {
       m.setExport('x', 1);
     });
     await m.evaluate();
     assert.strictEqual(m.namespace.x, 1);
   })();


Properties and constants
------------------------

* ``constants`` — object
* ``constants.DONT_CONTEXTIFY`` — symbol
* ``constants.USE_MAIN_CONTEXT_DEFAULT_LOADER`` — symbol
* ``DONT_CONTEXTIFY`` — symbol

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
