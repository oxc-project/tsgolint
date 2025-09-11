package no_deprecated

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
		// Helper function to check if a symbol or its declarations are deprecated
		isSymbolDeprecated := func(symbol *ast.Symbol) bool {
			if symbol == nil || symbol.Declarations == nil {
				return false
			}

			// Check each declaration for deprecation
			for _, declaration := range symbol.Declarations {
				if declaration.Flags&ast.NodeFlagsDeprecated != 0 {
					return true
				}
			}

			return false
		}

		// Helper function to report deprecation
		reportDeprecation := func(node *ast.Node, name string) {
			// For now, we don't extract the reason from JSDoc
			// This would require additional work to parse JSDoc comments
			msg := buildDeprecatedMessage(name)
			ctx.ReportNode(node, msg)
		}

		// Helper function to check if we should skip this node
		shouldSkipNode := func(node *ast.Node) bool {
			if node.Parent == nil {
				return false
			}
			
			parent := node.Parent
			
			// Skip if this is the name of a declaration
			if ast.IsDeclaration(parent) && ast.GetNameOfDeclaration(parent) == node {
				return true
			}
			
			// Skip if this is part of a type node (type references, etc.)
			if ast.IsTypeNode(parent) {
				return true
			}
			
			// Skip if this is in a property declaration context (defining the property)
			if ast.IsPropertyDeclaration(parent) || ast.IsPropertySignatureDeclaration(parent) ||
			   ast.IsMethodDeclaration(parent) || ast.IsMethodSignatureDeclaration(parent) ||
			   ast.IsGetAccessorDeclaration(parent) || ast.IsSetAccessorDeclaration(parent) {
				return true
			}
			
			// Skip if this is the name part of a property access (handled by PropertyAccessExpression listener)
			if ast.IsPropertyAccessExpression(parent) && parent.AsPropertyAccessExpression().Name() == node {
				return true
			}
			
			// Skip if this is in an object literal property assignment (the property name itself)
			if ast.IsPropertyAssignment(parent) && parent.AsPropertyAssignment().Name() == node {
				return true
			}
			
			// Skip if this is in binding pattern (destructuring)
			if ast.IsBindingElement(parent) && parent.AsBindingElement().Name() == node {
				// This is a destructuring pattern - we should check it
				return false
			}
			
			return false
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Kind != ast.KindIdentifier {
					return
				}
				
				// Skip if we should not check this identifier
				if shouldSkipNode(node) {
					return
				}
				
				// Skip if this is not a value reference (e.g., it's a type reference)
				// Get the symbol at this location
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return
				}

				// Check if the symbol is deprecated
				if isSymbolDeprecated(symbol) {
					name := node.AsIdentifier().Text
					reportDeprecation(node, name)
				}
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if !ast.IsPropertyAccessExpression(node) {
					return
				}
				
				prop := node.AsPropertyAccessExpression()
				memberNode := prop.Name()
				
				if !ast.IsIdentifier(memberNode) {
					return
				}
				
				memberName := memberNode.AsIdentifier().Text
				
				// Get the type of the object being accessed
				objectType := ctx.TypeChecker.GetTypeAtLocation(prop.Expression)
				if objectType == nil {
					return
				}

				// Get the property symbol
				propSymbol := ctx.TypeChecker.GetPropertyOfType(objectType, memberName)
				if propSymbol == nil {
					return
				}

				// Check if the property is deprecated
				if isSymbolDeprecated(propSymbol) {
					reportDeprecation(memberNode, memberName)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if !ast.IsElementAccessExpression(node) {
					return
				}
				
				elem := node.AsElementAccessExpression()
				if elem.ArgumentExpression == nil || !ast.IsStringLiteral(elem.ArgumentExpression) {
					return
				}
				
				memberName := elem.ArgumentExpression.AsStringLiteral().Text
				
				// Get the type of the object being accessed
				objectType := ctx.TypeChecker.GetTypeAtLocation(elem.Expression)
				if objectType == nil {
					return
				}

				// Get the property symbol
				propSymbol := ctx.TypeChecker.GetPropertyOfType(objectType, memberName)
				if propSymbol == nil {
					return
				}

				// Check if the property is deprecated
				if isSymbolDeprecated(propSymbol) {
					reportDeprecation(elem.ArgumentExpression, memberName)
				}
			},
		}
	},
}