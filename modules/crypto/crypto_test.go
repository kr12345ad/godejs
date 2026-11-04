package crypto

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	buf := jstest.Install(t, e, "buffer", buffer.Install)
	e.Global().Set("Buffer", buf.Get("Buffer"))
	jstest.Install(t, e, "crypto", Install)
	return e
}

func TestHashAlgorithms(t *testing.T) {
	e := setup(t)
	cases := []struct{ algo, hex string }{
		{"sha256", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"SHA-256", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"md5", "900150983cd24fb0d6963f7d28e17f72"},
		{"sha1", "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{"sha512", "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
	}
	for _, tc := range cases {
		e.Global().SetStr("__algo", tc.algo)
		got := jstest.Str(t, e, `crypto.createHash(__algo).update('abc').digest('hex')`)
		jstest.Equal(t, got, tc.hex)
	}
}

func TestHashChunksAndCopy(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`(function(){var h=crypto.createHash('sha256'); h.update('a'); h.update(new Uint8Array([98,99])); return h.digest('hex');})()`),
		"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
}

func TestHashCopy(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var h = crypto.createHash('sha256');
		h.update('abc');
		var c = h.copy();
		return JSON.stringify([c.digest('hex'), h.digest('hex'), c instanceof crypto.Hash]);
	})()`), `["ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",true]`)
}

func TestDigestTwice(t *testing.T) {
	e := setup(t)
	err := jstest.Throws(t, e, `(function(){
		var h = crypto.createHash('sha256');
		h.update('a');
		h.digest();
		h.digest();
	})()`)
	jstest.Equal(t, err.Message, "Digest already called")
	jstest.Equal(t, err.Name, "Error")
}

func TestCreateHashUnknownAlgorithm(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`(function(){try{crypto.createHash('no-such-algorithm')}catch(e){return e.code+'|'+e.message}})()`),
		"ERR_CRYPTO_INVALID_DIGEST|Digest method not supported")
}

func TestHmac(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e,
		`crypto.createHmac('sha256', 'key').update('The quick brown fox jumps over the lazy dog').digest('hex')`)
	jstest.Equal(t, got, "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8")

	jstest.Equal(t, jstest.Str(t, e,
		`(function(){var h=crypto.createHmac('sha256','key'); h.update('The quick brown fox jumps '); h.update('over the lazy dog'); return h.digest('hex');})()`),
		"f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8")
}

func TestHashOneShot(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `crypto.hash('sha256', 'abc', 'hex')`),
		"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	// Node's one-shot hash defaults to a hex string, not a Buffer.
	jstest.Equal(t, jstest.Str(t, e, `crypto.hash('sha256', 'abc').length`), "64")
}

func TestRandomHelpers(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `crypto.randomBytes(16).length`), "16")
	jstest.Equal(t, jstest.Str(t, e, `/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(crypto.randomUUID())`), "true")
	jstest.Equal(t, jstest.Str(t, e,
		`(function(){var b=new Uint8Array(8); var r=crypto.randomFillSync(b,2,4); return JSON.stringify([r===b,b.length]);})()`),
		`[true,8]`)
	jstest.Equal(t, jstest.Str(t, e,
		`(function(){var a=Array.from(crypto.randomBytes(16)); var b=Array.from(crypto.randomBytes(16)); return a.join()!==b.join();})()`),
		"true")
}

func TestTimingSafeEqual(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`crypto.timingSafeEqual(new Uint8Array([1,2,3]), new Uint8Array([1,2,3]))`), "true")
	jstest.Equal(t, jstest.Str(t, e,
		`crypto.timingSafeEqual(new Uint8Array([1,2,3]), new Uint8Array([1,2,4]))`), "false")
	jstest.Equal(t, jstest.Str(t, e,
		`(function(){try{crypto.timingSafeEqual(new Uint8Array([1]), new Uint8Array([1,2]))}catch(e){return e.code}})()`),
		"ERR_CRYPTO_TIMING_SAFE_EQUAL_LENGTH")
}

func TestGetRandomValuesAndSubtle(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var a = new Uint8Array(16);
		var r = crypto.getRandomValues(a);
		return JSON.stringify([r === a, a.length, Array.from(a).some(function(x){return x !== 0;})]);
	})()`), `[true,16,true]`)
	jstest.Equal(t, jstest.Str(t, e,
		`typeof crypto.subtle.digest('sha256', new Uint8Array([97,98,99])).then`), "function")
	jstest.Equal(t, jstest.Str(t, e, `crypto.webcrypto.subtle === crypto.subtle`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof crypto.constants.RSA_PKCS1_PADDING`), "number")
}

func TestImplementedTracked(t *testing.T) {
	data, err := os.ReadFile("../../parity/baseline/crypto.json")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatal(err)
	}
	tracked := make(map[string]bool, len(keys))
	for _, k := range keys {
		tracked[k] = true
	}
	impl := Implemented()
	if len(impl) == 0 {
		t.Fatal("no implemented keys")
	}
	for _, k := range impl {
		if !tracked[k] {
			t.Errorf("Implemented key %q is not in parity/baseline/crypto.json", k)
		}
	}
}

func TestRandomInt(t *testing.T) {
	e := setup(t)
	for i := 0; i < 32; i++ {
		v := jstest.Eval(t, e, `crypto.randomInt(5, 10)`)
		if n := v.Int(); n < 5 || n > 9 {
			t.Fatalf("randomInt(5, 10) = %d", n)
		}
	}
	if n := jstest.Eval(t, e, `crypto.randomInt(1)`).Int(); n != 0 {
		t.Fatalf("randomInt(1) = %d", n)
	}
}

func TestCipherGCMRoundTrip(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		function cat(a, b) {
			var out = new Uint8Array(a.length + b.length);
			out.set(a); out.set(b, a.length); return out;
		}
		var key = new Uint8Array(32); key.fill(1);
		var iv = new Uint8Array(12); iv.fill(2);
		var c = crypto.createCipheriv('aes-256-gcm', key, iv);
		var ct = cat(c.update(new Uint8Array([104, 101, 108, 108, 111])), c.final());
		var tag = c.getAuthTag();
		var d = crypto.createDecipheriv('aes-256-gcm', key, iv);
		d.setAuthTag(tag);
		var pt = cat(d.update(ct), d.final());
		return String.fromCharCode.apply(null, pt) + ':' + tag.length + ':' + ct.length;
	})()`)
	jstest.Equal(t, got, "hello:16:5")
}

func TestCipherCBCRoundTrip(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		function cat(a, b) {
			var out = new Uint8Array(a.length + b.length);
			out.set(a); out.set(b, a.length); return out;
		}
		var key = new Uint8Array(16); key.fill(9);
		var iv = new Uint8Array(16); iv.fill(8);
		var c = crypto.createCipheriv('aes-128-cbc', key, iv);
		var ct = cat(c.update('hello world'), c.final());
		var d = crypto.createDecipheriv('aes-128-cbc', key, iv);
		var pt = cat(d.update(ct), d.final());
		return String.fromCharCode.apply(null, pt) + ':' + ct.length;
	})()`)
	jstest.Equal(t, got, "hello world:16")
}

func TestCipherErrors(t *testing.T) {
	e := setup(t)
	err := jstest.Throws(t, e, `crypto.createCipheriv('aes-256-cbc', new Uint8Array(3), new Uint8Array(16))`)
	jstest.Equal(t, err.Code, "ERR_CRYPTO_INVALID_KEYLEN")
	err = jstest.Throws(t, e, `crypto.createCipheriv('no-such-cipher', new Uint8Array(16), new Uint8Array(16))`)
	jstest.Equal(t, err.Code, "ERR_CRYPTO_UNKNOWN_CIPHER")
}

func TestECKeyGenSignVerify(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('ec', { namedCurve: 'P-256' });
		var msg = new Uint8Array([97, 98, 99]);
		var sig = crypto.sign('sha256', msg, kp.privateKey);
		return [kp.publicKey.type, kp.publicKey.asymmetricKeyType,
		        kp.publicKey.asymmetricKeyDetails.namedCurve,
		        crypto.verify('sha256', msg, kp.publicKey, sig),
		        crypto.verify('sha256', new Uint8Array([97, 98]), kp.publicKey, sig)].join('|');
	})()`)
	jstest.Equal(t, got, "public|ec|prime256v1|true|false")
}

func TestEd25519ExportImportPEM(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('ed25519');
		var privPem = kp.privateKey.export({ type: 'pkcs8', format: 'pem' });
		var pubPem = kp.publicKey.export({ type: 'spki', format: 'pem' });
		var priv = crypto.createPrivateKey(privPem);
		var pub = crypto.createPublicKey(pubPem);
		var msg = new Uint8Array([109]);
		var sig = crypto.sign(null, msg, priv);
		return [privPem.indexOf('BEGIN PRIVATE KEY') > 0,
		        pubPem.indexOf('BEGIN PUBLIC KEY') > 0,
		        crypto.verify(null, msg, pub, sig),
		        priv.equals(crypto.createPrivateKey(privPem))].join('|');
	})()`)
	jstest.Equal(t, got, "true|true|true|true")
}

func TestRSASignVerifyAndJWK(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('rsa', { modulusLength: 1024 });
		var msg = new Uint8Array([1, 2, 3]);
		var sig = crypto.sign('sha256', msg, kp.privateKey);
		var jwk = kp.privateKey.export({ format: 'jwk' });
		var imported = crypto.createPrivateKey({ key: jwk, format: 'jwk' });
		return [kp.publicKey.asymmetricKeyDetails.modulusLength,
		        crypto.verify('sha256', msg, kp.publicKey, sig),
		        crypto.verify('sha256', msg, imported, sig),
		        jwk.kty, jwk['d'] !== undefined].join('|');
	})()`)
	jstest.Equal(t, got, "1024|true|true|RSA|true")
}

func TestCreateSignVerify(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('ec', { namedCurve: 'secp384r1' });
		var sig = crypto.createSign('sha384').update('payload').sign(kp.privateKey);
		return [sig.length > 0,
		        crypto.createVerify('sha384').update('payload').verify(kp.publicKey, sig),
		        crypto.createVerify('sha384').update('other').verify(kp.publicKey, sig)].join('|');
	})()`)
	jstest.Equal(t, got, "true|true|false")
}

func TestShakeAndHashStream(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `crypto.createHash('shake128').digest('hex')`),
		"7f9c2ba4e88f827d616045507605853e")
	jstest.Equal(t, jstest.Str(t, e, `crypto.createHash('shake128', { outputLength: 5 }).digest('hex')`),
		"7f9c2ba4e8")
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var h = crypto.createHash('sha512');
		h.end('Test123');
		return h.read().length;
	})()`), "64")
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var h = crypto.createHash('sha256');
		h.digest();
		try { h.update('x'); } catch (e) { return e.code; }
		return 'none';
	})()`), "ERR_CRYPTO_HASH_FINALIZED")
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var h = crypto.Hash('sha256');
		h.update('abc');
		return h.digest('hex');
	})()`), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
}

func TestRandomBytesValidation(t *testing.T) {
	e := setup(t)
	err := jstest.Throws(t, e, `crypto.randomBytes(-1)`)
	jstest.Equal(t, err.Code, "ERR_OUT_OF_RANGE")
	err = jstest.Throws(t, e, `crypto.randomBytes('x')`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
	jstest.Equal(t, jstest.Str(t, e, `crypto.randomBytes(2.9).length`), "2")
	jstest.Equal(t, jstest.Str(t, e,
		`Object.getOwnPropertyDescriptor(crypto, 'rng').enumerable`), "false")
}

func TestHKDFAndPBKDF2Vectors(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`crypto.pbkdf2Sync('password', 'salt', 1, 32, 'sha256').length`), "32")

	got := jstest.Str(t, e, `(function(){
		function hex(u8) {
			var s = '';
			for (var i = 0; i < u8.length; i++) s += ('0' + u8[i].toString(16)).slice(-2);
			return s;
		}
		var ikm = new Uint8Array(22); ikm.fill(11);
		var salt = new Uint8Array(13);
		for (var i = 0; i < 13; i++) salt[i] = i;
		var info = new Uint8Array(10);
		for (var j = 0; j < 10; j++) info[j] = 240 + j;
		return hex(crypto.hkdfSync('sha256', ikm, salt, info, 42));
	})()`)
	jstest.Equal(t, got, "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")
}

func TestGetCipherInfoAndHashes(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var info = crypto.getCipherInfo('aes-256-gcm');
		return [info.keyLength, info.ivLength, info.mode].join('|');
	})()`), "32|12|gcm")
	jstest.Equal(t, jstest.Str(t, e, `crypto.getHashes().indexOf('sha3-256') >= 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `crypto.getCiphers().indexOf('aes-128-cbc') >= 0`), "true")
}

func TestDiffieHellmanGroups(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var a = crypto.getDiffieHellman('modp2');
		var b = crypto.createDiffieHellmanGroup('modp2');
		var pubA = a.generateKeys();
		var pubB = b.generateKeys('hex');
		var s1 = a.computeSecret(pubB, 'hex', 'hex');
		var s2 = b.computeSecret(pubA, null, 'hex');
		return [s1 === s2, a.verifyError, s1.length,
		        a instanceof crypto.DiffieHellman,
		        a.getPrime('hex').length > 0].join('|');
	})()`)
	jstest.Equal(t, got, "true|0|256|true|true")

	err := jstest.Throws(t, e, `(function(){
		var dh = crypto.getDiffieHellman('modp2');
		dh.generateKeys();
		return dh.computeSecret(new Uint8Array([1]));
	})()`)
	jstest.Equal(t, err.Code, "ERR_CRYPTO_INVALID_KEYLEN")
	jstest.Equal(t, err.Message, "Supplied key is too small")
}

func TestECDHSharedSecret(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var a = crypto.createECDH('prime256v1');
		var b = crypto.createECDH('prime256v1');
		var pa = a.generateKeys();
		var pb = b.generateKeys();
		var s1 = a.computeSecret(pb, 'buffer', 'hex');
		var s2 = b.computeSecret(pa, 'buffer', 'hex');
		var comp = a.getPublicKey('buffer', 'compressed')[0];
		var unc = a.getPublicKey('buffer', 'uncompressed')[0];
		var conv = crypto.ECDH.convertKey(a.getPublicKey('buffer', 'compressed'),
			'prime256v1', 'buffer', 'hex', 'uncompressed');
		return [s1 === s2, comp === 2 || comp === 3, unc, conv.length,
		        crypto.getCurves().indexOf('prime256v1') >= 0].join('|');
	})()`)
	jstest.Equal(t, got, "true|true|4|130|true")
}

func TestScryptVector(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var out = crypto.scryptSync('', '', 64, { N: 16, r: 1, p: 1 });
		var s = '';
		for (var i = 0; i < out.length; i++) s += ('0' + out[i].toString(16)).slice(-2);
		return s;
	})()`)
	jstest.Equal(t, got, "77d6576238657b203b19ca42c18a0497f16b4844e3074ae8dfdffa3fede21442"+
		"fcd0069ded0948f8326a753a0fc81f17e8d3e0fb2e0d3628cf35e20c38d18906")

	err := jstest.Throws(t, e, `crypto.scryptSync('', '', 16, { N: 3 })`)
	jstest.Equal(t, err.Code, "ERR_CRYPTO_INVALID_SCRYPT_PARAMS")
}

func TestFipsFlag(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `String(crypto.getFips())`), "0")
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		crypto.setFips(true);
		var on = crypto.getFips();
		crypto.setFips(false);
		return [on, crypto.getFips()].join('|');
	})()`), "1|0")
	err := jstest.Throws(t, e, `crypto.setFips('yes')`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
}

func TestKeyObjectFromCryptoKey(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var data = new Uint8Array(32);
		for (var i = 0; i < data.length; i++) data[i] = i;
		crypto.subtle.importKey('raw', data, { name: 'AES-GCM' }, false, ['encrypt'])
			.then(function (k) { globalThis.__key = crypto.KeyObject.from(k); });
		return 'pending';
	})()`), "pending")
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e,
		`[globalThis.__key.type, globalThis.__key.symmetricKeySize].join('|')`), "secret|32")

	err := jstest.Throws(t, e, `crypto.KeyObject.from({})`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
}

// Node accepts 512-bit RSA keys (its fixtures use them); crypto/rsa refuses
// anything below 1024 bits unless GODEBUG rsa1024min=0 is in force.
func TestGenerateKeyPair512(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		try {
			var kp = crypto.generateKeyPairSync('rsa', { modulusLength: 512 });
			return kp.publicKey.asymmetricKeyDetails.modulusLength + '|' +
				kp.publicKey.asymmetricKeyType;
		} catch (e) { return 'ERR:' + e.message; }
	})()`), "512|rsa")
}

func TestHmacKeyObjectKey(t *testing.T) {
	e := setup(t)
	// A secret KeyObject is valid key material, and must hash identically to
	// the raw bytes it wraps.
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var k = crypto.createSecretKey(Buffer.from('key'));
		return crypto.createHmac('sha256', k).update('abc').digest('hex') ===
			crypto.createHmac('sha256', 'key').update('abc').digest('hex');
	})()`), "true")

	err := jstest.Throws(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('ed25519');
		crypto.createHmac('sha256', kp.publicKey);
	})()`)
	jstest.Equal(t, err.Code, "ERR_CRYPTO_INVALID_KEYTYPE")
}

// Node's Hmac/Hash/Sign/Verify constructors are callable without `new`; the
// official tests exercise that form directly.
func TestLegacyConstructorsCallable(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var s = crypto.Sign('sha256');
		var v = crypto.Verify('sha256');
		var h = crypto.Hmac('sha256', 'key');
		var d = crypto.Hash('sha1');
		return [s instanceof crypto.Sign, v instanceof crypto.Verify,
		        h instanceof crypto.Hmac, d instanceof crypto.Hash,
		        h.update('abc').digest('hex') ===
		          crypto.createHmac('sha256', 'key').update('abc').digest('hex')].join('|');
	})()`), "true|true|true|true|true")
}

// KeyObject validates its constructor arguments before rejecting construction.
func TestKeyObjectConstructorValidation(t *testing.T) {
	e := setup(t)
	err := jstest.Throws(t, e, `new crypto.KeyObject('wrong_type')`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_VALUE")
	jstest.Equal(t, err.Message, "The argument 'type' is invalid. Received 'wrong_type'")

	err = jstest.Throws(t, e, `new crypto.KeyObject('secret', '')`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
	jstest.Equal(t, err.Message, `The "handle" argument must be of type object. Received type string ('')`)
}

// The key object prototypes are split per branch so the brand checks in
// test-crypto-keyobject-brand-check find the accessors where Node puts them.
func TestKeyObjectPrototypeBranches(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var k = crypto.createSecretKey(Buffer.alloc(4));
		var size = Object.getOwnPropertyDescriptor(
		  Object.getPrototypeOf(k), 'symmetricKeySize').get;
		var out = [];
		try { size.call({}); } catch (e) { out.push(e.code); }
		var p = crypto.generateKeyPairSync('rsa', { modulusLength: 512 }).publicKey;
		var a = Object.getOwnPropertyDescriptor(
		  Object.getPrototypeOf(Object.getPrototypeOf(p)),
		  'asymmetricKeyType').get;
		out.push(p.asymmetricKeyType, a.call(p), String(k.asymmetricKeyType),
		         k.symmetricKeySize, 'symmetricKeySize' in p);
		try { a.call(k); } catch (e) { out.push(e.code); }
		return out.join('|');
	})()`), "ERR_INVALID_THIS|rsa|rsa|undefined|4|false|ERR_INVALID_THIS")
}

// OpenSSL lists single and triple DES; Go's crypto/des backs the same names.
func TestDESCiphers(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var key = Buffer.from('0123456789abcd0123456789');
		var iv = Buffer.from('12345678');
		var c = crypto.createCipheriv('des-ede3-cbc', key, iv);
		var ct = Buffer.concat([c.update('hello'), c.final()]);
		var d = crypto.createDecipheriv('des-ede3-cbc', key, iv);
		return [Buffer.concat([d.update(ct), d.final()]).toString(),
		        crypto.getCiphers().indexOf('des-ede3-cbc') >= 0,
		        c instanceof crypto.Cipheriv,			        crypto.getCipherInfo('des-ede3-cbc').keyLength].join('|');
	})()`), "hello|true|true|24")
}

// A custom Matcher (createMac) covers the EVP_MAC HMAC provider Node 24 lists.
func TestCreateMac(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var key = Buffer.alloc(16, 7);
		var mac = crypto.createMac('hmac', key, { digest: 'sha256' });
		var out = mac.update('abc').final();
		var want = crypto.createHmac('sha256', key).update('abc').digest();
		return [crypto.getMacs().indexOf('hmac') >= 0, out.equals(want)].join('|');
	})()`), "true|true")

	err := jstest.Throws(t, e, `crypto.createMac('nope', Buffer.alloc(1), { digest: 'sha256' })`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_VALUE")
}

// crypto.hash() validates its output encoding differently from digest().
func TestHashOneShotEncodingValidation(t *testing.T) {
	e := setup(t)
	err := jstest.Throws(t, e, `crypto.hash('sha1', 'test', 0)`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")

	err = jstest.Throws(t, e, `crypto.hash('sha1', 'test', 'not an encoding')`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_VALUE")

	jstest.Equal(t, jstest.Str(t, e, `crypto.hash('sha1', 'test', 'hex')`),
		"a94a8fe5ccb19ba61c4c0873d391e987982fbbd3")
}

func TestX509Certificate(t *testing.T) {
	e := setup(t)
	e.Global().SetStr("__cert", testCertificatePEM(t))

	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var c = new crypto.X509Certificate(__cert);
		return [c instanceof crypto.X509Certificate,
		        c.subject.indexOf('CN=example.test') >= 0,
		        c.subjectAltName.indexOf('DNS:example.test') >= 0,
		        c.ca, c.checkHost('example.test'), c.checkHost('nope.test'),
		        c.checkEmail('dev@example.test'), c.checkIP('127.0.0.1'),
		        c.checkIP('10.0.0.1'), Buffer.isBuffer(c.raw),
		        c.toString().indexOf('BEGIN CERTIFICATE') >= 0,
		        c.serialNumber, c.keyUsage.join(',')].join('|');
	})()`), "true|true|true|false|true|false|true|true|false|true|true|1234|digitalSignature,keyCertSign")

	jstest.Equal(t, jstest.Str(t, e, `(function(){
		var c = new crypto.X509Certificate(__cert);
		var getter = Object.getOwnPropertyDescriptor(
		  crypto.X509Certificate.prototype, 'subject').get;
		var out = [];
		try { getter.call({}); } catch (e) { out.push(e.code); }
		try { crypto.X509Certificate.prototype.toString.call(1); }
		catch (e) { out.push(e.code); }
		out.push(c.subject === getter.call(c));
		return out.join('|');
	})()`), "ERR_INVALID_THIS|ERR_INVALID_THIS|true")

	err := jstest.Throws(t, e, `crypto.X509Certificate.prototype.checkHost.call(
		new crypto.X509Certificate(__cert), 123)`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
}

func testCertificatePEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(0x1234),
		Subject:               pkix.Name{CommonName: "example.test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		DNSNames:              []string{"example.test"},
		EmailAddresses:        []string{"dev@example.test"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestRSAEncryptDecrypt(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('rsa', { modulusLength: 1024 });
		var msg = Buffer.from('secret message');
		var ct = crypto.publicEncrypt(kp.publicKey, msg);
		var pt = crypto.privateDecrypt(kp.privateKey, ct);
		var pkcs1 = crypto.constants.RSA_PKCS1_PADDING;
		var ct2 = crypto.publicEncrypt({ key: kp.publicKey, padding: pkcs1 }, msg);
		var pt2 = crypto.privateDecrypt({ key: kp.privateKey, padding: pkcs1 }, ct2);
		var ct3 = crypto.privateEncrypt(kp.privateKey, msg);
		var pt3 = crypto.publicDecrypt(kp.publicKey, ct3);
		var pem = kp.publicKey.export({ type: 'spki', format: 'pem' });
		var pt4 = crypto.privateDecrypt(kp.privateKey, crypto.publicEncrypt(pem, msg));
		var empty = crypto.privateDecrypt(kp.privateKey,
			crypto.publicEncrypt(kp.publicKey, Buffer.alloc(0)));
		return [pt.toString(), pt2.toString(), pt3.toString(), pt4.toString(),
		        empty.length, ct.length].join('|');
	})()`)
	jstest.Equal(t, got, "secret message|secret message|secret message|secret message|0|128")

	err := jstest.Throws(t, e, `crypto.privateDecrypt(
		crypto.generateKeyPairSync('rsa', { modulusLength: 1024 }).publicKey,
		Buffer.alloc(128))`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")
}

func TestSignOptions(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var kp = crypto.generateKeyPairSync('ec', { namedCurve: 'P-256' });
		var msg = Buffer.from('abc');
		var sig = crypto.sign('sha256', msg, { key: kp.privateKey, dsaEncoding: 'ieee-p1363' });
		var ok = crypto.verify('sha256', msg,
			{ key: kp.publicKey, dsaEncoding: 'ieee-p1363' }, sig);
		var der = crypto.sign('sha256', msg, kp.privateKey);
		return [sig.length, ok, der[0] === 0x30].join('|');
	})()`)
	jstest.Equal(t, got, "64|true|true")
}

func TestCipherSecretKeyObject(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var key = crypto.createSecretKey(Buffer.alloc(16, 1));
		var iv = Buffer.alloc(16, 2);
		var c = crypto.createCipheriv('aes-128-cbc', key, iv);
		var ct = Buffer.concat([c.update('hi'), c.final()]);
		return [ct.length, c instanceof crypto.Cipheriv].join('|');
	})()`)
	jstest.Equal(t, got, "16|true")
}

// DSA keys cannot be marshalled by crypto/x509, so the module carries its own
// ASN.1 encoder/decoder for them.
func TestDSAKeyMarshalRoundTrip(t *testing.T) {
	priv := &dsa.PrivateKey{}
	priv.Parameters = dsa.Parameters{P: big.NewInt(23), Q: big.NewInt(11), G: big.NewInt(2)}
	priv.X = big.NewInt(3)
	priv.Y = new(big.Int).Exp(big.NewInt(2), big.NewInt(3), big.NewInt(23))

	der, err := dsaMarshalPKCS8(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	back, ok := dsaParsePKCS8(der)
	if !ok {
		t.Fatalf("parse pkcs8 failed")
	}
	if back.X.Cmp(priv.X) != 0 || back.P.Cmp(priv.P) != 0 || back.Q.Cmp(priv.Q) != 0 {
		t.Fatalf("pkcs8 round trip mismatch")
	}

	spki, err := dsaMarshalSPKI(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal spki: %v", err)
	}
	pub, ok := dsaParseSPKI(spki)
	if !ok {
		t.Fatalf("parse spki failed")
	}
	if pub.Y.Cmp(priv.Y) != 0 || pub.P.Cmp(priv.P) != 0 {
		t.Fatalf("spki round trip mismatch")
	}
}

func TestRandomUUIDv7(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function(){
		var a = crypto.randomUUIDv7();
		var b = crypto.randomUUIDv7();
		return [/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(a),
		        a !== b, a.length].join('|');
	})()`)
	jstest.Equal(t, got, "true|true|36")
}
