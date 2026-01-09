package prefer_find

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferFindMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFind",
		Description: "Prefer .find(...) instead of .filter(...)[0].",
	}
}

func buildPreferFindSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFindSuggestion",
		Description: "Use .find(...) instead of .filter(...)[0].",
	}
}

type filterExpressionData struct {
	filterNode               *ast.Node
	isBracketSyntaxForFilter bool
}

var PreferFindRule = rule.Rule{
	Name: "prefer-find",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		isArrayish := func(t *checker.Type) bool {
			isAtLeastOneArrayishComponent := false

			for _, unionPart := range utils.UnionTypeParts(t) {
				if utils.IsTypeFlagSet(unionPart, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
					continue
				}

				intersectionParts := utils.IntersectionTypeParts(unionPart)
				isArrayOrIntersectionThereof := utils.Every(intersectionParts, func(intersectionPart *checker.Type) bool {
					return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, intersectionPart)
				})

				if !isArrayOrIntersectionThereof {
					return false
				}

				isAtLeastOneArrayishComponent = true
			}

			return isAtLeastOneArrayishComponent
		}

		var getStaticStringValue func(node *ast.Node) (string, bool)
		getStaticStringValue = func(node *ast.Node) (string, bool) {
			node = ast.SkipParentheses(node)

			switch node.Kind {
			case ast.KindStringLiteral:
				return node.AsStringLiteral().Text, true

			case ast.KindNoSubstitutionTemplateLiteral:
				return node.AsNoSubstitutionTemplateLiteral().Text, true

			case ast.KindNumericLiteral:
				return node.AsNumericLiteral().Text, true

			case ast.KindBigIntLiteral:
				return strings.TrimSuffix(node.AsBigIntLiteral().Text, "n"), true

			case ast.KindPrefixUnaryExpression:
				prefixExpr := node.AsPrefixUnaryExpression()
				if prefixExpr.Operator == ast.KindMinusToken {
					if prefixExpr.Operand.Kind == ast.KindNumericLiteral {
						return "-" + prefixExpr.Operand.AsNumericLiteral().Text, true
					}
					if prefixExpr.Operand.Kind == ast.KindBigIntLiteral {
						text := strings.TrimSuffix(prefixExpr.Operand.AsBigIntLiteral().Text, "n")
						return "-" + text, true
					}
				}
				return "", false

			case ast.KindIdentifier:
				if node.AsIdentifier().Text == "NaN" {
					return "NaN", true
				}
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil || symbol.ValueDeclaration == nil {
					return "", false
				}
				if symbol.ValueDeclaration.Kind == ast.KindVariableDeclaration {
					varDecl := symbol.ValueDeclaration.AsVariableDeclaration()
					if varDecl.Initializer != nil {
						return getStaticStringValue(varDecl.Initializer)
					}
				}
				return "", false

			case ast.KindTemplateExpression:
				templateExpr := node.AsTemplateExpression()
				if len(templateExpr.TemplateSpans.Nodes) == 0 {
					return templateExpr.Head.AsTemplateHead().Text, true
				}
				var builder strings.Builder
				builder.WriteString(templateExpr.Head.AsTemplateHead().Text)
				for _, spanNode := range templateExpr.TemplateSpans.Nodes {
					span := spanNode.AsTemplateSpan()
					spanValue, ok := getStaticStringValue(span.Expression)
					if !ok {
						return "", false
					}
					builder.WriteString(spanValue)
					switch span.Literal.Kind {
					case ast.KindTemplateMiddle:
						builder.WriteString(span.Literal.AsTemplateMiddle().Text)
					case ast.KindTemplateTail:
						builder.WriteString(span.Literal.AsTemplateTail().Text)
					}
				}
				return builder.String(), true
			}

			return "", false
		}

		var parseArrayFilterExpressions func(expression *ast.Node) []filterExpressionData
		parseArrayFilterExpressions = func(expression *ast.Node) []filterExpressionData {
			node := ast.SkipParentheses(expression)

			if ast.IsBinaryExpression(node) && node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
				return parseArrayFilterExpressions(node.AsBinaryExpression().Right)
			}

			if node.Kind == ast.KindConditionalExpression {
				condExpr := node.AsConditionalExpression()

				consequentResult := parseArrayFilterExpressions(condExpr.WhenTrue)
				if len(consequentResult) == 0 {
					return nil
				}

				alternateResult := parseArrayFilterExpressions(condExpr.WhenFalse)
				if len(alternateResult) == 0 {
					return nil
				}

				return append(consequentResult, alternateResult...)
			}

			if node.Kind == ast.KindCallExpression {
				callExpr := node.AsCallExpression()

				if callExpr.QuestionDotToken != nil {
					return nil
				}

				callee := callExpr.Expression

				if !ast.IsAccessExpression(callee) {
					return nil
				}

				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !found || propertyName != "filter" {
					return nil
				}

				isBracketSyntax := callee.Kind == ast.KindElementAccessExpression

				var filterNode *ast.Node
				switch callee.Kind {
				case ast.KindPropertyAccessExpression:
					filterNode = callee.AsPropertyAccessExpression().Name()
				case ast.KindElementAccessExpression:
					filterNode = callee.AsElementAccessExpression().ArgumentExpression
				default:
					return nil
				}

				filteredObjectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				if isArrayish(filteredObjectType) {
					return []filterExpressionData{
						{
							filterNode:               filterNode,
							isBracketSyntaxForFilter: isBracketSyntax,
						},
					}
				}
			}

			return nil
		}

		isTreatedAsZeroByMemberAccess := func(value string) bool {
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return value == "0"
			}
			return num == 0
		}

		isTreatedAsZeroByArrayAt := func(value string) bool {
			if value == "NaN" {
				return true
			}

			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return true
			}

			truncated := int64(num)
			return truncated == 0
		}

		getObjectIfArrayAtZeroExpression := func(node *ast.CallExpression) *ast.Node {
			if len(node.Arguments.Nodes) != 1 {
				return nil
			}

			callee := node.Expression

			if !ast.IsAccessExpression(callee) {
				return nil
			}

			switch callee.Kind {
			case ast.KindPropertyAccessExpression:
				if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
					return nil
				}
			case ast.KindElementAccessExpression:
				if callee.AsElementAccessExpression().QuestionDotToken != nil {
					return nil
				}
			}

			propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
			if !found || propertyName != "at" {
				return nil
			}

			atArg := node.Arguments.Nodes[0]
			value, ok := getStaticStringValue(atArg)
			if !ok {
				return nil
			}

			if isTreatedAsZeroByArrayAt(value) {
				return callee.Expression()
			}

			return nil
		}

		isMemberAccessOfZero := func(node *ast.ElementAccessExpression) bool {
			if node.QuestionDotToken != nil {
				return false
			}

			value, ok := getStaticStringValue(node.ArgumentExpression)
			if !ok {
				return false
			}

			return isTreatedAsZeroByMemberAccess(value)
		}

		generateFixes := func(
			filterExpressions []filterExpressionData,
			arrayNode *ast.Node,
			wholeExpressionNode *ast.Node,
		) []rule.RuleFix {
			fixes := []rule.RuleFix{}

			for _, filterExpr := range filterExpressions {
				filterNodeRange := utils.TrimNodeTextRange(ctx.SourceFile, filterExpr.filterNode)
				if filterExpr.isBracketSyntaxForFilter {
					fixes = append(fixes, rule.RuleFixReplaceRange(filterNodeRange, "\"find\""))
				} else {
					fixes = append(fixes, rule.RuleFixReplaceRange(filterNodeRange, "find"))
				}
			}

			s := scanner.GetScannerForSourceFile(ctx.SourceFile, arrayNode.End())
			accessTokenStart := s.TokenRange().Pos()
			wholeExprEnd := wholeExpressionNode.End()

			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(accessTokenStart, wholeExprEnd)))

			return fixes
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()

				object := getObjectIfArrayAtZeroExpression(callExpr)
				if object == nil {
					return
				}

				filterExpressions := parseArrayFilterExpressions(object)
				if len(filterExpressions) == 0 {
					return
				}

				ctx.ReportNodeWithSuggestions(node, buildPreferFindMessage(), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message:  buildPreferFindSuggestionMessage(),
						FixesArr: generateFixes(filterExpressions, object, node),
					}}
				})
			},

			ast.KindElementAccessExpression: func(node *ast.Node) {
				elemAccess := node.AsElementAccessExpression()

				if !isMemberAccessOfZero(elemAccess) {
					return
				}

				object := elemAccess.Expression
				filterExpressions := parseArrayFilterExpressions(object)
				if len(filterExpressions) == 0 {
					return
				}

				ctx.ReportNodeWithSuggestions(node, buildPreferFindMessage(), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message:  buildPreferFindSuggestionMessage(),
						FixesArr: generateFixes(filterExpressions, object, node),
					}}
				})
			},
		}
	},
}
