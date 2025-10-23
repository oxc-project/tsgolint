package prefer_reduce_type_parameter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferTypeParameterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferTypeParameter",
		Description: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
	}
}

var PreferReduceTypeParameterRule = rule.Rule{
	Name: "prefer-reduce-type-parameter",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if len(expr.Arguments.Nodes) < 2 {
					return
				}

				secondArg := expr.Arguments.Nodes[1]

				if secondArg.Kind != ast.KindAsExpression && secondArg.Kind != ast.KindTypeAssertionExpression {
					return
				}

				callee := expr.Expression
				if !ast.IsAccessExpression(callee) {
					return
				}

				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !found || propertyName != "reduce" {
					return
				}

				assertionExpr := secondArg.Expression()
				assertionType := secondArg.Type()

				initializerType := ctx.TypeChecker.GetTypeAtLocation(assertionExpr)
				assertedType := ctx.TypeChecker.GetTypeAtLocation(assertionType)

				// don't report this if the resulting fix will be a type error
				if !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, initializerType, assertedType) {
					return
				}

				calleeObjType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				if utils.TypeRecurser(calleeObjType, func(t *checker.Type) bool {
					return !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
				}) {
					return
				}

				fixes := make([]rule.RuleFix, 0, 2)
				if secondArg.Kind == ast.KindAsExpression {
					fixes = append(fixes, rule.RuleFixRemoveRange(assertionType.Loc.WithPos(assertionExpr.End())))
				} else {
					fixes = append(fixes, rule.RuleFixRemoveRange(secondArg.Loc.WithEnd(assertionExpr.Pos())))
				}
				if expr.TypeArguments == nil {
					fixes = append(fixes, rule.RuleFixInsertAfter(callee, "<"+ctx.SourceFile.Text()[assertionType.Pos():assertionType.End()]+">"))
				}
				ctx.ReportNodeWithFixes(secondArg, buildPreferTypeParameterMessage(), func() []rule.RuleFix { return fixes })
			},
		}
	},
}
