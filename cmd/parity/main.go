// Command parity runs the test suite, executes the differential program corpus
// with both node and gode, scores both against the tracked Node 24 baseline,
// and writes progress.svg.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"godejs/internal/all"
	"godejs/internal/compat"
	"godejs/internal/parity"
)

func main() {
	out := flag.String("out", "progress.svg", "SVG output path")
	baseline := flag.String("baseline", "parity/baseline", "baseline directory")
	corpus := flag.String("corpus", "parity/cases", "differential program corpus")
	official := flag.String("official", "parity/official.json", "official Node test suite summary (optional)")
	node := flag.String("node", "node", "node executable")
	gode := flag.String("gode", "", "gode executable (built into a temp dir when empty)")
	skipTests := flag.Bool("skip-tests", false, "do not run go test first")
	skipPrograms := flag.Bool("skip-programs", false, "score APIs only; do not run the corpus")
	flag.Parse()

	if !*skipTests {
		cmd := exec.Command("go", "test", "./...", "-count=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "parity: tests failed; refusing to report parity")
			os.Exit(1)
		}
	}

	mods := make([]parity.Module, 0, len(all.Modules()))
	for _, m := range all.Modules() {
		mods = append(mods, parity.Module{Name: m.Name, Implemented: m.Implemented()})
	}
	report, err := parity.Compute(*baseline, mods)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parity:", err)
		os.Exit(1)
	}

	if *official != "" {
		if o, err := parity.LoadOfficial(*official); err == nil {
			report.Official = o
		} else if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "parity: official summary:", err)
			os.Exit(1)
		}
	}

	if !*skipPrograms {
		bin := *gode
		if bin == "" {
			dir, err := os.MkdirTemp("", "gode-parity-")
			if err != nil {
				fmt.Fprintln(os.Stderr, "parity:", err)
				os.Exit(1)
			}
			defer os.RemoveAll(dir)
			bin = filepath.Join(dir, "gode")
			build := exec.Command("go", "build", "-o", bin, "./cmd/gode")
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			if err := build.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "parity: building gode failed")
				os.Exit(1)
			}
		}
		results, err := compat.Run(*corpus, *node, bin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "parity:", err)
			os.Exit(1)
		}
		for _, r := range results {
			report.Programs = append(report.Programs, parity.Program{
				Name: r.Name, Passed: r.Passed, Detail: r.Err,
			})
		}
	}

	if err := os.WriteFile(*out, []byte(report.SVG()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "parity:", err)
		os.Exit(1)
	}
	fmt.Print(report.Table())
	fmt.Printf("\nwrote %s\n", *out)
}
