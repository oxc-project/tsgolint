package no_unnecessary_type_assertion

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// typescript-go represents JSDoc casts as assertions. Locate the cast's comment
// so it can be reported and removed.
func jsDocAssertionRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, bool) {
	if !ast.IsInJSFile(node) || node.Type() == nil || node.Type().Pos() >= node.Expression().Pos() {
		return core.TextRange{}, false
	}
	searchStart := node.Pos()
	if ast.IsParenthesizedExpression(node.Parent) {
		searchStart = node.Parent.Pos()
	}
	beforeExpression := ctx.SourceFile.Text()[searchStart:node.Expression().Pos()]
	start := strings.LastIndex(beforeExpression, "/**")
	end := strings.LastIndex(beforeExpression, "*/")
	if start < 0 || end < start {
		return core.TextRange{}, false
	}
	return core.NewTextRange(searchStart+start, searchStart+end+len("*/")), true
}

func assertionRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	if assertion, ok := jsDocAssertionRange(ctx, node); ok {
		return assertion
	}
	// Report the whole assertion, excluding surrounding trivia.
	return utils.TrimNodeTextRange(ctx.SourceFile, node)
}

// Remove the assertion while preserving valid expression syntax.
func createAssertionFixer(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	if assertion, ok := jsDocAssertionRange(ctx, node); ok {
		end := assertion.End()
		for end < node.Expression().Pos() && utils.IsStrWhiteSpace(rune(ctx.SourceFile.Text()[end])) {
			end++
		}
		return []rule.RuleFix{rule.RuleFixRemoveRange(assertion.WithEnd(end))}
	}
	typeNode := node.Type()
	if ast.IsTypeAssertion(node) {
		s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
		openingAngleBracket := s.TokenRange()
		s.ResetPos(typeNode.End())
		s.Scan()
		closingAngleBracket := s.TokenRange()
		s.Scan()
		firstOperandToken := s.Token()
		statementStartToken := firstOperandToken
		// At the start of a statement, async function needs parentheses to remain
		// an expression after the assertion is removed.
		if firstOperandToken == ast.KindAsyncKeyword && s.Scan() == ast.KindFunctionKeyword && !s.HasPrecedingLineBreak() {
			statementStartToken = ast.KindFunctionKeyword
		}
		needsParens := isStartOfExpressionStatementNeedingParentheses(ctx, node, statementStartToken) ||
			isStartOfArrowFunctionBodyNeedingParentheses(ctx, node, firstOperandToken)

		var fixes []rule.RuleFix
		if needsParens {
			fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, node, "("))
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(openingAngleBracket.WithEnd(closingAngleBracket.End())))
		if needsParens {
			fixes = append(fixes, rule.RuleFixInsertAfter(node, ")"))
		}
		return fixes
	}

	// Preserve the token or comment before `as` when removing the assertion.
	expression := node.Expression()
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
	asToken := s.TokenRange()
	tokenBeforeAsEnd := expression.End()
	for comment := range utils.GetCommentsInRange(ctx.SourceFile, core.NewTextRange(expression.End(), asToken.Pos())) {
		tokenBeforeAsEnd = max(tokenBeforeAsEnd, comment.End())
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(tokenBeforeAsEnd, node.End()))}
}

// Matching start positions limits traversal to the beginning of a statement
// and stops at existing parentheses.
func isStartOfExpressionStatementNeedingParentheses(ctx rule.RuleContext, node *ast.Node, firstToken ast.Kind) bool {
	if firstToken != ast.KindOpenBraceToken && firstToken != ast.KindClassKeyword && firstToken != ast.KindFunctionKeyword {
		return false
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	for ancestor := node.Parent; ancestor != nil && utils.TrimNodeTextRange(ctx.SourceFile, ancestor).Pos() == start; ancestor = ancestor.Parent {
		if ast.IsExpressionStatement(ancestor) {
			return true
		}
	}
	return false
}

func isStartOfArrowFunctionBodyNeedingParentheses(ctx rule.RuleContext, node *ast.Node, firstToken ast.Kind) bool {
	if firstToken != ast.KindOpenBraceToken {
		return false
	}
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			return false
		}
		if ast.IsArrowFunction(parent) && parent.Body() == current {
			return true
		}
		if utils.TrimNodeTextRange(ctx.SourceFile, parent).Pos() != utils.TrimNodeTextRange(ctx.SourceFile, current).Pos() {
			return false
		}
	}
	return false
}
