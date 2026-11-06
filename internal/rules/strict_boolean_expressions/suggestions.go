package strict_boolean_expressions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func reportPrimitiveCondition(ctx rule.RuleContext, node *ast.Node, message rule.RuleMessage, variant typeVariant, suggest bool) {
	if !suggest {
		ctx.ReportNode(node, message)
		return
	}
	ctx.ReportNodeWithSuggestions(node, message, func() []rule.RuleSuggestion {
		return buildPrimitiveConditionSuggestions(ctx, node, variant)
	})
}

func buildPrimitiveConditionSuggestions(ctx rule.RuleContext, node *ast.Node, variant typeVariant) []rule.RuleSuggestion {
	target := node
	for ast.IsParenthesizedExpression(target.Parent) {
		target = target.Parent
	}
	negated := target.Parent != nil && target.Parent.Kind == ast.KindPrefixUnaryExpression && target.Parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	inner := node
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	code := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	if negated {
		target = target.Parent
		inner = target.AsPrefixUnaryExpression().Operand
		targetRange := utils.TrimNodeTextRange(ctx.SourceFile, target)
		code = strings.TrimSpace(ctx.SourceFile.Text()[targetRange.Pos()+1 : targetRange.End()])
	} else {
		target = node
	}
	if !utils.IsStrongPrecedenceNode(inner) {
		code = "(" + code + ")"
	}

	var suggestions []rule.RuleSuggestion
	add := func(id, description, replacement string) {
		s := scanner.NewScanner()
		s.SetText(replacement)
		firstToken := s.Scan()
		hasLeadingLineBreak := s.HasPrecedingLineBreak()
		statementStartToken := firstToken
		if firstToken == ast.KindAsyncKeyword && s.Scan() == ast.KindFunctionKeyword && !s.HasPrecedingLineBreak() {
			statementStartToken = ast.KindFunctionKeyword
		}
		if hasLeadingLineBreak || conditionNeedsParentheses(target) ||
			utils.IsStartOfExpressionStatementNeedingParentheses(ctx.SourceFile, target, statementStartToken) ||
			utils.IsStartOfArrowFunctionBodyNeedingParentheses(ctx.SourceFile, target, firstToken) {
			replacement = "(" + replacement + ")"
		}
		firstTokenPos := scanner.SkipTrivia(replacement, 0)
		if firstTokenPos < len(replacement) && strings.IndexByte("([`/", replacement[firstTokenPos]) >= 0 && conditionNeedsSemicolon(ctx.SourceFile, target) {
			replacement = ";" + replacement
		}
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message:  rule.RuleMessage{Id: id, Description: description},
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, target, replacement)},
		})
	}
	if variant == typeVariantString {
		lengthComparison := code + ".length > 0"
		emptyComparison := code + ` !== ""`
		if negated {
			lengthComparison = code + ".length === 0"
			emptyComparison = code + ` === ""`
		}
		add("conditionFixCompareStringLength", "Change condition to check string's length (`value.length !== 0`)", lengthComparison)
		add("conditionFixCompareEmptyString", "Change condition to check for empty string (`value !== \"\"`)", emptyComparison)
	} else {
		if isArrayLengthExpression(ctx, node) {
			if negated {
				add("conditionFixCompareArrayLengthZero", "Change condition to check array's length (`value.length === 0`)", code+" === 0")
			} else {
				add("conditionFixCompareArrayLengthNonzero", "Change condition to check array's length (`value.length > 0`)", code+" > 0")
			}
			return suggestions
		}
		zero := "0"
		if variant == typeVariantBigInt {
			zero = "0n"
		}
		comparison := code + " !== " + zero
		if negated {
			comparison = code + " === " + zero
		}
		add("conditionFixCompareZero", "Change condition to check for "+zero+" (`value !== "+zero+"`)", comparison)
		if variant != typeVariantBigInt {
			comparison = "!Number.isNaN(" + code + ")"
			if negated {
				comparison = "Number.isNaN(" + code + ")"
			}
			add("conditionFixCompareNaN", "Change condition to check for NaN (`!Number.isNaN(value)`)", comparison)
		}
	}
	conversion := "Boolean(" + code + ")"
	if negated {
		conversion = "!" + conversion
	}
	add("conditionFixCastBoolean", "Explicitly convert value to a boolean (`Boolean(value)`)", conversion)
	return suggestions
}

func isArrayLengthExpression(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsPropertyAccessExpression(node) || node.AsPropertyAccessExpression().Name().Text() != "length" {
		return false
	}
	t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.AsPropertyAccessExpression().Expression)
	for _, part := range utils.UnionTypeParts(t) {
		if !ctx.TypeChecker.IsArrayType(part) && !checker.IsTupleType(part) {
			return false
		}
	}
	return true
}

func conditionNeedsParentheses(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPostfixUnaryExpression, ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression,
		ast.KindBinaryExpression, ast.KindConditionalExpression, ast.KindAwaitExpression,
		ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return true
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression:
		return parent.Expression() == node
	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}

func conditionNeedsSemicolon(sourceFile *ast.SourceFile, node *ast.Node) bool {
	start := utils.TrimNodeTextRange(sourceFile, node).Pos()
	for current := node; current != nil && utils.TrimNodeTextRange(sourceFile, current).Pos() == start; current = current.Parent {
		if !ast.IsExpressionStatement(current) || current.Parent == nil {
			continue
		}
		if !current.Parent.CanHaveStatements() {
			return false
		}
		statements := current.Parent.Statements()
		for i, statement := range statements {
			if statement == current && i > 0 {
				previous := utils.TrimNodeTextRange(sourceFile, statements[i-1])
				return sourceFile.Text()[previous.End()-1] != ';'
			}
		}
		return false
	}
	return false
}
