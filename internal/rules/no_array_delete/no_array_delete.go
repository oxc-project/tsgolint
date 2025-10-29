package no_array_delete

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildNoArrayDeleteMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noArrayDelete",
		Description: "Using the `delete` operator with an array expression is unsafe.",
	}
}
func buildUseSpliceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useSplice",
		Description: "Use `array.splice()` instead.",
	}
}

var NoArrayDeleteRule = rule.Rule{
	Name: "no-array-delete",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		isUnderlyingTypeArray := func(t *checker.Type) bool {
			if utils.IsTypeFlagSet(t, checker.TypeFlagsUnion) {
				for _, t := range t.Types() {
					if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) {
						return false
					}
				}
				return true
			}

			if utils.IsTypeFlagSet(t, checker.TypeFlagsIntersection) {
				for _, t := range t.Types() {
					if checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) {
						return true
					}
				}
				return false
			}

			return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
		}

		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				if node.Kind != ast.KindDeleteExpression {
					return
				}
				deleteExpression := ast.SkipParentheses(node.AsDeleteExpression().Expression)

				if !ast.IsElementAccessExpression(deleteExpression) {
					return
				}

				expression := deleteExpression.AsElementAccessExpression()

				argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression.Expression)

				if !isUnderlyingTypeArray(argType) {
					return
				}

				expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, expression.Expression)
				argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, expression.ArgumentExpression)

				deleteTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				leftBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expressionRange.End())
				rightBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, argumentRange.End())

				ctx.ReportNodeWithSuggestions(node, buildNoArrayDeleteMessage(), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message: buildUseSpliceMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixRemoveRange(deleteTokenRange),
							rule.RuleFixReplaceRange(leftBracketTokenRange, ".splice("),
							rule.RuleFixReplaceRange(rightBracketTokenRange, ", 1)"),
						},
					}}
				})
			},
		}
	},
}
