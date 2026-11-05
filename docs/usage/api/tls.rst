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


Server.close
~~~~~~~~~~~~

.. code-block:: text

   server.close([callback])

*Kind:* function · *Arguments:* 1

The ``server.close()`` method stops the server from accepting new connections.

This function operates asynchronously. The ``'close'`` event will be emitted when the server has no more open connections.

**Arguments**

* ``callback`` (*Function*) — optional: A listener callback that will be registered to listen


Server.getConnections
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   server.getConnections(callback)

*Kind:* function · *Arguments:* 1

Asynchronously get the number of concurrent connections on the server. Works when sockets were sent to forks.

Callback should take two arguments ``err`` and ``count``.

**Arguments**

* ``callback`` (*Function*)


Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   server.listen()

*Kind:* function · *Arguments:* 0

Starts the server listening for encrypted connections. This method is identical to ``server.listen()`` from ``net.Server``.

This callable takes no arguments.


Server.ref
~~~~~~~~~~

.. code-block:: text

   server.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed server will *not* let the program exit if it's the only server left (the default behavior). If the server is ``ref``\ ed calling ``ref()`` again will have no effect.

This callable takes no arguments.


Server.unref
~~~~~~~~~~~~

.. code-block:: text

   server.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a server will allow the program to exit if this is the only active server in the event system. If the server is already ``unref``\ ed calling ``unref()`` again will have no effect.

This callable takes no arguments.


TLSSocket class
---------------

**TLSSocket** — Performs transparent encryption of written data and all required TLS negotiation.

Instances of ``tls.TLSSocket`` implement the duplex Stream interface.

TLSSocket._receive
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket._receive(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.address
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.address()

*Kind:* function · *Arguments:* 0

Returns the bound ``address``, the address ``family`` name, and ``port`` of the underlying socket as reported by the operating system: ``{ port: 12346, family: 'IPv4', address: '127.0.0.1' }``.

This callable takes no arguments.


TLSSocket.destroy
~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.destroy(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.end
~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.end(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

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

   { subject:
      { OU: [ 'Domain Control Validated', 'PositiveSSL Wildcard' ],
        CN: '*.example.com' },
     issuer:
      { C: 'GB',
        ST: 'Greater Manchester',
        L: 'Salford',
        O: 'COMODO CA Limited',
        CN: 'COMODO RSA Domain Validation Secure Server CA' },
     subjectaltname: 'DNS:*.example.com, DNS:example.com',
     infoAccess:
      { 'CA Issuers - URI':
         [ 'http://crt.comodoca.com/COMODORSADomainValidationSecureServerCA.crt' ],
        'OCSP - URI': [ 'http://ocsp.comodoca.com' ] },
     modulus: 'B56CE45CB740B09A13F64AC543B712FF9EE8E4C284B542A1708A27E82A8D151CA178153E12E6DDA15BF70FFD96CB8A88618641BDFCCA03527E665B70D779C8A349A6F88FD4EF6557180BD4C98192872BCFE3AF56E863C09DDD8BC1EC58DF9D94F914F0369102B2870BECFA1348A0838C9C49BD1C20124B442477572347047506B1FCD658A80D0C44BCC16BC5C5496CFE6E4A8428EF654CD3D8972BF6E5BFAD59C93006830B5EB1056BBB38B53D1464FA6E02BFDF2FF66CD949486F0775EC43034EC2602AEFBF1703AD221DAA2A88353C3B6A688EFE8387811F645CEED7B3FE46E1F8B9F59FAD028F349B9BC14211D5830994D055EEA3D547911E07A0ADDEB8A82B9188E58720D95CD478EEC9AF1F17BE8141BE80906F1A339445A7EB5B285F68039B0F294598A7D1C0005FC22B5271B0752F58CCDEF8C8FD856FB7AE21C80B8A2CE983AE94046E53EDE4CB89F42502D31B5360771C01C80155918637490550E3F555E2EE75CC8C636DDE3633CFEDD62E91BF0F7688273694EEEBA20C2FC9F14A2A435517BC1D7373922463409AB603295CEB0BB53787A334C9CA3CA8B30005C5A62FC0715083462E00719A8FA3ED0A9828C3871360A73F8B04A4FC1E71302844E9BB9940B77E745C9D91F226D71AFCAD4B113AAF68D92B24DDB4A2136B55A1CD1ADF39605B63CB639038ED0F4C987689866743A68769CC55847E4A06D6E2E3F1',
     exponent: '0x10001',
     pubkey: <Buffer ... >,
     valid_from: 'Aug 14 00:00:00 2017 GMT',
     valid_to: 'Nov 20 23:59:59 2019 GMT',
     fingerprint: '01:02:59:D9:C3:D2:0D:08:F7:82:4E:44:A4:B4:53:C5:E2:3A:87:4D',
     fingerprint256: '69:AE:1A:6A:D4:3D:C6:C1:1B:EA:C6:23:DE:BA:2A:14:62:62:93:5C:7A:EA:06:41:9B:0B:BC:87:CE:48:4E:02',
     fingerprint512: '19:2B:3E:C3:B3:5B:32:E8:AE:BB:78:97:27:E4:BA:6C:39:C9:92:79:4F:31:46:39:E2:70:E5:5F:89:42:17:C9:E8:64:CA:FF:BB:72:56:73:6E:28:8A:92:7E:A3:2A:15:8B:C2:E0:45:CA:C3:BC:EA:40:52:EC:CA:A2:68:CB:32',
     ext_key_usage: [ '1.3.6.1.5.5.7.3.1', '1.3.6.1.5.5.7.3.2' ],
     serialNumber: '66593D57F20CBC573E433381B5FEC280',
     raw: <Buffer ... > }


TLSSocket.getProtocol
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getProtocol()

*Kind:* function · *Arguments:* 0

Returns a string containing the negotiated SSL/TLS protocol version of the current connection. The value ``'unknown'`` will be returned for connected sockets that have not completed the handshaking process. The value ``null`` will be returned for server sockets or disconnected client sockets.

Protocol versions are:

This callable takes no arguments.


TLSSocket.getSession
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.getSession()

*Kind:* function · *Arguments:* 0

Returns the TLS session data or ``undefined`` if no session was negotiated. On the client, the data can be provided to the ``session`` option of ``tls.connect()`` to resume the connection. On the server, it may be useful for debugging.

See Session Resumption for more information.

This callable takes no arguments.


TLSSocket.isSessionReused
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tlsSocket.isSessionReused()

*Kind:* function · *Arguments:* 0

See Session Resumption for more information.

This callable takes no arguments.


TLSSocket.pause
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.pause(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.ref
~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.ref(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.resume
~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.resume(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.setEncoding
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.setEncoding(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.setServername
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.setServername(arg0)

*Kind:* function · *Arguments:* 1

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.unref
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.unref(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


TLSSocket.write
~~~~~~~~~~~~~~~

.. code-block:: text

   tls.TLSSocket.write(arg0, arg1, arg2)

*Kind:* function · *Arguments:* 3

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
