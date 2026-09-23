package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// These opt-in helpers run identically in baseline and candidate test binaries.
// See benchmarks/checker-affinity/README.md. Each measurement needs a fresh
// process: utils' parsed-source cache is global, even across new Programs.
type affinityCase struct {
	Root     string   `json:"root"`
	Config   string   `json:"config"`
	Files    []string `json:"files"`
	Rules    []string `json:"rules"`
	Semantic bool     `json:"semantic"`
	Warm     bool     `json:"warm"`
	Workers  int      `json:"workers"`
	Mode     string   `json:"mode"`
}

func readAffinityCase(tb testing.TB) affinityCase {
	tb.Helper()
	name := os.Getenv("TSGOLINT_AFFINITY_CASE")
	if name == "" {
		tb.Skip("set TSGOLINT_AFFINITY_CASE to an experiment case JSON file")
	}
	data, err := os.ReadFile(name)
	if err != nil {
		tb.Fatal(err)
	}
	var c affinityCase
	if err := json.Unmarshal(data, &c); err != nil {
		tb.Fatal(err)
	}
	if c.Workers <= 0 {
		c.Workers = runtime.GOMAXPROCS(0)
	}
	if !filepath.IsAbs(c.Root) || !filepath.IsAbs(c.Config) {
		tb.Fatal("root and config must be absolute paths")
	}
	return c
}

func (c affinityCase) program(tb testing.TB) (*compiler.Program, []*ast.SourceFile) {
	tb.Helper()
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(c.Root, fs)
	program, diagnostics, err := utils.CreateProgram(false, fs, c.Root, c.Config, host, false)
	if err != nil || len(diagnostics) > 0 || program == nil {
		tb.Fatalf("create program: %v; diagnostics: %+v", err, diagnostics)
	}
	var files []*ast.SourceFile
	if c.Files != nil {
		for _, path := range c.Files {
			file := program.GetSourceFile(filepath.ToSlash(filepath.Join(c.Root, path)))
			if file == nil {
				tb.Fatalf("selected file missing from program: %s", path)
			}
			files = append(files, file)
		}
	} else {
		files = c.eligibleFiles(program)
	}
	if len(files) == 0 {
		tb.Fatal("no eligible files")
	}
	return program, files
}

func (c affinityCase) eligibleFiles(program *compiler.Program) []*ast.SourceFile {
	var files []*ast.SourceFile
	for _, file := range program.GetSourceFiles() {
		// Imported JSON belongs to the program but is not lintable source code.
		if file.ScriptKind == core.ScriptKindJSON {
			continue
		}
		path, err := filepath.Rel(c.Root, file.FileName())
		if err == nil && path != ".." && !strings.HasPrefix(path, ".."+string(filepath.Separator)) &&
			!strings.Contains(filepath.ToSlash(path), "node_modules/") && !file.IsDeclarationFile {
			files = append(files, file)
		}
	}
	slices.SortFunc(files, func(a, b *ast.SourceFile) int { return strings.Compare(a.FileName(), b.FileName()) })
	return files
}

func TestCheckerAffinityEligibleFiles(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())
	for name, content := range map[string]string{
		"tsconfig.json": `{"compilerOptions":{"module":"preserve","target":"esnext","resolveJsonModule":true,"types":[]},"files":["index.ts","types.d.ts"]}`,
		"index.ts":      `import data from "./package.json"; export const version = data.version;`,
		"package.json":  `{"version":"1.0.0"}`,
		"types.d.ts":    `declare const external: string;`,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	c := affinityCase{Root: root, Config: filepath.Join(root, "tsconfig.json")}
	program, files := c.program(t)
	if program.GetSourceFile(filepath.Join(root, "package.json")) == nil {
		t.Fatal("expected imported JSON in the program")
	}
	if len(files) != 1 || files[0].FileName() != filepath.Join(root, "index.ts") {
		t.Fatalf("expected only index.ts, got %d eligible files", len(files))
	}
}

func (c affinityCase) rules(tb testing.TB) []linter.ConfiguredRule {
	tb.Helper()
	names := c.Rules
	if len(names) == 0 {
		for _, r := range allRules {
			names = append(names, r.Name)
		}
	}
	rules := make([]linter.ConfiguredRule, 0, len(names))
	for _, name := range names {
		r, ok := allRulesByName[name]
		if !ok {
			tb.Fatalf("unknown rule: %s", name)
		}
		rules = append(rules, linter.ConfiguredRule{Name: name, Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return r.Run(ctx, nil)
		}})
	}
	return rules
}

func (c affinityCase) warm(program *compiler.Program) {
	if c.Warm {
		program.GetSemanticDiagnosticsWithoutNoEmitFiltering(
			core.WithRequestID(context.Background(), "__single_run__"), c.eligibleFiles(program))
	}
}

func BenchmarkCheckerAffinity(b *testing.B) {
	b.StopTimer()
	c := readAffinityCase(b)
	if b.N != 1 {
		b.Fatal("use -benchtime=1x and a fresh process per sample")
	}
	rules := c.rules(b)
	var ruleCount, typeCount atomic.Int64
	b.ReportAllocs()
	if !c.Warm {
		b.StartTimer()
	}
	program, files := c.program(b)
	c.warm(program)
	if c.Warm {
		b.StartTimer()
	}
	err := linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
		Program: program, Files: files, Workers: c.Workers,
		GetRulesForFile:      func(*ast.SourceFile) []linter.ConfiguredRule { return rules },
		TypeErrors:           linter.TypeErrors{ReportSemantic: c.Semantic && !c.Warm},
		OnDiagnostic:         func(rule.RuleDiagnostic) { ruleCount.Add(1) },
		OnInternalDiagnostic: func(diagnostic.Internal) { typeCount.Add(1) },
	})
	b.StopTimer()
	if err != nil {
		b.Fatal(err)
	}
	var checkers atomic.Int64
	program.ForEachCheckerParallel(func(int, *checker.Checker) { checkers.Add(1) })
	b.ReportMetric(float64(checkers.Load()), "checkers")
	b.ReportMetric(float64(c.Workers), "workers")
	b.ReportMetric(float64(runtime.GOMAXPROCS(0)), "gomaxprocs")
	b.ReportMetric(float64(len(files)), "files/op")
	b.ReportMetric(float64(ruleCount.Load()), "diagnostics/op")
	b.ReportMetric(float64(typeCount.Load()), "type-diagnostics/op")
}

type affinityCheckerProfile struct {
	AssignedFiles int                    `json:"assigned_files"`
	AssignedNodes int                    `json:"assigned_nodes"`
	ProgramNodes  int                    `json:"program_nodes"`
	Files         int                    `json:"files"`
	Calls         int                    `json:"calls"`
	RuleNS        int64                  `json:"rule_ns"`
	StolenFiles   int                    `json:"stolen_files"`
	FileProfiles  []*affinityFileProfile `json:"file_profiles"`
}

type affinityFileProfile struct {
	Path   string `json:"path"`
	RuleNS int64  `json:"rule_ns"`
}

func TestCheckerAffinityProbe(t *testing.T) {
	c := readAffinityCase(t)
	program, files := c.program(t)
	output := os.Getenv("TSGOLINT_AFFINITY_OUTPUT")
	if output == "" {
		t.Fatal("set TSGOLINT_AFFINITY_OUTPUT")
	}
	relative := func(path string) string {
		path, err := filepath.Rel(c.Root, path)
		if err != nil {
			t.Fatal(err)
		}
		return filepath.ToSlash(path)
	}
	write := func(value any) {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(output, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if c.Mode == "manifest" {
		paths := make([]string, len(files))
		for i, file := range files {
			paths[i] = relative(file.FileName())
		}
		write(paths)
		return
	}
	c.warm(program)
	rules := c.rules(t)
	var mu sync.Mutex
	var diagnostics []headlessDiagnostic
	profiles := make(map[*checker.Checker]*affinityCheckerProfile)
	indices := make(map[*checker.Checker]int)
	program.ForEachCheckerParallel(func(idx int, ch *checker.Checker) {
		mu.Lock()
		defer mu.Unlock()
		indices[ch] = idx
		profiles[ch] = &affinityCheckerProfile{}
	})
	if c.Mode == "profile" {
		for _, file := range program.GetSourceFiles() {
			ch, release := program.GetTypeCheckerForFile(t.Context(), file)
			profiles[ch].ProgramNodes += file.NodeCount
			release()
		}
		owners := make(map[*ast.SourceFile]*checker.Checker, len(files))
		for _, file := range files {
			ch, release := program.GetTypeCheckerForFile(t.Context(), file)
			owners[file] = ch
			profiles[ch].AssignedFiles++
			profiles[ch].AssignedNodes += file.NodeCount
			release()
		}
		instrumented := make([]linter.ConfiguredRule, len(rules))
		for i, r := range rules {
			instrumented[i] = linter.ConfiguredRule{Name: r.Name, Run: func(ctx rule.RuleContext) rule.RuleListeners {
				profile := profiles[ctx.TypeChecker]
				// Each checker has a single consumer in both implementations.
				if i == 0 {
					profile.Files++
					if ctx.TypeChecker != owners[ctx.SourceFile] {
						profile.StolenFiles++
					}
					profile.FileProfiles = append(profile.FileProfiles, &affinityFileProfile{Path: relative(ctx.SourceFile.FileName())})
				}
				fileProfile := profile.FileProfiles[len(profile.FileProfiles)-1]
				start := time.Now()
				listeners := r.Run(ctx)
				elapsed := time.Since(start).Nanoseconds()
				profile.RuleNS += elapsed
				fileProfile.RuleNS += elapsed
				profile.Calls++
				for kind, listener := range listeners {
					listeners[kind] = func(node *ast.Node) {
						start := time.Now()
						listener(node)
						elapsed := time.Since(start).Nanoseconds()
						profile.RuleNS += elapsed
						fileProfile.RuleNS += elapsed
						profile.Calls++
					}
				}
				return listeners
			}}
		}
		rules = instrumented
	}
	appendDiagnostic := func(d headlessDiagnostic) {
		mu.Lock()
		defer mu.Unlock()
		diagnostics = append(diagnostics, d)
	}
	err := linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
		Program: program, Files: files, Workers: c.Workers,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule { return rules },
		TypeErrors:      linter.TypeErrors{ReportSemantic: c.Semantic && !c.Warm},
		Fixes:           linter.Fixes{Fix: c.Mode != "profile", FixSuggestions: c.Mode != "profile"},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if c.Mode == "profile" {
				return
			}
			path := relative(d.SourceFile.FileName())
			hd := headlessDiagnostic{Kind: headlessDiagnosticKindRule, FilePath: &path, Rule: &d.RuleName,
				Range: headlessRangeFromRange(d.Range), Message: headlessRuleMessageFromRuleMessage(d.Message),
				Fixes: headlessFixesFromRuleFixes(d.Fixes())}
			for _, suggestion := range d.GetSuggestions() {
				hd.Suggestions = append(hd.Suggestions, headlessSuggestion{
					Message: headlessRuleMessageFromRuleMessage(suggestion.Message), Fixes: headlessFixesFromRuleFixes(suggestion.Fixes()),
				})
			}
			for _, label := range d.LabeledRanges {
				hd.LabeledRanges = append(hd.LabeledRanges, headlessLabeledRange{Label: label.Label, Range: *headlessRangeFromRange(label.Range)})
			}
			appendDiagnostic(hd)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) {
			if c.Mode == "profile" {
				return
			}
			var path *string
			if d.FilePath != nil {
				p := relative(*d.FilePath)
				path = &p
			}
			appendDiagnostic(headlessDiagnostic{Kind: headlessDiagnosticKindTsconfig, FilePath: path,
				Range: headlessRangeFromRange(d.Range), Message: headlessRuleMessage{Id: d.Id, Description: d.Description, Help: d.Help}})
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode == "profile" {
		ordered := make([]*affinityCheckerProfile, len(profiles))
		for ch, profile := range profiles {
			ordered[indices[ch]] = profile
		}
		write(ordered)
	} else {
		// The runner canonicalizes complete records; emission order is concurrent.
		write(diagnostics)
	}
}
