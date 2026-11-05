package main

// --watch: run the entry script in a child process and restart it when the
// script or an explicitly watched path changes. The child is an ordinary gode
// process (GODE_WATCH_CHILD=1 stops it from watching again).

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func watchLoop(argv []string, opts options) int {
	if opts.script == "" {
		fmt.Fprintln(os.Stderr, "gode: --watch requires a script file")
		return 9
	}
	if len(opts.watchPaths) == 0 {
		opts.watchPaths = []string{opts.script}
	}
	files := collectWatchFiles(opts.watchPaths)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "gode: --watch: no files to watch")
		return 9
	}
	args := stripWatchArgs(argv[1:])
	exe, err := os.Executable()
	if err != nil {
		exe = argv[0]
	}

	start := func() (*exec.Cmd, chan error) {
		cmd := exec.Command(exe, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Env = append(os.Environ(), "GODE_WATCH_CHILD=1")
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "gode: --watch: %v\n", err)
			return nil, nil
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		return cmd, done
	}

	stamps := snapshot(files)
	cmd, done := start()
	if cmd == nil {
		return 1
	}
	running := true
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		var doneCh chan error
		if running {
			doneCh = done
		}
		select {
		case err := <-doneCh:
			// A watch session stays alive when the child finishes: Node waits
			// for a file change and then starts a fresh child.
			running = false
			cmd = nil
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stdout, "Failed running %q (exit code %d). Waiting for file changes before restarting...\n", opts.script, exit.ExitCode())
				} else {
					fmt.Fprintf(os.Stderr, "gode: --watch: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stdout, "Completed running %q. Waiting for file changes before restarting...\n", opts.script)
			}
		case <-signals:
			if running && cmd != nil {
				_ = cmd.Process.Signal(watchSignal(opts.watchKill))
				<-done
			}
			return 130
		case <-ticker.C:
			next := snapshot(files)
			if !changed(stamps, next) {
				continue
			}
			stamps = next
			// Pick up files created after startup when a directory is watched.
			files = collectWatchFiles(opts.watchPaths)
			stamps = snapshot(files)
			fmt.Fprintln(os.Stdout, "Restarting...")
			if running && cmd != nil {
				_ = cmd.Process.Signal(watchSignal(opts.watchKill))
				<-done
			}
			cmd, done = start()
			if cmd == nil {
				return 1
			}
			running = true
		}
	}
}

func watchSignal(name string) os.Signal {
	switch name {
	case "SIGINT":
		return syscall.SIGINT
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGUSR1":
		return syscall.SIGUSR1
	case "SIGUSR2":
		return syscall.SIGUSR2
	default:
		return syscall.SIGTERM
	}
}

func collectWatchFiles(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(path string) {
		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		if !seen[abs] {
			seen[abs] = true
			out = append(out, abs)
		}
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			add(path)
			continue
		}
		if !info.IsDir() {
			add(path)
			continue
		}
		count := 0
		_ = filepath.WalkDir(path, func(p string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			if strings0(entry.Name(), ".git", "node_modules") {
				return nil
			}
			add(p)
			count++
			if count > 20000 {
				return filepath.SkipAll
			}
			return nil
		})
	}
	return out
}

// strings0 reports whether name equals any of the ignored directory names.
func strings0(name string, ignored ...string) bool {
	for _, item := range ignored {
		if name == item {
			return true
		}
	}
	return false
}

type fileStamp struct {
	mod  int64
	size int64
}

func snapshot(files []string) map[string]fileStamp {
	out := make(map[string]fileStamp, len(files))
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		out[path] = fileStamp{mod: info.ModTime().UnixNano(), size: info.Size()}
	}
	return out
}

func changed(before, after map[string]fileStamp) bool {
	if len(before) != len(after) {
		return true
	}
	for path, stamp := range after {
		if before[path] != stamp {
			return true
		}
	}
	return false
}

// stripWatchArgs removes the watch flags from a child command line.
func stripWatchArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--watch", arg == "--watch-preserve-output":
			continue
		case arg == "--watch-path", arg == "--watch-kill-signal":
			i++
			continue
		case len(arg) > len("--watch-path=") && arg[:len("--watch-path=")] == "--watch-path=":
			continue
		case len(arg) > len("--watch-kill-signal=") && arg[:len("--watch-kill-signal=")] == "--watch-kill-signal=":
			continue
		}
		out = append(out, arg)
	}
	return out
}
