inspector
=========

The ``node:inspector`` module provides an API for interacting with the V8 inspector.

It can be accessed using:

Loading
-------

.. code-block:: javascript

   const inspector = require('node:inspector');
   // ESM
   import * as inspector from 'node:inspector';

Quick start
-----------

.. code-block:: javascript

   const inspector = require('node:inspector/promises');

API
---

close
~~~~~

.. code-block:: text

   inspector.close()

*Kind:* function · *Arguments:* 0

Deactivates the inspector. If there are active connections, they are forcibly terminated. Blocks until the inspector server has fully stopped.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#inspectorclose>`__

isOpen
~~~~~~

.. code-block:: text

   database.isOpen

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/sqlite.html#databaseisopen>`__

open
~~~~

.. code-block:: text

   inspector.open([port[, host[, wait]]])

*Kind:* function · *Arguments:* 3

Activate inspector on host and port. Equivalent to ``node --inspect=[[host:]port]``, but can be done programmatically after node has started.

If wait is ``true``, will block until a client has connected to the inspect port and flow control has been passed to the debugger client.

**Arguments**

* ``port`` (*number*) — optional: Port to listen on for inspector connections. Optional.
* ``host`` (*string*) — optional: Host to listen on for inspector connections. Optional.
* ``wait`` (*boolean*) — optional: Block until a client has connected. Optional.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#inspectoropenport-host-wait>`__

Session
~~~~~~~

.. code-block:: text

   inspector.Session

*Kind:* class · *Arguments:* 0

The ``inspector.Session`` is used for dispatching messages to the V8 inspector back-end and receiving message responses and notifications.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   session.on('inspectorNotification', (message) => console.log(message.method));
   // Debugger.paused
   // Debugger.resumed

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#class-inspectorsession>`__

url
~~~

.. code-block:: text

   inspector.url()

*Kind:* function · *Arguments:* 0

Return the URL of the active inspector, or ``undefined`` if there is none.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#inspectorurl>`__

waitForDebugger
~~~~~~~~~~~~~~~

.. code-block:: text

   inspector.waitForDebugger()

*Kind:* function · *Arguments:* 0

Blocks until a client (existing or connected later) has sent ``Runtime.runIfWaitingForDebugger`` command.

An exception will be thrown if there is no active inspector.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#inspectorwaitfordebugger>`__

Session class
-------------

**Session** — The ``inspector.Session`` is used for dispatching messages to the V8 inspector back-end and receiving message responses and notifications.

Session.connect
~~~~~~~~~~~~~~~

.. code-block:: text

   session.connect()

*Kind:* function · *Arguments:* 0

Connects a session to the inspector back-end.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#sessionconnect>`__

Session.connectToMainThread
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   session.connectToMainThread()

*Kind:* function · *Arguments:* 0

Connects a session to the main thread inspector back-end. An exception will be thrown if this API was not called on a Worker thread.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#sessionconnecttomainthread>`__

Session.disconnect
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   session.disconnect()

*Kind:* function · *Arguments:* 0

Immediately close the session. All pending message callbacks will be called with an error. ``session.connect()`` will need to be called to be able to send messages again. Reconnected session will lose all inspector state, such as enabled agents or configured breakpoints.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#sessiondisconnect>`__

Session.post
~~~~~~~~~~~~

.. code-block:: text

   session.post(method[, params][, callback])

*Kind:* function · *Arguments:* 3

Posts a message to the inspector back-end. ``callback`` will be notified when a response is received. ``callback`` is a function that accepts two optional arguments: error and message-specific result.

The latest version of the V8 inspector protocol is published on the Chrome DevTools Protocol Viewer.

**Arguments**

* ``method`` (*string*)
* ``params`` (*Object*) — optional
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   session.post('Runtime.evaluate', { expression: '2 + 2' },
                (error, { result }) => console.log(result));
   // Output: { type: 'number', value: 4, description: '4' }

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/inspector.html#sessionpostmethod-params-callback>`__

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
