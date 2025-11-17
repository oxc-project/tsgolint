package no_meaningless_void_operator

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMeaninglessVoidOperatorMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "meaninglessVoidOperator",
		Description: fmt.Sprintf("void operator shouldn't be used on %v; it should convey that a return value is being ignored", t),
	}
}
func buildRemoveVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeVoid",
		Description: "Remove 'void'",
	}
}

var NoMeaninglessVoidOperatorRule = rule.Rule{
	Name: "no-meaningless-void-operator",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoMeaninglessVoidOperatorOptions](options, "no-meaningless-void-operator")

		return rule.RuleListeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				arg := node.AsVoidExpression().Expression
				argType := ctx.TypeChecker.GetTypeAtLocation(arg)

				mask := checker.TypeFlagsVoidLike | checker.TypeFlagsNever

				for _, t := range utils.UnionTypeParts(argType) {
					mask &= checker.Type_flags(t)
				}

				fixRemoveVoidKeyword := func() rule.RuleFix {
					return rule.RuleFixRemoveRange(utils.TrimNodeTextRange(ctx.SourceFile, node).WithEnd(arg.Pos()))
				}

				if mask&checker.TypeFlagsVoidLike != 0 {
					ctx.ReportNodeWithFixes(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), func() []rule.RuleFix { return []rule.RuleFix{fixRemoveVoidKeyword()} })
				} else if opts.CheckNever && mask&checker.TypeFlagsNever != 0 {
					ctx.ReportNodeWithSuggestions(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildRemoveVoidMessage(),
							FixesArr: []rule.RuleFix{fixRemoveVoidKeyword()},
						}}
					})
				}
			},
		}
	},
}
