# zlib: 14 failing official tests

## test-zlib-brotli-16GB.js (exit 124)
```

```

## test-zlib-brotli-dictionary.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-zlib-brotli-kmaxlength-rangeerror.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-zlib-flush.js (exit 1)
```
TAP version 13
not ok 1 - zlib flush
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
    + <Buffer 78 01 00 10 00 ef ff ff d8 ff e0 00 10 4a 46 49 46 00 01 01 01 00 48 01 00 00 ff ff 40 8a 05 31>
    - <Buffer 78 01>
    
    @/tmp/node24tests/test/parallel/test-zlib-flush.js:36:25
  code: ERR_ASSERTION
  ...
1..1
# tests 1
# pass 0
# fail 1
```

## test-zlib-invalid-input.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-zlib-kmaxlength-rangeerror.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-zlib-premature-end.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

43 !== 18

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

43 !== 18

─────╯
```

## test-zlib-reject-garbage-after-end.js (exit 1)
```
TAP version 13
ok 1 - rejectGarbageAfterEnd rejects trailing input for inflate
ok 2 - rejectGarbageAfterEnd rejects trailing input for inflateRaw
ok 3 - rejectGarbageAfterEnd rejects trailing input for gunzip
ok 4 - rejectGarbageAfterEnd rejects trailing input for unzip
ok 5 - rejectGarbageAfterEnd rejects trailing input for brotli
ok 6 - rejectGarbageAfterEnd rejects trailing input for zstd
ok 7 - zstd decompresses concatenated frames regardless of chunking
ok 8 - zstd trailing junk handling is independent of chunking
ok 9 - zstd handles incomplete trailing frame identifiers as junk
not ok 10 - zstd reports errors in subsequent frames
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
    + Comparison { code: 'Z_DATA_ERROR' }
    - Comparison { code: 'ZSTD_error_checksum_wrong' }
    
    @/tmp/node24tests/test/parallel/test-zlib-reject-garbage-after-end.js:233:16
  code: ERR_ASSERTION
  ...
```

## test-zlib-truncated.js (exit 1)
```
gode: Error: unexpected end of file
    ╭─[ /tmp/node24tests/test/parallel/test-zlib-truncated.js:60:5 ]
    │
 60 │     const result = toUTF8(zlib[methods.decompSync](truncated, partialFlushOpt));
    │     ┬
    │     ╰── unexpected end of file
────╯
```

## test-zlib-type-error.js (exit 1)
```
TAP version 13
not ok 1 - DecompressionStream deflate emits TypeError on trailing data
  ---
  error: |-
    TypeError: undefined is not a constructor (evaluating 'new DecompressionStream(format)')
    assertDecompressionStreamRejects@/tmp/node24tests/test/parallel/test-zlib-type-error.js:11:68
    @/tmp/node24tests/test/parallel/test-zlib-type-error.js:24:43
    runTest@[gode:test]:250:74
    @[gode:test]:165:18
    runHooks@[gode:test]:149:9
    step@[gode:test]:164:17
    @[gode:test]:170:11
    runHooks@[gode:test]:149:9
    runSuite@[gode:test]:153:13
    pump@[gode:test]:109:13
  code: ERR_TEST_FAILURE
  ...
not ok 2 - DecompressionStream gzip emits TypeError on trailing data
  ---
  error: |-
    TypeError: undefined is not a constructor (evaluating 'new DecompressionStream(format)')
    assertDecompressionStreamRejects@/tmp/node24tests/test/parallel/test-zlib-type-error.js:11:68
```

## test-zlib-unused-weak.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-gc', '--no-concurrent-array-buffer-sweeping' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected after-GC delta 0 to be less than 5 % of before-GC delta 0
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── Expected after-GC delta 0 to be less than 5 % of before-GC delta 0
──────╯
```

## test-zlib-write-after-flush.js (exit 1)
```
gode: AssertionError: function should not have been called at /tmp/node24tests/test/parallel/test-zlib-write-after-flush.js:42
called with arguments: Error: brotli: excessive input
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── function should not have been called at /tmp/node24tests/test/parallel/test-zlib-write-after-flush.js:42
called with arguments: Error: brotli: excessive input
──────╯
```

## test-zlib-zstd-kmaxlength-rangeerror.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-zlib-zstd-reset.js (exit 1)
```
TAP version 13
ok 1 - ZstdCompress reset preserves its initial options
ok 2 - ZstdDecompress reset preserves its dictionary
not ok 3 - ZstdDecompress reset preserves its parameters
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected rejection.
  code: ERR_ASSERTION
  ...
1..3
# tests 3
# pass 2
# fail 1
```
