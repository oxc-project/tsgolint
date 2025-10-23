package non_nullable_type_assertion_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNonNullAssertion",
		Description: "Use a ! assertion to more succinctly remove null and undefined from the type.",
	}
}

var NonNullableTypeAssertionStyleRule = rule.Rule{
	Name: "non-nullable-type-assertion-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		getTypesIfNotLoose := func(node *ast.Node) []*checker.Type {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			if utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return nil
			}
			return utils.UnionTypeParts(t)
		}

		couldBeNullable := func(t *checker.Type) bool {
			if utils.IsTypeParameter(t) {
				t = checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
				if t == nil {
					return true
				}
			}
			for _, p := range utils.UnionTypeParts(t) {
				if utils.IsTypeFlagSet(p, checker.TypeFlagsNullable) {
					return true
				}
			}
			return false
		}

		checkAssertion := func(node *ast.Node) {
			if ast.IsConstAssertion(node) {
				return
			}

			expression := node.Expression()
			originalTypes := getTypesIfNotLoose(expression)
			if originalTypes == nil {
				return
			}

			typeAnnotation := ast.GetAssertedTypeNode(node)
			assertedTypes := getTypesIfNotLoose(typeAnnotation)
			if assertedTypes == nil {
				return
			}

			nonNullableOriginalType := utils.NewSetWithSizeHint[*checker.Type](len(originalTypes))
			for _, t := range originalTypes {
				if !utils.IsTypeFlagSet(t, checker.TypeFlagsNullable) {
					nonNullableOriginalType.Add(t)
				}
			}
			if nonNullableOriginalType.Len() == len(originalTypes) {
				return
			}

			assertedTypesSet := utils.NewSetWithSizeHint[*checker.Type](len(assertedTypes))
			for _, t := range assertedTypes {
				if couldBeNullable(t) || !nonNullableOriginalType.Has(t) {
					return
				}
				assertedTypesSet.Add(t)
			}

			for originalType := range nonNullableOriginalType.Keys() {
				if !assertedTypesSet.Has(originalType) {
					return
				}
			}

			higherPrecedenceThanUnary := ast.GetExpressionPrecedence(expression) > ast.OperatorPrecedenceUnary

			var removeRange core.TextRange
			if ast.IsAssertionExpression(node) {
				removeRange = node.Loc.WithPos(expression.End())
			} else {
				removeRange = node.Loc.WithEnd(expression.Pos())
			}
			if higherPrecedenceThanUnary {
				ctx.ReportNodeWithFixes(node, buildPreferNonNullAssertionMessage(), func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(removeRange), rule.RuleFixInsertAfter(expression, "!")}
				})
			} else {
				ctx.ReportNodeWithFixes(node, buildPreferNonNullAssertionMessage(), func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(removeRange), rule.RuleFixInsertBefore(ctx.SourceFile, expression, "("), rule.RuleFixInsertAfter(expression, ")!")}
				})
			}
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkAssertion,
			ast.KindTypeAssertionExpression: checkAssertion,
		}
	},
}
