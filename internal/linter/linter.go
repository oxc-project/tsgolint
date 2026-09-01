package linter

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

type ConfiguredRule struct {
	Name string
	Run  func(ctx rule.RuleContext) rule.RuleListeners
}

type Workload struct {
	Programs       map[string][]string
	UnmatchedFiles []string
}

type Fixes struct {
	Fix            bool
	FixSuggestions bool
}

type TypeErrors struct {
	ReportSyntactic bool
	ReportSemantic  bool
}

type checkerWorkload struct {
	checker *checker.Checker
	program *compiler.Program
	queue   chan *ast.SourceFile
}

type RunLinterOptions struct {
	LogLevel                   utils.LogLevel
	CurrentDirectory           string
	Workload                   Workload
	Workers                    int
	FS                         vfs.FS
	GetRulesForFile            func(sourceFile *ast.SourceFile) []ConfiguredRule
	OnRuleDiagnostic           func(diagnostic rule.RuleDiagnostic)
	OnInternalDiagnostic       func(d diagnostic.Internal)
	Fixes                      Fixes
	TypeErrors                 TypeErrors
	SuppressProgramDiagnostics bool
	TimingStore                *RuleTimingStore
}

// This is same as `RunLinterOptions` but for a single program.
type RunLinterOnProgramOptions struct {
	LogLevel             utils.LogLevel
	Program              *compiler.Program
	Files                []*ast.SourceFile
	Workers              int
	GetRulesForFile      func(sourceFile *ast.SourceFile) []ConfiguredRule
	OnDiagnostic         func(diagnostic rule.RuleDiagnostic)
	OnInternalDiagnostic func(d diagnostic.Internal)
	Fixes                Fixes
	TypeErrors           TypeErrors
	TimingStore          *RuleTimingStore
}

func RunLinter(options RunLinterOptions) error {
	logLevel := options.LogLevel
	currentDirectory := options.CurrentDirectory
	workload := options.Workload
	workers := options.Workers
	fs := options.FS
	getRulesForFile := options.GetRulesForFile
	onRuleDiagnostic := options.OnRuleDiagnostic
	onInternalDiagnostic := options.OnInternalDiagnostic
	fixState := options.Fixes
	typeErrors := options.TypeErrors
	suppressProgramDiagnostics := options.SuppressProgramDiagnostics
	timingStore := options.TimingStore

	idx := 0
	for configFileName, filePaths := range workload.Programs {
		if logLevel == utils.LogLevelDebug {
			log.Printf("[%d/%d] Running linter on program: %s", idx+1, len(workload.Programs), configFileName)
		}

		currentDirectory := tspath.GetDirectoryPath(configFileName)
		host := utils.NewCachedFSCompilerHost(currentDirectory, fs, bundled.LibPath(), nil, nil)

		program, diagnostics, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host, suppressProgramDiagnostics)

		if err != nil {
			return err
		}

		if program == nil {
			for _, d := range diagnostics {
				onInternalDiagnostic(d)
			}
			idx++
			continue
		}

		if logLevel == utils.LogLevelDebug {
			log.Printf("Program created with %d source files", len(program.GetSourceFiles()))
		}

		fileSet := make(map[string]struct{}, len(filePaths))
		for _, f := range filePaths {
			fileSet[f] = struct{}{}
		}

		sourceFiles := make([]*ast.SourceFile, 0, len(filePaths))
		for _, sf := range program.SourceFiles() {
			if _, ok := fileSet[sf.FileName()]; ok {
				sourceFiles = append(sourceFiles, sf)
				delete(fileSet, sf.FileName())
			}
		}

		if len(fileSet) > 0 {
			var unmatchedFiles []string
			for k := range fileSet {
				unmatchedFiles = append(unmatchedFiles, k)
			}
			unmatchedFilesString := strings.Join(unmatchedFiles, ", ")
			log.Println("Unmatched files found:", unmatchedFilesString)

			var programFiles []string
			for _, k := range program.SourceFiles() {
				programFiles = append(programFiles, k.FileName())
			}
			log.Printf("Program source files (%d): %s", len(programFiles), strings.Join(programFiles, ", "))

			panic(fmt.Sprintf("Expected file '%s' to be in program '%s'", unmatchedFilesString, configFileName))
		}

		err = RunLinterOnProgram(RunLinterOnProgramOptions{
			LogLevel:             logLevel,
			Program:              program,
			Files:                sourceFiles,
			Workers:              workers,
			GetRulesForFile:      getRulesForFile,
			OnDiagnostic:         onRuleDiagnostic,
			OnInternalDiagnostic: onInternalDiagnostic,
			Fixes:                fixState,
			TypeErrors:           typeErrors,
			TimingStore:          timingStore,
		})
		if err != nil {
			return err
		}

		idx++
	}

	{
		// A file that belongs to no tsconfig goes into an inferred project, which has no content
		// mappers, so its extension has to be one TypeScript itself can parse. Report the rest instead
		// of handing them to the parser, which has no script kind for them.
		inferredFiles := make([]string, 0, len(workload.UnmatchedFiles))
		for _, f := range workload.UnmatchedFiles {
			if core.GetScriptKindFromFileName(f) == core.ScriptKindUnknown {
				onInternalDiagnostic(unsupportedFileExtensionDiagnostic(f))
				continue
			}
			inferredFiles = append(inferredFiles, f)
		}

		host := utils.NewCachedFSCompilerHost(currentDirectory, fs, bundled.LibPath(), nil, nil)
		program, diagnostics, err := utils.CreateInferredProjectProgram(false, fs, currentDirectory, host, inferredFiles)

		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			for _, d := range diagnostics {
				onInternalDiagnostic(d)
			}
		}

		files := make([]*ast.SourceFile, 0, len(inferredFiles))
		for _, f := range inferredFiles {
			sf := program.GetSourceFile(f)
			if sf == nil {
				panic(fmt.Sprintf("Expected file '%s' to be in inferred program", f))
			}
			files = append(files, sf)
		}

		err = RunLinterOnProgram(RunLinterOnProgramOptions{
			LogLevel:             logLevel,
			Program:              program,
			Files:                files,
			Workers:              workers,
			GetRulesForFile:      getRulesForFile,
			OnDiagnostic:         onRuleDiagnostic,
			OnInternalDiagnostic: onInternalDiagnostic,
			Fixes:                fixState,
			TypeErrors:           typeErrors,
			TimingStore:          timingStore,
		})
		if err != nil {
			return err
		}
	}

	return nil

}

// unsupportedFileExtensionDiagnostic reports a file tsgolint was asked to lint but cannot parse. It is
// what a file with a languageOptions.parser override lands on when no content mapper claims its
// extension: either the tsconfig that includes it declares no mapper for it, or the mapper it declares
// failed to resolve, so the file was never registered and fell through to the inferred project.
func unsupportedFileExtensionDiagnostic(fileName string) diagnostic.Internal {
	extension := tspath.GetAnyExtensionFromPath(fileName, nil, false)
	if extension == "" {
		extension = "this file"
	}
	return diagnostic.Internal{
		Range:       core.NewTextRange(0, 0),
		Id:          "unsupported-file-extension",
		Description: "Unsupported file extension",
		Help: "tsgolint cannot type-check " + extension + " files on their own. Register a TypeScript " +
			"content mapper for the extension in the \"contentMappers\" of the tsconfig that includes this file.",
		FilePath: &fileName,
	}
}

// ruleContextBuilder is a per-worker struct that provides the RuleContext
// reporting methods. Instead of allocating 8 new closures per file, per rule, a
// single builder is created per worker goroutine and its mutable fields
// are updated before each rule invocation to match the current rule and file.
type ruleContextBuilder struct {
	file         *ast.SourceFile
	ruleName     string
	program      *compiler.Program
	checker      *checker.Checker
	fixState     Fixes
	onDiagnostic func(rule.RuleDiagnostic)
}

// Calls `onDiagnostic` with the given diagnostic's information, but sets the
// rule name and source file to match the file and rule currently being run.
func (b *ruleContextBuilder) emitDiagnostic(d rule.RuleDiagnostic) {
	d.RuleName = b.ruleName
	d.SourceFile = b.file
	if utils.IsContentMapped(b.file) && !mapContentMappedRuleDiagnostic(&d) {
		return
	}
	b.onDiagnostic(d)
}

// mapContentMappedRuleDiagnostic rewrites a diagnostic reported against a content mapper's virtual
// TypeScript so it points into the original file. It returns false when the diagnostic has no place in
// the original file and should be dropped.
//
// Fixes and suggestions survive only where every range they touch maps verbatim; see
// mapContentMappedFixes.
func mapContentMappedRuleDiagnostic(d *rule.RuleDiagnostic) bool {
	mapped, ok := utils.MapContentMappedLintRange(d.SourceFile, d.Range)
	if !ok {
		return false
	}
	d.Range = mapped
	if d.FixesPtr != nil {
		fixes, ok := mapContentMappedFixes(d.SourceFile, *d.FixesPtr)
		if !ok {
			fixes = nil
		}
		d.FixesPtr = &fixes
	}
	if d.Suggestions != nil {
		suggestions := make([]rule.RuleSuggestion, 0, len(*d.Suggestions))
		for _, suggestion := range *d.Suggestions {
			fixes, ok := mapContentMappedFixes(d.SourceFile, suggestion.FixesArr)
			if !ok {
				continue
			}
			suggestion.FixesArr = fixes
			suggestions = append(suggestions, suggestion)
		}
		d.Suggestions = &suggestions
	}
	if len(d.LabeledRanges) > 0 {
		labeled := make([]rule.RuleLabeledRange, 0, len(d.LabeledRanges))
		for _, labeledRange := range d.LabeledRanges {
			if mapped, ok := utils.MapContentMappedLintRange(d.SourceFile, labeledRange.Range); ok {
				labeledRange.Range = mapped
				labeled = append(labeled, labeledRange)
			}
		}
		d.LabeledRanges = labeled
	}
	return true
}

// mapContentMappedFixes rewrites an autofix's ranges into the original file. It returns false when any
// range is not an exact verbatim mapping, in which case the whole fix has to be dropped: applying part
// of a fix would be worse than applying none. In practice this keeps fixes for a mapped file's copied
// script content and drops them for anything the mapper synthesized.
func mapContentMappedFixes(file *ast.SourceFile, fixes []rule.RuleFix) ([]rule.RuleFix, bool) {
	mapped := make([]rule.RuleFix, len(fixes))
	for i, fix := range fixes {
		r, ok := utils.MapContentMappedEditRange(file, fix.Range)
		if !ok {
			return nil, false
		}
		fix.Range = r
		mapped[i] = fix
	}
	return mapped, true
}

func (b *ruleContextBuilder) reportDiagnosticWithFixes(d rule.RuleDiagnostic, fixesFn func() []rule.RuleFix) {
	var fixes []rule.RuleFix
	if b.fixState.Fix {
		fixes = fixesFn()
	}
	d.FixesPtr = &fixes
	b.emitDiagnostic(d)
}

func (b *ruleContextBuilder) reportDiagnosticWithSuggestions(d rule.RuleDiagnostic, suggestionsFn func() []rule.RuleSuggestion) {
	var suggestions []rule.RuleSuggestion
	if b.fixState.FixSuggestions {
		suggestions = suggestionsFn()
	}
	d.Suggestions = &suggestions
	b.emitDiagnostic(d)
}

func (b *ruleContextBuilder) reportRange(textRange core.TextRange, msg rule.RuleMessage) {
	b.emitDiagnostic(rule.RuleDiagnostic{
		Range:   textRange,
		Message: msg,
	})
}

func (b *ruleContextBuilder) reportRangeWithSuggestions(textRange core.TextRange, msg rule.RuleMessage, suggestionsFn func() []rule.RuleSuggestion) {
	var suggestions []rule.RuleSuggestion
	if b.fixState.FixSuggestions {
		suggestions = suggestionsFn()
	}
	b.emitDiagnostic(rule.RuleDiagnostic{
		Range:       textRange,
		Message:     msg,
		Suggestions: &suggestions,
	})
}

func (b *ruleContextBuilder) reportNode(node *ast.Node, msg rule.RuleMessage) {
	b.emitDiagnostic(rule.RuleDiagnostic{
		Range:   utils.TrimNodeTextRange(b.file, node),
		Message: msg,
	})
}

func (b *ruleContextBuilder) reportNodeWithFixes(node *ast.Node, msg rule.RuleMessage, fixesFn func() []rule.RuleFix) {
	var fixes []rule.RuleFix
	if b.fixState.Fix {
		fixes = fixesFn()
	}
	b.emitDiagnostic(rule.RuleDiagnostic{
		Range:    utils.TrimNodeTextRange(b.file, node),
		Message:  msg,
		FixesPtr: &fixes,
	})
}

func (b *ruleContextBuilder) reportNodeWithSuggestions(node *ast.Node, msg rule.RuleMessage, suggestionsFn func() []rule.RuleSuggestion) {
	var suggestions []rule.RuleSuggestion
	if b.fixState.FixSuggestions {
		suggestions = suggestionsFn()
	}
	b.emitDiagnostic(rule.RuleDiagnostic{
		Range:       utils.TrimNodeTextRange(b.file, node),
		Message:     msg,
		Suggestions: &suggestions,
	})
}

func newRuleContext(ctxBuilder *ruleContextBuilder) rule.RuleContext {
	return rule.RuleContext{
		ReportDiagnostic:                ctxBuilder.emitDiagnostic,
		ReportDiagnosticWithFixes:       ctxBuilder.reportDiagnosticWithFixes,
		ReportDiagnosticWithSuggestions: ctxBuilder.reportDiagnosticWithSuggestions,
		ReportRange:                     ctxBuilder.reportRange,
		ReportRangeWithSuggestions:      ctxBuilder.reportRangeWithSuggestions,
		ReportNode:                      ctxBuilder.reportNode,
		ReportNodeWithFixes:             ctxBuilder.reportNodeWithFixes,
		ReportNodeWithSuggestions:       ctxBuilder.reportNodeWithSuggestions,
	}
}

func reportTypeScriptDiagnostics(program *compiler.Program, files []*ast.SourceFile, typeErrors TypeErrors, onInternalDiagnostic func(d diagnostic.Internal)) {
	ctx := core.WithRequestID(context.Background(), "__single_run__")

	if typeErrors.ReportSyntactic {
		for _, file := range files {
			fileName := file.FileName()

			syntacticDiagnostics := program.GetSyntacticDiagnostics(ctx, file)
			for _, d := range syntacticDiagnostics {
				if d.File() != nil && d.File().FileName() == fileName {
					loc, ok := utils.MapContentMappedDiagnosticRange(file, d)
					if !ok {
						continue
					}
					onInternalDiagnostic(diagnostic.Internal{
						Range:       loc,
						Id:          "TS" + strconv.Itoa(int(d.Code())),
						Description: utils.GetDiagnosticMessage(d),
						FilePath:    &fileName,
					})
				}
			}
		}
	}

	if typeErrors.ReportSemantic {
		semanticDiagnosticsByFile := program.GetSemanticDiagnosticsWithoutNoEmitFiltering(ctx, files)

		programOption := program.Options()

		for _, file := range files {
			fileName := file.FileName()
			finalDiagnostics := compiler.FilterNoEmitSemanticDiagnostics(semanticDiagnosticsByFile[file], programOption)
			includeProcessorDiagnostics := program.GetIncludeProcessorDiagnostics(file)
			if len(finalDiagnostics) == 0 && len(includeProcessorDiagnostics) == 0 {
				continue
			}
			finalDiagnostics = append(append(make([]*ast.Diagnostic, 0, len(finalDiagnostics)+len(includeProcessorDiagnostics)), finalDiagnostics...), includeProcessorDiagnostics...)
			if len(finalDiagnostics) > 1 {
				finalDiagnostics = compiler.SortAndDeduplicateDiagnostics(finalDiagnostics)
			}

			for _, d := range finalDiagnostics {
				if d.File() != nil && d.File().FileName() == fileName {
					loc, ok := utils.MapContentMappedDiagnosticRange(file, d)
					if !ok {
						continue
					}
					onInternalDiagnostic(diagnostic.Internal{
						Range:       loc,
						Id:          "TS" + strconv.Itoa(int(d.Code())),
						Description: utils.GetDiagnosticMessage(d),
						FilePath:    &fileName,
					})
				}
			}
		}
	}
}

func makeSourceFileQueue(files []*ast.SourceFile) chan *ast.SourceFile {
	queue := make(chan *ast.SourceFile, len(files))
	for _, file := range files {
		queue <- file
	}
	close(queue)
	return queue
}

func makeCheckerWorkloadQueue(program *compiler.Program, files []*ast.SourceFile) chan checkerWorkload {
	queue := makeSourceFileQueue(files)
	flatQueue := []checkerWorkload{}
	var flatQueueMu sync.Mutex
	program.ForEachCheckerParallel(func(idx int, ch *checker.Checker) {
		flatQueueMu.Lock()
		flatQueue = append(flatQueue, checkerWorkload{ch, program, queue})
		flatQueueMu.Unlock()
	})

	workloadQueue := make(chan checkerWorkload, len(flatQueue))
	for _, w := range flatQueue {
		workloadQueue <- w
	}
	close(workloadQueue)
	return workloadQueue
}

func visitLintNodes(file *ast.SourceFile, runListeners func(kind ast.Kind, node *ast.Node)) {
	/* convert.ts -> allowPattern:
	catch name
	variabledeclaration name
	forinstatement initializer
	forofstatement initializer
	(propagation) allowPattern > arrayliteralexpression elements
	(propagation) allowPattern > objectliteralexpression properties
	(propagation) allowPattern > spreadassignment,spreadelement expression
	(propagation) allowPattern > propertyassignment value
	arraybindingpattern elements
	objectbindingpattern elements
	(init) binaryexpression(with '=' operator') left
	*/

	var childVisitor ast.Visitor
	var patternVisitor func(node *ast.Node)
	patternVisitor = func(node *ast.Node) {
		runListeners(node.Kind, node)
		kind := rule.ListenerOnAllowPattern(node.Kind)
		runListeners(kind, node)

		switch node.Kind {
		case ast.KindArrayLiteralExpression:
			for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
				patternVisitor(element)
			}
		case ast.KindObjectLiteralExpression:
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				patternVisitor(property)
			}
		case ast.KindSpreadElement, ast.KindSpreadAssignment:
			patternVisitor(node.Expression())
		case ast.KindPropertyAssignment:
			patternVisitor(node.Initializer())
		default:
			node.ForEachChild(childVisitor)
		}

		runListeners(rule.ListenerOnExit(kind), node)
		runListeners(rule.ListenerOnExit(node.Kind), node)
	}
	childVisitor = func(node *ast.Node) bool {
		runListeners(node.Kind, node)

		switch node.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			kind := rule.ListenerOnNotAllowPattern(node.Kind)
			runListeners(kind, node)
			node.ForEachChild(childVisitor)
			runListeners(rule.ListenerOnExit(kind), node)
		default:
			if ast.IsAssignmentExpression(node, true) {
				expr := node.AsBinaryExpression()
				patternVisitor(expr.Left)
				childVisitor(expr.OperatorToken)
				childVisitor(expr.Right)
			} else {
				node.ForEachChild(childVisitor)
			}
		}

		runListeners(rule.ListenerOnExit(node.Kind), node)

		return false
	}
	file.Node.ForEachChild(childVisitor)
}

func RunLinterOnProgram(options RunLinterOnProgramOptions) error {
	logLevel := options.LogLevel
	program := options.Program
	files := options.Files
	workers := options.Workers
	getRulesForFile := options.GetRulesForFile
	onDiagnostic := options.OnDiagnostic
	onInternalDiagnostic := options.OnInternalDiagnostic
	fixState := options.Fixes
	typeErrors := options.TypeErrors
	timingStore := options.TimingStore

	reportTypeScriptDiagnostics(program, files, typeErrors, onInternalDiagnostic)
	workloadQueue := makeCheckerWorkloadQueue(program, files)

	wg := core.NewWorkGroup(workers == 1)
	for range workers {
		wg.Queue(func() {
			ctxBuilder := &ruleContextBuilder{
				fixState:     fixState,
				onDiagnostic: onDiagnostic,
			}

			// These closures remain valid for the length of linting, as we mutate the fields
			// of `ctxBuilder`, but `ctxBuilder` itself will not change.
			ctx := newRuleContext(ctxBuilder)

			if timingStore == nil {
				// Listeners are tagged with the rule that is associated with, so that when a diagnostic
				// is emitted we know what rule it is coming from.
				type taggedListener struct {
					ruleName string
					fn       func(node *ast.Node)
				}
				registeredListeners := make(map[ast.Kind][]taggedListener, 20)

				for w := range workloadQueue {
					ctxBuilder.program = w.program
					ctxBuilder.checker = w.checker
					ctx.Program = w.program
					ctx.TypeChecker = w.checker

					for file := range w.queue {
						if logLevel == utils.LogLevelDebug {
							log.Print(file.FileName())
						}
						ctxBuilder.file = file
						ctx.SourceFile = file

						rules := getRulesForFile(file)
						for _, r := range rules {
							ctxBuilder.ruleName = r.Name
							for kind, listener := range r.Run(ctx) {
								listeners, ok := registeredListeners[kind]
								if !ok {
									listeners = make([]taggedListener, 0, len(rules))
								}
								registeredListeners[kind] = append(listeners, taggedListener{ruleName: r.Name, fn: listener})
							}
						}

						runListeners := func(kind ast.Kind, node *ast.Node) {
							if listeners, ok := registeredListeners[kind]; ok {
								for _, listener := range listeners {
									ctxBuilder.ruleName = listener.ruleName
									listener.fn(node)
								}
							}
						}

						visitLintNodes(file, runListeners)
						// Instead of clearing the map, we clear the slices in-place to avoid re-allocating memory for the listeners on each file.
						for k := range registeredListeners {
							registeredListeners[k] = registeredListeners[k][:0]
						}
					}
				}

				return
			}

			type timedTaggedListener struct {
				ruleName string
				ruleIdx  int
				fn       func(node *ast.Node)
			}
			registeredListeners := make(map[ast.Kind][]timedTaggedListener, 20)
			localTimings := make(map[string]RuleTimingStat, 64)

			recordTiming := func(stat *RuleTimingStat, duration time.Duration) {
				stat.Duration += duration
				stat.Calls++
			}

			for w := range workloadQueue {
				ctxBuilder.program = w.program
				ctxBuilder.checker = w.checker
				ctx.Program = w.program
				ctx.TypeChecker = w.checker

				for file := range w.queue {
					if logLevel == utils.LogLevelDebug {
						log.Print(file.FileName())
					}
					ctxBuilder.file = file
					ctx.SourceFile = file

					rules := getRulesForFile(file)
					timingStats := make([]RuleTimingStat, len(rules))
					for ruleIdx, r := range rules {
						ctxBuilder.ruleName = r.Name
						start := time.Now()
						listenersByKind := r.Run(ctx)
						recordTiming(&timingStats[ruleIdx], time.Since(start))
						for kind, listener := range listenersByKind {
							listeners, ok := registeredListeners[kind]
							if !ok {
								listeners = make([]timedTaggedListener, 0, len(rules))
							}
							registeredListeners[kind] = append(listeners, timedTaggedListener{ruleName: r.Name, ruleIdx: ruleIdx, fn: listener})
						}
					}

					runListeners := func(kind ast.Kind, node *ast.Node) {
						if listeners, ok := registeredListeners[kind]; ok {
							for _, listener := range listeners {
								ctxBuilder.ruleName = listener.ruleName
								start := time.Now()
								listener.fn(node)
								recordTiming(&timingStats[listener.ruleIdx], time.Since(start))
							}
						}
					}

					visitLintNodes(file, runListeners)
					for idx, stat := range timingStats {
						if stat.Calls == 0 {
							continue
						}
						merged := localTimings[rules[idx].Name]
						merged.add(stat)
						localTimings[rules[idx].Name] = merged
					}
					// Instead of clearing the map, we clear the slices in-place to avoid re-allocating memory for the listeners on each file.
					for k := range registeredListeners {
						registeredListeners[k] = registeredListeners[k][:0]
					}
				}
			}

			timingStore.merge(localTimings)
		})
	}
	wg.RunAndWait()

	return nil
}
