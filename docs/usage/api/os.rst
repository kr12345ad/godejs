os
===

The ``node:os`` module provides operating system-related utility methods and properties. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const os = require('node:os');
   // ESM
   import * as os from 'node:os';

Quick start
-----------

.. code-block:: javascript

   const os = require('node:os');

API
---

arch
~~~~

.. code-block:: text

   os.arch()

*Kind:* function · *Arguments:* 0

Returns the operating system CPU architecture for which the Node.js binary was compiled. Possible values are ``'arm'``, ``'arm64'``, ``'ia32'``, ``'loong64'``, ``'mips'``, ``'mipsel'``, ``'ppc64'``, ``'riscv64'``, ``'s390x'``, and ``'x64'``.

The return value is equivalent to ``process.arch``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.arch();
   console.log(result);


availableParallelism
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   os.availableParallelism()

*Kind:* function · *Arguments:* 0

Returns an estimate of the default amount of parallelism a program should use. Always returns a value greater than zero.

This function is a small wrapper about libuv's ``uv_available_parallelism()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.availableParallelism();
   console.log(result);


cpus
~~~~

.. code-block:: text

   os.cpus()

*Kind:* function · *Arguments:* 0

Returns an array of objects containing information about each logical CPU core. The array will be empty if no CPU information is available, such as if the ``/proc`` file system is unavailable.

The properties included on each object include:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.cpus();
   console.log(result);


endianness
~~~~~~~~~~

.. code-block:: text

   os.endianness()

*Kind:* function · *Arguments:* 0

Returns a string identifying the endianness of the CPU for which the Node.js binary was compiled.

Possible values are ``'BE'`` for big endian and ``'LE'`` for little endian.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.endianness();
   console.log(result);


freemem
~~~~~~~

.. code-block:: text

   os.freemem()

*Kind:* function · *Arguments:* 0

Returns the amount of free system memory in bytes as an integer.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.freemem();
   console.log(result);


getPriority
~~~~~~~~~~~

.. code-block:: text

   os.getPriority([pid])

*Kind:* function · *Arguments:* 1

Returns the scheduling priority for the process specified by ``pid``. If ``pid`` is not provided or is ``0``, the priority of the current process is returned.

**Arguments**

* ``pid`` (*integer*) — optional: The process ID to retrieve scheduling priority for.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.getPriority(1);
   console.log(result);


homedir
~~~~~~~

.. code-block:: text

   os.homedir()

*Kind:* function · *Arguments:* 0

Returns the string path of the current user's home directory.

On POSIX, it uses the ``$HOME`` environment variable if defined. Otherwise it uses the effective UID to look up the user's home directory.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.homedir();
   console.log(result);


hostname
~~~~~~~~

.. code-block:: text

   os.hostname()

*Kind:* function · *Arguments:* 0

Returns the host name of the operating system as a string.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.hostname();
   console.log(result);


loadavg
~~~~~~~

.. code-block:: text

   os.loadavg()

*Kind:* function · *Arguments:* 0

Returns an array containing the 1, 5, and 15 minute load averages.

The load average is a measure of system activity calculated by the operating system and expressed as a fractional number.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.loadavg();
   console.log(result);


machine
~~~~~~~

.. code-block:: text

   os.machine()

*Kind:* function · *Arguments:* 0

Returns the machine type as a string, such as ``arm``, ``arm64``, ``aarch64``, ``mips``, ``mips64``, ``ppc64``, ``ppc64le``, ``s390x``, ``i386``, ``i686``, ``x86_64``.

On POSIX systems, the machine type is determined by calling ``uname(3)``. On Windows, ``RtlGetVersion()`` is used, and if it is not available, ``GetVersionExW()`` will be used. See <https://en.wikipedia.org/wiki/Uname#Examples> for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.machine();
   console.log(result);


networkInterfaces
~~~~~~~~~~~~~~~~~

.. code-block:: text

   os.networkInterfaces()

*Kind:* function · *Arguments:* 0

Returns an object containing network interfaces that have been assigned a network address.

Each key on the returned object identifies a network interface. The associated value is an array of objects that each describe an assigned network address.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.networkInterfaces();
   console.log(result);


platform
~~~~~~~~

.. code-block:: text

   os.platform()

*Kind:* function · *Arguments:* 0

Returns a string identifying the operating system platform for which the Node.js binary was compiled. The value is set at compile time. Possible values are ``'aix'``, ``'darwin'``, ``'freebsd'``,``'linux'``, ``'openbsd'``, ``'sunos'``, and ``'win32'``.

The return value is equivalent to ``process.platform``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.platform();
   console.log(result);


release
~~~~~~~

.. code-block:: text

   os.release()

*Kind:* function · *Arguments:* 0

Returns the operating system as a string.

On POSIX systems, the operating system release is determined by calling ``uname(3)``. On Windows, ``GetVersionExW()`` is used. See <https://en.wikipedia.org/wiki/Uname#Examples> for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.release();
   console.log(result);


setPriority
~~~~~~~~~~~

.. code-block:: text

   os.setPriority([pid, ]priority)

*Kind:* function · *Arguments:* 2

Attempts to set the scheduling priority for the process specified by ``pid``. If ``pid`` is not provided or is ``0``, the process ID of the current process is used.

The ``priority`` input must be an integer between ``-20`` (high priority) and ``19`` (low priority). Due to differences between Unix priority levels and Windows priority classes, ``priority`` is mapped to one of six priority constants in ``os.constants.priority``. When retrieving a process priority level, this range mapping may cause the return value to be slightly different on Windows. To avoid confusion, set ``priority`` to one of the priority constants.

**Arguments**

* ``pid`` (*integer*) — optional: The process ID to set scheduling priority for.
* ``priority`` (*integer*) — optional: The scheduling priority to assign to the process.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.setPriority(1, 1);
   console.log(result);


tmpdir
~~~~~~

.. code-block:: text

   os.tmpdir()

*Kind:* function · *Arguments:* 0

Returns the operating system's default directory for temporary files as a string.

On Windows, the result can be overridden by ``TEMP`` and ``TMP`` environment variables, and ``TEMP`` takes precedence over ``TMP``. If neither is set, it defaults to ``%SystemRoot%\temp`` or ``%windir%\temp``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.tmpdir();
   console.log(result);


totalmem
~~~~~~~~

.. code-block:: text

   os.totalmem()

*Kind:* function · *Arguments:* 0

Returns the total amount of system memory in bytes as an integer.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.totalmem();
   console.log(result);


type
~~~~

.. code-block:: text

   os.type()

*Kind:* function · *Arguments:* 0

Returns the operating system name as returned by ``uname(3)``. For example, it returns ``'Linux'`` on Linux, ``'Darwin'`` on macOS, and ``'Windows_NT'`` on Windows.

See <https://en.wikipedia.org/wiki/Uname#Examples> for additional information about the output of running ``uname(3)`` on various operating systems.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.type();
   console.log(result);


uptime
~~~~~~

.. code-block:: text

   os.uptime()

*Kind:* function · *Arguments:* 0

Returns the system uptime in number of seconds.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.uptime();
   console.log(result);


userInfo
~~~~~~~~

.. code-block:: text

   os.userInfo([options])

*Kind:* function · *Arguments:* 1

Returns information about the currently effective user. On POSIX platforms, this is typically a subset of the password file. The returned object includes the ``username``, ``uid``, ``gid``, ``shell``, and ``homedir``. On Windows, the ``uid`` and ``gid`` fields are ``-1``, and ``shell`` is ``null``.

The value of ``homedir`` returned by ``os.userInfo()`` is provided by the operating system. This differs from the result of ``os.homedir()``, which queries environment variables for the home directory before falling back to the operating system response.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.userInfo({});
   console.log(result);


version
~~~~~~~

.. code-block:: text

   os.version()

*Kind:* function · *Arguments:* 0

Returns a string identifying the kernel version.

On POSIX systems, the operating system release is determined by calling ``uname(3)``. On Windows, ``RtlGetVersion()`` is used, and if it is not available, ``GetVersionExW()`` will be used. See <https://en.wikipedia.org/wiki/Uname#Examples> for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const os = require('node:os');

   const result = os.version();
   console.log(result);


Properties and constants
------------------------

* ``constants`` — object
* ``constants.dlopen`` — object
* ``constants.dlopen.RTLD_GLOBAL`` — number
* ``constants.dlopen.RTLD_LAZY`` — number
* ``constants.dlopen.RTLD_LOCAL`` — number
* ``constants.dlopen.RTLD_NOW`` — number
* ``constants.errno`` — object
* ``constants.errno.E2BIG`` — number
* ``constants.errno.EACCES`` — number
* ``constants.errno.EAGAIN`` — number
* ``constants.errno.EBADF`` — number
* ``constants.errno.EBUSY`` — number
* ``constants.errno.ECHILD`` — number
* ``constants.errno.EDOM`` — number
* ``constants.errno.EEXIST`` — number
* ``constants.errno.EFAULT`` — number
* ``constants.errno.EFBIG`` — number
* ``constants.errno.EINTR`` — number
* ``constants.errno.EINVAL`` — number
* ``constants.errno.EIO`` — number
* ``constants.errno.EISDIR`` — number
* ``constants.errno.EMFILE`` — number
* ``constants.errno.EMLINK`` — number
* ``constants.errno.ENFILE`` — number
* ``constants.errno.ENODEV`` — number
* ``constants.errno.ENOENT`` — number
* ``constants.errno.ENOEXEC`` — number
* ``constants.errno.ENOMEM`` — number
* ``constants.errno.ENOSPC`` — number
* ``constants.errno.ENOTDIR`` — number
* ``constants.errno.ENOTTY`` — number
* ``constants.errno.ENXIO`` — number
* ``constants.errno.EPERM`` — number
* ``constants.errno.EPIPE`` — number
* ``constants.errno.ERANGE`` — number
* ``constants.errno.EROFS`` — number
* ``constants.errno.ESPIPE`` — number
* ``constants.errno.ESRCH`` — number
* ``constants.errno.EXDEV`` — number
* ``constants.libuv`` — object
* ``constants.priority`` — object
* ``constants.priority.PRIORITY_ABOVE_NORMAL`` — number
* ``constants.priority.PRIORITY_BELOW_NORMAL`` — number
* ``constants.priority.PRIORITY_HIGH`` — number
* ``constants.priority.PRIORITY_HIGHEST`` — number
* ``constants.priority.PRIORITY_LOW`` — number
* ``constants.priority.PRIORITY_NORMAL`` — number
* ``constants.signals`` — object
* ``constants.signals.SIGABRT`` — number
* ``constants.signals.SIGALRM`` — number
* ``constants.signals.SIGBUS`` — number
* ``constants.signals.SIGCHLD`` — number
* ``constants.signals.SIGCONT`` — number
* ``constants.signals.SIGFPE`` — number
* ``constants.signals.SIGHUP`` — number
* ``constants.signals.SIGILL`` — number
* ``constants.signals.SIGINT`` — number
* ``constants.signals.SIGIO`` — number
* ``constants.signals.SIGKILL`` — number
* ``constants.signals.SIGPIPE`` — number
* ``constants.signals.SIGPROF`` — number
* ``constants.signals.SIGPWR`` — number
* ``constants.signals.SIGQUIT`` — number
* ``constants.signals.SIGSEGV`` — number
* ``constants.signals.SIGSTKFLT`` — number
* ``constants.signals.SIGSTOP`` — number
* ``constants.signals.SIGSYS`` — number
* ``constants.signals.SIGTERM`` — number
* ``constants.signals.SIGTRAP`` — number
* ``constants.signals.SIGTSTP`` — number
* ``constants.signals.SIGTTIN`` — number
* ``constants.signals.SIGTTOU`` — number
* ``constants.signals.SIGURG`` — number
* ``constants.signals.SIGUSR1`` — number
* ``constants.signals.SIGUSR2`` — number
* ``constants.signals.SIGVTALRM`` — number
* ``constants.signals.SIGWINCH`` — number
* ``constants.signals.SIGXCPU`` — number
* ``constants.signals.SIGXFSZ`` — number
* ``constants.uv`` — object
* ``constants.uv.UV_E2BIG`` — number
* ``constants.uv.UV_EACCES`` — number
* ``constants.uv.UV_EAGAIN`` — number
* ``constants.uv.UV_EBADF`` — number
* ``constants.uv.UV_EBUSY`` — number
* ``constants.uv.UV_ECHILD`` — number
* ``constants.uv.UV_EDOM`` — number
* ``constants.uv.UV_EEXIST`` — number
* ``constants.uv.UV_EFAULT`` — number
* ``constants.uv.UV_EFBIG`` — number
* ``constants.uv.UV_EINTR`` — number
* ``constants.uv.UV_EINVAL`` — number
* ``constants.uv.UV_EIO`` — number
* ``constants.uv.UV_EISDIR`` — number
* ``constants.uv.UV_EMFILE`` — number
* ``constants.uv.UV_EMLINK`` — number
* ``constants.uv.UV_ENFILE`` — number
* ``constants.uv.UV_ENODEV`` — number
* ``constants.uv.UV_ENOENT`` — number
* ``constants.uv.UV_ENOEXEC`` — number
* ``constants.uv.UV_ENOMEM`` — number
* ``constants.uv.UV_ENOSPC`` — number
* ``constants.uv.UV_ENOTDIR`` — number
* ``constants.uv.UV_ENOTTY`` — number
* ``constants.uv.UV_ENXIO`` — number
* ``constants.uv.UV_EOF`` — number
* ``constants.uv.UV_EPERM`` — number
* ``constants.uv.UV_EPIPE`` — number
* ``constants.uv.UV_ERANGE`` — number
* ``constants.uv.UV_EROFS`` — number
* ``constants.uv.UV_ESPIPE`` — number
* ``constants.uv.UV_ESRCH`` — number
* ``constants.uv.UV_EXDEV`` — number
* ``devNull`` — string
* ``EOL`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
