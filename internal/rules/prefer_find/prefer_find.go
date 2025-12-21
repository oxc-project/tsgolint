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

// filterExpressionData holds information about a filter call expression
// that may need to be transformed to a find call.
type filterExpressionData struct {
	filterNode               *ast.Node
	isBracketSyntaxForFilter bool
}

var PreferFindRule = rule.Rule{
	Name: "prefer-find",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		// isArrayish checks if the type is a possibly nullable array/tuple or union thereof.
		// It returns true only if all non-nullish parts of the type are arrays or tuples.
		isArrayish := func(t *checker.Type) bool {
			isAtLeastOneArrayishComponent := false

			for _, unionPart := range utils.UnionTypeParts(t) {
				// Skip null and undefined types
				if utils.IsTypeFlagSet(unionPart, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
					continue
				}

				// For intersection types, all parts must be arrays/tuples
				intersectionParts := utils.IntersectionTypeParts(unionPart)
				isArrayOrIntersectionThereof := utils.Every(intersectionParts, func(intersectionPart *checker.Type) bool {
					return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, intersectionPart)
				})

				if !isArrayOrIntersectionThereof {
					// There is a non-array, non-nullish type component
					return false
				}

				isAtLeastOneArrayishComponent = true
			}

			return isAtLeastOneArrayishComponent
		}

		// skipChainExpression skips over chain expressions to get to the underlying expression
		skipChainExpression := func(node *ast.Node) *ast.Node {
			// In typescript-go AST, there's no separate ChainExpression node type
			// The chain is represented via optional tokens on member/call expressions
			return node
		}

		// getStaticStringValue attempts to get a static string value from a node.
		// It handles string literals, numeric literals, identifiers with const initializers, etc.
		var getStaticStringValue func(node *ast.Node) (string, bool)
		getStaticStringValue = func(node *ast.Node) (string, bool) {
			node = ast.SkipParentheses(node)

			switch node.Kind {
			case ast.KindStringLiteral:
				return node.AsStringLiteral().Text, true

			case ast.KindNoSubstitutionTemplateLiteral:
				// Template literal like `0` or `at` - use Text() method
				return node.AsNoSubstitutionTemplateLiteral().Text, true

			case ast.KindNumericLiteral:
				return node.AsNumericLiteral().Text, true

			case ast.KindBigIntLiteral:
				text := node.AsBigIntLiteral().Text
				// Remove trailing 'n' from bigint literal
				if strings.HasSuffix(text, "n") {
					text = text[:len(text)-1]
				}
				return text, true

			case ast.KindPrefixUnaryExpression:
				prefixExpr := node.AsPrefixUnaryExpression()
				if prefixExpr.Operator == ast.KindMinusToken {
					// Handle negative numbers: -0, -0n, -0.12635678
					if prefixExpr.Operand.Kind == ast.KindNumericLiteral {
						return "-" + prefixExpr.Operand.AsNumericLiteral().Text, true
					}
					if prefixExpr.Operand.Kind == ast.KindBigIntLiteral {
						text := prefixExpr.Operand.AsBigIntLiteral().Text
						if strings.HasSuffix(text, "n") {
							text = text[:len(text)-1]
						}
						return "-" + text, true
					}
				}
				return "", false

			case ast.KindIdentifier:
				// Check for NaN first
				if node.AsIdentifier().Text == "NaN" {
					return "NaN", true
				}
				// Try to resolve the identifier to its value
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil || symbol.ValueDeclaration == nil {
					return "", false
				}
				if symbol.ValueDeclaration.Kind == ast.KindVariableDeclaration {
					varDecl := symbol.ValueDeclaration.AsVariableDeclaration()
					if varDecl.Initializer != nil {
						// Recursively get the value
						return getStaticStringValue(varDecl.Initializer)
					}
				}
				return "", false

			case ast.KindTemplateExpression:
				// Handle template literals with substitutions like `${0}`
				templateExpr := node.AsTemplateExpression()
				// Only handle simple template literals without substitutions
				if len(templateExpr.TemplateSpans.Nodes) == 0 {
					return templateExpr.Head.AsTemplateHead().Text, true
				}
				// For complex template expressions with substitutions, try to evaluate
				// Only handle the case where all spans evaluate to static values
				result := templateExpr.Head.AsTemplateHead().Text
				for _, spanNode := range templateExpr.TemplateSpans.Nodes {
					span := spanNode.AsTemplateSpan()
					spanValue, ok := getStaticStringValue(span.Expression)
					if !ok {
						return "", false
					}
					result += spanValue
					if span.Literal.Kind == ast.KindTemplateMiddle {
						result += span.Literal.AsTemplateMiddle().Text
					} else if span.Literal.Kind == ast.KindTemplateTail {
						result += span.Literal.AsTemplateTail().Text
					}
				}
				return result, true
			}

			return "", false
		}

		// parseArrayFilterExpressions collects all filter call expressions that should be
		// transformed to find calls. It handles sequence expressions and conditional expressions.
		var parseArrayFilterExpressions func(expression *ast.Node) []filterExpressionData
		parseArrayFilterExpressions = func(expression *ast.Node) []filterExpressionData {
			node := skipChainExpression(expression)
			node = ast.SkipParentheses(node)

			// Handle sequence (comma) expressions - only the last expression matters
			if ast.IsBinaryExpression(node) && node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
				// In a sequence expression, only the last (rightmost) value matters
				return parseArrayFilterExpressions(node.AsBinaryExpression().Right)
			}

			// Handle conditional (ternary) expressions - both branches must be filter expressions
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

				// Accumulate the results from both sides
				return append(consequentResult, alternateResult...)
			}

			// Check if it's a call expression (but not optional: foo?.filter(...))
			if node.Kind == ast.KindCallExpression {
				callExpr := node.AsCallExpression()

				// Reject if it's an optional call: foo?.filter(...)
				if callExpr.QuestionDotToken != nil {
					return nil
				}

				callee := callExpr.Expression

				// Check if callee is a member expression (property or element access)
				if !ast.IsAccessExpression(callee) {
					return nil
				}

				// Check if the property name is "filter"
				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !found || propertyName != "filter" {
					return nil
				}

				// Determine if bracket syntax is used: arr["filter"](...) vs arr.filter(...)
				isBracketSyntax := callee.Kind == ast.KindElementAccessExpression

				// Get the filter property node for fixing
				var filterNode *ast.Node
				if callee.Kind == ast.KindPropertyAccessExpression {
					filterNode = callee.AsPropertyAccessExpression().Name()
				} else if callee.Kind == ast.KindElementAccessExpression {
					filterNode = callee.AsElementAccessExpression().ArgumentExpression
				} else {
					return nil
				}

				// Get the type of the object that filter is called on
				filteredObjectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				// Check if the object is an array type (possibly nullable)
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

		// isTreatedAsZeroByMemberAccess checks if a value would be treated as index 0
		// when used as an array member access index.
		// String(value) === '0' is the JavaScript behavior.
		// Note: -0 and -0n both become "0" when converted to string for property access.
		isTreatedAsZeroByMemberAccess := func(value string) bool {
			return value == "0" || value == "-0"
		}

		// isTreatedAsZeroByArrayAt checks if a value would be treated as index 0
		// when used with Array.prototype.at().
		// According to MDN: the index is converted to a number, then Math.trunc()'d.
		isTreatedAsZeroByArrayAt := func(value string) bool {
			// Handle NaN - converted to 0 by Math.trunc
			if value == "NaN" {
				return true
			}

			// Try to parse as a number
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				// If it's not a valid number, it becomes NaN, which becomes 0
				return true
			}

			// Math.trunc(num) === 0
			truncated := int64(num)
			return truncated == 0
		}

		// getObjectIfArrayAtZeroExpression checks if a call expression is .at(0) or .at(something that evaluates to 0)
		// and returns the object it's called on.
		getObjectIfArrayAtZeroExpression := func(node *ast.CallExpression) *ast.Node {
			// .at() should take exactly one argument
			if len(node.Arguments.Nodes) != 1 {
				return nil
			}

			callee := node.Expression

			// Check for property/element access expression
			if !ast.IsAccessExpression(callee) {
				return nil
			}

			// Check if it's an optional call: foo?.at(0) - not allowed
			if callee.Kind == ast.KindPropertyAccessExpression {
				propAccess := callee.AsPropertyAccessExpression()
				if propAccess.QuestionDotToken != nil {
					return nil
				}
			} else if callee.Kind == ast.KindElementAccessExpression {
				elemAccess := callee.AsElementAccessExpression()
				if elemAccess.QuestionDotToken != nil {
					return nil
				}
			}

			// Check if the property name is "at"
			propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
			if !found || propertyName != "at" {
				return nil
			}

			// Get the argument value
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

		// isMemberAccessOfZero checks if a computed member expression accesses index 0
		isMemberAccessOfZero := func(node *ast.ElementAccessExpression) bool {
			// Check for optional chaining: foo?.[0] - not allowed
			if node.QuestionDotToken != nil {
				return false
			}

			value, ok := getStaticStringValue(node.ArgumentExpression)
			if !ok {
				return false
			}

			return isTreatedAsZeroByMemberAccess(value)
		}

		// generateFixes creates the fixes to transform filter()[0] to find()
		generateFixes := func(
			filterExpressions []filterExpressionData,
			arrayNode *ast.Node,
			wholeExpressionNode *ast.Node,
		) []rule.RuleFix {
			fixes := []rule.RuleFix{}

			// Replace each "filter" with "find"
			for _, filterExpr := range filterExpressions {
				filterNodeRange := utils.TrimNodeTextRange(ctx.SourceFile, filterExpr.filterNode)
				if filterExpr.isBracketSyntaxForFilter {
					// For bracket syntax: arr["filter"](...) -> arr["find"](...)
					fixes = append(fixes, rule.RuleFixReplaceRange(filterNodeRange, "\"find\""))
				} else {
					// For dot syntax: arr.filter(...) -> arr.find(...)
					fixes = append(fixes, rule.RuleFixReplaceRange(filterNodeRange, "find"))
				}
			}

			// Remove the array element access ([0] or .at(0))
			// We need to find the token that starts the access (. or [) to preserve comments
			// Use scanner to find the next token after the array node
			s := scanner.GetScannerForSourceFile(ctx.SourceFile, arrayNode.End())
			accessTokenStart := s.TokenRange().Pos()
			wholeExprEnd := wholeExpressionNode.End()

			// Remove from the start of the access token to the end of the whole expression
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(accessTokenStart, wholeExprEnd)))

			return fixes
		}

		return rule.RuleListeners{
			// Handle .at(0) case: filteredResults.at(0)
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

			// Handle [0] case: filteredResults[0]
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
