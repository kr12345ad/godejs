Global objects
==============

GodeJS exposes these host globals alongside the Web APIs it implements. Standard ECMAScript built-ins (``Object``, ``Array``, ``Promise``, ``Map``, typed arrays, and so on) are not repeated here; see the ECMAScript specification for those.

Functions
---------

AbortSignal
~~~~~~~~~~~

.. code-block:: text

   AbortSignal

*Kind:* function · *Arguments:* 1

Type: Documentation-only

Using ``AbortSignal`` to destroy long-lived resources is deprecated. Prefer ``using`` for resource cleanup.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   // Deprecated
   async function example() {
     const ac = new AbortController();
     const server = http.createServer(handler);
     server.listen({ port: 3000, signal: ac.signal });

     await doWork();
     ac.abort();
   }


atob
~~~~

.. code-block:: text

   atob(data)

*Kind:* function · *Arguments:* 1

Global alias for ``buffer.atob()``.

An automated migration is available (source):

**Arguments**

* ``data``


btoa
~~~~

.. code-block:: text

   btoa(data)

*Kind:* function · *Arguments:* 1

Global alias for ``buffer.btoa()``.

An automated migration is available (source):

**Arguments**

* ``data``


clearImmediate
~~~~~~~~~~~~~~

.. code-block:: text

   clearImmediate(immediateObject)

*Kind:* function · *Arguments:* 1

``clearImmediate`` is described in the timers section.

**Arguments**

* ``immediateObject``


clearInterval
~~~~~~~~~~~~~

.. code-block:: text

   clearInterval(intervalObject)

*Kind:* function · *Arguments:* 1

``clearInterval`` is described in the timers section.

**Arguments**

* ``intervalObject``


clearTimeout
~~~~~~~~~~~~

.. code-block:: text

   clearTimeout(timeoutObject)

*Kind:* function · *Arguments:* 1

``clearTimeout`` is described in the timers section.

**Arguments**

* ``timeoutObject``


CryptoKey
~~~~~~~~~

.. code-block:: text

   CryptoKey

*Kind:* function · *Arguments:* 0

Type: Documentation-only

Passing a ``CryptoKey`` to ``node:crypto`` functions is deprecated and will throw an error in a future version. This includes ``crypto.createPublicKey()``, ``crypto.createPrivateKey()``, ``crypto.sign()``, ``crypto.verify()``, ``crypto.publicEncrypt()``, ``crypto.publicDecrypt()``, ``crypto.privateEncrypt()``, ``crypto.privateDecrypt()``, ``Sign.prototype.sign()``, ``Verify.prototype.verify()``, ``crypto.createHmac()``, ``crypto.createCipheriv()``, ``crypto.createDecipheriv()``, ``crypto.encapsulate()``, and ``crypto.decapsulate()``.

This callable takes no arguments.


EventTarget
~~~~~~~~~~~

.. code-block:: text

   EventTarget

*Kind:* function · *Arguments:* 0

The ``EventTarget`` and ``Event`` objects are a Node.js-specific implementation of the ``EventTarget`` Web API that are exposed by some Node.js core APIs.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const target = new EventTarget();

   target.addEventListener('foo', (event) => {
     console.log('foo event happened!');
   });


fetch
~~~~~

.. code-block:: text

   fetch

*Kind:* function · *Arguments:* 2

A browser-compatible implementation of the ``fetch()`` function.

The implementation is based upon undici, an HTTP/1.1 client written from scratch for Node.js. You can figure out which version of ``undici`` is bundled in your Node.js process reading the ``process.versions.undici`` property.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   fetch(url, { dispatcher: new MyAgent() });


queueMicrotask
~~~~~~~~~~~~~~

.. code-block:: text

   queueMicrotask(callback)

*Kind:* function · *Arguments:* 1

The ``queueMicrotask()`` method queues a microtask to invoke ``callback``. If ``callback`` throws an exception, the ``process`` object ``'uncaughtException'`` event will be emitted.

The microtask queue is managed by V8 and may be used in a similar manner to the ``process.nextTick()`` queue, which is managed by Node.js. The ``process.nextTick()`` queue is always processed before the microtask queue within each turn of the Node.js event loop.

**Arguments**

* ``callback`` (*Function*): Function to be queued.

**Example**

.. code-block:: javascript

   // Here, `queueMicrotask()` is used to ensure the 'load' event is always
   // emitted asynchronously, and therefore consistently. Using
   // `process.nextTick()` here would result in the 'load' event always emitting
   // before any other promise jobs.

   DataHandler.prototype.load = async function load(key) {
     const hit = this._cache.get(key);
     if (hit !== undefined) {
       queueMicrotask(() => {
         this.emit('load', hit);
       });
       return;
     }

     const data = await fetchData(key);
     this._cache.set(key, data);
     this.emit('load', data);
   };


QuotaExceededError
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.QuotaExceededError(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


require
~~~~~~~

.. code-block:: text

   require(id)

*Kind:* function · *Arguments:* 1

Load a CommonJS builtin or file relative to the current module. In ``-e`` and stdin code the base directory is the working directory.

**Arguments**

* ``id`` (*string*): Builtin name, package name, or relative/absolute path.

**Returns**

any


setImmediate
~~~~~~~~~~~~

.. code-block:: text

   setImmediate(callback[, ...args])

*Kind:* function · *Arguments:* 2

``setImmediate`` is described in the timers section.

**Arguments**

* ``callback``
* ``...args`` — optional


setInterval
~~~~~~~~~~~

.. code-block:: text

   setInterval(callback, delay[, ...args])

*Kind:* function · *Arguments:* 3

``setInterval`` is described in the timers section.

**Arguments**

* ``callback``
* ``delay``
* ``...args`` — optional


setTimeout
~~~~~~~~~~

.. code-block:: text

   setTimeout(callback, delay[, ...args])

*Kind:* function · *Arguments:* 3

``setTimeout`` is described in the timers section.

**Arguments**

* ``callback``
* ``delay``
* ``...args`` — optional


structuredClone
~~~~~~~~~~~~~~~

.. code-block:: text

   structuredClone(value[, options])

*Kind:* function · *Arguments:* 2

The WHATWG ``structuredClone`` method.

**Arguments**

* ``value``
* ``options`` — optional


localStorage
------------

A persistent Web Storage area. Values are strings, keys keep insertion order, and data is stored in the file selected by ``--localstorage-file`` or the per-working-directory default. ``localStorage`` is an object, not a function.

localStorage.clear
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   localStorage.clear()

*Kind:* function · *Arguments:* 0

Delete every key in the area.

**Returns**

undefined


localStorage.getItem
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   localStorage.getItem(key)

*Kind:* function · *Arguments:* 1

Return the value stored under ``key``, or ``null`` when the key does not exist.

**Arguments**

* ``key`` (*string*): Key to look up; converted to a string.

**Returns**

string | null


localStorage.key
~~~~~~~~~~~~~~~~

.. code-block:: text

   localStorage.key(index)

*Kind:* function · *Arguments:* 1

Return the key at ``index`` in insertion order, or ``null`` when the index is out of range.

**Arguments**

* ``index`` (*number*): Zero-based key index.

**Returns**

string | null


localStorage.removeItem
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   localStorage.removeItem(key)

*Kind:* function · *Arguments:* 1

Delete ``key`` from the area. Removing a missing key is a no-op.

**Arguments**

* ``key`` (*string*): Key to delete.

**Returns**

undefined


localStorage.setItem
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   localStorage.setItem(key, value)

*Kind:* function · *Arguments:* 2

Store ``value`` under ``key``. Both arguments are converted to strings.

**Arguments**

* ``key`` (*string*): Key to write.
* ``value`` (*string*): Value to store.

**Returns**

undefined


AbortController
---------------

A utility class used to signal cancelation in selected ``Promise``-based APIs. The API is based on the Web API ``AbortController``.

AbortController.abort
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   abortController.abort([reason])

*Kind:* function · *Arguments:* 1

Triggers the abort signal, causing the ``abortController.signal`` to emit the ``'abort'`` event.

**Arguments**

* ``reason`` (*any*) — optional: An optional reason, retrievable on the ``AbortSignal``'s


AbortSignal
-----------

AbortSignal.abort
~~~~~~~~~~~~~~~~~

.. code-block:: text

   AbortSignal.abort([reason])

*Kind:* function · *Arguments:* 1

Returns a new already aborted ``AbortSignal``.

**Arguments**

* ``reason`` (*any*) — optional


AbortSignal.any
~~~~~~~~~~~~~~~

.. code-block:: text

   AbortSignal.any(signals)

*Kind:* function · *Arguments:* 1

Returns a new ``AbortSignal`` which will be aborted if any of the provided signals are aborted. Its ``abortSignal.reason`` will be set to whichever one of the ``signals`` caused it to be aborted.

**Arguments**

* ``signals`` (*AbortSignal\[]*): The ``AbortSignal``\ s of which to compose a new ``AbortSignal``.


AbortSignal._abort
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal._abort(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal._fire
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal._fire(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal._makeEvent
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal._makeEvent(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal.addEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal.addEventListener(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal.dispatchEvent
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal.dispatchEvent(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal.removeEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.AbortSignal.removeEventListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


AbortSignal.throwIfAborted
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   abortSignal.throwIfAborted()

*Kind:* function · *Arguments:* 0

If ``abortSignal.aborted`` is ``true``, throws ``abortSignal.reason``.

This callable takes no arguments.


AbortSignal.timeout
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   AbortSignal.timeout(delay)

*Kind:* function · *Arguments:* 1

Returns a new ``AbortSignal`` which will be aborted in ``delay`` milliseconds.

**Arguments**

* ``delay`` (*number*): The number of milliseconds to wait before triggering


Blob
----

A ``Blob`` encapsulates immutable, raw data that can be safely shared across multiple worker threads.

Blob.arrayBuffer
~~~~~~~~~~~~~~~~

.. code-block:: text

   blob.arrayBuffer()

*Kind:* function · *Arguments:* 0

Returns a promise that fulfills with an ``ArrayBuffer`` containing a copy of the ``Blob`` data.

This callable takes no arguments.


Blob.bytes
~~~~~~~~~~

.. code-block:: text

   blob.bytes()

*Kind:* function · *Arguments:* 0

The ``blob.bytes()`` method returns the byte of the ``Blob`` object as a ``Promise<Uint8Array>``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const blob = new Blob(['hello']);
   blob.bytes().then((bytes) => {
     console.log(bytes); // Outputs: Uint8Array(5) [ 104, 101, 108, 108, 111 ]
   });


Blob.slice
~~~~~~~~~~

.. code-block:: text

   blob.slice([start[, end[, type]]])

*Kind:* function · *Arguments:* 3

Creates and returns a new ``Blob`` containing a subset of this ``Blob`` objects data. The original ``Blob`` is not altered.

**Arguments**

* ``start`` (*number*) — optional: The starting index.
* ``end`` (*number*) — optional: The ending index.
* ``type`` (*string*) — optional: The content-type for the new ``Blob``


Blob.stream
~~~~~~~~~~~

.. code-block:: text

   blob.stream()

*Kind:* function · *Arguments:* 0

Returns a new ``ReadableStream`` that allows the content of the ``Blob`` to be read.

This callable takes no arguments.


Blob.text
~~~~~~~~~

.. code-block:: text

   blob.text()

*Kind:* function · *Arguments:* 0

Returns a promise that fulfills with the contents of the ``Blob`` decoded as a UTF-8 string.

This callable takes no arguments.


Blob.textStream
~~~~~~~~~~~~~~~

.. code-block:: text

   blob.textStream()

*Kind:* function · *Arguments:* 0

Returns a new ``ReadableStream`` that allows the content of the ``Blob`` to be read as a stream of UTF-8 decoded strings. It is equivalent to piping ``blob.stream()`` through a ``TextDecoderStream`` set up with UTF-8.

This callable takes no arguments.


**Properties**

* ``Blob.type`` — undefined

BroadcastChannel
----------------

See ``BroadcastChannel``.

BroadcastChannel.close
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   broadcastChannel.close()

*Kind:* function · *Arguments:* 0

Closes the ``BroadcastChannel`` connection.

This callable takes no arguments.


BroadcastChannel.postMessage
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   broadcastChannel.postMessage(message)

*Kind:* function · *Arguments:* 1

**Arguments**

* ``message`` (*any*): Any cloneable JavaScript value.


BroadcastChannel.ref
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   broadcastChannel.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``. Calling ``ref()`` on a previously ``unref()``\ ed BroadcastChannel does *not* let the program exit if it's the only active handle left (the default behavior). If the port is ``ref()``\ ed, calling ``ref()`` again has no effect.

This callable takes no arguments.


BroadcastChannel.unref
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   broadcastChannel.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a BroadcastChannel allows the thread to exit if this is the only active handle in the event system. If the BroadcastChannel is already ``unref()``\ ed calling ``unref()`` again has no effect.

This callable takes no arguments.


Buffer
------

The ``Buffer`` class is a global type for dealing with binary data directly. It can be constructed in a variety of ways.

Buffer.alloc
~~~~~~~~~~~~

.. code-block:: text

   Buffer.alloc(size[, fill[, encoding]])

*Kind:* function · *Arguments:* 3

Allocates a new ``Buffer`` of ``size`` bytes. If ``fill`` is ``undefined``, the ``Buffer`` will be zero-filled.

If ``size`` is larger than ``buffer.constants.MAX_LENGTH`` or smaller than 0, ``ERR_OUT_OF_RANGE`` is thrown.

**Arguments**

* ``size`` (*integer*): The desired length of the new ``Buffer``.
* ``fill`` (*string|Buffer|Uint8Array|integer*) — optional: A value to pre-fill the new ``Buffer``
* ``encoding`` (*string*) — optional: If ``fill`` is a string, this is its encoding.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.alloc(5);

   console.log(buf);
   // Prints: <Buffer 00 00 00 00 00>


Buffer.allocUnsafe
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.allocUnsafe(size)

*Kind:* function · *Arguments:* 1

Allocates a new ``Buffer`` of ``size`` bytes. If ``size`` is larger than ``buffer.constants.MAX_LENGTH`` or smaller than 0, ``ERR_OUT_OF_RANGE`` is thrown.

The underlying memory for ``Buffer`` instances created in this way is \_not initialized\_. The contents of the newly created ``Buffer`` are unknown and *may contain sensitive data*. Use ``Buffer.alloc()`` instead to initialize ``Buffer`` instances with zeroes.

**Arguments**

* ``size`` (*integer*): The desired length of the new ``Buffer``.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(10);

   console.log(buf);
   // Prints (contents may vary): <Buffer a0 8b 28 3f 01 00 00 00 50 32>

   buf.fill(0);

   console.log(buf);
   // Prints: <Buffer 00 00 00 00 00 00 00 00 00 00>


Buffer.allocUnsafeSlow
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.allocUnsafeSlow(size)

*Kind:* function · *Arguments:* 1

Allocates a new ``Buffer`` of ``size`` bytes. If ``size`` is larger than ``buffer.constants.MAX_LENGTH`` or smaller than 0, ``ERR_OUT_OF_RANGE`` is thrown. A zero-length ``Buffer`` is created if ``size`` is 0.

The underlying memory for ``Buffer`` instances created in this way is \_not initialized\_. The contents of the newly created ``Buffer`` are unknown and *may contain sensitive data*. Use ``buf.fill(0)`` to initialize such ``Buffer`` instances with zeroes.

**Arguments**

* ``size`` (*integer*): The desired length of the new ``Buffer``.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Need to keep around a few small chunks of memory.
   const store = [];

   socket.on('readable', () => {
     let data;
     while (null !== (data = readable.read())) {
       // Allocate for retained data.
       const sb = Buffer.allocUnsafeSlow(10);

       // Copy the data into the new allocation.
       data.copy(sb, 0, 0, 10);

       store.push(sb);
     }
   });


Buffer.byteLength
~~~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.byteLength(string[, encoding])

*Kind:* function · *Arguments:* 2

Returns the byte length of a string when encoded using ``encoding``. This is not the same as ``String.prototype.length``, which does not account for the encoding that is used to convert the string into bytes.

For ``'base64'``, ``'base64url'``, and ``'hex'``, this function assumes valid input. For strings that contain non-base64/hex-encoded data (e.g. whitespace), the return value might be greater than the length of a ``Buffer`` created from the string.

**Arguments**

* ``string`` (*string|Buffer|TypedArray|DataView|ArrayBuffer|SharedArrayBuffer*): A
* ``encoding`` (*string*) — optional: If ``string`` is a string, this is its encoding.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const str = '\u00bd + \u00bc = \u00be';

   console.log(`${str}: ${str.length} characters, ` +
               `${Buffer.byteLength(str, 'utf8')} bytes`);
   // Prints: ½ + ¼ = ¾: 9 characters, 12 bytes


Buffer.compare
~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.compare(buf1, buf2)

*Kind:* function · *Arguments:* 2

Compares ``buf1`` to ``buf2``, typically for the purpose of sorting arrays of ``Buffer`` instances. This is equivalent to calling ``buf1.compare(buf2)``.

**Arguments**

* ``buf1`` (*Buffer|Uint8Array*)
* ``buf2`` (*Buffer|Uint8Array*)

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from('1234');
   const buf2 = Buffer.from('0123');
   const arr = [buf1, buf2];

   console.log(arr.sort(Buffer.compare));
   // Prints: [ <Buffer 30 31 32 33>, <Buffer 31 32 33 34> ]
   // (This result is equal to: [buf2, buf1].)


Buffer.concat
~~~~~~~~~~~~~

.. code-block:: text

   Buffer.concat(list[, totalLength])

*Kind:* function · *Arguments:* 2

Returns a new ``Buffer`` which is the result of concatenating all the ``Buffer`` instances in the ``list`` together.

If the list has no items, or if the ``totalLength`` is 0, then a new zero-length ``Buffer`` is returned.

**Arguments**

* ``list`` (*Buffer\[] | Uint8Array\[]*): List of ``Buffer`` or ``Uint8Array``
* ``totalLength`` (*integer*) — optional: Total length of the ``Buffer`` instances in ``list``

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Create a single `Buffer` from a list of three `Buffer` instances.

   const buf1 = Buffer.alloc(10);
   const buf2 = Buffer.alloc(14);
   const buf3 = Buffer.alloc(18);
   const totalLength = buf1.length + buf2.length + buf3.length;

   console.log(totalLength);
   // Prints: 42

   const bufA = Buffer.concat([buf1, buf2, buf3], totalLength);

   console.log(bufA);
   // Prints: <Buffer 00 00 00 00 ...>
   console.log(bufA.length);
   // Prints: 42


Buffer.copyBytesFrom
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.copyBytesFrom(view[, offset[, length]])

*Kind:* function · *Arguments:* 3

Copies the underlying memory of ``view`` into a new ``Buffer``.

**Arguments**

* ``view`` (*TypedArray*): The ``TypedArray`` to copy.
* ``offset`` (*integer*) — optional: The starting offset within ``view``. **Default:** ``0``.
* ``length`` (*integer*) — optional: The number of elements from ``view`` to copy.

**Example**

.. code-block:: javascript

   const u16 = new Uint16Array([0, 0xffff]);
   const buf = Buffer.copyBytesFrom(u16, 1, 1);
   u16[1] = 0;
   console.log(buf.length); // 2
   console.log(buf[0]); // 255
   console.log(buf[1]); // 255


Buffer.from
~~~~~~~~~~~

.. code-block:: text

   Buffer.from(array)

*Kind:* function · *Arguments:* 1

Allocates a new ``Buffer`` using an ``array`` of bytes in the range ``0`` – ``255``. Array entries outside that range will be truncated to fit into it.

If ``array`` is an ``Array``-like object (that is, one with a ``length`` property of type ``number``), it is treated as if it is an array, unless it is a ``Buffer`` or a ``Uint8Array``. This means all other ``TypedArray`` variants get treated as an ``Array``. To create a ``Buffer`` from the bytes backing a ``TypedArray``, use ``Buffer.copyBytesFrom()``.

**Arguments**

* ``array`` (*integer\[]*)

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Creates a new Buffer containing the UTF-8 bytes of the string 'buffer'.
   const buf = Buffer.from([0x62, 0x75, 0x66, 0x66, 0x65, 0x72]);


Buffer.isBuffer
~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.isBuffer(obj)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``obj`` is a ``Buffer``, ``false`` otherwise.

**Arguments**

* ``obj`` (*Object*)

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   Buffer.isBuffer(Buffer.alloc(10)); // true
   Buffer.isBuffer(Buffer.from('foo')); // true
   Buffer.isBuffer('a string'); // false
   Buffer.isBuffer([]); // false
   Buffer.isBuffer(new Uint8Array(1024)); // false


Buffer.isEncoding
~~~~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.isEncoding(encoding)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``encoding`` is the name of a supported character encoding, or ``false`` otherwise.

**Arguments**

* ``encoding`` (*string*): A character encoding name to check.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   console.log(Buffer.isEncoding('utf8'));
   // Prints: true

   console.log(Buffer.isEncoding('hex'));
   // Prints: true

   console.log(Buffer.isEncoding('utf/8'));
   // Prints: false

   console.log(Buffer.isEncoding(''));
   // Prints: false


Buffer.of
~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.of(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.asciiSlice
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.asciiSlice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.asciiWrite
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.asciiWrite(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64Slice
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.base64Slice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64urlSlice
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.base64urlSlice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64urlWrite
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.base64urlWrite(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64Write
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.base64Write(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.compare
~~~~~~~~~~~~~~

.. code-block:: text

   Buffer.compare(buf1, buf2)

*Kind:* function · *Arguments:* 2

Compares ``buf1`` to ``buf2``, typically for the purpose of sorting arrays of ``Buffer`` instances. This is equivalent to calling ``buf1.compare(buf2)``.

**Arguments**

* ``buf1`` (*Buffer|Uint8Array*)
* ``buf2`` (*Buffer|Uint8Array*)

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from('1234');
   const buf2 = Buffer.from('0123');
   const arr = [buf1, buf2];

   console.log(arr.sort(Buffer.compare));
   // Prints: [ <Buffer 30 31 32 33>, <Buffer 31 32 33 34> ]
   // (This result is equal to: [buf2, buf1].)


Buffer.copy
~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.copy(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.equals
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.equals(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.fill
~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.fill(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.hexSlice
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.hexSlice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.hexWrite
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.hexWrite(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.includes
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.includes(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.indexOf
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.indexOf(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.inspect
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.inspect(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.lastIndexOf
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.lastIndexOf(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.latin1Slice
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.latin1Slice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.latin1Write
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.latin1Write(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigInt64BE
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigInt64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigInt64LE
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigInt64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigUInt64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigUInt64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigUint64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigUint64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigUInt64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigUInt64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigUint64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readBigUint64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readDoubleBE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readDoubleBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readDoubleLE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readDoubleLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readFloatBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readFloatBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readFloatLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readFloatLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readInt16BE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readInt16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readInt16LE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readInt16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readInt32BE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readInt32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readInt32LE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readInt32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readInt8
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readInt8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readIntBE
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readIntBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readIntLE
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readIntLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUInt16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUint16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUint16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUInt16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUint16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUint16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUInt32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUint32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUint32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUInt32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUint32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUint32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt8
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUInt8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUint8
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUint8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUIntBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUIntBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUintBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUintBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUIntLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUIntLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUintLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.readUintLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.slice
~~~~~~~~~~~~

.. code-block:: text

   buffer.slice(start, end)

*Kind:* function · *Arguments:* 2

Type: Documentation-only

This method was deprecated because it is not compatible with ``Uint8Array.prototype.slice()``, which is a superclass of ``Buffer``.

**Arguments**

* ``start``
* ``end``


Buffer.subarray
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.subarray(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.swap16
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.swap16(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.swap32
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.swap32(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.swap64
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.swap64(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.toJSON
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.toJSON(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.toLocaleString
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.toLocaleString(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.toString
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.toString(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.ucs2Slice
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.ucs2Slice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.ucs2Write
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.ucs2Write(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.utf8Slice
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.utf8Slice(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.utf8Write
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.utf8Write(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.write
~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.write(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigInt64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigInt64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigInt64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigInt64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigUInt64BE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigUInt64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigUint64BE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigUint64BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigUInt64LE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigUInt64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigUint64LE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeBigUint64LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeDoubleBE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeDoubleBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeDoubleLE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeDoubleLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeFloatBE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeFloatBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeFloatLE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeFloatLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeInt16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeInt16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeInt16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeInt16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeInt32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeInt32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeInt32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeInt32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeInt8
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeInt8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeIntBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeIntBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeIntLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeIntLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt16BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUInt16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUint16BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUint16BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt16LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUInt16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUint16LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUint16LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt32BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUInt32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUint32BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUint32BE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt32LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUInt32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUint32LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUint32LE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt8
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUInt8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUint8
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUint8(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUIntBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUIntBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUintBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUintBE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUIntLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUIntLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUintLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Buffer.writeUintLE(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


**Properties**

* ``Buffer.poolSize`` — number
* ``Buffer.offset`` — undefined
* ``Buffer.parent`` — undefined

ByteLengthQueuingStrategy
-------------------------

A browser-compatible implementation of ``ByteLengthQueuingStrategy``.

ByteLengthQueuingStrategy.size
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   byteLengthQueuingStrategy.size

*Kind:* function · *Arguments:* 1

This callable takes no arguments.


CompressionStream
-----------------

A browser-compatible implementation of ``CompressionStream``.

**Properties**

* ``CompressionStream.readable`` — undefined
* ``CompressionStream.writable`` — undefined

console
-------

Used to print to stdout and stderr. See the ``console`` section.

CountQueuingStrategy
--------------------

A browser-compatible implementation of ``CountQueuingStrategy``.

CountQueuingStrategy.size
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   countQueuingStrategy.size

*Kind:* function · *Arguments:* 0

This callable takes no arguments.


crypto
------

crypto.getRandomValues
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.getRandomValues(typedArray)

*Kind:* function · *Arguments:* 1

A convenient alias for ``crypto.webcrypto.getRandomValues()``. This implementation is not compliant with the Web Crypto spec, to write web-compatible code use ``crypto.webcrypto.getRandomValues()`` instead.

**Arguments**

* ``typedArray`` (*Buffer|TypedArray|DataView|ArrayBuffer*)


crypto.randomUUID
~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.randomUUID([options])

*Kind:* function · *Arguments:* 1

Generates a random RFC 4122 version 4 UUID. The UUID is generated using a cryptographic pseudorandom number generator.

**Arguments**

* ``options`` (*Object*) — optional


**Properties**

* ``crypto.subtle`` — object

Crypto
------

A browser-compatible implementation of ``Crypto``. This global is available only if the Node.js binary was compiled with including support for the ``node:crypto`` module.

CryptoKey
---------

**Properties**

* ``CryptoKey.$gode$id`` — number

CustomEvent
-----------

The ``CustomEvent`` object is an adaptation of the ``CustomEvent`` Web API. Instances are created internally by Node.js.

**Properties**

* ``CustomEvent.detail`` — undefined

DecompressionStream
-------------------

A browser-compatible implementation of ``DecompressionStream``.

**Properties**

* ``DecompressionStream.readable`` — undefined
* ``DecompressionStream.writable`` — undefined

DOMException
------------

The Web IDL ``DOMException`` class. These errors are thrown by web-platform APIs in Node.js such as ``fetch()``, ``AbortController``, ``AbortSignal``, and Web Streams. For details, see also ``Class: DOMException`` on the Globals page.

The ``domException.name`` property identifies the type of the exception (for example, ``'AbortError'``). Unlike most errors in Node.js, the ``domException.code`` property is a number that corresponds to a legacy error code name (for example, ``20`` for ``ABORT_ERR``).

Event
-----

The ``Event`` object is an adaptation of the ``Event`` Web API. Instances are created internally by Node.js.

Event.composedPath
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.composedPath()

*Kind:* function · *Arguments:* 0

Returns an array containing the current ``EventTarget`` as the only entry or empty if the event is not being dispatched. This is not used in Node.js and is provided purely for completeness.

This callable takes no arguments.


Event.initEvent
~~~~~~~~~~~~~~~

.. code-block:: text

   event.initEvent(type[, bubbles[, cancelable]])

*Kind:* function · *Arguments:* 3

Redundant with event constructors and incapable of setting ``composed``. This is not used in Node.js and is provided purely for completeness.

**Arguments**

* ``type`` (*string*)
* ``bubbles`` (*boolean*) — optional
* ``cancelable`` (*boolean*) — optional


Event.preventDefault
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.preventDefault()

*Kind:* function · *Arguments:* 0

Sets the ``defaultPrevented`` property to ``true`` if ``cancelable`` is ``true``.

This callable takes no arguments.


Event.stopImmediatePropagation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.stopImmediatePropagation()

*Kind:* function · *Arguments:* 0

Stops the invocation of event listeners after the current one completes.

This callable takes no arguments.


Event.stopPropagation
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.stopPropagation()

*Kind:* function · *Arguments:* 0

This is not used in Node.js and is provided purely for completeness.

This callable takes no arguments.


**Properties**

* ``Event.AT_TARGET`` — number
* ``Event.BUBBLING_PHASE`` — number
* ``Event.CAPTURING_PHASE`` — number
* ``Event.NONE`` — number
* ``Event.bubbles`` — undefined
* ``Event.cancelable`` — undefined
* ``Event.cancelBubble`` — undefined
* ``Event.composed`` — undefined
* ``Event.currentTarget`` — undefined
* ``Event.defaultPrevented`` — undefined
* ``Event.eventPhase`` — undefined
* ``Event.isTrusted`` — undefined
* ``Event.returnValue`` — boolean
* ``Event.srcElement`` — undefined
* ``Event.target`` — undefined
* ``Event.timeStamp`` — undefined
* ``Event.type`` — undefined

EventTarget
-----------

EventTarget.addEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.addEventListener(type, listener[, options])

*Kind:* function · *Arguments:* 3

Adds a new handler for the ``type`` event. Any given ``listener`` is added only once per ``type`` and per ``capture`` option value.

If the ``once`` option is ``true``, the ``listener`` is removed after the next time a ``type`` event is dispatched.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   function handler(event) {}

   const target = new EventTarget();
   target.addEventListener('foo', handler, { capture: true });  // first
   target.addEventListener('foo', handler, { capture: false }); // second

   // Removes the second instance of handler
   target.removeEventListener('foo', handler);

   // Removes the first instance of handler
   target.removeEventListener('foo', handler, { capture: true });


EventTarget.dispatchEvent
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.dispatchEvent(event)

*Kind:* function · *Arguments:* 1

Dispatches the ``event`` to the list of handlers for ``event.type``.

The registered event listeners is synchronously invoked in the order they were registered.

**Arguments**

* ``event`` (*Event*)


EventTarget.removeEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.removeEventListener(type, listener[, options])

*Kind:* function · *Arguments:* 3

Removes the ``listener`` from the list of handlers for event ``type``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional


File
----

A ``File`` provides information about files.

**Properties**

* ``File.lastModified`` — undefined
* ``File.name`` — undefined

FormData
--------

A browser-compatible implementation of ``FormData``.

FormData.append
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.append(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.delete
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.delete(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.entries
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.entries(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.forEach
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.forEach(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.get
~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.get(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.getAll
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.getAll(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.has
~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.has(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.keys
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.keys(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.set
~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.set(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


FormData.values
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.FormData.values(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


global
------

An alias for ``globalThis``, matching Node's CommonJS global.

Headers
-------

A browser-compatible implementation of ``Headers``.

Headers.append
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.append(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.delete
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.delete(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.entries
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.entries(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.forEach
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.forEach(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.get
~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.get(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.getSetCookie
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.getSetCookie(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.has
~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.has(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.keys
~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.keys(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.set
~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.set(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Headers.values
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Headers.values(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessageChannel
--------------

The ``MessageChannel`` class. See ``MessageChannel`` for more details.

MessageEvent
------------

A browser-compatible implementation of ``MessageEvent``.

MessagePort
-----------

The ``MessagePort`` class. See ``MessagePort`` for more details.

MessagePort.close
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.close(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessagePort.hasRef
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.hasRef(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessagePort.postMessage
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.postMessage(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessagePort.ref
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.ref(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessagePort.start
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.start(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


MessagePort.unref
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.MessagePort.unref(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


navigator
---------

A partial implementation of ``window.navigator``.

**Properties**

* ``navigator.hardwareConcurrency`` — number
* ``navigator.language`` — string
* ``navigator.languages`` — array
* ``navigator.platform`` — string
* ``navigator.userAgent`` — string

performance
-----------

The ``perf_hooks.performance`` object.

performance.clearMarks
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.clearMarks([name])

*Kind:* function · *Arguments:* 1

If ``name`` is not provided, removes all ``PerformanceMark`` objects from the Performance Timeline. If ``name`` is provided, removes only the named mark.

**Arguments**

* ``name`` (*string*) — optional


performance.clearMeasures
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.clearMeasures([name])

*Kind:* function · *Arguments:* 1

If ``name`` is not provided, removes all ``PerformanceMeasure`` objects from the Performance Timeline. If ``name`` is provided, removes only the named measure.

**Arguments**

* ``name`` (*string*) — optional


performance.getEntries
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntries()

*Kind:* function · *Arguments:* 0

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime``. If you are only interested in performance entries of certain types or that have certain names, see ``performance.getEntriesByType()`` and ``performance.getEntriesByName()``.

This callable takes no arguments.


performance.getEntriesByName
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntriesByName(name[, type])

*Kind:* function · *Arguments:* 2

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.name`` is equal to ``name``, and optionally, whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``name`` (*string*)
* ``type`` (*string*) — optional


performance.getEntriesByType
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntriesByType(type)

*Kind:* function · *Arguments:* 1

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``type`` (*string*)


performance.mark
~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.mark(name[, options])

*Kind:* function · *Arguments:* 2

Creates a new ``PerformanceMark`` entry in the Performance Timeline. A ``PerformanceMark`` is a subclass of ``PerformanceEntry`` whose ``performanceEntry.entryType`` is always ``'mark'``, and whose ``performanceEntry.duration`` is always ``0``. Performance marks are used to mark specific significant moments in the Performance Timeline.

The created ``PerformanceMark`` entry is put in the global Performance Timeline and can be queried with ``performance.getEntries``, ``performance.getEntriesByName``, and ``performance.getEntriesByType``. When the observation is performed, the entries should be cleared from the global Performance Timeline manually with ``performance.clearMarks``.

**Arguments**

* ``name`` (*string*)
* ``options`` (*Object*) — optional


performance.measure
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.measure(name[, startMarkOrOptions[, endMark]])

*Kind:* function · *Arguments:* 3

Creates a new ``PerformanceMeasure`` entry in the Performance Timeline. A ``PerformanceMeasure`` is a subclass of ``PerformanceEntry`` whose ``performanceEntry.entryType`` is always ``'measure'``, and whose ``performanceEntry.duration`` measures the number of milliseconds elapsed since ``startMark`` and ``endMark``.

The ``startMark`` argument may identify any *existing* ``PerformanceMark`` in the Performance Timeline, or *may* identify any of the timestamp properties provided by the ``PerformanceNodeTiming`` class. If the named ``startMark`` does not exist, an error is thrown.

**Arguments**

* ``name`` (*string*)
* ``startMarkOrOptions`` (*string|Object*) — optional: Optional.
* ``endMark`` (*string*) — optional: Optional. Must be omitted if ``startMarkOrOptions`` is an


performance.now
~~~~~~~~~~~~~~~

.. code-block:: text

   performance.now()

*Kind:* function · *Arguments:* 0

Returns the current high resolution millisecond timestamp, where 0 represents the start of the current ``node`` process.

This callable takes no arguments.


performance.toJSON
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.toJSON()

*Kind:* function · *Arguments:* 0

An object which is JSON representation of the ``performance`` object. It is similar to ``window.performance.toJSON`` in browsers.

This callable takes no arguments.


**Properties**

* ``performance.timeOrigin`` — number

process
-------

The process object. See the ``process`` object section.

ReadableStream
--------------

A browser-compatible implementation of ``ReadableStream``.

ReadableStream._next
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.ReadableStream._next(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ReadableStream._wake
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.ReadableStream._wake(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ReadableStream.cancel
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.cancel([reason])

*Kind:* function · *Arguments:* 1

**Arguments**

* ``reason`` (*any*) — optional


ReadableStream.getReader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.getReader([options])

*Kind:* function · *Arguments:* 1

Causes the ``readableStream.locked`` to be ``true``.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { ReadableStream } = require('node:stream/web');

   const stream = new ReadableStream();

   const reader = stream.getReader();

   reader.read().then(console.log);


ReadableStream.pipeThrough
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.pipeThrough(transform[, options])

*Kind:* function · *Arguments:* 2

Connects this ``ReadableStream`` to the pair of ``ReadableStream`` and ``WritableStream`` provided in the ``transform`` argument such that the data from this ``ReadableStream`` is written in to ``transform.writable``, possibly transformed, then pushed to ``transform.readable``. Once the pipeline is configured, ``transform.readable`` is returned.

Causes the ``readableStream.locked`` to be ``true`` while the pipe operation is active.

**Arguments**

* ``transform`` (*Object*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const {
     ReadableStream,
     TransformStream,
   } = require('node:stream/web');

   const stream = new ReadableStream({
     start(controller) {
       controller.enqueue('a');
     },
   });

   const transform = new TransformStream({
     transform(chunk, controller) {
       controller.enqueue(chunk.toUpperCase());
     },
   });

   const transformedStream = stream.pipeThrough(transform);

   (async () => {
     for await (const chunk of transformedStream)
       console.log(chunk);
       // Prints: A
   })();


ReadableStream.pipeTo
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.pipeTo(destination[, options])

*Kind:* function · *Arguments:* 2

Causes the ``readableStream.locked`` to be ``true`` while the pipe operation is active.

**Arguments**

* ``destination`` (*WritableStream*): A ``WritableStream`` to which this
* ``options`` (*Object*) — optional


ReadableStream.tee
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.tee()

*Kind:* function · *Arguments:* 0

Returns a pair of new ``ReadableStream`` instances to which this ``ReadableStream``'s data will be forwarded. Each will receive the same data.

Causes the ``readableStream.locked`` to be ``true``.

This callable takes no arguments.


ReadableStream.values
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   readableStream.values([options])

*Kind:* function · *Arguments:* 1

Creates and returns an async iterator usable for consuming this ``ReadableStream``'s data.

Causes the ``readableStream.locked`` to be ``true`` while the async iterator is active.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { Buffer } from 'node:buffer';

   const stream = new ReadableStream(getSomeSource());

   for await (const chunk of stream.values({ preventCancel: true }))
     console.log(Buffer.from(chunk).toString());


Request
-------

A browser-compatible implementation of ``Request``.

Request.arrayBuffer
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.arrayBuffer(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.blob
~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.blob(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.bytes
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.bytes(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.clone
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.clone(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.formData
~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.formData(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.json
~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.json(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Request.text
~~~~~~~~~~~~

.. code-block:: text

   globalThis.Request.text(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


require
-------

require.resolve
~~~~~~~~~~~~~~~

.. code-block:: text

   require.resolve

*Kind:* function · *Arguments:* 0

Relative resolution can be handled via ``new URL('./local', import.meta.url)``.

For a complete ``require.resolve`` replacement, there is the import.meta.resolve API.

This callable takes no arguments.


**Properties**

* ``require.cache`` — object
* ``require.extensions`` — object

Response
--------

A browser-compatible implementation of ``Response``.

Response.error
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.error(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.json
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.json(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.arrayBuffer
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.arrayBuffer(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.blob
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.blob(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.bytes
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.bytes(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.clone
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.clone(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.formData
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.formData(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.json
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.json(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.text
~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.text(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Response.redirect
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.Response.redirect(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


sessionStorage
--------------

An in-memory Web Storage area for the lifetime of the process. ``sessionStorage`` is an object, not a function.

sessionStorage.clear
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sessionStorage.clear()

*Kind:* function · *Arguments:* 0

Delete every key in the area.

**Returns**

undefined


sessionStorage.getItem
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sessionStorage.getItem(key)

*Kind:* function · *Arguments:* 1

Return the value stored under ``key``, or ``null`` when the key does not exist.

**Arguments**

* ``key`` (*string*): Key to look up.

**Returns**

string | null


sessionStorage.key
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sessionStorage.key(index)

*Kind:* function · *Arguments:* 1

Return the key at ``index`` in insertion order, or ``null`` when out of range.

**Arguments**

* ``index`` (*number*): Zero-based key index.

**Returns**

string | null


sessionStorage.removeItem
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sessionStorage.removeItem(key)

*Kind:* function · *Arguments:* 1

Delete ``key`` from the area.

**Arguments**

* ``key`` (*string*): Key to delete.

**Returns**

undefined


sessionStorage.setItem
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   sessionStorage.setItem(key, value)

*Kind:* function · *Arguments:* 2

Store ``value`` under ``key`` for the lifetime of the process.

**Arguments**

* ``key`` (*string*): Key to write.
* ``value`` (*string*): Value to store.

**Returns**

undefined


Storage
-------

The Web Storage interface implemented by ``localStorage`` and ``sessionStorage``. Constructing it directly is illegal; use one of the two global areas.

Storage.clear
~~~~~~~~~~~~~

.. code-block:: text

   storage.clear()

*Kind:* function · *Arguments:* 0

Delete every key in the area.

**Returns**

undefined


Storage.getItem
~~~~~~~~~~~~~~~

.. code-block:: text

   storage.getItem(key)

*Kind:* function · *Arguments:* 1

Return the value stored under ``key``, or ``null`` when absent.

**Arguments**

* ``key`` (*string*): Key to look up.

**Returns**

string | null


Storage.key
~~~~~~~~~~~

.. code-block:: text

   storage.key(index)

*Kind:* function · *Arguments:* 1

Return the key at ``index`` in insertion order, or ``null`` when out of range.

**Arguments**

* ``index`` (*number*): Zero-based key index.

**Returns**

string | null


Storage.removeItem
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   storage.removeItem(key)

*Kind:* function · *Arguments:* 1

Delete ``key`` from the area.

**Arguments**

* ``key`` (*string*): Key to delete.

**Returns**

undefined


Storage.setItem
~~~~~~~~~~~~~~~

.. code-block:: text

   storage.setItem(key, value)

*Kind:* function · *Arguments:* 2

Store a string value under ``key``.

**Arguments**

* ``key`` (*string*): Key to write.
* ``value`` (*string*): Value to store.

**Returns**

undefined


SubtleCrypto
------------

A browser-compatible implementation of ``SubtleCrypto``. This global is available only if the Node.js binary was compiled with including support for the ``node:crypto`` module.

TextDecoder
-----------

The WHATWG ``TextDecoder`` class. See the ``TextDecoder`` section.

TextDecoder.decode
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textDecoder.decode([input[, options]])

*Kind:* function · *Arguments:* 2

Decodes the ``input`` and returns a string. If ``options.stream`` is ``true``, any incomplete byte sequences occurring at the end of the ``input`` are buffered internally and emitted after the next call to ``textDecoder.decode()``.

If ``textDecoder.fatal`` is ``true``, decoding errors that occur will result in a ``TypeError`` being thrown.

**Arguments**

* ``input`` (*ArrayBuffer|DataView|TypedArray*) — optional: An ``ArrayBuffer``, ``DataView``, or
* ``options`` (*Object*) — optional


TextDecoderStream
-----------------

A browser-compatible implementation of ``TextDecoderStream``.

**Properties**

* ``TextDecoderStream.readable`` — undefined
* ``TextDecoderStream.writable`` — undefined

TextEncoder
-----------

The WHATWG ``TextEncoder`` class. See the ``TextEncoder`` section.

TextEncoder.encode
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textEncoder.encode([input])

*Kind:* function · *Arguments:* 1

UTF-8 encodes the ``input`` string and returns a ``Uint8Array`` containing the encoded bytes.

**Arguments**

* ``input`` (*string*) — optional: The text to encode. **Default:** an empty string.


TextEncoder.encodeInto
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   textEncoder.encodeInto(src, dest)

*Kind:* function · *Arguments:* 2

UTF-8 encodes the ``src`` string to the ``dest`` Uint8Array and returns an object containing the read Unicode code units and written UTF-8 bytes.

**Arguments**

* ``src`` (*string*): The text to encode.
* ``dest`` (*Uint8Array*): The array to hold the encode result.

**Example**

.. code-block:: javascript

   const encoder = new TextEncoder();
   const src = 'this is some data';
   const dest = new Uint8Array(10);
   const { read, written } = encoder.encodeInto(src, dest);


**Properties**

* ``TextEncoder.encoding`` — string

TextEncoderStream
-----------------

A browser-compatible implementation of ``TextEncoderStream``.

**Properties**

* ``TextEncoderStream.encoding`` — string
* ``TextEncoderStream.readable`` — undefined
* ``TextEncoderStream.writable`` — undefined

TransformStream
---------------

A browser-compatible implementation of ``TransformStream``.

URL
---

The WHATWG ``URL`` class. See the ``URL`` section.

URL.canParse
~~~~~~~~~~~~

.. code-block:: text

   URL.canParse(input[, base])

*Kind:* function · *Arguments:* 2

Checks if an ``input`` relative to the ``base`` can be parsed to a ``URL``.

**Arguments**

* ``input`` (*string*): The absolute or relative input URL to parse. If ``input``
* ``base`` (*string*) — optional: The base URL to resolve against if the ``input`` is not

**Example**

.. code-block:: javascript

   const isValid = URL.canParse('/foo', 'https://example.org/'); // true

   const isNotValid = URL.canParse('/foo'); // false


URL.createObjectURL
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   URL.createObjectURL(blob)

*Kind:* function · *Arguments:* 1

Creates a ``'blob:nodedata:...'`` URL string that represents the given ``Blob`` object and can be used to retrieve the ``Blob`` later.

The data stored by the registered ``Blob`` will be retained in memory until ``URL.revokeObjectURL()`` is called to remove it.

**Arguments**

* ``blob`` (*Blob*)

**Example**

.. code-block:: javascript

   const {
     Blob,
     resolveObjectURL,
   } = require('node:buffer');

   const blob = new Blob(['hello']);
   const id = URL.createObjectURL(blob);

   // later...

   const otherBlob = resolveObjectURL(id);
   console.log(otherBlob.size);


URL.parse
~~~~~~~~~

.. code-block:: text

   URL.parse(input[, base])

*Kind:* function · *Arguments:* 2

Parses a string as a URL. If ``base`` is provided, it will be used as the base URL for the purpose of resolving non-absolute ``input`` URLs. Returns ``null`` if the parameters can't be resolved to a valid URL.

**Arguments**

* ``input`` (*string*): The absolute or relative input URL to parse. If ``input``
* ``base`` (*string*) — optional: The base URL to resolve against if the ``input`` is not


URL.toJSON
~~~~~~~~~~

.. code-block:: text

   url.toJSON()

*Kind:* function · *Arguments:* 0

The ``toJSON()`` method on the ``URL`` object returns the serialized URL. The value returned is equivalent to that of ``url.href`` and ``url.toString()``.

This method is automatically called when an ``URL`` object is serialized with ``JSON.stringify()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const myURLs = [
     new URL('https://www.example.com'),
     new URL('https://test.example.org'),
   ];
   console.log(JSON.stringify(myURLs));
   // Prints ["https://www.example.com/","https://test.example.org/"]


URL.toString
~~~~~~~~~~~~

.. code-block:: text

   url.toString()

*Kind:* function · *Arguments:* 0

The ``toString()`` method on the ``URL`` object returns the serialized URL. The value returned is equivalent to that of ``url.href`` and ``url.toJSON()``.

This callable takes no arguments.


URL.revokeObjectURL
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   URL.revokeObjectURL(id)

*Kind:* function · *Arguments:* 1

Removes the stored ``Blob`` identified by the given ID. Attempting to revoke a ID that isn't registered will silently fail.

**Arguments**

* ``id`` (*string*): A ``'blob:nodedata:...`` URL string returned by a prior call to


**Properties**

* ``URL.$params`` — object
* ``URL.auth`` — string
* ``URL.hash`` — string
* ``URL.host`` — string
* ``URL.hostname`` — string
* ``URL.href`` — string
* ``URL.origin`` — string
* ``URL.password`` — string
* ``URL.path`` — string
* ``URL.pathname`` — string
* ``URL.port`` — string
* ``URL.protocol`` — string
* ``URL.query`` — string
* ``URL.search`` — string
* ``URL.searchParams`` — object
* ``URL.slashes`` — boolean
* ``URL.username`` — string

URLPattern
----------

The WHATWG ``URLPattern`` class. See the ``URLPattern`` section.

URLPattern.exec
~~~~~~~~~~~~~~~

.. code-block:: text

   urlPattern.exec(input[, baseURL])

*Kind:* function · *Arguments:* 2

Input can be a string or an object providing the individual URL parts. The object members can be any of ``protocol``, ``username``, ``password``, ``hostname``, ``port``, ``pathname``, ``search``, ``hash`` or ``baseURL``.

If ``baseURL`` is not specified, it will default to ``undefined``.

**Arguments**

* ``input`` (*string | Object*): A URL or URL parts
* ``baseURL`` (*string | undefined*) — optional: A base URL string

**Example**

.. code-block:: javascript

   const myPattern = new URLPattern('https://example.com/docs/latest/api/*.html');
   console.log(myPattern.exec('https://example.com/docs/latest/api/dns.html'));
   // Prints:
   // {
   //  "hash": { "groups": {  "0": "" },  "input": "" },
   //  "hostname": { "groups": {}, "input": "example.com" },
   //  "inputs": [
   //    "https://example.com/docs/latest/api/dns.html"
   //  ],
   //  "password": { "groups": { "0": "" }, "input": "" },
   //  "pathname": { "groups": { "0": "dns" }, "input": "/docs/latest/api/dns.html" },
   //  "port": { "groups": {}, "input": "" },
   //  "protocol": { "groups": {}, "input": "https" },
   //  "search": { "groups": { "0": "" }, "input": "" },
   //  "username": { "groups": { "0": "" }, "input": "" }
   // }


URLPattern.test
~~~~~~~~~~~~~~~

.. code-block:: text

   urlPattern.test(input[, baseURL])

*Kind:* function · *Arguments:* 2

Input can be a string or an object providing the individual URL parts. The object members can be any of ``protocol``, ``username``, ``password``, ``hostname``, ``port``, ``pathname``, ``search``, ``hash`` or ``baseURL``.

If ``baseURL`` is not specified, it will default to ``undefined``.

**Arguments**

* ``input`` (*string | Object*): A URL or URL parts
* ``baseURL`` (*string | undefined*) — optional: A base URL string

**Example**

.. code-block:: javascript

   const myPattern = new URLPattern('https://example.com/docs/latest/api/*.html');
   console.log(myPattern.test('https://example.com/docs/latest/api/dns.html'));
   // Prints: true


**Properties**

* ``URLPattern.hasRegExpGroups`` — boolean

URLSearchParams
---------------

The WHATWG ``URLSearchParams`` class. See the ``URLSearchParams`` section.

URLSearchParams.append
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.append(name, value)

*Kind:* function · *Arguments:* 2

Append a new name-value pair to the query string.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*)


URLSearchParams.delete
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.delete(name[, value])

*Kind:* function · *Arguments:* 2

If ``value`` is provided, removes all name-value pairs where name is ``name`` and value is ``value``..

If ``value`` is not provided, removes all name-value pairs whose name is ``name``.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*) — optional


URLSearchParams.entries
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.entries()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over each of the name-value pairs in the query. Each item of the iterator is a JavaScript ``Array``. The first item of the ``Array`` is the ``name``, the second item of the ``Array`` is the ``value``.

Alias for ``urlSearchParams[Symbol.iterator]()``.

This callable takes no arguments.


URLSearchParams.forEach
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.forEach(fn[, thisArg])

*Kind:* function · *Arguments:* 2

Iterates over each name-value pair in the query and invokes the given function.

**Arguments**

* ``fn`` (*Function*): Invoked for each name-value pair in the query
* ``thisArg`` (*Object*) — optional: To be used as ``this`` value for when ``fn`` is called

**Example**

.. code-block:: javascript

   const myURL = new URL('https://example.org/?a=b&c=d');
   myURL.searchParams.forEach((value, name, searchParams) => {
     console.log(name, value, myURL.searchParams === searchParams);
   });
   // Prints:
   //   a b true
   //   c d true


URLSearchParams.get
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.get(name)

*Kind:* function · *Arguments:* 1

Returns the value of the first name-value pair whose name is ``name``. If there are no such pairs, ``null`` is returned.

**Arguments**

* ``name`` (*string*)


URLSearchParams.getAll
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.getAll(name)

*Kind:* function · *Arguments:* 1

Returns the values of all name-value pairs whose name is ``name``. If there are no such pairs, an empty array is returned.

**Arguments**

* ``name`` (*string*)


URLSearchParams.has
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.has(name[, value])

*Kind:* function · *Arguments:* 2

Checks if the ``URLSearchParams`` object contains key-value pair(s) based on ``name`` and an optional ``value`` argument.

If ``value`` is provided, returns ``true`` when name-value pair with same ``name`` and ``value`` exists.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*) — optional


URLSearchParams.keys
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.keys()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over the names of each name-value pair.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const params = new URLSearchParams('foo=bar&foo=baz');
   for (const name of params.keys()) {
     console.log(name);
   }
   // Prints:
   //   foo
   //   foo


URLSearchParams.set
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.set(name, value)

*Kind:* function · *Arguments:* 2

Sets the value in the ``URLSearchParams`` object associated with ``name`` to ``value``. If there are any pre-existing name-value pairs whose names are ``name``, set the first such pair's value to ``value`` and remove all others. If not, append the name-value pair to the query string.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*)

**Example**

.. code-block:: javascript

   const params = new URLSearchParams();
   params.append('foo', 'bar');
   params.append('foo', 'baz');
   params.append('abc', 'def');
   console.log(params.toString());
   // Prints foo=bar&foo=baz&abc=def

   params.set('foo', 'def');
   params.set('xyz', 'opq');
   console.log(params.toString());
   // Prints foo=def&abc=def&xyz=opq


URLSearchParams.sort
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.sort()

*Kind:* function · *Arguments:* 0

Sort all existing name-value pairs in-place by their names. Sorting is done with a stable sorting algorithm, so relative order between name-value pairs with the same name is preserved.

This method can be used, in particular, to increase cache hits.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const params = new URLSearchParams('query[]=abc&type=search&query[]=123');
   params.sort();
   console.log(params.toString());
   // Prints query%5B%5D=abc&query%5B%5D=123&type=search


URLSearchParams.toString
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.toString()

*Kind:* function · *Arguments:* 0

Returns the search parameters serialized as a string, with characters percent-encoded where necessary.

This callable takes no arguments.


URLSearchParams.values
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.values()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over the values of each name-value pair.

This callable takes no arguments.


**Properties**

* ``URLSearchParams.size`` — number

WebSocket
---------

A browser-compatible implementation of ``WebSocket``. Disable this API with the ``--no-experimental-websocket`` CLI flag.

WebSocket._deliver
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._deliver(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._dispatch
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._dispatch(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._fail
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._fail(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._finishClose
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._finishClose(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._handleFrame
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._handleFrame(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._handshakeRequest
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._handshakeRequest(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._onData
~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._onData(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._parseFrames
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._parseFrames(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._sendClose
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._sendClose(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket._validateHandshake
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket._validateHandshake(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket.close
~~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket.close(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


WebSocket.send
~~~~~~~~~~~~~~

.. code-block:: text

   globalThis.WebSocket.send(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


**Properties**

* ``WebSocket.CLOSED`` — number
* ``WebSocket.CLOSING`` — number
* ``WebSocket.CONNECTING`` — number
* ``WebSocket.OPEN`` — number
* ``WebSocket.CLOSED`` — number
* ``WebSocket.CLOSING`` — number
* ``WebSocket.CONNECTING`` — number
* ``WebSocket.onclose`` — null
* ``WebSocket.onerror`` — null
* ``WebSocket.onmessage`` — null
* ``WebSocket.onopen`` — null
* ``WebSocket.OPEN`` — number

WritableStream
--------------

A browser-compatible implementation of ``WritableStream``.

WritableStream.abort
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writableStream.abort([reason])

*Kind:* function · *Arguments:* 1

Abruptly terminates the ``WritableStream``. All queued writes will be canceled with their associated promises rejected.

**Arguments**

* ``reason`` (*any*) — optional


WritableStream.close
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writableStream.close()

*Kind:* function · *Arguments:* 0

Closes the ``WritableStream`` when no additional writes are expected.

This callable takes no arguments.


WritableStream.getWriter
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   writableStream.getWriter()

*Kind:* function · *Arguments:* 0

Creates and returns a new writer instance that can be used to write data into the ``WritableStream``.

This callable takes no arguments.

