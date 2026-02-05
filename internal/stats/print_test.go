package stats

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrintReport(t *testing.T) {
	r := NewReport("v0.11.3", "7.0.0-dev.20260107.1")

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
	PrintReport(&buf, r, "/test/dir")
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

func TestPrintReport_NoRules(t *testing.T) {
	r := NewReport("dev", "unknown")
	r.AddProgram("tsconfig.json", 100*time.Millisecond, 50)
	r.AddLintCPU(10 * time.Millisecond)
	r.SetTotal(100 * time.Millisecond)

	var buf bytes.Buffer
	PrintReport(&buf, r, "")
	output := buf.String()

	// Should not panic with empty rules
	if !strings.Contains(output, "Lint:") {
		t.Errorf("Lint section should still be present:\n%s", output)
	}
	if !strings.Contains(output, "Traversal+overhead") {
		t.Errorf("Traversal+overhead line should still be present:\n%s", output)
	}
}

func TestPrintReport_FewRules(t *testing.T) {
	r := NewReport("dev", "unknown")
	r.AddProgram("tsconfig.json", 100*time.Millisecond, 50)
	r.AddLintCPU(20 * time.Millisecond)
	r.SetTotal(100 * time.Millisecond)

	// Only 3 rules - should not show "more rules" line
	r.AddRule("rule1", 10*time.Millisecond)
	r.AddRule("rule2", 5*time.Millisecond)
	r.AddRule("rule3", 1*time.Millisecond)

	var buf bytes.Buffer
	PrintReport(&buf, r, "")
	output := buf.String()

	if strings.Contains(output, "more rules") {
		t.Errorf("Should not show 'more rules' when there are fewer than 5 rules:\n%s", output)
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name       string
		currentDir string
		input      string
		want       string
	}{
		{
			name:       "relative path within cwd",
			currentDir: "/test/dir",
			input:      "/test/dir/tsconfig.json",
			want:       "tsconfig.json",
		},
		{
			name:       "nested relative path",
			currentDir: "/test/dir",
			input:      "/test/dir/packages/core/tsconfig.json",
			want:       "packages/core/tsconfig.json",
		},
		{
			name:       "non-path string returned as-is",
			currentDir: "/test/dir",
			input:      "inferred program",
			want:       "inferred program",
		},
		{
			name:       "empty currentDir returns name as-is",
			currentDir: "",
			input:      "/some/absolute/tsconfig.json",
			want:       "/some/absolute/tsconfig.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayName(tt.currentDir, tt.input)
			if got != tt.want {
				t.Errorf("displayName(%q, %q) = %q, want %q", tt.currentDir, tt.input, got, tt.want)
			}
		})
	}
}

// TestDisplayName_Symlink reproduces the real bug: on macOS, os.Getwd() resolves
// symlinks while tsconfig paths (from VS Code file paths) use the unresolved
// symlink path. This causes displayName to fail to compute a relative path.
func TestDisplayName_Symlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create real directory with a tsconfig.json
	realDir := filepath.Join(tmpDir, "real", "project")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tsconfigPath := filepath.Join(realDir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink: tmpDir/link -> tmpDir/real
	symlinkDir := filepath.Join(tmpDir, "link")
	if err := os.Symlink(filepath.Join(tmpDir, "real"), symlinkDir); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	// Simulate: cwd is the resolved path, tsconfig path uses the symlink
	resolvedCwd := filepath.Join(tmpDir, "real", "project")
	symlinkTsconfig := filepath.Join(symlinkDir, "project", "tsconfig.json")

	got := displayName(resolvedCwd, symlinkTsconfig)
	if got != "tsconfig.json" {
		t.Errorf("displayName(%q, %q)\ngot  %q\nwant %q", resolvedCwd, symlinkTsconfig, got, "tsconfig.json")
	}

	// Also test the reverse: cwd uses symlink, tsconfig is resolved
	got2 := displayName(filepath.Join(symlinkDir, "project"), tsconfigPath)
	if got2 != "tsconfig.json" {
		t.Errorf("displayName(%q, %q)\ngot  %q\nwant %q", filepath.Join(symlinkDir, "project"), tsconfigPath, got2, "tsconfig.json")
	}
}
