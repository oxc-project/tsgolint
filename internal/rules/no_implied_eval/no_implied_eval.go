package no_implied_eval

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildNoFunctionConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noFunctionConstructor",
		Description: "Implied eval. Do not use the Function constructor to create functions.",
	}
}
func buildNoImpliedEvalErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noImpliedEvalError",
		Description: "Implied eval. Consider passing a function.",
	}
}

var globalCandidates = []string{"global", "globalThis", "window"}
var evalLikeFunctions = []string{"execScript", "setImmediate", "setInterval", "setTimeout"}

var NoImpliedEvalRule = rule.Rule{
	Name: "no-implied-eval",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		getCalleeName := func(node *ast.Expression) string {
			if ast.IsIdentifier(node) {
				return node.AsIdentifier().Text
			}

			if ast.IsAccessExpression(node) && ast.IsIdentifier(node.Expression()) && slices.Contains(globalCandidates, node.Expression().AsIdentifier().Text) {
				if ast.IsPropertyAccessExpression(node) {
					if ast.IsIdentifier(node.Name()) {
						return node.Name().AsIdentifier().Text
					}
				} else if ast.IsElementAccessExpression(node) {
					expr := node.AsElementAccessExpression()
					if ast.IsStringLiteral(expr.ArgumentExpression) {
						return expr.ArgumentExpression.AsStringLiteral().Text
					}
				}
			}

			return ""
		}

		isFunctionType := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			symbol := checker.Type_symbol(t)

			if symbol != nil && utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsFunction|ast.SymbolFlagsMethod) {
				return true
			}

			if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "Function") {
				return true
			}

			return len(utils.GetCallSignatures(ctx.TypeChecker, t)) > 0
		}

		isBind := func(node *ast.Node) bool {
			if ast.IsPropertyAccessExpression(node) {
				node = node.AsPropertyAccessExpression().Name()
			}
			return ast.IsIdentifier(node) && node.AsIdentifier().Text == "bind"
		}

		isFunction := func(node *ast.Node) bool {
			if ast.IsFunctionLike(node) {
				return true
			}
			if ast.IsLiteralExpression(node) {
				return false
			}
			if ast.IsCallExpression(node) {
				return isBind(node.AsCallExpression().Expression) || isFunctionType(node)
			}
			return isFunctionType(node)
		}

		checkImpliedEval := func(
			node *ast.Node,
		) {
			calleeName := getCalleeName(node.Expression())
			if calleeName == "" {
				return
			}

			if calleeName == "Function" {
				t := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
				symbol := checker.Type_symbol(t)

				if symbol != nil {
					if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "FunctionConstructor") {
						ctx.ReportNode(node, buildNoFunctionConstructorMessage())
						return
					}
				} else {
					ctx.ReportNode(node, buildNoFunctionConstructorMessage())
				}
			}

			if len(node.Arguments()) == 0 {
				return
			}

			handler := node.Arguments()[0]

			if slices.Contains(evalLikeFunctions, calleeName) && !isFunction(handler) {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node.Expression())
				if symbol == nil || !utils.Some(symbol.Declarations, func(d *ast.Node) bool {
					return ast.GetSourceFileOfNode(d) == ctx.SourceFile
				}) {
					ctx.ReportNode(handler, buildNoImpliedEvalErrorMessage())
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: checkImpliedEval,
			ast.KindNewExpression:  checkImpliedEval,
		}
	},
}
