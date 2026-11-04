package crypto

import (
	stdcrypto "crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	stdmlkem "crypto/mlkem"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pkcs12"

	"godejs/internal/js"
)

// installExtra wires the long tail of Node 24's crypto surface: the SPKAC
// Certificate class, the Mac class name, the OpenSSL-style feature flags, the
// WebCrypto bridge on KeyObject, Argon2, PKCS#12 and Diffie-Hellman.
func (m *module) installExtra(exports js.Object) {
	m.installCertificate(exports)

	// Node exposes the Mac created by createMac under a public class name.
	macCtor := m.class("Mac", m.mac2Proto)
	exports.Set("Mac", macCtor.V())

	// crypto.fips is a getter backed by setFips, not a snapshot.
	if fn, err := m.e.JS.Eval(`(function (exports) {
  Object.defineProperty(exports, 'fips', {
    enumerable: true,
    configurable: true,
    get: function () { return exports.getFips() === 1; },
  });
  return exports;
})`, "[gode:crypto/fips]"); err == nil {
		_, _ = fn.Object().Call(js.Object{}, exports.V())
	}

	exports.SetFunc("secureHeapUsed", func(this js.Object, args []js.Value) (js.Value, error) {
		o := m.e.JS.Obj()
		o.SetNum("total", 0)
		o.SetNum("used", 0)
		o.SetNum("minimum", 0)
		o.SetNum("utilized", 0)
		return o.V(), nil
	})

	exports.SetFunc("setEngine", func(this js.Object, args []js.Value) (js.Value, error) {
		id := arg(args, 0)
		if !id.IsString() {
			return js.Value{}, typeErr(`The "id" argument must be of type string. Received %s`, describe(id))
		}
		if f := arg(args, 1); f.Valid() && !f.IsUndefined() && !f.IsNull() && !f.IsNumber() {
			return js.Value{}, typeErr(`The "flags" argument must be of type number. Received %s`, describe(f))
		}
		// This build has no OpenSSL engine support, which is exactly the case
		// Node reports with ERR_CRYPTO_CUSTOM_ENGINE_NOT_SUPPORTED.
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_CUSTOM_ENGINE_NOT_SUPPORTED",
			"Custom engine not supported")
	})

	exports.SetFunc("diffieHellman", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.diffieHellman(args)
	})

	exports.SetFunc("argon2Sync", func(this js.Object, args []js.Value) (js.Value, error) {
		out, _, err := m.argon2(args, false)
		return out, err
	})
	exports.SetFunc("argon2", func(this js.Object, args []js.Value) (js.Value, error) {
		out, params, err := m.argon2(args, true)
		if err != nil {
			return js.Value{}, err
		}
		cb := arg(args, 2)
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		_ = params
		m.e.NextTick(cb, m.e.JS.Null(), out)
		return js.Value{}, nil
	})

	exports.SetFunc("parsePKCS12", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.parsePKCS12(args)
	})

	// ML-KEM key encapsulation, backed by Go's crypto/mlkem (768 and 1024).
	exports.SetFunc("encapsulate", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.kemEncapsulate(args)
	})
	exports.SetFunc("decapsulate", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.kemDecapsulate(args)
	})

	// KeyObject#toCryptoKey bridges a secret KeyObject to a WebCrypto CryptoKey.
	m.keyProto.SetFuncHidden("toCryptoKey", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.keyStateOf(this)
		if err != nil {
			return js.Value{}, err
		}
		alg := arg(args, 0)
		if !alg.IsObject() {
			return js.Value{}, typeErr(`The "algorithm" argument must be of type object. Received %s`, describe(alg))
		}
		extractable := arg(args, 1)
		if !extractable.IsBool() {
			return js.Value{}, typeErr(`The "extractable" argument must be of type boolean. Received %s`, describe(extractable))
		}
		usages := arg(args, 2)
		if !usages.IsArray() {
			return js.Value{}, typeErr(`The "keyUsages" argument must be an instance of Array. Received %s`, describe(usages))
		}
		if st.kind != "secret" {
			return js.Value{}, js.Throw("NotSupportedError",
				"Only secret keys can be converted to CryptoKey")
		}
		key := m.newCryptoKey(st.secretBytes())
		key.SetStr("type", "secret")
		key.Set("algorithm", alg)
		key.SetBool("extractable", extractable.Bool())
		key.Set("usages", usages)
		return key.V(), nil
	})
}

// ---------------------------------------------------------------------------
// Certificate (NETSCAPE SPKAC)
// ---------------------------------------------------------------------------

type spkacInfo struct {
	challenge []byte
	spki      []byte
	pub       any
	algorithm pkix.AlgorithmIdentifier
	signature []byte
	signed    []byte
}

const certificatePrelude = `(function (host) {
  // OpenSSL refuses SPKAC buffers at or above 2 GiB; Node reports that as an
  // out-of-range range check before any parsing happens.
  function checkSize(spkac) {
    var n = spkac && typeof spkac.length === 'number' ? spkac.length : 0;
    if (n > 0x7fffffff) {
      var e = new RangeError('The value of "spkac" is out of range. It must be <= 2 ** 31 - 1');
      e.code = 'ERR_OUT_OF_RANGE';
      throw e;
    }
  }
  function Certificate() {
    if (!(this instanceof Certificate)) return new Certificate();
  }
  Certificate.prototype.exportChallenge = function (spkac, encoding) {
    checkSize(spkac);
    return host.exportChallenge(spkac, encoding);
  };
  Certificate.prototype.exportPublicKey = function (spkac, encoding) {
    checkSize(spkac);
    return host.exportPublicKey(spkac, encoding);
  };
  Certificate.prototype.verifySpkac = function (spkac) {
    checkSize(spkac);
    return host.verifySpkac(spkac);
  };
  // Node also exposes the SPKAC helpers as static methods.
  Certificate.exportChallenge = Certificate.prototype.exportChallenge;
  Certificate.exportPublicKey = Certificate.prototype.exportPublicKey;
  Certificate.verifySpkac = Certificate.prototype.verifySpkac;
  return Certificate;
})`

func (m *module) installCertificate(exports js.Object) {
	host := m.e.JS.Obj()
	host.SetFunc("exportChallenge", func(this js.Object, args []js.Value) (js.Value, error) {
		info, _ := m.parseSpkac(arg(args, 0))
		if info == nil {
			return m.bytesValue(nil), nil
		}
		return m.spkacOutput(info.challenge, arg(args, 1))
	})
	host.SetFunc("exportPublicKey", func(this js.Object, args []js.Value) (js.Value, error) {
		info, err := m.parseSpkac(arg(args, 0))
		if err != nil {
			return js.Value{}, err
		}
		out := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: info.spki})
		return m.spkacOutput(out, arg(args, 1))
	})
	host.SetFunc("verifySpkac", func(this js.Object, args []js.Value) (js.Value, error) {
		info, err := m.parseSpkac(arg(args, 0))
		if err != nil || info == nil {
			return m.e.JS.Bool(false), nil
		}
		return m.e.JS.Bool(spkacVerify(info)), nil
	})
	fn, err := m.e.JS.Eval(certificatePrelude, "[gode:crypto/certificate]")
	if err != nil {
		return
	}
	ctor, err := fn.Object().Call(js.Object{}, host.V())
	if err != nil {
		return
	}
	exports.Set("Certificate", ctor)
}

// spkacOutput honours the optional encoding argument, defaulting to a Buffer.
func (m *module) spkacOutput(data []byte, enc js.Value) (js.Value, error) {
	if enc.Valid() && !enc.IsUndefined() && !enc.IsNull() {
		if !enc.IsString() {
			return js.Value{}, typeErr(`The "encoding" argument must be of type string. Received %s`, describe(enc))
		}
		if !knownEncoding(enc.Str()) {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				fmt.Sprintf("The argument 'encoding' is invalid. Received '%s'", enc.Str()))
		}
		return m.encodeDigest(data, enc.Str())
	}
	return m.bytesValue(data), nil
}

func (m *module) parseSpkac(v js.Value) (*spkacInfo, error) {
	var data []byte
	if v.IsString() {
		data = []byte(v.Str())
	} else if b, ok := m.bytesOf(v); ok {
		data = b
	} else {
		return nil, typeErr(`The "spkac" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
	}
	var outer struct {
		PKAC      asn1.RawValue
		Algorithm pkix.AlgorithmIdentifier
		Signature asn1.BitString
	}
	if _, err := asn1.Unmarshal(data, &outer); err != nil || len(outer.PKAC.FullBytes) == 0 {
		return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_SPKAC", "Invalid SPKAC")
	}
	var pkc struct {
		SPKI      asn1.RawValue
		Challenge asn1.RawValue `asn1:"optional"`
	}
	if _, err := asn1.Unmarshal(outer.PKAC.FullBytes, &pkc); err != nil || len(pkc.SPKI.FullBytes) == 0 {
		return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_SPKAC", "Invalid SPKAC")
	}
	pub, err := x509.ParsePKIXPublicKey(pkc.SPKI.FullBytes)
	if err != nil {
		return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_SPKAC", "Invalid SPKAC")
	}
	return &spkacInfo{
		challenge: pkc.Challenge.Bytes,
		spki:      pkc.SPKI.FullBytes,
		pub:       pub,
		algorithm: outer.Algorithm,
		signature: outer.Signature.Bytes,
		signed:    outer.PKAC.FullBytes,
	}, nil
}

func spkacVerify(info *spkacInfo) bool {
	oid := info.algorithm.Algorithm
	switch {
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}):
		return spkacRSA(info, stdcrypto.SHA1)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}):
		return spkacRSA(info, stdcrypto.SHA256)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}):
		return spkacRSA(info, stdcrypto.SHA384)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}):
		return spkacRSA(info, stdcrypto.SHA512)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}):
		key, ok := info.pub.(*ecdsa.PublicKey)
		if !ok {
			return false
		}
		h := sha256.Sum256(info.signed)
		return ecdsa.VerifyASN1(key, h[:], info.signature)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 3}):
		key, ok := info.pub.(*ecdsa.PublicKey)
		if !ok {
			return false
		}
		h := sha512.Sum384(info.signed)
		return ecdsa.VerifyASN1(key, h[:], info.signature)
	case oid.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 4}):
		key, ok := info.pub.(*ecdsa.PublicKey)
		if !ok {
			return false
		}
		h := sha512.Sum512(info.signed)
		return ecdsa.VerifyASN1(key, h[:], info.signature)
	case oid.Equal(asn1.ObjectIdentifier{1, 3, 101, 112}):
		key, ok := info.pub.(ed25519.PublicKey)
		if !ok {
			return false
		}
		return ed25519.Verify(key, info.signed, info.signature)
	}
	return false
}

func spkacRSA(info *spkacInfo, hash stdcrypto.Hash) bool {
	key, ok := info.pub.(*rsa.PublicKey)
	if !ok {
		return false
	}
	var digest []byte
	switch hash {
	case stdcrypto.SHA1:
		h := sha1.Sum(info.signed)
		digest = h[:]
	case stdcrypto.SHA256:
		h := sha256.Sum256(info.signed)
		digest = h[:]
	case stdcrypto.SHA384:
		h := sha512.Sum384(info.signed)
		digest = h[:]
	default:
		h := sha512.Sum512(info.signed)
		digest = h[:]
	}
	return rsa.VerifyPKCS1v15(key, hash, digest, info.signature) == nil
}

// ---------------------------------------------------------------------------
// Diffie-Hellman
// ---------------------------------------------------------------------------

func (m *module) diffieHellman(args []js.Value) (js.Value, error) {
	opts := arg(args, 0)
	if !opts.IsObject() {
		return js.Value{}, typeErr(`The "options" argument must be of type object. Received %s`, describe(opts))
	}
	o := opts.Object()
	privV := o.Get("privateKey")
	if !privV.Valid() || privV.IsUndefined() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'options.privateKey' is invalid")
	}
	pubV := o.Get("publicKey")
	if !pubV.Valid() || pubV.IsUndefined() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'options.publicKey' is invalid")
	}
	priv, err := m.keyObjectArg(privV, "options.privateKey")
	if err != nil {
		return js.Value{}, err
	}
	if priv.kind != "private" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'options.privateKey' is invalid")
	}
	pub, err := m.keyObjectArg(pubV, "options.publicKey")
	if err != nil {
		return js.Value{}, err
	}

	var privKey *ecdh.PrivateKey
	switch k := priv.key.(type) {
	case *ecdh.PrivateKey:
		privKey = k
	case *ecdsa.PrivateKey:
		conv, cerr := k.ECDH()
		if cerr != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INCOMPATIBLE_KEY",
				"Incompatible key for Diffie-Hellman")
		}
		privKey = conv
	default:
		return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INCOMPATIBLE_KEY",
			"Incompatible key for Diffie-Hellman")
	}

	var pubKey *ecdh.PublicKey
	switch k := pub.key.(type) {
	case *ecdh.PublicKey:
		pubKey = k
	case *ecdh.PrivateKey:
		pubKey = k.PublicKey()
	case *ecdsa.PrivateKey:
		conv, cerr := k.PublicKey.ECDH()
		if cerr != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INCOMPATIBLE_KEY",
				"Incompatible key for Diffie-Hellman")
		}
		pubKey = conv
	case *ecdsa.PublicKey:
		conv, cerr := k.ECDH()
		if cerr != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INCOMPATIBLE_KEY",
				"Incompatible key for Diffie-Hellman")
		}
		pubKey = conv
	default:
		return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INCOMPATIBLE_KEY",
			"Incompatible key for Diffie-Hellman")
	}

	secret, err := privKey.ECDH(pubKey)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_OPERATION_FAILED", err.Error())
	}
	out := m.bytesValue(secret)
	if cb := arg(args, 1); cb.Valid() && !cb.IsUndefined() {
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		m.e.NextTick(cb, m.e.JS.Null(), out)
		return js.Value{}, nil
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// ML-KEM key encapsulation
// ---------------------------------------------------------------------------

func (m *module) kemEncapsulate(args []js.Value) (js.Value, error) {
	st, err := m.keyObjectArg(arg(args, 0), "key")
	if err != nil {
		return js.Value{}, err
	}
	var shared, ciphertext []byte
	switch k := st.key.(type) {
	case *stdmlkem.EncapsulationKey768:
		shared, ciphertext = k.Encapsulate()
	case *stdmlkem.DecapsulationKey768:
		shared, ciphertext = k.EncapsulationKey().Encapsulate()
	case *stdmlkem.EncapsulationKey1024:
		shared, ciphertext = k.Encapsulate()
	case *stdmlkem.DecapsulationKey1024:
		shared, ciphertext = k.EncapsulationKey().Encapsulate()
	default:
		return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_KEYTYPE",
			"Invalid key type for KEM encapsulation")
	}
	res := m.e.JS.Obj()
	res.Set("sharedKey", m.bytesValue(shared))
	res.Set("ciphertext", m.bytesValue(ciphertext))
	if cb := arg(args, 1); cb.Valid() && !cb.IsUndefined() {
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		m.e.NextTick(cb, m.e.JS.Null(), res.V())
		return js.Value{}, nil
	}
	return res.V(), nil
}

func (m *module) kemDecapsulate(args []js.Value) (js.Value, error) {
	st, err := m.keyObjectArg(arg(args, 0), "key")
	if err != nil {
		return js.Value{}, err
	}
	ciphertext, err := m.bytesArg(arg(args, 1), "ciphertext")
	if err != nil {
		return js.Value{}, err
	}
	var shared []byte
	switch k := st.key.(type) {
	case *stdmlkem.DecapsulationKey768:
		shared, err = k.Decapsulate(ciphertext)
	case *stdmlkem.DecapsulationKey1024:
		shared, err = k.Decapsulate(ciphertext)
	default:
		return js.Value{}, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_KEYTYPE",
			"Invalid key type for KEM decapsulation")
	}
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_OPERATION_FAILED", err.Error())
	}
	out := m.bytesValue(shared)
	if cb := arg(args, 2); cb.Valid() && !cb.IsUndefined() {
		if !cb.IsFunction() {
			return js.Value{}, typeErr(`The "callback" argument must be of type function. Received %s`, describe(cb))
		}
		m.e.NextTick(cb, m.e.JS.Null(), out)
		return js.Value{}, nil
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Argon2
// ---------------------------------------------------------------------------

// argon2Params carries the validated derivation inputs.
type argon2Params struct {
	message     []byte
	nonce       []byte
	secret      []byte
	assocData   []byte
	parallelism uint32
	tagLength   uint32
	memory      uint32
	passes      uint32
}

func (m *module) argon2(args []js.Value, async bool) (js.Value, *argon2Params, error) {
	algo := arg(args, 0)
	if !algo.IsString() {
		return js.Value{}, nil, typeErr(`The "algorithm" argument must be of type string. Received %s`, describe(algo))
	}
	switch algo.Str() {
	case "argon2d", "argon2i", "argon2id":
	default:
		return js.Value{}, nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			fmt.Sprintf("The argument 'algorithm' must be one of: 'argon2d', 'argon2i', 'argon2id'. Received '%s'", algo.Str()))
	}
	paramsV := arg(args, 1)
	if !paramsV.IsObject() {
		return js.Value{}, nil, typeErr(`The "parameters" argument must be of type object. Received %s`, describe(paramsV))
	}
	p := paramsV.Object()

	message, err := m.bytesArg(p.Get("message"), "parameters.message")
	if err != nil {
		return js.Value{}, nil, err
	}
	nonce, err := m.bytesArg(p.Get("nonce"), "parameters.nonce")
	if err != nil {
		return js.Value{}, nil, err
	}
	if len(nonce) < 8 {
		return js.Value{}, nil, argonOutOfRange("parameters.nonce.byteLength", ">= 8", len(nonce))
	}
	if uint64(len(message)) > math.MaxUint32 {
		return js.Value{}, nil, argonOutOfRange("parameters.message.byteLength", "<= 4294967295", len(message))
	}
	parallelism, err := m.argonUint(p.Get("parallelism"), "parameters.parallelism", 1, 1<<24-1)
	if err != nil {
		return js.Value{}, nil, err
	}
	if parallelism > 255 {
		// Argon2's thread count is a byte in this implementation.
		return js.Value{}, nil, js.ThrowCode("Error", "ERR_CRYPTO_ARGON2_NOT_SUPPORTED",
			"Argon2 parallelism above 255 is not supported by this build")
	}
	tagLength, err := m.argonUint(p.Get("tagLength"), "parameters.tagLength", 4, math.MaxUint32)
	if err != nil {
		return js.Value{}, nil, err
	}
	memory, err := m.argonUint(p.Get("memory"), "parameters.memory", uint64(8*parallelism), math.MaxUint32)
	if err != nil {
		return js.Value{}, nil, err
	}
	passes, err := m.argonUint(p.Get("passes"), "parameters.passes", 1, math.MaxUint32)
	if err != nil {
		return js.Value{}, nil, err
	}
	var secret, ad []byte
	if v := p.Get("secret"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		secret, err = m.bytesArg(v, "parameters.secret")
		if err != nil {
			return js.Value{}, nil, err
		}
	}
	if v := p.Get("associatedData"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		ad, err = m.bytesArg(v, "parameters.associatedData")
		if err != nil {
			return js.Value{}, nil, err
		}
	}

	par := &argon2Params{
		message: message, nonce: nonce, secret: secret, assocData: ad,
		parallelism: uint32(parallelism), tagLength: uint32(tagLength),
		memory: uint32(memory), passes: uint32(passes),
	}
	// golang.org/x/crypto/argon2 can only derive without a secret or
	// associated data. Node's libargon2 mixes both into the initial hash, so
	// silently ignoring them would produce a different key.
	if len(secret) > 0 || len(ad) > 0 {
		return js.Value{}, nil, js.ThrowCode("Error", "ERR_CRYPTO_ARGON2_NOT_SUPPORTED",
			"Argon2 secret and associated data are not supported by this build")
	}
	var out []byte
	switch algo.Str() {
	case "argon2i":
		out = argon2.Key(message, nonce, par.passes, par.memory, uint8(par.parallelism), par.tagLength)
	case "argon2id":
		out = argon2.IDKey(message, nonce, par.passes, par.memory, uint8(par.parallelism), par.tagLength)
	default: // argon2d is not provided by golang.org/x/crypto/argon2.
		return js.Value{}, nil, js.ThrowCode("Error", "ERR_CRYPTO_ARGON2_NOT_SUPPORTED",
			"Argon2d is not supported by this build")
	}
	return m.bytesValue(out), par, nil
}

func (m *module) argonUint(v js.Value, name string, min, max uint64) (uint64, error) {
	if !v.Valid() || !v.IsNumber() {
		return 0, typeErr(`The "%s" argument must be an integer.%s`, name, invalidArgHelper(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) || f < 0 || f > math.MaxUint32 {
		return 0, argonOutOfRange(name, fmt.Sprintf(">= %d and <= %d", min, max), f)
	}
	n := uint64(f)
	if n < min || n > max {
		return 0, argonOutOfRange(name, fmt.Sprintf(">= %d and <= %d", min, max), f)
	}
	return n, nil
}

func argonOutOfRange(name, bound string, got any) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		fmt.Sprintf(`The value of "%s" is out of range. It must be %s. Received %v`, name, bound, got))
}

// ---------------------------------------------------------------------------
// PKCS#12
// ---------------------------------------------------------------------------

func (m *module) parsePKCS12(args []js.Value) (js.Value, error) {
	pfx, err := m.bytesArg(arg(args, 0), "pfx")
	if err != nil {
		return js.Value{}, err
	}
	password := ""
	switch v := arg(args, 1); {
	case v.Valid() && v.IsString():
		password = v.Str()
	case v.Valid() && !v.IsUndefined() && !v.IsNull():
		if b, ok := m.bytesOf(v); ok {
			password = string(b)
		} else {
			return js.Value{}, typeErr(`The "passphrase" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
		}
	}
	key, cert, err := pkcs12.Decode(pfx, password)
	if err != nil {
		return js.Value{}, js.Throw("Error", "Unable to load PKCS12: "+err.Error())
	}
	out := m.e.JS.Obj()
	if st, ok := wrapKey(key); ok {
		out.Set("key", m.newKeyObject(st).V())
	} else {
		out.SetNull("key")
	}
	if cert != nil {
		out.Set("cert", m.newX509Instance(cert.Raw))
		out.Set("ca", m.e.JS.Arr(nil).V())
	} else {
		out.SetNull("cert")
		out.Set("ca", m.e.JS.Arr(nil).V())
	}
	return out.V(), nil
}

// newX509Instance builds an X509Certificate from DER through the JS class so
// the normal brand/slot bookkeeping applies.
func (m *module) newX509Instance(der []byte) js.Value {
	if !m.x509Ctor.Valid() || !m.x509Ctor.IsFunction() {
		return m.e.JS.Undefined()
	}
	fn, err := m.e.JS.Eval("(function (C, b) { return new C(b); })", "[gode:crypto/x509new]")
	if err != nil {
		return m.e.JS.Undefined()
	}
	out, err := fn.Object().Call(js.Object{}, m.x509Ctor.V(), m.bytesValue(der))
	if err != nil {
		return m.e.JS.Undefined()
	}
	return out
}
