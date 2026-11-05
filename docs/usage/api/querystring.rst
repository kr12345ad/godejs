querystring
===========

The ``node:querystring`` module provides utilities for parsing and formatting URL query strings. It can be accessed using:

``querystring`` is more performant than ``URLSearchParams`` but is not a standardized API. Use ``URLSearchParams`` when performance is not critical or when compatibility with browser code is desirable.

Loading
-------

.. code-block:: javascript

   const querystring = require('node:querystring');
   // ESM
   import * as querystring from 'node:querystring';

Quick start
-----------

.. code-block:: javascript

   const querystring = require('node:querystring');

API
---

decode
~~~~~~

.. code-block:: text

   querystring.decode()

*Kind:* function · *Arguments:* 0

The ``querystring.decode()`` function is an alias for ``querystring.parse()``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringdecode>`__

encode
~~~~~~

.. code-block:: text

   querystring.encode()

*Kind:* function · *Arguments:* 4

The ``querystring.encode()`` function is an alias for ``querystring.stringify()``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringencode>`__

escape
~~~~~~

.. code-block:: text

   querystring.escape(str)

*Kind:* function · *Arguments:* 1

The ``querystring.escape()`` method performs URL percent-encoding on the given ``str`` in a manner that is optimized for the specific requirements of URL query strings.

The ``querystring.escape()`` method is used by ``querystring.stringify()`` and is generally not expected to be used directly. It is exported primarily to allow application code to provide a replacement percent-encoding implementation if necessary by assigning ``querystring.escape`` to an alternative function.

**Arguments**

* ``str`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringescapestr>`__

parse
~~~~~

.. code-block:: text

   querystring.parse(str[, sep[, eq[, options]]])

*Kind:* function · *Arguments:* 4

The ``querystring.parse()`` method parses a URL query string (``str``) into a collection of key and value pairs.

For example, the query string ``'foo=bar&abc=xyz&abc=123'`` is parsed into:

**Arguments**

* ``str`` (*string*): The URL query string to parse
* ``sep`` (*string*) — optional: The substring used to delimit key and value pairs in the
* ``eq`` (*string*) — optional: . The substring used to delimit keys and values in the
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   // Assuming gbkDecodeURIComponent function already exists...

   querystring.parse('w=%D6%D0%CE%C4&foo=bar', null, null,
                     { decodeURIComponent: gbkDecodeURIComponent });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringparsestr-sep-eq-options>`__

stringify
~~~~~~~~~

.. code-block:: text

   querystring.stringify(obj[, sep[, eq[, options]]])

*Kind:* function · *Arguments:* 4

The ``querystring.stringify()`` method produces a URL query string from a given ``obj`` by iterating through the object's "own properties".

It serializes the following types of values passed in ``obj``: ``string|number|bigint|boolean|string\[]|number\[]|bigint\[]|boolean\[]`` The numeric values must be finite. Any other input values will be coerced to empty strings.

**Arguments**

* ``obj`` (*Object*): The object to serialize into a URL query string
* ``sep`` (*string*) — optional: The substring used to delimit key and value pairs in the
* ``eq`` (*string*) — optional: . The substring used to delimit keys and values in the
* ``options`` — optional

**Example**

.. code-block:: javascript

   querystring.stringify({ foo: 'bar', baz: ['qux', 'quux'], corge: '' });
   // Returns 'foo=bar&baz=qux&baz=quux&corge='

   querystring.stringify({ foo: 'bar', baz: 'qux' }, ';', ':');
   // Returns 'foo:bar;baz:qux'

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringstringifyobj-sep-eq-options>`__

unescape
~~~~~~~~

.. code-block:: text

   querystring.unescape(str)

*Kind:* function · *Arguments:* 1

The ``querystring.unescape()`` method performs decoding of URL percent-encoded characters on the given ``str``.

The ``querystring.unescape()`` method is used by ``querystring.parse()`` and is generally not expected to be used directly. It is exported primarily to allow application code to provide a replacement decoding implementation if necessary by assigning ``querystring.unescape`` to an alternative function.

**Arguments**

* ``str`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/querystring.html#querystringunescapestr>`__

unescapeBuffer
~~~~~~~~~~~~~~

.. code-block:: text

   querystring.unescapeBuffer(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
