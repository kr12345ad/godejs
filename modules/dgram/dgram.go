// Package dgram implements Node's UDP socket builtin on top of Go's
// net.UDPConn. Reads happen on a goroutine and are delivered on the JS thread
// through a pump timer, so 'message' listeners always run in JavaScript order.
package dgram

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"

	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/buffer"
	_ "godejs/modules/events"
)

func init() { env.Register("dgram", Install) }

const pumpEveryMS = 1

type module struct {
	e *env.Env

	mu     sync.Mutex
	queued []func()
	pumpID int
	active int
	socks  map[float64]*socket
	seq    float64

	exports    js.Object
	host       js.Object
	bufferCtor js.Object
	bufferFrom js.Object
}

type socket struct {
	id      float64
	obj     js.Object
	network string

	mu        sync.Mutex
	conn      *net.UDPConn
	remote    *net.UDPAddr
	listening bool
	closed    bool
	handle    bool
	gen       int
}

func Install(e *env.Env) js.Object {
	m := &module{e: e, socks: map[float64]*socket{}}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()
	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("init", m.hostInit)
	m.host.SetFuncHidden("bind", m.hostBind)
	m.host.SetFuncHidden("send", m.hostSend)
	m.host.SetFuncHidden("close", m.hostClose)
	m.host.SetFuncHidden("address", m.hostAddress)
	m.host.SetFuncHidden("remoteAddress", m.hostRemoteAddress)
	m.host.SetFuncHidden("setBroadcast", m.hostSetBroadcast)
	m.host.SetFuncHidden("setTTL", m.hostSetTTL)
	m.host.SetFuncHidden("setMulticastTTL", m.hostSetMulticastTTL)
	m.host.SetFuncHidden("setMulticastLoopback", m.hostSetMulticastLoopback)
	m.host.SetFuncHidden("setMulticastInterface", m.hostSetMulticastInterface)
	m.host.SetFuncHidden("addMembership", m.hostAddMembership)
	m.host.SetFuncHidden("dropMembership", m.hostDropMembership)
	m.host.SetFuncHidden("connect", m.hostConnect)
	m.host.SetFuncHidden("disconnect", m.hostDisconnect)
	m.host.SetFuncHidden("ref", m.hostRef)
	m.host.SetFuncHidden("unref", m.hostRef)

	buffer := e.Exports("buffer").Get("Buffer")
	if buffer.Valid() && buffer.IsFunction() {
		m.bufferCtor = buffer.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}
	events := e.Exports("events").Get("EventEmitter")
	if !events.Valid() || !events.IsFunction() {
		panic("godejs dgram: events.EventEmitter is unavailable")
	}

	prelude, err := e.JS.Eval(prelude, "dgram.js")
	if err != nil {
		panic("godejs dgram: " + err.Error())
	}
	built, err := prelude.Object().Call(js.Object{}, m.host.V(), events, buffer, m.bufferCtor.V())
	if err != nil {
		panic("godejs dgram: " + err.Error())
	}
	for _, name := range built.Object().OwnKeys() {
		m.exports.Set(name, built.Object().Get(name))
	}
	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

func Implemented() []string {
	return []string{
		"dgram.createSocket", "dgram.Socket", "dgram.Socket#address",
		"dgram.Socket#bind", "dgram.Socket#close", "dgram.Socket#connect",
		"dgram.Socket#disconnect", "dgram.Socket#send", "dgram.Socket#ref",
		"dgram.Socket#unref", "dgram.Socket#setBroadcast",
		"dgram.Socket#setMulticastLoopback", "dgram.Socket#setMulticastTTL",
		"dgram.Socket#addMembership", "dgram.Socket#dropMembership",
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
	socks := make([]*socket, 0, len(m.socks))
	for _, s := range m.socks {
		socks = append(socks, s)
	}
	id := m.pumpID
	m.pumpID = 0
	m.active = 0
	m.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
	for _, s := range socks {
		s.mu.Lock()
		if s.conn != nil {
			_ = s.conn.Close()
		}
		s.mu.Unlock()
	}
}

func (m *module) socketOf(this js.Object) *socket {
	v := this.Get("$godeDgramId")
	if !v.IsNumber() {
		return nil
	}
	m.mu.Lock()
	s := m.socks[v.Float()]
	m.mu.Unlock()
	return s
}

func (m *module) hostInit(this js.Object, args []js.Value) (js.Value, error) {
	network := argStr(args, 0)
	if network != "udp4" && network != "udp6" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			`The argument 'type' must be one of: 'udp4', 'udp6'`)
	}
	s := &socket{id: m.nextID(), obj: this, network: network}
	m.mu.Lock()
	m.socks[s.id] = s
	m.mu.Unlock()
	this.SetHidden("$godeDgramId", m.e.JS.Num(s.id))
	this.Protect()
	this.SetStr("type", network)
	this.SetBool("bound", false)
	this.SetBool("closed", false)
	this.SetBool("connecting", false)
	return this.V(), nil
}

func (m *module) hostBind(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	port := int(arg(args, 0).Int())
	address := argStr(args, 1)
	cb := arg(args, 2)
	s.mu.Lock()
	if s.conn != nil {
		s.mu.Unlock()
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_ALREADY_BOUND", "Socket is already bound")
	}
	if s.closed {
		s.mu.Unlock()
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_CLOSED", "Socket is closed")
	}
	addr := &net.UDPAddr{Port: port}
	if address != "" {
		ip := net.ParseIP(address)
		if ip == nil {
			s.mu.Unlock()
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				`The argument 'address' must be a valid IP address`)
		}
		addr.IP = ip
	}
	conn, err := net.ListenUDP(s.network, addr)
	if err != nil {
		s.mu.Unlock()
		return js.Value{}, m.listenError(err, address, port)
	}
	s.conn = conn
	s.listening = true
	s.handle = true
	s.gen++
	gen := s.gen
	s.mu.Unlock()
	m.addHandle()
	this.SetBool("bound", true)
	m.startReading(s, conn, gen)
	m.emitAsync(s.obj, "listening")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) listenError(err error, address string, port int) error {
	code := "EADDRINUSE"
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		code = "EACCES"
	case strings.Contains(err.Error(), "cannot assign requested address"):
		code = "EADDRNOTAVAIL"
	}
	e := m.e.JS.Error("bind " + code + " " + net.JoinHostPort(address, strconv.Itoa(port)))
	e.SetStr("code", code)
	e.SetStr("syscall", "bind")
	e.SetStr("address", address)
	e.SetNum("port", float64(port))
	return js.ThrowCode("Error", code, e.Get("message").Str())
}

func (m *module) startReading(s *socket, conn *net.UDPConn, gen int) {
	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if n > 0 {
				data := append([]byte(nil), buf[:n]...)
				m.push(func() { m.deliver(s, data, addr, gen) })
			}
			if err != nil {
				s.mu.Lock()
				closed := s.closed || s.gen != gen
				s.mu.Unlock()
				if !closed {
					m.push(func() {
						_ = m.emit(s.obj, "error", m.socketError(err, "recvfrom").V())
					})
				}
				return
			}
		}
	}()
}

func (m *module) deliver(s *socket, data []byte, addr *net.UDPAddr, gen int) {
	s.mu.Lock()
	closed := s.closed || s.gen != gen
	s.mu.Unlock()
	if closed {
		return
	}
	payload := m.buffer(data)
	rinfo := m.e.JS.Obj()
	rinfo.SetStr("address", addr.IP.String())
	rinfo.SetNum("family", float64(ipFamily(addr.IP)))
	rinfo.SetNum("port", float64(addr.Port))
	rinfo.SetNum("size", float64(len(data)))
	_ = m.emit(s.obj, "message", payload, rinfo.V())
}

func ipFamily(ip net.IP) int {
	if ip.To4() != nil {
		return 4
	}
	return 6
}

func (m *module) hostSend(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	buf := arg(args, 0)
	data, ok := buf.Bytes()
	if !ok {
		if buf.Valid() && buf.IsString() {
			data = []byte(buf.Str())
		} else {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "buffer" argument must be an instance of Buffer, TypedArray, or DataView.`)
		}
	}
	offset := int(arg(args, 1).Int())
	length := int(arg(args, 2).Int())
	if offset < 0 || length < 0 || offset+length > len(data) {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS",
			`The value of "offset" is out of range.`)
	}
	data = data[offset : offset+length]
	port := int(arg(args, 3).Int())
	address := argStr(args, 4)
	cb := arg(args, 5)

	s.mu.Lock()
	conn := s.conn
	remote := s.remote
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_CLOSED", "Socket is closed")
	}
	if conn == nil {
		// Node auto-binds an unbound socket on send().
		if _, err := m.hostBind(this, []js.Value{m.e.JS.Num(0), m.e.JS.Str("")}); err != nil {
			return js.Value{}, err
		}
		s.mu.Lock()
		conn = s.conn
		s.mu.Unlock()
	}
	var err error
	written := 0
	if remote != nil && address == "" {
		written, err = conn.Write(data)
	} else {
		var target *net.UDPAddr
		target, err = net.ResolveUDPAddr(s.network, net.JoinHostPort(address, strconv.Itoa(port)))
		if err == nil {
			written, err = conn.WriteToUDP(data, target)
		}
	}
	if err != nil {
		jsErr := m.socketError(err, "sendto")
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb, jsErr.V())
			return this.V(), nil
		}
		_ = m.emit(s.obj, "error", jsErr.V())
		return this.V(), nil
	}
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb, m.e.JS.Null(), m.e.JS.Num(float64(written)))
	}
	return this.V(), nil
}

func (m *module) socketError(err error, syscall string) js.Object {
	code := "EIO"
	if strings.Contains(err.Error(), "connection refused") {
		code = "ECONNREFUSED"
	} else if strings.Contains(err.Error(), "network is unreachable") {
		code = "ENETUNREACH"
	}
	e := m.e.JS.Error(err.Error())
	e.SetStr("code", code)
	e.SetStr("syscall", syscall)
	return e
}

func (m *module) hostClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	cb := arg(args, 0)
	s.mu.Lock()
	conn := s.conn
	if conn == nil || s.closed {
		s.mu.Unlock()
		err := js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_RUNNING", "Not running")
		if cb.Valid() && cb.IsFunction() {
			m.e.NextTick(cb, m.e.JS.ErrorValue(err))
			return this.V(), nil
		}
		return js.Value{}, err
	}
	s.closed = true
	s.listening = false
	s.conn = nil
	s.gen++
	s.mu.Unlock()
	_ = conn.Close()
	if s.handle {
		s.handle = false
		m.dropHandle()
	}
	this.SetBool("bound", false)
	this.SetBool("closed", true)
	m.emitAsync(s.obj, "close")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) hostAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_RUNNING", "Not running")
	}
	local := conn.LocalAddr().(*net.UDPAddr)
	o := m.e.JS.Obj()
	o.SetStr("address", local.IP.String())
	o.SetNum("family", float64(ipFamily(local.IP)))
	o.SetNum("port", float64(local.Port))
	return o.V(), nil
}

func (m *module) hostRemoteAddress(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	s.mu.Lock()
	remote := s.remote
	s.mu.Unlock()
	if remote == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_CONNECTED", "Not connected")
	}
	o := m.e.JS.Obj()
	o.SetStr("address", remote.IP.String())
	o.SetNum("family", float64(ipFamily(remote.IP)))
	o.SetNum("port", float64(remote.Port))
	return o.V(), nil
}

func (m *module) hostSetBroadcast(this js.Object, args []js.Value) (js.Value, error) {
	return m.withConn(this, func(conn *net.UDPConn) error {
		return udpSetInt(conn, unix.SOL_SOCKET, unix.SO_BROADCAST, boolInt(arg(args, 0).Bool()))
	})
}

func (m *module) hostSetTTL(this js.Object, args []js.Value) (js.Value, error) {
	return m.withConn(this, func(conn *net.UDPConn) error {
		return udpSetInt(conn, unix.IPPROTO_IP, unix.IP_TTL, int(arg(args, 0).Int()))
	})
}

func (m *module) hostSetMulticastTTL(this js.Object, args []js.Value) (js.Value, error) {
	return m.withConn(this, func(conn *net.UDPConn) error {
		return udpSetInt(conn, unix.IPPROTO_IP, unix.IP_MULTICAST_TTL, int(arg(args, 0).Int()))
	})
}

func (m *module) hostSetMulticastLoopback(this js.Object, args []js.Value) (js.Value, error) {
	return m.withConn(this, func(conn *net.UDPConn) error {
		return udpSetInt(conn, unix.IPPROTO_IP, unix.IP_MULTICAST_LOOP, boolInt(arg(args, 0).Bool()))
	})
}

func (m *module) hostSetMulticastInterface(this js.Object, args []js.Value) (js.Value, error) {
	ifaceName := argStr(args, 0)
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", "Invalid interface: "+ifaceName)
	}
	return m.withConn(this, func(conn *net.UDPConn) error {
		return udpSetMulticastInterface(conn, iface)
	})
}

func (m *module) withConn(this js.Object, fn func(*net.UDPConn) error) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_RUNNING", "Not running")
	}
	if err := fn(conn); err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET", err.Error())
	}
	return this.V(), nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func udpControl(conn *net.UDPConn, fn func(fd int) error) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	var setErr error
	if err := raw.Control(func(fd uintptr) {
		setErr = fn(int(fd))
	}); err != nil {
		return err
	}
	return setErr
}

func udpSetInt(conn *net.UDPConn, level, option, value int) error {
	return udpControl(conn, func(fd int) error {
		return unix.SetsockoptInt(fd, level, option, value)
	})
}

func udpSetMulticastInterface(conn *net.UDPConn, iface *net.Interface) error {
	if iface == nil {
		return errors.New("network interface is required")
	}
	return udpControl(conn, func(fd int) error {
		if ip := firstIPv4(iface); ip != nil {
			var address [4]byte
			copy(address[:], ip.To4())
			return unix.SetsockoptInet4Addr(fd, unix.IPPROTO_IP, unix.IP_MULTICAST_IF, address)
		}
		return unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_MULTICAST_IF, iface.Index)
	})
}

func firstIPv4(iface *net.Interface) net.IP {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.To4() != nil {
			return ip.To4()
		}
	}
	return nil
}

func (m *module) hostAddMembership(this js.Object, args []js.Value) (js.Value, error) {
	return m.membership(this, args, true)
}

func (m *module) hostDropMembership(this js.Object, args []js.Value) (js.Value, error) {
	return m.membership(this, args, false)
}

func (m *module) membership(this js.Object, args []js.Value, join bool) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	group := net.ParseIP(argStr(args, 0))
	if group == nil {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Invalid multicast group")
	}
	var iface *net.Interface
	if name := argStr(args, 1); name != "" {
		var err error
		iface, err = net.InterfaceByName(name)
		if err != nil {
			return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_ARG_VALUE", "Invalid interface: "+name)
		}
	}
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_RUNNING", "Not running")
	}
	if err := udpMembership(conn, group, iface, join); err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET", err.Error())
	}
	return this.V(), nil
}

func udpMembership(conn *net.UDPConn, group net.IP, iface *net.Interface, join bool) error {
	if ip4 := group.To4(); ip4 != nil {
		mreq := &unix.IPMreq{}
		copy(mreq.Multiaddr[:], ip4)
		if iface != nil {
			if local := firstIPv4(iface); local != nil {
				copy(mreq.Interface[:], local)
			}
		}
		option := unix.IP_ADD_MEMBERSHIP
		if !join {
			option = unix.IP_DROP_MEMBERSHIP
		}
		return udpControl(conn, func(fd int) error {
			return setsockoptRaw(fd, unix.IPPROTO_IP, option, unsafe.Pointer(mreq), unsafe.Sizeof(*mreq))
		})
	}
	mreq := &unix.IPv6Mreq{}
	copy(mreq.Multiaddr[:], group.To16())
	if iface != nil {
		mreq.Interface = uint32(iface.Index)
	}
	option := unix.IPV6_JOIN_GROUP
	if !join {
		option = unix.IPV6_LEAVE_GROUP
	}
	return udpControl(conn, func(fd int) error {
		return setsockoptRaw(fd, unix.IPPROTO_IPV6, option, unsafe.Pointer(mreq), unsafe.Sizeof(*mreq))
	})
}

// setsockoptRaw passes a struct sockopt value (ip_mreq, ipv6_mreq) straight to
// the kernel. x/sys/unix only ships the IPMreq convenience wrappers on Linux,
// so calling setsockopt directly keeps multicast working on macOS too.
func setsockoptRaw(fd, level, option int, value unsafe.Pointer, size uintptr) error {
	_, _, errno := unix.Syscall6(unix.SYS_SETSOCKOPT, uintptr(fd), uintptr(level),
		uintptr(option), uintptr(value), size, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func (m *module) hostConnect(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	port := int(arg(args, 0).Int())
	address := argStr(args, 1)
	cb := arg(args, 2)
	target, err := net.ResolveUDPAddr(s.network, net.JoinHostPort(address, strconv.Itoa(port)))
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET", err.Error())
	}
	s.mu.Lock()
	old := s.conn
	var local *net.UDPAddr
	if old != nil {
		local = old.LocalAddr().(*net.UDPAddr)
		_ = old.Close()
		s.conn = nil
	}
	s.gen++
	gen := s.gen
	s.mu.Unlock()

	conn, err := net.DialUDP(s.network, local, target)
	if err != nil {
		// Leave the socket usable if the connection attempt failed.
		if local != nil {
			if rebound, rebindErr := net.ListenUDP(s.network, local); rebindErr == nil {
				s.mu.Lock()
				s.conn = rebound
				s.gen++
				g := s.gen
				s.mu.Unlock()
				m.startReading(s, rebound, g)
			}
		}
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET", err.Error())
	}
	s.mu.Lock()
	s.conn = conn
	s.remote = target
	s.listening = true
	wasHandle := s.handle
	s.handle = true
	s.gen++
	gen = s.gen
	s.mu.Unlock()
	if !wasHandle {
		m.addHandle()
	}
	this.SetBool("bound", true)
	this.SetBool("connecting", false)
	m.startReading(s, conn, gen)
	m.emitAsync(s.obj, "connect")
	if cb.Valid() && cb.IsFunction() {
		m.e.NextTick(cb)
	}
	return this.V(), nil
}

func (m *module) hostDisconnect(this js.Object, args []js.Value) (js.Value, error) {
	s := m.socketOf(this)
	if s == nil {
		return js.Value{}, js.Throw("TypeError", "not a dgram.Socket")
	}
	s.mu.Lock()
	old := s.conn
	if old == nil || s.remote == nil {
		s.mu.Unlock()
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET_DGRAM_NOT_CONNECTED", "Not connected")
	}
	local := old.LocalAddr().(*net.UDPAddr)
	_ = old.Close()
	s.conn = nil
	s.remote = nil
	s.gen++
	s.mu.Unlock()
	conn, err := net.ListenUDP(s.network, local)
	if err != nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_SOCKET", err.Error())
	}
	s.mu.Lock()
	s.conn = conn
	s.gen++
	gen := s.gen
	s.mu.Unlock()
	m.startReading(s, conn, gen)
	return this.V(), nil
}

func (m *module) hostRef(this js.Object, args []js.Value) (js.Value, error) {
	return this.V(), nil
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

const prelude = `(function (host, EventEmitter, Buffer, BufferCtor) {
  'use strict';

  if (typeof EventEmitter !== 'function') throw new Error('events.EventEmitter is unavailable');

  function describe(value) {
    if (value === null) return 'null';
    var t = typeof value;
    if (t === 'string') return "type string ('" + value + "')";
    if (t === 'number' || t === 'boolean') return 'type ' + t + ' (' + value + ')';
    if (t === 'undefined') return 'undefined';
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }

  function invalidArg(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function bufferOf(value) {
    if (value instanceof Uint8Array) return value;
    if (typeof value === 'string') return Buffer.from(value);
    throw invalidArg('buffer', 'an instance of Buffer, TypedArray, or DataView', value);
  }

  class Socket extends EventEmitter {
    constructor(type, callback) {
      super();
      if (type !== 'udp4' && type !== 'udp6') {
        var err = new TypeError("The argument 'type' must be one of: 'udp4', 'udp6'. Received '" + type + "'");
        err.code = 'ERR_INVALID_ARG_VALUE';
        throw err;
      }
      host.init.call(this, type);
      if (typeof callback === 'function') this.on('message', callback);
    }
    bind(port, address, callback) {
      if (port && typeof port === 'object') {
        var options = port;
        port = options.port;
        address = options.address;
        callback = options.callback;
      }
      if (typeof port === 'function') { callback = port; port = 0; address = ''; }
      if (typeof address === 'function') { callback = address; address = ''; }
      host.bind.call(this, Number(port) || 0, address || '', callback);
      return this;
    }
    send(message, offset, length, port, address, callback) {
      var data = bufferOf(message);
      if (typeof offset === 'number' && typeof length === 'number') {
        // send(message, offset, length, port[, address][, callback])
        if (typeof address === 'function') { callback = address; address = ''; }
      } else {
        // send(message, port[, address][, callback])
        callback = (typeof callback === 'function') ? callback :
          (typeof address === 'function') ? address :
          (typeof port === 'function') ? port :
          (typeof length === 'function') ? length : undefined;
        address = (typeof length === 'string') ? length :
          (typeof address === 'string') ? address : '';
        port = (typeof offset === 'number') ? offset : 0;
        offset = 0;
        length = data.length;
      }
      if (address === undefined || address === null) address = '';
      host.send.call(this, data, offset, length, Number(port) || 0, String(address), callback);
      return this;
    }
    close(callback) { host.close.call(this, callback); return this; }
    address() { return host.address.call(this); }
    remoteAddress() { return host.remoteAddress.call(this); }
    setBroadcast(flag) { host.setBroadcast.call(this, !!flag); return this; }
    setTTL(ttl) { host.setTTL.call(this, Number(ttl)); return this; }
    setMulticastTTL(ttl) { host.setMulticastTTL.call(this, Number(ttl)); return this; }
    setMulticastLoopback(flag) { host.setMulticastLoopback.call(this, !!flag); return this; }
    setMulticastInterface(iface) { host.setMulticastInterface.call(this, String(iface)); return this; }
    addMembership(group, iface) { host.addMembership.call(this, String(group), iface ? String(iface) : ''); return this; }
    dropMembership(group, iface) { host.dropMembership.call(this, String(group), iface ? String(iface) : ''); return this; }
    connect(port, address, callback) {
      if (typeof port === 'object') { callback = address; address = port.address; port = port.port; }
      if (typeof address === 'function') { callback = address; address = ''; }
      host.connect.call(this, Number(port) || 0, String(address || ''), callback);
      return this;
    }
    disconnect() { host.disconnect.call(this); return this; }
    ref() { host.ref.call(this); return this; }
    unref() { host.unref.call(this); return this; }
    getRecvBufferSize() { return 0; }
    setRecvBufferSize(size) { return this; }
    getSendBufferSize() { return 0; }
    setSendBufferSize(size) { return this; }
  }

  function createSocket(type, callback) { return new Socket(type, callback); }

  return { createSocket: createSocket, Socket: Socket };
})
`
