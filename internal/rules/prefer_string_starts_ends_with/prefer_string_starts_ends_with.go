package prefer_string_starts_ends_with

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferStartsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStartsWith",
		Description: "Use 'String#startsWith' method instead.",
	}
}

func buildPreferEndsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEndsWith",
		Description: "Use 'String#endsWith' method instead.",
	}
}

type parsedRegExp struct {
	isEndsWith   bool
	isStartsWith bool
	text         string
}

type nullishKind int

const (
	nullishKindUnknown nullishKind = iota
	nullishKindNull
	nullishKindUndefined
)

var PreferStringStartsEndsWithRule = rule.Rule{
	Name: "prefer-string-starts-ends-with",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferStringStartsEndsWithOptions](options, "prefer-string-starts-ends-with")

		normalizedNodeText := func(node *ast.Node) string {
			r := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(node))
			text := ctx.SourceFile.Text()[r.Pos():r.End()]
			return strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) {
					return -1
				}
				return r
			}, text)
		}

		isSameTokens := func(node1 *ast.Node, node2 *ast.Node) bool {
			return normalizedNodeText(node1) == normalizedNodeText(node2)
		}

		isEqualityComparison := func(node *ast.Node) bool {
			if !ast.IsBinaryExpression(node) {
				return false
			}
			switch node.AsBinaryExpression().OperatorToken.Kind {
			case ast.KindEqualsEqualsToken,
				ast.KindEqualsEqualsEqualsToken,
				ast.KindExclamationEqualsToken,
				ast.KindExclamationEqualsEqualsToken:
				return true
			}
			return false
		}

		isNegativeEqualityOperator := func(kind ast.Kind) bool {
			return kind == ast.KindExclamationEqualsToken || kind == ast.KindExclamationEqualsEqualsToken
		}

		isLooseEqualityOperator := func(kind ast.Kind) bool {
			return kind == ast.KindEqualsEqualsToken || kind == ast.KindExclamationEqualsToken
		}

		memberObject := func(node *ast.Node) *ast.Node {
			if ast.IsPropertyAccessExpression(node) {
				return node.AsPropertyAccessExpression().Expression
			}
			if ast.IsElementAccessExpression(node) {
				return node.AsElementAccessExpression().Expression
			}
			return nil
		}

		memberOptional := func(node *ast.Node) bool {
			if ast.IsPropertyAccessExpression(node) {
				return node.AsPropertyAccessExpression().QuestionDotToken != nil
			}
			if ast.IsElementAccessExpression(node) {
				return node.AsElementAccessExpression().QuestionDotToken != nil
			}
			return false
		}

		memberOptionalOperator := func(node *ast.Node) string {
			if memberOptional(node) {
				return "?."
			}
			return "."
		}

		getPropertyName := func(node *ast.Node) (string, bool) {
			if ast.IsPropertyAccessExpression(node) {
				nameNode := node.AsPropertyAccessExpression().Name()
				if nameNode != nil {
					return nameNode.Text(), true
				}
				return "", false
			}

			if ast.IsElementAccessExpression(node) {
				arg := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
				if arg == nil {
					return "", false
				}
				if ast.IsStringLiteral(arg) || arg.Kind == ast.KindNoSubstitutionTemplateLiteral {
					return arg.Text(), true
				}
			}

			return "", false
		}

		getPropertyRange := func(node *ast.Node) core.TextRange {
			obj := memberObject(node)
			if obj == nil {
				return utils.TrimNodeTextRange(ctx.SourceFile, node)
			}
			start := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, obj.End())
			memberRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return start.WithEnd(memberRange.End())
		}

		callMemberCallee := func(callExpr *ast.CallExpression) *ast.Node {
			expr := ast.SkipParentheses(callExpr.Expression)
			if ast.IsPropertyAccessExpression(expr) || ast.IsElementAccessExpression(expr) {
				return expr
			}
			return nil
		}

		isStringType := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(ast.SkipParentheses(node))
			return utils.GetTypeName(ctx.TypeChecker, t) == "string"
		}

		var staticNumber func(node *ast.Node) (float64, bool)
		staticNumber = func(node *ast.Node) (float64, bool) {
			node = ast.SkipParentheses(node)
			if node == nil {
				return 0, false
			}
			if ast.IsNumericLiteral(node) {
				n, err := strconv.ParseFloat(strings.ReplaceAll(node.AsNumericLiteral().Text, "_", ""), 64)
				if err != nil {
					return 0, false
				}
				return n, true
			}
			if ast.IsPrefixUnaryExpression(node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix.Operator == ast.KindMinusToken {
					if n, ok := staticNumber(prefix.Operand); ok {
						return -n, true
					}
				}
			}
			return 0, false
		}

		var staticString func(node *ast.Node) (string, bool)
		staticString = func(node *ast.Node) (string, bool) {
			node = ast.SkipParentheses(node)
			if node == nil {
				return "", false
			}

			if ast.IsStringLiteral(node) || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
				return node.Text(), true
			}

			if ast.IsIdentifier(node) {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) {
					return "", false
				}
				decl := symbol.ValueDeclaration.AsVariableDeclaration()
				if decl.Initializer == nil {
					return "", false
				}
				return staticString(decl.Initializer)
			}

			return "", false
		}

		isCharacter := func(node *ast.Node) bool {
			value, ok := staticString(node)
			if !ok {
				return false
			}
			return len(utf16.Encode([]rune(value))) == 1
		}

		var getNullishKind func(node *ast.Node) nullishKind
		getNullishKind = func(node *ast.Node) nullishKind {
			node = ast.SkipParentheses(node)
			if node == nil {
				return nullishKindUnknown
			}
			if utils.IsNullLiteral(node) {
				return nullishKindNull
			}
			if utils.IsUndefinedLiteral(node) {
				return nullishKindUndefined
			}
			if ast.IsIdentifier(node) {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) {
					return nullishKindUnknown
				}
				decl := symbol.ValueDeclaration.AsVariableDeclaration()
				if decl.Initializer == nil {
					return nullishKindUnknown
				}
				return getNullishKind(decl.Initializer)
			}
			return nullishKindUnknown
		}

		isLengthExpression := func(node *ast.Node, expectedObject *ast.Node) bool {
			node = ast.SkipParentheses(node)
			expectedObject = ast.SkipParentheses(expectedObject)
			if node == nil || expectedObject == nil {
				return false
			}

			if ast.IsPropertyAccessExpression(node) || ast.IsElementAccessExpression(node) {
				name, ok := getPropertyName(node)
				if ok && name == "length" {
					obj := memberObject(node)
					if obj != nil && isSameTokens(obj, expectedObject) {
						return true
					}
				}
			}

			evaluatedLength, okLength := staticNumber(node)
			evaluatedString, okString := staticString(expectedObject)
			if !okLength || !okString {
				return false
			}
			return evaluatedLength == float64(len(utf16.Encode([]rune(evaluatedString))))
		}

		isLengthAheadOfEnd := func(node *ast.Node, substring *ast.Node, parentString *ast.Node) bool {
			node = ast.SkipParentheses(node)
			if node == nil {
				return false
			}

			if ast.IsPrefixUnaryExpression(node) {
				prefix := node.AsPrefixUnaryExpression()
				return prefix.Operator == ast.KindMinusToken && isLengthExpression(prefix.Operand, substring)
			}

			if ast.IsBinaryExpression(node) {
				bin := node.AsBinaryExpression()
				return bin.OperatorToken.Kind == ast.KindMinusToken &&
					isLengthExpression(bin.Left, parentString) &&
					isLengthExpression(bin.Right, substring)
			}

			return false
		}

		isLastIndexExpression := func(node *ast.Node, expectedObject *ast.Node) bool {
			node = ast.SkipParentheses(node)
			if !ast.IsBinaryExpression(node) {
				return false
			}
			bin := node.AsBinaryExpression()
			return bin.OperatorToken.Kind == ast.KindMinusToken &&
				isLengthExpression(bin.Left, expectedObject) &&
				func() bool {
					n, ok := staticNumber(bin.Right)
					return ok && n == 1
				}()
		}

		splitRegexLiteral := func(text string) (pattern string, flags string, ok bool) {
			if len(text) < 2 || text[0] != '/' {
				return "", "", false
			}
			closingSlash := -1
			for i := len(text) - 1; i > 0; i-- {
				if text[i] != '/' {
					continue
				}
				backslashes := 0
				for j := i - 1; j >= 0 && text[j] == '\\'; j-- {
					backslashes++
				}
				if backslashes%2 == 0 {
					closingSlash = i
					break
				}
			}
			if closingSlash <= 0 {
				return "", "", false
			}
			return text[1:closingSlash], text[closingSlash+1:], true
		}

		parseRegExpText := func(pattern string) (string, bool) {
			isStartsWith := strings.HasPrefix(pattern, "^")
			isEndsWith := strings.HasSuffix(pattern, "$")
			if isStartsWith == isEndsWith {
				return "", false
			}

			content := pattern
			if isStartsWith {
				content = content[1:]
			} else {
				content = content[:len(content)-1]
			}

			var builder strings.Builder
			escaped := false
			for _, ch := range content {
				if escaped {
					switch ch {
					case 'n':
						builder.WriteByte('\n')
					case 'r':
						builder.WriteByte('\r')
					case 't':
						builder.WriteByte('\t')
					case 'v':
						builder.WriteByte('\v')
					case 'f':
						builder.WriteByte('\f')
					case '0':
						builder.WriteByte(0)
					case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'c', 'x', 'u', 'k', 'p', 'P':
						return "", false
					default:
						if unicode.IsDigit(ch) {
							return "", false
						}
						builder.WriteRune(ch)
					}
					escaped = false
					continue
				}

				if ch == '\\' {
					escaped = true
					continue
				}

				switch ch {
				case '.', '*', '+', '?', '|', '^', '$', '[', ']', '(', ')', '{', '}':
					return "", false
				default:
					builder.WriteRune(ch)
				}
			}

			if escaped {
				return "", false
			}

			return builder.String(), true
		}

		var parseRegExp func(node *ast.Node) (*parsedRegExp, bool)
		parseRegExp = func(node *ast.Node) (*parsedRegExp, bool) {
			node = ast.SkipParentheses(node)
			if node == nil {
				return nil, false
			}

			var pattern string
			var flags string

			switch {
			case node.Kind == ast.KindRegularExpressionLiteral:
				parsedPattern, parsedFlags, ok := splitRegexLiteral(node.AsRegularExpressionLiteral().Text)
				if !ok {
					return nil, false
				}
				pattern = parsedPattern
				flags = parsedFlags
			case ast.IsNewExpression(node):
				newExpr := node.AsNewExpression()
				if !ast.IsIdentifier(newExpr.Expression) || newExpr.Expression.AsIdentifier().Text != "RegExp" {
					return nil, false
				}
				args := node.Arguments()
				if len(args) == 0 || !ast.IsStringLiteral(args[0]) {
					return nil, false
				}
				pattern = args[0].AsStringLiteral().Text
				if len(args) > 1 {
					if !ast.IsStringLiteral(args[1]) {
						return nil, false
					}
					flags = args[1].AsStringLiteral().Text
				}
			case ast.IsIdentifier(node):
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) {
					return nil, false
				}
				decl := symbol.ValueDeclaration.AsVariableDeclaration()
				if decl.Initializer == nil {
					return nil, false
				}
				return parseRegExp(decl.Initializer)
			default:
				return nil, false
			}

			if strings.ContainsAny(flags, "imgy") {
				return nil, false
			}

			isStartsWith := strings.HasPrefix(pattern, "^")
			isEndsWith := strings.HasSuffix(pattern, "$")
			if isStartsWith == isEndsWith {
				return nil, false
			}

			text, ok := parseRegExpText(pattern)
			if !ok {
				return nil, false
			}

			return &parsedRegExp{
				isEndsWith:   isEndsWith,
				isStartsWith: isStartsWith,
				text:         text,
			}, true
		}

		fixWithRightOperand := func(bin *ast.BinaryExpression, memberNode *ast.Node, kind string) []rule.RuleFix {
			fixes := []rule.RuleFix{}
			if isNegativeEqualityOperator(bin.OperatorToken.Kind) {
				fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, &bin.Node, "!"))
			}

			propertyRange := getPropertyRange(memberNode)
			rightRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(bin.Right))
			binRange := utils.TrimNodeTextRange(ctx.SourceFile, &bin.Node)

			fixes = append(fixes,
				rule.RuleFixReplaceRange(propertyRange.WithEnd(rightRange.Pos()), fmt.Sprintf("%s%ssWith(", memberOptionalOperator(memberNode), kind)),
				rule.RuleFixReplaceRange(rightRange.WithPos(rightRange.End()).WithEnd(binRange.End()), ")"),
			)

			return fixes
		}

		fixWithArgument := func(bin *ast.BinaryExpression, callExpr *ast.CallExpression, memberNode *ast.Node, kind string) []rule.RuleFix {
			fixes := []rule.RuleFix{}
			if isNegativeEqualityOperator(bin.OperatorToken.Kind) {
				fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, &bin.Node, "!"))
			}

			propertyRange := getPropertyRange(memberNode)
			callRange := utils.TrimNodeTextRange(ctx.SourceFile, &callExpr.Node)
			binRange := utils.TrimNodeTextRange(ctx.SourceFile, &bin.Node)

			fixes = append(fixes,
				rule.RuleFixReplaceRange(propertyRange, fmt.Sprintf("%s%ssWith", memberOptionalOperator(memberNode), kind)),
				rule.RuleFixRemoveRange(callRange.WithPos(callRange.End()).WithEnd(binRange.End())),
			)

			return fixes
		}

		reportStringIndexOrCharAt := func(bin *ast.BinaryExpression, memberNode *ast.Node, indexNode *ast.Node) {
			obj := memberObject(memberNode)
			if obj == nil || !isStringType(obj) {
				return
			}

			isEndsWith := isLastIndexExpression(indexNode, obj)
			allowSingleElementEquality := opts.AllowSingleElementEquality != nil &&
				*opts.AllowSingleElementEquality == PreferStringStartsEndsWithOptionsAllowSingleElementEqualityAlways
			if allowSingleElementEquality && isEndsWith {
				return
			}

			startsAtZero := false
			if !isEndsWith {
				if n, ok := staticNumber(indexNode); ok && n == 0 {
					startsAtZero = true
				}
			}

			if allowSingleElementEquality && startsAtZero {
				return
			}

			if !startsAtZero && !isEndsWith {
				return
			}

			message := buildPreferEndsWithMessage()
			kind := "end"
			if startsAtZero {
				message = buildPreferStartsWithMessage()
				kind = "start"
			}

			ctx.ReportNodeWithFixes(&bin.Node, message, func() []rule.RuleFix {
				if !isCharacter(bin.Right) {
					return nil
				}
				return fixWithRightOperand(bin, memberNode, kind)
			})
		}

		handleBinaryCallExpression := func(bin *ast.BinaryExpression, callExpr *ast.CallExpression) {
			memberNode := callMemberCallee(callExpr)
			if memberNode == nil {
				return
			}

			propName, ok := getPropertyName(memberNode)
			if !ok {
				return
			}

			switch propName {
			case "charAt":
				if len(callExpr.Arguments.Nodes) != 1 {
					return
				}
				reportStringIndexOrCharAt(bin, memberNode, callExpr.Arguments.Nodes[0])
			case "indexOf":
				if len(callExpr.Arguments.Nodes) != 1 {
					return
				}
				if n, ok := staticNumber(bin.Right); !ok || n != 0 {
					return
				}
				obj := memberObject(memberNode)
				if obj == nil || !isStringType(obj) {
					return
				}

				ctx.ReportNodeWithFixes(&bin.Node, buildPreferStartsWithMessage(), func() []rule.RuleFix {
					return fixWithArgument(bin, callExpr, memberNode, "start")
				})
			case "lastIndexOf":
				if len(callExpr.Arguments.Nodes) != 1 || !ast.IsBinaryExpression(bin.Right) {
					return
				}

				rightBin := bin.Right.AsBinaryExpression()
				if rightBin.OperatorToken.Kind != ast.KindMinusToken {
					return
				}

				obj := memberObject(memberNode)
				if obj == nil || !isStringType(obj) {
					return
				}

				if !isLengthExpression(rightBin.Left, obj) || !isLengthExpression(rightBin.Right, callExpr.Arguments.Nodes[0]) {
					return
				}

				ctx.ReportNodeWithFixes(&bin.Node, buildPreferEndsWithMessage(), func() []rule.RuleFix {
					return fixWithArgument(bin, callExpr, memberNode, "end")
				})
			case "match":
				if len(callExpr.Arguments.Nodes) != 1 {
					return
				}

				obj := memberObject(memberNode)
				operandNullishKind := getNullishKind(bin.Right)
				if obj == nil || !isStringType(obj) || operandNullishKind == nullishKindUnknown {
					return
				}
				if !isLooseEqualityOperator(bin.OperatorToken.Kind) && operandNullishKind != nullishKindNull {
					return
				}

				parsed, ok := parseRegExp(callExpr.Arguments.Nodes[0])
				if !ok {
					return
				}

				message := buildPreferEndsWithMessage()
				method := "endsWith"
				if parsed.isStartsWith {
					message = buildPreferStartsWithMessage()
					method = "startsWith"
				}

				ctx.ReportNodeWithFixes(&bin.Node, message, func() []rule.RuleFix {
					fixes := []rule.RuleFix{}
					if !isNegativeEqualityOperator(bin.OperatorToken.Kind) {
						fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, &bin.Node, "!"))
					}

					propertyRange := getPropertyRange(memberNode)
					callRange := utils.TrimNodeTextRange(ctx.SourceFile, &callExpr.Node)
					argRange := utils.TrimNodeTextRange(ctx.SourceFile, callExpr.Arguments.Nodes[0])
					binRange := utils.TrimNodeTextRange(ctx.SourceFile, &bin.Node)

					fixes = append(fixes,
						rule.RuleFixReplaceRange(propertyRange, fmt.Sprintf("%s%s", memberOptionalOperator(memberNode), method)),
						rule.RuleFixReplaceRange(argRange, strconv.Quote(parsed.text)),
						rule.RuleFixRemoveRange(callRange.WithPos(callRange.End()).WithEnd(binRange.End())),
					)

					return fixes
				})
			case "slice", "substring":
				obj := memberObject(memberNode)
				if obj == nil || !isStringType(obj) {
					return
				}

				args := callExpr.Arguments.Nodes
				isStartsWith := false
				isEndsWith := false

				if len(args) == 1 {
					if isLengthAheadOfEnd(args[0], bin.Right, obj) {
						isEndsWith = true
					}
				} else if len(args) == 2 {
					if n, ok := staticNumber(args[0]); ok && n == 0 && isLengthExpression(args[1], bin.Right) {
						isStartsWith = true
					} else if (isLengthExpression(args[1], obj) || func() bool {
						n, ok := staticNumber(args[1])
						return ok && n == 0
					}()) && isLengthAheadOfEnd(args[0], bin.Right, obj) {
						isEndsWith = true
					}
				}

				if !isStartsWith && !isEndsWith {
					return
				}

				message := buildPreferEndsWithMessage()
				kind := "end"
				if isStartsWith {
					message = buildPreferStartsWithMessage()
					kind = "start"
				}

				ctx.ReportNodeWithFixes(&bin.Node, message, func() []rule.RuleFix {
					if isLooseEqualityOperator(bin.OperatorToken.Kind) {
						r := ast.SkipParentheses(bin.Right)
						if !ast.IsStringLiteral(r) {
							return nil
						}
					}

					if isStartsWith {
						if len(args) < 2 || !isLengthExpression(args[1], bin.Right) {
							return nil
						}
					} else {
						posNode := ast.SkipParentheses(args[0])
						validPos := false
						if ast.IsBinaryExpression(posNode) {
							posBin := posNode.AsBinaryExpression()
							validPos = posBin.OperatorToken.Kind == ast.KindMinusToken &&
								isLengthExpression(posBin.Left, obj) &&
								isLengthExpression(posBin.Right, bin.Right)
						} else if propName == "slice" && ast.IsPrefixUnaryExpression(posNode) {
							prefix := posNode.AsPrefixUnaryExpression()
							validPos = prefix.Operator == ast.KindMinusToken && isLengthExpression(prefix.Operand, bin.Right)
						}
						if !validPos {
							return nil
						}
					}

					return fixWithRightOperand(bin, memberNode, kind)
				})
			}
		}

		handleBinary := func(node *ast.Node) {
			if !isEqualityComparison(node) {
				return
			}

			bin := node.AsBinaryExpression()
			left := ast.SkipParentheses(bin.Left)
			if left == nil {
				return
			}

			if ast.IsElementAccessExpression(left) {
				reportStringIndexOrCharAt(bin, left, left.AsElementAccessExpression().ArgumentExpression)
				return
			}

			if ast.IsCallExpression(left) {
				handleBinaryCallExpression(bin, left.AsCallExpression())
			}
		}

		handleRegexTestCall := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) != 1 {
				return
			}

			memberNode := callMemberCallee(callExpr)
			if memberNode == nil {
				return
			}

			name, ok := getPropertyName(memberNode)
			if !ok || name != "test" {
				return
			}

			obj := memberObject(memberNode)
			if obj == nil {
				return
			}

			parsed, ok := parseRegExp(obj)
			if !ok {
				return
			}

			message := buildPreferEndsWithMessage()
			method := "endsWith"
			if parsed.isStartsWith {
				message = buildPreferStartsWithMessage()
				method = "startsWith"
			}

			arg := callExpr.Arguments.Nodes[0]

			ctx.ReportNodeWithFixes(&callExpr.Node, message, func() []rule.RuleFix {
				arg = ast.SkipParentheses(arg)
				needsParen := !ast.IsLiteralExpression(arg) &&
					arg.Kind != ast.KindNoSubstitutionTemplateLiteral &&
					arg.Kind != ast.KindTemplateExpression &&
					!ast.IsIdentifier(arg) &&
					!ast.IsPropertyAccessExpression(arg) &&
					!ast.IsElementAccessExpression(arg) &&
					!ast.IsCallExpression(arg)

				callRange := utils.TrimNodeTextRange(ctx.SourceFile, &callExpr.Node)
				argRange := utils.TrimNodeTextRange(ctx.SourceFile, arg)
				fixes := []rule.RuleFix{
					rule.RuleFixRemoveRange(callRange.WithEnd(argRange.Pos())),
				}
				if needsParen {
					fixes = append(fixes,
						rule.RuleFixInsertBefore(ctx.SourceFile, arg, "("),
						rule.RuleFixInsertAfter(arg, ")"),
					)
				}
				fixes = append(fixes, rule.RuleFixInsertAfter(arg, fmt.Sprintf("%s%s(%s", memberOptionalOperator(memberNode), method, strconv.Quote(parsed.text))))

				return fixes
			})
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: handleBinary,
			ast.KindCallExpression:   handleRegexTestCall,
		}
	},
}
