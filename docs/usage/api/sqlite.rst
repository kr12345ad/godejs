sqlite
======

The ``node:sqlite`` module facilitates working with SQLite databases. To access it:

This module is only available under the ``node:`` scheme. SQL trace events can be observed via the ``diagnostics_channel`` module. See ``'sqlite.db.query'`` for details.

Loading
-------

.. code-block:: javascript

   const sqlite = require('node:sqlite');
   // ESM
   import * as sqlite from 'node:sqlite';

Quick start
-----------

.. code-block:: javascript

   const sqlite = require('node:sqlite');

API
---

backup
~~~~~~

.. code-block:: text

   sqlite.backup(sourceDb, path[, options])

*Kind:* function · *Arguments:* 3

This method makes a database backup. This method abstracts the ``sqlite3_backup_init()``, ``sqlite3_backup_step()`` and ``sqlite3_backup_finish()`` functions.

The backed-up database can be used normally during the backup process. Mutations coming from the same connection - same ``DatabaseSync`` - object will be reflected in the backup right away. However, mutations from other connections will cause the backup process to restart.

**Arguments**

* ``sourceDb`` (*DatabaseSync*): The database to backup. The source database must be open.
* ``path`` (*string | Buffer | URL*): The path where the backup will be created. If the file already exists,
* ``options`` (*Object*) — optional: Optional configuration for the backup. The

**Example**

.. code-block:: javascript

   const { backup, DatabaseSync } = require('node:sqlite');

   (async () => {
     const sourceDb = new DatabaseSync('source.db');
     const totalPagesTransferred = await backup(sourceDb, 'backup.db', {
       rate: 1, // Copy one page at a time.
       progress: ({ totalPages, remainingPages }) => {
         console.log('Backup in progress', { totalPages, remainingPages });
       },
     });

     console.log('Backup completed', totalPagesTransferred);
   })();


DatabaseSync
~~~~~~~~~~~~

.. code-block:: text

   DatabaseSync

*Kind:* class · *Arguments:* 2

This class represents a single connection to a SQLite database. All APIs exposed by this class execute synchronously.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');

   const db = new DatabaseSync(':memory:');
   db.exec(`
     CREATE TABLE t3(x, y);
     INSERT INTO t3 VALUES ('a', 4),
                           ('b', 5),
                           ('c', 3),
                           ('d', 8),
                           ('e', 1);
   `);

   db.aggregate('sumint', {
     start: 0,
     step: (acc, value) => acc + value,
   });

   using query = db.prepare('SELECT sumint(y) as total FROM t3');
   query.get(); // { total: 21 }


Session
~~~~~~~

.. code-block:: text

   Session

*Kind:* class · *Arguments:* 2

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.Session(undefined, undefined);
   console.log(instance);


StatementSync
~~~~~~~~~~~~~

.. code-block:: text

   StatementSync

*Kind:* class · *Arguments:* 3

This class represents a single prepared statement. This class cannot be instantiated via its constructor. Instead, instances are created via the ``database.prepare()`` method. All APIs exposed by this class execute synchronously.

A prepared statement is an efficient binary representation of the SQL used to create it. Prepared statements are parameterizable, and can be invoked multiple times with different bound values. Parameters also offer protection against SQL injection attacks. For these reasons, prepared statements are preferred over hand-crafted SQL strings when handling user input.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   db.prepare('SELECT ? AS a, ? AS b').get('x', 42);
   // { a: 'x', b: 42 }
   db.prepare('SELECT ?2 AS a, ?1 AS b').get('first', 'second');
   // { a: 'second', b: 'first' }


DatabaseSync class
------------------

**DatabaseSync** — This class represents a single connection to a SQLite database. All APIs exposed by this class execute synchronously.

DatabaseSync.aggregate
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.aggregate(name, options)

*Kind:* function · *Arguments:* 2

Registers a new aggregate function with the SQLite database. This method is a wrapper around ``sqlite3_create_window_function()``.

When used as a window function, the ``result`` function will be called multiple times.

**Arguments**

* ``name`` (*string*): The name of the SQLite function to create.
* ``options`` (*Object*): Function configuration settings.

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');

   const db = new DatabaseSync(':memory:');
   db.exec(`
     CREATE TABLE t3(x, y);
     INSERT INTO t3 VALUES ('a', 4),
                           ('b', 5),
                           ('c', 3),
                           ('d', 8),
                           ('e', 1);
   `);

   db.aggregate('sumint', {
     start: 0,
     step: (acc, value) => acc + value,
   });

   using query = db.prepare('SELECT sumint(y) as total FROM t3');
   query.get(); // { total: 21 }


DatabaseSync.applyChangeset
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.applyChangeset(changeset[, options])

*Kind:* function · *Arguments:* 2

An exception is thrown if the database is not open. This method is a wrapper around ``sqlite3changeset_apply()``.

**Arguments**

* ``changeset`` (*Uint8Array*): A binary changeset or patchset.
* ``options`` (*Object*) — optional: The configuration options for how the changes will be applied.

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');

   const sourceDb = new DatabaseSync(':memory:');
   const targetDb = new DatabaseSync(':memory:');

   sourceDb.exec('CREATE TABLE data(key INTEGER PRIMARY KEY, value TEXT)');
   targetDb.exec('CREATE TABLE data(key INTEGER PRIMARY KEY, value TEXT)');

   const session = sourceDb.createSession();

   using insert = sourceDb.prepare('INSERT INTO data (key, value) VALUES (?, ?)');
   insert.run(1, 'hello');
   insert.run(2, 'world');

   const changeset = session.changeset();
   targetDb.applyChangeset(changeset);
   // Now that the changeset has been applied, targetDb contains the same data as sourceDb.


DatabaseSync.close
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.close()

*Kind:* function · *Arguments:* 0

Closes the database connection. An exception is thrown if the database is not open. An ``ERR_INVALID_STATE`` error is thrown if the method is called while a statement is executing, such as inside a user-defined function, an aggregate function, an authorizer callback, or a ``'sqlite.db.query'`` subscriber. This method is a wrapper around ``sqlite3_close_v2()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.close();
   console.log(result);


DatabaseSync.createSession
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.createSession([options])

*Kind:* function · *Arguments:* 1

Creates and attaches a session to the database. This method is a wrapper around ``sqlite3session_create()`` and ``sqlite3session_attach()``.

**Arguments**

* ``options`` (*Object*) — optional: The configuration options for the session.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.createSession({});
   console.log(result);


DatabaseSync.deserialize
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.deserialize(buffer[, options])

*Kind:* function · *Arguments:* 2

Loads a serialized database into this connection, replacing the current database. The deserialized database is writable. Existing prepared statements are finalized before deserialization is attempted, even if the operation subsequently fails. An ``ERR_INVALID_STATE`` error is thrown if the method is called while a database callback is on the stack, for example a user-defined function, an aggregate function, an authorizer, or a changeset filter or conflict handler. This method is a wrapper around ``sqlite3_deserialize()``.

**Arguments**

* ``buffer`` (*Uint8Array*): A binary representation of a database, such as the
* ``options`` (*Object*) — optional: Optional configuration for the deserialization.

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');

   const original = new DatabaseSync(':memory:');
   original.exec('CREATE TABLE t(key INTEGER PRIMARY KEY, value TEXT)');
   original.exec("INSERT INTO t VALUES (1, 'hello')");
   const buffer = original.serialize();
   original.close();

   const clone = new DatabaseSync(':memory:');
   clone.deserialize(buffer);
   using query = clone.prepare('SELECT value FROM t');
   console.log(query.get());
   // Prints: { value: 'hello' }


DatabaseSync.enableLoadExtension
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.enableLoadExtension(allow)

*Kind:* function · *Arguments:* 1

Enables or disables the ``loadExtension`` SQL function, and the ``loadExtension()`` method. When ``allowExtension`` is ``false`` when constructing, you cannot enable loading extensions for security reasons.

**Arguments**

* ``allow`` (*boolean*): Whether to allow loading extensions.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.enableLoadExtension(true);
   console.log(result);


DatabaseSync.exec
~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.exec(sql)

*Kind:* function · *Arguments:* 1

This method allows one or more SQL statements to be executed without returning any results. This method is useful when executing SQL statements read from a file. This method is a wrapper around ``sqlite3_exec()``.

**Arguments**

* ``sql`` (*string*): A SQL string to execute.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.exec('value');
   console.log(result);


DatabaseSync.function
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.function(name[, options], fn)

*Kind:* function · *Arguments:* 3

This method is used to create SQLite user-defined functions. This method is a wrapper around ``sqlite3_create_function_v2()``.

**Arguments**

* ``name`` (*string*): The name of the SQLite function to create.
* ``options`` (*Object*) — optional: Optional configuration settings for the function. The
* ``fn`` (*Function*): The JavaScript function to call when the SQLite function is

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.function('example', {}, () => {});
   console.log(result);


DatabaseSync.isOpen
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.isOpen

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.isOpen();
   console.log(result);


DatabaseSync.loadExtension
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.loadExtension(path[, entryPoint])

*Kind:* function · *Arguments:* 2

Loads a shared library into the database connection. This method is a wrapper around ``sqlite3_load_extension()``. It is required to enable the ``allowExtension`` option when constructing the ``DatabaseSync`` instance.

**Arguments**

* ``path`` (*string*): The path to the shared library to load.
* ``entryPoint`` (*string*) — optional: The name of the extension's entry-point function. When

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');
   const database = new DatabaseSync(':memory:', { allowExtension: true });

   // Load using the entry point derived from the filename.
   database.loadExtension('./decimal.dylib');

   // Override the entry point when the derived name does not match.
   database.loadExtension('./base64.dylib', 'sqlite3_base64_init');


DatabaseSync.location
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.location([dbName])

*Kind:* function · *Arguments:* 1

This method is a wrapper around ``sqlite3_db_filename()``

**Arguments**

* ``dbName`` (*string*) — optional: Name of the database. This can be ``'main'`` (the default primary database) or any other

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.location('value');
   console.log(result);


DatabaseSync.open
~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.open()

*Kind:* function · *Arguments:* 0

Opens the database specified in the ``path`` argument of the ``DatabaseSync`` constructor. This method should only be used when the database is not opened via the constructor. An exception is thrown if the database is already open.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.open();
   console.log(result);


DatabaseSync.prepare
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.prepare(sql[, options])

*Kind:* function · *Arguments:* 2

Compiles a SQL statement into a prepared statement. This method is a wrapper around ``sqlite3_prepare_v2()``.

**Arguments**

* ``sql`` (*string*): A SQL string to compile to a prepared statement.
* ``options`` (*Object*) — optional: Optional configuration for the prepared statement.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.DatabaseSync();
   const result = instance.prepare('value', {});
   console.log(result);


DatabaseSync.serialize
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.serialize([dbName])

*Kind:* function · *Arguments:* 1

Serializes the database into a binary representation, returned as a ``Uint8Array``. This is useful for saving, cloning, or transferring an in-memory database. This method is a wrapper around ``sqlite3_serialize()``.

**Arguments**

* ``dbName`` (*string*) — optional: Name of the database to serialize. This can be ``'main'``

**Example**

.. code-block:: javascript

   const { DatabaseSync } = require('node:sqlite');

   const db = new DatabaseSync(':memory:');
   db.exec('CREATE TABLE t(key INTEGER PRIMARY KEY, value TEXT)');
   db.exec("INSERT INTO t VALUES (1, 'hello')");
   const buffer = db.serialize();
   console.log(buffer.length); // Prints the byte length of the database


DatabaseSync.setAuthorizer
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.setAuthorizer(callback)

*Kind:* function · *Arguments:* 1

Sets an authorizer callback that SQLite will invoke whenever it attempts to access data or modify the database schema through prepared statements. This can be used to implement security policies, audit access, or restrict certain operations. This method is a wrapper around ``sqlite3_set_authorizer()``.

When invoked, the callback receives five arguments:

**Arguments**

* ``callback`` (*Function|null*): The authorizer function to set, or ``null`` to

**Example**

.. code-block:: javascript

   const { DatabaseSync, constants } = require('node:sqlite');
   const db = new DatabaseSync(':memory:');

   // Set up an authorizer that denies all table creation
   db.setAuthorizer((actionCode) => {
     if (actionCode === constants.SQLITE_CREATE_TABLE) {
       return constants.SQLITE_DENY;
     }
     return constants.SQLITE_OK;
   });

   // This will work
   using query = db.prepare('SELECT 1');
   query.get();

   // This will throw an error due to authorization denial
   try {
     db.exec('CREATE TABLE blocked (id INTEGER)');
   } catch (err) {
     console.log('Operation blocked:', err.message);
   }


Session class
-------------

Session.changeset
~~~~~~~~~~~~~~~~~

.. code-block:: text

   session.changeset()

*Kind:* function · *Arguments:* 0

Retrieves a changeset containing all changes since the changeset was created. Can be called multiple times. An exception is thrown if the database or the session is not open. This method is a wrapper around ``sqlite3session_changeset()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.Session();
   const result = instance.changeset();
   console.log(result);


Session.close
~~~~~~~~~~~~~

.. code-block:: text

   session.close()

*Kind:* function · *Arguments:* 0

Closes the session. An exception is thrown if the database or the session is not open, or if the session is currently generating a changeset or patchset. This method is a wrapper around ``sqlite3session_delete()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.Session();
   const result = instance.close();
   console.log(result);


Session.patchset
~~~~~~~~~~~~~~~~

.. code-block:: text

   session.patchset()

*Kind:* function · *Arguments:* 0

Similar to the method above, but generates a more compact patchset. See Changesets and Patchsets in the documentation of SQLite. An exception is thrown if the database or the session is not open. This method is a wrapper around ``sqlite3session_patchset()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.Session();
   const result = instance.patchset();
   console.log(result);


StatementSync class
-------------------

**StatementSync** — This class represents a single prepared statement. This class cannot be instantiated via its constructor. Instead, instances are created via the ``database.prepare()`` method. All APIs exposed by this class execute synchronously.

A prepared statement is an efficient binary representation of the SQL used to create it. Prepared statements are parameterizable, and can be invoked multiple times with different bound values. Parameters also offer protection against SQL injection attacks. For these reasons, prepared statements are preferred over hand-crafted SQL strings when handling user input.

StatementSync.all
~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.all([namedParameters][, ...anonymousParameters])

*Kind:* function · *Arguments:* 2

This method executes a prepared statement and returns all results as an array of objects. If the prepared statement does not return any results, this method returns an empty array. The prepared statement parameters are bound using the values in ``namedParameters`` and ``anonymousParameters``. See Binding parameters.

**Arguments**

* ``namedParameters`` (*Object*) — optional: An optional object used to bind named parameters.
* ``...anonymousParameters`` — optional

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.all({}, undefined);
   console.log(result);


StatementSync.close
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   database.close()

*Kind:* function · *Arguments:* 0

Closes the database connection. An exception is thrown if the database is not open. An ``ERR_INVALID_STATE`` error is thrown if the method is called while a statement is executing, such as inside a user-defined function, an aggregate function, an authorizer callback, or a ``'sqlite.db.query'`` subscriber. This method is a wrapper around ``sqlite3_close_v2()``.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.close();
   console.log(result);


StatementSync.columns
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.columns()

*Kind:* function · *Arguments:* 0

This method is used to retrieve information about the columns returned by the prepared statement.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.columns();
   console.log(result);


StatementSync.get
~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.get([namedParameters][, ...anonymousParameters])

*Kind:* function · *Arguments:* 2

This method executes a prepared statement and returns the first result as an object. If the prepared statement does not return any results, this method returns ``undefined``. The prepared statement parameters are bound using the values in ``namedParameters`` and ``anonymousParameters``. See Binding parameters.

**Arguments**

* ``namedParameters`` (*Object*) — optional: An optional object used to bind named parameters.
* ``...anonymousParameters`` — optional

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.get({}, undefined);
   console.log(result);


StatementSync.iterate
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.iterate([namedParameters][, ...anonymousParameters])

*Kind:* function · *Arguments:* 2

This method executes a prepared statement and returns an iterator of objects. If the prepared statement does not return any results, this method returns an empty iterator. The prepared statement parameters are bound using the values in ``namedParameters`` and ``anonymousParameters``. See Binding parameters.

**Arguments**

* ``namedParameters`` (*Object*) — optional: An optional object used to bind named parameters.
* ``...anonymousParameters`` — optional

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.iterate({}, undefined);
   console.log(result);


StatementSync.run
~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.run([namedParameters][, ...anonymousParameters])

*Kind:* function · *Arguments:* 2

This method executes a prepared statement and returns an object summarizing the resulting changes. The prepared statement parameters are bound using the values in ``namedParameters`` and ``anonymousParameters``. See Binding parameters.

**Arguments**

* ``namedParameters`` (*Object*) — optional: An optional object used to bind named parameters.
* ``...anonymousParameters`` — optional

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.run({}, undefined);
   console.log(result);


StatementSync.setAllowBareNamedParameters
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.setAllowBareNamedParameters(enabled)

*Kind:* function · *Arguments:* 1

The names of SQLite parameters begin with a prefix character. However, with the exception of the dollar sign character, these prefix characters also require extra quoting when used in object keys.

To improve ergonomics, ``node:sqlite`` allows bare named parameters, which do not require the prefix character in JavaScript code, by default. This method can be used to disable that behavior, requiring the prefix character when binding. There are several caveats to be aware of when bare named parameters are allowed:

**Arguments**

* ``enabled`` (*boolean*): Enables or disables support for binding named parameters

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.setAllowBareNamedParameters(true);
   console.log(result);


StatementSync.setReadBigInts
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.setReadBigInts(enabled)

*Kind:* function · *Arguments:* 1

When reading from the database, SQLite ``INTEGER``\ s are mapped to JavaScript numbers by default. However, SQLite ``INTEGER``\ s can store values larger than JavaScript numbers are capable of representing. In such cases, this method can be used to read ``INTEGER`` data using JavaScript ``BigInt``\ s. This method has no impact on database write operations where numbers and ``BigInt``\ s are both supported at all times.

**Arguments**

* ``enabled`` (*boolean*): Enables or disables the use of ``BigInt``\ s when reading

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.setReadBigInts(true);
   console.log(result);


StatementSync.setReturnArrays
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   statement.setReturnArrays(enabled)

*Kind:* function · *Arguments:* 1

When enabled, query results returned by the ``all()``, ``get()``, and ``iterate()`` methods will be returned as arrays instead of objects.

**Arguments**

* ``enabled`` (*boolean*): Enables or disables the return of query results as arrays.

**Example**

.. code-block:: javascript

   const sqlite = require('node:sqlite');

   const instance = new sqlite.StatementSync();
   const result = instance.setReturnArrays(true);
   console.log(result);


Properties and constants
------------------------

* ``constants`` — object
* ``constants.SQLITE_ATTACH`` — number
* ``constants.SQLITE_BUSY`` — number
* ``constants.SQLITE_CHANGESET_ABORT`` — number
* ``constants.SQLITE_CHANGESET_CONFLICT`` — number
* ``constants.SQLITE_CHANGESET_CONSTRAINT`` — number
* ``constants.SQLITE_CHANGESET_DATA`` — number
* ``constants.SQLITE_CHANGESET_FOREIGN_KEY`` — number
* ``constants.SQLITE_CHANGESET_NOTFOUND`` — number
* ``constants.SQLITE_CHANGESET_OMIT`` — number
* ``constants.SQLITE_CHANGESET_REPLACE`` — number
* ``constants.SQLITE_CONSTRAINT`` — number
* ``constants.SQLITE_CREATE_TABLE`` — number
* ``constants.SQLITE_DELETE`` — number
* ``constants.SQLITE_DENY`` — number
* ``constants.SQLITE_DETACH`` — number
* ``constants.SQLITE_DROP_TABLE`` — number
* ``constants.SQLITE_ERROR`` — number
* ``constants.SQLITE_IGNORE`` — number
* ``constants.SQLITE_INSERT`` — number
* ``constants.SQLITE_LOCKED`` — number
* ``constants.SQLITE_MISMATCH`` — number
* ``constants.SQLITE_OK`` — number
* ``constants.SQLITE_PRAGMA`` — number
* ``constants.SQLITE_READONLY`` — number
* ``constants.SQLITE_SELECT`` — number
* ``constants.SQLITE_TRANSACTION`` — number
* ``constants.SQLITE_UPDATE`` — number
* ``DatabaseSync.isTransaction`` — boolean
* ``DatabaseSync.readOnly`` — undefined
* ``StatementSync.expandedSQL`` — string
* ``StatementSync.sourceSQL`` — string

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
