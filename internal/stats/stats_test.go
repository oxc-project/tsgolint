package stats

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLintStats_Print(t *testing.T) {
	stats := NewLintStats("v0.11.3", "7.0.0-dev.20260107.1", 8)

	// Add program stats
	stats.AddProgramStat("tsconfig.json", 42*time.Millisecond, 100)
	stats.AddProgramStat("packages/core/tsconfig.json", 80*time.Millisecond, 200)
	stats.AddProgramStat("packages/cli/tsconfig.json", 20*time.Millisecond, 300)
	stats.AddProgramStat("inferred program", 20*time.Millisecond, 500)

	// Add rule times
	stats.AddRuleTime("no_misused_promises", 200*time.Millisecond)
	stats.AddRuleTime("await_thenable", 3*time.Millisecond)
	stats.AddRuleTime("no_floating_promises", 1*time.Millisecond)
	stats.AddRuleTime("no_base_to_string", 1*time.Millisecond)
	stats.AddRuleTime("no_deprecated", 1*time.Millisecond)
	stats.AddRuleTime("rule6", 500*time.Microsecond)
	stats.AddRuleTime("rule7", 400*time.Microsecond)

	var buf bytes.Buffer
	stats.Print(&buf)
	output := buf.String()

	// Check header
	if !strings.Contains(output, "tsgolint stats (4 tsconfigs, 8 threads)") {
		t.Errorf("Header not found in output:\n%s", output)
	}

	// Check version section
	if !strings.Contains(output, "v0.11.3") {
		t.Errorf("tsgolint version not found in output:\n%s", output)
	}
	if !strings.Contains(output, "7.0.0-dev.20260107.1") {
		t.Errorf("tsgo version not found in output:\n%s", output)
	}

	// Check Typecheck section
	if !strings.Contains(output, "Typecheck:") {
		t.Errorf("Typecheck section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "tsconfig.json") {
		t.Errorf("tsconfig.json not found in output:\n%s", output)
	}
	if !strings.Contains(output, "inferred program") {
		t.Errorf("inferred program not found in output:\n%s", output)
	}

	// Check Lint section
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "no_misused_promises") {
		t.Errorf("no_misused_promises rule not found in output:\n%s", output)
	}
	// Should show "2 more rules" for rule6 and rule7
	if !strings.Contains(output, "2 more rules") {
		t.Errorf("collapsed rules line not found in output:\n%s", output)
	}

	// Check Summary section
	if !strings.Contains(output, "Summary:") {
		t.Errorf("Summary section not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Compile:") {
		t.Errorf("Compile line not found in output:\n%s", output)
	}
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint line not found in output:\n%s", output)
	}

	t.Logf("Output:\n%s", output)
}

func TestLintStats_NoRules(t *testing.T) {
	stats := NewLintStats("dev", "unknown", 4)
	stats.AddProgramStat("tsconfig.json", 100*time.Millisecond, 50)

	var buf bytes.Buffer
	stats.Print(&buf)
	output := buf.String()

	// Should not panic with empty rules
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint section should still be present:\n%s", output)
	}
}

func TestLintStats_FewRules(t *testing.T) {
	stats := NewLintStats("dev", "unknown", 4)
	stats.AddProgramStat("tsconfig.json", 100*time.Millisecond, 50)

	// Only 3 rules - should not show "more rules" line
	stats.AddRuleTime("rule1", 10*time.Millisecond)
	stats.AddRuleTime("rule2", 5*time.Millisecond)
	stats.AddRuleTime("rule3", 1*time.Millisecond)

	var buf bytes.Buffer
	stats.Print(&buf)
	output := buf.String()

	if strings.Contains(output, "more rules") {
		t.Errorf("Should not show 'more rules' when there are fewer than 5 rules:\n%s", output)
	}
}

func TestEnabled(t *testing.T) {
	// This test depends on environment, so just check it doesn't panic
	_ = Enabled()
}
