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

// isNullishCheckOperand checks if an operand is a null/undefined check (including OperandTypeComparison with null/undef).
// This is used in OR chain fix generation to determine if an operand should be included in the chain.
func isNullishCheckOperand(op Operand) bool {
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

// isAndOperator returns true if this is an AND operator (&&)
func isAndOperator(op ast.Kind) bool {
	return op == ast.KindAmpersandAmpersandToken
}

// isOrLikeOperator returns true if this is an OR-like operator (|| or ??)
// Both handle empty object patterns like (foo || {}).bar or (foo ?? {}).bar
func isOrLikeOperator(op ast.Kind) bool {
	return op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken
}

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

// isChainExtension checks if 'longer' extends 'shorter' in a chain expression.
// Returns true if 'shorter' is a prefix of 'longer' (e.g., foo.bar is a prefix of foo.bar.baz).
// This is an AST-based check that replaces text-based prefix matching.
//
// Examples:
//   - isChainExtension(foo, foo.bar) -> true (property access extends)
//   - isChainExtension(foo.bar, foo.bar[0]) -> true (element access extends)
//   - isChainExtension(foo.bar, foo.bar()) -> true (call extends)
//   - isChainExtension(foo.bar, foo.baz) -> false (different property)
//   - isChainExtension(foo, bar.foo) -> false (different base)
func (processor *chainProcessor) isChainExtension(shorter, longer *ast.Node) bool {
	if shorter == nil || longer == nil {
		return false
	}

	// Get normalized text for both - if shorter is a prefix of longer's normalized text,
	// we need to verify using AST that it's actually a chain extension
	shorterNorm := processor.getNormalizedNodeText(shorter)
	longerNorm := processor.getNormalizedNodeText(longer)

	// Quick check: if normalized texts are equal, it's not an extension
	if shorterNorm == longerNorm {
		return false
	}

	// Quick check: if shorter's text isn't a prefix of longer's text, can't be extension
	if !strings.HasPrefix(longerNorm, shorterNorm) {
		return false
	}

	// Now verify via AST that 'shorter' appears as a base in the chain of 'longer'
	// Walk up the chain of 'longer' and check if any base matches 'shorter'
	current := longer
	for {
		current = unwrapChainNode(current)
		if current == nil {
			return false
		}

		var base *ast.Node
		switch {
		case ast.IsPropertyAccessExpression(current):
			base = current.AsPropertyAccessExpression().Expression
		case ast.IsElementAccessExpression(current):
			base = current.AsElementAccessExpression().Expression
		case ast.IsCallExpression(current):
			base = current.AsCallExpression().Expression
		case ast.IsNonNullExpression(current):
			base = current.AsNonNullExpression().Expression
		case ast.IsTaggedTemplateExpression(current):
			base = current.AsTaggedTemplateExpression().Tag
		default:
			// Reached a terminal node (identifier, literal, etc.)
			return false
		}

		// Check if the base matches 'shorter'
		if processor.getNormalizedNodeText(base) == shorterNorm {
			return true
		}

		// Continue walking up the chain
		current = base
	}
}

// unwrapChainNode unwraps parentheses, type assertions, and non-null expressions
// to get to the underlying chain expression
func unwrapChainNode(node *ast.Node) *ast.Node {
	current := node
	for current != nil {
		switch {
		case ast.IsParenthesizedExpression(current):
			current = current.AsParenthesizedExpression().Expression
		case current.Kind == ast.KindAsExpression:
			current = current.AsAsExpression().Expression
		case current.Kind == ast.KindTypeAssertionExpression:
			current = current.AsTypeAssertion().Expression
		default:
			return current
		}
	}
	return current
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
	isCall      bool // true if this part is a call expression (foo() or foo<T>())
}

// baseText returns the text without the non-null assertion suffix (!)
// This is useful for comparing parts where ! should be ignored
func (p ChainPart) baseText() string {
	if p.hasNonNull && len(p.text) > 0 && p.text[len(p.text)-1] == '!' {
		return p.text[:len(p.text)-1]
	}
	return p.text
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

// validateChainRoot validates that a node is a valid root for chain processing.
// Returns the binary expression if valid, nil otherwise.
// Also returns whether the node has already been seen/processed.
func (processor *chainProcessor) validateChainRoot(node *ast.Node, operatorKind ast.Kind) (*ast.BinaryExpression, bool) {
	if !ast.IsBinaryExpression(node) {
		return nil, false
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != operatorKind {
		return nil, false
	}

	// Skip if inside JSX - semantic difference
	// In JSX, foo && foo.bar returns false/null/undefined (rendered as-is)
	// while foo?.bar always returns undefined
	if isInsideJSX(node) {
		return nil, false
	}

	return binExpr, true
}

// isAndChainAlreadySeen checks if an AND chain node has already been seen or processed.
// Returns true if the node should be skipped.
func (processor *chainProcessor) isAndChainAlreadySeen(node *ast.Node) bool {
	if processor.seenLogicals[node] {
		return true
	}

	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeStart, nodeEnd := nodeRange.Pos(), nodeRange.End()

	for _, processedRange := range processor.processedAndRanges {
		// Two ranges overlap if: start1 < end2 && start2 < end1
		if nodeStart < processedRange.end && processedRange.start < nodeEnd {
			processor.seenLogicals[node] = true
			return true
		}
	}
	return false
}

// isOrChainAlreadySeen checks if an OR chain node has already been seen.
// Returns true if the node should be skipped.
func (processor *chainProcessor) isOrChainAlreadySeen(node *ast.Node) bool {
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeTextRange := textRange{start: nodeRange.Pos(), end: nodeRange.End()}
	return processor.seenLogicalRanges[nodeTextRange]
}

// markAndChainAsSeen marks an AND chain node and its range as processed.
func (processor *chainProcessor) markAndChainAsSeen(node *ast.Node) {
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeStart, nodeEnd := nodeRange.Pos(), nodeRange.End()
	processor.processedAndRanges = append(processor.processedAndRanges, textRange{start: nodeStart, end: nodeEnd})
}

// markOrChainAsSeen marks an OR chain node as seen.
func (processor *chainProcessor) markOrChainAsSeen(node *ast.Node) {
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeTextRange := textRange{start: nodeRange.Pos(), end: nodeRange.End()}
	processor.seenLogicalRanges[nodeTextRange] = true
}

// isOrChainNestedInLargerChain checks if an OR chain node is part of a larger || chain.
// Returns true if this node should skip processing (parent will handle it).
func (processor *chainProcessor) isOrChainNestedInLargerChain(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if ast.IsParenthesizedExpression(parent) {
			parent = parent.Parent
			continue
		}
		if ast.IsBinaryExpression(parent) {
			parentBin := parent.AsBinaryExpression()
			if parentBin.OperatorToken.Kind == ast.KindBarBarToken {
				leftUnwrapped := unwrapParentheses(parentBin.Left)
				rightUnwrapped := unwrapParentheses(parentBin.Right)
				if leftUnwrapped == node || rightUnwrapped == node {
					return true
				}
			}
		}
		break
	}
	return false
}

// flattenAndMarkLogicals recursively flattens a logical expression and marks all nodes as seen.
// This prevents nested nodes from being processed separately.
func (processor *chainProcessor) flattenAndMarkLogicals(node *ast.Node, operatorKind ast.Kind) []*ast.Node {
	unwrapped := unwrapParentheses(node)
	if !ast.IsBinaryExpression(unwrapped) {
		return nil
	}
	binExpr := unwrapped.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != operatorKind {
		return nil
	}

	// Mark both wrapped and unwrapped versions
	processor.seenLogicals[node] = true
	processor.seenLogicals[unwrapped] = true

	result := []*ast.Node{node, unwrapped}
	result = append(result, processor.flattenAndMarkLogicals(binExpr.Left, operatorKind)...)
	result = append(result, processor.flattenAndMarkLogicals(binExpr.Right, operatorKind)...)
	return result
}

// hasAnyOperandBeenReported checks if any operand node has already been reported.
func (processor *chainProcessor) hasAnyOperandBeenReported(operandNodes []*ast.Node) bool {
	for _, n := range operandNodes {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
		opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
		if processor.reportedRanges[opTextRange] {
			return true
		}
	}
	return false
}

// markChainOperandsAsReported marks all operands in a chain as reported.
func (processor *chainProcessor) markChainOperandsAsReported(chain []Operand) {
	for _, op := range chain {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
		opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
		processor.reportedRanges[opTextRange] = true
	}
}

// reportChainWithFixes reports a chain issue with the appropriate fix or suggestion.
// This is a shared helper used by both AND and OR chain processing.
func (processor *chainProcessor) reportChainWithFixes(node *ast.Node, fixes []rule.RuleFix, useSuggestion bool) {
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
}

// isValidOperandForChainType checks if an operand type is valid for the given chain type.
// For AND chains, most operand types are valid.
// For OR chains, we need specific types like negations, comparisons, and null checks.
func isValidOperandForChainType(op Operand, operatorKind ast.Kind) bool {
	if isAndOperator(operatorKind) {
		// For AND chains, all types except Invalid are potentially valid
		// The chain-building logic will determine if they can be chained
		return op.typ != OperandTypeInvalid
	}

	// For OR chains, we need specific operand types
	switch op.typ {
	case OperandTypeNot,
		OperandTypeComparison,
		OperandTypePlain,
		OperandTypeTypeofCheck,
		OperandTypeNotStrictEqualNull,
		OperandTypeNotStrictEqualUndef,
		OperandTypeNotEqualBoth,
		OperandTypeStrictEqualNull,
		OperandTypeStrictEqualUndef,
		OperandTypeEqualNull:
		return true
	default:
		return false
	}
}

// isInvalidOrChainStartingOperand checks if an operand should NOT start an OR chain.
// This catches patterns like `foo != null || foo.bar` which have opposite semantics
// compared to optional chaining.
func (processor *chainProcessor) isInvalidOrChainStartingOperand(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return false
	}

	unwrapped := unwrapParentheses(op.node)
	if !ast.IsBinaryExpression(unwrapped) {
		return false
	}

	binExpr := unwrapped.AsBinaryExpression()
	binOp := binExpr.OperatorToken.Kind

	// Check for != or !== operators
	if binOp != ast.KindExclamationEqualsToken && binOp != ast.KindExclamationEqualsEqualsToken {
		return false
	}

	left := unwrapParentheses(binExpr.Left)
	right := unwrapParentheses(binExpr.Right)

	// Check if comparing to null/undefined
	isLeftNullish := left.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(left)
	isRightNullish := right.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(right)

	if !isLeftNullish && !isRightNullish {
		return false
	}

	// Determine which side is the checked expression
	var checkedExpr *ast.Node
	if isRightNullish {
		checkedExpr = left
	} else {
		checkedExpr = right
	}

	// If the checked expression is a base identifier, don't start chain
	isBaseIdentifier := ast.IsIdentifier(checkedExpr) || checkedExpr.Kind == ast.KindThisKeyword
	return isBaseIdentifier
}

// shouldAllowCallChainExtension determines if we should allow extending through a call expression
// when compareNodes returns NodeInvalid. This is allowed in specific patterns:
// - AND chains: when both prev and current operands are plain truthiness checks
// - OR chains: when both are negations or both are nullish comparisons
// - Either chain: when the unsafe option is enabled
func (processor *chainProcessor) shouldAllowCallChainExtension(prevOp, currentOp Operand, operatorKind ast.Kind) bool {
	// Unsafe option always allows extension
	if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		return true
	}

	if isAndOperator(operatorKind) {
		// AND chains: allow when both operands are plain truthiness checks
		// Pattern: foo && foo() && foo().bar
		return prevOp.typ == OperandTypePlain && currentOp.typ == OperandTypePlain
	}

	// OR chains: allow when both are negations or both are nullish comparisons
	// Pattern: !foo() || !foo().bar
	// Pattern: foo.bar() === null || foo.bar().baz === null
	isNegationChain := prevOp.typ == OperandTypeNot && currentOp.typ == OperandTypeNot
	isNullishComparisonChain := isOrChainNullishCheck(prevOp) && isOrChainNullishCheck(currentOp)
	return isNegationChain || isNullishComparisonChain
}

// tryExtendThroughCallExpression checks if we can extend the chain through a call expression
// when compareNodes returned NodeInvalid. Returns the updated comparison result.
func (processor *chainProcessor) tryExtendThroughCallExpression(lastExpr *ast.Node, currentOp Operand, firstOpExpr *ast.Node, operatorKind ast.Kind) NodeComparisonResult {
	if lastExpr == nil {
		return NodeInvalid
	}

	lastUnwrapped := lastExpr
	for ast.IsParenthesizedExpression(lastUnwrapped) {
		lastUnwrapped = lastUnwrapped.AsParenthesizedExpression().Expression
	}

	// Only applies to call or new expressions
	if !ast.IsCallExpression(lastUnwrapped) && !ast.IsNewExpression(lastUnwrapped) {
		return NodeInvalid
	}

	// For AND chains, check that the first operand is not rooted in a new expression
	// (each `new X()` creates a fresh instance)
	if isAndOperator(operatorKind) && firstOpExpr != nil {
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
				break
			}
		}

		if baseExpr != nil && ast.IsNewExpression(baseExpr) {
			return NodeInvalid
		}
	}

	// Check if currentOp extends lastExpr using AST-based chain extension check
	if processor.isChainExtension(lastExpr, currentOp.comparedExpr) {
		return NodeSubset
	}

	return NodeInvalid
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
	// Use AST-based chain extension check
	if processor.isChainExtension(left, right) {
		return NodeSubset
	}

	// Check if right is a subset of left
	if processor.isChainExtension(right, left) {
		return NodeSuperset
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
				isCall:      true,
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
				isLastPart := i == len(parts)-1
				if part.isCall && isLastPart && callShouldBeOptional {
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
		if stripNonNullAssertions && i < len(parts)-1 && optionalParts[i+1] && part.hasNonNull {
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
func (processor *chainProcessor) parseOperand(node *ast.Node, operatorKind ast.Kind) Operand {
	isAndChain := isAndOperator(operatorKind)
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

// validateChain performs common validation checks on a chain.
// Returns true if the chain is valid and should be processed.
// This is a shared helper used by both AND and OR chain processing.
func (processor *chainProcessor) validateChain(chain []Operand, operatorKind ast.Kind) bool {
	// Need at least 2 operands
	if len(chain) < 2 {
		return false
	}

	// Check if all operands have the same base identifier
	if !processor.hasSameBaseIdentifier(chain) {
		return false
	}

	// Ensure at least one operand involves property/element/call access
	if !processor.hasPropertyAccessInChain(chain) {
		return false
	}

	// Check requireNullish option
	if processor.shouldSkipForRequireNullish(chain, operatorKind) {
		return false
	}

	return true
}

// filterOverlappingChains filters out chains that overlap with longer chains.
// This is used when multiple chains are detected in the same expression.
func (processor *chainProcessor) filterOverlappingChains(chains [][]Operand) [][]Operand {
	if len(chains) <= 1 {
		return chains
	}

	// Build a list of chain ranges
	type chainWithRange struct {
		chain    []Operand
		startPos int
		endPos   int
		length   int
	}

	chainRanges := make([]chainWithRange, len(chains))
	for i, chain := range chains {
		if len(chain) == 0 {
			continue
		}
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
	return filteredChains
}

// hasChainOverlapWithReported checks if any operand in the chain overlaps with a previously reported range.
func (processor *chainProcessor) hasChainOverlapWithReported(chain []Operand) bool {
	for _, op := range chain {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
		opStart, opEnd := opRange.Pos(), opRange.End()

		for reportedRange := range processor.reportedRanges {
			if opStart < reportedRange.end && reportedRange.start < opEnd {
				return true
			}
		}
	}
	return false
}

// shouldSkipForRequireNullish checks if the chain should be skipped based on requireNullish option.
// When requireNullish is true, only convert chains that have explicit nullish checks or nullable types.
func (processor *chainProcessor) shouldSkipForRequireNullish(chain []Operand, operatorKind ast.Kind) bool {
	if !processor.opts.RequireNullish {
		return false
	}

	// For OR chains starting with negation, skip entirely
	if !isAndOperator(operatorKind) && len(chain) > 0 && chain[0].typ == OperandTypeNot {
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
		if isAndOperator(operatorKind) && i < len(chain)-1 && op.comparedExpr != nil {
			if processor.includesExplicitNullish(op.comparedExpr) {
				return false // Has nullish type, don't skip
			}
		}
	}
	return true // No nullish context found, skip
}

// processChain is the unified chain processor that handles AND, OR, and nullish coalescing chains.
// This follows typescript-eslint's analyzeChain pattern with a single entry point
// that dispatches to chain-type-specific analyzers.
func (processor *chainProcessor) processChain(node *ast.Node, operatorKind ast.Kind) {
	// For OR-like chains (|| and ??), also check for empty object pattern
	// This handles patterns like (foo || {}).bar or (foo ?? {}).bar
	if isOrLikeOperator(operatorKind) {
		processor.handleEmptyObjectPattern(node)
	}

	// For nullish coalescing, we only handle empty object patterns
	// The ?? operator doesn't create the same chain patterns as && or ||
	if operatorKind == ast.KindQuestionQuestionToken {
		return
	}

	// Validate this is a valid chain root
	_, ok := processor.validateChainRoot(node, operatorKind)
	if !ok {
		return
	}

	// Check if already seen - use chain-type-specific tracking
	if isAndOperator(operatorKind) {
		if processor.isAndChainAlreadySeen(node) {
			return
		}
		processor.markAndChainAsSeen(node)
		// Flatten and mark ALL logical expressions in this chain
		_ = processor.flattenAndMarkLogicals(node, operatorKind)
	} else {
		// For OR chains, also check if nested in larger chain
		if processor.isOrChainNestedInLargerChain(node) {
			return
		}
		if processor.isOrChainAlreadySeen(node) {
			return
		}
		processor.markOrChainAsSeen(node)
	}

	// Collect operands
	var operandNodes []*ast.Node
	if isAndOperator(operatorKind) {
		operandNodes = processor.collectOperands(node, operatorKind)
	} else {
		var collectedBinaryRanges []textRange
		operandNodes, collectedBinaryRanges = processor.collectOperandsWithRanges(node, operatorKind)
		// Mark all collected binary expression ranges as seen
		for _, r := range collectedBinaryRanges {
			processor.seenLogicalRanges[r] = true
		}
	}

	if len(operandNodes) < 2 {
		return
	}

	// Check if any operand has already been reported
	if processor.hasAnyOperandBeenReported(operandNodes) {
		return
	}

	// Parse operands
	operands := make([]Operand, len(operandNodes))
	for i, n := range operandNodes {
		operands[i] = processor.parseOperand(n, operatorKind)
	}

	// Build chains using the unified chain builder
	chains := processor.buildChains(operands, operatorKind)

	// Filter out chains that overlap with longer chains
	chains = processor.filterOverlappingChains(chains)

	// Process each chain
	for _, chain := range chains {
		// Skip if any operand overlaps with a previously reported range
		if processor.hasChainOverlapWithReported(chain) {
			continue
		}

		// Validate the chain (may return modified/truncated chain or nil)
		validatedChain := processor.validateChainForReporting(chain, operatorKind)
		if validatedChain == nil {
			continue
		}

		// Generate fix and report
		processor.generateFixAndReport(node, validatedChain, operandNodes, operatorKind)
	}
}

// buildChains builds chains from operands using chain-type-specific logic.
// For AND chains, this handles multiple independent chains.
// For OR chains, this builds a single chain.
func (processor *chainProcessor) buildChains(operands []Operand, operatorKind ast.Kind) [][]Operand {
	if isAndOperator(operatorKind) {
		return processor.buildAndChains(operands)
	}
	return processor.buildOrChains(operands)
}

// buildAndChains builds chains for AND expressions.
// This handles multiple independent chains (e.g., foo && foo.bar && bar && bar.baz).
func (processor *chainProcessor) buildAndChains(operands []Operand) [][]Operand {
	var allChains [][]Operand
	var currentChain []Operand
	var lastExpr *ast.Node
	var lastCheckType OperandType
	var chainComplete bool
	var stopProcessing bool

	i := 0
	for i < len(operands) && !stopProcessing {
		op := operands[i]

		// Check if operand is invalid for AND chains
		// These operand types should break the chain:
		// - OperandTypeInvalid: obviously invalid
		// - OperandTypeNegatedAndOperand: negations (!foo) in AND chains have wrong semantics
		// - OperandTypeEqualNull: foo == null in AND chains has opposite semantics
		// - OperandTypeStrictEqualNull/Undef: strict equality checks break AND chains
		if op.typ == OperandTypeInvalid ||
			op.typ == OperandTypeNegatedAndOperand ||
			op.typ == OperandTypeEqualNull ||
			op.typ == OperandTypeStrictEqualNull ||
			op.typ == OperandTypeStrictEqualUndef {
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
			currentChain = append(currentChain, op)
			lastExpr = op.comparedExpr
			if op.typ != OperandTypePlain {
				lastCheckType = op.typ
			}
			chainComplete = false
			i++
			continue
		}

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

		cmp := processor.compareNodes(lastExpr, op.comparedExpr)

		// Check for strict nullish check on call result pattern
		if len(currentChain) > 0 {
			prevOp := currentChain[len(currentChain)-1]
			if shouldStop := processor.shouldStopAtStrictNullishCheck(prevOp); shouldStop {
				if len(currentChain) >= 2 {
					allChains = append(allChains, currentChain)
				}
				currentChain = nil
				break
			}
		}

		// Try call chain extension if compareNodes returned Invalid
		if cmp == NodeInvalid && len(currentChain) > 0 {
			prevOp := currentChain[len(currentChain)-1]
			if processor.shouldAllowCallChainExtension(prevOp, op, ast.KindAmpersandAmpersandToken) {
				firstOpExpr := currentChain[0].comparedExpr
				if extCmp := processor.tryExtendThroughCallExpression(lastExpr, op, firstOpExpr, ast.KindAmpersandAmpersandToken); extCmp == NodeSubset {
					cmp = NodeSubset
				}
			}
		}

		// Handle complementary null+undefined pairs
		if cmp == NodeEqual && i+1 < len(operands) {
			if pair := processor.tryMergeComplementaryPair(op, operands[i+1], ast.KindAmpersandAmpersandToken); pair != nil {
				currentChain = append(currentChain, *pair...)
				lastExpr = (*pair)[len(*pair)-1].comparedExpr
				i += 2
				continue
			}
		}

		// Check for inconsistent nullish checks
		if isNullishCheckType(op.typ) && lastCheckType != OperandTypeInvalid {
			if !processor.areNullishChecksConsistent(lastCheckType, op.typ) {
				if len(currentChain) >= 2 {
					allChains = append(allChains, currentChain)
				}
				currentChain = nil
				break
			}
		}

		if cmp == NodeSubset || cmp == NodeEqual {
			currentChain = append(currentChain, op)
			lastExpr = op.comparedExpr
			if op.typ != OperandTypePlain && isNullishCheckType(op.typ) {
				lastCheckType = op.typ
			}
			i++
			continue
		}

		// Chain broken
		if len(currentChain) >= 2 {
			allChains = append(allChains, currentChain)
		}
		currentChain = []Operand{op}
		lastExpr = op.comparedExpr
		lastCheckType = OperandTypeInvalid
		if op.typ != OperandTypePlain {
			lastCheckType = op.typ
		}
		i++
	}

	// Finalize remaining chain
	if len(currentChain) >= 2 {
		allChains = append(allChains, currentChain)
	}

	// Post-process chains to remove trailing duplicate Plain operands
	// A duplicate is when the last operand has NodeEqual comparison with second-to-last
	// AND both are Plain type (not nullish checks forming a pair)
	// Example to trim: [foo, foo.toString(), foo.toString()] -> [foo, foo.toString()]
	// Example to keep: [foo, foo.bar.baz, foo.bar.baz, foo.bar.baz.buzz] -> keep all (duplicate in middle, not at end)
	for i := range allChains {
		chain := allChains[i]
		for len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			secondToLastOp := chain[len(chain)-2]
			// Check if last two operands are both Plain and represent the same expression
			if lastOp.typ == OperandTypePlain && secondToLastOp.typ == OperandTypePlain {
				cmp := processor.compareNodes(secondToLastOp.comparedExpr, lastOp.comparedExpr)
				if cmp == NodeEqual {
					// Trailing duplicate - remove the last operand
					chain = chain[:len(chain)-1]
					allChains[i] = chain
					continue // Check again in case there are multiple trailing duplicates
				}
			}
			break // No more trailing duplicates
		}
	}

	return allChains
}

// buildOrChains builds chains for OR expressions.
// OR chains typically have a single chain (e.g., !foo || !foo.bar || !foo.bar.baz).
func (processor *chainProcessor) buildOrChains(operands []Operand) [][]Operand {
	var chain []Operand
	var lastExpr *ast.Node

	for i := range operands {
		op := operands[i]

		if !isValidOperandForChainType(op, ast.KindBarBarToken) {
			if len(chain) >= 2 {
				break
			}
			chain = nil
			lastExpr = nil
			continue
		}

		if len(chain) == 0 {
			if processor.isInvalidOrChainStartingOperand(op) {
				continue
			}
			chain = append(chain, op)
			lastExpr = op.comparedExpr
			continue
		}

		cmp := processor.compareNodes(lastExpr, op.comparedExpr)

		// Try call chain extension
		if cmp == NodeInvalid && len(chain) > 0 {
			prevOp := chain[len(chain)-1]
			if processor.shouldAllowCallChainExtension(prevOp, op, ast.KindBarBarToken) {
				if extCmp := processor.tryExtendThroughCallExpression(lastExpr, op, nil, ast.KindBarBarToken); extCmp == NodeSubset {
					cmp = NodeSubset
				}
			}
		}

		if cmp == NodeSubset || cmp == NodeEqual {
			// Skip non-nullish comparisons on same expression after negation
			if cmp == NodeEqual && op.typ == OperandTypeComparison && len(chain) > 0 && !isNullishComparison(op) {
				lastOp := chain[len(chain)-1]
				if lastOp.typ == OperandTypeNot || lastOp.typ == OperandTypeNotStrictEqualNull ||
					lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeNotEqualBoth ||
					lastOp.typ == OperandTypePlain {
					if len(chain) >= 2 {
						break
					}
				}
			}

			chain = append(chain, op)
			lastExpr = op.comparedExpr
			continue
		}

		// Chain broken
		if len(chain) >= 2 {
			break
		}
		chain = []Operand{op}
		lastExpr = op.comparedExpr
	}

	if len(chain) < 2 {
		return nil
	}

	return [][]Operand{chain}
}

// shouldStopAtStrictNullishCheck checks if we should stop extending the chain
// at a strict nullish check on a call/element access result.
func (processor *chainProcessor) shouldStopAtStrictNullishCheck(prevOp Operand) bool {
	if !isStrictNullishCheck(prevOp.typ) || prevOp.comparedExpr == nil {
		return false
	}

	prevUnwrapped := prevOp.comparedExpr
	for ast.IsParenthesizedExpression(prevUnwrapped) {
		prevUnwrapped = prevUnwrapped.AsParenthesizedExpression().Expression
	}

	isCallOrNew := ast.IsCallExpression(prevUnwrapped) || ast.IsNewExpression(prevUnwrapped)
	isElementAccess := ast.IsElementAccessExpression(prevUnwrapped)

	if !isCallOrNew && !isElementAccess {
		return false
	}

	isAnyOrUnknown := processor.typeIsAnyOrUnknown(prevOp.comparedExpr)
	hasNull := processor.typeIncludesNull(prevOp.comparedExpr)
	hasUndefined := processor.typeIncludesUndefined(prevOp.comparedExpr)

	isIncomplete := !isAnyOrUnknown && hasNull && hasUndefined

	isMismatched := false
	if !isAnyOrUnknown {
		if prevOp.typ == OperandTypeNotStrictEqualUndef && !hasUndefined && !hasNull {
			isMismatched = true
		}
		if prevOp.typ == OperandTypeNotStrictEqualNull && !hasNull && !hasUndefined {
			isMismatched = true
		}
	}

	if isCallOrNew {
		return isIncomplete || isMismatched
	}
	if isElementAccess {
		if isMismatched {
			return true
		}
		if isIncomplete && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
			return true
		}
	}

	return false
}

// tryMergeComplementaryPair checks if two operands form a complementary null+undefined pair.
func (processor *chainProcessor) tryMergeComplementaryPair(op1, op2 Operand, operatorKind ast.Kind) *[]Operand {
	if op1.comparedExpr == nil || op2.comparedExpr == nil {
		return nil
	}

	cmp := processor.compareNodes(op1.comparedExpr, op2.comparedExpr)
	if cmp != NodeEqual {
		return nil
	}

	if isAndOperator(operatorKind) {
		// AND chains: !== null && !== undefined
		isOp1Null := op1.typ == OperandTypeNotStrictEqualNull
		isOp1Undef := op1.typ == OperandTypeNotStrictEqualUndef || op1.typ == OperandTypeTypeofCheck
		isOp2Null := op2.typ == OperandTypeNotStrictEqualNull
		isOp2Undef := op2.typ == OperandTypeNotStrictEqualUndef || op2.typ == OperandTypeTypeofCheck

		if (isOp1Null && isOp2Undef) || (isOp1Undef && isOp2Null) {
			result := []Operand{op1, op2}
			return &result
		}
	} else {
		// OR chains: === null || === undefined
		isOp1Null := op1.typ == OperandTypeStrictEqualNull
		isOp1Undef := op1.typ == OperandTypeStrictEqualUndef
		isOp2Null := op2.typ == OperandTypeStrictEqualNull
		isOp2Undef := op2.typ == OperandTypeStrictEqualUndef

		if (isOp1Null && isOp2Undef) || (isOp1Undef && isOp2Null) {
			result := []Operand{op1, op2}
			return &result
		}
	}

	return nil
}

// areNullishChecksConsistent checks if two nullish check types are consistent.
func (processor *chainProcessor) areNullishChecksConsistent(type1, type2 OperandType) bool {
	// Loose checks (!=, ==) are always consistent with each other and with strict checks
	if type1 == OperandTypeNotEqualBoth || type1 == OperandTypeEqualNull {
		return true
	}
	if type2 == OperandTypeNotEqualBoth || type2 == OperandTypeEqualNull {
		return true
	}

	// Strict checks must match
	isType1Null := type1 == OperandTypeNotStrictEqualNull || type1 == OperandTypeStrictEqualNull
	isType1Undef := type1 == OperandTypeNotStrictEqualUndef || type1 == OperandTypeStrictEqualUndef || type1 == OperandTypeTypeofCheck
	isType2Null := type2 == OperandTypeNotStrictEqualNull || type2 == OperandTypeStrictEqualNull
	isType2Undef := type2 == OperandTypeNotStrictEqualUndef || type2 == OperandTypeStrictEqualUndef || type2 == OperandTypeTypeofCheck

	// Allow null+undefined pairs (complementary)
	if (isType1Null && isType2Undef) || (isType1Undef && isType2Null) {
		return true
	}

	// Same type is consistent
	return (isType1Null && isType2Null) || (isType1Undef && isType2Undef)
}

// validateChainForReporting performs additional validation before reporting a chain.
// Returns the validated chain (possibly modified/truncated) or nil if the chain should be skipped.
func (processor *chainProcessor) validateChainForReporting(chain []Operand, operatorKind ast.Kind) []Operand {
	if !processor.validateChain(chain, operatorKind) {
		return nil
	}

	// Chain-type-specific validation
	if isAndOperator(operatorKind) {
		return processor.validateAndChainForReporting(chain)
	}
	return processor.validateOrChainForReporting(chain)
}

// validateAndChainForReporting performs AND-chain-specific validation.
// Returns the validated chain or nil if the chain should be skipped.
func (processor *chainProcessor) validateAndChainForReporting(chain []Operand) []Operand {
	// Skip if first operand is Plain but contains optional chaining with different base
	if len(chain) >= 2 {
		firstOp := chain[0]
		if firstOp.typ == OperandTypePlain && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
			return nil
		}
	}

	// Skip chains where the first operand ALREADY has optional chaining AND a STRICT check
	// These patterns result from a previous partial fix that intentionally stopped
	// Example: foo?.bar?.baz !== undefined && foo.bar.baz.buzz
	// EXCEPTION: Don't skip if this is a "split strict equals" pattern
	if len(chain) >= 2 {
		firstOp := chain[0]
		if isStrictNullishCheck(firstOp.typ) && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
			if !processor.isSplitStrictEqualsPattern(chain) {
				return nil
			}
		}
	}

	// Skip single-operand chains (need at least 2 operands to form a chain)
	if len(chain) < 2 {
		return nil
	}

	// Ensure at least one operand involves property/element/call access
	if !processor.hasPropertyAccessInChain(chain) {
		return nil
	}

	// Skip if all operands check the same expression (nothing to chain)
	if len(chain) >= 2 {
		if processor.allOperandsCheckSameExpression(chain) {
			// Exception: split strict equals pattern
			if !processor.isSplitStrictEqualsPattern(chain) {
				return nil
			}
		}
	}

	// Skip chains with strict checks and optimal optional chaining
	if processor.shouldSkipOptimalStrictChecks(chain) {
		return nil
	}

	// Additional check for 2-operand chains where second has optional in extension
	if len(chain) == 2 {
		firstOp := chain[0]
		secondOp := chain[1]

		if processor.containsOptionalChain(secondOp.comparedExpr) {
			firstParts := processor.flattenForFix(firstOp.node)
			secondParts := processor.flattenForFix(secondOp.node)

			// Check for redundant check pattern: foo && foo?.()
			isRedundantCheck := false
			if len(secondParts) == len(firstParts)+1 {
				allMatch := true
				for i := range firstParts {
					if firstParts[i].text != secondParts[i].text || firstParts[i].optional != secondParts[i].optional {
						allMatch = false
						break
					}
				}
				if allMatch && secondParts[len(secondParts)-1].optional {
					isRedundantCheck = true
				}
			}

			if !isRedundantCheck {
				if len(secondParts) > len(firstParts) && len(firstParts) > 0 {
					basesMatch := true
					for i := range firstParts {
						if firstParts[i].text != secondParts[i].text {
							basesMatch = false
							break
						}
					}

					if basesMatch {
						hasOptionalInExtension := false
						for i := len(firstParts); i < len(secondParts); i++ {
							if secondParts[i].optional {
								hasOptionalInExtension = true
								break
							}
						}
						if hasOptionalInExtension {
							return nil
						}
					}
				}
			}
		}
	}

	// Check if we should apply requireNullish option
	if processor.shouldSkipForRequireNullish(chain, ast.KindAmpersandAmpersandToken) {
		return nil
	}

	// Skip for void type in plain chains
	if len(chain) > 0 && chain[0].typ == OperandTypePlain {
		if chain[0].comparedExpr != nil && processor.hasVoidType(chain[0].comparedExpr) {
			return nil
		}
	}

	// Check for non-null assertions without unsafe option
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		for _, op := range chain {
			if op.node != nil && ast.IsNonNullExpression(op.node) {
				return nil
			}
		}
	}

	// Check for incomplete nullish checks
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		if processor.hasIncompleteNullishCheck(chain) {
			return nil
		}
	}

	// Check type-checking options for "loose boolean" operands
	for i, op := range chain {
		if op.typ == OperandTypePlain {
			if i == 0 {
				if processor.shouldSkipByType(op.comparedExpr) {
					return nil
				}
				if processor.wouldChangeReturnType(op.comparedExpr) {
					if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
						return nil
					}
				}
			}
		}
		// Special case: typeof check on non-nullable expression with call expression
		if op.typ == OperandTypeTypeofCheck && op.comparedExpr != nil && !processor.includesNullish(op.comparedExpr) {
			if len(chain) == 2 {
				lastOp := chain[len(chain)-1]
				if lastOp.comparedExpr != nil {
					unwrapped := unwrapParentheses(lastOp.comparedExpr)
					if ast.IsCallExpression(unwrapped) {
						return nil
					}
				}
			}
		}
	}

	// Check for trailing comparison safety
	if len(chain) >= 2 {
		lastOp := chain[len(chain)-1]
		if processor.isUnsafeTrailingComparison(chain, lastOp) {
			return nil
		}
	}

	return chain
}

// hasIncompleteNullishCheck checks if the chain has an incomplete nullish check pattern.
// Optional chaining checks for BOTH null AND undefined, so if the chain only checks
// for one but not both, it's unsafe to convert.
func (processor *chainProcessor) hasIncompleteNullishCheck(chain []Operand) bool {
	hasNullCheck := false
	hasUndefinedCheck := false
	hasBothCheck := false
	hasPlainTruthinessCheck := false

	// Determine guard operands (exclude trailing comparisons/accesses)
	guardOperands := chain
	if len(chain) >= 2 {
		lastOp := chain[len(chain)-1]
		if isTrailingComparisonType(lastOp.typ) {
			guardOperands = chain[:len(chain)-1]
		} else if lastOp.typ == OperandTypePlain && lastOp.comparedExpr != nil {
			prevOp := chain[len(chain)-2]
			if prevOp.comparedExpr != nil {
				lastParts := processor.flattenForFix(lastOp.comparedExpr)
				prevParts := processor.flattenForFix(prevOp.comparedExpr)
				if len(lastParts) > len(prevParts) {
					guardOperands = chain[:len(chain)-1]
				}
			}
		}
	}

	hasTypeofCheck := false
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
			hasTypeofCheck = true
			hasUndefinedCheck = true
		}
	}

	// Check for trailing optional chaining
	hasTrailingOptionalChaining := false
	if len(chain) >= 2 && len(guardOperands) < len(chain) {
		lastOp := chain[len(chain)-1]
		if lastOp.comparedExpr != nil && processor.containsOptionalChain(lastOp.comparedExpr) {
			hasTrailingOptionalChaining = true
		}
	}

	// Check if first op's type doesn't include nullish
	firstOpNotNullish := false
	if len(guardOperands) > 0 && guardOperands[0].comparedExpr != nil {
		if !processor.includesNullish(guardOperands[0].comparedExpr) {
			firstOpNotNullish = true
		}
	}

	// Check if strict check is complete for the type
	strictCheckIsComplete := false
	if len(guardOperands) > 0 && guardOperands[0].comparedExpr != nil {
		info := processor.getTypeInfo(guardOperands[0].comparedExpr)
		if !info.hasAny && !info.hasUnknown {
			hasNull := info.hasNull
			hasUndefined := info.hasUndefined
			hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
			hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck

			if hasOnlyNullCheck && hasNull && !hasUndefined {
				strictCheckIsComplete = true
			}
			if hasOnlyUndefinedCheck && hasUndefined && !hasNull {
				strictCheckIsComplete = true
			}
		}
	}

	if !hasPlainTruthinessCheck && !hasBothCheck && !hasTypeofCheck && !hasTrailingBothCheck && !hasTrailingOptionalChaining && !firstOpNotNullish && !strictCheckIsComplete {
		hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
		hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck
		if hasOnlyNullCheck || hasOnlyUndefinedCheck {
			return true // Incomplete nullish check
		}
	}

	return false
}

// isUnsafeTrailingComparison checks if the trailing comparison would change behavior
// when the base is nullish.
func (processor *chainProcessor) isUnsafeTrailingComparison(chain []Operand, lastOp Operand) bool {
	isTrailingComparison := lastOp.typ == OperandTypeComparison
	if !isTrailingComparison && len(chain) >= 2 &&
		(lastOp.typ == OperandTypeNotStrictEqualNull ||
			lastOp.typ == OperandTypeNotStrictEqualUndef ||
			lastOp.typ == OperandTypeNotEqualBoth) {
		prevOp := chain[len(chain)-2]
		if lastOp.comparedExpr != nil && prevOp.comparedExpr != nil {
			lastParts := processor.flattenForFix(lastOp.comparedExpr)
			prevParts := processor.flattenForFix(prevOp.comparedExpr)
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
					if isNullish || isUndeclaredVar {
						unsafe = true
					}
				case ast.KindEqualsEqualsEqualsToken: // ===
					if isUndefined || isUndeclaredVar {
						unsafe = true
					}
				case ast.KindExclamationEqualsToken: // !=
					if !isNullish {
						unsafe = true
					}
				case ast.KindExclamationEqualsEqualsToken: // !==
					if !isUndefined {
						unsafe = true
					}
				}

				if unsafe && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
					return true
				}
			}
		}
	}

	return false
}

// validateOrChainNullishChecks validates OR chain nullish check patterns and may truncate the chain.
// Returns the validated (possibly truncated) chain or nil if the chain should be skipped.
func (processor *chainProcessor) validateOrChainNullishChecks(chain []Operand) []Operand {
	if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		return chain
	}

	hasNullCheck := false
	hasUndefinedCheck := false
	hasBothCheck := false

	for _, op := range chain {
		if op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeStrictEqualNull {
			hasNullCheck = true
		} else if op.typ == OperandTypeNotStrictEqualUndef || op.typ == OperandTypeStrictEqualUndef {
			hasUndefinedCheck = true
		} else if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
			hasBothCheck = true
		} else if op.typ == OperandTypeTypeofCheck {
			hasUndefinedCheck = true
		} else if op.typ == OperandTypeNot {
			hasBothCheck = true
		} else if op.typ == OperandTypeComparison && op.node != nil {
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

	// Check first operand type
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

	// Check if strict check is complete
	strictCheckIsComplete := true
	hasAnyNullableOperand := false
	for _, op := range chain {
		if op.comparedExpr == nil {
			continue
		}
		info := processor.getTypeInfo(op.comparedExpr)
		if !info.hasNull && !info.hasUndefined && !info.hasAny && !info.hasUnknown {
			continue
		}
		hasAnyNullableOperand = true
		if info.hasAny || info.hasUnknown {
			strictCheckIsComplete = false
			break
		}
		if info.hasNull && info.hasUndefined {
			strictCheckIsComplete = false
			break
		}
		hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
		hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck
		if hasOnlyNullCheck && !info.hasNull && info.hasUndefined {
			strictCheckIsComplete = false
			break
		}
		if hasOnlyUndefinedCheck && info.hasNull && !info.hasUndefined {
			strictCheckIsComplete = false
			break
		}
	}
	if !hasAnyNullableOperand {
		strictCheckIsComplete = false
	}

	if !hasBothCheck && !firstOpNotNullish && !hasTrailingOptionalChaining && !strictCheckIsComplete {
		hasOnlyNullCheck := hasNullCheck && !hasUndefinedCheck
		hasOnlyUndefinedCheck := !hasNullCheck && hasUndefinedCheck
		if hasOnlyNullCheck || hasOnlyUndefinedCheck {
			return nil
		}
	}

	// Check for incomplete nullish checks on guard operands and truncate if needed
	truncateAt := -1
	for i, op := range chain {
		if i == len(chain)-1 {
			continue
		}
		if op.comparedExpr != nil && processor.containsOptionalChain(op.comparedExpr) {
			continue
		}

		if op.comparedExpr != nil {
			typeInfo := processor.getTypeInfo(op.comparedExpr)
			if typeInfo.hasNull && typeInfo.hasUndefined {
				switch op.typ {
				case OperandTypeNotStrictEqualNull, OperandTypeNotStrictEqualUndef,
					OperandTypeStrictEqualNull, OperandTypeStrictEqualUndef:
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

							if operator == ast.KindEqualsEqualsEqualsToken {
								isStrictNullCheck := binExpr.Right.Kind == ast.KindNullKeyword ||
									binExpr.Left.Kind == ast.KindNullKeyword
								isStrictUndefCheck := (ast.IsIdentifier(binExpr.Right) && binExpr.Right.AsIdentifier().Text == "undefined") ||
									(ast.IsIdentifier(binExpr.Left) && binExpr.Left.AsIdentifier().Text == "undefined") ||
									ast.IsVoidExpression(binExpr.Right) || ast.IsVoidExpression(binExpr.Left)

								if (isStrictNullCheck && !isStrictUndefCheck) || (isStrictUndefCheck && !isStrictNullCheck) {
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

	if len(chain) < 2 {
		return nil
	}

	// Check for === null semantics
	if hasNullCheck && !hasUndefinedCheck && !hasBothCheck && !strictCheckIsComplete {
		return nil
	}

	return chain
}

// validateOrChainForReporting performs OR-chain-specific validation.
// Returns the validated chain (possibly truncated) or nil if the chain should be skipped.
func (processor *chainProcessor) validateOrChainForReporting(chain []Operand) []Operand {
	if len(chain) < 2 {
		return nil
	}

	// Check if all operands in the chain have the same base identifier
	if !processor.hasSameBaseIdentifier(chain) {
		return nil
	}

	// Ensure at least one operand involves property/element/call access
	if !processor.hasPropertyAccessInChain(chain) {
		return nil
	}

	// Ensure at least one operand is an explicit check
	hasExplicitCheck := false
	for _, op := range chain {
		if op.typ != OperandTypePlain {
			hasExplicitCheck = true
			break
		}
	}
	if !hasExplicitCheck {
		return nil
	}

	// Skip optimal strict null checks on non-nullable types
	if processor.shouldSkipOrChainOptimalChecks(chain) {
		return nil
	}

	// Skip OR chains with strict checks and optimal optional chaining
	if len(chain) >= 2 {
		allSubsequentHaveOptionalChaining := true
		for i := 1; i < len(chain); i++ {
			if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
				allSubsequentHaveOptionalChaining = false
				break
			}
		}

		if allSubsequentHaveOptionalChaining {
			firstOp := chain[0]
			isExplicitNullCheck := firstOp.typ == OperandTypeStrictEqualNull ||
				firstOp.typ == OperandTypeEqualNull ||
				isStrictNullComparison(firstOp)
			if isExplicitNullCheck {
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
					return nil
				}
			}
		}

		// Check if all nullish checks are strict
		allStrictChecks := true
		for _, op := range chain {
			if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull ||
				op.typ == OperandTypePlain || op.typ == OperandTypeNot {
				allStrictChecks = false
				break
			}
			if op.typ == OperandTypeComparison && op.node != nil {
				if ast.IsBinaryExpression(op.node) {
					binExpr := op.node.AsBinaryExpression()
					binOp := binExpr.OperatorToken.Kind
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

		if allSubsequentHaveOptionalChaining && allStrictChecks {
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
			if !anyHasNull && !anyHasUndefined {
				return nil
			}
			if anyHasNull && anyHasUndefined {
				return nil
			}
		}
	}

	// Skip simple negation patterns that would change semantics
	if len(chain) >= 2 && chain[0].typ == OperandTypeNot && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := unwrapParentheses(firstExpr)
			isFirstSimpleNegation := !ast.IsPropertyAccessExpression(unwrappedFirst) &&
				!ast.IsElementAccessExpression(unwrappedFirst) &&
				!ast.IsCallExpression(unwrappedFirst)

			if isFirstSimpleNegation {
				allNegatedOrSafeComparisonOrNullCheck := true
				hasIntermediateNullishComp := false
				for i := 1; i < len(chain); i++ {
					isComparison := chain[i].typ == OperandTypeComparison
					isSafeComparison := isComparison && processor.isOrChainComparisonSafe(chain[i])
					isIntermediateNullishComp := isComparison && isNullishComparison(chain[i]) && i < len(chain)-1
					if isIntermediateNullishComp {
						hasIntermediateNullishComp = true
					}
					isAllowedPlainAtEnd := chain[i].typ == OperandTypePlain && i == len(chain)-1 && hasIntermediateNullishComp

					if chain[i].typ != OperandTypeNot && !isSafeComparison && !isIntermediateNullishComp && !isNullishCheckType(chain[i].typ) && !isAllowedPlainAtEnd {
						allNegatedOrSafeComparisonOrNullCheck = false
						break
					}
				}
				if !allNegatedOrSafeComparisonOrNullCheck {
					return nil
				}
			} else {
				for i := 1; i < len(chain); i++ {
					if chain[i].typ == OperandTypeComparison && !processor.isOrChainComparisonSafe(chain[i]) {
						return nil
					}
				}
			}
		}
	}

	// Check OR chains starting with null check for trailing comparison safety
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		if len(chain) >= 2 && (chain[0].typ == OperandTypeNotEqualBoth || chain[0].typ == OperandTypeNotStrictEqualNull || chain[0].typ == OperandTypeNotStrictEqualUndef) {
			firstTypeInfo := processor.getTypeInfo(chain[0].comparedExpr)

			for i := 1; i < len(chain); i++ {
				if chain[i].typ == OperandTypeComparison && !processor.isOrChainComparisonSafe(chain[i]) {
					isLastOperand := i == len(chain)-1
					if isLastOperand && isNullishComparison(chain[i]) {
						if firstTypeInfo.hasNull && firstTypeInfo.hasUndefined {
							return nil
						}
						if firstTypeInfo.hasAny || firstTypeInfo.hasUnknown {
							return nil
						}
					} else if !isNullishComparison(chain[i]) {
						return nil
					}
				}
			}
		}
	}

	// Check if we should apply requireNullish option
	if processor.shouldSkipForRequireNullish(chain, ast.KindBarBarToken) {
		return nil
	}

	// Skip OR chains starting with import.meta
	if len(chain) >= 2 && chain[0].typ == OperandTypePlain {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrapped := unwrapParentheses(firstExpr)
			if unwrapped.Kind == ast.KindMetaProperty {
				return nil
			}
		}
	}

	// Skip plain truthy check with comparison patterns
	if len(chain) >= 2 && chain[0].typ == OperandTypePlain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := unwrapParentheses(firstExpr)
			isFirstSimplePlain := !ast.IsPropertyAccessExpression(unwrappedFirst) &&
				!ast.IsElementAccessExpression(unwrappedFirst) &&
				!ast.IsCallExpression(unwrappedFirst)

			if isFirstSimplePlain {
				for i := 1; i < len(chain); i++ {
					if chain[i].typ == OperandTypeComparison {
						return nil
					}
				}
			}
		}
	}

	// Check for incomplete nullish checks in OR chains
	chain = processor.validateOrChainNullishChecks(chain)
	if chain == nil || len(chain) < 2 {
		return nil
	}

	// Check if conversion would change return type
	for _, op := range chain {
		if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
			if processor.wouldChangeReturnType(op.comparedExpr) && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
				return nil
			}
		}
	}

	// Skip pattern !a.b || a.b() where we negate a property then call it
	if len(chain) >= 2 {
		for i := range len(chain) - 1 {
			if chain[i].typ == OperandTypeNot {
				negatedExpr := chain[i].comparedExpr
				for j := i + 1; j < len(chain); j++ {
					if chain[j].typ == OperandTypeNot {
						continue
					}
					callExpr := chain[j].comparedExpr
					if callExpr != nil && ast.IsCallExpression(unwrapParentheses(callExpr)) {
						call := unwrapParentheses(callExpr).AsCallExpression()
						callBase := call.Expression
						cmp := processor.compareNodes(negatedExpr, callBase)
						if cmp == NodeEqual {
							return nil
						}
					}
				}
			}
		}
	}

	return chain
}

// allOperandsCheckSameExpression returns true if all operands check the same expression.
func (processor *chainProcessor) allOperandsCheckSameExpression(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}
	firstParts := processor.flattenForFix(chain[0].comparedExpr)
	for i := 1; i < len(chain); i++ {
		opParts := processor.flattenForFix(chain[i].comparedExpr)
		if len(opParts) != len(firstParts) {
			return false
		}
		for j := range firstParts {
			if firstParts[j].text != opParts[j].text {
				return false
			}
		}
	}
	return true
}

// isSplitStrictEqualsPattern checks for complementary null+undefined checks on same expression.
func (processor *chainProcessor) isSplitStrictEqualsPattern(chain []Operand) bool {
	if len(chain) != 2 {
		return false
	}
	firstOp := chain[0]
	lastOp := chain[1]
	isFirstUndef := firstOp.typ == OperandTypeNotStrictEqualUndef || firstOp.typ == OperandTypeTypeofCheck
	isFirstNull := firstOp.typ == OperandTypeNotStrictEqualNull
	isLastUndef := lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeTypeofCheck
	isLastNull := lastOp.typ == OperandTypeNotStrictEqualNull
	return (isFirstUndef && isLastNull) || (isFirstNull && isLastUndef)
}

// shouldSkipOptimalStrictChecks checks if the chain uses strict checks with optimal optional chaining.
func (processor *chainProcessor) shouldSkipOptimalStrictChecks(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}

	// Check if ALL operands after the first have optional chaining
	allSubsequentHaveOptionalChaining := true
	for i := 1; i < len(chain); i++ {
		if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
			allSubsequentHaveOptionalChaining = false
			break
		}
	}

	if !allSubsequentHaveOptionalChaining {
		return false
	}

	// Check if ALL nullish checks are strict (ignore Plain operands - they're not nullish checks)
	// We want to skip chains where:
	// - The nullish check operands use strict !== (not loose !=)
	// - Combined with expressions that already have optional chaining
	// This is an "optimal" pattern that shouldn't be converted
	allStrictChecks := true
	hasNullishCheck := false
	for _, op := range chain {
		// Skip Plain operands - they're not nullish checks
		if op.typ == OperandTypePlain {
			continue
		}
		// Any negation (!foo) breaks the pattern
		if op.typ == OperandTypeNot || op.typ == OperandTypeNegatedAndOperand {
			allStrictChecks = false
			break
		}
		// Loose nullish checks break the "all strict" pattern
		if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
			allStrictChecks = false
			break
		}
		// Track that we have at least one nullish check
		if isStrictNullishCheck(op.typ) {
			hasNullishCheck = true
		}
	}

	// Need at least one strict nullish check to be considered "optimal strict"
	if !hasNullishCheck {
		return false
	}

	if !allStrictChecks {
		return false
	}

	// Key logic: Only skip if the type has BOTH null AND undefined
	// When type has both, a strict check (!== null or !== undefined) only covers ONE,
	// so the check is intentional and we shouldn't suggest converting to optional chain.
	//
	// When type has only ONE of null/undefined:
	// - The strict check provides complete coverage for that type
	// - Converting to ?. is safe and preferred
	//
	// When type has neither:
	// - The check is meaningless and should be flagged
	firstOp := chain[0]
	firstTypeInfo := processor.getTypeInfo(firstOp.comparedExpr)

	// Only skip if type has BOTH null AND undefined
	// This means the strict check is intentionally covering only one of them
	if firstTypeInfo.hasNull && firstTypeInfo.hasUndefined {
		return true
	}

	// Type has only one or neither - don't skip, allow reporting
	return false
}

// shouldSkipOrChainOptimalChecks checks if OR chain should be skipped due to optimal structure.
func (processor *chainProcessor) shouldSkipOrChainOptimalChecks(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}

	// Check if ALL operands after the first have optional chaining
	allSubsequentHaveOptionalChaining := true
	for i := 1; i < len(chain); i++ {
		if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
			allSubsequentHaveOptionalChaining = false
			break
		}
	}

	if !allSubsequentHaveOptionalChaining {
		return false
	}

	// Check first operand
	firstOp := chain[0]
	if firstOp.typ != OperandTypeStrictEqualNull {
		return false
	}

	// Check if first operand's type includes undefined
	if firstOp.comparedExpr != nil {
		typeInfo := processor.getTypeInfo(firstOp.comparedExpr)
		// If type includes undefined, the === null check is intentional
		if !typeInfo.hasUndefined && !typeInfo.hasAny && !typeInfo.hasUnknown {
			return true
		}
	}

	return false
}

// generateFixAndReport generates the fix and reports the chain.
// This dispatches to chain-type-specific fix generation.
func (processor *chainProcessor) generateFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node, operatorKind ast.Kind) {
	if isAndOperator(operatorKind) {
		processor.generateAndChainFixAndReport(node, chain, operandNodes)
	} else {
		processor.generateOrChainFixAndReport(node, chain, operandNodes)
	}
}

// generateAndChainFixAndReport generates the fix and reports for AND chains (foo && foo.bar -> foo?.bar).
func (processor *chainProcessor) generateAndChainFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node) {
	// Build the optional chain
	// Find the last actual property access (plain or comparison)
	var lastPropertyAccess *ast.Node
	var hasTrailingComparison bool
	var hasTrailingTypeofCheck bool
	var hasComplementaryNullCheck bool      // true if last two operands form a complementary null+undefined check
	var complementaryTrailingNode *ast.Node // the last operand's node to append as trailing text
	var hasLooseStrictWithTrailingPlain bool
	var looseStrictTrailingPlainNode *ast.Node

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

	// Check for loose+strict transition with trailing Plain
	// Pattern: ... != null && ... !== undefined && <plain access>
	// Example: foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz
	// Expected: foo?.bar?.baz !== undefined && foo.bar.baz.buzz
	// - Convert up to and including the strict check (with optional chaining)
	// - Preserve the strict check comparison syntax
	// - Append the trailing Plain as trailing text
	//
	// Key distinction: This only applies when:
	// 1. The strict check is on a DEEPER expression than the loose check
	// 2. The strict check is NOT part of a complementary pair (no matching null/undefined check on same expr)
	// 3. There is no other check between the strict check and the trailing Plain
	if !hasComplementaryNullCheck && len(chain) >= 3 {
		lastOp := chain[len(chain)-1]
		secondLastOp := chain[len(chain)-2]

		// Check if last operand is Plain and second-to-last is a strict check
		if lastOp.typ == OperandTypePlain {
			isStrictCheck := secondLastOp.typ == OperandTypeNotStrictEqualNull ||
				secondLastOp.typ == OperandTypeNotStrictEqualUndef

			if isStrictCheck && secondLastOp.comparedExpr != nil {
				// Check if the strict check is part of a complementary pair by looking
				// for another strict check on the SAME expression earlier in the chain
				hasMatchingStrictCheck := false
				for i := len(chain) - 3; i >= 0; i-- {
					if chain[i].comparedExpr != nil {
						cmp := processor.compareNodes(chain[i].comparedExpr, secondLastOp.comparedExpr)
						if cmp == NodeEqual {
							// Found a check on the same expression
							// Check if it's a complementary type
							isSecondLastNull := secondLastOp.typ == OperandTypeNotStrictEqualNull
							isSecondLastUndef := secondLastOp.typ == OperandTypeNotStrictEqualUndef
							isOtherNull := chain[i].typ == OperandTypeNotStrictEqualNull
							isOtherUndef := chain[i].typ == OperandTypeNotStrictEqualUndef || chain[i].typ == OperandTypeTypeofCheck

							if (isSecondLastNull && isOtherUndef) || (isSecondLastUndef && isOtherNull) {
								// Complementary pair exists - this is covered
								hasMatchingStrictCheck = true
								break
							}
						}
					}
				}

				if !hasMatchingStrictCheck {
					// No complementary pair - check if there's a loose check on a shallower expression
					var closestLooseCheckExpr *ast.Node
					for i := len(chain) - 3; i >= 0; i-- {
						if chain[i].typ == OperandTypeNotEqualBoth || chain[i].typ == OperandTypeEqualNull {
							closestLooseCheckExpr = chain[i].comparedExpr
							break
						}
					}

					if closestLooseCheckExpr != nil {
						// The strict check must be on a DEEPER expression than the loose check
						// i.e., strict.comparedExpr should be a superset/extension of loose.comparedExpr
						cmp := processor.compareNodes(closestLooseCheckExpr, secondLastOp.comparedExpr)
						if cmp == NodeSubset {
							// Loose is a proper subset of strict - this is the pattern we want
							// e.g., loose on foo.bar, strict on foo.bar.baz
							hasLooseStrictWithTrailingPlain = true
							lastPropertyAccess = secondLastOp.comparedExpr
							looseStrictTrailingPlainNode = lastOp.node
							hasTrailingComparison = true
							hasTrailingTypeofCheck = false
						}
					}
				}
			}
		}
	}

	if !hasComplementaryNullCheck && !hasLooseStrictWithTrailingPlain {
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
		return // Skip this chain
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
						if lastPart.isCall {
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
			if parts[i].hasNonNull {
				parts[i].text = parts[i].baseText()
				parts[i].hasNonNull = false
			}
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
						// Compare base text without ! suffix
						if opParts[i].baseText() != parts[i].baseText() {
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
					// Only add "!" to text if it doesn't already have one
					if !parts[i].hasNonNull {
						parts[i].text = parts[i].text + "!"
					}
					parts[i].hasNonNull = true
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
	if len(parts) > 0 && parts[len(parts)-1].isCall {
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
		// For complementary null+undefined checks or loose+strict with trailing Plain,
		// use the SECOND-TO-LAST operand as the comparison and append the LAST operand as trailing text
		var operandForComparison Operand
		if hasComplementaryNullCheck || hasLooseStrictWithTrailingPlain {
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

		// For loose+strict transition with trailing Plain, append the Plain operand as trailing text
		if hasLooseStrictWithTrailingPlain && looseStrictTrailingPlainNode != nil {
			// Get the text of the last operand including the && before it
			// We need to find the && between the strict check and the plain operand
			secondLastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-2].node)
			lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, looseStrictTrailingPlainNode)
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

	processor.reportChainWithFixes(node, fixes, useSuggestion)

	// Mark all operands in this chain as reported to avoid overlapping diagnostics
	processor.markChainOperandsAsReported(chain)
}

// generateOrChainFixAndReport generates the fix and reports for OR chains (!foo || !foo.bar -> !foo?.bar).
func (processor *chainProcessor) generateOrChainFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node) {
	// Determine hasTrailingComparison based on the last operand
	hasTrailingComparison := false
	if len(chain) > 0 {
		lastOp := chain[len(chain)-1]
		hasTrailingComparison = isComparisonOrNullCheck(lastOp.typ)
	}

	// For || chains with >= 3 operands where the last one is Plain (not negated/checked),
	// we keep it separate to avoid changing semantics.
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
		if isNullishCheckOperand(secondLastOp) && lastOp.comparedExpr != nil && secondLastOp.comparedExpr != nil {
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
	if len(parts) > 0 && parts[len(parts)-1].isCall {
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

	processor.reportChainWithFixes(node, fixes, useSuggestion)

	// Mark all operands in this chain as reported to avoid overlapping diagnostics
	processor.markChainOperandsAsReported(chain)
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
				case ast.KindAmpersandAmpersandToken,
					ast.KindBarBarToken,
					ast.KindQuestionQuestionToken:
					processor.processChain(node, operator)
				}
			},
		}
	},
}
