tls
===

The ``node:tls`` module provides an implementation of the Transport Layer Security (TLS) and Secure Socket Layer (SSL) protocols that is built on top of OpenSSL. The module can be accessed using:

Loading
-------

.. code-block:: javascript

   const tls = require('node:tls');
   // ESM
   import * as tls from 'node:tls';

Quick start
-----------

.. code-block:: javascript

   const tls = require('node:tls');

API
---

checkServerIdentity
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.checkServerIdentity(hostname, cert)

*Kind:* function · *Arguments:* 2

Verifies the certificate ``cert`` is issued to ``hostname``.

Returns ``Error`` object, populating it with ``reason``, ``host``, and ``cert`` on failure. On success, returns ``undefined``.

**Arguments**

* ``hostname`` (*string*): The host name or IP address to verify the certificate
* ``cert`` (*Object*): A certificate object representing the peer's certificate.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const result = tls.checkServerIdentity('127.0.0.1', {});
   console.log(result);


connect
~~~~~~~

.. code-block:: text

   tls.connect(options[, callback])

*Kind:* function · *Arguments:* 2

The ``callback`` function, if specified, will be added as a listener for the ``'secureConnect'`` event.

``tls.connect()`` returns a ``tls.TLSSocket`` object.

**Arguments**

* ``options`` (*Object*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   // Assumes an echo server that is listening on port 8000.
   const { connect } = require('node:tls');
   const { readFileSync } = require('node:fs');

   const options = {
     // Necessary only if the server requires client certificate authentication.
     key: readFileSync('client-key.pem'),
     cert: readFileSync('client-cert.pem'),

     // Necessary only if the server uses a self-signed certificate.
     ca: [ readFileSync('server-cert.pem') ],

     // Necessary only if the server's cert isn't for "localhost".
     checkServerIdentity: () => { return null; },
   };

   const socket = connect(8000, options, () => {
     console.log('client connected',
                 socket.authorized ? 'authorized' : 'unauthorized');
     process.stdin.pipe(socket);
     process.stdin.resume();
   });
   socket.setEncoding('utf8');
   socket.on('data', (data) => {
     console.log(data);
   });
   socket.on('end', () => {
     console.log('server ends connection');
   });


createSecureContext
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.createSecureContext([options])

*Kind:* function · *Arguments:* 1

``tls.createServer()`` sets the default value of the ``honorCipherOrder`` option to ``true``, other APIs that create secure contexts leave it unset.

``tls.createServer()`` uses a 128 bit truncated SHA1 hash value generated from ``process.argv`` as the default value of the ``sessionIdContext`` option, other APIs that create secure contexts have no default value.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const result = tls.createSecureContext({});
   console.log(result);


createServer
~~~~~~~~~~~~

.. code-block:: text

   tls.createServer([options][, secureConnectionListener])

*Kind:* function · *Arguments:* 2

Creates a new ``tls.Server``. The ``secureConnectionListener``, if provided, is automatically set as a listener for the ``'secureConnection'`` event.

The ``ticketKeys`` option is automatically shared between ``node:cluster`` module workers.

**Arguments**

* ``options`` (*Object*) — optional
* ``secureConnectionListener`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const { createServer } = require('node:tls');
   const { readFileSync } = require('node:fs');

   const options = {
     key: readFileSync('server-key.pem'),
     cert: readFileSync('server-cert.pem'),

     // This is necessary only if using client certificate authentication.
     requestCert: true,

     // This is necessary only if the client uses a self-signed certificate.
     ca: [ readFileSync('client-cert.pem') ],
   };

   const server = createServer(options, (socket) => {
     console.log('server connected',
                 socket.authorized ? 'authorized' : 'unauthorized');
     socket.write('welcome!\n');
     socket.setEncoding('utf8');
     socket.pipe(socket);
   });
   server.listen(8000, () => {
     console.log('server bound');
   });


getCiphers
~~~~~~~~~~

.. code-block:: text

   tls.getCiphers()

*Kind:* function · *Arguments:* 0

Returns an array with the names of the supported TLS ciphers. The names are lower-case for historical reasons, but must be uppercased to be used in the ``ciphers`` option of ``tls.createSecureContext()``.

Not all supported ciphers are enabled by default. See Modifying the default TLS cipher suite.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   console.log(tls.getCiphers()); // ['aes128-gcm-sha256', 'aes128-sha', ...]


Server
~~~~~~

.. code-block:: text

   tls.Server

*Kind:* class · *Arguments:* 2

Accepts encrypted connections using TLS or SSL.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const logFile = fs.createWriteStream('/tmp/ssl-keys.log', { flags: 'a' });
   // ...
   server.on('keylog', (line, tlsSocket) => {
     if (tlsSocket.remoteAddress !== '...')
       return; // Only log keys for a particular IP
     logFile.write(line);
   });


TLSSocket
~~~~~~~~~

.. code-block:: text

   tls.TLSSocket

*Kind:* class · *Arguments:* 1

Performs transparent encryption of written data and all required TLS negotiation.

Instances of ``tls.TLSSocket`` implement the duplex Stream interface.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const logFile = fs.createWriteStream('/tmp/ssl-keys.log', { flags: 'a' });
   // ...
   tlsSocket.on('keylog', (line) => logFile.write(line));


Server class
------------

**Server** — Accepts encrypted connections using TLS or SSL.

Server.address
~~~~~~~~~~~~~~

.. code-block:: text

   server.address()

*Kind:* function · *Arguments:* 0

Returns the bound address, the address family name, and port of the server as reported by the operating system. See ``net.Server.address()`` for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.address();
   console.log(result);


Server.close
~~~~~~~~~~~~

.. code-block:: text

   server.close([callback])

*Kind:* function · *Arguments:* 1

The ``server.close()`` method stops the server from accepting new connections.

This function operates asynchronously. The ``'close'`` event will be emitted when the server has no more open connections.

**Arguments**

* ``callback`` (*Function*) — optional: A listener callback that will be registered to listen

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.close(() => {});
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

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.getConnections(() => {});
   console.log(result);


Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   server.listen()

*Kind:* function · *Arguments:* 0

Starts the server listening for encrypted connections. This method is identical to ``server.listen()`` from ``net.Server``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.listen();
   console.log(result);


Server.ref
~~~~~~~~~~

.. code-block:: text

   server.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed server will *not* let the program exit if it's the only server left (the default behavior). If the server is ``ref``\ ed calling ``ref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.ref();
   console.log(result);


Server.unref
~~~~~~~~~~~~

.. code-block:: text

   server.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a server will allow the program to exit if this is the only active server in the event system. If the server is already ``unref``\ ed calling ``unref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.Server();
   const result = instance.unref();
   console.log(result);


TLSSocket class
---------------

**TLSSocket** — Performs transparent encryption of written data and all required TLS negotiation.

Instances of ``tls.TLSSocket`` implement the duplex Stream interface.

TLSSocket._receive
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket._receive(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   console.log(typeof tls.TLSSocket.prototype._receive()); // 'function'
   console.log(tls.TLSSocket.prototype._receive().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.address
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.address()

*Kind:* function · *Arguments:* 0

Returns the bound ``address``, the address ``family`` name, and ``port`` of the underlying socket as reported by the operating system: ``{ port: 12346, family: 'IPv4', address: '127.0.0.1' }``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.address();
   console.log(result);


TLSSocket.destroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.destroy(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.destroy(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.end
~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.end(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.getCipher
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getCipher()

*Kind:* function · *Arguments:* 0

Returns an object containing information on the negotiated cipher suite.

For example, a TLSv1.2 protocol with AES256-SHA cipher:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.getCipher();
   console.log(result);


TLSSocket.getPeerCertificate
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getPeerCertificate([detailed])

*Kind:* function · *Arguments:* 1

Returns an object representing the peer's certificate. If the peer does not provide a certificate, an empty object will be returned. If the socket has been destroyed, ``null`` will be returned.

If the full certificate chain was requested, each certificate will include an ``issuerCertificate`` property containing an object representing its issuer's certificate.

**Arguments**

* ``detailed`` (*boolean*) — optional: Include the full certificate chain if ``true``, otherwise

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.getPeerCertificate(true);
   console.log(result);


TLSSocket.getProtocol
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getProtocol()

*Kind:* function · *Arguments:* 0

Returns a string containing the negotiated SSL/TLS protocol version of the current connection. The value ``'unknown'`` will be returned for connected sockets that have not completed the handshaking process. The value ``null`` will be returned for server sockets or disconnected client sockets.

Protocol versions are:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.getProtocol();
   console.log(result);


TLSSocket.getSession
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getSession()

*Kind:* function · *Arguments:* 0

Returns the TLS session data or ``undefined`` if no session was negotiated. On the client, the data can be provided to the ``session`` option of ``tls.connect()`` to resume the connection. On the server, it may be useful for debugging.

See Session Resumption for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.getSession();
   console.log(result);


TLSSocket.isSessionReused
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.isSessionReused()

*Kind:* function · *Arguments:* 0

See Session Resumption for more information.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.isSessionReused();
   console.log(result);


TLSSocket.pause
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.pause(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.pause();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.ref
~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.ref(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.ref();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.resume
~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.resume(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.resume();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.setEncoding
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.setEncoding(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.setServername
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.setServername(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.setServername(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.unref
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.unref(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.unref();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.write
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

**Example**

.. code-block:: javascript

   const tls = require('node:tls');

   const instance = new tls.TLSSocket();
   const result = instance.write(undefined, undefined, undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Properties and constants
------------------------

* ``CLIENT_RENEG_LIMIT`` — number
* ``CLIENT_RENEG_WINDOW`` — number
* ``DEFAULT_ECDH_CURVE`` — string
* ``DEFAULT_MAX_VERSION`` — string
* ``DEFAULT_MIN_VERSION`` — string
* ``rootCertificates`` — array

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
