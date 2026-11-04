// Package http implements Node's HTTP server API on top of Go's HTTP/1.1
// server. Go owns connection and protocol handling; the JavaScript prelude
// owns Node-style Server, IncomingMessage and ServerResponse objects.
package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	stdnet "net"
	stdhttp "net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"godejs/internal/clusterutil"
	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("http", Install) }

// pumpEveryMS only keeps the event loop alive while a server or client request
// is active. It is no longer a polling interval: requests wake the loop
// directly, so dispatch no longer waits for the next tick.
const pumpEveryMS = 1000

type module struct {
	e *env.Env

	mu        sync.Mutex
	idMu      sync.Mutex
	pumpID    int
	active    int
	servers   map[float64]*server
	responses map[float64]*responseBridge
	seq       float64

	exports          js.Object
	host             js.Object
	dispatchFn       js.Value
	clientDispatchFn js.Value
	bufferFrom       js.Object
	bufferCtor       js.Object
	emptyBuffer      js.Value
}

type server struct {
	id  float64
	obj js.Object

	hs        *stdhttp.Server
	ln        stdnet.Listener
	listening bool
	closed    bool
	closeSent bool
	closeCb   js.Value
	handle    bool

	address   string
	port      int
	family    string
	tlsConfig *tls.Config
	http3     bool
	packet    stdnet.PacketConn
	h3        any

	headersTimeout   time.Duration
	keepAliveTimeout time.Duration
	requestTimeout   time.Duration

	connMu sync.Mutex
	conns  map[stdnet.Conn]bool
}

type responseBridge struct {
	id float64
	ch chan responseCommand
	er chan error
}

type responseCommand struct {
	kind    string
	status  int
	headers []headerPair
	body    []byte
}

type headerPair struct {
	name  string
	value string
}

type clientResult struct {
	statusCode    int
	statusMessage string
	version       string
	headers       []headerPair
	body          []byte
	method        string
	url           string
}

type target struct {
	network         string
	address         string
	addressForError string
	portForError    int
}

func Install(e *env.Env) js.Object {
	m := &module{
		e:         e,
		servers:   map[float64]*server{},
		responses: map[float64]*responseBridge{},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()
	activeHTTP.Store(e, m)

	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("initServer", m.hostInitServer)
	m.host.SetFuncHidden("srvListen", m.hostSrvListen)
	m.host.SetFuncHidden("srvClose", m.hostSrvClose)
	m.host.SetFuncHidden("srvAddress", m.hostSrvAddress)
	m.host.SetFuncHidden("srvGetConnections", m.hostSrvGetConnections)
	m.host.SetFuncHidden("respHead", m.hostRespHead)
	m.host.SetFuncHidden("respWrite", m.hostRespWrite)
	m.host.SetFuncHidden("respWriteAll", m.hostRespWriteAll)
	m.host.SetFuncHidden("respEnd", m.hostRespEnd)
	m.host.SetFuncHidden("respEndAll", m.hostRespEndAll)
	m.host.SetFuncHidden("respAbort", m.hostRespAbort)
	m.host.SetFuncHidden("clientRequest", m.hostClientRequest)

	e.Exports("buffer")
	buffer := e.Exports("buffer").Get("Buffer")
	if buffer.Valid() && buffer.IsFunction() {
		m.bufferCtor = buffer.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}
	// Most HTTP requests have no body. Reuse one empty Buffer instead of
	// paying a Buffer.from host callback for every request.
	if empty := m.buffer(nil); empty.Valid() {
		empty.Protect()
		m.emptyBuffer = empty
	}

	events := e.Exports("events")
	ee := events.Get("EventEmitter")
	if !ee.Valid() || !ee.IsFunction() {
		panic("godejs http: events.EventEmitter is unavailable")
	}
	api, err := e.JS.Eval(prelude, "http.js")
	if err != nil {
		panic("godejs http: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V(), ee)
	if err != nil {
		panic("godejs http: " + err.Error())
	}
	obj := built.Object()
	m.dispatchFn = obj.Get("dispatch").Keep()
	m.clientDispatchFn = obj.Get("clientDispatch").Keep()
	for _, name := range obj.OwnKeys() {
		if name == "dispatch" || name == "clientDispatch" {
			continue
		}
		m.exports.Set(name, obj.Get(name))
	}

	m.exports.SetFunc("validateHeaderName", func(this js.Object, args []js.Value) (js.Value, error) {
		name := argStr(args, 0)
		if !validHeaderName(name) {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_HTTP_TOKEN",
				fmt.Sprintf("Header name must be a valid HTTP token [\"%s\"]", name))
		}
		return e.JS.Undefined(), nil
	})
	m.exports.SetFunc("validateHeaderValue", func(this js.Object, args []js.Value) (js.Value, error) {
		name, value := argStr(args, 0), argStr(args, 1)
		if strings.ContainsAny(value, "\r\n") {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_CHAR",
				fmt.Sprintf("Invalid character in header content [\"%s\"]", name))
		}
		return e.JS.Undefined(), nil
	})

	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

func (m *module) nextID() float64 {
	m.idMu.Lock()
	m.seq++
	id := m.seq
	m.idMu.Unlock()
	return id
}

// push runs fn on the thread that owns the JavaScript context. Running it
// through the loop's native queue means a request is dispatched as soon as the
// loop is free, instead of waiting for a timer to poll it.
func (m *module) push(fn func()) { m.e.EnqueueNative(fn) }

// pump is only a keepalive now: it keeps a refed timer in the loop so the
// process stays alive while a server is listening.
func (m *module) pump(this js.Object, args []js.Value) (js.Value, error) {
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
		s.listening = false
		s.closed = true
		if s.h3 != nil {
			m.closeHTTP3(s)
		} else if s.hs != nil {
			_ = s.hs.Close()
		}
		if s.ln != nil {
			_ = s.ln.Close()
		}
	}
}

func (m *module) srvOf(this js.Object) *server {
	v := this.Get("$godeHttpId")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	s := m.servers[v.Float()]
	m.mu.Unlock()
	return s
}

func (m *module) buffer(b []byte) js.Value {
	if len(b) == 0 && m.emptyBuffer.Valid() {
		return m.emptyBuffer
	}
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

func durationOption(options js.Object, name string, fallback time.Duration) time.Duration {
	v := options.Get(name)
	if !v.IsNumber() || v.Float() < 0 {
		return fallback
	}
	return time.Duration(v.Float()) * time.Millisecond
}

func (m *module) hostInitServer(this js.Object, args []js.Value) (js.Value, error) {
	id := m.nextID()
	s := &server{
		id:               id,
		obj:              this,
		conns:            map[stdnet.Conn]bool{},
		headersTimeout:   60 * time.Second,
		keepAliveTimeout: 5 * time.Second,
		requestTimeout:   300 * time.Second,
	}
	options := arg(args, 0)
	if options.Valid() && options.IsObject() {
		o := options.Object()
		s.headersTimeout = durationOption(o, "headersTimeout", s.headersTimeout)
		s.keepAliveTimeout = durationOption(o, "keepAliveTimeout", s.keepAliveTimeout)
		s.requestTimeout = durationOption(o, "requestTimeout", s.requestTimeout)
		s.http3 = o.Get("_http3").Bool()
		if o.Get("_tls").Bool() || s.http3 {
			cfg, err := tlsServerConfig(o)
			if err != nil {
				return js.Value{}, err
			}
			s.tlsConfig = cfg
		}
	}
	m.mu.Lock()
	m.servers[id] = s
	m.mu.Unlock()
	this.SetHidden("$godeHttpId", m.e.JS.Num(id))
	this.Protect()
	this.SetBool("listening", false)
	return this.V(), nil
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

func (m *module) parseTarget(args []js.Value) (target, js.Value, error) {
	var t target
	t.network = "tcp"
	var cb js.Value
	first := arg(args, 0)
	second := arg(args, 1)
	third := arg(args, 2)
	for _, a := range args {
		if a.Valid() && a.IsFunction() {
			cb = a
		}
	}

	switch {
	case first.Valid() && first.IsNumber():
		port := int(first.Float())
		host := "localhost"
		if second.Valid() && second.IsString() && second.Str() != "" {
			host = second.Str()
		}
		t.address = stdnet.JoinHostPort(host, strconv.Itoa(port))
		t.addressForError = host
		t.portForError = port
	case first.Valid() && first.IsString():
		t.network = "unix"
		t.address = first.Str()
		t.addressForError = t.address
	case first.Valid() && first.IsObject():
		o := first.Object()
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
					fmt.Sprintf("The argument 'options.port' is invalid. Received type string ('%s')", p.Str()))
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
			"The \"options\" argument must be of type object. Received "+describe(first))
	}
	_ = second
	_ = third
	return t, cb, nil
}

func (m *module) hostSrvListen(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not an http.Server")
	}
	if s.listening {
		return js.Value{}, js.ThrowCode("Error", "ERR_SERVER_ALREADY_LISTEN",
			"Listen method has been called more than once without closing.")
	}
	values := args
	if len(args) == 1 && args[0].IsArray() {
		arr := args[0].Object()
		values = make([]js.Value, arr.Length())
		for i := range values {
			values[i] = arr.At(i)
		}
	}
	t, cb, err := m.parseTarget(values)
	if err != nil {
		return js.Value{}, err
	}
	if s.http3 {
		return m.listenHTTP3(s, t, cb)
	}

	ln, listenErr := clusterutil.Listen(t.network, t.address)
	if listenErr != nil {
		s.obj.SetBool("listening", false)
		m.emitAsync(s.obj, "error", m.listenError(listenErr, t).V())
		return this.V(), nil
	}
	s.ln = ln
	s.listening = true
	s.closed = false
	s.closeSent = false
	s.address, s.port = splitHostPort(ln.Addr().String())
	s.family = familyOf(ln.Addr())
	s.obj.SetBool("listening", true)
	s.handle = true
	s.hs = &stdhttp.Server{
		Handler: stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			m.handle(s, w, r)
		}),
		ConnState:         s.track,
		ReadHeaderTimeout: s.headersTimeout,
		IdleTimeout:       s.keepAliveTimeout,
		ReadTimeout:       s.requestTimeout,
	}
	m.addHandle()

	serveLn := ln
	if s.tlsConfig != nil {
		serveLn = tls.NewListener(ln, s.tlsConfig)
	}
	go func() {
		serveErr := s.hs.Serve(serveLn)
		m.push(func() {
			if serveErr != nil && serveErr != stdhttp.ErrServerClosed && !s.closed {
				_ = m.emit(s.obj, "error", m.serverError(serveErr).V())
			}
			m.finishServerClose(s)
		})
	}()

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
	return this.V(), nil
}

func (s *server) track(c stdnet.Conn, state stdhttp.ConnState) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	switch state {
	case stdhttp.StateNew:
		s.conns[c] = true
	case stdhttp.StateClosed, stdhttp.StateHijacked:
		delete(s.conns, c)
	}
}

func (m *module) serverError(err error) js.Object {
	e := m.e.JS.Error(err.Error())
	e.SetStr("code", "EHTTP")
	return e
}

func (m *module) listenError(err error, t target) js.Object {
	code := "EADDRINUSE"
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		code = "EACCES"
	case strings.Contains(err.Error(), "cannot assign requested address"):
		code = "EADDRNOTAVAIL"
	case strings.Contains(err.Error(), "address already in use"):
		code = "EADDRINUSE"
	}
	e := m.e.JS.Error("listen " + code + " " + t.address)
	e.SetStr("code", code)
	e.SetStr("syscall", "listen")
	e.SetStr("address", t.addressForError)
	e.SetNum("port", float64(t.portForError))
	return e
}

func (m *module) hostSrvClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not an http.Server")
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
	s.closeCb = cb
	s.listening = false
	s.obj.SetBool("listening", false)
	if s.h3 != nil {
		m.closeHTTP3(s)
	} else if s.hs != nil {
		_ = s.hs.Close()
	} else if s.ln != nil {
		_ = s.ln.Close()
		m.finishServerClose(s)
	}
	return this.V(), nil
}

func (m *module) finishServerClose(s *server) {
	if s.closeSent {
		return
	}
	s.closeSent = true
	s.closed = true
	s.listening = false
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

func (m *module) hostSrvAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil || !s.listening {
		return m.e.JS.Null(), nil
	}
	if s.ln != nil && s.ln.Addr().Network() == "unix" {
		return m.e.JS.Str(s.address), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("address", s.address)
	o.SetNum("port", float64(s.port))
	o.SetStr("family", s.family)
	return o.V(), nil
}

func (m *module) hostSrvGetConnections(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	cb := arg(args, 0)
	n := 0
	if s != nil {
		s.connMu.Lock()
		n = len(s.conns)
		s.connMu.Unlock()
	}
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Num(float64(n)))
	}
	return this.V(), nil
}

// handle runs one Go HTTP request and bridges it to the JavaScript thread.
func (m *module) handle(s *server, w stdhttp.ResponseWriter, r *stdhttp.Request) {
	// Most requests (notably GET and HEAD) have no body. Reading an empty
	// body still allocates and synchronously polls the reader, so only touch
	// it when the request declares a body or uses chunked framing.
	var body []byte
	if r.ContentLength != 0 {
		var readErr error
		body, readErr = io.ReadAll(r.Body)
		if readErr != nil {
			httpError(w, 400, "Bad Request")
			return
		}
	}
	b := &responseBridge{
		id: m.nextID(),
		ch: make(chan responseCommand, 16),
		er: make(chan error, 1),
	}
	m.mu.Lock()
	m.responses[b.id] = b
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.responses, b.id)
		m.mu.Unlock()
	}()

	// Addresses travel to JavaScript as raw host:port strings. Node's socket
	// object parses them lazily, and most handlers never look, so parsing and
	// re-joining them on every request was wasted work.
	localAddr := ""
	if a, ok := r.Context().Value(stdhttp.LocalAddrContextKey).(stdnet.Addr); ok {
		localAddr = a.String()
	}

	m.push(func() {
		if s.closed {
			select {
			case b.er <- fmt.Errorf("server closed"):
			default:
			}
			return
		}
		if err := m.dispatchJS(s, b, r, body, r.RemoteAddr, localAddr); err != nil {
			select {
			case b.er <- err:
			default:
			}
		}
	})

	headersWritten := false
	for {
		select {
		case cmd := <-b.ch:
			switch cmd.kind {
			case "head":
				for _, h := range cmd.headers {
					w.Header().Add(h.name, h.value)
				}
				w.WriteHeader(cmd.status)
				headersWritten = true
			case "write":
				if !headersWritten {
					headersWritten = true
				}
				_, _ = w.Write(cmd.body)
				if f, ok := w.(stdhttp.Flusher); ok {
					f.Flush()
				}
			case "writefull":
				for _, h := range cmd.headers {
					w.Header().Add(h.name, h.value)
				}
				w.WriteHeader(cmd.status)
				headersWritten = true
				_, _ = w.Write(cmd.body)
				if f, ok := w.(stdhttp.Flusher); ok {
					f.Flush()
				}
			case "full":
				for _, h := range cmd.headers {
					w.Header().Add(h.name, h.value)
				}
				w.WriteHeader(cmd.status)
				headersWritten = true
				if len(cmd.body) > 0 {
					_, _ = w.Write(cmd.body)
				}
				return
			case "end":
				if !headersWritten {
					headersWritten = true
				}
				if len(cmd.body) > 0 {
					_, _ = w.Write(cmd.body)
				}
				return
			case "abort":
				return
			}
		case err := <-b.er:
			if !headersWritten {
				httpError(w, 500, "Internal Server Error")
			}
			fmt.Fprintf(m.e.Err(), "godejs http: request handler failed: %v\n", err)
			return
		case <-r.Context().Done():
			return
		}
	}
}

func (m *module) dispatchJS(s *server, b *responseBridge, r *stdhttp.Request, body []byte, remoteAddr, localAddr string) error {
	// Method/URL/version and every header pair travel as two packed strings.
	// Building a JS array and calling SetAt once per header crossed the cgo
	// bridge dozens of times per request, which dominated small responses.
	var head strings.Builder
	head.Grow(len(r.Method) + len(r.RequestURI) + len(r.Proto) + len(remoteAddr) + len(localAddr) + 5)
	head.WriteString(r.Method)
	head.WriteByte(0x1f)
	head.WriteString(r.RequestURI)
	head.WriteByte(0x1f)
	head.WriteString(r.Proto)
	head.WriteByte(0x1f)
	head.WriteString(remoteAddr)
	head.WriteByte(0x1f)
	head.WriteString(localAddr)

	var packed strings.Builder
	if r.Host != "" {
		packed.WriteByte(0x1e)
		packed.WriteString("Host")
		packed.WriteByte(0x1f)
		packed.WriteString(r.Host)
	}
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range r.Header.Values(key) {
			packed.WriteByte(0x1e)
			packed.WriteString(key)
			packed.WriteByte(0x1f)
			packed.WriteString(value)
		}
	}
	if r.ContentLength >= 0 && r.Header.Get("Content-Length") == "" {
		packed.WriteByte(0x1e)
		packed.WriteString("Content-Length")
		packed.WriteByte(0x1f)
		packed.WriteString(strconv.FormatInt(r.ContentLength, 10))
	}
	if len(r.TransferEncoding) > 0 && r.Header.Get("Transfer-Encoding") == "" {
		packed.WriteByte(0x1e)
		packed.WriteString("Transfer-Encoding")
		packed.WriteByte(0x1f)
		packed.WriteString(strings.Join(r.TransferEncoding, ", "))
	}

	result, err := m.dispatchFn.Object().Call(js.Object{},
		s.obj.V(),
		m.e.JS.Num(b.id),
		m.e.JS.Str(head.String()),
		m.e.JS.Str(packed.String()),
		m.buffer(body),
	)
	if err != nil {
		return err
	}
	// Decoding the result allocates JS values (notably toBytes), which can
	// trigger a JavaScriptCore collection. Keep the object graph alive until
	// every property has been read.
	result.Protect()
	defer result.Unprotect()
	// A synchronous handler returns its complete response instead of making a
	// second cgo round trip from res.end(). This is the common case for small
	// JSON/text endpoints and removes one host call per request.
	// dispatch() returns a three-element array (status, packed headers, body).
	// Reading by index avoids allocating a property-name string per field, and
	// a string body skips the Buffer round trip for text responses entirely.
	if result.Valid() && result.IsArray() {
		arr := result.Object()
		statusValue := arr.At(0)
		if statusValue.IsNumber() {
			status := statusValue.Int()
			bodyValue := arr.At(2)
			var data []byte
			if bodyValue.IsString() {
				data = []byte(bodyValue.Str())
			} else {
				b, bodyErr := m.toBytes(bodyValue, "")
				if bodyErr != nil {
					return bodyErr
				}
				data = b
			}
			headersValue := arr.At(1)
			var headers []headerPair
			if headersValue.IsString() {
				headers = parsePackedHeaders(headersValue.Str())
			} else {
				headers = m.headersFromJS(headersValue)
			}
			select {
			case b.ch <- responseCommand{kind: "full", status: status, headers: headers, body: data}:
			default:
				return fmt.Errorf("response queue full")
			}
		}
	}
	return nil
}

func parsePackedHeaders(packed string) []headerPair {
	if packed == "" {
		return nil
	}
	records := strings.Split(packed, "\x1e")
	out := make([]headerPair, 0, len(records))
	for _, record := range records {
		if record == "" {
			continue
		}
		at := strings.IndexByte(record, 0x1f)
		if at < 0 {
			continue
		}
		out = append(out, headerPair{name: record[:at], value: record[at+1:]})
	}
	return out
}

func (m *module) hostClientRequest(this js.Object, args []js.Value) (js.Value, error) {
	method := strings.ToUpper(argStr(args, 0))
	if method == "" {
		method = "GET"
	}
	target := argStr(args, 1)
	headers := m.headersFromJS(arg(args, 2))
	body, err := m.toBytes(arg(args, 3), "")
	if err != nil {
		return js.Value{}, err
	}
	req, err := stdhttp.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return js.Value{}, err
	}
	for _, h := range headers {
		if strings.EqualFold(h.name, "Host") {
			req.Host = h.value
			continue
		}
		req.Header.Add(h.name, h.value)
	}
	requestOptions := arg(args, 4)
	transport := &stdhttp.Transport{DisableCompression: true}
	if req.URL.Scheme == "https" {
		tlsCfg, tlsErr := tlsClientConfig(requestOptions, req.URL.Hostname())
		if tlsErr != nil {
			return js.Value{}, tlsErr
		}
		transport.TLSClientConfig = tlsCfg
	}
	this.Protect()
	m.addHandle()
	go func() {
		client := &stdhttp.Client{
			Transport: transport,
			CheckRedirect: func(_ *stdhttp.Request, _ []*stdhttp.Request) error {
				return stdhttp.ErrUseLastResponse
			},
		}
		resp, requestErr := client.Do(req)
		if requestErr != nil {
			m.push(func() {
				defer m.dropHandle()
				_ = m.emit(this, "error", m.e.JS.Error(requestErr.Error()).V())
			})
			return
		}
		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			m.push(func() {
				defer m.dropHandle()
				_ = m.emit(this, "error", m.e.JS.Error(readErr.Error()).V())
			})
			return
		}
		result := clientResult{
			statusCode:    resp.StatusCode,
			statusMessage: resp.Status,
			version:       resp.Proto,
			body:          data,
			method:        req.Method,
			url:           req.URL.String(),
		}
		if i := strings.IndexByte(result.statusMessage, ' '); i >= 0 {
			result.statusMessage = result.statusMessage[i+1:]
		}
		keys := make([]string, 0, len(resp.Header))
		for k := range resp.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			for _, v := range resp.Header.Values(k) {
				result.headers = append(result.headers, headerPair{name: k, value: v})
			}
		}
		if resp.ContentLength >= 0 && resp.Header.Get("Content-Length") == "" {
			result.headers = append(result.headers, headerPair{name: "Content-Length", value: strconv.FormatInt(resp.ContentLength, 10)})
		}
		m.push(func() {
			defer m.dropHandle()
			m.clientDispatch(this, result)
		})
	}()
	return this.V(), nil
}

func (m *module) clientDispatch(reqObj js.Object, result clientResult) {
	raw := m.e.JS.Arr(nil)
	n := 0
	for _, h := range result.headers {
		raw.SetAt(n, m.e.JS.Str(h.name))
		raw.SetAt(n+1, m.e.JS.Str(h.value))
		n += 2
	}
	_, _ = m.clientDispatchFn.Object().Call(js.Object{},
		reqObj.V(),
		m.e.JS.Num(float64(result.statusCode)),
		m.e.JS.Str(result.statusMessage),
		m.e.JS.Str(result.version),
		raw.V(),
		m.buffer(result.body),
		m.e.JS.Str(result.method),
		m.e.JS.Str(result.url),
	)
}

func (m *module) responseOf(this js.Object) *responseBridge {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	b := m.responses[v.Float()]
	m.mu.Unlock()
	return b
}

func (m *module) hostRespHead(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot write after the stream was destroyed")
	}
	status := 200
	if v := arg(args, 0); v.IsNumber() {
		status = int(v.Float())
	}
	headers := m.headersFromJS(arg(args, 2))
	select {
	case b.ch <- responseCommand{kind: "head", status: status, headers: headers}:
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	return this.V(), nil
}

func (m *module) hostRespWrite(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot write after the stream was destroyed")
	}
	data, err := m.toBytes(arg(args, 0), "")
	if err != nil {
		return js.Value{}, err
	}
	select {
	case b.ch <- responseCommand{kind: "write", body: data}:
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	return this.V(), nil
}

func (m *module) hostRespWriteAll(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot write after the stream was destroyed")
	}
	status := 200
	if v := arg(args, 0); v.IsNumber() {
		status = int(v.Float())
	}
	data, err := m.toBytes(arg(args, 2), "")
	if err != nil {
		return js.Value{}, err
	}
	select {
	case b.ch <- responseCommand{kind: "writefull", status: status, headers: m.headersFromJS(arg(args, 1)), body: data}:
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	return this.V(), nil
}

func (m *module) hostRespEndAll(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot end after the stream was destroyed")
	}
	status := 200
	if v := arg(args, 0); v.IsNumber() {
		status = int(v.Float())
	}
	data, err := m.toBytes(arg(args, 2), "")
	if err != nil {
		return js.Value{}, err
	}
	select {
	case b.ch <- responseCommand{kind: "full", status: status, headers: m.headersFromJS(arg(args, 1)), body: data}:
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	return this.V(), nil
}

func (m *module) hostRespEnd(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_DESTROYED", "Cannot end after the stream was destroyed")
	}
	data, err := m.toBytes(arg(args, 0), "")
	if err != nil {
		return js.Value{}, err
	}
	select {
	case b.ch <- responseCommand{kind: "end", body: data}:
	default:
		return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
	}
	return this.V(), nil
}

func (m *module) hostRespAbort(this js.Object, args []js.Value) (js.Value, error) {
	b := m.responseOf(this)
	if b == nil {
		return this.V(), nil
	}
	select {
	case b.ch <- responseCommand{kind: "abort"}:
	default:
	}
	return this.V(), nil
}

func (m *module) headersFromJS(v js.Value) []headerPair {
	if !v.Valid() || !v.IsObject() {
		return nil
	}
	obj := v.Object()
	var out []headerPair
	for _, name := range obj.OwnKeys() {
		value := obj.Get(name)
		if value.IsArray() {
			arr := value.Object()
			for i := 0; i < arr.Length(); i++ {
				out = append(out, headerPair{name: name, value: stringify(arr.At(i))})
			}
		} else {
			out = append(out, headerPair{name: name, value: stringify(value)})
		}
	}
	return out
}

func stringify(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return v.Str()
	case v.IsNumber():
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case v.IsBool():
		return strconv.FormatBool(v.Bool())
	}
	return v.Str()
}

func (m *module) toBytes(v js.Value, encoding string) ([]byte, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil, nil
	}
	if v.IsString() {
		return encodeString(v.Str(), encoding), nil
	}
	if b, ok := v.Bytes(); ok {
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "chunk" argument must be of type string or an instance of Buffer or Uint8Array. Received `+describe(v))
}

func describe(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsBool():
		return "type boolean"
	case v.IsArray():
		return "an instance of Array"
	}
	return "an instance of Object"
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
		out, _ := hex.DecodeString(strings.TrimSpace(s))
		return out
	case "base64":
		out, _ := base64.StdEncoding.DecodeString(s)
		return out
	case "base64url":
		out, _ := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
		return out
	}
	return []byte(s)
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

func httpError(w stdhttp.ResponseWriter, status int, message string) {
	stdhttp.Error(w, message, status)
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

func tlsVersion(name string, fallback uint16) uint16 {
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

func tlsOptionsBase(o js.Object, server bool) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS13}
	if v := o.Get("minVersion"); v.IsString() {
		cfg.MinVersion = tlsVersion(v.Str(), cfg.MinVersion)
	}
	if v := o.Get("maxVersion"); v.IsString() {
		cfg.MaxVersion = tlsVersion(v.Str(), cfg.MaxVersion)
	}
	if v := o.Get("ciphers"); v.IsString() && v.Str() != "" {
		var ids []uint16
		for _, wanted := range strings.Split(v.Str(), ":") {
			for _, suite := range tls.CipherSuites() {
				if strings.EqualFold(suite.Name, strings.TrimSpace(wanted)) {
					ids = append(ids, suite.ID)
				}
			}
		}
		if len(ids) > 0 {
			cfg.CipherSuites = ids
		}
	}
	if v := o.Get("ca"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(optionBytes(o, "ca")) {
			return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CA", "Unable to load CA certificates")
		}
		if server {
			cfg.ClientCAs = pool
		} else {
			cfg.RootCAs = pool
		}
	}
	if v := o.Get("requestCert"); v.IsBool() && v.Bool() {
		cfg.ClientAuth = tls.RequestClientCert
	}
	if v := o.Get("rejectUnauthorized"); v.IsBool() && !server {
		cfg.InsecureSkipVerify = !v.Bool()
	} else if !server {
		cfg.InsecureSkipVerify = false
	}
	if v := o.Get("servername"); v.IsString() && v.Str() != "" {
		cfg.ServerName = v.Str()
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

func tlsServerConfig(o js.Object) (*tls.Config, error) {
	cfg, err := tlsOptionsBase(o, true)
	if err != nil {
		return nil, err
	}
	certPEM, keyPEM := optionBytes(o, "cert"), optionBytes(o, "key")
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, js.ThrowCode("Error", "ERR_TLS_CERT_ALTNAME_INVALID",
			"An https server requires both 'cert' and 'key' options")
	}
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT", err.Error())
	}
	cfg.Certificates = []tls.Certificate{pair}
	return cfg, nil
}

func tlsClientConfig(o js.Value, hostname string) (*tls.Config, error) {
	if !o.Valid() || !o.IsObject() {
		return &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS13, ServerName: hostname}, nil
	}
	obj := o.Object()
	cfg, err := tlsOptionsBase(obj, false)
	if err != nil {
		return nil, err
	}
	if cfg.ServerName == "" {
		cfg.ServerName = hostname
	}
	certPEM, keyPEM := optionBytes(obj, "cert"), optionBytes(obj, "key")
	if len(certPEM) > 0 && len(keyPEM) > 0 {
		pair, pairErr := tls.X509KeyPair(certPEM, keyPEM)
		if pairErr != nil {
			return nil, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT", pairErr.Error())
		}
		cfg.Certificates = []tls.Certificate{pair}
	}
	return cfg, nil
}

func validHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		switch c {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		}
		return false
	}
	return true
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"http.Agent",
		"http.IncomingMessage",
		"http.METHODS",
		"http.Server",
		"http.Server#address",
		"http.Server#close",
		"http.Server#getConnections",
		"http.Server#listen",
		"http.Server#ref",
		"http.Server#unref",
		"http.ServerResponse",
		"http.ServerResponse#end",
		"http.ServerResponse#getHeader",
		"http.ServerResponse#getHeaderNames",
		"http.ServerResponse#getHeaders",
		"http.ServerResponse#hasHeader",
		"http.ServerResponse#removeHeader",
		"http.ServerResponse#setHeader",
		"http.ServerResponse#write",
		"http.ServerResponse#writeHead",
		"http.STATUS_CODES",
		"http.createServer",
		"http.get",
		"http.globalAgent",
		"http.request",
		"http.maxHeaderSize",
		"http.validateHeaderName",
		"http.validateHeaderValue",
	}
}
