// Package tls implements Node's TLS builtin with Go's crypto/tls. The
// JavaScript objects are thin event emitters; all handshakes and reads run on
// Go goroutines and cross to the JS thread through the same pump pattern used
// by net, http and dgram.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("tls", Install) }

const pumpEveryMS = 1

type module struct {
	e *env.Env

	mu     sync.Mutex
	queued []func()
	pumpID int
	active int

	servers map[float64]*server
	sockets map[float64]*socket
	seq     float64

	exports    js.Object
	host       js.Object
	makeSocket js.Value
	bufferCtor js.Object
	bufferFrom js.Object
}

type server struct {
	id  float64
	obj js.Object

	ln        net.Listener
	cfg       *tls.Config
	listening bool
	closed    bool
	closeSent bool
	closeCb   js.Value
	handle    bool

	address string
	port    int
	family  string

	conns map[float64]*socket
}

type socket struct {
	id  float64
	obj js.Object

	conn       *tls.Conn
	server     *server
	serverSide bool
	connecting bool
	secure     bool
	closed     bool
	readEnded  bool
	wroteEnd   bool
	handle     bool
	paused     bool
	encoding   string
	pending    [][]byte
	queue      [][]byte
	gen        int

	peer       js.Object
	authorized bool
	authErr    string
	protocol   string
	alpn       string
	cipher     string
	servername string

	remoteAddress string
	remotePort    int
	localAddress  string
	localPort     int
	bytesRead     int
	bytesWritten  int
}

func Install(e *env.Env) js.Object {
	m := &module{
		e:       e,
		servers: map[float64]*server{},
		sockets: map[float64]*socket{},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()
	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("initSock", m.hostInitSock)
	m.host.SetFuncHidden("initServer", m.hostInitServer)
	m.host.SetFuncHidden("listen", m.hostListen)
	m.host.SetFuncHidden("serverClose", m.hostServerClose)
	m.host.SetFuncHidden("serverAddress", m.hostServerAddress)
	m.host.SetFuncHidden("getConnections", m.hostGetConnections)
	m.host.SetFuncHidden("connect", m.hostConnect)
	m.host.SetFuncHidden("write", m.hostWrite)
	m.host.SetFuncHidden("end", m.hostEnd)
	m.host.SetFuncHidden("destroy", m.hostDestroy)
	m.host.SetFuncHidden("address", m.hostAddress)
	m.host.SetFuncHidden("peerCertificate", m.hostPeerCertificate)
	m.host.SetFuncHidden("getProtocol", m.hostGetProtocol)
	m.host.SetFuncHidden("getCipher", m.hostGetCipher)
	m.host.SetFuncHidden("isSessionReused", m.hostIsSessionReused)
	m.host.SetFuncHidden("getSession", m.hostGetSession)
	m.host.SetFuncHidden("setServername", m.hostSetServername)

	buffer := e.Exports("buffer").Get("Buffer")
	if buffer.Valid() && buffer.IsFunction() {
		m.bufferCtor = buffer.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}
	events := e.Exports("events")
	ee := events.Get("EventEmitter")
	if !ee.Valid() || !ee.IsFunction() {
		panic("godejs tls: events.EventEmitter is unavailable")
	}
	api, err := e.JS.Eval(prelude, "tls.js")
	if err != nil {
		panic("godejs tls: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V(), ee, buffer)
	if err != nil {
		panic("godejs tls: " + err.Error())
	}
	obj := built.Object()
	m.makeSocket = obj.Get("makeSocket").Keep()
	for _, name := range obj.OwnKeys() {
		if name == "makeSocket" {
			continue
		}
		m.exports.Set(name, obj.Get(name))
	}
	m.exports.Set("DEFAULT_MIN_VERSION", e.JS.Str("TLSv1.2"))
	m.exports.Set("DEFAULT_MAX_VERSION", e.JS.Str("TLSv1.3"))
	m.exports.Set("DEFAULT_ECDH_CURVE", e.JS.Str("auto"))
	m.exports.Set("CLIENT_RENEG_LIMIT", e.JS.Num(3))
	m.exports.Set("CLIENT_RENEG_WINDOW", e.JS.Num(600))
	m.exports.SetFunc("getCiphers", func(this js.Object, args []js.Value) (js.Value, error) {
		names := make([]string, 0)
		for _, c := range tls.CipherSuites() {
			names = append(names, strings.ToLower(c.Name))
		}
		for _, c := range tls.InsecureCipherSuites() {
			names = append(names, strings.ToLower(c.Name))
		}
		arr := e.JS.Arr(nil)
		for i, name := range names {
			arr.SetAt(i, e.JS.Str(name))
		}
		return arr.V(), nil
	})
	roots := e.JS.Arr(nil)
	for i, cert := range rootCertificates() {
		roots.SetAt(i, e.JS.Str(cert))
	}
	m.exports.Set("rootCertificates", roots.V())
	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

func Implemented() []string {
	return []string{
		"tls.Server",
		"tls.Server#address",
		"tls.Server#close",
		"tls.Server#getConnections",
		"tls.Server#listen",
		"tls.TLSSocket",
		"tls.TLSSocket#address",
		"tls.TLSSocket#destroy",
		"tls.TLSSocket#end",
		"tls.TLSSocket#getCipher",
		"tls.TLSSocket#getPeerCertificate",
		"tls.TLSSocket#getProtocol",
		"tls.TLSSocket#getSession",
		"tls.TLSSocket#isSessionReused",
		"tls.TLSSocket#setServername",
		"tls.TLSSocket#write",
		"tls.checkServerIdentity",
		"tls.connect",
		"tls.createSecureContext",
		"tls.createServer",
		"tls.getCiphers",
		"tls.rootCertificates",
	}
}

func (m *module) nextID() float64 {
	m.mu.Lock()
	m.seq++
	id := m.seq
	m.mu.Unlock()
	return id
}

func (m *module) push(fn func()) {
	m.mu.Lock()
	m.queued = append(m.queued, fn)
	m.mu.Unlock()
}

func (m *module) pump(this js.Object, args []js.Value) (js.Value, error) {
	for {
		m.mu.Lock()
		if len(m.queued) == 0 {
			m.mu.Unlock()
			break
		}
		fn := m.queued[0]
		m.queued = m.queued[1:]
		m.mu.Unlock()
		fn()
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) addHandle() {
	m.mu.Lock()
	m.active++
	m.mu.Unlock()
	m.startPump()
}

func (m *module) dropHandle() {
	m.mu.Lock()
	m.active--
	idle := m.active <= 0
	id := 0
	if idle {
		id, m.pumpID = m.pumpID, 0
	}
	m.mu.Unlock()
	if idle && id != 0 {
		m.e.ClearTimer(id)
	}
}

func (m *module) startPump() {
	m.mu.Lock()
	if m.pumpID != 0 || m.active <= 0 {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()
	id := m.e.SetTimeout(m.host.Get("pump"), pumpEveryMS, true, nil)
	m.mu.Lock()
	m.pumpID = id
	m.mu.Unlock()
}

func (m *module) shutdown() {
	m.mu.Lock()
	servers := make([]*server, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	id := m.pumpID
	m.pumpID = 0
	m.active = 0
	m.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
	for _, s := range servers {
		s.closed = true
		if s.ln != nil {
			_ = s.ln.Close()
		}
	}
}

func (m *module) sockOf(this js.Object) *socket {
	v := this.Get("$godeTLSId")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	s := m.sockets[v.Float()]
	m.mu.Unlock()
	return s
}

func (m *module) srvOf(this js.Object) *server {
	v := this.Get("$godeTLSId")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	s := m.servers[v.Float()]
	m.mu.Unlock()
	return s
}

func (m *module) hostInitSock(this js.Object, args []js.Value) (js.Value, error) {
	id := m.nextID()
	s := &socket{id: id, obj: this, connecting: false}
	if v := arg(args, 0); v.IsBool() {
		s.serverSide = v.Bool()
	}
	m.mu.Lock()
	m.sockets[id] = s
	m.mu.Unlock()
	this.SetHidden("$godeTLSId", m.e.JS.Num(id))
	this.Protect()
	this.SetBool("encrypted", true)
	this.SetBool("authorized", false)
	this.SetBool("connecting", false)
	this.SetBool("secureConnecting", false)
	this.SetBool("destroyed", false)
	this.SetBool("readable", true)
	this.SetBool("writable", true)
	this.SetBool("pending", false)
	this.SetStr("readyState", "open")
	this.Set("alpnProtocol", m.e.JS.Bool(false))
	return this.V(), nil
}

func (m *module) hostInitServer(this js.Object, args []js.Value) (js.Value, error) {
	options := arg(args, 0)
	cfg, err := serverConfig(options)
	if err != nil {
		return js.Value{}, err
	}
	id := m.nextID()
	s := &server{id: id, obj: this, cfg: cfg, conns: map[float64]*socket{}}
	m.mu.Lock()
	m.servers[id] = s
	m.mu.Unlock()
	this.SetHidden("$godeTLSId", m.e.JS.Num(id))
	this.Protect()
	this.SetBool("listening", false)
	return this.V(), nil
}

func (m *module) hostListen(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.Server")
	}
	if s.listening {
		return js.Value{}, js.ThrowCode("Error", "ERR_SERVER_ALREADY_LISTEN", "Listen method has been called more than once without closing.")
	}
	values := args
	if len(args) == 1 && args[0].IsArray() {
		arr := args[0].Object()
		values = make([]js.Value, arr.Length())
		for i := range values {
			values[i] = arr.At(i)
		}
	}
	port, host, cb := 0, "", js.Value{}
	for _, v := range values {
		if v.IsFunction() {
			cb = v
		} else if v.IsNumber() {
			port = int(v.Float())
		} else if v.IsString() && host == "" {
			host = v.Str()
		} else if v.IsObject() {
			o := v.Object()
			if p := o.Get("port"); p.IsNumber() {
				port = int(p.Float())
			}
			if h := o.Get("host"); h.IsString() {
				host = h.Str()
			}
		}
	}
	if host == "" {
		host = "0.0.0.0"
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "EADDRINUSE", err.Error())
	}
	s.ln = ln
	s.listening = true
	s.closed = false
	s.closeSent = false
	s.address, s.port = splitHostPort(ln.Addr().String())
	s.family = "IPv4"
	if ip := net.ParseIP(s.address); ip != nil && ip.To4() == nil {
		s.family = "IPv6"
	}
	s.handle = true
	this.SetBool("listening", true)
	m.addHandle()
	go m.acceptLoop(s)
	m.emitAsync(s.obj, "listening")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) acceptLoop(s *server) {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			m.push(func() { m.finishServerClose(s) })
			return
		}
		go m.handleServerConn(s, conn)
	}
}

func (m *module) handleServerConn(s *server, raw net.Conn) {
	conn := tls.Server(raw, s.cfg)
	if err := conn.Handshake(); err != nil {
		m.push(func() {
			_ = m.emit(s.obj, "tlsClientError", m.errorValue(err, "ERR_TLS_HANDSHAKE").V())
		})
		_ = raw.Close()
		return
	}
	m.push(func() {
		if s.closed {
			_ = conn.Close()
			return
		}
		v, err := m.makeSocket.Object().Call(js.Object{}, m.e.JS.Bool(true))
		if err != nil {
			_ = conn.Close()
			return
		}
		obj := v.Object()
		sock := m.sockOf(obj)
		if sock == nil {
			_ = conn.Close()
			return
		}
		m.attachSocket(sock, conn, s)
		s.conns[sock.id] = sock
		m.emit(s.obj, "secureConnection", obj.V())
	})
}

func (m *module) attachSocket(s *socket, conn *tls.Conn, srv *server) {
	state := conn.ConnectionState()
	s.conn = conn
	s.server = srv
	s.serverSide = srv != nil
	s.connecting = false
	s.secure = true
	s.gen++
	gen := s.gen
	s.handle = true
	s.remoteAddress, s.remotePort = splitHostPort(conn.RemoteAddr().String())
	s.localAddress, s.localPort = splitHostPort(conn.LocalAddr().String())
	s.protocol = state.NegotiatedProtocol
	s.alpn = state.NegotiatedProtocol
	s.cipher = tls.CipherSuiteName(state.CipherSuite)
	s.servername = state.ServerName
	s.authorized = len(state.VerifiedChains) > 0
	s.peer = m.peerCertificate(state)
	s.obj.SetBool("encrypted", true)
	s.obj.SetBool("authorized", s.authorized)
	s.obj.Set("authorizationError", m.e.JS.Null())
	s.obj.SetBool("connecting", false)
	s.obj.SetBool("secureConnecting", false)
	s.obj.SetBool("destroyed", false)
	s.obj.SetBool("readable", true)
	s.obj.SetBool("writable", true)
	s.obj.SetStr("readyState", "open")
	s.obj.Set("alpnProtocol", stringOrFalse(m.e, s.alpn))
	s.obj.SetStr("servername", s.servername)
	s.obj.SetStr("remoteAddress", s.remoteAddress)
	s.obj.SetNum("remotePort", float64(s.remotePort))
	s.obj.SetStr("localAddress", s.localAddress)
	s.obj.SetNum("localPort", float64(s.localPort))
	s.obj.SetNum("bytesRead", 0)
	s.obj.SetNum("bytesWritten", 0)
	m.addHandleIfNeeded(s)
	for _, data := range s.pending {
		_, _ = conn.Write(data)
	}
	s.pending = nil
	m.startReading(s, conn, gen)
}

func (m *module) addHandleIfNeeded(s *socket) {
	if s.handle {
		return
	}
	s.handle = true
	m.addHandle()
}

func (m *module) startReading(s *socket, conn *tls.Conn, gen int) {
	go func() {
		buf := make([]byte, 64*1024)
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				data := append([]byte(nil), buf[:n]...)
				m.push(func() { m.deliver(s, data, gen) })
			}
			if err != nil {
				m.push(func() { m.finishRead(s, err, gen) })
				return
			}
		}
	}()
}

func (m *module) deliver(s *socket, data []byte, gen int) {
	s.muSafe(func(x *socket) bool { return !x.closed && x.gen == gen })
	if s.paused {
		s.queue = append(s.queue, data)
		return
	}
	s.bytesRead += len(data)
	s.obj.SetNum("bytesRead", float64(s.bytesRead))
	fn := s.obj.Get("_receive")
	if fn.Valid() && fn.IsFunction() {
		_, _ = fn.Object().Call(s.obj, m.buffer(data))
		return
	}
	_ = m.emit(s.obj, "data", m.buffer(data))
}

func (s *socket) muSafe(fn func(*socket) bool) {
	// State is only touched on the JS thread; this helper keeps the read path
	// readable while the genuine synchronization is the pump queue.
	if !fn(s) {
		s.closed = true
	}
}

func (m *module) finishRead(s *socket, err error, gen int) {
	if s.closed || s.gen != gen {
		return
	}
	s.readEnded = true
	s.obj.SetBool("readable", false)
	m.emit(s.obj, "end")
	if err != nil && !errors.Is(err, io.EOF) {
		_ = m.emit(s.obj, "error", m.errorValue(err, "ERR_TLS_READ").V())
	}
	m.closeSocket(s, true)
}

func (m *module) closeSocket(s *socket, notify bool) {
	if s.closed {
		return
	}
	s.closed = true
	s.secure = false
	s.connecting = false
	s.obj.SetBool("destroyed", true)
	s.obj.SetBool("readable", false)
	s.obj.SetBool("writable", false)
	s.obj.SetBool("connecting", false)
	s.obj.SetStr("readyState", "closed")
	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.handle {
		s.handle = false
		m.dropHandle()
	}
	if s.server != nil {
		delete(s.server.conns, s.id)
	}
	if notify {
		m.emitAsync(s.obj, "close")
	}
}

func (m *module) hostConnect(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.TLSSocket")
	}
	opts := arg(args, 0)
	cb := arg(args, 1)
	if !opts.Valid() || !opts.IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"options\" argument must be an object")
	}
	o := opts.Object()
	port := int(o.Get("port").Int())
	host := ""
	if v := o.Get("host"); v.IsString() {
		host = v.Str()
	}
	if host == "" {
		host = "localhost"
	}
	path := ""
	if v := o.Get("path"); v.IsString() {
		path = v.Str()
	}
	cfg, err := clientConfig(o, host)
	if err != nil {
		return js.Value{}, err
	}
	s.connecting = true
	this.SetBool("connecting", true)
	this.SetBool("secureConnecting", true)
	s.handle = true
	m.addHandle()
	go func() {
		var conn *tls.Conn
		var dialErr error
		if path != "" {
			raw, dErr := net.Dial("unix", path)
			if dErr != nil {
				dialErr = dErr
			} else {
				conn = tls.Client(raw, cfg)
				dialErr = conn.Handshake()
			}
		} else {
			conn, dialErr = tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp",
				net.JoinHostPort(host, strconv.Itoa(port)), cfg)
		}
		m.push(func() {
			if s.closed {
				if conn != nil {
					_ = conn.Close()
				}
				return
			}
			if dialErr != nil {
				s.connecting = false
				s.handle = false
				m.dropHandle()
				s.obj.SetBool("connecting", false)
				s.obj.SetBool("secureConnecting", false)
				errObj := m.errorValue(dialErr, "ERR_TLS_HANDSHAKE")
				_ = m.emit(s.obj, "error", errObj.V())
				if cb.Valid() && cb.IsFunction() {
					_, _ = cb.Object().Call(s.obj, errObj.V())
				}
				return
			}
			m.attachSocket(s, conn, nil)
			m.emitAsync(s.obj, "secureConnect")
			if cb.Valid() && cb.IsFunction() {
				m.e.NextTick(cb)
			}
		})
	}()
	return this.V(), nil
}

func (m *module) hostWrite(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.TLSSocket")
	}
	data, err := bytesOf(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	cb := arg(args, 1)
	if s.conn == nil || !s.secure {
		s.pending = append(s.pending, data)
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb)
		}
		return this.V(), nil
	}
	n, err := s.conn.Write(data)
	s.bytesWritten += n
	s.obj.SetNum("bytesWritten", float64(s.bytesWritten))
	if err != nil {
		errObj := m.errorValue(err, "ERR_TLS_WRITE")
		_ = m.emit(s.obj, "error", errObj.V())
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb, errObj.V())
		}
		return this.V(), nil
	}
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) hostEnd(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.TLSSocket")
	}
	if data, err := bytesOf(arg(args, 0)); err == nil && len(data) > 0 {
		if s.conn != nil {
			_, _ = s.conn.Write(data)
		} else {
			s.pending = append(s.pending, data)
		}
	}
	if s.conn != nil {
		_ = s.conn.CloseWrite()
	}
	s.wroteEnd = true
	s.obj.SetBool("writable", false)
	if cb := arg(args, 1); cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) hostDestroy(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.TLSSocket")
	}
	m.closeSocket(s, true)
	return this.V(), nil
}

func (m *module) hostAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || s.conn == nil {
		return m.e.JS.Null(), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("address", s.localAddress)
	o.SetNum("port", float64(s.localPort))
	o.SetStr("family", familyOf(s.localAddress))
	return o.V(), nil
}

func (m *module) hostPeerCertificate(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || !s.peer.Valid() {
		return m.e.JS.Null(), nil
	}
	return s.peer.V(), nil
}

func (m *module) hostGetProtocol(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return m.e.JS.Null(), nil
	}
	return stringOrFalse(m.e, s.protocol), nil
}

func (m *module) hostGetCipher(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || s.cipher == "" {
		return m.e.JS.Undefined(), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("name", s.cipher)
	o.SetStr("standardName", s.cipher)
	o.SetStr("version", "TLSv1.3")
	return o.V(), nil
}

func (m *module) hostIsSessionReused(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || s.conn == nil {
		return m.e.JS.Bool(false), nil
	}
	return m.e.JS.Bool(s.conn.ConnectionState().DidResume), nil
}

func (m *module) hostGetSession(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || s.conn == nil {
		return m.e.JS.Undefined(), nil
	}
	session := s.conn.ConnectionState().TLSUnique
	if len(session) == 0 {
		return m.e.JS.Undefined(), nil
	}
	return m.buffer(session), nil
}

func (m *module) hostSetServername(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.TLSSocket")
	}
	s.servername = argStr(args, 0)
	return this.V(), nil
}

func (m *module) hostServerAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil || !s.listening {
		return m.e.JS.Null(), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("address", s.address)
	o.SetNum("port", float64(s.port))
	o.SetStr("family", s.family)
	return o.V(), nil
}

func (m *module) hostServerClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a tls.Server")
	}
	cb := arg(args, 0)
	if !s.listening {
		return this.V(), nil
	}
	s.listening = false
	s.closed = true
	s.obj.SetBool("listening", false)
	s.closeCb = cb
	if s.ln != nil {
		_ = s.ln.Close()
	}
	for id, sock := range s.conns {
		m.closeSocket(sock, true)
		delete(s.conns, id)
	}
	m.finishServerClose(s)
	return this.V(), nil
}

func (m *module) finishServerClose(s *server) {
	if s.closeSent {
		return
	}
	s.closeSent = true
	s.listening = false
	s.closed = true
	s.obj.SetBool("listening", false)
	if s.handle {
		s.handle = false
		m.dropHandle()
	}
	m.emitAsync(s.obj, "close")
	if s.closeCb.Valid() && s.closeCb.IsFunction() {
		m.e.NextTick(s.closeCb)
		s.closeCb = js.Value{}
	}
}

func (m *module) hostGetConnections(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	n := 0
	if s != nil {
		n = len(s.conns)
	}
	if cb := arg(args, 0); cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Num(float64(n)))
	}
	return this.V(), nil
}

func (m *module) peerCertificate(state tls.ConnectionState) js.Object {
	if len(state.PeerCertificates) == 0 {
		return js.Object{}
	}
	cert := state.PeerCertificates[0]
	o := m.e.JS.Obj()
	o.SetStr("subject", cert.Subject.String())
	o.SetStr("issuer", cert.Issuer.String())
	o.SetStr("valid_from", cert.NotBefore.Format(time.RFC1123))
	o.SetStr("valid_to", cert.NotAfter.Format(time.RFC1123))
	o.SetStr("serialNumber", fmt.Sprintf("%x", cert.SerialNumber))
	o.SetStr("fingerprint", fingerprint(cert.Raw))
	o.SetStr("fingerprint256", fingerprint256(cert.Raw))
	o.Set("raw", m.buffer(cert.Raw))
	o.SetStr("subjectaltname", altNames(cert))
	o.SetStr("pubkey", "")
	return o.Keep()
}

func fingerprint(raw []byte) string {
	return strings.ToUpper(hexUpper(raw[:min(len(raw), 32)]))
}

func fingerprint256(raw []byte) string {
	return strings.ToUpper(hexUpper(raw[:min(len(raw), 64)]))
}

func hexUpper(b []byte) string {
	const chars = "0123456789ABCDEF"
	out := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			out = append(out, ':')
		}
		out = append(out, chars[v>>4], chars[v&15])
	}
	return string(out)
}

func altNames(cert *x509.Certificate) string {
	var names []string
	for _, n := range cert.DNSNames {
		names = append(names, "DNS:"+n)
	}
	for _, n := range cert.IPAddresses {
		names = append(names, "IP Address:"+n.String())
	}
	return strings.Join(names, ",")
}

func (m *module) errorValue(err error, code string) js.Object {
	o := m.e.JS.Error(err.Error())
	if code != "" {
		o.SetStr("code", code)
	}
	return o
}

func (m *module) buffer(data []byte) js.Value {
	if m.bufferFrom.Valid() && m.bufferCtor.Valid() {
		if v, err := m.bufferFrom.Call(m.bufferCtor, m.e.JS.Bytes(data).V()); err == nil {
			return v
		}
	}
	return m.e.JS.Bytes(data).V()
}

func (m *module) emit(o js.Object, name string, args ...js.Value) error {
	fn := o.Get("emit")
	if !fn.Valid() || !fn.IsFunction() {
		return nil
	}
	_, err := fn.Object().Call(o, append([]js.Value{m.e.JS.Str(name)}, args...)...)
	return err
}

func (m *module) emitAsync(o js.Object, name string, args ...js.Value) {
	payload := append([]js.Value(nil), args...)
	holder := m.e.JS.Obj()
	holder.SetFuncHidden("fire", func(this js.Object, _ []js.Value) (js.Value, error) {
		_ = m.emit(o, name, payload...)
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(holder.Get("fire"))
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func argStr(args []js.Value, i int) string {
	v := arg(args, i)
	if v.Valid() && v.IsString() {
		return v.Str()
	}
	return ""
}

func bytesOf(v js.Value) ([]byte, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil, nil
	}
	if v.IsString() {
		return []byte(v.Str()), nil
	}
	if b, ok := v.Bytes(); ok {
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"data\" argument must be a string or Buffer")
}

func splitHostPort(addr string) (string, int) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, 0
	}
	n, _ := strconv.Atoi(port)
	return host, n
}

func familyOf(host string) string {
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		return "IPv6"
	}
	return "IPv4"
}

func stringOrFalse(e *env.Env, s string) js.Value {
	if s == "" {
		return e.JS.Bool(false)
	}
	return e.JS.Str(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func optionBytes(o js.Object, name string) []byte {
	v := o.Get(name)
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil
	}
	if v.IsString() {
		return []byte(v.Str())
	}
	if b, ok := v.Bytes(); ok {
		return b
	}
	if v.IsArray() {
		arr := v.Object()
		var out []byte
		for i := 0; i < arr.Length(); i++ {
			item := arr.At(i)
			if item.IsString() {
				out = append(out, []byte(item.Str())...)
			} else if b, ok := item.Bytes(); ok {
				out = append(out, b...)
			}
		}
		return out
	}
	return nil
}

func version(name string, fallback uint16) uint16 {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "TLSV1", "TLSV1.0", "TLS1", "TLS1.0":
		return tls.VersionTLS10
	case "TLSV1.1", "TLS1.1":
		return tls.VersionTLS11
	case "TLSV1.2", "TLS1.2":
		return tls.VersionTLS12
	case "TLSV1.3", "TLS1.3":
		return tls.VersionTLS13
	}
	return fallback
}

func commonConfig(o js.Object, serverSide bool, host string) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS13}
	if v := o.Get("minVersion"); v.IsString() {
		cfg.MinVersion = version(v.Str(), cfg.MinVersion)
	}
	if v := o.Get("maxVersion"); v.IsString() {
		cfg.MaxVersion = version(v.Str(), cfg.MaxVersion)
	}
	if v := o.Get("ciphers"); v.IsString() && v.Str() != "" {
		for _, wanted := range strings.Split(v.Str(), ":") {
			for _, suite := range tls.CipherSuites() {
				if strings.EqualFold(suite.Name, strings.TrimSpace(wanted)) {
					cfg.CipherSuites = append(cfg.CipherSuites, suite.ID)
				}
			}
		}
	}
	if v := o.Get("ca"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(optionBytes(o, "ca")) {
			return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CA", "Unable to load CA certificates")
		}
		if serverSide {
			cfg.ClientCAs = pool
		} else {
			cfg.RootCAs = pool
		}
	}
	if v := o.Get("requestCert"); v.IsBool() && v.Bool() {
		cfg.ClientAuth = tls.RequestClientCert
	}
	if !serverSide {
		cfg.ServerName = host
		if v := o.Get("servername"); v.IsString() && v.Str() != "" {
			cfg.ServerName = v.Str()
		}
		if v := o.Get("rejectUnauthorized"); v.IsBool() {
			cfg.InsecureSkipVerify = !v.Bool()
		}
	}
	if v := o.Get("ALPNProtocols"); v.IsArray() {
		arr := v.Object()
		for i := 0; i < arr.Length(); i++ {
			if item := arr.At(i); item.IsString() {
				cfg.NextProtos = append(cfg.NextProtos, item.Str())
			}
		}
	}
	return cfg, nil
}

func serverConfig(value js.Value) (*tls.Config, error) {
	if !value.Valid() || !value.IsObject() {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"options\" argument must be an object")
	}
	o := value.Object()
	cfg, err := commonConfig(o, true, "")
	if err != nil {
		return nil, err
	}
	certPEM, keyPEM := optionBytes(o, "cert"), optionBytes(o, "key")
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT", "A TLS server requires both 'cert' and 'key' options")
	}
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT", err.Error())
	}
	cfg.Certificates = []tls.Certificate{pair}
	return cfg, nil
}

func clientConfig(o js.Object, host string) (*tls.Config, error) {
	cfg, err := commonConfig(o, false, host)
	if err != nil {
		return nil, err
	}
	certPEM, keyPEM := optionBytes(o, "cert"), optionBytes(o, "key")
	if len(certPEM) > 0 && len(keyPEM) > 0 {
		pair, pairErr := tls.X509KeyPair(certPEM, keyPEM)
		if pairErr != nil {
			return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT", pairErr.Error())
		}
		cfg.Certificates = []tls.Certificate{pair}
	}
	return cfg, nil
}

func rootCertificates() []string {
	paths := []string{
		"/etc/ssl/certs/ca-certificates.crt",
		"/etc/pki/tls/certs/ca-bundle.crt",
		"/etc/ssl/ca-bundle.pem",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		parts := strings.Split(string(data), "-----END CERTIFICATE-----")
		var out []string
		for _, part := range parts {
			if strings.Contains(part, "-----BEGIN CERTIFICATE-----") {
				out = append(out, strings.TrimSpace(part)+"\n-----END CERTIFICATE-----\n")
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

const prelude = `(function (host, EventEmitter, Buffer) {
  'use strict';

  function describe(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (typeof value === 'string') return "type string ('" + value + "')";
    if (typeof value === 'number') return 'type number (' + value + ')';
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }

  function invalidArg(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function toBuffer(chunk, encoding) {
    if (typeof chunk === 'string') return Buffer.from(chunk, encoding || 'utf8');
    if (Buffer.isBuffer(chunk)) return chunk;
    if (chunk instanceof Uint8Array) return Buffer.from(chunk.buffer, chunk.byteOffset, chunk.byteLength);
    throw invalidArg('chunk', 'of type string or an instance of Buffer or Uint8Array', chunk);
  }

  class TLSSocket extends EventEmitter {
    constructor(serverSide) {
      super();
      this.encrypted = true;
      this.authorized = false;
      this.authorizationError = null;
      this.alpnProtocol = false;
      this.connecting = false;
      this.secureConnecting = false;
      this.destroyed = false;
      this.readable = true;
      this.writable = true;
      this._encoding = null;
      this._paused = false;
      this._queue = [];
      host.initSock.call(this, !!serverSide);
    }
    _receive(buf) {
      if (this._paused) { this._queue.push(buf); return; }
      if (this._encoding) buf = buf.toString(this._encoding);
      this.emit('data', buf);
    }
    setEncoding(encoding) { this._encoding = encoding; return this; }
    pause() { this._paused = true; return this; }
    resume() {
      this._paused = false;
      while (this._queue.length) this.emit('data', this._queue.shift());
      return this;
    }
    write(chunk, encoding, cb) {
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var data = toBuffer(chunk, encoding);
      host.write.call(this, data, cb);
      return true;
    }
    end(chunk, encoding, cb) {
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; encoding = undefined; }
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var data = chunk === undefined || chunk === null ? Buffer.alloc(0) : toBuffer(chunk, encoding);
      host.end.call(this, data, cb);
      return this;
    }
    destroy(err) {
      host.destroy.call(this);
      if (err) this.emit('error', err);
      return this;
    }
    address() { return host.address.call(this); }
    getPeerCertificate() { return host.peerCertificate.call(this); }
    getProtocol() { return host.getProtocol.call(this); }
    getCipher() { return host.getCipher.call(this); }
    isSessionReused() { return host.isSessionReused.call(this); }
    getSession() { return host.getSession.call(this); }
    setServername(name) { host.setServername.call(this, String(name)); return this; }
    ref() { return this; }
    unref() { return this; }
  }

  class Server extends EventEmitter {
    constructor(options, listener) {
      super();
      options = options || {};
      this._options = options;
      this.listening = false;
      host.initServer.call(this, options);
      if (typeof listener === 'function') this.on('secureConnection', listener);
    }
    listen() { return host.listen.call(this, Array.prototype.slice.call(arguments)); }
    close(cb) { return host.serverClose.call(this, cb); }
    address() { return host.serverAddress.call(this); }
    getConnections(cb) { return host.getConnections.call(this, cb); }
    ref() { return this; }
    unref() { return this; }
  }

  function createServer(options, listener) {
    if (typeof options === 'function') { listener = options; options = {}; }
    return new Server(options, listener);
  }

  function connect() {
    var options = {};
    var callback;
    if (typeof arguments[0] === 'object') {
      options = arguments[0] || {};
      callback = arguments[1];
    } else {
      options = { port: arguments[0], host: arguments[1] };
      if (arguments[2] && typeof arguments[2] === 'object') {
        for (var key in arguments[2]) {
          if (Object.prototype.hasOwnProperty.call(arguments[2], key)) options[key] = arguments[2][key];
        }
        callback = arguments[3];
      } else {
        callback = typeof arguments[2] === 'function' ? arguments[2] : arguments[3];
      }
    }
    if (!options || (options.port === undefined && !options.path)) {
      throw invalidArg('options', 'an object with a port or path', options);
    }
    var socket = new TLSSocket(false);
    if (typeof callback === 'function') socket.once('secureConnect', callback);
    host.connect.call(socket, options);
    return socket;
  }

  function createSecureContext(options) {
    return { context: options || {} };
  }

  function checkServerIdentity(hostname, cert) {
    if (!cert) {
      var missing = new Error('No certificate presented');
      missing.code = 'ERR_TLS_CERT_ALTNAME_INVALID';
      return missing;
    }
    var names = String(cert.subjectaltname || '').split(',');
    for (var i = 0; i < names.length; i++) {
      var name = names[i].trim();
      if (name.indexOf('DNS:') === 0 && name.slice(4) === hostname) return undefined;
      if (name.indexOf('IP Address:') === 0 && name.slice(11) === hostname) return undefined;
    }
    var err = new Error('Hostname/IP does not match certificate altnames: ' + (cert.subjectaltname || ''));
    err.code = 'ERR_TLS_CERT_ALTNAME_INVALID';
    return err;
  }

  return {
    Server: Server,
    TLSSocket: TLSSocket,
    createServer: createServer,
    connect: connect,
    createSecureContext: createSecureContext,
    checkServerIdentity: checkServerIdentity,
    makeSocket: function (serverSide) { return new TLSSocket(serverSide); }
  };
})`
