package no_unnecessary_type_assertion

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildContextuallyUnnecessaryMessage(assertion core.TextRange) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: assertion,
		Message: rule.RuleMessage{
			Id:          "contextuallyUnnecessary",
			Description: "This assertion is unnecessary since the receiver accepts the original type of the expression.",
		},
	}
}
func buildUnnecessaryAssertionDiagnostic(assertion core.TextRange) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: assertion,
		Message: rule.RuleMessage{
			Id:          "unnecessaryAssertion",
			Description: "This assertion is unnecessary since it does not change the type of the expression.",
		},
	}
}

// typescript-go represents parentheses as AST nodes. Expression and parent
// lookups skip these nodes to match ESTree semantics.
var NoUnnecessaryTypeAssertionRule = rule.Rule{
	Name: "no-unnecessary-type-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoUnnecessaryTypeAssertionOptions](options, "no-unnecessary-type-assertion")

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		parentThroughParens := func(node *ast.Node) *ast.Node {
			parent := node.Parent
			for parent != nil && ast.IsParenthesizedExpression(parent) {
				parent = parent.Parent
			}
			return parent
		}

		/**
		 * Returns true if there's a chance the variable has been used before a value has been assigned to it
		 */
		isPossiblyUsedBeforeAssigned := func(node *ast.Node) bool {
			declaration := utils.GetDeclaration(ctx.TypeChecker, node)
			if declaration == nil {
				// don't know what the declaration is for some reason, so just assume the worst
				return true
			}
			// non-strict mode doesn't care about used before assigned errors
			if !isStrictNullChecks {
				return false
			}
			// ignore class properties as they are compile time guarded
			// also ignore function arguments as they can't be used before defined
			if !ast.IsVariableDeclaration(declaration) {
				return false
			}

			decl := declaration.AsVariableDeclaration()

			// For var declarations, we need to check whether the node
			// is actually in a descendant of its declaration or not. If not,
			// it may be used before defined.

			// eg
			// if (Math.random() < 0.5) {
			//     var x: number  = 2;
			// } else {
			//     x!.toFixed();
			// }
			if ast.IsVariableDeclarationList(declaration.Parent) &&
				// var
				declaration.Parent.Flags == ast.NodeFlagsNone {
				// If they are not in the same file it will not exist.
				// This situation must not occur using before defined.
				declaratorScope := ast.GetEnclosingBlockScopeContainer(declaration)
				scope := ast.GetEnclosingBlockScopeContainer(node)

				parentScope := declaratorScope
				for {
					parentScope = ast.GetEnclosingBlockScopeContainer(parentScope)
					if parentScope == nil {
						break
					}
					if parentScope == scope {
						return true
					}
				}
			}

			if
			// is it `const x: number`
			decl.Initializer == nil &&
				decl.ExclamationToken == nil &&
				decl.Type != nil {
				// check if the defined variable type has changed since assignment
				declarationType := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declaration.Type())
				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
				if declarationType == t &&
					// `declare`s are never narrowed, so never skip them
					!(ast.IsVariableDeclarationList(declaration.Parent) &&
						ast.IsVariableStatement(declaration.Parent.Parent) &&
						utils.IncludesModifier(declaration.Parent.Parent.AsVariableStatement(), ast.KindDeclareKeyword)) {
					// possibly used before assigned, so just skip it
					// better to false negative and skip it, than false positive and fix to compile erroring code
					//
					// no better way to figure this out right now
					// https://github.com/Microsoft/TypeScript/issues/31124
					return true
				}
			}

			return false
		}
		isConstAssertion := func(node *ast.Node) bool {
			if !ast.IsTypeReferenceNode(node) {
				return false
			}
			typeName := node.AsTypeReferenceNode().TypeName
			return ast.IsIdentifier(typeName) && typeName.Text() == "const"
		}

		isImplicitlyNarrowedLiteralDeclaration := func(node *ast.Node) bool {
			expression := ast.SkipParentheses(node.Expression())
			parent := parentThroughParens(node)
			/**
			 * Even on `const` variable declarations, template literals with expressions can sometimes be widened without a type assertion.
			 * @see https://github.com/typescript-eslint/typescript-eslint/issues/8737
			 */
			if ast.IsTemplateExpression(expression) {
				return false
			}

			return (ast.IsVariableDeclaration(parent) && ast.IsVariableDeclarationList(parent.Parent) && parent.Parent.Flags&ast.NodeFlagsConst != 0) ||
				(ast.IsPropertyDeclaration(parent) && parent.ModifierFlags()&ast.ModifierFlagsReadonly != 0)

		}

		getTypeArguments := func(t *checker.Type) []*checker.Type {
			if alias := checker.Type_alias(t); alias != nil && len(alias.TypeArguments()) > 0 {
				return alias.TypeArguments()
			}
			if checker.Type_objectFlags(t)&checker.ObjectFlagsReference == 0 {
				return nil
			}
			return checker.Checker_getTypeArguments(ctx.TypeChecker, t)
		}

		var typeContains func(t *checker.Type, predicate func(*checker.Type) bool, seenTypes map[*checker.Type]struct{}, activeSignatures map[*checker.Signature]struct{}) bool
		typeContains = func(t *checker.Type, predicate func(*checker.Type) bool, seenTypes map[*checker.Type]struct{}, activeSignatures map[*checker.Signature]struct{}) bool {
			if t == nil {
				return false
			}
			if _, ok := seenTypes[t]; ok {
				return false
			}
			seenTypes[t] = struct{}{}
			if predicate(t) {
				return true
			}
			if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
				return slices.ContainsFunc(t.Types(), func(part *checker.Type) bool {
					return typeContains(part, predicate, seenTypes, activeSignatures)
				})
			}
			for _, typeArgument := range getTypeArguments(t) {
				if typeContains(typeArgument, predicate, seenTypes, activeSignatures) {
					return true
				}
			}
			for _, sig := range utils.GetCallSignatures(ctx.TypeChecker, t) {
				// Generic signature instantiations can produce fresh recursive return
				// types. If their shared original target is already active,
				// conservatively assume the predicate could occur in the unseen cycle
				// instead of making an unsafe autofix.
				signatureIdentity := sig
				for signatureIdentity.Target() != nil {
					signatureIdentity = signatureIdentity.Target()
				}
				if _, ok := activeSignatures[signatureIdentity]; ok {
					return true
				}
				activeSignatures[signatureIdentity] = struct{}{}

				for _, param := range checker.Signature_parameters(sig) {
					if typeContains(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, param), predicate, seenTypes, activeSignatures) {
						return true
					}
				}
				if typeContains(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig), predicate, seenTypes, activeSignatures) {
					return true
				}
				delete(activeSignatures, signatureIdentity)
			}
			return false
		}

		containsAny := func(t *checker.Type) bool {
			return typeContains(t, func(part *checker.Type) bool {
				return utils.IsTypeFlagSet(part, checker.TypeFlagsAny)
			}, map[*checker.Type]struct{}{}, map[*checker.Signature]struct{}{})
		}

		containsTypeVariable := func(t *checker.Type) bool {
			return typeContains(t, func(part *checker.Type) bool {
				return utils.IsTypeFlagSet(part, checker.TypeFlagsTypeVariable|checker.TypeFlagsIndex)
			}, map[*checker.Type]struct{}{}, map[*checker.Signature]struct{}{})
		}

		hasIndexSignature := func(t *checker.Type) bool {
			return slices.ContainsFunc(utils.UnionTypeParts(t), func(part *checker.Type) bool {
				return len(checker.Checker_getIndexInfosOfType(ctx.TypeChecker, part)) > 0
			})
		}

		hasSameProperties := func(uncast, cast *checker.Type) bool {
			uncastProps := checker.Checker_getPropertiesOfType(ctx.TypeChecker, uncast)
			castProps := checker.Checker_getPropertiesOfType(ctx.TypeChecker, cast)
			if len(uncastProps) != len(castProps) {
				return false
			}

			castPropsByName := make(map[string]*ast.Symbol, len(castProps))
			for _, prop := range castProps {
				castPropsByName[prop.Name] = prop
			}

			for _, prop := range uncastProps {
				castProp := castPropsByName[prop.Name]
				if castProp == nil ||
					checker.Checker_isReadonlySymbol(ctx.TypeChecker, prop) != checker.Checker_isReadonlySymbol(ctx.TypeChecker, castProp) {
					return false
				}
			}
			return true
		}

		haveSameTypeArguments := func(uncast, cast *checker.Type) bool {
			uncastArgs := getTypeArguments(uncast)
			castArgs := getTypeArguments(cast)
			if len(uncastArgs) != len(castArgs) {
				return false
			}
			for i, arg := range uncastArgs {
				if arg != castArgs[i] {
					return false
				}
			}
			return true
		}

		areMutuallyAssignable := func(a, b *checker.Type) bool {
			return checker.Checker_isTypeAssignableTo(ctx.TypeChecker, a, b) &&
				checker.Checker_isTypeAssignableTo(ctx.TypeChecker, b, a)
		}

		areUnionPartsEquivalentIgnoringUndefined := func(uncast, cast *checker.Type) bool {
			uncastParts := utils.Set[*checker.Type]{}
			for _, part := range utils.UnionTypeParts(uncast) {
				if !utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					uncastParts.Add(part)
				}
			}

			castPartsCount := 0
			for _, part := range utils.UnionTypeParts(cast) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					continue
				}
				if !uncastParts.Has(part) {
					return false
				}
				castPartsCount++
			}
			return uncastParts.Len() == castPartsCount
		}

		isEmptyObjectType := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsNonPrimitive) ||
				(len(checker.Checker_getPropertiesOfType(ctx.TypeChecker, t)) == 0 &&
					len(utils.GetCallSignatures(ctx.TypeChecker, t)) == 0 &&
					len(utils.GetConstructSignatures(ctx.TypeChecker, t)) == 0 &&
					len(checker.Checker_getIndexInfosOfType(ctx.TypeChecker, t)) == 0)
		}

		hasPhantomTypeArguments := func(t *checker.Type) bool {
			return isEmptyObjectType(t) && len(getTypeArguments(t)) > 0
		}

		isConceptuallyLiteral := func(node *ast.Node) bool {
			node = ast.SkipParentheses(node)
			return ast.IsArrayLiteralExpression(node) ||
				ast.IsObjectLiteralExpression(node) ||
				ast.IsClassExpression(node) ||
				ast.IsFunctionExpression(node) ||
				ast.IsArrowFunction(node) ||
				ast.IsJsxElement(node) ||
				ast.IsJsxSelfClosingElement(node) ||
				ast.IsJsxFragment(node) ||
				ast.IsStringLiteral(node) ||
				node.Kind == ast.KindNumericLiteral ||
				node.Kind == ast.KindBigIntLiteral ||
				node.Kind == ast.KindRegularExpressionLiteral ||
				node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
				node.Kind == ast.KindTrueKeyword ||
				node.Kind == ast.KindFalseKeyword ||
				node.Kind == ast.KindNullKeyword ||
				ast.IsTemplateExpression(node)
		}

		isTypeLiteral := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral|checker.TypeFlagsBooleanLiteral)
		}

		isReceiverOfWriteAccess := func(node *ast.Node) bool {
			current := node.Parent
			for current != nil && ast.IsParenthesizedExpression(current) {
				current = current.Parent
			}
			for current != nil && ast.IsAccessExpression(current) {
				if ast.IsWriteAccess(current) {
					return true
				}
				current = current.Parent
				for current != nil && ast.IsParenthesizedExpression(current) {
					current = current.Parent
				}
			}
			return false
		}

		isElementAccessArgument := func(node *ast.Node) bool {
			current := node
			parent := current.Parent
			for parent != nil && ast.IsParenthesizedExpression(parent) {
				current = parent
				parent = parent.Parent
			}
			return parent != nil &&
				ast.IsElementAccessExpression(parent) &&
				parent.AsElementAccessExpression().ArgumentExpression == current
		}

		var hasEnumType func(t *checker.Type) bool
		hasEnumType = func(t *checker.Type) bool {
			if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
				return true
			}
			return (utils.IsUnionType(t) || utils.IsIntersectionType(t)) && slices.ContainsFunc(t.Types(), hasEnumType)
		}

		isTypeUnchanged := func(node *ast.Node, expression *ast.Node, uncast, cast *checker.Type) bool {
			expression = ast.SkipParentheses(expression)
			if uncast == cast {
				return true
			}
			// Numeric enums and number are mutually assignable, but assertions
			// between them still change the type, including within unions and intersections.
			if (hasEnumType(uncast) || hasEnumType(cast)) &&
				!checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, uncast, cast) {
				return false
			}
			if utils.IsTypeParameter(uncast) && isReceiverOfWriteAccess(node) {
				// Local safeguard: widening a generic write receiver can make the write legal.
				return false
			}
			if compilerOptions.NoUncheckedIndexedAccess.IsTrue() && isElementAccessArgument(node) {
				// Local safeguard: the asserted key type can change indexed-access nullability.
				return false
			}

			typeNode := ast.SkipTypeParentheses(node.Type())
			if ast.IsIntersectionTypeNode(typeNode) && containsTypeVariable(cast) {
				return false
			}

			if compilerOptions.ExactOptionalPropertyTypes.IsTrue() &&
				utils.IsTypeFlagSet(uncast, checker.TypeFlagsUndefined) &&
				utils.IsTypeFlagSet(cast, checker.TypeFlagsUndefined) {
				return areUnionPartsEquivalentIgnoringUndefined(uncast, cast)
			}

			if (utils.IsTypeFlagSet(uncast, checker.TypeFlagsNonPrimitive) && !utils.IsTypeFlagSet(cast, checker.TypeFlagsNonPrimitive)) ||
				(hasIndexSignature(uncast) != hasIndexSignature(cast)) {
				return false
			}

			if isConceptuallyLiteral(expression) &&
				(!ast.IsObjectLiteralExpression(expression) ||
					len(expression.AsObjectLiteralExpression().Properties.Nodes) == 0 ||
					slices.ContainsFunc(checker.Checker_getPropertiesOfType(ctx.TypeChecker, cast), func(prop *ast.Symbol) bool {
						return isTypeLiteral(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, prop))
					})) {
				return false
			}

			if utils.IsIntersectionType(cast) && !utils.IsIntersectionType(uncast) {
				castParts := cast.Types()
				var otherPart *checker.Type
				for _, part := range castParts {
					if part != uncast {
						otherPart = part
						break
					}
				}
				if utils.IsTypeParameter(uncast) &&
					len(castParts) == 2 &&
					slices.Contains(castParts, uncast) &&
					otherPart != nil &&
					isEmptyObjectType(otherPart) &&
					!containsTypeVariable(otherPart) {
					constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, uncast)
					if constraint != nil && !utils.IsNullableType(ctx.TypeChecker, constraint) {
						return true
					}
				}
				return false
			}

			// Check shape and assignability before recursively walking nested types. Assertions between
			// incompatible callable types can otherwise traverse very large generic parameter graphs.
			if !hasSameProperties(uncast, cast) ||
				!haveSameTypeArguments(uncast, cast) ||
				!areMutuallyAssignable(uncast, cast) {
				return false
			}

			return !containsAny(uncast) &&
				!containsAny(cast) &&
				!(containsTypeVariable(cast) && !containsTypeVariable(uncast))
		}

		isTypeAny := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsAny)
		}

		isTypeUnknown := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsUnknown)
		}

		isNullableForNonNullAssertion := func(t *checker.Type) bool {
			if utils.IsNullableType(ctx.TypeChecker, t) {
				return true
			}
			for _, part := range utils.UnionTypeParts(t) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsVoid) {
					return true
				}
			}
			return false
		}

		isIIFE := func(expression *ast.Node) bool {
			expression = ast.SkipParentheses(expression)
			if !ast.IsCallExpression(expression) {
				return false
			}

			callee := ast.SkipParentheses(expression.AsCallExpression().Expression)
			return ast.IsArrowFunction(callee) || ast.IsFunctionExpression(callee)
		}

		var isContextSensitiveCallLikeExpression func(expression *ast.Node) bool
		isContextSensitiveCallLikeExpression = func(expression *ast.Node) bool {
			if ast.IsCallExpression(expression) || ast.IsNewExpression(expression) || ast.IsTaggedTemplateExpression(expression) {
				return true
			}

			if ast.IsAwaitExpression(expression) {
				return isContextSensitiveCallLikeExpression(ast.SkipParentheses(expression.Expression()))
			}

			return false
		}

		getUncastType := func(node *ast.Node) *checker.Type {
			expression := ast.SkipParentheses(node.Expression())

			if isIIFE(expression) {
				callee := ast.SkipParentheses(expression.AsCallExpression().Expression)
				functionType := ctx.TypeChecker.GetTypeAtLocation(callee)
				signatures := ctx.TypeChecker.GetCallSignatures(functionType)
				if len(signatures) > 0 {
					returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
					if callee.Type() == nil && utils.IsTypeFlagSet(returnType, checker.TypeFlagsUndefined) {
						return ctx.TypeChecker.GetVoidType()
					}
					return returnType
				}
			}

			// For call-like expressions, use the context-free expression type so
			// contextual typing from the assertion itself doesn't leak into generic
			// inference for the original expression.
			if isContextSensitiveCallLikeExpression(expression) {
				if t := checker.Checker_getContextFreeTypeOfExpression(ctx.TypeChecker, expression); t != nil {
					return t
				}
			}

			return ctx.TypeChecker.GetTypeAtLocation(expression)
		}

		getOriginalExpression := func(node *ast.Node) *ast.Node {
			current := ast.SkipParentheses(node.Expression())
			for ast.IsAsExpression(current) || ast.IsTypeAssertion(current) {
				current = ast.SkipParentheses(current.Expression())
			}
			return current
		}

		isArgumentToParentCallOrNew := func(node *ast.Node) (bool, int) {
			parent := parentThroughParens(node)
			if parent == nil || (!ast.IsCallExpression(parent) && !ast.IsNewExpression(parent)) {
				return false, -1
			}
			for i, argument := range parent.Arguments() {
				if argument == node || ast.SkipParentheses(argument) == node {
					return true, i
				}
			}
			return false, -1
		}

		hasTypeParams := func(sig *checker.Signature) bool {
			return len(sig.TypeParameters()) > 0
		}

		hasGenericCallSignature := func(t *checker.Type) bool {
			return slices.ContainsFunc(utils.GetCallSignatures(ctx.TypeChecker, t), hasTypeParams)
		}

		hasGenericInferenceParameterAtArgument := func(callOrNew *ast.Node, argIndex int, elementPath []int) bool {
			signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, callOrNew, nil, checker.CheckModeNormal)
			if signature == nil {
				return false
			}
			for signature.Target() != nil {
				signature = signature.Target()
			}
			if len(signature.TypeParameters()) == 0 {
				return false
			}

			params := checker.Signature_parameters(signature)
			if len(params) == 0 {
				return false
			}

			paramIndex := argIndex
			if paramIndex >= len(params) {
				paramIndex = len(params) - 1
			}
			param := params[paramIndex]
			paramType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, param)
			if valueDeclaration := param.ValueDeclaration; valueDeclaration != nil &&
				valueDeclaration.Kind == ast.KindParameter &&
				valueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
				if typeArguments := getTypeArguments(paramType); len(typeArguments) > 0 {
					typeArgumentIndex := 0
					if len(typeArguments) > 1 {
						typeArgumentIndex = min(argIndex-paramIndex, len(typeArguments)-1)
					}
					paramType = typeArguments[typeArgumentIndex]
				}
			}
			for _, elementIndex := range elementPath {
				var elementType *checker.Type
				if checker.IsTupleType(paramType) {
					typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, paramType)
					if len(typeArguments) > 0 {
						elementType = typeArguments[min(elementIndex, len(typeArguments)-1)]
					}
				} else {
					elementType = utils.GetNumberIndexType(ctx.TypeChecker, paramType)
				}
				if elementType == nil {
					break
				}
				paramType = elementType
			}

			return containsTypeVariable(paramType)
		}

		genericsMismatch := func(uncast, contextual *checker.Type) bool {
			return slices.ContainsFunc(checker.Checker_getPropertiesOfType(ctx.TypeChecker, contextual), func(prop *ast.Symbol) bool {
				contextualSigs := checker.Checker_getSignaturesOfType(
					ctx.TypeChecker,
					checker.Checker_getTypeOfSymbol(ctx.TypeChecker, prop),
					checker.SignatureKindCall,
				)
				if !slices.ContainsFunc(contextualSigs, hasTypeParams) {
					return false
				}

				uncastProp := checker.Checker_getPropertyOfType(ctx.TypeChecker, uncast, prop.Name)
				if uncastProp == nil {
					return true
				}

				uncastSigs := checker.Checker_getSignaturesOfType(
					ctx.TypeChecker,
					checker.Checker_getTypeOfSymbol(ctx.TypeChecker, uncastProp),
					checker.SignatureKindCall,
				)
				return !slices.ContainsFunc(uncastSigs, hasTypeParams)
			})
		}

		isArgumentToOverloadedFunction := func(node *ast.Node) bool {
			isArg, argIndex := isArgumentToParentCallOrNew(node)
			if !isArg {
				return false
			}

			parent := parentThroughParens(node)
			calleeType := checker.Checker_GetNonNullableType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(parent.Expression()))
			signatures := ctx.TypeChecker.GetCallSignatures(calleeType)
			if len(signatures) <= 1 {
				return false
			}

			paramTypes := make([]*checker.Type, 0, len(signatures))
			for _, sig := range signatures {
				params := sig.Parameters()
				if argIndex >= len(params) {
					return true
				}
				paramType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, params[argIndex])
				if valueDeclaration := params[argIndex].ValueDeclaration; valueDeclaration != nil &&
					valueDeclaration.Kind == ast.KindParameter &&
					valueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
					if typeArguments := getTypeArguments(paramType); len(typeArguments) > 0 {
						paramType = typeArguments[0]
					}
				}
				if paramType == nil {
					return true
				}
				paramTypes = append(paramTypes, paramType)
			}

			firstParamType := paramTypes[0]
			if slices.ContainsFunc(paramTypes, func(paramType *checker.Type) bool { return paramType != firstParamType }) {
				uncastType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
				return slices.ContainsFunc(paramTypes, func(paramType *checker.Type) bool {
					return !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, paramType)
				})
			}
			return false
		}

		isInDestructuringDeclaration := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			return ast.IsVariableDeclaration(parent) &&
				ast.SkipParentheses(parent.Initializer()) == node &&
				parent.Name() != nil && ast.IsBindingPattern(parent.Name())
		}

		isPropertyInProblematicContext := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			if parent == nil || !ast.IsPropertyAssignment(parent) || ast.SkipParentheses(parent.Initializer()) != node {
				return false
			}
			objectExpr := parent.Parent
			if objectExpr == nil || !ast.IsObjectLiteralExpression(objectExpr) {
				return false
			}
			if objectContextualType := checker.Checker_getContextualType(ctx.TypeChecker, objectExpr, checker.ContextFlagsNone); objectContextualType != nil && utils.IsUnionType(objectContextualType) {
				propContextualType := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
				if propContextualType == nil {
					return true
				}
				nonNullableContextualType := checker.Checker_GetNonNullableType(ctx.TypeChecker, propContextualType)
				if utils.IsUnionType(nonNullableContextualType) {
					return true
				}
				uncastType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
				return !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, nonNullableContextualType)
			}
			objectParent := parentThroughParens(objectExpr)
			// Also preserve casts whose property context comes from another assertion.
			// typescript-go uses that context during inference; removing the inner cast
			// can change the inferred property type.
			return objectParent != nil &&
				(ast.IsAsExpression(objectParent) ||
					ast.IsTypeAssertion(objectParent) ||
					ast.IsSatisfiesExpression(objectParent) ||
					(ast.IsCallExpression(objectParent) && objectParent.Parent != nil && ast.IsSatisfiesExpression(objectParent.Parent)))
		}

		isAssignmentInNonStatementContext := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			return parent != nil &&
				ast.IsAssignmentExpression(parent, false) &&
				ast.SkipParentheses(parent.AsBinaryExpression().Right) == node &&
				(parentThroughParens(parent) == nil || parentThroughParens(parent).Kind != ast.KindExpressionStatement)
		}

		isRightHandSideOfLogicalAssignment := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			return parent != nil &&
				ast.IsBinaryExpression(parent) &&
				ast.SkipParentheses(parent.AsBinaryExpression().Right) == node &&
				ast.IsLogicalOrCoalescingAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind)
		}

		isNestedInArrayLiteralArgumentToGenericCall := func(node *ast.Node) bool {
			// Local safeguard: contextual acceptance alone does not preserve inference
			// for a generic parameter inferred from an array element.
			elementPath := []int{}
			for child, current := node, node.Parent; current != nil; child, current = current, current.Parent {
				if ast.IsFunctionExpression(current) || ast.IsArrowFunction(current) {
					return false
				}
				if !ast.IsArrayLiteralExpression(current) {
					continue
				}
				elementIndex := slices.IndexFunc(current.AsArrayLiteralExpression().Elements.Nodes, func(element *ast.Node) bool {
					return element == child || ast.SkipParentheses(element) == ast.SkipParentheses(child)
				})
				if elementIndex == -1 {
					continue
				}
				elementPath = append([]int{elementIndex}, elementPath...)

				callArgument := current
				parent := parentThroughParens(callArgument)
				spreadOffset := 0
				if parent != nil && ast.IsSpreadElement(parent) {
					spreadOffset = elementIndex
					callArgument = parent
					parent = parentThroughParens(callArgument)
				}
				if parent == nil || (!ast.IsCallExpression(parent) && !ast.IsNewExpression(parent)) {
					continue
				}
				if parent.TypeArguments() != nil {
					return false
				}

				argIndex := slices.IndexFunc(parent.Arguments(), func(candidate *ast.Node) bool {
					return candidate == callArgument || ast.SkipParentheses(candidate) == callArgument
				})
				if argIndex == -1 {
					continue
				}

				parameterElementPath := elementPath
				if ast.IsSpreadElement(callArgument) {
					parameterElementPath = parameterElementPath[1:]
				}
				return hasGenericInferenceParameterAtArgument(parent, argIndex+spreadOffset, parameterElementPath)
			}
			return false
		}

		isInGenericContext := func(node *ast.Node) bool {
			seenFunction := false
			for current := node.Parent; current != nil; current = current.Parent {
				if current.Kind == ast.KindFunctionDeclaration {
					return false
				}
				if ast.IsFunctionExpression(current) || ast.IsArrowFunction(current) {
					if current.Body() != nil && current.Body().Kind == ast.KindBlock {
						return false
					}
					if seenFunction {
						return false
					}
					seenFunction = true
				}
				if ast.IsCallExpression(current) || ast.IsNewExpression(current) {
					if current.TypeArguments() != nil {
						continue
					}
					if ast.IsCallExpression(current) && ast.IsAccessExpression(current.Expression()) {
						if slices.ContainsFunc(current.Arguments(), func(argument *ast.Node) bool {
							return ast.SkipParentheses(argument) == node
						}) {
							continue
						}
					}
					calleeType := ctx.TypeChecker.GetTypeAtLocation(current.Expression())
					if hasGenericCallSignature(calleeType) {
						return true
					}
				}
			}
			return false
		}

		isPropertyInInferredCallbackReturn := func(node *ast.Node) bool {
			// Local safeguard for the same inference dependency in callback returns.
			parent := parentThroughParens(node)
			if parent == nil || !ast.IsPropertyAssignment(parent) || ast.SkipParentheses(parent.Initializer()) != node {
				return false
			}
			objectExpr := parent.Parent
			if objectExpr == nil || !ast.IsObjectLiteralExpression(objectExpr) {
				return false
			}

			callback := parentThroughParens(objectExpr)
			return callback != nil &&
				ast.IsArrowFunction(callback) &&
				ast.SkipParentheses(callback.Body()) == objectExpr &&
				isInGenericContext(node)
		}

		isSkipParentType := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			return parent != nil &&
				(ast.IsAsExpression(parent) ||
					ast.IsTypeAssertion(parent) ||
					parent.Kind == ast.KindSpreadElement ||
					parent.Kind == ast.KindSpreadAssignment ||
					ast.IsSatisfiesExpression(parent))
		}

		shouldSkipContextualTypeFallback := func(node *ast.Node, castIsAny bool, uncastType, castType *checker.Type) bool {
			parent := parentThroughParens(node)
			// An assignment can narrow the receiver for subsequent statements.
			// Accepting the original type does not make that narrowing unnecessary.
			if isInNarrowingAssignment(ctx, node, uncastType, castType) {
				return true
			}
			if castIsAny {
				return (parent != nil && ast.IsLogicalExpression(parent)) ||
					isInGenericContext(node) ||
					isPropertyInProblematicContext(node)
			}

			// Interpolated templates can widen to string even when the context accepts them.
			// https://github.com/typescript-eslint/typescript-eslint/issues/12276
			if ast.IsTemplateExpression(ast.SkipParentheses(node.Expression())) {
				return true
			}

			if isSkipParentType(node) ||
				ast.IsArrayLiteralExpression(ast.SkipParentheses(node.Expression())) ||
				isNestedInArrayLiteralArgumentToGenericCall(node) ||
				isInDestructuringDeclaration(node) ||
				isPropertyInProblematicContext(node) ||
				isPropertyInInferredCallbackReturn(node) ||
				isAssignmentInNonStatementContext(node) ||
				isRightHandSideOfLogicalAssignment(node) ||
				isArgumentToOverloadedFunction(node) {
				return true
			}

			if isInGenericContext(node) {
				originalExpr := getOriginalExpression(node)
				return !isConceptuallyLiteral(originalExpr) &&
					(parent == nil || !ast.IsPropertyAssignment(parent))
			}

			return false
		}

		hasPhantomTypeArgumentMismatch := func(node *ast.Node, uncastType, contextualType *checker.Type) bool {
			return isInGenericContext(node) &&
				(hasPhantomTypeArguments(uncastType) ||
					hasPhantomTypeArguments(contextualType)) &&
				!haveSameTypeArguments(uncastType, contextualType)
		}

		isNullishLiteralToUnion := func(node *ast.Node, castType *checker.Type) bool {
			expression := ast.SkipParentheses(node.Expression())
			return utils.IsUnionType(castType) &&
				(expression.Kind == ast.KindNullKeyword ||
					(ast.IsIdentifier(expression) && expression.Text() == "undefined"))
		}

		isDoubleAssertionUnnecessary := func(node *ast.Node, contextualType *checker.Type) string {
			innerExpression := ast.SkipParentheses(node.Expression())
			if !ast.IsAsExpression(innerExpression) && !ast.IsTypeAssertion(innerExpression) {
				return ""
			}

			originalExpr := getOriginalExpression(node)
			originalType := ctx.TypeChecker.GetTypeAtLocation(originalExpr)
			castType := ctx.TypeChecker.GetTypeAtLocation(node)
			var isConstrainedTo func(source, target *checker.Type, seen map[*checker.Type]struct{}) bool
			isConstrainedTo = func(source, target *checker.Type, seen map[*checker.Type]struct{}) bool {
				if source == target {
					return true
				}
				if source == nil {
					return false
				}
				if _, ok := seen[source]; ok {
					return false
				}
				seen[source] = struct{}{}

				if utils.IsTypeParameter(source) {
					return isConstrainedTo(ctx.TypeChecker.GetConstraintOfTypeParameter(source), target, seen)
				}
				if utils.IsIntersectionType(source) {
					return slices.ContainsFunc(source.Types(), func(part *checker.Type) bool {
						return isConstrainedTo(part, target, seen)
					})
				}
				return false
			}
			differentUnrelatedTypeParameters := originalType != castType &&
				utils.IsTypeParameter(originalType) &&
				utils.IsTypeParameter(castType) &&
				!isConstrainedTo(originalType, castType, map[*checker.Type]struct{}{})

			if isTypeUnchanged(node, innerExpression, originalType, castType) && !isTypeAny(castType) {
				return "unnecessaryAssertion"
			}
			if contextualType != nil && !differentUnrelatedTypeParameters {
				// Keep bridges between unrelated type parameters: typescript-go can
				// accept their contextual constraints without accepting the direct cast.
				intermediateType := ctx.TypeChecker.GetTypeAtLocation(innerExpression)
				if (isTypeAny(intermediateType) || isTypeUnknown(intermediateType)) &&
					checker.Checker_isTypeAssignableTo(ctx.TypeChecker, originalType, contextualType) {
					return "contextuallyUnnecessary"
				}
			}
			return ""
		}

		reportDoubleAssertionIfUnnecessary := func(node *ast.Node, contextualType *checker.Type) {
			messageId := isDoubleAssertionUnnecessary(node, contextualType)
			if messageId == "" {
				return
			}

			description := buildContextuallyUnnecessaryMessage(assertionRange(ctx, node)).Message.Description
			if messageId == "unnecessaryAssertion" {
				description = buildUnnecessaryAssertionDiagnostic(node.Loc).Message.Description
			}

			ctx.ReportDiagnosticWithFixes(rule.RuleDiagnostic{
				Range: assertionRange(ctx, node),
				Message: rule.RuleMessage{
					Id:          messageId,
					Description: description,
				},
			}, func() []rule.RuleFix {
				originalExpr := getOriginalExpression(node)
				textRange := utils.TrimNodeTextRange(ctx.SourceFile, originalExpr)
				text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
				if ast.IsObjectLiteralExpression(originalExpr) &&
					node.Parent != nil &&
					ast.IsArrowFunction(node.Parent) &&
					node.Parent.Body() == node {
					text = "(" + text + ")"
				}
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, text)}
			})
		}

		checkTypeAssertion := func(node *ast.Node) {
			typeNode := ast.SkipTypeParentheses(node.Type())
			typeAnnotationRange := utils.TrimNodeTextRange(ctx.SourceFile, typeNode)
			if slices.Contains(opts.TypesToIgnore, ctx.SourceFile.Text()[typeAnnotationRange.Pos():typeAnnotationRange.End()]) {
				return
			}

			castType := ctx.TypeChecker.GetTypeAtLocation(node)
			castTypeIsLiteral := isTypeLiteral(castType)
			typeAnnotationIsConstAssertion := isConstAssertion(typeNode)

			if !opts.CheckLiteralConstAssertions && castTypeIsLiteral && typeAnnotationIsConstAssertion {
				return
			}

			expression := node.Expression()
			uncastType := getUncastType(node)

			expressionForType := ast.SkipParentheses(expression)
			if uncastType == castType && ast.IsIdentifier(expressionForType) {
				// typescript-go may resolve a conditional type at the assertion site.
				// Retain its declared form when deciding whether the cast changed it.
				if symbol := ctx.TypeChecker.GetSymbolAtLocation(expressionForType); symbol != nil {
					symbolType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol)
					if symbolType != nil && checker.Type_flags(symbolType)&checker.TypeFlagsConditional != 0 {
						uncastType = symbolType
					}
				}
			}

			typeIsUnchanged := isTypeUnchanged(node, expression, uncastType, castType)

			var wouldSameTypeBeInferred bool
			if castTypeIsLiteral {
				wouldSameTypeBeInferred = isImplicitlyNarrowedLiteralDeclaration(node)
			} else {
				wouldSameTypeBeInferred = !typeAnnotationIsConstAssertion
			}

			if typeIsUnchanged && wouldSameTypeBeInferred {
				ctx.ReportDiagnosticWithFixes(buildUnnecessaryAssertionDiagnostic(assertionRange(ctx, node)), func() []rule.RuleFix {
					return createAssertionFixer(ctx, node)
				})
				return
			}

			castIsAny := isTypeAny(castType) && !isSkipParentType(node)
			var contextualType *checker.Type
			if !shouldSkipContextualTypeFallback(node, castIsAny, uncastType, castType) {
				contextualType = checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
			}

			if contextualType != nil {
				contextualTypeIsAny := isTypeAny(contextualType)
				isCallArgument, _ := isArgumentToParentCallOrNew(node)
				anyInvolvedInContextualCheck := (!contextualTypeIsAny && !containsAny(contextualType)) ||
					(contextualTypeIsAny && isCallArgument && !containsAny(castType))

				isContextuallyUnnecessary := !typeAnnotationIsConstAssertion &&
					!containsAny(uncastType) &&
					anyInvolvedInContextualCheck &&
					!hasPhantomTypeArgumentMismatch(node, uncastType, contextualType) &&
					(castIsAny || !genericsMismatch(uncastType, contextualType)) &&
					(contextualTypeIsAny || checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, contextualType)) &&
					!isNullishLiteralToUnion(node, castType)

				if isContextuallyUnnecessary {
					ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(assertionRange(ctx, node)), func() []rule.RuleFix {
						return createAssertionFixer(ctx, node)
					})
					return
				}
			}

			reportDoubleAssertionIfUnnecessary(node, contextualType)
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkTypeAssertion,
			ast.KindTypeAssertionExpression: checkTypeAssertion,

			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := ast.SkipParentheses(node.Expression())

				getExclamationTokenRange := func() core.TextRange {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Expression().End())
					return s.TokenRange()
				}

				buildRemoveExclamationFix := func(exclamation core.TextRange) rule.RuleFix {
					return rule.RuleFixRemoveRange(exclamation)
				}

				parent := parentThroughParens(node)
				if ast.IsAssignmentExpression(parent, true) {
					if ast.SkipParentheses(parent.AsBinaryExpression().Left) == node {
						exclamationRange := getExclamationTokenRange()
						ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(assertionRange(ctx, node)), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix(exclamationRange)} })
					}
					// for all other = assignments we ignore non-null checks
					// this is because non-null assertions can change the type-flow of the code
					// so whilst they might be unnecessary for the assignment - they are necessary
					// for following code
					return
				}

				constrainedType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
				actualType := ctx.TypeChecker.GetTypeAtLocation(expression)

				constrainedTypeIsNullable := isNullableForNonNullAssertion(constrainedType)
				actualTypeIsNullable := isNullableForNonNullAssertion(actualType)

				if !constrainedTypeIsNullable && !actualTypeIsNullable {
					if ast.IsIdentifier(expression) && isPossiblyUsedBeforeAssigned(expression) {
						return
					}
					exclamationRange := getExclamationTokenRange()
					ctx.ReportDiagnosticWithFixes(
						buildUnnecessaryAssertionDiagnostic(assertionRange(ctx, node)),
						func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix(exclamationRange)} },
					)
				} else {
					// we know it's a nullable type
					// so figure out if the variable is used in a place that accepts nullable types
					if constrainedType != actualType {
						return
					}

					var tFlags checker.TypeFlags
					for _, part := range utils.UnionTypeParts(constrainedType) {
						tFlags |= checker.Type_flags(part)
					}

					contextualType := utils.GetContextualType(ctx.TypeChecker, node)
					if contextualType != nil {
						var contextualFlags checker.TypeFlags
						for _, part := range utils.UnionTypeParts(contextualType) {
							contextualFlags |= checker.Type_flags(part)
						}

						if tFlags&checker.TypeFlagsUnknown != 0 && contextualFlags&checker.TypeFlagsUnknown == 0 {
							return
						}

						// in strict mode you can't assign null to undefined, so we have to make sure that
						// the two types share a nullable type
						typeIncludesUndefined := tFlags&checker.TypeFlagsUndefined != 0
						typeIncludesNull := tFlags&checker.TypeFlagsNull != 0
						typeIncludesVoid := tFlags&checker.TypeFlagsVoid != 0

						contextualTypeIncludesUndefined := contextualFlags&checker.TypeFlagsUndefined != 0
						contextualTypeIncludesNull := contextualFlags&checker.TypeFlagsNull != 0
						contextualTypeIncludesVoid := contextualFlags&checker.TypeFlagsVoid != 0

						// make sure that the parent accepts the same types
						// i.e. assigning `string | null | undefined` to `string | undefined` is invalid
						isValidUndefined := !typeIncludesUndefined || contextualTypeIncludesUndefined
						isValidNull := !typeIncludesNull || contextualTypeIncludesNull
						isValidVoid := !typeIncludesVoid || contextualTypeIncludesVoid

						if isValidUndefined && isValidNull && isValidVoid {
							exclamationRange := getExclamationTokenRange()
							ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(assertionRange(ctx, node)), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix(exclamationRange)} })
						}
					}
				}
			},
		}
	},
}
