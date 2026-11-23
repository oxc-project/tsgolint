package no_meaningless_void_operator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
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

func removeVoidKeyword(ctx rule.RuleContext, node, arg *ast.Node) []rule.RuleFix {
	keyword := utils.TrimNodeTextRange(ctx.SourceFile, node).WithEnd(arg.Pos())
	statement := node.Parent
	if !ast.IsExpressionStatement(statement) {
		return []rule.RuleFix{rule.RuleFixRemoveRange(keyword)}
	}

	s := scanner.GetScannerForSourceFile(ctx.SourceFile, arg.Pos())
	firstToken := s.Token()
	if firstToken == ast.KindAsyncKeyword && s.Scan() == ast.KindFunctionKeyword && !s.HasPrecedingLineBreak() {
		firstToken = ast.KindFunctionKeyword
	}
	wrap := ast.IsStringLiteral(arg) || firstToken == ast.KindAtToken ||
		(firstToken == ast.KindLetKeyword && s.Scan() == ast.KindOpenBracketToken) ||
		utils.IsStartOfExpressionStatementNeedingParentheses(ctx.SourceFile, node, firstToken)
	if wrap {
		firstToken = ast.KindOpenParenToken
	}

	prefix := ""
	switch firstToken {
	case ast.KindOpenParenToken, ast.KindOpenBracketToken, ast.KindPlusToken, ast.KindMinusToken,
		ast.KindSlashToken, ast.KindSlashEqualsToken, ast.KindLessThanToken,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead:
		if previousStatementLacksSemicolon(ctx.SourceFile, statement) {
			prefix = ";"
		}
	}
	if wrap {
		return []rule.RuleFix{rule.RuleFixReplaceRange(keyword, prefix+"("), rule.RuleFixInsertAfter(node, ")")}
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(keyword, prefix)}
}

func previousStatementLacksSemicolon(sourceFile *ast.SourceFile, statement *ast.Node) bool {
	if !statement.Parent.CanHaveStatements() {
		return false
	}
	statements := statement.Parent.Statements()
	i := sort.Search(len(statements), func(i int) bool { return statements[i].Pos() >= statement.Pos() })
	return i > 0 && !strings.HasSuffix(sourceFile.Text()[:statements[i-1].End()], ";")
}

var NoMeaninglessVoidOperatorRule = rule.Rule{
	Name: "no-meaningless-void-operator",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoMeaninglessVoidOperatorOptions](options, "no-meaningless-void-operator")

		return rule.RuleListeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				arg := node.AsVoidExpression().Expression
				inner := ast.SkipOuterExpressions(arg, ast.OEKParentheses|ast.OEKAssertions|ast.OEKComma)
				if ast.IsAssignmentExpression(inner, false) || (ast.IsNumericLiteral(inner) && inner.AsNumericLiteral().Text == "0") {
					return
				}
				argType := ctx.TypeChecker.GetTypeAtLocation(arg)
				fixRemoveVoidKeyword := func() []rule.RuleFix {
					return removeVoidKeyword(ctx, node, arg)
				}

				if !ast.IsCallExpression(inner) || ast.IsImportCall(inner) {
					if utils.IsThenableType(ctx.TypeChecker, arg, argType) {
						return
					}
					message := rule.RuleMessage{
						Id:          "meaninglessVoidOnNonCall",
						Description: "void operator is useless here; it should only discard a call's return value",
					}
					parent := node.Parent
					for ast.IsParenthesizedExpression(parent) {
						parent = parent.Parent
					}
					if ast.IsExpressionStatement(parent) {
						ctx.ReportNodeWithFixes(node, message, fixRemoveVoidKeyword)
					} else {
						ctx.ReportNode(node, message)
					}
					return
				}

				unionParts := utils.UnionTypeParts(argType)

				isAlwaysVoidLike := utils.Every(unionParts, func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike)
				})
				isAlwaysVoidLikeOrNever := utils.Every(unionParts, func(t *checker.Type) bool {
					return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike|checker.TypeFlagsNever)
				})

				if isAlwaysVoidLike {
					ctx.ReportNodeWithFixes(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), fixRemoveVoidKeyword)
				} else if opts.CheckNever && isAlwaysVoidLikeOrNever {
					ctx.ReportNodeWithSuggestions(node, buildMeaninglessVoidOperatorMessage(ctx.TypeChecker.TypeToString(argType)), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildRemoveVoidMessage(),
							FixesArr: fixRemoveVoidKeyword(),
						}}
					})
				}
			},
		}
	},
}
