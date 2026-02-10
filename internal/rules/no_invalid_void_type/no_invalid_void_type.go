package no_invalid_void_type

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildInvalidVoidForGenericMessage(generic string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidForGeneric",
		Description: fmt.Sprintf("%v may not have void as a type argument.", generic),
	}
}
func buildInvalidVoidNotReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturn",
		Description: "void is only valid as a return type.",
	}
}
func buildInvalidVoidNotReturnOrGenericMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrGeneric",
		Description: "void is only valid as a return type or generic type argument.",
	}
}
func buildInvalidVoidNotReturnOrThisParamMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrThisParam",
		Description: "void is only valid as return type or type of `this` parameter.",
	}
}
func buildInvalidVoidNotReturnOrThisParamOrGenericMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrThisParamOrGeneric",
		Description: "void is only valid as a return type or generic type argument or the type of a `this` parameter.",
	}
}
func buildInvalidVoidUnionConstituentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidUnionConstituent",
		Description: "void is not valid as a constituent in a union type",
	}
}

func isNodeInNodeList(list *ast.NodeList, node *ast.Node) bool {
	if list == nil {
		return false
	}
	for _, n := range list.Nodes {
		if n == node {
			return true
		}
	}
	return false
}

func compactWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func isVoidInTypeReferenceTypeArguments(node *ast.Node) (*ast.Node, bool) {
	parent := node.Parent
	if parent == nil || !ast.IsTypeReferenceNode(parent) {
		return nil, false
	}
	return parent, isNodeInNodeList(parent.AsTypeReference().TypeArguments, node)
}

func isVoidInNewExpressionTypeArguments(node *ast.Node) (*ast.Node, bool) {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindNewExpression {
		return nil, false
	}
	return parent, isNodeInNodeList(parent.AsNewExpression().TypeArguments, node)
}

func isThisParameterVoid(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || !ast.IsParameter(parent) || parent.Type() != node {
		return false
	}
	name := parent.Name()
	return name != nil && ast.IsIdentifier(name) && name.Text() == "this"
}

func isVoidInReturnTypePosition(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && ast.IsFunctionLike(parent) && parent.Type() == node
}

func getTypeReferenceNameText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil || node.Kind != ast.KindTypeReference {
		return ""
	}
	return compactWhitespace(scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, node.AsTypeReference().TypeName, false))
}

func isAllowedGenericName(typeNameText string, allowlist []string) bool {
	return utils.Some(allowlist, func(allowed string) bool {
		return compactWhitespace(allowed) == typeNameText
	})
}

func isValidUnionType(ctx rule.RuleContext, node *ast.Node, allowlist []string) bool {
	if node == nil || node.Kind != ast.KindUnionType {
		return false
	}

	for _, member := range node.AsUnionTypeNode().Types.Nodes {
		if member.Kind == ast.KindVoidKeyword || member.Kind == ast.KindNeverKeyword {
			continue
		}
		if member.Kind == ast.KindTypeReference {
			args := member.AsTypeReference().TypeArguments
			if args != nil && utils.Some(args.Nodes, func(typeArg *ast.Node) bool { return typeArg.Kind == ast.KindVoidKeyword }) {
				if allowlist == nil {
					continue
				}
				if isAllowedGenericName(getTypeReferenceNameText(ctx, member), allowlist) {
					continue
				}
			}
		}
		return false
	}

	return true
}

func getParentFunctionOrMethodDeclaration(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		if ast.IsFunctionDeclaration(current) {
			return current
		}
		if ast.IsMethodDeclaration(current) && current.Body() != nil {
			return current
		}
		current = current.Parent
	}
	return nil
}

func hasOverloadSignatures(declNode *ast.Node, typeChecker any) bool {
	checker, ok := typeChecker.(interface {
		GetSymbolAtLocation(node *ast.Node) *ast.Symbol
	})
	if !ok || declNode == nil {
		return false
	}

	name := declNode.Name()
	if name == nil {
		if ast.IsFunctionDeclaration(declNode) &&
			utils.IncludesModifier(declNode, ast.KindDefaultKeyword) &&
			utils.IncludesModifier(declNode, ast.KindExportKeyword) &&
			declNode.Body() != nil {
			parent := declNode.Parent
			if parent != nil && parent.CanHaveStatements() {
				for _, stmt := range parent.Statements() {
					if stmt == declNode || !ast.IsFunctionDeclaration(stmt) {
						continue
					}
					if stmt.Name() == nil &&
						stmt.Body() == nil &&
						utils.IncludesModifier(stmt, ast.KindDefaultKeyword) &&
						utils.IncludesModifier(stmt, ast.KindExportKeyword) {
						return true
					}
				}
			}
		}
		return false
	}

	symbol := checker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) <= 1 {
		return false
	}

	for _, decl := range symbol.Declarations {
		if decl == declNode {
			continue
		}

		if ast.IsFunctionDeclaration(decl) && decl.Body() == nil {
			return true
		}
		if ast.IsMethodDeclaration(decl) && decl.Body() == nil {
			return true
		}
		if ast.IsMethodSignatureDeclaration(decl) {
			return true
		}
	}

	return false
}

func getNotReturnOrGenericMessage(node *ast.Node) rule.RuleMessage {
	if node.Parent != nil && node.Parent.Kind == ast.KindUnionType {
		return buildInvalidVoidUnionConstituentMessage()
	}
	return buildInvalidVoidNotReturnOrGenericMessage()
}

var NoInvalidVoidTypeRule = rule.Rule{
	Name: "no-invalid-void-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoInvalidVoidTypeOptions](options, "no-invalid-void-type")

		getDefaultMessage := func(node *ast.Node) rule.RuleMessage {
			allowInGeneric := opts.AllowInGenericTypeArguments.Bool()

			if allowInGeneric && opts.AllowAsThisParameter {
				return buildInvalidVoidNotReturnOrThisParamOrGenericMessage()
			}
			if allowInGeneric {
				return getNotReturnOrGenericMessage(node)
			}
			if opts.AllowAsThisParameter {
				return buildInvalidVoidNotReturnOrThisParamMessage()
			}
			return buildInvalidVoidNotReturnMessage()
		}

		checkGenericTypeArgument := func(node *ast.Node) bool {
			typeRef, isTypeRefGenericArg := isVoidInTypeReferenceTypeArguments(node)
			_, isNewExprGenericArg := isVoidInNewExpressionTypeArguments(node)
			if !isTypeRefGenericArg && !isNewExprGenericArg {
				return false
			}

			allowlist := opts.AllowInGenericTypeArguments.Object()
			if allowlist != nil {
				if isTypeRefGenericArg {
					typeNameText := getTypeReferenceNameText(ctx, typeRef)
					if !isAllowedGenericName(typeNameText, *allowlist) {
						ctx.ReportNode(node, buildInvalidVoidForGenericMessage(typeNameText))
					}
					return true
				}
				// keep parity with typescript-eslint: only type-reference generics are allowlisted
				ctx.ReportNode(node, getNotReturnOrGenericMessage(node))
				return true
			}

			if !opts.AllowInGenericTypeArguments.Bool() {
				if opts.AllowAsThisParameter {
					ctx.ReportNode(node, buildInvalidVoidNotReturnOrThisParamMessage())
				} else {
					ctx.ReportNode(node, buildInvalidVoidNotReturnMessage())
				}
				return true
			}

			return true
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				var allowlist []string
				if list := opts.AllowInGenericTypeArguments.Object(); list != nil {
					allowlist = *list
				}

				if checkGenericTypeArgument(node) {
					return
				}

				if opts.AllowInGenericTypeArguments.Bool() && node.Parent != nil && ast.IsTypeParameterDeclaration(node.Parent) && node.Parent.AsTypeParameter().DefaultType == node {
					return
				}

				if node.Parent != nil && node.Parent.Kind == ast.KindUnionType && isValidUnionType(ctx, node.Parent, allowlist) {
					return
				}

				if node.Parent != nil && node.Parent.Kind == ast.KindUnionType {
					if decl := getParentFunctionOrMethodDeclaration(node.Parent); decl != nil && hasOverloadSignatures(decl, ctx.TypeChecker) {
						return
					}
				}

				if opts.AllowAsThisParameter && isThisParameterVoid(node) {
					return
				}

				if isVoidInReturnTypePosition(node) {
					return
				}

				ctx.ReportNode(node, getDefaultMessage(node))
			},
		}
	},
}
