# fs: 75 failing official tests

## test-fs-append-file-flush.js (exit 1)
```
TAP version 13
not ok 1 - synchronous version > validation
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception.
    @/tmp/node24tests/test/parallel/test-fs-append-file-flush.js:20:20
    Promise@[native code]
    @/tmp/node24tests/test/parallel/test-fs-append-file-flush.js:18:15
  code: ERR_ASSERTION
  ...
not ok 2 - synchronous version > performs flush
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    0 !== 1
    
    @/tmp/node24tests/test/parallel/test-fs-append-file-flush.js:31:23
    Promise@[native code]
    @/tmp/node24tests/test/parallel/test-fs-append-file-flush.js:26:15
  code: ERR_ASSERTION
  ...
```

## test-fs-append-file.js (exit 1)
```
gode: Error: Cannot find module 'function() { // eslint-disable-line func-style
    context.actual++;
    return fn.apply(this, arguments);
  }'
     ╭─[ /tmp/node24tests/test/parallel/test-fs-append-file.js:155:7 ]
     │
 155 │       fs.close(fd, common.mustSucceed(() => {
     │       ┬
     │       ╰── Cannot find module 'function() { // eslint-disable-line func-style
    context.actual++;
    return fn.apply(this, arguments);
  }'
─────╯
```

## test-fs-chmod-mask.js (exit 1)
```
gode: Error: Cannot find module 'function ifError(err) {
    if (err !== null && err !== undefined) {
      var message = 'ifError got unwanted exception: ';
      if (typeof err === 'object' && typeof err.message === 'string') {
        if (err.message.length === 0 && err.constructor) {
          message += err.constructor.name;
        } else {
          message += err.message;
        }
      } else {
        message += inspectValue(err);
      }

      var newErr = new AssertionError({
        actual: err,
        expected: null,
        operator: 'ifError',
        message: message,
        diff: diffOf(this)
      });

      var origStack = err.stack;
```

## test-fs-chmod.js (exit 1)
```
gode: TypeError: not a net.Server
     ╭─[ /tmp/node24tests/test/parallel/test-fs-chmod.js:119:5 ]
     │
 119 │     fs.close(fd, assert.ifError);
     │     ┬
     │     ╰── not a net.Server
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-fs-close-errors.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-fs-close.js (exit 1)
```
gode: TypeError: not a net.Server
    ╭─[ /tmp/node24tests/test/parallel/test-fs-close.js:10:1 ]
    │
 10 │ fs.close(fd, common.mustCall(function(...args) {
    │ ┬
    │ ╰── not a net.Server
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-cp-sync-dereference.js (exit 1)
```
gode: Error: EISDIR: illegal operation on a directory, cp '/tmp/node24tests/test/.tmp.test-fs-cp-sync-dereference/a'
   ╭─[ [fs:throw]:3:18 ]
   │
 3 │ // Refs: https://github.com/nodejs/node/issues/58939
   │                  ┬
   │                  ╰── EISDIR: illegal operation on a directory, cp '/tmp/node24tests/test/.tmp.test-fs-cp-sync-dereference/a'
───╯
```

## test-fs-empty-readStream.js (exit 1)
```
gode: AssertionError: end event should not emit at /tmp/node24tests/test/parallel/test-fs-empty-readStream.js:45
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── end event should not emit at /tmp/node24tests/test/parallel/test-fs-empty-readStream.js:45
──────╯
```

## test-fs-error-messages.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'ENOENT: no such file or directory, lstat \'/tmp/node24tests/test/.tmp.test-fs-error-messages/non-existent\''
- 'ENOENT: no such file or directory, realpath \'/tmp/node24tests/test/.tmp.test-fs-error-messages/non-existent\''

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'ENOENT: no such file or directory, lstat \'/tmp/node24tests/test/.tmp.test-fs-error-messages/non-existent\''
- 'ENOENT: no such file or directory, realpath \'/tmp/node24tests/test/.tmp.test-fs-error-messages/non-existent\''

─────╯
```

## test-fs-fchmod.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE' }

──────╯
```

## test-fs-fchown.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "cb" argument must be of type function. Received type string (\'\')',
-   message: /fd|uid|gid/,
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
+   message: 'The "cb" argument must be of type function. Received type string (\'\')',
-   message: /fd|uid|gid/,
    name: 'TypeError'
```

## test-fs-filehandle.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-gc', '--no-warnings', '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: host.binding is not a function. (In 'host.binding(id)', 'host.binding' is undefined)
    ╭─[ [gode:internal/test/binding]:18:56 ]
    │
 18 │                             stringToFlags('r'), 0o666, undefined, ctx).fd;
    │                                                        ┬
    │                                                        ╰── host.binding is not a function. (In 'host.binding(id)', 'host.binding' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-mkdir.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "options.recursive" property must be of type boolean. Received Array(0) []',
-   message: 'The "options.recursive" property must be of type boolean. Received an instance of Array',
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
+   message: 'The "options.recursive" property must be of type boolean. Received Array(0) []',
-   message: 'The "options.recursive" property must be of type boolean. Received an instance of Array',
    name: 'TypeError'
```

## test-fs-mkdtempDisposableSync.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(!fs.existsSync(fullSecondPath))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(!fs.existsSync(fullSecondPath))

──────╯
```

## test-fs-open-flags.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE', name: 'TypeError' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE', name: 'TypeError' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE', name: 'TypeError' }
- Comparison { code: 'ERR_INVALID_ARG_VALUE', name: 'TypeError' }

──────╯
```

## test-fs-promises-exists.js (exit 1)
```
gode: AssertionError: Values have same structure but are not reference-equal:

[Object: null prototype] {
  COPYFILE_EXCL: 1,
  COPYFILE_FICLONE: 2,
  COPYFILE_FICLONE_FORCE: 4,
  F_OK: 0,
  O_APPEND: 1024,
  O_CREAT: 64,
  O_DIRECTORY: 65536,
  O_DSYNC: 4096,
  O_EXCL: 128,
  O_NOCTTY: 256,
  O_NOFOLLOW: 131072,
  O_NONBLOCK: 2048,
  O_RDONLY: 0,
  O_RDWR: 2,
  O_SYNC: 1052672,
  O_TRUNC: 512,
  O_WRONLY: 1,
  R_OK: 4,
  S_IFBLK: 24576,
```

## test-fs-promises-file-handle-dispose.js (exit 1)
```
gode: SyntaxError: Unexpected identifier 'openHandle'
    ╭─[ /tmp/node24tests/test/parallel/test-fs-promises-file-handle-dispose.js:28:5 ]
    │
 28 │     await using openHandle = await fs.open(__filename);
    │     ┬
    │     ╰── Unexpected identifier 'openHandle'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-fs-promises-file-handle-writer.js (exit 1)
```
gode: SyntaxError: Unexpected identifier 'fh'
     ╭─[ /tmp/node24tests/test/parallel/test-fs-promises-file-handle-writer.js:411:5 ]
     │
 411 │     await using fh = await open(filePath, 'w');
     │     ┬
     │     ╰── Unexpected identifier 'fh'
     │
     │ Help: the file could not be parsed; check the highlighted line
─────╯
```

## test-fs-promises-readfile-buffer-option.js (exit 1)
```
gode: SyntaxError: Unexpected identifier 'handle'
    ╭─[ /tmp/node24tests/test/parallel/test-fs-promises-readfile-buffer-option.js:72:5 ]
    │
 72 │     await using handle = await open(file, 'r');
    │     ┬
    │     ╰── Unexpected identifier 'handle'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-fs-read-empty-buffer.js (exit 1)
```
gode: SyntaxError: Unexpected identifier 'filehandle'
    ╭─[ /tmp/node24tests/test/parallel/test-fs-read-empty-buffer.js:32:3 ]
    │
 32 │   await using filehandle = await fsPromises.open(filepath, 'r');
    │   ┬
    │   ╰── Unexpected identifier 'filehandle'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-fs-read-offset-null.js (exit 1)
```
gode: SyntaxError: Unexpected identifier 'filehandle'
    ╭─[ /tmp/node24tests/test/parallel/test-fs-read-offset-null.js:40:3 ]
    │
 40 │   await using filehandle = await fsPromises.open(filepath, 'r');
    │   ┬
    │   ╰── Unexpected identifier 'filehandle'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-fs-read-stream-fd-leak.js (exit 124)
```
testing for leaks from fs.createReadStream().close()...
testing for leaks from fs.createReadStream().destroy()...
```

## test-fs-read-stream-patch-open.js (exit 1)
```
gode: Error: ENOENT: no such file or directory, open 'asd'
   ╭─[ [fs:throw]:3:21 ]
   │
 3 │ const fs = require('fs');
   │                     ┬
   │                     ╰── ENOENT: no such file or directory, open 'asd'
───╯
```

## test-fs-read-stream-pos.js (exit 124)
```

```

## test-fs-read-type.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-fs-read.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE' }
- Comparison { code: 'ERR_MISSING_ARGS' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_TYPE' }
- Comparison { code: 'ERR_MISSING_ARGS' }

──────╯
```

## test-fs-readSync-optional-params.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-fs-readdir-stack-overflow.js (exit 124)
```

```

## test-fs-readdir-types.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: function is not a constructor (evaluating 'new fs.Dirent('foo', types[method])')
     ╭─[ /tmp/node24tests/test/parallel/test-fs-readdir-types.js:128:3 ]
     │
 128 │   const dirent = new fs.Dirent('foo', types[method]);
     │   ┬
     │   ╰── function is not a constructor (evaluating 'new fs.Dirent('foo', types[method])')
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-fs-readfile-error.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "path" argument must be of type string or an instance of Buffer or URL. Received function <anonymous>',
-   message: 'The "path" argument must be of type string or an instance of Buffer or URL. Received function ',
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
+   message: 'The "path" argument must be of type string or an instance of Buffer or URL. Received function <anonymous>',
-   message: 'The "path" argument must be of type string or an instance of Buffer or URL. Received function ',
    name: 'TypeError'
```

## test-fs-readfile.js (exit 124)
```

```

## test-fs-realpath-native.js (exit 1)
```
gode: Error: ENOENT: no such file or directory, lstat './test/parallel/test-fs-realpath-native.js'
   ╭─[ [fs:throw]:3:16 ]
   │
 3 │ const assert = require('assert');
   │                ┬
   │                ╰── ENOENT: no such file or directory, lstat './test/parallel/test-fs-realpath-native.js'
───╯
```

## test-fs-realpath-pipe.js (exit 1)
```
try {
    if (require('fs').realpathSync('/dev/stdin')) {
      process.exit(2);
    }
  } catch (e) {
    console.error(e);
    process.exit(1);
  }
Error: process.exit

gode: AssertionError: Expected values to be strictly equal:

1 !== 2

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 2

```

## test-fs-realpath-stale-stat-values.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '/tmp/node24tests/test/.tmp.test-fs-realpath-stale-stat-values/pkg-link/index.js'
- '/tmp/node24tests/test/.tmp.test-fs-realpath-stale-stat-values/pkg/index.js'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '/tmp/node24tests/test/.tmp.test-fs-realpath-stale-stat-values/pkg-link/index.js'
- '/tmp/node24tests/test/.tmp.test-fs-realpath-stale-stat-values/pkg/index.js'

─────╯
```

## test-fs-rm.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: validateRmOptionsSync is not a function. (In 'validateRmOptionsSync(filePath)', 'validateRmOptionsSync' is undefined)
     ╭─[ /tmp/node24tests/test/parallel/test-fs-rm.js:429:3 ]
     │
 429 │   assert.deepStrictEqual(validateRmOptionsSync(filePath), defaults);
     │   ┬
     │   ╰── validateRmOptionsSync is not a function. (In 'validateRmOptionsSync(filePath)', 'validateRmOptionsSync' is undefined)
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-fs-rmSync-special-char-additional-error.js (exit 1)
```
gode: Error: ENOENT: no such file or directory, chmod '/tmp/node24tests/test/.tmp.test-fs-rmSync-special-char-additional-error/速_dir'
   ╭─[ [fs:throw]:3:16 ]
   │
 3 │ const tmpdir = require('../common/tmpdir');
   │                ┬
   │                ╰── ENOENT: no such file or directory, chmod '/tmp/node24tests/test/.tmp.test-fs-rmSync-special-char-additional-error/速_dir'
───╯
```

## test-fs-rmdir-recursive.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: validateRmdirOptions is not a function. (In 'validateRmdirOptions()', 'validateRmdirOptions' is undefined)
     ╭─[ /tmp/node24tests/test/parallel/test-fs-rmdir-recursive.js:173:3 ]
     │
 173 │   assert.deepStrictEqual(validateRmdirOptions(), defaults);
     │   ┬
     │   ╰── validateRmdirOptions is not a function. (In 'validateRmdirOptions()', 'validateRmdirOptions' is undefined)
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-fs-stat.js (exit 1)
```
gode: TypeError: not a net.Server
    ╭─[ /tmp/node24tests/test/parallel/test-fs-stat.js:71:3 ]
    │
 71 │   fs.close(fd, common.mustSucceed());
    │   ┬
    │   ╰── not a net.Server
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-stream-construct-compat-error-write.js (exit 1)
```
11/3/2026, 2:15:38 PM start test
11/3/2026, 2:15:38 PM waiting for callbacks
11/3/2026, 2:15:38 PM fs open() callback
gode: Error: Cannot find module '() => { debuglog(`closed ${fd}`); }'
    ╭─[ /tmp/node24tests/test/parallel/test-fs-stream-construct-compat-error-write.js:37:5 ]
    │
 37 │     fs.close(fd, () => { debuglog(`closed ${fd}`); });
    │     ┬
    │     ╰── Cannot find module '() => { debuglog(`closed ${fd}`); }'
────╯
```

## test-fs-stream-windows-handle.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-fs-symlink.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-fs-sync-fd-leak.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-fs-syncwritestream.js (exit 1)
```
gode: SyntaxError: JSON Parse error: Unexpected EOF
    ╭─[ /tmp/node24tests/test/parallel/test-fs-syncwritestream.js:36:3 ]
    │
 36 │   assert.deepStrictEqual(JSON.parse(fs.readFileSync(filename, 'utf8')), [
    │   ┬
    │   ╰── JSON Parse error: Unexpected EOF
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-fs-truncate.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-fs-utils-get-dirents.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: getDirents is not a function. (In 'getDirents(
    tmpdir.path,
    [[filename], [UV_DIRENT_UNKNOWN]],
    common.mustCall((err, names) => {
      assert.strictEqual(err, null);
      assert.strictEqual(names.length, 1);
    },
    ))', 'getDirents' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-utils-get-dirents.js:22:3 ]
    │
 22 │   getDirents(
    │   ┬
    │   ╰── getDirents is not a function. (In 'getDirents(
    tmpdir.path,
    [[filename], [UV_DIRENT_UNKNOWN]],
    common.mustCall((err, names) => {
      assert.strictEqual(err, null);
      assert.strictEqual(names.length, 1);
    },
    ))', 'getDirents' is undefined)
    │
```

## test-fs-utimes.js (exit 1)
```
gode: AssertionError: FAILED: expect_ok [Arguments] {
  '0': 'lutimes',
  '1': '/tmp/node24tests/test/.tmp.test-fs-utimes/symlink',
  '2': Error: ENOENT: no such file or directory, utime '/tmp/node24tests/test/.tmp.test-fs-utimes/symlink',
  '3': 1982-09-10T20:37:00.000Z,
  '4': 400538220,
  '5': [Function: lstatSync]
}
     check_mtime: -1393205918.838
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── FAILED: expect_ok [Arguments] {
  '0': 'lutimes',
  '1': '/tmp/node24tests/test/.tmp.test-fs-utimes/symlink',
  '2': Error: ENOENT: no such file or directory, utime '/tmp/node24tests/test/.tmp.test-fs-utimes/symlink',
  '3': 1982-09-10T20:37:00.000Z,
  '4': 400538220,
  '5': [Function: lstatSync]
}
     check_mtime: -1393205918.838
```

## test-fs-watch-abort-signal.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AbortError: The operation was aborted
     ╭─[ [fs:streams]:757:1 ]
     │
 757 │ 
     │ ┬
     │ ╰── The operation was aborted
─────╯
```

## test-fs-watch-encoding.js (exit 124)
```

```

## test-fs-watch-enoent.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected "actual" to be reference-equal to "expected":
+ actual - expected

+ TypeError {}
- Error { code: 'ENOSPC' }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" to be reference-equal to "expected":
+ actual - expected

+ TypeError {}
- Error { code: 'ENOSPC' }

─────╯
```

## test-fs-watch-file-enoent-after-deletion.js (exit 124)
```

```

## test-fs-watch-recursive-delete-race.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](parent)', 'watcher[kFSWatchStart]' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watch-recursive-delete-race.js:39:3 ]
    │
 39 │   watcher[kFSWatchStart](parent);
    │   ┬
    │   ╰── watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](parent)', 'watcher[kFSWatchStart]' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-watch-recursive-linux-directory-watchers.js (exit 124)
```

```

## test-fs-watch-recursive-promise.js (exit 124)
```

```

## test-fs-watch-recursive-retry-directory-disappears-before-read.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watch-recursive-retry-directory-disappears-before-read.js:43:3 ]
    │
 43 │   watcher[kFSWatchStart](directory);
    │   ┬
    │   ╰── watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-watch-recursive-retry-directory-disappears-before-watch.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watch-recursive-retry-directory-disappears-before-watch.js:54:3 ]
    │
 54 │   watcher[kFSWatchStart](directory);
    │   ┬
    │   ╰── watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-watch-recursive-root-disappears-before-read.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison {
-   code: 'ENOENT',
-   filename: '/tmp/node24tests/test/.tmp.test-fs-watch-recursive-root-disappears-before-read/root-disappears-before-read'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison {
-   code: 'ENOENT',
-   filename: '/tmp/node24tests/test/.tmp.test-fs-watch-recursive-root-disappears-before-read/root-disappears-before-read'
- }
```

## test-fs-watch-recursive-root-disappears-before-watch-no-throw.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watch-recursive-root-disappears-before-watch-no-throw.js:35:3 ]
    │
 35 │   watcher[kFSWatchStart](directory);
    │   ┬
    │   ╰── watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-watch-recursive-root-disappears-before-watch.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
- undefined

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
- undefined

─────╯
```

## test-fs-watch-recursive-watch-before-read.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watch-recursive-watch-before-read.js:36:3 ]
    │
 36 │   watcher[kFSWatchStart](directory);
    │   ┬
    │   ╰── watcher[kFSWatchStart] is not a function. (In 'watcher[kFSWatchStart](directory)', 'watcher[kFSWatchStart]' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-watch-ref-unref.js (exit 124)
```

```

## test-fs-watchfile-bigint.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: undefined is not a constructor (evaluating 'new BigIntStats(
  0n,                                        // dev
  0n,                                        // mode
  0n,                                        // nlink
  0n,                                        // uid
  0n,                                        // gid
  0n,                                        // rdev
  0n,                                        // blksize
  0n,                                        // ino
  0n,                                        // size
  0n,                                        // blocks
  0n,                                        // atimeMs
  0n,                                        // mtimeMs
  0n,                                        // ctimeMs
  0n,                                        // birthtimeMs
  0n,                                        // atimeNs
  0n,                                        // mtimeNs
  0n,                                        // ctimeNs
  0n                                         // birthtimeNs
)')
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watchfile-bigint.js:14:1 ]
```

## test-fs-watchfile-ref-unref.js (exit 124)
```

```

## test-fs-watchfile.js (exit 1)
```
gode: TypeError: function is not a constructor (evaluating 'new fs.Stats(
  0,                                        // dev
  0,                                        // mode
  0,                                        // nlink
  0,                                        // uid
  0,                                        // gid
  0,                                        // rdev
  0,                                        // blksize
  0,                                        // ino
  0,                                        // size
  0,                                        // blocks
  Date.UTC(1970, 0, 1, 0, 0, 0),            // atime
  Date.UTC(1970, 0, 1, 0, 0, 0),            // mtime
  Date.UTC(1970, 0, 1, 0, 0, 0),            // ctime
  Date.UTC(1970, 0, 1, 0, 0, 0)             // birthtime
)')
    ╭─[ /tmp/node24tests/test/parallel/test-fs-watchfile.js:34:1 ]
    │
 34 │ const expectedStatObject = new fs.Stats(
    │ ┬
    │ ╰── function is not a constructor (evaluating 'new fs.Stats(
  0,                                        // dev
```

## test-fs-whatwg-url.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-fs-write-buffer-large.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-fs-write-file-flush.js (exit 1)
```
TAP version 13
not ok 1 - synchronous version > validation
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception.
    @/tmp/node24tests/test/parallel/test-fs-write-file-flush.js:20:20
    Promise@[native code]
    @/tmp/node24tests/test/parallel/test-fs-write-file-flush.js:18:15
  code: ERR_ASSERTION
  ...
not ok 2 - synchronous version > performs flush
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    0 !== 1
    
    @/tmp/node24tests/test/parallel/test-fs-write-file-flush.js:31:23
    Promise@[native code]
    @/tmp/node24tests/test/parallel/test-fs-write-file-flush.js:26:15
  code: ERR_ASSERTION
  ...
```

## test-fs-write-file.js (exit 1)
```
gode: TypeError: null is not an object (evaluating 'err.name')
    ╭─[ /tmp/node24tests/test/parallel/test-fs-write-file.js:72:5 ]
    │
 72 │     assert.strictEqual(err.name, 'AbortError');
    │     ┬
    │     ╰── null is not an object (evaluating 'err.name')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-fs-write-negativeoffset.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-fs-write-sigxfsz.js (exit 1)
```
gode: Error: ERRNO27: file too large, open '/tmp/node24tests/test/.tmp.test-fs-write-sigxfsz/efbig.txt'
   ╭─[ [fs:throw]:3:1 ]
   │
 3 │ 'use strict';
   │ ┬
   │ ╰── ERRNO27: file too large, open '/tmp/node24tests/test/.tmp.test-fs-write-sigxfsz/efbig.txt'
───╯

gode: AssertionError: The expression evaluated to a falsy value:

  assert(ok)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(ok)

──────╯
```

## test-fs-write-stream-autoclose-option.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-fs-write-stream-err.js (exit 1)
```
first cb
second cb
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ null
- Error: BAM

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ null
- Error: BAM

─────╯
```

## test-fs-write-stream-flush.js (exit 1)
```
TAP version 13
not ok 1 - validation
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception.
    @/tmp/node24tests/test/parallel/test-fs-write-stream-flush.js:19:18
  code: ERR_ASSERTION
  ...
```

## test-fs-write-stream-patch-open.js (exit 1)
```
gode: TypeError: s.open is not a function. (In 's.open()', 's.open' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-fs-write-stream-patch-open.js:30:1 ]
    │
 30 │ s.open();
    │ ┬
    │ ╰── s.open is not a function. (In 's.open()', 's.open' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
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

## test-fs-write-stream.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-fs-write.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose_externalize_string' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected "actual" to be strictly unequal to:

undefined
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" to be strictly unequal to:

undefined
─────╯
```
