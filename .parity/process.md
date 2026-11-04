# process: 31 failing official tests

## test-process-binding-internalbinding-allowlist.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: No such module: cares_wrap
    ╭─[ /tmp/node24tests/test/parallel/test-process-binding-internalbinding-allowlist.js:10:1 ]
    │
 10 │ assert(process.binding('cares_wrap'));
    │ ┬
    │ ╰── No such module: cares_wrap
────╯
```

## test-process-constants-noatime.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert('O_NOATIME' in constants)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert('O_NOATIME' in constants)

──────╯
```

## test-process-dlopen-error-message-crash.js (exit 1)
```
gode: Error: ENOENT: no such file or directory, access 'test/addons/not-a-binding'
   ╭─[ [fs:throw]:3:14 ]
   │
 3 │ // This is a regression test for some scenarios in which node would pass
   │              ┬
   │              ╰── ENOENT: no such file or directory, access 'test/addons/not-a-binding'
───╯
```

## test-process-env-allowed-flags-are-documented.js (exit 1)
```
gode: Error: ENOENT: no such file or directory, open '/tmp/node24tests/doc/api/cli.md'
   ╭─[ [fs:throw]:3:16 ]
   │
 3 │ const common = require('../common');
   │                ┬
   │                ╰── ENOENT: no such file or directory, open '/tmp/node24tests/doc/api/cli.md'
───╯
```

## test-process-env-tz.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /^Sat Apr 14 2018 14:34:56 GMT\+0200 \(.+\)$/. Input:

'Sat Apr 14 2018 05:34:56 GMT-0700 (Pacific Daylight Time)'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /^Sat Apr 14 2018 14:34:56 GMT\+0200 \(.+\)$/. Input:

'Sat Apr 14 2018 05:34:56 GMT-0700 (Pacific Daylight Time)'

──────╯
```

## test-process-exec-argv.js (exit 1)
```
gode: TypeError: null is not an object (evaluating 'child.stdout.setEncoding')
    ╭─[ /tmp/node24tests/test/parallel/test-process-exec-argv.js:55:7 ]
    │
 55 │       child.stdout.setEncoding('utf8');
    │       ┬
    │       ╰── null is not an object (evaluating 'child.stdout.setEncoding')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-process-execpath.js (exit 1)
```
[process 1317577]: --- stderr ---
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '/tmp/node24tests/test/.tmp.test-process-execpath/symlinked-node'
- '/tmp/gode'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '/tmp/node24tests/test/.tmp.test-process-execpath/symlinked-node'
- '/tmp/gode'

─────╯

[process 1317577]: --- stdout ---

[process 1317577]: status = 1, signal = null
```

## test-process-execve-abort.js (exit 1)
```
gode: AssertionError: gode: Error: ENOENT: no such file or directory, execve '/tmp/gode_non_existing'
    ╭─[ /tmp/node24tests/test/parallel/test-process-execve-abort.js:15:3 ]
    │
 15 │   process.execve(
    │   ┬
    │   ╰── ENOENT: no such file or directory, execve '/tmp/gode_non_existing'
────╯

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── gode: Error: ENOENT: no such file or directory, execve '/tmp/gode_non_existing'
    ╭─[ /tmp/node24tests/test/parallel/test-process-execve-abort.js:15:3 ]
    │
 15 │   process.execve(
    │   ┬
    │   ╰── ENOENT: no such file or directory, execve '/tmp/gode_non_existing'
────╯

──────╯
```

## test-process-execve-permission-fail.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--permission', '--allow-fs-read=*' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: process.execve should not have been allowed.
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── process.execve should not have been allowed.
──────╯
```

## test-process-execve-worker-threads.js (exit 1)
```
gode: Cannot find module '/tmp/node24tests/test/replaced'
   ╭─[ replaced:1:1 ]
   │
 1 │ 
   │ ┬
   │ ╰── Cannot find module '/tmp/node24tests/test/replaced'
───╯
```

## test-process-exit-after-fetch-throw.js (exit 1)
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

## test-process-exit-code-validation.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /Cannot delete property 'exitCode' of #<process>/. Input:

'TypeError: Unable to delete property.'

      ╭─[ node:assert:1263:1 ]
      │
 1263 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /Cannot delete property 'exitCode' of #<process>/. Input:

'TypeError: Unable to delete property.'

──────╯
```

## test-process-exit-code.js (exit 1)
```
gode: AssertionError: wrong exit for 3-exitWithOneOnUncaught
expected:1 but got:99

99 !== 1

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── wrong exit for 3-exitWithOneOnUncaught
expected:1 but got:99

99 !== 1

─────╯
```

## test-process-exit-recursive.js (exit 1)
```

```

## test-process-external-stdio-close-spawn.js (exit 1)
```
gode: Error: Channel closed
   ╭─[ parallel/test-process-external-stdio-close-spawn.js:1:1 ]
   │
 1 │ 'use strict';
   │ ┬
   │ ╰── Channel closed
───╯
```

## test-process-getactivehandles.js (exit 124)
```

```

## test-process-getactiverequests.js (exit 1)
```
gode: TypeError: process._getActiveRequests is not a function. (In 'process._getActiveRequests()', 'process._getActiveRequests' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-process-getactiverequests.js:10:1 ]
    │
 10 │ assert.strictEqual(process._getActiveRequests().length, 12);
    │ ┬
    │ ╰── process._getActiveRequests is not a function. (In 'process._getActiveRequests()', 'process._getActiveRequests' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-process-getactiveresources-track-active-handles.js (exit 124)
```

```

## test-process-getactiveresources-track-active-requests.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

0 !== 12

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

0 !== 12

─────╯
```

## test-process-hrtime-bigint.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--allow-natives-syntax', '--expose-internals', '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-process-hrtime-bigint.js:19:1 ]
    │
 19 │ eval('%PrepareFunctionForOptimization(process.hrtime.bigint)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-process-hrtime.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--allow-natives-syntax', '--expose-internals', '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-process-hrtime.js:79:1 ]
    │
 79 │ eval('%PrepareFunctionForOptimization(process.hrtime)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-process-kill-null.js (exit 124)
```

```

## test-process-kill-pid.js (exit 143)
```

```

## test-process-load-env-file.js (exit 1)
```
TAP version 13
ok 1 - process.loadEnvFile() > supports passing path
ok 2 - process.loadEnvFile() > supports not-passing a path
ok 3 - process.loadEnvFile() > should throw when file does not exist
not ok 4 - process.loadEnvFile() > should throw when `.env` does not exist
  ---
  error: |-
    Error [ENOENT]: ENOENT: no such file or directory, chdir /tmp/node24tests/test -> '/tmp/node24tests/test/lib'
    chdir@[native code]
    chdir@[gode:process]:150:24
    @/tmp/node24tests/test/parallel/test-process-load-env-file.js:63:22
    runTest@[gode:test]:250:74
    @[gode:test]:165:18
    runHooks@[gode:test]:149:9
    step@[gode:test]:164:17
    @[gode:test]:166:62
    runHooks@[gode:test]:149:9
    @[gode:test]:166:21
    finishOk@[gode:test]:210:9
    settle@[gode:test]:235:15
  code: ENOENT
  ...
```

## test-process-really-exit.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ ''
- 'really exited'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ ''
- 'really exited'

─────╯
```

## test-process-redirect-warnings-env.js (exit 1)
```
(node:1319507) Warning: a bad practice warning
(Use `node --trace-warnings ...` to show where the warning was created)
gode: AssertionError: ifError got unwanted exception: ENOENT: no such file or directory, open '/tmp/node24tests/test/.tmp.test-process-redirect-warnings-env/warnings.txt'
      ╭─[ node:assert:1430:1 ]
      │
 1430 │ 
      │ ┬
      │ ╰── ifError got unwanted exception: ENOENT: no such file or directory, open '/tmp/node24tests/test/.tmp.test-process-redirect-warnings-env/warnings.txt'
──────╯
```

## test-process-redirect-warnings.js (exit 1)
```
(node:1319681) Warning: a bad practice warning
(Use `node --trace-warnings ...` to show where the warning was created)
gode: AssertionError: ifError got unwanted exception: ENOENT: no such file or directory, open '/tmp/node24tests/test/.tmp.test-process-redirect-warnings/warnings.txt'
      ╭─[ node:assert:1430:1 ]
      │
 1430 │ 
      │ ┬
      │ ╰── ifError got unwanted exception: ENOENT: no such file or directory, open '/tmp/node24tests/test/.tmp.test-process-redirect-warnings/warnings.txt'
──────╯
```

## test-process-threadCpuUsage-worker-threads.js (exit 124)
```

```

## test-process-title-cli.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--title=foo' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:

'node' !== 'foo'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'node' !== 'foo'

─────╯
```

## test-process-uncaught-exception-monitor.js (exit 7)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ ''
- 'Monitored: Shall exit\n'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ ''
- 'Monitored: Shall exit\n'

─────╯
```

## test-process-versions.js (exit 1)
```
gode: Error: Cannot find module '../../deps/acorn/acorn/package.json'
   ╭─[ [gode:require]:1:1 ]
   │
 1 │ 'use strict';
   │ ┬
   │ ╰── Cannot find module '../../deps/acorn/acorn/package.json'
───╯
```
