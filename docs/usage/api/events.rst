events
======

Much of the Node.js core API is built around an idiomatic asynchronous event-driven architecture in which certain kinds of objects (called "emitters") emit named events that cause ``Function`` objects ("listeners") to be called.

For instance: a ``net.Server`` object emits an event each time a peer connects to it; a ``fs.ReadStream`` emits an event when the file is opened; a stream emits an event whenever data is available to be read.

Loading
-------

.. code-block:: javascript

   const events = require('node:events');
   // ESM
   import * as events from 'node:events';

Quick start
-----------

.. code-block:: javascript

   const EventEmitter = require('node:events');

   class MyEmitter extends EventEmitter {}

   const myEmitter = new MyEmitter();
   myEmitter.on('event', () => {
     console.log('an event occurred!');
   });
   myEmitter.emit('event');

API
---

addAbortListener
~~~~~~~~~~~~~~~~

.. code-block:: text

   events.addAbortListener(signal, listener)

*Kind:* function · *Arguments:* 2

Listens once to the ``abort`` event on the provided ``signal``.

Listening to the ``abort`` event on abort signals is unsafe and may lead to resource leaks since another third party with the signal can call ``e.stopImmediatePropagation()``. Unfortunately Node.js cannot change this since it would violate the web standard. Additionally, the original API makes it easy to forget to remove listeners.

**Arguments**

* ``signal`` (*AbortSignal*)
* ``listener`` (*Function|EventListener*)

**Example**

.. code-block:: javascript

   const { addAbortListener } = require('node:events');

   function example(signal) {
     signal.addEventListener('abort', (e) => e.stopImmediatePropagation());
     // addAbortListener() returns a disposable, so the `using` keyword ensures
     // the abort listener is automatically removed when this scope exits.
     using _ = addAbortListener(signal, (e) => {
       // Do something when signal is aborted.
     });
   }


CustomEvent
~~~~~~~~~~~

.. code-block:: text

   CustomEvent

*Kind:* class · *Arguments:* 1

The ``CustomEvent`` object is an adaptation of the ``CustomEvent`` Web API. Instances are created internally by Node.js.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.CustomEvent(undefined);
   console.log(instance);


Event
~~~~~

.. code-block:: text

   Event

*Kind:* class · *Arguments:* 2

The ``Event`` object is an adaptation of the ``Event`` Web API. Instances are created internally by Node.js.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event(undefined, undefined);
   console.log(instance);


EventEmitter
~~~~~~~~~~~~

.. code-block:: text

   EventEmitter

*Kind:* class · *Arguments:* 1

The ``EventEmitter`` class is defined and exposed by the ``node:events`` module:

All ``EventEmitter``\ s emit the event ``'newListener'`` when new listeners are added and ``'removeListener'`` when existing listeners are removed.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');


EventEmitterAsyncResource
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   events.EventEmitterAsyncResource([options])

*Kind:* class · *Arguments:* 1

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitterAsyncResource({});
   console.log(instance);


EventTarget
~~~~~~~~~~~

.. code-block:: text

   EventTarget

*Kind:* class · *Arguments:* 0

The ``EventTarget`` and ``Event`` objects are a Node.js-specific implementation of the ``EventTarget`` Web API that are exposed by some Node.js core APIs.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const target = new EventTarget();

   target.addEventListener('foo', (event) => {
     console.log('foo event happened!');
   });


getEventListeners
~~~~~~~~~~~~~~~~~

.. code-block:: text

   events.getEventListeners(emitterOrTarget, eventName)

*Kind:* function · *Arguments:* 2

Returns a copy of the array of listeners for the event named ``eventName``.

For ``EventEmitter``\ s this behaves exactly the same as calling ``.listeners`` on the emitter.

**Arguments**

* ``emitterOrTarget`` (*EventEmitter|EventTarget*)
* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   const { getEventListeners, EventEmitter } = require('node:events');

   {
     const ee = new EventEmitter();
     const listener = () => console.log('Events are fun');
     ee.on('foo', listener);
     console.log(getEventListeners(ee, 'foo')); // [ [Function: listener] ]
   }
   {
     const et = new EventTarget();
     const listener = () => console.log('Events are fun');
     et.addEventListener('foo', listener);
     console.log(getEventListeners(et, 'foo')); // [ [Function: listener] ]
   }


getMaxListeners
~~~~~~~~~~~~~~~

.. code-block:: text

   events.getMaxListeners(emitterOrTarget)

*Kind:* function · *Arguments:* 1

Returns the currently set max amount of listeners.

For ``EventEmitter``\ s this behaves exactly the same as calling ``.getMaxListeners`` on the emitter.

**Arguments**

* ``emitterOrTarget`` (*EventEmitter|EventTarget*)

**Example**

.. code-block:: javascript

   const { getMaxListeners, setMaxListeners, EventEmitter } = require('node:events');

   {
     const ee = new EventEmitter();
     console.log(getMaxListeners(ee)); // 10
     setMaxListeners(11, ee);
     console.log(getMaxListeners(ee)); // 11
   }
   {
     const et = new EventTarget();
     console.log(getMaxListeners(et)); // 10
     setMaxListeners(11, et);
     console.log(getMaxListeners(et)); // 11
   }


listenerCount
~~~~~~~~~~~~~

.. code-block:: text

   events.listenerCount(emitterOrTarget, eventName)

*Kind:* function · *Arguments:* 2

Returns the number of registered listeners for the event named ``eventName``.

For ``EventEmitter``\ s this behaves exactly the same as calling ``.listenerCount`` on the emitter.

**Arguments**

* ``emitterOrTarget`` (*EventEmitter|EventTarget*)
* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   const { EventEmitter, listenerCount } = require('node:events');

   {
     const ee = new EventEmitter();
     ee.on('event', () => {});
     ee.on('event', () => {});
     console.log(listenerCount(ee, 'event')); // 2
   }
   {
     const et = new EventTarget();
     et.addEventListener('event', () => {});
     et.addEventListener('event', () => {});
     console.log(listenerCount(et, 'event')); // 2
   }


NodeEventTarget
~~~~~~~~~~~~~~~

.. code-block:: text

   NodeEventTarget

*Kind:* class · *Arguments:* 0

The ``NodeEventTarget`` object implements a modified subset of the ``EventEmitter`` API that allows it to closely *emulate* an ``EventEmitter`` in certain situations. A ``NodeEventTarget`` is *not* an instance of ``EventEmitter`` and cannot be used in place of an ``EventEmitter`` in most cases.

1. Unlike ``EventEmitter``, any given ``listener`` can be registered at most once per event ``type``. Attempts to register a ``listener`` multiple times are ignored. 2. The ``NodeEventTarget`` does not emulate the full ``EventEmitter`` API. Specifically the ``prependListener()``, ``prependOnceListener()``, ``rawListeners()``, and ``errorMonitor`` APIs are not emulated. The ``'newListener'`` and ``'removeListener'`` events will also not be emitted. 3. The ``NodeEventTarget`` does not implement any special default behavior for events with type ``'error'``. 4. The ``NodeEventTarget`` supports ``EventListener`` objects as well as functions as handlers for all event types.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   console.log(instance);


on
~~~

.. code-block:: text

   events.on(emitter, eventName[, options])

*Kind:* function · *Arguments:* 3

Returns an ``AsyncIterator`` that iterates ``eventName`` events. It will throw if the ``EventEmitter`` emits ``'error'``. It removes all listeners when exiting the loop. The ``value`` returned by each iteration is an array composed of the emitted event arguments.

An ``AbortSignal`` can be used to cancel waiting on events:

**Arguments**

* ``emitter`` (*EventEmitter*)
* ``eventName`` (*string|symbol*): The name of the event being listened for
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { on, EventEmitter } = require('node:events');

   (async () => {
     const ee = new EventEmitter();

     // Emit later on
     process.nextTick(() => {
       ee.emit('foo', 'bar');
       ee.emit('foo', 42);
     });

     for await (const event of on(ee, 'foo')) {
       // The execution of this inner block is synchronous and it
       // processes one event at a time (even with await). Do not use
       // if concurrent execution is required.
       console.log(event); // prints ['bar'] [42]
     }
     // Unreachable here
   })();


once
~~~~

.. code-block:: text

   events.once(emitter, name[, options])

*Kind:* function · *Arguments:* 3

Creates a ``Promise`` that is fulfilled when the ``EventEmitter`` emits the given event or that is rejected if the ``EventEmitter`` emits ``'error'`` while waiting. The ``Promise`` will resolve with an array of all the arguments emitted to the given event.

This method is intentionally generic and works with the web platform EventTarget interface, which has no special ``'error'`` event semantics and does not listen to the ``'error'`` event.

**Arguments**

* ``emitter`` (*EventEmitter*)
* ``name`` (*string|symbol*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { once, EventEmitter } = require('node:events');

   async function run() {
     const ee = new EventEmitter();

     process.nextTick(() => {
       ee.emit('myevent', 42);
     });

     const [value] = await once(ee, 'myevent');
     console.log(value);

     const err = new Error('kaboom');
     process.nextTick(() => {
       ee.emit('error', err);
     });

     try {
       await once(ee, 'myevent');
     } catch (err) {
       console.error('error happened', err);
     }
   }

   run();


setMaxListeners
~~~~~~~~~~~~~~~

.. code-block:: text

   events.setMaxListeners(n[, ...eventTargets])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``n`` (*number*): A non-negative number. The maximum number of listeners per
* ``...eventTargets`` — optional

**Example**

.. code-block:: javascript

   const {
     setMaxListeners,
     EventEmitter,
   } = require('node:events');

   const target = new EventTarget();
   const emitter = new EventEmitter();

   setMaxListeners(5, target, emitter);


Event class
-----------

**Event** — The ``Event`` object is an adaptation of the ``Event`` Web API. Instances are created internally by Node.js.

Event.composedPath
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.composedPath()

*Kind:* function · *Arguments:* 0

Returns an array containing the current ``EventTarget`` as the only entry or empty if the event is not being dispatched. This is not used in Node.js and is provided purely for completeness.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event();
   const result = instance.composedPath();
   console.log(result);


Event.initEvent
~~~~~~~~~~~~~~~

.. code-block:: text

   event.initEvent(type[, bubbles[, cancelable]])

*Kind:* function · *Arguments:* 3

Redundant with event constructors and incapable of setting ``composed``. This is not used in Node.js and is provided purely for completeness.

**Arguments**

* ``type`` (*string*)
* ``bubbles`` (*boolean*) — optional
* ``cancelable`` (*boolean*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event();
   const result = instance.initEvent('example', true, true);
   console.log(result);


Event.preventDefault
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.preventDefault()

*Kind:* function · *Arguments:* 0

Sets the ``defaultPrevented`` property to ``true`` if ``cancelable`` is ``true``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event();
   const result = instance.preventDefault();
   console.log(result);


Event.stopImmediatePropagation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.stopImmediatePropagation()

*Kind:* function · *Arguments:* 0

Stops the invocation of event listeners after the current one completes.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event();
   const result = instance.stopImmediatePropagation();
   console.log(result);


Event.stopPropagation
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   event.stopPropagation()

*Kind:* function · *Arguments:* 0

This is not used in Node.js and is provided purely for completeness.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.Event();
   const result = instance.stopPropagation();
   console.log(result);


EventEmitter class
------------------

**EventEmitter** — The ``EventEmitter`` class is defined and exposed by the ``node:events`` module:

All ``EventEmitter``\ s emit the event ``'newListener'`` when new listeners are added and ``'removeListener'`` when existing listeners are removed.

EventEmitter._godeHasAnyListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   events.EventEmitter._godeHasAnyListeners(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const events = require('node:events');

   console.log(typeof events.EventEmitter.prototype._godeHasAnyListeners()); // 'function'
   console.log(events.EventEmitter.prototype._godeHasAnyListeners().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


EventEmitter.addListener
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.addListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Alias for ``emitter.on(eventName, listener)``.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.addListener('value', () => {});
   console.log(result);


EventEmitter.emit
~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.emit(eventName[, ...args])

*Kind:* function · *Arguments:* 2

Synchronously calls each of the listeners registered for the event named ``eventName``, in the order they were registered, passing the supplied arguments to each.

Returns ``true`` if the event had listeners, ``false`` otherwise.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``...args`` — optional

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const myEmitter = new EventEmitter();

   // First listener
   myEmitter.on('event', function firstListener() {
     console.log('Helloooo! first listener');
   });
   // Second listener
   myEmitter.on('event', function secondListener(arg1, arg2) {
     console.log(`event with parameters ${arg1}, ${arg2} in second listener`);
   });
   // Third listener
   myEmitter.on('event', function thirdListener(...args) {
     const parameters = args.join(', ');
     console.log(`event with parameters ${parameters} in third listener`);
   });

   console.log(myEmitter.listeners('event'));

   myEmitter.emit('event', 1, 2, 3, 4, 5);

   // Prints:
   // [
   //   [Function: firstListener],
   //   [Function: secondListener],
   //   [Function: thirdListener]
   // ]
   // Helloooo! first listener
   // event with parameters 1, 2 in second listener
   // event with parameters 1, 2, 3, 4, 5 in third listener


EventEmitter.eventNames
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.eventNames()

*Kind:* function · *Arguments:* 0

Returns an array listing the events for which the emitter has registered listeners.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');

   const myEE = new EventEmitter();
   myEE.on('foo', () => {});
   myEE.on('bar', () => {});

   const sym = Symbol('symbol');
   myEE.on(sym, () => {});

   console.log(myEE.eventNames());
   // Prints: [ 'foo', 'bar', Symbol(symbol) ]


EventEmitter.getMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.getMaxListeners()

*Kind:* function · *Arguments:* 0

Returns the current max listener value for the ``EventEmitter`` which is either set by ``emitter.setMaxListeners(n)`` or defaults to ``events.defaultMaxListeners``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.getMaxListeners();
   console.log(result);


EventEmitter.listenerCount
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   EventEmitter.listenerCount()

*Kind:* function · *Arguments:* 0

Type: Revoked

The ``events.listenerCount(emitter, eventName)`` API was deprecated, as it provided identical functionality to ``emitter.listenerCount(eventName)``. The deprecation was revoked because this function has been repurposed to also accept ``EventTarget`` arguments.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.listenerCount();
   console.log(result);


EventEmitter.listeners
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.listeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``.

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   server.on('connection', (stream) => {
     console.log('someone connected!');
   });
   console.log(util.inspect(server.listeners('connection')));
   // Prints: [ [Function] ]


EventEmitter.off
~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.off(eventName, listener)

*Kind:* function · *Arguments:* 2

Alias for ``emitter.removeListener()``.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.off('value', () => {});
   console.log(result);


EventEmitter.on
~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.on(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds the ``listener`` function to the end of the listeners array for the event named ``eventName``. No checks are made to see if the ``listener`` has already been added. Multiple calls passing the same combination of ``eventName`` and ``listener`` will result in the ``listener`` being added, and called, multiple times.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const myEE = new EventEmitter();
   myEE.on('foo', () => console.log('a'));
   myEE.prependListener('foo', () => console.log('b'));
   myEE.emit('foo');
   // Prints:
   //   b
   //   a


EventEmitter.once
~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.once(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds a **one-time** ``listener`` function for the event named ``eventName``. The next time ``eventName`` is triggered, this listener is removed and then invoked.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const myEE = new EventEmitter();
   myEE.once('foo', () => console.log('a'));
   myEE.prependOnceListener('foo', () => console.log('b'));
   myEE.emit('foo');
   // Prints:
   //   b
   //   a


EventEmitter.prependListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds the ``listener`` function to the *beginning* of the listeners array for the event named ``eventName``. No checks are made to see if the ``listener`` has already been added. Multiple calls passing the same combination of ``eventName`` and ``listener`` will result in the ``listener`` being added, and called, multiple times.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependListener('connection', (stream) => {
     console.log('someone connected!');
   });


EventEmitter.prependOnceListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependOnceListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds a **one-time** ``listener`` function for the event named ``eventName`` to the *beginning* of the listeners array. The next time ``eventName`` is triggered, this listener is removed, and then invoked.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependOnceListener('connection', (stream) => {
     console.log('Ah, we have our first user!');
   });


EventEmitter.rawListeners
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.rawListeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``, including any wrappers (such as those created by ``.once()``).

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const emitter = new EventEmitter();
   emitter.once('log', () => console.log('log once'));

   // Returns a new Array with a function `onceWrapper` which has a property
   // `listener` which contains the original listener bound above
   const listeners = emitter.rawListeners('log');
   const logFnWrapper = listeners[0];

   // Logs "log once" to the console and does not unbind the `once` event
   logFnWrapper.listener();

   // Logs "log once" to the console and removes the listener
   logFnWrapper();

   emitter.on('log', () => console.log('log persistently'));
   // Will return a new Array with a single function bound by `.on()` above
   const newListeners = emitter.rawListeners('log');

   // Logs "log persistently" twice
   newListeners[0]();
   emitter.emit('log');


EventEmitter.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.removeAllListeners([eventName])

*Kind:* function · *Arguments:* 1

Removes all listeners, or those of the specified ``eventName``.

It is bad practice to remove listeners added elsewhere in the code, particularly when the ``EventEmitter`` instance was created by some other component or module (e.g. sockets or file streams).

**Arguments**

* ``eventName`` (*string|symbol*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.removeAllListeners('value');
   console.log(result);


EventEmitter.removeListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.removeListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Removes the specified ``listener`` from the listener array for the event named ``eventName``.

``removeListener()`` will remove, at most, one instance of a listener from the listener array. If any single listener has been added multiple times to the listener array for the specified ``eventName``, then ``removeListener()`` must be called multiple times to remove each instance.

**Arguments**

* ``eventName`` (*string|symbol*)
* ``listener`` (*Function*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   class MyEmitter extends EventEmitter {}
   const myEmitter = new MyEmitter();

   const callbackA = () => {
     console.log('A');
     myEmitter.removeListener('event', callbackB);
   };

   const callbackB = () => {
     console.log('B');
   };

   myEmitter.on('event', callbackA);

   myEmitter.on('event', callbackB);

   // callbackA removes listener callbackB but it will still be called.
   // Internal listener array at time of emit [callbackA, callbackB]
   myEmitter.emit('event');
   // Prints:
   //   A
   //   B

   // callbackB is now removed.
   // Internal listener array [callbackA]
   myEmitter.emit('event');
   // Prints:
   //   A


EventEmitter.setMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.setMaxListeners(n)

*Kind:* function · *Arguments:* 1

By default ``EventEmitter``\ s will print a warning if more than ``10`` listeners are added for a particular event. This is a useful default that helps finding memory leaks. The ``emitter.setMaxListeners()`` method allows the limit to be modified for this specific ``EventEmitter`` instance. The value can be set to ``Infinity`` (or ``0``) to indicate an unlimited number of listeners.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``n`` (*integer*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitter();
   const result = instance.setMaxListeners(1);
   console.log(result);


EventEmitterAsyncResource class
-------------------------------

EventEmitterAsyncResource.emitDestroy
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventemitterasyncresource.emitDestroy()

*Kind:* function · *Arguments:* 0

Call all ``destroy`` hooks. This should only ever be called once. An error will be thrown if it is called more than once. This **must** be manually called. If the resource is left to be collected by the GC then the ``destroy`` hooks will never be called.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventEmitterAsyncResource();
   const result = instance.emitDestroy();
   console.log(result);


EventTarget class
-----------------

**EventTarget** — The ``EventTarget`` and ``Event`` objects are a Node.js-specific implementation of the ``EventTarget`` Web API that are exposed by some Node.js core APIs.

EventTarget.addEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.addEventListener(type, listener[, options])

*Kind:* function · *Arguments:* 3

Adds a new handler for the ``type`` event. Any given ``listener`` is added only once per ``type`` and per ``capture`` option value.

If the ``once`` option is ``true``, the ``listener`` is removed after the next time a ``type`` event is dispatched.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   function handler(event) {}

   const target = new EventTarget();
   target.addEventListener('foo', handler, { capture: true });  // first
   target.addEventListener('foo', handler, { capture: false }); // second

   // Removes the second instance of handler
   target.removeEventListener('foo', handler);

   // Removes the first instance of handler
   target.removeEventListener('foo', handler, { capture: true });


EventTarget.dispatchEvent
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.dispatchEvent(event)

*Kind:* function · *Arguments:* 1

Dispatches the ``event`` to the list of handlers for ``event.type``.

The registered event listeners is synchronously invoked in the order they were registered.

**Arguments**

* ``event`` (*Event*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventTarget();
   const result = instance.dispatchEvent(undefined);
   console.log(result);


EventTarget.removeEventListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   eventTarget.removeEventListener(type, listener[, options])

*Kind:* function · *Arguments:* 3

Removes the ``listener`` from the list of handlers for event ``type``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.EventTarget();
   const result = instance.removeEventListener('example', () => {}, {});
   console.log(result);


NodeEventTarget class
---------------------

**NodeEventTarget** — The ``NodeEventTarget`` object implements a modified subset of the ``EventEmitter`` API that allows it to closely *emulate* an ``EventEmitter`` in certain situations. A ``NodeEventTarget`` is *not* an instance of ``EventEmitter`` and cannot be used in place of an ``EventEmitter`` in most cases.

1. Unlike ``EventEmitter``, any given ``listener`` can be registered at most once per event ``type``. Attempts to register a ``listener`` multiple times are ignored. 2. The ``NodeEventTarget`` does not emulate the full ``EventEmitter`` API. Specifically the ``prependListener()``, ``prependOnceListener()``, ``rawListeners()``, and ``errorMonitor`` APIs are not emulated. The ``'newListener'`` and ``'removeListener'`` events will also not be emitted. 3. The ``NodeEventTarget`` does not implement any special default behavior for events with type ``'error'``. 4. The ``NodeEventTarget`` supports ``EventListener`` objects as well as functions as handlers for all event types.

NodeEventTarget._godeHasAnyListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   events.NodeEventTarget._godeHasAnyListeners(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const events = require('node:events');

   console.log(typeof events.NodeEventTarget.prototype._godeHasAnyListeners()); // 'function'
   console.log(events.NodeEventTarget.prototype._godeHasAnyListeners().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


NodeEventTarget.addListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.addListener(type, listener)

*Kind:* function · *Arguments:* 2

Node.js-specific extension to the ``EventTarget`` class that emulates the equivalent ``EventEmitter`` API. The only difference between ``addListener()`` and ``addEventListener()`` is that ``addListener()`` will return a reference to the ``EventTarget``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.addListener('example', () => {});
   console.log(result);


NodeEventTarget.emit
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.emit(type, arg)

*Kind:* function · *Arguments:* 2

Node.js-specific extension to the ``EventTarget`` class that dispatches the ``arg`` to the list of handlers for ``type``.

**Arguments**

* ``type`` (*string*)
* ``arg`` (*any*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.emit('example', undefined);
   console.log(result);


NodeEventTarget.eventNames
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.eventNames()

*Kind:* function · *Arguments:* 0

Node.js-specific extension to the ``EventTarget`` class that returns an array of event ``type`` names for which event listeners are registered.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.eventNames();
   console.log(result);


NodeEventTarget.getMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.getMaxListeners()

*Kind:* function · *Arguments:* 0

Node.js-specific extension to the ``EventTarget`` class that returns the number of max event listeners.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.getMaxListeners();
   console.log(result);


NodeEventTarget.listenerCount
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.listenerCount(type)

*Kind:* function · *Arguments:* 1

Node.js-specific extension to the ``EventTarget`` class that returns the number of event listeners registered for the ``type``.

**Arguments**

* ``type`` (*string*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.listenerCount('example');
   console.log(result);


NodeEventTarget.listeners
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.listeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``.

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   server.on('connection', (stream) => {
     console.log('someone connected!');
   });
   console.log(util.inspect(server.listeners('connection')));
   // Prints: [ [Function] ]


NodeEventTarget.off
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.off(type, listener[, options])

*Kind:* function · *Arguments:* 3

Node.js-specific alias for ``eventTarget.removeEventListener()``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.off('example', () => {}, {});
   console.log(result);


NodeEventTarget.on
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.on(type, listener)

*Kind:* function · *Arguments:* 2

Node.js-specific alias for ``eventTarget.addEventListener()``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.on('example', () => {});
   console.log(result);


NodeEventTarget.once
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.once(type, listener)

*Kind:* function · *Arguments:* 2

Node.js-specific extension to the ``EventTarget`` class that adds a ``once`` listener for the given event ``type``. This is equivalent to calling ``on`` with the ``once`` option set to ``true``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.once('example', () => {});
   console.log(result);


NodeEventTarget.prependListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds the ``listener`` function to the *beginning* of the listeners array for the event named ``eventName``. No checks are made to see if the ``listener`` has already been added. Multiple calls passing the same combination of ``eventName`` and ``listener`` will result in the ``listener`` being added, and called, multiple times.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependListener('connection', (stream) => {
     console.log('someone connected!');
   });


NodeEventTarget.prependOnceListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.prependOnceListener(eventName, listener)

*Kind:* function · *Arguments:* 2

Adds a **one-time** ``listener`` function for the event named ``eventName`` to the *beginning* of the listeners array. The next time ``eventName`` is triggered, this listener is removed, and then invoked.

Returns a reference to the ``EventEmitter``, so that calls can be chained.

**Arguments**

* ``eventName`` (*string|symbol*): The name of the event.
* ``listener`` (*Function*): The callback function

**Example**

.. code-block:: javascript

   server.prependOnceListener('connection', (stream) => {
     console.log('Ah, we have our first user!');
   });


NodeEventTarget.rawListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   emitter.rawListeners(eventName)

*Kind:* function · *Arguments:* 1

Returns a copy of the array of listeners for the event named ``eventName``, including any wrappers (such as those created by ``.once()``).

**Arguments**

* ``eventName`` (*string|symbol*)

**Example**

.. code-block:: javascript

   const EventEmitter = require('node:events');
   const emitter = new EventEmitter();
   emitter.once('log', () => console.log('log once'));

   // Returns a new Array with a function `onceWrapper` which has a property
   // `listener` which contains the original listener bound above
   const listeners = emitter.rawListeners('log');
   const logFnWrapper = listeners[0];

   // Logs "log once" to the console and does not unbind the `once` event
   logFnWrapper.listener();

   // Logs "log once" to the console and removes the listener
   logFnWrapper();

   emitter.on('log', () => console.log('log persistently'));
   // Will return a new Array with a single function bound by `.on()` above
   const newListeners = emitter.rawListeners('log');

   // Logs "log persistently" twice
   newListeners[0]();
   emitter.emit('log');


NodeEventTarget.removeAllListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.removeAllListeners([type])

*Kind:* function · *Arguments:* 1

Node.js-specific extension to the ``EventTarget`` class. If ``type`` is specified, removes all registered listeners for ``type``, otherwise removes all registered listeners.

**Arguments**

* ``type`` (*string*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.removeAllListeners('example');
   console.log(result);


NodeEventTarget.removeListener
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.removeListener(type, listener[, options])

*Kind:* function · *Arguments:* 3

Node.js-specific extension to the ``EventTarget`` class that removes the ``listener`` for the given ``type``. The only difference between ``removeListener()`` and ``removeEventListener()`` is that ``removeListener()`` will return a reference to the ``EventTarget``.

**Arguments**

* ``type`` (*string*)
* ``listener`` (*Function|EventListener*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.removeListener('example', () => {}, {});
   console.log(result);


NodeEventTarget.setMaxListeners
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   nodeEventTarget.setMaxListeners(n)

*Kind:* function · *Arguments:* 1

Node.js-specific extension to the ``EventTarget`` class that sets the number of max event listeners as ``n``.

**Arguments**

* ``n`` (*number*)

**Example**

.. code-block:: javascript

   const events = require('node:events');

   const instance = new events.NodeEventTarget();
   const result = instance.setMaxListeners(1);
   console.log(result);


Properties and constants
------------------------

* ``captureRejections`` — boolean
* ``captureRejectionSymbol`` — symbol
* ``CustomEvent.detail`` — undefined
* ``defaultMaxListeners`` — number
* ``errorMonitor`` — symbol
* ``Event.bubbles`` — undefined
* ``Event.cancelable`` — undefined
* ``Event.cancelBubble`` — undefined
* ``Event.composed`` — undefined
* ``Event.currentTarget`` — undefined
* ``Event.defaultPrevented`` — undefined
* ``Event.eventPhase`` — undefined
* ``Event.isTrusted`` — undefined
* ``Event.returnValue`` — boolean
* ``Event.srcElement`` — undefined
* ``Event.target`` — undefined
* ``Event.timeStamp`` — undefined
* ``Event.type`` — undefined
* ``length`` — number
* ``name`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
