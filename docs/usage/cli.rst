Command-line interface
======================

``gode`` accepts the Node.js 24 command line. Every flag in ``node --help`` is
recognized with Node's arity, aliases, ``--no-`` negations and ``--`` end of
options. Use ``gode --help`` for the complete list.

.. code-block:: sh

   gode app.js                     # run a file
   gode -e "console.log(1 + 1)"    # evaluate
   gode -p "process.version"       # evaluate and print
   echo "console.log(1)" | gode    # read a script from stdin
   gode                            # start the REPL

The REPL understands ``.break``, ``.clear``, ``.editor``, ``.exit``, ``.help``,
``.load`` and ``.save``. ``Ctrl-D`` also exits. Expressions are inspected with
``util.inspect``.

Input and execution
-------------------

.. list-table::
   :header-rows: 1
   :widths: 30 70

   * - Flag
     - Behavior
   * - ``-e``, ``--eval``
     - Evaluate the argument as JavaScript.
   * - ``-p``, ``--print``
     - Evaluate and print the completion value.
   * - ``-c``, ``--check``
     - Compile without running. Uses JavaScriptCore's syntax checker.
   * - ``-i``, ``--interactive``
     - Enter the REPL after the script.
   * - ``--input-type``
     - ``module`` or ``commonjs`` for ``--eval`` and stdin.
   * - ``--entry-url``
     - Accept a ``file:`` URL as the entry script.
   * - ``-r``, ``--require``
     - Preload a CommonJS module. May be repeated.
   * - ``--import``
     - Preload a module. May be repeated.
   * - ``--preserve-symlinks``
     - Do not resolve symlinks while loading modules.
   * - ``--preserve-symlinks-main``
     - Preserve the entry script's symlink (maps to the same resolver mode).
   * - ``--``
     - End option parsing.

Environment and package scripts
-------------------------------

``--env-file`` loads ``KEY=VALUE`` lines, ``export`` prefixes, comments,
blank lines and single- or double-quoted values. It may be repeated;
``--env-file-if-exists`` ignores a missing file.

``--run=<name>`` reads ``scripts[name]`` from ``package.json`` and runs it with
``sh -c``, setting ``npm_lifecycle_event`` and ``npm_package_json``.

Process and runtime
-------------------

.. list-table::
   :header-rows: 1
   :widths: 34 66

   * - Flag
     - Behavior
   * - ``--title``
     - Set ``process.title`` before user code runs.
   * - ``--stack-trace-limit``
     - Set ``Error.stackTraceLimit``.
   * - ``--expose-gc``
     - Install the ``gc()`` global.
   * - ``--frozen-intrinsics``
     - Freeze built-in constructors and prototypes.
   * - ``--disable-proto=delete|throw``
     - Remove ``Object.prototype.__proto__`` or make access throw
       ``ERR_PROTO_ACCESS``.
   * - ``--trace-exit``
     - Print a warning and stack when the environment exits.
   * - ``--localstorage-file``
     - Persist ``localStorage`` in the named file.
   * - ``--zero-fill-buffers``
     - Accepted; JavaScriptCore buffers are always zero-filled.

Warnings and reports
--------------------

``--no-warnings``, ``--no-deprecation``, ``--trace-warnings``,
``--trace-deprecation``, ``--throw-deprecation``, ``--pending-deprecation``,
``--disable-warning=CODE`` and ``--redirect-warnings=FILE`` are honored by
``process.emitWarning``.

``--report-compact``, ``--report-dir``, ``--report-filename``,
``--report-on-fatalerror``, ``--report-on-signal``, ``--report-signal`` and
``--report-uncaught-exception`` configure ``process.report``.
``--diagnostic-dir`` sets the report directory when ``--report-dir`` is absent.

Watch mode
----------

``--watch`` runs the entry script in a child process. It polls the script and
every ``--watch-path`` every 200 ms, restarts on a change, and stays alive
after the child exits so the next change starts a fresh run.
``--watch-kill-signal`` selects the stop signal (default ``SIGTERM``).
``--watch-preserve-output`` is accepted; output is never cleared.

Recognized fixed no-ops
-----------------------

These describe behavior that is already fixed by the runtime or that Node 24
itself treats as a no-op. They are accepted without error:

* V8 heap tuning: ``--max-old-space-size``, ``--max-semi-space-size``,
  ``--max-old-space-size-percentage``, ``--stack-size``.
* ``--use-largepages`` (Node documents this as a no-op).
* ``--no-experimental-detect-module``, ``--no-experimental-require-module``,
  ``--no-experimental-repl-await``, ``--no-strip-types``,
  ``--no-experimental-ffi``, ``--no-worker-snapshot``.
* Positive feature requests that are already available:
  ``--experimental-websocket``, ``--experimental-webstorage``
  (``--webstorage``), ``--experimental-sqlite``,
  ``--experimental-stream-iter``, ``--experimental-import-meta-resolve``,
  ``--experimental-vm-modules``.
* ``--use-system-ca`` (Go's default trust store) and
  ``--no-global-search-paths`` (there are no global search paths).

Explicitly refused
------------------

Flags whose behavior would otherwise be silently wrong exit with status 9 and
a one-line explanation. They are never ignored. The categories are:

* V8 engine control and profiling: ``--jitless``, ``--v8-options``,
  ``--v8-pool-size``, ``--cpu-prof*``, ``--heap-prof*``, ``--prof``,
  ``--prof-process``, ``--track-heap-objects``, ``--heapsnapshot-*``,
  ``--interpreted-frames-native-stack``.
* Inspector protocol: ``--inspect*``, ``--debug-port``, ``--disable-sigusr1``,
  ``--experimental-*-inspection``.
* Snapshots and single executable applications: ``--build-snapshot*``,
  ``--snapshot-blob``, ``--build-sea``, ``--experimental-sea-config``.
* Permissions: ``--permission*`` and every ``--allow-*`` capability.
* VFS, FFI, package maps and importable addon modules.
* The benchmark runner (``--bench*``) and the Node test runner
  (``--test*``, ``--experimental-test-*``).
* OpenSSL/FIPS and TLS tuning: ``--enable-fips*``, ``--force-fips``,
  ``--openssl-*``, ``--secure-heap*``, ``--tls-*``, ``--use-*-ca``.
* Custom loaders and configuration files: ``--loader``,
  ``--experimental-loader``, ``--experimental-config-file``,
  ``--config-file``.
* Behavioral toggles that cannot be honored here: disabling WebSocket, Web
  Storage, ``navigator`` or ``node:sqlite``; custom export conditions; DNS
  result order; HTTP header/parser limits; source maps; proxy environment
  handling; process timeouts; sync-I/O, promise, environment, SIGINT and TLS
  tracing.

A refused flag means the runtime cannot provide Node's documented effect.
Running code that depends on it would be misleading, so ``gode`` stops
instead.

Exit status
-----------

* ``0`` — success.
* ``1`` — script exception or syntax-check failure.
* ``9`` — bad command line or a refused flag.
* ``130`` — watch mode interrupted by ``SIGINT``.
