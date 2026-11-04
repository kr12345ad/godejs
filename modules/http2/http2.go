// Package http2 implements Node's HTTP/2 API on golang.org/x/net/http2.
// Go owns framing and TLS; the prelude owns Node-style sessions, streams and
// compatibility request/response objects.
package http2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	stdnet "net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"

	"godejs/internal/clusterutil"
	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("http2", Install) }

const pumpEveryMS = 1

type module struct {
	e      *env.Env
	mu     sync.Mutex
	idMu   sync.Mutex
	seq    float64
	queued []func()
	pumpID int
	active int

	servers  map[float64]*server
	bridges  map[float64]*bridge
	sessions map[float64]*session
	streams  map[float64]*clientStream

	exports    js.Object
	host       js.Object
	dispatchFn js.Value
	bufferCtor js.Object
	bufferFrom js.Object
}

type server struct {
	id        float64
	obj       js.Object
	secure    bool
	ln        stdnet.Listener
	h2        *http2.Server
	listening bool
	closed    bool
	address   string
	port      int
	connMu    sync.Mutex
	conns     map[stdnet.Conn]bool
}

type bridge struct {
	id float64
	ch chan command
	er chan error
	rw http.ResponseWriter
}

type command struct {
	kind     string
	status   int
	headers  []headerPair
	trailers []headerPair
	body     []byte
}

type headerPair struct{ name, value string }

type session struct {
	id        float64
	obj       js.Object
	authority *url.URL
	transport *http2.Transport
	closed    bool
}

type clientStream struct {
	id      float64
	obj     js.Object
	sess    *session
	headers js.Value
	body    []byte
	ended   bool
	started bool
	bodyCh  chan []byte
}

func Install(e *env.Env) js.Object {
	m := &module{
		e:        e,
		servers:  map[float64]*server{},
		bridges:  map[float64]*bridge{},
		sessions: map[float64]*session{},
		streams:  map[float64]*clientStream{},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()
	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("srvInit", m.hostSrvInit)
	m.host.SetFuncHidden("srvListen", m.hostSrvListen)
	m.host.SetFuncHidden("srvClose", m.hostSrvClose)
	m.host.SetFuncHidden("srvAddress", m.hostSrvAddress)
	m.host.SetFuncHidden("respHead", m.hostRespHead)
	m.host.SetFuncHidden("respWrite", m.hostRespWrite)
	m.host.SetFuncHidden("respEnd", m.hostRespEnd)
	m.host.SetFuncHidden("respAbort", m.hostRespAbort)
	m.host.SetFuncHidden("respTrailers", m.hostRespTrailers)
	m.host.SetFuncHidden("respPush", m.hostRespPush)
	m.host.SetFuncHidden("respFile", m.hostRespFile)
	m.host.SetFuncHidden("respFD", m.hostRespFD)
	m.host.SetFuncHidden("connect", m.hostConnect)
	m.host.SetFuncHidden("sessionSettings", m.hostSessionSettings)
	m.host.SetFuncHidden("streamWrite", m.hostStreamWrite)
	m.host.SetFuncHidden("request", m.hostRequest)
	m.host.SetFuncHidden("streamEnd", m.hostStreamEnd)
	m.host.SetFuncHidden("sessionClose", m.hostSessionClose)

	e.Exports("buffer")
	if buf := e.Exports("buffer").Get("Buffer"); buf.Valid() && buf.IsFunction() {
		m.bufferCtor = buf.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}
	events := e.Exports("events")
	ee := events.Get("EventEmitter")
	if !ee.Valid() || !ee.IsFunction() {
		panic("godejs http2: events.EventEmitter is unavailable")
	}
	api, err := e.JS.Eval(prelude, "http2.js")
	if err != nil {
		panic("godejs http2: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V(), ee)
	if err != nil {
		panic("godejs http2: " + err.Error())
	}
	obj := built.Object()
	m.dispatchFn = obj.Get("dispatch").Keep()
	for _, key := range obj.OwnKeys() {
		if key == "dispatch" {
			continue
		}
		m.exports.Set(key, obj.Get(key))
	}
	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

func Implemented() []string {
	return []string{
		"http2.createServer", "http2.createSecureServer", "http2.connect",
		"http2.Http2Server", "http2.Http2ServerRequest", "http2.Http2ServerResponse",
		"http2.Http2Stream", "http2.ClientHttp2Session", "http2.ClientHttp2Stream",
		"http2.Http2Stream#pushStream", "http2.Http2Stream#respondWithFile",
		"http2.Http2Stream#respondWithFD", "http2.Http2ServerResponse#addTrailers",
		"http2.ClientHttp2Session#settings", "http2.constants",
	}
}

func (m *module) nextID() float64 {
	m.idMu.Lock()
	m.seq++
	id := m.seq
	m.idMu.Unlock()
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
			return m.e.JS.Undefined(), nil
		}
		fn := m.queued[0]
		m.queued = m.queued[1:]
		m.mu.Unlock()
		fn()
	}
}

func (m *module) addHandle() {
	m.mu.Lock()
	m.active++
	if m.pumpID == 0 {
		m.pumpID = m.e.SetTimeout(m.host.Get("pump"), pumpEveryMS, true, nil)
	}
	m.mu.Unlock()
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

func (m *module) shutdown() {
	m.mu.Lock()
	servers := make([]*server, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	sessions := make([]*session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.active = 0
	id := m.pumpID
	m.pumpID = 0
	m.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
	for _, s := range servers {
		if s.ln != nil {
			_ = s.ln.Close()
		}
	}
	for _, s := range sessions {
		if s.transport != nil {
			s.transport.CloseIdleConnections()
		}
	}
}

func (m *module) hostSessionSettings(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sessionOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not an Http2Session")
	}
	settings := arg(args, 0)
	if settings.Valid() && settings.IsObject() {
		// golang.org/x/net/http2 exposes initial SETTINGS through Transport and
		// Server configuration, but no runtime SETTINGS-frame API. Record the
		// requested values on the session and acknowledge the request so code
		// that uses the callback path continues to run.
		this.Set("localSettings", settings)
	}
	return this.V(), nil
}

func (m *module) emit(obj js.Object, name string, args ...js.Value) error {
	fn := obj.Get("emit")
	if !fn.Valid() || !fn.IsFunction() {
		return nil
	}
	_, err := fn.Object().Call(obj, append([]js.Value{m.e.JS.Str(name)}, args...)...)
	return err
}

func (m *module) srvOf(this js.Object) *server {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	return m.servers[v.Float()]
}

func (m *module) hostSrvInit(this js.Object, args []js.Value) (js.Value, error) {
	secure := false
	if v := arg(args, 0); v.IsBool() {
		secure = v.Bool()
	}
	m.nextID()
	id := m.seq
	s := &server{id: id, obj: this, secure: secure, conns: map[stdnet.Conn]bool{}}
	m.mu.Lock()
	m.servers[id] = s
	m.mu.Unlock()
	this.SetNum("_id", id)
	this.Protect()
	return m.e.JS.Num(id), nil
}

func (m *module) parseListen(args []js.Value) (string, string, js.Value, error) {
	values := args
	if len(args) == 1 && args[0].IsArray() {
		arr := args[0].Object()
		values = make([]js.Value, arr.Length())
		for i := range values {
			values[i] = arr.At(i)
		}
	}
	first := arg(values, 0)
	second := arg(values, 1)
	var cb js.Value
	if first.Valid() && first.IsFunction() {
		cb = first
		first = js.Value{}
	}
	if second.Valid() && second.IsFunction() {
		cb = second
		second = js.Value{}
	}
	if len(values) >= 3 && values[2].IsFunction() {
		cb = values[2]
	}
	host := "localhost"
	port := 0
	switch {
	case first.Valid() && first.IsNumber():
		port = int(first.Float())
		if second.Valid() && second.IsString() && second.Str() != "" {
			host = second.Str()
		}
	case first.Valid() && first.IsString():
		return "unix", first.Str(), cb, nil
	case first.Valid() && first.IsObject():
		o := first.Object()
		if p := o.Get("path"); p.IsString() && p.Str() != "" {
			return "unix", p.Str(), cb, nil
		}
		if p := o.Get("port"); p.IsNumber() {
			port = int(p.Float())
		} else if p.IsString() {
			n, err := strconv.Atoi(p.Str())
			if err != nil {
				return "", "", js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "invalid port")
			}
			port = n
		}
		if h := o.Get("host"); h.IsString() && h.Str() != "" {
			host = h.Str()
		}
	default:
		return "", "", js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "invalid listen arguments")
	}
	return "tcp", stdnet.JoinHostPort(host, strconv.Itoa(port)), cb, nil
}

func (m *module) hostSrvListen(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not an Http2Server")
	}
	if s.listening {
		return js.Value{}, js.ThrowCode("Error", "ERR_SERVER_ALREADY_LISTEN", "Listen method has been called more than once without closing.")
	}
	network, address, cb, err := m.parseListen(args)
	if err != nil {
		return js.Value{}, err
	}
	ln, err := clusterutil.Listen(network, address)
	if err != nil {
		obj := m.e.JS.Error(err.Error())
		obj.SetStr("code", "EADDRINUSE")
		_ = m.emit(s.obj, "error", obj.V())
		return this.V(), nil
	}
	if s.secure {
		cfg, cfgErr := serverTLSConfig(this)
		if cfgErr != nil {
			_ = ln.Close()
			return js.Value{}, cfgErr
		}
		ln = tls.NewListener(ln, cfg)
	}
	s.ln = ln
	s.listening = true
	s.closed = false
	addr := ln.Addr().String()
	s.address, s.port = splitHostPort(addr)
	s.h2 = &http2.Server{}
	m.applyServerSettings(s)
	s.obj.SetBool("listening", true)
	m.addHandle()
	go m.acceptLoop(s)
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, this.V())
	}
	return this.V(), nil
}

func serverTLSConfig(obj js.Object) (*tls.Config, error) {
	o := obj.Get("_options")
	certPEM, keyPEM := "", ""
	if o.Valid() && o.IsObject() {
		certPEM = pemValue(o.Object().Get("cert"))
		keyPEM = pemValue(o.Object().Get("key"))
	}
	if certPEM == "" || keyPEM == "" {
		return nil, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", "HTTP/2 secure server requires cert and key")
	}
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", err.Error())
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"h2"}}, nil
}

func (m *module) acceptLoop(s *server) {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.connMu.Lock()
		s.conns[c] = true
		s.connMu.Unlock()
		go func(conn stdnet.Conn) {
			defer func() {
				_ = conn.Close()
				s.connMu.Lock()
				delete(s.conns, conn)
				s.connMu.Unlock()
			}()
			if tc, ok := conn.(*tls.Conn); ok {
				if err := tc.Handshake(); err != nil {
					return
				}
			}
			s.h2.ServeConn(conn, &http2.ServeConnOpts{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					m.handle(s, w, r)
				}),
			})
		}(c)
	}
}

func (m *module) hostSrvClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not an Http2Server")
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
	s.listening = false
	s.obj.SetBool("listening", false)
	if s.ln != nil {
		_ = s.ln.Close()
	}
	s.connMu.Lock()
	for c := range s.conns {
		_ = c.Close()
	}
	s.connMu.Unlock()
	s.closed = true
	m.dropHandle()
	_ = m.emit(s.obj, "close")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) hostSrvAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.srvOf(this)
	if s == nil || !s.listening || s.ln == nil {
		return m.e.JS.Null(), nil
	}
	o := m.e.JS.Obj()
	o.SetStr("address", s.address)
	o.SetNum("port", float64(s.port))
	o.SetStr("family", familyOf(s.ln.Addr()))
	return o.V(), nil
}

func (m *module) handle(s *server, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	b := &bridge{id: m.nextID(), rw: w, ch: make(chan command, 128), er: make(chan error, 1)}
	m.mu.Lock()
	m.bridges[b.id] = b
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.bridges, b.id)
		m.mu.Unlock()
	}()
	localAddr, localPort := splitHostPort(r.Host)
	if a, ok := r.Context().Value(http.LocalAddrContextKey).(stdnet.Addr); ok {
		localAddr, localPort = splitHostPort(a.String())
	}
	remoteAddr, remotePort := splitHostPort(r.RemoteAddr)
	m.push(func() {
		if s.closed {
			select {
			case b.er <- fmt.Errorf("server closed"):
			default:
			}
			return
		}
		if err := m.dispatch(s, b, r, body, remoteAddr, remotePort, localAddr, localPort); err != nil {
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
					if strings.HasPrefix(h.name, ":") || strings.EqualFold(h.name, "connection") {
						continue
					}
					w.Header().Add(h.name, h.value)
				}
				w.WriteHeader(cmd.status)
				headersWritten = true
			case "write":
				_, _ = w.Write(cmd.body)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case "trailers":
				for _, h := range cmd.trailers {
					if strings.HasPrefix(h.name, ":") || strings.EqualFold(h.name, "connection") {
						continue
					}
					w.Header().Add("Trailer", h.name)
					w.Header().Set(http.TrailerPrefix+h.name, h.value)
				}
			case "end":
				if len(cmd.body) > 0 {
					_, _ = w.Write(cmd.body)
				}
				if !headersWritten {
					w.WriteHeader(200)
				}
				return
			case "abort":
				return
			}
		case err := <-b.er:
			if !headersWritten {
				w.WriteHeader(500)
			}
			fmt.Fprintf(m.e.Err(), "godejs http2: request handler failed: %v\n", err)
			return
		case <-r.Context().Done():
			return
		}
	}
}

func (m *module) dispatch(s *server, b *bridge, r *http.Request, body []byte, remoteAddr string, remotePort int, localAddr string, localPort int) error {
	raw := m.e.JS.Arr(nil)
	n := 0
	add := func(k, v string) {
		raw.SetAt(n, m.e.JS.Str(k))
		raw.SetAt(n+1, m.e.JS.Str(v))
		n += 2
	}
	add(":method", r.Method)
	add(":path", r.URL.RequestURI())
	add(":scheme", schemeOf(r))
	add(":authority", r.Host)
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		for _, v := range r.Header.Values(k) {
			add(k, v)
		}
	}
	_, err := m.dispatchFn.Object().Call(js.Object{}, s.obj.V(), m.e.JS.Num(b.id),
		m.e.JS.Str(r.Method), m.e.JS.Str(r.URL.RequestURI()), m.e.JS.Str(schemeOf(r)),
		m.e.JS.Str(r.Host), raw.V(), m.buffer(body),
		m.e.JS.Str(remoteAddr), m.e.JS.Num(float64(remotePort)),
		m.e.JS.Str(localAddr), m.e.JS.Num(float64(localPort)))
	return err
}

func schemeOf(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func (m *module) hostRespHead(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V(), nil
	}
	status := int(arg(args, 0).Float())
	if status == 0 {
		status = 200
	}
	b.ch <- command{kind: "head", status: status, headers: headersFromJS(arg(args, 1))}
	return this.V(), nil
}

func (m *module) hostRespWrite(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V(), nil
	}
	body, _ := arg(args, 0).Bytes()
	select {
	case b.ch <- command{kind: "write", body: body}:
	default:
	}
	return this.V(), nil
}

func (m *module) hostRespEnd(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V(), nil
	}
	body, _ := arg(args, 0).Bytes()
	select {
	case b.ch <- command{kind: "end", body: body}:
	default:
	}
	return this.V(), nil
}

func (m *module) hostRespTrailers(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V(), nil
	}
	select {
	case b.ch <- command{kind: "trailers", trailers: headersFromJS(arg(args, 0))}:
	default:
	}
	return this.V(), nil
}

func (m *module) hostRespPush(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil || b.rw == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_HTTP2_PUSH_DISABLED", "HTTP/2 push is disabled")
	}
	pusher, ok := b.rw.(http.Pusher)
	if !ok {
		return js.Value{}, js.ThrowCode("Error", "ERR_HTTP2_PUSH_DISABLED", "HTTP/2 push is disabled")
	}
	headers := arg(args, 0)
	path := "/"
	method := "GET"
	po := &http.PushOptions{Header: http.Header{}}
	if headers.Valid() && headers.IsObject() {
		o := headers.Object()
		for _, key := range o.OwnKeys() {
			v := o.Get(key)
			switch strings.ToLower(key) {
			case ":path", "path":
				if v.IsString() {
					path = v.Str()
				}
			case ":method", "method":
				if v.IsString() {
					method = strings.ToUpper(v.Str())
				}
			case ":status", ":scheme", ":authority":
				continue
			default:
				if v.IsArray() {
					for i := 0; i < v.Object().Length(); i++ {
						po.Header.Add(key, v.Object().At(i).Str())
					}
				} else if !v.IsUndefined() && !v.IsNull() {
					po.Header.Add(key, v.Str())
				}
			}
		}
	}
	po.Method = method
	if err := pusher.Push(path, po); err != nil {
		if err == http.ErrNotSupported {
			return js.Value{}, js.ThrowCode("Error", "ERR_HTTP2_PUSH_DISABLED", "HTTP/2 push is disabled")
		}
		return js.Value{}, m.http2Error(err)
	}
	return this.V(), nil
}

func (m *module) hostRespFile(this js.Object, args []js.Value) (js.Value, error) {
	path := arg(args, 0)
	if !path.IsString() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "path" argument must be a string.`)
	}
	data, err := os.ReadFile(path.Str())
	if err != nil {
		return js.Value{}, m.http2Error(err)
	}
	return m.sendWholeResponse(this, data, arg(args, 1)), nil
}

func (m *module) hostRespFD(this js.Object, args []js.Value) (js.Value, error) {
	fd := arg(args, 0)
	if !fd.IsNumber() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "fd" argument must be of type number.`)
	}
	f := os.NewFile(uintptr(int(fd.Int())), "http2-fd")
	if f == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", "invalid file descriptor")
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return js.Value{}, m.http2Error(err)
	}
	return m.sendWholeResponse(this, data, arg(args, 1)), nil
}

func (m *module) sendWholeResponse(this js.Object, data []byte, opts js.Value) js.Value {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V()
	}
	status := 200
	headers := []headerPair{}
	if opts.Valid() && opts.IsObject() {
		o := opts.Object()
		if v := o.Get(":status"); v.IsNumber() {
			status = int(v.Float())
		}
		if v := o.Get("headers"); v.Valid() && v.IsObject() {
			headers = headersFromJS(v)
		}
		if v := o.Get("offset"); v.IsNumber() && int(v.Int()) >= 0 && int(v.Int()) < len(data) {
			data = data[int(v.Int()):]
		}
		if v := o.Get("length"); v.IsNumber() && int(v.Int()) >= 0 && int(v.Int()) < len(data) {
			data = data[:int(v.Int())]
		}
	}
	headers = append(headers, headerPair{"content-length", strconv.Itoa(len(data))})
	b.ch <- command{kind: "head", status: status, headers: headers}
	b.ch <- command{kind: "end", body: data}
	return this.V()
}

func (m *module) http2Error(err error) error {
	return js.ThrowCode("Error", "ERR_HTTP2_ERROR", err.Error())
}

func (m *module) applyServerSettings(s *server) {
	if s == nil || s.obj.Get("_options").IsUndefined() {
		return
	}
	o := s.obj.Get("_options")
	if !o.IsObject() {
		return
	}
	settings := o.Object().Get("settings")
	if !settings.Valid() || !settings.IsObject() {
		return
	}
	so := settings.Object()
	if v := so.Get("maxConcurrentStreams"); v.IsNumber() && v.Int() > 0 {
		s.h2.MaxConcurrentStreams = uint32(v.Int())
	}
	if v := so.Get("maxFrameSize"); v.IsNumber() && v.Int() > 0 {
		s.h2.MaxReadFrameSize = uint32(v.Int())
	}
	if v := so.Get("maxSessionMemory"); v.IsNumber() && v.Int() > 0 {
		s.h2.MaxUploadBufferPerConnection = int32(v.Int())
	}
	if v := so.Get("streamIdleTimeout"); v.IsNumber() && v.Int() > 0 {
		s.h2.IdleTimeout = time.Duration(v.Int()) * time.Millisecond
	}
}

func (m *module) hostRespAbort(this js.Object, args []js.Value) (js.Value, error) {
	b := m.bridgeOf(this)
	if b == nil {
		return this.V(), nil
	}
	select {
	case b.ch <- command{kind: "abort"}:
	default:
	}
	return this.V(), nil
}

func (m *module) bridgeOf(this js.Object) *bridge {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	b := m.bridges[v.Float()]
	m.mu.Unlock()
	return b
}

func (m *module) buffer(b []byte) js.Value {
	if m.bufferCtor.Valid() && m.bufferFrom.Valid() {
		if v, err := m.bufferFrom.Call(m.bufferCtor, m.e.JS.Bytes(b).V()); err == nil {
			return v
		}
	}
	return m.e.JS.Bytes(b).V()
}

func headersFromJS(v js.Value) []headerPair {
	var out []headerPair
	if !v.Valid() || !v.IsObject() {
		return out
	}
	o := v.Object()
	for _, key := range o.OwnKeys() {
		val := o.Get(key)
		if val.IsArray() {
			arr := val.Object()
			for i := 0; i < arr.Length(); i++ {
				out = append(out, headerPair{key, arr.At(i).Str()})
			}
		} else if !val.IsUndefined() && !val.IsNull() {
			out = append(out, headerPair{key, val.Str()})
		}
	}
	return out
}

func splitHostPort(address string) (string, int) {
	host, portStr, err := stdnet.SplitHostPort(address)
	if err != nil {
		return address, 0
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func familyOf(a stdnet.Addr) string {
	tcp, ok := a.(*stdnet.TCPAddr)
	if ok && tcp.IP.To4() == nil {
		return "IPv6"
	}
	return "IPv4"
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func pemValue(v js.Value) string {
	if v.IsString() {
		return v.Str()
	}
	if b, ok := v.Bytes(); ok {
		return string(b)
	}
	return ""
}

func (m *module) hostConnect(this js.Object, args []js.Value) (js.Value, error) {
	authority := arg(args, 0).Str()
	opts := arg(args, 1)
	u, err := url.Parse(authority)
	if err != nil || u.Host == "" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid authority: "+authority)
	}
	tr := &http2.Transport{}
	if u.Scheme == "http" {
		tr.AllowHTTP = true
		tr.DialTLSContext = func(ctx context.Context, network, addr string, cfg *tls.Config) (stdnet.Conn, error) {
			return (&stdnet.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
		}
	} else {
		cfg := &tls.Config{ServerName: u.Hostname(), NextProtos: []string{"h2"}}
		if opts.Valid() && opts.IsObject() {
			if r := opts.Object().Get("rejectUnauthorized"); r.IsBool() && !r.Bool() {
				cfg.InsecureSkipVerify = true
			}
			if ca := pemValue(opts.Object().Get("ca")); ca != "" {
				pool := x509.NewCertPool()
				if pool.AppendCertsFromPEM([]byte(ca)) {
					cfg.RootCAs = pool
				}
			}
			certPEM := pemValue(opts.Object().Get("cert"))
			keyPEM := pemValue(opts.Object().Get("key"))
			if certPEM != "" && keyPEM != "" {
				if cert, e := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); e == nil {
					cfg.Certificates = []tls.Certificate{cert}
				}
			}
		}
		tr.TLSClientConfig = cfg
	}
	m.nextID()
	s := &session{id: m.seq, obj: this, authority: u, transport: tr}
	m.mu.Lock()
	m.sessions[s.id] = s
	m.mu.Unlock()
	this.SetNum("_id", s.id)
	this.Protect()
	m.addHandle()
	return this.V(), nil
}

func (m *module) sessionOf(this js.Object) *session {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	s := m.sessions[v.Float()]
	m.mu.Unlock()
	return s
}

func (m *module) hostRequest(this js.Object, args []js.Value) (js.Value, error) {
	sess := m.sessionOf(this)
	if sess == nil {
		return js.Value{}, js.Throw("TypeError", "not an Http2Session")
	}
	streamObj := arg(args, 0).Object()
	headers := arg(args, 1)
	m.nextID()
	st := &clientStream{id: m.seq, obj: streamObj, sess: sess, headers: headers}
	m.mu.Lock()
	m.streams[st.id] = st
	m.mu.Unlock()
	streamObj.SetNum("_id", st.id)
	streamObj.Protect()
	return m.e.JS.Num(st.id), nil
}

func (m *module) streamOf(this js.Object) *clientStream {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	st := m.streams[v.Float()]
	m.mu.Unlock()
	return st
}

func (m *module) hostStreamWrite(this js.Object, args []js.Value) (js.Value, error) {
	st := m.streamOf(this)
	if st == nil || st.ended {
		return this.V(), nil
	}
	body, _ := arg(args, 0).Bytes()
	m.ensureClientStarted(st)
	if len(body) > 0 {
		st.bodyCh <- append([]byte(nil), body...)
	}
	return this.V(), nil
}

func (m *module) hostStreamEnd(this js.Object, args []js.Value) (js.Value, error) {
	st := m.streamOf(this)
	if st == nil || st.ended {
		return this.V(), nil
	}
	body, _ := arg(args, 0).Bytes()
	m.ensureClientStarted(st)
	if len(body) > 0 {
		st.bodyCh <- append([]byte(nil), body...)
	}
	st.ended = true
	close(st.bodyCh)
	return this.V(), nil
}

func (m *module) ensureClientStarted(st *clientStream) {
	if st.started {
		return
	}
	st.started = true
	st.bodyCh = make(chan []byte, 16)
	go m.doClientRequest(st, st.bodyCh)
}

type channelBody struct {
	ch  <-chan []byte
	buf []byte
}

func (r *channelBody) Read(p []byte) (int, error) {
	for len(r.buf) == 0 {
		b, ok := <-r.ch
		if !ok {
			return 0, io.EOF
		}
		r.buf = b
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

func (m *module) doClientRequest(st *clientStream, bodyCh <-chan []byte) {
	m.doClientRequestReader(st, &channelBody{ch: bodyCh})
}

func (m *module) doClientRequestReader(st *clientStream, body io.Reader) {
	defer func() {
		m.mu.Lock()
		delete(m.streams, st.id)
		m.mu.Unlock()
	}()
	method, path := "GET", "/"
	if st.headers.Valid() && st.headers.IsObject() {
		o := st.headers.Object()
		if v := o.Get(":method"); v.IsString() && v.Str() != "" {
			method = strings.ToUpper(v.Str())
		}
		if v := o.Get(":path"); v.IsString() && v.Str() != "" {
			path = v.Str()
		}
	}
	ref := st.sess.authority
	u := *ref
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	// :path is an origin-form request target. Assigning it to URL.Path would
	// percent-encode the query separator, so split it explicitly.
	if i := strings.IndexByte(path, '?'); i >= 0 {
		u.Path = path[:i]
		u.RawQuery = path[i+1:]
	} else {
		u.Path = path
		u.RawQuery = ""
	}
	req, err := http.NewRequest(method, u.String(), body)
	req.ContentLength = -1
	if err != nil {
		m.deliverError(st, err)
		return
	}
	if st.headers.Valid() && st.headers.IsObject() {
		o := st.headers.Object()
		for _, k := range o.OwnKeys() {
			if strings.HasPrefix(k, ":") {
				continue
			}
			v := o.Get(k)
			if v.IsArray() {
				for i := 0; i < v.Object().Length(); i++ {
					req.Header.Add(k, v.Object().At(i).Str())
				}
			} else {
				req.Header.Set(k, v.Str())
			}
		}
		if a := o.Get(":authority"); a.IsString() && a.Str() != "" {
			req.Host = a.Str()
		} else {
			req.Host = u.Host
		}
	}
	res, err := st.sess.transport.RoundTrip(req)
	if err != nil {
		m.deliverError(st, err)
		return
	}
	defer res.Body.Close()
	h := m.e.JS.Obj()
	h.SetStr(":status", strconv.Itoa(res.StatusCode))
	for k, vals := range res.Header {
		if strings.EqualFold(k, "Set-Cookie") {
			arr := m.e.JS.Arr(nil)
			for i, v := range vals {
				arr.SetAt(i, m.e.JS.Str(v))
			}
			h.Set(k, arr.V())
		} else if len(vals) > 0 {
			h.SetStr(strings.ToLower(k), vals[0])
		}
	}
	m.push(func() {
		_ = m.emit(st.obj, "response", h.V())
		_ = m.emit(st.obj, "headers", h.V())
	})
	data, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		m.deliverError(st, readErr)
		return
	}
	trailers := m.e.JS.Obj()
	for k, vals := range res.Trailer {
		if len(vals) > 0 {
			trailers.SetStr(strings.ToLower(k), vals[0])
		}
	}
	m.push(func() {
		if len(data) > 0 {
			_ = m.emit(st.obj, "data", m.buffer(data))
		}
		if len(res.Trailer) > 0 {
			_ = m.emit(st.obj, "trailers", trailers.V())
		}
		_ = m.emit(st.obj, "end")
		_ = m.emit(st.obj, "close")
	})
}

func (m *module) deliverError(st *clientStream, err error) {
	m.push(func() {
		e := m.e.JS.Error(err.Error())
		e.SetStr("code", "ERR_HTTP2_ERROR")
		_ = m.emit(st.obj, "error", e.V())
		_ = m.emit(st.obj, "close")
	})
}

func (m *module) hostSessionClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sessionOf(this)
	if s != nil && s.transport != nil {
		s.transport.CloseIdleConnections()
		m.mu.Lock()
		s.closed = true
		delete(m.sessions, s.id)
		m.mu.Unlock()
		m.dropHandle()
		_ = m.emit(this, "close")
	}
	return this.V(), nil
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}
