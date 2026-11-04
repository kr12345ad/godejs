package internals

import (
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"

	// The shims adapt these public modules, so they must be registered in the
	// test binary as well.
	_ "godejs/modules/assert"
	_ "godejs/modules/buffer"
	_ "godejs/modules/child_process"
	_ "godejs/modules/crypto"
	_ "godejs/modules/events"
	_ "godejs/modules/fs"
	_ "godejs/modules/os"
	_ "godejs/modules/path"
	_ "godejs/modules/process"
	_ "godejs/modules/timers"
	_ "godejs/modules/url"
	_ "godejs/modules/util"
)

// setup installs every shim and exposes each namespace as a global named
// __internal_<name>, so tests can poke at it from JavaScript.
func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	for _, name := range Implemented() {
		global := "__" + strings.NewReplacer("/", "_", "-", "_").Replace(name)
		e.Global().Set(global, e.Exports(name).V())
	}
	return e
}

func TestUtilShimMirrorsPublicModule(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `__internal_util.inspect(1)`), "1")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_util.promisify`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_util.callbackify`), "function")
	jstest.Equal(t, jstest.Str(t, e, `Object.isFrozen(__internal_util.kEmptyObject)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(__internal_util.isPrimitive(null))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(__internal_util.isPrimitive({}))`), "false")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_util.getCallSites`), "function")
	jstest.Equal(t, jstest.Str(t, e, `__internal_util.formatList(['a', 1])`), "'a', 1")
}

func TestErrorsShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `__internal_errors.codes.ERR_INVALID_ARG_TYPE`), "ERR_INVALID_ARG_TYPE")
	jstest.Equal(t, jstest.Str(t, e, `new __internal_errors.AbortError('stop').code`), "ABORT_ERR")
	jstest.Equal(t, jstest.Str(t, e, `new __internal_errors.AbortError('stop').name`), "AbortError")
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const outer = new Error('outer');
			outer.code = 'OUTER';
			const err = __internal_errors.aggregateTwoErrors(new Error('inner'), outer);
			return err.code;
		})()`), "OUTER")
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const err = __internal_errors.uvException({ errno: -2, syscall: 'open', path: '/nope' });
			return err.code + ' ' + err.syscall + ' ' + err.path;
		})()`), "ENOENT open /nope")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_errors.hideStackFrames(() => 1)`), "function")
}

func TestLinkedListShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const list = __internal_linkedlist.init({});
			const a = { name: 'a' };
			const b = { name: 'b' };
			const emptyAtStart = __internal_linkedlist.isEmpty(list);
			__internal_linkedlist.append(list, a);
			__internal_linkedlist.append(list, b);
			const head = __internal_linkedlist.peek(list).name;
			const tail = list._idleNext.name;
			__internal_linkedlist.remove(b);
			const afterRemove = __internal_linkedlist.peek(list).name;
			return emptyAtStart + ' ' + head + ' ' + tail + ' ' + afterRemove;
		})()`), "true a b a")
}

func TestMyersDiffShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const diff = __internal_assert_myers_diff.myersDiff(['a', 'b'], ['a', 'c']);
			const kinds = diff.map((entry) => entry[0]).sort().join(',');
			return diff.length + ' ' + kinds;
		})()`), "3 -1,0,1")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_assert_myers_diff.printSimpleMyersDiff([[-1, 'gone']]).length > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
		typeof __internal_assert_myers_diff.printMyersDiff([[-1, 'x']], 'strictEqual').message`), "string")
}

func TestValidatorsShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			try { __internal_validators.validateString(1, 'name'); }
			catch (err) { return err.name + ' ' + err.code + ' ' + err.message; }
			return 'no throw';
		})()`), `TypeError ERR_INVALID_ARG_TYPE The "name" argument must be of type string. Received type number (1)`)
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			try { __internal_validators.validateUint32(-1, 'n'); }
			catch (err) { return err.code; }
			return 'no throw';
		})()`), "ERR_OUT_OF_RANGE")
	jstest.Equal(t, jstest.Str(t, e, `__internal_validators.validatePort('8080')`), "8080")
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			__internal_validators.validateObject({ a: 1 }, 'opts');
			__internal_validators.validateStringArray(['a'], 'list');
			return 'ok';
		})()`), "ok")
}

func TestFSAndURLShims(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `__internal_fs_utils.toUnixTimestamp('5')`), "5")
	jstest.Equal(t, jstest.Str(t, e, `__internal_fs_utils.toUnixTimestamp(new Date(7))`), "7")
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const opts = __internal_fs_utils.getOptions({ a: 1 }, { b: 2 });
			return opts.a + ' ' + opts.b;
		})()`), "1 2")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_fs_utils.DirentFromStats`), "function")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_url.getPathFromURLPosix(new URL('file:///a%20b'))`), "/a b")
	jstest.Equal(t, jstest.Str(t, e, `
		typeof __internal_url.pathToFileURL`), "function")
}

func TestRootAndOptionsShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `String(__internal_root.isLinux === (process.platform === 'linux'))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(__internal_options.getOptionValue('--expose-internals'))`), "false")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_util_colors.red`), "string")
}

func TestEventTargetShimDefinesOnAccessor(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			const emitter = {
				_handlers: [],
				on(name, fn) { this._handlers.push(fn); return this; },
				listeners(name) { return this._handlers.slice(); },
				removeListener(name, fn) {
					this._handlers = this._handlers.filter((entry) => entry !== fn);
					return this;
				}
			};
			const fn = () => {};
			__internal_event_target.defineEventHandler(emitter, 'data');
			emitter.ondata = fn;
			return String(emitter.ondata === fn) + ' ' + String(emitter._handlers.length);
		})()`), "true 1")
}

func TestTestBindingShim(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_test_binding.internalBinding`), "function")

	// Node guarantees a stable namespace per id, and tests patch it in place.
	jstest.Equal(t, jstest.Str(t, e, `String(
		__internal_test_binding.internalBinding('util') ===
		__internal_test_binding.internalBinding('util'))`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.internalBinding('util').getConstructorName({})`), "Object")
	jstest.Equal(t, jstest.Str(t, e, `String(
		__internal_test_binding.internalBinding('util').isInsideNodeModules())`), "false")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.internalBinding('util').isNativeError(new TypeError('x'))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.internalBinding('util').previewEntries(new Map([['a', 1]])).join('|')`), "a,1|true")
	jstest.Equal(t, jstest.Str(t, e, `
		typeof __internal_test_binding.internalBinding('util').getCallSites`), "function")

	jstest.Equal(t, jstest.Str(t, e, `
		typeof __internal_test_binding.internalBinding('js_stream').JSStream`), "function")
	jstest.Equal(t, jstest.Str(t, e, `
		String(new (__internal_test_binding.internalBinding('js_stream').JSStream)() instanceof
			__internal_test_binding.internalBinding('js_stream').JSStream)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
		String(__internal_test_binding.internalBinding('uv').UV_ENOENT)`), "-2")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.internalBinding('constants').fs.UV_DIRENT_DIR`), "2")
	jstest.Equal(t, jstest.Str(t, e, `String(
		typeof __internal_test_binding.internalBinding('config').hasSmallICU === 'function' &&
		__internal_test_binding.internalBinding('config').hasSmallICU())`), "false")

	// internal_only_* bindings must stay hidden even from the test accessor.
	jstest.Equal(t, jstest.Str(t, e, `
		(() => {
			try { __internal_test_binding.internalBinding('internal_only_v8'); }
			catch (e) { return e.code; }
			return 'did not throw';
		})()`), "ERR_INVALID_MODULE")
}

func TestPrimordials(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.primordials.StringPrototypeStartsWith('abc', 'ab')`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
		(() => { const a = [1]; __internal_test_binding.primordials.ArrayPrototypePush(a, 2); return a.join(','); })()`), "1,2")
	jstest.Equal(t, jstest.Str(t, e, `
		String(__internal_test_binding.primordials.ArrayIsArray([]))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `
		__internal_test_binding.primordials.ObjectKeys({a: 1}).join(',')`), "a")
}

func TestRemainingShimsResolve(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_fs_recursive_watch.watch`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_fs_recursive_watch.kFSWatchStart`), "symbol")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_js_stream_socket.JSStreamSocket`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_child_process.ChildProcess`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof __internal_dgram.newSocketHandle`), "function")
}
