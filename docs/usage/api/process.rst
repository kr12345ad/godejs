process
=======

The ``process`` object provides information about, and control over, the current Node.js process.

Loading
-------

.. code-block:: javascript

   const process = require('node:process');
   // ESM
   import * as process from 'node:process';

Quick start
-----------

.. code-block:: javascript

   const process = require('node:process');

API
---

_rawDebug
~~~~~~~~~

.. code-block:: text

   process._rawDebug(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   console.log(typeof process._rawDebug); // 'function'
   console.log(process._rawDebug.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


abort
~~~~~

.. code-block:: text

   process.abort()

*Kind:* function · *Arguments:* 0

The ``process.abort()`` method causes the Node.js process to exit immediately and generate a core file.

This feature is not available in ``Worker`` threads.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.abort();
   console.log(result);


addListener
~~~~~~~~~~~

.. code-block:: text

   emitter.addListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Alias for ``emitter.on(eventName, listener)``.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.addListener('value', () => {});
   console.log(result);


availableMemory
~~~~~~~~~~~~~~~

.. code-block:: text

   process.availableMemory()

*Kind:* function · *Arguments:* 0

Gets the amount of free memory that is still available to the process (in bytes).

See ``uv_get_available_memory`` for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.availableMemory();
   console.log(result);


binding
~~~~~~~

.. code-block:: text

   process.binding('util').is[...]

*Kind:* function · *Arguments:* 1

Type: Documentation-only (supports ``--pending-deprecation``)

Using ``process.binding()`` in general should be avoided. The type checking methods in particular can be replaced by using ``util.types``.

**Arguments**

* ``util``

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.binding(undefined);
   console.log(result);


chdir
~~~~~

.. code-block:: text

   process.chdir(directory)

*Kind:* function · *Arguments:* 1

The ``process.chdir()`` method changes the current working directory of the Node.js process or throws an exception if doing so fails (for instance, if the specified ``directory`` does not exist).

This feature is not available in ``Worker`` threads.

**Arguments**

* ``directory`` (*string*)

**Example**

.. code-block:: javascript

   const { chdir, cwd } = require('node:process');

   console.log(`Starting directory: ${cwd()}`);
   try {
     chdir('/tmp');
     console.log(`New directory: ${cwd()}`);
   } catch (err) {
     console.error(`chdir: ${err}`);
   }


constrainedMemory
~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.constrainedMemory()

*Kind:* function · *Arguments:* 0

Gets the amount of memory available to the process (in bytes) based on limits imposed by the OS. If there is no such constraint, or the constraint is unknown, ``0`` is returned.

See ``uv_get_constrained_memory`` for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.constrainedMemory();
   console.log(result);


cpuUsage
~~~~~~~~

.. code-block:: text

   process.cpuUsage([previousValue])

*Kind:* function · *Arguments:* 1

The ``process.cpuUsage()`` method returns the user and system CPU time usage of the current process, in an object with properties ``user`` and ``system``, whose values are microsecond values (millionth of a second). These values measure time spent in user and system code respectively, and may end up being greater than actual elapsed time if multiple CPU cores are performing work for this process.

The result of a previous call to ``process.cpuUsage()`` can be passed as the argument to the function, to get a diff reading.

**Arguments**

* ``previousValue`` (*Object*) — optional: A previous return value from calling

**Example**

.. code-block:: javascript

   const { cpuUsage } = require('node:process');

   const startUsage = cpuUsage();
   // { user: 38579, system: 6986 }

   // spin the CPU for 500 milliseconds
   const now = Date.now();
   while (Date.now() - now < 500);

   console.log(cpuUsage(startUsage));
   // { user: 514883, system: 11226 }


cwd
~~~

.. code-block:: text

   process.cwd()

*Kind:* function · *Arguments:* 0

The ``process.cwd()`` method returns the current working directory of the Node.js process.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { cwd } = require('node:process');

   console.log(`Current directory: ${cwd()}`);


disconnect
~~~~~~~~~~

.. code-block:: text

   process.disconnect()

*Kind:* function · *Arguments:* 0

If the Node.js process is spawned with an IPC channel (see the Child Process and Cluster documentation), the ``process.disconnect()`` method will close the IPC channel to the parent process, allowing the child process to exit gracefully once there are no other connections keeping it alive.

The effect of calling ``process.disconnect()`` is the same as calling ``ChildProcess.disconnect()`` from the parent process.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.disconnect();
   console.log(result);


dlopen
~~~~~~

.. code-block:: text

   process.dlopen(module, filename[, flags])

*Kind:* function · *Arguments:* 3

The ``process.dlopen()`` method allows dynamically loading shared objects. It is primarily used by ``require()`` to load C++ Addons, and should not be used directly, except in special cases. In other words, ``require()`` should be preferred over ``process.dlopen()`` unless there are specific reasons such as custom dlopen flags or loading from ES modules.

The ``flags`` argument is an integer that allows to specify dlopen behavior. See the ``os.constants.dlopen`` documentation for details.

**Arguments**

* ``module`` (*Object*)
* ``filename`` (*string*)
* ``flags`` (*os.constants.dlopen*) — optional: **Default:** ``os.constants.dlopen.RTLD_LAZY``

**Example**

.. code-block:: javascript

   const { dlopen } = require('node:process');
   const { constants } = require('node:os');
   const { join } = require('node:path');

   const module = { exports: {} };
   dlopen(module, join(__dirname, 'local.node'), constants.dlopen.RTLD_NOW);
   module.exports.foo();


emit
~~~~

.. code-block:: text

   emitter.emit(eventName[, ...args])

*Kind:* function · *Arguments:* 2

Synchronously calls each of the listeners registered for the event named ``eventName``, in the order they were registered, passing the supplied arguments to each.

Returns ``true`` if the event had listeners, ``false`` otherwise.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const myEmitter = new EventEmitter();

   // First listener
   myEmitter.on('event', function firstListener() {
     console.log('Helloooo! first listener');
   });
   // Second listener
   myEmitter.on('event', function secondListener(arg1, arg2) {
     console.log(`event with parameters ${arg1}, ${arg2} in second listener`);
   });
   // Third listener
   myEmitter.on('event', function thirdListener(...args) {
     const parameters = args.join(', ');
     console.log(`event with parameters ${parameters} in third listener`);
   });

   console.log(myEmitter.listeners('event'));

   myEmitter.emit('event', 1, 2, 3, 4, 5);

   // Prints:
   // [
   //   [Function: firstListener],
   //   [Function: secondListener],
   //   [Function: thirdListener]
   // ]
   // Helloooo! first listener
   // event with parameters 1, 2 in second listener
   // event with parameters 1, 2, 3, 4, 5 in third listener


emitWarning
~~~~~~~~~~~

.. code-block:: text

   process.emitWarning(warning[, options])

*Kind:* function · *Arguments:* 2

The ``process.emitWarning()`` method can be used to emit custom or application specific process warnings. These can be listened for by adding a handler to the ``'warning'`` event.

In this example, an ``Error`` object is generated internally by ``process.emitWarning()`` and passed through to the ``'warning'`` handler.

**Arguments**

* ``warning`` (*string|Error*): The warning to emit.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { emitWarning } = require('node:process');

   // Emit a warning with a code and additional detail.
   emitWarning('Something happened!', {
     code: 'MY_WARNING',
     detail: 'This is some additional information',
   });
   // Emits:
   // (node:56338) [MY_WARNING] Warning: Something happened!
   // This is some additional information


emitWarningSync
~~~~~~~~~~~~~~~

.. code-block:: text

   process.emitWarningSync(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.emitWarningSync();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


eventNames
~~~~~~~~~~

.. code-block:: text

   emitter.eventNames()

*Kind:* function · *Arguments:* 0

Returns an array listing the events for which the emitter has registered listeners.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');

   const myEE = new EventEmitter();
   myEE.on('foo', () => {});
   myEE.on('bar', () => {});

   const sym = Symbol('symbol');
   myEE.on(sym, () => {});

   console.log(myEE.eventNames());
   // Prints: [ 'foo', 'bar', Symbol(symbol) ]


execve
~~~~~~

.. code-block:: text

   process.execve(file[, args[, env]])

*Kind:* function · *Arguments:* 3

Replaces the current process with a new process.

This is achieved by using the ``execve`` POSIX function and therefore no memory or other resources from the current process are preserved, except for the standard input, standard output and standard error file descriptor.

**Arguments**

* ``file`` (*string*): The name or path of the executable file to run.
* ``args`` (*string\[]*) — optional: List of string arguments. No argument can contain a null-byte (``\u0000``).
* ``env`` (*Object*) — optional: Environment key-value pairs.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.execve('./file.txt', [], {});
   console.log(result);


exit
~~~~

.. code-block:: text

   process.exit(code)

*Kind:* function · *Arguments:* 1

Type: End-of-Life

Values other than ``undefined``, ``null``, integer numbers, and integer strings (e.g., ``'1'``) are deprecated as value for the ``code`` parameter in ``process.exit()`` and as value to assign to ``process.exitCode``.

**Arguments**

* ``code``

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.exit(undefined);
   console.log(result);


getActiveRequests
~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.getActiveRequests(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.getActiveRequests();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


getActiveResourcesInfo
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.getActiveResourcesInfo()

*Kind:* function · *Arguments:* 0

The ``process.getActiveResourcesInfo()`` method returns an array of strings containing the types of the active resources that are currently keeping the event loop alive.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { getActiveResourcesInfo } = require('node:process');
   const { setTimeout } = require('node:timers');

   console.log('Before:', getActiveResourcesInfo());
   setTimeout(() => {}, 1000);
   console.log('After:', getActiveResourcesInfo());
   // Prints:
   //   Before: [ 'TTYWrap', 'TTYWrap', 'TTYWrap' ]
   //   After: [ 'TTYWrap', 'TTYWrap', 'TTYWrap', 'Timeout' ]


getBuiltinModule
~~~~~~~~~~~~~~~~

.. code-block:: text

   process.getBuiltinModule(id)

*Kind:* function · *Arguments:* 1

``process.getBuiltinModule(id)`` provides a way to load built-in modules in a globally available function. ES Modules that need to support other environments can use it to conditionally load a Node.js built-in when it is run in Node.js, without having to deal with the resolution error that can be thrown by ``import`` in a non-Node.js environment or having to use dynamic ``import()`` which either turns the module into an asynchronous module, or turns a synchronous API into an asynchronous one.

If ``id`` specifies a built-in module available in the current Node.js process, ``process.getBuiltinModule(id)`` method returns the corresponding built-in module. If ``id`` does not correspond to any built-in module, ``undefined`` is returned.

**Arguments**

* ``id`` (*string*): ID of the built-in module being requested.

**Example**

.. code-block:: javascript

   if (globalThis.process?.getBuiltinModule) {
     // Run in Node.js, use the Node.js fs module.
     const fs = globalThis.process.getBuiltinModule('fs');
     // If `require()` is needed to load user-modules, use createRequire()
     const module = globalThis.process.getBuiltinModule('module');
     const require = module.createRequire(import.meta.url);
     const foo = require('foo');
   }


getegid
~~~~~~~

.. code-block:: text

   process.getegid()

*Kind:* function · *Arguments:* 0

The ``process.getegid()`` method returns the numerical effective group identity of the Node.js process. (See getegid(2).)

This function is only available on POSIX platforms (i.e. not Windows or Android).

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (process.getegid) {
     console.log(`Current gid: ${process.getegid()}`);
   }


geteuid
~~~~~~~

.. code-block:: text

   process.geteuid()

*Kind:* function · *Arguments:* 0

The ``process.geteuid()`` method returns the numerical effective user identity of the process. (See geteuid(2).)

This function is only available on POSIX platforms (i.e. not Windows or Android).

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (process.geteuid) {
     console.log(`Current uid: ${process.geteuid()}`);
   }


getgid
~~~~~~

.. code-block:: text

   process.getgid()

*Kind:* function · *Arguments:* 0

The ``process.getgid()`` method returns the numerical group identity of the process. (See getgid(2).)

This function is only available on POSIX platforms (i.e. not Windows or Android).

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (process.getgid) {
     console.log(`Current gid: ${process.getgid()}`);
   }


getgroups
~~~~~~~~~

.. code-block:: text

   process.getgroups()

*Kind:* function · *Arguments:* 0

The ``process.getgroups()`` method returns an array with the supplementary group IDs. POSIX leaves it unspecified if the effective group ID is included but Node.js ensures it always is.

This function is only available on POSIX platforms (i.e. not Windows or Android).

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (process.getgroups) {
     console.log(process.getgroups()); // [ 16, 21, 297 ]
   }


getuid
~~~~~~

.. code-block:: text

   process.getuid()

*Kind:* function · *Arguments:* 0

The ``process.getuid()`` method returns the numeric user identity of the process. (See getuid(2).)

This function not available on Windows.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (process.getuid) {
     console.log(`Current uid: ${process.getuid()}`);
   }


hasUncaughtExceptionCaptureCallback
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.hasUncaughtExceptionCaptureCallback()

*Kind:* function · *Arguments:* 0

Indicates whether a callback has been set using ``process.setUncaughtExceptionCaptureCallback()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.hasUncaughtExceptionCaptureCallback();
   console.log(result);


hrtime
~~~~~~

.. code-block:: text

   process.hrtime([time])

*Kind:* function · *Arguments:* 1

This is the legacy version of ``process.hrtime.bigint()`` before ``bigint`` was introduced in JavaScript.

The ``process.hrtime()`` method returns the current high-resolution real time in a ``[seconds, nanoseconds]`` tuple ``Array``, where ``nanoseconds`` is the remaining part of the real time that can't be represented in second precision.

**Arguments**

* ``time`` (*integer\[]*) — optional: The result of a previous call to ``process.hrtime()``

**Example**

.. code-block:: javascript

   const { hrtime } = require('node:process');

   const NS_PER_SEC = 1e9;
   const time = hrtime();
   // [ 1800216, 25 ]

   setTimeout(() => {
     const diff = hrtime(time);
     // [ 1, 552 ]

     console.log(`Benchmark took ${diff[0] * NS_PER_SEC + diff[1]} nanoseconds`);
     // Benchmark took 1000000552 nanoseconds
   }, 1000);


initgroups
~~~~~~~~~~

.. code-block:: text

   process.initgroups(user, extraGroup)

*Kind:* function · *Arguments:* 2

The ``process.initgroups()`` method reads the ``/etc/group`` file and initializes the group access list, using all groups of which the user is a member. This is a privileged operation that requires that the Node.js process either have ``root`` access or the ``CAP_SETGID`` capability.

Use care when dropping privileges:

**Arguments**

* ``user`` (*string|number*): The user name or numeric identifier.
* ``extraGroup`` (*string|number*): A group name or numeric identifier.

**Example**

.. code-block:: javascript

   const { getgroups, initgroups, setgid } = require('node:process');

   console.log(getgroups());         // [ 0 ]
   initgroups('nodeuser', 1000);     // switch user
   console.log(getgroups());         // [ 27, 30, 46, 1000, 0 ]
   setgid(1000);                     // drop root gid
   console.log(getgroups());         // [ 27, 30, 46, 1000 ]


kill
~~~~

.. code-block:: text

   process.kill(pid[, signal])

*Kind:* function · *Arguments:* 2

The ``process.kill()`` method sends the ``signal`` to the process identified by ``pid``.

Signal names are strings such as ``'SIGINT'`` or ``'SIGHUP'``. See Signal Events and kill(2) for more information.

**Arguments**

* ``pid`` (*number*): A process ID
* ``signal`` (*string|number*) — optional: The signal to send, either as a string or number.

**Example**

.. code-block:: javascript

   process.on('SIGHUP', () => {
     console.log('Got SIGHUP signal.');
   });

   setTimeout(() => {
     console.log('Exiting.');
     process.exit(0);
   }, 100);

   process.kill(process.pid, 'SIGHUP');


listenerCount
~~~~~~~~~~~~~

.. code-block:: text

   EventEmitter.listenerCount()

*Kind:* function · *Arguments:* 0

Type: Revoked

The ``events.listenerCount(emitter, eventName)`` API was deprecated, as it provided identical functionality to ``emitter.listenerCount(eventName)``. The deprecation was revoked because this function has been repurposed to also accept ``EventTarget`` arguments.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.listenerCount();
   console.log(result);


listeners
~~~~~~~~~

.. code-block:: text

   emitter.listeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``.

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   server.on('connection', (stream) => {
     console.log('someone connected!');
   });
   console.log(util.inspect(server.listeners('connection')));
   // Prints: [ [Function] ]


loadEnvFile
~~~~~~~~~~~

.. code-block:: text

   process.loadEnvFile(path)

*Kind:* function · *Arguments:* 1

Loads the ``.env`` file into ``process.env``. Usage of ``NODE_OPTIONS`` in the ``.env`` file will not have any effect on Node.js.

**Arguments**

* ``path`` (*string | URL | Buffer | undefined*): . **Default:** ``'./.env'``

**Example**

.. code-block:: javascript

   const { loadEnvFile } = require('node:process');
   loadEnvFile();


memoryUsage
~~~~~~~~~~~

.. code-block:: text

   process.memoryUsage()

*Kind:* function · *Arguments:* 0

Returns an object describing the memory usage of the Node.js process measured in bytes.

* ``heapTotal`` and ``heapUsed`` refer to V8's memory usage. * ``external`` refers to the memory usage of C++ objects bound to JavaScript objects managed by V8. * ``rss``, Resident Set Size, is the amount of space occupied in the main memory device (that is a subset of the total allocated memory) for the process, including all C++ and JavaScript objects and code. * ``arrayBuffers`` refers to memory allocated for ``ArrayBuffer``\ s and ``SharedArrayBuffer``\ s, including all Node.js ``Buffer``\ s. This is also included in the ``external`` value. When Node.js is used as an embedded library, this value may be ``0`` because allocations for ``ArrayBuffer``\ s may not be tracked in that case.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { memoryUsage } = require('node:process');

   console.log(memoryUsage());
   // Prints:
   // {
   //  rss: 4935680,
   //  heapTotal: 1826816,
   //  heapUsed: 650472,
   //  external: 49879,
   //  arrayBuffers: 9386
   // }


nextTick
~~~~~~~~

.. code-block:: text

   process.nextTick(callback[, ...args])

*Kind:* function · *Arguments:* 2

``process.nextTick()`` adds ``callback`` to the "next tick queue". This queue is fully drained after the current operation on the JavaScript stack runs to completion and before the event loop is allowed to continue. It's possible to create an infinite loop if one were to recursively call ``process.nextTick()``. See the Event Loop guide for more background.

This is important when developing APIs in order to give users the opportunity to assign event handlers *after* an object has been constructed but before any I/O has occurred:

**Arguments**

* ``callback`` (*Function*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const { nextTick } = require('node:process');

   console.log('start');
   nextTick(() => {
     console.log('nextTick callback');
   });
   console.log('scheduled');
   // Output:
   // start
   // scheduled
   // nextTick callback


off
~~~

.. code-block:: text

   emitter.off(eventName, listener)

*Kind:* function · *Arguments:* 2

Alias for ``emitter.removeListener()``.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.off('value', () => {});
   console.log(result);


on
~~~

.. code-block:: text

   process.on('multipleResolves', handler)

*Kind:* function · *Arguments:* 2

Type: Runtime

This event was deprecated because it did not work with V8 promise combinators which diminished its usefulness.

**Arguments**

* ``multipleResolves``
* ``handler``

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.on(undefined, () => {});
   console.log(result);


once
~~~~

.. code-block:: text

   emitter.once(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds a **one-time** ``listener`` function for the event named ``eventName``. The next time ``eventName`` is triggered, this listener is removed and then invoked.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const myEE = new EventEmitter();
   myEE.once('foo', () => console.log('a'));
   myEE.prependOnceListener('foo', () => console.log('b'));
   myEE.emit('foo');
   // Prints:
   //   b
   //   a


prependListener
~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds the ``listener`` function to the *beginning* of the listeners array for the event named ``eventName``. No checks are made to see if the ``listener`` has already been added. Multiple calls passing the same combination of ``eventName`` and ``listener`` will result in the ``listener`` being added, and called, multiple times.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependListener('connection', (stream) => {
     console.log('someone connected!');
   });


prependOnceListener
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependOnceListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds a **one-time** ``listener`` function for the event named ``eventName`` to the *beginning* of the listeners array. The next time ``eventName`` is triggered, this listener is removed, and then invoked.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependOnceListener('connection', (stream) => {
     console.log('Ah, we have our first user!');
   });


rawDebug
~~~~~~~~

.. code-block:: text

   process.rawDebug(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.rawDebug();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


rawListeners
~~~~~~~~~~~~

.. code-block:: text

   emitter.rawListeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``, including any wrappers (such as those created by ``.once()``).

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const emitter = new EventEmitter();
   emitter.once('log', () => console.log('log once'));

   // Returns a new Array with a function `onceWrapper` which has a property
   // `listener` which contains the original listener bound above
   const listeners = emitter.rawListeners('log');
   const logFnWrapper = listeners[0];

   // Logs "log once" to the console and does not unbind the `once` event
   logFnWrapper.listener();

   // Logs "log once" to the console and removes the listener
   logFnWrapper();

   emitter.on('log', () => console.log('log persistently'));
   // Will return a new Array with a single function bound by `.on()` above
   const newListeners = emitter.rawListeners('log');

   // Logs "log persistently" twice
   newListeners[0]();
   emitter.emit('log');


reallyExit
~~~~~~~~~~

.. code-block:: text

   process.reallyExit(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.reallyExit();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ref
~~~

.. code-block:: text

   process.ref(maybeRefable)

*Kind:* function · *Arguments:* 1

An object is "refable" if it implements the Node.js "Refable protocol". Specifically, this means that the object implements the ``Symbol.for('nodejs.ref')`` and ``Symbol.for('nodejs.unref')`` methods. "Ref'd" objects will keep the Node.js event loop alive, while "unref'd" objects will not. Historically, this was implemented by using ``ref()`` and ``unref()`` methods directly on the objects. This pattern, however, is being deprecated in favor of the "Refable protocol" in order to better support Web Platform API types whose APIs cannot be modified to add ``ref()`` and ``unref()`` methods but still need to support that behavior.

**Arguments**

* ``maybeRefable`` (*any*): An object that may be "refable".

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.ref(undefined);
   console.log(result);


register
~~~~~~~~

.. code-block:: text

   process.finalization.register(ref, callback)

*Kind:* function · *Arguments:* 2

This function registers a callback to be called when the process emits the ``exit`` event if the ``ref`` object was not garbage collected. If the object ``ref`` was garbage collected before the ``exit`` event is emitted, the callback will be removed from the finalization registry, and it will not be called on process exit.

Inside the callback you can release the resources allocated by the ``ref`` object. Be aware that all limitations applied to the ``beforeExit`` event are also applied to the ``callback`` function, this means that there is a possibility that the callback will not be called under special circumstances.

**Arguments**

* ``ref`` (*Object | Function*): The reference to the resource that is being tracked.
* ``callback`` (*Function*): The callback function to be called when the resource

**Example**

.. code-block:: javascript

   const { finalization } = require('node:process');

   // Please make sure that the function passed to finalization.register()
   // does not create a closure around unnecessary objects.
   function onFinalize(obj, event) {
     // You can do whatever you want with the object
     obj.dispose();
   }

   function setup() {
     // This object can be safely garbage collected,
     // and the resulting shutdown function will not be called.
     // There are no leaks.
     const myDisposableObject = {
       dispose() {
         // Free your resources synchronously
       },
     };

     finalization.register(myDisposableObject, onFinalize);
   }

   setup();


registerBeforeExit
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.finalization.registerBeforeExit(ref, callback)

*Kind:* function · *Arguments:* 2

This function behaves exactly like the ``register``, except that the callback will be called when the process emits the ``beforeExit`` event if ``ref`` object was not garbage collected.

Be aware that all limitations applied to the ``beforeExit`` event are also applied to the ``callback`` function, this means that there is a possibility that the callback will not be called under special circumstances.

**Arguments**

* ``ref`` (*Object | Function*): The reference
* ``callback`` (*Function*): The callback function to be called when the resource

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.registerBeforeExit(() => {}, () => {});
   console.log(result);


removeAllListeners
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.removeAllListeners([eventName])

*Kind:* function · *Arguments:* 1

Removes all listeners, or those of the specified ``eventName``.

It is bad practice to remove listeners added elsewhere in the code, particularly when the ``EventEmitter`` instance was created by some other component or module (e.g. sockets or file streams).

**Arguments**

* ``eventName`` (*string|symbol*) — optional

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.removeAllListeners('value');
   console.log(result);


removeListener
~~~~~~~~~~~~~~

.. code-block:: text

   emitter.removeListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Removes the specified ``listener`` from the listener array for the event named ``eventName``.

``removeListener()`` will remove, at most, one instance of a listener from the listener array. If any single listener has been added multiple times to the listener array for the specified ``eventName``, then ``removeListener()`` must be called multiple times to remove each instance.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   class MyEmitter extends EventEmitter {}
   const myEmitter = new MyEmitter();

   const callbackA = () => {
     console.log('A');
     myEmitter.removeListener('event', callbackB);
   };

   const callbackB = () => {
     console.log('B');
   };

   myEmitter.on('event', callbackA);

   myEmitter.on('event', callbackB);

   // callbackA removes listener callbackB but it will still be called.
   // Internal listener array at time of emit [callbackA, callbackB]
   myEmitter.emit('event');
   // Prints:
   //   A
   //   B

   // callbackB is now removed.
   // Internal listener array [callbackA]
   myEmitter.emit('event');
   // Prints:
   //   A


resourceUsage
~~~~~~~~~~~~~

.. code-block:: text

   process.resourceUsage()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { resourceUsage } = require('node:process');

   console.log(resourceUsage());
   /*
     Will output:
     {
       userCPUTime: 82872,
       systemCPUTime: 4143,
       maxRSS: 33164,
       sharedMemorySize: 0,
       unsharedDataSize: 0,
       unsharedStackSize: 0,
       minorPageFault: 2469,
       majorPageFault: 0,
       swappedOut: 0,
       fsRead: 0,
       fsWrite: 8,
       ipcSent: 0,
       ipcReceived: 0,
       signalsCount: 0,
       voluntaryContextSwitches: 79,
       involuntaryContextSwitches: 1
     }
   */


rss
~~~

.. code-block:: text

   process.memoryUsage.rss()

*Kind:* function · *Arguments:* 0

The ``process.memoryUsage.rss()`` method returns an integer representing the Resident Set Size (RSS) in bytes.

The Resident Set Size, is the amount of space occupied in the main memory device (that is a subset of the total allocated memory) for the process, including all C++ and JavaScript objects and code.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { memoryUsage } = require('node:process');

   console.log(memoryUsage.rss());
   // 35655680


send
~~~~

.. code-block:: text

   process.send(message[, sendHandle[, options]][, callback])

*Kind:* function · *Arguments:* 4

If Node.js is spawned with an IPC channel, the ``process.send()`` method can be used to send messages to the parent process. Messages will be received as a ``'message'`` event on the parent's ``ChildProcess`` object.

If Node.js was not spawned with an IPC channel, ``process.send`` will be ``undefined``.

**Arguments**

* ``message`` (*Object*)
* ``sendHandle`` (*net.Server|net.Socket*) — optional
* ``options`` (*Object*) — optional: used to parameterize the sending of certain types of
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.send({}, undefined, {}, () => {});
   console.log(result);


setegid
~~~~~~~

.. code-block:: text

   process.setegid(id)

*Kind:* function · *Arguments:* 1

The ``process.setegid()`` method sets the effective group identity of the process. (See setegid(2).) The ``id`` can be passed as either a numeric ID or a group name string. If a group name is specified, this method blocks while resolving the associated a numeric ID.

This function is only available on POSIX platforms (i.e. not Windows or Android). This feature is not available in ``Worker`` threads.

**Arguments**

* ``id`` (*string|number*): A group name or ID

**Example**

.. code-block:: javascript

   if (process.getegid && process.setegid) {
     console.log(`Current gid: ${process.getegid()}`);
     try {
       process.setegid(501);
       console.log(`New gid: ${process.getegid()}`);
     } catch (err) {
       console.error(`Failed to set gid: ${err}`);
     }
   }


seteuid
~~~~~~~

.. code-block:: text

   process.seteuid(id)

*Kind:* function · *Arguments:* 1

The ``process.seteuid()`` method sets the effective user identity of the process. (See seteuid(2).) The ``id`` can be passed as either a numeric ID or a username string. If a username is specified, the method blocks while resolving the associated numeric ID.

This function is only available on POSIX platforms (i.e. not Windows or Android). This feature is not available in ``Worker`` threads.

**Arguments**

* ``id`` (*string|number*): A user name or ID

**Example**

.. code-block:: javascript

   if (process.geteuid && process.seteuid) {
     console.log(`Current uid: ${process.geteuid()}`);
     try {
       process.seteuid(501);
       console.log(`New uid: ${process.geteuid()}`);
     } catch (err) {
       console.error(`Failed to set uid: ${err}`);
     }
   }


setgid
~~~~~~

.. code-block:: text

   process.setgid(id)

*Kind:* function · *Arguments:* 1

The ``process.setgid()`` method sets the group identity of the process. (See setgid(2).) The ``id`` can be passed as either a numeric ID or a group name string. If a group name is specified, this method blocks while resolving the associated numeric ID.

This function is only available on POSIX platforms (i.e. not Windows or Android). This feature is not available in ``Worker`` threads.

**Arguments**

* ``id`` (*string|number*): The group name or ID

**Example**

.. code-block:: javascript

   if (process.getgid && process.setgid) {
     console.log(`Current gid: ${process.getgid()}`);
     try {
       process.setgid(501);
       console.log(`New gid: ${process.getgid()}`);
     } catch (err) {
       console.error(`Failed to set gid: ${err}`);
     }
   }


setgroups
~~~~~~~~~

.. code-block:: text

   process.setgroups(groups)

*Kind:* function · *Arguments:* 1

The ``process.setgroups()`` method sets the supplementary group IDs for the Node.js process. This is a privileged operation that requires the Node.js process to have ``root`` or the ``CAP_SETGID`` capability.

The ``groups`` array can contain numeric group IDs, group names, or both.

**Arguments**

* ``groups`` (*integer\[]*)

**Example**

.. code-block:: javascript

   if (process.getgroups && process.setgroups) {
     try {
       process.setgroups([501]);
       console.log(process.getgroups()); // new groups
     } catch (err) {
       console.error(`Failed to set groups: ${err}`);
     }
   }


setSourceMapsEnabled
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.setSourceMapsEnabled(val)

*Kind:* function · *Arguments:* 1

This function enables or disables the Source Map support for stack traces.

It provides same features as launching Node.js process with commandline options ``--enable-source-maps``.

**Arguments**

* ``val`` (*boolean*)

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.setSourceMapsEnabled(true);
   console.log(result);


setuid
~~~~~~

.. code-block:: text

   process.setuid(id)

*Kind:* function · *Arguments:* 1

The ``process.setuid(id)`` method sets the user identity of the process. (See setuid(2).) The ``id`` can be passed as either a numeric ID or a username string. If a username is specified, the method blocks while resolving the associated numeric ID.

This function is only available on POSIX platforms (i.e. not Windows or Android). This feature is not available in ``Worker`` threads.

**Arguments**

* ``id`` (*integer | string*)

**Example**

.. code-block:: javascript

   if (process.getuid && process.setuid) {
     console.log(`Current uid: ${process.getuid()}`);
     try {
       process.setuid(501);
       console.log(`New uid: ${process.getuid()}`);
     } catch (err) {
       console.error(`Failed to set uid: ${err}`);
     }
   }


setUncaughtExceptionCaptureCallback
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.setUncaughtExceptionCaptureCallback(fn)

*Kind:* function · *Arguments:* 1

The ``process.setUncaughtExceptionCaptureCallback()`` function sets a function that will be invoked when an uncaught exception occurs, which will receive the exception value itself as its first argument.

If such a function is set, the ``'uncaughtException'`` event will not be emitted. If ``--abort-on-uncaught-exception`` was passed from the command line or set through ``v8.setFlagsFromString()``, the process will not abort. Actions configured to take place on exceptions such as report generations will be affected too

**Arguments**

* ``fn`` (*Function|null*)

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.setUncaughtExceptionCaptureCallback(() => {});
   console.log(result);


threadCpuUsage
~~~~~~~~~~~~~~

.. code-block:: text

   process.threadCpuUsage([previousValue])

*Kind:* function · *Arguments:* 1

The ``process.threadCpuUsage()`` method returns the user and system CPU time usage of the current worker thread, in an object with properties ``user`` and ``system``, whose values are microsecond values (millionth of a second).

The result of a previous call to ``process.threadCpuUsage()`` can be passed as the argument to the function, to get a diff reading.

**Arguments**

* ``previousValue`` (*Object*) — optional: A previous return value from calling

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.threadCpuUsage({});
   console.log(result);


traceProcessWarnings
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.traceProcessWarnings

*Kind:* function · *Arguments:* 0

The ``process.traceProcessWarnings`` property indicates whether the ``--trace-warnings`` flag is set on the current Node.js process. This property allows programmatic control over the tracing of warnings, enabling or disabling stack traces for warnings at runtime.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   // Enable trace warnings
   process.traceProcessWarnings = true;

   // Emit a warning with a stack trace
   process.emitWarning('Warning with stack trace');

   // Disable trace warnings
   process.traceProcessWarnings = false;


umask
~~~~~

.. code-block:: text

   process.umask()

*Kind:* function · *Arguments:* 1

Type: Documentation-only

Calling ``process.umask()`` with no argument causes the process-wide umask to be written twice. This introduces a race condition between threads, and is a potential security vulnerability. There is no safe, cross-platform alternative API.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.umask(undefined);
   console.log(result);


unref
~~~~~

.. code-block:: text

   process.unref(maybeRefable)

*Kind:* function · *Arguments:* 1

An object is "unrefable" if it implements the Node.js "Refable protocol". Specifically, this means that the object implements the ``Symbol.for('nodejs.ref')`` and ``Symbol.for('nodejs.unref')`` methods. "Ref'd" objects will keep the Node.js event loop alive, while "unref'd" objects will not. Historically, this was implemented by using ``ref()`` and ``unref()`` methods directly on the objects. This pattern, however, is being deprecated in favor of the "Refable protocol" in order to better support Web Platform API types whose APIs cannot be modified to add ``ref()`` and ``unref()`` methods but still need to support that behavior.

**Arguments**

* ``maybeRefable`` (*any*): An object that may be "unref'd".

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.unref(undefined);
   console.log(result);


unregister
~~~~~~~~~~

.. code-block:: text

   process.finalization.unregister(ref)

*Kind:* function · *Arguments:* 1

This function remove the register of the object from the finalization registry, so the callback will not be called anymore.

**Arguments**

* ``ref`` (*Object | Function*): The reference

**Example**

.. code-block:: javascript

   const { finalization } = require('node:process');

   // Please make sure that the function passed to finalization.register()
   // does not create a closure around unnecessary objects.
   function onFinalize(obj, event) {
     // You can do whatever you want with the object
     obj.dispose();
   }

   function setup() {
     // This object can be safely garbage collected,
     // and the resulting shutdown function will not be called.
     // There are no leaks.
     const myDisposableObject = {
       dispose() {
         // Free your resources synchronously
       },
     };

     finalization.register(myDisposableObject, onFinalize);

     // Do something

     myDisposableObject.dispose();
     finalization.unregister(myDisposableObject);
   }

   setup();


uptime
~~~~~~

.. code-block:: text

   process.uptime()

*Kind:* function · *Arguments:* 0

The ``process.uptime()`` method returns the number of seconds the current Node.js process has been running.

The return value includes fractions of a second. Use ``Math.floor()`` to get whole seconds.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.uptime();
   console.log(result);


allowedNodeEnvironmentFlags
---------------------------

**allowedNodeEnvironmentFlags** — The ``process.allowedNodeEnvironmentFlags`` property is a special, read-only ``Set`` of flags allowable within the ``NODE_OPTIONS`` environment variable.

``process.allowedNodeEnvironmentFlags`` extends ``Set``, but overrides ``Set.prototype.has`` to recognize several different possible flag representations. ``process.allowedNodeEnvironmentFlags.has()`` will return ``true`` in the following cases:

allowedNodeEnvironmentFlags.add
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.add(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.add();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.add
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.add(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.add(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.clear
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.clear(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.clear();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.clear
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.clear(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.clear();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.delete
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.delete(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.delete();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.delete
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.delete(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.delete(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.difference
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.difference(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.difference(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.entries
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.entries(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.entries();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.entries
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.entries(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.entries();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.forEach
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.forEach(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.forEach(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.forEach
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.forEach(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.forEach(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.has
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.permission.has(scope[, reference])

*Kind:* function · *Arguments:* 2

Verifies that the process is able to access the given scope and reference. If no reference is provided, a global scope is assumed, for instance, ``process.permission.has('fs.read')`` will check if the process has ALL file system read permissions.

In audit mode (``--permission-audit``), this method still returns the actual permission status, but denied operations will not throw ``ERR_ACCESS_DENIED``.

**Arguments**

* ``scope`` (*string*)
* ``reference`` (*string*) — optional

**Example**

.. code-block:: javascript

   // Check if the process has permission to read the README file
   process.permission.has('fs.read', './README.md');
   // Check if the process has read permission operations
   process.permission.has('fs.read');


allowedNodeEnvironmentFlags.has
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.permission.has(scope[, reference])

*Kind:* function · *Arguments:* 2

Verifies that the process is able to access the given scope and reference. If no reference is provided, a global scope is assumed, for instance, ``process.permission.has('fs.read')`` will check if the process has ALL file system read permissions.

In audit mode (``--permission-audit``), this method still returns the actual permission status, but denied operations will not throw ``ERR_ACCESS_DENIED``.

**Arguments**

* ``scope`` (*string*)
* ``reference`` (*string*) — optional

**Example**

.. code-block:: javascript

   // Check if the process has permission to read the README file
   process.permission.has('fs.read', './README.md');
   // Check if the process has read permission operations
   process.permission.has('fs.read');


allowedNodeEnvironmentFlags.intersection
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.intersection(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.intersection(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.isDisjointFrom
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.isDisjointFrom(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.isDisjointFrom(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.isSubsetOf
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.isSubsetOf(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.isSubsetOf(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.isSupersetOf
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.isSupersetOf(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.isSupersetOf(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.keys
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.keys(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.keys();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.keys
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.keys(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.keys();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.symmetricDifference
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.symmetricDifference(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.symmetricDifference(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.union
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.union(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.union(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.values
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.values(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.values();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


allowedNodeEnvironmentFlags.values
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.allowedNodeEnvironmentFlags.values(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.allowedNodeEnvironmentFlags.values();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


report
------

**report** — ``process.report`` is an object whose methods are used to generate diagnostic reports for the current process. Additional documentation is available in the report documentation.

report.getReport
~~~~~~~~~~~~~~~~

.. code-block:: text

   process.report.getReport([err])

*Kind:* function · *Arguments:* 1

Returns a JavaScript Object representation of a diagnostic report for the running process. The report's JavaScript stack trace is taken from ``err``, if present.

Additional documentation is available in the report documentation.

**Arguments**

* ``err`` (*Error*) — optional: A custom error used for reporting the JavaScript stack.

**Example**

.. code-block:: javascript

   const { report } = require('node:process');
   const util = require('node:util');

   const data = report.getReport();
   console.log(data.header.nodejsVersion);

   // Similar to process.report.writeReport()
   const fs = require('node:fs');
   fs.writeFileSync('my-report.log', util.inspect(data), 'utf8');


report.writeReport
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   process.report.writeReport([filename][, err])

*Kind:* function · *Arguments:* 2

Writes a diagnostic report to a file. If ``filename`` is not provided, the default filename includes the date, time, PID, and a sequence number. The report's JavaScript stack trace is taken from ``err``, if present.

If the value of ``filename`` is set to ``'stdout'`` or ``'stderr'``, the report is written to the stdout or stderr of the process respectively.

**Arguments**

* ``filename`` (*string*) — optional: Name of the file where the report is written. This
* ``err`` (*Error*) — optional: A custom error used for reporting the JavaScript stack.

**Example**

.. code-block:: javascript

   const { report } = require('node:process');

   report.writeReport();


stderr
------

**stderr** — The ``process.stderr`` property returns a stream connected to ``stderr`` (fd ``2``). It is a ``net.Socket`` (which is a Duplex stream) unless fd ``2`` refers to a file, in which case it is a Writable stream.

``process.stderr`` differs from other Node.js streams in important ways. See note on process I/O for more information.

stderr.end
~~~~~~~~~~

.. code-block:: text

   process.stderr.end(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stderr.end();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stderr.write
~~~~~~~~~~~~

.. code-block:: text

   process.stderr.write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stderr.write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdin
-----

**stdin** — The ``process.stdin`` property returns a stream connected to ``stdin`` (fd ``0``). It is a ``net.Socket`` (which is a Duplex stream) unless fd ``0`` refers to a file, in which case it is a Readable stream.

For details of how to read from ``stdin`` see ``readable.read()``.

stdin._read
~~~~~~~~~~~

.. code-block:: text

   process.stdin._read(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   console.log(typeof process.stdin._read()); // 'function'
   console.log(process.stdin._read().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdin.destroy
~~~~~~~~~~~~~

.. code-block:: text

   process.stdin.destroy(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdin.destroy(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdin.pause
~~~~~~~~~~~

.. code-block:: text

   process.stdin.pause(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdin.pause();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdin.resume
~~~~~~~~~~~~

.. code-block:: text

   process.stdin.resume(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdin.resume();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdin.setRawMode
~~~~~~~~~~~~~~~~

.. code-block:: text

   process.stdin.setRawMode(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdin.setRawMode(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdout
------

**stdout** — The ``process.stdout`` property returns a stream connected to ``stdout`` (fd ``1``). It is a ``net.Socket`` (which is a Duplex stream) unless fd ``1`` refers to a file, in which case it is a Writable stream.

For example, to copy ``process.stdin`` to ``process.stdout``:

stdout.end
~~~~~~~~~~

.. code-block:: text

   process.stdout.end(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdout.end();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


stdout.write
~~~~~~~~~~~~

.. code-block:: text

   process.stdout.write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const process = require('node:process');

   const result = process.stdout.write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``_exiting`` — boolean
* ``allowedNodeEnvironmentFlags`` — object
* ``allowedNodeEnvironmentFlags.size`` — number
* ``arch`` — string
* ``argv`` — array
* ``argv0`` — string
* ``config`` — object
* ``config.target_defaults`` — object
* ``config.target_defaults.default_configuration`` — string
* ``config.variables`` — object
* ``config.variables.asan`` — number
* ``config.variables.napi_build_version`` — string
* ``config.variables.node_builtin_shareable_builtins`` — array
* ``config.variables.node_module_version`` — number
* ``config.variables.node_shared`` — number
* ``config.variables.node_shared_openssl`` — number
* ``config.variables.node_use_amaro`` — number
* ``config.variables.openssl`` — number
* ``config.variables.v8_enable_i18n_support`` — number
* ``debugPort`` — number
* ``domain`` — null
* ``env`` — object
* ``env._`` — string
* ``env.__MODULES_LMINIT`` — string
* ``env.__MODULES_SHARE_MANPATH`` — string
* ``env.BASH_FUNC__module_raw%`` — string
* ``env.BASH_FUNC_ml%`` — string
* ``env.BASH_FUNC_module%`` — string
* ``env.BASH_FUNC_scl%`` — string
* ``env.BASH_FUNC_which%`` — string
* ``env.BUN_INSTALL`` — string
* ``env.CAELESTIA_LIB_DIR`` — string
* ``env.CAELESTIA_PREFIX`` — string
* ``env.CC`` — string
* ``env.CMAKE_INCLUDE_PATH`` — string
* ``env.CMAKE_LIBRARY_PATH`` — string
* ``env.CMAKE_PREFIX_PATH`` — string
* ``env.CODEBUFF_LAUNCHER_PID`` — string
* ``env.CODEBUFF_LAUNCHER_VERSION`` — string
* ``env.CODEBUFF_RG_PATH`` — string
* ``env.CODEBUFF_TREE_SITTER_WASM_PATH`` — string
* ``env.CXX`` — string
* ``env.DBUS_SESSION_BUS_ADDRESS`` — string
* ``env.DEBUGINFOD_IMA_CERT_PATH`` — string
* ``env.DOCKER_HOST`` — string
* ``env.DOTNET_BUNDLE_EXTRACT_BASE_DIR`` — string
* ``env.DOTNET_ROOT`` — string
* ``env.GODEBUG`` — string
* ``env.HISTCONTROL`` — string
* ``env.HISTSIZE`` — string
* ``env.HOME`` — string
* ``env.HOSTNAME`` — string
* ``env.KDEDIRS`` — string
* ``env.LANG`` — string
* ``env.LD_LIBRARY_PATH`` — string
* ``env.LESSOPEN`` — string
* ``env.LOADEDMODULES`` — string
* ``env.LOGNAME`` — string
* ``env.LS_COLORS`` — string
* ``env.MAIL`` — string
* ``env.MANPATH`` — string
* ``env.MODULEPATH`` — string
* ``env.MODULES_CMD`` — string
* ``env.MODULES_RUN_QUARANTINE`` — string
* ``env.MODULESHOME`` — string
* ``env.MOTD_SHOWN`` — string
* ``env.NVM_BIN`` — string
* ``env.NVM_CD_FLAGS`` — string
* ``env.NVM_DIR`` — string
* ``env.NVM_INC`` — string
* ``env.OLDPWD`` — string
* ``env.PATH`` — string
* ``env.PKG_CONFIG_PATH`` — string
* ``env.PWD`` — string
* ``env.PYTHONPATH`` — string
* ``env.QML_IMPORT_PATH`` — string
* ``env.QT_PLUGIN_PATH`` — string
* ``env.SHELL`` — string
* ``env.SHLVL`` — string
* ``env.SSH_CLIENT`` — string
* ``env.SSH_CONNECTION`` — string
* ``env.SSH_TTY`` — string
* ``env.TERM`` — string
* ``env.USER`` — string
* ``env.which_declare`` — string
* ``env.XDG_CONFIG_DIRS`` — string
* ``env.XDG_DATA_DIRS`` — string
* ``env.XDG_RUNTIME_DIR`` — string
* ``env.XDG_SESSION_CLASS`` — string
* ``env.XDG_SESSION_ID`` — string
* ``env.XDG_SESSION_TYPE`` — string
* ``execArgv`` — array
* ``execPath`` — string
* ``exitCode`` — number
* ``features`` — object
* ``features.cached_builtins`` — boolean
* ``features.debug`` — boolean
* ``features.inspector`` — boolean
* ``features.ipv6`` — boolean
* ``features.openssl_is_boringssl`` — boolean
* ``features.quic`` — boolean
* ``features.require_module`` — boolean
* ``features.tls`` — boolean
* ``features.tls_alpn`` — boolean
* ``features.tls_ocsp`` — boolean
* ``features.tls_sni`` — boolean
* ``features.typescript`` — string
* ``features.uv`` — boolean
* ``pid`` — number
* ``platform`` — string
* ``ppid`` — number
* ``release`` — object
* ``release.headersUrl`` — string
* ``release.libUrl`` — string
* ``release.lts`` — string
* ``release.name`` — string
* ``release.sourceUrl`` — string
* ``report`` — object
* ``report.compact`` — boolean
* ``report.directory`` — string
* ``report.filename`` — string
* ``report.reportOnFatalError`` — boolean
* ``report.reportOnSignal`` — boolean
* ``report.reportOnUncaughtException`` — boolean
* ``report.signal`` — string
* ``stderr`` — object
* ``stderr.fd`` — number
* ``stderr.isTTY`` — boolean
* ``stdin`` — object
* ``stdin._ended`` — boolean
* ``stdin._pending`` — boolean
* ``stdin._poll`` — number
* ``stdin._readableState`` — object
* ``stdin._readableState.autoDestroy`` — boolean
* ``stdin._readableState.buffer`` — array
* ``stdin._readableState.closed`` — boolean
* ``stdin._readableState.closeEmitted`` — boolean
* ``stdin._readableState.constructed`` — boolean
* ``stdin._readableState.decoder`` — null
* ``stdin._readableState.destroyed`` — boolean
* ``stdin._readableState.didRead`` — boolean
* ``stdin._readableState.emitClose`` — boolean
* ``stdin._readableState.emittedReadable`` — boolean
* ``stdin._readableState.encoding`` — null
* ``stdin._readableState.ended`` — boolean
* ``stdin._readableState.endEmitted`` — boolean
* ``stdin._readableState.errored`` — null
* ``stdin._readableState.errorEmitted`` — boolean
* ``stdin._readableState.flowing`` — null
* ``stdin._readableState.highWaterMark`` — number
* ``stdin._readableState.length`` — number
* ``stdin._readableState.needReadable`` — boolean
* ``stdin._readableState.objectMode`` — boolean
* ``stdin._readableState.pipes`` — array
* ``stdin._readableState.readable`` — boolean
* ``stdin._readableState.readableListening`` — boolean
* ``stdin._readableState.reading`` — boolean
* ``stdin._readableState.sync`` — boolean
* ``stdin.fd`` — number
* ``stdin.isTTY`` — boolean
* ``stdout`` — object
* ``stdout.fd`` — number
* ``stdout.isTTY`` — boolean
* ``title`` — string
* ``type`` — string
* ``version`` — string
* ``versions`` — object
* ``versions.acorn`` — string
* ``versions.ada`` — string
* ``versions.ares`` — string
* ``versions.brotli`` — string
* ``versions.cldr`` — string
* ``versions.icu`` — string
* ``versions.llhttp`` — string
* ``versions.merve`` — string
* ``versions.modules`` — string
* ``versions.napi`` — string
* ``versions.nbytes`` — string
* ``versions.ncrypto`` — string
* ``versions.nghttp2`` — string
* ``versions.nghttp3`` — string
* ``versions.ngtcp2`` — string
* ``versions.node`` — string
* ``versions.openssl`` — string
* ``versions.simdjson`` — string
* ``versions.simdutf`` — string
* ``versions.tz`` — string
* ``versions.undici`` — string
* ``versions.unicode`` — string
* ``versions.uv`` — string
* ``versions.uvwasi`` — string
* ``versions.v8`` — string
* ``versions.zlib`` — string
* ``versions.zstd`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
