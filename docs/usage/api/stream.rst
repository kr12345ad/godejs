stream
======

A stream is an abstract interface for working with streaming data in Node.js. The ``node:stream`` module provides an API for implementing the stream interface.

There are many stream objects provided by Node.js. For instance, a request to an HTTP server and ``process.stdout`` are both stream instances.

Loading
-------

.. code-block:: javascript

   const stream = require('node:stream');
   // ESM
   import * as stream from 'node:stream';

Quick start
-----------

.. code-block:: javascript

   const stream = require('node:stream');

API
---

_isUint8Array
~~~~~~~~~~~~~

.. code-block:: text

   stream._isUint8Array(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   console.log(typeof stream._isUint8Array); // 'function'
   console.log(stream._isUint8Array.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


addAbortSignal
~~~~~~~~~~~~~~

.. code-block:: text

   stream.addAbortSignal(signal, stream)

*Kind:* function · *Arguments:* 2

Attaches an AbortSignal to a readable or writable stream. This lets code control stream destruction using an ``AbortController``.

Calling ``abort`` on the ``AbortController`` corresponding to the passed ``AbortSignal`` will behave the same way as calling ``.destroy(new AbortError())`` on the stream, and ``controller.error(new AbortError())`` for webstreams.

**Arguments**

* ``signal`` (*AbortSignal*): A signal representing possible cancellation
* ``stream`` (*Stream|ReadableStream|WritableStream*): A stream to attach a signal

**Example**

.. code-block:: javascript

   const fs = require('node:fs');

   const controller = new AbortController();
   const read = addAbortSignal(
     controller.signal,
     fs.createReadStream(('object.json')),
   );
   // Later, abort the operation closing the stream
   controller.abort();


addAbortSignalNoValidate
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.addAbortSignalNoValidate(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.addAbortSignalNoValidate(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


compose
~~~~~~~

.. code-block:: text

   stream.compose(...streams)

*Kind:* function · *Arguments:* 1

Combines two or more streams into a ``Duplex`` stream that writes to the first stream and reads from the last. Each provided stream is piped into the next, using ``stream.pipeline``. If any of the streams error then all are destroyed, including the outer ``Duplex`` stream.

Because ``stream.compose`` returns a new stream that in turn can (and should) be piped into other streams, it enables composition. In contrast, when passing streams to ``stream.pipeline``, typically the first stream is a readable stream and the last a writable stream, forming a closed circuit.

**Arguments**

* ``...streams``

**Example**

.. code-block:: javascript

   import { compose, Transform } from 'node:stream';

   const removeSpaces = new Transform({
     transform(chunk, encoding, callback) {
       callback(null, String(chunk).replace(' ', ''));
     },
   });

   async function* toUpper(source) {
     for await (const chunk of source) {
       yield String(chunk).toUpperCase();
     }
   }

   let res = '';
   for await (const buf of compose(removeSpaces, toUpper).end('hello world')) {
     res += buf;
   }

   console.log(res); // prints 'HELLOWORLD'


destroy
~~~~~~~

.. code-block:: text

   stream.destroy([error[, options]])

*Kind:* function · *Arguments:* 2

Immediately and abruptly destroys the stream. If ``error`` is provided and ``stream.onerror`` is set, the ``onerror`` callback is invoked before destruction. The ``stream.closed`` promise rejects with the error.

When the stream is destroyed with an ``error`` (or with an explicit ``options.code``), the QUIC stack signals the abort to the peer:

**Arguments**

* ``error`` (*any*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.destroy(new Error('example'), {});
   console.log(result);


Duplex
~~~~~~

.. code-block:: text

   stream.Duplex

*Kind:* class · *Arguments:* 1

Duplex streams are streams that implement both the ``Readable`` and ``Writable`` interfaces.

Examples of ``Duplex`` streams include:

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex(undefined);
   console.log(instance);


duplexPair
~~~~~~~~~~

.. code-block:: text

   stream.duplexPair([options])

*Kind:* function · *Arguments:* 1

The utility function ``duplexPair`` returns an Array with two items, each being a ``Duplex`` stream connected to the other side:

Whatever is written to one stream is made readable on the other. It provides behavior analogous to a network connection, where the data written by the client becomes readable by the server, and vice-versa.

**Arguments**

* ``options`` (*Object*) — optional: A value to pass to both ``Duplex`` constructors,

**Example**

.. code-block:: javascript

   const [ sideA, sideB ] = duplexPair();


finished
~~~~~~~~

.. code-block:: text

   stream.finished(stream[, options])

*Kind:* function · *Arguments:* 2

The ``finished`` API also provides a callback version.

``stream.finished()`` leaves dangling event listeners (in particular ``'error'``, ``'end'``, ``'finish'`` and ``'close'``) after the returned promise is resolved or rejected. The reason for this is so that unexpected ``'error'`` events (due to incorrect stream implementations) do not cause unexpected crashes. If this is unwanted behavior then ``options.cleanup`` should be set to ``true``:

**Arguments**

* ``stream`` (*Stream|ReadableStream|WritableStream*): A readable and/or writable
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { finished } = require('node:stream/promises');
   const fs = require('node:fs');

   const rs = fs.createReadStream('archive.tar');

   async function run() {
     await finished(rs);
     console.log('Stream is done reading.');
   }

   run().catch(console.error);
   rs.resume(); // Drain the stream.


from
~~~~

.. code-block:: text

   stream.Readable.from(iterable[, options])

*Kind:* function · *Arguments:* 2

A utility method for creating readable streams out of iterators.

Calling ``Readable.from(string)`` or ``Readable.from(buffer)`` will not have the strings or buffers be iterated to match the other streams semantics for performance reasons.

**Arguments**

* ``iterable`` (*Iterable*): Object implementing the ``Symbol.asyncIterator`` or
* ``options`` (*Object*) — optional: Options provided to ``new stream.Readable([options])``.

**Example**

.. code-block:: javascript

   const { Readable } = require('node:stream');

   async function * generate() {
     yield 'hello';
     yield 'streams';
   }

   const readable = Readable.from(generate());

   readable.on('data', (chunk) => {
     console.log(chunk);
   });


getDefaultHighWaterMark
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.getDefaultHighWaterMark(objectMode)

*Kind:* function · *Arguments:* 1

Returns the default highWaterMark used by streams. Defaults to ``16`` for ``objectMode``. For byte streams, it defaults to ``65536`` (64 KiB) on non-Windows platforms and ``16384`` (16 KiB) on Windows.

**Arguments**

* ``objectMode`` (*boolean*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.getDefaultHighWaterMark(true);
   console.log(result);


isDestroyed
~~~~~~~~~~~

.. code-block:: text

   stream.isDestroyed(stream)

*Kind:* function · *Arguments:* 1

Returns whether the stream has been destroyed.

**Arguments**

* ``stream`` (*Readable|Writable|Duplex*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.isDestroyed(process.stdout);
   console.log(result);


isDisturbed
~~~~~~~~~~~

.. code-block:: text

   stream.Readable.isDisturbed(stream)

*Kind:* function · *Arguments:* 1

Returns whether the stream has been read from or cancelled.

**Arguments**

* ``stream`` (*stream.Readable|ReadableStream*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.isDisturbed(process.stdout);
   console.log(result);


isErrored
~~~~~~~~~

.. code-block:: text

   stream.isErrored(stream)

*Kind:* function · *Arguments:* 1

Returns whether the stream has encountered an error.

**Arguments**

* ``stream`` (*Readable|Writable|Duplex|WritableStream|ReadableStream*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.isErrored(process.stdout);
   console.log(result);


isReadable
~~~~~~~~~~

.. code-block:: text

   stream.isReadable(stream)

*Kind:* function · *Arguments:* 1

Returns whether the stream is readable.

**Arguments**

* ``stream`` (*Readable|Duplex|ReadableStream*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.isReadable(process.stdout);
   console.log(result);


isUint8Array
~~~~~~~~~~~~

.. code-block:: text

   util.types.isUint8Array(value)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the value is a built-in ``Uint8Array`` instance.

**Arguments**

* ``value`` (*any*)

**Example**

.. code-block:: javascript

   util.types.isUint8Array(new ArrayBuffer());  // Returns false
   util.types.isUint8Array(new Uint8Array());  // Returns true
   util.types.isUint8Array(new Float64Array());  // Returns false


isWritable
~~~~~~~~~~

.. code-block:: text

   stream.isWritable(stream)

*Kind:* function · *Arguments:* 1

Returns whether the stream is writable.

**Arguments**

* ``stream`` (*Writable|Duplex|WritableStream*)

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.isWritable(process.stdout);
   console.log(result);


PassThrough
~~~~~~~~~~~

.. code-block:: text

   stream.PassThrough

*Kind:* class · *Arguments:* 1

The ``stream.PassThrough`` class is a trivial implementation of a ``Transform`` stream that simply passes the input bytes across to the output. Its purpose is primarily for examples and testing, but there are some use cases where ``stream.PassThrough`` is useful as a building block for novel sorts of streams.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.PassThrough(undefined);
   console.log(instance);


pipeline
~~~~~~~~

.. code-block:: text

   stream.pipeline(streams[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``streams``
* ``options`` — optional

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.pipeline(undefined, {});
   console.log(result);


push
~~~~

.. code-block:: text

   readable.push(chunk[, encoding])

*Kind:* function · *Arguments:* 2

When ``chunk`` is a ``Buffer``, ``TypedArray``, ``DataView`` or ``string``, the ``chunk`` of data will be added to the internal queue for users of the stream to consume. Passing ``chunk`` as ``null`` signals the end of the stream (EOF), after which no more data can be written.

When the ``Readable`` is operating in paused mode, the data added with ``readable.push()`` can be read out by calling the ``readable.read()`` method when the ``'readable'`` event is emitted.

**Arguments**

* ``chunk`` (*Buffer|TypedArray|DataView|string|null|any*): Chunk of data to push
* ``encoding`` (*string*) — optional: Encoding of string chunks. Must be a valid

**Example**

.. code-block:: javascript

   // `_source` is an object with readStop() and readStart() methods,
   // and an `ondata` member that gets called when it has data, and
   // an `onend` member that gets called when the data is over.

   class SourceWrapper extends Readable {
     constructor(options) {
       super(options);

       this._source = getLowLevelSourceObject();

       // Every time there's data, push it into the internal buffer.
       this._source.ondata = (chunk) => {
         // If push() returns false, then stop reading from source.
         if (!this.push(chunk))
           this._source.readStop();
       };

       // When the source ends, push the EOF-signaling `null` chunk.
       this._source.onend = () => {
         this.push(null);
       };
     }
     // _read() will be called when the stream wants to pull more data in.
     // The advisory size argument is ignored in this case.
     _read(size) {
       this._source.readStart();
     }
   }


read
~~~~

.. code-block:: text

   readable.read(0)

*Kind:* function · *Arguments:* 1

There are some cases where it is necessary to trigger a refresh of the underlying readable stream mechanisms, without actually consuming any data. In such cases, it is possible to call ``readable.read(0)``, which will always return ``null``.

If the internal read buffer is below the ``highWaterMark``, and the stream is not currently reading, then calling ``stream.read(0)`` will trigger a low-level ``stream._read()`` call.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.read(undefined);
   console.log(result);


Readable
~~~~~~~~

.. code-block:: text

   stream.Readable

*Kind:* class · *Arguments:* 1

##### Event: ``'close'``

The ``'close'`` event is emitted when the stream and any of its underlying resources (a file descriptor, for example) have been closed. The event indicates that no more events will be emitted, and no further computation will occur.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Readable } = require('node:stream');
   const { text, from } = require('node:stream/iter');

   async function run() {
     const readable = new Readable({
       read() { this.push('hello'); this.push(null); },
     });

     console.log(await text(from(readable))); // 'hello'
   }

   run().catch(console.error);


setDefaultHighWaterMark
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.setDefaultHighWaterMark(objectMode, value)

*Kind:* function · *Arguments:* 2

Sets the default highWaterMark used by streams.

**Arguments**

* ``objectMode`` (*boolean*)
* ``value`` (*integer*): highWaterMark value

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.setDefaultHighWaterMark(true, 'hello');
   console.log(result);


Stream
~~~~~~

.. code-block:: text

   stream.Stream(arg0)

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Stream(undefined);
   console.log(instance);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Transform
~~~~~~~~~

.. code-block:: text

   stream.Transform

*Kind:* class · *Arguments:* 1

Transform streams are ``Duplex`` streams where the output is in some way related to the input. Like all ``Duplex`` streams, ``Transform`` streams implement both the ``Readable`` and ``Writable`` interfaces.

Examples of ``Transform`` streams include:

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Transform(undefined);
   console.log(instance);


Writable
~~~~~~~~

.. code-block:: text

   stream.Writable

*Kind:* class · *Arguments:* 1

##### Event: ``'close'``

The ``'close'`` event is emitted when the stream and any of its underlying resources (a file descriptor, for example) have been closed. The event indicates that no more events will be emitted, and no further computation will occur.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Writable } = require('node:stream');

   const myStream = new Writable();

   const fooErr = new Error('foo error');
   myStream.destroy(fooErr);
   myStream.on('error', (fooErr) => console.error(fooErr.message)); // foo error


Duplex class
------------

**Duplex** — Duplex streams are streams that implement both the ``Readable`` and ``Writable`` interfaces.

Examples of ``Duplex`` streams include:

Duplex._destroy
~~~~~~~~~~~~~~~

.. code-block:: text

   writable._destroy(err, callback)

*Kind:* function · *Arguments:* 2

The ``_destroy()`` method is called by ``writable.destroy()``. It can be overridden by child classes but it **must not** be called directly.

**Arguments**

* ``err`` (*Error*): A possible error.
* ``callback`` (*Function*): A callback function that takes an optional error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance._destroy(new Error('example'), () => {});
   console.log(result);


Duplex._undestroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex._undestroy(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   console.log(typeof stream.Duplex.prototype._undestroy()); // 'function'
   console.log(stream.Duplex.prototype._undestroy().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex._write
~~~~~~~~~~~~~

.. code-block:: text

   writable._write(chunk, encoding, callback)

*Kind:* function · *Arguments:* 3

All ``Writable`` stream implementations must provide a ``writable._write()`` and/or ``writable._writev()`` method to send data to the underlying resource.

``Transform`` streams provide their own implementation of the ``writable._write()``.

**Arguments**

* ``chunk`` (*Buffer|string|any*): The ``Buffer`` to be written, converted from the
* ``encoding`` (*string*): If the chunk is a string, then ``encoding`` is the
* ``callback`` (*Function*): Call this function (optionally with an error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance._write(Buffer.from('hello'), 'utf8', () => {});
   console.log(result);


Duplex.cork
~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.cork(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.cork();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.destroy
~~~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.destroy(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.destroy(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.destroySoon
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.destroySoon(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.destroySoon();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.end
~~~~~~~~~~

.. code-block:: text

   stream.Duplex.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.end(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.setDefaultEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.setDefaultEncoding(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.setDefaultEncoding(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.uncork
~~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.uncork(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.uncork();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Duplex.write
~~~~~~~~~~~~

.. code-block:: text

   stream.Duplex.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Duplex();
   const result = instance.write(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


PassThrough class
-----------------

**PassThrough** — The ``stream.PassThrough`` class is a trivial implementation of a ``Transform`` stream that simply passes the input bytes across to the output. Its purpose is primarily for examples and testing, but there are some use cases where ``stream.PassThrough`` is useful as a building block for novel sorts of streams.

PassThrough._transform
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   transform._transform(chunk, encoding, callback)

*Kind:* function · *Arguments:* 3

This function MUST NOT be called by application code directly. It should be implemented by child classes, and called by the internal ``Readable`` class methods only.

All ``Transform`` stream implementations must provide a ``_transform()`` method to accept input and produce output. The ``transform._transform()`` implementation handles the bytes being written, computes an output, then passes that output off to the readable portion using the ``transform.push()`` method.

**Arguments**

* ``chunk`` (*Buffer|string|any*): The ``Buffer`` to be transformed, converted from
* ``encoding`` (*string*): If the chunk is a string, then this is the
* ``callback`` (*Function*): A callback function (optionally with an error

**Example**

.. code-block:: javascript

   transform.prototype._transform = function(data, encoding, callback) {
     this.push(data);
     callback();
   };

   transform.prototype._transform = function(data, encoding, callback) {
     callback(null, data);
   };


promises
--------

promises.finished
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.finished(stream[, options])

*Kind:* function · *Arguments:* 2

The ``finished`` API also provides a callback version.

``stream.finished()`` leaves dangling event listeners (in particular ``'error'``, ``'end'``, ``'finish'`` and ``'close'``) after the returned promise is resolved or rejected. The reason for this is so that unexpected ``'error'`` events (due to incorrect stream implementations) do not cause unexpected crashes. If this is unwanted behavior then ``options.cleanup`` should be set to ``true``:

**Arguments**

* ``stream`` (*Stream|ReadableStream|WritableStream*): A readable and/or writable
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { finished } = require('node:stream/promises');
   const fs = require('node:fs');

   const rs = fs.createReadStream('archive.tar');

   async function run() {
     await finished(rs);
     console.log('Stream is done reading.');
   }

   run().catch(console.error);
   rs.resume(); // Drain the stream.


promises.pipeline
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.pipeline(streams[, options])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``streams``
* ``options`` — optional

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const result = stream.promises.pipeline(undefined, {});
   console.log(result);


Readable class
--------------

**Readable** — ##### Event: ``'close'``

The ``'close'`` event is emitted when the stream and any of its underlying resources (a file descriptor, for example) have been closed. The event indicates that no more events will be emitted, and no further computation will occur.

Readable._destroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   readable._destroy(err, callback)

*Kind:* function · *Arguments:* 2

The ``_destroy()`` method is called by ``readable.destroy()``. It can be overridden by child classes but it **must not** be called directly.

**Arguments**

* ``err`` (*Error*): A possible error.
* ``callback`` (*Function*): A callback function that takes an optional error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance._destroy(new Error('example'), () => {});
   console.log(result);


Readable._read
~~~~~~~~~~~~~~

.. code-block:: text

   readable._read(size)

*Kind:* function · *Arguments:* 1

This function MUST NOT be called by application code directly. It should be implemented by child classes, and called by the internal ``Readable`` class methods only.

All ``Readable`` stream implementations must provide an implementation of the ``readable._read()`` method to fetch data from the underlying resource.

**Arguments**

* ``size`` (*number*): Number of bytes to read asynchronously

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance._read(0);
   console.log(result);


Readable._undestroy
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable._undestroy(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   console.log(typeof stream.Readable.prototype._undestroy()); // 'function'
   console.log(stream.Readable.prototype._undestroy().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.addListener
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.addListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.addListener(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.asIndexedPairs
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.asIndexedPairs(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.asIndexedPairs();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.compose
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.compose(...streams)

*Kind:* function · *Arguments:* 1

Combines two or more streams into a ``Duplex`` stream that writes to the first stream and reads from the last. Each provided stream is piped into the next, using ``stream.pipeline``. If any of the streams error then all are destroyed, including the outer ``Duplex`` stream.

Because ``stream.compose`` returns a new stream that in turn can (and should) be piped into other streams, it enables composition. In contrast, when passing streams to ``stream.pipeline``, typically the first stream is a readable stream and the last a writable stream, forming a closed circuit.

**Arguments**

* ``...streams``

**Example**

.. code-block:: javascript

   import { compose, Transform } from 'node:stream';

   const removeSpaces = new Transform({
     transform(chunk, encoding, callback) {
       callback(null, String(chunk).replace(' ', ''));
     },
   });

   async function* toUpper(source) {
     for await (const chunk of source) {
       yield String(chunk).toUpperCase();
     }
   }

   let res = '';
   for await (const buf of compose(removeSpaces, toUpper).end('hello world')) {
     res += buf;
   }

   console.log(res); // prints 'HELLOWORLD'


Readable.destroy
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.destroy(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.destroy(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.drop
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.drop(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.drop(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.every
~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.every(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.every(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.filter
~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.filter(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.filter(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.find
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.find(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.find(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.flatMap
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.flatMap(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.flatMap(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.forEach
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.forEach(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.forEach(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.isPaused
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.isPaused(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.isPaused();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.iterator
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.iterator(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.iterator(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.map
~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.map(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.map(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.on
~~~~~~~~~~~

.. code-block:: text

   stream.Readable.on(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.on(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.once
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.once(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.once(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.pause
~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.pause(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.pause();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.pipe
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.pipe(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.pipe(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.push
~~~~~~~~~~~~~

.. code-block:: text

   readable.push(chunk[, encoding])

*Kind:* function · *Arguments:* 2

When ``chunk`` is a ``Buffer``, ``TypedArray``, ``DataView`` or ``string``, the ``chunk`` of data will be added to the internal queue for users of the stream to consume. Passing ``chunk`` as ``null`` signals the end of the stream (EOF), after which no more data can be written.

When the ``Readable`` is operating in paused mode, the data added with ``readable.push()`` can be read out by calling the ``readable.read()`` method when the ``'readable'`` event is emitted.

**Arguments**

* ``chunk`` (*Buffer|TypedArray|DataView|string|null|any*): Chunk of data to push
* ``encoding`` (*string*) — optional: Encoding of string chunks. Must be a valid

**Example**

.. code-block:: javascript

   // `_source` is an object with readStop() and readStart() methods,
   // and an `ondata` member that gets called when it has data, and
   // an `onend` member that gets called when the data is over.

   class SourceWrapper extends Readable {
     constructor(options) {
       super(options);

       this._source = getLowLevelSourceObject();

       // Every time there's data, push it into the internal buffer.
       this._source.ondata = (chunk) => {
         // If push() returns false, then stop reading from source.
         if (!this.push(chunk))
           this._source.readStop();
       };

       // When the source ends, push the EOF-signaling `null` chunk.
       this._source.onend = () => {
         this.push(null);
       };
     }
     // _read() will be called when the stream wants to pull more data in.
     // The advisory size argument is ignored in this case.
     _read(size) {
       this._source.readStart();
     }
   }


Readable.read
~~~~~~~~~~~~~

.. code-block:: text

   readable.read(0)

*Kind:* function · *Arguments:* 1

There are some cases where it is necessary to trigger a refresh of the underlying readable stream mechanisms, without actually consuming any data. In such cases, it is possible to call ``readable.read(0)``, which will always return ``null``.

If the internal read buffer is below the ``highWaterMark``, and the stream is not currently reading, then calling ``stream.read(0)`` will trigger a low-level ``stream._read()`` call.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.read(undefined);
   console.log(result);


Readable.reduce
~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.reduce(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.reduce(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.resume
~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.resume(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.resume();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.setEncoding
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.setEncoding(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.some
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.some(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.some(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.take
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.take(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.take(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.toArray
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.toArray(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.toArray();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.unpipe
~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.unpipe(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.unpipe(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.unshift
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.unshift(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.unshift(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Readable.wrap
~~~~~~~~~~~~~

.. code-block:: text

   stream.Readable.wrap(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Readable();
   const result = instance.wrap(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Stream class
------------

Stream.off
~~~~~~~~~~

.. code-block:: text

   stream.Stream.off(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Stream();
   const result = instance.off();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Stream.pipe
~~~~~~~~~~~

.. code-block:: text

   stream.Stream.pipe(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Stream();
   const result = instance.pipe(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Transform class
---------------

**Transform** — Transform streams are ``Duplex`` streams where the output is in some way related to the input. Like all ``Duplex`` streams, ``Transform`` streams implement both the ``Readable`` and ``Writable`` interfaces.

Examples of ``Transform`` streams include:

Transform._final
~~~~~~~~~~~~~~~~

.. code-block:: text

   writable._final(callback)

*Kind:* function · *Arguments:* 1

The ``_final()`` method **must not** be called directly. It may be implemented by child classes, and if so, will be called by the internal ``Writable`` class methods only.

This optional function will be called before the stream closes, delaying the ``'finish'`` event until ``callback`` is called. This is useful to close resources or write buffered data before a stream ends.

**Arguments**

* ``callback`` (*Function*): Call this function (optionally with an error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Transform();
   const result = instance._final(() => {});
   console.log(result);


Transform._read
~~~~~~~~~~~~~~~

.. code-block:: text

   readable._read(size)

*Kind:* function · *Arguments:* 1

This function MUST NOT be called by application code directly. It should be implemented by child classes, and called by the internal ``Readable`` class methods only.

All ``Readable`` stream implementations must provide an implementation of the ``readable._read()`` method to fetch data from the underlying resource.

**Arguments**

* ``size`` (*number*): Number of bytes to read asynchronously

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Transform();
   const result = instance._read(0);
   console.log(result);


Transform._transform
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   transform._transform(chunk, encoding, callback)

*Kind:* function · *Arguments:* 3

This function MUST NOT be called by application code directly. It should be implemented by child classes, and called by the internal ``Readable`` class methods only.

All ``Transform`` stream implementations must provide a ``_transform()`` method to accept input and produce output. The ``transform._transform()`` implementation handles the bytes being written, computes an output, then passes that output off to the readable portion using the ``transform.push()`` method.

**Arguments**

* ``chunk`` (*Buffer|string|any*): The ``Buffer`` to be transformed, converted from
* ``encoding`` (*string*): If the chunk is a string, then this is the
* ``callback`` (*Function*): A callback function (optionally with an error

**Example**

.. code-block:: javascript

   transform.prototype._transform = function(data, encoding, callback) {
     this.push(data);
     callback();
   };

   transform.prototype._transform = function(data, encoding, callback) {
     callback(null, data);
   };


Transform._write
~~~~~~~~~~~~~~~~

.. code-block:: text

   writable._write(chunk, encoding, callback)

*Kind:* function · *Arguments:* 3

All ``Writable`` stream implementations must provide a ``writable._write()`` and/or ``writable._writev()`` method to send data to the underlying resource.

``Transform`` streams provide their own implementation of the ``writable._write()``.

**Arguments**

* ``chunk`` (*Buffer|string|any*): The ``Buffer`` to be written, converted from the
* ``encoding`` (*string*): If the chunk is a string, then ``encoding`` is the
* ``callback`` (*Function*): Call this function (optionally with an error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Transform();
   const result = instance._write(Buffer.from('hello'), 'utf8', () => {});
   console.log(result);


Transform._writev
~~~~~~~~~~~~~~~~~

.. code-block:: text

   writable._writev(chunks, callback)

*Kind:* function · *Arguments:* 2

This function MUST NOT be called by application code directly. It should be implemented by child classes, and called by the internal ``Writable`` class methods only.

The ``writable._writev()`` method may be implemented in addition or alternatively to ``writable._write()`` in stream implementations that are capable of processing multiple chunks of data at once. If implemented and if there is buffered data from previous writes, ``_writev()`` will be called instead of ``_write()``.

**Arguments**

* ``chunks`` (*Object\[]*): The data to be written. The value is an array of ``Object``
* ``callback`` (*Function*): A callback function (optionally with an error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Transform();
   const result = instance._writev([], () => {});
   console.log(result);


Writable class
--------------

**Writable** — ##### Event: ``'close'``

The ``'close'`` event is emitted when the stream and any of its underlying resources (a file descriptor, for example) have been closed. The event indicates that no more events will be emitted, and no further computation will occur.

Writable._destroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   writable._destroy(err, callback)

*Kind:* function · *Arguments:* 2

The ``_destroy()`` method is called by ``writable.destroy()``. It can be overridden by child classes but it **must not** be called directly.

**Arguments**

* ``err`` (*Error*): A possible error.
* ``callback`` (*Function*): A callback function that takes an optional error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance._destroy(new Error('example'), () => {});
   console.log(result);


Writable._undestroy
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable._undestroy(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   console.log(typeof stream.Writable.prototype._undestroy()); // 'function'
   console.log(stream.Writable.prototype._undestroy().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable._write
~~~~~~~~~~~~~~~

.. code-block:: text

   writable._write(chunk, encoding, callback)

*Kind:* function · *Arguments:* 3

All ``Writable`` stream implementations must provide a ``writable._write()`` and/or ``writable._writev()`` method to send data to the underlying resource.

``Transform`` streams provide their own implementation of the ``writable._write()``.

**Arguments**

* ``chunk`` (*Buffer|string|any*): The ``Buffer`` to be written, converted from the
* ``encoding`` (*string*): If the chunk is a string, then ``encoding`` is the
* ``callback`` (*Function*): Call this function (optionally with an error

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance._write(Buffer.from('hello'), 'utf8', () => {});
   console.log(result);


Writable.cork
~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.cork(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.cork();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.destroy
~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.destroy(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.destroy(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.destroySoon
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.destroySoon(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.destroySoon();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.end
~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.end(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.setDefaultEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.setDefaultEncoding(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.setDefaultEncoding(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.uncork
~~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.uncork(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.uncork();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Writable.write
~~~~~~~~~~~~~~

.. code-block:: text

   stream.Writable.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const stream = require('node:stream');

   const instance = new stream.Writable();
   const result = instance.write(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``Duplex.closed`` — boolean
* ``Duplex.destroyed`` — boolean
* ``Duplex.errored`` — null
* ``Duplex.readable`` — boolean
* ``Duplex.readableAborted`` — boolean
* ``Duplex.readableBuffer`` — null
* ``Duplex.readableDidRead`` — boolean
* ``Duplex.readableEncoding`` — null
* ``Duplex.readableEnded`` — boolean
* ``Duplex.readableFlowing`` — null
* ``Duplex.readableHighWaterMark`` — number
* ``Duplex.readableLength`` — number
* ``Duplex.readableObjectMode`` — boolean
* ``Duplex.writable`` — boolean
* ``Duplex.writableBuffer`` — null
* ``Duplex.writableCorked`` — number
* ``Duplex.writableEnded`` — boolean
* ``Duplex.writableFinished`` — boolean
* ``Duplex.writableHighWaterMark`` — number
* ``Duplex.writableLength`` — number
* ``Duplex.writableNeedDrain`` — boolean
* ``Duplex.writableObjectMode`` — boolean
* ``length`` — number
* ``name`` — string
* ``promises`` — object
* ``Readable.closed`` — boolean
* ``Readable.destroyed`` — boolean
* ``Readable.errored`` — null
* ``Readable.readable`` — boolean
* ``Readable.readableAborted`` — boolean
* ``Readable.readableBuffer`` — null
* ``Readable.readableDidRead`` — boolean
* ``Readable.readableEncoding`` — null
* ``Readable.readableEnded`` — boolean
* ``Readable.readableFlowing`` — null
* ``Readable.readableHighWaterMark`` — number
* ``Readable.readableLength`` — number
* ``Readable.readableObjectMode`` — boolean
* ``Stream.closed`` — boolean
* ``Stream.destroyed`` — boolean
* ``Stream.errored`` — null
* ``Stream.readable`` — boolean
* ``Stream.readableAborted`` — boolean
* ``Stream.readableBuffer`` — null
* ``Stream.readableDidRead`` — boolean
* ``Stream.readableEncoding`` — null
* ``Stream.readableEnded`` — boolean
* ``Stream.readableFlowing`` — null
* ``Stream.readableHighWaterMark`` — number
* ``Stream.readableLength`` — number
* ``Stream.readableObjectMode`` — boolean
* ``Stream.writable`` — boolean
* ``Stream.writableBuffer`` — null
* ``Stream.writableCorked`` — number
* ``Stream.writableEnded`` — boolean
* ``Stream.writableFinished`` — boolean
* ``Stream.writableHighWaterMark`` — number
* ``Stream.writableLength`` — number
* ``Stream.writableNeedDrain`` — boolean
* ``Stream.writableObjectMode`` — boolean
* ``Writable.closed`` — boolean
* ``Writable.destroyed`` — boolean
* ``Writable.errored`` — null
* ``Writable.writable`` — boolean
* ``Writable.writableBuffer`` — null
* ``Writable.writableCorked`` — number
* ``Writable.writableEnded`` — boolean
* ``Writable.writableFinished`` — boolean
* ``Writable.writableHighWaterMark`` — number
* ``Writable.writableLength`` — number
* ``Writable.writableNeedDrain`` — boolean
* ``Writable.writableObjectMode`` — boolean

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
