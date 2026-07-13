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

			// Only a symbol with merged declarations needs reordering so the
			// type-context declaration is consulted first. The common case is a
			// single declaration, where cloning and sorting would just allocate.
			declarations := symbol.Declarations
			if len(declarations) > 1 {
				declarations = slices.Clone(declarations)

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
			}

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

		// declarationHasDefaultedTypeParameter reports whether a function-like
		// declaration has at least one type parameter with a default.
		declarationHasDefaultedTypeParameter := func(decl *ast.Node) bool {
			for _, typeParameter := range decl.TypeParameters() {
				if typeParameter.AsTypeParameterDeclaration().DefaultType != nil {
					return true
				}
			}
			return false
		}

		// declarationCannotAcceptArgCount reports whether a function-like
		// declaration definitely cannot be the overload resolved for a call with
		// argCount value arguments. It only returns true when certain: the
		// declaration has fewer parameter slots than the call has arguments and
		// no rest parameter to absorb the extras. Conservative (returns false)
		// whenever argCount is unknown (<0) or a rest parameter is present.
		declarationCannotAcceptArgCount := func(decl *ast.Node, argCount int) bool {
			if argCount < 0 {
				return false
			}
			parameters := decl.Parameters()
			if len(parameters) >= argCount {
				return false
			}
			for _, parameter := range parameters {
				if parameter.AsParameterDeclaration().DotDotDotToken != nil {
					return false
				}
			}
			return true
		}

		// callCanReportUnnecessaryTypeArgument is a cheap gate before the
		// expensive signature resolution below. The rule can only report when the
		// resolved overload's type parameter (at the position of the last type
		// argument) has a default. Resolving the callee symbol and scanning its
		// overload declarations is far cheaper than full overload resolution, so
		// when no overload that could match this call's argument count carries a
		// defaulted type parameter, the resolution is skipped entirely. Any
		// uncertainty falls back to resolving (returns true) to preserve behavior.
		callCanReportUnnecessaryTypeArgument := func(node *ast.Node) bool {
			var callee *ast.Node
			argCount := -1
			switch node.Kind {
			case ast.KindCallExpression, ast.KindNewExpression:
				callee = node.Expression()
				argCount = len(node.Arguments())
			case ast.KindTaggedTemplateExpression:
				callee = node.AsTaggedTemplateExpression().Tag
			default:
				// JSX and anything else: argument count is ambiguous, so don't
				// attempt to gate; resolve as before.
				return true
			}
			if callee == nil {
				return true
			}

			symbol := ctx.TypeChecker.GetSymbolAtLocation(callee)
			if symbol == nil {
				return true
			}
			if symbol.Flags&ast.SymbolFlagsAlias != 0 {
				var found bool
				symbol, found = ctx.TypeChecker.ResolveAlias(symbol)
				if !found {
					return true
				}
			}
			if symbol.Declarations == nil {
				return true
			}

			for _, decl := range symbol.Declarations {
				// Only plain function/method overloads have a syntactic parameter
				// list we can reason about. For classes, variables holding
				// call/construct signatures, etc., fall back to resolving.
				if decl.FunctionLikeData() == nil {
					return true
				}
				if declarationHasDefaultedTypeParameter(decl) && !declarationCannotAcceptArgCount(decl, argCount) {
					return true
				}
			}
			return false
		}

		getTypeParametersFromCall := func(node *ast.Node) []*ast.Node {
			if !callCanReportUnnecessaryTypeArgument(node) {
				return nil
			}
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

		typeNodeReferencesTypeParameter := func(typeNode *ast.Node, typeParameterSymbols map[*ast.Symbol]struct{}) bool {
			var visit func(node *ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node == nil {
					return false
				}

				if ast.IsIdentifier(node) {
					if symbol := ctx.TypeChecker.GetSymbolAtLocation(node); symbol != nil {
						if _, ok := typeParameterSymbols[symbol]; ok {
							return true
						}
					}
				}

				return node.ForEachChild(func(child *ast.Node) bool {
					return visit(child)
				})
			}

			return visit(typeNode)
		}

		constructorArgumentsCanInferTypeParameters := func(node *ast.Node, parameters []*ast.Node) bool {
			if !ast.IsNewExpression(node) || len(node.Arguments()) == 0 {
				return false
			}

			typeParameterSymbols := make(map[*ast.Symbol]struct{}, len(parameters))
			for _, parameter := range parameters {
				name := parameter.Name()
				if name == nil {
					continue
				}
				if symbol := ctx.TypeChecker.GetSymbolAtLocation(name); symbol != nil {
					typeParameterSymbols[symbol] = struct{}{}
				}
			}
			if len(typeParameterSymbols) == 0 {
				return false
			}

			signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
			if signature == nil {
				return false
			}
			declaration := checker.Signature_declaration(signature)
			if declaration == nil || declaration.FunctionLikeData() == nil {
				return false
			}

			for _, parameter := range declaration.Parameters() {
				if typeNode := parameter.Type(); typeNode != nil && typeNodeReferencesTypeParameter(typeNode, typeParameterSymbols) {
					return true
				}
			}

			return false
		}

		checkArgsAndParameters := func(node *ast.Node, arguments *ast.NodeList, parameters []*ast.Node) {
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

			// Removing the entire type argument list from a constructor can re-enable
			// inference for all class type parameters used by constructor parameters.
			if lastParamIndex == 0 && constructorArgumentsCanInferTypeParameters(node, parameters) {
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

		// A node can only be reported when it has an explicit type argument list.
		// Most call/new/type-reference nodes have none, so gate the expensive
		// type-parameter queries (getResolvedSignature / symbol resolution) behind
		// this cheap syntactic check before they run.
		hasTypeArguments := func(typeArguments *ast.NodeList) bool {
			return typeArguments != nil && len(typeArguments.Nodes) != 0
		}

		return rule.RuleListeners{
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				expr := node.AsExpressionWithTypeArguments()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromType(node, expr.Expression))
			},
			ast.KindTypeReference: func(node *ast.Node) {
				expr := node.AsTypeReferenceNode()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromType(node, expr.TypeName))
			},

			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindNewExpression: func(node *ast.Node) {
				expr := node.AsNewExpression()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				expr := node.AsTaggedTemplateExpression()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				expr := node.AsJsxOpeningElement()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromCall(node))
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				expr := node.AsJsxSelfClosingElement()
				if !hasTypeArguments(expr.TypeArguments) {
					return
				}
				checkArgsAndParameters(node, expr.TypeArguments, getTypeParametersFromCall(node))
			},
		}
	},
}
