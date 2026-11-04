// Command paritygen writes the tracked Node.js 24 LTS API baseline from the
// official documentation JSON.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var moduleNames = []string{
	"assert", "buffer", "child_process", "console", "crypto", "events", "fs",
	"module", "os", "path", "process", "querystring", "stream",
	"string_decoder", "timers", "tty", "url", "util", "v8", "vm", "zlib",
}

type property struct {
	Name    string
	TextRaw string `json:"textRaw"`
}
type method struct {
	Name string
}
type class struct {
	Name         string
	Methods      []method
	ClassMethods []method `json:"classMethods"`
	Properties   []property
}
type moduleDoc struct {
	Name        string
	DisplayName string `json:"displayName"`
	Methods     []method
	Properties  []property
	Classes     []class
	Modules     []moduleDoc
}
type globalDoc struct {
	Name       string
	Methods    []method
	Properties []property
}
type doc struct {
	Modules []moduleDoc
	Globals []globalDoc
}

// extras are module exports that the per-module docs JSON does not list (they
// live in global or other pages).
var extras = map[string][]string{
	"timers": {
		"timers.setTimeout", "timers.clearTimeout",
		"timers.setInterval", "timers.clearInterval",
		"timers.setImmediate", "timers.clearImmediate",
		"timers.scheduler",
	},
	"buffer": {
		"buffer.atob", "buffer.btoa", "buffer.constants", "buffer.kMaxLength",
		"buffer.isAscii", "buffer.isUtf8", "buffer.transcode",
		"buffer.SlowBuffer", "buffer.INSPECT_MAX_BYTES",
	},
	"console": {"console.Console"},
}

func main() {
	out := flag.String("out", "parity/baseline", "output directory")
	base := flag.String("docs", "https://nodejs.org/docs/latest-v24.x/api/", "Node docs JSON base URL")
	flag.Parse()
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fatal(err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	for _, name := range moduleNames {
		keys, err := fetch(client, *base+name+".json", name)
		if err != nil {
			fatal(err)
		}
		data, err := json.MarshalIndent(keys, "", "  ")
		if err != nil {
			fatal(err)
		}
		path := filepath.Join(*out, name+".json")
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("%-16s %3d tracked APIs\n", name, len(keys))
	}
}

func fetch(client *http.Client, url, name string) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	var d doc
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	set := map[string]bool{}
	add := func(k string) {
		k = strings.TrimSpace(k)
		if k != "" && !strings.HasSuffix(k, ".") {
			set[k] = true
		}
	}
	var walk func(m moduleDoc, prefix string)
	walk = func(m moduleDoc, prefix string) {
		for _, x := range m.Methods {
			add(prefix + "." + x.Name)
		}
		for _, p := range m.Properties {
			add(prefix + "." + propertyName(p))
		}
		for _, c := range m.Classes {
			cn := strings.TrimPrefix(c.Name, name+".")
			if name == "console" {
				// require('console') is an instance: its methods are the
				// module's own methods.
				for _, x := range c.Methods {
					add("console." + x.Name)
				}
				continue
			}
			add(prefix + "." + cn)
			for _, x := range c.Methods {
				add(prefix + "." + cn + "#" + x.Name)
			}
			for _, x := range c.ClassMethods {
				add(prefix + "." + cn + "." + x.Name)
			}
			for _, p := range c.Properties {
				add(prefix + "." + cn + "." + propertyName(p))
			}
		}
		for _, child := range m.Modules {
			childPrefix := prefix
			if ns := apiNamespace(child.Name); ns != "" {
				childPrefix = prefix + "." + ns
			}
			walk(child, childPrefix)
		}
	}
	for _, m := range d.Modules {
		walk(m, name)
	}
	if name == "process" {
		for _, g := range d.Globals {
			for _, x := range g.Methods {
				add("process." + x.Name)
			}
			for _, p := range g.Properties {
				add("process." + p.Name)
			}
		}
	}
	for _, k := range extras[name] {
		add(k)
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// apiNamespace returns the export namespace of a docs section. Only the
// promises API is a real sub-object of the module exports; callback and
// synchronous sections document the module's own functions, and prose
// sections only group classes that live on the module itself.
func apiNamespace(section string) string {
	switch section {
	case "promises_api":
		return "promises"
	}
	return ""
}

// propertyName recovers the documented name for properties whose docs entry
// is typed (for example path's `delimiter`, `sep`, `posix` and `win32`).
func propertyName(p property) string {
	if p.Name != "Type" {
		return p.Name
	}
	if i := strings.Index(p.TextRaw, "`"); i >= 0 {
		rest := p.TextRaw[i+1:]
		if j := strings.Index(rest, "`"); j > 0 {
			return rest[:j]
		}
	}
	return ""
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "paritygen:", err)
	os.Exit(1)
}
