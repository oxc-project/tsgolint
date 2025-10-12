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
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type headlessRange struct {
	Pos int `json:"pos"`
	End int `json:"end"`
}

func headlessRangeFromRange(r core.TextRange) headlessRange {
	return headlessRange{
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
type headlessDiagnostic struct {
	Range       headlessRange        `json:"range"`
	Rule        string               `json:"rule"`
	Message     headlessRuleMessage  `json:"message"`
	Fixes       []headlessFix        `json:"fixes"`
	Suggestions []headlessSuggestion `json:"suggestions"`
	FilePath    string               `json:"file_path"`
}

type headlessMessageType uint8

const (
	headlessMessageTypeError headlessMessageType = iota
	headlessMessageTypeDiagnostic
)

type headlessMessagePayloadError struct {
	Error string `json:"error"`
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
		traceOut   string
		cpuprofOut string
		heapOut    string
		allocsOut  string
	)
	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.StringVar(&heapOut, "heap", "", "file to put heap profiling to")
	flag.StringVar(&allocsOut, "allocs", "", "file to put allocs profiling to")
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

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

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

	idx := 0
	for _, config := range payload.Configs {
		for _, filePath := range config.FilePaths {
			if logLevel == utils.LogLevelDebug {
				log.Printf("[%d/%d] Processing file: %s", idx+1, totalFileCount, filePath)
			}

			normalizedFilePath := tspath.NormalizeSlashes(filePath)

			tsconfig, found := tsConfigResolver.FindTsconfigForFile(normalizedFilePath, false)
			if logLevel == utils.LogLevelDebug {
				tsconfigStr := "<none>"
				if found {
					tsconfigStr = tsconfig
				}
				log.Printf("Got tsconfig for file %s: %s", normalizedFilePath, tsconfigStr)
			}

			if !found {
				workload.UnmatchedFiles = append(workload.UnmatchedFiles, normalizedFilePath)
			} else {
				workload.Programs[tsconfig] = append(workload.Programs[tsconfig], normalizedFilePath)
			}
			fileConfigs[normalizedFilePath] = config.Rules
			idx++
		}
	}

	if logLevel == utils.LogLevelDebug {
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

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)

	wg.Go(func() {
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			hd := headlessDiagnostic{
				Range:       headlessRangeFromRange(d.Range),
				Rule:        d.RuleName,
				Message:     headlessRuleMessageFromRuleMessage(d.Message),
				Fixes:       make([]headlessFix, len(d.Fixes())),
				Suggestions: make([]headlessSuggestion, len(d.GetSuggestions())),
				FilePath:    d.SourceFile.FileName(),
			}
			for i, fix := range d.Fixes() {
				hd.Fixes[i] = headlessFix{
					Text:  fix.Text,
					Range: headlessRangeFromRange(fix.Range),
				}
			}
			for i, suggestion := range d.GetSuggestions() {
				hd.Suggestions[i] = headlessSuggestion{
					Message: headlessRuleMessageFromRuleMessage(d.Message),
					Fixes:   make([]headlessFix, len(suggestion.Fixes())),
				}
				for j, fix := range suggestion.Fixes() {
					hd.Suggestions[i].Fixes[j] = headlessFix{
						Text:  fix.Text,
						Range: headlessRangeFromRange(fix.Range),
					}
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
						return r.Run(ctx, nil)
					},
				}
			}

			return rules
		},
		func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
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
