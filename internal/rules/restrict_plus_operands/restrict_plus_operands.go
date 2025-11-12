package restrict_plus_operands

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildBigintAndNumberMessage(left, right string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bigintAndNumber",
		Description: fmt.Sprintf("Numeric '+' operations must either be both bigints or both numbers. Got `%v` + `%v`.", left, right),
	}
}
func buildInvalidMessage(stringLike, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalid",
		Description: fmt.Sprintf("Invalid operand for a '+' operation. Operands must each be a number or %v. Got `%v`.", stringLike, t),
	}
}
func buildMismatchedMessage(stringLike, left, right string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatched",
		Description: fmt.Sprintf("Operands of '+' operations must be a number or %v. Got `%v` + `%v`.", stringLike, left, right),
	}
}

var RestrictPlusOperandsRule = rule.Rule{
	Name: "restrict-plus-operands",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[RestrictPlusOperandsOptions](options, "restrict-plus-operands")

		stringLikes := make([]string, 0, 5)
		if opts.AllowAny {
			stringLikes = append(stringLikes, "`any`")
		}
		if opts.AllowBoolean {
			stringLikes = append(stringLikes, "`boolean`")
		}
		if opts.AllowNullish {
			stringLikes = append(stringLikes, "`null`")
		}
		if opts.AllowRegExp {
			stringLikes = append(stringLikes, "`RegExp`")
		}
		if opts.AllowNullish {
			stringLikes = append(stringLikes, "`undefined`")
		}
		var stringLike string
		switch len(stringLikes) {
		case 0:
			stringLike = "string"
		case 1:
			stringLike = "string, allowing a string + " + stringLikes[0]
		default:
			stringLike = "string, allowing a string + any of: " + strings.Join(stringLikes, ", ")
		}

		getTypeConstrained := func(node *ast.Node) *checker.Type {
			return checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node))
		}

		globalRegexpType := checker.Checker_globalRegExpType(ctx.TypeChecker)

		invalidFlags := checker.TypeFlagsESSymbolLike |
			checker.TypeFlagsNever |
			checker.TypeFlagsUnknown
		if !opts.AllowAny {
			invalidFlags |= checker.TypeFlagsAny
		}
		if !opts.AllowBoolean {
			invalidFlags |= checker.TypeFlagsBooleanLike
		}
		if !opts.AllowNullish {
			invalidFlags |= checker.TypeFlagsNullable
		}

		checkInvalidPlusOperand := func(baseNode *ast.Node, baseType, otherType *checker.Type) (checker.TypeFlags, bool) {
			foundRegexp := false

			var flags checker.TypeFlags

			reported := false
			for _, part := range utils.UnionTypeParts(baseType) {
				flags |= checker.Type_flags(part)
				if reported {
					continue
				}
				if utils.IsTypeFlagSet(part, invalidFlags) {
					ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, ctx.TypeChecker.TypeToString(baseType)))
					reported = true
					continue
				}

				// RegExps also contain checker.TypeFlagsAny & checker.TypeFlagsObject
				if part == globalRegexpType {
					if opts.AllowRegExp && !utils.IsTypeFlagSet(otherType, checker.TypeFlagsNumberLike) {
						continue
					}
				} else if (opts.AllowAny || !utils.IsTypeAnyType(part)) && !utils.Every(utils.IntersectionTypeParts(part), utils.IsObjectType) {
					continue
				}
				foundRegexp = true
			}

			if !reported && foundRegexp {
				ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, ctx.TypeChecker.TypeToString(globalRegexpType)))
				return flags, true
			}

			return flags, reported
		}

		checkPlusOperands := func(
			node *ast.BinaryExpression,
		) {
			leftType := getTypeConstrained(node.Left)
			rightType := getTypeConstrained(node.Right)

			if leftType == rightType &&
				utils.IsTypeFlagSet(
					leftType,
					checker.TypeFlagsBigIntLike|
						checker.TypeFlagsNumberLike|
						checker.TypeFlagsStringLike,
				) {
				return
			}

			leftTypeFlags, leftInvalid := checkInvalidPlusOperand(node.Left, leftType, rightType)
			rightTypeFlags, rightInvalid := checkInvalidPlusOperand(node.Right, rightType, leftType)
			if leftInvalid || rightInvalid {
				return
			}

			checkMismatchedPlusOperands := func(baseTypeFlags, otherTypeFlags checker.TypeFlags) bool {
				if !opts.AllowNumberAndString &&
					baseTypeFlags&checker.TypeFlagsStringLike != 0 &&
					otherTypeFlags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) != 0 {
					ctx.ReportNode(&node.Node, buildMismatchedMessage(stringLike, ctx.TypeChecker.TypeToString(leftType), ctx.TypeChecker.TypeToString(rightType)))
					return true
				}

				if baseTypeFlags&checker.TypeFlagsNumberLike != 0 && otherTypeFlags&checker.TypeFlagsBigIntLike != 0 {
					ctx.ReportNode(&node.Node, buildBigintAndNumberMessage(ctx.TypeChecker.TypeToString(leftType), ctx.TypeChecker.TypeToString(rightType)))
					return true
				}

				return false
			}

			if checkMismatchedPlusOperands(leftTypeFlags, rightTypeFlags) {
				return
			}
			checkMismatchedPlusOperands(rightTypeFlags, leftTypeFlags)
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				if expr.OperatorToken.Kind == ast.KindPlusToken || (!opts.SkipCompoundAssignments && expr.OperatorToken.Kind == ast.KindPlusEqualsToken) {
					checkPlusOperands(expr)
				}
			},
		}
	},
}
