# crypto: 102 failing official tests

## test-crypto-argon2-job.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals', '--no-warnings' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: undefined is not a constructor (evaluating 'new Argon2Job(
      kCryptoJobSync, pass, salt, lanes, keylen, memcost, iter,
      empty, empty, kTypeArgon2id)')
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-argon2-job.js:53:5 ]
    │
 53 │     const job = new Argon2Job(
    │     ┬
    │     ╰── undefined is not a constructor (evaluating 'new Argon2Job(
      kCryptoJobSync, pass, salt, lanes, keylen, memcost, iter,
      empty, empty, kTypeArgon2id)')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-argon2.js (exit 1)
```
gode: Error: Argon2 secret and associated data are not supported by this build
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-argon2.js:15:3 ]
    │
 15 │   const syncResult = crypto.argon2Sync(algorithm, options);
    │   ┬
    │   ╰── Argon2 secret and associated data are not supported by this build
────╯
```

## test-crypto-async-sign-verify.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-async-sign-verify.js:23:3 ]
    │
 23 │   let privateKey = crypto.createPrivateKey(privatePem);
    │   ┬
    │   ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-authenticated.js (exit 1)
```
1..0 # Skipped: unsupported aes-128-siv test
1..0 # Skipped: unsupported aes-128-siv test
1..0 # Skipped: unsupported aes-128-gcm-siv test
1..0 # Skipped: unsupported aes-256-gcm-siv test
1..0 # Skipped: unsupported sm4-gcm test
1..0 # Skipped: unsupported sm4-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-192-ccm test
1..0 # Skipped: unsupported aes-192-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ccm test
1..0 # Skipped: unsupported aes-128-ocb test
1..0 # Skipped: unsupported aes-128-ocb test
```

## test-crypto-certificate.js (exit 1)
```
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

## test-crypto-cipherbase-options-fast-path.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'authTagLength' ]
- [ 'authTagLength', 'ctsMode', 'xtsStandard' ]

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ [ 'authTagLength' ]
- [ 'authTagLength', 'ctsMode', 'xtsStandard' ]

─────╯
```

## test-crypto-cipheriv-cbc-cts.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-cipheriv-decipheriv.js (exit 1)
```
gode: TypeError: Illegal constructor
     ╭─[ /tmp/node24tests/test/parallel/test-crypto-cipheriv-decipheriv.js:173:3 ]
     │
 173 │   const instance = Cipheriv(algorithm, key, iv);
     │   ┬
     │   ╰── Illegal constructor
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-crypto-dh-constructor.js (exit 1)
```
gode: TypeError: Illegal constructor
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-dh-constructor.js:15:3 ]
    │
 15 │   const dh = DiffieHellman(prime, 'buffer');
    │   ┬
    │   ╰── Illegal constructor
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-dh-curves.js (exit 1)
```
gode: Error: bad generator
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-dh-curves.js:68:3 ]
    │
 68 │   crypto.createDiffieHellman('02', 'hex');
    │   ┬
    │   ╰── bad generator
────╯
```

## test-crypto-dh-errors.js (exit 1)
```
gode: AssertionError: Missing expected exception (RangeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (RangeError).
─────╯
```

## test-crypto-dh-generate-keys.js (exit 1)
```
gode: Error: Failed to get DH public key
    ╭─[ [gode:crypto/dh]:12:1 ]
    │
 12 │ 
    │ ┬
    │ ╰── Failed to get DH public key
────╯
```

## test-crypto-dh-group-setters.js (exit 1)
```
gode: Error: Unknown group
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-dh-group-setters.js:16:1 ]
    │
 16 │ const dhg = crypto.getDiffieHellman('modp1');
    │ ┬
    │ ╰── Unknown group
────╯
```

## test-crypto-dh-modp2-views.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ '6682ae8b3fedb8a8ce226080a3ddb16381bfff4075615d2f7277af3cfc763df0a38f35ed63358faa8474a0a637ed83b8776d9a4edee8978c74cddeb319ce0442cf9b212f1981b2e1231f99f249d9fed0bec7f8a58997c113f7e9529133958a9a0bd6ad40e6c7fd8c575f554012f984ac220e7a0a5e8b3c1f4ff57f2e6c09e855'
- 'fcc79fc1fd0631c2c1e5840ad399b047cf610796c93980a09e161039dc0e2ae5b00f49fbb042e70e6d8448af0aee210fe7054e3c1758fef9c65b8d7a626847235fcd4d30c38d3afdd6f2819ca152f7ddadbae6f496481080a6b82d32a1dc109ec07026b9bb403707fe3592c8177f64874cbbfc67517de84246b636a90f8e70f4'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ '6682ae8b3fedb8a8ce226080a3ddb16381bfff4075615d2f7277af3cfc763df0a38f35ed63358faa8474a0a637ed83b8776d9a4edee8978c74cddeb319ce0442cf9b212f1981b2e1231f99f249d9fed0bec7f8a58997c113f7e9529133958a9a0bd6ad40e6c7fd8c575f554012f984ac220e7a0a5e8b3c1f4ff57f2e6c09e855'
- 'fcc79fc1fd0631c2c1e5840ad399b047cf610796c93980a09e161039dc0e2ae5b00f49fbb042e70e6d8448af0aee210fe7054e3c1758fef9c65b8d7a626847235fcd4d30c38d3afdd6f2819ca152f7ddadbae6f496481080a6b82d32a1dc109ec07026b9bb403707fe3592c8177f64874cbbfc67517de84246b636a90f8e70f4'

─────╯
```

## test-crypto-dh-stateless.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'options.privateKey\' is invalid',
-   code: 'ERR_INVALID_ARG_TYPE',
-   message: 'The "options" argument must be of type object. Received an instance of Array',
    name: 'TypeError'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'options.privateKey\' is invalid',
-   code: 'ERR_INVALID_ARG_TYPE',
```

## test-crypto-dh.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ 'Vsczas7i+iQN525cd7uAy/dp5nq98hZYvzyFw/JXUivqo/MkUN6LCAgYv8tGcP5jBYal/c4/M73cbWs1AwuxjqaSWImgWB4rWEgk9/IUK1vYK2EgxcBsVAQUowEqczAX/q9usm8Ag3gKXdc852FrwwhMomEjNmtpSHegl0TwXEM='
- 'DnHhKp5JXaSM1Fvm+c24pNnBQAFBHzmoy5IaeWdeqDFFWW8ZJZVdy/REcGSkcbB8y/TWsGvpP0NCRMxygUILTFSImdeQYG8gDv4/m1I2G+50dae841lYL4XwkTsxvvY4bRBFCbD71jlQX74daS/JRTqxp9SOG6OkpgtnUT3tDGM='

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ 'Vsczas7i+iQN525cd7uAy/dp5nq98hZYvzyFw/JXUivqo/MkUN6LCAgYv8tGcP5jBYal/c4/M73cbWs1AwuxjqaSWImgWB4rWEgk9/IUK1vYK2EgxcBsVAQUowEqczAX/q9usm8Ag3gKXdc852FrwwhMomEjNmtpSHegl0TwXEM='
- 'DnHhKp5JXaSM1Fvm+c24pNnBQAFBHzmoy5IaeWdeqDFFWW8ZJZVdy/REcGSkcbB8y/TWsGvpP0NCRMxygUILTFSImdeQYG8gDv4/m1I2G+50dae841lYL4XwkTsxvvY4bRBFCbD71jlQX74daS/JRTqxp9SOG6OkpgtnUT3tDGM='

─────╯
```

## test-crypto-domain.js (exit 1)
```
gode: Error: BAM
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-domain.js:32:3 ]
    │
 32 │   const ex = new Error('BAM');
    │   ┬
    │   ╰── BAM
────╯
```

## test-crypto-domains.js (exit 1)
```
gode: Error: pbkdf2
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-domains.js:44:7 ]
    │
 44 │       throw new Error('pbkdf2');
    │       ┬
    │       ╰── pbkdf2
────╯
```

## test-crypto-ecdh-convert-key.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_CRYPTO_UNKNOWN_CURVE', name: 'Error' }
- Comparison { code: 'ERR_INVALID_ARG_TYPE', name: 'TypeError' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_CRYPTO_UNKNOWN_CURVE', name: 'Error' }
- Comparison { code: 'ERR_INVALID_ARG_TYPE', name: 'TypeError' }

──────╯
```

## test-crypto-eddsa-variants.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ <Buffer 14 35 04 d6 85 91 01 35 b4 cf cd c9 1d 48 d1 bd 5d 29 2c f5 4a b3 d3 69 a5 a8 12 2a a9 13 3e 97 01 ea b4 01 63 05 5c e5 a6 4d 74 a2 10 cf 7d 42 26 8d ... 14 more bytes>
- <Buffer 55 a4 cc 2f 70 a5 4e 04 28 8c 5f 4c d1 e4 5a 7b b5 20 b3 62 92 91 18 76 ca da 73 23 19 8d d8 7a 8b 36 95 0b 95 13 00 22 90 7a 7f b7 c4 e9 b2 d5 f6 cc ... 14 more bytes>

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ <Buffer 14 35 04 d6 85 91 01 35 b4 cf cd c9 1d 48 d1 bd 5d 29 2c f5 4a b3 d3 69 a5 a8 12 2a a9 13 3e 97 01 ea b4 01 63 05 5c e5 a6 4d 74 a2 10 cf 7d 42 26 8d ... 14 more bytes>
- <Buffer 55 a4 cc 2f 70 a5 4e 04 28 8c 5f 4c d1 e4 5a 7b b5 20 b3 62 92 91 18 76 ca da 73 23 19 8d d8 7a 8b 36 95 0b 95 13 00 22 90 7a 7f b7 c4 e9 b2 d5 f6 cc ... 14 more bytes>

─────╯
```

## test-crypto-encap-decap.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "key" argument must be an instance of KeyObject. Received undefined'
-   message: /The "key" argument must be of type/
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
    code: 'ERR_INVALID_ARG_TYPE',
+   message: 'The "key" argument must be an instance of KeyObject. Received undefined'
-   message: /The "key" argument must be of type/
  }

```

## test-crypto-encoding-validation-error.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-crypto-fips.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: testFipsCrypto is not a function. (In 'testFipsCrypto()', 'testFipsCrypto' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-fips.js:83:3 ]
    │
 83 │   testFipsCrypto() ? 'stdout' : 'stderr',
    │   ┬
    │   ╰── testFipsCrypto is not a function. (In 'testFipsCrypto()', 'testFipsCrypto' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-gcm-explicit-short-tag.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--pending-deprecation' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: Invalid initialization vector
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-gcm-explicit-short-tag.js:29:3 ]
    │
 29 │   createDecipheriv('aes-256-gcm', key, iv, {
    │   ┬
    │   ╰── Invalid initialization vector
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-getcipherinfo.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ {
+   blockSize: 1,
+   ivLength: 0,
+   keyLength: 16,
+   mode: 'ecb',
+   name: 'aes-128-ecb',
+   nid: 0
+ }
- undefined

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ {
+   blockSize: 1,
```

## test-crypto-hash-stream-error.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: h.on is not a function. (In 'h.on('error', common.mustCall((err) => {
  assert.strictEqual(err.code, 'ERR_CRYPTO_HASH_UPDATE_FAILED');
}))', 'h.on' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-hash-stream-error.js:27:1 ]
    │
 27 │ h.on('error', common.mustCall((err) => {
    │ ┬
    │ ╰── h.on is not a function. (In 'h.on('error', common.mustCall((err) => {
  assert.strictEqual(err.code, 'ERR_CRYPTO_HASH_UPDATE_FAILED');
}))', 'h.on' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-hash-stream-pipe.js (exit 1)
```
gode: TypeError: dest.emit is not a function. (In 'dest.emit('pipe', this)', 'dest.emit' is undefined)
     ╭─[ stream.js:887:1 ]
     │
 887 │ 
     │ ┬
     │ ╰── dest.emit is not a function. (In 'dest.emit('pipe', this)', 'dest.emit' is undefined)
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-crypto-hkdf.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_CRYPTO_INVALID_KEYLEN' }
- Comparison { code: 'ERR_OUT_OF_RANGE' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_CRYPTO_INVALID_KEYLEN' }
- Comparison { code: 'ERR_OUT_OF_RANGE' }

──────╯
```

## test-crypto-hmac.js (exit 1)
```
gode: TypeError: str.end is not a function. (In 'str.end(rfc4231[i].data)', 'str.end' is undefined)
     ╭─[ /tmp/node24tests/test/parallel/test-crypto-hmac.js:291:5 ]
     │
 291 │     str.end(rfc4231[i].data);
     │     ┬
     │     ╰── str.end is not a function. (In 'str.end(rfc4231[i].data)', 'str.end' is undefined)
     │
     │ Help: the value does not have the type this operation requires
─────╯
```

## test-crypto-jwk-raw-validation.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-jwk-raw-validation.js:20:3 ]
    │
 20 │   const jwk = createPrivateKey(fixtures.readKey(file)).export({ format: 'jwk' });
    │   ┬
    │   ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-key-encoding-provider-cipher.js (exit 1)
```
gode: Error: Unknown cipher
    ╭─[ [gode:crypto/keyobject]:19:3 ]
    │
 19 │   verify,
    │   ┬
    │   ╰── Unknown cipher
────╯
```

## test-crypto-key-objects-ec-jwk-private.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-key-objects-raw.js (exit 1)
```
gode: TypeError: The argument 'options.format' must be one of: 'pem', 'der', 'jwk'. Received 'raw-public'
    ╭─[ [gode:crypto/keyobject]:19:3 ]
    │
 19 │   const rawPub = pubKeyObj.export({ format: 'raw-public' });
    │   ┬
    │   ╰── The argument 'options.format' must be one of: 'pem', 'der', 'jwk'. Received 'raw-public'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-key-objects-to-crypto-key.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: undefined is not an object (evaluating 'kSupportedAlgorithms.digest')
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-key-objects-to-crypto-key.js:20:1 ]
    │
 20 │ const hashes = Object.keys(kSupportedAlgorithms.digest).filter((name) => {
    │ ┬
    │ ╰── undefined is not an object (evaluating 'kSupportedAlgorithms.digest')
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-key-objects.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:7 ]
     │
 916 │       } else {
     │       ┬
     │       ╰── Missing expected exception (TypeError).
─────╯
```

## test-crypto-key-store.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-key-store.js:43:3 ]
    │
 43 │   const pk = createPrivateKey(url);
    │   ┬
    │   ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-async-dsa.js (exit 1)
```
gode: AssertionError: Key (617 bytes) is shorter than expected (648 bytes)
      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── Key (617 bytes) is shorter than expected (648 bytes)
──────╯
```

## test-crypto-keygen-async-elliptic-curve-jwk-ec.js (exit 1)
```
gode: TypeError: Invalid EC curve name
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-async-elliptic-curve-jwk-ec.js:20:3 ]
    │
 20 │   generateKeyPair('ec', {
    │   ┬
    │   ╰── Invalid EC curve name
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-async-elliptic-curve-jwk.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'ed448'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-async-elliptic-curve-jwk.js:21:5 ]
    │
 21 │     generateKeyPair(type, {
    │     ┬
    │     ╰── The argument 'type' must be a supported key type. Received 'ed448'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-async-encrypted-private-key.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-crypto-keygen-async-explicit-elliptic-curve-encrypted-p256.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /^\-\-\-\-\-BEGIN ENCRYPTED PRIVATE KEY\-\-\-\-\-\n([a-zA-Z0-9\+/=]{64}\n)*[a-zA-Z0-9\+/=]{1,64}\n\-\-\-\-\-END ENCRYPTED PRIVATE KEY\-\-\-\-\-\n$/. Input:

'-----BEGIN PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC,b449488518778a646fd38b98108cd965\n\nE+QuHpzEYjVRDpE/bM/80JANu9kQ3z1E7fuRCVMVnnnQud2a5B4tO4nphbZiY1YD\naGSyabd2psV46RtGNx+Lw/PXh/afTc4Tn2Vp4esX7pVkyiG3eyYcNP+cgRaBqOmw\nisa7phu0WSWe6hfxMhHiBtuTJhVqnOtve8VUit+HlinYBDy91hlJ6qdYq7qmfQkO\n-----END PRIVATE KEY-----\n'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /^\-\-\-\-\-BEGIN ENCRYPTED PRIVATE KEY\-\-\-\-\-\n([a-zA-Z0-9\+/=]{64}\n)*[a-zA-Z0-9\+/=]{1,64}\n\-\-\-\-\-END ENCRYPTED PRIVATE KEY\-\-\-\-\-\n$/. Input:

'-----BEGIN PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC,b449488518778a646fd38b98108cd965\n\nE+QuHpzEYjVRDpE/bM/80JANu9kQ3z1E7fuRCVMVnnnQud2a5B4tO4nphbZiY1YD\naGSyabd2psV46RtGNx+Lw/PXh/afTc4Tn2Vp4esX7pVkyiG3eyYcNP+cgRaBqOmw\nisa7phu0WSWe6hfxMhHiBtuTJhVqnOtve8VUit+HlinYBDy91hlJ6qdYq7qmfQkO\n-----END PRIVATE KEY-----\n'

──────╯
```

## test-crypto-keygen-async-explicit-elliptic-curve-encrypted.js.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

──────╯
```

## test-crypto-keygen-async-named-elliptic-curve-encrypted-p256.js (exit 1)
```
gode: AssertionError: The input did not match the regular expression /^\-\-\-\-\-BEGIN ENCRYPTED PRIVATE KEY\-\-\-\-\-\n([a-zA-Z0-9\+/=]{64}\n)*[a-zA-Z0-9\+/=]{1,64}\n\-\-\-\-\-END ENCRYPTED PRIVATE KEY\-\-\-\-\-\n$/. Input:

'-----BEGIN PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC,ea99eec8f267eb5b8d8d9354513baeaf\n\nOjnqk4zb3D5yiQ29Giximbn2MCbCN4FebtAzWe+TRlxKijOFAyZCBjVqLRmjBsWO\nG71UXRnnYtPy0ZWC6se/4oPIr+VF/2q6Vv71BdVad7uZ/5/AbVfyiPyrhJi6qwqC\nWfEP4See9fAXq7l72G4RrRQ3DTBCWf78BTj07un77YVBX7uQdURWvG+XiQd7n1JZ\n-----END PRIVATE KEY-----\n'

      ╭─[ node:assert:1474:1 ]
      │
 1474 │ 
      │ ┬
      │ ╰── The input did not match the regular expression /^\-\-\-\-\-BEGIN ENCRYPTED PRIVATE KEY\-\-\-\-\-\n([a-zA-Z0-9\+/=]{64}\n)*[a-zA-Z0-9\+/=]{1,64}\n\-\-\-\-\-END ENCRYPTED PRIVATE KEY\-\-\-\-\-\n$/. Input:

'-----BEGIN PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC,ea99eec8f267eb5b8d8d9354513baeaf\n\nOjnqk4zb3D5yiQ29Giximbn2MCbCN4FebtAzWe+TRlxKijOFAyZCBjVqLRmjBsWO\nG71UXRnnYtPy0ZWC6se/4oPIr+VF/2q6Vv71BdVad7uZ/5/AbVfyiPyrhJi6qwqC\nWfEP4See9fAXq7l72G4RrRQ3DTBCWf78BTj07un77YVBX7uQdURWvG+XiQd7n1JZ\n-----END PRIVATE KEY-----\n'

──────╯
```

## test-crypto-keygen-async-named-elliptic-curve-encrypted.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

──────╯
```

## test-crypto-keygen-async-rsa.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect', name: 'Error' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled',
-   name: 'Error'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect', name: 'Error' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled',
-   name: 'Error'
- }

```

## test-crypto-keygen-bit-length.js (exit 1)
```
gode: TypeError: Invalid DSA modulus and divisor length
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-bit-length.js:53:5 ]
    │
 53 │     generateKeyPair('dsa', {
    │     ┬
    │     ╰── Invalid DSA modulus and divisor length
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-dh-classic.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'dh'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-dh-classic.js:19:3 ]
    │
 19 │   generateKeyPair('dh', {
    │   ┬
    │   ╰── The argument 'type' must be a supported key type. Received 'dh'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-eddsa.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'ed448'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-eddsa.js:21:5 ]
    │
 21 │     generateKeyPair(keyType, common.mustCall((err, publicKey, privateKey) => {
    │     ┬
    │     ╰── The argument 'type' must be a supported key type. Received 'ed448'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-empty-passphrase-no-prompt.js (exit 1)
```
gode: AssertionError: Missing expected exception (Error).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (Error).
─────╯
```

## test-crypto-keygen-invalid-parameter-encoding-dsa.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_MISSING_OPTION',
+   message: 'options.divisorLength is required',
+   name: 'TypeError'
-   code: 'ERR_CRYPTO_JWK_UNSUPPORTED_KEY_TYPE',
-   message: 'Unsupported JWK Key Type.',
-   name: 'Error'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_MISSING_OPTION',
+   message: 'options.divisorLength is required',
```

## test-crypto-keygen-invalid-parameter-encoding-ec.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_CRYPTO_INVALID_CURVE',
+   message: 'Invalid EC curve name',
+   name: 'TypeError'
-   code: 'ERR_CRYPTO_JWK_UNSUPPORTED_CURVE',
-   message: 'Unsupported JWK EC curve: secp224r1.',
-   name: 'Error'
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_CRYPTO_INVALID_CURVE',
+   message: 'Invalid EC curve name',
```

## test-crypto-keygen-no-rsassa-pss-params.js (exit 1)
```
gode: AssertionError: 305c300d06092a864886f70d0101010500034b003048024100a262b756fda6f4ab531f4cc55eb0d98f1df6eda3b3d5aa4e45a2ea08c13e6c60d68b7bc2dcabec4c32b30b27de05f0e5db9f5f1c1755ae61ecdcccacc9a26e090203010001

13 !== 11

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── 305c300d06092a864886f70d0101010500034b003048024100a262b756fda6f4ab531f4cc55eb0d98f1df6eda3b3d5aa4e45a2ea08c13e6c60d68b7bc2dcabec4c32b30b27de05f0e5db9f5f1c1755ae61ecdcccacc9a26e090203010001

13 !== 11

─────╯
```

## test-crypto-keygen-non-standard-public-exponent.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ { modulusLength: 512, publicExponent: 65537n }
- { modulusLength: 512, publicExponent: 3n }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ { modulusLength: 512, publicExponent: 65537n }
- { modulusLength: 512, publicExponent: 3n }

─────╯
```

## test-crypto-keygen-raw.js (exit 1)
```
gode: TypeError: The argument 'options.format' must be one of: 'pem', 'der', 'jwk'. Received 'raw-public'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keygen-raw.js:36:5 ]
    │
 36 │     const { publicKey, privateKey } = generateKeyPairSync(type, options);
    │     ┬
    │     ╰── The argument 'options.format' must be one of: 'pem', 'der', 'jwk'. Received 'raw-public'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-keygen-rfc8017-a-2-3.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  {
    hashAlgorithm: 'sha512',
    mgf1HashAlgorithm: 'sha512',
    modulusLength: 512,
    publicExponent: 65537n,
+   saltLength: 64
-   saltLength: 0
  }

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  {
    hashAlgorithm: 'sha512',
    mgf1HashAlgorithm: 'sha512',
```

## test-crypto-keygen-rsa-pss.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-keygen.js (exit 1)
```
gode: AssertionError: Missing expected exception (TypeError).
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception (TypeError).
─────╯
```

## test-crypto-keyobject-hidden-slots.js (exit 1)
```
gode: TypeError: The "ikm" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received an instance of KeyObject
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-keyobject-hidden-slots.js:71:7 ]
    │
 71 │       hkdfSync('sha256', secret, Buffer.alloc(0), Buffer.alloc(0), 16)
    │       ┬
    │       ╰── The "ikm" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received an instance of KeyObject
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-lazy-transform-writable.js (exit 1)
```
gode: TypeError: The superclass is not a constructor.
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-lazy-transform-writable.js:20:25 ]
    │
 20 │ class OldStream extends Stream {
    │                         ┬
    │                         ╰── The superclass is not a constructor.
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-mac-errors.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_VALUE' }
- Comparison { code: 'ERR_INVALID_ARG_TYPE' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_INVALID_ARG_VALUE' }
- Comparison { code: 'ERR_INVALID_ARG_TYPE' }

──────╯
```

## test-crypto-mac.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly equal:
+ actual - expected

+ [Function: Mac]
- undefined

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:
+ actual - expected

+ [Function: Mac]
- undefined

─────╯
```

## test-crypto-negative-zero.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { name: 'RangeError' }
- Comparison { name: 'Error' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { name: 'RangeError' }
- Comparison { name: 'Error' }

──────╯
```

## test-crypto-no-algorithm.js (exit 1)
```
gode: TypeError: The "fips" argument must be of type boolean. Received type number (1)
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-no-algorithm.js:19:3 ]
    │
 19 │   crypto.setFips(1);
    │   ┬
    │   ╰── The "fips" argument must be of type boolean. Received type number (1)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-padding.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_OSSL_EVP_DATA_NOT_MULTIPLE_OF_BLOCK_LENGTH',
+   message: 'data not multiple of block length'
-   code: /ERR_OSSL(_EVP)?_WRONG_FINAL_BLOCK_LENGTH/,
-   message: /wrong[\s_]final[\s_]block[\s_]length/i,
-   reason: /wrong[\s_]final[\s_]block[\s_]length/i
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_OSSL_EVP_DATA_NOT_MULTIPLE_OF_BLOCK_LENGTH',
+   message: 'data not multiple of block length'
-   code: /ERR_OSSL(_EVP)?_WRONG_FINAL_BLOCK_LENGTH/,
```

## test-crypto-pbkdf2.js (exit 124)
```

```

## test-crypto-pkcs12.js (exit 1)
```
gode: TypeError: The "passphrase" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received an instance of Object
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pkcs12.js:49:5 ]
    │
 49 │     crypto.parsePKCS12(bundle, { passphrase: keyAndCert.passphrase });
    │     ┬
    │     ╰── The "passphrase" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received an instance of Object
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-pqc-encrypted-pkcs8.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'ml-dsa-44'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-encrypted-pkcs8.js:64:3 ]
    │
 64 │   const { privateKey } = generateKeyPairSync(asymmetricKeyType);
    │   ┬
    │   ╰── The argument 'type' must be a supported key type. Received 'ml-dsa-44'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-pqc-key-objects-ml-dsa.js (exit 1)
```
gode: Error: Failed to parse key
     ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-key-objects-ml-dsa.js:114:5 ]
     │
 114 │     const publicKey = createPublicKey(keys.public);
     │     ┬
     │     ╰── Failed to parse key
─────╯
```

## test-crypto-pqc-key-objects-ml-kem.js (exit 1)
```
gode: Error: Failed to parse key
     ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-key-objects-ml-kem.js:122:5 ]
     │
 122 │     const publicKey = createPublicKey(keys.public);
     │     ┬
     │     ╰── Failed to parse key
─────╯
```

## test-crypto-pqc-key-objects-slh-dsa.js (exit 1)
```
gode: Error: Failed to parse key
     ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-key-objects-slh-dsa.js:111:5 ]
     │
 111 │     const publicKey = createPublicKey(keys.public);
     │     ┬
     │     ╰── Failed to parse key
─────╯
```

## test-crypto-pqc-keygen-ml-dsa.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'ml-dsa-44'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-keygen-ml-dsa.js:56:7 ]
    │
 56 │       generateKeyPair(asymmetricKeyType, { publicKeyEncoding }, common.mustSucceed(validate));
    │       ┬
    │       ╰── The argument 'type' must be a supported key type. Received 'ml-dsa-44'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-pqc-keygen-ml-kem.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'ml-kem-512'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-keygen-ml-kem.js:56:7 ]
    │
 56 │       generateKeyPair(asymmetricKeyType, { publicKeyEncoding }, common.mustSucceed(validate));
    │       ┬
    │       ╰── The argument 'type' must be a supported key type. Received 'ml-kem-512'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-pqc-keygen-slh-dsa.js (exit 1)
```
gode: TypeError: The argument 'type' must be a supported key type. Received 'slh-dsa-sha2-128f'
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-keygen-slh-dsa.js:62:7 ]
    │
 62 │       generateKeyPair(asymmetricKeyType, { publicKeyEncoding }, common.mustSucceed(validate));
    │       ┬
    │       ╰── The argument 'type' must be a supported key type. Received 'slh-dsa-sha2-128f'
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-pqc-sign-verify-ml-dsa.js (exit 1)
```
gode: Error: Failed to parse key
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-pqc-sign-verify-ml-dsa.js:49:9 ]
    │
 49 │         const signature = sign(undefined, data, privateKey);
    │         ┬
    │         ╰── Failed to parse key
────╯
```

## test-crypto-prepare-asymmetric-cryptokey.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-prepare-asymmetric-cryptokey.js:25:1 ]
    │
 25 │ const privateKeyObject = crypto.createPrivateKey(fixtures.readKey('ec_p256_private.pem'));
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-prepare-asymmetric-key.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-prepare-asymmetric-key.js:20:1 ]
    │
 20 │ const keyObject = crypto.createPrivateKey(pem);
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-prime.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-private-decrypt-gh32240.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'x509: decryption password incorrect' }
- Comparison {
-   message: 'error:07880109:common libcrypto routines::interrupted or cancelled'
- }

──────╯
```

## test-crypto-provider-cache-snapshot.js (exit 1)
```
[stderr]
gode: Cannot find module '/tmp/node24tests/test/.tmp.test-crypto-provider-cache-snapshot/snapshot.blob'
   ╭─[ /tmp/node24tests/test/.tmp.test-crypto-provider-cache-snapshot/snapshot.blob:1:1 ]
   │
 1 │ 
   │ ┬
   │ ╰── Cannot find module '/tmp/node24tests/test/.tmp.test-crypto-provider-cache-snapshot/snapshot.blob'
───╯

[stdout]

gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0
```

## test-crypto-provider-hashes.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: The expression evaluated to a falsy value:

  assert(hashes.includes('RSA-SHA1'))

      ╭─[ node:assert:1042:1 ]
      │
 1042 │ 
      │ ┬
      │ ╰── The expression evaluated to a falsy value:

  assert(hashes.includes('RSA-SHA1'))

──────╯
```

## test-crypto-publicDecrypt-fails-first-time.js (exit 1)
```
gode: Error: Unknown cipher
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-publicDecrypt-fails-first-time.js:21:1 ]
    │
 21 │ const { privateKey, publicKey } = crypto.generateKeyPairSync('rsa', {
    │ ┬
    │ ╰── Unknown cipher
────╯
```

## test-crypto-random-cache-snapshot.js (exit 1)
```
[stderr]
gode: Cannot find module '/tmp/node24tests/test/.tmp.test-crypto-random-cache-snapshot/snapshot.blob'
   ╭─[ /tmp/node24tests/test/.tmp.test-crypto-random-cache-snapshot/snapshot.blob:1:1 ]
   │
 1 │ 
   │ ┬
   │ ╰── Cannot find module '/tmp/node24tests/test/.tmp.test-crypto-random-cache-snapshot/snapshot.blob'
───╯

[stdout]

gode: AssertionError: Expected values to be strictly equal:

1 !== 0

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

1 !== 0
```

## test-crypto-random.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--pending-deprecation' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: AssertionError: Expected values to be strictly deep-equal:

'0000' !== '00ab'

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly deep-equal:

'0000' !== '00ab'

─────╯
```

## test-crypto-raw-key-type-validation.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_OSSL_PEM_NO_START_LINE' }
- Comparison { code: 'ERR_CRYPTO_INCOMPATIBLE_KEY_OPTIONS' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { code: 'ERR_OSSL_PEM_NO_START_LINE' }
- Comparison { code: 'ERR_CRYPTO_INCOMPATIBLE_KEY_OPTIONS' }

──────╯
```

## test-crypto-rsa-dsa.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-rsa-dsa.js:90:5 ]
    │
 90 │     crypto.publicEncrypt(ab, ab2enc);
    │     ┬
    │     ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-rsa-multiprime-jwk.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-rsa-multiprime-jwk.js:20:1 ]
    │
 20 │ const privateKey = createPrivateKey(
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-rsa-oaep-mgf1.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-rsa-pkcs1-public-key.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-rsa-pkcs1-public-key.js:12:1 ]
    │
 12 │ const privateKey = createPrivateKey(fixtures.readKey('rsa_private_2048.pem'));
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-rsa-pss-parameters.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-rsa-pss-parameters.js:13:1 ]
    │
 13 │ const publicKey = createPublicKey(
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-scrypt.js (exit 1)
```
gode: AssertionError: Missing expected exception.
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Missing expected exception.
─────╯
```

## test-crypto-secret-keygen.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'options.length\' is required'
-   code: 'ERR_INVALID_ARG_TYPE',
-   message: /The "options" argument must be /
  }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

  Comparison {
+   code: 'ERR_INVALID_ARG_VALUE',
+   message: 'The argument \'options.length\' is required'
-   code: 'ERR_INVALID_ARG_TYPE',
-   message: /The "options" argument must be /
```

## test-crypto-secure-heap.js (exit 1)
```
gode: AssertionError: Expected values to be strictly equal:

0 !== 65536

     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected values to be strictly equal:

0 !== 65536

─────╯
gode: AssertionError: Expected "actual" to be strictly unequal to: 0
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" to be strictly unequal to: 0
─────╯
```

## test-crypto-sign-verify.js (exit 1)
```
gode: AssertionError: Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'error:0E079065:configuration file routines::no start line' }
- Comparison { message: 'bye, bye, library' }

      ╭─[ node:assert:1176:1 ]
      │
 1176 │ 
      │ ┬
      │ ╰── Expected values to be strictly deep-equal:
+ actual - expected

+ Comparison { message: 'error:0E079065:configuration file routines::no start line' }
- Comparison { message: 'bye, bye, library' }

──────╯
```

## test-crypto-stream.js (exit 1)
```
gode: TypeError: hasher1.pipe is not a function. (In 'hasher1.pipe(new Stream2buffer(common.mustCall(function end(err, hash) {
    assert.strictEqual(err, null);
    assert.strictEqual(
      hash.toString('hex'), '06460dadb35d3d503047ce750ceb2d07'
    );
  })))', 'hasher1.pipe' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-stream.js:53:3 ]
    │
 53 │   hasher1.pipe(new Stream2buffer(common.mustCall(function end(err, hash) {
    │   ┬
    │   ╰── hasher1.pipe is not a function. (In 'hasher1.pipe(new Stream2buffer(common.mustCall(function end(err, hash) {
    assert.strictEqual(err, null);
    assert.strictEqual(
      hash.toString('hex'), '06460dadb35d3d503047ce750ceb2d07'
    );
  })))', 'hasher1.pipe' is undefined)
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-subtle-cross-realm.js (exit 1)
```
gode: AssertionError: Expected "actual" not to be reference-equal to "expected":

ArrayBuffer { byteLength: undefined }
     ╭─[ node:assert:916:1 ]
     │
 916 │ 
     │ ┬
     │ ╰── Expected "actual" not to be reference-equal to "expected":

ArrayBuffer { byteLength: undefined }
─────╯
```

## test-crypto-verify-failure.js (exit 1)
```
gode: TypeError: tls.Server is not a function. (In 'tls.Server(options, (socket) => {
  setImmediate(() => {
    verify();
    setImmediate(() => {
      socket.destroy();
    });
  });
})', 'tls.Server' is undefined)
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-verify-failure.js:39:1 ]
    │
 39 │ const server = tls.Server(options, (socket) => {
    │ ┬
    │ ╰── tls.Server is not a function. (In 'tls.Server(options, (socket) => {
  setImmediate(() => {
    verify();
    setImmediate(() => {
      socket.destroy();
    });
  });
})', 'tls.Server' is undefined)
    │
    │ Help: the value does not have the type this operation requires
```

## test-crypto-x509-brand-check.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: Value of "this" must be of type X509Certificate
   ╭─[ [gode:crypto/x509]:5:1 ]
   │
 5 │ if (!common.hasCrypto)
   │ ┬
   │ ╰── Value of "this" must be of type X509Certificate
   │
   │ Help: the value does not have the type this operation requires
───╯
```

## test-crypto-x509-clone-transfer.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-x509-clone-transfer.js:31:1 ]
    │
 31 │ const privateKey = createPrivateKey(keyData);
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-x509-hidden-slots.js (exit 1)
```
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-x509-hidden-slots.js:22:1 ]
    │
 22 │ const privateKey = createPrivateKey(keyData);
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto-x509-name-non-string.js (exit 1)
```
gode: Error: Unable to load certificate: x509: invalid RDNSequence: invalid attribute value: unsupported string type: 3
    ╭─[ [gode:crypto/x509]:10:1 ]
    │
 10 │ const fixtures = require('../common/fixtures');
    │ ┬
    │ ╰── Unable to load certificate: x509: invalid RDNSequence: invalid attribute value: unsupported string type: 3
────╯
```

## test-crypto-x509.js (exit 1)
```
NOTE: The test started as a child_process using these flags: [ '--expose-internals' ] And these environment variables: {} Use NODE_SKIP_FLAG_CHECK to run the test with the original flags.
gode: TypeError: The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    ╭─[ /tmp/node24tests/test/parallel/test-crypto-x509.js:31:1 ]
    │
 31 │ const privateKey = createPrivateKey(key);
    │ ┬
    │ ╰── The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received undefined
    │
    │ Help: the value does not have the type this operation requires
────╯
```

## test-crypto.js (exit 1)
```
gode: AssertionError: The validation function is expected to return "true". Received false

Caught error:

Error [ERR_NOT_IMPLEMENTED]: GodeJS does not implement tls.createSecureContext yet
      ╭─[ node:assert:1263:1 ]
      │
 1263 │ 
      │ ┬
      │ ╰── The validation function is expected to return "true". Received false

Caught error:

Error [ERR_NOT_IMPLEMENTED]: GodeJS does not implement tls.createSecureContext yet
──────╯
```
