sea
===

This feature allows the distribution of a Node.js application conveniently to a system that does not have Node.js installed.

Node.js supports the creation of single executable applications by allowing the injection of a blob prepared by Node.js, which can contain a bundled script, into the ``node`` binary. During start up, the program checks if anything has been injected. If the blob is found, it executes the script in the blob. Otherwise Node.js operates as it normally does.

Loading
-------

.. code-block:: javascript

   const sea = require('node:sea');
   // ESM
   import * as sea from 'node:sea';

API
---

getAsset
~~~~~~~~

.. code-block:: text

   sea.getAsset(key[, encoding])

*Kind:* function · *Arguments:* 2

This method can be used to retrieve the assets configured to be bundled into the single-executable application at build time. An error is thrown when no matching asset can be found.

* ``key``  ``string`` the key for the asset in the dictionary specified by the ``assets`` field in the single-executable application configuration. * Returns: ``string|ArrayBuffer``

**Arguments**

* ``key``
* ``encoding`` (*string*) — optional: If specified, the asset will be decoded as


getAssetAsBlob
~~~~~~~~~~~~~~

.. code-block:: text

   sea.getAssetAsBlob(key[, options])

*Kind:* function · *Arguments:* 2

Similar to ``sea.getAsset()``, but returns the result in a ``Blob``. An error is thrown when no matching asset can be found.

* ``key``  ``string`` the key for the asset in the dictionary specified by the ``assets`` field in the single-executable application configuration. * Returns: ``Blob``

**Arguments**

* ``key``
* ``options`` (*Object*) — optional


getRawAsset
~~~~~~~~~~~

.. code-block:: text

   sea.getRawAsset(key)

*Kind:* function · *Arguments:* 1

This method can be used to retrieve the assets configured to be bundled into the single-executable application at build time. An error is thrown when no matching asset can be found.

Unlike ``sea.getAsset()`` or ``sea.getAssetAsBlob()``, this method does not return a copy. Instead, it returns the raw asset bundled inside the executable.

**Arguments**

* ``key``


isSea
~~~~~

.. code-block:: text

   sea.isSea()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
