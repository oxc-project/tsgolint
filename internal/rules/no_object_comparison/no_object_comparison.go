package no_object_comparison

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildObjectComparisonMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "objectComparison",
		Description: "Do not compare objects - use one of eq, gt, gte, lt, lte functions instead.",
	}
}

func isComparisonOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindLessThanToken,
		ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken,
		ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindExclamationEqualsEqualsToken:
		return true
	default:
		return false
	}
}

func isNullOrUndefined(node *ast.Node) bool {
	unwrapped := ast.SkipParentheses(node)
	return utils.IsNullLiteralOrUndefinedIdentifier(unwrapped)
}

func getConstrainedType(typeChecker *checker.Checker, node *ast.Node) *checker.Type {
	constraintType, isTypeParameter := utils.GetConstraintInfo(typeChecker, typeChecker.GetTypeAtLocation(node))
	if isTypeParameter && constraintType == nil {
		return nil
	}
	return constraintType
}

func isDisallowedNamedObjectType(t *checker.Type, classNames *utils.Set[string]) bool {
	for _, part := range utils.UnionTypeParts(t) {
		current := part
		if utils.IsTypeFlagSet(current, checker.TypeFlagsObject) && checker.Type_objectFlags(current)&checker.ObjectFlagsReference != 0 {
			current = current.Target()
		}

		if !utils.IsTypeFlagSet(current, checker.TypeFlagsObject) || checker.Type_objectFlags(current)&checker.ObjectFlagsClassOrInterface == 0 {
			continue
		}

		symbol := checker.Type_symbol(current)
		if symbol != nil && classNames.Has(symbol.Name) {
			return true
		}
	}

	return false
}

var NoObjectComparisonRule = rule.Rule{
	Name: "no-object-comparison",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoObjectComparisonOptions](options, "no-object-comparison")

		disallowedNames := utils.NewSetWithSizeHint[string](len(opts.ClassNames))
		for _, className := range opts.ClassNames {
			disallowedNames.Add(className)
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				if !isComparisonOperator(expr.OperatorToken.Kind) {
					return
				}

				if isNullOrUndefined(expr.Left) || isNullOrUndefined(expr.Right) {
					return
				}

				leftType := getConstrainedType(ctx.TypeChecker, expr.Left)
				rightType := getConstrainedType(ctx.TypeChecker, expr.Right)

				if isDisallowedNamedObjectType(leftType, disallowedNames) || isDisallowedNamedObjectType(rightType, disallowedNames) {
					ctx.ReportNode(node, buildObjectComparisonMessage())
				}
			},
		}
	},
}
