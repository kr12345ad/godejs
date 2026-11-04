# vm: 54 failing official tests

## test-vm-api-handles-getter-errors.js (exit 1)
```
gode: AssertionError: createScript crashes
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── createScript crashes
──────╯
```

## test-vm-basic.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '[object Object]'
- '[object process]'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '[object Object]'
- '[object process]'

─────╯
```

## test-vm-cached-data.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(!script.cachedDataRejected)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(!script.cachedDataRejected)

──────╯
```

## test-vm-codegen.js (exit 1)
```
gode: TypeError: first argument must be an ArrayBufferView or an ArrayBuffer (evaluating 'new WebAssembly.Module(WASM_BYTES)')
    ╭─[ /tmp/node24tests/test/parallel/test-vm-codegen.js:14:3 ]
    │
 14 │   runInContext(test, ctx);
    │   ┬
    │   ╰── first argument must be an ArrayBufferView or an ArrayBuffer (evaluating 'new WebAssembly.Module(WASM_BYTES)')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-vm-context-async-script.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

undefined !== 3

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

undefined !== 3

─────╯
```

## test-vm-context-dont-contextify.js (exit 1)
```
gode: AssertionError: Values have same structure but are not reference-equal:

{}

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Values have same structure but are not reference-equal:

{}

─────╯
```

## test-vm-context-property-forwarding.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-vm-context.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /expected-filename/. Input:

'runInContext@[native code]\n@/tmp/node24tests/test/parallel/test-vm-context.js:57:18'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /expected-filename/. Input:

'runInContext@[native code]\n@/tmp/node24tests/test/parallel/test-vm-context.js:57:18'

──────╯
```

## test-vm-create-context-accessors.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'test' !== 'ok=test'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'test' !== 'ok=test'

─────╯
```

## test-vm-function-declaration.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'bridged' !== 'b'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'bridged' !== 'b'

─────╯
```

## test-vm-function-redefinition.js (exit 1)
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

## test-vm-getters.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'value', 'writable', 'enumerable', 'configurable' ]
- [ 'get', 'set', 'enumerable', 'configurable' ]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'value', 'writable', 'enumerable', 'configurable' ]
- [ 'get', 'set', 'enumerable', 'configurable' ]

─────╯
```

## test-vm-global-define-property.js (exit 1)
```
gode: AssertionError: Values have same structure but are not reference-equal:

{}

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Values have same structure but are not reference-equal:

{}

─────╯
```

## test-vm-global-get-own.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert.ok(Object.getOwnPropertySymbols(data).includes(property))

    ╭─[ /tmp/node24tests/test/parallel/test-vm-global-get-own.js:33:3 ]
    │
 33 │   vm.runInContext(builder('this'), sandboxContext);
    │   ┬
    │   ╰── The expression evaluated to a falsy value:

  assert.ok(Object.getOwnPropertySymbols(data).includes(property))

────╯
```

## test-vm-global-identity.js (exit 1)
```
gode: AssertionError: Values have same structure but are not reference-equal:

{ window: { window: { window: { window: undefined } } } }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Values have same structure but are not reference-equal:

{ window: { window: { window: { window: undefined } } } }

─────╯
```

## test-vm-global-non-writable-properties.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /Cannot assign to read only property 'x'/. Input:

'TypeError: Attempted to assign to readonly property.'

      ╭─[ node:assert:1263:1 ]
      │
 1263 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /Cannot assign to read only property 'x'/. Input:

'TypeError: Attempted to assign to readonly property.'

──────╯
```

## test-vm-global-property-enumerator.js (exit 1)
```
gode: AssertionError: Case 6 failed: Object.keys
+ actual - expected

+ [ 'length' ]
- []

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Case 6 failed: Object.keys
+ actual - expected

+ [ 'length' ]
- []

─────╯
```

## test-vm-global-property-interceptors.js (exit 1)
```
gode: AssertionError: Expected values to be loosely deep-equal:

{
  a: { configurable: true, enumerable: true, value: 'a', writable: true },
  b: { configurable: true, enumerable: true, value: 'b', writable: true },
  c: { configurable: true, enumerable: true, value: 'c', writable: true },
  d: {
  configurable: false,
  enumerable: false,
  value: undefined,
  writable: false
},
  e: { configurable: true, enumerable: true, value: 'e', writable: true },
  f: {
  configurable: true,
  enumerable: true,
  value: undefined,
  writable: true
},
  g: undefined
}

```

## test-vm-global-property-prototype.js (exit 1)
```
gode: AssertionError: Expected values to be loosely deep-equal:

{
  resultDesc: {
  bothProto: undefined,
  bothProtoGetter: undefined,
  bothProtoIndexed: undefined,
  onInnerProto: undefined,
  onInnerProtoGetter: undefined,
  onInnerProtoIndexed: undefined,
  onOuterProto: undefined,
  onOuterProtoGetter: undefined,
  onOuterProtoIndexed: undefined,
  onSelf: {
  configurable: true,
  enumerable: true,
  value: 'onSelf',
  writable: true
},
  onSelfGetter: {
  configurable: true,
  enumerable: true,
```

## test-vm-global-setter.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

─────╯
```

## test-vm-harmony-symbols.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

─────╯
```

## test-vm-module-basic.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'SourceTextModule {\n  status: \'unlinked\',\n  identifier: \'vm:module(4)\',\n  context: { foo: \'bar\' }\n}'
- 'SourceTextModule {\n  status: \'unlinked\',\n  identifier: \'vm:module(0)\',\n  context: { foo: \'bar\' }\n}'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'SourceTextModule {\n  status: \'unlinked\',\n  identifier: \'vm:module(4)\',\n  context: { foo: \'bar\' }\n}'
- 'SourceTextModule {\n  status: \'unlinked\',\n  identifier: \'vm:module(0)\',\n  context: { foo: \'bar\' }\n}'

─────╯
```

## test-vm-module-cached-data.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-vm-module-errors.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-vm-module-evaluate-source-text-module.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: synchronous source text module
   ╭─[ parallel/test-vm-module-evaluate-source-text-module.js:5:16 ]
   │
 5 │ const common = require('../common');
   │                ┬
   │                ╰── synchronous source text module
───╯
```

## test-vm-module-evaluate-synthethic-module-rejection.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: Module status unlinked does not match the expected status linked
    ╭─[ [gode:vm modules]:65:1 ]
    │
 65 │ 
    │ ┬
    │ ╰── Module status unlinked does not match the expected status linked
────╯
```

## test-vm-module-evaluate-synthethic-module.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: Module status unlinked does not match the expected status linked
    ╭─[ [gode:vm modules]:65:22 ]
    │
 65 │     throw new Error('synchronous synthethic module');
    │                      ┬
    │                      ╰── Module status unlinked does not match the expected status linked
────╯
```

## test-vm-module-evaluate-while-evaluating.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: Module status unlinked does not match the expected status linked
    ╭─[ [gode:vm modules]:65:1 ]
    │
 65 │ 
    │ ┬
    │ ╰── Module status unlinked does not match the expected status linked
────╯
```

## test-vm-module-hasasyncgraph.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - module is not instantiated yet
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception.
    @/tmp/node24tests/test/parallel/test-vm-module-hasasyncgraph.js:17:16
  code: ERR_ASSERTION
  ...
ok 2 - simple module with top-level await
not ok 3 - simple module with non top-level await
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    true !== false
    
    @/tmp/node24tests/test/parallel/test-vm-module-hasasyncgraph.js:47:21
  code: ERR_ASSERTION
  ...
not ok 4 - module with a dependency containing top-level await
  ---
```

## test-vm-module-hastoplevelawait.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
ok 1 - simple module
ok 2 - simple module with top-level await
not ok 3 - simple module with non top-level await
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    true !== false
    
    @/tmp/node24tests/test/parallel/test-vm-module-hastoplevelawait.js:39:21
  code: ERR_ASSERTION
  ...
ok 4 - module with a dependency containing top-level await
1..4
# tests 4
# pass 3
# fail 1
```

## test-vm-module-instantiate.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - simple module
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly deep-equal:
    + actual - expected
    
    + []
    - [ 'default', 'foo', Symbol(Symbol.toStringTag) ]
    
    @/tmp/node24tests/test/parallel/test-vm-module-instantiate.js:20:25
  code: ERR_ASSERTION
  ...
not ok 2 - linkRequests can not be skipped
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception.
    @/tmp/node24tests/test/parallel/test-vm-module-instantiate.js:31:16
  code: ERR_ASSERTION
  ...
not ok 3 - re-export simple name
```

## test-vm-module-linkmodulerequests-circular.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules', '--js-source-phase-imports' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - basic circular linking
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    + actual - expected
    
    + 'unlinked'
    - 'linked'
    
    circular@/tmp/node24tests/test/parallel/test-vm-module-linkmodulerequests-circular.js:29:21
  code: ERR_ASSERTION
  ...
not ok 2 - circular linking graph
  ---
  error: |-
    ReferenceError: Can't find variable: a
    anonymous@
    @[gode:vm modules]:131:7
    evaluate@[gode:vm modules]:69:21
    circular2@/tmp/node24tests/test/parallel/test-vm-module-linkmodulerequests-circular.js:71:28
```

## test-vm-module-linkmodulerequests-deep.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules', '--js-source-phase-imports' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - deep linking
  ---
  error: |-
    ReferenceError: Can't find variable: baz
    anonymous@
    @[gode:vm modules]:131:7
    evaluate@[gode:vm modules]:69:21
    depth@/tmp/node24tests/test/parallel/test-vm-module-linkmodulerequests-deep.js:31:22
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
1..1
# tests 1
```

## test-vm-module-linkmodulerequests.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules', '--js-source-phase-imports' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - simple graph
  ---
  error: |-
    ReferenceError: Can't find variable: five
    anonymous@
    @[gode:vm modules]:131:7
    evaluate@[gode:vm modules]:69:21
    simple@/tmp/node24tests/test/parallel/test-vm-module-linkmodulerequests.js:22:21
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
not ok 2 - invalid link values
  ---
```

## test-vm-module-modulerequests.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--experimental-vm-modules', '--js-source-phase-imports' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
TAP version 13
not ok 1 - SourceTextModule.moduleRequests should return module requests
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
    
    11 !== 6
    
    @/tmp/node24tests/test/parallel/test-vm-module-modulerequests.js:29:21
  code: ERR_ASSERTION
  ...
not ok 2 - SourceTextModule.moduleRequests items are frozen
  ---
  error: |-
    AssertionError [ERR_ASSERTION]: Missing expected exception (TypeError).
    @/tmp/node24tests/test/parallel/test-vm-module-modulerequests.js:105:18
  code: ERR_ASSERTION
  ...
1..2
# tests 2
# pass 0
```

## test-vm-new-script-new-context.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

1 !== 2

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 2

─────╯
```

## test-vm-not-strict.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'direct', 'this' ]
- [ 'direct', 'this', 'new' ]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'direct', 'this' ]
- [ 'direct', 'this', 'new' ]

─────╯
```

## test-vm-options-validation.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-vm-ownkeys.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'a', 'b' ]
- [ 'a', 'b', Symbol(1), Symbol(2) ]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'a', 'b' ]
- [ 'a', 'b', Symbol(1), Symbol(2) ]

─────╯
```

## test-vm-ownpropertysymbols.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ []
- [ Symbol(1), Symbol(2) ]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ []
- [ Symbol(1), Symbol(2) ]

─────╯
```

## test-vm-preserves-property.js (exit 1)
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

## test-vm-property-not-on-sandbox.js (exit 1)
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

## test-vm-proxies.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'undefined'
- 'function'

─────╯
```

## test-vm-run-in-new-context.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-gc' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:

1 !== 2

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 2

─────╯
```

## test-vm-set-property-proxy.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-vm-sigint-existing-handler.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_IPC_CHANNEL_CLOSED', message: 'Channel closed' }
- Comparison {
-   code: 'ERR_SCRIPT_EXECUTION_INTERRUPTED',
-   message: 'Script execution was interrupted by `SIGINT`'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_IPC_CHANNEL_CLOSED', message: 'Channel closed' }
- Comparison {
-   code: 'ERR_SCRIPT_EXECUTION_INTERRUPTED',
-   message: 'Script execution was interrupted by `SIGINT`'
- }

```

## test-vm-sigint.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   message: 'process.send is not a function. (In \'process.send(\'runInContext\')\', \'process.send\' is undefined)'
-   code: 'ERR_SCRIPT_EXECUTION_INTERRUPTED',
-   message: 'Script execution was interrupted by `SIGINT`'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   message: 'process.send is not a function. (In \'process.send(\'runInContext\')\', \'process.send\' is undefined)'
-   code: 'ERR_SCRIPT_EXECUTION_INTERRUPTED',
-   message: 'Script execution was interrupted by `SIGINT`'
  }

```

## test-vm-strict-assign.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /Cannot assign to read only property 'x'/. Input:

'TypeError: Attempted to assign to readonly property.'

      ╭─[ node:assert:1263:1 ]
      │
 1263 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /Cannot assign to read only property 'x'/. Input:

'TypeError: Attempted to assign to readonly property.'

──────╯
```

## test-vm-symbols.js (exit 1)
```
gode: TypeError: this.getSymbolValue is not a function. (In 'this.getSymbolValue()', 'this.getSymbolValue' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-vm-symbols.js:23:1 ]
    │
 23 │ assert.strictEqual(vm.runInContext('this.getSymbolValue()', context), 'foo');
    │ ┬
    │ ╰── this.getSymbolValue is not a function. (In 'this.getSymbolValue()', 'this.getSymbolValue' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-vm-syntax-error-stderr.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /BEGIN CERT/. Input:

'gode: Error: Cannot find module \'undefined\'\n   ╭─[ [gode:require]:1:1 ]\n   │\n 1 │ require(process.argv[1]);\n   │ ┬\n   │ ╰── Cannot find module \'undefined\'\n───╯\n'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /BEGIN CERT/. Input:

'gode: Error: Cannot find module \'undefined\'\n   ╭─[ [gode:require]:1:1 ]\n   │\n 1 │ require(process.argv[1]);\n   │ ┬\n   │ ╰── Cannot find module \'undefined\'\n───╯\n'

──────╯
```

## test-vm-timeout-escape-promise-2.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-vm-timeout-escape-promise.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'escaped timeout at 2000 milliseconds!' }
- Comparison {
-   code: 'ERR_SCRIPT_EXECUTION_TIMEOUT',
-   message: 'Script execution timed out after 5ms'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'escaped timeout at 2000 milliseconds!' }
- Comparison {
-   code: 'ERR_SCRIPT_EXECUTION_TIMEOUT',
-   message: 'Script execution timed out after 5ms'
- }

```

## test-vm-timeout.js (exit 124)
```

```

## test-vm-util-lazy-properties.js (exit 1)
```
gode: AssertionError: Expected "actual" to be reference-equal to "expected":
+ actual - expected

+ [Function: bridged]
- [Function: parseArgs]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" to be reference-equal to "expected":
+ actual - expected

+ [Function: bridged]
- [Function: parseArgs]

─────╯
```
