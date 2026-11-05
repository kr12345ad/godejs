https
=====

HTTPS is the HTTP protocol over TLS/SSL. In Node.js this is implemented as a separate module.

Loading
-------

.. code-block:: javascript

   const https = require('node:https');
   // ESM
   import * as https from 'node:https';

API
---

_createSecureServer
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https._createSecureServer(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https._createSecureServer); // 'function'
   console.log(https._createSecureServer.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_getSecure
~~~~~~~~~~

.. code-block:: text

   https._getSecure(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https._getSecure); // 'function'
   console.log(https._getSecure.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


_requestSecure
~~~~~~~~~~~~~~

.. code-block:: text

   https._requestSecure(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https._requestSecure); // 'function'
   console.log(https._requestSecure.length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Agent
~~~~~

.. code-block:: text

   https.Agent

*Kind:* class · *Arguments:* 1

An ``Agent`` object for HTTPS similar to ``http.Agent``. See ``https.request()`` for more information.

Like ``http.Agent``, the ``createConnection(options[, callback])`` method can be overridden to customize how TLS connections are established.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   // ...
   https.globalAgent.on('keylog', (line, tlsSocket) => {
     fs.appendFileSync('/tmp/ssl-keys.log', line, { mode: 0o600 });
   });


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


ClientResponse
~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientResponse(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)

*Kind:* class · *Arguments:* 8

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientResponse(undefined, undefined, undefined, undefined, undefined, undefined, undefined, undefined);
   console.log(instance);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


createServer
~~~~~~~~~~~~

.. code-block:: text

   https.createServer([options][, requestListener])

*Kind:* function · *Arguments:* 2

Or

To generate the certificate and key for this example, run:

**Arguments**

* ``options`` (*Object*) — optional: Accepts ``options`` from ``tls.createServer()``,
* ``requestListener`` (*Function*) — optional: A listener to be added to the ``'request'`` event.

**Example**

.. code-block:: javascript

   // curl -k https://localhost:8000/
   const https = require('node:https');
   const fs = require('node:fs');

   const options = {
     key: fs.readFileSync('private-key.pem'),
     cert: fs.readFileSync('certificate.pem'),
   };

   https.createServer(options, (req, res) => {
     res.writeHead(200);
     res.end('hello world\n');
   }).listen(8000);


get
~~~

.. code-block:: text

   https.get(options[, callback])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``options``
* ``callback`` — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const result = https.get({}, () => {});
   console.log(result);


IncomingMessage
~~~~~~~~~~~~~~~

.. code-block:: text

   http.IncomingMessage

*Kind:* class · *Arguments:* 8

Type: Documentation-only

In a future version of Node.js, ``message.headers``, ``message.headersDistinct``, ``message.trailers``, and ``message.trailersDistinct`` will be read-only.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage(undefined, undefined, undefined, undefined, undefined, undefined, undefined, undefined);
   console.log(instance);


request
~~~~~~~

.. code-block:: text

   https.request(options[, callback])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``options``
* ``callback`` — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const result = https.request({}, () => {});
   console.log(result);


Server
~~~~~~

.. code-block:: text

   https.Server

*Kind:* class · *Arguments:* 2

See ``http.Server`` for more information.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server(undefined, undefined);
   console.log(instance);


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


Agent class
-----------

**Agent** — An ``Agent`` object for HTTPS similar to ``http.Agent``. See ``https.request()`` for more information.

Like ``http.Agent``, the ``createConnection(options[, callback])`` method can be overridden to customize how TLS connections are established.

Agent.destroy
~~~~~~~~~~~~~

.. code-block:: text

   agent.destroy()

*Kind:* function · *Arguments:* 0

Destroy any sockets that are currently in use by the agent.

It is usually not necessary to do this. However, if using an agent with ``keepAlive`` enabled, then it is best to explicitly shut down the agent when it is no longer needed. Otherwise, sockets might stay open for quite a long time before the server terminates them.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Agent();
   const result = instance.destroy();
   console.log(result);


ClientRequest class
-------------------

**ClientRequest** — This object is created internally and returned from ``http.request()``. It represents an *in-progress* request whose header has already been queued. The header is still mutable using the ``setHeader(name, value)``, ``getHeader(name)``, ``removeHeader(name)`` API. The actual header will be sent along with the first data chunk or when calling ``request.end()``.

To get the response, add a listener for ``'response'`` to the request object. ``'response'`` will be emitted from the request object when the response headers have been received. The ``'response'`` event is executed with one argument which is an instance of ``http.IncomingMessage``.

ClientRequest._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest._headersObject(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.ClientRequest.prototype._headersObject()); // 'function'
   console.log(https.ClientRequest.prototype._headersObject().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.abort
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.abort(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.abort();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.end
~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.end(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.flushHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.flushHeaders();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.getHeader
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.getHeader(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.getHeader(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.getHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.getHeaders();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.removeHeader(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.removeHeader(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.setHeader
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.setHeader(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.setHeader(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ClientRequest.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.setTimeout([msecs][, callback])

*Kind:* function · *Arguments:* 2

See ``server.setTimeout()`` in the ``node:http`` module.

**Arguments**

* ``msecs`` (*number*) — optional: **Default:** ``120000`` (2 minutes)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.setTimeout(1, () => {});
   console.log(result);


ClientRequest.write
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ClientRequest.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ClientRequest();
   const result = instance.write(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage class
---------------------

**IncomingMessage** — Type: Documentation-only

In a future version of Node.js, ``message.headers``, ``message.headersDistinct``, ``message.trailers``, and ``message.trailersDistinct`` will be read-only.

IncomingMessage._completeWithoutListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage._completeWithoutListeners(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.IncomingMessage.prototype._completeWithoutListeners()); // 'function'
   console.log(https.IncomingMessage.prototype._completeWithoutListeners().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage._emitBody
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage._emitBody(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.IncomingMessage.prototype._emitBody()); // 'function'
   console.log(https.IncomingMessage.prototype._emitBody().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage._hasBodyListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage._hasBodyListeners(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.IncomingMessage.prototype._hasBodyListeners()); // 'function'
   console.log(https.IncomingMessage.prototype._hasBodyListeners().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage._parseHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage._parseHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.IncomingMessage.prototype._parseHeaders()); // 'function'
   console.log(https.IncomingMessage.prototype._parseHeaders().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.destroy
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.destroy(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.destroy(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.isPaused
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.isPaused(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.isPaused();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.pause
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.pause(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.pause();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.read
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.read(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.read();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.resume
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.resume(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.resume();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.setEncoding
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.IncomingMessage.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.setEncoding(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


IncomingMessage.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.setTimeout([msecs][, callback])

*Kind:* function · *Arguments:* 2

See ``server.setTimeout()`` in the ``node:http`` module.

**Arguments**

* ``msecs`` (*number*) — optional: **Default:** ``120000`` (2 minutes)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.IncomingMessage();
   const result = instance.setTimeout(1, () => {});
   console.log(result);


Server class
------------

**Server** — See ``http.Server`` for more information.

Server.addListener
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.Server.addListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.addListener(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


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


Server.close
~~~~~~~~~~~~

.. code-block:: text

   server.close([callback])

*Kind:* function · *Arguments:* 1

See ``server.close()`` in the ``node:http`` module.

**Arguments**

* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.close(() => {});
   console.log(result);


Server.closeAllConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.closeAllConnections()

*Kind:* function · *Arguments:* 0

See ``server.closeAllConnections()`` in the ``node:http`` module.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.closeAllConnections();
   console.log(result);


Server.closeIdleConnections
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.closeIdleConnections()

*Kind:* function · *Arguments:* 0

See ``server.closeIdleConnections()`` in the ``node:http`` module.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.closeIdleConnections();
   console.log(result);


Server.getConnections
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.getConnections(callback)

*Kind:* function · *Arguments:* 1

Asynchronously get the number of concurrent connections on the server. Works when sockets were sent to forks.

Callback should take two arguments ``err`` and ``count``.

**Arguments**

* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.getConnections(() => {});
   console.log(result);


Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   server.listen()

*Kind:* function · *Arguments:* 0

Starts the HTTPS server listening for encrypted connections. This method is identical to ``server.listen()`` from ``net.Server``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.listen();
   console.log(result);


Server.off
~~~~~~~~~~

.. code-block:: text

   https.Server.off(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.off(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.on
~~~~~~~~~

.. code-block:: text

   https.Server.on(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.on(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.once
~~~~~~~~~~~

.. code-block:: text

   https.Server.once(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.once(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.prependListener
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.Server.prependListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.prependListener(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.prependOnceListener
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.Server.prependOnceListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.prependOnceListener(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.ref
~~~~~~~~~~

.. code-block:: text

   server.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed server will *not* let the program exit if it's the only server left (the default behavior). If the server is ``ref``\ ed calling ``ref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.ref();
   console.log(result);


Server.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.Server.removeAllListeners(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.removeAllListeners(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.removeListener
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.Server.removeListener(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.removeListener(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Server.unref
~~~~~~~~~~~~

.. code-block:: text

   server.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a server will allow the program to exit if this is the only active server in the event system. If the server is already ``unref``\ ed calling ``unref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.Server();
   const result = instance.unref();
   console.log(result);


ServerResponse class
--------------------

**ServerResponse** — This object is created internally by an HTTP server, not by the user. It is passed as the second parameter to the ``'request'`` event.

ServerResponse._flushHead
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse._flushHead(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.ServerResponse.prototype._flushHead()); // 'function'
   console.log(https.ServerResponse.prototype._flushHead().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse._headersObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse._headersObject(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.ServerResponse.prototype._headersObject()); // 'function'
   console.log(https.ServerResponse.prototype._headersObject().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse._packedHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse._packedHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.ServerResponse.prototype._packedHeaders()); // 'function'
   console.log(https.ServerResponse.prototype._packedHeaders().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse._sendHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse._sendHead(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   console.log(typeof https.ServerResponse.prototype._sendHead()); // 'function'
   console.log(https.ServerResponse.prototype._sendHead().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.addTrailers
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.addTrailers(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.addTrailers(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.appendHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.appendHeader(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.appendHeader(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.destroy
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.destroy(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.destroy(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.end
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.end(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.flushHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.flushHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.flushHeaders();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.getHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.getHeader(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.getHeader(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.getHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.getHeaderNames(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.getHeaderNames();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.getHeaders
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.getHeaders(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.getHeaders();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.getRawHeaderNames
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.getRawHeaderNames(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.getRawHeaderNames();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.hasHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.hasHeader(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.hasHeader(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.removeHeader
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.removeHeader(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.removeHeader(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.setHeader
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.setHeader(arg0, arg1)

*Kind:* function · *Arguments:* 2

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.setHeader(undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.setTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.setTimeout([msecs][, callback])

*Kind:* function · *Arguments:* 2

See ``server.setTimeout()`` in the ``node:http`` module.

**Arguments**

* ``msecs`` (*number*) — optional: **Default:** ``120000`` (2 minutes)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.setTimeout(1, () => {});
   console.log(result);


ServerResponse.write
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.write(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.writeContinue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.writeContinue(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.writeContinue();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.writeHead
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.writeHead(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.writeHead(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ServerResponse.writeProcessing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   https.ServerResponse.writeProcessing(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const https = require('node:https');

   const instance = new https.ServerResponse();
   const result = instance.writeProcessing();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


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

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
