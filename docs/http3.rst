HTTP/3
======

GodeJS includes a native HTTP/3 server through Go's ``quic-go`` package. Node
and Bun do not include an HTTP/3 server in their standard libraries.

Usage
-----

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

Generate a local certificate:

.. code-block:: sh

   openssl req -x509 -newkey rsa:2048 -nodes -days 1 \
     -keyout key.pem -out cert.pem -subj '/CN=localhost' \
     -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1'

Behavior
--------

``http3.createServer`` reuses the existing Node-style ``IncomingMessage`` and
``ServerResponse`` classes. Request URLs, headers, streaming writes,
``statusCode``, and ``server.close()`` behave like ``node:http``.

The listener is HTTP/3 only. Run an HTTP/1.1 or HTTP/2 listener alongside it if
clients need a fallback. UDP port 4433 must be reachable; ``curl`` 7.76 and
older cannot test HTTP/3, so use a QUIC client or the Go integration test.

Limitations
-----------

* TLS is required. There is no cleartext HTTP/3 server.
* The module does not advertise ``Alt-Svc`` for an HTTP/1.1 listener.
* HTTP/3 is implemented by GodeJS. It is not part of Node's API and is not
  available in Node or Bun without a native addon or proxy.
