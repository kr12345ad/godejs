child_process
=============

The ``node:child_process`` module provides the ability to spawn subprocesses in a manner that is similar, but not identical, to popen(3). This capability is primarily provided by the ``child_process.spawn()`` function:

By default, pipes for ``stdin``, ``stdout``, and ``stderr`` are established between the parent Node.js process and the spawned subprocess. These pipes have limited (and platform-specific) capacity. If the subprocess writes to stdout in excess of that limit without the output being captured, the subprocess blocks, waiting for the pipe buffer to accept more data. This is identical to the behavior of pipes in the shell. Use the ``{ stdio: 'ignore' }`` option if the output will not be consumed.

Loading
-------

.. code-block:: javascript

   const child_process = require('node:child_process');
   // ESM
   import * as child_process from 'node:child_process';

Quick start
-----------

.. code-block:: javascript

   const { spawn } = require('node:child_process');
   const ls = spawn('ls', ['-lh', '/usr']);

   ls.stdout.on('data', (data) => {
     console.log(`stdout: ${data}`);
   });

   ls.stderr.on('data', (data) => {
     console.error(`stderr: ${data}`);
   });

   ls.on('close', (code) => {
     console.log(`child process exited with code ${code}`);
   });

API
---

ChildProcess
~~~~~~~~~~~~

.. code-block:: text

   ChildProcess

*Kind:* class · *Arguments:* 0

Instances of the ``ChildProcess`` represent spawned child processes.

Instances of ``ChildProcess`` are not intended to be created directly. Rather, use the ``child_process.spawn()``, ``child_process.exec()``, ``child_process.execFile()``, or ``child_process.fork()`` methods to create instances of ``ChildProcess``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { spawn } = require('node:child_process');
   const ls = spawn('ls', ['-lh', '/usr']);

   ls.stdout.on('data', (data) => {
     console.log(`stdout: ${data}`);
   });

   ls.on('close', (code) => {
     console.log(`child process close all stdio with code ${code}`);
   });

   ls.on('exit', (code) => {
     console.log(`child process exited with code ${code}`);
   });


exec
~~~~

.. code-block:: text

   child_process.exec(command[, options][, callback])

*Kind:* function · *Arguments:* 3

Spawns a shell then executes the ``command`` within that shell, buffering any generated output. The ``command`` string passed to the exec function is processed directly by the shell and special characters (vary based on \`shell <https://en.wikipedia.org/wiki/List_of_command-line_interpreters>\`_) need to be dealt with accordingly:

**Never pass unsanitized user input to this function. Any input containing shell metacharacters may be used to trigger arbitrary command execution.**

**Arguments**

* ``command`` (*string*): The command to run, with space-separated arguments.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*) — optional: called with the output when process terminates.

**Example**

.. code-block:: javascript

   const { exec } = require('node:child_process');

   exec('"/path/to/test file/test.sh" arg1 arg2');
   // Double quotes are used so that the space in the path is not interpreted as
   // a delimiter of multiple arguments.

   exec('echo "The \\$HOME variable is $HOME"');
   // The $HOME variable is escaped in the first instance, but not in the second.


execFile
~~~~~~~~

.. code-block:: text

   child_process.execFile(file[, args][, options][, callback])

*Kind:* function · *Arguments:* 4

The ``child_process.execFile()`` function is similar to ``child_process.exec()`` except that it does not spawn a shell by default. Rather, the specified executable ``file`` is spawned directly as a new process making it slightly more efficient than ``child_process.exec()``.

The same options as ``child_process.exec()`` are supported. Since a shell is not spawned, behaviors such as I/O redirection and file globbing are not supported.

**Arguments**

* ``file`` (*string*): The name or path of the executable file to run.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*) — optional: Called with the output when process terminates.

**Example**

.. code-block:: javascript

   const { execFile } = require('node:child_process');
   const child = execFile('node', ['--version'], (error, stdout, stderr) => {
     if (error) {
       throw error;
     }
     console.log(stdout);
   });


execFileSync
~~~~~~~~~~~~

.. code-block:: text

   child_process.execFileSync(file[, args][, options])

*Kind:* function · *Arguments:* 3

The ``child_process.execFileSync()`` method is generally identical to ``child_process.execFile()`` with the exception that the method will not return until the child process has fully closed. When a timeout has been encountered and ``killSignal`` is sent, the method won't return until the process has completely exited.

If the child process intercepts and handles the ``SIGTERM`` signal and does not exit, the parent process will still wait until the child process has exited.

**Arguments**

* ``file`` (*string*): The name or path of the executable file to run.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { execFileSync } = require('node:child_process');

   try {
     const stdout = execFileSync('my-script.sh', ['my-arg'], {
       // Capture stdout and stderr from child process. Overrides the
       // default behavior of streaming child stderr to the parent stderr
       stdio: 'pipe',

       // Use utf8 encoding for stdio pipes
       encoding: 'utf8',
     });

     console.log(stdout);
   } catch (err) {
     if (err.code) {
       // Spawning child process failed
       console.error(err.code);
     } else {
       // Child was spawned but exited with non-zero exit code
       // Error contains any stdout and stderr from the child
       const { stdout, stderr } = err;

       console.error({ stdout, stderr });
     }
   }


execSync
~~~~~~~~

.. code-block:: text

   child_process.execSync(command[, options])

*Kind:* function · *Arguments:* 2

The ``child_process.execSync()`` method is generally identical to ``child_process.exec()`` with the exception that the method will not return until the child process has fully closed. When a timeout has been encountered and ``killSignal`` is sent, the method won't return until the process has completely exited. If the child process intercepts and handles the ``SIGTERM`` signal and doesn't exit, the parent process will wait until the child process has exited.

If the process times out or has a non-zero exit code, this method will throw. The ``Error`` object will contain the entire result from ``child_process.spawnSync()``.

**Arguments**

* ``command`` (*string*): The command to run.
* ``options`` (*Object*) — optional


fork
~~~~

.. code-block:: text

   child_process.fork(modulePath[, args][, options])

*Kind:* function · *Arguments:* 3

The ``child_process.fork()`` method is a special case of ``child_process.spawn()`` used specifically to spawn new Node.js processes. Like ``child_process.spawn()``, a ``ChildProcess`` object is returned. The returned ``ChildProcess`` will have an additional communication channel built-in that allows messages to be passed back and forth between the parent and child. See ``subprocess.send()`` for details.

Keep in mind that spawned Node.js child processes are independent of the parent with exception of the IPC communication channel that is established between the two. Each process has its own memory, with their own V8 instances. Because of the additional resource allocations required, spawning a large number of child Node.js processes is not recommended.

**Arguments**

* ``modulePath`` (*string|URL*): The module to run in the child.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { fork } = require('node:child_process');

   if (process.argv[2] === 'child') {
     setTimeout(() => {
       console.log(`Hello from ${process.argv[2]}!`);
     }, 1_000);
   } else {
     const controller = new AbortController();
     const { signal } = controller;
     const child = fork(__filename, ['child'], { signal });
     child.on('error', (err) => {
       // This will be called with err being an AbortError if the controller aborts
     });
     controller.abort(); // Stops the child process
   }


spawn
~~~~~

.. code-block:: text

   child_process.spawn(command[, args][, options])

*Kind:* function · *Arguments:* 3

The ``child_process.spawn()`` method spawns a new process using the given ``command``, with command-line arguments in ``args``. If omitted, ``args`` defaults to an empty array.

**If the ``shell`` option is enabled, do not pass unsanitized user input to this function. Any input containing shell metacharacters may be used to trigger arbitrary command execution.**

**Arguments**

* ``command`` (*string*): The command to run.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { spawn } = require('node:child_process');
   const ls = spawn('ls', ['-lh', '/usr']);

   ls.stdout.on('data', (data) => {
     console.log(`stdout: ${data}`);
   });

   ls.stderr.on('data', (data) => {
     console.error(`stderr: ${data}`);
   });

   ls.on('close', (code) => {
     console.log(`child process exited with code ${code}`);
   });


spawnSync
~~~~~~~~~

.. code-block:: text

   child_process.spawnSync(command[, args][, options])

*Kind:* function · *Arguments:* 3

The ``child_process.spawnSync()`` method is generally identical to ``child_process.spawn()`` with the exception that the function will not return until the child process has fully closed. When a timeout has been encountered and ``killSignal`` is sent, the method won't return until the process has completely exited. If the process intercepts and handles the ``SIGTERM`` signal and doesn't exit, the parent process will wait until the child process has exited.

**If the ``shell`` option is enabled, do not pass unsanitized user input to this function. Any input containing shell metacharacters may be used to trigger arbitrary command execution.**

**Arguments**

* ``command`` (*string*): The command to run.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional


ChildProcess class
------------------

**ChildProcess** — Instances of the ``ChildProcess`` represent spawned child processes.

Instances of ``ChildProcess`` are not intended to be created directly. Rather, use the ``child_process.spawn()``, ``child_process.exec()``, ``child_process.execFile()``, or ``child_process.fork()`` methods to create instances of ``ChildProcess``.

ChildProcess.addListener
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.addListener(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.disconnect
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subprocess.disconnect()

*Kind:* function · *Arguments:* 0

Closes the IPC channel between parent and child processes, allowing the child process to exit gracefully once there are no other connections keeping it alive. After calling this method the ``subprocess.connected`` and ``process.connected`` properties in both the parent and child processes (respectively) will be set to ``false``, and it will be no longer possible to pass messages between the processes.

The ``'disconnect'`` event will be emitted when there are no messages in the process of being received. This will most often be triggered immediately after calling ``subprocess.disconnect()``.

This callable takes no arguments.


ChildProcess.emit
~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.emit(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.getMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.getMaxListeners(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.kill
~~~~~~~~~~~~~~~~~

.. code-block:: text

   subprocess.kill([signal])

*Kind:* function · *Arguments:* 1

The ``subprocess.kill()`` method sends a signal to the child process. If no argument is given, the process will be sent the ``'SIGTERM'`` signal. See signal(7) for a list of available signals. This function returns ``true`` if kill(2) succeeds, and ``false`` otherwise.

The ``ChildProcess`` object may emit an ``'error'`` event if the signal cannot be delivered. Sending a signal to a child process that has already exited is not an error but may have unforeseen consequences. Specifically, if the process identifier (PID) has been reassigned to another process, the signal will be delivered to that process instead which can have unexpected results.

**Arguments**

* ``signal`` (*number|string*) — optional

**Example**

.. code-block:: javascript

   const { spawn } = require('node:child_process');
   const grep = spawn('grep', ['ssh']);

   grep.on('close', (code, signal) => {
     console.log(
       `child process terminated due to receipt of signal ${signal}`);
   });

   // Send SIGHUP to process.
   grep.kill('SIGHUP');


ChildProcess.listenerCount
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.listenerCount(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.off
~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.off(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.on
~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.on(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.once
~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.once(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.ref
~~~~~~~~~~~~~~~~

.. code-block:: text

   subprocess.channel.ref()

*Kind:* function · *Arguments:* 0

This method makes the IPC channel keep the event loop of the parent process running if ``.unref()`` has been called before.

This callable takes no arguments.


ChildProcess.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.removeAllListeners(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.removeListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.removeListener(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.send
~~~~~~~~~~~~~~~~~

.. code-block:: text

   subprocess.send(message[, sendHandle[, options]][, callback])

*Kind:* function · *Arguments:* 4

When an IPC channel has been established between the parent and child processes ( i.e. when using ``child_process.fork()``), the ``subprocess.send()`` method can be used to send messages to the child process. When the child process is a Node.js instance, these messages can be received via the ``'message'`` event.

The message goes through serialization and parsing. The resulting message might not be the same as what is originally sent.

**Arguments**

* ``message`` (*Object*)
* ``sendHandle`` (*Handle|undefined*) — optional: ``undefined``, or a ``net.Socket``,
* ``options`` (*Object*) — optional: The ``options`` argument, if present, is an object used to
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const { fork } = require('node:child_process');
   const forkedProcess = fork(`${__dirname}/sub.js`);

   forkedProcess.on('message', (message) => {
     console.log('PARENT got message:', message);
   });

   // Causes the child to print: CHILD got message: { hello: 'world' }
   forkedProcess.send({ hello: 'world' });


ChildProcess.setMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.ChildProcess.setMaxListeners(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ChildProcess.spawn
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   child_process.spawn(command[, args][, options])

*Kind:* function · *Arguments:* 3

The ``child_process.spawn()`` method spawns a new process using the given ``command``, with command-line arguments in ``args``. If omitted, ``args`` defaults to an empty array.

**If the ``shell`` option is enabled, do not pass unsanitized user input to this function. Any input containing shell metacharacters may be used to trigger arbitrary command execution.**

**Arguments**

* ``command`` (*string*): The command to run.
* ``args`` (*string\[]*) — optional: List of string arguments.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { spawn } = require('node:child_process');
   const ls = spawn('ls', ['-lh', '/usr']);

   ls.stdout.on('data', (data) => {
     console.log(`stdout: ${data}`);
   });

   ls.stderr.on('data', (data) => {
     console.error(`stderr: ${data}`);
   });

   ls.on('close', (code) => {
     console.log(`child process exited with code ${code}`);
   });


ChildProcess.unref
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subprocess.channel.unref()

*Kind:* function · *Arguments:* 0

This method makes the IPC channel not keep the event loop of the parent process running, and lets it finish even while the channel is open.

This callable takes no arguments.


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
