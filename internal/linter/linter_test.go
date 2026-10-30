package linter

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
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

func TestRunLinter_ProgramsSharingAFileParseItSeparately(t *testing.T) {
	// RunLinter clears the parse cache after each program, so a finished program's files are not kept alive by the
	// cache; a later program that includes the same file parses it again and must lint it the same way.
	dir := filepath.ToSlash(t.TempDir())
	write := func(name string, text string) string {
		path := dir + "/" + name
		assert.NilError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NilError(t, os.WriteFile(path, []byte(text), 0o644))
		return path
	}
	shared := write("shared.ts", "export const shared: Promise<number> = Promise.resolve(1);\n")
	programs := map[string][]string{}
	for _, name := range []string{"a", "b"} {
		config := write(name+"/tsconfig.json", `{"compilerOptions":{"strict":true,"target":"es2022","lib":["es2022"],"types":[]},"files":["../shared.ts"]}`)
		programs[config] = []string{shared}
	}

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic
	var internal []diagnostic.Internal
	err := RunLinter(RunLinterOptions{
		LogLevel:         utils.LogLevelNormal,
		CurrentDirectory: dir,
		Workload:         Workload{Programs: programs},
		Workers:          1,
		FS:               bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name: "declared-type",
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return rule.RuleListeners{
						ast.KindVariableDeclaration: func(node *ast.Node) {
							declared := ctx.TypeChecker.GetTypeAtLocation(node.Name())
							ctx.ReportNode(node, rule.RuleMessage{Id: "declaredType", Description: ctx.TypeChecker.TypeToString(declared)})
						},
					}
				},
			}}
		},
		OnRuleDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) { internal = append(internal, d) },
		Fixes:                Fixes{},
		TypeErrors:           TypeErrors{},
	})
	assert.NilError(t, err)
	assert.Equal(t, len(internal), 0)
	assert.Equal(t, len(diagnostics), 2, "one diagnostic per program")
	assert.Assert(t, diagnostics[0].SourceFile != diagnostics[1].SourceFile, "the second program parsed the shared file again")
	for _, d := range diagnostics {
		assert.Equal(t, d.SourceFile.FileName(), shared)
		assert.Equal(t, d.Message.Description, "Promise<number>")
		assert.Equal(t, d.Range, diagnostics[0].Range)
	}
}
