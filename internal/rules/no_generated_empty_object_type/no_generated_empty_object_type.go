package no_generated_empty_object_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var NoGeneratedEmptyObjectTypeRule = rule.Rule{
	Name: "no-generated-empty-object-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		isEmptyObjectType := func(t *checker.Type) bool {
			return utils.IsObjectType(t) &&
				checker.Type_objectFlags(t)&checker.ObjectFlagsClassOrInterface == 0 &&
				len(checker.Checker_getPropertiesOfType(ctx.TypeChecker, t)) == 0 &&
				len(checker.Checker_getIndexInfosOfType(ctx.TypeChecker, t)) == 0 &&
				len(checker.Checker_getSignaturesOfType(ctx.TypeChecker, t, checker.SignatureKindCall)) == 0 &&
				len(checker.Checker_getSignaturesOfType(ctx.TypeChecker, t, checker.SignatureKindConstruct)) == 0 &&
				// Unresolved mapped types can have no members and accept number,
				// but reject string. Both primitives must be assignable to {}.
				checker.Checker_isTypeAssignableTo(ctx.TypeChecker, checker.Checker_numberType(ctx.TypeChecker), t) &&
				checker.Checker_isTypeAssignableTo(ctx.TypeChecker, checker.Checker_stringType(ctx.TypeChecker), t)
		}
		checkNode := func(node *ast.Node) {
			for _, t := range utils.UnionTypeParts(ctx.TypeChecker.GetTypeAtLocation(node)) {
				if isEmptyObjectType(t) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noGeneratedEmptyObjectType",
						Description: "This type resolves to `{}`, the empty object type. This was likely not intentional.",
					})
					return
				}
			}
		}
		return rule.RuleListeners{
			ast.KindIntersectionType: checkNode,
			ast.KindTypeReference: func(node *ast.Node) {
				parent := node.Parent
				for parent.Kind == ast.KindParenthesizedType {
					parent = parent.Parent
				}
				if node.AsTypeReferenceNode().TypeArguments != nil && parent.Kind != ast.KindIntersectionType {
					checkNode(node)
				}
			},
		}
	},
}
