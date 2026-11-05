url
===

The ``node:url`` module provides utilities for URL resolution and parsing. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const url = require('node:url');
   // ESM
   import * as url from 'node:url';

Quick start
-----------

.. code-block:: javascript

   const url = require('node:url');

API
---

domainToASCII
~~~~~~~~~~~~~

.. code-block:: text

   url.domainToASCII(domain)

*Kind:* function · *Arguments:* 1

Returns the Punycode ASCII serialization of the ``domain``. If ``domain`` is an invalid domain, the empty string is returned.

It performs the inverse operation to ``url.domainToUnicode()``.

**Arguments**

* ``domain`` (*string*)

**Example**

.. code-block:: javascript

   const url = require('node:url');

   console.log(url.domainToASCII('español.com'));
   // Prints xn--espaol-zwa.com
   console.log(url.domainToASCII('中文.com'));
   // Prints xn--fiq228c.com
   console.log(url.domainToASCII('xn--iñvalid.com'));
   // Prints an empty string


domainToUnicode
~~~~~~~~~~~~~~~

.. code-block:: text

   url.domainToUnicode(domain)

*Kind:* function · *Arguments:* 1

Returns the Unicode serialization of the ``domain``. If ``domain`` is an invalid domain, the empty string is returned.

It performs the inverse operation to ``url.domainToASCII()``.

**Arguments**

* ``domain`` (*string*)

**Example**

.. code-block:: javascript

   const url = require('node:url');

   console.log(url.domainToUnicode('xn--espaol-zwa.com'));
   // Prints español.com
   console.log(url.domainToUnicode('xn--fiq228c.com'));
   // Prints 中文.com
   console.log(url.domainToUnicode('xn--iñvalid.com'));
   // Prints an empty string


fileURLToPath
~~~~~~~~~~~~~

.. code-block:: text

   url.fileURLToPath(url[, options])

*Kind:* function · *Arguments:* 2

This function ensures the correct decodings of percent-encoded characters as well as ensuring a cross-platform valid absolute path string.

**Security Considerations:**

**Arguments**

* ``url`` (*URL | string*): The file URL string or URL object to convert to a path.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { fileURLToPath } = require('node:url');
   new URL('file:///C:/path/').pathname;      // Incorrect: /C:/path/
   fileURLToPath('file:///C:/path/');         // Correct:   C:\path\ (Windows)

   new URL('file://nas/foo.txt').pathname;    // Incorrect: /foo.txt
   fileURLToPath('file://nas/foo.txt');       // Correct:   \\nas\foo.txt (Windows)

   new URL('file:///你好.txt').pathname;      // Incorrect: /%E4%BD%A0%E5%A5%BD.txt
   fileURLToPath('file:///你好.txt');         // Correct:   /你好.txt (POSIX)

   new URL('file:///hello world').pathname;   // Incorrect: /hello%20world
   fileURLToPath('file:///hello world');      // Correct:   /hello world (POSIX)


fileURLToPathBuffer
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   url.fileURLToPathBuffer(url[, options])

*Kind:* function · *Arguments:* 2

Like ``url.fileURLToPath(...)`` except that instead of returning a string representation of the path, a ``Buffer`` is returned. This conversion is helpful when the input URL contains percent-encoded segments that are not valid UTF-8 / Unicode sequences.

**Security Considerations:**

**Arguments**

* ``url`` (*URL | string*): The file URL string or URL object to convert to a path.
* ``options`` (*Object*) — optional


format
~~~~~~

.. code-block:: text

   url.format(URL[, options])

*Kind:* function · *Arguments:* 2

Returns a customizable serialization of a URL ``String`` representation of a WHATWG URL object.

The URL object has both a ``toString()`` method and ``href`` property that return string serializations of the URL. These are not, however, customizable in any way. The ``url.format(URL[, options])`` method allows for basic customization of the output.

**Arguments**

* ``URL`` (*URL*): A WHATWG URL object
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const url = require('node:url');
   const myURL = new URL('https://a:b@測試?abc#foo');

   console.log(myURL.href);
   // Prints https://a:b@xn--g6w251d/?abc#foo

   console.log(myURL.toString());
   // Prints https://a:b@xn--g6w251d/?abc#foo

   console.log(url.format(myURL, { fragment: false, unicode: true, auth: false }));
   // Prints 'https://測試/?abc'


parse
~~~~~

.. code-block:: text

   url.parse()

*Kind:* function · *Arguments:* 0

Type: Runtime

``url.parse()`` accepts URLs with ports that are not numbers. This behavior might result in host name spoofing with unexpected input. These URLs will throw an error in future versions of Node.js, as the WHATWG URL API does already.

This callable takes no arguments.


pathToFileURL
~~~~~~~~~~~~~

.. code-block:: text

   url.pathToFileURL(path[, options])

*Kind:* function · *Arguments:* 2

This function ensures that ``path`` is resolved absolutely, and that the URL control characters are correctly encoded when converting into a File URL.

**Arguments**

* ``path`` (*string*): The path to convert to a File URL.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const { pathToFileURL } = require('node:url');
   new URL(__filename);                  // Incorrect: throws (POSIX)
   new URL(__filename);                  // Incorrect: C:\... (Windows)
   pathToFileURL(__filename);            // Correct:   file:///... (POSIX)
   pathToFileURL(__filename);            // Correct:   file:///C:/... (Windows)

   new URL('/foo#1', 'file:');           // Incorrect: file:///foo#1
   pathToFileURL('/foo#1');              // Correct:   file:///foo%231 (POSIX)

   new URL('/some/path%.c', 'file:');    // Incorrect: file:///some/path%.c
   pathToFileURL('/some/path%.c');       // Correct:   file:///some/path%25.c (POSIX)


resolve
~~~~~~~

.. code-block:: text

   url.resolve(from, to)

*Kind:* function · *Arguments:* 2

The ``url.resolve()`` method resolves a target URL relative to a base URL in a manner similar to that of a web browser resolving an anchor tag.

Because it invokes the deprecated ``url.parse()`` internally, ``url.resolve()`` is itself deprecated.

**Arguments**

* ``from`` (*string*): The base URL to use if ``to`` is a relative URL.
* ``to`` (*string*): The target URL to resolve.

**Example**

.. code-block:: javascript

   const url = require('node:url');
   url.resolve('/one/two/three', 'four');         // '/one/two/four'
   url.resolve('http://example.com/', '/one');    // 'http://example.com/one'
   url.resolve('http://example.com/one', '/two'); // 'http://example.com/two'


resolveObject
~~~~~~~~~~~~~

.. code-block:: text

   url.resolveObject(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


URL
~~~

.. code-block:: text

   URL

*Kind:* class · *Arguments:* 2

Browser-compatible ``URL`` class, implemented by following the WHATWG URL Standard. Examples of parsed URLs may be found in the Standard itself. The ``URL`` class is also available on the global object.

In accordance with browser conventions, all properties of ``URL`` objects are implemented as getters and setters on the class prototype, rather than as data properties on the object itself. Thus, unlike legacy ``urlObject``\ s, using the ``delete`` keyword on any properties of ``URL`` objects (e.g. ``delete myURL.protocol``, ``delete myURL.pathname``, etc) has no effect but will still return ``true``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   console.log(URL === require('node:url').URL); // Prints 'true'.


Url
~~~

.. code-block:: text

   url.Url(...args)

*Kind:* function · *Arguments:* 0

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


URLPattern
~~~~~~~~~~

.. code-block:: text

   URLPattern

*Kind:* class · *Arguments:* 3

The ``URLPattern`` API provides an interface to match URLs or parts of URLs against a pattern.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const myPattern = new URLPattern('https://example.com/docs/latest/api/*.html');
   console.log(myPattern.exec('https://example.com/docs/latest/api/dns.html'));
   // Prints:
   // {
   //  "hash": { "groups": {  "0": "" },  "input": "" },
   //  "hostname": { "groups": {}, "input": "example.com" },
   //  "inputs": [
   //    "https://example.com/docs/latest/api/dns.html"
   //  ],
   //  "password": { "groups": { "0": "" }, "input": "" },
   //  "pathname": { "groups": { "0": "dns" }, "input": "/docs/latest/api/dns.html" },
   //  "port": { "groups": {}, "input": "" },
   //  "protocol": { "groups": {}, "input": "https" },
   //  "search": { "groups": { "0": "" }, "input": "" },
   //  "username": { "groups": { "0": "" }, "input": "" }
   // }

   console.log(myPattern.test('https://example.com/docs/latest/api/dns.html'));
   // Prints: true


URLSearchParams
~~~~~~~~~~~~~~~

.. code-block:: text

   URLSearchParams

*Kind:* class · *Arguments:* 1

The ``URLSearchParams`` API provides read and write access to the query of a ``URL``. The ``URLSearchParams`` class can also be used standalone with one of the four following constructors. The ``URLSearchParams`` class is also available on the global object.

The WHATWG ``URLSearchParams`` interface and the ``querystring`` module have similar purpose, but the purpose of the ``querystring`` module is more general, as it allows the customization of delimiter characters (``&`` and ``=``). On the other hand, this API is designed purely for URL query strings.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const myURL = new URL('https://example.org/?abc=123');
   console.log(myURL.searchParams.get('abc'));
   // Prints 123

   myURL.searchParams.append('abc', 'xyz');
   console.log(myURL.href);
   // Prints https://example.org/?abc=123&abc=xyz

   myURL.searchParams.delete('abc');
   myURL.searchParams.set('a', 'b');
   console.log(myURL.href);
   // Prints https://example.org/?a=b

   const newSearchParams = new URLSearchParams(myURL.searchParams);
   // The above is equivalent to
   // const newSearchParams = new URLSearchParams(myURL.search);

   newSearchParams.append('a', 'c');
   console.log(myURL.href);
   // Prints https://example.org/?a=b
   console.log(newSearchParams.toString());
   // Prints a=b&a=c

   // newSearchParams.toString() is implicitly called
   myURL.search = newSearchParams;
   console.log(myURL.href);
   // Prints https://example.org/?a=b&a=c
   newSearchParams.delete('a');
   console.log(myURL.href);
   // Prints https://example.org/?a=b&a=c


urlToHttpOptions
~~~~~~~~~~~~~~~~

.. code-block:: text

   url.urlToHttpOptions(url)

*Kind:* function · *Arguments:* 1

This utility function converts a URL object into an ordinary options object as expected by the ``http.request()`` and ``https.request()`` APIs.

**Arguments**

* ``url`` (*URL*): The WHATWG URL object to convert to an options object.

**Example**

.. code-block:: javascript

   const { urlToHttpOptions } = require('node:url');
   const myURL = new URL('https://a:b@測試?abc#foo');

   console.log(urlToHttpOptions(myURL));
   /*
   {
     protocol: 'https:',
     hostname: 'xn--g6w251d',
     hash: '#foo',
     search: '?abc',
     pathname: '/',
     path: '/?abc',
     href: 'https://a:b@xn--g6w251d/?abc#foo',
     auth: 'a:b'
   }
   */


URL class
---------

**URL** — Browser-compatible ``URL`` class, implemented by following the WHATWG URL Standard. Examples of parsed URLs may be found in the Standard itself. The ``URL`` class is also available on the global object.

In accordance with browser conventions, all properties of ``URL`` objects are implemented as getters and setters on the class prototype, rather than as data properties on the object itself. Thus, unlike legacy ``urlObject``\ s, using the ``delete`` keyword on any properties of ``URL`` objects (e.g. ``delete myURL.protocol``, ``delete myURL.pathname``, etc) has no effect but will still return ``true``.

URL.toJSON
~~~~~~~~~~

.. code-block:: text

   url.toJSON()

*Kind:* function · *Arguments:* 0

The ``toJSON()`` method on the ``URL`` object returns the serialized URL. The value returned is equivalent to that of ``url.href`` and ``url.toString()``.

This method is automatically called when an ``URL`` object is serialized with ``JSON.stringify()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const myURLs = [
     new URL('https://www.example.com'),
     new URL('https://test.example.org'),
   ];
   console.log(JSON.stringify(myURLs));
   // Prints ["https://www.example.com/","https://test.example.org/"]


URL.toString
~~~~~~~~~~~~

.. code-block:: text

   url.toString()

*Kind:* function · *Arguments:* 0

The ``toString()`` method on the ``URL`` object returns the serialized URL. The value returned is equivalent to that of ``url.href`` and ``url.toJSON()``.

This callable takes no arguments.


URLPattern class
----------------

**URLPattern** — The ``URLPattern`` API provides an interface to match URLs or parts of URLs against a pattern.

URLPattern.exec
~~~~~~~~~~~~~~~

.. code-block:: text

   urlPattern.exec(input[, baseURL])

*Kind:* function · *Arguments:* 2

Input can be a string or an object providing the individual URL parts. The object members can be any of ``protocol``, ``username``, ``password``, ``hostname``, ``port``, ``pathname``, ``search``, ``hash`` or ``baseURL``.

If ``baseURL`` is not specified, it will default to ``undefined``.

**Arguments**

* ``input`` (*string | Object*): A URL or URL parts
* ``baseURL`` (*string | undefined*) — optional: A base URL string

**Example**

.. code-block:: javascript

   const myPattern = new URLPattern('https://example.com/docs/latest/api/*.html');
   console.log(myPattern.exec('https://example.com/docs/latest/api/dns.html'));
   // Prints:
   // {
   //  "hash": { "groups": {  "0": "" },  "input": "" },
   //  "hostname": { "groups": {}, "input": "example.com" },
   //  "inputs": [
   //    "https://example.com/docs/latest/api/dns.html"
   //  ],
   //  "password": { "groups": { "0": "" }, "input": "" },
   //  "pathname": { "groups": { "0": "dns" }, "input": "/docs/latest/api/dns.html" },
   //  "port": { "groups": {}, "input": "" },
   //  "protocol": { "groups": {}, "input": "https" },
   //  "search": { "groups": { "0": "" }, "input": "" },
   //  "username": { "groups": { "0": "" }, "input": "" }
   // }


URLPattern.test
~~~~~~~~~~~~~~~

.. code-block:: text

   urlPattern.test(input[, baseURL])

*Kind:* function · *Arguments:* 2

Input can be a string or an object providing the individual URL parts. The object members can be any of ``protocol``, ``username``, ``password``, ``hostname``, ``port``, ``pathname``, ``search``, ``hash`` or ``baseURL``.

If ``baseURL`` is not specified, it will default to ``undefined``.

**Arguments**

* ``input`` (*string | Object*): A URL or URL parts
* ``baseURL`` (*string | undefined*) — optional: A base URL string

**Example**

.. code-block:: javascript

   const myPattern = new URLPattern('https://example.com/docs/latest/api/*.html');
   console.log(myPattern.test('https://example.com/docs/latest/api/dns.html'));
   // Prints: true


URLSearchParams class
---------------------

**URLSearchParams** — The ``URLSearchParams`` API provides read and write access to the query of a ``URL``. The ``URLSearchParams`` class can also be used standalone with one of the four following constructors. The ``URLSearchParams`` class is also available on the global object.

The WHATWG ``URLSearchParams`` interface and the ``querystring`` module have similar purpose, but the purpose of the ``querystring`` module is more general, as it allows the customization of delimiter characters (``&`` and ``=``). On the other hand, this API is designed purely for URL query strings.

URLSearchParams.append
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.append(name, value)

*Kind:* function · *Arguments:* 2

Append a new name-value pair to the query string.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*)


URLSearchParams.delete
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.delete(name[, value])

*Kind:* function · *Arguments:* 2

If ``value`` is provided, removes all name-value pairs where name is ``name`` and value is ``value``..

If ``value`` is not provided, removes all name-value pairs whose name is ``name``.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*) — optional


URLSearchParams.entries
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.entries()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over each of the name-value pairs in the query. Each item of the iterator is a JavaScript ``Array``. The first item of the ``Array`` is the ``name``, the second item of the ``Array`` is the ``value``.

Alias for ``urlSearchParams[Symbol.iterator]()``.

This callable takes no arguments.


URLSearchParams.forEach
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.forEach(fn[, thisArg])

*Kind:* function · *Arguments:* 2

Iterates over each name-value pair in the query and invokes the given function.

**Arguments**

* ``fn`` (*Function*): Invoked for each name-value pair in the query
* ``thisArg`` (*Object*) — optional: To be used as ``this`` value for when ``fn`` is called

**Example**

.. code-block:: javascript

   const myURL = new URL('https://example.org/?a=b&c=d');
   myURL.searchParams.forEach((value, name, searchParams) => {
     console.log(name, value, myURL.searchParams === searchParams);
   });
   // Prints:
   //   a b true
   //   c d true


URLSearchParams.get
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.get(name)

*Kind:* function · *Arguments:* 1

Returns the value of the first name-value pair whose name is ``name``. If there are no such pairs, ``null`` is returned.

**Arguments**

* ``name`` (*string*)


URLSearchParams.getAll
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.getAll(name)

*Kind:* function · *Arguments:* 1

Returns the values of all name-value pairs whose name is ``name``. If there are no such pairs, an empty array is returned.

**Arguments**

* ``name`` (*string*)


URLSearchParams.has
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.has(name[, value])

*Kind:* function · *Arguments:* 2

Checks if the ``URLSearchParams`` object contains key-value pair(s) based on ``name`` and an optional ``value`` argument.

If ``value`` is provided, returns ``true`` when name-value pair with same ``name`` and ``value`` exists.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*) — optional


URLSearchParams.keys
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.keys()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over the names of each name-value pair.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const params = new URLSearchParams('foo=bar&foo=baz');
   for (const name of params.keys()) {
     console.log(name);
   }
   // Prints:
   //   foo
   //   foo


URLSearchParams.set
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.set(name, value)

*Kind:* function · *Arguments:* 2

Sets the value in the ``URLSearchParams`` object associated with ``name`` to ``value``. If there are any pre-existing name-value pairs whose names are ``name``, set the first such pair's value to ``value`` and remove all others. If not, append the name-value pair to the query string.

**Arguments**

* ``name`` (*string*)
* ``value`` (*string*)

**Example**

.. code-block:: javascript

   const params = new URLSearchParams();
   params.append('foo', 'bar');
   params.append('foo', 'baz');
   params.append('abc', 'def');
   console.log(params.toString());
   // Prints foo=bar&foo=baz&abc=def

   params.set('foo', 'def');
   params.set('xyz', 'opq');
   console.log(params.toString());
   // Prints foo=def&abc=def&xyz=opq


URLSearchParams.sort
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.sort()

*Kind:* function · *Arguments:* 0

Sort all existing name-value pairs in-place by their names. Sorting is done with a stable sorting algorithm, so relative order between name-value pairs with the same name is preserved.

This method can be used, in particular, to increase cache hits.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const params = new URLSearchParams('query[]=abc&type=search&query[]=123');
   params.sort();
   console.log(params.toString());
   // Prints query%5B%5D=abc&query%5B%5D=123&type=search


URLSearchParams.toString
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.toString()

*Kind:* function · *Arguments:* 0

Returns the search parameters serialized as a string, with characters percent-encoded where necessary.

This callable takes no arguments.


URLSearchParams.values
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   urlSearchParams.values()

*Kind:* function · *Arguments:* 0

Returns an ES6 ``Iterator`` over the values of each name-value pair.

This callable takes no arguments.


Properties and constants
------------------------

* ``URL.auth`` — string
* ``URL.hash`` — string
* ``URL.host`` — string
* ``URL.hostname`` — string
* ``URL.href`` — string
* ``URL.origin`` — string
* ``URL.password`` — string
* ``URL.path`` — string
* ``URL.pathname`` — string
* ``URL.port`` — string
* ``URL.protocol`` — string
* ``URL.query`` — string
* ``URL.search`` — string
* ``URL.searchParams`` — object
* ``URL.slashes`` — boolean
* ``URL.username`` — string
* ``URLPattern.hasRegExpGroups`` — boolean
* ``URLSearchParams.size`` — number

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
