// Package parity scores GodeJS against the tracked Node.js 24 LTS API surface
// and renders progress.svg. The top-line score is the differential program
// corpus: APIs only count once real Node code runs through them.
package parity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Module is one builtin's implemented API list.
type Module struct {
	Name        string
	Implemented []string
}

// ModuleReport is the score for one module.
type ModuleReport struct {
	Name      string
	Done      int
	Total     int
	Percent   float64
	Untracked int
}

// Program is one real-world program in the differential corpus.
type Program struct {
	Name   string
	Passed bool
	Detail string
}

// OfficialModule is one module's score from Node's own test suite.
type OfficialModule struct {
	Name   string `json:"name"`
	Passed int    `json:"passed"`
	Total  int    `json:"total"`
}

// OfficialReport is the result of running Node's test/parallel suite against
// gode. It is the strongest compatibility signal: unmodified Node tests.
type OfficialReport struct {
	Passed    int              `json:"passed"`
	Total     int              `json:"total"`
	Generated string           `json:"generated"`
	Modules   []OfficialModule `json:"modules"`
}

// Report is the whole scoreboard.
type Report struct {
	Modules   []ModuleReport
	Programs  []Program
	Official  *OfficialReport
	Done      int
	Total     int
	Percent   float64
	Generated time.Time
}

// OfficialPercent is the official test suite pass rate.
func (r *Report) OfficialPercent() float64 {
	if r.Official == nil || r.Official.Total == 0 {
		return 0
	}
	return 100 * float64(r.Official.Passed) / float64(r.Official.Total)
}

// LoadOfficial reads a summary produced by running Node's official suite.
func LoadOfficial(path string) (*OfficialReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var o OfficialReport
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	sort.SliceStable(o.Modules, func(i, j int) bool {
		if o.Modules[i].Total != o.Modules[j].Total {
			return o.Modules[i].Total > o.Modules[j].Total
		}
		return o.Modules[i].Name < o.Modules[j].Name
	})
	return &o, nil
}

// ProgramsPassed counts identical program runs.
func (r *Report) ProgramsPassed() int {
	n := 0
	for _, p := range r.Programs {
		if p.Passed {
			n++
		}
	}
	return n
}

// ProgramsPercent is the differential corpus pass rate.
func (r *Report) ProgramsPercent() float64 {
	if len(r.Programs) == 0 {
		return 0
	}
	return 100 * float64(r.ProgramsPassed()) / float64(len(r.Programs))
}

// Compute scores modules against every baseline file in baselineDir. A
// baseline without a matching builtin scores zero rather than disappearing,
// so the headline number never counts an unimplemented module as done.
func Compute(baselineDir string, mods []Module) (*Report, error) {
	r := &Report{Generated: time.Now().UTC()}
	byName := make(map[string]Module, len(mods))
	for _, m := range mods {
		byName[m.Name] = m
	}
	files, err := filepath.Glob(filepath.Join(baselineDir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		baseline, err := loadBaseline(path)
		if err != nil {
			return nil, err
		}
		m := byName[name]
		total := len(baseline)
		done := 0
		untracked := 0
		for _, k := range m.Implemented {
			if _, ok := baseline[k]; ok {
				done++
			} else {
				untracked++
			}
		}
		pct := 0.0
		if total > 0 {
			pct = 100 * float64(done) / float64(total)
		}
		r.Modules = append(r.Modules, ModuleReport{
			Name: name, Done: done, Total: total, Percent: pct, Untracked: untracked,
		})
		r.Done += done
		r.Total += total
	}
	if r.Total > 0 {
		r.Percent = 100 * float64(r.Done) / float64(r.Total)
	}
	return r, nil
}

func loadBaseline(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return set, nil
}

// Table renders a plain-text scoreboard.
func (r *Report) Table() string {
	var b strings.Builder
	if len(r.Programs) > 0 {
		fmt.Fprintf(&b, "%-16s %6s %6s %7s\n", "program", "pass", "total", "rate")
		fmt.Fprintf(&b, "%-16s %6s %6s %7s\n", "-------", "----", "-----", "----")
		for _, p := range r.Programs {
			status := "FAIL"
			if p.Passed {
				status = "PASS"
			}
			fmt.Fprintf(&b, "%-16s %6s\n", p.Name, status)
			if !p.Passed && p.Detail != "" {
				fmt.Fprintf(&b, "  %s\n", p.Detail)
			}
		}
		fmt.Fprintf(&b, "%-16s %6d %6d %6.1f%%\n\n", "programs", r.ProgramsPassed(), len(r.Programs), r.ProgramsPercent())
	}
	fmt.Fprintf(&b, "%-16s %6s %6s %7s\n", "module", "done", "total", "parity")
	fmt.Fprintf(&b, "%-16s %6s %6s %7s\n", "------", "----", "-----", "------")
	for _, m := range r.Modules {
		fmt.Fprintf(&b, "%-16s %6d %6d %6.1f%%\n", m.Name, m.Done, m.Total, m.Percent)
	}
	fmt.Fprintf(&b, "%-16s %6d %6d %6.1f%%\n", "overall", r.Done, r.Total, r.Percent)
	if r.Official != nil {
		fmt.Fprintf(&b, "\n%-16s %6s %6s %7s\n", "official", "pass", "total", "rate")
		for _, m := range r.Official.Modules {
			pct := 0.0
			if m.Total > 0 {
				pct = 100 * float64(m.Passed) / float64(m.Total)
			}
			fmt.Fprintf(&b, "%-16s %6d %6d %6.1f%%\n", m.Name, m.Passed, m.Total, pct)
		}
		fmt.Fprintf(&b, "%-16s %6d %6d %6.1f%%\n", "suite", r.Official.Passed, r.Official.Total, r.OfficialPercent())
	}
	return b.String()
}

const (
	svgWidth  = 1000
	svgHeader = 178
	svgRow    = 34
	svgFooter = 74
	barX      = 228
	barWidth  = 610
	labelX    = 48
	pctX      = 902
	countX    = 956
)

// SVG renders progress.svg. The headline is the differential program corpus;
// the API rows below it explain what is available to those programs.
func (r *Report) SVG() string {
	programs := len(r.Programs)
	programsHeight := 0
	if programs > 0 {
		programsHeight = 38 + programs*24 + 36
	}
	officialModules := 0
	officialHeight := 0
	if r.Official != nil {
		officialModules = len(r.Official.Modules)
		officialHeight = 38 + officialModules*svgRow + 40
	}
	height := svgHeader + officialHeight + programsHeight + svgRow*len(r.Modules) + svgFooter
	headline := r.Percent
	headlineLabel := "of the tracked core API surface implemented"
	headlineCount := fmt.Sprintf("%d / %d APIs", r.Done, r.Total)
	if programs > 0 {
		headline = r.ProgramsPercent()
		headlineLabel = "of the real Node programs in the differential corpus run byte-identically"
		headlineCount = fmt.Sprintf("%d / %d programs", r.ProgramsPassed(), programs)
	}
	if r.Official != nil {
		headline = r.OfficialPercent()
		headlineLabel = "of Node's official test suite passing unchanged"
		headlineCount = fmt.Sprintf("%d / %d tests", r.Official.Passed, r.Official.Total)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="GodeJS Node.js 24 LTS parity: %.1f percent">`+"\n",
		svgWidth, height, svgWidth, height, headline)
	b.WriteString(`<rect width="100%" height="100%" fill="#0d1117"/>` + "\n")
	b.WriteString(`<rect x="24" y="24" width="952" height="` + fmt.Sprint(height-48) + `" rx="14" fill="#11161d" stroke="#21262d"/>` + "\n")
	b.WriteString(`<text x="48" y="62" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#8b949e" letter-spacing="2">GODEJS · NODE.JS 24 LTS COMPATIBILITY</text>` + "\n")
	fmt.Fprintf(&b, `<text x="48" y="126" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="52" font-weight="700" fill="#e6edf3">%.1f%%</text>`+"\n", headline)
	fmt.Fprintf(&b, `<text x="48" y="154" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="#8b949e">%s</text>`+"\n", escape(headlineLabel))
	fmt.Fprintf(&b, `<text x="952" y="126" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="15" fill="#c9d1d9">%s</text>`+"\n", escape(headlineCount))
	fmt.Fprintf(&b, `<text x="952" y="152" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#6e7681">%s</text>`+"\n", escape(fmt.Sprintf("%.1f%% API surface · %d/%d", r.Percent, r.Done, r.Total)))
	b.WriteString(`<line x1="48" y1="170" x2="952" y2="170" stroke="#21262d"/>` + "\n")

	y := svgHeader
	if r.Official != nil {
		fmt.Fprintf(&b, `<text x="48" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#8b949e" letter-spacing="1">NODE OFFICIAL TEST SUITE · test/parallel run unmodified against gode</text>`+"\n", y+14)
		y += 38
		for _, m := range r.Official.Modules {
			pct := 0.0
			if m.Total > 0 {
				pct = 100 * float64(m.Passed) / float64(m.Total)
			}
			fill := barColor(pct)
			w := 0.0
			if m.Total > 0 {
				w = float64(m.Passed) / float64(m.Total) * barWidth
			}
			fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="#c9d1d9">%s</text>`+"\n",
				labelX, y+15, escape(m.Name))
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="10" rx="5" fill="#21262d"/>`+"\n", barX, y+7, barWidth)
			if w > 0 {
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%.1f" height="10" rx="5" fill="%s"/>`+"\n", barX, y+7, w, fill)
			}
			fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="%s">%.1f%%</text>`+"\n",
				pctX, y+15, fill, pct)
			fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#6e7681">%d/%d</text>`+"\n",
				countX, y+15, m.Passed, m.Total)
			y += svgRow
		}
		y += 24
	}
	if programs > 0 {
		fmt.Fprintf(&b, `<text x="48" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#8b949e" letter-spacing="1">REAL NODE PROGRAMS · run with node and gode, compare stdout and exit status</text>`+"\n", y+14)
		y += 38
		for _, p := range r.Programs {
			color := "#f85149"
			status := "FAIL"
			if p.Passed {
				color = "#3fb950"
				status = "PASS"
			}
			fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="#c9d1d9">%s</text>`+"\n",
				labelX, y+15, escape(p.Name))
			fmt.Fprintf(&b, `<text x="952" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="%s">%s</text>`+"\n",
				y+15, color, status)
			y += 24
		}
		y += 12
		fmt.Fprintf(&b, `<text x="48" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#8b949e" letter-spacing="1">CORE API SURFACE · implemented and unit-tested</text>`+"\n", y+14)
		y += 30
	} else {
		y += 12
	}

	for _, m := range r.Modules {
		fill := barColor(m.Percent)
		w := 0.0
		if m.Total > 0 {
			w = float64(m.Done) / float64(m.Total) * barWidth
		}
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="#c9d1d9">%s</text>`+"\n",
			labelX, y+15, escape(m.Name))
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="10" rx="5" fill="#21262d"/>`+"\n", barX, y+7, barWidth)
		if w > 0 {
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%.1f" height="10" rx="5" fill="%s"/>`+"\n", barX, y+7, w, fill)
		}
		fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="13" fill="%s">%.1f%%</text>`+"\n",
			pctX, y+15, fill, m.Percent)
		fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#6e7681">%d/%d</text>`+"\n",
			countX, y+15, m.Done, m.Total)
		y += svgRow
	}

	// Legend / footer.
	fy := height - 44
	fmt.Fprintf(&b, `<text x="48" y="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="11" fill="#8b949e">%s</text>`+"\n",
		fy, escape("baseline: nodejs.org/docs/latest-v24.x · gode v0.1.0 · "+r.Generated.Format("2006-01-02 15:04 UTC")))
	fmt.Fprintf(&b, `<text x="952" y="%d" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="11" fill="#6e7681">%s</text>`+"\n",
		fy, escape("generated from unit tests and differential program runs"))
	b.WriteString("</svg>\n")
	return b.String()
}

func barColor(pct float64) string {
	switch {
	case pct >= 75:
		return "#3fb950"
	case pct >= 40:
		return "#58a6ff"
	case pct >= 20:
		return "#d29922"
	default:
		return "#f85149"
	}
}

func escape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

// SortedNames is a helper for deterministic output.
func SortedNames(mods []Module) []string {
	out := make([]string, 0, len(mods))
	for _, m := range mods {
		out = append(out, m.Name)
	}
	sort.Strings(out)
	return out
}
