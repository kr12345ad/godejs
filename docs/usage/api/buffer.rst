buffer
======

``Buffer`` objects are used to represent a fixed-length sequence of bytes. Many Node.js APIs support ``Buffer``\ s.

The ``Buffer`` class is a subclass of JavaScript's ``Uint8Array`` class and extends it with methods that cover additional use cases. Node.js APIs accept plain ``Uint8Array``\ s wherever ``Buffer``\ s are supported as well.

Loading
-------

.. code-block:: javascript

   const buffer = require('node:buffer');
   // ESM
   import * as buffer from 'node:buffer';

Quick start
-----------

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Creates a zero-filled Buffer of length 10.
   const buf1 = Buffer.alloc(10);

   // Creates a Buffer of length 10,
   // filled with bytes which all have the value `1`.
   const buf2 = Buffer.alloc(10, 1);

   // Creates an uninitialized buffer of length 10.
   // This is faster than calling Buffer.alloc() but the returned
   // Buffer instance might contain old data that needs to be
   // overwritten using fill(), write(), or other functions that fill the Buffer's
   // contents.
   const buf3 = Buffer.allocUnsafe(10);

   // Creates a Buffer containing the bytes [1, 2, 3].
   const buf4 = Buffer.from([1, 2, 3]);

   // Creates a Buffer containing the bytes [1, 1, 1, 1] – the entries
   // are all truncated using `(value & 255)` to fit into the range 0–255.
   const buf5 = Buffer.from([257, 257.5, -255, '1']);

   // Creates a Buffer containing the UTF-8-encoded bytes for the string 'tést':
   // [0x74, 0xc3, 0xa9, 0x73, 0x74] (in hexadecimal notation)
   // [116, 195, 169, 115, 116] (in decimal notation)
   const buf6 = Buffer.from('tést');

   // Creates a Buffer containing the Latin-1 bytes [0x74, 0xe9, 0x73, 0x74].
   const buf7 = Buffer.from('tést', 'latin1');

API
---

atob
~~~~

.. code-block:: text

   buffer.atob(data)

*Kind:* function · *Arguments:* 1

Decodes a string of Base64-encoded data into bytes, and encodes those bytes into a string using Latin-1 (ISO-8859-1).

The ``data`` may be any JavaScript-value that can be coerced into a string.

**Arguments**

* ``data`` (*any*): The Base64-encoded input string.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const result = buffer.atob('hello');
   console.log(result);


Blob
~~~~

.. code-block:: text

   buffer.Blob([sources[, options]])

*Kind:* class · *Arguments:* 2

Creates a new ``Blob`` object containing a concatenation of the given sources.

``ArrayBuffer``, ``TypedArray``, ``DataView``, and ``Buffer`` sources are copied into the 'Blob' and can therefore be safely modified after the 'Blob' is created.

**Arguments**

* ``sources`` (*string\[]|ArrayBuffer\[]|TypedArray\[]|DataView\[]|Blob\[]*) — optional: An
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob([], {});
   console.log(instance);


btoa
~~~~

.. code-block:: text

   buffer.btoa(data)

*Kind:* function · *Arguments:* 1

Decodes a string into bytes using Latin-1 (ISO-8859), and encodes those bytes into a string using Base64.

The ``data`` may be any JavaScript-value that can be coerced into a string.

**Arguments**

* ``data`` (*any*): An ASCII (Latin1) string.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const result = buffer.btoa('hello');
   console.log(result);


Buffer
~~~~~~

.. code-block:: text

   Buffer

*Kind:* class · *Arguments:* 3

The ``Buffer`` class is a global type for dealing with binary data directly. It can be constructed in a variety of ways.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.alloc(5);

   console.log(buf);
   // Prints: <Buffer 00 00 00 00 00>


File
~~~~

.. code-block:: text

   buffer.File(sources, fileName[, options])

*Kind:* class · *Arguments:* 3

**Arguments**

* ``sources`` (*string\[]|ArrayBuffer\[]|TypedArray\[]|DataView\[]|Blob\[]|File\[]*)
* ``fileName`` (*string*): The name of the file.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.File([], './file.txt', {});
   console.log(instance);


isAscii
~~~~~~~

.. code-block:: text

   buffer.isAscii(input)

*Kind:* function · *Arguments:* 1

This function returns ``true`` if ``input`` contains only valid ASCII-encoded data, including the case in which ``input`` is empty.

A detached ``ArrayBuffer``, or a ``TypedArray`` backed by one, is treated as empty.

**Arguments**

* ``input`` (*Buffer | ArrayBuffer | TypedArray*): The input to validate.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const result = buffer.isAscii([]);
   console.log(result);


isUtf8
~~~~~~

.. code-block:: text

   buffer.isUtf8(input)

*Kind:* function · *Arguments:* 1

This function returns ``true`` if ``input`` contains only valid UTF-8-encoded data, including the case in which ``input`` is empty.

A detached ``ArrayBuffer``, or a ``TypedArray`` backed by one, is treated as empty.

**Arguments**

* ``input`` (*Buffer | ArrayBuffer | TypedArray*): The input to validate.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const result = buffer.isUtf8([]);
   console.log(result);


resolveObjectURL
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.resolveObjectURL(id)

*Kind:* function · *Arguments:* 1

Resolves a ``'blob:nodedata:...'`` an associated ``Blob`` object registered using a prior call to ``URL.createObjectURL()``.

**Arguments**

* ``id`` (*string*): A ``'blob:nodedata:...`` URL string returned by a prior call to

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const result = buffer.resolveObjectURL('example');
   console.log(result);


SlowBuffer
~~~~~~~~~~

.. code-block:: text

   SlowBuffer

*Kind:* class · *Arguments:* 1

See ``Buffer.allocUnsafeSlow()``. This was never a class in the sense that the constructor always returned a ``Buffer`` instance, rather than a ``SlowBuffer`` instance.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.SlowBuffer(undefined);
   console.log(instance);


transcode
~~~~~~~~~

.. code-block:: text

   buffer.transcode(source, fromEnc, toEnc)

*Kind:* function · *Arguments:* 3

Re-encodes the given ``Buffer`` or ``Uint8Array`` instance from one character encoding to another. Returns a new ``Buffer`` instance.

Throws if the ``fromEnc`` or ``toEnc`` specify invalid character encodings or if conversion from ``fromEnc`` to ``toEnc`` is not permitted.

**Arguments**

* ``source`` (*Buffer|Uint8Array*): A ``Buffer`` or ``Uint8Array`` instance.
* ``fromEnc`` (*string*): The current encoding.
* ``toEnc`` (*string*): To target encoding.

**Example**

.. code-block:: javascript

   const { Buffer, transcode } = require('node:buffer');

   const newBuf = transcode(Buffer.from('€'), 'utf8', 'ascii');
   console.log(newBuf.toString('ascii'));
   // Prints: '?'


Blob class
----------

**Blob** — Creates a new ``Blob`` object containing a concatenation of the given sources.

``ArrayBuffer``, ``TypedArray``, ``DataView``, and ``Buffer`` sources are copied into the 'Blob' and can therefore be safely modified after the 'Blob' is created.

Blob.arrayBuffer
~~~~~~~~~~~~~~~~

.. code-block:: text

   blob.arrayBuffer()

*Kind:* function · *Arguments:* 0

Returns a promise that fulfills with an ``ArrayBuffer`` containing a copy of the ``Blob`` data.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob();
   const result = instance.arrayBuffer();
   console.log(result);


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

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob();
   const result = instance.slice(0, 0, 'example');
   console.log(result);


Blob.stream
~~~~~~~~~~~

.. code-block:: text

   blob.stream()

*Kind:* function · *Arguments:* 0

Returns a new ``ReadableStream`` that allows the content of the ``Blob`` to be read.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob();
   const result = instance.stream();
   console.log(result);


Blob.text
~~~~~~~~~

.. code-block:: text

   blob.text()

*Kind:* function · *Arguments:* 0

Returns a promise that fulfills with the contents of the ``Blob`` decoded as a UTF-8 string.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob();
   const result = instance.text();
   console.log(result);


Blob.textStream
~~~~~~~~~~~~~~~

.. code-block:: text

   blob.textStream()

*Kind:* function · *Arguments:* 0

Returns a new ``ReadableStream`` that allows the content of the ``Blob`` to be read as a stream of UTF-8 decoded strings. It is equivalent to piping ``blob.stream()`` through a ``TextDecoderStream`` set up with UTF-8.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Blob();
   const result = instance.textStream();
   console.log(result);


Buffer class
------------

**Buffer** — The ``Buffer`` class is a global type for dealing with binary data directly. It can be constructed in a variety of ways.

Buffer.asciiSlice
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.asciiSlice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.asciiSlice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.asciiWrite
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.asciiWrite(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.asciiWrite();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64Slice
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.base64Slice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.base64Slice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64urlSlice
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.base64urlSlice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.base64urlSlice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64urlWrite
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.base64urlWrite(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.base64urlWrite();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.base64Write
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.base64Write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.base64Write();
   console.log(result);

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

   buf.copy(target[, targetStart[, sourceStart[, sourceEnd]]])

*Kind:* function · *Arguments:* 4

Copies data from a region of ``buf`` to a region in ``target``, even if the ``target`` memory region overlaps with ``buf``.

``TypedArray.prototype.set()`` performs the same operation, and is available for all TypedArrays, including Node.js ``Buffer``\ s, although it takes different function arguments.

**Arguments**

* ``target`` (*Buffer|Uint8Array*): A ``Buffer`` or ``Uint8Array`` to copy into.
* ``targetStart`` (*integer*) — optional: The offset within ``target`` at which to begin
* ``sourceStart`` (*integer*) — optional: The offset within ``buf`` from which to begin copying.
* ``sourceEnd`` (*integer*) — optional: The offset within ``buf`` at which to stop copying (not

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Create two `Buffer` instances.
   const buf1 = Buffer.allocUnsafe(26);
   const buf2 = Buffer.allocUnsafe(26).fill('!');

   for (let i = 0; i < 26; i++) {
     // 97 is the decimal ASCII value for 'a'.
     buf1[i] = i + 97;
   }

   // Copy `buf1` bytes 16 through 19 into `buf2` starting at byte 8 of `buf2`.
   buf1.copy(buf2, 8, 16, 20);
   // This is equivalent to:
   // buf2.set(buf1.subarray(16, 20), 8);

   console.log(buf2.toString('ascii', 0, 25));
   // Prints: !!!!!!!!qrst!!!!!!!!!!!!!


Buffer.equals
~~~~~~~~~~~~~

.. code-block:: text

   buf.equals(otherBuffer)

*Kind:* function · *Arguments:* 1

Returns ``true`` if both ``buf`` and ``otherBuffer`` have exactly the same bytes, ``false`` otherwise. Equivalent to ``buf.compare(otherBuffer) === 0``.

**Arguments**

* ``otherBuffer`` (*Buffer|Uint8Array*): A ``Buffer`` or ``Uint8Array`` with which to

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from('ABC');
   const buf2 = Buffer.from('414243', 'hex');
   const buf3 = Buffer.from('ABCD');

   console.log(buf1.equals(buf2));
   // Prints: true
   console.log(buf1.equals(buf3));
   // Prints: false


Buffer.fill
~~~~~~~~~~~

.. code-block:: text

   buf.fill(value[, offset[, end]][, encoding])

*Kind:* function · *Arguments:* 4

Fills ``buf`` with the specified ``value``. If the ``offset`` and ``end`` are not given, the entire ``buf`` will be filled:

``value`` is coerced to a ``uint32`` value if it is not a string, ``Buffer``, or integer. If the resulting integer is greater than ``255`` (decimal), ``buf`` will be filled with ``value & 255``.

**Arguments**

* ``value`` (*string|Buffer|Uint8Array|integer*): The value with which to fill ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to fill ``buf``.
* ``end`` (*integer*) — optional: Where to stop filling ``buf`` (not inclusive). **Default:**
* ``encoding`` (*string*) — optional: The encoding for ``value`` if ``value`` is a string.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Fill a `Buffer` with the ASCII character 'h'.

   const b = Buffer.allocUnsafe(50).fill('h');

   console.log(b.toString());
   // Prints: hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh

   // Fill a buffer with empty string
   const c = Buffer.allocUnsafe(5).fill('');

   console.log(c.fill(''));
   // Prints: <Buffer 00 00 00 00 00>


Buffer.hexSlice
~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.hexSlice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.hexSlice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.hexWrite
~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.hexWrite(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.hexWrite();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.includes
~~~~~~~~~~~~~~~

.. code-block:: text

   buf.includes(value[, start[, end]][, encoding])

*Kind:* function · *Arguments:* 4

Equivalent to ``buf.indexOf() !== -1``.

**Arguments**

* ``value`` (*string|Buffer|Uint8Array|integer*): What to search for.
* ``start`` (*integer*) — optional: Where to begin searching in ``buf``. If negative, then
* ``end`` (*integer*) — optional: Where to stop searching in ``buf`` (exclusive). **Default:**
* ``encoding`` (*string*) — optional: If ``value`` is a string, this is its encoding.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from('this is a buffer');

   console.log(buf.includes('this'));
   // Prints: true
   console.log(buf.includes('is'));
   // Prints: true
   console.log(buf.includes(Buffer.from('a buffer')));
   // Prints: true
   console.log(buf.includes(97));
   // Prints: true (97 is the decimal ASCII value for 'a')
   console.log(buf.includes(Buffer.from('a buffer example')));
   // Prints: false
   console.log(buf.includes(Buffer.from('a buffer example').slice(0, 8)));
   // Prints: true
   console.log(buf.includes('this', 4));
   // Prints: false


Buffer.indexOf
~~~~~~~~~~~~~~

.. code-block:: text

   buf.indexOf(value[, start[, end]][, encoding])

*Kind:* function · *Arguments:* 4

If ``value`` is:

* a string, ``value`` is interpreted according to the character encoding in ``encoding``. * a ``Buffer`` or ``Uint8Array``, ``value`` will be used in its entirety. To compare a partial ``Buffer``, use ``buf.subarray``. * a number, ``value`` will be interpreted as an unsigned 8-bit integer value between ``0`` and ``255``.

**Arguments**

* ``value`` (*string|Buffer|Uint8Array|integer*): What to search for.
* ``start`` (*integer*) — optional: Where to begin searching in ``buf``. If negative, then
* ``end`` (*integer*) — optional: Where to stop searching in ``buf`` (exclusive). **Default:**
* ``encoding`` (*string*) — optional: If ``value`` is a string, this is the encoding used to

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from('this is a buffer');

   console.log(buf.indexOf('this'));
   // Prints: 0
   console.log(buf.indexOf('is'));
   // Prints: 2
   console.log(buf.indexOf(Buffer.from('a buffer')));
   // Prints: 8
   console.log(buf.indexOf(97));
   // Prints: 8 (97 is the decimal ASCII value for 'a')
   console.log(buf.indexOf(Buffer.from('a buffer example')));
   // Prints: -1
   console.log(buf.indexOf(Buffer.from('a buffer example').slice(0, 8)));
   // Prints: 8

   const utf16Buffer = Buffer.from('\u039a\u0391\u03a3\u03a3\u0395', 'utf16le');

   console.log(utf16Buffer.indexOf('\u03a3', 0, 'utf16le'));
   // Prints: 4
   console.log(utf16Buffer.indexOf('\u03a3', -4, 'utf16le'));
   // Prints: 6


Buffer.inspect
~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.inspect(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.inspect(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.lastIndexOf
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.lastIndexOf(value[, start[, end]][, encoding])

*Kind:* function · *Arguments:* 4

Identical to ``buf.indexOf()``, except the last occurrence of ``value`` is found rather than the first occurrence.

If ``value`` is not a string, number, or ``Buffer``, this method will throw a ``TypeError``. If ``value`` is a number, it will be coerced to a valid byte value, an integer between 0 and 255.

**Arguments**

* ``value`` (*string|Buffer|Uint8Array|integer*): What to search for.
* ``start`` (*integer*) — optional: Where to begin searching in ``buf``. If negative, then
* ``end`` (*integer*) — optional: Where to stop searching in ``buf`` (exclusive). **Default:**
* ``encoding`` (*string*) — optional: If ``value`` is a string, this is the encoding used to

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from('this buffer is a buffer');

   console.log(buf.lastIndexOf('this'));
   // Prints: 0
   console.log(buf.lastIndexOf('buffer'));
   // Prints: 17
   console.log(buf.lastIndexOf(Buffer.from('buffer')));
   // Prints: 17
   console.log(buf.lastIndexOf(97));
   // Prints: 15 (97 is the decimal ASCII value for 'a')
   console.log(buf.lastIndexOf(Buffer.from('yolo')));
   // Prints: -1
   console.log(buf.lastIndexOf('buffer', 5));
   // Prints: 5
   console.log(buf.lastIndexOf('buffer', 4));
   // Prints: -1

   const utf16Buffer = Buffer.from('\u039a\u0391\u03a3\u03a3\u0395', 'utf16le');

   console.log(utf16Buffer.lastIndexOf('\u03a3', undefined, 'utf16le'));
   // Prints: 6
   console.log(utf16Buffer.lastIndexOf('\u03a3', -5, 'utf16le'));
   // Prints: 4


Buffer.latin1Slice
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.latin1Slice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.latin1Slice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.latin1Write
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.latin1Write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.latin1Write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigInt64BE
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readBigInt64BE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, big-endian 64-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readBigInt64BE(0);
   console.log(result);


Buffer.readBigInt64LE
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readBigInt64LE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, little-endian 64-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readBigInt64LE(0);
   console.log(result);


Buffer.readBigUInt64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readBigUInt64BE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, big-endian 64-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readBigUint64BE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);

   console.log(buf.readBigUInt64BE(0));
   // Prints: 4294967295n


Buffer.readBigUint64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readBigUint64BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readBigUint64BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readBigUInt64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readBigUInt64LE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, little-endian 64-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readBigUint64LE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);

   console.log(buf.readBigUInt64LE(0));
   // Prints: 18446744069414584320n


Buffer.readBigUint64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readBigUint64LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readBigUint64LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readDoubleBE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readDoubleBE([offset])

*Kind:* function · *Arguments:* 1

Reads a 64-bit, big-endian double from ``buf`` at the specified ``offset``.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);

   console.log(buf.readDoubleBE(0));
   // Prints: 8.20788039913184e-304


Buffer.readDoubleLE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readDoubleLE([offset])

*Kind:* function · *Arguments:* 1

Reads a 64-bit, little-endian double from ``buf`` at the specified ``offset``.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);

   console.log(buf.readDoubleLE(0));
   // Prints: 5.447603722011605e-270
   console.log(buf.readDoubleLE(1));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readFloatBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readFloatBE([offset])

*Kind:* function · *Arguments:* 1

Reads a 32-bit, big-endian float from ``buf`` at the specified ``offset``.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([1, 2, 3, 4]);

   console.log(buf.readFloatBE(0));
   // Prints: 2.387939260590663e-38


Buffer.readFloatLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readFloatLE([offset])

*Kind:* function · *Arguments:* 1

Reads a 32-bit, little-endian float from ``buf`` at the specified ``offset``.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([1, 2, 3, 4]);

   console.log(buf.readFloatLE(0));
   // Prints: 1.539989614439558e-36
   console.log(buf.readFloatLE(1));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readInt16BE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readInt16BE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, big-endian 16-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0, 5]);

   console.log(buf.readInt16BE(0));
   // Prints: 5


Buffer.readInt16LE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readInt16LE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, little-endian 16-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0, 5]);

   console.log(buf.readInt16LE(0));
   // Prints: 1280
   console.log(buf.readInt16LE(1));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readInt32BE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readInt32BE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, big-endian 32-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0, 0, 0, 5]);

   console.log(buf.readInt32BE(0));
   // Prints: 5


Buffer.readInt32LE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readInt32LE([offset])

*Kind:* function · *Arguments:* 1

Reads a signed, little-endian 32-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0, 0, 0, 5]);

   console.log(buf.readInt32LE(0));
   // Prints: 83886080
   console.log(buf.readInt32LE(1));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readInt8
~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readInt8([offset])

*Kind:* function · *Arguments:* 1

Reads a signed 8-bit integer from ``buf`` at the specified ``offset``.

Integers read from a ``Buffer`` are interpreted as two's complement signed values.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([-1, 5]);

   console.log(buf.readInt8(0));
   // Prints: -1
   console.log(buf.readInt8(1));
   // Prints: 5
   console.log(buf.readInt8(2));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readIntBE
~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readIntBE(offset, byteLength)

*Kind:* function · *Arguments:* 2

Reads ``byteLength`` number of bytes from ``buf`` at the specified ``offset`` and interprets the result as a big-endian, two's complement signed value supporting up to 48 bits of accuracy.

**Arguments**

* ``offset`` (*integer*): Number of bytes to skip before starting to read. Must
* ``byteLength`` (*integer*): Number of bytes to read. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);

   console.log(buf.readIntBE(0, 6).toString(16));
   // Prints: 1234567890ab
   console.log(buf.readIntBE(1, 6).toString(16));
   // Throws ERR_OUT_OF_RANGE.
   console.log(buf.readIntBE(1, 0).toString(16));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readIntLE
~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readIntLE(offset, byteLength)

*Kind:* function · *Arguments:* 2

Reads ``byteLength`` number of bytes from ``buf`` at the specified ``offset`` and interprets the result as a little-endian, two's complement signed value supporting up to 48 bits of accuracy.

**Arguments**

* ``offset`` (*integer*): Number of bytes to skip before starting to read. Must
* ``byteLength`` (*integer*): Number of bytes to read. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);

   console.log(buf.readIntLE(0, 6).toString(16));
   // Prints: -546f87a9cbee


Buffer.readUInt16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUInt16BE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, big-endian 16-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readUint16BE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56]);

   console.log(buf.readUInt16BE(0).toString(16));
   // Prints: 1234
   console.log(buf.readUInt16BE(1).toString(16));
   // Prints: 3456


Buffer.readUint16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUint16BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUint16BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUInt16LE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, little-endian 16-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readUint16LE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56]);

   console.log(buf.readUInt16LE(0).toString(16));
   // Prints: 3412
   console.log(buf.readUInt16LE(1).toString(16));
   // Prints: 5634
   console.log(buf.readUInt16LE(2).toString(16));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readUint16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUint16LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUint16LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUInt32BE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, big-endian 32-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readUint32BE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78]);

   console.log(buf.readUInt32BE(0).toString(16));
   // Prints: 12345678


Buffer.readUint32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUint32BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUint32BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUInt32LE([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned, little-endian 32-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readUint32LE`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78]);

   console.log(buf.readUInt32LE(0).toString(16));
   // Prints: 78563412
   console.log(buf.readUInt32LE(1).toString(16));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readUint32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUint32LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUint32LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUInt8
~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUInt8([offset])

*Kind:* function · *Arguments:* 1

Reads an unsigned 8-bit integer from ``buf`` at the specified ``offset``.

This function is also available under the ``readUint8`` alias.

**Arguments**

* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to read. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([1, -2]);

   console.log(buf.readUInt8(0));
   // Prints: 1
   console.log(buf.readUInt8(1));
   // Prints: 254
   console.log(buf.readUInt8(2));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readUint8
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUint8(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUint8();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUIntBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUIntBE(offset, byteLength)

*Kind:* function · *Arguments:* 2

Reads ``byteLength`` number of bytes from ``buf`` at the specified ``offset`` and interprets the result as an unsigned big-endian integer supporting up to 48 bits of accuracy.

This function is also available under the ``readUintBE`` alias.

**Arguments**

* ``offset`` (*integer*): Number of bytes to skip before starting to read. Must
* ``byteLength`` (*integer*): Number of bytes to read. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);

   console.log(buf.readUIntBE(0, 6).toString(16));
   // Prints: 1234567890ab
   console.log(buf.readUIntBE(1, 6).toString(16));
   // Throws ERR_OUT_OF_RANGE.


Buffer.readUintBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUintBE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUintBE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.readUIntLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.readUIntLE(offset, byteLength)

*Kind:* function · *Arguments:* 2

Reads ``byteLength`` number of bytes from ``buf`` at the specified ``offset`` and interprets the result as an unsigned, little-endian integer supporting up to 48 bits of accuracy.

This function is also available under the ``readUintLE`` alias.

**Arguments**

* ``offset`` (*integer*): Number of bytes to skip before starting to read. Must
* ``byteLength`` (*integer*): Number of bytes to read. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);

   console.log(buf.readUIntLE(0, 6).toString(16));
   // Prints: ab9078563412


Buffer.readUintLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.readUintLE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.readUintLE();
   console.log(result);

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

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.slice(0, 0);
   console.log(result);


Buffer.subarray
~~~~~~~~~~~~~~~

.. code-block:: text

   buf.subarray([start[, end]])

*Kind:* function · *Arguments:* 2

Returns a new ``Buffer`` that references the same memory as the original, but offset and cropped by the ``start`` and ``end`` indexes.

Specifying ``end`` greater than ``buf.length`` will return the same result as that of ``end`` equal to ``buf.length``.

**Arguments**

* ``start`` (*integer*) — optional: Where the new ``Buffer`` will start. **Default:** ``0``.
* ``end`` (*integer*) — optional: Where the new ``Buffer`` will end (not inclusive).

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   // Create a `Buffer` with the ASCII alphabet, take a slice, and modify one byte
   // from the original `Buffer`.

   const buf1 = Buffer.allocUnsafe(26);

   for (let i = 0; i < 26; i++) {
     // 97 is the decimal ASCII value for 'a'.
     buf1[i] = i + 97;
   }

   const buf2 = buf1.subarray(0, 3);

   console.log(buf2.toString('ascii', 0, buf2.length));
   // Prints: abc

   buf1[0] = 33;

   console.log(buf2.toString('ascii', 0, buf2.length));
   // Prints: !bc


Buffer.swap16
~~~~~~~~~~~~~

.. code-block:: text

   buf.swap16()

*Kind:* function · *Arguments:* 0

Interprets ``buf`` as an array of unsigned 16-bit integers and swaps the byte order *in-place*. Throws ``ERR_INVALID_BUFFER_SIZE`` if ``buf.length`` is not a multiple of 2.

One convenient use of ``buf.swap16()`` is to perform a fast in-place conversion between UTF-16 little-endian and UTF-16 big-endian:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from([0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8]);

   console.log(buf1);
   // Prints: <Buffer 01 02 03 04 05 06 07 08>

   buf1.swap16();

   console.log(buf1);
   // Prints: <Buffer 02 01 04 03 06 05 08 07>

   const buf2 = Buffer.from([0x1, 0x2, 0x3]);

   buf2.swap16();
   // Throws ERR_INVALID_BUFFER_SIZE.


Buffer.swap32
~~~~~~~~~~~~~

.. code-block:: text

   buf.swap32()

*Kind:* function · *Arguments:* 0

Interprets ``buf`` as an array of unsigned 32-bit integers and swaps the byte order *in-place*. Throws ``ERR_INVALID_BUFFER_SIZE`` if ``buf.length`` is not a multiple of 4.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from([0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8]);

   console.log(buf1);
   // Prints: <Buffer 01 02 03 04 05 06 07 08>

   buf1.swap32();

   console.log(buf1);
   // Prints: <Buffer 04 03 02 01 08 07 06 05>

   const buf2 = Buffer.from([0x1, 0x2, 0x3]);

   buf2.swap32();
   // Throws ERR_INVALID_BUFFER_SIZE.


Buffer.swap64
~~~~~~~~~~~~~

.. code-block:: text

   buf.swap64()

*Kind:* function · *Arguments:* 0

Interprets ``buf`` as an array of 64-bit numbers and swaps byte order *in-place*. Throws ``ERR_INVALID_BUFFER_SIZE`` if ``buf.length`` is not a multiple of 8.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.from([0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8]);

   console.log(buf1);
   // Prints: <Buffer 01 02 03 04 05 06 07 08>

   buf1.swap64();

   console.log(buf1);
   // Prints: <Buffer 08 07 06 05 04 03 02 01>

   const buf2 = Buffer.from([0x1, 0x2, 0x3]);

   buf2.swap64();
   // Throws ERR_INVALID_BUFFER_SIZE.


Buffer.toJSON
~~~~~~~~~~~~~

.. code-block:: text

   buf.toJSON()

*Kind:* function · *Arguments:* 0

Returns a JSON representation of ``buf``. ``JSON.stringify()`` implicitly calls this function when stringifying a ``Buffer`` instance.

``Buffer.from()`` accepts objects in the format returned from this method. In particular, ``Buffer.from(buf.toJSON())`` works like ``Buffer.from(buf)``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.from([0x1, 0x2, 0x3, 0x4, 0x5]);
   const json = JSON.stringify(buf);

   console.log(json);
   // Prints: {"type":"Buffer","data":[1,2,3,4,5]}

   const copy = JSON.parse(json, (key, value) => {
     return value && value.type === 'Buffer' ?
       Buffer.from(value) :
       value;
   });

   console.log(copy);
   // Prints: <Buffer 01 02 03 04 05>


Buffer.toLocaleString
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.toLocaleString(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.toLocaleString();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.toString
~~~~~~~~~~~~~~~

.. code-block:: text

   buf.toString([encoding[, start[, end]]])

*Kind:* function · *Arguments:* 3

Decodes ``buf`` to a string according to the specified character encoding in ``encoding``. ``start`` and ``end`` may be passed to decode only a subset of ``buf``.

If ``encoding`` is ``'utf8'`` and a byte sequence in the input is not valid UTF-8, then each invalid byte is replaced with the replacement character ``U+FFFD``.

**Arguments**

* ``encoding`` (*string*) — optional: The character encoding to use. **Default:** ``'utf8'``.
* ``start`` (*integer*) — optional: The byte offset to start decoding at. **Default:** ``0``.
* ``end`` (*integer*) — optional: The byte offset to stop decoding at (not inclusive).

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf1 = Buffer.allocUnsafe(26);

   for (let i = 0; i < 26; i++) {
     // 97 is the decimal ASCII value for 'a'.
     buf1[i] = i + 97;
   }

   console.log(buf1.toString('utf8'));
   // Prints: abcdefghijklmnopqrstuvwxyz
   console.log(buf1.toString('utf8', 0, 5));
   // Prints: abcde

   const buf2 = Buffer.from('tést');

   console.log(buf2.toString('hex'));
   // Prints: 74c3a97374
   console.log(buf2.toString('utf8', 0, 3));
   // Prints: té
   console.log(buf2.toString(undefined, 0, 3));
   // Prints: té


Buffer.ucs2Slice
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.ucs2Slice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.ucs2Slice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.ucs2Write
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.ucs2Write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.ucs2Write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.utf8Slice
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.utf8Slice(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.utf8Slice();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.utf8Write
~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.utf8Write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.utf8Write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.write
~~~~~~~~~~~~

.. code-block:: text

   buf.write(string[, offset[, length]][, encoding])

*Kind:* function · *Arguments:* 4

Writes ``string`` to ``buf`` at ``offset`` according to the character encoding in ``encoding``. The ``length`` parameter is the number of bytes to write. If ``buf`` did not contain enough space to fit the entire string, only part of ``string`` will be written. However, partially encoded characters will not be written.

**Arguments**

* ``string`` (*string*): String to write to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write ``string``.
* ``length`` (*integer*) — optional: Maximum number of bytes to write (written bytes will not
* ``encoding`` (*string*) — optional: The character encoding of ``string``. **Default:** ``'utf8'``.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.alloc(256);

   const len = buf.write('\u00bd + \u00bc = \u00be', 0);

   console.log(`${len} bytes: ${buf.toString('utf8', 0, len)}`);
   // Prints: 12 bytes: ½ + ¼ = ¾

   const buffer = Buffer.alloc(10);

   const length = buffer.write('abcd', 8);

   console.log(`${length} bytes: ${buffer.toString('utf8', 8, 10)}`);
   // Prints: 2 bytes : ab


Buffer.writeBigInt64BE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeBigInt64BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian.

``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*bigint*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeBigInt64BE(0x0102030405060708n, 0);

   console.log(buf);
   // Prints: <Buffer 01 02 03 04 05 06 07 08>


Buffer.writeBigInt64LE
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeBigInt64LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian.

``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*bigint*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeBigInt64LE(0x0102030405060708n, 0);

   console.log(buf);
   // Prints: <Buffer 08 07 06 05 04 03 02 01>


Buffer.writeBigUInt64BE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeBigUInt64BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian.

This function is also available under the ``writeBigUint64BE`` alias.

**Arguments**

* ``value`` (*bigint*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeBigUInt64BE(0xdecafafecacefaden, 0);

   console.log(buf);
   // Prints: <Buffer de ca fa fe ca ce fa de>


Buffer.writeBigUint64BE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeBigUint64BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeBigUint64BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeBigUInt64LE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeBigUInt64LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian

This function is also available under the ``writeBigUint64LE`` alias.

**Arguments**

* ``value`` (*bigint*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeBigUInt64LE(0xdecafafecacefaden, 0);

   console.log(buf);
   // Prints: <Buffer de fa ce ca fe fa ca de>


Buffer.writeBigUint64LE
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeBigUint64LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeBigUint64LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeDoubleBE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeDoubleBE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian. The ``value`` must be a JavaScript number. Behavior is undefined when ``value`` is anything other than a JavaScript number.

**Arguments**

* ``value`` (*number*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeDoubleBE(123.456, 0);

   console.log(buf);
   // Prints: <Buffer 40 5e dd 2f 1a 9f be 77>


Buffer.writeDoubleLE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeDoubleLE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian. The ``value`` must be a JavaScript number. Behavior is undefined when ``value`` is anything other than a JavaScript number.

**Arguments**

* ``value`` (*number*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(8);

   buf.writeDoubleLE(123.456, 0);

   console.log(buf);
   // Prints: <Buffer 77 be 9f 1a 2f dd 5e 40>


Buffer.writeFloatBE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeFloatBE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian. Behavior is undefined when ``value`` is anything other than a JavaScript number.

**Arguments**

* ``value`` (*number*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeFloatBE(0xcafebabe, 0);

   console.log(buf);
   // Prints: <Buffer 4f 4a fe bb>


Buffer.writeFloatLE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeFloatLE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian. Behavior is undefined when ``value`` is anything other than a JavaScript number.

**Arguments**

* ``value`` (*number*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeFloatLE(0xcafebabe, 0);

   console.log(buf);
   // Prints: <Buffer bb fe 4a 4f>


Buffer.writeInt16BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeInt16BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian.  The ``value`` must be a valid signed 16-bit integer. Behavior is undefined when ``value`` is anything other than a signed 16-bit integer.

The ``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(2);

   buf.writeInt16BE(0x0102, 0);

   console.log(buf);
   // Prints: <Buffer 01 02>


Buffer.writeInt16LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeInt16LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian.  The ``value`` must be a valid signed 16-bit integer. Behavior is undefined when ``value`` is anything other than a signed 16-bit integer.

The ``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(2);

   buf.writeInt16LE(0x0304, 0);

   console.log(buf);
   // Prints: <Buffer 04 03>


Buffer.writeInt32BE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeInt32BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian. The ``value`` must be a valid signed 32-bit integer. Behavior is undefined when ``value`` is anything other than a signed 32-bit integer.

The ``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeInt32BE(0x01020304, 0);

   console.log(buf);
   // Prints: <Buffer 01 02 03 04>


Buffer.writeInt32LE
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeInt32LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian. The ``value`` must be a valid signed 32-bit integer. Behavior is undefined when ``value`` is anything other than a signed 32-bit integer.

The ``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeInt32LE(0x05060708, 0);

   console.log(buf);
   // Prints: <Buffer 08 07 06 05>


Buffer.writeInt8
~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeInt8(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset``. ``value`` must be a valid signed 8-bit integer. Behavior is undefined when ``value`` is anything other than a signed 8-bit integer.

``value`` is interpreted and written as a two's complement signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(2);

   buf.writeInt8(2, 0);
   buf.writeInt8(-2, 1);

   console.log(buf);
   // Prints: <Buffer 02 fe>


Buffer.writeIntBE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeIntBE(value, offset, byteLength)

*Kind:* function · *Arguments:* 3

Writes ``byteLength`` bytes of ``value`` to ``buf`` at the specified ``offset`` as big-endian. Supports up to 48 bits of accuracy. Behavior is undefined when ``value`` is anything other than a signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*): Number of bytes to skip before starting to write. Must
* ``byteLength`` (*integer*): Number of bytes to write. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(6);

   buf.writeIntBE(0x1234567890ab, 0, 6);

   console.log(buf);
   // Prints: <Buffer 12 34 56 78 90 ab>


Buffer.writeIntLE
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeIntLE(value, offset, byteLength)

*Kind:* function · *Arguments:* 3

Writes ``byteLength`` bytes of ``value`` to ``buf`` at the specified ``offset`` as little-endian. Supports up to 48 bits of accuracy. Behavior is undefined when ``value`` is anything other than a signed integer.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*): Number of bytes to skip before starting to write. Must
* ``byteLength`` (*integer*): Number of bytes to write. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(6);

   buf.writeIntLE(0x1234567890ab, 0, 6);

   console.log(buf);
   // Prints: <Buffer ab 90 78 56 34 12>


Buffer.writeUInt16BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUInt16BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian. The ``value`` must be a valid unsigned 16-bit integer. Behavior is undefined when ``value`` is anything other than an unsigned 16-bit integer.

This function is also available under the ``writeUint16BE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeUInt16BE(0xdead, 0);
   buf.writeUInt16BE(0xbeef, 2);

   console.log(buf);
   // Prints: <Buffer de ad be ef>


Buffer.writeUint16BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUint16BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUint16BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt16LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUInt16LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian. The ``value`` must be a valid unsigned 16-bit integer. Behavior is undefined when ``value`` is anything other than an unsigned 16-bit integer.

This function is also available under the ``writeUint16LE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeUInt16LE(0xdead, 0);
   buf.writeUInt16LE(0xbeef, 2);

   console.log(buf);
   // Prints: <Buffer ad de ef be>


Buffer.writeUint16LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUint16LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUint16LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt32BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUInt32BE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as big-endian. The ``value`` must be a valid unsigned 32-bit integer. Behavior is undefined when ``value`` is anything other than an unsigned 32-bit integer.

This function is also available under the ``writeUint32BE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeUInt32BE(0xfeedface, 0);

   console.log(buf);
   // Prints: <Buffer fe ed fa ce>


Buffer.writeUint32BE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUint32BE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUint32BE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt32LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUInt32LE(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset`` as little-endian. The ``value`` must be a valid unsigned 32-bit integer. Behavior is undefined when ``value`` is anything other than an unsigned 32-bit integer.

This function is also available under the ``writeUint32LE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeUInt32LE(0xfeedface, 0);

   console.log(buf);
   // Prints: <Buffer ce fa ed fe>


Buffer.writeUint32LE
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUint32LE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUint32LE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUInt8
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUInt8(value[, offset])

*Kind:* function · *Arguments:* 2

Writes ``value`` to ``buf`` at the specified ``offset``. ``value`` must be a valid unsigned 8-bit integer. Behavior is undefined when ``value`` is anything other than an unsigned 8-bit integer.

This function is also available under the ``writeUint8`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*) — optional: Number of bytes to skip before starting to write. Must

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(4);

   buf.writeUInt8(0x3, 0);
   buf.writeUInt8(0x4, 1);
   buf.writeUInt8(0x23, 2);
   buf.writeUInt8(0x42, 3);

   console.log(buf);
   // Prints: <Buffer 03 04 23 42>


Buffer.writeUint8
~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUint8(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUint8();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUIntBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUIntBE(value, offset, byteLength)

*Kind:* function · *Arguments:* 3

Writes ``byteLength`` bytes of ``value`` to ``buf`` at the specified ``offset`` as big-endian. Supports up to 48 bits of accuracy. Behavior is undefined when ``value`` is anything other than an unsigned integer.

This function is also available under the ``writeUintBE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*): Number of bytes to skip before starting to write. Must
* ``byteLength`` (*integer*): Number of bytes to write. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(6);

   buf.writeUIntBE(0x1234567890ab, 0, 6);

   console.log(buf);
   // Prints: <Buffer 12 34 56 78 90 ab>


Buffer.writeUintBE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUintBE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUintBE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Buffer.writeUIntLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buf.writeUIntLE(value, offset, byteLength)

*Kind:* function · *Arguments:* 3

Writes ``byteLength`` bytes of ``value`` to ``buf`` at the specified ``offset`` as little-endian. Supports up to 48 bits of accuracy. Behavior is undefined when ``value`` is anything other than an unsigned integer.

This function is also available under the ``writeUintLE`` alias.

**Arguments**

* ``value`` (*integer*): Number to be written to ``buf``.
* ``offset`` (*integer*): Number of bytes to skip before starting to write. Must
* ``byteLength`` (*integer*): Number of bytes to write. Must satisfy

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');

   const buf = Buffer.allocUnsafe(6);

   buf.writeUIntLE(0x1234567890ab, 0, 6);

   console.log(buf);
   // Prints: <Buffer ab 90 78 56 34 12>


Buffer.writeUintLE
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   buffer.Buffer.writeUintLE(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const buffer = require('node:buffer');

   const instance = new buffer.Buffer();
   const result = instance.writeUintLE();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``Blob.type`` — undefined
* ``Buffer.offset`` — undefined
* ``Buffer.parent`` — undefined
* ``constants`` — object
* ``constants.MAX_LENGTH`` — number
* ``constants.MAX_STRING_LENGTH`` — number
* ``File.lastModified`` — undefined
* ``File.name`` — undefined
* ``INSPECT_MAX_BYTES`` — number
* ``kMaxLength`` — number
* ``kStringMaxLength`` — number
* ``MAX_LENGTH`` — number
* ``MAX_STRING_LENGTH`` — number

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
