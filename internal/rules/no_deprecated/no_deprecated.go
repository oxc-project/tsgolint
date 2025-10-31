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

var NoDeprecatedRule = rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Extract the deprecation reason from JSDoc comments using node.JSDoc()
		// This uses the built-in JSDoc parsing from typescript-go
		getDeprecationReason := func(node *ast.Node) string {
			if node == nil {
				return ""
			}

			// Collect nodes to check: node and its ancestors up to certain boundaries
			// JSDoc comments can be on various parent nodes depending on the declaration type
			nodesToCheck := []*ast.Node{}
			current := node
			for current != nil {
				nodesToCheck = append(nodesToCheck, current)

				// Stop at certain boundaries
				kind := current.Kind
				if kind == ast.KindSourceFile || kind == ast.KindBlock {
					break
				}
				// For variable declarations, check up to the statement
				if kind == ast.KindVariableStatement {
					break
				}
				// For other declarations, stop at the declaration itself
				if ast.IsDeclaration(current) && current != node {
					break
				}

				current = current.Parent
			}

			// Check each node for JSDoc comments
			for _, checkNode := range nodesToCheck {
				if checkNode == nil {
					continue
				}

				// Get JSDoc comments using the built-in method
				jsdocs := checkNode.JSDoc(ctx.SourceFile)
				for _, jsdoc := range jsdocs {
					jsDocNode := jsdoc.AsJSDoc()
					if jsDocNode.Tags == nil {
						continue
					}

					// Look for @deprecated tag
					for _, tag := range jsDocNode.Tags.Nodes {
						if !ast.IsJSDocDeprecatedTag(tag) {
							continue
						}

						// Found @deprecated tag, extract the reason
						depTag := tag.AsJSDocDeprecatedTag()
						if depTag.Comment == nil || len(depTag.Comment.Nodes) == 0 {
							// No reason provided
							return ""
						}

						// Extract text from comment nodes
						var reasonParts []string
						for _, commentNode := range depTag.Comment.Nodes {
							text := commentNode.Text()
							if text != "" {
								reasonParts = append(reasonParts, strings.TrimSpace(text))
							}
						}

						if len(reasonParts) > 0 {
							return strings.Join(reasonParts, " ")
						}
						return ""
					}
				}
			}

			return ""
		}

		// Check if a node has a @deprecated JSDoc tag
		hasDeprecatedTag := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			// Collect nodes to check: node and its ancestors
			nodesToCheck := []*ast.Node{}
			current := node
			for current != nil {
				nodesToCheck = append(nodesToCheck, current)

				// Stop at certain boundaries
				kind := current.Kind
				if kind == ast.KindSourceFile || kind == ast.KindBlock {
					break
				}
				if kind == ast.KindVariableStatement {
					break
				}
				if ast.IsDeclaration(current) && current != node {
					break
				}

				current = current.Parent
			}

			// Check each node for JSDoc comments
			for _, checkNode := range nodesToCheck {
				if checkNode == nil {
					continue
				}

				// Get JSDoc comments using the built-in method
				jsdocs := checkNode.JSDoc(ctx.SourceFile)
				for _, jsdoc := range jsdocs {
					jsDocNode := jsdoc.AsJSDoc()
					if jsDocNode.Tags == nil {
						continue
					}

					// Look for @deprecated tag
					for _, tag := range jsDocNode.Tags.Nodes {
						if ast.IsJSDocDeprecatedTag(tag) {
							return true
						}
					}
				}
			}
			return false
		}

		// Check if a symbol is deprecated and optionally get the deprecation reason
		// This approach parses JSDoc comments directly following Parser.withJSDoc
		// Returns: (isDeprecated bool, reason string)
		checkDeprecation := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil || len(symbol.Declarations) == 0 {
				return false, ""
			}

			// Check each declaration for deprecation by parsing JSDoc
			for _, decl := range symbol.Declarations {
				if decl == nil {
					continue
				}

				// Parse JSDoc comments to find @deprecated tag
				reason := getDeprecationReason(decl)
				if reason != "" {
					// Found @deprecated with a reason
					return true, reason
				}

				// Check if @deprecated tag exists without a reason
				if hasDeprecatedTag(decl) {
					return true, ""
				}
			}

			return false, ""
		}

		// Check if a node is a declaration (should not report on declarations)
		isDeclaration := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			switch parent.Kind {
			case ast.KindVariableDeclaration:
				// Check if this is the name of the variable being declared
				varDecl := parent.AsVariableDeclaration()
				return varDecl.Name() == node

			case ast.KindParameter:
				param := parent.AsParameterDeclaration()
				return param.Name() == node

			case ast.KindFunctionDeclaration,
				ast.KindMethodDeclaration,
				ast.KindConstructor,
				ast.KindGetAccessor,
				ast.KindSetAccessor:
				// Function-like declarations
				return true

			case ast.KindClassDeclaration:
				classDecl := parent.AsClassDeclaration()
				return classDecl.Name() == node

			case ast.KindInterfaceDeclaration:
				interfaceDecl := parent.AsInterfaceDeclaration()
				return interfaceDecl.Name() == node

			case ast.KindTypeAliasDeclaration:
				typeAlias := parent.AsTypeAliasDeclaration()
				return typeAlias.Name() == node

			case ast.KindEnumDeclaration:
				enumDecl := parent.AsEnumDeclaration()
				return enumDecl.Name() == node

			case ast.KindPropertyDeclaration,
				ast.KindPropertySignature:
				// Property declarations
				return true

			case ast.KindImportSpecifier,
				ast.KindImportClause:
				// Import declarations
				return true

			default:
				return false
			}
		}

		// Check if we're inside an import statement
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
			// Skip if this is a declaration
			if isDeclaration(node) {
				return
			}

			// Skip if inside an import
			if isInsideImport(node) {
				return
			}

			// Get the symbol for this identifier
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil {
				// Debug: print when symbol is nil
				// fmt.Printf("DEBUG: No symbol for identifier '%s' at line %d\n", node.Text(), ctx.SourceFile.GetLineAndCharacterOfPosition(node.Pos()).Line+1)
				return
			}

			// Check for deprecation through the alias chain
			var isDeprecated bool
			var deprecationReason string

			// Check if the symbol is an alias and follow the chain
			if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				// Check the alias itself
				isDeprecated, deprecationReason = checkDeprecation(symbol)

				// If not deprecated, check the aliased symbol
				if !isDeprecated {
					aliasedSymbol := ctx.TypeChecker.GetAliasedSymbol(symbol)
					if aliasedSymbol != nil {
						isDeprecated, deprecationReason = checkDeprecation(aliasedSymbol)
					}
				}
			} else {
				// Not an alias, check the symbol directly
				isDeprecated, deprecationReason = checkDeprecation(symbol)
			}

			// If deprecated, report it
			if isDeprecated {
				name := node.Text()

				if deprecationReason == "" {
					ctx.ReportNode(node, buildDeprecatedMessage(name))
				} else {
					ctx.ReportNode(node, buildDeprecatedWithReasonMessage(name, strings.TrimSpace(deprecationReason)))
				}
			}
		}

		checkMemberExpression := func(node *ast.Node) {
			memberExpr := node.AsPropertyAccessExpression()

			// For property access (a.b), check if 'b' is deprecated
			property := memberExpr.Name()
			if property == nil {
				return
			}

			// Get the type of the object
			objectType := ctx.TypeChecker.GetTypeAtLocation(memberExpr.Expression)
			if objectType == nil {
				return
			}

			// Get the property symbol
			propertyName := property.Text()
			propertySymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, propertyName)
			if propertySymbol == nil {
				return
			}

			// Check if deprecated
			isDeprecated, deprecationReason := checkDeprecation(propertySymbol)
			if isDeprecated {
				if deprecationReason == "" {
					ctx.ReportNode(property, buildDeprecatedMessage(propertyName))
				} else {
					ctx.ReportNode(property, buildDeprecatedWithReasonMessage(propertyName, strings.TrimSpace(deprecationReason)))
				}
			}
		}

		checkCallExpression := func(node *ast.Node) {
			callExpr := node.AsCallExpression()

			// Get the signature of the call
			signature := ctx.TypeChecker.GetResolvedSignature(node)
			if signature == nil {
				return
			}

			// Check if the function is deprecated
			symbol := ctx.TypeChecker.GetSymbolAtLocation(callExpr.Expression)
			if symbol == nil {
				return
			}

			isDeprecated, deprecationReason := checkDeprecation(symbol)
			if isDeprecated {
				// Use the symbol name instead of expression text to avoid issues with complex expressions
				exprText := symbol.Name
				if deprecationReason == "" {
					ctx.ReportNode(callExpr.Expression, buildDeprecatedMessage(exprText))
				} else {
					ctx.ReportNode(callExpr.Expression, buildDeprecatedWithReasonMessage(exprText, strings.TrimSpace(deprecationReason)))
				}
			}
		}

		checkNewExpression := func(node *ast.Node) {
			newExpr := node.AsNewExpression()

			// Get the signature of the constructor call
			signature := ctx.TypeChecker.GetResolvedSignature(node)
			if signature == nil {
				return
			}

			// Check if the constructor or class is deprecated
			symbol := ctx.TypeChecker.GetSymbolAtLocation(newExpr.Expression)
			if symbol == nil {
				return
			}

			isDeprecated, deprecationReason := checkDeprecation(symbol)
			if isDeprecated {
				// Use the symbol name instead of expression text to avoid issues with complex expressions
				exprText := symbol.Name
				if deprecationReason == "" {
					ctx.ReportNode(newExpr.Expression, buildDeprecatedMessage(exprText))
				} else {
					ctx.ReportNode(newExpr.Expression, buildDeprecatedWithReasonMessage(exprText, strings.TrimSpace(deprecationReason)))
				}
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier:               checkIdentifier,
			ast.KindPropertyAccessExpression: checkMemberExpression,
			ast.KindCallExpression:           checkCallExpression,
			ast.KindNewExpression:            checkNewExpression,
		}
	},
}
