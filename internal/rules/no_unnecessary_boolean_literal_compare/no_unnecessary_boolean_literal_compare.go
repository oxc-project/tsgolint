package no_unnecessary_boolean_literal_compare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildComparingNullableToFalseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparingNullableToFalse",
		Description: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
	}
}
func buildComparingNullableToTrueDirectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparingNullableToTrueDirect",
		Description: "This expression unnecessarily compares a nullable boolean value to true instead of using it directly.",
	}
}
func buildComparingNullableToTrueNegatedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparingNullableToTrueNegated",
		Description: "This expression unnecessarily compares a nullable boolean value to true instead of negating it.",
	}
}
func buildDirectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "direct",
		Description: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
	}
}
func buildNegatedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "negated",
		Description: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
	}
}
func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

type NoUnnecessaryBooleanLiteralCompareOptions struct {
	AllowComparingNullableBooleansToFalse                  *bool
	AllowComparingNullableBooleansToTrue                   *bool
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool
}

type booleanComparison struct {
	expression                  *ast.Expression
	literalBooleanInComparison  bool
	negated                     bool
	expressionIsNullableBoolean bool
}

func isBooleanType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike)
}

var NoUnnecessaryBooleanLiteralCompareRule = rule.Rule{
	Name: "no-unnecessary-boolean-literal-compare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoUnnecessaryBooleanLiteralCompareOptions)
		if !ok {
			opts = NoUnnecessaryBooleanLiteralCompareOptions{}
		}
		if opts.AllowComparingNullableBooleansToFalse == nil {
			opts.AllowComparingNullableBooleansToFalse = utils.Ref(true)
		}
		if opts.AllowComparingNullableBooleansToTrue == nil {
			opts.AllowComparingNullableBooleansToTrue = utils.Ref(true)
		}
		if opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(false)
		}

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		/**
		 * checks if the expressionType is a union that
		 *   1) contains at least one nullish type (null or undefined)
		 *   2) contains at least once boolean type (true or false or boolean)
		 *   3) does not contain any types besides nullish and boolean types
		 */
		isNullableBoolean := func(t *checker.Type) bool {
			if !utils.IsUnionType(t) {
				return false
			}

			var flags checker.TypeFlags
			for _, t := range t.Types() {
				flags |= checker.Type_flags(t)
			}
			return flags&checker.TypeFlagsNullable != 0 && flags&checker.TypeFlagsBooleanLike != 0
		}

		getBooleanComparison := func(node *ast.BinaryExpression) (booleanComparison, bool) {
			res := booleanComparison{}
			if node.OperatorToken.Kind == ast.KindExclamationEqualsToken || node.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
				res.negated = true
			} else if node.OperatorToken.Kind != ast.KindEqualsEqualsToken && node.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
				return res, false
			}

			deconstruct := func(against, expression *ast.Node) bool {
				if against.Kind == ast.KindTrueKeyword {
					res.literalBooleanInComparison = true
				} else if against.Kind != ast.KindFalseKeyword {
					return false
				}
				res.expression = expression
				return true
			}

			if !deconstruct(ast.SkipParentheses(node.Right), node.Left) && !deconstruct(ast.SkipParentheses(node.Left), node.Right) {
				return res, false
			}

			constraintType, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(res.expression))

			if isTypeParameter && constraintType == nil {
				return res, false
			}

			if isBooleanType(constraintType) {
				return res, true
			}

			if isNullableBoolean(constraintType) {
				res.expressionIsNullableBoolean = true
				return res, true
			}

			return res, false
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				comparison, found := getBooleanComparison(node.AsBinaryExpression())
				if !found {
					return
				}

				if comparison.expressionIsNullableBoolean && ((comparison.literalBooleanInComparison && *opts.AllowComparingNullableBooleansToTrue) || (!comparison.literalBooleanInComparison && *opts.AllowComparingNullableBooleansToFalse)) {
					return
				}

				var msg rule.RuleMessage
				if comparison.expressionIsNullableBoolean {
					if comparison.literalBooleanInComparison {
						if comparison.negated {
							msg = buildComparingNullableToTrueNegatedMessage()
						} else {
							msg = buildComparingNullableToTrueDirectMessage()
						}
					} else {
						msg = buildComparingNullableToFalseMessage()
					}
				} else if comparison.negated {
					msg = buildNegatedMessage()
				} else {
					msg = buildDirectMessage()
				}

				parent := node.Parent
				for ast.IsParenthesizedExpression(parent) {
					parent = parent.Parent
				}
				isUnaryNegation := ast.IsPrefixUnaryExpression(parent) && parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken

				shouldNegate := comparison.negated != comparison.literalBooleanInComparison

				mutatedNode := node
				if isUnaryNegation {
					mutatedNode = parent
				}

				fixes := make([]rule.RuleFix, 0, 6)

				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, mutatedNode, ctx.SourceFile.Text()[comparison.expression.Pos():comparison.expression.End()]))

				if shouldNegate == isUnaryNegation {
					fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, mutatedNode, "!"))

					if !utils.IsStrongPrecedenceNode(comparison.expression) {
						fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, mutatedNode, "("), rule.RuleFixInsertAfter(mutatedNode, ")"))
					}
				}

				if comparison.expressionIsNullableBoolean && !comparison.literalBooleanInComparison {
					fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, mutatedNode, "("), rule.RuleFixInsertAfter(mutatedNode, " ?? true)"))
				}

				ctx.ReportNodeWithFixes(node, msg, func() []rule.RuleFix { return fixes })
			},
		}
	},
}
