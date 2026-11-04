package web

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"godejs/internal/jstest"
	_ "godejs/modules/crypto"
	_ "godejs/modules/net"
	_ "godejs/modules/tls"
)

// TestWebSocketClient drives a real RFC 6455 handshake and frame exchange
// against a hand-written TCP server, so the client is exercised end to end.
func TestWebSocketClient(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		key := ""
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(strings.ToLower(line), "sec-websocket-key:") {
				key = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			}
			if line == "\r\n" {
				break
			}
		}
		sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		accept := base64.StdEncoding.EncodeToString(sum[:])
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\nConnection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n", accept)
		conn.Write([]byte{0x81, 0x02, 'h', 'i'})

		// Read the client's masked close frame, then acknowledge it.
		header := make([]byte, 2)
		if _, err := io.ReadFull(reader, header); err != nil {
			return
		}
		length := int(header[1] & 0x7f)
		if length > 0 {
			if _, err := io.ReadFull(reader, make([]byte, 4+length)); err != nil {
				return
			}
		}
		conn.Write([]byte{0x88, 0x00})
		time.Sleep(50 * time.Millisecond)
	}()

	e := modernSetup(t)
	port := listener.Addr().(*net.TCPAddr).Port
	jstest.Str(t, e, fmt.Sprintf(`(function () {
	  globalThis.__ws = 'pending';
	  var socket = new WebSocket('ws://127.0.0.1:%d/chat');
	  socket.onopen = function () { globalThis.__ws = 'open'; };
	  socket.onmessage = function (event) {
	    globalThis.__ws = 'msg:' + event.data;
	    socket.close();
	  };
	  socket.onclose = function (event) {
	    globalThis.__ws += ':closed:' + event.wasClean;
	  };
	  socket.onerror = function () {
	    if (globalThis.__ws === 'pending') globalThis.__ws = 'error';
	  };
	  return 'scheduled';
	})()`, port))

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__ws`), "msg:hi:closed:true")
}
