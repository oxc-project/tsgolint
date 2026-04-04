package await_thenable

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "await",
		Description: "Unexpected `await` of a non-Promise (non-\"Thenable\") value.",
		Help:        "Remove `await` if the value is synchronous, or change the expression to return a Promise or Thenable before awaiting it.",
	}
}

func buildRemoveAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAwait",
		Description: "Remove unnecessary `await`.",
	}
}

func buildForAwaitOfNonAsyncIterableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "forAwaitOfNonAsyncIterable",
		Description: "Unexpected `for await...of` of a value that is not async iterable.",
		Help:        "Use `for...of` for synchronous iterables, or change the iterable to implement `Symbol.asyncIterator`.",
	}
}

func buildConvertToOrdinaryForMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "convertToOrdinaryFor",
		Description: "Convert to an ordinary `for...of` loop.",
	}
}

func buildAwaitUsingOfNonAsyncDisposableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "awaitUsingOfNonAsyncDisposable",
		Description: "Unexpected `await using` of a value that is not async disposable.",
		Help:        "Use plain `using` for synchronous disposables, or change the value to implement `Symbol.asyncDispose`.",
	}
}

func buildInvalidPromiseAggregatorInputMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidPromiseAggregatorInput",
		Description: "Unexpected iterable of non-Promise (non-\"Thenable\") values passed to promise aggregator.",
		Help:        "Pass an iterable of Promise-like values, or wrap each synchronous value in `Promise.resolve(...)` before calling the aggregator.",
	}
}

func buildAwaitedExpressionLabel(label string, nodeRange core.TextRange) rule.RuleLabeledRange {
	return rule.RuleLabeledRange{
		Label: label,
		Range: nodeRange,
	}
}

func buildDiagnosticLabelText(messageID string) string {
	switch messageID {
	case "forAwaitOfNonAsyncIterable":
		return "This value is not async iterable"
	case "awaitUsingOfNonAsyncDisposable":
		return "This value is not async disposable"
	default:
		return "This expression is not Promise-like"
	}
}

func buildAwaitThenableDiagnostic(diagnosticRange core.TextRange, message rule.RuleMessage, labelRange core.TextRange, extraLabels ...rule.RuleLabeledRange) rule.RuleDiagnostic {
	labels := []rule.RuleLabeledRange{
		buildAwaitedExpressionLabel(buildDiagnosticLabelText(message.Id), labelRange),
	}
	labels = append(labels, extraLabels...)
	return rule.RuleDiagnostic{
		Range:         diagnosticRange,
		Message:       message,
		LabeledRanges: labels,
	}
}

var promiseAggregatorMethods = []string{
	"all",
	"allSettled",
	"any",
	"race",
}

var AwaitThenableRule = rule.Rule{
	Name: "await-thenable",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				awaitArgument := node.AsAwaitExpression().Expression
				awaitArgumentType := ctx.TypeChecker.GetTypeAtLocation(awaitArgument)
				certainty := utils.NeedsToBeAwaited(ctx.TypeChecker, awaitArgument, awaitArgumentType)

				if certainty == utils.TypeAwaitableNever {
					awaitTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
					ctx.ReportDiagnosticWithSuggestions(buildAwaitThenableDiagnostic(
						awaitTokenRange,
						buildAwaitMessage(),
						utils.TrimNodeTextRange(ctx.SourceFile, awaitArgument),
					), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message: buildRemoveAwaitMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(awaitTokenRange),
							},
						}}
					})
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if !isPromiseAggregatorMethod(ctx, expr) {
					return
				}

				if len(expr.Arguments.Nodes) == 0 {
					return
				}

				argument := expr.Arguments.Nodes[0]
				if argument == nil {
					return
				}

				if ast.IsArrayLiteralExpression(argument) {
					for _, element := range argument.AsArrayLiteralExpression().Elements.Nodes {
						if element == nil || ast.IsOmittedExpression(element) {
							continue
						}

						t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, element)
						if isAlwaysNonAwaitableType(ctx.TypeChecker, element, t) {
							argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, argument)
							elementRange := utils.TrimNodeTextRange(ctx.SourceFile, element)
							ctx.ReportDiagnostic(buildAwaitThenableDiagnostic(
								elementRange,
								buildInvalidPromiseAggregatorInputMessage(),
								elementRange,
								rule.RuleLabeledRange{
									Label: "Promise aggregator input",
									Range: argumentRange,
								},
							))
						}
					}

					return
				}

				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
				if isInvalidPromiseAggregatorInput(ctx.TypeChecker, argument, t) {
					argRange := utils.TrimNodeTextRange(ctx.SourceFile, argument)
					ctx.ReportDiagnostic(buildAwaitThenableDiagnostic(
						argRange,
						buildInvalidPromiseAggregatorInputMessage(),
						argRange,
					))
				}
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				stmt := node.AsForInOrOfStatement()
				if stmt.AwaitModifier == nil {
					return
				}

				exprType := ctx.TypeChecker.GetTypeAtLocation(stmt.Expression)
				if utils.IsTypeAnyType(exprType) {
					return
				}

				for _, typePart := range utils.UnionTypeParts(exprType) {
					if utils.GetWellKnownSymbolPropertyOfType(typePart, "asyncIterator", ctx.TypeChecker) != nil {
						return
					}
				}

				ctx.ReportDiagnosticWithSuggestions(
					buildAwaitThenableDiagnostic(
						utils.GetForStatementHeadLoc(ctx.SourceFile, node),
						buildForAwaitOfNonAsyncIterableMessage(),
						utils.TrimNodeTextRange(ctx.SourceFile, stmt.Expression),
					),
					func() []rule.RuleSuggestion {
						// Note that this suggestion causes broken code for sync iterables
						// of promises, since the loop variable is not awaited.
						return []rule.RuleSuggestion{{
							Message: buildConvertToOrdinaryForMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemove(ctx.SourceFile, stmt.AwaitModifier),
							},
						}}
					},
				)
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !ast.IsVarAwaitUsing(node) {
					return
				}

				declaration := node.AsVariableDeclarationList()
			DeclaratorLoop:
				for _, declarator := range declaration.Declarations.Nodes {
					init := declarator.Initializer()
					if init == nil {
						continue
					}
					initType := ctx.TypeChecker.GetTypeAtLocation(init)
					if utils.IsTypeAnyType(initType) {
						continue
					}

					for _, typePart := range utils.UnionTypeParts(initType) {
						if utils.GetWellKnownSymbolPropertyOfType(typePart, "asyncDispose", ctx.TypeChecker) != nil {
							continue DeclaratorLoop
						}
					}

					var suggestions []rule.RuleSuggestion
					// let the user figure out what to do if there's
					// await using a = b, c = d, e = f;
					// it's rare and not worth the complexity to handle.
					if len(declaration.Declarations.Nodes) == 1 {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message: buildRemoveAwaitMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())),
							},
						})
					}

					initRange := utils.TrimNodeTextRange(ctx.SourceFile, init)
					ctx.ReportDiagnosticWithSuggestions(buildAwaitThenableDiagnostic(
						initRange,
						buildAwaitUsingOfNonAsyncDisposableMessage(),
						initRange,
					), func() []rule.RuleSuggestion { return suggestions })
				}
			},
		}
	},
}

func isPromiseAggregatorMethod(
	ctx rule.RuleContext,
	callExpression *ast.CallExpression,
) bool {
	callee := ast.SkipParentheses(callExpression.Expression)
	if !ast.IsAccessExpression(callee) {
		return false
	}

	methodName, ok := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
	if !ok || !slices.Contains(promiseAggregatorMethods, methodName) {
		return false
	}

	return utils.IsPromiseConstructorLike(
		ctx.Program,
		ctx.TypeChecker,
		utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression()),
	)
}

func isInvalidPromiseAggregatorInput(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	if !isIterable(typeChecker, t) {
		return false
	}

	for _, part := range utils.UnionTypeParts(t) {
		valueTypes := getValueTypesOfArrayLike(typeChecker, part)
		for _, valueType := range valueTypes {
			if containsNonAwaitableType(typeChecker, node, valueType) {
				return true
			}
		}
	}

	return false
}

func getValueTypesOfArrayLike(
	typeChecker *checker.Checker,
	t *checker.Type,
) []*checker.Type {
	if checker.IsTupleType(t) {
		return checker.Checker_getTypeArguments(typeChecker, t)
	}

	if numberIndexType := utils.GetNumberIndexType(typeChecker, t); numberIndexType != nil {
		return []*checker.Type{numberIndexType}
	}

	typeArguments := checker.Checker_getTypeArguments(typeChecker, t)
	if len(typeArguments) != 0 {
		return typeArguments[:1]
	}

	return nil
}

func isAlwaysNonAwaitableType(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	return utils.Every(utils.UnionTypeParts(t), func(typePart *checker.Type) bool {
		return utils.NeedsToBeAwaited(typeChecker, node, typePart) == utils.TypeAwaitableNever
	})
}

func containsNonAwaitableType(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	return utils.Some(utils.UnionTypeParts(t), func(typePart *checker.Type) bool {
		return utils.NeedsToBeAwaited(typeChecker, node, typePart) == utils.TypeAwaitableNever
	})
}

func isIterable(
	typeChecker *checker.Checker,
	t *checker.Type,
) bool {
	return utils.Every(utils.UnionTypeParts(t), func(typePart *checker.Type) bool {
		return utils.GetWellKnownSymbolPropertyOfType(typePart, "iterator", typeChecker) != nil
	})
}
