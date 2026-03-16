package require_using_for_disposable

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildRequireUsingMessage(typeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireUsing",
		Description: fmt.Sprintf("Type %q is disposable and must be declared with \"using\" or \"await using\"", typeName),
	}
}

func isDisposableType(typeChecker *checker.Checker, t *checker.Type) bool {
	for _, typePart := range utils.UnionTypeParts(t) {
		if utils.GetWellKnownSymbolPropertyOfType(typePart, "dispose", typeChecker) != nil ||
			utils.GetWellKnownSymbolPropertyOfType(typePart, "asyncDispose", typeChecker) != nil {
			return true
		}
	}

	return false
}

var RequireUsingForDisposableRule = rule.Rule{
	Name: "require-using-for-disposable",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarUsing(node) || ast.IsVarAwaitUsing(node) {
					return
				}

				declarationList := node.AsVariableDeclarationList()
				for _, declarator := range declarationList.Declarations.Nodes {
					init := declarator.Initializer()
					if init == nil {
						continue
					}

					t := ctx.TypeChecker.GetTypeAtLocation(init)
					if !isDisposableType(ctx.TypeChecker, t) {
						continue
					}

					ctx.ReportNode(declarator, buildRequireUsingMessage(ctx.TypeChecker.TypeToString(t)))
				}
			},
		}
	},
}
