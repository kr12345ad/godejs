// Package path implements Node.js's path module on top of the GodeJS runtime.
//
// The POSIX semantics are the default; win32 semantics are provided by
// path.win32 (and path.posix always refers to the POSIX implementation).
package path

import (
	"fmt"
	"strconv"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("path", Install)
	env.Register("path/posix", func(e *env.Env) js.Object { return e.Exports("path").Get("posix").Object() })
	env.Register("path/win32", func(e *env.Env) js.Object { return e.Exports("path").Get("win32").Object() })
}

// Install builds the path module's exports object.
func Install(e *env.Env) js.Object {
	posix := e.JS.Obj()
	win := e.JS.Obj()
	buildPosix(e, posix)
	buildWin32(e, win)

	exports := e.JS.Obj()
	for _, key := range []string{
		"basename", "delimiter", "dirname", "extname", "format",
		"isAbsolute", "join", "matchesGlob", "normalize", "parse",
		"relative", "resolve", "sep", "toNamespacedPath",
	} {
		exports.Set(key, posix.Get(key))
	}
	exports.SetObj("posix", posix)
	exports.SetObj("win32", win)
	return exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"path.basename",
		"path.delimiter",
		"path.dirname",
		"path.extname",
		"path.format",
		"path.isAbsolute",
		"path.join",
		"path.matchesGlob",
		"path.normalize",
		"path.parse",
		"path.posix",
		"path.relative",
		"path.resolve",
		"path.sep",
		"path.toNamespacedPath",
		"path.win32",
	}
}

// currentCwd prefers process.cwd() (which process.chdir updates) and falls
// back to the environment's initial working directory.
func currentCwd(e *env.Env) string {
	proc := e.Global().Get("process")
	if proc.Valid() && proc.IsObject() {
		fn := proc.Object().Get("cwd")
		if fn.Valid() && fn.IsFunction() {
			if v, err := fn.Object().Call(js.Object{}); err == nil && v.IsString() {
				return v.Str()
			}
		}
	}
	return e.Cwd()
}

func buildPosix(e *env.Env, o js.Object) {
	o.SetStr("sep", "/")
	o.SetStr("delimiter", ":")

	o.SetFunc("basename", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		ext, err := optString(args, 1, "ext")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixBasename(p, ext)), nil
	})

	o.SetFunc("dirname", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixDirname(p)), nil
	})

	o.SetFunc("extname", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixExtname(p)), nil
	})

	o.SetFunc("format", func(this js.Object, args []js.Value) (js.Value, error) {
		obj, err := reqObject(args, 0, "pathObject")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(formatPath(obj, "/")), nil
	})

	o.SetFunc("isAbsolute", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Bool(posixIsAbsolute(p)), nil
	})

	o.SetFunc("join", func(this js.Object, args []js.Value) (js.Value, error) {
		parts, err := allStrings(args, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixJoin(parts)), nil
	})

	o.SetFunc("matchesGlob", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		pattern, err := reqString(args, 1, "globPattern")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Bool(globMatch(p, pattern, '/', false)), nil
	})

	o.SetFunc("normalize", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixNormalize(p)), nil
	})

	o.SetFunc("parse", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		root, dir, base, ext, name := posixParse(p)
		obj := e.JS.Obj()
		obj.SetStr("root", root)
		obj.SetStr("dir", dir)
		obj.SetStr("base", base)
		obj.SetStr("ext", ext)
		obj.SetStr("name", name)
		return obj.V(), nil
	})

	o.SetFunc("relative", func(this js.Object, args []js.Value) (js.Value, error) {
		from, err := reqString(args, 0, "from")
		if err != nil {
			return js.Value{}, err
		}
		to, err := reqString(args, 1, "to")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixRelative(currentCwd(e), from, to)), nil
	})

	o.SetFunc("resolve", func(this js.Object, args []js.Value) (js.Value, error) {
		parts, err := allStrings(args, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(posixResolve(currentCwd(e), parts)), nil
	})

	o.SetFunc("toNamespacedPath", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		// Node's POSIX implementation returns the input unchanged, including
		// non-strings.
		if !args[0].IsString() {
			return args[0], nil
		}
		return args[0], nil
	})
}

func buildWin32(e *env.Env, o js.Object) {
	o.SetStr("sep", `\`)
	o.SetStr("delimiter", ";")

	o.SetFunc("basename", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		ext, err := optString(args, 1, "ext")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winBasename(p, ext)), nil
	})

	o.SetFunc("dirname", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winDirname(p)), nil
	})

	o.SetFunc("extname", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winExtname(p)), nil
	})

	o.SetFunc("format", func(this js.Object, args []js.Value) (js.Value, error) {
		obj, err := reqObject(args, 0, "pathObject")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(formatPath(obj, `\`)), nil
	})

	o.SetFunc("isAbsolute", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Bool(winIsAbsolute(p)), nil
	})

	o.SetFunc("join", func(this js.Object, args []js.Value) (js.Value, error) {
		parts, err := allStrings(args, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winJoin(parts)), nil
	})

	o.SetFunc("matchesGlob", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		pattern, err := reqString(args, 1, "globPattern")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Bool(globMatch(p, pattern, '/', true)), nil
	})

	o.SetFunc("normalize", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winNormalize(p)), nil
	})

	o.SetFunc("parse", func(this js.Object, args []js.Value) (js.Value, error) {
		p, err := reqString(args, 0, "path")
		if err != nil {
			return js.Value{}, err
		}
		root, dir, base, ext, name := winParse(p)
		obj := e.JS.Obj()
		obj.SetStr("root", root)
		obj.SetStr("dir", dir)
		obj.SetStr("base", base)
		obj.SetStr("ext", ext)
		obj.SetStr("name", name)
		return obj.V(), nil
	})

	o.SetFunc("relative", func(this js.Object, args []js.Value) (js.Value, error) {
		from, err := reqString(args, 0, "from")
		if err != nil {
			return js.Value{}, err
		}
		to, err := reqString(args, 1, "to")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winRelative(currentCwd(e), from, to)), nil
	})

	o.SetFunc("resolve", func(this js.Object, args []js.Value) (js.Value, error) {
		parts, err := allStrings(args, "path")
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(winResolve(currentCwd(e), parts)), nil
	})

	o.SetFunc("toNamespacedPath", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		if !args[0].IsString() {
			return args[0], nil
		}
		return e.JS.Str(winToNamespacedPath(currentCwd(e), args[0].Str())), nil
	})
}

func formatPath(obj js.Object, sep string) string {
	get := func(key string) string {
		v := obj.Get(key)
		if v.IsString() {
			return v.Str()
		}
		return ""
	}
	dir := get("dir")
	if dir == "" {
		dir = get("root")
	}
	root := get("root")
	base := get("base")
	if base == "" {
		ext := get("ext")
		// Node's formatExt adds the dot when the ext property omits it.
		if ext != "" && ext[0] != '.' {
			ext = "." + ext
		}
		base = get("name") + ext
	}
	if dir == "" {
		return base
	}
	if dir == root {
		return dir + base
	}
	return dir + sep + base
}

func reqString(args []js.Value, i int, name string) (string, error) {
	if i >= len(args) {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			fmt.Sprintf("The %q argument must be of type string. Received undefined", name))
	}
	v := args[i]
	if v.IsString() {
		return v.Str(), nil
	}
	return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		fmt.Sprintf("The %q argument must be of type string. Received %s", name, describe(v)))
}

func optString(args []js.Value, i int, name string) (*string, error) {
	if i >= len(args) || args[i].IsUndefined() {
		return nil, nil
	}
	v := args[i]
	if !v.IsString() {
		return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			fmt.Sprintf("The %q argument must be of type string. Received %s", name, describe(v)))
	}
	s := v.Str()
	return &s, nil
}

func allStrings(args []js.Value, name string) ([]string, error) {
	out := make([]string, 0, len(args))
	for _, v := range args {
		if !v.IsString() {
			return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The %q argument must be of type string. Received %s", name, describe(v)))
		}
		out = append(out, v.Str())
	}
	return out, nil
}

func reqObject(args []js.Value, i int, name string) (js.Object, error) {
	if i >= len(args) {
		return js.Object{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			fmt.Sprintf("The %q argument must be of type object. Received undefined", name))
	}
	v := args[i]
	if !v.IsObject() || v.IsNull() || v.IsArray() {
		return js.Object{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			fmt.Sprintf("The %q argument must be of type object. Received %s", name, describe(v)))
	}
	return v.Object(), nil
}

func describe(v js.Value) string {
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return fmt.Sprintf("type boolean (%t)", v.Bool())
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsArray():
		return "an instance of Array"
	case v.IsObject():
		tag := v.Tag()
		if tag == "" {
			tag = "Object"
		}
		return "an instance of " + tag
	default:
		return "type " + strings.ToLower(v.Tag())
	}
}
