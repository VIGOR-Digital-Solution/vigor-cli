package audit

import (
	"bytes"
	"context"
	"testing"
)

type fakeCheck struct {
	id, category string
	result       Result
}

func (f fakeCheck) ID() string       { return f.id }
func (f fakeCheck) Category() string { return f.category }
func (f fakeCheck) Run(_ context.Context, _ string) Result {
	return f.result
}

func TestAudit_runsAllRegisteredChecks(t *testing.T) {
	// Snapshot + restore the registry so the test is hermetic.
	prev := registry
	t.Cleanup(func() { registry = prev })

	registry = []Check{
		fakeCheck{
			id:       "x.passes",
			category: "X",
			result:   Result{ID: "x.passes", Category: "X", Severity: SeverityPass, Message: "ok"},
		},
		fakeCheck{
			id:       "x.fails",
			category: "X",
			result:   Result{ID: "x.fails", Category: "X", Severity: SeverityFail, Message: "no"},
		},
	}

	report := Audit(context.Background(), "/some/path")
	if len(report.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(report.Results))
	}
	if report.Root != "/some/path" {
		t.Errorf("root = %q", report.Root)
	}

	summary := report.Summary()
	if summary[SeverityPass] != 1 || summary[SeverityFail] != 1 {
		t.Errorf("summary: %v", summary)
	}
}

func TestReport_WriteMarkdown(t *testing.T) {
	r := Report{
		Root: "/root",
		Results: []Result{
			{ID: "a.b", Category: "Cat", Severity: SeverityPass, Message: "fine"},
		},
	}
	var buf bytes.Buffer
	r.WriteMarkdown(&buf)
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("/root")) {
		t.Errorf("missing root: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("a.b")) {
		t.Errorf("missing check id: %s", out)
	}
}
