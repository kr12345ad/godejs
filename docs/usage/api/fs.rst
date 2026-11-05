fs
===

The ``node:fs`` module enables interacting with the file system in a way modeled on standard POSIX functions.

To use the promise-based APIs:

Loading
-------

.. code-block:: javascript

   const fs = require('node:fs');
   // ESM
   import * as fs from 'node:fs';

Quick start
-----------

.. code-block:: javascript

   const fs = require('node:fs/promises');

API
---

_toUnixTimestamp
~~~~~~~~~~~~~~~~

.. code-block:: text

   fs._toUnixTimestamp(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


access
~~~~~~

.. code-block:: text

   fs.access(path[, mode], callback)

*Kind:* function · *Arguments:* 3

Tests a user's permissions for the file or directory specified by ``path``. The ``mode`` argument is an optional integer that specifies the accessibility checks to be performed. ``mode`` should be either the value ``fs.constants.F_OK`` or a mask consisting of the bitwise OR of any of ``fs.constants.R_OK``, ``fs.constants.W_OK``, and ``fs.constants.X_OK`` (e.g. ``fs.constants.W_OK | fs.constants.R_OK``). Check File access constants for possible values of ``mode``.

The final argument, ``callback``, is a callback function that is invoked with a possible error argument. If any of the accessibility checks fail, the error argument will be an ``Error`` object. The following examples check if ``package.json`` exists, and if it is readable or writable.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*integer*) — optional: **Default:** ``fs.constants.F_OK``
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { access, constants } from 'node:fs';

   const file = 'package.json';

   // Check if the file exists in the current directory.
   access(file, constants.F_OK, (err) => {
     console.log(`${file} ${err ? 'does not exist' : 'exists'}`);
   });

   // Check if the file is readable.
   access(file, constants.R_OK, (err) => {
     console.log(`${file} ${err ? 'is not readable' : 'is readable'}`);
   });

   // Check if the file is writable.
   access(file, constants.W_OK, (err) => {
     console.log(`${file} ${err ? 'is not writable' : 'is writable'}`);
   });

   // Check if the file is readable and writable.
   access(file, constants.R_OK | constants.W_OK, (err) => {
     console.log(`${file} ${err ? 'is not' : 'is'} readable and writable`);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsaccesspath-mode-callback>`__

accessSync
~~~~~~~~~~

.. code-block:: text

   fs.accessSync(path[, mode])

*Kind:* function · *Arguments:* 2

Synchronously tests a user's permissions for the file or directory specified by ``path``. The ``mode`` argument is an optional integer that specifies the accessibility checks to be performed. ``mode`` should be either the value ``fs.constants.F_OK`` or a mask consisting of the bitwise OR of any of ``fs.constants.R_OK``, ``fs.constants.W_OK``, and ``fs.constants.X_OK`` (e.g. ``fs.constants.W_OK | fs.constants.R_OK``). Check File access constants for possible values of ``mode``.

If any of the accessibility checks fail, an ``Error`` will be thrown. Otherwise, the method will return ``undefined``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*integer*) — optional: **Default:** ``fs.constants.F_OK``

**Example**

.. code-block:: javascript

   import { accessSync, constants } from 'node:fs';

   try {
     accessSync('etc/passwd', constants.R_OK | constants.W_OK);
     console.log('can read/write');
   } catch (err) {
     console.error('no access!');
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsaccesssyncpath-mode>`__

appendFile
~~~~~~~~~~

.. code-block:: text

   fs.appendFile(path, data[, options], callback)

*Kind:* function · *Arguments:* 4

Asynchronously append data to a file, creating the file if it does not yet exist. ``data`` can be a string or a ``Buffer``.

The ``mode`` option only affects the newly created file. See ``fs.open()`` for more details.

**Arguments**

* ``path`` (*string|Buffer|URL|number*): filename or file descriptor
* ``data`` (*string|Buffer*)
* ``options`` (*Object|string*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { appendFile } from 'node:fs';

   appendFile('message.txt', 'data to append', (err) => {
     if (err) throw err;
     console.log('The "data to append" was appended to file!');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsappendfilepath-data-options-callback>`__

appendFileSync
~~~~~~~~~~~~~~

.. code-block:: text

   fs.appendFileSync(path, data[, options])

*Kind:* function · *Arguments:* 3

Synchronously append data to a file, creating the file if it does not yet exist. ``data`` can be a string or a ``Buffer``.

The ``mode`` option only affects the newly created file. See ``fs.open()`` for more details.

**Arguments**

* ``path`` (*string|Buffer|URL|number*): filename or file descriptor
* ``data`` (*string|Buffer*)
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   import { appendFileSync } from 'node:fs';

   try {
     appendFileSync('message.txt', 'data to append');
     console.log('The "data to append" was appended to file!');
   } catch (err) {
     /* Handle the error */
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsappendfilesyncpath-data-options>`__

chmod
~~~~~

.. code-block:: text

   fs.chmod(path, mode, callback)

*Kind:* function · *Arguments:* 3

Asynchronously changes the permissions of a file. No arguments other than a possible exception are given to the completion callback.

See the POSIX chmod(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*string|integer*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { chmod } from 'node:fs';

   chmod('my_file.txt', 0o775, (err) => {
     if (err) throw err;
     console.log('The permissions for file "my_file.txt" have been changed!');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fschmodpath-mode-callback>`__

chmodSync
~~~~~~~~~

.. code-block:: text

   fs.chmodSync(path, mode)

*Kind:* function · *Arguments:* 2

For detailed information, see the documentation of the asynchronous version of this API: ``fs.chmod()``.

See the POSIX chmod(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*string|integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fschmodsyncpath-mode>`__

chown
~~~~~

.. code-block:: text

   fs.chown(path, uid, gid, callback)

*Kind:* function · *Arguments:* 4

Asynchronously changes owner and group of a file. No arguments other than a possible exception are given to the completion callback.

See the POSIX chown(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``uid`` (*integer*)
* ``gid`` (*integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fschownpath-uid-gid-callback>`__

chownSync
~~~~~~~~~

.. code-block:: text

   fs.chownSync(path, uid, gid)

*Kind:* function · *Arguments:* 3

Synchronously changes owner and group of a file. Returns ``undefined``. This is the synchronous version of ``fs.chown()``.

See the POSIX chown(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``uid`` (*integer*)
* ``gid`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fschownsyncpath-uid-gid>`__

close
~~~~~

.. code-block:: text

   fs.close(fd[, callback])

*Kind:* function · *Arguments:* 2

Closes the file descriptor. No arguments other than a possible exception are given to the completion callback.

Calling ``fs.close()`` on any file descriptor (``fd``) that is currently in use through any other ``fs`` operation may lead to undefined behavior.

**Arguments**

* ``fd`` (*integer*)
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsclosefd-callback>`__

closeSync
~~~~~~~~~

.. code-block:: text

   fs.closeSync(fd)

*Kind:* function · *Arguments:* 1

Closes the file descriptor. Returns ``undefined``.

Calling ``fs.closeSync()`` on any file descriptor (``fd``) that is currently in use through any other ``fs`` operation may lead to undefined behavior.

**Arguments**

* ``fd`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsclosesyncfd>`__

copyFile
~~~~~~~~

.. code-block:: text

   fs.copyFile(src, dest[, mode], callback)

*Kind:* function · *Arguments:* 4

Asynchronously copies ``src`` to ``dest``. By default, ``dest`` is overwritten if it already exists. No arguments other than a possible exception are given to the callback function. Node.js makes no guarantees about the atomicity of the copy operation. If an error occurs after the destination file has been opened for writing, Node.js will attempt to remove the destination.

Symbolic links are followed. If ``src`` is a symbolic link, the target file is copied. If ``dest`` is a symbolic link, the target file is overwritten unless ``mode`` contains ``fs.constants.COPYFILE_EXCL``.

**Arguments**

* ``src`` (*string|Buffer|URL*): source filename to copy
* ``dest`` (*string|Buffer|URL*): destination filename of the copy operation
* ``mode`` (*integer*) — optional: modifiers for copy operation. **Default:** ``0``.
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { copyFile, constants } from 'node:fs';

   function callback(err) {
     if (err) throw err;
     console.log('source.txt was copied to destination.txt');
   }

   // destination.txt will be created or overwritten by default.
   copyFile('source.txt', 'destination.txt', callback);

   // By using COPYFILE_EXCL, the operation will fail if destination.txt exists.
   copyFile('source.txt', 'destination.txt', constants.COPYFILE_EXCL, callback);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscopyfilesrc-dest-mode-callback>`__

copyFileSync
~~~~~~~~~~~~

.. code-block:: text

   fs.copyFileSync(src, dest[, mode])

*Kind:* function · *Arguments:* 3

Synchronously copies ``src`` to ``dest``. By default, ``dest`` is overwritten if it already exists. Returns ``undefined``. Node.js makes no guarantees about the atomicity of the copy operation. If an error occurs after the destination file has been opened for writing, Node.js will attempt to remove the destination.

Symbolic links are followed. If ``src`` is a symbolic link, the target file is copied. If ``dest`` is a symbolic link, the target file is overwritten unless ``mode`` contains ``fs.constants.COPYFILE_EXCL``.

**Arguments**

* ``src`` (*string|Buffer|URL*): source filename to copy
* ``dest`` (*string|Buffer|URL*): destination filename of the copy operation
* ``mode`` (*integer*) — optional: modifiers for copy operation. **Default:** ``0``.

**Example**

.. code-block:: javascript

   import { copyFileSync, constants } from 'node:fs';

   // destination.txt will be created or overwritten by default.
   copyFileSync('source.txt', 'destination.txt');
   console.log('source.txt was copied to destination.txt');

   // By using COPYFILE_EXCL, the operation will fail if destination.txt exists.
   copyFileSync('source.txt', 'destination.txt', constants.COPYFILE_EXCL);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscopyfilesyncsrc-dest-mode>`__

cp
~~~

.. code-block:: text

   fs.cp(src, dest[, options], callback)

*Kind:* function · *Arguments:* 4

Asynchronously copies the entire directory structure from ``src`` to ``dest``, including subdirectories and files.

When copying a directory to another directory, globs are not supported and behavior is similar to ``cp dir1/ dir2/``.

**Arguments**

* ``src`` (*string|URL*): source path to copy.
* ``dest`` (*string|URL*): destination path to copy to.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscpsrc-dest-options-callback>`__

cpSync
~~~~~~

.. code-block:: text

   fs.cpSync(src, dest[, options])

*Kind:* function · *Arguments:* 3

Synchronously copies the entire directory structure from ``src`` to ``dest``, including subdirectories and files.

When copying a directory to another directory, globs are not supported and behavior is similar to ``cp dir1/ dir2/``.

**Arguments**

* ``src`` (*string|URL*): source path to copy.
* ``dest`` (*string|URL*): destination path to copy to.
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscpsyncsrc-dest-options>`__

createReadStream
~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.createReadStream(path[, options])

*Kind:* function · *Arguments:* 2

``options`` can include ``start`` and ``end`` values to read a range of bytes from the file instead of the entire file. Both ``start`` and ``end`` are inclusive and start counting at 0, allowed values are in the \0, [``Number.MAX_SAFE_INTEGER``] range. If ``fd`` is specified and ``start`` is omitted or ``undefined``, ``fs.createReadStream()`` reads sequentially from the current file position. The ``encoding`` can be any one of those accepted by ``Buffer``.

If ``fd`` is specified, ``ReadStream`` will ignore the ``path`` argument and will use the specified file descriptor. This means that no ``'open'`` event will be emitted. ``fd`` should be blocking; non-blocking ``fd``\ s should be passed to ``net.Socket``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

**Example**

.. code-block:: javascript

   import { createReadStream } from 'node:fs';

   // Create a stream from some character device.
   const stream = createReadStream('/dev/input/event0');
   setTimeout(() => {
     stream.close(); // This may not close the stream.
     // Artificially marking end-of-stream, as if the underlying resource had
     // indicated end-of-file by itself, allows the stream to close.
     // This does not cancel pending read operations, and if there is such an
     // operation, the process may still not be able to exit successfully
     // until it finishes.
     stream.push(null);
     stream.read(0);
   }, 100);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscreatereadstreampath-options>`__

createWriteStream
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.createWriteStream(path[, options])

*Kind:* function · *Arguments:* 2

``options`` may also include a ``start`` option to allow writing data at some position past the beginning of the file, allowed values are in the \0, [``Number.MAX_SAFE_INTEGER``] range. Modifying a file rather than replacing it may require the ``flags`` option to be set to ``r+`` rather than the default ``w``. The ``encoding`` can be any one of those accepted by ``Buffer``.

If ``autoClose`` is set to true (default behavior) on ``'error'`` or ``'finish'`` the file descriptor will be closed automatically. If ``autoClose`` is false, then the file descriptor won't be closed, even if there's an error. It is the application's responsibility to close it and make sure there's no file descriptor leak.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fscreatewritestreampath-options>`__

Dir
~~~

.. code-block:: text

   fs.Dir

*Kind:* class · *Arguments:* 2

A class representing a directory stream.

Created by ``fs.opendir()``, ``fs.opendirSync()``, or ``fsPromises.opendir()``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   import { opendir } from 'node:fs/promises';

   try {
     const dir = await opendir('./');
     for await (const dirent of dir)
       console.log(dirent.name);
   } catch (err) {
     console.error(err);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsdir>`__

Dirent
~~~~~~

.. code-block:: text

   fs.Dirent

*Kind:* class · *Arguments:* 0

A representation of a directory entry, which can be a file or a subdirectory within the directory, as returned by reading from an ``fs.Dir``. The directory entry is a combination of the file name and file type pairs.

Additionally, when ``fs.readdir()`` or ``fs.readdirSync()`` is called with the ``withFileTypes`` option set to ``true``, the resulting array is filled with ``fs.Dirent`` objects, rather than strings or ``Buffer``\ s.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsdirent>`__

exists
~~~~~~

.. code-block:: text

   fs.exists(path, callback)

*Kind:* function · *Arguments:* 2

Type: Documentation-only

The ``fs.exists(path, callback)`` API is deprecated. Please use ``fs.stat()`` or ``fs.access()`` instead.

**Arguments**

* ``path``
* ``callback``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0034-fsexistspath-callback>`__

existsSync
~~~~~~~~~~

.. code-block:: text

   fs.existsSync

*Kind:* function · *Arguments:* 0

Type: Runtime

Passing non-supported argument types is deprecated and, instead of returning ``false``, will throw an error in a future version.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0187-passing-invalid-argument-types-to-fsexistssync>`__

fchmod
~~~~~~

.. code-block:: text

   fs.fchmod(fd, mode, callback)

*Kind:* function · *Arguments:* 3

Sets the permissions on the file. No arguments other than a possible exception are given to the completion callback.

See the POSIX fchmod(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``mode`` (*string|integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfchmodfd-mode-callback>`__

fchmodSync
~~~~~~~~~~

.. code-block:: text

   fs.fchmodSync(fd, mode)

*Kind:* function · *Arguments:* 2

Sets the permissions on the file. Returns ``undefined``.

See the POSIX fchmod(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``mode`` (*string|integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfchmodsyncfd-mode>`__

fchown
~~~~~~

.. code-block:: text

   fs.fchown(fd, uid, gid, callback)

*Kind:* function · *Arguments:* 4

Sets the owner of the file. No arguments other than a possible exception are given to the completion callback.

See the POSIX fchown(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``uid`` (*integer*)
* ``gid`` (*integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfchownfd-uid-gid-callback>`__

fchownSync
~~~~~~~~~~

.. code-block:: text

   fs.fchownSync(fd, uid, gid)

*Kind:* function · *Arguments:* 3

Sets the owner of the file. Returns ``undefined``.

See the POSIX fchown(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``uid`` (*integer*): The file's new owner's user id.
* ``gid`` (*integer*): The file's new group's group id.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfchownsyncfd-uid-gid>`__

fdatasync
~~~~~~~~~

.. code-block:: text

   fs.fdatasync(fd, callback)

*Kind:* function · *Arguments:* 2

Forces all currently queued I/O operations associated with the file to the operating system's synchronized I/O completion state. Refer to the POSIX fdatasync(2) documentation for details. No arguments other than a possible exception are given to the completion callback.

**Arguments**

* ``fd`` (*integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfdatasyncfd-callback>`__

fdatasyncSync
~~~~~~~~~~~~~

.. code-block:: text

   fs.fdatasyncSync(fd)

*Kind:* function · *Arguments:* 1

Forces all currently queued I/O operations associated with the file to the operating system's synchronized I/O completion state. Refer to the POSIX fdatasync(2) documentation for details. Returns ``undefined``.

**Arguments**

* ``fd`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfdatasyncsyncfd>`__

fstat
~~~~~

.. code-block:: text

   fs.fstat(fd[, options], callback)

*Kind:* function · *Arguments:* 3

Invokes the callback with the ``fs.Stats`` for the file descriptor.

See the POSIX fstat(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfstatfd-options-callback>`__

fstatSync
~~~~~~~~~

.. code-block:: text

   fs.fstatSync(fd[, options])

*Kind:* function · *Arguments:* 2

Retrieves the ``fs.Stats`` for the file descriptor.

See the POSIX fstat(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfstatsyncfd-options>`__

FSWatcher
~~~~~~~~~

.. code-block:: text

   fs.FSWatcher

*Kind:* class · *Arguments:* 0

A successful call to ``fs.watch()`` method will return a new ``fs.FSWatcher`` object.

All ``fs.FSWatcher`` objects emit a ``'change'`` event whenever a specific watched file is modified.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   import { watch } from 'node:fs';
   // Example when handled through fs.watch() listener
   watch('./tmp', { encoding: 'buffer' }, (eventType, filename) => {
     if (filename) {
       console.log(filename);
       // Prints: <Buffer ...>
     }
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsfswatcher>`__

fsync
~~~~~

.. code-block:: text

   fs.fsync(fd, callback)

*Kind:* function · *Arguments:* 2

Request that all data for the open file descriptor is flushed to the storage device. The specific implementation is operating system and device specific. Refer to the POSIX fsync(2) documentation for more detail. No arguments other than a possible exception are given to the completion callback.

**Arguments**

* ``fd`` (*integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfsyncfd-callback>`__

fsyncSync
~~~~~~~~~

.. code-block:: text

   fs.fsyncSync(fd)

*Kind:* function · *Arguments:* 1

Request that all data for the open file descriptor is flushed to the storage device. The specific implementation is operating system and device specific. Refer to the POSIX fsync(2) documentation for more detail. Returns ``undefined``.

**Arguments**

* ``fd`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfsyncsyncfd>`__

ftruncate
~~~~~~~~~

.. code-block:: text

   fs.ftruncate(fd[, len], callback)

*Kind:* function · *Arguments:* 3

Truncates the file descriptor. No arguments other than a possible exception are given to the completion callback.

See the POSIX ftruncate(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``len`` (*integer*) — optional: **Default:** ``0``
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { open, close, ftruncate } from 'node:fs';

   function closeFd(fd) {
     close(fd, (err) => {
       if (err) throw err;
     });
   }

   open('temp.txt', 'r+', (err, fd) => {
     if (err) throw err;

     try {
       ftruncate(fd, 4, (err) => {
         closeFd(fd);
         if (err) throw err;
       });
     } catch (err) {
       closeFd(fd);
       if (err) throw err;
     }
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsftruncatefd-len-callback>`__

ftruncateSync
~~~~~~~~~~~~~

.. code-block:: text

   fs.ftruncateSync(fd[, len])

*Kind:* function · *Arguments:* 2

Truncates the file descriptor. Returns ``undefined``.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.ftruncate()``.

**Arguments**

* ``fd`` (*integer*)
* ``len`` (*integer*) — optional: **Default:** ``0``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsftruncatesyncfd-len>`__

futimes
~~~~~~~

.. code-block:: text

   fs.futimes(fd, atime, mtime, callback)

*Kind:* function · *Arguments:* 4

Change the file system timestamps of the object referenced by the supplied file descriptor. See ``fs.utimes()``.

**Arguments**

* ``fd`` (*integer*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfutimesfd-atime-mtime-callback>`__

futimesSync
~~~~~~~~~~~

.. code-block:: text

   fs.futimesSync(fd, atime, mtime)

*Kind:* function · *Arguments:* 3

Synchronous version of ``fs.futimes()``. Returns ``undefined``.

**Arguments**

* ``fd`` (*integer*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfutimessyncfd-atime-mtime>`__

glob
~~~~

.. code-block:: text

   fs.glob(pattern[, options], callback)

*Kind:* function · *Arguments:* 3

When ``followSymlinks`` is enabled, detected symbolic link cycles are not traversed recursively.

**Arguments**

* ``pattern`` (*string|string\[]*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const { glob } = require('node:fs');

   glob('**/*.js', (err, matches) => {
     if (err) throw err;
     console.log(matches);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsglobpattern-options-callback>`__

globSync
~~~~~~~~

.. code-block:: text

   fs.globSync(pattern[, options])

*Kind:* function · *Arguments:* 2

When ``followSymlinks`` is enabled, detected symbolic link cycles are not traversed recursively.

**Arguments**

* ``pattern`` (*string|string\[]*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { globSync } = require('node:fs');

   console.log(globSync('**/*.js'));

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsglobsyncpattern-options>`__

lchown
~~~~~~

.. code-block:: text

   fs.lchown(path, uid, gid, callback)

*Kind:* function · *Arguments:* 4

Type: Deprecation revoked

The ``fs.lchown(path, uid, gid, callback)`` API was deprecated. The deprecation was revoked because the requisite supporting APIs were added in libuv.

**Arguments**

* ``path``
* ``uid``
* ``gid``
* ``callback``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0037-fslchownpath-uid-gid-callback>`__

lchownSync
~~~~~~~~~~

.. code-block:: text

   fs.lchownSync(path, uid, gid)

*Kind:* function · *Arguments:* 3

Type: Deprecation revoked

The ``fs.lchownSync(path, uid, gid)`` API was deprecated. The deprecation was revoked because the requisite supporting APIs were added in libuv.

**Arguments**

* ``path``
* ``uid``
* ``gid``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0038-fslchownsyncpath-uid-gid>`__

link
~~~~

.. code-block:: text

   fs.link(existingPath, newPath, callback)

*Kind:* function · *Arguments:* 3

Creates a new link from the ``existingPath`` to the ``newPath``. See the POSIX link(2) documentation for more detail. No arguments other than a possible exception are given to the completion callback.

**Arguments**

* ``existingPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslinkexistingpath-newpath-callback>`__

linkSync
~~~~~~~~

.. code-block:: text

   fs.linkSync(existingPath, newPath)

*Kind:* function · *Arguments:* 2

Creates a new link from the ``existingPath`` to the ``newPath``. See the POSIX link(2) documentation for more detail. Returns ``undefined``.

**Arguments**

* ``existingPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslinksyncexistingpath-newpath>`__

lstat
~~~~~

.. code-block:: text

   fs.lstat(path[, options], callback)

*Kind:* function · *Arguments:* 3

Retrieves the ``fs.Stats`` for the symbolic link referred to by the path. The callback gets two arguments ``(err, stats)`` where ``stats`` is a ``fs.Stats`` object. ``lstat()`` is identical to ``stat()``, except that if ``path`` is a symbolic link, then the link itself is stat-ed, not the file that it refers to.

See the POSIX lstat(2) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslstatpath-options-callback>`__

lstatSync
~~~~~~~~~

.. code-block:: text

   fs.lstatSync(path[, options])

*Kind:* function · *Arguments:* 2

Retrieves the ``fs.Stats`` for the symbolic link referred to by ``path``.

See the POSIX lstat(2) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslstatsyncpath-options>`__

lutimes
~~~~~~~

.. code-block:: text

   fs.lutimes(path, atime, mtime, callback)

*Kind:* function · *Arguments:* 4

Changes the access and modification times of a file in the same way as ``fs.utimes()``, with the difference that if the path refers to a symbolic link, then the link is not dereferenced: instead, the timestamps of the symbolic link itself are changed.

No arguments other than a possible exception are given to the completion callback.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslutimespath-atime-mtime-callback>`__

lutimesSync
~~~~~~~~~~~

.. code-block:: text

   fs.lutimesSync(path, atime, mtime)

*Kind:* function · *Arguments:* 3

Change the file system timestamps of the symbolic link referenced by ``path``. Returns ``undefined``, or throws an exception when parameters are incorrect or the operation fails. This is the synchronous version of ``fs.lutimes()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fslutimessyncpath-atime-mtime>`__

mkdir
~~~~~

.. code-block:: text

   fs.mkdir(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronously creates a directory.

The callback is given a possible exception and, if ``recursive`` is ``true``, the first directory path created, ``(err[, path])``. ``path`` can still be ``undefined`` when ``recursive`` is ``true``, if no directory was created (for instance, if it was previously created).

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object|integer*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { mkdir } from 'node:fs';

   // Create ./tmp/a/apple, regardless of whether ./tmp and ./tmp/a exist.
   mkdir('./tmp/a/apple', { recursive: true }, (err) => {
     if (err) throw err;
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsmkdirpath-options-callback>`__

mkdirSync
~~~~~~~~~

.. code-block:: text

   fs.mkdirSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronously creates a directory. Returns ``undefined``, or if ``recursive`` is ``true``, the first directory path created. This is the synchronous version of ``fs.mkdir()``.

See the POSIX mkdir(2) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object|integer*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsmkdirsyncpath-options>`__

mkdtemp
~~~~~~~

.. code-block:: text

   fs.mkdtemp(prefix[, options], callback)

*Kind:* function · *Arguments:* 3

Creates a unique temporary directory.

Generates six random characters to be appended behind a required ``prefix`` to create a unique temporary directory. Due to platform inconsistencies, avoid trailing ``X`` characters in ``prefix``. Some platforms, notably the BSDs, can return more than six random characters, and replace trailing ``X`` characters in ``prefix`` with random characters.

**Arguments**

* ``prefix`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { mkdtemp } from 'node:fs';
   import { join } from 'node:path';
   import { tmpdir } from 'node:os';

   mkdtemp(join(tmpdir(), 'foo-'), (err, directory) => {
     if (err) throw err;
     console.log(directory);
     // Prints: /tmp/foo-itXde2 or C:\Users\...\AppData\Local\Temp\foo-itXde2
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsmkdtempprefix-options-callback>`__

mkdtempDisposableSync
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.mkdtempDisposableSync(prefix[, options])

*Kind:* function · *Arguments:* 2

Returns a disposable object whose ``path`` property holds the created directory path. If ``encoding`` is ``'buffer'``, the ``path`` will be a ``Buffer``. When the object is disposed, the directory and its contents will be removed if it still exists. If the directory cannot be deleted, disposal will throw an error. The object has a ``remove()`` method which will perform the same task.

See the MDN documentation on ``using`` statements for more information about explicit resource management.

**Arguments**

* ``prefix`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsmkdtempdisposablesyncprefix-options>`__

mkdtempSync
~~~~~~~~~~~

.. code-block:: text

   fs.mkdtempSync(prefix[, options])

*Kind:* function · *Arguments:* 2

Returns the created directory path. If ``encoding`` is ``'buffer'``, then the resulting directory path is returned as a ``Buffer``. Otherwise, the path is returned as a ``string`` using the specified encoding.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.mkdtemp()``.

**Arguments**

* ``prefix`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsmkdtempsyncprefix-options>`__

open
~~~~

.. code-block:: text

   fs.open(path[, flags[, mode]], callback)

*Kind:* function · *Arguments:* 4

Asynchronous file open. See the POSIX open(2) documentation for more details.

``mode`` sets the file mode (permission and sticky bits), but only if the file was created. On Windows, only the write permission can be manipulated; see ``fs.chmod()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``flags`` (*string|number*) — optional: See support of file system ``flags``.
* ``mode`` (*string|integer*) — optional: **Default:** ``0o666`` (readable and writable)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsopenpath-flags-mode-callback>`__

openAsBlob
~~~~~~~~~~

.. code-block:: text

   fs.openAsBlob(path[, options])

*Kind:* function · *Arguments:* 2

Returns a ``Blob`` whose data is backed by the given file.

The file must not be modified after the ``Blob`` is created. Any modifications will cause reading the ``Blob`` data to fail with a ``DOMException`` error. Synchronous stat operations on the file when the ``Blob`` is created, and before each read in order to detect whether the file data has been modified on disk.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { openAsBlob } = require('node:fs');

   (async () => {
     const blob = await openAsBlob('the.file.txt');
     const ab = await blob.arrayBuffer();
     blob.stream();
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsopenasblobpath-options>`__

opendir
~~~~~~~

.. code-block:: text

   fs.opendir(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronously open a directory. See the POSIX opendir(3) documentation for more details.

Creates an ``fs.Dir``, which contains all further functions for reading from and cleaning up the directory.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsopendirpath-options-callback>`__

opendirSync
~~~~~~~~~~~

.. code-block:: text

   fs.opendirSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronously open a directory. See opendir(3).

Creates an ``fs.Dir``, which contains all further functions for reading from and cleaning up the directory.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsopendirsyncpath-options>`__

openSync
~~~~~~~~

.. code-block:: text

   fs.openSync(path[, flags[, mode]])

*Kind:* function · *Arguments:* 3

Returns an integer representing the file descriptor.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.open()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``flags`` (*string|number*) — optional: **Default:** ``'r'``.
* ``mode`` (*string|integer*) — optional: **Default:** ``0o666``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsopensyncpath-flags-mode>`__

read
~~~~

.. code-block:: text

   fs.read

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``fs.read()`` legacy ``String`` interface is deprecated. Use the ``Buffer`` API as mentioned in the documentation instead.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0014-fsread-legacy-string-interface>`__

readdir
~~~~~~~

.. code-block:: text

   fs.readdir(path[, options], callback)

*Kind:* function · *Arguments:* 3

Reads the contents of a directory. The callback gets two arguments ``(err, files)`` where ``files`` is an array of the names of the files in the directory excluding ``'.'`` and ``'..'``.

See the POSIX readdir(3) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreaddirpath-options-callback>`__

readdirSync
~~~~~~~~~~~

.. code-block:: text

   fs.readdirSync(path[, options])

*Kind:* function · *Arguments:* 2

Reads the contents of the directory.

See the POSIX readdir(3) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreaddirsyncpath-options>`__

readFile
~~~~~~~~

.. code-block:: text

   fs.readFile(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronously reads the entire contents of a file.

The callback is passed two arguments ``(err, data)``, where ``data`` is the contents of the file.

**Arguments**

* ``path`` (*string|Buffer|URL|integer*): filename or file descriptor
* ``options`` (*Object|string*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { readFile } from 'node:fs';

   readFile('/etc/passwd', (err, data) => {
     if (err) throw err;
     console.log(data);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadfilepath-options-callback>`__

readFileSync
~~~~~~~~~~~~

.. code-block:: text

   fs.readFileSync(path[, options])

*Kind:* function · *Arguments:* 2

Returns the contents of the ``path``.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.readFile()``.

**Arguments**

* ``path`` (*string|Buffer|URL|integer*): filename or file descriptor
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   import { readFileSync } from 'node:fs';

   // macOS, Linux, and Windows
   readFileSync('<directory>');
   // => [Error: EISDIR: illegal operation on a directory, read <directory>]

   //  FreeBSD
   readFileSync('<directory>'); // => <data>

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadfilesyncpath-options>`__

readlink
~~~~~~~~

.. code-block:: text

   fs.readlink(path[, options], callback)

*Kind:* function · *Arguments:* 3

Reads the contents of the symbolic link referred to by ``path``. The callback gets two arguments ``(err, linkString)``.

See the POSIX readlink(2) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadlinkpath-options-callback>`__

readlinkSync
~~~~~~~~~~~~

.. code-block:: text

   fs.readlinkSync(path[, options])

*Kind:* function · *Arguments:* 2

Returns the symbolic link's string value.

See the POSIX readlink(2) documentation for more details.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadlinksyncpath-options>`__

ReadStream
~~~~~~~~~~

.. code-block:: text

   fs.ReadStream

*Kind:* class · *Arguments:* 2

Instances of ``fs.ReadStream`` cannot be constructed directly. They are created and returned using the ``fs.createReadStream()`` function.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsreadstream>`__

readSync
~~~~~~~~

.. code-block:: text

   fs.readSync

*Kind:* function · *Arguments:* 0

Type: End-of-Life

The ``fs.readSync()`` legacy ``String`` interface is deprecated. Use the ``Buffer`` API as mentioned in the documentation instead.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0015-fsreadsync-legacy-string-interface>`__

readv
~~~~~

.. code-block:: text

   fs.readv(fd, buffers[, position], callback)

*Kind:* function · *Arguments:* 4

Read from a file specified by ``fd`` and write to an array of ``ArrayBufferView``\ s using ``readv()``.

``position`` is the offset from the beginning of the file from where data should be read. If ``typeof position !== 'number'``, the data will be read from the current position.

**Arguments**

* ``fd`` (*integer*)
* ``buffers`` (*ArrayBufferView\[]*)
* ``position`` (*integer|null*) — optional: **Default:** ``null``
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadvfd-buffers-position-callback>`__

readvSync
~~~~~~~~~

.. code-block:: text

   fs.readvSync(fd, buffers[, position])

*Kind:* function · *Arguments:* 3

For detailed information, see the documentation of the asynchronous version of this API: ``fs.readv()``.

**Arguments**

* ``fd`` (*integer*)
* ``buffers`` (*ArrayBufferView\[]*)
* ``position`` (*integer|null*) — optional: **Default:** ``null``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadvsyncfd-buffers-position>`__

realpath
~~~~~~~~

.. code-block:: text

   fs.realpath(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronously computes the canonical pathname by resolving ``.``, ``..``, and symbolic links.

A canonical pathname is not necessarily unique. Hard links and bind mounts can expose a file system entity through many pathnames.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrealpathpath-options-callback>`__

realpathSync
~~~~~~~~~~~~

.. code-block:: text

   fs.realpathSync(path[, options])

*Kind:* function · *Arguments:* 2

Returns the resolved pathname.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.realpath()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrealpathsyncpath-options>`__

rename
~~~~~~

.. code-block:: text

   fs.rename(oldPath, newPath, callback)

*Kind:* function · *Arguments:* 3

Asynchronously rename file at ``oldPath`` to the pathname provided as ``newPath``. In the case that ``newPath`` already exists, it will be overwritten. If there is a directory at ``newPath``, an error will be raised instead. No arguments other than a possible exception are given to the completion callback.

See also: rename(2).

**Arguments**

* ``oldPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { rename } from 'node:fs';

   rename('oldFile.txt', 'newFile.txt', (err) => {
     if (err) throw err;
     console.log('Rename complete!');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrenameoldpath-newpath-callback>`__

renameSync
~~~~~~~~~~

.. code-block:: text

   fs.renameSync(oldPath, newPath)

*Kind:* function · *Arguments:* 2

Renames the file from ``oldPath`` to ``newPath``. Returns ``undefined``.

See the POSIX rename(2) documentation for more details.

**Arguments**

* ``oldPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrenamesyncoldpath-newpath>`__

rm
~~~

.. code-block:: text

   fs.rm(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronously removes files and directories (modeled on the standard POSIX ``rm`` utility). No arguments other than a possible exception are given to the completion callback.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrmpath-options-callback>`__

rmdir
~~~~~

.. code-block:: text

   fs.rmdir(path, { recursive: true })

*Kind:* function · *Arguments:* 3

Type: Runtime

In future versions of Node.js, ``recursive`` option will be ignored for ``fs.rmdir``, ``fs.rmdirSync``, and ``fs.promises.rmdir``.

**Arguments**

* ``path``
* ``recursive``
* ``true`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0147-fsrmdirpath-recursive-true>`__

rmdirSync
~~~~~~~~~

.. code-block:: text

   fs.rmdirSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronous rmdir(2). Returns ``undefined``.

Using ``fs.rmdirSync()`` on a file (not a directory) results in an ``ENOENT`` error on Windows and an ``ENOTDIR`` error on POSIX.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrmdirsyncpath-options>`__

rmSync
~~~~~~

.. code-block:: text

   fs.rmSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronously removes files and directories (modeled on the standard POSIX ``rm`` utility). Returns ``undefined``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrmsyncpath-options>`__

stat
~~~~

.. code-block:: text

   fs.stat(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronous stat(2). The callback gets two arguments ``(err, stats)`` where ``stats`` is an ``fs.Stats`` object.

In case of an error, the ``err.code`` will be one of Common System Errors.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { stat } from 'node:fs';

   const pathsToCheck = ['./txtDir', './txtDir/file.txt'];

   for (let i = 0; i < pathsToCheck.length; i++) {
     stat(pathsToCheck[i], (err, stats) => {
       console.log(stats.isDirectory());
       console.log(stats);
     });
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsstatpath-options-callback>`__

statfs
~~~~~~

.. code-block:: text

   fs.statfs(path[, options], callback)

*Kind:* function · *Arguments:* 3

Asynchronous statfs(2). Returns information about the mounted file system which contains ``path``. The callback gets two arguments ``(err, stats)`` where ``stats`` is an ``fs.StatFs`` object.

In case of an error, the ``err.code`` will be one of Common System Errors.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsstatfspath-options-callback>`__

StatFs
~~~~~~

.. code-block:: text

   fs.StatFs

*Kind:* class · *Arguments:* 0

Provides information about a mounted file system.

Objects returned from ``fs.statfs()`` and its synchronous counterpart are of this type. If ``bigint`` in the ``options`` passed to those methods is ``true``, the numeric values will be ``bigint`` instead of ``number``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { statfs } = require('node:fs/promises');

   (async () => {
     const stats = await statfs('/');
     const availableBytes = stats.bsize * stats.bavail;
     console.log(`Available space: ${availableBytes} bytes`);
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsstatfs>`__

statfsSync
~~~~~~~~~~

.. code-block:: text

   fs.statfsSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronous statfs(2). Returns information about the mounted file system which contains ``path``.

In case of an error, the ``err.code`` will be one of Common System Errors.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsstatfssyncpath-options>`__

Stats
~~~~~

.. code-block:: text

   fs.Stats

*Kind:* function · *Arguments:* 0

Type: Runtime

Calling ``fs.Stats`` class directly with ``Stats()`` or ``new Stats()`` is deprecated due to being internals, not intended for public use.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0180-fsstats-constructor>`__

statSync
~~~~~~~~

.. code-block:: text

   fs.statSync(path[, options])

*Kind:* function · *Arguments:* 2

Retrieves the ``fs.Stats`` for the path.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsstatsyncpath-options>`__

StatWatcher
~~~~~~~~~~~

.. code-block:: text

   fs.StatWatcher

*Kind:* class · *Arguments:* 0

A successful call to ``fs.watchFile()`` method will return a new ``fs.StatWatcher`` object.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsstatwatcher>`__

symlink
~~~~~~~

.. code-block:: text

   fs.symlink(target, path[, type], callback)

*Kind:* function · *Arguments:* 4

Creates the link called ``path`` pointing to ``target``. No arguments other than a possible exception are given to the completion callback.

See the POSIX symlink(2) documentation for more details.

**Arguments**

* ``target`` (*string|Buffer|URL*)
* ``path`` (*string|Buffer|URL*)
* ``type`` (*string|null*) — optional: **Default:** ``null``
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { symlink } from 'node:fs';

   symlink('./mew', './mewtwo', callback);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fssymlinktarget-path-type-callback>`__

symlinkSync
~~~~~~~~~~~

.. code-block:: text

   fs.symlinkSync(target, path[, type])

*Kind:* function · *Arguments:* 3

For detailed information, see the documentation of the asynchronous version of this API: ``fs.symlink()``.

**Arguments**

* ``target`` (*string|Buffer|URL*)
* ``path`` (*string|Buffer|URL*)
* ``type`` (*string|null*) — optional: **Default:** ``null``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fssymlinksynctarget-path-type>`__

truncate
~~~~~~~~

.. code-block:: text

   fs.truncate()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

``fs.truncate()`` ``fs.truncateSync()`` usage with a file descriptor is deprecated. Please use ``fs.ftruncate()`` or ``fs.ftruncateSync()`` to work with file descriptors.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0081-fstruncate-using-a-file-descriptor>`__

truncateSync
~~~~~~~~~~~~

.. code-block:: text

   fs.truncateSync(path[, len])

*Kind:* function · *Arguments:* 2

Truncates the file. Returns ``undefined``. A file descriptor can also be passed as the first argument. In this case, ``fs.ftruncateSync()`` is called.

Passing a file descriptor is deprecated and may result in an error being thrown in the future.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``len`` (*integer*) — optional: **Default:** ``0``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fstruncatesyncpath-len>`__

unlink
~~~~~~

.. code-block:: text

   fs.unlink(path, callback)

*Kind:* function · *Arguments:* 2

Asynchronously removes a file or symbolic link. No arguments other than a possible exception are given to the completion callback.

``fs.unlink()`` will not work on a directory, empty or otherwise. To remove a directory, use ``fs.rmdir()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { unlink } from 'node:fs';
   // Assuming that 'path/file.txt' is a regular file.
   unlink('path/file.txt', (err) => {
     if (err) throw err;
     console.log('path/file.txt was deleted');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsunlinkpath-callback>`__

unlinkSync
~~~~~~~~~~

.. code-block:: text

   fs.unlinkSync(path)

*Kind:* function · *Arguments:* 1

Synchronous unlink(2). Returns ``undefined``.

**Arguments**

* ``path`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsunlinksyncpath>`__

unwatchFile
~~~~~~~~~~~

.. code-block:: text

   fs.unwatchFile(filename[, listener])

*Kind:* function · *Arguments:* 2

Stop watching for changes on ``filename``. If ``listener`` is specified, only that particular listener is removed. Otherwise, *all* listeners are removed, effectively stopping watching of ``filename``.

Calling ``fs.unwatchFile()`` with a filename that is not being watched is a no-op, not an error.

**Arguments**

* ``filename`` (*string|Buffer|URL*)
* ``listener`` (*Function*) — optional: Optional, a listener previously attached using

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsunwatchfilefilename-listener>`__

Utf8Stream
~~~~~~~~~~

.. code-block:: text

   fs.Utf8Stream

*Kind:* class · *Arguments:* 1

An optimized UTF-8 stream writer that allows for flushing all the internal buffering on demand. It handles ``EAGAIN`` errors correctly, allowing for customization, for example, by dropping content if the disk is busy.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsutf8stream>`__

utimes
~~~~~~

.. code-block:: text

   fs.utimes(path, atime, mtime, callback)

*Kind:* function · *Arguments:* 4

Change the file system timestamps of the object referenced by ``path``.

The ``atime`` and ``mtime`` arguments follow these rules:

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsutimespath-atime-mtime-callback>`__

utimesSync
~~~~~~~~~~

.. code-block:: text

   fs.utimesSync(path, atime, mtime)

*Kind:* function · *Arguments:* 3

For detailed information, see the documentation of the asynchronous version of this API: ``fs.utimes()``.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsutimessyncpath-atime-mtime>`__

watch
~~~~~

.. code-block:: text

   fs.watch(filename[, options][, listener])

*Kind:* function · *Arguments:* 3

Watch for changes on ``filename``, where ``filename`` is either a file or a directory.

The second argument is optional. If ``options`` is provided as a string, it specifies the ``encoding``. Otherwise ``options`` should be passed as an object.

**Arguments**

* ``filename`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional
* ``listener`` (*Function|undefined*) — optional: **Default:** ``undefined``

**Example**

.. code-block:: javascript

   import { watch } from 'node:fs';
   watch('somedir', (eventType, filename) => {
     console.log(`event type is: ${eventType}`);
     if (filename) {
       console.log(`filename provided: ${filename}`);
     } else {
       console.log('filename not provided');
     }
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswatchfilename-options-listener>`__

watchFile
~~~~~~~~~

.. code-block:: text

   fs.watchFile(filename[, options], listener)

*Kind:* function · *Arguments:* 3

Watch for changes on ``filename``. The callback ``listener`` will be called each time the file is accessed.

The ``options`` argument may be omitted. If provided, it should be an object. The ``options`` object may contain a boolean named ``persistent`` that indicates whether the process should continue to run as long as files are being watched. The ``options`` object may specify an ``interval`` property indicating how often the target should be polled in milliseconds.

**Arguments**

* ``filename`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   import { watchFile } from 'node:fs';

   watchFile('message.text', (curr, prev) => {
     console.log(`the current mtime is: ${curr.mtime}`);
     console.log(`the previous mtime was: ${prev.mtime}`);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswatchfilefilename-options-listener>`__

write
~~~~~

.. code-block:: text

   fs.write()

*Kind:* function · *Arguments:* 0

Type: End-of-Life

Implicit coercion of objects with own ``toString`` property, passed as second parameter in ``fs.write()``, ``fs.writeFile()``, ``fs.appendFile()``, ``fs.writeFileSync()``, and ``fs.appendFileSync()`` is deprecated. Convert them to primitive strings.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0162-fswrite-fswritefilesync-coercion-to-string>`__

writeFile
~~~~~~~~~

.. code-block:: text

   fs.writeFile(file, data[, options], callback)

*Kind:* function · *Arguments:* 4

When ``file`` is a filename, asynchronously writes data to the file, replacing the file if it already exists. ``data`` can be a string or a buffer.

When ``file`` is a file descriptor, the behavior is similar to calling ``fs.write()`` directly (which is recommended). See the notes below on using a file descriptor.

**Arguments**

* ``file`` (*string|Buffer|URL|integer*): filename or file descriptor
* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``options`` (*Object|string*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   import { writeFile } from 'node:fs';
   import { Buffer } from 'node:buffer';

   const data = new Uint8Array(Buffer.from('Hello Node.js'));
   writeFile('message.txt', data, (err) => {
     if (err) throw err;
     console.log('The file has been saved!');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswritefilefile-data-options-callback>`__

writeFileSync
~~~~~~~~~~~~~

.. code-block:: text

   fs.writeFileSync(file, data[, options])

*Kind:* function · *Arguments:* 3

The ``mode`` option only affects the newly created file. See ``fs.open()`` for more details.

For detailed information, see the documentation of the asynchronous version of this API: ``fs.writeFile()``.

**Arguments**

* ``file`` (*string|Buffer|URL|integer*): filename or file descriptor
* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``options`` (*Object|string*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswritefilesyncfile-data-options>`__

WriteStream
~~~~~~~~~~~

.. code-block:: text

   fs.WriteStream

*Kind:* class · *Arguments:* 2

Instances of ``fs.WriteStream`` cannot be constructed directly. They are created and returned using the ``fs.createWriteStream()`` function.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fswritestream>`__

writeSync
~~~~~~~~~

.. code-block:: text

   fs.writeSync(fd, buffer, offset[, length[, position]])

*Kind:* function · *Arguments:* 5

For detailed information, see the documentation of the asynchronous version of this API: ``fs.write(fd, buffer...)``.

**Arguments**

* ``fd`` (*integer*)
* ``buffer`` (*Buffer|TypedArray|DataView*)
* ``offset`` (*integer*): **Default:** ``0``
* ``length`` (*integer*) — optional: **Default:** ``buffer.byteLength - offset``
* ``position`` (*integer|null*) — optional: **Default:** ``null``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswritesyncfd-buffer-offset-length-position>`__

writev
~~~~~~

.. code-block:: text

   fs.writev(fd, buffers[, position], callback)

*Kind:* function · *Arguments:* 4

Write an array of ``ArrayBufferView``\ s to the file specified by ``fd`` using ``writev()``.

``position`` is the offset from the beginning of the file where this data should be written. If ``typeof position !== 'number'``, the data will be written at the current position.

**Arguments**

* ``fd`` (*integer*)
* ``buffers`` (*ArrayBufferView\[]*)
* ``position`` (*integer|null*) — optional: **Default:** ``null``
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswritevfd-buffers-position-callback>`__

writevSync
~~~~~~~~~~

.. code-block:: text

   fs.writevSync(fd, buffers[, position])

*Kind:* function · *Arguments:* 3

For detailed information, see the documentation of the asynchronous version of this API: ``fs.writev()``.

**Arguments**

* ``fd`` (*integer*)
* ``buffers`` (*ArrayBufferView\[]*)
* ``position`` (*integer|null*) — optional: **Default:** ``null``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fswritevsyncfd-buffers-position>`__

Dir class
---------

**Dir** — A class representing a directory stream.

Created by ``fs.opendir()``, ``fs.opendirSync()``, or ``fsPromises.opendir()``.

Dir.close
~~~~~~~~~

.. code-block:: text

   dir.close(callback)

*Kind:* function · *Arguments:* 1

Asynchronously close the directory's underlying resource handle. Subsequent reads will result in errors.

The ``callback`` will be called after the resource handle has been closed.

**Arguments**

* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirclosecallback>`__

Dir.closeSync
~~~~~~~~~~~~~

.. code-block:: text

   dir.closeSync()

*Kind:* function · *Arguments:* 0

Synchronously close the directory's underlying resource handle. Subsequent reads will result in errors.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirclosesync>`__

Dir.read
~~~~~~~~

.. code-block:: text

   dir.read(callback)

*Kind:* function · *Arguments:* 1

Asynchronously read the next directory entry via readdir(3) as an ``fs.Dirent``.

After the read is completed, the ``callback`` will be called with an ``fs.Dirent``, or ``null`` if there are no more directory entries to read.

**Arguments**

* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirreadcallback>`__

Dir.readSync
~~~~~~~~~~~~

.. code-block:: text

   dir.readSync()

*Kind:* function · *Arguments:* 0

Synchronously read the next directory entry as an ``fs.Dirent``. See the POSIX readdir(3) documentation for more detail.

If there are no more directory entries to read, ``null`` will be returned.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirreadsync>`__

Dirent class
------------

**Dirent** — A representation of a directory entry, which can be a file or a subdirectory within the directory, as returned by reading from an ``fs.Dir``. The directory entry is a combination of the file name and file type pairs.

Additionally, when ``fs.readdir()`` or ``fs.readdirSync()`` is called with the ``withFileTypes`` option set to ``true``, the resulting array is filled with ``fs.Dirent`` objects, rather than strings or ``Buffer``\ s.

Dirent.isBlockDevice
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dirent.isBlockDevice()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a block device.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntisblockdevice>`__

Dirent.isCharacterDevice
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dirent.isCharacterDevice()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a character device.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntischaracterdevice>`__

Dirent.isDirectory
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dirent.isDirectory()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a file system directory.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntisdirectory>`__

Dirent.isFIFO
~~~~~~~~~~~~~

.. code-block:: text

   dirent.isFIFO()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a first-in-first-out (FIFO) pipe.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntisfifo>`__

Dirent.isFile
~~~~~~~~~~~~~

.. code-block:: text

   dirent.isFile()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a regular file.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntisfile>`__

Dirent.isSocket
~~~~~~~~~~~~~~~

.. code-block:: text

   dirent.isSocket()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a socket.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntissocket>`__

Dirent.isSymbolicLink
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dirent.isSymbolicLink()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Dirent`` object describes a symbolic link.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#direntissymboliclink>`__

FSWatcher class
---------------

**FSWatcher** — A successful call to ``fs.watch()`` method will return a new ``fs.FSWatcher`` object.

All ``fs.FSWatcher`` objects emit a ``'change'`` event whenever a specific watched file is modified.

FSWatcher.close
~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.close()

*Kind:* function · *Arguments:* 0

Closes the file handle after waiting for any pending operation on the handle to complete.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   import { open } from 'node:fs/promises';

   let filehandle;
   try {
     filehandle = await open('thefile.txt', 'r');
   } finally {
     await filehandle?.close();
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandleclose>`__

FSWatcher.ref
~~~~~~~~~~~~~

.. code-block:: text

   watcher.ref()

*Kind:* function · *Arguments:* 0

When called, requests that the Node.js event loop *not* exit so long as the ``fs.FSWatcher`` is active. Calling ``watcher.ref()`` multiple times will have no effect.

By default, all ``fs.FSWatcher`` objects are "ref'ed", making it normally unnecessary to call ``watcher.ref()`` unless ``watcher.unref()`` had been called previously.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#watcherref>`__

FSWatcher.unref
~~~~~~~~~~~~~~~

.. code-block:: text

   watcher.unref()

*Kind:* function · *Arguments:* 0

When called, the active ``fs.FSWatcher`` object will not require the Node.js event loop to remain active. If there is no other activity keeping the event loop running, the process may exit before the ``fs.FSWatcher`` object's callback is invoked. Calling ``watcher.unref()`` multiple times will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#watcherunref>`__

native
------

**native** — Asynchronous realpath(3).

The ``callback`` gets two arguments ``(err, resolvedPath)``.

native.close
~~~~~~~~~~~~

.. code-block:: text

   filehandle.close()

*Kind:* function · *Arguments:* 0

Closes the file handle after waiting for any pending operation on the handle to complete.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   import { open } from 'node:fs/promises';

   let filehandle;
   try {
     filehandle = await open('thefile.txt', 'r');
   } finally {
     await filehandle?.close();
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandleclose>`__

native.fstat
~~~~~~~~~~~~

.. code-block:: text

   fs.fstat(fd[, options], callback)

*Kind:* function · *Arguments:* 3

Invokes the callback with the ``fs.Stats`` for the file descriptor.

See the POSIX fstat(2) documentation for more detail.

**Arguments**

* ``fd`` (*integer*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsfstatfd-options-callback>`__

native.open
~~~~~~~~~~~

.. code-block:: text

   fsPromises.open(path, flags[, mode])

*Kind:* function · *Arguments:* 3

Opens a ``FileHandle``.

Refer to the POSIX open(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``flags`` (*string|number*): See support of file system ``flags``.
* ``mode`` (*string|integer*) — optional: Sets the file mode (permission and sticky bits)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesopenpath-flags-mode>`__

native.read
~~~~~~~~~~~

.. code-block:: text

   dir.read()

*Kind:* function · *Arguments:* 0

Asynchronously read the next directory entry via readdir(3) as an ``fs.Dirent``.

A promise is returned that will be fulfilled with an ``fs.Dirent``, or ``null`` if there are no more directory entries to read.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirread>`__

native.statfs
~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.statfs(path[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesstatfspath-options>`__

native.statfsSync
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.statfsSync(path[, options])

*Kind:* function · *Arguments:* 2

Synchronous statfs(2). Returns information about the mounted file system which contains ``path``.

In case of an error, the ``err.code`` will be one of Common System Errors.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fsstatfssyncpath-options>`__

native.write
~~~~~~~~~~~~

.. code-block:: text

   filehandle.write(buffer, offset[, length[, position]])

*Kind:* function · *Arguments:* 4

Write ``buffer`` to the file.

The promise is fulfilled with an object containing two properties:

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView*): a reference to the
* ``offset`` (*integer*): The start position from within ``buffer`` where the data
* ``length`` (*integer*) — optional: The number of bytes from ``buffer`` to write. **Default:**
* ``position`` (*integer|null*) — optional: The offset from the beginning of the file where the

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlewritebuffer-offset-length-position>`__

promises
--------

promises.access
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.access(path[, mode])

*Kind:* function · *Arguments:* 2

Tests a user's permissions for the file or directory specified by ``path``. The ``mode`` argument is an optional integer that specifies the accessibility checks to be performed. ``mode`` should be either the value ``fs.constants.F_OK`` or a mask consisting of the bitwise OR of any of ``fs.constants.R_OK``, ``fs.constants.W_OK``, and ``fs.constants.X_OK`` (e.g. ``fs.constants.W_OK | fs.constants.R_OK``). Check File access constants for possible values of ``mode``.

If the accessibility check is successful, the promise is fulfilled with no value. If any of the accessibility checks fail, the promise is rejected with an ``Error`` object. The following example checks if the file ``/etc/passwd`` can be read and written by the current process.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*integer*) — optional: **Default:** ``fs.constants.F_OK``

**Example**

.. code-block:: javascript

   import { access, constants } from 'node:fs/promises';

   try {
     await access('/etc/passwd', constants.R_OK | constants.W_OK);
     console.log('can access');
   } catch {
     console.error('cannot access');
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesaccesspath-mode>`__

promises.appendFile
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.appendFile(path, data[, options])

*Kind:* function · *Arguments:* 3

Asynchronously append data to a file, creating the file if it does not yet ``data`` can be a string, a buffer, an ``AsyncIterable``, or an ``Iterable`` object.

If ``options`` is a string, then it specifies the ``encoding``.

**Arguments**

* ``path`` (*string|Buffer|URL|FileHandle*): filename or ``FileHandle``
* ``data`` (*string|Buffer|TypedArray|DataView|AsyncIterable|Iterable*)
* ``options`` (*Object|string*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesappendfilepath-data-options>`__

promises.chmod
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.chmod(path, mode)

*Kind:* function · *Arguments:* 2

Changes the permissions of a file.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``mode`` (*string|integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseschmodpath-mode>`__

promises.chown
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.chown(path, uid, gid)

*Kind:* function · *Arguments:* 3

Changes the ownership of a file.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``uid`` (*integer*)
* ``gid`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseschownpath-uid-gid>`__

promises.copyFile
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.copyFile(src, dest[, mode])

*Kind:* function · *Arguments:* 3

Asynchronously copies ``src`` to ``dest``. By default, ``dest`` is overwritten if it already exists.

Symbolic links are followed. If ``src`` is a symbolic link, the target file is copied. If ``dest`` is a symbolic link, the target file is overwritten unless ``mode`` contains ``fs.constants.COPYFILE_EXCL``.

**Arguments**

* ``src`` (*string|Buffer|URL*): source filename to copy
* ``dest`` (*string|Buffer|URL*): destination filename of the copy operation
* ``mode`` (*integer*) — optional: Optional modifiers that specify the behavior of the copy

**Example**

.. code-block:: javascript

   import { copyFile, constants } from 'node:fs/promises';

   try {
     await copyFile('source.txt', 'destination.txt');
     console.log('source.txt was copied to destination.txt');
   } catch {
     console.error('The file could not be copied');
   }

   // By using COPYFILE_EXCL, the operation will fail if destination.txt exists.
   try {
     await copyFile('source.txt', 'destination.txt', constants.COPYFILE_EXCL);
     console.log('source.txt was copied to destination.txt');
   } catch {
     console.error('The file could not be copied');
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisescopyfilesrc-dest-mode>`__

promises.cp
~~~~~~~~~~~

.. code-block:: text

   fsPromises.cp(src, dest[, options])

*Kind:* function · *Arguments:* 3

Asynchronously copies the entire directory structure from ``src`` to ``dest``, including subdirectories and files.

When copying a directory to another directory, globs are not supported and behavior is similar to ``cp dir1/ dir2/``.

**Arguments**

* ``src`` (*string|URL*): source path to copy.
* ``dest`` (*string|URL*): destination path to copy to.
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisescpsrc-dest-options>`__

promises.FileHandle
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   FileHandle

*Kind:* class · *Arguments:* 0

A ``FileHandle`` object is an object wrapper for a numeric file descriptor.

Instances of the ``FileHandle`` object are created by the ``fsPromises.open()`` method.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { open } = require('node:fs/promises');
   const { text } = require('node:stream/iter');
   const { compressGzip } = require('node:zlib/iter');

   async function run() {
     const fh = await open('input.txt', 'r');

     // Read as text
     console.log(await text(fh.pull({ autoClose: true })));

     // Read 1 KB starting at byte 100
     const fh2 = await open('input.txt', 'r');
     console.log(await text(fh2.pull({ start: 100, limit: 1024, autoClose: true })));

     // Read with compression
     const fh3 = await open('input.txt', 'r');
     const compressed = fh3.pull(compressGzip(), { autoClose: true });
   }

   run().catch(console.error);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#class-filehandle>`__

promises.glob
~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.glob(pattern[, options])

*Kind:* function · *Arguments:* 2

When ``followSymlinks`` is enabled, detected symbolic link cycles are not traversed recursively.

**Arguments**

* ``pattern`` (*string|string\[]*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { glob } = require('node:fs/promises');

   (async () => {
     for await (const entry of glob('**/*.js'))
       console.log(entry);
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesglobpattern-options>`__

promises.lchown
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.lchown(path, uid, gid)

*Kind:* function · *Arguments:* 3

Changes the ownership on a symbolic link.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``uid`` (*integer*)
* ``gid`` (*integer*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseslchownpath-uid-gid>`__

promises.link
~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.link(existingPath, newPath)

*Kind:* function · *Arguments:* 2

Creates a new link from the ``existingPath`` to the ``newPath``. See the POSIX link(2) documentation for more detail.

**Arguments**

* ``existingPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseslinkexistingpath-newpath>`__

promises.lstat
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.lstat(path[, options])

*Kind:* function · *Arguments:* 2

Equivalent to ``fsPromises.stat()`` unless ``path`` refers to a symbolic link, in which case the link itself is stat-ed, not the file that it refers to. Refer to the POSIX lstat(2) document for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseslstatpath-options>`__

promises.lutimes
~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.lutimes(path, atime, mtime)

*Kind:* function · *Arguments:* 3

Changes the access and modification times of a file in the same way as ``fsPromises.utimes()``, with the difference that if the path refers to a symbolic link, then the link is not dereferenced: instead, the timestamps of the symbolic link itself are changed.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseslutimespath-atime-mtime>`__

promises.mkdir
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.mkdir(path[, options])

*Kind:* function · *Arguments:* 2

Asynchronously creates a directory.

The optional ``options`` argument can be an integer specifying ``mode`` (permission and sticky bits), or an object with a ``mode`` property and a ``recursive`` property indicating whether parent directories should be created. Calling ``fsPromises.mkdir()`` when ``path`` is a directory that exists results in a rejection only when ``recursive`` is false.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object|integer*) — optional

**Example**

.. code-block:: javascript

   const { mkdir } = require('node:fs/promises');
   const { join } = require('node:path');

   async function makeDirectory() {
     const projectFolder = join(__dirname, 'test', 'project');
     const dirCreation = await mkdir(projectFolder, { recursive: true });

     console.log(dirCreation);
     return dirCreation;
   }

   makeDirectory().catch(console.error);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesmkdirpath-options>`__

promises.mkdtemp
~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.mkdtemp(prefix[, options])

*Kind:* function · *Arguments:* 2

Creates a unique temporary directory. A unique directory name is generated by appending six random characters to the end of the provided ``prefix``. Due to platform inconsistencies, avoid trailing ``X`` characters in ``prefix``. Some platforms, notably the BSDs, can return more than six random characters, and replace trailing ``X`` characters in ``prefix`` with random characters.

The optional ``options`` argument can be a string specifying an encoding, or an object with an ``encoding`` property specifying the character encoding to use.

**Arguments**

* ``prefix`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

**Example**

.. code-block:: javascript

   import { mkdtemp } from 'node:fs/promises';
   import { join } from 'node:path';
   import { tmpdir } from 'node:os';

   try {
     await mkdtemp(join(tmpdir(), 'foo-'));
   } catch (err) {
     console.error(err);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesmkdtempprefix-options>`__

promises.mkdtempDisposable
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.mkdtempDisposable(prefix[, options])

*Kind:* function · *Arguments:* 2

The resulting Promise holds an async-disposable object whose ``path`` property holds the created directory path. If ``encoding`` is ``'buffer'``, the ``path`` will also be a ``Buffer``, otherwise a ``string``. When the object is disposed, the directory and its contents will be removed asynchronously if it still exists. If the directory cannot be deleted, disposal will throw an error. The object has an async ``remove()`` method which will perform the same task.

Both this function and the disposal function on the resulting object are async, so it should be used with ``await`` + ``await using`` as in ``await using dir = await fsPromises.mkdtempDisposable('prefix')``.

**Arguments**

* ``prefix`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesmkdtempdisposableprefix-options>`__

promises.open
~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.open(path, flags[, mode])

*Kind:* function · *Arguments:* 3

Opens a ``FileHandle``.

Refer to the POSIX open(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``flags`` (*string|number*): See support of file system ``flags``.
* ``mode`` (*string|integer*) — optional: Sets the file mode (permission and sticky bits)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesopenpath-flags-mode>`__

promises.opendir
~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.opendir(path[, options])

*Kind:* function · *Arguments:* 2

Asynchronously open a directory for iterative scanning. See the POSIX opendir(3) documentation for more detail.

Creates an ``fs.Dir``, which contains all further functions for reading from and cleaning up the directory.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { opendir } from 'node:fs/promises';

   try {
     const dir = await opendir('./');
     for await (const dirent of dir)
       console.log(dirent.name);
   } catch (err) {
     console.error(err);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesopendirpath-options>`__

promises.readdir
~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.readdir(path[, options])

*Kind:* function · *Arguments:* 2

Reads the contents of a directory.

The optional ``options`` argument can be a string specifying an encoding, or an object with an ``encoding`` property specifying the character encoding to use for the filenames. If the ``encoding`` is set to ``'buffer'``, the filenames returned will be passed as ``Buffer`` objects.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

**Example**

.. code-block:: javascript

   import { readdir } from 'node:fs/promises';

   try {
     const files = await readdir(path);
     for (const file of files)
       console.log(file);
   } catch (err) {
     console.error(err);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesreaddirpath-options>`__

promises.readFile
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.readFile(path[, options])

*Kind:* function · *Arguments:* 2

Asynchronously reads the entire contents of a file.

If no encoding is specified (using ``options.encoding``), the data is returned as a ``Buffer`` object. Otherwise, the data will be a string.

**Arguments**

* ``path`` (*string|Buffer|URL|FileHandle*): filename or ``FileHandle``
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const { readFile } = require('node:fs/promises');
   const { resolve } = require('node:path');
   async function logFile() {
     try {
       const filePath = resolve('./package.json');
       const contents = await readFile(filePath, { encoding: 'utf8' });
       console.log(contents);
     } catch (err) {
       console.error(err.message);
     }
   }
   logFile();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesreadfilepath-options>`__

promises.readlink
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.readlink(path[, options])

*Kind:* function · *Arguments:* 2

Reads the contents of the symbolic link referred to by ``path``. See the POSIX readlink(2) documentation for more detail. The promise is fulfilled with the ``linkString`` upon success.

The optional ``options`` argument can be a string specifying an encoding, or an object with an ``encoding`` property specifying the character encoding to use for the link path returned. If the ``encoding`` is set to ``'buffer'``, the link path returned will be passed as a ``Buffer`` object.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesreadlinkpath-options>`__

promises.realpath
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.realpath(path[, options])

*Kind:* function · *Arguments:* 2

Determines the actual location of ``path`` using the same semantics as the ``fs.realpath.native()`` function.

Only paths that can be converted to UTF8 strings are supported.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesrealpathpath-options>`__

promises.rename
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.rename(oldPath, newPath)

*Kind:* function · *Arguments:* 2

Renames ``oldPath`` to ``newPath``.

**Arguments**

* ``oldPath`` (*string|Buffer|URL*)
* ``newPath`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesrenameoldpath-newpath>`__

promises.rm
~~~~~~~~~~~

.. code-block:: text

   fsPromises.rm(path[, options])

*Kind:* function · *Arguments:* 2

Removes files and directories (modeled on the standard POSIX ``rm`` utility).

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesrmpath-options>`__

promises.rmdir
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.rmdir(path[, options])

*Kind:* function · *Arguments:* 2

Removes the directory identified by ``path``.

Using ``fsPromises.rmdir()`` on a file (not a directory) results in the promise being rejected with an ``ENOENT`` error on Windows and an ``ENOTDIR`` error on POSIX.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesrmdirpath-options>`__

promises.stat
~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.stat(path[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesstatpath-options>`__

promises.statfs
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.statfs(path[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesstatfspath-options>`__

promises.symlink
~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.symlink(target, path[, type])

*Kind:* function · *Arguments:* 3

Creates a symbolic link.

The ``type`` argument is only used on Windows platforms and can be one of ``'dir'``, ``'file'``, or ``'junction'``. If the ``type`` argument is ``null``, Node.js will autodetect ``target`` type and use ``'file'`` or ``'dir'``. If the ``target`` does not exist, ``'file'`` will be used. Windows junction points require the destination path to be absolute. When using ``'junction'``, the ``target`` argument will automatically be normalized to absolute path. Junction points on NTFS volumes can only point to directories.

**Arguments**

* ``target`` (*string|Buffer|URL*)
* ``path`` (*string|Buffer|URL*)
* ``type`` (*string|null*) — optional: **Default:** ``null``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisessymlinktarget-path-type>`__

promises.truncate
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.truncate(path[, len])

*Kind:* function · *Arguments:* 2

Truncates (shortens or extends the length) of the content at ``path`` to ``len`` bytes.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``len`` (*integer*) — optional: **Default:** ``0``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisestruncatepath-len>`__

promises.unlink
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.unlink(path)

*Kind:* function · *Arguments:* 1

If ``path`` refers to a symbolic link, then the link is removed without affecting the file or directory to which that link refers. If the ``path`` refers to a file path that is not a symbolic link, the file is deleted. See the POSIX unlink(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesunlinkpath>`__

promises.utimes
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.utimes(path, atime, mtime)

*Kind:* function · *Arguments:* 3

Change the file system timestamps of the object referenced by ``path``.

The ``atime`` and ``mtime`` arguments follow these rules:

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesutimespath-atime-mtime>`__

promises.watch
~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.watch(filename[, options])

*Kind:* function · *Arguments:* 2

Returns an async iterator that watches for changes on ``filename``, where ``filename`` is either a file or a directory.

On most platforms, ``'rename'`` is emitted whenever a filename appears or disappears in the directory.

**Arguments**

* ``filename`` (*string|Buffer|URL*)
* ``options`` (*string|Object*) — optional

**Example**

.. code-block:: javascript

   const { watch } = require('node:fs/promises');

   const ac = new AbortController();
   const { signal } = ac;
   setTimeout(() => ac.abort(), 10000);

   (async () => {
     try {
       const watcher = watch(__filename, { signal });
       for await (const event of watcher)
         console.log(event);
     } catch (err) {
       if (err.name === 'AbortError')
         return;
       throw err;
     }
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseswatchfilename-options>`__

promises.writeFile
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.writeFile(file, data[, options])

*Kind:* function · *Arguments:* 3

Asynchronously writes data to a file, replacing the file if it already exists. ``data`` can be a string, a buffer, an ``AsyncIterable``, or an ``Iterable`` object.

The ``encoding`` option is ignored if ``data`` is a buffer.

**Arguments**

* ``file`` (*string|Buffer|URL|FileHandle*): filename or ``FileHandle``
* ``data`` (*string|Buffer|TypedArray|DataView|AsyncIterable|Iterable*)
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   import { writeFile } from 'node:fs/promises';
   import { Buffer } from 'node:buffer';

   try {
     const controller = new AbortController();
     const { signal } = controller;
     const data = new Uint8Array(Buffer.from('Hello Node.js'));
     const promise = writeFile('message.txt', data, { signal });

     // Abort the request before the promise settles.
     controller.abort();

     await promise;
   } catch (err) {
     // When a request is aborted - err is an AbortError
     console.error(err);
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromiseswritefilefile-data-options>`__

promises.FileHandle
-------------------

**promises.FileHandle** — A ``FileHandle`` object is an object wrapper for a numeric file descriptor.

Instances of the ``FileHandle`` object are created by the ``fsPromises.open()`` method.

promises.FileHandle.appendFile
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.appendFile(data[, options])

*Kind:* function · *Arguments:* 2

Alias of ``filehandle.writeFile()``.

When operating on file handles, the mode cannot be changed from what it was set to with ``fsPromises.open()``. Therefore, this is equivalent to ``filehandle.writeFile()``.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView|AsyncIterable|Iterable*)
* ``options`` (*Object|string*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandleappendfiledata-options>`__

promises.FileHandle.chmod
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.chmod(mode)

*Kind:* function · *Arguments:* 1

Modifies the permissions on the file. See chmod(2).

**Arguments**

* ``mode`` (*integer*): the file mode bit mask.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlechmodmode>`__

promises.FileHandle.chown
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.chown(uid, gid)

*Kind:* function · *Arguments:* 2

Changes the ownership of the file. A wrapper for chown(2).

**Arguments**

* ``uid`` (*integer*): The file's new owner's user id.
* ``gid`` (*integer*): The file's new group's group id.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlechownuid-gid>`__

promises.FileHandle.close
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.close()

*Kind:* function · *Arguments:* 0

Closes the file handle after waiting for any pending operation on the handle to complete.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   import { open } from 'node:fs/promises';

   let filehandle;
   try {
     filehandle = await open('thefile.txt', 'r');
   } finally {
     await filehandle?.close();
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandleclose>`__

promises.FileHandle.createReadStream
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.createReadStream([options])

*Kind:* function · *Arguments:* 1

``options`` can include ``start`` and ``end`` values to read a range of bytes from the file instead of the entire file. Both ``start`` and ``end`` are inclusive and start counting at 0, allowed values are in the \0, [``Number.MAX_SAFE_INTEGER``] range. If ``start`` is omitted or ``undefined``, ``filehandle.createReadStream()`` reads sequentially from the current file position. The ``encoding`` can be any one of those accepted by ``Buffer``.

If the ``FileHandle`` points to a character device that only supports blocking reads (such as keyboard or sound card), read operations do not finish until data is available. This can prevent the process from exiting and the stream from closing naturally.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { open } from 'node:fs/promises';

   const fd = await open('/dev/input/event0');
   // Create a stream from some character device.
   const stream = fd.createReadStream();
   setTimeout(() => {
     stream.close(); // This may not close the stream.
     // Artificially marking end-of-stream, as if the underlying resource had
     // indicated end-of-file by itself, allows the stream to close.
     // This does not cancel pending read operations, and if there is such an
     // operation, the process may still not be able to exit successfully
     // until it finishes.
     stream.push(null);
     stream.read(0);
   }, 100);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlecreatereadstreamoptions>`__

promises.FileHandle.createWriteStream
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.createWriteStream([options])

*Kind:* function · *Arguments:* 1

``options`` may also include a ``start`` option to allow writing data at some position past the beginning of the file, allowed values are in the \0, [``Number.MAX_SAFE_INTEGER``] range. Modifying a file rather than replacing it may require the ``flags`` ``open`` option to be set to ``r+`` rather than the default ``r``. The ``encoding`` can be any one of those accepted by ``Buffer``.

If ``autoClose`` is set to true (default behavior) on ``'error'`` or ``'finish'`` the file descriptor will be closed automatically. If ``autoClose`` is false, then the file descriptor won't be closed, even if there's an error. It is the application's responsibility to close it and make sure there's no file descriptor leak.

**Arguments**

* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlecreatewritestreamoptions>`__

promises.FileHandle.datasync
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.datasync()

*Kind:* function · *Arguments:* 0

Forces all currently queued I/O operations associated with the file to the operating system's synchronized I/O completion state. Refer to the POSIX fdatasync(2) documentation for details.

Unlike ``filehandle.sync`` this method does not flush modified metadata.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandledatasync>`__

promises.FileHandle.pull
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.pull([...transforms][, options])

*Kind:* function · *Arguments:* 2

Return the file contents as an async iterable using the ``node:stream/iter`` pull model. Reads are performed in ``chunkSize``-byte chunks (default 128 KB). If transforms are provided, they are applied via ``stream/iter pull()``.

The file handle is locked while the iterable is being consumed and unlocked when iteration completes, an error occurs, or the consumer breaks.

**Arguments**

* ``...transforms`` — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { open } = require('node:fs/promises');
   const { text } = require('node:stream/iter');
   const { compressGzip } = require('node:zlib/iter');

   async function run() {
     const fh = await open('input.txt', 'r');

     // Read as text
     console.log(await text(fh.pull({ autoClose: true })));

     // Read 1 KB starting at byte 100
     const fh2 = await open('input.txt', 'r');
     console.log(await text(fh2.pull({ start: 100, limit: 1024, autoClose: true })));

     // Read with compression
     const fh3 = await open('input.txt', 'r');
     const compressed = fh3.pull(compressGzip(), { autoClose: true });
   }

   run().catch(console.error);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlepulltransforms-options>`__

promises.FileHandle.pullSync
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.pullSync([...transforms][, options])

*Kind:* function · *Arguments:* 2

Synchronous counterpart of ``filehandle.pull()``. Returns a sync iterable that reads the file using synchronous I/O on the main thread. Reads are performed in ``chunkSize``-byte chunks (default 128 KB).

The file handle is locked while the iterable is being consumed. Unlike the async ``pull()``, this method does not support ``AbortSignal`` since all operations are synchronous.

**Arguments**

* ``...transforms`` — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { open } = require('node:fs/promises');
   const { textSync, pipeToSync } = require('node:stream/iter');
   const { compressGzipSync, decompressGzipSync } = require('node:zlib/iter');

   async function run() {
     const fh = await open('input.txt', 'r');

     // Read as text (sync)
     console.log(textSync(fh.pullSync({ autoClose: true })));

     // Sync compress pipeline: file -> gzip -> file
     const src = await open('input.txt', 'r');
     const dst = await open('output.gz', 'w');
     pipeToSync(
       src.pullSync(compressGzipSync(), { autoClose: true }),
       dst.writer({ autoClose: true }),
     );
   }

   run().catch(console.error);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlepullsynctransforms-options>`__

promises.FileHandle.read
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.read(buffer, offset, length, position)

*Kind:* function · *Arguments:* 4

Reads data from the file and stores that in the given buffer.

If the file is not modified concurrently, the end-of-file is reached when the number of bytes read is zero.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView*): A buffer that will be filled with the
* ``offset`` (*integer*): The location in the buffer at which to start filling.
* ``length`` (*integer*): The number of bytes to read. **Default:**
* ``position`` (*integer|bigint|null*): The location where to begin reading data

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadbuffer-offset-length-position>`__

promises.FileHandle.readableWebStream
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.readableWebStream([options])

*Kind:* function · *Arguments:* 1

Returns a byte-oriented ``ReadableStream`` that may be used to read the file's contents.

An error will be thrown if this method is called more than once or is called after the ``FileHandle`` is closed or closing.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const {
     open,
   } = require('node:fs/promises');

   (async () => {
     const file = await open('./some/file/to/read');

     for await (const chunk of file.readableWebStream())
       console.log(chunk);

     await file.close();
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadablewebstreamoptions>`__

promises.FileHandle.readFile
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.readFile(options)

*Kind:* function · *Arguments:* 1

Asynchronously reads the entire contents of a file.

If ``options`` is a string, then it specifies the ``encoding``.

**Arguments**

* ``options`` (*Object|string*)

**Example**

.. code-block:: javascript

   import { Buffer } from 'node:buffer';
   import { open } from 'node:fs/promises';

   const file = await open('./some/file/to/read');
   try {
     const buf = Buffer.alloc(16384);
     const contents = await file.readFile({ buffer: buf });
     console.log(contents); // A view over `buf` containing only the bytes read
   } finally {
     await file.close();
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadfileoptions>`__

promises.FileHandle.readLines
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.readLines([options])

*Kind:* function · *Arguments:* 1

Convenience method to create a ``readline`` interface and stream over the file. See ``filehandle.createReadStream()`` for the options.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { open } = require('node:fs/promises');

   (async () => {
     const file = await open('./some/file/to/read');

     for await (const line of file.readLines()) {
       console.log(line);
     }
   })();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadlinesoptions>`__

promises.FileHandle.readv
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.readv(buffers[, position])

*Kind:* function · *Arguments:* 2

Read from a file and write to an array of ``ArrayBufferView``\ s

**Arguments**

* ``buffers`` (*Buffer\[]|TypedArray\[]|DataView\[]*)
* ``position`` (*integer|null*) — optional: The offset from the beginning of the file where

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadvbuffers-position>`__

promises.FileHandle.stat
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.stat([options])

*Kind:* function · *Arguments:* 1

**Arguments**

* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlestatoptions>`__

promises.FileHandle.sync
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.sync()

*Kind:* function · *Arguments:* 0

Request that all data for the open file descriptor is flushed to the storage device. The specific implementation is operating system and device specific. Refer to the POSIX fsync(2) documentation for more detail.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlesync>`__

promises.FileHandle.truncate
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.truncate(len)

*Kind:* function · *Arguments:* 1

Truncates the file.

If the file was larger than ``len`` bytes, only the first ``len`` bytes will be retained in the file.

**Arguments**

* ``len`` (*integer*): **Default:** ``0``

**Example**

.. code-block:: javascript

   import { open } from 'node:fs/promises';

   let filehandle = null;
   try {
     filehandle = await open('temp.txt', 'r+');
     await filehandle.truncate(4);
   } finally {
     await filehandle?.close();
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandletruncatelen>`__

promises.FileHandle.utimes
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.utimes(atime, mtime)

*Kind:* function · *Arguments:* 2

Change the file system timestamps of the object referenced by the ``FileHandle`` then fulfills the promise with no arguments upon success.

**Arguments**

* ``atime`` (*number|string|Date*)
* ``mtime`` (*number|string|Date*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandleutimesatime-mtime>`__

promises.FileHandle.write
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.write(buffer, offset[, length[, position]])

*Kind:* function · *Arguments:* 4

Write ``buffer`` to the file.

The promise is fulfilled with an object containing two properties:

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView*): a reference to the
* ``offset`` (*integer*): The start position from within ``buffer`` where the data
* ``length`` (*integer*) — optional: The number of bytes from ``buffer`` to write. **Default:**
* ``position`` (*integer|null*) — optional: The offset from the beginning of the file where the

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlewritebuffer-offset-length-position>`__

promises.FileHandle.writeFile
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.writeFile(data, options)

*Kind:* function · *Arguments:* 2

Asynchronously writes data to a file, replacing the file if it already exists. ``data`` can be a string, a buffer, an ``AsyncIterable``, or an ``Iterable`` object. The promise is fulfilled with no arguments upon success.

If ``options`` is a string, then it specifies the ``encoding``.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView|AsyncIterable|Iterable*)
* ``options`` (*Object|string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlewritefiledata-options>`__

promises.FileHandle.writer
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.writer([options])

*Kind:* function · *Arguments:* 1

Return a ``node:stream/iter`` writer backed by this file handle.

The writer supports both ``Symbol.asyncDispose`` and ``Symbol.dispose``:

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { open } = require('node:fs/promises');
   const { from, pipeTo } = require('node:stream/iter');
   const { compressGzip } = require('node:zlib/iter');

   async function run() {
     // Async pipeline
     const fh = await open('output.gz', 'w');
     await pipeTo(from('Hello!'), compressGzip(), fh.writer({ autoClose: true }));

     // Sync pipeline with limit
     const src = await open('input.txt', 'r');
     const dst = await open('output.txt', 'w');
     const w = dst.writer({ limit: 1024 * 1024 }); // Max 1 MB
     await pipeTo(src.pull({ autoClose: true }), w);
     await w.end();
     await dst.close();
   }

   run().catch(console.error);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlewriteroptions>`__

promises.FileHandle.writev
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.writev(buffers[, position])

*Kind:* function · *Arguments:* 2

Write an array of ``ArrayBufferView``\ s to the file.

The promise is fulfilled with an object containing a two properties:

**Arguments**

* ``buffers`` (*Buffer\[]|TypedArray\[]|DataView\[]*): a reference to the ``buffers``
* ``position`` (*integer|null*) — optional: The offset from the beginning of the file where the

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlewritevbuffers-position>`__

ReadStream class
----------------

**ReadStream** — Instances of ``fs.ReadStream`` cannot be constructed directly. They are created and returned using the ``fs.createReadStream()`` function.

ReadStream._finishRead
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.ReadStream._finishRead(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ReadStream._pumpRead
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.ReadStream._pumpRead(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ReadStream._read
~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.ReadStream._read(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ReadStream.close
~~~~~~~~~~~~~~~~

.. code-block:: text

   dir.close(callback)

*Kind:* function · *Arguments:* 1

Asynchronously close the directory's underlying resource handle. Subsequent reads will result in errors.

The ``callback`` will be called after the resource handle has been closed.

**Arguments**

* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#dirclosecallback>`__

ReadStream.destroy
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.destroy()

*Kind:* function · *Arguments:* 1

Close the stream immediately, without flushing the internal buffer.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamdestroy>`__

ReadStream.on
~~~~~~~~~~~~~

.. code-block:: text

   fs.ReadStream.on(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ReadStream.open
~~~~~~~~~~~~~~~

.. code-block:: text

   fsPromises.open(path, flags[, mode])

*Kind:* function · *Arguments:* 3

Opens a ``FileHandle``.

Refer to the POSIX open(2) documentation for more detail.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``flags`` (*string|number*): See support of file system ``flags``.
* ``mode`` (*string|integer*) — optional: Sets the file mode (permission and sticky bits)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#fspromisesopenpath-flags-mode>`__

ReadStream.read
~~~~~~~~~~~~~~~

.. code-block:: text

   filehandle.read([options])

*Kind:* function · *Arguments:* 1

Reads data from the file and stores that in the given buffer.

If the file is not modified concurrently, the end-of-file is reached when the number of bytes read is zero.

**Arguments**

* ``options`` (*Object*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#filehandlereadoptions>`__

ReadStream.resume
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.ReadStream.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Stats class
-----------

**Stats** — Type: Runtime

Calling ``fs.Stats`` class directly with ``Stats()`` or ``new Stats()`` is deprecated due to being internals, not intended for public use.

Stats.isBlockDevice
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stats.isBlockDevice()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a block device.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsisblockdevice>`__

Stats.isCharacterDevice
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stats.isCharacterDevice()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a character device.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsischaracterdevice>`__

Stats.isDirectory
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stats.isDirectory()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a file system directory.

If the ``fs.Stats`` object was obtained from calling ``fs.lstat()`` on a symbolic link which resolves to a directory, this method will return ``false``. This is because ``fs.lstat()`` returns information about a symbolic link itself and not the path it resolves to.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsisdirectory>`__

Stats.isFIFO
~~~~~~~~~~~~

.. code-block:: text

   stats.isFIFO()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a first-in-first-out (FIFO) pipe.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsisfifo>`__

Stats.isFile
~~~~~~~~~~~~

.. code-block:: text

   stats.isFile()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a regular file.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsisfile>`__

Stats.isSocket
~~~~~~~~~~~~~~

.. code-block:: text

   stats.isSocket()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a socket.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsissocket>`__

Stats.isSymbolicLink
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stats.isSymbolicLink()

*Kind:* function · *Arguments:* 0

Returns ``true`` if the ``fs.Stats`` object describes a symbolic link.

This method is only valid when using ``fs.lstat()``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#statsissymboliclink>`__

StatWatcher class
-----------------

**StatWatcher** — A successful call to ``fs.watchFile()`` method will return a new ``fs.StatWatcher`` object.

StatWatcher.ref
~~~~~~~~~~~~~~~

.. code-block:: text

   watcher.ref()

*Kind:* function · *Arguments:* 0

When called, requests that the Node.js event loop *not* exit so long as the ``fs.FSWatcher`` is active. Calling ``watcher.ref()`` multiple times will have no effect.

By default, all ``fs.FSWatcher`` objects are "ref'ed", making it normally unnecessary to call ``watcher.ref()`` unless ``watcher.unref()`` had been called previously.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#watcherref>`__

StatWatcher.unref
~~~~~~~~~~~~~~~~~

.. code-block:: text

   watcher.unref()

*Kind:* function · *Arguments:* 0

When called, the active ``fs.FSWatcher`` object will not require the Node.js event loop to remain active. If there is no other activity keeping the event loop running, the process may exit before the ``fs.FSWatcher`` object's callback is invoked. Calling ``watcher.unref()`` multiple times will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#watcherunref>`__

Utf8Stream class
----------------

**Utf8Stream** — An optimized UTF-8 stream writer that allows for flushing all the internal buffering on demand. It handles ``EAGAIN`` errors correctly, allowing for customization, for example, by dropping content if the disk is busy.

Utf8Stream._drain
~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.Utf8Stream._drain(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Utf8Stream._open
~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.Utf8Stream._open(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Utf8Stream.destroy
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.destroy()

*Kind:* function · *Arguments:* 0

Close the stream immediately, without flushing the internal buffer.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamdestroy>`__

Utf8Stream.end
~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.end()

*Kind:* function · *Arguments:* 1

Close the stream gracefully, flushing the internal buffer before closing.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamend>`__

Utf8Stream.flush
~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.flush(callback)

*Kind:* function · *Arguments:* 1

Writes the current buffer to the file if a write was not in progress. Do nothing if ``minLength`` is zero or if it is already writing.

**Arguments**

* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamflushcallback>`__

Utf8Stream.flushSync
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.flushSync()

*Kind:* function · *Arguments:* 0

Flushes the buffered data synchronously. This is a costly operation.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamflushsync>`__

Utf8Stream.reopen
~~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.reopen(file)

*Kind:* function · *Arguments:* 1

Reopen the file in place, useful for log rotation.

**Arguments**

* ``file``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamreopenfile>`__

Utf8Stream.write
~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.write(data)

*Kind:* function · *Arguments:* 1

When the ``options.contentMode`` is set to ``'utf8'`` when the stream is created, the ``data`` argument must be a string. If the ``contentMode`` is set to ``'buffer'``, the ``data`` argument must be a ``Buffer``.

**Arguments**

* ``data`` (*string|Buffer*): The data to write.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamwritedata>`__

WriteStream class
-----------------

**WriteStream** — Instances of ``fs.WriteStream`` cannot be constructed directly. They are created and returned using the ``fs.createWriteStream()`` function.

WriteStream._close
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.WriteStream._close(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


WriteStream._write
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   fs.WriteStream._write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


WriteStream.close
~~~~~~~~~~~~~~~~~

.. code-block:: text

   writeStream.close([callback])

*Kind:* function · *Arguments:* 1

Closes ``writeStream``. Optionally accepts a callback that will be executed once the ``writeStream`` is closed.

**Arguments**

* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#writestreamclosecallback>`__

WriteStream.destroy
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.destroy()

*Kind:* function · *Arguments:* 1

Close the stream immediately, without flushing the internal buffer.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamdestroy>`__

WriteStream.end
~~~~~~~~~~~~~~~

.. code-block:: text

   utf8Stream.end()

*Kind:* function · *Arguments:* 3

Close the stream gracefully, flushing the internal buffer before closing.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/fs.html#utf8streamend>`__

Properties and constants
------------------------

* ``constants`` — object
* ``constants.COPYFILE_EXCL`` — number
* ``constants.COPYFILE_FICLONE`` — number
* ``constants.COPYFILE_FICLONE_FORCE`` — number
* ``constants.F_OK`` — number
* ``constants.O_APPEND`` — number
* ``constants.O_CREAT`` — number
* ``constants.O_DIRECTORY`` — number
* ``constants.O_DSYNC`` — number
* ``constants.O_EXCL`` — number
* ``constants.O_NOCTTY`` — number
* ``constants.O_NOFOLLOW`` — number
* ``constants.O_NONBLOCK`` — number
* ``constants.O_RDONLY`` — number
* ``constants.O_RDWR`` — number
* ``constants.O_SYNC`` — number
* ``constants.O_TRUNC`` — number
* ``constants.O_WRONLY`` — number
* ``constants.R_OK`` — number
* ``constants.S_IFBLK`` — number
* ``constants.S_IFCHR`` — number
* ``constants.S_IFDIR`` — number
* ``constants.S_IFIFO`` — number
* ``constants.S_IFLNK`` — number
* ``constants.S_IFMT`` — number
* ``constants.S_IFREG`` — number
* ``constants.S_IFSOCK`` — number
* ``constants.S_IRGRP`` — number
* ``constants.S_IROTH`` — number
* ``constants.S_IRUSR`` — number
* ``constants.S_IRWXG`` — number
* ``constants.S_IRWXO`` — number
* ``constants.S_IRWXU`` — number
* ``constants.S_IWGRP`` — number
* ``constants.S_IWOTH`` — number
* ``constants.S_IWUSR`` — number
* ``constants.S_IXGRP`` — number
* ``constants.S_IXOTH`` — number
* ``constants.S_IXUSR`` — number
* ``constants.UV_FS_COPYFILE_EXCL`` — number
* ``constants.UV_FS_COPYFILE_FICLONE`` — number
* ``constants.UV_FS_COPYFILE_FICLONE_FORCE`` — number
* ``constants.W_OK`` — number
* ``constants.X_OK`` — number
* ``F_OK`` — number
* ``native`` — object
* ``promises`` — object
* ``promises.constants`` — object
* ``promises.constants.COPYFILE_EXCL`` — number
* ``promises.constants.COPYFILE_FICLONE`` — number
* ``promises.constants.COPYFILE_FICLONE_FORCE`` — number
* ``promises.constants.F_OK`` — number
* ``promises.constants.O_APPEND`` — number
* ``promises.constants.O_CREAT`` — number
* ``promises.constants.O_DIRECTORY`` — number
* ``promises.constants.O_DSYNC`` — number
* ``promises.constants.O_EXCL`` — number
* ``promises.constants.O_NOCTTY`` — number
* ``promises.constants.O_NOFOLLOW`` — number
* ``promises.constants.O_NONBLOCK`` — number
* ``promises.constants.O_RDONLY`` — number
* ``promises.constants.O_RDWR`` — number
* ``promises.constants.O_SYNC`` — number
* ``promises.constants.O_TRUNC`` — number
* ``promises.constants.O_WRONLY`` — number
* ``promises.constants.R_OK`` — number
* ``promises.constants.S_IFBLK`` — number
* ``promises.constants.S_IFCHR`` — number
* ``promises.constants.S_IFDIR`` — number
* ``promises.constants.S_IFIFO`` — number
* ``promises.constants.S_IFLNK`` — number
* ``promises.constants.S_IFMT`` — number
* ``promises.constants.S_IFREG`` — number
* ``promises.constants.S_IFSOCK`` — number
* ``promises.constants.S_IRGRP`` — number
* ``promises.constants.S_IROTH`` — number
* ``promises.constants.S_IRUSR`` — number
* ``promises.constants.S_IRWXG`` — number
* ``promises.constants.S_IRWXO`` — number
* ``promises.constants.S_IRWXU`` — number
* ``promises.constants.S_IWGRP`` — number
* ``promises.constants.S_IWOTH`` — number
* ``promises.constants.S_IWUSR`` — number
* ``promises.constants.S_IXGRP`` — number
* ``promises.constants.S_IXOTH`` — number
* ``promises.constants.S_IXUSR`` — number
* ``promises.constants.UV_FS_COPYFILE_EXCL`` — number
* ``promises.constants.UV_FS_COPYFILE_FICLONE`` — number
* ``promises.constants.UV_FS_COPYFILE_FICLONE_FORCE`` — number
* ``promises.constants.W_OK`` — number
* ``promises.constants.X_OK`` — number
* ``R_OK`` — number
* ``ReadStream.readable`` — boolean
* ``W_OK`` — number
* ``WriteStream.writable`` — boolean
* ``X_OK`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
