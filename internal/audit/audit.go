// Package audit runs the Vigor standards audit on a directory.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// Severity grades a check result.
type Severity string

const (
	SeverityPass Severity = "pass"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
	SeveritySkip Severity = "skip"
)

// Result is one entry in the audit report.
type Result struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
	Standard   string   `json:"standard,omitempty"`
}

// Check is a single audit step. Implementations live in checks/*.go.
type Check interface {
	ID() string
	Category() string
	Run(ctx context.Context, root string) Result
}

var registry []Check

// Register attaches a Check at init() time. Audit() runs every registered check.
func Register(c Check) { registry = append(registry, c) }

// Audit runs all registered checks against `root` and returns a Report.
func Audit(ctx context.Context, root string) Report {
	results := make([]Result, 0, len(registry))
	for _, c := range registry {
		results = append(results, c.Run(ctx, root))
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return Report{Root: root, Results: results}
}

// Report is the aggregate audit output.
type Report struct {
	Root    string   `json:"root"`
	Results []Result `json:"results"`
}

// Summary tallies counts per severity.
func (r Report) Summary() map[Severity]int {
	out := map[Severity]int{SeverityPass: 0, SeverityWarn: 0, SeverityFail: 0, SeveritySkip: 0}
	for _, res := range r.Results {
		out[res.Severity]++
	}
	return out
}

// WriteMarkdown emits a human-readable report.
func (r Report) WriteMarkdown(w io.Writer) {
	fmt.Fprintf(w, "# Vigor Standards Audit\n\n")
	fmt.Fprintf(w, "**Root:** `%s`\n\n", r.Root)
	s := r.Summary()
	fmt.Fprintf(w, "**Summary:** ✓ %d pass · ⚠ %d warn · ✗ %d fail · — %d skip\n\n",
		s[SeverityPass], s[SeverityWarn], s[SeverityFail], s[SeveritySkip])
	fmt.Fprintln(w, "| Severity | Category | Check | Message | Fix |")
	fmt.Fprintln(w, "|----------|----------|-------|---------|-----|")
	for _, res := range r.Results {
		icon := "—"
		switch res.Severity {
		case SeverityPass:
			icon = "✓"
		case SeverityWarn:
			icon = "⚠"
		case SeverityFail:
			icon = "✗"
		}
		fmt.Fprintf(w, "| %s | %s | `%s` | %s | %s |\n",
			icon, res.Category, res.ID, escape(res.Message), escape(res.Suggestion))
	}
}

// WriteJSON emits the report as a JSON document for piping into other tools.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("encode audit report: %w", err)
	}
	return nil
}

func escape(s string) string {
	out := ""
	for _, r := range s {
		switch r {
		case '|':
			out += "\\|"
		case '\n':
			out += " "
		default:
			out += string(r)
		}
	}
	return out
}
