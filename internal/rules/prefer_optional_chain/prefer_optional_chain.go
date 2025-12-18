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

type OperandType int

const (
	OperandTypeInvalid OperandType = iota
	OperandTypePlain
	OperandTypeNotEqualNull
	OperandTypeNotStrictEqualNull
	OperandTypeNotStrictEqualUndef
	OperandTypeNotEqualBoth
	OperandTypeNot
	OperandTypeNegatedAndOperand
	OperandTypeTypeofCheck
	OperandTypeComparison
	OperandTypeEqualNull
	OperandTypeStrictEqualNull
	OperandTypeStrictEqualUndef
)

func isNullishCheckType(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull ||
		typ == OperandTypeNotStrictEqualUndef ||
		typ == OperandTypeNotEqualBoth ||
		typ == OperandTypeStrictEqualNull ||
		typ == OperandTypeEqualNull ||
		typ == OperandTypeStrictEqualUndef ||
		typ == OperandTypeTypeofCheck
}

func isStrictNullishCheck(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull || typ == OperandTypeNotStrictEqualUndef
}

func isTrailingComparisonType(typ OperandType) bool {
	return typ == OperandTypeNotStrictEqualNull ||
		typ == OperandTypeNotStrictEqualUndef ||
		typ == OperandTypeNotEqualBoth ||
		typ == OperandTypeComparison
}

func isComparisonOrNullCheck(typ OperandType) bool {
	return typ == OperandTypeComparison || isNullishCheckType(typ)
}

// isNullishComparison checks if an OperandTypeComparison is comparing to null/undefined.
// In OR chains, `a.b == null` is parsed as OperandTypeComparison for property accesses
// but should still be treated as a nullish check for chain building.
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

	if binOp != ast.KindEqualsEqualsEqualsToken && binOp != ast.KindEqualsEqualsToken {
		return false
	}

	left := unwrapParentheses(binExpr.Left)
	right := unwrapParentheses(binExpr.Right)

	isLeftNull := left.Kind == ast.KindNullKeyword
	isRightNull := right.Kind == ast.KindNullKeyword

	return isLeftNull || isRightNull
}

// isOrChainNullishCheck returns true if the operand is a nullish check usable in an OR chain.
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

func isNullishCheckOperand(op Operand) bool {
	if op.typ == OperandTypeNotStrictEqualNull ||
		op.typ == OperandTypeNotStrictEqualUndef ||
		op.typ == OperandTypeNotEqualBoth ||
		op.typ == OperandTypeStrictEqualNull ||
		op.typ == OperandTypeStrictEqualUndef ||
		op.typ == OperandTypeEqualNull {
		return true
	}
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
	comparedExpr *ast.Node
}

type NodeComparisonResult int

const (
	NodeEqual NodeComparisonResult = iota
	NodeSubset
	NodeSuperset
	NodeInvalid
)

func isAndOperator(op ast.Kind) bool {
	return op == ast.KindAmpersandAmpersandToken
}

func isOrLikeOperator(op ast.Kind) bool {
	return op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken
}

func unwrapParentheses(n *ast.Node) *ast.Node {
	for ast.IsParenthesizedExpression(n) {
		n = n.AsParenthesizedExpression().Expression
	}
	return n
}

// unwrapForComparison unwraps parentheses, non-null assertions, and type assertions.
// Used for operand comparison where we want foo.bar! to match foo.bar.
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

// getNormalizedNodeText builds a normalized text representation of an AST node.
// Normalization: unwraps parentheses, normalizes ?. to ., strips ! and type assertions.
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

func (processor *chainProcessor) buildNormalizedText(n *ast.Node, result *strings.Builder) {
	if n == nil {
		return
	}

	switch {
	case ast.IsParenthesizedExpression(n):
		processor.buildNormalizedText(n.AsParenthesizedExpression().Expression, result)

	case ast.IsNonNullExpression(n):
		processor.buildNormalizedText(n.AsNonNullExpression().Expression, result)

	case n.Kind == ast.KindAsExpression:
		processor.buildNormalizedText(n.AsAsExpression().Expression, result)

	case n.Kind == ast.KindTypeAssertionExpression:
		processor.buildNormalizedText(n.AsTypeAssertion().Expression, result)

	case ast.IsPropertyAccessExpression(n):
		propAccess := n.AsPropertyAccessExpression()
		processor.buildNormalizedText(propAccess.Expression, result)
		result.WriteByte('.')
		nameRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, propAccess.Name())
		result.WriteString(processor.sourceText[nameRange.Pos():nameRange.End()])

	case ast.IsElementAccessExpression(n):
		elemAccess := n.AsElementAccessExpression()
		processor.buildNormalizedText(elemAccess.Expression, result)
		result.WriteByte('[')
		argRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, elemAccess.ArgumentExpression)
		result.WriteString(processor.sourceText[argRange.Pos():argRange.End()])
		result.WriteByte(']')

	case ast.IsCallExpression(n):
		callExpr := n.AsCallExpression()
		processor.buildNormalizedText(callExpr.Expression, result)
		if callExpr.TypeArguments != nil && len(callExpr.TypeArguments.Nodes) > 0 {
			result.WriteByte('<')
			typeArgsStart := callExpr.TypeArguments.Loc.Pos()
			typeArgsEnd := callExpr.TypeArguments.Loc.End()
			result.WriteString(processor.sourceText[typeArgsStart:typeArgsEnd])
			result.WriteByte('>')
		}
		result.WriteByte('(')
		if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
			argsStart := callExpr.Arguments.Loc.Pos()
			callEnd := n.End()
			result.WriteString(processor.sourceText[argsStart : callEnd-1])
		}
		result.WriteByte(')')

	default:
		textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
		result.WriteString(processor.sourceText[textRange.Pos():textRange.End()])
	}
}

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
// E.g., isChainExtension(foo, foo.bar) -> true, isChainExtension(foo.bar, foo.baz) -> false.
func (processor *chainProcessor) isChainExtension(shorter, longer *ast.Node) bool {
	if shorter == nil || longer == nil {
		return false
	}

	// Get normalized text for both - if shorter is a prefix of longer's normalized text,
	// verify using AST that it's actually a chain extension
	shorterNorm := processor.getNormalizedNodeText(shorter)
	longerNorm := processor.getNormalizedNodeText(longer)

	if shorterNorm == longerNorm {
		return false
	}

	if !strings.HasPrefix(longerNorm, shorterNorm) {
		return false
	}

	// Verify via AST that 'shorter' appears as a base in the chain of 'longer'
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
			return false
		}

		if processor.getNormalizedNodeText(base) == shorterNorm {
			return true
		}

		current = base
	}
}

// unwrapChainNode unwraps parentheses and type assertions to get the underlying chain expression.
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

// isInsideJSX checks if a node is inside a JSX context.
// In JSX, foo && foo.bar has different semantics than foo?.bar:
// foo && foo.bar returns false/null/undefined, while foo?.bar returns undefined.
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

// getBaseIdentifier extracts the base identifier from an expression chain.
// For foo.bar.baz, returns foo. For (foo as any).bar, returns foo.
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
			current = current.AsAsExpression().Expression
		} else {
			return current
		}
	}
}

// hasSideEffects checks if an expression has side effects (++, --, yield, assignment).
func hasSideEffects(node *ast.Node) bool {
	if node == nil {
		return false
	}

	if ast.IsPrefixUnaryExpression(node) {
		op := node.AsPrefixUnaryExpression().Operator
		if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return true
		}
	}

	if node.Kind == ast.KindPostfixUnaryExpression {
		return true
	}

	if ast.IsYieldExpression(node) {
		return true
	}

	// NOTE: Await expressions are NOT checked here for side effects.
	// Await expressions can be safely included in property chains like (await foo).bar.
	// The check for problematic await patterns like "(await foo) && (await foo).bar"
	// is handled separately in compareNodes.

	if ast.IsBinaryExpression(node) {
		op := node.AsBinaryExpression().OperatorToken.Kind
		if op == ast.KindEqualsToken ||
			op == ast.KindPlusEqualsToken ||
			op == ast.KindMinusEqualsToken ||
			op == ast.KindAsteriskEqualsToken ||
			op == ast.KindSlashEqualsToken {
			return true
		}
	}

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

type textRange struct {
	start int
	end   int
}

// ChainPart represents a component of a chain expression for reconstruction.
type ChainPart struct {
	text        string
	optional    bool
	requiresDot bool
	isPrivate   bool
	hasNonNull  bool
	isCall      bool
}

// baseText returns the text without the non-null assertion suffix (!).
func (p ChainPart) baseText() string {
	if p.hasNonNull && len(p.text) > 0 && p.text[len(p.text)-1] == '!' {
		return p.text[:len(p.text)-1]
	}
	return p.text
}

// TypeInfo caches computed type information to avoid repeated type checker calls.
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
	hasBigIntLike    bool
	hasBooleanLike   bool
	hasNumberLike    bool
	hasStringLike    bool
}

// chainProcessor manages state for processing optional chain candidates.
type chainProcessor struct {
	ctx                rule.RuleContext
	opts               PreferOptionalChainOptions
	sourceText         string
	seenLogicals       map[*ast.Node]bool
	processedAndRanges []textRange
	seenLogicalRanges  map[textRange]bool
	reportedRanges     map[textRange]bool
	typeCache          map[*ast.Node]*TypeInfo
	normalizedCache    map[*ast.Node]string
	flattenCache       map[*ast.Node][]ChainPart
	callSigCache       map[*ast.Node]map[string]string
}

func newChainProcessor(ctx rule.RuleContext, opts PreferOptionalChainOptions) *chainProcessor {
	return &chainProcessor{
		ctx:                ctx,
		opts:               opts,
		sourceText:         ctx.SourceFile.Text(),
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

func (processor *chainProcessor) getTypeInfo(node *ast.Node) *TypeInfo {
	if info, ok := processor.typeCache[node]; ok {
		return info
	}

	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	parts := utils.UnionTypeParts(nodeType)

	info := &TypeInfo{
		parts: parts,
	}

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

// extractCallSignatures extracts call signatures from a node.
// Returns a map of "base expression" -> "full call text" for all call expressions.
func (processor *chainProcessor) extractCallSignatures(node *ast.Node) map[string]string {
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

	processor.callSigCache[node] = signatures
	return signatures
}

// validateChainRoot validates that a node is a valid root for chain processing.
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
func (processor *chainProcessor) flattenAndMarkLogicals(node *ast.Node, operatorKind ast.Kind) []*ast.Node {
	unwrapped := unwrapParentheses(node)
	if !ast.IsBinaryExpression(unwrapped) {
		return nil
	}
	binExpr := unwrapped.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != operatorKind {
		return nil
	}

	processor.seenLogicals[node] = true
	processor.seenLogicals[unwrapped] = true

	result := []*ast.Node{node, unwrapped}
	result = append(result, processor.flattenAndMarkLogicals(binExpr.Left, operatorKind)...)
	result = append(result, processor.flattenAndMarkLogicals(binExpr.Right, operatorKind)...)
	return result
}

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

func (processor *chainProcessor) markChainOperandsAsReported(chain []Operand) {
	for _, op := range chain {
		opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, op.node)
		opTextRange := textRange{start: opRange.Pos(), end: opRange.End()}
		processor.reportedRanges[opTextRange] = true
	}
}

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

func isValidOperandForChainType(op Operand, operatorKind ast.Kind) bool {
	if isAndOperator(operatorKind) {
		return op.typ != OperandTypeInvalid
	}

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

// Catches patterns like `foo != null || foo.bar` which have opposite semantics to optional chaining.
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

	isLeftNullish := left.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(left) && left.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(left)
	isRightNullish := right.Kind == ast.KindNullKeyword ||
		(ast.IsIdentifier(right) && right.AsIdentifier().Text == "undefined") ||
		ast.IsVoidExpression(right)

	if !isLeftNullish && !isRightNullish {
		return false
	}

	var checkedExpr *ast.Node
	if isRightNullish {
		checkedExpr = left
	} else {
		checkedExpr = right
	}

	isBaseIdentifier := ast.IsIdentifier(checkedExpr) || checkedExpr.Kind == ast.KindThisKeyword
	return isBaseIdentifier
}

// Allows extending through a call expression when compareNodes returns NodeInvalid.
// AND chains: both operands must be plain truthiness checks.
// OR chains: both must be negations or both must be nullish comparisons.
func (processor *chainProcessor) shouldAllowCallChainExtension(prevOp, currentOp Operand, operatorKind ast.Kind) bool {
	if processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		return true
	}

	if isAndOperator(operatorKind) {
		return prevOp.typ == OperandTypePlain && currentOp.typ == OperandTypePlain
	}

	isNegationChain := prevOp.typ == OperandTypeNot && currentOp.typ == OperandTypeNot
	isNullishComparisonChain := isOrChainNullishCheck(prevOp) && isOrChainNullishCheck(currentOp)
	return isNegationChain || isNullishComparisonChain
}

func (processor *chainProcessor) tryExtendThroughCallExpression(lastExpr *ast.Node, currentOp Operand, firstOpExpr *ast.Node, operatorKind ast.Kind) NodeComparisonResult {
	if lastExpr == nil {
		return NodeInvalid
	}

	lastUnwrapped := lastExpr
	for ast.IsParenthesizedExpression(lastUnwrapped) {
		lastUnwrapped = lastUnwrapped.AsParenthesizedExpression().Expression
	}

	if !ast.IsCallExpression(lastUnwrapped) && !ast.IsNewExpression(lastUnwrapped) {
		return NodeInvalid
	}

	// Each `new X()` creates a fresh instance, so can't chain through it
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

	if processor.isChainExtension(lastExpr, currentOp.comparedExpr) {
		return NodeSubset
	}

	return NodeInvalid
}

func (processor *chainProcessor) compareNodes(left, right *ast.Node) NodeComparisonResult {
	if hasSideEffects(left) || hasSideEffects(right) {
		return NodeInvalid
	}

	leftUnwrapped := unwrapParentheses(left)

	// Block standalone calls (getFoo()) and new expressions (new Date()) since they may have
	// side effects or create different instances. Allow method calls (foo.bar()) and expressions
	// with existing optional chaining. Also block literals ([], {}, functions, classes, JSX)
	// which create new instances each time.
	if !hasOptionalChaining(left) {
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

		isRootedInNew := false
		if rootExpr != nil {
			unwrappedRoot := unwrapParentheses(rootExpr)
			isRootedInNew = ast.IsNewExpression(unwrappedRoot)
		}

		isStandaloneCall := false
		if ast.IsCallExpression(leftUnwrapped) {
			callee := unwrapParentheses(leftUnwrapped.AsCallExpression().Expression)
			isStandaloneCall = ast.IsIdentifier(callee)
		} else if ast.IsNewExpression(leftUnwrapped) {
			isStandaloneCall = true
		}

		if isStandaloneCall || isRootedInNew ||
			ast.IsArrayLiteralExpression(leftUnwrapped) ||
			ast.IsObjectLiteralExpression(leftUnwrapped) ||
			ast.IsFunctionExpression(leftUnwrapped) ||
			ast.IsArrowFunction(leftUnwrapped) ||
			ast.IsClassExpression(leftUnwrapped) ||
			ast.IsJsxElement(leftUnwrapped) ||
			ast.IsJsxSelfClosingElement(leftUnwrapped) ||
			ast.IsJsxFragment(leftUnwrapped) ||
			leftUnwrapped.Kind == ast.KindTemplateExpression ||
			leftUnwrapped.Kind == ast.KindAwaitExpression {
			return NodeInvalid
		}
	}

	leftSigs := processor.extractCallSignatures(left)
	rightSigs := processor.extractCallSignatures(right)

	// Check if any call expressions have matching base but different signatures
	for baseExpr, leftSig := range leftSigs {
		if rightSig, exists := rightSigs[baseExpr]; exists && leftSig != rightSig {
			return NodeInvalid
		}
	}

	leftNormalized := processor.getNormalizedNodeText(left)
	rightNormalized := processor.getNormalizedNodeText(right)

	if leftNormalized == rightNormalized {
		return NodeEqual
	}

	if processor.isChainExtension(left, right) {
		return NodeSubset
	}

	if processor.isChainExtension(right, left) {
		return NodeSuperset
	}

	return NodeInvalid
}

func (processor *chainProcessor) includesNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasNull || info.hasUndefined || info.hasAny || info.hasUnknown
}

// Does NOT return true for 'any' or 'unknown' types.
func (processor *chainProcessor) includesExplicitNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasNull || info.hasUndefined
}

func (processor *chainProcessor) typeIsAnyOrUnknown(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	if len(info.parts) == 0 {
		return false
	}
	for _, part := range info.parts {
		if !utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return false
		}
	}
	return true
}

func (processor *chainProcessor) typeIncludesNull(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasNull || info.hasAny || info.hasUnknown
}

func (processor *chainProcessor) typeIncludesUndefined(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasUndefined || info.hasAny || info.hasUnknown
}

// Returns true when type has falsy non-nullish values (false, 0, "", 0n) but no null/undefined.
func (processor *chainProcessor) wouldChangeReturnType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	hasNullish := info.hasNull || info.hasUndefined
	hasFalsyNonNullish := info.hasBoolLiteral || info.hasNumLiteral || info.hasStrLiteral || info.hasBigIntLiteral
	return hasFalsyNonNullish && !hasNullish
}

// void is always falsy but not nullish, causing issues with && to ?. conversion.
// Example: x && x() where x is void | (() => void) - converting to x?.() would TypeError.
func (processor *chainProcessor) hasVoidType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasVoid
}

// For OR chains (!foo || foo.bar OP VALUE), checks if the comparison is safe to convert.
// Safe: !== X (literals/null), === undefined, == null/undefined, != X (literals only)
// Unsafe: === X (non-undefined), != null/undefined (undefined != null is false in JS!)
func (processor *chainProcessor) isOrChainComparisonSafe(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return true
	}

	unwrapped := op.node
	for ast.IsParenthesizedExpression(unwrapped) {
		unwrapped = unwrapped.AsParenthesizedExpression().Expression
	}

	if !ast.IsBinaryExpression(unwrapped) {
		return true
	}

	binExpr := unwrapped.AsBinaryExpression()
	operator := binExpr.OperatorToken.Kind

	left := binExpr.Left
	right := binExpr.Right

	for ast.IsParenthesizedExpression(left) {
		left = left.AsParenthesizedExpression().Expression
	}
	for ast.IsParenthesizedExpression(right) {
		right = right.AsParenthesizedExpression().Expression
	}

	var value *ast.Node
	if ast.IsPropertyAccessExpression(left) || ast.IsElementAccessExpression(left) || ast.IsCallExpression(left) {
		value = right
	} else if ast.IsPropertyAccessExpression(right) || ast.IsElementAccessExpression(right) || ast.IsCallExpression(right) {
		value = left
	} else {
		return true
	}

	isNull := value.Kind == ast.KindNullKeyword
	isUndefined := (ast.IsIdentifier(value) && value.AsIdentifier().Text == "undefined") || ast.IsVoidExpression(value)
	isLiteral := value.Kind == ast.KindNumericLiteral ||
		value.Kind == ast.KindStringLiteral ||
		value.Kind == ast.KindTrueKeyword ||
		value.Kind == ast.KindFalseKeyword ||
		value.Kind == ast.KindObjectLiteralExpression ||
		value.Kind == ast.KindArrayLiteralExpression

	switch operator {
	case ast.KindExclamationEqualsEqualsToken:
		// !== undefined is NOT safe: undefined !== undefined is false
		return isLiteral || isNull

	case ast.KindEqualsEqualsEqualsToken:
		return isUndefined

	case ast.KindExclamationEqualsToken:
		if isNull || isUndefined {
			return false
		}
		if ast.IsIdentifier(value) && !isLiteral {
			return false
		}
		return isLiteral

	case ast.KindEqualsEqualsToken:
		return isNull || isUndefined
	}

	return true
}

// Checks base identifier's type (e.g., in (foo as any).bar, checks foo's type).
func (processor *chainProcessor) shouldSkipByType(node *ast.Node) bool {
	baseNode := getBaseIdentifier(node)
	info := processor.getTypeInfo(baseNode)

	if processor.opts.RequireNullish && (info.hasNull || info.hasUndefined) {
		return false
	}

	if info.hasAny && !processor.opts.CheckAny {
		return true
	}
	if info.hasBigIntLike && !processor.opts.CheckBigInt {
		return true
	}
	if info.hasBooleanLike && !processor.opts.CheckBoolean {
		return true
	}
	if info.hasNumberLike && !processor.opts.CheckNumber {
		return true
	}
	if info.hasStringLike && !processor.opts.CheckString {
		return true
	}
	if info.hasUnknown && !processor.opts.CheckUnknown {
		return true
	}

	return false
}

func (processor *chainProcessor) flattenForFix(node *ast.Node) []ChainPart {
	if cached, ok := processor.flattenCache[node]; ok {
		return cached
	}

	parts := []ChainPart{}

	var visit func(n *ast.Node, parentIsNonNull bool)
	visit = func(n *ast.Node, parentIsNonNull bool) {
		switch {
		case ast.IsParenthesizedExpression(n):
			inner := n.AsParenthesizedExpression().Expression
			if ast.IsAwaitExpression(inner) || ast.IsYieldExpression(inner) {
				textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
				text := processor.sourceText[textRange.Pos():textRange.End()]

				parts = append(parts, ChainPart{
					text:        text,
					optional:    false,
					requiresDot: false,
				})
				return
			}
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

			typeArgsText := ""
			if callExpr.TypeArguments != nil && len(callExpr.TypeArguments.Nodes) > 0 {
				typeArgsStart := callExpr.TypeArguments.Loc.Pos()
				typeArgsEnd := callExpr.TypeArguments.Loc.End()
				typeArgsText = "<" + processor.sourceText[typeArgsStart:typeArgsEnd] + ">"
			}

			argsText := "()"
			if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
				argsStart := callExpr.Arguments.Loc.Pos()
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
			textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			text := processor.sourceText[textRange.Pos():textRange.End()]

			if parentIsNonNull && ast.IsIdentifier(n) {
				text = text + "!"
			}

			// Type assertions need parentheses when used as base of property access
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

	processor.flattenCache[node] = parts
	return parts
}

// Returns empty string if the chain would result in invalid syntax (e.g., ?.#private).
// stripNonNullAssertions: true for OR chains (strip !), false for AND chains (preserve !).
func (processor *chainProcessor) buildOptionalChain(parts []ChainPart, checkedLengths map[int]bool, callShouldBeOptional bool, stripNonNullAssertions bool) string {
	maxCheckedLength := 0
	for length := range checkedLengths {
		if length > maxCheckedLength {
			maxCheckedLength = length
		}
	}

	optionalParts := make([]bool, len(parts))
	for i, part := range parts {
		if i > 0 {
			if checkedLengths[i] {
				optionalParts[i] = true
			} else if part.optional {
				optionalParts[i] = true
			} else {
				isLastPart := i == len(parts)-1
				if part.isCall && isLastPart && callShouldBeOptional {
					optionalParts[i] = true
				}
			}
		}

		if optionalParts[i] && part.isPrivate {
			return ""
		}
	}

	var result strings.Builder
	for i, part := range parts {
		partText := part.text

		// Strip ! from parts within the checked region when the next part becomes optional (OR chains only)
		if stripNonNullAssertions && i < len(parts)-1 && optionalParts[i+1] && part.hasNonNull {
			if i < maxCheckedLength {
				partText = partText[:len(partText)-1]
			}
		}

		if i > 0 && optionalParts[i] {
			result.WriteString("?.")
		} else if i > 0 {
			// Strip existing ?. from parts within the checked region (the earlier check validated them)
			if part.optional && i > maxCheckedLength {
				result.WriteString("?.")
			} else if part.requiresDot {
				result.WriteString(".")
			}
		}
		result.WriteString(partText)
	}
	return result.String()
}

func (processor *chainProcessor) containsOptionalChain(n *ast.Node) bool {
	unwrapped := unwrapParentheses(n)

	if ast.IsPropertyAccessExpression(unwrapped) {
		if unwrapped.AsPropertyAccessExpression().QuestionDotToken != nil {
			return true
		}
		return processor.containsOptionalChain(unwrapped.AsPropertyAccessExpression().Expression)
	}
	if ast.IsElementAccessExpression(unwrapped) {
		if unwrapped.AsElementAccessExpression().QuestionDotToken != nil {
			return true
		}
		return processor.containsOptionalChain(unwrapped.AsElementAccessExpression().Expression)
	}
	if ast.IsCallExpression(unwrapped) {
		callExpr := unwrapped.AsCallExpression()
		if callExpr.QuestionDotToken != nil {
			return true
		}
		return processor.containsOptionalChain(callExpr.Expression)
	}
	if ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		return processor.containsOptionalChain(binExpr.Left) || processor.containsOptionalChain(binExpr.Right)
	}

	return false
}

func (processor *chainProcessor) parseOperand(node *ast.Node, operatorKind ast.Kind) Operand {
	isAndChain := isAndOperator(operatorKind)
	unwrapped := unwrapForComparison(node)

	// Bare 'this' cannot be converted because it's not nullable in TypeScript.
	// However, this.foo CAN be converted because the property might be nullable.
	if unwrapped.Kind == ast.KindThisKeyword {
		return Operand{typ: OperandTypeInvalid, node: node}
	}

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

	// Skip patterns with nested logical operators at base level (e.g., (x || y) && ...)
	if ast.IsBinaryExpression(baseExpr) {
		binOp := baseExpr.AsBinaryExpression().OperatorToken.Kind
		if binOp == ast.KindAmpersandAmpersandToken || binOp == ast.KindBarBarToken {
			return Operand{typ: OperandTypeInvalid, node: node}
		}
	}

	if ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		op := binExpr.OperatorToken.Kind

		var expr, value *ast.Node

		if binExpr.Right.Kind == ast.KindNullKeyword {
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsIdentifier(binExpr.Right) && binExpr.Right.AsIdentifier().Text == "undefined" {
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsVoidExpression(binExpr.Right) {
			expr = binExpr.Left
			value = binExpr.Right
		} else if ast.IsStringLiteral(binExpr.Right) {
			expr = binExpr.Left
			value = binExpr.Right
		} else if binExpr.Left.Kind == ast.KindNullKeyword {
			// Yoda style
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsIdentifier(binExpr.Left) && binExpr.Left.AsIdentifier().Text == "undefined" {
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsVoidExpression(binExpr.Left) {
			expr = binExpr.Right
			value = binExpr.Left
		} else if ast.IsStringLiteral(binExpr.Left) {
			expr = binExpr.Right
			value = binExpr.Left
		}

		if expr != nil && value != nil {
			expr = unwrapParentheses(expr)

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

			isNull := value.Kind == ast.KindNullKeyword
			isUndefined := (ast.IsIdentifier(value) && value.AsIdentifier().Text == "undefined") || ast.IsVoidExpression(value)

			if isAndChain {
				switch op {
				case ast.KindExclamationEqualsEqualsToken:
					if isNull {
						return Operand{typ: OperandTypeNotStrictEqualNull, node: node, comparedExpr: expr}
					}
					if isUndefined {
						return Operand{typ: OperandTypeNotStrictEqualUndef, node: node, comparedExpr: expr}
					}
				case ast.KindExclamationEqualsToken:
					if isNull || isUndefined {
						return Operand{typ: OperandTypeNotEqualBoth, node: node, comparedExpr: expr}
					}
				case ast.KindEqualsEqualsEqualsToken:
					if isNull {
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
					if (isNull || isUndefined) && (ast.IsIdentifier(expr) || expr.Kind == ast.KindThisKeyword) {
						return Operand{typ: OperandTypeEqualNull, node: node, comparedExpr: expr}
					}
				}
			} else {
				// OR chain: base identifier null checks become optional chain starts,
				// property null checks become comparison operands at chain end
				isPropertyOrElement := ast.IsPropertyAccessExpression(expr) || ast.IsElementAccessExpression(expr) || ast.IsCallExpression(expr)

				if isPropertyOrElement && (isNull || isUndefined) {
					return Operand{typ: OperandTypeComparison, node: node, comparedExpr: expr}
				} else if !isPropertyOrElement {
					switch op {
					case ast.KindEqualsEqualsEqualsToken:
						if isNull {
							return Operand{typ: OperandTypeNotStrictEqualNull, node: node, comparedExpr: expr}
						}
						if isUndefined {
							return Operand{typ: OperandTypeNotStrictEqualUndef, node: node, comparedExpr: expr}
						}
					case ast.KindEqualsEqualsToken:
						if isNull || isUndefined {
							return Operand{typ: OperandTypeNotEqualBoth, node: node, comparedExpr: expr}
						}
					}
				}
			}
		}
	}

	if ast.IsPrefixUnaryExpression(unwrapped) {
		prefixExpr := unwrapped.AsPrefixUnaryExpression()
		if prefixExpr.Operator == ast.KindExclamationToken {
			if prefixExpr.Operand.Kind == ast.KindThisKeyword {
				return Operand{typ: OperandTypeInvalid, node: node}
			}

			if !isAndChain {
				return Operand{typ: OperandTypeNot, node: node, comparedExpr: prefixExpr.Operand}
			}
			return Operand{typ: OperandTypeNegatedAndOperand, node: node, comparedExpr: prefixExpr.Operand}
		}
	}

	if isAndChain && ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()

		// Determine which side is the property being checked (handles yoda style)
		comparedExpr := unwrapParentheses(binExpr.Left)
		hasPropertyAccess := ast.IsPropertyAccessExpression(comparedExpr) ||
			ast.IsElementAccessExpression(comparedExpr) ||
			ast.IsCallExpression(comparedExpr)

		if ast.IsPropertyAccessExpression(binExpr.Right) || ast.IsElementAccessExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
			hasPropertyAccess = true
		} else if ast.IsCallExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
			hasPropertyAccess = true
		}

		if !hasPropertyAccess {
			return Operand{typ: OperandTypeInvalid, node: node}
		}

		return Operand{typ: OperandTypeComparison, node: node, comparedExpr: comparedExpr}
	}

	if !isAndChain && ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		comparedExpr := unwrapParentheses(binExpr.Left)
		if ast.IsPropertyAccessExpression(binExpr.Right) || ast.IsElementAccessExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
		} else if ast.IsCallExpression(binExpr.Right) {
			comparedExpr = unwrapParentheses(binExpr.Right)
		}
		return Operand{typ: OperandTypeComparison, node: node, comparedExpr: comparedExpr}
	}

	if isAndChain {
		if ast.IsBinaryExpression(unwrapped) {
			binExpr := unwrapped.AsBinaryExpression()
			op := binExpr.OperatorToken.Kind

			isComparison := op == ast.KindEqualsEqualsToken ||
				op == ast.KindExclamationEqualsToken ||
				op == ast.KindEqualsEqualsEqualsToken ||
				op == ast.KindExclamationEqualsEqualsToken ||
				op == ast.KindLessThanToken ||
				op == ast.KindGreaterThanToken ||
				op == ast.KindLessThanEqualsToken ||
				op == ast.KindGreaterThanEqualsToken

			if isComparison {
				return Operand{typ: OperandTypeInvalid, node: node}
			}
		}

		return Operand{typ: OperandTypePlain, node: node, comparedExpr: unwrapped}
	}

	if !isAndChain {
		return Operand{typ: OperandTypePlain, node: node, comparedExpr: unwrapped}
	}

	return Operand{typ: OperandTypeInvalid, node: node}
}

func (processor *chainProcessor) collectOperands(node *ast.Node, operatorKind ast.Kind) []*ast.Node {
	operandNodes := []*ast.Node{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == operatorKind {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			processor.seenLogicals[unwrapped] = true
		} else {
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)
	return operandNodes
}

func (processor *chainProcessor) collectOperandsWithRanges(node *ast.Node, operatorKind ast.Kind) ([]*ast.Node, []textRange) {
	operandNodes := []*ast.Node{}
	binaryRanges := []textRange{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == operatorKind {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			binRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, unwrapped)
			binaryRanges = append(binaryRanges, textRange{start: binRange.Pos(), end: binRange.End()})
		} else {
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)
	return operandNodes, binaryRanges
}

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

func (processor *chainProcessor) validateChain(chain []Operand, operatorKind ast.Kind) bool {
	if len(chain) < 2 {
		return false
	}

	if !processor.hasSameBaseIdentifier(chain) {
		return false
	}

	if !processor.hasPropertyAccessInChain(chain) {
		return false
	}

	if processor.shouldSkipForRequireNullish(chain, operatorKind) {
		return false
	}

	return true
}

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

// When requireNullish is true, only convert chains with explicit nullish checks or nullable types.
func (processor *chainProcessor) shouldSkipForRequireNullish(chain []Operand, operatorKind ast.Kind) bool {
	if !processor.opts.RequireNullish {
		return false
	}

	// For OR chains starting with negation, skip entirely
	if !isAndOperator(operatorKind) && len(chain) > 0 && chain[0].typ == OperandTypeNot {
		return true
	}

	for i, op := range chain {
		if op.typ != OperandTypePlain {
			return false
		}
		// For plain && checks, allow if the type explicitly includes null/undefined
		if isAndOperator(operatorKind) && i < len(chain)-1 && op.comparedExpr != nil {
			if processor.includesExplicitNullish(op.comparedExpr) {
				return false
			}
		}
	}
	return true
}

func (processor *chainProcessor) processChain(node *ast.Node, operatorKind ast.Kind) {
	if isOrLikeOperator(operatorKind) {
		processor.handleEmptyObjectPattern(node)
	}

	if operatorKind == ast.KindQuestionQuestionToken {
		return
	}

	_, ok := processor.validateChainRoot(node, operatorKind)
	if !ok {
		return
	}

	if isAndOperator(operatorKind) {
		if processor.isAndChainAlreadySeen(node) {
			return
		}
		processor.markAndChainAsSeen(node)
		_ = processor.flattenAndMarkLogicals(node, operatorKind)
	} else {
		if processor.isOrChainNestedInLargerChain(node) {
			return
		}
		if processor.isOrChainAlreadySeen(node) {
			return
		}
		processor.markOrChainAsSeen(node)
	}

	var operandNodes []*ast.Node
	if isAndOperator(operatorKind) {
		operandNodes = processor.collectOperands(node, operatorKind)
	} else {
		var collectedBinaryRanges []textRange
		operandNodes, collectedBinaryRanges = processor.collectOperandsWithRanges(node, operatorKind)
		for _, r := range collectedBinaryRanges {
			processor.seenLogicalRanges[r] = true
		}
	}

	if len(operandNodes) < 2 {
		return
	}

	if processor.hasAnyOperandBeenReported(operandNodes) {
		return
	}

	operands := make([]Operand, len(operandNodes))
	for i, n := range operandNodes {
		operands[i] = processor.parseOperand(n, operatorKind)
	}

	chains := processor.buildChains(operands, operatorKind)
	chains = processor.filterOverlappingChains(chains)

	for _, chain := range chains {
		if processor.hasChainOverlapWithReported(chain) {
			continue
		}

		validatedChain := processor.validateChainForReporting(chain, operatorKind)
		if validatedChain == nil {
			continue
		}

		processor.generateFixAndReport(node, validatedChain, operandNodes, operatorKind)
	}
}

func (processor *chainProcessor) buildChains(operands []Operand, operatorKind ast.Kind) [][]Operand {
	if isAndOperator(operatorKind) {
		return processor.buildAndChains(operands)
	}
	return processor.buildOrChains(operands)
}

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

		// Invalid operand types that should break the chain
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

	// Post-process: remove trailing duplicate Plain operands
	for i := range allChains {
		chain := allChains[i]
		for len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			secondToLastOp := chain[len(chain)-2]
			if lastOp.typ == OperandTypePlain && secondToLastOp.typ == OperandTypePlain {
				cmp := processor.compareNodes(secondToLastOp.comparedExpr, lastOp.comparedExpr)
				if cmp == NodeEqual {
					chain = chain[:len(chain)-1]
					allChains[i] = chain
					continue
				}
			}
			break
		}
	}

	return allChains
}

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

func (processor *chainProcessor) tryMergeComplementaryPair(op1, op2 Operand, operatorKind ast.Kind) *[]Operand {
	if op1.comparedExpr == nil || op2.comparedExpr == nil {
		return nil
	}

	cmp := processor.compareNodes(op1.comparedExpr, op2.comparedExpr)
	if cmp != NodeEqual {
		return nil
	}

	if isAndOperator(operatorKind) {
		isOp1Null := op1.typ == OperandTypeNotStrictEqualNull
		isOp1Undef := op1.typ == OperandTypeNotStrictEqualUndef || op1.typ == OperandTypeTypeofCheck
		isOp2Null := op2.typ == OperandTypeNotStrictEqualNull
		isOp2Undef := op2.typ == OperandTypeNotStrictEqualUndef || op2.typ == OperandTypeTypeofCheck

		if (isOp1Null && isOp2Undef) || (isOp1Undef && isOp2Null) {
			result := []Operand{op1, op2}
			return &result
		}
	} else {
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

func (processor *chainProcessor) areNullishChecksConsistent(type1, type2 OperandType) bool {
	if type1 == OperandTypeNotEqualBoth || type1 == OperandTypeEqualNull {
		return true
	}
	if type2 == OperandTypeNotEqualBoth || type2 == OperandTypeEqualNull {
		return true
	}

	isType1Null := type1 == OperandTypeNotStrictEqualNull || type1 == OperandTypeStrictEqualNull
	isType1Undef := type1 == OperandTypeNotStrictEqualUndef || type1 == OperandTypeStrictEqualUndef || type1 == OperandTypeTypeofCheck
	isType2Null := type2 == OperandTypeNotStrictEqualNull || type2 == OperandTypeStrictEqualNull
	isType2Undef := type2 == OperandTypeNotStrictEqualUndef || type2 == OperandTypeStrictEqualUndef || type2 == OperandTypeTypeofCheck

	if (isType1Null && isType2Undef) || (isType1Undef && isType2Null) {
		return true
	}

	return (isType1Null && isType2Null) || (isType1Undef && isType2Undef)
}

func (processor *chainProcessor) validateChainForReporting(chain []Operand, operatorKind ast.Kind) []Operand {
	if !processor.validateChain(chain, operatorKind) {
		return nil
	}

	if isAndOperator(operatorKind) {
		return processor.validateAndChainForReporting(chain)
	}
	return processor.validateOrChainForReporting(chain)
}

func (processor *chainProcessor) validateAndChainForReporting(chain []Operand) []Operand {
	// Skip if first operand is Plain but contains optional chaining with different base
	if len(chain) >= 2 {
		firstOp := chain[0]
		if firstOp.typ == OperandTypePlain && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
			return nil
		}
	}

	// Skip chains where first operand has optional chaining AND a strict check (previous partial fix)
	if len(chain) >= 2 {
		firstOp := chain[0]
		if isStrictNullishCheck(firstOp.typ) && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
			if !processor.isSplitStrictEqualsPattern(chain) {
				return nil
			}
		}
	}

	if len(chain) < 2 {
		return nil
	}

	if !processor.hasPropertyAccessInChain(chain) {
		return nil
	}

	if len(chain) >= 2 {
		if processor.allOperandsCheckSameExpression(chain) {
			if !processor.isSplitStrictEqualsPattern(chain) {
				return nil
			}
		}
	}

	if processor.shouldSkipOptimalStrictChecks(chain) {
		return nil
	}

	if len(chain) == 2 {
		firstOp := chain[0]
		secondOp := chain[1]

		if processor.containsOptionalChain(secondOp.comparedExpr) {
			firstParts := processor.flattenForFix(firstOp.node)
			secondParts := processor.flattenForFix(secondOp.node)

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

	if processor.shouldSkipForRequireNullish(chain, ast.KindAmpersandAmpersandToken) {
		return nil
	}

	if len(chain) > 0 && chain[0].typ == OperandTypePlain {
		if chain[0].comparedExpr != nil && processor.hasVoidType(chain[0].comparedExpr) {
			return nil
		}
	}

	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		for _, op := range chain {
			if op.node != nil && ast.IsNonNullExpression(op.node) {
				return nil
			}
		}
	}

	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		if processor.hasIncompleteNullishCheck(chain) {
			return nil
		}
	}

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

	if len(chain) >= 2 {
		lastOp := chain[len(chain)-1]
		if processor.isUnsafeTrailingComparison(chain, lastOp) {
			return nil
		}
	}

	return chain
}

// Optional chaining checks for BOTH null AND undefined, so if the chain only checks
// for one but not both, it's unsafe to convert.
func (processor *chainProcessor) hasIncompleteNullishCheck(chain []Operand) bool {
	hasNullCheck := false
	hasUndefinedCheck := false
	hasBothCheck := false
	hasPlainTruthinessCheck := false

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

	hasTrailingOptionalChaining := false
	if len(chain) >= 2 && len(guardOperands) < len(chain) {
		lastOp := chain[len(chain)-1]
		if lastOp.comparedExpr != nil && processor.containsOptionalChain(lastOp.comparedExpr) {
			hasTrailingOptionalChaining = true
		}
	}

	firstOpNotNullish := false
	if len(guardOperands) > 0 && guardOperands[0].comparedExpr != nil {
		if !processor.includesNullish(guardOperands[0].comparedExpr) {
			firstOpNotNullish = true
		}
	}

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
			return true
		}
	}

	return false
}

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

	if truncateAt > 0 && truncateAt < len(chain) {
		chain = chain[:truncateAt]
	}

	if len(chain) < 2 {
		return nil
	}

	if hasNullCheck && !hasUndefinedCheck && !hasBothCheck && !strictCheckIsComplete {
		return nil
	}

	return chain
}

func (processor *chainProcessor) validateOrChainForReporting(chain []Operand) []Operand {
	if len(chain) < 2 {
		return nil
	}

	if !processor.hasSameBaseIdentifier(chain) {
		return nil
	}

	if !processor.hasPropertyAccessInChain(chain) {
		return nil
	}

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

	if processor.shouldSkipOrChainOptimalChecks(chain) {
		return nil
	}

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

	if processor.shouldSkipForRequireNullish(chain, ast.KindBarBarToken) {
		return nil
	}

	if len(chain) >= 2 && chain[0].typ == OperandTypePlain {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrapped := unwrapParentheses(firstExpr)
			if unwrapped.Kind == ast.KindMetaProperty {
				return nil
			}
		}
	}

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

	chain = processor.validateOrChainNullishChecks(chain)
	if chain == nil || len(chain) < 2 {
		return nil
	}

	for _, op := range chain {
		if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
			if processor.wouldChangeReturnType(op.comparedExpr) && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
				return nil
			}
		}
	}

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

func (processor *chainProcessor) shouldSkipOptimalStrictChecks(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}

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

	// Skip chains where strict !== checks are combined with expressions that already have ?.
	// This is an "optimal" pattern that shouldn't be converted.
	allStrictChecks := true
	hasNullishCheck := false
	for _, op := range chain {
		if op.typ == OperandTypePlain {
			continue
		}
		if op.typ == OperandTypeNot || op.typ == OperandTypeNegatedAndOperand {
			allStrictChecks = false
			break
		}
		if op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeEqualNull {
			allStrictChecks = false
			break
		}
		if isStrictNullishCheck(op.typ) {
			hasNullishCheck = true
		}
	}

	if !hasNullishCheck {
		return false
	}

	if !allStrictChecks {
		return false
	}

	// Only skip if type has BOTH null AND undefined (strict check intentionally covers one)
	firstOp := chain[0]
	firstTypeInfo := processor.getTypeInfo(firstOp.comparedExpr)

	if firstTypeInfo.hasNull && firstTypeInfo.hasUndefined {
		return true
	}

	return false
}

func (processor *chainProcessor) shouldSkipOrChainOptimalChecks(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}

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

	firstOp := chain[0]
	if firstOp.typ != OperandTypeStrictEqualNull {
		return false
	}

	if firstOp.comparedExpr != nil {
		typeInfo := processor.getTypeInfo(firstOp.comparedExpr)
		if !typeInfo.hasUndefined && !typeInfo.hasAny && !typeInfo.hasUnknown {
			return true
		}
	}

	return false
}

func (processor *chainProcessor) generateFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node, operatorKind ast.Kind) {
	if isAndOperator(operatorKind) {
		processor.generateAndChainFixAndReport(node, chain, operandNodes)
	} else {
		processor.generateOrChainFixAndReport(node, chain, operandNodes)
	}
}

// generateAndChainFixAndReport handles AND chains: foo && foo.bar -> foo?.bar
func (processor *chainProcessor) generateAndChainFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node) {
	var lastPropertyAccess *ast.Node
	var hasTrailingComparison bool
	var hasTrailingTypeofCheck bool
	var hasComplementaryNullCheck bool
	var complementaryTrailingNode *ast.Node
	var hasLooseStrictWithTrailingPlain bool
	var looseStrictTrailingPlainNode *ast.Node

	// Check for complementary pair (null + undefined checks on same expression).
	// When true, use second-to-last as chain endpoint and append last as trailing text.
	if len(chain) >= 2 {
		lastOp := chain[len(chain)-1]
		secondLastOp := chain[len(chain)-2]

		if lastOp.comparedExpr != nil && secondLastOp.comparedExpr != nil {
			cmpResult := processor.compareNodes(lastOp.comparedExpr, secondLastOp.comparedExpr)
			if cmpResult == NodeEqual {
				isLastUndef := lastOp.typ == OperandTypeNotStrictEqualUndef || lastOp.typ == OperandTypeTypeofCheck
				isLastNull := lastOp.typ == OperandTypeNotStrictEqualNull
				isSecondLastUndef := secondLastOp.typ == OperandTypeNotStrictEqualUndef || secondLastOp.typ == OperandTypeTypeofCheck
				isSecondLastNull := secondLastOp.typ == OperandTypeNotStrictEqualNull

				if (isLastUndef && isSecondLastNull) || (isLastNull && isSecondLastUndef) {
					hasComplementaryNullCheck = true
					lastPropertyAccess = secondLastOp.comparedExpr
					complementaryTrailingNode = lastOp.node
					hasTrailingComparison = true
					hasTrailingTypeofCheck = secondLastOp.typ == OperandTypeTypeofCheck
				}
			}
		}
	}

	// Loose+strict transition with trailing Plain:
	// foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz
	// -> foo?.bar?.baz !== undefined && foo.bar.baz.buzz
	if !hasComplementaryNullCheck && len(chain) >= 3 {
		lastOp := chain[len(chain)-1]
		secondLastOp := chain[len(chain)-2]

		if lastOp.typ == OperandTypePlain {
			isStrictCheck := secondLastOp.typ == OperandTypeNotStrictEqualNull ||
				secondLastOp.typ == OperandTypeNotStrictEqualUndef

			if isStrictCheck && secondLastOp.comparedExpr != nil {
				// Check if strict check has a complementary pair earlier in the chain
				hasMatchingStrictCheck := false
				for i := len(chain) - 3; i >= 0; i-- {
					if chain[i].comparedExpr != nil {
						cmp := processor.compareNodes(chain[i].comparedExpr, secondLastOp.comparedExpr)
						if cmp == NodeEqual {
							isSecondLastNull := secondLastOp.typ == OperandTypeNotStrictEqualNull
							isSecondLastUndef := secondLastOp.typ == OperandTypeNotStrictEqualUndef
							isOtherNull := chain[i].typ == OperandTypeNotStrictEqualNull
							isOtherUndef := chain[i].typ == OperandTypeNotStrictEqualUndef || chain[i].typ == OperandTypeTypeofCheck

							if (isSecondLastNull && isOtherUndef) || (isSecondLastUndef && isOtherNull) {
								hasMatchingStrictCheck = true
								break
							}
						}
					}
				}

				if !hasMatchingStrictCheck {
					var closestLooseCheckExpr *ast.Node
					for i := len(chain) - 3; i >= 0; i-- {
						if chain[i].typ == OperandTypeNotEqualBoth || chain[i].typ == OperandTypeEqualNull {
							closestLooseCheckExpr = chain[i].comparedExpr
							break
						}
					}

					if closestLooseCheckExpr != nil {
						// Strict check must be on a deeper expression than the loose check
						cmp := processor.compareNodes(closestLooseCheckExpr, secondLastOp.comparedExpr)
						if cmp == NodeSubset {
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
				lastPropertyAccess = chain[i].node
				hasTrailingComparison = false
				hasTrailingTypeofCheck = false
				break
			} else if chain[i].typ == OperandTypeComparison ||
				chain[i].typ == OperandTypeNotStrictEqualNull ||
				chain[i].typ == OperandTypeNotStrictEqualUndef ||
				chain[i].typ == OperandTypeNotEqualBoth {
				lastPropertyAccess = chain[i].comparedExpr
				hasTrailingComparison = true
				hasTrailingTypeofCheck = false
				break
			} else if chain[i].typ == OperandTypeTypeofCheck {
				lastPropertyAccess = chain[i].comparedExpr
				hasTrailingComparison = true
				hasTrailingTypeofCheck = true
				break
			} else if chain[i].comparedExpr != nil {
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

	// For type assertions like (foo as T | null) && (foo as T).bar, use the first operand's
	// more complete type annotation for the base.
	if len(chain) > 0 && len(parts) > 0 && chain[0].typ == OperandTypePlain && chain[0].node != nil {
		firstParts := processor.flattenForFix(chain[0].node)
		if len(firstParts) == 1 && len(firstParts) <= len(parts) {
			if len(firstParts[0].text) > len(parts[0].text) {
				parts[0] = firstParts[0]
			}
		}
	}

	// Map of chain lengths that were explicitly checked, determining which properties get ?.
	checkedLengths := make(map[int]bool)

	checksToConsider := []Operand{}
	for i := range chain {
		op := chain[i]
		isLastOperand := i == len(chain)-1
		isCallAccess := op.comparedExpr != nil && ast.IsCallExpression(op.comparedExpr)

		if isLastOperand && (op.typ == OperandTypePlain || (op.typ == OperandTypeNot && isCallAccess)) {
			continue
		}

		checksToConsider = append(checksToConsider, op)
	}

	hasNonTypeofCheck := false
	for _, operand := range checksToConsider {
		if operand.typ != OperandTypeTypeofCheck && operand.comparedExpr != nil {
			hasNonTypeofCheck = true
			break
		}
	}

	for _, operand := range checksToConsider {
		if operand.comparedExpr != nil {
			// Skip typeof checks when there are other checks - typeof verifies existence,
			// not nullability, so the next property shouldn't be optional when there's
			// a middle guard that does the actual null check.
			if operand.typ == OperandTypeTypeofCheck && hasNonTypeofCheck {
				continue
			}
			checkedParts := processor.flattenForFix(operand.comparedExpr)
			checkedLengths[len(checkedParts)] = true
		}
	}

	// Fill in gaps: for single check at start, fill up to second-to-last part.
	// For multiple checks, use exact check lengths only.
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
			if minChecked == 1 {
				if len(chain) > 0 && chain[len(chain)-1].typ == OperandTypePlain {
					lastPlainParts := processor.flattenForFix(chain[len(chain)-1].node)
					fillUpTo = len(lastPlainParts) - 1

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
			}
		}
	}

	// Replace parts from earlier operands for the checked prefix to preserve existing ?. and !.
	if len(checksToConsider) > 0 && len(parts) > 1 {
		maxCheckedLen := 0
		for _, op := range checksToConsider {
			if op.comparedExpr != nil {
				opParts := processor.flattenForFix(op.comparedExpr)
				if len(opParts) > maxCheckedLen {
					maxCheckedLen = len(opParts)
				}
			}
		}

		// Strip non-null assertions from parts within the checked range
		for i := 0; i < maxCheckedLen && i < len(parts); i++ {
			if parts[i].hasNonNull {
				parts[i].text = parts[i].baseText()
				parts[i].hasNonNull = false
			}
		}

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
				if len(opParts) <= len(parts) {
					isPrefix := true
					for i := range opParts {
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

		// For each part, use the shortest covering operand's optional and non-null flags
		for i := range parts {
			var shortestCoveringOp *opPartsInfo
			for j := range allOpParts {
				op := &allOpParts[j]
				if i < op.len {
					if shortestCoveringOp == nil || op.len < shortestCoveringOp.len {
						shortestCoveringOp = op
					}
				}
			}

			if shortestCoveringOp != nil && i < shortestCoveringOp.len {
				parts[i].optional = shortestCoveringOp.parts[i].optional
			}

			if shortestCoveringOp != nil && i < shortestCoveringOp.len {
				if shortestCoveringOp.parts[i].hasNonNull {
					if !parts[i].hasNonNull {
						parts[i].text = parts[i].text + "!"
					}
					parts[i].hasNonNull = true
				}
			}
		}

	}

	callShouldBeOptional := false
	if len(parts) > 0 && parts[len(parts)-1].isCall {
		partsWithoutCall := len(processor.flattenForFix(lastPropertyAccess.AsCallExpression().Expression))

		for _, op := range chain[:len(chain)-1] {
			if op.comparedExpr != nil {
				checkedParts := processor.flattenForFix(op.comparedExpr)

				if len(checkedParts) == partsWithoutCall {
					callShouldBeOptional = true
					break
				}
			}
		}
	}

	newCode := processor.buildOptionalChain(parts, checkedLengths, callShouldBeOptional, false)

	if newCode == "" {
		return
	}

	// Preserve leading trivia (comments) from operands after the first one.
	if len(chain) > 1 {
		var leadingTrivia strings.Builder
		for i := 1; i < len(chain); i++ {
			opNode := chain[i].node
			if opNode != nil {
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
			triviaStr := strings.TrimLeft(leadingTrivia.String(), " \t\n\r")
			if triviaStr != "" {
				newCode = triviaStr + newCode
			}
		}
	}

	if hasTrailingComparison {
		var operandForComparison Operand
		if hasComplementaryNullCheck || hasLooseStrictWithTrailingPlain {
			operandForComparison = chain[len(chain)-2]
		} else {
			operandForComparison = chain[len(chain)-1]
		}

		if ast.IsBinaryExpression(operandForComparison.node) {
			binExpr := operandForComparison.node.AsBinaryExpression()

			// typeof checks need special wrapping: typeof foo.bar !== 'undefined'
			if hasTrailingTypeofCheck {
				leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)
				comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.comparedExpr)

				typeofPrefix := processor.sourceText[leftRange.Pos():comparedExprRange.Pos()]

				binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).End()
				comparisonSuffix := processor.sourceText[comparedExprRange.End():binExprEnd]

				newCode = typeofPrefix + newCode + comparisonSuffix
			} else {
				// Yoda condition: literal on left (e.g., '123' == foo.bar.baz)
				isYoda := false
				comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.comparedExpr)
				leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)

				if comparedExprRange.Pos() > leftRange.Pos() {
					isYoda = true
				}

				if isYoda {
					binExprStart := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).Pos()
					comparedExprStart := comparedExprRange.Pos()
					yodaPrefix := processor.sourceText[binExprStart:comparedExprStart]
					newCode = yodaPrefix + newCode
				} else {
					comparedExprEnd := comparedExprRange.End()
					binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).End()
					comparisonSuffix := processor.sourceText[comparedExprEnd:binExprEnd]
					newCode = newCode + comparisonSuffix
				}
			}
		}

		if hasComplementaryNullCheck && complementaryTrailingNode != nil {
			secondLastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-2].node)
			lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, complementaryTrailingNode)
			betweenText := processor.sourceText[secondLastRange.End():lastRange.Pos()]
			lastText := processor.sourceText[lastRange.Pos():lastRange.End()]
			newCode = newCode + betweenText + lastText
		}

		if hasLooseStrictWithTrailingPlain && looseStrictTrailingPlainNode != nil {
			secondLastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-2].node)
			lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, looseStrictTrailingPlainNode)
			betweenText := processor.sourceText[secondLastRange.End():lastRange.Pos()]
			lastText := processor.sourceText[lastRange.Pos():lastRange.End()]
			newCode = newCode + betweenText + lastText
		}
	}

	var replaceStart, replaceEnd int

	// Preserve typeof check on undeclared variables to avoid ReferenceError.
	effectiveChainStart := 0
	if len(chain) >= 2 && chain[0].typ == OperandTypeTypeofCheck {
		hasNonTypeofAfterFirst := false
		for i := 1; i < len(chain); i++ {
			if chain[i].typ != OperandTypeTypeofCheck {
				hasNonTypeofAfterFirst = true
				break
			}
		}
		if hasNonTypeofAfterFirst && chain[0].comparedExpr != nil {
			typeofTarget := chain[0].comparedExpr
			symbol := processor.ctx.TypeChecker.GetSymbolAtLocation(typeofTarget)
			isUndeclared := symbol == nil || len(symbol.Declarations) == 0

			if isUndeclared {
				effectiveChainStart = 1
			}
		}
	}

	if effectiveChainStart == 0 && len(chain) == len(operandNodes) {
		nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
		replaceStart = nodeRange.Pos()
		replaceEnd = nodeRange.End()
	} else {
		firstNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[effectiveChainStart].node)
		lastNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-1].node)
		replaceStart = firstNodeRange.Pos()
		replaceEnd = lastNodeRange.End()
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(replaceStart, replaceEnd), newCode),
	}

	// Autofix is safe when: unsafe option enabled, trailing comparison present,
	// operand type includes undefined/any/unknown, or certain nullish comparison types.
	useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

	if useSuggestion && len(chain) > 0 {
		lastOp := chain[len(chain)-1]

		switch lastOp.typ {
		case OperandTypeEqualNull,
			OperandTypeNotEqualBoth,
			OperandTypeStrictEqualUndef,
			OperandTypeNotStrictEqualUndef:
			useSuggestion = false
		case OperandTypeTypeofCheck:
			useSuggestion = false
		}

		if useSuggestion && hasTrailingComparison {
			useSuggestion = false
		}

		// Safe if type includes undefined (since ?. returns undefined)
		if useSuggestion {
			for _, op := range chain {
				if op.comparedExpr != nil {
					info := processor.getTypeInfo(op.comparedExpr)
					if info.hasUndefined || info.hasAny || info.hasUnknown {
						useSuggestion = false
						break
					}
				}
			}
		}

		// Return type changes from "false | T" to "undefined | T" when first operand
		// is explicit nullish check and last is plain access.
		if !useSuggestion && len(chain) > 0 {
			firstOp := chain[0]
			lastOp := chain[len(chain)-1]
			isExplicitNullishCheck := firstOp.typ == OperandTypeNotEqualBoth ||
				firstOp.typ == OperandTypeNotStrictEqualNull ||
				firstOp.typ == OperandTypeNotStrictEqualUndef
			isPlainAccess := lastOp.typ == OperandTypePlain

			if isExplicitNullishCheck && isPlainAccess {
				if firstOp.comparedExpr != nil {
					info := processor.getTypeInfo(firstOp.comparedExpr)
					if !info.hasAny && !info.hasUnknown {
						useSuggestion = true
					}
				}
			}
		}

		// Complementary checks: use suggestion when trailing uses typeof or
		// when the included check is !== null (returns true for undefined).
		if hasComplementaryNullCheck && len(chain) >= 2 {
			lastOp := chain[len(chain)-1]
			secondLastOp := chain[len(chain)-2]
			if lastOp.typ == OperandTypeTypeofCheck {
				useSuggestion = true
			}
			if secondLastOp.typ == OperandTypeNotStrictEqualNull {
				useSuggestion = true
			}
		}
	}

	processor.reportChainWithFixes(node, fixes, useSuggestion)

	processor.markChainOperandsAsReported(chain)
}

// generateOrChainFixAndReport handles OR chains: !foo || !foo.bar -> !foo?.bar
func (processor *chainProcessor) generateOrChainFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node) {
	hasTrailingComparison := false
	if len(chain) > 0 {
		lastOp := chain[len(chain)-1]
		hasTrailingComparison = isComparisonOrNullCheck(lastOp.typ)
	}

	// For >= 3 operand chains where last is Plain, keep it separate to avoid changing semantics.
	// But fully convert when last is negated or when unsafe option is enabled.
	trailingPlainOperand := ""
	chainForOptional := chain
	if len(chain) >= 3 && chain[len(chain)-1].typ == OperandTypePlain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		lastOp := chain[len(chain)-1]
		secondLastOp := chain[len(chain)-2]
		if isNullishCheckOperand(secondLastOp) && lastOp.comparedExpr != nil && secondLastOp.comparedExpr != nil {
			lastParts := processor.flattenForFix(lastOp.comparedExpr)
			secondLastParts := processor.flattenForFix(secondLastOp.comparedExpr)
			if len(lastParts) > len(secondLastParts) {
				lastOpRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOp.node)
				trailingPlainOperand = processor.sourceText[lastOpRange.Pos():lastOpRange.End()]
				chainForOptional = chain[:len(chain)-1]
			}
		}
	}

	// Prevent infinite fix loops when remaining chain already has optional chaining.
	if len(chainForOptional) == 1 && trailingPlainOperand != "" {
		singleOp := chainForOptional[0]
		if singleOp.comparedExpr != nil {
			if hasOptionalChaining(singleOp.comparedExpr) {
				return
			}
		}
	}

	if len(chain) == 2 && trailingPlainOperand == "" {
		firstOp := chain[0]
		if firstOp.comparedExpr != nil && hasOptionalChaining(firstOp.comparedExpr) {
			return
		}
		if firstOp.node != nil && hasOptionalChaining(firstOp.node) {
			return
		}
	}

	lastOp := chainForOptional[len(chainForOptional)-1]
	var lastPropertyAccess *ast.Node
	if lastOp.typ == OperandTypePlain {
		lastPropertyAccess = lastOp.node
	} else {
		lastPropertyAccess = lastOp.comparedExpr
	}
	parts := processor.flattenForFix(lastPropertyAccess)

	checkedLengths := make(map[int]bool)

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

	callShouldBeOptional := false
	if len(parts) > 0 && parts[len(parts)-1].isCall {
		partsWithoutCall := len(parts) - 1
		for _, op := range chainForOptional[:len(chainForOptional)-1] {
			checkedParts := processor.flattenForFix(op.node)
			if len(checkedParts) == partsWithoutCall {
				callShouldBeOptional = true
				break
			}
		}
	}

	optionalChainCode := processor.buildOptionalChain(parts, checkedLengths, callShouldBeOptional, true)

	if optionalChainCode == "" {
		return
	}

	var newCode string
	hasTrailingComparisonForFix := false
	if len(chainForOptional) > 0 {
		lastOpForFix := chainForOptional[len(chainForOptional)-1]
		hasTrailingComparisonForFix = isComparisonOrNullCheck(lastOpForFix.typ)
	}

	if hasTrailingComparisonForFix {
		lastOpForFix := chainForOptional[len(chainForOptional)-1]
		if ast.IsBinaryExpression(lastOpForFix.node) {
			binExpr := lastOpForFix.node.AsBinaryExpression()
			// Normalize Yoda style to non-Yoda
			isYoda := false
			comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOpForFix.comparedExpr)
			leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)

			if comparedExprRange.Pos() > leftRange.Pos() {
				isYoda = true
			}

			if isYoda {
				opRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.OperatorToken)
				opText := processor.sourceText[opRange.Pos():opRange.End()]
				valueText := strings.TrimSpace(processor.sourceText[leftRange.Pos():leftRange.End()])
				newCode = optionalChainCode + " " + opText + " " + valueText
			} else {
				opStart := binExpr.OperatorToken.Pos()
				rightEnd := binExpr.Right.End()
				comparisonText := processor.sourceText[opStart:rightEnd]
				newCode = optionalChainCode + comparisonText
			}
		} else {
			newCode = optionalChainCode
		}
	} else {
		// Add negation if both first and last operands are negated
		firstOpIsNegated := chainForOptional[0].typ == OperandTypeNot
		lastOpIsNegated := chainForOptional[len(chainForOptional)-1].typ == OperandTypeNot

		if firstOpIsNegated && lastOpIsNegated {
			newCode = "!" + optionalChainCode
		} else {
			newCode = optionalChainCode
		}
	}

	if trailingPlainOperand != "" {
		lastChainOp := chain[len(chain)-2]
		trailingOp := chain[len(chain)-1]
		lastChainEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastChainOp.node).End()
		trailingStart := utils.TrimNodeTextRange(processor.ctx.SourceFile, trailingOp.node).Pos()
		separator := processor.sourceText[lastChainEnd:trailingStart]
		newCode = newCode + separator + trailingPlainOperand
	}

	var replaceStart, replaceEnd int
	if len(chain) == len(operandNodes) {
		nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
		replaceStart = nodeRange.Pos()
		replaceEnd = nodeRange.End()
	} else {
		firstNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[0].node)
		lastNodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, chain[len(chain)-1].node)
		replaceStart = firstNodeRange.Pos()
		replaceEnd = lastNodeRange.End()
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(replaceStart, replaceEnd), newCode),
	}

	useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

	if useSuggestion && len(chain) > 0 {
		lastOp := chain[len(chain)-1]

		switch lastOp.typ {
		case OperandTypeNot:
			useSuggestion = false
		case OperandTypeEqualNull,
			OperandTypeNotEqualBoth,
			OperandTypeStrictEqualUndef,
			OperandTypeNotStrictEqualUndef:
			useSuggestion = false
		case OperandTypeTypeofCheck:
			useSuggestion = false
		}
	}

	// For strict null/undefined checks on types that only have one of them,
	// use suggestion because optional chaining checks for BOTH.
	if useSuggestion && hasTrailingComparison {
		strictCheckRequiresSuggestion := false
		if len(chain) > 0 && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
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

			if hasOnlyNullCheck || hasOnlyUndefinedCheck {
				allTypesMatchCheck := true
				hasAnyNullableType := false
				for _, op := range chain {
					if op.comparedExpr == nil {
						continue
					}
					info := processor.getTypeInfo(op.comparedExpr)
					if !info.hasNull && !info.hasUndefined && !info.hasAny && !info.hasUnknown {
						continue
					}
					hasAnyNullableType = true
					if info.hasAny || info.hasUnknown {
						allTypesMatchCheck = false
						break
					}
					if info.hasNull && info.hasUndefined {
						allTypesMatchCheck = false
						break
					}
					if hasOnlyNullCheck && (!info.hasNull || info.hasUndefined) {
						allTypesMatchCheck = false
						break
					}
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

	// Safe if type includes BOTH null AND undefined (or any/unknown)
	if useSuggestion && len(chain) > 0 {
		for _, op := range chain {
			if op.comparedExpr != nil {
				info := processor.getTypeInfo(op.comparedExpr)
				if info.hasAny || info.hasUnknown || (info.hasNull && info.hasUndefined) {
					useSuggestion = false
					break
				}
			}
		}
	}

	processor.reportChainWithFixes(node, fixes, useSuggestion)

	processor.markChainOperandsAsReported(chain)
}

func (processor *chainProcessor) handleEmptyObjectPattern(node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	operator := binExpr.OperatorToken.Kind

	if operator != ast.KindBarBarToken && operator != ast.KindQuestionQuestionToken {
		return
	}

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

	// Only process if the left expression's type includes null/undefined
	if processor.opts.RequireNullish {
		leftExpr := binExpr.Left
		if !processor.includesExplicitNullish(leftExpr) {
			return
		}
	}

	// Pattern: (foo || {}).bar or foo || {}).bar
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

	// Parenthesize complex expressions that need different precedence
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

	// (foo || {}).bar returns {} when foo is falsy, while foo?.bar returns undefined
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
