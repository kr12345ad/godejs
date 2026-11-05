async_hooks
===========

> We do not recommend using the ``createHook``, ``AsyncHook``, and > ``executionAsyncResource`` APIs as they have usability issues, safety risks, > and performance implications. Async context tracking use cases are better > served by the stable ``AsyncLocalStorage`` API. If you have a use case for > ``createHook``, ``AsyncHook``, or ``executionAsyncResource`` beyond the context > tracking need solved by ``AsyncLocalStorage`` or diagnostics data currently > provided by Diagnostics Channel, please open an issue at > <https://github.com/nodejs/node/issues> describing your use case so we can > create a more purpose-focused API.

We strongly discourage the use of the ``async_hooks`` API. Other APIs that can cover most of its use cases include:

Loading
-------

.. code-block:: javascript

   const async_hooks = require('node:async_hooks');
   // ESM
   import * as async_hooks from 'node:async_hooks';

Quick start
-----------

.. code-block:: javascript

   const async_hooks = require('node:async_hooks');

API
---

AsyncLocalStorage
~~~~~~~~~~~~~~~~~

.. code-block:: text

   AsyncLocalStorage

*Kind:* class · *Arguments:* 0

The documentation for this class has moved ``AsyncLocalStorage``.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#class-asynclocalstorage>`__

AsyncResource
~~~~~~~~~~~~~

.. code-block:: text

   AsyncResource

*Kind:* class · *Arguments:* 2

The documentation for this class has moved ``AsyncResource``.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#class-asyncresource>`__

createHook
~~~~~~~~~~

.. code-block:: text

   async_hooks.createHook(options)

*Kind:* function · *Arguments:* 1

Registers functions to be called for different lifetime events of each async operation.

The callbacks ``init()``/``before()``/``after()``/``destroy()`` are called for the respective asynchronous event during a resource's lifetime.

**Arguments**

* ``options`` (*Object*): The Hook Callbacks to register

**Example**

.. code-block:: javascript

   const async_hooks = require('node:async_hooks');

   const asyncHook = async_hooks.createHook({
     init(asyncId, type, triggerAsyncId, resource) { },
     destroy(asyncId) { },
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#async_hookscreatehookoptions>`__

executionAsyncId
~~~~~~~~~~~~~~~~

.. code-block:: text

   async_hooks.executionAsyncId()

*Kind:* function · *Arguments:* 0

The ID returned from ``executionAsyncId()`` is related to execution timing, not causality (which is covered by ``triggerAsyncId()``):

Promise contexts may not get precise ``executionAsyncIds`` by default. See the section on promise execution tracking.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const async_hooks = require('node:async_hooks');
   const fs = require('node:fs');

   console.log(async_hooks.executionAsyncId());  // 1 - bootstrap
   const path = '.';
   fs.open(path, 'r', (err, fd) => {
     console.log(async_hooks.executionAsyncId());  // 6 - open()
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#async_hooksexecutionasyncid>`__

executionAsyncResource
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   async_hooks.executionAsyncResource()

*Kind:* function · *Arguments:* 0

Resource objects returned by ``executionAsyncResource()`` are most often internal Node.js handle objects with undocumented APIs. Using any functions or properties on the object is likely to crash your application and should be avoided.

Using ``executionAsyncResource()`` in the top-level execution context will return an empty object as there is no handle or request object to use, but having an object representing the top-level can be helpful.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { open } = require('node:fs');
   const { executionAsyncId, executionAsyncResource } = require('node:async_hooks');

   console.log(executionAsyncId(), executionAsyncResource());  // 1 {}
   open(__filename, 'r', (err, fd) => {
     console.log(executionAsyncId(), executionAsyncResource());  // 7 FSReqWrap
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#async_hooksexecutionasyncresource>`__

triggerAsyncId
~~~~~~~~~~~~~~

.. code-block:: text

   async_hooks.triggerAsyncId()

*Kind:* function · *Arguments:* 0

Promise contexts may not get valid ``triggerAsyncId``\ s by default. See the section on promise execution tracking.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const server = net.createServer((conn) => {
     // Returns the triggerAsyncId of the server, not of the new connection,
     // because the callback runs in the execution scope of the server's
     // MakeCallback().
     async_hooks.triggerAsyncId();

   }).listen(port, () => {
     // Even though all callbacks passed to .listen() are wrapped in a nextTick()
     // the callback itself exists because the call to the server's .listen()
     // was made. So the return value would be the ID of the server.
     async_hooks.triggerAsyncId();
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_hooks.html#async_hookstriggerasyncid>`__

AsyncLocalStorage class
-----------------------

**AsyncLocalStorage** — The documentation for this class has moved ``AsyncLocalStorage``.

AsyncLocalStorage.bind
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   AsyncLocalStorage.bind(fn)

*Kind:* function · *Arguments:* 1

Binds the given function to the current execution context.

**Arguments**

* ``fn`` (*Function*): The function to bind to the current execution context.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#static-method-asynclocalstoragebindfn>`__

AsyncLocalStorage.disable
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncLocalStorage.disable()

*Kind:* function · *Arguments:* 0

Disables the instance of ``AsyncLocalStorage``. All subsequent calls to ``asyncLocalStorage.getStore()`` will return ``undefined`` until ``asyncLocalStorage.run()`` or ``asyncLocalStorage.enterWith()`` is called again.

When calling ``asyncLocalStorage.disable()``, all current contexts linked to the instance will be exited.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asynclocalstoragedisable>`__

AsyncLocalStorage.enterWith
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncLocalStorage.enterWith(store)

*Kind:* function · *Arguments:* 1

Transitions into the context for the remainder of the current synchronous execution and then persists the store through any following asynchronous calls.

Example:

**Arguments**

* ``store`` (*any*)

**Example**

.. code-block:: javascript

   const store = { id: 1 };
   // Replaces previous store with the given store object
   asyncLocalStorage.enterWith(store);
   asyncLocalStorage.getStore(); // Returns the store object
   someAsyncOperation(() => {
     asyncLocalStorage.getStore(); // Returns the same object
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asynclocalstorageenterwithstore>`__

AsyncLocalStorage.exit
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncLocalStorage.exit(callback[, ...args])

*Kind:* function · *Arguments:* 2

Runs a function synchronously outside of a context and returns its return value. The store is not accessible within the callback function or the asynchronous operations created within the callback. Any ``getStore()`` call done within the callback function will always return ``undefined``.

The optional ``args`` are passed to the callback function.

**Arguments**

* ``callback`` (*Function*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   // Within a call to run
   try {
     asyncLocalStorage.getStore(); // Returns the store object or value
     asyncLocalStorage.exit(() => {
       asyncLocalStorage.getStore(); // Returns undefined
       throw new Error();
     });
   } catch (e) {
     asyncLocalStorage.getStore(); // Returns the same object or value
     // The error will be caught here
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asynclocalstorageexitcallback-args>`__

AsyncLocalStorage.getStore
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncLocalStorage.getStore()

*Kind:* function · *Arguments:* 0

Returns the current store. If called outside of an asynchronous context initialized by calling ``asyncLocalStorage.run()`` or ``asyncLocalStorage.enterWith()``, it returns ``undefined``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asynclocalstoragegetstore>`__

AsyncLocalStorage.run
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncLocalStorage.run(store, callback[, ...args])

*Kind:* function · *Arguments:* 3

Runs a function synchronously within a context and returns its return value. The store is not accessible outside of the callback function. The store is accessible to any asynchronous operations created within the callback.

The optional ``args`` are passed to the callback function.

**Arguments**

* ``store`` (*any*)
* ``callback`` (*Function*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const store = { id: 2 };
   try {
     asyncLocalStorage.run(store, () => {
       asyncLocalStorage.getStore(); // Returns the store object
       setTimeout(() => {
         asyncLocalStorage.getStore(); // Returns the store object
       }, 200);
       throw new Error();
     });
   } catch (e) {
     asyncLocalStorage.getStore(); // Returns undefined
     // The error will be caught here
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asynclocalstoragerunstore-callback-args>`__

AsyncResource class
-------------------

**AsyncResource** — The documentation for this class has moved ``AsyncResource``.

AsyncResource.asyncId
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncResource.asyncId()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asyncresourceasyncid>`__

AsyncResource.bind
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncResource.bind(fn[, thisArg])

*Kind:* function · *Arguments:* 2

Binds the given function to execute to this ``AsyncResource``'s scope.

**Arguments**

* ``fn`` (*Function*): The function to bind to the current ``AsyncResource``.
* ``thisArg`` (*any*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asyncresourcebindfn-thisarg>`__

AsyncResource.emitDestroy
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncResource.emitDestroy()

*Kind:* function · *Arguments:* 0

Call all ``destroy`` hooks. This should only ever be called once. An error will be thrown if it is called more than once. This **must** be manually called. If the resource is left to be collected by the GC then the ``destroy`` hooks will never be called.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asyncresourceemitdestroy>`__

AsyncResource.runInAsyncScope
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncResource.runInAsyncScope(fn[, thisArg, ...args])

*Kind:* function · *Arguments:* 3

Call the provided function with the provided arguments in the execution context of the async resource. This will establish the context, trigger the AsyncHooks before callbacks, call the function, trigger the AsyncHooks after callbacks, and then restore the original execution context.

**Arguments**

* ``fn`` (*Function*): The function to call in the execution context of this async
* ``thisArg`` (*any*) — optional: The receiver to be used for the function call.
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asyncresourceruninasyncscopefn-thisarg-args>`__

AsyncResource.triggerAsyncId
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   asyncResource.triggerAsyncId()

*Kind:* function · *Arguments:* 0

<a id="async-resource-worker-pool"></a>

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/async_context.html#asyncresourcetriggerasyncid>`__

Properties and constants
------------------------

* ``asyncWrapProviders`` — object
* ``asyncWrapProviders.CHECKPRIMEREQUEST`` — number
* ``asyncWrapProviders.CIPHERREQUEST`` — number
* ``asyncWrapProviders.DERIVEBITSREQUEST`` — number
* ``asyncWrapProviders.DIRHANDLE`` — number
* ``asyncWrapProviders.DNSCHANNEL`` — number
* ``asyncWrapProviders.ELDHISTOGRAM`` — number
* ``asyncWrapProviders.FILEHANDLE`` — number
* ``asyncWrapProviders.FILEHANDLECLOSEREQ`` — number
* ``asyncWrapProviders.FIXEDSIZEBLOBCOPY`` — number
* ``asyncWrapProviders.FSEVENTWRAP`` — number
* ``asyncWrapProviders.FSREQCALLBACK`` — number
* ``asyncWrapProviders.FSREQPROMISE`` — number
* ``asyncWrapProviders.GETADDRINFOREQWRAP`` — number
* ``asyncWrapProviders.GETNAMEINFOREQWRAP`` — number
* ``asyncWrapProviders.HASHREQUEST`` — number
* ``asyncWrapProviders.HEAPSNAPSHOT`` — number
* ``asyncWrapProviders.HTTP2PING`` — number
* ``asyncWrapProviders.HTTP2SESSION`` — number
* ``asyncWrapProviders.HTTP2SETTINGS`` — number
* ``asyncWrapProviders.HTTP2STREAM`` — number
* ``asyncWrapProviders.HTTPCLIENTREQUEST`` — number
* ``asyncWrapProviders.HTTPINCOMINGMESSAGE`` — number
* ``asyncWrapProviders.INSPECTORJSBINDING`` — number
* ``asyncWrapProviders.JSSTREAM`` — number
* ``asyncWrapProviders.JSUDPWRAP`` — number
* ``asyncWrapProviders.KEYEXPORTREQUEST`` — number
* ``asyncWrapProviders.KEYGENREQUEST`` — number
* ``asyncWrapProviders.KEYPAIRGENREQUEST`` — number
* ``asyncWrapProviders.MESSAGEPORT`` — number
* ``asyncWrapProviders.NONE`` — number
* ``asyncWrapProviders.PBKDF2REQUEST`` — number
* ``asyncWrapProviders.PIPECONNECTWRAP`` — number
* ``asyncWrapProviders.PIPESERVERWRAP`` — number
* ``asyncWrapProviders.PIPEWRAP`` — number
* ``asyncWrapProviders.PROCESSWRAP`` — number
* ``asyncWrapProviders.PROMISE`` — number
* ``asyncWrapProviders.QWRAP`` — number
* ``asyncWrapProviders.RANDOMBYTESREQUEST`` — number
* ``asyncWrapProviders.RANDOMPRIMEREQUEST`` — number
* ``asyncWrapProviders.SCRYPTREQUEST`` — number
* ``asyncWrapProviders.SHUTDOWNWRAP`` — number
* ``asyncWrapProviders.SIGINTWATCHDOG`` — number
* ``asyncWrapProviders.SIGNALWRAP`` — number
* ``asyncWrapProviders.SIGNREQUEST`` — number
* ``asyncWrapProviders.STATWATCHER`` — number
* ``asyncWrapProviders.STREAMPIPE`` — number
* ``asyncWrapProviders.TCPCONNECTWRAP`` — number
* ``asyncWrapProviders.TCPSERVERWRAP`` — number
* ``asyncWrapProviders.TCPWRAP`` — number
* ``asyncWrapProviders.TLSWRAP`` — number
* ``asyncWrapProviders.TTYWRAP`` — number
* ``asyncWrapProviders.UDPSENDWRAP`` — number
* ``asyncWrapProviders.UDPWRAP`` — number
* ``asyncWrapProviders.VERIFYREQUEST`` — number
* ``asyncWrapProviders.WORKER`` — number
* ``asyncWrapProviders.WORKERHEAPSNAPSHOT`` — number
* ``asyncWrapProviders.WRITEWRAP`` — number
* ``asyncWrapProviders.ZLIB`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
