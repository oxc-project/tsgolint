package linter

import (
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

// TestUnsupportedExtensionIsReported covers a file whose extension TypeScript cannot parse. It
// belongs to no tsconfig, so it reaches the inferred project, where the parser has no script kind for
// it and used to panic. It has to land on a diagnostic instead.
func TestUnsupportedExtensionIsReported(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	unparsableFile := tspath.ResolvePath(rootDir, "component.gts")

	fs := utils.NewOverlayVFS(cachedBaseFS, map[string]string{
		unparsableFile: "const a = 1;\n<template>{{a}}</template>\n",
	})

	var mu sync.Mutex
	var internalDiags []diagnostic.Internal
	err := RunLinter(RunLinterOptions{
		LogLevel:         utils.LogLevelNormal,
		CurrentDirectory: rootDir,
		FS:               fs,
		Workload:         Workload{Programs: map[string][]string{}, UnmatchedFiles: []string{unparsableFile}},
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
	assert.Equal(t, *internalDiags[0].FilePath, unparsableFile)
	assert.Assert(t, strings.Contains(internalDiags[0].Help, ".gts"), "help should name the extension: %v", internalDiags[0].Help)
}
