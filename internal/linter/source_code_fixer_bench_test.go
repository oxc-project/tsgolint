package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// mockLintMessage implements LintMessage interface for benchmarking
type mockLintMessage struct {
	fixes []rule.RuleFix
}

func (m mockLintMessage) Fixes() []rule.RuleFix {
	return m.fixes
}

// BenchmarkSourceCodeFixer benchmarks the source code fixing functionality
func BenchmarkSourceCodeFixer(b *testing.B) {
	sourceCode := `
function example() {
  const x = 1;
  const y = 2;
  const z = 3;
  return x + y + z;
}

function another() {
  let a = "hello";
  let b = "world";
  return a + " " + b;
}
`

	// Create mock diagnostics with fixes
	diagnostics := []mockLintMessage{
		{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(50, 55), // Replace "const x = 1"
					Text:  "let x = 1",
				},
			},
		},
		{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(65, 70), // Replace "const y = 2"
					Text:  "let y = 2",
				},
			},
		},
		{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(80, 85), // Replace "const z = 3"
					Text:  "let z = 3",
				},
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = ApplyRuleFixes(sourceCode, diagnostics)
	}
}

// BenchmarkSourceCodeFixerMultipleFixes benchmarks fixing with multiple fixes per diagnostic
func BenchmarkSourceCodeFixerMultipleFixes(b *testing.B) {
	sourceCode := `
class TestClass {
  method1() {
    console.log("test");
  }
  
  method2() {
    console.log("another test");
  }
}
`

	// Create diagnostics with multiple fixes each
	diagnostics := []mockLintMessage{
		{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(35, 42), // Replace "console"
					Text:  "window.console",
				},
				{
					Range: core.NewTextRange(43, 46), // Replace "log"
					Text:  "info",
				},
			},
		},
		{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(85, 92), // Replace "console"
					Text:  "window.console",
				},
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = ApplyRuleFixes(sourceCode, diagnostics)
	}
}

// BenchmarkSourceCodeFixerLargeFile benchmarks fixing a larger file
func BenchmarkSourceCodeFixerLargeFile(b *testing.B) {
	// Generate a larger source code file
	sourceCode := ""
	for i := 0; i < 100; i++ {
		sourceCode += `
function example` + string(rune('0'+(i%10))) + `() {
  const value = ` + string(rune('0'+(i%10))) + `;
  return value * 2;
}
`
	}

	// Create many diagnostics with fixes
	var diagnostics []mockLintMessage
	for i := 0; i < 50; i++ {
		offset := i * 80 // Approximate offset for each function
		diagnostics = append(diagnostics, mockLintMessage{
			fixes: []rule.RuleFix{
				{
					Range: core.NewTextRange(offset+30, offset+35), // Replace "const"
					Text:  "let",
				},
			},
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = ApplyRuleFixes(sourceCode, diagnostics)
	}
}