package all_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"godejs/internal/all"
	"godejs/internal/jstest"
)

// TestEveryClaimedKeyExists resolves every key a module advertises in
// Implemented() against the real installed object graph. The parity report
// counts names; this test proves the names are actually reachable, so a typo
// or a prelude that fails to install cannot inflate the score.
func TestEveryClaimedKeyExists(t *testing.T) {
	e := jstest.New(t)
	all.InstallGlobals(e)
	installed := map[string]bool{}
	for _, info := range all.Modules() {
		e.Global().Set(info.Name, info.Install(e).V())
		installed[info.Name] = true
	}

	baselineFiles, err := filepath.Glob(filepath.Join("..", "..", "parity", "baseline", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	tracked := map[string]map[string]bool{}
	for _, file := range baselineFiles {
		mod := strings.TrimSuffix(filepath.Base(file), ".json")
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var keys []string
		if err := json.Unmarshal(data, &keys); err != nil {
			t.Fatal(err)
		}
		set := map[string]bool{}
		for _, key := range keys {
			set[key] = true
		}
		tracked[mod] = set
	}

	total, bad := 0, 0
	for _, info := range all.Modules() {
		set, ok := tracked[info.Name]
		if !ok {
			continue
		}
		for _, key := range info.Implemented() {
			if !set[key] {
				t.Errorf("%s: claims untracked key %q", info.Name, key)
				bad++
				continue
			}
			// Docs spell instance properties as Module.Class.prop, which no
			// static walk can distinguish from a class static; only "#" keys
			// name a member that must live on the prototype.
			if !strings.Contains(key, "#") {
				continue
			}
			total++
			expr := resolveExpr(key)
			got, err := e.JS.Eval(expr, "apicheck.js")
			if err != nil {
				t.Errorf("%s: eval %q: %v", key, expr, err)
				bad++
				continue
			}
			switch v := got.Str(); {
			case v == "absent":
				t.Errorf("%s: claimed but not reachable at runtime", key)
				bad++
			case v == "undefined":
				t.Errorf("%s: claimed but resolves to undefined", key)
				bad++
			case strings.HasPrefix(v, "throw:"):
				t.Errorf("%s: resolution threw %s", key, strings.TrimPrefix(v, "throw:"))
				bad++
			case v != "function":
				t.Errorf("%s: instance member is %s, want function", key, v)
				bad++
			}
		}
	}
	t.Logf("resolved %d prototype keys, %d problems, %d modules installed", total, bad, len(installed))
}

// resolveExpr builds JavaScript that walks the key's owner path from the
// module export and reports whether the final member is present. Dot keys can
// name either a class static or an instance property (the docs-derived scheme
// spells both the same way), so both holders are tried; "#" keys must resolve
// on the prototype.
func resolveExpr(key string) string {
	owner, member, instance := key, "", false
	if i := strings.Index(key, "#"); i >= 0 {
		owner, member, instance = key[:i], key[i+1:], true
	}
	head, rest, _ := strings.Cut(owner, ".")
	segs := strings.Split(rest, ".")
	if !instance {
		member = segs[len(segs)-1]
		segs = segs[:len(segs)-1]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "(function(){try{var o=globalThis[%q];", head)
	for _, seg := range segs {
		fmt.Fprintf(&b, "if(o==null)return 'absent';o=o[%q];", seg)
	}
	b.WriteString("if(o==null)return 'absent';")
	if strings.HasPrefix(member, "[Symbol.") && strings.HasSuffix(member, "]") {
		name := strings.TrimSuffix(strings.TrimPrefix(member, "[Symbol."), "]")
		fmt.Fprintf(&b, "var m=Symbol[%q];", name)
	} else {
		fmt.Fprintf(&b, "var m=%q;", member)
	}
	// The docs JSON categorizes some class statics (URL.parse, MIMEType.parse)
	// as instance methods, so "#" tries the prototype and then the constructor.
	b.WriteString("if(!(m in o)&&typeof o==='function'&&o.prototype!=null&&(m in o.prototype))o=o.prototype;")
	b.WriteString("if(!(m in o))return 'absent';return typeof o[m];")
	b.WriteString("}catch(err){return 'throw:'+err.message}})()")
	return b.String()
}
