package prefer_regexp_exec

import (
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildRegExpExecOverStringMatchMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "regExpExecOverStringMatch",
		Description: "Use the `RegExp#exec()` method instead.",
	}
}

const (
	argumentTypeOther  = 0
	argumentTypeString = 1 << iota
	argumentTypeRegExp
)

type staticArgumentValueKind int

const (
	staticArgumentValueUnknown staticArgumentValueKind = iota
	staticArgumentValueString
	staticArgumentValueRegExp
	staticArgumentValueOther
)

type staticArgumentValue struct {
	kind        staticArgumentValueKind
	regExpFlags string
}

var PreferRegexpExecRule = rule.Rule{
	Name: "prefer-regexp-exec",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()

		isNodeParenthesized := func(node *ast.Node) bool {
			return ast.IsParenthesizedExpression(node.Parent) && node.Parent.AsParenthesizedExpression().Expression == node
		}

		isWeakPrecedenceParent := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			switch parent.Kind {
			case ast.KindPostfixUnaryExpression,
				ast.KindPrefixUnaryExpression,
				ast.KindBinaryExpression,
				ast.KindConditionalExpression,
				ast.KindAwaitExpression:
				return true
			}

			if ast.IsPropertyAccessExpression(parent) {
				return parent.AsPropertyAccessExpression().Expression == node
			}
			if ast.IsElementAccessExpression(parent) {
				return parent.AsElementAccessExpression().Expression == node
			}
			if ast.IsCallExpression(parent) || ast.IsNewExpression(parent) {
				return parent.Expression() == node
			}
			if ast.IsTaggedTemplateExpression(parent) {
				return parent.AsTaggedTemplateExpression().Tag == node
			}

			return false
		}

		getNodeText := func(node *ast.Node) string {
			r := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return sourceText[r.Pos():r.End()]
		}

		buildWrappingFix := func(node *ast.Node, innerNodes []*ast.Node, wrap func(code ...string) string) rule.RuleFix {
			innerCodes := make([]string, len(innerNodes))
			for i, innerNode := range innerNodes {
				code := getNodeText(innerNode)
				if !utils.IsStrongPrecedenceNode(innerNode) {
					code = "(" + code + ")"
				}
				innerCodes[i] = code
			}

			code := strings.Join(innerCodes, "")
			if wrap != nil {
				code = wrap(innerCodes...)
			}
			if isWeakPrecedenceParent(node) && !isNodeParenthesized(node) {
				code = "(" + code + ")"
			}

			return rule.RuleFixReplace(ctx.SourceFile, node, code)
		}

		extractRegexLiteralFlags := func(node *ast.Node) (string, bool) {
			if node.Kind != ast.KindRegularExpressionLiteral {
				return "", false
			}

			text := node.AsRegularExpressionLiteral().Text
			if len(text) < 2 || text[0] != '/' {
				return "", false
			}

			isEscaped := func(s string, idx int) bool {
				backslashes := 0
				for i := idx - 1; i >= 0 && s[i] == '\\'; i-- {
					backslashes++
				}
				return backslashes%2 == 1
			}

			closingSlash := -1
			for i := len(text) - 1; i > 0; i-- {
				if text[i] == '/' && !isEscaped(text, i) {
					closingSlash = i
					break
				}
			}

			if closingSlash < 0 {
				return "", false
			}

			return text[closingSlash+1:], true
		}

		isStringLiteral := func(node *ast.Node) bool {
			return node != nil && node.Kind == ast.KindStringLiteral
		}

		isRegExpConstructorCall := func(node *ast.Node) bool {
			if !ast.IsCallExpression(node) && !ast.IsNewExpression(node) {
				return false
			}
			callee := node.Expression()
			return ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "RegExp"
		}

		definitelyDoesNotContainGlobalFlag := func(node *ast.Node) bool {
			node = ast.SkipParentheses(node)
			if !isRegExpConstructorCall(node) {
				return false
			}

			arguments := node.Arguments()
			if len(arguments) < 2 {
				return true
			}

			flags := ast.SkipParentheses(arguments[1])
			if utils.IsUndefinedLiteral(flags) {
				return true
			}
			if !isStringLiteral(flags) {
				return false
			}
			return !strings.Contains(flags.AsStringLiteral().Text, "g")
		}

		var getStaticArgumentValue func(node *ast.Node, visited map[*ast.Symbol]struct{}) staticArgumentValue
		getStaticArgumentValue = func(node *ast.Node, visited map[*ast.Symbol]struct{}) staticArgumentValue {
			node = ast.SkipParentheses(node)

			switch node.Kind {
			case ast.KindStringLiteral:
				return staticArgumentValue{kind: staticArgumentValueString}
			case ast.KindRegularExpressionLiteral:
				flags, ok := extractRegexLiteralFlags(node)
				if !ok {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}
				return staticArgumentValue{kind: staticArgumentValueRegExp, regExpFlags: flags}
			case ast.KindNoSubstitutionTemplateLiteral,
				ast.KindNumericLiteral,
				ast.KindTrueKeyword,
				ast.KindFalseKeyword,
				ast.KindNullKeyword:
				return staticArgumentValue{kind: staticArgumentValueOther}
			case ast.KindIdentifier:
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}
				if _, seen := visited[symbol]; seen {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}
				visited[symbol] = struct{}{}
				defer delete(visited, symbol)

				if symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}

				declaration := symbol.ValueDeclaration.AsVariableDeclaration()
				if declaration.Initializer == nil {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}

				// Keep this conservative and match getStaticValue behavior for mutable bindings.
				if !ast.IsVariableDeclarationList(symbol.ValueDeclaration.Parent) || symbol.ValueDeclaration.Parent.Flags&ast.NodeFlagsConst == 0 {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}

				return getStaticArgumentValue(declaration.Initializer, visited)
			case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
				return getStaticArgumentValue(node.Expression(), visited)
			}

			if isRegExpConstructorCall(node) {
				arguments := node.Arguments()
				if len(arguments) > 0 && !isStringLiteral(arguments[0]) {
					return staticArgumentValue{kind: staticArgumentValueUnknown}
				}

				flags := ""
				if len(arguments) > 1 {
					if !isStringLiteral(arguments[1]) {
						return staticArgumentValue{kind: staticArgumentValueUnknown}
					}
					flags = arguments[1].AsStringLiteral().Text
				}

				if len(arguments) > 0 {
					if _, err := regexp2.Compile(arguments[0].AsStringLiteral().Text, regexp2.ECMAScript); err != nil {
						return staticArgumentValue{kind: staticArgumentValueUnknown}
					}
				}

				return staticArgumentValue{kind: staticArgumentValueRegExp, regExpFlags: flags}
			}

			return staticArgumentValue{kind: staticArgumentValueUnknown}
		}

		collectArgumentTypes := func(types []*checker.Type) int {
			result := argumentTypeOther
			for _, t := range types {
				typeName := utils.GetTypeName(ctx.TypeChecker, t)
				switch typeName {
				case "RegExp":
					result |= argumentTypeRegExp
				case "string":
					result |= argumentTypeString
				}
			}
			return result
		}

		buildStringToRegExpLiteral := func(pattern string) string {
			return "/" + strings.ReplaceAll(pattern, "/", "\\/") + "/"
		}

		reportWithFix := func(
			reportNode *ast.Node,
			callNode *ast.Node,
			objectNode *ast.Node,
			argumentNode *ast.Node,
			buildExpression func(objectCode string, argumentCode string) string,
		) {
			ctx.ReportNodeWithFixes(reportNode, buildRegExpExecOverStringMatchMessage(), func() []rule.RuleFix {
				return []rule.RuleFix{
					buildWrappingFix(callNode, []*ast.Node{objectNode, argumentNode}, func(code ...string) string {
						return buildExpression(code[0], code[1])
					}),
				}
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpression := node.AsCallExpression()
				if len(callExpression.Arguments.Nodes) != 1 {
					return
				}

				callee := callExpression.Expression
				if !ast.IsPropertyAccessExpression(callee) && !ast.IsElementAccessExpression(callee) {
					return
				}

				propertyName, ok := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !ok || propertyName != "match" {
					return
				}

				objectNode := callee.Expression()
				objectType := ctx.TypeChecker.GetTypeAtLocation(objectNode)
				if utils.GetTypeName(ctx.TypeChecker, objectType) != "string" {
					return
				}

				argumentNode := callExpression.Arguments.Nodes[0]
				staticArgument := getStaticArgumentValue(argumentNode, map[*ast.Symbol]struct{}{})
				argumentType := ctx.TypeChecker.GetTypeAtLocation(argumentNode)
				argumentTypes := collectArgumentTypes(utils.UnionTypeParts(argumentType))

				if staticArgument.kind == staticArgumentValueRegExp && strings.Contains(staticArgument.regExpFlags, "g") {
					return
				}
				if staticArgument.kind == staticArgumentValueUnknown &&
					argumentTypes&argumentTypeRegExp != 0 &&
					!definitelyDoesNotContainGlobalFlag(argumentNode) {
					return
				}

				reportNode := callee
				if ast.IsPropertyAccessExpression(callee) {
					reportNode = callee.AsPropertyAccessExpression().Name().AsNode()
				} else if ast.IsElementAccessExpression(callee) {
					reportNode = callee.AsElementAccessExpression().ArgumentExpression
				}

				if isStringLiteral(argumentNode) {
					pattern := argumentNode.AsStringLiteral().Text
					if _, err := regexp2.Compile(pattern, regexp2.ECMAScript); err != nil {
						return
					}
					regExpLiteral := buildStringToRegExpLiteral(pattern)
					reportWithFix(reportNode, node, objectNode, argumentNode, func(objectCode string, _ string) string {
						return regExpLiteral + ".exec(" + objectCode + ")"
					})
					return
				}

				switch argumentTypes {
				case argumentTypeRegExp:
					reportWithFix(reportNode, node, objectNode, argumentNode, func(objectCode string, argumentCode string) string {
						return argumentCode + ".exec(" + objectCode + ")"
					})
				case argumentTypeString:
					reportWithFix(reportNode, node, objectNode, argumentNode, func(objectCode string, argumentCode string) string {
						return "RegExp(" + argumentCode + ").exec(" + objectCode + ")"
					})
				}
			},
		}
	},
}
