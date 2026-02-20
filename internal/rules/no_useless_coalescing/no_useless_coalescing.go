package no_useless_coalescing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const noUselessCoalescingRuleName = "no-useless-coalescing"

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildUselessCoalescingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessCoalescing",
		Description: "This coalescing/defaulting operation is unnecessary and can be removed.",
	}
}

func buildRedundantUndefinedFallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redundantUndefinedFallback",
		Description: "Fallback to `undefined` is redundant for this expression.",
	}
}

func buildFalsyUndefinedNormalizationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "falsyUndefinedNormalization",
		Description: "This `|| undefined` fallback normalizes falsy primitive values and is likely unnecessary noise.",
	}
}

type orIdentityFallbackKind uint8

const (
	orIdentityFallbackNone orIdentityFallbackKind = iota
	orIdentityFallbackEmptyString
	orIdentityFallbackFalse
	orIdentityFallbackZeroBigInt
)

func isIndeterminateType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIndex|checker.TypeFlagsIndexedAccess) != 0
}

func hasIndeterminateConstituent(t *checker.Type) bool {
	var visit func(tp *checker.Type) bool
	visit = func(tp *checker.Type) bool {
		if tp == nil {
			return true
		}

		if isIndeterminateType(tp) {
			return true
		}

		if utils.IsUnionType(tp) || utils.IsIntersectionType(tp) {
			for _, part := range tp.Types() {
				if visit(part) {
					return true
				}
			}
		}

		return false
	}

	return visit(t)
}

func checkTypeCondition(t *checker.Type) (isTruthy bool, isFalsy bool) {
	flags := checker.Type_flags(t)

	if flags&checker.TypeFlagsNever != 0 {
		return false, true
	}

	if flags&checker.TypeFlagsIndexedAccess != 0 {
		return false, false
	}

	if utils.IsUnionType(t) {
		allTruthy := true
		allFalsy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(part)
			if !partTruthy {
				allTruthy = false
			}
			if !partFalsy {
				allFalsy = false
			}
		}

		return allTruthy, allFalsy
	}

	if utils.IsIntersectionType(t) {
		allTruthy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(part)
			if partFalsy {
				return false, true
			}
			if !partTruthy {
				allTruthy = false
			}
		}

		return allTruthy, false
	}

	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false, true
	}

	if flags&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive) != 0 {
		return true, false
	}

	if flags&(checker.TypeFlagsESSymbol|checker.TypeFlagsUniqueESSymbol) != 0 {
		return true, false
	}

	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		if utils.IsIntrinsicType(t) {
			intrinsicName := t.AsIntrinsicType().IntrinsicName()
			if intrinsicName == "true" {
				return true, false
			}
			if intrinsicName == "false" {
				return false, true
			}
		} else if t.AsLiteralType() != nil {
			litStr := t.AsLiteralType().String()
			if litStr == "true" {
				return true, false
			}
			if litStr == "false" {
				return false, true
			}
		}
	}

	if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if literal.Value() == "" {
				return false, true
			}
			return true, false
		}
	}

	if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			value := literal.String()
			if value == "0" || value == "NaN" {
				return false, true
			}
			return true, false
		}
	}

	if flags&checker.TypeFlagsBigIntLiteral != 0 && t.IsBigIntLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			value := literal.String()
			if value == "0" || value == "0n" {
				return false, true
			}
			return true, false
		}
	}

	return false, false
}

func typeCanBeNullish(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if checker.Type_flags(part)&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return true
		}
	}
	return false
}

func typeCanBeNull(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if checker.Type_flags(part)&checker.TypeFlagsNull != 0 {
			return true
		}
	}
	return false
}

func typeCanBeUndefined(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if checker.Type_flags(part)&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return true
		}
	}
	return false
}

func partCanBeNonNullishFalsy(part *checker.Type) bool {
	flags := checker.Type_flags(part)

	if flags&checker.TypeFlagsNever != 0 {
		return false
	}

	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false
	}

	truthy, falsy := checkTypeCondition(part)
	if truthy {
		return false
	}
	if falsy {
		return true
	}

	if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsEnumLike) != 0 {
		return true
	}

	return false
}

func typeHasNonNullishFalsyPotential(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if partCanBeNonNullishFalsy(part) {
			return true
		}
	}
	return false
}

func partHasPrimitiveFalsyCapability(part *checker.Type) bool {
	flags := checker.Type_flags(part)

	if flags&checker.TypeFlagsNever != 0 {
		return false
	}

	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false
	}

	if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike) == 0 {
		return false
	}

	truthy, _ := checkTypeCondition(part)
	return !truthy
}

func typeHasPrimitiveFalsyCapability(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if partHasPrimitiveFalsyCapability(part) {
			return true
		}
	}
	return false
}

func classifyOrIdentityFallback(rhs *ast.Node) orIdentityFallbackKind {
	rhs = ast.SkipParentheses(rhs)

	switch rhs.Kind {
	case ast.KindFalseKeyword:
		return orIdentityFallbackFalse
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if rhs.Text() == "" {
			return orIdentityFallbackEmptyString
		}
	case ast.KindBigIntLiteral:
		if rhs.Text() == "0" || rhs.Text() == "0n" {
			return orIdentityFallbackZeroBigInt
		}
	}

	return orIdentityFallbackNone
}

func typeIsIdentityNoopForOrFallback(t *checker.Type, fallback orIdentityFallbackKind) bool {
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)

		if flags&checker.TypeFlagsNever != 0 {
			continue
		}

		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return false
		}

		truthy, _ := checkTypeCondition(part)
		if truthy {
			continue
		}

		switch fallback {
		case orIdentityFallbackEmptyString:
			if flags&checker.TypeFlagsStringLike == 0 {
				return false
			}
		case orIdentityFallbackFalse:
			if flags&checker.TypeFlagsBooleanLike == 0 {
				return false
			}
		case orIdentityFallbackZeroBigInt:
			if flags&checker.TypeFlagsBigIntLike == 0 {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func isUndefinedLikeNode(node *ast.Node) bool {
	node = ast.SkipParentheses(node)

	if utils.IsUndefinedIdentifier(node) {
		return true
	}

	if !ast.IsVoidExpression(node) {
		return false
	}

	voidExpr := node.AsVoidExpression()
	if voidExpr == nil || voidExpr.Expression == nil {
		return false
	}

	operand := ast.SkipParentheses(voidExpr.Expression)
	if operand != nil && operand.Kind == ast.KindNumericLiteral {
		return operand.Text() == "0"
	}

	return false
}

func isOrUndefinedNoopByDefault(t *checker.Type) bool {
	if !typeCanBeUndefined(t) {
		return false
	}

	if typeCanBeNull(t) {
		return false
	}

	if typeHasNonNullishFalsyPotential(t) {
		return false
	}

	return true
}

func buildRemoveFallbackFix(ctx rule.RuleContext, binExpr *ast.BinaryExpression) rule.RuleFix {
	opRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, binExpr.OperatorToken.Pos())
	rightRange := utils.TrimNodeTextRange(ctx.SourceFile, binExpr.Right)

	start := opRange.Pos()
	minStart := binExpr.Left.End()
	sourceText := ctx.SourceFile.Text()
	for start > minStart {
		previousChar := sourceText[start-1]
		if previousChar != ' ' && previousChar != '\t' && previousChar != '\n' && previousChar != '\r' {
			break
		}
		start--
	}

	return rule.RuleFixRemoveRange(core.NewTextRange(start, rightRange.End()))
}

func getStableTypeForAnalysis(ctx rule.RuleContext, node *ast.Node) *checker.Type {
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	constraint, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, t)
	if isTypeParameter {
		if constraint == nil {
			return nil
		}
		t = constraint
	}

	if t == nil || hasIndeterminateConstituent(t) {
		return nil
	}

	return t
}

var NoUselessCoalescingRule = rule.Rule{
	Name: noUselessCoalescingRuleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoUselessCoalescingOptions](options, noUselessCoalescingRuleName)

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil || binExpr.OperatorToken == nil {
					return
				}

				opKind := binExpr.OperatorToken.Kind
				if opKind != ast.KindBarBarToken && opKind != ast.KindQuestionQuestionToken {
					return
				}

				leftType := getStableTypeForAnalysis(ctx, binExpr.Left)
				if leftType == nil {
					return
				}

				if opKind == ast.KindQuestionQuestionToken {
					if !typeCanBeNullish(leftType) {
						ctx.ReportNodeWithFixes(node, buildUselessCoalescingMessage(), func() []rule.RuleFix {
							return []rule.RuleFix{buildRemoveFallbackFix(ctx, binExpr)}
						})
						return
					}

					if isUndefinedLikeNode(binExpr.Right) && typeCanBeUndefined(leftType) && !typeCanBeNull(leftType) {
						ctx.ReportNodeWithFixes(node, buildRedundantUndefinedFallbackMessage(), func() []rule.RuleFix {
							return []rule.RuleFix{buildRemoveFallbackFix(ctx, binExpr)}
						})
					}
					return
				}

				leftAlwaysTruthy, _ := checkTypeCondition(leftType)
				if leftAlwaysTruthy {
					ctx.ReportNodeWithFixes(node, buildUselessCoalescingMessage(), func() []rule.RuleFix {
						return []rule.RuleFix{buildRemoveFallbackFix(ctx, binExpr)}
					})
					return
				}

				identityFallback := classifyOrIdentityFallback(binExpr.Right)
				if identityFallback != orIdentityFallbackNone && typeIsIdentityNoopForOrFallback(leftType, identityFallback) {
					ctx.ReportNodeWithFixes(node, buildUselessCoalescingMessage(), func() []rule.RuleFix {
						return []rule.RuleFix{buildRemoveFallbackFix(ctx, binExpr)}
					})
					return
				}

				if !isUndefinedLikeNode(binExpr.Right) {
					return
				}

				if isOrUndefinedNoopByDefault(leftType) {
					ctx.ReportNodeWithFixes(node, buildRedundantUndefinedFallbackMessage(), func() []rule.RuleFix {
						return []rule.RuleFix{buildRemoveFallbackFix(ctx, binExpr)}
					})
					return
				}

				if opts.DetectFalsyValues && typeCanBeNullish(leftType) && typeHasPrimitiveFalsyCapability(leftType) {
					ctx.ReportNode(node, buildFalsyUndefinedNormalizationMessage())
				}
			},
		}
	},
}
