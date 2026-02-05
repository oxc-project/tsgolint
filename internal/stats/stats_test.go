package stats

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestReport_Print(t *testing.T) {
	r := NewReport("v0.11.3", "7.0.0-dev.20260107.1", 8, "/test/dir")

	// Add program stats with absolute paths
	r.AddProgram("/test/dir/tsconfig.json", 42*time.Millisecond, 100)
	r.AddProgram("/test/dir/packages/core/tsconfig.json", 80*time.Millisecond, 200)
	r.AddProgram("/test/dir/packages/cli/tsconfig.json", 20*time.Millisecond, 300)
	r.AddProgram("inferred program", 20*time.Millisecond, 500)

	// Add rule times
	r.AddRule("no_misused_promises", 200*time.Millisecond)
	r.AddRule("await_thenable", 3*time.Millisecond)
	r.AddRule("no_floating_promises", 1*time.Millisecond)
	r.AddRule("no_base_to_string", 1*time.Millisecond)
	r.AddRule("no_deprecated", 1*time.Millisecond)
	r.AddRule("rule6", 500*time.Microsecond)
	r.AddRule("rule7", 400*time.Microsecond)
	r.AddLintCPU(300 * time.Millisecond)
	r.SetTotal(500 * time.Millisecond)

	var buf bytes.Buffer
	r.Print(&buf)
	output := buf.String()

	// Check version section
	if !strings.Contains(output, "v0.11.3") {
		t.Errorf("tsgolint version not found in output:\n%s", output)
	}
	if !strings.Contains(output, "7.0.0-dev.20260107.1") {
		t.Errorf("tsgo version not found in output:\n%s", output)
	}

	// Check Typecheck section with "Wall Time" header
	if !strings.Contains(output, "Typecheck:") {
		t.Errorf("Typecheck section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Wall Time") {
		t.Errorf("Wall Time header not found in output:\n%s", output)
	}
	// Should show relative paths
	if !strings.Contains(output, "tsconfig.json") {
		t.Errorf("tsconfig.json not found in output:\n%s", output)
	}
	if !strings.Contains(output, "packages/core/tsconfig.json") {
		t.Errorf("relative path packages/core/tsconfig.json not found in output:\n%s", output)
	}
	if !strings.Contains(output, "inferred program") {
		t.Errorf("inferred program not found in output:\n%s", output)
	}

	// Check Lint section with "CPU Time" header
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "CPU Time") {
		t.Errorf("CPU Time header not found in output:\n%s", output)
	}
	if !strings.Contains(output, "no_misused_promises") {
		t.Errorf("no_misused_promises rule not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Traversal+overhead") {
		t.Errorf("Traversal+overhead line not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Total") {
		t.Errorf("Total line not found in output:\n%s", output)
	}
	// Should show "2 more rules" for rule6 and rule7
	if !strings.Contains(output, "2 more rules") {
		t.Errorf("collapsed rules line not found in output:\n%s", output)
	}

	// Check Summary section
	if !strings.Contains(output, "Summary:") {
		t.Errorf("Summary section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Category") {
		t.Errorf("Category header not found in output:\n%s", output)
	}
	if !strings.Contains(output, "typecheck") {
		t.Errorf("typecheck row not found in output:\n%s", output)
	}
	if !strings.Contains(output, "lint") {
		t.Errorf("lint row not found in output:\n%s", output)
	}

	// Check Unicode separator
	if !strings.Contains(output, "─") {
		t.Errorf("Unicode separator not found in output:\n%s", output)
	}

	t.Logf("Output:\n%s", output)
}

func TestReport_NoRules(t *testing.T) {
	r := NewReport("dev", "unknown", 4, "")
	r.AddProgram("tsconfig.json", 100*time.Millisecond, 50)
	r.AddLintCPU(10 * time.Millisecond)
	r.SetTotal(100 * time.Millisecond)

	var buf bytes.Buffer
	r.Print(&buf)
	output := buf.String()

	// Should not panic with empty rules
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint section should still be present:\n%s", output)
	}
	if !strings.Contains(output, "Traversal+overhead") {
		t.Errorf("Traversal+overhead line should still be present:\n%s", output)
	}
}

func TestReport_FewRules(t *testing.T) {
	r := NewReport("dev", "unknown", 4, "")
	r.AddProgram("tsconfig.json", 100*time.Millisecond, 50)
	r.AddLintCPU(20 * time.Millisecond)
	r.SetTotal(100 * time.Millisecond)

	// Only 3 rules - should not show "more rules" line
	r.AddRule("rule1", 10*time.Millisecond)
	r.AddRule("rule2", 5*time.Millisecond)
	r.AddRule("rule3", 1*time.Millisecond)

	var buf bytes.Buffer
	r.Print(&buf)
	output := buf.String()

	if strings.Contains(output, "more rules") {
		t.Errorf("Should not show 'more rules' when there are fewer than 5 rules:\n%s", output)
	}
}

func TestEnabled(t *testing.T) {
	// This test depends on environment, so just check it doesn't panic
	_ = Enabled()
}
