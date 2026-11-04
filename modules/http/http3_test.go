package http

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	stdnet "net"
	stdhttp "net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"

	"godejs/internal/js"
	"godejs/internal/jstest"
	"godejs/modules/timers"
)

func testCertificate(t *testing.T) (keyPEM, certPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []stdnet.IP{stdnet.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestHTTP3ServerRequestResponse(t *testing.T) {
	e := jstest.New(t)
	var stderr bytes.Buffer
	e.Opts.Stderr = &stderr
	keyPEM, certPEM := testCertificate(t)
	e.Global().SetStr("__h3Key", keyPEM)
	e.Global().SetStr("__h3Cert", certPEM)

	portCh := make(chan int, 1)
	e.Global().SetFuncHidden("__h3Ready", func(this js.Object, args []js.Value) (js.Value, error) {
		port := 0
		if len(args) > 0 && args[0].IsNumber() {
			port = args[0].Int()
		}
		select {
		case portCh <- port:
		default:
		}
		return e.JS.Undefined(), nil
	})

	jstest.Install(t, e, "timers", timers.Install)
	jstest.Install(t, e, "http3", InstallHTTP3)
	jstest.Eval(t, e, `globalThis.__h3Error = '';
var h3server = http3.createServer({ key: __h3Key, cert: __h3Cert }, function (req, res) {
  res.statusCode = 200;
  res.setHeader('content-type', 'text/plain');
  res.end('h3:' + req.method + ':' + req.url);
  setTimeout(function () { h3server.close(); }, 200);
});
h3server.on('error', function (err) { __h3Error = String(err); __h3Ready(-1); });
h3server.listen(0, '127.0.0.1', function () { __h3Ready(h3server.address().port); });`)

	done := make(chan error, 1)
	go func() { done <- e.RunMain() }()

	var port int
	select {
	case port = <-portCh:
	case <-time.After(5 * time.Second):
		e.Exit(0)
		<-done
		t.Fatal("http3 server did not start")
	}
	if port <= 0 {
		e.Exit(0)
		<-done
		t.Fatalf("http3 server reported port %d", port)
	}

	transport := &http3.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	defer transport.Close()
	client := &stdhttp.Client{Transport: transport}
	resp, err := client.Get(fmt.Sprintf("https://127.0.0.1:%d/h3?x=1", port))
	if err != nil {
		e.Exit(0)
		<-done
		t.Fatalf("http3 GET: %v (js error: %s)", err, jstest.Str(t, e, `__h3Error`))
	}
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if resp.ProtoMajor != 3 {
		t.Fatalf("response protocol = %s, want HTTP/3", resp.Proto)
	}
	if got := string(body); got != "h3:GET:/h3?x=1" {
		t.Fatalf("body = %q (stderr: %s)", got, stderr.String())
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunMain: %v", err)
		}
	case <-time.After(10 * time.Second):
		e.Exit(0)
		<-done
		t.Fatal("http3 server did not close")
	}
	if got := jstest.Str(t, e, `__h3Error`); got != "" {
		t.Fatalf("http3 error: %s", got)
	}
}
