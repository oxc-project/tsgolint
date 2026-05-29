package no_unnecessary_type_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// This file ports the `contextuallyUnnecessary` ("the receiver accepts the
// original type of the expression") path of typescript-eslint's
// `no-unnecessary-type-assertion` rule for `as` / `<T>` assertions.
//
// Source of truth: typescript-eslint v8.59.x (introduced in PR #11789, hardened
// by #12246/#12257/#12269/#12278). Helper names mirror the upstream functions so
// the two implementations can be cross-checked.

// SKIP_PARENT_TYPES upstream.
func isSkipParentKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression,
		ast.KindSpreadElement, ast.KindSatisfiesExpression:
		return true
	}
	return false
}

// skipUpwardParentheses returns the outermost ParenthesizedExpression wrapping
// node (or node itself). typescript-eslint operates on an ESTree where parens
// are not nodes, so a guard that inspects `node.parent` there must look past TS
// AST ParenthesizedExpression wrappers here.
func skipUpwardParentheses(node *ast.Node) *ast.Node {
	n := node
	for n.Parent != nil && ast.IsParenthesizedExpression(n.Parent) {
		n = n.Parent
	}
	return n
}

// getTypeArguments upstream: prefer alias type arguments, otherwise the type
// reference's arguments (guarded by the reference object flag, per #12246).
func getTypeArgumentsForAssertion(tc *checker.Checker, t *checker.Type) []*checker.Type {
	if alias := checker.Type_alias(t); alias != nil {
		if args := alias.TypeArguments(); len(args) > 0 {
			return args
		}
	}
	if checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		return checker.Checker_getTypeArguments(tc, t)
	}
	return nil
}

// typeContains upstream: recursively walk unions/intersections, type arguments,
// and call-signature return/parameter types looking for a matching type.
func typeContains(tc *checker.Checker, t *checker.Type, predicate func(*checker.Type) bool, seen *utils.Set[*checker.Type]) bool {
	if seen.Has(t) {
		return false
	}
	seen.Add(t)
	if predicate(t) {
		return true
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, part := range t.Types() {
			if typeContains(tc, part, predicate, seen) {
				return true
			}
		}
		return false
	}
	for _, arg := range getTypeArgumentsForAssertion(tc, t) {
		if typeContains(tc, arg, predicate, seen) {
			return true
		}
	}
	for _, sig := range tc.GetCallSignatures(t) {
		if ret := tc.GetReturnTypeOfSignature(sig); ret != nil && typeContains(tc, ret, predicate, seen) {
			return true
		}
		for _, p := range checker.Signature_parameters(sig) {
			if pt := checker.Checker_getTypeOfSymbol(tc, p); pt != nil && typeContains(tc, pt, predicate, seen) {
				return true
			}
		}
	}
	return false
}

// containsAny upstream.
func containsAny(tc *checker.Checker, t *checker.Type) bool {
	seen := &utils.Set[*checker.Type]{}
	return typeContains(tc, t, utils.IsTypeAnyType, seen)
}

// hasTypeParams upstream.
func hasTypeParams(sig *checker.Signature) bool {
	return len(checker.Signature_typeParameters(sig)) > 0
}

// hasGenericCallSignature upstream.
func hasGenericCallSignature(tc *checker.Checker, t *checker.Type) bool {
	for _, sig := range tc.GetCallSignatures(t) {
		if hasTypeParams(sig) {
			return true
		}
	}
	return false
}

// genericsMismatch upstream: a generic call-signature property on the contextual
// type that the uncast type does not also provide generically.
func genericsMismatch(tc *checker.Checker, uncast, contextual *checker.Type) bool {
	for _, prop := range checker.Checker_getPropertiesOfType(tc, contextual) {
		propType := checker.Checker_getTypeOfSymbol(tc, prop)
		contextualHasGeneric := false
		for _, s := range checker.Checker_getSignaturesOfType(tc, propType, checker.SignatureKindCall) {
			if hasTypeParams(s) {
				contextualHasGeneric = true
				break
			}
		}
		if !contextualHasGeneric {
			continue
		}
		uncastProp := checker.Checker_getPropertyOfType(tc, uncast, prop.Name)
		if uncastProp == nil {
			return true
		}
		uncastPropType := checker.Checker_getTypeOfSymbol(tc, uncastProp)
		uncastHasGeneric := false
		for _, s := range checker.Checker_getSignaturesOfType(tc, uncastPropType, checker.SignatureKindCall) {
			if hasTypeParams(s) {
				uncastHasGeneric = true
				break
			}
		}
		if !uncastHasGeneric {
			return true
		}
	}
	return false
}

// isEmptyObjectType upstream.
func isEmptyObjectType(tc *checker.Checker, t *checker.Type) bool {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsNonPrimitive) {
		return true
	}
	return len(checker.Checker_getPropertiesOfType(tc, t)) == 0 &&
		len(tc.GetCallSignatures(t)) == 0 &&
		len(tc.GetConstructSignatures(t)) == 0 &&
		tc.GetStringIndexType(t) == nil &&
		tc.GetNumberIndexType(t) == nil
}

// haveSameTypeArguments upstream.
func haveSameTypeArguments(tc *checker.Checker, uncast, cast *checker.Type) bool {
	uncastArgs := getTypeArgumentsForAssertion(tc, uncast)
	castArgs := getTypeArgumentsForAssertion(tc, cast)
	if len(uncastArgs) != len(castArgs) {
		return false
	}
	for i := range uncastArgs {
		if uncastArgs[i] != castArgs[i] {
			return false
		}
	}
	return true
}

// hasPhantomTypeArguments upstream (#12269).
func hasPhantomTypeArguments(tc *checker.Checker, t *checker.Type) bool {
	return isEmptyObjectType(tc, t) && len(getTypeArgumentsForAssertion(tc, t)) > 0
}

// hasPhantomTypeArgumentMismatch upstream (#12269).
func hasPhantomTypeArgumentMismatch(tc *checker.Checker, node *ast.Node, uncast, contextual *checker.Type) bool {
	return isInGenericContext(tc, node) &&
		(hasPhantomTypeArguments(tc, uncast) || hasPhantomTypeArguments(tc, contextual)) &&
		!haveSameTypeArguments(tc, uncast, contextual)
}

// getOriginalExpression upstream: unwrap chained assertions. Parentheses are not
// ESTree nodes upstream, so skip TS AST ParenthesizedExpression wrappers to match.
func getOriginalExpression(node *ast.Node) *ast.Node {
	current := ast.SkipParentheses(node.Expression())
	for current.Kind == ast.KindAsExpression || current.Kind == ast.KindTypeAssertionExpression {
		current = ast.SkipParentheses(current.Expression())
	}
	return current
}

// isConceptuallyLiteral upstream: CONCEPTUALLY_LITERAL_TYPES.
func isConceptuallyLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression,
		ast.KindTemplateExpression,
		ast.KindClassExpression, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	}
	return false
}

// nodeInArguments reports whether node is one of call's arguments.
func nodeInArguments(call *ast.Node, node *ast.Node) bool {
	for _, arg := range call.Arguments() {
		if arg == node {
			return true
		}
	}
	return false
}

// isInGenericContext upstream.
func isInGenericContext(tc *checker.Checker, node *ast.Node) bool {
	seenFunction := false
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration:
			return false
		case ast.KindFunctionExpression, ast.KindArrowFunction:
			if body := current.Body(); body != nil && body.Kind == ast.KindBlock {
				return false
			}
			if seenFunction {
				return false
			}
			seenFunction = true
		case ast.KindCallExpression, ast.KindNewExpression:
			if current.TypeArguments() != nil {
				continue
			}
			callee := current.Expression()
			if current.Kind == ast.KindCallExpression &&
				(ast.IsPropertyAccessExpression(callee) || ast.IsElementAccessExpression(callee)) &&
				nodeInArguments(current, skipUpwardParentheses(node)) {
				continue
			}
			if hasGenericCallSignature(tc, tc.GetTypeAtLocation(callee)) {
				return true
			}
		}
	}
	return false
}

// isInDestructuringDeclaration upstream.
func isInDestructuringDeclaration(node *ast.Node) bool {
	outer := skipUpwardParentheses(node)
	parent := outer.Parent
	if parent.Kind != ast.KindVariableDeclaration || parent.Initializer() != outer {
		return false
	}
	name := parent.Name()
	return name != nil && (name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern)
}

// isPropertyInProblematicContext upstream.
func isPropertyInProblematicContext(tc *checker.Checker, node *ast.Node) bool {
	outer := skipUpwardParentheses(node)
	parent := outer.Parent
	if parent.Kind != ast.KindPropertyAssignment || parent.Initializer() != outer {
		return false
	}
	objectExpr := parent.Parent
	if objectExpr.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	if objContextual := checker.Checker_getContextualType(tc, objectExpr, checker.ContextFlagsNone); objContextual != nil && utils.IsUnionType(objContextual) {
		propContextual := checker.Checker_getContextualType(tc, node, checker.ContextFlagsNone)
		if propContextual == nil {
			return true
		}
		nonNullable := checker.Checker_GetNonNullableType(tc, propContextual)
		if utils.IsUnionType(nonNullable) {
			return true
		}
		uncast := tc.GetTypeAtLocation(node.Expression())
		return !checker.Checker_isTypeAssignableTo(tc, uncast, nonNullable)
	}
	objectParent := objectExpr.Parent
	if objectParent.Kind == ast.KindSatisfiesExpression {
		return true
	}
	return objectParent.Kind == ast.KindCallExpression && objectParent.Parent.Kind == ast.KindSatisfiesExpression
}

// isAssignmentInNonStatementContext upstream.
func isAssignmentInNonStatementContext(node *ast.Node) bool {
	outer := skipUpwardParentheses(node)
	parent := outer.Parent
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	if bin.OperatorToken.Kind != ast.KindEqualsToken || bin.Right != outer {
		return false
	}
	return parent.Parent.Kind != ast.KindExpressionStatement
}

// isRightHandSideOfLogicalAssignment upstream (#12278).
func isRightHandSideOfLogicalAssignment(node *ast.Node) bool {
	outer := skipUpwardParentheses(node)
	parent := outer.Parent
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	if bin.Right != outer {
		return false
	}
	switch bin.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// isArgumentToOverloadedFunction upstream.
func isArgumentToOverloadedFunction(tc *checker.Checker, node *ast.Node) bool {
	outer := skipUpwardParentheses(node)
	parent := outer.Parent
	if parent.Kind != ast.KindCallExpression && parent.Kind != ast.KindNewExpression {
		return false
	}
	argIndex := -1
	for i, a := range parent.Arguments() {
		if a == outer {
			argIndex = i
			break
		}
	}
	if argIndex < 0 {
		return false
	}
	signatures := tc.GetCallSignatures(tc.GetTypeAtLocation(parent.Expression()))
	if len(signatures) <= 1 {
		return false
	}

	paramTypes := make([]*checker.Type, len(signatures))
	for i, sig := range signatures {
		params := checker.Signature_parameters(sig)
		if argIndex >= len(params) {
			// missing parameter for this overload -> treat as ambiguous below
			return true
		}
		param := params[argIndex]
		paramType := checker.Checker_getTypeOfSymbol(tc, param)
		if decl := param.ValueDeclaration; decl != nil && ast.IsParameterDeclaration(decl) && decl.AsParameterDeclaration().DotDotDotToken != nil {
			if typeArgs := getTypeArgumentsForAssertion(tc, paramType); len(typeArgs) > 0 {
				paramType = typeArgs[0]
			}
		}
		paramTypes[i] = paramType
	}

	first := paramTypes[0]
	allSame := true
	for _, pt := range paramTypes {
		if pt != first {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	uncast := tc.GetTypeAtLocation(node.Expression())
	for _, pt := range paramTypes {
		if !checker.Checker_isTypeAssignableTo(tc, uncast, pt) {
			return true
		}
	}
	return false
}

// shouldSkipContextualTypeFallback upstream.
func shouldSkipContextualTypeFallback(tc *checker.Checker, node *ast.Node, castIsAny bool) bool {
	parent := skipUpwardParentheses(node).Parent
	if castIsAny {
		return ast.IsLogicalExpression(parent) || isInGenericContext(tc, node)
	}

	if isSkipParentKind(parent.Kind) ||
		ast.IsArrayLiteralExpression(ast.SkipParentheses(node.Expression())) ||
		isInDestructuringDeclaration(node) ||
		isPropertyInProblematicContext(tc, node) ||
		isAssignmentInNonStatementContext(node) ||
		isRightHandSideOfLogicalAssignment(node) ||
		isArgumentToOverloadedFunction(tc, node) {
		return true
	}

	if isInGenericContext(tc, node) {
		return !isConceptuallyLiteral(getOriginalExpression(node)) &&
			parent.Kind != ast.KindPropertyAssignment
	}

	return false
}
