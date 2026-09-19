package linter

import (
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

var cachedBaseFS = cachedvfs.From(bundled.WrapFS(osvfs.FS()))

func TestRunLinterOnProgram_MultipleListenersDiagnostics(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	filePath := tspath.ResolvePath(rootDir, fileName)
	code := `
function greet(name: string): string {
	const greeting = "hello";
	return greeting + " " + name;
}
`

	fs := utils.NewOverlayVFS(
		cachedBaseFS,
		map[string]string{filePath: code},
	)
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	sourceFiles := []*ast.SourceFile{program.GetSourceFile(filePath)}

	const ruleName = "multi-listener-rule"
	funcMessage := rule.RuleMessage{
		Id:          "noFunction",
		Description: "Found a function declaration",
	}
	varMessage := rule.RuleMessage{
		Id:          "noVariable",
		Description: "Found a variable statement",
	}

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic

	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    sourceFiles,
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{
				{
					Name: ruleName,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{
							ast.KindFunctionDeclaration: func(node *ast.Node) {
								ctx.ReportNode(node, funcMessage)
							},
							ast.KindVariableStatement: func(node *ast.Node) {
								ctx.ReportNode(node, varMessage)
							},
						}
					},
				},
			}
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) {},
		Fixes:                Fixes{Fix: false, FixSuggestions: false},
		TypeErrors:           TypeErrors{ReportSyntactic: false, ReportSemantic: false},
	})
	assert.NilError(t, err, "unexpected error from RunLinterOnProgram")

	assert.Equal(t, len(diagnostics), 2, "expected exactly two diagnostics")

	// Both diagnostics should have the same rule name and source file
	for i, d := range diagnostics {
		assert.Equal(t, d.RuleName, ruleName, "diagnostic %d should have the correct rule name", i)
		assert.Assert(t, d.SourceFile != nil, "diagnostic %d should have a non-nil source file", i)
		assert.Equal(t, d.SourceFile.FileName(), filePath, "diagnostic %d source file should match the input file", i)
	}

	// Verify each listener produced the expected diagnostic (order matches AST traversal)
	assert.Equal(t, diagnostics[0].Message.Id, funcMessage.Id, "first diagnostic should be from the function listener")
	assert.Equal(t, diagnostics[1].Message.Id, varMessage.Id, "second diagnostic should be from the variable listener")
}

func TestRunLinterOnProgram_MultipleRules(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	filePath := tspath.ResolvePath(rootDir, fileName)
	code := `
const x: number = 1;
function add(a: number, b: number): number {
	return a + b;
}
`

	fs := utils.NewOverlayVFS(
		cachedBaseFS,
		map[string]string{filePath: code},
	)
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	sourceFiles := []*ast.SourceFile{program.GetSourceFile(filePath)}

	const ruleA = "no-variables"
	const ruleB = "no-functions"
	msgA := rule.RuleMessage{
		Id:          "noVar",
		Description: "Variable statements are not allowed",
	}
	msgB := rule.RuleMessage{
		Id:          "noFunc",
		Description: "Function declarations are not allowed",
	}

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic

	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    sourceFiles,
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{
				{
					Name: ruleA,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{
							ast.KindVariableStatement: func(node *ast.Node) {
								ctx.ReportNode(node, msgA)
							},
						}
					},
				},
				{
					Name: ruleB,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{
							ast.KindFunctionDeclaration: func(node *ast.Node) {
								ctx.ReportNode(node, msgB)
							},
						}
					},
				},
			}
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) {},
		Fixes:                Fixes{Fix: false, FixSuggestions: false},
		TypeErrors:           TypeErrors{ReportSyntactic: false, ReportSemantic: false},
	})
	assert.NilError(t, err, "unexpected error from RunLinterOnProgram")

	assert.Equal(t, len(diagnostics), 2, "expected exactly two diagnostics")

	// All diagnostics should reference the correct source file
	for i, d := range diagnostics {
		assert.Assert(t, d.SourceFile != nil, "diagnostic %d should have a non-nil source file", i)
		assert.Equal(t, d.SourceFile.FileName(), filePath, "diagnostic %d source file should match the input file", i)
	}

	// Each diagnostic should be attributed to the correct rule.
	// The variable statement appears first in the source, so ruleA fires first,
	// then the function declaration triggers ruleB.
	assert.Equal(t, diagnostics[0].RuleName, ruleA, "first diagnostic should come from ruleA")
	assert.Equal(t, diagnostics[0].Message.Id, msgA.Id, "first diagnostic should have ruleA's message id")

	assert.Equal(t, diagnostics[1].RuleName, ruleB, "second diagnostic should come from ruleB")
	assert.Equal(t, diagnostics[1].Message.Id, msgB.Id, "second diagnostic should have ruleB's message id")
}

func TestRunLinterOnProgram_DiagnosticsEmittedInRun(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	filePath := tspath.ResolvePath(rootDir, fileName)
	code := `const x = 1;`

	fs := utils.NewOverlayVFS(
		cachedBaseFS,
		map[string]string{filePath: code},
	)
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	sourceFiles := []*ast.SourceFile{program.GetSourceFile(filePath)}

	const ruleA = "file-checker-a"
	const ruleB = "file-checker-b"
	msgA := rule.RuleMessage{
		Id:          "fileCheckA",
		Description: "Rule A checked this file",
	}
	msgB := rule.RuleMessage{
		Id:          "fileCheckB",
		Description: "Rule B checked this file",
	}

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic

	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    sourceFiles,
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{
				{
					Name: ruleA,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						// Emit a diagnostic directly in Run, without any listeners
						ctx.ReportDiagnostic(rule.RuleDiagnostic{
							Message: msgA,
						})
						return rule.RuleListeners{}
					},
				},
				{
					Name: ruleB,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						// Emit a diagnostic directly in Run, without any listeners
						ctx.ReportDiagnostic(rule.RuleDiagnostic{
							Message: msgB,
						})
						return rule.RuleListeners{}
					},
				},
			}
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) {},
		Fixes:                Fixes{Fix: false, FixSuggestions: false},
		TypeErrors:           TypeErrors{ReportSyntactic: false, ReportSemantic: false},
	})
	assert.NilError(t, err, "unexpected error from RunLinterOnProgram")

	assert.Equal(t, len(diagnostics), 2, "expected exactly two diagnostics")

	// Both diagnostics should have the correct source file set by the linter
	for i, d := range diagnostics {
		assert.Assert(t, d.SourceFile != nil, "diagnostic %d should have a non-nil source file", i)
		assert.Equal(t, d.SourceFile.FileName(), filePath, "diagnostic %d source file should match the input file", i)
	}

	// Verify each diagnostic is attributed to the correct rule
	assert.Equal(t, diagnostics[0].RuleName, ruleA, "first diagnostic should come from ruleA")
	assert.Equal(t, diagnostics[0].Message.Id, msgA.Id, "first diagnostic should have ruleA's message")

	assert.Equal(t, diagnostics[1].RuleName, ruleB, "second diagnostic should come from ruleB")
	assert.Equal(t, diagnostics[1].Message.Id, msgB.Id, "second diagnostic should have ruleB's message")
}

func TestRunLinterOnProgramWithTimings(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	filePath := tspath.ResolvePath(rootDir, fileName)
	code := `
const x = 1;
function greet() {
	return x;
}
`

	fs := utils.NewOverlayVFS(
		cachedBaseFS,
		map[string]string{filePath: code},
	)
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	sourceFiles := []*ast.SourceFile{program.GetSourceFile(filePath)}

	const ruleA = "timed-variable-rule"
	const ruleB = "timed-function-rule"
	const ruleC = "timed-run-only-rule"

	timingStore := NewRuleTimingStore()
	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    sourceFiles,
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{
				{
					Name: ruleA,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						time.Sleep(time.Microsecond)
						return rule.RuleListeners{
							ast.KindVariableStatement: func(*ast.Node) {
								time.Sleep(time.Microsecond)
							},
						}
					},
				},
				{
					Name: ruleB,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{
							ast.KindFunctionDeclaration: func(*ast.Node) {
								time.Sleep(time.Microsecond)
							},
						}
					},
				},
				{
					Name: ruleC,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						time.Sleep(time.Microsecond)
						return rule.RuleListeners{}
					},
				},
			}
		},
		OnDiagnostic:         func(d rule.RuleDiagnostic) {},
		OnInternalDiagnostic: func(d diagnostic.Internal) {},
		Fixes:                Fixes{Fix: false, FixSuggestions: false},
		TypeErrors:           TypeErrors{ReportSyntactic: false, ReportSemantic: false},
		TimingStore:          timingStore,
	})
	assert.NilError(t, err, "unexpected error from RunLinterOnProgramWithTimings")

	records := timingStore.Collect()
	assert.Equal(t, len(records), 3, "expected timings for each configured rule")

	recordsByRule := make(map[string]RuleTimingRecord, len(records))
	for _, record := range records {
		recordsByRule[record.RuleName] = record
		assert.Assert(t, record.Duration > 0, "timing for %s should record a positive duration", record.RuleName)
	}

	assert.Equal(t, recordsByRule[ruleA].Calls, uint64(2), "rule A should count Run plus its variable listener")
	assert.Equal(t, recordsByRule[ruleB].Calls, uint64(2), "rule B should count Run plus its function listener")
	assert.Equal(t, recordsByRule[ruleC].Calls, uint64(1), "rule C should count its Run call")
}

// setupTypeErrorProgram creates a program containing two files that each have a
// semantic error, and returns the source files in the given order.
func setupTypeErrorProgram(t *testing.T) (*compiler.Program, []*ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	filePathA := tspath.ResolvePath(rootDir, "type-errors-a.ts")
	filePathB := tspath.ResolvePath(rootDir, "type-errors-b.ts")

	fs := utils.NewOverlayVFS(
		cachedBaseFS,
		map[string]string{
			filePathA: "const a: number = \"not a number\";\n",
			filePathB: "const b: number = \"not a number\";\n",
		},
	)
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	sourceFileA := program.GetSourceFile(filePathA)
	assert.Assert(t, sourceFileA != nil, "expected %s to be in the program", filePathA)
	sourceFileB := program.GetSourceFile(filePathB)
	assert.Assert(t, sourceFileB != nil, "expected %s to be in the program", filePathB)

	return program, []*ast.SourceFile{sourceFileA, sourceFileB}
}

// runLinterCollectingInternalDiagnostics lints the given files and returns the
// file names the internal diagnostics were reported for.
func runLinterCollectingInternalDiagnostics(t *testing.T, program *compiler.Program, files []*ast.SourceFile, typeErrors TypeErrors) []string {
	t.Helper()

	var mu sync.Mutex
	var fileNames []string

	err := RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    files,
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return nil
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {},
		OnInternalDiagnostic: func(d diagnostic.Internal) {
			mu.Lock()
			defer mu.Unlock()
			if d.FilePath != nil {
				fileNames = append(fileNames, *d.FilePath)
			}
		},
		Fixes:      Fixes{Fix: false, FixSuggestions: false},
		TypeErrors: typeErrors,
	})
	assert.NilError(t, err, "unexpected error from RunLinterOnProgram")

	return fileNames
}

func TestRunLinterOnProgram_TypeErrorsReportedForEveryFileByDefault(t *testing.T) {
	program, files := setupTypeErrorProgram(t)

	// A nil callback reports the diagnostics for every linted file.
	fileNames := runLinterCollectingInternalDiagnostics(t, program, files, TypeErrors{
		ReportSemantic: true,
	})

	reportedFiles := make(map[string]struct{}, len(fileNames))
	for _, fileName := range fileNames {
		reportedFiles[fileName] = struct{}{}
	}

	assert.Equal(t, len(reportedFiles), 2, "expected type errors for both files, got %v", fileNames)
	for _, file := range files {
		_, ok := reportedFiles[file.FileName()]
		assert.Assert(t, ok, "expected a type error for %s, got %v", file.FileName(), fileNames)
	}
}

func TestRunLinterOnProgram_TypeErrorsReportedForAcceptedFilesOnly(t *testing.T) {
	program, files := setupTypeErrorProgram(t)

	fileNames := runLinterCollectingInternalDiagnostics(t, program, files, TypeErrors{
		ReportSemantic: true,
		ReportTypeErrorsForFile: func(sourceFile *ast.SourceFile) bool {
			return sourceFile == files[0]
		},
	})

	assert.Assert(t, len(fileNames) > 0, "expected type errors for the accepted file")
	for _, fileName := range fileNames {
		assert.Equal(t, fileName, files[0].FileName(), "type errors should only be reported for the accepted file")
	}
}

func TestRunLinterOnProgram_TypeErrorsRejectedForEveryFile(t *testing.T) {
	program, files := setupTypeErrorProgram(t)

	// A callback rejecting every file reports nothing, the behavior of a
	// payload opting every config group out of type checking.
	fileNames := runLinterCollectingInternalDiagnostics(t, program, files, TypeErrors{
		ReportSemantic:  true,
		ReportSyntactic: true,
		ReportTypeErrorsForFile: func(sourceFile *ast.SourceFile) bool {
			return false
		},
	})

	assert.Equal(t, len(fileNames), 0, "expected no type errors, got %v", fileNames)
}
