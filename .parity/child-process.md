# child-process: 63 failing official tests

## test-child-process-advanced-serialization.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-child-process-bad-stdio.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
ok 1 - normal execution of a child process is handled
not ok 2 - execution with an error event is handled
  ---
  error: |-
    Error: foo
    emit@[native code]
    @/tmp/node24tests/test/parallel/test-child-process-bad-stdio.js:56:13
    runTest@[gode:test]:250:40
    @[gode:test]:165:18
    runHooks@[gode:test]:149:9
    step@[gode:test]:164:17
    @[gode:test]:166:62
    runHooks@[gode:test]:149:9
    @[gode:test]:166:21
    finishOk@[gode:test]:210:9
    settle@[gode:test]:235:15
    doneArg@[gode:test]:245:13
    @/tmp/node24tests/test/parallel/test-child-process-bad-stdio.js:43:9
    f@[native code]
    f@[native code]
```

## test-child-process-constructor.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "options.file" property must be of type string. Received undefined',
-   message: 'The "options.envPairs" property must be an instance of Array. Received null',
    name: 'TypeError'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "options.file" property must be of type string. Received undefined',
-   message: 'The "options.envPairs" property must be an instance of Array. Received null',
    name: 'TypeError'
```

## test-child-process-disconnect.js (exit 124)
```

```

## test-child-process-exec-abortcontroller-promisified.js (exit 1)
```
gode: ReferenceError: Can't find variable: DOMException
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-exec-abortcontroller-promisified.js:23:5 ]
    │
 23 │     cause: new DOMException('This operation was aborted', 'AbortError'),
    │     ┬
    │     ╰── Can't find variable: DOMException
    │
    │ Help: check for a typo, or declare the identifier before using it
────╯
```

## test-child-process-exec-encoding.js (exit 2)
```
runtime: g 66: unexpected return pc for os/exec.(*Cmd).Start.gowrap1 called from 0xd3479aed267b6400
stack: frame={sp:0x35dfe2d4efc8, fp:0x35dfe2d4efe0} stack=[0x35dfe2d4e000,0x35dfe2d4f000)
                   7 6 5 4  3 2 1 0   f e d c  b a 9 8  0123456789abcdef
000035dfe2d4eec0:                    00000000 00000000          ........
000035dfe2d4eed0: 00000000 00000000  00000000 00000000  ................
000035dfe2d4eee0: 00000000 00000000  00000000 00000000  ................
000035dfe2d4eef0: 000035df e2d4ef50  00000000 005679d4  P....5...yV.....
                                     ^ <os/exec.(*Cmd).writerDescriptor.func1+0x34>
000035dfe2d4ef00: 00000000 00a26900  000035df e2da5200  .i.......R...5..
000035dfe2d4ef10: 00000000 00a26260  000035df e320e018  `b........ ..5..
000035dfe2d4ef20: 00000000 00000000  00000000 00000000  ................
000035dfe2d4ef30: 00000000 00000000  00000000 00000000  ................
000035dfe2d4ef40: 000035df e320e018  00000000 00000000  .. ..5..........
000035dfe2d4ef50: 000035df e2d4efb8  00000000 005685ac  .....5....V.....
                                     ^ <os/exec.(*Cmd).Start.func2+0x2c>
000035dfe2d4ef60: 00000000 00000000  00000000 00000000  ................
000035dfe2d4ef70: 00000000 00000000  000035df e31520e0  ......... ...5..
000035dfe2d4ef80: 000035df e3152070  00000000 00000000  p ...5..........
000035dfe2d4ef90: 00000000 00000000  00000000 00000000  ................
000035dfe2d4efa0: 00000000 00000000  00000000 00000000  ................
000035dfe2d4efb0: 00000000 00000000  000035df e2d4efd0  .............5..
000035dfe2d4efc0: 00000000 0056855b  00000000 00000000  [.V.............
```

## test-child-process-exec-kill-throws.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: null is not an object (evaluating 'err.message')
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-exec-kill-throws.js:26:5 ]
    │
 26 │     assert.strictEqual(err.message, 'mock error', err);
    │     ┬
    │     ╰── null is not an object (evaluating 'err.message')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-child-process-exec-maxbuf.js (exit 1)
```
gode: TypeError: null is not an object (evaluating 'err.message')
   ╭─[ /tmp/node24tests/test/parallel/test-child-process-exec-maxbuf.js:7:3 ]
   │
 7 │   assert.strictEqual(err.message, `${streamName} maxBuffer length exceeded`);
   │   ┬
   │   ╰── null is not an object (evaluating 'err.message')
   │
   │ Help: the value does not have the type this operation requires
───╯
```

## test-child-process-execFile-promisified-abortController.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-child-process-execfile-maxbuf.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'name'
- 'stdout maxBuffer length exceeded'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'name'
- 'stdout maxBuffer length exceeded'

─────╯
```

## test-child-process-execfile.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-execfilesync-maxbuf.js (exit 1)
```
gode: Error: Unknown system error -105
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-execfilesync-maxbuf.js:25:5 ]
    │
 25 │     assert.strictEqual(getSystemErrorName(e.errno), 'ENOBUFS');
    │     ┬
    │     ╰── Unknown system error -105
────╯
```

## test-child-process-execsync-maxbuf.js (exit 1)
```
gode: Error: Unknown system error -105
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-execsync-maxbuf.js:22:5 ]
    │
 22 │     assert.strictEqual(getSystemErrorName(e.errno), 'ENOBUFS');
    │     ┬
    │     ╰── Unknown system error -105
────╯
```

## test-child-process-flush-stdio.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'' !== '123'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'' !== '123'

─────╯
```

## test-child-process-fork-advanced-header-serialization.js (exit 124)
```

```

## test-child-process-fork-and-spawn.js (exit 124)
```

```

## test-child-process-fork-args.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
gode: Error: Channel closed
   ╭─[ /tmp/node24tests/test/fixtures/child-process-echo-options.js:1:1 ]
   │
 1 │ process.send({ env: process.env });
   │ ┬
   │ ╰── Channel closed
───╯
gode: Error: Channel closed
   ╭─[ /tmp/node24tests/test/fixtures/child-process-echo-options.js:1:1 ]
   │
 1 │ process.send({ env: process.env });
   │ ┬
   │ ╰── Channel closed
───╯
gode: Error: Channel closed
```

## test-child-process-fork-close.js (exit 124)
```

```

## test-child-process-fork-closed-channel-segfault.js (exit 1)
```
gode: Error: GodeJS does not implement cluster.fork yet
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-fork-closed-channel-segfault.js:32:5 ]
    │
 32 │     const worker = cluster.fork();
    │     ┬
    │     ╰── GodeJS does not implement cluster.fork yet
────╯
```

## test-child-process-fork-dgram.js (exit 1)
```
gode: Error: GodeJS does not implement dgram.createSocket yet
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-fork-dgram.js:53:3 ]
    │
 53 │   const parentServer = dgram.createSocket('udp4');
    │   ┬
    │   ╰── GodeJS does not implement dgram.createSocket yet
────╯
```

## test-child-process-fork-exec-path.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '/tmp/gode'
- '/tmp/node24tests/test/.tmp.test-child-process-fork-exec-path/node-copy.exe'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '/tmp/gode'
- '/tmp/node24tests/test/.tmp.test-child-process-fork-exec-path/node-copy.exe'

─────╯
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
```

## test-child-process-fork-getconnections.js (exit 124)
```

```

## test-child-process-fork-net-server.js (exit 124)
```

```

## test-child-process-fork-net-socket.js (exit 124)
```

```

## test-child-process-fork-net.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(!closeEmitted)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(!closeEmitted)

──────╯
```

## test-child-process-fork-ref.js (exit 124)
```

```

## test-child-process-fork-stdio-string-variant.js (exit 1)
```
gode: Cannot find module '/tmp/node24tests/test/fixtures/child-process-spawn-node'
   ╭─[ /tmp/node24tests/test/fixtures/child-process-spawn-node:1:1 ]
   │
 1 │ 
   │ ┬
   │ ╰── Cannot find module '/tmp/node24tests/test/fixtures/child-process-spawn-node'
───╯
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-fork-stdio.js (exit 1)
```
gode: TypeError: undefined is not an object (evaluating 'child.stdio[4].write')
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-fork-stdio.js:53:3 ]
    │
 53 │   child.stdio[4].write(buf);
    │   ┬
    │   ╰── undefined is not an object (evaluating 'child.stdio[4].write')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-child-process-fork-timeout-kill-signal.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-fork.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ {
+   _handle: null,
+   close: [Function: close],
+   ref: [Function: ref],
+   unref: [Function: unref]
+ }
- undefined

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ {
+   _handle: null,
+   close: [Function: close],
```

## test-child-process-fork3.js (exit 124)
```

```

## test-child-process-http-socket-leak.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: GodeJS does not implement http.createServer yet
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-http-socket-leak.js:26:1 ]
    │
 26 │ const server = http.createServer(common.mustCall((req, res) => {
    │ ┬
    │ ╰── GodeJS does not implement http.createServer yet
────╯
```

## test-child-process-ipc-next-tick.js (exit 124)
```

```

## test-child-process-ipc.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-child-process-ipc.js:38
called with arguments: <Buffer 67 6f 64 65 3a 20 54 79 70 65 45 72 72 6f 72 3a 20 70 72 6f 63 65 73 73 2e 6f 70 65 6e 53 74 64 69 6e 20 69 73 20 6e 6f 74 20 61 20 66 75 6e 63 74 69 ... 395 more bytes>
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-child-process-ipc.js:38
called with arguments: <Buffer 67 6f 64 65 3a 20 54 79 70 65 45 72 72 6f 72 3a 20 70 72 6f 63 65 73 73 2e 6f 70 65 6e 53 74 64 69 6e 20 69 73 20 6e 6f 74 20 61 20 66 75 6e 63 74 69 ... 395 more bytes>
──────╯
```

## test-child-process-kill-sigwinch.js (exit 1)
```
gode: TypeError: child.removeAllListeners is not a function. (In 'child.removeAllListeners('exit')', 'child.removeAllListeners' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-kill-sigwinch.js:31:5 ]
    │
 31 │     child.removeAllListeners('exit');
    │     ┬
    │     ╰── child.removeAllListeners is not a function. (In 'child.removeAllListeners('exit')', 'child.removeAllListeners' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-child-process-kill.js (exit 124)
```

```

## test-child-process-pipe-dataflow.js (exit 124)
```

```

## test-child-process-promisified.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(promise.child instanceof child_process.ChildProcess)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(promise.child instanceof child_process.ChildProcess)

──────╯
```

## test-child-process-recv-handle.js (exit 1)
```
gode: TypeError: undefined is not an object (evaluating 'process[kChannelHandle].readStop')
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-recv-handle.js:63:3 ]
    │
 63 │   process[kChannelHandle].readStop();  // Make messages batch up.
    │   ┬
    │   ╰── undefined is not an object (evaluating 'process[kChannelHandle].readStop')
    │
    │ Help: the value does not have the type this operation requires
────╯
gode: TypeError: null is not an object (evaluating 'handle.close')
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-recv-handle.js:46:5 ]
    │
 46 │     handle.close();
    │     ┬
    │     ╰── null is not an object (evaluating 'handle.close')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-child-process-reject-null-bytes.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-child-process-send-boundsocket.js (exit 1)
```
gode: TypeError: undefined is not a constructor (evaluating 'new net.BoundSocket({ path: common.PIPE })')
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-send-boundsocket.js:43:1 ]
    │
 43 │ const pipeBound = new net.BoundSocket({ path: common.PIPE });
    │ ┬
    │ ╰── undefined is not a constructor (evaluating 'new net.BoundSocket({ path: common.PIPE })')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-child-process-send-keep-open.js (exit 124)
```

```

## test-child-process-send-returns-boolean.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

false !== true

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

false !== true

─────╯
```

## test-child-process-send-type-error.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-send-utf8.js (exit 1)
```
gode: ReferenceError: Can't find variable: assert
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-send-utf8.js:33:5 ]
    │
 33 │     assert.strictEqual(actual, expected);
    │     ┬
    │     ╰── Can't find variable: assert
    │
    │ Help: check for a typo, or declare the identifier before using it
────╯
```

## test-child-process-server-close.js (exit 124)
```

```

## test-child-process-spawn-argv0.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "options.argv0" property must be of type string. Received object (Array)',
-   message: 'The "options.argv0" property must be of type string. Received an instance of Array',
    name: 'TypeError'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "options.argv0" property must be of type string. Received object (Array)',
-   message: 'The "options.argv0" property must be of type string. Received an instance of Array',
    name: 'TypeError'
```

## test-child-process-spawn-shell.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'' !== 'buzz'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'' !== 'buzz'

─────╯
```

## test-child-process-spawn-timeout-kill-signal.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-spawn-typeerror.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-child-process-spawn-windows-batch-file.js (exit 2)
```
SIGSEGV: segmentation violation
PC=0x7fc9a00302c7 m=11 sigcode=1 addr=0x5
signal arrived during cgo execution

goroutine 1 gp=0x3c15e722e1e0 m=11 mp=0x3c15e7301808 [syscall, locked to thread]:
runtime.cgocall(0x897a20, 0x3c15e72673e0)
	/usr/lib/golang/src/runtime/cgocall.go:167 +0x4b fp=0x3c15e72673b8 sp=0x3c15e7267380 pc=0x481f4b
godejs/internal/js._Cfunc_JSObjectCallAsFunction(0x7fc9ef220088, 0x7fc9ef34c400, 0x7fc97a0ff140, 0x1, 0x3c15e7267460, 0x7fc96c000d40)
	_cgo_gotypes.go:139 +0x47 fp=0x3c15e72673e0 sp=0x3c15e72673b8 pc=0x4fc247
godejs/internal/js.Object.Call({{0x3c15e73325a0?, 0x3c15e73325a0?}}, {{0x3c15e73325a0?, 0x2?}}, {0x3c15e7267668, 0x1, 0x0?})
	/home/krishnaa/godejs/internal/js/value.go:603 +0x20e fp=0x3c15e72674c8 sp=0x3c15e72673e0 pc=0x50610e
godejs/modules/child_process.(*module).emitValue(0x3c15e7588050, {{0x3c15e73325a0?, 0x458e71?}}, {0x950662, 0x5}, {0x3c15e7267668, 0x1, 0x1})
	/home/krishnaa/godejs/modules/child_process/child_process.go:2455 +0x26e fp=0x3c15e72675b0 sp=0x3c15e72674c8 pc=0x57fa0e
godejs/modules/child_process.(*module).queueFail(0x3c15e7588050, 0x3c15e75a2580, {{0x3c15e73325a0?, 0x4ffadb?}})
	/home/krishnaa/godejs/modules/child_process/child_process.go:1916 +0x11f fp=0x3c15e7267688 sp=0x3c15e72675b0 pc=0x57b1ff
godejs/modules/child_process.(*module).spawnChildWith.func1({{0xfffe000000000465?, 0x0?}}, {0x4091940000000000?, 0xfffe000000000465?, 0x3c15e7267710?})
	/home/krishnaa/godejs/modules/child_process/child_process.go:1736 +0x28 fp=0x3c15e72676c8 sp=0x3c15e7267688 pc=0x57aa88
godejs/internal/js.godeCallback.func1(0x3c15e72677d8, 0x3c15e72677b0, 0x465?, 0x3c15e72677a0?, 0x3c15e73325a0?, {0xe909a0?, 0x3c15e7240cc0?, 0xfffe0000000000c4?})
	/home/krishnaa/godejs/internal/js/callback.go:52 +0x69 fp=0x3c15e7267720 sp=0x3c15e72676c8 pc=0x4ff6a9
godejs/internal/js.godeCallback(0x7fc9ef220088, 0x7fc9efe7a280, 0x7fc9ee05dec8, 0x0, 0x7fc9817f9910, 0x7fc9817f9880)
	/home/krishnaa/godejs/internal/js/callback.go:53 +0x255 fp=0x3c15e7267800 sp=0x3c15e7267720 pc=0x4ff315
_cgoexp_ac897f297743_godeCallback(0x7fc9817f97f0)
```

## test-child-process-spawnsync-input.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-child-process-spawnsync-maxbuf.js (exit 1)
```
gode: Error: Unknown system error -105
    ╭─[ /tmp/node24tests/test/parallel/test-child-process-spawnsync-maxbuf.js:24:3 ]
    │
 24 │   assert.strictEqual(getSystemErrorName(ret.error.errno), 'ENOBUFS');
    │   ┬
    │   ╰── Unknown system error -105
────╯
```

## test-child-process-spawnsync-shell.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:

'' !== 'buzz'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'' !== 'buzz'

─────╯
```

## test-child-process-spawnsync-validation-errors.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-child-process-stdio-big-write-end.js (exit 124)
```

```

## test-child-process-stdio-merge-stdouts-into-cat.js (exit 124)
```

```

## test-child-process-stdio-reuse-readable-stdio.js (exit 124)
```

```

## test-child-process-stdout-flush.js (exit 2)
```
SIGSEGV: segmentation violation
PC=0x7f3f922e758a m=5 sigcode=128 addr=0x0
signal arrived during cgo execution

goroutine 1 gp=0x3d0c3df041e0 m=5 mp=0x3d0c3e000008 [syscall, locked to thread]:
runtime.cgocall(0x897a20, 0x3d0c3df3d1a0)
	/usr/lib/golang/src/runtime/cgocall.go:167 +0x4b fp=0x3d0c3df3d178 sp=0x3d0c3df3d140 pc=0x481f4b
godejs/internal/js._Cfunc_JSObjectCallAsFunction(0x7f3f27220088, 0x7f3eba0a6820, 0x7f3eba0bd3c0, 0x1, 0x3d0c3df3d220, 0x7f3f30000df0)
	_cgo_gotypes.go:139 +0x47 fp=0x3d0c3df3d1a0 sp=0x3d0c3df3d178 pc=0x4fc247
godejs/internal/js.Object.Call({{0x3d0c3df165e0?, 0x3d0c3df165e0?}}, {{0x3d0c3df165e0?, 0x2?}}, {0x3d0c3df3d438, 0x1, 0x504bfc?})
	/home/krishnaa/godejs/internal/js/value.go:603 +0x20e fp=0x3d0c3df3d288 sp=0x3d0c3df3d1a0 pc=0x50610e
godejs/modules/child_process.(*module).emitValue(0x3d0c3e250050, {{0x3d0c3df165e0?, 0x95746a?}}, {0x95015c, 0x4}, {0x3d0c3df3d438, 0x1, 0x1})
	/home/krishnaa/godejs/modules/child_process/child_process.go:2455 +0x26e fp=0x3d0c3df3d370 sp=0x3d0c3df3d288 pc=0x57fa0e
godejs/modules/child_process.(*module).emitStream(0x3d0c3e250050, {{0x3d0c3df165e0?, 0x3d0c3df3d400?}}, {0x3d0c3e508000, 0x7a120, 0x7c000})
	/home/krishnaa/godejs/modules/child_process/child_process.go:2305 +0x1e7 fp=0x3d0c3df3d488 sp=0x3d0c3df3d370 pc=0x57e287
godejs/modules/child_process.(*module).poll(0x3d0c3e250050, 0x3d0c3e058160)
	/home/krishnaa/godejs/modules/child_process/child_process.go:2020 +0x6ca fp=0x3d0c3df3d710 sp=0x3d0c3df3d488 pc=0x57bdea
godejs/modules/child_process.(*module).spawnChildWith.func5({{0xfffe000000000385?, 0x0?}}, {0x408c280000000000?, 0xfffe000000000385?, 0x3d0c3df3d788?})
	/home/krishnaa/godejs/modules/child_process/child_process.go:1817 +0x25 fp=0x3d0c3df3d740 sp=0x3d0c3df3d710 pc=0x57a5e5
godejs/internal/js.godeCallback.func1(0x3d0c3df3d850, 0x3d0c3df3d828, 0x385?, 0x3d0c3df3d818?, 0xa?, {0xe909a0?, 0x3d0c3e2aed30?, 0x7f3f8e251aa8?})
	/home/krishnaa/godejs/internal/js/callback.go:52 +0x69 fp=0x3d0c3df3d798 sp=0x3d0c3df3d740 pc=0x4ff6a9
godejs/internal/js.godeCallback(0x7f3f27220088, 0x7f3f27e7a380, 0x7f3f2605dec8, 0x0, 0x7f3f45e5b910, 0x7f3f45e5b880)
```

## test-child-process-stdout-ipc.js (exit 1)
```
gode: Error: Channel closed
   ╭─[ /tmp/node24tests/test/parallel/test-child-process-stdout-ipc.js:8:3 ]
   │
 8 │   process.send('hahah');
   │   ┬
   │   ╰── Channel closed
───╯
gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0

─────╯
```

## test-child-process-uid-gid.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-child-process-validate-stdio.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { name: 'TypeError' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE', name: 'TypeError' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { name: 'TypeError' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE', name: 'TypeError' }

──────╯
```

## test-child-process-windows-hide.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - spawnSync() passes windowsHide correctly
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    0 !== 1
    
    @/tmp/node24tests/test/parallel/test-child-process-windows-hide.js:25:21
  code: ERR_ASSERTION
  ...
gode: AssertionError: Expected values to be strictly equal:

0 !== 1

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

```
