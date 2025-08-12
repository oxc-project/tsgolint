package no_unused_vars

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/typescript-eslint/tsgolint/internal/gorule"
)

func buildUnusedVarMessage(varName string) gorule.GoRuleMessage {
	return gorule.GoRuleMessage{
		Id:          "unusedVar",
		Description: "Variable '" + varName + "' is declared but never used.",
	}
}

func buildRemoveVarMessage() gorule.GoRuleMessage {
	return gorule.GoRuleMessage{
		Id:          "removeVar",
		Description: "Remove unused variable.",
	}
}

var NoUnusedVarsRule = gorule.GoRule{
	Name: "no-unused-vars",
	Run: func(ctx gorule.GoRuleContext, options any) gorule.GoRuleListeners {
		// Track variable declarations and usages
		declaredVars := make(map[string]*ast.Ident)
		usedVars := make(map[string]bool)

		return gorule.GoRuleListeners{
			"FuncDecl": func(node ast.Node) {
				funcDecl := node.(*ast.FuncDecl)
				if funcDecl.Body == nil {
					return
				}

				// Clear for each function scope
				declaredVars = make(map[string]*ast.Ident)
				usedVars = make(map[string]bool)

				// Collect parameter names (these are implicitly used)
				if funcDecl.Type.Params != nil {
					for _, param := range funcDecl.Type.Params.List {
						for _, name := range param.Names {
							if name.Name != "_" {
								usedVars[name.Name] = true
							}
						}
					}
				}

				// Walk function body to find variable declarations and usages
				ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.AssignStmt:
						if node.Tok == token.DEFINE { // := assignment
							for _, lhs := range node.Lhs {
								if ident, ok := lhs.(*ast.Ident); ok {
									if ident.Name != "_" && !strings.HasPrefix(ident.Name, "_") {
										declaredVars[ident.Name] = ident
									}
								}
							}
						}
						// Mark right-hand side variables as used
						for _, rhs := range node.Rhs {
							ast.Inspect(rhs, func(subNode ast.Node) bool {
								if ident, ok := subNode.(*ast.Ident); ok {
									usedVars[ident.Name] = true
								}
								return true
							})
						}
					case *ast.ValueSpec:
						// Variable declarations in var blocks
						for _, name := range node.Names {
							if name.Name != "_" && !strings.HasPrefix(name.Name, "_") {
								declaredVars[name.Name] = name
							}
						}
						// Mark values as used
						for _, value := range node.Values {
							ast.Inspect(value, func(subNode ast.Node) bool {
								if ident, ok := subNode.(*ast.Ident); ok {
									usedVars[ident.Name] = true
								}
								return true
							})
						}
					case *ast.Ident:
						// Mark identifier as used if it's referencing a variable
						if node.Obj == nil && node.Name != "nil" && node.Name != "true" && node.Name != "false" {
							usedVars[node.Name] = true
						}
					}
					return true
				})

				// Report unused variables
				for varName, identNode := range declaredVars {
					if !usedVars[varName] {
						ctx.ReportNodeWithSuggestions(
							identNode,
							buildUnusedVarMessage(varName),
							gorule.GoRuleSuggestion{
								Message: buildRemoveVarMessage(),
								FixesArr: []gorule.GoRuleFix{
									{
										Text:  "_",
										Range: identNode.Pos(),
										End:   identNode.End(),
									},
								},
							},
						)
					}
				}
			},
		}
	},
}