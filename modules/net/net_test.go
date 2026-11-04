package net

import (
	"testing"
	"time"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	buffer.Install(e)
	jstest.Install(t, e, "net", Install)
	return e
}

func TestIsIP(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `String(net.isIP("127.0.0.1"))`), "4")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIP("::1"))`), "6")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIP("256.0.0.1"))`), "0")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIP(""))`), "0")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIP("localhost"))`), "0")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIPv4("127.0.0.1"))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIPv4("::1"))`), "false")
	jstest.Equal(t, jstest.Str(t, e, `String(net.isIPv6("::1"))`), "true")
}

func TestBlockList(t *testing.T) {
	e := setup(t)

	src := `(function () {
		var list = new net.BlockList();
		list.addAddress("127.0.0.1");
		list.addRange("10.0.0.1", "10.0.0.9");
		list.addSubnet("192.168.0.0", 16);
		return [
			list.check("127.0.0.1"),
			list.check("127.0.0.2"),
			list.check("10.0.0.5"),
			list.check("10.0.1.5"),
			list.check("192.168.4.4"),
			list.check("192.169.0.1"),
			list.check("::1"),
			list.rules.length
		].join(",");
	})()`
	jstest.Equal(t, jstest.Str(t, e, src), "true,false,true,false,true,false,false,3")
}

func TestSocketAddress(t *testing.T) {
	e := setup(t)

	src := `(function () {
		var v4 = new net.SocketAddress({ address: "1.2.3.4", port: 80 });
		var v6 = new net.SocketAddress({ address: "::1", family: "ipv6", port: 443 });
		return [
			v4.family, v4.address, v4.port,
			v4.toString(),
			JSON.stringify(v4),
			v6.family, v6.flowlabel
		].join("|");
	})()`
	jstest.Equal(t, jstest.Str(t, e, src), "ipv4|1.2.3.4|80|1.2.3.4:80|{\"address\":\"1.2.3.4\",\"family\":\"ipv4\",\"port\":80}|ipv6|0")
}

// TestEchoRoundTrip drives a real loopback server and client. The pump timer
// is what lets the run loop deliver the asynchronous connect/data/close
// events, so this also covers the keep-alive accounting.
func TestEchoRoundTrip(t *testing.T) {
	e := setup(t)

	jstest.Eval(t, e, `globalThis.__probe = (function () {
		var received = null;
		var closed = false;
		var error = null;
		var server = net.createServer(function (socket) {
			socket.on('data', function (d) { socket.write(d); });
		});
		server.on('error', function (err) { error = String(err); });
		server.listen(0, '127.0.0.1', function () {
			var port = server.address().port;
			var client = net.createConnection(port, '127.0.0.1');
			client.setEncoding('utf8');
			client.on('connect', function () { client.write('hello'); });
			client.on('error', function (err) { error = String(err); });
			client.on('data', function (d) {
				received = d;
				client.end();
			});
			client.on('close', function () {
				server.close(function () { closed = true; });
			});
		});
		return function () {
			return [String(received), closed, String(error)].join('|');
		};
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
		t.Fatal("echo round trip did not finish")
	}

	jstest.Equal(t, jstest.Str(t, e, `__probe()`), "hello|true|null")
}

// TestListenTwice checks Node's ERR_SERVER_ALREADY_LISTEN guard.
func TestListenTwice(t *testing.T) {
	e := setup(t)

	src := `(function () {
		var server = net.createServer();
		server.listen(0, '127.0.0.1');
		var code;
		try {
			server.listen(0, '127.0.0.1');
			code = "no-throw";
		} catch (err) {
			code = err.code;
		}
		server.close();
		return code;
	})()`
	jstest.Equal(t, jstest.Str(t, e, src), "ERR_SERVER_ALREADY_LISTEN")
}
