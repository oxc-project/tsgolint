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
	// OperandTypeNotEqualBoth represents loose inequality (!= null or != undefined),
	// which checks for BOTH null and undefined due to JavaScript coercion.
	OperandTypeNotEqualBoth
	OperandTypeNot
	OperandTypeNegatedAndOperand
	OperandTypeTypeofCheck
	OperandTypeComparison
	// OperandTypeEqualNull represents loose equality (== null or == undefined),
	// which checks for BOTH null and undefined due to JavaScript coercion.
	OperandTypeEqualNull
	OperandTypeStrictEqualNull
	OperandTypeStrictEqualUndef
)

func (t OperandType) IsNullishCheck() bool {
	switch t {
	case OperandTypeNotStrictEqualNull, OperandTypeNotStrictEqualUndef,
		OperandTypeNotEqualBoth, OperandTypeStrictEqualNull,
		OperandTypeEqualNull, OperandTypeStrictEqualUndef, OperandTypeTypeofCheck:
		return true
	}
	return false
}

func (t OperandType) IsStrictNullishCheck() bool {
	return t == OperandTypeNotStrictEqualNull || t == OperandTypeNotStrictEqualUndef
}

// IsLooseNullishCheck returns true for loose equality checks that cover both null and undefined
// due to JavaScript's coercion semantics (== null or != null checks both null and undefined).
func (t OperandType) IsLooseNullishCheck() bool {
	return t == OperandTypeNotEqualBoth || t == OperandTypeEqualNull
}

func (t OperandType) IsTrailingComparison() bool {
	switch t {
	case OperandTypeNotStrictEqualNull, OperandTypeNotStrictEqualUndef,
		OperandTypeNotEqualBoth, OperandTypeComparison:
		return true
	}
	return false
}

func (t OperandType) IsComparisonOrNullCheck() bool {
	return t == OperandTypeComparison || t.IsNullishCheck()
}

func isComparisonAgainst(op Operand, predicate func(*ast.Node) bool) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return false
	}
	bin := unwrapBinary(op.node)
	if bin == nil {
		return false
	}

	if bin.operator != ast.KindEqualsEqualsToken && bin.operator != ast.KindEqualsEqualsEqualsToken {
		return false
	}

	return predicate(bin.left) || predicate(bin.right)
}

func isNullishComparison(op Operand) bool {
	return isComparisonAgainst(op, utils.IsNullishLiteral)
}

func isStrictNullComparison(op Operand) bool {
	return isComparisonAgainst(op, utils.IsNullLiteral)
}

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
	// Use IsNullishCheck for most types, but exclude OperandTypeTypeofCheck
	// which is technically a nullish check but not for this function's purpose
	switch op.typ {
	case OperandTypeNotStrictEqualNull, OperandTypeNotStrictEqualUndef,
		OperandTypeNotEqualBoth, OperandTypeStrictEqualNull,
		OperandTypeStrictEqualUndef, OperandTypeEqualNull:
		return true
	case OperandTypeComparison:
		if op.node == nil {
			return false
		}
		if bin := unwrapBinary(op.node); bin != nil {
			return utils.IsNullLiteral(bin.right) || utils.IsUndefinedIdentifier(bin.right) ||
				utils.IsNullLiteral(bin.left) || utils.IsUndefinedIdentifier(bin.left)
		}
	}
	return false
}

// NullishCheckAnalysis tracks what kinds of null/undefined checks are present in a chain.
// Optional chaining checks for BOTH null AND undefined, so incomplete checks are unsafe.
type NullishCheckAnalysis struct {
	HasNullCheck      bool
	HasUndefinedCheck bool
	HasBothCheck      bool
}

func (a NullishCheckAnalysis) HasOnlyNullCheck() bool {
	return a.HasNullCheck && !a.HasUndefinedCheck && !a.HasBothCheck
}

func (a NullishCheckAnalysis) HasOnlyUndefinedCheck() bool {
	return !a.HasNullCheck && a.HasUndefinedCheck && !a.HasBothCheck
}

func (a NullishCheckAnalysis) HasIncompleteCheck() bool {
	return a.HasOnlyNullCheck() || a.HasOnlyUndefinedCheck()
}

// orChainMode: true for OR chains (=== null), false for AND chains (!== null)
func analyzeNullishChecks(chain []Operand, orChainMode bool) NullishCheckAnalysis {
	analysis := NullishCheckAnalysis{}

	for _, op := range chain {
		switch op.typ {
		case OperandTypeNotStrictEqualNull, OperandTypeStrictEqualNull:
			analysis.HasNullCheck = true
		case OperandTypeNotStrictEqualUndef, OperandTypeStrictEqualUndef:
			analysis.HasUndefinedCheck = true
		case OperandTypeNotEqualBoth, OperandTypeEqualNull:
			analysis.HasBothCheck = true
		case OperandTypeTypeofCheck:
			analysis.HasUndefinedCheck = true
		case OperandTypeNot:
			if orChainMode {
				analysis.HasBothCheck = true
			}
		case OperandTypePlain:
			// Plain truthiness checks only count as both in AND chain mode
			if !orChainMode {
				analysis.HasBothCheck = true
			}
		case OperandTypeComparison:
			if bin := unwrapBinary(op.node); bin != nil {
				isNull := utils.IsNullLiteral(bin.left) || utils.IsNullLiteral(bin.right)
				isUndefined := utils.IsUndefinedLiteral(bin.left) || utils.IsUndefinedLiteral(bin.right)

				if bin.operator == ast.KindEqualsEqualsToken && (isNull || isUndefined) {
					analysis.HasBothCheck = true
				} else if bin.operator == ast.KindEqualsEqualsEqualsToken {
					if isNull {
						analysis.HasNullCheck = true
					}
					if isUndefined {
						analysis.HasUndefinedCheck = true
					}
				}
			}
		}
	}

	return analysis
}

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

// binaryParts holds the components of a binary expression after unwrapping parentheses.
type binaryParts struct {
	expr     *ast.BinaryExpression
	operator ast.Kind
	left     *ast.Node // parentheses-skipped
	right    *ast.Node // parentheses-skipped
}

func unwrapBinary(node *ast.Node) *binaryParts {
	if node == nil {
		return nil
	}
	unwrapped := ast.SkipParentheses(node)
	if !ast.IsBinaryExpression(unwrapped) {
		return nil
	}
	binExpr := unwrapped.AsBinaryExpression()
	return &binaryParts{
		expr:     binExpr,
		operator: binExpr.OperatorToken.Kind,
		left:     ast.SkipParentheses(binExpr.Left),
		right:    ast.SkipParentheses(binExpr.Right),
	}
}

func unwrapForComparison(n *ast.Node) *ast.Node {
	for {
		switch {
		case ast.IsParenthesizedExpression(n):
			n = n.AsParenthesizedExpression().Expression
		case ast.IsNonNullExpression(n):
			n = n.AsNonNullExpression().Expression
		case n.Kind == ast.KindAsExpression:
			n = n.AsAsExpression().Expression
		case n.Kind == ast.KindTypeAssertionExpression:
			n = n.AsTypeAssertion().Expression
		default:
			return n
		}
	}
}

// Ignores parentheses, non-null assertions, type assertions.
func areNodesStructurallyEqual(a, b *ast.Node) bool {
	a = unwrapForComparison(a)
	b = unwrapForComparison(b)

	if a == nil || b == nil {
		return a == b
	}

	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case ast.KindIdentifier:
		return a.Text() == b.Text()

	case ast.KindThisKeyword, ast.KindNullKeyword, ast.KindSuperKeyword:
		return true

	case ast.KindPropertyAccessExpression:
		aProp := a.AsPropertyAccessExpression()
		bProp := b.AsPropertyAccessExpression()
		aName, bName := aProp.Name(), bProp.Name()
		if aName == nil || bName == nil {
			return aName == bName
		}
		return aName.Text() == bName.Text() &&
			areNodesStructurallyEqual(aProp.Expression, bProp.Expression)

	case ast.KindElementAccessExpression:
		aElem := a.AsElementAccessExpression()
		bElem := b.AsElementAccessExpression()
		return areNodesStructurallyEqual(aElem.ArgumentExpression, bElem.ArgumentExpression) &&
			areNodesStructurallyEqual(aElem.Expression, bElem.Expression)

	case ast.KindCallExpression:
		aCall := a.AsCallExpression()
		bCall := b.AsCallExpression()
		if !areNodesStructurallyEqual(aCall.Expression, bCall.Expression) {
			return false
		}
		aArgs := aCall.Arguments
		bArgs := bCall.Arguments
		if aArgs == nil && bArgs == nil {
			return true
		}
		if aArgs == nil || bArgs == nil || len(aArgs.Nodes) != len(bArgs.Nodes) {
			return false
		}
		for i := range aArgs.Nodes {
			if !areNodesStructurallyEqual(aArgs.Nodes[i], bArgs.Nodes[i]) {
				return false
			}
		}
		// Also compare type arguments
		aTypeArgs := aCall.TypeArguments
		bTypeArgs := bCall.TypeArguments
		if aTypeArgs == nil && bTypeArgs == nil {
			return true
		}
		if aTypeArgs == nil || bTypeArgs == nil || len(aTypeArgs.Nodes) != len(bTypeArgs.Nodes) {
			return false
		}
		for i := range aTypeArgs.Nodes {
			if !areNodesStructurallyEqual(aTypeArgs.Nodes[i], bTypeArgs.Nodes[i]) {
				return false
			}
		}
		return true

	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return a.Text() == b.Text()

	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true

	// Type keywords (for type arguments comparison)
	case ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindBooleanKeyword,
		ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindNeverKeyword,
		ast.KindVoidKeyword, ast.KindUndefinedKeyword, ast.KindObjectKeyword,
		ast.KindSymbolKeyword, ast.KindBigIntKeyword:
		return true

	case ast.KindTypeReference:
		aRef := a.AsTypeReferenceNode()
		bRef := b.AsTypeReferenceNode()
		return areNodesStructurallyEqual(aRef.TypeName, bRef.TypeName)

	case ast.KindQualifiedName:
		aQual := a.AsQualifiedName()
		bQual := b.AsQualifiedName()
		return areNodesStructurallyEqual(aQual.Left, bQual.Left) &&
			areNodesStructurallyEqual(aQual.Right, bQual.Right)

	case ast.KindMetaProperty:
		// Handles import.meta and new.target
		aMeta := a.AsMetaProperty()
		bMeta := b.AsMetaProperty()
		aName, bName := aMeta.Name(), bMeta.Name()
		if aName == nil || bName == nil {
			return aName == bName
		}
		return aMeta.KeywordToken == bMeta.KeywordToken && aName.Text() == bName.Text()

	case ast.KindBinaryExpression:
		aBin := a.AsBinaryExpression()
		bBin := b.AsBinaryExpression()
		if aBin.OperatorToken.Kind != bBin.OperatorToken.Kind {
			return false
		}
		return areNodesStructurallyEqual(aBin.Left, bBin.Left) &&
			areNodesStructurallyEqual(aBin.Right, bBin.Right)

	case ast.KindPrefixUnaryExpression:
		aPrefix := a.AsPrefixUnaryExpression()
		bPrefix := b.AsPrefixUnaryExpression()
		if aPrefix.Operator != bPrefix.Operator {
			return false
		}
		return areNodesStructurallyEqual(aPrefix.Operand, bPrefix.Operand)

	case ast.KindTypeOfExpression:
		aTypeof := a.AsTypeOfExpression()
		bTypeof := b.AsTypeOfExpression()
		return areNodesStructurallyEqual(aTypeof.Expression, bTypeof.Expression)

	case ast.KindTemplateExpression:
		aTemplate := a.AsTemplateExpression()
		bTemplate := b.AsTemplateExpression()
		// Compare head
		if aTemplate.Head == nil || bTemplate.Head == nil {
			if aTemplate.Head != bTemplate.Head {
				return false
			}
		} else if aTemplate.Head.Text() != bTemplate.Head.Text() {
			return false
		}
		// Compare template spans
		aSpans := aTemplate.TemplateSpans
		bSpans := bTemplate.TemplateSpans
		if aSpans == nil && bSpans == nil {
			return true
		}
		if aSpans == nil || bSpans == nil || len(aSpans.Nodes) != len(bSpans.Nodes) {
			return false
		}
		for i := range aSpans.Nodes {
			aSpan := aSpans.Nodes[i].AsTemplateSpan()
			bSpan := bSpans.Nodes[i].AsTemplateSpan()
			if !areNodesStructurallyEqual(aSpan.Expression, bSpan.Expression) {
				return false
			}
			if aSpan.Literal == nil || bSpan.Literal == nil {
				if aSpan.Literal != bSpan.Literal {
					return false
				}
			} else if aSpan.Literal.Text() != bSpan.Literal.Text() {
				return false
			}
		}
		return true

	case ast.KindAwaitExpression:
		aAwait := a.AsAwaitExpression()
		bAwait := b.AsAwaitExpression()
		return areNodesStructurallyEqual(aAwait.Expression, bAwait.Expression)
	}

	return false
}

// e.g., foo.bar is a prefix of foo.bar.baz
func isNodePrefixOf(shorter, longer *ast.Node) bool {
	shorter = unwrapForComparison(shorter)
	longer = unwrapForComparison(longer)

	if areNodesStructurallyEqual(shorter, longer) {
		return true
	}

	var base *ast.Node
	switch {
	case ast.IsPropertyAccessExpression(longer):
		base = longer.AsPropertyAccessExpression().Expression
	case ast.IsElementAccessExpression(longer):
		base = longer.AsElementAccessExpression().Expression
	case ast.IsCallExpression(longer):
		base = longer.AsCallExpression().Expression
	case ast.IsNonNullExpression(longer):
		base = longer.AsNonNullExpression().Expression
	default:
		return false
	}

	return isNodePrefixOf(shorter, base)
}

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

func getBaseIdentifier(node *ast.Node) *ast.Node {
	current := node
	for {
		switch {
		case ast.IsPropertyAccessExpression(current):
			current = current.AsPropertyAccessExpression().Expression
		case ast.IsElementAccessExpression(current):
			current = current.AsElementAccessExpression().Expression
		case ast.IsCallExpression(current):
			current = current.AsCallExpression().Expression
		case ast.IsNonNullExpression(current):
			current = current.AsNonNullExpression().Expression
		case ast.IsParenthesizedExpression(current):
			current = current.AsParenthesizedExpression().Expression
		case current.Kind == ast.KindAsExpression:
			current = current.AsAsExpression().Expression
		default:
			return current
		}
	}
}

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

type ChainPart struct {
	text        string
	optional    bool
	requiresDot bool
	isPrivate   bool
	hasNonNull  bool
	isCall      bool
}

func (p ChainPart) baseText() string {
	if p.hasNonNull && len(p.text) > 0 && p.text[len(p.text)-1] == '!' {
		return p.text[:len(p.text)-1]
	}
	return p.text
}

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

func (t *TypeInfo) IsNullish() bool {
	return t.hasNull || t.hasUndefined || t.hasAny || t.hasUnknown
}

// HasExplicitNullish excludes any/unknown (which are implicitly nullish).
func (t *TypeInfo) HasExplicitNullish() bool {
	return t.hasNull || t.hasUndefined
}

func (t *TypeInfo) IsAnyOrUnknown() bool {
	return t.hasAny || t.hasUnknown
}

func (t *TypeInfo) HasBothNullAndUndefined() bool {
	return t.hasNull && t.hasUndefined
}

func (t *TypeInfo) HasNoNullableTypes() bool {
	return !t.hasNull && !t.hasUndefined && !t.hasAny && !t.hasUnknown
}

type chainProcessor struct {
	ctx                rule.RuleContext
	opts               PreferOptionalChainOptions
	sourceText         string
	seenRanges         map[textRange]bool
	typeCache          map[*ast.Node]*TypeInfo
	flattenCache       map[*ast.Node][]ChainPart
	callSigCache       map[*ast.Node]map[string]string
	optionalChainCache map[*ast.Node]bool
}

func newChainProcessor(ctx rule.RuleContext, opts PreferOptionalChainOptions) *chainProcessor {
	return &chainProcessor{
		ctx:                ctx,
		opts:               opts,
		sourceText:         ctx.SourceFile.Text(),
		seenRanges:         make(map[textRange]bool, 16),
		typeCache:          make(map[*ast.Node]*TypeInfo, 32),
		flattenCache:       make(map[*ast.Node][]ChainPart, 16),
		callSigCache:       make(map[*ast.Node]map[string]string, 8),
		optionalChainCache: make(map[*ast.Node]bool, 16),
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
		switch {
		case ast.IsCallExpression(n):
			call := n.AsCallExpression()
			exprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, call.Expression)
			exprText := processor.sourceText[exprRange.Pos():exprRange.End()]
			fullRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			fullText := processor.sourceText[fullRange.Pos():fullRange.End()]
			signatures[exprText] = fullText
			visit(call.Expression)
		case ast.IsPropertyAccessExpression(n):
			visit(n.AsPropertyAccessExpression().Expression)
		case ast.IsElementAccessExpression(n):
			visit(n.AsElementAccessExpression().Expression)
		case ast.IsNonNullExpression(n):
			visit(n.AsNonNullExpression().Expression)
		}
	}
	visit(node)

	processor.callSigCache[node] = signatures
	return signatures
}

func (processor *chainProcessor) validateChainRoot(node *ast.Node, operatorKind ast.Kind) (*ast.BinaryExpression, bool) {
	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != operatorKind {
		return nil, false
	}

	if isInsideJSX(node) {
		return nil, false
	}

	return binExpr, true
}

func (processor *chainProcessor) isChainAlreadySeen(node *ast.Node) bool {
	nodeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
	nodeTextRange := textRange{start: nodeRange.Pos(), end: nodeRange.End()}
	return processor.seenRanges[nodeTextRange]
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

	bin := unwrapBinary(op.node)
	if bin == nil {
		return false
	}

	if bin.operator != ast.KindExclamationEqualsToken && bin.operator != ast.KindExclamationEqualsEqualsToken {
		return false
	}

	isLeftNullish := utils.IsNullishLiteral(bin.left)
	isRightNullish := utils.IsNullishLiteral(bin.right)

	if !isLeftNullish && !isRightNullish {
		return false
	}

	var checkedExpr *ast.Node
	if isRightNullish {
		checkedExpr = bin.left
	} else {
		checkedExpr = bin.right
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

	lastUnwrapped := ast.SkipParentheses(lastExpr)

	if !ast.IsCallExpression(lastUnwrapped) && !ast.IsNewExpression(lastUnwrapped) {
		return NodeInvalid
	}

	// Each `new X()` creates a fresh instance, so can't chain through it
	if isAndOperator(operatorKind) && firstOpExpr != nil {
		baseExpr := firstOpExpr
	outer:
		for baseExpr != nil {
			unwrapped := ast.SkipParentheses(baseExpr)
			switch {
			case ast.IsPropertyAccessExpression(unwrapped):
				baseExpr = unwrapped.AsPropertyAccessExpression().Expression
			case ast.IsElementAccessExpression(unwrapped):
				baseExpr = unwrapped.AsElementAccessExpression().Expression
			case ast.IsCallExpression(unwrapped):
				baseExpr = unwrapped.AsCallExpression().Expression
			default:
				break outer
			}
		}

		if baseExpr != nil && ast.IsNewExpression(ast.SkipParentheses(baseExpr)) {
			return NodeInvalid
		}
	}

	if isNodePrefixOf(lastExpr, currentOp.comparedExpr) {
		return NodeSubset
	}

	return NodeInvalid
}

func (processor *chainProcessor) compareNodes(left, right *ast.Node) NodeComparisonResult {
	if hasSideEffects(left) || hasSideEffects(right) {
		return NodeInvalid
	}

	leftUnwrapped := ast.SkipParentheses(left)

	// Block standalone calls (getFoo()) and new expressions (new Date()) since they may have
	// side effects or create different instances. Allow method calls (foo.bar()) and expressions
	// with existing optional chaining. Also block literals ([], {}, functions, classes, JSX)
	// which create new instances each time.
	if !processor.containsOptionalChain(left) {
		rootExpr := leftUnwrapped
	outer2:
		for {
			unwrapped := ast.SkipParentheses(rootExpr)
			switch {
			case ast.IsPropertyAccessExpression(unwrapped):
				rootExpr = unwrapped.AsPropertyAccessExpression().Expression
			case ast.IsElementAccessExpression(unwrapped):
				rootExpr = unwrapped.AsElementAccessExpression().Expression
			case ast.IsCallExpression(unwrapped):
				rootExpr = unwrapped.AsCallExpression().Expression
			default:
				break outer2
			}
		}

		isRootedInNew := false
		if rootExpr != nil {
			unwrappedRoot := ast.SkipParentheses(rootExpr)
			isRootedInNew = ast.IsNewExpression(unwrappedRoot)
		}

		isStandaloneCall := false
		if ast.IsCallExpression(leftUnwrapped) {
			callee := ast.SkipParentheses(leftUnwrapped.AsCallExpression().Expression)
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

	for baseExpr, leftSig := range leftSigs {
		if rightSig, exists := rightSigs[baseExpr]; exists && leftSig != rightSig {
			return NodeInvalid
		}
	}

	if areNodesStructurallyEqual(left, right) {
		return NodeEqual
	}

	if isNodePrefixOf(left, right) {
		return NodeSubset
	}

	if isNodePrefixOf(right, left) {
		return NodeSuperset
	}

	return NodeInvalid
}

func (processor *chainProcessor) includesNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.IsNullish()
}

func (processor *chainProcessor) includesExplicitNullish(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.HasExplicitNullish()
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

func (processor *chainProcessor) wouldChangeReturnType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	hasFalsyNonNullish := info.hasBoolLiteral || info.hasNumLiteral || info.hasStrLiteral || info.hasBigIntLiteral
	return hasFalsyNonNullish && !info.HasExplicitNullish()
}

// void is falsy but not nullish - x?.() on void would TypeError.
func (processor *chainProcessor) hasVoidType(node *ast.Node) bool {
	info := processor.getTypeInfo(node)
	return info.hasVoid
}

// Unsafe: === X (non-undefined), != null/undefined (since undefined != null is false!)
func (processor *chainProcessor) isOrChainComparisonSafe(op Operand) bool {
	if op.typ != OperandTypeComparison || op.node == nil {
		return true
	}

	bin := unwrapBinary(op.node)
	if bin == nil {
		return true
	}

	isAccessExpr := utils.IsAccessExpression

	var value *ast.Node
	switch {
	case isAccessExpr(bin.left):
		value = bin.right
	case isAccessExpr(bin.right):
		value = bin.left
	default:
		return true
	}

	isNull := utils.IsNullLiteral(value)
	isUndefined := utils.IsUndefinedLiteral(value)
	isLiteral := value.Kind == ast.KindNumericLiteral ||
		value.Kind == ast.KindStringLiteral ||
		value.Kind == ast.KindTrueKeyword ||
		value.Kind == ast.KindFalseKeyword ||
		value.Kind == ast.KindObjectLiteralExpression ||
		value.Kind == ast.KindArrayLiteralExpression

	switch bin.operator {
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

func (processor *chainProcessor) shouldSkipByType(node *ast.Node) bool {
	baseNode := getBaseIdentifier(node)
	info := processor.getTypeInfo(baseNode)

	if processor.opts.RequireNullish && info.HasExplicitNullish() {
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
			nonNullExpr := n.AsNonNullExpression()
			visit(nonNullExpr.Expression, true)

		case ast.IsPropertyAccessExpression(n):
			propAccess := n.AsPropertyAccessExpression()
			visit(propAccess.Expression, false)
			nameRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, propAccess.Name())
			nameText := processor.sourceText[nameRange.Pos():nameRange.End()]

			hasNonNull := parentIsNonNull
			if hasNonNull {
				nameText += "!"
			}

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
				text += "!"
			}

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
			switch {
			case checkedLengths[i]:
				optionalParts[i] = true
			case part.optional:
				optionalParts[i] = true
			default:
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

		if stripNonNullAssertions && i < len(parts)-1 && optionalParts[i+1] && part.hasNonNull {
			if i < maxCheckedLength {
				partText = partText[:len(partText)-1]
			}
		}

		if i > 0 && optionalParts[i] {
			result.WriteString("?.")
		} else if i > 0 {
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
	if n == nil {
		return false
	}

	// Check cache first
	if cached, ok := processor.optionalChainCache[n]; ok {
		return cached
	}

	result := processor.containsOptionalChainUncached(n)
	processor.optionalChainCache[n] = result
	return result
}

func (processor *chainProcessor) containsOptionalChainUncached(n *ast.Node) bool {
	unwrapped := ast.SkipParentheses(n)

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

// allSubsequentHaveOptionalChaining checks whether all operands in the chain
// (starting from index 1) already have optional chaining.
// This is used to determine if we should skip transformation in certain cases.
func (processor *chainProcessor) allSubsequentHaveOptionalChaining(chain []Operand) bool {
	for i := 1; i < len(chain); i++ {
		if chain[i].comparedExpr != nil && !processor.containsOptionalChain(chain[i].comparedExpr) {
			return false
		}
	}
	return true
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
outer3:
	for {
		switch {
		case ast.IsPropertyAccessExpression(baseExpr):
			baseExpr = baseExpr.AsPropertyAccessExpression().Expression
		case ast.IsElementAccessExpression(baseExpr):
			baseExpr = baseExpr.AsElementAccessExpression().Expression
		case ast.IsCallExpression(baseExpr):
			baseExpr = baseExpr.AsCallExpression().Expression
		case ast.IsNonNullExpression(baseExpr):
			baseExpr = baseExpr.AsNonNullExpression().Expression
		case ast.IsParenthesizedExpression(baseExpr):
			baseExpr = baseExpr.AsParenthesizedExpression().Expression
		case baseExpr.Kind == ast.KindAsExpression:
			baseExpr = baseExpr.AsAsExpression().Expression
		case baseExpr.Kind == ast.KindTypeAssertionExpression:
			baseExpr = baseExpr.AsTypeAssertion().Expression
		default:
			break outer3
		}
	}

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

		if utils.IsNullishLiteral(binExpr.Right) || ast.IsStringLiteral(binExpr.Right) {
			expr = binExpr.Left
			value = binExpr.Right
		} else if utils.IsNullishLiteral(binExpr.Left) || ast.IsStringLiteral(binExpr.Left) {
			expr = binExpr.Right
			value = binExpr.Left
		}

		if expr != nil && value != nil {
			expr = ast.SkipParentheses(expr)

			if ast.IsTypeOfExpression(expr) {
				typeofExpr := expr.AsTypeOfExpression()
				if ast.IsStringLiteral(value) && value.AsStringLiteral().Text == "undefined" {
					if (op == ast.KindExclamationEqualsEqualsToken || op == ast.KindExclamationEqualsToken) && isAndChain {
						return Operand{typ: OperandTypeTypeofCheck, node: node, comparedExpr: typeofExpr.Expression}
					}
					if (op == ast.KindEqualsEqualsEqualsToken || op == ast.KindEqualsEqualsToken) && !isAndChain {
						return Operand{typ: OperandTypeTypeofCheck, node: node, comparedExpr: typeofExpr.Expression}
					}
				}
			}

			isNull := utils.IsNullLiteral(value)
			isUndefined := utils.IsUndefinedLiteral(value)

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
				isPropertyOrElement := utils.IsAccessExpression(expr)

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

		left := ast.SkipParentheses(binExpr.Left)
		right := ast.SkipParentheses(binExpr.Right)
		leftIsAccess := utils.IsAccessExpression(left)
		rightIsAccess := utils.IsAccessExpression(right)

		// If both sides are access expressions with the same base, we can't convert
		// to optional chaining (e.g., foo && foo.bar === foo.baz)
		if leftIsAccess && rightIsAccess {
			leftBase := getBaseIdentifier(left)
			rightBase := getBaseIdentifier(right)
			if areNodesStructurallyEqual(leftBase, rightBase) {
				return Operand{typ: OperandTypeInvalid, node: node}
			}
		}

		comparedExpr := left
		hasPropertyAccess := leftIsAccess

		if rightIsAccess {
			comparedExpr = right
			hasPropertyAccess = true
		}

		if !hasPropertyAccess {
			return Operand{typ: OperandTypeInvalid, node: node}
		}

		return Operand{typ: OperandTypeComparison, node: node, comparedExpr: comparedExpr}
	}

	if !isAndChain && ast.IsBinaryExpression(unwrapped) {
		binExpr := unwrapped.AsBinaryExpression()
		left := ast.SkipParentheses(binExpr.Left)
		right := ast.SkipParentheses(binExpr.Right)
		leftIsAccess := utils.IsAccessExpression(left)
		rightIsAccess := utils.IsAccessExpression(right)

		// If both sides are access expressions with the same base, we can't convert
		// to optional chaining (e.g., foo == null || foo.bar === foo.baz)
		if leftIsAccess && rightIsAccess {
			leftBase := getBaseIdentifier(left)
			rightBase := getBaseIdentifier(right)
			if areNodesStructurallyEqual(leftBase, rightBase) {
				return Operand{typ: OperandTypeInvalid, node: node}
			}
		}

		comparedExpr := left
		if rightIsAccess {
			comparedExpr = right
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

	return Operand{typ: OperandTypePlain, node: node, comparedExpr: unwrapped}
}

func (processor *chainProcessor) collectOperands(node *ast.Node, operatorKind ast.Kind) []*ast.Node {
	operandNodes := []*ast.Node{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		unwrapped := ast.SkipParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == operatorKind {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			// Mark this binary expression as seen to prevent re-processing
			binRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, unwrapped)
			processor.seenRanges[textRange{start: binRange.Pos(), end: binRange.End()}] = true
		} else {
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)
	return operandNodes
}

func (processor *chainProcessor) hasPropertyAccessInChain(chain []Operand) bool {
	for _, op := range chain {
		if op.comparedExpr != nil {
			unwrapped := ast.SkipParentheses(op.comparedExpr)
			if utils.IsAccessExpression(unwrapped) {
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
		} else if !areNodesStructurallyEqual(firstBase, base) {
			return false
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

func (processor *chainProcessor) shouldSkipForRequireNullish(chain []Operand, operatorKind ast.Kind) bool {
	if !processor.opts.RequireNullish {
		return false
	}

	if !isAndOperator(operatorKind) && len(chain) > 0 && chain[0].typ == OperandTypeNot {
		return true
	}

	for i, op := range chain {
		if op.typ != OperandTypePlain {
			return false
		}
		if isAndOperator(operatorKind) && i < len(chain)-1 && op.comparedExpr != nil {
			if processor.includesExplicitNullish(op.comparedExpr) {
				return false
			}
		}
	}
	return true
}

func (processor *chainProcessor) processChain(node *ast.Node, operatorKind ast.Kind) {
	if operatorKind == ast.KindBarBarToken || operatorKind == ast.KindQuestionQuestionToken {
		processor.handleEmptyObjectPattern(node)
	}

	if operatorKind == ast.KindQuestionQuestionToken {
		return
	}

	_, ok := processor.validateChainRoot(node, operatorKind)
	if !ok {
		return
	}

	if processor.isChainAlreadySeen(node) {
		return
	}

	operandNodes := processor.collectOperands(node, operatorKind)

	if len(operandNodes) < 2 {
		return
	}

	operands := make([]Operand, len(operandNodes))
	for i, n := range operandNodes {
		operands[i] = processor.parseOperand(n, operatorKind)
	}

	chains := processor.buildChains(operands, operatorKind)

	for _, chain := range chains {
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

	i := 0
	for i < len(operands) {
		op := operands[i]

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

		if cmp == NodeInvalid && len(currentChain) > 0 {
			prevOp := currentChain[len(currentChain)-1]
			if processor.shouldAllowCallChainExtension(prevOp, op, ast.KindAmpersandAmpersandToken) {
				firstOpExpr := currentChain[0].comparedExpr
				if extCmp := processor.tryExtendThroughCallExpression(lastExpr, op, firstOpExpr, ast.KindAmpersandAmpersandToken); extCmp == NodeSubset {
					cmp = NodeSubset
				}
			}
		}

		if cmp == NodeEqual && i+1 < len(operands) {
			if pair := processor.tryMergeComplementaryPair(op, operands[i+1], ast.KindAmpersandAmpersandToken); pair != nil {
				currentChain = append(currentChain, pair...)
				lastExpr = pair[len(pair)-1].comparedExpr
				i += 2
				continue
			}
		}

		if op.typ.IsNullishCheck() && lastCheckType != OperandTypeInvalid {
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
			if op.typ != OperandTypePlain && op.typ.IsNullishCheck() {
				lastCheckType = op.typ
			}
			i++
			continue
		}

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

	if len(currentChain) >= 2 {
		allChains = append(allChains, currentChain)
	}

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
	if !prevOp.typ.IsStrictNullishCheck() || prevOp.comparedExpr == nil {
		return false
	}

	prevUnwrapped := ast.SkipParentheses(prevOp.comparedExpr)

	isCallOrNew := ast.IsCallExpression(prevUnwrapped) || ast.IsNewExpression(prevUnwrapped)
	isElementAccess := ast.IsElementAccessExpression(prevUnwrapped)

	if !isCallOrNew && !isElementAccess {
		return false
	}

	isAnyOrUnknown := processor.typeIsAnyOrUnknown(prevOp.comparedExpr)
	typeInfo := processor.getTypeInfo(prevOp.comparedExpr)
	hasNull := typeInfo.hasNull || typeInfo.hasAny || typeInfo.hasUnknown
	hasUndefined := typeInfo.hasUndefined || typeInfo.hasAny || typeInfo.hasUnknown

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

func (processor *chainProcessor) tryMergeComplementaryPair(op1, op2 Operand, operatorKind ast.Kind) []Operand {
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
			return []Operand{op1, op2}
		}
	} else {
		isOp1Null := op1.typ == OperandTypeStrictEqualNull
		isOp1Undef := op1.typ == OperandTypeStrictEqualUndef
		isOp2Null := op2.typ == OperandTypeStrictEqualNull
		isOp2Undef := op2.typ == OperandTypeStrictEqualUndef

		if (isOp1Null && isOp2Undef) || (isOp1Undef && isOp2Null) {
			return []Operand{op1, op2}
		}
	}

	return nil
}

func (processor *chainProcessor) areNullishChecksConsistent(type1, type2 OperandType) bool {
	if type1.IsLooseNullishCheck() {
		return true
	}
	if type2.IsLooseNullishCheck() {
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
	// Note: chain length >= 2, hasSameBaseIdentifier, and hasPropertyAccessInChain are already
	// validated by validateChain() which is called before this function.

	firstOp := chain[0]
	if firstOp.typ == OperandTypePlain && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
		return nil
	}

	if firstOp.typ.IsStrictNullishCheck() && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
		if !processor.isSplitStrictEqualsPattern(chain) {
			return nil
		}
	}

	if processor.allOperandsCheckSameExpression(chain) {
		if !processor.isSplitStrictEqualsPattern(chain) {
			return nil
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
					unwrapped := ast.SkipParentheses(lastOp.comparedExpr)
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
	guardOperands := chain
	if len(chain) >= 2 {
		lastOp := chain[len(chain)-1]
		if lastOp.typ.IsTrailingComparison() {
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

	analysis := analyzeNullishChecks(guardOperands, false)

	hasTypeofCheck := false
	for _, op := range guardOperands {
		if op.typ == OperandTypeTypeofCheck {
			hasTypeofCheck = true
			break
		}
	}

	hasTrailingBothCheck := false
	if len(chain) >= 2 && len(guardOperands) < len(chain) {
		lastOp := chain[len(chain)-1]
		if lastOp.typ == OperandTypeNotEqualBoth {
			hasTrailingBothCheck = true
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
		if !info.IsAnyOrUnknown() {
			hasOnlyNullCheck := analysis.HasNullCheck && !analysis.HasUndefinedCheck
			hasOnlyUndefinedCheck := !analysis.HasNullCheck && analysis.HasUndefinedCheck

			if hasOnlyNullCheck && info.hasNull && !info.hasUndefined {
				strictCheckIsComplete = true
			}
			if hasOnlyUndefinedCheck && info.hasUndefined && !info.hasNull {
				strictCheckIsComplete = true
			}
		}
	}

	hasPlainTruthinessCheck := analysis.HasBothCheck

	if !hasPlainTruthinessCheck && !hasTypeofCheck && !hasTrailingBothCheck && !hasTrailingOptionalChaining && !firstOpNotNullish && !strictCheckIsComplete {
		if analysis.HasIncompleteCheck() {
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
		unwrappedNode := ast.SkipParentheses(lastOp.node)
		if ast.IsBinaryExpression(unwrappedNode) {
			binExpr := unwrappedNode.AsBinaryExpression()
			op := binExpr.OperatorToken.Kind

			var value *ast.Node
			if utils.IsAccessExpression(binExpr.Left) {
				value = binExpr.Right
			} else if utils.IsAccessExpression(binExpr.Right) {
				value = binExpr.Left
			}

			if value != nil {
				isNull := utils.IsNullLiteral(value)
				isUndefined := utils.IsUndefinedLiteral(value)
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
				case ast.KindEqualsEqualsToken:
					if isNullish || isUndeclaredVar {
						unsafe = true
					}
				case ast.KindEqualsEqualsEqualsToken:
					if isUndefined || isUndeclaredVar {
						unsafe = true
					}
				case ast.KindExclamationEqualsToken:
					if !isNullish {
						unsafe = true
					}
				case ast.KindExclamationEqualsEqualsToken:
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

	analysis := analyzeNullishChecks(chain, true)

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
		if info.HasNoNullableTypes() {
			continue
		}
		hasAnyNullableOperand = true
		if info.IsAnyOrUnknown() {
			strictCheckIsComplete = false
			break
		}
		if info.HasBothNullAndUndefined() {
			strictCheckIsComplete = false
			break
		}
		if analysis.HasOnlyNullCheck() && !info.hasNull && info.hasUndefined {
			strictCheckIsComplete = false
			break
		}
		if analysis.HasOnlyUndefinedCheck() && info.hasNull && !info.hasUndefined {
			strictCheckIsComplete = false
			break
		}
	}
	if !hasAnyNullableOperand {
		strictCheckIsComplete = false
	}

	if !analysis.HasBothCheck && !firstOpNotNullish && !hasTrailingOptionalChaining && !strictCheckIsComplete {
		if analysis.HasIncompleteCheck() {
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
						unwrapped := ast.SkipParentheses(op.node)
						if ast.IsBinaryExpression(unwrapped) {
							binExpr := unwrapped.AsBinaryExpression()
							operator := binExpr.OperatorToken.Kind

							if operator == ast.KindEqualsEqualsEqualsToken {
								isStrictNullCheck := utils.IsNullLiteral(binExpr.Right) || utils.IsNullLiteral(binExpr.Left)
								isStrictUndefCheck := utils.IsUndefinedLiteral(binExpr.Right) || utils.IsUndefinedLiteral(binExpr.Left)

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

	if analysis.HasNullCheck && !analysis.HasUndefinedCheck && !analysis.HasBothCheck && !strictCheckIsComplete {
		return nil
	}

	return chain
}

func (processor *chainProcessor) validateOrChainForReporting(chain []Operand) []Operand {
	// Note: chain length >= 2, hasSameBaseIdentifier, and hasPropertyAccessInChain are already
	// validated by validateChain() which is called before this function.

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

	allSubsequentHaveOptionalChaining := processor.allSubsequentHaveOptionalChaining(chain)
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
		if op.typ.IsLooseNullishCheck() ||
			op.typ == OperandTypePlain || op.typ == OperandTypeNot {
			allStrictChecks = false
			break
		}
		if op.typ == OperandTypeComparison && op.node != nil {
			if ast.IsBinaryExpression(op.node) {
				binExpr := op.node.AsBinaryExpression()
				binOp := binExpr.OperatorToken.Kind
				if binOp == ast.KindEqualsEqualsToken {
					left := ast.SkipParentheses(binExpr.Left)
					right := ast.SkipParentheses(binExpr.Right)
					isNullish := utils.IsNullLiteral(left) || utils.IsNullLiteral(right) ||
						utils.IsUndefinedIdentifier(left) || utils.IsUndefinedIdentifier(right)
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

	if chain[0].typ == OperandTypeNot && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := ast.SkipParentheses(firstExpr)
			isFirstSimpleNegation := !utils.IsAccessExpression(unwrappedFirst)

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

					if chain[i].typ != OperandTypeNot && !isSafeComparison && !isIntermediateNullishComp && !chain[i].typ.IsNullishCheck() && !isAllowedPlainAtEnd {
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
		if chain[0].typ == OperandTypeNotEqualBoth || chain[0].typ == OperandTypeNotStrictEqualNull || chain[0].typ == OperandTypeNotStrictEqualUndef {
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

	if chain[0].typ == OperandTypePlain {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrapped := ast.SkipParentheses(firstExpr)
			if unwrapped.Kind == ast.KindMetaProperty {
				return nil
			}
		}
	}

	if chain[0].typ == OperandTypePlain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		firstExpr := chain[0].comparedExpr
		if firstExpr != nil {
			unwrappedFirst := ast.SkipParentheses(firstExpr)
			isFirstSimplePlain := !utils.IsAccessExpression(unwrappedFirst)

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
	if len(chain) < 2 {
		return nil
	}

	for _, op := range chain {
		if op.typ == OperandTypePlain || op.typ == OperandTypeNot {
			if processor.wouldChangeReturnType(op.comparedExpr) && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
				return nil
			}
		}
	}

	for i := range len(chain) - 1 {
		if chain[i].typ == OperandTypeNot {
			negatedExpr := chain[i].comparedExpr
			for j := i + 1; j < len(chain); j++ {
				if chain[j].typ == OperandTypeNot {
					continue
				}
				callExpr := chain[j].comparedExpr
				if callExpr != nil && ast.IsCallExpression(ast.SkipParentheses(callExpr)) {
					call := ast.SkipParentheses(callExpr).AsCallExpression()
					callBase := call.Expression
					cmp := processor.compareNodes(negatedExpr, callBase)
					if cmp == NodeEqual {
						return nil
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

	if !processor.allSubsequentHaveOptionalChaining(chain) {
		return false
	}

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
		if op.typ.IsLooseNullishCheck() {
			allStrictChecks = false
			break
		}
		if op.typ.IsStrictNullishCheck() {
			hasNullishCheck = true
		}
	}

	if !hasNullishCheck {
		return false
	}

	if !allStrictChecks {
		return false
	}

	// Strict checks (=== null or === undefined) intentionally cover one - skip only when type has BOTH
	firstOp := chain[0]
	firstTypeInfo := processor.getTypeInfo(firstOp.comparedExpr)

	return firstTypeInfo.HasBothNullAndUndefined()
}

func (processor *chainProcessor) shouldSkipOrChainOptimalChecks(chain []Operand) bool {
	if len(chain) < 2 {
		return false
	}

	if !processor.allSubsequentHaveOptionalChaining(chain) {
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
						if chain[i].typ.IsLooseNullishCheck() {
							closestLooseCheckExpr = chain[i].comparedExpr
							break
						}
					}

					if closestLooseCheckExpr != nil {
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
	chainLoop:
		for i := len(chain) - 1; i >= 0; i-- {
			switch {
			case chain[i].typ == OperandTypePlain:
				lastPropertyAccess = chain[i].node
				hasTrailingComparison = false
				hasTrailingTypeofCheck = false
				break chainLoop
			case chain[i].typ == OperandTypeComparison ||
				chain[i].typ == OperandTypeNotStrictEqualNull ||
				chain[i].typ == OperandTypeNotStrictEqualUndef ||
				chain[i].typ == OperandTypeNotEqualBoth:
				lastPropertyAccess = chain[i].comparedExpr
				hasTrailingComparison = true
				hasTrailingTypeofCheck = false
				break chainLoop
			case chain[i].typ == OperandTypeTypeofCheck:
				lastPropertyAccess = chain[i].comparedExpr
				hasTrailingComparison = true
				hasTrailingTypeofCheck = true
				break chainLoop
			case chain[i].comparedExpr != nil:
				lastPropertyAccess = chain[i].comparedExpr
				hasTrailingComparison = false
				hasTrailingTypeofCheck = false
				break chainLoop
			}
		}
	}

	if lastPropertyAccess == nil {
		return
	}

	parts := processor.flattenForFix(lastPropertyAccess)

	if len(chain) > 0 && len(parts) > 0 && chain[0].typ == OperandTypePlain && chain[0].node != nil {
		firstParts := processor.flattenForFix(chain[0].node)
		if len(firstParts) == 1 && len(firstParts) <= len(parts) {
			if len(firstParts[0].text) > len(parts[0].text) {
				parts[0] = firstParts[0]
			}
		}
	}

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
						parts[i].text += "!"
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

			if hasTrailingTypeofCheck {
				leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)
				comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.comparedExpr)

				typeofPrefix := processor.sourceText[leftRange.Pos():comparedExprRange.Pos()]

				binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, operandForComparison.node).End()
				comparisonSuffix := processor.sourceText[comparedExprRange.End():binExprEnd]

				newCode = typeofPrefix + newCode + comparisonSuffix
			} else {
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
					newCode += comparisonSuffix
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

		if useSuggestion {
			for _, op := range chain {
				if op.comparedExpr != nil {
					info := processor.getTypeInfo(op.comparedExpr)
					if info.hasUndefined || info.IsAnyOrUnknown() {
						useSuggestion = false
						break
					}
				}
			}
		}

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
					if !info.IsAnyOrUnknown() {
						useSuggestion = true
					}
				}
			}
		}

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
}

func (processor *chainProcessor) generateOrChainFixAndReport(node *ast.Node, chain []Operand, operandNodes []*ast.Node) {
	hasTrailingComparison := false
	if len(chain) > 0 {
		lastOp := chain[len(chain)-1]
		hasTrailingComparison = lastOp.typ.IsComparisonOrNullCheck()
	}

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

	if len(chainForOptional) == 1 && trailingPlainOperand != "" {
		singleOp := chainForOptional[0]
		if singleOp.comparedExpr != nil {
			if processor.containsOptionalChain(singleOp.comparedExpr) {
				return
			}
		}
	}

	if len(chain) == 2 && trailingPlainOperand == "" {
		firstOp := chain[0]
		if firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
			return
		}
		if firstOp.node != nil && processor.containsOptionalChain(firstOp.node) {
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
		hasTrailingComparisonForFix = lastOpForFix.typ.IsComparisonOrNullCheck()
	}

	if hasTrailingComparisonForFix {
		lastOpForFix := chainForOptional[len(chainForOptional)-1]
		if ast.IsBinaryExpression(lastOpForFix.node) {
			binExpr := lastOpForFix.node.AsBinaryExpression()
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
			analysis := analyzeNullishChecks(chain, true)

			if analysis.HasIncompleteCheck() {
				allTypesMatchCheck := true
				hasAnyNullableType := false
				for _, op := range chain {
					if op.comparedExpr == nil {
						continue
					}
					info := processor.getTypeInfo(op.comparedExpr)
					if info.HasNoNullableTypes() {
						continue
					}
					hasAnyNullableType = true
					if info.IsAnyOrUnknown() {
						allTypesMatchCheck = false
						break
					}
					if info.HasBothNullAndUndefined() {
						allTypesMatchCheck = false
						break
					}
					if analysis.HasOnlyNullCheck() && (!info.hasNull || info.hasUndefined) {
						allTypesMatchCheck = false
						break
					}
					if analysis.HasOnlyUndefinedCheck() && (info.hasNull || !info.hasUndefined) {
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

	if useSuggestion && len(chain) > 0 {
		for _, op := range chain {
			if op.comparedExpr != nil {
				info := processor.getTypeInfo(op.comparedExpr)
				if info.IsAnyOrUnknown() || info.HasBothNullAndUndefined() {
					useSuggestion = false
					break
				}
			}
		}
	}

	processor.reportChainWithFixes(node, fixes, useSuggestion)
}

func (processor *chainProcessor) handleEmptyObjectPattern(node *ast.Node) {
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

	if processor.opts.RequireNullish {
		leftExpr := binExpr.Left
		if !processor.includesExplicitNullish(leftExpr) {
			return
		}
	}

	var accessExpr *ast.Node
	if utils.IsPropertyOrElementAccess(node.Parent) {
		accessExpr = node.Parent
	} else if ast.IsParenthesizedExpression(node.Parent) {
		grandParent := node.Parent.Parent
		if grandParent != nil && utils.IsPropertyOrElementAccess(grandParent) {
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

	leftNode := binExpr.Left
	leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, leftNode)
	leftText := processor.sourceText[leftRange.Pos():leftRange.End()]

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
