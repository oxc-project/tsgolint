package no_floating_promises

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var messageBase = "Promises must be awaited."

var messageBaseHelp = "The promise must end with a call to .catch, or end with a call to .then with a rejection handler."

var messageBaseVoid = "Promises must be awaited."

var messageBaseVoidHelp = "The promise must end with a call to .catch, or end with a call to .then with a rejection handler, or be explicitly marked as ignored with the `void` operator."

var messageRejectionHandler = "A rejection handler that is not a function will be ignored."

func buildFloatingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floating",
		Description: messageBase,
		Help:        messageBaseHelp,
	}
}
func buildFloatingFixAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingFixAwait",
		Description: "Add await operator.",
	}
}
func buildFloatingFixVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingFixVoid",
		Description: "Add void operator to ignore.",
	}
}
func buildFloatingPromiseArrayMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingPromiseArray",
		Description: "An array of Promises may be unintentional.",
		Help:        "Consider handling the promises' fulfillment or rejection with Promise.all or similar.",
	}
}
func buildFloatingPromiseArrayVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingPromiseArrayVoid",
		Description: "An array of Promises may be unintentional.",
		Help: "Consider handling the promises' fulfillment or rejection with Promise.all or similar," +
			" or explicitly marking the expression as ignored with the `void` operator.",
	}
}
func buildFloatingUselessRejectionHandlerMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingUselessRejectionHandler",
		Description: messageBase,
		Help:        messageBaseHelp + " " + messageRejectionHandler,
	}
}
func buildFloatingUselessRejectionHandlerVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingUselessRejectionHandlerVoid",
		Description: messageBaseVoid,
		Help:        messageBaseVoidHelp + " " + messageRejectionHandler,
	}
}
func buildFloatingVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "floatingVoid",
		Description: messageBaseVoid,
		Help:        messageBaseVoidHelp,
	}
}

var NoFloatingPromisesRule = rule.Rule{
	Name: "no-floating-promises",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoFloatingPromisesOptions](options, "no-floating-promises")

		isHigherPrecedenceThanUnary := func(node *ast.Node) bool {
			operator := ast.KindUnknown
			if ast.IsBinaryExpression(node) {
				operator = node.AsBinaryExpression().OperatorToken.Kind
			}
			nodePrecedence := ast.GetOperatorPrecedence(node.Kind, operator, ast.OperatorPrecedenceFlagsNone)
			return nodePrecedence > ast.OperatorPrecedenceUnary
		}

		addAwait := func(
			expression *ast.Expression,
			node *ast.ExpressionStatement,
		) []rule.RuleFix {
			if ast.IsVoidExpression(expression) {
				voidTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expression.Pos())
				return []rule.RuleFix{rule.RuleFixReplaceRange(voidTokenRange, "await")}
			}
			if isHigherPrecedenceThanUnary(node.Expression) {
				return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, &node.Node, "await ")}
			}
			return []rule.RuleFix{
				rule.RuleFixInsertBefore(ctx.SourceFile, &node.Node, "await ("),
				rule.RuleFixInsertAfter(expression, ")"),
			}
		}
		hasMatchingSignature := func(
			t *checker.Type,
			matcher func(signature *checker.Signature) bool,
		) bool {
			for _, part := range utils.UnionTypeParts(t) {
				if utils.Some(utils.GetCallSignatures(ctx.TypeChecker, part), matcher) {
					return true
				}
			}

			return false
		}

		isFunctionParam := func(
			param *ast.Symbol,
			node *ast.Node,
		) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeOfSymbolAtLocation(param, node))

			for _, part := range utils.UnionTypeParts(t) {
				if len(utils.GetCallSignatures(ctx.TypeChecker, part)) != 0 {
					return true
				}
			}
			return false
		}
		isPromiseLike := func(node *ast.Node, t *checker.Type) bool {
			if t == nil {
				t = ctx.TypeChecker.GetTypeAtLocation(node)
			}

			// The highest priority is to allow anything allowlisted
			if utils.TypeMatchesSomeSpecifier(
				t,
				opts.AllowForKnownSafePromises,
				ctx.Program,
			) {
				return false
			}

			// Otherwise, we always consider the built-in Promise to be Promise-like...
			typeParts := utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, t))
			if utils.Some(typeParts, func(typePart *checker.Type) bool {
				return utils.IsPromiseLike(ctx.Program, ctx.TypeChecker, typePart)
			}) {
				return true
			}

			// ...and only check all Thenables if explicitly told to
			if !opts.CheckThenables {
				return false
			}

			// Modified from tsutils.isThenable() to only consider thenables which can be
			// rejected/caught via a second parameter. Original source (MIT licensed):
			//
			//   https://github.com/ajafff/tsutils/blob/49d0d31050b44b81e918eae4fbaf1dfe7b7286af/util/type.ts#L95-L125
			for _, typePart := range typeParts {
				then := checker.Checker_getPropertyOfType(ctx.TypeChecker, typePart, "then")
				if then == nil {
					continue
				}

				thenType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(then, node)
				if hasMatchingSignature(
					thenType,
					func(signature *checker.Signature) bool {
						params := checker.Signature_parameters(signature)
						return len(params) >= 2 && isFunctionParam(params[0], node) && isFunctionParam(params[1], node)
					}) {
					return true
				}
			}
			return false
		}
		isPromiseArray := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			for _, typePart := range utils.UnionTypeParts(t) {
				apparent := checker.Checker_getApparentType(ctx.TypeChecker, typePart)

				if checker.Checker_isArrayType(ctx.TypeChecker, apparent) {
					arrayType := checker.Checker_getTypeArguments(ctx.TypeChecker, apparent)[0]
					if isPromiseLike(node, arrayType) {
						return true
					}
				}

				if checker.IsTupleType(apparent) {
					for _, tupleElementType := range checker.Checker_getTypeArguments(ctx.TypeChecker, apparent) {
						if isPromiseLike(node, tupleElementType) {
							return true
						}
					}
				}
			}
			return false
		}

		isKnownSafePromiseReturn := func(node *ast.Node) bool {
			if !ast.IsCallExpression(node) {
				return false
			}

			callExpression := node.AsCallExpression()

			t := ctx.TypeChecker.GetTypeAtLocation(callExpression.Expression)

			if utils.ValueMatchesSomeSpecifier(
				callExpression.Expression,
				opts.AllowForKnownSafeCalls,
				ctx.Program,
				t,
			) {
				return true
			}

			return utils.TypeMatchesSomeSpecifier(
				t,
				opts.AllowForKnownSafeCalls,
				ctx.Program,
			)
		}

		isAsyncIife := func(node *ast.ExpressionStatement) bool {
			if !ast.IsCallExpression(node.Expression) {
				return false
			}

			callee := ast.SkipParentheses(node.Expression.AsCallExpression().Expression)

			return ast.IsArrowFunction(callee) || ast.IsFunctionExpression(callee)
		}

		isValidRejectionHandler := func(rejectionHandler *ast.Node) bool {
			return len(utils.GetCallSignatures(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(rejectionHandler))) > 0
		}

		var isUnhandledPromise func(
			node *ast.Node,
		) (
			bool, // isUnhandled
			bool, // nonFunctionHandler
			bool, // promiseArray
		)
		isUnhandledPromise = func(
			node *ast.Node,
		) (
			bool, // isUnhandled
			bool, // nonFunctionHandler
			bool, // promiseArray
		) {
			if ast.IsAssignmentExpression(node, false) {
				return false, false, false
			}

			// First, check expressions whose resulting types may not be promise-like
			if ast.IsCommaExpression(node) {
				expr := node.AsBinaryExpression()
				// Any child in a comma expression could return a potentially unhandled
				// promise, so we check them all regardless of whether the final returned
				// value is promise-like.
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.Left)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.Right)
			}

			if !opts.IgnoreVoid && ast.IsVoidExpression(node) {
				// Similarly, a `void` expression always returns undefined, so we need to
				// see what's inside it without checking the type of the overall expression.
				return isUnhandledPromise(node.Expression())
			}

			// Check the type. At this point it can't be unhandled if it isn't a promise
			// or array thereof.

			if isPromiseArray(node) {
				return true, false, true
			}

			// await expression addresses promises, but not promise arrays.
			if ast.IsAwaitExpression(node) {
				// you would think this wouldn't be strictly necessary, since we're
				// anyway checking the type of the expression, but, unfortunately TS
				// reports the result of `await (promise as Promise<number> & number)`
				// as `Promise<number> & number` instead of `number`.
				return false, false, false
			}

			if !isPromiseLike(node, nil) {
				return false, false, false
			}

			if ast.IsCallExpression(node) {
				// If the outer expression is a call, a `.catch()` or `.then()` with
				// rejection handler handles the promise.

				callExpr := node.AsCallExpression()
				callee := callExpr.Expression
				if ast.IsAccessExpression(callee) {
					// TODO(port): getStaticMemberAccessValue -> GetAccessedPropertyName is an
					// enhancement, we should probably add tests for it
					// const methodName = getStaticMemberAccessValue(callee, context);
					methodName, _ := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)

					if methodName == "catch" && len(callExpr.Arguments.Nodes) >= 1 {
						if isValidRejectionHandler(callExpr.Arguments.Nodes[0]) {
							return false, false, false
						}
						return true, true, false
					}
					if methodName == "then" && len(callExpr.Arguments.Nodes) >= 2 {
						if isValidRejectionHandler(callExpr.Arguments.Nodes[1]) {
							return false, false, false
						}
						return true, true, false
					}
					// `x.finally()` is transparent to resolution of the promise, so check `x`.
					// ("object" in this context is the `x` in `x.finally()`)
					if methodName == "finally" {
						return isUnhandledPromise(callee.Expression())
					}
				}

				// All other cases are unhandled.
				return true, false, false
			}

			if node.Kind == ast.KindConditionalExpression {
				expr := node.AsConditionalExpression()
				// We must be getting the promise-like value from one of the branches of the
				// ternary. Check them directly.
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.WhenFalse)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.WhenTrue)
			}

			if ast.IsLogicalOrCoalescingBinaryExpression(node) {
				expr := node.AsBinaryExpression()
				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expr.Left)
				if isUnhandled {
					return isUnhandled, nonFunctionHandler, promiseArray
				}
				return isUnhandledPromise(expr.Right)
			}

			// Anything else is unhandled.
			return true, false, false
		}

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				exprStatement := node.AsExpressionStatement()

				if opts.IgnoreIIFE && isAsyncIife(exprStatement) {
					return
				}

				expression := ast.SkipParentheses(exprStatement.Expression)

				if isKnownSafePromiseReturn(expression) {
					return
				}

				isUnhandled, nonFunctionHandler, promiseArray := isUnhandledPromise(expression)

				if !isUnhandled {
					return
				}
				if promiseArray {
					var msg rule.RuleMessage
					if opts.IgnoreVoid {
						msg = buildFloatingPromiseArrayVoidMessage()
					} else {
						msg = buildFloatingPromiseArrayMessage()
					}
					ctx.ReportNode(node, msg)
				} else if opts.IgnoreVoid {
					var msg rule.RuleMessage
					if nonFunctionHandler {
						msg = buildFloatingUselessRejectionHandlerVoidMessage()
					} else {
						msg = buildFloatingVoidMessage()
					}

					ctx.ReportNodeWithSuggestions(node, msg, func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{
							{
								Message: buildFloatingFixVoidMessage(),
								FixesArr: func() []rule.RuleFix {
									if isHigherPrecedenceThanUnary(exprStatement.Expression) {
										return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, node, "void ")}
									}
									return []rule.RuleFix{
										rule.RuleFixInsertBefore(ctx.SourceFile, node, "void ("),
										rule.RuleFixInsertAfter(expression, ")"),
									}
								}(),
							},
							{
								Message:  buildFloatingFixAwaitMessage(),
								FixesArr: addAwait(expression, exprStatement),
							},
						}
					})
				} else {
					var msg rule.RuleMessage
					if nonFunctionHandler {
						msg = buildFloatingUselessRejectionHandlerMessage()
					} else {
						msg = buildFloatingMessage()
					}
					ctx.ReportNodeWithSuggestions(node, msg, func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildFloatingFixAwaitMessage(),
							FixesArr: addAwait(expression, exprStatement),
						}}
					})
				}
			},
		}
	},
}
