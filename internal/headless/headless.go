package headless

import (
	"bufio"
	"fmt"
	"io"
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

type Range struct { Pos int `json:"pos"`; End int `json:"end"` }
func fromRange(r core.TextRange) Range { return Range{Pos: r.Pos(), End: r.End()} }

type RuleMessage struct { Id string `json:"id"`; Description string `json:"description"`; Help string `json:"help,omitempty"` }
func fromRuleMessage(m rule.RuleMessage) RuleMessage { return RuleMessage{Id: m.Id, Description: m.Description, Help: m.Help} }

type Fix struct { Text string `json:"text"`; Range Range `json:"range"` }

type Suggestion struct { Message RuleMessage `json:"message"`; Fixes []Fix `json:"fixes"` }

type Diagnostic struct {
	Range Range `json:"range"`
	Rule string `json:"rule"`
	Message RuleMessage `json:"message"`
	Fixes []Fix `json:"fixes"`
	Suggestions []Suggestion `json:"suggestions"`
	FilePath string `json:"file_path"`
}

type messageType uint8
const (
	messageTypeError messageType = iota
	messageTypeDiagnostic
)

type errorPayload struct { Error string `json:"error"` }

func writeMessage(w io.Writer, mt messageType, payload any) error {
	b, err := json.Marshal(payload); if err != nil { return err }
	var header [5]byte
	// little endian length then type byte
	header[0] = byte(len(b))
	header[1] = byte(len(b) >> 8)
	header[2] = byte(len(b) >> 16)
	header[3] = byte(len(b) >> 24)
	header[4] = byte(mt)
	w.Write(header[:]); w.Write(b)
	return nil
}

// Run executes the headless lint using provided stdin payload bytes and rules map.
// rulesByName maps rule names to implementations (usually provided by the CLI layer).
func Run(stdin []byte, rulesByName map[string]rule.Rule, cwd string, logLevel utils.LogLevel, out io.Writer) int {
	payload, err := deserializePayload(stdin)
	if err != nil {
		writeMessage(out, messageTypeError, errorPayload{Error: fmt.Sprintf("error parsing config: %v", err)})
		return 1
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	workload := linter.Workload{Programs: make(map[string][]string), UnmatchedFiles: []string{}}

	fileConfigs := make(map[string][]headlessRule)
	// Count & assign files to programs
	total := 0
	for _, c := range payload.Configs { total += len(c.FilePaths) }
	resolver := utils.NewTsConfigResolver(fs, cwd)
	idx := 0
	for _, c := range payload.Configs {
		for _, fp := range c.FilePaths {
			if logLevel == utils.LogLevelDebug { fmt.Printf("[%d/%d] Processing file: %s\n", idx+1, total, fp) }
			nfp := tspath.NormalizeSlashes(fp)
			tsconfig, found := resolver.FindTsconfigForFile(nfp, false)
			if !found { workload.UnmatchedFiles = append(workload.UnmatchedFiles, nfp) } else { workload.Programs[tsconfig] = append(workload.Programs[tsconfig], nfp) }
			fileConfigs[nfp] = c.Rules
			idx++
		}
	}
	for _, files := range workload.Programs {
		slices.SortFunc(files, func(a,b string) int { return len(b)-len(a) })
	}

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		defer wg.Done()
		w := bufio.NewWriterSize(out, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			diag := Diagnostic{
				Range: fromRange(d.Range), Rule: d.RuleName, Message: fromRuleMessage(d.Message),
				Fixes: make([]Fix, len(d.Fixes())), Suggestions: make([]Suggestion, len(d.GetSuggestions())), FilePath: d.SourceFile.FileName(),
			}
			for i, f := range d.Fixes() { diag.Fixes[i] = Fix{Text: f.Text, Range: fromRange(f.Range)} }
			for i, s := range d.GetSuggestions() {
				diag.Suggestions[i] = Suggestion{Message: fromRuleMessage(d.Message), Fixes: make([]Fix, len(s.Fixes()))}
				for j, f := range s.Fixes() { diag.Suggestions[i].Fixes[j] = Fix{Text: f.Text, Range: fromRange(f.Range)} }
			}
			writeMessage(w, messageTypeDiagnostic, diag)
			if w.Available() < 4096 { w.Flush() }
		}
	}()

	err = linter.RunLinter(
		logLevel,
		cwd,
		workload,
		runtime.GOMAXPROCS(0),
		func(sf *ast.SourceFile) []linter.ConfiguredRule {
			cfg := fileConfigs[sf.FileName()]
			res := make([]linter.ConfiguredRule, len(cfg))
			for i, hr := range cfg {
				ruleImpl, ok := rulesByName[hr.Name]
				if !ok { panic(fmt.Sprintf("unknown rule: %s", hr.Name)) }
				res[i] = linter.ConfiguredRule{Name: ruleImpl.Name, Run: func(ctx rule.RuleContext) rule.RuleListeners { return ruleImpl.Run(ctx, nil) }}
			}
			return res
		},
		func(d rule.RuleDiagnostic){ diagnosticsChan <- d },
	)
	close(diagnosticsChan)
	if err != nil {
		writeMessage(out, messageTypeError, errorPayload{Error: fmt.Sprintf("error running linter: %v", err)})
		wg.Wait()
		return 1
	}
	wg.Wait()
	return 0
}
