package prefer_optional_chain

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferOptionalChain",
		Description: "Prefer using an optional chain expression instead, as it's more concise and easier to read.",
	}
}

func buildOptionalChainSuggestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "optionalChainSuggest",
		Description: "Change to an optional chain.",
	}
}

// OperandType represents what kind of operand we're dealing with
type OperandType int

const (
	OperandTypeInvalid             OperandType = iota
	OperandTypePlain                           // foo
	OperandTypeNotEqualNull                    // foo != null
	OperandTypeNotStrictEqualNull              // foo !== null
	OperandTypeNotStrictEqualUndef             // foo !== undefined
	OperandTypeNotEqualBoth                    // foo != null (covers both)
	OperandTypeNot                             // !foo
	OperandTypeNegatedAndOperand               // !foo in && chains (special handling needed)
	OperandTypeTypeofCheck                     // typeof foo !== 'undefined'
	OperandTypeComparison                      // foo.bar == 0 (comparison at end of chain)
	OperandTypeEqualNull                       // foo == null (inverted check in && chains)
	OperandTypeStrictEqualNull                 // foo === null (inverted check in && chains)
	OperandTypeStrictEqualUndef                // foo === undefined (inverted check in && chains)
)

// isNullishCheckType returns true if the operand type represents any kind of null/undefined check.
// This includes both strict (===, !==) and loose (==, !=) null checks, as well as typeof checks.
func isNullishCheckType(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull ||
		typ == OperandTypeNotStrictEqualUndef ||
		typ == OperandTypeNotEqualBoth ||
		typ == OperandTypeStrictEqualNull ||
		typ == OperandTypeEqualNull ||
		typ == OperandTypeStrictEqualUndef ||
		typ == OperandTypeTypeofCheck
}

// isStrictNullishCheck returns true if the operand type is a strict (!==) nullish check.
// Used to detect incomplete nullish check patterns.
func isStrictNullishCheck(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull || typ == OperandTypeNotStrictEqualUndef
}

// isExplicitNullishCheck returns true if the operand type is any explicit nullish check
// (not a plain truthiness check). Includes strict, loose, and typeof checks.
func isExplicitNullishCheck(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull ||
		typ == OperandTypeNotStrictEqualUndef ||
		typ == OperandTypeNotEqualBoth ||
		typ == OperandTypeTypeofCheck
}

// isTrailingComparisonType returns true if the operand type could be a trailing comparison/check.
// This includes strict null checks, loose null checks, and value comparisons.
// Used when determining which operands to exclude from guard checks.
func isTrailingComparisonType(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull ||
		typ == OperandTypeNotStrictEqualUndef ||
		typ == OperandTypeNotEqualBoth ||
		typ == OperandTypeComparison
}

// isComparisonOrNullCheck returns true if the operand type is a comparison or any null/undefined check.
// Used to determine if an operand should be treated as a trailing comparison in the output.
func isComparisonOrNullCheck(typ OperandType) bool {
	return typ == OperandTypeComparison || isNullishCheckType(typ)
}

// isNullishComparison checks if an OperandTypeComparison operand is actually comparing to null/undefined.
// This is used in OR chains where property null checks like `a.b == null` are parsed as OperandTypeComparison
// but should still be treated as nullish checks for chain building purposes.
func isNullishComparison(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return false
	}
	unwrapped := unwrapParentheses(op.node)
	if !ast.IsBinaryExpression(unwrapped) {
		return false
	}
	binExpr := unwrapped.AsBinaryExpression()
	binOp := binExpr.OperatorToken.Kind

	// Only == and === null/undefined checks are nullish comparisons in OR chains
	// != and !== are the opposite (checking if NOT null)
	if binOp != ast.KindEqualsEqualsToken && binOp != ast.KindEqualsEqualsEqualsToken {
		return false
	}

	left := unwrapParentheses(binExpr.Left)
	right := unwrapParentheses(binExpr.Right)

	isLeftNullish := left.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(left)
	isRightNullish := right.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(right)

	return isLeftNullish || isRightNullish
}

// isStrictNullComparison checks if an OperandTypeComparison operand is comparing to null (not undefined).
// This is used to distinguish between null checks and undefined checks in OR chains.
func isStrictNullComparison(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return false
	}
	unwrapped := unwrapParentheses(op.node)
	if !ast.IsBinaryExpression(unwrapped) {
		return false
	}
	binExpr := unwrapped.AsBinaryExpression()
	binOp := binExpr.OperatorToken.Kind

	// Only === null checks (not undefined)
	if binOp != ast.KindEqualsEqualsEqualsToken && binOp != ast.KindEqualsEqualsToken {
		return false
	}

	left := unwrapParentheses(binExpr.Left)
	right := unwrapParentheses(binExpr.Right)

	// Check only for null keyword, NOT undefined
	isLeftNull := left.Kind == ast.KindNullKeyword
	isRightNull := right.Kind == ast.KindNullKeyword

	return isLeftNull || isRightNull
}

// isOrChainNullishCheck returns true if the operand is any kind of nullish check that can be
// used in an OR chain. This includes:
// - OperandTypeStrictEqualNull (foo === null)
// - OperandTypeStrictEqualUndef (foo === undefined)
// - OperandTypeEqualNull (foo == null)
// - OperandTypeComparison when comparing to null/undefined
// Used to allow extending OR chains through call expressions for nullish comparison patterns.
func isOrChainNullishCheck(op Operand) bool {
	switch op.typ {
	case OperandTypeStrictEqualNull, OperandTypeStrictEqualUndef, OperandTypeEqualNull:
		return true
	case OperandTypeComparison:
		return isNullishComparison(op)
	default:
		return false
	}
}

// Operand represents a parsed operand in a logical chain
type Operand struct {
	typ          OperandType
	node         *ast.Node
	comparedExpr *ast.Node // The expression being checked (e.g., 'foo' in 'foo !== null')
}

// NodeComparisonResult indicates how two nodes compare
type NodeComparisonResult int

const (
	NodeEqual    NodeComparisonResult = iota
	NodeSubset                        // left is a subset of right (foo vs foo.bar)
	NodeSuperset                      // left is a superset of right
	NodeInvalid                       // incomparable
)

// unwrapParentheses unwraps parenthesized expressions
func unwrapParentheses(n *ast.Node) *ast.Node {
	for ast.IsParenthesizedExpression(n) {
		n = n.AsParenthesizedExpression().Expression
	}
	return n
}

// unwrapForComparison unwraps both parentheses AND non-null assertions AND type assertions
// Used for operand comparison where we want foo.bar! to match foo.bar
// and (foo as Type) to match foo, and (<Type>foo) to match foo
func unwrapForComparison(n *ast.Node) *ast.Node {
	for {
		if ast.IsParenthesizedExpression(n) {
			n = n.AsParenthesizedExpression().Expression
		} else if ast.IsNonNullExpression(n) {
			n = n.AsNonNullExpression().Expression
		} else if n.Kind == ast.KindAsExpression {
			n = n.AsAsExpression().Expression
		} else if n.Kind == ast.KindTypeAssertionExpression {
			n = n.AsTypeAssertion().Expression
		} else {
			break
		}
	}
	return n
}

// getNormalizedNodeText builds a normalized text representation of an AST node for comparison.
// This replaces text-based removeAllParens and related functions with proper AST traversal.
// Normalization:
// - Unwraps parenthesized expressions (grouping parens are removed)
// - Normalizes optional chaining (?. → .)
// - Strips non-null assertions (!)
// - Strips type assertions (as Type, <Type>)
// - Preserves call/element access structure
func (processor *chainProcessor) getNormalizedNodeText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	// Check cache first
	if cached, ok := processor.normalizedCache[node]; ok {
		return cached
	}

	var result strings.Builder
	processor.buildNormalizedText(node, &result)
	normalized := result.String()

	// Cache the result
	processor.normalizedCache[node] = normalized
	return normalized
}

// buildNormalizedText recursively builds normalized text from an AST node
func (processor *chainProcessor) buildNormalizedText(n *ast.Node, result *strings.Builder) {
	if n == nil {
		return
	}

	switch {
	case ast.IsParenthesizedExpression(n):
		// Unwrap parentheses - just process the inner expression
		processor.buildNormalizedText(n.AsParenthesizedExpression().Expression, result)

	case ast.IsNonNullExpression(n):
		// Strip non-null assertions - just process the inner expression
		processor.buildNormalizedText(n.AsNonNullExpression().Expression, result)

	case n.Kind == ast.KindAsExpression:
		// Strip type assertions - just process the expression being asserted
		processor.buildNormalizedText(n.AsAsExpression().Expression, result)

	case n.Kind == ast.KindTypeAssertionExpression:
		// Strip angle bracket type assertions - just process the inner expression
		processor.buildNormalizedText(n.AsTypeAssertion().Expression, result)

	case ast.IsPropertyAccessExpression(n):
		propAccess := n.AsPropertyAccessExpression()
		// Build the base expression
		processor.buildNormalizedText(propAccess.Expression, result)
		// Always use regular dot (normalize ?. to .)
		result.WriteByte('.')
		// Add the property name
		nameRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, propAccess.Name())
		result.WriteString(processor.sourceText[nameRange.Pos():nameRange.End()])

	case ast.IsElementAccessExpression(n):
		elemAccess := n.AsElementAccessExpression()
		// Build the base expression
		processor.buildNormalizedText(elemAccess.Expression, result)
		// Always use regular bracket (normalize ?.[ to [)
		result.WriteByte('[')
		// Add the argument expression (use raw text to preserve computed expressions)
		argRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, elemAccess.ArgumentExpression)
		result.WriteString(processor.sourceText[argRange.Pos():argRange.End()])
		result.WriteByte(']')

	case ast.IsCallExpression(n):
		callExpr := n.AsCallExpression()
		// Build the callee expression
		processor.buildNormalizedText(callExpr.Expression, result)
		// Add type arguments if present
		if callExpr.TypeArguments != nil && len(callExpr.TypeArguments.Nodes) > 0 {
			result.WriteByte('<')
			typeArgsStart := callExpr.TypeArguments.Loc.Pos()
			typeArgsEnd := callExpr.TypeArguments.Loc.End()
			result.WriteString(processor.sourceText[typeArgsStart:typeArgsEnd])
			result.WriteByte('>')
		}
		// Always use regular paren (normalize ?.( to ()
		result.WriteByte('(')
		// Add arguments
		if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
			argsStart := callExpr.Arguments.Loc.Pos()
			callEnd := n.End()
			result.WriteString(processor.sourceText[argsStart : callEnd-1])
		}
		result.WriteByte(')')

	default:
		// Base case - identifiers, literals, or other expressions
		// Get the raw text
		textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
		result.WriteString(processor.sourceText[textRange.Pos():textRange.End()])
	}
}

// hasOptionalChaining checks if an expression contains optional chaining (?.)
// This is an AST-based check that replaces text-based strings.Contains(text, "?.")
func hasOptionalChaining(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch {
	case ast.IsParenthesizedExpression(node):
		return hasOptionalChaining(node.AsParenthesizedExpression().Expression)

	case ast.IsNonNullExpression(node):
		return hasOptionalChaining(node.AsNonNullExpression().Expression)

	case ast.IsPropertyAccessExpression(node):
		propAccess := node.AsPropertyAccessExpression()
		if propAccess.QuestionDotToken != nil {
			return true
		}
		return hasOptionalChaining(propAccess.Expression)

	case ast.IsElementAccessExpression(node):
		elemAccess := node.AsElementAccessExpression()
		if elemAccess.QuestionDotToken != nil {
			return true
		}
		return hasOptionalChaining(elemAccess.Expression)

	case ast.IsCallExpression(node):
		callExpr := node.AsCallExpression()
		if callExpr.QuestionDotToken != nil {
			return true
		}
		return hasOptionalChaining(callExpr.Expression)
	}

	return false
}

// isInsideJSX checks if a node is inside a JSX context
// In JSX, foo && foo.bar has different semantics than foo?.bar
// (foo && foo.bar returns false/null/undefined, while foo?.bar returns undefined)
func isInsideJSX(node *ast.Node) bool {
	current := node
	for current != nil {
		if ast.IsJsxExpression(current) ||
			ast.IsJsxAttribute(current) ||
			ast.IsJsxAttributes(current) ||
			ast.IsJsxElement(current) ||
			ast.IsJsxSelfClosingElement(current) ||
			ast.IsJsxOpeningElement(current) ||
			ast.IsJsxClosingElement(current) ||
			ast.IsJsxFragment(current) {
			return true
		}
		current = current.Parent
	}
	return false
}

// getBaseIdentifier extracts the base identifier from an expression chain
// For foo.bar.baz, returns foo
// For (foo as any).bar, returns foo
// For foo!.bar, returns foo
func getBaseIdentifier(node *ast.Node) *ast.Node {
	current := node
	for {
		if ast.IsPropertyAccessExpression(current) {
			current = current.AsPropertyAccessExpression().Expression
		} else if ast.IsElementAccessExpression(current) {
			current = current.AsElementAccessExpression().Expression
		} else if ast.IsCallExpression(current) {
			current = current.AsCallExpression().Expression
		} else if ast.IsNonNullExpression(current) {
			current = current.AsNonNullExpression().Expression
		} else if ast.IsParenthesizedExpression(current) {
			current = current.AsParenthesizedExpression().Expression
		} else if current.Kind == ast.KindAsExpression {
			// Type assertion - get the expression being asserted
			current = current.AsAsExpression().Expression
		} else {
			// Base case - return the current node
			return current
		}
	}
}

// hasSideEffects checks if an expression has side effects
// This includes: ++, --, yield, assignment operators
func hasSideEffects(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for prefix increment/decrement
	if ast.IsPrefixUnaryExpression(node) {
		op := node.AsPrefixUnaryExpression().Operator
		if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return true
		}
	}

	// Check for postfix increment/decrement
	if node.Kind == ast.KindPostfixUnaryExpression {
		return true // postfix ++ and -- always have side effects
	}

	// Check for yield expressions
	if ast.IsYieldExpression(node) {
		return true
	}

	// NOTE: We do NOT check await expressions here for side effects
	// Await expressions can be safely included in property chains like (await foo).bar
	// The check for problematic await patterns like "(await foo) && (await foo).bar"
	// is handled separately in compareNodes

	// Check for assignment operators
	if ast.IsBinaryExpression(node) {
		op := node.AsBinaryExpression().OperatorToken.Kind
		// Assignment operators
		if op == ast.KindEqualsToken ||
			op == ast.KindPlusEqualsToken ||
			op == ast.KindMinusEqualsToken ||
			op == ast.KindAsteriskEqualsToken ||
			op == ast.KindSlashEqualsToken {
			return true
		}
	}

	// Recursively check children
	if ast.IsPropertyAccessExpression(node) {
		return hasSideEffects(node.AsPropertyAccessExpression().Expression)
	}
	if ast.IsElementAccessExpression(node) {
		elem := node.AsElementAccessExpression()
		return hasSideEffects(elem.Expression) || hasSideEffects(elem.ArgumentExpression)
	}
	if ast.IsCallExpression(node) {
		return hasSideEffects(node.AsCallExpression().Expression)
	}
	if ast.IsParenthesizedExpression(node) {
		return hasSideEffects(node.AsParenthesizedExpression().Expression)
	}

	return false
}

// textRange represents a range in the source text
type textRange struct {
	start int
	end   int
}

// ChainPart represents a component part of a chain expression for reconstruction
type ChainPart struct {
	text        string
	optional    bool
	requiresDot bool
	isPrivate   bool // true if this part is a private identifier (#foo)
	hasNonNull  bool // true if this part has a non-null assertion (!)
}

// TypeInfo caches computed type information for a node to avoid repeated type checker calls
type TypeInfo struct {
	parts            []*checker.Type
	hasNull          bool
	hasUndefined     bool
	hasVoid          bool
	hasAny           bool
	hasUnknown       bool
	hasBoolLiteral   bool
	hasNumLiteral    bool
	hasStrLiteral    bool
	hasBigIntLiteral bool
	// Additional flags for shouldSkipByType
	hasBigIntLike  bool
	hasBooleanLike bool
	hasNumberLike  bool
	hasStringLike  bool
}

// chainProcessor manages state and provides helper methods for processing optional chain candidates
type chainProcessor struct {
	ctx                rule.RuleContext
	opts               PreferOptionalChainOptions
	sourceText         string // Cached source text to avoid repeated calls
	seenLogicals       map[*ast.Node]bool
	processedAndRanges []textRange
	seenLogicalRanges  map[textRange]bool
	reportedRanges     map[textRange]bool
	// Caches to avoid repeated computations
	typeCache       map[*ast.Node]*TypeInfo
	normalizedCache map[*ast.Node]string
	flattenCache    map[*ast.Node][]ChainPart
	callSigCache    map[*ast.Node]map[string]string
}

// newChainProcessor creates a new chainProcessor with initialized state
func newChainProcessor(ctx rule.RuleContext, opts PreferOptionalChainOptions) *chainProcessor {
	return &chainProcessor{
		ctx:                ctx,
		opts:               opts,
		sourceText:         ctx.SourceFile.Text(), // Cache source text once
		seenLogicals:       make(map[*ast.Node]bool, 16),
		processedAndRanges: make([]textRange, 0, 8),
		seenLogicalRanges:  make(map[textRange]bool, 16),
		reportedRanges:     make(map[textRange]bool, 8),
		typeCache:          make(map[*ast.Node]*TypeInfo, 32),
		normalizedCache:    make(map[*ast.Node]string, 32),
		flattenCache:       make(map[*ast.Node][]ChainPart, 16),
		callSigCache:       make(map[*ast.Node]map[string]string, 8),
	}
}

// getTypeInfo returns cached type information for a node, computing it if not already cached.
// This avoids repeated calls to GetTypeAtLocation and UnionTypeParts for the same node.
func (processor *chainProcessor) getTypeInfo(node *ast.Node) *TypeInfo {
	if info, ok := processor.typeCache[node]; ok {
		return info
	}

	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	parts := utils.UnionTypeParts(nodeType)

	info := &TypeInfo{
		parts: parts,
	}

	// Use UnionTypeParts to detect nullability from the type's constituent parts.
	// This is more reliable than string parsing and handles all type kinds correctly.
	for _, part := range parts {
		if utils.IsTypeNullType(part) {
			info.hasNull = true
		}
		if utils.IsTypeUndefinedType(part) {
			info.hasUndefined = true
		}
		if utils.IsTypeVoidType(part) {
			info.hasVoid = true
		}
		if utils.IsTypeAnyType(part) {
			info.hasAny = true
		}
		if utils.IsTypeUnknownType(part) {
			info.hasUnknown = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLiteral) {
			info.hasBoolLiteral = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral) {
			info.hasNumLiteral = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) {
			info.hasStrLiteral = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLiteral) {
			info.hasBigIntLiteral = true
		}
		// Additional flags for shouldSkipByType
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLike) {
			info.hasBigIntLike = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLike) {
			info.hasBooleanLike = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLike) {
			info.hasNumberLike = true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsStringLike) {
			info.hasStringLike = true
		}
	}

	processor.typeCache[node] = info
	return info
}

// extractCallSignatures extracts call signatures from a node for comparison.
// Returns a map of "base expression" -> "full call text" for all call expressions in the node.
func (processor *chainProcessor) extractCallSignatures(node *ast.Node) map[string]string {
	// Check cache first
	if cached, ok := processor.callSigCache[node]; ok {
		return cached
	}

	signatures := make(map[string]string, 4)
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ast.IsCallExpression(n) {
			call := n.AsCallExpression()
			// Get base expression text
			exprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, call.Expression)
			exprText := processor.sourceText[exprRange.Pos():exprRange.End()]
			// Get full call text (including args and type args)
			fullRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			fullText := processor.sourceText[fullRange.Pos():fullRange.End()]
			signatures[exprText] = fullText
			visit(call.Expression)
		} else if ast.IsPropertyAccessExpression(n) {
			visit(n.AsPropertyAccessExpression().Expression)
		} else if ast.IsElementAccessExpression(n) {
			visit(n.AsElementAccessExpression().Expression)
		} else if ast.IsNonNullExpression(n) {
			visit(n.AsNonNullExpression().Expression)
		}
	}
	visit(node)

	// Cache the result
	processor.callSigCache[node] = signatures
	return signatures
}

// compareNodes compares two nodes to determine their relationship.
func (processor *chainProcessor) compareNodes(left, right *ast.Node) NodeComparisonResult {
	// Check for side effects in either expression
	// Example: foo[x++] && foo[x++].bar -> Cannot convert (x++ has side effects)
	// Example: foo[yield x] && foo[yield x].bar -> Cannot convert (yield has side effects)
	if hasSideEffects(left) || hasSideEffects(right) {
		return NodeInvalid
	}

	// Unwrap parentheses for AST-based checks
	leftUnwrapped := unwrapParentheses(left)

	// Check if the left operand is a CallExpression or NewExpression at the base level
	// If so, we cannot safely chain because calling the function/constructor multiple times may have side effects
	// Example: getFoo() && getFoo().bar -> Cannot convert (getFoo() might have side effects)
	// Example: new Date() && new Date().getTime() -> Cannot convert (different instances)
	// EXCEPTION 1: If the expression already contains optional chaining (?.), it's safe to extend
	// Example: foo?.() || foo?.().bar -> CAN convert to foo?.()?.bar (single evaluation)
	// EXCEPTION 2: If the call is a method call on an object (e.g., foo.bar()), allow it because:
	//   - We've already verified the base (foo.bar) in a previous operand
	//   - The original code already calls the method multiple times
	//   - Converting to optional chain actually REDUCES call count
	// Also check for literal expressions (arrays, objects, functions, classes) which create new instances
	// Example: [] && [].length -> Cannot convert (different arrays)
	// Example: (class Foo {}) && class Foo {}.name -> Cannot convert (different classes)

	// Allow call expressions if they contain optional chaining (already safe)
	if !hasOptionalChaining(left) {
		// For CallExpressions and NewExpressions, only block if the callee is a standalone identifier
		// (like getFoo() or new Date()). Allow method calls like foo.bar() because:
		// 1. The base object was already verified in a previous chain operand
		// 2. The original code already calls the method in both operands
		// 3. Converting to optional chain reduces the number of calls
		//
		// EXCEPTION: If the root of the expression is a NewExpression, block it because
		// each `new X()` creates a different instance.
		// Example: new Map().get('a') && new Map().get('a').what -> SHOULD NOT CONVERT

		// Find the root of the expression (traverse through property/element/call chains)
		rootExpr := leftUnwrapped
		for {
			unwrapped := unwrapParentheses(rootExpr)
			if ast.IsPropertyAccessExpression(unwrapped) {
				rootExpr = unwrapped.AsPropertyAccessExpression().Expression
			} else if ast.IsElementAccessExpression(unwrapped) {
				rootExpr = unwrapped.AsElementAccessExpression().Expression
			} else if ast.IsCallExpression(unwrapped) {
				rootExpr = unwrapped.AsCallExpression().Expression
			} else {
				break
			}
		}

		// Check if the root is a NewExpression (problematic - creates different instances)
		isRootedInNew := false
		if rootExpr != nil {
			unwrappedRoot := unwrapParentheses(rootExpr)
			isRootedInNew = ast.IsNewExpression(unwrappedRoot)
		}

		isStandaloneCall := false
		if ast.IsCallExpression(leftUnwrapped) {
			callee := unwrapParentheses(leftUnwrapped.AsCallExpression().Expression)
			// Standalone call if callee is just an identifier (not a property/element access)
			isStandaloneCall = ast.IsIdentifier(callee)
		} else if ast.IsNewExpression(leftUnwrapped) {
			// new expressions are always problematic (create different instances)
			isStandaloneCall = true
		}

		if isStandaloneCall || isRootedInNew ||
			ast.IsArrayLiteralExpression(leftUnwrapped) ||
			ast.IsObjectLiteralExpression(leftUnwrapped) ||
			ast.IsFunctionExpression(leftUnwrapped) ||
			ast.IsArrowFunction(leftUnwrapped) ||
			ast.IsClassExpression(leftUnwrapped) ||
			// JSX elements are always new instances, like object/array literals
			ast.IsJsxElement(leftUnwrapped) ||
			ast.IsJsxSelfClosingElement(leftUnwrapped) ||
			ast.IsJsxFragment(leftUnwrapped) ||
			leftUnwrapped.Kind == ast.KindTemplateExpression ||
			leftUnwrapped.Kind == ast.KindAwaitExpression {
			return NodeInvalid
		}
	}

	// Extract call signatures from both nodes BEFORE normalization
	leftSigs := processor.extractCallSignatures(left)
	rightSigs := processor.extractCallSignatures(right)

	// Check if any call expressions have matching base but different signatures
	for baseExpr, leftSig := range leftSigs {
		if rightSig, exists := rightSigs[baseExpr]; exists && leftSig != rightSig {
			// Same function called with different arguments or type parameters
			return NodeInvalid
		}
	}

	// Use AST-based normalization instead of text manipulation
	// This handles: parentheses unwrapping, optional chaining normalization,
	// type assertion stripping, and non-null assertion stripping
	leftNormalized := processor.getNormalizedNodeText(left)
	rightNormalized := processor.getNormalizedNodeText(right)

	if leftNormalized == rightNormalized {
		// If normalized forms are equal but one has optional chaining and the other doesn't,
		// they represent the same path but with different nullability handling.
		// For most cases where they ARE the same normalized expression, we consider them equal.
		// This handles cases like:
		//   foo?.bar?.baz !== null && typeof foo.bar.baz !== 'undefined'
		// where both refer to the same property path.
		//
		// The only exception is when one side has optional CALL syntax that would change behavior:
		// Example: (foo?.a)() vs foo.a() - the optional chaining affects whether the call happens
		// But: foo?.bar?.baz vs foo.bar.baz - same property path, safe to treat as equal
		return NodeEqual
	}

	// Check if left is a subset of right (foo vs foo.bar or foo vs foo<T>())
	// Use normalized text for comparison
	if strings.HasPrefix(rightNormalized, leftNormalized) {
		remainder := strings.TrimPrefix(rightNormalized, leftNormalized)
		// Allow ., [, (, <, and ! (for non-null assertions) as valid continuations
		if len(remainder) > 0 && (remainder[0] == '.' || remainder[0] == '[' || remainder[0] == '(' || remainder[0] == '<' || remainder[0] == '!') {
			// Allow optional chaining in left when extending to right
			// Example: foo?.bar (left) vs foo.bar.baz (right) is valid
			// We normalize both before comparison, so the optional chaining is already stripped
			// The key insight: if left has ?. and is a subset of right, we're building a longer chain
			// The optional chaining in left will be replaced by the final chain anyway
			return NodeSubset
		}
	}

	// Check if right is a subset of left
	if strings.HasPrefix(leftNormalized, rightNormalized) {
		remainder := strings.TrimPrefix(leftNormalized, rightNormalized)
		// Allow ., [, (, <, and ! (for non-null assertions) as valid continuations
		if len(remainder) > 0 && (remainder[0] == '.' || remainder[0] == '[' || remainder[0] == '(' || remainder[0] == '<' || remainder[0] == '!') {
			return NodeSuperset
		}
	}

	return NodeInvalid
}

// includesNullish checks if a type includes nullish flags.
// Also returns true for 'any' and 'unknown' types since they can be nullish at runtime.
func (processor *chainProcessor) includesNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasNull || info.hasUndefined || info.hasAny || info.hasUnknown
}

// includesExplicitNullish checks if a type includes explicit nullish types (null | undefined).
// This does NOT return true for 'any' or 'unknown' types.
// Used to determine if autofix is safe when allowPotentiallyUnsafe is false.
func (processor *chainProcessor) includesExplicitNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasNull || info.hasUndefined
}

// typeIsAnyOrUnknown checks if a type is any or unknown (where we can't determine exact nullishness).
func (processor *chainProcessor) typeIsAnyOrUnknown(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	// If the type has any or unknown and nothing else that's non-nullish
	if len(info.parts) == 0 {
		return false
	}
	// Check if all parts are any/unknown
	for _, part := range info.parts {
		if !utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return false
		}
	}
	return true
}

// typeIncludesNull checks if a type includes the null type specifically.
func (processor *chainProcessor) typeIncludesNull(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	// any and unknown can be null at runtime
	return info.hasNull || info.hasAny || info.hasUnknown
}

// typeIncludesUndefined checks if a type includes the undefined type specifically.
func (processor *chainProcessor) typeIncludesUndefined(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	// any and unknown can be undefined at runtime
	return info.hasUndefined || info.hasAny || info.hasUnknown
}

// wouldChangeReturnType checks if converting to optional chaining would change the return type.
// This happens when the type includes falsy non-nullish values (false, 0, ", 0n)
// but does NOT include null/undefined.
func (processor *chainProcessor) wouldChangeReturnType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)

	hasNullish := info.hasNull || info.hasUndefined
	hasFalsyNonNullish := info.hasBoolLiteral || info.hasNumLiteral || info.hasStrLiteral || info.hasBigIntLiteral

	// Return type changes if we have falsy non-nullish but no nullish
	return hasFalsyNonNullish && !hasNullish
}

// hasVoidType checks if the type includes void (always falsy, but not nullish).
// void can cause issues when converting && to optional chaining
// because && checks truthiness, while ?. only checks for null/undefined.
// Example: x && x() where x is void | (() => void)
// - Original: if x is void (falsy), returns void (no call)
// - Converted: x?.() would try to call void (TypeError!)
//
// Note: We ONLY check for void here. Other falsy values like false/0/""
// are handled by the existing checkBoolean/checkNumber/checkString options.
// void is special because it's ALWAYS falsy (never truthy like true/1/"x")
func (processor *chainProcessor) hasVoidType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasVoid
}

// isOrChainComparisonSafe checks if a comparison operand in an OR chain is safe to convert to optional chaining.
// For OR chains with !foo || foo.bar OP VALUE:
// - != X with literals (0, 1, '123', true, false, {}, []) - SAFE (undefined != X evaluates correctly)
// - !== X with literals - SAFE (undefined !== X is always true for non-undefined literals)
// - === undefined - SAFE (undefined === undefined is true)
// - == null or == undefined - SAFE (covers both null and undefined)
// - === X where X is NOT undefined - NOT SAFE (undefined === 'foo' is false, changes behavior)
// - != null or != undefined - NOT SAFE (undefined != null is false in JS!)
func (processor *chainProcessor) isOrChainComparisonSafe(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return true // Not a comparison, skip this check
	}

	unwrapped := op.node
	for ast.IsParenthesizedExpression(unwrapped) {
		unwrapped = unwrapped.AsParenthesizedExpression().Expression
	}

	if !ast.IsBinaryExpression(unwrapped) {
		return true // Not a binary expression, can't analyze
	}

	binExpr := unwrapped.AsBinaryExpression()
	operator := binExpr.OperatorToken.Kind

	// Determine the value being compared (not the property access)
	left := binExpr.Left
	right := binExpr.Right

	// Unwrap parentheses
	for ast.IsParenthesizedExpression(left) {
		left = left.AsParenthesizedExpression().Expression
	}
	for ast.IsParenthesizedExpression(right) {
		right = right.AsParenthesizedExpression().Expression
	}

	// Determine which side is the value (not the property/element/call access)
	var value *ast.Node
	if ast.IsPropertyAccessExpression(left) || ast.IsElementAccessExpression(left) || ast.IsCallExpression(left) {
		value = right
	} else if ast.IsPropertyAccessExpression(right) || ast.IsElementAccessExpression(right) || ast.IsCallExpression(right) {
		value = left
	} else {
		// Neither side is a property access, can't determine
		return true
	}

	// Helper to check if value is null keyword
	isNull := value.Kind == ast.KindNullKeyword

	// Helper to check if value is undefined
	isUndefined := (ast.IsIdentifier(value) && value.AsIdentifier().Text == "undefined") || ast.IsVoidExpression(value)

	// Helper to check if value is a literal (safe for comparisons)
	isLiteral := value.Kind == ast.KindNumericLiteral ||
		value.Kind == ast.KindStringLiteral ||
		value.Kind == ast.KindTrueKeyword ||
		value.Kind == ast.KindFalseKeyword ||
		value.Kind == ast.KindObjectLiteralExpression ||
		value.Kind == ast.KindArrayLiteralExpression

	switch operator {
	case ast.KindExclamationEqualsEqualsToken:
		// !== is SAFE for literals and null (undefined !== X is true for non-undefined X)
		// Example: !foo || foo.bar !== 0
		// Original: if foo is falsy, returns true (due to !foo); if truthy, returns foo.bar !== 0
		// Converted: foo?.bar !== 0 -> if foo is nullish, returns undefined !== 0 = true
		//                           -> if foo is falsy non-nullish (0, "", false), foo.bar is undefined, returns true
		// These are equivalent for literals and null!
		//
		// !== null is SAFE: undefined !== null is true
		// !== undefined is NOT SAFE: undefined !== undefined is false (DIFFERENT from original which returns true)
		return isLiteral || isNull

	case ast.KindEqualsEqualsEqualsToken:
		// === is only safe if comparing to undefined
		// Example: !foo || foo.bar === undefined -> foo?.bar === undefined (safe)
		// Example: !foo || foo.bar === 'foo' -> NOT safe
		//   - if foo is nullish: !foo is true, returns true
		//   - converted: foo?.bar === 'foo' -> undefined === 'foo' = false (DIFFERENT!)
		return isUndefined

	case ast.KindExclamationEqualsToken:
		// != is safe for literals, but NOT for null/undefined
		// Example: !foo || foo.bar != 0 -> foo?.bar != 0 (safe: undefined != 0 is true)
		// Example: !foo || foo.bar != null -> NOT safe
		//   - Original: if foo is nullish, returns true; if defined, returns foo.bar != null
		//   - Converted: foo?.bar != null -> if foo is nullish, undefined != null is FALSE!
		if isNull || isUndefined {
			return false
		}
		// Also reject variables (undeclared identifiers could be undefined at runtime)
		if ast.IsIdentifier(value) && !isLiteral {
			return false
		}
		return isLiteral

	case ast.KindEqualsEqualsToken:
		// == is safe for null/undefined (covers both), but risky for other values
		// Example: !foo || foo.bar == null -> foo?.bar == null (safe)
		// Example: !foo || foo.bar == 0 -> risky (type coercion)
		return isNull || isUndefined
	}

	// Other operators (>, <, >=, <=, instanceof, in) - generally safe
	return true
}

// shouldSkipByType checks if we should skip this operand based on type-checking options.
// For plain operands, check the base identifier's type.
// For example, in (foo as any).bar, we want to check foo's type, not any.
func (processor *chainProcessor) shouldSkipByType(node *ast.Node) bool {
	baseNode := getBaseIdentifier(node)
	info := processor.getTypeInfo(baseNode)

	// If requireNullish is true and the type explicitly includes null/undefined,
	// do NOT skip - the chain is specifically checking for nullish values.
	// The check* options are for "loose boolean" cases where we're checking
	// falsy non-nullish values (like empty string, 0, false).
	if processor.opts.RequireNullish && (info.hasNull || info.hasUndefined) {
		return false
	}

	// Skip nullish types - they're always allowed
	// We need to check each non-nullish type against the options
	// If the type has any flag that should be skipped (based on options), return true

	// Check any type
	if info.hasAny && !processor.opts.CheckAny {
		return true
	}
	// Check bigint type
	if info.hasBigIntLike && !processor.opts.CheckBigInt {
		return true
	}
	// Check boolean type
	if info.hasBooleanLike && !processor.opts.CheckBoolean {
		return true
	}
	// Check number type
	if info.hasNumberLike && !processor.opts.CheckNumber {
		return true
	}
	// Check string type
	if info.hasStringLike && !processor.opts.CheckString {
		return true
	}
	// Check unknown type
	if info.hasUnknown && !processor.opts.CheckUnknown {
		return true
	}

	return false
}

// flattenForFix flattens a chain expression to its component parts for reconstruction
func (processor *chainProcessor) flattenForFix(node *ast.Node) []ChainPart {
	// Check cache first
	if cached, ok := processor.flattenCache[node]; ok {
		return cached
	}

	parts := []ChainPart{}

	var visit func(n *ast.Node, parentIsNonNull bool)
	visit = func(n *ast.Node, parentIsNonNull bool) {
		switch {
		case ast.IsParenthesizedExpression(n):
			// Check if the inner expression requires parentheses
			inner := n.AsParenthesizedExpression().Expression
			// Keep parentheses around await, yield, and other expressions that need them
			if ast.IsAwaitExpression(inner) || ast.IsYieldExpression(inner) {
				// Keep the parentheses - get the full text including parens
				textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
				text := processor.sourceText[textRange.Pos():textRange.End()]

				parts = append(parts, ChainPart{
					text:        text,
					optional:    false,
					requiresDot: false,
				})
				return
			}
			// Otherwise skip parentheses and visit the inner expression
			visit(inner, parentIsNonNull)

		case ast.IsNonNullExpression(n):
			// Handle non-null assertions: foo!.bar
			// Visit the inner expression and mark it as having a non-null assertion
			nonNullExpr := n.AsNonNullExpression()
			visit(nonNullExpr.Expression, true)

		case ast.IsPropertyAccessExpression(n):
			propAccess := n.AsPropertyAccessExpression()
			visit(propAccess.Expression, false)
			nameRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, propAccess.Name())
			nameText := processor.sourceText[nameRange.Pos():nameRange.End()]

			// If this property access is wrapped in a NonNullExpression (parentIsNonNull),
			// append ! to the property name
			hasNonNull := parentIsNonNull
			if hasNonNull {
				nameText = nameText + "!"
			}

			// Check if this is a private identifier
			isPrivate := propAccess.Name().Kind == ast.KindPrivateIdentifier

			parts = append(parts, ChainPart{
				text:        nameText,
				optional:    propAccess.QuestionDotToken != nil,
				requiresDot: true,
				isPrivate:   isPrivate,
				hasNonNull:  hasNonNull,
			})

		case ast.IsElementAccessExpression(n):
			elemAccess := n.AsElementAccessExpression()
			visit(elemAccess.Expression, false)
			argRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, elemAccess.ArgumentExpression)
			argText := processor.sourceText[argRange.Pos():argRange.End()]

			// If this element access is wrapped in a NonNullExpression (parentIsNonNull),
			// we need to handle it, but element access already uses brackets
			// so we'll add the ! after the closing bracket
			hasNonNull := parentIsNonNull
			suffix := ""
			if hasNonNull {
				suffix = "!"
			}

			parts = append(parts, ChainPart{
				text:        "[" + argText + "]" + suffix,
				optional:    elemAccess.QuestionDotToken != nil,
				requiresDot: false,
				hasNonNull:  hasNonNull,
			})

		case ast.IsCallExpression(n):
			callExpr := n.AsCallExpression()
			visit(callExpr.Expression, false)

			// Get type arguments text if present - use the full TypeArguments list range
			typeArgsText := ""
			if callExpr.TypeArguments != nil && len(callExpr.TypeArguments.Nodes) > 0 {
				// Use the NodeList's Loc to get the full range including whitespace
				typeArgsStart := callExpr.TypeArguments.Loc.Pos()
				typeArgsEnd := callExpr.TypeArguments.Loc.End()
				typeArgsText = "<" + processor.sourceText[typeArgsStart:typeArgsEnd] + ">"
			}

			// Get the arguments text - extract from opening ( to closing )
			// to preserve comments, whitespace, and trailing commas
			// The call expression ends with ), so we need to find the ( and extract everything
			argsText := "()"
			if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
				// Use the NodeList's Loc.Pos() for the start (after the opening paren)
				// and the call expression's End()-1 for the end (before the closing paren)
				// Actually, extract from Arguments start to just before the closing paren
				argsStart := callExpr.Arguments.Loc.Pos()
				// The call expression's End() points to right after the closing )
				// So End()-1 is the ), and we want everything from argsStart to End()-1
				callEnd := n.End()
				argsText = "(" + processor.sourceText[argsStart:callEnd-1] + ")"
			}

			parts = append(parts, ChainPart{
				text:        typeArgsText + argsText,
				optional:    callExpr.QuestionDotToken != nil,
				requiresDot: false,
			})

		default:
			// Base case - identifier or other expression
			textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			text := processor.sourceText[textRange.Pos():textRange.End()]

			// If this base expression is wrapped in a NonNullExpression (parentIsNonNull),
			// the ! is already part of the text range
			// But if parentIsNonNull is true and this is an identifier, we should append !
			if parentIsNonNull && ast.IsIdentifier(n) {
				text = text + "!"
			}

			// Type assertions need parentheses when used as base of property access
			// foo as any -> (foo as any) when followed by .bar
			// <Type>foo -> (<Type>foo) when followed by .bar
			if n.Kind == ast.KindAsExpression || n.Kind == ast.KindTypeAssertionExpression {
				text = "(" + text + ")"
			}

			parts = append(parts, ChainPart{
				text:        text,
				optional:    false,
				requiresDot: false,
				hasNonNull:  parentIsNonNull,
			})
		}
	}

	visit(node, false)

	// Cache the result
	processor.flattenCache[node] = parts
	return parts
}

// buildOptionalChain builds optional chain code from parts.
// Returns empty string if the chain would result in invalid syntax (e.g., ?.#private).
// stripNonNullAssertions: if true, strip ! when the next part becomes optional (for OR chains)
//
//	if false, preserve ! assertions (for AND chains)
func (processor *chainProcessor) buildOptionalChain(parts []ChainPart, checkedLengths map[int]bool, callShouldBeOptional bool, stripNonNullAssertions bool) string {
	// Find the maximum checked length - parts at or before this index were part of the nullish check
	maxCheckedLength := 0
	for length := range checkedLengths {
		if length > maxCheckedLength {
			maxCheckedLength = length
		}
	}

	// First pass: determine which parts should be optional
	optionalParts := make([]bool, len(parts))
	for i, part := range parts {
		if i > 0 {
			// Priority 1: If we have an explicit check at this length, this part should be optional
			if checkedLengths[i] {
				optionalParts[i] = true
			} else if part.optional {
				// Priority 2: Keep existing optional chain from the checked expression
				// When we copy parts from the first operand (the check), we preserve their optional flags
				// Example: foo?.bar.baz != null && foo.bar?.baz.bam != null
				// - The first operand has foo?.bar, so parts[1].optional=true
				// - We should keep this optional, resulting in foo?.bar.baz?.bam
				optionalParts[i] = true
			} else {
				// Priority 3: Check for call expressions
				isCall := strings.HasPrefix(part.text, "(") || strings.HasPrefix(part.text, "<(")
				isLastPart := i == len(parts)-1
				if isCall && isLastPart && callShouldBeOptional {
					optionalParts[i] = true
				}
			}
		}

		// TypeScript doesn't allow optional chaining on private identifiers (?.#foo)
		// If we would make a private identifier optional, abort the fix
		if optionalParts[i] && part.isPrivate {
			return ""
		}
	}

	var result strings.Builder
	for i, part := range parts {
		partText := part.text

		// If the NEXT part is being made optional and this part has a non-null assertion (!),
		// strip the ! ONLY if:
		// 1. stripNonNullAssertions is true (OR chains)
		// 2. This part was part of the checked portion (not the extension)
		// e.g., !foo!.bar!.baz || !foo!.bar!.baz!.paz -> foo!.bar!.baz?.paz
		//       (strip ! from baz at index 2, max checked is 3, 2 < 3 so strip)
		// For AND chains (stripNonNullAssertions=false), preserve all ! assertions
		// e.g., foo! && foo!.bar! && foo!.bar!.baz -> foo!?.bar!?.baz
		if stripNonNullAssertions && i < len(parts)-1 && optionalParts[i+1] && part.hasNonNull && strings.HasSuffix(partText, "!") {
			// Only strip if this part is within the checked region
			if i < maxCheckedLength {
				partText = partText[:len(partText)-1]
			}
		}

		if i > 0 && optionalParts[i] {
			// Make this optional
			result.WriteString("?.")
		} else if i > 0 {
			// Not making it optional
			// If this part was within the checked region (i <= maxCheckedLength),
			// strip any existing ?. because the earlier check already validated it
			// Example: foo.bar.baz != null && foo?.bar?.baz.bam != null
			// - maxCheckedLength = 3 (from foo.bar.baz)
			// - For i=1,2 (bar, baz), don't use ?. even if part.optional is true
			// - The earlier foo.bar.baz check already validated these parts
			if part.optional && i > maxCheckedLength {
				// Keep existing optional chaining only if it's beyond the checked region
				result.WriteString("?.")
			} else if part.requiresDot {
				result.WriteString(".")
			}
			// For calls and element access, no separator needed (requiresDot is false)
		}
		result.WriteString(partText)
	}
	return result.String()
}

// containsOptionalChain checks if an expression contains any optional chains
func (processor *chainProcessor) containsOptionalChain(n *ast.Node) bool {
	unwrapped := unwrapParentheses(n)

	// Check if this node itself is an optional chain
	if ast.IsPropertyAccessExpression(unwrapped) {
		if unwrapped.AsPropertyAccessExpression().QuestionDotToken != nil {
			return true
		}
		// Recursively check the left side
		return processor.containsOptionalChain(unwrapped.AsPropertyAccessExpression().Expression)
	}
	if ast.IsElementAccessExpression(unwrapped) {
		if unwrapped.AsElementAccessExpression().QuestionDotToken != nil {
			return true
		}
		// Recursively check the left side
		return processor.containsOptionalChain(unwrapped.AsElementAccessExpression().Expression)
	}
	if ast.IsCallExpression(unwrapped) {
		callExpr := unwrapped.AsCallExpression()
		if callExpr.QuestionDotToken != nil {
			return true
		}
		// Recursively check the expression being called
		return processor.containsOptionalChain(callExpr.Expression)
	}
	if ast.IsBinaryExpression(unwrapped) {
		// Check both sides of binary expression
		binExpr := unwrapped.AsBinaryExpression()
		return processor.containsOptionalChain(binExpr.Left) || processor.containsOptionalChain(binExpr.Right)
	}

	return false
}

// parseOperand parses an operand to determine its type and what it's checking
func (processor *chainProcessor) parseOperand(node *ast.Node, isAndChain bool) Operand {
	// Unwrap parentheses AND non-null assertions for analysis
	// but keep original node for range calculation and fix generation
	unwrapped := unwrapForComparison(node)

	// Note: Private identifiers are handled in buildOptionalChain - we allow operands
	// with private identifiers here, and check during fix generation whether we'd
	// create invalid ?.#private syntax

	// Skip ONLY bare 'this' keyword (not this.foo)
	// Pattern: this && ... or !this || ...
	// Bare 'this' cannot be converted because it's not nullable in TypeScript
	// However, this.foo CAN be converted because the property might be nullable
	if unwrapped.Kind == ast.KindThisKeyword {
		return Operand{typ: OperandTypeInvalid, node: node}
	}

	// Extract the base expression for further checks
	baseExpr := unwrapped
	for {
		if ast.IsPropertyAccessExpression(baseExpr) {
			baseExpr = baseExpr.AsPropertyAccessExpression().Expression
		} else if ast.IsElementAccessExpression(baseExpr) {
			baseExpr = baseExpr.AsElementAccessExpression().Expression
		} else if ast.IsCallExpression(baseExpr) {
			baseExpr = baseExpr.AsCallExpression().Expression
		} else if ast.IsNonNullExpression(baseExpr) {
			baseExpr = baseExpr.AsNonNullExpression().Expression
		} else if ast.IsParenthesizedExpression(baseExpr) {
			baseExpr = baseExpr.AsParenthesizedExpression().Expression
		} else if baseExpr.Kind == ast.KindAsExpression {
			baseExpr = baseExpr.AsAsExpression().Expression
		} else if baseExpr.Kind == ast.KindTypeAssertionExpression {
			baseExpr = baseExpr.AsTypeAssertion().Expression
		} else {
			break
		}
	}
	// Note: We do NOT reject if baseExpr is 'this' - only if the whole unwrapped expression is bare 'this'
	// Example: this.bar && this.bar.baz -> this.bar?.baz is valid
	// Example: this && this.foo -> invalid (already caught above)

	// Skip patterns with nested logical operators at the base level
	// Example: (x || y) && (x || y).foo
	// The (x || y) expression should not be used as a base for chaining
	// because it contains short-circuiting logic
	// However, we still want to handle comparison operators below
	if ast.IsBinaryExpression(baseExpr) {
		binOp := baseExpr.AsBinaryExpression().OperatorToken.Kind
		if binOp == ast.KindAmpersandAmpersandToken || binOp == ast.KindBarBarToken {
			return Operand{typ: OperandTypeInvalid, node: node}
		}
	}

	// Handle binary expressions (comparisons)
	if ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		op := binExpr.OperatorToken.Kind

		// Determine which side is the expression and which is the value
		var expr, value *ast.Node

		// Check right side first (non-yoda: foo !== null)
		if binExpr.Right.Kind == ast.KindNullKeyword {
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsIdentifier(binExpr.Right) && binExpr.Right.AsIdentifier().Text == "undefined" {
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsVoidExpression(binExpr.Right) {
			// void 0, void(0), void 123, etc. all evaluate to undefined
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsStringLiteral(binExpr.Right) {
			// For typeof checks: typeof foo !== 'undefined'
			expr = binExpr.Left
			value = binExpr.Right
		} else if binExpr.Left.Kind == ast.KindNullKeyword {
			// Yoda style: null !== foo
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsIdentifier(binExpr.Left) && binExpr.Left.AsIdentifier().Text == "undefined" {
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsVoidExpression(binExpr.Left) {
			// Yoda style: void 0 !== foo
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsStringLiteral(binExpr.Left) {
			// Yoda style typeof check: 'undefined' !== typeof foo
			expr = binExpr.Right
			value = binExpr.Left
		}

		if expr != nil && value != nil {
			// Unwrap parentheses from the expression being checked
			expr = unwrapParentheses(expr)

			// Check for typeof expression
			// Note: Only typeof checks with string literals count as undefined checks
			// Regular string comparisons like foo === 'undefined' are NOT typeof checks
			if ast.IsTypeOfExpression(expr) {
				typeofExpr := expr.AsTypeOfExpression()
				if ast.IsStringLiteral(value) && value.AsStringLiteral().Text == "undefined" {
					// AND chain: typeof foo !== 'undefined' && foo.bar
					if (op == ast.KindExclamationEqualsEqualsToken || op == ast.KindExclamationEqualsToken) && isAndChain {
						return Operand{typ: OperandTypeTypeofCheck, node: node, comparedExpr: typeofExpr.Expression}
					}
					// OR chain: typeof foo === 'undefined' || foo.bar
					if (op == ast.KindEqualsEqualsEqualsToken || op == ast.KindEqualsEqualsToken) && !isAndChain {
						return Operand{typ: OperandTypeTypeofCheck, node: node, comparedExpr: typeofExpr.Expression}
					}
				}
			}

			// Regular null/undefined checks
			// Note: Only treat the IDENTIFIER undefined as an undefined check
			// String literal 'undefined' is just a regular string comparison
			// void 0 and other void expressions are treated as undefined
			isNull := value.Kind == ast.KindNullKeyword
			isUndefined := (ast.IsIdentifier(value) && value.AsIdentifier().Text == "undefined") || ast.IsVoidExpression(value)

			// For && chains, we typically want !== checks
			// But we also handle == and === for consistency (even though they're illogical)
			// Pattern: foo == null && foo.bar -> treat same as foo != null || foo.bar
			// For || chains, we want === checks
			if isAndChain {
				// For && chains, check for !== or != with null/undefined
				// These are null/undefined checks that can be converted to optional chaining
				// Example: foo !== null && foo.bar !== null && foo.bar.baz -> foo?.bar?.baz
				// The chain-building code will determine if the last check is a "trailing comparison"
				// that should be preserved in the output
				switch op {
				case ast.KindExclamationEqualsEqualsToken:
					// !== null or !== undefined
					if isNull {
						return Operand{typ: OperandTypeNotStrictEqualNull, node: node, comparedExpr: expr}
					}
					if isUndefined {
						return Operand{typ: OperandTypeNotStrictEqualUndef, node: node, comparedExpr: expr}
					}
				case ast.KindExclamationEqualsToken:
					// != null covers both null and undefined
					if isNull || isUndefined {
						return Operand{typ: OperandTypeNotEqualBoth, node: node, comparedExpr: expr}
					}
				// Handle === and == in AND chains differently:
				// These check if the value IS null/undefined (inverted check)
				// Example: foo == null && foo.bar -> inverted null check
				// These are illogical (would error if foo is null) but can be converted with unsafe option
				case ast.KindEqualsEqualsEqualsToken:
					if isNull {
						// Only treat as inverted check if expr is a simple identifier (base variable)
						// Not for property accesses like foo.bar === null
						if ast.IsIdentifier(expr) || expr.Kind == ast.KindThisKeyword {
							return Operand{typ: OperandTypeStrictEqualNull, node: node, comparedExpr: expr}
						}
					}
					if isUndefined {
						if ast.IsIdentifier(expr) || expr.Kind == ast.KindThisKeyword {
							return Operand{typ: OperandTypeStrictEqualUndef, node: node, comparedExpr: expr}
						}
					}
				case ast.KindEqualsEqualsToken:
					// == null checks for both null and undefined
					if (isNull || isUndefined) && (ast.IsIdentifier(expr) || expr.Kind == ast.KindThisKeyword) {
						return Operand{typ: OperandTypeEqualNull, node: node, comparedExpr: expr}
					}
					// Note: === and == for properties (not base identifiers) fall through to OperandTypeComparison
				}
			} else {
				// OR chain - look for === or == checks (opposite of AND chains)
				// We want === checks (checking that something IS null/undefined)
				// Example: foo === null || foo.bar (base variable check)
				// Example: !foo || foo.bar === null (property comparison at end - treat as Comparison, not null check)
				//
				// Key distinction:
				// - Base identifier: foo === null -> null check operand
				// - Property access: foo.bar === null -> comparison operand (keep the comparison in output)
				//
				// Check if this is a property/element access (comparison at end of chain)
				// vs a base identifier (null check at start of chain)
				isPropertyOrElement := ast.IsPropertyAccessExpression(expr) || ast.IsElementAccessExpression(expr) || ast.IsCallExpression(expr)

				if isPropertyOrElement && (isNull || isUndefined) {
					// Property comparisons against null/undefined should be treated as regular comparisons
					// They will appear at the END of the chain and the comparison will be preserved
					// Example: !foo || foo.bar === null -> foo?.bar === null
					// Example: !foo || foo.bar === undefined -> foo?.bar === undefined
					// Return as a comparison operand so the comparison is kept in the output
					return Operand{typ: OperandTypeComparison, node: node, comparedExpr: expr}
				} else if !isPropertyOrElement {
					// Base identifier null/undefined checks
					// Example: foo === null || foo.bar
					// Note: We use OperandTypeNotStrictEqual* types here because in OR chains,
					// foo === null has the same semantics as foo !== null in AND chains
					// (both filter out null values before accessing properties)
					switch op {
					case ast.KindEqualsEqualsEqualsToken:
						if isNull {
							return Operand{typ: OperandTypeNotStrictEqualNull, node: node, comparedExpr: expr}
						}
						if isUndefined {
							return Operand{typ: OperandTypeNotStrictEqualUndef, node: node, comparedExpr: expr}
						}
					case ast.KindEqualsEqualsToken:
						// == null covers both null and undefined in OR chains
						if isNull || isUndefined {
							return Operand{typ: OperandTypeNotEqualBoth, node: node, comparedExpr: expr}
						}
					}
				}
				// For property null/undefined checks and !== / != operators, treat as comparison
				// Example: !foo || foo.bar !== null (property check - treat as comparison)
				// Example: !foo || foo.bar === undefined (property check - treat as comparison)
			}
		}
	}

	// Handle unary expressions (!foo)
	if ast.IsPrefixUnaryExpression(unwrapped) {
		prefixExpr := unwrapped.AsPrefixUnaryExpression()
		if prefixExpr.Operator == ast.KindExclamationToken {
			// Check if the operand is BARE 'this' - if so, reject
			// Example: !this || ... -> invalid (this is never null)
			// But: !this.bar || !this.bar.baz -> valid (this.bar might be null)
			if prefixExpr.Operand.Kind == ast.KindThisKeyword {
				return Operand{typ: OperandTypeInvalid, node: node}
			}

			// !foo in || chain is like foo !== null in && chain
			if !isAndChain {
				return Operand{typ: OperandTypeNot, node: node, comparedExpr: prefixExpr.Operand}
			}
			// In AND chains, !foo is treated as a negated operand that needs special handling
			// It can only be converted if followed by a comparison (not another negation)
			// Example: !foo && foo.bar === 0 -> foo?.bar === 0 (valid)
			// Example: !foo && !foo.bar -> cannot convert (invalid)
			return Operand{typ: OperandTypeNegatedAndOperand, node: node, comparedExpr: prefixExpr.Operand}
		}
	}

	// If we reach here with a binary expression in an && chain, it's a comparison like foo.bar == 0
	if isAndChain && ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()

		// Determine which side is the property being checked
		// For yoda: '123' == foo.bar.baz -> comparedExpr = foo.bar.baz (right side)
		// For normal: foo.bar.baz == '123' -> comparedExpr = foo.bar.baz (left side)
		// For instanceof: foo.bar.baz instanceof Error -> comparedExpr = foo.bar.baz (left side)
		comparedExpr := unwrapParentheses(binExpr.Left)
		hasPropertyAccess := ast.IsPropertyAccessExpression(comparedExpr) ||
			ast.IsElementAccessExpression(comparedExpr) ||
			ast.IsCallExpression(comparedExpr)

		if ast.IsPropertyAccessExpression(binExpr.Right) || ast.IsElementAccessExpression(binExpr.Right) {
			// Right side is a property access - likely yoda style
			comparedExpr = unwrapParentheses(binExpr.Right)
			hasPropertyAccess = true
		} else if ast.IsCallExpression(binExpr.Right) {
			// Right side is a call - might be yoda style
			comparedExpr = unwrapParentheses(binExpr.Right)
			hasPropertyAccess = true
		}

		// Only treat as a comparison operand if there's a property access
		// Otherwise it's something like "x !== false" which should not trigger the rule
		if !hasPropertyAccess {
			return Operand{typ: OperandTypeInvalid, node: node}
		}

		return Operand{typ: OperandTypeComparison, node: node, comparedExpr: comparedExpr}
	}

	// If we reach here with a binary expression in an || chain, it's a comparison like foo.bar != 0
	if !isAndChain && ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		// Determine which side is the property being checked
		comparedExpr := unwrapParentheses(binExpr.Left)
		if ast.IsPropertyAccessExpression(binExpr.Right) || ast.IsElementAccessExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
		} else if ast.IsCallExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
		}
		return Operand{typ: OperandTypeComparison, node: node, comparedExpr: comparedExpr}
	}

	// Plain expression (foo in && chain, or part of a property access)
	if isAndChain {
		// Reject binary comparisons that are not null/undefined checks
		// Examples: x !== false, x > 0, x < 100, etc.
		if ast.IsBinaryExpression(unwrapped) {
			binExpr := unwrapped.AsBinaryExpression()
			op := binExpr.OperatorToken.Kind

			// Check if this is a comparison operator
			isComparison := op == ast.KindEqualsEqualsToken ||
				op == ast.KindExclamationEqualsToken ||
				op == ast.KindEqualsEqualsEqualsToken ||
				op == ast.KindExclamationEqualsEqualsToken ||
				op == ast.KindLessThanToken ||
				op == ast.KindGreaterThanToken ||
				op == ast.KindLessThanEqualsToken ||
				op == ast.KindGreaterThanEqualsToken

			if isComparison {
				// This is a comparison operator but wasn't recognized as a null/undefined check
				// Don't treat it as a plain truthy check
				return Operand{typ: OperandTypeInvalid, node: node}
			}
		}

		return Operand{typ: OperandTypePlain, node: node, comparedExpr: unwrapped}
	}

	// For OR chains, allow any expression (identifier, property access, etc.)
	// Pattern: foo || foo.bar, foo == null || foo.bar
	if !isAndChain {
		return Operand{typ: OperandTypePlain, node: node, comparedExpr: unwrapped}
	}

	return Operand{typ: OperandTypeInvalid, node: node}
}

// collectOperands collects all operands from a binary expression tree with the given operator kind.
// Returns the operand nodes (preserving parentheses for range calculation) and marks binary expressions as seen.
func (processor *chainProcessor) collectOperands(node *ast.Node, operatorKind ast.Kind) []*ast.Node {
	operandNodes := []*ast.Node{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		// Check the unwrapped node for the operator type
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == operatorKind {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			processor.seenLogicals[unwrapped] = true
		} else {
			// Store the original node (with parentheses) for range calculation
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)
	return operandNodes
}

// collectOperandsWithRanges collects all operands and also returns ranges of binary expressions.
// This is used by OR chains which use range-based seen tracking instead of node-based.
func (processor *chainProcessor) collectOperandsWithRanges(node *ast.Node, operatorKind ast.Kind) ([]*ast.Node, []textRange) {
	operandNodes := []*ast.Node{}
	binaryRanges := []textRange{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		// Check the unwrapped node for the operator type
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == operatorKind {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			// Collect the range for marking as seen
			binRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, unwrapped)
			binaryRanges = append(binaryRanges, textRange{start: binRange.Pos(), end: binRange.End()})
		} else {
			// Store the original node (with parentheses) for range calculation
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)
	return operandNodes, binaryRanges
}

// hasPropertyAccessInChain checks if at least one operand in the chain involves property/element/call access.
// Returns false for patterns like: foo != null && foo !== undefined (just null checks, no access)
func (processor *chainProcessor) hasPropertyAccessInChain(chain []Operand) bool {
	for _, op := range chain {
		if op.comparedExpr != nil {
			unwrapped := unwrapParentheses(op.comparedExpr)
			if ast.IsPropertyAccessExpression(unwrapped) ||
				ast.IsElementAccessExpression(unwrapped) ||
				ast.IsCallExpression(unwrapped) {
				return true
			}
		}
	}
	return false
}

// hasSameBaseIdentifier checks if all operands in the chain have the same base identifier.
// Returns false if different bases are found (e.g., a === undefined || b === null)
func (processor *chainProcessor) hasSameBaseIdentifier(chain []Operand) bool {
	var firstBase *ast.Node
	for _, op := range chain {
		if op.comparedExpr == nil {
			continue
		}
		base := getBaseIdentifier(op.comparedExpr)
		if firstBase == nil {
			firstBase = base
		} else {
			// Compare base identifiers
			firstBaseRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, firstBase)
			baseRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, base)
			if firstBaseRange.Pos() >= 0 && firstBaseRange.End() <= len(processor.sourceText) &&
				baseRange.Pos() >= 0 && baseRange.End() <= len(processor.sourceText) {
				firstBaseText := processor.sourceText[firstBaseRange.Pos():firstBaseRange.End()]
				baseText := processor.sourceText[baseRange.Pos():baseRange.End()]
				if firstBaseText != baseText {
					return false
				}
			}
		}
	}
	return true
}

// shouldSkipForRequireNullish checks if the chain should be skipped based on requireNullish option.
// When requireNullish is true, only convert chains that have explicit nullish checks or nullable types.
func (processor *chainProcessor) shouldSkipForRequireNullish(chain []Operand, isAndChain bool) bool {
	if !processor.opts.RequireNullish {
		return false
	}

	// For OR chains starting with negation, skip entirely
	if !isAndChain && len(chain) > 0 && chain[0].typ == OperandTypeNot {
		return true
	}

	// Check if any operand has an explicit nullish context
	for i, op := range chain {
		// Check for explicit nullish check operators
		if op.typ != OperandTypePlain {
			return false // Has nullish context, don't skip
		}
		// For plain && checks, allow if the type explicitly includes null/undefined
		// (but only for intermediate operands, not the last one)
		if isAndChain && i < len(chain)-1 && op.comparedExpr != nil {
			if processor.includesExplicitNullish(op.comparedExpr) {
				return false // Has nullish type, don't skip
			}
		}
	}
	return true // No nullish context found, skip
}

// processAndChain processes && chains: foo && foo.bar -> foo?.bar
func (processor *chainProcessor) processAndChain(node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return
	}

	// Skip if already seen
	if processor.seenLogicals[node] {
		return
	}

	// Skip if inside JSX - semantic difference
	// In JSX, foo && foo.bar returns false/null/undefined (rendered as-is)
	// while foo?.bar always returns undefined
	if isInsideJSX(node) {
		return
	}

	// Skip if this node is contained within an already-processed && expression
	// OR if this node contains an already-processed && expression (child was processed first)
	// This prevents processing nested && nodes separately
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeStart, nodeEnd := nodeRange.Pos(), nodeRange.End()

	for _, processedRange := range processor.processedAndRanges {
		// Two ranges overlap if: start1 < end2 && start2 < end1
		// Skip any node that overlaps with an already-processed range
		// This prevents processing nested chains or subsequent chains in the same expression
		if nodeStart < processedRange.end && processedRange.start < nodeEnd {
			// This node overlaps with an already-processed range
			processor.seenLogicals[node] = true
			return
		}
	}

	// Mark this range as processed BEFORE doing anything else
	processor.processedAndRanges = append(processor.processedAndRanges, textRange{start: nodeStart, end: nodeEnd})

	// FIRST: Flatten and mark ALL logical expressions in this chain
	// This is critical - we must mark all nested && nodes BEFORE processing
	// to prevent them from being visited separately
	var flattenAndMarkLogicals func(*ast.Node) []*ast.Node
	flattenAndMarkLogicals = func(n *ast.Node) []*ast.Node {
		unwrapped := unwrapParentheses(n)
		if !ast.IsBinaryExpression(unwrapped) {
			return nil
		}
		binExpr := unwrapped.AsBinaryExpression()
		if binExpr.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
			return nil
		}

		// Mark both wrapped and unwrapped versions
		processor.seenLogicals[n] = true
		processor.seenLogicals[unwrapped] = true

		result := []*ast.Node{n, unwrapped}
		// Recursively flatten children
		result = append(result, flattenAndMarkLogicals(binExpr.Left)...)
		result = append(result, flattenAndMarkLogicals(binExpr.Right)...)
		return result
	}

	_ = flattenAndMarkLogicals(node)

	// Collect all && operands using the shared helper
	operandNodes := processor.collectOperands(node, ast.KindAmpersandAmpersandToken)

	if len(operandNodes) < 2 {
		return
	}

	// Parse operands
	operands := make([]Operand, len(operandNodes))
	for i, n := range operandNodes {
		operands[i] = processor.parseOperand(n, true)
	}

	// NOTE: We used to check for conflicting call signatures upfront and mark operands invalid,
	// but this was too aggressive. It prevented partial chains from being detected.
	// For example: foo && foo.bar(a) && foo.bar(a, b).baz
	// Should detect chain [foo, foo.bar(a)] and report it, then naturally break when
	// comparing foo.bar(a) vs foo.bar(a, b).baz in compareNodes.
	// The compareNodes function already handles signature conflicts correctly.

	// Try to find ALL valid chains in the expression
	// Pattern 1: foo && foo.bar && foo.bar.baz
	// Pattern 2: foo !== null && foo.bar
	// Pattern 3: foo !== null && foo !== undefined && foo.bar
	// Pattern 4: foo1 && foo1.bar && foo2 && foo2.bar (multiple independent chains)
	// Pattern 5: foo && foo.bar != null && foo.bar.baz !== undefined (inconsistent checks - break chain)

	var allChains [][]Operand
	var currentChain []Operand
	var lastExpr *ast.Node
	var lastCheckType OperandType // Track the type of the last nullish check
	var chainComplete bool        // Mark when chain should not accept more operands
	var stopProcessing bool       // Stop processing after inconsistent check
	i := 0

	for i < len(operands) && !stopProcessing {
		op := operands[i]

		if op.typ == OperandTypeInvalid {
			// Invalid operand, finalize current chain if valid
			if len(currentChain) >= 2 {
				allChains = append(allChains, currentChain)
			}
			currentChain = nil
			lastExpr = nil
			lastCheckType = OperandTypeInvalid
			chainComplete = false
			i++
			continue
		}

		if len(currentChain) == 0 {
			// Start a new chain
			currentChain = append(currentChain, op)
			lastExpr = op.comparedExpr
			if op.typ != OperandTypePlain {
				lastCheckType = op.typ
			}
			chainComplete = false
			i++
			continue
		}

		// If chain is marked complete, finalize it and start a new one
		if chainComplete {
			if len(currentChain) >= 2 {
				allChains = append(allChains, currentChain)
			}
			currentChain = []Operand{op}
			lastExpr = op.comparedExpr
			lastCheckType = OperandTypeInvalid
			if op.typ != OperandTypePlain {
				lastCheckType = op.typ
			}
			chainComplete = false
			i++
			continue
		}

		// Check if this operand continues the chain
		cmp := processor.compareNodes(lastExpr, op.comparedExpr)

		// IMPORTANT: Check for "STRICT explicit check on call result" pattern FIRST
		// This must happen before any special handling for cmp == NodeInvalid
		// because we need to stop extending even when cmp shows a valid extension.
		//
		// When the previous operand is a STRICT nullish check (!== null or !== undefined)
		// on a call result, we may need to stop extending the chain.
		//
		// RULES:
		// 1. LOOSE checks (!= null): ALWAYS continue - same semantics as optional chaining
		// 2. STRICT checks (!== null or !== undefined):
		//    - If type has BOTH null AND undefined: STOP (incomplete check)
		//    - If type has ONLY what we're checking: CONTINUE (complete check)
		//
		// Example that should STOP:
		//   declare const foo: {bar: () => X | null | undefined};
		//   foo.bar() !== undefined && foo.bar().baz
		//   - Type has BOTH null AND undefined
		//   - !== undefined only checks one, so extending changes behavior
		//
		// Example that should CONTINUE:
		//   declare const foo: {bar: () => X | undefined};  // NO null
		//   foo.bar() !== undefined && foo.bar().baz
		//   - Type has ONLY undefined
		//   - !== undefined is a COMPLETE check, same as optional chaining
		if len(currentChain) > 0 {
			prevOp := currentChain[len(currentChain)-1]
			// Only consider STRICT checks, not loose checks (!= null)
			if isStrictNullishCheck(prevOp.typ) && prevOp.comparedExpr != nil {
				prevUnwrapped := prevOp.comparedExpr
				for ast.IsParenthesizedExpression(prevUnwrapped) {
					prevUnwrapped = prevUnwrapped.AsParenthesizedExpression().Expression
				}
				isCallOrNew := ast.IsCallExpression(prevUnwrapped) || ast.IsNewExpression(prevUnwrapped)
				isElementAccess := ast.IsElementAccessExpression(prevUnwrapped)
				if isCallOrNew || isElementAccess {
					// Check if this is an INCOMPLETE or MISMATCHED strict check
					//
					// Cases to handle:
					// 1. Type has BOTH null AND undefined but user only checks one → incomplete
					// 2. User wrote !== undefined but type has NO undefined → mismatched (preserve check)
					// 3. User wrote !== null but type has NO null → mismatched (preserve check)
					// 4. Type is any/unknown → can't determine, allow conversion
					//
					// IMPORTANT: For any/unknown types, we can't determine exact nullishness,
					// so we should NOT consider these as incomplete checks.
					isAnyOrUnknown := processor.typeIsAnyOrUnknown(prevOp.comparedExpr)
					hasNull := processor.typeIncludesNull(prevOp.comparedExpr)
					hasUndefined := processor.typeIncludesUndefined(prevOp.comparedExpr)

					// Incomplete: type has both but user only checks one
					isIncomplete := !isAnyOrUnknown && hasNull && hasUndefined

					// Mismatched: user checks for something the type doesn't have
					// This indicates user is doing a runtime check that should be preserved
					isMismatched := false
					if !isAnyOrUnknown {
						if prevOp.typ == OperandTypeNotStrictEqualUndef && !hasUndefined && !hasNull {
							// User wrote !== undefined but type has no undefined (and no null)
							// This is a "defensive" check that should be preserved
							isMismatched = true
						}
						if prevOp.typ == OperandTypeNotStrictEqualNull && !hasNull && !hasUndefined {
							// User wrote !== null but type has no null (and no undefined)
							isMismatched = true
						}
					}

					// For call/new expressions: ALWAYS stop at incomplete/mismatched strict check
					// (each call might return different value, semantics always change)
					//
					// For element access with incomplete check: Only stop if unsafe option is NOT enabled
					// For element access with mismatched check: Always stop (user intent should be preserved)
					shouldStop := false
					if isCallOrNew {
						shouldStop = isIncomplete || isMismatched
					} else if isElementAccess {
						if isMismatched {
							shouldStop = true // Always preserve user's defensive checks
						} else if isIncomplete && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
							shouldStop = true
						}
					}

					if shouldStop {
						// Previous operand is an INCOMPLETE/MISMATCHED strict check
						// Stop extending - finalize current chain
						if len(currentChain) >= 2 {
							allChains = append(allChains, currentChain)
						}
						currentChain = nil
						chainComplete = true
						stopProcessing = true
						break
					}
					// If check is COMPLETE (type only has what we check), continue extending
				}
			}
		}

		// Special case for AND chains:
		// Allow extending call expressions even though they may have side effects when:
		// 1. The unsafe option is enabled, OR
		// 2. Both the previous and current operand are plain truthiness checks
		//
		// For case 2: foo && foo<string>() && foo<string>().bar
		// - All operands are plain truthiness checks
		// - The user's intent is clear: chain through the call result
		// - This is a common pattern that typescript-eslint converts
		//
		// This is different from: getFoo() && getFoo().bar (different calls, always unsafe)
		// Track if we used special handling to allow call chain extension
		usedCallChainExtension := false
		_ = usedCallChainExtension // May be set but not used after simplification
		if cmp == NodeInvalid {
			// Check if we should allow extending through call expression
			// Either via unsafe option OR via plain truthiness chain pattern
			prevOp := currentChain[len(currentChain)-1]
			isPlainTruthinessChain := prevOp.typ == OperandTypePlain && op.typ == OperandTypePlain

			if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing || isPlainTruthinessChain {
				// Check if lastExpr is a call/new expression and op.comparedExpr extends it
				lastUnwrapped := lastExpr
				if lastUnwrapped != nil {
					for ast.IsParenthesizedExpression(lastUnwrapped) {
						lastUnwrapped = lastUnwrapped.AsParenthesizedExpression().Expression
					}
					if ast.IsCallExpression(lastUnwrapped) || ast.IsNewExpression(lastUnwrapped) {
						// IMPORTANT: Only allow text-based extension if the FIRST operand
						// in the chain is NOT rooted in a new/call expression.
						//
						// Valid: foo && foo() && foo().bar
						//   - First operand `foo` is an identifier
						//   - All refer to the same base object
						//
						// Invalid: new Map().get('a') && new Map().get('a').what
						//   - First operand `new Map().get('a')` is rooted in `new Map()`
						//   - Each `new Map()` creates a fresh instance, they're not the same
						//
						// Check if the first operand's base is a new/call expression
						firstOpExpr := currentChain[0].comparedExpr
						baseExpr := firstOpExpr
						for baseExpr != nil {
							unwrapped := baseExpr
							for ast.IsParenthesizedExpression(unwrapped) {
								unwrapped = unwrapped.AsParenthesizedExpression().Expression
							}
							if ast.IsPropertyAccessExpression(unwrapped) {
								baseExpr = unwrapped.AsPropertyAccessExpression().Expression
							} else if ast.IsElementAccessExpression(unwrapped) {
								baseExpr = unwrapped.AsElementAccessExpression().Expression
							} else if ast.IsCallExpression(unwrapped) {
								baseExpr = unwrapped.AsCallExpression().Expression
							} else {
								// Found the base - check if it's a new expression
								break
							}
						}

						// If the base of the first operand is a new expression, don't allow extension
						// (each `new X()` creates a fresh instance)
						firstOpRootedInNew := baseExpr != nil && ast.IsNewExpression(baseExpr)

						if !firstOpRootedInNew {
							// Try text-based comparison to see if op extends lastExpr
							lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastExpr)
							opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.comparedExpr)
							sourceText := processor.sourceText
							if lastRange.Pos() >= 0 && lastRange.End() <= len(sourceText) &&
								opRange.Pos() >= 0 && opRange.End() <= len(sourceText) {
								lastText := sourceText[lastRange.Pos():lastRange.End()]
								opText := sourceText[opRange.Pos():opRange.End()]
								if strings.HasPrefix(opText, lastText) {
									remainder := strings.TrimPrefix(opText, lastText)
									if len(remainder) > 0 && (remainder[0] == '.' || remainder[0] == '[' || remainder[0] == '(') {
										// op extends lastExpr, treat as NodeSubset
										cmp = NodeSubset
										usedCallChainExtension = true
									}
								}
							}
						}
					}
				}
			}
		}

		if cmp == NodeEqual {
			// Same expression, might be additional nullish check
			// foo !== null && foo !== undefined
			// OR might be a duplicate plain check: foo.bar.baz && foo.bar.baz
			// OR a check followed by access: foo.bar !== null && foo.bar
			if isExplicitNullishCheck(op.typ) {

				// Check for inconsistent check types
				// If we had a "both" check (!= null) and now have a specific check (!== undefined or !== null),
				// This is redundant but not incorrect - include it and continue
				// We DON'T mark the chain as complete because subsequent property accesses should be included
				// Example: foo != null && foo !== undefined && foo.bar -> foo?.bar (all checks on foo)
				if lastCheckType == OperandTypeNotEqualBoth && isStrictNullishCheck(op.typ) {
					// Include this redundant check
					currentChain = append(currentChain, op)
					// Update lastCheckType to the more specific check
					lastCheckType = op.typ
					i++
					continue
				}

				// Check for incomplete paired nullish check pattern
				// Example: foo.bar.baz !== undefined && foo.bar.baz !== null
				// When we have:
				// - Previous operand checking ONLY undefined (NotStrictEqualUndef or TypeofCheck)
				// - Current operand checking ONLY null (NotStrictEqualNull)
				// - Both on the SAME expression
				// - No subsequent property access follows
				// Then we should STOP the chain at the previous operand (as a trailing comparison)
				// and leave the current operand unchanged
				isComplementaryCheck := false
				if (lastCheckType == OperandTypeNotStrictEqualUndef || lastCheckType == OperandTypeTypeofCheck) &&
					op.typ == OperandTypeNotStrictEqualNull {
					isComplementaryCheck = true
				} else if lastCheckType == OperandTypeNotStrictEqualNull &&
					(op.typ == OperandTypeNotStrictEqualUndef || op.typ == OperandTypeTypeofCheck) {
					isComplementaryCheck = true
				}

				if isComplementaryCheck {
					// This is a complementary check (null + undefined on same expression)
					// Together they are equivalent to `!= null`
					// Include this operand in the chain
					currentChain = append(currentChain, op)
					if op.typ != OperandTypePlain {
						lastCheckType = op.typ
					}

					// Check if there's a subsequent property access
					hasSubsequentPropertyAccess := false
					if i+1 < len(operandNodes) {
						nextOp := operands[i+1]
						if nextOp.comparedExpr != nil {
							nextCmp := processor.compareNodes(op.comparedExpr, nextOp.comparedExpr)
							if nextCmp == NodeSubset {
								// There IS a subsequent property access - continue building chain
								hasSubsequentPropertyAccess = true
							}
						}
					}

					if !hasSubsequentPropertyAccess {
						// No subsequent property access - this pair is the end of the chain
						// Mark the chain as complete (the complementary pair will be simplified to != null during fix)
						chainComplete = true
					}
					i++
					continue
				}

				currentChain = append(currentChain, op)
				if op.typ != OperandTypePlain {
					lastCheckType = op.typ
				}
				i++
				continue
			} else if op.typ == OperandTypePlain {
				// Plain operand matching the last checked expression
				// Check if the previous operand was a null check on the same expression
				// If so, this is the pattern: foo.bar !== null && foo.bar
				// We should include this in the chain as it's the actual access after the check
				prevOp := currentChain[len(currentChain)-1]
				if isExplicitNullishCheck(prevOp.typ) {
					// Previous was a null check on the same expression, include this access
					currentChain = append(currentChain, op)
					// Don't update lastExpr since we're accessing the same thing
					i++
					continue
				}
				// Otherwise, it's a true duplicate - skip it
				// Example: foo && foo.bar && foo.bar && foo.bar.baz
				// The duplicate foo.bar is redundant
				i++
				continue
			} else if op.typ == OperandTypeComparison && len(currentChain) > 0 {
				// Comparison operand on the same expression as the previous check
				// Example: typeof foo.bar.baz !== 'undefined' && foo.bar.baz <= 100
				// The comparison is a trailing comparison that should be included
				prevOp := currentChain[len(currentChain)-1]
				if isExplicitNullishCheck(prevOp.typ) {
					// Previous was a null/typeof check on the same expression
					// Include this trailing comparison and mark chain as complete
					currentChain = append(currentChain, op)
					chainComplete = true
					i++
					continue
				}
			}
		} else if cmp == NodeSubset {
			// Property access of previous expression

			// Note: Previously we had conservative logic here that would stop the chain
			// when there were calls. But this was too restrictive - typescript-eslint
			// converts chains with calls as long as each step is checked.
			// Since we're in `cmp == NodeSubset` (the current operand extends lastExpr),
			// and we've already verified the previous operands, it's safe to continue.
			// The semantic differences (if any) are handled by useSuggestion logic later.

			// Special check: if previous operand is NegatedAndOperand or inverted null check, handle carefully
			// !a checks ALL falsy values (0, "", false, null, undefined)
			// foo == null checks if value IS null/undefined (inverted)
			// while optional chaining only checks null/undefined
			// Examples:
			// - !a && !a.b -> cannot convert (both negated)
			// - !a && a.b -> cannot convert (negation + plain property)
			// - !a && a.b === 0 -> cannot convert (negation + comparison) - semantics differ!
			// - foo == null && foo.bar == 0 -> cannot convert (inverted + comparison) - semantics differ!
			// The inverted check means the code path only executes when foo IS null/undefined,
			// but optional chaining would skip the property access when foo IS null/undefined.
			if len(currentChain) > 0 {
				lastOp := currentChain[len(currentChain)-1]
				isInvertedCheck := lastOp.typ == OperandTypeNegatedAndOperand ||
					lastOp.typ == OperandTypeEqualNull ||
					lastOp.typ == OperandTypeStrictEqualNull ||
					lastOp.typ == OperandTypeStrictEqualUndef
				if isInvertedCheck {
					// Cannot convert ANY patterns with inverted checks
					// Break chain and start new one (which will also be invalid)
					currentChain = nil
					lastExpr = nil
					lastCheckType = OperandTypeInvalid
					chainComplete = false
					i++
					continue
				}
			}

			// Check if we should stop the chain due to:
			// 1. INCONSISTENT nullish check (truthiness + strict mixing)
			// 2. Strict check on a CALL expression result
			// 3. Multiple strict checks followed by a Plain operand (should preserve trailing access)
			//
			// Case 1: Inconsistent checks
			// When we have a strict check (!== null or !== undefined) on an expression
			// but a PARENT expression was checked with a TRUTHINESS check (plain `foo`)
			// Example: foo && foo.bar !== null && foo.bar.baz !== undefined && foo.bar.baz.buzz
			// -> foo?.bar?.baz !== undefined && foo.bar.baz.buzz
			//
			// Case 2: Strict check on call result
			// When there's a strict check on a CALL expression result, we should NOT chain through
			// because converting to ?. would change semantics (checks both null and undefined)
			// Example: foo.bar !== undefined && foo.bar?.() !== undefined && foo.bar?.().baz
			// -> foo.bar?.() !== undefined && foo.bar?.().baz
			// The `!== undefined` on the call result should be preserved
			//
			// Case 3: Strict checks followed by Plain
			// When we have multiple strict checks (typeof/!== null/!== undefined) and then
			// a Plain operand, the Plain operand should be preserved as trailing, not converted
			// Example: typeof foo !== 'undefined' && typeof foo.bar !== 'undefined' && foo.bar.baz
			// -> typeof foo?.bar !== 'undefined' && foo.bar.baz
			// NOT: foo?.bar?.baz (which would lose the typeof check)
			//
			// Note: Mixing strict-null (!== null) with strict-undefined (!== undefined) on
			// PROPERTY accesses is FINE because both are checking for nullish values.
			// Example: foo !== null && foo.bar !== undefined && foo.bar.baz()
			// -> foo?.bar?.baz() (full conversion - ends with call, not plain access)
			if len(currentChain) > 0 {
				lastChainOp := currentChain[len(currentChain)-1]
				shouldStopChain := false

				// Check if last operand is a strict check (undefined only or null only)
				isStrictUndef := lastChainOp.typ == OperandTypeNotStrictEqualUndef || lastChainOp.typ == OperandTypeTypeofCheck
				isStrictNull := lastChainOp.typ == OperandTypeNotStrictEqualNull

				if isStrictUndef || isStrictNull {
					// Case 2: Check if the strict check is on a CALL expression result
					// If so, we should NOT chain through it UNLESS:
					// 1. The call already has optional chaining in source (e.g., foo?.())
					// 2. The call's base expression was checked earlier in our chain (e.g., foo.bar !== null && foo.bar() !== null)
					//    In case 2, the base will be converted to optional chain, so foo.bar() becomes foo?.bar?.()
					// Example (should NOT chain): foo() !== null && foo().bar -> can't convert (calling foo() twice)
					// Example (SHOULD chain): foo?.() !== null && foo?.().bar -> foo?.()?.bar (single call)
					// Example (SHOULD chain): foo.bar !== null && foo.bar() !== null && foo.bar().baz -> foo?.bar?.()?.baz
					if lastChainOp.comparedExpr != nil {
						unwrappedLast := lastChainOp.comparedExpr
						for ast.IsParenthesizedExpression(unwrappedLast) {
							unwrappedLast = unwrappedLast.AsParenthesizedExpression().Expression
						}
						// Check if it's a call expression (including optional call)
						if ast.IsCallExpression(unwrappedLast) {
							// First check if the source already has optional chaining
							hasOptionalChaining := hasOptionalChaining(lastChainOp.comparedExpr)

							// Also check if the call's base expression was checked earlier in our chain
							// If so, the base will be converted to optional chain
							callBaseWasChecked := false
							if !hasOptionalChaining {
								callExpr := unwrappedLast.AsCallExpression()
								if callExpr != nil && callExpr.Expression != nil {
									// Get the callee (e.g., foo.bar in foo.bar())
									calleeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, callExpr.Expression)
									sourceText := processor.sourceText
									calleeText := ""
									if calleeRange.Pos() >= 0 && calleeRange.End() <= len(sourceText) {
										calleeText = sourceText[calleeRange.Pos():calleeRange.End()]
									}

									// Check if any earlier operand in the chain checked this callee
									for _, prevOp := range currentChain[:len(currentChain)-1] {
										if prevOp.comparedExpr != nil {
											prevRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, prevOp.comparedExpr)
											prevText := ""
											if prevRange.Pos() >= 0 && prevRange.End() <= len(sourceText) {
												prevText = sourceText[prevRange.Pos():prevRange.End()]
											}
											if prevText == calleeText {
												// The callee was checked earlier, so it will be converted to optional chain
												callBaseWasChecked = true
												break
											}
										}
									}
								}
							}

							if !hasOptionalChaining && !callBaseWasChecked {
								// Strict check on call result without optional chaining - don't chain through
								shouldStopChain = true
							}
							// If it already has optional chaining (e.g., foo?.()), or base was checked, we can chain through
						}
					}

					// Case 4: Check if the strict check is INCOMPLETE
					// A strict check is incomplete when the type includes BOTH null AND undefined
					// but the check only covers one of them.
					// This case only applies when we have MULTIPLE strict checks in the chain
					// (not just one strict check followed by a plain property access).
					// Example that should trigger Case 4:
					//   foo !== undefined && foo.bar !== undefined && foo.bar.baz
					//   - foo type has both null and undefined
					//   - Second strict check should become trailing comparison
					//   -> foo?.bar !== undefined && foo.bar.baz
					// Example that should NOT trigger Case 4:
					//   foo !== null && foo.bar
					//   - Just one strict check followed by plain access
					//   -> foo?.bar (full conversion, even if unsafe)
					//
					// IMPORTANT: When allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing
					// is enabled, we should NOT stop the chain - the user has opted into potentially
					// unsafe conversions that change return type semantics.
					// Count strict checks in the chain
					strictCheckCount := 0
					for _, chainOp := range currentChain {
						if isStrictNullishCheck(chainOp.typ) {
							strictCheckCount++
						}
					}
					// Only apply Case 4 if:
					// - NOT using unsafe option (if unsafe, allow full conversion)
					// - We have at least one strict check already AND
					// - Current operand is also a strict check (resulting in 2+ strict checks)
					if !shouldStopChain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing &&
						strictCheckCount >= 1 && lastChainOp.comparedExpr != nil &&
						isStrictNullishCheck(op.typ) {
						isAnyOrUnknown := processor.typeIsAnyOrUnknown(lastChainOp.comparedExpr)
						hasNull := processor.typeIncludesNull(lastChainOp.comparedExpr)
						hasUndefined := processor.typeIncludesUndefined(lastChainOp.comparedExpr)
						// Check is incomplete if type has both null/undefined but we only check one
						// For any/unknown types, we can't determine exact nullishness, so don't stop
						if !isAnyOrUnknown && hasNull && hasUndefined {
							// Type has both, so strict check is incomplete
							// When the unsafe option is enabled, we still need to be careful:
							// - We CAN convert the chain up to and including the strict check
							// - But we should preserve the rest as-is (not convert further)
							// This handles cases like: foo !== undefined && foo.bar !== undefined && foo.bar.baz
							//   -> foo?.bar !== undefined && foo.bar.baz
							// The first part is converted, but the second !== undefined is preserved as trailing
							shouldStopChain = true
						}
					}

					// Case 3: Check if we just added a typeof check and the current operand is NOT a typeof check
					// typeof checks act as chain boundaries - they can absorb previous checks,
					// but subsequent non-typeof checks should start a new chain
					// Example: foo != null && typeof foo.bar !== 'undefined' && foo.bar != null && foo.bar.baz
					// Should produce TWO chains:
					// 1. foo != null && typeof foo.bar !== 'undefined' → typeof foo?.bar !== 'undefined'
					// 2. foo.bar != null && foo.bar.baz → foo.bar?.baz
					//
					// IMPORTANT: When allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing
					// is enabled, we should NOT stop at typeof boundaries - allow full conversion.
					if !shouldStopChain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing &&
						len(currentChain) >= 2 && op.typ != OperandTypeTypeofCheck {
						lastChainOp := currentChain[len(currentChain)-1]
						if lastChainOp.typ == OperandTypeTypeofCheck {
							// Last operand is a typeof check and current is not typeof
							// Stop the chain here - the typeof check is the boundary
							shouldStopChain = true
						}
					}

					// Case 5: Strict check followed by plain access that extends it
					// When the last operand in the chain is a strict nullish check (!== null, !== undefined, typeof)
					// and the current operand is a plain access that EXTENDS the checked expression,
					// we should stop the chain. This is the "inconsistent checks" case.
					//
					// The issue is: optional chaining (?.) checks for BOTH null and undefined,
					// but a strict check only checks for ONE. If we convert the plain access to
					// optional chaining, we'd be adding a check that wasn't in the original code.
					//
					// Example: foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz
					// - foo.bar.baz !== undefined only checks for undefined, not null
					// - foo.bar.baz.buzz is a plain access extending the strict-checked expression
					// - Converting to foo?.bar?.baz?.buzz would add a null check on foo.bar.baz.buzz
					// -> foo?.bar?.baz !== undefined && foo.bar.baz.buzz (preserve the plain access)
					//
					// EXCEPTION: If the strict check is part of a complementary pair (both null and undefined
					// are checked on the same expression), then we CAN continue the chain because the combined
					// checks are equivalent to optional chaining.
					//
					// IMPORTANT: This only applies when:
					// - NOT using unsafe option
					// - Current operand is a plain access (not another nullish check)
					// - Current operand EXTENDS the strict-checked expression (not just any access)
					// - The strict check is NOT part of a complementary pair
					if !shouldStopChain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing &&
						op.typ == OperandTypePlain && isStrictNullishCheck(lastChainOp.typ) {
						// Check if current operand extends the last chain operand's expression
						if lastChainOp.comparedExpr != nil && op.comparedExpr != nil {
							cmpResult := processor.compareNodes(lastChainOp.comparedExpr, op.comparedExpr)
							if cmpResult == NodeSubset {
								// Before stopping, check if the strict check is part of a complementary pair
								// Look for another check on the same expression that completes the null+undefined check
								hasComplementaryCheck := false
								isLastNull := lastChainOp.typ == OperandTypeNotStrictEqualNull
								isLastUndef := lastChainOp.typ == OperandTypeNotStrictEqualUndef || lastChainOp.typ == OperandTypeTypeofCheck

								for j := range len(currentChain) - 1 {
									prevOp := currentChain[j]
									if prevOp.comparedExpr != nil {
										prevCmp := processor.compareNodes(prevOp.comparedExpr, lastChainOp.comparedExpr)
										if prevCmp == NodeEqual {
											// Same expression - check if it's a complementary check
											isPrevNull := prevOp.typ == OperandTypeNotStrictEqualNull
											isPrevUndef := prevOp.typ == OperandTypeNotStrictEqualUndef || prevOp.typ == OperandTypeTypeofCheck
											if (isLastNull && isPrevUndef) || (isLastUndef && isPrevNull) {
												hasComplementaryCheck = true
												break
											}
										}
									}
								}

								if !hasComplementaryCheck {
									// Before stopping, check if the strict check is COMPLETE for the type.
									// A strict check is complete when:
									// - `!== null` and type has null but NOT undefined
									// - `!== undefined` and type has undefined but NOT null
									// In these cases, the check is semantically equivalent to optional chaining
									// for that type, so we should continue the chain (with suggestion fixer).
									strictCheckComplete := false

									// IMPORTANT: Due to type narrowing, we need to find the FIRST occurrence
									// of this expression in the chain to get the un-narrowed type.
									// Otherwise, duplicate checks like foo.bar.baz !== null && foo.bar.baz !== null
									// would have the second occurrence already narrowed to non-null.
									exprToCheck := lastChainOp.comparedExpr
									for j := range len(currentChain) - 1 {
										prevOp := currentChain[j]
										if prevOp.comparedExpr != nil {
											prevCmp := processor.compareNodes(prevOp.comparedExpr, lastChainOp.comparedExpr)
											if prevCmp == NodeEqual {
												// Found an earlier occurrence - use its node for type checking
												exprToCheck = prevOp.comparedExpr
												break
											}
										}
									}

									if exprToCheck != nil {
										isAnyOrUnknown := processor.typeIsAnyOrUnknown(exprToCheck)
										hasNull := processor.typeIncludesNull(exprToCheck)
										hasUndefined := processor.typeIncludesUndefined(exprToCheck)

										// For any/unknown types, we can't determine, so treat as incomplete
										if !isAnyOrUnknown {
											// !== null is complete if type has null but NOT undefined
											if isLastNull && hasNull && !hasUndefined {
												strictCheckComplete = true
											}
											// !== undefined is complete if type has undefined but NOT null
											if isLastUndef && hasUndefined && !hasNull {
												strictCheckComplete = true
											}
										}
									}

									if !strictCheckComplete {
										// Current operand extends the strict-checked expression
										// Stop the chain - the strict check becomes trailing, plain access is preserved
										shouldStopChain = true
									}
								}
							}
						}
					}
				}

				// If we should stop the chain and current is a property access, do so
				if shouldStopChain && (op.typ == OperandTypePlain || op.typ == OperandTypeComparison) {
					// The last operand becomes a trailing comparison
					// Don't add current operand, mark chain as complete
					// Finalize the current chain before breaking
					if len(currentChain) >= 2 {
						allChains = append(allChains, currentChain)
					}
					chainComplete = true
					stopProcessing = true
					// Break out of the loop entirely to avoid finding additional chains
					break
				}
			}

			currentChain = append(currentChain, op)
			lastExpr = op.comparedExpr
			if op.typ != OperandTypePlain {
				lastCheckType = op.typ
			}

			// Check if we just added a strict check operand AND any PREVIOUS operand
			// in the chain has an incomplete strict check (type has both null and undefined).
			// If so, mark the chain as complete to prevent further operands from being added.
			// This handles cases like: foo !== undefined && foo.bar !== undefined && foo.bar.baz
			//   -> foo?.bar !== undefined && foo.bar.baz
			// The chain stops at foo.bar !== undefined (which becomes trailing comparison)
			//
			// IMPORTANT: When allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing
			// is enabled, we should NOT stop at incomplete strict checks - allow full conversion.
			if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing &&
				isStrictNullishCheck(op.typ) {
				// Check if any PREVIOUS operand (not the one we just added) has an incomplete strict check
				for j := range len(currentChain) - 1 {
					prevOp := currentChain[j]
					if isStrictNullishCheck(prevOp.typ) && prevOp.comparedExpr != nil {
						isAnyOrUnknown := processor.typeIsAnyOrUnknown(prevOp.comparedExpr)
						hasNull := processor.typeIncludesNull(prevOp.comparedExpr)
						hasUndefined := processor.typeIncludesUndefined(prevOp.comparedExpr)
						// For any/unknown types, we can't determine exact nullishness, so don't stop
						if !isAnyOrUnknown && hasNull && hasUndefined {
							// Previous operand has incomplete strict check
							// Mark chain as complete but don't stop processing (allow the chain to be finalized)
							chainComplete = true
							break
						}
					}
				}
			}

			i++
			continue
		} else if cmp == NodeInvalid {
			// Nodes are incomparable (e.g., different call signatures)
			// Break chain and start new one
			if len(currentChain) >= 2 {
				allChains = append(allChains, currentChain)
			}
			currentChain = []Operand{op}
			lastExpr = op.comparedExpr
			lastCheckType = OperandTypeInvalid
			if op.typ != OperandTypePlain {
				lastCheckType = op.typ
			}
			chainComplete = false
			i++
			continue
		}

		// Chain broken - finalize current chain if valid
		if len(currentChain) >= 2 {
			allChains = append(allChains, currentChain)
		}
		// Start new chain with current operand
		currentChain = []Operand{op}
		lastExpr = op.comparedExpr
		lastCheckType = OperandTypeInvalid
		if op.typ != OperandTypePlain {
			lastCheckType = op.typ
		}
		chainComplete = false
		i++
	}

	// Finalize last chain if valid
	// If stopProcessing is true, only finalize if the chain was marked as complete
	// (meaning it was intentionally stopped, not just broken)
	shouldFinalize := len(currentChain) >= 2
	if stopProcessing && !chainComplete {
		// Don't finalize incomplete chains when we stopped processing
		shouldFinalize = false
	}
	if shouldFinalize {
		allChains = append(allChains, currentChain)
	}

	// Need at least one valid chain
	if len(allChains) == 0 {
		return
	}

	// If we have multiple chains, check if they have different base identifiers
	// Previously we used to skip these entirely, but now we should report all chains
	// Example: foo && foo.a && bar && bar.a should report TWO errors
	// This matches typescript-eslint behavior

	// If we stopped processing due to inconsistent checks, only report the first chain
	// to avoid reporting additional chains that come after the inconsistent check
	chainsToReport := allChains
	if stopProcessing {
		// When stopProcessing is true, we should only have 1 chain (the one that triggered stop)
		// If we have more, something went wrong in the chain building logic
		if len(allChains) > 1 {
			// Only use first chain if we got multiple
			chainsToReport = allChains[:1]
		}
	}

	// Note: Previously we skipped all chains when multiple bases were present
	// without the unsafe option. However, typescript-eslint reports each chain
	// separately, so we now process all chains regardless of base differences.

	// Filter out chains that overlap with longer chains
	// For example: foo !== null && foo.bar !== null && foo.bar.baz !== null && foo.bar.baz.qux
	// This may create multiple overlapping chains, and we only want to report the longest one
	if len(chainsToReport) > 1 {
		// Build a list of chain ranges
		type chainWithRange struct {
			chain    []Operand
			startPos int
			endPos   int
			length   int
		}

		chainRanges := make([]chainWithRange, len(chainsToReport))
		for i, chain := range chainsToReport {
			if len(chain) == 0 {
				continue
			}
			// Get the text range of the entire chain
			firstOpRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[0].node)
			lastOpRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-1].node)
			chainRanges[i] = chainWithRange{
				chain:    chain,
				startPos: firstOpRange.Pos(),
				endPos:   lastOpRange.End(),
				length:   len(chain),
			}
		}

		// Filter: keep only chains that don't overlap with a longer chain
		filteredChains := [][]Operand{}
		for i, cr1 := range chainRanges {
			if len(cr1.chain) == 0 {
				continue
			}
			isOverlappedByLonger := false
			for j, cr2 := range chainRanges {
				if i == j || len(cr2.chain) == 0 {
					continue
				}
				// Check if chains overlap and cr2 is longer
				overlaps := cr1.startPos < cr2.endPos && cr2.startPos < cr1.endPos
				if overlaps && cr2.length > cr1.length {
					isOverlappedByLonger = true
					break
				}
			}
			if !isOverlappedByLonger {
				filteredChains = append(filteredChains, cr1.chain)
			}
		}
		chainsToReport = filteredChains
	}

	// Process each chain
	for _, chain := range chainsToReport {
		// Check if any operand in this chain overlaps with ANY previously reported range
		hasOverlap := false
		for _, op := range chain {
			opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
			opStart, opEnd := opRange.Pos(), opRange.End()

			// Check if this operand's range overlaps with any reported range
			for reportedRange := range processor.reportedRanges {
				// Two ranges overlap if: opStart < reportedEnd && reportedStart < opEnd
				if opStart < reportedRange.end && reportedRange.start < opEnd {
					hasOverlap = true
					break
				}
			}
			if hasOverlap {
				break
			}
		}
		if hasOverlap {
			continue // Skip this chain as it overlaps with a previously reported one
		}

		// Skip if the first operand is Plain but contains optional chaining
		// AND represents a different base than subsequent operands
		// Example to SKIP: (foo?.a)() && foo.a().b
		//   - First: calling (foo?.a)()
		//   - Second: accessing foo.a().b
		//   - These are fundamentally different expressions, can't be merged
		// Example to SKIP: (foo?.a).b && foo.a.b.c
		//   - First: accessing (foo?.a).b
		//   - Second: accessing foo.a.b.c
		//   - These use different bases (foo?.a vs foo.a), can't be merged
		// Example to ALLOW: x?.a != null && x.a.b
		//   - First: checking x?.a (explicit null check)
		//   - Second: accessing x.a.b
		//   - Same base (x), can optimize to x?.a?.b (the explicit check makes this valid)
		if len(chain) >= 2 {
			firstOp := chain[0]
			if firstOp.typ == OperandTypePlain && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
				// Plain operand with optional chaining - this is risky
				// The subsequent operands likely use different bases
				// Skip this pattern for safety
				continue
			}
		}

		// Skip chains where the first operand ALREADY has optional chaining AND a STRICT check
		// These patterns result from a previous partial fix that intentionally stopped
		// Example: foo?.bar?.baz !== undefined && foo.bar.baz.buzz
		// - The !== undefined is a STRICT check (only catches undefined, not null)
		// - The ?. handles both null and undefined
		// - Converting further would change semantics
		// - Don't continue optimizing - the partial fix was intentional
		//
		// EXCEPTION: Don't skip if this is a "split strict equals" pattern where
		// the last two operands form a complementary pair (null + undefined check on same expression)
		// Example: foo?.bar?.baz !== null && typeof foo.bar.baz !== 'undefined'
		// This should be optimized to: foo?.bar?.baz != null
		if len(chain) >= 2 {
			firstOp := chain[0]
			if isStrictNullishCheck(firstOp.typ) && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
				// Check if this is a split strict equals pattern
				isSplitStrictEquals := false
				if len(chain) == 2 {
					lastOp := chain[1]
					// Check if they're on the same expression (when normalized)
					if lastOp.comparedExpr != nil {
						cmpResult := processor.compareNodes(firstOp.comparedExpr, lastOp.comparedExpr)
						if cmpResult == NodeEqual {
							// Check if they form a complementary pair
							isFirstUndef := firstOp.typ == OperandTypeNotStrictEqualUndef || firstOp.typ == OperandTypeTypeofCheck
							isFirstNull := firstOp.typ == OperandTypeNotStrictEqualNull
							isLastUndef := lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeTypeofCheck
							isLastNull := lastOp.typ == OperandTypeNotStrictEqualNull
							if (isFirstUndef && isLastNull) || (isFirstNull && isLastUndef) {
								isSplitStrictEquals = true
							}
						}
					}
				}
				if !isSplitStrictEquals {
					continue
				}
			}
		}

		// NOTE: We intentionally do NOT skip chains with "inconsistent" optional chaining
		// Example: foo?.bar.baz != null && foo.bar?.baz.bam != null
		// Even though the optional tokens are in different positions, this is valid
		// because the null check semantics mean if we reach the second operand,
		// the first check passed, so we know the base chain is not null
		// The fix will merge them correctly: foo?.bar.baz?.bam != null

		// Also skip single-operand chains (need at least 2 operands to form a chain)
		if len(chain) < 2 {
			continue
		}

		// Ensure at least one operand involves property/element/call access
		// Pattern to skip: foo != null && foo !== undefined (just null checks, no access)
		// Pattern to allow: foo != null && foo.bar (has property access)
		if !processor.hasPropertyAccessInChain(chain) {
			continue // No property access, nothing to chain
		}

		// Skip chains where all operands check the SAME expression
		// Pattern to skip: x['y'] !== undefined && x['y'] !== null
		// This is a complete nullish check on a SINGLE property, not a chain
		// A valid chain requires operands that EXTEND each other (e.g., foo && foo.bar)
		//
		// EXCEPTION: Don't skip if this is a "split strict equals" pattern
		// Example: foo !== null && typeof foo !== 'undefined' -> should become foo != null
		// Example: foo?.bar?.baz !== null && typeof foo.bar.baz !== 'undefined' -> foo?.bar?.baz != null
		if len(chain) >= 2 {
			allSameExpr := true
			firstParts := processor.flattenForFix(chain[0].comparedExpr)
			for i := 1; i < len(chain); i++ {
				opParts := processor.flattenForFix(chain[i].comparedExpr)
				if len(opParts) != len(firstParts) {
					allSameExpr = false
					break
				}
				for j := range firstParts {
					if firstParts[j].text != opParts[j].text {
						allSameExpr = false
						break
					}
				}
				if !allSameExpr {
					break
				}
			}
			if allSameExpr {
				// Check if this is a split strict equals pattern (complementary null + undefined checks)
				// If so, we should NOT skip - we want to simplify to != null
				isSplitStrictEquals := false
				if len(chain) == 2 {
					firstOp := chain[0]
					lastOp := chain[1]
					isFirstUndef := firstOp.typ == OperandTypeNotStrictEqualUndef || firstOp.typ == OperandTypeTypeofCheck
					isFirstNull := firstOp.typ == OperandTypeNotStrictEqualNull
					isLastUndef := lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeTypeofCheck
					isLastNull := lastOp.typ == OperandTypeNotStrictEqualNull
					if (isFirstUndef && isLastNull) || (isFirstNull && isLastUndef) {
						isSplitStrictEquals = true
					}
				}
				if !isSplitStrictEquals {
					continue // All operands check the same expression, nothing to chain
				}
			}
		}

		// Note: Comparison operands in AND chains are now allowed without the unsafe option
		// for patterns like `foo && foo.bar == 0` which can be safely converted to `foo?.bar == 0`
		// The semantics are preserved:
		// - Original: if foo is nullish, returns foo (falsy); otherwise returns (foo.bar == 0)
		// - Converted: foo?.bar == 0 -> if foo is nullish, returns (undefined == 0) = false
		// Both are falsy when foo is nullish, so the conversion is safe.
		//
		// However, inverted checks (!foo && ..., foo == null && ...) are blocked earlier
		// in the chain building logic because those have different semantics.

		// Skip chains where ALL operands use STRICT nullish checks AND subsequent operands
		// already have optimal optional chaining.
		//
		// Key insight:
		// - Strict checks (!== null, !== undefined) only guard against ONE nullish value
		// - Optional chaining (?.) guards against BOTH null and undefined
		// - So foo !== undefined && foo?.bar is intentional: strict check + optional for the other
		// - But foo != undefined && foo?.bar is redundant: loose check already covers both
		//
		// Only skip if ALL checks in the chain are STRICT (not loose) AND subsequent operands
		// have optional chaining. For loose checks, the optional chaining is redundant and
		// should be optimized away.
		if len(chain) >= 2 {
			// Check if ALL operands after the first have optional chaining
			allSubsequentHaveOptionalChaining := true
			for i := 1; i < len(chain); i++ {
				if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
					allSubsequentHaveOptionalChaining = false
					break
				}
			}

			// Check if ALL nullish checks in the chain are STRICT (not loose)
			allStrictChecks := true
			for _, op := range chain {
				// Loose checks that cover both null and undefined
				if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
					allStrictChecks = false
					break
				}
				// Plain/Not operands also check for all falsy values, not just one nullish
				if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
					allStrictChecks = false
					break
				}
			}

			// Only skip if BOTH conditions are met:
			// 1. All subsequent operands have optional chaining
			// 2. All checks are strict (so the pattern is intentional)
			// EXCEPTION: If the first type is not nullable (no null or undefined),
			// then we should still flag for conversion - the strict check is meaningless.
			if allSubsequentHaveOptionalChaining && allStrictChecks {
				firstOp := chain[0]
				firstTypeInfo := processor.getTypeInfo(firstOp.comparedExpr)
				typeHasNeitherNullNorUndefined := !firstTypeInfo.hasNull && !firstTypeInfo.hasUndefined
				if !typeHasNeitherNullNorUndefined {
					continue // Chain uses strict checks with optimal optional chaining, skip it
				}
				// If first type is not nullable, continue processing - the check is meaningless
			}
		}

		// Additional check for 2-operand chains where second has optional in extension
		if len(chain) == 2 {
			firstOp := chain[0]
			secondOp := chain[1]

			// Check if second operand contains optional chaining
			if processor.containsOptionalChain(secondOp.comparedExpr) {
				// If the second operand already has optional chaining and is longer than the first,
				// it's likely already optimal
				// We can check this by seeing if the chain parts already have the right structure
				firstParts := processor.flattenForFix(firstOp.node)
				secondParts := processor.flattenForFix(secondOp.node)

				// SPECIAL CASE: foo && foo?.() or foo.bar && foo.bar?.()
				// If the first and second operands are the same expression,
				// but the second has optional chaining added at the END,
				// this is a redundant check that should be optimized
				// Example: foo && foo?.() -> foo?.()
				// Example: foo.bar && foo.bar?.() -> foo.bar?.()
				isRedundantCheck := false
				if len(secondParts) == len(firstParts)+1 {
					// Check if all parts match except the last
					allMatch := true
					for i := range firstParts {
						if firstParts[i].text != secondParts[i].text || firstParts[i].optional != secondParts[i].optional {
							allMatch = false
							break
						}
					}
					// And the last part of second is optional
					if allMatch && secondParts[len(secondParts)-1].optional {
						// This is expr && expr?.xxx - redundant check, allow optimization
						isRedundantCheck = true
					}
				}

				// If not a redundant check, apply the normal logic
				if !isRedundantCheck {
					// If second is longer and has optional chaining, check if they share a common base
					// and the extension has optional chaining
					if len(secondParts) > len(firstParts) && len(firstParts) > 0 {
						// Check if the bases match (first N parts are the same)
						basesMatch := true
						for i := range firstParts {
							if firstParts[i].text != secondParts[i].text {
								basesMatch = false
								break
							}
						}

						if basesMatch {
							// Check if any part after the first expression's length is optional
							hasOptionalInExtension := false
							for i := len(firstParts); i < len(secondParts); i++ {
								if secondParts[i].optional {
									hasOptionalInExtension = true
									break
								}
							}
							// If we have optional chaining in the extension part, it's already optimal
							if hasOptionalInExtension {
								continue
							}
						}
					}
				}
			}
		}

		// Check if we should apply requireNullish option
		// When requireNullish is true, only convert chains that either:
		// 1. Have explicit nullish check operators (!=null, !==undefined, etc.), OR
		// 2. Have types that explicitly include null/undefined
		if processor.shouldSkipForRequireNullish(chain, true) {
			continue // Skip chains without nullish context when requireNullish is true
		}

		// CRITICAL: Check for void type in plain && chains
		// Optional chaining only checks for null/undefined, not general truthiness
		// void is ALWAYS falsy but not nullish, so converting && to ?. changes behavior
		// Example: x && x() where x is void | (() => void)
		// - Original: if x is void (falsy), short-circuits, returns void
		// - Converted: x?.() attempts to call void (TypeError!)
		// This ONLY applies to plain && chains (not explicit null checks)
		// because explicit checks like x !== null are already checking for nullishness
		if len(chain) > 0 && chain[0].typ == OperandTypePlain {
			// Check if the first operand has void type
			if chain[0].comparedExpr != nil {
				hasVoid := processor.hasVoidType(chain[0].comparedExpr)
				if hasVoid {
					continue // Skip conversion when base has void type
				}
			}
		}

		// Check for non-null assertions without unsafe option
		// Pattern: foo! && foo!.bar should not be converted without unsafe option
		// because the non-null assertion already asserts foo is not null
		// With unsafe option, we can optimize to foo!?.bar
		if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
			// Check if any operand in the chain is a NonNullExpression
			hasNonNullAssertion := false
			for _, op := range chain {
				if op.node != nil && ast.IsNonNullExpression(op.node) {
					hasNonNullAssertion = true
					break
				}
			}
			if hasNonNullAssertion {
				continue // Skip conversion when non-null assertions are present without unsafe option
			}
		}

		// CRITICAL: Check for incomplete nullish checks
		// Optional chaining checks for BOTH null AND undefined
		// If the chain only checks for null OR only for undefined (not both), it's NOT equivalent
		// Example: x !== undefined && x.prop - if x is null, throws error
		//          x?.prop - if x is null, returns undefined safely
		// SAFE patterns:
		// - x != null && x.prop (== checks both)
		// - x !== null && x !== undefined && x.prop (both checks)
		// - typeof x !== 'undefined' && x.prop (special case, safe)
		// - x && x.prop (truthy check, different but acceptable)
		// - x && x.prop !== undefined (truthiness guard + trailing comparison)
		// HOWEVER: Allow incomplete checks when the unsafe option is enabled
		if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
			hasNullCheck := false
			hasUndefinedCheck := false
			hasBothCheck := false
			hasPlainTruthinessCheck := false

			// Only look at GUARD operands - exclude the last operand
			// A trailing comparison/access is when the last operand extends a previous check
			// Example: foo && foo.bar !== undefined
			//   - foo is a truthiness guard (Plain)
			//   - foo.bar !== undefined is a trailing comparison (should be excluded)
			// Example: foo !== null && foo !== undefined && foo.bar
			//   - foo !== null is a null guard
			//   - foo !== undefined is an undefined guard
			//   - foo.bar is an access (Plain) - this is the accessed expression, not a guard
			// Example: foo !== null && foo.bar
			//   - foo !== null is a null guard (incomplete!)
			//   - foo.bar is an access (Plain) - should NOT count as truthiness check
			guardOperands := chain
			if len(chain) >= 2 {
				lastOp := chain[len(chain)-1]
				// Exclude the last operand from guard checks in most cases:
				// - If it's a comparison type, it's a trailing comparison
				// - If it's Plain but extends a previous operand, it's the accessed expression
				if isTrailingComparisonType(lastOp.typ) {
					// Trailing comparison - always exclude
					guardOperands = chain[:len(chain)-1]
				} else if lastOp.typ == OperandTypePlain && lastOp.comparedExpr != nil {
					// Plain operand - check if it extends a previous operand
					prevOp := chain[len(chain)-2]
					if prevOp.comparedExpr != nil {
						lastParts := processor.flattenForFix(lastOp.comparedExpr)
						prevParts := processor.flattenForFix(prevOp.comparedExpr)
						// If last operand is longer (extends previous), it's an access, not a guard
						if len(lastParts) > len(prevParts) {
							guardOperands = chain[:len(chain)-1]
						}
					}
				}
			}

			hasTypeofCheck := false
			// Also check if the TRAILING operand (excluded from guardOperands) is a "both" check
			// If so, the chain is still safe because the result is a boolean nullish check
			hasTrailingBothCheck := false
			if len(chain) >= 2 && len(guardOperands) < len(chain) {
				lastOp := chain[len(chain)-1]
				if lastOp.typ == OperandTypeNotEqualBoth {
					hasTrailingBothCheck = true
				}
			}
			for _, op := range guardOperands {
				if op.typ == OperandTypePlain {
					hasPlainTruthinessCheck = true
				} else if op.typ == OperandTypeNotStrictEqualNull {
					hasNullCheck = true
				} else if op.typ == OperandTypeNotStrictEqualUndef {
					hasUndefinedCheck = true
				} else if op.typ == OperandTypeNotEqualBoth {
					hasBothCheck = true
				} else if op.typ == OperandTypeTypeofCheck {
					// typeof checks are conceptually undefined-only checks
					// BUT converting to optional chaining is SAFE because ?. handles both null AND undefined
					// So typeof checks should NOT be flagged as "incomplete nullish checks"
					hasTypeofCheck = true
					hasUndefinedCheck = true
				}
			}

			// If we have a plain truthiness check as a guard, this is NOT an incomplete nullish check chain
			// Example: foo && foo.bar !== undefined - foo is a truthiness check, not a nullish check
			// The conversion to foo?.bar !== undefined is safe
			// Also, typeof checks should NOT be flagged as incomplete because optional chaining is strictly safer
			// (typeof x !== 'undefined' only checks undefined, but x?.foo checks both null and undefined)
			// Also, if the trailing operand is a "both" check (!= null), the chain is safe
			// ALSO: If the trailing operand already has optional chaining, allow the conversion
			// This handles patterns like: foo.bar !== undefined && foo.bar?.() !== undefined
			// where the user has already opted into optional chaining semantics
			hasTrailingOptionalChaining := false
			if len(chain) >= 2 && len(guardOperands) < len(chain) {
				lastOp := chain[len(chain)-1]
				if lastOp.comparedExpr != nil && processor.containsOptionalChain(lastOp.comparedExpr) {
					hasTrailingOptionalChaining = true
				}
			}
			// ALSO: If the first operand's expression type doesn't include nullish, skip this check
			// The incomplete nullish check is only dangerous when the type COULD be null but we only check undefined
			// If the type can never be null, then checking only for undefined is fine
			firstOpNotNullish := false
			if len(guardOperands) > 0 && guardOperands[0].comparedExpr != nil {
				if !processor.includesNullish(guardOperands[0].comparedExpr) {
					firstOpNotNullish = true
				}
			}

			// Check if the strict check is COMPLETE for the type
			// A strict check is complete when:
			// - `!== null` and type has null but NOT undefined
			// - `!== undefined` and type has undefined but NOT null
			// In these cases, the check is semantically complete and we should report with suggestion
			strictCheckIsComplete := false
			if len(guardOperands) > 0 && guardOperands[0].comparedExpr != nil {
				info := processor.getTypeInfo(guardOperands[0].comparedExpr)
				// Don't consider complete if type is any or unknown (can't determine)
				if !info.hasAny && !info.hasUnknown {
					hasNull := info.hasNull
					hasUndefined := info.hasUndefined
					hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
					hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck

					// !== null is complete if type has null but NOT undefined
					if hasOnlyNullCheck && hasNull && !hasUndefined {
						strictCheckIsComplete = true
					}
					// !== undefined is complete if type has undefined but NOT null
					if hasOnlyUndefinedCheck && hasUndefined && !hasNull {
						strictCheckIsComplete = true
					}
				}
			}

			if !hasPlainTruthinessCheck && !hasBothCheck && !hasTypeofCheck && !hasTrailingBothCheck && !hasTrailingOptionalChaining && !firstOpNotNullish && !strictCheckIsComplete {
				// If we have a strict null check or strict undefined check (but not both), skip
				// This is unsafe regardless of other checks in the chain
				// UNLESS we also have a "both" check (!=) or a typeof check
				// Note: typeof checks count as undefined checks, so if we have typeof + null check, that's complete
				// Check if we have exactly one type of strict check (not both)
				hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
				hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck

				if hasOnlyNullCheck || hasOnlyUndefinedCheck {
					// Skip - incomplete nullish check
					continue
				}
			}
		}

		// Check type-checking options for "loose boolean" operands
		// These options only apply to plain operands (not explicit nullish checks)
		// and only to the FIRST operand (the guard) - subsequent operands are just accesses
		shouldSkip := false
		for i, op := range chain {
			if op.typ == OperandTypePlain {
				// Check if we should skip based on type - only for the first operand (the guard)
				// Subsequent operands are just accesses, not truthiness guards
				if i == 0 {
					if processor.shouldSkipByType(op.comparedExpr) {
						shouldSkip = true
						break
					}
					// Check if conversion would change return type for the FIRST operand only
					// This happens when the type has falsy non-nullish values (like '', 0, false)
					// but does NOT have null/undefined. In this case, && checks for these falsy values
					// but ?. would not, so conversion would change behavior.
					// Example: foo: { bar: string } | '' - && guards against '', ?. doesn't
					// We only check the first operand because that's what && is checking for truthiness.
					// Subsequent operands don't affect whether we can convert - they're accessed after the guard.
					if processor.wouldChangeReturnType(op.comparedExpr) {
						// If allowUnsafe is true, we can still convert (user opted in)
						// If allowUnsafe is false, skip entirely (not even suggestion)
						if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
							shouldSkip = true
							break
						}
					}
				}
				// NOTE: We previously had a check here to skip comparison chains where the
				// plain operand doesn't include nullish. However, upstream TypeScript-ESLint
				// converts these patterns anyway, even if they're semantically redundant.
				// Example: declare const foo: { bar: number }; foo && foo.bar != null;
				// -> foo?.bar != null (converted even though foo is never null)
			}
			// Also check explicit nullish checks, negated operands, and inverted checks for non-nullish types
			// Example: foo != null && foo.bar != 0 where foo: { bar: number } (never nullish)
			// Example: !foo && foo.bar === 0 where foo: { bar: number } (never nullish)
			// Example: foo == null && foo.bar === 0 where foo: { bar: number } (never nullish)
			// NOTE: We previously skipped these when there was a comparison and the expression
			// didn't include nullish. However, upstream TypeScript-ESLint converts these patterns
			// anyway, treating them as stylistic improvements even when semantically redundant.
			//
			// The only cases we still skip are:
			// 1. typeof checks on non-nullable expressions with call expressions
			//    (which could cause runtime errors if converted)
			if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeNotStrictEqualNull ||
				op.typ == OperandTypeNotStrictEqualUndef || op.typ == OperandTypeTypeofCheck ||
				op.typ == OperandTypeNegatedAndOperand || op.typ == OperandTypeEqualNull ||
				op.typ == OperandTypeStrictEqualNull || op.typ == OperandTypeStrictEqualUndef {

				// Special case: typeof check on non-nullable expression with call expression
				// Example: typeof globalThis !== 'undefined' && globalThis.Array()
				// If globalThis is never undefined, this pattern shouldn't be converted
				// UNLESS there's a middle guard operand that checks the property:
				// Example: typeof globalThis !== 'undefined' && globalThis.Array && globalThis.Array()
				// The middle guard (globalThis.Array) makes the conversion safe.
				if op.typ == OperandTypeTypeofCheck && op.comparedExpr != nil && !processor.includesNullish(op.comparedExpr) {
					// Only skip if this is a 2-operand chain (typeof + call without middle guard)
					// If there are more operands, the middle ones provide guards
					if len(chain) == 2 {
						// Check if last operand is a call expression
						lastOp := chain[len(chain)-1]
						if lastOp.comparedExpr != nil {
							unwrapped := unwrapParentheses(lastOp.comparedExpr)
							if ast.IsCallExpression(unwrapped) {
								shouldSkip = true
								break
							}
						}
					}
				}
			}
		}
		if shouldSkip {
			continue // Skip this chain, but process others
		}

		// Check if trailing comparison would change falsy to truthy when base is nullish.
		// For AND chains like `foo && foo.bar OP X`:
		// - If foo is nullish, the original returns foo (falsy)
		// - Converting to `foo?.bar OP X` returns `undefined OP X`
		// The conversion is SAFE only if `undefined OP X` is also FALSY.
		//
		// Safe patterns (undefined OP X is falsy):
		// - == literal (undefined == 0 is false)
		// - === literal (undefined === 0 is false, undefined === null is false)
		// - != null/undefined (undefined != null is false)
		// - !== undefined (undefined !== undefined is false)
		//
		// Unsafe patterns (undefined OP X is truthy):
		// - == null/undefined (undefined == null is true)
		// - === undefined (undefined === undefined is true)
		// - != literal (undefined != 0 is true)
		// - !== null (undefined !== null is true)
		// - !== literal (undefined !== 'x' is true)
		if len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			// Check if this is a trailing comparison (either OperandTypeComparison or a null/undefined check
			// that is on a SUPERSET of the previous operand - meaning it's a comparison on the accessed value,
			// not a guard for accessing it)
			isTrailingComparison := lastOp.typ == OperandTypeComparison
			if !isTrailingComparison && len(chain) >= 2 &&
				(lastOp.typ == OperandTypeNotStrictEqualNull ||
					lastOp.typ == OperandTypeNotStrictEqualUndef ||
					lastOp.typ == OperandTypeNotEqualBoth) {
				// Check if the last operand extends a previous operand (making it a trailing comparison)
				prevOp := chain[len(chain)-2]
				if lastOp.comparedExpr != nil && prevOp.comparedExpr != nil {
					lastParts := processor.flattenForFix(lastOp.comparedExpr)
					prevParts := processor.flattenForFix(prevOp.comparedExpr)
					// If last operand is longer (extends previous), it's a trailing comparison
					if len(lastParts) > len(prevParts) {
						isTrailingComparison = true
					}
				}
			}

			if isTrailingComparison && lastOp.node != nil {
				unwrappedNode := unwrapParentheses(lastOp.node)
				if ast.IsBinaryExpression(unwrappedNode) {
					binExpr := unwrappedNode.AsBinaryExpression()
					op := binExpr.OperatorToken.Kind

					// Determine the comparison value (not the property access)
					var value *ast.Node
					if ast.IsPropertyAccessExpression(binExpr.Left) || ast.IsElementAccessExpression(binExpr.Left) || ast.IsCallExpression(binExpr.Left) {
						value = binExpr.Right
					} else if ast.IsPropertyAccessExpression(binExpr.Right) || ast.IsElementAccessExpression(binExpr.Right) || ast.IsCallExpression(binExpr.Right) {
						value = binExpr.Left
					}

					if value != nil {
						isNull := value.Kind == ast.KindNullKeyword
						isUndefined := (ast.IsIdentifier(value) && value.AsIdentifier().Text == "undefined") || ast.IsVoidExpression(value)
						isNullish := isNull || isUndefined
						isLiteral := value.Kind == ast.KindNumericLiteral ||
							value.Kind == ast.KindStringLiteral ||
							value.Kind == ast.KindTrueKeyword ||
							value.Kind == ast.KindFalseKeyword ||
							value.Kind == ast.KindObjectLiteralExpression ||
							value.Kind == ast.KindArrayLiteralExpression
						isUndeclaredVar := ast.IsIdentifier(value) && !isUndefined && !isLiteral

						unsafe := false
						switch op {
						case ast.KindEqualsEqualsToken: // ==
							// == null/undefined → unsafe (undefined == null is true)
							// == literal → safe (undefined == 0 is false)
							// == undeclaredVar → unsafe (could be anything)
							if isNullish || isUndeclaredVar {
								unsafe = true
							}
						case ast.KindEqualsEqualsEqualsToken: // ===
							// === undefined → unsafe (undefined === undefined is true)
							// === null → safe (undefined === null is false)
							// === literal → safe (undefined === 0 is false)
							// === undeclaredVar → unsafe (could be undefined)
							if isUndefined || isUndeclaredVar {
								unsafe = true
							}
						case ast.KindExclamationEqualsToken: // !=
							// != null/undefined → safe (undefined != null is false)
							// != literal → unsafe (undefined != 0 is true)
							// != undeclaredVar → unsafe (could be anything)
							if !isNullish {
								unsafe = true
							}
						case ast.KindExclamationEqualsEqualsToken: // !==
							// !== undefined → safe (undefined !== undefined is false)
							// !== null → unsafe (undefined !== null is true)
							// !== literal → unsafe (undefined !== 'x' is true)
							// !== undeclaredVar → unsafe (could be anything)
							if !isUndefined {
								unsafe = true
							}
						}

						if unsafe && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
							continue // Skip this chain
						}
					}
				}
			}
		}

		// Build the optional chain
		// Find the last actual property access (plain or comparison)
		var lastPropertyAccess *ast.Node
		var hasTrailingComparison bool
		var hasTrailingTypeofCheck bool
		var hasComplementaryNullCheck bool      // true if last two operands form a complementary null+undefined check
		var complementaryTrailingNode *ast.Node // the last operand's node to append as trailing text

		// Check if the last two operands form a complementary pair (null + undefined checks on same expression)
		// When this is true, we DON'T simplify to `!= null`. Instead:
		// - Use the SECOND-TO-LAST operand as the chain endpoint
		// - Append the LAST operand as trailing text (preserving it)
		// - Make it a SUGGESTION because the code can't be fully simplified
		if len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			secondLastOp := chain[len(chain)-2]

			// Check if they're on the same expression
			if lastOp.comparedExpr != nil && secondLastOp.comparedExpr != nil {
				cmpResult := processor.compareNodes(lastOp.comparedExpr, secondLastOp.comparedExpr)
				if cmpResult == NodeEqual {
					// Same expression - check if they form a complementary pair
					isLastUndef := lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeTypeofCheck
					isLastNull := lastOp.typ == OperandTypeNotStrictEqualNull
					isSecondLastUndef := secondLastOp.typ == OperandTypeNotStrictEqualUndef || secondLastOp.typ == OperandTypeTypeofCheck
					isSecondLastNull := secondLastOp.typ == OperandTypeNotStrictEqualNull

					if (isLastUndef && isSecondLastNull) || (isLastNull && isSecondLastUndef) {
						// Complementary pair! Don't simplify - use second-to-last as chain end, keep last as trailing
						hasComplementaryNullCheck = true
						lastPropertyAccess = secondLastOp.comparedExpr
						complementaryTrailingNode = lastOp.node
						// Set flags based on the second-to-last operand (the one we're including in the chain)
						hasTrailingComparison = true
						hasTrailingTypeofCheck = secondLastOp.typ == OperandTypeTypeofCheck
					}
				}
			}
		}

		if !hasComplementaryNullCheck {
			for i := len(chain) - 1; i >= 0; i-- {
				if chain[i].typ == OperandTypePlain {
					// For plain operands, use node to preserve NonNull assertions
					lastPropertyAccess = chain[i].node
					hasTrailingComparison = false
					hasTrailingTypeofCheck = false
					break
				} else if chain[i].typ == OperandTypeComparison ||
					chain[i].typ == OperandTypeNotStrictEqualNull ||
					chain[i].typ == OperandTypeNotStrictEqualUndef ||
					chain[i].typ == OperandTypeNotEqualBoth {
					// For comparison operands, use comparedExpr and mark as having trailing comparison
					lastPropertyAccess = chain[i].comparedExpr
					hasTrailingComparison = true
					hasTrailingTypeofCheck = false
					break
				} else if chain[i].typ == OperandTypeTypeofCheck {
					// For typeof checks, use comparedExpr and mark as having trailing typeof
					lastPropertyAccess = chain[i].comparedExpr
					hasTrailingComparison = true
					hasTrailingTypeofCheck = true
					break
				} else if chain[i].comparedExpr != nil {
					// For other operands with comparedExpr (like OperandTypeNot), use comparedExpr
					// but don't mark as trailing comparison
					lastPropertyAccess = chain[i].comparedExpr
					hasTrailingComparison = false
					hasTrailingTypeofCheck = false
					break
				}
			}
		}

		if lastPropertyAccess == nil {
			continue // Skip this chain, but process others
		}

		parts := processor.flattenForFix(lastPropertyAccess)

		// For type assertions, the first operand may have a more complete type annotation
		// e.g., (foo as T | null) && (foo as T).bar
		// The first operand's base (foo as T | null) should be used instead of (foo as T)
		// Check if the first operand's base has a longer type assertion
		// Note: Only do this for OperandTypePlain - for other types (comparisons etc.),
		// the node includes the comparison operator which we don't want
		if len(chain) > 0 && len(parts) > 0 && chain[0].typ == OperandTypePlain && chain[0].node != nil {
			firstParts := processor.flattenForFix(chain[0].node)
			if len(firstParts) == 1 && len(firstParts) <= len(parts) {
				// Only replace base if the first operand is a single expression (not a chain)
				// and if first is longer (has more text), use it
				// This handles cases like (foo as T | null) vs (foo as T)
				if len(firstParts[0].text) > len(parts[0].text) {
					parts[0] = firstParts[0]
				}
			}
		}

		// Find all checked lengths to determine which properties should be optional
		// For: foo && foo.bar && foo.bar.baz.buzz && foo.bar.baz.buzz()
		// Checks: [foo, foo.bar, foo.bar.baz.buzz] with lengths [1, 2, 4]
		// Make optional at indices where we checked the parent:
		// - Index 1 (bar): checked length 1 (foo) ✓
		// - Index 2 (baz): checked length 2 (foo.bar) ✓
		// - Index 3 (buzz): checked length 3? NO, jumped from 2 to 4 ✗
		// - Index 4 (call): checked length 4 (foo.bar.baz.buzz) ✓
		checkedLengths := make(map[int]bool)

		// Find all checks (not including the last operand if it's just an access)
		// We want to exclude the final access that we're converting, but include
		// all the checks that happen before it
		checksToConsider := []Operand{}
		for i := range chain {
			op := chain[i]
			// Skip the last operand if it's the final access (not a check)
			isLastOperand := i == len(chain)-1
			isCallAccess := false
			if op.comparedExpr != nil && ast.IsCallExpression(op.comparedExpr) {
				isCallAccess = true
			}

			// Include all operands except:
			// 1. The last plain operand (foo.bar.baz in foo && foo.bar && foo.bar.baz)
			// 2. The last call operand (foo.bar() in !foo.bar || !foo.bar())
			if isLastOperand && (op.typ == OperandTypePlain || (op.typ == OperandTypeNot && isCallAccess)) {
				continue
			}

			checksToConsider = append(checksToConsider, op)
		}

		// First, count how many non-typeof checks we have
		hasNonTypeofCheck := false
		for _, operand := range checksToConsider {
			if operand.typ != OperandTypeTypeofCheck && operand.comparedExpr != nil {
				hasNonTypeofCheck = true
				break
			}
		}

		for _, operand := range checksToConsider {
			if operand.comparedExpr != nil {
				// Skip typeof checks when populating checkedLengths IF there are other checks
				// typeof checks verify existence (not nullability) of the identifier
				// They should NOT cause the immediate next property to be optional
				// WHEN there's a middle guard that does the actual null check.
				//
				// Example: typeof globalThis !== 'undefined' && globalThis.foo && globalThis.foo.bar
				// - typeof check on globalThis should NOT make .foo optional (there's a middle guard)
				// - Check on globalThis.foo SHOULD make .bar optional
				// Result: globalThis.foo?.bar (not globalThis?.foo?.bar)
				//
				// But: typeof foo !== 'undefined' && foo.bar
				// - typeof check is the ONLY check, so it SHOULD make .bar optional
				// Result: foo?.bar
				if operand.typ == OperandTypeTypeofCheck && hasNonTypeofCheck {
					continue
				}
				checkedParts := processor.flattenForFix(operand.comparedExpr)
				checkedLengths[len(checkedParts)] = true
			}
		}

		// Fill in gaps in checkedLengths
		// Two cases to handle:
		//
		// Case 1: Single check with deep access
		// Example: foo && (foo.bar).baz
		// - Check foo (length 1), access foo.bar.baz (length 3)
		// - Make index 1 (.bar) and 2 (.baz) optional
		// - checkedLengths before: {1}
		// - checkedLengths after: {1, 2}
		//
		// Case 2: Multiple checks
		// Example: foo && foo.bar && foo.bar.baz.buzz
		// - Check foo (length 1), check foo.bar (length 2), access foo.bar.baz.buzz (length 4)
		// - Make index 1 (.bar) and 2 (.baz) optional
		// - DON'T make index 3 (.buzz) optional
		// - checkedLengths before: {1, 2}
		// - checkedLengths after: {1, 2}
		//
		// Strategy: If we have only one check, fill up to (last plain length - 1)
		//           If we have multiple checks, only fill gaps between checks

		minChecked := -1
		maxChecked := -1
		numChecks := len(checkedLengths)
		for length := range checkedLengths {
			if minChecked == -1 || length < minChecked {
				minChecked = length
			}
			if maxChecked == -1 || length > maxChecked {
				maxChecked = length
			}
		}

		if minChecked > 0 {
			var fillUpTo int
			if numChecks == 1 {
				// Single check: behavior depends on where the check is
				if minChecked == 1 {
					// Check at the start (e.g., foo && foo.bar.baz)
					// Fill up to the second-to-last part
					if len(chain) > 0 && chain[len(chain)-1].typ == OperandTypePlain {
						lastPlainParts := processor.flattenForFix(chain[len(chain)-1].node)
						fillUpTo = len(lastPlainParts) - 1

						// Don't fill up to include a call - calls are handled separately by callShouldBeOptional
						if fillUpTo > 0 && len(lastPlainParts) > 0 {
							lastPart := lastPlainParts[len(lastPlainParts)-1]
							isCall := strings.HasPrefix(lastPart.text, "(") || strings.HasPrefix(lastPart.text, "<(")
							if isCall {
								fillUpTo--
							}
						}
					} else {
						fillUpTo = maxChecked
					}
					for i := minChecked; i <= fillUpTo; i++ {
						if !checkedLengths[i] {
							checkedLengths[i] = true
						}
					}
				} else {
					// Check in the middle (e.g., foo.bar && foo.bar.baz.buzz)
					// Only make the immediate next property optional
					// No filling needed - checkedLengths already has the right index
				}
			} else {
				// Multiple checks: don't fill gaps!
				// Only use the exact check lengths that were found
				// Example: foo && foo.bar && foo.bar.baz.buzz
				// Checks at [1, 2], don't auto-add index 3 or 4
				// The checkedLengths map already has the right values from line 2227
			}
		}

		// Replace parts from earlier operands for the checked prefix
		// This ensures we use foo?.bar.baz (preserving ?.) when the first operand had optional chains
		// or foo!.bar.baz (not foo.bar.baz) when an earlier operand had non-null assertions
		// Example: foo?.bar.baz != null && foo.bar?.baz.bam != null
		// - First operand: foo?.bar.baz, parts: [foo, bar(?.opt), baz]
		// - Last operand: foo.bar?.baz.bam, parts: [foo, bar, baz(?.opt), bam]
		// - First operand is a PROPER prefix of last, so replace indices 0,1,2 from first operand
		// - Result: [foo, bar(?.opt), baz, bam] with first operand's optional flags preserved
		if len(checksToConsider) > 0 && len(parts) > 1 {
			// Find the maximum checked length (the longest check operand)
			maxCheckedLen := 0
			for _, op := range checksToConsider {
				if op.comparedExpr != nil {
					opParts := processor.flattenForFix(op.comparedExpr)
					if len(opParts) > maxCheckedLen {
						maxCheckedLen = len(opParts)
					}
				}
			}

			// First, STRIP non-null assertions from parts that are within the checked range
			// This ensures we use the check operands' ! state, not the last operand's
			// Example: foo!.bar != null && foo.bar!.baz != null
			// - parts from last operand: [foo, bar!, baz]
			// - maxCheckedLen = 2 (from foo!.bar)
			// - We strip ! from parts[0] and parts[1] (indices < maxCheckedLen)
			for i := 0; i < maxCheckedLen && i < len(parts); i++ {
				parts[i].hasNonNull = false
				parts[i].text = strings.TrimSuffix(parts[i].text, "!")
			}

			// Collect all check operand parts for later use
			type opPartsInfo struct {
				parts []ChainPart
				len   int
			}
			var allOpParts []opPartsInfo
			for _, op := range checksToConsider {
				if op.comparedExpr != nil {
					exprToFlatten := op.comparedExpr
					if op.typ == OperandTypePlain {
						exprToFlatten = op.node
					}
					opParts := processor.flattenForFix(exprToFlatten)
					// Check if opParts is a prefix of parts
					if len(opParts) <= len(parts) {
						isPrefix := true
						for i := range opParts {
							opText := strings.TrimSuffix(opParts[i].text, "!")
							partText := strings.TrimSuffix(parts[i].text, "!")
							if opText != partText {
								isPrefix = false
								break
							}
						}
						if isPrefix {
							allOpParts = append(allOpParts, opPartsInfo{parts: opParts, len: len(opParts)})
						}
					}
				}
			}

			// For each part index, merge optional and non-null flags
			// For optional chains (?.), use the SHORTEST operand that covers this index
			// This ensures we preserve the optional state from the earliest check (the one that validates the path)
			// Example: foo.bar.baz != null && foo?.bar?.baz.bam != null
			// - First operand (foo.bar.baz) is shorter and has NO optional chaining
			// - So we should NOT make bar/baz optional, even though second operand has them optional
			// - Result: foo.bar.baz?.bam (only the extension is optional)
			//
			// For non-null assertions (!), use the SHORTEST operand that covers this index
			// This ensures we preserve ! from the earliest check, not from later extended checks
			for i := range parts {
				// Find the shortest operand that covers this index
				var shortestCoveringOp *opPartsInfo
				for j := range allOpParts {
					op := &allOpParts[j]
					if i < op.len {
						if shortestCoveringOp == nil || op.len < shortestCoveringOp.len {
							shortestCoveringOp = op
						}
					}
				}

				// Use the shortest covering operand for optional flag
				// This ensures we respect the first check's optional state
				if shortestCoveringOp != nil && i < shortestCoveringOp.len {
					parts[i].optional = shortestCoveringOp.parts[i].optional
				}

				// Use only the shortest covering operand for non-null assertion
				if shortestCoveringOp != nil && i < shortestCoveringOp.len {
					if shortestCoveringOp.parts[i].hasNonNull {
						parts[i].hasNonNull = true
						if !strings.HasSuffix(parts[i].text, "!") {
							parts[i].text = parts[i].text + "!"
						}
					}
				}
			}

		}

		// Don't normalize parts - we want to preserve existing optional flags
		// and only ADD new ones based on checks
		// For example: foo?.bar.baz should keep bar as non-optional if that's how it appears

		// Check if we're explicitly checking the function being called
		// e.g., foo && foo.bar && foo.bar.baz && foo.bar.baz()
		// We check foo.bar.baz before calling it, so the call should be optional
		callShouldBeOptional := false
		if len(parts) > 0 && strings.HasPrefix(parts[len(parts)-1].text, "(") {
			// Last part is a call expression
			// Check if we have a check for the expression without the call
			partsWithoutCall := len(processor.flattenForFix(lastPropertyAccess.AsCallExpression().Expression))

			for _, op := range chain[:len(chain)-1] { // Don't check the last operand (the call itself)
				// Use comparedExpr to get the actual expression being checked (without ! or comparisons)
				if op.comparedExpr != nil {
					checkedParts := processor.flattenForFix(op.comparedExpr)

					// If we checked all parts except the call, the call should be optional
					if len(checkedParts) == partsWithoutCall {
						callShouldBeOptional = true
						break
					}
				}
			}
		}

		newCode := processor.buildOptionalChain(parts, checkedLengths, callShouldBeOptional, false) // false = preserve ! assertions for AND chains

		// If buildOptionalChain returned empty string, it means we'd create invalid syntax
		// (e.g., ?.#privateIdentifier which TypeScript doesn't allow)
		if newCode == "" {
			return
		}

		// Preserve leading trivia (comments, whitespace) from operands after the first one.
		// For: foo && /* important */ foo.bar
		// The comment /* important */ is leading trivia of foo.bar and should be preserved.
		// We extract trivia from each operand after the first and prepend it to newCode.
		// Note: We trim leading whitespace since that was between the && and the comment.
		if len(chain) > 1 {
			var leadingTrivia strings.Builder
			for i := 1; i < len(chain); i++ {
				opNode := chain[i].node
				if opNode != nil {
					// Full position includes leading trivia, trimmed position skips it
					fullPos := opNode.Pos()
					trimmedRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, opNode)
					trimmedPos := trimmedRange.Pos()
					if fullPos < trimmedPos {
						trivia := processor.sourceText[fullPos:trimmedPos]
						leadingTrivia.WriteString(trivia)
					}
				}
			}
			if leadingTrivia.Len() > 0 {
				// Trim leading whitespace but preserve comments
				triviaStr := strings.TrimLeft(leadingTrivia.String(), " \t\n\r")
				if triviaStr != "" {
					newCode = triviaStr + newCode
				}
			}
		}

		// Check if the last operand is a comparison - if so, append/prepend it
		if hasTrailingComparison {
			// For complementary null+undefined checks, use the SECOND-TO-LAST operand
			// as the comparison and append the LAST operand as trailing text
			var operandForComparison Operand
			if hasComplementaryNullCheck {
				operandForComparison = chain[len(chain)-2]
			} else {
				operandForComparison = chain[len(chain)-1]
			}

			if ast.IsBinaryExpression(operandForComparison.node) {
				binExpr := operandForComparison.node.AsBinaryExpression()

				// Special handling for typeof checks: typeof foo.bar !== 'undefined'
				// The binary expression is: (typeof foo.bar) !== 'undefined'
				// We need to wrap the optional chain with: typeof ... !== 'undefined'
				if hasTrailingTypeofCheck {
					// For typeof checks, we need to:
					// 1. Get the "typeof " prefix from the left side
					// 2. Get the " !== 'undefined'" suffix from after the comparedExpr
					leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)
					comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.comparedExpr)

					// typeof prefix: from start of left side to start of comparedExpr
					typeofPrefix := processor.sourceText[leftRange.Pos():comparedExprRange.Pos()]

					// comparison suffix: from end of comparedExpr to end of binary expression
					binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).End()
					comparisonSuffix := processor.sourceText[comparedExprRange.End():binExprEnd]

					newCode = typeofPrefix + newCode + comparisonSuffix
				} else {
					// Check if this is a yoda condition (literal/constant on left, property on right)
					// In yoda: '123' == foo.bar.baz
					// Not yoda: foo.bar.baz == '123'
					isYoda := false
					comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.comparedExpr)
					leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)

					// If comparedExpr is on the right side, it's yoda
					if comparedExprRange.Pos() > leftRange.Pos() {
						isYoda = true
					}

					if isYoda {
						// Yoda: prepend the left side + operator
						binExprStart := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).Pos()
						comparedExprStart := comparedExprRange.Pos()
						yodaPrefix := processor.sourceText[binExprStart:comparedExprStart]
						newCode = yodaPrefix + newCode
					} else {
						// Normal: append the operator + right side
						comparedExprEnd := comparedExprRange.End()
						binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).End()
						comparisonSuffix := processor.sourceText[comparedExprEnd:binExprEnd]
						newCode = newCode + comparisonSuffix
					}
				}
			}

			// For complementary null+undefined checks, append the last operand as trailing text
			if hasComplementaryNullCheck && complementaryTrailingNode != nil {
				// Get the text of the last operand including the && before it
				// We need to find the && between the second-to-last and last operands
				secondLastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-2].node)
				lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, complementaryTrailingNode)
				// The text between includes " && " or similar
				betweenText := processor.sourceText[secondLastRange.End():lastRange.Pos()]
				lastText := processor.sourceText[lastRange.Pos():lastRange.End()]
				newCode = newCode + betweenText + lastText
			}
		}

		// Use trimmed ranges to preserve leading/trailing whitespace
		// If we're replacing the entire logical expression (all operands from this node),
		// use the node's range to include any wrapping parentheses
		// Otherwise, use the operand ranges
		var replaceStart, replaceEnd int

		// Determine the effective chain start for replacement
		// When the first operand is a typeof check on an UNDECLARED variable AND there are
		// other non-typeof checks, we must PRESERVE the typeof check because it guards
		// against ReferenceError for potentially-undeclared globals.
		//
		// Example: typeof globalThis !== 'undefined' && globalThis.Array && globalThis.Array()
		// - globalThis might not exist in older environments (no declaration)
		// - The typeof check prevents ReferenceError if globalThis doesn't exist
		// - We can only transform: globalThis.Array && globalThis.Array() -> globalThis.Array?.()
		// - Result: typeof globalThis !== 'undefined' && globalThis.Array?.()
		//
		// But: function foo(globalThis?: ...) { typeof globalThis !== 'undefined' && globalThis.Array() }
		// - globalThis is a DECLARED parameter, so it always exists (might be undefined, but won't throw)
		// - Can be fully transformed to: globalThis?.Array()
		effectiveChainStart := 0
		if len(chain) >= 2 && chain[0].typ == OperandTypeTypeofCheck {
			// Check if there are non-typeof checks after the first operand
			hasNonTypeofAfterFirst := false
			for i := 1; i < len(chain); i++ {
				if chain[i].typ != OperandTypeTypeofCheck {
					hasNonTypeofAfterFirst = true
					break
				}
			}
			if hasNonTypeofAfterFirst && chain[0].comparedExpr != nil {
				// Check if the typeof target is a declared variable
				// If it has no symbol or no declarations, it's potentially undeclared
				typeofTarget := chain[0].comparedExpr
				symbol := processor.ctx.TypeChecker.GetSymbolAtLocation(typeofTarget)
				isUndeclared := symbol == nil || len(symbol.Declarations) == 0

				if isUndeclared {
					// Preserve the typeof check - start replacement from second operand
					effectiveChainStart = 1
				}
			}
		}

		if effectiveChainStart == 0 && len(chain) == len(operandNodes) {
			// We're replacing all operands - use the top-level node range
			nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
			replaceStart = nodeRange.Pos()
			replaceEnd = nodeRange.End()
		} else {
			// We're replacing a subset - use operand ranges
			// Start from effectiveChainStart to preserve leading typeof checks
			firstNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[effectiveChainStart].node)
			lastNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-1].node)
			replaceStart = firstNodeRange.Pos()
			replaceEnd = lastNodeRange.End()
		}

		fixes := []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(replaceStart, replaceEnd), newCode),
		}

		// Determine if we should autofix or suggest
		// Following typescript-eslint's logic:
		// 1. If allowPotentiallyUnsafe option is enabled, always autofix
		// 2. If there's a trailing comparison or the last operand has a nullish comparison type, autofix
		// 3. Otherwise, check if ANY operand includes undefined (including any/unknown) - if so, autofix
		// 4. Default to suggestion
		useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

		if useSuggestion && len(chain) > 0 {
			lastOp := chain[len(chain)-1]

			// Check if the last operand has a comparison type that makes autofix safe
			// These comparison types ensure the return type remains consistent (boolean)
			switch lastOp.typ {
			case OperandTypeEqualNull, // == null (covers both null and undefined)
				OperandTypeNotEqualBoth,        // != null (covers both)
				OperandTypeStrictEqualUndef,    // === undefined
				OperandTypeNotStrictEqualUndef: // !== undefined
				useSuggestion = false
			case OperandTypeTypeofCheck:
				// typeof checks are safe to autofix
				useSuggestion = false
			}

			// For AND chains with trailing comparisons (non-nullish), always provide a fix
			// The comparison ensures the return type remains consistent
			// Example: foo && foo.bar === 0 -> foo?.bar === 0 (both return boolean)
			if useSuggestion && hasTrailingComparison {
				useSuggestion = false
			}

			// If still using suggestion, check if ANY operand includes undefined
			// (including any/unknown types). If so, the autofix is safe because
			// optional chaining will union undefined into a type that already has it.
			// NOTE: We specifically check for UNDEFINED, not null. If the type only
			// has null (not undefined), converting would change return type from null
			// to undefined, which is an unsafe change.
			if useSuggestion {
				for _, op := range chain {
					if op.comparedExpr != nil {
						info := processor.getTypeInfo(op.comparedExpr)
						// Safe to autofix if type includes undefined, any, or unknown
						// (because ?. returns undefined, so adding undefined is safe)
						if info.hasUndefined || info.hasAny || info.hasUnknown {
							useSuggestion = false
							break
						}
					}
				}
			}

			// EXCEPTION: When the first operand is an explicit nullish comparison
			// (like foo != null or foo !== undefined) AND the last operand is a
			// plain access (not a comparison), the return type changes from
			// "false | T" to "undefined | T". If the operand's type doesn't include
			// any/unknown (which would mask this difference), use suggestion.
			if !useSuggestion && len(chain) > 0 {
				firstOp := chain[0]
				lastOp := chain[len(chain)-1]
				// Check if first operand is an explicit nullish comparison
				isExplicitNullishCheck := firstOp.typ == OperandTypeNotEqualBoth ||
					firstOp.typ == OperandTypeNotStrictEqualNull ||
					firstOp.typ == OperandTypeNotStrictEqualUndef
				// Check if last operand is a plain access (not a comparison)
				isPlainAccess := lastOp.typ == OperandTypePlain

				if isExplicitNullishCheck && isPlainAccess {
					// Check if the first operand's type includes any/unknown
					// If not, the return type change is observable, so use suggestion
					if firstOp.comparedExpr != nil {
						info := processor.getTypeInfo(firstOp.comparedExpr)
						if !info.hasAny && !info.hasUnknown {
							useSuggestion = true
						}
					}
				}
			}

			// For complementary null+undefined checks, use suggestion when:
			// 1. The trailing operand uses typeof, OR
			// 2. The included operand (second-to-last) is a null check (!== null)
			//
			// Reason: If we convert foo?.bar !== null, and foo is undefined,
			// the result is (undefined !== null) = true, which might not be expected.
			// But foo?.bar !== undefined returns false when foo is undefined, which is expected.
			//
			// Example: null !== foo.bar.baz && undefined !== foo.bar.baz -> SUGGESTION
			//          (because the chain uses !== null which returns true for undefined)
			// Example: undefined !== foo.bar.baz && null !== foo.bar.baz -> FIX
			//          (because the chain uses !== undefined which returns false for undefined)
			if hasComplementaryNullCheck && len(chain) >= 2 {
				lastOp := chain[len(chain)-1]
				secondLastOp := chain[len(chain)-2]
				// Check if trailing uses typeof
				if lastOp.typ == OperandTypeTypeofCheck {
					useSuggestion = true
				}
				// Check if the included check (second-to-last) is a null check
				if secondLastOp.typ == OperandTypeNotStrictEqualNull {
					useSuggestion = true
				}
			}
		}

		if useSuggestion {
			processor.ctx.ReportNodeWithSuggestions(node, buildPreferOptionalChainMessage(), func() []rule.RuleSuggestion {
				return []rule.RuleSuggestion{{
					Message:  buildOptionalChainSuggestMessage(),
					FixesArr: fixes,
				}}
			})
		} else {
			processor.ctx.ReportNodeWithFixes(node, buildPreferOptionalChainMessage(), func() []rule.RuleFix {
				return fixes
			})
		}

		// Mark all operands in this chain as reported to avoid overlapping diagnostics
		for _, op := range chain {
			opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
			opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
			processor.reportedRanges[opTextRange] = true
		}
	} // End of for _, chain := range allChains
}

func (processor *chainProcessor) processOrChain(node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindBarBarToken {
		return
	}

	// Check if this node is part of a LARGER || chain (i.e., it has a || parent)
	// If so, skip processing - the parent will handle the entire chain
	// We traverse up the parent chain, skipping parentheses, to find if there's
	// an enclosing || expression where this node is an operand
	parent := node.Parent
	for parent != nil {
		if ast.IsParenthesizedExpression(parent) {
			// Skip parentheses and continue up
			parent = parent.Parent
			continue
		}
		if ast.IsBinaryExpression(parent) {
			parentBin := parent.AsBinaryExpression()
			if parentBin.OperatorToken.Kind == ast.KindBarBarToken {
				// Check if this node is actually an operand of this || (not just nested inside)
				// This node is nested inside if it's the Left or Right (possibly wrapped in parens)
				leftUnwrapped := unwrapParentheses(parentBin.Left)
				rightUnwrapped := unwrapParentheses(parentBin.Right)
				if leftUnwrapped == node || rightUnwrapped == node {
					// This is a nested || expression, skip it
					return
				}
			}
		}
		// Stop traversing - we only skip nested || parens
		break
	}

	// When requireNullish is true, only skip || chains with negation (!foo || !foo.bar)
	// Allow || chains with explicit null checks (foo == null || foo.bar)
	// We'll filter later based on operand types

	// Skip if already seen (use range-based check for reliability)
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeTextRange := textRange{start: nodeRange.Pos(), end: nodeRange.End()}
	if processor.seenLogicalRanges[nodeTextRange] {
		return
	}
	processor.seenLogicalRanges[nodeTextRange] = true

	// Skip if inside JSX - semantic difference
	if isInsideJSX(node) {
		return
	}

	// Collect all || operands and binary expression ranges using the shared helper
	operandNodes, collectedBinaryRanges := processor.collectOperandsWithRanges(node, ast.KindBarBarToken)

	// Mark all collected binary expression ranges as seen
	for _, r := range collectedBinaryRanges {
		processor.seenLogicalRanges[r] = true
	}

	if len(operandNodes) < 2 {
		return
	}

	// Check if any operand has already been reported
	for _, n := range operandNodes {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
		opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
		if processor.reportedRanges[opTextRange] {
			return
		}
	}

	// Parse operands
	operands := make([]Operand, len(operandNodes))
	for i, n := range operandNodes {
		operands[i] = processor.parseOperand(n, false)
	}

	// Look for pattern: !foo || !foo.bar or foo == null || foo.bar != 0
	var chain []Operand
	var lastExpr *ast.Node
	var hasTrailingComparison bool

	for i := range operands {
		op := operands[i]

		// Accept OperandTypeNot, OperandTypeComparison, OperandTypePlain, typeof checks, and null check types
		// In OR chains, both !== and === null checks are valid:
		// - !== null: foo !== null || foo.bar (checks if NOT null, short-circuits if null)
		// - === null: !a || a.b === null || !a.b.c (checks if IS null, returns true if null)
		// Both patterns can be converted to optional chaining because they're checking for nullish values
		validOrOperand := op.typ == OperandTypeNot ||
			op.typ == OperandTypeComparison ||
			op.typ == OperandTypePlain ||
			op.typ == OperandTypeTypeofCheck ||
			op.typ == OperandTypeNotStrictEqualNull ||
			op.typ == OperandTypeNotStrictEqualUndef ||
			op.typ == OperandTypeNotEqualBoth ||
			op.typ == OperandTypeStrictEqualNull ||
			op.typ == OperandTypeStrictEqualUndef ||
			op.typ == OperandTypeEqualNull

		if !validOrOperand {
			// Not a valid operand type for OR chain
			if len(chain) >= 2 {
				break
			}
			chain = nil
			lastExpr = nil
			continue
		}

		if len(chain) == 0 {
			// CRITICAL: In OR chains, do NOT start a chain with `foo != null` or `foo !== null`
			// on a BASE identifier. These patterns are semantically OPPOSITE of optional chaining.
			//
			// Example: foo != null || foo.bar
			// - If foo is NOT null (true), short-circuit to true - we never evaluate foo.bar
			// - If foo IS null (false), we try foo.bar which THROWS because foo is null!
			//
			// This is the OPPOSITE of what optional chaining does:
			// - foo?.bar: If foo is null/undefined, return undefined; otherwise access foo.bar
			//
			// Valid OR chain starting patterns:
			// - foo == null || foo.bar (if foo IS null, short-circuit; otherwise access foo.bar)
			// - foo === null || foo.bar (same)
			// - !foo || foo.bar (if foo is falsy, short-circuit; otherwise access foo.bar)
			//
			// Invalid OR chain starting patterns (should not be flagged):
			// - foo != null || foo.bar (wrong semantics - throws if foo is null!)
			// - foo !== null || foo.bar !== X (same issue)
			if op.typ == OperandTypeComparison && op.node != nil {
				unwrapped := unwrapParentheses(op.node)
				if ast.IsBinaryExpression(unwrapped) {
					binExpr := unwrapped.AsBinaryExpression()
					binOp := binExpr.OperatorToken.Kind
					// Check for != or !== operators
					if binOp == ast.KindExclamationEqualsToken || binOp == ast.KindExclamationEqualsEqualsToken {
						left := unwrapParentheses(binExpr.Left)
						right := unwrapParentheses(binExpr.Right)
						// Check if comparing to null/undefined
						isLeftNullish := left.Kind == ast.KindNullKeyword ||
							(ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
							ast.IsVoidExpression(left)
						isRightNullish := right.Kind == ast.KindNullKeyword ||
							(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
							ast.IsVoidExpression(right)
						if isLeftNullish || isRightNullish {
							// Determine which side is the checked expression
							var checkedExpr *ast.Node
							if isRightNullish {
								checkedExpr = left
							} else {
								checkedExpr = right
							}
							// If the checked expression is a base identifier, don't start chain
							isBaseIdentifier := ast.IsIdentifier(checkedExpr) || checkedExpr.Kind == ast.KindThisKeyword
							if isBaseIdentifier {
								// Skip this operand - cannot start a valid OR chain
								continue
							}
						}
					}
				}
			}
			chain = append(chain, op)
			lastExpr = op.comparedExpr
			// Set hasTrailingComparison for both value comparisons AND null checks
			// Null checks like foo.bar == null should be preserved in the output
			if isComparisonOrNullCheck(op.typ) {
				hasTrailingComparison = true
			}
			continue
		}

		// Check if this continues the chain
		cmp := processor.compareNodes(lastExpr, op.comparedExpr)

		// Special case for OR chains:
		// Allow extending call expressions even though they may have side effects when:
		// 1. The unsafe option is enabled, OR
		// 2. Both the previous and current operand are negations (OperandTypeNot)
		//
		// For case 2: !foo() || !foo().bar
		// - Both operands are negations checking for falsy values
		// - The user's intent is clear: chain through the call result
		// - This is a common pattern that typescript-eslint converts
		if cmp == NodeInvalid && len(chain) > 0 {
			prevOp := chain[len(chain)-1]
			isNegationChain := prevOp.typ == OperandTypeNot && op.typ == OperandTypeNot
			// Also allow extending through call expressions for nullish comparison chains
			// Pattern: foo.bar() === null || foo.bar().baz === null
			// Both operands are checking for null/undefined, so extending through the call is safe
			isNullishComparisonChain := isOrChainNullishCheck(prevOp) && isOrChainNullishCheck(op)

			if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing || isNegationChain || isNullishComparisonChain {
				// Check if lastExpr is a call/new expression and op.comparedExpr extends it
				lastUnwrapped := lastExpr
				if lastUnwrapped != nil {
					for ast.IsParenthesizedExpression(lastUnwrapped) {
						lastUnwrapped = lastUnwrapped.AsParenthesizedExpression().Expression
					}
					if ast.IsCallExpression(lastUnwrapped) || ast.IsNewExpression(lastUnwrapped) {
						// Try text-based comparison to see if op extends lastExpr
						lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastExpr)
						opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.comparedExpr)
						sourceText := processor.sourceText
						if lastRange.Pos() >= 0 && lastRange.End() <= len(sourceText) &&
							opRange.Pos() >= 0 && opRange.End() <= len(sourceText) {
							lastText := sourceText[lastRange.Pos():lastRange.End()]
							opText := sourceText[opRange.Pos():opRange.End()]
							if strings.HasPrefix(opText, lastText) {
								remainder := strings.TrimPrefix(opText, lastText)
								if len(remainder) > 0 && (remainder[0] == '.' || remainder[0] == '[' || remainder[0] == '(') {
									// op extends lastExpr, treat as NodeSubset
									cmp = NodeSubset
								}
							}
						}
					}
				}
			}
		}

		if cmp == NodeSubset || cmp == NodeEqual {
			// Special case: Don't add a value comparison (OperandTypeComparison) when it's
			// on the same expression as a previous negation/null check.
			// Pattern: !foo || !foo.bar || foo.bar > 5
			//   -> !foo?.bar || foo.bar > 5 (NOT foo?.bar > 5)
			// The comparison to a non-nullish value (> 5) is a different semantic check
			// and should NOT be part of the optional chain.
			//
			// EXCEPTION: If the comparison is actually a null/undefined check (e.g., a.b == null),
			// then it SHOULD extend the chain because it's checking for nullish values.
			// Pattern: !a || a.b == null || !a.b.c
			//   -> !a?.b?.c (the a.b == null is a nullish check that extends the chain)
			if cmp == NodeEqual && op.typ == OperandTypeComparison && len(chain) > 0 && !isNullishComparison(op) {
				lastOp := chain[len(chain)-1]
				// If the previous operand was a negation or null check, don't add this comparison
				if lastOp.typ == OperandTypeNot ||
					lastOp.typ == OperandTypeNotStrictEqualNull ||
					lastOp.typ == OperandTypeNotStrictEqualUndef ||
					lastOp.typ == OperandTypeNotEqualBoth ||
					lastOp.typ == OperandTypePlain {
					// Stop the chain here - the value comparison is semantically different
					if len(chain) >= 2 {
						break
					}
				}
			}

			// NodeSubset: foo vs foo.bar (growing chain)
			// NodeEqual: foo.bar vs foo.bar (duplicate check, continue chain)
			chain = append(chain, op)
			lastExpr = op.comparedExpr
			// Set hasTrailingComparison for both value comparisons AND null checks
			if isComparisonOrNullCheck(op.typ) {
				hasTrailingComparison = true
			}

			continue
		}

		// Chain broken
		if len(chain) >= 2 {
			break
		}
		chain = []Operand{op}
		lastExpr = op.comparedExpr
		hasTrailingComparison = (op.typ == OperandTypeComparison)
	}

	if len(chain) < 2 {
		return
	}

	// Check if all operands in the chain have the same base identifier
	// Example: a === undefined || b === null - different bases (a vs b), skip
	// Example: foo === null || foo.bar - same base (foo), allow
	if !processor.hasSameBaseIdentifier(chain) {
		return // Different base identifiers in the same chain
	}

	// Ensure at least one operand involves property/element/call access
	// Pattern to skip: foo === null || foo === undefined (just null checks, no access)
	// Pattern to allow: foo === null || foo.bar (has property access)
	if !processor.hasPropertyAccessInChain(chain) {
		return // No property access, nothing to chain
	}

	// Ensure at least one operand is an explicit check (not just plain truthy/falsy)
	// Pattern to skip: foo || foo.bar (plain truthy checks, different semantics)
	// Pattern to allow: foo == null || foo.bar (has explicit null check)
	// Pattern to allow: !foo || !foo.bar (has negation)
	hasExplicitCheck := false
	for _, op := range chain {
		if op.typ != OperandTypePlain {
			hasExplicitCheck = true
			break
		}
	}
	if !hasExplicitCheck {
		return // No explicit checks, just plain operands - can't convert
	}

	// Skip OR chains where ALL operands use STRICT nullish checks AND subsequent operands
	// already have optimal optional chaining.
	// Same logic as for AND chains - strict checks are intentional when combined with ?.
	if len(chain) >= 2 {
		// Check if ALL operands after the first have optional chaining
		allSubsequentHaveOptionalChaining := true
		for i := 1; i < len(chain); i++ {
			if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
				allSubsequentHaveOptionalChaining = false
				break
			}
		}

		// Early check: If all subsequent operands already have optimal optional chaining,
		// and the first operand is an explicit NULL check (not undefined) on a non-nullable type, skip.
		// Example: foo.bar === null || foo.bar?.() === null || foo.bar?.().baz
		// where foo.bar is a function type (not nullable) - the code is already optimal.
		//
		// BUT: Don't skip for === undefined checks! These should be flagged for conversion.
		// This matches typescript-eslint behavior where === null on non-nullable with ?. is valid,
		// but === undefined on non-nullable with ?. should be converted to optional chaining only.
		//
		// Also don't skip if the first operand is a negation (!foo) or plain truthiness check,
		// as those can still be converted to optional chaining regardless of nullability.
		if allSubsequentHaveOptionalChaining {
			firstOp := chain[0]
			// Only check type when first operand is an explicit NULL equality check (not undefined)
			// For OR chains, property null comparisons are classified as OperandTypeComparison,
			// so we need to check the actual comparison too
			isExplicitNullCheck := firstOp.typ == OperandTypeStrictEqualNull ||
				firstOp.typ == OperandTypeEqualNull ||
				isStrictNullComparison(firstOp)
			if isExplicitNullCheck {
				// Check if ANY type in the chain is nullable
				// We should only skip if ALL types in the chain are non-nullable
				// Example: foo.bar === null || foo.bar?.() === null
				// foo.bar is a function (not nullable), but foo.bar() is nullable
				// We should still report because there are nullable types that can be consolidated
				anyTypeHasNullOrUndefined := false
				for _, op := range chain {
					if op.comparedExpr != nil {
						typeInfo := processor.getTypeInfo(op.comparedExpr)
						if typeInfo.hasNull || typeInfo.hasUndefined {
							anyTypeHasNullOrUndefined = true
							break
						}
					}
				}
				if !anyTypeHasNullOrUndefined {
					return // No type in the chain is nullable and subsequent operands already use ?., skip
				}
			}
		}

		// Check if ALL nullish checks in the chain are STRICT (not loose)
		allStrictChecks := true
		for _, op := range chain {
			// Loose checks that cover both null and undefined
			if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
				allStrictChecks = false
				break
			}
			// Plain/Not operands also check for all falsy values, not just one nullish
			if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
				allStrictChecks = false
				break
			}
			// For OperandTypeComparison, check if it's a loose nullish comparison
			if op.typ == OperandTypeComparison && op.node != nil {
				if ast.IsBinaryExpression(op.node) {
					binExpr := op.node.AsBinaryExpression()
					binOp := binExpr.OperatorToken.Kind
					// == null/undefined is loose (covers both)
					if binOp == ast.KindEqualsEqualsToken {
						left := unwrapParentheses(binExpr.Left)
						right := unwrapParentheses(binExpr.Right)
						isNullish := left.Kind == ast.KindNullKeyword || right.Kind == ast.KindNullKeyword ||
							(ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
							(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined")
						if isNullish {
							allStrictChecks = false
							break
						}
					}
				}
			}
		}

		// Only skip if BOTH conditions are met:
		// 1. All subsequent operands have optional chaining
		// 2. All checks are strict (so the pattern is intentional)
		// EXCEPTION: If the type only includes null OR only includes undefined (but not both),
		// then strict checks are appropriate for the type and we should still report.
		// Example: type is `| null` only - using === null is just being type-correct, not
		// intentionally distinguishing between null and undefined.
		if allSubsequentHaveOptionalChaining && allStrictChecks {
			// Check types across ALL operands in the chain, not just the first
			// Example: foo.bar === null || foo.bar?.() === null
			// foo.bar is a function (not nullable), but foo.bar() is nullable
			anyHasNull := false
			anyHasUndefined := false
			for _, op := range chain {
				if op.comparedExpr != nil {
					typeInfo := processor.getTypeInfo(op.comparedExpr)
					if typeInfo.hasNull {
						anyHasNull = true
					}
					if typeInfo.hasUndefined {
						anyHasUndefined = true
					}
				}
			}

			// Skip if NO type in the chain has null or undefined
			// In this case, all checks are meaningless (dead code)
			if !anyHasNull && !anyHasUndefined {
				return // No types are nullable, skip it
			}
			// Skip if types include BOTH null AND undefined across the chain
			// In that case, strict checks are intentionally distinguishing between them
			if anyHasNull && anyHasUndefined {
				return // Chain uses strict checks with optimal optional chaining, skip it
			}
			// If types only include null OR only undefined (but not both), continue processing
			// as the strict check is just type-appropriate, not distinguishing
		}
	}

	// Note: OR chains with trailing comparisons do NOT require the unsafe option
	// The semantics are preserved because:
	// - Original: !foo || foo.bar != 0 returns true if foo is falsy, otherwise (foo.bar != 0)
	// - Converted: foo?.bar != 0 returns (undefined != 0) = true if foo is null/undefined
	// For other falsy values (0, "", false), the result is also preserved due to JS semantics
	// So we allow OR chains with comparisons without requiring the unsafe option

	// Skip chains that start with simple negation (!foo || foo.bar)
	// These check ALL falsy values (0, "", false, null, undefined),
	// while optional chaining only checks null/undefined
	// Converting would change semantics
	// HOWEVER, allow chains where ALL operands are consistently negated (!foo || !foo.bar)
	// ALSO, allow all chains when the unsafe option is enabled
	if len(chain) >= 2 && chain[0].typ == OperandTypeNot && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		// Check if it's a simple negation (not a negation with property access)
		// Pattern to skip: !x || x.prop (mixed - first is negation, second is not)
		// Pattern to allow: !x.prop || x.prop.bar (negation of property access)
		// Pattern to allow: !x || !x.prop (all negated consistently)
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := unwrapParentheses(firstExpr)
			// If the first negated expression is NOT a property/element/call access
			isFirstSimpleNegation := !ast.IsPropertyAccessExpression(unwrappedFirst) &&
				!ast.IsElementAccessExpression(unwrappedFirst) &&
				!ast.IsCallExpression(unwrappedFirst)

			if isFirstSimpleNegation {
				// Check if ALL subsequent operands are also negations OR safe comparisons OR null checks
				// Allow patterns like:
				// - !foo || !foo.bar (all negated)
				// - !foo || foo.bar != 0 (negation + SAFE comparison with literal)
				// - !foo || foo.bar === undefined (negation + null check) - SAFE
				// - !foo || foo.bar == null (negation + null check) - SAFE
				// - !a || a.b === null || !a.b.c (nullish comparison as INTERMEDIATE check - SAFE)
				// - !a || a.b == null || a.b.c (nullish comparisons + plain at end - SAFE)
				// Block patterns like:
				// - !foo || foo.bar (negation + plain property access without intermediate checks)
				// - !foo || foo.bar === 'foo' (strict equality with non-undefined - NOT safe)
				// - !foo || foo.bar !== 'foo' (strict not-equal - NOT safe)
				// - !foo || foo.bar != null (loose not-equal with null/undefined - NOT safe)
				// - !foo || foo.bar === null (2-operand chain ending in === null - NOT safe, changes semantics for falsy non-nullish)
				allNegatedOrSafeComparisonOrNullCheck := true
				hasIntermediateNullishComp := false // Track if we have intermediate nullish checks
				for i := 1; i < len(chain); i++ {
					isComparison := chain[i].typ == OperandTypeComparison
					isSafeComparison := isComparison && processor.isOrChainComparisonSafe(chain[i])
					// Allow nullish comparisons (== null, === null, === undefined) ONLY as intermediate checks
					// i.e., when this is NOT the last operand in the chain
					// For 2-operand chains like !foo || foo.bar === null, the conversion changes semantics
					// for falsy non-nullish values (0, "", false), so we don't allow it
					isIntermediateNullishComp := isComparison && isNullishComparison(chain[i]) && i < len(chain)-1
					if isIntermediateNullishComp {
						hasIntermediateNullishComp = true
					}

					// Allow OperandTypePlain at the END of the chain if there were intermediate nullish checks
					// Pattern: !a || a.b == null || a.b.c.d.e.f.g.h (guarded by nullish checks)
					isAllowedPlainAtEnd := chain[i].typ == OperandTypePlain && i == len(chain)-1 && hasIntermediateNullishComp

					if chain[i].typ != OperandTypeNot && !isSafeComparison && !isIntermediateNullishComp && !isNullishCheckType(chain[i].typ) && !isAllowedPlainAtEnd {
						allNegatedOrSafeComparisonOrNullCheck = false
						break
					}
				}
				// If mixed with plain property access or unsafe comparisons, skip
				if !allNegatedOrSafeComparisonOrNullCheck {
					return
				}
			} else {
				// First operand is negation of property/element/call access (!foo.bar || ...)
				// Still need to check if trailing comparisons are safe
				// Pattern: !array[1] || array[1].b === 'foo' - NOT safe to convert
				// Pattern: !array[1] || array[1].b != 0 - SAFE to convert
				for i := 1; i < len(chain); i++ {
					if chain[i].typ == OperandTypeComparison && !processor.isOrChainComparisonSafe(chain[i]) {
						return // Unsafe comparison, skip this chain
					}
				}
			}
		}
	}

	// Also check for OR chains starting with foo == null || foo.bar OP value
	// These need to verify the trailing comparison is safe too
	// Note: In OR chains, foo == null is treated as OperandTypeNotEqualBoth (covers both null and undefined)
	// Skip this check when the unsafe option is enabled (allow potentially unsafe transformations)
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		if len(chain) >= 2 && (chain[0].typ == OperandTypeNotEqualBoth || chain[0].typ == OperandTypeNotStrictEqualNull || chain[0].typ == OperandTypeNotStrictEqualUndef) {
			// Get type info for the first operand to determine what values it can be
			firstTypeInfo := processor.getTypeInfo(chain[0].comparedExpr)

			// Check if trailing comparisons are safe
			for i := 1; i < len(chain); i++ {
				if chain[i].typ == OperandTypeComparison && !processor.isOrChainComparisonSafe(chain[i]) {
					isLastOperand := i == len(chain)-1

					if isLastOperand && isNullishComparison(chain[i]) {
						// Trailing nullish comparison - only unsafe if the first operand's type
						// includes BOTH null AND undefined (meaning == null catches values that
						// === null doesn't, changing semantics when converted)
						//
						// Examples:
						// - foo == null || foo.bar === null (foo: any) -> UNSAFE
						//   When foo is undefined, original returns true, converted returns false
						// - foo === null || foo.bar === null (foo: T | null) -> SAFE
						//   foo can never be undefined, so semantics are preserved
						if firstTypeInfo.hasNull && firstTypeInfo.hasUndefined {
							return // Type includes both null and undefined, trailing === null is unsafe
						}
						if firstTypeInfo.hasAny || firstTypeInfo.hasUnknown {
							return // any/unknown could be anything, be conservative
						}
						// Type only includes one of null/undefined (or neither), so trailing comparison is safe
					} else if !isNullishComparison(chain[i]) {
						// Non-nullish comparison that's not safe - skip chain
						return
					}
					// Intermediate nullish comparison is safe (it's a guard)
				}
			}
		}
	}

	// When requireNullish is true, skip chains that start with negation (!foo || !foo.bar)
	// Only allow chains that start with explicit null checks (foo == null || foo.bar)
	if processor.shouldSkipForRequireNullish(chain, false) {
		return
	}

	// Skip OR chains starting with import.meta (import.meta || import.meta.url)
	// import.meta is always defined (non-nullable), so the second part is unreachable
	// This is similar to skipping 'this' patterns
	if len(chain) >= 2 && chain[0].typ == OperandTypePlain {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrapped := unwrapParentheses(firstExpr)
			if unwrapped.Kind == ast.KindMetaProperty {
				return
			}
		}
	}

	// Skip OR chains starting with plain truthy check (foo || foo.bar != 0)
	// These check ALL falsy values (0, "", false, null, undefined),
	// while optional chaining only checks null/undefined
	// Original: foo || foo.bar != 0 -> returns foo if falsy, otherwise foo.bar != 0
	// Converted: foo?.bar != 0 -> returns undefined != 0 = true if foo is null/undefined
	// This changes semantics, so skip unless the unsafe option is enabled
	// HOWEVER, allow plain chains without comparisons (foo || foo.bar) as these are valid
	if len(chain) >= 2 && chain[0].typ == OperandTypePlain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := unwrapParentheses(firstExpr)
			// If the first plain expression is NOT a property/element/call access (i.e., it's just a simple identifier)
			isFirstSimplePlain := !ast.IsPropertyAccessExpression(unwrappedFirst) &&
				!ast.IsElementAccessExpression(unwrappedFirst) &&
				!ast.IsCallExpression(unwrappedFirst)

			if isFirstSimplePlain {
				// Check if any subsequent operand is a comparison
				// Block patterns like: foo || foo.bar != 0 (truthy + comparison)
				// Allow patterns like: foo || foo.bar (truthy + plain property access)
				for i := 1; i < len(chain); i++ {
					if chain[i].typ == OperandTypeComparison {
						return
					}
				}
			}
		}
	}

	// CRITICAL: Check for incomplete nullish checks in OR chains
	// Optional chaining checks for BOTH null AND undefined
	// If the chain only checks for null OR only for undefined (not both), it's NOT equivalent
	// Example: x === undefined || x.prop - if x is null, evaluates x.prop (throws!)
	//          x?.prop - if x is null, returns undefined safely
	// HOWEVER: Allow incomplete checks when the unsafe option is enabled
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		hasNullCheck := false
		hasUndefinedCheck := false
		hasBothCheck := false
		// Debug mode for invalid-225
		for _, op := range chain {
			if op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeStrictEqualNull {
				hasNullCheck = true
			} else if op.typ == OperandTypeNotStrictEqualUndef || op.typ == OperandTypeStrictEqualUndef {
				hasUndefinedCheck = true
			} else if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
				hasBothCheck = true
			} else if op.typ == OperandTypeTypeofCheck {
				// typeof checks are equivalent to undefined checks
				// (typeof x === 'undefined' in OR chains means undefined check)
				hasUndefinedCheck = true
			} else if op.typ == OperandTypeNot {
				// !foo checks all falsy values including null and undefined
				hasBothCheck = true
			} else if op.typ == OperandTypeComparison && op.node != nil {
				// Check if this is a nullish comparison (== null, === null, === undefined)
				unwrapped := op.node
				for ast.IsParenthesizedExpression(unwrapped) {
					unwrapped = unwrapped.AsParenthesizedExpression().Expression
				}
				if ast.IsBinaryExpression(unwrapped) {
					binExpr := unwrapped.AsBinaryExpression()
					binOp := binExpr.OperatorToken.Kind
					left := binExpr.Left
					right := binExpr.Right
					for ast.IsParenthesizedExpression(left) {
						left = left.AsParenthesizedExpression().Expression
					}
					for ast.IsParenthesizedExpression(right) {
						right = right.AsParenthesizedExpression().Expression
					}

					isNull := left.Kind == ast.KindNullKeyword || right.Kind == ast.KindNullKeyword
					isUndefined := (ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
						(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
						ast.IsVoidExpression(left) || ast.IsVoidExpression(right)

					if binOp == ast.KindEqualsEqualsToken && (isNull || isUndefined) {
						// == null covers both null and undefined
						hasBothCheck = true
					} else if binOp == ast.KindEqualsEqualsEqualsToken {
						if isNull {
							hasNullCheck = true
						}
						if isUndefined {
							hasUndefinedCheck = true
						}
					}
				}
			}
		}

		// If we have a strict null check or strict undefined check (but not both), skip
		// This is unsafe regardless of other checks in the chain
		// Note: typeof checks count as undefined checks
		// EXCEPTION: If the first operand's expression type doesn't include nullish, skip this check
		// The incomplete nullish check is only dangerous when the type COULD be null but we only check undefined
		// If the type can never be null, then checking only for undefined is fine
		// ALSO: If the trailing operand already has optional chaining, allow the conversion
		firstOpNotNullish := false
		hasTrailingOptionalChaining := false
		if len(chain) > 0 && chain[0].comparedExpr != nil {
			if !processor.includesNullish(chain[0].comparedExpr) {
				firstOpNotNullish = true
			}
		}
		if len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			if lastOp.comparedExpr != nil && processor.containsOptionalChain(lastOp.comparedExpr) {
				hasTrailingOptionalChaining = true
			}
		}

		// Check if the strict check is COMPLETE for the types in the chain.
		// A strict check is complete when:
		// - `=== null` (for OR) and ALL nullable types have null but NOT undefined
		// - `=== undefined` (for OR) and ALL nullable types have undefined but NOT null
		// In these cases, the check is semantically complete and we should report.
		//
		// We check all operands because the first operand might not be nullable (e.g., function type),
		// but subsequent operands (return values) might be.
		strictCheckIsComplete := true // Assume complete, set false if any operand has both
		hasAnyNullableOperand := false
		for _, op := range chain {
			if op.comparedExpr == nil {
				continue
			}
			info := processor.getTypeInfo(op.comparedExpr)
			// Skip non-nullable operands (like function types) - they don't affect completeness
			if !info.hasNull && !info.hasUndefined && !info.hasAny && !info.hasUnknown {
				continue
			}
			hasAnyNullableOperand = true
			// If any nullable operand has any/unknown, we can't determine completeness
			if info.hasAny || info.hasUnknown {
				strictCheckIsComplete = false
				break
			}
			// If any nullable operand has BOTH null AND undefined, the check is incomplete
			if info.hasNull && info.hasUndefined {
				strictCheckIsComplete = false
				break
			}
			// Check if the check type matches the type's nullability
			hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
			hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck
			// === null is incomplete if type only has undefined (not null)
			if hasOnlyNullCheck && !info.hasNull && info.hasUndefined {
				strictCheckIsComplete = false
				break
			}
			// === undefined is incomplete if type only has null (not undefined)
			if hasOnlyUndefinedCheck && info.hasNull && !info.hasUndefined {
				strictCheckIsComplete = false
				break
			}
		}
		// If no nullable operands found, can't be complete
		if !hasAnyNullableOperand {
			strictCheckIsComplete = false
		}

		if !hasBothCheck && !firstOpNotNullish && !hasTrailingOptionalChaining && !strictCheckIsComplete {
			hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
			hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck

			if hasOnlyNullCheck || hasOnlyUndefinedCheck {
				return // Skip chains with incomplete nullish checks
			}
		}

		// CRITICAL: Also check for OperandTypeComparison operands that are strict equality checks
		// against null or undefined on property accesses. If the type includes BOTH null AND undefined
		// but the check only covers one, the conversion would be unsafe.
		// Example: foo.bar === undefined || foo.bar.baz where foo.bar has type T | null | undefined
		//          This is unsafe because if foo.bar is null, the original throws but foo.bar?.baz doesn't
		// IMPORTANT: Only check guard operands (not the last one), because the last operand's check
		// is preserved in the output.
		// ALSO: Skip operands that already have optional chaining - they're intermediate operands
		// whose checks will be preserved in the output.
		//
		// MODIFICATION: Instead of rejecting the entire chain, truncate it so that the operand
		// with the incomplete check becomes the last operand (whose check will be preserved).
		truncateAt := -1 // -1 means no truncation needed
		for i, op := range chain {
			// Skip the last operand - its check is preserved in the output
			if i == len(chain)-1 {
				continue
			}
			// Skip operands that already have optional chaining - they're preserved
			if op.comparedExpr != nil && processor.containsOptionalChain(op.comparedExpr) {
				continue
			}

			// Check for incomplete nullish checks on guard operands.
			// This covers:
			// 1. OperandTypeNotStrictEqualNull in OR chains (foo === null parsed as equivalent to !== null in AND)
			// 2. OperandTypeNotStrictEqualUndef in OR chains (foo === undefined parsed as equivalent to !== undefined in AND)
			// 3. OperandTypeComparison with === null or === undefined
			//
			// If the type has BOTH null AND undefined but the check only covers one,
			// truncate the chain so this operand becomes the last (preserving its check).
			if op.comparedExpr != nil {
				typeInfo := processor.getTypeInfo(op.comparedExpr)
				if typeInfo.hasNull && typeInfo.hasUndefined {
					// Type has both null and undefined
					// In OR chains, OperandTypeNotStrictEqualNull represents "foo === null"
					// (semantically equivalent to !== null in AND chains)
					switch op.typ {
					case OperandTypeNotStrictEqualNull:
						// foo === null in OR chain on type that also includes undefined - incomplete check
						// Truncate chain so this operand becomes the last (preserving its check)
						if truncateAt == -1 || i+1 < truncateAt {
							truncateAt = i + 1
						}
					case OperandTypeNotStrictEqualUndef:
						// foo === undefined in OR chain on type that also includes null - incomplete check
						if truncateAt == -1 || i+1 < truncateAt {
							truncateAt = i + 1
						}
					case OperandTypeStrictEqualNull:
						// foo === null (for inverted AND chain checks) on type with both - incomplete
						if truncateAt == -1 || i+1 < truncateAt {
							truncateAt = i + 1
						}
					case OperandTypeStrictEqualUndef:
						// foo === undefined (for inverted AND chain checks) on type with both - incomplete
						if truncateAt == -1 || i+1 < truncateAt {
							truncateAt = i + 1
						}
					case OperandTypeComparison:
						if op.node != nil {
							unwrapped := op.node
							for ast.IsParenthesizedExpression(unwrapped) {
								unwrapped = unwrapped.AsParenthesizedExpression().Expression
							}
							if ast.IsBinaryExpression(unwrapped) {
								binExpr := unwrapped.AsBinaryExpression()
								operator := binExpr.OperatorToken.Kind

								// Only check strict equality operators (=== null or === undefined)
								if operator == ast.KindEqualsEqualsEqualsToken {
									// Check if comparing to null or undefined
									isStrictNullCheck := binExpr.Right.Kind == ast.KindNullKeyword ||
										binExpr.Left.Kind == ast.KindNullKeyword
									isStrictUndefCheck := (ast.IsIdentifier(binExpr.Right) && binExpr.Right.AsIdentifier().Text == "undefined") ||
										(ast.IsIdentifier(binExpr.Left) && binExpr.Left.AsIdentifier().Text == "undefined") ||
										ast.IsVoidExpression(binExpr.Right) || ast.IsVoidExpression(binExpr.Left)

									// If type has both null and undefined, but we only check one, truncate
									if isStrictNullCheck && !isStrictUndefCheck {
										if truncateAt == -1 || i+1 < truncateAt {
											truncateAt = i + 1
										}
									}
									if isStrictUndefCheck && !isStrictNullCheck {
										if truncateAt == -1 || i+1 < truncateAt {
											truncateAt = i + 1
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// Apply truncation if needed
		if truncateAt > 0 && truncateAt < len(chain) {
			chain = chain[:truncateAt]
		}

		// After truncation, verify we still have a valid chain
		if len(chain) < 2 {
			return
		}

		// CRITICAL: For === null OR chains, converting to ?. changes semantics.
		//
		// Optional chaining (?.) returns UNDEFINED when the value is nullish:
		//   x?.y returns undefined if x is null OR undefined
		//
		// This means:
		//   - x === null || x.y === null  →  returns true if x is null
		//   - x?.y === null               →  returns false if x is null (undefined === null is false)
		//
		// For === undefined, the semantics ARE preserved:
		//   - x === undefined || x.y === undefined  →  returns true if x is undefined
		//   - x?.y === undefined                    →  returns true if x is undefined (undefined === undefined is true)
		//
		// So we skip === null chains unless the type only includes null (no undefined),
		// in which case the strict check is semantically complete for that type.
		if hasNullCheck && !hasUndefinedCheck && !hasBothCheck && !strictCheckIsComplete {
			return
		}
	}

	// Check if conversion would change return type for plain operands in OR chains
	// Skip unless the unsafe option is enabled
	for _, op := range chain {
		if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
			if processor.wouldChangeReturnType(op.comparedExpr) && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
				return
			}
		}
	}

	// Skip pattern !a.b || a.b() where we negate a property then call it
	// This checks if a function exists before calling it, which is a valid pattern
	// Converting would change semantics: !a.b || a.b() !== !a.b?.()
	// HOWEVER, allow !a.b || !a.b() (both negated) - this CAN be converted to !a.b?.()
	if len(chain) >= 2 {
		for i := range len(chain) - 1 {
			if chain[i].typ == OperandTypeNot {
				// Check if any subsequent operand is a call to the same base
				// BUT only if that subsequent operand is NOT also negated
				negatedExpr := chain[i].comparedExpr
				for j := i + 1; j < len(chain); j++ {
					// Skip if this operand is also negated - that's allowed
					if chain[j].typ == OperandTypeNot {
						continue
					}
					callExpr := chain[j].comparedExpr
					if callExpr != nil && ast.IsCallExpression(unwrapParentheses(callExpr)) {
						// Check if the call is to the negated expression
						call := unwrapParentheses(callExpr).AsCallExpression()
						callBase := call.Expression
						// Compare the negated expression with the call base
						cmp := processor.compareNodes(negatedExpr, callBase)
						if cmp == NodeEqual {
							// Negation followed by call to same property - skip
							return
						}
					}
				}
			}
		}
	}

	// Note: For OR chains, we don't need the same null check skip logic as AND chains
	// because the semantics are equivalent:
	//   Original: !data || data.value !== null -> true if data is null
	//   Converted: data?.value !== null -> true if data is null (undefined !== null)
	// The AND chain case is handled separately in the && handler

	// Helper to check if an operand is a null/undefined check (including OperandTypeComparison with null/undef)
	isNullishCheck := func(op Operand) bool {
		if op.typ == OperandTypeNotStrictEqualNull ||
			op.typ == OperandTypeNotStrictEqualUndef ||
			op.typ == OperandTypeNotEqualBoth ||
			op.typ == OperandTypeStrictEqualNull ||
			op.typ == OperandTypeStrictEqualUndef ||
			op.typ == OperandTypeEqualNull {
			return true
		}
		// OperandTypeComparison can also be a null/undefined check in OR chains
		// e.g., a.b == null is classified as OperandTypeComparison for property accesses
		if op.typ == OperandTypeComparison && op.node != nil && ast.IsBinaryExpression(op.node) {
			binExpr := op.node.AsBinaryExpression()
			// Check if right side is null or undefined
			rightIsNull := binExpr.Right.Kind == ast.KindNullKeyword
			rightIsUndefined := ast.IsIdentifier(binExpr.Right) &&
				binExpr.Right.AsIdentifier().Text == "undefined"
			// Check if left side is null or undefined (Yoda style)
			leftIsNull := binExpr.Left.Kind == ast.KindNullKeyword
			leftIsUndefined := ast.IsIdentifier(binExpr.Left) &&
				binExpr.Left.AsIdentifier().Text == "undefined"
			return rightIsNull || rightIsUndefined || leftIsNull || leftIsUndefined
		}
		return false
	}

	// Special case: OR chain with trailing PLAIN operand after null checks
	// When the last operand is OperandTypePlain (not negated), it should be kept SEPARATE
	// because the conversion is only partial - we convert the guarded part, not the final access.
	//
	// Pattern: !a || a.b == null || ... || a.b.c.d.e.f.g == null || a.b.c.d.e.f.g.h
	// Expected: a?.b?.c?.d?.e?.f?.g == null || a.b.c.d.e.f.g.h (keep plain separate)
	//
	// BUT when the last operand is negated (!a.b.c.d.e.f.g.h), fully convert:
	// Pattern: !a || a.b == null || ... || !a.b.c.d.e.f.g.h
	// Expected: !a?.b?.c?.d?.e?.f?.g?.h (full conversion)
	//
	// For simple 2-operand chains like: foo == null || foo.bar
	// Expected: foo?.bar (fully converted)
	//
	// NOTE: When unsafe option is enabled, we allow full conversion even with trailing plain
	trailingPlainOperand := ""
	chainForOptional := chain
	if len(chain) >= 3 && chain[len(chain)-1].typ == OperandTypePlain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		lastOp := chain[len(chain)-1]
		secondLastOp := chain[len(chain)-2]
		// Check if second-to-last is a null check
		if isNullishCheck(secondLastOp) && lastOp.comparedExpr != nil && secondLastOp.comparedExpr != nil {
			// Check if plain operand extends the null check
			lastParts := processor.flattenForFix(lastOp.comparedExpr)
			secondLastParts := processor.flattenForFix(secondLastOp.comparedExpr)
			if len(lastParts) > len(secondLastParts) {
				// Plain extends null check - keep plain operand separate
				lastOpRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOp.node)
				trailingPlainOperand = processor.sourceText[lastOpRange.Pos():lastOpRange.End()]
				chainForOptional = chain[:len(chain)-1]
			}
		}
	}

	// After separating the trailing plain operand, if the remaining chain is just 1 operand
	// AND it already contains optional chaining, there's nothing more to transform
	// This prevents infinite fix loops on patterns like: a?.b?.c == null || a.b.c.d
	if len(chainForOptional) == 1 && trailingPlainOperand != "" {
		// Check if the single remaining operand already has optional chaining
		singleOp := chainForOptional[0]
		if singleOp.comparedExpr != nil {
			if hasOptionalChaining(singleOp.comparedExpr) {
				return
			}
		}
	}

	// Also check for 2-operand chains where the first operand already has optional chaining
	// This prevents the second fix pass on: a?.b?.c == null || a.b.c.d
	// which would try to convert to a.b.c?.d
	if len(chain) == 2 && trailingPlainOperand == "" {
		firstOp := chain[0]
		// Check both node and comparedExpr since for comparison operands,
		// comparedExpr contains the actual expression being checked
		if firstOp.comparedExpr != nil && hasOptionalChaining(firstOp.comparedExpr) {
			return
		}
		if firstOp.node != nil && hasOptionalChaining(firstOp.node) {
			return
		}
	}

	// Build the optional chain with negation (or comparison)
	// For OR chains, the last operand is typically a plain expression or comparison
	// Use node for plain (preserves NonNull), comparedExpr for comparisons
	lastOp := chainForOptional[len(chainForOptional)-1]
	var lastPropertyAccess *ast.Node
	if lastOp.typ == OperandTypePlain {
		lastPropertyAccess = lastOp.node
	} else {
		lastPropertyAccess = lastOp.comparedExpr
	}
	parts := processor.flattenForFix(lastPropertyAccess)

	// Find all checked lengths to determine which properties should be optional
	checkedLengths := make(map[int]bool)

	// Find all checks (not including the last plain operand if any)
	checksToConsider := chainForOptional
	if len(chainForOptional) > 0 && chainForOptional[len(chainForOptional)-1].typ == OperandTypePlain {
		checksToConsider = chainForOptional[:len(chainForOptional)-1]
	}

	for _, operand := range checksToConsider {
		if operand.comparedExpr != nil {
			checkedParts := processor.flattenForFix(operand.comparedExpr)
			checkedLengths[len(checkedParts)] = true
		}
	}

	// Check if we're explicitly checking the function being called
	// e.g., !foo || !foo.bar || !foo.bar.baz || !foo.bar.baz()
	// We check foo.bar.baz before calling it, so the call should be optional
	callShouldBeOptional := false
	if len(parts) > 0 && strings.HasPrefix(parts[len(parts)-1].text, "(") {
		// Last part is a call expression
		// Check if we have a check for the expression without the call
		partsWithoutCall := len(parts) - 1
		for _, op := range chainForOptional[:len(chainForOptional)-1] { // Don't check the last operand (the call itself)
			checkedParts := processor.flattenForFix(op.node)
			// If we checked all parts except the call, the call should be optional
			if len(checkedParts) == partsWithoutCall {
				callShouldBeOptional = true
				break
			}
		}
	}

	optionalChainCode := processor.buildOptionalChain(parts, checkedLengths, callShouldBeOptional, true) // true = strip ! assertions for OR chains

	// If buildOptionalChain returned empty string, it means we'd create invalid syntax
	// (e.g., ?.#privateIdentifier which TypeScript doesn't allow)
	if optionalChainCode == "" {
		return
	}

	var newCode string
	// Update hasTrailingComparison based on chainForOptional (after removing trailing plain)
	hasTrailingComparisonForFix := false
	if len(chainForOptional) > 0 {
		lastOpForFix := chainForOptional[len(chainForOptional)-1]
		hasTrailingComparisonForFix = isComparisonOrNullCheck(lastOpForFix.typ)
	}

	if hasTrailingComparisonForFix {
		// Extract the comparison operator and right side from the last operand
		lastOpForFix := chainForOptional[len(chainForOptional)-1]
		if ast.IsBinaryExpression(lastOpForFix.node) {
			binExpr := lastOpForFix.node.AsBinaryExpression()
			// Check for Yoda style: undefined === foo.bar.baz vs foo.bar.baz === undefined
			// In Yoda: 'undefined' === foo.bar.baz
			// Not Yoda: foo.bar.baz === 'undefined'
			// Note: We normalize Yoda to non-Yoda style to match typescript-eslint behavior
			isYoda := false
			comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOpForFix.comparedExpr)
			leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)

			// If comparedExpr is on the right side, it's Yoda
			if comparedExprRange.Pos() > leftRange.Pos() {
				isYoda = true
			}

			if isYoda {
				// Yoda: normalize to non-Yoda style (optionalChain OP value)
				// Extract operator text (trim trivia to avoid extra spaces)
				opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.OperatorToken)
				opText := processor.sourceText[opRange.Pos():opRange.End()]
				// Extract left side (the value being compared)
				valueText := strings.TrimSpace(processor.sourceText[leftRange.Pos():leftRange.End()])
				newCode = optionalChainCode + " " + opText + " " + valueText
			} else {
				// Normal: append the operator + right side
				opStart := binExpr.OperatorToken.Pos()
				rightEnd := binExpr.Right.End()
				comparisonText := processor.sourceText[opStart:rightEnd]
				newCode = optionalChainCode + comparisonText
			}
		} else {
			newCode = optionalChainCode
		}
	} else {
		// Determine if we should add negation based on the chain pattern:
		// 1. If ALL operands are negated: !foo || !foo.bar -> !foo?.bar
		// 2. If first AND last operands are negated: !a || ... || !a.b.c -> !a?.b?.c
		// 3. Otherwise no negation: foo || foo.bar -> foo?.bar
		//                           foo == null || foo.bar -> foo?.bar
		//                           !a || a.b == null || a.b.c -> a?.b?.c (plain last)
		firstOpIsNegated := chainForOptional[0].typ == OperandTypeNot
		lastOpIsNegated := chainForOptional[len(chainForOptional)-1].typ == OperandTypeNot

		// Add negation if both first and last operands are negated
		// This handles: !foo || !foo.bar -> !foo?.bar
		// And: !a || a.b == null || !a.b.c -> !a?.b?.c
		if firstOpIsNegated && lastOpIsNegated {
			newCode = "!" + optionalChainCode
		} else {
			newCode = optionalChainCode
		}
	}

	// Append trailing plain operand if we kept one separate
	if trailingPlainOperand != "" {
		// Extract the original separator between the last two operands
		// to preserve formatting (spaces vs newlines)
		lastChainOp := chain[len(chain)-2] // Second to last (last in chainForOptional before trailing)
		trailingOp := chain[len(chain)-1]  // The trailing plain operand
		lastChainEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastChainOp.node).End()
		trailingStart := utils.TrimNodeTextRange(processor.ctx.SourceFile, trailingOp.node).Pos()
		// Extract text between them (includes ||)
		separator := processor.sourceText[lastChainEnd:trailingStart]
		// Clean up the separator: should be " || " or " ||\n..."
		// Just use the original separator as-is
		newCode = newCode + separator + trailingPlainOperand
	}

	// Use trimmed ranges to preserve leading/trailing whitespace
	// If we're replacing the entire logical expression, use the node's range
	var replaceStart, replaceEnd int
	if len(chain) == len(operandNodes) {
		// We're replacing all operands - use the top-level node range
		nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
		replaceStart = nodeRange.Pos()
		replaceEnd = nodeRange.End()
	} else {
		// We're replacing a subset - use operand ranges
		firstNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[0].node)
		lastNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-1].node)
		replaceStart = firstNodeRange.Pos()
		replaceEnd = lastNodeRange.End()
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(replaceStart, replaceEnd), newCode),
	}

	// Determine if we should autofix or suggest
	// Following typescript-eslint's logic for OR chains:
	// 1. If allowPotentiallyUnsafe option is enabled, always autofix
	// 2. If the last operand is !foo (NotBoolean) or has a nullish comparison type, autofix
	// 3. If there's a trailing comparison, autofix
	// 4. Otherwise, check if ANY operand includes undefined (including any/unknown) - if so, autofix
	// 5. Default to suggestion
	useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

	if useSuggestion && len(chain) > 0 {
		lastOp := chain[len(chain)-1]

		// Check if the last operand has a comparison type that makes autofix safe
		switch lastOp.typ {
		case OperandTypeNot: // !foo - for OR chains, this is safe
			useSuggestion = false
		case OperandTypeEqualNull, // == null (covers both null and undefined)
			OperandTypeNotEqualBoth,        // != null (covers both)
			OperandTypeStrictEqualUndef,    // === undefined
			OperandTypeNotStrictEqualUndef: // !== undefined
			useSuggestion = false
		case OperandTypeTypeofCheck:
			useSuggestion = false
		}
	}

	// For OR chains with trailing comparisons, check if direct fix is safe
	// The comparison ensures the return type remains consistent
	// Example: !foo || foo.bar === undefined -> foo?.bar === undefined (both return boolean)
	// HOWEVER: For strict null/undefined checks on types that only have one of them,
	// we should use suggestion because converting intermediate checks to optional chaining
	// changes semantics (optional chaining checks for BOTH null and undefined).
	if useSuggestion && hasTrailingComparison {
		// Check if this is a strict check that's "complete" for the type
		// If so, still use suggestion because the intermediate conversions change semantics
		strictCheckRequiresSuggestion := false
		if len(chain) > 0 && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
			// Check what kind of checks we have
			hasNullCheck := false
			hasUndefinedCheck := false
			hasBothCheck := false
			for _, op := range chain {
				if op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeStrictEqualNull {
					hasNullCheck = true
				} else if op.typ == OperandTypeNotStrictEqualUndef || op.typ == OperandTypeStrictEqualUndef {
					hasUndefinedCheck = true
				} else if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull || op.typ == OperandTypeNot {
					hasBothCheck = true
				} else if op.typ == OperandTypeComparison && isNullishComparison(op) {
					// Check what kind of nullish comparison
					if op.node != nil && ast.IsBinaryExpression(op.node) {
						binExpr := op.node.AsBinaryExpression()
						left := unwrapParentheses(binExpr.Left)
						right := unwrapParentheses(binExpr.Right)
						isNull := left.Kind == ast.KindNullKeyword || right.Kind == ast.KindNullKeyword
						isUndefined := (ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
							(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
							ast.IsVoidExpression(left) || ast.IsVoidExpression(right)
						binOp := binExpr.OperatorToken.Kind
						if binOp == ast.KindEqualsEqualsToken {
							hasBothCheck = true
						} else if binOp == ast.KindEqualsEqualsEqualsToken {
							if isNull {
								hasNullCheck = true
							}
							if isUndefined {
								hasUndefinedCheck = true
							}
						}
					}
				}
			}
			hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck && !hasBothCheck
			hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck && !hasBothCheck

			// Check ALL nullable operands to determine if the check is complete for all types
			// A check is complete when all nullable types in the chain match the check type
			if hasOnlyNullCheck || hasOnlyUndefinedCheck {
				allTypesMatchCheck := true
				hasAnyNullableType := false
				for _, op := range chain {
					if op.comparedExpr == nil {
						continue
					}
					info := processor.getTypeInfo(op.comparedExpr)
					// Skip non-nullable operands
					if !info.hasNull && !info.hasUndefined && !info.hasAny && !info.hasUnknown {
						continue
					}
					hasAnyNullableType = true
					// If any type has any/unknown, we can't determine
					if info.hasAny || info.hasUnknown {
						allTypesMatchCheck = false
						break
					}
					// If any type has BOTH, check is incomplete
					if info.hasNull && info.hasUndefined {
						allTypesMatchCheck = false
						break
					}
					// For null-only checks, type should have null but not undefined
					if hasOnlyNullCheck && (!info.hasNull || info.hasUndefined) {
						allTypesMatchCheck = false
						break
					}
					// For undefined-only checks, type should have undefined but not null
					if hasOnlyUndefinedCheck && (info.hasNull || !info.hasUndefined) {
						allTypesMatchCheck = false
						break
					}
				}
				if hasAnyNullableType && allTypesMatchCheck {
					strictCheckRequiresSuggestion = true
				}
			}
		}
		if !strictCheckRequiresSuggestion {
			useSuggestion = false
		}
	}

	// If still using suggestion, check if ANY operand includes BOTH null and undefined
	// (or any/unknown types). If so, the autofix is safe because optional chaining
	// checks for both null and undefined, matching the expected behavior.
	// BUT: If the type only has null OR only undefined (not both), keep using suggestion
	// because the conversion would change semantics.
	if useSuggestion && len(chain) > 0 {
		for _, op := range chain {
			if op.comparedExpr != nil {
				info := processor.getTypeInfo(op.comparedExpr)
				// Safe if type is any/unknown (can't determine, assume safe)
				// or if type includes BOTH null AND undefined
				if info.hasAny || info.hasUnknown || (info.hasNull && info.hasUndefined) {
					useSuggestion = false
					break
				}
			}
		}
	}

	if useSuggestion {
		processor.ctx.ReportNodeWithSuggestions(node, buildPreferOptionalChainMessage(), func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  buildOptionalChainSuggestMessage(),
				FixesArr: fixes,
			}}
		})
	} else {
		processor.ctx.ReportNodeWithFixes(node, buildPreferOptionalChainMessage(), func() []rule.RuleFix {
			return fixes
		})
	}

	// Mark all operands in this chain as reported to avoid overlapping diagnostics
	for _, op := range chain {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
		opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
		processor.reportedRanges[opTextRange] = true
	}
}

func (processor *chainProcessor) handleEmptyObjectPattern(node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	operator := binExpr.OperatorToken.Kind

	// Only for || and ?? operators
	if operator != ast.KindBarBarToken && operator != ast.KindQuestionQuestionToken {
		return
	}

	// Check if right side is empty object literal
	// It can be either {} or ({})
	rightNode := binExpr.Right
	var objLit *ast.ObjectLiteralExpression

	if ast.IsObjectLiteralExpression(rightNode) {
		objLit = rightNode.AsObjectLiteralExpression()
	} else if ast.IsParenthesizedExpression(rightNode) {
		innerExpr := rightNode.AsParenthesizedExpression().Expression
		if ast.IsObjectLiteralExpression(innerExpr) {
			objLit = innerExpr.AsObjectLiteralExpression()
		}
	}

	if objLit == nil {
		return
	}
	if len(objLit.Properties.Nodes) != 0 {
		return
	}

	// When requireNullish is true, only process if the left expression's type includes null/undefined
	// This ensures we only flag patterns where the nullish check is semantically meaningful
	if processor.opts.RequireNullish {
		leftExpr := binExpr.Left
		if !processor.includesExplicitNullish(leftExpr) {
			return
		}
	}

	// Check if parent (or grandparent through parenthesized expression) is property access expression
	// The pattern can be either: foo || {}).bar or (foo || {}).bar
	var accessExpr *ast.Node
	if ast.IsPropertyAccessExpression(node.Parent) || ast.IsElementAccessExpression(node.Parent) {
		accessExpr = node.Parent
	} else if ast.IsParenthesizedExpression(node.Parent) {
		grandParent := node.Parent.Parent
		if grandParent != nil && (ast.IsPropertyAccessExpression(grandParent) || ast.IsElementAccessExpression(grandParent)) {
			accessExpr = grandParent
		}
	}

	if accessExpr == nil {
		return
	}

	// Get the property access info
	var isOptional bool
	var isComputed bool
	var propNode *ast.Node

	if ast.IsPropertyAccessExpression(accessExpr) {
		parentProp := accessExpr.AsPropertyAccessExpression()
		isOptional = parentProp.QuestionDotToken != nil
		isComputed = false
		propNode = parentProp.Name()
	} else {
		parentElem := accessExpr.AsElementAccessExpression()
		isOptional = parentElem.QuestionDotToken != nil
		isComputed = true
		propNode = parentElem.ArgumentExpression
	}

	if isOptional {
		return
	}

	processor.seenLogicals[node] = true

	leftNode := binExpr.Left
	leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, leftNode)
	leftText := processor.sourceText[leftRange.Pos():leftRange.End()]

	// Determine if we need to add parentheses around the left expression
	// We need parentheses when the left side is a complex expression that would
	// have different precedence without them (await, binary ops, etc.)
	// We do NOT add parens just because the original binary expression was parenthesized
	needsParens := ast.IsAwaitExpression(leftNode) ||
		ast.IsBinaryExpression(leftNode) ||
		ast.IsConditionalExpression(leftNode) ||
		ast.IsPrefixUnaryExpression(leftNode) ||
		leftNode.Kind == ast.KindAsExpression ||
		ast.IsVoidExpression(leftNode) ||
		ast.IsTypeOfExpression(leftNode) ||
		leftNode.Kind == ast.KindPostfixUnaryExpression ||
		leftNode.Kind == ast.KindDeleteExpression

	if needsParens {
		leftText = "(" + leftText + ")"
	}

	propRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, propNode)
	propertyText := ""
	if isComputed {
		propertyText = "[" + processor.sourceText[propRange.Pos():propRange.End()] + "]"
	} else {
		propertyText = processor.sourceText[propRange.Pos():propRange.End()]
	}

	newCode := leftText + "?." + propertyText
	accessRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, accessExpr)

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(accessRange, newCode),
	}

	// Use suggestion unless the unsafe option is enabled
	// This pattern changes return type: (foo || {}).bar returns {} when foo is falsy,
	// while foo?.bar returns undefined
	if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		processor.ctx.ReportNodeWithFixes(accessExpr, buildPreferOptionalChainMessage(), func() []rule.RuleFix {
			return fixes
		})
	} else {
		processor.ctx.ReportNodeWithSuggestions(accessExpr, buildPreferOptionalChainMessage(), func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  buildOptionalChainSuggestMessage(),
				FixesArr: fixes,
			}}
		})
	}
}

var PreferOptionalChainRule = rule.Rule{
	Name: "prefer-optional-chain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferOptionalChainOptions](options, "prefer-optional-chain")

		// Create processor instance to manage state
		processor := newChainProcessor(ctx, opts)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsBinaryExpression(node) {
					return
				}

				binExpr := node.AsBinaryExpression()
				operator := binExpr.OperatorToken.Kind

				switch operator {
				case ast.KindAmpersandAmpersandToken:
					processor.processAndChain(node)
				case ast.KindBarBarToken:
					processor.processOrChain(node)
					processor.handleEmptyObjectPattern(node)
				case ast.KindQuestionQuestionToken:
					processor.handleEmptyObjectPattern(node)
				}
			},
		}
	},
}
