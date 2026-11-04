package async_hooks

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/timers"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	// timers installs the setTimeout/setInterval/setImmediate globals that the
	// module wraps for id and store propagation.
	jstest.Install(t, e, "timers", timers.Install)
	jstest.Install(t, e, "async_hooks", Install)
	return e
}

func TestIdsAndResource(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `typeof async_hooks.executionAsyncId() + ':' +
		typeof async_hooks.triggerAsyncId() + ':' +
		typeof async_hooks.executionAsyncResource()`), "number:number:object")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const resource = new async_hooks.AsyncResource('X');
		return resource.type + ':' + (typeof resource.asyncId());
	})()`), "X:number")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const resource = new async_hooks.AsyncResource('X');
		return String(resource.runInAsyncScope(() => 42));
	})()`), "42")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const resource = new async_hooks.AsyncResource('X');
		const bound = resource.bind(function (a) { return a + 1; });
		return String(bound(1));
	})()`), "2")

	// The execution id changes while a scope is entered.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const resource = new async_hooks.AsyncResource('X');
		const outer = async_hooks.executionAsyncId();
		const inner = resource.runInAsyncScope(() => async_hooks.executionAsyncId());
		return String(inner !== outer) + ':' + String(async_hooks.executionAsyncId() === outer);
	})()`), "true:true")
}

func TestCreateHookTracksTimers(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const types = [];
		const hook = async_hooks.createHook({
			init(id, type) { types.push(type); }
		});
		hook.enable();
		const timer = setTimeout(() => {}, 5);
		clearTimeout(timer);
		hook.disable();
		return types.join(',');
	})()`), "Timeout")

	// init must not fire once the hook is disabled.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const types = [];
		const hook = async_hooks.createHook({ init(id, type) { types.push(type); } });
		hook.enable();
		hook.disable();
		const timer = setTimeout(() => {}, 5);
		clearTimeout(timer);
		return String(types.length);
	})()`), "0")
}

func TestDestroyAfterCallback(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		globalThis.__destroyed = [];
		const hook = async_hooks.createHook({
			destroy(id) { globalThis.__destroyed.push(id); }
		});
		hook.enable();
		setTimeout(() => {}, 0);
		return 'queued';
	})()`), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__destroyed.length >= 1)`), "true")
}

func TestAsyncLocalStorageSync(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const als = new async_hooks.AsyncLocalStorage();
		const inside = als.run('v', () => als.getStore());
		return inside + ':' + String(als.getStore());
	})()`), "v:undefined")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const als = new async_hooks.AsyncLocalStorage();
		als.enterWith('entered');
		const current = als.getStore();
		als.exit(() => {});
		return current + ':' + als.getStore();
	})()`), "entered:entered")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const als = new async_hooks.AsyncLocalStorage();
		const seen = [];
		als.run('outer', () => {
			seen.push(als.getStore());
			als.run('inner', () => seen.push(als.getStore()));
			seen.push(als.getStore());
		});
		return seen.join(',');
	})()`), "outer,inner,outer")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const als = new async_hooks.AsyncLocalStorage();
		const snapshot = async_hooks.AsyncLocalStorage.snapshot();
		return String(typeof snapshot);
	})()`), "function")
}

func TestCreateHookValidatesCallbacks(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const names = ['init', 'before', 'after', 'destroy', 'promiseResolve'];
		const out = [];
		for (const name of names) {
			try { async_hooks.createHook({ [name]: 1 }); out.push('no-error'); }
			catch (err) { out.push(err.name + ':' + err.code + ':' + err.message); }
		}
		return out.join('|');
	})()`), "TypeError:ERR_ASYNC_CALLBACK:hook.init must be a function|TypeError:ERR_ASYNC_CALLBACK:hook.before must be a function|TypeError:ERR_ASYNC_CALLBACK:hook.after must be a function|TypeError:ERR_ASYNC_CALLBACK:hook.destroy must be a function|TypeError:ERR_ASYNC_CALLBACK:hook.promiseResolve must be a function")
}

func TestAsyncResourceValidation(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const out = [];
		for (const make of [
			() => new async_hooks.AsyncResource(),
			() => new async_hooks.AsyncResource(''),
			() => new async_hooks.AsyncResource('type', -4),
			() => new async_hooks.AsyncResource('type', Math.PI)
		]) {
			try { make(); out.push('no-error'); }
			catch (err) { out.push(err.name + ':' + err.code); }
		}
		return out.join(',');
	})()`), "TypeError:ERR_INVALID_ARG_TYPE,TypeError:ERR_ASYNC_TYPE,RangeError:ERR_INVALID_ASYNC_ID,RangeError:ERR_INVALID_ASYNC_ID")
}

func TestClearedImmediateIsDestroyed(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		globalThis.__asyncEvents = [];
		globalThis.__asyncResource = undefined;
		const hook = async_hooks.createHook({
			init(id, type, trigger, resource) {
				if (type === 'Immediate') {
					globalThis.__asyncResource = resource;
					globalThis.__asyncEvents.push('init:' + id);
				}
			},
			destroy(id) { globalThis.__asyncEvents.push('destroy:' + id); }
		}).enable();
		const immediate = setImmediate(() => { throw new Error('cleared immediate ran'); });
		globalThis.__sameResource = immediate === globalThis.__asyncResource;
		clearImmediate(immediate);
		hook.disable();
		return 'queued';
	})()`), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__sameResource)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__asyncEvents.join(',')`), "init:2,destroy:2")
}

func TestIntervalRetainsOneId(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		globalThis.__intervalIds = [];
		globalThis.__ticks = [];
		globalThis.__destroys = 0;
		const hook = async_hooks.createHook({
			init(id, type) { if (type === 'Interval') globalThis.__intervalIds.push(id); },
			before(id) { globalThis.__ticks.push(id); },
			destroy() { globalThis.__destroys++; }
		}).enable();
		const interval = setInterval(() => {
			if (globalThis.__ticks.length === 2) clearInterval(interval);
		}, 0);
		return 'queued';
	})()`), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__intervalIds.length === 1)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__intervalIds[0] === globalThis.__ticks[0] && globalThis.__ticks[0] === globalThis.__ticks[1])`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(globalThis.__destroys)`), "1")
}

func TestAsyncLocalStorageBind(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const als = new async_hooks.AsyncLocalStorage();
		const bound = als.run('bound', () => als.bind(() => als.getStore()));
		return bound() + ':' + String(als.getStore());
	})()`), "bound:undefined")
}

func TestAsyncLocalStoragePropagatesToTimers(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		globalThis.__als = new async_hooks.AsyncLocalStorage();
		globalThis.__got = 'unset';
		globalThis.__als.run('stored', () => {
			setTimeout(() => { globalThis.__got = String(globalThis.__als.getStore()); }, 0);
		});
		return 'queued';
	})()`), "queued")

	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__got`), "stored")
}
