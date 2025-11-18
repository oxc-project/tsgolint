package restrict_template_expressions

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildInvalidTypeMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidType",
		Description: fmt.Sprintf("Invalid type \"%v\" of template literal expression.", t),
	}
}

var RestrictTemplateExpressionsRule = rule.Rule{
	Name: "restrict-template-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[RestrictTemplateExpressionsOptions](options, "restrict-template-expressions")

		allowedFlags := checker.TypeFlagsStringLike
		if opts.AllowAny {
			allowedFlags |= checker.TypeFlagsAny
		}
		if opts.AllowBoolean {
			allowedFlags |= checker.TypeFlagsBooleanLike
		}
		if opts.AllowNullish {
			allowedFlags |= checker.TypeFlagsNullable
		}
		if opts.AllowNumber {
			allowedFlags |= checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike
		}
		if opts.AllowNever {
			allowedFlags |= checker.TypeFlagsNever
		}

		globalRegexpType := checker.Checker_globalRegExpType(ctx.TypeChecker)

		var isTypeAllowed func(innerType *checker.Type) bool
		isTypeAllowed = func(innerType *checker.Type) bool {
			return utils.Every(utils.UnionTypeParts(innerType), func(t *checker.Type) bool {
				return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, allowedFlags) ||
						utils.TypeMatchesSomeSpecifier(t, opts.Allow, ctx.Program) ||
						(opts.AllowArray && checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) && isTypeAllowed(utils.GetNumberIndexType(ctx.TypeChecker, t))) ||
						(opts.AllowRegExp && t == globalRegexpType)
				})
			})
		}

		return rule.RuleListeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				// don't check tagged template literals
				if ast.IsTaggedTemplateExpression(node.Parent) {
					return
				}

				for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
					expression := span.Expression()
					expressionType := utils.GetConstrainedTypeAtLocation(
						ctx.TypeChecker,
						expression,
					)
					if !isTypeAllowed(expressionType) {
						ctx.ReportNode(expression, buildInvalidTypeMessage(ctx.TypeChecker.TypeToString(expressionType)))
					}
				}
			},
		}
	},
}
