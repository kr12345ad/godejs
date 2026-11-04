// Package wasi implements Node's WebAssembly System Interface binding. The
// import object is backed by the same synchronous JavaScript thread as the
// WebAssembly instance, so preview1 calls operate directly on memory views.
package wasi

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("wasi", Install) }

type module struct {
	e     *env.Env
	state map[float64]*wasiState
	seq   float64
}

type preopen struct {
	fd     uint32
	mapped string
	real   string
}

type wasiState struct {
	m        *module
	id       float64
	args     []string
	env      []string
	preopens []preopen
	stdio    [3]int
	exitCode uint32
	files    map[uint32]*os.File
	nextFD   uint32
	dirFds   map[uint32]preopen
}

const (
	errSuccess     = 0
	err2big        = 1
	errAcces       = 2
	errBadf        = 8
	errExist       = 20
	errFault       = 21
	errInval       = 28
	errIO          = 29
	errIsdir       = 31
	errLoop        = 32
	errMfile       = 33
	errMlink       = 34
	errNametoolong = 37
	errNoent       = 44
	errNosys       = 52
	errNotdir      = 54
	errNotempty    = 55
	errNotcap      = 76
	errOverflow    = 61
)

func Install(e *env.Env) js.Object {
	m := &module{e: e, state: map[float64]*wasiState{}}
	host := e.JS.Obj()
	host.SetFuncHidden("new", m.hostNew)
	host.SetFuncHidden("call", m.hostCall)
	api := e.JS.Obj()
	api.Set("new", host.Get("new"))
	api.Set("call", host.Get("call"))
	exports := e.JS.Obj()
	fn, err := e.JS.Eval(`(function (host) {
  'use strict';
  var errno = { success: 0, badf: 8, inval: 28, io: 29, noent: 44, nosys: 52, notdir: 54, notempty: 55 };

  class WASI {
    constructor(options) {
      options = options || {};
      if (typeof options !== 'object') throw new TypeError('The "options" argument must be of type object');
      if (options.version !== 'preview1' && options.version !== 'unstable') {
        var e = new TypeError('The argument \'options.version\' must be one of: \'unstable\', \'preview1\'. Received \'' + options.version + '\'');
        e.code = 'ERR_INVALID_ARG_VALUE';
        throw e;
      }
      var args = (options.args || []).map(String);
      var envPairs = [];
      if (options.env) Object.keys(options.env).forEach(function (k) {
        if (options.env[k] !== undefined) envPairs.push(k + '=' + options.env[k]);
      });
      var preopens = [];
      if (options.preopens) Object.keys(options.preopens).forEach(function (k) {
        preopens.push(String(k), String(options.preopens[k]));
      });
      var stdio = [options.stdin === undefined ? 0 : options.stdin,
                   options.stdout === undefined ? 1 : options.stdout,
                   options.stderr === undefined ? 2 : options.stderr];
      this._id = host.new.call(this, args, envPairs, preopens, stdio, options.version);
      this._exitCode = 0;
      this._started = false;
      this._memory = null;
      this._returnOnExit = options.returnOnExit !== false;
      var self = this;
      var api = {};
      var names = ['args_get','args_sizes_get','environ_get','environ_sizes_get',
        'clock_res_get','clock_time_get','fd_advise','fd_allocate','fd_close',
        'fd_datasync','fd_fdstat_get','fd_fdstat_set_flags','fd_fdstat_set_rights',
        'fd_filestat_get','fd_filestat_set_size','fd_filestat_set_times','fd_pread',
        'fd_prestat_get','fd_prestat_dir_name','fd_pwrite','fd_read','fd_readdir',
        'fd_renumber','fd_seek','fd_sync','fd_tell','fd_write','path_create_directory',
        'path_filestat_get','path_filestat_set_times','path_link','path_open',
        'path_readlink','path_remove_directory','path_rename','path_symlink',
        'path_unlink_file','poll_oneoff','proc_raise','random_get','sched_yield',
        'sock_accept','sock_recv','sock_send','sock_shutdown'];
      function memoryView() {
        if (!self._memory) {
          var e = new Error('WASI is not started');
          e.code = 'ERR_WASI_NOT_STARTED';
          throw e;
        }
        return new Uint8Array(self._memory.buffer);
      }
      names.forEach(function (name) {
        api[name] = function () {
          var a = [name, memoryView()];
          for (var i = 0; i < arguments.length; i++) a.push(arguments[i]);
          return host.call.apply(self, a);
        };
      });
      api.proc_exit = function (code) {
        code = code >>> 0;
        if (self._returnOnExit) {
          self._exitCode = code;
          var stop = new Error('WASI proc_exit');
          stop.code = 'ERR_WASI_EXIT';
          throw stop;
        }
        if (typeof process !== 'undefined' && process.exit) process.exit(code);
      };
      api._wasiId = this._id;
      this.wasiImport = api;
      this._bindingName = options.version === 'unstable' ? 'wasi_unstable' : 'wasi_snapshot_preview1';
    }
    finalizeBindings(instance, options) {
      if (this._started) {
        var e = new Error('WASI is already started');
        e.code = 'ERR_WASI_ALREADY_STARTED';
        throw e;
      }
      if (!instance || typeof instance !== 'object' || !instance.exports ||
          !instance.exports.memory) {
        throw new TypeError('The "instance" argument must be a WebAssembly instance with exported memory');
      }
      this._memory = instance.exports.memory;
      this._started = true;
    }
    start(instance) {
      this.finalizeBindings(instance);
      var exp = instance.exports;
      if (typeof exp._start !== 'function') throw new TypeError('The "instance.exports._start" property must be a function');
      if (exp._initialize !== undefined) throw new TypeError('The "instance.exports._initialize" property must be undefined');
      try { exp._start(); }
      catch (err) {
        if (!err || err.code !== 'ERR_WASI_EXIT') throw err;
      }
      return this._exitCode;
    }
    initialize(instance) {
      this.finalizeBindings(instance);
      var exp = instance.exports;
      if (exp._start !== undefined) throw new TypeError('The "instance.exports._start" property must be undefined');
      if (exp._initialize !== undefined) {
        if (typeof exp._initialize !== 'function') throw new TypeError('The "instance.exports._initialize" property must be a function');
        exp._initialize();
      }
    }
    getImportObject() {
      var out = {};
      out[this._bindingName] = this.wasiImport;
      return out;
    }
  }
  return { WASI: WASI, errno: errno };
})`, "wasi.js")
	if err != nil {
		panic("godejs wasi: " + err.Error())
	}
	built, err := fn.Object().Call(js.Object{}, api.V())
	if err != nil {
		panic("godejs wasi: " + err.Error())
	}
	for _, key := range built.Object().OwnKeys() {
		exports.Set(key, built.Object().Get(key))
	}
	return exports
}

func Implemented() []string {
	return []string{
		"wasi.WASI", "wasi.WASI#getImportObject", "wasi.WASI#start", "wasi.WASI#initialize",
		"wasi.WASI#finalizeBindings", "wasi.args_get", "wasi.args_sizes_get",
		"wasi.environ_get", "wasi.environ_sizes_get", "wasi.clock_res_get",
		"wasi.clock_time_get", "wasi.fd_read", "wasi.fd_write", "wasi.fd_close",
		"wasi.fd_fdstat_get", "wasi.fd_filestat_get", "wasi.fd_seek", "wasi.fd_tell",
		"wasi.fd_prestat_get", "wasi.fd_prestat_dir_name", "wasi.fd_readdir",
		"wasi.path_open", "wasi.path_create_directory", "wasi.path_remove_directory",
		"wasi.path_unlink_file", "wasi.path_rename", "wasi.path_filestat_get",
		"wasi.path_readlink", "wasi.path_symlink", "wasi.path_link",
		"wasi.random_get", "wasi.sched_yield", "wasi.proc_exit", "wasi.poll_oneoff",
	}
}

func (m *module) hostNew(this js.Object, args []js.Value) (js.Value, error) {
	m.seq++
	s := &wasiState{
		m:      m,
		id:     m.seq,
		files:  map[uint32]*os.File{},
		dirFds: map[uint32]preopen{},
		nextFD: 3,
	}
	if len(args) > 0 && args[0].IsArray() {
		a := args[0].Object()
		for i := 0; i < a.Length(); i++ {
			s.args = append(s.args, a.At(i).Str())
		}
	}
	if len(args) > 1 && args[1].IsArray() {
		a := args[1].Object()
		for i := 0; i < a.Length(); i++ {
			s.env = append(s.env, a.At(i).Str())
		}
	}
	if len(args) > 2 && args[2].IsArray() {
		a := args[2].Object()
		for i := 0; i+1 < a.Length(); i += 2 {
			p := preopen{fd: s.nextFD, mapped: a.At(i).Str(), real: a.At(i + 1).Str()}
			s.nextFD++
			s.preopens = append(s.preopens, p)
			s.dirFds[p.fd] = p
		}
	}
	if len(args) > 3 && args[3].IsArray() {
		a := args[3].Object()
		for i := 0; i < 3 && i < a.Length(); i++ {
			s.stdio[i] = int(a.At(i).Int())
		}
	}
	m.state[s.id] = s
	this.SetNum("_wasiId", s.id)
	return m.e.JS.Num(s.id), nil
}

func (m *module) stateOf(this js.Object) *wasiState {
	v := this.Get("_wasiId")
	if !v.IsNumber() {
		return nil
	}
	return m.state[v.Float()]
}

func (m *module) hostCall(this js.Object, args []js.Value) (js.Value, error) {
	s := m.stateOf(this)
	if s == nil {
		return m.e.JS.Num(errInval), nil
	}
	if len(args) == 0 {
		return m.e.JS.Num(errInval), nil
	}
	op := args[0].Str()
	if len(args) < 2 {
		return m.e.JS.Num(errInval), nil
	}
	// ByteView aliases WebAssembly memory. Bytes() would copy it, so writes
	// such as fd_write's nwritten out-parameter would never reach the guest.
	mem, ok := args[1].ByteView()
	if !ok {
		return m.e.JS.Num(errFault), nil
	}
	a := args[2:]
	u32 := func(i int) uint32 {
		if i >= len(a) {
			return 0
		}
		return uint32(a[i].Int())
	}
	u64 := func(i int) uint64 {
		if i >= len(a) {
			return 0
		}
		if a[i].IsBigInt() {
			n, _ := parseInt64(a[i].Str())
			return uint64(n)
		}
		return uint64(a[i].Int())
	}
	code := s.call(op, mem, u32, u64, a)
	return m.e.JS.Num(float64(code)), nil
}

func parseInt64(s string) (int64, bool) {
	var n int64
	if _, err := fmt.Sscan(strings.TrimSuffix(s, "n"), &n); err != nil {
		return 0, false
	}
	return n, true
}

func (s *wasiState) call(op string, mem []byte, u32 func(int) uint32, u64 func(int) uint64, a []js.Value) uint32 {
	switch op {
	case "args_sizes_get":
		return s.sizes(mem, u32(0), u32(1), s.args)
	case "args_get":
		return s.stringsGet(mem, u32(0), u32(1), s.args)
	case "environ_sizes_get":
		return s.sizes(mem, u32(0), u32(1), s.env)
	case "environ_get":
		return s.stringsGet(mem, u32(0), u32(1), s.env)
	case "clock_res_get":
		return writeU64(mem, u32(1), 1)
	case "clock_time_get":
		return s.clockTime(mem, u32(0), u32(2))
	case "fd_write":
		return s.fdWrite(mem, u32(0), u32(1), u32(2), u32(3))
	case "fd_read":
		return s.fdRead(mem, u32(0), u32(1), u32(2), u32(3))
	case "fd_close":
		return s.fdClose(u32(0))
	case "fd_fdstat_get":
		return s.fdFdstatGet(mem, u32(0), u32(1))
	case "fd_filestat_get":
		return s.fdFilestatGet(mem, u32(0), u32(1))
	case "fd_filestat_set_size":
		return s.fdSetSize(u32(0), u64(1))
	case "fd_filestat_set_times":
		return errSuccess
	case "fd_seek":
		return s.fdSeek(mem, u32(0), int64(u64(1)), u32(2), u32(3))
	case "fd_tell":
		return s.fdTell(mem, u32(0), u32(1))
	case "fd_sync", "fd_datasync":
		return s.fdSync(u32(0))
	case "fd_advise", "fd_allocate", "fd_fdstat_set_flags", "fd_fdstat_set_rights":
		return errSuccess
	case "fd_renumber":
		return s.fdRenumber(u32(0), u32(1))
	case "fd_prestat_get":
		return s.fdPrestatGet(mem, u32(0), u32(1))
	case "fd_prestat_dir_name":
		return s.fdPrestatDirName(mem, u32(0), u32(1), u32(2))
	case "fd_readdir":
		return s.fdReaddir(mem, u32(0), u32(1), u32(2), u64(3), u32(4))
	case "path_open":
		return s.pathOpen(mem, u32(0), u32(2), u32(3), u32(4), u64(5), u64(6), u32(7), u32(8))
	case "path_create_directory":
		return s.pathCreateDirectory(mem, u32(0), u32(1), u32(2))
	case "path_remove_directory":
		return s.pathRemoveDirectory(mem, u32(0), u32(1), u32(2))
	case "path_unlink_file":
		return s.pathUnlink(mem, u32(0), u32(1), u32(2))
	case "path_rename":
		return s.pathRename(mem, u32(0), u32(1), u32(2), u32(3), u32(4), u32(5))
	case "path_filestat_get":
		return s.pathFilestatGet(mem, u32(0), u32(1), u32(2), u32(3), u32(4))
	case "path_readlink":
		return s.pathReadlink(mem, u32(0), u32(1), u32(2), u32(3), u32(4), u32(5))
	case "path_symlink":
		return s.pathSymlink(mem, u32(0), u32(1), u32(2), u32(3), u32(4))
	case "path_link":
		return s.pathLink(mem, u32(0), u32(1), u32(2), u32(3), u32(4), u32(5), u32(6))
	case "path_filestat_set_times":
		return errSuccess
	case "random_get":
		return s.randomGet(mem, u32(0), u32(1))
	case "sched_yield":
		return errSuccess
	case "poll_oneoff":
		return s.pollOneoff(mem, u32(0), u32(1), u32(2), u32(3))
	case "proc_raise":
		return errNosys
	case "sock_accept", "sock_recv", "sock_send", "sock_shutdown":
		return errNosys
	}
	return errNosys
}

func (s *wasiState) sizes(mem []byte, countPtr, sizePtr uint32, values []string) uint32 {
	total := uint32(0)
	for _, v := range values {
		total += uint32(len(v) + 1)
	}
	if !putU32(mem, countPtr, uint32(len(values))) {
		return errFault
	}
	if !putU32(mem, sizePtr, total) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) stringsGet(mem []byte, ptrs, buf uint32, values []string) uint32 {
	cur := buf
	for i, v := range values {
		if !putU32(mem, ptrs+uint32(i)*4, cur) {
			return errFault
		}
		if !putBytes(mem, cur, []byte(v+"\x00")) {
			return errFault
		}
		cur += uint32(len(v) + 1)
	}
	return errSuccess
}

func (s *wasiState) clockTime(mem []byte, id, ptr uint32) uint32 {
	var t time.Time
	switch id {
	case 0:
		t = time.Now()
	case 1:
		t = time.Now()
	case 2, 3:
		t = time.Now()
	default:
		return errInval
	}
	return writeU64(mem, ptr, uint64(t.UnixNano()))
}

func (s *wasiState) fdWrite(mem []byte, fd, iovs, iovsLen, nwritten uint32) uint32 {
	if fd != 1 && fd != 2 && fd != uint32(s.stdio[1]) && fd != uint32(s.stdio[2]) {
		if f := s.files[fd]; f != nil {
			return s.writeFile(mem, f, iovs, iovsLen, nwritten)
		}
		return errBadf
	}
	w := s.m.e.Out()
	if fd == uint32(s.stdio[2]) || fd == 2 {
		w = s.m.e.Err()
	}
	var total uint32
	for i := uint32(0); i < iovsLen; i++ {
		p, n, ok := iovec(mem, iovs+i*8)
		if !ok {
			return errFault
		}
		b, ok := getBytes(mem, p, n)
		if !ok {
			return errFault
		}
		if _, err := w.Write(b); err != nil {
			return errIO
		}
		total += n
	}
	if !putU32(mem, nwritten, total) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) writeFile(mem []byte, f *os.File, iovs, iovsLen, nwritten uint32) uint32 {
	var total uint32
	for i := uint32(0); i < iovsLen; i++ {
		p, n, ok := iovec(mem, iovs+i*8)
		if !ok {
			return errFault
		}
		b, ok := getBytes(mem, p, n)
		if !ok {
			return errFault
		}
		written, err := f.Write(b)
		total += uint32(written)
		if err != nil {
			return errIO
		}
	}
	if !putU32(mem, nwritten, total) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) fdRead(mem []byte, fd, iovs, iovsLen, nread uint32) uint32 {
	var r io.Reader
	if fd == 0 || fd == uint32(s.stdio[0]) {
		r = s.m.e.In()
	} else if f := s.files[fd]; f != nil {
		r = f
	} else {
		return errBadf
	}
	var total uint32
	for i := uint32(0); i < iovsLen; i++ {
		p, n, ok := iovec(mem, iovs+i*8)
		if !ok {
			return errFault
		}
		buf := make([]byte, n)
		got, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return errIO
		}
		if got > 0 {
			if !putBytes(mem, p, buf[:got]) {
				return errFault
			}
			total += uint32(got)
		}
		if err == io.EOF {
			break
		}
	}
	if !putU32(mem, nread, total) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) fdClose(fd uint32) uint32 {
	if fd < 3 {
		return errSuccess
	}
	if f := s.files[fd]; f != nil {
		_ = f.Close()
		delete(s.files, fd)
		return errSuccess
	}
	if _, ok := s.dirFds[fd]; ok {
		return errSuccess
	}
	return errBadf
}

func (s *wasiState) fdFdstatGet(mem []byte, fd, ptr uint32) uint32 {
	ft := uint8(4)
	if fd < 3 {
		if fd == 0 {
			ft = 2
		} else {
			ft = 2
		}
	} else if _, ok := s.dirFds[fd]; ok {
		ft = 3
	} else if f := s.files[fd]; f != nil {
		if st, err := f.Stat(); err == nil && st.IsDir() {
			ft = 3
		}
	} else {
		return errBadf
	}
	// 2-byte flags at offset 2, then 64-bit rights.
	if !putBytes(mem, ptr, make([]byte, 24)) {
		return errFault
	}
	mem[ptr] = ft
	putU64(mem, ptr+8, ^uint64(0))
	putU64(mem, ptr+16, ^uint64(0))
	return errSuccess
}

func (s *wasiState) fdFilestatGet(mem []byte, fd, ptr uint32) uint32 {
	var st os.FileInfo
	var err error
	if fd < 3 {
		return writeFilestat(mem, ptr, nil, 2)
	}
	if f := s.files[fd]; f != nil {
		st, err = f.Stat()
	} else if p, ok := s.dirFds[fd]; ok {
		st, err = os.Stat(p.real)
	} else {
		return errBadf
	}
	if err != nil {
		return errno(err)
	}
	return writeFilestat(mem, ptr, st, 0)
}

func (s *wasiState) fdSetSize(fd uint32, size uint64) uint32 {
	f := s.files[fd]
	if f == nil {
		return errBadf
	}
	if err := f.Truncate(int64(size)); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) fdSeek(mem []byte, fd uint32, offset int64, whence, ptr uint32) uint32 {
	f := s.files[fd]
	if f == nil {
		return errBadf
	}
	w := int(whence)
	if w < 0 || w > 2 {
		return errInval
	}
	n, err := f.Seek(offset, w)
	if err != nil {
		return errno(err)
	}
	if code := writeU64(mem, ptr, uint64(n)); code != errSuccess {
		return code
	}
	return errSuccess
}

func (s *wasiState) fdTell(mem []byte, fd, ptr uint32) uint32 {
	f := s.files[fd]
	if f == nil {
		return errBadf
	}
	n, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return errno(err)
	}
	if code := writeU64(mem, ptr, uint64(n)); code != errSuccess {
		return code
	}
	return errSuccess
}

func (s *wasiState) fdSync(fd uint32) uint32 {
	if f := s.files[fd]; f != nil {
		if err := f.Sync(); err != nil {
			return errno(err)
		}
		return errSuccess
	}
	if fd < 3 {
		return errSuccess
	}
	return errBadf
}

func (s *wasiState) fdRenumber(from, to uint32) uint32 {
	f := s.files[from]
	if f == nil {
		return errBadf
	}
	if old := s.files[to]; old != nil {
		_ = old.Close()
	}
	s.files[to] = f
	delete(s.files, from)
	return errSuccess
}

func (s *wasiState) fdPrestatGet(mem []byte, fd, ptr uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	if !putBytes(mem, ptr, make([]byte, 8)) {
		return errFault
	}
	mem[ptr] = 0
	putU32(mem, ptr+4, uint32(len(p.mapped)))
	return errSuccess
}

func (s *wasiState) fdPrestatDirName(mem []byte, fd, ptr, length uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	if length < uint32(len(p.mapped)) {
		return errNametoolong
	}
	if !putBytes(mem, ptr, []byte(p.mapped)) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) fdReaddir(mem []byte, fd, buf, bufLen uint32, cookie uint64, used uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	entries, err := os.ReadDir(p.real)
	if err != nil {
		return errno(err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	start := int(cookie)
	if start > len(entries) {
		start = len(entries)
	}
	off := uint32(0)
	for i := start; i < len(entries); i++ {
		name := entries[i].Name()
		need := uint32(24 + len(name))
		if off+need > bufLen {
			break
		}
		base := buf + off
		if !putBytes(mem, base, make([]byte, 24)) {
			return errFault
		}
		putU64(mem, base, uint64(i+1))
		putU64(mem, base+8, 0)
		putU32(mem, base+16, uint32(len(name)))
		kind := uint8(0)
		if entries[i].IsDir() {
			kind = 3
		} else if entries[i].Type().IsRegular() {
			kind = 4
		} else if entries[i].Type()&os.ModeSymlink != 0 {
			kind = 7
		}
		mem[base+20] = kind
		copy(mem[base+24:], name)
		off += need
	}
	if !putU32(mem, used, off) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) pathOpen(mem []byte, dirfd, pathPtr, pathLen, oflags uint32, rightsBase, rightsInh uint64, fdflags, fdPtr uint32) uint32 {
	p, ok := s.dirFds[dirfd]
	if !ok {
		return errBadf
	}
	name, ok := getBytes(mem, pathPtr, pathLen)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(name))
	if !ok {
		return errNotcap
	}
	flags := os.O_RDWR
	if oflags&1 != 0 {
		flags |= os.O_CREATE
	}
	if oflags&4 != 0 {
		flags |= os.O_EXCL
	}
	if oflags&8 != 0 {
		flags |= os.O_TRUNC
	}
	if fdflags&1 != 0 {
		flags |= os.O_APPEND
	}
	if oflags&2 != 0 {
		f, err := os.Open(full)
		if err != nil {
			return errno(err)
		}
		fd := s.allocFD(f)
		if !putU32(mem, fdPtr, fd) {
			_ = f.Close()
			return errFault
		}
		return errSuccess
	}
	f, err := os.OpenFile(full, flags, 0o666)
	if err != nil {
		return errno(err)
	}
	fd := s.allocFD(f)
	if !putU32(mem, fdPtr, fd) {
		_ = f.Close()
		return errFault
	}
	return errSuccess
}

func (s *wasiState) allocFD(f *os.File) uint32 {
	fd := s.nextFD
	for s.files[fd] != nil {
		fd++
	}
	if fd >= s.nextFD {
		s.nextFD = fd + 1
	}
	s.files[fd] = f
	return fd
}

func (s *wasiState) resolve(base, name string) (string, bool) {
	full := filepath.Clean(filepath.Join(base, name))
	absBase, err1 := filepath.Abs(base)
	absFull, err2 := filepath.Abs(full)
	if err1 != nil || err2 != nil {
		return "", false
	}
	rel, err := filepath.Rel(absBase, absFull)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return full, true
}

func (s *wasiState) pathCreateDirectory(mem []byte, fd, ptr, length uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	name, ok := getBytes(mem, ptr, length)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(name))
	if !ok {
		return errNotcap
	}
	if err := os.Mkdir(full, 0o777); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) pathRemoveDirectory(mem []byte, fd, ptr, length uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	name, ok := getBytes(mem, ptr, length)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(name))
	if !ok {
		return errNotcap
	}
	if err := os.Remove(full); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) pathUnlink(mem []byte, fd, ptr, length uint32) uint32 {
	return s.pathRemoveDirectory(mem, fd, ptr, length)
}

func (s *wasiState) pathRename(mem []byte, oldFD, oldPtr, oldLen, newFD, newPtr, newLen uint32) uint32 {
	op, ok := s.dirFds[oldFD]
	if !ok {
		return errBadf
	}
	np, ok := s.dirFds[newFD]
	if !ok {
		return errBadf
	}
	oldName, ok := getBytes(mem, oldPtr, oldLen)
	if !ok {
		return errFault
	}
	newName, ok := getBytes(mem, newPtr, newLen)
	if !ok {
		return errFault
	}
	oldFull, ok := s.resolve(op.real, string(oldName))
	if !ok {
		return errNotcap
	}
	newFull, ok := s.resolve(np.real, string(newName))
	if !ok {
		return errNotcap
	}
	if err := os.Rename(oldFull, newFull); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) pathFilestatGet(mem []byte, fd, flags, ptr, length, buf uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	name, ok := getBytes(mem, ptr, length)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(name))
	if !ok {
		return errNotcap
	}
	st, err := os.Stat(full)
	if err != nil {
		return errno(err)
	}
	return writeFilestat(mem, buf, st, 0)
}

func (s *wasiState) pathReadlink(mem []byte, fd, ptr, length, buf, bufLen, used uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	name, ok := getBytes(mem, ptr, length)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(name))
	if !ok {
		return errNotcap
	}
	target, err := os.Readlink(full)
	if err != nil {
		return errno(err)
	}
	if uint32(len(target)) > bufLen {
		return errOverflow
	}
	if !putBytes(mem, buf, []byte(target)) {
		return errFault
	}
	if !putU32(mem, used, uint32(len(target))) {
		return errFault
	}
	return errSuccess
}

func (s *wasiState) pathSymlink(mem []byte, oldPtr, oldLen, fd, newPtr, newLen uint32) uint32 {
	p, ok := s.dirFds[fd]
	if !ok {
		return errBadf
	}
	oldName, ok := getBytes(mem, oldPtr, oldLen)
	if !ok {
		return errFault
	}
	newName, ok := getBytes(mem, newPtr, newLen)
	if !ok {
		return errFault
	}
	full, ok := s.resolve(p.real, string(newName))
	if !ok {
		return errNotcap
	}
	if err := os.Symlink(string(oldName), full); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) pathLink(mem []byte, oldFD, oldFlags, oldPtr, oldLen, newFD, newPtr, newLen uint32) uint32 {
	op, ok := s.dirFds[oldFD]
	if !ok {
		return errBadf
	}
	np, ok := s.dirFds[newFD]
	if !ok {
		return errBadf
	}
	oldName, ok := getBytes(mem, oldPtr, oldLen)
	if !ok {
		return errFault
	}
	newName, ok := getBytes(mem, newPtr, newLen)
	if !ok {
		return errFault
	}
	oldFull, ok := s.resolve(op.real, string(oldName))
	if !ok {
		return errNotcap
	}
	newFull, ok := s.resolve(np.real, string(newName))
	if !ok {
		return errNotcap
	}
	if err := os.Link(oldFull, newFull); err != nil {
		return errno(err)
	}
	return errSuccess
}

func (s *wasiState) randomGet(mem []byte, ptr, length uint32) uint32 {
	if ptr+length > uint32(len(mem)) {
		return errFault
	}
	f, err := os.Open("/dev/urandom")
	if err == nil {
		defer f.Close()
		if _, err := io.ReadFull(f, mem[ptr:ptr+length]); err == nil {
			return errSuccess
		}
	}
	for i := uint32(0); i < length; i++ {
		mem[ptr+i] = byte(time.Now().UnixNano() >> (i % 8 * 8))
	}
	return errSuccess
}

func (s *wasiState) pollOneoff(mem []byte, inPtr, outPtr, nsubs, nevents uint32) uint32 {
	// Each subscription is 48 bytes. Clock subscriptions are the common case
	// for preview1 programs; sleep for the first clock and report one event.
	var wait time.Duration
	for i := uint32(0); i < nsubs; i++ {
		base := inPtr + i*48
		if base+48 > uint32(len(mem)) {
			return errFault
		}
		if mem[base+8] == 0 {
			timeout := binary.LittleEndian.Uint64(mem[base+24:])
			wait = time.Duration(timeout)
		}
	}
	if wait > 0 {
		time.Sleep(wait)
	}
	if nsubs > 0 {
		if !putBytes(mem, outPtr, make([]byte, 32)) {
			return errFault
		}
		if !putU32(mem, nevents, 1) {
			return errFault
		}
	}
	return errSuccess
}

func writeFilestat(mem []byte, ptr uint32, st os.FileInfo, kind uint8) uint32 {
	if !putBytes(mem, ptr, make([]byte, 64)) {
		return errFault
	}
	if st == nil {
		mem[ptr+16] = kind
		return errSuccess
	}
	if st.IsDir() {
		kind = 3
	} else if st.Mode().IsRegular() {
		kind = 4
	} else if st.Mode()&os.ModeSymlink != 0 {
		kind = 7
	}
	mem[ptr+16] = kind
	putU64(mem, ptr+32, uint64(st.Size()))
	mod := st.ModTime().UnixNano()
	putU64(mem, ptr+40, uint64(mod))
	putU64(mem, ptr+48, uint64(mod))
	putU64(mem, ptr+56, uint64(mod))
	return errSuccess
}

func errno(err error) uint32 {
	if err == nil {
		return errSuccess
	}
	if os.IsNotExist(err) {
		return errNoent
	}
	if os.IsPermission(err) {
		return errAcces
	}
	if os.IsExist(err) {
		return errExist
	}
	if strings.Contains(err.Error(), "not a directory") {
		return errNotdir
	}
	if strings.Contains(err.Error(), "directory not empty") {
		return errNotempty
	}
	return errIO
}

func iovec(mem []byte, ptr uint32) (uint32, uint32, bool) {
	if ptr+8 > uint32(len(mem)) {
		return 0, 0, false
	}
	return binary.LittleEndian.Uint32(mem[ptr:]), binary.LittleEndian.Uint32(mem[ptr+4:]), true
}

func getBytes(mem []byte, ptr, length uint32) ([]byte, bool) {
	if ptr+length > uint32(len(mem)) {
		return nil, false
	}
	return mem[ptr : ptr+length], true
}

func putBytes(mem []byte, ptr uint32, b []byte) bool {
	if ptr+uint32(len(b)) > uint32(len(mem)) {
		return false
	}
	copy(mem[ptr:], b)
	return true
}

func putU32(mem []byte, ptr, v uint32) bool {
	if ptr+4 > uint32(len(mem)) {
		return false
	}
	binary.LittleEndian.PutUint32(mem[ptr:], v)
	return true
}

func putU64(mem []byte, ptr uint32, v uint64) bool {
	if ptr+8 > uint32(len(mem)) {
		return false
	}
	binary.LittleEndian.PutUint64(mem[ptr:], v)
	return true
}

func writeU64(mem []byte, ptr uint32, v uint64) uint32 {
	if !putU64(mem, ptr, v) {
		return errFault
	}
	return errSuccess
}
