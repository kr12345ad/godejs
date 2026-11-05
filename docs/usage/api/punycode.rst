punycode
========

**The version of the punycode module bundled in Node.js is being deprecated.** In a future major version of Node.js this module will be removed. Users currently depending on the ``punycode`` module should switch to using the userland-provided Punycode.js module instead. For punycode-based URL encoding, see ``url.domainToASCII`` or, more generally, the WHATWG URL API.

The ``punycode`` module is a bundled version of the Punycode.js module. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const punycode = require('node:punycode');
   // ESM
   import * as punycode from 'node:punycode';

Quick start
-----------

.. code-block:: javascript

   const punycode = require('node:punycode');

API
---

decode
~~~~~~

.. code-block:: text

   punycode.decode(string)

*Kind:* function · *Arguments:* 1

The ``punycode.decode()`` method converts a Punycode string of ASCII-only characters to the equivalent string of Unicode codepoints.

**Arguments**

* ``string`` (*string*)

**Example**

.. code-block:: javascript

   punycode.decode('maana-pta'); // 'mañana'
   punycode.decode('--dqo34k'); // '☃-⌘'


encode
~~~~~~

.. code-block:: text

   punycode.encode(string)

*Kind:* function · *Arguments:* 1

The ``punycode.encode()`` method converts a string of Unicode codepoints to a Punycode string of ASCII-only characters.

**Arguments**

* ``string`` (*string*)

**Example**

.. code-block:: javascript

   punycode.encode('mañana'); // 'maana-pta'
   punycode.encode('☃-⌘'); // '--dqo34k'


toASCII
~~~~~~~

.. code-block:: text

   punycode.toASCII(domain)

*Kind:* function · *Arguments:* 1

The ``punycode.toASCII()`` method converts a Unicode string representing an Internationalized Domain Name to Punycode. Only the non-ASCII parts of the domain name will be converted. Calling ``punycode.toASCII()`` on a string that already only contains ASCII characters will have no effect.

**Arguments**

* ``domain`` (*string*)

**Example**

.. code-block:: javascript

   // encode domain names
   punycode.toASCII('mañana.com');  // 'xn--maana-pta.com'
   punycode.toASCII('☃-⌘.com');   // 'xn----dqo34k.com'
   punycode.toASCII('example.com'); // 'example.com'


toUnicode
~~~~~~~~~

.. code-block:: text

   punycode.toUnicode(domain)

*Kind:* function · *Arguments:* 1

The ``punycode.toUnicode()`` method converts a string representing a domain name containing Punycode encoded characters into Unicode. Only the Punycode encoded parts of the domain name are converted.

**Arguments**

* ``domain`` (*string*)

**Example**

.. code-block:: javascript

   // decode domain names
   punycode.toUnicode('xn--maana-pta.com'); // 'mañana.com'
   punycode.toUnicode('xn----dqo34k.com');  // '☃-⌘.com'
   punycode.toUnicode('example.com');       // 'example.com'


ucs2
----

ucs2.decode
~~~~~~~~~~~

.. code-block:: text

   punycode.ucs2.decode(string)

*Kind:* function · *Arguments:* 1

The ``punycode.ucs2.decode()`` method returns an array containing the numeric codepoint values of each Unicode symbol in the string.

**Arguments**

* ``string`` (*string*)

**Example**

.. code-block:: javascript

   punycode.ucs2.decode('abc'); // [0x61, 0x62, 0x63]
   // surrogate pair for U+1D306 tetragram for centre:
   punycode.ucs2.decode('\uD834\uDF06'); // [0x1D306]


ucs2.encode
~~~~~~~~~~~

.. code-block:: text

   punycode.ucs2.encode(codePoints)

*Kind:* function · *Arguments:* 1

The ``punycode.ucs2.encode()`` method returns a string based on an array of numeric code point values.

**Arguments**

* ``codePoints`` (*integer\[]*)

**Example**

.. code-block:: javascript

   punycode.ucs2.encode([0x61, 0x62, 0x63]); // 'abc'
   punycode.ucs2.encode([0x1D306]); // '\uD834\uDF06'


Properties and constants
------------------------

* ``ucs2`` — object
* ``version`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
