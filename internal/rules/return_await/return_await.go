package return_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildDisallowedPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "disallowedPromiseAwait",
		Description: "Returning an awaited promise is not allowed in this context.",
	}
}
func buildDisallowedPromiseAwaitSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "disallowedPromiseAwaitSuggestion",
		Description: "Remove `await` before the expression. Use caution as this may impact control flow.",
	}
}
func buildNonPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonPromiseAwait",
		Description: "Returning an awaited value that is not a promise is not allowed.",
	}
}
func buildRequiredPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requiredPromiseAwait",
		Description: "Returning an awaited promise is required in this context.",
	}
}
func buildRequiredPromiseAwaitSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requiredPromiseAwaitSuggestion",
		Description: "Add `await` before the expression. Use caution as this may impact control flow.",
	}
}

type scopeInfo struct {
	hasAsync   bool
	owningFunc *ast.Node
	parent     *scopeInfo
}

type containingTryStatementBlock uint8

const (
	containingTryStatementBlockCatch containingTryStatementBlock = iota
	containingTryStatementBlockFinally
	containingTryStatementBlockTry
)

type whetherToAwait uint8

const (
	whetherToAwaitDontCare whetherToAwait = iota
	whetherToAwaitAwait
	whetherToAwaitNoAwait
)

func getWhetherToAwait(affectsErrorHandling bool, option ReturnAwaitOptionsOption) whetherToAwait {
	switch option {
	case ReturnAwaitOptionsOptionAlways:
		return whetherToAwaitAwait
	case ReturnAwaitOptionsOptionErrorHandlingCorrectnessOnly:
		if affectsErrorHandling {
			return whetherToAwaitAwait
		}
		return whetherToAwaitDontCare
	case ReturnAwaitOptionsOptionInTryCatch:
		if affectsErrorHandling {
			return whetherToAwaitAwait
		}
		return whetherToAwaitNoAwait
	case ReturnAwaitOptionsOptionNever:
		return whetherToAwaitNoAwait
	default:
		panic("unexpected ReturnAwaitOptionsOption")
	}
}

var ReturnAwaitRule = rule.Rule{
	Name: "return-await",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[ReturnAwaitOptions](options, "return-await")

		var scope *scopeInfo

		enterFunction := func(node *ast.Node) {
			scope = &scopeInfo{
				hasAsync:   ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync),
				owningFunc: node,
				parent:     scope,
			}
		}
		exitFunction := func(_node *ast.Node) {
			scope = scope.parent
		}

		affectsExplicitResourceManagement := func(node *ast.Node) bool {
			if !ast.IsBlock(scope.owningFunc.Body()) {
				return false
			}

			for declarationScope := ast.GetEnclosingBlockScopeContainer(node); declarationScope != nil; declarationScope = ast.GetEnclosingBlockScopeContainer(declarationScope) {
				locals := declarationScope.Locals()
				if locals != nil {
					for _, local := range locals {
						decl := local.ValueDeclaration
						// if it's a using/await using declaration, and it comes _before_ the
						// node we're checking, it affects control flow for that node.

						if decl != nil && ast.IsVariableDeclaration(decl) && decl.Parent.Flags&ast.NodeFlagsUsing != 0 && decl.Pos() < node.Pos() {
							return true
						}
					}
				}
				if scope.owningFunc == declarationScope {
					break
				}
			}

			return false
		}

		/**
		 * A try _statement_ is the whole thing that encompasses try block,
		 * catch clause, and finally block. This function finds the nearest
		 * enclosing try statement (if present) for a given node, and reports which
		 * part of the try statement the node is in.
		 */
		findContainingTryStatement := func(node *ast.Node) (containingTryStatementBlock, *ast.TryStatement, bool) {
			child := node
			ancestor := node.Parent

			for ancestor != nil && !ast.IsFunctionLike(ancestor) {
				if !ast.IsTryStatement(ancestor) {
					child = ancestor
					ancestor = ancestor.Parent
					continue
				}

				statement := ancestor.AsTryStatement()

				var block containingTryStatementBlock
				switch {
				case child == statement.TryBlock:
					block = containingTryStatementBlockTry
				case child == statement.CatchClause:
					block = containingTryStatementBlockCatch
				case child == statement.FinallyBlock:
					block = containingTryStatementBlockFinally
				}

				return block, statement, true
			}

			return 0, nil, false
		}

		/**
		 * Tests whether a node is inside of an explicit error handling context
		 * (try/catch/finally) in a way that throwing an exception will have an
		 * impact on the program's control flow.
		 */
		affectsExplicitErrorHandling := func(node *ast.Node) bool {
			for {
				// If an error-handling block is followed by another error-handling block,
				// control flow is affected by whether promises in it are awaited or not.
				// Otherwise, we need to check recursively for nested try statements until
				// we get to the top level of a function or the program. If by then,
				// there's no offending error-handling blocks, it doesn't affect control
				// flow.
				containingTryBlock, tryStatement, found := findContainingTryStatement(node)
				if !found {
					return false
				}

				switch containingTryBlock {
				case containingTryStatementBlockCatch:
					// Exceptions thrown in catch blocks followed by a finally block affect
					// control flow.
					if tryStatement.FinallyBlock != nil {
						return true
					}
					// Otherwise recurse.
					node = tryStatement.AsNode()
				case containingTryStatementBlockFinally:
					node = tryStatement.AsNode()
				case containingTryStatementBlockTry:
					// Try blocks are always followed by either a catch or finally,
					// so exceptions thrown here always affect control flow.
					return true
				}
			}
		}

		removeAwaitFix := func(node *ast.Node) rule.RuleFix {
			return rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos()))
		}
		insertAwaitFix := func(node *ast.Node, isHighPrecedence bool) []rule.RuleFix {
			if isHighPrecedence {
				return []rule.RuleFix{
					rule.RuleFixInsertBefore(ctx.SourceFile, node, "await "),
				}
			}
			return []rule.RuleFix{
				rule.RuleFixInsertBefore(ctx.SourceFile, node, "await ("),
				rule.RuleFixInsertAfter(node, ")"),
			}
		}

		test := func(node *ast.Node) {
			var child *ast.Node
			isAwait := ast.IsAwaitExpression(node)

			if isAwait {
				child = node.Expression()
			} else {
				child = node
			}

			t := ctx.TypeChecker.GetTypeAtLocation(child)
			certainty := utils.NeedsToBeAwaited(ctx.TypeChecker, node, t)

			if certainty != utils.TypeAwaitableAlways {
				if isAwait {
					if certainty == utils.TypeAwaitableMay {
						return
					}

					ctx.ReportNodeWithFixes(node, buildNonPromiseAwaitMessage(), func() []rule.RuleFix { return []rule.RuleFix{removeAwaitFix(node)} })
				}
				return
			}

			// At this point it's definitely a thenable.
			affectsErrorHandling := affectsExplicitErrorHandling(node) || affectsExplicitResourceManagement(node)
			useAutoFix := !affectsErrorHandling

			shouldAwaitInCurrentContext := getWhetherToAwait(affectsErrorHandling, opts.Option)

			switch shouldAwaitInCurrentContext {
			case whetherToAwaitAwait:
				if isAwait {
					break
				}
				rule.ReportNodeWithFixesOrSuggestions(ctx, node, useAutoFix, buildRequiredPromiseAwaitMessage(), buildRequiredPromiseAwaitSuggestionMessage(), insertAwaitFix(node, utils.IsHigherPrecedenceThanAwait(node))...)
			case whetherToAwaitDontCare:
				break
			case whetherToAwaitNoAwait:
				if !isAwait {
					break
				}
				rule.ReportNodeWithFixesOrSuggestions(ctx, node, useAutoFix, buildDisallowedPromiseAwaitMessage(), buildDisallowedPromiseAwaitSuggestionMessage(), removeAwaitFix(node))
			}
		}

		var testEachPossiblyReturnedNode func(node *ast.Node)
		testEachPossiblyReturnedNode = func(node *ast.Node) {
			node = ast.SkipParentheses(node)
			if node.Kind == ast.KindConditionalExpression {
				expr := node.AsConditionalExpression()
				testEachPossiblyReturnedNode(expr.WhenFalse)
				testEachPossiblyReturnedNode(expr.WhenTrue)
			} else {
				test(node)
			}
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:       enterFunction,
			ast.KindFunctionDeclaration: enterFunction,
			ast.KindFunctionExpression:  enterFunction,
			ast.KindMethodDeclaration:   enterFunction,
			ast.KindConstructor:         enterFunction,
			ast.KindGetAccessor:         enterFunction,
			ast.KindSetAccessor:         enterFunction,

			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				body := node.Body()
				if !ast.IsBlock(body) {
					testEachPossiblyReturnedNode(body)
				}

				exitFunction(node)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,

			ast.KindReturnStatement: func(node *ast.Node) {
				expr := node.AsReturnStatement().Expression
				if scope == nil || !scope.hasAsync || expr == nil {
					return
				}

				testEachPossiblyReturnedNode(expr)
			},
		}
	},
}
