package http

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"godejs/internal/jstest"
	netmod "godejs/modules/net"
)

func TestServerRequestResponse(t *testing.T) {
	e := jstest.New(t)
	var stderr bytes.Buffer
	e.Opts.Stderr = &stderr
	jstest.Install(t, e, "net", netmod.Install)
	jstest.Install(t, e, "http", Install)
	jstest.Eval(t, e, `globalThis.__httpProbe = (function () {
		var result = null;
		var error = null;
		var server = http.createServer(function (req, res) {
			var body = '';
			req.setEncoding('utf8');
			req.on('data', function (chunk) { body += chunk; });
			req.on('end', function () {
				res.statusCode = 202;
				res.setHeader('X-Test', 'ok');
				res.end('got:' + req.method + ':' + body);
			});
		});
		server.on('error', function (err) { error = String(err); });
		server.listen(0, '127.0.0.1', function () {
			var socket = net.createConnection(server.address().port, '127.0.0.1');
			var data = '';
			socket.setEncoding('utf8');
			socket.on('connect', function () {
				socket.write('POST /submit?x=1 HTTP/1.1\r\n' +
					'Host: test\r\nContent-Length: 3\r\nConnection: close\r\n\r\nabc');
			});
			socket.on('data', function (chunk) { data += chunk; });
			socket.on('end', function () {
				result = data;
				server.close();
			});
			socket.on('error', function (err) { error = String(err); });
		});
		return function () { return result || error || 'pending'; };
	})()`)

	done := make(chan error, 1)
	go func() { done <- e.RunMain() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunMain: %v", err)
		}
	case <-time.After(10 * time.Second):
		e.Exit(0)
		<-done
		t.Fatal("http round trip did not finish")
	}

	got := jstest.Str(t, e, `__httpProbe()`)
	t.Logf("stderr: %s", stderr.String())
	if !strings.Contains(got, "HTTP/1.1 202") {
		t.Fatalf("missing status: %q (stderr: %s)", got, stderr.String())
	}
	if !strings.Contains(strings.ToLower(got), "x-test: ok") {
		t.Fatalf("missing response header: %q", got)
	}
	if !strings.Contains(got, "got:POST:abc") {
		t.Fatalf("missing echoed body: %q", got)
	}
}

func TestClientRequestAndGet(t *testing.T) {
	e := jstest.New(t)
	jstest.Install(t, e, "http", Install)
	jstest.Eval(t, e, `globalThis.__clientProbe = (function () {
		var result = null;
		var error = null;
		var server = http.createServer(function (req, res) {
			var body = '';
			req.setEncoding('utf8');
			req.on('data', function (chunk) { body += chunk; });
			req.on('end', function () {
				res.setHeader('X-Echo', req.method);
				res.end(req.method + ':' + req.url + ':' + body);
			});
		});
		server.on('error', function (err) { error = String(err); });
		server.listen(0, '127.0.0.1', function () {
			var port = server.address().port;
			http.get({ host: '127.0.0.1', port: port, path: '/g' }, function (res) {
				var data = '';
				res.setEncoding('utf8');
				res.on('data', function (chunk) { data += chunk; });
				res.on('end', function () {
					var getResult = res.statusCode + ':' + data;
					var req = http.request({ host: '127.0.0.1', port: port, method: 'POST', path: '/p' }, function (postRes) {
						var post = '';
						postRes.setEncoding('utf8');
						postRes.on('data', function (chunk) { post += chunk; });
						postRes.on('end', function () {
							result = getResult + '|' + postRes.statusCode + ':' + post;
							server.close();
						});
					});
					req.on('error', function (err) { error = String(err); });
					req.end('body');
				});
			}).on('error', function (err) { error = String(err); });
		});
		return function () { return result || error || 'pending'; };
	})()`)

	done := make(chan error, 1)
	go func() { done <- e.RunMain() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunMain: %v", err)
		}
	case <-time.After(10 * time.Second):
		e.Exit(0)
		<-done
		t.Fatal("http client round trip did not finish")
	}

	jstest.Equal(t, jstest.Str(t, e, `__clientProbe()`), "200:GET:/g:|200:POST:/p:body")
}
