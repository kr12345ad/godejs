package main

import "testing"

func TestFormatForInputWithSourceDetectsEvalESM(t *testing.T) {
	tests := []struct {
		name      string
		inputType string
		file      string
		source    string
		want      string
	}{
		{"static import", "", "[eval]", `import os from "node:os";`, "module"},
		{"named import", "", "[eval]", `import { readFile } from "node:fs";`, "module"},
		{"export", "", "[eval]", `export default 1;`, "module"},
		{"import meta", "", "[eval]", `console.log(import.meta.url);`, "module"},
		{"dynamic import", "", "[eval]", `import("node:fs");`, "commonjs"},
		{"import in string", "", "[eval]", `const source = "import os from 'node:os'";`, "commonjs"},
		{"plain script", "", "[eval]", `console.log("ok");`, "commonjs"},
		{"explicit commonjs", "commonjs", "[eval]", `import os from "node:os";`, "commonjs"},
		{"explicit module", "module", "[eval]", `console.log("ok");`, "module"},
		{"mjs file", "", "app.mjs", `console.log("ok");`, "module"},
		{"stdin import", "", "[stdin]", `import os from "node:os";`, "module"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatForInputWithSource(tt.inputType, tt.file, tt.source); got != tt.want {
				t.Fatalf("formatForInputWithSource(%q, %q, %q) = %q, want %q",
					tt.inputType, tt.file, tt.source, got, tt.want)
			}
		})
	}
}
