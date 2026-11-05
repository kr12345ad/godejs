http2
=====

The ``node:http2`` module provides an implementation of the HTTP/2 protocol. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const http2 = require('node:http2');
   // ESM
   import * as http2 from 'node:http2';

Quick start
-----------

.. code-block:: javascript

   const http2 = require('node:http2');

API
---

ClientHttp2Session
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   ClientHttp2Session

*Kind:* class · *Arguments:* 2

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const client = http2.connect('https://example.org');

   client.on('altsvc', (alt, origin, streamId) => {
     console.log(alt);
     console.log(origin);
     console.log(streamId);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-clienthttp2session>`__

ClientHttp2Stream
~~~~~~~~~~~~~~~~~

.. code-block:: text

   ClientHttp2Stream

*Kind:* class · *Arguments:* 2

The ``ClientHttp2Stream`` class is an extension of ``Http2Stream`` that is used exclusively on HTTP/2 Clients. ``Http2Stream`` instances on the client provide events such as ``'response'`` and ``'push'`` that are only relevant on the client.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const client = http2.connect('https://localhost');
   const req = client.request({ ':path': '/' });
   req.on('response', (headers, flags) => {
     console.log(headers[':status']);
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-clienthttp2stream>`__

connect
~~~~~~~

.. code-block:: text

   http2.connect(authority[, options][, listener])

*Kind:* function · *Arguments:* 3

Returns a ``ClientHttp2Session`` instance.

**Arguments**

* ``authority`` (*string|URL*): The remote HTTP/2 server to connect to. This must
* ``options`` (*Object*) — optional
* ``listener`` (*Function*) — optional: Will be registered as a one-time listener of the

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const client = http2.connect('https://localhost:1234');

   /* Use the client */

   client.close();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2connectauthority-options-listener>`__

createSecureServer
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.createSecureServer(options[, onRequestHandler])

*Kind:* function · *Arguments:* 2

Returns a ``tls.Server`` instance that creates and manages ``Http2Session`` instances.

**Arguments**

* ``options`` (*Object*)
* ``onRequestHandler`` (*Function*) — optional: See Compatibility API

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const fs = require('node:fs');

   const options = {
     key: fs.readFileSync('server-key.pem'),
     cert: fs.readFileSync('server-cert.pem'),
   };

   // Create a secure HTTP/2 server
   const server = http2.createSecureServer(options);

   server.on('stream', (stream, headers) => {
     stream.respond({
       'content-type': 'text/html; charset=utf-8',
       ':status': 200,
     });
     stream.end('<h1>Hello World</h1>');
   });

   server.listen(8443);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2createsecureserveroptions-onrequesthandler>`__

createServer
~~~~~~~~~~~~

.. code-block:: text

   http2.createServer([options][, onRequestHandler])

*Kind:* function · *Arguments:* 2

Returns a ``net.Server`` instance that creates and manages ``Http2Session`` instances.

Since there are no browsers known that support unencrypted HTTP/2, the use of ``http2.createSecureServer()`` is necessary when communicating with browser clients.

**Arguments**

* ``options`` (*Object*) — optional
* ``onRequestHandler`` (*Function*) — optional: See Compatibility API

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');

   // Create an unencrypted HTTP/2 server.
   // Since there are no browsers known that support
   // unencrypted HTTP/2, the use of `http2.createSecureServer()`
   // is necessary when communicating with browser clients.
   const server = http2.createServer();

   server.on('stream', (stream, headers) => {
     stream.respond({
       'content-type': 'text/html; charset=utf-8',
       ':status': 200,
     });
     stream.end('<h1>Hello World</h1>');
   });

   server.listen(8000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2createserveroptions-onrequesthandler>`__

Http2Server
~~~~~~~~~~~

.. code-block:: text

   Http2Server

*Kind:* class · *Arguments:* 3

Instances of ``Http2Server`` are created using the ``http2.createServer()`` function. The ``Http2Server`` class is not exported directly by the ``node:http2`` module.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const {
     HTTP2_HEADER_METHOD,
     HTTP2_HEADER_PATH,
     HTTP2_HEADER_STATUS,
     HTTP2_HEADER_CONTENT_TYPE,
   } = http2.constants;

   const server = http2.createServer();
   server.on('stream', (stream, headers, flags) => {
     const method = headers[HTTP2_HEADER_METHOD];
     const path = headers[HTTP2_HEADER_PATH];
     // ...
     stream.respond({
       [HTTP2_HEADER_STATUS]: 200,
       [HTTP2_HEADER_CONTENT_TYPE]: 'text/plain; charset=utf-8',
     });
     stream.write('hello ');
     stream.end('world');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-http2server>`__

Http2ServerRequest
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest

*Kind:* class · *Arguments:* 12

A ``Http2ServerRequest`` object is created by ``http2.Server`` or ``http2.SecureServer`` and passed as the first argument to the ``'request'`` event. It may be used to access a request status, headers, and data.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   // Prints something like:
   //
   // { 'user-agent': 'curl/7.22.0',
   //   host: '127.0.0.1:8000',
   //   accept: '*/*' }
   console.log(request.headers);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-http2http2serverrequest>`__

Http2ServerResponse
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse

*Kind:* class · *Arguments:* 3

This object is created internally by an HTTP server, not by the user. It is passed as the second parameter to the ``'request'`` event.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const server = http2.createServer((req, res) => {
     const ip = req.socket.remoteAddress;
     const port = req.socket.remotePort;
     res.end(`Your IP address is ${ip} and your source port is ${port}.`);
   }).listen(3000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-http2http2serverresponse>`__

Http2Stream
~~~~~~~~~~~

.. code-block:: text

   Http2Stream

*Kind:* class · *Arguments:* 4

Each instance of the ``Http2Stream`` class represents a bidirectional HTTP/2 communications stream over an ``Http2Session`` instance. Any single ``Http2Session`` may have up to 2<sup>31</sup>-1 ``Http2Stream`` instances over its lifetime.

User code will not construct ``Http2Stream`` instances directly. Rather, these are created, managed, and provided to user code through the ``Http2Session`` instance. On the server, ``Http2Stream`` instances are created either in response to an incoming HTTP request (and handed off to user code via the ``'stream'`` event), or in response to a call to the ``http2stream.pushStream()`` method. On the client, ``Http2Stream`` instances are created and returned when either the ``http2session.request()`` method is called, or in response to an incoming ``'push'`` event.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const client = http2.connect('http://example.org:8000');
   const { NGHTTP2_CANCEL } = http2.constants;
   const req = client.request({ ':path': '/' });

   // Cancel the stream if there's no activity after 5 seconds
   req.setTimeout(5000, () => req.close(NGHTTP2_CANCEL));

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#class-http2stream>`__

ClientHttp2Session class
------------------------

ClientHttp2Session.close
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.close([callback])

*Kind:* function · *Arguments:* 1

Gracefully closes the ``Http2Session``, allowing any existing streams to complete on their own and preventing new ``Http2Stream`` instances from being created. Once closed, ``http2session.destroy()`` *might* be called if there are no open ``Http2Stream`` instances.

If specified, the ``callback`` function is registered as a handler for the ``'close'`` event.

**Arguments**

* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionclosecallback>`__

ClientHttp2Session.destroy
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.destroy([error][, code])

*Kind:* function · *Arguments:* 2

Immediately terminates the ``Http2Session`` and the associated ``net.Socket`` or ``tls.TLSSocket``.

Once destroyed, the ``Http2Session`` will emit the ``'close'`` event. If ``error`` is not undefined, an ``'error'`` event will be emitted immediately before the ``'close'`` event.

**Arguments**

* ``error`` (*Error*) — optional: An ``Error`` object if the ``Http2Session`` is being destroyed
* ``code`` (*number*) — optional: The HTTP/2 error code to send in the final ``GOAWAY`` frame.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessiondestroyerror-code>`__

ClientHttp2Session.goaway
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.goaway([code[, lastStreamID[, opaqueData]]])

*Kind:* function · *Arguments:* 3

Transmits a ``GOAWAY`` frame to the connected peer *without* shutting down the ``Http2Session``.

**Arguments**

* ``code`` (*number*) — optional: An HTTP/2 error code
* ``lastStreamID`` (*number*) — optional: The numeric ID of the last processed ``Http2Stream``
* ``opaqueData`` (*Buffer|TypedArray|DataView*) — optional: A ``TypedArray`` or ``DataView``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessiongoawaycode-laststreamid-opaquedata>`__

ClientHttp2Session.ping
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.ping([payload, ]callback)

*Kind:* function · *Arguments:* 2

Sends a ``PING`` frame to the connected HTTP/2 peer. A ``callback`` function must be provided. The method will return ``true`` if the ``PING`` was sent, ``false`` otherwise.

The maximum number of outstanding (unacknowledged) pings is determined by the ``maxOutstandingPings`` configuration option. The default maximum is 10.

**Arguments**

* ``payload`` (*Buffer|TypedArray|DataView*) — optional: Optional ping payload.
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   session.ping(Buffer.from('abcdefgh'), (err, duration, payload) => {
     if (!err) {
       console.log(`Ping acknowledged in ${duration} milliseconds`);
       console.log(`With payload '${payload.toString()}'`);
     }
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionpingpayload-callback>`__

ClientHttp2Session.ref
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.ref()

*Kind:* function · *Arguments:* 0

Calls ``ref()`` on this ``Http2Session`` instance's underlying ``net.Socket``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionref>`__

ClientHttp2Session.request
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   clienthttp2session.request(headers[, options])

*Kind:* function · *Arguments:* 2

For HTTP/2 Client ``Http2Session`` instances only, the ``http2session.request()`` creates and returns an ``Http2Stream`` instance that can be used to send an HTTP/2 request to the connected server.

When a ``ClientHttp2Session`` is first created, the socket may not yet be connected. If ``clienthttp2session.request()`` is called during this time, the actual request will be deferred until the socket is ready to go.

**Arguments**

* ``headers`` (*HTTP/2 Headers Object|HTTP/2 Raw Headers*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const clientSession = http2.connect('https://localhost:1234');
   const {
     HTTP2_HEADER_PATH,
     HTTP2_HEADER_STATUS,
   } = http2.constants;

   const req = clientSession.request({ [HTTP2_HEADER_PATH]: '/' });
   req.on('response', (headers) => {
     console.log(headers[HTTP2_HEADER_STATUS]);
     req.on('data', (chunk) => { /* .. */ });
     req.on('end', () => { /* .. */ });
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#clienthttp2sessionrequestheaders-options>`__

ClientHttp2Session.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

Used to set a callback function that is called when there is no activity on the ``Http2Session`` after ``msecs`` milliseconds. The given ``callback`` is registered as a listener on the ``'timeout'`` event.

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettimeoutmsecs-callback>`__

ClientHttp2Session.settings
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.settings([settings][, callback])

*Kind:* function · *Arguments:* 2

Updates the current local settings for this ``Http2Session`` and sends a new ``SETTINGS`` frame to the connected HTTP/2 peer.

Once called, the ``http2session.pendingSettingsAck`` property will be ``true`` while the session is waiting for the remote peer to acknowledge the new settings.

**Arguments**

* ``settings`` (*HTTP/2 Settings Object*) — optional
* ``callback`` (*Function*) — optional: Callback that is called once the session is connected or

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettingssettings-callback>`__

ClientHttp2Session.unref
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.unref()

*Kind:* function · *Arguments:* 0

Calls ``unref()`` on this ``Http2Session`` instance's underlying ``net.Socket``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionunref>`__

ClientHttp2Stream class
-----------------------

**ClientHttp2Stream** — The ``ClientHttp2Stream`` class is an extension of ``Http2Stream`` that is used exclusively on HTTP/2 Clients. ``Http2Stream`` instances on the client provide events such as ``'response'`` and ``'push'`` that are only relevant on the client.

ClientHttp2Stream.close
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.close(code[, callback])

*Kind:* function · *Arguments:* 2

Closes the ``Http2Stream`` instance by sending an ``RST_STREAM`` frame to the connected HTTP/2 peer.

**Arguments**

* ``code`` (*number*): Unsigned 32-bit integer identifying the error code.
* ``callback`` (*Function*) — optional: An optional function registered to listen for the

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamclosecode-callback>`__

ClientHttp2Stream.destroy
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Calls ``destroy()`` on the ``Http2Stream`` that received the ``Http2ServerRequest``. If ``error`` is provided, an ``'error'`` event is emitted and ``error`` is passed as an argument to any listeners on the event.

It does nothing if the stream was already destroyed.

**Arguments**

* ``error`` (*Error*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#requestdestroyerror>`__

ClientHttp2Stream.end
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

This method signals to the server that all of the response headers and body have been sent; that server should consider this message complete. The method, ``response.end()``, MUST be called on each response.

If ``data`` is specified, it is equivalent to calling ``response.write(data, encoding)`` followed by ``response.end(callback)``.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseenddata-encoding-callback>`__

ClientHttp2Stream.isPaused
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.ClientHttp2Stream.isPaused(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ClientHttp2Stream.pause
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.ClientHttp2Stream.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ClientHttp2Stream.priority
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.priority(options)

*Kind:* function · *Arguments:* 1

Empty method, only there to maintain some backward compatibility.

**Arguments**

* ``options``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streampriorityoptions>`__

ClientHttp2Stream.resume
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.ClientHttp2Stream.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ClientHttp2Stream.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.ClientHttp2Stream.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ClientHttp2Stream.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

Used to set a callback function that is called when there is no activity on the ``Http2Session`` after ``msecs`` milliseconds. The given ``callback`` is registered as a listener on the ``'timeout'`` event.

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettimeoutmsecs-callback>`__

ClientHttp2Stream.write
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.write(chunk[, encoding][, callback])

*Kind:* function · *Arguments:* 3

If this method is called and ``response.writeHead()`` has not been called, it will switch to implicit header mode and flush the implicit headers.

This sends a chunk of the response body. This method may be called multiple times to provide successive parts of the body.

**Arguments**

* ``chunk`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsewritechunk-encoding-callback>`__

Http2Server class
-----------------

**Http2Server** — Instances of ``Http2Server`` are created using the ``http2.createServer()`` function. The ``Http2Server`` class is not exported directly by the ``node:http2`` module.

Http2Server.address
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Server.address(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Server.close
~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.close([callback])

*Kind:* function · *Arguments:* 1

Gracefully closes the ``Http2Session``, allowing any existing streams to complete on their own and preventing new ``Http2Stream`` instances from being created. Once closed, ``http2session.destroy()`` *might* be called if there are no open ``Http2Stream`` instances.

If specified, the ``callback`` function is registered as a handler for the ``'close'`` event.

**Arguments**

* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionclosecallback>`__

Http2Server.getConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Server.getConnections(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Server.listen
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Server.listen(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Server.ref
~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.ref()

*Kind:* function · *Arguments:* 0

Calls ``ref()`` on this ``Http2Session`` instance's underlying ``net.Socket``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionref>`__

Http2Server.setTimeout
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

Used to set a callback function that is called when there is no activity on the ``Http2Session`` after ``msecs`` milliseconds. The given ``callback`` is registered as a listener on the ``'timeout'`` event.

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettimeoutmsecs-callback>`__

Http2Server.unref
~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.unref()

*Kind:* function · *Arguments:* 0

Calls ``unref()`` on this ``Http2Session`` instance's underlying ``net.Socket``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionunref>`__

Http2ServerRequest class
------------------------

**Http2ServerRequest** — A ``Http2ServerRequest`` object is created by ``http2.Server`` or ``http2.SecureServer`` and passed as the first argument to the ``'request'`` event. It may be used to access a request status, headers, and data.

Http2ServerRequest._emitBody
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest._emitBody(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.destroy
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Calls ``destroy()`` on the ``Http2Stream`` that received the ``Http2ServerRequest``. If ``error`` is provided, an ``'error'`` event is emitted and ``error`` is passed as an argument to any listeners on the event.

It does nothing if the stream was already destroyed.

**Arguments**

* ``error`` (*Error*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#requestdestroyerror>`__

Http2ServerRequest.isPaused
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest.isPaused(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.pause
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.read
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest.read(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.resume
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerRequest.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerRequest.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

Used to set a callback function that is called when there is no activity on the ``Http2Session`` after ``msecs`` milliseconds. The given ``callback`` is registered as a listener on the ``'timeout'`` event.

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettimeoutmsecs-callback>`__

Http2ServerResponse class
-------------------------

**Http2ServerResponse** — This object is created internally by an HTTP server, not by the user. It is passed as the second parameter to the ``'request'`` event.

Http2ServerResponse._emitTrailers
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse._emitTrailers(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerResponse._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse._headersObject(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerResponse._sendHead
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse._sendHead(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerResponse.addTrailers
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.addTrailers(headers)

*Kind:* function · *Arguments:* 1

This method adds HTTP trailing headers (a header but at the end of the message) to the response.

Attempting to set a header field name or value that contains invalid characters will result in a ``TypeError`` being thrown.

**Arguments**

* ``headers`` (*Object*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseaddtrailersheaders>`__

Http2ServerResponse.appendHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.appendHeader(name, value)

*Kind:* function · *Arguments:* 2

Append a single header value to the header object.

If the value is an array, this is equivalent to calling this method multiple times.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string|string\[]*)

**Example**

.. code-block:: javascript

   // Returns headers including "set-cookie: a" and "set-cookie: b"
   const server = http2.createServer((req, res) => {
     res.setHeader('set-cookie', 'a');
     res.appendHeader('set-cookie', 'b');
     res.writeHead(200);
     res.end('ok');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseappendheadername-value>`__

Http2ServerResponse.destroy
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Calls ``destroy()`` on the ``Http2Stream`` that received the ``Http2ServerRequest``. If ``error`` is provided, an ``'error'`` event is emitted and ``error`` is passed as an argument to any listeners on the event.

It does nothing if the stream was already destroyed.

**Arguments**

* ``error`` (*Error*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#requestdestroyerror>`__

Http2ServerResponse.end
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

This method signals to the server that all of the response headers and body have been sent; that server should consider this message complete. The method, ``response.end()``, MUST be called on each response.

If ``data`` is specified, it is equivalent to calling ``response.write(data, encoding)`` followed by ``response.end(callback)``.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseenddata-encoding-callback>`__

Http2ServerResponse.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse.flushHeaders(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2ServerResponse.getHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.getHeader(name)

*Kind:* function · *Arguments:* 1

Reads out a header that has already been queued but not sent to the client. The name is case-insensitive.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const contentType = response.getHeader('content-type');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsegetheadername>`__

Http2ServerResponse.getHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.getHeaderNames()

*Kind:* function · *Arguments:* 0

Returns an array containing the unique names of the current outgoing headers. All header names are lowercase.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   response.setHeader('Foo', 'bar');
   response.setHeader('Set-Cookie', ['foo=bar', 'bar=baz']);

   const headerNames = response.getHeaderNames();
   // headerNames === ['foo', 'set-cookie']

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsegetheadernames>`__

Http2ServerResponse.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.getHeaders()

*Kind:* function · *Arguments:* 0

Returns a shallow copy of the current outgoing headers. Since a shallow copy is used, array values may be mutated without additional calls to various header-related http module methods. The keys of the returned object are the header names and the values are the respective header values. All header names are lowercase.

The object returned by the ``response.getHeaders()`` method *does not* prototypically inherit from the JavaScript ``Object``. This means that typical ``Object`` methods such as ``obj.toString()``, ``obj.hasOwnProperty()``, and others are not defined and *will not work*.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   response.setHeader('Foo', 'bar');
   response.setHeader('Set-Cookie', ['foo=bar', 'bar=baz']);

   const headers = response.getHeaders();
   // headers === { foo: 'bar', 'set-cookie': ['foo=bar', 'bar=baz'] }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsegetheaders>`__

Http2ServerResponse.hasHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.hasHeader(name)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the header identified by ``name`` is currently set in the outgoing headers. The header name matching is case-insensitive.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const hasContentType = response.hasHeader('content-type');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsehasheadername>`__

Http2ServerResponse.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.removeHeader(name)

*Kind:* function · *Arguments:* 1

Removes a header that has been queued for implicit sending.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   response.removeHeader('Content-Encoding');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseremoveheadername>`__

Http2ServerResponse.setHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.setHeader(name, value)

*Kind:* function · *Arguments:* 2

Sets a single header value for implicit headers. If this header already exists in the to-be-sent headers, its value will be replaced. Use an array of strings here to send multiple headers with the same name.

or

**Arguments**

* ``name`` (*string*)
* ``value`` (*string|string\[]*)

**Example**

.. code-block:: javascript

   response.setHeader('Content-Type', 'text/html; charset=utf-8');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsesetheadername-value>`__

Http2ServerResponse.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2session.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

Used to set a callback function that is called when there is no activity on the ``Http2Session`` after ``msecs`` milliseconds. The given ``callback`` is registered as a listener on the ``'timeout'`` event.

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2sessionsettimeoutmsecs-callback>`__

Http2ServerResponse.write
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.write(chunk[, encoding][, callback])

*Kind:* function · *Arguments:* 3

If this method is called and ``response.writeHead()`` has not been called, it will switch to implicit header mode and flush the implicit headers.

This sends a chunk of the response body. This method may be called multiple times to provide successive parts of the body.

**Arguments**

* ``chunk`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsewritechunk-encoding-callback>`__

Http2ServerResponse.writeContinue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.writeContinue()

*Kind:* function · *Arguments:* 0

Sends a status ``100 Continue`` to the client, indicating that the request body should be sent. See the ``'checkContinue'`` event on ``Http2Server`` and ``Http2SecureServer``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsewritecontinue>`__

Http2ServerResponse.writeHead
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.writeHead(statusCode[, statusMessage][, headers])

*Kind:* function · *Arguments:* 3

Sends a response header to the request. The status code is a 3-digit HTTP status code, like ``404``. The last argument, ``headers``, are the response headers.

Returns a reference to the ``Http2ServerResponse``, so that calls can be chained.

**Arguments**

* ``statusCode`` (*number*)
* ``statusMessage`` (*string*) — optional
* ``headers`` (*HTTP/2 Headers Object|HTTP/2 Raw Headers*) — optional

**Example**

.. code-block:: javascript

   const body = 'hello world';
   response.writeHead(200, {
     'Content-Length': Buffer.byteLength(body),
     'Content-Type': 'text/plain; charset=utf-8',
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsewriteheadstatuscode-statusmessage-headers>`__

Http2ServerResponse.writeProcessing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2ServerResponse.writeProcessing(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream class
-----------------

**Http2Stream** — Each instance of the ``Http2Stream`` class represents a bidirectional HTTP/2 communications stream over an ``Http2Session`` instance. Any single ``Http2Session`` may have up to 2<sup>31</sup>-1 ``Http2Stream`` instances over its lifetime.

User code will not construct ``Http2Stream`` instances directly. Rather, these are created, managed, and provided to user code through the ``Http2Session`` instance. On the server, ``Http2Stream`` instances are created either in response to an incoming HTTP request (and handed off to user code via the ``'stream'`` event), or in response to a call to the ``http2stream.pushStream()`` method. On the client, ``Http2Stream`` instances are created and returned when either the ``http2session.request()`` method is called, or in response to an incoming ``'push'`` event.

Http2Stream.additionalHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.additionalHeaders(headers)

*Kind:* function · *Arguments:* 1

Sends an additional informational ``HEADERS`` frame to the connected HTTP/2 peer.

**Arguments**

* ``headers`` (*HTTP/2 Headers Object*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamadditionalheadersheaders>`__

Http2Stream.addTrailers
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.addTrailers(headers)

*Kind:* function · *Arguments:* 1

This method adds HTTP trailing headers (a header but at the end of the message) to the response.

Attempting to set a header field name or value that contains invalid characters will result in a ``TypeError`` being thrown.

**Arguments**

* ``headers`` (*Object*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseaddtrailersheaders>`__

Http2Stream.close
~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.close(code[, callback])

*Kind:* function · *Arguments:* 2

Closes the ``Http2Stream`` instance by sending an ``RST_STREAM`` frame to the connected HTTP/2 peer.

**Arguments**

* ``code`` (*number*): Unsigned 32-bit integer identifying the error code.
* ``callback`` (*Function*) — optional: An optional function registered to listen for the

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamclosecode-callback>`__

Http2Stream.destroy
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Calls ``destroy()`` on the ``Http2Stream`` that received the ``Http2ServerRequest``. If ``error`` is provided, an ``'error'`` event is emitted and ``error`` is passed as an argument to any listeners on the event.

It does nothing if the stream was already destroyed.

**Arguments**

* ``error`` (*Error*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#requestdestroyerror>`__

Http2Stream.end
~~~~~~~~~~~~~~~

.. code-block:: text

   response.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

This method signals to the server that all of the response headers and body have been sent; that server should consider this message complete. The method, ``response.end()``, MUST be called on each response.

If ``data`` is specified, it is equivalent to calling ``response.write(data, encoding)`` followed by ``response.end(callback)``.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responseenddata-encoding-callback>`__

Http2Stream.isPaused
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Stream.isPaused(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream.pause
~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Stream.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream.pushStream
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.pushStream(headers[, options], callback)

*Kind:* function · *Arguments:* 3

Initiates a push stream. The callback is invoked with the new ``Http2Stream`` instance created for the push stream passed as the second argument, or an ``Error`` passed as the first argument.

Setting the weight of a push stream is not allowed in the ``HEADERS`` frame. Pass a ``weight`` value to ``http2stream.priority`` with the ``silent`` option set to ``true`` to enable server-side bandwidth balancing between concurrent streams.

**Arguments**

* ``headers`` (*HTTP/2 Headers Object*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*): Callback that is called once the push stream has been

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const server = http2.createServer();
   server.on('stream', (stream) => {
     stream.respond({ ':status': 200 });
     stream.pushStream({ ':path': '/' }, (err, pushStream, headers) => {
       if (err) throw err;
       pushStream.respond({ ':status': 200 });
       pushStream.end('some pushed data');
     });
     stream.end('some data');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streampushstreamheaders-options-callback>`__

Http2Stream.read
~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Stream.read(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream.respond
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.respond([headers[, options]])

*Kind:* function · *Arguments:* 2

Initiates a response. When the ``options.waitForTrailers`` option is set, the ``'wantTrailers'`` event will be emitted immediately after queuing the last chunk of payload data to be sent. The ``http2stream.sendTrailers()`` method can then be used to send trailing header fields to the peer.

When ``options.waitForTrailers`` is set, the ``Http2Stream`` will not automatically close when the final ``DATA`` frame is transmitted. User code must call either ``http2stream.sendTrailers()`` or ``http2stream.close()`` to close the ``Http2Stream``.

**Arguments**

* ``headers`` (*HTTP/2 Headers Object|HTTP/2 Raw Headers*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const server = http2.createServer();
   server.on('stream', (stream) => {
     stream.respond({ ':status': 200 });
     stream.end('some data');
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamrespondheaders-options>`__

Http2Stream.respondWithFD
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.respondWithFD(fd[, headers[, options]])

*Kind:* function · *Arguments:* 3

Initiates a response whose data is read from the given file descriptor. No validation is performed on the given file descriptor. If an error occurs while attempting to read data using the file descriptor, the ``Http2Stream`` will be closed using an ``RST_STREAM`` frame using the standard ``INTERNAL_ERROR`` code.

When used, the ``Http2Stream`` object's ``Duplex`` interface will be closed automatically.

**Arguments**

* ``fd`` (*number|FileHandle*): A readable file descriptor.
* ``headers`` (*HTTP/2 Headers Object*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const fs = require('node:fs');

   const server = http2.createServer();
   server.on('stream', (stream) => {
     const fd = fs.openSync('/some/file', 'r');

     const stat = fs.fstatSync(fd);
     const headers = {
       'content-length': stat.size,
       'last-modified': stat.mtime.toUTCString(),
       'content-type': 'text/plain; charset=utf-8',
     };
     stream.respondWithFD(fd, headers);
     stream.on('close', () => fs.closeSync(fd));
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamrespondwithfdfd-headers-options>`__

Http2Stream.respondWithFile
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.respondWithFile(path[, headers[, options]])

*Kind:* function · *Arguments:* 3

Sends a regular file as the response. The ``path`` must specify a regular file or an ``'error'`` event will be emitted on the ``Http2Stream`` object.

When used, the ``Http2Stream`` object's ``Duplex`` interface will be closed automatically.

**Arguments**

* ``path`` (*string|Buffer|URL*)
* ``headers`` (*HTTP/2 Headers Object*) — optional
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const server = http2.createServer();
   server.on('stream', (stream) => {
     function statCheck(stat, headers) {
       headers['last-modified'] = stat.mtime.toUTCString();
     }

     function onError(err) {
       // stream.respond() can throw if the stream has been destroyed by
       // the other side.
       try {
         if (err.code === 'ENOENT') {
           stream.respond({ ':status': 404 });
         } else {
           stream.respond({ ':status': 500 });
         }
       } catch (err) {
         // Perform actual error handling.
         console.error(err);
       }
       stream.end();
     }

     stream.respondWithFile('/some/file',
                            { 'content-type': 'text/plain; charset=utf-8' },
                            { statCheck, onError });
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamrespondwithfilepath-headers-options>`__

Http2Stream.resume
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Stream.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2.Http2Stream.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Http2Stream.setTimeout
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http2stream.setTimeout(msecs, callback)

*Kind:* function · *Arguments:* 2

**Arguments**

* ``msecs`` (*number*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const http2 = require('node:http2');
   const client = http2.connect('http://example.org:8000');
   const { NGHTTP2_CANCEL } = http2.constants;
   const req = client.request({ ':path': '/' });

   // Cancel the stream if there's no activity after 5 seconds
   req.setTimeout(5000, () => req.close(NGHTTP2_CANCEL));

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#http2streamsettimeoutmsecs-callback>`__

Http2Stream.write
~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.write(chunk[, encoding][, callback])

*Kind:* function · *Arguments:* 3

If this method is called and ``response.writeHead()`` has not been called, it will switch to implicit header mode and flush the implicit headers.

This sends a chunk of the response body. This method may be called multiple times to provide successive parts of the body.

**Arguments**

* ``chunk`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http2.html#responsewritechunk-encoding-callback>`__

Properties and constants
------------------------

* ``constants`` — object
* ``constants.HTTP2_HEADER_AUTHORITY`` — string
* ``constants.HTTP2_HEADER_CONTENT_LENGTH`` — string
* ``constants.HTTP2_HEADER_CONTENT_TYPE`` — string
* ``constants.HTTP2_HEADER_METHOD`` — string
* ``constants.HTTP2_HEADER_PATH`` — string
* ``constants.HTTP2_HEADER_SCHEME`` — string
* ``constants.HTTP2_HEADER_STATUS`` — string
* ``constants.HTTP2_HEADER_USER_AGENT`` — string
* ``constants.NGHTTP2_CANCEL`` — number
* ``constants.NGHTTP2_COMPRESSION_ERROR`` — number
* ``constants.NGHTTP2_CONNECT_ERROR`` — number
* ``constants.NGHTTP2_ENHANCE_YOUR_CALM`` — number
* ``constants.NGHTTP2_FLOW_CONTROL_ERROR`` — number
* ``constants.NGHTTP2_FRAME_SIZE_ERROR`` — number
* ``constants.NGHTTP2_HTTP_1_1_REQUIRED`` — number
* ``constants.NGHTTP2_INADEQUATE_SECURITY`` — number
* ``constants.NGHTTP2_INTERNAL_ERROR`` — number
* ``constants.NGHTTP2_NO_ERROR`` — number
* ``constants.NGHTTP2_PROTOCOL_ERROR`` — number
* ``constants.NGHTTP2_REFUSED_STREAM`` — number
* ``constants.NGHTTP2_SETTINGS_TIMEOUT`` — number
* ``constants.NGHTTP2_STREAM_CLOSED`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
