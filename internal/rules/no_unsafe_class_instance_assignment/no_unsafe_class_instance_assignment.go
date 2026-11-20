package no_unsafe_class_instance_assignment

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildNonInstanceMessage(typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonInstance",
		Description: fmt.Sprintf("This value is used where an instance of class `%s` is expected, but was not created by its constructor.", typeName),
	}
}

func buildObjectLiteralMessage(typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "objectLiteral",
		Description: fmt.Sprintf("This object literal is used where an instance of class `%s` is expected.", typeName),
	}
}

type classInstanceChecker struct {
	typeChecker *checker.Checker
}

func isTypeReference(t *checker.Type) bool {
	return utils.IsObjectType(t) && checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0
}

func getClassTarget(t *checker.Type) *checker.Type {
	if !utils.IsObjectType(t) {
		return nil
	}

	target := t
	if isTypeReference(t) {
		target = t.Target()
	}

	if checker.Type_objectFlags(target)&checker.ObjectFlagsClass != 0 {
		return target
	}
	return nil
}

func (c *classInstanceChecker) getNominalClass(t *checker.Type) *checker.Type {
	var found *checker.Type
	utils.MatchesTypeOrBaseType(c.typeChecker, t, func(candidate *checker.Type) bool {
		if getClassTarget(candidate) != nil {
			found = candidate
		}
		return found != nil
	})
	return found
}

func (c *classInstanceChecker) resolveConstraint(t *checker.Type) *checker.Type {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsInstantiable) {
		return c.typeChecker.GetBaseConstraintOfType(t)
	}
	return t
}

func (c *classInstanceChecker) containsClass(t *checker.Type, seen map[*checker.Type]struct{}) bool {
	if _, ok := seen[t]; ok {
		return false
	}
	seen[t] = struct{}{}

	resolved := c.resolveConstraint(t)
	if resolved == nil {
		return false
	}

	if utils.IsTypeFlagSet(resolved, checker.TypeFlagsUnionOrIntersection) {
		for _, part := range resolved.Types() {
			if c.containsClass(part, seen) {
				return true
			}
		}
		return false
	}

	if c.getNominalClass(resolved) != nil {
		return true
	}

	if isTypeReference(resolved) {
		for _, typeArgument := range c.typeChecker.GetTypeArguments(resolved) {
			if c.containsClass(typeArgument, seen) {
				return true
			}
		}
	}

	for _, indexInfo := range c.typeChecker.GetIndexInfosOfType(resolved) {
		if c.containsClass(indexInfo.ValueType(), seen) {
			return true
		}
	}
	return false
}

func (c *classInstanceChecker) hasClass(t *checker.Type) bool {
	return c.containsClass(t, map[*checker.Type]struct{}{})
}

func (c *classInstanceChecker) isInstanceOf(source *checker.Type, classType *checker.Type) bool {
	classTarget := getClassTarget(classType)

	for _, part := range utils.IntersectionTypeParts(source) {
		if utils.IsObjectType(part) && checker.Type_objectFlags(part)&checker.ObjectFlagsMapped != 0 {
			return true
		}
		if utils.MatchesTypeOrBaseType(c.typeChecker, part, func(t *checker.Type) bool {
			return getClassTarget(t) == classTarget
		}) {
			return true
		}
	}
	return false
}

func (c *classInstanceChecker) findViolation(source *checker.Type, target *checker.Type, shallow bool) *checker.Type {
	resolvedSource := c.resolveConstraint(source)
	resolvedTarget := c.resolveConstraint(target)

	if resolvedSource == nil ||
		resolvedTarget == nil ||
		resolvedSource == resolvedTarget ||
		utils.IsTypeFlagSet(resolvedSource, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNever) ||
		!c.hasClass(resolvedTarget) {
		return nil
	}

	if resolvedSource.IsUnion() {
		for _, part := range resolvedSource.Types() {
			if !c.typeChecker.IsTypeAssignableTo(part, resolvedTarget) {
				continue
			}
			if violation := c.findViolation(part, resolvedTarget, shallow); violation != nil {
				return violation
			}
		}
		return nil
	}

	if !utils.IsObjectType(resolvedSource) &&
		!(resolvedSource.IsIntersection() && !c.typeChecker.IsTypeAssignableTo(resolvedSource, c.typeChecker.GetNeverType())) {
		return nil
	}

	if resolvedTarget.IsUnion() {
		candidates := make([]*checker.Type, 0, len(resolvedTarget.Types()))
		for _, part := range resolvedTarget.Types() {
			if c.typeChecker.IsTypeAssignableTo(resolvedSource, part) {
				candidates = append(candidates, part)
			}
		}
		if len(candidates) == 0 {
			return nil
		}
		for _, candidate := range candidates {
			if !c.hasClass(candidate) {
				return nil
			}
		}

		var firstViolation *checker.Type
		for _, candidate := range candidates {
			violation := c.findViolation(resolvedSource, candidate, shallow)
			if violation == nil {
				return nil
			}
			if firstViolation == nil {
				firstViolation = violation
			}
		}
		return firstViolation
	}

	if resolvedTarget.IsIntersection() {
		for _, part := range resolvedTarget.Types() {
			if violation := c.findViolation(resolvedSource, part, shallow); violation != nil {
				return violation
			}
		}
		return nil
	}

	nominalClass := c.getNominalClass(resolvedTarget)
	if nominalClass != nil && !c.isInstanceOf(resolvedSource, nominalClass) {
		return nominalClass
	}

	if shallow {
		return nil
	}

	if isTypeReference(resolvedSource) && isTypeReference(resolvedTarget) && resolvedSource.Target() == resolvedTarget.Target() {
		sourceArguments := c.typeChecker.GetTypeArguments(resolvedSource)
		targetArguments := c.typeChecker.GetTypeArguments(resolvedTarget)
		for i, targetArgument := range targetArguments[:min(len(sourceArguments), len(targetArguments))] {
			if violation := c.findViolation(sourceArguments[i], targetArgument, false); violation != nil {
				return violation
			}
		}
		return nil
	}

	for _, keyType := range []*checker.Type{c.typeChecker.GetNumberType(), c.typeChecker.GetStringType()} {
		sourceIndexInfo := c.typeChecker.GetIndexInfoOfType(resolvedSource, keyType)
		targetIndexInfo := c.typeChecker.GetIndexInfoOfType(resolvedTarget, keyType)
		if sourceIndexInfo == nil || targetIndexInfo == nil {
			continue
		}
		if violation := c.findViolation(sourceIndexInfo.ValueType(), targetIndexInfo.ValueType(), false); violation != nil {
			return violation
		}
	}
	return nil
}

func isSkippedSite(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindConditionalExpression, ast.KindFunctionExpression:
		return true
	case ast.KindBinaryExpression:
		switch node.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken, ast.KindCommaToken:
			return true
		}
	}
	return false
}

var NoUnsafeClassInstanceAssignmentRule = rule.Rule{
	Name: "no-unsafe-class-instance-assignment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		c := &classInstanceChecker{typeChecker: ctx.TypeChecker}

		checkSite := func(node *ast.Node) {
			if node == nil {
				return
			}
			node = ast.SkipParentheses(node)
			if isSkippedSite(node) {
				return
			}

			contextualType := ctx.TypeChecker.GetContextualType(node, checker.ContextFlagsNone)
			if contextualType == nil {
				return
			}

			isLiteral := ast.IsObjectLiteralExpression(node) || ast.IsArrayLiteralExpression(node)
			violation := c.findViolation(ctx.TypeChecker.GetTypeAtLocation(node), contextualType, isLiteral)
			if violation == nil {
				return
			}

			typeName := ctx.TypeChecker.TypeToString(violation)
			if ast.IsObjectLiteralExpression(node) {
				ctx.ReportNode(node, buildObjectLiteralMessage(typeName))
			} else {
				ctx.ReportNode(node, buildNonInstanceMessage(typeName))
			}
		}

		checkArguments := func(arguments *ast.NodeList) {
			if arguments == nil {
				return
			}
			for _, argument := range arguments.Nodes {
				if !ast.IsSpreadElement(argument) {
					checkSite(argument)
				}
			}
		}

		checkInitializer := func(node *ast.Node) {
			checkSite(node.Initializer())
		}

		checkExpression := func(node *ast.Node) {
			checkSite(node.Expression())
		}

		return rule.RuleListeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				if ast.IsAssignmentTarget(node) {
					return
				}
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					if !ast.IsOmittedExpression(element) && !ast.IsSpreadElement(element) {
						checkSite(element)
					}
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if !ast.IsBlock(body) {
					checkSite(body)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()
				switch expression.OperatorToken.Kind {
				case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken:
					checkSite(expression.Right)
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
					checkSite(expression.Left)
					checkSite(expression.Right)
				case ast.KindCommaToken:
					checkSite(expression.Right)
				}
			},
			ast.KindBindingElement: checkInitializer,
			ast.KindCallExpression: func(node *ast.Node) {
				if ast.IsImportCall(node) {
					return
				}
				checkArguments(node.AsCallExpression().Arguments)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				expression := node.AsConditionalExpression()
				checkSite(expression.WhenTrue)
				checkSite(expression.WhenFalse)
			},
			ast.KindJsxExpression: func(node *ast.Node) {
				expression := node.AsJsxExpression()
				if expression.DotDotDotToken == nil {
					checkSite(expression.Expression)
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				checkArguments(node.AsNewExpression().Arguments)
			},
			ast.KindParameter: checkInitializer,
			ast.KindPropertyAssignment: func(node *ast.Node) {
				if !ast.IsAssignmentTarget(node.Parent) {
					checkInitializer(node)
				}
			},
			ast.KindPropertyDeclaration: checkInitializer,
			ast.KindReturnStatement:     checkExpression,
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				property := node.AsShorthandPropertyAssignment()
				if property.ObjectAssignmentInitializer != nil {
					checkSite(property.ObjectAssignmentInitializer)
				} else if !ast.IsAssignmentTarget(node.Parent) {
					checkSite(property.Name())
				}
			},
			ast.KindAsExpression:            checkExpression,
			ast.KindSatisfiesExpression:     checkExpression,
			ast.KindTypeAssertionExpression: checkExpression,
			ast.KindVariableDeclaration:     checkInitializer,
			ast.KindYieldExpression:         checkExpression,
		}
	},
}
