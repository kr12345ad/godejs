# HTTP/3 servers, only in GodeJS

Node and Bun ship HTTP/1.1 and HTTP/2 servers. GodeJS adds a native **QUIC /
HTTP/3** server backed by Go's `quic-go`, while keeping the same Node-style
`IncomingMessage` and `ServerResponse` objects used by `node:http`.

```js
const fs = require('node:fs');
const http3 = require('gode:http3');

const server = http3.createServer({
  key: fs.readFileSync('key.pem'),
  cert: fs.readFileSync('cert.pem'),
}, (req, res) => {
  res.setHeader('content-type', 'application/json');
  res.end(JSON.stringify({ ok: true, url: req.url, version: req.httpVersion }));
});

server.listen(4433, '0.0.0.0', () => {
  console.log('HTTP/3 on', server.address());
});
```

Generate a local certificate for development:

```sh
openssl req -x509 -newkey rsa:2048 -nodes -days 1 \
  -keyout key.pem -out cert.pem -subj '/CN=localhost' \
  -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1'
```

Test with a client that supports HTTP/3, or use the Go integration test in
`modules/http/http3_test.go`.

## API

`gode:http3` exports:

- `createServer(options, listener)`
- `createSecureServer(options, listener)` (alias)
- `Server`
- `IncomingMessage`
- `ServerResponse`
- `STATUS_CODES`

The handler receives the same request/response classes as `node:http`, including
streaming `write()`, `setHeader()`, `statusCode`, `req.url`, `req.method`, and
body events. Only the transport changes: QUIC over UDP instead of TCP.

## Why this is useful

- **Lower latency on lossy networks.** QUIC avoids head-of-line blocking between
  independent HTTP/3 streams.
- **One runtime for HTTP/1.1, HTTP/2, and HTTP/3.** Reuse the same handler logic
  and streaming response code across protocols.
- **Go's QUIC implementation.** Connection management, TLS 1.3, and congestion
  control live in Go, not in a native addon.

Node and Bun cannot do this with their built-in HTTP server APIs. They need a
native addon or an external proxy to terminate HTTP/3.

## Operational notes

- HTTP/3 always uses TLS. `key` and `cert` are required.
- Open the UDP port, not just TCP. A firewall that only allows TCP will make the
  server appear offline to HTTP/3 clients.
- This module is HTTP/3 only. It does not advertise `Alt-Svc` or fall back to
  HTTP/1.1; run an HTTP/1.1 or HTTP/2 listener alongside it when clients need a
  fallback.
- `server.close()` closes the QUIC listener and active connections. `address()`
  returns the UDP address, including the port chosen when listening on port 0.
