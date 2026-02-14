package no_for_in_array

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildForInViolationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "forInViolation",
		Description: "For-in loops over arrays skips holes, returns indices as strings, and may visit the prototype chain or other enumerable properties.",
		Help:        "Use a more robust iteration method such as for-of or array.forEach instead.",
	}
}

var NoForInArrayRule = rule.Rule{
	Name: "no-for-in-array",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		hasArrayishLength := func(t *checker.Type) bool {
			lengthProperty := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "length")
			if lengthProperty == nil {
				return false
			}

			return utils.IsTypeFlagSet(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, lengthProperty), checker.TypeFlagsNumberLike)
		}
		isArrayLike := func(t *checker.Type) bool {
			return utils.TypeRecurser(t, func(t *checker.Type) bool {
				return utils.GetNumberIndexType(ctx.TypeChecker, t) != nil && hasArrayishLength(t)
			})
		}

		return rule.RuleListeners{
			ast.KindForInStatement: func(node *ast.Node) {
				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.AsForInOrOfStatement().Expression)

				if isArrayLike(t) {
					ctx.ReportRange(
						utils.GetForStatementHeadLoc(ctx.SourceFile, node),
						buildForInViolationMessage(),
					)
				}
			},
		}
	},
}
