package rule_tester

import (
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

type ValidTestCase struct {
	Code     string
	Only     bool
	Skip     bool
	Options  any
	TSConfig string
	Tsx      bool
}

type InvalidTestCaseError struct {
	MessageId   string
	Line        int
	Column      int
	EndLine     int
	EndColumn   int
	Suggestions []InvalidTestCaseSuggestion
}

type InvalidTestCaseSuggestion struct {
	MessageId string
	Output    string
}

type InvalidTestCase struct {
	Code     string
	Only     bool
	Skip     bool
	Output   []string
	Errors   []InvalidTestCaseError
	TSConfig string
	Options  any
	Tsx      bool
}

func RunRuleTester(rootDir string, tsconfigPath string, t *testing.T, r *rule.Rule, validTestCases []ValidTestCase, invalidTestCases []InvalidTestCase) {
	onlyMode := slices.ContainsFunc(validTestCases, func(c ValidTestCase) bool { return c.Only }) ||
		slices.ContainsFunc(invalidTestCases, func(c InvalidTestCase) bool { return c.Only })

	runLinter := func(t *testing.T, code string, options any, tsconfigPathOverride string, tsx bool) []rule.RuleDiagnostic {
		var diagnosticsMu sync.Mutex
		diagnostics := make([]rule.RuleDiagnostic, 0, 3)

		fileName := "file.ts"
		if tsx {
			fileName = "react.tsx"
		}

		fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
		host := utils.CreateCompilerHost(rootDir, fs)

		tsconfigPath := tsconfigPath
		if tsconfigPathOverride != "" {
			tsconfigPath = tsconfigPathOverride
		}

		program, err := utils.CreateProgram(true, fs, rootDir, tsconfigPath, host)
		assert.NilError(t, err, "couldn't create program. code: "+code)

		files := []*ast.SourceFile{program.GetSourceFile(fileName)}

		err = linter.RunLinterOnProgram(
			utils.LogLevelNormal,
			program,
			files,
			1,
			func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{
					{
						Name: "test",
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return r.Run(ctx, options)
						},
					},
				}
			},
			func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				defer diagnosticsMu.Unlock()

				diagnostics = append(diagnostics, diagnostic)
			},
		)

		assert.NilError(t, err, "error running linter. code:\n", code)

		return diagnostics
	}

	for i, testCase := range validTestCases {
		t.Run("valid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if (onlyMode && !testCase.Only) || testCase.Skip {
				t.SkipNow()
			}

			diagnostics := runLinter(t, testCase.Code, testCase.Options, testCase.TSConfig, testCase.Tsx)
			if len(diagnostics) != 0 {
				// TODO: pretty errors
				t.Errorf("Expected valid test case not to contain errors. Code:\n%v", testCase.Code)
				for i, d := range diagnostics {
					t.Errorf("error %v - (%v-%v) %v", i+1, d.Range.Pos(), d.Range.End(), d.Message.Description)
				}
				t.FailNow()
			}
		})
	}

	for i, testCase := range invalidTestCases {
		t.Run("invalid-"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if (onlyMode && !testCase.Only) || testCase.Skip {
				t.SkipNow()
			}

			var initialDiagnostics []rule.RuleDiagnostic
			outputs := make([]string, 0, 1)
			code := testCase.Code

			for i := range 10 {
				diagnostics := runLinter(t, code, testCase.Options, testCase.TSConfig, testCase.Tsx)
				if i == 0 {
					initialDiagnostics = diagnostics
				}

				fixedCode, _, fixed := linter.ApplyRuleFixes(code, diagnostics)

				if !fixed {
					break
				}
				code = fixedCode
				outputs = append(outputs, fixedCode)
			}

			if len(testCase.Output) == len(outputs) {
				for i, expected := range testCase.Output {
					assert.Equal(t, expected, outputs[i], "Expected code after fix")
				}
			} else {
				t.Errorf("Expected to have %v outputs but had %v: %v", len(testCase.Output), len(outputs), outputs)
			}

			if len(initialDiagnostics) != len(testCase.Errors) {
				t.Fatalf("Expected invalid test case to contain exactly %v errors (reported %v errors - %v). Code:\n%v", len(testCase.Errors), len(initialDiagnostics), initialDiagnostics, testCase.Code)
			}

			for i, expected := range testCase.Errors {
				diagnostic := initialDiagnostics[i]

				if expected.MessageId != diagnostic.Message.Id {
					t.Errorf("Invalid message id %v. Expected %v", diagnostic.Message.Id, expected.MessageId)
				}

				lineIndex, columnIndex := scanner.GetLineAndCharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
				line, column := lineIndex+1, columnIndex+1
				endLineIndex, endColumnIndex := scanner.GetLineAndCharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.End())
				endLine, endColumn := endLineIndex+1, endColumnIndex+1

				if expected.Line != 0 && expected.Line != line {
					t.Errorf("Error line should be %v. Got %v", expected.Line, line)
				}
				if expected.Column != 0 && expected.Column != column {
					t.Errorf("Error column should be %v. Got %v", expected.Column, column)
				}
				if expected.EndLine != 0 && expected.EndLine != endLine {
					t.Errorf("Error end line should be %v. Got %v", expected.EndLine, endLine)
				}
				if expected.EndColumn != 0 && expected.EndColumn != endColumn {
					t.Errorf("Error end column should be %v. Got %v", expected.EndColumn, endColumn)
				}

				suggestionsCount := 0
				if diagnostic.Suggestions != nil {
					suggestionsCount = len(*diagnostic.Suggestions)
				}
				if len(expected.Suggestions) != suggestionsCount {
					t.Errorf("Expected to have %v suggestions but had %v", len(expected.Suggestions), suggestionsCount)
				} else {
					for i, expectedSuggestion := range expected.Suggestions {
						suggestion := (*diagnostic.Suggestions)[i]
						if expectedSuggestion.MessageId != suggestion.Message.Id {
							t.Errorf("Invalid suggestion message id %v. Expected %v", suggestion.Message.Id, expectedSuggestion.MessageId)
						} else {
							output, _, _ := linter.ApplyRuleFixes(testCase.Code, []rule.RuleSuggestion{suggestion})

							assert.Equal(t, expectedSuggestion.Output, output, "Expected code after suggestion fix")
						}
					}
				}
			}
		})
	}
}
