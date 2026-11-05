dns
===

The ``node:dns`` module enables name resolution. For example, use it to look up IP addresses of host names.

Although named for the Domain Name System (DNS), it does not always use the DNS protocol for lookups. ``dns.lookup()`` uses the operating system facilities to perform name resolution. It may not need to perform any network communication. To perform name resolution the way other applications on the same system do, use ``dns.lookup()``.

Loading
-------

.. code-block:: javascript

   const dns = require('node:dns');
   // ESM
   import * as dns from 'node:dns';

Quick start
-----------

.. code-block:: javascript

   const dns = require('node:dns');

   dns.lookup('example.org', (err, address, family) => {
     console.log('address: %j family: IPv%s', address, family);
   });
   // address: "2606:2800:21f:cb07:6820:80da:af6b:8b2c" family: IPv6

API
---

getServers
~~~~~~~~~~

.. code-block:: text

   dns.getServers()

*Kind:* function · *Arguments:* 0

Returns an array of IP address strings, formatted according to RFC 5952, that are currently configured for DNS resolution. A string will include a port section if a custom port is used.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsgetservers>`__

lookup
~~~~~~

.. code-block:: text

   dns.lookup()

*Kind:* function · *Arguments:* 3

Type: Runtime

Previous versions of Node.js supported ``dns.lookup()`` with a falsy host name like ``dns.lookup(false)`` due to backward compatibility. This behavior is undocumented and is thought to be unused in real world apps. It will become an error in future versions of Node.js.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/deprecations.html#dep0118-dnslookup-support-for-a-falsy-host-name>`__

lookupService
~~~~~~~~~~~~~

.. code-block:: text

   dns.lookupService(address, port, callback)

*Kind:* function · *Arguments:* 3

Resolves the given ``address`` and ``port`` into a host name and service using the operating system's underlying ``getnameinfo`` implementation.

If ``address`` is not a valid IP address, a ``TypeError`` will be thrown. The ``port`` will be coerced to a number. If it is not a legal port, a ``TypeError`` will be thrown.

**Arguments**

* ``address`` (*string*)
* ``port`` (*number*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const dns = require('node:dns');
   dns.lookupService('127.0.0.1', 22, (err, hostname, service) => {
     console.log(hostname, service);
     // Prints: localhost ssh
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnslookupserviceaddress-port-callback>`__

resolve
~~~~~~~

.. code-block:: text

   dns.resolve(hostname[, rrtype], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve a host name (e.g. ``'nodejs.org'``) into an array of the resource records. The ``callback`` function has arguments ``(err, records)``. When successful, ``records`` will be an array of resource records. The type and structure of individual results varies based on ``rrtype``:

On error, ``err`` is an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``rrtype`` (*string*) — optional: Resource record type. **Default:** ``'A'``.
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvehostname-rrtype-callback>`__

resolve4
~~~~~~~~

.. code-block:: text

   dns.resolve4(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve an IPv4 addresses (``A`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv4 addresses (e.g. ``['74.125.79.104', '74.125.79.105', '74.125.79.106']``).

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve4hostname-options-callback>`__

resolve6
~~~~~~~~

.. code-block:: text

   dns.resolve6(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve IPv6 addresses (``AAAA`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv6 addresses.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve6hostname-options-callback>`__

resolveAny
~~~~~~~~~~

.. code-block:: text

   dns.resolveAny(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve all records (also known as ``ANY`` or ``*`` query). The ``ret`` argument passed to the ``callback`` function will be an array containing various types of records. Each object has a property ``type`` that indicates the type of the current record. And depending on the ``type``, additional properties will be present on the object:

Here is an example of the ``ret`` object passed to the callback:

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   [ { type: 'A', address: '127.0.0.1', ttl: 299 },
     { type: 'CNAME', value: 'example.com' },
     { type: 'MX', exchange: 'alt4.aspmx.l.example.com', priority: 50 },
     { type: 'NS', value: 'ns1.example.com' },
     { type: 'TXT', entries: [ 'v=spf1 include:_spf.example.com ~all' ] },
     { type: 'SOA',
       nsname: 'ns1.example.com',
       hostmaster: 'admin.example.com',
       serial: 156696742,
       refresh: 900,
       retry: 900,
       expire: 1800,
       minttl: 60 } ];

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolveanyhostname-callback>`__

resolveCaa
~~~~~~~~~~

.. code-block:: text

   dns.resolveCaa(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve ``CAA`` records for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of certification authority authorization records available for the ``hostname`` (e.g. ``[{critical: 0, iodef: 'mailto:pki@example.com'}, {critical: 128, issue: 'pki.example.com'}]``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvecaahostname-callback>`__

resolveCname
~~~~~~~~~~~~

.. code-block:: text

   dns.resolveCname(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve ``CNAME`` records for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of canonical name records available for the ``hostname`` (e.g. ``['bar.example.com']``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvecnamehostname-callback>`__

resolveMx
~~~~~~~~~

.. code-block:: text

   dns.resolveMx(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve mail exchange records (``MX`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of objects containing both a ``priority`` and ``exchange`` property (e.g. ``[{priority: 10, exchange: 'mx.example.com'}, ...]``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvemxhostname-callback>`__

resolveNaptr
~~~~~~~~~~~~

.. code-block:: text

   dns.resolveNaptr(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve regular expression-based records (``NAPTR`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of objects with the following properties:

* ``flags`` * ``service`` * ``regexp`` * ``replacement`` * ``order`` * ``preference``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     flags: 's',
     service: 'SIP+D2U',
     regexp: '',
     replacement: '_sip._udp.example.com',
     order: 30,
     preference: 100,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvenaptrhostname-callback>`__

resolveNs
~~~~~~~~~

.. code-block:: text

   dns.resolveNs(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve name server records (``NS`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of name server records available for ``hostname`` (e.g. ``['ns1.example.com', 'ns2.example.com']``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvenshostname-callback>`__

resolvePtr
~~~~~~~~~~

.. code-block:: text

   dns.resolvePtr(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve pointer records (``PTR`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will be an array of strings containing the reply records.

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolveptrhostname-callback>`__

Resolver
~~~~~~~~

.. code-block:: text

   dns.Resolver

*Kind:* class · *Arguments:* 1

An independent resolver for DNS requests.

Creating a new resolver uses the default server settings. Setting the servers used for a resolver using ``resolver.setServers()`` does not affect other resolvers:

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Resolver } = require('node:dns');
   const resolver = new Resolver();
   resolver.setServers(['4.4.4.4']);

   // This request will use the server at 4.4.4.4, independent of global settings.
   resolver.resolve4('example.org', (err, addresses) => {
     // ...
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#class-dnsresolver>`__

resolveSoa
~~~~~~~~~~

.. code-block:: text

   dns.resolveSoa(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve a start of authority record (``SOA`` record) for the ``hostname``. The ``address`` argument passed to the ``callback`` function will be an object with the following properties:

* ``nsname`` * ``hostmaster`` * ``serial`` * ``refresh`` * ``retry`` * ``expire`` * ``minttl``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     nsname: 'ns.example.com',
     hostmaster: 'root.example.com',
     serial: 2013101809,
     refresh: 10000,
     retry: 2400,
     expire: 604800,
     minttl: 3600,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvesoahostname-callback>`__

resolveSrv
~~~~~~~~~~

.. code-block:: text

   dns.resolveSrv(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve service records (``SRV`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will be an array of objects with the following properties:

* ``priority`` * ``weight`` * ``port`` * ``name``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     priority: 10,
     weight: 5,
     port: 21223,
     name: 'service.example.com',
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvesrvhostname-callback>`__

resolveTxt
~~~~~~~~~~

.. code-block:: text

   dns.resolveTxt(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve text queries (``TXT`` records) for the ``hostname``. The ``records`` argument passed to the ``callback`` function is a two-dimensional array of the text records available for ``hostname`` (e.g. ``[ ['v=spf1 ip4:0.0.0.0 ', '~all' ] ]``). Each sub-array contains TXT chunks of one record. Depending on the use case, these could be either joined together or treated separately.

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvetxthostname-callback>`__

reverse
~~~~~~~

.. code-block:: text

   dns.reverse(ip, callback)

*Kind:* function · *Arguments:* 2

Performs a reverse DNS query that resolves an IPv4 or IPv6 address to an array of host names.

On error, ``err`` is an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``ip`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsreverseip-callback>`__

setServers
~~~~~~~~~~

.. code-block:: text

   dns.setServers(servers)

*Kind:* function · *Arguments:* 1

Sets the IP address and port of servers to be used when performing DNS resolution. The ``servers`` argument is an array of RFC 5952 formatted addresses. If the port is the IANA default DNS port (53) it can be omitted.

An error will be thrown if an invalid address is provided.

**Arguments**

* ``servers`` (*string\[]*): array of RFC 5952 formatted addresses

**Example**

.. code-block:: javascript

   dns.setServers([
     '8.8.8.8',
     '[2001:4860:4860::8888]',
     '8.8.8.8:1053',
     '[2001:4860:4860::8888]:1053',
   ]);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnssetserversservers>`__

promises
--------

promises.lookup
~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.lookup(hostname[, options])

*Kind:* function · *Arguments:* 2

Resolves a host name (e.g. ``'nodejs.org'``) into the first found A (IPv4) or AAAA (IPv6) record. All ``option`` properties are optional. If ``options`` is an integer, then it must be ``4`` or ``6`` – if ``options`` is not provided, then either IPv4 or IPv6 addresses, or both, are returned if found.

With the ``all`` option set to ``true``, the ``Promise`` is resolved with ``addresses`` being an array of objects with the properties ``address`` and ``family``.

**Arguments**

* ``hostname`` (*string*)
* ``options`` (*integer | Object*) — optional

**Example**

.. code-block:: javascript

   const dns = require('node:dns');
   const dnsPromises = dns.promises;
   const options = {
     family: 6,
     hints: dns.ADDRCONFIG | dns.V4MAPPED,
   };

   dnsPromises.lookup('example.org', options).then((result) => {
     console.log('address: %j family: IPv%s', result.address, result.family);
     // address: "2606:2800:21f:cb07:6820:80da:af6b:8b2c" family: IPv6
   });

   // When options.all is true, the result will be an Array.
   options.all = true;
   dnsPromises.lookup('example.org', options).then((result) => {
     console.log('addresses: %j', result);
     // addresses: [{"address":"2606:2800:21f:cb07:6820:80da:af6b:8b2c","family":6}]
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromiseslookuphostname-options>`__

promises.lookupService
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.lookupService(address, port)

*Kind:* function · *Arguments:* 2

Resolves the given ``address`` and ``port`` into a host name and service using the operating system's underlying ``getnameinfo`` implementation.

If ``address`` is not a valid IP address, a ``TypeError`` will be thrown. The ``port`` will be coerced to a number. If it is not a legal port, a ``TypeError`` will be thrown.

**Arguments**

* ``address`` (*string*)
* ``port`` (*number*)

**Example**

.. code-block:: javascript

   const dnsPromises = require('node:dns').promises;
   dnsPromises.lookupService('127.0.0.1', 22).then((result) => {
     console.log(result.hostname, result.service);
     // Prints: localhost ssh
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromiseslookupserviceaddress-port>`__

promises.resolve
~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve(hostname[, rrtype], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve a host name (e.g. ``'nodejs.org'``) into an array of the resource records. The ``callback`` function has arguments ``(err, records)``. When successful, ``records`` will be an array of resource records. The type and structure of individual results varies based on ``rrtype``:

On error, ``err`` is an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``rrtype`` (*string*) — optional: Resource record type. **Default:** ``'A'``.
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvehostname-rrtype-callback>`__

promises.resolve4
~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve4(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve an IPv4 addresses (``A`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv4 addresses (e.g. ``['74.125.79.104', '74.125.79.105', '74.125.79.106']``).

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve4hostname-options-callback>`__

promises.resolve6
~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve6(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve IPv6 addresses (``AAAA`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv6 addresses.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve6hostname-options-callback>`__

promises.resolveAny
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveAny(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve all records (also known as ``ANY`` or ``*`` query). On success, the ``Promise`` is resolved with an array containing various types of records. Each object has a property ``type`` that indicates the type of the current record. And depending on the ``type``, additional properties will be present on the object:

Here is an example of the result object:

**Arguments**

* ``hostname`` (*string*)

**Example**

.. code-block:: javascript

   [ { type: 'A', address: '127.0.0.1', ttl: 299 },
     { type: 'CNAME', value: 'example.com' },
     { type: 'MX', exchange: 'alt4.aspmx.l.example.com', priority: 50 },
     { type: 'NS', value: 'ns1.example.com' },
     { type: 'TXT', entries: [ 'v=spf1 include:_spf.example.com ~all' ] },
     { type: 'SOA',
       nsname: 'ns1.example.com',
       hostmaster: 'admin.example.com',
       serial: 156696742,
       refresh: 900,
       retry: 900,
       expire: 1800,
       minttl: 60 } ];

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolveanyhostname>`__

promises.resolveCaa
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveCaa(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve ``CAA`` records for the ``hostname``. On success, the ``Promise`` is resolved with an array of objects containing available certification authority authorization records available for the ``hostname`` (e.g. ``[{critical: 0, iodef: 'mailto:pki@example.com'},{critical: 128, issue: 'pki.example.com'}]``).

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvecaahostname>`__

promises.resolveCname
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveCname(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve ``CNAME`` records for the ``hostname``. On success, the ``Promise`` is resolved with an array of canonical name records available for the ``hostname`` (e.g. ``['bar.example.com']``).

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvecnamehostname>`__

promises.resolveMx
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveMx(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve mail exchange records (``MX`` records) for the ``hostname``. On success, the ``Promise`` is resolved with an array of objects containing both a ``priority`` and ``exchange`` property (e.g. ``[{priority: 10, exchange: 'mx.example.com'}, ...]``).

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvemxhostname>`__

promises.resolveNaptr
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveNaptr(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve regular expression-based records (``NAPTR`` records) for the ``hostname``. On success, the ``Promise`` is resolved with an array of objects with the following properties:

* ``flags`` * ``service`` * ``regexp`` * ``replacement`` * ``order`` * ``preference``

**Arguments**

* ``hostname`` (*string*)

**Example**

.. code-block:: javascript

   ({
     flags: 's',
     service: 'SIP+D2U',
     regexp: '',
     replacement: '_sip._udp.example.com',
     order: 30,
     preference: 100,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvenaptrhostname>`__

promises.resolveNs
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveNs(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve name server records (``NS`` records) for the ``hostname``. On success, the ``Promise`` is resolved with an array of name server records available for ``hostname`` (e.g. ``['ns1.example.com', 'ns2.example.com']``).

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvenshostname>`__

promises.resolvePtr
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolvePtr(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve pointer records (``PTR`` records) for the ``hostname``. On success, the ``Promise`` is resolved with an array of strings containing the reply records.

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolveptrhostname>`__

promises.resolveSoa
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveSoa(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve a start of authority record (``SOA`` record) for the ``hostname``. On success, the ``Promise`` is resolved with an object with the following properties:

* ``nsname`` * ``hostmaster`` * ``serial`` * ``refresh`` * ``retry`` * ``expire`` * ``minttl``

**Arguments**

* ``hostname`` (*string*)

**Example**

.. code-block:: javascript

   ({
     nsname: 'ns.example.com',
     hostmaster: 'root.example.com',
     serial: 2013101809,
     refresh: 10000,
     retry: 2400,
     expire: 604800,
     minttl: 3600,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvesoahostname>`__

promises.resolveSrv
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveSrv(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve service records (``SRV`` records) for the ``hostname``. On success, the ``Promise`` is resolved with an array of objects with the following properties:

* ``priority`` * ``weight`` * ``port`` * ``name``

**Arguments**

* ``hostname`` (*string*)

**Example**

.. code-block:: javascript

   ({
     priority: 10,
     weight: 5,
     port: 21223,
     name: 'service.example.com',
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvesrvhostname>`__

promises.resolveTxt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.resolveTxt(hostname)

*Kind:* function · *Arguments:* 1

Uses the DNS protocol to resolve text queries (``TXT`` records) for the ``hostname``. On success, the ``Promise`` is resolved with a two-dimensional array of the text records available for ``hostname`` (e.g. ``[ ['v=spf1 ip4:0.0.0.0 ', '~all' ] ]``). Each sub-array contains TXT chunks of one record. Depending on the use case, these could be either joined together or treated separately.

**Arguments**

* ``hostname`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesresolvetxthostname>`__

promises.reverse
~~~~~~~~~~~~~~~~

.. code-block:: text

   dnsPromises.reverse(ip)

*Kind:* function · *Arguments:* 1

Performs a reverse DNS query that resolves an IPv4 or IPv6 address to an array of host names.

On error, the ``Promise`` is rejected with an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``ip`` (*string*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnspromisesreverseip>`__

Resolver class
--------------

**Resolver** — An independent resolver for DNS requests.

Creating a new resolver uses the default server settings. Setting the servers used for a resolver using ``resolver.setServers()`` does not affect other resolvers:

Resolver.cancel
~~~~~~~~~~~~~~~

.. code-block:: text

   resolver.cancel()

*Kind:* function · *Arguments:* 0

Cancel all outstanding DNS queries made by this resolver. The corresponding callbacks will be called with an error with code ``ECANCELLED``.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#resolvercancel>`__

Resolver.getServers
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.getServers()

*Kind:* function · *Arguments:* 0

Returns an array of IP address strings, formatted according to RFC 5952, that are currently configured for DNS resolution. A string will include a port section if a custom port is used.

This callable takes no arguments.

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsgetservers>`__

Resolver.lookup
~~~~~~~~~~~~~~~

.. code-block:: text

   dns.lookup(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Resolves a host name (e.g. ``'nodejs.org'``) into the first found A (IPv4) or AAAA (IPv6) record. All ``option`` properties are optional. If ``options`` is an integer, then it must be ``4`` or ``6`` – if ``options`` is not provided, then either IPv4 or IPv6 addresses, or both, are returned if found.

With the ``all`` option set to ``true``, the arguments for ``callback`` change to ``(err, addresses)``, with ``addresses`` being an array of objects with the properties ``address`` and ``family``.

**Arguments**

* ``hostname`` (*string*)
* ``options`` (*integer | Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const dns = require('node:dns');
   const options = {
     family: 6,
     hints: dns.ADDRCONFIG | dns.V4MAPPED,
   };
   dns.lookup('example.org', options, (err, address, family) =>
     console.log('address: %j family: IPv%s', address, family));
   // address: "2606:2800:21f:cb07:6820:80da:af6b:8b2c" family: IPv6

   // When options.all is true, the result will be an Array.
   options.all = true;
   dns.lookup('example.org', options, (err, addresses) =>
     console.log('addresses: %j', addresses));
   // addresses: [{"address":"2606:2800:21f:cb07:6820:80da:af6b:8b2c","family":6}]

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnslookuphostname-options-callback>`__

Resolver.resolve
~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve(hostname[, rrtype], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve a host name (e.g. ``'nodejs.org'``) into an array of the resource records. The ``callback`` function has arguments ``(err, records)``. When successful, ``records`` will be an array of resource records. The type and structure of individual results varies based on ``rrtype``:

On error, ``err`` is an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``rrtype`` (*string*) — optional: Resource record type. **Default:** ``'A'``.
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvehostname-rrtype-callback>`__

Resolver.resolve4
~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve4(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve an IPv4 addresses (``A`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv4 addresses (e.g. ``['74.125.79.104', '74.125.79.105', '74.125.79.106']``).

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve4hostname-options-callback>`__

Resolver.resolve6
~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolve6(hostname[, options], callback)

*Kind:* function · *Arguments:* 3

Uses the DNS protocol to resolve IPv6 addresses (``AAAA`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of IPv6 addresses.

**Arguments**

* ``hostname`` (*string*): Host name to resolve.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolve6hostname-options-callback>`__

Resolver.resolveAny
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveAny(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve all records (also known as ``ANY`` or ``*`` query). The ``ret`` argument passed to the ``callback`` function will be an array containing various types of records. Each object has a property ``type`` that indicates the type of the current record. And depending on the ``type``, additional properties will be present on the object:

Here is an example of the ``ret`` object passed to the callback:

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   [ { type: 'A', address: '127.0.0.1', ttl: 299 },
     { type: 'CNAME', value: 'example.com' },
     { type: 'MX', exchange: 'alt4.aspmx.l.example.com', priority: 50 },
     { type: 'NS', value: 'ns1.example.com' },
     { type: 'TXT', entries: [ 'v=spf1 include:_spf.example.com ~all' ] },
     { type: 'SOA',
       nsname: 'ns1.example.com',
       hostmaster: 'admin.example.com',
       serial: 156696742,
       refresh: 900,
       retry: 900,
       expire: 1800,
       minttl: 60 } ];

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolveanyhostname-callback>`__

Resolver.resolveCaa
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveCaa(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve ``CAA`` records for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of certification authority authorization records available for the ``hostname`` (e.g. ``[{critical: 0, iodef: 'mailto:pki@example.com'}, {critical: 128, issue: 'pki.example.com'}]``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvecaahostname-callback>`__

Resolver.resolveCname
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveCname(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve ``CNAME`` records for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of canonical name records available for the ``hostname`` (e.g. ``['bar.example.com']``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvecnamehostname-callback>`__

Resolver.resolveMx
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveMx(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve mail exchange records (``MX`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of objects containing both a ``priority`` and ``exchange`` property (e.g. ``[{priority: 10, exchange: 'mx.example.com'}, ...]``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvemxhostname-callback>`__

Resolver.resolveNaptr
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveNaptr(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve regular expression-based records (``NAPTR`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of objects with the following properties:

* ``flags`` * ``service`` * ``regexp`` * ``replacement`` * ``order`` * ``preference``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     flags: 's',
     service: 'SIP+D2U',
     regexp: '',
     replacement: '_sip._udp.example.com',
     order: 30,
     preference: 100,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvenaptrhostname-callback>`__

Resolver.resolveNs
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveNs(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve name server records (``NS`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will contain an array of name server records available for ``hostname`` (e.g. ``['ns1.example.com', 'ns2.example.com']``).

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvenshostname-callback>`__

Resolver.resolvePtr
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolvePtr(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve pointer records (``PTR`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will be an array of strings containing the reply records.

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolveptrhostname-callback>`__

Resolver.resolveSoa
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveSoa(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve a start of authority record (``SOA`` record) for the ``hostname``. The ``address`` argument passed to the ``callback`` function will be an object with the following properties:

* ``nsname`` * ``hostmaster`` * ``serial`` * ``refresh`` * ``retry`` * ``expire`` * ``minttl``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     nsname: 'ns.example.com',
     hostmaster: 'root.example.com',
     serial: 2013101809,
     refresh: 10000,
     retry: 2400,
     expire: 604800,
     minttl: 3600,
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvesoahostname-callback>`__

Resolver.resolveSrv
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveSrv(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve service records (``SRV`` records) for the ``hostname``. The ``addresses`` argument passed to the ``callback`` function will be an array of objects with the following properties:

* ``priority`` * ``weight`` * ``port`` * ``name``

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   ({
     priority: 10,
     weight: 5,
     port: 21223,
     name: 'service.example.com',
   });

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvesrvhostname-callback>`__

Resolver.resolveTxt
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.resolveTxt(hostname, callback)

*Kind:* function · *Arguments:* 2

Uses the DNS protocol to resolve text queries (``TXT`` records) for the ``hostname``. The ``records`` argument passed to the ``callback`` function is a two-dimensional array of the text records available for ``hostname`` (e.g. ``[ ['v=spf1 ip4:0.0.0.0 ', '~all' ] ]``). Each sub-array contains TXT chunks of one record. Depending on the use case, these could be either joined together or treated separately.

**Arguments**

* ``hostname`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsresolvetxthostname-callback>`__

Resolver.reverse
~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.reverse(ip, callback)

*Kind:* function · *Arguments:* 2

Performs a reverse DNS query that resolves an IPv4 or IPv6 address to an array of host names.

On error, ``err`` is an ``Error`` object, where ``err.code`` is one of the DNS error codes.

**Arguments**

* ``ip`` (*string*)
* ``callback`` (*Function*)

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnsreverseip-callback>`__

Resolver.setServers
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   dns.setServers(servers)

*Kind:* function · *Arguments:* 1

Sets the IP address and port of servers to be used when performing DNS resolution. The ``servers`` argument is an array of RFC 5952 formatted addresses. If the port is the IANA default DNS port (53) it can be omitted.

An error will be thrown if an invalid address is provided.

**Arguments**

* ``servers`` (*string\[]*): array of RFC 5952 formatted addresses

**Example**

.. code-block:: javascript

   dns.setServers([
     '8.8.8.8',
     '[2001:4860:4860::8888]',
     '8.8.8.8:1053',
     '[2001:4860:4860::8888]:1053',
   ]);

`Node.js reference <https://nodejs.org/docs/latest-v24.x/api/dns.html#dnssetserversservers>`__

Properties and constants
------------------------

* ``ADDRCONFIG`` — number
* ``ADDRGETNETWORKPARAMS`` — string
* ``ALL`` — number
* ``BADFAMILY`` — string
* ``BADFLAGS`` — string
* ``BADHINTS`` — string
* ``BADNAME`` — string
* ``BADQUERY`` — string
* ``BADRESP`` — string
* ``BADSTR`` — string
* ``CANCELLED`` — string
* ``CONNREFUSED`` — string
* ``DESTRUCTION`` — string
* ``EOF`` — string
* ``FILE`` — string
* ``FORMERR`` — string
* ``LOADIPHLPAPI`` — string
* ``NODATA`` — string
* ``NOMEM`` — string
* ``NONAME`` — string
* ``NOTFOUND`` — string
* ``NOTIMP`` — string
* ``NOTINITIALIZED`` — string
* ``promises`` — object
* ``REFUSED`` — string
* ``SERVFAIL`` — string
* ``TIMEOUT`` — string
* ``V4MAPPED`` — number

.. note::

   Signatures, argument names, return values and examples for Node APIs come from the Node.js v24.22.2 documentation (MIT). Entries marked as private or GodeJS extensions are generated from the runtime snapshot.
