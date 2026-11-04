package parity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeScoresBaselinesAndMissingModules(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, "alpha", []string{"alpha.a", "alpha.b", "alpha.c"})
	writeBaseline(t, dir, "beta", []string{"beta.a", "beta.b"})

	report, err := Compute(dir, []Module{
		{Name: "alpha", Implemented: []string{"alpha.a", "alpha.b", "alpha.unknown"}},
		// beta has no builtin: it must still count against the total.
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Done != 2 || report.Total != 5 {
		t.Fatalf("done/total = %d/%d, want 2/5", report.Done, report.Total)
	}
	if len(report.Modules) != 2 || report.Modules[0].Untracked != 1 {
		t.Fatalf("modules = %+v", report.Modules)
	}
	if report.Percent < 39.9 || report.Percent > 40.1 {
		t.Fatalf("percent = %f", report.Percent)
	}
}

func TestSVGHeadlineFollowsPrograms(t *testing.T) {
	report := &Report{
		Modules:  []ModuleReport{{Name: "alpha", Done: 1, Total: 2, Percent: 50}},
		Programs: []Program{{Name: "real", Passed: true}, {Name: "broken", Passed: false}},
		Done:     1,
		Total:    2,
		Percent:  50,
	}
	svg := report.SVG()
	for _, want := range []string{
		"50.0%",
		"1 / 2 programs",
		"REAL NODE PROGRAMS",
		"real",
		"broken",
		"PASS",
		"FAIL",
		"CORE API SURFACE",
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if strings.Count(svg, "PASS") != 1 || strings.Count(svg, "FAIL") != 1 {
		t.Fatalf("program statuses are wrong:\n%s", svg)
	}
}

func TestTableReportsProgramsFirst(t *testing.T) {
	report := &Report{
		Modules:  []ModuleReport{{Name: "alpha", Done: 1, Total: 2, Percent: 50}},
		Programs: []Program{{Name: "real", Passed: true}},
		Done:     1,
		Total:    2,
		Percent:  50,
	}
	table := report.Table()
	if strings.Index(table, "program") > strings.Index(table, "module") {
		t.Fatalf("programs should be reported first:\n%s", table)
	}
	if !strings.Contains(table, "real") || !strings.Contains(table, "PASS") {
		t.Fatalf("table missing program result:\n%s", table)
	}
}

func writeBaseline(t *testing.T, dir, name string, keys []string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("[")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`"` + k + `"`)
	}
	b.WriteString("]")
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
