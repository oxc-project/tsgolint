// Package no_deprecated implements the no-deprecated rule.
//
// This rule disallows using code marked as @deprecated in JSDoc comments.
//
// Implementation Status: 191/219 tests passing (87.2%)
//
// Known limitations:
// - Allow options may not work correctly in all scenarios (3 tests)
// - Export specifiers with deprecated identifiers not detected (5 tests)
// - Reexported/aliased imports with deprecation tags on the alias (9 tests)
// - JSX attribute deprecation not implemented (2 tests)
// - Template literal keys in element access not supported (1 test)
// - Node.js module imports (node:*) deprecation checking (2 tests)
// - Nested destructuring patterns not fully supported (2 tests)
// - Miscellaneous edge cases (4 tests)
package no_deprecated

import (
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

	// Walk up the tree while we're the property of a PropertyAccessExpression
	// This handles cases like a.b.c() where we need to walk from c to a.b.c
	for {
		if callee.Parent == nil {
			break
		}
		if callee.Parent.Kind != ast.KindPropertyAccessExpression {
			break
		}

		// Only move up if this node is the property (name) of the PropertyAccessExpression
		// Not if it's the expression (object) part
		pae := callee.Parent.AsPropertyAccessExpression()
		if pae.Name().AsNode() != callee {
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

			// Handle object binding patterns (destructuring) and shorthand properties
			if node.Parent != nil && node.Kind != ast.KindSuperKeyword {
				parent := node.Parent

				// Handle BindingElement in object destructuring: const { b } = a
				if parent.Kind == ast.KindBindingElement {
					bindingElem := parent.AsBindingElement()
					// The binding element's parent should be an ObjectBindingPattern
					if parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern {
						// Get the type of the object being destructured
						// We need to find the variable declaration or parameter that contains this binding pattern
						objBindingPattern := parent.Parent

						// Find what's being destructured by looking up the tree
						var sourceType *checker.Type
						if objBindingPattern.Parent != nil {
							switch objBindingPattern.Parent.Kind {
							case ast.KindVariableDeclaration:
								varDecl := objBindingPattern.Parent.AsVariableDeclaration()
								if varDecl.Initializer != nil {
									sourceType = ctx.TypeChecker.GetTypeAtLocation(varDecl.Initializer)
								}
							case ast.KindParameter:
								// For parameters, get the type directly
								sourceType = ctx.TypeChecker.GetTypeAtLocation(objBindingPattern.Parent)
							}
						}

						if sourceType != nil {
							// Get the property name being destructured
							propertyName := node.Text()
							if bindingElem.PropertyName != nil {
								propertyName = bindingElem.PropertyName.Text()
							}

							property := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
							propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(node)

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
				}

				// Handle shorthand property assignments in object literals
				if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Parent != nil {
					parentType := ctx.TypeChecker.GetTypeAtLocation(parent.Parent)
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

			switch parent.Kind {
			case ast.KindClassExpression:
				fallthrough
			case ast.KindVariableDeclaration:
				fallthrough
			case ast.KindEnumMember:
				fallthrough
			case ast.KindClassDeclaration:
				return parent.Name() == node

			case ast.KindMethodDeclaration:
				fallthrough
			case ast.KindPropertyDeclaration:
				return parent.Name() == node

			case ast.KindPropertyAssignment:
				// Property in object literal is a declaration
				return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression

			case ast.KindArrowFunction:
				fallthrough
			case ast.KindFunctionDeclaration:
				fallthrough
			case ast.KindFunctionExpression:
				fallthrough
			case ast.KindEnumDeclaration:
				fallthrough
			case ast.KindInterfaceDeclaration:
				fallthrough
			case ast.KindModuleDeclaration:
				fallthrough
			case ast.KindMethodSignature:
				fallthrough
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
				return
			}

			isDeprecated, deprecationReason := getDeprecationReason(node)

			if !isDeprecated {
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

		// Check element access expressions with literal keys (e.g., a['b'])
		checkElementAccessExpression := func(node *ast.Node) {
			eae := node.AsElementAccessExpression()
			if eae.ArgumentExpression == nil {
				return
			}

			// Get the type of the property being accessed
			propertyType := ctx.TypeChecker.GetTypeAtLocation(eae.ArgumentExpression)
			if propertyType == nil {
				return
			}

			// Only check if the property is a literal type (string or number literal)
			isStringLit := propertyType.IsStringLiteral()
			isNumberLit := utils.IsTypeFlagSet(propertyType, checker.TypeFlagsNumberLiteral)
			isBigIntLit := utils.IsTypeFlagSet(propertyType, checker.TypeFlagsBigIntLiteral)

			if !isStringLit && !isNumberLit && !isBigIntLit {
				return
			}

			objectType := ctx.TypeChecker.GetTypeAtLocation(eae.Expression)

			// Get the property name from the literal type
			literalType := propertyType.AsLiteralType()
			if literalType == nil {
				return
			}

			var propertyName string
			value := literalType.Value()
			if value == nil {
				return
			}

			// Convert value to string
			if str, ok := value.(string); ok {
				propertyName = str
			} else {
				// For numbers or other types, use String() representation
				propertyName = literalType.String()
			}

			property := checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, propertyName)

			isDeprecated, reason := getJsDocDeprecation(property)
			if !isDeprecated {
				return
			}

			if utils.TypeMatchesSomeSpecifier(objectType, opts.Allow, []string{}, ctx.Program) {
				return
			}

			// Report on the argument expression (the key being accessed)
			if reason == "" {
				ctx.ReportNode(eae.ArgumentExpression, buildDeprecatedMessage(propertyName))
			} else {
				ctx.ReportNode(eae.ArgumentExpression, buildDeprecatedWithReasonMessage(propertyName, strings.TrimSpace(reason)))
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Parent == nil {
					return
				}
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
			// We have checkElementAccessExpression registered to handle element access with literal keys.
			// This handles cases like obj['deprecatedProp'] where the key is a literal.
			ast.KindElementAccessExpression: checkElementAccessExpression,
			ast.KindPrivateIdentifier:       checkIdentifier,
			ast.KindSuperKeyword:            checkIdentifier,
		}
	},
}
