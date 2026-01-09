package linter

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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

func RunLinter(
	logLevel utils.LogLevel,
	currentDirectory string,
	workload Workload,
	workers int,
	fs vfs.FS,
	getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule,
	onRuleDiagnostic func(diagnostic rule.RuleDiagnostic),
	onInternalDiagnostic func(d diagnostic.Internal),
	fixState Fixes,
	typeErrors TypeErrors,
) error {

	idx := 0
	for configFileName, filePaths := range workload.Programs {
		if logLevel == utils.LogLevelDebug {
			log.Printf("[%d/%d] Running linter on program: %s", idx+1, len(workload.Programs), configFileName)
		}

		currentDirectory := tspath.GetDirectoryPath(configFileName)
		host := utils.NewCachedFSCompilerHost(currentDirectory, fs, bundled.LibPath(), nil, nil)

		program, diagnostics, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host)

		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
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

		err = RunLinterOnProgram(logLevel, program, sourceFiles, workers, getRulesForFile, onRuleDiagnostic, onInternalDiagnostic, fixState, typeErrors)
		if err != nil {
			return err
		}

		idx++
	}

	{
		host := utils.NewCachedFSCompilerHost(currentDirectory, fs, bundled.LibPath(), nil, nil)
		program, diagnostics, err := utils.CreateInferredProjectProgram(false, fs, currentDirectory, host, workload.UnmatchedFiles)

		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			for _, d := range diagnostics {
				onInternalDiagnostic(d)
			}
		}

		files := make([]*ast.SourceFile, 0, len(workload.UnmatchedFiles))
		for _, f := range workload.UnmatchedFiles {
			sf := program.GetSourceFile(f)
			if sf == nil {
				panic(fmt.Sprintf("Expected file '%s' to be in inferred program", f))
			}
			files = append(files, sf)
		}

		err = RunLinterOnProgram(logLevel, program, files, workers, getRulesForFile, onRuleDiagnostic, onInternalDiagnostic, fixState, typeErrors)
		if err != nil {
			return err
		}
	}

	return nil

}

func RunLinterOnProgram(logLevel utils.LogLevel, program *compiler.Program, files []*ast.SourceFile, workers int, getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule, onDiagnostic func(diagnostic rule.RuleDiagnostic), onInternalDiagnostic func(d diagnostic.Internal), fixState Fixes, typeErrors TypeErrors) error {
	type checkerWorkload struct {
		checker *checker.Checker
		program *compiler.Program
		queue   chan *ast.SourceFile
	}
	flatQueue := []checkerWorkload{}
	queue := make(chan *ast.SourceFile, len(files))

	for _, f := range files {
		queue <- f
	}

	close(queue)
	program.BindSourceFiles()

	ctx := core.WithRequestID(context.Background(), "__single_run__")

	if typeErrors.ReportSyntactic || typeErrors.ReportSemantic {
		for _, file := range files {
			fileName := file.FileName()

			if typeErrors.ReportSyntactic {
				syntacticDiagnostics := program.GetSyntacticDiagnostics(ctx, file)
				for _, d := range syntacticDiagnostics {
					if d.File() != nil && d.File().FileName() == fileName {
						onInternalDiagnostic(diagnostic.Internal{
							Range:       d.Loc(),
							Id:          "TS" + strconv.Itoa(int(d.Code())),
							Description: utils.GetDiagnosticMessage(d),
							FilePath:    &fileName,
						})
					}
				}
			}

			if typeErrors.ReportSemantic {
				semanticDiagnostics := program.GetSemanticDiagnostics(ctx, file)
				for _, d := range semanticDiagnostics {
					if d.File() != nil && d.File().FileName() == fileName {
						onInternalDiagnostic(diagnostic.Internal{
							Range:       d.Loc(),
							Id:          "TS" + strconv.Itoa(int(d.Code())),
							Description: utils.GetDiagnosticMessage(d),
							FilePath:    &fileName,
						})
					}
				}
			}
		}
	}

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

	wg := core.NewWorkGroup(workers == 1)
	for range workers {
		wg.Queue(func() {
			registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)

			for w := range workloadQueue {
				for file := range w.queue {
					if logLevel == utils.LogLevelDebug {
						log.Print(file.FileName())
					}
					rules := getRulesForFile(file)
					for _, r := range rules {
						ctx := rule.RuleContext{
							SourceFile:  file,
							Program:     w.program,
							TypeChecker: w.checker,
							ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      textRange,
									Message:    msg,
									SourceFile: file,
								})
							},
							ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestionsFn func() []rule.RuleSuggestion) {
								var suggestions []rule.RuleSuggestion = nil
								if fixState.FixSuggestions {
									suggestions = suggestionsFn()
								}
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:    r.Name,
									Range:       textRange,
									Message:     msg,
									Suggestions: &suggestions,
									SourceFile:  file,
								})
							},
							ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      utils.TrimNodeTextRange(file, node),
									Message:    msg,
									SourceFile: file,
								})
							},
							ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixesFn func() []rule.RuleFix) {
								var fixes []rule.RuleFix = nil
								if fixState.Fix {
									fixes = fixesFn()
								}
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      utils.TrimNodeTextRange(file, node),
									Message:    msg,
									FixesPtr:   &fixes,
									SourceFile: file,
								})
							},

							ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestionsFn func() []rule.RuleSuggestion) {
								suggestions := suggestionsFn()
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:    r.Name,
									Range:       utils.TrimNodeTextRange(file, node),
									Message:     msg,
									Suggestions: &suggestions,
									SourceFile:  file,
								})
							},
						}

						for kind, listener := range r.Run(ctx) {
							listeners, ok := registeredListeners[kind]
							if !ok {
								listeners = make([](func(node *ast.Node)), 0, len(rules))
							}
							registeredListeners[kind] = append(listeners, listener)
						}
					}

					runListeners := func(kind ast.Kind, node *ast.Node) {
						if listeners, ok := registeredListeners[kind]; ok {
							for _, listener := range listeners {
								listener(node)
							}
						}
					}

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
					clear(registeredListeners)
				}
			}
		})
	}
	wg.RunAndWait()

	return nil
}
