package no_unnecessary_type_arguments

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildCanBeInferredMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "canBeInferred",
		Description: "This value can be trivially inferred for this type parameter, so it can be omitted.",
	}
}

func buildIsDefaultParameterValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "isDefaultParameterValue",
		Description: "This is the default value for this type parameter, so it can be omitted.",
	}
}

func isTypeContextDeclaration(decl *ast.Node) bool {
	return ast.IsTypeAliasDeclaration(decl) || ast.IsInterfaceDeclaration(decl)
}

func isInTypeContext(node *ast.Node) bool {
	return ast.IsTypeReferenceNode(node) || ast.IsInterfaceDeclaration(node.Parent) || ast.IsTypeReferenceNode(node.Parent) || (ast.IsHeritageClause(node.Parent) && node.Parent.AsHeritageClause().Token == ast.KindImplementsKeyword)
}

func isEmptyObjectType(typeChecker *checker.Checker, t *checker.Type) bool {
	if !utils.IsObjectType(t) {
		return false
	}

	if len(checker.Checker_getPropertiesOfType(typeChecker, t)) != 0 {
		return false
	}

	return len(checker.Checker_getIndexInfosOfType(typeChecker, t)) == 0
}

func areTypesEquivalent(typeChecker *checker.Checker, a *checker.Type, b *checker.Type) bool {
	// If either type is `any` (including unresolved `error`-adjacent cases) or `{}`,
	// only treat them as equivalent when they are literally the same type object.
	if utils.IsTypeAnyType(a) || utils.IsTypeAnyType(b) || isEmptyObjectType(typeChecker, a) || isEmptyObjectType(typeChecker, b) {
		return a == b
	}

	return checker.Checker_isTypeAssignableTo(typeChecker, a, b) && checker.Checker_isTypeAssignableTo(typeChecker, b, a)
}

var NoUnnecessaryTypeArgumentsRule = rule.Rule{
	Name: "no-unnecessary-type-arguments",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		getTypeParametersFromType := func(node *ast.Node, nodeName *ast.Node) []*ast.Node {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(nodeName)
			if symbol == nil {
				return nil
			}

			if symbol.Flags&ast.SymbolFlagsAlias != 0 {
				var found bool
				symbol, found = ctx.TypeChecker.ResolveAlias(symbol)
				if !found {
					return nil
				}
			}

			if symbol.Declarations == nil {
				return nil
			}

			declarations := slices.Clone(symbol.Declarations)

			nodeInTypeContext := isInTypeContext(node)
			slices.SortFunc(declarations, func(a *ast.Node, b *ast.Node) int {
				if !nodeInTypeContext {
					a, b = b, a
				}
				res := 0

				if isTypeContextDeclaration(a) {
					res -= 1
				}
				if isTypeContextDeclaration(b) {
					res += 1
				}

				return res
			})

			for _, decl := range declarations {
				if ast.IsTypeAliasDeclaration(decl) || ast.IsInterfaceDeclaration(decl) || ast.IsClassLike(decl) {
					return decl.TypeParameters()
				}

				if ast.IsVariableDeclaration(decl) {
					t := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol)
					signatures := utils.GetConstructSignatures(ctx.TypeChecker, t)
					if len(signatures) == 0 {
						continue
					}
					decl := checker.Signature_declaration(signatures[0])
					if decl != nil {
						return decl.TypeParameters()
					}
				}
			}

			return nil
		}

		getTypeParametersFromCall := func(node *ast.Node) []*ast.Node {
			signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
			if signature != nil {
				if declaration := checker.Signature_declaration(signature); declaration != nil {
					if typeParameters := declaration.TypeParameters(); len(typeParameters) != 0 {
						return typeParameters
					}
				}
			}
			if ast.IsNewExpression(node) {
				return getTypeParametersFromType(node, node.AsNewExpression().Expression)
			}
			return nil
		}

		checkArgsAndParameters := func(arguments *ast.NodeList, parameters []*ast.Node, callOrNewExpr *ast.Node) {
			if arguments == nil || parameters == nil || len(arguments.Nodes) == 0 || len(parameters) == 0 {
				return
			}

			// Just check the last one. Must specify previous type parameters if the last one is specified.
			lastParamIndex := len(arguments.Nodes) - 1

			if lastParamIndex >= len(parameters) {
				return
			}

			typeArgument := arguments.Nodes[lastParamIndex]
			typeParameter := parameters[lastParamIndex]

			typeArgumentType := ctx.TypeChecker.GetTypeAtLocation(typeArgument)

			if callOrNewExpr != nil {
				signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, callOrNewExpr, nil, checker.CheckModeNormal)
				for argumentIndex, argument := range callOrNewExpr.Arguments() {
					if signature == nil {
						break
					}

					parameters := checker.Signature_parameters(signature)
					if argumentIndex >= len(parameters) {
						continue
					}

					parameter := parameters[argumentIndex]
					if parameter == nil || parameter.ValueDeclaration == nil || !ast.IsParameter(parameter.ValueDeclaration) {
						continue
					}

					parameterDecl := parameter.ValueDeclaration.AsParameterDeclaration()
					if parameterDecl.Type == nil {
						continue
					}

					typeParameterType := ctx.TypeChecker.GetTypeAtLocation(typeParameter)
					parameterTypeFromDeclaration := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, parameterDecl.Type)
					if utils.IsTypeAnyType(parameterTypeFromDeclaration) || !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, typeParameterType, parameterTypeFromDeclaration) {
						continue
					}

					argumentType := checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(argument))
					if areTypesEquivalent(ctx.TypeChecker, typeArgumentType, argumentType) {
						ctx.ReportNodeWithFixes(typeArgument, buildCanBeInferredMessage(), func() []rule.RuleFix {
							var removeRange core.TextRange
							if lastParamIndex == 0 {
								removeRange = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arguments.End()).WithPos(arguments.Pos() - 1)
							} else {
								removeRange = typeArgument.Loc.WithPos(arguments.Nodes[lastParamIndex-1].End())
							}
							return []rule.RuleFix{rule.RuleFixRemoveRange(removeRange)}
						})
					}
				}
			}

			defaultType := typeParameter.AsTypeParameter().DefaultType
			if defaultType == nil {
				return
			}

			defaultTypeValue := ctx.TypeChecker.GetTypeAtLocation(defaultType)
			if !areTypesEquivalent(ctx.TypeChecker, defaultTypeValue, typeArgumentType) {
				return
			}

			ctx.ReportNodeWithFixes(typeArgument, buildIsDefaultParameterValueMessage(), func() []rule.RuleFix {
				var removeRange core.TextRange
				if lastParamIndex == 0 {
					removeRange = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arguments.End()).WithPos(arguments.Pos() - 1)
				} else {
					removeRange = typeArgument.Loc.WithPos(arguments.Nodes[lastParamIndex-1].End())
				}
				return []rule.RuleFix{rule.RuleFixRemoveRange(removeRange)}
			})
		}

		return rule.RuleListeners{
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				expr := node.AsExpressionWithTypeArguments()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromType(node, expr.Expression), nil)
			},
			ast.KindTypeReference: func(node *ast.Node) {
				expr := node.AsTypeReference()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromType(node, expr.TypeName), nil)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node), node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				expr := node.AsNewExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node), node)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				expr := node.AsTaggedTemplateExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node), nil)
			},
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				expr := node.AsJsxOpeningElement()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node), nil)
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				expr := node.AsJsxSelfClosingElement()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node), nil)
			},
		}
	},
}
