package main

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-json-experiment/json"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/no_misused_promises"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// BenchmarkNoMisusedPromises runs ONLY the no-misused-promises rule over a
// corpus directory (set NMP_CORPUS to override; defaults to the basic fixtures).
//
// A FRESH program is created each iteration so the checker's type cache is COLD.
// A real lint run touches each file (and thus each type) once, so the cold path
// is what matters; reusing one program across iterations warms the cache and
// hides the type-query savings. Program-creation cost is identical before/after,
// so the before/after delta cleanly isolates the rule's cold-path cost.
func BenchmarkNoMisusedPromises(b *testing.B) {
	b.ReportAllocs()

	dir := os.Getenv("NMP_CORPUS")
	if dir == "" {
		dir = fixtureDir
	}
	tsconfigPath := filepath.Join(dir, "tsconfig.json")

	// NMP_OPTS is a JSON object of rule options (empty/unset = rule defaults).
	// Used to measure the cost of individual config options empirically.
	var ruleOpts any
	if raw := os.Getenv("NMP_OPTS"); raw != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			b.Fatalf("bad NMP_OPTS: %v", err)
		}
		ruleOpts = m
	}

	getRulesForFile := func(_ *ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{
			{
				Name: no_misused_promises.NoMisusedPromisesRule.Name,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_misused_promises.NoMisusedPromisesRule.Run(ctx, ruleOpts)
				},
			},
		}
	}

	buildRun := func() func() int64 {
		fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := utils.CreateCompilerHost(dir, fs)
		program, diags, err := utils.CreateProgram(false, fs, dir, tsconfigPath, host, false)
		if err != nil {
			b.Fatal("failed to create program:", err)
		}
		if len(diags) > 0 {
			b.Fatal("tsconfig diagnostics:", diags[0].Description)
		}
		var files []*ast.SourceFile
		prefix := string(tspath.ToPath("", dir, fs.UseCaseSensitiveFileNames()).EnsureTrailingDirectorySeparator())
		for _, sf := range program.SourceFiles() {
			if strings.HasPrefix(string(sf.Path()), prefix) {
				files = append(files, sf)
			}
		}
		if len(files) == 0 {
			b.Fatal("no source files found in corpus directory")
		}
		return func() int64 {
			var count int64
			err := linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
				LogLevel:             utils.LogLevelNormal,
				Program:              program,
				Files:                files,
				Workers:              runtime.GOMAXPROCS(0),
				GetRulesForFile:      getRulesForFile,
				OnDiagnostic:         func(_ rule.RuleDiagnostic) { atomic.AddInt64(&count, 1) },
				OnInternalDiagnostic: func(_ diagnostic.Internal) {},
			})
			if err != nil {
				b.Fatal("linter failed:", err)
			}
			return count
		}
	}

	// Disable GC so a collection of the previous ~50MB program never lands
	// inside a timed run() and adds noise. We reclaim it manually (below) while
	// the timer is stopped, keeping only one program live at a time.
	prevGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGC)

	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		run := buildRun() // fresh cold program, not timed
		runtime.GC()      // reclaim the previous iteration's program now, off the clock
		b.StartTimer()
		run()
	}
}
