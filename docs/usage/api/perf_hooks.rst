perf_hooks
==========

This module provides an implementation of a subset of the W3C Web Performance APIs as well as additional APIs for Node.js-specific performance measurements.

Node.js supports the following Web Performance APIs:

Loading
-------

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');
   // ESM
   import * as perf_hooks from 'node:perf_hooks';

Quick start
-----------

.. code-block:: javascript

   const { PerformanceObserver, performance } = require('node:perf_hooks');

   const obs = new PerformanceObserver((items) => {
     console.log(items.getEntries()[0].duration);
   });
   obs.observe({ type: 'measure' });
   performance.measure('Start to Now');

   performance.mark('A');
   (async function doSomeLongRunningProcess() {
     await new Promise((r) => setTimeout(r, 5000));
     performance.measure('A to Now', 'A');

     performance.mark('B');
     performance.measure('A to B', 'A', 'B');
   })();

API
---

createHistogram
~~~~~~~~~~~~~~~

.. code-block:: text

   perf_hooks.createHistogram([options])

*Kind:* function · *Arguments:* 1

Returns a ``RecordableHistogram``.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.createHistogram({});
   console.log(result);


histogram
~~~~~~~~~

.. code-block:: text

   perf_hooks.histogram(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.histogram(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Histogram
~~~~~~~~~

.. code-block:: text

   Histogram

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { createHistogram } = require('node:perf_hooks');

   // Track latency with a 200ms SLO threshold, half-life of 100 samples
   const h = createHistogram({ halfLife: 100, threshold: 200_000_000 });

   // ... record latency values ...

   // Check burn rate against a 99.9% SLO
   const rate = h.burnRate(0.999);
   if (rate > 1) {
     console.log(`SLO burn rate: ${rate.toFixed(2)}x — error budget depleting`);
   }


monitorEventLoopDelay
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   perf_hooks.monitorEventLoopDelay([options])

*Kind:* function · *Arguments:* 1

*This property is an extension by Node.js. It is not available in Web browsers.*

Creates a histogram object that samples and reports the event loop delay over time. The delays will be reported in nanoseconds.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { monitorEventLoopDelay } = require('node:perf_hooks');
   const h = monitorEventLoopDelay({ resolution: 20 });
   h.enable();
   // Do something.
   h.disable();
   console.log(h.min);
   console.log(h.max);
   console.log(h.mean);
   console.log(h.stddev);
   console.log(h.percentiles);
   console.log(h.percentile(50));
   console.log(h.percentile(99));


PerformanceEntry
~~~~~~~~~~~~~~~~

.. code-block:: text

   PerformanceEntry

*Kind:* class · *Arguments:* 4

The constructor of this class is not exposed to users directly.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.PerformanceEntry(undefined, undefined, undefined, undefined);
   console.log(instance);


PerformanceNodeTiming
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   PerformanceNodeTiming

*Kind:* class · *Arguments:* 0

*This property is an extension by Node.js. It is not available in Web browsers.*

Provides timing details for Node.js itself. The constructor of this class is not exposed to users.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { performance } = require('node:perf_hooks');

   setImmediate(() => {
     console.log(performance.nodeTiming.uvMetricsInfo);
   });


PerformanceObserver
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   PerformanceObserver

*Kind:* class · *Arguments:* 1

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((list, observer) => {
     console.log(list.getEntries());

     performance.clearMarks();
     performance.clearMeasures();
     observer.disconnect();
   });
   obs.observe({ entryTypes: ['mark'], buffered: true });

   performance.mark('test');


PerformanceObserverEntryList
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   PerformanceObserverEntryList

*Kind:* class · *Arguments:* 1

The ``PerformanceObserverEntryList`` class is used to provide access to the ``PerformanceEntry`` instances passed to a ``PerformanceObserver``. The constructor of this class is not exposed to users.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((perfObserverList, observer) => {
     console.log(perfObserverList.getEntries());
     /**
      * [
      *   PerformanceEntry {
      *     name: 'test',
      *     entryType: 'mark',
      *     startTime: 81.465639,
      *     duration: 0,
      *     detail: null
      *   },
      *   PerformanceEntry {
      *     name: 'meow',
      *     entryType: 'mark',
      *     startTime: 81.860064,
      *     duration: 0,
      *     detail: null
      *   }
      * ]
      */

     performance.clearMarks();
     performance.clearMeasures();
     observer.disconnect();
   });
   obs.observe({ type: 'mark' });

   performance.mark('test');
   performance.mark('meow');


Histogram class
---------------

Histogram.percentile
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   histogram.percentile(percentile)

*Kind:* function · *Arguments:* 1

Returns the value at the given percentile.

**Arguments**

* ``percentile`` (*number*): A percentile value in the range (0, 100].

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.Histogram();
   const result = instance.percentile(1);
   console.log(result);


Histogram.record
~~~~~~~~~~~~~~~~

.. code-block:: text

   histogram.record(val)

*Kind:* function · *Arguments:* 1

**Arguments**

* ``val`` (*number|bigint*): The amount to record in the histogram.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.Histogram();
   const result = instance.record(1);
   console.log(result);


Histogram.reset
~~~~~~~~~~~~~~~

.. code-block:: text

   histogram.reset()

*Kind:* function · *Arguments:* 0

Resets the collected histogram data.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.Histogram();
   const result = instance.reset();
   console.log(result);


Histogram.toJSON
~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.toJSON()

*Kind:* function · *Arguments:* 0

An object which is JSON representation of the ``performance`` object. It is similar to ``window.performance.toJSON`` in browsers.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.Histogram();
   const result = instance.toJSON();
   console.log(result);


performance
-----------

**performance** — An object that can be used to collect performance metrics from the current Node.js instance. It is similar to ``window.performance`` in browsers.

performance.clearMarks
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.clearMarks([name])

*Kind:* function · *Arguments:* 1

If ``name`` is not provided, removes all ``PerformanceMark`` objects from the Performance Timeline. If ``name`` is provided, removes only the named mark.

**Arguments**

* ``name`` (*string*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.clearMarks('example');
   console.log(result);


performance.clearMeasures
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.clearMeasures([name])

*Kind:* function · *Arguments:* 1

If ``name`` is not provided, removes all ``PerformanceMeasure`` objects from the Performance Timeline. If ``name`` is provided, removes only the named measure.

**Arguments**

* ``name`` (*string*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.clearMeasures('example');
   console.log(result);


performance.eventLoopUtilization
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.eventLoopUtilization([utilization1[, utilization2]])

*Kind:* function · *Arguments:* 2

This is an alias of ``perf_hooks.eventLoopUtilization()``.

*This property is an extension by Node.js. It is not available in Web browsers.*

**Arguments**

* ``utilization1`` (*Object*) — optional: The result of a previous call to
* ``utilization2`` (*Object*) — optional: The result of a previous call to

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.eventLoopUtilization({}, {});
   console.log(result);


performance.getEntries
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntries()

*Kind:* function · *Arguments:* 0

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime``. If you are only interested in performance entries of certain types or that have certain names, see ``performance.getEntriesByType()`` and ``performance.getEntriesByName()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.getEntries();
   console.log(result);


performance.getEntriesByName
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntriesByName(name[, type])

*Kind:* function · *Arguments:* 2

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.name`` is equal to ``name``, and optionally, whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``name`` (*string*)
* ``type`` (*string*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.getEntriesByName('example', 'example');
   console.log(result);


performance.getEntriesByType
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.getEntriesByType(type)

*Kind:* function · *Arguments:* 1

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``type`` (*string*)

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.getEntriesByType('example');
   console.log(result);


performance.mark
~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.mark(name[, options])

*Kind:* function · *Arguments:* 2

Creates a new ``PerformanceMark`` entry in the Performance Timeline. A ``PerformanceMark`` is a subclass of ``PerformanceEntry`` whose ``performanceEntry.entryType`` is always ``'mark'``, and whose ``performanceEntry.duration`` is always ``0``. Performance marks are used to mark specific significant moments in the Performance Timeline.

The created ``PerformanceMark`` entry is put in the global Performance Timeline and can be queried with ``performance.getEntries``, ``performance.getEntriesByName``, and ``performance.getEntriesByType``. When the observation is performed, the entries should be cleared from the global Performance Timeline manually with ``performance.clearMarks``.

**Arguments**

* ``name`` (*string*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.mark('example', {});
   console.log(result);


performance.measure
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.measure(name[, startMarkOrOptions[, endMark]])

*Kind:* function · *Arguments:* 3

Creates a new ``PerformanceMeasure`` entry in the Performance Timeline. A ``PerformanceMeasure`` is a subclass of ``PerformanceEntry`` whose ``performanceEntry.entryType`` is always ``'measure'``, and whose ``performanceEntry.duration`` measures the number of milliseconds elapsed since ``startMark`` and ``endMark``.

The ``startMark`` argument may identify any *existing* ``PerformanceMark`` in the Performance Timeline, or *may* identify any of the timestamp properties provided by the ``PerformanceNodeTiming`` class. If the named ``startMark`` does not exist, an error is thrown.

**Arguments**

* ``name`` (*string*)
* ``startMarkOrOptions`` (*string|Object*) — optional: Optional.
* ``endMark`` (*string*) — optional: Optional. Must be omitted if ``startMarkOrOptions`` is an

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.measure('example', 'value', 'value');
   console.log(result);


performance.now
~~~~~~~~~~~~~~~

.. code-block:: text

   performance.now()

*Kind:* function · *Arguments:* 0

Returns the current high resolution millisecond timestamp, where 0 represents the start of the current ``node`` process.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.now();
   console.log(result);


performance.timerify
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.timerify(fn[, options])

*Kind:* function · *Arguments:* 2

This is an alias of ``perf_hooks.timerify()``.

*This property is an extension by Node.js. It is not available in Web browsers.*

**Arguments**

* ``fn`` (*Function*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.timerify(() => {}, {});
   console.log(result);


performance.toJSON
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.toJSON()

*Kind:* function · *Arguments:* 0

An object which is JSON representation of the ``performance`` object. It is similar to ``window.performance.toJSON`` in browsers.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const result = perf_hooks.performance.toJSON();
   console.log(result);


PerformanceEntry class
----------------------

**PerformanceEntry** — The constructor of this class is not exposed to users directly.

PerformanceEntry.toJSON
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.toJSON()

*Kind:* function · *Arguments:* 0

An object which is JSON representation of the ``performance`` object. It is similar to ``window.performance.toJSON`` in browsers.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.PerformanceEntry();
   const result = instance.toJSON();
   console.log(result);


PerformanceNodeTiming class
---------------------------

**PerformanceNodeTiming** — *This property is an extension by Node.js. It is not available in Web browsers.*

Provides timing details for Node.js itself. The constructor of this class is not exposed to users.

PerformanceNodeTiming.toJSON
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performance.toJSON()

*Kind:* function · *Arguments:* 0

An object which is JSON representation of the ``performance`` object. It is similar to ``window.performance.toJSON`` in browsers.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.PerformanceNodeTiming();
   const result = instance.toJSON();
   console.log(result);


PerformanceObserver class
-------------------------

PerformanceObserver._enqueue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   perf_hooks.PerformanceObserver._enqueue(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   console.log(typeof perf_hooks.PerformanceObserver.prototype._enqueue()); // 'function'
   console.log(perf_hooks.PerformanceObserver.prototype._enqueue().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


PerformanceObserver._flush
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   perf_hooks.PerformanceObserver._flush(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   console.log(typeof perf_hooks.PerformanceObserver.prototype._flush()); // 'function'
   console.log(perf_hooks.PerformanceObserver.prototype._flush().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


PerformanceObserver._matches
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   perf_hooks.PerformanceObserver._matches(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   console.log(typeof perf_hooks.PerformanceObserver.prototype._matches()); // 'function'
   console.log(perf_hooks.PerformanceObserver.prototype._matches().length);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


PerformanceObserver.disconnect
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserver.disconnect()

*Kind:* function · *Arguments:* 0

Disconnects the ``PerformanceObserver`` instance from all notifications.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.PerformanceObserver();
   const result = instance.disconnect();
   console.log(result);


PerformanceObserver.observe
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserver.observe(options)

*Kind:* function · *Arguments:* 1

Subscribes the ``PerformanceObserver`` instance to notifications of new ``PerformanceEntry`` instances identified either by ``options.entryTypes`` or ``options.type``:

**Arguments**

* ``options`` (*Object*)

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((list, observer) => {
     // Called once asynchronously. `list` contains three items.
   });
   obs.observe({ type: 'mark' });

   for (let n = 0; n < 3; n++)
     performance.mark(`test${n}`);


PerformanceObserver.takeRecords
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserver.takeRecords()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const perf_hooks = require('node:perf_hooks');

   const instance = new perf_hooks.PerformanceObserver();
   const result = instance.takeRecords();
   console.log(result);


PerformanceObserverEntryList class
----------------------------------

**PerformanceObserverEntryList** — The ``PerformanceObserverEntryList`` class is used to provide access to the ``PerformanceEntry`` instances passed to a ``PerformanceObserver``. The constructor of this class is not exposed to users.

PerformanceObserverEntryList.getEntries
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserverEntryList.getEntries()

*Kind:* function · *Arguments:* 0

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((perfObserverList, observer) => {
     console.log(perfObserverList.getEntries());
     /**
      * [
      *   PerformanceEntry {
      *     name: 'test',
      *     entryType: 'mark',
      *     startTime: 81.465639,
      *     duration: 0,
      *     detail: null
      *   },
      *   PerformanceEntry {
      *     name: 'meow',
      *     entryType: 'mark',
      *     startTime: 81.860064,
      *     duration: 0,
      *     detail: null
      *   }
      * ]
      */

     performance.clearMarks();
     performance.clearMeasures();
     observer.disconnect();
   });
   obs.observe({ type: 'mark' });

   performance.mark('test');
   performance.mark('meow');


PerformanceObserverEntryList.getEntriesByName
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserverEntryList.getEntriesByName(name[, type])

*Kind:* function · *Arguments:* 2

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.name`` is equal to ``name``, and optionally, whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``name`` (*string*)
* ``type`` (*string*) — optional

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((perfObserverList, observer) => {
     console.log(perfObserverList.getEntriesByName('meow'));
     /**
      * [
      *   PerformanceEntry {
      *     name: 'meow',
      *     entryType: 'mark',
      *     startTime: 98.545991,
      *     duration: 0,
      *     detail: null
      *   }
      * ]
      */
     console.log(perfObserverList.getEntriesByName('nope')); // []

     console.log(perfObserverList.getEntriesByName('test', 'mark'));
     /**
      * [
      *   PerformanceEntry {
      *     name: 'test',
      *     entryType: 'mark',
      *     startTime: 63.518931,
      *     duration: 0,
      *     detail: null
      *   }
      * ]
      */
     console.log(perfObserverList.getEntriesByName('test', 'measure')); // []

     performance.clearMarks();
     performance.clearMeasures();
     observer.disconnect();
   });
   obs.observe({ entryTypes: ['mark', 'measure'] });

   performance.mark('test');
   performance.mark('meow');


PerformanceObserverEntryList.getEntriesByType
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   performanceObserverEntryList.getEntriesByType(type)

*Kind:* function · *Arguments:* 1

Returns a list of ``PerformanceEntry`` objects in chronological order with respect to ``performanceEntry.startTime`` whose ``performanceEntry.entryType`` is equal to ``type``.

**Arguments**

* ``type`` (*string*)

**Example**

.. code-block:: javascript

   const {
     performance,
     PerformanceObserver,
   } = require('node:perf_hooks');

   const obs = new PerformanceObserver((perfObserverList, observer) => {
     console.log(perfObserverList.getEntriesByType('mark'));
     /**
      * [
      *   PerformanceEntry {
      *     name: 'test',
      *     entryType: 'mark',
      *     startTime: 55.897834,
      *     duration: 0,
      *     detail: null
      *   },
      *   PerformanceEntry {
      *     name: 'meow',
      *     entryType: 'mark',
      *     startTime: 56.350146,
      *     duration: 0,
      *     detail: null
      *   }
      * ]
      */
     performance.clearMarks();
     performance.clearMeasures();
     observer.disconnect();
   });
   obs.observe({ type: 'mark' });

   performance.mark('test');
   performance.mark('meow');


Properties and constants
------------------------

* ``performance`` — object
* ``performance.timeOrigin`` — number
* ``performance_node`` — object

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
