package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

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

type headlessConfigForFile struct {
	FilePath string   `json:"file_path"`
	Rules    []string `json:"rules"`
}
type headlessConfig struct {
	Files []headlessConfigForFile `json:"files"`
}

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
}

func headlessRuleMessageFromRuleMessage(msg rule.RuleMessage) headlessRuleMessage {
	return headlessRuleMessage{
		Id:          msg.Id,
		Description: msg.Description,
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
	var (
		tsconfig string
		cwd      string
	)

	flag.StringVar(&tsconfig, "tsconfig", "", "which tsconfig to use")
	flag.StringVar(&cwd, "cwd", "", "current directory")
	flag.CommandLine.Parse(args)

	if tsconfig == "" {
		fmt.Fprint(os.Stderr, "--tsconfig is required\n")
		return 1
	}
	if cwd == "" {
		fmt.Fprint(os.Stderr, "--cwd is required\n")
		return 1
	}

	cwd = tspath.NormalizePath(cwd)
	tsconfig = tspath.NormalizePath(tsconfig)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	host := utils.CreateCompilerHost(cwd, fs)

	program, err := utils.CreateProgram(false, fs, cwd, tsconfig, host)
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error creating TS program: %v", err))
		return 1
	}

	configRaw, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeErrorMessage(fmt.Sprintf("error reading from stdin: %v", err))
		return 1
	}

	var config headlessConfig

	if err := json.Unmarshal(configRaw, &config); err != nil {
		writeErrorMessage(fmt.Sprintf("error parsing config: %v", err))
		return 1
	}

	fileConfigs := make(map[*ast.SourceFile]headlessConfigForFile, len(config.Files))
	files := make([]*ast.SourceFile, len(config.Files))
	for i, fileConfig := range config.Files {
		file := program.GetSourceFile(fileConfig.FilePath)
		if file == nil {
			writeErrorMessage(fmt.Sprintf("file %v is not matched by tsconfig", fileConfig.FilePath))
			return 1
		}
		files[i] = file
		fileConfigs[file] = fileConfig
	}

	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text()) - len(a.Text())
	})

	var wg sync.WaitGroup

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)

	wg.Add(1)
	go func() {
		defer wg.Done()
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
	}()

	err = linter.RunLinter(
		program,
		false,
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			cfg := fileConfigs[sourceFile]
			rules := make([]linter.ConfiguredRule, len(cfg.Rules))

			for i, ruleName := range cfg.Rules {
				r, ok := allRulesByName[ruleName]
				if !ok {
					panic(fmt.Sprintf("unknown rule: %v", ruleName))
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
		writeErrorMessage(fmt.Sprintf("error running linter: %v", err))
		return 1
	}

	wg.Wait()

	return 0
}
