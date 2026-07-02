package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"slices"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type headlessOptions struct {
	traceOut       string
	cpuprofOut     string
	heapOut        string
	allocsOut      string
	fix            bool
	fixSuggestions bool
	debugTimings   bool
}

var suppressProgramDiagnostics = sync.OnceValue(func() bool {
	return os.Getenv("OXLINT_TSGOLINT_DANGEROUSLY_SUPPRESS_PROGRAM_DIAGNOSTICS") == "true"
})

func parseHeadlessOptions(args []string) (*headlessOptions, error) {
	var opts headlessOptions
	var debug string

	flag.StringVar(&opts.traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&opts.cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.StringVar(&opts.heapOut, "heap", "", "file to put heap profiling to")
	flag.StringVar(&opts.allocsOut, "allocs", "", "file to put allocs profiling to")
	flag.BoolVar(&opts.fix, "fix", false, "generate fixes for code problems")
	flag.BoolVar(&opts.fixSuggestions, "fix-suggestions", false, "generate suggestions for code problems")
	flag.StringVar(&debug, "debug", "", "enable debug output options")

	if err := flag.CommandLine.Parse(args); err != nil {
		return nil, err
	}

	debugTimings, err := parseDebugTimings(debug)
	if err != nil {
		return nil, err
	}
	opts.debugTimings = debugTimings

	return &opts, nil
}

// Unified diagnostic type for channel
type anyDiagnostic struct {
	ruleDiagnostic     *rule.RuleDiagnostic
	internalDiagnostic *diagnostic.Internal
}

func ruleToAny(d rule.RuleDiagnostic) anyDiagnostic {
	return anyDiagnostic{ruleDiagnostic: &d}
}

func internalToAny(d diagnostic.Internal) anyDiagnostic {
	return anyDiagnostic{internalDiagnostic: &d}
}

func writeErrorMessage(text string) error {
	return writeMessage(os.Stdout, headlessMessageTypeError, headlessMessagePayloadError{
		Error: text,
	})
}

func runHeadless(args []string) int {
	logLevel := utils.GetLogLevel()
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	if logLevel == utils.LogLevelDebug {
		log.Printf("Starting tsgolint")
	}

	opts, err := parseHeadlessOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing options: %v", err)
		return 1
	}

	if cleanup, err := setupProfiling(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error starting profiling: %v", err)
		return 1
	} else {
		defer cleanup()
	}

	cwd, err := os.Getwd()
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error getting current directory: %v", err))
		return 1
	}

	jsonPayload, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error reading from stdin: %v", err))
		return 1
	}

	payload, err := deserializePayload(jsonPayload)
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error parsing config: %v", err))
		return 1
	}

	baseFS := osvfs.FS()
	if len(payload.SourceOverrides) > 0 {
		baseFS = newOverlayFS(baseFS, payload.SourceOverrides)
	}
	fs := bundled.WrapFS(cachedvfs.From(baseFS))

	workload := linter.Workload{
		Programs:       make(map[string][]string),
		UnmatchedFiles: []string{},
	}

	totalFileCount := 0
	for _, config := range payload.Configs {
		totalFileCount += len(config.FilePaths)
	}
	if logLevel == utils.LogLevelDebug {
		log.Printf("Starting to assign files to programs. Total files: %d", totalFileCount)
	}

	tsConfigResolver := utils.NewTsConfigResolver(fs, cwd)

	normalizedFiles := make([]string, 0, totalFileCount)
	fileConfigs := make(map[string][]headlessRule, totalFileCount)
	for _, config := range payload.Configs {
		for _, filePath := range config.FilePaths {
			normalized := tspath.NormalizeSlashes(filePath)
			normalizedFiles = append(normalizedFiles, normalized)

			fileConfigs[normalized] = config.Rules
		}
	}

	result := tsConfigResolver.FindTsConfigParallel(normalizedFiles)
	for file, tsconfig := range result {
		if tsconfig == "" {
			workload.UnmatchedFiles = append(workload.UnmatchedFiles, file)
		} else {
			workload.Programs[tsconfig] = append(workload.Programs[tsconfig], file)
		}
	}

	if logLevel == utils.LogLevelDebug {
		for file, tsconfig := range result {
			tsconfigStr := "<none>"
			if tsconfig != "" {
				tsconfigStr = tsconfig
			}
			log.Printf("Got tsconfig for file %s: %s", file, tsconfigStr)
		}

		log.Printf("Done assigning files to programs. Total programs: %d. Unmatched files: %d", len(workload.Programs), len(workload.UnmatchedFiles))
		for program, files := range workload.Programs {
			log.Printf("  Program %s: %d files", program, len(files))
		}
		for _, file := range workload.UnmatchedFiles {
			log.Printf("  Unmatched file: %s", file)
		}
	}

	allRulesByName := make(map[string]rule.Rule, len(allRules))
	for _, r := range allRules {
		allRulesByName[r.Name] = r
	}

	for _, files := range workload.Programs {
		slices.SortFunc(files, func(a, b string) int {
			return len(b) - len(a)
		})
	}

	if logLevel == utils.LogLevelDebug {
		log.Printf("Starting linter with %d workers", runtime.GOMAXPROCS(0))
		log.Printf("Workload distribution: %d programs", len(workload.Programs))
	}

	var wg sync.WaitGroup

	diagnosticsChan := make(chan anyDiagnostic, 4096)

	// Handle all diagnostics
	wg.Go(func() {
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			if d.ruleDiagnostic != nil {
				writeMessage(w, headlessMessageTypeDiagnostic, headlessDiagnosticFromRuleDiagnostic(
					d.ruleDiagnostic,
					opts.fix,
					opts.fixSuggestions,
				))
			} else if d.internalDiagnostic != nil {
				writeMessage(w, headlessMessageTypeDiagnostic, headlessDiagnosticFromInternalDiagnostic(d.internalDiagnostic))
			}

			if w.Available() < 4096 {
				w.Flush()
			}
		}
	})

	var timingStore *linter.RuleTimingStore
	if opts.debugTimings {
		timingStore = linter.NewRuleTimingStore()
	}

	if logLevel == utils.LogLevelDebug {
		log.Printf("Running Linter")
	}

	err = linter.RunLinter(linter.RunLinterOptions{
		LogLevel:         logLevel,
		CurrentDirectory: cwd,
		Workload:         workload,
		Workers:          runtime.GOMAXPROCS(0),
		FS:               fs,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			cfg := fileConfigs[sourceFile.FileName()]
			rules := make([]linter.ConfiguredRule, len(cfg))

			for i, headlessRule := range cfg {
				r, ok := allRulesByName[headlessRule.Name]
				if !ok {
					panic(fmt.Sprintf("unknown rule: %v", headlessRule.Name))
				}
				rules[i] = linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, headlessRule.Options)
					},
				}
			}

			return rules
		},
		OnRuleDiagnostic:     func(d rule.RuleDiagnostic) { diagnosticsChan <- ruleToAny(d) },
		OnInternalDiagnostic: func(d diagnostic.Internal) { diagnosticsChan <- internalToAny(d) },
		Fixes: linter.Fixes{
			Fix:            opts.fix,
			FixSuggestions: opts.fixSuggestions,
		},
		TypeErrors: linter.TypeErrors{
			ReportSyntactic: payload.ReportSyntactic,
			ReportSemantic:  payload.ReportSemantic,
		},
		SuppressProgramDiagnostics: suppressProgramDiagnostics(),
		TimingStore:                timingStore,
	})

	close(diagnosticsChan)
	if err != nil {
		log.Printf("ERROR: Linter failed: %v", err)
		writeErrorMessage(fmt.Sprintf("error running linter: %v", err))
		return 1
	}

	wg.Wait()

	if opts.debugTimings {
		if err := writeMessage(os.Stdout, headlessMessageTypeTiming, headlessTimingPayloadFromRecords(timingStore.Collect())); err != nil {
			log.Printf("ERROR: failed to write timing output: %v", err)
			return 1
		}
	}

	if logLevel == utils.LogLevelDebug {
		log.Printf("Linting Complete")
	}

	return 0
}
