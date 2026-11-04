// Package fs implements Node.js's fs module on top of the GodeJS runtime.
//
// The host performs file I/O synchronously. Promise-returning APIs expose an
// already-settled Promise, and callback APIs deliver their result through
// env.NextTick so that callbacks run after the current script, matching Node's
// ordering for programs that do not depend on real concurrent I/O.
package fs

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"

	"godejs/internal/env"
	"godejs/internal/js"
	// fs.ReadStream/WriteStream build on the stream classes; importing the
	// package registers it for embedders and module tests that only install fs.
	_ "godejs/modules/events"
	_ "godejs/modules/stream"
)

func init() {
	env.Register("fs", Install)
	env.Register("fs/promises", installPromises)
}

// installPromises exposes the fs.promises surface directly, as
// require('fs/promises') does in Node. It must be the very same object as
// fs.promises (and share fs.constants), so reuse the fs exports instead of
// building a second module instance.
func installPromises(e *env.Env) js.Object {
	if p := e.Exports("fs").Get("promises"); p.Valid() && p.IsObject() {
		return p.Object()
	}
	return Install(e).Get("promises").Object()
}

type module struct {
	e           *env.Env
	exports     js.Object
	statsProto  js.Object
	direntProto js.Object
	statfsProto js.Object
	fhProto     js.Object
	blobFactory js.Value
	invoke      js.Value
	throwing    js.Value
}

// Install builds the fs exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e}
	// Node invokes every fs callback with `this` bound to undefined (strict
	// callbacks observe it); the runtime's tick helper passes the global
	// otherwise, so funnel callbacks through a tiny JS apply shim.
	if fn, err := e.JS.Eval(`(function () {
  return function (fn) {
    var args = Array.prototype.slice.call(arguments, 1);
    return fn.apply(undefined, args);
  };
})()`, "[fs:invoke]"); err == nil {
		// Rooted explicitly: a lone function is only reachable from Go, so
		// JavaScriptCore may otherwise collect it and every later callback
		// through this shim silently disappears.
		m.invoke = fn.Protect()
	}
	// Node's fs errors carry errno/code/syscall/path; raising them from JS keeps
	// every field on the object the caller catches.
	if fn, err := e.JS.Eval(`(function () {
  return function (props) {
    var e = new Error(props.message);
    e.name = 'Error';
    if (props.code) e.code = props.code;
    e.errno = props.errno;
    if (props.syscall) e.syscall = props.syscall;
    if (props.path) e.path = props.path;
    if (props.dest) e.dest = props.dest;
    throw e;
  };
})()`, "[fs:throw]"); err == nil {
		m.throwing = fn.Protect()
	}
	m.statsProto = m.makeStatsProto().Keep()
	m.direntProto = m.makeDirentProto().Keep()
	m.statfsProto = m.makeStatFsProto().Keep()
	// FileHandle instances and the FileHandle class share one prototype so
	// instanceof works and the stream/iter helpers can decorate it once.
	m.fhProto = m.makeFileHandleProto().Keep()
	// fs.openAsBlob builds real Blobs through the global constructor; calling it
	// as a plain function would throw, so keep a `new`-capable factory.
	if factory, ferr := e.JS.Eval(`(function (C) {
  return function (data, type) { return new C([data], { type: type }); };
})`, "[fs:blob]"); ferr == nil {
		m.blobFactory = factory.Protect()
	}

	exports := e.JS.Obj()
	m.exports = exports
	m.installSync(exports)
	m.installCallbacks(exports)
	m.installPromises(exports)
	m.installAdvanced(exports)

	// fs.watch/Dir and the fs stream classes read fs.constants while the
	// prelude is evaluated, so the table has to exist before installStreams.
	consts := m.constants()
	exports.SetObj("constants", consts)

	m.installStreams(exports)
	exports.Set("Stats", m.ctorClass("Stats", m.statsProto, func(this js.Object, args []js.Value) (js.Value, error) {
		o := e.JS.Obj()
		o.SetProto(m.statsProto)
		// Node's Stats constructor also accepts the legacy positional fields
		// (dev, mode, nlink, uid, gid, rdev, blksize, ino, size, blocks, then
		// the four millisecond timestamps), which watchFile uses for the
		// all-zero stat it passes when the file does not exist yet.
		if len(args) > 0 {
			for i, name := range []string{"dev", "mode", "nlink", "uid", "gid", "rdev", "blksize", "ino", "size", "blocks"} {
				if v := arg(args, i); v.Valid() {
					o.Set(name, v)
				}
			}
			for i, pair := range [][2]string{{"atime", "atimeMs"}, {"mtime", "mtimeMs"}, {"ctime", "ctimeMs"}, {"birthtime", "birthtimeMs"}} {
				v := arg(args, 10+i)
				if !v.Valid() {
					continue
				}
				o.Set(pair[1], v)
				if v.IsNumber() {
					o.Set(pair[0], m.date(v.Float()))
				}
			}
			o.SetHidden("_kind", e.JS.Str("file"))
		}
		return o.V(), nil
	}))
	exports.Set("Dirent", m.ctorClass("Dirent", m.direntProto, func(this js.Object, args []js.Value) (js.Value, error) {
		o := e.JS.Obj()
		o.SetProto(m.direntProto)
		return o.V(), nil
	}))
	// StatFs mirrors Node's class: the constructor records the eight fields and
	// statfs()/statfsSync() return instances of it.
	exports.Set("StatFs", m.ctorClass("StatFs", m.statfsProto, func(this js.Object, args []js.Value) (js.Value, error) {
		o := e.JS.Obj()
		o.SetProto(m.statfsProto)
		for i, name := range []string{"type", "bsize", "frsize", "blocks", "bfree", "bavail", "files", "ffree"} {
			if v := arg(args, i); v.Valid() {
				o.Set(name, v)
			}
		}
		return o.V(), nil
	}))
	m.installExtras(exports)
	// Node exposes the native realpath under realpath.native and
	// realpathSync.native as well as the bare names.
	for _, name := range []string{"realpath", "realpathSync"} {
		if fn := exports.Get(name); fn.Valid() && fn.IsFunction() {
			fn.Object().Set("native", fn)
		}
	}
	return exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"fs.Dirent",
		"fs.Dirent#isBlockDevice",
		"fs.Dirent#isCharacterDevice",
		"fs.Dirent#isDirectory",
		"fs.Dirent#isFIFO",
		"fs.Dirent#isFile",
		"fs.Dirent#isSocket",
		"fs.Dirent#isSymbolicLink",
		"fs.Dirent.name",
		"fs.Dirent.parentPath",
		"fs.Stats",
		"fs.Stats#isBlockDevice",
		"fs.Stats#isCharacterDevice",
		"fs.Stats#isDirectory",
		"fs.Stats#isFIFO",
		"fs.Stats#isFile",
		"fs.Stats#isSocket",
		"fs.Stats#isSymbolicLink",
		"fs.Stats.atime",
		"fs.Stats.atimeMs",
		"fs.Stats.birthtime",
		"fs.Stats.birthtimeMs",
		"fs.Stats.blksize",
		"fs.Stats.blocks",
		"fs.Stats.ctime",
		"fs.Stats.ctimeMs",
		"fs.Stats.dev",
		"fs.Stats.gid",
		"fs.Stats.ino",
		"fs.Stats.mode",
		"fs.Stats.mtime",
		"fs.Stats.mtimeMs",
		"fs.Stats.nlink",
		"fs.Stats.rdev",
		"fs.Stats.size",
		"fs.Stats.uid",
		"fs.accessSync",
		"fs.appendFileSync",
		"fs.chmod",
		"fs.chmodSync",
		"fs.chown",
		"fs.chownSync",
		"fs.close",
		"fs.closeSync",
		"fs.constants",
		"fs.Dir",
		"fs.Dir#[Symbol.asyncIterator]",
		"fs.Dir#close",
		"fs.Dir#closeSync",
		"fs.Dir#read",
		"fs.Dir#readSync",
		"fs.Dir.path",
		"fs.ReadStream",
		"fs.ReadStream.bytesRead",
		"fs.ReadStream.path",
		"fs.ReadStream.pending",
		"fs.WriteStream",
		"fs.WriteStream#close",
		"fs.WriteStream.bytesWritten",
		"fs.WriteStream.path",
		"fs.WriteStream.pending",
		"fs.copyFileSync",
		"fs.cpSync",
		"fs.createReadStream",
		"fs.createWriteStream",
		"fs.exists",
		"fs.existsSync",
		"fs.fchmod",
		"fs.fchmodSync",
		"fs.fchown",
		"fs.fchownSync",
		"fs.fdatasync",
		"fs.fdatasyncSync",
		"fs.fstat",
		"fs.fstatSync",
		"fs.fsync",
		"fs.fsyncSync",
		"fs.ftruncate",
		"fs.ftruncateSync",
		"fs.futimes",
		"fs.futimesSync",
		"fs.lchown",
		"fs.lchownSync",
		"fs.link",
		"fs.linkSync",
		"fs.lstatSync",
		"fs.mkdirSync",
		"fs.mkdtempSync",
		"fs.openSync",
		"fs.opendir",
		"fs.opendirSync",
		"fs.promises.FileHandle",
		"fs.promises.FileHandle#appendFile",
		"fs.promises.FileHandle#chmod",
		"fs.promises.FileHandle#chown",
		"fs.promises.FileHandle#close",
		"fs.promises.FileHandle#datasync",
		"fs.promises.FileHandle#read",
		"fs.promises.FileHandle#readFile",
		"fs.promises.FileHandle#readv",
		"fs.promises.FileHandle#stat",
		"fs.promises.FileHandle#sync",
		"fs.promises.FileHandle#truncate",
		"fs.promises.FileHandle#utimes",
		"fs.promises.FileHandle#write",
		"fs.promises.FileHandle#writeFile",
		"fs.promises.FileHandle#writev",
		"fs.promises.FileHandle.fd",
		"fs.promises.access",
		"fs.promises.appendFile",
		"fs.promises.chmod",
		"fs.promises.opendir",
		"fs.promises.watch",
		"fs.promises.chown",
		"fs.promises.constants",
		"fs.promises.copyFile",
		"fs.promises.cp",
		"fs.promises.lchmod",
		"fs.promises.lchown",
		"fs.promises.link",
		"fs.promises.lstat",
		"fs.promises.mkdir",
		"fs.promises.mkdtemp",
		"fs.promises.open",
		"fs.promises.readFile",
		"fs.promises.readdir",
		"fs.promises.readlink",
		"fs.promises.realpath",
		"fs.promises.rename",
		"fs.promises.rm",
		"fs.promises.rmdir",
		"fs.promises.stat",
		"fs.promises.statfs",
		"fs.promises.symlink",
		"fs.promises.truncate",
		"fs.promises.unlink",
		"fs.promises.utimes",
		"fs.promises.writeFile",
		"fs.read",
		"fs.readFile",
		"fs.readFileSync",
		"fs.readSync",
		"fs.readv",
		"fs.readvSync",
		"fs.readdir",
		"fs.readdirSync",
		"fs.readlinkSync",
		"fs.realpathSync",
		"fs.renameSync",
		"fs.rmSync",
		"fs.rmdirSync",
		"fs.stat",
		"fs.statSync",
		"fs.statfs",
		"fs.statfsSync",
		"fs.symlinkSync",
		"fs.truncateSync",
		"fs.unlinkSync",
		"fs.utimes",
		"fs.unwatchFile",
		"fs.utimesSync",
		"fs.watch",
		"fs.watchFile",
		"fs.write",
		"fs.writeFile",
		"fs.writeFileSync",
		"fs.writeSync",
		"fs.writev",
		"fs.writevSync",
		// The remaining tracked entries, including the extras installed from
		// extras.go (Utf8Stream, glob, StatWatcher, FileHandle helpers).
		"fs.access",
		"fs.appendFile",
		"fs.copyFile",
		"fs.cp",
		"fs.Dir#[Symbol.asyncDispose]",
		"fs.Dir#[Symbol.dispose]",
		"fs.FSWatcher",
		"fs.FSWatcher#close",
		"fs.FSWatcher#ref",
		"fs.FSWatcher#unref",
		"fs.StatFs",
		"fs.StatFs.bavail",
		"fs.StatFs.bfree",
		"fs.StatFs.blocks",
		"fs.StatFs.bsize",
		"fs.StatFs.ffree",
		"fs.StatFs.files",
		"fs.StatFs.frsize",
		"fs.StatFs.type",
		"fs.StatWatcher",
		"fs.StatWatcher#ref",
		"fs.StatWatcher#unref",
		"fs.Stats.atimeNs",
		"fs.Stats.birthtimeNs",
		"fs.Stats.ctimeNs",
		"fs.Stats.mtimeNs",
		"fs.Utf8Stream",
		"fs.Utf8Stream#[Symbol.dispose]",
		"fs.Utf8Stream#destroy",
		"fs.Utf8Stream#end",
		"fs.Utf8Stream#flush",
		"fs.Utf8Stream#flushSync",
		"fs.Utf8Stream#reopen",
		"fs.Utf8Stream#write",
		"fs.Utf8Stream.append",
		"fs.Utf8Stream.contentMode",
		"fs.Utf8Stream.fd",
		"fs.Utf8Stream.file",
		"fs.Utf8Stream.fsync",
		"fs.Utf8Stream.maxLength",
		"fs.Utf8Stream.minLength",
		"fs.Utf8Stream.mkdir",
		"fs.Utf8Stream.mode",
		"fs.Utf8Stream.periodicFlush",
		"fs.Utf8Stream.sync",
		"fs.Utf8Stream.writing",
		"fs.glob",
		"fs.globSync",
		"fs.lchmod",
		"fs.lchmodSync",
		"fs.lstat",
		"fs.lutimes",
		"fs.lutimesSync",
		"fs.mkdir",
		"fs.mkdtemp",
		"fs.mkdtempDisposableSync",
		"fs.native",
		"fs.open",
		"fs.openAsBlob",
		"fs.promises.FileHandle#[Symbol.asyncDispose]",
		"fs.promises.FileHandle#createReadStream",
		"fs.promises.FileHandle#createWriteStream",
		"fs.promises.FileHandle#pull",
		"fs.promises.FileHandle#pullSync",
		"fs.promises.FileHandle#readLines",
		"fs.promises.FileHandle#readableWebStream",
		"fs.promises.FileHandle#writer",
		"fs.promises.glob",
		"fs.promises.lutimes",
		"fs.promises.mkdtempDisposable",
		"fs.readlink",
		"fs.realpath",
		"fs.rename",
		"fs.rm",
		"fs.rmdir",
		"fs.symlink",
		"fs.truncate",
		"fs.unlink",
	}
}

// ---------------------------------------------------------------- objects --

func (m *module) class(name string, proto js.Object, call func(js.Object, []js.Value) (js.Value, error)) js.Value {
	holder := m.e.JS.Obj()
	holder.SetFunc(name, call)
	ctor := holder.Get(name).Object()
	ctor.Set("prototype", proto.V())
	proto.Set("constructor", ctor.V())
	return ctor.V()
}

// ctorClass builds a class whose instances come from call. A Go function
// value is not constructible in JavaScriptCore, so the exported class is a JS
// wrapper that delegates to the Go implementation, keeping `new fs.Stats(...)`
// working for the tests that build stat records by hand.
func (m *module) ctorClass(name string, proto js.Object, call func(js.Object, []js.Value) (js.Value, error)) js.Value {
	holder := m.e.JS.Obj()
	holder.SetFunc("impl", call)
	makeClass, err := m.e.JS.Eval(`(function (impl, proto, name) {
  var Cls = function (...args) { return impl.apply(this, args); };
  Object.defineProperty(Cls, 'name', { value: name, configurable: true });
  Cls.prototype = proto;
  Object.defineProperty(proto, 'constructor', {
    value: Cls, writable: true, configurable: true, enumerable: false
  });
  return Cls;
})`, "[gode:fs/class]")
	if err != nil {
		panic("godejs fs: " + err.Error())
	}
	v, cerr := makeClass.Object().Call(js.Object{}, holder.Get("impl"), proto.V(), m.e.JS.Str(name))
	if cerr != nil {
		panic("godejs fs: " + cerr.Error())
	}
	return v
}

func (m *module) makeStatFsProto() js.Object {
	return m.e.JS.Obj()
}

func (m *module) makeStatsProto() js.Object {
	p := m.e.JS.Obj()
	// The predicates read the recorded file kind rather than `mode`, because a
	// BigIntStats stores `mode` as a BigInt and must answer identically.
	is := func(kind string) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			return m.e.JS.Bool(this.Get("_kind").Str() == kind), nil
		}
	}
	p.SetFunc("isFile", is("file"))
	p.SetFunc("isDirectory", is("dir"))
	p.SetFunc("isSymbolicLink", is("symlink"))
	p.SetFunc("isBlockDevice", is("block"))
	p.SetFunc("isCharacterDevice", is("char"))
	p.SetFunc("isFIFO", is("fifo"))
	p.SetFunc("isSocket", is("socket"))
	return p
}

// modeKind names the S_IFMT class of a stat(2) mode, matching the values
// kindOf produces for io/fs modes.
func modeKind(mode uint32) string {
	switch mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		return "dir"
	case syscall.S_IFLNK:
		return "symlink"
	case syscall.S_IFCHR:
		return "char"
	case syscall.S_IFBLK:
		return "block"
	case syscall.S_IFIFO:
		return "fifo"
	case syscall.S_IFSOCK:
		return "socket"
	}
	return "file"
}

func (m *module) makeDirentProto() js.Object {
	p := m.e.JS.Obj()
	kind := func(this js.Object) string { return this.Get("_kind").Str() }
	p.SetFunc("isFile", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "file"), nil
	})
	p.SetFunc("isDirectory", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "dir"), nil
	})
	p.SetFunc("isSymbolicLink", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "symlink"), nil
	})
	p.SetFunc("isBlockDevice", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "block"), nil
	})
	p.SetFunc("isCharacterDevice", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "char"), nil
	})
	p.SetFunc("isFIFO", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "fifo"), nil
	})
	p.SetFunc("isSocket", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(kind(this) == "socket"), nil
	})
	return p
}

func (m *module) stats(fi os.FileInfo) js.Object { return m.statsWith(fi, false) }

// statsWith builds a Stats (or BigIntStats when big is set) from an FileInfo.
func (m *module) statsWith(fi os.FileInfo, big bool) js.Object {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return m.statsFromStatWith(st, big)
	}
	o := m.e.JS.Obj()
	o.SetProto(m.statsProto)
	o.SetHidden("_kind", m.e.JS.Str(kindOf(fi.Mode())))
	mod := fi.ModTime()
	o.SetNum("mode", float64(fi.Mode().Perm()))
	o.SetNum("size", float64(fi.Size()))
	o.SetNum("atimeMs", ms(mod))
	o.SetNum("mtimeMs", ms(mod))
	o.SetNum("ctimeMs", ms(mod))
	o.SetNum("birthtimeMs", ms(mod))
	o.Set("atime", m.date(ms(mod)))
	o.Set("mtime", m.date(ms(mod)))
	o.Set("ctime", m.date(ms(mod)))
	o.Set("birthtime", m.date(ms(mod)))
	return o
}

func (m *module) statsFromStat(st *syscall.Stat_t) js.Object {
	return m.statsFromStatWith(st, false)
}

// statsFromStatWith renders the fields of Stats / BigIntStats. Node exposes the
// nanosecond fields (atimeNs, ...) as BigInt in both shapes, and the *Ms fields
// as BigInt only for BigIntStats.
func (m *module) statsFromStatWith(st *syscall.Stat_t, big bool) js.Object {
	o := m.e.JS.Obj()
	o.SetProto(m.statsProto)
	o.SetHidden("_kind", m.e.JS.Str(modeKind(st.Mode)))
	atime := time.Unix(st.Atim.Sec, st.Atim.Nsec)
	mtime := time.Unix(st.Mtim.Sec, st.Mtim.Nsec)
	ctime := time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
	set := func(k string, v int64) {
		if big {
			o.Set(k, m.bigint(v))
			return
		}
		o.SetNum(k, float64(v))
	}
	set("dev", int64(st.Dev))
	set("ino", int64(st.Ino))
	set("mode", int64(st.Mode))
	set("nlink", int64(st.Nlink))
	set("uid", int64(st.Uid))
	set("gid", int64(st.Gid))
	set("rdev", int64(st.Rdev))
	set("size", int64(st.Size))
	set("blksize", int64(st.Blksize))
	set("blocks", int64(st.Blocks))
	stamp := func(name string, t time.Time) {
		o.Set(name, m.date(ms(t)))
		// Node keeps the nanosecond fields out of Object.keys(stats): they are
		// prototype getters, not own enumerable properties like the rest.
		o.SetHidden(name+"Ns", m.bigint(t.UnixNano()))
		if big {
			o.Set(name+"Ms", m.bigint(int64(ms(t))))
			return
		}
		o.SetNum(name+"Ms", ms(t))
	}
	stamp("atime", atime)
	stamp("mtime", mtime)
	stamp("ctime", ctime)
	// Linux has no creation time in stat(2); Node reports ctime for birthtime.
	stamp("birthtime", ctime)
	return o
}

func ms(t time.Time) float64 { return float64(t.UnixNano()) / 1e6 }

func (m *module) date(ms float64) js.Value {
	ctor := m.e.Global().Get("Date")
	reflect := m.e.Global().Get("Reflect")
	if reflect.IsObject() {
		construct := reflect.Object().Get("construct")
		if construct.IsFunction() {
			args := m.e.JS.Arr([]js.Value{ctor, m.e.JS.Arr([]js.Value{m.e.JS.Num(ms)}).V()})
			if v, err := construct.Object().Call(js.Object{}, args.At(0), args.At(1)); err == nil {
				return v
			}
		}
	}
	return m.e.JS.Undefined()
}

func (m *module) dirent(de os.DirEntry, parent string) js.Object {
	o := m.e.JS.Obj()
	o.SetProto(m.direntProto)
	o.SetStr("name", de.Name())
	o.SetStr("parentPath", parent)
	// `path` is the deprecated alias of parentPath that Node still exposes.
	o.SetStr("path", parent)
	o.SetHidden("_kind", m.e.JS.Str(kindOf(de.Type())))
	return o
}

func kindOf(mode iofs.FileMode) string {
	switch {
	case mode.IsDir():
		return "dir"
	case mode&iofs.ModeSymlink != 0:
		return "symlink"
	case mode&iofs.ModeDevice != 0 && mode&iofs.ModeCharDevice != 0:
		return "char"
	case mode&iofs.ModeDevice != 0:
		return "block"
	case mode&iofs.ModeNamedPipe != 0:
		return "fifo"
	case mode&iofs.ModeSocket != 0:
		return "socket"
	default:
		return "file"
	}
}

func (m *module) constants() js.Object {
	// Node's fs.constants has a null prototype; test-fs-constants checks it.
	o := m.e.JS.Obj()
	if v, err := m.e.JS.Eval("Object.create(null)", "[fs:constants]"); err == nil && v.IsObject() {
		o = v.Object()
	}
	values := map[string]int{
		"F_OK": 0, "R_OK": 4, "W_OK": 2, "X_OK": 1,
		"O_RDONLY": syscall.O_RDONLY, "O_WRONLY": syscall.O_WRONLY, "O_RDWR": syscall.O_RDWR,
		"O_CREAT": syscall.O_CREAT, "O_TRUNC": syscall.O_TRUNC, "O_APPEND": syscall.O_APPEND,
		"O_EXCL": syscall.O_EXCL, "O_SYNC": syscall.O_SYNC, "O_NOFOLLOW": syscall.O_NOFOLLOW,
		"O_DIRECTORY": syscall.O_DIRECTORY, "O_NOCTTY": syscall.O_NOCTTY,
		"O_NONBLOCK": syscall.O_NONBLOCK, "O_DSYNC": syscall.O_DSYNC,
		"COPYFILE_EXCL": 1, "COPYFILE_FICLONE": 2, "COPYFILE_FICLONE_FORCE": 4,
		// Node also exposes the libuv spellings on fs.constants; tests and
		// libraries use them interchangeably with the COPYFILE_* names.
		"UV_FS_COPYFILE_EXCL": 1, "UV_FS_COPYFILE_FICLONE": 2, "UV_FS_COPYFILE_FICLONE_FORCE": 4,
		"S_IFMT": syscall.S_IFMT, "S_IFREG": syscall.S_IFREG, "S_IFDIR": syscall.S_IFDIR,
		"S_IFCHR": syscall.S_IFCHR, "S_IFBLK": syscall.S_IFBLK, "S_IFIFO": syscall.S_IFIFO,
		"S_IFLNK": syscall.S_IFLNK, "S_IFSOCK": syscall.S_IFSOCK,
		"S_IRWXU": syscall.S_IRWXU, "S_IRUSR": syscall.S_IRUSR, "S_IWUSR": syscall.S_IWUSR,
		"S_IXUSR": syscall.S_IXUSR, "S_IRWXG": syscall.S_IRWXG, "S_IRGRP": syscall.S_IRGRP,
		"S_IWGRP": syscall.S_IWGRP, "S_IXGRP": syscall.S_IXGRP,
		"S_IRWXO": syscall.S_IRWXO, "S_IROTH": syscall.S_IROTH, "S_IWOTH": syscall.S_IWOTH,
		"S_IXOTH": syscall.S_IXOTH,
	}
	for k, v := range values {
		o.SetNum(k, float64(v))
	}
	return o
}

// ------------------------------------------------------------- arguments --

func arg(args []js.Value, i int) js.Value {
	if i >= len(args) {
		return js.Value{}
	}
	return args[i]
}

func optObject(args []js.Value, i int) js.Object {
	v := arg(args, i)
	if v.Valid() && v.IsObject() && !v.IsNull() && !v.IsFunction() {
		return v.Object()
	}
	return js.Object{}
}

func (m *module) pathArg(args []js.Value, i int, name string) (string, error) {
	v := arg(args, i)
	if v.Valid() && v.IsString() {
		s := v.Str()
		if strings.IndexByte(s, 0) >= 0 {
			return "", m.invalidArgValue(name, v, "must be a string, Uint8Array, or URL without null bytes")
		}
		return s, nil
	}
	if b, ok := v.Bytes(); ok {
		if strings.IndexByte(string(b), 0) >= 0 {
			return "", m.invalidArgValue(name, v, "must be a string, Uint8Array, or URL without null bytes")
		}
		return string(b), nil
	}
	if v.Valid() && v.IsObject() && !v.IsNull() {
		o := v.Object()
		// A WHATWG URL for any other scheme is rejected with
		// ERR_INVALID_URL_SCHEME, not a generic type error.
		if proto := o.Get("protocol"); proto.Valid() && proto.IsString() && proto.Str() != "" && !strings.HasPrefix(proto.Str(), "file:") {
			return "", js.ThrowCode("TypeError", "ERR_INVALID_URL_SCHEME", "The URL must be of scheme file")
		}
		if strings.HasPrefix(o.Get("protocol").Str(), "file:") {
			p := o.Get("pathname").Str()
			decoded, derr := fileURLPath(p)
			if derr != nil {
				return "", derr
			}
			if strings.IndexByte(decoded, 0) >= 0 {
				return "", m.invalidArgValue(name, v, "must be a string, Uint8Array, or URL without null bytes")
			}
			return decoded, nil
		}
	}
	return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", fmt.Sprintf("The %q argument must be of type string or an instance of Buffer or URL. Received %s", name, describe(v)))
}

func fileURLPath(p string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		if p[i] != '%' || i+2 >= len(p) {
			b.WriteByte(p[i])
			continue
		}
		hi, ok1 := hexNibble(p[i+1])
		lo, ok2 := hexNibble(p[i+2])
		if !ok1 || !ok2 {
			b.WriteByte(p[i])
			continue
		}
		b.WriteByte(hi<<4 | lo)
		i += 2
	}
	return b.String(), nil
}

func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// describe builds the `Received ...` clause of a Node argument error. Node's
// validate* helpers print the value for primitives and the constructor name for
// objects, and the official tests compare those strings exactly.
func describe(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return fmt.Sprintf("type boolean (%t)", v.Bool())
	case v.IsNumber():
		f := v.Float()
		switch {
		case math.IsNaN(f):
			return "type number (NaN)"
		case math.IsInf(f, 1):
			return "type number (Infinity)"
		case math.IsInf(f, -1):
			return "type number (-Infinity)"
		case f == math.Trunc(f) && math.Abs(f) < 1e21:
			return fmt.Sprintf("type number (%d)", int64(f))
		default:
			return fmt.Sprintf("type number (%s)", strconv.FormatFloat(f, 'g', -1, 64))
		}
	case v.IsString():
		// Node's validators quote the received string with single quotes
		// (`Received type string ('')`), not Go's %q double quotes.
		return "type string ('" + v.Str() + "')"
	case v.IsFunction():
		// Node's validators print `function ` plus the function's name, which
		// is empty for anonymous functions (the official tests compare it).
		if n := v.Object().Get("name"); n.IsString() {
			return "function " + n.Str()
		}
		return "function "
	case v.IsArray():
		return "an instance of Array"
	case v.IsObject():
		name := ""
		if ctor := v.Object().Get("constructor"); ctor.Valid() && ctor.IsObject() {
			if n := ctor.Object().Get("name"); n.IsString() {
				name = n.Str()
			}
		}
		// Arrays from another realm may not satisfy IsArray; their constructor
		// still names them Array, and Node reports `an instance of Array`.
		if name == "Array" {
			return "an instance of Array"
		}
		// Typed arrays and Buffers print like util.inspect: Name(len) [items].
		typedArray := name == "Buffer" || strings.HasSuffix(name, "Array")
		if n := v.Object().Get("length"); typedArray && n.Valid() && n.IsNumber() {
			total := int(n.Float())
			limit := total
			if limit > 100 {
				limit = 100
			}
			items := make([]string, 0, limit)
			for i := 0; i < limit; i++ {
				items = append(items, strconv.FormatFloat(v.Object().At(i).Float(), 'f', -1, 64))
			}
			inner := strings.Join(items, ", ")
			if total > limit {
				inner += ", ... " + strconv.Itoa(total-limit) + " more items"
			}
			return name + "(" + strconv.Itoa(total) + ") [" + inner + "]"
		}
		if name != "" {
			return "an instance of " + name
		}
		return "an instance of Object"
	default:
		return "an unsupported value"
	}
}

func normalizeEncoding(enc string) (string, error) {
	switch strings.ToLower(enc) {
	case "", "buffer":
		return "", nil
	case "utf8", "utf-8":
		return "utf8", nil
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		return "utf16le", nil
	case "latin1", "binary":
		return "latin1", nil
	case "ascii":
		return "ascii", nil
	case "base64":
		return "base64", nil
	case "base64url":
		return "base64url", nil
	case "hex":
		return "hex", nil
	default:
		// Node's fs layer reports invalid encodings as ERR_INVALID_ARG_VALUE
		// (assertEncoding), not as Buffer's ERR_UNKNOWN_ENCODING.
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'encoding' is invalid encoding. Received '"+enc+"'")
	}
}

// encodingOpt parses the encoding an options slot carries, keeping def when the
// caller supplies none. A bare string is the encoding; "buffer" comes back as
// the empty encoding so encode() returns a Buffer.
func (m *module) encodingOpt(args []js.Value, i int, def string) (string, error) {
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() || v.IsNull() || v.IsFunction() {
		return def, nil
	}
	if v.IsString() {
		return normalizeEncoding(v.Str())
	}
	if !v.IsObject() {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options" argument must be of type string or an instance of Object. Received `+describe(v))
	}
	o := v.Object()
	e := o.Get("encoding")
	if !e.Valid() || e.IsUndefined() || e.IsNull() {
		return def, nil
	}
	if !e.IsString() {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options.encoding" argument must be of type string. Received `+describe(e))
	}
	return normalizeEncoding(e.Str())
}

// readOpts parses the argument used by readFile/readdir: a string encoding or
// an options object with encoding and flag.
func (m *module) readOpts(args []js.Value, i int, defFlag string) (string, string, error) {
	flag := defFlag
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return "", flag, nil
	}
	if v.IsString() {
		enc, err := normalizeEncoding(v.Str())
		return enc, flag, err
	}
	if v.IsFunction() {
		// The callback occupies the options slot in callback-style calls.
		return "", flag, nil
	}
	if v.IsObject() && !v.IsFunction() {
		o := v.Object()
		enc := ""
		if e := o.Get("encoding"); e.Valid() && !e.IsUndefined() && !e.IsNull() {
			if !e.IsString() {
				return "", "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.encoding" argument must be of type string. Received `+describe(e))
			}
			var err error
			enc, err = normalizeEncoding(e.Str())
			if err != nil {
				return "", "", err
			}
		}
		if f := o.Get("flag"); f.Valid() && !f.IsUndefined() && !f.IsNull() {
			switch {
			case f.IsString():
				flag = f.Str()
			case f.IsNumber():
				flag = strconv.Itoa(f.Int())
			default:
				return "", "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.flag" argument must be of type string or an integer. Received `+describe(f))
			}
		}
		return enc, flag, nil
	}
	return "", "", js.Throw("TypeError", fmt.Sprintf("The \"options\" argument must be of type string or an instance of Object. Received %s", describe(v)))
}

func (m *module) modeOpt(args []js.Value, i int, def os.FileMode) os.FileMode {
	o := optObject(args, i)
	if !o.Valid() {
		return def
	}
	if v := o.Get("mode"); v.IsNumber() {
		return os.FileMode(v.Int())
	}
	return def
}

func boolOpt(args []js.Value, i int, key string) bool {
	o := optObject(args, i)
	if !o.Valid() {
		return false
	}
	return o.Get(key).Bool()
}

func intArg(args []js.Value, i, def int) int {
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return def
	}
	return v.Int()
}

func (m *module) dataArg(v js.Value, enc string) ([]byte, error) {
	if v.Valid() && v.IsString() {
		return encodeString(v.Str(), enc)
	}
	if b, ok := v.Bytes(); ok {
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"data\" argument must be of type string or an instance of Buffer, TypedArray, or DataView")
}

// bytesValue returns a Buffer when the buffer builtin is installed, falling
// back to a plain Uint8Array. Node's fs always returns Buffers.
func (m *module) bytesValue(data []byte) js.Value {
	ctor := m.e.Global().Get("Buffer")
	if ctor.Valid() && ctor.IsFunction() {
		from := ctor.Object().Get("from")
		if from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), m.e.JS.Bytes(data).V()); err == nil {
				return v
			}
		}
	}
	return m.e.JS.Bytes(data).V()
}

func (m *module) encode(data []byte, enc string) (js.Value, error) {
	switch enc {
	case "":
		return m.bytesValue(data), nil
	case "utf8":
		return m.e.JS.Str(string(data)), nil
	case "utf16le":
		units := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			units = append(units, uint16(data[i])|uint16(data[i+1])<<8)
		}
		return m.e.JS.Str(string(utf16.Decode(units))), nil
	case "latin1":
		var b strings.Builder
		for _, c := range data {
			b.WriteRune(rune(c))
		}
		return m.e.JS.Str(b.String()), nil
	case "ascii":
		var b strings.Builder
		for _, c := range data {
			b.WriteRune(rune(c & 0x7f))
		}
		return m.e.JS.Str(b.String()), nil
	case "base64":
		return m.e.JS.Str(base64.StdEncoding.EncodeToString(data)), nil
	case "base64url":
		return m.e.JS.Str(base64.RawURLEncoding.EncodeToString(data)), nil
	case "hex":
		return m.e.JS.Str(hex.EncodeToString(data)), nil
	}
	return m.e.JS.Undefined(), nil
}

func encodeString(s, enc string) ([]byte, error) {
	switch enc {
	case "", "utf8":
		return []byte(s), nil
	case "utf16le":
		units := utf16.Encode([]rune(s))
		out := make([]byte, 0, len(units)*2)
		for _, u := range units {
			out = append(out, byte(u), byte(u>>8))
		}
		return out, nil
	case "latin1":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			out = append(out, byte(r))
		}
		return out, nil
	case "ascii":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			out = append(out, byte(r)&0x7f)
		}
		return out, nil
	case "base64":
		clean := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, s)
		for len(clean)%4 != 0 {
			clean += "="
		}
		b, err := base64.StdEncoding.DecodeString(clean)
		if err != nil {
			return nil, js.Throw("TypeError", "The input must be a valid base64 string")
		}
		return b, nil
	case "base64url":
		clean := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, s)
		for len(clean)%4 != 0 {
			clean += "="
		}
		b, err := base64.URLEncoding.DecodeString(clean)
		if err != nil {
			return nil, js.Throw("TypeError", "The input must be a valid base64url string")
		}
		return b, nil
	case "hex":
		clean := s
		if len(clean) > 0 && len(clean)%2 != 0 {
			clean = clean[:len(clean)-1]
		}
		b, err := hex.DecodeString(clean)
		if err != nil {
			return nil, js.Throw("TypeError", "The input must be a valid hex string")
		}
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", "Unknown encoding: "+enc)
}

// ---------------------------------------------------------------- errors --

func errnoCode(errno syscall.Errno) string {
	switch errno {
	case syscall.ENOENT:
		return "ENOENT"
	case syscall.EACCES:
		return "EACCES"
	case syscall.EEXIST:
		return "EEXIST"
	case syscall.ENOTDIR:
		return "ENOTDIR"
	case syscall.EISDIR:
		return "EISDIR"
	case syscall.ENOTEMPTY:
		return "ENOTEMPTY"
	case syscall.EBADF:
		return "EBADF"
	case syscall.EINVAL:
		return "EINVAL"
	case syscall.ELOOP:
		return "ELOOP"
	case syscall.ENAMETOOLONG:
		return "ENAMETOOLONG"
	case syscall.EMFILE:
		return "EMFILE"
	case syscall.ENOSPC:
		return "ENOSPC"
	case syscall.EPERM:
		return "EPERM"
	case syscall.EROFS:
		return "EROFS"
	case syscall.EXDEV:
		return "EXDEV"
	case syscall.ENFILE:
		return "ENFILE"
	case syscall.ESPIPE:
		return "ESPIPE"
	}
	return fmt.Sprintf("ERRNO%d", int(errno))
}

// fsError is a Node-shaped filesystem error. Node's fs errors carry the errno
// code in the message and expose it as `code`, `errno`, `syscall`, `path` and
// (for two-path operations) `dest`; the official tests assert those fields.
type fsError struct {
	Code    string
	Message string
	Errno   int
	Syscall string
	Path    string
	Dest    string
}

func (e *fsError) Error() string { return e.Message }

// object renders the error the way Node exposes it to JavaScript.
func (e *fsError) object(c *js.Context) js.Object {
	o := c.Error(e.Message)
	o.SetStr("code", e.Code)
	o.SetNum("errno", float64(e.Errno))
	if e.Syscall != "" {
		o.SetStr("syscall", e.Syscall)
	}
	if e.Path != "" {
		o.SetStr("path", e.Path)
	}
	if e.Dest != "" {
		o.SetStr("dest", e.Dest)
	}
	return o
}

// JSErrorObject lets the bridge build the full fs error (code, errno,
// syscall, path) for promise rejections and uncaught exceptions instead of
// degrading to a plain Error with only the message.
func (e *fsError) JSErrorObject(c *js.Context) js.Object { return e.object(c) }

// nodeErr converts a Go/OS error into the shape Node throws: an Error whose
// message starts with the stable errno code.
func (m *module) nodeErr(err error, op, path string) error {
	return m.nodeErrDest(err, op, path, "")
}

// nodeErrDest is nodeErr for the operations that take two paths (rename, link,
// copyFile, symlink); Node reports the second one as `dest`.
func (m *module) nodeErrDest(err error, op, path, dest string) error {
	if err == nil {
		return nil
	}
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return err
	}
	code := errnoCode(errno)
	text := errno.Error()
	// libuv words EISDIR differently from Go's syscall package, and the
	// official tests compare the message text.
	if errno == syscall.EISDIR {
		text = "illegal operation on a directory"
	}
	// libuv's strerror text for EEXIST is "file already exists"; Go's syscall
	// package says just "file exists".
	if errno == syscall.EEXIST {
		text = "file already exists"
	}
	msg := code + ": " + text + ", " + op
	if path != "" {
		msg += " '" + path + "'"
		if dest != "" {
			msg += " -> '" + dest + "'"
		}
	}
	return &fsError{Code: code, Message: msg, Errno: -int(errno), Syscall: op, Path: path, Dest: dest}
}

func (m *module) errorValue(err error) js.Value {
	if err == nil {
		return m.e.JS.Null()
	}
	if fe, ok := err.(*fsError); ok {
		return fe.object(m.e.JS).V()
	}
	obj := m.e.JS.Error(err.Error())
	if je, ok := err.(*js.Error); ok {
		if je.Name != "" {
			obj.SetStr("name", je.Name)
		}
		if je.Code != "" {
			obj.SetStr("code", je.Code)
		}
	}
	return obj.V()
}

// syncCall rethrows a rich filesystem error as a real JavaScript error. The
// host bridge only copies `name` and `code` from a Go error, so a synchronous
// API has to raise the populated object itself for `err.path`/`err.syscall`
// to be observable, the way Node's tests expect.
func (m *module) syncCall(fn func([]js.Value) (js.Value, error)) js.Func {
	return func(this js.Object, args []js.Value) (js.Value, error) {
		v, err := fn(args)
		if fe, ok := err.(*fsError); ok && m.throwing.Valid() {
			props := m.e.JS.Obj()
			props.SetStr("message", fe.Message)
			props.SetStr("code", fe.Code)
			props.SetStr("syscall", fe.Syscall)
			props.SetStr("path", fe.Path)
			props.SetStr("dest", fe.Dest)
			props.SetNum("errno", float64(fe.Errno))
			if _, thrown := m.throwing.Object().Call(js.Object{}, props.V()); thrown != nil {
				return js.Value{}, thrown
			}
			// The helper above always throws; if it ever does not, keep the code
			// observable rather than degrading to a bare Error.
			return js.Value{}, js.ThrowCode("Error", fe.Code, fe.Message)
		}
		return v, err
	}
}

// --------------------------------------------------------------- helpers --

func readAll(path, flag string) ([]byte, error) {
	if flag == "" || flag == "r" {
		return os.ReadFile(path)
	}
	fd, err := parseFlags(flag)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, fd, 0o666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func parseFlags(flag string) (int, error) {
	// Numeric flags are the platform's open(2) bits and are passed through
	// unchanged, as Node does for `{ flag: fs.constants.O_WRONLY | ... }`.
	if flag != "" {
		if n, err := strconv.Atoi(flag); err == nil && n >= 0 {
			return n, nil
		}
	}
	switch flag {
	case "", "r", "rs", "sr":
		return syscall.O_RDONLY, nil
	case "r+", "rs+", "sr+":
		return syscall.O_RDWR, nil
	case "w", "wx", "xw":
		f := syscall.O_WRONLY | syscall.O_CREAT | syscall.O_TRUNC
		if flag != "w" {
			f |= syscall.O_EXCL
		}
		return f, nil
	case "w+", "wx+", "xw+":
		f := syscall.O_RDWR | syscall.O_CREAT | syscall.O_TRUNC
		if flag != "w+" {
			f |= syscall.O_EXCL
		}
		return f, nil
	case "a", "ax", "xa":
		f := syscall.O_WRONLY | syscall.O_CREAT | syscall.O_APPEND
		if flag != "a" {
			f |= syscall.O_EXCL
		}
		return f, nil
	case "a+", "ax+", "xa+":
		f := syscall.O_RDWR | syscall.O_CREAT | syscall.O_APPEND
		if flag != "a+" {
			f |= syscall.O_EXCL
		}
		return f, nil
	}
	return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", fmt.Sprintf("The argument 'flags' is invalid. Received '%s'", flag))
}

func flagArg(args []js.Value, i int, def string) (string, error) {
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return def, nil
	}
	if v.IsString() {
		return v.Str(), nil
	}
	if v.IsNumber() {
		return fmt.Sprintf("%d", v.Int()), nil
	}
	return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", fmt.Sprintf("The argument 'flags' is invalid. Received %s", describe(v)))
}

// writeAll writes data with the requested flag. flush mirrors Node 24's
// `flush: true` option by fsyncing before the descriptor is closed.
func (m *module) writeAll(path string, data []byte, flag string, mode os.FileMode, flush bool) error {
	if (flag == "" || flag == "w") && !flush {
		return os.WriteFile(path, data, mode)
	}
	fd, err := parseFlags(flag)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, fd, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if flush {
		return f.Sync()
	}
	return nil
}

func copyFile(src, dst string, force bool) error {
	if !force {
		if _, err := os.Lstat(dst); err == nil {
			return &fsError{Code: "EEXIST", Message: "EEXIST: file already exists, copyfile '" + src + "' -> '" + dst + "'",
				Errno: -int(syscall.EEXIST), Syscall: "copyfile", Path: src, Dest: dst}
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	// Node's copyFile also reproduces the source permission bits when the
	// destination already existed.
	return os.Chmod(dst, info.Mode().Perm())
}

func copyTree(src, dst string, recursive, force, deref bool) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if deref && info.Mode()&os.ModeSymlink != 0 {
		if resolved, statErr := os.Stat(src); statErr == nil {
			info = resolved
		}
	}
	switch {
	case info.IsDir():
		if !recursive {
			return &js.Error{Code: "ERR_FS_EISDIR", Message: "Path is a directory: cp returned EISDIR (is a directory)"}
		}
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, de := range entries {
			if err := copyTree(filepath.Join(src, de.Name()), filepath.Join(dst, de.Name()), true, force, deref); err != nil {
				return err
			}
		}
		return nil
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		if !force {
			if _, err := os.Lstat(dst); err == nil {
				return &fsError{Code: "EEXIST", Message: "EEXIST: file already exists, symlink '" + src + "' -> '" + dst + "'",
					Errno: -int(syscall.EEXIST), Syscall: "symlink", Path: src, Dest: dst}
			}
		}
		return os.Symlink(target, dst)
	default:
		return copyFile(src, dst, force)
	}
}

// --------------------------------------------------------------- sync API --

func (m *module) installSync(exports js.Object) {
	sync := func(name string, fn func([]js.Value) (js.Value, error)) {
		exports.SetFunc(name+"Sync", m.syncCall(fn))
	}

	sync("readFile", m.readFileCore)
	sync("writeFile", m.writeFileCore)
	sync("appendFile", m.appendFileCore)
	sync("readdir", m.readdirCore)
	sync("stat", m.statCore)
	sync("lstat", m.lstatCore)
	sync("realpath", m.realpathCore)
	sync("readlink", m.readlinkCore)
	sync("mkdir", m.mkdirCore)
	sync("mkdtemp", m.mkdtempCore)
	sync("rm", m.rmCore)
	sync("rmdir", m.rmdirCore)
	sync("unlink", m.unlinkCore)
	sync("rename", m.renameCore)
	sync("copyFile", m.copyFileCore)
	sync("cp", m.cpCore)
	sync("truncate", m.truncateCore)
	sync("access", m.accessCore)
	sync("chmod", m.chmodCore)
	sync("utimes", m.utimesCore)
	sync("symlink", m.symlinkCore)
	sync("open", func(args []js.Value) (js.Value, error) {
		fd, err := m.openCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(fd)), nil
	})

	exports.SetFunc("existsSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		// Node swallows argument validation here: a path of the wrong type is
		// simply reported as missing rather than thrown.
		path, err := m.pathArg(args, 0, "path")
		if err != nil {
			return m.e.JS.Bool(false), nil
		}
		_, statErr := os.Stat(path)
		return m.e.JS.Bool(statErr == nil), nil
	}))

	exports.SetFunc("readSync", m.syncCall(m.readSyncCore))
	exports.SetFunc("writeSync", m.syncCall(m.writeSyncCore))
	exports.SetFunc("closeSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		if err := syscall.Close(arg(args, 0).Int()); err != nil {
			return js.Value{}, m.nodeErr(err, "close", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	exports.SetFunc("fstatSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		fd, err := m.fdArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.fstatWith(fd, statOption(args, "bigint", false))
	}))
	exports.SetFunc("ftruncateSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		fd, err := m.fdArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		length, err := m.lenArg(args, 1)
		if err != nil {
			return js.Value{}, err
		}
		if err := syscall.Ftruncate(fd, int64(length)); err != nil {
			return js.Value{}, m.nodeErr(err, "ftruncate", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	exports.SetFunc("fsyncSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		return m.syncFd(args, true)
	}))
	exports.SetFunc("fdatasyncSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		return m.syncFd(args, false)
	}))
}

func (m *module) syncFd(args []js.Value, full bool) (js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return js.Value{}, err
	}
	if full {
		err = syscall.Fsync(fd)
	} else {
		err = syscall.Fdatasync(fd)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "fsync", "")
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) readFileCore(args []js.Value) (js.Value, error) {
	if v := arg(args, 0); v.Valid() && v.IsNumber() {
		// Node also accepts a file descriptor; it reads from the current
		// position and leaves the descriptor open.
		enc, _, err := m.readOpts(args, 1, "r")
		if err != nil {
			return js.Value{}, err
		}
		data, err := readAllFd(v.Int())
		if err != nil {
			return js.Value{}, m.nodeErr(err, "read", "")
		}
		if buf, used, berr := m.fillReadFileBuffer(optObject(args, 1), data); berr != nil {
			return js.Value{}, berr
		} else if used {
			return m.encodeBufferValue(buf, enc)
		}
		return m.encode(data, enc)
	}
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	enc, flag, err := m.readOpts(args, 1, "r")
	if err != nil {
		return js.Value{}, err
	}
	data, err := readAll(path, flag)
	if err != nil {
		op := "read"
		var errno syscall.Errno
		if errors.As(err, &errno) && (errno == syscall.ENOENT || errno == syscall.EACCES || errno == syscall.ENOTDIR) {
			op = "open"
		}
		return js.Value{}, m.nodeErr(err, op, path)
	}
	if buf, used, berr := m.fillReadFileBuffer(optObject(args, 1), data); berr != nil {
		return js.Value{}, berr
	} else if used {
		return m.encodeBufferValue(buf, enc)
	}
	return m.encode(data, enc)
}

func (m *module) writeFileCore(args []js.Value) (js.Value, error) {
	if v := arg(args, 0); v.Valid() && v.IsNumber() {
		// Node also accepts a file descriptor: it writes at the current
		// position and leaves the descriptor open (no truncation).
		enc, _, err := m.readOpts(args, 2, "w")
		if err != nil {
			return js.Value{}, err
		}
		data, err := m.dataArg(arg(args, 1), enc)
		if err != nil {
			return js.Value{}, err
		}
		fd := v.Int()
		if _, err := writeAllFd(fd, data); err != nil {
			return js.Value{}, m.nodeErr(err, "write", "")
		}
		if opts := optObject(args, 2); opts.Valid() && opts.Get("flush").Bool() {
			if err := syscall.Fsync(fd); err != nil {
				return js.Value{}, m.nodeErr(err, "fsync", "")
			}
		}
		return m.e.JS.Undefined(), nil
	}
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	enc, flag, err := m.readOpts(args, 2, "w")
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.dataArg(arg(args, 1), enc)
	if err != nil {
		return js.Value{}, err
	}
	mode := m.modeOpt(args, 2, 0o666)
	flush := optObject(args, 2).Valid() && optObject(args, 2).Get("flush").Bool()
	if err := m.writeAll(path, data, flag, mode, flush); err != nil {
		return js.Value{}, m.nodeErr(err, "open", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) appendFileCore(args []js.Value) (js.Value, error) {
	if v := arg(args, 0); v.Valid() && v.IsNumber() {
		enc, _, err := m.readOpts(args, 2, "a")
		if err != nil {
			return js.Value{}, err
		}
		data, err := m.dataArg(arg(args, 1), enc)
		if err != nil {
			return js.Value{}, err
		}
		fd := v.Int()
		var st syscall.Stat_t
		if err := syscall.Fstat(fd, &st); err != nil {
			return js.Value{}, m.nodeErr(err, "fstat", "")
		}
		if _, err := syscall.Pwrite(fd, data, st.Size); err != nil {
			return js.Value{}, m.nodeErr(err, "write", "")
		}
		return m.e.JS.Undefined(), nil
	}
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	enc, flag, err := m.readOpts(args, 2, "a")
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.dataArg(arg(args, 1), enc)
	if err != nil {
		return js.Value{}, err
	}
	mode := m.modeOpt(args, 2, 0o666)
	flush := optObject(args, 2).Valid() && optObject(args, 2).Get("flush").Bool()
	if err := m.writeAll(path, data, flag, mode, flush); err != nil {
		return js.Value{}, m.nodeErr(err, "open", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) readdirCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	withTypes := false
	recursive := false
	enc := ""
	if o := optObject(args, 1); o.Valid() {
		if v := o.Get("withFileTypes"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
			if !v.IsBool() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.withFileTypes" argument must be of type boolean. Received `+describe(v))
			}
			withTypes = v.Bool()
		}
		if v := o.Get("recursive"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
			if !v.IsBool() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.recursive" argument must be of type boolean. Received `+describe(v))
			}
			recursive = v.Bool()
		}
		if e := o.Get("encoding"); e.Valid() && !e.IsUndefined() && !e.IsNull() {
			if !e.IsString() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.encoding" argument must be of type string. Received `+describe(e))
			}
			enc, err = readdirEncoding(e.Str())
			if err != nil {
				return js.Value{}, err
			}
		}
	} else if v := arg(args, 1); v.Valid() && v.IsString() {
		enc, err = readdirEncoding(v.Str())
		if err != nil {
			return js.Value{}, err
		}
	}
	// Collect first so a recursive walk reports the same scandir error shape as
	// a flat one, and so the returned order is depth-first like Node's.
	type dirEntry struct {
		de     os.DirEntry
		parent string
		name   string
	}
	var list []dirEntry
	var walk func(dir, rel string) error
	walk = func(dir, rel string) error {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return readErr
		}
		for _, de := range entries {
			name := de.Name()
			if rel != "" {
				name = rel + "/" + name
			}
			list = append(list, dirEntry{de: de, parent: dir, name: name})
			if recursive && de.IsDir() {
				if err := walk(filepath.Join(dir, de.Name()), name); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(path, ""); err != nil {
		return js.Value{}, m.nodeErr(err, "scandir", path)
	}
	renderName := func(name string) js.Value {
		switch enc {
		case "", "utf8":
			return m.e.JS.Str(name)
		case "buffer":
			return m.bytesValue([]byte(name))
		default:
			return m.encodeNoErr([]byte(name), enc)
		}
	}
	arr := m.e.JS.Arr(nil)
	for i, entry := range list {
		if withTypes {
			arr.SetAt(i, m.dirent(entry.de, entry.parent).V())
			continue
		}
		arr.SetAt(i, renderName(entry.name))
	}
	return arr.V(), nil
}

// readdirEncoding is normalizeEncoding plus readdir's `buffer` encoding, which
// returns the entry names as Buffers rather than strings.
func readdirEncoding(raw string) (string, error) {
	if strings.EqualFold(raw, "buffer") {
		return "buffer", nil
	}
	return normalizeEncoding(raw)
}

func (m *module) encodeNoErr(data []byte, enc string) js.Value {
	v, _ := m.encode(data, enc)
	return v
}

func (m *module) statCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	fi, err := os.Stat(path)
	if err != nil {
		if statOption(args, "throwIfNoEntry", true) == false && errors.Is(err, os.ErrNotExist) {
			return m.e.JS.Undefined(), nil
		}
		return js.Value{}, m.nodeErr(err, "stat", path)
	}
	return m.statsWith(fi, statOption(args, "bigint", false)).V(), nil
}

// statOption reads a boolean stat option, applying Node's default when absent.
func statOption(args []js.Value, name string, def bool) bool {
	if o := optObject(args, 1); o.Valid() {
		if v := o.Get(name); v.Valid() && v.IsBool() {
			return v.Bool()
		}
	}
	return def
}

func (m *module) lstatCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	fi, err := os.Lstat(path)
	if err != nil {
		if statOption(args, "throwIfNoEntry", true) == false && errors.Is(err, os.ErrNotExist) {
			return m.e.JS.Undefined(), nil
		}
		return js.Value{}, m.nodeErr(err, "lstat", path)
	}
	return m.statsWith(fi, statOption(args, "bigint", false)).V(), nil
}

func (m *module) fstatCore(fd int) (js.Value, error) { return m.fstatWith(fd, false) }

func (m *module) fstatWith(fd int, big bool) (js.Value, error) {
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		return js.Value{}, m.nodeErr(err, "fstat", "")
	}
	return m.statsFromStatWith(&st, big).V(), nil
}

func (m *module) realpathCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	// realpath defaults to a string result; 'buffer' (or a 'buffer' option)
	// returns the same path bytes as a Buffer.
	enc, err := m.encodingOpt(args, 1, "utf8")
	if err != nil {
		return js.Value{}, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return js.Value{}, m.nodeErr(err, "realpath", path)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// Go reports symlink loops as a plain error; Node surfaces ELOOP.
		if strings.Contains(err.Error(), "too many links") {
			return js.Value{}, &fsError{Code: "ELOOP", Message: "ELOOP: too many symbolic links encountered, realpath '" + path + "'",
				Errno: -int(syscall.ELOOP), Syscall: "realpath", Path: path}
		}
		// Node reports realpath as the syscall for both missing paths and
		// other resolution failures (test-fs-error-messages asserts it).
		return js.Value{}, m.nodeErr(err, "realpath", path)
	}
	return m.encode([]byte(resolved), enc)
}

func (m *module) readlinkCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	enc, err := m.encodingOpt(args, 1, "utf8")
	if err != nil {
		return js.Value{}, err
	}
	target, err := os.Readlink(path)
	if err != nil {
		return js.Value{}, m.nodeErr(err, "readlink", path)
	}
	return m.encode([]byte(target), enc)
}

func (m *module) mkdirCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	o := optObject(args, 1)
	recursive := false
	if o.Valid() {
		if rv := o.Get("recursive"); rv.Valid() && !rv.IsUndefined() && !rv.IsNull() {
			if !rv.IsBool() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.recursive" property must be of type boolean. Received `+describe(rv))
			}
			recursive = rv.Bool()
		}
	}
	// The second mkdir argument is either an options object or a bare mode
	// (number or octal string), and the mode applies only below 0o1000.
	mode := os.FileMode(0o777)
	if v := arg(args, 1); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		switch {
		case v.IsNumber():
			mode = os.FileMode(uint32(v.Int())) & os.ModePerm
		case v.IsString():
			n, perr := strconv.ParseUint(v.Str(), 8, 32)
			if perr != nil {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
					"The argument 'mode' must be a 32-bit unsigned integer or an octal string. Received '"+v.Str()+"'")
			}
			mode = os.FileMode(uint32(n)) & os.ModePerm
		default:
			mode = m.modeOpt(args, 1, 0o777)
		}
	}
	if !recursive {
		if err := os.Mkdir(path, mode); err != nil {
			return js.Value{}, m.nodeErr(err, "mkdir", path)
		}
		return m.e.JS.Undefined(), nil
	}
	first := ""
	for cur := filepath.Clean(path); ; cur = filepath.Dir(cur) {
		if _, err := os.Stat(cur); err == nil {
			break
		}
		first = cur
		if parent := filepath.Dir(cur); parent == cur {
			break
		}
	}
	if st, statErr := os.Stat(path); statErr == nil && !st.IsDir() {
		// mkdir -p over an existing file is EEXIST, not ENOTDIR.
		return js.Value{}, m.nodeErr(syscall.EEXIST, "mkdir", path)
	}
	if err := os.MkdirAll(path, mode); err != nil {
		return js.Value{}, m.nodeErr(err, "mkdir", path)
	}
	if first == "" {
		return m.e.JS.Undefined(), nil
	}
	return m.e.JS.Str(first), nil
}

// mkdtempWarned mirrors Node's once-per-process mkdtemp portability warning.
var mkdtempWarned bool

func (m *module) mkdtempCore(args []js.Value) (js.Value, error) {
	prefix, err := m.pathArg(args, 0, "prefix")
	if err != nil {
		return js.Value{}, err
	}
	enc, err := m.encodingOpt(args, 1, "utf8")
	if err != nil {
		return js.Value{}, err
	}
	if strings.HasSuffix(prefix, "X") && !strings.HasSuffix(prefix, "XXXXXX") && !mkdtempWarned {
		mkdtempWarned = true
		if p := m.e.Exports("process"); p.Valid() {
			if w := p.Get("emitWarning"); w.IsFunction() {
				_, _ = w.Object().Call(p, m.e.JS.Str("mkdtemp() templates ending with X are not portable. For details see: https://nodejs.org/api/fs.html"), m.e.JS.Str("Warning"))
			}
		}
	}
	// Node replaces the last six X placeholders (adding them when the prefix
	// does not end with any) with random alphanumerics: the basename grows by
	// exactly six characters. os.MkdirTemp's nine-character suffix would not
	// match that length.
	template := prefix
	if !strings.HasSuffix(template, "XXXXXX") {
		template += "XXXXXX"
	}
	base := template[:len(template)-6]
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for attempt := 0; attempt < 100; attempt++ {
		suffix := make([]byte, 6)
		for i := range suffix {
			suffix[i] = letters[rand.Intn(len(letters))]
		}
		path := base + string(suffix)
		mkErr := os.Mkdir(path, 0o700)
		if mkErr == nil {
			return m.encode([]byte(path), enc)
		}
		if !errors.Is(mkErr, iofs.ErrExist) {
			return js.Value{}, m.nodeErr(mkErr, "mkdtemp", prefix)
		}
	}
	return js.Value{}, m.nodeErr(syscall.EEXIST, "mkdtemp", prefix)
}

func (m *module) rmCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	o := optObject(args, 1)
	recursive := o.Valid() && o.Get("recursive").Bool()
	force := o.Valid() && o.Get("force").Bool()
	if recursive {
		// Node stats first unless `force` is set, so a missing path is ENOENT
		// rather than a silent success.
		if !force {
			if _, statErr := os.Lstat(path); statErr != nil {
				// Node's rm stats first; a missing path reports the lstat
				// syscall, not rm.
				return js.Value{}, m.nodeErr(statErr, "lstat", path)
			}
		}
		if err := removeAll(path); err != nil {
			return js.Value{}, m.nodeErr(err, "rm", path)
		}
		return m.e.JS.Undefined(), nil
	}
	// fs.rm without recursive must never remove a directory: Node reports
	// ERR_FS_EISDIR rather than falling back to rmdir(2).
	st, statErr := os.Lstat(path)
	if statErr != nil {
		if force && errors.Is(statErr, os.ErrNotExist) {
			return m.e.JS.Undefined(), nil
		}
		return js.Value{}, m.nodeErr(statErr, "lstat", path)
	}
	if st.IsDir() {
		return js.Value{}, &fsError{Code: "ERR_FS_EISDIR",
			Message: "Path is a directory: rm returned EISDIR (is a directory) '" + path + "'",
			Errno:   -int(syscall.EISDIR), Syscall: "rm", Path: path}
	}
	if err := os.Remove(path); err != nil {
		if force && errors.Is(err, os.ErrNotExist) {
			return m.e.JS.Undefined(), nil
		}
		return js.Value{}, m.nodeErr(err, "rm", path)
	}
	return m.e.JS.Undefined(), nil
}

// removeAll deletes path and everything below it. Node's fs.rm chmods
// directories that lack write permission instead of failing with EACCES or
// ENOTEMPTY, and several official tests remove read-only trees.
func removeAll(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	mode := info.Mode()
	if mode&os.ModeSymlink != 0 || !mode.IsDir() {
		return os.Remove(path)
	}
	if mode.Perm()&0o700 != 0o700 {
		_ = os.Chmod(path, mode.Perm()|0o700)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, de := range entries {
		if err := removeAll(filepath.Join(path, de.Name())); err != nil {
			return err
		}
	}
	return os.Remove(path)
}

// rmdirRecursiveWarned mirrors Node's once-per-process DEP0147 warning.
var rmdirRecursiveWarned bool

func (m *module) emitRecursiveRmdirWarning() {
	if rmdirRecursiveWarned {
		return
	}
	rmdirRecursiveWarned = true
	p := m.e.Exports("process")
	if !p.Valid() {
		return
	}
	if nd := p.Get("noDeprecation"); nd.Valid() && nd.Bool() {
		return
	}
	w := p.Get("emitWarning")
	if !w.Valid() || !w.IsFunction() {
		return
	}
	_, _ = w.Object().Call(p,
		m.e.JS.Str("In future versions of Node.js, fs.rmdir(path, { recursive: true }) will be removed. Use fs.rm(path, { recursive: true }) instead"),
		m.e.JS.Str("DeprecationWarning"),
		m.e.JS.Str("DEP0147"))
}

func (m *module) rmdirCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	// Node still honours the deprecated { recursive: true } form of rmdir,
	// which it implements through rm(force: false): missing paths error and a
	// regular file is ENOTDIR because the recursive form expects a directory.
	if o := optObject(args, 1); o.Valid() && o.Get("recursive").Bool() {
		m.emitRecursiveRmdirWarning()
		st, statErr := os.Lstat(path)
		if statErr != nil {
			return js.Value{}, m.nodeErr(statErr, "rmdir", path)
		}
		if !st.IsDir() {
			return js.Value{}, m.nodeErr(syscall.ENOTDIR, "rmdir", path)
		}
		if err := removeAll(path); err != nil {
			return js.Value{}, m.nodeErr(err, "rmdir", path)
		}
		return m.e.JS.Undefined(), nil
	}
	// rmdir(2), not remove(3): a regular file must fail with ENOTDIR.
	if err := syscall.Rmdir(path); err != nil {
		return js.Value{}, m.nodeErr(err, "rmdir", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) unlinkCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	// unlink(2) rejects directories with EISDIR; os.Remove would delete an
	// empty one.
	if err := syscall.Unlink(path); err != nil {
		return js.Value{}, m.nodeErr(err, "unlink", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) renameCore(args []js.Value) (js.Value, error) {
	from, err := m.pathArg(args, 0, "oldPath")
	if err != nil {
		return js.Value{}, err
	}
	to, err := m.pathArg(args, 1, "newPath")
	if err != nil {
		return js.Value{}, err
	}
	if err := os.Rename(from, to); err != nil {
		return js.Value{}, m.nodeErrDest(err, "rename", from, to)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) copyFileCore(args []js.Value) (js.Value, error) {
	src, err := m.pathArg(args, 0, "src")
	if err != nil {
		return js.Value{}, err
	}
	dst, err := m.pathArg(args, 1, "dest")
	if err != nil {
		return js.Value{}, err
	}
	// Node validates the copy mode before touching the filesystem: a
	// non-integer is ERR_INVALID_ARG_TYPE and anything outside the three
	// libuv copyfile bits is ERR_OUT_OF_RANGE.
	// The callback occupies the mode slot in copyFile(src, dest, cb).
	flags := 0
	if v := arg(args, 2); v.Valid() && !v.IsUndefined() && !v.IsNull() && !v.IsFunction() {
		if !v.IsNumber() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The \"mode\" argument must be of type integer. Received %s", describe(v)))
		}
		flags = v.Int()
		if flags < 0 || flags > 7 {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				fmt.Sprintf("The value of \"mode\" is out of range. It must be >= 0 && <= 7. Received %d", flags))
		}
	}
	force := flags&1 == 0
	if err := copyFile(src, dst, force); err != nil {
		return js.Value{}, m.nodeErrDest(err, "copyfile", src, dst)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) cpCore(args []js.Value) (js.Value, error) {
	src, err := m.pathArg(args, 0, "src")
	if err != nil {
		return js.Value{}, err
	}
	dst, err := m.pathArg(args, 1, "dest")
	if err != nil {
		return js.Value{}, err
	}
	o := optObject(args, 2)
	recursive := o.Valid() && o.Get("recursive").Bool()
	deref := o.Valid() && o.Get("dereference").Bool()
	force := true
	if o.Valid() {
		if v := o.Get("force"); v.Valid() && v.IsBool() {
			force = v.Bool()
		}
	}
	if err := copyTree(src, dst, recursive, force, deref); err != nil {
		return js.Value{}, m.nodeErr(err, "cp", src)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) truncateCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	length, err := m.lenArg(args, 1)
	if err != nil {
		return js.Value{}, err
	}
	if err := os.Truncate(path, int64(length)); err != nil {
		return js.Value{}, m.nodeErr(err, "truncate", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) accessCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	mode := arg(args, 1)
	if mode.Valid() && !mode.IsUndefined() && !mode.IsNull() && !mode.IsFunction() {
		if !mode.IsNumber() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "mode" argument must be of type number. Received `+describe(mode))
		}
		f := mode.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "mode" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 || f > 7 {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "mode" is out of range. It must be >= 0 && <= 7. Received `+strconv.FormatFloat(f, 'f', -1, 64))
		}
		if err := syscall.Access(path, uint32(f)); err != nil {
			return js.Value{}, m.nodeErr(err, "access", path)
		}
		return m.e.JS.Undefined(), nil
	}
	if _, err := os.Stat(path); err != nil {
		return js.Value{}, m.nodeErr(err, "access", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) chmodCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	mode, err := m.modeArg(args, 1)
	if err != nil {
		return js.Value{}, err
	}
	if err := os.Chmod(path, mode); err != nil {
		return js.Value{}, m.nodeErr(err, "chmod", path)
	}
	return m.e.JS.Undefined(), nil
}

// modeArg accepts the numeric and octal-string forms Node allows for modes.
func (m *module) modeArg(args []js.Value, i int) (os.FileMode, error) {
	v := arg(args, i)
	if v.Valid() && v.IsNumber() {
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "mode" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f < 0 || f > 4294967295 {
			return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "mode" is out of range. It must be >= 0 && <= 4294967295. Received `+strconv.FormatFloat(f, 'f', -1, 64))
		}
		return os.FileMode(uint32(f)), nil
	}
	if v.Valid() && v.IsString() {
		s := v.Str()
		n, err := strconv.ParseUint(s, 8, 32)
		if err != nil || s == "" {
			return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				"The argument 'mode' must be a 32-bit unsigned integer or an octal string. Received '"+s+"'")
		}
		return os.FileMode(uint32(n)), nil
	}
	return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "mode" argument must be of type number. Received `+describe(v))
}

func (m *module) linkCore(args []js.Value) (js.Value, error) {
	oldPath, err := m.pathArg(args, 0, "oldPath")
	if err != nil {
		return js.Value{}, err
	}
	newPath, err := m.pathArg(args, 1, "newPath")
	if err != nil {
		return js.Value{}, err
	}
	if err := os.Link(oldPath, newPath); err != nil {
		return js.Value{}, m.nodeErrDest(err, "link", oldPath, newPath)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) chownCore(args []js.Value, follow bool) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	uid, gid, err := m.idArgs(args)
	if err != nil {
		return js.Value{}, err
	}
	if follow {
		err = os.Chown(path, uid, gid)
	} else {
		err = os.Lchown(path, uid, gid)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "chown", path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) idArgs(args []js.Value) (int, int, error) {
	id := func(i int, name string) (int, error) {
		v := arg(args, i)
		if !v.Valid() || !v.IsNumber() || v.Float() != math.Trunc(v.Float()) {
			return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "`+name+`" argument must be of type number. Received `+describe(v))
		}
		n := v.Int()
		if n < -1 || n > 0xFFFFFFFF {
			return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "`+name+`" is out of range. It must be >= -1 && <= 4294967295. Received `+strconv.Itoa(n))
		}
		return n, nil
	}
	uid, err := id(1, "uid")
	if err != nil {
		return 0, 0, err
	}
	gid, err := id(2, "gid")
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

func (m *module) fchownCore(args []js.Value) (js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return js.Value{}, err
	}
	uid, gid, err := m.idArgs(args)
	if err != nil {
		return js.Value{}, err
	}
	if err := syscall.Fchown(fd, uid, gid); err != nil {
		return js.Value{}, m.nodeErr(err, "fchown", "")
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) fchmodCore(args []js.Value) (js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return js.Value{}, err
	}
	mode, err := m.modeArg(args, 1)
	if err != nil {
		return js.Value{}, err
	}
	if err := syscall.Fchmod(fd, uint32(mode)); err != nil {
		return js.Value{}, m.nodeErr(err, "fchmod", "")
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) bigint(n int64) js.Value {
	f := m.e.Global().Get("BigInt")
	if f.Valid() && f.IsFunction() {
		if v, err := f.Object().Call(js.Object{}, m.e.JS.Str(strconv.FormatInt(n, 10))); err == nil {
			return v
		}
	}
	return m.e.JS.Num(float64(n))
}

func (m *module) statfsCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return js.Value{}, m.nodeErr(err, "statfs", path)
	}
	big := false
	if o := optObject(args, 1); o.Valid() {
		big = o.Get("bigint").Bool()
	}
	o := m.e.JS.Obj()
	o.SetProto(m.statfsProto)
	set := func(k string, v int64) {
		if big {
			o.Set(k, m.bigint(v))
			return
		}
		o.SetNum(k, float64(v))
	}
	set("type", int64(st.Type))
	set("bsize", st.Bsize)
	set("frsize", st.Frsize)
	set("blocks", int64(st.Blocks))
	set("bfree", int64(st.Bfree))
	set("bavail", int64(st.Bavail))
	set("files", int64(st.Files))
	set("ffree", int64(st.Ffree))
	return o.V(), nil
}

func (m *module) symlinkCore(args []js.Value) (js.Value, error) {
	target, err := m.pathArg(args, 0, "target")
	if err != nil {
		return js.Value{}, err
	}
	path, err := m.pathArg(args, 1, "path")
	if err != nil {
		return js.Value{}, err
	}
	// Windows-only type hints are still validated on every platform.
	// fs.symlink(target, path, callback) passes the callback in the type slot.
	if t := arg(args, 2); t.Valid() && !t.IsUndefined() && !t.IsNull() && !t.IsFunction() {
		if s := t.Str(); !t.IsString() || (s != "dir" && s != "file" && s != "junction") {
			received := describe(t)
			if t.IsString() {
				received = "'" + s + "'"
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				"The argument 'type' must be one of: 'dir', 'file', 'junction'. Received "+received)
		}
	}
	if err := os.Symlink(target, path); err != nil {
		return js.Value{}, m.nodeErrDest(err, "symlink", target, path)
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) openCore(args []js.Value) (int, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return 0, err
	}
	flag := "r"
	if v := arg(args, 1); v.Valid() && !v.IsUndefined() && !v.IsNull() && !v.IsFunction() {
		if v.IsString() {
			flag = v.Str()
		} else if v.IsNumber() {
			flag = fmt.Sprintf("%d", v.Int())
		} else {
			return 0, js.Throw("TypeError", fmt.Sprintf("The \"flags\" argument must be of type string or an integer. Received %s", describe(v)))
		}
	}
	flags, err := parseFlags(flag)
	if err != nil {
		if flag != "" && flag[0] >= '0' && flag[0] <= '9' {
			// Numeric flags are passed straight to open(2).
			var n int
			if _, scanErr := fmt.Sscanf(flag, "%d", &n); scanErr == nil {
				flags = n
				err = nil
			}
		}
		if err != nil {
			return 0, err
		}
	}
	mode := 0o666
	if v := arg(args, 2); v.Valid() && !v.IsUndefined() && !v.IsNull() && !v.IsFunction() {
		switch {
		case v.IsNumber():
			mode = v.Int()
		case v.IsString():
			// Node accepts the octal string form of a mode, e.g. '10644'.
			n, perr := strconv.ParseUint(v.Str(), 8, 32)
			if perr != nil || v.Str() == "" {
				return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
					"The argument 'mode' must be a 32-bit unsigned integer or an octal string. Received '"+v.Str()+"'")
			}
			mode = int(uint32(n))
		default:
			return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "mode" argument must be of type number. Received `+describe(v))
		}
	}
	fd, err := syscall.Open(path, flags, uint32(mode))
	if err != nil {
		return 0, m.nodeErr(err, "open", path)
	}
	return fd, nil
}

func (m *module) readSyncCore(args []js.Value) (js.Value, error) {
	fd := arg(args, 0).Int()
	bufVal := arg(args, 1)
	if _, ok := bufVal.Bytes(); !ok {
		return js.Value{}, js.Throw("TypeError", "The \"buffer\" argument must be an instance of Buffer or Uint8Array")
	}
	total := len(mustBytes(bufVal))
	offset := intArg(args, 2, 0)
	length := intArg(args, 3, total-offset)
	position := -1
	if v := arg(args, 4); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		position = v.Int()
	}
	if offset < 0 || offset > total || length < 0 || offset+length > total {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "The value of \"offset\" is out of range.")
	}
	tmp := make([]byte, length)
	var n int
	var err error
	if position >= 0 {
		n, err = syscall.Pread(fd, tmp, int64(position))
	} else {
		n, err = syscall.Read(fd, tmp)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "read", "")
	}
	if n > 0 {
		if err := m.fillBuffer(bufVal, tmp[:n], offset); err != nil {
			return js.Value{}, err
		}
	}
	return m.e.JS.Num(float64(n)), nil
}

func mustBytes(v js.Value) []byte {
	b, _ := v.Bytes()
	return b
}

func (m *module) fillBuffer(buf js.Value, data []byte, offset int) error {
	proto := m.e.Global().Get("Uint8Array").Object().Get("prototype").Object()
	set := proto.Get("set")
	if !set.IsFunction() {
		return js.Throw("TypeError", "The \"buffer\" argument must be an instance of Buffer or Uint8Array")
	}
	_, err := set.Object().Call(buf.Object(), m.e.JS.Bytes(data).V(), m.e.JS.Num(float64(offset)))
	return err
}

func (m *module) writeSyncCore(args []js.Value) (js.Value, error) {
	fd := arg(args, 0).Int()
	v := arg(args, 1)
	var data []byte
	var position int
	if v.Valid() && v.IsString() {
		enc := "utf8"
		if e := arg(args, 3); e.Valid() && e.IsString() {
			var err error
			enc, err = normalizeEncoding(e.Str())
			if err != nil {
				return js.Value{}, err
			}
		}
		var err error
		data, err = encodeString(v.Str(), enc)
		if err != nil {
			return js.Value{}, err
		}
		position = -1
		if p := arg(args, 2); p.Valid() && p.IsNumber() {
			position = p.Int()
		}
	} else {
		b, ok := v.Bytes()
		if !ok {
			return js.Value{}, js.Throw("TypeError", "The \"buffer\" argument must be an instance of Buffer or Uint8Array")
		}
		offset := intArg(args, 2, 0)
		length := intArg(args, 3, len(b)-offset)
		if offset < 0 || offset > len(b) || length < 0 || offset+length > len(b) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "The value of \"offset\" is out of range.")
		}
		data = b[offset : offset+length]
		position = -1
		if p := arg(args, 4); p.Valid() && p.IsNumber() {
			position = p.Int()
		}
	}
	var n int
	var err error
	if position >= 0 {
		n, err = syscall.Pwrite(fd, data, int64(position))
	} else {
		n, err = syscall.Write(fd, data)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "write", "")
	}
	return m.e.JS.Num(float64(n)), nil
}

// ----------------------------------------------------------- callback API --

// abortedSignal reports whether one of the options arguments carries an
// already aborted AbortSignal. Node checks the signal before issuing the
// filesystem request and fails with AbortError instead.
func abortedSignal(args []js.Value) bool {
	for _, a := range args {
		if !a.Valid() || !a.IsObject() || a.IsNull() || a.IsFunction() {
			continue
		}
		sig := a.Object().Get("signal")
		if !sig.Valid() || !sig.IsObject() || sig.IsNull() {
			continue
		}
		if ab := sig.Object().Get("aborted"); ab.Valid() && ab.IsBool() && ab.Bool() {
			return true
		}
	}
	return false
}

func abortErr() error {
	return js.ThrowCode("AbortError", "ABORT_ERR", "The operation was aborted")
}

func (m *module) installCallbacks(exports js.Object) {
	cb := func(name string, fn func([]js.Value) (js.Value, error), withValue bool) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			if _, err := callbackFunc(args); err != nil {
				return js.Value{}, err
			}
			if abortedSignal(args) {
				m.deliver(args, abortErr(), js.Value{}, withValue)
				return m.e.JS.Undefined(), nil
			}
			v, err := fn(args)
			if validationErr(err) {
				return js.Value{}, err
			}
			m.deliver(args, err, v, withValue)
			return m.e.JS.Undefined(), nil
		})
	}
	cb("readFile", m.readFileCore, true)
	cb("writeFile", m.writeFileCore, false)
	cb("appendFile", m.appendFileCore, false)
	cb("readdir", m.readdirCore, true)
	cb("stat", m.statCore, true)
	cb("lstat", m.lstatCore, true)
	cb("realpath", m.realpathCore, true)
	cb("readlink", m.readlinkCore, true)
	cb("mkdir", m.mkdirCore, true)
	cb("mkdtemp", m.mkdtempCore, true)
	cb("rm", m.rmCore, false)
	cb("rmdir", m.rmdirCore, false)
	cb("unlink", m.unlinkCore, false)
	cb("rename", m.renameCore, false)
	cb("copyFile", m.copyFileCore, false)
	cb("cp", m.cpCore, false)
	cb("truncate", m.truncateCore, false)
	cb("utimes", m.utimesCore, false)
	cb("access", m.accessCore, false)
	cb("chmod", m.chmodCore, false)
	cb("symlink", m.symlinkCore, false)
	cb("open", func(args []js.Value) (js.Value, error) {
		fd, err := m.openCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(fd)), nil
	}, true)
	// Node's bindings validate the descriptor before the callback, so a bad fd
	// must produce the fd error even when the callback is not a function.
	// The generic helper above looks for the callback first, so wrap the fd
	// operations it registered.
	for _, name := range []string{"fchmod", "fchown", "fdatasync", "fsync", "fstat", "ftruncate", "futimes"} {
		orig := exports.Get(name)
		if !orig.Valid() || !orig.IsFunction() {
			continue
		}
		fn := orig
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			if _, ferr := m.fdArg(args, 0); ferr != nil {
				return js.Value{}, ferr
			}
			return fn.Object().Call(this, args...)
		})
	}
	// fs.close(fd, cb) is declared here rather than through cb(): the generic
	// helper registers the callback before the descriptor, and Node validates
	// the descriptor first. It is also kept out of the descriptor loop above so
	// the implementation is installed exactly once.
	exports.SetFunc("close", func(this js.Object, args []js.Value) (js.Value, error) {
		fd, ferr := m.fdArg(args, 0)
		if ferr != nil {
			return js.Value{}, ferr
		}
		if _, cerr := callbackFunc(args); cerr != nil {
			return js.Value{}, cerr
		}
		err := syscall.Close(fd)
		if err != nil {
			err = m.nodeErr(err, "close", "")
		}
		m.deliver(args, err, js.Value{}, false)
		return m.e.JS.Undefined(), nil
	})

	// fchmod and fchown validate every numeric argument before the callback,
	// so a bad mode or uid/gid must report itself even without a function.
	if orig := exports.Get("fchmod"); orig.Valid() && orig.IsFunction() {
		fn := orig
		exports.SetFunc("fchmod", func(this js.Object, args []js.Value) (js.Value, error) {
			if _, err := m.fdArg(args, 0); err != nil {
				return js.Value{}, err
			}
			if _, err := m.modeArg(args, 1); err != nil {
				return js.Value{}, err
			}
			return fn.Object().Call(this, args...)
		})
	}
	if orig := exports.Get("fchown"); orig.Valid() && orig.IsFunction() {
		fn := orig
		exports.SetFunc("fchown", func(this js.Object, args []js.Value) (js.Value, error) {
			if _, err := m.fdArg(args, 0); err != nil {
				return js.Value{}, err
			}
			if _, _, err := m.idArgs(args); err != nil {
				return js.Value{}, err
			}
			return fn.Object().Call(this, args...)
		})
	}
	if orig := exports.Get("ftruncate"); orig.Valid() && orig.IsFunction() {
		fn := orig
		exports.SetFunc("ftruncate", func(this js.Object, args []js.Value) (js.Value, error) {
			if _, err := m.fdArg(args, 0); err != nil {
				return js.Value{}, err
			}
			if _, err := m.lenArg(args, 1); err != nil {
				return js.Value{}, err
			}
			return fn.Object().Call(this, args...)
		})
	}
}

// validationErr reports whether err is Node's synchronous argument-validation
// error (ERR_INVALID_ARG_TYPE and friends), which must be thrown instead of
// passed to the callback.
func validationErr(err error) bool {
	je, ok := err.(*js.Error)
	if !ok {
		return false
	}
	return strings.HasPrefix(je.Code, "ERR_") || je.Name == "TypeError" || je.Name == "RangeError"
}

func (m *module) deliver(args []js.Value, err error, v js.Value, withValue bool) {
	var fn js.Value
	for i := len(args) - 1; i >= 0; i-- {
		if args[i].IsFunction() {
			fn = args[i]
			break
		}
	}
	if !fn.Valid() {
		return
	}
	errVal := m.errorValue(err)
	if withValue {
		if !v.Valid() {
			v = m.e.JS.Undefined()
		}
		m.tick(fn, errVal, v)
		return
	}
	m.tick(fn, errVal)
}

// tick schedules a callback with `this` set to undefined, as Node does.
func (m *module) tick(fn js.Value, args ...js.Value) {
	if m.invoke.Valid() {
		all := make([]js.Value, 0, len(args)+1)
		all = append(all, fn)
		m.e.NextTick(m.invoke, append(all, args...)...)
		return
	}
	m.e.NextTick(fn, args...)
}

// ------------------------------------------------------------ promise API --

func (m *module) promise(fn func([]js.Value) (js.Value, error)) js.Func {
	return func(this js.Object, args []js.Value) (js.Value, error) {
		if abortedSignal(args) {
			return m.e.JS.Promise(js.Value{}, abortErr()), nil
		}
		v, err := fn(args)
		return m.e.JS.Promise(v, err), nil
	}
}

func (m *module) installPromises(exports js.Object) {
	p := m.e.JS.Obj()
	p.SetFunc("readFile", m.promise(m.readFileCore))
	p.SetFunc("writeFile", m.promise(m.writeFileCore))
	p.SetFunc("appendFile", m.promise(m.appendFileCore))
	p.SetFunc("readdir", m.promise(m.readdirCore))
	p.SetFunc("stat", m.promise(m.statCore))
	p.SetFunc("lstat", m.promise(m.lstatCore))
	p.SetFunc("realpath", m.promise(m.realpathCore))
	p.SetFunc("readlink", m.promise(m.readlinkCore))
	p.SetFunc("mkdir", m.promise(m.mkdirCore))
	p.SetFunc("mkdtemp", m.promise(m.mkdtempCore))
	p.SetFunc("rm", m.promise(m.rmCore))
	p.SetFunc("rmdir", m.promise(m.rmdirCore))
	p.SetFunc("unlink", m.promise(m.unlinkCore))
	p.SetFunc("rename", m.promise(m.renameCore))
	p.SetFunc("copyFile", m.promise(m.copyFileCore))
	p.SetFunc("cp", m.promise(m.cpCore))
	p.SetFunc("truncate", m.promise(m.truncateCore))
	p.SetFunc("access", m.promise(m.accessCore))
	p.SetFunc("chmod", m.promise(m.chmodCore))
	p.SetFunc("symlink", m.promise(m.symlinkCore))
	p.SetFunc("utimes", m.promise(m.utimesCore))
	p.SetFunc("open", func(this js.Object, args []js.Value) (js.Value, error) {
		fd, err := m.openCore(args)
		if err != nil {
			return m.e.JS.Promise(js.Value{}, err), nil
		}
		return m.e.JS.Promise(m.handle(fd).V(), nil), nil
	})
	p.SetObj("constants", m.constants())
	p.Set("FileHandle", m.class("FileHandle", m.makeFileHandleProto(), func(this js.Object, args []js.Value) (js.Value, error) {
		o := m.e.JS.Obj()
		o.SetProto(m.makeFileHandleProto())
		if len(args) > 0 {
			o.SetNum("fd", float64(args[0].Int()))
		}
		return o.V(), nil
	}))
	exports.SetObj("promises", p)
}

func (m *module) makeFileHandleProto() js.Object {
	// One prototype shared by every handle and by fs.promises.FileHandle.
	if m.fhProto.Valid() {
		return m.fhProto
	}
	o := m.e.JS.Obj()
	o.SetFunc("close", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		if this.Get("_closed").Bool() {
			return m.e.JS.Undefined(), nil
		}
		fd := this.Get("fd").Int()
		this.SetHidden("_closed", m.e.JS.Bool(true))
		this.SetNum("fd", -1)
		if err := syscall.Close(fd); err != nil {
			return js.Value{}, m.nodeErr(err, "close", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	o.SetFunc("readFile", m.handleMethod(m.handleReadFile))
	o.SetFunc("writeFile", m.handleMethod(m.handleWriteFile))
	o.SetFunc("appendFile", m.handleMethod(m.handleAppendFile))
	o.SetFunc("read", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		n, buf, err := m.handleReadCore(this, args)
		if err != nil {
			return js.Value{}, err
		}
		out := m.e.JS.Obj()
		out.SetNum("bytesRead", float64(n))
		out.Set("buffer", buf)
		return out.V(), nil
	}))
	o.SetFunc("write", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		n, buf, err := m.handleWriteCore(this, args)
		if err != nil {
			return js.Value{}, err
		}
		out := m.e.JS.Obj()
		out.SetNum("bytesWritten", float64(n))
		out.Set("buffer", buf)
		return out.V(), nil
	}))
	o.SetFunc("readv", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		n, bufs, err := m.handleReadv(this, args)
		if err != nil {
			return js.Value{}, err
		}
		out := m.e.JS.Obj()
		out.SetNum("bytesRead", float64(n))
		out.Set("buffers", m.e.JS.Arr(bufs).V())
		return out.V(), nil
	}))
	o.SetFunc("writev", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		n, bufs, err := m.handleWritev(this, args)
		if err != nil {
			return js.Value{}, err
		}
		out := m.e.JS.Obj()
		out.SetNum("bytesWritten", float64(n))
		out.Set("buffers", m.e.JS.Arr(bufs).V())
		return out.V(), nil
	}))
	o.SetFunc("stat", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		// FileHandle methods take the options object as their first argument.
		big := false
		if o := optObject(args, 0); o.Valid() {
			big = o.Get("bigint").Bool()
		}
		return m.fstatWith(this.Get("fd").Int(), big)
	}))
	o.SetFunc("truncate", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		if err := syscall.Ftruncate(this.Get("fd").Int(), int64(intArg(args, 0, 0))); err != nil {
			return js.Value{}, m.nodeErr(err, "ftruncate", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	o.SetFunc("sync", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		if err := syscall.Fsync(this.Get("fd").Int()); err != nil {
			return js.Value{}, m.nodeErr(err, "fsync", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	o.SetFunc("datasync", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		if err := syscall.Fdatasync(this.Get("fd").Int()); err != nil {
			return js.Value{}, m.nodeErr(err, "fdatasync", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	o.SetFunc("chmod", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		if err := syscall.Fchmod(this.Get("fd").Int(), uint32(intArg(args, 0, 0))); err != nil {
			return js.Value{}, m.nodeErr(err, "fchmod", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	o.SetFunc("utimes", m.handleMethod(m.handleUtimes))
	o.SetFunc("chown", m.handleMethod(func(this js.Object, args []js.Value) (js.Value, error) {
		uid, gid, err := m.idArgs(args)
		if err != nil {
			return js.Value{}, err
		}
		if err := syscall.Fchown(this.Get("fd").Int(), uid, gid); err != nil {
			return js.Value{}, m.nodeErr(err, "fchown", "")
		}
		return m.e.JS.Undefined(), nil
	}))
	return o
}

// handleMethod converts a core (args) function into a host method that runs
// against `this` and returns a settled Promise.
func (m *module) handleMethod(fn func(js.Object, []js.Value) (js.Value, error)) js.Func {
	return func(this js.Object, args []js.Value) (js.Value, error) {
		v, err := fn(this, args)
		return m.e.JS.Promise(v, err), nil
	}
}

func (m *module) handle(fd int) js.Object {
	fh := m.e.JS.Obj()
	fh.SetProto(m.makeFileHandleProto())
	fh.SetNum("fd", float64(fd))
	fh.SetHidden("_closed", m.e.JS.Bool(false))
	return fh
}

func (m *module) handleReadFile(this js.Object, args []js.Value) (js.Value, error) {
	enc, _, err := m.readOpts(args, 0, "r")
	if err != nil {
		return js.Value{}, err
	}
	data, err := readAllFd(this.Get("fd").Int())
	if err != nil {
		return js.Value{}, m.nodeErr(err, "read", "")
	}
	if buf, used, berr := m.fillReadFileBuffer(optObject(args, 0), data); berr != nil {
		return js.Value{}, berr
	} else if used {
		return m.encodeBufferValue(buf, enc)
	}
	return m.encode(data, enc)
}

// fillReadFileBuffer writes data into a caller-provided options.buffer and
// returns the value Node returns for that call: a Buffer view over the same
// memory, sliced to the bytes read. used reports whether the option existed.
func (m *module) fillReadFileBuffer(opts js.Object, data []byte) (js.Value, bool, error) {
	buf, ok, err := m.readFileBufferOption(opts, len(data))
	if err != nil || !ok {
		return js.Value{}, ok, err
	}
	if err := m.fillBuffer(buf, data, 0); err != nil {
		return js.Value{}, true, err
	}
	view, err := m.bufferSlice(buf, len(data))
	if err != nil {
		return js.Value{}, true, err
	}
	return view, true, nil
}

// encodeBufferValue applies an encoding to an already-materialised Buffer, or
// returns it unchanged when no encoding was requested.
func (m *module) encodeBufferValue(v js.Value, enc string) (js.Value, error) {
	if enc == "" {
		return v, nil
	}
	b, ok := v.Bytes()
	if !ok {
		return v, nil
	}
	return m.encode(b, enc)
}

// bufferSlice returns a Buffer over v's first n bytes. TypedArrays and
// DataViews are converted through Buffer.from(view.buffer, offset, length) so
// the result keeps sharing the caller's memory, exactly like Node's
// getReadFileBuffer normalisation.
func (m *module) bufferSlice(v js.Value, n int) (js.Value, error) {
	if !v.Valid() || !v.IsObject() {
		return v, nil
	}
	ctor := m.e.Global().Get("Buffer")
	if ctor.Valid() && ctor.IsFunction() {
		isBuffer := ctor.Object().Get("isBuffer")
		from := ctor.Object().Get("from")
		if isBuffer.IsFunction() && from.IsFunction() {
			if res, err := isBuffer.Object().Call(ctor.Object(), v); err == nil && !res.Bool() {
				o := v.Object()
				if conv, cerr := from.Object().Call(ctor.Object(), o.Get("buffer"), o.Get("byteOffset"), o.Get("byteLength")); cerr == nil {
					v = conv
				}
			}
		}
	}
	if !v.IsObject() {
		return v, nil
	}
	sub := v.Object().Get("subarray")
	if sub.IsFunction() {
		return sub.Object().Call(v.Object(), m.e.JS.Num(0), m.e.JS.Num(float64(n)))
	}
	return v, nil
}

// readFileBufferOption resolves a `buffer` option. It reports false when the
// option is absent.
func (m *module) readFileBufferOption(o js.Object, size int) (js.Value, bool, error) {
	if !o.Valid() {
		return js.Value{}, false, nil
	}
	b := o.Get("buffer")
	if !b.Valid() || b.IsUndefined() {
		return js.Value{}, false, nil
	}
	factory := false
	if b.IsFunction() {
		factory = true
		made, callErr := b.Object().Call(o, m.e.JS.Num(float64(size)))
		if callErr != nil {
			return js.Value{}, false, callErr
		}
		b = made
	}
	name := "options.buffer"
	if factory {
		name = "options.buffer()"
	}
	if _, isView := b.Bytes(); !isView {
		return js.Value{}, false, m.invalidArgType(name, "an instance of Buffer, TypedArray, or DataView", b)
	}
	n := len(mustBytes(b))
	if size > n {
		return js.Value{}, false, m.invalidArgValue(name+".byteLength", m.e.JS.Num(float64(n)),
			fmt.Sprintf("is smaller than the file size of %d bytes", size))
	}
	return b, true, nil
}

func (m *module) handleWriteFile(this js.Object, args []js.Value) (js.Value, error) {
	enc, _, err := m.readOpts(args, 1, "w")
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.dataArg(arg(args, 0), enc)
	if err != nil {
		return js.Value{}, err
	}
	fd := this.Get("fd").Int()
	if _, err := writeAllFd(fd, data); err != nil {
		return js.Value{}, m.nodeErr(err, "write", "")
	}
	if optObject(args, 1).Valid() && optObject(args, 1).Get("flush").Bool() {
		if err := syscall.Fsync(fd); err != nil {
			return js.Value{}, m.nodeErr(err, "fsync", "")
		}
	}
	return m.e.JS.Undefined(), nil
}

// readAllFd reads from the fd's current offset without taking ownership of it
// (os.NewFile would install a finalizer that closes the caller's fd).
func readAllFd(fd int) ([]byte, error) {
	var out []byte
	buf := make([]byte, 32*1024)
	for {
		n, err := syscall.Read(fd, buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			return out, err
		}
		if n == 0 {
			return out, nil
		}
	}
}

func writeAllFd(fd int, data []byte) (int, error) {
	total := 0
	for total < len(data) {
		n, err := syscall.Write(fd, data[total:])
		if err != nil {
			return total, err
		}
		if n == 0 {
			break
		}
		total += n
	}
	return total, nil
}

func (m *module) handleAppendFile(this js.Object, args []js.Value) (js.Value, error) {
	enc, _, err := m.readOpts(args, 1, "a")
	if err != nil {
		return js.Value{}, err
	}
	data, err := m.dataArg(arg(args, 0), enc)
	if err != nil {
		return js.Value{}, err
	}
	fd := this.Get("fd").Int()
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		return js.Value{}, m.nodeErr(err, "fstat", "")
	}
	if _, err := syscall.Pwrite(fd, data, st.Size); err != nil {
		return js.Value{}, m.nodeErr(err, "write", "")
	}
	if optObject(args, 1).Valid() && optObject(args, 1).Get("flush").Bool() {
		if err := syscall.Fsync(fd); err != nil {
			return js.Value{}, m.nodeErr(err, "fsync", "")
		}
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) withFd(this js.Object, args []js.Value) []js.Value {
	all := make([]js.Value, 0, len(args)+1)
	all = append(all, m.e.JS.Num(float64(this.Get("fd").Int())))
	return append(all, args...)
}

func (m *module) handleReadCore(this js.Object, args []js.Value) (int, js.Value, error) {
	return m.readCore(m.withFd(this, args))
}

func (m *module) handleWriteCore(this js.Object, args []js.Value) (int, js.Value, error) {
	return m.writeCore(m.withFd(this, args))
}

func (m *module) handleReadv(this js.Object, args []js.Value) (int, []js.Value, error) {
	return m.readvCore(m.withFd(this, args))
}

func (m *module) handleWritev(this js.Object, args []js.Value) (int, []js.Value, error) {
	return m.writevCore(m.withFd(this, args))
}

func (m *module) handleRead(this js.Object, args []js.Value) (js.Value, error) {
	fd := this.Get("fd").Int()
	bufVal := arg(args, 0)
	if _, ok := bufVal.Bytes(); !ok {
		return js.Value{}, js.Throw("TypeError", "The \"buffer\" argument must be an instance of Buffer or Uint8Array")
	}
	total := len(mustBytes(bufVal))
	offset := intArg(args, 1, 0)
	length := intArg(args, 2, total-offset)
	position := -1
	if v := arg(args, 3); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		position = v.Int()
	}
	if offset < 0 || offset > total || length < 0 || offset+length > total {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "The value of \"offset\" is out of range.")
	}
	tmp := make([]byte, length)
	var n int
	var err error
	if position >= 0 {
		n, err = syscall.Pread(fd, tmp, int64(position))
	} else {
		n, err = syscall.Read(fd, tmp)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "read", "")
	}
	if n > 0 {
		if err := m.fillBuffer(bufVal, tmp[:n], offset); err != nil {
			return js.Value{}, err
		}
	}
	out := m.e.JS.Obj()
	out.SetNum("bytesRead", float64(n))
	out.Set("buffer", bufVal)
	return out.V(), nil
}

func (m *module) handleWrite(this js.Object, args []js.Value) (js.Value, error) {
	fd := this.Get("fd").Int()
	v := arg(args, 0)
	var data []byte
	var buf js.Value
	position := -1
	if v.Valid() && v.IsString() {
		enc := "utf8"
		if e := arg(args, 2); e.Valid() && e.IsString() {
			var err error
			enc, err = normalizeEncoding(e.Str())
			if err != nil {
				return js.Value{}, err
			}
		}
		var err error
		data, err = encodeString(v.Str(), enc)
		if err != nil {
			return js.Value{}, err
		}
		// Node returns the Buffer it wrote from, not a bare Uint8Array, so callers
		// that treat result.buffer as a Buffer keep working.
		buf = m.bytesValue(data)
		if p := arg(args, 1); p.Valid() && p.IsNumber() {
			position = p.Int()
		}
	} else {
		b, ok := v.Bytes()
		if !ok {
			return js.Value{}, js.Throw("TypeError", "The \"buffer\" argument must be an instance of Buffer or Uint8Array")
		}
		offset := intArg(args, 1, 0)
		length := intArg(args, 2, len(b)-offset)
		if offset < 0 || offset > len(b) || length < 0 || offset+length > len(b) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "The value of \"offset\" is out of range.")
		}
		data = b[offset : offset+length]
		buf = v
		if p := arg(args, 3); p.Valid() && p.IsNumber() {
			position = p.Int()
		}
	}
	var n int
	var err error
	if position >= 0 {
		n, err = syscall.Pwrite(fd, data, int64(position))
	} else {
		n, err = syscall.Write(fd, data)
	}
	if err != nil {
		return js.Value{}, m.nodeErr(err, "write", "")
	}
	out := m.e.JS.Obj()
	out.SetNum("bytesWritten", float64(n))
	out.Set("buffer", buf)
	return out.V(), nil
}

func (m *module) handleUtimes(this js.Object, args []js.Value) (js.Value, error) {
	toMs := func(v js.Value) int64 {
		if v.IsDate() {
			return int64(v.Float())
		}
		return int64(v.Float() * 1000)
	}
	atime := time.Unix(0, toMs(arg(args, 0))*int64(time.Millisecond))
	mtime := time.Unix(0, toMs(arg(args, 1))*int64(time.Millisecond))
	tv := []syscall.Timeval{
		syscall.NsecToTimeval(atime.UnixNano()),
		syscall.NsecToTimeval(mtime.UnixNano()),
	}
	if err := syscall.Futimes(this.Get("fd").Int(), tv); err != nil {
		return js.Value{}, m.nodeErr(err, "futimes", "")
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) futimesCore(args []js.Value) (js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return js.Value{}, err
	}
	toTime := func(v js.Value) time.Time {
		if v.IsDate() {
			return time.Unix(0, int64(v.Float())*int64(time.Millisecond))
		}
		return time.Unix(0, int64(v.Float()*1000)*int64(time.Millisecond))
	}
	tv := []syscall.Timeval{
		syscall.NsecToTimeval(toTime(arg(args, 1)).UnixNano()),
		syscall.NsecToTimeval(toTime(arg(args, 2)).UnixNano()),
	}
	if err := syscall.Futimes(fd, tv); err != nil {
		return js.Value{}, m.nodeErr(err, "futime", "")
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) utimesCore(args []js.Value) (js.Value, error) {
	path, err := m.pathArg(args, 0, "path")
	if err != nil {
		return js.Value{}, err
	}
	toTime := func(v js.Value) time.Time {
		if v.IsDate() {
			return time.Unix(0, int64(v.Float())*int64(time.Millisecond))
		}
		return time.Unix(0, int64(v.Float()*1000)*int64(time.Millisecond))
	}
	atime := toTime(arg(args, 1))
	mtime := toTime(arg(args, 2))
	if err := os.Chtimes(path, atime, mtime); err != nil {
		return js.Value{}, m.nodeErr(err, "utime", path)
	}
	return m.e.JS.Undefined(), nil
}

// ---- descriptor-centric APIs with Node's overloads ------------------------

func (m *module) fdArg(args []js.Value, i int) (int, error) {
	v := arg(args, i)
	if !v.Valid() || !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "fd" argument must be of type number. Received `+describe(v))
	}
	f := v.Float()
	if math.IsNaN(f) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "fd" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
	}
	if f < 0 || f > 2147483647 {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "fd" is out of range. It must be >= 0 && <= 2147483647. Received `+strconv.FormatFloat(f, 'f', -1, 64))
	}
	return int(f), nil
}

// lenArg validates truncate/ftruncate's optional length. Only undefined means
// "default to zero"; null and other non-numbers are rejected like Node.
func (m *module) lenArg(args []js.Value, i int) (int, error) {
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() {
		return 0, nil
	}
	if !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "len" argument must be of type number. Received `+describe(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "len" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
	}
	// libuv clamps a negative ftruncate length to zero (test-fs-truncate
	// relies on ftruncate(fd, -1) emptying the file).
	if f < 0 {
		f = 0
	}
	return int(f), nil
}

// numField reads an optional integer field, distinguishing "absent" from 0 so
// callers can apply Node's defaults.
func numField(v js.Value, name string) (int, bool, error) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return 0, false, nil
	}
	if !v.IsNumber() || math.IsNaN(v.Float()) || v.Float() != math.Trunc(v.Float()) {
		return 0, false, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "`+name+`" argument must be of type number. Received `+describe(v))
	}
	return v.Int(), true, nil
}

func rangeErr(name string, value int, min, max int) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		`The value of "`+name+`" is out of range. It must be >= `+strconv.Itoa(min)+
			` && <= `+strconv.Itoa(max)+`. Received `+strconv.Itoa(value))
}

// rangeMinErr is Node's validateInteger message shape (a single lower bound).
func rangeMinErr(name string, value, min int64) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		`The value of "`+name+`" is out of range. It must be >= `+strconv.FormatInt(min, 10)+
			`. Received `+strconv.FormatInt(value, 10))
}

// inspectValue renders a value the way Node's ERR_INVALID_ARG_VALUE and
// ERR_INVALID_ARG_TYPE messages do: through util.inspect when it is available.
func (m *module) inspectValue(v js.Value) string {
	if env.Registered("util") {
		if util := m.e.Exports("util"); util.Valid() {
			if insp := util.Get("inspect"); insp.Valid() && insp.IsFunction() {
				if out, err := insp.Object().Call(js.Object{}, v); err == nil && out.IsString() {
					return out.Str()
				}
			}
		}
	}
	return describe(v)
}

// invalidArgType builds ERR_INVALID_ARG_TYPE with Node's message shape.
func (m *module) invalidArgType(name, expected string, actual js.Value) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "`+name+`" argument must be `+expected+`. Received `+m.inspectValue(actual))
}

// invalidArgValue builds ERR_INVALID_ARG_VALUE with Node's message shape:
// "The argument '<name>' <reason>. Received <value>".
func (m *module) invalidArgValue(name string, value js.Value, reason string) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
		`The argument '`+name+`' `+reason+`. Received `+m.inspectValue(value))
}

// readCore implements fs.read and FileHandle#read, including the options
// object form that Node 24 accepts.
func (m *module) readCore(args []js.Value) (int, js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return 0, js.Value{}, err
	}
	var buf js.Value
	offset, length, position := 0, -1, -1
	v := arg(args, 1)
	if _, ok := v.Bytes(); ok {
		buf = v
		if o := optObject(args, 2); o.Valid() {
			offset, length, position, err = m.bufferOptions(o, len(mustBytes(buf)))
		} else {
			if p := arg(args, 2); p.Valid() && !p.IsUndefined() && !p.IsNull() && !p.IsNumber() && !p.IsFunction() {
				return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "offset" argument must be of type number. Received `+describe(p))
			}
			// Node validates offset before length, so NaN must be reported as an
			// offset problem even when the length also overflows.
			if raw := arg(args, 2); raw.Valid() && raw.IsNumber() {
				f := raw.Float()
				if math.IsNaN(f) || f != math.Trunc(f) {
					return 0, js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
						`The value of "offset" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
				}
			}
			offset = intArg(args, 2, 0)
			// A null/undefined length (or the callback sitting in that slot)
			// means "the rest of the buffer", not zero bytes.
			if lv := arg(args, 3); !lv.Valid() || lv.IsUndefined() || lv.IsNull() || lv.IsFunction() {
				length = -1
			} else {
				length = lv.Int()
			}
			position = m.positionArg(arg(args, 4))
		}
	} else if v.Valid() && v.IsObject() && !v.IsFunction() {
		o := v.Object()
		buf = o.Get("buffer")
		if buf.IsNull() {
			// An explicitly null buffer is invalid; only an omitted one means
			// "allocate a fresh 16 KiB buffer". Node describes null as an
			// instance of Object in this message.
			return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received an instance of Object`)
		}
		if !buf.Valid() || buf.IsUndefined() {
			// Node allocates a fresh 16 KiB buffer when options omit one.
			buf = m.allocBuffer(16 * 1024)
			offset, length, position = 0, -1, -1
			if o.Length() >= 0 {
				offset, length, position, err = m.bufferOptions(o, len(mustBytes(buf)))
			}
		} else if _, ok := buf.Bytes(); !ok {
			return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received `+describe(buf))
		} else {
			offset, length, position, err = m.bufferOptions(o, len(mustBytes(buf)))
		}
	} else if !v.Valid() || v.IsUndefined() || v.IsNull() || v.IsFunction() {
		// read(fd, callback) puts the callback in the buffer slot; Node then
		// allocates a 16 KiB buffer and reports it through the callback.
		buf = m.allocBuffer(16 * 1024)
		offset, length, position = 0, -1, -1
	} else {
		return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "buffer" argument must be an instance of Buffer, TypedArray, or DataView. Received `+describe(v))
	}
	if err != nil {
		return 0, js.Value{}, err
	}
	total := len(mustBytes(buf))
	// Node validates in this order (lib/fs.js read): position, then the
	// zero-length short circuit, then the empty-buffer rejection, then the
	// offset/length pair.
	if position < -1 {
		return 0, js.Value{}, rangeMinErr("position", int64(position), -1)
	}
	if length == 0 {
		return 0, buf, nil
	}
	if total == 0 {
		return 0, js.Value{}, m.invalidArgValue("buffer", buf, "is empty and cannot be written")
	}
	if offset < 0 || offset > total {
		return 0, js.Value{}, rangeErr("offset", offset, 0, total)
	}
	if length < 0 {
		length = total - offset
	}
	if length < 0 || offset+length > total {
		return 0, js.Value{}, rangeErr("length", length, 0, total-offset)
	}
	tmp := make([]byte, length)
	var n int
	if position >= 0 {
		n, err = syscall.Pread(fd, tmp, int64(position))
	} else {
		n, err = syscall.Read(fd, tmp)
	}
	if err != nil {
		return 0, js.Value{}, m.nodeErr(err, "read", "")
	}
	if n > 0 {
		if err := m.fillBuffer(buf, tmp[:n], offset); err != nil {
			return 0, js.Value{}, err
		}
	}
	return n, buf, nil
}

func (m *module) bufferOptions(o js.Object, size int) (int, int, int, error) {
	offset, _, err := numField(o.Get("offset"), "offset")
	if err != nil {
		return 0, 0, 0, err
	}
	length, hasLen, err := numField(o.Get("length"), "length")
	if err != nil {
		return 0, 0, 0, err
	}
	if !hasLen {
		length = -1
	}
	position := m.positionArg(o.Get("position"))
	if offset < 0 || offset > size {
		return 0, 0, 0, rangeErr("offset", offset, 0, size)
	}
	if hasLen && (length < 0 || length > size-offset) {
		return 0, 0, 0, rangeErr("length", length, 0, size-offset)
	}
	return offset, length, position, nil
}

func (m *module) positionArg(v js.Value) int {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return -1
	}
	return v.Int()
}

func (m *module) writeCore(args []js.Value) (int, js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return 0, js.Value{}, err
	}
	v := arg(args, 1)
	var data []byte
	var buf js.Value
	position := -1
	if v.Valid() && v.IsString() {
		enc := "utf8"
		if o := optObject(args, 2); o.Valid() {
			position = m.positionArg(o.Get("position"))
			if e := o.Get("encoding"); e.IsString() {
				if enc, err = normalizeEncoding(e.Str()); err != nil {
					return 0, js.Value{}, err
				}
			}
		} else {
			if p := arg(args, 2); p.Valid() && p.IsNumber() {
				position = p.Int()
			}
			if e := arg(args, 3); e.Valid() && e.IsString() {
				if enc, err = normalizeEncoding(e.Str()); err != nil {
					return 0, js.Value{}, err
				}
			}
		}
		if data, err = encodeString(v.Str(), enc); err != nil {
			return 0, js.Value{}, err
		}
		buf = v
	} else {
		b, ok := v.Bytes()
		if !ok {
			return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "buffer" argument must be of type string or an instance of Buffer, TypedArray, or DataView. Received `+describe(v))
		}
		offset, length := 0, -1
		if o := optObject(args, 2); o.Valid() {
			offset, length, position, err = m.bufferOptions(o, len(b))
		} else {
			if p := arg(args, 2); p.Valid() && !p.IsUndefined() && !p.IsNull() && !p.IsNumber() && !p.IsFunction() {
				return 0, js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "offset" argument must be of type number. Received `+describe(p))
			}
			// Node validates offset before length, so NaN must be reported as an
			// offset problem even when the length also overflows.
			if raw := arg(args, 2); raw.Valid() && raw.IsNumber() {
				f := raw.Float()
				if math.IsNaN(f) || f != math.Trunc(f) {
					return 0, js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
						`The value of "offset" is out of range. It must be an integer. Received `+strconv.FormatFloat(f, 'f', -1, 64))
				}
			}
			offset = intArg(args, 2, 0)
			// A null/undefined length (or the callback sitting in that slot)
			// means "the rest of the buffer", not zero bytes.
			if lv := arg(args, 3); !lv.Valid() || lv.IsUndefined() || lv.IsNull() || lv.IsFunction() {
				length = -1
			} else {
				length = lv.Int()
				// An explicit negative length is out of range; only the
				// omitted sentinel means "the rest of the buffer".
				if length < 0 {
					return 0, js.Value{}, rangeErr("length", length, 0, len(b)-offset)
				}
			}
			position = m.positionArg(arg(args, 4))
		}
		if err != nil {
			return 0, js.Value{}, err
		}
		if offset < 0 || offset > len(b) {
			return 0, js.Value{}, rangeErr("offset", offset, 0, len(b))
		}
		if length < 0 {
			length = len(b) - offset
		}
		if offset+length > len(b) {
			return 0, js.Value{}, rangeErr("length", length, 0, len(b)-offset)
		}
		data = b[offset : offset+length]
		buf = v
	}
	var n int
	if position >= 0 {
		n, err = syscall.Pwrite(fd, data, int64(position))
	} else {
		n, err = syscall.Write(fd, data)
	}
	if err != nil {
		return 0, js.Value{}, m.nodeErr(err, "write", "")
	}
	return n, buf, nil
}

func (m *module) allocBuffer(n int) js.Value {
	if b := m.e.Global().Get("Buffer"); b.Valid() && b.IsObject() {
		if alloc := b.Object().Get("alloc"); alloc.IsFunction() {
			if v, err := alloc.Object().Call(b.Object(), m.e.JS.Num(float64(n))); err == nil {
				return v
			}
		}
	}
	return m.e.JS.Bytes(make([]byte, n)).V()
}

func (m *module) buffersArg(args []js.Value, i int) ([]js.Value, error) {
	v := arg(args, i)
	if !v.Valid() || !v.IsArray() {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "buffers" argument must be an instance of Array. Received `+describe(v))
	}
	n := v.Object().Length()
	out := make([]js.Value, n)
	for j := 0; j < n; j++ {
		b := v.Object().At(j)
		if _, ok := b.Bytes(); !ok {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "buffers[`+strconv.Itoa(j)+`]" argument must be an instance of Buffer or Uint8Array. Received `+describe(b))
		}
		out[j] = b
	}
	return out, nil
}

func (m *module) readvCore(args []js.Value) (int, []js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return 0, nil, err
	}
	bufs, err := m.buffersArg(args, 1)
	if err != nil {
		return 0, nil, err
	}
	position := -1
	if p := arg(args, 2); p.Valid() && p.IsNumber() {
		position = p.Int()
	}
	total := 0
	for _, b := range bufs {
		total += len(mustBytes(b))
	}
	tmp := make([]byte, total)
	var n int
	if position >= 0 {
		n, err = syscall.Pread(fd, tmp, int64(position))
	} else {
		n, err = syscall.Read(fd, tmp)
	}
	if err != nil {
		return 0, nil, m.nodeErr(err, "read", "")
	}
	left := n
	for _, b := range bufs {
		if left == 0 {
			break
		}
		sz := len(mustBytes(b))
		if sz > left {
			sz = left
		}
		if err := m.fillBuffer(b, tmp[:sz], 0); err != nil {
			return 0, nil, err
		}
		tmp = tmp[sz:]
		left -= sz
	}
	return n, bufs, nil
}

func (m *module) writevCore(args []js.Value) (int, []js.Value, error) {
	fd, err := m.fdArg(args, 0)
	if err != nil {
		return 0, nil, err
	}
	bufs, err := m.buffersArg(args, 1)
	if err != nil {
		return 0, nil, err
	}
	position := -1
	if p := arg(args, 2); p.Valid() && p.IsNumber() {
		position = p.Int()
	}
	var data []byte
	for _, b := range bufs {
		data = append(data, mustBytes(b)...)
	}
	var n int
	if position >= 0 {
		n, err = syscall.Pwrite(fd, data, int64(position))
	} else {
		n, err = syscall.Write(fd, data)
	}
	if err != nil {
		return 0, nil, m.nodeErr(err, "write", "")
	}
	return n, bufs, nil
}

func (m *module) orUndefined(v js.Value) js.Value {
	if v.Valid() {
		return v
	}
	return m.e.JS.Undefined()
}

func callbackFunc(args []js.Value) (js.Value, error) {
	for i := len(args) - 1; i >= 0; i-- {
		if args[i].IsFunction() {
			return args[i], nil
		}
	}
	return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "cb" argument must be of type function. Received `+describe(arg(args, len(args)-1)))
}

// installAdvanced adds the descriptor, chown, link and statfs families and the
// overloads that Node 24 supports for read/write/readv/writev.
// installStreams builds fs.ReadStream, fs.WriteStream, fs.watch, fs.watchFile,
// fs.unwatchFile and fs.Dir on top of the JavaScript stream classes. The state
// machine lives in JavaScript so streams share Node's event ordering, while
// the actual I/O goes through this module's synchronous host calls.
func (m *module) installStreams(exports js.Object) {
	stream := m.e.Exports("stream")
	if !stream.Valid() || !stream.IsObject() {
		return
	}
	helpers := m.e.JS.Obj()
	helpers.SetFunc("tick", func(this js.Object, args []js.Value) (js.Value, error) {
		if fn := arg(args, 0); fn.Valid() && fn.IsFunction() {
			m.e.NextTick(fn, args[1:]...)
		}
		return m.e.JS.Undefined(), nil
	})
	helpers.SetFunc("setInterval", func(this js.Object, args []js.Value) (js.Value, error) {
		fn := arg(args, 0)
		delay := 100.0
		if v := arg(args, 1); v.Valid() && v.IsNumber() {
			delay = v.Float()
		}
		if !fn.Valid() || !fn.IsFunction() {
			return m.e.JS.Num(0), nil
		}
		return m.e.JS.Num(float64(m.e.SetTimeout(fn, delay, true, nil))), nil
	})
	helpers.SetFunc("clearInterval", func(this js.Object, args []js.Value) (js.Value, error) {
		if v := arg(args, 0); v.Valid() && v.IsNumber() {
			m.e.ClearTimer(v.Int())
		}
		return m.e.JS.Undefined(), nil
	})
	fn, err := m.e.JS.Eval(streamPrelude, "[fs:streams]")
	if err != nil {
		return
	}
	emitter := m.e.JS.Undefined()
	if env.Registered("events") {
		if ev := m.e.Exports("events").Get("EventEmitter"); ev.Valid() {
			emitter = ev
		}
	}
	if _, cerr := fn.Object().Call(js.Object{}, exports.V(), stream.V(), helpers.V(), emitter); cerr != nil {
		panic("godejs fs streams: " + cerr.Error())
	}
}

func (m *module) installAdvanced(exports js.Object) {
	exports.SetFunc("read", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node validates the descriptor before the callback argument.
		if _, ferr := m.fdArg(args, 0); ferr != nil {
			return js.Value{}, ferr
		}
		cb, err := callbackFunc(args)
		if err != nil {
			return js.Value{}, err
		}
		n, buf, err := m.readCore(args)
		if validationErr(err) {
			return js.Value{}, err
		}
		m.tick(cb, m.errorValue(err), m.e.JS.Num(float64(n)), m.orUndefined(buf))
		return m.e.JS.Undefined(), nil
	})
	exports.SetFunc("readSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		n, _, err := m.readCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(n)), nil
	}))
	exports.SetFunc("write", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node validates the descriptor before the callback argument.
		if _, ferr := m.fdArg(args, 0); ferr != nil {
			return js.Value{}, ferr
		}
		cb, err := callbackFunc(args)
		if err != nil {
			return js.Value{}, err
		}
		n, buf, err := m.writeCore(args)
		if validationErr(err) {
			return js.Value{}, err
		}
		m.tick(cb, m.errorValue(err), m.e.JS.Num(float64(n)), m.orUndefined(buf))
		return m.e.JS.Undefined(), nil
	})
	exports.SetFunc("writeSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		n, _, err := m.writeCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(n)), nil
	}))
	exports.SetFunc("readv", func(this js.Object, args []js.Value) (js.Value, error) {
		cb, err := callbackFunc(args)
		if err != nil {
			return js.Value{}, err
		}
		n, bufs, err := m.readvCore(args)
		if validationErr(err) {
			return js.Value{}, err
		}
		m.tick(cb, m.errorValue(err), m.e.JS.Num(float64(n)), m.e.JS.Arr(bufs).V())
		return m.e.JS.Undefined(), nil
	})
	exports.SetFunc("writev", func(this js.Object, args []js.Value) (js.Value, error) {
		cb, err := callbackFunc(args)
		if err != nil {
			return js.Value{}, err
		}
		n, bufs, err := m.writevCore(args)
		if validationErr(err) {
			return js.Value{}, err
		}
		m.tick(cb, m.errorValue(err), m.e.JS.Num(float64(n)), m.e.JS.Arr(bufs).V())
		return m.e.JS.Undefined(), nil
	})
	exports.SetFunc("readvSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		n, _, err := m.readvCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(n)), nil
	}))
	exports.SetFunc("writevSync", m.syncCall(func(args []js.Value) (js.Value, error) {
		n, _, err := m.writevCore(args)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(n)), nil
	}))

	fdCall := func(name string, fn func([]js.Value) (js.Value, error)) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			cb, err := callbackFunc(args)
			if err != nil {
				return js.Value{}, err
			}
			v, err := fn(args)
			if validationErr(err) {
				return js.Value{}, err
			}
			var val js.Value
			if err == nil && v.Valid() {
				val = v
			} else {
				val = m.e.JS.Undefined()
			}
			m.tick(cb, m.errorValue(err), val)
			return m.e.JS.Undefined(), nil
		})
	}
	simple := func(name string, fn func([]js.Value) (js.Value, error)) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			cb, err := callbackFunc(args)
			if err != nil {
				return js.Value{}, err
			}
			_, ferr := fn(args)
			if validationErr(ferr) {
				return js.Value{}, ferr
			}
			m.tick(cb, m.errorValue(ferr))
			return m.e.JS.Undefined(), nil
		})
	}

	fdCall("fstat", func(args []js.Value) (js.Value, error) {
		fd, err := m.fdArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.fstatWith(fd, statOption(args, "bigint", false))
	})
	simple("fsync", func(args []js.Value) (js.Value, error) { return m.syncFd(args, true) })
	simple("fdatasync", func(args []js.Value) (js.Value, error) { return m.syncFd(args, false) })
	simple("ftruncate", func(args []js.Value) (js.Value, error) {
		fd, err := m.fdArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		length, err := m.lenArg(args, 1)
		if err != nil {
			return js.Value{}, err
		}
		if err := syscall.Ftruncate(fd, int64(length)); err != nil {
			return js.Value{}, m.nodeErr(err, "ftruncate", "")
		}
		return m.e.JS.Undefined(), nil
	})
	simple("fchmod", m.fchmodCore)
	simple("fchown", m.fchownCore)
	simple("chown", func(args []js.Value) (js.Value, error) { return m.chownCore(args, true) })
	simple("lchown", func(args []js.Value) (js.Value, error) { return m.chownCore(args, false) })
	simple("link", m.linkCore)
	simple("futimes", m.futimesCore)
	fdCall("statfs", m.statfsCore)

	exports.SetFunc("fchmodSync", m.syncCall(m.fchmodCore))
	exports.SetFunc("fchownSync", m.syncCall(m.fchownCore))
	exports.SetFunc("chownSync", m.syncCall(func(args []js.Value) (js.Value, error) { return m.chownCore(args, true) }))
	exports.SetFunc("lchownSync", m.syncCall(func(args []js.Value) (js.Value, error) { return m.chownCore(args, false) }))
	exports.SetFunc("linkSync", m.syncCall(m.linkCore))
	exports.SetFunc("futimesSync", m.syncCall(m.futimesCore))
	exports.SetFunc("statfsSync", m.syncCall(m.statfsCore))
	exports.SetFunc("exists", func(this js.Object, args []js.Value) (js.Value, error) {
		cb := arg(args, 1)
		if !cb.Valid() || !cb.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "cb" argument must be of type function. Received `+describe(cb))
		}
		path, err := m.pathArg(args, 0, "path")
		ok := err == nil
		if ok {
			_, statErr := os.Stat(path)
			ok = statErr == nil
		}
		m.tick(cb, m.e.JS.Bool(ok))
		return m.e.JS.Undefined(), nil
	})

	// fs/promises additions.
	if p := exports.Get("promises"); p.Valid() && p.IsObject() {
		po := p.Object()
		po.SetFunc("chown", m.promise(func(args []js.Value) (js.Value, error) { return m.chownCore(args, true) }))
		po.SetFunc("lchown", m.promise(func(args []js.Value) (js.Value, error) { return m.chownCore(args, false) }))
		po.SetFunc("lchmod", m.promise(func(args []js.Value) (js.Value, error) { return m.chmodCore(args) }))
		po.SetFunc("link", m.promise(m.linkCore))
		po.SetFunc("statfs", m.promise(m.statfsCore))
	}
}

// streamPrelude defines the Node stream subclasses and watchers in JavaScript.
// It is evaluated once per environment with (fs, stream, helpers).
const streamPrelude = `(function (fs, stream, helpers, EventEmitter) {
  'use strict';
  if (!stream || typeof stream.Readable !== 'function') return;
  var Readable = stream.Readable;
  var Writable = stream.Writable;

  function basename(p) {
    p = String(p);
    var i = p.lastIndexOf('/');
    return i === -1 ? p : p.slice(i + 1);
  }
  function dirname(p) {
    p = String(p);
    var i = p.lastIndexOf('/');
    return i <= 0 ? '/' : p.slice(0, i);
  }
  function statOrNull(p) {
    try { return fs.statSync(p); } catch (e) { return null; }
  }

  // --------------------------------------------------------- ReadStream --
  // Node accepts an options object, a bare encoding string, or undefined for
  // its fs streams; anything else is a TypeError before any I/O happens.
  function invalidEncodingValue(encoding) {
    var err = new TypeError("The argument 'encoding' is invalid encoding. Received '" +
      encoding + "'");
    err.code = 'ERR_INVALID_ARG_VALUE';
    return err;
  }
  function validateEncodingOption(options) {
    var enc = options.encoding;
    if (enc === undefined || enc === null) return;
    if (typeof enc !== 'string') throw streamArgType('options.encoding', 'string', enc);
    if (enc !== 'buffer' && !Buffer.isEncoding(enc)) throw invalidEncodingValue(enc);
  }
  function normaliseStreamOptions(options) {
    if (options === undefined || options === null) return {};
    if (typeof options === 'string') options = { encoding: options };
    else if (typeof options !== 'object') throw streamArgType('options', 'object', options);
    validateEncodingOption(options);
    return options;
  }
  function streamArgType(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be of type ' + expected +
      '. Received ' + (value === null ? 'null' : typeof value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }
  function validateStreamRange(value, name, allowInfinity) {
    if (value === undefined) return;
    if (typeof value !== 'number') throw streamArgType('options.' + name, 'number', value);
    if (allowInfinity && value === Infinity) return;
    if (!Number.isSafeInteger(value) || value < 0) {
      var err = new RangeError('The value of "options.' + name + '" is out of range. ' +
        'It must be >= 0 && <= Number.MAX_SAFE_INTEGER. Received ' + value);
      err.code = 'ERR_OUT_OF_RANGE';
      throw err;
    }
  }

  // Node validates the injectable fs implementation with "property" wording
  // ('The "options.fs.open" property must be of type function').
  function streamArgTypeProperty(name, value) {
    var received = value === null ? 'null'
      : typeof value === 'string' ? "type string ('" + value + "')"
      : typeof value;
    var err = new TypeError('The "' + name + '" property must be of type function. Received ' + received);
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function validateStreamPathAndFd(path, options) {
    if (options.fd !== undefined && options.fd !== null) {
      if (typeof options.fd !== 'number') throw streamArgType('options.fd', 'number', options.fd);
      return;
    }
    var urlCtor = typeof URL === 'function' ? URL : null;
    var isURL = urlCtor !== null && path instanceof urlCtor;
    var isBuffer = typeof Buffer === 'function' && Buffer.isBuffer(path);
    if (typeof path !== 'string' && !isURL && !isBuffer) {
      throw streamArgType('path', 'string or an instance of Buffer or URL', path);
    }
  }

  function ReadStream(path, options) {
    if (!(this instanceof ReadStream)) return new ReadStream(path, options);
    options = normaliseStreamOptions(options);
    validateStreamPathAndFd(path, options);
    validateStreamRange(options.start, 'start', false);
    validateStreamRange(options.end, 'end', true);
    var impl = options.fs || fs;
    if (options.fd === undefined || options.fd === null) {
      if (typeof impl.open !== 'function') throw streamArgTypeProperty('options.fs.open', impl.open);
    }
    if (typeof impl.read !== 'function') throw streamArgTypeProperty('options.fs.read', impl.read);
    if (options.autoClose !== false && typeof impl.close !== 'function') {
      throw streamArgTypeProperty('options.fs.close', impl.close);
    }
    if (options.start !== undefined && options.start !== null &&
        options.end !== undefined && options.end !== null &&
        options.start > options.end) {
      var rangeErr = new RangeError('The value of "start" is out of range. It must be <= "end" ' +
        '(here: ' + options.end + '). Received ' + options.start);
      rangeErr.code = 'ERR_OUT_OF_RANGE';
      throw rangeErr;
    }
    // Node ties autoDestroy to autoClose, so a stream opened with
    // autoClose: false stays usable (not destroyed) after 'end'. The options
    // may be a frozen object, so derive into a copy instead of mutating it.
    var streamOptions = {
      flags: options.flags,
      mode: options.mode,
      start: options.start,
      end: options.end,
      autoClose: options.autoClose,
      autoDestroy: options.autoDestroy,
      emitClose: options.emitClose,
      encoding: options.encoding,
      highWaterMark: options.highWaterMark,
      fd: options.fd,
      signal: options.signal
    };
    if (streamOptions.autoDestroy === undefined) {
      streamOptions.autoDestroy = streamOptions.autoClose === undefined
        ? true
        : streamOptions.autoClose;
    }
    Readable.call(this, streamOptions);
    // Node's fs streams decode for you when an encoding is requested, so the
    // 'data' events carry strings rather than Buffers.
    if (typeof options.encoding === 'string' && options.encoding !== 'buffer') {
      this.setEncoding(options.encoding);
    }
    // Node leaves 'path' undefined when the stream was opened from a fd.
    this.path = options.fd === undefined || options.fd === null
      ? (path === null || path === undefined ? undefined : path)
      : undefined;
    this.flags = options.flags || 'r';
    this.mode = options.mode === undefined ? 438 : options.mode;
    this.start = options.start;
    this.end = options.end;
    this.autoClose = options.autoClose !== false;
    this.emitClose = options.emitClose !== false;
    // Node's fs streams default to a 64 KiB read size, larger than the generic
    // Readable default; keep our own copy so the pump never depends on which
    // field the stream base class chose to record it in.
    this._hwm = options.highWaterMark === undefined
      ? 65536
      : Number(options.highWaterMark);
    this.bytesRead = 0;
    this.fd = options.fd === undefined || options.fd === null ? null : options.fd;
    this.pending = this.fd === null;
    var fromPath = this.fd === null;
    this.pos = this.start === undefined || this.start === null ? (fromPath ? 0 : null) : this.start;
    this._ready = !fromPath;
    if (this.fd === null) {
      try {
        this.fd = fs.openSync(path, this.flags, this.mode);
      } catch (openErr) {
        this._openErr = openErr;
      }
    }
    var self = this;
    helpers.tick(function () {
      self.pending = false;
      if (self._openErr) {
        var err = self._openErr;
        self._openErr = null;
        self.destroyed = true;
        self._destroyed = true;
        self.emit('error', err);
        if (self.emitClose) self.emit('close');
        return;
      }
      if (fromPath) self.emit('open', self.fd);
      self.emit('ready');
      self._ready = true;
      if (self._pendingPump) { self._pendingPump = false; self._pumpRead(); }
    });
  }
  ReadStream.prototype = Object.create(Readable.prototype);
  ReadStream.prototype.constructor = ReadStream;
  Object.defineProperty(ReadStream.prototype, 'readable', { value: true, writable: true });

  // The Readable base class drives reads through _read(); this stream fills the
  // buffer itself from the file descriptor.
  ReadStream.prototype._read = function () { this._pumpRead(); };

  ReadStream.prototype._pumpRead = function () {
    if (this._destroyed || this._readDone || this._filling) return;
    if (!this._ready) { this._pendingPump = true; return; }
    if (this.fd === null) { this._pendingPump = true; return; }
    this._filling = true;
    var hwm = this._hwm > 0 ? this._hwm : 65536;
    try {
      var remaining = hwm;
      if (this.end !== undefined && this.end !== null) {
        var at = this.pos === null ? 0 : this.pos;
        remaining = this.end - at + 1;
      }
      if (remaining <= 0) { this._finishRead(); return; }
      var len = Math.min(hwm, remaining);
      var buf = Buffer.allocUnsafe(len);
      var bytesRead;
      try {
        bytesRead = fs.readSync(this.fd, buf, 0, len, this.pos);
      } catch (readErr) {
        this._filling = false;
        // With autoDestroy off (autoClose: false) Node reports the read error
        // but leaves the stream, and its descriptor, intact.
        if (this._readableState && this._readableState.autoDestroy === false) {
          this.emit('error', readErr);
          return;
        }
        this.destroy(readErr);
        return;
      }
      if (bytesRead === 0) { this._finishRead(); return; }
      this.bytesRead += bytesRead;
      if (this.pos !== null) this.pos += bytesRead;
      this.push(buf.slice(0, bytesRead));
      var self = this;
      var more = bytesRead >= len &&
        (this.end === undefined || this.end === null || this.pos === null || this.pos <= this.end);
      if (more) helpers.tick(function () { self._pumpRead(); });
      else this._finishRead();
    } finally {
      this._filling = false;
    }
  };
  ReadStream.prototype._finishRead = function () {
    if (this._readDone) return;
    this._readDone = true;
    if (this.autoClose && this.fd !== null) {
      try { fs.closeSync(this.fd); } catch (e) {}
      this.fd = null;
    }
    this.push(null);
    // A stream that closed its own descriptor emits 'close' after 'end'.
    if (this.emitClose !== false && !this._closeEmitted) {
      this._closeEmitted = true;
      var self = this;
      helpers.tick(function () { self.emit('close'); });
    }
  };
  ReadStream.prototype.on = function (name, fn) {
    var out = Readable.prototype.on.call(this, name, fn);
    if (name === 'data' || name === 'readable' || name === 'end') {
      var self = this;
      helpers.tick(function () { self._pumpRead(); });
    }
    return out;
  };
  ReadStream.prototype.resume = function () {
    Readable.prototype.resume.call(this);
    // Node never performs file I/O synchronously from resume(): a read error
    // must be reported through 'error' after listeners have been attached.
    var self = this;
    helpers.tick(function () { self._pumpRead(); });
    return this;
  };
  ReadStream.prototype.read = function (size) {
    var buffered = this._buffer ? this._buffer.length : 0;
    if (!buffered && !this._readDone) this._pumpRead();
    return Readable.prototype.read.call(this, size);
  };
  ReadStream.prototype.destroy = function (err) {
    if (this._destroyed) return this;
    this._readDone = true;
    this._closeEmitted = true;
    if (this.fd !== null && this.autoClose) {
      try { fs.closeSync(this.fd); } catch (e) {}
    }
    this.fd = null;
    return Readable.prototype.destroy.call(this, err);
  };
  // Node's ReadStream has a public open() that (re)opens the path through
  // fs.open and re-arms the pump once the descriptor arrives. It is what
  // test-fs-read-stream-patch-open patches and calls.
  ReadStream.prototype.open = function () {
    if (this.fd !== null) return;
    var self = this;
    this.pending = true;
    this._openErr = null;
    fs.open(this.path, this.flags, this.mode, function (err, fd) {
      self.pending = false;
      if (err) {
        self._openErr = null;
        self.emit('error', err);
        if (self.emitClose) self.emit('close');
        return;
      }
      self.fd = fd;
      self._ready = true;
      self.emit('open', fd);
      self.emit('ready');
      if (self._pendingPump) { self._pendingPump = false; self._pumpRead(); }
    });
  };
  ReadStream.prototype.close = function (cb) {
    this.destroy();
    if (typeof cb === 'function') helpers.tick(cb);
  };

  // -------------------------------------------------------- WriteStream --
  function WriteStream(path, options) {
    if (!(this instanceof WriteStream)) return new WriteStream(path, options);
    options = normaliseStreamOptions(options);
    validateStreamPathAndFd(path, options);
    validateStreamRange(options.start, 'start', false);
    var impl = options.fs || fs;
    if (options.fd === undefined || options.fd === null) {
      if (typeof impl.open !== 'function') throw streamArgTypeProperty('options.fs.open', impl.open);
    }
    if (typeof impl.write !== 'function') throw streamArgTypeProperty('options.fs.write', impl.write);
    if (impl.writev !== undefined && typeof impl.writev !== 'function') {
      throw streamArgTypeProperty('options.fs.writev', impl.writev);
    }
    if (options.autoClose !== false && typeof impl.close !== 'function') {
      throw streamArgTypeProperty('options.fs.close', impl.close);
    }
    var streamOptions = {
      flags: options.flags,
      mode: options.mode,
      start: options.start,
      end: options.end,
      autoClose: options.autoClose,
      autoDestroy: options.autoDestroy,
      emitClose: options.emitClose,
      encoding: options.encoding,
      highWaterMark: options.highWaterMark,
      fd: options.fd,
      signal: options.signal
    };
    if (streamOptions.autoDestroy === undefined) {
      streamOptions.autoDestroy = streamOptions.autoClose === undefined
        ? true
        : streamOptions.autoClose;
    }
    Writable.call(this, streamOptions);
    // Node leaves 'path' undefined when the stream was opened from a fd.
    this.path = options.fd === undefined || options.fd === null
      ? (path === null || path === undefined ? undefined : path)
      : undefined;
    this.flags = options.flags || 'w';
    this.mode = options.mode === undefined ? 438 : options.mode;
    this.start = options.start;
    this.pos = this.start === undefined || this.start === null ? null : this.start;
    this.autoClose = options.autoClose !== false;
    this.emitClose = options.emitClose !== false;
    this._flush = options.flush === true;
    this.bytesWritten = 0;
    this.fd = options.fd === undefined || options.fd === null ? null : options.fd;
    this.pending = this.fd === null;
    this._fromPath = this.fd === null;
    if (this.fd === null) {
      try {
        this.fd = fs.openSync(path, this.flags, this.mode);
      } catch (openErr) {
        this._openErr = openErr;
      }
    }
    this.writable = true;
    var self = this;
    helpers.tick(function () {
      self.pending = false;
      if (self._openErr) {
        var err = self._openErr;
        self._openErr = null;
        self._errored = true;
        self.emit('error', err);
        if (self.emitClose) self.emit('close');
        return;
      }
      if (self._fromPath) self.emit('open', self.fd);
      self.emit('ready');
      // Writes queued before the descriptor opened are retried by _write,
      // which re-ticks itself until this.fd is set.
    });
  }
  WriteStream.prototype = Object.create(Writable.prototype);
  WriteStream.prototype.constructor = WriteStream;
  Object.defineProperty(WriteStream.prototype, 'writable', { value: true, writable: true });

  WriteStream.prototype._write = function (chunk, encoding, cb) {
    if (this.fd === null) {
      if (this._openErr) { cb(this._openErr); return; }
      var self0 = this;
      helpers.tick(function () { self0._write(chunk, encoding, cb); });
      return;
    }
    var data = chunk;
    if (typeof data === 'string') data = Buffer.from(data, encoding || 'utf8');
    var offset = 0;
    try {
      while (offset < data.length) {
        var n = this.pos === null
          ? fs.writeSync(this.fd, data, offset, data.length - offset)
          : fs.writeSync(this.fd, data, offset, data.length - offset, this.pos);
        if (n <= 0) throw new Error('Unable to write to file');
        offset += n;
        if (this.pos !== null) this.pos += n;
        this.bytesWritten += n;
      }
    } catch (writeErr) {
      cb(writeErr);
      return;
    }
    cb();
  };
  WriteStream.prototype._close = function (force) {
    // autoClose: false means the caller owns the descriptor: finishing the
    // stream must not close it and must not emit 'close'.
    if (!this.autoClose && !force) return;
    if (this._closed) return;
    this._closed = true;
    if (this.fd !== null && this._flush) {
      try { fs.fsyncSync(this.fd); } catch (e) { this.emit('error', e); }
    }
    if (this.fd !== null) {
      try { fs.closeSync(this.fd); } catch (e) {}
    }
    this.fd = null;
    // An autoDestroying stream reaches 'close' through destroy(); emitting it
    // here as well would fire 'close' twice.
    var baseCloses = this._writableState ? this._writableState.autoDestroy !== false : false;
    if (baseCloses) return;
    if (this._writableState) this._writableState.closed = true;
    if (this.emitClose !== false) {
      var self = this;
      helpers.tick(function () { self.emit('close'); });
    }
  };
  WriteStream.prototype.end = function (chunk, encoding, cb) {
    var self = this;
    if (!this._closeHooked) {
      this._closeHooked = true;
      this.on('finish', function () { self._close(); });
    }
    return Writable.prototype.end.call(this, chunk, encoding, cb);
  };
  WriteStream.prototype.close = function (cb) {
    if (this.autoClose) {
      this.destroy();
      if (typeof cb === 'function') this.once('close', cb);
      return;
    }
    this._close(true);
    if (typeof cb === 'function') helpers.tick(cb);
  };
  WriteStream.prototype.destroy = function (err) {
    if (this.destroyed) return this;
    this._close();
    return Writable.prototype.destroy.call(this, err);
  };

  fs.ReadStream = ReadStream;
  fs.WriteStream = WriteStream;
  fs.createReadStream = function (path, options) { return new ReadStream(path, options); };
  fs.createWriteStream = function (path, options) { return new WriteStream(path, options); };

  // Node exposes _toUnixTimestamp for its own utimes validation; userland and
  // the official tests reach for it, so keep the same semantics: numeric
  // strings pass through, finite numbers are seconds (negative means now),
  // and Date objects become fractional seconds.
  fs._toUnixTimestamp = function (time) {
    if (typeof time === 'string' && +time == time) return +time;
    if (typeof time === 'number' && Number.isFinite(time)) {
      if (time < 0) return Date.now() / 1000;
      return time;
    }
    if (time instanceof Date) return time.getTime() / 1000;
    var err = new TypeError('The "time" argument must be of type Date or Time in seconds. ' +
      'Received ' + (typeof time === 'object' ? 'an instance of Object' : typeof time));
    err.code = 'ERR_INVALID_ARG_TYPE';
    throw err;
  };

  // ------------------------------------------------------------- watchers --
  function snapshot(target, recursive) {
    var map = Object.create(null);
    var st = statOrNull(target);
    if (!st) return map;
    map[target] = (st.isDirectory() ? 'd:' : 'f:') + st.mtimeMs + ':' + st.size;
    if (!st.isDirectory()) return map;
    function walk(dir) {
      var entries;
      try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch (e) { return; }
      for (var i = 0; i < entries.length; i++) {
        var d = entries[i];
        var full = dir + '/' + d.name;
        var child = statOrNull(full);
        map[full] = child ? ((child.isDirectory() ? 'd:' : 'f:') + child.mtimeMs + ':' + child.size) : 'missing';
        if (recursive && d.isDirectory()) walk(full);
      }
    }
    walk(target);
    return map;
  }
  // diffSnapshots reports every difference the way libuv would: entries that
  // appeared or vanished are 'rename', changed entries are 'change'. A whole
  // new subtree can appear between two polls, so this must not stop at the
  // first difference (a new directory would otherwise hide the file created
  // inside it, and recursive watchers would never see that file).
  function diffSnapshots(previous, next) {
    var changes = [];
    var k;
    for (k in next) if (!(k in previous)) changes.push({ type: 'rename', name: k });
    for (k in previous) if (!(k in next)) changes.push({ type: 'rename', name: k });
    for (k in previous) if ((k in next) && previous[k] !== next[k]) {
      // libuv reports directory modifications as its own 'change' events only
      // for files; a directory's mtime moves whenever a child is touched, and
      // Node does not surface that as an extra event.
      var wasDir = previous[k].charAt(0) === 'd';
      var isDir = next[k].charAt(0) === 'd';
      if (wasDir && isDir) continue;
      changes.push({ type: 'change', name: k });
    }
    return changes;
  }
  // FSWatcher is Node's public watcher class: an EventEmitter that fs.watch()
  // returns and that is also constructible on its own (arguments are ignored,
  // as in Node, where only fs.watch() starts a watcher).
  // fs always links the events module, but a build that installs fs alone must
  // still produce a working watcher, so fall back to a bare emitter.
  var Emitter = typeof EventEmitter === 'function' ? EventEmitter : null;
  // Node watchers are libuv handles, so ref()/unref() must actually control
  // whether the polling keeps the event loop alive. The timers module returns
  // Timeout objects with those methods; the Go helper only knows ids, so use
  // the module when the runtime exposes it.
  var watcherTimers = null;
  try {
    if (typeof require === 'function') watcherTimers = require('timers');
  } catch (e) { watcherTimers = null; }
  function setWatcherInterval(fn, delay) {
    if (watcherTimers && typeof watcherTimers.setInterval === 'function') {
      return watcherTimers.setInterval(fn, delay);
    }
    return helpers.setInterval(fn, delay);
  }
  function clearWatcherInterval(timer) {
    if (timer === null || timer === undefined) return;
    try {
      if (watcherTimers && typeof watcherTimers.clearInterval === 'function') watcherTimers.clearInterval(timer);
    } catch (e) {}
    try { helpers.clearInterval(timer); } catch (e) {}
  }
  function FSWatcher() {
    if (!(this instanceof FSWatcher)) return new FSWatcher();
    if (Emitter) Emitter.call(this);
    this._listeners = Object.create(null);
    this._timer = null;
    this._closed = false;
  }
  FSWatcher.prototype = Emitter ? Object.create(Emitter.prototype) : {};
  FSWatcher.prototype.constructor = FSWatcher;
  if (!Emitter) {
    FSWatcher.prototype.on = function (name, fn) {
      (this._listeners[name] || (this._listeners[name] = [])).push(fn);
      return this;
    };
    FSWatcher.prototype.addListener = FSWatcher.prototype.on;
    FSWatcher.prototype.once = function (name, fn) {
      var self = this;
      function wrapper() { self.removeListener(name, wrapper); fn.apply(self, arguments); }
      return this.on(name, wrapper);
    };
    FSWatcher.prototype.removeListener = function (name, fn) {
      var list = this._listeners[name];
      if (list) {
        var i = list.indexOf(fn);
        if (i !== -1) list.splice(i, 1);
      }
      return this;
    };
    FSWatcher.prototype.emit = function (name) {
      var list = this._listeners[name];
      if (!list || !list.length) return false;
      var args = Array.prototype.slice.call(arguments, 1);
      list.slice().forEach(function (fn) { fn.apply(this, args); }, this);
      return true;
    };
  }
  FSWatcher.prototype.close = function () {
    if (this._closed) return;
    this._closed = true;
    if (this._timer !== null) { clearWatcherInterval(this._timer); this._timer = null; }
    var self = this;
    // Node emits 'close' on the next tick, after the handle is released.
    helpers.tick(function () { self.emit('close'); });
  };
  FSWatcher.prototype.ref = function () { return this; };
  FSWatcher.prototype.unref = function () { return this; };
  if (typeof Symbol === 'function' && typeof Symbol.dispose === 'symbol') {
    FSWatcher.prototype[Symbol.dispose] = function () { this.close(); };
  }

  function argError(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be of type ' + expected +
      '. Received ' + (typeof value === 'string' ? "type string ('" + value + "')" : typeof value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function isPathLike(value) {
    if (typeof value === 'string') return true;
    if (value && typeof value === 'object') {
      if (typeof Buffer === 'function' && Buffer.isBuffer && Buffer.isBuffer(value)) return true;
      if (typeof value.href === 'string') return true;
    }
    return false;
  }

  function pathHasNullBytes(value) {
    if (typeof value === 'string') return value.indexOf('\u0000') !== -1;
    if (value === null || typeof value !== 'object') return false;
    if (typeof value.byteLength === 'number' && typeof value.length === 'number') {
      for (var i = 0; i < value.length; i++) if (value[i] === 0) return true;
      return false;
    }
    if (typeof value.pathname === 'string') {
      var p = value.pathname;
      try { p = decodeURIComponent(p); } catch (e) { /* keep encoded form */ }
      return p.indexOf('\u0000') !== -1;
    }
    return false;
  }

  function assertPathHasNoNullBytes(value, name) {
    if (!pathHasNullBytes(value)) return;
    var err = new TypeError("The argument '" + name + "' must be a string, Uint8Array, " +
      "or URL without null bytes. Received '" + String(value) + "'");
    err.code = 'ERR_INVALID_ARG_VALUE';
    throw err;
  }

  function watch(filename, options, listener) {
    assertPathHasNoNullBytes(filename, 'filename');
    if (!isPathLike(filename)) {
      var pathErr = new TypeError('The "filename" argument must be of type string or an ' +
        'instance of Buffer or URL. Received ' +
        (typeof filename === 'string' ? "type string ('" + filename + "')" : typeof filename));
      pathErr.code = 'ERR_INVALID_ARG_TYPE';
      throw pathErr;
    }
    if (typeof options === 'function') { listener = options; options = undefined; }
    if (typeof options === 'string') options = { encoding: options };
    if (options === null || options === undefined) options = {};
    if (typeof options !== 'object') throw argError('options', 'object', options);
    if (options.recursive != null && typeof options.recursive !== 'boolean') {
      throw argError('options.recursive', 'boolean', options.recursive);
    }
    if (options.persistent != null && typeof options.persistent !== 'boolean') {
      throw argError('options.persistent', 'boolean', options.persistent);
    }
    if (options.throwIfNoEntry != null && typeof options.throwIfNoEntry !== 'boolean') {
      throw argError('options.throwIfNoEntry', 'boolean', options.throwIfNoEntry);
    }
    if (options.encoding != null) {
      if (typeof options.encoding !== 'string') {
        var encErr = new TypeError('The argument \'options.encoding\' is invalid. Received ' +
          typeof options.encoding);
        encErr.code = 'ERR_INVALID_ARG_VALUE';
        throw encErr;
      }
      if (options.encoding !== 'buffer' && !Buffer.isEncoding(options.encoding)) {
        throw invalidEncodingValue(options.encoding);
      }
    }
    // Node's ignore option filters events by a glob, RegExp, function or
    // array of those; a glob without a slash is matched against the basename.
    var ignore = options.ignore;
    var ignoreMatcher = null;
    function globRegExp(pattern) {
      var out = '^';
      for (var i = 0; i < pattern.length; i++) {
        var ch = pattern.charAt(i);
        if (ch === '*') {
          if (pattern.charAt(i + 1) === '*') {
            out += '.*';
            i++;
            if (pattern.charAt(i + 1) === '/') i++;
          } else {
            out += '[^/]*';
          }
        } else if (ch === '?') {
          out += '[^/]';
        } else if ('\\^$.|+()[]{}'.indexOf(ch) !== -1) {
          out += '\\' + ch;
        } else {
          out += ch;
        }
      }
      return new RegExp(out + '$');
    }
    function compileIgnore(value) {
      if (typeof value === 'function') return value;
      if (value instanceof RegExp) return function (name) { return value.test(name); };
      if (typeof value === 'string') {
        if (value === '') {
          var emptyErr = new TypeError('The argument \'ignore\' must not be an empty string');
          emptyErr.code = 'ERR_INVALID_ARG_VALUE';
          throw emptyErr;
        }
        var re = globRegExp(value);
        var basenameOnly = value.indexOf('/') === -1;
        return function (name) {
          if (basenameOnly) {
            var slash = name.lastIndexOf('/');
            return re.test(slash === -1 ? name : name.slice(slash + 1));
          }
          return re.test(name);
        };
      }
      if (Array.isArray(value)) {
        var parts = value.map(compileIgnore);
        return function (name, type) {
          for (var i = 0; i < parts.length; i++) if (parts[i](name, type)) return true;
          return false;
        };
      }
      var typeErr = new TypeError('The "ignore" argument must be of type string, ' +
        'RegExp, function, or an array of those. Received ' + typeof value);
      typeErr.code = 'ERR_INVALID_ARG_TYPE';
      throw typeErr;
    }
    if (ignore !== undefined) ignoreMatcher = compileIgnore(ignore);
    var emitter = new FSWatcher();
    var path = String(filename);
    var recursive = options.recursive === true;
    var wantBuffer = options.encoding === 'buffer';
    var encName = wantBuffer || options.encoding === undefined ? null : options.encoding;
    var throwIfNoEntry = options.throwIfNoEntry !== false;
    var initial = statOrNull(path);
    // Node's watcher start is synchronous, so a missing path throws unless
    // throwIfNoEntry is false; a path that disappears later emits 'error'.
    if (!initial && throwIfNoEntry) {
      var missing = new Error("ENOENT: no such file or directory, watch '" + path + "'");
      missing.code = 'ENOENT';
      missing.errno = -2;
      missing.syscall = 'watch';
      missing.path = path;
      missing.filename = path;
      throw missing;
    }
    emitter.path = path;
    // An AbortSignal closes the watcher when aborted; an already-aborted one
    // makes fs.watch throw AbortError without starting anything.
    if (options.signal !== undefined) {
      var signal = options.signal;
      if (!signal || typeof signal.addEventListener !== 'function') {
        var sigErr = new TypeError('The "options.signal" property must be an instance of AbortSignal. Received ' + typeof signal);
        sigErr.code = 'ERR_INVALID_ARG_TYPE';
        throw sigErr;
      }
      // Node returns an already-closed watcher for an aborted signal rather
      // than throwing; 'close' fires on the next tick so listeners added by
      // the caller still observe it.
      if (signal.aborted) {
        emitter.close();
        return emitter;
      }
      signal.addEventListener('abort', function () { emitter.close(); }, { once: true });
    }
    var previous = initial ? snapshot(path, recursive) : Object.create(null);
    var interval = options.interval || 100;
    emitter._timer = setWatcherInterval(function () {
      if (emitter._closed) return;
      var next = snapshot(path, recursive);
      var changes = diffSnapshots(previous, next);
      if (!changes.length) return;
      previous = next;
      // Node reports paths relative to the watched directory for recursive
      // watchers and the bare basename for non-recursive ones.
      var root = path.replace(/\/+$/, '');
      for (var ci = 0; ci < changes.length; ci++) {
        var change = changes[ci];
        var reported = change.name;
        if (recursive && reported.length > root.length) reported = reported.slice(root.length + 1);
        else reported = basename(reported);
        if (ignoreMatcher && ignoreMatcher(reported, change.type)) continue;
        if (wantBuffer) reported = Buffer.from(reported);
        else if (encName !== null && encName !== 'utf8' && encName !== 'utf-8') {
          reported = Buffer.from(reported).toString(encName);
        }
        emitter.emit('change', change.type, reported);
        if (emitter._closed) return;
      }
    }, interval);
    if (typeof listener === 'function') emitter.on('change', listener);
    return emitter;
  }

  var statWatchers = Object.create(null);
  // zeroStats builds the all-zero Stats Node passes when a watched file does
  // not exist. It avoids 'new' so it also works in builds where host classes
  // are not constructible.
  function zeroStats() {
    var epoch = Date.UTC(1970, 0, 1);
    var o = Object.create(fs.Stats.prototype);
    o.dev = 0; o.mode = 0; o.nlink = 0; o.uid = 0; o.gid = 0; o.rdev = 0;
    o.blksize = 0; o.ino = 0; o.size = 0; o.blocks = 0;
    o.atimeMs = epoch; o.mtimeMs = epoch; o.ctimeMs = epoch; o.birthtimeMs = epoch;
    o.atime = new Date(epoch); o.mtime = new Date(epoch);
    o.ctime = new Date(epoch); o.birthtime = new Date(epoch);
    Object.defineProperty(o, '_kind', { value: 'file' });
    return o;
  }
  function watchFile(filename, options, listener) {
    assertPathHasNoNullBytes(filename, 'filename');
    if (typeof options === 'function') { listener = options; options = {}; }
    if (!isPathLike(filename)) throw argError('filename', 'string or an instance of Buffer or URL', filename);
    if (typeof listener !== 'function') {
      var listenerErr = new TypeError('The "listener" argument must be of type function. Received ' +
        (typeof listener === 'string' ? "type string ('" + listener + "')" : typeof listener));
      listenerErr.code = 'ERR_INVALID_ARG_TYPE';
      throw listenerErr;
    }
    if (options === null || options === undefined) options = {};
    if (typeof options !== 'object') throw argError('options', 'object', options);
    var key = String(filename);
    if (statWatchers[key]) { statWatchers[key].on('change', listener); return statWatchers[key]; }
    var emitter = new EventEmitter();
    var prev = statOrNull(key);
    var interval = typeof options.interval === 'number' && options.interval >= 0 ? options.interval : 5007;
    var timer = setWatcherInterval(function () {
      var curr = statOrNull(key);
      if (!curr) return;
      if (!prev || curr.mtimeMs !== prev.mtimeMs || curr.size !== prev.size || curr.ino !== prev.ino) {
        var old = prev || zeroStats();
        prev = curr;
        emitter.emit('change', curr, old);
      }
    }, interval);
    // Node reports a missing file once with two all-zero Stats objects, so a
    // listener can observe the ENOENT start state before the file appears.
    if (!prev) {
      helpers.tick(function () {
        if (emitter._stopped || emitter._closed) return;
        var zero = zeroStats();
        emitter.emit('change', zero, zero);
      });
    }
    emitter.close = function () {
      clearWatcherInterval(timer);
      delete statWatchers[key];
    };
    // Node's StatWatcher.stop() releases the handle and emits 'stop' on the
    // next tick, so a listener added before stop() still fires afterwards.
    emitter.stop = function () {
      if (emitter._stopped) return;
      emitter._stopped = true;
      emitter.close();
      helpers.tick(function () { emitter.emit('stop'); });
    };
    emitter.ref = function () {
      if (timer && typeof timer.ref === 'function') timer.ref();
      return emitter;
    };
    emitter.unref = function () {
      if (timer && typeof timer.unref === 'function') timer.unref();
      return emitter;
    };
    statWatchers[key] = emitter;
    if (typeof listener === 'function') emitter.on('change', listener);
    return emitter;
  }
  function unwatchFile(filename, listener) {
    assertPathHasNoNullBytes(filename, 'filename');
    var key = String(filename);
    var emitter = statWatchers[key];
    if (!emitter) return;
    if (typeof listener === 'function') {
      emitter.removeListener('change', listener);
      if (emitter.listenerCount('change') > 0) return;
    }
    emitter.stop();
  }

  // Node only defines lchmod where the platform can open a symlink without
  // following it (O_SYMLINK); on Linux the API is absent and test-fs-chmod-mask
  // skips it. extras.go installs working stubs for the platforms that have it.
  if (fs.constants.O_SYMLINK === undefined) {
    delete fs.lchmod;
    delete fs.lchmodSync;
    if (fs.promises) delete fs.promises.lchmod;
  }
  fs.watch = watch;
  fs.watchFile = watchFile;
  fs.unwatchFile = unwatchFile;
  fs.FSWatcher = FSWatcher;

  // fs.F_OK and friends are deprecated aliases of fs.constants: Node exposes
  // them as getters that warn (DEP0176) on access and reject assignment.
  ['F_OK', 'R_OK', 'W_OK', 'X_OK'].forEach(function (name) {
    Object.defineProperty(fs, name, {
      configurable: true,
      enumerable: true,
      get: function () {
        if (typeof process === 'object' && typeof process.emitWarning === 'function') {
          process.emitWarning('fs.' + name + ' is deprecated, use fs.constants.' + name + ' instead',
            { type: 'DeprecationWarning', code: 'DEP0176' });
        }
        return fs.constants[name];
      },
      set: function () {
        throw new TypeError('Cannot set property ' + name + ' of #<Object> which has only a getter');
      }
    });
  });

  // ------------------------------------------------------------------ Dir --
  function Dir(path, options) {
    if (!(this instanceof Dir)) return new Dir(path, options);
    if (arguments.length < 1) {
      var missingPath = new TypeError('The "path" argument must be specified');
      missingPath.code = 'ERR_MISSING_ARGS';
      throw missingPath;
    }
    if (!isPathLike(path)) {
      var dirPathErr = new TypeError('The "path" argument must be of type string or ' +
        'an instance of Buffer or URL. Received ' +
        (path === null ? 'null' : (Array.isArray(path) ? 'an instance of Array' : typeof path)));
      dirPathErr.code = 'ERR_INVALID_ARG_TYPE';
      throw dirPathErr;
    }
    options = normaliseStreamOptions(options);
    if (options.bufferSize !== undefined) {
      var size = options.bufferSize;
      if (typeof size !== 'number') {
        var sizeType = new TypeError('The "options.bufferSize" property must be of type ' +
          'number. Received ' + (size === null ? 'null' : 'type ' + typeof size));
        sizeType.code = 'ERR_INVALID_ARG_TYPE';
        throw sizeType;
      }
      if (!Number.isInteger(size) || size <= 0 || size > 0xFFFFFFFF) {
        var sizeErr = new RangeError('The value of "options.bufferSize" is out of range. ' +
          'It must be > 0 && <= 4294967295. Received ' + size);
        sizeErr.code = 'ERR_OUT_OF_RANGE';
        throw sizeErr;
      }
    }
    this._path = path;
    this._options = options;
    this._reading = false;
    var readdirOpts = { withFileTypes: true };
    if (options.encoding !== undefined) readdirOpts.encoding = options.encoding;
    this._entries = fs.readdirSync(path, readdirOpts);
    this._index = 0;
    this._closed = false;
  }
  // Node's dir.path is a prototype accessor so reading it off the prototype
  // itself throws ERR_INVALID_THIS instead of returning undefined.
  Object.defineProperty(Dir.prototype, 'path', {
    configurable: true,
    enumerable: false,
    get: function () {
      if (!(this instanceof Dir)) {
        var thisErr = new TypeError('Value of "this" must be of type Dir');
        thisErr.code = 'ERR_INVALID_THIS';
        throw thisErr;
      }
      return this._path;
    }
  });
  function dirStateError(code, message) {
    var err = new Error(message);
    err.code = code;
    return err;
  }
  Dir.prototype.read = function (cb) {
    if (cb !== undefined && typeof cb !== 'function') {
      var readCbErr = new TypeError('The "callback" argument must be of type function. ' +
        'Received ' + typeof cb);
      readCbErr.code = 'ERR_INVALID_ARG_TYPE';
      throw readCbErr;
    }
    var self = this;
    if (!this._closed) this._reading = true;
    var p = new Promise(function (resolve, reject) {
      helpers.tick(function () {
        self._reading = false;
        if (self._closed) {
          reject(dirStateError('ERR_DIR_CLOSED', 'Directory handle was closed'));
          return;
        }
        resolve(self._index < self._entries.length ? self._entries[self._index++] : null);
      });
    });
    if (typeof cb === 'function') {
      p.then(function (v) { cb(null, v); }, function (e) { cb(e); });
      return undefined;
    }
    return p;
  };
  Dir.prototype.readSync = function () {
    if (this._reading) throw dirStateError('ERR_DIR_CONCURRENT_OPERATION',
      'Cannot read from directory handle as the request is already pending');
    if (this._closed) throw dirStateError('ERR_DIR_CLOSED', 'Directory handle was closed');
    return this._index < this._entries.length ? this._entries[this._index++] : null;
  };
  Dir.prototype.close = function (cb) {
    if (cb !== undefined && typeof cb !== 'function') {
      var cbType = new TypeError('The "callback" argument must be of type function. ' +
        'Received ' + typeof cb);
      cbType.code = 'ERR_INVALID_ARG_TYPE';
      throw cbType;
    }
    var self = this;
    var closedErr = this._closed
      ? dirStateError('ERR_DIR_CLOSED', 'Directory handle was closed') : null;
    this._closed = true;
    var p = closedErr ? Promise.reject(closedErr) : Promise.resolve();
    if (typeof cb === 'function') { p.then(function () { cb(null); }, cb); return undefined; }
    return p;
  };
  Dir.prototype.closeSync = function () {
    if (this._reading) throw dirStateError('ERR_DIR_CONCURRENT_OPERATION',
      'Cannot close directory handle as the request is already pending');
    if (this._closed) throw dirStateError('ERR_DIR_CLOSED', 'Directory handle was closed');
    this._closed = true;
  };
  Dir.prototype[Symbol.asyncIterator] = function () {
    var self = this;
    return {
      next: function () {
        return self.read().then(function (d) {
          return d === null ? { value: undefined, done: true } : { value: d, done: false };
        });
      },
      return: function () { return self.close().then(function () { return { value: undefined, done: true }; }); }
    };
  };
  function opendirSync(path, options) { return new Dir(path, options); }
  function opendir(path, options, cb) {
    if (typeof options === 'function') { cb = options; options = {}; }
    if (!isPathLike(path)) {
      var odPathErr = new TypeError('The "path" argument must be of type string or ' +
        'an instance of Buffer or URL. Received ' +
        (path === null ? 'null' : (Array.isArray(path) ? 'an instance of Array' : typeof path)));
      odPathErr.code = 'ERR_INVALID_ARG_TYPE';
      throw odPathErr;
    }
    if (typeof cb !== 'function') {
      // fs.opendir is the callback form; only fs.promises.opendir returns a
      // Promise, so a missing callback is an argument error in Node.
      var cbErr = new TypeError('The "callback" argument must be of type function. ' +
        'Received ' + (cb === undefined ? 'undefined' : typeof cb));
      cbErr.code = 'ERR_INVALID_ARG_TYPE';
      throw cbErr;
    }
    if (typeof cb === 'function') {
      // The async form never throws: an ENOTDIR/EACCES failure is delivered to
      // the callback instead (Node constructs the handle in the request).
      helpers.tick(function () {
        var d;
        try { d = new Dir(path, options); } catch (err) { cb(err); return; }
        cb(null, d);
      });
      return undefined;
    }
    return new Promise(function (resolve, reject) {
      helpers.tick(function () {
        try { resolve(new Dir(path, options)); } catch (err) { reject(err); }
      });
    });
  }
  fs.Dir = Dir;
  fs.opendirSync = opendirSync;
  fs.opendir = opendir;
  if (fs.promises) {
    fs.promises.opendir = function (path, options) {
      return new Promise(function (resolve, reject) {
        helpers.tick(function () {
          try { resolve(new Dir(path, options)); } catch (err) { reject(err); }
        });
      });
    };
    fs.promises.watch = function (filename, options) {
      // An async generator never throws synchronously: a bad path or option
      // rejects the first next() call instead.
      var watcher = null, startErr = null;
      try { watcher = watch(filename, options); } catch (startFailure) { startErr = startFailure; }
      var queue = [], waiters = [], closed = false;
      function push(ev) {
        if (waiters.length) waiters.shift().resolve({ value: ev, done: false });
        else queue.push(ev);
      }
      if (watcher) {
        watcher.on('change', function (type, name) { push({ eventType: type, filename: name }); });
        watcher.on('error', function (err) {
          closed = true;
          while (waiters.length) waiters.shift().reject(err);
        });
        watcher.on('close', function () {
          closed = true;
          while (waiters.length) waiters.shift().resolve({ value: undefined, done: true });
        });
      } else {
        closed = true;
      }
      var iterator = {
        next: function () {
          if (startErr) {
            var pending = startErr;
            startErr = null;
            return Promise.reject(pending);
          }
          if (queue.length) return Promise.resolve({ value: queue.shift(), done: false });
          if (closed) return Promise.resolve({ value: undefined, done: true });
          return new Promise(function (resolve, reject) { waiters.push({ resolve: resolve, reject: reject }); });
        },
        return: function () {
          if (watcher) watcher.close();
          return Promise.resolve({ value: undefined, done: true });
        }
      };
      iterator[Symbol.asyncIterator] = function () { return this; };
      return iterator;
    };
  }
})
`
