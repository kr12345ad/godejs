# os: 5 failing official tests

## test-os-checked-function.js (exit 1)
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

## test-os-fast.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings', '--allow-natives-syntax' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-os-fast.js:20:1 ]
    │
 20 │ eval('%PrepareFunctionForOptimization(testFastOs)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-os-process-priority.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

-1 !== 1

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

-1 !== 1

─────╯
```

## test-os-userinfo-handles-getter-errors.js (exit 1)
```
gode: AssertionError: userInfo crashes
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── userInfo crashes
──────╯
```

## test-os.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

-10 !== 10

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

-10 !== 10

─────╯
```
