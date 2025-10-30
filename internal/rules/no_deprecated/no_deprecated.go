package no_deprecated

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
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
		// Get the JSDoc deprecation tag for a symbol or signature
		// NOTE: This requires Symbol.GetJsDocTags() to be exposed in the typescript-go shim
		// The method exists in TypeScript's compiler API: symbol.getJsDocTags(checker)
		getJsDocDeprecation := func(symbol *ast.Symbol) string {
			if symbol == nil {
				return ""
			}

			// TODO: This method needs to be added to the typescript-go shim
			// For now, this is a placeholder implementation
			// The actual TypeScript API is: symbol.getJsDocTags(checker)
			// which returns an array of JSDocTagInfo objects
			
			// Check each declaration of the symbol for JSDoc comments
			if symbol.Declarations == nil {
				return ""
			}

			for _, decl := range symbol.Declarations {
				if decl == nil {
					continue
				}
				
				// TODO: Parse JSDoc comments from the declaration
				// This would involve using parser.GetJSDocCommentRanges
				// and then parsing the comment text for @deprecated tag
				// For now, this is incomplete and needs the shim to be updated
			}

			return ""
		}

		// Check if a node is a declaration (should not report on declarations)
		isDeclaration := func(node *ast.Node) bool {
			parent := node.Parent()
			if parent == nil {
				return false
			}

			switch parent.Kind() {
			case ast.KindVariableDeclaration:
				// Check if this is the name of the variable being declared
				varDecl := parent.AsVariableDeclaration()
				return varDecl.Name == node

			case ast.KindParameter:
				param := parent.AsParameter()
				return param.Name == node

			case ast.KindFunctionDeclaration,
				ast.KindMethodDeclaration,
				ast.KindConstructor,
				ast.KindGetAccessor,
				ast.KindSetAccessor:
				// Function-like declarations
				return true

			case ast.KindClassDeclaration:
				classDecl := parent.AsClassDeclaration()
				return classDecl.Name == node

			case ast.KindInterfaceDeclaration:
				interfaceDecl := parent.AsInterfaceDeclaration()
				return interfaceDecl.Name == node

			case ast.KindTypeAliasDeclaration:
				typeAlias := parent.AsTypeAliasDeclaration()
				return typeAlias.Name == node

			case ast.KindEnumDeclaration:
				enumDecl := parent.AsEnumDeclaration()
				return enumDecl.Name == node

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
				kind := current.Kind()
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
				current = current.Parent()
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
				return
			}

			// Check for deprecation through the alias chain
			var deprecationReason string

			// Check if the symbol is an alias and follow the chain
			if checker.IsSymbolFlagSet(symbol, checker.SymbolFlagsAlias) {
				// Check the alias itself
				deprecationReason = getJsDocDeprecation(symbol)

				// If not deprecated, check the aliased symbol
				if deprecationReason == "" {
					aliasedSymbol := ctx.TypeChecker.GetAliasedSymbol(symbol)
					if aliasedSymbol != nil {
						deprecationReason = getJsDocDeprecation(aliasedSymbol)
					}
				}
			} else {
				// Not an alias, check the symbol directly
				deprecationReason = getJsDocDeprecation(symbol)
			}

			// If deprecated, report it
			if deprecationReason != "" {
				name := node.Text(ctx.SourceFile)

				if deprecationReason == " " || deprecationReason == "" {
					ctx.ReportNode(node, buildDeprecatedMessage(name))
				} else {
					ctx.ReportNode(node, buildDeprecatedWithReasonMessage(name, strings.TrimSpace(deprecationReason)))
				}
			}
		}

		checkMemberExpression := func(node *ast.Node) {
			memberExpr := node.AsMemberExpression()

			// For property access (a.b), check if 'b' is deprecated
			property := memberExpr.Name
			if property == nil {
				return
			}

			// Get the type of the object
			objectType := ctx.TypeChecker.GetTypeAtLocation(memberExpr.Expression)
			if objectType == nil {
				return
			}

			// Get the property symbol
			propertyName := property.Text(ctx.SourceFile)
			propertySymbol := objectType.GetProperty(propertyName)
			if propertySymbol == nil {
				return
			}

			// Check if deprecated
			deprecationReason := getJsDocDeprecation(propertySymbol)
			if deprecationReason != "" {
				if deprecationReason == " " || deprecationReason == "" {
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

			// Check if the signature is deprecated
			deprecationReason := ""
			if signature != nil {
				// Try to get JSDoc from the signature
				// Note: TypeScript's signature interface may have GetJsDocTags method
				// but we need to check the symbol instead
				symbol := ctx.TypeChecker.GetSymbolAtLocation(callExpr.Expression)
				if symbol != nil {
					deprecationReason = getJsDocDeprecation(symbol)
				}
			}

			if deprecationReason != "" {
				exprText := callExpr.Expression.Text(ctx.SourceFile)
				if deprecationReason == " " || deprecationReason == "" {
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

			deprecationReason := getJsDocDeprecation(symbol)
			if deprecationReason != "" {
				exprText := newExpr.Expression.Text(ctx.SourceFile)
				if deprecationReason == " " || deprecationReason == "" {
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
