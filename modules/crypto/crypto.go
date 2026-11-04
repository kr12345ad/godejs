// Package crypto implements the Node.js `crypto` builtin on top of the Go
// standard library. Only the algorithms Go ships are exposed; anything else
// fails with Node's ERR_CRYPTO_INVALID_DIGEST instead of pretending.
package crypto

import (
	"bytes"
	stdcrypto "crypto"
	"crypto/aes"
	stdcipher "crypto/cipher"
	"crypto/des"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	stdhash "hash"
	"math"
	"math/big"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	// Node's own fixtures include 512-bit RSA keys, while crypto/rsa refuses
	// anything under 1024 bits unless rsa1024min=0 is set. Merge into any
	// inherited GODEBUG instead of overwriting it.
	const want = "rsa1024min=0"
	if cur := os.Getenv("GODEBUG"); !strings.Contains(cur, "rsa1024min=") {
		if cur == "" {
			_ = os.Setenv("GODEBUG", want)
		} else {
			_ = os.Setenv("GODEBUG", cur+","+want)
		}
	}
	env.Register("crypto", Install)
}

const (
	maxRandomBytes = 2147483647
	maxEntropy     = 65536
)

// hashState buffers everything written to a Hash/Hmac. Go's hash.Hash does not
// expose a clone for every algorithm, and buffering lets copy() work uniformly.
type hashState struct {
	algo   string
	key    []byte
	mac    bool
	buf    []byte
	done   bool
	xof    string
	outLen int
	stream bool
	sum    []byte
}

type cipherState struct {
	decrypt   bool
	name      string
	mode      string
	block     stdcipher.Block
	blockMode stdcipher.BlockMode
	stream    stdcipher.Stream
	iv        []byte
	aead      stdcipher.AEAD
	aad       []byte
	tag       []byte
	tagLen    int
	tagGiven  bool
	autoPad   bool
	buf       []byte
	done      bool
	updated   bool
	// outEnc is the output encoding locked in by the first update()/final()
	// call; Node rejects a later call that asks for a different one.
	outEnc    string
	outEncSet bool
}

type keyState struct {
	kind        string // "secret", "private" or "public"
	typ         string // rsa, rsa-pss, ec, ed25519, x25519, dsa
	key         any
	curve       string
	hashAlg     string
	mgf1HashAlg string
	saltLen     int
	// padding and dsaEncoding come from the per-call key options of
	// sign/verify and select the RSA padding mode or the ECDSA signature
	// encoding; clone() keeps them from leaking into cached KeyObjects.
	padding     int
	dsaEncoding string
}

type signState struct {
	algo  string
	buf   []byte
	done  bool
	isSig bool
}

// dhState is the mutable side of a DiffieHellman/DiffieHellmanGroup object.
type dhState struct {
	prime   *big.Int
	gen     *big.Int
	priv    *big.Int
	pub     *big.Int
	other   *big.Int
	trusted bool // prime/generator chosen by OpenSSL (or a MODP group)
}

// ecdhState mirrors Node's ECDH object: a curve plus an optional key pair.
type ecdhState struct {
	curve ecdh.Curve
	name  string
	priv  *ecdh.PrivateKey
	pub   *ecdh.PublicKey
}

type module struct {
	e              *env.Env
	hashProto      js.Object
	macProto       js.Object
	cipherProto    js.Object
	decipherProto  js.Object
	keyProto       js.Object
	secretProto    js.Object
	asymProto      js.Object
	pubProto       js.Object
	privProto      js.Object
	signProto      js.Object
	verifyProto    js.Object
	cryptoKeyProto js.Object
	mac2Proto      js.Object
	x509Proto      js.Object
	x509Ctor       js.Object
	x509s          map[int]*x509State
	cipherCtor     js.Object
	decipherCtor   js.Object
	keyCtor        js.Object
	signCtor       js.Object
	verifyCtor     js.Object
	cryptoKeyCtor  js.Object
	dhProto        js.Object
	dhGroupProto   js.Object
	ecdhProto      js.Object
	dhCtor         js.Object
	dhGroupCtor    js.Object
	ecdhCtor       js.Object

	bigintFn js.Value

	mu         sync.Mutex
	seq        int
	states     map[int]*hashState
	ciphers    map[int]*cipherState
	keys       map[int]*keyState
	signers    map[int]*signState
	cryptoKeys map[int][]byte
	dhs        map[int]*dhState
	ecdhs      map[int]*ecdhState

	// fips is the flag process.crypto.setFips toggles; getFips reports it.
	fips bool
}

// Install builds the crypto module's exports object.
func Install(e *env.Env) js.Object {
	m := &module{
		e:          e,
		states:     map[int]*hashState{},
		ciphers:    map[int]*cipherState{},
		keys:       map[int]*keyState{},
		signers:    map[int]*signState{},
		cryptoKeys: map[int][]byte{},
		dhs:        map[int]*dhState{},
		ecdhs:      map[int]*ecdhState{},
		x509s:      map[int]*x509State{},
	}
	exports := e.JS.Obj()

	if fn, err := e.JS.Eval("(function(s){return BigInt(s);})", "[gode:crypto/bigint]"); err == nil {
		m.bigintFn = fn.Protect()
	}

	m.hashProto = e.JS.Obj().Keep()
	m.macProto = e.JS.Obj().Keep()
	// Hash is constructible (new crypto.Hash('sha256')) and also callable
	// without new, so it goes through the same class builder as Hmac/Sign.
	hashCtor := m.classWithCtor("Hash", m.hashProto, func(args []js.Value) (js.Value, error) {
		return m.newHash(args)
	})
	// Hmac, like Hash, is callable without `new`: Hmac('sha256', key).
	macCtor := m.classWithCtor("Hmac", m.macProto, func(args []js.Value) (js.Value, error) {
		algo := arg(args, 0)
		if !algo.IsString() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(algo))
		}
		_, canonical, ok := lookup(algo.Str())
		if !ok {
			return js.Value{}, invalidDigest()
		}
		key, err := m.keyArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		return m.instance(&hashState{algo: canonical, key: key, mac: true}, m.macProto).V(), nil
	})

	m.cipherProto = e.JS.Obj().Keep()
	m.decipherProto = e.JS.Obj().Keep()
	// Node's Cipheriv/Decipheriv are plain functions: calling them without
	// `new` returns a new instance, exactly as `new` does.
	m.cipherCtor = m.classWithCtor("Cipheriv", m.cipherProto, func(args []js.Value) (js.Value, error) {
		return m.newCipher(args, false)
	}).Keep()
	m.decipherCtor = m.classWithCtor("Decipheriv", m.decipherProto, func(args []js.Value) (js.Value, error) {
		return m.newCipher(args, true)
	}).Keep()
	m.installCipherProto(m.cipherProto, false)
	m.installCipherProto(m.decipherProto, true)

	m.keyProto = e.JS.Obj().Keep()
	// Node's key objects form a small prototype tree: shared KeyObject
	// members, then a secret branch carrying symmetricKeySize and an
	// asymmetric branch carrying the asymmetric accessors.
	// These four are only reachable from Go, so they must be rooted
	// explicitly or JavaScriptCore may collect them between install and use.
	m.secretProto = e.JS.Obj().Keep()
	m.asymProto = e.JS.Obj().Keep()
	m.pubProto = e.JS.Obj().Keep()
	m.privProto = e.JS.Obj().Keep()
	// KeyObject is constructible only far enough to validate its arguments;
	// Node uses the constructor to report invalid types and handles.
	m.keyCtor = m.classWithCtor("KeyObject", m.keyProto, func(args []js.Value) (js.Value, error) {
		t := arg(args, 0)
		if !t.IsString() {
			return js.Value{}, typeErr(`The "type" argument must be of type string. Received %s`, describe(t))
		}
		switch t.Str() {
		case "secret", "public", "private":
		default:
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				fmt.Sprintf("The argument 'type' is invalid. Received '%s'", t.Str()))
		}
		handle := arg(args, 1)
		if !handle.IsObject() {
			return js.Value{}, typeErr(`The "handle" argument must be of type object. Received %s`, describe(handle))
		}
		return js.Value{}, js.Throw("TypeError", "Illegal constructor")
	}).Keep()
	m.installKeyProto()
	// KeyObject.from(CryptoKey) turns a WebCrypto key into a KeyObject. Only
	// symmetric CryptoKeys exist here (importKey accepts raw AES material),
	// so the result is always a secret KeyObject.
	m.keyCtor.SetFunc("from", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if !v.IsObject() {
			return js.Value{}, typeErr(`The "key" argument must be an instance of CryptoKey. Received %s`, describe(v))
		}
		id := v.Object().Get("$godeCryptoKeyId")
		if !id.Valid() || !id.IsNumber() {
			return js.Value{}, typeErr(`The "key" argument must be an instance of CryptoKey. Received %s`, describe(v))
		}
		data, err := m.cryptoKeyBytes(v)
		if err != nil {
			return js.Value{}, err
		}
		return m.newKeyObject(&keyState{kind: "secret", typ: "secret", key: append([]byte{}, data...)}).V(), nil
	})

	m.signProto = e.JS.Obj().Keep()
	m.verifyProto = e.JS.Obj().Keep()
	// Node's Sign and Verify are callable without `new`: Sign('sha256')
	// returns an instance just like createSign does.
	m.signCtor = m.classWithCtor("Sign", m.signProto, func(args []js.Value) (js.Value, error) {
		if !arg(args, 0).IsString() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(arg(args, 0)))
		}
		return m.newSigner(m.algoArg(arg(args, 0)), false)
	}).Keep()
	m.verifyCtor = m.classWithCtor("Verify", m.verifyProto, func(args []js.Value) (js.Value, error) {
		if !arg(args, 0).IsString() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(arg(args, 0)))
		}
		return m.newSigner(m.algoArg(arg(args, 0)), true)
	}).Keep()
	m.installSignProto(m.signProto, false)
	m.installSignProto(m.verifyProto, true)

	// A Mac (crypto.createMac) exposes update/final, unlike Hmac's digest.
	m.mac2Proto = e.JS.Obj().Keep()
	m.mac2Proto.SetFuncHidden("update", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.update(this, args)
	})
	m.mac2Proto.SetFuncHidden("final", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.macFinal(this, args)
	})

	m.cryptoKeyProto = e.JS.Obj().Keep()
	m.cryptoKeyCtor = m.class("CryptoKey", m.cryptoKeyProto).Keep()
	e.Global().Set("CryptoKey", m.cryptoKeyCtor.V())

	m.dhProto = e.JS.Obj().Keep()
	m.dhGroupProto = e.JS.Obj().Keep()
	m.ecdhProto = e.JS.Obj().Keep()
	// A DiffieHellmanGroup is a DiffieHellman with a fixed prime/generator;
	// the key setters are hidden below because Node's group type lacks them.
	m.dhGroupProto.SetProto(m.dhProto)
	m.dhCtor = m.classWithCtor("DiffieHellman", m.dhProto, func(args []js.Value) (js.Value, error) {
		return m.createDH(args, false)
	}).Keep()
	m.dhGroupCtor = m.classWithCtor("DiffieHellmanGroup", m.dhGroupProto, func(args []js.Value) (js.Value, error) {
		return m.createDHGroup(args)
	}).Keep()
	m.ecdhCtor = m.classWithCtor("ECDH", m.ecdhProto, func(args []js.Value) (js.Value, error) {
		return m.createECDH(args)
	}).Keep()
	m.installDHProto()
	// Node's DiffieHellmanGroup has no key setters at all.
	m.dhGroupProto.Set("setPrivateKey", m.e.JS.Undefined())
	m.dhGroupProto.Set("setPublicKey", m.e.JS.Undefined())
	m.installECDHProto()

	m.hashProto.SetFuncHidden("update", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.update(this, args)
	})
	m.hashProto.SetFuncHidden("digest", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.digest(this, args)
	})
	m.hashProto.SetFuncHidden("copy", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.copyHash(this, args)
	})
	m.macProto.SetFuncHidden("update", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.update(this, args)
	})
	m.macProto.SetFuncHidden("digest", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.digest(this, args)
	})
	m.hashProto.SetFuncHidden("write", func(this js.Object, args []js.Value) (js.Value, error) {
		if _, err := m.update(this, args); err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(true), nil
	})
	m.hashProto.SetFuncHidden("end", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.stateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if d := arg(args, 0); d.Valid() && !d.IsUndefined() && !d.IsNull() {
			if _, err := m.update(this, args); err != nil {
				return js.Value{}, err
			}
		}
		if !st.done {
			st.done = true
			st.stream = true
			sum, err := m.sum(st)
			if err != nil {
				return js.Value{}, err
			}
			st.sum = sum
		}
		return this.V(), nil
	})
	m.hashProto.SetFuncHidden("read", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.stateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if !st.stream || st.sum == nil {
			st.done = true
			st.stream = true
			sum, err := m.sum(st)
			if err != nil {
				return js.Value{}, err
			}
			st.sum = sum
		}
		return m.bytesValue(st.sum), nil
	})

	exports.Set("Hash", hashCtor.V())
	exports.Set("Hmac", macCtor.V())
	exports.Set("Cipheriv", m.cipherCtor.V())
	exports.Set("Decipheriv", m.decipherCtor.V())
	exports.Set("KeyObject", m.keyCtor.V())
	exports.Set("Sign", m.signCtor.V())
	exports.Set("Verify", m.verifyCtor.V())
	exports.Set("DiffieHellman", m.dhCtor.V())
	exports.Set("DiffieHellmanGroup", m.dhGroupCtor.V())
	exports.Set("ECDH", m.ecdhCtor.V())

	exports.SetFunc("createHash", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.newHash(args)
	})
	exports.SetFunc("createHmac", func(this js.Object, args []js.Value) (js.Value, error) {
		algo := arg(args, 0)
		if !algo.IsString() {
			return js.Value{}, typeErr(`The "hmac" argument must be of type string. Received %s`, describe(algo))
		}
		fn, canonical, ok := lookup(algo.Str())
		if !ok {
			return js.Value{}, invalidDigest()
		}
		_ = fn
		key, err := m.keyArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		st := &hashState{algo: canonical, key: key, mac: true}
		return m.instance(st, m.macProto).V(), nil
	})
	exports.SetFunc("hash", func(this js.Object, args []js.Value) (js.Value, error) {
		algo := arg(args, 0)
		if !algo.IsString() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(algo))
		}
		data, err := m.dataArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		// Node defaults one-shot hash output to a hex string; a Buffer is only
		// returned when the caller asks for outputEncoding: 'buffer'. The third
		// argument may also be an options object.
		enc := "hex"
		outLen := -1
		opts := arg(args, 2)
		if opts.Valid() && opts.IsString() {
			if !knownEncoding(opts.Str()) {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
					fmt.Sprintf("The argument 'outputEncoding' is invalid. Received '%s'", opts.Str()))
			}
			enc = opts.Str()
		} else if opts.Valid() && opts.IsObject() && !opts.IsNull() {
			o := opts.Object()
			if v := o.Get("outputEncoding"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
				if !v.IsString() {
					return js.Value{}, typeErr(`The "outputEncoding" argument must be of type string. Received %s`, describe(v))
				}
				if !knownEncoding(v.Str()) {
					return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
						fmt.Sprintf("The argument 'outputEncoding' is invalid. Received '%s'", v.Str()))
				}
				enc = v.Str()
			}
			if v := o.Get("outputLength"); v.Valid() && !v.IsUndefined() {
				n, lenErr := m.outputLengthArg(v)
				if lenErr != nil {
					return js.Value{}, lenErr
				}
				outLen = n
			}
		} else if opts.Valid() && !opts.IsUndefined() && !opts.IsNull() {
			return js.Value{}, typeErr(`The "options" argument must be of type object. Received %s`, describe(opts))
		}
		var sum []byte
		if xof, def, ok := lookupXOF(algo.Str()); ok {
			if outLen >= 0 {
				def = outLen
			}
			var sh *sha3.SHAKE
			if xof == "shake128" {
				sh = sha3.NewSHAKE128()
			} else {
				sh = sha3.NewSHAKE256()
			}
			sh.Write(data)
			sum = make([]byte, def)
			if _, err := sh.Read(sum); err != nil {
				return js.Value{}, js.Throw("Error", err.Error())
			}
		} else {
			fn, _, ok := lookup(algo.Str())
			if !ok {
				return js.Value{}, invalidDigest()
			}
			if outLen >= 0 && outLen != fn().Size() {
				return js.Value{}, js.Throw("Error",
					fmt.Sprintf("Output length %d is invalid for %s, which does not support XOF", outLen, algo.Str()))
			}
			h := fn()
			h.Write(data)
			sum = h.Sum(nil)
		}
		if strings.EqualFold(enc, "buffer") {
			return m.bytesValue(sum), nil
		}
		return m.encodeDigest(sum, enc)
	})
	exports.SetFunc("randomBytes", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomBytes(args)
	})
	exports.SetFunc("randomFillSync", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomFill(args, false)
	})
	exports.SetFunc("randomFill", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomFill(args, true)
	})
	exports.SetFunc("randomInt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomInt(args)
	})
	exports.SetFunc("randomUUID", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomUUID(args)
	})
	exports.SetFunc("timingSafeEqual", func(this js.Object, args []js.Value) (js.Value, error) {
		a, ok := m.bytesOf(arg(args, 0))
		if !ok {
			return js.Value{}, typeErr(`The "a" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 0)))
		}
		b, ok := m.bytesOf(arg(args, 1))
		if !ok {
			return js.Value{}, typeErr(`The "b" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 1)))
		}
		if len(a) != len(b) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_CRYPTO_TIMING_SAFE_EQUAL_LENGTH", "Input buffers must have the same byte length")
		}
		return e.JS.Bool(subtle.ConstantTimeCompare(a, b) == 1), nil
	})
	getRandomValues := func(this js.Object, args []js.Value) (js.Value, error) {
		return m.getRandomValues(args)
	}
	exports.SetFunc("getRandomValues", getRandomValues)

	exports.SetFunc("createCipheriv", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.newCipher(args, false)
	})
	exports.SetFunc("createDecipheriv", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.newCipher(args, true)
	})
	exports.SetFunc("createPublicKey", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.loadKey(arg(args, 0), false)
		if err != nil {
			return js.Value{}, err
		}
		return m.newKeyObject(st).V(), nil
	})
	exports.SetFunc("createPrivateKey", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.loadKey(arg(args, 0), true)
		if err != nil {
			return js.Value{}, err
		}
		if st.kind != "private" {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a private key")
		}
		return m.newKeyObject(st).V(), nil
	})
	exports.SetFunc("createSecretKey", func(this js.Object, args []js.Value) (js.Value, error) {
		var data []byte
		v := arg(args, 0)
		if v.IsString() {
			enc := "utf8"
			if e := arg(args, 1); e.Valid() && !e.IsUndefined() && !e.IsNull() {
				enc = e.Str()
			}
			d, err := decodeInput(v.Str(), enc)
			if err != nil {
				return js.Value{}, err
			}
			data = d
		} else {
			b, ok := m.bytesOf(v)
			if !ok {
				return js.Value{}, typeErr(`The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
			}
			data = b
		}
		return m.newKeyObject(&keyState{kind: "secret", typ: "secret", key: append([]byte{}, data...)}).V(), nil
	})
	exports.SetFunc("generateKeyPairSync", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.generateKeyPair(args, false)
	})
	exports.SetFunc("generateKeyPair", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.generateKeyPair(args, true)
	})
	exports.SetFunc("generateKeySync", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.generateSecretKey(args, false)
	})
	exports.SetFunc("generateKey", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.generateSecretKey(args, true)
	})
	exports.SetFunc("sign", func(this js.Object, args []js.Value) (js.Value, error) {
		data, err := m.dataArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		st, err := m.loadKeyForOp(arg(args, 2), true)
		if err != nil {
			return js.Value{}, err
		}
		sig, err := m.signData(m.algoArg(arg(args, 0)), data, st)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(sig), nil
	})
	exports.SetFunc("verify", func(this js.Object, args []js.Value) (js.Value, error) {
		data, err := m.dataArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		st, err := m.loadKeyForOp(arg(args, 2), false)
		if err != nil {
			return js.Value{}, err
		}
		sig, ok := m.bytesOf(arg(args, 3))
		if !ok {
			return js.Value{}, typeErr(`The "signature" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 3)))
		}
		okVerify, err := m.verifyData(m.algoArg(arg(args, 0)), data, st, sig)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(okVerify), nil
	})
	exports.SetFunc("createSign", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.newSigner(m.algoArg(arg(args, 0)), false)
	})
	exports.SetFunc("createVerify", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.newSigner(m.algoArg(arg(args, 0)), true)
	})
	exports.SetFunc("getHashes", func(this js.Object, args []js.Value) (js.Value, error) {
		names := []string{"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha512-224", "sha512-256", "sha3-256", "sha3-384", "sha3-512", "shake128", "shake256"}
		vals := make([]js.Value, len(names))
		for i, n := range names {
			vals[i] = m.e.JS.Str(n)
		}
		return m.e.JS.Arr(vals).V(), nil
	})
	exports.SetFunc("getCiphers", func(this js.Object, args []js.Value) (js.Value, error) {
		names := cipherNames()
		vals := make([]js.Value, len(names))
		for i, n := range names {
			vals[i] = m.e.JS.Str(n)
		}
		return m.e.JS.Arr(vals).V(), nil
	})
	exports.SetFunc("getCipherInfo", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.getCipherInfo(args)
	})
	// OpenSSL 3 exposes EVP_MAC; Go only has HMAC, so that is what getMacs
	// advertises and createMac accepts.
	exports.SetFunc("getMacs", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Arr([]js.Value{m.e.JS.Str("hmac")}).V(), nil
	})
	exports.SetFunc("createMac", func(this js.Object, args []js.Value) (js.Value, error) {
		algo := arg(args, 0)
		if !algo.IsString() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(algo))
		}
		if !strings.EqualFold(algo.Str(), "hmac") {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				fmt.Sprintf("The argument 'algorithm' is invalid. Received '%s'", algo.Str()))
		}
		digest := "sha256"
		if opts := arg(args, 2); opts.IsObject() {
			if d := opts.Object().Get("digest"); d.IsString() {
				digest = d.Str()
			} else {
				return js.Value{}, typeErr(`The "options.digest" argument must be of type string. Received %s`, describe(d))
			}
		} else if !arg(args, 2).Valid() || arg(args, 2).IsUndefined() {
			return js.Value{}, typeErr(`The "options.digest" argument must be of type string. Received undefined`)
		}
		fn, canonical, ok := lookup(digest)
		if !ok {
			return js.Value{}, invalidDigest()
		}
		_ = fn
		key, err := m.keyArg(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		st := &hashState{algo: canonical, key: key, mac: true}
		return m.instance(st, m.mac2Proto).V(), nil
	})
	exports.SetFunc("hkdfSync", func(this js.Object, args []js.Value) (js.Value, error) {
		out, err := m.hkdf(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(out), nil
	})
	exports.SetFunc("hkdf", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node validates the derivation arguments before the callback.
		vars := firstArgs(args, 5)
		out, err := m.hkdf(vars)
		if err != nil {
			return js.Value{}, err
		}
		cb := arg(args, 5)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		m.e.NextTick(cb, m.e.JS.Null(), m.bytesValue(out))
		return js.Value{}, nil
	})
	exports.SetFunc("pbkdf2Sync", func(this js.Object, args []js.Value) (js.Value, error) {
		out, err := m.pbkdf2(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(out), nil
	})
	exports.SetFunc("pbkdf2", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node validates the derivation arguments before the callback.
		out, err := m.pbkdf2(firstArgs(args, 5))
		if err != nil {
			return js.Value{}, err
		}
		cb := arg(args, 5)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		m.e.NextTick(cb, m.e.JS.Null(), m.bytesValue(out))
		return js.Value{}, nil
	})
	exports.SetFunc("checkPrimeSync", func(this js.Object, args []js.Value) (js.Value, error) {
		ok, err := m.checkPrime(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(ok), nil
	})
	exports.SetFunc("checkPrime", func(this js.Object, args []js.Value) (js.Value, error) {
		cb := arg(args, 2)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		ok, err := m.checkPrime(firstArgs(args, 2))
		if err != nil {
			return js.Value{}, err
		}
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Bool(ok))
		return js.Value{}, nil
	})
	exports.SetFunc("generatePrimeSync", func(this js.Object, args []js.Value) (js.Value, error) {
		buf, err := m.generatePrime(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(buf), nil
	})
	exports.SetFunc("generatePrime", func(this js.Object, args []js.Value) (js.Value, error) {
		cb := arg(args, 2)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		buf, err := m.generatePrime(firstArgs(args, 2))
		if err != nil {
			return js.Value{}, err
		}
		m.e.NextTick(cb, m.e.JS.Null(), m.bytesValue(buf))
		return js.Value{}, nil
	})
	exports.SetFunc("getFips", func(this js.Object, args []js.Value) (js.Value, error) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.fips {
			return m.e.JS.Num(1), nil
		}
		return m.e.JS.Num(0), nil
	})
	exports.SetFunc("setFips", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if !v.IsBool() {
			return js.Value{}, typeErr(`The "fips" argument must be of type boolean. Received %s`, describe(v))
		}
		m.mu.Lock()
		m.fips = v.Bool()
		m.mu.Unlock()
		return m.e.JS.Undefined(), nil
	})
	exports.SetFunc("createDiffieHellman", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.createDH(args, false)
	})
	groupFn := func(this js.Object, args []js.Value) (js.Value, error) {
		return m.createDHGroup(args)
	}
	exports.SetFunc("getDiffieHellman", groupFn)
	exports.SetFunc("createDiffieHellmanGroup", groupFn)
	exports.SetFunc("createECDH", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.createECDH(args)
	})
	exports.SetFunc("getCurves", func(this js.Object, args []js.Value) (js.Value, error) {
		names := []string{"prime256v1", "secp384r1", "secp521r1"}
		vals := make([]js.Value, len(names))
		for i, n := range names {
			vals[i] = m.e.JS.Str(n)
		}
		return m.e.JS.Arr(vals).V(), nil
	})
	exports.SetFunc("randomUUIDv7", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomUUIDv7(args)
	})
	exports.SetFunc("scryptSync", func(this js.Object, args []js.Value) (js.Value, error) {
		out, err := m.scrypt(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(out), nil
	})
	exports.SetFunc("scrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		out, cb, err := m.scryptAsync(args)
		if err != nil {
			return js.Value{}, err
		}
		m.e.NextTick(cb, m.e.JS.Null(), m.bytesValue(out))
		return js.Value{}, nil
	})

	// Deprecated aliases Node keeps next to randomBytes. They are
	// non-enumerable and configurable, as in Node.
	for _, alias := range []string{"pseudoRandomBytes", "prng", "rng"} {
		exports.SetFuncHidden(alias, func(this js.Object, args []js.Value) (js.Value, error) {
			return m.randomBytes(args)
		})
	}

	constants := e.JS.Obj()
	constants.SetNum("RSA_PKCS1_PADDING", 1)
	constants.SetNum("RSA_SSLV23_PADDING", 2)
	constants.SetNum("RSA_NO_PADDING", 3)
	constants.SetNum("RSA_PKCS1_OAEP_PADDING", 4)
	constants.SetNum("POINT_CONVERSION_COMPRESSED", 2)
	constants.SetNum("POINT_CONVERSION_UNCOMPRESSED", 4)
	constants.SetNum("ENGINE_METHOD_ALL", 0xFFFF)
	constants.SetNum("DH_CHECK_P_NOT_PRIME", dhCheckNotPrime)
	constants.SetNum("DH_CHECK_P_NOT_SAFE_PRIME", dhCheckNotSafePrime)
	constants.SetNum("DH_UNABLE_TO_CHECK_GENERATOR", dhUnableToCheckGenerator)
	constants.SetNum("DH_NOT_SUITABLE_GENERATOR", dhNotSuitableGenerator)
	exports.Set("constants", constants.V())

	m.installX509(exports)
	m.installAsym(exports)
	m.installExtra(exports)

	subtle := e.JS.Obj()
	subtle.SetFunc("digest", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.subtleDigest(args)
	})
	subtle.SetFunc("importKey", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.subtleImportKey(args)
	})
	subtle.SetFunc("encrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.subtleCipher(args, false)
	})
	subtle.SetFunc("decrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.subtleCipher(args, true)
	})
	webcrypto := e.JS.Obj()
	webcrypto.SetFunc("getRandomValues", getRandomValues)
	webcrypto.SetFunc("randomUUID", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.randomUUID(args)
	})
	webcrypto.Set("subtle", subtle.V())
	exports.Set("webcrypto", webcrypto.V())
	exports.Set("subtle", subtle.V())
	// Node also exposes Web Crypto as the global `crypto`; libraries such as
	// uuid call crypto.randomUUID without requiring the module.
	e.Global().Set("crypto", webcrypto.V())
	// Node brands the Web Crypto objects: crypto is a Crypto instance and
	// crypto.subtle a SubtleCrypto instance. The methods stay own properties,
	// so the prototype swap only fixes instanceof and constructor identity.
	if init, err := e.JS.Eval(`(function () {
  class Crypto {}
  class SubtleCrypto {}
  Object.setPrototypeOf(globalThis.crypto, Crypto.prototype);
  Object.setPrototypeOf(globalThis.crypto.subtle, SubtleCrypto.prototype);
  globalThis.Crypto = Crypto;
  globalThis.SubtleCrypto = SubtleCrypto;
})`, "[gode:crypto identity]"); err == nil && init.Valid() && init.IsFunction() {
		_, _ = init.Object().Call(js.Object{})
	}

	return exports
}

// class exposes a Node-style class whose public constructor is illegal. The
// exported value exists so `instanceof crypto.Hash` works, exactly as in Node.
func (m *module) class(name string, proto js.Object) js.Object {
	holder := m.e.JS.Obj()
	holder.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
		return js.Value{}, js.Throw("TypeError", "Illegal constructor")
	})
	ctor := holder.Get(name).Object()
	ctor.Set("prototype", proto.V())
	proto.Set("constructor", ctor.V())
	return ctor
}

// classWithCtor builds a Node-style class whose constructor runs fn. A JS
// function that returns an object produces it for both `C(...)` and
// `new C(...)`, which is exactly how Node's Sign/Verify and KeyObject behave.
func (m *module) classWithCtor(name string, proto js.Object, fn func(args []js.Value) (js.Value, error)) js.Object {
	holder := m.e.JS.Obj()
	holder.SetFunc("impl", func(this js.Object, args []js.Value) (js.Value, error) {
		return fn(args)
	})
	// A Go function object is not constructible in JavaScriptCore, so the
	// public class is a JS function whose prototype is the Go-side proto and
	// whose body delegates to the Go implementation. It stays callable without
	// `new` as well, which is how Node's Hash, Hmac, Sign and Verify behave.
	makeClass, err := m.e.JS.Eval(`(function (impl, proto, name) {
  var Cls = function (...args) { return impl.apply(this, args); };
  Object.defineProperty(Cls, 'name', { value: name, configurable: true });
  Object.defineProperty(Cls, 'length', { value: 0, configurable: true });
  Cls.prototype = proto;
  Object.defineProperty(proto, 'constructor', {
    value: Cls, writable: true, configurable: true, enumerable: false
  });
  return Cls;
})`, "[gode:crypto/class]")
	if err != nil {
		panic("godejs crypto: " + err.Error())
	}
	built, callErr := makeClass.Object().Call(js.Object{}, holder.Get("impl"), proto.V(), m.e.JS.Str(name))
	if callErr != nil {
		panic("godejs crypto: " + callErr.Error())
	}
	return built.Object()
}

// invalidThis mirrors Node's ERR_INVALID_THIS, which brand checks report when
// a prototype method is invoked on the wrong receiver.
func invalidThis(what string) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_THIS",
		fmt.Sprintf(`Value of "this" must be of type %s`, what))
}

func (m *module) instance(st *hashState, proto js.Object) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.states[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeCryptoId", m.e.JS.Num(float64(id)))
	o.SetProto(proto)
	return o
}

func (m *module) stateOf(this js.Object) (*hashState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[this.Get("$godeCryptoId").Int()]
	if st == nil {
		return nil, invalidThis("Hash")
	}
	return st, nil
}

func (m *module) update(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.stateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, hashFinalized()
	}
	v := arg(args, 0)
	var data []byte
	if v.IsString() {
		enc := "utf8"
		if e := arg(args, 1); e.Valid() && !e.IsUndefined() && !e.IsNull() {
			enc = e.Str()
		}
		data, err = decodeInput(v.Str(), enc)
		if err != nil {
			return js.Value{}, err
		}
	} else if b, ok := m.bytesOf(v); ok {
		data = b
	} else {
		return js.Value{}, typeErr(`The "data" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
	}
	st.buf = append(st.buf, data...)
	return this.V(), nil
}

func (m *module) digest(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.stateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, hashFinalized()
	}
	st.done = true
	sum, err := m.sum(st)
	if err != nil {
		return js.Value{}, err
	}
	enc, err := m.encodingArg(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	if enc == "" || enc == "buffer" {
		return m.bytesValue(sum), nil
	}
	return m.encodeDigest(sum, enc)
}

func (m *module) copyHash(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.stateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_HASH_FINALIZED", "Digest already called")
	}
	cp := &hashState{
		algo:   st.algo,
		key:    append([]byte{}, st.key...),
		mac:    st.mac,
		buf:    append([]byte{}, st.buf...),
		xof:    st.xof,
		outLen: st.outLen,
	}
	// copy() without an explicit outputLength returns an XOF hash to its
	// default length, even when the source was created with a short one.
	if cp.xof != "" {
		if _, def, ok := lookupXOF(cp.xof); ok {
			cp.outLen = def
		}
	}
	if opts := arg(args, 0); opts.IsObject() {
		if v := opts.Object().Get("outputLength"); v.Valid() && !v.IsUndefined() {
			n, err := m.outputLengthArg(v)
			if err != nil {
				return js.Value{}, err
			}
			if cp.xof == "" {
				size, ok := m.digestSize(cp.algo)
				if !ok || n != size {
					return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_EVP_NOT_XOF_OR_INVALID_LENGTH", "invalid digest length")
				}
			} else {
				cp.outLen = n
			}
		}
	}
	proto := m.hashProto
	if st.mac {
		proto = m.macProto
	}
	return m.instance(cp, proto).V(), nil
}

func (m *module) sum(st *hashState) ([]byte, error) {
	if st.xof != "" {
		var sh *sha3.SHAKE
		if st.xof == "shake128" {
			sh = sha3.NewSHAKE128()
		} else {
			sh = sha3.NewSHAKE256()
		}
		sh.Write(st.buf)
		out := make([]byte, st.outLen)
		if len(out) > 0 {
			if _, err := sh.Read(out); err != nil {
				return nil, js.Throw("Error", err.Error())
			}
		}
		return out, nil
	}
	fn, _, ok := lookup(st.algo)
	if !ok {
		return nil, invalidDigest()
	}
	if st.mac {
		h := hmac.New(fn, st.key)
		h.Write(st.buf)
		return h.Sum(nil), nil
	}
	h := fn()
	h.Write(st.buf)
	return h.Sum(nil), nil
}

func (m *module) encodeDigest(sum []byte, enc string) (js.Value, error) {
	switch strings.ToLower(enc) {
	case "buffer":
		return m.bytesValue(sum), nil
	case "hex":
		return m.e.JS.Str(hex.EncodeToString(sum)), nil
	case "base64":
		return m.e.JS.Str(base64.StdEncoding.EncodeToString(sum)), nil
	case "base64url":
		return m.e.JS.Str(base64.RawURLEncoding.EncodeToString(sum)), nil
	case "latin1", "binary":
		return m.e.JS.Str(latin1String(sum)), nil
	case "utf8", "utf-8":
		return m.e.JS.Str(string(sum)), nil
	case "ucs2", "ucs-2", "utf16le", "utf-16le":
		return m.e.JS.Str(utf16leString(sum)), nil
	}
	return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", enc))
}

// macFinal is Mac#final(): always a Buffer, and finalized only once.
func (m *module) macFinal(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.stateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, hashFinalized()
	}
	st.done = true
	sum, err := m.sum(st)
	if err != nil {
		return js.Value{}, err
	}
	return m.bytesValue(sum), nil
}

// knownEncoding reports whether Buffer accepts the encoding name.
func knownEncoding(enc string) bool {
	switch strings.ToLower(enc) {
	case "buffer", "hex", "base64", "base64url", "latin1", "binary", "utf8", "utf-8",
		"ucs2", "ucs-2", "utf16le", "utf-16le", "ascii":
		return true
	}
	return false
}

func (m *module) keyArg(v js.Value) ([]byte, error) {
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	if v.IsObject() {
		// createHmac/createHash accept KeyObjects and CryptoKeys as key
		// material, not just raw bytes.
		if _, ok := keyIDOf(v); ok {
			st, err := m.keyStateOf(v.Object())
			if err != nil {
				return nil, err
			}
			if st.kind != "secret" {
				return nil, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_KEYTYPE", "Invalid key type")
			}
			return st.secretBytes(), nil
		}
		if id := v.Object().Get("$godeCryptoKeyId"); id.Valid() && id.IsNumber() {
			return m.cryptoKeyBytes(v)
		}
	}
	return nil, typeErr(`The "key" argument must be of type string or an instance of Buffer, TypedArray, DataView, KeyObject, or CryptoKey. Received %s`, describe(v))
}

func (m *module) dataArg(v js.Value) ([]byte, error) {
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	return nil, typeErr(`The "data" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
}

func (m *module) randomBytes(args []js.Value) (js.Value, error) {
	n, err := randomSizeArg(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	var cb js.Value
	if c := arg(args, 1); c.Valid() && !c.IsUndefined() {
		if !c.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(c))
		}
		cb = c
	}
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	buf := m.bytesValue(data)
	if cb.Valid() {
		m.e.NextTick(cb, m.e.JS.Null(), buf)
		return js.Value{}, nil
	}
	return buf, nil
}

func (m *module) randomFill(args []js.Value, async bool) (js.Value, error) {
	buf := arg(args, 0)
	if !buf.IsObject() {
		return js.Value{}, typeErr(`The "buf" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(buf))
	}
	tag := buf.Tag()
	switch tag {
	case "ArrayBuffer", "SharedArrayBuffer":
		// Filled as raw bytes; the caller gets the same object back.
	default:
		if !isBufferViewTag(tag) && tag != "DataView" {
			return js.Value{}, typeErr(`The "buf" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(buf))
		}
	}
	byteLength := buf.Object().Get("byteLength").Int()
	view := buf
	if tag == "ArrayBuffer" || tag == "SharedArrayBuffer" {
		v, err := m.newU8(buf, 0, byteLength)
		if err != nil {
			return js.Value{}, err
		}
		view = v
	}
	idx := 1
	offset, size := 0, byteLength
	if idx < len(args) && !args[idx].IsFunction() {
		off, err := fillIntArg(args[idx], "offset", byteLength)
		if err != nil {
			return js.Value{}, err
		}
		offset = off
		size = byteLength - offset
		idx++
	}
	if idx < len(args) && !args[idx].IsFunction() {
		sz, err := fillIntArg(args[idx], "size", maxRandomBytes)
		if err != nil {
			return js.Value{}, err
		}
		size = sz
		idx++
	}
	var cb js.Value
	if idx < len(args) && args[idx].IsFunction() {
		cb = args[idx]
	}
	if async && !cb.Valid() {
		return js.Value{}, typeErr(`The "cb" argument must be of type function. Received %s`, describe(arg(args, idx)))
	}
	if offset < 0 || offset > byteLength {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "offset" is out of range. It must be >= 0 && <= %d. Received %d`, byteLength, offset))
	}
	if size < 0 || size > maxRandomBytes {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "size" is out of range. It must be >= 0 && <= %d. Received %d`, maxRandomBytes, size))
	}
	if offset+size > byteLength {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "size + offset" is out of range. It must be <= %d. Received %d`, byteLength, offset+size))
	}
	if size > 0 {
		data := make([]byte, size)
		if _, err := rand.Read(data); err != nil {
			return js.Value{}, js.Throw("Error", err.Error())
		}
		if err := m.fillRegion(view, offset, data); err != nil {
			return js.Value{}, err
		}
	}
	if async {
		m.e.NextTick(cb, m.e.JS.Null(), buf)
		return js.Value{}, nil
	}
	return buf, nil
}

// randomSizeArg accepts the fractional sizes randomBytes truncates.
func randomSizeArg(v js.Value) (int, error) {
	if !v.Valid() || !v.IsNumber() {
		return 0, typeErr(`The "size" argument must be of type number.%s`, invalidArgHelper(v))
	}
	f := math.Floor(v.Float())
	if math.IsNaN(f) || f < 0 || f > maxRandomBytes {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "size" is out of range. It must be >= 0 && <= %d. Received %s`, maxRandomBytes, numberString(v.Float())))
	}
	return int(f), nil
}

func fillIntArg(v js.Value, name string, max int) (int, error) {
	if !v.IsNumber() {
		return 0, typeErr(`The "%s" argument must be of type number.%s`, name, invalidArgHelper(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) || f < 0 || f > float64(max) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "%s" is out of range. It must be >= 0 && <= %d. Received %s`, name, max, numberString(f)))
	}
	return int(f), nil
}

func numberString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func firstArgs(args []js.Value, n int) []js.Value {
	if len(args) < n {
		return args
	}
	return args[:n]
}

func invalidArgHelper(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return " Received undefined"
	}
	if v.IsNull() {
		return " Received null"
	}
	if v.IsFunction() {
		return " Received function "
	}
	if v.IsString() {
		return fmt.Sprintf(" Received type string ('%s')", v.Str())
	}
	if v.IsBool() {
		return fmt.Sprintf(" Received type boolean (%v)", v.Bool())
	}
	if v.IsNumber() {
		return " Received type number (" + numberString(v.Float()) + ")"
	}
	if v.IsObject() {
		tag := v.Tag()
		if tag == "" {
			tag = "Object"
		}
		return " Received an instance of " + tag
	}
	return ""
}

func (m *module) fillRegion(v js.Value, offset int, data []byte) error {
	base := v.Object().Get("byteOffset").Int()
	buf := v.Object().Get("buffer")
	view, err := m.newU8(buf, base+offset, len(data))
	if err != nil {
		return err
	}
	return m.setBytes(view, data)
}

func (m *module) setBytes(view js.Value, data []byte) error {
	set := m.e.Global().Get("Uint8Array").Object().Get("prototype").Object().Get("set")
	_, err := set.Object().Call(view.Object(), m.e.JS.Bytes(data).V(), m.e.JS.Num(0))
	return err
}

func (m *module) getRandomValues(args []js.Value) (js.Value, error) {
	v := arg(args, 0)
	if !v.IsObject() || !isTypedArrayTag(v.Tag()) {
		return js.Value{}, typeErr(`The "typedArray" argument must be an instance of ArrayBufferView. Received %s`, describe(v))
	}
	n := v.Object().Get("byteLength").Int()
	if n > maxEntropy {
		return js.Value{}, js.ThrowCode("QuotaExceededError", "22",
			fmt.Sprintf("The ArrayBufferView's byte length (%d) exceeds the number of bytes of entropy available via this API (65536)", n))
	}
	if n > 0 {
		data := make([]byte, n)
		if _, err := rand.Read(data); err != nil {
			return js.Value{}, js.Throw("Error", err.Error())
		}
		if err := m.fillRegion(v, 0, data); err != nil {
			return js.Value{}, err
		}
	}
	return v, nil
}

// maxRandomIntRange is Node's 2^48-1 ceiling on the drawable range.
const maxRandomIntRange = 1<<48 - 1

func (m *module) randomInt(args []js.Value) (js.Value, error) {
	vars := args
	var cb js.Value
	if n := len(vars); n > 1 && vars[n-1].IsFunction() {
		cb = vars[n-1]
		vars = vars[:n-1]
	} else if n := len(vars); n == 3 || (n > 1 && !vars[n-1].IsNumber()) {
		return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(vars[n-1]))
	}
	min, max := int64(0), int64(0)
	switch len(vars) {
	case 1:
		v, err := safeIntArg(vars[0], "max")
		if err != nil {
			return js.Value{}, err
		}
		max = v
	case 2:
		lo, err := safeIntArg(vars[0], "min")
		if err != nil {
			return js.Value{}, err
		}
		hi, err := safeIntArg(vars[1], "max")
		if err != nil {
			return js.Value{}, err
		}
		min, max = lo, hi
	default:
		return js.Value{}, typeErr(`The "max" argument must be a safe integer. Received %s`, describe(arg(vars, 0)))
	}
	if max <= min {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf("The value of \"max\" is out of range. It must be greater than the value of \"min\" (%d). Received %d", min, max))
	}
	rangeSize := max - min
	if rangeSize > maxRandomIntRange {
		field := "max - min"
		if len(vars) == 1 {
			field = "max"
		}
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf("The value of \"%s\" is out of range. It must be <= %d. Received %s", field, int64(maxRandomIntRange), formatInt(rangeSize)))
	}
	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	value := n.Int64() + min
	if cb.Valid() {
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Num(float64(value)))
		return js.Value{}, nil
	}
	return m.e.JS.Num(float64(value)), nil
}

func safeIntArg(v js.Value, name string) (int64, error) {
	if !v.Valid() || !v.IsNumber() {
		return 0, typeErr(`The "%s" argument must be a safe integer.%s`, name, invalidArgHelper(v))
	}
	f := v.Float()
	if math.Trunc(f) != f || math.Abs(f) > 1<<53-1 {
		return 0, typeErr(`The "%s" argument must be a safe integer.%s`, name, invalidArgHelper(v))
	}
	return int64(f), nil
}

// formatInt matches the digit grouping Node uses in range errors.
func formatInt(v int64) string {
	s := fmt.Sprintf("%d", v)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "_" + s[i:]
	}
	return s
}

// uuidOptions validates the shared randomUUID/randomUUIDv7 options bag.
func uuidOptions(args []js.Value) error {
	opts := arg(args, 0)
	if !opts.Valid() || opts.IsUndefined() {
		return nil
	}
	if !opts.IsObject() || opts.IsNull() {
		return typeErr(`The "options" argument must be of type object. Received %s`, describe(opts))
	}
	if v := opts.Object().Get("disableEntropyCache"); v.Valid() && !v.IsUndefined() {
		if !v.IsBool() {
			return typeErr(`The "options.disableEntropyCache" argument must be of type boolean. Received %s`, describe(v))
		}
	}
	return nil
}

func (m *module) randomUUID(args []js.Value) (js.Value, error) {
	if err := uuidOptions(args); err != nil {
		return js.Value{}, err
	}

	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	const digits = "0123456789abcdef"
	out := make([]byte, 36)
	pos := 0
	for i, c := range b {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			out[pos] = '-'
			pos++
		}
		out[pos] = digits[c>>4]
		out[pos+1] = digits[c&0x0f]
		pos += 2
	}
	return m.e.JS.Str(string(out)), nil
}

func (m *module) subtleDigest(args []js.Value) (js.Value, error) {
	alg := arg(args, 0)
	name := ""
	if alg.IsString() {
		name = alg.Str()
	} else if alg.IsObject() {
		name = alg.Object().Get("name").Str()
	}
	fn, _, ok := lookup(name)
	if !ok {
		return m.e.JS.Promise(js.Value{}, js.Throw("TypeError", "Unrecognized algorithm name")), nil
	}
	data, ok := m.bytesOf(arg(args, 1))
	if !ok {
		return m.e.JS.Promise(js.Value{}, typeErr(`The "data" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 1)))), nil
	}
	h := fn()
	h.Write(data)
	sum := h.Sum(nil)
	ab := m.e.JS.Bytes(sum).Get("buffer")
	return m.e.JS.Promise(ab, nil), nil
}

func (m *module) newCryptoKey(data []byte) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.cryptoKeys[id] = append([]byte{}, data...)
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeCryptoKeyId", m.e.JS.Num(float64(id)))
	o.SetProto(m.cryptoKeyProto)
	return o
}

func (m *module) cryptoKeyBytes(v js.Value) ([]byte, error) {
	if !v.IsObject() {
		return nil, typeErr(`The "key" argument must be a CryptoKey. Received %s`, describe(v))
	}
	id := v.Object().Get("$godeCryptoKeyId")
	if !id.Valid() || !id.IsNumber() {
		return nil, typeErr(`The "key" argument must be a CryptoKey. Received %s`, describe(v))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	data := m.cryptoKeys[id.Int()]
	if data == nil {
		return nil, typeErr(`The "key" argument must be a CryptoKey. Received %s`, describe(v))
	}
	return data, nil
}

func (m *module) subtleImportKey(args []js.Value) (js.Value, error) {
	format := arg(args, 0)
	if !format.IsString() {
		return m.e.JS.Promise(js.Value{}, typeErr(`The "format" argument must be of type string. Received %s`, describe(format))), nil
	}
	name := ""
	alg := arg(args, 2)
	if alg.IsObject() {
		name = alg.Object().Get("name").Str()
	} else if alg.IsString() {
		name = alg.Str()
	}
	upper := strings.ToUpper(strings.ReplaceAll(name, "-", ""))
	if upper != "AESGCM" && upper != "AESCTR" && upper != "AESCBC" {
		return m.e.JS.Promise(js.Value{}, js.Throw("TypeError", "Unrecognized algorithm name")), nil
	}
	if strings.ToLower(format.Str()) != "raw" {
		return m.e.JS.Promise(js.Value{}, js.Throw("TypeError", "Only raw keys are supported")), nil
	}
	data, ok := m.bytesOf(arg(args, 1))
	if !ok {
		return m.e.JS.Promise(js.Value{}, typeErr(`The "keyData" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 1)))), nil
	}
	if len(data) != 16 && len(data) != 24 && len(data) != 32 {
		return m.e.JS.Promise(js.Value{}, js.Throw("DataError", "Invalid key length")), nil
	}
	return m.e.JS.Promise(m.newCryptoKey(data).V(), nil), nil
}

func (m *module) subtleCipher(args []js.Value, decrypt bool) (js.Value, error) {
	alg := arg(args, 0)
	if !alg.IsObject() {
		return m.e.JS.Promise(js.Value{}, typeErr(`The "algorithm" argument must be of type object. Received %s`, describe(alg))), nil
	}
	params := alg.Object()
	name := params.Get("name").Str()
	if strings.ToUpper(strings.ReplaceAll(name, "-", "")) != "AESGCM" {
		return m.e.JS.Promise(js.Value{}, js.Throw("NotSupportedError", "Unsupported algorithm")), nil
	}
	key, err := m.cryptoKeyBytes(arg(args, 1))
	if err != nil {
		return m.e.JS.Promise(js.Value{}, err), nil
	}
	data, ok := m.bytesOf(arg(args, 2))
	if !ok {
		return m.e.JS.Promise(js.Value{}, typeErr(`The "data" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(arg(args, 2)))), nil
	}
	iv, _ := m.bytesOf(params.Get("iv"))
	if len(iv) == 0 {
		return m.e.JS.Promise(js.Value{}, js.Throw("TypeError", "The AES-GCM algorithm.iv parameter is required")), nil
	}
	var aad []byte
	if v := params.Get("additionalData"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		aad, _ = m.bytesOf(v)
	}
	tagBits := 128
	if v := params.Get("tagLength"); v.Valid() && v.IsNumber() {
		tagBits = int(v.Float())
	}
	tagLen := tagBits / 8
	if tagLen < 12 || tagLen > 16 {
		return m.e.JS.Promise(js.Value{}, js.Throw("OperationError", "Invalid tag length")), nil
	}
	if tagLen != 16 && len(iv) != 12 {
		return m.e.JS.Promise(js.Value{}, js.Throw("OperationError", "Invalid nonce length")), nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return m.e.JS.Promise(js.Value{}, js.Throw("Error", err.Error())), nil
	}
	var aead stdcipher.AEAD
	if tagLen == 16 {
		aead, err = stdcipher.NewGCMWithNonceSize(block, len(iv))
	} else {
		aead, err = stdcipher.NewGCMWithTagSize(block, tagLen)
	}
	if err != nil {
		return m.e.JS.Promise(js.Value{}, js.Throw("Error", err.Error())), nil
	}
	if decrypt {
		plain, err := aead.Open(nil, iv, data, aad)
		if err != nil {
			return m.e.JS.Promise(js.Value{}, js.Throw("OperationError", "The operation failed for an operation-specific reason")), nil
		}
		return m.e.JS.Promise(m.e.JS.Bytes(plain).Get("buffer"), nil), nil
	}
	sealed := aead.Seal(nil, iv, data, aad)
	return m.e.JS.Promise(m.e.JS.Bytes(sealed).Get("buffer"), nil), nil
}

// bytesValue returns a real Buffer when the buffer builtin is installed and a
// plain Uint8Array otherwise (unit tests install modules in isolation).
func (m *module) bytesValue(b []byte) js.Value {
	g := m.e.Global().Get("Buffer")
	if g.IsFunction() {
		from := g.Object().Get("from")
		if from.IsFunction() {
			if out, err := from.Object().Call(g.Object(), m.e.JS.Bytes(b).V()); err == nil && out.Valid() {
				return out
			}
		}
	}
	return m.e.JS.Bytes(b).V()
}

func (m *module) bytesOf(v js.Value) ([]byte, bool) {
	if b, ok := v.Bytes(); ok {
		return b, true
	}
	switch v.Tag() {
	case "DataView":
		view, err := m.u8View(v)
		if err != nil {
			return nil, false
		}
		return view.Bytes()
	case "ArrayBuffer":
		n := v.Object().Get("byteLength").Int()
		out, err := m.newU8(v, 0, n)
		if err != nil {
			return nil, false
		}
		return out.Bytes()
	}
	return nil, false
}

func (m *module) u8View(v js.Value) (js.Value, error) {
	buf := v.Object().Get("buffer")
	off := v.Object().Get("byteOffset").Int()
	n := v.Object().Get("byteLength").Int()
	return m.newU8(buf, off, n)
}

// newU8 constructs a Uint8Array view. Calling the constructor as a function
// is a TypeError in JavaScriptCore, so use Reflect.construct.
func (m *module) newU8(buf js.Value, offset, length int) (js.Value, error) {
	reflect := m.e.Global().Get("Reflect")
	if !reflect.IsObject() {
		return js.Value{}, typeErr("Reflect is unavailable")
	}
	construct := reflect.Object().Get("construct")
	if !construct.IsFunction() {
		return js.Value{}, typeErr("Reflect.construct is unavailable")
	}
	ctor := m.e.Global().Get("Uint8Array")
	args := m.e.JS.Arr([]js.Value{
		buf,
		m.e.JS.Num(float64(offset)),
		m.e.JS.Num(float64(length)),
	})
	return construct.Object().Call(js.Object{}, ctor, args.V())
}

func lookup(algo string) (func() stdhash.Hash, string, bool) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(algo), "-", "")) {
	case "dss1":
		// Node keeps the historical DSS1 alias for SHA-1.
		return sha1.New, "sha1", true
	case "md5":
		return md5.New, "md5", true
	case "sha1":
		return sha1.New, "sha1", true
	case "sha224":
		return sha256.New224, "sha224", true
	case "sha256":
		return sha256.New, "sha256", true
	case "sha384":
		return sha512.New384, "sha384", true
	case "sha512":
		return sha512.New, "sha512", true
	case "sha512224":
		return sha512.New512_224, "sha512-224", true
	case "sha512256":
		return sha512.New512_256, "sha512-256", true
	case "sha3256":
		return func() stdhash.Hash { return sha3.New256() }, "sha3-256", true
	case "sha3384":
		return func() stdhash.Hash { return sha3.New384() }, "sha3-384", true
	case "sha3512":
		return func() stdhash.Hash { return sha3.New512() }, "sha3-512", true
	}
	return nil, "", false
}

func decodeInput(s, enc string) ([]byte, error) {
	switch strings.ToLower(enc) {
	case "", "utf8", "utf-8":
		return []byte(s), nil
	case "latin1", "binary":
		return latin1Bytes(s), nil
	case "ascii":
		out := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			out[i] = s[i] & 0x7f
		}
		return out, nil
	case "hex":
		out := make([]byte, 0, len(s)/2)
		for i := 0; i+1 < len(s); i += 2 {
			hi, ok1 := hexVal(s[i])
			lo, ok2 := hexVal(s[i+1])
			if !ok1 || !ok2 {
				break
			}
			out = append(out, hi<<4|lo)
		}
		return out, nil
	case "base64":
		return base64.StdEncoding.DecodeString(padBase64(s))
	case "base64url":
		return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	}
	return nil, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", enc))
}

func padBase64(s string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '/', r == '=':
			return r
		}
		return -1
	}, s)
	if pad := len(clean) % 4; pad != 0 {
		clean += strings.Repeat("=", 4-pad)
	}
	return clean
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

func latin1String(b []byte) string {
	out := make([]rune, len(b))
	for i, c := range b {
		out[i] = rune(c)
	}
	return string(out)
}

func latin1Bytes(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		out = append(out, byte(r))
	}
	return out
}

func sizeArg(v js.Value, name string, max int) (int, error) {
	if !v.Valid() || !v.IsNumber() {
		return 0, typeErr(`The "%s" argument must be of type number. Received %s`, name, describe(v))
	}
	f := v.Float()
	if f < 0 || math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "%s" is out of range. It must be an integer. Received %v`, name, f))
	}
	if f > float64(max) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "%s" is out of range. It must be >= 0 && <= %d. Received %v`, name, max, f))
	}
	return int(f), nil
}

// isTypedArrayTag lists the views WebCrypto getRandomValues accepts. Float
// views are excluded because Node rejects them there.
func isTypedArrayTag(tag string) bool {
	switch tag {
	case "Int8Array", "Uint8Array", "Uint8ClampedArray",
		"Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
		"BigInt64Array", "BigUint64Array":
		return true
	}
	return false
}

// isBufferViewTag lists every typed array Node accepts for randomFill.
func isBufferViewTag(tag string) bool {
	return isTypedArrayTag(tag) || tag == "Float32Array" || tag == "Float64Array"
}

func invalidDigest() error {
	return js.ThrowCode("Error", "ERR_CRYPTO_INVALID_DIGEST", "Digest method not supported")
}

func hashFinalized() error {
	return js.ThrowCode("Error", "ERR_CRYPTO_HASH_FINALIZED", "Digest already called")
}

// utf16leString decodes bytes the way Buffer#toString('utf16le') does.
func utf16leString(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		runes = append(runes, rune(uint16(b[i])|uint16(b[i+1])<<8))
	}
	return string(runes)
}

func (m *module) encodingArg(v js.Value) (string, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return "", nil
	}
	if v.IsString() {
		return v.Str(), nil
	}
	if v.IsObject() || v.IsFunction() {
		ts := v.Object().Get("toString")
		if ts.IsFunction() {
			out, err := ts.Object().Call(v.Object())
			if err != nil {
				return "", err
			}
			if out.IsString() {
				return out.Str(), nil
			}
		}
	}
	return "", nil
}

// newHash builds a Hash instance. It handles SHAKE output lengths and the
// constructor's argument validation, which createHash and crypto.Hash share.
func (m *module) newHash(args []js.Value) (js.Value, error) {
	algo := arg(args, 0)
	if !algo.IsString() {
		return js.Value{}, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(algo))
	}
	st := &hashState{}
	if xof, def, ok := lookupXOF(algo.Str()); ok {
		st.algo = xof
		st.xof = xof
		st.outLen = def
	} else {
		_, canonical, ok := lookup(algo.Str())
		if !ok {
			return js.Value{}, invalidDigest()
		}
		st.algo = canonical
	}
	if opts := arg(args, 1); opts.IsObject() {
		if v := opts.Object().Get("outputLength"); v.Valid() && !v.IsUndefined() {
			n, err := m.outputLengthArg(v)
			if err != nil {
				return js.Value{}, err
			}
			if st.xof != "" {
				st.outLen = n
			} else {
				size, ok := m.digestSize(st.algo)
				if !ok || n != size {
					return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_EVP_NOT_XOF_OR_INVALID_LENGTH", "invalid digest length")
				}
			}
		}
	}
	return m.instance(st, m.hashProto).V(), nil
}

func lookupXOF(algo string) (string, int, bool) {
	switch strings.ToLower(strings.TrimSpace(algo)) {
	case "shake128", "shake-128":
		return "shake128", 16, true
	case "shake256", "shake-256":
		return "shake256", 32, true
	}
	return "", 0, false
}

func (m *module) digestSize(algo string) (int, bool) {
	fn, _, ok := lookup(algo)
	if !ok {
		return 0, false
	}
	return fn().Size(), true
}

func (m *module) outputLengthArg(v js.Value) (int, error) {
	if !v.IsNumber() {
		return 0, typeErr(`The "options.outputLength" argument must be of type number. Received %s`, describe(v))
	}
	f := v.Float()
	if f < 0 || math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "options.outputLength" is out of range. It must be a non-negative integer. Received %v`, f))
	}
	if f > 1<<31-1 {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "options.outputLength" is out of range. It must be <= 2147483647. Received %v`, f))
	}
	return int(f), nil
}

func typeErr(format string, a ...any) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", fmt.Sprintf(format, a...))
}

func describe(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	if v.IsString() {
		return "type string ('" + v.Str() + "')"
	}
	if v.IsNumber() {
		return fmt.Sprintf("type number (%v)", v.Float())
	}
	if v.IsBool() {
		return fmt.Sprintf("type boolean (%v)", v.Bool())
	}
	tag := v.Tag()
	if tag == "" {
		tag = "Object"
	}
	return "an instance of " + tag
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func (m *module) installKeyProto() {
	m.keyProto.SetFuncHidden("$godeType", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Str(st.kind), nil
	})
	m.keyProto.SetFuncHidden("$godeAsymType", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if st.kind == "secret" {
			return js.Value{}, invalidThis("AsymmetricKeyObject")
		}
		return m.e.JS.Str(st.typ), nil
	})
	m.keyProto.SetFuncHidden("$godeDetails", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if st.kind == "secret" {
			return js.Value{}, invalidThis("AsymmetricKeyObject")
		}
		return m.keyDetails(st), nil
	})
	m.keyProto.SetFuncHidden("$godeSymSize", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if st.kind != "secret" {
			return js.Value{}, invalidThis("SecretKeyObject")
		}
		return m.e.JS.Num(float64(len(st.secretBytes()))), nil
	})
	m.keyProto.SetFuncHidden("$godeExport", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		hasOpts := false
		var opts js.Object
		if len(args) > 0 && args[0].IsObject() {
			opts, hasOpts = args[0].Object(), true
		}
		return m.exportKey(st, hasOpts, opts)
	})
	m.keyProto.SetFuncHidden("$godeEquals", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		otherV := arg(args, 0)
		var other *keyState
		if otherV.IsObject() {
			if _, ok := keyIDOf(otherV); ok {
				other, err = m.keyStateOf(otherV.Object())
			} else {
				other, err = m.loadKey(otherV, false)
			}
		} else {
			other, err = m.loadKey(otherV, false)
		}
		if err != nil {
			return js.Value{}, err
		}
		if other.kind == "secret" || st.kind == "secret" {
			return m.e.JS.Bool(bytes.Equal(st.secretBytes(), other.secretBytes())), nil
		}
		a, err := m.rawKey(st)
		if err != nil {
			return js.Value{}, err
		}
		b, err := m.rawKey(other)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(bytes.Equal(a, b)), nil
	})
	prelude, err := m.e.JS.Eval(`(function(base, secret, asym, pub, priv) {
  // The hidden host functions are looked up on base and invoked with the
  // receiver as its this value, so a foreign receiver ({}) still reaches the Go brand
  // check and reports ERR_INVALID_THIS instead of a plain TypeError.
  Object.defineProperty(base, Symbol.toStringTag, { value: 'KeyObject', configurable: true });
  // The numeric key id lives in a WeakMap instead of an own property, so a
  // KeyObject exposes no own names or symbols (Node keeps it in a native slot).
  var ids = new WeakMap();
  Object.defineProperty(base, '$godeAttachId', {
    value: function(id) { ids.set(this, id); return this; }, configurable: true
  });
  Object.defineProperty(base, '$godeOwnId', {
    value: function() { return ids.get(this); }, configurable: true
  });
  Object.defineProperty(base, 'type', { configurable: true, enumerable: true, get: function() { return base.$godeType.call(this); } });
  Object.defineProperty(secret, 'symmetricKeySize', { configurable: true, enumerable: true, get: function() { return base.$godeSymSize.call(this); } });
  Object.defineProperty(asym, 'asymmetricKeyType', { configurable: true, enumerable: true, get: function() { return base.$godeAsymType.call(this); } });
  Object.defineProperty(asym, 'asymmetricKeyDetails', { configurable: true, enumerable: true, get: function() { return base.$godeDetails.call(this); } });
  base['export'] = function(options) { return base.$godeExport.call(this, options); };
  base.equals = function(other) { return base.$godeEquals.call(this, other); };
  Object.setPrototypeOf(secret, base);
  Object.setPrototypeOf(asym, base);
  Object.setPrototypeOf(pub, asym);
  Object.setPrototypeOf(priv, asym);
  return base;
})`, "[gode:crypto/keyobject]")
	if err == nil {
		_, _ = prelude.Object().Call(js.Object{}, m.keyProto.V(), m.secretProto.V(), m.asymProto.V(), m.pubProto.V(), m.privProto.V())
	}
}

func (st *keyState) secretBytes() []byte {
	if b, ok := st.key.([]byte); ok {
		return b
	}
	return nil
}

func (m *module) keyStateOf(this js.Object) (*keyState, error) {
	id, ok := keyIDOf(this.V())
	m.mu.Lock()
	defer m.mu.Unlock()
	if !ok {
		return nil, invalidThis("KeyObject")
	}
	st := m.keys[id]
	if st == nil {
		return nil, invalidThis("KeyObject")
	}
	return st, nil
}

func (m *module) newKeyObject(st *keyState) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.keys[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	proto := m.pubProto
	switch {
	case st.kind == "secret":
		proto = m.secretProto
	case st.kind == "private":
		proto = m.privProto
	}
	o.SetProto(proto)
	if attach := m.keyProto.Get("$godeAttachId"); attach.Valid() && attach.IsFunction() {
		_, _ = attach.Object().Call(o, m.e.JS.Num(float64(id)))
	} else {
		o.SetHidden("$godeKeyId", m.e.JS.Num(float64(id)))
	}
	return o
}

func wrapKey(k any) (*keyState, bool) {
	switch v := k.(type) {
	case *rsa.PrivateKey:
		return &keyState{kind: "private", typ: "rsa", key: v}, true
	case *rsa.PublicKey:
		return &keyState{kind: "public", typ: "rsa", key: v}, true
	case *ecdsa.PrivateKey:
		return &keyState{kind: "private", typ: "ec", key: v, curve: curveName(v.Curve)}, true
	case *ecdsa.PublicKey:
		return &keyState{kind: "public", typ: "ec", key: v, curve: curveName(v.Curve)}, true
	case ed25519.PrivateKey:
		return &keyState{kind: "private", typ: "ed25519", key: v}, true
	case ed25519.PublicKey:
		return &keyState{kind: "public", typ: "ed25519", key: v}, true
	case *ecdh.PrivateKey:
		return &keyState{kind: "private", typ: "x25519", key: v}, true
	case *ecdh.PublicKey:
		return &keyState{kind: "public", typ: "x25519", key: v}, true
	}
	if st, ok := wrapDSA(k); ok {
		return st, true
	}
	return nil, false
}

func (m *module) keyDetails(st *keyState) js.Value {
	if st.kind == "secret" {
		return m.e.JS.Undefined()
	}
	o := m.e.JS.Obj()
	switch st.typ {
	case "rsa", "rsa-pss":
		var pub *rsa.PublicKey
		switch k := st.key.(type) {
		case *rsa.PrivateKey:
			pub = &k.PublicKey
		case *rsa.PublicKey:
			pub = k
		}
		if pub == nil {
			return o.V()
		}
		o.SetNum("modulusLength", float64(pub.N.BitLen()))
		o.Set("publicExponent", m.bigintValue(big.NewInt(int64(pub.E))))
		if st.typ == "rsa-pss" {
			if st.hashAlg != "" {
				o.SetStr("hashAlgorithm", st.hashAlg)
			}
			if st.mgf1HashAlg != "" {
				o.SetStr("mgf1HashAlgorithm", st.mgf1HashAlg)
			}
			// Node reports the effective salt length: the digest size when the
			// caller did not ask for one (RFC 8017 A.2.3).
			salt := st.saltLen
			if salt == 0 && st.hashAlg != "" {
				if fn, _, ok := lookup(st.hashAlg); ok {
					salt = fn().Size()
				}
			}
			if salt != 0 {
				o.SetNum("saltLength", float64(salt))
			}
		}
	case "ec":
		o.SetStr("namedCurve", st.curve)
	case "dsa":
		if modulus, divisor, ok := dsaKeyInfo(st); ok {
			o.SetNum("modulusLength", float64(modulus))
			o.SetNum("divisorLength", float64(divisor))
		}
	}
	return o.V()
}

func (m *module) bigintValue(n *big.Int) js.Value {
	if !m.bigintFn.Valid() {
		return m.e.JS.Num(0)
	}
	out, err := m.bigintFn.Object().Call(js.Object{}, m.e.JS.Str(n.String()))
	if err != nil {
		return m.e.JS.Num(0)
	}
	return out
}

func (m *module) loadKey(v js.Value, private bool) (*keyState, error) {
	if _, ok := keyIDOf(v); ok {
		st, err := m.keyStateOf(v.Object())
		if err != nil {
			return nil, err
		}
		return m.coerceKey(st, private)
	}
	var data []byte
	format := "pem"
	typ := ""
	passphrase := ""
	jwk := js.Object{}
	hasJWK := false
	// A Buffer/TypedArray/DataView argument is raw key material, not an
	// options bag, so it must skip the destructuring path Node uses for
	// { key, format, type, passphrase } objects.
	_, isBinaryKey := m.bytesOf(v)
	if v.IsObject() && !isBinaryKey {
		opts := v.Object()
		keyVal := opts.Get("key")
		if !keyVal.Valid() {
			return nil, typeErr(`The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
		}
		if f := opts.Get("format"); f.IsString() {
			format = strings.ToLower(f.Str())
		}
		if t := opts.Get("type"); t.IsString() {
			typ = strings.ToLower(t.Str())
		}
		if p := opts.Get("passphrase"); p.IsString() {
			passphrase = p.Str()
		}
		v = keyVal
		if format == "jwk" && v.IsObject() {
			jwk, hasJWK = v.Object(), true
		}
		// {key: KeyObject} wrappers carry the passphrase and signature
		// options, so unwrap to the KeyObject rather than treating it as
		// raw key bytes.
		if isKeyObjectValue(v) {
			st, err := m.keyStateOf(v.Object())
			if err != nil {
				return nil, err
			}
			return m.coerceKey(st, private)
		}
	}
	if hasJWK {
		st, err := m.keyFromJWK(jwk)
		if err != nil {
			return nil, err
		}
		return m.coerceKey(st, private)
	}
	if v.IsString() {
		data = []byte(v.Str())
	} else if b, ok := m.bytesOf(v); ok {
		data = b
	} else {
		return nil, typeErr(`The "key" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
	}
	var st *keyState
	var err error
	if format == "der" {
		st, err = parseDER(data, typ)
	} else {
		st, err = parsePEM(data, passphrase)
	}
	if err != nil {
		return nil, err
	}
	return m.coerceKey(st, private)
}

func (m *module) coerceKey(st *keyState, private bool) (*keyState, error) {
	if private {
		if st.kind != "private" {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a private key")
		}
		return st, nil
	}
	if st.kind == "private" {
		return publicState(st)
	}
	return st, nil
}

func publicState(st *keyState) (*keyState, error) {
	switch k := st.key.(type) {
	case *rsa.PrivateKey:
		return &keyState{kind: "public", typ: st.typ, key: &k.PublicKey, hashAlg: st.hashAlg, mgf1HashAlg: st.mgf1HashAlg, saltLen: st.saltLen, padding: st.padding}, nil
	case *ecdsa.PrivateKey:
		return &keyState{kind: "public", typ: "ec", key: &k.PublicKey, curve: st.curve, dsaEncoding: st.dsaEncoding}, nil
	case ed25519.PrivateKey:
		return &keyState{kind: "public", typ: "ed25519", key: k.Public().(ed25519.PublicKey)}, nil
	case *ecdh.PrivateKey:
		return &keyState{kind: "public", typ: "x25519", key: k.PublicKey()}, nil
	}
	if pub, ok := dsaPublicState(st); ok {
		return pub, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a public key")
}

func parsePEM(data []byte, passphrase string) (*keyState, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, js.ThrowCode("Error", "ERR_OSSL_PEM_NO_START_LINE", "error:0E079065:configuration file routines::no start line")
	}
	der := block.Bytes
	if x509.IsEncryptedPEMBlock(block) {
		plain, err := x509.DecryptPEMBlock(block, []byte(passphrase))
		if err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		der = plain
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		k, err := x509.ParsePKCS1PrivateKey(der)
		if err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "private", typ: "rsa", key: k}, nil
	case "RSA PUBLIC KEY":
		k, err := x509.ParsePKCS1PublicKey(der)
		if err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "public", typ: "rsa", key: k}, nil
	case "EC PRIVATE KEY":
		k, err := x509.ParseECPrivateKey(der)
		if err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "private", typ: "ec", key: k, curve: curveName(k.Curve)}, nil
	case "DSA PRIVATE KEY":
		k, ok := dsaParseTraditional(der)
		if !ok {
			return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYTYPE", "Failed to parse key")
		}
		return &keyState{kind: "private", typ: "dsa", key: k}, nil
	}
	return parseDER(der, "")
}

func parseDER(der []byte, typ string) (*keyState, error) {
	try := func(k any, err error) (*keyState, error, bool) {
		if err != nil {
			return nil, nil, false
		}
		st, ok := wrapKey(k)
		if !ok {
			return nil, nil, false
		}
		return st, nil, true
	}
	var attempts []func() (any, error)
	switch typ {
	case "pkcs1":
		attempts = []func() (any, error){
			func() (any, error) { return x509.ParsePKCS1PrivateKey(der) },
			func() (any, error) { return x509.ParsePKCS1PublicKey(der) },
		}
	case "pkcs8":
		attempts = []func() (any, error){func() (any, error) { return x509.ParsePKCS8PrivateKey(der) }}
	case "spki":
		attempts = []func() (any, error){func() (any, error) { return x509.ParsePKIXPublicKey(der) }}
	case "sec1":
		attempts = []func() (any, error){func() (any, error) { return x509.ParseECPrivateKey(der) }}
	default:
		attempts = []func() (any, error){
			func() (any, error) { return x509.ParsePKCS8PrivateKey(der) },
			func() (any, error) { return x509.ParsePKIXPublicKey(der) },
			func() (any, error) { return x509.ParsePKCS1PrivateKey(der) },
			func() (any, error) { return x509.ParsePKCS1PublicKey(der) },
			func() (any, error) { return x509.ParseECPrivateKey(der) },
		}
	}
	var lastErr error
	for _, fn := range attempts {
		st, err, ok := try(fn())
		if ok {
			return st, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	if k, ok := dsaParsePKCS8(der); ok {
		return &keyState{kind: "private", typ: "dsa", key: k}, nil
	}
	if k, ok := dsaParseSPKI(der); ok {
		return &keyState{kind: "public", typ: "dsa", key: k}, nil
	}
	if lastErr != nil {
		return nil, js.Throw("Error", lastErr.Error())
	}
	return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYTYPE", "Failed to parse key")
}

func curveByName(name string) (elliptic.Curve, string, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "prime256v1", "p-256", "p256", "secp256r1":
		return elliptic.P256(), "prime256v1", true
	case "secp384r1", "p-384", "p384":
		return elliptic.P384(), "secp384r1", true
	case "secp521r1", "p-521", "p521":
		return elliptic.P521(), "secp521r1", true
	}
	return nil, "", false
}

func curveName(c elliptic.Curve) string {
	if c == nil {
		return ""
	}
	switch c.Params().BitSize {
	case 256:
		return "prime256v1"
	case 384:
		return "secp384r1"
	case 521:
		return "secp521r1"
	}
	return c.Params().Name
}

func (m *module) rawKey(st *keyState) ([]byte, error) {
	switch st.kind {
	case "secret":
		return st.secretBytes(), nil
	case "public":
		return x509.MarshalPKIXPublicKey(st.key)
	default:
		return x509.MarshalPKCS8PrivateKey(st.key)
	}
}

func (m *module) exportKey(st *keyState, hasOpts bool, opts js.Object) (js.Value, error) {
	if st.kind == "secret" {
		return m.bytesValue(st.secretBytes()), nil
	}
	format := "pem"
	typ := ""
	cipherName := ""
	passphrase := ""
	if hasOpts {
		if v := opts.Get("format"); v.IsString() {
			format = strings.ToLower(v.Str())
		}
		if v := opts.Get("type"); v.IsString() {
			typ = strings.ToLower(v.Str())
		}
		if v := opts.Get("cipher"); v.IsString() {
			cipherName = v.Str()
		}
		if v := opts.Get("passphrase"); v.IsString() {
			passphrase = v.Str()
		}
	}
	if format == "jwk" {
		return m.keyToJWK(st)
	}
	if format != "pem" && format != "der" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", fmt.Sprintf("The argument 'options.format' must be one of: 'pem', 'der', 'jwk'. Received '%s'", format))
	}
	der, label, err := m.keyDER(st, typ)
	if err != nil {
		return js.Value{}, err
	}
	if format == "der" {
		return m.bytesValue(der), nil
	}
	block := &pem.Block{Type: label, Bytes: der}
	if cipherName != "" {
		enc, ok := pemCipherFor(cipherName)
		if !ok {
			return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_CIPHER", "Unknown cipher")
		}
		block, err = x509.EncryptPEMBlock(rand.Reader, label, der, []byte(passphrase), enc)
		if err != nil {
			return js.Value{}, js.Throw("Error", err.Error())
		}
	}
	return m.e.JS.Str(string(pem.EncodeToMemory(block))), nil
}

func pemCipherFor(name string) (x509.PEMCipher, bool) {
	switch strings.ToLower(name) {
	case "aes-128-cbc":
		return x509.PEMCipherAES128, true
	case "aes-192-cbc":
		return x509.PEMCipherAES192, true
	case "aes-256-cbc":
		return x509.PEMCipherAES256, true
	}
	return 0, false
}

func (m *module) keyDER(st *keyState, typ string) ([]byte, string, error) {
	// Go's x509 package cannot marshal DSA keys, so those use the ASN.1
	// helpers in asym.go.
	if der, label, ok, err := dsaExportDER(st, typ); ok {
		return der, label, err
	}
	switch st.kind {
	case "public":
		if typ == "" || typ == "spki" {
			der, err := x509.MarshalPKIXPublicKey(st.key)
			return der, "PUBLIC KEY", err
		}
		if typ == "pkcs1" {
			if pub, ok := st.key.(*rsa.PublicKey); ok {
				return x509.MarshalPKCS1PublicKey(pub), "RSA PUBLIC KEY", nil
			}
		}
	case "private":
		if typ == "" || typ == "pkcs8" {
			der, err := x509.MarshalPKCS8PrivateKey(st.key)
			return der, "PRIVATE KEY", err
		}
		if typ == "pkcs1" {
			if priv, ok := st.key.(*rsa.PrivateKey); ok {
				return x509.MarshalPKCS1PrivateKey(priv), "RSA PRIVATE KEY", nil
			}
		}
		if typ == "sec1" {
			if priv, ok := st.key.(*ecdsa.PrivateKey); ok {
				der, err := x509.MarshalECPrivateKey(priv)
				return der, "EC PRIVATE KEY", err
			}
		}
	}
	return nil, "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid key object type")
}

func b64u(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func fromB64u(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

func (m *module) keyToJWK(st *keyState) (js.Value, error) {
	o := m.e.JS.Obj()
	pub := st
	if st.kind == "private" {
		p, err := publicState(st)
		if err != nil {
			return js.Value{}, err
		}
		pub = p
	}
	switch k := pub.key.(type) {
	case *rsa.PublicKey:
		o.SetStr("kty", "RSA")
		o.SetStr("n", b64u(k.N.Bytes()))
		o.SetStr("e", b64u(big.NewInt(int64(k.E)).Bytes()))
		if st.kind == "private" {
			priv := st.key.(*rsa.PrivateKey)
			priv.Precompute()
			o.SetStr("d", b64u(priv.D.Bytes()))
			o.SetStr("p", b64u(priv.Primes[0].Bytes()))
			o.SetStr("q", b64u(priv.Primes[1].Bytes()))
			o.SetStr("dp", b64u(priv.Precomputed.Dp.Bytes()))
			o.SetStr("dq", b64u(priv.Precomputed.Dq.Bytes()))
			o.SetStr("qi", b64u(priv.Precomputed.Qinv.Bytes()))
		}
	case *ecdsa.PublicKey:
		size := (k.Curve.Params().BitSize + 7) / 8
		o.SetStr("kty", "EC")
		o.SetStr("crv", jwkCurve(k.Curve))
		o.SetStr("x", b64u(padBytes(k.X.Bytes(), size)))
		o.SetStr("y", b64u(padBytes(k.Y.Bytes(), size)))
		if st.kind == "private" {
			priv := st.key.(*ecdsa.PrivateKey)
			o.SetStr("d", b64u(padBytes(priv.D.Bytes(), size)))
		}
	case ed25519.PublicKey:
		o.SetStr("kty", "OKP")
		o.SetStr("crv", "Ed25519")
		o.SetStr("x", b64u(k))
		if st.kind == "private" {
			priv := st.key.(ed25519.PrivateKey)
			o.SetStr("d", b64u(priv.Seed()))
		}
	case *ecdh.PublicKey:
		o.SetStr("kty", "OKP")
		o.SetStr("crv", "X25519")
		o.SetStr("x", b64u(k.Bytes()))
		if st.kind == "private" {
			priv := st.key.(*ecdh.PrivateKey)
			o.SetStr("d", b64u(priv.Bytes()))
		}
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_JWK_UNSUPPORTED_KEY_TYPE", "Unsupported JWK Key Type")
	}
	return o.V(), nil
}

func jwkCurve(c elliptic.Curve) string {
	switch c.Params().BitSize {
	case 256:
		return "P-256"
	case 384:
		return "P-384"
	case 521:
		return "P-521"
	}
	return c.Params().Name
}

func padBytes(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}

func (m *module) keyFromJWK(o js.Object) (*keyState, error) {
	kty := o.Get("kty")
	if !kty.IsString() {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid JWK: missing kty")
	}
	get := func(name string) ([]byte, error) {
		v := o.Get(name)
		if !v.IsString() {
			return nil, nil
		}
		b, err := fromB64u(v.Str())
		if err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		return b, nil
	}
	switch kty.Str() {
	case "RSA":
		nb, err := get("n")
		if err != nil {
			return nil, err
		}
		eb, err := get("e")
		if err != nil {
			return nil, err
		}
		if len(nb) == 0 || len(eb) == 0 {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid JWK RSA key")
		}
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: int(new(big.Int).SetBytes(eb).Int64())}
		db, err := get("d")
		if err != nil {
			return nil, err
		}
		if len(db) == 0 {
			return &keyState{kind: "public", typ: "rsa", key: pub}, nil
		}
		pb, _ := get("p")
		qb, _ := get("q")
		if len(pb) == 0 || len(qb) == 0 {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid JWK RSA private key")
		}
		priv := &rsa.PrivateKey{PublicKey: *pub, D: new(big.Int).SetBytes(db), Primes: []*big.Int{new(big.Int).SetBytes(pb), new(big.Int).SetBytes(qb)}}
		priv.Precompute()
		if err := priv.Validate(); err != nil {
			return nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "private", typ: "rsa", key: priv}, nil
	case "EC":
		crv := o.Get("crv")
		if !crv.IsString() {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid JWK EC key")
		}
		curve, canonical, ok := curveByName(crv.Str())
		if !ok {
			return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_CURVE", "Invalid EC curve name")
		}
		xb, err := get("x")
		if err != nil {
			return nil, err
		}
		yb, err := get("y")
		if err != nil {
			return nil, err
		}
		pub := &ecdsa.PublicKey{Curve: curve, X: new(big.Int).SetBytes(xb), Y: new(big.Int).SetBytes(yb)}
		db, err := get("d")
		if err != nil {
			return nil, err
		}
		invalid := func() error {
			return js.ThrowCode("Error", "ERR_CRYPTO_INVALID_JWK", "Invalid JWK EC key")
		}
		if dv := o.Get("d"); dv.Valid() && !dv.IsUndefined() && !dv.IsNull() && !dv.IsString() {
			return nil, invalid()
		}
		if len(db) == 0 {
			if !curve.IsOnCurve(pub.X, pub.Y) {
				return nil, invalid()
			}
			return &keyState{kind: "public", typ: "ec", key: pub, curve: canonical}, nil
		}
		// A private JWK is only valid when its scalar is in [1, n-1] and the
		// x/y point it carries is the public key derived from d.
		d := new(big.Int).SetBytes(db)
		if d.Sign() <= 0 || d.Cmp(curve.Params().N) >= 0 {
			return nil, invalid()
		}
		px, py := curve.ScalarBaseMult(db)
		if px == nil || px.Cmp(pub.X) != 0 || py.Cmp(pub.Y) != 0 {
			return nil, invalid()
		}
		priv := &ecdsa.PrivateKey{PublicKey: *pub, D: d}
		return &keyState{kind: "private", typ: "ec", key: priv, curve: canonical}, nil
	case "OKP":
		crv := o.Get("crv").Str()
		xb, err := get("x")
		if err != nil {
			return nil, err
		}
		db, err := get("d")
		if err != nil {
			return nil, err
		}
		invalidOKP := func() error {
			return js.ThrowCode("Error", "ERR_CRYPTO_INVALID_JWK", "Invalid JWK OKP key")
		}
		if dv := o.Get("d"); dv.Valid() && !dv.IsUndefined() && !dv.IsNull() && !dv.IsString() {
			return nil, invalidOKP()
		}
		switch crv {
		case "Ed25519":
			if len(db) > 0 {
				if len(db) != ed25519.SeedSize {
					return nil, invalidOKP()
				}
				priv := ed25519.NewKeyFromSeed(db)
				// When the JWK also carries x, it must be the public key the
				// seed derives; a mismatch means the JWK is inconsistent.
				if len(xb) == 0 || !bytes.Equal(priv.Public().(ed25519.PublicKey), xb) {
					return nil, invalidOKP()
				}
				return &keyState{kind: "private", typ: "ed25519", key: priv}, nil
			}
			if len(xb) != ed25519.PublicKeySize {
				return nil, invalidOKP()
			}
			return &keyState{kind: "public", typ: "ed25519", key: ed25519.PublicKey(xb)}, nil
		case "X25519":
			if len(db) > 0 {
				if len(db) != 32 {
					return nil, invalidOKP()
				}
				k, err := ecdh.X25519().NewPrivateKey(db)
				if err != nil {
					return nil, invalidOKP()
				}
				if len(xb) == 0 || !bytes.Equal(k.PublicKey().Bytes(), xb) {
					return nil, invalidOKP()
				}
				return &keyState{kind: "private", typ: "x25519", key: k}, nil
			}
			if len(xb) != 32 {
				return nil, invalidOKP()
			}
			k, err := ecdh.X25519().NewPublicKey(xb)
			if err != nil {
				return nil, invalidOKP()
			}
			return &keyState{kind: "public", typ: "x25519", key: k}, nil
		default:
			return nil, invalidOKP()
		}
	}
	return nil, js.ThrowCode("Error", "ERR_CRYPTO_JWK_UNSUPPORTED_KEY_TYPE", "Unsupported JWK Key Type")
}

func (m *module) algoArg(v js.Value) string {
	if v.Valid() && !v.IsUndefined() && !v.IsNull() {
		if v.IsString() {
			return v.Str()
		}
		if v.IsObject() {
			if n := v.Object().Get("name"); n.IsString() {
				return n.Str()
			}
		}
	}
	return ""
}

func (m *module) bytesArg(v js.Value, name string) ([]byte, error) {
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	return nil, typeErr(`The "%s" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, name, describe(v))
}

func cryptoHashFor(algo string) stdcrypto.Hash {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(algo), "-", "")) {
	case "md5":
		return stdcrypto.MD5
	case "sha1":
		return stdcrypto.SHA1
	case "sha224":
		return stdcrypto.SHA224
	case "sha256":
		return stdcrypto.SHA256
	case "sha384":
		return stdcrypto.SHA384
	case "sha512":
		return stdcrypto.SHA512
	case "sha512224":
		return stdcrypto.SHA512_224
	case "sha512256":
		return stdcrypto.SHA512_256
	}
	return 0
}

func (m *module) digestFor(algo string, data []byte) ([]byte, stdcrypto.Hash, error) {
	fn, _, ok := lookup(algo)
	if !ok {
		return nil, 0, invalidDigest()
	}
	h := fn()
	h.Write(data)
	return h.Sum(nil), cryptoHashFor(algo), nil
}

func (m *module) signData(algo string, data []byte, st *keyState) ([]byte, error) {
	if st.kind == "secret" {
		fn, _, ok := lookup(algo)
		if !ok {
			return nil, invalidDigest()
		}
		h := hmac.New(fn, st.secretBytes())
		h.Write(data)
		return h.Sum(nil), nil
	}
	if st.kind != "private" {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a private key")
	}
	switch k := st.key.(type) {
	case ed25519.PrivateKey:
		return ed25519.Sign(k, data), nil
	case *ed25519.PrivateKey:
		return ed25519.Sign(*k, data), nil
	}
	digest, ch, err := m.digestFor(algo, data)
	if err != nil {
		return nil, err
	}
	if ch == 0 {
		return nil, invalidDigest()
	}
	if sig, ok, signErr := dsaSignState(st, digest); ok {
		return sig, signErr
	}
	switch k := st.key.(type) {
	case *rsa.PrivateKey:
		if st.typ == "rsa-pss" || st.padding == rsaPaddingPSS {
			saltLen := rsa.PSSSaltLengthEqualsHash
			if st.saltLen > 0 {
				saltLen = st.saltLen
			}
			return rsa.SignPSS(rand.Reader, k, ch, digest, &rsa.PSSOptions{SaltLength: saltLen, Hash: ch})
		}
		return rsa.SignPKCS1v15(rand.Reader, k, ch, digest)
	case *ecdsa.PrivateKey:
		if st.dsaEncoding == "ieee-p1363" {
			sig, err := ecdsa.SignASN1(rand.Reader, k, digest)
			if err != nil {
				return nil, js.Throw("Error", err.Error())
			}
			return ecdsaSigToP1363(sig, k.Curve)
		}
		return ecdsa.SignASN1(rand.Reader, k, digest)
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a private key")
}

func (m *module) verifyData(algo string, data []byte, st *keyState, sig []byte) (bool, error) {
	if st.kind == "secret" {
		fn, _, ok := lookup(algo)
		if !ok {
			return false, invalidDigest()
		}
		h := hmac.New(fn, st.secretBytes())
		h.Write(data)
		return subtle.ConstantTimeCompare(h.Sum(nil), sig) == 1, nil
	}
	if st.kind == "private" {
		pub, err := publicState(st)
		if err != nil {
			return false, err
		}
		st = pub
	}
	switch pub := st.key.(type) {
	case ed25519.PublicKey:
		return ed25519.Verify(pub, data, sig), nil
	}
	digest, ch, err := m.digestFor(algo, data)
	if err != nil {
		return false, err
	}
	if ch == 0 {
		return false, invalidDigest()
	}
	if okSig, ok := dsaVerifyState(st, digest, sig); ok {
		return okSig, nil
	}
	switch k := st.key.(type) {
	case *rsa.PublicKey:
		if st.typ == "rsa-pss" || st.padding == rsaPaddingPSS {
			return rsa.VerifyPSS(k, ch, digest, sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto, Hash: ch}) == nil, nil
		}
		return rsa.VerifyPKCS1v15(k, ch, digest, sig) == nil, nil
	case *ecdsa.PublicKey:
		if st.dsaEncoding == "ieee-p1363" {
			der, err := p1363ToECDSASig(sig, k.Curve)
			if err != nil {
				return false, err
			}
			sig = der
		}
		return ecdsa.VerifyASN1(k, digest, sig), nil
	}
	return false, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"key\" argument must be a public key or private key")
}

func (m *module) installSignProto(p js.Object, verify bool) {
	p.SetFuncHidden("update", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.signerStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		v := arg(args, 0)
		enc := "utf8"
		if e := arg(args, 1); e.Valid() && e.IsString() {
			enc = e.Str()
		}
		if v.IsString() {
			d, err := decodeInput(v.Str(), enc)
			if err != nil {
				return js.Value{}, err
			}
			st.buf = append(st.buf, d...)
		} else if b, ok := m.bytesOf(v); ok {
			st.buf = append(st.buf, b...)
		} else {
			return js.Value{}, typeErr(`The "data" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
		}
		return this.V(), nil
	})
	if verify {
		p.SetFuncHidden("verify", func(this js.Object, args []js.Value) (js.Value, error) {
			st, err := m.signerStateOf(this)
			if err != nil {
				return js.Value{}, err
			}
			key, err := m.loadKeyForOp(arg(args, 0), false)
			if err != nil {
				return js.Value{}, err
			}
			sig := arg(args, 1)
			var sigBytes []byte
			if sig.IsString() {
				sigBytes, err = decodeInput(sig.Str(), "hex")
				if err != nil {
					return js.Value{}, err
				}
			} else if b, ok := m.bytesOf(sig); ok {
				sigBytes = b
			} else {
				return js.Value{}, typeErr(`The "signature" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(sig))
			}
			ok, err := m.verifyData(st.algo, st.buf, key, sigBytes)
			if err != nil {
				return js.Value{}, err
			}
			return m.e.JS.Bool(ok), nil
		})
	} else {
		p.SetFuncHidden("sign", func(this js.Object, args []js.Value) (js.Value, error) {
			st, err := m.signerStateOf(this)
			if err != nil {
				return js.Value{}, err
			}
			key, err := m.loadKeyForOp(arg(args, 0), true)
			if err != nil {
				return js.Value{}, err
			}
			sig, err := m.signData(st.algo, st.buf, key)
			if err != nil {
				return js.Value{}, err
			}
			if enc := arg(args, 1); enc.IsString() {
				return m.cipherOutput(sig, enc.Str())
			}
			return m.bytesValue(sig), nil
		})
	}
}

func (m *module) signerStateOf(this js.Object) (*signState, error) {
	id := this.Get("$godeSignId")
	m.mu.Lock()
	defer m.mu.Unlock()
	if !id.Valid() || !id.IsNumber() {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	st := m.signers[id.Int()]
	if st == nil {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	return st, nil
}

func (m *module) newSigner(algo string, verify bool) (js.Value, error) {
	st := &signState{algo: algo, isSig: verify}
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.signers[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeSignId", m.e.JS.Num(float64(id)))
	proto := m.signProto
	if verify {
		proto = m.verifyProto
	}
	o.SetProto(proto)
	return o.V(), nil
}

func (m *module) generatePair(typ string, hasOpts bool, opts js.Object) (*keyState, *keyState, error) {
	switch typ {
	case "rsa", "rsa-pss":
		bits := 2048
		if hasOpts {
			if v := opts.Get("modulusLength"); v.Valid() && v.IsNumber() {
				bits = int(v.Float())
			}
		}
		if bits < 512 {
			bits = 512
		}
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		key.Precompute()
		priv := &keyState{kind: "private", typ: "rsa", key: key}
		if typ == "rsa-pss" {
			priv.typ = "rsa-pss"
			// Node does not add an RSASSA-PSS-params sequence unless the
			// caller supplied parameters (test-crypto-keygen-no-rsassa-pss-params).
			if hasOpts {
				if v := opts.Get("hashAlgorithm"); v.IsString() {
					priv.hashAlg = v.Str()
				}
				if v := opts.Get("hash"); v.IsString() {
					priv.hashAlg = v.Str()
				}
				if v := opts.Get("mgf1HashAlgorithm"); v.IsString() {
					priv.mgf1HashAlg = v.Str()
				}
				if v := opts.Get("mgf1Hash"); v.IsString() {
					priv.mgf1HashAlg = v.Str()
				}
				if v := opts.Get("saltLength"); v.IsNumber() {
					priv.saltLen = int(v.Float())
				}
				// RFC 8017 9.1: MGF1 uses the same hash as the signature by
				// default, so only an explicit mgf1 option overrides it.
				if priv.mgf1HashAlg == "" {
					priv.mgf1HashAlg = priv.hashAlg
				}
			}
		}
		pub, err := publicState(priv)
		if err != nil {
			return nil, nil, err
		}
		return pub, priv, nil
	case "ec":
		name := ""
		if hasOpts {
			if v := opts.Get("namedCurve"); v.IsString() {
				name = v.Str()
			}
		}
		curve, canonical, ok := curveByName(name)
		if !ok {
			return nil, nil, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_CURVE", "Invalid EC curve name")
		}
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		priv := &keyState{kind: "private", typ: "ec", key: key, curve: canonical}
		return &keyState{kind: "public", typ: "ec", key: &key.PublicKey, curve: canonical}, priv, nil
	case "ed25519":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "public", typ: "ed25519", key: pub}, &keyState{kind: "private", typ: "ed25519", key: priv}, nil
	case "x25519":
		priv, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "public", typ: "x25519", key: priv.PublicKey()}, &keyState{kind: "private", typ: "x25519", key: priv}, nil
	case "ml-kem-768":
		dk, err := mlkem.GenerateKey768()
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "public", typ: "ml-kem-768", key: dk.EncapsulationKey()},
			&keyState{kind: "private", typ: "ml-kem-768", key: dk}, nil
	case "ml-kem-1024":
		dk, err := mlkem.GenerateKey1024()
		if err != nil {
			return nil, nil, js.Throw("Error", err.Error())
		}
		return &keyState{kind: "public", typ: "ml-kem-1024", key: dk.EncapsulationKey()},
			&keyState{kind: "private", typ: "ml-kem-1024", key: dk}, nil
	case "dsa":
		modulusBits, divisorBits := 0, 0
		if hasOpts {
			if v := opts.Get("modulusLength"); v.Valid() && v.IsNumber() {
				modulusBits = int(v.Float())
			}
			if v := opts.Get("divisorLength"); v.Valid() && v.IsNumber() {
				divisorBits = int(v.Float())
			}
		}
		if modulusBits == 0 {
			return nil, nil, js.ThrowCode("TypeError", "ERR_MISSING_OPTION", "options.modulusLength is required")
		}
		if divisorBits == 0 {
			return nil, nil, js.ThrowCode("TypeError", "ERR_MISSING_OPTION", "options.divisorLength is required")
		}
		key, err := dsaGenerateKey(modulusBits, divisorBits)
		if err != nil {
			return nil, nil, err
		}
		priv := &keyState{kind: "private", typ: "dsa", key: key}
		return &keyState{kind: "public", typ: "dsa", key: &key.PublicKey}, priv, nil
	}
	return nil, nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", fmt.Sprintf("The argument 'type' must be a supported key type. Received '%s'", typ))
}

func (m *module) keyResult(st *keyState, hasOpts bool, opts js.Object, field string) (js.Value, error) {
	if hasOpts {
		if enc := opts.Get(field); enc.IsObject() {
			return m.exportKey(st, true, enc.Object())
		}
	}
	return m.newKeyObject(st).V(), nil
}

func (m *module) generateKeyPair(args []js.Value, async bool) (js.Value, error) {
	typV := arg(args, 0)
	if !typV.IsString() {
		return js.Value{}, typeErr(`The "type" argument must be of type string. Received %s`, describe(typV))
	}
	typ := strings.ToLower(typV.Str())
	hasOpts := false
	var opts js.Object
	cbIndex := 2
	if async && arg(args, 1).IsFunction() {
		// generateKeyPair(type, callback) omits the options object.
		cbIndex = 1
	} else if v := arg(args, 1); v.IsObject() {
		opts, hasOpts = v.Object(), true
	} else if v.Valid() && !v.IsUndefined() && !v.IsNull() {
		return js.Value{}, typeErr(`The "options" argument must be of type object. Received %s`, describe(v))
	}
	var cb js.Value
	if async {
		cb = arg(args, cbIndex)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
	}
	pub, priv, err := m.generatePair(typ, hasOpts, opts)
	if err != nil {
		return js.Value{}, err
	}
	pubV, err := m.keyResult(pub, hasOpts, opts, "publicKeyEncoding")
	if err != nil {
		return js.Value{}, err
	}
	privV, err := m.keyResult(priv, hasOpts, opts, "privateKeyEncoding")
	if err != nil {
		return js.Value{}, err
	}
	if async {
		m.e.NextTick(cb, m.e.JS.Null(), pubV, privV)
		return js.Value{}, nil
	}
	o := m.e.JS.Obj()
	o.Set("publicKey", pubV)
	o.Set("privateKey", privV)
	return o.V(), nil
}

func (m *module) generateSecretKey(args []js.Value, async bool) (js.Value, error) {
	typV := arg(args, 0)
	if !typV.IsString() {
		return js.Value{}, typeErr(`The "type" argument must be of type string. Received %s`, describe(typV))
	}
	typ := strings.ToLower(typV.Str())
	if typ != "hmac" && typ != "aes" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", fmt.Sprintf("The argument 'type' must be a supported key type. Received '%s'", typ))
	}
	var opts js.Object
	hasOpts := false
	if v := arg(args, 1); v.IsObject() {
		opts, hasOpts = v.Object(), true
	}
	var cb js.Value
	if async {
		cb = arg(args, 2)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
	}
	bits := 0
	if hasOpts {
		if v := opts.Get("length"); v.IsNumber() {
			bits = int(v.Float())
		}
	}
	if bits <= 0 || bits%8 != 0 {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "The argument 'options.length' is required")
	}
	data := make([]byte, bits/8)
	if _, err := rand.Read(data); err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	st := &keyState{kind: "secret", typ: "secret", key: data}
	key := m.newKeyObject(st).V()
	if async {
		m.e.NextTick(cb, m.e.JS.Null(), key)
		return js.Value{}, nil
	}
	return key, nil
}

func (m *module) hkdf(args []js.Value) ([]byte, error) {
	digest := arg(args, 0)
	if !digest.IsString() {
		return nil, typeErr(`The "digest" argument must be of type string. Received %s`, describe(digest))
	}
	algo := digest.Str()
	fn, _, ok := lookup(algo)
	if !ok {
		return nil, invalidDigest()
	}
	ikm, err := m.bytesArg(arg(args, 1), "ikm")
	if err != nil {
		return nil, err
	}
	var salt, info []byte
	if v := arg(args, 2); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		salt, err = m.bytesArg(v, "salt")
		if err != nil {
			return nil, err
		}
	}
	if v := arg(args, 3); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		info, err = m.bytesArg(v, "info")
		if err != nil {
			return nil, err
		}
	}
	kl := arg(args, 4)
	if !kl.IsNumber() {
		return nil, typeErr(`The "length" argument must be of type number.%s`, invalidArgHelper(kl))
	}
	length := int(kl.Float())
	if length < 0 || kl.Float() != math.Trunc(kl.Float()) {
		return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "length" is out of range. It must be a non-negative integer. Received %s`, numberString(kl.Float())))
	}
	return hkdfDerive(fn, ikm, salt, info, length)
}

func hkdfDerive(fn func() stdhash.Hash, ikm, salt, info []byte, length int) ([]byte, error) {
	hashLen := fn().Size()
	if salt == nil {
		salt = make([]byte, hashLen)
	}
	mac := hmac.New(fn, salt)
	mac.Write(ikm)
	prk := mac.Sum(nil)
	if length > 255*hashLen {
		return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYLEN", "Invalid key length")
	}
	out := make([]byte, 0, length)
	var t []byte
	for i := byte(1); len(out) < length; i++ {
		h := hmac.New(fn, prk)
		h.Write(t)
		h.Write(info)
		h.Write([]byte{i})
		t = h.Sum(nil)
		out = append(out, t...)
	}
	return out[:length], nil
}

func (m *module) pbkdf2(args []js.Value) ([]byte, error) {
	password, err := m.passwordBytes(arg(args, 0), "password")
	if err != nil {
		return nil, err
	}
	salt, err := m.passwordBytes(arg(args, 1), "salt")
	if err != nil {
		return nil, err
	}
	itV := arg(args, 2)
	if !itV.IsNumber() {
		return nil, typeErr(`The "iterations" argument must be of type number.%s`, invalidArgHelper(itV))
	}
	iterations := int(itV.Float())
	if itV.Float() < 1 || itV.Float() > 2147483647 || itV.Float() != math.Trunc(itV.Float()) {
		return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "iterations" is out of range. It must be >= 1 && <= 2147483647. Received %s`, numberString(itV.Float())))
	}
	klV := arg(args, 3)
	if !klV.IsNumber() {
		return nil, typeErr(`The "keylen" argument must be of type number.%s`, invalidArgHelper(klV))
	}
	kl := klV.Float()
	if math.IsNaN(kl) || math.IsInf(kl, 0) || kl != math.Trunc(kl) {
		return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "keylen" is out of range. It must be an integer. Received %s`, numberString(kl)))
	}
	if kl < 0 || kl > 2147483647 {
		return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "keylen" is out of range. It must be >= 0 && <= 2147483647. Received %s`, numberString(kl)))
	}
	keylen := int(kl)
	algo := arg(args, 4)
	if !algo.IsString() {
		return nil, typeErr(`The "digest" argument must be of type string. Received %s`, describe(algo))
	}
	fn, _, ok := lookup(algo.Str())
	if !ok {
		return nil, invalidDigest()
	}
	return pbkdf2Derive(fn, password, salt, iterations, keylen), nil
}

func (m *module) passwordBytes(v js.Value, name string) ([]byte, error) {
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	return nil, typeErr(`The "%s" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, name, describe(v))
}

func pbkdf2Derive(fn func() stdhash.Hash, password, salt []byte, iterations, keylen int) []byte {
	hashLen := fn().Size()
	blocks := (keylen + hashLen - 1) / hashLen
	out := make([]byte, 0, blocks*hashLen)
	for block := 1; block <= blocks; block++ {
		mac := hmac.New(fn, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte{}, u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(fn, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keylen]
}

func (m *module) checkPrime(args []js.Value) (bool, error) {
	candidate, err := m.bytesArg(arg(args, 0), "candidate")
	if err != nil {
		return false, err
	}
	checks := 0
	if opts := arg(args, 1); opts.IsObject() {
		if v := opts.Object().Get("checks"); v.IsNumber() {
			checks = int(v.Float())
		}
	}
	return new(big.Int).SetBytes(candidate).ProbablyPrime(checks), nil
}

func (m *module) generatePrime(args []js.Value) ([]byte, error) {
	sizeV := arg(args, 0)
	if !sizeV.IsNumber() {
		return nil, typeErr(`The "size" argument must be of type number.%s`, invalidArgHelper(sizeV))
	}
	size := int(sizeV.Float())
	if size < 2 {
		return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "size" is out of range. It must be >= 2 && <= 2147483647. Received %s`, numberString(sizeV.Float())))
	}
	p, err := rand.Prime(rand.Reader, size)
	if err != nil {
		return nil, js.Throw("Error", err.Error())
	}
	return p.Bytes(), nil
}

type cipherSpec struct {
	keyLen int
	ivLen  int
	mode   string
}

var cipherTable = []struct {
	name string
	spec cipherSpec
}{
	{"aes-128-ecb", cipherSpec{16, 0, "ecb"}},
	{"aes-192-ecb", cipherSpec{24, 0, "ecb"}},
	{"aes-256-ecb", cipherSpec{32, 0, "ecb"}},
	{"aes-128-cbc", cipherSpec{16, 16, "cbc"}},
	{"aes-192-cbc", cipherSpec{24, 16, "cbc"}},
	{"aes-256-cbc", cipherSpec{32, 16, "cbc"}},
	{"aes-128-ctr", cipherSpec{16, 16, "ctr"}},
	{"aes-192-ctr", cipherSpec{24, 16, "ctr"}},
	{"aes-256-ctr", cipherSpec{32, 16, "ctr"}},
	{"aes-128-ofb", cipherSpec{16, 16, "ofb"}},
	{"aes-192-ofb", cipherSpec{24, 16, "ofb"}},
	{"aes-256-ofb", cipherSpec{32, 16, "ofb"}},
	{"aes-128-cfb", cipherSpec{16, 16, "cfb"}},
	{"aes-192-cfb", cipherSpec{24, 16, "cfb"}},
	{"aes-256-cfb", cipherSpec{32, 16, "cfb"}},
	{"aes-128-gcm", cipherSpec{16, 12, "gcm"}},
	{"aes-192-gcm", cipherSpec{24, 12, "gcm"}},
	{"aes-256-gcm", cipherSpec{32, 12, "gcm"}},
	// Single DES and three-key triple DES come from crypto/des. Two-key
	// (des-ede-*) needs a 16-byte key, which Go's TripleDESCipher does not
	// accept, so only these names are listed.
	{"des-ede3-cbc", cipherSpec{24, 8, "cbc"}},
	{"des-ede3-ecb", cipherSpec{24, 0, "ecb"}},
	{"des-ede3", cipherSpec{24, 0, "ecb"}},
	{"des-ede3-cfb", cipherSpec{24, 8, "cfb"}},
	{"des-ede3-ofb", cipherSpec{24, 8, "ofb"}},
	{"des-cbc", cipherSpec{8, 8, "cbc"}},
	{"des-ecb", cipherSpec{8, 0, "ecb"}},
	{"des-cfb", cipherSpec{8, 8, "cfb"}},
	{"des-ofb", cipherSpec{8, 8, "ofb"}},
}

func cipherSpecFor(name string) (cipherSpec, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, c := range cipherTable {
		if c.name == name {
			return c.spec, true
		}
	}
	return cipherSpec{}, false
}

func cipherNames() []string {
	out := make([]string, len(cipherTable))
	for i, c := range cipherTable {
		out[i] = c.name
	}
	return out
}

func (m *module) getCipherInfo(args []js.Value) (js.Value, error) {
	v := arg(args, 0)
	if !v.IsString() {
		return m.e.JS.Undefined(), nil
	}
	name := strings.ToLower(strings.TrimSpace(v.Str()))
	spec, ok := cipherSpecFor(name)
	if !ok {
		return m.e.JS.Undefined(), nil
	}
	// A requested key or IV length that does not match the cipher makes the
	// lookup fail (Node returns undefined rather than a partial record).
	if opts := arg(args, 1); opts.IsObject() && !opts.IsNull() {
		o := opts.Object()
		if k := o.Get("keyLength"); k.Valid() && !k.IsUndefined() && (!k.IsNumber() || int(k.Float()) != spec.keyLen) {
			return m.e.JS.Undefined(), nil
		}
		if iv := o.Get("ivLength"); iv.Valid() && !iv.IsUndefined() && (!iv.IsNumber() || int(iv.Float()) != spec.ivLen) {
			return m.e.JS.Undefined(), nil
		}
	}
	blockSize := 16
	if spec.mode != "cbc" {
		blockSize = 1
	}
	o := m.e.JS.Obj()
	o.SetStr("name", name)
	o.SetNum("nid", 0)
	o.SetNum("blockSize", float64(blockSize))
	o.SetNum("ivLength", float64(spec.ivLen))
	o.SetNum("keyLength", float64(spec.keyLen))
	o.SetStr("mode", spec.mode)
	return o.V(), nil
}

func (m *module) newCipher(args []js.Value, decrypt bool) (js.Value, error) {
	algo := arg(args, 0)
	if !algo.IsString() {
		// Node names this parameter "cipher" for both createCipheriv and
		// createDecipheriv.
		return js.Value{}, typeErr(`The "cipher" argument must be of type string. Received %s`, describe(algo))
	}
	name := strings.ToLower(strings.TrimSpace(algo.Str()))
	spec, ok := cipherSpecFor(name)
	if !ok {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_CIPHER", "Unknown cipher")
	}
	key, err := m.cipherBytesArg(arg(args, 1), "key")
	if err != nil {
		return js.Value{}, err
	}
	if len(key) != spec.keyLen {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYLEN", "Invalid key length")
	}
	ivArg := arg(args, 2)
	var iv []byte
	if ivArg.Valid() && !ivArg.IsNull() && !ivArg.IsUndefined() {
		iv, err = m.cipherBytesArg(ivArg, "iv")
		if err != nil {
			return js.Value{}, err
		}
	}
	tagLen := 16
	if opts := arg(args, 3); opts.IsObject() {
		if v := opts.Object().Get("authTagLength"); v.Valid() && v.IsNumber() {
			tagLen = int(v.Float())
		}
	}
	if spec.mode == "gcm" {
		if len(iv) == 0 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_IV", "Invalid initialization vector")
		}
		if tagLen < 12 || tagLen > 16 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_AUTH_TAG", "Invalid authentication tag length")
		}
		if tagLen != 16 && len(iv) != 12 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_IV", "Invalid initialization vector")
		}
	} else if len(iv) != spec.ivLen {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_IV", "Invalid initialization vector")
	}

	var block stdcipher.Block
	switch {
	case strings.HasPrefix(name, "des-ede3"):
		block, err = des.NewTripleDESCipher(key)
	case strings.HasPrefix(name, "des-"):
		block, err = des.NewCipher(key)
	default:
		block, err = aes.NewCipher(key)
	}
	if err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	st := &cipherState{
		decrypt: decrypt,
		name:    name,
		mode:    spec.mode,
		block:   block,
		iv:      append([]byte{}, iv...),
		tagLen:  tagLen,
		autoPad: true,
	}
	switch spec.mode {
	case "gcm":
		if tagLen == 16 {
			st.aead, err = stdcipher.NewGCMWithNonceSize(block, len(iv))
		} else {
			st.aead, err = stdcipher.NewGCMWithTagSize(block, tagLen)
		}
	case "ecb":
		st.blockMode = ecbMode{block, decrypt}
	case "cbc":
		if decrypt {
			st.blockMode = stdcipher.NewCBCDecrypter(block, iv)
		} else {
			st.blockMode = stdcipher.NewCBCEncrypter(block, iv)
		}
	case "ctr":
		st.stream = stdcipher.NewCTR(block, iv)
	case "ofb":
		st.stream = stdcipher.NewOFB(block, iv)
	case "cfb":
		if decrypt {
			st.stream = stdcipher.NewCFBDecrypter(block, iv)
		} else {
			st.stream = stdcipher.NewCFBEncrypter(block, iv)
		}
	}
	if err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	return m.newCipherObject(st, decrypt).V(), nil
}

func (m *module) cipherBytesArg(v js.Value, name string) ([]byte, error) {
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	// Node accepts a secret KeyObject wherever a raw key is expected.
	if isKeyObjectValue(v) {
		st, err := m.keyStateOf(v.Object())
		if err != nil {
			return nil, err
		}
		if st.kind != "secret" {
			return nil, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_KEYTYPE", "Invalid key type")
		}
		return st.secretBytes(), nil
	}
	return nil, typeErr(`The "%s" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, name, describe(v))
}

func (m *module) newCipherObject(st *cipherState, decrypt bool) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.ciphers[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeCipherId", m.e.JS.Num(float64(id)))
	proto := m.cipherProto
	if decrypt {
		proto = m.decipherProto
	}
	o.SetProto(proto)
	return o
}

func (m *module) cipherStateOf(this js.Object) (*cipherState, error) {
	id := this.Get("$godeCipherId")
	m.mu.Lock()
	defer m.mu.Unlock()
	if !id.Valid() || !id.IsNumber() {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	st := m.ciphers[id.Int()]
	if st == nil {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	return st, nil
}

func (m *module) installCipherProto(p js.Object, decrypt bool) {
	p.SetFuncHidden("update", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.cipherUpdate(this, args)
	})
	p.SetFuncHidden("final", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.cipherFinal(this, args)
	})
	p.SetFuncHidden("setAAD", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.cipherSetAAD(this, args)
	})
	p.SetFuncHidden("setAutoPadding", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.cipherStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		b := arg(args, 0)
		if !b.IsBool() {
			return js.Value{}, typeErr(`The "autoPadding" argument must be of type boolean. Received %s`, describe(b))
		}
		st.autoPad = b.Bool()
		return this.V(), nil
	})
	if decrypt {
		p.SetFuncHidden("setAuthTag", func(this js.Object, args []js.Value) (js.Value, error) {
			return m.cipherSetAuthTag(this, args)
		})
	} else {
		p.SetFuncHidden("getAuthTag", func(this js.Object, args []js.Value) (js.Value, error) {
			st, err := m.cipherStateOf(this)
			if err != nil {
				return js.Value{}, err
			}
			// OpenSSL's getAuthTag control fails both for non-AEAD ciphers and
			// while the cipher is still streaming; Node reports both as an
			// invalid state ("Invalid state for operation getAuthTag").
			if st.mode != "gcm" || st.tag == nil {
				return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_STATE",
					"Invalid state for operation getAuthTag")
			}
			return m.bytesValue(st.tag), nil
		})
	}
}

func (m *module) cipherSetAAD(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.cipherStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.mode != "gcm" || st.updated || st.done {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_STATE",
			"Invalid state for operation setAAD")
	}
	data, err := m.cipherBytesArg(arg(args, 0), "aad")
	if err != nil {
		return js.Value{}, err
	}
	st.aad = append(st.aad, data...)
	return this.V(), nil
}

func (m *module) cipherSetAuthTag(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.cipherStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.mode != "gcm" || st.updated || st.done {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_STATE",
			"Invalid state for operation setAuthTag")
	}
	tag, err := m.cipherBytesArg(arg(args, 0), "authTag")
	if err != nil {
		return js.Value{}, err
	}
	if len(tag) < 12 || len(tag) > 16 {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_AUTH_TAG", "Invalid authentication tag length")
	}
	st.tag = append([]byte{}, tag...)
	st.tagLen = len(tag)
	st.tagGiven = true
	return this.V(), nil
}

func (m *module) cipherInput(args []js.Value) ([]byte, string, error) {
	v := arg(args, 0)
	outEnc := ""
	if v.IsString() {
		inEnc := ""
		if e := arg(args, 1); e.Valid() && e.IsString() {
			inEnc = e.Str()
		}
		if e := arg(args, 2); e.Valid() && e.IsString() {
			outEnc = e.Str()
		}
		data, err := decodeInput(v.Str(), inEnc)
		return data, outEnc, err
	}
	b, ok := m.bytesOf(v)
	if !ok {
		return nil, "", typeErr(`The "data" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
	}
	if e := arg(args, 1); e.Valid() && e.IsString() {
		outEnc = e.Str()
	}
	return b, outEnc, nil
}

func (m *module) cipherUpdate(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.cipherStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_CIPHER_FINALIZED", "Cipher already finalized")
	}
	data, outEnc, err := m.cipherInput(args)
	if err != nil {
		return js.Value{}, err
	}
	if err := lockCipherEncoding(st, outEnc); err != nil {
		return js.Value{}, err
	}
	st.updated = true
	out, err := m.cipherDoUpdate(st, data)
	if err != nil {
		return js.Value{}, err
	}
	return m.cipherOutput(out, outEnc)
}

func (m *module) cipherDoUpdate(st *cipherState, data []byte) ([]byte, error) {
	switch st.mode {
	case "gcm":
		st.buf = append(st.buf, data...)
		return nil, nil
	case "ctr", "ofb", "cfb":
		if st.stream == nil {
			return nil, js.Throw("Error", "Cipher is not initialized")
		}
		out := make([]byte, len(data))
		st.stream.XORKeyStream(out, data)
		return out, nil
	default: // cbc
		st.buf = append(st.buf, data...)
		n := len(st.buf)
		if st.decrypt {
			keep := aes.BlockSize
			if !st.autoPad {
				keep = 0
			}
			n -= keep
		}
		n -= n % aes.BlockSize
		if n <= 0 {
			return nil, nil
		}
		out := make([]byte, n)
		st.blockMode.CryptBlocks(out, st.buf[:n])
		st.buf = st.buf[n:]
		return out, nil
	}
}

func (m *module) cipherFinal(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.cipherStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.done {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_CIPHER_FINALIZED", "Cipher already finalized")
	}
	outEnc := ""
	if e := arg(args, 0); e.Valid() && e.IsString() {
		outEnc = e.Str()
	}
	if err := lockCipherEncoding(st, outEnc); err != nil {
		return js.Value{}, err
	}
	st.done = true
	var out []byte
	switch st.mode {
	case "gcm":
		if st.decrypt {
			if !st.tagGiven {
				return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_AUTH_TAG", "Invalid authentication tag length")
			}
			combined := append(append([]byte{}, st.buf...), st.tag...)
			plain, err := st.aead.Open(nil, st.iv, combined, st.aad)
			if err != nil {
				return js.Value{}, js.Throw("Error", "Unsupported state or unable to authenticate data")
			}
			out = plain
		} else {
			sealed := st.aead.Seal(nil, st.iv, st.buf, st.aad)
			st.tag = sealed[len(sealed)-st.tagLen:]
			out = sealed[:len(sealed)-st.tagLen]
		}
	case "ctr", "ofb", "cfb":
		out = nil
	default: // cbc
		if st.decrypt {
			if len(st.buf) > 0 {
				if len(st.buf)%aes.BlockSize != 0 {
					return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_EVP_WRONG_FINAL_BLOCK_LENGTH", "wrong final block length")
				}
				plain := make([]byte, len(st.buf))
				st.blockMode.CryptBlocks(plain, st.buf)
				st.buf = nil
				if st.autoPad {
					plain, err = pkcs7Unpad(plain)
					if err != nil {
						return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_BAD_DECRYPT", "bad decrypt")
					}
				}
				out = plain
			} else if st.autoPad {
				return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_EVP_WRONG_FINAL_BLOCK_LENGTH", "wrong final block length")
			}
		} else {
			if st.autoPad {
				pad := aes.BlockSize - len(st.buf)%aes.BlockSize
				st.buf = append(st.buf, bytes.Repeat([]byte{byte(pad)}, pad)...)
			} else if len(st.buf)%aes.BlockSize != 0 {
				return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_EVP_DATA_NOT_MULTIPLE_OF_BLOCK_LENGTH", "data not multiple of block length")
			}
			if len(st.buf) > 0 {
				out = make([]byte, len(st.buf))
				st.blockMode.CryptBlocks(out, st.buf)
				st.buf = nil
			}
		}
	}
	return m.cipherOutput(out, outEnc)
}

// canonicalCipherEncoding normalizes a cipher output encoding the way Node's
// getEncoding does; ok is false for an unknown label.
func canonicalCipherEncoding(enc string) (string, bool) {
	switch strings.ToLower(enc) {
	case "utf8", "utf-8":
		return "utf8", true
	case "hex":
		return "hex", true
	case "base64":
		return "base64", true
	case "base64url":
		return "base64url", true
	case "latin1", "binary":
		return "latin1", true
	case "ascii":
		return "ascii", true
	case "ucs2", "ucs-2", "utf16le", "utf-16le":
		return "utf16le", true
	}
	return "", false
}

// lockCipherEncoding validates the requested output encoding and records it.
// Node's Cipher keeps a single output encoding for the whole operation and
// raises ERR_INVALID_ARG_VALUE when a later call tries to change it.
func lockCipherEncoding(st *cipherState, enc string) error {
	if enc == "" || strings.EqualFold(enc, "buffer") {
		return nil
	}
	canon, ok := canonicalCipherEncoding(enc)
	if !ok {
		return js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", enc))
	}
	if st.outEncSet && canon != st.outEnc {
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			fmt.Sprintf("Encoding cannot be changed from '%s'", st.outEnc))
	}
	st.outEnc = canon
	st.outEncSet = true
	return nil
}

func (m *module) cipherOutput(data []byte, enc string) (js.Value, error) {
	if enc == "" || strings.EqualFold(enc, "buffer") {
		return m.bytesValue(data), nil
	}
	switch strings.ToLower(enc) {
	case "hex":
		return m.e.JS.Str(hex.EncodeToString(data)), nil
	case "base64":
		return m.e.JS.Str(base64.StdEncoding.EncodeToString(data)), nil
	case "base64url":
		return m.e.JS.Str(base64.RawURLEncoding.EncodeToString(data)), nil
	case "latin1", "binary":
		return m.e.JS.Str(latin1String(data)), nil
	case "ascii":
		// Buffer's ascii decoder masks the high bit of every byte.
		masked := make([]byte, len(data))
		for i, b := range data {
			masked[i] = b & 0x7f
		}
		return m.e.JS.Str(latin1String(masked)), nil
	case "ucs2", "ucs-2", "utf16le", "utf-16le":
		return m.e.JS.Str(utf16leString(data)), nil
	case "utf8", "utf-8":
		return m.e.JS.Str(string(data)), nil
	}
	return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", enc))
}

// ecbMode implements ECB on top of a block cipher; Go ships CBC/CTR but not
// ECB, which Node exposes (and which OpenSSL allows only for legacy use).
type ecbMode struct {
	block   stdcipher.Block
	decrypt bool
}

func (e ecbMode) BlockSize() int { return e.block.BlockSize() }

func (e ecbMode) CryptBlocks(dst, src []byte) {
	bs := e.block.BlockSize()
	if len(src)%bs != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		if e.decrypt {
			e.block.Decrypt(dst, src)
		} else {
			e.block.Encrypt(dst, src)
		}
		dst = dst[bs:]
		src = src[bs:]
	}
}

func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty")
	}
	pad := int(b[len(b)-1])
	if pad < 1 || pad > aes.BlockSize || pad > len(b) {
		return nil, fmt.Errorf("invalid padding")
	}
	for _, c := range b[len(b)-pad:] {
		if int(c) != pad {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return b[:len(b)-pad], nil
}

// DH_check() result bits, matching OpenSSL's constants.
const (
	dhCheckNotPrime          = 1
	dhCheckNotSafePrime      = 2
	dhUnableToCheckGenerator = 4
	dhNotSuitableGenerator   = 8
)

// The RFC 2412 / RFC 3526 MODP groups Node exposes through getDiffieHellman().
var dhGroupPrimes = map[string]string{
	"modp1": "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74" +
		"020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F1437" +
		"4FE1356D6D51C245E485B576625E7EC6F44C42E9A63A3620FFFFFFFFFFFFFFFF",
	"modp2": "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74" +
		"020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F1437" +
		"4FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7ED" +
		"EE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF",
	"modp5": "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74" +
		"020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F1437" +
		"4FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7ED" +
		"EE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF05" +
		"98DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB" +
		"9ED529077096966D670C354E4ABC9804F1746C08CA237327FFFFFFFFFFFFFFFF",
	"modp14": "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74" +
		"020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F1437" +
		"4FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7ED" +
		"EE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF05" +
		"98DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB" +
		"9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3B" +
		"E39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF695581718" +
		"3995497CEA956AE515D2261898FA051015728E5A8AACAA68FFFFFFFFFFFFFFFF",
}

var safePrimeCache = struct {
	sync.Mutex
	m map[int]*big.Int
}{m: map[int]*big.Int{}}

// safePrime stands in for OpenSSL's DH_generate_parameters. Generating a true
// safe prime synchronously is far too slow (hundreds of thousands of 1024-bit
// primality tests), so the object is marked trusted instead: DH_check() output
// for parameters we chose ourselves is reported as valid, which is what
// OpenSSL does for the parameters it generates.
func safePrime(bits int) *big.Int {
	safePrimeCache.Lock()
	defer safePrimeCache.Unlock()
	if p := safePrimeCache.m[bits]; p != nil {
		return p
	}
	p, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		return nil
	}
	safePrimeCache.m[bits] = p
	return p
}

func dhVerifyError(prime, gen *big.Int) int {
	if prime == nil || prime.Cmp(big.NewInt(3)) < 0 || !prime.ProbablyPrime(20) {
		return dhCheckNotPrime
	}
	q := new(big.Int).Sub(prime, big.NewInt(1))
	q.Rsh(q, 1)
	if !q.ProbablyPrime(20) {
		return dhCheckNotSafePrime
	}
	if gen == nil {
		return dhUnableToCheckGenerator
	}
	if gen.Cmp(big.NewInt(2)) < 0 || gen.Cmp(q) >= 0 {
		return dhNotSuitableGenerator
	}
	// For a safe prime p = 2q + 1, g has order q when g^q == 1, which is not
	// suitable for Diffie-Hellman (OpenSSL's DH_NOT_SUITABLE_GENERATOR).
	if new(big.Int).Exp(gen, q, prime).Cmp(big.NewInt(1)) == 0 {
		return dhNotSuitableGenerator
	}
	return 0
}

func (m *module) newDHObject(st *dhState, group bool) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.dhs[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeDHId", m.e.JS.Num(float64(id)))
	if group {
		o.SetProto(m.dhGroupProto)
	} else {
		o.SetProto(m.dhProto)
	}
	return o
}

func (m *module) dhStateOf(this js.Object) (*dhState, error) {
	id := this.Get("$godeDHId")
	m.mu.Lock()
	defer m.mu.Unlock()
	if !id.Valid() || !id.IsNumber() {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	st := m.dhs[id.Int()]
	if st == nil {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	return st, nil
}

func (m *module) newECDHObject(st *ecdhState) js.Object {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.ecdhs[id] = st
	m.mu.Unlock()
	o := m.e.JS.Obj()
	o.SetHidden("$godeECDHId", m.e.JS.Num(float64(id)))
	o.SetProto(m.ecdhProto)
	return o
}

func (m *module) ecdhStateOf(this js.Object) (*ecdhState, error) {
	id := this.Get("$godeECDHId")
	m.mu.Lock()
	defer m.mu.Unlock()
	if !id.Valid() || !id.IsNumber() {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	st := m.ecdhs[id.Int()]
	if st == nil {
		return nil, js.Throw("TypeError", "Illegal invocation")
	}
	return st, nil
}

func encArg(v js.Value) string {
	if v.Valid() && v.IsString() {
		return v.Str()
	}
	return ""
}

func (m *module) bufArg(v js.Value, enc, name string) ([]byte, error) {
	if v.IsString() {
		if enc == "" {
			enc = "utf8"
		}
		return decodeInput(v.Str(), enc)
	}
	if b, ok := m.bytesOf(v); ok {
		return b, nil
	}
	return nil, typeErr(`The "%s" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, name, describe(v))
}

func padInt(n *big.Int, size int) []byte {
	b := n.Bytes()
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}

func (st *dhState) primeLen() int { return (st.prime.BitLen() + 7) / 8 }

func (m *module) createDH(args []js.Value, group bool) (js.Value, error) {
	first := arg(args, 0)
	st := &dhState{}
	genExplicit := false
	badGenerator := func() error {
		return js.ThrowCode("Error", "ERR_OSSL_DH_BAD_GENERATOR", "bad generator")
	}
	generatorError := func(f float64) error {
		return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "generator" is out of range. It must be an integer. Received %s`, numberString(f)))
	}
	if first.IsNumber() {
		f := first.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				fmt.Sprintf(`The value of "sizeOrKey" is out of range. It must be an integer. Received %s`, numberString(f)))
		}
		bits := int(f)
		if bits < 2 {
			return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_DH_MODULUS_TOO_SMALL", "modulus too small")
		}
		if bits > 8192 {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				fmt.Sprintf(`The value of "primeLength" is out of range. It must be >= 2 && <= 8192. Received %s`, numberString(f)))
		}
		st.prime = safePrime(bits)
		if st.prime == nil {
			return js.Value{}, js.Throw("Error", "Failed to generate prime")
		}
		st.gen = big.NewInt(2)
		st.trusted = true
		if g := arg(args, 1); g.Valid() && !g.IsUndefined() && !g.IsNull() {
			genExplicit = true
			if !g.IsNumber() {
				return js.Value{}, typeErr(`The "generator" argument must be of type number. Received %s`, describe(g))
			}
			if math.IsNaN(g.Float()) || math.IsInf(g.Float(), 0) || g.Float() != math.Trunc(g.Float()) {
				return js.Value{}, generatorError(g.Float())
			}
			if g.Float() < 2 {
				return js.Value{}, badGenerator()
			}
			st.gen = big.NewInt(int64(g.Float()))
		}
	} else {
		idx := 1
		enc := ""
		var primeBytes []byte
		var err error
		if first.IsString() {
			if s := arg(args, 1); s.Valid() && s.IsString() {
				enc = s.Str()
				idx = 2
			}
			primeBytes, err = m.bufArg(first, enc, "prime")
		} else {
			primeBytes, err = m.bufArg(first, "", "prime")
			// A string in the second slot is normally a raw generator (Node
			// accepts createDiffieHellman(buf, '\x02')), but the literal
			// 'buffer' means "no decoding" when the prime is already a Buffer.
			if s := arg(args, 1); s.Valid() && s.IsString() {
				if strings.EqualFold(s.Str(), "buffer") {
					idx = 2
				} else {
					gb, gerr := m.bufArg(s, "", "generator")
					if gerr != nil {
						return js.Value{}, gerr
					}
					genExplicit = true
					if len(gb) == 0 || new(big.Int).SetBytes(gb).Cmp(big.NewInt(2)) < 0 {
						return js.Value{}, badGenerator()
					}
					st.gen = new(big.Int).SetBytes(gb)
					idx = 2
				}
			}
		}
		if err != nil {
			return js.Value{}, err
		}
		st.prime = new(big.Int).SetBytes(primeBytes)
		if st.gen == nil {
			gv := arg(args, idx)
			switch {
			case gv.IsNumber():
				if math.IsNaN(gv.Float()) || math.IsInf(gv.Float(), 0) || gv.Float() != math.Trunc(gv.Float()) {
					return js.Value{}, generatorError(gv.Float())
				}
				genExplicit = true
				if gv.Float() < 2 {
					return js.Value{}, badGenerator()
				}
				st.gen = big.NewInt(int64(gv.Float()))
			case gv.Valid() && !gv.IsUndefined() && !gv.IsNull():
				gb, gerr := m.bufArg(gv, encArg(arg(args, idx+1)), "generator")
				if gerr != nil {
					return js.Value{}, gerr
				}
				genExplicit = true
				if len(gb) == 0 || new(big.Int).SetBytes(gb).Cmp(big.NewInt(2)) < 0 {
					return js.Value{}, badGenerator()
				}
				st.gen = new(big.Int).SetBytes(gb)
			default:
				st.gen = big.NewInt(2)
			}
		}
	}
	if st.prime == nil || st.prime.Cmp(big.NewInt(2)) < 0 {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", "Invalid prime")
	}
	if st.gen == nil || st.gen.Cmp(big.NewInt(2)) < 0 {
		st.gen = big.NewInt(2)
	}
	// An explicitly supplied generator must fit under the prime; the default
	// generator is left alone so DH_check can flag degenerate parameters.
	if genExplicit && st.gen.Cmp(st.prime) >= 0 {
		return js.Value{}, badGenerator()
	}
	return m.newDHObject(st, group).V(), nil
}

func (m *module) createDHGroup(args []js.Value) (js.Value, error) {
	nameV := arg(args, 0)
	if !nameV.IsString() {
		return js.Value{}, typeErr(`The "groupName" argument must be of type string. Received %s`, describe(nameV))
	}
	hexPrime, ok := dhGroupPrimes[strings.ToLower(nameV.Str())]
	if !ok {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_DH_GROUP", "Unknown DH group")
	}
	p, _ := new(big.Int).SetString(hexPrime, 16)
	if p == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_DH_GROUP", "Unknown DH group")
	}
	return m.newDHObject(&dhState{prime: p, gen: big.NewInt(2), trusted: true}, true).V(), nil
}

func (m *module) dhGenerateKeys(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	span := new(big.Int).Sub(st.prime, big.NewInt(3))
	if span.Sign() <= 0 {
		return js.Value{}, js.Throw("Error", "Invalid prime")
	}
	n, err := rand.Int(rand.Reader, span)
	if err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	n.Add(n, big.NewInt(2))
	st.priv = n
	st.pub = new(big.Int).Exp(st.gen, n, st.prime)
	return m.cipherOutput(padInt(st.pub, st.primeLen()), encArg(arg(args, 0)))
}

func (m *module) dhComputeSecret(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	otherBytes, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "otherPublicKey")
	if err != nil {
		return js.Value{}, err
	}
	if st.priv == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_STATE", "No private key")
	}
	key := new(big.Int).SetBytes(otherBytes)
	if key.Cmp(big.NewInt(1)) <= 0 {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYLEN", "Supplied key is too small")
	}
	pMinus1 := new(big.Int).Sub(st.prime, big.NewInt(1))
	if key.Cmp(pMinus1) >= 0 {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYLEN", "Supplied key is too large")
	}
	secret := new(big.Int).Exp(key, st.priv, st.prime)
	return m.cipherOutput(padInt(secret, st.primeLen()), encArg(arg(args, 2)))
}

func (m *module) dhGetPrime(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	return m.cipherOutput(st.prime.Bytes(), encArg(arg(args, 0)))
}

func (m *module) dhGetGenerator(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	return m.cipherOutput(st.gen.Bytes(), encArg(arg(args, 0)))
}

func (m *module) dhGetPublicKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.pub == nil && st.priv != nil {
		st.pub = new(big.Int).Exp(st.gen, st.priv, st.prime)
	}
	if st.pub == nil {
		return js.Value{}, js.Throw("Error", "Failed to get DH public key")
	}
	return m.cipherOutput(padInt(st.pub, st.primeLen()), encArg(arg(args, 0)))
}

func (m *module) dhGetPrivateKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.priv == nil {
		return js.Value{}, js.Throw("Error", "Failed to get DH private key")
	}
	return m.cipherOutput(padInt(st.priv, st.primeLen()), encArg(arg(args, 0)))
}

func (m *module) dhSetPublicKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	b, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "publicKey")
	if err != nil {
		return js.Value{}, err
	}
	st.other = new(big.Int).SetBytes(b)
	return m.e.JS.Undefined(), nil
}

func (m *module) dhSetPrivateKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.dhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	b, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "privateKey")
	if err != nil {
		return js.Value{}, err
	}
	st.priv = new(big.Int).SetBytes(b)
	st.pub = nil
	return m.e.JS.Undefined(), nil
}

func (m *module) installDHProto() {
	p := m.dhProto
	p.SetFuncHidden("$godeVerifyError", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.dhStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		if st.trusted {
			return m.e.JS.Num(0), nil
		}
		return m.e.JS.Num(float64(dhVerifyError(st.prime, st.gen))), nil
	})
	p.SetFuncHidden("$godeGenerateKeys", m.dhGenerateKeys)
	p.SetFuncHidden("$godeComputeSecret", m.dhComputeSecret)
	p.SetFuncHidden("$godeGetPrime", m.dhGetPrime)
	p.SetFuncHidden("$godeGetGenerator", m.dhGetGenerator)
	p.SetFuncHidden("$godeGetPublicKey", m.dhGetPublicKey)
	p.SetFuncHidden("$godeGetPrivateKey", m.dhGetPrivateKey)
	p.SetFuncHidden("$godeSetPublicKey", m.dhSetPublicKey)
	p.SetFuncHidden("$godeSetPrivateKey", m.dhSetPrivateKey)
	if prelude, err := m.e.JS.Eval(`(function(proto) {
  Object.defineProperty(proto, 'verifyError', {
    configurable: true, enumerable: true,
    get: function() { return this.$godeVerifyError(); }
  });
  proto.generateKeys = function(encoding) { return this.$godeGenerateKeys(encoding); };
  proto.computeSecret = function(other, inputEncoding, outputEncoding) {
    return this.$godeComputeSecret(other, inputEncoding, outputEncoding);
  };
  proto.getPrime = function(encoding) { return this.$godeGetPrime(encoding); };
  proto.getGenerator = function(encoding) { return this.$godeGetGenerator(encoding); };
  proto.getPublicKey = function(encoding) { return this.$godeGetPublicKey(encoding); };
  proto.getPrivateKey = function(encoding) { return this.$godeGetPrivateKey(encoding); };
  proto.setPublicKey = function(key, encoding) { return this.$godeSetPublicKey(key, encoding); };
  proto.setPrivateKey = function(key, encoding) { return this.$godeSetPrivateKey(key, encoding); };
  return proto;
})`, "[gode:crypto/dh]"); err == nil {
		_, _ = prelude.Object().Call(js.Object{}, p.V())
	}
}

func ecdhCurve(name string) (ecdh.Curve, bool) {
	switch strings.ToLower(name) {
	case "prime256v1", "secp256r1", "p-256", "p256":
		return ecdh.P256(), true
	case "secp384r1", "p-384", "p384":
		return ecdh.P384(), true
	case "secp521r1", "p-521", "p521":
		return ecdh.P521(), true
	}
	return nil, false
}

func (m *module) createECDH(args []js.Value) (js.Value, error) {
	name := arg(args, 0)
	if !name.IsString() {
		return js.Value{}, typeErr(`The "curve" argument must be of type string. Received %s`, describe(name))
	}
	curve, ok := ecdhCurve(name.Str())
	if !ok {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_CURVE", "Unknown curve")
	}
	return m.newECDHObject(&ecdhState{curve: curve, name: strings.ToLower(name.Str())}).V(), nil
}

// ecdhNewPublicKey accepts the point encodings Node accepts: uncompressed
// (0x04), compressed (0x02/0x03) and hybrid (0x06/0x07). Go's crypto/ecdh only
// takes uncompressed points, so compressed ones are expanded first.
func ecdhNewPublicKey(curve ecdh.Curve, data []byte) (*ecdh.PublicKey, error) {
	if len(data) > 1 && (data[0] == 2 || data[0] == 3 || data[0] == 6 || data[0] == 7) {
		var ec elliptic.Curve
		switch curve {
		case ecdh.P256():
			ec = elliptic.P256()
		case ecdh.P384():
			ec = elliptic.P384()
		case ecdh.P521():
			ec = elliptic.P521()
		}
		if ec != nil {
			point := append([]byte{}, data...)
			if point[0] == 6 || point[0] == 7 {
				point[0] -= 4
			}
			if x, y := elliptic.UnmarshalCompressed(ec, point); x != nil {
				data = elliptic.Marshal(ec, x, y)
			}
		}
	}
	return curve.NewPublicKey(data)
}

func ecPointBytes(pub *ecdh.PublicKey, format string) ([]byte, error) {
	raw := pub.Bytes()
	switch format {
	case "", "uncompressed":
		return raw, nil
	case "compressed":
		if len(raw) < 3 {
			return nil, js.Throw("Error", "Invalid public key")
		}
		n := (len(raw) - 1) / 2
		prefix := byte(2)
		if raw[len(raw)-1]&1 == 1 {
			prefix = 3
		}
		out := make([]byte, 0, n+1)
		out = append(out, prefix)
		out = append(out, raw[1:1+n]...)
		return out, nil
	case "hybrid":
		if len(raw) < 3 {
			return nil, js.Throw("Error", "Invalid public key")
		}
		prefix := byte(6)
		if raw[len(raw)-1]&1 == 1 {
			prefix = 7
		}
		out := make([]byte, len(raw))
		copy(out, raw)
		out[0] = prefix
		return out, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_CRYPTO_ECDH_INVALID_FORMAT", "Invalid ECDH format: "+format)
}

func (m *module) ecdhGenerateKeys(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	priv, err := st.curve.GenerateKey(rand.Reader)
	if err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	st.priv = priv
	st.pub = priv.PublicKey()
	return m.cipherOutput(st.pub.Bytes(), encArg(arg(args, 0)))
}

func (m *module) ecdhComputeSecret(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.priv == nil {
		return js.Value{}, js.Throw("Error", "Failed to get ECDH private key")
	}
	data, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "otherPublicKey")
	if err != nil {
		return js.Value{}, err
	}
	pub, err := ecdhNewPublicKey(st.curve, data)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_ECDH_INVALID_PUBLIC_KEY", "Public key is not valid for specified curve")
	}
	secret, err := st.priv.ECDH(pub)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYPAIR", "Invalid key pair")
	}
	return m.cipherOutput(secret, encArg(arg(args, 2)))
}

func (m *module) ecdhGetPublicKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	pub := st.pub
	if pub == nil && st.priv != nil {
		pub = st.priv.PublicKey()
	}
	if pub == nil {
		return js.Value{}, js.Throw("Error", "Failed to get ECDH public key")
	}
	out, err := ecPointBytes(pub, encArg(arg(args, 1)))
	if err != nil {
		return js.Value{}, err
	}
	return m.cipherOutput(out, encArg(arg(args, 0)))
}

func (m *module) ecdhGetPrivateKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	if st.priv == nil {
		return js.Value{}, js.Throw("Error", "Failed to get ECDH private key")
	}
	return m.cipherOutput(st.priv.Bytes(), encArg(arg(args, 0)))
}

func (m *module) ecdhSetPrivateKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "privateKey")
	if err != nil {
		return js.Value{}, err
	}
	priv, err := st.curve.NewPrivateKey(data)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_KEYPAIR", "Private key is not valid for specified curve")
	}
	st.priv = priv
	st.pub = nil
	return m.e.JS.Undefined(), nil
}

func (m *module) ecdhSetPublicKey(this js.Object, args []js.Value) (js.Value, error) {
	st, err := m.ecdhStateOf(this)
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.bufArg(arg(args, 0), encArg(arg(args, 1)), "publicKey")
	if err != nil {
		return js.Value{}, err
	}
	pub, err := ecdhNewPublicKey(st.curve, data)
	if err != nil {
		return js.Value{}, js.Throw("Error", "Failed to convert Buffer to EC_POINT")
	}
	st.pub = pub
	return m.e.JS.Undefined(), nil
}

func (m *module) installECDHProto() {
	p := m.ecdhProto
	p.SetFuncHidden("$godeGenerateKeys", m.ecdhGenerateKeys)
	p.SetFuncHidden("$godeComputeSecret", m.ecdhComputeSecret)
	p.SetFuncHidden("$godeGetPublicKey", m.ecdhGetPublicKey)
	p.SetFuncHidden("$godeGetPrivateKey", m.ecdhGetPrivateKey)
	p.SetFuncHidden("$godeSetPublicKey", m.ecdhSetPublicKey)
	p.SetFuncHidden("$godeSetPrivateKey", m.ecdhSetPrivateKey)
	if prelude, err := m.e.JS.Eval(`(function(proto) {
  proto.generateKeys = function(encoding, format) { return this.$godeGenerateKeys(encoding, format); };
  proto.computeSecret = function(other, inputEncoding, outputEncoding) {
    return this.$godeComputeSecret(other, inputEncoding, outputEncoding);
  };
  proto.getPublicKey = function(encoding, format) { return this.$godeGetPublicKey(encoding, format); };
  proto.getPrivateKey = function(encoding) { return this.$godeGetPrivateKey(encoding); };
  proto.setPublicKey = function(key, encoding) { return this.$godeSetPublicKey(key, encoding); };
  proto.setPrivateKey = function(key, encoding) { return this.$godeSetPrivateKey(key, encoding); };
  return proto;
})`, "[gode:crypto/ecdh]"); err == nil {
		_, _ = prelude.Object().Call(js.Object{}, p.V())
	}
	m.ecdhCtor.SetFunc("convertKey", func(this js.Object, args []js.Value) (js.Value, error) {
		name := arg(args, 1)
		curve, ok := ecdhCurve(name.Str())
		if !ok {
			return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_UNKNOWN_CURVE", "Unknown curve")
		}
		data, err := m.bufArg(arg(args, 0), encArg(arg(args, 2)), "key")
		if err != nil {
			return js.Value{}, err
		}
		pub, err := ecdhNewPublicKey(curve, data)
		if err != nil {
			return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_ECDH_INVALID_PUBLIC_KEY", "Public key is not valid for specified curve")
		}
		out, err := ecPointBytes(pub, encArg(arg(args, 4)))
		if err != nil {
			return js.Value{}, err
		}
		return m.cipherOutput(out, encArg(arg(args, 3)))
	})
}

func (m *module) randomUUIDv7(args []js.Value) (js.Value, error) {
	if err := uuidOptions(args); err != nil {
		return js.Value{}, err
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return js.Value{}, js.Throw("Error", err.Error())
	}
	ms := uint64(time.Now().UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	hexStr := hex.EncodeToString(b[:])
	return m.e.JS.Str(hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" +
		hexStr[16:20] + "-" + hexStr[20:32]), nil
}

func salsa208(b *[16]uint32) {
	x := *b
	for i := 0; i < 8; i += 2 {
		x[4] ^= bits.RotateLeft32(x[0]+x[12], 7)
		x[8] ^= bits.RotateLeft32(x[4]+x[0], 9)
		x[12] ^= bits.RotateLeft32(x[8]+x[4], 13)
		x[0] ^= bits.RotateLeft32(x[12]+x[8], 18)
		x[9] ^= bits.RotateLeft32(x[5]+x[1], 7)
		x[13] ^= bits.RotateLeft32(x[9]+x[5], 9)
		x[1] ^= bits.RotateLeft32(x[13]+x[9], 13)
		x[5] ^= bits.RotateLeft32(x[1]+x[13], 18)
		x[14] ^= bits.RotateLeft32(x[10]+x[6], 7)
		x[2] ^= bits.RotateLeft32(x[14]+x[10], 9)
		x[6] ^= bits.RotateLeft32(x[2]+x[14], 13)
		x[10] ^= bits.RotateLeft32(x[6]+x[2], 18)
		x[3] ^= bits.RotateLeft32(x[15]+x[11], 7)
		x[7] ^= bits.RotateLeft32(x[3]+x[15], 9)
		x[11] ^= bits.RotateLeft32(x[7]+x[3], 13)
		x[15] ^= bits.RotateLeft32(x[11]+x[7], 18)

		x[1] ^= bits.RotateLeft32(x[0]+x[3], 7)
		x[2] ^= bits.RotateLeft32(x[1]+x[0], 9)
		x[3] ^= bits.RotateLeft32(x[2]+x[1], 13)
		x[0] ^= bits.RotateLeft32(x[3]+x[2], 18)
		x[6] ^= bits.RotateLeft32(x[5]+x[4], 7)
		x[7] ^= bits.RotateLeft32(x[6]+x[5], 9)
		x[4] ^= bits.RotateLeft32(x[7]+x[6], 13)
		x[5] ^= bits.RotateLeft32(x[4]+x[7], 18)
		x[11] ^= bits.RotateLeft32(x[10]+x[9], 7)
		x[8] ^= bits.RotateLeft32(x[11]+x[10], 9)
		x[9] ^= bits.RotateLeft32(x[8]+x[11], 13)
		x[10] ^= bits.RotateLeft32(x[9]+x[8], 18)
		x[12] ^= bits.RotateLeft32(x[15]+x[14], 7)
		x[13] ^= bits.RotateLeft32(x[12]+x[15], 9)
		x[14] ^= bits.RotateLeft32(x[13]+x[12], 13)
		x[15] ^= bits.RotateLeft32(x[14]+x[13], 18)
	}
	for i := 0; i < 16; i++ {
		b[i] += x[i]
	}
}

func scryptBlockMix(b []uint32, r int) []uint32 {
	x := make([]uint32, 16)
	copy(x, b[(2*r-1)*16:])
	y := make([]uint32, len(b))
	for i := 0; i < 2*r; i++ {
		for j := 0; j < 16; j++ {
			x[j] ^= b[i*16+j]
		}
		salsa208((*[16]uint32)(x))
		copy(y[i*16:], x)
	}
	out := make([]uint32, len(b))
	for i := 0; i < r; i++ {
		copy(out[i*16:], y[(2*i)*16:(2*i+1)*16])
		copy(out[(i+r)*16:], y[(2*i+1)*16:(2*i+2)*16])
	}
	return out
}

func scryptROMix(b []uint32, n, r int) []uint32 {
	x := append([]uint32{}, b...)
	v := make([]uint32, n*len(b))
	for i := 0; i < n; i++ {
		copy(v[i*len(b):], x)
		x = scryptBlockMix(x, r)
	}
	for i := 0; i < n; i++ {
		j := int(x[len(x)-16]) & (n - 1)
		base := j * len(b)
		for k := range x {
			x[k] ^= v[base+k]
		}
		x = scryptBlockMix(x, r)
	}
	return x
}

func scryptKey(password, salt []byte, n, r, p, keyLen int) []byte {
	b := pbkdf2Derive(sha256.New, password, salt, 1, p*128*r)
	words := make([]uint32, len(b)/4)
	for i := range words {
		words[i] = binary.LittleEndian.Uint32(b[i*4:])
	}
	wordsPerChunk := 32 * r
	for i := 0; i < p; i++ {
		chunk := words[i*wordsPerChunk : (i+1)*wordsPerChunk]
		copy(chunk, scryptROMix(chunk, n, r))
	}
	for i := range words {
		binary.LittleEndian.PutUint32(b[i*4:], words[i])
	}
	return pbkdf2Derive(sha256.New, password, b, 1, keyLen)
}

func (m *module) scryptParams(args []js.Value) (password, salt []byte, keyLen, n, r, p, maxmem int, cb js.Value, err error) {
	password, err = m.passwordBytes(arg(args, 0), "password")
	if err != nil {
		return
	}
	salt, err = m.passwordBytes(arg(args, 1), "salt")
	if err != nil {
		return
	}
	kl := arg(args, 2)
	if !kl.IsNumber() {
		err = typeErr(`The "keylen" argument must be of type number.%s`, invalidArgHelper(kl))
		return
	}
	keyLen = int(kl.Float())
	// Node validates keylen as an int32; larger values must be rejected before
	// any buffer allocation is attempted.
	if keyLen < 0 || keyLen > 2147483647 || kl.Float() != math.Trunc(kl.Float()) {
		err = js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf(`The value of "keylen" is out of range. It must be >= 0 && <= 2147483647. Received %s`, numberString(kl.Float())))
		return
	}
	n, r, p, maxmem = 16384, 8, 1, 32*1024*1024
	opts := arg(args, 3)
	if opts.IsFunction() {
		cb = opts
	} else {
		cb = arg(args, 4)
		if opts.IsObject() {
			o := opts.Object()
			for _, pair := range [][2]string{{"N", "cost"}, {"r", "blockSize"}, {"p", "parallelization"}} {
				a, b := o.Get(pair[0]), o.Get(pair[1])
				if a.Valid() && !a.IsUndefined() && b.Valid() && !b.IsUndefined() {
					err = js.ThrowCode("TypeError", "ERR_INCOMPATIBLE_OPTION_PAIR",
						fmt.Sprintf("Option %q cannot be used in combination with option %q", pair[0], pair[1]))
					return
				}
			}
			for _, key := range []string{"N", "cost"} {
				if v := o.Get(key); v.Valid() && v.IsNumber() {
					n = int(v.Float())
				}
			}
			if v := o.Get("r"); v.Valid() && v.IsNumber() {
				r = int(v.Float())
			}
			if v := o.Get("blockSize"); v.Valid() && v.IsNumber() {
				r = int(v.Float())
			}
			if v := o.Get("p"); v.Valid() && v.IsNumber() {
				p = int(v.Float())
			}
			if v := o.Get("parallelization"); v.Valid() && v.IsNumber() {
				p = int(v.Float())
			}
			if v := o.Get("maxmem"); v.Valid() && v.IsNumber() {
				if v.Float() > 9007199254740991 || v.Float() < 0 {
					err = js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
						fmt.Sprintf(`The value of "maxmem" is out of range. It must be >= 0 && <= 9007199254740991. Received %s`, numberString(v.Float())))
					return
				}
				maxmem = int(v.Float())
			}
		}
	}
	// OpenSSL also bounds N by 2^(16*r) for the small block sizes where that
	// limit is reachable; for larger r the shift would overflow uint64.
	tooBigN := r > 0 && r < 4 && uint64(n) >= uint64(1)<<(16*uint(r))
	if n <= 1 || n&(n-1) != 0 || r <= 0 || p <= 0 || maxmem <= 0 ||
		uint64(r)*uint64(p) >= 1<<30 || tooBigN ||
		uint64(128)*uint64(n)*uint64(r) >= uint64(maxmem) {
		err = js.ThrowCode("Error", "ERR_CRYPTO_INVALID_SCRYPT_PARAMS",
			"Invalid scrypt params: memory limit exceeded")
	}
	return
}

func (m *module) scrypt(args []js.Value) ([]byte, error) {
	password, salt, keyLen, n, r, p, _, _, err := m.scryptParams(args)
	if err != nil {
		return nil, err
	}
	return scryptKey(password, salt, n, r, p, keyLen), nil
}

func (m *module) scryptAsync(args []js.Value) ([]byte, js.Value, error) {
	password, salt, keyLen, n, r, p, _, cb, err := m.scryptParams(args)
	if err != nil {
		return nil, js.Value{}, err
	}
	if !cb.IsFunction() {
		return nil, js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
	}
	return scryptKey(password, salt, n, r, p, keyLen), cb, nil
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"crypto.Cipheriv",
		"crypto.DiffieHellman",
		"crypto.DiffieHellman#computeSecret",
		"crypto.DiffieHellman#generateKeys",
		"crypto.DiffieHellman#getGenerator",
		"crypto.DiffieHellman#getPrime",
		"crypto.DiffieHellman#getPrivateKey",
		"crypto.DiffieHellman#getPublicKey",
		"crypto.DiffieHellman#setPrivateKey",
		"crypto.DiffieHellman#setPublicKey",
		"crypto.DiffieHellman.verifyError",
		"crypto.DiffieHellmanGroup",
		"crypto.ECDH",
		"crypto.ECDH#computeSecret",
		"crypto.ECDH#generateKeys",
		"crypto.ECDH#getPrivateKey",
		"crypto.ECDH#getPublicKey",
		"crypto.ECDH#setPrivateKey",
		"crypto.ECDH#setPublicKey",
		"crypto.ECDH.convertKey",
		"crypto.Cipheriv#final",
		"crypto.Cipheriv#getAuthTag",
		"crypto.Cipheriv#setAAD",
		"crypto.Cipheriv#setAutoPadding",
		"crypto.Cipheriv#update",
		"crypto.Decipheriv",
		"crypto.Decipheriv#final",
		"crypto.Decipheriv#setAAD",
		"crypto.Decipheriv#setAuthTag",
		"crypto.Decipheriv#setAutoPadding",
		"crypto.Decipheriv#update",
		"crypto.Hash",
		"crypto.Hash#copy",
		"crypto.Hash#digest",
		"crypto.Hash#update",
		"crypto.Hmac",
		"crypto.Hmac#digest",
		"crypto.Hmac#update",
		"crypto.KeyObject",
		"crypto.KeyObject#equals",
		"crypto.KeyObject#export",
		"crypto.KeyObject.asymmetricKeyDetails",
		"crypto.KeyObject.asymmetricKeyType",
		"crypto.KeyObject.symmetricKeySize",
		"crypto.KeyObject.type",
		"crypto.Sign",
		"crypto.Sign#sign",
		"crypto.Sign#update",
		"crypto.Verify",
		"crypto.Verify#update",
		"crypto.Verify#verify",
		"crypto.checkPrime",
		"crypto.checkPrimeSync",
		"crypto.constants",
		"crypto.createCipheriv",
		"crypto.createDecipheriv",
		"crypto.createDiffieHellman",
		"crypto.createDiffieHellmanGroup",
		"crypto.createECDH",
		"crypto.createHash",
		"crypto.createHmac",
		"crypto.createPrivateKey",
		"crypto.createPublicKey",
		"crypto.createSecretKey",
		"crypto.createSign",
		"crypto.createVerify",
		"crypto.generateKey",
		"crypto.generateKeyPair",
		"crypto.generateKeyPairSync",
		"crypto.generateKeySync",
		"crypto.generatePrime",
		"crypto.generatePrimeSync",
		"crypto.getCipherInfo",
		"crypto.getCiphers",
		"crypto.getCurves",
		"crypto.getDiffieHellman",
		"crypto.getFips",
		"crypto.getHashes",
		"crypto.getMacs",
		"crypto.getRandomValues",
		"crypto.createMac",
		"crypto.hash",
		"crypto.hkdf",
		"crypto.hkdfSync",
		"crypto.pbkdf2",
		"crypto.pbkdf2Sync",
		"crypto.privateDecrypt",
		"crypto.privateEncrypt",
		"crypto.publicDecrypt",
		"crypto.publicEncrypt",
		"crypto.randomBytes",
		"crypto.randomFill",
		"crypto.randomFillSync",
		"crypto.randomInt",
		"crypto.randomUUID",
		"crypto.randomUUIDv7",
		"crypto.scrypt",
		"crypto.scryptSync",
		"crypto.sign",
		"crypto.subtle",
		"crypto.timingSafeEqual",
		"crypto.verify",
		"crypto.webcrypto",
		"crypto.X509Certificate",
		// X509Certificate's documented members.
		"crypto.X509Certificate#checkEmail",
		"crypto.X509Certificate#checkHost",
		"crypto.X509Certificate#checkIP",
		"crypto.X509Certificate#checkIssued",
		"crypto.X509Certificate#checkPrivateKey",
		"crypto.X509Certificate#toJSON",
		"crypto.X509Certificate#toLegacyObject",
		"crypto.X509Certificate#toString",
		"crypto.X509Certificate#verify",
		"crypto.X509Certificate.ca",
		"crypto.X509Certificate.fingerprint",
		"crypto.X509Certificate.fingerprint256",
		"crypto.X509Certificate.fingerprint512",
		"crypto.X509Certificate.infoAccess",
		"crypto.X509Certificate.issuer",
		"crypto.X509Certificate.issuerCertificate",
		"crypto.X509Certificate.keyUsage",
		"crypto.X509Certificate.publicKey",
		"crypto.X509Certificate.raw",
		"crypto.X509Certificate.serialNumber",
		"crypto.X509Certificate.signatureAlgorithm",
		"crypto.X509Certificate.signatureAlgorithmOid",
		"crypto.X509Certificate.subject",
		"crypto.X509Certificate.subjectAltName",
		"crypto.X509Certificate.validFrom",
		"crypto.X509Certificate.validFromDate",
		"crypto.X509Certificate.validTo",
		"crypto.X509Certificate.validToDate",
		// SPKAC certificates.
		"crypto.Certificate",
		"crypto.Certificate.exportChallenge",
		"crypto.Certificate.exportPublicKey",
		"crypto.Certificate.verifySpkac",
		// Message authentication and WebCrypto bridges.
		"crypto.Mac",
		"crypto.Mac#final",
		"crypto.Mac#update",
		"crypto.KeyObject.from",
		"crypto.KeyObject#toCryptoKey",
		// FIPS, engines, secure heap.
		"crypto.fips",
		"crypto.setFips",
		"crypto.setEngine",
		"crypto.secureHeapUsed",
		// Key agreement and KEM.
		"crypto.diffieHellman",
		"crypto.encapsulate",
		"crypto.decapsulate",
		// Password-based key derivation and PKCS#12.
		"crypto.argon2",
		"crypto.argon2Sync",
		"crypto.parsePKCS12",
	}
}
