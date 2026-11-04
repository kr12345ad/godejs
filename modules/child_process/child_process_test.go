package child_process

import (
	"reflect"
	"strings"
	"testing"

	"godejs/internal/jstest"
)

func TestResolveCommand(t *testing.T) {
	file, argv := resolveCommand(runOptions{cmd: "echo", args: []string{"hi"}})
	if file != "echo" || !reflect.DeepEqual(argv, []string{"hi"}) {
		t.Fatalf("plain spawn resolved to %q %v", file, argv)
	}
	file, argv = resolveCommand(runOptions{cmd: "echo hi", shell: true})
	if file != "/bin/sh" || !reflect.DeepEqual(argv, []string{"-c", "echo hi"}) {
		t.Fatalf("shell spawn resolved to %q %v", file, argv)
	}
	file, argv = resolveCommand(runOptions{cmd: "echo", args: []string{"hi"}, shell: true, shellArg: "/bin/bash"})
	if file != "/bin/bash" || !reflect.DeepEqual(argv, []string{"-c", "echo hi"}) {
		t.Fatalf("custom shell resolved to %q %v", file, argv)
	}
}

func TestValidSignal(t *testing.T) {
	for _, sig := range []string{"SIGTERM", "SIGKILL", "SIGINT", "SIGUSR2"} {
		if !validSignal(sig) {
			t.Errorf("%s should be accepted", sig)
		}
	}
	for _, sig := range []string{"SIGFOO", "TERM", ""} {
		if validSignal(sig) {
			t.Errorf("%s should be rejected", sig)
		}
	}
}

func TestStripIPCEnv(t *testing.T) {
	in := []string{
		"PATH=/bin",
		"GODE_IPC_RECV_FD=3",
		"GODE_IPC_SEND_FD=4",
		"NODE_CHANNEL_FD=3",
		"NODE_CHANNEL_SERIALIZATION_MODE=json",
		"HOME=/root",
	}
	got := stripIPCEnv(in)
	want := []string{"PATH=/bin", "HOME=/root"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stripIPCEnv = %v, want %v", got, want)
	}
	// The caller's slice must not be rewritten in place.
	if in[1] != "GODE_IPC_RECV_FD=3" {
		t.Fatalf("stripIPCEnv mutated its input: %v", in)
	}
}

func TestSpawnArgValidation(t *testing.T) {
	e := jstest.New(t)
	jstest.Install(t, e, "cp", Install)

	cases := []struct{ src, want string }{
		{
			// Node stringifies array entries (including numbers and booleans)
			// instead of rejecting them.
			`(function(){ return cp.execFileSync('/bin/echo', [1, true], { encoding: 'utf8' }).trim(); })()`,
			`1 true`,
		},
		{
			`(function(){ try { cp.spawn(1); } catch (e) { return e.code + '|' + e.message; } })()`,
			`ERR_INVALID_ARG_TYPE|The "file" argument must be of type string. Received type number (1)`,
		},
		{
			`(function(){ try { cp.spawn('x', [], { timeout: -1 }); } catch (e) { return e.code + '|' + e.message; } })()`,
			`ERR_OUT_OF_RANGE|The value of "timeout" is out of range. It must be >= 0 && <= 9007199254740991. Received -1`,
		},
		{
			`(function(){ try { cp.fork('x', [], { stdio: 'ipc' }); } catch (e) { return e.code + '|' + e.message; } })()`,
			`ERR_INVALID_ARG_VALUE|The argument 'stdio' is invalid. Received 'ipc'`,
		},
	}
	for _, c := range cases {
		got := jstest.Str(t, e, c.src)
		if got != c.want {
			t.Errorf("%s\n got %q\nwant %q", c.src, got, c.want)
		}
	}
}

// TestForkStdioModes checks the fork() stdio shorthands and the requirement
// that an explicit array carries an ipc entry.
func TestForkStdioModes(t *testing.T) {
	var m *module
	cases := []struct {
		src  string
		want [3]string
	}{
		{`'ignore'`, [3]string{"ignore", "ignore", "ignore"}},
		{`'inherit'`, [3]string{"inherit", "inherit", "inherit"}},
		{`['pipe', 'ignore', 'inherit', 'ipc']`, [3]string{"pipe", "ignore", "inherit"}},
	}
	for _, c := range cases {
		e := jstest.New(t)
		m = &module{e: e}
		v := jstest.Eval(t, e, c.src)
		opts := runOptions{}
		if err := m.forkStdio(v, &opts); err != nil {
			t.Fatalf("forkStdio(%s): %v", c.src, err)
		}
		for i := 0; i < 3; i++ {
			got := "pipe"
			if opts.inherit[i] {
				got = "inherit"
			} else if opts.ignore[i] {
				got = "ignore"
			}
			if got != c.want[i] {
				t.Errorf("forkStdio(%s) fd %d = %s, want %s", c.src, i, got, c.want[i])
			}
		}
	}

	e := jstest.New(t)
	m = &module{e: e}
	opts := runOptions{}
	err := m.forkStdio(jstest.Eval(t, e, `['pipe', 'pipe', 'pipe']`), &opts)
	if err == nil || !strings.Contains(err.Error(), "IPC channel") {
		t.Fatalf("expected an ipc-channel error, got %v", err)
	}
}
