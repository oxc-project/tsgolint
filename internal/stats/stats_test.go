package stats

import (
	"testing"
	"time"
)

func TestNewReport(t *testing.T) {
	r := NewReport("v1.0.0", "7.0.0")
	if r.TsgolintVersion != "v1.0.0" {
		t.Errorf("TsgolintVersion = %q, want %q", r.TsgolintVersion, "v1.0.0")
	}
	if r.TsgoVersion != "7.0.0" {
		t.Errorf("TsgoVersion = %q, want %q", r.TsgoVersion, "7.0.0")
	}
	if len(r.Programs) != 0 {
		t.Errorf("Programs should be empty, got %d", len(r.Programs))
	}
	if len(r.Rules) != 0 {
		t.Errorf("Rules should be empty, got %d", len(r.Rules))
	}
}

func TestAddProgram(t *testing.T) {
	r := NewReport("dev", "unknown")
	r.AddProgram("tsconfig.json", 100*time.Millisecond, 50)
	r.AddProgram("tsconfig.app.json", 200*time.Millisecond, 30)

	if len(r.Programs) != 2 {
		t.Fatalf("Programs count = %d, want 2", len(r.Programs))
	}
	if r.TsconfigCount != 2 {
		t.Errorf("TsconfigCount = %d, want 2", r.TsconfigCount)
	}
	if r.Compile != 300*time.Millisecond {
		t.Errorf("Compile = %v, want 300ms", r.Compile)
	}
}

func TestAddRule(t *testing.T) {
	r := NewReport("dev", "unknown")
	r.AddRule("no_misused_promises", 100*time.Millisecond)
	r.AddRule("no_misused_promises", 50*time.Millisecond) // same rule, should accumulate

	if got := r.Rules["no_misused_promises"]; got != 150*time.Millisecond {
		t.Errorf("Rules[no_misused_promises] = %v, want 150ms", got)
	}
}

func TestAddLintTimings(t *testing.T) {
	r := NewReport("dev", "unknown")
	r.AddLintWall(100 * time.Millisecond)
	r.AddLintCPU(200 * time.Millisecond)
	r.SetTotal(500 * time.Millisecond)

	if r.LintWall != 100*time.Millisecond {
		t.Errorf("LintWall = %v, want 100ms", r.LintWall)
	}
	if r.LintCPU != 200*time.Millisecond {
		t.Errorf("LintCPU = %v, want 200ms", r.LintCPU)
	}
	if r.Total != 500*time.Millisecond {
		t.Errorf("Total = %v, want 500ms", r.Total)
	}
}
