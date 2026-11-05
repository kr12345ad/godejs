cluster
=======

Clusters of Node.js processes can be used to run multiple instances of Node.js that can distribute workloads among their application threads. When process isolation is not needed, use the ``worker_threads`` module instead, which allows running multiple application threads within a single Node.js instance.

The cluster module allows easy creation of child processes that all share server ports.

Loading
-------

.. code-block:: javascript

   const cluster = require('node:cluster');
   // ESM
   import * as cluster from 'node:cluster';

Quick start
-----------

.. code-block:: javascript

   const cluster = require('node:cluster');
   const http = require('node:http');
   const numCPUs = require('node:os').availableParallelism();

   if (cluster.isPrimary) {
     console.log(`Primary ${process.pid} is running`);

     // Fork workers.
     for (let i = 0; i < numCPUs; i++) {
       cluster.fork();
     }

     cluster.on('exit', (worker, code, signal) => {
       console.log(`worker ${worker.process.pid} died`);
     });
   } else {
     // Workers can share any TCP connection
     // In this case it is an HTTP server
     http.createServer((req, res) => {
       res.writeHead(200);
       res.end('hello world\n');
     }).listen(8000);

     console.log(`Worker ${process.pid} started`);
   }

API
---

disconnect
~~~~~~~~~~

.. code-block:: text

   cluster.disconnect([callback])

*Kind:* function · *Arguments:* 1

Calls ``.disconnect()`` on each worker in ``cluster.workers``.

When they are disconnected all internal handles will be closed, allowing the primary process to die gracefully if no other event is waiting.

**Arguments**

* ``callback`` (*Function*) — optional: Called when all workers are disconnected and handles are

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   const result = cluster.disconnect(() => {});
   console.log(result);


fork
~~~~

.. code-block:: text

   cluster.fork([env])

*Kind:* function · *Arguments:* 1

Spawn a new worker process.

This can only be called from the primary process.

**Arguments**

* ``env`` (*Object*) — optional: Key/value pairs to add to worker process environment.

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   const result = cluster.fork({});
   console.log(result);


setupMaster
~~~~~~~~~~~

.. code-block:: text

   cluster.setupMaster([settings])

*Kind:* function · *Arguments:* 1

Deprecated alias for ``.setupPrimary()``.

**Arguments**

* ``settings`` — optional

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   const result = cluster.setupMaster(undefined);
   console.log(result);


setupPrimary
~~~~~~~~~~~~

.. code-block:: text

   cluster.setupPrimary([settings])

*Kind:* function · *Arguments:* 1

``setupPrimary`` is used to change the default 'fork' behavior. Once called, the settings will be present in ``cluster.settings``.

Any settings changes only affect future calls to ``.fork()`` and have no effect on workers that are already running.

**Arguments**

* ``settings`` (*Object*) — optional: See ``cluster.settings``.

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   cluster.setupPrimary({
     exec: 'worker.js',
     args: ['--use', 'https'],
     silent: true,
   });
   cluster.fork(); // https worker
   cluster.setupPrimary({
     exec: 'worker.js',
     args: ['--use', 'http'],
   });
   cluster.fork(); // http worker


Worker
~~~~~~

.. code-block:: text

   Worker

*Kind:* class · *Arguments:* 2

A ``Worker`` object contains all public information and method about a worker. In the primary it can be obtained using ``cluster.workers``. In a worker it can be obtained using ``cluster.worker``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   if (cluster.isPrimary) {
     const worker = cluster.fork();
     worker.on('exit', (code, signal) => {
       if (signal) {
         console.log(`worker was killed by signal: ${signal}`);
       } else if (code !== 0) {
         console.log(`worker exited with error code: ${code}`);
       } else {
         console.log('worker success!');
       }
     });
   }


Worker class
------------

**Worker** — A ``Worker`` object contains all public information and method about a worker. In the primary it can be obtained using ``cluster.workers``. In a worker it can be obtained using ``cluster.worker``.

Worker.disconnect
~~~~~~~~~~~~~~~~~

.. code-block:: text

   worker.disconnect()

*Kind:* function · *Arguments:* 0

In a worker, this function will close all servers, wait for the ``'close'`` event on those servers, and then disconnect the IPC channel.

In the primary, an internal message is sent to the worker causing it to call ``.disconnect()`` on itself.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   if (cluster.isPrimary) {
     const worker = cluster.fork();
     let timeout;

     worker.on('listening', (address) => {
       worker.send('shutdown');
       worker.disconnect();
       timeout = setTimeout(() => {
         worker.kill();
       }, 2000);
     });

     worker.on('disconnect', () => {
       clearTimeout(timeout);
     });

   } else if (cluster.isWorker) {
     const net = require('node:net');
     const server = net.createServer((socket) => {
       // Connections never end
     });

     server.listen(8000);

     process.on('message', (msg) => {
       if (msg === 'shutdown') {
         // Initiate graceful close of any connections to server
       }
     });
   }


Worker.kill
~~~~~~~~~~~

.. code-block:: text

   worker.kill([signal])

*Kind:* function · *Arguments:* 1

This function will kill the worker. In the primary worker, it does this by disconnecting the ``worker.process``, and once disconnected, killing with ``signal``. In the worker, it does it by killing the process with ``signal``.

The ``kill()`` function kills the worker process without waiting for a graceful disconnect, it has the same behavior as ``worker.process.kill()``.

**Arguments**

* ``signal`` (*string*) — optional: Name of the kill signal to send to the worker

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');

   const instance = new cluster.Worker();
   const result = instance.kill(new AbortController().signal);
   console.log(result);


Worker.send
~~~~~~~~~~~

.. code-block:: text

   worker.send(message[, sendHandle[, options]][, callback])

*Kind:* function · *Arguments:* 4

Send a message to a worker or primary, optionally with a handle.

In the primary, this sends a message to a specific worker. It is identical to ``ChildProcess.send()``.

**Arguments**

* ``message`` (*Object*)
* ``sendHandle`` (*Handle*) — optional
* ``options`` (*Object*) — optional: The ``options`` argument, if present, is an object used to
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   if (cluster.isPrimary) {
     const worker = cluster.fork();
     worker.send('hi there');

   } else if (cluster.isWorker) {
     process.on('message', (msg) => {
       process.send(msg);
     });
   }


Properties and constants
------------------------

* ``isMaster`` — boolean
* ``isPrimary`` — boolean
* ``isWorker`` — boolean
* ``SCHED_NONE`` — number
* ``SCHED_RR`` — number
* ``schedulingPolicy`` — number
* ``settings`` — object
* ``settings.args`` — array
* ``settings.cwd`` — undefined
* ``settings.exec`` — string
* ``settings.execArgv`` — array
* ``settings.gid`` — undefined
* ``settings.inspectPort`` — undefined
* ``settings.serialization`` — string
* ``settings.silent`` — boolean
* ``settings.stdio`` — string
* ``settings.uid`` — undefined
* ``settings.windowsHide`` — boolean
* ``worker`` — null
* ``workers`` — object

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
