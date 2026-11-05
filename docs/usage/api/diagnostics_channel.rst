diagnostics_channel
===================

The ``node:diagnostics_channel`` module provides an API to create named channels to report arbitrary message data for diagnostics purposes.

It can be accessed using:

Loading
-------

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');
   // ESM
   import * as diagnostics_channel from 'node:diagnostics_channel';

Quick start
-----------

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

API
---

channel
~~~~~~~

.. code-block:: text

   diagnostics_channel.channel(name)

*Kind:* function · *Arguments:* 1

This is the primary entry-point for anyone wanting to publish to a named channel. It produces a channel object which is optimized to reduce overhead at publish time as much as possible.

**Arguments**

* ``name`` (*string|symbol*): The channel name

**Example**

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

   const channel = diagnostics_channel.channel('my-channel');


hasSubscribers
~~~~~~~~~~~~~~

.. code-block:: text

   diagnostics_channel.hasSubscribers(name)

*Kind:* function · *Arguments:* 1

Check if there are active subscribers to the named channel. This is helpful if the message you want to send might be expensive to prepare.

This API is optional but helpful when trying to publish messages from very performance-sensitive code.

**Arguments**

* ``name`` (*string|symbol*): The channel name

**Example**

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

   if (diagnostics_channel.hasSubscribers('my-channel')) {
     // There are subscribers, prepare and publish message
   }


subscribe
~~~~~~~~~

.. code-block:: text

   diagnostics_channel.subscribe(name, onMessage)

*Kind:* function · *Arguments:* 2

Register a message handler to subscribe to this channel. This message handler will be run synchronously whenever a message is published to the channel. Any errors thrown in the message handler will trigger an ``'uncaughtException'``.

**Arguments**

* ``name`` (*string|symbol*): The channel name
* ``onMessage`` (*Function*): The handler to receive channel messages

**Example**

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

   diagnostics_channel.subscribe('my-channel', (message, name) => {
     // Received data
   });


tracingChannel
~~~~~~~~~~~~~~

.. code-block:: text

   diagnostics_channel.tracingChannel(nameOrChannels)

*Kind:* function · *Arguments:* 1

Creates a ``TracingChannel`` wrapper for the given TracingChannel Channels. If a name is given, the corresponding tracing channels will be created in the form of ``tracing:${name}:${eventType}`` where ``eventType`` corresponds to the types of TracingChannel Channels.

**Arguments**

* ``nameOrChannels`` (*string|TracingChannel*): Channel name or

**Example**

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

   const channelsByName = diagnostics_channel.tracingChannel('my-channel');

   // or...

   const channelsByCollection = diagnostics_channel.tracingChannel({
     start: diagnostics_channel.channel('tracing:my-channel:start'),
     end: diagnostics_channel.channel('tracing:my-channel:end'),
     asyncStart: diagnostics_channel.channel('tracing:my-channel:asyncStart'),
     asyncEnd: diagnostics_channel.channel('tracing:my-channel:asyncEnd'),
     error: diagnostics_channel.channel('tracing:my-channel:error'),
   });


unsubscribe
~~~~~~~~~~~

.. code-block:: text

   diagnostics_channel.unsubscribe(name, onMessage)

*Kind:* function · *Arguments:* 2

Remove a message handler previously registered to this channel with ``diagnostics_channel.subscribe(name, onMessage)``.

**Arguments**

* ``name`` (*string|symbol*): The channel name
* ``onMessage`` (*Function*): The previous subscribed handler to remove

**Example**

.. code-block:: javascript

   const diagnostics_channel = require('node:diagnostics_channel');

   function onMessage(message, name) {
     // Received data
   }

   diagnostics_channel.subscribe('my-channel', onMessage);

   diagnostics_channel.unsubscribe('my-channel', onMessage);


.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
