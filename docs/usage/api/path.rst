path
====

The ``node:path`` module provides utilities for working with file and directory paths. It can be accessed using:

Loading
-------

.. code-block:: javascript

   const path = require('node:path');
   // ESM
   import * as path from 'node:path';

Quick start
-----------

.. code-block:: javascript

   const path = require('node:path');

API
---

basename
~~~~~~~~

.. code-block:: text

   path.basename(path[, suffix])

*Kind:* function · *Arguments:* 2

The ``path.basename()`` method returns the last portion of a ``path``, similar to the Unix ``basename`` command. Trailing directory separators are ignored.

Although Windows usually treats file names, including file extensions, in a case-insensitive manner, this function does not. For example, ``C:\\foo.html`` and ``C:\\foo.HTML`` refer to the same file, but ``basename`` treats the extension as a case-sensitive string:

**Arguments**

* ``path`` (*string*)
* ``suffix`` (*string*) — optional: An optional suffix to remove

**Example**

.. code-block:: javascript

   path.basename('/foo/bar/baz/asdf/quux.html');
   // Returns: 'quux.html'

   path.basename('/foo/bar/baz/asdf/quux.html', '.html');
   // Returns: 'quux'


dirname
~~~~~~~

.. code-block:: text

   path.dirname(path)

*Kind:* function · *Arguments:* 1

The ``path.dirname()`` method returns the directory name of a ``path``, similar to the Unix ``dirname`` command. Trailing directory separators are ignored, see ``path.sep``.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.dirname('/foo/bar/baz/asdf/quux');
   // Returns: '/foo/bar/baz/asdf'


extname
~~~~~~~

.. code-block:: text

   path.extname(path)

*Kind:* function · *Arguments:* 1

The ``path.extname()`` method returns the extension of the ``path``, from the last occurrence of the ``.`` (period) character to end of string in the last portion of the ``path``. If there is no ``.`` in the last portion of the ``path``, or if there are no ``.`` characters other than the first character of the basename of ``path`` (see ``path.basename()``) , an empty string is returned.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.extname('index.html');
   // Returns: '.html'

   path.extname('index.coffee.md');
   // Returns: '.md'

   path.extname('index.');
   // Returns: '.'

   path.extname('index');
   // Returns: ''

   path.extname('.index');
   // Returns: ''

   path.extname('.index.md');
   // Returns: '.md'


format
~~~~~~

.. code-block:: text

   path.format(pathObject)

*Kind:* function · *Arguments:* 1

The ``path.format()`` method returns a path string from an object. This is the opposite of ``path.parse()``.

When providing properties to the ``pathObject`` remember that there are combinations where one property has priority over another:

**Arguments**

* ``pathObject`` (*Object*): Any JavaScript object having the following properties:

**Example**

.. code-block:: javascript

   // If `dir`, `root` and `base` are provided,
   // `${dir}${path.sep}${base}`
   // will be returned. `root` is ignored.
   path.format({
     root: '/ignored',
     dir: '/home/user/dir',
     base: 'file.txt',
   });
   // Returns: '/home/user/dir/file.txt'

   // `root` will be used if `dir` is not specified.
   // If only `root` is provided or `dir` is equal to `root` then the
   // platform separator will not be included. `ext` will be ignored.
   path.format({
     root: '/',
     base: 'file.txt',
     ext: 'ignored',
   });
   // Returns: '/file.txt'

   // `name` + `ext` will be used if `base` is not specified.
   path.format({
     root: '/',
     name: 'file',
     ext: '.txt',
   });
   // Returns: '/file.txt'

   // The dot will be added if it is not specified in `ext`.
   path.format({
     root: '/',
     name: 'file',
     ext: 'txt',
   });
   // Returns: '/file.txt'


isAbsolute
~~~~~~~~~~

.. code-block:: text

   path.isAbsolute(path)

*Kind:* function · *Arguments:* 1

The ``path.isAbsolute()`` method determines if the literal ``path`` is absolute. Therefore, it’s not safe for mitigating path traversals.

If the given ``path`` is a zero-length string, ``false`` will be returned.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.isAbsolute('/foo/bar');   // true
   path.isAbsolute('/baz/..');    // true
   path.isAbsolute('/baz/../..'); // true
   path.isAbsolute('qux/');       // false
   path.isAbsolute('.');          // false


join
~~~~

.. code-block:: text

   path.join([...paths])

*Kind:* function · *Arguments:* 1

The ``path.join()`` method joins all given ``path`` segments together using the platform-specific separator as a delimiter, then normalizes the resulting path.

Zero-length ``path`` segments are ignored. If the joined path string is a zero-length string then ``'.'`` will be returned, representing the current working directory.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.join('/foo', 'bar', 'baz/asdf', 'quux', '..');
   // Returns: '/foo/bar/baz/asdf'

   path.join('foo', {}, 'bar');
   // Throws 'TypeError: Path must be a string. Received {}'


matchesGlob
~~~~~~~~~~~

.. code-block:: text

   path.matchesGlob(path, pattern)

*Kind:* function · *Arguments:* 2

The ``path.matchesGlob()`` method determines if ``path`` matches the ``pattern``.

For example:

**Arguments**

* ``path`` (*string*): The path to glob-match against.
* ``pattern`` (*string*): The glob to check the path against.

**Example**

.. code-block:: javascript

   path.matchesGlob('/foo/bar', '/foo/*'); // true
   path.matchesGlob('/foo/bar*', 'foo/bird'); // false


normalize
~~~~~~~~~

.. code-block:: text

   path.normalize(path)

*Kind:* function · *Arguments:* 1

The ``path.normalize()`` method normalizes the given ``path``, resolving ``'..'`` and ``'.'`` segments.

When multiple, sequential path segment separation characters are found (e.g. ``/`` on POSIX and either ``\`` or ``/`` on Windows), they are replaced by a single instance of the platform-specific path segment separator (``/`` on POSIX and ``\`` on Windows). Trailing separators are preserved.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.normalize('/foo/bar//baz/asdf/quux/..');
   // Returns: '/foo/bar/baz/asdf'


parse
~~~~~

.. code-block:: text

   path.parse(path)

*Kind:* function · *Arguments:* 1

The ``path.parse()`` method returns an object whose properties represent significant elements of the ``path``. Trailing directory separators are ignored, see ``path.sep``.

The returned object will have the following properties:

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.parse('/home/user/dir/file.txt');
   // Returns:
   // { root: '/',
   //   dir: '/home/user/dir',
   //   base: 'file.txt',
   //   ext: '.txt',
   //   name: 'file' }


relative
~~~~~~~~

.. code-block:: text

   path.relative(from, to)

*Kind:* function · *Arguments:* 2

The ``path.relative()`` method returns the relative path from ``from`` to ``to`` based on the current working directory. If ``from`` and ``to`` each resolve to the same path (after calling ``path.resolve()`` on each), a zero-length string is returned.

If a zero-length string is passed as ``from`` or ``to``, the current working directory will be used instead of the zero-length strings.

**Arguments**

* ``from`` (*string*)
* ``to`` (*string*)

**Example**

.. code-block:: javascript

   path.relative('/data/orandea/test/aaa', '/data/orandea/impl/bbb');
   // Returns: '../../impl/bbb'


resolve
~~~~~~~

.. code-block:: text

   path.resolve([...paths])

*Kind:* function · *Arguments:* 1

The ``path.resolve()`` method resolves a sequence of paths or path segments into an absolute path.

The given sequence of paths is processed from right to left, with each subsequent ``path`` prepended until an absolute path is constructed. For instance, given the sequence of path segments: ``/foo``, ``/bar``, ``baz``, calling ``path.resolve('/foo', '/bar', 'baz')`` would return ``/bar/baz`` because ``'baz'`` is not an absolute path but ``'/bar' + '/' + 'baz'`` is.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.resolve('/foo/bar', './baz');
   // Returns: '/foo/bar/baz'

   path.resolve('/foo/bar', '/tmp/file/');
   // Returns: '/tmp/file'

   path.resolve('wwwroot', 'static_files/png/', '../gif/image.gif');
   // If the current working directory is /home/myself/node,
   // this returns '/home/myself/node/wwwroot/static_files/gif/image.gif'


toNamespacedPath
~~~~~~~~~~~~~~~~

.. code-block:: text

   path.toNamespacedPath(path)

*Kind:* function · *Arguments:* 1

On Windows systems only, returns an equivalent namespace-prefixed path for the given ``path``. If ``path`` is not a string, ``path`` will be returned without modifications.

This method is meaningful only on Windows systems. On POSIX systems, the method is non-operational and always returns ``path`` without modifications.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   const path = require('node:path');

   const result = path.toNamespacedPath('./file.txt');
   console.log(result);


posix
-----

**posix** — The ``path.posix`` property provides access to POSIX specific implementations of the ``path`` methods.

The API is accessible via ``require('node:path').posix`` or ``require('node:path/posix')``.

posix.basename
~~~~~~~~~~~~~~

.. code-block:: text

   path.basename(path[, suffix])

*Kind:* function · *Arguments:* 2

The ``path.basename()`` method returns the last portion of a ``path``, similar to the Unix ``basename`` command. Trailing directory separators are ignored.

Although Windows usually treats file names, including file extensions, in a case-insensitive manner, this function does not. For example, ``C:\\foo.html`` and ``C:\\foo.HTML`` refer to the same file, but ``basename`` treats the extension as a case-sensitive string:

**Arguments**

* ``path`` (*string*)
* ``suffix`` (*string*) — optional: An optional suffix to remove

**Example**

.. code-block:: javascript

   path.basename('/foo/bar/baz/asdf/quux.html');
   // Returns: 'quux.html'

   path.basename('/foo/bar/baz/asdf/quux.html', '.html');
   // Returns: 'quux'


posix.dirname
~~~~~~~~~~~~~

.. code-block:: text

   path.dirname(path)

*Kind:* function · *Arguments:* 1

The ``path.dirname()`` method returns the directory name of a ``path``, similar to the Unix ``dirname`` command. Trailing directory separators are ignored, see ``path.sep``.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.dirname('/foo/bar/baz/asdf/quux');
   // Returns: '/foo/bar/baz/asdf'


posix.extname
~~~~~~~~~~~~~

.. code-block:: text

   path.extname(path)

*Kind:* function · *Arguments:* 1

The ``path.extname()`` method returns the extension of the ``path``, from the last occurrence of the ``.`` (period) character to end of string in the last portion of the ``path``. If there is no ``.`` in the last portion of the ``path``, or if there are no ``.`` characters other than the first character of the basename of ``path`` (see ``path.basename()``) , an empty string is returned.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.extname('index.html');
   // Returns: '.html'

   path.extname('index.coffee.md');
   // Returns: '.md'

   path.extname('index.');
   // Returns: '.'

   path.extname('index');
   // Returns: ''

   path.extname('.index');
   // Returns: ''

   path.extname('.index.md');
   // Returns: '.md'


posix.format
~~~~~~~~~~~~

.. code-block:: text

   path.format(pathObject)

*Kind:* function · *Arguments:* 1

The ``path.format()`` method returns a path string from an object. This is the opposite of ``path.parse()``.

When providing properties to the ``pathObject`` remember that there are combinations where one property has priority over another:

**Arguments**

* ``pathObject`` (*Object*): Any JavaScript object having the following properties:

**Example**

.. code-block:: javascript

   // If `dir`, `root` and `base` are provided,
   // `${dir}${path.sep}${base}`
   // will be returned. `root` is ignored.
   path.format({
     root: '/ignored',
     dir: '/home/user/dir',
     base: 'file.txt',
   });
   // Returns: '/home/user/dir/file.txt'

   // `root` will be used if `dir` is not specified.
   // If only `root` is provided or `dir` is equal to `root` then the
   // platform separator will not be included. `ext` will be ignored.
   path.format({
     root: '/',
     base: 'file.txt',
     ext: 'ignored',
   });
   // Returns: '/file.txt'

   // `name` + `ext` will be used if `base` is not specified.
   path.format({
     root: '/',
     name: 'file',
     ext: '.txt',
   });
   // Returns: '/file.txt'

   // The dot will be added if it is not specified in `ext`.
   path.format({
     root: '/',
     name: 'file',
     ext: 'txt',
   });
   // Returns: '/file.txt'


posix.isAbsolute
~~~~~~~~~~~~~~~~

.. code-block:: text

   path.isAbsolute(path)

*Kind:* function · *Arguments:* 1

The ``path.isAbsolute()`` method determines if the literal ``path`` is absolute. Therefore, it’s not safe for mitigating path traversals.

If the given ``path`` is a zero-length string, ``false`` will be returned.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.isAbsolute('/foo/bar');   // true
   path.isAbsolute('/baz/..');    // true
   path.isAbsolute('/baz/../..'); // true
   path.isAbsolute('qux/');       // false
   path.isAbsolute('.');          // false


posix.join
~~~~~~~~~~

.. code-block:: text

   path.join([...paths])

*Kind:* function · *Arguments:* 1

The ``path.join()`` method joins all given ``path`` segments together using the platform-specific separator as a delimiter, then normalizes the resulting path.

Zero-length ``path`` segments are ignored. If the joined path string is a zero-length string then ``'.'`` will be returned, representing the current working directory.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.join('/foo', 'bar', 'baz/asdf', 'quux', '..');
   // Returns: '/foo/bar/baz/asdf'

   path.join('foo', {}, 'bar');
   // Throws 'TypeError: Path must be a string. Received {}'


posix.matchesGlob
~~~~~~~~~~~~~~~~~

.. code-block:: text

   path.matchesGlob(path, pattern)

*Kind:* function · *Arguments:* 2

The ``path.matchesGlob()`` method determines if ``path`` matches the ``pattern``.

For example:

**Arguments**

* ``path`` (*string*): The path to glob-match against.
* ``pattern`` (*string*): The glob to check the path against.

**Example**

.. code-block:: javascript

   path.matchesGlob('/foo/bar', '/foo/*'); // true
   path.matchesGlob('/foo/bar*', 'foo/bird'); // false


posix.normalize
~~~~~~~~~~~~~~~

.. code-block:: text

   path.normalize(path)

*Kind:* function · *Arguments:* 1

The ``path.normalize()`` method normalizes the given ``path``, resolving ``'..'`` and ``'.'`` segments.

When multiple, sequential path segment separation characters are found (e.g. ``/`` on POSIX and either ``\`` or ``/`` on Windows), they are replaced by a single instance of the platform-specific path segment separator (``/`` on POSIX and ``\`` on Windows). Trailing separators are preserved.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.normalize('/foo/bar//baz/asdf/quux/..');
   // Returns: '/foo/bar/baz/asdf'


posix.parse
~~~~~~~~~~~

.. code-block:: text

   path.parse(path)

*Kind:* function · *Arguments:* 1

The ``path.parse()`` method returns an object whose properties represent significant elements of the ``path``. Trailing directory separators are ignored, see ``path.sep``.

The returned object will have the following properties:

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.parse('/home/user/dir/file.txt');
   // Returns:
   // { root: '/',
   //   dir: '/home/user/dir',
   //   base: 'file.txt',
   //   ext: '.txt',
   //   name: 'file' }


posix.relative
~~~~~~~~~~~~~~

.. code-block:: text

   path.relative(from, to)

*Kind:* function · *Arguments:* 2

The ``path.relative()`` method returns the relative path from ``from`` to ``to`` based on the current working directory. If ``from`` and ``to`` each resolve to the same path (after calling ``path.resolve()`` on each), a zero-length string is returned.

If a zero-length string is passed as ``from`` or ``to``, the current working directory will be used instead of the zero-length strings.

**Arguments**

* ``from`` (*string*)
* ``to`` (*string*)

**Example**

.. code-block:: javascript

   path.relative('/data/orandea/test/aaa', '/data/orandea/impl/bbb');
   // Returns: '../../impl/bbb'


posix.resolve
~~~~~~~~~~~~~

.. code-block:: text

   path.resolve([...paths])

*Kind:* function · *Arguments:* 1

The ``path.resolve()`` method resolves a sequence of paths or path segments into an absolute path.

The given sequence of paths is processed from right to left, with each subsequent ``path`` prepended until an absolute path is constructed. For instance, given the sequence of path segments: ``/foo``, ``/bar``, ``baz``, calling ``path.resolve('/foo', '/bar', 'baz')`` would return ``/bar/baz`` because ``'baz'`` is not an absolute path but ``'/bar' + '/' + 'baz'`` is.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.resolve('/foo/bar', './baz');
   // Returns: '/foo/bar/baz'

   path.resolve('/foo/bar', '/tmp/file/');
   // Returns: '/tmp/file'

   path.resolve('wwwroot', 'static_files/png/', '../gif/image.gif');
   // If the current working directory is /home/myself/node,
   // this returns '/home/myself/node/wwwroot/static_files/gif/image.gif'


posix.toNamespacedPath
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   path.toNamespacedPath(path)

*Kind:* function · *Arguments:* 1

On Windows systems only, returns an equivalent namespace-prefixed path for the given ``path``. If ``path`` is not a string, ``path`` will be returned without modifications.

This method is meaningful only on Windows systems. On POSIX systems, the method is non-operational and always returns ``path`` without modifications.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   const path = require('node:path');

   const result = path.posix.toNamespacedPath('./file.txt');
   console.log(result);


win32
-----

**win32** — The ``path.win32`` property provides access to Windows-specific implementations of the ``path`` methods.

The API is accessible via ``require('node:path').win32`` or ``require('node:path/win32')``.

win32.basename
~~~~~~~~~~~~~~

.. code-block:: text

   path.basename(path[, suffix])

*Kind:* function · *Arguments:* 2

The ``path.basename()`` method returns the last portion of a ``path``, similar to the Unix ``basename`` command. Trailing directory separators are ignored.

Although Windows usually treats file names, including file extensions, in a case-insensitive manner, this function does not. For example, ``C:\\foo.html`` and ``C:\\foo.HTML`` refer to the same file, but ``basename`` treats the extension as a case-sensitive string:

**Arguments**

* ``path`` (*string*)
* ``suffix`` (*string*) — optional: An optional suffix to remove

**Example**

.. code-block:: javascript

   path.basename('/foo/bar/baz/asdf/quux.html');
   // Returns: 'quux.html'

   path.basename('/foo/bar/baz/asdf/quux.html', '.html');
   // Returns: 'quux'


win32.dirname
~~~~~~~~~~~~~

.. code-block:: text

   path.dirname(path)

*Kind:* function · *Arguments:* 1

The ``path.dirname()`` method returns the directory name of a ``path``, similar to the Unix ``dirname`` command. Trailing directory separators are ignored, see ``path.sep``.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.dirname('/foo/bar/baz/asdf/quux');
   // Returns: '/foo/bar/baz/asdf'


win32.extname
~~~~~~~~~~~~~

.. code-block:: text

   path.extname(path)

*Kind:* function · *Arguments:* 1

The ``path.extname()`` method returns the extension of the ``path``, from the last occurrence of the ``.`` (period) character to end of string in the last portion of the ``path``. If there is no ``.`` in the last portion of the ``path``, or if there are no ``.`` characters other than the first character of the basename of ``path`` (see ``path.basename()``) , an empty string is returned.

A ``TypeError`` is thrown if ``path`` is not a string.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.extname('index.html');
   // Returns: '.html'

   path.extname('index.coffee.md');
   // Returns: '.md'

   path.extname('index.');
   // Returns: '.'

   path.extname('index');
   // Returns: ''

   path.extname('.index');
   // Returns: ''

   path.extname('.index.md');
   // Returns: '.md'


win32.format
~~~~~~~~~~~~

.. code-block:: text

   path.format(pathObject)

*Kind:* function · *Arguments:* 1

The ``path.format()`` method returns a path string from an object. This is the opposite of ``path.parse()``.

When providing properties to the ``pathObject`` remember that there are combinations where one property has priority over another:

**Arguments**

* ``pathObject`` (*Object*): Any JavaScript object having the following properties:

**Example**

.. code-block:: javascript

   // If `dir`, `root` and `base` are provided,
   // `${dir}${path.sep}${base}`
   // will be returned. `root` is ignored.
   path.format({
     root: '/ignored',
     dir: '/home/user/dir',
     base: 'file.txt',
   });
   // Returns: '/home/user/dir/file.txt'

   // `root` will be used if `dir` is not specified.
   // If only `root` is provided or `dir` is equal to `root` then the
   // platform separator will not be included. `ext` will be ignored.
   path.format({
     root: '/',
     base: 'file.txt',
     ext: 'ignored',
   });
   // Returns: '/file.txt'

   // `name` + `ext` will be used if `base` is not specified.
   path.format({
     root: '/',
     name: 'file',
     ext: '.txt',
   });
   // Returns: '/file.txt'

   // The dot will be added if it is not specified in `ext`.
   path.format({
     root: '/',
     name: 'file',
     ext: 'txt',
   });
   // Returns: '/file.txt'


win32.isAbsolute
~~~~~~~~~~~~~~~~

.. code-block:: text

   path.isAbsolute(path)

*Kind:* function · *Arguments:* 1

The ``path.isAbsolute()`` method determines if the literal ``path`` is absolute. Therefore, it’s not safe for mitigating path traversals.

If the given ``path`` is a zero-length string, ``false`` will be returned.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.isAbsolute('/foo/bar');   // true
   path.isAbsolute('/baz/..');    // true
   path.isAbsolute('/baz/../..'); // true
   path.isAbsolute('qux/');       // false
   path.isAbsolute('.');          // false


win32.join
~~~~~~~~~~

.. code-block:: text

   path.join([...paths])

*Kind:* function · *Arguments:* 1

The ``path.join()`` method joins all given ``path`` segments together using the platform-specific separator as a delimiter, then normalizes the resulting path.

Zero-length ``path`` segments are ignored. If the joined path string is a zero-length string then ``'.'`` will be returned, representing the current working directory.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.join('/foo', 'bar', 'baz/asdf', 'quux', '..');
   // Returns: '/foo/bar/baz/asdf'

   path.join('foo', {}, 'bar');
   // Throws 'TypeError: Path must be a string. Received {}'


win32.matchesGlob
~~~~~~~~~~~~~~~~~

.. code-block:: text

   path.matchesGlob(path, pattern)

*Kind:* function · *Arguments:* 2

The ``path.matchesGlob()`` method determines if ``path`` matches the ``pattern``.

For example:

**Arguments**

* ``path`` (*string*): The path to glob-match against.
* ``pattern`` (*string*): The glob to check the path against.

**Example**

.. code-block:: javascript

   path.matchesGlob('/foo/bar', '/foo/*'); // true
   path.matchesGlob('/foo/bar*', 'foo/bird'); // false


win32.normalize
~~~~~~~~~~~~~~~

.. code-block:: text

   path.normalize(path)

*Kind:* function · *Arguments:* 1

The ``path.normalize()`` method normalizes the given ``path``, resolving ``'..'`` and ``'.'`` segments.

When multiple, sequential path segment separation characters are found (e.g. ``/`` on POSIX and either ``\`` or ``/`` on Windows), they are replaced by a single instance of the platform-specific path segment separator (``/`` on POSIX and ``\`` on Windows). Trailing separators are preserved.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.normalize('/foo/bar//baz/asdf/quux/..');
   // Returns: '/foo/bar/baz/asdf'


win32.parse
~~~~~~~~~~~

.. code-block:: text

   path.parse(path)

*Kind:* function · *Arguments:* 1

The ``path.parse()`` method returns an object whose properties represent significant elements of the ``path``. Trailing directory separators are ignored, see ``path.sep``.

The returned object will have the following properties:

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   path.parse('/home/user/dir/file.txt');
   // Returns:
   // { root: '/',
   //   dir: '/home/user/dir',
   //   base: 'file.txt',
   //   ext: '.txt',
   //   name: 'file' }


win32.relative
~~~~~~~~~~~~~~

.. code-block:: text

   path.relative(from, to)

*Kind:* function · *Arguments:* 2

The ``path.relative()`` method returns the relative path from ``from`` to ``to`` based on the current working directory. If ``from`` and ``to`` each resolve to the same path (after calling ``path.resolve()`` on each), a zero-length string is returned.

If a zero-length string is passed as ``from`` or ``to``, the current working directory will be used instead of the zero-length strings.

**Arguments**

* ``from`` (*string*)
* ``to`` (*string*)

**Example**

.. code-block:: javascript

   path.relative('/data/orandea/test/aaa', '/data/orandea/impl/bbb');
   // Returns: '../../impl/bbb'


win32.resolve
~~~~~~~~~~~~~

.. code-block:: text

   path.resolve([...paths])

*Kind:* function · *Arguments:* 1

The ``path.resolve()`` method resolves a sequence of paths or path segments into an absolute path.

The given sequence of paths is processed from right to left, with each subsequent ``path`` prepended until an absolute path is constructed. For instance, given the sequence of path segments: ``/foo``, ``/bar``, ``baz``, calling ``path.resolve('/foo', '/bar', 'baz')`` would return ``/bar/baz`` because ``'baz'`` is not an absolute path but ``'/bar' + '/' + 'baz'`` is.

**Arguments**

* ``...paths`` — optional

**Example**

.. code-block:: javascript

   path.resolve('/foo/bar', './baz');
   // Returns: '/foo/bar/baz'

   path.resolve('/foo/bar', '/tmp/file/');
   // Returns: '/tmp/file'

   path.resolve('wwwroot', 'static_files/png/', '../gif/image.gif');
   // If the current working directory is /home/myself/node,
   // this returns '/home/myself/node/wwwroot/static_files/gif/image.gif'


win32.toNamespacedPath
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   path.toNamespacedPath(path)

*Kind:* function · *Arguments:* 1

On Windows systems only, returns an equivalent namespace-prefixed path for the given ``path``. If ``path`` is not a string, ``path`` will be returned without modifications.

This method is meaningful only on Windows systems. On POSIX systems, the method is non-operational and always returns ``path`` without modifications.

**Arguments**

* ``path`` (*string*)

**Example**

.. code-block:: javascript

   const path = require('node:path');

   const result = path.win32.toNamespacedPath('./file.txt');
   console.log(result);


Properties and constants
------------------------

* ``delimiter`` — string
* ``posix`` — object
* ``posix.delimiter`` — string
* ``posix.sep`` — string
* ``sep`` — string
* ``win32`` — object
* ``win32.delimiter`` — string
* ``win32.sep`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
