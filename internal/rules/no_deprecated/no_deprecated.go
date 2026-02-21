package no_deprecated

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

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

func getReportedNodeName(node *ast.Node) string {
	if node.Kind == ast.KindSuperKeyword {
		return "super"
	}
	if node.Kind == ast.KindPrivateIdentifier {
		return "#" + node.Text()
	}
	return node.Text()
}

var NoDeprecatedRule = rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoDeprecatedOptions](options, "no-deprecated")

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

		getJsDocDeprecation := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil {
				return false, ""
			}

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

				if symbol.Flags&ast.SymbolFlagsAlias == 0 {
					break
				}

				if checker.Checker_getDeclarationOfAliasSymbol(ctx.TypeChecker, symbol) == nil {
					break
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
				// Fallback: if the signature had no declaration (typescript-go limitation for
				// inherited methods), check the symbol's own declarations directly.
				// Guarded by signatureDecl == nil so this never fires for overloaded functions
				// (which always produce a resolved signatureDecl), preventing false positives on
				// non-deprecated overloads.
				if signatureDecl == nil {
					isDeprecated, reason = getJsDocDeprecation(aliasedSymbol)
					if isDeprecated {
						return true, reason
					}
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
			switch elementNode.Kind {
			case ast.KindJsxSelfClosingElement:
				tagName = elementNode.AsJsxSelfClosingElement().TagName
			case ast.KindJsxOpeningElement:
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

		// Helper to get the source type for a binding pattern by walking up the tree
		// This is declared as a variable to allow recursive calls
		var getBindingPatternSourceType func(bindingPattern *ast.Node) *checker.Type
		getBindingPatternSourceType = func(bindingPattern *ast.Node) *checker.Type {
			current := bindingPattern

			for current != nil {
				switch current.Kind {
				case ast.KindVariableDeclaration:
					varDecl := current.AsVariableDeclaration()
					if varDecl.Initializer != nil {
						return ctx.TypeChecker.GetTypeAtLocation(varDecl.Initializer)
					}
					return nil

				case ast.KindParameter:
					return ctx.TypeChecker.GetTypeAtLocation(current)

				case ast.KindBindingElement:
					// For nested destructuring like { bar: { anchor } }
					// We need to get the type of the parent property
					bindingElem := current.AsBindingElement()

					// Get the parent binding pattern
					if current.Parent != nil {
						parentPattern := current.Parent

						// Get the source type of the parent pattern
						parentSourceType := getBindingPatternSourceType(parentPattern)
						if parentSourceType == nil {
							return nil
						}

						// Get the property name for this binding element
						propertyName := ""
						if bindingElem.PropertyName != nil {
							if bindingElem.PropertyName.Kind == ast.KindComputedPropertyName {
								return nil
							}
							propertyName = bindingElem.PropertyName.Text()
						} else if bindingElem.Name() != nil {
							name := bindingElem.Name()
							// The name might be a binding pattern for nested destructuring
							// For example: const [{ anchor }] = x where x is a tuple
							if name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern {
								// Special case: when destructuring from an array/tuple
								// e.g., const [{ anchor }] = x where x: [item: Props]
								// We can't get a property name, so we need to use the type checker differently
								// Return nil for now - the caller will need to use TypeChecker.GetTypeAtLocation
								return nil
							} else {
								propertyName = name.Text()
							}
						}

						if propertyName == "" {
							return nil
						}

						// Get the type of this property
						property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentSourceType, propertyName)
						if property != nil {
							return ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, current)
						}
					}
					return nil

				case ast.KindArrayBindingPattern:
					// For array destructuring, get the source type
					parentSourceType := getBindingPatternSourceType(current.Parent)
					if parentSourceType == nil {
						return nil
					}
					// For arrays/tuples, try to get the first element type (index 0)
					// Try getting property "0" for tuple types
					property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentSourceType, "0")
					if property != nil {
						return ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, current)
					}
					return parentSourceType

				case ast.KindObjectBindingPattern:
					// Continue walking up
					current = current.Parent
					continue
				}

				current = current.Parent
			}

			return nil
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
					// The binding element's parent should be an ObjectBindingPattern or ArrayBindingPattern
					if parent.Parent != nil && (parent.Parent.Kind == ast.KindObjectBindingPattern || parent.Parent.Kind == ast.KindArrayBindingPattern) {
						// Get the type of the object/array being destructured
						bindingPattern := parent.Parent
						sourceType := getBindingPatternSourceType(bindingPattern)

						// If getBindingPatternSourceType returns nil (e.g., for nested destructuring),
						// try using TypeChecker.GetTypeAtLocation as a fallback
						if sourceType == nil {
							sourceType = ctx.TypeChecker.GetTypeAtLocation(bindingPattern)
						}

						if sourceType != nil {
							// Get the property name being destructured
							propertyName := node.Text()
							if bindingElem.PropertyName != nil {
								if bindingElem.PropertyName.Kind == ast.KindComputedPropertyName {
									return false, ""
								}
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
				fallthrough
			case ast.KindGetAccessor:
				fallthrough
			case ast.KindSetAccessor:
				return parent.Name() == node

			case ast.KindPropertyAssignment:
				// Property keys in object literals are declarations
				// But property values are uses (not declarations)
				propAssign := parent.AsPropertyAssignment()
				// Check if node is the value (initializer)
				if propAssign.Initializer == node {
					// This is the value side (e.g., test in { prop: test })
					return false
				}
				// This is the key side - it's a declaration if parent is ObjectLiteralExpression
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

			if utils.TypeMatchesSomeSpecifier(ty, opts.Allow, ctx.Program) ||
				utils.ValueMatchesSomeSpecifier(node, opts.Allow, ctx.Program, ty) {
				return
			}

			name := getReportedNodeName(node)
			if deprecationReason == "" {
				ctx.ReportNode(node, buildDeprecatedMessage(name))
			} else {
				ctx.ReportNode(node, buildDeprecatedWithReasonMessage(name, strings.TrimSpace(deprecationReason)))
			}
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

			if utils.TypeMatchesSomeSpecifier(objectType, opts.Allow, ctx.Program) {
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

				// Skip JSX closing elements to avoid duplicate reports
				if node.Parent.Kind == ast.KindJsxClosingElement {
					return
				}

				// Skip identifiers directly in export declarations (not in specifiers)
				if node.Parent.Kind == ast.KindExportDeclaration {
					return
				}

				// Skip namespace exports like: export * as ns from 'module'
				// The identifier 'ns' is the export name, not a usage
				if node.Parent.Kind == ast.KindNamespaceExport {
					return
				}

				// Handle export specifiers: export { foo } or export { foo as bar }
				if node.Parent.Kind == ast.KindExportSpecifier {
					exportSpec := node.Parent.AsExportSpecifier()

					// In "export { foo as bar }":
					//   - PropertyName points to "foo" (the local symbol being exported)
					//   - Name() returns "bar" (the exported name)
					// In "export { foo }":
					//   - PropertyName is nil
					//   - Name() returns "foo"

					// Check which identifier we're looking at
					isPropertyName := exportSpec.PropertyName != nil && exportSpec.PropertyName.AsNode() == node

					if isPropertyName {
						// This is the local binding (foo in "export { foo as bar }")
						// We should NOT report on the local name, only on the export
						return
					}

					// This is the exported identifier (the alias)
					// Check if the export specifier itself has a @deprecated tag
					// If so, we should NOT report (the re-export is explicitly marked as deprecated)
					jsdocs := node.Parent.JSDoc(nil)
					hasDeprecatedTag := false
					for _, jsdoc := range jsdocs {
						tags := jsdoc.AsJSDoc().Tags
						if tags == nil {
							continue
						}
						for _, tag := range tags.Nodes {
							if ast.IsJSDocDeprecatedTag(tag) {
								hasDeprecatedTag = true
								break
							}
						}
						if hasDeprecatedTag {
							break
						}
					}

					if hasDeprecatedTag {
						// The export specifier itself is marked deprecated
						// Don't report - this is intentional documentation of the deprecation
						return
					}

					// Fall through to check if the underlying symbol is deprecated
				}

				checkIdentifier(node)
			},

			// TODO: TypeScript implementation listens to MemberExpression for computed property access.
			// We have checkElementAccessExpression registered to handle element access with literal keys.
			// This handles cases like obj['deprecatedProp'] where the key is a literal.
			ast.KindElementAccessExpression: checkElementAccessExpression,
			ast.KindPrivateIdentifier:       checkIdentifier,
			ast.KindSuperKeyword:            checkIdentifier,
		}
	},
}
