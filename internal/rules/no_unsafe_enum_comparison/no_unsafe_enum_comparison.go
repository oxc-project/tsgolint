package no_unsafe_enum_comparison

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMismatchedCaseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatchedCase",
		Description: "The case statement does not have a shared enum type with the switch predicate.",
	}
}
func buildMismatchedConditionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatchedCondition",
		Description: "The two values in this comparison do not have a shared enum type.",
	}
}

// func buildReplaceValueWithEnumMessage() rule.RuleMessage {
// return rule.RuleMessage{
// Id:          "replaceValueWithEnum",
// Description: "Replace with an enum value comparison.",
// }
// }

/**
 * @returns What type a type's enum value is (number or string), if either.
 */
func getEnumValueType(t *checker.Type) checker.TypeFlags {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
		if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
			return checker.TypeFlagsNumber
		}
		return checker.TypeFlagsString
	}
	return checker.TypeFlagsNone
}

/**
 * @returns Whether the right type is an unsafe comparison against any left type.
 */
func typeViolates(leftTypeParts []*checker.Type, rightType *checker.Type) bool {
	rightNumberLike, rightStringLike := false, false
	for _, typePart := range utils.IntersectionTypeParts(rightType) {
		if utils.IsTypeFlagSet(typePart, checker.TypeFlagsNumberLike) {
			rightNumberLike = true
		}
		if utils.IsTypeFlagSet(typePart, checker.TypeFlagsStringLike) {
			rightStringLike = true
		}
		if rightNumberLike && rightStringLike {
			break
		}
	}
	if !rightNumberLike && !rightStringLike {
		return false
	}

	for _, typePart := range leftTypeParts {
		t := getEnumValueType(typePart)
		if t == checker.TypeFlagsNumber && rightNumberLike {
			return true
		} else if t == checker.TypeFlagsString && rightStringLike {
			return true
		}
	}
	return false
}

var NoUnsafeEnumComparisonRule = rule.Rule{
	Name: "no-unsafe-enum-comparison",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		isMismatchedComparison := func(
			leftType *checker.Type,
			rightType *checker.Type,
		) bool {
			// Allow comparisons that don't have anything to do with enums:
			//
			// ```ts
			// 1 === 2;
			// ```
			leftEnumTypes := utils.GetEnumTypes(ctx.TypeChecker, leftType)
			rightEnumTypes := utils.NewSetFromItems(utils.GetEnumTypes(ctx.TypeChecker, rightType)...)
			if len(leftEnumTypes) == 0 && rightEnumTypes.Len() == 0 {
				return false
			}

			// Allow comparisons that share an enum type:
			//
			// ```ts
			// Fruit.Apple === Fruit.Banana;
			// ```
			for _, leftEnumType := range leftEnumTypes {
				if rightEnumTypes.Has(leftEnumType) {
					return false
				}
			}

			// We need to split the type into the union type parts in order to find
			// valid enum comparisons like:
			//
			// ```ts
			// declare const something: Fruit | Vegetable;
			// something === Fruit.Apple;
			// ```
			leftTypeParts := utils.UnionTypeParts(leftType)
			rightTypeParts := utils.UnionTypeParts(rightType)

			// If a type exists in both sides, we consider this comparison safe:
			//
			// ```ts
			// declare const fruit: Fruit.Apple | 0;
			// fruit === 0;
			// ```
			for _, leftTypePart := range leftTypeParts {
				if slices.Contains(rightTypeParts, leftTypePart) {
					return false
				}
			}

			l := typeViolates(leftTypeParts, rightType)

			return (l || typeViolates(rightTypeParts, leftType))
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				opKind := expr.OperatorToken.Kind
				if !(opKind == ast.KindLessThanToken || opKind == ast.KindLessThanEqualsToken || opKind == ast.KindGreaterThanToken || opKind == ast.KindGreaterThanEqualsToken || opKind == ast.KindEqualsEqualsToken || opKind == ast.KindEqualsEqualsEqualsToken || opKind == ast.KindExclamationEqualsToken || opKind == ast.KindExclamationEqualsEqualsToken) {
					return
				}

				leftType := ctx.TypeChecker.GetTypeAtLocation(expr.Left)
				rightType := ctx.TypeChecker.GetTypeAtLocation(expr.Right)

				if isMismatchedComparison(leftType, rightType) {
					// TODO(port): port suggestion
					ctx.ReportNode(node, buildMismatchedConditionMessage())
				}
			},

			ast.KindCaseClause: func(node *ast.Node) {
				leftType := ctx.TypeChecker.GetTypeAtLocation(node.Parent.Parent.Expression())
				rightType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())

				if isMismatchedComparison(leftType, rightType) {
					ctx.ReportNode(node, buildMismatchedCaseMessage())
				}
			},
		}
	},
}
