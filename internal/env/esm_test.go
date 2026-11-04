package env_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"godejs/internal/env"
)

// runModule writes files into a temporary project, runs entry, and returns the
// string the module assigned to globalThis.__out.
func runModule(t *testing.T, entry string, files map[string]string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	e := env.New(env.Options{
		Args:   []string{"gode", entry},
		Cwd:    dir,
		Env:    []string{"GODE_TEST=1"},
		Stdout: &bytes.Buffer{},
		Stderr: io.Discard,
		Stdin:  strings.NewReader(""),
	})
	t.Cleanup(func() { e.JS.Close() })
	if _, err := e.RunFile(filepath.Join(dir, entry)); err != nil {
		return "", err
	}
	v := e.Global().Get("__out")
	if !v.Valid() {
		t.Fatal("module did not set globalThis.__out")
	}
	return v.Str(), nil
}

func mustRun(t *testing.T, entry string, files map[string]string) string {
	t.Helper()
	got, err := runModule(t, entry, files)
	if err != nil {
		t.Fatalf("run %s: %v", entry, err)
	}
	return got
}

func TestESMDefaultNamedNamespaceImports(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"dep.mjs": `export default function add(a, b) { return a + b; }
export const tag = 'dep';
export function named() { return 'named'; }
`,
		"main.mjs": `import add, {tag, named as n} from './dep.mjs';
import * as all from './dep.mjs';
globalThis.__out = [add(1, 2), tag, n(), Object.keys(all).sort().join(',')].join('|');
`,
	})
	if got != "3|dep|named|default,named,tag" {
		t.Fatalf("got %q", got)
	}
}

func TestESMDefaultFunctionNameStaysVisible(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"util.mjs": `export default function ui() { return 'ui'; }
export {ui};
`,
		"main.mjs": `import ui, {ui as named} from './util.mjs';
globalThis.__out = ui() + '|' + named();
`,
	})
	if got != "ui|ui" {
		t.Fatalf("got %q", got)
	}
}

func TestESMReExports(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"a.mjs": `export const x = 1;
export default 2;
`,
		"b.mjs": `export * as ns from './a.mjs';
export {default as d} from './a.mjs';
`,
		"main.mjs": `import {ns, d} from './b.mjs';
globalThis.__out = [ns.x, ns.default, d].join('|');
`,
	})
	if got != "1|2|2" {
		t.Fatalf("got %q", got)
	}
}

func TestESMImportMeta(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"dep.mjs": `export const value = 'dep';`,
		"main.mjs": `const resolved = import.meta.resolve('./dep.mjs');
globalThis.__out = [
  import.meta.url.startsWith('file://'),
  import.meta.filename.endsWith('main.mjs'),
  import.meta.dirname.length > 0,
  resolved.endsWith('dep.mjs'),
].join('|');
`,
	})
	if got != "true|true|true|true" {
		t.Fatalf("got %q", got)
	}
}

func TestESMLeavesLiteralsAndCommentsAlone(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"tricky.mjs": "const s = \"import x from 'y'\";\n" +
			"const tpl = `export default ${1 + 1}`;\n" +
			"const re = /import\\s+foo/;\n" +
			"// export const commented = 1;\n" +
			"/* import nope from 'nope' */\n" +
			"export {s, tpl, re};\n",
		"main.mjs": `import {s, tpl, re} from './tricky.mjs';
globalThis.__out = [s, tpl, String(re)].join('|');
`,
	})
	if got != "import x from 'y'|export default 2|/import\\s+foo/" {
		t.Fatalf("got %q", got)
	}
}

func TestPackageExportsSubpathsAndWildcards(t *testing.T) {
	got := mustRun(t, "main.js", map[string]string{
		"node_modules/widget/package.json": `{"name":"widget","exports":{
  ".":{"require":"./cjs.cjs","default":"./cjs.cjs"},
  "./feature":"./feature.js",
  "./lib/*":"./lib/*.js"
}}`,
		"node_modules/widget/cjs.cjs":     `module.exports = {kind: 'root'};`,
		"node_modules/widget/feature.js":  `module.exports = {kind: 'feature'};`,
		"node_modules/widget/lib/tool.js": `module.exports = {kind: 'tool'};`,
		"main.js": `const root = require('widget');
const feature = require('widget/feature');
const tool = require('widget/lib/tool');
globalThis.__out = [root.kind, feature.kind, tool.kind].join('|');
`,
	})
	if got != "root|feature|tool" {
		t.Fatalf("got %q", got)
	}
}

func TestPackageImportsMap(t *testing.T) {
	got := mustRun(t, "pkg/main.js", map[string]string{
		"pkg/package.json": `{"name":"pkg","imports":{"#config":"./config.js"}}`,
		"pkg/config.js":    `module.exports = 'configured';`,
		"pkg/main.js":      `globalThis.__out = require('#config');`,
	})
	if got != "configured" {
		t.Fatalf("got %q", got)
	}
}

func TestESMImportsFromPackageExports(t *testing.T) {
	got := mustRun(t, "main.mjs", map[string]string{
		"node_modules/esm-only/package.json": `{"name":"esm-only","type":"module","exports":"./index.mjs"}`,
		"node_modules/esm-only/index.mjs":    `export const value = 'esm';`,
		"main.mjs": `import {value} from 'esm-only';
globalThis.__out = value;
`,
	})
	if got != "esm" {
		t.Fatalf("got %q", got)
	}
}
