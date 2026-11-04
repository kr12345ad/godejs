// Package compat runs real Node programs twice — once with node and once with
// gode — and reports whether stdout and the exit status are identical. API
// presence is a weaker signal than this: a module can expose every name and
// still fail the first real package that imports it.
package compat

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Result is the outcome for one program.
type Result struct {
	Name     string
	Dir      string
	Entry    string
	Passed   bool
	NodeExit int
	GodeExit int
	NodeOut  []byte
	GodeOut  []byte
	Err      string
}

// Discover finds every program directory under root. A program is a directory
// containing main.mjs or main.js; main.mjs wins when both exist.
func Discover(root string) ([]Result, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []Result
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		main := filepath.Join(dir, "main.mjs")
		if _, err := os.Stat(main); err != nil {
			main = filepath.Join(dir, "main.js")
			if _, err := os.Stat(main); err != nil {
				continue
			}
		}
		out = append(out, Result{Name: entry.Name(), Dir: dir, Entry: filepath.Base(main)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Run executes each case with node and gode. Cases are compared on stdout and
// exit status; stderr is kept for reporting gode failures.
func Run(root, nodeBin, godeBin string) ([]Result, error) {
	cases, err := Discover(root)
	if err != nil {
		return nil, err
	}
	for i := range cases {
		c := &cases[i]
		nodeOut, _, nodeExit, err := run(nodeBin, c.Entry, c.Dir)
		if err != nil {
			return nil, fmt.Errorf("%s: node: %w", c.Name, err)
		}
		godeOut, godeErr, godeExit, err := run(godeBin, c.Entry, c.Dir)
		if err != nil {
			return nil, fmt.Errorf("%s: gode: %w", c.Name, err)
		}
		c.NodeOut, c.GodeOut = nodeOut, godeOut
		c.NodeExit, c.GodeExit = nodeExit, godeExit
		c.Passed = nodeExit == godeExit && bytes.Equal(nodeOut, godeOut)
		if !c.Passed {
			c.Err = describeMismatch(nodeOut, godeOut, nodeExit, godeExit, godeErr)
		}
	}
	return cases, nil
}

func run(bin, entry, dir string) (stdout, stderr []byte, exit int, err error) {
	cmd := exec.Command(bin, entry)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	runErr := cmd.Run()
	if runErr != nil {
		var exitErr *exec.ExitError
		if !asExitError(runErr, &exitErr) {
			return nil, nil, 0, runErr
		}
		exit = exitErr.ExitCode()
	}
	return out.Bytes(), errOut.Bytes(), exit, nil
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

func describeMismatch(nodeOut, godeOut []byte, nodeExit, godeExit int, godeErr []byte) string {
	var b strings.Builder
	fmt.Fprintf(&b, "exit node=%d gode=%d", nodeExit, godeExit)
	if !bytes.Equal(nodeOut, godeOut) {
		fmt.Fprintf(&b, "; stdout differs: %s", firstDifference(nodeOut, godeOut))
	}
	if len(godeErr) > 0 {
		lines := strings.Split(strings.TrimSpace(string(godeErr)), "\n")
		if len(lines) > 1 {
			lines = lines[:1]
		}
		fmt.Fprintf(&b, "; gode stderr: %s", strings.TrimSpace(lines[0]))
	}
	return b.String()
}

func firstDifference(want, got []byte) string {
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			return fmt.Sprintf("first difference at byte %d: node %q, gode %q", i, snippet(want, i), snippet(got, i))
		}
	}
	if len(want) != len(got) {
		return fmt.Sprintf("length node=%d gode=%d", len(want), len(got))
	}
	return "none"
}

func snippet(b []byte, at int) string {
	start := at - 20
	if start < 0 {
		start = 0
	}
	end := at + 40
	if end > len(b) {
		end = len(b)
	}
	return string(b[start:end])
}
