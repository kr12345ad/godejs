http3 (GodeJS extension)
========================

A native HTTP/3 server implemented with Go's ``quic-go`` package. HTTP/3 is not part of Node's standard library; this module exposes the same ``IncomingMessage`` and ``ServerResponse`` objects as ``node:http``.

Loading
-------

.. code-block:: javascript

   const http3 = require('gode:http3');
   // ESM
   import http3 from 'gode:http3';

Quick start
-----------

.. code-block:: javascript

   const fs = require('node:fs');
   const http3 = require('gode:http3');

   const server = http3.createServer({
     key: fs.readFileSync('key.pem'),
     cert: fs.readFileSync('cert.pem'),
   }, (req, res) => {
     res.setHeader('content-type', 'application/json');
     res.end(JSON.stringify({ ok: true, url: req.url }));
   });

   server.listen(4433, '0.0.0.0', () => {
     console.log('HTTP/3 on', server.address());
   });

API
---

createSecureServer
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.createSecureServer([options][, requestListener])

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

Alias of ``http3.createServer``. The listener is always TLS because QUIC requires encryption.

**Arguments**

* ``options`` (*Object*) — optional: TLS key/certificate and HTTP/3 options.
* ``requestListener`` (*Function*) — optional: Handler receiving IncomingMessage and ServerResponse.

**Returns**

http3.Server

`GodeJS HTTP/3 guide <../../http3.html>`__

createServer
~~~~~~~~~~~~

.. code-block:: text

   http3.createServer([options][, requestListener])

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

Create an HTTP/3 server. TLS options must include ``key`` and ``cert``; ``requestListener`` is called with ``(req, res)`` exactly like ``http.createServer``.

**Arguments**

* ``options`` (*Object*) — optional: TLS key/certificate and HTTP/3 options.
* ``requestListener`` (*Function*) — optional: Handler receiving IncomingMessage and ServerResponse.

**Returns**

http3.Server

`GodeJS HTTP/3 guide <../../http3.html>`__

IncomingMessage
~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)

*Kind:* function · *Arguments:* 8 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server
~~~~~~

.. code-block:: text

   new http3.Server([options][, requestListener])

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

The HTTP/3 server class. Its request, response, and lifecycle methods mirror ``http.Server``.

**Arguments**

* ``options`` (*Object*) — optional: TLS and HTTP/3 options.
* ``requestListener`` (*Function*) — optional: Request handler.

**Returns**

http3.Server

`GodeJS HTTP/3 guide <../../http3.html>`__

ServerResponse
~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage class
---------------------

IncomingMessage._completeWithoutListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage._completeWithoutListeners(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._emitBody
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage._emitBody(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._hasBodyListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage._hasBodyListeners(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._parseHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage._parseHeaders(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.destroy
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.destroy(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.isPaused
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.isPaused(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.pause
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.pause(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.read
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.read(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.resume
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.resume(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.setEncoding(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.IncomingMessage.setTimeout(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server class
------------

**Server** — The HTTP/3 server class. Its request, response, and lifecycle methods mirror ``http.Server``.

Server.addListener
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.addListener(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.address
~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.address(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.close
~~~~~~~~~~~~

.. code-block:: text

   http3.Server.close(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.closeAllConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.closeAllConnections(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.closeIdleConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.closeIdleConnections(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.getConnections
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.getConnections(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.listen(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.off
~~~~~~~~~~

.. code-block:: text

   http3.Server.off(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.on
~~~~~~~~~

.. code-block:: text

   http3.Server.on(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.once
~~~~~~~~~~~

.. code-block:: text

   http3.Server.once(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.prependListener
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.prependListener(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.prependOnceListener
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.prependOnceListener(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.ref
~~~~~~~~~~

.. code-block:: text

   http3.Server.ref(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.removeAllListeners(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.removeListener
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.Server.removeListener(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.unref
~~~~~~~~~~~~

.. code-block:: text

   http3.Server.unref(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse class
--------------------

ServerResponse._flushHead
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse._flushHead(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse._headersObject(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._packedHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse._packedHeaders(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._sendHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse._sendHead(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.addTrailers
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.addTrailers(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.appendHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.appendHeader(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.destroy
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.destroy(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.end
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.flushHeaders(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.getHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.getHeader(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.getHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.getHeaderNames(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.getHeaders(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.getRawHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.getRawHeaderNames(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.hasHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.hasHeader(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.removeHeader(arg0)

*Kind:* function · *Arguments:* 1 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.setHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.setHeader(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.setTimeout(arg0, arg1)

*Kind:* function · *Arguments:* 2 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.write
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.writeContinue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.writeContinue(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.writeHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.writeHead(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.writeProcessing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http3.ServerResponse.writeProcessing(...args)

*Kind:* function · *Arguments:* 0 · *GodeJS extension*

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Properties and constants
------------------------

* ``IncomingMessage.connection`` — undefined
* ``IncomingMessage.headers`` — undefined
* ``IncomingMessage.headersDistinct`` — undefined
* ``IncomingMessage.rawHeaders`` — undefined
* ``IncomingMessage.socket`` — undefined
* ``IncomingMessage.trailers`` — undefined
* ``IncomingMessage.trailersDistinct`` — undefined
* ``ServerResponse.connection`` — undefined
* ``ServerResponse.socket`` — undefined
* ``STATUS_CODES`` — object
* ``STATUS_CODES.100`` — string
* ``STATUS_CODES.101`` — string
* ``STATUS_CODES.102`` — string
* ``STATUS_CODES.103`` — string
* ``STATUS_CODES.200`` — string
* ``STATUS_CODES.201`` — string
* ``STATUS_CODES.202`` — string
* ``STATUS_CODES.203`` — string
* ``STATUS_CODES.204`` — string
* ``STATUS_CODES.205`` — string
* ``STATUS_CODES.206`` — string
* ``STATUS_CODES.300`` — string
* ``STATUS_CODES.301`` — string
* ``STATUS_CODES.302`` — string
* ``STATUS_CODES.303`` — string
* ``STATUS_CODES.304`` — string
* ``STATUS_CODES.305`` — string
* ``STATUS_CODES.307`` — string
* ``STATUS_CODES.308`` — string
* ``STATUS_CODES.400`` — string
* ``STATUS_CODES.401`` — string
* ``STATUS_CODES.402`` — string
* ``STATUS_CODES.403`` — string
* ``STATUS_CODES.404`` — string
* ``STATUS_CODES.405`` — string
* ``STATUS_CODES.406`` — string
* ``STATUS_CODES.407`` — string
* ``STATUS_CODES.408`` — string
* ``STATUS_CODES.409`` — string
* ``STATUS_CODES.410`` — string
* ``STATUS_CODES.411`` — string
* ``STATUS_CODES.412`` — string
* ``STATUS_CODES.413`` — string
* ``STATUS_CODES.414`` — string
* ``STATUS_CODES.415`` — string
* ``STATUS_CODES.416`` — string
* ``STATUS_CODES.417`` — string
* ``STATUS_CODES.418`` — string
* ``STATUS_CODES.421`` — string
* ``STATUS_CODES.422`` — string
* ``STATUS_CODES.423`` — string
* ``STATUS_CODES.424`` — string
* ``STATUS_CODES.425`` — string
* ``STATUS_CODES.426`` — string
* ``STATUS_CODES.428`` — string
* ``STATUS_CODES.429`` — string
* ``STATUS_CODES.431`` — string
* ``STATUS_CODES.451`` — string
* ``STATUS_CODES.500`` — string
* ``STATUS_CODES.501`` — string
* ``STATUS_CODES.502`` — string
* ``STATUS_CODES.503`` — string
* ``STATUS_CODES.504`` — string
* ``STATUS_CODES.505`` — string
* ``STATUS_CODES.506`` — string
* ``STATUS_CODES.507`` — string
* ``STATUS_CODES.508`` — string
* ``STATUS_CODES.510`` — string
* ``STATUS_CODES.511`` — string

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
