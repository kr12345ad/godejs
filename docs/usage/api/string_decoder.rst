string_decoder
==============

The ``node:string_decoder`` module provides an API for decoding ``Buffer`` objects into strings in a manner that preserves encoded multi-byte UTF-8 and UTF-16 characters. It can be accessed using:

The following example shows the basic use of the ``StringDecoder`` class.

Loading
-------

.. code-block:: javascript

   const string_decoder = require('node:string_decoder');
   // ESM
   import * as string_decoder from 'node:string_decoder';

Quick start
-----------

.. code-block:: javascript

   const { StringDecoder } = require('node:string_decoder');

API
---

StringDecoder
~~~~~~~~~~~~~

.. code-block:: text

   StringDecoder

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const string_decoder = require('node:string_decoder');

   const instance = new string_decoder.StringDecoder(undefined);
   console.log(instance);


StringDecoder class
-------------------

StringDecoder.end
~~~~~~~~~~~~~~~~~

.. code-block:: text

   stringDecoder.end([buffer])

*Kind:* function · *Arguments:* 1

Returns any remaining input stored in the internal buffer as a string. Bytes representing incomplete UTF-8 and UTF-16 characters will be replaced with substitution characters appropriate for the character encoding.

If the ``buffer`` argument is provided, one final call to ``stringDecoder.write()`` is performed before returning the remaining input. After ``end()`` is called, the ``stringDecoder`` object can be reused for new input.

**Arguments**

* ``buffer`` (*string|Buffer|TypedArray|DataView*) — optional: The bytes to decode.

**Example**

.. code-block:: javascript

   const string_decoder = require('node:string_decoder');

   const instance = new string_decoder.StringDecoder();
   const result = instance.end([]);
   console.log(result);


StringDecoder.text
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   string_decoder.StringDecoder.text(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const string_decoder = require('node:string_decoder');

   const instance = new string_decoder.StringDecoder();
   const result = instance.text(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


StringDecoder.write
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   stringDecoder.write(buffer)

*Kind:* function · *Arguments:* 1

Returns a decoded string, ensuring that any incomplete multibyte characters at the end of the ``Buffer``, or ``TypedArray``, or ``DataView`` are omitted from the returned string and stored in an internal buffer for the next call to ``stringDecoder.write()`` or ``stringDecoder.end()``.

**Arguments**

* ``buffer`` (*string|Buffer|TypedArray|DataView*): The bytes to decode.

**Example**

.. code-block:: javascript

   const string_decoder = require('node:string_decoder');

   const instance = new string_decoder.StringDecoder();
   const result = instance.write([]);
   console.log(result);


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
