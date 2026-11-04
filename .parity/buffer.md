# buffer: 20 failing official tests

## test-buffer-alloc.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'length larger than (2 ** 53) - 1' }
- Comparison { message: 'Invalid typed array length: 9007199254740992' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ assert.throws(() => {
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'length larger than (2 ** 53) - 1' }
- Comparison { message: 'Invalid typed array length: 9007199254740992' }

──────╯
```

## test-buffer-backing-arraybuffer.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: mismatch: false vs true for Uint8Array, length = 0

false !== true

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── mismatch: false vs true for Uint8Array, length = 0

false !== true

─────╯
```

## test-buffer-bytelength.js (exit 1)
```
gode: TypeError: The "string" argument must be of type string or an instance of Buffer or ArrayBuffer. Received an instance of Object
     ╭─[ /tmp/node24tests/test/parallel/test-buffer-bytelength.js:119:1 ]
     │
 119 │ assert.strictEqual(Buffer.byteLength(arrayBuf), 0);
     │ ┬
     │ ╰── The "string" argument must be of type string or an instance of Buffer or ArrayBuffer. Received an instance of Object
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-buffer-constants.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-buffer-constructor-node-modules-paths.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'gode: ReferenceError: Can\'t find variable: vm\n   ╭─[ [eval]:4:3 ]\n   │\n 4 │   vm.runInNewContext(\'new Buffer(10)\', { Buffer }, {\n   │   ┬\n   │   ╰── Can\'t find variable: vm\n   │\n   │ Help: check for a typo, or declare the identifier before using it\n───╯'
- ''

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'gode: ReferenceError: Can\'t find variable: vm\n   ╭─[ [eval]:4:3 ]\n   │\n 4 │   vm.runInNewContext(\'new Buffer(10)\', { Buffer }, {\n   │   ┬\n   │   ╰── Can\'t find variable: vm\n   │\n   │ Help: check for a typo, or declare the identifier before using it\n───╯'
- ''

─────╯
```

## test-buffer-constructor-outside-node-modules.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: function is not a constructor (evaluating 'new Buffer(10)')
    ╭─[ /tmp/node24tests/test/parallel/test-buffer-constructor-outside-node-modules.js:21:1 ]
    │
 21 │ vm.runInNewContext('new Buffer(10)', { Buffer }, {
    │ ┬
    │ ╰── function is not a constructor (evaluating 'new Buffer(10)')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-buffer-fill.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'ERR_OUT_OF_RANGE' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'ERR_OUT_OF_RANGE' }

──────╯
```

## test-buffer-generic-methods.js (exit 1)
```
gode: TypeError: Receiver should be a typed array view
   ╭─[ [gode:buffer/legacy]:4:1 ]
   │
 4 │ 
   │ ┬
   │ ╰── Receiver should be a typed array view
   │
   │ Help: the value does not have the type this operation requires
───╯
```

## test-buffer-indexof.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   message: 'function is not a constructor (evaluating \'new buffer.Buffer.prototype.lastIndexOf(1, \'str\')\')',
-   code: 'ERR_INVALID_ARG_TYPE',
-   message: 'The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received an instance of lastIndexOf',
    name: 'TypeError'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   message: 'function is not a constructor (evaluating \'new buffer.Buffer.prototype.lastIndexOf(1, \'str\')\')',
-   code: 'ERR_INVALID_ARG_TYPE',
-   message: 'The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received an instance of lastIndexOf',
    name: 'TypeError'
```

## test-buffer-inspect.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '<Buffer 31 32 33 34>'
- '<Buffer 31 32 ... 2 more bytes>'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '<Buffer 31 32 33 34>'
- '<Buffer 31 32 ... 2 more bytes>'

─────╯
```

## test-buffer-isascii.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

true !== false

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

true !== false

─────╯
```

## test-buffer-isutf8-isascii-fast.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings', '--allow-natives-syntax' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-buffer-isutf8-isascii-fast.js:19:1 ]
    │
 19 │ eval('%PrepareFunctionForOptimization(isAscii)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-buffer-isutf8.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

true !== false

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

true !== false

─────╯
```

## test-buffer-pool-untransferable.js (exit 1)
```
gode: AssertionError: Values have same structure but are not reference-equal:

ArrayBuffer { byteLength: 11 }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Values have same structure but are not reference-equal:

ArrayBuffer { byteLength: 11 }

─────╯
```

## test-buffer-prototype-inspect.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /^<Buffer (?:78 ){50}\.\.\. 1 more byte>$/. Input:

'<Buffer 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 ... 1 more bytes>'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /^<Buffer (?:78 ){50}\.\.\. 1 more byte>$/. Input:

'<Buffer 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 78 ... 1 more bytes>'

──────╯
```

## test-buffer-resizable.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - Using resizable ArrayBuffer with Buffer... > works as expected
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    9 !== 14
    
    @/tmp/node24tests/test/parallel/test-buffer-resizable.js:15:23
  code: ERR_ASSERTION
  ...
not ok 2 - Using resizable ArrayBuffer with Buffer... > works with the deprecated constructor also
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    9 !== 14
    
    @/tmp/node24tests/test/parallel/test-buffer-resizable.js:25:23
  code: ERR_ASSERTION
  ...
```

## test-buffer-swap-fast.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings', '--allow-natives-syntax' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: SyntaxError: Unexpected token '%'
    ╭─[ /tmp/node24tests/test/parallel/test-buffer-swap-fast.js:34:1 ]
    │
 34 │ eval('%PrepareFunctionForOptimization(Buffer.prototype.swap16)');
    │ ┬
    │ ╰── Unexpected token '%'
    │
    │ Help: the file could not be parsed; check the highlighted line
────╯
```

## test-buffer-tostring-4gb.js (exit 2)
```
SIGSEGV: segmentation violation
PC=0x7f696e165e33 m=5 sigcode=1 addr=0x0
signal arrived during cgo execution

goroutine 1 gp=0x3d7bc3db41e0 m=5 mp=0x3d7bc3e80008 [syscall, locked to thread]:
runtime.cgocall(0x8975f0, 0x3d7bc3ded4f0)
	/usr/lib/golang/src/runtime/cgocall.go:167 +0x4b fp=0x3d7bc3ded4c8 sp=0x3d7bc3ded490 pc=0x481f4b
godejs/internal/js._Cfunc_JSObjectSetPrototype(0x7f6913220088, 0x0, 0x7f69130be3c0)
	_cgo_gotypes.go:357 +0x45 fp=0x3d7bc3ded4f0 sp=0x3d7bc3ded4c8 pc=0x4fd285
godejs/internal/js.Object.SetProto({{0x3d7bc3eaa5a0?, 0x3d7bc4482000?}}, {{0x3d7bc3eaa5a0?, 0x0?}})
	/home/krishnaa/godejs/internal/js/value.go:581 +0x3b fp=0x3d7bc3ded520 sp=0x3d7bc3ded4f0 pc=0x505e9b
godejs/modules/buffer.(*module).newBuffer(0x3d7bc3eee0a0, {0x3d7bc4482000?, 0x48c5da?, 0x3d7bc3ded5e0?})
	/home/krishnaa/godejs/modules/buffer/buffer.go:1472 +0x89 fp=0x3d7bc3ded590 sp=0x3d7bc3ded520 pc=0x559d69
godejs/modules/buffer.(*module).alloc(0x3d7bc3eee0a0, {0x3d7bc41b8770, 0x1, 0x3d7bc41b8770?})
	/home/krishnaa/godejs/modules/buffer/buffer.go:2051 +0x190 fp=0x3d7bc3ded608 sp=0x3d7bc3ded590 pc=0x55dcb0
godejs/modules/buffer.(*module).installStatics.func5({{0x10?, 0x943920?}}, {0x3d7bc41b8770?, 0xfffe000000000078?, 0x3d7bc3ded680?})
	/home/krishnaa/godejs/modules/buffer/buffer.go:251 +0x25 fp=0x3d7bc3ded638 sp=0x3d7bc3ded608 pc=0x54bae5
godejs/internal/js.godeCallback.func1(0x3d7bc3ded748, 0x3d7bc3ded720, 0x78?, 0x3d7bc3ded710?, 0x7f696a865060?, {0x3d7bc41b8770?, 0x3d7bc3ef16e0?, 0x7f696a865060?})
	/home/krishnaa/godejs/internal/js/callback.go:52 +0x69 fp=0x3d7bc3ded690 sp=0x3d7bc3ded638 pc=0x4ff6a9
godejs/internal/js.godeCallback(0x7f6913220088, 0x7f69130e9d00, 0x7f69132be0c0, 0x1, 0x7f692245b850, 0x7f692245b7c0)
	/home/krishnaa/godejs/internal/js/callback.go:53 +0x255 fp=0x3d7bc3ded770 sp=0x3d7bc3ded690 pc=0x4ff315
_cgoexp_ac897f297743_godeCallback(0x7f692245b730)
```

## test-buffer-write-fast.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings', '--allow-natives-syntax' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: Cannot find module 'internal/buffer'
   ╭─[ [gode:require]:1:60 ]
   │
 1 │ // Flags: --expose-internals --no-warnings --allow-natives-syntax
   │                                                            ┬
   │                                                            ╰── Cannot find module 'internal/buffer'
───╯
```

## test-buffer-write.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_OUT_OF_RANGE',
+   message: 'The value of "offset" is out of range. It must be >= 0 and <= 9. Received -1',
-   message: 'The value of "offset" is out of range. It must be >= 0 && <= 9. Received -1',
    name: 'RangeError'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_OUT_OF_RANGE',
+   message: 'The value of "offset" is out of range. It must be >= 0 and <= 9. Received -1',
-   message: 'The value of "offset" is out of range. It must be >= 0 && <= 9. Received -1',
    name: 'RangeError'
```
