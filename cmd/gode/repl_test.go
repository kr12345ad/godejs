package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/version"
)

func runReplSession(t *testing.T, input string) (string, string) {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	e := env.New(env.Options{
		Args:   []string{"gode"},
		Cwd:    dir,
		Stdin:  strings.NewReader(input),
		Stdout: &stdout,
		Stderr: &stderr,
	})
	defer e.JS.Close()
	e.InstallAll()
	if err := repl(e); err != nil {
		t.Fatalf("repl: %v", err)
	}
	return stdout.String(), stderr.String()
}

func TestReplWelcomeShowsGodeVersion(t *testing.T) {
	stdout, _ := runReplSession(t, ".exit\n")
	want := "Welcome to gode v" + version.Gode + "!"
	if !strings.Contains(stdout, want) {
		t.Fatalf("welcome message %q not found:\n%s", want, stdout)
	}
}

func TestReplHelpListsEveryCommand(t *testing.T) {
	stdout, _ := runReplSession(t, ".help\n.exit\n")
	for _, command := range []string{".break", ".clear", ".editor", ".exit", ".help", ".load", ".save"} {
		if !strings.Contains(stdout, command) {
			t.Errorf("help output does not mention %s:\n%s", command, stdout)
		}
	}
}

func TestReplBreakAndClearAbandonMultilineInput(t *testing.T) {
	for _, command := range []string{".break", ".clear"} {
		t.Run(command, func(t *testing.T) {
			stdout, stderr := runReplSession(t, "function broken() {\n"+command+"\n1 + 1\n.exit\n")
			if strings.Contains(stderr, "SyntaxError") {
				t.Fatalf("buffer was not cleared:\n%s", stderr)
			}
			if !strings.Contains(stdout, "2") {
				t.Fatalf("expected the next expression to run:\n%s", stdout)
			}
		})
	}
}

func TestReplEditorEvaluatesCollectedSource(t *testing.T) {
	stdout, stderr := runReplSession(t, ".editor\nconst editorValue = 41\neditorValue + 1\n.\neditorValue\n.exit\n")
	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "42") || !strings.Contains(stdout, "41") {
		t.Fatalf("editor source did not evaluate and retain editorValue:\n%s", stdout)
	}
}

func TestReplLoadEvaluatesFileInSession(t *testing.T) {
	file := filepath.Join(t.TempDir(), "loaded.js")
	if err := os.WriteFile(file, []byte("const loadedValue = 19;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := runReplSession(t, ".load "+file+"\nloadedValue\n.exit\n")
	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "19") {
		t.Fatalf("expected loadedValue to persist:\n%s", stdout)
	}
}

func TestReplSaveWritesEvaluatedCommands(t *testing.T) {
	file := filepath.Join(t.TempDir(), "session.js")
	stdout, stderr := runReplSession(t, "const savedValue = 7\n.save "+file+"\n.exit\n")
	if stderr != "" {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "Saved REPL session") {
		t.Fatalf("missing save confirmation:\n%s", stdout)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "const savedValue = 7") {
		t.Fatalf("saved session is missing the evaluated command:\n%s", data)
	}
}
