package related_getter_setter_pairs

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMismatchMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatch",
		Description: "`get()` type should be assignable to its equivalent `set()` type.",
	}
}

var RelatedGetterSetterPairsRule = rule.Rule{
	Name: "related-getter-setter-pairs",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkAccessorsPair := func(getter *ast.GetAccessorDeclaration, setter *ast.SetAccessorDeclaration) {
			getType := ctx.TypeChecker.GetTypeAtLocation(getter.AsNode())
			setType := ctx.TypeChecker.GetTypeAtLocation(setter.Parameters.Nodes[0])

			if !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, getType, setType) {
				ctx.ReportNode(getter.Type, buildMismatchMessage())
			}
		}
		checkMembers := func(node *ast.Node) {
			members := node.Members()
			if members == nil {
				return
			}

			getAccessors := map[string]*ast.GetAccessorDeclaration{}
			setAccessors := map[string]*ast.SetAccessorDeclaration{}

			for _, member := range members {
				if ast.IsGetAccessorDeclaration(member) {
					m := member.AsGetAccessorDeclaration()
					if m.Type == nil {
						continue
					}
					name, _ := utils.GetNameFromMember(ctx.SourceFile, m.Name())
					if setAccessor, found := setAccessors[name]; found {
						checkAccessorsPair(m, setAccessor)
					} else {
						getAccessors[name] = m
					}
				} else if ast.IsSetAccessorDeclaration(member) {
					m := member.AsSetAccessorDeclaration()
					if len(m.Parameters.Nodes) != 1 {
						continue
					}
					name, _ := utils.GetNameFromMember(ctx.SourceFile, m.Name())
					if getAccessor, found := getAccessors[name]; found {
						checkAccessorsPair(getAccessor, m)
					} else {
						setAccessors[name] = m
					}
				}
			}
		}
		return rule.RuleListeners{
			ast.KindClassDeclaration:     checkMembers,
			ast.KindClassExpression:      checkMembers,
			ast.KindInterfaceDeclaration: checkMembers,
			ast.KindTypeLiteral:          checkMembers,
		}
	},
}
