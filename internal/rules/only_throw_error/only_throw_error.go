package only_throw_error

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type OnlyThrowErrorOptions struct {
	Allow                []utils.TypeOrValueSpecifier
	AllowInline          []string
	AllowThrowingAny     *bool
	AllowThrowingUnknown *bool
}

func buildObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object",
		Description: "Expected an error object to be thrown.",
	}
}
func buildUndefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "undef",
		Description: "Do not throw undefined.",
	}
}

var OnlyThrowErrorRule = rule.Rule{
	Name: "only-throw-error",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(OnlyThrowErrorOptions)
		if !ok {
			opts = OnlyThrowErrorOptions{}
		}
		if opts.Allow == nil {
			opts.Allow = []utils.TypeOrValueSpecifier{}
		}
		if opts.AllowInline == nil {
			opts.AllowInline = []string{}
		}
		if opts.AllowThrowingAny == nil {
			opts.AllowThrowingAny = utils.Ref(true)
		}
		if opts.AllowThrowingUnknown == nil {
			opts.AllowThrowingUnknown = utils.Ref(true)
		}

		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				expr := node.Expression()
				t := ctx.TypeChecker.GetTypeAtLocation(expr)

				if utils.TypeMatchesSomeSpecifier(t, opts.Allow, opts.AllowInline, ctx.Program) {
					return
				}

				if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					ctx.ReportNode(node, buildUndefMessage())
					return
				}

				if *opts.AllowThrowingAny && utils.IsTypeAnyType(t) {
					return
				}

				if *opts.AllowThrowingUnknown && utils.IsTypeUnknownType(t) {
					return
				}

				if utils.IsErrorLike(ctx.Program, ctx.TypeChecker, t) {
					return
				}

				ctx.ReportNode(expr, buildObjectMessage())
			},
		}
	},
}
