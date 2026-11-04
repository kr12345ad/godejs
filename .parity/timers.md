# timers: 19 failing official tests

## test-timers-async-store-leak.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--expose-gc' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
[process 1327940]: --- stderr ---
gode: Error: Cannot find module 'internal/async_context_frame'
   ╭─[ [gode:require]:1:39 ]
   │
 1 │ // Flags: --expose-internals --expose-gc
   │                                       ┬
   │                                       ╰── Cannot find module 'internal/async_context_frame'
───╯

[process 1327940]: --- stdout ---

[process 1327940]: status = 1, signal = null
gode: Error: - process terminated with status 1, expected 0
    ╭─[ /tmp/node24tests/test/common/child_process.js:88:26 ]
    │
 88 │     assertStore(timeout, asyncLocalStorage, store);
    │                          ┬
    │                          ╰── - process terminated with status 1, expected 0
────╯
```

## test-timers-destroyed.js (exit 124)
```

```

## test-timers-fast-calls.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--allow-natives-syntax', '--expose-internals', '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-timers-fast-calls.js:18:1 ]
    │
 18 │ eval('%PrepareFunctionForOptimization(testFastCalls)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-timers-immediate-queue-throw.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

0 !== 1

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

0 !== 1

─────╯
```

## test-timers-immediate-unref-nested-once.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref-nested-once.js:8
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref-nested-once.js:8
──────╯
```

## test-timers-immediate-unref-simple.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref-simple.js:13
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref-simple.js:13
──────╯
```

## test-timers-immediate-unref.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref.js:40
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-timers-immediate-unref.js:40
──────╯
```

## test-timers-interval-promisified.js (exit 124)
```
NOTE: The test started as a child_process using these flags: [ '--no-warnings', '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
```

## test-timers-now.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings', '--allow-natives-syntax' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: The expression evaluated to a falsy value:

  assert(binding.getLibuvNow() < 0x3ffffff)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(binding.getLibuvNow() < 0x3ffffff)

──────╯
```

## test-timers-promises-scheduler.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-timers-refresh.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'ERR_INVALID_ARG_TYPE' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'ERR_INVALID_ARG_TYPE' }

──────╯
```

## test-timers-reset-process-domain-on-throw.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ null
- undefined

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ null
- undefined

─────╯
```

## test-timers-this.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ undefined
- Immediate {
-   _argv: undefined,
-   _destroyed: false,
-   _idleNext: null,
-   _idlePrev: null,
-   _onImmediate: [Function (anonymous)]
- }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ undefined
- Immediate {
-   _argv: undefined,
```

## test-timers-unenroll-unref-interval.js (exit 124)
```

```

## test-timers-unref.js (exit 1)
```
gode: AssertionError: Interval should not fire at /tmp/node24tests/test/parallel/test-timers-unref.js:43
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── Interval should not fire at /tmp/node24tests/test/parallel/test-timers-unref.js:43
──────╯
```

## test-timers-unrefed-in-beforeexit.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-timers-unrefed-in-beforeexit.js:6
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-timers-unrefed-in-beforeexit.js:6
──────╯
```

## test-timers-unrefed-in-callback.js (exit 124)
```

```

## test-timers-user-call.js (exit 1)
```
gode: TypeError: fn.apply is not a function. (In 'fn.apply(thisArg, args)', 'fn.apply' is "also not a function")
     ╭─[ [gode:async_hooks]:103:1 ]
     │
 103 │ 
     │ ┬
     │ ╰── fn.apply is not a function. (In 'fn.apply(thisArg, args)', 'fn.apply' is "also not a function")
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-timers.js (exit 124)
```
(node:1329790) TimeoutNaNWarning: NaN is not a number.
Timeout duration was set to 1.
(Use `node --trace-warnings ...` to show where the warning was created)
(node:1329790) TimeoutOverflowWarning: Infinity does not fit into a 32-bit signed integer.
Timeout duration was set to 1.
(node:1329790) TimeoutNegativeWarning: -Infinity is a negative number.
Timeout duration was set to 1.
```
