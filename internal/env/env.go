// Package env is the shared runtime surface every GodeJS module is built on:
// one JavaScriptCore context, a module registry, Node-style require(), a timer
// queue, process state and stdio.
package env

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"godejs/internal/js"
)

// Installer builds a module's exports object inside an environment.
type Installer func(*Env) js.Object

var (
	regMu        sync.RWMutex
	registry     = map[string]Installer{}
	nativeLoader func(*Env, string) (js.Object, error)
)

// SetNativeLoader installs the runtime's Node-API addon loader. The native
// package registers it during init; keeping the hook here avoids an import
// cycle between env and the addon implementation.
func SetNativeLoader(fn func(*Env, string) (js.Object, error)) {
	regMu.Lock()
	nativeLoader = fn
	regMu.Unlock()
}

// LoadNative loads a .node addon through the registered Node-API loader.
func (e *Env) LoadNative(path string) (js.Object, error) {
	regMu.RLock()
	fn := nativeLoader
	regMu.RUnlock()
	if fn == nil {
		return js.Object{}, &js.Error{Name: "Error", Code: "ERR_DLOPEN_FAILED",
			Message: "ERR_DLOPEN_FAILED: Node-API loader is not linked into this binary"}
	}
	return fn(e, path)
}

// Register makes a builtin module available to require() and InstallAll.
// Modules call this from init.
func Register(name string, fn Installer) {
	regMu.Lock()
	registry[name] = fn
	regMu.Unlock()
}

// Registered reports whether a builtin exists.
func Registered(name string) bool {
	regMu.RLock()
	_, ok := registry[name]
	regMu.RUnlock()
	return ok
}

// Builtins lists registered module names, sorted.
func Builtins() []string {
	regMu.RLock()
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	regMu.RUnlock()
	sort.Strings(out)
	return out
}

// Options configures an Env.
type Options struct {
	// Args is the full process.argv (argv[0] is the executable).
	Args []string
	// Env is KEY=VALUE pairs, as in os.Environ.
	Env []string
	// Cwd is the working directory (defaults to os.Getwd()).
	Cwd string
	// ExecPath is process.execPath (defaults to os.Executable()).
	ExecPath string
	// Version is process.version (defaults to version.Node).
	Version string
	// ExecArgv is process.execArgv: the runtime flags that preceded the
	// entry script. When nil it is derived from Args.
	ExecArgv []string
	// PreserveSymlinks keeps symlinked module paths intact, as Node's
	// --preserve-symlinks flag does.
	PreserveSymlinks bool
	// Stdin, Stdout and Stderr default to the process streams.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// ErrProcessExit is returned from host functions to unwind JS after
// process.exit().
var ErrProcessExit = errors.New("process.exit")

// Env is one GodeJS runtime instance.
type Env struct {
	JS   *js.Context
	Opts Options

	// ExitCode is the value passed to process.exit or set by the CLI.
	ExitCode int
	// Exited is set when the program asked to terminate.
	Exited bool
	// StartTime is when the environment was created.
	StartTime time.Time

	mu           sync.Mutex
	exports      map[string]js.Object
	stackScripts map[string]*stackScript
	installing   map[string]bool
	timers     []*Timer
	timerSeq   int
	ticks      []tick
	native     []func()
	cache      map[string]js.Object
	sources    map[string]string
	mainModule js.Object
	lastErr    error
	exitHooks  []func(int)

	// uncaught delivers an exception that escaped the loop to the process
	// builtin (an 'uncaughtException' listener or the capture callback).
	// It returns true when something absorbed the exception.
	uncaught func(js.Value, string) bool
	// beforeExit runs when the loop would finish; returning true means the
	// handler scheduled more work, so the loop keeps running.
	beforeExit func(int) bool
	// delivering guards against delivering an exception raised by the
	// delivery itself (Node exits with 7 in that case).
	delivering bool
	// pendingWork counts native asynchronous operations (for example Node-API
	// async work) whose completion will arrive on a later event-loop turn.
	pendingWork int
	// wakeTimer is reused by the loop's park/wake path. The loop now wakes for
	// every batch of queued native work, so allocating a fresh timer on each
	// park was hot-path garbage.
	wakeTimer *time.Timer
}

type tick struct {
	fn   js.Value
	args []js.Value
}

// Timer is a pending timeout or interval.
type Timer struct {
	ID       int
	When     time.Time
	Interval time.Duration
	Repeat   bool
	Fn       js.Value
	Args     []js.Value
	cleared  bool
}

// New creates an environment. Call InstallAll (or install modules manually)
// before evaluating user code.
func New(o Options) *Env {
	if o.Cwd == "" {
		o.Cwd, _ = os.Getwd()
	}
	if o.ExecPath == "" {
		o.ExecPath, _ = os.Executable()
	}
	if o.Version == "" {
		o.Version = "v24.22.2"
	}
	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	e := &Env{
		JS:         js.New(),
		Opts:       o,
		StartTime:  time.Now(),
		exports:    map[string]js.Object{},
		installing: map[string]bool{},
		cache:      map[string]js.Object{},
		sources:    map[string]string{},
	}
	g := e.JS.Global()
	g.Set("global", g.V())
	e.installStack()
	return e
}

// Out is the stdout writer.
func (e *Env) Out() io.Writer { return e.Opts.Stdout }

// Err is the stderr writer.
func (e *Env) Err() io.Writer { return e.Opts.Stderr }

// In is the stdin reader.
func (e *Env) In() io.Reader { return e.Opts.Stdin }

// Argv is process.argv.
func (e *Env) Argv() []string { return e.Opts.Args }

// Environ is the process environment as KEY=VALUE pairs.
func (e *Env) Environ() []string { return e.Opts.Env }

// Cwd is process.cwd().
func (e *Env) Cwd() string { return e.Opts.Cwd }

// Global is the JS global object.
func (e *Env) Global() js.Object { return e.JS.Global() }

// Exports installs (once) and returns a builtin module's exports object.
func (e *Env) Exports(name string) js.Object {
	name = normalizeBuiltin(name)
	e.mu.Lock()
	if obj, ok := e.exports[name]; ok {
		e.mu.Unlock()
		return obj
	}
	if e.installing[name] {
		// Dependency cycle: hand back an empty placeholder rather than
		// recursing forever.
		e.mu.Unlock()
		return e.JS.Obj()
	}
	e.installing[name] = true
	e.mu.Unlock()

	regMu.RLock()
	install := registry[name]
	regMu.RUnlock()
	if install == nil {
		e.mu.Lock()
		delete(e.installing, name)
		e.mu.Unlock()
		panic("godejs: builtin module not registered: " + name)
	}
	obj := install(e)

	e.mu.Lock()
	e.exports[name] = obj
	delete(e.installing, name)
	e.mu.Unlock()
	obj.V().Protect()
	return obj
}

// coreBuiltins are installed eagerly because they define globals (process,
// Buffer, timers, console, events, crypto) that user code may use without a
// require(). Every other builtin is initialized on first require; eagerly
// building the full registry cost roughly as much as starting JavaScriptCore.
var coreBuiltins = []string{
	"process", "buffer", "timers", "async_hooks", "console", "events", "crypto",
}

// InstallAll prepares global require and initializes the builtins that own
// process-wide globals. Other registered builtins are installed lazily by
// Exports when require() first resolves them.
func (e *Env) InstallAll() {
	for _, name := range coreBuiltins {
		if Registered(name) {
			e.Exports(name)
		}
	}
	e.setDefaultRequire(e.Opts.Cwd)
}

func (e *Env) setDefaultRequire(dir string) {
	req := e.makeRequire(dir, e.mainModule)
	e.Global().Set("require", req.V())
}

// Eval evaluates source in the global context.
func (e *Env) Eval(src string) (js.Value, error) { return e.JS.Eval(src, "[eval]") }

// RunScript evaluates source and returns its completion value.
func (e *Env) RunScript(src, filename string) (js.Value, error) {
	if filename != "" && !strings.HasPrefix(filename, "[") {
		e.registerScript(filename, src, e.JS.Undefined(), 0, 0)
	}
	return e.JS.Eval(src, filename)
}

// MainFile returns the entry script path, if one was run.
func (e *Env) MainFile() string {
	args := e.Opts.Args
	if len(args) > 1 {
		return args[1]
	}
	return ""
}

// RunMain runs the timer/microtask loop until no work remains. It returns the
// first error that no 'uncaughtException' listener absorbed.
func (e *Env) RunMain() error {
	// JavaScriptCore keeps thread-affine state for the VM. Go may migrate a
	// goroutine between OS threads after any cgo call, so pin the loop that
	// owns the context; without this a multi-P run can enter JSC from two
	// threads at once and crash inside the API.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for {
		e.drainNative()
		if err := e.DrainTicks(); err != nil {
			if e.DeliverUncaught(err, "uncaughtException") {
				continue
			}
			return err
		}
		if e.Exited {
			return nil
		}
		t, ok := e.nextTimer()
		if !ok {
			// JavaScriptCore has no public API to run the microtask
			// checkpoint. Returning from a top-level evaluation does run
			// it, so evaluate an empty script and re-check for work.
			if _, err := e.JS.Eval("void 0", "[microtask]"); err != nil {
				if e.DeliverUncaught(err, "uncaughtException") {
					continue
				}
				return err
			}
			if e.hasPendingWork() {
				continue
			}
			// The loop is empty: give the program one 'beforeExit' chance to
			// schedule more work, which is how Node keeps running after it.
			if e.fireBeforeExit() {
				continue
			}
			return nil
		}
		if wait := time.Until(t.When); wait > 0 {
			// A native completion, thread-safe function, or HTTP request can
			// queue work while the loop is parked on a long timer; wake
			// immediately instead of sleeping until that timer fires.
			select {
			case <-wake:
			case <-e.wakeAfter(wait):
			}
		}
		if t.Repeat {
			t.When = time.Now().Add(t.Interval)
			e.mu.Lock()
			e.timers = append(e.timers, t)
			e.mu.Unlock()
		} else {
			t.Fn.Unprotect()
			for _, a := range t.Args {
				a.Unprotect()
			}
		}
		if _, err := t.Fn.Object().Call(js.Object{}, t.Args...); err != nil {
			if e.DeliverUncaught(err, "uncaughtException") {
				continue
			}
			return err
		}
	}
}

// wakeAfter returns a channel that fires after d, reusing one timer per
// environment. The loop parks and wakes once per batch of queued work, so a
// new timer for every park allocated on the path between requests.
func (e *Env) wakeAfter(d time.Duration) <-chan time.Time {
	if e.wakeTimer == nil {
		e.wakeTimer = time.NewTimer(d)
		return e.wakeTimer.C
	}
	if !e.wakeTimer.Stop() {
		select {
		case <-e.wakeTimer.C:
		default:
		}
	}
	e.wakeTimer.Reset(d)
	return e.wakeTimer.C
}

// DrainTicks runs queued nextTick callbacks.
func (e *Env) DrainTicks() error {
	for {
		e.mu.Lock()
		if len(e.ticks) == 0 {
			e.mu.Unlock()
			return nil
		}
		t := e.ticks[0]
		e.ticks = e.ticks[1:]
		e.mu.Unlock()
		if _, err := t.fn.Object().Call(js.Object{}, t.args...); err != nil {
			return err
		}
	}
}

// wake is signalled when another goroutine queues event-loop work. The timer
// loop selects on it so native async completions and thread-safe functions do
// not have to wait for the next timer to fire.
var wake = make(chan struct{}, 1)

func signalWake() {
	select {
	case wake <- struct{}{}:
	default:
	}
}

// NextTick queues a callback to run after the current script finishes.
func (e *Env) NextTick(fn js.Value, args ...js.Value) {
	fn.Protect()
	for _, a := range args {
		a.Protect()
	}
	e.mu.Lock()
	e.ticks = append(e.ticks, tick{fn: fn, args: args})
	e.mu.Unlock()
	signalWake()
}

// EnqueueNative schedules fn to run on the thread that owns the JavaScript
// context. Native worker goroutines must use it instead of touching
// JavaScriptCore themselves: the context is thread-affine and their goroutine
// may be running on any OS thread.
func (e *Env) EnqueueNative(fn func()) {
	e.mu.Lock()
	e.native = append(e.native, fn)
	e.mu.Unlock()
	signalWake()
}

// drainNative runs native callbacks queued from worker goroutines. It is only
// called by the loop that owns the context.
func (e *Env) drainNative() {
	for {
		e.mu.Lock()
		if len(e.native) == 0 {
			e.mu.Unlock()
			return
		}
		fn := e.native[0]
		e.native = e.native[1:]
		e.mu.Unlock()
		fn()
	}
}

// SetTimeout schedules fn. repeat turns it into an interval. delay is in
// milliseconds; negative or NaN delays fire on the next turn.
func (e *Env) SetTimeout(fn js.Value, delay float64, repeat bool, args []js.Value) int {
	if delay < 0 || delay != delay {
		delay = 0
	}
	e.mu.Lock()
	e.timerSeq++
	id := e.timerSeq
	e.mu.Unlock()
	fn.Protect()
	for _, a := range args {
		a.Protect()
	}
	t := &Timer{
		ID:       id,
		When:     time.Now().Add(time.Duration(delay * float64(time.Millisecond))),
		Interval: time.Duration(delay * float64(time.Millisecond)),
		Repeat:   repeat,
		Fn:       fn,
		Args:     args,
	}
	e.mu.Lock()
	e.timers = append(e.timers, t)
	e.mu.Unlock()
	return id
}

// ClearTimer cancels a pending timer.
func (e *Env) ClearTimer(id int) {
	e.mu.Lock()
	for _, t := range e.timers {
		if t.ID == id && !t.cleared {
			t.cleared = true
			t.Fn.Unprotect()
			for _, a := range t.Args {
				a.Unprotect()
			}
		}
	}
	e.mu.Unlock()
}

// RefreshTimer reschedules a timer from now.
func (e *Env) RefreshTimer(id int, delay float64) bool {
	if delay < 0 || delay != delay {
		delay = 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, t := range e.timers {
		if t.ID == id && !t.cleared {
			t.When = time.Now().Add(time.Duration(delay * float64(time.Millisecond)))
			t.Interval = time.Duration(delay * float64(time.Millisecond))
			return true
		}
	}
	return false
}

// TimerActive reports whether a timer is still pending.
func (e *Env) TimerActive(id int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, t := range e.timers {
		if t.ID == id && !t.cleared {
			return true
		}
	}
	return false
}

// AddPendingWork keeps the event loop alive for a native async operation.
func (e *Env) AddPendingWork() {
	e.mu.Lock()
	e.pendingWork++
	e.mu.Unlock()
	signalWake()
}

// DonePendingWork releases one native async operation.
func (e *Env) DonePendingWork() {
	e.mu.Lock()
	if e.pendingWork > 0 {
		e.pendingWork--
	}
	e.mu.Unlock()
}

// hasPendingWork reports whether ticks, timers or native async work remain
// after a microtask checkpoint.
func (e *Env) hasPendingWork() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.ticks) > 0 || e.pendingWork > 0 {
		return true
	}
	for _, t := range e.timers {
		if !t.cleared {
			return true
		}
	}
	return false
}

func (e *Env) nextTimer() (*Timer, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var best *Timer
	for _, t := range e.timers {
		if t.cleared {
			continue
		}
		if best == nil || t.When.Before(best.When) {
			best = t
		}
	}
	if best == nil {
		return nil, false
	}
	out := e.timers[:0]
	for _, t := range e.timers {
		if t != best {
			out = append(out, t)
		}
	}
	e.timers = out
	return best, true
}

// Exit requests process termination. Host functions should return
// ErrProcessExit after calling it.
func (e *Env) Exit(code int) {
	e.ExitCode = code
	e.Exited = true
}

// OnExit registers a callback fired once when the runtime finishes.
func (e *Env) OnExit(fn func(int)) {
	e.mu.Lock()
	e.exitHooks = append(e.exitHooks, fn)
	e.mu.Unlock()
}

// SetUncaughtHandler installs the process builtin's exception delivery. The
// handler receives the thrown value and its origin ('uncaughtException' or
// 'unhandledRejection') and reports whether a listener absorbed it.
func (e *Env) SetUncaughtHandler(fn func(js.Value, string) bool) {
	e.mu.Lock()
	e.uncaught = fn
	e.mu.Unlock()
}

// SetBeforeExitHook installs the callback run when the loop empties. It
// returns true when it scheduled new work.
func (e *Env) SetBeforeExitHook(fn func(int) bool) {
	e.mu.Lock()
	e.beforeExit = fn
	e.mu.Unlock()
}

// DeliverUncaught hands err to the process builtin. It returns false when no
// handler is installed, when a previous delivery is still running (Node exits
// with code 7 in that case) or when the handler itself threw.
func (e *Env) DeliverUncaught(err error, origin string) bool {
	if err == nil {
		return false
	}
	// process.exit() is terminal, never an exception: it must not be handed
	// to an 'uncaughtException' listener (npm's exit handler calls
	// process.exit from one, which would otherwise loop forever).
	if e.Exited || errors.Is(err, ErrProcessExit) {
		return false
	}
	e.mu.Lock()
	h := e.uncaught
	if h == nil || e.delivering {
		if e.delivering {
			e.ExitCode = 7
		}
		e.mu.Unlock()
		return false
	}
	e.delivering = true
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.delivering = false
		e.mu.Unlock()
	}()
	return h(e.JS.ErrorValue(err), origin)
}

// fireBeforeExit runs the beforeExit hook and reports whether it queued work.
func (e *Env) fireBeforeExit() bool {
	e.mu.Lock()
	h := e.beforeExit
	e.mu.Unlock()
	if h == nil || e.Exited {
		return false
	}
	if !h(e.ExitCode) {
		return false
	}
	return e.hasPendingWork()
}

// FireExitHooks runs the registered exit callbacks with the final exit code.
func (e *Env) FireExitHooks() {
	e.mu.Lock()
	hooks := append([]func(int){}, e.exitHooks...)
	e.mu.Unlock()
	for _, fn := range hooks {
		fn(e.ExitCode)
	}
}

// RunFile loads and executes a CommonJS module as the entry point.
func (e *Env) RunFile(path string) (js.Object, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return js.Object{}, err
	}
	if st, statErr := os.Stat(abs); statErr == nil && st.IsDir() {
		resolved, ok := resolveAsFile(abs)
		if !ok {
			return js.Object{}, &ModuleError{Module: abs, Message: "Cannot find module '" + abs + "'"}
		}
		abs = resolved
	}
	obj, err := e.loadFile(abs, true)
	if err != nil {
		return js.Object{}, err
	}
	return obj, nil
}

// SetSourceOverride makes loadFile evaluate src for path instead of reading
// the file from disk. The debugger uses it to run an instrumented copy while
// preserving the real filename, require paths and __dirname.
func (e *Env) SetSourceOverride(path, src string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	e.mu.Lock()
	if e.sources == nil {
		e.sources = map[string]string{}
	}
	e.sources[abs] = src
	e.mu.Unlock()
}

func (e *Env) defineOwn(obj js.Object, name string, val js.Value) {
	fn, err := e.JS.Eval(`(function (o, n, v) {
  Object.defineProperty(o, n, { value: v, writable: true, enumerable: true, configurable: true });
})`, "[gode:define]")
	if err != nil {
		return
	}
	_, _ = fn.Object().Call(js.Object{}, obj.V(), e.JS.Str(name), val)
}

func (e *Env) loadFile(abs string, main bool) (js.Object, error) {
	if !e.Opts.PreserveSymlinks {
		if real, err := filepath.EvalSymlinks(abs); err == nil {
			abs = real
		}
	}
	e.mu.Lock()
	if cached, ok := e.cache[abs]; ok {
		e.mu.Unlock()
		// A cache hit must hand back the module's current exports, not the
		// module object. During a require cycle the exports object may be
		// only partially populated, which is exactly CommonJS behavior.
		if exp := cached.Get("exports"); exp.Valid() {
			return exp.Object(), nil
		}
		return cached, nil
	}
	e.mu.Unlock()

	e.mu.Lock()
	override, hasOverride := e.sources[abs]
	e.mu.Unlock()
	var src []byte
	if hasOverride {
		src = []byte(override)
	} else {
		var err error
		src, err = os.ReadFile(abs)
		if err != nil {
			return js.Object{}, &ModuleError{Module: abs, Message: "Cannot find module '" + abs + "'"}
		}
	}
	body := prepareSource(string(src))
	dir := filepath.Dir(abs)

	moduleObj := e.JS.Obj()
	exports := e.JS.Obj()
	moduleObj.SetStr("id", abs)
	moduleObj.SetStr("filename", abs)
	// Define exports as an own data property. A plain Set would go through a
	// user-installed Object.prototype.exports setter, which Node's loader
	// deliberately avoids during module creation.
	e.defineOwn(moduleObj, "exports", exports.V())
	moduleObj.SetBool("loaded", false)
	moduleObj.Set("children", e.JS.Arr(nil).V())
	if main {
		// Node exposes module.parent as a deprecated accessor: null for the
		// entry module, with DEP0144 emitted only when it is read.
		if fn, err := e.JS.Eval(moduleParentGetter, "[gode:module-parent]"); err == nil {
			_, _ = fn.Object().Call(js.Object{}, moduleObj.V())
		}
	}

	e.mu.Lock()
	e.cache[abs] = moduleObj
	if main && !e.mainModule.Valid() {
		e.mainModule = moduleObj
	}
	mainModule := e.mainModule
	e.mu.Unlock()

	requireFn := e.makeRequire(dir, mainModule)
	requireFn.SetHidden("$module", moduleObj.V())

	if fn, ok := e.extensionHandler(abs); ok {
		if _, err := fn.Object().Call(js.Object{}, moduleObj.V(), e.JS.Str(abs)); err != nil {
			return js.Object{}, err
		}
		moduleObj.SetBool("loaded", true)
		return moduleObj.Get("exports").Object(), nil
	}

	if strings.HasSuffix(abs, ".node") {
		exp, loadErr := e.LoadNative(abs)
		if loadErr != nil {
			return js.Object{}, loadErr
		}
		moduleObj.Set("exports", exp.V())
		moduleObj.SetBool("loaded", true)
		e.mu.Lock()
		e.cache[abs] = moduleObj
		e.mu.Unlock()
		return exp, nil
	}

	if moduleFormat(abs) == "json" {
		v, err := e.JS.ParseJSON(body)
		if err != nil {
			return js.Object{}, err
		}
		moduleObj.Set("exports", v)
		moduleObj.SetBool("loaded", true)
		return v.Object(), nil
	}
	isESM := moduleFormat(abs) == "esm"
	if isESM {
		body = transformESM(body, abs)
	}

	var wrapped string
	var before string
	if isESM {
		// Node ESM has no __filename/__dirname/require/exports/module
		// bindings; ESM code may legally declare consts with those names.
		prefix := "(function (" + esmExports + ", " + esmRequire + ", " + esmModule + ") {"
		before = prefix
		wrapped = prefix + body + "\n})"
	} else {
		// CommonJS modules are compiled through Module.wrapper so tests (and
		// loaders) that patch Module.wrapper take effect.
		body = rewriteDynamicImports(body)
		helper := ""
		if strings.Contains(body, "__godeImport(") {
			helper = dynamicImportHelper + "\n"
		}
		prefix, suffix := "(function (exports, require, module, __filename, __dirname) {", "\n})"
		if Registered("module") {
			if me := e.Exports("module"); me.Valid() {
				if w := me.Get("wrapper"); w.IsObject() && !w.IsNull() {
					wo := w.Object()
					if wo.Get("0").IsString() && wo.Get("1").IsString() {
						prefix = wo.Get("0").Str()
						suffix = wo.Get("1").Str()
					}
				}
			}
		}
		before = prefix + "\n" + helper
		wrapped = before + body + suffix
	}
	lineOffset, colShift := stackOffsets(before)
	topThis := e.JS.Undefined()
	if !isESM {
		topThis = exports.V()
	}
	e.registerScript(abs, body, topThis, lineOffset, colShift)
	fnVal, err := e.JS.Eval(wrapped, abs)
	if err != nil {
		return js.Object{}, err
	}
	var callErr error
	if isESM {
		_, callErr = fnVal.Object().Call(js.Object{}, exports.V(), requireFn.V(), moduleObj.V())
	} else {
		_, callErr = fnVal.Object().Call(js.Object{},
			exports.V(), requireFn.V(), moduleObj.V(), e.JS.Str(abs), e.JS.Str(dir))
	}
	if callErr != nil {
		if jsErr, ok := callErr.(*js.Error); ok && strings.HasPrefix(jsErr.Message, "Can't find variable: ") {
			name := strings.TrimPrefix(jsErr.Message, "Can't find variable: ")
			return js.Object{}, &js.Error{Name: "ReferenceError", Message: name + " is not defined"}
		}
		return js.Object{}, callErr
	}
	moduleObj.SetBool("loaded", true)
	// module.exports may have been reassigned.
	exp := moduleObj.Get("exports")
	if !exp.Valid() {
		return js.Object{}, errors.New("module.exports is not set")
	}
	e.mu.Lock()
	e.cache[abs] = moduleObj
	e.mu.Unlock()
	return exp.Object(), nil
}

func (e *Env) makeRequire(dir string, mainModule js.Object) js.Object {
	callable := e.requireCallable(dir)
	callable.Set("cache", e.requireCacheObject().V())
	if Registered("module") {
		if me := e.Exports("module"); me.Valid() {
			if ext := me.Get("_extensions"); ext.Valid() && ext.IsObject() {
				callable.Set("extensions", ext)
			}
		}
	}
	if mainModule.Valid() {
		callable.Set("main", mainModule.V())
	}
	callable.SetFunc("resolve", func(this js.Object, args []js.Value) (js.Value, error) {
		spec := ""
		if len(args) > 0 {
			spec = args[0].Str()
		}
		if isBuiltinSpec(spec) {
			return e.JS.Str(normalizeBuiltin(spec)), nil
		}
		resolved, err := e.resolveFile(dir, spec)
		if err != nil {
			return js.Value{}, &ModuleError{Module: spec, Message: "Cannot find module '" + spec + "'"}
		}
		return e.JS.Str(resolved), nil
	})
	return callable
}

func (e *Env) requireCacheObject() js.Object {
	cache := e.Obj()
	e.mu.Lock()
	keys := make([]string, 0, len(e.cache))
	for k := range e.cache {
		keys = append(keys, k)
	}
	e.mu.Unlock()
	sort.Strings(keys)
	for _, k := range keys {
		e.mu.Lock()
		mod := e.cache[k]
		e.mu.Unlock()
		cache.Set(k, mod.V())
	}
	return cache
}

func (e *Env) requireCallable(dir string) js.Object {
	host := e.JS.Obj()
	host.SetFunc("load", func(this js.Object, args []js.Value) (js.Value, error) {
		spec := ""
		if len(args) > 0 {
			spec = args[0].Str()
		}
		var parent js.Object
		if len(args) > 1 && args[1].IsObject() && !args[1].IsNull() {
			parent = args[1].Object()
		}
		v, err := e.requireFromParent(dir, spec, parent)
		out := e.JS.Obj()
		if err != nil {
			// Returning the error as data and throwing it from the JS wrapper
			// is deliberate. A host callback that throws while a module load
			// is already reentrant can lose the exception in JavaScriptCore,
			// making a failing require look successful.
			out.SetBool("ok", false)
			out.Set("error", e.JS.ErrorValue(err))
			return out.V(), nil
		}
		out.SetBool("ok", true)
		out.Set("value", v)
		return out.V(), nil
	})
	factory, err := e.JS.Eval(`(function (host) { return function require(id) {
		if (typeof id !== 'string') {
			const e = new TypeError('The "id" argument must be of type string. Received ' + typeof id);
			e.code = 'ERR_INVALID_ARG_TYPE'; throw e;
		}
		if (id === '') {
			const e = new TypeError("The argument 'id' must be a non-empty string. Received ''");
			e.code = 'ERR_INVALID_ARG_VALUE'; throw e;
		}
		const r = host.load(String(id), require.$module); if (!r.ok) throw r.error; return r.value;
	}; })`, "[gode:require]")
	if err != nil {
		panic("godejs: require wrapper: " + err.Error())
	}
	wrapper, err := factory.Object().Call(js.Object{}, host.V())
	if err != nil {
		panic("godejs: require wrapper: " + err.Error())
	}
	return wrapper.Object()
}

func (e *Env) requireFromParent(dir, spec string, parent js.Object) (js.Value, error) {
	v, err := e.requireFrom(dir, spec)
	if err != nil || !parent.Valid() || isBuiltinSpec(spec) {
		return v, err
	}
	resolved, rerr := e.resolveFile(dir, spec)
	if rerr != nil {
		return v, nil
	}
	e.mu.Lock()
	mod := e.cache[resolved]
	e.mu.Unlock()
	if !mod.Valid() {
		return v, nil
	}
	children := parent.Get("children")
	if !children.IsObject() || children.IsNull() {
		return v, nil
	}
	arr := children.Object()
	n := arr.Length()
	id := mod.Get("id").Str()
	for i := 0; i < n; i++ {
		if cur := arr.At(i); cur.IsObject() && cur.Object().Get("id").Str() == id {
			return v, nil
		}
	}
	arr.Set(fmt.Sprintf("%d", n), mod.V())
	return v, nil
}

func (e *Env) requireFrom(dir, spec string) (js.Value, error) {
	if spec == "" {
		return js.Value{}, &ModuleError{Message: "Cannot find module ''", Module: spec}
	}
	if isBuiltinSpec(spec) {
		name := normalizeBuiltin(spec)
		if Registered(name) {
			return e.Exports(name).V(), nil
		}
		return js.Value{}, &ModuleError{Module: spec, Message: "Cannot find module '" + spec + "'"}
	}
	resolved, err := e.resolveFile(dir, spec)
	if err != nil {
		return js.Value{}, &ModuleError{Module: spec, Message: "Cannot find module '" + spec + "'"}
	}
	obj, err := e.loadFile(resolved, false)
	if err != nil {
		return js.Value{}, err
	}
	return obj.V(), nil
}

// RequireFrom loads a module as if require() were called from dir. Builtins
// accept the "node:" prefix. It is the host half of module.createRequire.
func (e *Env) RequireFrom(dir, spec string) (js.Value, error) {
	return e.requireFrom(dir, spec)
}

// RequireResolve resolves spec from dir without loading it.
func (e *Env) RequireResolve(dir, spec string) (string, error) {
	if isBuiltinSpec(spec) {
		return normalizeBuiltin(spec), nil
	}
	return e.resolveFile(dir, spec)
}

// Obj creates a fresh object in the environment's context.
func (e *Env) Obj() js.Object { return e.JS.Obj() }

// ModuleError is a module resolution/loading failure.
type ModuleError struct {
	Module  string
	Message string
}

func (m *ModuleError) Error() string { return m.Message }

// JSErrorCode makes a failed require surface as Node's coded error, so
// callers can branch on err.code === 'MODULE_NOT_FOUND'.
func (m *ModuleError) JSErrorCode() (string, string) { return "Error", "MODULE_NOT_FOUND" }

// Snapshot returns the runtime's exported module names and their objects.
func (e *Env) Snapshot() map[string]js.Object {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make(map[string]js.Object, len(e.exports))
	for k, v := range e.exports {
		out[k] = v
	}
	return out
}

func normalizeBuiltin(name string) string {
	return strings.TrimPrefix(name, "node:")
}

func isBuiltinSpec(spec string) bool {
	name := strings.TrimPrefix(spec, "node:")
	return Registered(name)
}

func prepareSource(src string) string {
	src = strings.TrimPrefix(src, "\ufeff")
	if strings.HasPrefix(src, "#!") {
		if i := strings.IndexByte(src, '\n'); i >= 0 {
			src = "//" + src[2:]
		}
	}
	return stripV8IntrinsicEvals(src)
}

func (e *Env) envValue(key string) string {
	prefix := key + "="
	for _, kv := range e.Opts.Env {
		if strings.HasPrefix(kv, prefix) {
			return kv[len(prefix):]
		}
	}
	return ""
}

func (e *Env) extensionKeys() []string {
	if !Registered("module") {
		return nil
	}
	ext := e.Exports("module").Get("_extensions")
	if !ext.IsObject() || ext.IsNull() {
		return nil
	}
	fn, err := e.JS.Eval("(function (o) { return Object.keys(o); })", "[gode:extensions]")
	if err != nil {
		return nil
	}
	v, err := fn.Object().Call(js.Object{}, ext)
	if err != nil || !v.IsObject() || v.IsNull() {
		return nil
	}
	arr := v.Object()
	keys := make([]string, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		if k := arr.At(i); k.IsString() {
			keys = append(keys, k.Str())
		}
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	return keys
}

func (e *Env) extensionHandler(abs string) (js.Value, bool) {
	base := filepath.Base(abs)
	var ext string
	if strings.HasPrefix(base, ".") && strings.Count(base, ".") == 1 {
		return js.Value{}, false
	}
	if strings.HasPrefix(base, ".") {
		ext = filepath.Ext(abs)
	} else {
		for _, k := range e.extensionKeys() {
			if strings.HasSuffix(base, k) {
				ext = k
				break
			}
		}
	}
	if ext == "" || ext == ".js" || ext == ".json" || ext == ".node" {
		return js.Value{}, false
	}
	if !Registered("module") {
		return js.Value{}, false
	}
	fn := e.Exports("module").Get("_extensions").Object().Get(ext)
	if !fn.IsObject() || fn.IsNull() {
		return js.Value{}, false
	}
	return fn, true
}

func (e *Env) globalRoots() []string {
	var roots []string
	if np := e.envValue("NODE_PATH"); np != "" {
		for _, p := range strings.Split(np, string(os.PathListSeparator)) {
			if p != "" {
				roots = append(roots, p)
			}
		}
	}
	home := e.envValue("HOME")
	if home == "" {
		home = e.envValue("USERPROFILE")
	}
	if home != "" {
		roots = append(roots, filepath.Join(home, ".node_modules"), filepath.Join(home, ".node_libraries"))
	}
	if exe := e.Opts.ExecPath; exe != "" {
		roots = append(roots, filepath.Join(filepath.Dir(filepath.Dir(exe)), "lib", "node"))
	}
	return roots
}

func (e *Env) resolveFile(dir, spec string) (string, error) {
	resolved, err := resolveFile(dir, spec)
	if err == nil {
		return resolved, nil
	}
	if strings.HasPrefix(spec, "#") || isBuiltinSpec(spec) {
		return "", err
	}
	base := spec
	if !filepath.IsAbs(base) {
		base = filepath.Join(dir, base)
	}
	for _, ext := range e.extensionKeys() {
		if ext == ".js" || ext == ".json" || ext == ".node" || strings.HasSuffix(base, ext) {
			continue
		}
		if st, statErr := os.Stat(base + ext); statErr == nil && !st.IsDir() {
			if abs, absErr := filepath.Abs(base + ext); absErr == nil {
				return abs, nil
			}
		}
	}
	if !strings.HasPrefix(spec, ".") && !filepath.IsAbs(spec) {
		for _, root := range e.globalRoots() {
			if abs, ok := resolveAsFile(filepath.Join(root, filepath.FromSlash(spec))); ok {
				return abs, nil
			}
		}
	}
	return "", err
}

func resolveFile(dir, spec string) (string, error) {
	if strings.HasPrefix(spec, "#") {
		return resolveImport(dir, spec)
	}
	if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") ||
		spec == "." || spec == ".." || filepath.IsAbs(spec) {
		base := spec
		if !filepath.IsAbs(base) {
			base = filepath.Join(dir, base)
		}
		if resolved, ok := resolveAsFile(base); ok {
			return resolved, nil
		}
		return "", fmt.Errorf("not found: %s", spec)
	}
	return resolvePackage(dir, spec)
}

// moduleParentGetter installs Node's deprecated module.parent accessor on the
// entry module. The warning fires on read, not at startup.
const moduleParentGetter = `(function (m) {
  if (typeof process === 'undefined' || !process.execArgv ||
      process.execArgv.indexOf('--pending-deprecation') === -1) {
    m.parent = null;
    return;
  }
  Object.defineProperty(m, 'parent', {
    configurable: true,
    enumerable: true,
    get: function () {
      process.emitWarning(
        'module.parent is deprecated due to accuracy issues. Please use ' +
        'require.main to find program entry point instead.',
        'DeprecationWarning', 'DEP0144');
      return null;
    }
  });
})`
