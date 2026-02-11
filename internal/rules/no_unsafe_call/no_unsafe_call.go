package no_unsafe_call

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUnsafeCallMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeCall",
		Description: fmt.Sprintf("Unsafe call of a(n) %v typed value.", t),
	}
}
func buildUnsafeCallThisMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeCallThis",
		Description: fmt.Sprintf("Unsafe call of a(n) %v typed value. `this` is typed as %v.\n", t, t),
		Help:        "You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}
func buildUnsafeNewMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeNew",
		Description: fmt.Sprintf("Unsafe construction of a(n) %v typed value.", t),
	}
}
func buildUnsafeTemplateTagMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTemplateTag",
		Description: fmt.Sprintf("Unsafe use of a(n) %v typed template tag.", t),
	}
}

var NoUnsafeCallRule = rule.Rule{
	Name: "no-unsafe-call",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		isNoImplicitThis := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.NoImplicitThis,
		)

		checkCall := func(
			node *ast.Node,
			reportingNode *ast.Node,
			messageBuilder func(t string) rule.RuleMessage,
			newCall bool,
		) {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)

			if utils.IsTypeAnyType(t) {
				if !isNoImplicitThis {
					// `this()` or `this.foo()` or `this.foo[bar]()`
					thisExpression := utils.GetThisExpression(node)
					if thisExpression != nil && utils.IsTypeAnyType(
						utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression),
					) {
						messageBuilder = buildUnsafeCallThisMessage
					}
				}

				isErrorType := utils.IsIntrinsicErrorType(t)

				msg := "`any`"
				if isErrorType {
					msg = "`error` type"
				}
				ctx.ReportNode(reportingNode, messageBuilder(msg))
				return
			}

			if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "Function") {
				// this also matches subtypes of `Function`, like `interface Foo extends Function {}`.
				//
				// For weird TS reasons that I don't understand, these are
				//
				// safe to construct if:
				// - they have at least one call signature _that is not void-returning_,
				// - OR they have at least one construct signature.
				//
				// safe to call (including as template) if:
				// - they have at least one call signature
				// - OR they have at least one construct signature.

				constructSignatures := utils.GetConstructSignatures(ctx.TypeChecker, t)
				if len(constructSignatures) > 0 {
					return
				}

				callSignatures := utils.GetCallSignatures(ctx.TypeChecker, t)
				if newCall {
					if utils.Some(callSignatures, func(signature *checker.Signature) bool {
						return !utils.IsIntrinsicVoidType(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
					}) {
						return
					}
				} else if len(callSignatures) > 0 {
					return
				}

				ctx.ReportNode(reportingNode, messageBuilder("`Function`"))
				return
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.Expression()
				if callee.Kind == ast.KindImportKeyword {
					return
				}
				checkCall(callee, callee, buildUnsafeCallMessage, false)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				callee := node.Expression()
				checkCall(callee, node, buildUnsafeNewMessage, true)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tag := node.AsTaggedTemplateExpression().Tag
				checkCall(tag, tag, buildUnsafeTemplateTagMessage, false)
			},
		}
	},
}
