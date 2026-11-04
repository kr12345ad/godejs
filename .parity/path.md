# path: 1 failing official tests

## test-path.js (exit 1)
```
gode: AssertionError: Expected "actual" to be reference-equal to "expected":
+ actual - expected

  {
    basename: [Function: basename],
    delimiter: ':',
    dirname: [Function: dirname],
    extname: [Function: extname],
...
    parse: [Function: parse],
+   posix: {
+   basename: [Function: basename],
+   delimiter: ':',
+   dirname: [Function: dirname],
+   extname: [Function: extname],
+   format: [Function: format],
+   isAbsolute: [Function: isAbsolute],
+   join: [Function: join],
+   matchesGlob: [Function: matchesGlob],
+   normalize: [Function: normalize],
+   parse: [Function: parse],
    relative: [Function: relative],
```
