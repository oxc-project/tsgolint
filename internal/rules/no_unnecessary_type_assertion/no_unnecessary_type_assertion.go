package no_unnecessary_type_assertion

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildContextuallyUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "contextuallyUnnecessary",
		Description: "This assertion is unnecessary since the receiver accepts the original type of the expression.",
	}
}
func buildUnnecessaryAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssertion",
		Description: "This assertion is unnecessary since it does not change the type of the expression.",
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

		isTypeUnchanged := func(uncast, cast *checker.Type) bool {
			if uncast == cast {
				return true
			}

			if compilerOptions.ExactOptionalPropertyTypes.IsFalseOrUnknown() {
				return false
			}

			// if !utils.IsTypeFlagSet(uncast, checker.TypeFlagsUndefined) || !utils.IsTypeFlagSet(cast, checker.TypeFlagsUndefined) || !compilerOptions.ExactOptionalPropertyTypes.IsTrue() {
			// 	return false
			// }

			uncastParts := utils.Set[*checker.Type]{}
			uncastHasUndefined := false
			for _, part := range utils.UnionTypeParts(uncast) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					uncastHasUndefined = true
				} else {
					uncastParts.Add(part)
				}
			}

			if !uncastHasUndefined {
				return false
			}

			uncastPartsCount := uncastParts.Len()

			castPartsCount := 0
			castHasUndefined := false
			for _, part := range utils.UnionTypeParts(cast) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					castHasUndefined = true
				} else {
					if !uncastParts.Has(part) {
						return false
					}
					castPartsCount++
					if castPartsCount > uncastPartsCount {
						return false
					}
				}
			}

			return castHasUndefined && uncastPartsCount == castPartsCount
		}

		checkTypeAssertion := func(node *ast.Node) {
			typeNode := node.Type()
			if slices.Contains(opts.TypesToIgnore, strings.TrimSpace(ctx.SourceFile.Text()[typeNode.Pos():typeNode.End()])) {
				return
			}

			castType := ctx.TypeChecker.GetTypeAtLocation(node)

			if !utils.IsTypeFlagSet(castType, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) {
				if isConstAssertion(typeNode) {
					return
				}
			} else {
				if !isImplicitlyNarrowedLiteralDeclaration(node) {
					return
				}
			}

			expression := node.Expression()
			uncastType := ctx.TypeChecker.GetTypeAtLocation(expression)
			if !isTypeUnchanged(uncastType, castType) {
				return
			}

			msg := buildUnnecessaryAssertionMessage()

			if node.Kind == ast.KindAsExpression {
				ctx.ReportNodeWithFixes(node, msg, func() []rule.RuleFix {
					// Remove everything from expression end to type end (including whitespace around 'as')
					return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(expression.End(), typeNode.End()))}
				})
			} else {
				ctx.ReportNodeWithFixes(node, msg, func() []rule.RuleFix {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
					openingAngleBracket := s.TokenRange()
					s.ResetPos(typeNode.End())
					s.Scan()
					closingAngleBracket := s.TokenRange()
					return []rule.RuleFix{rule.RuleFixRemoveRange(openingAngleBracket.WithEnd(closingAngleBracket.End()))}
				})
			}
			// TODO - add contextually unnecessary check for this
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkTypeAssertion,
			ast.KindTypeAssertionExpression: checkTypeAssertion,

			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.Expression()

				buildRemoveExclamationFix := func() rule.RuleFix {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					return rule.RuleFixRemoveRange(s.TokenRange())
				}

				if ast.IsAssignmentExpression(node.Parent, true) {
					if node.Parent.AsBinaryExpression().Left == node {
						ctx.ReportNodeWithFixes(node, buildContextuallyUnnecessaryMessage(), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix()} })
					}
					// for all other = assignments we ignore non-null checks
					// this is because non-null assertions can change the type-flow of the code
					// so whilst they might be unnecessary for the assignment - they are necessary
					// for following code
					return
				}

				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)

				var tFlags checker.TypeFlags
				for _, part := range utils.UnionTypeParts(t) {
					tFlags |= checker.Type_flags(part)
				}

				if tFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|
					checker.TypeFlagsNull|
					checker.TypeFlagsUndefined|
					checker.TypeFlagsVoid) == 0 {
					if ast.IsIdentifier(expression) && isPossiblyUsedBeforeAssigned(expression) {
						return
					}
					ctx.ReportNodeWithFixes(node, buildUnnecessaryAssertionMessage(), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix()} })
				} else {
					// we know it's a nullable type
					// so figure out if the variable is used in a place that accepts nullable types
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
							ctx.ReportNodeWithFixes(node, buildContextuallyUnnecessaryMessage(), func() []rule.RuleFix { return []rule.RuleFix{buildRemoveExclamationFix()} })
						}
					}
				}
			},
		}
	},
}
