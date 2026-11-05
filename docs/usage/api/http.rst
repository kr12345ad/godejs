http
====

This module, containing both a client and server, can be imported via ``require('node:http')`` (CommonJS) or ``import * as http from 'node:http'`` (ES module).

The HTTP interfaces in Node.js are designed to support many features of the protocol which have been traditionally difficult to use. In particular, large, possibly chunk-encoded, messages. The interface is careful to never buffer entire requests or responses, so the user is able to stream data.

Loading
-------

.. code-block:: javascript

   const http = require('node:http');
   // ESM
   import * as http from 'node:http';

API
---

_createSecureServer
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http._createSecureServer(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


_getSecure
~~~~~~~~~~

.. code-block:: text

   http._getSecure(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


_requestSecure
~~~~~~~~~~~~~~

.. code-block:: text

   http._requestSecure(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Agent
~~~~~

.. code-block:: text

   http.Agent

*Kind:* class · *Arguments:* 1

An ``Agent`` is responsible for managing connection persistence and reuse for HTTP clients. It maintains a queue of pending requests for a given host and port, reusing a single socket connection for each until the queue is empty, at which time the socket is either destroyed or put into a pool where it is kept to be used again for requests to the same host and port. Whether it is destroyed or pooled depends on the ``keepAlive`` \`option <#new-agentoptions>\`_.

Pooled connections have TCP Keep-Alive enabled for them, but servers may still close idle connections, in which case they will be removed from the pool and a new connection will be made when a new HTTP request is made for that host and port. Servers may also refuse to allow multiple requests over the same connection, in which case the connection will have to be remade for every request and cannot be pooled. The ``Agent`` will still make the requests to that server, but each one will occur over a new connection.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http = require('node:http');
   const keepAliveAgent = new http.Agent({ keepAlive: true });
   options.agent = keepAliveAgent;
   http.request(options, onResponseCallback);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#class-httpagent>`__

ClientRequest
~~~~~~~~~~~~~

.. code-block:: text

   http.ClientRequest

*Kind:* class · *Arguments:* 2

This object is created internally and returned from ``http.request()``. It represents an *in-progress* request whose header has already been queued. The header is still mutable using the ``setHeader(name, value)``, ``getHeader(name)``, ``removeHeader(name)`` API. The actual header will be sent along with the first data chunk or when calling ``request.end()``.

To get the response, add a listener for ``'response'`` to the request object. ``'response'`` will be emitted from the request object when the response headers have been received. The ``'response'`` event is executed with one argument which is an instance of ``http.IncomingMessage``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http = require('node:http');
   const net = require('node:net');
   const { URL } = require('node:url');

   // Create an HTTP tunneling proxy
   const proxy = http.createServer((req, res) => {
     res.writeHead(200, { 'Content-Type': 'text/plain' });
     res.end('okay');
   });
   proxy.on('connect', (req, clientSocket, head) => {
     // Connect to an origin server
     const { port, hostname } = new URL(`http://${req.url}`);
     const serverSocket = net.connect(port || 80, hostname, () => {
       clientSocket.write('HTTP/1.1 200 Connection Established\r\n' +
                       'Proxy-agent: Node.js-Proxy\r\n' +
                       '\r\n');
       serverSocket.write(head);
       serverSocket.pipe(clientSocket);
       clientSocket.pipe(serverSocket);
     });
   });

   // Now that proxy is running
   proxy.listen(1337, '127.0.0.1', () => {

     // Make a request to a tunneling proxy
     const options = {
       port: 1337,
       host: '127.0.0.1',
       method: 'CONNECT',
       path: 'www.google.com:80',
     };

     const req = http.request(options);
     req.end();

     req.on('connect', (res, socket, head) => {
       console.log('got connected!');

       // Make a request over an HTTP tunnel
       socket.write('GET / HTTP/1.1\r\n' +
                    'Host: www.google.com:80\r\n' +
                    'Connection: close\r\n' +
                    '\r\n');
       socket.on('data', (chunk) => {
         console.log(chunk.toString());
       });
       socket.on('end', () => {
         proxy.close();
       });
     });
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#class-httpclientrequest>`__

ClientResponse
~~~~~~~~~~~~~~

.. code-block:: text

   http.ClientResponse(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)

*Kind:* function · *Arguments:* 8

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


createServer
~~~~~~~~~~~~

.. code-block:: text

   http.createServer([options][, requestListener])

*Kind:* function · *Arguments:* 2

Returns a new instance of ``http.Server``.

The ``requestListener`` is a function which is automatically added to the ``'request'`` event.

**Arguments**

* ``options`` (*Object*) — optional
* ``requestListener`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const http = require('node:http');

   // Create a local server to receive data from
   const server = http.createServer((req, res) => {
     res.writeHead(200, { 'Content-Type': 'application/json' });
     res.end(JSON.stringify({
       data: 'Hello World!',
     }));
   });

   server.listen(8000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#httpcreateserveroptions-requestlistener>`__

get
~~~

.. code-block:: text

   http.get(options[, callback])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``options``
* ``callback`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#httpgetoptions-callback>`__

IncomingMessage
~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage

*Kind:* function · *Arguments:* 8

Type: Documentation-only

In a future version of Node.js, ``message.headers``, ``message.headersDistinct``, ``message.trailers``, and ``message.trailersDistinct`` will be read-only.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0171-setters-for-httpincomingmessage-headers-and-trailers>`__

request
~~~~~~~

.. code-block:: text

   http.request(options[, callback])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``options``
* ``callback`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#httprequestoptions-callback>`__

Server
~~~~~~

.. code-block:: text

   http.Server

*Kind:* class · *Arguments:* 2

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http = require('node:http');

   const server = http.createServer((req, res) => {
     res.end();
   });
   server.on('clientError', (err, socket) => {
     socket.end('HTTP/1.1 400 Bad Request\r\n\r\n');
   });
   server.listen(8000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#class-httpserver>`__

ServerResponse
~~~~~~~~~~~~~~

.. code-block:: text

   http.ServerResponse

*Kind:* class · *Arguments:* 2

This object is created internally by an HTTP server, not by the user. It is passed as the second parameter to the ``'request'`` event.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const http = require('node:http');
   const server = http.createServer((req, res) => {
     const ip = res.socket.remoteAddress;
     const port = res.socket.remotePort;
     res.end(`Your IP address is ${ip} and your source port is ${port}.`);
   }).listen(3000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#class-httpserverresponse>`__

validateHeaderName
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.validateHeaderName(name[, label])

*Kind:* function · *Arguments:* 2

Performs the low-level validations on the provided ``name`` that are done when ``res.setHeader(name, value)`` is called.

Passing illegal value as ``name`` will result in a ``TypeError`` being thrown, identified by ``code: 'ERR_INVALID_HTTP_TOKEN'``.

**Arguments**

* ``name`` (*string*)
* ``label`` (*string*) — optional: Label for error message. **Default:** ``'Header name'``.

**Example**

.. code-block:: javascript

   const { validateHeaderName } = require('node:http');

   try {
     validateHeaderName('');
   } catch (err) {
     console.error(err instanceof TypeError); // --> true
     console.error(err.code); // --> 'ERR_INVALID_HTTP_TOKEN'
     console.error(err.message); // --> 'Header name must be a valid HTTP token [""]'
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#httpvalidateheadernamename-label>`__

validateHeaderValue
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.validateHeaderValue(name, value)

*Kind:* function · *Arguments:* 2

Performs the low-level validations on the provided ``value`` that are done when ``res.setHeader(name, value)`` is called.

Passing illegal value as ``value`` will result in a ``TypeError`` being thrown.

**Arguments**

* ``name`` (*string*)
* ``value`` (*any*)

**Example**

.. code-block:: javascript

   const { validateHeaderValue } = require('node:http');

   try {
     validateHeaderValue('x-my-header', undefined);
   } catch (err) {
     console.error(err instanceof TypeError); // --> true
     console.error(err.code === 'ERR_HTTP_INVALID_HEADER_VALUE'); // --> true
     console.error(err.message); // --> 'Invalid value "undefined" for header "x-my-header"'
   }

   try {
     validateHeaderValue('x-my-header', 'oʊmɪɡə');
   } catch (err) {
     console.error(err instanceof TypeError); // --> true
     console.error(err.code === 'ERR_INVALID_CHAR'); // --> true
     console.error(err.message); // --> 'Invalid character in header content ["x-my-header"]'
   }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#httpvalidateheadervaluename-value>`__

Agent class
-----------

**Agent** — An ``Agent`` is responsible for managing connection persistence and reuse for HTTP clients. It maintains a queue of pending requests for a given host and port, reusing a single socket connection for each until the queue is empty, at which time the socket is either destroyed or put into a pool where it is kept to be used again for requests to the same host and port. Whether it is destroyed or pooled depends on the ``keepAlive`` \`option <#new-agentoptions>\`_.

Pooled connections have TCP Keep-Alive enabled for them, but servers may still close idle connections, in which case they will be removed from the pool and a new connection will be made when a new HTTP request is made for that host and port. Servers may also refuse to allow multiple requests over the same connection, in which case the connection will have to be remade for every request and cannot be pooled. The ``Agent`` will still make the requests to that server, but each one will occur over a new connection.

Agent.destroy
~~~~~~~~~~~~~

.. code-block:: text

   agent.destroy()

*Kind:* function · *Arguments:* 0

Destroy any sockets that are currently in use by the agent.

It is usually not necessary to do this. However, if using an agent with ``keepAlive`` enabled, then it is best to explicitly shut down the agent when it is no longer needed. Otherwise, sockets might stay open for quite a long time before the server terminates them.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#agentdestroy>`__

ClientRequest class
-------------------

**ClientRequest** — This object is created internally and returned from ``http.request()``. It represents an *in-progress* request whose header has already been queued. The header is still mutable using the ``setHeader(name, value)``, ``getHeader(name)``, ``removeHeader(name)`` API. The actual header will be sent along with the first data chunk or when calling ``request.end()``.

To get the response, add a listener for ``'response'`` to the request object. ``'response'`` will be emitted from the request object when the response headers have been received. The ``'response'`` event is executed with one argument which is an instance of ``http.IncomingMessage``.

ClientRequest._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.ClientRequest._headersObject(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ClientRequest.abort
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.abort()

*Kind:* function · *Arguments:* 0

Marks the request as aborting. Calling this will cause remaining data in the response to be dropped and the socket to be destroyed.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestabort>`__

ClientRequest.end
~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

Finishes sending the request. If any parts of the body are unsent, it will flush them to the stream. If the request is chunked, this will send the terminating ``'0\r\n\r\n'``.

If ``data`` is specified, it is equivalent to calling ``request.write(data, encoding)`` followed by ``request.end(callback)``.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestenddata-encoding-callback>`__

ClientRequest.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.flushHeaders()

*Kind:* function · *Arguments:* 0

Flushes the request headers.

For efficiency reasons, Node.js normally buffers the request headers until ``request.end()`` is called or the first chunk of request data is written. It then tries to pack the request headers and data into a single TCP packet.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestflushheaders>`__

ClientRequest.getHeader
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getHeader(name)

*Kind:* function · *Arguments:* 1

Reads out a header on the request. The name is case-insensitive. The type of the return value depends on the arguments provided to ``request.setHeader()``.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   request.setHeader('content-type', 'text/html');
   request.setHeader('Content-Length', Buffer.byteLength(body));
   request.setHeader('Cookie', ['type=ninja', 'language=javascript']);
   const contentType = request.getHeader('Content-Type');
   // 'contentType' is 'text/html'
   const contentLength = request.getHeader('Content-Length');
   // 'contentLength' is of type number
   const cookie = request.getHeader('Cookie');
   // 'cookie' is of type string[]

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetheadername>`__

ClientRequest.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getHeaders()

*Kind:* function · *Arguments:* 0

Returns a shallow copy of the current outgoing headers. Since a shallow copy is used, array values may be mutated without additional calls to various header-related http module methods. The keys of the returned object are the header names and the values are the respective header values. All header names are lowercase.

The object returned by the ``request.getHeaders()`` method *does not* prototypically inherit from the JavaScript ``Object``. This means that typical ``Object`` methods such as ``obj.toString()``, ``obj.hasOwnProperty()``, and others are not defined and *will not work*.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   request.setHeader('Foo', 'bar');
   request.setHeader('Cookie', ['foo=bar', 'bar=baz']);

   const headers = request.getHeaders();
   // headers === { foo: 'bar', 'cookie': ['foo=bar', 'bar=baz'] }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetheaders>`__

ClientRequest.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.removeHeader(name)

*Kind:* function · *Arguments:* 1

Removes a header that's already defined into headers object.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   request.removeHeader('Content-Type');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestremoveheadername>`__

ClientRequest.setHeader
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.setHeader(name, value)

*Kind:* function · *Arguments:* 2

Sets a single header value for headers object. If this header already exists in the to-be-sent headers, its value will be replaced. Use an array of strings here to send multiple headers with the same name. Non-string values will be stored without modification. Therefore, ``request.getHeader()`` may return non-string values. However, the non-string values will be converted to strings for network transmission.

or

**Arguments**

* ``name`` (*string*)
* ``value`` (*any*)

**Example**

.. code-block:: javascript

   request.setHeader('Content-Type', 'application/json');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestsetheadername-value>`__

ClientRequest.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.setTimeout(timeout[, callback])

*Kind:* function · *Arguments:* 2

Once a socket is assigned to this request and is connected ``socket.setTimeout()`` will be called.

**Arguments**

* ``timeout`` (*number*): Milliseconds before a request times out.
* ``callback`` (*Function*) — optional: Optional function to be called when a timeout occurs.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestsettimeouttimeout-callback>`__

ClientRequest.write
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.write(chunk[, encoding][, callback])

*Kind:* function · *Arguments:* 3

Sends a chunk of the body. This method can be called multiple times. If no ``Content-Length`` is set, data will automatically be encoded in HTTP Chunked transfer encoding, so that server knows when the data ends. The ``Transfer-Encoding: chunked`` header is added. Calling ``request.end()`` is necessary to finish sending the request.

The ``encoding`` argument is optional and only applies when ``chunk`` is a string. Defaults to ``'utf8'``.

**Arguments**

* ``chunk`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestwritechunk-encoding-callback>`__

IncomingMessage class
---------------------

**IncomingMessage** — Type: Documentation-only

In a future version of Node.js, ``message.headers``, ``message.headersDistinct``, ``message.trailers``, and ``message.trailersDistinct`` will be read-only.

IncomingMessage._completeWithoutListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage._completeWithoutListeners(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._emitBody
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage._emitBody(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._hasBodyListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage._hasBodyListeners(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage._parseHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage._parseHeaders(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.destroy
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Destroy the request. Optionally emit an ``'error'`` event, and emit a ``'close'`` event. Calling this will cause remaining data in the response to be dropped, and the socket to be destroyed if used, or returned to the corresponding Agent pool otherwise if possible.

See ``writable.destroy()`` for further details.

**Arguments**

* ``error`` (*Error*) — optional: Optional, an error to emit with ``'error'`` event.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestdestroyerror>`__

IncomingMessage.isPaused
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage.isPaused(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.pause
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.read
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage.read(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.resume
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


IncomingMessage.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.setTimeout(timeout[, callback])

*Kind:* function · *Arguments:* 2

Once a socket is assigned to this request and is connected ``socket.setTimeout()`` will be called.

**Arguments**

* ``timeout`` (*number*): Milliseconds before a request times out.
* ``callback`` (*Function*) — optional: Optional function to be called when a timeout occurs.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestsettimeouttimeout-callback>`__

Server class
------------

Server.addListener
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.Server.addListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.address
~~~~~~~~~~~~~~

.. code-block:: text

   server.address()

*Kind:* function · *Arguments:* 0

Returns the bound ``address``, the address ``family`` name, and ``port`` of the server as reported by the operating system if listening on an IP socket (useful to find which port was assigned when getting an OS-assigned address): ``{ port: 12346, family: 'IPv4', address: '127.0.0.1' }``.

For a server listening on a pipe or Unix domain socket, the name is returned as a string.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const server = net.createServer((socket) => {
     socket.end('goodbye\n');
   }).on('error', (err) => {
     // Handle errors here.
     throw err;
   });

   // Grab an arbitrary unused port.
   server.listen(() => {
     console.log('opened server on', server.address());
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/net.html#serveraddress>`__

Server.close
~~~~~~~~~~~~

.. code-block:: text

   server.close([callback])

*Kind:* function · *Arguments:* 1

Stops the server from accepting new connections and closes all connections connected to this server which are not sending a request or waiting for a response. See ``net.Server.close()``.

**Arguments**

* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const http = require('node:http');

   const server = http.createServer({ keepAliveTimeout: 60000 }, (req, res) => {
     res.writeHead(200, { 'Content-Type': 'application/json' });
     res.end(JSON.stringify({
       data: 'Hello World!',
     }));
   });

   server.listen(8000);
   // Close the server after 10 seconds
   setTimeout(() => {
     server.close(() => {
       console.log('server on port 8000 closed successfully');
     });
   }, 10000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#serverclosecallback>`__

Server.closeAllConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.closeAllConnections()

*Kind:* function · *Arguments:* 0

Closes all established HTTP(S) connections connected to this server, including active connections connected to this server which are sending a request or waiting for a response. This does *not* destroy sockets upgraded to a different protocol, such as WebSocket or HTTP/2.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const http = require('node:http');

   const server = http.createServer({ keepAliveTimeout: 60000 }, (req, res) => {
     res.writeHead(200, { 'Content-Type': 'application/json' });
     res.end(JSON.stringify({
       data: 'Hello World!',
     }));
   });

   server.listen(8000);
   // Close the server after 10 seconds
   setTimeout(() => {
     server.close(() => {
       console.log('server on port 8000 closed successfully');
     });
     // Closes all connections, ensuring the server closes successfully
     server.closeAllConnections();
   }, 10000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#servercloseallconnections>`__

Server.closeIdleConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.closeIdleConnections()

*Kind:* function · *Arguments:* 0

Closes all connections connected to this server which are not sending a request or waiting for a response.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const http = require('node:http');

   const server = http.createServer({ keepAliveTimeout: 60000 }, (req, res) => {
     res.writeHead(200, { 'Content-Type': 'application/json' });
     res.end(JSON.stringify({
       data: 'Hello World!',
     }));
   });

   server.listen(8000);
   // Close the server after 10 seconds
   setTimeout(() => {
     server.close(() => {
       console.log('server on port 8000 closed successfully');
     });
     // Closes idle connections, such as keep-alive connections. Server will close
     // once remaining active connections are terminated
     server.closeIdleConnections();
   }, 10000);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#servercloseidleconnections>`__

Server.getConnections
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.getConnections(callback)

*Kind:* function · *Arguments:* 1

Asynchronously get the number of concurrent connections on the server. Works when sockets were sent to forks.

Callback should take two arguments ``err`` and ``count``.

**Arguments**

* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/net.html#servergetconnectionscallback>`__

Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   server.listen()

*Kind:* function · *Arguments:* 0

Starts the HTTP server listening for connections. This method is identical to ``server.listen()`` from ``net.Server``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#serverlisten>`__

Server.off
~~~~~~~~~~

.. code-block:: text

   http.Server.off(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.on
~~~~~~~~~

.. code-block:: text

   http.Server.on(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.once
~~~~~~~~~~~

.. code-block:: text

   http.Server.once(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.prependListener
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.Server.prependListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.prependOnceListener
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.Server.prependOnceListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.ref
~~~~~~~~~~

.. code-block:: text

   server.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed server will *not* let the program exit if it's the only server left (the default behavior). If the server is ``ref``\ ed calling ``ref()`` again will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/net.html#serverref>`__

Server.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.Server.removeAllListeners(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.removeListener
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.Server.removeListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


Server.unref
~~~~~~~~~~~~

.. code-block:: text

   server.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a server will allow the program to exit if this is the only active server in the event system. If the server is already ``unref``\ ed calling ``unref()`` again will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/net.html#serverunref>`__

ServerResponse class
--------------------

**ServerResponse** — This object is created internally by an HTTP server, not by the user. It is passed as the second parameter to the ``'request'`` event.

ServerResponse._flushHead
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.ServerResponse._flushHead(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.ServerResponse._headersObject(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._packedHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.ServerResponse._packedHeaders(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse._sendHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   http.ServerResponse._sendHead(...args)

*Kind:* function · *Arguments:* 0

.. note::

   No matching Node.js 24 reference section was found. The signature and arity come from the runtime snapshot; this is likely a private helper or a GodeJS extension.


ServerResponse.addTrailers
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.addTrailers(headers)

*Kind:* function · *Arguments:* 1

This method adds HTTP trailing headers (a header but at the end of the message) to the response.

Trailers will **only** be emitted if chunked encoding is used for the response; if it is not (e.g. if the request was HTTP/1.0), they will be silently discarded.

**Arguments**

* ``headers`` (*Object*)

**Example**

.. code-block:: javascript

   response.writeHead(200, { 'Content-Type': 'text/plain',
                             'Trailer': 'Content-MD5' });
   response.write(fileData);
   response.addTrailers({ 'Content-MD5': '7895bf4b8828b55ceaf47747b4bca667' });
   response.end();

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#responseaddtrailersheaders>`__

ServerResponse.appendHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   outgoingMessage.appendHeader(name, value)

*Kind:* function · *Arguments:* 2

Append a single header value to the header object.

If the value is an array, this is equivalent to calling this method multiple times.

**Arguments**

* ``name`` (*string*): Header name
* ``value`` (*string|string\[]*): Header value

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#outgoingmessageappendheadername-value>`__

ServerResponse.destroy
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.destroy([error])

*Kind:* function · *Arguments:* 1

Destroy the request. Optionally emit an ``'error'`` event, and emit a ``'close'`` event. Calling this will cause remaining data in the response to be dropped, and the socket to be destroyed if used, or returned to the corresponding Agent pool otherwise if possible.

See ``writable.destroy()`` for further details.

**Arguments**

* ``error`` (*Error*) — optional: Optional, an error to emit with ``'error'`` event.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestdestroyerror>`__

ServerResponse.end
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

Finishes sending the request. If any parts of the body are unsent, it will flush them to the stream. If the request is chunked, this will send the terminating ``'0\r\n\r\n'``.

If ``data`` is specified, it is equivalent to calling ``request.write(data, encoding)`` followed by ``request.end(callback)``.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestenddata-encoding-callback>`__

ServerResponse.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.flushHeaders()

*Kind:* function · *Arguments:* 0

Flushes the request headers.

For efficiency reasons, Node.js normally buffers the request headers until ``request.end()`` is called or the first chunk of request data is written. It then tries to pack the request headers and data into a single TCP packet.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestflushheaders>`__

ServerResponse.getHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getHeader(name)

*Kind:* function · *Arguments:* 1

Reads out a header on the request. The name is case-insensitive. The type of the return value depends on the arguments provided to ``request.setHeader()``.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   request.setHeader('content-type', 'text/html');
   request.setHeader('Content-Length', Buffer.byteLength(body));
   request.setHeader('Cookie', ['type=ninja', 'language=javascript']);
   const contentType = request.getHeader('Content-Type');
   // 'contentType' is 'text/html'
   const contentLength = request.getHeader('Content-Length');
   // 'contentLength' is of type number
   const cookie = request.getHeader('Cookie');
   // 'cookie' is of type string[]

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetheadername>`__

ServerResponse.getHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getHeaderNames()

*Kind:* function · *Arguments:* 0

Returns an array containing the unique names of the current outgoing headers. All header names are lowercase.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   request.setHeader('Foo', 'bar');
   request.setHeader('Cookie', ['foo=bar', 'bar=baz']);

   const headerNames = request.getHeaderNames();
   // headerNames === ['foo', 'cookie']

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetheadernames>`__

ServerResponse.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getHeaders()

*Kind:* function · *Arguments:* 0

Returns a shallow copy of the current outgoing headers. Since a shallow copy is used, array values may be mutated without additional calls to various header-related http module methods. The keys of the returned object are the header names and the values are the respective header values. All header names are lowercase.

The object returned by the ``request.getHeaders()`` method *does not* prototypically inherit from the JavaScript ``Object``. This means that typical ``Object`` methods such as ``obj.toString()``, ``obj.hasOwnProperty()``, and others are not defined and *will not work*.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   request.setHeader('Foo', 'bar');
   request.setHeader('Cookie', ['foo=bar', 'bar=baz']);

   const headers = request.getHeaders();
   // headers === { foo: 'bar', 'cookie': ['foo=bar', 'bar=baz'] }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetheaders>`__

ServerResponse.getRawHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.getRawHeaderNames()

*Kind:* function · *Arguments:* 0

Returns an array containing the unique names of the current outgoing raw headers. Header names are returned with their exact casing being set.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   request.setHeader('Foo', 'bar');
   request.setHeader('Set-Cookie', ['foo=bar', 'bar=baz']);

   const headerNames = request.getRawHeaderNames();
   // headerNames === ['Foo', 'Set-Cookie']

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestgetrawheadernames>`__

ServerResponse.hasHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.hasHeader(name)

*Kind:* function · *Arguments:* 1

Returns ``true`` if the header identified by ``name`` is currently set in the outgoing headers. The header name matching is case-insensitive.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const hasContentType = request.hasHeader('content-type');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requesthasheadername>`__

ServerResponse.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.removeHeader(name)

*Kind:* function · *Arguments:* 1

Removes a header that's already defined into headers object.

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   request.removeHeader('Content-Type');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestremoveheadername>`__

ServerResponse.setHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.setHeader(name, value)

*Kind:* function · *Arguments:* 2

Sets a single header value for headers object. If this header already exists in the to-be-sent headers, its value will be replaced. Use an array of strings here to send multiple headers with the same name. Non-string values will be stored without modification. Therefore, ``request.getHeader()`` may return non-string values. However, the non-string values will be converted to strings for network transmission.

or

**Arguments**

* ``name`` (*string*)
* ``value`` (*any*)

**Example**

.. code-block:: javascript

   request.setHeader('Content-Type', 'application/json');

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestsetheadername-value>`__

ServerResponse.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.setTimeout(timeout[, callback])

*Kind:* function · *Arguments:* 2

Once a socket is assigned to this request and is connected ``socket.setTimeout()`` will be called.

**Arguments**

* ``timeout`` (*number*): Milliseconds before a request times out.
* ``callback`` (*Function*) — optional: Optional function to be called when a timeout occurs.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestsettimeouttimeout-callback>`__

ServerResponse.write
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   request.write(chunk[, encoding][, callback])

*Kind:* function · *Arguments:* 3

Sends a chunk of the body. This method can be called multiple times. If no ``Content-Length`` is set, data will automatically be encoded in HTTP Chunked transfer encoding, so that server knows when the data ends. The ``Transfer-Encoding: chunked`` header is added. Calling ``request.end()`` is necessary to finish sending the request.

The ``encoding`` argument is optional and only applies when ``chunk`` is a string. Defaults to ``'utf8'``.

**Arguments**

* ``chunk`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional
* ``callback`` (*Function*) — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#requestwritechunk-encoding-callback>`__

ServerResponse.writeContinue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.writeContinue()

*Kind:* function · *Arguments:* 0

Sends an HTTP/1.1 100 Continue message to the client, indicating that the request body should be sent. See the ``'checkContinue'`` event on ``Server``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#responsewritecontinue>`__

ServerResponse.writeHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.writeHead(statusCode[, statusMessage][, headers])

*Kind:* function · *Arguments:* 3

Sends a response header to the request. The status code is a 3-digit HTTP status code, like ``404``. The last argument, ``headers``, are the response headers. Optionally one can give a human-readable ``statusMessage`` as the second argument.

``headers`` may be an ``Array`` where the keys and values are in the same list. It is *not* a list of tuples. So, the even-numbered offsets are key values, and the odd-numbered offsets are the associated values. The array is in the same format as ``request.rawHeaders``.

**Arguments**

* ``statusCode`` (*number*)
* ``statusMessage`` (*string*) — optional
* ``headers`` (*Object|Array*) — optional

**Example**

.. code-block:: javascript

   const body = 'hello world';
   response
     .writeHead(200, {
       'Content-Length': Buffer.byteLength(body),
       'Content-Type': 'text/plain',
     })
     .end(body);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#responsewriteheadstatuscode-statusmessage-headers>`__

ServerResponse.writeProcessing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   response.writeProcessing()

*Kind:* function · *Arguments:* 0

Sends an HTTP/1.1 102 Processing message to the client, indicating that the request body should be sent.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/http.html#responsewriteprocessing>`__

Properties and constants
------------------------

* ``globalAgent`` — object
* ``globalAgent.keepAlive`` — boolean
* ``globalAgent.maxSockets`` — number
* ``globalAgent.options`` — object
* ``globalAgent.requests`` — object
* ``globalAgent.sockets`` — object
* ``IncomingMessage.connection`` — undefined
* ``IncomingMessage.headers`` — undefined
* ``IncomingMessage.headersDistinct`` — undefined
* ``IncomingMessage.rawHeaders`` — undefined
* ``IncomingMessage.socket`` — undefined
* ``IncomingMessage.trailers`` — undefined
* ``IncomingMessage.trailersDistinct`` — undefined
* ``maxHeaderSize`` — number
* ``METHODS`` — array
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
