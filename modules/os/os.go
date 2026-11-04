// Package os implements Node's os builtin on top of the Go runtime and the
// Linux /proc filesystem.
package os

import (
	"fmt"
	"math"
	"net"
	stdos "os"
	gouser "os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("os", Install)
	// Node still exposes the legacy `constants` builtin as os.constants.
	env.Register("constants", func(e *env.Env) js.Object {
		return e.Exports("os").Get("constants").Object()
	})
}

// Install builds the os module's exports object.
func Install(e *env.Env) js.Object {
	c := e.JS
	m := c.Obj()
	// Node reads the environment through process.env, which JavaScript code
	// can change at any time; os.tmpdir()/os.homedir() must observe those
	// changes, so look the value up on every call instead of caching Go's
	// process environment.
	envGet := func(key string) string {
		proc := e.Global().Get("process")
		if !proc.IsObject() {
			return ""
		}
		envv := proc.Object().Get("env")
		if !envv.IsObject() {
			return ""
		}
		v := envv.Object().Get(key)
		if !v.IsString() {
			return ""
		}
		return v.Str()
	}

	m.SetStr("EOL", "\n")
	m.SetStr("devNull", devNull())

	m.SetFunc("arch", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(archName()), nil
	})
	m.SetFunc("platform", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(platformName()), nil
	})
	m.SetFunc("type", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(typeName()), nil
	})
	m.SetFunc("release", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(releaseName()), nil
	})
	m.SetFunc("version", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(readTrim("/proc/sys/kernel/version")), nil
	})
	m.SetFunc("machine", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(machineName()), nil
	})
	m.SetFunc("hostname", func(js.Object, []js.Value) (js.Value, error) {
		name, err := stdos.Hostname()
		if err != nil {
			return js.Value{}, &js.Error{Name: "Error", Message: err.Error()}
		}
		return c.Str(name), nil
	})
	m.SetFunc("homedir", func(js.Object, []js.Value) (js.Value, error) {
		// The internal os binding can be monkey-patched (Node's tests do this);
		// honor a patched getHomeDirectory and surface its ctx as a SystemError.
		if fn := m.Get("getHomeDirectory"); fn.IsFunction() {
			ctx := c.Obj()
			if _, err := fn.Object().Call(js.Object{}, ctx.V()); err != nil {
				return js.Value{}, err
			}
			if syscall := ctx.Get("syscall").Str(); syscall != "" {
				msg := fmt.Sprintf("A system error occurred: %s returned %s (%s)",
					syscall, ctx.Get("code").Str(), ctx.Get("message").Str())
				return js.Value{}, js.Throw("Error", msg)
			}
		}
		return c.Str(homedir(envGet)), nil
	})
	m.SetFunc("tmpdir", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(tmpdir(envGet)), nil
	})
	m.SetFunc("totalmem", func(js.Object, []js.Value) (js.Value, error) {
		return c.Num(float64(memInfo("MemTotal"))), nil
	})
	m.SetFunc("freemem", func(js.Object, []js.Value) (js.Value, error) {
		total := memInfo("MemAvailable")
		if total == 0 {
			total = memInfo("MemFree")
		}
		return c.Num(float64(total)), nil
	})
	m.SetFunc("availableParallelism", func(js.Object, []js.Value) (js.Value, error) {
		return c.Num(float64(runtime.NumCPU())), nil
	})
	m.SetFunc("cpus", func(js.Object, []js.Value) (js.Value, error) {
		return cpus(c), nil
	})
	m.SetFunc("loadavg", func(js.Object, []js.Value) (js.Value, error) {
		loads := loadAvg()
		vals := make([]js.Value, 3)
		for i := range loads {
			vals[i] = c.Num(loads[i])
		}
		return c.Arr(vals).V(), nil
	})
	m.SetFunc("uptime", func(js.Object, []js.Value) (js.Value, error) {
		return c.Num(uptime()), nil
	})
	m.SetFunc("endianness", func(js.Object, []js.Value) (js.Value, error) {
		return c.Str(endianness()), nil
	})
	m.SetFunc("networkInterfaces", func(js.Object, []js.Value) (js.Value, error) {
		return networkInterfaces(c), nil
	})
	m.SetFunc("_userInfo", func(js.Object, []js.Value) (js.Value, error) {
		info, err := userInfo(c)
		if err != nil {
			return js.Value{}, err
		}
		return info.V(), nil
	})
	// libuv rejects negative pids and reports the failure through the
	// SystemError contract, so the host functions return a plain result object
	// and the JavaScript wrapper below builds the error with its `info` bag.
	priorityResult := func(err error) js.Object {
		o := c.Obj()
		if err == nil {
			o.SetBool("ok", true)
			return o
		}
		code := ""
		if errno, ok := err.(syscall.Errno); ok {
			code = errnoNames[errno]
		}
		o.SetBool("ok", false)
		o.SetStr("code", code)
		o.SetStr("message", errnoMessage(code))
		return o
	}
	m.SetFunc("_getPriority", func(_ js.Object, args []js.Value) (js.Value, error) {
		pid := 0
		if len(args) > 0 && args[0].IsNumber() {
			pid = int(args[0].Int())
		}
		if pid < 0 {
			return priorityResult(syscall.EINVAL).V(), nil
		}
		prio, err := syscall.Getpriority(syscall.PRIO_PROCESS, pid)
		if err != nil {
			return priorityResult(err).V(), nil
		}
		// The raw getpriority(2) value is 20 - nice (so errors stay negative
		// and lower means higher priority); Node reports the nice value.
		prio = 20 - prio
		o := priorityResult(nil)
		o.SetNum("priority", float64(prio))
		return o.V(), nil
	})
	m.SetFunc("_setPriority", func(_ js.Object, args []js.Value) (js.Value, error) {
		pid, priority := 0, 0
		if len(args) > 0 && args[0].IsNumber() {
			pid = int(args[0].Int())
		}
		if len(args) > 1 && args[1].IsNumber() {
			priority = int(args[1].Int())
		}
		if pid < 0 {
			return priorityResult(syscall.EINVAL).V(), nil
		}
		if err := syscall.Setpriority(syscall.PRIO_PROCESS, pid, priority); err != nil {
			return priorityResult(err).V(), nil
		}
		return priorityResult(nil).V(), nil
	})

	cons := c.Obj()
	signals := c.Obj()
	for name, n := range signalNumbers {
		signals.SetNum(name, float64(n))
	}
	cons.SetObj("signals", signals)
	errnos := c.Obj()
	for name, n := range errnoNumbers {
		errnos.SetNum(name, float64(n))
	}
	cons.SetObj("errno", errnos)

	priorities := c.Obj()
	priorities.SetNum("PRIORITY_LOW", 19)
	priorities.SetNum("PRIORITY_BELOW_NORMAL", 10)
	priorities.SetNum("PRIORITY_NORMAL", 0)
	priorities.SetNum("PRIORITY_ABOVE_NORMAL", -5)
	priorities.SetNum("PRIORITY_HIGH", -14)
	priorities.SetNum("PRIORITY_HIGHEST", -20)
	cons.SetObj("priority", priorities)

	dlopen := c.Obj()
	dlopen.SetNum("RTLD_LAZY", 1)
	dlopen.SetNum("RTLD_NOW", 2)
	dlopen.SetNum("RTLD_GLOBAL", 0x100)
	dlopen.SetNum("RTLD_LOCAL", 0)
	cons.SetObj("dlopen", dlopen)

	// Node exposes the libuv errno table as constants.uv and aliases it as
	// constants.libuv; reuse the errno names where libuv has one.
	libuv := c.Obj()
	for name, n := range errnoNumbers {
		libuv.SetNum("UV_"+name, float64(n))
	}
	libuv.SetNum("UV_E2BIG", 7)
	libuv.SetNum("UV_EOF", -4095)
	libuv.SetNum("UV_ENODEV", 19)
	cons.SetObj("uv", libuv)
	cons.SetObj("libuv", libuv)

	m.SetObj("constants", cons)

	// Node defines EOL, devNull, the frozen signal table, the coercible
	// function wrappers and the priority/userInfo validators in JavaScript.
	// Doing the same here keeps property attributes and SystemError `info`
	// bags exactly as the official tests expect them.
	if post, err := c.Eval(osPostlude, "[os:postlude]"); err == nil && post.Valid() && post.IsFunction() {
		if _, cerr := post.Object().Call(js.Object{}, m.V()); cerr != nil {
			panic("godejs os postlude: " + cerr.Error())
		}
	}
	return m
}

// osPostlude rewrites the parts of the os module that Node implements in
// JavaScript: non-writable EOL/devNull, the frozen signal constants, the
// Symbol.toPrimitive coercions and the validating priority/userInfo wrappers.
const osPostlude = `(function (os) {
  'use strict';
  var defineProperty = Object.defineProperty;
  defineProperty(os, 'EOL', { configurable: true, enumerable: true, writable: false, value: '\n' });
  defineProperty(os, 'devNull', { configurable: true, enumerable: true, writable: false, value: '/dev/null' });
  if (os.constants && os.constants.signals) Object.freeze(os.constants.signals);
  // Keep the newline independent of how the literal above was escaped.
  defineProperty(os, 'EOL', { configurable: true, enumerable: true, writable: false, value: String.fromCharCode(10) });

  var names = ['arch', 'platform', 'tmpdir', 'endianness', 'release', 'type',
    'version', 'machine', 'hostname', 'homedir', 'totalmem', 'freemem',
    'availableParallelism', 'uptime'];
  for (var i = 0; i < names.length; i++) {
    (function (fn) {
      if (typeof fn !== 'function') return;
      try {
        defineProperty(fn, Symbol.toPrimitive, {
          configurable: true,
          value: function (hint) { return fn(); }
        });
      } catch (e) {}
    })(os[names[i]]);
  }

  function describe(value) {
    if (value === null) return 'null';
    var t = typeof value;
    if (t === 'undefined') return 'undefined';
    if (t === 'string') return "type string ('" + value + "')";
    if (t === 'number' || t === 'boolean' || t === 'bigint') return 'type ' + t + ' (' + String(value) + ')';
    if (t === 'function') return 'function ' + (value.name || '<anonymous>');
    if (Array.isArray(value)) return 'an instance of Array';
    if (value && value.constructor && value.constructor.name) return 'an instance of ' + value.constructor.name;
    return 'an instance of Object';
  }
  function invalidArgType(name, value) {
    var err = new TypeError('The "' + name + '" argument must be of type number. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }
  function outOfRange(name, value, min, max) {
    var err = new RangeError('The value of "' + name + '" is out of range. It must be >= ' + min + ' && <= ' + max + '. Received ' + value);
    err.code = 'ERR_OUT_OF_RANGE';
    return err;
  }
  function validateInt32(value, name, min, max) {
    if (typeof value !== 'number') throw invalidArgType(name, value);
    if (min === undefined) { min = -2147483648; max = 2147483647; }
    if (!isFinite(value) || Math.floor(value) !== value || value < min || value > max) {
      throw outOfRange(name, value, min, max);
    }
    return value;
  }
  function systemError(syscall, result) {
    var err = new Error('A system error occurred: ' + syscall + ' returned ' + result.code + ' (' + result.message + ')');
    err.name = 'SystemError';
    err.code = 'ERR_SYSTEM_ERROR';
    err.syscall = syscall;
    err.info = { code: result.code, message: result.message, syscall: syscall };
    return err;
  }

  var getPriorityHost = os._getPriority;
  var setPriorityHost = os._setPriority;
  var userInfoHost = os._userInfo;
  delete os._getPriority;
  delete os._setPriority;
  delete os._userInfo;

  function getPriority(pid) {
    if (pid === undefined) pid = 0;
    else pid = validateInt32(pid, 'pid');
    var r = getPriorityHost(pid);
    if (!r.ok) throw systemError('uv_os_getpriority', r);
    return r.priority;
  }
  function setPriority(pid, priority) {
    if (priority === undefined) { priority = pid; pid = 0; }
    pid = validateInt32(pid, 'pid');
    priority = validateInt32(priority, 'priority', -20, 19);
    var r = setPriorityHost(pid, priority);
    if (!r.ok) throw systemError('uv_os_setpriority', r);
  }
  function toBuffer(value) {
    if (typeof Buffer === 'function') return Buffer.from(value, 'utf8');
    return value;
  }
  function userInfo(options) {
    if (options === null || typeof options !== 'object') options = null;
    var encoding;
    if (options !== null) encoding = options.encoding;
    var raw = userInfoHost();
    if (encoding === 'buffer') {
      return {
        uid: raw.uid,
        gid: raw.gid,
        username: toBuffer(raw.username),
        homedir: toBuffer(raw.homedir),
        shell: raw.shell === null ? null : toBuffer(raw.shell)
      };
    }
    return raw;
  }
  os.getPriority = getPriority;
  os.setPriority = setPriority;
  os.userInfo = userInfo;
  return os;
})`

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"os.EOL",
		"os.arch",
		"os.availableParallelism",
		"os.constants",
		"os.cpus",
		"os.devNull",
		"os.endianness",
		"os.freemem",
		"os.getPriority",
		"os.homedir",
		"os.hostname",
		"os.loadavg",
		"os.machine",
		"os.networkInterfaces",
		"os.platform",
		"os.release",
		"os.setPriority",
		"os.tmpdir",
		"os.totalmem",
		"os.type",
		"os.uptime",
		"os.userInfo",
		"os.version",
	}
}

func archName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	case "arm64":
		return "arm64"
	case "arm":
		return "arm"
	case "ppc64", "ppc64le":
		return "ppc64"
	case "s390x":
		return "s390x"
	case "riscv64":
		return "riscv64"
	case "mips", "mipsle":
		return "mips"
	case "mips64", "mips64le":
		return "mips64el"
	case "loong64":
		return "loong64"
	default:
		return runtime.GOARCH
	}
}

func platformName() string {
	switch runtime.GOOS {
	case "windows":
		return "win32"
	case "solaris":
		return "sunos"
	default:
		return runtime.GOOS
	}
}

func typeName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows_NT"
	case "darwin":
		return "Darwin"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	case "netbsd":
		return "NetBSD"
	case "solaris":
		return "SunOS"
	case "aix":
		return "AIX"
	case "android":
		return "Android"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}

func releaseName() string {
	if s := readTrim("/proc/sys/kernel/osrelease"); s != "" {
		return s
	}
	return runtime.GOOS
}

func machineName() string {
	if s := readTrim("/proc/sys/kernel/arch"); s != "" {
		return s
	}
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "i686"
	case "arm64":
		return "aarch64"
	case "arm":
		return "armv7l"
	default:
		return runtime.GOARCH
	}
}

func endianness() string {
	switch runtime.GOARCH {
	case "s390x", "ppc64", "mips", "mips64":
		return "BE"
	default:
		return "LE"
	}
}

func devNull() string {
	if runtime.GOOS == "windows" {
		return "\\\\.\\nul"
	}
	return "/dev/null"
}

func homedir(getenv func(string) string) string {
	if home := getenv("HOME"); home != "" {
		return home
	}
	if u, err := gouser.Current(); err == nil && u.HomeDir != "" {
		return u.HomeDir
	}
	return ""
}

func tmpdir(getenv func(string) string) string {
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		if dir := getenv(key); dir != "" {
			dir = strings.TrimRight(dir, "/")
			if dir == "" {
				return "/"
			}
			return dir
		}
	}
	return "/tmp"
}

func uptime() float64 {
	fields, err := readFields("/proc/uptime")
	if err != nil || len(fields) == 0 {
		return 0
	}
	f, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || f < 0 {
		return 0
	}
	return f
}

func loadAvg() [3]float64 {
	var out [3]float64
	fields, err := readFields("/proc/loadavg")
	if err != nil {
		return out
	}
	for i := 0; i < 3 && i < len(fields); i++ {
		f, err := strconv.ParseFloat(fields[i], 64)
		if err == nil && !math.IsNaN(f) {
			out[i] = f
		}
	}
	return out
}

func memInfo(key string) int64 {
	data, err := readFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		name, rest, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) != key {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0
		}
		return n * 1024
	}
	return 0
}

type cpuInfo struct {
	model string
	mhz   float64
}

func cpus(c *js.Context) js.Value {
	infos := readCPUInfo()
	times := readCPUTimes()
	if len(infos) == 0 {
		n := runtime.NumCPU()
		if n < 1 {
			n = 1
		}
		infos = make([]cpuInfo, n)
		for i := range infos {
			infos[i] = cpuInfo{model: archName()}
		}
	}
	vals := make([]js.Value, 0, len(infos))
	for i, info := range infos {
		entry := c.Obj()
		model := info.model
		if model == "" {
			model = archName()
		}
		entry.SetStr("model", model)
		entry.SetNum("speed", info.mhz)
		t := times[i]
		tm := c.Obj()
		tm.SetNum("user", t[0])
		tm.SetNum("nice", t[1])
		tm.SetNum("sys", t[2])
		tm.SetNum("idle", t[3])
		tm.SetNum("irq", t[4])
		entry.SetObj("times", tm)
		vals = append(vals, entry.V())
	}
	return c.Arr(vals).V()
}

func readCPUInfo() []cpuInfo {
	data, err := readFile("/proc/cpuinfo")
	if err != nil {
		return nil
	}
	var out []cpuInfo
	var cur cpuInfo
	seen := false
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "processor":
			if seen {
				out = append(out, cur)
			}
			cur = cpuInfo{}
			seen = true
		case "model name", "Model", "cpu model":
			if cur.model == "" {
				cur.model = value
			}
		case "cpu MHz", "clock":
			if cur.mhz == 0 {
				if f, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "MHz")), 64); err == nil {
					cur.mhz = f
				}
			}
		}
	}
	if seen {
		out = append(out, cur)
	}
	return out
}

// readCPUTimes returns per-CPU jiffies converted to milliseconds, matching
// Node's os.cpus()[].times (user, nice, sys, idle, irq).
func readCPUTimes() map[int][5]float64 {
	out := map[int][5]float64{}
	data, err := readFile("/proc/stat")
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		rest := line[3:]
		if rest == "" || rest[0] == ' ' {
			continue // aggregate "cpu" line
		}
		idx := 0
		for idx < len(rest) && rest[idx] >= '0' && rest[idx] <= '9' {
			idx++
		}
		if idx == 0 {
			continue
		}
		n, err := strconv.Atoi(rest[:idx])
		if err != nil {
			continue
		}
		fields := strings.Fields(rest[idx:])
		var jiffies [5]float64
		// user, nice, system, idle, iowait, irq, softirq, steal
		for i := 0; i < len(fields) && i < 8; i++ {
			f, err := strconv.ParseFloat(fields[i], 64)
			if err != nil {
				continue
			}
			switch i {
			case 0:
				jiffies[0] = f
			case 1:
				jiffies[1] = f
			case 2:
				jiffies[2] = f
			case 3:
				jiffies[3] = f
			case 5, 6:
				jiffies[4] += f
			}
		}
		// /proc/stat reports USER_HZ ticks (100 on Linux); Node exposes ms.
		for i := range jiffies {
			jiffies[i] *= 10
		}
		out[n] = jiffies
	}
	return out
}

func networkInterfaces(c *js.Context) js.Value {
	out := c.Obj()
	ifaces, err := net.Interfaces()
	if err != nil {
		return out.V()
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			out.Set(iface.Name, c.Null())
			continue
		}
		entries := make([]js.Value, 0, len(addrs))
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP
			family := "IPv6"
			if ip.To4() != nil {
				family = "IPv4"
				ip = ip.To4()
			}
			mask := ipnet.Mask
			if family == "IPv4" && len(mask) == net.IPv6len {
				mask = mask[12:]
			}
			entry := c.Obj()
			entry.SetStr("address", ip.String())
			entry.SetStr("netmask", net.IP(mask).String())
			entry.SetStr("family", family)
			entry.SetStr("mac", iface.HardwareAddr.String())
			entry.SetBool("internal", iface.Flags&net.FlagLoopback != 0 || ip.IsLoopback())
			entry.SetStr("cidr", ipnet.String())
			if family == "IPv6" {
				entry.SetNum("scopeid", 0)
				if ip.IsLinkLocalUnicast() {
					entry.SetStr("zoneid", iface.Name)
				}
			}
			entries = append(entries, entry.V())
		}
		if len(entries) == 0 {
			out.Set(iface.Name, c.Null())
			continue
		}
		out.Set(iface.Name, c.Arr(entries).V())
	}
	return out.V()
}

func userInfo(c *js.Context) (js.Object, error) {
	u, err := gouser.Current()
	if err != nil {
		return c.Obj(), &js.Error{Name: "Error", Message: err.Error()}
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	shell := stdos.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	info := c.Obj()
	info.SetNum("uid", float64(uid))
	info.SetNum("gid", float64(gid))
	info.SetStr("username", u.Username)
	info.SetStr("homedir", u.HomeDir)
	info.SetStr("shell", shell)
	return info, nil
}

func intArg(v js.Value, name string) (int, error) {
	if !v.IsNumber() {
		return 0, typeErr(fmt.Sprintf("The %q argument must be of type number. Received %s", name, describe(v)))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, rangeErr(fmt.Sprintf("The value of %q is out of range. It must be an integer. Received %v", name, f))
	}
	return int(f), nil
}

// errnoMessage renders libuv's strerror text for the errno names os exposes.
func errnoMessage(code string) string {
	switch code {
	case "EINVAL":
		return "invalid argument"
	case "EPERM":
		return "operation not permitted"
	case "EACCES":
		return "permission denied"
	case "ESRCH":
		return "no such process"
	case "ENOSYS":
		return "function not implemented"
	}
	return "unknown error"
}

func describe(v js.Value) string {
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		if v.Bool() {
			return "true"
		}
		return "false"
	case v.IsNumber():
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case v.IsString():
		return "'" + v.Str() + "'"
	case v.IsArray():
		return "an instance of Array"
	case v.IsObject():
		return "an instance of Object"
	default:
		return "an unsupported value"
	}
}

func typeErr(msg string) *js.Error {
	return &js.Error{Name: "TypeError", Code: "ERR_INVALID_ARG_TYPE", Message: msg}
}

func rangeErr(msg string) *js.Error {
	return &js.Error{Name: "RangeError", Code: "ERR_OUT_OF_RANGE", Message: msg}
}

func sysError(err error, call string) *js.Error {
	code := ""
	if errno, ok := err.(syscall.Errno); ok {
		code = errnoNames[errno]
	}
	msg := err.Error() + ", " + call
	if code != "" {
		msg = code + ": " + msg
	}
	return &js.Error{Name: "Error", Code: code, Message: msg}
}

func readFile(path string) ([]byte, error) {
	return stdos.ReadFile(path)
}

func readTrim(path string) string {
	b, err := stdos.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func readFields(path string) ([]string, error) {
	b, err := stdos.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(b)), nil
}

var errnoNames = map[syscall.Errno]string{
	syscall.EPERM:   "EPERM",
	syscall.ENOENT:  "ENOENT",
	syscall.ESRCH:   "ESRCH",
	syscall.EINTR:   "EINTR",
	syscall.EIO:     "EIO",
	syscall.EBADF:   "EBADF",
	syscall.EACCES:  "EACCES",
	syscall.EEXIST:  "EEXIST",
	syscall.ENOTDIR: "ENOTDIR",
	syscall.EISDIR:  "EISDIR",
	syscall.EINVAL:  "EINVAL",
	syscall.EMFILE:  "EMFILE",
	syscall.ENOSPC:  "ENOSPC",
	syscall.EPIPE:   "EPIPE",
}

var signalNumbers = map[string]int{
	"SIGHUP":    1,
	"SIGINT":    2,
	"SIGQUIT":   3,
	"SIGILL":    4,
	"SIGTRAP":   5,
	"SIGABRT":   6,
	"SIGBUS":    7,
	"SIGFPE":    8,
	"SIGKILL":   9,
	"SIGUSR1":   10,
	"SIGSEGV":   11,
	"SIGUSR2":   12,
	"SIGPIPE":   13,
	"SIGALRM":   14,
	"SIGTERM":   15,
	"SIGSTKFLT": 16,
	"SIGCHLD":   17,
	"SIGCONT":   18,
	"SIGSTOP":   19,
	"SIGTSTP":   20,
	"SIGTTIN":   21,
	"SIGTTOU":   22,
	"SIGURG":    23,
	"SIGXCPU":   24,
	"SIGXFSZ":   25,
	"SIGVTALRM": 26,
	"SIGPROF":   27,
	"SIGWINCH":  28,
	"SIGIO":     29,
	"SIGPWR":    30,
	"SIGSYS":    31,
}

var errnoNumbers = map[string]int{
	"EPERM":   1,
	"ENOENT":  2,
	"ESRCH":   3,
	"EINTR":   4,
	"EIO":     5,
	"ENXIO":   6,
	"E2BIG":   7,
	"ENOEXEC": 8,
	"EBADF":   9,
	"ECHILD":  10,
	"EAGAIN":  11,
	"ENOMEM":  12,
	"EACCES":  13,
	"EFAULT":  14,
	"EBUSY":   16,
	"EEXIST":  17,
	"EXDEV":   18,
	"ENODEV":  19,
	"ENOTDIR": 20,
	"EISDIR":  21,
	"EINVAL":  22,
	"ENFILE":  23,
	"EMFILE":  24,
	"ENOTTY":  25,
	"EFBIG":   27,
	"ENOSPC":  28,
	"ESPIPE":  29,
	"EROFS":   30,
	"EMLINK":  31,
	"EPIPE":   32,
	"EDOM":    33,
	"ERANGE":  34,
}
