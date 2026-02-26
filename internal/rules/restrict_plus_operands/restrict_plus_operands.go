package restrict_plus_operands

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildBigintAndNumberDiagnostic(exprRange core.TextRange, leftRange core.TextRange, rightRange core.TextRange, leftType string, rightType string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: exprRange,
		Message: rule.RuleMessage{
			Id:          "bigintAndNumber",
			Description: "Numeric '+' operations must either be both bigints or both numbers.",
		},
		LabeledRanges: []rule.RuleLabeledRange{
			{Label: fmt.Sprintf("Type: %v", leftType), Range: leftRange},
			{Label: fmt.Sprintf("Type: %v", rightType), Range: rightRange},
		},
	}
}

func buildInvalidDiagnostic(exprRange core.TextRange, invalidType string, stringLike string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: exprRange,
		Message: rule.RuleMessage{
			Id:          "invalid",
			Description: fmt.Sprintf("Invalid operand of type '%v' for a '+' operation.", invalidType),
			Help:        fmt.Sprintf("Operands must each be a number or %v.", stringLike),
		},
	}
}

func buildMismatchedDiagnostic(exprRange core.TextRange, leftRange core.TextRange, rightRange core.TextRange, stringLike string, leftType string, rightType string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range: exprRange,
		Message: rule.RuleMessage{
			Id:          "mismatched",
			Description: "Operands of '+' operations must be of the same type.",
			Help:        fmt.Sprintf("Operands must both be a number or both be %v.", stringLike),
		},
		LabeledRanges: []rule.RuleLabeledRange{
			{Label: fmt.Sprintf("Type: %v", leftType), Range: leftRange},
			{Label: fmt.Sprintf("Type: %v", rightType), Range: rightRange},
		},
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

		checkInvalidPlusOperand := func(baseType, otherType *checker.Type) (checker.TypeFlags, string, bool) {
			foundRegexp := false

			var flags checker.TypeFlags
			baseTypeString := ctx.TypeChecker.TypeToString(baseType)

			for _, part := range utils.UnionTypeParts(baseType) {
				flags |= checker.Type_flags(part)
				if utils.IsTypeFlagSet(part, invalidFlags) {
					return flags, baseTypeString, true
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

			if foundRegexp {
				return flags, ctx.TypeChecker.TypeToString(globalRegexpType), true
			}

			return flags, "", false
		}

		checkPlusOperands := func(
			node *ast.BinaryExpression,
		) {
			leftType := getTypeConstrained(node.Left)
			rightType := getTypeConstrained(node.Right)
			leftTypeString := ctx.TypeChecker.TypeToString(leftType)
			rightTypeString := ctx.TypeChecker.TypeToString(rightType)
			leftRange := utils.TrimNodeTextRange(ctx.SourceFile, node.Left)
			rightRange := utils.TrimNodeTextRange(ctx.SourceFile, node.Right)

			if leftType == rightType &&
				utils.IsTypeFlagSet(
					leftType,
					checker.TypeFlagsBigIntLike|
						checker.TypeFlagsNumberLike|
						checker.TypeFlagsStringLike,
				) {
				return
			}

			leftTypeFlags, leftInvalidType, leftInvalid := checkInvalidPlusOperand(leftType, rightType)
			rightTypeFlags, rightInvalidType, rightInvalid := checkInvalidPlusOperand(rightType, leftType)

			if leftInvalid {
				ctx.ReportDiagnostic(buildInvalidDiagnostic(leftRange, leftInvalidType, stringLike))
			}
			if rightInvalid {
				ctx.ReportDiagnostic(buildInvalidDiagnostic(rightRange, rightInvalidType, stringLike))
			}
			if leftInvalid || rightInvalid {
				return
			}

			checkMismatchedPlusOperands := func(baseTypeFlags, otherTypeFlags checker.TypeFlags) bool {
				if !opts.AllowNumberAndString &&
					baseTypeFlags&checker.TypeFlagsStringLike != 0 &&
					otherTypeFlags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) != 0 {
					ctx.ReportDiagnostic(buildMismatchedDiagnostic(utils.TrimNodeTextRange(ctx.SourceFile, &node.Node), leftRange, rightRange, stringLike, leftTypeString, rightTypeString))
					return true
				}

				if baseTypeFlags&checker.TypeFlagsNumberLike != 0 && otherTypeFlags&checker.TypeFlagsBigIntLike != 0 {
					ctx.ReportDiagnostic(buildBigintAndNumberDiagnostic(utils.TrimNodeTextRange(ctx.SourceFile, &node.Node), leftRange, rightRange, leftTypeString, rightTypeString))
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
