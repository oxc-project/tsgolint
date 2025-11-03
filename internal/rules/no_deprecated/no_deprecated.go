package no_deprecated

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type NoDeprecatedOptions struct {
	Allow []utils.TypeOrValueSpecifier `json:"allow"`
}

func buildDeprecatedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "deprecated",
		Description: "`" + name + "` is deprecated.",
	}
}

func buildDeprecatedWithReasonMessage(name string, reason string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "deprecatedWithReason",
		Description: "`" + name + "` is deprecated. " + reason,
	}
}

func isNodeCalleeOfParent(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindNewExpression:
		newExpr := node.Parent.AsNewExpression()
		return newExpr.Expression == node
	case ast.KindCallExpression:
		callExpr := node.Parent.AsCallExpression()
		return callExpr.Expression == node
	case ast.KindTaggedTemplateExpression:
		taggedTemplate := node.Parent.AsTaggedTemplateExpression()
		return taggedTemplate.Tag == node
	case ast.KindJsxOpeningElement:
		jsxOpening := node.Parent.AsJsxOpeningElement()
		return jsxOpening.TagName == node
	default:
		return false
	}
}

func getCallLikeNode(node *ast.Node) *ast.Node {
	callee := node

	for {
		if callee.Parent == nil {
			break
		}
		if callee.Parent.Kind != ast.KindPropertyAccessExpression {
			break
		}
		callee = callee.Parent
	}

	if isNodeCalleeOfParent(callee) {
		return callee
	}
	return nil
}

// Helper to get the reported name for a node
func getReportedNodeName(node *ast.Node) string {
	if node.Kind == ast.KindSuperKeyword {
		return "super"
	}
	if node.Kind == ast.KindPrivateIdentifier {
		return "#" + node.Text()
	}
	// For most identifiers, use Text()
	return node.Text()
}

var NoDeprecatedRule = rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoDeprecatedOptions)
		if !ok {
			opts = NoDeprecatedOptions{}
		}
		if opts.Allow == nil {
			opts.Allow = []utils.TypeOrValueSpecifier{}
		}

		// Helper to extract deprecation reason from a JSDoc deprecated tag
		getJsDocDeprecationFromNode := func(node *ast.Node) string {
			if node == nil {
				return ""
			}

			jsdocs := node.JSDoc(nil)
			for _, jsdoc := range jsdocs {
				tags := jsdoc.AsJSDoc().Tags
				if tags == nil {
					continue
				}
				for _, tag := range tags.Nodes {
					if ast.IsJSDocDeprecatedTag(tag) {
						deprecatedTag := tag.AsJSDocDeprecatedTag()
						if deprecatedTag.Comment != nil && len(deprecatedTag.Comment.Nodes) > 0 {
							var text strings.Builder
							for _, commentNode := range deprecatedTag.Comment.Nodes {
								text.WriteString(commentNode.Text())
							}
							return text.String()
						}
						return ""
					}
				}
			}
			return ""
		}

		// Helper to check if a symbol or its declarations are deprecated
		getJsDocDeprecation := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil {
				return false, ""
			}

			// TODO: TypeScript implementation uses symbol.getJsDocTags(checker) which includes
			// JSDoc tags from all symbol declarations combined. We currently check each declaration
			// individually using the Checker_IsDeprecatedDeclaration helper instead.
			// This may miss some edge cases where JSDoc tags are inherited differently.

			// Check all declarations for @deprecated tag
			for _, decl := range symbol.Declarations {
				if checker.Checker_IsDeprecatedDeclaration(ctx.TypeChecker, decl) {
					reason := getJsDocDeprecationFromNode(decl)
					return true, reason
				}
			}

			return false, ""
		}

		searchForDeprecationInAliasesChain := func(
			symbol *ast.Symbol,
			checkDeprecationsOfAliasedSymbol bool,
		) (bool, string) {
			if symbol == nil {
				return false, ""
			}

			if !utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				if checkDeprecationsOfAliasedSymbol {
					return getJsDocDeprecation(symbol)
				}
				return false, ""
			}

			targetSymbol := ctx.TypeChecker.GetAliasedSymbol(symbol)

			for utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				isDeprecated, reason := getJsDocDeprecation(symbol)
				if isDeprecated {
					return true, reason
				}

				immediateAliasedSymbol := checker.Checker_getImmediateAliasedSymbol(ctx.TypeChecker, symbol)

				if immediateAliasedSymbol == nil {
					break
				}

				symbol = immediateAliasedSymbol

				if checkDeprecationsOfAliasedSymbol && symbol == targetSymbol {
					return getJsDocDeprecation(symbol)
				}
			}

			return false, ""
		}

		// Helper to get deprecation for call-like expressions (function calls, new expressions, etc.)
		getCallLikeDeprecation := func(node *ast.Node) (bool, string) {
			if node == nil || node.Parent == nil {
				return false, ""
			}

			tsNode := node.Parent
			// Get the resolved signature for the call
			signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, tsNode, nil, 0)
			if signature == nil {
				return false, ""
			}

			// Check if the signature itself is deprecated
			// TODO: TypeScript implementation also calls signature.getJsDocTags() to get JSDoc deprecation
			// on the signature object itself. We rely on checking the signature's declaration node instead.
			signatureDecl := signature.Declaration()
			if signatureDecl != nil {
				if checker.Checker_IsDeprecatedDeclaration(ctx.TypeChecker, signatureDecl) {
					reason := getJsDocDeprecationFromNode(signatureDecl)
					return true, reason
				}
			}

			// Also check the symbol
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil {
				return false, ""
			}

			aliasedSymbol := symbol
			if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				aliasedSymbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
			}

			// For property-like signatures, check the symbol itself first
			var symbolDeclarationKind ast.Kind
			if aliasedSymbol != nil && len(aliasedSymbol.Declarations) > 0 {
				symbolDeclarationKind = aliasedSymbol.Declarations[0].Kind
			}

			// Properties with function-like types have @deprecated on their symbols, not signatures
			if symbolDeclarationKind != ast.KindMethodDeclaration &&
				symbolDeclarationKind != ast.KindFunctionDeclaration &&
				symbolDeclarationKind != ast.KindMethodSignature {
				isDeprecated, reason := searchForDeprecationInAliasesChain(symbol, true)
				if isDeprecated {
					return true, reason
				}
			} else {
				// For function/method declarations, don't check the aliased symbol
				// but rely on the signature deprecation (checked above)
				isDeprecated, reason := searchForDeprecationInAliasesChain(symbol, false)
				if isDeprecated {
					return true, reason
				}
			}

			return false, ""
		}

		// Helper to get deprecation for JSX attributes
		getJSXAttributeDeprecation := func(elementNode *ast.Node, propertyName string) (bool, string) {
			if elementNode == nil {
				return false, ""
			}

			var tagName *ast.Node
			// Handle both JsxSelfClosingElement and JsxOpeningElement
			if elementNode.Kind == ast.KindJsxSelfClosingElement {
				tagName = elementNode.AsJsxSelfClosingElement().TagName
			} else if elementNode.Kind == ast.KindJsxOpeningElement {
				tagName = elementNode.AsJsxOpeningElement().TagName
			}

			if tagName == nil {
				return false, ""
			}

			// Get the contextual type for the JSX element
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, tagName, 0)
			if contextualType == nil {
				return false, ""
			}

			// Get the property symbol
			symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, contextualType, propertyName)

			return getJsDocDeprecation(symbol)
		}

		// Extract the deprecation reason from JSDoc comments
		getDeprecationReason := func(node *ast.Node) (bool, string) {
			callLikeNode := getCallLikeNode(node)
			if callLikeNode != nil {
				return getCallLikeDeprecation(callLikeNode)
			}

			if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute && node.Kind != ast.KindSuperKeyword {
				// node.Parent is JsxAttribute, node.Parent.Parent is JsxAttributes, node.Parent.Parent.Parent is the element
				if node.Parent.Parent != nil && node.Parent.Parent.Parent != nil {
					return getJSXAttributeDeprecation(node.Parent.Parent.Parent, node.Text())
				}
			}

			// log parent kind
			fmt.Println("Node parent kind:", node.Parent.Kind)

			// Handle property assignments in object literals
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Kind != ast.KindSuperKeyword && node.Parent.Parent != nil {
				parentType := ctx.TypeChecker.GetTypeAtLocation(node.Parent.Parent)
				if parentType != nil {
					propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(node)
					property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentType, node.Text())

					// Check alias chain first
					isDeprecated, reason := searchForDeprecationInAliasesChain(propertySymbol, true)
					if isDeprecated {
						return true, reason
					}

					// Check the property on the type
					isDeprecated, reason = getJsDocDeprecation(property)
					if isDeprecated {
						return true, reason
					}

					// Check the property symbol itself
					isDeprecated, reason = getJsDocDeprecation(propertySymbol)
					if isDeprecated {
						return true, reason
					}

					// Check shorthand assignment value symbol
					if propertySymbol != nil && propertySymbol.ValueDeclaration != nil {
						valueSymbol := checker.Checker_GetShorthandAssignmentValueSymbol(ctx.TypeChecker, propertySymbol.ValueDeclaration)
						isDeprecated, reason = getJsDocDeprecation(valueSymbol)
						if isDeprecated {
							return true, reason
						}
					}
				}
			}

			return searchForDeprecationInAliasesChain(
				ctx.TypeChecker.GetSymbolAtLocation(node),
				true,
			)
		}

		// Check if a node is a declaration (should not report on declarations)
		// TODO: TypeScript implementation handles more complex cases like:
		// - ArrayPattern elements
		// - ClassExpression
		// - TSEnumMember
		// - MethodDefinition/PropertyDefinition/AccessorProperty key checks
		// - Property shorthand and value checks with ObjectPattern
		// - AssignmentPattern left side
		// - More function-like and type declaration kinds
		// We handle the most common cases but may miss some edge cases.
		isDeclaration := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			// debug print parent kind
			fmt.Println("Parent kind:", parent.Kind)

			switch parent.Kind {
    		// case AST_NODE_TYPES.ArrayPattern:
    		// return parent.elements.includes(node as TSESTree.Identifier);
			case ast.KindClassExpression:
				fallthrough
			case ast.KindVariableDeclaration:
				fallthrough
			case ast.KindEnumMember:
				fallthrough
			case ast.KindClassDeclaration:
				// log parent.Name()
				fmt.Println("Parent name:", parent.Name())
				return parent.Name() == node

			case ast.KindMethodDeclaration:
				fallthrough
			// case ast.KindAccessorProperty:
			case ast.KindPropertyDeclaration:
				return parent.Name() == node
			case ast.KindPropertyAssignment:
				// propertyAssignment := parent.AsPropertyAssignment()
				// foo in "const { foo } = bar" will be processed twice, as parent.key
				// and parent.value. The second is treated as a declaration.

				// fmt.Println("Property access expression value:", propertyAccessExpression)

				// if propertyAccessExpression.Shorthand && propertyAccessExpression.Value == node {
				// 	return parent.Parent.Kind == ast.KindObjectBindingPattern
				// }
				// if propertyAccessExpression.Value == node {
				// 	return false
				// }
				return parent.Parent.Kind == ast.KindObjectLiteralExpression
			case ast.KindArrowFunction:
				fallthrough
			case ast.KindFunctionDeclaration:
				fallthrough
			case ast.KindFunctionExpression:
				fallthrough
			// case ast.TSDeclareFunction:
			// case ast.TSEmptyBodyFunctionExpression:
			case ast.KindEnumDeclaration:
				fallthrough
			// case ast.TSInterfaceDeclaration:
			// case ast.TSMethodSignature:
			// case ast.TSModuleDeclaration:
			// case ast.TSParameterProperty:
			case ast.KindPropertySignature:
				fallthrough
			case ast.KindTypeAliasDeclaration:
				fallthrough
			case ast.KindTypeParameter:
				fallthrough
			case ast.KindParameter:
				return true
			case ast.KindImportEqualsDeclaration:
				return parent.Name() == node
			default:
				return false
			}
		}

		// Check if we're inside an import statement
		// TODO: TypeScript implementation checks more boundary node types:
		// - ArrowFunctionExpression
		// - ExportAllDeclaration
		// - ExportNamedDeclaration
		// - TSInterfaceDeclaration
		// - FunctionExpression
		// - Program
		// - TSUnionType
		// - VariableDeclarator
		// We check the most common boundaries but may not catch all cases.
		isInsideImport := func(node *ast.Node) bool {
			current := node
			for current != nil {
				kind := current.Kind
				if kind == ast.KindImportDeclaration {
					return true
				}
				// Stop at certain boundaries
				if kind == ast.KindSourceFile ||
					kind == ast.KindBlock ||
					kind == ast.KindFunctionDeclaration ||
					kind == ast.KindClassDeclaration {
					return false
				}
				current = current.Parent
			}
			return false
		}

		checkIdentifier := func(node *ast.Node) {
			if isDeclaration(node) || isInsideImport(node) {
				fmt.Println("Skipping declaration or import")
				return
			}

			isDeprecated, deprecationReason := getDeprecationReason(node)

			if !isDeprecated {
				fmt.Println("Not deprecated")
				return
			}

			ty := ctx.TypeChecker.GetTypeAtLocation(node)

			// TODO: if type OR value is allowed, skip

			if utils.TypeMatchesSomeSpecifier(ty, opts.Allow, []string{}, ctx.Program) ||
				utils.ValueMatchesSomeSpecifier(node, opts.Allow, ctx.Program, ty) {
				return
			}

			name := getReportedNodeName(node)
			if deprecationReason == "" {
				ctx.ReportNode(node, buildDeprecatedMessage(name))
			} else {
				ctx.ReportNode(node, buildDeprecatedWithReasonMessage(name, strings.TrimSpace(deprecationReason)))
			}

			return
		}

		checkPropertyAccessExpression := func(node *ast.Node) {
			fmt.Println("Checking property access expression")
			pae := node.AsPropertyAccessExpression()
			propertyType := ctx.TypeChecker.GetTypeAtLocation(pae.Name().AsNode())

			if propertyType == nil {
				fmt.Println("Property type is nil")
				return
			}

			fmt.Println("Property type flags:", propertyType.Flags())

			if !utils.IsTypeFlagSet(propertyType, checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBigIntLiteral) {
				return
			}

			fmt.Println("Property is a literal type")

			objectType := ctx.TypeChecker.GetTypeAtLocation(pae.Expression)


			propertyName := propertyType.AsLiteralType().String()

			fmt.Println("Checking property:", propertyName)

			property := checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, propertyName)

			fmt.Println("Got property symbol:", property)

			isDeprecated, reason := getJsDocDeprecation(property)

			if !isDeprecated {
				return
			}

			if utils.TypeMatchesSomeSpecifier(objectType, opts.Allow, []string{}, ctx.Program) {
				return
			}

			if reason == "" {
				ctx.ReportNode(node, buildDeprecatedMessage(pae.Name().Text()))
			} else {
				ctx.ReportNode(node, buildDeprecatedWithReasonMessage(pae.Name().Text(), strings.TrimSpace(reason)))
			}


		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				fmt.Println("Visiting identifier:", node.Text())
				if node.Parent == nil {
					return
				}
				// print parent kind
				fmt.Println("Node parent kind:", node.Parent.Kind)
				if node.Parent.Kind == ast.KindExportDeclaration {
					return
				}
				if node.Parent.Kind == ast.KindExportSpecifier {
					// TODO: TypeScript implementation checks if this is the exported alias vs local binding.
					// It only checks the exported side (parent.exported === node), and if the alias itself
					// has a deprecation tag, it returns early. We currently don't distinguish between
					// the local and exported identifiers in export specifiers, which may cause issues
					// with export { /** @deprecated */ foo as bar } scenarios.
					return
				}

				checkIdentifier(node)
			},
			// TODO: TypeScript implementation has a JSXIdentifier listener separate from Identifier.
			// In TypeScript AST, JSX identifiers may be represented differently than in ESTree.
			// We currently handle JSX through the Identifier listener and parent kind checks.
			// This works for most cases but may miss some JSX-specific scenarios.

			// TODO: TypeScript implementation listens to MemberExpression for computed property access.
			// We have checkPropertyAccessExpression registered but it's not fully implemented.
			// This causes test failures for cases like obj['deprecatedProp'] where the key is a literal.
			ast.KindPropertyAccessExpression: checkPropertyAccessExpression,
			ast.KindPrivateIdentifier:        checkIdentifier,
			ast.KindSuperKeyword:             checkIdentifier,
		}
	},
}
