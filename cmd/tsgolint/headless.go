package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

// findTsConfigForFile walks up the directory tree to find the appropriate tsconfig.json
func findTsConfigForFile(filePath, cwd string, fs bundled.FS) string {
	dir := filepath.Dir(filePath)
	for {
		tsconfigPath := filepath.Join(dir, "tsconfig.json")
		if fs.FileExists(tsconfigPath) {
			return tsconfigPath
		}

		parent := filepath.Dir(dir)
		if parent == dir || parent == "." || parent == "/" {
			// Reached root, return default tsconfig in cwd
			defaultTsconfig := filepath.Join(cwd, "tsconfig.json")
			if fs.FileExists(defaultTsconfig) {
				return defaultTsconfig
			}
			// Return a path for the default config even if it doesn't exist
			// (CreateProgram will handle this case)
			return defaultTsconfig
		}
		dir = parent
	}
}

func runHeadless(args []string) int {
	var (
		traceOut   string
		cpuprofOut string
	)
	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.CommandLine.Parse(args)

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
	cwd = tspath.NormalizePath(cwd)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

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

	// Group files by their tsconfig.json to minimize program creation
	tsconfigToFiles := make(map[string][]headlessConfigForFile)
	for _, fileConfig := range config.Files {
		absPath, err := filepath.Abs(fileConfig.FilePath)
		if err != nil {
			writeErrorMessage(fmt.Sprintf("error resolving absolute path for %v: %v", fileConfig.FilePath, err))
			return 1
		}

		tsconfigPath := findTsConfigForFile(absPath, cwd, fs)
		tsconfigToFiles[tsconfigPath] = append(tsconfigToFiles[tsconfigPath], fileConfig)
	}

	fileConfigs := make(map[*ast.SourceFile]headlessConfigForFile, len(config.Files))
	workload := linter.Workload{}

	// Create one program per tsconfig.json
	for tsconfigPath, filesForTsconfig := range tsconfigToFiles {
		currentDirectory := tspath.GetDirectoryPath(tsconfigPath)

		// Create overlay VFS with file contents
		overlayFiles := make(map[string]string)
		for _, fileConfig := range filesForTsconfig {
			source, err := os.ReadFile(fileConfig.FilePath)
			if err != nil {
				writeErrorMessage(fmt.Sprintf("error reading %v file: %v", fileConfig.FilePath, err))
				return 1
			}
			overlayFiles[fileConfig.FilePath] = string(source)
		}

		// If tsconfig doesn't exist, create a default one
		if !fs.FileExists(tsconfigPath) {
			overlayFiles[tsconfigPath] = "{}"
		}

		overlayFS := utils.NewOverlayVFS(fs, overlayFiles)
		host := utils.CreateCompilerHost(currentDirectory, overlayFS)

		// Create program using the same approach as main.go
		program, err := utils.CreateProgram(false, overlayFS, currentDirectory, tsconfigPath, host)
		if err != nil {
			writeErrorMessage(fmt.Sprintf("error creating TS program for %v: %v", tsconfigPath, err))
			return 1
		}

		// Map each requested file to its source file in the program
		for _, fileConfig := range filesForTsconfig {
			file := program.GetSourceFile(fileConfig.FilePath)
			if file == nil {
				// Try to find by normalized path
				normalizedPath := tspath.NormalizePath(fileConfig.FilePath)
				for _, sourceFile := range program.SourceFiles() {
					if tspath.NormalizePath(sourceFile.FileName()) == normalizedPath {
						file = sourceFile
						break
					}
				}
			}

			if file == nil {
				writeErrorMessage(fmt.Sprintf("file %v is not matched by tsconfig %v", fileConfig.FilePath, tsconfigPath))
				return 1
			}

			fileConfigs[file] = fileConfig
			workload[program] = append(workload[program], file)
		}
	}

	for _, files := range workload {
		slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
			return len(b.Text()) - len(a.Text())
		})
	}

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
		workload,
		runtime.GOMAXPROCS(0),
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
