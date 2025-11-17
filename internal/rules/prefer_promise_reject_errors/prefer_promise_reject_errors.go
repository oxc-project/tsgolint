package prefer_promise_reject_errors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildRejectAnErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "rejectAnError",
		Description: "Expected the Promise rejection reason to be an Error.",
	}
}

var PreferPromiseRejectErrorsRule = rule.Rule{
	Name: "prefer-promise-reject-errors",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferPromiseRejectErrorsOptions](options, "prefer-promise-reject-errors")

		checkRejectCall := func(callExpression *ast.CallExpression) {
			if len(callExpression.Arguments.Nodes) != 0 {
				argument := callExpression.Arguments.Nodes[0]
				t := ctx.TypeChecker.GetTypeAtLocation(argument)

				if opts.AllowThrowingAny && utils.IsTypeAnyType(t) {
					return
				}
				if opts.AllowThrowingUnknown && utils.IsTypeUnknownType(t) {
					return
				}

				if utils.IsErrorLike(ctx.Program, ctx.TypeChecker, t) || utils.IsReadonlyErrorLike(ctx.Program, ctx.TypeChecker, t) {
					return
				}
			} else if opts.AllowEmptyReject {
				return
			}
			ctx.ReportNode(&callExpression.Node, buildRejectAnErrorMessage())
		}

		typeAtLocationIsLikePromise := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			return (utils.IsPromiseConstructorLike(ctx.Program, ctx.TypeChecker, t) || utils.IsPromiseLike(ctx.Program, ctx.TypeChecker, t))
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				callee := ast.SkipParentheses(expr.Expression)

				// Promise.reject(...)
				if ast.IsAccessExpression(callee) {
					// TODO(port): getStaticMemberAccessValue -> GetAccessedPropertyName is an
					// enhancement, we should probably add tests for it
					methodName, _ := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
					if methodName == "reject" && typeAtLocationIsLikePromise(callee.Expression()) {
						checkRejectCall(expr)
					}
					// reject(...)
				} else if ast.IsIdentifier(callee) {
					symbol := ctx.TypeChecker.GetSymbolAtLocation(callee)
					if symbol == nil {
						return
					}
					param := symbol.ValueDeclaration

					if param == nil || !ast.IsParameter(param) {
						return
					}

					parentNode := symbol.ValueDeclaration.Parent
					if !ast.IsFunctionExpression(parentNode) && !ast.IsArrowFunction(parentNode) {
						return
					}

					params := parentNode.Parameters()
					if len(params) < 2 || params[1] != param {
						return
					}

					for {
						parentNode = parentNode.Parent
						if ast.IsParenthesizedExpression(parentNode) {
							continue
						}
						if !ast.IsNewExpression(parentNode) {
							return
						}
						break
					}

					if !utils.IsPromiseConstructorLike(ctx.Program, ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(parentNode.Expression())) {
						return
					}

					checkRejectCall(expr)
				}
			},
		}
	},
}
