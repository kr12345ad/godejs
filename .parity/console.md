# console: 8 failing official tests

## test-console-diagnostics-channels.js (exit 1)
```
string { key: /value/ } [ 1, 2, 3 ]
string { key: /value/ } [ 1, 2, 3 ]
string { key: /value/ } [ 1, 2, 3 ]
string { key: /value/ } [ 1, 2, 3 ]
string { key: /value/ } [ 1, 2, 3 ]
gode: AssertionError: The expression evaluated to a falsy value:

  assert.ok(intercepted)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert.ok(intercepted)

──────╯
```

## test-console-formatTime.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '100.0096'
- '100.01ms'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '100.0096'
- '100.01ms'

─────╯
```

## test-console-instance.js (exit 1)
```
gode: TypeError: Object is not a constructor (evaluating 'new Stream()')
    ╭─[ /tmp/node24tests/test/parallel/test-console-instance.js:29:1 ]
    │
 29 │ const out = new Stream();
    │ ┬
    │ ╰── Object is not a constructor (evaluating 'new Stream()')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-console-no-swallow-stack-overflow.js (exit 124)
```

```

## test-console-table.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  [
+   '┌─────────┬──────────┐',
+   '│ (index) │ Values   │',
+   '├─────────┼──────────┤',
+   '│ 0       │ Symbol() │',
+   '│ 1       │ 5        │',
+   '│ 2       │ [ 10 ]   │',
+   '└─────────┴──────────┘',
-   '┌─────────┬────┬──────────┐',
-   '│ (index) │ 0  │ Values   │',
-   '├─────────┼────┼──────────┤',
-   '│ 0       │    │ Symbol() │',
-   '│ 1       │    │ 5        │',
-   '│ 2       │ 10 │          │',
-   '└─────────┴────┴──────────┘',
    ''
  ]

     ╭─[ node:assert:916:1 ]
```

## test-console-tty-colors-per-stream.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'Hello 42\n'
- 'Hello \x1b[33m42\x1b[39m\n'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'Hello 42\n'
- 'Hello \x1b[33m42\x1b[39m\n'

─────╯
```

## test-console-tty-colors.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-console.js (exit 1)
```
gode: AssertionError: The expression evaluated to a falsy value:

  assert.ok(process.stdout.writable)

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert.ok(process.stdout.writable)

──────╯
```
