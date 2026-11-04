package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// packageJSON is the subset of package.json that module resolution uses.
type packageJSON struct {
	Name    string          `json:"name"`
	Main    string          `json:"main"`
	Type    string          `json:"type"`
	Exports json.RawMessage `json:"exports"`
	Imports json.RawMessage `json:"imports"`
}

// readPackageJSON reads dir/package.json.
func readPackageJSON(dir string) (*packageJSON, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, false
	}
	var p packageJSON
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, false
	}
	return &p, true
}

// moduleFormat reports how a file should be evaluated: "esm", "cjs" or
// "json". A .js file follows the nearest package.json "type" field.
func moduleFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mjs":
		return "esm"
	case ".cjs":
		return "cjs"
	case ".json":
		return "json"
	}
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if p, ok := readPackageJSON(dir); ok {
			if strings.EqualFold(p.Type, "module") {
				return "esm"
			}
			return "cjs"
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "cjs"
		}
	}
}

// resolvePackage resolves a bare module specifier to a file. It supports
// package.json "exports" (string, conditional object, subpath map and the
// common "./*" wildcard), "main", and directory index files.
func resolvePackage(dir, spec string) (string, error) {
	name, sub := splitPackageSpec(spec)
	if name == "" {
		return "", fmt.Errorf("invalid module specifier: %s", spec)
	}
	for base := dir; ; {
		root := filepath.Join(base, "node_modules", name)
		if st, err := os.Stat(root); err == nil && st.IsDir() {
			if resolved, ok := resolvePackageRoot(root, sub); ok {
				return resolved, nil
			}
			return "", fmt.Errorf("not found: %s", spec)
		}
		if sub == "" {
			if resolved, ok := resolveAsFile(root); ok {
				return resolved, nil
			}
		}
		parent := filepath.Dir(base)
		if parent == base {
			break
		}
		base = parent
	}
	return "", fmt.Errorf("not found: %s", spec)
}

// splitPackageSpec separates "pkg" / "@scope/pkg/sub/path" into the package
// name and its subpath (without a leading slash).
func splitPackageSpec(spec string) (name, sub string) {
	parts := strings.Split(spec, "/")
	if len(parts) == 0 {
		return "", ""
	}
	n := 1
	if strings.HasPrefix(parts[0], "@") && len(parts) > 1 {
		n = 2
	}
	name = strings.Join(parts[:n], "/")
	if len(parts) > n {
		sub = strings.Join(parts[n:], "/")
	}
	return name, sub
}

// resolvePackageRoot resolves the package entry or a subpath inside root.
func resolvePackageRoot(root, sub string) (string, bool) {
	pkg, _ := readPackageJSON(root)
	if pkg != nil && len(pkg.Exports) > 0 {
		if target, ok := exportsTarget(pkg.Exports, sub); ok {
			path := filepath.Join(root, filepath.FromSlash(target))
			if resolved, ok := resolveAsFile(path); ok {
				return resolved, true
			}
			return "", false
		}
	}
	if sub == "" {
		if pkg != nil && pkg.Main != "" {
			if resolved, ok := resolveAsFile(filepath.Join(root, pkg.Main)); ok {
				return resolved, true
			}
			if resolved, ok := resolveAsFile(filepath.Join(root, pkg.Main, "index.js")); ok {
				return resolved, true
			}
		}
		// No "main": legacy CommonJS directory resolution falls back to
		// index.js/index.json/index.node, not index.js alone. Packages such as
		// spdx-license-ids ship only index.json.
		return resolveAsFile(root)
	}
	return resolveAsFile(filepath.Join(root, filepath.FromSlash(sub)))
}

// exportsTarget walks the "exports" value for a subpath. sub is "" for the
// package's "." entry.
func exportsTarget(raw json.RawMessage, sub string) (string, bool) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", false
	}
	key := "."
	if sub != "" {
		key = "./" + sub
	}
	return exportsLookup(v, key, sub)
}

func exportsLookup(v any, key, sub string) (string, bool) {
	switch node := v.(type) {
	case string:
		if key == "." {
			return node, true
		}
		return "", false
	case []any:
		for _, item := range node {
			if s, ok := exportsLookup(item, key, sub); ok {
				return s, true
			}
		}
	case map[string]any:
		// A conditions object (no keys beginning with ".") applies to ".".
		isSubpathMap := false
		for k := range node {
			if strings.HasPrefix(k, ".") {
				isSubpathMap = true
				break
			}
		}
		if !isSubpathMap {
			if key != "." {
				return "", false
			}
			return conditionTarget(node)
		}
		if target, ok := node[key]; ok {
			return exportsLookup(target, ".", "")
		}
		// Wildcard patterns, most specific first.
		var bestKey string
		var bestTarget any
		for pattern, target := range node {
			star := strings.IndexByte(pattern, '*')
			if star < 0 {
				continue
			}
			prefix, suffix := pattern[:star], pattern[star+1:]
			if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix) && len(key) >= len(prefix)+len(suffix) {
				if len(pattern) > len(bestKey) {
					bestKey, bestTarget = pattern, target
				}
			}
		}
		if bestTarget != nil {
			star := strings.IndexByte(bestKey, '*')
			prefix, suffix := bestKey[:star], bestKey[star+1:]
			replacement := key[len(prefix) : len(key)-len(suffix)]
			return replaceTarget(bestTarget, replacement)
		}
	}
	return "", false
}

// conditionTarget picks a condition branch for require(). "require" is
// preferred, then "node", then "default"; "import" is a last resort so that
// packages shipping only ESM can still be loaded by the transform.
func conditionTarget(node map[string]any) (string, bool) {
	for _, name := range []string{"require", "node", "default", "import"} {
		if target, ok := node[name]; ok {
			if s, ok := exportsLookup(target, ".", ""); ok {
				return s, true
			}
		}
	}
	return "", false
}

func replaceTarget(target any, replacement string) (string, bool) {
	switch node := target.(type) {
	case string:
		return strings.ReplaceAll(node, "*", replacement), true
	case []any:
		for _, item := range node {
			if s, ok := replaceTarget(item, replacement); ok {
				return s, true
			}
		}
	case map[string]any:
		return conditionTarget(node)
	}
	return "", false
}

// resolveImport resolves a package-internal "#name" specifier through the
// nearest package.json "imports" map, as used by packages such as chalk.
func resolveImport(dir, spec string) (string, error) {
	root := dir
	var pkg *packageJSON
	for {
		if p, ok := readPackageJSON(root); ok {
			pkg = p
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("not found: %s", spec)
		}
		root = parent
	}
	if len(pkg.Imports) == 0 {
		return "", fmt.Errorf("not found: %s", spec)
	}
	var v any
	if err := json.Unmarshal(pkg.Imports, &v); err != nil {
		return "", fmt.Errorf("not found: %s", spec)
	}
	target, ok := importsLookup(v, spec)
	if !ok {
		return "", fmt.Errorf("not found: %s", spec)
	}
	if strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../") {
		if resolved, ok := resolveAsFile(filepath.Join(root, filepath.FromSlash(target))); ok {
			return resolved, nil
		}
		return "", fmt.Errorf("not found: %s", spec)
	}
	// An imports target may itself be a package specifier.
	return resolvePackage(root, target)
}

func importsLookup(v any, spec string) (string, bool) {
	switch node := v.(type) {
	case string:
		return node, true
	case []any:
		for _, item := range node {
			if s, ok := importsLookup(item, spec); ok {
				return s, true
			}
		}
	case map[string]any:
		if target, ok := node[spec]; ok {
			return conditionTargetAny(target)
		}
		var bestKey string
		var bestTarget any
		for pattern, target := range node {
			star := strings.IndexByte(pattern, '*')
			if star < 0 {
				continue
			}
			prefix, suffix := pattern[:star], pattern[star+1:]
			if strings.HasPrefix(spec, prefix) && strings.HasSuffix(spec, suffix) && len(spec) >= len(prefix)+len(suffix) {
				if len(pattern) > len(bestKey) {
					bestKey, bestTarget = pattern, target
				}
			}
		}
		if bestTarget != nil {
			star := strings.IndexByte(bestKey, '*')
			replacement := spec[star:]
			if suffix := bestKey[star+1:]; suffix != "" {
				replacement = spec[star : len(spec)-len(suffix)]
			}
			return replaceTarget(bestTarget, replacement)
		}
	}
	return "", false
}

func conditionTargetAny(v any) (string, bool) {
	switch node := v.(type) {
	case string:
		return node, true
	case map[string]any:
		return conditionTarget(node)
	case []any:
		for _, item := range node {
			if s, ok := conditionTargetAny(item); ok {
				return s, true
			}
		}
	}
	return "", false
}

// resolveAsFile resolves a path that may be a file, a file plus a known
// extension, or a directory with package.json/index.js.
func resolveAsFile(base string) (string, bool) {
	candidates := []string{base}
	// Node appends each extension after trying the exact name, even when the
	// basename contains a dot. `object-inspect` relies on this: it requires
	// `./util.inspect`, which resolves to `util.inspect.js`.
	candidates = append(candidates, base+".js", base+".json", base+".cjs")
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs, true
			}
		}
	}
	if st, err := os.Stat(base); err == nil && st.IsDir() {
		if pkg, ok := readPackageJSON(base); ok {
			if pkg.Main != "" {
				if abs, ok := resolveAsFile(filepath.Join(base, pkg.Main)); ok {
					return abs, true
				}
			}
		}
		for _, name := range []string{"index.js", "index.json", "index.cjs"} {
			if abs, ok := resolveAsFile(filepath.Join(base, name)); ok {
				return abs, true
			}
		}
	}
	return "", false
}
