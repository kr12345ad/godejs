package child_process

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"

	"godejs/internal/js"
)

// The fork() channel is newline-delimited JSON. Node itself uses a length
// prefixed JSON protocol on fd 3; because both ends here are this runtime, the
// framing only has to be unambiguous, and JSON.stringify never emits a raw
// newline inside a string.
//
// Two real OS pipes are used rather than io.Pipe: os/exec has to spawn a
// copying goroutine for a generic Reader, and that goroutine blocks the child's
// Wait() until the write end closes. A child that never reads stdin would then
// deadlock the parent.
const (
	ipcRecvEnv = "GODE_IPC_RECV_FD"
	ipcSendEnv = "GODE_IPC_SEND_FD"
)

var errIPCClosed = errors.New("Channel closed")

// ipcChannel is the parent end of a fork() channel.
type ipcChannel struct {
	mu     sync.Mutex
	w      *os.File
	r      *os.File
	closed bool
}

func (c *ipcChannel) write(line string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.w == nil {
		return errIPCClosed
	}
	_, err := c.w.WriteString(line + "\n")
	return err
}

func (c *ipcChannel) closeWrite() {
	c.mu.Lock()
	if !c.closed {
		c.closed = true
		if c.w != nil {
			_ = c.w.Close()
		}
	}
	c.mu.Unlock()
}

func (c *ipcChannel) closeAll() {
	c.closeWrite()
	c.mu.Lock()
	if c.r != nil {
		_ = c.r.Close()
		c.r = nil
	}
	c.mu.Unlock()
}

// makeIPCPipes creates the two pipes a fork() needs. It returns the parent's
// write end, the files the child inherits (fd 3 is its read end, fd 4 its write
// end) and the parent's read end.
func makeIPCPipes() (parentW *os.File, childFiles []*os.File, parentR *os.File, err error) {
	childR, parentW, err := os.Pipe()
	if err != nil {
		return nil, nil, nil, err
	}
	parentR, childW, err := os.Pipe()
	if err != nil {
		_ = childR.Close()
		_ = parentW.Close()
		return nil, nil, nil, err
	}
	return parentW, []*os.File{childR, childW}, parentR, nil
}

// readIPCLines reads frames until the pipe closes. onLine runs on a reader
// goroutine, so it may only touch Go state guarded by a mutex; onDone runs once
// after the last frame (or when the fd is unusable).
func readIPCLines(r *os.File, onLine func(string), onDone func()) {
	go func() {
		defer onDone()
		br := bufio.NewReaderSize(r, 64*1024)
		for {
			line, err := br.ReadString('\n')
			if len(line) > 0 {
				line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
				if line != "" {
					onLine(line)
				}
			}
			if err != nil {
				return
			}
		}
	}()
}

// childIPC is the child end of a fork() channel, stored on the module so
// process.send and the pump can reach it.
type childIPC struct {
	mu        sync.Mutex
	w         *os.File
	queue     []string
	dead      bool
	connected bool
	refed     bool
	timer     int
}

// installChildIPC wires process.send/process.disconnect/channel when this
// process was started by fork(). The module is installed eagerly in every
// process, so this runs before user code.
func (m *module) installChildIPC() {
	if m.childIPC != nil {
		// A repeat call (the install-time re-run) must not start a second
		// reader on the same descriptor; it only restores the process
		// bindings the process module may have overwritten.
		if m.ipcChan.Valid() {
			m.wireChildIPCProcess()
		}
		return
	}
	recvFD, err := strconv.Atoi(os.Getenv(ipcRecvEnv))
	if err != nil || recvFD <= 2 {
		return
	}
	sendFD, err := strconv.Atoi(os.Getenv(ipcSendEnv))
	if err != nil || sendFD <= 2 {
		return
	}
	recv := os.NewFile(uintptr(recvFD), "gode-ipc-in")
	send := os.NewFile(uintptr(sendFD), "gode-ipc-out")
	if recv == nil || send == nil {
		return
	}
	ch := &childIPC{w: send, connected: true}
	m.childIPC = ch

	proc := m.e.Exports("process")
	if !proc.Valid() {
		return
	}

	// The reader goroutine cannot touch JavaScriptCore, so it only appends to
	// the queue; a repeating timer on the JS thread drains it. The timer is
	// what keeps an idle IPC child alive, exactly like a referenced channel.
	readIPCLines(recv, func(line string) {
		ch.mu.Lock()
		ch.queue = append(ch.queue, line)
		ch.mu.Unlock()
	}, func() {
		ch.mu.Lock()
		ch.dead = true
		ch.connected = false
		ch.mu.Unlock()
	})

	// Node's channel is unref'd until someone listens for 'message'/'disconnect'
	// or calls channel.ref(); a forked child with no such listener must be able
	// to exit on its own instead of being pinned alive by the channel.
	kick := m.e.JS.Obj()
	kick.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
		m.syncChildIPCTimer(proc)
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(kick.Get("f"))

	chann := m.e.JS.Obj()
	chann.SetFunc("ref", func(this js.Object, args []js.Value) (js.Value, error) {
		ch.mu.Lock()
		ch.refed = true
		ch.mu.Unlock()
		m.syncChildIPCTimer(proc)
		return this.V(), nil
	})
	chann.SetFunc("unref", func(this js.Object, args []js.Value) (js.Value, error) {
		ch.mu.Lock()
		ch.refed = false
		ch.mu.Unlock()
		m.syncChildIPCTimer(proc)
		return this.V(), nil
	})
	chann.SetFunc("close", func(this js.Object, args []js.Value) (js.Value, error) {
		m.childDisconnect(proc)
		return m.e.JS.Undefined(), nil
	})
	chann.Set("_handle", m.e.JS.Null())
	m.ipcChan = chann.Keep()
	m.wireChildIPCProcess()
}

// wireChildIPCProcess publishes the child end of the channel on `process`.
// It is idempotent so it can be re-run after every builtin has been installed:
// the process module's own Install defines channel-less send/disconnect stubs,
// and module install order decides which definition wins.
func (m *module) wireChildIPCProcess() {
	ch := m.childIPC
	proc := m.e.Exports("process")
	if ch == nil || !proc.Valid() || !m.ipcChan.Valid() {
		return
	}
	proc.SetBool("connected", true)
	proc.Set("channel", m.ipcChan.V())
	proc.SetFunc("send", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.childSend(proc, args)
	})
	proc.SetFunc("disconnect", func(this js.Object, args []js.Value) (js.Value, error) {
		m.childDisconnect(proc)
		return m.e.JS.Undefined(), nil
	})
}

// syncChildIPCTimer starts or stops the drain timer. Only a channel that is
// ref'd, or one that has a listener for 'message'/'disconnect', holds the loop
// open (see test-child-process-send-cb and test-child-process-fork-ref).
func (m *module) syncChildIPCTimer(proc js.Object) {
	ch := m.childIPC
	if ch == nil {
		return
	}
	ch.mu.Lock()
	connected := ch.connected
	refed := ch.refed
	running := ch.timer != 0
	ch.mu.Unlock()
	need := connected && refed
	if connected && !need {
		if lc := proc.Get("listenerCount"); lc.Valid() && lc.IsFunction() {
			if v, err := lc.Object().Call(proc, m.e.JS.Str("message")); err == nil && v.Float() > 0 {
				need = true
			}
			if !need {
				if v, err := lc.Object().Call(proc, m.e.JS.Str("disconnect")); err == nil && v.Float() > 0 {
					need = true
				}
			}
		}
	}
	if need && !running {
		pump := m.e.JS.Obj()
		pump.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
			m.syncChildIPCTimer(proc)
			return m.childIPCPump(proc), nil
		})
		id := m.e.SetTimeout(pump.Get("f"), 2, true, nil)
		ch.mu.Lock()
		ch.timer = id
		ch.mu.Unlock()
		return
	}
	if !need && running {
		ch.mu.Lock()
		id := ch.timer
		ch.timer = 0
		ch.mu.Unlock()
		m.e.ClearTimer(id)
	}
}

// childIPCPump delivers queued frames and keeps the loop alive while the
// channel is open.
func (m *module) childIPCPump(proc js.Object) js.Value {
	ch := m.childIPC
	if ch == nil {
		return m.e.JS.Undefined()
	}
	ch.mu.Lock()
	lines := ch.queue
	ch.queue = nil
	dead := ch.dead
	connected := ch.connected
	ch.mu.Unlock()

	emit := proc.Get("emit")
	for _, line := range lines {
		if !emit.Valid() || !emit.IsFunction() {
			break
		}
		v, err := m.e.JS.Eval(`JSON.parse(`+strconv.Quote(line)+`)`, "[gode:ipc]")
		if err != nil {
			continue
		}
		if _, err := emit.Object().Call(proc, m.e.JS.Str("message"), v); err != nil {
			panic(err)
		}
	}
	if dead || !connected {
		proc.SetBool("connected", false)
		ch.mu.Lock()
		id := ch.timer
		ch.timer = 0
		ch.mu.Unlock()
		if id != 0 {
			m.e.ClearTimer(id)
		}
	}
	return m.e.JS.Undefined()
}

func (m *module) childSend(proc js.Object, args []js.Value) (js.Value, error) {
	ch := m.childIPC
	if ch == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_IPC_CHANNEL_CLOSED", "Channel closed")
	}
	msg := arg(args, 0)
	if !msg.Valid() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_MISSING_ARGS",
			`The "message" argument must be specified`)
	}
	ch.mu.Lock()
	connected := ch.connected
	w := ch.w
	ch.mu.Unlock()
	if !connected || w == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_IPC_CHANNEL_CLOSED", "Channel closed")
	}
	text, err := m.stringify(msg)
	if err != nil {
		return js.Value{}, err
	}
	if _, err := w.WriteString(text + "\n"); err != nil {
		return m.e.JS.Bool(false), nil
	}
	if cb := lastFunction(args); cb.Valid() {
		m.e.NextTick(cb, m.e.JS.Null())
	}
	return m.e.JS.Bool(true), nil
}

func (m *module) childDisconnect(proc js.Object) {
	ch := m.childIPC
	if ch == nil {
		return
	}
	ch.mu.Lock()
	if !ch.connected {
		ch.mu.Unlock()
		return
	}
	ch.connected = false
	w := ch.w
	ch.mu.Unlock()
	ch.closeWriteOnly(w)
	proc.SetBool("connected", false)
	if ch.timer != 0 {
		m.e.ClearTimer(ch.timer)
		ch.timer = 0
	}
	if emit := proc.Get("emit"); emit.Valid() && emit.IsFunction() {
		if _, err := emit.Object().Call(proc, m.e.JS.Str("disconnect")); err != nil {
			panic(err)
		}
	}
}

func (ch *childIPC) closeWriteOnly(w *os.File) {
	if w != nil {
		_ = w.Close()
	}
	ch.mu.Lock()
	ch.w = nil
	ch.mu.Unlock()
}

// stringify renders a value the way JSON.stringify does.
func (m *module) stringify(v js.Value) (string, error) {
	fn, err := m.e.JS.Eval(`JSON.stringify`, "[gode:ipc]")
	if err != nil || !fn.IsFunction() {
		return "", js.Throw("Error", "IPC message could not be serialized")
	}
	out, err := fn.Object().Call(js.Object{}, v)
	if err != nil {
		return "", err
	}
	if !out.Valid() || out.IsUndefined() {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "message" argument must be serializable`)
	}
	return out.Str(), nil
}

func lastFunction(args []js.Value) js.Value {
	for i := len(args) - 1; i >= 0; i-- {
		if args[i].IsFunction() {
			return args[i]
		}
	}
	return js.Value{}
}

// stripIPCEnv removes channel markers from an inherited environment so a
// grandchild does not believe it owns a pipe that was never created for it.
func stripIPCEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		switch {
		case strings.HasPrefix(kv, ipcRecvEnv+"="),
			strings.HasPrefix(kv, ipcSendEnv+"="),
			strings.HasPrefix(kv, "NODE_CHANNEL_FD="),
			strings.HasPrefix(kv, "NODE_CHANNEL_SERIALIZATION_MODE="):
			continue
		}
		out = append(out, kv)
	}
	return out
}
