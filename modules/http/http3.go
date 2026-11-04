// HTTP/3 server support. It reuses modules/http's Node-style request,
// response and event-loop bridge; only the transport is different (QUIC/UDP
// instead of TCP, via quic-go's HTTP/3 server).
package http

import (
	"errors"
	stdnet "net"
	stdhttp "net/http"
	"sync"

	"github.com/quic-go/quic-go/http3"

	"godejs/internal/env"
	"godejs/internal/js"
)

// activeHTTP maps an Env to the single http module instance. The http3 module
// installs after http and reuses its server bridge and host functions.
var activeHTTP sync.Map

func init() { env.Register("http3", InstallHTTP3) }

const http3Prelude = `(function (http) {
  'use strict';

  function copyOptions(options) {
    var out = {};
    if (options && typeof options === 'object') {
      for (var key in options) {
        if (Object.prototype.hasOwnProperty.call(options, key)) out[key] = options[key];
      }
    }
    return out;
  }

  function createServer(options, listener) {
    if (typeof options === 'function') {
      listener = options;
      options = {};
    }
    var opts = copyOptions(options);
    // HTTP/3 is always TLS, and the shared Server bridge reads these flags.
    opts._tls = true;
    opts._http3 = true;
    return http.createServer(opts, listener);
  }

  return {
    createServer: createServer,
    createSecureServer: createServer,
    Server: http.Server,
    IncomingMessage: http.IncomingMessage,
    ServerResponse: http.ServerResponse,
    STATUS_CODES: http.STATUS_CODES
  };
})`

// InstallHTTP3 exposes the gode-only HTTP/3 API. Unlike Node's HTTP modules,
// this serves real QUIC/HTTP/3 over the same request and response classes as
// http, including streaming writes and the existing event loop.
func InstallHTTP3(e *env.Env) js.Object {
	httpExports := e.Exports("http")
	if _, ok := activeHTTP.Load(e); !ok {
		panic("godejs http3: http module is not installed")
	}

	api, err := e.JS.Eval(http3Prelude, "http3.js")
	if err != nil {
		panic("godejs http3: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, httpExports.V())
	if err != nil {
		panic("godejs http3: " + err.Error())
	}

	exports := e.JS.Obj().Keep()
	built.Protect()
	defer built.Unprotect()
	for _, key := range built.Object().OwnKeys() {
		exports.Set(key, built.Object().Get(key))
	}
	return exports
}

func (m *module) listenHTTP3(s *server, t target, cb js.Value) (js.Value, error) {
	if s.tlsConfig == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_TLS_INVALID_CERT",
			"An http3 server requires both 'cert' and 'key' options")
	}
	pc, err := stdnet.ListenPacket("udp", t.address)
	if err != nil {
		s.obj.SetBool("listening", false)
		m.emitAsync(s.obj, "error", m.listenError(err, t).V())
		return s.obj.V(), nil
	}

	s.packet = pc
	s.address, s.port = splitHostPort(pc.LocalAddr().String())
	s.family = familyOf(pc.LocalAddr())
	s.listening = true
	s.closed = false
	s.closeSent = false
	s.obj.SetBool("listening", true)
	s.handle = true

	srv := &http3.Server{
		TLSConfig: s.tlsConfig,
		Handler: stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			m.handle(s, w, r)
		}),
	}
	s.h3 = srv
	m.addHandle()

	go func() {
		serveErr := srv.Serve(pc)
		m.push(func() {
			if serveErr != nil && !errors.Is(serveErr, stdhttp.ErrServerClosed) && !s.closed {
				_ = m.emit(s.obj, "error", m.serverError(serveErr).V())
			}
			m.finishServerClose(s)
		})
	}()

	holder := m.e.JS.Obj()
	holder.SetFuncHidden("fire", func(this js.Object, _ []js.Value) (js.Value, error) {
		_ = m.emit(s.obj, "listening")
		if cb.Valid() && cb.IsFunction() {
			if _, err := cb.Object().Call(s.obj); err != nil {
				return js.Value{}, err
			}
		}
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(holder.Get("fire"))
	return s.obj.V(), nil
}

func (m *module) closeHTTP3(s *server) {
	if srv, ok := s.h3.(*http3.Server); ok && srv != nil {
		_ = srv.Close()
	}
	if s.packet != nil {
		_ = s.packet.Close()
		s.packet = nil
	}
	s.h3 = nil
	m.finishServerClose(s)
}

// HTTP3Implemented lists the public surface this gode-only module provides.
func HTTP3Implemented() []string {
	return []string{
		"http3.Server",
		"http3.IncomingMessage",
		"http3.ServerResponse",
		"http3.STATUS_CODES",
		"http3.createServer",
		"http3.createSecureServer",
	}
}
