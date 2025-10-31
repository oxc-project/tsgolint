package no_deprecated

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
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

var NoDeprecatedRule = rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Extract the deprecation reason from JSDoc comments
		// This implementation is based on Parser.withJSDoc approach:
		// - Get leading comment ranges
		// - Filter for JSDoc comments (/** ... */)
		// - Parse the comment text to find @deprecated tag
		// - Extract the reason text after @deprecated
		getDeprecationReason := func(node *ast.Node) string {
			// Check if node has JSDoc
			if node.Flags&ast.NodeFlagsHasJSDoc == 0 {
				return ""
			}

			// Get the source text
			sourceText := ctx.SourceFile.Text()

			// Get leading comments for this node
			for commentRange := range scanner.GetLeadingCommentRanges(nil, sourceText, node.Pos()) {
				// Check if this is a JSDoc comment (starts with /**)
				start := commentRange.Pos()
				end := commentRange.End()

				if end <= start+3 {
					continue
				}

				// Check for /** but not /**/ (which is not JSDoc)
				if sourceText[start+1] != '*' || sourceText[start+2] != '*' || sourceText[start+3] == '/' {
					continue
				}

				// Extract comment text
				commentText := sourceText[start:end]

				// Look for @deprecated tag
				deprecatedIndex := strings.Index(commentText, "@deprecated")
				if deprecatedIndex == -1 {
					continue
				}

				// Extract the reason after @deprecated
				// Skip past "@deprecated" and any whitespace
				reasonStart := deprecatedIndex + len("@deprecated")

				// Find the end of the reason (next @ tag or end of comment)
				reasonText := commentText[reasonStart:]

				// Trim whitespace and extract until next tag or end
				lines := strings.Split(reasonText, "\n")
				var reasonParts []string

				for _, line := range lines {
					// Remove leading * and whitespace
					line = strings.TrimSpace(line)
					line = strings.TrimPrefix(line, "*")
					line = strings.TrimSpace(line)

					// Stop at next @ tag or end of comment
					if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "*/") {
						break
					}

					if line != "" {
						reasonParts = append(reasonParts, line)
					}
				}

				if len(reasonParts) > 0 {
					return strings.Join(reasonParts, " ")
				}
			}

			return ""
		}

		// Check if a symbol is deprecated and optionally get the deprecation reason
		// This approach is based on typescript-go's symbol_display.go implementation
		// Returns: (isDeprecated bool, reason string)
		checkDeprecation := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil || len(symbol.Declarations) == 0 {
				return false, ""
			}

			// Check if any declaration has the deprecated flag
			// Following the typescript-go approach: GetCombinedModifierFlags includes JSDoc node flags
			for _, decl := range symbol.Declarations {
				if decl == nil {
					continue
				}
				if ast.IsDeclaration(decl) {
					flags := ast.GetCombinedModifierFlags(decl)
					if flags&ast.ModifierFlagsDeprecated != 0 {
						// Found deprecation, now try to extract the reason from JSDoc
						reason := getDeprecationReason(decl)
						return true, reason
					}
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
				exprText := callExpr.Expression.Text()
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
				exprText := newExpr.Expression.Text()
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
