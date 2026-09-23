package no_unnecessary_type_assertion

import (
	"slices"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// These checks supplement typescript-eslint's contextual fallback for issue #1122.
// A receiver accepting the original type does not imply that removing the assertion
// preserves the receiver's type in later statements. Keep this analysis separate
// from the upstream rule and use typescript-go's assignment reduction semantics.

func getDestructuringDefault(target *ast.Node, propertyPath []string) *ast.Node {
	for i := len(propertyPath) - 1; i >= 0; i-- {
		target = ast.SkipParentheses(target)
		var element *ast.Node
		if ast.IsArrayLiteralExpression(target) {
			index, err := strconv.Atoi(propertyPath[i])
			elements := target.AsArrayLiteralExpression().Elements.Nodes
			if err != nil || index < 0 || index >= len(elements) {
				return nil
			}
			element = elements[index]
		} else if ast.IsObjectLiteralExpression(target) {
			for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
				if ast.IsSpreadAssignment(property) || property.Name() == nil {
					continue
				}
				name, known := ast.TryGetTextOfPropertyName(property.Name())
				if !known {
					return nil
				}
				if name == propertyPath[i] {
					// Every target reading this property must preserve the narrowing.
					// A default on only one of several targets is not sufficient.
					if element != nil {
						return nil
					}
					element = property
				}
			}
		}
		if element == nil {
			return nil
		}
		if i == 0 {
			if ast.IsShorthandPropertyAssignment(element) {
				return element.AsShorthandPropertyAssignment().ObjectAssignmentInitializer
			}
			if ast.IsPropertyAssignment(element) {
				element = element.Initializer()
			}
			if ast.IsAssignmentExpression(element, true) {
				return element.AsBinaryExpression().Right
			}
			return nil
		}
		target = ast.GetTargetOfBindingOrAssignmentElement(element)
		if target == nil {
			return nil
		}
	}
	return nil
}

func getLogicalResultType(ctx rule.RuleContext, operator ast.Kind, left, right *checker.Type) *checker.Type {
	switch operator {
	case ast.KindAmpersandAmpersandToken:
		if checker.Checker_hasTypeFacts(ctx.TypeChecker, left, checker.TypeFactsTruthy) {
			falsySource := left
			if !utils.IsStrictCompilerOptionEnabled(ctx.Program.Options(), ctx.Program.Options().StrictNullChecks) {
				falsySource = checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, right)
			}
			falsy := checker.Checker_extractDefinitelyFalsyTypes(ctx.TypeChecker, falsySource)
			return checker.Checker_getUnionTypeEx(ctx.TypeChecker, []*checker.Type{falsy, right}, checker.UnionReductionLiteral, nil, nil)
		}
	case ast.KindBarBarToken:
		if checker.Checker_hasTypeFacts(ctx.TypeChecker, left, checker.TypeFactsFalsy) {
			truthy := checker.Checker_GetNonNullableType(ctx.TypeChecker, checker.Checker_removeDefinitelyFalsyTypes(ctx.TypeChecker, left))
			return checker.Checker_getUnionTypeEx(ctx.TypeChecker, []*checker.Type{truthy, right}, checker.UnionReductionSubtype, nil, nil)
		}
	case ast.KindQuestionQuestionToken:
		if checker.Checker_hasTypeFacts(ctx.TypeChecker, left, checker.TypeFactsEQUndefinedOrNull) {
			nonNullable := checker.Checker_GetNonNullableType(ctx.TypeChecker, left)
			return checker.Checker_getUnionTypeEx(ctx.TypeChecker, []*checker.Type{nonNullable, right}, checker.UnionReductionSubtype, nil, nil)
		}
	}
	return left
}

func isInNarrowingAssignment(ctx rule.RuleContext, node *ast.Node, uncastType, castType *checker.Type) bool {
	inLiteral := false
	var propertyPath []string
	pathKnown := true
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if !inLiteral && ast.IsLogicalOrCoalescingBinaryExpression(parent) && parent.AsBinaryExpression().Right == current {
			binary := parent.AsBinaryExpression()
			// The left operand can also supply the value assigned to the receiver.
			leftType := ctx.TypeChecker.GetTypeAtLocation(binary.Left)
			uncastType = getLogicalResultType(ctx, binary.OperatorToken.Kind, leftType, uncastType)
			castType = getLogicalResultType(ctx, binary.OperatorToken.Kind, leftType, castType)
		}
		if pathKnown && ast.IsConditionalExpression(parent) && parent.AsConditionalExpression().Condition != current {
			conditional := parent.AsConditionalExpression()
			other := conditional.WhenTrue
			if other == current {
				other = conditional.WhenFalse
			}
			// Either branch can supply the assigned value. Preserve the other
			// branch's types when comparing narrowing with and without the cast.
			otherType := ctx.TypeChecker.GetTypeAtLocation(other)
			for i := len(propertyPath) - 1; i >= 0 && otherType != nil; i-- {
				propertyType := checker.Checker_getTypeOfPropertyOfType(ctx.TypeChecker, otherType, propertyPath[i])
				if propertyType == nil {
					if _, err := strconv.Atoi(propertyPath[i]); err == nil {
						propertyType = utils.GetNumberIndexType(ctx.TypeChecker, otherType)
					}
				}
				otherType = propertyType
			}
			if otherType != nil {
				uncastType = checker.Checker_getUnionTypeEx(ctx.TypeChecker, []*checker.Type{uncastType, otherType}, checker.UnionReductionSubtype, nil, nil)
				castType = checker.Checker_getUnionTypeEx(ctx.TypeChecker, []*checker.Type{castType, otherType}, checker.UnionReductionSubtype, nil, nil)
			}
		}
		if !inLiteral && ast.IsBinaryExpression(parent) && parent.AsBinaryExpression().Left == current &&
			(parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken || parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken) {
			// These operators already discard nullish values from their left
			// operand, so a purely non-nullable assertion cannot change narrowing.
			uncastNonNullable := checker.Checker_GetNonNullableType(ctx.TypeChecker, uncastType)
			castNonNullable := checker.Checker_GetNonNullableType(ctx.TypeChecker, castType)
			if checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, uncastNonNullable, castNonNullable) {
				return false
			}
		}
		if ast.IsParenthesizedExpression(parent) || ast.IsLogicalOrCoalescingBinaryExpression(parent) ||
			(ast.IsConditionalExpression(parent) && parent.AsConditionalExpression().Condition != current) ||
			(ast.IsBinaryExpression(parent) && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken && parent.AsBinaryExpression().Right == current) {
			continue
		}
		if ast.IsArrayLiteralExpression(parent) || ast.IsObjectLiteralExpression(parent) || ast.IsSpreadAssignment(parent) ||
			(ast.IsPropertyAssignment(parent) && parent.Initializer() == current) {
			inLiteral = true
			if ast.IsPropertyAssignment(parent) {
				name, known := ast.TryGetTextOfPropertyName(parent.Name())
				pathKnown = pathKnown && known
				propertyPath = append(propertyPath, name)
			} else if ast.IsArrayLiteralExpression(parent) {
				elements := parent.AsArrayLiteralExpression().Elements.Nodes
				index := slices.Index(elements, current)
				// A spread can shift the runtime index away from the syntax index.
				pathKnown = pathKnown && index >= 0 && !slices.ContainsFunc(elements[:index], ast.IsSpreadElement)
				propertyPath = append(propertyPath, strconv.Itoa(index))
			} else if ast.IsObjectLiteralExpression(parent) && pathKnown && len(propertyPath) > 0 {
				// A later required property replaces this value, so its assertion
				// cannot narrow the target. Match the checker's spread-property rules.
				properties := parent.AsObjectLiteralExpression().Properties.Nodes
				for _, property := range properties[slices.Index(properties, current)+1:] {
					if ast.IsSpreadAssignment(property) {
						spreadType := ctx.TypeChecker.GetTypeAtLocation(property.Expression())
						overwriting := checker.Checker_getPropertyOfType(ctx.TypeChecker, spreadType, propertyPath[len(propertyPath)-1])
						if overwriting != nil && overwriting.Flags&ast.SymbolFlagsOptional == 0 && overwriting.CheckFlags&ast.CheckFlagsPartial == 0 &&
							checker.GetDeclarationModifierFlagsFromSymbol(overwriting)&(ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected) == 0 &&
							checker.Checker_isSpreadableProperty(ctx.TypeChecker, overwriting) {
							return false
						}
						continue
					}
					if property.Name() == nil {
						continue
					}
					name, known := ast.TryGetTextOfPropertyName(property.Name())
					if known && name == propertyPath[len(propertyPath)-1] {
						return false
					}
				}
			}
			continue
		}
		if !ast.IsAssignmentExpression(parent, true) || parent.AsBinaryExpression().Right != current {
			return false
		}

		left := ast.SkipParentheses(parent.AsBinaryExpression().Left)
		var receiverType *checker.Type
		if inLiteral {
			if !ast.IsAssignmentPattern(left) {
				return false
			}
			// Destructuring context supplies the type of the corresponding target.
			receiverType = checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
			if pathKnown {
				if initializer := getDestructuringDefault(left, propertyPath); initializer != nil {
					uncastType = checker.Checker_getTypeWithDefault(ctx.TypeChecker, uncastType, initializer)
					castType = checker.Checker_getTypeWithDefault(ctx.TypeChecker, castType, initializer)
				}
			}
		} else {
			receiverType = ctx.TypeChecker.GetTypeAtLocation(left)
		}
		if receiverType == nil {
			return false
		}
		if constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, receiverType); constraint != nil {
			receiverType = constraint
		}
		if !utils.IsUnionType(receiverType) {
			return false
		}

		// Use the checker's assignment reduction so assertions that leave the
		// same union members reachable can still be reported as unnecessary.
		uncastReduced := checker.Checker_getAssignmentReducedType(ctx.TypeChecker, receiverType, uncastType)
		castReduced := checker.Checker_getAssignmentReducedType(ctx.TypeChecker, receiverType, castType)
		return !checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, uncastReduced, castReduced)
	}
	return false
}
