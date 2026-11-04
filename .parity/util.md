# util: 13 failing official tests

## test-util-callbackify.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ ''
- 'ifError got unwanted exception: /tmp/node24tests/test/fixtures/uncaught-exceptions/callbackify2.js'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ ''
- 'ifError got unwanted exception: /tmp/node24tests/test/fixtures/uncaught-exceptions/callbackify2.js'

─────╯
```

## test-util-exception-with-host-port.js (exit 1)
```
gode: AssertionError: expected local info in "connect ENOENT 127.0.0.1:80 - 0.0.0.0:12345"
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── expected local info in "connect ENOENT 127.0.0.1:80 - 0.0.0.0:12345"
──────╯
```

## test-util-format.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '118059162071741140000 118059162071741140000 123123123'
- '118_059_162_071_741_140_000 118_059_162_071_741_140_000 123_123_123'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '118059162071741140000 118059162071741140000 123123123'
- '118_059_162_071_741_140_000 118_059_162_071_741_140_000 123_123_123'

─────╯
```

## test-util-getcallsites-sourcemap.js (exit 1)
```
gode: AssertionError: expected original source in scriptName, got "/tmp/node24tests/test/fixtures/source-map/get-call-sites-mapped.js"
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── expected original source in scriptName, got "/tmp/node24tests/test/fixtures/source-map/get-call-sites-mapped.js"
──────╯
```

## test-util-inspect-long-running.js (exit 124)
```

```

## test-util-inspect-proxy.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: undefined is not an object (evaluating 'details[0]')
    ╭─[ /tmp/node24tests/test/parallel/test-util-inspect-proxy.js:50:1 ]
    │
 50 │ assert.strictEqual(target, details[0]);
    │ ┬
    │ ╰── undefined is not an object (evaluating 'details[0]')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-util-inspect.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '[GeneratorFunction (anonymous)]'
- '[GeneratorFunction (anonymous)] AsyncFunction'

     ╭─[ node:assert:916:3 ]
     │
 916 │   );
     │   ┬
     │   ╰── Expected values to be strictly equal:
+ actual - expected

+ '[GeneratorFunction (anonymous)]'
- '[GeneratorFunction (anonymous)] AsyncFunction'

─────╯
```

## test-util-internal.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: Cannot destructure property 'arrow_message_private_symbol' from null or undefined value
    ╭─[ /tmp/node24tests/test/parallel/test-util-internal.js:13:1 ]
    │
 13 │ } = internalBinding('util');
    │ ┬
    │ ╰── Cannot destructure property 'arrow_message_private_symbol' from null or undefined value
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-util-isDeepStrictEqual.js (exit 1)
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

## test-util-promisify.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected "actual" not to be reference-equal to "expected":

[Function (anonymous)]
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" not to be reference-equal to "expected":

[Function (anonymous)]
─────╯
```

## test-util-sigint-watchdog.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - [object Object] > with no signal observed
  ---
  error: |-
    TypeError: binding.startSigintWatchdog is not a function. (In 'binding.startSigintWatchdog()', 'binding.startSigintWatchdog' is undefined)
    @/tmp/node24tests/test/parallel/test-util-sigint-watchdog.js:16:32
    runTest@[gode:test]:250:40
    @[gode:test]:165:18
    runHooks@[gode:test]:149:9
    step@[gode:test]:164:17
    @[gode:test]:170:11
    runHooks@[gode:test]:149:9
    runSuite@[gode:test]:153:13
    step@[gode:test]:161:19
    @[gode:test]:170:11
    runHooks@[gode:test]:149:9
    runSuite@[gode:test]:153:13
    pump@[gode:test]:109:13
  code: ERR_TEST_FAILURE
  ...
not ok 2 - [object Object] > with one call to the watchdog, one signal
```

## test-util-types.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [
  '--experimental-vm-modules',
  '--expose-internals',
  '--allow-natives-syntax'
] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Want undefined to match isExternal
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── Want undefined to match isExternal
──────╯
```

## test-util.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: Error is not a constructor (evaluating 'new (context('Error'))()')
    ╭─[ /tmp/node24tests/test/parallel/test-util.js:58:3 ]
    │
 58 │   assert.strictEqual(util.types.isNativeError(new (context('Error'))()), true);
    │   ┬
    │   ╰── Error is not a constructor (evaluating 'new (context('Error'))()')
    │
    │ Help: the value does not have the type this operation requires
────╯
```
