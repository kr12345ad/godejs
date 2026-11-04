package crypto

import (
	"crypto/dsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/asn1"
	"fmt"
	"math/big"
	"strings"

	"godejs/internal/js"
)

// RSA padding modes, matching OpenSSL's numbering and crypto.constants.
const (
	rsaPaddingPKCS1 = 1
	rsaPaddingNone  = 3
	rsaPaddingOAEP  = 4
	rsaPaddingPSS   = 6
)

// rsaOp is which of the four RSA wrappers is running.
type rsaOp int

const (
	rsaPublicEncrypt rsaOp = iota
	rsaPrivateDecrypt
	rsaPrivateEncrypt
	rsaPublicDecrypt
)

func (m *module) installAsym(exports js.Object) {
	exports.SetFunc("publicEncrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.rsaCipher(args, rsaPublicEncrypt)
	})
	exports.SetFunc("privateDecrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.rsaCipher(args, rsaPrivateDecrypt)
	})
	exports.SetFunc("privateEncrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.rsaCipher(args, rsaPrivateEncrypt)
	})
	exports.SetFunc("publicDecrypt", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.rsaCipher(args, rsaPublicDecrypt)
	})
}

// keyIDOf returns the id registered for one of this module's KeyObjects. The
// id is kept in a WeakMap held by the KeyObject prototype, so an instance has
// no own names or symbols (matching Node's native slot storage).
func keyIDOf(v js.Value) (int, bool) {
	if !v.IsObject() {
		return 0, false
	}
	fn := v.Object().Get("$godeOwnId")
	if !fn.Valid() || !fn.IsFunction() {
		return 0, false
	}
	id, err := fn.Object().Call(v.Object())
	if err != nil || !id.Valid() || !id.IsNumber() {
		return 0, false
	}
	return int(id.Int()), true
}

// isKeyObjectValue reports whether v is one of this module's KeyObjects.
func isKeyObjectValue(v js.Value) bool {
	_, ok := keyIDOf(v)
	return ok
}

// rsaCipher implements publicEncrypt/privateDecrypt/privateEncrypt/publicDecrypt.
// The first argument may be a key (string, Buffer, KeyObject) or an options
// object of the form {key, padding, oaepHash, oaepLabel, passphrase}.
func (m *module) rsaCipher(args []js.Value, op rsaOp) (js.Value, error) {
	keyV := arg(args, 0)
	padding := rsaPaddingOAEP
	oaepHash := "sha1"
	var label []byte
	if keyV.IsObject() {
		o := keyV.Object()
		if !isKeyObjectValue(keyV) && o.Get("key").Valid() {
			if v := o.Get("padding"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
				if !v.IsNumber() {
					return js.Value{}, typeErr(`The "options.padding" argument must be of type number. Received %s`, describe(v))
				}
				padding = int(v.Float())
			}
			if v := o.Get("oaepHash"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
				if !v.IsString() {
					return js.Value{}, typeErr(`The "options.oaepHash" argument must be of type string. Received %s`, describe(v))
				}
				oaepHash = v.Str()
			}
			if v := o.Get("oaepLabel"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
				b, ok := m.bytesOf(v)
				if !ok {
					return js.Value{}, typeErr(`The "options.oaepLabel" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
				}
				label = b
			}
		}
	}
	switch padding {
	case rsaPaddingPKCS1, rsaPaddingNone, rsaPaddingOAEP:
	default:
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			fmt.Sprintf("The argument 'options.padding' is invalid. Received %d", padding))
	}
	dataV := arg(args, 1)
	data, ok := m.bytesOf(dataV)
	if !ok {
		return js.Value{}, typeErr(`The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received %s`, describe(dataV))
	}

	wantPrivate := op == rsaPrivateDecrypt || op == rsaPrivateEncrypt
	st, err := m.loadKey(keyV, wantPrivate)
	if err != nil {
		return js.Value{}, err
	}
	priv, isPriv := st.key.(*rsa.PrivateKey)
	pub, isPub := st.key.(*rsa.PublicKey)
	if isPriv && !wantPrivate {
		pub, isPub = &priv.PublicKey, true
	}
	if !isPriv && !isPub {
		if wantPrivate {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The argument 'key' must be a private key")
		}
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The argument 'key' must be a public key")
	}

	switch op {
	case rsaPublicEncrypt:
		switch padding {
		case rsaPaddingOAEP:
			fn, _, ok := lookup(oaepHash)
			if !ok {
				return js.Value{}, invalidDigest()
			}
			out, err := rsa.EncryptOAEP(fn(), rand.Reader, pub, data, label)
			if err != nil {
				return js.Value{}, rsaFailure(err)
			}
			return m.bytesValue(out), nil
		case rsaPaddingPKCS1:
			out, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
			if err != nil {
				return js.Value{}, rsaFailure(err)
			}
			return m.bytesValue(out), nil
		default:
			size := (pub.N.BitLen() + 7) / 8
			if len(data) != size {
				return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_PADDING", "Invalid padding")
			}
			return m.bytesValue(rsaRawPublic(pub, data)), nil
		}
	case rsaPrivateDecrypt:
		switch padding {
		case rsaPaddingOAEP:
			fn, _, ok := lookup(oaepHash)
			if !ok {
				return js.Value{}, invalidDigest()
			}
			out, err := rsa.DecryptOAEP(fn(), nil, priv, data, label)
			if err != nil {
				return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_RSA_OAEP_DECODING_ERROR",
					"error:1C800064:Provider routines::oaep decoding error")
			}
			return m.bytesValue(out), nil
		case rsaPaddingPKCS1:
			out, err := rsa.DecryptPKCS1v15(rand.Reader, priv, data)
			if err != nil {
				return js.Value{}, js.ThrowCode("Error", "ERR_OSSL_RSA_PKCS_DECODING_ERROR",
					"error:0200009F:rsa routines::pkcs decoding error")
			}
			return m.bytesValue(out), nil
		default:
			size := (priv.N.BitLen() + 7) / 8
			if len(data) != size {
				return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_PADDING", "Invalid padding")
			}
			return m.bytesValue(rsaRawPrivate(priv, data)), nil
		}
	case rsaPrivateEncrypt:
		size := (priv.N.BitLen() + 7) / 8
		if padding == rsaPaddingNone {
			if len(data) != size {
				return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_PADDING", "Invalid padding")
			}
			return m.bytesValue(rsaRawPrivate(priv, data)), nil
		}
		padded, err := pkcs1Pad(data, 0x01, size)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(rsaRawPrivate(priv, padded)), nil
	default: // rsaPublicDecrypt
		size := (pub.N.BitLen() + 7) / 8
		raw := rsaRawPublic(pub, data)
		if len(raw) != size && len(data) != size {
			return js.Value{}, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_PADDING", "Invalid padding")
		}
		out, err := pkcs1Unpad(raw, 0x01)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(out), nil
	}
	return js.Value{}, js.Throw("Error", "unknown RSA operation")
}

func rsaFailure(err error) error {
	return js.ThrowCode("Error", "ERR_CRYPTO_OPERATION_FAILED", err.Error())
}

// rsaRawPublic computes input^e mod n, left-padded to the modulus length.
func rsaRawPublic(pub *rsa.PublicKey, in []byte) []byte {
	size := (pub.N.BitLen() + 7) / 8
	m := new(big.Int).Exp(new(big.Int).SetBytes(in), big.NewInt(int64(pub.E)), pub.N)
	return leftPad(m.Bytes(), size)
}

// rsaRawPrivate computes input^d mod n, left-padded to the modulus length.
func rsaRawPrivate(priv *rsa.PrivateKey, in []byte) []byte {
	size := (priv.N.BitLen() + 7) / 8
	m := new(big.Int).Exp(new(big.Int).SetBytes(in), priv.D, priv.N)
	return leftPad(m.Bytes(), size)
}

func leftPad(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}

// pkcs1Pad builds an EMSA-PKCS1-v1_5 block of the given type (1 for
// signatures, 2 for encryption).
func pkcs1Pad(msg []byte, blockType byte, size int) ([]byte, error) {
	if len(msg) > size-11 {
		return nil, js.ThrowCode("Error", "ERR_CRYPTO_INVALID_PADDING", "Invalid padding")
	}
	out := make([]byte, size)
	out[1] = blockType
	for i := 2; i < size-len(msg)-1; i++ {
		out[i] = 0xFF
	}
	out[size-len(msg)-1] = 0
	copy(out[size-len(msg):], msg)
	return out, nil
}

// pkcs1Unpad validates and strips an EMSA-PKCS1-v1_5 block.
func pkcs1Unpad(msg []byte, blockType byte) ([]byte, error) {
	if len(msg) < 11 || msg[0] != 0 || msg[1] != blockType {
		return nil, js.ThrowCode("Error", "ERR_OSSL_RSA_PKCS_DECODING_ERROR",
			"error:0200009F:rsa routines::pkcs decoding error")
	}
	i := 2
	for i < len(msg) && msg[i] == 0xFF {
		i++
	}
	if i < 10 || i >= len(msg) || msg[i] != 0 {
		return nil, js.ThrowCode("Error", "ERR_OSSL_RSA_PKCS_DECODING_ERROR",
			"error:0200009F:rsa routines::pkcs decoding error")
	}
	return msg[i+1:], nil
}

// ---------------------------------------------------------------------------
// DSA
// ---------------------------------------------------------------------------

var (
	oidDSA = asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 1}
)

// dsaAlgorithm identifies a DSA key and carries its domain parameters.
type dsaAlgorithm struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue
}

type dsaParams struct {
	P *big.Int
	Q *big.Int
	G *big.Int
}

type dsaSPKI struct {
	Algorithm dsaAlgorithm
	PublicKey asn1.BitString
}

type dsaPKCS8 struct {
	Version    int
	Algorithm  dsaAlgorithm
	PrivateKey []byte
}

type dsaTraditional struct {
	Version int
	P       *big.Int
	Q       *big.Int
	G       *big.Int
	Y       *big.Int
	X       *big.Int
}

func dsaParametersDER(p, q, g *big.Int) ([]byte, error) {
	return asn1.Marshal(dsaParams{P: p, Q: q, G: g})
}

// dsaGenerateKey creates a DSA key for one of the four parameter sizes
// OpenSSL (and therefore Node) accepts.
func dsaGenerateKey(modulusBits, divisorBits int) (*dsa.PrivateKey, error) {
	var size dsa.ParameterSizes
	switch {
	case modulusBits == 1024 && divisorBits == 160:
		size = dsa.L1024N160
	case modulusBits == 2048 && divisorBits == 224:
		size = dsa.L2048N224
	case modulusBits == 2048 && divisorBits == 256:
		size = dsa.L2048N256
	case modulusBits == 3072 && divisorBits == 256:
		size = dsa.L3072N256
	default:
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid DSA modulus and divisor length")
	}
	var params dsa.Parameters
	if err := dsa.GenerateParameters(&params, rand.Reader, size); err != nil {
		return nil, js.Throw("Error", err.Error())
	}
	key := &dsa.PrivateKey{}
	key.Parameters = params
	if err := dsa.GenerateKey(key, rand.Reader); err != nil {
		return nil, js.Throw("Error", err.Error())
	}
	return key, nil
}

func dsaMarshalSPKI(pub *dsa.PublicKey) ([]byte, error) {
	params, err := dsaParametersDER(pub.P, pub.Q, pub.G)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(dsaSPKI{
		Algorithm: dsaAlgorithm{Algorithm: oidDSA, Parameters: asn1.RawValue{FullBytes: params}},
		PublicKey: asn1.BitString{Bytes: pub.Y.Bytes(), BitLength: len(pub.Y.Bytes()) * 8},
	})
}

func dsaMarshalPKCS8(priv *dsa.PrivateKey) ([]byte, error) {
	params, err := dsaParametersDER(priv.P, priv.Q, priv.G)
	if err != nil {
		return nil, err
	}
	inner, err := asn1.Marshal(priv.X)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(dsaPKCS8{
		Version:    0,
		Algorithm:  dsaAlgorithm{Algorithm: oidDSA, Parameters: asn1.RawValue{FullBytes: params}},
		PrivateKey: inner,
	})
}

// dsaParseSPKI decodes an x509 SubjectPublicKeyInfo holding a DSA key.
func dsaParseSPKI(der []byte) (*dsa.PublicKey, bool) {
	var spki dsaSPKI
	if _, err := asn1.Unmarshal(der, &spki); err != nil {
		return nil, false
	}
	if !spki.Algorithm.Algorithm.Equal(oidDSA) {
		return nil, false
	}
	var p dsaParams
	if _, err := asn1.Unmarshal(spki.Algorithm.Parameters.FullBytes, &p); err != nil {
		return nil, false
	}
	if p.P == nil || p.Q == nil || p.G == nil {
		return nil, false
	}
	y := new(big.Int).SetBytes(spki.PublicKey.Bytes)
	if spki.PublicKey.BitLength > 0 && spki.PublicKey.BitLength < len(spki.PublicKey.Bytes)*8 {
		y = new(big.Int).Rsh(y, uint(len(spki.PublicKey.Bytes)*8-spki.PublicKey.BitLength))
	}
	return &dsa.PublicKey{Parameters: dsa.Parameters{P: p.P, Q: p.Q, G: p.G}, Y: y}, true
}

// dsaParsePKCS8 decodes a PKCS#8 PrivateKeyInfo holding a DSA key.
func dsaParsePKCS8(der []byte) (*dsa.PrivateKey, bool) {
	var p8 dsaPKCS8
	if _, err := asn1.Unmarshal(der, &p8); err != nil {
		return nil, false
	}
	if !p8.Algorithm.Algorithm.Equal(oidDSA) {
		return nil, false
	}
	var p dsaParams
	if _, err := asn1.Unmarshal(p8.Algorithm.Parameters.FullBytes, &p); err != nil {
		return nil, false
	}
	if p.P == nil || p.Q == nil || p.G == nil {
		return nil, false
	}
	var x *big.Int
	if _, err := asn1.Unmarshal(p8.PrivateKey, &x); err != nil {
		return nil, false
	}
	key := &dsa.PrivateKey{}
	key.Parameters = dsa.Parameters{P: p.P, Q: p.Q, G: p.G}
	key.X = x
	key.Y = new(big.Int).Exp(p.G, x, p.P)
	return key, true
}

// dsaParseTraditional decodes OpenSSL's "DSA PRIVATE KEY" SEQUENCE.
func dsaParseTraditional(der []byte) (*dsa.PrivateKey, bool) {
	var t dsaTraditional
	if _, err := asn1.Unmarshal(der, &t); err != nil {
		return nil, false
	}
	if t.P == nil || t.Q == nil || t.G == nil || t.X == nil {
		return nil, false
	}
	key := &dsa.PrivateKey{}
	key.Parameters = dsa.Parameters{P: t.P, Q: t.Q, G: t.G}
	key.X = t.X
	if t.Y != nil {
		key.Y = t.Y
	} else {
		key.Y = new(big.Int).Exp(t.G, t.X, t.P)
	}
	return key, true
}

type dsaSignature struct {
	R *big.Int
	S *big.Int
}

// dsaHash truncates a digest the way OpenSSL does for DSA: only the leftmost
// q.BitLen() bits take part in the signature.
func dsaHash(q *big.Int, digest []byte) []byte {
	n := (q.BitLen() + 7) / 8
	if len(digest) > n {
		return digest[:n]
	}
	return digest
}

func dsaSignDER(priv *dsa.PrivateKey, digest []byte) ([]byte, error) {
	r, s, err := dsa.Sign(rand.Reader, priv, dsaHash(priv.Q, digest))
	if err != nil {
		return nil, js.Throw("Error", err.Error())
	}
	return asn1.Marshal(dsaSignature{R: r, S: s})
}

func dsaVerifyDER(pub *dsa.PublicKey, digest, sig []byte) bool {
	var s dsaSignature
	if rest, err := asn1.Unmarshal(sig, &s); err != nil || len(rest) != 0 {
		return false
	}
	if s.R == nil || s.S == nil {
		return false
	}
	return dsa.Verify(pub, dsaHash(pub.Q, digest), s.R, s.S)
}

// dsaDetails mirrors the asymmetricKeyDetails Node reports for DSA keys.
func dsaDetails(p, q *big.Int) (int, int) {
	modulus, divisor := 0, 0
	if p != nil {
		modulus = p.BitLen()
	}
	if q != nil {
		divisor = q.BitLen()
	}
	return modulus, divisor
}

// ---------------------------------------------------------------------------
// ECDSA signature encodings
// ---------------------------------------------------------------------------

type ecdsaSignature struct {
	R *big.Int
	S *big.Int
}

// ecdsaSigToP1363 rewrites a DER ECDSA signature as the fixed-width r||s form
// Node calls 'ieee-p1363'.
func ecdsaSigToP1363(sig []byte, curve elliptic.Curve) ([]byte, error) {
	var s ecdsaSignature
	if _, err := asn1.Unmarshal(sig, &s); err != nil {
		return nil, js.Throw("Error", err.Error())
	}
	if s.R == nil || s.S == nil {
		return nil, js.Throw("Error", "Invalid ECDSA signature")
	}
	size := (curve.Params().BitSize + 7) / 8
	out := make([]byte, 2*size)
	s.R.FillBytes(out[:size])
	s.S.FillBytes(out[size:])
	return out, nil
}

// p1363ToECDSASig is the inverse of ecdsaSigToP1363.
func p1363ToECDSASig(raw []byte, curve elliptic.Curve) ([]byte, error) {
	size := (curve.Params().BitSize + 7) / 8
	if len(raw) != 2*size {
		return nil, js.ThrowCode("TypeError", "ERR_CRYPTO_INVALID_SIGNATURE",
			fmt.Sprintf("Invalid signature length: %d", len(raw)))
	}
	r := new(big.Int).SetBytes(raw[:size])
	s := new(big.Int).SetBytes(raw[size:])
	return asn1.Marshal(ecdsaSignature{R: r, S: s})
}

// keyClone returns a copy so per-call options never leak into a cached
// KeyObject.
func (st *keyState) clone() *keyState {
	if st == nil {
		return nil
	}
	cp := *st
	return &cp
}

// applySignOptions reads the per-call signature options Node accepts as part
// of the key argument: padding, saltLength and dsaEncoding.
func (m *module) applySignOptions(st *keyState, v js.Value) {
	if st == nil || !v.IsObject() {
		return
	}
	o := v.Object()
	// Only the wrapped form {key, ...} carries options; a bare KeyObject does
	// not, and must not be mutated.
	if !o.Get("key").Valid() {
		return
	}
	if p := o.Get("padding"); p.Valid() && p.IsNumber() {
		st.padding = int(p.Float())
	}
	if p := o.Get("saltLength"); p.Valid() && p.IsNumber() {
		st.saltLen = int(p.Float())
	}
	if e := o.Get("dsaEncoding"); e.Valid() && e.IsString() {
		st.dsaEncoding = strings.ToLower(e.Str())
	}
}

// loadKeyForOp loads a key and applies any signature options bundled with it.
func (m *module) loadKeyForOp(v js.Value, private bool) (*keyState, error) {
	st, err := m.loadKey(v, private)
	if err != nil {
		return nil, err
	}
	cp := st.clone()
	m.applySignOptions(cp, v)
	return cp, nil
}

// dsaPublicState derives the public half of a DSA private key.
func dsaPublicState(st *keyState) (*keyState, bool) {
	if priv, ok := st.key.(*dsa.PrivateKey); ok {
		return &keyState{kind: "public", typ: "dsa", key: &priv.PublicKey}, true
	}
	return nil, false
}

// wrapDSA turns a parsed DSA key into a keyState, if it is one.
func wrapDSA(k any) (*keyState, bool) {
	switch v := k.(type) {
	case *dsa.PrivateKey:
		return &keyState{kind: "private", typ: "dsa", key: v}, true
	case *dsa.PublicKey:
		return &keyState{kind: "public", typ: "dsa", key: v}, true
	}
	return nil, false
}

// dsaKeyInfo reports a DSA key's modulus and divisor lengths.
func dsaKeyInfo(st *keyState) (int, int, bool) {
	switch k := st.key.(type) {
	case *dsa.PrivateKey:
		modulus, divisor := dsaDetails(k.P, k.Q)
		return modulus, divisor, true
	case *dsa.PublicKey:
		modulus, divisor := dsaDetails(k.P, k.Q)
		return modulus, divisor, true
	}
	return 0, 0, false
}

// dsaExportDER marshals a DSA key as PKCS#8 (private) or SPKI (public).
func dsaExportDER(st *keyState, typ string) ([]byte, string, bool, error) {
	switch k := st.key.(type) {
	case *dsa.PrivateKey:
		if typ != "" && typ != "pkcs8" {
			return nil, "", false, nil
		}
		der, err := dsaMarshalPKCS8(k)
		return der, "PRIVATE KEY", true, err
	case *dsa.PublicKey:
		if typ != "" && typ != "spki" {
			return nil, "", false, nil
		}
		der, err := dsaMarshalSPKI(k)
		return der, "PUBLIC KEY", true, err
	}
	return nil, "", false, nil
}

// dsaSignState signs digest when st holds a DSA private key.
func dsaSignState(st *keyState, digest []byte) ([]byte, bool, error) {
	if priv, ok := st.key.(*dsa.PrivateKey); ok {
		sig, err := dsaSignDER(priv, digest)
		return sig, true, err
	}
	return nil, false, nil
}

// dsaVerifyState verifies a DER DSA signature when st holds a DSA public key.
func dsaVerifyState(st *keyState, digest, sig []byte) (bool, bool) {
	if pub, ok := st.key.(*dsa.PublicKey); ok {
		return dsaVerifyDER(pub, digest, sig), true
	}
	return false, false
}
