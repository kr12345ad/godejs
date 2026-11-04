# string-decoder: 2 failing official tests

## test-string-decoder-fuzz.js (exit 1)
```
gode: AssertionError: Mismatch:
[
  '�\x0f�î��\x04��q�y',
  'x\x1a��:��Zm]u\x04B�\x1aN.�\x17k\x15\x1b',
  '��\x05\x1a���\f���\x0f��q�v�g�y��g\x17V��Kl�o�o�۞�\x01�',
  '�'
]
[
  'cd0fddc3aee9e30488b471ad79',
  '781af6a1ec3afcaa5a6d5d750442f11a4e2ef3176b151b',
  'fbf1b4051a83d7cd0cb9e9a8ec0fa1e771b076e767ea79d5ff671756fb884b6ca96ffb6ff1db9ef901eed8'
]
for encoding utf8
+ actual - expected

+ '�\x0f�î��\x04��q�yx\x1a��:��Zm]u\x04B�\x1aN.�\x17k\x15\x1b��\x05\x1a���\f���\x0f��q�v�g�y��g\x17V��Kl�o�o�۞�\x01��'
- '�\x0f�î��\x04��q�yx\x1a���:��Zm]u\x04B�\x1aN.�\x17k\x15\x1b���\x05\x1a���\f����\x0f��q�v�g�y��g\x17V��Kl�o�o�۞�\x01��'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
```

## test-string-decoder.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

'' !== '�'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

'' !== '�'

─────╯
```
