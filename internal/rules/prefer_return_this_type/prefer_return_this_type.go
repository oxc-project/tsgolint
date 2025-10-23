package prefer_return_this_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

func buildUseThisTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useThisType",
		Description: "Use `this` type instead.",
	}
}

var PreferReturnThisTypeRule = rule.Rule{
	Name: "prefer-return-this-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var tryGetNameInTypeNode func(name string, node *ast.Node) *ast.Node
		tryGetNameInTypeNode = func(name string, node *ast.Node) *ast.Node {
			node = ast.SkipParentheses(node)

			if ast.IsTypeReferenceNode(node) {
				n := node.AsTypeReference().TypeName
				if ast.IsIdentifier(n) && n.AsIdentifier().Text == name {
					return node
				}
			} else if node.Kind == ast.KindUnionType {
				for _, t := range node.AsUnionTypeNode().Types.Nodes {
					if found := tryGetNameInTypeNode(name, t); found != nil {
						return found
					}
				}
			}

			return nil
		}

		checkFunction := func(fn *ast.Node, originalClass *ast.Node) {
			returnType := fn.Type()
			body := fn.Body()
			if returnType == nil || body == nil {
				return
			}

			className := originalClass.Name()

			if className == nil {
				return
			}

			node := tryGetNameInTypeNode(className.Text(), returnType)
			if node == nil {
				return
			}

			params := fn.Parameters()
			if len(params) >= 1 {
				firstArg := params[0].Name()
				if ast.IsIdentifier(firstArg) && firstArg.AsIdentifier().Text == "this" {
					return
				}
			}

			classType := ctx.TypeChecker.GetTypeAtLocation(originalClass).AsInterfaceType()

			if ast.IsBlock(body) {
				hasReturnThis := false
				hasReturnClassType := ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
					expr := stmt.Expression()
					if expr == nil {
						return false
					}

					if expr.Kind == ast.KindThisKeyword {
						hasReturnThis = true
						return false
					}

					t := ctx.TypeChecker.GetTypeAtLocation(expr)
					if classType.AsType() == t {
						return true
					}

					if checker.InterfaceType_thisType(classType) == t {
						hasReturnThis = true
					}
					return false
				})

				if hasReturnClassType || !hasReturnThis {
					return
				}
			} else {
				t := ctx.TypeChecker.GetTypeAtLocation(body)
				if checker.InterfaceType_thisType(classType) != t {
					return
				}
			}

			ctx.ReportNodeWithFixes(node, buildUseThisTypeMessage(), func() []rule.RuleFix { return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "this")} })
		}

		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if !ast.IsClassLike(node.Parent) {
					return
				}

				property := node.AsPropertyDeclaration()

				if property.Initializer == nil {
					return
				}

				if ast.IsFunctionExpression(property.Initializer) || ast.IsArrowFunction(property.Initializer) {
					checkFunction(property.Initializer, node.Parent)
				}
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				if !ast.IsClassLike(node.Parent) {
					return
				}
				checkFunction(node, node.Parent)
			},
		}
	},
}
