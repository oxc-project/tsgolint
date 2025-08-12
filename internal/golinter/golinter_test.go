package golinter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/gorule"
	"github.com/typescript-eslint/tsgolint/internal/gorules/no_unused_vars"
	"github.com/typescript-eslint/tsgolint/internal/gorules/inefficient_string_concat"
)

func TestGoLinter(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a test Go file
	testCode := `package main

func test() {
	unusedVar := "hello"
	
	result := ""
	for i := 0; i < 10; i++ {
		result += "test"  
	}
}
`
	
	testFile := filepath.Join(tmpDir, "test_lint.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Parse the test file
	files, err := ParseGoFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse Go files: %v", err)
	}
	
	if len(files) == 0 {
		t.Fatal("No test files found")
	}
	
	// Collect diagnostics
	var diagnostics []gorule.GoRuleDiagnostic
	
	err = RunGoLinter(
		files,
		func(file GoFile) []ConfiguredGoRule {
			return []ConfiguredGoRule{
				{
					Name: no_unused_vars.NoUnusedVarsRule.Name,
					Run: func(ctx gorule.GoRuleContext) gorule.GoRuleListeners {
						return no_unused_vars.NoUnusedVarsRule.Run(ctx, nil)
					},
				},
				{
					Name: inefficient_string_concat.InefficientStringConcatRule.Name,
					Run: func(ctx gorule.GoRuleContext) gorule.GoRuleListeners {
						return inefficient_string_concat.InefficientStringConcatRule.Run(ctx, nil)
					},
				},
			}
		},
		func(d gorule.GoRuleDiagnostic) {
			diagnostics = append(diagnostics, d)
		},
	)
	
	if err != nil {
		t.Fatalf("Failed to run Go linter: %v", err)
	}
	
	// Verify we got the expected diagnostics
	if len(diagnostics) < 2 {
		t.Errorf("Expected at least 2 diagnostics, got %d", len(diagnostics))
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s - %s", d.RuleName, d.Message.Description)
		}
	}
	
	// Check for unused variable diagnostic
	foundUnusedVar := false
	foundInefficientConcat := false
	
	for _, diag := range diagnostics {
		if diag.RuleName == "no-unused-vars" {
			foundUnusedVar = true
		}
		if diag.RuleName == "inefficient-string-concat" {
			foundInefficientConcat = true
		}
	}
	
	if !foundUnusedVar {
		t.Error("Expected to find unused variable diagnostic")
	}
	
	if !foundInefficientConcat {
		t.Error("Expected to find inefficient string concatenation diagnostic")
	}
}