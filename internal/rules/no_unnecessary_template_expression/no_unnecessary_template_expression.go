package no_unnecessary_template_expression

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/jsnum"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildNoUnnecessaryTemplateExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUnnecessaryTemplateExpression",
		Description: "Template literal expression is unnecessary and can be simplified.",
	}
}

func isUnderlyingTypeString(t *checker.Type) bool {
	return utils.Every(utils.UnionTypeParts(t), func(t *checker.Type) bool {
		return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
		})
	})
}

func isAnyLiteral(node *ast.Node) bool {
	return ast.IsLiteralExpression(node) || ast.IsBooleanLiteral(node) || node.Kind == ast.KindNullKeyword
}

func isFixableIdentifier(node *ast.Node) bool {
	if ast.IsIdentifier(node) {
		name := node.AsIdentifier().Text
		return name == "undefined" || name == "Infinity" || name == "NaN"
	}
	return node.Kind == ast.KindUndefinedKeyword
}

func startsWithNewline(str string) bool {
	return strings.HasPrefix(str, "\n") || strings.HasPrefix(str, "\r\n")
}

func isWhitespace(str string) bool {
	// allow empty string too since we went to allow
	// `      ${''}
	// `;
	//
	// in addition to
	// `${'        '}
	// `;

	for _, r := range str {
		if !utils.IsStrWhiteSpace(r) {
			return false
		}
	}
	return true
}

func endsWithUnescapedDollarSign(str string) bool {
	if !strings.HasSuffix(str, "$") {
		return false
	}

	backslashes := 0
	for i := len(str) - 2; i >= 0 && str[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 0
}

func escapeTemplateRawText(text string) string {
	var builder strings.Builder
	for i := range len(text) {
		needsEscape := text[i] == '`' || (text[i] == '$' && i+1 < len(text) && text[i+1] == '{')
		if needsEscape {
			backslashes := 0
			for j := i - 1; j >= 0 && text[j] == '\\'; j-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				builder.WriteByte('\\')
			}
		}
		builder.WriteByte(text[i])
	}
	return builder.String()
}

func canonicalNumericLiteralText(text string) string {
	text = strings.ReplaceAll(text, "_", "")
	return jsnum.FromString(text).String()
}

func canonicalBigIntLiteralText(text string) string {
	return jsnum.ParsePseudoBigInt(strings.ReplaceAll(text, "_", ""))
}

var NoUnnecessaryTemplateExpressionRule = rule.Rule{
	Name: "no-unnecessary-template-expression",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()

		nodeText := func(node *ast.Node) string {
			textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return sourceText[textRange.Pos():textRange.End()]
		}

		reportSingleInterpolation := func(template *ast.Node, interpolation *ast.Node, spanLiteral *ast.Node) {
			text := nodeText(interpolation)
			isMemberReceiver := ast.IsPropertyAccessExpression(template.Parent) && template.Parent.AsPropertyAccessExpression().Expression == template ||
				ast.IsElementAccessExpression(template.Parent) && template.Parent.AsElementAccessExpression().Expression == template
			if isMemberReceiver && ast.GetExpressionPrecedence(interpolation) < ast.OperatorPrecedenceMember {
				text = "(" + text + ")"
			}

			fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, template, text)}
			ctx.ReportDiagnosticWithFixes(rule.RuleDiagnostic{
				Range:   core.NewTextRange(interpolation.Pos()-2, spanLiteral.Pos()+1),
				Message: buildNoUnnecessaryTemplateExpressionMessage(),
			}, func() []rule.RuleFix {
				return fixes
			})
		}

		isUnnecessaryValueInterpolation := func(expression *ast.Node, prevQuasiEnd int, nextQuasiLiteral *ast.TemplateMiddleOrTail) bool {
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(prevQuasiEnd, nextQuasiLiteral.Pos())) || utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(nextQuasiLiteral.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, nextQuasiLiteral).Pos())) {
				return false
			}

			if ast.IsLiteralTypeNode(expression) {
				expression = expression.AsLiteralTypeNode().Literal
			}

			if isFixableIdentifier(expression) {
				return true
			}

			if ast.IsStringLiteralLike(expression) {
				var raw string
				if nextQuasiLiteral.Kind == ast.KindTemplateMiddle {
					raw = nextQuasiLiteral.AsTemplateMiddle().RawText
				} else {
					raw = nextQuasiLiteral.AsTemplateTail().RawText
				}

				// allow trailing whitespace literal
				return !startsWithNewline(raw) || !isWhitespace(expression.Text())
			}

			return isAnyLiteral(expression) || ast.IsTemplateExpression(expression)
		}

		getRawText := func(literal *ast.TemplateMiddleOrTail) string {
			if literal.Kind == ast.KindTemplateMiddle {
				return literal.AsTemplateMiddle().RawText
			}
			return literal.AsTemplateTail().RawText
		}

		getLiteral := func(node *ast.Node) *ast.Node {
			if ast.IsLiteralTypeNode(node) {
				node = node.AsLiteralTypeNode().Literal
			}
			if ast.IsLiteralExpression(node) {
				return node
			}
			return nil
		}

		getTemplateLiteral := func(node *ast.Node) *ast.Node {
			if ast.IsLiteralTypeNode(node) {
				node = node.AsLiteralTypeNode().Literal
			}
			if ast.IsTemplateExpression(node) {
				return node
			}
			return nil
		}

		buildLiteralReplacementText := func(literal *ast.Node, nextCharacterIsOpeningCurlyBrace bool) (string, bool, bool) {
			textRange := utils.TrimNodeTextRange(ctx.SourceFile, literal)
			var text string
			switch literal.Kind {
			case ast.KindStringLiteral:
				rawText := sourceText[textRange.Pos()+1 : textRange.End()-1]
				text = rawText
			case ast.KindNoSubstitutionTemplateLiteral:
				rawText := sourceText[textRange.Pos()+1 : textRange.End()-1]
				text = rawText
			case ast.KindNumericLiteral:
				text = canonicalNumericLiteralText(literal.Text())
			case ast.KindBigIntLiteral:
				text = canonicalBigIntLiteralText(literal.Text())
			case ast.KindRegularExpressionLiteral:
				text = strings.ReplaceAll(literal.Text(), `\`, `\\`)
			default:
				text = nodeText(literal)
			}

			text = escapeTemplateRawText(text)
			if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(text) {
				text = text[:len(text)-1] + `\$`
			}
			return text, strings.HasPrefix(text, "{"), text != ""
		}

		isTrivialInterpolation := func(templateSpans *ast.NodeList, head *ast.TemplateHeadNode, firstSpanLiteral *ast.Node) bool {
			return len(templateSpans.Nodes) == 1 && head.AsTemplateHead().Text == "" && firstSpanLiteral.Text() == "" && !utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(head.End(), firstSpanLiteral.Pos())) && !utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(firstSpanLiteral.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, firstSpanLiteral).Pos()))
		}

		isEnumMemberType := func(t *checker.Type) bool {
			return utils.TypeRecurser(t, func(t *checker.Type) bool {
				symbol := checker.Type_symbol(t)
				return symbol != nil && symbol.ValueDeclaration != nil && ast.IsEnumMember(symbol.ValueDeclaration)
			})
		}

		checkTemplateSpans := func(templateSpans *ast.NodeList, head *ast.TemplateHeadNode) {
			nextCharacterIsOpeningCurlyBrace := false

			for i := len(templateSpans.Nodes) - 1; i >= 0; i-- {
				span := templateSpans.Nodes[i]
				var prevQuasiEnd int
				if i == 0 {
					prevQuasiEnd = head.End()
				} else {
					prevQuasiEnd = templateSpans.Nodes[i-1].End()
				}

				var expr *ast.Node
				var literal *ast.TemplateMiddleOrTail
				if span.Kind == ast.KindTemplateSpan {
					s := span.AsTemplateSpan()
					expr = s.Expression
					literal = s.Literal
				} else {
					s := span.AsTemplateLiteralTypeSpan()
					expr = s.Type
					literal = s.Literal
				}

				if !isUnnecessaryValueInterpolation(expr, prevQuasiEnd, literal) {
					continue
				}

				raw := getRawText(literal)
				if raw != "" {
					nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(raw, "{")
				}

				exprRange := utils.TrimNodeTextRange(ctx.SourceFile, expr)
				fixes := []rule.RuleFix{
					rule.RuleFixRemoveRange(core.NewTextRange(prevQuasiEnd-2, exprRange.Pos())),
					rule.RuleFixRemoveRange(core.NewTextRange(exprRange.End(), utils.TrimNodeTextRange(ctx.SourceFile, literal).Pos()+1)),
				}

				if literal := getLiteral(expr); literal != nil {
					replacement, startsWithOpeningCurlyBrace, shouldUpdateNextCharacter := buildLiteralReplacementText(literal, nextCharacterIsOpeningCurlyBrace)
					if shouldUpdateNextCharacter {
						nextCharacterIsOpeningCurlyBrace = startsWithOpeningCurlyBrace
					}
					fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, literal, replacement))
				} else if templateLiteral := getTemplateLiteral(expr); templateLiteral != nil {
					templateLiteralRange := utils.TrimNodeTextRange(ctx.SourceFile, templateLiteral)
					templateExpr := templateLiteral.AsTemplateExpression()
					lastSpan := templateExpr.TemplateSpans.Nodes[len(templateExpr.TemplateSpans.Nodes)-1].AsTemplateSpan()
					if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(lastSpan.Literal.RawText()) {
						fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(templateLiteralRange.End()-2, templateLiteralRange.End()-2), `\`))
					}
					if templateExpr.Head.AsTemplateHead().RawText != "" {
						nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(templateExpr.Head.AsTemplateHead().RawText, "{")
					}
					fixes = append(fixes,
						rule.RuleFixRemoveRange(core.NewTextRange(templateLiteralRange.Pos(), templateLiteralRange.Pos()+1)),
						rule.RuleFixRemoveRange(core.NewTextRange(templateLiteralRange.End()-1, templateLiteralRange.End())),
					)
				} else {
					nextCharacterIsOpeningCurlyBrace = false
				}

				prevRaw := ""
				if i == 0 {
					prevRaw = head.AsTemplateHead().RawText
				} else {
					prevSpan := templateSpans.Nodes[i-1]
					if prevSpan.Kind == ast.KindTemplateSpan {
						prevRaw = prevSpan.AsTemplateSpan().Literal.RawText()
					} else {
						prevRaw = prevSpan.AsTemplateLiteralTypeSpan().Literal.RawText()
					}
				}
				if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(prevRaw) {
					fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(prevQuasiEnd-3, prevQuasiEnd-2), `\$`))
				}

				reportRange := core.NewTextRange(prevQuasiEnd-2, utils.TrimNodeTextRange(ctx.SourceFile, literal).Pos()+1)
				ctx.ReportDiagnosticWithFixes(rule.RuleDiagnostic{
					Range:   reportRange,
					Message: buildNoUnnecessaryTemplateExpressionMessage(),
				}, func() []rule.RuleFix {
					return fixes
				})
			}
		}

		return rule.RuleListeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				if ast.IsTaggedTemplateExpression(node.Parent) {
					return
				}

				expr := node.AsTemplateExpression()
				firstSpan := expr.TemplateSpans.Nodes[0].AsTemplateSpan()

				if isTrivialInterpolation(expr.TemplateSpans, expr.Head, firstSpan.Literal) {
					constraintType, _ := utils.GetConstraintInfo(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(firstSpan.Expression))

					if constraintType != nil && isUnderlyingTypeString(constraintType) {
						reportSingleInterpolation(node, firstSpan.Expression, firstSpan.Literal)
						return
					}
				}

				checkTemplateSpans(expr.TemplateSpans, expr.Head)
			},
			ast.KindTemplateLiteralType: func(node *ast.Node) {
				expr := node.AsTemplateLiteralTypeNode()
				firstSpan := expr.TemplateSpans.Nodes[0].AsTemplateLiteralTypeSpan()

				if isTrivialInterpolation(expr.TemplateSpans, expr.Head, firstSpan.Literal) {
					constraintType, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(firstSpan.Type))

					if constraintType != nil && !isTypeParameter && isUnderlyingTypeString(constraintType) && !isEnumMemberType(constraintType) {
						reportSingleInterpolation(node, firstSpan.Type, firstSpan.Literal)
						return
					}
				}

				checkTemplateSpans(expr.TemplateSpans, expr.Head)
			},
		}
	},
}
