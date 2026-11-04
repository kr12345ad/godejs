package nodetest

import (
	"bytes"
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/jstest"
	"godejs/modules/assert"
	"godejs/modules/timers"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "assert", assert.Install)
	jstest.Install(t, e, "timers", timers.Install)
	jstest.Install(t, e, "test", Install)
	// Only node:test/assert are installed here, so provide a require() that
	// resolves exactly those two builtins instead of installing the world.
	e.Global().SetFunc("require", func(_ js.Object, args []js.Value) (js.Value, error) {
		spec := ""
		if len(args) > 0 {
			spec = args[0].Str()
		}
		return e.RequireFrom(e.Cwd(), spec)
	})
	return e
}

// run evaluates src and drains the event loop so the runner finishes.
func run(t *testing.T, e *env.Env, src string) string {
	t.Helper()
	if _, err := e.RunScript(src, "test.js"); err != nil {
		t.Fatalf("run script: %v", err)
	}
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	out, ok := e.Out().(*bytes.Buffer)
	if !ok {
		t.Fatal("stdout is not a buffer")
	}
	return out.String()
}

func TestPassingSkippedAndTodo(t *testing.T) {
	e := setup(t)
	out := run(t, e, `
const test = require('node:test');
const assert = require('assert');
test('passes sync', () => { assert.strictEqual(1, 1); });
test('passes async', async () => {
  await new Promise((resolve) => setTimeout(resolve, 1));
  assert.ok(true);
});
test('skipped', { skip: 'not today' }, () => { throw new Error('must not run'); });
test('todo', { todo: true });
test.skip('skipped entrypoint', () => { throw new Error('must not run'); });
`)
	for _, want := range []string{
		"TAP version 13",
		"ok 1 - passes sync",
		"ok 2 - passes async",
		"ok 3 - skipped # SKIP not today",
		"ok 4 - todo # TODO",
		"ok 5 - skipped entrypoint # SKIP",
		"1..5",
		"# tests 5",
		"# fail",
	} {
		if want == "# fail" {
			if strings.Contains(out, "# fail") {
				t.Errorf("unexpected failures in output:\n%s", out)
			}
			continue
		}
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if e.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", e.ExitCode)
	}
}

func TestFailingTestSetsExitCode(t *testing.T) {
	e := setup(t)
	out := run(t, e, `
const test = require('node:test');
const assert = require('assert');
test('fails', () => { assert.strictEqual(1, 2); });
`)
	if !strings.Contains(out, "not ok 1 - fails") {
		t.Fatalf("missing failure line:\n%s", out)
	}
	if !strings.Contains(out, "AssertionError") {
		t.Errorf("failure block does not describe the error:\n%s", out)
	}
	if e.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", e.ExitCode)
	}
}

func TestDescribeHooksAndSubtests(t *testing.T) {
	e := setup(t)
	out := run(t, e, `
const test = require('node:test');
const assert = require('assert');
const order = [];
test.before(() => order.push('before'));
test.after(() => { globalThis.__order = order.join(','); });
test.describe('suite', () => {
  test.beforeEach(() => order.push('each'));
  test.it('runs subtests', async (t) => {
    await t.test('nested', () => { assert.ok(true); });
    assert.deepStrictEqual(order, ['before', 'each']);
  });
  test.it('sees the hook again', () => {
    assert.deepStrictEqual(order, ['before', 'each', 'each']);
  });
});
`)
	for _, want := range []string{
		"ok 1 - suite > runs subtests > nested",
		"ok 2 - suite > runs subtests",
		"ok 3 - suite > sees the hook again",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if e.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0\n%s", e.ExitCode, out)
	}
	if got := jstest.Str(t, e, `globalThis.__order`); got != "before,each,each" {
		t.Errorf("hook order = %q, want %q", got, "before,each,each")
	}
}

func TestMockAPI(t *testing.T) {
	e := setup(t)
	out := run(t, e, `
const test = require('node:test');
const assert = require('assert');
test('mocks', () => {
  const add = test.mock.fn((a, b) => a + b);
  assert.strictEqual(add(1, 2), 3);
  assert.strictEqual(add.mock.callCount(), 1);
  assert.strictEqual(add.mock.calls[0].arguments[0], 1);
  assert.strictEqual(add.mock.results[0].value, 3);
  add.mock.resetCalls();
  assert.strictEqual(add.mock.callCount(), 0);

  const target = { greet() { return 'real'; } };
  const greeted = test.mock.method(target, 'greet', () => 'mocked');
  assert.strictEqual(target.greet(), 'mocked');
  assert.strictEqual(greeted.mock.callCount(), 1);
  test.mock.restoreAll();
  assert.strictEqual(target.greet(), 'real');
});
`)
	if !strings.Contains(out, "ok 1 - mocks") {
		t.Fatalf("mock test did not pass:\n%s", out)
	}
	if e.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", e.ExitCode)
	}
}

func TestPlanWaitsForAssertions(t *testing.T) {
	e := setup(t)
	out := run(t, e, `
const test = require('node:test');
test('planned', (t) => {
  t.plan(2);
  setTimeout(() => {
    t.assert.ok(true);
    t.assert.ok(true);
  }, 1);
});
`)
	if !strings.Contains(out, "ok 1 - planned") {
		t.Fatalf("planned test did not pass:\n%s", out)
	}
}

func TestTestModuleIsCallableAndDestructurable(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof require('node:test')`), "function")
	jstest.Equal(t, jstest.Str(t, e, `String(require('node:test') === require('node:test').test)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
const { describe, it, mock, beforeEach } = require('node:test');
[typeof describe, typeof it, typeof mock.fn, typeof beforeEach].join(',')`), "function,function,function,function")
}

func TestImplementedTracked(t *testing.T) {
	if len(Implemented()) == 0 {
		t.Fatal("Implemented() is empty")
	}
}
