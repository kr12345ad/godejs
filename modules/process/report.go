package process

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

// installReport installs process.report. GodeJS writes the same diagnostic
// report document shape Node produces, built from the live process rather than
// a captured template.
func installReport(e *env.Env, proc js.Object) {
	report := e.JS.Obj()
	cwd, _ := os.Getwd()
	report.SetStr("directory", cwd)
	report.SetStr("filename", "")
	report.SetBool("compact", false)
	report.SetBool("reportOnFatalError", false)
	report.SetBool("reportOnSignal", false)
	report.SetBool("reportOnUncaughtException", false)
	report.SetStr("signal", "SIGUSR2")

	data := func(trigger, message string) []byte {
		payload, err := json.Marshal(reportPayload(e, trigger, message))
		if err != nil {
			payload = []byte("{}")
		}
		return payload
	}

	report.SetFunc("getReport", func(this js.Object, args []js.Value) (js.Value, error) {
		message := ""
		if len(args) > 0 && args[0].IsObject() {
			if m := args[0].Object().Get("message"); m.Valid() && m.IsString() {
				message = m.Str()
			}
		}
		v, err := e.JS.Eval("JSON.parse("+strconv.Quote(string(data("JavaScript API", message)))+")",
			"[gode:report]")
		if err != nil || !v.Valid() {
			return e.JS.Obj().V(), nil
		}
		return v, nil
	})

	report.SetFunc("writeReport", func(this js.Object, args []js.Value) (js.Value, error) {
		name := ""
		message := ""
		if len(args) > 0 {
			switch {
			case args[0].IsString():
				name = args[0].Str()
			case args[0].IsObject():
				if m := args[0].Object().Get("message"); m.Valid() && m.IsString() {
					message = m.Str()
				}
			}
		}
		if name == "" {
			name = defaultReportName()
		}
		path := name
		if !filepath.IsAbs(path) {
			directory := ""
			if d := report.Get("directory"); d.Valid() && d.IsString() {
				directory = d.Str()
			}
			if directory == "" {
				directory, _ = os.Getwd()
			}
			path = filepath.Join(directory, name)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return js.Value{}, err
		}
		if err := os.WriteFile(path, data("JavaScript API", message), 0o600); err != nil {
			return js.Value{}, err
		}
		report.SetStr("filename", filepath.Base(path))
		return e.JS.Str(path), nil
	})

	proc.Set("report", report.V())
}

func defaultReportName() string {
	now := time.Now()
	return "report." + now.Format("20060102") + "." + now.Format("150405") + "." +
		strconv.Itoa(os.Getpid()) + ".0.001.json"
}

func reportPayload(e *env.Env, trigger, message string) map[string]any {
	cwd, _ := os.Getwd()
	hostname, _ := os.Hostname()
	release := strings.TrimSpace(readFileString("/proc/sys/kernel/osrelease"))
	if release == "" {
		release = runtime.GOOS
	}

	argv := make([]any, 0, len(e.Argv()))
	for _, arg := range e.Argv() {
		argv = append(argv, arg)
	}

	envVars := map[string]string{}
	for _, entry := range os.Environ() {
		if at := strings.IndexByte(entry, '='); at > 0 {
			envVars[entry[:at]] = entry[at+1:]
		}
	}

	return map[string]any{
		"header": map[string]any{
			"event":              trigger,
			"trigger":            trigger,
			"filename":           "",
			"dumpEventTime":      time.Now().Format(time.RFC3339),
			"dumpEventTimeStamp": time.Now().UnixMilli(),
			"processId":          os.Getpid(),
			"threadId":           0,
			"cwd":                cwd,
			"commandLine":        argv,
			"nodejsVersion":      e.Opts.Version,
			"wordSize":           strconv.IntSize,
			"arch":               nodeArch(),
			"platform":           runtime.GOOS,
			"release":            release,
			"osVersion":          strings.TrimSpace(readFileString("/proc/version")),
			"osRelease":          release,
			"osMachine":          runtime.GOARCH,
			"host":               hostname,
			"networkInterfaces":  []any{},
		},
		"javascriptStack": map[string]any{
			"message": message,
			"stack":   []string{},
		},
		"nativeStack":           []any{},
		"javascriptHeap":        heapStats(),
		"resourceUsage":         resourceUsage(),
		"uvthreadResourceUsage": map[string]any{},
		"libuv":                 []any{},
		"workers":               []any{},
		"environmentVariables":  envVars,
	}
}

func readFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func heapStats() map[string]any {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return map[string]any{
		"total_heap_size":             stats.HeapSys,
		"total_heap_size_executable":  uint64(0),
		"total_physical_size":         stats.HeapSys,
		"total_available_size":        stats.HeapIdle,
		"used_heap_size":              stats.HeapAlloc,
		"heap_size_limit":             stats.HeapSys + stats.HeapIdle,
		"malloced_memory":             stats.Mallocs,
		"peak_malloced_memory":        stats.TotalAlloc,
		"does_zap_garbage":            false,
		"number_of_native_contexts":   1,
		"number_of_detached_contexts": 0,
	}
}

func resourceUsage() map[string]any {
	usage := map[string]any{
		"userCPUTime":                0,
		"systemCPUTime":              0,
		"maxRSS":                     0,
		"sharedMemorySize":           0,
		"unsharedDataSize":           0,
		"unsharedStackSize":          0,
		"minorPageFault":             0,
		"majorPageFault":             0,
		"swappedOut":                 0,
		"fsRead":                     0,
		"fsWrite":                    0,
		"ipcSent":                    0,
		"ipcReceived":                0,
		"signalsCount":               0,
		"voluntaryContextSwitches":   0,
		"involuntaryContextSwitches": 0,
	}
	var usageInfo syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usageInfo); err != nil {
		return usage
	}
	usage["userCPUTime"] = usageInfo.Utime.Sec*1000 + usageInfo.Utime.Usec/1000
	usage["systemCPUTime"] = usageInfo.Stime.Sec*1000 + usageInfo.Stime.Usec/1000
	usage["maxRSS"] = usageInfo.Maxrss
	usage["minorPageFault"] = usageInfo.Minflt
	usage["majorPageFault"] = usageInfo.Majflt
	usage["swappedOut"] = usageInfo.Nswap
	usage["fsRead"] = usageInfo.Inblock
	usage["fsWrite"] = usageInfo.Oublock
	usage["ipcSent"] = usageInfo.Msgsnd
	usage["ipcReceived"] = usageInfo.Msgrcv
	usage["signalsCount"] = usageInfo.Nsignals
	usage["voluntaryContextSwitches"] = usageInfo.Nvcsw
	usage["involuntaryContextSwitches"] = usageInfo.Nivcsw
	return usage
}
