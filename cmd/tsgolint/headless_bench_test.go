package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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

	program, diags, err := utils.CreateProgram(singleThreaded, fs, dir, tsconfigPath, host, false)
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

// BenchmarkE2E benchmarks the true end-to-end path for all fixture files:
// FS creation, tsconfig resolution, program creation, linting with all rules,
// and diagnostic serialization via RunLinter. This measures the full cost that
// a real oxlint invocation would pay, including tsconfig discovery.
// File collection is excluded from the timed section since oxlint handles that.
func BenchmarkE2E(b *testing.B) {
	b.Helper()
	b.ReportAllocs()

	dir := fixtureDir
	baseFS := osvfs.FS()

	// Collect all .ts files in the fixtures directory (simulates what oxlint sends us).
	var allFiles []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".d.ts") {
			allFiles = append(allFiles, tspath.NormalizeSlashes(path))
		}
		return nil
	})
	if len(allFiles) == 0 {
		b.Fatal("no .ts fixture files found")
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

	// buildWorkload creates a fresh FS, resolves tsconfigs for all files, and
	// returns the workload and FS needed by RunLinter.
	buildWorkload := func() (linter.Workload, vfs.FS) {
		fs := bundled.WrapFS(cachedvfs.From(baseFS))
		resolver := utils.NewTsConfigResolver(fs, dir)
		result := resolver.FindTsConfigParallel(allFiles)

		workload := linter.Workload{
			Programs:       make(map[string][]string),
			UnmatchedFiles: []string{},
		}
		for file, tsconfig := range result {
			if tsconfig == "" {
				workload.UnmatchedFiles = append(workload.UnmatchedFiles, file)
			} else {
				workload.Programs[tsconfig] = append(workload.Programs[tsconfig], file)
			}
		}
		return workload, fs
	}

	workers := runtime.GOMAXPROCS(0)

	// Warm up once to verify everything works.
	{
		workload, fs := buildWorkload()
		if len(workload.Programs) == 0 && len(workload.UnmatchedFiles) == 0 {
			b.Fatal("no files resolved to any program")
		}

		var diagnosticCount int64
		err := linter.RunLinter(
			utils.LogLevelNormal,
			dir,
			workload,
			workers,
			fs,
			getRulesForFile,
			func(_ rule.RuleDiagnostic) { atomic.AddInt64(&diagnosticCount, 1) },
			func(_ diagnostic.Internal) {},
			linter.Fixes{},
			linter.TypeErrors{},
			false,
		)
		if err != nil {
			b.Fatal("warmup linter failed:", err)
		}
		b.Logf("files: %d, programs: %d, diagnostics: %d, workers: %d",
			len(allFiles), len(workload.Programs), diagnosticCount, workers)
	}

	b.ResetTimer()
	for b.Loop() {
		// Full end-to-end: fresh FS, tsconfig resolution, program creation, lint,
		// and diagnostic serialization to io.Discard.
		workload, fs := buildWorkload()

		err := linter.RunLinter(
			utils.LogLevelNormal,
			dir,
			workload,
			workers,
			fs,
			getRulesForFile,
			func(_ rule.RuleDiagnostic) {},
			func(_ diagnostic.Internal) {},
			linter.Fixes{},
			linter.TypeErrors{},
			false,
		)
		if err != nil {
			b.Fatal("linter failed:", err)
		}
	}
}
