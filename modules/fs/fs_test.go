package fs

import (
	"os"
	"path/filepath"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
	"godejs/modules/timers"
)

func setup(t *testing.T) (*env.Env, string) {
	t.Helper()
	e := jstest.New(t)
	// fs tests use Buffer for byte-level assertions and timers for watchers.
	buffer.Install(e)
	jstest.Install(t, e, "timers", timers.Install)
	jstest.Install(t, e, "fs", Install)
	dir := t.TempDir()
	e.Global().Set("dir", e.JS.Str(dir))
	return e, dir
}

func TestReadWriteFile(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `fs.writeFileSync(dir + "/a.txt", "hé\n"); fs.readFileSync(dir + "/a.txt", "utf8")`), "hé\n")
	jstest.Equal(t, jstest.Str(t, e, `var b = fs.readFileSync(dir + "/a.txt"); b.length + ":" + b[0] + "," + b[1] + "," + b[2] + "," + b[3]`), "4:104,195,169,10")
	jstest.Equal(t, jstest.Str(t, e, `fs.readFileSync(dir + "/a.txt", "hex")`), "68c3a90a")

	jstest.Equal(t, jstest.Str(t, e, `fs.appendFileSync(dir + "/a.txt", "!"); fs.readFileSync(dir + "/a.txt", "utf8")`), "hé\n!")
	jstest.Equal(t, jstest.Str(t, e, `fs.readFileSync(dir + "/a.txt", {encoding: "utf8"})`), "hé\n!")
	jstest.Equal(t, jstest.Str(t, e, `var buf = fs.readFileSync(dir + "/a.txt"); Buffer.isBuffer(buf) + ":" + buf.toString("utf8")`), "true:hé\n!")

	jstest.Equal(t, jstest.Str(t, e, `fs.writeFileSync(dir + "/b.bin", Buffer.from([0, 255, 65])); var c = fs.readFileSync(dir + "/b.bin"); c.length + ":" + c[1] + ":" + c[2]`), "3:255:65")

	jstest.Equal(t, jstest.Str(t, e, `fs.writeFileSync(dir + "/c.txt", "aGk=", "base64"); fs.readFileSync(dir + "/c.txt", "utf8")`), "hi")
	jstest.Equal(t, jstest.Str(t, e, `fs.writeFileSync(dir + "/d.txt", "6869", "hex"); fs.readFileSync(dir + "/d.txt", "utf8")`), "hi")
}

func TestMkdirAndReaddir(t *testing.T) {
	e, dir := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `fs.mkdirSync(dir + "/x/y", {recursive: true})`), filepath.Join(dir, "x"))
	jstest.Equal(t, jstest.Str(t, e, `fs.mkdirSync(dir + "/x/y", {recursive: true}) === undefined`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/x/y/a.txt", "a");
var d = fs.readdirSync(dir + "/x/y", {withFileTypes: true})[0];
d.name + ":" + d.isFile() + ":" + d.isDirectory() + ":" + (d.parentPath === dir + "/x/y")`), "a.txt:true:false:true")

	jstest.Equal(t, jstest.Str(t, e, `fs.readdirSync(dir + "/x").join(",")`), "y")
}

func TestStat(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/s.txt", "abcd");
var s = fs.statSync(dir + "/s.txt");
[s.isFile(), s.isDirectory(), s.isSymbolicLink(), s.size, s.mode > 0, typeof s.mtimeMs, s.mtime instanceof Date, s.birthtime instanceof Date].join(",")`),
		"true,false,false,4,true,number,true,true")

	jstest.Equal(t, jstest.Str(t, e, `fs.statSync(dir).isDirectory()`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof fs.Stats === "function" && typeof fs.Dirent === "function"`), "true")
}

func TestRenameCopyRemove(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/r.txt", "one");
fs.copyFileSync(dir + "/r.txt", dir + "/c.txt");
fs.renameSync(dir + "/c.txt", dir + "/c2.txt");
fs.readFileSync(dir + "/c2.txt", "utf8")`), "one")

	jstest.Equal(t, jstest.Str(t, e, `
fs.rmSync(dir + "/c2.txt");
fs.existsSync(dir + "/c2.txt")`), "false")

	jstest.Equal(t, jstest.Str(t, e, `
fs.mkdirSync(dir + "/tree/sub", {recursive: true});
fs.writeFileSync(dir + "/tree/sub/leaf.txt", "leaf");
fs.cpSync(dir + "/tree", dir + "/tree-copy", {recursive: true});
fs.readFileSync(dir + "/tree-copy/sub/leaf.txt", "utf8")`), "leaf")

	jstest.Equal(t, jstest.Str(t, e, `
fs.rmSync(dir + "/tree", {recursive: true});
fs.rmSync(dir + "/tree-copy", {recursive: true});
fs.existsSync(dir + "/tree") + ":" + fs.existsSync(dir + "/tree-copy")`), "false:false")

	jstest.Equal(t, jstest.Str(t, e, `fs.rmSync(dir + "/never", {force: true}); "ok"`), "ok")

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/u.txt", "x");
fs.unlinkSync(dir + "/u.txt");
fs.mkdirSync(dir + "/empty");
fs.rmdirSync(dir + "/empty");
fs.writeFileSync(dir + "/t.txt", "12345");
fs.truncateSync(dir + "/t.txt", 2);
fs.existsSync(dir + "/u.txt") + ":" + fs.existsSync(dir + "/empty") + ":" + fs.readFileSync(dir + "/t.txt", "utf8")`), "false:false:12")
}

func TestSymlinkRealpathMkdtemp(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/target.txt", "t");
fs.symlinkSync(dir + "/target.txt", dir + "/link.txt");
var target = fs.readlinkSync(dir + "/link.txt");
var same = fs.realpathSync(dir + "/target.txt") === fs.realpathSync(dir) + "/target.txt";
(target === dir + "/target.txt") + ":" + same + ":" + fs.lstatSync(dir + "/link.txt").isSymbolicLink() + ":" + fs.statSync(dir + "/link.txt").isFile()`),
		"true:true:true:true")

	jstest.Equal(t, jstest.Str(t, e, `
var tmp = fs.mkdtempSync(dir + "/tmp-");
(tmp.indexOf(dir + "/tmp-") === 0) + ":" + fs.statSync(tmp).isDirectory()`), "true:true")
}

func TestErrors(t *testing.T) {
	e, dir := setup(t)

	want := "ENOENT|ENOENT: no such file or directory, open '" + filepath.Join(dir, "nope.txt") + "'"
	jstest.Equal(t, jstest.Str(t, e, `try { fs.readFileSync(dir + "/nope.txt"); "no" } catch (e) { e.code + "|" + e.message }`), want)

	jstest.Equal(t, jstest.Str(t, e, `try { fs.statSync(dir + "/nope.txt") } catch (e) { e.code }`), "ENOENT")
	// Node accepts an integer fd here, so the invalid-type case is a value that
	// is neither a path nor an fd.
	jstest.Equal(t, jstest.Str(t, e, `try { fs.readFileSync({}) } catch (e) { e.name }`), "TypeError")
	jstest.Equal(t, jstest.Str(t, e, `fs.existsSync(dir + "/nope.txt")`), "false")
	jstest.Equal(t, jstest.Str(t, e, `
var ok = "no";
try { fs.accessSync(dir); ok = "yes"; } catch (e) { ok = e.code; }
var bad = "no";
try { fs.accessSync(dir + "/nope.txt", fs.constants.F_OK); } catch (e) { bad = e.code; }
ok + ":" + bad`), "yes:ENOENT")
}

func TestFileDescriptors(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
var fd = fs.openSync(dir + "/fd.bin", "w+");
var w = fs.writeSync(fd, "abc");
var b = Buffer.alloc(3);
var n = fs.readSync(fd, b, 0, 3, 0);
var size = fs.fstatSync(fd).size;
fs.ftruncateSync(fd, 1);
var truncated = fs.fstatSync(fd).size;
fs.closeSync(fd);
[w, n, b.toString(), size, truncated].join(",")`), "3,3,abc,3,1")

	jstest.Equal(t, jstest.Str(t, e, `var fd = fs.openSync(dir + "/fd.bin", "r"); fs.closeSync(fd); try { fs.fstatSync(fd) } catch (e) { e.code }`), "EBADF")
}

func TestPromises(t *testing.T) {
	e, _ := setup(t)

	jstest.Eval(t, e, `
globalThis.__p = "pending";
globalThis.__pe = "pending";	globalThis.__fh = "pending";
	globalThis.__fh2 = "pending";
	globalThis.__fh2r = "pending";
	globalThis.__fh2w = "pending";
(async function () {
  try {
    await fs.promises.writeFile(dir + "/p.txt", "hello");
    var v = await fs.promises.readFile(dir + "/p.txt", "utf8");
    var st = await fs.promises.stat(dir + "/p.txt");
    __p = v + ":" + st.size;
  } catch (e) { __p = "ERR:" + e.code; }
  try {
    await fs.promises.readFile(dir + "/nope.txt");
  } catch (e) { __pe = e.code; }
  try {
    var fh = await fs.promises.open(dir + "/fh.txt", "w+");
    await fh.writeFile("abc");
    await fh.close();
    __fh = fs.readFileSync(dir + "/fh.txt", "utf8");
  } catch (e) { __fh = "ERR:" + e.code; }
  try {
    var h = await fs.promises.open(dir + "/fh2.txt", "w+");
    await h.writeFile("abcdef");
    await h.truncate(3);
    await h.appendFile("Z");
    await h.sync();
    await h.datasync();
    await h.chmod(0o600);
    var hs = await h.stat();
    var hb = Buffer.alloc(2);	var hr = await h.read(hb, 0, 2, 0);
    var hw = await h.write(Buffer.from("Q"), 0, 1, 0);
    __fh2w = hw.bytesWritten + ":" + hw.buffer.toString();
    await h.utimes(new Date(0), new Date(0));
    await h.close();    __fh2 = hs.size + ":" + hr.bytesRead + ":" + hb.toString() + ":" + (h.fd === -1);
    var r = await fs.promises.open(dir + "/fh2.txt", "r");
    __fh2r = await r.readFile("utf8");
    await r.close();
  } catch (e) { __fh2 = "ERR:" + e.code; __fh2r = "ERR:" + e.code; __fh2w = "ERR:" + e.code; }
})()`)

	jstest.Equal(t, jstest.Str(t, e, `__p`), "hello:5")
	jstest.Equal(t, jstest.Str(t, e, `__pe`), "ENOENT")
	jstest.Equal(t, jstest.Str(t, e, `__fh`), "abc")
	jstest.Equal(t, jstest.Str(t, e, `__fh2`), "4:2:ab:true")
	jstest.Equal(t, jstest.Str(t, e, `__fh2r`), "QbcZ")
	jstest.Equal(t, jstest.Str(t, e, `__fh2w`), "1:Q")
}

func TestCallbacks(t *testing.T) {
	e, _ := setup(t)

	jstest.Str(t, e, `fs.writeFileSync(dir + "/cb.txt", "cb-data")`)
	jstest.Str(t, e, `	globalThis.__cb = "pending";
globalThis.__cbe = "pending";
globalThis.__cbw = "pending";
globalThis.__cbd = "pending";
globalThis.__cbs = "pending";
fs.readFile(dir + "/cb.txt", "utf8", function (err, data) { __cb = err ? "ERR:" + err.code : data; });
fs.readFile(dir + "/nope.txt", function (err) { __cbe = err && err.code; });
fs.writeFile(dir + "/cb2.txt", "w", function (err) { __cbw = err ? "ERR:" + err.code : fs.readFileSync(dir + "/cb2.txt", "utf8"); });
fs.readdir(dir, {withFileTypes: true}, function (err, entries) { __cbd = err ? "ERR:" + err.code : entries.length > 0; });
fs.stat(dir + "/cb2.txt", function (err, st) { __cbs = err ? "ERR:" + err.code : st.isFile(); });
"queued"`)

	if err := e.RunMain(); err != nil {
		t.Fatalf("RunMain: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__cb`), "cb-data")
	jstest.Equal(t, jstest.Str(t, e, `__cbe`), "ENOENT")
	jstest.Equal(t, jstest.Str(t, e, `__cbw`), "w")
	jstest.Equal(t, jstest.Str(t, e, `__cbd`), "true")
	jstest.Equal(t, jstest.Str(t, e, `__cbs`), "true")
}

func TestStatfsAndLink(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/l1.txt", "hello");
fs.linkSync(dir + "/l1.txt", dir + "/l2.txt");
fs.readFileSync(dir + "/l2.txt", "utf8")`), "hello")

	jstest.Equal(t, jstest.Str(t, e, `
var st = fs.statfsSync(dir);
["type", "bsize", "frsize", "blocks", "bfree", "bavail", "files", "ffree"]
  .every((k) => typeof st[k] === "number")`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
var st = fs.statfsSync(dir, { bigint: true });
["type", "bsize", "blocks"].every((k) => typeof st[k] === "bigint")`), "true")

	jstest.Equal(t, jstest.Str(t, e, `
try { fs.chownSync(dir + "/l1.txt", -1, -1); "ok" } catch (e) { e.code }`), "ok")

	jstest.Equal(t, jstest.Str(t, e, `
try { fs.openSync(1, "r") } catch (e) { e.code }`), "ERR_INVALID_ARG_TYPE")
}

func TestReadWriteOptions(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/rw.txt", "abcdef");
var fd = fs.openSync(dir + "/rw.txt", "r");
var b = Buffer.alloc(3);
var n = fs.readSync(fd, b, { offset: 0, length: 3, position: 1 });
fs.closeSync(fd);
n + ":" + b.toString()`), "3:bcd")

	jstest.Equal(t, jstest.Str(t, e, `
var fd = fs.openSync(dir + "/rw2.txt", "w+");
var b = Buffer.from("xyz");
var n = fs.writeSync(fd, b, { offset: 1, length: 2, position: 0 });
fs.closeSync(fd);
n + ":" + fs.readFileSync(dir + "/rw2.txt", "utf8")`), "2:yz")

	jstest.Equal(t, jstest.Str(t, e, `
var fd = fs.openSync(dir + "/rw.txt", "r");
var bufs = [Buffer.alloc(2), Buffer.alloc(2)];
var n = fs.readvSync(fd, bufs, 0);
fs.closeSync(fd);	n + ":" + bufs[0].toString() + bufs[1].toString()`), "4:abcd")

	jstest.Equal(t, jstest.Str(t, e, `
try { fs.readSync(0, Buffer.alloc(1), { offset: false }) } catch (e) { e.code }`), "ERR_INVALID_ARG_TYPE")
}

func TestConstants(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.constants.F_OK === 0 && fs.constants.O_RDONLY === 0 &&
fs.constants.O_CREAT > 0 && typeof fs.promises.readFile === "function" &&
typeof fs.readFileSync === "function"`), "true")
}

func TestStreams(t *testing.T) {
	e, _ := setup(t)

	jstest.Str(t, e, `
fs.writeFileSync(dir + "/stream.txt", "hello world");
globalThis.__chunks = "";
globalThis.__ended = false;
var rs = fs.createReadStream(dir + "/stream.txt", { highWaterMark: 4 });
rs.setEncoding("utf8");
rs.on("data", function (c) { __chunks += c; });
rs.on("end", function () { __ended = true; });
"started"`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("read stream run: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__chunks`), "hello world")
	jstest.Equal(t, jstest.Str(t, e, `String(__ended)`), "true")

	jstest.Str(t, e, `
var ws = fs.createWriteStream(dir + "/out.txt");
ws.write("ab");
ws.end("cd");
"queued"`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("write stream run: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `fs.readFileSync(dir + "/out.txt", "utf8")`), "abcd")

	jstest.Equal(t, jstest.Str(t, e, `
var out = [];
fs.writeFileSync(dir + "/a.txt", "a");
fs.writeFileSync(dir + "/b.txt", "b");
var d = fs.opendirSync(dir);
var e1 = d.readSync();
if (e1) out.push(e1.isFile());
d.closeSync();
out.length === 1 && out[0] === true`), "true")
}

func TestWatch(t *testing.T) {
	e, _ := setup(t)

	jstest.Str(t, e, `
fs.writeFileSync(dir + "/w.txt", "a");
globalThis.__watch = "pending";
var w = fs.watch(dir + "/w.txt", { interval: 20 });
w.on("change", function () { __watch = "changed"; });
setTimeout(function () { fs.writeFileSync(dir + "/w.txt", "b"); }, 60);
setTimeout(function () { __watch = __watch === "changed" ? __watch : "timeout"; w.close(); }, 400);
"armed"`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("watch run: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `__watch`), "changed")
}

func TestReadValidationMatchesNode(t *testing.T) {
	e, _ := setup(t)

	// An empty destination buffer is ERR_INVALID_ARG_VALUE, not an offset error.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  fs.writeFileSync(dir + "/rv.txt", "xyz");
  var fd = fs.openSync(dir + "/rv.txt", "r");
  try {
    fs.readSync(fd, new Uint8Array(), 0, 10, 0);
  } catch (err) {
    fs.closeSync(fd);
    return err.code + "|" + err.message;
  }
  return "no throw";
})()`), "ERR_INVALID_ARG_VALUE|The argument 'buffer' is empty and cannot be written. Received Uint8Array(0) []")

	// fs.read accepts an options object, including { offset: null }.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  globalThis.__rv = "pending";
  var fd = fs.openSync(dir + "/rv.txt", "r");
  fs.read(fd, { offset: null, buffer: Buffer.alloc(1) }, function (err, bytesRead, buffer) {
    globalThis.__rv = String(err) + ":" + bytesRead + ":" + String.fromCharCode(buffer[0]);
    fs.closeSync(fd);
  });
  return "armed";
})()`), "armed")
	if err := e.RunMain(); err != nil {
		t.Fatalf("read run: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__rv`), "null:1:x")
}

func TestStatOptions(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/stat.txt", "abc");
var st = fs.statSync(dir + "/stat.txt", { bigint: true });
(typeof st.size) + ":" + st.size + ":" + (typeof st.atimeNs) + ":" + st.isFile()`), "bigint:3:bigint:true")

	// atimeNs is BigInt even for the number-valued Stats shape.
	jstest.Equal(t, jstest.Str(t, e, `
var st = fs.statSync(dir + "/stat.txt");
(typeof st.atimeNs) + ":" + (typeof st.size)`), "bigint:number")

	jstest.Equal(t, jstest.Str(t, e, `
String(fs.statSync(dir + "/missing-stat", { throwIfNoEntry: false }))`), "undefined")

	// A caller-provided buffer keeps its old bytes past the read prefix.
	jstest.Equal(t, jstest.Str(t, e, `
var buf = Buffer.alloc(8, 120);
var out = fs.readFileSync(dir + "/stat.txt", { buffer: buf });
out.toString() + ":" + buf.toString() + ":" + buf[3]`), "abc:abcxxxxx:120")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { fs.readFileSync(dir + "/stat.txt", { buffer: Buffer.alloc(1) }); }
  catch (err) { return err.code; }
})()`), "ERR_INVALID_ARG_VALUE")
}

func TestReaddirRecursiveAndRemovalRules(t *testing.T) {
	e, _ := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
fs.mkdirSync(dir + "/r/a/b", { recursive: true });
fs.writeFileSync(dir + "/r/a/b/f.txt", "1");
fs.readdirSync(dir + "/r", { recursive: true }).join(",")`), "a,a/b,a/b/f.txt")

	// unlink(2) and rmdir(2) keep their own errno; rm needs recursive: true.
	jstest.Equal(t, jstest.Str(t, e, `
fs.mkdirSync(dir + "/d1");
(function () { try { fs.unlinkSync(dir + "/d1"); } catch (err) { return err.code; } })()`), "EISDIR")

	jstest.Equal(t, jstest.Str(t, e, `
(function () { try { fs.rmSync(dir + "/d1"); } catch (err) { return err.code; } })()`), "ERR_FS_EISDIR")

	jstest.Equal(t, jstest.Str(t, e, `
fs.writeFileSync(dir + "/plain.txt", "x");
(function () { try { fs.rmdirSync(dir + "/plain.txt"); } catch (err) { return err.code; } })()`), "ENOTDIR")

	jstest.Equal(t, jstest.Str(t, e, `
(function () { try { fs.mkdirSync(dir + "/plain.txt", { recursive: true }); } catch (err) { return err.code; } })()`), "EEXIST")

	// readdir's `buffer` encoding returns the names as Buffers.
	jstest.Equal(t, jstest.Str(t, e, `
var names = fs.readdirSync(dir + "/r", { encoding: "buffer" });
Buffer.isBuffer(names[0]) + ":" + names[0].toString()`), "true:a")
}

func TestFileContentsOnDisk(t *testing.T) {
	e, dir := setup(t)

	jstest.Str(t, e, `fs.writeFileSync(dir + "/on-disk.txt", "disk")`)
	data, err := os.ReadFile(filepath.Join(dir, "on-disk.txt"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "disk" {
		t.Fatalf("file contents = %q, want %q", data, "disk")
	}
}
