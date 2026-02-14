package no_unnecessary_qualifier

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUnnecessaryQualifierMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryQualifier",
		Description: fmt.Sprintf("Qualifier is unnecessary since '%s' is in scope.", name),
	}
}

var NoUnnecessaryQualifierRule = rule.Rule{
	Name: "no-unnecessary-qualifier",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		namespacesInScope := []*ast.Node{}
		var currentFailedNamespaceExpression *ast.Node

		enterDeclaration := func(node *ast.Node) {
			namespacesInScope = append(namespacesInScope, node)
		}

		exitDeclaration := func() {
			if len(namespacesInScope) == 0 {
				return
			}
			namespacesInScope = namespacesInScope[:len(namespacesInScope)-1]
		}

		var symbolIsNamespaceInScope func(symbol *ast.Symbol) bool
		symbolIsNamespaceInScope = func(symbol *ast.Symbol) bool {
			if symbol == nil {
				return false
			}

			for _, declaration := range symbol.Declarations {
				if slices.Contains(namespacesInScope, declaration) {
					return true
				}
			}

			if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsAlias) {
				return symbolIsNamespaceInScope(ctx.TypeChecker.GetAliasedSymbol(symbol))
			}

			return false
		}

		getSymbolInScope := func(node *ast.Node, flags ast.SymbolFlags, name string) *ast.Symbol {
			scopeSymbols := ctx.TypeChecker.GetSymbolsInScope(node, flags)
			for _, scopeSymbol := range scopeSymbols {
				if scopeSymbol.Name == name {
					return scopeSymbol
				}
			}
			return nil
		}

		symbolsAreEqual := func(accessed *ast.Symbol, inScope *ast.Symbol) bool {
			if accessed == nil || inScope == nil {
				return false
			}
			return accessed == ctx.TypeChecker.GetExportSymbolOfSymbol(inScope)
		}

		qualifierIsUnnecessary := func(qualifier *ast.Node, name *ast.Node) bool {
			namespaceSymbol := ctx.TypeChecker.GetSymbolAtLocation(qualifier)
			if namespaceSymbol == nil || !symbolIsNamespaceInScope(namespaceSymbol) {
				return false
			}

			accessedSymbol := ctx.TypeChecker.GetSymbolAtLocation(name)
			if accessedSymbol == nil {
				return false
			}

			inScopeSymbol := getSymbolInScope(qualifier, accessedSymbol.Flags, name.Text())
			return inScopeSymbol != nil && symbolsAreEqual(accessedSymbol, inScopeSymbol)
		}

		visitNamespaceAccess := func(node *ast.Node, qualifier *ast.Node, name *ast.Node) {
			if currentFailedNamespaceExpression != nil {
				return
			}
			if !qualifierIsUnnecessary(qualifier, name) {
				return
			}

			currentFailedNamespaceExpression = node
			ctx.ReportNodeWithFixes(qualifier, buildUnnecessaryQualifierMessage(name.Text()), func() []rule.RuleFix {
				qualifierStart := utils.TrimNodeTextRange(ctx.SourceFile, qualifier).Pos()
				nameStart := utils.TrimNodeTextRange(ctx.SourceFile, name).Pos()
				return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(qualifierStart, nameStart))}
			})
		}

		resetCurrentNamespaceExpression := func(node *ast.Node) {
			if node == currentFailedNamespaceExpression {
				currentFailedNamespaceExpression = nil
			}
		}

		var isEntityNameExpression func(node *ast.Node) bool
		isEntityNameExpression = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindIdentifier:
				return true
			case ast.KindPropertyAccessExpression:
				return isEntityNameExpression(node.AsPropertyAccessExpression().Expression)
			default:
				return false
			}
		}

		return rule.RuleListeners{
			ast.KindEnumDeclaration: enterDeclaration,
			rule.ListenerOnExit(ast.KindEnumDeclaration): func(_ *ast.Node) {
				exitDeclaration()
			},

			ast.KindModuleBlock: func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
					enterDeclaration(node.Parent)
				}
			},
			rule.ListenerOnExit(ast.KindModuleBlock): func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
					exitDeclaration()
				}
			},

			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propertyAccess := node.AsPropertyAccessExpression()
				name := propertyAccess.Name()
				if name == nil {
					return
				}
				if isEntityNameExpression(propertyAccess.Expression) {
					visitNamespaceAccess(node, propertyAccess.Expression, name)
				}
			},
			rule.ListenerOnExit(ast.KindPropertyAccessExpression): func(node *ast.Node) {
				resetCurrentNamespaceExpression(node)
			},

			ast.KindQualifiedName: func(node *ast.Node) {
				qualifiedName := node.AsQualifiedName()
				visitNamespaceAccess(node, qualifiedName.Left, qualifiedName.Right)
			},
			rule.ListenerOnExit(ast.KindQualifiedName): func(node *ast.Node) {
				resetCurrentNamespaceExpression(node)
			},
		}
	},
}
