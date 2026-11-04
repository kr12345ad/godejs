# assert: 7 failing official tests

## test-assert-class.js (exit 1)
```
TAP version 13
ok 1 - Assert constructor requires new
not ok 2 - Assert class non strict
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
        code: 'ERR_ASSERTION',
    +   message: 'custom message\n+ actual - expected\n\n+ { a: true }\n- { a: false }\n',
    -   message: 'custom message\n+ actual - expected\n\n  {\n+   a: true\n-   a: false\n  }\n',
        name: 'AssertionError'
      }
    
    @/tmp/node24tests/test/parallel/test-assert-class.js:66:24
  code: ERR_ASSERTION
  ...
ok 3 - Assert class strict
ok 4 - Assert class with invalid diff option
not ok 5 - Assert class non strict with full diff
  ---
```

## test-assert-deep-with-error.js (exit 1)
```
TAP version 13
not ok 1 - Handle error causes
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
    +   message: 'Values have same structure but are not reference-equal:\n\nError {}\n'
    -   message: 'Expected values to be strictly deep-equal:\n+ actual - expected\n\n  [Error: a] {\n+   [cause]: [Error: x]\n-   [cause]: [Error: y]\n  }\n'
      }
    
    @/tmp/node24tests/test/parallel/test-assert-deep-with-error.js:24:16
  code: ERR_ASSERTION
  ...
not ok 2 - Handle undefined causes
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
```

## test-assert-deep.js (exit 1)
```
TAP version 13
not ok 1 - deepEqual
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
        code: 'ERR_ASSERTION',
    +   message: 'Expected values to be strictly deep-equal:\n+ actual - expected\n\n+ Uint8Array(4) [ 120, 121, 122, 10 ]\n- <Buffer 78 79 7a 0a>\n'
    -   message: 'Expected values to be strictly deep-equal:\n+ actual - expected\n\n+ Uint8Array(4) [\n- Buffer(4) [Uint8Array] [\n    120,\n    121,\n    122,\n    10\n  ]\n'
      }
    
    @/tmp/node24tests/test/parallel/test-assert-deep.js:68:16
  code: ERR_ASSERTION
  ...
ok 2 - loose deepEqual
not ok 3 - date
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
```

## test-assert-first-line.js (exit 1)
```
TAP version 13
not ok 1 - Verify that asserting in the very first line produces the expected result
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
    +   message: 'The expression evaluated to a falsy value:\n\n  \'use strict\'; const ässört = require(\'assert\'); ässört(true); ässört.ok(\'\'); ässört(null);\n',
    -   message: 'The expression evaluated to a falsy value:\n\n  ässört.ok(\'\')\n',
        name: 'AssertionError'
      }
    
    @/tmp/node24tests/test/parallel/test-assert-first-line.js:11:16
  code: ERR_ASSERTION
  ...
1..1
# tests 1
# pass 0
# fail 1
```

## test-assert-partial-deep-equal.js (exit 1)
```
TAP version 13
not ok 1 - Object Comparison Tests > partialDeepStrictEqual > throws an error > describe block
  ---
  error: |-
    TypeError: function is not a constructor (evaluating 'new vm.runInNewContext('new Set(["value1", "value2"])')')
    runInNewContext@[native code]
    @/tmp/node24tests/test/parallel/test-assert-partial-deep-equal.js:320:41
    describe@[gode:test]:661:44
    @/tmp/node24tests/test/parallel/test-assert-partial-deep-equal.js:42:13
    describe@[gode:test]:661:44
    @/tmp/node24tests/test/parallel/test-assert-partial-deep-equal.js:41:11
    describe@[gode:test]:661:44
    @/tmp/node24tests/test/parallel/test-assert-partial-deep-equal.js:40:9
  code: ERR_TEST_FAILURE
  ...
not ok 2 - Object Comparison Tests > does not throw an error > describe block
  ---
  error: |-
    TypeError: function is not a constructor (evaluating 'new vm.runInNewContext(
              'new Map([["key1", "value1"], ["key2", "value2"]])'
            )')
    runInNewContext@[native code]
```

## test-assert-typedarray-deepequal.js (exit 1)
```
TAP version 13
ok 1 - equalArrayPairs
ok 2 - equalArrayPairs
ok 3 - equalArrayPairs
ok 4 - equalArrayPairs
ok 5 - equalArrayPairs
ok 6 - equalArrayPairs
ok 7 - equalArrayPairs
ok 8 - equalArrayPairs
ok 9 - equalArrayPairs
ok 10 - equalArrayPairs
ok 11 - equalArrayPairs
ok 12 - equalArrayPairs
ok 13 - equalArrayPairs
ok 14 - equalArrayPairs
ok 15 - equalArrayPairs
ok 16 - equalArrayPairs
ok 17 - looseEqualArrayPairs
ok 18 - looseEqualArrayPairs
ok 19 - looseEqualArrayPairs
not ok 20 - notEqualArrayPairs
  ---
```

## test-assert.js (exit 1)
```
TAP version 13
ok 1 - some basics
ok 2 - Throw message if the message is instanceof Error
ok 3 - Errors created in different contexts are handled as any other custom error
not ok 4 - assert.throws()
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
      Comparison {
    +   message: 'Cannot call a class constructor AssertionError without |new|',
    +   name: 'TypeError'
    -   actual: TypeError: [object Object],
    -   code: 'ERR_ASSERTION',
    -   expected: [class AssertionError],
    -   generatedMessage: true,
    -   message: 'The error is expected to be an instance of "AssertionError". Received "TypeError"\n\nError message:\n\n[object Object]',
    -   name: 'AssertionError',
    -   operator: 'throws'
      }
    
```
