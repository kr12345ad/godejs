net
===

The ``node:net`` module provides an asynchronous network API for creating stream-based TCP or IPC servers (``net.createServer()``) and clients (``net.createConnection()``).

It can be accessed using:

Loading
-------

.. code-block:: javascript

   const net = require('node:net');
   // ESM
   import * as net from 'node:net';

Quick start
-----------

.. code-block:: javascript

   const net = require('node:net');

API
---

BlockList
~~~~~~~~~

.. code-block:: text

   net.BlockList

*Kind:* class · *Arguments:* 0

The ``BlockList`` object can be used with some network APIs to specify rules for disabling inbound or outbound access to specific IP addresses, IP ranges, or IP subnets.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const blockList = new net.BlockList();
   blockList.addAddress('123.123.123.123');
   blockList.addRange('10.0.0.1', '10.0.0.10');
   blockList.addSubnet('8592:757c:efae:4e45::', 64, 'ipv6');

   console.log(blockList.check('123.123.123.123'));  // Prints: true
   console.log(blockList.check('10.0.0.3'));  // Prints: true
   console.log(blockList.check('222.111.111.222'));  // Prints: false

   // IPv6 notation for IPv4 addresses works:
   console.log(blockList.check('::ffff:7b7b:7b7b', 'ipv6')); // Prints: true
   console.log(blockList.check('::ffff:123.123.123.123', 'ipv6')); // Prints: true


connect
~~~~~~~

.. code-block:: text

   net.connect()

*Kind:* function · *Arguments:* 0

Aliases to ``net.createConnection()``.

Possible signatures:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const result = net.connect();
   console.log(result);


createConnection
~~~~~~~~~~~~~~~~

.. code-block:: text

   net.createConnection()

*Kind:* function · *Arguments:* 0

A factory function, which creates a new ``net.Socket``, immediately initiates connection with ``socket.connect()``, then returns the ``net.Socket`` that starts the connection.

When the connection is established, a ``'connect'`` event will be emitted on the returned socket. The last parameter ``connectListener``, if supplied, will be added as a listener for the ``'connect'`` event **once**.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');
   const client = net.createConnection({ port: 8124 }, () => {
     // 'connect' listener.
     console.log('connected to server!');
     client.write('world!\r\n');
   });
   client.on('data', (data) => {
     console.log(data.toString());
     client.end();
   });
   client.on('end', () => {
     console.log('disconnected from server');
   });


createServer
~~~~~~~~~~~~

.. code-block:: text

   net.createServer([options][, connectionListener])

*Kind:* function · *Arguments:* 2

Creates a new TCP or IPC server.

If ``allowHalfOpen`` is set to ``true``, when the other end of the socket signals the end of transmission, the server will only send back the end of transmission when ``socket.end()`` is explicitly called. For example, in the context of TCP, when a FIN packet is received, a FIN packet is sent back only when ``socket.end()`` is explicitly called. Until then the connection is half-closed (non-readable but still writable). See ``'end'`` event and RFC 1122 (section 4.2.2.13) for more information.

**Arguments**

* ``options`` (*Object*) — optional
* ``connectionListener`` (*Function*) — optional: Automatically set as a listener for the

**Example**

.. code-block:: javascript

   const net = require('node:net');
   const server = net.createServer((c) => {
     // 'connection' listener.
     console.log('client connected');
     c.on('end', () => {
       console.log('client disconnected');
     });
     c.write('hello\r\n');
     c.pipe(c);
   });
   server.on('error', (err) => {
     throw err;
   });
   server.listen(8124, () => {
     console.log('server bound');
   });


getDefaultAutoSelectFamily
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   net.getDefaultAutoSelectFamily()

*Kind:* function · *Arguments:* 0

Gets the current default value of the ``autoSelectFamily`` option of ``socket.connect(options)``. The initial default value is ``true``, unless the command line option ``--no-network-family-autoselection`` is provided.

* Returns: ``boolean`` The current default value of the ``autoSelectFamily`` option.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const result = net.getDefaultAutoSelectFamily();
   console.log(result);


getDefaultAutoSelectFamilyAttemptTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   net.getDefaultAutoSelectFamilyAttemptTimeout()

*Kind:* function · *Arguments:* 0

Gets the current default value of the ``autoSelectFamilyAttemptTimeout`` option of ``socket.connect(options)``. The initial default value is ``250`` or the value specified via the command line option ``--network-family-autoselection-attempt-timeout``.

* Returns: ``number`` The current default value of the ``autoSelectFamilyAttemptTimeout`` option.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const result = net.getDefaultAutoSelectFamilyAttemptTimeout();
   console.log(result);


isIP
~~~~

.. code-block:: text

   net.isIP(input)

*Kind:* function · *Arguments:* 1

Returns ``6`` if ``input`` is an IPv6 address, including an IPv4-mapped IPv6 address. Returns ``4`` if ``input`` is an IPv4 address in dot-decimal notation with no leading zeroes. Otherwise, returns ``0``.

**Arguments**

* ``input`` (*string*)

**Example**

.. code-block:: javascript

   net.isIP('::1'); // returns 6
   net.isIP('::ffff:127.0.0.1'); // returns 6
   net.isIP('127.0.0.1'); // returns 4
   net.isIP('127.000.000.001'); // returns 0
   net.isIP('127.0.0.1/24'); // returns 0
   net.isIP('fhqwhgads'); // returns 0


isIPv4
~~~~~~

.. code-block:: text

   net.isIPv4(input)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``input`` is an IPv4 address in dot-decimal notation with no leading zeroes. Otherwise, returns ``false``.

**Arguments**

* ``input`` (*string*)

**Example**

.. code-block:: javascript

   net.isIPv4('127.0.0.1'); // returns true
   net.isIPv4('127.000.000.001'); // returns false
   net.isIPv4('127.0.0.1/24'); // returns false
   net.isIPv4('fhqwhgads'); // returns false


isIPv6
~~~~~~

.. code-block:: text

   net.isIPv6(input)

*Kind:* function · *Arguments:* 1

Returns ``true`` if ``input`` is an IPv6 address, including an IPv4-mapped IPv6 address. Otherwise, returns ``false``.

**Arguments**

* ``input`` (*string*)

**Example**

.. code-block:: javascript

   net.isIPv6('::1'); // returns true
   net.isIPv6('::ffff:127.0.0.1'); // returns true
   net.isIPv6('fhqwhgads'); // returns false


Server
~~~~~~

.. code-block:: text

   net.Server

*Kind:* class · *Arguments:* 2

Type: End-of-Life

Accessing several internal, undocumented properties of ``net.Server`` instances with inappropriate names is deprecated.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Server(undefined, undefined);
   console.log(instance);


setDefaultAutoSelectFamily
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   net.setDefaultAutoSelectFamily(value)

*Kind:* function · *Arguments:* 1

Sets the default value of the ``autoSelectFamily`` option of ``socket.connect(options)``.

**Arguments**

* ``value`` (*boolean*): The new default value.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const result = net.setDefaultAutoSelectFamily('hello');
   console.log(result);


setDefaultAutoSelectFamilyAttemptTimeout
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   net.setDefaultAutoSelectFamilyAttemptTimeout(value)

*Kind:* function · *Arguments:* 1

Sets the default value of the ``autoSelectFamilyAttemptTimeout`` option of ``socket.connect(options)``.

**Arguments**

* ``value`` (*number*): The new default value, which must be a positive number. If the number is less than ``10``,

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const result = net.setDefaultAutoSelectFamilyAttemptTimeout('hello');
   console.log(result);


Socket
~~~~~~

.. code-block:: text

   net.Socket

*Kind:* class · *Arguments:* 1

This class is an abstraction of a TCP socket or a streaming IPC endpoint (uses named pipes on Windows, and Unix domain sockets otherwise). It is also an ``EventEmitter``.

A ``net.Socket`` can be created by the user and used directly to interact with a server. For example, it is returned by ``net.createConnection()``, so the user can use it to talk to the server.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const net = require('node:net');
   const { Worker } = require('node:worker_threads');

   // worker.js receives `{ socket }` messages and handles each connection.
   const worker = new Worker('./worker.js');

   const server = net.createServer((socket) => {
     // Hand the freshly accepted connection off to the worker thread.
     worker.postMessage({ socket }, [socket]);
   });
   server.listen(8000);


SocketAddress
~~~~~~~~~~~~~

.. code-block:: text

   net.SocketAddress

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.SocketAddress(undefined);
   console.log(instance);


BlockList class
---------------

**BlockList** — The ``BlockList`` object can be used with some network APIs to specify rules for disabling inbound or outbound access to specific IP addresses, IP ranges, or IP subnets.

BlockList.addAddress
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   blockList.addAddress(address[, type])

*Kind:* function · *Arguments:* 2

Adds a rule to block the given IP address.

**Arguments**

* ``address`` (*string|net.SocketAddress*): An IPv4 or IPv6 address.
* ``type`` (*string*) — optional: Either ``'ipv4'`` or ``'ipv6'``. **Default:** ``'ipv4'``.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.BlockList();
   const result = instance.addAddress('127.0.0.1', 'example');
   console.log(result);


BlockList.addRange
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   blockList.addRange(start, end[, type])

*Kind:* function · *Arguments:* 3

Adds a rule to block a range of IP addresses from ``start`` (inclusive) to ``end`` (inclusive).

**Arguments**

* ``start`` (*string|net.SocketAddress*): The starting IPv4 or IPv6 address in the
* ``end`` (*string|net.SocketAddress*): The ending IPv4 or IPv6 address in the range.
* ``type`` (*string*) — optional: Either ``'ipv4'`` or ``'ipv6'``. **Default:** ``'ipv4'``.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.BlockList();
   const result = instance.addRange(0, 0, 'example');
   console.log(result);


BlockList.addSubnet
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   blockList.addSubnet(net, prefix[, type])

*Kind:* function · *Arguments:* 3

Adds a rule to block a range of IP addresses specified as a subnet mask.

**Arguments**

* ``net`` (*string|net.SocketAddress*): The network IPv4 or IPv6 address.
* ``prefix`` (*number*): The number of CIDR prefix bits. For IPv4, this
* ``type`` (*string*) — optional: Either ``'ipv4'`` or ``'ipv6'``. **Default:** ``'ipv4'``.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.BlockList();
   const result = instance.addSubnet('value', 1, 'example');
   console.log(result);


BlockList.check
~~~~~~~~~~~~~~~

.. code-block:: text

   blockList.check(address[, type])

*Kind:* function · *Arguments:* 2

Returns ``true`` if the given IP address matches any of the rules added to the ``BlockList``.

**Arguments**

* ``address`` (*string|net.SocketAddress*): The IP address to check
* ``type`` (*string*) — optional: Either ``'ipv4'`` or ``'ipv6'``. **Default:** ``'ipv4'``.

**Example**

.. code-block:: javascript

   const blockList = new net.BlockList();
   blockList.addAddress('123.123.123.123');
   blockList.addRange('10.0.0.1', '10.0.0.10');
   blockList.addSubnet('8592:757c:efae:4e45::', 64, 'ipv6');

   console.log(blockList.check('123.123.123.123'));  // Prints: true
   console.log(blockList.check('10.0.0.3'));  // Prints: true
   console.log(blockList.check('222.111.111.222'));  // Prints: false

   // IPv6 notation for IPv4 addresses works:
   console.log(blockList.check('::ffff:7b7b:7b7b', 'ipv6')); // Prints: true
   console.log(blockList.check('::ffff:123.123.123.123', 'ipv6')); // Prints: true


Server class
------------

**Server** — Type: End-of-Life

Accessing several internal, undocumented properties of ``net.Server`` instances with inappropriate names is deprecated.

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

Stops the server from accepting new connections and keeps existing connections. This function is asynchronous, the server is finally closed when all connections are ended and the server emits a ``'close'`` event. The optional ``callback`` will be called once the ``'close'`` event occurs. Unlike that event, it will be called with an ``Error`` as its only argument if the server was not open when it was closed.

**Arguments**

* ``callback`` (*Function*) — optional: Called when the server is closed.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Server();
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

   const net = require('node:net');

   const instance = new net.Server();
   const result = instance.getConnections(() => {});
   console.log(result);


Server.listen
~~~~~~~~~~~~~

.. code-block:: text

   server.listen()

*Kind:* function · *Arguments:* 0

Start a server listening for connections. A ``net.Server`` can be a TCP or an IPC server depending on what it listens to.

Possible signatures:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   server.on('error', (e) => {
     if (e.code === 'EADDRINUSE') {
       console.error('Address in use, retrying...');
       setTimeout(() => {
         server.close();
         server.listen(PORT, HOST);
       }, 1000);
     }
   });


Server.ref
~~~~~~~~~~

.. code-block:: text

   server.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed server will *not* let the program exit if it's the only server left (the default behavior). If the server is ``ref``\ ed calling ``ref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Server();
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

   const net = require('node:net');

   const instance = new net.Server();
   const result = instance.unref();
   console.log(result);


Socket class
------------

**Socket** — This class is an abstraction of a TCP socket or a streaming IPC endpoint (uses named pipes on Windows, and Unix domain sockets otherwise). It is also an ``EventEmitter``.

A ``net.Socket`` can be created by the user and used directly to interact with a server. For example, it is returned by ``net.createConnection()``, so the user can use it to talk to the server.

Socket.address
~~~~~~~~~~~~~~

.. code-block:: text

   socket.address()

*Kind:* function · *Arguments:* 0

Returns the bound ``address``, the address ``family`` name and ``port`` of the socket as reported by the operating system: ``{ port: 12346, family: 'IPv4', address: '127.0.0.1' }``

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.address();
   console.log(result);


Socket.connect
~~~~~~~~~~~~~~

.. code-block:: text

   socket.connect()

*Kind:* function · *Arguments:* 0

Initiate a connection on a given socket.

Possible signatures:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.connect();
   console.log(result);


Socket.destroy
~~~~~~~~~~~~~~

.. code-block:: text

   socket.destroy([error])

*Kind:* function · *Arguments:* 1

Ensures that no more I/O activity happens on the current connection. Destroys the stream and closes the connection.

See ``writable.destroy()`` for further details.

**Arguments**

* ``error`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.destroy(new Error('example'));
   console.log(result);


Socket.end
~~~~~~~~~~

.. code-block:: text

   socket.end([data[, encoding]][, callback])

*Kind:* function · *Arguments:* 3

Half-closes the socket. i.e., it sends a FIN packet. It is possible the server will still send some data.

See ``writable.end()`` for further details.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*) — optional
* ``encoding`` (*string*) — optional: Only used when data is ``string``. **Default:** ``'utf8'``.
* ``callback`` (*Function*) — optional: Optional callback for when the socket is finished.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.end('hello', 'utf8', () => {});
   console.log(result);


Socket.pause
~~~~~~~~~~~~

.. code-block:: text

   socket.pause()

*Kind:* function · *Arguments:* 0

Pauses the reading of data. That is, ``'data'`` events will not be emitted. Useful to throttle back an upload.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.pause();
   console.log(result);


Socket.pipe
~~~~~~~~~~~

.. code-block:: text

   net.Socket.pipe(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.pipe(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Socket.ref
~~~~~~~~~~

.. code-block:: text

   socket.ref()

*Kind:* function · *Arguments:* 0

Opposite of ``unref()``, calling ``ref()`` on a previously ``unref``\ ed socket will *not* let the program exit if it's the only socket left (the default behavior). If the socket is ``ref``\ ed calling ``ref`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.ref();
   console.log(result);


Socket.resetAndDestroy
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.resetAndDestroy()

*Kind:* function · *Arguments:* 0

Close the TCP connection by sending an RST packet and destroy the stream. If this TCP socket is in connecting status, it will send an RST packet and destroy this TCP socket once it is connected. Otherwise, it will call ``socket.destroy`` with an ``ERR_SOCKET_CLOSED`` Error. If this is not a TCP socket (for example, a pipe), calling this method will immediately throw an ``ERR_INVALID_HANDLE_TYPE`` Error.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.resetAndDestroy();
   console.log(result);


Socket.resume
~~~~~~~~~~~~~

.. code-block:: text

   socket.resume()

*Kind:* function · *Arguments:* 0

Resumes reading after a call to ``socket.pause()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.resume();
   console.log(result);


Socket.setEncoding
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setEncoding([encoding])

*Kind:* function · *Arguments:* 1

Set the encoding for the socket as a Readable Stream. See ``readable.setEncoding()`` for more information.

**Arguments**

* ``encoding`` (*string*) — optional

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.setEncoding('utf8');
   console.log(result);


Socket.setKeepAlive
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setKeepAlive()

*Kind:* function · *Arguments:* 0

Enable/disable keep-alive functionality, and optionally configure the keepalive probe timing. Returns the socket itself.

Possible signatures:

This callable takes no arguments.

**Example**

.. code-block:: javascript

   socket.setKeepAlive({ enable: true, initialDelay: 1000, interval: 1000, count: 10 });


Socket.setNoDelay
~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setNoDelay([noDelay])

*Kind:* function · *Arguments:* 1

Enable/disable the use of Nagle's algorithm.

When a TCP connection is created, it will have Nagle's algorithm enabled.

**Arguments**

* ``noDelay`` (*boolean*) — optional: **Default:** ``true``

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.setNoDelay(true);
   console.log(result);


Socket.setTimeout
~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setTimeout(timeout[, callback])

*Kind:* function · *Arguments:* 2

Sets the socket to timeout after ``timeout`` milliseconds of inactivity on the socket. By default ``net.Socket`` do not have a timeout.

When an idle timeout is triggered the socket will receive a ``'timeout'`` event but the connection will not be severed. The user must manually call ``socket.end()`` or ``socket.destroy()`` to end the connection.

**Arguments**

* ``timeout`` (*number*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   socket.setTimeout(3000);
   socket.on('timeout', () => {
     console.log('socket timeout');
     socket.end();
   });


Socket.unref
~~~~~~~~~~~~

.. code-block:: text

   socket.unref()

*Kind:* function · *Arguments:* 0

Calling ``unref()`` on a socket will allow the program to exit if this is the only active socket in the event system. If the socket is already ``unref``\ ed calling ``unref()`` again will have no effect.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.unref();
   console.log(result);


Socket.write
~~~~~~~~~~~~

.. code-block:: text

   socket.write(data[, encoding][, callback])

*Kind:* function · *Arguments:* 3

Sends data on the socket. The second parameter specifies the encoding in the case of a string. It defaults to UTF8 encoding.

Returns ``true`` if the entire data was flushed successfully to the kernel buffer. Returns ``false`` if all or part of the data was queued in user memory. ``'drain'`` will be emitted when the buffer is again free.

**Arguments**

* ``data`` (*string|Buffer|Uint8Array*)
* ``encoding`` (*string*) — optional: Only used when data is ``string``. **Default:** ``utf8``.
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.Socket();
   const result = instance.write('hello', 'utf8', () => {});
   console.log(result);


SocketAddress class
-------------------

SocketAddress.toJSON
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   blockList.toJSON()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.SocketAddress();
   const result = instance.toJSON();
   console.log(result);


SocketAddress.toString
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   net.SocketAddress.toString(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const net = require('node:net');

   const instance = new net.SocketAddress();
   const result = instance.toString();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
