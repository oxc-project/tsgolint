package linter

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_condition"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unnecessary_type_assertion"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_assignment"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_unsafe_call"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

// TestContentMappedDiagnostics lints a file transformed by testdata/contentmapper's mapper, which runs
// as a real child process over the content mapper protocol. The mapper wraps the file's text in
// scaffolding, so this covers the whole path: spawning the mapper, parsing its output into the program,
// and reporting diagnostics against the original file rather than the virtual TypeScript.
func TestContentMappedDiagnostics(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is required to run a content mapper")
	}
	if !utils.ContentMappersEnabled() {
		t.Skip(utils.ContentMappersDisabledEnvVar + " is set")
	}
	t.Cleanup(utils.ShutdownContentMappers)

	rootDir, err := filepath.Abs(filepath.Join("testdata", "contentmapper"))
	assert.NilError(t, err)
	rootDir = tspath.NormalizePath(rootDir)
	mappedFile := tspath.ResolvePath(rootDir, "src/mapped.ext")

	host := utils.CreateCompilerHost(rootDir, cachedBaseFS)
	program, internalDiags, err := utils.CreateProgram(true, cachedBaseFS, rootDir, "tsconfig.json", host, false)
	assert.NilError(t, err)
	assert.Equal(t, len(internalDiags), 0, "unexpected program diagnostics: %v", internalDiags)

	sourceFile := program.GetSourceFile(mappedFile)
	assert.Assert(t, sourceFile != nil, "mapped file is not in the program")
	assert.Assert(t, utils.IsContentMapped(sourceFile), "mapped file was not transformed by the mapper")

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic
	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    []*ast.SourceFile{sourceFile},
		Workers:  1,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			rules := []rule.Rule{
				no_unnecessary_condition.NoUnnecessaryConditionRule,
				no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
				no_unsafe_assignment.NoUnsafeAssignmentRule,
				no_unsafe_call.NoUnsafeCallRule,
			}
			configured := make([]ConfiguredRule, len(rules))
			for i, r := range rules {
				configured[i] = ConfiguredRule{
					Name: r.Name,
					Run:  func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}
			}
			return configured
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(diagnostic.Internal) {},
		Fixes:                Fixes{Fix: true},
	})
	assert.NilError(t, err)

	sort.Slice(diagnostics, func(i, j int) bool { return diagnostics[i].Range.Pos() < diagnostics[j].Range.Pos() })

	original := sourceFile.OriginalText()
	got := make([]string, len(diagnostics))
	for i, d := range diagnostics {
		got[i] = d.RuleName + ": " + original[d.Range.Pos():d.Range.End()]
	}

	// Only the verbatim body is reported on: the unmapped `JSON.parse` assignment (no-unsafe-assignment)
	// and the `__scaffold__.trailing()` call anchored on a zero-length original range (no-unsafe-call)
	// are both mapper scaffolding, and are dropped.
	assert.DeepEqual(t, got, []string{
		"no-unnecessary-condition: always",
		"no-unnecessary-type-assertion: as number",
	})

	// A fix inside the verbatim body maps back exactly, so it stays applicable to the original file.
	var fixes []rule.RuleFix
	for _, d := range diagnostics {
		if d.RuleName == "no-unnecessary-type-assertion" {
			fixes = d.Fixes()
		}
	}
	assert.Equal(t, len(fixes), 1)
	assert.Equal(t, original[fixes[0].Range.Pos():fixes[0].Range.End()], " as number")
	assert.Equal(t, fixes[0].Text, "")
}

// TestContentMapperOwnDiagnostics checks that a diagnostic the mapper produced itself is reported at
// the offsets the mapper gave. Those are already original offsets, so mapping them a second time would
// move them.
func TestContentMapperOwnDiagnostics(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is required to run a content mapper")
	}
	if !utils.ContentMappersEnabled() {
		t.Skip(utils.ContentMappersDisabledEnvVar + " is set")
	}
	t.Cleanup(utils.ShutdownContentMappers)

	rootDir, err := filepath.Abs(filepath.Join("testdata", "contentmapper"))
	assert.NilError(t, err)
	rootDir = tspath.NormalizePath(rootDir)
	mappedFile := tspath.ResolvePath(rootDir, "src/mapper-diagnostic.ext")

	host := utils.CreateCompilerHost(rootDir, cachedBaseFS)
	program, _, err := utils.CreateProgram(true, cachedBaseFS, rootDir, "tsconfig.json", host, false)
	assert.NilError(t, err)

	sourceFile := program.GetSourceFile(mappedFile)
	assert.Assert(t, sourceFile != nil, "mapped file is not in the program")

	var mu sync.Mutex
	var internalDiags []diagnostic.Internal
	err = RunLinterOnProgram(RunLinterOnProgramOptions{
		LogLevel:        utils.LogLevelNormal,
		Program:         program,
		Files:           []*ast.SourceFile{sourceFile},
		Workers:         1,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		OnDiagnostic:    func(rule.RuleDiagnostic) {},
		OnInternalDiagnostic: func(d diagnostic.Internal) {
			mu.Lock()
			defer mu.Unlock()
			internalDiags = append(internalDiags, d)
		},
		TypeErrors: TypeErrors{ReportSyntactic: true},
	})
	assert.NilError(t, err)

	assert.Equal(t, len(internalDiags), 1)
	assert.Equal(t, internalDiags[0].Description, "test mapper syntax error")
	original := sourceFile.OriginalText()
	assert.Equal(t, original[internalDiags[0].Range.Pos():internalDiags[0].Range.End()], "@@mapper-error@@")
}

// TestUnmappedExtensionIsReported covers a file whose extension no content mapper claims — an oxlint
// languageOptions.parser override pointing at a tsconfig that registers no mapper for it. It has to
// land on a diagnostic; the parser has no script kind for the file and used to panic on it.
func TestUnmappedExtensionIsReported(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	mappedFile := tspath.ResolvePath(rootDir, "component.gts")

	fs := utils.NewOverlayVFS(cachedBaseFS, map[string]string{
		mappedFile: "const a = 1;\n<template>{{a}}</template>\n",
	})

	var mu sync.Mutex
	var internalDiags []diagnostic.Internal
	err := RunLinter(RunLinterOptions{
		LogLevel:         utils.LogLevelNormal,
		CurrentDirectory: rootDir,
		FS:               fs,
		Workload:         Workload{Programs: map[string][]string{}, UnmatchedFiles: []string{mappedFile}},
		Workers:          1,
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return nil },
		OnRuleDiagnostic: func(rule.RuleDiagnostic) {},
		OnInternalDiagnostic: func(d diagnostic.Internal) {
			mu.Lock()
			defer mu.Unlock()
			internalDiags = append(internalDiags, d)
		},
	})
	assert.NilError(t, err)

	assert.Equal(t, len(internalDiags), 1)
	assert.Equal(t, internalDiags[0].Id, "unsupported-file-extension")
	assert.Equal(t, *internalDiags[0].FilePath, mappedFile)
	assert.Assert(t, strings.Contains(internalDiags[0].Help, ".gts"), "help should name the extension: %v", internalDiags[0].Help)
}
