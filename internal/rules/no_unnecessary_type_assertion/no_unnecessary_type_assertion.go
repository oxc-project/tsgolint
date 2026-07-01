package no_unnecessary_type_assertion

import (
	"fmt"
	"slices"
	"strings"

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
func buildUnnecessaryAssertionDiagnostic(assertion core.TextRange, expression core.TextRange, expressionType string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: assertion,
		Message: rule.RuleMessage{
			Id:          "unnecessaryAssertion",
			Description: "This assertion is unnecessary since it does not change the type of the expression.",
		},
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("This expression already has the type '%s'", expressionType),
				Range: expression,
			},
		},
	}
}

func buildUnnecessaryTypeAssertionDiagnostic(assertion core.TextRange, expression core.TextRange, expressionType string, assertedType string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: assertion,
		Message: rule.RuleMessage{
			Id:          "unnecessaryAssertion",
			Description: "This assertion is unnecessary since it does not change the type of the expression.",
		},
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("This expression already has the type '%s'", expressionType),
				Range: expression,
			},
			{
				Label: fmt.Sprintf("Casting it to '%s' is unnecessary", assertedType),
				Range: assertion,
			},
		},
	}
}

var NoUnnecessaryTypeAssertionRule = rule.Rule{
	Name: "no-unnecessary-type-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoUnnecessaryTypeAssertionOptions](options, "no-unnecessary-type-assertion")

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

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
			expression := node.Expression()
			/**
			 * Even on `const` variable declarations, template literals with expressions can sometimes be widened without a type assertion.
			 * @see https://github.com/typescript-eslint/typescript-eslint/issues/8737
			 */
			if ast.IsTemplateExpression(expression) {
				return false
			}

			return (ast.IsVariableDeclaration(node.Parent) && ast.IsVariableDeclarationList(node.Parent.Parent) && node.Parent.Parent.Flags&ast.NodeFlagsConst != 0) ||
				(ast.IsPropertyDeclaration(node.Parent) && node.Parent.ModifierFlags()&ast.ModifierFlagsReadonly != 0)

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

		var typeContains func(t *checker.Type, predicate func(*checker.Type) bool, seen map[*checker.Type]struct{}) bool
		typeContains = func(t *checker.Type, predicate func(*checker.Type) bool, seen map[*checker.Type]struct{}) bool {
			if t == nil {
				return false
			}
			if _, ok := seen[t]; ok {
				return false
			}
			seen[t] = struct{}{}
			if predicate(t) {
				return true
			}
			for _, part := range utils.UnionTypeParts(t) {
				if part != t && typeContains(part, predicate, seen) {
					return true
				}
			}
			for _, part := range utils.IntersectionTypeParts(t) {
				if part != t && typeContains(part, predicate, seen) {
					return true
				}
			}
			for _, typeArgument := range getTypeArguments(t) {
				if typeContains(typeArgument, predicate, seen) {
					return true
				}
			}
			for _, sig := range utils.GetCallSignatures(ctx.TypeChecker, t) {
				if typeContains(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig), predicate, seen) {
					return true
				}
				for _, param := range checker.Signature_parameters(sig) {
					if typeContains(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, param), predicate, seen) {
						return true
					}
				}
			}
			return false
		}

		containsAny := func(t *checker.Type) bool {
			return typeContains(t, func(part *checker.Type) bool {
				return utils.IsTypeFlagSet(part, checker.TypeFlagsAny)
			}, map[*checker.Type]struct{}{})
		}

		containsTypeVariable := func(t *checker.Type) bool {
			return typeContains(t, func(part *checker.Type) bool {
				return utils.IsTypeFlagSet(part, checker.TypeFlagsTypeVariable|checker.TypeFlagsIndex)
			}, map[*checker.Type]struct{}{})
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
			return ast.IsArrayLiteralExpression(node) ||
				ast.IsObjectLiteralExpression(node) ||
				ast.IsClassExpression(node) ||
				ast.IsFunctionExpression(node) ||
				ast.IsArrowFunction(node) ||
				ast.IsJsxElement(node) ||
				ast.IsJsxFragment(node) ||
				ast.IsStringLiteral(node) ||
				node.Kind == ast.KindNumericLiteral ||
				node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
				node.Kind == ast.KindTrueKeyword ||
				node.Kind == ast.KindFalseKeyword ||
				node.Kind == ast.KindNullKeyword ||
				ast.IsTemplateExpression(node)
		}

		isTypeLiteral := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral|checker.TypeFlagsBooleanLiteral)
		}

		isTypeUnchanged := func(node *ast.Node, expression *ast.Node, uncast, cast *checker.Type) bool {
			if uncast == cast {
				return true
			}

			typeNode := node.Type()
			if ast.IsIntersectionTypeNode(typeNode) && containsTypeVariable(cast) {
				return false
			}

			if compilerOptions.ExactOptionalPropertyTypes.IsTrue() &&
				utils.IsTypeFlagSet(uncast, checker.TypeFlagsUndefined) &&
				utils.IsTypeFlagSet(cast, checker.TypeFlagsUndefined) {
				return areUnionPartsEquivalentIgnoringUndefined(uncast, cast)
			}

			if (utils.IsTypeFlagSet(uncast, checker.TypeFlagsNonPrimitive) && !utils.IsTypeFlagSet(cast, checker.TypeFlagsNonPrimitive)) ||
				(hasIndexSignature(uncast) && !hasIndexSignature(cast)) ||
				containsAny(uncast) ||
				containsAny(cast) ||
				(containsTypeVariable(cast) && !containsTypeVariable(uncast)) {
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

			if !hasSameProperties(uncast, cast) || !haveSameTypeArguments(uncast, cast) {
				return false
			}

			return checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncast, cast) &&
				checker.Checker_isTypeAssignableTo(ctx.TypeChecker, cast, uncast)
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

		parentThroughParens := func(node *ast.Node) *ast.Node {
			parent := node.Parent
			for parent != nil && ast.IsParenthesizedExpression(parent) {
				parent = parent.Parent
			}
			return parent
		}

		buildAssertionFixes := func(node *ast.Node) []rule.RuleFix {
			typeNode := node.Type()
			expression := node.Expression()
			if node.Kind == ast.KindAsExpression {
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
				asKeywordRange := s.TokenRange()
				typeNodeRange := typeNode.Loc

				for {
					previousCharPos := asKeywordRange.Pos() - 1
					if previousCharPos < expression.End() {
						break
					}
					previousChar := ctx.SourceFile.Text()[previousCharPos]
					if !utils.IsStrWhiteSpace(rune(previousChar)) {
						break
					}
					asKeywordRange = asKeywordRange.WithPos(previousCharPos)
				}

				typeNodePos := utils.TrimNodeTextRange(ctx.SourceFile, typeNode).Pos()
				if asKeywordRange.End() > typeNodePos {
					return []rule.RuleFix{
						rule.RuleFixRemoveRange(core.NewTextRange(expression.End(), typeNode.Loc.End())),
					}
				}
				betweenText := ctx.SourceFile.Text()[asKeywordRange.End():typeNodePos]
				if !utils.IsStringWhiteSpace(betweenText) {
					return []rule.RuleFix{
						rule.RuleFixRemoveRange(asKeywordRange),
						rule.RuleFixRemove(ctx.SourceFile, typeNode),
					}
				}

				return []rule.RuleFix{
					rule.RuleFixRemoveRange(core.NewTextRange(asKeywordRange.Pos(), typeNodeRange.End())),
				}
			}

			s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
			openingAngleBracket := s.TokenRange()
			s.ResetPos(typeNode.End())
			s.Scan()
			closingAngleBracket := s.TokenRange()
			return []rule.RuleFix{rule.RuleFixRemoveRange(openingAngleBracket.WithEnd(closingAngleBracket.End()))}
		}

		reportUnnecessaryTypeAssertion := func(node *ast.Node, uncastType, castType *checker.Type) {
			typeNode := node.Type()
			expression := node.Expression()
			expressionForType := ast.SkipParentheses(expression)

			if typeNode.Pos() < expression.Pos() {
				searchStart := node.Pos()
				if ast.IsParenthesizedExpression(node.Parent) {
					searchStart = node.Parent.Pos()
				}

				beforeExpression := ctx.SourceFile.Text()[searchStart:expression.Pos()]
				commentStartOffset := strings.LastIndex(beforeExpression, "/**")
				commentEndOffset := strings.LastIndex(beforeExpression, "*/")
				if commentStartOffset != -1 && commentEndOffset != -1 && commentEndOffset >= commentStartOffset {
					commentStart := searchStart + commentStartOffset
					commentEnd := searchStart + commentEndOffset + len("*/")
					fixEnd := commentEnd
					for fixEnd < expression.Pos() && utils.IsStrWhiteSpace(rune(ctx.SourceFile.Text()[fixEnd])) {
						fixEnd++
					}

					assertionRange := core.NewTextRange(commentStart, commentEnd)
					ctx.ReportDiagnosticWithFixes(
						buildUnnecessaryTypeAssertionDiagnostic(
							assertionRange,
							utils.TrimNodeTextRange(ctx.SourceFile, expressionForType),
							ctx.TypeChecker.TypeToString(uncastType),
							ctx.TypeChecker.TypeToString(castType),
						),
						func() []rule.RuleFix {
							return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(commentStart, fixEnd))}
						},
					)
					return
				}
			}

			if node.Kind == ast.KindAsExpression {
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
				asKeywordRange := s.TokenRange()
				assertionRange := asKeywordRange.WithEnd(typeNode.Loc.End())
				ctx.ReportDiagnosticWithFixes(
					buildUnnecessaryTypeAssertionDiagnostic(
						assertionRange,
						utils.TrimNodeTextRange(ctx.SourceFile, expressionForType),
						ctx.TypeChecker.TypeToString(uncastType),
						ctx.TypeChecker.TypeToString(castType),
					), func() []rule.RuleFix {
						return buildAssertionFixes(node)
					})
				return
			}

			{
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
				openingAngleBracket := s.TokenRange()
				s.ResetPos(typeNode.End())
				s.Scan()
				closingAngleBracket := s.TokenRange()
				assertionRange := openingAngleBracket.WithEnd(closingAngleBracket.End())
				ctx.ReportDiagnosticWithFixes(
					buildUnnecessaryTypeAssertionDiagnostic(
						assertionRange,
						utils.TrimNodeTextRange(ctx.SourceFile, expressionForType),
						ctx.TypeChecker.TypeToString(uncastType),
						ctx.TypeChecker.TypeToString(castType),
					),
					func() []rule.RuleFix {
						return []rule.RuleFix{rule.RuleFixRemoveRange(assertionRange)}
					})
			}
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

		hasGenericCallSignature := func(t *checker.Type) bool {
			return slices.ContainsFunc(utils.GetCallSignatures(ctx.TypeChecker, t), func(sig *checker.Signature) bool {
				return len(sig.TypeParameters()) > 0
			})
		}

		hasGenericInferenceParameterAtArgument := func(callOrNew *ast.Node, argIndex int) bool {
			calleeType := ctx.TypeChecker.GetTypeAtLocation(callOrNew.Expression())
			var signatures []*checker.Signature
			if ast.IsNewExpression(callOrNew) {
				signatures = utils.GetConstructSignatures(ctx.TypeChecker, calleeType)
			} else {
				signatures = utils.GetCallSignatures(ctx.TypeChecker, calleeType)
			}

			return slices.ContainsFunc(signatures, func(sig *checker.Signature) bool {
				if len(sig.TypeParameters()) == 0 {
					return false
				}

				params := checker.Signature_parameters(sig)
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
						paramType = typeArguments[0]
					}
				}

				return containsTypeVariable(paramType)
			})
		}

		genericsMismatch := func(uncast, contextual *checker.Type) bool {
			return slices.ContainsFunc(checker.Checker_getPropertiesOfType(ctx.TypeChecker, contextual), func(prop *ast.Symbol) bool {
				contextualSigs := checker.Checker_getSignaturesOfType(
					ctx.TypeChecker,
					checker.Checker_getTypeOfSymbol(ctx.TypeChecker, prop),
					checker.SignatureKindCall,
				)
				if !slices.ContainsFunc(contextualSigs, func(sig *checker.Signature) bool {
					return len(sig.TypeParameters()) > 0
				}) {
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
				return !slices.ContainsFunc(uncastSigs, func(sig *checker.Signature) bool {
					return len(sig.TypeParameters()) > 0
				})
			})
		}

		isArgumentToOverloadedFunction := func(node *ast.Node) bool {
			isArg, argIndex := isArgumentToParentCallOrNew(node)
			if !isArg {
				return false
			}

			parent := parentThroughParens(node)
			calleeType := ctx.TypeChecker.GetTypeAtLocation(parent.Expression())
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
			return ast.IsVariableDeclaration(node.Parent) &&
				node.Parent.Initializer() == node &&
				node.Parent.Name() != nil &&
				ast.IsBindingPattern(node.Parent.Name())
		}

		isPropertyInProblematicContext := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil || !ast.IsPropertyAssignment(parent) || parent.Initializer() != node {
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
			objectParent := objectExpr.Parent
			return objectParent != nil &&
				(ast.IsSatisfiesExpression(objectParent) ||
					(ast.IsCallExpression(objectParent) && objectParent.Parent != nil && ast.IsSatisfiesExpression(objectParent.Parent)))
		}

		isAssignmentInNonStatementContext := func(node *ast.Node) bool {
			parent := node.Parent
			return parent != nil &&
				ast.IsAssignmentExpression(parent, false) &&
				parent.AsBinaryExpression().Right == node &&
				(parent.Parent == nil || parent.Parent.Kind != ast.KindExpressionStatement)
		}

		isRightHandSideOfLogicalAssignment := func(node *ast.Node) bool {
			parent := node.Parent
			return parent != nil &&
				ast.IsBinaryExpression(parent) &&
				parent.AsBinaryExpression().Right == node &&
				ast.IsLogicalOrCoalescingAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind)
		}

		isNestedInArrayLiteralArgumentToGenericCall := func(node *ast.Node) bool {
			for current := node.Parent; current != nil; current = current.Parent {
				if ast.IsFunctionExpression(current) || ast.IsArrowFunction(current) {
					return false
				}
				if !ast.IsArrayLiteralExpression(current) {
					continue
				}

				parent := parentThroughParens(current)
				if parent == nil || (!ast.IsCallExpression(parent) && !ast.IsNewExpression(parent)) {
					continue
				}
				if parent.TypeArguments() != nil {
					return false
				}

				argIndex := slices.IndexFunc(parent.Arguments(), func(argument *ast.Node) bool {
					return argument == current || ast.SkipParentheses(argument) == current
				})
				if argIndex == -1 {
					continue
				}

				return hasGenericInferenceParameterAtArgument(parent, argIndex)
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
						if slices.Contains(current.Arguments(), node) {
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

		skipParentTypeForContextualAny := func(node *ast.Node) bool {
			parent := parentThroughParens(node)
			return parent != nil &&
				(ast.IsAsExpression(parent) ||
					ast.IsTypeAssertion(parent) ||
					parent.Kind == ast.KindSpreadElement ||
					parent.Kind == ast.KindSpreadAssignment ||
					ast.IsSatisfiesExpression(parent))
		}

		shouldSkipContextualTypeFallback := func(node *ast.Node, castIsAny bool) bool {
			parent := parentThroughParens(node)
			if castIsAny {
				return (parent != nil && ast.IsLogicalExpression(parent)) || isInGenericContext(node)
			}

			if skipParentTypeForContextualAny(node) ||
				ast.IsArrayLiteralExpression(node.Expression()) ||
				isNestedInArrayLiteralArgumentToGenericCall(node) ||
				isInDestructuringDeclaration(node) ||
				isPropertyInProblematicContext(node) ||
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

		reportDoubleAssertionIfUnnecessary := func(node *ast.Node, contextualType *checker.Type) bool {
			innerExpression := ast.SkipParentheses(node.Expression())
			if !ast.IsAsExpression(innerExpression) && !ast.IsTypeAssertion(innerExpression) {
				return false
			}

			originalExpr := getOriginalExpression(node)
			originalType := ctx.TypeChecker.GetTypeAtLocation(originalExpr)
			castType := ctx.TypeChecker.GetTypeAtLocation(node)

			messageId := ""
			if isTypeUnchanged(node, innerExpression, originalType, castType) && !isTypeAny(castType) {
				messageId = "unnecessaryAssertion"
			} else if contextualType != nil {
				intermediateType := ctx.TypeChecker.GetTypeAtLocation(innerExpression)
				if (isTypeAny(intermediateType) || isTypeUnknown(intermediateType)) &&
					checker.Checker_isTypeAssignableTo(ctx.TypeChecker, originalType, contextualType) {
					messageId = "contextuallyUnnecessary"
				}
			}
			if messageId == "" {
				return false
			}

			description := buildContextuallyUnnecessaryMessage(node.Loc).Message.Description
			if messageId == "unnecessaryAssertion" {
				description = buildUnnecessaryAssertionDiagnostic(node.Loc, originalExpr.Loc, ctx.TypeChecker.TypeToString(originalType)).Message.Description
			}

			ctx.ReportDiagnosticWithFixes(rule.RuleDiagnostic{
				Range: node.Loc,
				Message: rule.RuleMessage{
					Id:          messageId,
					Description: description,
				},
			}, func() []rule.RuleFix {
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
			return true
		}

		checkTypeAssertion := func(node *ast.Node) {
			typeNode := node.Type()
			if slices.Contains(opts.TypesToIgnore, strings.TrimSpace(ctx.SourceFile.Text()[typeNode.Pos():typeNode.End()])) {
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
				reportUnnecessaryTypeAssertion(node, uncastType, castType)
				return
			}

			castIsAny := isTypeAny(castType) && !skipParentTypeForContextualAny(node)
			var contextualType *checker.Type
			if !shouldSkipContextualTypeFallback(node, castIsAny) {
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
					ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(node.Loc), func() []rule.RuleFix {
						return buildAssertionFixes(node)
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
				expression := node.Expression()

				getExclamationTokenRange := func() core.TextRange {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					return s.TokenRange()
				}

				buildRemoveExclamationFix := func(exclamation core.TextRange) rule.RuleFix {
					return rule.RuleFixRemoveRange(exclamation)
				}

				if ast.IsAssignmentExpression(node.Parent, true) {
					if node.Parent.AsBinaryExpression().Left == node {
						exclamationRange := getExclamationTokenRange()
						ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(exclamationRange), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix(exclamationRange)} })
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
						buildUnnecessaryAssertionDiagnostic(
							exclamationRange,
							expression.Loc,
							ctx.TypeChecker.TypeToString(constrainedType),
						),
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
							ctx.ReportDiagnosticWithFixes(buildContextuallyUnnecessaryMessage(exclamationRange), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix(exclamationRange)} })
						}
					}
				}
			},
		}
	},
}
