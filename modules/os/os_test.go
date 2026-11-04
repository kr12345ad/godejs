package os

import (
	"encoding/json"
	stdos "os"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "os", Install)
	return e
}

func TestStaticNames(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `os.EOL`), "\n")
	jstest.Equal(t, jstest.Str(t, e, `os.platform()`), "linux")
	jstest.Equal(t, jstest.Str(t, e, `os.type()`), "Linux")
	jstest.Equal(t, jstest.Str(t, e, `os.endianness()`), "LE")
	jstest.Equal(t, jstest.Str(t, e, `os.devNull`), "/dev/null")
	jstest.Equal(t, jstest.Str(t, e, `os.arch()`), archName())
	jstest.Equal(t, jstest.Str(t, e, `typeof os.release() === "string"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof os.version() === "string"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof os.machine() === "string"`), "true")
}

func TestSystemInfo(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `os.hostname().length > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.homedir().length > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.tmpdir().startsWith("/")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.tmpdir() === "/" || !os.tmpdir().endsWith("/")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.totalmem() > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.freemem() > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.availableParallelism() >= 1`), "true")
	jstest.Equal(t, jstest.Str(t, e, `os.uptime() > 0`), "true")
}

func TestCpus(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `Array.isArray(os.cpus()) && os.cpus().length >= 1`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof os.cpus()[0].model === "string" && typeof os.cpus()[0].speed === "number"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(Object.keys(os.cpus()[0].times))`), `["user","nice","sys","idle","irq"]`)
	jstest.Equal(t, jstest.Str(t, e, `os.cpus()[0].times.user >= 0 &&
		os.cpus()[0].times.nice >= 0 &&
		os.cpus()[0].times.sys >= 0 &&
		os.cpus()[0].times.idle >= 0 &&
		os.cpus()[0].times.irq >= 0`), "true")
}

func TestLoadavg(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `os.loadavg().length === 3 && os.loadavg().every(n => typeof n === "number")`), "true")
}

func TestNetworkInterfaces(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `Object.prototype.toString.call(os.networkInterfaces()) === "[object Object]"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `Object.keys(os.networkInterfaces()).every(function (name) {
		var value = os.networkInterfaces()[name];
		if (value === null) return true;
		return Array.isArray(value) && value.every(function (a) {
			return typeof a.address === "string" &&
				(typeof a.netmask === "string" || a.netmask === null) &&
				(a.family === "IPv4" || a.family === "IPv6") &&
				typeof a.mac === "string" &&
				typeof a.internal === "boolean" &&
				typeof a.cidr === "string";
		});
	})`), "true")
}

func TestUserInfo(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(Object.keys(os.userInfo()))`), `["uid","gid","username","homedir","shell"]`)
	jstest.Equal(t, jstest.Str(t, e, `typeof os.userInfo().uid === "number" && typeof os.userInfo().gid === "number"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof os.userInfo().username === "string" && typeof os.userInfo().homedir === "string" && typeof os.userInfo().shell === "string"`), "true")
}

func TestConstants(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `os.constants.signals.SIGTERM`), "15")
	jstest.Equal(t, jstest.Str(t, e, `os.constants.signals.SIGKILL`), "9")
	jstest.Equal(t, jstest.Str(t, e, `os.constants.errno.ENOENT`), "2")
	jstest.Equal(t, jstest.Str(t, e, `os.constants.errno.EACCES`), "13")
}

func TestPriority(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof os.getPriority()`), "number")
	jstest.Equal(t, jstest.Str(t, e, `typeof os.setPriority(os.getPriority())`), "undefined")
	jstest.Equal(t, jstest.Str(t, e, `os.getPriority() === os.getPriority()`), "true")
}

func TestImplementedMatchesBaseline(t *testing.T) {
	data, err := stdos.ReadFile("../../parity/baseline/os.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	tracked := map[string]bool{}
	for _, k := range baseline {
		tracked[k] = true
	}
	got := Implemented()
	if len(got) != len(baseline) {
		t.Fatalf("Implemented has %d keys, baseline has %d", len(got), len(baseline))
	}
	seen := map[string]bool{}
	for _, k := range got {
		if !tracked[k] {
			t.Fatalf("Implemented lists untracked key %q", k)
		}
		if seen[k] {
			t.Fatalf("Implemented repeats %q", k)
		}
		seen[k] = true
	}
}
