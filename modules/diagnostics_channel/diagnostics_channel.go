// Package diagnostics_channel implements the subscription core of Node's
// diagnostics_channel module. Libraries use it for optional instrumentation, so
// the important fidelity is that channel() and tracingChannel() always return
// usable objects and that publish is a no-op when nobody subscribed.
package diagnostics_channel

import (
	"sync"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("diagnostics_channel", Install) }

type subscriber struct {
	fn js.Value
}

type channelState struct {
	name        string
	subscribers []subscriber
}

type module struct {
	e   *env.Env
	mu  sync.Mutex
	all map[string]*channelState
}

// Install builds the diagnostics_channel exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e, all: map[string]*channelState{}}
	exports := e.JS.Obj()

	exports.SetFunc("channel", func(this js.Object, args []js.Value) (js.Value, error) {
		name := ""
		if len(args) > 0 {
			name = args[0].Str()
		}
		return m.channel(name).V(), nil
	})
	exports.SetFunc("hasSubscribers", func(this js.Object, args []js.Value) (js.Value, error) {
		name := ""
		if len(args) > 0 {
			name = args[0].Str()
		}
		m.mu.Lock()
		state := m.all[name]
		has := state != nil && len(state.subscribers) > 0
		m.mu.Unlock()
		return e.JS.Bool(has), nil
	})
	exports.SetFunc("subscribe", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return e.JS.Undefined(), nil
		}
		name := args[0].Str()
		ch := m.channel(name)
		ch.Get("subscribe").Object().Call(ch, args[1])
		return e.JS.Undefined(), nil
	})
	exports.SetFunc("unsubscribe", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return e.JS.Undefined(), nil
		}
		name := args[0].Str()
		ch := m.channel(name)
		ch.Get("unsubscribe").Object().Call(ch, args[1])
		return e.JS.Undefined(), nil
	})
	exports.SetFunc("tracingChannel", func(this js.Object, args []js.Value) (js.Value, error) {
		name := ""
		if len(args) > 0 {
			name = args[0].Str()
		}
		return m.tracingChannel(name).V(), nil
	})
	return exports
}

func (m *module) channel(name string) js.Object {
	m.mu.Lock()
	state, ok := m.all[name]
	if !ok {
		state = &channelState{name: name}
		m.all[name] = state
	}
	m.mu.Unlock()

	ch := m.e.JS.Obj()
	ch.SetStr("name", name)
	ch.SetFunc("subscribe", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 && args[0].IsFunction() {
			m.mu.Lock()
			state.subscribers = append(state.subscribers, subscriber{fn: args[0]})
			has := len(state.subscribers) > 0
			m.mu.Unlock()
			ch.SetBool("hasSubscribers", has)
		}
		return this.V(), nil
	})
	ch.SetFunc("unsubscribe", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			m.mu.Lock()
			kept := state.subscribers[:0]
			for _, s := range state.subscribers {
				if !s.fn.Same(args[0]) {
					kept = append(kept, s)
				}
			}
			state.subscribers = kept
			has := len(state.subscribers) > 0
			m.mu.Unlock()
			ch.SetBool("hasSubscribers", has)
		}
		return this.V(), nil
	})
	ch.SetFunc("publish", func(this js.Object, args []js.Value) (js.Value, error) {
		m.mu.Lock()
		list := append([]subscriber(nil), state.subscribers...)
		m.mu.Unlock()
		message := m.e.JS.Undefined()
		if len(args) > 0 {
			message = args[0]
		}
		for _, s := range list {
			if _, err := s.fn.Object().Call(m.e.JS.Obj(), message, m.e.JS.Str(name)); err != nil {
				return js.Value{}, err
			}
		}
		return m.e.JS.Undefined(), nil
	})
	ch.SetBool("hasSubscribers", false)
	return ch
}

func (m *module) tracingChannel(name string) js.Object {
	tc := m.e.JS.Obj()
	tc.SetStr("name", name)
	// lru-cache and similar libraries only gate reporting on these flags until
	// somebody subscribes, which GodeJS does not model yet.
	tc.SetBool("hasSubscribers", false)
	for _, event := range []string{"start", "end", "asyncStart", "asyncEnd", "error"} {
		ch := m.channel(name + ":" + event)
		tc.Set(event, ch.V())
	}
	noop := func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return m.e.JS.Undefined(), nil
	}
	holder := m.e.JS.Obj()
	for _, fn := range []string{"traceSync", "tracePromise", "traceCallback", "subscribe", "unsubscribe"} {
		holder.SetFunc(fn, noop)
		tc.Set(fn, holder.Get(fn))
	}
	return tc
}

// Implemented lists the tracked diagnostics_channel API entries.
func Implemented() []string {
	return []string{
		"diagnostics_channel.channel",
		"diagnostics_channel.hasSubscribers",
		"diagnostics_channel.subscribe",
		"diagnostics_channel.unsubscribe",
		"diagnostics_channel.tracingChannel",
	}
}
