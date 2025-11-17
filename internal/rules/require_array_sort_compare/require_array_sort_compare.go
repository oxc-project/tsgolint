package require_array_sort_compare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildRequireCompareMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireCompare",
		Description: "Require 'compare' argument.",
	}
}

var RequireArraySortCompareRule = rule.Rule{
	Name: "require-array-sort-compare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[RequireArraySortCompareOptions](options, "require-array-sort-compare")

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if len(expr.Arguments.Nodes) != 0 {
					return
				}
				callee := expr.Expression

				if !ast.IsAccessExpression(callee) {
					return
				}

				if propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee); !found || (propertyName != "sort" && propertyName != "toSorted") {
					return
				}

				calleeObjType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				if opts.IgnoreStringArrays && checker.Checker_isArrayOrTupleType(ctx.TypeChecker, calleeObjType) {
					if utils.Every(checker.Checker_getTypeArguments(ctx.TypeChecker, calleeObjType), func(t *checker.Type) bool {
						return utils.IsTypeFlagSet(t, checker.TypeFlagsString)
					}) {
						return
					}
				}

				if utils.Every(utils.UnionTypeParts(calleeObjType), func(t *checker.Type) bool {
					return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
				}) {
					ctx.ReportNode(node, buildRequireCompareMessage())
				}
			},
		}
	},
}
