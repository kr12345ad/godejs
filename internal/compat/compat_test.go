package compat

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "b", "main.js"), "")
	writeFile(t, filepath.Join(root, "a", "main.mjs"), "")
	writeFile(t, filepath.Join(root, "a", "main.js"), "")
	writeFile(t, filepath.Join(root, "skip", "other.js"), "")

	cases, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 || cases[0].Name != "a" || cases[1].Name != "b" {
		t.Fatalf("cases = %+v", cases)
	}
	if cases[0].Entry != "main.mjs" {
		t.Fatalf("entry = %q, want main.mjs", cases[0].Entry)
	}
}

func TestRunComparesStdoutAndExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake executables are shell scripts")
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "same", "main.js"), "")
	writeFile(t, filepath.Join(root, "different", "main.js"), "")
	writeFile(t, filepath.Join(root, "status", "main.js"), "")

	node := fakeBin(t, "node", `case "$PWD" in
  */different) printf 'node\n' ;;
  */status) exit 7 ;;
  *) printf 'ok\n' ;;
esac`)
	gode := fakeBin(t, "gode", `case "$PWD" in
  */different) printf 'gode\n' ;;
  */status) exit 7 ;;
  *) printf 'ok\n' ;;
esac`)

	results, err := Run(root, node, gode)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Result{}
	for _, r := range results {
		byName[r.Name] = r
	}
	if !byName["same"].Passed {
		t.Fatalf("same should pass: %+v", byName["same"])
	}
	if !byName["status"].Passed {
		t.Fatalf("matching non-zero exits should pass: %+v", byName["status"])
	}
	diff := byName["different"]
	if diff.Passed {
		t.Fatalf("different stdout should fail: %+v", diff)
	}
	if !strings.Contains(diff.Err, "stdout differs") || !strings.Contains(diff.Err, "first difference") {
		t.Fatalf("error does not describe the difference: %q", diff.Err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fakeBin(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
