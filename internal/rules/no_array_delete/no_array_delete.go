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

				deleteTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				arrayExprRange := utils.TrimNodeTextRange(ctx.SourceFile, expression.Expression)

				ctx.ReportDiagnosticWithSuggestions(rule.RuleDiagnostic{
					Range:   deleteTokenRange,
					Message: buildNoArrayDeleteMessage(),
					LabeledRanges: []rule.RuleLabeledRange{
						{Label: "This expression evaluates to an array.", Range: arrayExprRange},
					},
				}, func() []rule.RuleSuggestion {
					argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, expression.ArgumentExpression)

					leftBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arrayExprRange.End())
					rightBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, argumentRange.End())

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
