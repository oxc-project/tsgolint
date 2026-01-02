package jsx_no_leaked_render

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildLeakedRenderMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noLeakedConditionalRendering",
		Description: "Potential leaked value that might cause unintentionally rendered values or rendering crashes.",
	}
}

// checkLeakyType returns true if the type could leak a falsy value to JSX
func checkLeakyType(t *checker.Type) bool {
	// any type is always problematic
	if utils.IsTypeFlagSet(t, checker.TypeFlagsAny) {
		return true
	}

	flags := checker.Type_flags(t)

	// Check for number types
	if flags&checker.TypeFlagsNumberLike != 0 {
		// If it's a number literal, check if it's 0 (falsy)
		if t.IsNumberLiteral() {
			literal := t.AsLiteralType()
			if literal != nil && literal.String() != "0" {
				// Non-zero number literal is safe (truthy)
				return false
			}
			// 0 is falsy and can leak
			return true
		}
		// Generic number type - could be 0, so it's leaky
		return true
	}

	// Check for bigint types (0n is also falsy and renders "0")
	if flags&checker.TypeFlagsBigIntLike != 0 {
		// If it's a bigint literal, check if it's 0n (falsy)
		if t.IsBigIntLiteral() {
			literal := t.AsLiteralType()
			if literal != nil && literal.String() != "0" {
				// Non-zero bigint literal is safe (truthy)
				return false
			}
			// 0n is falsy and can leak
			return true
		}
		// Generic bigint type - could be 0n, so it's leaky
		return true
	}

	// Other types (string, boolean, object, null, undefined) are safe
	// - string: empty string doesn't render visibly
	// - boolean: false doesn't render
	// - object: always truthy
	// - null/undefined: don't render
	return false
}

var JsxNoLeakedRenderRule = rule.Rule{
	Name: "jsx-no-leaked-render",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()

				// Only check && operator
				if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
					return
				}

				// Check if inside JSX expression container
				parent := node.Parent
				if parent == nil || !ast.IsJsxExpression(parent) {
					return
				}

				// Get type of left operand (resolves generics via GetBaseConstraintOfType)
				leftType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, binary.Left)

				// Check if type could leak (including union members)
				if utils.TypeRecurser(leftType, func(t *checker.Type) bool {
					return checkLeakyType(t)
				}) {
					ctx.ReportNode(binary.Left, buildLeakedRenderMessage())
				}
			},
		}
	},
}
