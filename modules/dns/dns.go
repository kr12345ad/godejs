// Package dns implements Node's dns builtin using Go's resolver. The resolver
// calls run on goroutines and are delivered through a JS-thread pump, matching
// Node's asynchronous callback API.
package dns

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("dns", Install) }

const pumpEveryMS = 1

type module struct {
	e *env.Env

	mu     sync.Mutex
	queued []func()
	pumpID int
	active int
	server []string

	exports js.Object
	host    js.Object
}

// Install builds the dns module and its JS wrapper.
func Install(e *env.Env) js.Object {
	m := &module{e: e}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()

	m.host.SetFuncHidden("pump", m.pump)
	m.host.SetFuncHidden("lookup", m.hostLookup)
	m.host.SetFuncHidden("resolve", m.hostResolve)
	m.host.SetFuncHidden("reverse", m.hostReverse)
	m.host.SetFuncHidden("lookupService", m.hostLookupService)
	m.host.SetFuncHidden("setServers", m.hostSetServers)
	m.host.SetFuncHidden("getServers", m.hostGetServers)

	prelude, err := e.JS.Eval(prelude, "dns.js")
	if err != nil {
		panic("godejs dns: " + err.Error())
	}
	built, err := prelude.Object().Call(js.Object{}, m.host.V())
	if err != nil {
		panic("godejs dns: " + err.Error())
	}
	for _, name := range built.Object().OwnKeys() {
		m.exports.Set(name, built.Object().Get(name))
	}
	e.OnExit(func(int) { m.shutdown() })
	return m.exports
}

// Implemented lists the tracked dns API entries this package provides.
func Implemented() []string {
	return []string{
		"dns.lookup", "dns.lookupService", "dns.resolve", "dns.resolve4",
		"dns.resolve6", "dns.resolveAny", "dns.resolveCaa", "dns.resolveCname",
		"dns.resolveMx", "dns.resolveNaptr", "dns.resolveNs", "dns.resolvePtr",
		"dns.resolveSoa", "dns.resolveSrv", "dns.resolveTxt", "dns.reverse",
		"dns.getServers", "dns.setServers", "dns.Resolver",
		"dns.promises.lookup", "dns.promises.resolve", "dns.promises.resolve4",
		"dns.promises.resolve6", "dns.promises.reverse",
	}
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
	id := m.pumpID
	m.pumpID = 0
	m.active = 0
	m.mu.Unlock()
	if id != 0 {
		m.e.ClearTimer(id)
	}
}

// async runs work off the JS thread, then invokes cb on the JS thread with a
// null error and the values produced by build.
func (m *module) async(cb js.Value, host, syscall string, work func() (func() []js.Value, error)) {
	cb.Protect()
	m.addHandle()
	go func() {
		build, err := work()
		m.push(func() {
			defer cb.Unprotect()
			defer m.dropHandle()
			if err != nil {
				_, _ = cb.Object().Call(js.Object{}, m.errorValue(err, host, syscall))
				return
			}
			values := build()
			args := make([]js.Value, 0, len(values)+1)
			args = append(args, m.e.JS.Null())
			args = append(args, values...)
			_, _ = cb.Object().Call(js.Object{}, args...)
		})
	}()
}

func (m *module) resolver() *net.Resolver {
	m.mu.Lock()
	servers := append([]string(nil), m.server...)
	m.mu.Unlock()
	if len(servers) == 0 {
		return net.DefaultResolver
	}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			for _, server := range servers {
				target := server
				if _, _, err := net.SplitHostPort(target); err != nil {
					target = net.JoinHostPort(target, "53")
				}
				conn, err := (&net.Dialer{}).DialContext(ctx, network, target)
				if err == nil {
					return conn, nil
				}
			}
			return nil, errors.New("all DNS servers failed")
		},
	}
}

func (m *module) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func (m *module) hostLookup(this js.Object, args []js.Value) (js.Value, error) {
	hostname := argString(args, 0)
	family := int(arg(args, 1).Int())
	all := arg(args, 2).Bool()
	cb := arg(args, 3)
	if !cb.Valid() || !cb.IsFunction() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "callback" argument must be of type function.`)
	}
	network := "ip"
	if family == 4 {
		network = "ip4"
	} else if family == 6 {
		network = "ip6"
	}
	m.async(cb, hostname, "queryA", func() (func() []js.Value, error) {
		ctx, cancel := m.context()
		defer cancel()
		ips, err := m.resolver().LookupIP(ctx, network, hostname)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, &net.DNSError{Err: "no such host", Name: hostname, IsNotFound: true}
		}
		type ipRecord struct {
			address string
			family  int
		}
		records := make([]ipRecord, 0, len(ips))
		for _, ip := range ips {
			fam := 6
			if ip.To4() != nil {
				fam = 4
			}
			if family != 0 && fam != family {
				continue
			}
			records = append(records, ipRecord{address: ip.String(), family: fam})
		}
		if len(records) == 0 {
			return nil, &net.DNSError{Err: "no such host", Name: hostname, IsNotFound: true}
		}
		return func() []js.Value {
			if all {
				arr := m.e.JS.Arr(nil)
				for i, rec := range records {
					obj := m.e.JS.Obj()
					obj.SetStr("address", rec.address)
					obj.SetNum("family", float64(rec.family))
					arr.SetAt(i, obj.V())
				}
				return []js.Value{arr.V()}
			}
			return []js.Value{m.e.JS.Str(records[0].address), m.e.JS.Num(float64(records[0].family))}
		}, nil
	})
	return this.V(), nil
}

// resolveRecords performs one DNS record lookup and returns the Node-shaped
// JavaScript value for it.
func (m *module) resolveRecords(hostname, recordType string, ttl bool) (func() []js.Value, error) {
	ctx, cancel := m.context()
	defer cancel()
	r := m.resolver()
	recordType = strings.ToUpper(recordType)
	buildStrings := func(values []string) func() []js.Value {
		return func() []js.Value {
			arr := m.e.JS.Arr(nil)
			for i, value := range values {
				arr.SetAt(i, m.e.JS.Str(value))
			}
			return []js.Value{arr.V()}
		}
	}
	buildObjects := func(values []func(js.Object)) func() []js.Value {
		return func() []js.Value {
			arr := m.e.JS.Arr(nil)
			for i, fill := range values {
				obj := m.e.JS.Obj()
				fill(obj)
				arr.SetAt(i, obj.V())
			}
			return []js.Value{arr.V()}
		}
	}
	switch recordType {
	case "A", "AAAA":
		network := "ip4"
		if recordType == "AAAA" {
			network = "ip6"
		}
		ips, err := r.LookupIP(ctx, network, hostname)
		if err != nil {
			return nil, err
		}
		if ttl {
			objects := make([]func(js.Object), 0, len(ips))
			for _, ip := range ips {
				address := ip.String()
				objects = append(objects, func(o js.Object) {
					o.SetStr("address", address)
					o.SetNum("ttl", 0)
				})
			}
			return buildObjects(objects), nil
		}
		values := make([]string, 0, len(ips))
		for _, ip := range ips {
			values = append(values, ip.String())
		}
		return buildStrings(values), nil
	case "CNAME":
		name, err := r.LookupCNAME(ctx, hostname)
		if err != nil {
			return nil, err
		}
		return buildStrings([]string{name}), nil
	case "MX":
		mxs, err := r.LookupMX(ctx, hostname)
		if err != nil {
			return nil, err
		}
		objects := make([]func(js.Object), 0, len(mxs))
		for _, mx := range mxs {
			exchange, priority := mx.Host, mx.Pref
			objects = append(objects, func(o js.Object) {
				o.SetStr("exchange", exchange)
				o.SetNum("priority", float64(priority))
			})
		}
		return buildObjects(objects), nil
	case "NS":
		nss, err := r.LookupNS(ctx, hostname)
		if err != nil {
			return nil, err
		}
		values := make([]string, 0, len(nss))
		for _, ns := range nss {
			values = append(values, ns.Host)
		}
		return buildStrings(values), nil
	case "TXT":
		txts, err := r.LookupTXT(ctx, hostname)
		if err != nil {
			return nil, err
		}
		return func() []js.Value {
			arr := m.e.JS.Arr(nil)
			for i, value := range txts {
				inner := m.e.JS.Arr([]js.Value{m.e.JS.Str(value)})
				arr.SetAt(i, inner.V())
			}
			return []js.Value{arr.V()}
		}, nil
	case "SRV":
		_, srvs, err := r.LookupSRV(ctx, "", "", hostname)
		if err != nil {
			return nil, err
		}
		objects := make([]func(js.Object), 0, len(srvs))
		for _, srv := range srvs {
			priority, weight, port, name := srv.Priority, srv.Weight, srv.Port, srv.Target
			objects = append(objects, func(o js.Object) {
				o.SetNum("priority", float64(priority))
				o.SetNum("weight", float64(weight))
				o.SetNum("port", float64(port))
				o.SetStr("name", name)
			})
		}
		return buildObjects(objects), nil
	case "PTR":
		names, err := r.LookupAddr(ctx, hostname)
		if err != nil {
			return nil, err
		}
		return buildStrings(names), nil
	case "SOA", "CAA", "NAPTR", "ANY":
		return nil, &net.DNSError{Err: "record type not supported by this resolver", Name: hostname, IsNotFound: true}
	default:
		return nil, &net.DNSError{Err: "unknown record type " + recordType, Name: hostname, IsNotFound: true}
	}
}

func (m *module) hostResolve(this js.Object, args []js.Value) (js.Value, error) {
	hostname := argString(args, 0)
	recordType := argString(args, 1)
	ttl := arg(args, 2).Bool()
	cb := arg(args, 3)
	if !cb.Valid() || !cb.IsFunction() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "callback" argument must be of type function.`)
	}
	m.async(cb, hostname, "queryA", func() (func() []js.Value, error) {
		return m.resolveRecords(hostname, recordType, ttl)
	})
	return this.V(), nil
}

func (m *module) hostReverse(this js.Object, args []js.Value) (js.Value, error) {
	ip := argString(args, 0)
	cb := arg(args, 1)
	if !cb.Valid() || !cb.IsFunction() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "callback" argument must be of type function.`)
	}
	m.async(cb, ip, "getHostByAddr", func() (func() []js.Value, error) {
		ctx, cancel := m.context()
		defer cancel()
		names, err := m.resolver().LookupAddr(ctx, ip)
		if err != nil {
			return nil, err
		}
		return func() []js.Value {
			arr := m.e.JS.Arr(nil)
			for i, name := range names {
				arr.SetAt(i, m.e.JS.Str(name))
			}
			return []js.Value{arr.V()}
		}, nil
	})
	return this.V(), nil
}

func (m *module) hostLookupService(this js.Object, args []js.Value) (js.Value, error) {
	address := argString(args, 0)
	port := int(arg(args, 1).Int())
	cb := arg(args, 2)
	if !cb.Valid() || !cb.IsFunction() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "callback" argument must be of type function.`)
	}
	m.async(cb, address, "getnameinfo", func() (func() []js.Value, error) {
		ctx, cancel := m.context()
		defer cancel()
		names, err := m.resolver().LookupAddr(ctx, address)
		if err != nil {
			return nil, err
		}
		service := strconv.Itoa(port)
		if name, err := net.LookupPort("tcp", service); err == nil {
			service = strconv.Itoa(name)
		}
		return func() []js.Value {
			host := ""
			if len(names) > 0 {
				host = strings.TrimSuffix(names[0], ".")
			}
			return []js.Value{m.e.JS.Str(host), m.e.JS.Str(service)}
		}, nil
	})
	return this.V(), nil
}

func (m *module) hostSetServers(this js.Object, args []js.Value) (js.Value, error) {
	v := arg(args, 0)
	if !v.Valid() || !v.IsArray() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "servers" argument must be an instance of Array.`)
	}
	arr := v.Object()
	servers := make([]string, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		server := arr.At(i).Str()
		if server == "" {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "DNS server must be a non-empty string")
		}
		servers = append(servers, server)
	}
	m.mu.Lock()
	m.server = servers
	m.mu.Unlock()
	return this.V(), nil
}

func (m *module) hostGetServers(this js.Object, args []js.Value) (js.Value, error) {
	m.mu.Lock()
	servers := append([]string(nil), m.server...)
	m.mu.Unlock()
	if len(servers) == 0 {
		servers = []string{"127.0.0.1"}
	}
	arr := m.e.JS.Arr(nil)
	for i, server := range servers {
		arr.SetAt(i, m.e.JS.Str(server))
	}
	return arr.V(), nil
}

func (m *module) errorValue(err error, host, syscall string) js.Value {
	code := "ENOTFOUND"
	message := err.Error()
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			code = "ETIMEOUT"
		} else if dnsErr.IsTemporary {
			code = "EAI_AGAIN"
		} else if dnsErr.IsNotFound {
			code = "ENOTFOUND"
		} else if strings.Contains(strings.ToLower(dnsErr.Err), "server misbehaving") {
			code = "EAI_FAIL"
		}
		if dnsErr.Name != "" {
			host = dnsErr.Name
		}
	}
	if strings.Contains(strings.ToLower(message), "no such host") {
		code = "ENOTFOUND"
	}
	e := m.e.JS.Error(message)
	e.SetStr("code", code)
	e.SetStr("syscall", syscall)
	if host != "" {
		e.SetStr("hostname", host)
	}
	return e.V()
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func argString(args []js.Value, i int) string {
	v := arg(args, i)
	if v.Valid() && v.IsString() {
		return v.Str()
	}
	return ""
}

const prelude = `(function (host) {
  'use strict';

  function typeError(message) {
    var err = new TypeError(message);
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function validateHost(hostname) {
    if (typeof hostname !== 'string') throw typeError('The "hostname" argument must be of type string.');
    if (hostname.indexOf('\0') !== -1) throw typeError('The "hostname" argument must not contain null bytes.');
    return hostname;
  }

  function optionsArg(options) {
    if (typeof options === 'number') return { family: options };
    return options && typeof options === 'object' ? options : {};
  }

  function lookup(hostname, options, callback) {
    if (typeof options === 'function') { callback = options; options = {}; }
    validateHost(hostname);
    options = optionsArg(options);
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    var family = options.family === undefined ? 0 : Number(options.family);
    if (family !== 0 && family !== 4 && family !== 6) throw typeError('The "family" argument must be 0, 4, or 6.');
    host.lookup(hostname, family, !!options.all, callback);
  }

  function resolve(hostname, rrtype, callback) {
    if (typeof rrtype === 'function') { callback = rrtype; rrtype = 'A'; }
    validateHost(hostname);
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    host.resolve(hostname, String(rrtype || 'A'), false, callback);
  }

  function resolveWithTTL(hostname, options, callback) {
    var ttl = false;
    if (typeof options === 'function') { callback = options; }
    else if (options && typeof options === 'object') { ttl = !!options.ttl; }
    validateHost(hostname);
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    host.resolve(hostname, 'A', ttl, callback);
  }

  function resolve6WithTTL(hostname, options, callback) {
    var ttl = false;
    if (typeof options === 'function') { callback = options; }
    else if (options && typeof options === 'object') { ttl = !!options.ttl; }
    validateHost(hostname);
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    host.resolve(hostname, 'AAAA', ttl, callback);
  }

  function simpleResolve(type) {
    return function (hostname, callback) {
      validateHost(hostname);
      if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
      host.resolve(hostname, type, false, callback);
    };
  }

  function reverse(ip, callback) {
    if (typeof ip !== 'string') throw typeError('The "ip" argument must be of type string.');
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    host.reverse(ip, callback);
  }

  function lookupService(address, port, callback) {
    if (typeof address !== 'string') throw typeError('The "address" argument must be of type string.');
    if (typeof callback !== 'function') throw typeError('The "callback" argument must be of type function.');
    host.lookupService(address, Number(port), callback);
  }

  function setServers(servers) { host.setServers(servers); }
  function getServers() { return host.getServers(); }

  function promisesLookup(hostname, options) {
    return new Promise(function (resolvePromise, reject) {
      var all = options && typeof options === 'object' && !!options.all;
      lookup(hostname, options || {}, function (err, address, family) {
        if (err) return reject(err);
        if (all) resolvePromise(address);
        else resolvePromise({ address: address, family: family });
      });
    });
  }

  function promisify(fn) {
    return function (hostname) {
      return new Promise(function (resolvePromise, reject) {
        fn(hostname, function (err, result) {
          if (err) reject(err); else resolvePromise(result);
        });
      });
    };
  }

  var promises = {
    lookup: promisesLookup,
    resolve: promisify(resolve),
    resolve4: promisify(resolveWithTTL),
    resolve6: promisify(resolve6WithTTL),
    resolveAny: promisify(simpleResolve('ANY')),
    resolveCaa: promisify(simpleResolve('CAA')),
    resolveCname: promisify(simpleResolve('CNAME')),
    resolveMx: promisify(simpleResolve('MX')),
    resolveNaptr: promisify(simpleResolve('NAPTR')),
    resolveNs: promisify(simpleResolve('NS')),
    resolvePtr: promisify(simpleResolve('PTR')),
    resolveSoa: promisify(simpleResolve('SOA')),
    resolveSrv: promisify(simpleResolve('SRV')),
    resolveTxt: promisify(simpleResolve('TXT')),
    reverse: promisify(reverse),
    lookupService: function (address, port) {
      return new Promise(function (resolvePromise, reject) {
        lookupService(address, port, function (err, hostname, service) {
          if (err) reject(err); else resolvePromise({ hostname: hostname, service: service });
        });
      });
    }
  };

  class Resolver {
    constructor(options) { this.options = options || {}; }
    lookup(hostname, options, callback) { return lookup(hostname, options, callback); }
    resolve(hostname, rrtype, callback) { return resolve(hostname, rrtype, callback); }
    resolve4(hostname, options, callback) { return resolveWithTTL(hostname, options, callback); }
    resolve6(hostname, options, callback) { return resolve6WithTTL(hostname, options, callback); }
    resolveAny(hostname, callback) { return simpleResolve('ANY')(hostname, callback); }
    resolveCaa(hostname, callback) { return simpleResolve('CAA')(hostname, callback); }
    resolveCname(hostname, callback) { return simpleResolve('CNAME')(hostname, callback); }
    resolveMx(hostname, callback) { return simpleResolve('MX')(hostname, callback); }
    resolveNaptr(hostname, callback) { return simpleResolve('NAPTR')(hostname, callback); }
    resolveNs(hostname, callback) { return simpleResolve('NS')(hostname, callback); }
    resolvePtr(hostname, callback) { return simpleResolve('PTR')(hostname, callback); }
    resolveSoa(hostname, callback) { return simpleResolve('SOA')(hostname, callback); }
    resolveSrv(hostname, callback) { return simpleResolve('SRV')(hostname, callback); }
    resolveTxt(hostname, callback) { return simpleResolve('TXT')(hostname, callback); }
    reverse(ip, callback) { return reverse(ip, callback); }
    setServers(servers) { setServers(servers); }
    getServers() { return getServers(); }
    cancel() {}
  }

  return {
    lookup: lookup,
    resolve: resolve,
    resolve4: resolveWithTTL,
    resolve6: resolve6WithTTL,
    resolveAny: simpleResolve('ANY'),
    resolveCaa: simpleResolve('CAA'),
    resolveCname: simpleResolve('CNAME'),
    resolveMx: simpleResolve('MX'),
    resolveNaptr: simpleResolve('NAPTR'),
    resolveNs: simpleResolve('NS'),
    resolvePtr: simpleResolve('PTR'),
    resolveSoa: simpleResolve('SOA'),
    resolveSrv: simpleResolve('SRV'),
    resolveTxt: simpleResolve('TXT'),
    reverse: reverse,
    lookupService: lookupService,
    setServers: setServers,
    getServers: getServers,
    Resolver: Resolver,
    promises: promises,
    ADDRCONFIG: 1,
    V4MAPPED: 2,
    ALL: 4,
    NODATA: 'ENODATA',
    FORMERR: 'EFORMERR',
    SERVFAIL: 'ESERVFAIL',
    NOTFOUND: 'ENOTFOUND',
    NOTIMP: 'ENOTIMP',
    REFUSED: 'EREFUSED',
    BADQUERY: 'EBADQUERY',
    BADNAME: 'EBADNAME',
    BADFAMILY: 'EBADFAMILY',
    BADRESP: 'EBADRESP',
    CONNREFUSED: 'ECONNREFUSED',
    TIMEOUT: 'ETIMEOUT',
    EOF: 'EOF',
    FILE: 'EFILE',
    NOMEM: 'ENOMEM',
    DESTRUCTION: 'EDESTRUCTION',
    BADSTR: 'EBADSTR',
    BADFLAGS: 'EBADFLAGS',
    NONAME: 'ENONAME',
    BADHINTS: 'EBADHINTS',
    NOTINITIALIZED: 'ENOTINITIALIZED',
    LOADIPHLPAPI: 'ELOADIPHLPAPI',
    ADDRGETNETWORKPARAMS: 'EADDRGETNETWORKPARAMS',
    CANCELLED: 'ECANCELLED'
  };
})
`
