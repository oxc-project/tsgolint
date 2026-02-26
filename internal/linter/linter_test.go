package linter

import (
	"sync"
	"testing"

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

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host)
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

	err = RunLinterOnProgram(
		utils.LogLevelNormal,
		program,
		sourceFiles,
		1,
		func(sourceFile *ast.SourceFile) []ConfiguredRule {
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
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		func(d diagnostic.Internal) {},
		Fixes{Fix: false, FixSuggestions: false},
		TypeErrors{ReportSyntactic: false, ReportSemantic: false},
		nil,
	)
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

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host)
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

	err = RunLinterOnProgram(
		utils.LogLevelNormal,
		program,
		sourceFiles,
		1,
		func(sourceFile *ast.SourceFile) []ConfiguredRule {
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
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		func(d diagnostic.Internal) {},
		Fixes{Fix: false, FixSuggestions: false},
		TypeErrors{ReportSyntactic: false, ReportSemantic: false},
		nil,
	)
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

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host)
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

	err = RunLinterOnProgram(
		utils.LogLevelNormal,
		program,
		sourceFiles,
		1,
		func(sourceFile *ast.SourceFile) []ConfiguredRule {
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
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		func(d diagnostic.Internal) {},
		Fixes{Fix: false, FixSuggestions: false},
		TypeErrors{ReportSyntactic: false, ReportSemantic: false},
		nil,
	)
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
