package linter

import (
	"context"

	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

type ConfiguredRule struct {
	Name string
	Run  func(ctx rule.RuleContext) rule.RuleListeners
}

type Workload = map[*compiler.Program][]*ast.SourceFile

func RunLinter(workload Workload, workers int, getRulesForFile func(sourceFile *ast.SourceFile) []ConfiguredRule, onDiagnostic func(diagnostic rule.RuleDiagnostic)) error {
	type checkerWorkload struct {
		checker *checker.Checker
		program *compiler.Program
		queue   chan *ast.SourceFile
	}
	flatQueue := []checkerWorkload{}
	for program, files := range workload {
		queue := make(chan *ast.SourceFile, len(files))
		for _, file := range files {
			queue <- file
		}
		close(queue)
		program.BindSourceFiles()
		checkers, done := program.GetTypeCheckers(core.WithRequestID(context.Background(), "__single_run__"))
		defer done()
		for _, ch := range checkers {
			flatQueue = append(flatQueue, checkerWorkload{ch, program, queue})
		}
	}
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
							ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
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
							ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixes ...rule.RuleFix) {
								onDiagnostic(rule.RuleDiagnostic{
									RuleName:   r.Name,
									Range:      utils.TrimNodeTextRange(file, node),
									Message:    msg,
									FixesPtr:   &fixes,
									SourceFile: file,
								})
							},

							ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
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
