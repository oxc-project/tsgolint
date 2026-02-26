package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// fixtureDir is the path to the e2e/fixtures/basic directory, relative to the
// cmd/tsgolint package (two levels up from the repo root).
var fixtureDir = func() string {
	abs, err := filepath.Abs(filepath.Join("..", "..", "e2e", "fixtures", "basic"))
	if err != nil {
		panic(err)
	}
	return abs
}()

type benchmarkEnv struct {
	files           []*ast.SourceFile
	program         *compiler.Program
	getRulesForFile func(_ *ast.SourceFile) []linter.ConfiguredRule
}

func setupBenchmarkEnv(b *testing.B, singleThreaded bool) benchmarkEnv {
	b.Helper()

	dir := fixtureDir
	tsconfigPath := filepath.Join(dir, "tsconfig.json")

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)

	program, diags, err := utils.CreateProgram(singleThreaded, fs, dir, tsconfigPath, host)
	if err != nil {
		b.Fatal("failed to create program:", err)
	}
	if len(diags) > 0 {
		b.Fatal("tsconfig diagnostics:", diags[0].Description)
	}

	// Collect all source files under the fixture directory (skip node_modules/lib files).
	var files []*ast.SourceFile
	prefix := string(tspath.ToPath("", dir, fs.UseCaseSensitiveFileNames()).EnsureTrailingDirectorySeparator())
	for _, sf := range program.SourceFiles() {
		if strings.HasPrefix(string(sf.Path()), prefix) {
			files = append(files, sf)
		}
	}
	if len(files) == 0 {
		b.Fatal("no source files found in fixture directory")
	}

	getRulesForFile := func(_ *ast.SourceFile) []linter.ConfiguredRule {
		rules := make([]linter.ConfiguredRule, len(allRules))
		for i, r := range allRules {
			rules[i] = linter.ConfiguredRule{
				Name: r.Name,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return r.Run(ctx, nil)
				},
			}
		}
		return rules
	}

	return benchmarkEnv{
		files:           files,
		program:         program,
		getRulesForFile: getRulesForFile,
	}
}

func runAllRulesBenchmark(b *testing.B, singleThreaded bool) {
	b.Helper()
	b.ReportAllocs()

	env := setupBenchmarkEnv(b, singleThreaded)
	workers := runtime.GOMAXPROCS(0)
	if singleThreaded {
		workers = 1
	}

	// Warm up: run once to ensure everything is initialized
	var diagnosticCount int64
	err := linter.RunLinterOnProgram(
		utils.LogLevelNormal,
		env.program,
		env.files,
		workers,
		env.getRulesForFile,
		func(_ rule.RuleDiagnostic) { atomic.AddInt64(&diagnosticCount, 1) },
		func(_ diagnostic.Internal) {},
		linter.Fixes{},
		linter.TypeErrors{},
		nil, // stats
	)
	if err != nil {
		b.Fatal("warmup linter failed:", err)
	}
	if diagnosticCount == 0 {
		b.Fatal("no diagnostics were emitted, expected at least one")
	}

	b.ResetTimer()
	for b.Loop() {
		err := linter.RunLinterOnProgram(
			utils.LogLevelNormal,
			env.program,
			env.files,
			workers,
			env.getRulesForFile,
			func(_ rule.RuleDiagnostic) {},
			func(_ diagnostic.Internal) {},
			linter.Fixes{},
			linter.TypeErrors{},
			nil,
		)
		if err != nil {
			b.Fatal("linter failed:", err)
		}
	}
}

// BenchmarkAllRulesHeadless benchmarks running all rules in headless mode on a single file. This should be
// somewhat correlated to real-world performance, minus the overhead for things like program creation and streaming
// data back to oxlint.
func BenchmarkAllRulesHeadless(b *testing.B) {
	runAllRulesBenchmark(b, false)
}

// BenchmarkAllRulesHeadlessSingleThread benchmarks with a single worker to measure per-core throughput.
func BenchmarkAllRulesHeadlessSingleThread(b *testing.B) {
	runAllRulesBenchmark(b, true)
}
