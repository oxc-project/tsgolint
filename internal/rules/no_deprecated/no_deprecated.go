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

var NoDeprecatedRule = rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Extract the deprecation reason from JSDoc comments
		// This implementation is based on Parser.withJSDoc approach:
		// - Get leading comment ranges
		// - Filter for JSDoc comments (/** ... */)
		// - Parse the comment text to find @deprecated tag
		// - Extract the reason text after @deprecated
		getDeprecationReason := func(node *ast.Node) (bool, string) {
			// TODO

			callLikeNode := getCallLikeNode(node)
			if callLikeNode != nil {
				return getCallLikeDeprecation(callLikeNode)
			}

			if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute && node.Kind != ast.KindSuperKeyword {
				return getJSXAttributeDeprecation(node.Parent.Parent, node.Name())
			}

			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Kind != ast.KindSuperKeyword {



    //     const property = services
    //       .getTypeAtLocation(node.parent.parent)
    //       .getProperty(node.name);
    //     const propertySymbol = services.getSymbolAtLocation(node);
    //     const valueSymbol = checker.getShorthandAssignmentValueSymbol(
    //       propertySymbol?.valueDeclaration,
    //     );
    //     return (
    //       searchForDeprecationInAliasesChain(propertySymbol, true) ??
    //       getJsDocDeprecation(property) ??
    //       getJsDocDeprecation(propertySymbol) ??
    //       getJsDocDeprecation(valueSymbol)
    //     );
			}

			return searchForDeprecationInAliasesChain(
			ctx.TypeChecker.GetSymbolAtLocation(node),
			true,
		)

	// 		      if (
    //     node.parent.type === AST_NODE_TYPES.Property &&
    //     node.type !== AST_NODE_TYPES.Super
    //   ) {

    //   }

    //   return searchForDeprecationInAliasesChain(
    //     services.getSymbolAtLocation(node),
    //     true,
    //   );














			// // Check if node has JSDoc
			// if node.Flags&ast.NodeFlagsHasJSDoc == 0 {
			// 	// 	return ""
			// }

			// // Get the source text
			// sourceText := ctx.SourceFile.Text()

			// // Get leading comments for this node
			// for commentRange := range scanner.GetLeadingCommentRanges(nil, sourceText, node.Pos()) {

			// 	// Check if this is a JSDoc comment (starts with /**)
			// 	start := commentRange.Pos()
			// 	end := commentRange.End()

			// 	if end <= start+3 {
			// 		continue
			// 	}

			// 	// Check for /** but not /**/ (which is not JSDoc)
			// 	if sourceText[start+1] != '*' || sourceText[start+2] != '*' || sourceText[start+3] == '/' {
			// 		continue
			// 	}

			// 	// Extract comment text
			// 	commentText := sourceText[start:end]

			// 	// Look for @deprecated tag
			// 	deprecatedIndex := strings.Index(commentText, "@deprecated")
			// 	if deprecatedIndex == -1 {
			// 		continue
			// 	}

			// 	// Extract the reason after @deprecated
			// 	// Skip past "@deprecated" and any whitespace
			// 	reasonStart := deprecatedIndex + len("@deprecated")

			// 	// Find the end of the reason (next @ tag or end of comment)
			// 	reasonText := commentText[reasonStart:]

			// 	// Remove trailing */ if present
			// 	reasonText = strings.TrimSuffix(reasonText, "*/")
			// 	reasonText = strings.TrimSpace(reasonText)

			// 	// For multi-line JSDoc, process line by line
			// 	lines := strings.Split(reasonText, "\n")
			// 	var reasonParts []string

			// 	for _, line := range lines {
			// 		// Remove leading * and whitespace
			// 		line = strings.TrimSpace(line)
			// 		line = strings.TrimPrefix(line, "*")
			// 		line = strings.TrimSpace(line)

			// 		// Stop at next @ tag
			// 		if strings.HasPrefix(line, "@") {
			// 			break
			// 		}

			// 		if line != "" {
			// 			reasonParts = append(reasonParts, line)
			// 		}
			// 	}

			// 	if len(reasonParts) > 0 {
			// 		return strings.Join(reasonParts, " ")
			// 	}
			// }

			// return ""
		}

		// Check if a symbol is deprecated and optionally get the deprecation reason
		// Use TypeScript's IsDeprecatedDeclaration API which properly checks for @deprecated JSDoc tags
		// Returns: (isDeprecated bool, reason string)
		checkDeprecation := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil || len(symbol.Declarations) == 0 {
				return false, ""
			}

			// Check if any declaration has the deprecated flag
			for _, decl := range symbol.Declarations {
				if decl == nil {
					continue
				}
				if ast.IsDeclaration(decl) {
					if ctx.TypeChecker.IsDeprecatedDeclaration(decl) {
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
				// Property declarations (in classes and interfaces)
				return true

			case ast.KindPropertyAssignment:
				// For property assignments in object literals, only the name is a declaration
				// The value is a reference
				propAssign := parent.AsPropertyAssignment()
				return propAssign.Name() == node

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
			fmt.Println("check ident")
			if isDeclaration(node) || isInsideImport(node) {
				return
			}

			isDeprecated, deprecationReason := getDeprecationReason(node)

			if !isDeprecated {
				return
			}

			ty := ctx.TypeChecker.GetTypeAtLocation(node)

			// TODO: if type OR value is allowed, skip

			if deprecationReason == "" {
				ctx.ReportNode(node, buildDeprecatedMessage(node.Text()))
			} else {
				ctx.ReportNode(node, buildDeprecatedWithReasonMessage(node.Text(), strings.TrimSpace(deprecationReason)))
			}


			return
		}

		checkComputedPropertyName := func(node *ast.Node) {
			computedPropertyName := node.AsComputedPropertyName()

			propertyType := ctx.TypeChecker.GetTypeAtLocation(computedPropertyName.Expression)

			if propertyType == nil {
				return
			}

			if utils.IsTypeFlagSet(propertyType, checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBigIntLiteral ) {
				// objectType := ctx.TypeChecker.GetTypeAtLocation(computed)

			 	// var propertyName string
				// if propertyType.IsStringLiteral() {
				// 	 propertyName = propertyType.AsLiteralType().String()
				// } else if propertyType.IsNumberLiteral() {
				// 	 propertyName = fmt.Sprintf("%v", propertyType.AsLiteralType().Value())
				// } else if propertyType.IsBigIntLiteral() {
				// 	 propertyName = propertyType.AsBigIntLiteralType().Text()
				// } else {
				// 	 return
				// }

				// TODO
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Parent != nil {
					return
				}
				if node.Parent.Kind == ast.KindExportDeclaration {
					return
				}
				if node.Parent.Kind == ast.KindExportSpecifier {
					// only deal with the alias (exported) side, not the local binding
        //   if (parent.exported != node) {
        //     return;
        //   }

					symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
        // aliasDeprecation := getJsDocDeprecation(symbol)
					if aliasDeprecation != nil {
						return
					}
				}

				checkIdentifier(node)
			},
			ast.KindComputedPropertyName:     checkComputedPropertyName,
			ast.KindPropertyAccessExpression: checkMemberExpression,
			ast.KindPrivateIdentifier:        checkIdentifier,
			ast.KindSuperKeyword:             checkIdentifier,
		}
	},
}
