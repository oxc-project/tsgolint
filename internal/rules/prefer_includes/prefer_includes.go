package prefer_includes

import (
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIncludes",
		Description: "Use 'includes()' method instead.",
	}
}

func buildPreferStringIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStringIncludes",
		Description: "Use `String#includes()` method with a string instead.",
	}
}

var PreferIncludesRule = rule.Rule{
	Name: "prefer-includes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		// Escape special characters for string literal
		// The pattern from regex already has escape sequences, we only need to escape apostrophes
		escapeString := func(s string) string {
			return strings.ReplaceAll(s, "'", "\\'")
		}

		// Check if a pattern string contains only simple literal characters
		// The TypeScript version uses a proper ECMAScript regex parser (@eslint-community/regexpp),
		// but we just check for unescaped metacharacters.
		isSimpleLiteralPattern := func(pattern string) bool {
			prevRune := rune(0)
			for _, ch := range pattern {
				if prevRune != '\\' {
					switch ch {
					case '.', '*', '+', '?', '|', '^', '$', '[', ']', '(', ')', '{', '}':
						return false
					}
				}
				prevRune = ch
			}
			return true
		}

		// Extract pattern from regex literal: /bar/ -> "bar"
		extractRegexLiteralPattern := func(node *ast.Node) string {
			if node.Kind != ast.KindRegularExpressionLiteral {
				return ""
			}

			regexLit := node.AsRegularExpressionLiteral()
			text := regexLit.Text // e.g., "/bar/" or "/bar/i"

			// Parse the regex literal: /pattern/flags
			if len(text) < 3 || text[0] != '/' {
				return ""
			}

			// Find the closing /
			lastSlash := -1
			for i := len(text) - 1; i > 0; i-- {
				if text[i] == '/' {
					lastSlash = i
					break
				}
			}
			if lastSlash <= 0 {
				return ""
			}

			pattern := text[1:lastSlash]
			flags := text[lastSlash+1:]

			// Reject patterns with any flags
			if len(flags) > 0 {
				return ""
			}

			// Validate pattern compiles and is simple literal
			if _, err := regexp2.Compile(pattern, regexp2.ECMAScript); err != nil {
				return ""
			}

			if !isSimpleLiteralPattern(pattern) {
				return ""
			}

			return pattern
		}

		// Extract pattern from RegExp constructor: new RegExp('bar') -> "bar"
		extractRegExpConstructorPattern := func(node *ast.Node) string {
			if node.Kind != ast.KindNewExpression {
				return ""
			}

			newExpr := node.AsNewExpression()
			if newExpr.Expression.Kind != ast.KindIdentifier {
				return ""
			}

			if newExpr.Expression.AsIdentifier().Text != "RegExp" {
				return ""
			}

			args := node.Arguments()
			if len(args) == 0 || args[0].Kind != ast.KindStringLiteral {
				return ""
			}

			pattern := args[0].AsStringLiteral().Text

			// Validate pattern compiles and is simple literal
			if _, err := regexp2.Compile(pattern, regexp2.ECMAScript); err != nil {
				return ""
			}

			if !isSimpleLiteralPattern(pattern) {
				return ""
			}

			return pattern
		}

		// Resolve pattern from variable: const p = /bar/; p.test(...) -> "bar"
		resolveVariablePattern := func(node *ast.Node) string {
			if !ast.IsIdentifier(node) {
				return ""
			}

			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil || symbol.ValueDeclaration == nil {
				return ""
			}

			valueDecl := symbol.ValueDeclaration
			if valueDecl.Kind != ast.KindVariableDeclaration {
				return ""
			}

			varDecl := valueDecl.AsVariableDeclaration()
			if varDecl.Initializer == nil {
				return ""
			}

			// Try regex literal: const p = /bar/
			if pattern := extractRegexLiteralPattern(varDecl.Initializer); pattern != "" {
				return pattern
			}

			// Try RegExp constructor: const p = new RegExp('bar')
			return extractRegExpConstructorPattern(varDecl.Initializer)
		}

		// Resolve a regex pattern from a node, handling:
		// 1. Direct regex literal: /bar/
		// 2. Variable reference: const p = /bar/; p.test(...)
		// 3. RegExp constructor: new RegExp('bar')
		//
		// The TypeScript ESLint version uses getStaticValue() from ESLint to evaluate
		// more complex cases like string concatenation. We handle the common patterns
		// by resolving symbols to their initializers.
		resolveRegexPattern := func(node *ast.Node) string {
			// Try direct regex literal
			if pattern := extractRegexLiteralPattern(node); pattern != "" {
				return pattern
			}

			// Try variable reference
			return resolveVariablePattern(node)
		}

		// Check if two function declarations have matching parameter signatures
		// Compares the full text of each parameter (name, type annotation, and optionality)
		hasSameParameters := func(declA, declB *ast.Node) bool {
			if !ast.IsFunctionLike(declA) || !ast.IsFunctionLike(declB) {
				return false
			}

			paramsA := declA.Parameters()
			paramsB := declB.Parameters()

			if len(paramsA) != len(paramsB) {
				return false
			}

			// Compare the text of each parameter
			for i := range paramsA {
				paramA := paramsA[i]
				paramB := paramsB[i]

				sourceFileA := ast.GetSourceFileOfNode(paramA)
				sourceFileB := ast.GetSourceFileOfNode(paramB)
				if sourceFileA == nil || sourceFileB == nil {
					return false
				}

				textA := sourceFileA.Text()[paramA.Pos():paramA.End()]
				textB := sourceFileB.Text()[paramB.Pos():paramB.End()]
				if textA != textB {
					return false
				}
			}

			return true
		}

		// Check if the indexOf symbol has a compatible includes method
		// Verifies that for every indexOf declaration, there exists an includes
		// declaration on the same type with matching parameters
		indexOfHasCompatibleIncludes := func(indexOfSymbol *ast.Symbol) bool {
			if indexOfSymbol == nil || indexOfSymbol.Declarations == nil || len(indexOfSymbol.Declarations) == 0 {
				return false
			}

			// Check every declaration of indexOf to ensure it has a compatible includes
			for _, indexOfDecl := range indexOfSymbol.Declarations {
				// Get the type that contains this indexOf declaration
				typeDecl := indexOfDecl.Parent
				if typeDecl == nil {
					return false
				}

				// Get the type at this location
				t := ctx.TypeChecker.GetTypeAtLocation(typeDecl)
				if t == nil {
					return false
				}

				// Check if this type has an includes method
				includesSymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "includes")
				if includesSymbol == nil || includesSymbol.Declarations == nil {
					return false
				}

				// Check if any includes declaration has the same parameters as this indexOf
				hasMatchingIncludes := false
				for _, includesDecl := range includesSymbol.Declarations {
					if hasSameParameters(indexOfDecl, includesDecl) {
						hasMatchingIncludes = true
						break
					}
				}

				if !hasMatchingIncludes {
					return false
				}
			}

			return true
		}

		// Check if the node is a number literal with specific value
		// Handles both numeric literals (0) and prefix unary expressions (-1)
		isNumberLiteral := func(node *ast.Node, value int) bool {
			if node.Kind == ast.KindNumericLiteral {
				if num, err := strconv.Atoi(node.AsNumericLiteral().Text); err == nil {
					return num == value
				}
			}

			// Handle negative numbers as prefix unary expressions: -1 is PrefixUnaryExpression(-, 1)
			if node.Kind == ast.KindPrefixUnaryExpression {
				prefixExpr := node.AsPrefixUnaryExpression()
				if prefixExpr.Operator == ast.KindMinusToken && prefixExpr.Operand.Kind == ast.KindNumericLiteral {
					if num, err := strconv.Atoi(prefixExpr.Operand.AsNumericLiteral().Text); err == nil {
						return -num == value
					}
				}
			}

			return false
		}

		// Determine if this is a positive check (should use includes)
		// Patterns: !== -1, != -1, > -1, >= 0
		isPositiveCheck := func(binaryExpr *ast.BinaryExpression) bool {
			operator := binaryExpr.OperatorToken.Kind
			right := binaryExpr.Right

			switch operator {
			case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindGreaterThanToken:
				return isNumberLiteral(right, -1)
			case ast.KindGreaterThanEqualsToken:
				return isNumberLiteral(right, 0)
			}
			return false
		}

		// Determine if this is a negative check (should use !includes)
		// Patterns: === -1, == -1, <= -1, < 0
		isNegativeCheck := func(binaryExpr *ast.BinaryExpression) bool {
			operator := binaryExpr.OperatorToken.Kind
			right := binaryExpr.Right

			switch operator {
			case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindLessThanEqualsToken:
				return isNumberLiteral(right, -1)
			case ast.KindLessThanToken:
				return isNumberLiteral(right, 0)
			}
			return false
		}

		return rule.RuleListeners{
			// Handle: /regex/.test(str) -> str.includes('literal')
			ast.KindCallExpression: func(node *ast.Node) {
				if node.Kind != ast.KindCallExpression {
					return
				}

				callExpr := node.AsCallExpression()

				// Check if it's a member access (e.g., /regex/.test or pattern.test)
				if callExpr.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}

				propAccess := callExpr.Expression.AsPropertyAccessExpression()

				// Check if the method name is "test"
				nameNode := propAccess.Name()
				if !ast.IsIdentifier(nameNode) || nameNode.AsIdentifier().Text != "test" {
					return
				}

				// Check if there's exactly one argument
				if len(callExpr.Arguments.Nodes) != 1 {
					return
				}

				// The regex is either:
				// 1. Direct literal: /bar/.test(a)
				// 2. Variable: pattern.test(a) where pattern = /bar/ or new RegExp('bar')
				regexNode := propAccess.Expression
				pattern := resolveRegexPattern(regexNode)
				if pattern == "" {
					return
				}

				// Check the argument type has includes method
				argument := callExpr.Arguments.Nodes[0]
				argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
				if argType == nil {
					return
				}

				includesSymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, argType, "includes")
				if includesSymbol == nil || includesSymbol.Declarations == nil {
					return
				}

				fixes := []rule.RuleFix{}

				// Check if argument needs wrapping in parentheses for .includes() call
				// Member access (.) has high precedence, so we only need parens for expressions
				// that would parse incorrectly, like: a + b.includes() vs (a + b).includes()
				// Safe types: literals, identifiers, member/element access, calls, already-parenthesized
				needsParens := false
				switch argument.Kind {
				case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral,
					ast.KindNoSubstitutionTemplateLiteral, ast.KindPropertyAccessExpression,
					ast.KindCallExpression, ast.KindElementAccessExpression, ast.KindParenthesizedExpression:
					needsParens = false
				default:
					needsParens = true
				}

				// Use TrimNodeTextRange to preserve leading whitespace
				callExprRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

				// Remove everything before the argument
				fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(callExprRange.Pos(), argument.Pos())))

				// Remove everything after the argument
				fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(argument.End(), callExprRange.End())))

				// Add parentheses if needed
				if needsParens {
					fixes = append(fixes, rule.RuleFix{
						Range: core.NewTextRange(argument.Pos(), argument.Pos()),
						Text:  "(",
					})
					fixes = append(fixes, rule.RuleFix{
						Range: core.NewTextRange(argument.End(), argument.End()),
						Text:  ")",
					})
				}

				// Add .includes('pattern') after the argument
				escapedPattern := escapeString(pattern)
				fixes = append(fixes, rule.RuleFix{
					Range: core.NewTextRange(argument.End(), argument.End()),
					Text:  ".includes('" + escapedPattern + "')",
				})

				ctx.ReportNodeWithFixes(node, buildPreferStringIncludesMessage(), func() []rule.RuleFix { return fixes })
			},
			// Handle: array.indexOf(item) !== -1 -> array.includes(item)
			ast.KindBinaryExpression: func(node *ast.Node) {
				if node.Kind != ast.KindBinaryExpression {
					return
				}

				binaryExpr := node.AsBinaryExpression()
				left := binaryExpr.Left

				// Skip if left side is not a call expression
				// Handle: array.indexOf(item) !== -1
				if left.Kind != ast.KindCallExpression {
					return
				}

				callExpr := left.AsCallExpression()

				// Check if it's a member access (e.g., array.indexOf)
				if callExpr.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}

				propAccess := callExpr.Expression.AsPropertyAccessExpression()

				// Check if the method name is "indexOf"
				nameNode := propAccess.Name()
				if !ast.IsIdentifier(nameNode) || nameNode.AsIdentifier().Text != "indexOf" {
					return
				}

				// Check if it's a positive or negative check
				isPositive := isPositiveCheck(binaryExpr)
				isNegative := isNegativeCheck(binaryExpr)

				if !isPositive && !isNegative {
					return
				}

				// Get the symbol of indexOf method
				indexOfSymbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if indexOfSymbol == nil {
					return
				}

				// Check if the type has includes method with matching parameters
				if !indexOfHasCompatibleIncludes(indexOfSymbol) {
					return
				}

				fixes := []rule.RuleFix{}

				// Replace "indexOf" with "includes"
				indexOfRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				fixes = append(fixes, rule.RuleFixReplaceRange(indexOfRange, "includes"))

				// Remove the comparison part (e.g., " !== -1")
				comparisonStart := callExpr.End()
				comparisonEnd := binaryExpr.End()
				fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(comparisonStart, comparisonEnd)))

				// If negative check, add "!" before the call expression
				// Use TrimNodeTextRange to get the actual start without leading trivia
				if isNegative {
					callExprRange := utils.TrimNodeTextRange(ctx.SourceFile, left)
					fixes = append(fixes, rule.RuleFix{
						Range: core.NewTextRange(callExprRange.Pos(), callExprRange.Pos()),
						Text:  "!",
					})
				}

				ctx.ReportNodeWithFixes(node, buildPreferIncludesMessage(), func() []rule.RuleFix { return fixes })
			},
		}
	},
}
