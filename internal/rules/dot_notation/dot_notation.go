package dot_notation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var es3Keywords = map[string]struct{}{
	"abstract": {}, "boolean": {}, "break": {}, "byte": {}, "case": {}, "catch": {}, "char": {}, "class": {}, "const": {}, "continue": {}, "debugger": {},
	"default": {}, "delete": {}, "do": {}, "double": {}, "else": {}, "enum": {}, "export": {}, "extends": {}, "false": {}, "final": {}, "finally": {},
	"float": {}, "for": {}, "function": {}, "goto": {}, "if": {}, "implements": {}, "import": {}, "in": {}, "instanceof": {}, "int": {}, "interface": {},
	"long": {}, "native": {}, "new": {}, "null": {}, "package": {}, "private": {}, "protected": {}, "public": {}, "return": {}, "short": {}, "static": {},
	"super": {}, "switch": {}, "synchronized": {}, "this": {}, "throw": {}, "throws": {}, "transient": {}, "true": {}, "try": {}, "typeof": {}, "var": {},
	"void": {}, "volatile": {}, "while": {}, "with": {},
}

func buildUseDotMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useDot",
		Description: fmt.Sprintf("[%s] is better written in dot notation.", key),
	}
}

func buildUseBracketsMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useBrackets",
		Description: fmt.Sprintf(".%s is a syntax error.", key),
	}
}

func isKeyword(name string) bool {
	_, ok := es3Keywords[name]
	return ok
}

func getComputedPropertyValue(node *ast.Node) (value string, formatted string, ok bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		value = node.AsStringLiteral().Text
		return value, strconv.Quote(value), true
	case ast.KindTrueKeyword:
		return "true", "true", true
	case ast.KindFalseKeyword:
		return "false", "false", true
	case ast.KindNullKeyword:
		return "null", "null", true
	case ast.KindNoSubstitutionTemplateLiteral:
		value = node.AsNoSubstitutionTemplateLiteral().Text
		return value, "`" + value + "`", true
	default:
		return "", "", false
	}
}

func getStaticPropertyName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	default:
		return "", false
	}
}

func isDecimalInteger(node *ast.Node) bool {
	if !ast.IsNumericLiteral(node) {
		return false
	}

	text := node.AsNumericLiteral().Text
	if text == "" {
		return false
	}

	if strings.ContainsAny(text, ".eExXoObBnN") {
		return false
	}

	for i, ch := range text {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '_' && i > 0 && i < len(text)-1 {
			continue
		}
		return false
	}

	return true
}

func needsSpaceAfterIdentifier(sourceText string, memberExprEnd int) bool {
	nextPos := scanner.SkipTrivia(sourceText, memberExprEnd)
	if nextPos != memberExprEnd || nextPos >= len(sourceText) {
		return false
	}

	nextRune, _ := utf8.DecodeRuneInString(sourceText[nextPos:])
	return scanner.IsIdentifierPart(nextRune)
}

func getFirstDeclaration(symbol *ast.Symbol) *ast.Node {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil
	}
	return symbol.Declarations[0]
}

var DotNotationRule = rule.Rule{
	Name: "dot-notation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[DotNotationOptions](options, "dot-notation")

		allowPattern := (*regexp.Regexp)(nil)
		if opts.AllowPattern != "" {
			compiled, err := regexp.Compile(opts.AllowPattern)
			if err != nil {
				panic(fmt.Sprintf("dot-notation: invalid allowPattern %q: %v", opts.AllowPattern, err))
			}
			allowPattern = compiled
		}

		allowIndexSignaturePropertyAccess :=
			opts.AllowIndexSignaturePropertyAccess ||
				ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue()

		reportComputedProperty := func(node *ast.Node) {
			elementAccess := node.AsElementAccessExpression()
			property := elementAccess.ArgumentExpression
			value, formatted, ok := getComputedPropertyValue(property)
			if !ok {
				return
			}

			if !scanner.IsIdentifierText(value, core.LanguageVariantStandard) {
				return
			}

			if !opts.AllowKeywords && isKeyword(value) {
				return
			}

			if allowPattern != nil && allowPattern.MatchString(value) {
				return
			}

			msg := buildUseDotMessage(formatted)

			ctx.ReportNodeWithFixes(property, msg, func() []rule.RuleFix {
				objectRange := utils.TrimNodeTextRange(ctx.SourceFile, node.Expression())
				propertyRange := utils.TrimNodeTextRange(ctx.SourceFile, property)

				leftBracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, objectRange.End())
				rightBracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, propertyRange.End())

				fixes := make([]rule.RuleFix, 0, 3)
				if elementAccess.QuestionDotToken == nil {
					dotText := "."
					if isDecimalInteger(node.Expression()) {
						dotText = " ."
					}
					fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(leftBracketRange.Pos(), leftBracketRange.Pos()), dotText))
				}

				fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(leftBracketRange.Pos(), rightBracketRange.End()), value))

				if needsSpaceAfterIdentifier(ctx.SourceFile.Text(), node.End()) {
					fixes = append(fixes, rule.RuleFixInsertAfter(node, " "))
				}

				return fixes
			})
		}

		reportDotKeywordAccess := func(node *ast.Node) {
			if opts.AllowKeywords {
				return
			}

			propertyAccess := node.AsPropertyAccessExpression()
			property := propertyAccess.Name()
			if !ast.IsIdentifier(property) {
				return
			}

			propertyName := property.Text()
			if !isKeyword(propertyName) {
				return
			}

			msg := buildUseBracketsMessage(propertyName)
			ctx.ReportNodeWithFixes(property.AsNode(), msg, func() []rule.RuleFix {
				if propertyAccess.QuestionDotToken == nil && ast.IsIdentifier(node.Expression()) && node.Expression().AsIdentifier().Text == "let" {
					return nil
				}

				objectRange := utils.TrimNodeTextRange(ctx.SourceFile, node.Expression())
				dotRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, objectRange.End())
				fixes := make([]rule.RuleFix, 0, 2)
				if propertyAccess.QuestionDotToken == nil {
					fixes = append(fixes, rule.RuleFixRemoveRange(dotRange))
				}
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, property.AsNode(), fmt.Sprintf(`["%s"]`, propertyName)))

				return fixes
			})
		}

		shouldSkipForTypeAwareAllowances := func(node *ast.Node) bool {
			if !opts.AllowPrivateClassPropertyAccess && !opts.AllowProtectedClassPropertyAccess && !allowIndexSignaturePropertyAccess {
				return false
			}

			elementAccess := node.AsElementAccessExpression()
			property := elementAccess.ArgumentExpression

			propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(property)
			propertyName, hasStaticPropertyName := getStaticPropertyName(property)
			if propertySymbol == nil && hasStaticPropertyName {
				objectType := utils.GetNonNullableType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node.Expression()))
				if objectType != nil {
					for _, candidate := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, objectType) {
						if candidate.Name == propertyName {
							propertySymbol = candidate
							break
						}
					}
				}
			}

			declaration := getFirstDeclaration(propertySymbol)
			if declaration != nil {
				if opts.AllowPrivateClassPropertyAccess && ast.HasModifier(declaration, ast.ModifierFlagsPrivate) {
					return true
				}
				if opts.AllowProtectedClassPropertyAccess && ast.HasModifier(declaration, ast.ModifierFlagsProtected) {
					return true
				}
				if allowIndexSignaturePropertyAccess && declaration.Kind == ast.KindIndexSignature {
					return true
				}
			}

			if (propertySymbol == nil || len(propertySymbol.Declarations) == 0) && allowIndexSignaturePropertyAccess {
				if ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue() {
					return true
				}

				objectType := utils.GetNonNullableType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node.Expression()))
				if objectType != nil {
					objectType = checker.Checker_getApparentType(ctx.TypeChecker, objectType)
					keyType := ctx.TypeChecker.GetTypeAtLocation(property)
					baseKeyType := checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, keyType)
					if checker.Checker_getIndexTypeOfType(ctx.TypeChecker, objectType, keyType) != nil ||
						(baseKeyType != nil && checker.Checker_getIndexTypeOfType(ctx.TypeChecker, objectType, baseKeyType) != nil) {
						return true
					}
				}
			}

			return false
		}

		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindElementAccessExpression): func(node *ast.Node) {
				if shouldSkipForTypeAwareAllowances(node) {
					return
				}
				reportComputedProperty(node)
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				reportDotKeywordAccess(node)
			},
		}
	},
}
