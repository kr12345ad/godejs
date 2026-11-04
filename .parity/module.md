# module: 22 failing official tests

## test-module-builtin.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ [
+   'internal/assert',
+   'internal/assert/myers_diff',
+   'internal/child_process',
+   'internal/crypto/keys',
+   'internal/crypto/util',
+   'internal/crypto/x509',
+   'internal/dgram',
+   'internal/errors',
+   'internal/event_target',
+   'internal/fs/promises',
+   'internal/fs/recursive_watch',
+   'internal/fs/utils',
+   'internal/fs/watchers',
+   'internal/js_stream_socket',
+   'internal/linkedlist',
+   'internal/options',
+   'internal/root',
+   'internal/streams/consumers',
```

## test-module-cache.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'MODULE_NOT_FOUND' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison {}
- Comparison { code: 'MODULE_NOT_FOUND' }

──────╯
```

## test-module-children.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--no-deprecation' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
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

## test-module-circular-symlinks.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert.ok(!obj.b.a.b)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert.ok(!obj.b.a.b)

──────╯
```

## test-module-create-require.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'filename\' must be a file URL object, file URL string, or absolute path string. Received \'[object Object]\''
-   message: 'The argument \'filename\' must be a file URL object, file URL string, or absolute path string. Received {}'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'filename\' must be a file URL object, file URL string, or absolute path string. Received \'[object Object]\''
-   message: 'The argument \'filename\' must be a file URL object, file URL string, or absolute path string. Received {}'
  }

```

## test-module-globalpaths-nodepath.js (exit 1)
```
gode: TypeError: mod._initPaths is not a function. (In 'mod._initPaths()', 'mod._initPaths' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-module-globalpaths-nodepath.js:40:1 ]
    │
 40 │ mod._initPaths();
    │ ┬
    │ ╰── mod._initPaths is not a function. (In 'mod._initPaths()', 'mod._initPaths' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-module-isBuiltin.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(!isBuiltin('internal/errors'))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(!isBuiltin('internal/errors'))

──────╯
```

## test-module-loading-error.js (exit 1)
```
gode: AssertionError: The validation function is expected to return "true". Received false

Caught error:

ReferenceError: Can't find variable: foobar
      ╭─[ node:assert:1263:1 ]
      │
 1263 │ 
      │ ┬
      │ ╰── The validation function is expected to return "true". Received false

Caught error:

ReferenceError: Can't find variable: foobar
──────╯
```

## test-module-loading-globalpaths.js (exit 1)
```
gode: Error: Command failed: /tmp/node24tests/test/.tmp.test-module-loading-globalpaths/install/bin/gode /tmp/node24tests/test/parallel/test-module-loading-globalpaths.js child
gode: Error: Cannot find module 'foo'
   ╭─[ [gode:require]:1:1 ]
   │
 1 │ 'use strict';
   │ ┬
   │ ╰── Cannot find module 'foo'
───╯
    ╭─[ child_process.js:78:3 ]
    │
 78 │   fs.mkdirSync(prefixLibPath);
    │   ┬
    │   ╰── Command failed: /tmp/node24tests/test/.tmp.test-module-loading-globalpaths/install/bin/gode /tmp/node24tests/test/parallel/test-module-loading-globalpaths.js child
gode: Error: Cannot find module 'foo'
   ╭─[ [gode:require]:1:1 ]
   │
 1 │ 'use strict';
   │ ┬
   │ ╰── Cannot find module 'foo'
───╯
────╯
```

## test-module-main-fail.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /MODULE_NOT_FOUND/. Input:

'Error: Command failed: /tmp/gode iDoNotExist\ngode: Cannot find module \'/tmp/node24tests/test/iDoNotExist\'\n   ╭─[ iDoNotExist:1:1 ]\n   │\n 1 │ \n   │ ┬\n   │ ╰── Cannot find module \'/tmp/node24tests/test/iDoNotExist\'\n───╯'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /MODULE_NOT_FOUND/. Input:

'Error: Command failed: /tmp/gode iDoNotExist\ngode: Cannot find module \'/tmp/node24tests/test/iDoNotExist\'\n   ╭─[ iDoNotExist:1:1 ]\n   │\n 1 │ \n   │ ┬\n   │ ╰── Cannot find module \'/tmp/node24tests/test/iDoNotExist\'\n───╯'

──────╯
```

## test-module-main-preserve-symlinks-fail.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(e.toString().match(/Error: Cannot find module/))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(e.toString().match(/Error: Cannot find module/))

──────╯
```

## test-module-multi-extensions.js (exit 1)
```
gode: TypeError: undefined is not an object (evaluating 'require.extensions['.bar'] = common.mustNotCall()')
    ╭─[ /tmp/node24tests/test/parallel/test-module-multi-extensions.js:20:3 ]
    │
 20 │   require.extensions['.bar'] = common.mustNotCall();
    │   ┬
    │   ╰── undefined is not an object (evaluating 'require.extensions['.bar'] = common.mustNotCall()')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-module-nearest-parent-package-json-cache.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: Error: Cannot find module 'internal/modules/package_json_reader'
   ╭─[ [gode:require]:1:1 ]
   │
 1 │ 'use strict';
   │ ┬
   │ ╰── Cannot find module 'internal/modules/package_json_reader'
───╯
```

## test-module-parent-deprecation.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--pending-deprecation' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ undefined
- null

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ undefined
- null

─────╯
```

## test-module-prototype-mutation.js (exit 1)
```
gode: AssertionError: set %Object.prototype%.exports at /tmp/node24tests/test/parallel/test-module-prototype-mutation.js:36
called with arguments: 'no main field'
      ╭─[ node:assert:1518:1 ]
      │
 1518 │ 
      │ ┬
      │ ╰── set %Object.prototype%.exports at /tmp/node24tests/test/parallel/test-module-prototype-mutation.js:36
called with arguments: 'no main field'
──────╯
```

## test-module-relative-lookup.js (exit 1)
```
gode: TypeError: _module._resolveLookupPaths is not a function. (In '_module._resolveLookupPaths(moduleName)', '_module._resolveLookupPaths' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-module-relative-lookup.js:13:3 ]
    │
 13 │   let paths = _module._resolveLookupPaths(moduleName);
    │   ┬
    │   ╰── _module._resolveLookupPaths is not a function. (In '_module._resolveLookupPaths(moduleName)', '_module._resolveLookupPaths' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-module-run-main-monkey-patch.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(child.stdout.toString().includes('runMain is monkey patched!'))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(child.stdout.toString().includes('runMain is monkey patched!'))

──────╯
```

## test-module-stat.js (exit 1)
```
gode: TypeError: Module._stat is not a function. (In 'Module._stat(directory)', 'Module._stat' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-module-stat.js:19:1 ]
    │
 19 │ assert.strictEqual(Module._stat(directory), 1); // Returns 1 for directories.
    │ ┬
    │ ╰── Module._stat is not a function. (In 'Module._stat(directory)', 'Module._stat' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-module-unreadable-package-json.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'' !== 'real'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'' !== 'real'

─────╯
```

## test-module-version.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert(Object.hasOwn(process.config.variables, 'node_module_version'))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(Object.hasOwn(process.config.variables, 'node_module_version'))

──────╯
```

## test-module-wrap.js (exit 1)
```
gode: Error: Command failed: /tmp/gode /tmp/node24tests/test/fixtures/cjs-module-wrap.js
gode: TypeError: undefined is not an object (evaluating 'm.wrapper[0]')
   ╭─[ /tmp/node24tests/test/fixtures/cjs-module-wrap.js:6:1 ]
   │
 6 │ m.wrapper[0] += 'global.mwc = (global.mwc || 0 ) + 1;';
   │ ┬
   │ ╰── undefined is not an object (evaluating 'm.wrapper[0]')
   │
   │ Help: the value does not have the type this operation requires
───╯
    ╭─[ child_process.js:78:1 ]
    │
 78 │ 
    │ ┬
    │ ╰── Command failed: /tmp/gode /tmp/node24tests/test/fixtures/cjs-module-wrap.js
gode: TypeError: undefined is not an object (evaluating 'm.wrapper[0]')
   ╭─[ /tmp/node24tests/test/fixtures/cjs-module-wrap.js:6:1 ]
   │
 6 │ m.wrapper[0] += 'global.mwc = (global.mwc || 0 ) + 1;';
   │ ┬
   │ ╰── undefined is not an object (evaluating 'm.wrapper[0]')
   │
```

## test-module-wrapper.js (exit 1)
```
gode: Error: Command failed: /tmp/gode /tmp/node24tests/test/fixtures/cjs-module-wrapper.js
gode: AssertionError: Expected values to be strictly equal:

0 !== 1

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

0 !== 1

─────╯
    ╭─[ child_process.js:78:1 ]
    │
 78 │ 
    │ ┬
    │ ╰── Command failed: /tmp/gode /tmp/node24tests/test/fixtures/cjs-module-wrapper.js
gode: AssertionError: Expected values to be strictly equal:

0 !== 1
```
