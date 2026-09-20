package rule_tester

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type BenchmarkTestCase struct {
	Name                string
	Code                string
	FileName            string
	Options             any
	TSConfig            string
	Tsx                 bool
	Files               map[string]string
	ExpectedDiagnostics int
}

// RunRuleBenchmark benchmarks one rule against independently prepared TypeScript programs.
// Program creation is excluded so results focus on rule execution and AST traversal.
func RunRuleBenchmark(rootDir string, tsconfigPath string, b *testing.B, r *rule.Rule, testCases []BenchmarkTestCase) {
	b.Helper()

	for _, testCase := range testCases {
		b.Run(testCase.Name, func(b *testing.B) {
			fileName := testCase.FileName
			if fileName == "" {
				fileName = "file.ts"
				if testCase.Tsx {
					fileName = "react.tsx"
				}
			}

			resolvedFileName := tspath.ResolvePath(rootDir, fileName)
			virtualFiles := map[string]string{resolvedFileName: testCase.Code}
			for relativePath, source := range testCase.Files {
				virtualFiles[tspath.ResolvePath(rootDir, relativePath)] = source
			}
			fs := utils.NewOverlayVFS(cachedBaseFS, virtualFiles)
			host := utils.CreateCompilerHost(rootDir, fs)

			caseTSConfigPath := tsconfigPath
			if testCase.TSConfig != "" {
				caseTSConfigPath = testCase.TSConfig
			}

			program, internalDiagnostics, err := utils.CreateProgram(true, fs, rootDir, caseTSConfigPath, host, false)
			if err != nil {
				b.Fatal("couldn't create program:", err)
			}
			if len(internalDiagnostics) > 0 {
				b.Fatalf("couldn't create program due to internal diagnostics: %+v", internalDiagnostics)
			}
			if program == nil {
				b.Fatal("couldn't create program")
			}

			sourceFile := program.GetSourceFile(fileName)
			if sourceFile == nil {
				sourceFile = program.GetSourceFile(resolvedFileName)
			}
			if sourceFile == nil {
				programFiles := make([]string, 0, len(program.SourceFiles()))
				for _, sourceFile := range program.SourceFiles() {
					programFiles = append(programFiles, sourceFile.FileName())
				}
				slices.Sort(programFiles)
				b.Fatalf("couldn't get source file %s (resolved: %s); program source files (%s): %s", fileName, resolvedFileName, strconv.Itoa(len(programFiles)), strings.Join(programFiles, ", "))
			}

			configuredRule := linter.ConfiguredRule{
				Name: r.Name,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return r.Run(ctx, testCase.Options)
				},
			}
			run := func(onDiagnostic func(rule.RuleDiagnostic)) error {
				return linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
					LogLevel: utils.LogLevelNormal,
					Program:  program,
					Files:    []*ast.SourceFile{sourceFile},
					Workers:  1,
					GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
						return []linter.ConfiguredRule{configuredRule}
					},
					OnDiagnostic:         onDiagnostic,
					OnInternalDiagnostic: func(diagnostic.Internal) {},
					Fixes: linter.Fixes{
						Fix:            true,
						FixSuggestions: true,
					},
				})
			}

			diagnosticCount := 0
			if err := run(func(rule.RuleDiagnostic) { diagnosticCount++ }); err != nil {
				b.Fatal("warmup linter failed:", err)
			}
			if diagnosticCount != testCase.ExpectedDiagnostics {
				b.Fatalf("expected %d diagnostics, got %d", testCase.ExpectedDiagnostics, diagnosticCount)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := run(func(rule.RuleDiagnostic) {}); err != nil {
					b.Fatal("linter failed:", err)
				}
			}
		})
	}
}
