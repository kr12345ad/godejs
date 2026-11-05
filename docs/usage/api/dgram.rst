dgram
=====

The ``node:dgram`` module provides an implementation of UDP datagram sockets.

Loading
-------

.. code-block:: javascript

   const dgram = require('node:dgram');
   // ESM
   import * as dgram from 'node:dgram';

Quick start
-----------

.. code-block:: javascript

   const dgram = require('node:dgram');
   const server = dgram.createSocket('udp4');

   server.on('error', (err) => {
     console.error(`server error:\n${err.stack}`);
     server.close();
   });

   server.on('message', (msg, rinfo) => {
     console.log(`server got: ${msg} from ${rinfo.address}:${rinfo.port}`);
   });

   server.on('listening', () => {
     const address = server.address();
     console.log(`server listening ${address.address}:${address.port}`);
   });

   server.bind(41234);
   // Prints: server listening 0.0.0.0:41234

API
---

createSocket
~~~~~~~~~~~~

.. code-block:: text

   dgram.createSocket(options[, callback])

*Kind:* function · *Arguments:* 2

Creates a ``dgram.Socket`` object. Once the socket is created, calling ``socket.bind()`` will instruct the socket to begin listening for datagram messages. When ``address`` and ``port`` are not passed to ``socket.bind()`` the method will bind the socket to the "all interfaces" address on a random port (it does the right thing for both ``udp4`` and ``udp6`` sockets). The bound address and port can be retrieved using ``socket.address().address`` and ``socket.address().port``.

If the ``signal`` option is enabled, calling ``.abort()`` on the corresponding ``AbortController`` is similar to calling ``.close()`` on the socket:

**Arguments**

* ``options`` (*Object*): Available options are:
* ``callback`` (*Function*) — optional: Attached as a listener for ``'message'`` events. Optional.

**Example**

.. code-block:: javascript

   const controller = new AbortController();
   const { signal } = controller;
   const server = dgram.createSocket({ type: 'udp4', signal });
   server.on('message', (msg, rinfo) => {
     console.log(`server got: ${msg} from ${rinfo.address}:${rinfo.port}`);
   });
   // Later, when you want to close the server.
   controller.abort();


Socket
~~~~~~

.. code-block:: text

   dgram.Socket

*Kind:* class · *Arguments:* 2

Encapsulates the datagram functionality.

New instances of ``dgram.Socket`` are created using ``dgram.createSocket()``. The ``new`` keyword is not to be used to create ``dgram.Socket`` instances.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');
   const dgram = require('node:dgram');

   if (cluster.isPrimary) {
     cluster.fork(); // Works ok.
     cluster.fork(); // Fails with EADDRINUSE.
   } else {
     const s = dgram.createSocket('udp4');
     s.bind(1234, () => {
       s.addMembership('224.0.0.114');
     });
   }


Socket class
------------

**Socket** — Encapsulates the datagram functionality.

New instances of ``dgram.Socket`` are created using ``dgram.createSocket()``. The ``new`` keyword is not to be used to create ``dgram.Socket`` instances.

Socket.addMembership
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.addMembership(multicastAddress[, multicastInterface])

*Kind:* function · *Arguments:* 2

Tells the kernel to join a multicast group at the given ``multicastAddress`` and ``multicastInterface`` using the ``IP_ADD_MEMBERSHIP`` socket option. If the ``multicastInterface`` argument is not specified, the operating system will choose one interface and will add membership to it. To add membership to every available interface, call ``addMembership`` multiple times, once per interface.

When called on an unbound socket, this method will implicitly bind to a random port, listening on all interfaces.

**Arguments**

* ``multicastAddress`` (*string*)
* ``multicastInterface`` (*string*) — optional

**Example**

.. code-block:: javascript

   const cluster = require('node:cluster');
   const dgram = require('node:dgram');

   if (cluster.isPrimary) {
     cluster.fork(); // Works ok.
     cluster.fork(); // Fails with EADDRINUSE.
   } else {
     const s = dgram.createSocket('udp4');
     s.bind(1234, () => {
       s.addMembership('224.0.0.114');
     });
   }


Socket.address
~~~~~~~~~~~~~~

.. code-block:: text

   socket.address()

*Kind:* function · *Arguments:* 0

Returns an object containing the address information for a socket. For UDP sockets, this object will contain ``address``, ``family``, and ``port`` properties.

This method throws ``EBADF`` if called on an unbound socket.

This callable takes no arguments.


Socket.bind
~~~~~~~~~~~

.. code-block:: text

   socket.bind([port][, address][, callback])

*Kind:* function · *Arguments:* 3

For UDP sockets, causes the ``dgram.Socket`` to listen for datagram messages on a named ``port`` and optional ``address``. If ``port`` is not specified or is ``0``, the operating system will attempt to bind to a random port. If ``address`` is not specified, the operating system will attempt to listen on all addresses. Once binding is complete, a ``'listening'`` event is emitted and the optional ``callback`` function is called.

Specifying both a ``'listening'`` event listener and passing a ``callback`` to the ``socket.bind()`` method is not harmful but not very useful.

**Arguments**

* ``port`` (*integer*) — optional
* ``address`` (*string*) — optional
* ``callback`` (*Function*) — optional: with no parameters. Called when binding is complete.

**Example**

.. code-block:: javascript

   const dgram = require('node:dgram');
   const server = dgram.createSocket('udp4');

   server.on('error', (err) => {
     console.error(`server error:\n${err.stack}`);
     server.close();
   });

   server.on('message', (msg, rinfo) => {
     console.log(`server got: ${msg} from ${rinfo.address}:${rinfo.port}`);
   });

   server.on('listening', () => {
     const address = server.address();
     console.log(`server listening ${address.address}:${address.port}`);
   });

   server.bind(41234);
   // Prints: server listening 0.0.0.0:41234


Socket.close
~~~~~~~~~~~~

.. code-block:: text

   socket.close([callback])

*Kind:* function · *Arguments:* 1

Close the underlying socket and stop listening for data on it. If a callback is provided, it is added as a listener for the ``'close'`` event.

**Arguments**

* ``callback`` (*Function*) — optional: Called when the socket has been closed.


Socket.connect
~~~~~~~~~~~~~~

.. code-block:: text

   socket.connect(port[, address][, callback])

*Kind:* function · *Arguments:* 3

Associates the ``dgram.Socket`` to a remote address and port. Every message sent by this handle is automatically sent to that destination. Also, the socket will only receive messages from that remote peer. Trying to call ``connect()`` on an already connected socket will result in an ``ERR_SOCKET_DGRAM_IS_CONNECTED`` exception. If ``address`` is not provided, ``'127.0.0.1'`` (for ``udp4`` sockets) or ``'::1'`` (for ``udp6`` sockets) will be used by default. Once the connection is complete, a ``'connect'`` event is emitted and the optional ``callback`` function is called. In case of failure, the ``callback`` is called or, failing this, an ``'error'`` event is emitted.

**Arguments**

* ``port`` (*integer*)
* ``address`` (*string*) — optional
* ``callback`` (*Function*) — optional: Called when the connection is completed or on error.


Socket.disconnect
~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.disconnect()

*Kind:* function · *Arguments:* 0

A synchronous function that disassociates a connected ``dgram.Socket`` from its remote address. Trying to call ``disconnect()`` on an unbound or already disconnected socket will result in an ``ERR_SOCKET_DGRAM_NOT_CONNECTED`` exception.

This callable takes no arguments.


Socket.dropMembership
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.dropMembership(multicastAddress[, multicastInterface])

*Kind:* function · *Arguments:* 2

Instructs the kernel to leave a multicast group at ``multicastAddress`` using the ``IP_DROP_MEMBERSHIP`` socket option. This method is automatically called by the kernel when the socket is closed or the process terminates, so most apps will never have reason to call this.

If ``multicastInterface`` is not specified, the operating system will attempt to drop membership on all valid interfaces.

**Arguments**

* ``multicastAddress`` (*string*)
* ``multicastInterface`` (*string*) — optional


Socket.getRecvBufferSize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.getRecvBufferSize()

*Kind:* function · *Arguments:* 0

This method throws ``ERR_SOCKET_BUFFER_SIZE`` if called on an unbound socket.

This callable takes no arguments.


Socket.getSendBufferSize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.getSendBufferSize()

*Kind:* function · *Arguments:* 0

This method throws ``ERR_SOCKET_BUFFER_SIZE`` if called on an unbound socket.

This callable takes no arguments.


Socket.ref
~~~~~~~~~~

.. code-block:: text

   socket.ref()

*Kind:* function · *Arguments:* 0

By default, binding a socket will cause it to block the Node.js process from exiting as long as the socket is open. The ``socket.unref()`` method can be used to exclude the socket from the reference counting that keeps the Node.js process active. The ``socket.ref()`` method adds the socket back to the reference counting and restores the default behavior.

Calling ``socket.ref()`` multiple times will have no additional effect.

This callable takes no arguments.


Socket.remoteAddress
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.remoteAddress()

*Kind:* function · *Arguments:* 0

Returns an object containing the ``address``, ``family``, and ``port`` of the remote endpoint. This method throws an ``ERR_SOCKET_DGRAM_NOT_CONNECTED`` exception if the socket is not connected.

This callable takes no arguments.


Socket.send
~~~~~~~~~~~

.. code-block:: text

   socket.send(msg[, offset, length][, port][, address][, callback])

*Kind:* function · *Arguments:* 6

Broadcasts a datagram on the socket. For connectionless sockets, the destination ``port`` and ``address`` must be specified. Connected sockets, on the other hand, will use their associated remote endpoint, so the ``port`` and ``address`` arguments must not be set.

The ``msg`` argument contains the message to be sent. Depending on its type, different behavior can apply. If ``msg`` is a ``Buffer``, any ``TypedArray`` or a ``DataView``, the ``offset`` and ``length`` specify the offset within the ``Buffer`` where the message begins and the number of bytes in the message, respectively. If ``msg`` is a ``String``, then it is automatically converted to a ``Buffer`` with ``'utf8'`` encoding. With messages that contain multi-byte characters, ``offset`` and ``length`` will be calculated with respect to byte length and not the character position. If ``msg`` is an array, ``offset`` and ``length`` must not be specified.

**Arguments**

* ``msg`` (*Buffer|TypedArray|DataView|string|Array*): Message to be sent.
* ``offset`` (*integer*) — optional: Offset in the buffer where the message starts.
* ``length`` (*integer*) — optional: Number of bytes in the message.
* ``port`` (*integer*) — optional: Destination port.
* ``address`` (*string*) — optional: Destination host name or IP address.
* ``callback`` (*Function*) — optional: Called when the message has been sent.

**Example**

.. code-block:: javascript

   const dgram = require('node:dgram');
   const { Buffer } = require('node:buffer');

   const message = Buffer.from('Some bytes');
   const client = dgram.createSocket('udp4');
   client.send(message, 41234, 'localhost', (err) => {
     client.close();
   });


Socket.setBroadcast
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setBroadcast(flag)

*Kind:* function · *Arguments:* 1

Sets or clears the ``SO_BROADCAST`` socket option. When set to ``true``, UDP packets may be sent to a local interface's broadcast address.

This method throws ``EBADF`` if called on an unbound socket.

**Arguments**

* ``flag`` (*boolean*)


Socket.setMulticastInterface
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setMulticastInterface(multicastInterface)

*Kind:* function · *Arguments:* 1

\_All references to scope in this section are referring to IPv6 Zone Indexes, which are defined by RFC 4007. In string form, an IP with a scope index is written as ``'IP%scope'`` where scope is an interface name or interface number.\_

Sets the default outgoing multicast interface of the socket to a chosen interface or back to system interface selection. The ``multicastInterface`` must be a valid string representation of an IP from the socket's family.

**Arguments**

* ``multicastInterface`` (*string*)

**Example**

.. code-block:: javascript

   const socket = dgram.createSocket('udp6');

   socket.bind(1234, () => {
     socket.setMulticastInterface('::%eth1');
   });


Socket.setMulticastLoopback
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setMulticastLoopback(flag)

*Kind:* function · *Arguments:* 1

Sets or clears the ``IP_MULTICAST_LOOP`` socket option. When set to ``true``, multicast packets will also be received on the local interface.

This method throws ``EBADF`` if called on an unbound socket.

**Arguments**

* ``flag`` (*boolean*)


Socket.setMulticastTTL
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setMulticastTTL(ttl)

*Kind:* function · *Arguments:* 1

Sets the ``IP_MULTICAST_TTL`` socket option. While TTL generally stands for "Time to Live", in this context it specifies the number of IP hops that a packet is allowed to travel through, specifically for multicast traffic. Each router or gateway that forwards a packet decrements the TTL. If the TTL is decremented to 0 by a router, it will not be forwarded.

The ``ttl`` argument may be between 0 and 255. The default on most systems is ``1``.

**Arguments**

* ``ttl`` (*integer*)


Socket.setRecvBufferSize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setRecvBufferSize(size)

*Kind:* function · *Arguments:* 1

Sets the ``SO_RCVBUF`` socket option. Sets the maximum socket receive buffer in bytes.

This method throws ``ERR_SOCKET_BUFFER_SIZE`` if called on an unbound socket.

**Arguments**

* ``size`` (*integer*)


Socket.setSendBufferSize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   socket.setSendBufferSize(size)

*Kind:* function · *Arguments:* 1

Sets the ``SO_SNDBUF`` socket option. Sets the maximum socket send buffer in bytes.

This method throws ``ERR_SOCKET_BUFFER_SIZE`` if called on an unbound socket.

**Arguments**

* ``size`` (*integer*)


Socket.setTTL
~~~~~~~~~~~~~

.. code-block:: text

   socket.setTTL(ttl)

*Kind:* function · *Arguments:* 1

Sets the ``IP_TTL`` socket option. While TTL generally stands for "Time to Live", in this context it specifies the number of IP hops that a packet is allowed to travel through. Each router or gateway that forwards a packet decrements the TTL. If the TTL is decremented to 0 by a router, it will not be forwarded. Changing TTL values is typically done for network probes or when multicasting.

The ``ttl`` argument may be between 1 and 255. The default on most systems is 64.

**Arguments**

* ``ttl`` (*integer*)


Socket.unref
~~~~~~~~~~~~

.. code-block:: text

   socket.unref()

*Kind:* function · *Arguments:* 0

By default, binding a socket will cause it to block the Node.js process from exiting as long as the socket is open. The ``socket.unref()`` method can be used to exclude the socket from the reference counting that keeps the Node.js process active, allowing the process to exit even if the socket is still listening.

Calling ``socket.unref()`` multiple times will have no additional effect.

This callable takes no arguments.


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
