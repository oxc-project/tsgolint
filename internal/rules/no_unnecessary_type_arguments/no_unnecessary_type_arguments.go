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

func buildUnnecessaryTypeParameterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryTypeParameter",
		Description: "This is the default value for this type parameter, so it can be omitted.",
	}
}

func isTypeContextDeclaration(decl *ast.Node) bool {
	return ast.IsTypeAliasDeclaration(decl) || ast.IsInterfaceDeclaration(decl)
}

func isInTypeContext(node *ast.Node) bool {
	return ast.IsTypeReferenceNode(node) || ast.IsInterfaceDeclaration(node.Parent) || ast.IsTypeReferenceNode(node.Parent) || (ast.IsHeritageClause(node.Parent) && node.Parent.AsHeritageClause().Token == ast.KindImplementsKeyword)
}

type typeForComparison struct {
	typeValue     *checker.Type
	typeArguments []*checker.Type
}

func getTypeForComparison(typeChecker *checker.Checker, t *checker.Type) typeForComparison {
	if checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		return typeForComparison{
			typeValue:     t.Target(),
			typeArguments: checker.Checker_getTypeArguments(typeChecker, t),
		}
	}

	return typeForComparison{
		typeValue: t,
	}
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

		checkArgsAndParameters := func(arguments *ast.NodeList, parameters []*ast.Node) {
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

			defaultTypeNode := typeParameter.AsTypeParameterDeclaration().DefaultType
			if defaultTypeNode == nil {
				return
			}

			defaultType := ctx.TypeChecker.GetTypeAtLocation(defaultTypeNode)
			argType := ctx.TypeChecker.GetTypeAtLocation(typeArgument)

			if defaultType == nil || argType == nil {
				return
			}

			typesMatch := defaultType == argType
			if !typesMatch {
				// For more complex types (such as generic object types), TS won't always create a
				// global shared type object for the type, so fall back to comparing the
				// reference type and the passed type arguments.
				defaultTypeResolved := getTypeForComparison(ctx.TypeChecker, defaultType)
				argTypeResolved := getTypeForComparison(ctx.TypeChecker, argType)
				typesMatch = defaultTypeResolved.typeValue == argTypeResolved.typeValue &&
					len(defaultTypeResolved.typeArguments) == len(argTypeResolved.typeArguments)

				if typesMatch {
					for i, defaultTypeArgument := range defaultTypeResolved.typeArguments {
						if defaultTypeArgument != argTypeResolved.typeArguments[i] {
							typesMatch = false
							break
						}
					}
				}
			}

			if !typesMatch {
				return
			}

			ctx.ReportNodeWithFixes(typeArgument, buildUnnecessaryTypeParameterMessage(), func() []rule.RuleFix {
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
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromType(node, expr.Expression))
			},
			ast.KindTypeReference: func(node *ast.Node) {
				expr := node.AsTypeReferenceNode()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromType(node, expr.TypeName))
			},

			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindNewExpression: func(node *ast.Node) {
				expr := node.AsNewExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				expr := node.AsTaggedTemplateExpression()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				expr := node.AsJsxOpeningElement()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				expr := node.AsJsxSelfClosingElement()
				checkArgsAndParameters(expr.TypeArguments, getTypeParametersFromCall(node))
			},
		}
	},
}
