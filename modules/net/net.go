// Package net implements Node's net builtin on top of Go's net package.
//
// GodeJS runs one JavaScriptCore context on one goroutine, so socket I/O
// cannot call into JavaScript directly. Every goroutine instead queues a
// closure with push(); a 1ms repeating pump timer, installed only while a
// handle is active, drains that queue on the JavaScript thread and emits the
// events. An active listener or socket therefore keeps the run loop alive, and
// the timer is cleared as soon as the last handle closes.
package net

import (
	"encoding/base64"
	"encoding/hex"
	stdnet "net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"godejs/internal/clusterutil"
	"godejs/internal/env"
	"godejs/internal/js"

	// net.Socket extends events.EventEmitter and emits Buffers, so both
	// modules must be registered in this package's own test binary too.
	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("net", Install) }

// pumpEveryMS is how often the pump wakes while a handle is active.
const pumpEveryMS = 1

type module struct {
	e *env.Env

	mu     sync.Mutex
	queued []func()
	pumpID int
	active int

	socks map[float64]*sock
	srvs  map[float64]*srv
	seq   float64

	exports    js.Object
	host       js.Object
	makeSocket js.Value
	bufferFrom js.Object
	bufferCtor js.Object
}

type sock struct {
	id  float64
	obj js.Object

	conn stdnet.Conn
	wch  chan []byte

	connecting bool
	destroyed  bool
	readEnded  bool
	wroteEnd   bool
	closedSent bool
	paused     bool
	enc        string
	queue      [][]byte

	allowHalfOpen bool
	// pendingWrites buffers write() calls made before the connection is
	// established, the way Node's Socket does.
	pendingWrites [][]byte

	remoteAddress string
	remotePort    int
	remoteFamily  string
	localAddress  string
	localPort     int

	bytesRead    int
	bytesWritten int

	timeoutMS  float64
	timeoutCb  js.Value
	timeoutID  int
	lastActive time.Time
	needDrain  bool
	handleHeld bool

	server *srv
}

type srv struct {
	id  float64
	obj js.Object

	ln        stdnet.Listener
	listening bool
	closed    bool

	address string
	port    int
	family  string

	connections map[float64]*sock

	timeoutMS float64
	refed     bool
	closeCb   js.Value
}

// Install builds the net module's exports object.
func Install(e *env.Env) js.Object {
	m := &module{
		e:     e,
		socks: map[float64]*sock{},
		srvs:  map[float64]*srv{},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()

	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("initSock", m.hostInitSock)
	m.host.SetFuncHidden("sockConnect", m.hostSockConnect)
	m.host.SetFuncHidden("sockWrite", m.hostSockWrite)
	m.host.SetFuncHidden("sockEnd", m.hostSockEnd)
	m.host.SetFuncHidden("sockDestroy", m.hostSockDestroy)
	m.host.SetFuncHidden("sockPause", m.hostSockPause)
	m.host.SetFuncHidden("sockSetTimeout", m.hostSockSetTimeout)
	m.host.SetFuncHidden("sockSetEncoding", m.hostSockSetEncoding)
	m.host.SetFuncHidden("sockAddress", m.hostSockAddress)
	m.host.SetFuncHidden("sockRef", m.hostSockRef)
	m.host.SetFuncHidden("initServer", m.hostInitServer)
	m.host.SetFuncHidden("srvListen", m.hostSrvListen)
	m.host.SetFuncHidden("srvClose", m.hostSrvClose)
	m.host.SetFuncHidden("srvAddress", m.hostSrvAddress)
	m.host.SetFuncHidden("srvGetConnections", m.hostSrvGetConnections)
	m.host.SetFuncHidden("parseIP", m.hostParseIP)

	events := e.Exports("events")
	ee := events.Get("EventEmitter")
	if !ee.Valid() || !ee.IsFunction() {
		panic("godejs net: events.EventEmitter is unavailable")
	}
	buffer := e.Exports("buffer").Get("Buffer")
	if buffer.Valid() && buffer.IsFunction() {
		m.bufferCtor = buffer.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}

	api, err := e.JS.Eval(prelude, "net.js")
	if err != nil {
		panic("godejs net: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V(), ee)
	if err != nil {
		panic("godejs net: " + err.Error())
	}
	obj := built.Object()
	m.makeSocket = obj.Get("makeSocket").Keep()
	for _, name := range obj.OwnKeys() {
		if name == "makeSocket" || name == "makeServer" {
			continue
		}
		m.exports.Set(name, obj.Get(name))
	}

	m.exports.SetFunc("isIP", func(this js.Object, args []js.Value) (js.Value, error) {
		s := argStr(args, 0)
		switch ipVersion(s) {
		case 4:
			return e.JS.Num(4), nil
		case 6:
			return e.JS.Num(6), nil
		}
		return e.JS.Num(0), nil
	})
	m.exports.SetFunc("isIPv4", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Bool(ipVersion(argStr(args, 0)) == 4), nil
	})
	m.exports.SetFunc("isIPv6", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Bool(ipVersion(argStr(args, 0)) == 6), nil
	})

	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"net.BlockList",
		"net.BlockList#addAddress",
		"net.BlockList#addRange",
		"net.BlockList#addSubnet",
		"net.BlockList#check",
		"net.Server",
		"net.Server#address",
		"net.Server#close",
		"net.Server#getConnections",
		"net.Server#listen",
		"net.Server#ref",
		"net.Server#unref",
		"net.Socket",
		"net.Socket#address",
		"net.Socket#connect",
		"net.Socket#destroy",
		"net.Socket#end",
		"net.Socket#pause",
		"net.Socket#resume",
		"net.Socket#setEncoding",
		"net.Socket#setKeepAlive",
		"net.Socket#setNoDelay",
		"net.Socket#setTimeout",
		"net.Socket#write",
		"net.SocketAddress",
		"net.connect",
		"net.createConnection",
		"net.createServer",
		"net.getDefaultAutoSelectFamily",
		"net.getDefaultAutoSelectFamilyAttemptTimeout",
		"net.isIP",
		"net.isIPv4",
		"net.isIPv6",
		"net.setDefaultAutoSelectFamily",
		"net.setDefaultAutoSelectFamilyAttemptTimeout",
	}
}

// ------------------------------------------------------------------- helpers --

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

func ipVersion(s string) int {
	ip := stdnet.ParseIP(s)
	if ip == nil {
		return 0
	}
	if ip.To4() != nil {
		return 4
	}
	return 6
}

// push queues work for the pump, which runs it on the JavaScript thread.
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
	tick := m.host.Get("pump")
	id := m.e.SetTimeout(tick, pumpEveryMS, true, nil)
	m.mu.Lock()
	m.pumpID = id
	m.mu.Unlock()
}

func (m *module) shutdown() {
	m.mu.Lock()
	socks := make([]*sock, 0, len(m.socks))
	for _, s := range m.socks {
		socks = append(socks, s)
	}
	srvs := make([]*srv, 0, len(m.srvs))
	for _, s := range m.srvs {
		srvs = append(srvs, s)
	}
	id := m.pumpID
	m.pumpID = 0
	m.active = 0
	m.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
	for _, s := range srvs {
		if s.ln != nil {
			_ = s.ln.Close()
		}
	}
	for _, s := range socks {
		if s.conn != nil {
			_ = s.conn.Close()
		}
	}
}

func (m *module) nextID() float64 {
	m.seq++
	return m.seq
}

// buffer returns a Node Buffer holding a copy of b.
func (m *module) buffer(b []byte) js.Value {
	if m.bufferFrom.Valid() && m.bufferCtor.Valid() {
		if v, err := m.bufferFrom.Call(m.bufferCtor, m.e.JS.Bytes(b).V()); err == nil {
			return v
		}
	}
	return m.e.JS.Bytes(b).V()
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

// toBytes converts a write chunk (string or Buffer/TypedArray) to bytes.
func (m *module) toBytes(v js.Value, encoding string) ([]byte, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "chunk" argument must be of type string or an instance of Buffer or Uint8Array. Received `+valueKind(v))
	}
	if v.IsString() {
		return encodeString(v.Str(), encoding), nil
	}
	if b, ok := v.Bytes(); ok {
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "chunk" argument must be of type string or an instance of Buffer or Uint8Array. Received `+valueKind(v))
}

func valueKind(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsBool():
		return "type boolean"
	case v.IsArray():
		return "an instance of Array"
	case v.IsObject():
		return "an instance of Object"
	}
	return "type " + v.Str()
}

func encodeString(s, encoding string) []byte {
	switch strings.ToLower(encoding) {
	case "", "utf8", "utf-8":
		return []byte(s)
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		units := utf16.Encode([]rune(s))
		out := make([]byte, 0, len(units)*2)
		for _, u := range units {
			out = append(out, byte(u), byte(u>>8))
		}
		return out
	case "latin1", "binary":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			out = append(out, byte(r))
		}
		return out
	case "ascii":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			if r > 0x7f {
				r = 0x7f
			}
			out = append(out, byte(r))
		}
		return out
	case "hex":
		out, err := hex.DecodeString(strings.TrimSpace(s))
		if err != nil {
			return nil
		}
		return out
	case "base64":
		out, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil
		}
		return out
	case "base64url":
		out, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
		if err != nil {
			return nil
		}
		return out
	}
	return []byte(s)
}

// ---------------------------------------------------------------- socket host --

func (m *module) sockOf(this js.Object) *sock {
	if v := this.Get("$godeNetId"); v.IsNumber() {
		m.mu.Lock()
		s := m.socks[v.Float()]
		m.mu.Unlock()
		return s
	}
	return nil
}

func (m *module) srvOf(this js.Object) *srv {
	if v := this.Get("$godeNetId"); v.IsNumber() {
		m.mu.Lock()
		s := m.srvs[v.Float()]
		m.mu.Unlock()
		return s
	}
	return nil
}

func (m *module) hostInitSock(this js.Object, args []js.Value) (js.Value, error) {
	id := m.nextID()
	s := &sock{id: id, obj: this, lastActive: time.Now()}
	m.mu.Lock()
	m.socks[id] = s
	m.mu.Unlock()
	this.SetHidden("$godeNetId", m.e.JS.Num(id))
	this.Protect()

	opts := arg(args, 0)
	if opts.Valid() && opts.IsObject() {
		o := opts.Object()
		if v := o.Get("allowHalfOpen"); v.IsBool() {
			s.allowHalfOpen = v.Bool()
		}
		if v := o.Get("encoding"); v.IsString() {
			s.enc = strings.ToLower(v.Str())
		}
		if v := o.Get("timeout"); v.IsNumber() {
			s.timeoutMS = v.Float()
		}
	}
	s.setProps()
	return this.V(), nil
}

// setProps mirrors the Go state onto the visible JavaScript properties.
func (s *sock) setProps() {
	s.obj.SetBool("connecting", s.connecting)
	s.obj.SetBool("destroyed", s.destroyed)
	s.obj.SetBool("pending", s.connecting)
	state := "open"
	switch {
	case s.destroyed:
		state = "closed"
	case s.connecting:
		state = "opening"
	case s.readEnded || s.wroteEnd:
		state = "readOnly"
	}
	s.obj.SetStr("readyState", state)
	if s.remoteAddress != "" {
		s.obj.SetStr("remoteAddress", s.remoteAddress)
		s.obj.SetNum("remotePort", float64(s.remotePort))
		if s.remoteFamily == "" {
			s.remoteFamily = "IPv4"
		}
		s.obj.SetStr("remoteFamily", s.remoteFamily)
	}
	if s.localAddress != "" {
		s.obj.SetStr("localAddress", s.localAddress)
		s.obj.SetNum("localPort", float64(s.localPort))
	}
	s.obj.SetNum("bytesRead", float64(s.bytesRead))
	s.obj.SetNum("bytesWritten", float64(s.bytesWritten))
	s.obj.SetBool("readable", !s.readEnded && !s.destroyed)
	s.obj.SetBool("writable", !s.wroteEnd && !s.destroyed)
}

func (m *module) hostSockConnect(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Socket")
	}
	target, cb, err := m.parseTarget(arg(args, 0), arg(args, 1), arg(args, 2))
	if err != nil {
		return js.Value{}, err
	}
	if s.connecting || s.conn != nil || s.destroyed {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_ALREADY_BOUND", "Socket is already bound")
	}
	s.connecting = true
	s.setProps()
	s.handleHeld = true
	m.addHandle()
	go func() {
		conn, dialErr := stdnet.Dial(target.network, target.address)
		m.push(func() {
			if s.destroyed {
				if conn != nil {
					_ = conn.Close()
				}
				return
			}
			if dialErr != nil {
				s.connecting = false
				s.setProps()
				_ = m.emit(s.obj, "error", m.connectionError(dialErr, target).V())
				m.teardown(s, true)
				return
			}
			m.attach(s, conn)
			s.connecting = false
			s.setProps()
			_ = m.emit(s.obj, "connect")
			_ = m.emit(s.obj, "ready")
			if cb.Valid() && cb.IsFunction() {
				_, _ = cb.Object().Call(s.obj)
			}
		})
	}()
	return this.V(), nil
}

func (m *module) connectionError(err error, t target) js.Object {
	code := "ECONNREFUSED"
	switch {
	case strings.Contains(err.Error(), "no such file"):
		code = "ENOENT"
	case strings.Contains(err.Error(), "i/o timeout"):
		code = "ETIMEDOUT"
	case strings.Contains(err.Error(), "no route"):
		code = "EHOSTUNREACH"
	case strings.Contains(err.Error(), "network is unreachable"):
		code = "ENETUNREACH"
	}
	e := m.e.JS.Error("connect " + code + " " + t.address)
	e.SetStr("code", code)
	e.SetNum("errno", float64(errnoOf(code)))
	e.SetStr("syscall", "connect")
	e.SetStr("address", t.addressForError)
	e.SetNum("port", float64(t.portForError))
	return e
}

func errnoOf(code string) int {
	switch code {
	case "ENOENT":
		return -2
	case "ECONNREFUSED":
		return -111
	case "ETIMEDOUT":
		return -110
	case "EHOSTUNREACH":
		return -113
	case "ENETUNREACH":
		return -101
	}
	return 0
}

func (m *module) attach(s *sock, conn stdnet.Conn) {
	s.conn = conn
	s.wch = make(chan []byte, 4096)
	if a := conn.RemoteAddr(); a != nil {
		s.remoteAddress, s.remotePort = splitHostPort(a.String())
		s.remoteFamily = familyOf(a)
	}
	if a := conn.LocalAddr(); a != nil {
		s.localAddress, s.localPort = splitHostPort(a.String())
	}
	m.startReading(s, conn)
	m.startWriting(s, conn)
	if len(s.pendingWrites) > 0 && s.wch != nil {
		for _, w := range s.pendingWrites {
			s.wch <- w
		}
		s.pendingWrites = nil
	}
	if s.wroteEnd && s.wch != nil {
		// nil is the writer goroutine's "half-close after draining" marker.
		s.wch <- nil
	}
	if s.timeoutMS > 0 {
		m.armTimeout(s, s.timeoutMS)
	}
}

func splitHostPort(addr string) (string, int) {
	host, portStr, err := stdnet.SplitHostPort(addr)
	if err != nil {
		return addr, 0
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func familyOf(a stdnet.Addr) string {
	host, _ := splitHostPort(a.String())
	if ip := stdnet.ParseIP(host); ip != nil && ip.To4() == nil {
		return "IPv6"
	}
	return "IPv4"
}

func (m *module) startReading(s *sock, conn stdnet.Conn) {
	go func() {
		buf := make([]byte, 64*1024)
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				data := append([]byte(nil), buf[:n]...)
				m.push(func() { m.onData(s, data) })
			}
			if err != nil {
				m.push(func() { m.onReadEnd(s) })
				return
			}
		}
	}()
}

func (m *module) startWriting(s *sock, conn stdnet.Conn) {
	go func() {
		for b := range s.wch {
			if b == nil {
				if tcp, ok := conn.(*stdnet.TCPConn); ok && tcp != nil {
					_ = tcp.CloseWrite()
				}
				continue
			}
			if _, err := conn.Write(b); err != nil {
				n := len(b)
				m.push(func() { m.onWriteError(s, n) })
				return
			}
			n := len(b)
			m.push(func() { m.afterWrite(s, n) })
		}
	}()
}

func (m *module) onData(s *sock, data []byte) {
	if s.destroyed || s.readEnded {
		return
	}
	s.bytesRead += len(data)
	s.lastActive = time.Now()
	if s.paused {
		s.queue = append(s.queue, data)
		s.setProps()
		return
	}
	m.emitData(s, data)
	s.setProps()
}

func (m *module) emitData(s *sock, data []byte) {
	var payload js.Value
	if s.enc != "" {
		payload = m.e.JS.Str(decodeString(data, s.enc))
	} else {
		payload = m.buffer(data)
	}
	_ = m.emit(s.obj, "data", payload)
}

func decodeString(b []byte, encoding string) string {
	switch strings.ToLower(encoding) {
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		units := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			units = append(units, uint16(b[i])|uint16(b[i+1])<<8)
		}
		return string(utf16.Decode(units))
	case "latin1", "binary":
		runes := make([]rune, len(b))
		for i, c := range b {
			runes[i] = rune(c)
		}
		return string(runes)
	case "hex":
		return hex.EncodeToString(b)
	case "base64":
		return base64.StdEncoding.EncodeToString(b)
	case "base64url":
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return string(b)
}

func (m *module) afterWrite(s *sock, n int) {
	if s.destroyed {
		return
	}
	s.bytesWritten += n
	s.lastActive = time.Now()
	s.setProps()
	if s.needDrain {
		s.needDrain = false
		_ = m.emit(s.obj, "drain")
	}
}

func (m *module) onWriteError(s *sock, n int) {
	if s.destroyed {
		return
	}
	s.bytesWritten += n
	m.teardown(s, true)
}

func (m *module) onReadEnd(s *sock) {
	if s.destroyed {
		return
	}
	if s.readEnded {
		return
	}
	s.readEnded = true
	// Deliver anything still buffered by pause() before 'end'.
	if len(s.queue) > 0 {
		pending := s.queue
		s.queue = nil
		for _, chunk := range pending {
			m.emitData(s, chunk)
		}
	}
	s.setProps()
	_ = m.emit(s.obj, "end")
	if s.wroteEnd || !s.allowHalfOpen {
		m.teardown(s, false)
	}
}

// teardown closes the connection, releases the handle and emits 'close' once.
func (m *module) teardown(s *sock, hadError bool) {
	if s.destroyed {
		return
	}
	s.destroyed = true
	s.connecting = false
	if s.timeoutID != 0 {
		m.e.ClearTimer(s.timeoutID)
		s.timeoutID = 0
	}
	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.wch != nil {
		close(s.wch)
		s.wch = nil
	}
	if s.server != nil {
		delete(s.server.connections, s.id)
	}
	if !s.closedSent {
		s.closedSent = true
		if s.handleHeld {
			s.handleHeld = false
			m.dropHandle()
		}
	}
	s.setProps()
	s.obj.SetBool("closed", true)
	_ = m.emit(s.obj, "close", m.e.JS.Bool(hadError))
}

func (m *module) hostSockWrite(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Socket")
	}
	if s.destroyed {
		return m.e.JS.Bool(false), js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot call write after a stream was destroyed")
	}
	if s.wroteEnd {
		return m.e.JS.Bool(false), js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	data, err := m.toBytes(arg(args, 0), argStr(args, 1))
	if err != nil {
		return js.Value{}, err
	}
	if s.conn == nil {
		s.pendingWrites = append(s.pendingWrites, data)
		return m.e.JS.Bool(true), nil
	}
	if s.wch != nil {
		s.wch <- data
	}
	s.lastActive = time.Now()
	if cb := arg(args, 2); cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, m.e.JS.Undefined())
	}
	return m.e.JS.Bool(true), nil
}

func (m *module) hostSockEnd(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Socket")
	}
	chunk := arg(args, 0)
	enc := argStr(args, 1)
	var cb js.Value
	if chunk.Valid() && chunk.IsFunction() {
		cb, chunk = chunk, js.Value{}
	} else if enc != "" {
		if v := arg(args, 2); v.Valid() && v.IsFunction() {
			cb = v
		}
	}
	if chunk.Valid() && !chunk.IsUndefined() && !chunk.IsNull() {
		data, err := m.toBytes(chunk, enc)
		if err != nil {
			return js.Value{}, err
		}
		switch {
		case s.conn == nil:
			s.pendingWrites = append(s.pendingWrites, data)
		case s.wch != nil:
			s.wch <- data
		}
	}
	if s.wroteEnd || s.destroyed {
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb)
		}
		return this.V(), nil
	}
	s.wroteEnd = true
	s.setProps()
	if s.conn != nil && s.wch != nil {
		// Ask the writer to half-close once the queued bytes are on the wire.
		s.wch <- nil
	}
	m.emitAsync(s.obj, "finish")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	if s.readEnded {
		m.teardown(s, false)
	}
	return this.V(), nil
}

func (m *module) hostSockDestroy(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Socket")
	}
	if s.destroyed {
		return this.V(), nil
	}
	errV := arg(args, 0)
	hadError := false
	if errV.Valid() && !errV.IsUndefined() && !errV.IsNull() {
		hadError = true
		if errV.IsObject() {
			_ = m.emit(s.obj, "error", errV)
		} else {
			_ = m.emit(s.obj, "error", m.e.JS.Error(errV.Str()).V())
		}
	}
	m.teardown(s, hadError)
	return this.V(), nil
}

func (m *module) hostSockPause(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return this.V(), nil
	}
	want := arg(args, 0).Bool()
	s.paused = want
	if !want && len(s.queue) > 0 {
		pending := s.queue
		s.queue = nil
		for _, chunk := range pending {
			m.emitData(s, chunk)
		}
	}
	return this.V(), nil
}

func (m *module) hostSockSetTimeout(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return this.V(), nil
	}
	ms := 0.0
	if v := arg(args, 0); v.IsNumber() {
		ms = v.Float()
	}
	s.timeoutMS = ms
	s.timeoutCb = arg(args, 1)
	if s.timeoutID != 0 {
		m.e.ClearTimer(s.timeoutID)
		s.timeoutID = 0
	}
	if ms > 0 && s.conn != nil {
		m.armTimeout(s, ms)
	}
	return this.V(), nil
}

func (m *module) armTimeout(s *sock, ms float64) {
	holder := m.e.JS.Obj()
	holder.SetFuncHidden("fire", func(this js.Object, _ []js.Value) (js.Value, error) {
		s.timeoutID = 0
		if s.destroyed || s.conn == nil {
			return m.e.JS.Undefined(), nil
		}
		elapsed := float64(time.Since(s.lastActive)) / float64(time.Millisecond)
		if elapsed+0.5 < ms {
			m.armTimeout(s, ms-elapsed)
			return m.e.JS.Undefined(), nil
		}
		if s.timeoutCb.Valid() && s.timeoutCb.IsFunction() {
			if _, err := s.timeoutCb.Object().Call(s.obj); err != nil {
				return js.Value{}, err
			}
		} else {
			_ = m.emit(s.obj, "timeout")
		}
		return m.e.JS.Undefined(), nil
	})
	s.timeoutID = m.e.SetTimeout(holder.Get("fire"), ms, false, nil)
}

func (m *module) hostSockSetEncoding(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil {
		return this.V(), nil
	}
	s.enc = strings.ToLower(argStr(args, 0))
	return this.V(), nil
}

func (m *module) hostSockAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sockOf(this)
	if s == nil || s.conn == nil {
		return m.e.JS.Obj().V(), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("address", s.localAddress)
	o.SetNum("port", float64(s.localPort))
	family := "IPv4"
	if ip := stdnet.ParseIP(s.localAddress); ip != nil && ip.To4() == nil {
		family = "IPv6"
	}
	o.SetStr("family", family)
	return o.V(), nil
}

func (m *module) hostSockRef(this js.Object, args []js.Value) (js.Value, error) {
	return this.V(), nil
}

// ---------------------------------------------------------------- server host --

func (m *module) hostInitServer(this js.Object, args []js.Value) (js.Value, error) {
	id := m.nextID()
	s := &srv{id: id, obj: this, connections: map[float64]*sock{}, refed: true}
	m.mu.Lock()
	m.srvs[id] = s
	m.mu.Unlock()
	this.SetHidden("$godeNetId", m.e.JS.Num(id))
	this.Protect()
	opts := arg(args, 0)
	if opts.Valid() && opts.IsObject() {
		if v := opts.Object().Get("requestTimeout"); v.IsNumber() {
			s.timeoutMS = v.Float()
		}
	}
	this.SetBool("listening", false)
	this.SetNum("maxConnections", 0)
	this.SetNum("connections", 0)
	return this.V(), nil
}

type target struct {
	network         string
	address         string
	addressForError string
	portForError    int
}

// parseTarget normalizes connect()/listen() arguments: (port[, host][, cb]),
// (path[, cb]) or (options[, cb]).
func (m *module) parseTarget(first, second, third js.Value) (target, js.Value, error) {
	var t target
	t.network = "tcp"
	var cb js.Value

	switch {
	case first.Valid() && first.IsNumber():
		port := int(first.Float())
		host := ""
		if second.Valid() && second.IsString() && second.Str() != "" {
			host = second.Str()
		} else if second.Valid() && second.IsFunction() {
			cb = second
		}
		if third.Valid() && third.IsFunction() {
			cb = third
		}
		if host == "" {
			host = "localhost"
		}
		t.address = stdnet.JoinHostPort(host, strconv.Itoa(port))
		t.addressForError = host
		t.portForError = port
	case first.Valid() && first.IsString():
		t.network = "unix"
		t.address = first.Str()
		t.addressForError = t.address
		if second.Valid() && second.IsFunction() {
			cb = second
		}
		if third.Valid() && third.IsFunction() {
			cb = third
		}
	case first.Valid() && first.IsObject():
		o := first.Object()
		if second.Valid() && second.IsFunction() {
			cb = second
		} else if third.Valid() && third.IsFunction() {
			cb = third
		}
		if p := o.Get("path"); p.IsString() && p.Str() != "" {
			t.network = "unix"
			t.address = p.Str()
			t.addressForError = t.address
			break
		}
		port := 0
		if p := o.Get("port"); p.IsNumber() {
			port = int(p.Float())
		} else if p.IsString() {
			n, err := strconv.Atoi(p.Str())
			if err != nil {
				return t, cb, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
					`The argument 'options.port' is invalid. Received type string ('`+p.Str()+`')`)
			}
			port = n
		}
		host := "localhost"
		if h := o.Get("host"); h.IsString() && h.Str() != "" {
			host = h.Str()
		}
		t.address = stdnet.JoinHostPort(host, strconv.Itoa(port))
		t.addressForError = host
		t.portForError = port
	default:
		return t, cb, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options" argument must be of type object. Received `+valueKind(first))
	}
	return t, cb, nil
}

func (m *module) hostSrvListen(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Server")
	}
	if s.listening {
		return js.Value{}, js.ThrowCode("Error", "ERR_SERVER_ALREADY_LISTEN", "Listen method has been called more than once without closing.")
	}
	t, cb, err := m.parseTarget(arg(args, 0), arg(args, 1), arg(args, 2))
	if err != nil {
		return js.Value{}, err
	}
	// listen(port[, host][, backlog][, cb]): the callback is the last function
	// argument regardless of where the backlog number landed.
	for _, a := range args[1:] {
		if a.Valid() && a.IsFunction() {
			cb = a
		}
	}

	ln, listenErr := clusterutil.Listen(t.network, t.address)
	if listenErr != nil {
		e := m.listenError(listenErr, t)
		s.obj.SetBool("listening", false)
		m.emitAsync(s.obj, "error", e.V())
		return this.V(), nil
	}
	s.ln = ln
	s.listening = true
	s.closed = false
	addr := ln.Addr()
	s.address, s.port = splitHostPort(addr.String())
	s.family = familyOf(addr)
	s.obj.SetBool("listening", true)
	m.addHandle()

	cbV := cb
	holder := m.e.JS.Obj()
	holder.SetFuncHidden("fire", func(this js.Object, _ []js.Value) (js.Value, error) {
		_ = m.emit(s.obj, "listening")
		if cbV.Valid() && cbV.IsFunction() {
			if _, err := cbV.Object().Call(s.obj); err != nil {
				return js.Value{}, err
			}
		}
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(holder.Get("fire"))
	m.startAccepting(s, ln)
	return this.V(), nil
}

func (m *module) listenError(err error, t target) js.Object {
	code := "EADDRINUSE"
	msg := "listen " + code + " " + t.address
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		code = "EACCES"
		msg = "listen " + code + " " + t.address
	case strings.Contains(err.Error(), "cannot assign requested address"):
		code = "EADDRNOTAVAIL"
		msg = "listen " + code + " " + t.address
	case strings.Contains(err.Error(), "address already in use"):
		code = "EADDRINUSE"
	}
	e := m.e.JS.Error(msg)
	e.SetStr("code", code)
	e.SetNum("errno", float64(errnoOf(code)))
	e.SetStr("syscall", "listen")
	e.SetStr("address", t.addressForError)
	e.SetNum("port", float64(t.portForError))
	return e
}

func (m *module) startAccepting(s *srv, ln stdnet.Listener) {
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				c := conn
				m.push(func() {
					if c != nil {
						_ = c.Close()
					}
					m.finishServerClose(s)
				})
				return
			}
			c := conn
			m.push(func() { m.accept(s, c) })
		}
	}()
}

func (m *module) accept(s *srv, conn stdnet.Conn) {
	if s.closed || !s.listening {
		_ = conn.Close()
		return
	}
	mk := m.makeSocket
	if !mk.Valid() || !mk.IsFunction() {
		_ = conn.Close()
		return
	}
	v, err := mk.Object().Call(js.Object{})
	if err != nil || !v.IsObject() {
		_ = conn.Close()
		return
	}
	obj := v.Object()
	so := m.sockOf(obj)
	if so == nil {
		_ = conn.Close()
		return
	}
	so.server = s
	s.connections[so.id] = so
	so.handleHeld = true
	m.addHandle()
	m.attach(so, conn)
	so.setProps()
	_ = m.emit(s.obj, "connection", obj.V())
}

func (m *module) hostSrvClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a net.Server")
	}
	cb := arg(args, 0)
	if !s.listening {
		err := js.ThrowCode("Error", "ERR_SERVER_NOT_RUNNING", "Server is not running.")
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb, m.e.JS.ErrorValue(err))
			return this.V(), nil
		}
		return js.Value{}, err
	}
	// Record the callback before closing so the accept loop's completion can
	// never observe a stale empty slot.
	s.closeCb = cb
	s.listening = false
	_ = s.ln.Close()
	return this.V(), nil
}

func (m *module) finishServerClose(s *srv) {
	if s.closed {
		return
	}
	s.closed = true
	s.obj.SetBool("listening", false)
	m.dropHandle()
	m.emitAsync(s.obj, "close")
	if s.closeCb.Valid() && s.closeCb.IsFunction() {
		m.e.NextTick(s.closeCb)
		s.closeCb = js.Value{}
	}
}

func (m *module) hostSrvAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil || !s.listening {
		return m.e.JS.Null(), nil
	}
	o := m.e.JS.Obj()
	addr := s.address
	family := "IPv4"
	if ip := stdnet.ParseIP(addr); ip != nil && ip.To4() == nil {
		family = "IPv6"
	} else if ip == nil && strings.Contains(addr, ":") {
		family = "IPv6"
	}
	o.SetStr("address", addr)
	o.SetNum("port", float64(s.port))
	o.SetStr("family", family)
	if strings.HasPrefix(addr, "/") || strings.HasPrefix(addr, ".") {
		return m.e.JS.Str(addr), nil
	}
	return o.V(), nil
}

func (m *module) hostSrvGetConnections(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	cb := arg(args, 0)
	n := 0
	if s != nil {
		n = len(s.connections)
	}
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Num(float64(n)))
	}
	return this.V(), nil
}

// hostParseIP returns {family, hex} for BlockList comparisons, or null.
func (m *module) hostParseIP(this js.Object, args []js.Value) (js.Value, error) {
	ip := stdnet.ParseIP(argStr(args, 0))
	if ip == nil {
		return m.e.JS.Null(), nil
	}
	o := m.e.JS.Obj()
	if v4 := ip.To4(); v4 != nil {
		o.SetStr("family", "ipv4")
		o.SetStr("hex", hex.EncodeToString(v4))
		return o.V(), nil
	}
	o.SetStr("family", "ipv6")
	o.SetStr("hex", hex.EncodeToString(ip.To16()))
	return o.V(), nil
}

const prelude = `(function (host, EventEmitter) {
  'use strict';

  function describe(v) {
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    if (typeof v === 'string') return "type string ('" + v + "')";
    if (typeof v === 'number') return 'type number (' + v + ')';
    if (typeof v === 'boolean') return 'type boolean (' + v + ')';
    if (Array.isArray(v)) return 'an instance of Array';
    if (typeof v === 'function') return 'function ' + (v.name || '(anonymous)');
    return 'an instance of ' + ((v.constructor && v.constructor.name) || 'Object');
  }
  function invalidArgType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function invalidArgValue(name, value, reason) {
    var e = new TypeError("The argument '" + name + "' " + (reason || 'is invalid') +
      '. Received ' + describe(value));
    e.code = 'ERR_INVALID_ARG_VALUE';
    return e;
  }

  class Socket extends EventEmitter {
    constructor(options) {
      super();
      host.initSock.call(this, options || {});
    }
    connect() {
      return host.sockConnect.call(this, arguments[0], arguments[1], arguments[2]);
    }
    write(chunk, encoding, cb) {
      return host.sockWrite.call(this, chunk, encoding, cb);
    }
    end(chunk, encoding, cb) {
      return host.sockEnd.call(this, chunk, encoding, cb);
    }
    destroy(err) {
      return host.sockDestroy.call(this, err);
    }
    resetAndDestroy() {
      return host.sockDestroy.call(this, undefined);
    }
    pause() {
      return host.sockPause.call(this, true);
    }
    resume() {
      return host.sockPause.call(this, false);
    }
    setTimeout(ms, cb) {
      return host.sockSetTimeout.call(this, ms, cb);
    }
    setNoDelay() { return this; }
    setKeepAlive() { return this; }
    setEncoding(enc) {
      return host.sockSetEncoding.call(this, enc);
    }
    address() {
      return host.sockAddress.call(this);
    }
    ref() {
      return host.sockRef.call(this, true);
    }
    unref() {
      return host.sockRef.call(this, false);
    }
    pipe(dest) {
      this.on('data', function (chunk) {
        if (dest && typeof dest.write === 'function') dest.write(chunk);
      });
      this.on('end', function () {
        if (dest && typeof dest.end === 'function') dest.end();
      });
      return dest;
    }
  }

  class Server extends EventEmitter {
    constructor(options, connectionListener) {
      super();
      if (typeof options === 'function') {
        connectionListener = options;
        options = {};
      }
      host.initServer.call(this, options || {});
      if (typeof connectionListener === 'function') {
        this.on('connection', connectionListener);
      }
    }
    listen() {
      return host.srvListen.call(this, arguments[0], arguments[1], arguments[2], arguments[3]);
    }
    close(cb) {
      return host.srvClose.call(this, cb);
    }
    address() {
      return host.srvAddress.call(this);
    }
    getConnections(cb) {
      return host.srvGetConnections.call(this, cb);
    }
    ref() { return this; }
    unref() { return this; }
  }

  function createServer(options, connectionListener) {
    if (typeof options === 'function') {
      connectionListener = options;
      options = {};
    }
    return new Server(options || {}, connectionListener);
  }

  function createConnection() {
    var args = Array.prototype.slice.call(arguments);
    var socket = new Socket();
    if (typeof args[args.length - 1] === 'function') {
      socket.once('connect', args.pop());
    }
    socket.connect.apply(socket, args);
    return socket;
  }

  function normalizeIP(address, type) {
    var parsed = host.parseIP(address);
    if (parsed === null) {
      throw invalidArgValue('address', address);
    }
    if (type !== undefined && type !== null) {
      if (type !== 'ipv4' && type !== 'ipv6') {
        throw invalidArgValue('type', type);
      }
      if (parsed.family !== type) {
        throw invalidArgValue('address', address);
      }
    }
    return { family: parsed.family, value: BigInt('0x' + parsed.hex) };
  }

  const BITS = { ipv4: 32n, ipv6: 128n };

  class BlockList {
    constructor() {
      Object.defineProperty(this, 'rules', {
        value: [], writable: false, enumerable: true, configurable: true
      });
    }
    addAddress(address, type) {
      return this.addRange(address, address, type);
    }
    addRange(start, end, type) {
      var a = normalizeIP(start, type);
      var b = normalizeIP(end, a.family);
      if (a.value > b.value) {
        throw invalidArgValue('start', start, 'must come before end');
      }
      this.rules.push({
        type: a.family, start: a.value, end: b.value,
        toString: function () {
          return start + ':' + a.family + '=>' + end;
        }
      });
      return this;
    }
    addSubnet(net, prefix, type) {
      var a = normalizeIP(net, type);
      var bits = BITS[a.family];
      if (typeof prefix !== 'number' || prefix < 0 || BigInt(prefix) > bits) {
        throw invalidArgValue('prefix', prefix);
      }
      var hostBits = bits - BigInt(prefix);
      var mask = hostBits === 0n ? ~0n : ((1n << bits) - 1n) & ~((1n << hostBits) - 1n);
      var start = a.value & mask;
      var end = start | (~mask & ((1n << bits) - 1n));
      this.rules.push({
        type: a.family, start: start, end: end,
        toString: function () { return net + '/' + prefix + ':' + a.family; }
      });
      return this;
    }
    check(address, type) {
      var a;
      try {
        a = normalizeIP(address, type);
      } catch (err) {
        return false;
      }
      for (var i = 0; i < this.rules.length; i++) {
        var rule = this.rules[i];
        if (rule.type === a.family && a.value >= rule.start && a.value <= rule.end) {
          return true;
        }
      }
      return false;
    }
    [Symbol.for('nodejs.util.inspect.custom')]() {
      return this.rules;
    }
  }

  class SocketAddress {
    constructor(options) {
      if (options === null || typeof options !== 'object') {
        throw invalidArgType('options', 'object', options);
      }
      var family = options.family === undefined ? 'ipv4' : options.family;
      if (family !== 'ipv4' && family !== 'ipv6') {
        throw invalidArgValue('options.family', family);
      }
      this.address = options.address === undefined ? '127.0.0.1' : String(options.address);
      this.family = family;
      this.port = options.port === undefined ? 0 : Number(options.port);
      if (family === 'ipv6') {
        this.flowlabel = options.flowlabel === undefined ? 0 : Number(options.flowlabel);
      }
    }
    toJSON() {
      var out = {
        address: this.address, family: this.family, port: this.port
      };
      if (this.family === 'ipv6') out.flowlabel = this.flowlabel;
      return out;
    }
    toString() {
      return this.address + ':' + this.port;
    }
  }

  var autoSelectFamily = true;
  var autoSelectFamilyAttemptTimeout = 250;

  return {
    Socket: Socket,
    Server: Server,
    BlockList: BlockList,
    SocketAddress: SocketAddress,
    createServer: createServer,
    createConnection: createConnection,
    connect: createConnection,
    makeSocket: function () { return new Socket(); },
    makeServer: function (options, listener) { return new Server(options, listener); },
    getDefaultAutoSelectFamily: function () { return autoSelectFamily; },
    setDefaultAutoSelectFamily: function (value) {
      autoSelectFamily = !!value;
    },
    getDefaultAutoSelectFamilyAttemptTimeout: function () {
      return autoSelectFamilyAttemptTimeout;
    },
    setDefaultAutoSelectFamilyAttemptTimeout: function (value) {
      autoSelectFamilyAttemptTimeout = Number(value);
    }
  };
})`
