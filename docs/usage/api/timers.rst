timers
======

The ``timer`` module exposes a global API for scheduling functions to be called at some future period of time. Because the timer functions are globals, there is no need to call ``require('node:timers')`` to use the API.

The timer functions within Node.js implement a similar API as the timers API provided by Web Browsers but use a different internal implementation that is built around the Node.js Event Loop.

Loading
-------

.. code-block:: javascript

   const timers = require('node:timers');
   // ESM
   import * as timers from 'node:timers';

API
---

clearImmediate
~~~~~~~~~~~~~~

.. code-block:: text

   clearImmediate(immediate)

*Kind:* function · *Arguments:* 1

Cancels an ``Immediate`` object created by ``setImmediate()``.

**Arguments**

* ``immediate`` (*Immediate*): An ``Immediate`` object as returned by

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#clearimmediateimmediate>`__

clearInterval
~~~~~~~~~~~~~

.. code-block:: text

   clearInterval(timeout)

*Kind:* function · *Arguments:* 1

Cancels a ``Timeout`` object created by ``setInterval()``.

**Arguments**

* ``timeout`` (*Timeout|string|number*): A ``Timeout`` object as returned by ``setInterval()``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#clearintervaltimeout>`__

clearTimeout
~~~~~~~~~~~~

.. code-block:: text

   clearTimeout(timeout)

*Kind:* function · *Arguments:* 1

Cancels a ``Timeout`` object created by ``setTimeout()``.

**Arguments**

* ``timeout`` (*Timeout|string|number*): A ``Timeout`` object as returned by ``setTimeout()``

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#cleartimeouttimeout>`__

Immediate
~~~~~~~~~

.. code-block:: text

   Immediate

*Kind:* class · *Arguments:* 2

This object is created internally and is returned from ``setImmediate()``. It can be passed to ``clearImmediate()`` in order to cancel the scheduled actions.

By default, when an immediate is scheduled, the Node.js event loop will continue running as long as the immediate is active. The ``Immediate`` object returned by ``setImmediate()`` exports both ``immediate.ref()`` and ``immediate.unref()`` functions that can be used to control this default behavior.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#class-immediate>`__

setImmediate
~~~~~~~~~~~~

.. code-block:: text

   setImmediate(callback[, ...args])

*Kind:* function · *Arguments:* 2

Schedules the "immediate" execution of the ``callback`` after I/O events' callbacks.

When multiple calls to ``setImmediate()`` are made, the ``callback`` functions are queued for execution in the order in which they are created. The entire callback queue is processed every event loop iteration. If an immediate timer is queued from inside an executing callback, that timer will not be triggered until the next event loop iteration.

**Arguments**

* ``callback`` (*Function*): The function to call at the end of this turn of
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#setimmediatecallback-args>`__

setInterval
~~~~~~~~~~~

.. code-block:: text

   setInterval(callback[, delay[, ...args]])

*Kind:* function · *Arguments:* 3

Schedules repeated execution of ``callback`` every ``delay`` milliseconds.

When ``delay`` is larger than ``2147483647`` or less than ``1`` or ``NaN``, the ``delay`` will be set to ``1``. Non-integer delays are truncated to an integer.

**Arguments**

* ``callback`` (*Function*): The function to call when the timer elapses.
* ``delay`` (*number*) — optional: The number of milliseconds to wait before calling the
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#setintervalcallback-delay-args>`__

setTimeout
~~~~~~~~~~

.. code-block:: text

   setTimeout(callback[, delay[, ...args]])

*Kind:* function · *Arguments:* 3

Schedules execution of a one-time ``callback`` after ``delay`` milliseconds.

The ``callback`` will likely not be invoked in precisely ``delay`` milliseconds. Node.js makes no guarantees about the exact timing of when callbacks will fire, nor of their ordering. The callback will be called as close as possible to the time specified.

**Arguments**

* ``callback`` (*Function*): The function to call when the timer elapses.
* ``delay`` (*number*) — optional: The number of milliseconds to wait before calling the
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#settimeoutcallback-delay-args>`__

Timeout
~~~~~~~

.. code-block:: text

   Timeout

*Kind:* class · *Arguments:* 5

This object is created internally and is returned from ``setTimeout()`` and ``setInterval()``. It can be passed to either ``clearTimeout()`` or ``clearInterval()`` in order to cancel the scheduled actions.

By default, when a timer is scheduled using either ``setTimeout()`` or ``setInterval()``, the Node.js event loop will continue running as long as the timer is active. Each of the ``Timeout`` objects returned by these functions export both ``timeout.ref()`` and ``timeout.unref()`` functions that can be used to control this default behavior.

This is a class; construct instances with ``new``.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#class-timeout>`__

Immediate class
---------------

**Immediate** — This object is created internally and is returned from ``setImmediate()``. It can be passed to ``clearImmediate()`` in order to cancel the scheduled actions.

By default, when an immediate is scheduled, the Node.js event loop will continue running as long as the immediate is active. The ``Immediate`` object returned by ``setImmediate()`` exports both ``immediate.ref()`` and ``immediate.unref()`` functions that can be used to control this default behavior.

Immediate.hasRef
~~~~~~~~~~~~~~~~

.. code-block:: text

   immediate.hasRef()

*Kind:* function · *Arguments:* 0

If true, the ``Immediate`` object will keep the Node.js event loop active.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#immediatehasref>`__

Immediate.ref
~~~~~~~~~~~~~

.. code-block:: text

   immediate.ref()

*Kind:* function · *Arguments:* 0

When called, requests that the Node.js event loop *not* exit so long as the ``Immediate`` is active. Calling ``immediate.ref()`` multiple times will have no effect.

By default, all ``Immediate`` objects are "ref'ed", making it normally unnecessary to call ``immediate.ref()`` unless ``immediate.unref()`` had been called previously.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#immediateref>`__

Immediate.unref
~~~~~~~~~~~~~~~

.. code-block:: text

   immediate.unref()

*Kind:* function · *Arguments:* 0

When called, the active ``Immediate`` object will not require the Node.js event loop to remain active. If there is no other activity keeping the event loop running, the process may exit before the ``Immediate`` object's callback is invoked. Calling ``immediate.unref()`` multiple times will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#immediateunref>`__

promises
--------

promises.scheduler.wait
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   timersPromises.scheduler.wait(delay[, options])

*Kind:* function · *Arguments:* 2

An experimental API defined by the Scheduling APIs draft specification being developed as a standard Web Platform API.

Calling ``timersPromises.scheduler.wait(delay, options)`` is equivalent to calling ``timersPromises.setTimeout(delay, undefined, options)``.

**Arguments**

* ``delay`` (*number*): The number of milliseconds to wait before resolving the
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { scheduler } from 'node:timers/promises';

   await scheduler.wait(1000); // Wait one second before continuing

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timerspromisesschedulerwaitdelay-options>`__

promises.scheduler.yield
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   timersPromises.scheduler.yield()

*Kind:* function · *Arguments:* 0

An experimental API defined by the Scheduling APIs draft specification being developed as a standard Web Platform API.

Calling ``timersPromises.scheduler.yield()`` is equivalent to calling ``timersPromises.setImmediate()`` with no arguments.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timerspromisesscheduleryield>`__

promises.setImmediate
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   setImmediate(callback[, ...args])

*Kind:* function · *Arguments:* 2

Schedules the "immediate" execution of the ``callback`` after I/O events' callbacks.

When multiple calls to ``setImmediate()`` are made, the ``callback`` functions are queued for execution in the order in which they are created. The entire callback queue is processed every event loop iteration. If an immediate timer is queued from inside an executing callback, that timer will not be triggered until the next event loop iteration.

**Arguments**

* ``callback`` (*Function*): The function to call at the end of this turn of
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#setimmediatecallback-args>`__

promises.setInterval
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   setInterval(callback[, delay[, ...args]])

*Kind:* function · *Arguments:* 3

Schedules repeated execution of ``callback`` every ``delay`` milliseconds.

When ``delay`` is larger than ``2147483647`` or less than ``1`` or ``NaN``, the ``delay`` will be set to ``1``. Non-integer delays are truncated to an integer.

**Arguments**

* ``callback`` (*Function*): The function to call when the timer elapses.
* ``delay`` (*number*) — optional: The number of milliseconds to wait before calling the
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#setintervalcallback-delay-args>`__

promises.setTimeout
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   setTimeout(callback[, delay[, ...args]])

*Kind:* function · *Arguments:* 3

Schedules execution of a one-time ``callback`` after ``delay`` milliseconds.

The ``callback`` will likely not be invoked in precisely ``delay`` milliseconds. Node.js makes no guarantees about the exact timing of when callbacks will fire, nor of their ordering. The callback will be called as close as possible to the time specified.

**Arguments**

* ``callback`` (*Function*): The function to call when the timer elapses.
* ``delay`` (*number*) — optional: The number of milliseconds to wait before calling the
* ``...args`` — optional

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#settimeoutcallback-delay-args>`__

Timeout class
-------------

**Timeout** — This object is created internally and is returned from ``setTimeout()`` and ``setInterval()``. It can be passed to either ``clearTimeout()`` or ``clearInterval()`` in order to cancel the scheduled actions.

By default, when a timer is scheduled using either ``setTimeout()`` or ``setInterval()``, the Node.js event loop will continue running as long as the timer is active. Each of the ``Timeout`` objects returned by these functions export both ``timeout.ref()`` and ``timeout.unref()`` functions that can be used to control this default behavior.

Timeout.close
~~~~~~~~~~~~~

.. code-block:: text

   timeout.close()

*Kind:* function · *Arguments:* 0

Cancels the timeout.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timeoutclose>`__

Timeout.hasRef
~~~~~~~~~~~~~~

.. code-block:: text

   timeout.hasRef()

*Kind:* function · *Arguments:* 0

If true, the ``Timeout`` object will keep the Node.js event loop active.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timeouthasref>`__

Timeout.ref
~~~~~~~~~~~

.. code-block:: text

   timeout.ref()

*Kind:* function · *Arguments:* 0

When called, requests that the Node.js event loop *not* exit so long as the ``Timeout`` is active. Calling ``timeout.ref()`` multiple times will have no effect.

By default, all ``Timeout`` objects are "ref'ed", making it normally unnecessary to call ``timeout.ref()`` unless ``timeout.unref()`` had been called previously.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timeoutref>`__

Timeout.refresh
~~~~~~~~~~~~~~~

.. code-block:: text

   timeout.refresh()

*Kind:* function · *Arguments:* 0

Sets the timer's start time to the current time, and reschedules the timer to call its callback at the previously specified duration adjusted to the current time. This is useful for refreshing a timer without allocating a new JavaScript object.

Using this on a timer that has already called its callback will reactivate the timer.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timeoutrefresh>`__

Timeout.unref
~~~~~~~~~~~~~

.. code-block:: text

   timeout.unref()

*Kind:* function · *Arguments:* 0

When called, the active ``Timeout`` object will not require the Node.js event loop to remain active. If there is no other activity keeping the event loop running, the process may exit before the ``Timeout`` object's callback is invoked. Calling ``timeout.unref()`` multiple times will have no effect.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/timers.html#timeoutunref>`__

Properties and constants
------------------------

* ``promises`` — object
* ``promises.scheduler`` — object
* ``scheduler`` — object
* ``wait`` — undefined
* ``yield`` — undefined

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
