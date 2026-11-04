# tty: 1 failing official tests

## test-tty-backwards-api.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: host.binding is not a function. (In 'host.binding(id)', 'host.binding' is undefined)
    ╭─[ [gode:internal/test/binding]:18:1 ]
    │
 18 │ [
    │ ┬
    │ ╰── host.binding is not a function. (In 'host.binding(id)', 'host.binding' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```
