zlib
====

The ``node:zlib`` module provides compression functionality implemented using Gzip, Deflate/Inflate, Brotli, and Zstd.

To access it:

Loading
-------

.. code-block:: javascript

   const zlib = require('node:zlib');
   // ESM
   import * as zlib from 'node:zlib';

Quick start
-----------

.. code-block:: javascript

   const zlib = require('node:zlib');

API
---

BrotliCompress
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.BrotliCompress

*Kind:* class · *Arguments:* 1

Compress data using the Brotli algorithm.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibbrotlicompress>`__

brotliCompress
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.brotliCompress(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*brotli options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibbrotlicompressbuffer-options-callback>`__

brotliCompressSync
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.brotliCompressSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Compress a chunk of data with ``BrotliCompress``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*brotli options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibbrotlicompresssyncbuffer-options>`__

BrotliDecompress
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.BrotliDecompress

*Kind:* class · *Arguments:* 1

Decompress data using the Brotli algorithm.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibbrotlidecompress>`__

brotliDecompress
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.brotliDecompress(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*brotli options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibbrotlidecompressbuffer-options-callback>`__

brotliDecompressSync
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.brotliDecompressSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``BrotliDecompress``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*brotli options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibbrotlidecompresssyncbuffer-options>`__

crc32
~~~~~

.. code-block:: text

   zlib.crc32(data[, value])

*Kind:* function · *Arguments:* 2

Computes a 32-bit Cyclic Redundancy Check checksum of ``data``. If ``value`` is specified, it is used as the starting value of the checksum, otherwise, 0 is used as the starting value.

The CRC algorithm is designed to compute checksums and to detect error in data transmission. It's not suitable for cryptographic authentication.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*): When ``data`` is a string,
* ``value`` (*integer*) — optional: An optional starting value. It must be a 32-bit unsigned

**Example**

.. code-block:: javascript

   const zlib = require('node:zlib');
   const { Buffer } = require('node:buffer');

   let crc = zlib.crc32('hello');  // 907060870
   crc = zlib.crc32('world', crc);  // 4192936109

   crc = zlib.crc32(Buffer.from('hello', 'utf16le'));  // 1427272415
   crc = zlib.crc32(Buffer.from('world', 'utf16le'), crc);  // 4150509955

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcrc32data-value>`__

createBrotliCompress
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createBrotliCompress([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``BrotliCompress`` object.

**Arguments**

* ``options`` (*brotli options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatebrotlicompressoptions>`__

createBrotliDecompress
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createBrotliDecompress([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``BrotliDecompress`` object.

**Arguments**

* ``options`` (*brotli options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatebrotlidecompressoptions>`__

createDeflate
~~~~~~~~~~~~~

.. code-block:: text

   zlib.createDeflate([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``Deflate`` object.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatedeflateoptions>`__

createDeflateRaw
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createDeflateRaw([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``DeflateRaw`` object.

An upgrade of zlib from 1.2.8 to 1.2.11 changed behavior when ``windowBits`` is set to 8 for raw deflate streams. zlib would automatically set ``windowBits`` to 9 if was initially set to 8. Newer versions of zlib will throw an exception, so Node.js restored the original behavior of upgrading a value of 8 to 9, since passing ``windowBits = 9`` to zlib actually results in a compressed stream that effectively uses an 8-bit window only.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatedeflaterawoptions>`__

createGunzip
~~~~~~~~~~~~

.. code-block:: text

   zlib.createGunzip([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``Gunzip`` object.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreategunzipoptions>`__

createGzip
~~~~~~~~~~

.. code-block:: text

   zlib.createGzip([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``Gzip`` object. See example.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreategzipoptions>`__

createInflate
~~~~~~~~~~~~~

.. code-block:: text

   zlib.createInflate([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``Inflate`` object.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreateinflateoptions>`__

createInflateRaw
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createInflateRaw([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``InflateRaw`` object.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreateinflaterawoptions>`__

createUnzip
~~~~~~~~~~~

.. code-block:: text

   zlib.createUnzip([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``Unzip`` object.

**Arguments**

* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreateunzipoptions>`__

createZstdCompress
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createZstdCompress([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``ZstdCompress`` object.

**Arguments**

* ``options`` (*zstd options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatezstdcompressoptions>`__

createZstdDecompress
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.createZstdDecompress([options])

*Kind:* function · *Arguments:* 1

Creates and returns a new ``ZstdDecompress`` object.

**Arguments**

* ``options`` (*zstd options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibcreatezstddecompressoptions>`__

Deflate
~~~~~~~

.. code-block:: text

   zlib.Deflate

*Kind:* class · *Arguments:* 1

Compress data using deflate.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibdeflate>`__

deflate
~~~~~~~

.. code-block:: text

   zlib.deflate(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibdeflatebuffer-options-callback>`__

DeflateRaw
~~~~~~~~~~

.. code-block:: text

   zlib.DeflateRaw

*Kind:* class · *Arguments:* 1

Compress data using deflate, and do not append a ``zlib`` header.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibdeflateraw>`__

deflateRaw
~~~~~~~~~~

.. code-block:: text

   zlib.deflateRaw(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibdeflaterawbuffer-options-callback>`__

deflateRawSync
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.deflateRawSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Compress a chunk of data with ``DeflateRaw``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibdeflaterawsyncbuffer-options>`__

deflateSync
~~~~~~~~~~~

.. code-block:: text

   zlib.deflateSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Compress a chunk of data with ``Deflate``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibdeflatesyncbuffer-options>`__

Gunzip
~~~~~~

.. code-block:: text

   zlib.Gunzip

*Kind:* class · *Arguments:* 1

Decompress a gzip stream.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibgunzip>`__

gunzip
~~~~~~

.. code-block:: text

   zlib.gunzip(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibgunzipbuffer-options-callback>`__

gunzipSync
~~~~~~~~~~

.. code-block:: text

   zlib.gunzipSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``Gunzip``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibgunzipsyncbuffer-options>`__

Gzip
~~~~

.. code-block:: text

   zlib.Gzip

*Kind:* class · *Arguments:* 1

Compress data using gzip.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibgzip>`__

gzip
~~~~

.. code-block:: text

   zlib.gzip(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibgzipbuffer-options-callback>`__

gzipSync
~~~~~~~~

.. code-block:: text

   zlib.gzipSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Compress a chunk of data with ``Gzip``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibgzipsyncbuffer-options>`__

Inflate
~~~~~~~

.. code-block:: text

   zlib.Inflate

*Kind:* class · *Arguments:* 1

Decompress a deflate stream.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibinflate>`__

inflate
~~~~~~~

.. code-block:: text

   zlib.inflate(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibinflatebuffer-options-callback>`__

InflateRaw
~~~~~~~~~~

.. code-block:: text

   zlib.InflateRaw

*Kind:* class · *Arguments:* 1

Decompress a raw deflate stream.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibinflateraw>`__

inflateRaw
~~~~~~~~~~

.. code-block:: text

   zlib.inflateRaw(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibinflaterawbuffer-options-callback>`__

inflateRawSync
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.inflateRawSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``InflateRaw``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibinflaterawsyncbuffer-options>`__

inflateSync
~~~~~~~~~~~

.. code-block:: text

   zlib.inflateSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``Inflate``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibinflatesyncbuffer-options>`__

Unzip
~~~~~

.. code-block:: text

   zlib.Unzip

*Kind:* class · *Arguments:* 1

Decompress either a Gzip- or Deflate-compressed stream by auto-detecting the header.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibunzip>`__

unzip
~~~~~

.. code-block:: text

   zlib.unzip(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibunzipbuffer-options-callback>`__

unzipSync
~~~~~~~~~

.. code-block:: text

   zlib.unzipSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``Unzip``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zlib options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibunzipsyncbuffer-options>`__

ZlibBase
~~~~~~~~

.. code-block:: text

   zlib.ZlibBase

*Kind:* class · *Arguments:* 2

Not exported by the ``node:zlib`` module. It is documented here because it is the base class of the compressor/decompressor classes.

This class inherits from ``stream.Transform``, allowing ``node:zlib`` objects to be used in pipes and similar stream operations.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibzlibbase>`__

ZstdCompress
~~~~~~~~~~~~

.. code-block:: text

   zlib.ZstdCompress

*Kind:* class · *Arguments:* 1

Compress data using the Zstd algorithm.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibzstdcompress>`__

zstdCompress
~~~~~~~~~~~~

.. code-block:: text

   zlib.zstdCompress(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zstd options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibzstdcompressbuffer-options-callback>`__

zstdCompressSync
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.zstdCompressSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Compress a chunk of data with ``ZstdCompress``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zstd options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibzstdcompresssyncbuffer-options>`__

ZstdDecompress
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZstdDecompress

*Kind:* class · *Arguments:* 1

Decompress data using the Zstd algorithm. Concatenated Zstd and skippable frames are decoded as a single stream.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#class-zlibzstddecompress>`__

zstdDecompress
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.zstdDecompress(buffer[, options], callback)

*Kind:* function · *Arguments:* 3

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zstd options*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibzstddecompressbuffer-options-callback>`__

zstdDecompressSync
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.zstdDecompressSync(buffer[, options])

*Kind:* function · *Arguments:* 2

Decompress a chunk of data with ``ZstdDecompress``.

**Arguments**

* ``buffer`` (*Buffer|TypedArray|DataView|ArrayBuffer|string*)
* ``options`` (*zstd options*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibzstddecompresssyncbuffer-options>`__

ZlibBase class
--------------

**ZlibBase** — Not exported by the ``node:zlib`` module. It is documented here because it is the base class of the compressor/decompressor classes.

This class inherits from ``stream.Transform``, allowing ``node:zlib`` objects to be used in pipes and similar stream operations.

ZlibBase._destroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._destroy(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._emitClose
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._emitClose(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._emitData
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._emitData(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._final
~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._final(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._finishOutput
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._finishOutput(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._flush
~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._flush(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._maybeDrain
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._maybeDrain(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._processChunk
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._processChunk(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase._transform
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase._transform(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.addListener
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.addListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.close
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.close([callback])

*Kind:* function · *Arguments:* 1

Close the underlying handle.

**Arguments**

* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibclosecallback>`__

ZlibBase.destroy
~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.destroy(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.emit
~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.emit(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.end
~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.eventNames
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.eventNames(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.flush
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.flush([kind, ]callback)

*Kind:* function · *Arguments:* 2

Flush pending data. Don't call this frivolously, premature flushes negatively impact the effectiveness of the compression algorithm.

Calling this only flushes data from the internal ``zlib`` state, and does not perform flushing of any kind on the streams level. Rather, it behaves like a normal call to ``.write()``, i.e. it will be queued up behind other pending writes and will only produce output when data is being read from the stream.

**Arguments**

* ``kind`` — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibflushkind-callback>`__

ZlibBase.isPaused
~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.isPaused(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.listenerCount
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.listenerCount(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.listeners
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.listeners(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.off
~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.off(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.on
~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.on(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.once
~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.once(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.params
~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.params(level, strategy, callback)

*Kind:* function · *Arguments:* 3

This function is only available for zlib-based streams, i.e. not Brotli.

Dynamically update the compression level and compression strategy. Only applicable to deflate algorithm.

**Arguments**

* ``level`` (*integer*)
* ``strategy`` (*integer*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibparamslevel-strategy-callback>`__

ZlibBase.pause
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.pipe
~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.pipe(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.prependListener
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.prependListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.push
~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.push(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.read
~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.read(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.removeAllListeners(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.removeListener
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.removeListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.reset
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.reset()

*Kind:* function · *Arguments:* 0

For inflate and deflate streams, reset the compressor/decompressor to factory defaults.

For Zstd streams, cancel the current frame and start a new session while preserving the configured parameters and dictionary. If ``pledgedSrcSize`` was configured for a Zstd compressor, it applies again to the next frame.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/zlib.html#zlibreset>`__

ZlibBase.resume
~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.setEncoding
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.setEncoding(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.unpipe
~~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.unpipe(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ZlibBase.write
~~~~~~~~~~~~~~

.. code-block:: text

   zlib.ZlibBase.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Properties and constants
------------------------

* ``codes`` — object
* ``codes.-1`` — string
* ``codes.-2`` — string
* ``codes.-3`` — string
* ``codes.-4`` — string
* ``codes.-5`` — string
* ``codes.-6`` — string
* ``codes.0`` — string
* ``codes.1`` — string
* ``codes.2`` — string
* ``codes.Z_BUF_ERROR`` — number
* ``codes.Z_DATA_ERROR`` — number
* ``codes.Z_ERRNO`` — number
* ``codes.Z_MEM_ERROR`` — number
* ``codes.Z_NEED_DICT`` — number
* ``codes.Z_OK`` — number
* ``codes.Z_STREAM_END`` — number
* ``codes.Z_STREAM_ERROR`` — number
* ``codes.Z_VERSION_ERROR`` — number
* ``constants`` — object
* ``constants.BROTLI_DECODE`` — number
* ``constants.BROTLI_DECODER_PARAM_DISABLE_RING_BUFFER_REALLOCATION`` — number
* ``constants.BROTLI_DECODER_PARAM_LARGE_WINDOW`` — number
* ``constants.BROTLI_DECODER_RESULT_ERROR`` — number
* ``constants.BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT`` — number
* ``constants.BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT`` — number
* ``constants.BROTLI_DECODER_RESULT_SUCCESS`` — number
* ``constants.BROTLI_DEFAULT_MODE`` — number
* ``constants.BROTLI_DEFAULT_QUALITY`` — number
* ``constants.BROTLI_DEFAULT_WINDOW`` — number
* ``constants.BROTLI_ENCODE`` — number
* ``constants.BROTLI_MAX_INPUT_BLOCK_BITS`` — number
* ``constants.BROTLI_MAX_QUALITY`` — number
* ``constants.BROTLI_MAX_WINDOW_BITS`` — number
* ``constants.BROTLI_MIN_INPUT_BLOCK_BITS`` — number
* ``constants.BROTLI_MIN_QUALITY`` — number
* ``constants.BROTLI_MIN_WINDOW_BITS`` — number
* ``constants.BROTLI_MODE_FONT`` — number
* ``constants.BROTLI_MODE_GENERIC`` — number
* ``constants.BROTLI_MODE_TEXT`` — number
* ``constants.BROTLI_OPERATION_EMIT_METADATA`` — number
* ``constants.BROTLI_OPERATION_FINISH`` — number
* ``constants.BROTLI_OPERATION_FLUSH`` — number
* ``constants.BROTLI_OPERATION_PROCESS`` — number
* ``constants.BROTLI_PARAM_DISABLE_LITERAL_CONTEXT_MODELING`` — number
* ``constants.BROTLI_PARAM_LARGE_WINDOW`` — number
* ``constants.BROTLI_PARAM_LGBLOCK`` — number
* ``constants.BROTLI_PARAM_LGWIN`` — number
* ``constants.BROTLI_PARAM_MODE`` — number
* ``constants.BROTLI_PARAM_NDIRECT`` — number
* ``constants.BROTLI_PARAM_NPOSTFIX`` — number
* ``constants.BROTLI_PARAM_QUALITY`` — number
* ``constants.BROTLI_PARAM_SIZE_HINT`` — number
* ``constants.DEFLATE`` — number
* ``constants.DEFLATERAW`` — number
* ``constants.GUNZIP`` — number
* ``constants.GZIP`` — number
* ``constants.INFLATE`` — number
* ``constants.INFLATERAW`` — number
* ``constants.UNZIP`` — number
* ``constants.Z_ASCII`` — number
* ``constants.Z_BEST_COMPRESSION`` — number
* ``constants.Z_BEST_SPEED`` — number
* ``constants.Z_BINARY`` — number
* ``constants.Z_BLOCK`` — number
* ``constants.Z_BUF_ERROR`` — number
* ``constants.Z_DATA_ERROR`` — number
* ``constants.Z_DEFAULT_CHUNK`` — number
* ``constants.Z_DEFAULT_COMPRESSION`` — number
* ``constants.Z_DEFAULT_LEVEL`` — number
* ``constants.Z_DEFAULT_MEMLEVEL`` — number
* ``constants.Z_DEFAULT_STRATEGY`` — number
* ``constants.Z_DEFAULT_WINDOWBITS`` — number
* ``constants.Z_DEFLATED`` — number
* ``constants.Z_ERRNO`` — number
* ``constants.Z_FILTERED`` — number
* ``constants.Z_FINISH`` — number
* ``constants.Z_FIXED`` — number
* ``constants.Z_FULL_FLUSH`` — number
* ``constants.Z_HUFFMAN_ONLY`` — number
* ``constants.Z_MAX_CHUNK`` — number
* ``constants.Z_MAX_LEVEL`` — number
* ``constants.Z_MAX_MEMLEVEL`` — number
* ``constants.Z_MAX_WINDOWBITS`` — number
* ``constants.Z_MEM_ERROR`` — number
* ``constants.Z_MIN_CHUNK`` — number
* ``constants.Z_MIN_LEVEL`` — number
* ``constants.Z_MIN_MEMLEVEL`` — number
* ``constants.Z_MIN_WINDOWBITS`` — number
* ``constants.Z_NEED_DICT`` — number
* ``constants.Z_NO_COMPRESSION`` — number
* ``constants.Z_NO_FLUSH`` — number
* ``constants.Z_NULL`` — number
* ``constants.Z_OK`` — number
* ``constants.Z_PARTIAL_FLUSH`` — number
* ``constants.Z_RLE`` — number
* ``constants.Z_STREAM_END`` — number
* ``constants.Z_STREAM_ERROR`` — number
* ``constants.Z_SYNC_FLUSH`` — number
* ``constants.Z_TEXT`` — number
* ``constants.Z_TREES`` — number
* ``constants.Z_UNKNOWN`` — number
* ``constants.Z_VERSION_ERROR`` — number
* ``constants.ZSTD_c_chainLog`` — number
* ``constants.ZSTD_c_checksumFlag`` — number
* ``constants.ZSTD_c_compressionLevel`` — number
* ``constants.ZSTD_c_contentSizeFlag`` — number
* ``constants.ZSTD_c_dictIDFlag`` — number
* ``constants.ZSTD_c_enableLongDistanceMatching`` — number
* ``constants.ZSTD_c_hashLog`` — number
* ``constants.ZSTD_c_jobSize`` — number
* ``constants.ZSTD_c_ldmBucketSizeLog`` — number
* ``constants.ZSTD_c_ldmHashLog`` — number
* ``constants.ZSTD_c_ldmHashRateLog`` — number
* ``constants.ZSTD_c_ldmMinMatch`` — number
* ``constants.ZSTD_c_minMatch`` — number
* ``constants.ZSTD_c_nbWorkers`` — number
* ``constants.ZSTD_c_overlapLog`` — number
* ``constants.ZSTD_c_searchLog`` — number
* ``constants.ZSTD_c_strategy`` — number
* ``constants.ZSTD_c_targetLength`` — number
* ``constants.ZSTD_c_windowLog`` — number
* ``constants.ZSTD_COMPRESS`` — number
* ``constants.ZSTD_d_windowLogMax`` — number
* ``constants.ZSTD_DECOMPRESS`` — number
* ``constants.ZSTD_e_continue`` — number
* ``constants.ZSTD_e_end`` — number
* ``constants.ZSTD_e_flush`` — number
* ``constants.ZSTD_error_checksum_wrong`` — number
* ``constants.ZSTD_error_corruption_detected`` — number
* ``constants.ZSTD_error_dictionary_corrupted`` — number
* ``constants.ZSTD_error_dictionary_wrong`` — number
* ``constants.ZSTD_error_dictionaryCreation_failed`` — number
* ``constants.ZSTD_error_dstBuffer_null`` — number
* ``constants.ZSTD_error_dstSize_tooSmall`` — number
* ``constants.ZSTD_error_frameParameter_unsupported`` — number
* ``constants.ZSTD_error_frameParameter_windowTooLarge`` — number
* ``constants.ZSTD_error_GENERIC`` — number
* ``constants.ZSTD_error_init_missing`` — number
* ``constants.ZSTD_error_literals_headerWrong`` — number
* ``constants.ZSTD_error_maxSymbolValue_tooLarge`` — number
* ``constants.ZSTD_error_maxSymbolValue_tooSmall`` — number
* ``constants.ZSTD_error_memory_allocation`` — number
* ``constants.ZSTD_error_no_error`` — number
* ``constants.ZSTD_error_noForwardProgress_destFull`` — number
* ``constants.ZSTD_error_noForwardProgress_inputEmpty`` — number
* ``constants.ZSTD_error_parameter_combination_unsupported`` — number
* ``constants.ZSTD_error_parameter_outOfBound`` — number
* ``constants.ZSTD_error_parameter_unsupported`` — number
* ``constants.ZSTD_error_prefix_unknown`` — number
* ``constants.ZSTD_error_srcSize_wrong`` — number
* ``constants.ZSTD_error_stabilityCondition_notRespected`` — number
* ``constants.ZSTD_error_stage_wrong`` — number
* ``constants.ZSTD_error_tableLog_tooLarge`` — number
* ``constants.ZSTD_error_version_unsupported`` — number
* ``constants.ZSTD_error_workSpace_tooSmall`` — number
* ``DEFLATE`` — number
* ``DEFLATERAW`` — number
* ``GUNZIP`` — number
* ``GZIP`` — number
* ``INFLATE`` — number
* ``INFLATERAW`` — number
* ``UNZIP`` — number
* ``Z_ASCII`` — number
* ``Z_BEST_COMPRESSION`` — number
* ``Z_BEST_SPEED`` — number
* ``Z_BINARY`` — number
* ``Z_BLOCK`` — number
* ``Z_BUF_ERROR`` — number
* ``Z_DATA_ERROR`` — number
* ``Z_DEFAULT_CHUNK`` — number
* ``Z_DEFAULT_COMPRESSION`` — number
* ``Z_DEFAULT_LEVEL`` — number
* ``Z_DEFAULT_MEMLEVEL`` — number
* ``Z_DEFAULT_STRATEGY`` — number
* ``Z_DEFAULT_WINDOWBITS`` — number
* ``Z_DEFLATED`` — number
* ``Z_ERRNO`` — number
* ``Z_FILTERED`` — number
* ``Z_FINISH`` — number
* ``Z_FIXED`` — number
* ``Z_FULL_FLUSH`` — number
* ``Z_HUFFMAN_ONLY`` — number
* ``Z_MAX_CHUNK`` — number
* ``Z_MAX_LEVEL`` — number
* ``Z_MAX_MEMLEVEL`` — number
* ``Z_MAX_WINDOWBITS`` — number
* ``Z_MEM_ERROR`` — number
* ``Z_MIN_CHUNK`` — number
* ``Z_MIN_LEVEL`` — number
* ``Z_MIN_MEMLEVEL`` — number
* ``Z_MIN_WINDOWBITS`` — number
* ``Z_NEED_DICT`` — number
* ``Z_NO_COMPRESSION`` — number
* ``Z_NO_FLUSH`` — number
* ``Z_NULL`` — number
* ``Z_OK`` — number
* ``Z_PARTIAL_FLUSH`` — number
* ``Z_RLE`` — number
* ``Z_STREAM_END`` — number
* ``Z_STREAM_ERROR`` — number
* ``Z_SYNC_FLUSH`` — number
* ``Z_TEXT`` — number
* ``Z_TREES`` — number
* ``Z_UNKNOWN`` — number
* ``Z_VERSION_ERROR`` — number
* ``ZSTD_c_chainLog`` — number
* ``ZSTD_c_checksumFlag`` — number
* ``ZSTD_c_compressionLevel`` — number
* ``ZSTD_c_contentSizeFlag`` — number
* ``ZSTD_c_dictIDFlag`` — number
* ``ZSTD_c_enableLongDistanceMatching`` — number
* ``ZSTD_c_hashLog`` — number
* ``ZSTD_c_jobSize`` — number
* ``ZSTD_c_ldmBucketSizeLog`` — number
* ``ZSTD_c_ldmHashLog`` — number
* ``ZSTD_c_ldmHashRateLog`` — number
* ``ZSTD_c_ldmMinMatch`` — number
* ``ZSTD_c_minMatch`` — number
* ``ZSTD_c_nbWorkers`` — number
* ``ZSTD_c_overlapLog`` — number
* ``ZSTD_c_searchLog`` — number
* ``ZSTD_c_strategy`` — number
* ``ZSTD_c_targetLength`` — number
* ``ZSTD_c_windowLog`` — number
* ``ZSTD_COMPRESS`` — number
* ``ZSTD_d_windowLogMax`` — number
* ``ZSTD_DECOMPRESS`` — number
* ``ZSTD_e_continue`` — number
* ``ZSTD_e_end`` — number
* ``ZSTD_e_flush`` — number
* ``ZSTD_error_checksum_wrong`` — number
* ``ZSTD_error_corruption_detected`` — number
* ``ZSTD_error_dictionary_corrupted`` — number
* ``ZSTD_error_dictionary_wrong`` — number
* ``ZSTD_error_dictionaryCreation_failed`` — number
* ``ZSTD_error_dstBuffer_null`` — number
* ``ZSTD_error_dstSize_tooSmall`` — number
* ``ZSTD_error_frameParameter_unsupported`` — number
* ``ZSTD_error_frameParameter_windowTooLarge`` — number
* ``ZSTD_error_GENERIC`` — number
* ``ZSTD_error_init_missing`` — number
* ``ZSTD_error_literals_headerWrong`` — number
* ``ZSTD_error_maxSymbolValue_tooLarge`` — number
* ``ZSTD_error_maxSymbolValue_tooSmall`` — number
* ``ZSTD_error_memory_allocation`` — number
* ``ZSTD_error_no_error`` — number
* ``ZSTD_error_noForwardProgress_destFull`` — number
* ``ZSTD_error_noForwardProgress_inputEmpty`` — number
* ``ZSTD_error_parameter_combination_unsupported`` — number
* ``ZSTD_error_parameter_outOfBound`` — number
* ``ZSTD_error_parameter_unsupported`` — number
* ``ZSTD_error_prefix_unknown`` — number
* ``ZSTD_error_srcSize_wrong`` — number
* ``ZSTD_error_stabilityCondition_notRespected`` — number
* ``ZSTD_error_stage_wrong`` — number
* ``ZSTD_error_tableLog_tooLarge`` — number
* ``ZSTD_error_version_unsupported`` — number
* ``ZSTD_error_workSpace_tooSmall`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
