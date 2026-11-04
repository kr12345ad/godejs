package crypto

import (
	stdcrypto "crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"time"

	"godejs/internal/js"
)

// x509State is the parsed certificate behind one X509Certificate instance.
type x509State struct {
	cert *x509.Certificate
}

// x509Prelude builds the class in JS so the prototype carries real accessors
// (brand checks inspect their descriptors) while the parsing stays in Go.
const x509Prelude = `(function(host) {
  const slot = Symbol('gode.x509');
  function id(self) {
    const v = self == null ? undefined : self[slot];
    if (typeof v !== 'number') host.invalid();
    return v;
  }
  class X509Certificate {
    constructor(data) {
      Object.defineProperty(this, slot, { value: host.init(data) });
    }
    get subject() { return host.field(id(this), 'subject'); }
    get subjectAltName() { return host.field(id(this), 'subjectAltName'); }
    get issuer() { return host.field(id(this), 'issuer'); }
    get issuerCertificate() { return host.field(id(this), 'issuerCertificate'); }
    get infoAccess() { return host.field(id(this), 'infoAccess'); }
    get validFrom() { return host.field(id(this), 'validFrom'); }
    get validTo() { return host.field(id(this), 'validTo'); }
    get validFromDate() { return host.field(id(this), 'validFromDate'); }
    get validToDate() { return host.field(id(this), 'validToDate'); }
    get fingerprint() { return host.field(id(this), 'fingerprint'); }
    get fingerprint256() { return host.field(id(this), 'fingerprint256'); }
    get fingerprint512() { return host.field(id(this), 'fingerprint512'); }
    get keyUsage() { return host.field(id(this), 'keyUsage'); }
    get serialNumber() { return host.field(id(this), 'serialNumber'); }
    get signatureAlgorithm() { return host.field(id(this), 'signatureAlgorithm'); }
    get signatureAlgorithmOid() { return host.field(id(this), 'signatureAlgorithmOid'); }
    get raw() { return host.field(id(this), 'raw'); }
    get publicKey() { return host.field(id(this), 'publicKey'); }
    get ca() { return host.field(id(this), 'ca'); }
    toString() { return host.method(id(this), 'toString', undefined); }
    toJSON() { return host.method(id(this), 'toJSON', undefined); }
    toLegacyObject() { return host.method(id(this), 'toLegacyObject', undefined); }
    checkHost(name) { return host.method(id(this), 'checkHost', name); }
    checkEmail(email) { return host.method(id(this), 'checkEmail', email); }
    checkIP(ip) { return host.method(id(this), 'checkIP', ip); }
    checkIssued(other) { return host.method(id(this), 'checkIssued', id(other)); }
    checkPrivateKey(key) { return host.method(id(this), 'checkPrivateKey', key); }
    verify(key) { return host.method(id(this), 'verify', key); }
  }
  return X509Certificate;
})`

func (m *module) installX509(exports js.Object) {
	m.x509Proto = m.e.JS.Obj()
	host := m.e.JS.Obj()
	host.SetFunc("init", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.parseX509(arg(args, 0))
		if err != nil {
			return js.Value{}, err
		}
		m.mu.Lock()
		m.seq++
		id := m.seq
		m.x509s[id] = st
		m.mu.Unlock()
		return m.e.JS.Num(float64(id)), nil
	})
	host.SetFunc("invalid", func(this js.Object, args []js.Value) (js.Value, error) {
		return js.Value{}, invalidThis("X509Certificate")
	})
	host.SetFunc("field", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.x509Of(arg(args, 0))
		if err != nil {
			return js.Value{}, err
		}
		return m.x509Field(st, arg(args, 1).Str())
	})
	host.SetFunc("method", func(this js.Object, args []js.Value) (js.Value, error) {
		st, err := m.x509Of(arg(args, 0))
		if err != nil {
			return js.Value{}, err
		}
		return m.x509Method(st, arg(args, 1).Str(), arg(args, 2))
	})
	fn, err := m.e.JS.Eval(x509Prelude, "[gode:crypto/x509]")
	if err != nil {
		return
	}
	ctorValue, err := fn.Object().Call(js.Object{}, host.V())
	if err != nil {
		return
	}
	ctor := ctorValue.Object()
	// The prelude's class owns its prototype (that is where the accessors
	// live), so adopt it instead of replacing it.
	m.x509Proto = ctor.Get("prototype").Object()
	m.x509Ctor = ctor
	exports.Set("X509Certificate", ctor.V())
}

func (m *module) x509Of(v js.Value) (*x509State, error) {
	if !v.Valid() || !v.IsNumber() {
		return nil, invalidThis("X509Certificate")
	}
	m.mu.Lock()
	st := m.x509s[int(v.Float())]
	m.mu.Unlock()
	if st == nil {
		return nil, invalidThis("X509Certificate")
	}
	return st, nil
}

func (m *module) parseX509(v js.Value) (*x509State, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil, typeErr(`The "buffer" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received %s`, describe(v))
	}
	data, err := m.bytesArg(v, "buffer")
	if err != nil {
		return nil, err
	}
	der := data
	if block, _ := pem.Decode(data); block != nil && block.Type == "CERTIFICATE" {
		der = block.Bytes
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, js.Throw("Error", "Unable to load certificate: "+err.Error())
	}
	st := &x509State{cert: cert}
	return st, nil
}

func (m *module) x509Field(st *x509State, name string) (js.Value, error) {
	c := st.cert
	switch name {
	case "subject":
		return m.e.JS.Str(c.Subject.String()), nil
	case "issuer":
		return m.e.JS.Str(c.Issuer.String()), nil
	case "subjectAltName":
		return m.e.JS.Str(x509AltNames(c)), nil
	case "issuerCertificate":
		return m.e.JS.Undefined(), nil
	case "infoAccess":
		return m.e.JS.Str(x509InfoAccess(c)), nil
	case "validFrom":
		return m.e.JS.Str(c.NotBefore.UTC().Format("Jan 2 15:04:05 2006 MST")), nil
	case "validTo":
		return m.e.JS.Str(c.NotAfter.UTC().Format("Jan 2 15:04:05 2006 MST")), nil
	case "validFromDate":
		return m.dateValue(c.NotBefore), nil
	case "validToDate":
		return m.dateValue(c.NotAfter), nil
	case "fingerprint":
		return m.e.JS.Str(x509Fingerprint(c.Raw, 1)), nil
	case "fingerprint256":
		return m.e.JS.Str(x509Fingerprint(c.Raw, 256)), nil
	case "fingerprint512":
		return m.e.JS.Str(x509Fingerprint(c.Raw, 512)), nil
	case "keyUsage":
		names := x509KeyUsage(c)
		vals := make([]js.Value, len(names))
		for i, n := range names {
			vals[i] = m.e.JS.Str(n)
		}
		return m.e.JS.Arr(vals).V(), nil
	case "serialNumber":
		return m.e.JS.Str(strings.ToUpper(c.SerialNumber.Text(16))), nil
	case "signatureAlgorithm":
		return m.e.JS.Str(x509SignatureName(c)), nil
	case "signatureAlgorithmOid":
		return m.e.JS.Str(x509SignatureOID(c)), nil
	case "raw":
		return m.bytesValue(c.Raw), nil
	case "publicKey":
		st, ok := wrapKey(c.PublicKey)
		if !ok {
			return m.e.JS.Undefined(), nil
		}
		return m.newKeyObject(st).V(), nil
	case "ca":
		return m.e.JS.Bool(c.IsCA), nil
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) dateValue(t time.Time) js.Value {
	fn, err := m.e.JS.Eval("(function(ms){return new Date(ms);})", "[gode:crypto/date]")
	if err != nil {
		return m.e.JS.Undefined()
	}
	out, err := fn.Object().Call(js.Object{}, m.e.JS.Num(float64(t.UnixMilli())))
	if err != nil {
		return m.e.JS.Undefined()
	}
	return out
}

func (m *module) x509Method(st *x509State, name string, arg js.Value) (js.Value, error) {
	c := st.cert
	switch name {
	case "toString", "toJSON":
		out := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
		return m.e.JS.Str(string(out)), nil
	case "toLegacyObject":
		o := m.e.JS.Obj()
		o.SetStr("subject", c.Subject.String())
		o.SetStr("issuer", c.Issuer.String())
		o.SetStr("valid_from", c.NotBefore.UTC().Format("Jan 2 15:04:05 2006 MST"))
		o.SetStr("valid_to", c.NotAfter.UTC().Format("Jan 2 15:04:05 2006 MST"))
		o.SetStr("serialNumber", strings.ToUpper(c.SerialNumber.Text(16)))
		o.SetStr("fingerprint", x509Fingerprint(c.Raw, 1))
		o.Set("raw", m.bytesValue(c.Raw))
		return o.V(), nil
	case "checkHost":
		if !arg.IsString() {
			return js.Value{}, typeErr(`The "name" argument must be of type string. Received %s`, describe(arg))
		}
		return m.e.JS.Bool(c.VerifyHostname(arg.Str()) == nil), nil
	case "checkEmail":
		if !arg.IsString() {
			return js.Value{}, typeErr(`The "email" argument must be of type string. Received %s`, describe(arg))
		}
		for _, e := range c.EmailAddresses {
			if strings.EqualFold(e, arg.Str()) {
				return m.e.JS.Bool(true), nil
			}
		}
		return m.e.JS.Bool(false), nil
	case "checkIP":
		if !arg.IsString() {
			return js.Value{}, typeErr(`The "ip" argument must be of type string. Received %s`, describe(arg))
		}
		ip := net.ParseIP(arg.Str())
		if ip == nil {
			return m.e.JS.Bool(false), nil
		}
		for _, want := range c.IPAddresses {
			if want.Equal(ip) {
				return m.e.JS.Bool(true), nil
			}
		}
		return m.e.JS.Bool(false), nil
	case "checkIssued":
		other, err := m.x509Arg(arg, "X509Certificate")
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(string(c.RawIssuer) == string(other.cert.RawSubject)), nil
	case "checkPrivateKey":
		st, err := m.keyObjectArg(arg, "key")
		if err != nil {
			return js.Value{}, err
		}
		pub, ok := publicKeyOf(st)
		if !ok {
			return m.e.JS.Bool(false), nil
		}
		a, err1 := x509.MarshalPKIXPublicKey(pub)
		b, err2 := x509.MarshalPKIXPublicKey(c.PublicKey)
		return m.e.JS.Bool(err1 == nil && err2 == nil && string(a) == string(b)), nil
	case "verify":
		st, err := m.keyObjectArg(arg, "key")
		if err != nil {
			return js.Value{}, err
		}
		pub, ok := publicKeyOf(st)
		if !ok {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "key" argument must be a public or private key`)
		}
		return m.e.JS.Bool(x509VerifySignature(c, pub)), nil
	}
	return m.e.JS.Undefined(), nil
}

// x509Arg resolves another certificate instance passed through the JS class.
func (m *module) x509Arg(v js.Value, what string) (*x509State, error) {
	// The JS class resolves the peer instance to its id after its own brand
	// check, so only a number reaches this point.
	if !v.IsNumber() {
		return nil, typeErr(`The "cert" argument must be an instance of %s. Received %s`, what, describe(v))
	}
	return m.x509Of(v)
}

func (m *module) keyObjectArg(v js.Value, name string) (*keyState, error) {
	if _, ok := keyIDOf(v); ok {
		return m.keyStateOf(v.Object())
	}
	return nil, typeErr(`The "%s" argument must be an instance of KeyObject. Received %s`, name, describe(v))
}

func publicKeyOf(st *keyState) (any, bool) {
	switch k := st.key.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, true
	case *rsa.PublicKey:
		return k, true
	case *ecdsa.PrivateKey:
		return &k.PublicKey, true
	case *ecdsa.PublicKey:
		return k, true
	case ed25519.PrivateKey:
		return k.Public(), true
	case ed25519.PublicKey:
		return k, true
	}
	return nil, false
}

func x509VerifySignature(c *x509.Certificate, pub any) bool {
	sig := c.Signature
	signed := c.RawTBSCertificate
	switch key := pub.(type) {
	case *rsa.PublicKey:
		pss := false
		switch c.SignatureAlgorithm {
		case x509.SHA1WithRSA:
			h := sha1.Sum(signed)
			return rsa.VerifyPKCS1v15(key, stdcrypto.SHA1, h[:], sig) == nil
		case x509.SHA256WithRSA:
			h := sha256.Sum256(signed)
			return rsa.VerifyPKCS1v15(key, stdcrypto.SHA256, h[:], sig) == nil
		case x509.SHA384WithRSA:
			h := sha512.Sum384(signed)
			return rsa.VerifyPKCS1v15(key, stdcrypto.SHA384, h[:], sig) == nil
		case x509.SHA512WithRSA:
			h := sha512.Sum512(signed)
			return rsa.VerifyPKCS1v15(key, stdcrypto.SHA512, h[:], sig) == nil
		case x509.SHA256WithRSAPSS:
			pss = true
			h := sha256.Sum256(signed)
			return rsa.VerifyPSS(key, stdcrypto.SHA256, h[:], sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}) == nil
		case x509.SHA384WithRSAPSS:
			pss = true
			h := sha512.Sum384(signed)
			return rsa.VerifyPSS(key, stdcrypto.SHA384, h[:], sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}) == nil
		case x509.SHA512WithRSAPSS:
			pss = true
			h := sha512.Sum512(signed)
			return rsa.VerifyPSS(key, stdcrypto.SHA512, h[:], sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}) == nil
		}
		_ = pss
	case *ecdsa.PublicKey:
		switch c.SignatureAlgorithm {
		case x509.ECDSAWithSHA256:
			h := sha256.Sum256(signed)
			return ecdsa.VerifyASN1(key, h[:], sig)
		case x509.ECDSAWithSHA384:
			h := sha512.Sum384(signed)
			return ecdsa.VerifyASN1(key, h[:], sig)
		case x509.ECDSAWithSHA512:
			h := sha512.Sum512(signed)
			return ecdsa.VerifyASN1(key, h[:], sig)
		}
	case ed25519.PublicKey:
		return ed25519.Verify(key, signed, sig)
	}
	return false
}

func x509Fingerprint(der []byte, bits int) string {
	var sum []byte
	switch bits {
	case 1:
		s := sha1.Sum(der)
		sum = s[:]
	case 256:
		s := sha256.Sum256(der)
		sum = s[:]
	default:
		s := sha512.Sum512(der)
		sum = s[:]
	}
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = strings.ToUpper(hex.EncodeToString([]byte{b}))
	}
	return strings.Join(parts, ":")
}

func x509AltNames(c *x509.Certificate) string {
	var parts []string
	for _, d := range c.DNSNames {
		parts = append(parts, "DNS:"+d)
	}
	for _, e := range c.EmailAddresses {
		parts = append(parts, "email:"+e)
	}
	for _, ip := range c.IPAddresses {
		parts = append(parts, "IP Address:"+ip.String())
	}
	for _, u := range c.URIs {
		parts = append(parts, "URI:"+u.String())
	}
	return strings.Join(parts, ", ")
}

func x509InfoAccess(c *x509.Certificate) string {
	var parts []string
	for _, u := range c.OCSPServer {
		parts = append(parts, "OCSP - URI:"+u)
	}
	for _, u := range c.IssuingCertificateURL {
		parts = append(parts, "CA Issuers - URI:"+u)
	}
	return strings.Join(parts, "\n")
}

func x509KeyUsage(c *x509.Certificate) []string {
	table := []struct {
		bit  x509.KeyUsage
		name string
	}{
		{x509.KeyUsageDigitalSignature, "digitalSignature"},
		{x509.KeyUsageContentCommitment, "nonRepudiation"},
		{x509.KeyUsageKeyEncipherment, "keyEncipherment"},
		{x509.KeyUsageDataEncipherment, "dataEncipherment"},
		{x509.KeyUsageKeyAgreement, "keyAgreement"},
		{x509.KeyUsageCertSign, "keyCertSign"},
		{x509.KeyUsageCRLSign, "cRLSign"},
		{x509.KeyUsageEncipherOnly, "encipherOnly"},
		{x509.KeyUsageDecipherOnly, "decipherOnly"},
	}
	var out []string
	for _, t := range table {
		if c.KeyUsage&t.bit != 0 {
			out = append(out, t.name)
		}
	}
	return out
}

func x509SignatureName(c *x509.Certificate) string {
	switch c.SignatureAlgorithm {
	case x509.MD5WithRSA:
		return "md5WithRSAEncryption"
	case x509.SHA1WithRSA:
		return "sha1WithRSAEncryption"
	case x509.SHA256WithRSA:
		return "sha256WithRSAEncryption"
	case x509.SHA384WithRSA:
		return "sha384WithRSAEncryption"
	case x509.SHA512WithRSA:
		return "sha512WithRSAEncryption"
	case x509.ECDSAWithSHA256:
		return "ecdsa-with-SHA256"
	case x509.ECDSAWithSHA384:
		return "ecdsa-with-SHA384"
	case x509.ECDSAWithSHA512:
		return "ecdsa-with-SHA512"
	case x509.PureEd25519:
		return "ED25519"
	case x509.SHA256WithRSAPSS:
		return "rsassaPss"
	}
	return fmt.Sprintf("unknown (%d)", int(c.SignatureAlgorithm))
}

func x509SignatureOID(c *x509.Certificate) string {
	switch c.SignatureAlgorithm {
	case x509.MD5WithRSA:
		return "1.2.840.113549.1.1.4"
	case x509.SHA1WithRSA:
		return "1.2.840.113549.1.1.5"
	case x509.SHA256WithRSA:
		return "1.2.840.113549.1.1.11"
	case x509.SHA384WithRSA:
		return "1.2.840.113549.1.1.12"
	case x509.SHA512WithRSA:
		return "1.2.840.113549.1.1.13"
	case x509.ECDSAWithSHA256:
		return "1.2.840.10045.4.3.2"
	case x509.ECDSAWithSHA384:
		return "1.2.840.10045.4.3.3"
	case x509.ECDSAWithSHA512:
		return "1.2.840.10045.4.3.4"
	case x509.PureEd25519:
		return "1.3.101.112"
	case x509.SHA256WithRSAPSS:
		return "1.2.840.113549.1.1.10"
	}
	return "0.0"
}
