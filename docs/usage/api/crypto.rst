crypto
======

The ``node:crypto`` module provides cryptographic functionality that includes a set of wrappers for OpenSSL's hash, message authentication code (MAC), cipher, decipher, sign, verify, and key encapsulation mechanism (KEM) functions.

Loading
-------

.. code-block:: javascript

   const crypto = require('node:crypto');
   // ESM
   import * as crypto from 'node:crypto';

Quick start
-----------

.. code-block:: javascript

   const { createHmac } = require('node:crypto');

   const secret = 'abcdefg';
   const hash = createHmac('sha256', secret)
                  .update('I love cupcakes')
                  .digest('hex');
   console.log(hash);
   // Prints:
   //   c0fa1bc00531bd78ef38c628449c5102aeabd49b5dc3a2a516ea6ea959d6658e

API
---

argon2
~~~~~~

.. code-block:: text

   crypto.argon2(algorithm, parameters, callback)

*Kind:* function · *Arguments:* 3

Provides an asynchronous Argon2 implementation. Argon2 is a password-based key derivation function that is designed to be expensive computationally and memory-wise in order to make brute-force attacks unrewarding.

The ``nonce`` should be as unique as possible. It is recommended that a nonce is random and at least 16 bytes long. See NIST SP 800-132 for details.

**Arguments**

* ``algorithm`` (*string*): Variant of Argon2, one of ``"argon2d"``, ``"argon2i"`` or ``"argon2id"``.
* ``parameters`` (*Object*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const { argon2, randomBytes } = require('node:crypto');

   const parameters = {
     message: 'password',
     nonce: randomBytes(16),
     parallelism: 4,
     tagLength: 64,
     memory: 65536,
     passes: 3,
   };

   argon2('argon2id', parameters, (err, derivedKey) => {
     if (err) throw err;
     console.log(derivedKey.toString('hex'));  // 'af91dad...9520f15'
   });


argon2Sync
~~~~~~~~~~

.. code-block:: text

   crypto.argon2Sync(algorithm, parameters)

*Kind:* function · *Arguments:* 2

Provides a synchronous Argon2 implementation. Argon2 is a password-based key derivation function that is designed to be expensive computationally and memory-wise in order to make brute-force attacks unrewarding.

The ``nonce`` should be as unique as possible. It is recommended that a nonce is random and at least 16 bytes long. See NIST SP 800-132 for details.

**Arguments**

* ``algorithm`` (*string*): Variant of Argon2, one of ``"argon2d"``, ``"argon2i"`` or ``"argon2id"``.
* ``parameters`` (*Object*)

**Example**

.. code-block:: javascript

   const { argon2Sync, randomBytes } = require('node:crypto');

   const parameters = {
     message: 'password',
     nonce: randomBytes(16),
     parallelism: 4,
     tagLength: 64,
     memory: 65536,
     passes: 3,
   };

   const derivedKey = argon2Sync('argon2id', parameters);
   console.log(derivedKey.toString('hex'));  // 'af91dad...9520f15'


Certificate
~~~~~~~~~~~

.. code-block:: text

   crypto.Certificate()

*Kind:* class · *Arguments:* 0

Instances of the ``Certificate`` class can be created using the ``new`` keyword or by calling ``crypto.Certificate()`` as a function:

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { Certificate } = require('node:crypto');

   const cert1 = new Certificate();
   const cert2 = Certificate();


checkPrime
~~~~~~~~~~

.. code-block:: text

   crypto.checkPrime(candidate[, options], callback)

*Kind:* function · *Arguments:* 3

Checks the primality of the ``candidate``.

**Arguments**

* ``candidate`` (*ArrayBuffer|SharedArrayBuffer|TypedArray|Buffer|DataView|bigint*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.checkPrime([], {}, () => {});
   console.log(result);


checkPrimeSync
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.checkPrimeSync(candidate[, options])

*Kind:* function · *Arguments:* 2

Checks the primality of the ``candidate``.

**Arguments**

* ``candidate`` (*ArrayBuffer|SharedArrayBuffer|TypedArray|Buffer|DataView|bigint*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.checkPrimeSync([], {});
   console.log(result);


Cipheriv
~~~~~~~~

.. code-block:: text

   Cipheriv

*Kind:* class · *Arguments:* 0

Instances of the ``Cipheriv`` class are used to encrypt data. The class can be used in one of two ways:

* As a stream that is both readable and writable, where plain unencrypted data is written to produce encrypted data on the readable side, or * Using the ``cipher.update()`` and ``cipher.final()`` methods to produce the encrypted data.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     scrypt,
     randomFill,
     createCipheriv,
   } = require('node:crypto');

   const algorithm = 'aes-192-cbc';
   const password = 'Password used to generate key';

   // First, we'll generate the key. The key length is dependent on the algorithm.
   // In this case for aes192, it is 24 bytes (192 bits).
   scrypt(password, 'salt', 24, (err, key) => {
     if (err) throw err;
     // Then, we'll generate a random initialization vector
     randomFill(new Uint8Array(16), (err, iv) => {
       if (err) throw err;

       // Once we have the key and iv, we can create and use the cipher...
       const cipher = createCipheriv(algorithm, key, iv);

       let encrypted = '';
       cipher.setEncoding('hex');

       cipher.on('data', (chunk) => encrypted += chunk);
       cipher.on('end', () => console.log(encrypted));

       cipher.write('some clear text data');
       cipher.end();
     });
   });


createCipheriv
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createCipheriv(algorithm, key, iv[, options])

*Kind:* function · *Arguments:* 4

Creates and returns a ``Cipheriv`` object, with the given ``algorithm``, ``key`` and initialization vector (``iv``).

The ``options`` argument controls cipher-specific settings and stream behavior. It is optional except when a cipher in CCM or OCB mode (e.g. ``'aes-128-ccm'``) is used. In that case, the ``authTagLength`` option is required and specifies the length of the authentication tag in bytes, see CCM mode. In GCM mode, the ``authTagLength`` option is not required but can be used to set the length of the authentication tag that will be returned by ``getAuthTag()`` and defaults to 16 bytes. For ``SIV``, ``GCM-SIV``, and ``chacha20-poly1305``, the ``authTagLength`` option defaults to 16 bytes. ``SIV`` and ``GCM-SIV`` only support 16-byte authentication tags.

**Arguments**

* ``algorithm`` (*string*)
* ``key`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``iv`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|null*)
* ``options`` (*Object*) — optional: ``stream.transform`` options with these additional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createCipheriv('value', 'example', [], {});
   console.log(result);


createDecipheriv
~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createDecipheriv(algorithm, key, iv[, options])

*Kind:* function · *Arguments:* 4

Creates and returns a ``Decipheriv`` object that uses the given ``algorithm``, ``key`` and initialization vector (``iv``).

The ``options`` argument controls cipher-specific settings and stream behavior. It is optional except when a cipher in CCM or OCB mode (e.g. ``'aes-128-ccm'``) is used. In that case, the ``authTagLength`` option is required and specifies the length of the authentication tag in bytes, see CCM mode. For ``chacha20-poly1305``, the ``authTagLength`` option defaults to 16 bytes and must be set to a different value if a different length is used. For AES-GCM, the ``authTagLength`` option has no default value when decrypting, and ``setAuthTag()`` will accept arbitrarily short authentication tags. This behavior is deprecated and subject to change (see DEP0182). <strong class="critical"> In the meantime, applications should either set the ``authTagLength`` option or check the actual authentication tag length before passing it to ``setAuthTag()``.</strong> For ``SIV`` and ``GCM-SIV``, the ``authTagLength`` option defaults to 16 bytes and only 16-byte authentication tags are supported.

**Arguments**

* ``algorithm`` (*string*)
* ``key`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``iv`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|null*)
* ``options`` (*Object*) — optional: ``stream.transform`` options with these additional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createDecipheriv('value', 'example', [], {});
   console.log(result);


createDiffieHellman
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createDiffieHellman(prime[, primeEncoding][, generator][, generatorEncoding])

*Kind:* function · *Arguments:* 4

Creates a ``DiffieHellman`` key exchange object using the supplied ``prime`` and an optional specific ``generator``.

The ``generator`` argument can be a number, string, or ``Buffer``. If ``generator`` is not specified, the value ``2`` is used.

**Arguments**

* ``prime`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``primeEncoding`` (*string*) — optional: The encoding of the ``prime`` string.
* ``generator`` (*number|string|ArrayBuffer|Buffer|TypedArray|DataView*) — optional
* ``generatorEncoding`` (*string*) — optional: The encoding of the ``generator`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createDiffieHellman([], 'utf8', [], 'utf8');
   console.log(result);


createDiffieHellmanGroup
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createDiffieHellmanGroup(name)

*Kind:* function · *Arguments:* 1

An alias for ``crypto.getDiffieHellman()``

**Arguments**

* ``name`` (*string*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createDiffieHellmanGroup('example');
   console.log(result);


createECDH
~~~~~~~~~~

.. code-block:: text

   crypto.createECDH(curveName)

*Kind:* function · *Arguments:* 1

Creates an Elliptic Curve Diffie-Hellman (``ECDH``) key exchange object using a predefined curve specified by the ``curveName`` string. Use ``crypto.getCurves()`` to obtain a list of available curve names. On recent OpenSSL releases, ``openssl ecparam -list_curves`` will also display the name and description of each available elliptic curve.

**Arguments**

* ``curveName`` (*string*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createECDH('value');
   console.log(result);


createHash
~~~~~~~~~~

.. code-block:: text

   crypto.createHash(algorithm[, options])

*Kind:* function · *Arguments:* 2

Creates and returns a ``Hash`` object that can be used to generate hash digests using the given ``algorithm``. Optional ``options`` argument controls stream behavior. For XOF hash functions such as ``'shake256'``, the ``outputLength`` option can be used to specify the desired output length in bytes.

The ``functionName`` and ``customization`` options apply only to cSHAKE-128 and cSHAKE-256. They are supported only when Node.js is built with OpenSSL 4.0 or later and the selected provider supports the corresponding digest parameters. Strings are encoded as UTF-8, and neither strings nor byte values may contain NUL bytes. Both options default to an empty byte string. For OpenSSL's built-in providers, ``functionName`` is case-sensitive and must be ``''``, ``'TupleHash'``, ``'ParallelHash'``, or ``'KMAC'``. Other providers can impose different restrictions. With both options empty, cSHAKE produces the same output as the corresponding SHAKE function for the same output length. ``cshake-128`` and ``cshake-256`` default to output lengths of 32 and 64 bytes, respectively.

**Arguments**

* ``algorithm`` (*string*)
* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   const {
     createReadStream,
   } = require('node:fs');
   const {
     createHash,
   } = require('node:crypto');
   const { argv } = require('node:process');

   const filename = argv[2];

   const hash = createHash('sha256');

   const input = createReadStream(filename);
   input.on('readable', () => {
     // Only one element is going to be produced by the
     // hash stream.
     const data = input.read();
     if (data)
       hash.update(data);
     else {
       console.log(`${hash.digest('hex')} ${filename}`);
     }
   });


createHmac
~~~~~~~~~~

.. code-block:: text

   crypto.createHmac(algorithm, key[, options])

*Kind:* function · *Arguments:* 3

Creates and returns an ``Hmac`` object that uses the given ``algorithm`` and ``key``. Optional ``options`` argument controls stream behavior.

The available algorithms depend on the version and configuration of OpenSSL on the platform. Examples are ``'sha256'`` and ``'sha512'``. ``crypto.getHashes()`` lists algorithms available to the hashing APIs, but HMAC imposes additional restrictions, so not every listed algorithm is suitable.

**Arguments**

* ``algorithm`` (*string*)
* ``key`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   const {
     createReadStream,
   } = require('node:fs');
   const {
     createHmac,
   } = require('node:crypto');
   const { argv } = require('node:process');

   const filename = argv[2];

   const hmac = createHmac('sha256', 'a secret');

   const input = createReadStream(filename);
   input.on('readable', () => {
     // Only one element is going to be produced by the
     // hash stream.
     const data = input.read();
     if (data)
       hmac.update(data);
     else {
       console.log(`${hmac.digest('hex')} ${filename}`);
     }
   });


createMac
~~~~~~~~~

.. code-block:: text

   crypto.createMac(algorithm, key[, options])

*Kind:* function · *Arguments:* 3

``algorithm`` must be a non-empty provider MAC name. The MAC-specific properties listed above are extensions to the standard ``stream.transform`` options and are passed only when the selected provider implementation advertises the corresponding parameter with the expected type. A supplied MAC-specific option that the selected implementation does not support causes an error.

The following table summarizes the MAC-specific options accepted by MAC implementations in OpenSSL's built-in providers. The ``key`` argument is required for every MAC. The table lists only MAC-specific options; standard ``stream.transform`` options remain available for every family.

**Arguments**

* ``algorithm`` (*string*): The name of the MAC algorithm.
* ``key`` (*ArrayBuffer|Buffer|TypedArray|DataView|KeyObject*)
* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createMac('value', 'example', {});
   console.log(result);


createPrivateKey
~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createPrivateKey(key)

*Kind:* function · *Arguments:* 1

Creates and returns a new key object containing a private key. If ``key`` is a string or ``Buffer``, ``format`` is assumed to be ``'pem'``; otherwise, ``key`` must be an object with the properties described above.

If the private key is encrypted, a ``passphrase`` must be specified. The length of the passphrase is limited to 1024 bytes.

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|URL*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createPrivateKey('example');
   console.log(result);


createPublicKey
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createPublicKey(key)

*Kind:* function · *Arguments:* 1

Creates and returns a new key object containing a public key. If ``key`` is a string or ``Buffer``, ``format`` is assumed to be ``'pem'``; if ``key`` is a ``KeyObject`` with type ``'private'``, the public key is derived from the given private key; otherwise, ``key`` must be an object with the properties described above.

If the format is ``'pem'``, the ``'key'`` may also be an X.509 certificate.

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createPublicKey('example');
   console.log(result);


createSecretKey
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.createSecretKey(key[, encoding])

*Kind:* function · *Arguments:* 2

Creates and returns a new key object containing a secret key for symmetric encryption or ``Hmac``.

**Arguments**

* ``key`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The string encoding when ``key`` is a string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createSecretKey('example', 'utf8');
   console.log(result);


createSign
~~~~~~~~~~

.. code-block:: text

   crypto.createSign(algorithm[, options])

*Kind:* function · *Arguments:* 2

Creates and returns a ``Sign`` object that uses the given ``algorithm``. Use ``crypto.getHashes()`` to obtain the names available to the hashing APIs. The key type and signature scheme can impose additional restrictions on which digests can be used. Optional ``options`` argument controls the ``stream.Writable`` behavior.

In some cases, a ``Sign`` instance can be created using the name of a signature algorithm, such as ``'RSA-SHA256'``, instead of a digest algorithm. This will use the corresponding digest algorithm. This does not work for all signature algorithms, such as ``'ecdsa-with-SHA256'``, so it is best to always use digest algorithm names.

**Arguments**

* ``algorithm`` (*string*)
* ``options`` (*Object*) — optional: ``stream.Writable`` options

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createSign('value', {});
   console.log(result);


createVerify
~~~~~~~~~~~~

.. code-block:: text

   crypto.createVerify(algorithm[, options])

*Kind:* function · *Arguments:* 2

Creates and returns a ``Verify`` object that uses the given algorithm. Use ``crypto.getHashes()`` to obtain the names available to the hashing APIs. The key type and signature scheme can impose additional restrictions on which digests can be used. Optional ``options`` argument controls the ``stream.Writable`` behavior.

In some cases, a ``Verify`` instance can be created using the name of a signature algorithm, such as ``'RSA-SHA256'``, instead of a digest algorithm. This will use the corresponding digest algorithm. This does not work for all signature algorithms, such as ``'ecdsa-with-SHA256'``, so it is best to always use digest algorithm names.

**Arguments**

* ``algorithm`` (*string*)
* ``options`` (*Object*) — optional: ``stream.Writable`` options

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.createVerify('value', {});
   console.log(result);


decapsulate
~~~~~~~~~~~

.. code-block:: text

   crypto.decapsulate(key, ciphertext[, callback])

*Kind:* function · *Arguments:* 3

Key decapsulation using a KEM algorithm with a private key.

Supported key types and their KEM algorithms are:

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|URL*): Private Key
* ``ciphertext`` (*ArrayBuffer|Buffer|TypedArray|DataView*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.decapsulate('example', [], () => {});
   console.log(result);


Decipheriv
~~~~~~~~~~

.. code-block:: text

   Decipheriv

*Kind:* class · *Arguments:* 0

Instances of the ``Decipheriv`` class are used to decrypt data. The class can be used in one of two ways:

* As a stream that is both readable and writable, where plain encrypted data is written to produce unencrypted data on the readable side, or * Using the ``decipher.update()`` and ``decipher.final()`` methods to produce the unencrypted data.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     scryptSync,
     createDecipheriv,
   } = require('node:crypto');
   const { Buffer } = require('node:buffer');

   const algorithm = 'aes-192-cbc';
   const password = 'Password used to generate key';
   // Key length is dependent on the algorithm. In this case for aes192, it is
   // 24 bytes (192 bits).
   // Use the async `crypto.scrypt()` instead.
   const key = scryptSync(password, 'salt', 24);
   // The IV is usually passed along with the ciphertext.
   const iv = Buffer.alloc(16, 0); // Initialization vector.

   const decipher = createDecipheriv(algorithm, key, iv);

   let decrypted = '';
   decipher.on('readable', () => {
     let chunk;
     while (null !== (chunk = decipher.read())) {
       decrypted += chunk.toString('utf8');
     }
   });
   decipher.on('end', () => {
     console.log(decrypted);
     // Prints: some clear text data
   });

   // Encrypted with same algorithm, key and iv.
   const encrypted =
     'e5f79c5915c02171eec6b212d5520d44480993d7d622a7c4c2da32f6efda0ffa';
   decipher.write(encrypted, 'hex');
   decipher.end();


DiffieHellman
~~~~~~~~~~~~~

.. code-block:: text

   DiffieHellman

*Kind:* class · *Arguments:* 0

The ``DiffieHellman`` class is a utility for creating Diffie-Hellman key exchanges.

Instances of the ``DiffieHellman`` class can be created using the ``crypto.createDiffieHellman()`` function.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const {
     createDiffieHellman,
   } = require('node:crypto');

   // Generate Alice's keys...
   const alice = createDiffieHellman(2048);
   const aliceKey = alice.generateKeys();

   // Generate Bob's keys...
   const bob = createDiffieHellman(alice.getPrime(), alice.getGenerator());
   const bobKey = bob.generateKeys();

   // Exchange and generate the secret...
   const aliceSecret = alice.computeSecret(bobKey);
   const bobSecret = bob.computeSecret(aliceKey);

   // OK
   assert.strictEqual(aliceSecret.toString('hex'), bobSecret.toString('hex'));


diffieHellman
~~~~~~~~~~~~~

.. code-block:: text

   crypto.diffieHellman(options[, callback])

*Kind:* function · *Arguments:* 2

Computes the Diffie-Hellman shared secret based on a ``privateKey`` and a ``publicKey``. Both keys must represent the same asymmetric key type and must support either the DH or ECDH operation.

If ``options.privateKey`` is not a ``KeyObject``, this function behaves as if ``options.privateKey`` had been passed to ``crypto.createPrivateKey()``.

**Arguments**

* ``options`` (*Object*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.diffieHellman({}, () => {});
   console.log(result);


DiffieHellmanGroup
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   DiffieHellmanGroup

*Kind:* class · *Arguments:* 0

The ``DiffieHellmanGroup`` class takes a well-known modp group as its argument. It works the same as ``DiffieHellman``, except that it does not allow changing its keys after creation. In other words, it does not implement ``setPublicKey()`` or ``setPrivateKey()`` methods.

The following groups are supported:

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { createDiffieHellmanGroup } = require('node:crypto');
   const dh = createDiffieHellmanGroup('modp16');


ECDH
~~~~

.. code-block:: text

   ECDH

*Kind:* class · *Arguments:* 0

The ``ECDH`` class is a utility for creating Elliptic Curve Diffie-Hellman (ECDH) key exchanges.

Instances of the ``ECDH`` class can be created using the ``crypto.createECDH()`` function.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const assert = require('node:assert');

   const {
     createECDH,
   } = require('node:crypto');

   // Generate Alice's keys...
   const alice = createECDH('secp521r1');
   const aliceKey = alice.generateKeys();

   // Generate Bob's keys...
   const bob = createECDH('secp521r1');
   const bobKey = bob.generateKeys();

   // Exchange and generate the secret...
   const aliceSecret = alice.computeSecret(bobKey);
   const bobSecret = bob.computeSecret(aliceKey);

   assert.strictEqual(aliceSecret.toString('hex'), bobSecret.toString('hex'));
   // OK


encapsulate
~~~~~~~~~~~

.. code-block:: text

   crypto.encapsulate(key[, callback])

*Kind:* function · *Arguments:* 2

Key encapsulation using a KEM algorithm with a public key.

Supported key types and their KEM algorithms are:

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject*): Public Key
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.encapsulate('example', () => {});
   console.log(result);


generateKey
~~~~~~~~~~~

.. code-block:: text

   crypto.generateKey(type, options, callback)

*Kind:* function · *Arguments:* 3

Asynchronously generates a new random secret key of the given ``length``. The ``type`` will determine which validations will be performed on the ``length``.

The size of a generated HMAC key should not exceed the block size of the underlying hash function. See ``crypto.createHmac()`` for more information.

**Arguments**

* ``type`` (*string*): The intended use of the generated secret key. Currently
* ``options`` (*Object*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const {
     generateKey,
   } = require('node:crypto');

   generateKey('hmac', { length: 512 }, (err, key) => {
     if (err) throw err;
     console.log(key.export().toString('hex'));  // 46e..........620
   });


generateKeyPair
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.generateKeyPair(type, options, callback)

*Kind:* function · *Arguments:* 3

Generates a new asymmetric key pair of the given ``type``. See the supported asymmetric key types.

For RSA-PSS keys, ``crypto.getHashes()`` lists algorithms available to the hashing APIs, but not every listed digest can be encoded in RSA-PSS parameters or is supported by the active RSA implementation.

**Arguments**

* ``type`` (*string*): The asymmetric key type to generate. See the
* ``options`` (*Object*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const {
     generateKeyPair,
   } = require('node:crypto');

   generateKeyPair('rsa', {
     modulusLength: 4096,
     publicKeyEncoding: {
       type: 'spki',
       format: 'pem',
     },
     privateKeyEncoding: {
       type: 'pkcs8',
       format: 'pem',
       cipher: 'aes-256-cbc',
       passphrase: 'top secret',
     },
   }, (err, publicKey, privateKey) => {
     // Handle errors and use the generated key pair.
   });


generateKeyPairSync
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.generateKeyPairSync(type, options)

*Kind:* function · *Arguments:* 2

Generates a new asymmetric key pair of the given ``type``. See the supported asymmetric key types.

For RSA-PSS keys, ``crypto.getHashes()`` lists algorithms available to the hashing APIs, but not every listed digest can be encoded in RSA-PSS parameters or is supported by the active RSA implementation.

**Arguments**

* ``type`` (*string*): The asymmetric key type to generate. See the
* ``options`` (*Object*)

**Example**

.. code-block:: javascript

   const {
     generateKeyPairSync,
   } = require('node:crypto');

   const {
     publicKey,
     privateKey,
   } = generateKeyPairSync('rsa', {
     modulusLength: 4096,
     publicKeyEncoding: {
       type: 'spki',
       format: 'pem',
     },
     privateKeyEncoding: {
       type: 'pkcs8',
       format: 'pem',
       cipher: 'aes-256-cbc',
       passphrase: 'top secret',
     },
   });


generateKeySync
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.generateKeySync(type, options)

*Kind:* function · *Arguments:* 2

Synchronously generates a new random secret key of the given ``length``. The ``type`` will determine which validations will be performed on the ``length``.

The size of a generated HMAC key should not exceed the block size of the underlying hash function. See ``crypto.createHmac()`` for more information.

**Arguments**

* ``type`` (*string*): The intended use of the generated secret key. Currently
* ``options`` (*Object*)

**Example**

.. code-block:: javascript

   const {
     generateKeySync,
   } = require('node:crypto');

   const key = generateKeySync('hmac', { length: 512 });
   console.log(key.export().toString('hex'));  // e89..........41e


generatePrime
~~~~~~~~~~~~~

.. code-block:: text

   crypto.generatePrime(size[, options], callback)

*Kind:* function · *Arguments:* 3

Generates a pseudorandom prime of ``size`` bits.

If ``options.safe`` is ``true``, the prime will be a safe prime -- that is, ``(prime - 1) / 2`` will also be a prime.

**Arguments**

* ``size`` (*number*): The size (in bits) of the prime to generate.
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.generatePrime(0, {}, () => {});
   console.log(result);


generatePrimeSync
~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.generatePrimeSync(size[, options])

*Kind:* function · *Arguments:* 2

Generates a pseudorandom prime of ``size`` bits.

If ``options.safe`` is ``true``, the prime will be a safe prime -- that is, ``(prime - 1) / 2`` will also be a prime.

**Arguments**

* ``size`` (*number*): The size (in bits) of the prime to generate.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.generatePrimeSync(0, {});
   console.log(result);


getCipherInfo
~~~~~~~~~~~~~

.. code-block:: text

   crypto.getCipherInfo(nameOrNid[, options])

*Kind:* function · *Arguments:* 2

Returns information about a given cipher.

Some ciphers accept variable length keys and initialization vectors. By default, the ``crypto.getCipherInfo()`` method will return the default values for these ciphers. To test if a given key length or iv length is acceptable for given cipher, use the ``keyLength`` and ``ivLength`` options. If the given values are unacceptable, ``undefined`` will be returned.

**Arguments**

* ``nameOrNid`` (*string|number*): The name or nid of the cipher to query.
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.getCipherInfo('value', {});
   console.log(result);


getCiphers
~~~~~~~~~~

.. code-block:: text

   crypto.getCiphers()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const {
     getCiphers,
   } = require('node:crypto');

   console.log(getCiphers()); // ['aes-128-cbc', 'aes-128-ccm', ...]


getCurves
~~~~~~~~~

.. code-block:: text

   crypto.getCurves()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const {
     getCurves,
   } = require('node:crypto');

   console.log(getCurves()); // ['Oakley-EC2N-3', 'Oakley-EC2N-4', ...]


getDiffieHellman
~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.getDiffieHellman(groupName)

*Kind:* function · *Arguments:* 1

Creates a predefined ``DiffieHellmanGroup`` key exchange object. The supported groups are listed in the documentation for ``DiffieHellmanGroup``.

The returned object mimics the interface of objects created by ``crypto.createDiffieHellman()``, but will not allow changing the keys (with ``diffieHellman.setPublicKey()``, for example). The advantage of using this method is that the parties do not have to generate nor exchange a group modulus beforehand, saving both processor and communication time.

**Arguments**

* ``groupName`` (*string*)

**Example**

.. code-block:: javascript

   const {
     getDiffieHellman,
   } = require('node:crypto');

   const alice = getDiffieHellman('modp14');
   const bob = getDiffieHellman('modp14');

   alice.generateKeys();
   bob.generateKeys();

   const aliceSecret = alice.computeSecret(bob.getPublicKey(), null, 'hex');
   const bobSecret = bob.computeSecret(alice.getPublicKey(), null, 'hex');

   /* aliceSecret and bobSecret should be the same */
   console.log(aliceSecret === bobSecret);


getFips
~~~~~~~

.. code-block:: text

   crypto.getFips()

*Kind:* function · *Arguments:* 0

With OpenSSL 3, this reports whether the default property query includes ``fips=yes``. It does not establish that a FIPS provider is loaded or validated. It can return ``1`` even when a requested cryptographic implementation cannot be fetched because no loaded provider supplies a match for ``fips=yes``. See FIPS mode.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.getFips();
   console.log(result);


getHashes
~~~~~~~~~

.. code-block:: text

   crypto.getHashes()

*Kind:* function · *Arguments:* 0

This is the authoritative Node.js list of hash algorithms available to ``crypto.createHash()`` and ``crypto.hash()`` in the current process. With OpenSSL 3 or later, the list depends on the loaded providers and the default property query in effect when the list is first generated. Some listed algorithms can require API-specific options, such as ``outputLength`` for XOF hash functions.

A listed hash algorithm is not necessarily supported by APIs that combine a digest with another cryptographic operation, such as HMAC, key derivation, or signing. Those operations can apply additional restrictions.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const {
     getHashes,
   } = require('node:crypto');

   console.log(getHashes()); // ['DSA', 'DSA-SHA', 'DSA-SHA1', ...]


getMacs
~~~~~~~

.. code-block:: text

   crypto.getMacs()

*Kind:* function · *Arguments:* 0

Returns MAC names exposed by loaded OpenSSL providers that match the active default property query. Duplicate names and numeric OID aliases are omitted. On builds without OpenSSL ``EVP_MAC`` support, this function returns an empty array.

The returned names describe implementations that OpenSSL can fetch. They do not guarantee that ``crypto.createMac()`` can initialize the MAC without additional options. A provider can require additional parameters or a key with algorithm-specific properties, and it can expose parameters that this API does not support.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const { getMacs } = await import('node:crypto');

   console.log(getMacs());
   // ['blake2bmac', 'blake2smac', 'cmac', 'gmac', 'hmac', ...]


getRandomValues
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.getRandomValues(typedArray)

*Kind:* function · *Arguments:* 1

A convenient alias for ``crypto.webcrypto.getRandomValues()``. This implementation is not compliant with the Web Crypto spec, to write web-compatible code use ``crypto.webcrypto.getRandomValues()`` instead.

**Arguments**

* ``typedArray`` (*Buffer|TypedArray|DataView|ArrayBuffer*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.getRandomValues([]);
   console.log(result);


Hash
~~~~

.. code-block:: text

   Hash

*Kind:* class · *Arguments:* 0

The ``Hash`` class is a utility for creating hash digests of data. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written to produce a computed hash digest on the readable side, or * Using the ``hash.update()`` and ``hash.digest()`` methods to produce the computed hash.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     createHash,
   } = require('node:crypto');

   const hash = createHash('sha256');

   hash.on('readable', () => {
     // Only one element is going to be produced by the
     // hash stream.
     const data = hash.read();
     if (data) {
       console.log(data.toString('hex'));
       // Prints:
       //   6a2da20943931e9834fc12cfe5bb47bbd9ae43489a30726962b576f4e3993e50
     }
   });

   hash.write('some data to hash');
   hash.end();


hash
~~~~

.. code-block:: text

   crypto.hash(algorithm, data[, options])

*Kind:* function · *Arguments:* 3

A utility for creating one-shot hash digests of data. It can be faster than the object-based ``crypto.createHash()`` when hashing a smaller amount of data (<= 5MB) that's readily available. If the data can be big or if it is streamed, it's still recommended to use ``crypto.createHash()`` instead.

The available algorithms depend on the version and configuration of OpenSSL on the platform. Examples are ``'sha256'`` and ``'sha512'``. Use ``crypto.getHashes()`` to obtain the list of hash algorithms available to the Node.js process.

**Arguments**

* ``algorithm`` (*string|undefined*)
* ``data`` (*string|Buffer|TypedArray|DataView*): When ``data`` is a
* ``options`` (*Object|string*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');
   const { Buffer } = require('node:buffer');

   // Hashing a string and return the result as a hex-encoded string.
   const string = 'Node.js';
   // 10b3493287f831e81a438811a1ffba01f8cec4b7
   console.log(crypto.hash('sha1', string));

   // Encode a base64-encoded string into a Buffer, hash it and return
   // the result as a buffer.
   const base64 = 'Tm9kZS5qcw==';
   // <Buffer 10 b3 49 32 87 f8 31 e8 1a 43 88 11 a1 ff ba 01 f8 ce c4 b7>
   console.log(crypto.hash('sha1', Buffer.from(base64, 'base64'), 'buffer'));


hkdf
~~~~

.. code-block:: text

   crypto.hkdf(digest, ikm, salt, info, keylen, callback)

*Kind:* function · *Arguments:* 6

HKDF is a simple key derivation function defined in RFC 5869. The given ``ikm``, ``salt`` and ``info`` are used with the ``digest`` to derive a key of ``keylen`` bytes. The available digest algorithms depend on the version and configuration of OpenSSL. HKDF uses HMAC internally. ``crypto.getHashes()`` lists algorithms available to the hashing APIs, but not every listed algorithm is necessarily suitable for HMAC or HKDF.

The supplied ``callback`` function is called with two arguments: ``err`` and ``derivedKey``. If an error occurs while deriving the key, ``err`` will be set; otherwise ``err`` will be ``null``. The successfully generated ``derivedKey`` will be passed to the callback as an ``ArrayBuffer``. An error will be thrown if any of the input arguments specify invalid values or types.

**Arguments**

* ``digest`` (*string*): The digest algorithm to use.
* ``ikm`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject*): The input
* ``salt`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*): The salt value. Must
* ``info`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*): Additional info value.
* ``keylen`` (*number*): The length of the key to generate. Must be greater than 0.
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const {
     hkdf,
   } = require('node:crypto');
   const { Buffer } = require('node:buffer');

   hkdf('sha512', 'key', 'salt', 'info', 64, (err, derivedKey) => {
     if (err) throw err;
     console.log(Buffer.from(derivedKey).toString('hex'));  // '24156e2...5391653'
   });


hkdfSync
~~~~~~~~

.. code-block:: text

   crypto.hkdfSync(digest, ikm, salt, info, keylen)

*Kind:* function · *Arguments:* 5

Provides a synchronous HKDF key derivation function as defined in RFC 5869. The given ``ikm``, ``salt`` and ``info`` are used with the ``digest`` to derive a key of ``keylen`` bytes. The available digest algorithms depend on the version and configuration of OpenSSL. HKDF uses HMAC internally. ``crypto.getHashes()`` lists algorithms available to the hashing APIs, but not every listed algorithm is necessarily suitable for HMAC or HKDF.

The successfully generated ``derivedKey`` will be returned as an ``ArrayBuffer``.

**Arguments**

* ``digest`` (*string*): The digest algorithm to use.
* ``ikm`` (*string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject*): The input
* ``salt`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*): The salt value. Must
* ``info`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*): Additional info value.
* ``keylen`` (*number*): The length of the key to generate. Must be greater than 0.

**Example**

.. code-block:: javascript

   const {
     hkdfSync,
   } = require('node:crypto');
   const { Buffer } = require('node:buffer');

   const derivedKey = hkdfSync('sha512', 'key', 'salt', 'info', 64);
   console.log(Buffer.from(derivedKey).toString('hex'));  // '24156e2...5391653'


Hmac
~~~~

.. code-block:: text

   Hmac

*Kind:* class · *Arguments:* 0

The ``Hmac`` class is a utility for creating cryptographic HMAC digests. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written to produce a computed HMAC digest on the readable side, or * Using the ``hmac.update()`` and ``hmac.digest()`` methods to produce the computed HMAC digest.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     createHmac,
   } = require('node:crypto');

   const hmac = createHmac('sha256', 'a secret');

   hmac.on('readable', () => {
     // Only one element is going to be produced by the
     // hash stream.
     const data = hmac.read();
     if (data) {
       console.log(data.toString('hex'));
       // Prints:
       //   7fd04df92f636fd450bc841c9418e5825c17f33ad9c87c518115a45971f7f77e
     }
   });

   hmac.write('some data to hash');
   hmac.end();


KeyObject
~~~~~~~~~

.. code-block:: text

   KeyObject

*Kind:* class · *Arguments:* 0

Node.js uses a ``KeyObject`` class to represent a symmetric or asymmetric key, and each kind of key exposes different functions. The ``crypto.createSecretKey()``, ``crypto.createPublicKey()`` and ``crypto.createPrivateKey()`` methods are used to create ``KeyObject`` instances. ``KeyObject`` objects are not to be created directly using the ``new`` keyword.

Most applications should consider using the new ``KeyObject`` API instead of passing keys as strings or ``Buffer``\ s due to improved security features.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { KeyObject } = require('node:crypto');
   const { subtle } = globalThis.crypto;

   (async function() {
     const key = await subtle.generateKey({
       name: 'HMAC',
       hash: 'SHA-256',
       length: 256,
     }, true, ['sign', 'verify']);

     const keyObject = KeyObject.from(key);
     console.log(keyObject.symmetricKeySize);
     // Prints: 32 (symmetric key size in bytes)
   })();


Mac
~~~

.. code-block:: text

   Mac

*Kind:* class · *Arguments:* 0

The ``Mac`` class computes message authentication codes using MAC implementations supplied by OpenSSL providers. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written and one authentication tag is produced on the readable side when the writable side ends; or * By calling ``mac.update()`` one or more times followed by ``mac.final()``.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { createMac, randomBytes } = await import('node:crypto');

   const key = randomBytes(16);
   const mac = createMac('CMAC', key, {
     cipher: 'AES-128-CBC',
   });

   mac.update('some data to authenticate');
   console.log(mac.final('hex'));


parsePKCS12
~~~~~~~~~~~

.. code-block:: text

   crypto.parsePKCS12(bundle[, options])

*Kind:* function · *Arguments:* 2

Parses a PKCS#12 bundle, commonly stored with a ``.p12`` or ``.pfx`` extension, and returns its private key and certificates.

**Arguments**

* ``bundle`` (*ArrayBuffer|Buffer|TypedArray|DataView*): A DER-encoded PKCS#12
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   import { parsePKCS12 } from 'node:crypto';
   import { readFileSync } from 'node:fs';

   const { privateKey, certificate, additionalCertificates } = parsePKCS12(
     readFileSync('bundle.p12'),
     { passphrase: 'secret' },
   );


pbkdf2
~~~~~~

.. code-block:: text

   crypto.pbkdf2(password, salt, iterations, keylen, digest, callback)

*Kind:* function · *Arguments:* 6

Provides an asynchronous Password-Based Key Derivation Function 2 (PBKDF2) implementation. A selected HMAC digest algorithm specified by ``digest`` is applied to derive a key of the requested byte length (``keylen``) from the ``password``, ``salt`` and ``iterations``.

The supplied ``callback`` function is called with two arguments: ``err`` and ``derivedKey``. If an error occurs while deriving the key, ``err`` will be set; otherwise ``err`` will be ``null``. By default, the successfully generated ``derivedKey`` will be passed to the callback as a ``Buffer``. An error will be thrown if any of the input arguments specify invalid values or types.

**Arguments**

* ``password`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``salt`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``iterations`` (*number*)
* ``keylen`` (*number*)
* ``digest`` (*string*)
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const {
     pbkdf2,
   } = require('node:crypto');

   pbkdf2('secret', 'salt', 100000, 64, 'sha512', (err, derivedKey) => {
     if (err) throw err;
     console.log(derivedKey.toString('hex'));  // '3745e48...08d59ae'
   });


pbkdf2Sync
~~~~~~~~~~

.. code-block:: text

   crypto.pbkdf2Sync(password, salt, iterations, keylen, digest)

*Kind:* function · *Arguments:* 5

Provides a synchronous Password-Based Key Derivation Function 2 (PBKDF2) implementation. A selected HMAC digest algorithm specified by ``digest`` is applied to derive a key of the requested byte length (``keylen``) from the ``password``, ``salt`` and ``iterations``.

If an error occurs an ``Error`` will be thrown, otherwise the derived key will be returned as a ``Buffer``.

**Arguments**

* ``password`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``salt`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``iterations`` (*number*)
* ``keylen`` (*number*)
* ``digest`` (*string*)

**Example**

.. code-block:: javascript

   const {
     pbkdf2Sync,
   } = require('node:crypto');

   const key = pbkdf2Sync('secret', 'salt', 100000, 64, 'sha512');
   console.log(key.toString('hex'));  // '3745e48...08d59ae'


privateDecrypt
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.privateDecrypt(privateKey, buffer)

*Kind:* function · *Arguments:* 2

Decrypts ``buffer`` with ``privateKey``. ``buffer`` was previously encrypted using the corresponding public key, for example using ``crypto.publicEncrypt()``.

``crypto.getHashes()`` lists algorithms available to the hashing APIs, but the active RSA implementation can impose additional restrictions on digests used for OAEP or MGF1.

**Arguments**

* ``privateKey`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey|URL*)
* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.privateDecrypt([], []);
   console.log(result);


privateEncrypt
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.privateEncrypt(privateKey, buffer)

*Kind:* function · *Arguments:* 2

Encrypts ``buffer`` with ``privateKey``. The returned data can be decrypted using the corresponding public key, for example using ``crypto.publicDecrypt()``.

If ``privateKey`` is not a ``KeyObject``, this function behaves as if ``privateKey`` had been passed to ``crypto.createPrivateKey()``. If it is an object, the ``padding`` property can be passed. Otherwise, this function uses ``RSA_PKCS1_PADDING``.

**Arguments**

* ``privateKey`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey|URL*)
* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.privateEncrypt([], []);
   console.log(result);


prng
~~~~

.. code-block:: text

   crypto.prng()

*Kind:* function · *Arguments:* 0

Type: Documentation-only (supports ``--pending-deprecation``)

In recent versions of Node.js, there is no difference between ``crypto.randomBytes()`` and ``crypto.pseudoRandomBytes()``. The latter is deprecated along with the undocumented aliases ``crypto.prng()`` and ``crypto.rng()`` in favor of ``crypto.randomBytes()`` and might be removed in a future release.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.prng();
   console.log(result);


pseudoRandomBytes
~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.pseudoRandomBytes(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.pseudoRandomBytes();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


publicDecrypt
~~~~~~~~~~~~~

.. code-block:: text

   crypto.publicDecrypt(key, buffer)

*Kind:* function · *Arguments:* 2

Decrypts ``buffer`` with ``key``. ``buffer`` was previously encrypted using the corresponding private key, for example using ``crypto.privateEncrypt()``.

If ``key`` is not a ``KeyObject``, this function behaves as if ``key`` had been passed to ``crypto.createPublicKey()``. If it is an object, the ``padding`` property can be passed. Otherwise, this function uses ``RSA_PKCS1_PADDING``.

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.publicDecrypt('example', []);
   console.log(result);


publicEncrypt
~~~~~~~~~~~~~

.. code-block:: text

   crypto.publicEncrypt(key, buffer)

*Kind:* function · *Arguments:* 2

Encrypts the content of ``buffer`` with ``key`` and returns a new ``Buffer`` with encrypted content. The returned data can be decrypted using the corresponding private key, for example using ``crypto.privateDecrypt()``.

``crypto.getHashes()`` lists algorithms available to the hashing APIs, but the active RSA implementation can impose additional restrictions on digests used for OAEP or MGF1.

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.publicEncrypt('example', []);
   console.log(result);


randomBytes
~~~~~~~~~~~

.. code-block:: text

   crypto.randomBytes(size[, callback])

*Kind:* function · *Arguments:* 2

Generates cryptographically strong pseudorandom data. The ``size`` argument is a number indicating the number of bytes to generate.

If a ``callback`` function is provided, the bytes are generated asynchronously and the ``callback`` function is invoked with two arguments: ``err`` and ``buf``. If an error occurs, ``err`` will be an ``Error`` object; otherwise it is ``null``. The ``buf`` argument is a ``Buffer`` containing the generated bytes.

**Arguments**

* ``size`` (*number*): The number of bytes to generate.  The ``size`` must
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   // Asynchronous
   const {
     randomBytes,
   } = require('node:crypto');

   randomBytes(256, (err, buf) => {
     if (err) throw err;
     console.log(`${buf.length} bytes of random data: ${buf.toString('hex')}`);
   });


randomFill
~~~~~~~~~~

.. code-block:: text

   crypto.randomFill(buffer[, offset][, size], callback)

*Kind:* function · *Arguments:* 4

This function is similar to ``crypto.randomBytes()`` but requires the first argument to be a ``Buffer`` that will be filled. It also requires that a callback is passed in.

If the ``callback`` function is not provided, an error will be thrown.

**Arguments**

* ``buffer`` (*ArrayBuffer|Buffer|TypedArray|DataView*): Must be supplied. The
* ``offset`` (*number*) — optional: The start position, in elements for a ``TypedArray`` and in
* ``size`` (*number*) — optional: The amount to fill, in the same units as ``offset``.
* ``callback`` (*Function*): ``function(err, buf) {}``.

**Example**

.. code-block:: javascript

   const { randomFill } = require('node:crypto');
   const { Buffer } = require('node:buffer');

   const buf = Buffer.alloc(10);
   randomFill(buf, (err, buf) => {
     if (err) throw err;
     console.log(buf.toString('hex'));
   });

   randomFill(buf, 5, (err, buf) => {
     if (err) throw err;
     console.log(buf.toString('hex'));
   });

   // The above is equivalent to the following:
   randomFill(buf, 5, 5, (err, buf) => {
     if (err) throw err;
     console.log(buf.toString('hex'));
   });


randomFillSync
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.randomFillSync(buffer[, offset][, size])

*Kind:* function · *Arguments:* 3

Synchronous version of ``crypto.randomFill()``.

Any ``ArrayBuffer``, ``TypedArray`` or ``DataView`` instance may be passed as ``buffer``.

**Arguments**

* ``buffer`` (*ArrayBuffer|Buffer|TypedArray|DataView*): Must be supplied. The
* ``offset`` (*number*) — optional: The start position, in elements for a ``TypedArray`` and in
* ``size`` (*number*) — optional: The amount to fill, in the same units as ``offset``.

**Example**

.. code-block:: javascript

   const { randomFillSync } = require('node:crypto');
   const { Buffer } = require('node:buffer');

   const buf = Buffer.alloc(10);
   console.log(randomFillSync(buf).toString('hex'));

   randomFillSync(buf, 5);
   console.log(buf.toString('hex'));

   // The above is equivalent to the following:
   randomFillSync(buf, 5, 5);
   console.log(buf.toString('hex'));


randomInt
~~~~~~~~~

.. code-block:: text

   crypto.randomInt([min, ]max[, callback])

*Kind:* function · *Arguments:* 3

Return a random integer ``n`` such that ``min <= n < max``.  This implementation avoids modulo bias.

The range (``max - min``) must be less than 2<sup>48</sup>. ``min`` and ``max`` must be safe integers.

**Arguments**

* ``min`` (*integer*) — optional: Start of random range (inclusive). **Default:** ``0``.
* ``max`` (*integer*) — optional: End of random range (exclusive).
* ``callback`` (*Function*) — optional: ``function(err, n) {}``.

**Example**

.. code-block:: javascript

   // Asynchronous
   const {
     randomInt,
   } = require('node:crypto');

   randomInt(3, (err, n) => {
     if (err) throw err;
     console.log(`Random number chosen from (0, 1, 2): ${n}`);
   });


randomUUID
~~~~~~~~~~

.. code-block:: text

   crypto.randomUUID([options])

*Kind:* function · *Arguments:* 1

Generates a random RFC 4122 version 4 UUID. The UUID is generated using a cryptographic pseudorandom number generator.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.randomUUID({});
   console.log(result);


randomUUIDv7
~~~~~~~~~~~~

.. code-block:: text

   crypto.randomUUIDv7([options])

*Kind:* function · *Arguments:* 1

Generates a random RFC 9562 version 7 UUID. The UUID contains a millisecond precision Unix timestamp in the most significant 48 bits, followed by cryptographically secure random bits for the remaining fields, making it suitable for use as a database key with time-based sorting. The embedded timestamp relies on a non-monotonic clock and is not guaranteed to be strictly increasing.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.randomUUIDv7({});
   console.log(result);


rng
~~~

.. code-block:: text

   crypto.rng(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.rng();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


scrypt
~~~~~~

.. code-block:: text

   crypto.scrypt(password, salt, keylen[, options], callback)

*Kind:* function · *Arguments:* 5

Provides an asynchronous scrypt implementation. Scrypt is a password-based key derivation function that is designed to be expensive computationally and memory-wise in order to make brute-force attacks unrewarding.

The ``salt`` should be as unique as possible. It is recommended that a salt is random and at least 16 bytes long. See NIST SP 800-132 for details.

**Arguments**

* ``password`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``salt`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``keylen`` (*number*)
* ``options`` (*Object*) — optional
* ``callback`` (*Function*)

**Example**

.. code-block:: javascript

   const {
     scrypt,
   } = require('node:crypto');

   // Using the factory defaults.
   scrypt('password', 'salt', 64, (err, derivedKey) => {
     if (err) throw err;
     console.log(derivedKey.toString('hex'));  // '3745e48...08d59ae'
   });
   // Using a custom N parameter. Must be a power of two.
   scrypt('password', 'salt', 64, { N: 1024 }, (err, derivedKey) => {
     if (err) throw err;
     console.log(derivedKey.toString('hex'));  // '3745e48...aa39b34'
   });


scryptSync
~~~~~~~~~~

.. code-block:: text

   crypto.scryptSync(password, salt, keylen[, options])

*Kind:* function · *Arguments:* 4

Provides a synchronous scrypt implementation. Scrypt is a password-based key derivation function that is designed to be expensive computationally and memory-wise in order to make brute-force attacks unrewarding.

The ``salt`` should be as unique as possible. It is recommended that a salt is random and at least 16 bytes long. See NIST SP 800-132 for details.

**Arguments**

* ``password`` (*string|Buffer|TypedArray|DataView*)
* ``salt`` (*string|Buffer|TypedArray|DataView*)
* ``keylen`` (*number*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const {
     scryptSync,
   } = require('node:crypto');
   // Using the factory defaults.

   const key1 = scryptSync('password', 'salt', 64);
   console.log(key1.toString('hex'));  // '3745e48...08d59ae'
   // Using a custom N parameter. Must be a power of two.
   const key2 = scryptSync('password', 'salt', 64, { N: 1024 });
   console.log(key2.toString('hex'));  // '3745e48...aa39b34'


secureHeapUsed
~~~~~~~~~~~~~~

.. code-block:: text

   crypto.secureHeapUsed()

*Kind:* function · *Arguments:* 0

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.secureHeapUsed();
   console.log(result);


setEngine
~~~~~~~~~

.. code-block:: text

   crypto.setEngine(engine[, flags])

*Kind:* function · *Arguments:* 2

Load and set the ``engine`` for some or all OpenSSL functions (selected by flags). Use of this API is deprecated because custom engine support has been deprecated since OpenSSL 3.

``engine`` could be either an id or a path to the engine's shared library.

**Arguments**

* ``engine`` (*string*)
* ``flags`` (*crypto.constants*) — optional: **Default:** ``crypto.constants.ENGINE_METHOD_ALL``

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.setEngine('value', 'r');
   console.log(result);


setFips
~~~~~~~

.. code-block:: text

   crypto.setFips(bool)

*Kind:* function · *Arguments:* 1

Changes FIPS mode. With OpenSSL 3, this only adds or removes ``fips=yes`` in the default property query. It does not install, load, initialize, or validate a FIPS provider. For a usable FIPS configuration, install the provider and configure OpenSSL to load it when Node.js starts, as described in FIPS mode.

If no loaded provider supplies a requested cryptographic implementation matching ``fips=yes``, the call can still succeed and ``crypto.getFips()`` can still return ``1``, but fetching that implementation fails. Affected ``node:crypto`` operations typically fail with ``ERR_OSSL_EVP_UNSUPPORTED``. Operations that do not require a new fetch, including those using previously fetched implementations or initialized operation contexts, may still succeed. Call this method during application initialization, before application code uses other OpenSSL-backed APIs.

**Arguments**

* ``bool`` (*boolean*): ``true`` to enable FIPS mode, ``false`` to disable it.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.setFips(true);
   console.log(result);


Sign
~~~~

.. code-block:: text

   Sign

*Kind:* class · *Arguments:* 0

The ``Sign`` class is a utility for generating signatures. It can be used in one of two ways:

* As a writable stream, where data to be signed is written and the ``sign.sign()`` method is used to generate and return the signature, or * Using the ``sign.update()`` and ``sign.sign()`` methods to produce the signature.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const {
     generateKeyPairSync,
     createSign,
     createVerify,
   } = require('node:crypto');

   const { privateKey, publicKey } = generateKeyPairSync('ec', {
     namedCurve: 'sect239k1',
   });

   const sign = createSign('SHA256');
   sign.write('some data to sign');
   sign.end();
   const signature = sign.sign(privateKey, 'hex');

   const verify = createVerify('SHA256');
   verify.write('some data to sign');
   verify.end();
   console.log(verify.verify(publicKey, signature, 'hex'));
   // Prints: true


sign
~~~~

.. code-block:: text

   crypto.sign(algorithm, data, key[, callback])

*Kind:* function · *Arguments:* 4

Calculates and returns the signature for ``data`` using the given private key and algorithm. If ``algorithm`` is ``null`` or ``undefined``, then the algorithm is dependent upon the key type.

``algorithm`` is required to be ``null`` or ``undefined`` for Ed25519, Ed448, and ML-DSA.

**Arguments**

* ``algorithm`` (*string | null | undefined*)
* ``data`` (*ArrayBuffer|Buffer|SharedArrayBuffer|TypedArray|DataView|string*)
* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey|URL*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.sign('value', 'hello', 'example', () => {});
   console.log(result);


timingSafeEqual
~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.timingSafeEqual(a, b)

*Kind:* function · *Arguments:* 2

This function compares the underlying bytes that represent the given ``ArrayBuffer``, ``TypedArray``, or ``DataView`` instances using a constant-time algorithm.

This function does not leak timing information that would allow an attacker to guess one of the values. This is suitable for comparing HMAC digests or secret values like authentication cookies or \`capability urls <https://www.w3.org/TR/capability-urls/>\`_.

**Arguments**

* ``a`` (*ArrayBuffer|Buffer|TypedArray|DataView*)
* ``b`` (*ArrayBuffer|Buffer|TypedArray|DataView*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.timingSafeEqual([], []);
   console.log(result);


Verify
~~~~~~

.. code-block:: text

   Verify

*Kind:* class · *Arguments:* 0

The ``Verify`` class is a utility for verifying signatures. It can be used in one of two ways:

* As a writable stream where written data is used to validate against the supplied signature, or * Using the ``verify.update()`` and ``verify.verify()`` methods to verify the signature.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Verify();
   console.log(instance);


verify
~~~~~~

.. code-block:: text

   crypto.verify(algorithm, data, key, signature[, callback])

*Kind:* function · *Arguments:* 5

Verifies the given signature for ``data`` using the given key and algorithm. If ``algorithm`` is ``null`` or ``undefined``, then the algorithm is dependent upon the key type.

``algorithm`` is required to be ``null`` or ``undefined`` for Ed25519, Ed448, and ML-DSA.

**Arguments**

* ``algorithm`` (*string|null|undefined*)
* ``data`` (*ArrayBuffer|Buffer|SharedArrayBuffer|TypedArray|DataView|string*)
* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``signature`` (*ArrayBuffer|Buffer|SharedArrayBuffer|TypedArray|DataView*)
* ``callback`` (*Function*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.verify('value', 'hello', 'example', [], () => {});
   console.log(result);


X509Certificate
~~~~~~~~~~~~~~~

.. code-block:: text

   X509Certificate

*Kind:* class · *Arguments:* 1

Encapsulates an X509 certificate and provides read-only access to its information.

This is a class; construct instances with ``new``.

**Example**

.. code-block:: javascript

   const { X509Certificate } = require('node:crypto');

   const x509 = new X509Certificate('{... pem encoded cert ...}');

   console.log(x509.subject);


Certificate class
-----------------

**Certificate** — Instances of the ``Certificate`` class can be created using the ``new`` keyword or by calling ``crypto.Certificate()`` as a function:

Certificate.exportChallenge
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Certificate.exportChallenge(spkac[, encoding])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``spkac`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``spkac`` string.

**Example**

.. code-block:: javascript

   const { Certificate } = require('node:crypto');
   const spkac = getSpkacSomehow();
   const challenge = Certificate.exportChallenge(spkac);
   console.log(challenge.toString('utf8'));
   // Prints: the challenge as a UTF8 string


Certificate.exportPublicKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Certificate.exportPublicKey(spkac[, encoding])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``spkac`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``spkac`` string.

**Example**

.. code-block:: javascript

   const { Certificate } = require('node:crypto');
   const spkac = getSpkacSomehow();
   const publicKey = Certificate.exportPublicKey(spkac);
   console.log(publicKey);
   // Prints: the public key as <Buffer ...>


Certificate.verifySpkac
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   Certificate.verifySpkac(spkac[, encoding])

*Kind:* function · *Arguments:* 2

**Arguments**

* ``spkac`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``spkac`` string.

**Example**

.. code-block:: javascript

   const { Buffer } = require('node:buffer');
   const { Certificate } = require('node:crypto');

   const spkac = getSpkacSomehow();
   console.log(Certificate.verifySpkac(Buffer.from(spkac)));
   // Prints: true or false


Cipheriv class
--------------

**Cipheriv** — Instances of the ``Cipheriv`` class are used to encrypt data. The class can be used in one of two ways:

* As a stream that is both readable and writable, where plain unencrypted data is written to produce encrypted data on the readable side, or * Using the ``cipher.update()`` and ``cipher.final()`` methods to produce the encrypted data.

Cipheriv.final
~~~~~~~~~~~~~~

.. code-block:: text

   cipher.final([outputEncoding])

*Kind:* function · *Arguments:* 1

If an output encoding was specified in a previous call to ``cipher.update()``, ``outputEncoding`` must use the same encoding.

Once the ``cipher.final()`` method has been called, the ``Cipheriv`` object can no longer be used to encrypt data. Attempts to call ``cipher.final()`` more than once will result in an error being thrown.

**Arguments**

* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Cipheriv();
   const result = instance.final('utf8');
   console.log(result);


Cipheriv.getAuthTag
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.getAuthTag()

*Kind:* function · *Arguments:* 0

The ``cipher.getAuthTag()`` method should only be called after encryption has been completed using the ``cipher.final()`` method.

If the ``authTagLength`` option was set during the ``cipher`` instance's creation, this function will return exactly ``authTagLength`` bytes.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Cipheriv();
   const result = instance.getAuthTag();
   console.log(result);


Cipheriv.setAAD
~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.setAAD(buffer[, options])

*Kind:* function · *Arguments:* 2

When using an authenticated encryption mode (``GCM``, ``CCM``, ``OCB``, ``SIV``, ``GCM-SIV``, and ``chacha20-poly1305`` are currently supported), the ``cipher.setAAD()`` method sets the value used for the \_additional authenticated data\_ (AAD) input parameter.

The ``plaintextLength`` option is optional for ``GCM``, ``OCB``, ``SIV``, and ``GCM-SIV``. When using ``CCM``, the ``plaintextLength`` option must be specified and its value must match the length of the plaintext in bytes. See CCM mode.

**Arguments**

* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Cipheriv();
   const result = instance.setAAD([], {});
   console.log(result);


Cipheriv.setAutoPadding
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.setAutoPadding([autoPadding])

*Kind:* function · *Arguments:* 1

When using block ciphers that use standard block padding, the ``Cipheriv`` class will automatically add padding to the input data to the appropriate block size. To disable the default padding call ``cipher.setAutoPadding(false)``.

For block ciphers that use standard block padding, when ``autoPadding`` is ``false``, the length of the entire input data must be a multiple of the cipher's block size or ``cipher.final()`` will throw an error. Disabling automatic padding is useful for non-standard padding, for instance using ``0x0`` instead of PKCS padding.

**Arguments**

* ``autoPadding`` (*boolean*) — optional: **Default:** ``true``

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Cipheriv();
   const result = instance.setAutoPadding(true);
   console.log(result);


Cipheriv.update
~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.update(data[, inputEncoding][, outputEncoding])

*Kind:* function · *Arguments:* 3

Updates the cipher with ``data``. If the ``inputEncoding`` argument is given, the ``data`` argument is a string using the specified encoding. If the ``inputEncoding`` argument is not given, ``data`` must be a ``Buffer``, ``TypedArray``, or ``DataView``. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

The ``outputEncoding`` specifies the output format of the enciphered data. If the ``outputEncoding`` is specified, a string using the specified encoding is returned. If no ``outputEncoding`` is provided, a ``Buffer`` is returned. When ``outputEncoding`` is specified, it must use the same encoding as previous calls to ``cipher.update()``.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the data.
* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Cipheriv();
   const result = instance.update('hello', 'utf8', 'utf8');
   console.log(result);


Decipheriv class
----------------

**Decipheriv** — Instances of the ``Decipheriv`` class are used to decrypt data. The class can be used in one of two ways:

* As a stream that is both readable and writable, where plain encrypted data is written to produce unencrypted data on the readable side, or * Using the ``decipher.update()`` and ``decipher.final()`` methods to produce the unencrypted data.

Decipheriv.final
~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.final([outputEncoding])

*Kind:* function · *Arguments:* 1

If an output encoding was specified in a previous call to ``cipher.update()``, ``outputEncoding`` must use the same encoding.

Once the ``cipher.final()`` method has been called, the ``Cipheriv`` object can no longer be used to encrypt data. Attempts to call ``cipher.final()`` more than once will result in an error being thrown.

**Arguments**

* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Decipheriv();
   const result = instance.final('utf8');
   console.log(result);


Decipheriv.setAAD
~~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.setAAD(buffer[, options])

*Kind:* function · *Arguments:* 2

When using an authenticated encryption mode (``GCM``, ``CCM``, ``OCB``, ``SIV``, ``GCM-SIV``, and ``chacha20-poly1305`` are currently supported), the ``cipher.setAAD()`` method sets the value used for the \_additional authenticated data\_ (AAD) input parameter.

The ``plaintextLength`` option is optional for ``GCM``, ``OCB``, ``SIV``, and ``GCM-SIV``. When using ``CCM``, the ``plaintextLength`` option must be specified and its value must match the length of the plaintext in bytes. See CCM mode.

**Arguments**

* ``buffer`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Decipheriv();
   const result = instance.setAAD([], {});
   console.log(result);


Decipheriv.setAuthTag
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   decipher.setAuthTag(buffer[, encoding])

*Kind:* function · *Arguments:* 2

When using an authenticated encryption mode (``GCM``, ``CCM``, ``OCB``, ``SIV``, ``GCM-SIV``, and ``chacha20-poly1305`` are currently supported), the ``decipher.setAuthTag()`` method is used to pass in the received *authentication tag*. If no tag is provided, or if the cipher text has been tampered with, ``decipher.final()`` will throw, indicating that the cipher text should be discarded due to failed authentication. If the tag length is invalid according to NIST SP 800-38D or does not match the value of the ``authTagLength`` option, ``decipher.setAuthTag()`` will throw an error.

The ``decipher.setAuthTag()`` method must be called before ``decipher.update()`` for ``CCM``, ``SIV``, and ``GCM-SIV`` modes or before ``decipher.final()`` for ``GCM`` and ``OCB`` modes and ``chacha20-poly1305``. ``decipher.setAuthTag()`` can only be called once.

**Arguments**

* ``buffer`` (*string|Buffer|ArrayBuffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: String encoding to use when ``buffer`` is a string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Decipheriv();
   const result = instance.setAuthTag([], 'utf8');
   console.log(result);


Decipheriv.setAutoPadding
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.setAutoPadding([autoPadding])

*Kind:* function · *Arguments:* 1

When using block ciphers that use standard block padding, the ``Cipheriv`` class will automatically add padding to the input data to the appropriate block size. To disable the default padding call ``cipher.setAutoPadding(false)``.

For block ciphers that use standard block padding, when ``autoPadding`` is ``false``, the length of the entire input data must be a multiple of the cipher's block size or ``cipher.final()`` will throw an error. Disabling automatic padding is useful for non-standard padding, for instance using ``0x0`` instead of PKCS padding.

**Arguments**

* ``autoPadding`` (*boolean*) — optional: **Default:** ``true``

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Decipheriv();
   const result = instance.setAutoPadding(true);
   console.log(result);


Decipheriv.update
~~~~~~~~~~~~~~~~~

.. code-block:: text

   cipher.update(data[, inputEncoding][, outputEncoding])

*Kind:* function · *Arguments:* 3

Updates the cipher with ``data``. If the ``inputEncoding`` argument is given, the ``data`` argument is a string using the specified encoding. If the ``inputEncoding`` argument is not given, ``data`` must be a ``Buffer``, ``TypedArray``, or ``DataView``. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

The ``outputEncoding`` specifies the output format of the enciphered data. If the ``outputEncoding`` is specified, a string using the specified encoding is returned. If no ``outputEncoding`` is provided, a ``Buffer`` is returned. When ``outputEncoding`` is specified, it must use the same encoding as previous calls to ``cipher.update()``.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the data.
* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Decipheriv();
   const result = instance.update('hello', 'utf8', 'utf8');
   console.log(result);


DiffieHellman class
-------------------

**DiffieHellman** — The ``DiffieHellman`` class is a utility for creating Diffie-Hellman key exchanges.

Instances of the ``DiffieHellman`` class can be created using the ``crypto.createDiffieHellman()`` function.

DiffieHellman.$godeComputeSecret
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeComputeSecret(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeComputeSecret();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeGenerateKeys
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeGenerateKeys(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeGenerateKeys();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeGetGenerator
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeGetGenerator(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeGetGenerator();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeGetPrime
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeGetPrime(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeGetPrime();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeGetPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeGetPrivateKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeGetPrivateKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeGetPublicKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeGetPublicKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeGetPublicKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeSetPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeSetPrivateKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeSetPrivateKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeSetPublicKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeSetPublicKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeSetPublicKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.$godeVerifyError
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.DiffieHellman.$godeVerifyError(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.$godeVerifyError();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


DiffieHellman.computeSecret
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.computeSecret(otherPublicKey[, inputEncoding][, outputEncoding])

*Kind:* function · *Arguments:* 3

Computes the shared secret using ``otherPublicKey`` as the other party's public key and returns the computed shared secret. The supplied key is interpreted using the specified ``inputEncoding``, and secret is encoded using specified ``outputEncoding``. If the ``inputEncoding`` is not provided, ``otherPublicKey`` is expected to be a ``Buffer``, ``TypedArray``, or ``DataView``.

If ``outputEncoding`` is given a string is returned; otherwise, a ``Buffer`` is returned.

**Arguments**

* ``otherPublicKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of an ``otherPublicKey`` string.
* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.computeSecret([], 'utf8', 'utf8');
   console.log(result);


DiffieHellman.generateKeys
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.generateKeys([encoding])

*Kind:* function · *Arguments:* 1

Generates private and public Diffie-Hellman key values unless they have been generated or computed already, and returns the public key in the specified ``encoding``. This key should be transferred to the other party. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned.

This function is a thin wrapper around ``DH_generate_key()``. In particular, once a private key has been generated or set, calling this function only recomputes the public key from the existing private key. Since the public key is determined by the private key, the result will be the same unless the private key has been changed via ``diffieHellman.setPrivateKey()``.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.generateKeys('utf8');
   console.log(result);


DiffieHellman.getGenerator
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.getGenerator([encoding])

*Kind:* function · *Arguments:* 1

Returns the Diffie-Hellman generator in the specified ``encoding``. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.getGenerator('utf8');
   console.log(result);


DiffieHellman.getPrime
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.getPrime([encoding])

*Kind:* function · *Arguments:* 1

Returns the Diffie-Hellman prime in the specified ``encoding``. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.getPrime('utf8');
   console.log(result);


DiffieHellman.getPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.getPrivateKey([encoding])

*Kind:* function · *Arguments:* 1

Returns the Diffie-Hellman private key in the specified ``encoding``. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.getPrivateKey('utf8');
   console.log(result);


DiffieHellman.getPublicKey
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.getPublicKey([encoding])

*Kind:* function · *Arguments:* 1

Returns the Diffie-Hellman public key in the specified ``encoding``. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.getPublicKey('utf8');
   console.log(result);


DiffieHellman.setPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.setPrivateKey(privateKey[, encoding])

*Kind:* function · *Arguments:* 2

Sets the Diffie-Hellman private key. If the ``encoding`` argument is provided, ``privateKey`` is expected to be a string. If no ``encoding`` is provided, ``privateKey`` is expected to be a ``Buffer``, ``TypedArray``, or ``DataView``.

This function does not automatically compute the associated public key. Either ``diffieHellman.setPublicKey()`` or ``diffieHellman.generateKeys()`` can be used to manually provide the public key or to automatically derive it.

**Arguments**

* ``privateKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``privateKey`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.setPrivateKey([], 'utf8');
   console.log(result);


DiffieHellman.setPublicKey
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   diffieHellman.setPublicKey(publicKey[, encoding])

*Kind:* function · *Arguments:* 2

Sets the Diffie-Hellman public key. If the ``encoding`` argument is provided, ``publicKey`` is expected to be a string. If no ``encoding`` is provided, ``publicKey`` is expected to be a ``Buffer``, ``TypedArray``, or ``DataView``.

**Arguments**

* ``publicKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``publicKey`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.DiffieHellman();
   const result = instance.setPublicKey([], 'utf8');
   console.log(result);


ECDH class
----------

**ECDH** — The ``ECDH`` class is a utility for creating Elliptic Curve Diffie-Hellman (ECDH) key exchanges.

Instances of the ``ECDH`` class can be created using the ``crypto.createECDH()`` function.

ECDH.$godeComputeSecret
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeComputeSecret(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeComputeSecret();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.$godeGenerateKeys
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeGenerateKeys(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeGenerateKeys();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.$godeGetPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeGetPrivateKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeGetPrivateKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.$godeGetPublicKey
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeGetPublicKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeGetPublicKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.$godeSetPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeSetPrivateKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeSetPrivateKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.$godeSetPublicKey
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.ECDH.$godeSetPublicKey(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.$godeSetPublicKey();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


ECDH.computeSecret
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.computeSecret(otherPublicKey[, inputEncoding][, outputEncoding])

*Kind:* function · *Arguments:* 3

Computes the shared secret using ``otherPublicKey`` as the other party's public key and returns the computed shared secret. The supplied key is interpreted using specified ``inputEncoding``, and the returned secret is encoded using the specified ``outputEncoding``. If the ``inputEncoding`` is not provided, ``otherPublicKey`` is expected to be a ``Buffer``, ``TypedArray``, or ``DataView``.

If ``outputEncoding`` is given a string will be returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``otherPublicKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``otherPublicKey`` string.
* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.computeSecret([], 'utf8', 'utf8');
   console.log(result);


ECDH.generateKeys
~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.generateKeys([encoding[, format]])

*Kind:* function · *Arguments:* 2

Generates private and public EC Diffie-Hellman key values, and returns the public key in the specified ``format`` and ``encoding``. This key should be transferred to the other party.

The ``format`` argument specifies point encoding and can be ``'compressed'`` or ``'uncompressed'``. If ``format`` is not specified, the point will be returned in ``'uncompressed'`` format.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.
* ``format`` (*string*) — optional: **Default:** ``'uncompressed'``

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.generateKeys('utf8', 'value');
   console.log(result);


ECDH.getPrivateKey
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.getPrivateKey([encoding])

*Kind:* function · *Arguments:* 1

If ``encoding`` is specified, a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.getPrivateKey('utf8');
   console.log(result);


ECDH.getPublicKey
~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.getPublicKey([encoding][, format])

*Kind:* function · *Arguments:* 2

The ``format`` argument specifies point encoding and can be ``'compressed'`` or ``'uncompressed'``. If ``format`` is not specified the point will be returned in ``'uncompressed'`` format.

If ``encoding`` is specified, a string is returned; otherwise a ``Buffer`` is returned.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.
* ``format`` (*string*) — optional: **Default:** ``'uncompressed'``

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.getPublicKey('utf8', 'value');
   console.log(result);


ECDH.setPrivateKey
~~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.setPrivateKey(privateKey[, encoding])

*Kind:* function · *Arguments:* 2

Sets the EC Diffie-Hellman private key. If ``encoding`` is provided, ``privateKey`` is expected to be a string; otherwise ``privateKey`` is expected to be a ``Buffer``, ``TypedArray``, or ``DataView``.

If ``privateKey`` is not valid for the curve specified when the ``ECDH`` object was created, an error is thrown. Upon setting the private key, the associated public point (key) is also generated and set in the ``ECDH`` object.

**Arguments**

* ``privateKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``privateKey`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.ECDH();
   const result = instance.setPrivateKey([], 'utf8');
   console.log(result);


ECDH.setPublicKey
~~~~~~~~~~~~~~~~~

.. code-block:: text

   ecdh.setPublicKey(publicKey[, encoding])

*Kind:* function · *Arguments:* 2

Sets the EC Diffie-Hellman public key. If ``encoding`` is provided ``publicKey`` is expected to be a string; otherwise a ``Buffer``, ``TypedArray``, or ``DataView`` is expected.

There is not normally a reason to call this method because ``ECDH`` only requires a private key and the other party's public key to compute the shared secret. Typically either ``ecdh.generateKeys()`` or ``ecdh.setPrivateKey()`` will be called. The ``ecdh.setPrivateKey()`` method attempts to generate the public point/key associated with the private key being set.

**Arguments**

* ``publicKey`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``encoding`` (*string*) — optional: The encoding of the ``publicKey`` string.

**Example**

.. code-block:: javascript

   const {
     createECDH,
     createHash,
   } = require('node:crypto');

   const alice = createECDH('secp256k1');
   const bob = createECDH('secp256k1');

   // This is a shortcut way of specifying one of Alice's previous private
   // keys. It would be unwise to use such a predictable private key in a real
   // application.
   alice.setPrivateKey(
     createHash('sha256').update('alice', 'utf8').digest(),
   );

   // Bob uses a newly generated cryptographically strong
   // pseudorandom key pair
   bob.generateKeys();

   const aliceSecret = alice.computeSecret(bob.getPublicKey(), null, 'hex');
   const bobSecret = bob.computeSecret(alice.getPublicKey(), null, 'hex');

   // aliceSecret and bobSecret should be the same shared secret value
   console.log(aliceSecret === bobSecret);


Hash class
----------

**Hash** — The ``Hash`` class is a utility for creating hash digests of data. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written to produce a computed hash digest on the readable side, or * Using the ``hash.update()`` and ``hash.digest()`` methods to produce the computed hash.

Hash.copy
~~~~~~~~~

.. code-block:: text

   hash.copy([options])

*Kind:* function · *Arguments:* 1

Creates a new ``Hash`` object that contains a deep copy of the internal state of the current ``Hash`` object.

The optional ``options`` argument controls stream behavior. For XOF hash functions such as ``'shake256'``, the ``outputLength`` option can be used to specify the desired output length in bytes.

**Arguments**

* ``options`` (*Object*) — optional: ``stream.transform`` options

**Example**

.. code-block:: javascript

   // Calculate a rolling hash.
   const {
     createHash,
   } = require('node:crypto');

   const hash = createHash('sha256');

   hash.update('one');
   console.log(hash.copy().digest('hex'));

   hash.update('two');
   console.log(hash.copy().digest('hex'));

   hash.update('three');
   console.log(hash.copy().digest('hex'));

   // Etc.


Hash.digest
~~~~~~~~~~~

.. code-block:: text

   hash.digest([encoding])

*Kind:* function · *Arguments:* 1

Calculates the digest of all of the data passed to be hashed (using the ``hash.update()`` method). If ``encoding`` is provided a string will be returned; otherwise a ``Buffer`` is returned.

The ``Hash`` object can not be used again after ``hash.digest()`` method has been called. Multiple calls will cause an error to be thrown.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hash();
   const result = instance.digest('utf8');
   console.log(result);


Hash.end
~~~~~~~~

.. code-block:: text

   crypto.Hash.end(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hash();
   const result = instance.end();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Hash.read
~~~~~~~~~

.. code-block:: text

   crypto.Hash.read(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hash();
   const result = instance.read();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Hash.update
~~~~~~~~~~~

.. code-block:: text

   hash.update(data[, inputEncoding])

*Kind:* function · *Arguments:* 2

Updates the hash content with the given ``data``, the encoding of which is given in ``inputEncoding``. If ``encoding`` is not provided, and the ``data`` is a string, an encoding of ``'utf8'`` is enforced. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

This can be called many times with new data as it is streamed.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``data`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hash();
   const result = instance.update('hello', 'utf8');
   console.log(result);


Hash.write
~~~~~~~~~~

.. code-block:: text

   crypto.Hash.write(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hash();
   const result = instance.write();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


Hmac class
----------

**Hmac** — The ``Hmac`` class is a utility for creating cryptographic HMAC digests. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written to produce a computed HMAC digest on the readable side, or * Using the ``hmac.update()`` and ``hmac.digest()`` methods to produce the computed HMAC digest.

Hmac.digest
~~~~~~~~~~~

.. code-block:: text

   hmac.digest([encoding])

*Kind:* function · *Arguments:* 1

Calculates the HMAC digest of all of the data passed using ``hmac.update()``. If ``encoding`` is provided a string is returned; otherwise a ``Buffer`` is returned;

The ``Hmac`` object can not be used again after ``hmac.digest()`` has been called. Multiple calls to ``hmac.digest()`` will result in an error being thrown.

**Arguments**

* ``encoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hmac();
   const result = instance.digest('utf8');
   console.log(result);


Hmac.update
~~~~~~~~~~~

.. code-block:: text

   hmac.update(data[, inputEncoding])

*Kind:* function · *Arguments:* 2

Updates the ``Hmac`` content with the given ``data``, the encoding of which is given in ``inputEncoding``. If ``encoding`` is not provided, and the ``data`` is a string, an encoding of ``'utf8'`` is enforced. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

This can be called many times with new data as it is streamed.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``data`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Hmac();
   const result = instance.update('hello', 'utf8');
   console.log(result);


KeyObject class
---------------

**KeyObject** — Node.js uses a ``KeyObject`` class to represent a symmetric or asymmetric key, and each kind of key exposes different functions. The ``crypto.createSecretKey()``, ``crypto.createPublicKey()`` and ``crypto.createPrivateKey()`` methods are used to create ``KeyObject`` instances. ``KeyObject`` objects are not to be created directly using the ``new`` keyword.

Most applications should consider using the new ``KeyObject`` API instead of passing keys as strings or ``Buffer``\ s due to improved security features.

KeyObject.$godeAsymType
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeAsymType(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeAsymType();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeAttachId
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeAttachId(arg0)

*Kind:* function · *Arguments:* 1

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeAttachId(undefined);
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeDetails
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeDetails(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeDetails();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeEquals
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeEquals(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeEquals();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeExport
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeExport(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeExport();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeOwnId
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeOwnId(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeOwnId();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeSymSize
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeSymSize(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeSymSize();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.$godeType
~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.KeyObject.$godeType(...args)

*Kind:* function · *Arguments:* 0

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.$godeType();
   console.log(result);

.. note::

   This entry comes from the runtime surface. A fuller description is not available yet, and argument names may be incomplete.


KeyObject.equals
~~~~~~~~~~~~~~~~

.. code-block:: text

   keyObject.equals(otherKeyObject)

*Kind:* function · *Arguments:* 1

Returns ``true`` or ``false`` depending on whether the keys have exactly the same type, value, and parameters. This method is not \`constant time <https://en.wikipedia.org/wiki/Timing_attack>\`_.

**Arguments**

* ``otherKeyObject`` (*KeyObject*): A ``KeyObject`` with which to

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.equals(undefined);
   console.log(result);


KeyObject.export
~~~~~~~~~~~~~~~~

.. code-block:: text

   keyObject.export([options])

*Kind:* function · *Arguments:* 1

For symmetric keys, the following encoding options can be used:

For public keys, the following encoding options can be used:

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.export({});
   console.log(result);


KeyObject.toCryptoKey
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   keyObject.toCryptoKey(algorithm, extractable, keyUsages)

*Kind:* function · *Arguments:* 3

Converts a ``KeyObject`` instance to a ``CryptoKey``.

**Arguments**

* ``algorithm`` (*string|Algorithm|RsaHashedImportParams|EcKeyImportParams|HmacImportParams*)
* ``extractable`` (*boolean*)
* ``keyUsages`` (*string\[]*): See Key usages.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.KeyObject();
   const result = instance.toCryptoKey('value', true, []);
   console.log(result);


Mac class
---------

**Mac** — The ``Mac`` class computes message authentication codes using MAC implementations supplied by OpenSSL providers. It can be used in one of two ways:

* As a stream that is both readable and writable, where data is written and one authentication tag is produced on the readable side when the writable side ends; or * By calling ``mac.update()`` one or more times followed by ``mac.final()``.

Mac.final
~~~~~~~~~

.. code-block:: text

   mac.final([outputEncoding])

*Kind:* function · *Arguments:* 1

Completes the MAC computation and returns the authentication tag. If ``outputEncoding`` is omitted or is ``'buffer'``, a ``Buffer`` is returned. Otherwise, a string is returned.

To verify an authentication tag, compare equal-length ``Buffer`` values using ``crypto.timingSafeEqual()``.

**Arguments**

* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Mac();
   const result = instance.final('utf8');
   console.log(result);


Mac.update
~~~~~~~~~~

.. code-block:: text

   mac.update(data[, inputEncoding])

*Kind:* function · *Arguments:* 2

Updates the MAC with ``data`` and returns the ``Mac`` object so that calls can be chained. When ``data`` is a string, ``inputEncoding`` defaults to ``'utf8'``. When ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, ``inputEncoding`` is ignored.

This method can be called multiple times before finalization. If an underlying MAC update fails, the ``Mac`` object cannot be used again. Calling this method after a previous underlying MAC update failure or after finalization throws ``ERR_CRYPTO_MAC_FINALIZED``.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``data`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Mac();
   const result = instance.update('hello', 'utf8');
   console.log(result);


Sign class
----------

**Sign** — The ``Sign`` class is a utility for generating signatures. It can be used in one of two ways:

* As a writable stream, where data to be signed is written and the ``sign.sign()`` method is used to generate and return the signature, or * Using the ``sign.update()`` and ``sign.sign()`` methods to produce the signature.

Sign.sign
~~~~~~~~~

.. code-block:: text

   sign.sign(privateKey[, outputEncoding])

*Kind:* function · *Arguments:* 2

Calculates the signature on all the data passed through using either ``sign.update()`` or ``sign.write()``.

If ``privateKey`` is not a ``KeyObject``, this function behaves as if ``privateKey`` had been passed to ``crypto.createPrivateKey()``. When ``privateKey`` is a string, ``ArrayBuffer``, ``Buffer``, ``TypedArray``, or ``DataView``, it must contain PEM-encoded key material. If it is an object, the following additional properties can be passed:

**Arguments**

* ``privateKey`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey|URL*)
* ``outputEncoding`` (*string*) — optional: The encoding of the return value.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Sign();
   const result = instance.sign([], 'utf8');
   console.log(result);


Sign.update
~~~~~~~~~~~

.. code-block:: text

   sign.update(data[, inputEncoding])

*Kind:* function · *Arguments:* 2

Updates the ``Sign`` content with the given ``data``, the encoding of which is given in ``inputEncoding``. If ``encoding`` is not provided, and the ``data`` is a string, an encoding of ``'utf8'`` is enforced. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

This can be called many times with new data as it is streamed.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``data`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Sign();
   const result = instance.update('hello', 'utf8');
   console.log(result);


Verify class
------------

**Verify** — The ``Verify`` class is a utility for verifying signatures. It can be used in one of two ways:

* As a writable stream where written data is used to validate against the supplied signature, or * Using the ``verify.update()`` and ``verify.verify()`` methods to verify the signature.

Verify.update
~~~~~~~~~~~~~

.. code-block:: text

   verify.update(data[, inputEncoding])

*Kind:* function · *Arguments:* 2

Updates the ``Verify`` content with the given ``data``, the encoding of which is given in ``inputEncoding``. If ``inputEncoding`` is not provided, and the ``data`` is a string, an encoding of ``'utf8'`` is enforced. If ``data`` is a ``Buffer``, ``TypedArray``, or ``DataView``, then ``inputEncoding`` is ignored.

This can be called many times with new data as it is streamed.

**Arguments**

* ``data`` (*string|Buffer|TypedArray|DataView*)
* ``inputEncoding`` (*string*) — optional: The encoding of the ``data`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Verify();
   const result = instance.update('hello', 'utf8');
   console.log(result);


Verify.verify
~~~~~~~~~~~~~

.. code-block:: text

   verify.verify(key, signature[, signatureEncoding])

*Kind:* function · *Arguments:* 3

Verifies the provided data using the given ``key`` and ``signature``.

If ``key`` is not a ``KeyObject``, this function behaves as if ``key`` had been passed to ``crypto.createPublicKey()``. When ``key`` is a string, ``ArrayBuffer``, ``Buffer``, ``TypedArray``, or ``DataView``, it must contain PEM-encoded key material. If it is an object, the following additional properties can be passed:

**Arguments**

* ``key`` (*Object|string|ArrayBuffer|Buffer|TypedArray|DataView|KeyObject|CryptoKey*)
* ``signature`` (*string|ArrayBuffer|Buffer|TypedArray|DataView*)
* ``signatureEncoding`` (*string*) — optional: The encoding of the ``signature`` string.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.Verify();
   const result = instance.verify('example', [], 'utf8');
   console.log(result);


webcrypto
---------

**webcrypto** — Type: ``Crypto`` An implementation of the Web Crypto API standard.

See the Web Crypto API documentation for details.

webcrypto.getRandomValues
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.getRandomValues(typedArray)

*Kind:* function · *Arguments:* 1

A convenient alias for ``crypto.webcrypto.getRandomValues()``. This implementation is not compliant with the Web Crypto spec, to write web-compatible code use ``crypto.webcrypto.getRandomValues()`` instead.

**Arguments**

* ``typedArray`` (*Buffer|TypedArray|DataView|ArrayBuffer*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.getRandomValues([]);
   console.log(result);


webcrypto.randomUUID
~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   crypto.randomUUID([options])

*Kind:* function · *Arguments:* 1

Generates a random RFC 4122 version 4 UUID. The UUID is generated using a cryptographic pseudorandom number generator.

**Arguments**

* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.randomUUID({});
   console.log(result);


webcrypto.subtle.decrypt
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subtle.decrypt(algorithm, key, data)

*Kind:* function · *Arguments:* 3

Using the method and parameters specified in ``algorithm`` and the keying material provided by ``key``, this method attempts to decipher the provided ``data``. If successful, the returned promise will be resolved with an ``ArrayBuffer`` containing the plaintext result.

The algorithms currently supported include:

**Arguments**

* ``algorithm`` (*RsaOaepParams|AesCtrParams|AesCbcParams|AeadParams*)
* ``key`` (*CryptoKey*)
* ``data`` (*ArrayBuffer|TypedArray|DataView|Buffer*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.subtle.decrypt(undefined, 'example', 'hello');
   console.log(result);


webcrypto.subtle.digest
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subtle.digest(algorithm, data)

*Kind:* function · *Arguments:* 2

Using the method identified by ``algorithm``, this method attempts to generate a digest of ``data``. If successful, the returned promise is resolved with an ``ArrayBuffer`` containing the computed digest.

If ``algorithm`` is provided as a ``string``, it must be one of:

**Arguments**

* ``algorithm`` (*string|Algorithm|CShakeParams|TurboShakeParams|KangarooTwelveParams*)
* ``data`` (*ArrayBuffer|TypedArray|DataView|Buffer*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.subtle.digest('value', 'hello');
   console.log(result);


webcrypto.subtle.encrypt
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subtle.encrypt(algorithm, key, data)

*Kind:* function · *Arguments:* 3

Using the method and parameters specified by ``algorithm`` and the keying material provided by ``key``, this method attempts to encipher ``data``. If successful, the returned promise is resolved with an ``ArrayBuffer`` containing the encrypted result.

The algorithms currently supported include:

**Arguments**

* ``algorithm`` (*RsaOaepParams|AesCtrParams|AesCbcParams|AeadParams*)
* ``key`` (*CryptoKey*)
* ``data`` (*ArrayBuffer|TypedArray|DataView|Buffer*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.subtle.encrypt(undefined, 'example', 'hello');
   console.log(result);


webcrypto.subtle.importKey
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   subtle.importKey(format, keyData, algorithm, extractable, keyUsages)

*Kind:* function · *Arguments:* 5

This method attempts to interpret the provided ``keyData`` as the given ``format`` to create a ``CryptoKey`` instance using the provided ``algorithm``, ``extractable``, and ``keyUsages`` arguments. If the import is successful, the returned promise will be resolved with a ``CryptoKey`` representation of the key material.

If importing KDF algorithm keys, ``extractable`` must be ``false``.

**Arguments**

* ``format`` (*string*): Must be one of ``'raw'``, ``'pkcs8'``, ``'spki'``, ``'jwk'``, ``'raw-secret'``,
* ``keyData`` (*ArrayBuffer|TypedArray|DataView|Buffer|Object*)
* ``algorithm`` (*string|Algorithm|RsaHashedImportParams|EcKeyImportParams|HmacImportParams|KmacImportParams*)
* ``extractable`` (*boolean*)
* ``keyUsages`` (*string\[]*): See Key usages.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const result = crypto.webcrypto.subtle.importKey('value', [], 'value', true, []);
   console.log(result);


X509Certificate class
---------------------

**X509Certificate** — Encapsulates an X509 certificate and provides read-only access to its information.

X509Certificate.checkEmail
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.checkEmail(email[, options])

*Kind:* function · *Arguments:* 2

Checks whether the certificate matches the given email address.

If the ``'subject'`` option is undefined or set to ``'default'``, the certificate subject is considered according to OpenSSL's default behavior.

**Arguments**

* ``email`` (*string*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.checkEmail('value', {});
   console.log(result);


X509Certificate.checkHost
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.checkHost(name[, options])

*Kind:* function · *Arguments:* 2

Checks whether the certificate matches the given host name.

If the certificate matches the given host name, the matching subject name is returned. The returned name might be an exact match (e.g., ``foo.example.com``) or it might contain wildcards (e.g., ``*.example.com``). Because host name comparisons are case-insensitive, the returned subject name might also differ from the given ``name`` in capitalization.

**Arguments**

* ``name`` (*string*)
* ``options`` (*Object*) — optional

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.checkHost('example', {});
   console.log(result);


X509Certificate.checkIP
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.checkIP(ip)

*Kind:* function · *Arguments:* 1

Checks whether the certificate matches the given IP address (IPv4 or IPv6).

Only RFC 5280 ``iPAddress`` subject alternative names are considered, and they must match the given ``ip`` address exactly. Other subject alternative names as well as the subject field of the certificate are ignored.

**Arguments**

* ``ip`` (*string*)

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.checkIP('value');
   console.log(result);


X509Certificate.checkIssued
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.checkIssued(otherCert)

*Kind:* function · *Arguments:* 1

Checks whether this certificate was potentially issued by the given ``otherCert`` by comparing the certificate metadata.

This is useful for pruning a list of possible issuer certificates which have been selected using a more rudimentary filtering routine, i.e. just based on subject and issuer names.

**Arguments**

* ``otherCert`` (*X509Certificate*)

**Example**

.. code-block:: javascript

   if (!x509.verify(otherCert.publicKey)) {
     throw new Error('otherCert did not issue x509');
   }


X509Certificate.checkPrivateKey
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.checkPrivateKey(privateKey)

*Kind:* function · *Arguments:* 1

Checks whether the public key for this certificate is consistent with the given private key.

**Arguments**

* ``privateKey`` (*KeyObject*): A private key.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.checkPrivateKey(undefined);
   console.log(result);


X509Certificate.toJSON
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.toJSON()

*Kind:* function · *Arguments:* 0

There is no standard JSON encoding for X509 certificates. The ``toJSON()`` method returns a string containing the PEM encoded certificate.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.toJSON();
   console.log(result);


X509Certificate.toLegacyObject
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.toLegacyObject()

*Kind:* function · *Arguments:* 0

Returns information about this certificate using the legacy certificate object encoding.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.toLegacyObject();
   console.log(result);


X509Certificate.toString
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.toString()

*Kind:* function · *Arguments:* 0

Returns the PEM-encoded certificate.

This callable takes no arguments.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.toString();
   console.log(result);


X509Certificate.verify
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: text

   x509.verify(publicKey)

*Kind:* function · *Arguments:* 1

Verifies that this certificate was signed by the given public key. Does not perform any other validation checks on the certificate.

**Arguments**

* ``publicKey`` (*KeyObject*): A public key.

**Example**

.. code-block:: javascript

   const crypto = require('node:crypto');

   const instance = new crypto.X509Certificate();
   const result = instance.verify(undefined);
   console.log(result);


Properties and constants
------------------------

* ``constants`` — object
* ``constants.DH_CHECK_P_NOT_PRIME`` — number
* ``constants.DH_CHECK_P_NOT_SAFE_PRIME`` — number
* ``constants.DH_NOT_SUITABLE_GENERATOR`` — number
* ``constants.DH_UNABLE_TO_CHECK_GENERATOR`` — number
* ``constants.ENGINE_METHOD_ALL`` — number
* ``constants.POINT_CONVERSION_COMPRESSED`` — number
* ``constants.POINT_CONVERSION_UNCOMPRESSED`` — number
* ``constants.RSA_NO_PADDING`` — number
* ``constants.RSA_PKCS1_OAEP_PADDING`` — number
* ``constants.RSA_PKCS1_PADDING`` — number
* ``constants.RSA_SSLV23_PADDING`` — number
* ``DiffieHellmanGroup.setPrivateKey`` — undefined
* ``DiffieHellmanGroup.setPublicKey`` — undefined
* ``fips`` — boolean
* ``subtle`` — object
* ``webcrypto`` — object
* ``webcrypto.subtle`` — object

.. note::

   This page is generated from the GodeJS runtime snapshot. Signatures and argument counts reflect the running runtime; private helpers and GodeJS extensions are labelled.
