package prefer_find

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
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

type staticValueKind uint8

const (
	staticValueString staticValueKind = iota
	staticValueNumber
	staticValueBoolean
	staticValueNull
	staticValueBigInt
	staticValueUndefined
	staticValueSymbol
)

type staticValue struct {
	kind        staticValueKind
	stringValue string
	numberValue float64
	boolValue   bool
	bigIntValue *big.Int
}

type filterExpressionData struct {
	filterNode               *ast.Node
	isBracketSyntaxForFilter bool
}

var PreferFindRule = rule.Rule{
	Name: "prefer-find",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var getStaticValue func(node *ast.Node, visited map[*ast.Symbol]struct{}) (staticValue, bool)

		numberToString := func(v float64) string {
			if math.IsNaN(v) {
				return "NaN"
			}
			if math.IsInf(v, 1) {
				return "Infinity"
			}
			if math.IsInf(v, -1) {
				return "-Infinity"
			}
			if v == 0 {
				return "0"
			}
			return strconv.FormatFloat(v, 'g', -1, 64)
		}

		toNumberFromString := func(s string) float64 {
			s = strings.TrimSpace(s)
			if s == "" {
				return 0
			}

			if s == "Infinity" || s == "+Infinity" {
				return math.Inf(1)
			}
			if s == "-Infinity" {
				return math.Inf(-1)
			}

			base := 0
			trimmed := s
			sign := ""
			if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "-") {
				sign = trimmed[:1]
				trimmed = trimmed[1:]
			}
			if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
				base = 16
				trimmed = trimmed[2:]
			} else if strings.HasPrefix(trimmed, "0b") || strings.HasPrefix(trimmed, "0B") {
				base = 2
				trimmed = trimmed[2:]
			} else if strings.HasPrefix(trimmed, "0o") || strings.HasPrefix(trimmed, "0O") {
				base = 8
				trimmed = trimmed[2:]
			}

			if base != 0 {
				if trimmed == "" {
					return math.NaN()
				}
				intVal, ok := new(big.Int).SetString(sign+trimmed, base)
				if !ok {
					return math.NaN()
				}
				floatVal, _ := new(big.Float).SetInt(intVal).Float64()
				return floatVal
			}

			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return math.NaN()
			}
			return n
		}

		toNumber := func(value staticValue) (float64, bool) {
			switch value.kind {
			case staticValueNumber:
				return value.numberValue, true
			case staticValueString:
				return toNumberFromString(value.stringValue), true
			case staticValueBoolean:
				if value.boolValue {
					return 1, true
				}
				return 0, true
			case staticValueNull:
				return 0, true
			case staticValueUndefined:
				return math.NaN(), true
			case staticValueBigInt, staticValueSymbol:
				return 0, false
			}

			return 0, false
		}

		staticValueToPropertyName := func(value staticValue) (string, bool) {
			switch value.kind {
			case staticValueString:
				return value.stringValue, true
			case staticValueNumber:
				return numberToString(value.numberValue), true
			case staticValueBoolean:
				if value.boolValue {
					return "true", true
				}
				return "false", true
			case staticValueNull:
				return "null", true
			case staticValueUndefined:
				return "undefined", true
			case staticValueBigInt:
				return value.bigIntValue.String(), true
			}

			return "", false
		}

		isTreatedAsZeroByArrayAt := func(value staticValue) bool {
			if value.kind == staticValueSymbol {
				return false
			}

			asNumber, ok := toNumber(value)
			if !ok {
				return false
			}

			if math.IsNaN(asNumber) {
				return true
			}

			return math.Trunc(asNumber) == 0
		}

		isTreatedAsZeroByMemberAccess := func(value staticValue) bool {
			switch value.kind {
			case staticValueString:
				return value.stringValue == "0"
			case staticValueNumber:
				return value.numberValue == 0
			case staticValueBigInt:
				return value.bigIntValue.Sign() == 0
			}

			return false
		}

		parseBigIntLiteral := func(text string) (staticValue, bool) {
			trimmed := strings.TrimSpace(text)
			trimmed = strings.TrimSuffix(trimmed, "n")
			trimmed = strings.ReplaceAll(trimmed, "_", "")
			if trimmed == "" {
				return staticValue{}, false
			}

			result, ok := new(big.Int).SetString(trimmed, 0)
			if !ok {
				return staticValue{}, false
			}
			return staticValue{kind: staticValueBigInt, bigIntValue: result}, true
		}

		getConstInitializer := func(symbol *ast.Symbol) (*ast.Node, bool) {
			if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) {
				return nil, false
			}

			if !ast.IsVariableDeclarationList(symbol.ValueDeclaration.Parent) || symbol.ValueDeclaration.Parent.Flags&ast.NodeFlagsConst == 0 {
				return nil, false
			}

			declaration := symbol.ValueDeclaration.AsVariableDeclaration()
			if declaration.Initializer == nil {
				return nil, false
			}

			return declaration.Initializer, true
		}

		getStaticValue = func(node *ast.Node, visited map[*ast.Symbol]struct{}) (staticValue, bool) {
			node = ast.SkipParentheses(node)
			if node == nil {
				return staticValue{}, false
			}

			switch node.Kind {
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				return staticValue{kind: staticValueString, stringValue: node.Text()}, true
			case ast.KindNumericLiteral:
				value, err := strconv.ParseFloat(strings.ReplaceAll(node.AsNumericLiteral().Text, "_", ""), 64)
				if err != nil {
					return staticValue{}, false
				}
				return staticValue{kind: staticValueNumber, numberValue: value}, true
			case ast.KindTrueKeyword:
				return staticValue{kind: staticValueBoolean, boolValue: true}, true
			case ast.KindFalseKeyword:
				return staticValue{kind: staticValueBoolean, boolValue: false}, true
			case ast.KindNullKeyword:
				return staticValue{kind: staticValueNull}, true
			case ast.KindBigIntLiteral:
				return parseBigIntLiteral(node.Text())
			case ast.KindIdentifier:
				identifier := node.AsIdentifier().Text
				switch identifier {
				case "undefined":
					return staticValue{kind: staticValueUndefined}, true
				case "NaN":
					return staticValue{kind: staticValueNumber, numberValue: math.NaN()}, true
				case "Infinity":
					return staticValue{kind: staticValueNumber, numberValue: math.Inf(1)}, true
				}

				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return staticValue{}, false
				}
				if _, ok := visited[symbol]; ok {
					return staticValue{}, false
				}

				initializer, ok := getConstInitializer(symbol)
				if !ok {
					return staticValue{}, false
				}

				visited[symbol] = struct{}{}
				defer delete(visited, symbol)
				return getStaticValue(initializer, visited)
			case ast.KindPrefixUnaryExpression:
				prefix := node.AsPrefixUnaryExpression()
				operandValue, ok := getStaticValue(prefix.Operand, visited)
				if !ok {
					return staticValue{}, false
				}

				switch prefix.Operator {
				case ast.KindMinusToken:
					if operandValue.kind == staticValueBigInt {
						result := new(big.Int).Set(operandValue.bigIntValue)
						result.Neg(result)
						return staticValue{kind: staticValueBigInt, bigIntValue: result}, true
					}
					asNumber, ok := toNumber(operandValue)
					if !ok {
						return staticValue{}, false
					}
					return staticValue{kind: staticValueNumber, numberValue: -asNumber}, true
				case ast.KindPlusToken:
					asNumber, ok := toNumber(operandValue)
					if !ok {
						return staticValue{}, false
					}
					return staticValue{kind: staticValueNumber, numberValue: asNumber}, true
				}
			case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
				return getStaticValue(node.Expression(), visited)
			}

			if ast.IsCallExpression(node) {
				callExpression := node.AsCallExpression()
				if callExpression.QuestionDotToken != nil {
					return staticValue{}, false
				}

				callee := ast.SkipParentheses(callExpression.Expression)
				if ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Symbol" {
					return staticValue{kind: staticValueSymbol}, true
				}

				if ast.IsPropertyAccessExpression(callee) {
					propertyAccess := callee.AsPropertyAccessExpression()
					object := ast.SkipParentheses(propertyAccess.Expression)
					if ast.IsIdentifier(object) && object.AsIdentifier().Text == "Symbol" && propertyAccess.Name().Text() == "for" {
						return staticValue{kind: staticValueSymbol}, true
					}
				}
			}

			return staticValue{}, false
		}

		isStaticMemberAccessOfValue := func(node *ast.Node, value string) bool {
			node = ast.SkipParentheses(node)
			if node == nil {
				return false
			}

			if ast.IsPropertyAccessExpression(node) {
				name := node.AsPropertyAccessExpression().Name()
				return name != nil && name.Text() == value
			}

			if ast.IsElementAccessExpression(node) {
				argument := node.AsElementAccessExpression().ArgumentExpression
				propertyValue, ok := getStaticValue(argument, map[*ast.Symbol]struct{}{})
				if !ok {
					return false
				}
				propertyName, ok := staticValueToPropertyName(propertyValue)
				return ok && propertyName == value
			}

			return false
		}

		isArrayish := func(t *checker.Type) bool {
			isAtLeastOneArrayishComponent := false

			for _, unionPart := range utils.UnionTypeParts(t) {
				if utils.IsTypeNullType(unionPart) || utils.IsTypeUndefinedType(unionPart) {
					continue
				}

				isArrayOrIntersectionThereof := true
				for _, intersectionPart := range utils.IntersectionTypeParts(unionPart) {
					if !checker.Checker_isArrayType(ctx.TypeChecker, intersectionPart) && !checker.IsTupleType(intersectionPart) {
						isArrayOrIntersectionThereof = false
						break
					}
				}

				if !isArrayOrIntersectionThereof {
					return false
				}

				isAtLeastOneArrayishComponent = true
			}

			return isAtLeastOneArrayishComponent
		}

		var parseArrayFilterExpressions func(expression *ast.Node) []filterExpressionData
		parseArrayFilterExpressions = func(expression *ast.Node) []filterExpressionData {
			node := ast.SkipParentheses(expression)
			if node == nil {
				return nil
			}

			if ast.IsCommaExpression(node) {
				lastExpression := node.AsBinaryExpression().Right
				return parseArrayFilterExpressions(lastExpression)
			}

			if node.Kind == ast.KindConditionalExpression {
				conditionalExpression := node.AsConditionalExpression()
				consequentResult := parseArrayFilterExpressions(conditionalExpression.WhenTrue)
				if len(consequentResult) == 0 {
					return nil
				}

				alternateResult := parseArrayFilterExpressions(conditionalExpression.WhenFalse)
				if len(alternateResult) == 0 {
					return nil
				}

				return append(consequentResult, alternateResult...)
			}

			if ast.IsCallExpression(node) {
				callExpression := node.AsCallExpression()
				if callExpression.QuestionDotToken != nil {
					return nil
				}

				callee := ast.SkipParentheses(callExpression.Expression)
				if !ast.IsPropertyAccessExpression(callee) && !ast.IsElementAccessExpression(callee) {
					return nil
				}

				if !isStaticMemberAccessOfValue(callee, "filter") {
					return nil
				}

				filteredObjectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())
				if !isArrayish(filteredObjectType) {
					return nil
				}

				if ast.IsPropertyAccessExpression(callee) {
					return []filterExpressionData{{
						filterNode:               callee.AsPropertyAccessExpression().Name().AsNode(),
						isBracketSyntaxForFilter: false,
					}}
				}

				return []filterExpressionData{{
					filterNode:               callee.AsElementAccessExpression().ArgumentExpression,
					isBracketSyntaxForFilter: true,
				}}
			}

			return nil
		}

		getObjectIfArrayAtZeroExpression := func(node *ast.Node) *ast.Node {
			callExpression := node.AsCallExpression()
			if len(callExpression.Arguments.Nodes) != 1 {
				return nil
			}

			if callExpression.QuestionDotToken != nil {
				return nil
			}

			callee := ast.SkipParentheses(callExpression.Expression)
			if callee == nil || (!ast.IsPropertyAccessExpression(callee) && !ast.IsElementAccessExpression(callee)) {
				return nil
			}

			if !isStaticMemberAccessOfValue(callee, "at") {
				return nil
			}

			atArgument := callExpression.Arguments.Nodes[0]
			atArgumentValue, ok := getStaticValue(atArgument, map[*ast.Symbol]struct{}{})
			if ok && isTreatedAsZeroByArrayAt(atArgumentValue) {
				return callee.Expression()
			}

			return nil
		}

		isMemberAccessOfZero := func(node *ast.Node) bool {
			if !ast.IsElementAccessExpression(node) {
				return false
			}

			elementAccessExpression := node.AsElementAccessExpression()
			if elementAccessExpression.QuestionDotToken != nil {
				return false
			}

			propertyValue, ok := getStaticValue(elementAccessExpression.ArgumentExpression, map[*ast.Symbol]struct{}{})
			return ok && isTreatedAsZeroByMemberAccess(propertyValue)
		}

		findArrayElementAccessStart := func(arrayNode *ast.Node, wholeExpression *ast.Node) (int, bool) {
			arrayRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(arrayNode))
			searchStart := arrayRange.End()
			searchEnd := wholeExpression.End()
			text := ctx.SourceFile.Text()

			for i := searchStart; i < searchEnd; {
				r, size := utf8.DecodeRuneInString(text[i:])
				if unicode.IsSpace(r) {
					i += size
					continue
				}

				if text[i] == '/' && i+1 < searchEnd {
					if text[i+1] == '/' {
						i += 2
						for i < searchEnd && text[i] != '\n' && text[i] != '\r' {
							i++
						}
						continue
					}
					if text[i+1] == '*' {
						i += 2
						for i+1 < searchEnd && !(text[i] == '*' && text[i+1] == '/') {
							i++
						}
						if i+1 < searchEnd {
							i += 2
						}
						continue
					}
				}

				if text[i] == '.' || text[i] == '[' {
					return i, true
				}

				i += size
			}

			return 0, false
		}

		generateFixToRemoveArrayElementAccess := func(arrayNode *ast.Node, wholeExpressionBeingFlagged *ast.Node) (rule.RuleFix, bool) {
			start, ok := findArrayElementAccessStart(arrayNode, wholeExpressionBeingFlagged)
			if !ok {
				return rule.RuleFix{}, false
			}

			return rule.RuleFixRemoveRange(core.NewTextRange(start, wholeExpressionBeingFlagged.End())), true
		}

		generateFixToReplaceFilterWithFind := func(filterExpression filterExpressionData) rule.RuleFix {
			replacement := "find"
			if filterExpression.isBracketSyntaxForFilter {
				replacement = `"find"`
			}

			return rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, filterExpression.filterNode), replacement)
		}

		reportPreferFind := func(node *ast.Node, arrayNode *ast.Node, filterExpressions []filterExpressionData) {
			removeFix, ok := generateFixToRemoveArrayElementAccess(arrayNode, node)
			if !ok {
				return
			}

			ctx.ReportNodeWithSuggestions(node, buildPreferFindMessage(), func() []rule.RuleSuggestion {
				fixes := make([]rule.RuleFix, 0, len(filterExpressions)+1)
				for _, filterExpression := range filterExpressions {
					fixes = append(fixes, generateFixToReplaceFilterWithFind(filterExpression))
				}
				fixes = append(fixes, removeFix)

				return []rule.RuleSuggestion{{
					Message:  buildPreferFindSuggestionMessage(),
					FixesArr: fixes,
				}}
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				object := getObjectIfArrayAtZeroExpression(node)
				if object == nil {
					return
				}

				filterExpressions := parseArrayFilterExpressions(object)
				if len(filterExpressions) == 0 {
					return
				}

				reportPreferFind(node, object, filterExpressions)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if !isMemberAccessOfZero(node) {
					return
				}

				object := node.AsElementAccessExpression().Expression
				filterExpressions := parseArrayFilterExpressions(object)
				if len(filterExpressions) == 0 {
					return
				}

				reportPreferFind(node, object, filterExpressions)
			},
		}
	},
}
