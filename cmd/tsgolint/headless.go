package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"slices"
	"sync"

	"github.com/go-json-experiment/json"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type headlessRange struct {
	Pos int `json:"pos"`
	End int `json:"end"`
}

func headlessRangeFromRange(r core.TextRange) *headlessRange {
	if !r.IsValid() {
		return nil
	}
	return &headlessRange{
		Pos: r.Pos(),
		End: r.End(),
	}
}

type headlessRuleMessage struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Help        string `json:"help,omitempty"`
}

func headlessRuleMessageFromRuleMessage(msg rule.RuleMessage) headlessRuleMessage {
	return headlessRuleMessage{
		Id:          msg.Id,
		Description: msg.Description,
		Help:        msg.Help,
	}
}

type headlessFix struct {
	Text  string        `json:"text"`
	Range headlessRange `json:"range"`
}
type headlessSuggestion struct {
	Message headlessRuleMessage `json:"message"`
	Fixes   []headlessFix       `json:"fixes"`
}

// Diagnostic kind discriminator
type headlessDiagnosticKind uint8

const (
	headlessDiagnosticKindRule headlessDiagnosticKind = iota
	headlessDiagnosticKindTsconfig
)

type headlessDiagnostic struct {
	Kind     headlessDiagnosticKind `json:"kind"`
	Range    *headlessRange         `json:"range,omitempty"`
	Message  headlessRuleMessage    `json:"message"`
	FilePath *string                `json:"file_path"`

	// Only for kind="rule"
	Rule        *string              `json:"rule,omitempty"`
	Fixes       []headlessFix        `json:"fixes,omitempty"`
	Suggestions []headlessSuggestion `json:"suggestions,omitempty"`
}

type headlessMessageType uint8

const (
	headlessMessageTypeError headlessMessageType = iota
	headlessMessageTypeDiagnostic
)

type headlessMessagePayloadError struct {
	Error string `json:"error"`
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

func writeMessage(w io.Writer, messageType headlessMessageType, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var header [5]byte
	binary.LittleEndian.PutUint32(header[:], uint32(len(payloadBytes)))
	header[4] = byte(messageType)
	w.Write(header[:])
	w.Write(payloadBytes)
	return nil
}

func writeErrorMessage(text string) error {
	return writeMessage(os.Stdout, headlessMessageTypeError, headlessMessagePayloadError{
		Error: text,
	})
}

func runHeadless(args []string) int {
	logLevel := utils.GetLogLevel()

	var (
		traceOut       string
		cpuprofOut     string
		heapOut        string
		allocsOut      string
		fix            bool
		fixSuggestions bool
	)
	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.StringVar(&heapOut, "heap", "", "file to put heap profiling to")
	flag.StringVar(&allocsOut, "allocs", "", "file to put allocs profiling to")
	flag.BoolVar(&fix, "fix", false, "generate fixes for code problems")
	flag.BoolVar(&fixSuggestions, "fix-suggestions", false, "generate suggestions for code problems")
	flag.CommandLine.Parse(args)

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	if logLevel == utils.LogLevelDebug {
		log.Printf("Starting tsgolint")
	}

	if done, err := recordTrace(traceOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}
	if done, err := recordCpuprof(cpuprofOut); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	} else {
		defer done()
	}

	cwd, err := os.Getwd()
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error getting current directory: %v", err))
		return 1
	}

	configRaw, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error reading from stdin: %v", err))
		return 1
	}

	payload, err := deserializePayload(configRaw)

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

	fileConfigs := make(map[string][]headlessRule, totalFileCount)
	for _, config := range payload.Configs {
		for _, filePath := range config.FilePaths {
			fileConfigs[filePath] = config.Rules
		}
	}

	normalizedFiles := make([]string, 0, totalFileCount)
	for _, config := range payload.Configs {
		for _, filePath := range config.FilePaths {
			normalized := tspath.NormalizeSlashes(filePath)
			normalizedFiles = append(normalizedFiles, normalized)
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
			var hd headlessDiagnostic

			if d.ruleDiagnostic != nil {
				// Rule diagnostic
				rd := d.ruleDiagnostic
				filePath := rd.SourceFile.FileName()
				hd = headlessDiagnostic{
					Kind:        headlessDiagnosticKindRule,
					Range:       headlessRangeFromRange(rd.Range),
					Rule:        &rd.RuleName,
					Message:     headlessRuleMessageFromRuleMessage(rd.Message),
					Fixes:       nil,
					Suggestions: nil,
					FilePath:    &filePath,
				}

				if fix {
					hd.Fixes = make([]headlessFix, len(rd.Fixes()))
					for i, fix := range rd.Fixes() {
						hd.Fixes[i] = headlessFix{
							Text:  fix.Text,
							Range: *headlessRangeFromRange(fix.Range),
						}
					}
				}
				if fixSuggestions {
					hd.Suggestions = make([]headlessSuggestion, len(rd.GetSuggestions()))
					for i, suggestion := range rd.GetSuggestions() {
						hd.Suggestions[i] = headlessSuggestion{
							Message: headlessRuleMessageFromRuleMessage(rd.Message),
							Fixes:   make([]headlessFix, len(suggestion.Fixes())),
						}
						for j, fix := range suggestion.Fixes() {
							hd.Suggestions[i].Fixes[j] = headlessFix{
								Text:  fix.Text,
								Range: *headlessRangeFromRange(fix.Range),
							}
						}
					}
				}
			} else if d.internalDiagnostic != nil {
				// Internal diagnostic (tsconfig, type error, etc.)
				internalDiagnostic := d.internalDiagnostic

				hd = headlessDiagnostic{
					Kind:  headlessDiagnosticKindTsconfig,
					Range: headlessRangeFromRange(internalDiagnostic.Range),
					Rule:  nil, // Internal diagnostics don't have a rule
					Message: headlessRuleMessage{
						Id:          internalDiagnostic.Id,
						Description: internalDiagnostic.Description,
						Help:        internalDiagnostic.Help,
					},
					Fixes:       nil,
					Suggestions: nil,
					FilePath:    internalDiagnostic.FilePath,
				}
			}

			writeMessage(w, headlessMessageTypeDiagnostic, hd)
			if w.Available() < 4096 {
				w.Flush()
			}
		}
	})

	if logLevel == utils.LogLevelDebug {
		log.Printf("Running Linter")
	}

	err = linter.RunLinter(
		logLevel,
		cwd,
		workload,
		runtime.GOMAXPROCS(0),
		fs,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
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
		func(d rule.RuleDiagnostic) {
			diagnosticsChan <- ruleToAny(d)
		},
		func(d diagnostic.Internal) {
			diagnosticsChan <- internalToAny(d)
		},
		linter.Fixes{
			Fix:            fix,
			FixSuggestions: fixSuggestions,
		},
		linter.TypeErrors{
			ReportSyntactic: payload.ReportSyntactic,
			ReportSemantic:  payload.ReportSemantic,
		},
	)

	close(diagnosticsChan)
	if err != nil {
		log.Printf("ERROR: Linter failed: %v", err)
		writeErrorMessage(fmt.Sprintf("error running linter: %v", err))
		return 1
	}

	wg.Wait()

	if logLevel == utils.LogLevelDebug {
		log.Printf("Linting Complete")
	}

	writeMemProfiles(heapOut, allocsOut)

	return 0
}
