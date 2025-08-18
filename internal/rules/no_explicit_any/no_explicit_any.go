package no_explicit_any

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildExplicitAnyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "explicitAny",
		Description: "Unexpected `any`. Specify a different type.",
	}
}

func buildExplicitAnyWithIgnoreRestArgsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "explicitAnyWithIgnoreRestArgs",
		Description: "Unexpected `any`. Specify a different type. Use `...args: never[]` to ignore rest args.",
	}
}

func buildExplicitAnyWithIgnoreRestArgsSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "explicitAnyWithIgnoreRestArgsSuggestion",
		Description: "Use `...args: never[]` to ignore rest args.",
	}
}

func buildExplicitAnyWithIgnoreRestArgsSuggestion(node *ast.Node, ctx rule.RuleContext) rule.RuleSuggestion {
	return rule.RuleSuggestion{
		Message: buildExplicitAnyWithIgnoreRestArgsSuggestionMessage(),
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, node, "never[]"),
		},
	}
}

var NoExplicitAnyRule = rule.Rule{
	Name: "no-explicit-any",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkVariableDeclaration := func(node *ast.Node) {
			// Check if the variable type annotation is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkParameter := func(node *ast.Node) {
			// Check if the parameter type is `any`
			if node.Type() != nil {
				paramType := ctx.TypeChecker.GetTypeAtLocation(node.Type())
				if utils.IsTypeAnyType(paramType) {
					// For rest parameters, provide a suggestion to use `never[]`
					if node.AsParameterDeclaration().DotDotDotToken != nil {
						ctx.ReportNodeWithSuggestions(node.Type(), buildExplicitAnyWithIgnoreRestArgsMessage(), buildExplicitAnyWithIgnoreRestArgsSuggestion(node.Type(), ctx))
					} else {
						ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
					}
				}
			}
		}

		checkFunctionDeclaration := func(node *ast.Node) {
			// Check if the function return type is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkMethodDeclaration := func(node *ast.Node) {
			// Check if the method return type is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkPropertyDeclaration := func(node *ast.Node) {
			// Check if the property type is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkInterfaceDeclaration := func(node *ast.Node) {
			// Check if any property in the interface has type `any`
			// This would need to traverse the interface body
			// For now, we'll check the interface itself
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkPropertySignature := func(node *ast.Node) {
			// Check if the property type is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkIndexSignature := func(node *ast.Node) {
			// Check if the index signature type is `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkTypeAliasDeclaration := func(node *ast.Node) {
			// Check if the type alias resolves to `any`
			if node.Type() != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.Type())) {
				ctx.ReportNode(node.Type(), buildExplicitAnyMessage())
			}
		}

		checkTypeParameter := func(node *ast.Node) {
			// Check if the type parameter default is `any`
			if node.AsTypeParameter().DefaultType != nil && utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node.AsTypeParameter().DefaultType)) {
				ctx.ReportNode(node.AsTypeParameter().DefaultType, buildExplicitAnyMessage())
			}
		}

		// Check type references (for type aliases, generic parameters, etc.)
		checkTypeReference := func(node *ast.Node) {
			if utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node)) {
				ctx.ReportNode(node, buildExplicitAnyMessage())
			}
		}

		// Check array types
		checkArrayType := func(node *ast.Node) {
			// Check if the array element type is `any`
			elementType := ctx.TypeChecker.GetTypeAtLocation(node.AsArrayTypeNode().ElementType)
			if utils.IsTypeAnyType(elementType) {
				// Check if this array type is used in a rest parameter context
				if node.Parent != nil && node.Parent.Kind == ast.KindParameter && node.Parent.AsParameterDeclaration().DotDotDotToken != nil {
					ctx.ReportNodeWithSuggestions(node, buildExplicitAnyWithIgnoreRestArgsMessage(), buildExplicitAnyWithIgnoreRestArgsSuggestion(node, ctx))
				} else {
					ctx.ReportNode(node.AsArrayTypeNode().ElementType, buildExplicitAnyMessage())
				}
			}
		}

		// Check union types
		checkUnionType := func(node *ast.Node) {
			if utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node)) {
				ctx.ReportNode(node, buildExplicitAnyMessage())
			}
		}

		// Check type annotations specifically
		checkTypeAnnotation := func(node *ast.Node) {
			if utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(node)) {
				ctx.ReportNode(node, buildExplicitAnyMessage())
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration:      checkVariableDeclaration,
			ast.KindParameter:                checkParameter,
			ast.KindFunctionDeclaration:      checkFunctionDeclaration,
			ast.KindMethodDeclaration:        checkMethodDeclaration,
			ast.KindPropertyDeclaration:      checkPropertyDeclaration,
			ast.KindPropertySignature:        checkPropertySignature,
			ast.KindIndexSignature:           checkIndexSignature,
			ast.KindInterfaceDeclaration:     checkInterfaceDeclaration,
			ast.KindTypeAliasDeclaration:     checkTypeAliasDeclaration,
			ast.KindTypeParameter:            checkTypeParameter,
			ast.KindTypeReference:            checkTypeReference,
			ast.KindArrayType:                checkArrayType,
			ast.KindUnionType:                checkUnionType,
			ast.KindTypeKeyword:              checkTypeAnnotation,
		}
	},
}
