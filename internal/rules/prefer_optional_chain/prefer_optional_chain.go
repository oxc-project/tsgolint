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

// stripParens strips surrounding parentheses from text
func stripParens(text string) string {
	text = strings.TrimSpace(text)
	// Keep stripping outer parentheses as long as they're balanced
	for len(text) > 2 && text[0] == '(' && text[len(text)-1] == ')' {
		// Check if the opening and closing parens are paired
		depth := 0
		paired := true
		for i := 0; i < len(text); i++ {
			if text[i] == '(' {
				depth++
			} else if text[i] == ')' {
				depth--
				if depth == 0 && i < len(text)-1 {
					// Found closing paren before end - not fully paired
					paired = false
					break
				}
			}
		}
		if paired {
			text = strings.TrimSpace(text[1 : len(text)-1])
		} else {
			break
		}
	}
	return text
}

// removeAllParens removes ALL parentheses from text for normalization
func removeAllParens(text string) string {
	// Remove all parentheses that are not part of function calls
	// This is a simple approach: remove ( and ) but keep the content
	var result strings.Builder
	inCall := false
	depth := 0

	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch == '(' {
			// Check if this looks like a function call (preceded by identifier, ], ), or > for generic calls)
			// Added > to handle foo<T>() pattern where ( follows the >
			if i > 0 && (text[i-1] == ']' || text[i-1] == ')' || text[i-1] == '>' || (text[i-1] >= 'a' && text[i-1] <= 'z') || (text[i-1] >= 'A' && text[i-1] <= 'Z') || text[i-1] == '_' || text[i-1] == '$' || (text[i-1] >= '0' && text[i-1] <= '9')) {
				// Likely a function call, keep the parentheses
				inCall = true
				result.WriteByte(ch)
				depth++
			} else {
				// Grouping parentheses, skip it
				depth++
			}
		} else if ch == ')' {
			depth--
			if inCall && depth == 0 {
				inCall = false
				result.WriteByte(ch)
			} else if !inCall {
				// Grouping parentheses, skip it
			} else {
				result.WriteByte(ch)
			}
		} else {
			result.WriteByte(ch)
		}
	}
	return result.String()
}

// removeTypeAnnotations removes TypeScript type annotations from text for comparison
func removeTypeAnnotations(text string) string {
	// Remove angle bracket type assertions: <Type>expr -> expr
	// Pattern: <{...}>expr or <SomeType>expr
	// We need to be careful to match balanced brackets
	for {
		ltIndex := strings.Index(text, "<")
		if ltIndex == -1 {
			break
		}
		// Find the matching >
		depth := 1
		gtIndex := -1
		for i := ltIndex + 1; i < len(text); i++ {
			if text[i] == '<' {
				depth++
			} else if text[i] == '>' {
				depth--
				if depth == 0 {
					gtIndex = i
					break
				}
			}
		}
		if gtIndex == -1 {
			// No matching >, skip this <
			break
		}
		// Remove the <Type> part, keeping the expression after it
		text = text[:ltIndex] + text[gtIndex+1:]
	}

	// Remove "as Type" patterns
	// This is a simple regex-like approach
	text = strings.ReplaceAll(text, " as any", "")
	text = strings.ReplaceAll(text, " as unknown", "")
	// Remove generic "as SomeType" by finding " as " and skipping until we hit a property access
	// For simplicity, we'll use a more aggressive approach
	for {
		asIndex := strings.Index(text, " as ")
		if asIndex == -1 {
			break
		}
		// Find the end of the type assertion (next . or [ or ! or ? or end of string)
		endIndex := len(text)
		for i := asIndex + 4; i < len(text); i++ {
			if text[i] == '.' || text[i] == '[' || text[i] == '!' || text[i] == '?' {
				endIndex = i
				break
			}
		}
		text = text[:asIndex] + text[endIndex:]
	}

	// Remove "!" non-null assertions at the end of identifiers (before . or [)
	// foo! -> foo, but keep foo!.bar as foo.bar
	text = strings.ReplaceAll(text, "!.", ".")
	text = strings.ReplaceAll(text, "![", "[")

	return text
}

// stripTrailingNonNull strips trailing non-null assertions from text
// foo.bar! -> foo.bar
// foo.bar!.baz! -> foo.bar.baz (strip after each segment)
func stripTrailingNonNull(text string) string {
	// Remove trailing ! at the very end
	for len(text) > 0 && text[len(text)-1] == '!' {
		text = text[:len(text)-1]
	}
	// Also remove ! before property accesses (foo!.bar -> foo.bar)
	text = strings.ReplaceAll(text, "!.", ".")
	text = strings.ReplaceAll(text, "![", "[")
	return text
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

// chainProcessor manages state and provides helper methods for processing optional chain candidates
type chainProcessor struct {
	ctx                rule.RuleContext
	opts               PreferOptionalChainOptions
	seenLogicals       map[*ast.Node]bool
	processedAndRanges []textRange
	seenLogicalRanges  map[textRange]bool
	reportedRanges     map[textRange]bool
}

// newChainProcessor creates a new chainProcessor with initialized state
func newChainProcessor(ctx rule.RuleContext, opts PreferOptionalChainOptions) *chainProcessor {
	return &chainProcessor{
		ctx:                ctx,
		opts:               opts,
		seenLogicals:       make(map[*ast.Node]bool),
		processedAndRanges: []textRange{},
		seenLogicalRanges:  make(map[textRange]bool),
		reportedRanges:     make(map[textRange]bool),
	}
}

// extractCallSignatures extracts call signatures from a node for comparison.
// Returns a map of "base expression" -> "full call text" for all call expressions in the node.
func (processor *chainProcessor) extractCallSignatures(node *ast.Node) map[string]string {
	signatures := make(map[string]string)
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ast.IsCallExpression(n) {
			call := n.AsCallExpression()
			// Get base expression text
			exprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, call.Expression)
			exprText := processor.ctx.SourceFile.Text()[exprRange.Pos():exprRange.End()]
			// Get full call text (including args and type args)
			fullRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			fullText := processor.ctx.SourceFile.Text()[fullRange.Pos():fullRange.End()]
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
	return signatures
}

// compareNodes compares two nodes to determine their relationship.
func (processor *chainProcessor) compareNodes(left, right *ast.Node) NodeComparisonResult {
	leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, left)
	rightRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, right)

	// Bounds check to prevent panic - if ranges are invalid, nodes are not equal
	sourceText := processor.ctx.SourceFile.Text()
	if leftRange.Pos() < 0 || leftRange.End() > len(sourceText) || leftRange.Pos() > leftRange.End() {
		return NodeInvalid
	}
	if rightRange.Pos() < 0 || rightRange.End() > len(sourceText) || rightRange.Pos() > rightRange.End() {
		return NodeInvalid
	}

	// Check for side effects in either expression
	// Example: foo[x++] && foo[x++].bar -> Cannot convert (x++ has side effects)
	// Example: foo[yield x] && foo[yield x].bar -> Cannot convert (yield has side effects)
	if hasSideEffects(left) || hasSideEffects(right) {
		return NodeInvalid
	}

	leftText := sourceText[leftRange.Pos():leftRange.End()]
	rightText := sourceText[rightRange.Pos():rightRange.End()]

	// Strip surrounding parentheses for comparison
	leftText = stripParens(leftText)
	rightText = stripParens(rightText)

	// Check if the left operand is a CallExpression or NewExpression at the base level
	// If so, we cannot safely chain because calling the function/constructor multiple times may have side effects
	// Example: getFoo() && getFoo().bar -> Cannot convert (getFoo() might have side effects)
	// Example: new Date() && new Date().getTime() -> Cannot convert (different instances)
	// EXCEPTION: If the expression already contains optional chaining (?.), it's safe to extend
	// Example: foo?.() || foo?.().bar -> CAN convert to foo?.()?.bar (single evaluation)
	// Also check for literal expressions (arrays, objects, functions, classes) which create new instances
	// Example: [] && [].length -> Cannot convert (different arrays)
	// Example: (class Foo {}) && class Foo {}.name -> Cannot convert (different classes)
	// Unwrap parentheses manually since unwrapParentheses is defined later
	leftUnwrapped := left
	for ast.IsParenthesizedExpression(leftUnwrapped) {
		leftUnwrapped = leftUnwrapped.AsParenthesizedExpression().Expression
	}
	// Allow call expressions if they contain optional chaining (already safe)
	hasOptionalChaining := strings.Contains(leftText, "?.")
	if !hasOptionalChaining {
		if ast.IsCallExpression(leftUnwrapped) || ast.IsNewExpression(leftUnwrapped) ||
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

	// Normalize: remove ALL parentheses (not part of calls), type annotations, optional chain operators, and non-null assertions
	// (foo as any)?.bar?.baz! should be compared as foo.bar.baz
	// foo?.bar?.baz should be compared as foo.bar.baz
	// foo?.() should be compared as foo()
	// foo?.[bar] should be compared as foo[bar]
	// foo.bar! should be compared as foo.bar
	leftNormalized := removeAllParens(leftText)
	rightNormalized := removeAllParens(rightText)
	leftNormalized = removeTypeAnnotations(leftNormalized)
	rightNormalized = removeTypeAnnotations(rightNormalized)
	// Remove optional chaining operators while preserving valid syntax:
	// - ?.( -> ( (optional call)
	// - ?.[ -> [ (optional element access)
	// - ?. -> . (optional property access)
	leftNormalized = strings.ReplaceAll(leftNormalized, "?.(", "(")
	leftNormalized = strings.ReplaceAll(leftNormalized, "?.[", "[")
	leftNormalized = strings.ReplaceAll(leftNormalized, "?.", ".")
	rightNormalized = strings.ReplaceAll(rightNormalized, "?.(", "(")
	rightNormalized = strings.ReplaceAll(rightNormalized, "?.[", "[")
	rightNormalized = strings.ReplaceAll(rightNormalized, "?.", ".")
	// Remove trailing non-null assertions for comparison
	// foo.bar! should equal foo.bar
	// But be careful not to remove ! from other contexts
	// We strip trailing ! that's not inside brackets or parens
	leftNormalized = stripTrailingNonNull(leftNormalized)
	rightNormalized = stripTrailingNonNull(rightNormalized)

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
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)
	for _, part := range typeParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			return true
		}
		// any and unknown can be nullish at runtime
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return true
		}
	}
	return false
}

// includesExplicitNullish checks if a type includes explicit nullish types (null | undefined).
// This does NOT return true for 'any' or 'unknown' types.
// Used to determine if autofix is safe when allowPotentiallyUnsafe is false.
func (processor *chainProcessor) includesExplicitNullish(node *ast.Node) bool {
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)
	for _, part := range typeParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			return true
		}
	}
	return false
}

// typeIsAnyOrUnknown checks if a type is any or unknown (where we can't determine exact nullishness).
func (processor *chainProcessor) typeIsAnyOrUnknown(node *ast.Node) bool {
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)
	// If all parts are any/unknown, return true
	for _, part := range typeParts {
		if !utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return false
		}
	}
	return len(typeParts) > 0
}

// typeIncludesNull checks if a type includes the null type specifically.
func (processor *chainProcessor) typeIncludesNull(node *ast.Node) bool {
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)
	for _, part := range typeParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull) {
			return true
		}
		// any and unknown can be null at runtime
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return true
		}
	}
	return false
}

// typeIncludesUndefined checks if a type includes the undefined type specifically.
func (processor *chainProcessor) typeIncludesUndefined(node *ast.Node) bool {
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)
	for _, part := range typeParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
			return true
		}
		// any and unknown can be undefined at runtime
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return true
		}
	}
	return false
}

// wouldChangeReturnType checks if converting to optional chaining would change the return type.
// This happens when the type includes falsy non-nullish values (false, 0, ”, 0n)
// but does NOT include null/undefined.
func (processor *chainProcessor) wouldChangeReturnType(node *ast.Node) bool {
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)

	hasNullish := false
	hasFalsyNonNullish := false

	for _, part := range typeParts {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			hasNullish = true
		}
		// Check for falsy non-nullish values
		// Note: We check for literal types like 'false', '0', '', '0n'
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLiteral) ||
			utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral) ||
			utils.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) ||
			utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLiteral) {
			// Check if it's a falsy literal by looking at the type text
			// TODO: This is a heuristic; ideally we'd check the actual value
			hasFalsyNonNullish = true
		}
	}

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
	nodeType := processor.ctx.TypeChecker.GetTypeAtLocation(node)
	typeParts := utils.UnionTypeParts(nodeType)

	for _, part := range typeParts {
		// Skip nullish types
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}

		// Check for void type (always falsy, not nullish)
		if utils.IsTypeFlagSet(part, checker.TypeFlagsVoid) {
			return true
		}
	}

	return false
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
	t := processor.ctx.TypeChecker.GetTypeAtLocation(baseNode)
	types := utils.UnionTypeParts(t)

	for _, part := range types {
		// Skip nullish types - they're always allowed
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}

		// Check each type flag
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny) && !processor.opts.CheckAny {
			return true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLike) && !processor.opts.CheckBigInt {
			return true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLike) && !processor.opts.CheckBoolean {
			return true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLike) && !processor.opts.CheckNumber {
			return true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsStringLike) && !processor.opts.CheckString {
			return true
		}
		if utils.IsTypeFlagSet(part, checker.TypeFlagsUnknown) && !processor.opts.CheckUnknown {
			return true
		}
	}

	return false
}

// flattenForFix flattens a chain expression to its component parts for reconstruction
func (processor *chainProcessor) flattenForFix(node *ast.Node) []ChainPart {
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
				text := processor.ctx.SourceFile.Text()[textRange.Pos():textRange.End()]

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
			nameText := processor.ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

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
			argText := processor.ctx.SourceFile.Text()[argRange.Pos():argRange.End()]

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
				typeArgsText = "<" + processor.ctx.SourceFile.Text()[typeArgsStart:typeArgsEnd] + ">"
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
				argsText = "(" + processor.ctx.SourceFile.Text()[argsStart:callEnd-1] + ")"
			}

			parts = append(parts, ChainPart{
				text:        typeArgsText + argsText,
				optional:    callExpr.QuestionDotToken != nil,
				requiresDot: false,
			})

		default:
			// Base case - identifier or other expression
			textRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, n)
			text := processor.ctx.SourceFile.Text()[textRange.Pos():textRange.End()]

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

	allLogicalNodes := flattenAndMarkLogicals(node)

	// Collect all && operands (keeping track of original nodes with parentheses)
	operandNodes := []*ast.Node{}
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		// Check the unwrapped node for the operator type
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
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
	var lastCheckType OperandType      // Track the type of the last nullish check
	var chainComplete bool             // Mark when chain should not accept more operands
	var stopProcessing bool            // Stop processing after inconsistent check
	var chainHasSafeCallExtension bool // Track if chain has been safely extended through a call
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
			chainHasSafeCallExtension = false
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
			chainHasSafeCallExtension = false
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
			chainHasSafeCallExtension = false
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
			isStrictExplicitCheck := prevOp.typ == OperandTypeNotStrictEqualNull ||
				prevOp.typ == OperandTypeNotStrictEqualUndef
			if isStrictExplicitCheck && prevOp.comparedExpr != nil {
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

		// Special case for AND chains with unsafe option enabled:
		// Allow extending call expressions even though they may have side effects
		// Example: foo.bar() && foo.bar().baz with unsafe option
		// This is different from: getFoo() && getFoo().bar (different calls, always unsafe)
		// Track if we used special handling to allow call chain extension
		usedCallChainExtension := false
		if cmp == NodeInvalid && processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
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
					sourceText := processor.ctx.SourceFile.Text()
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
								chainHasSafeCallExtension = true // Mark chain-level flag
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
			if op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeNotStrictEqualUndef ||
				op.typ == OperandTypeNotEqualBoth || op.typ == OperandTypeTypeofCheck {

				// Check for inconsistent check types
				// If we had a "both" check (!= null) and now have a specific check (!== undefined or !== null),
				// This is redundant but not incorrect - include it and continue
				// We DON'T mark the chain as complete because subsequent property accesses should be included
				// Example: foo != null && foo !== undefined && foo.bar -> foo?.bar (all checks on foo)
				if lastCheckType == OperandTypeNotEqualBoth &&
					(op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeNotStrictEqualUndef) {
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
				if prevOp.typ == OperandTypeNotStrictEqualNull ||
					prevOp.typ == OperandTypeNotStrictEqualUndef ||
					prevOp.typ == OperandTypeNotEqualBoth ||
					prevOp.typ == OperandTypeTypeofCheck {
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
				if prevOp.typ == OperandTypeNotStrictEqualNull ||
					prevOp.typ == OperandTypeNotStrictEqualUndef ||
					prevOp.typ == OperandTypeNotEqualBoth ||
					prevOp.typ == OperandTypeTypeofCheck {
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

			// Check if current operand is an explicit null/undefined check
			// and we already have at least one check in the chain
			// If so, CONTINUE adding to the chain for simple property/element access chains
			// But FINALIZE for call expression chains (to avoid multiple evaluations)
			// Example to ALLOW: foo != null && foo.bar != null && foo.bar.baz (property chain)
			// Example to FINALIZE: foo.bar !== undefined && foo.bar() !== undefined (call chain)
			isExplicitCheck := op.typ == OperandTypeNotStrictEqualNull ||
				op.typ == OperandTypeNotStrictEqualUndef ||
				op.typ == OperandTypeNotEqualBoth ||
				op.typ == OperandTypeTypeofCheck

			if isExplicitCheck && len(currentChain) >= 2 {
				// Check if we already have at least one explicit check in the chain
				hasExplicitCheck := false
				for _, chainOp := range currentChain {
					if chainOp.typ == OperandTypeNotStrictEqualNull ||
						chainOp.typ == OperandTypeNotStrictEqualUndef ||
						chainOp.typ == OperandTypeNotEqualBoth ||
						chainOp.typ == OperandTypeTypeofCheck {
						hasExplicitCheck = true
						break
					}
				}

				// Check if this is a call expression chain (has side effects)
				// If so, use conservative approach (finalize chain)
				// Otherwise, continue adding to the chain
				hasCallInChain := false
				isCallingCheckedExpr := false // True if current operand is calling the expression we just checked
				if op.comparedExpr != nil {
					unwrappedOp := unwrapParentheses(op.comparedExpr)
					if ast.IsCallExpression(unwrappedOp) {
						hasCallInChain = true
						// Check if this call is calling the lastExpr (the expression we just checked)
						// Example: foo.bar != null && foo.bar() - foo.bar() is calling foo.bar
						callExpr := unwrappedOp.AsCallExpression()
						if callExpr != nil && callExpr.Expression != nil {
							calleeText := ""
							calleeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, callExpr.Expression)
							if calleeRange.Pos() >= 0 && calleeRange.End() <= len(processor.ctx.SourceFile.Text()) {
								calleeText = processor.ctx.SourceFile.Text()[calleeRange.Pos():calleeRange.End()]
							}
							lastText := ""
							lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastExpr)
							if lastRange.Pos() >= 0 && lastRange.End() <= len(processor.ctx.SourceFile.Text()) {
								lastText = processor.ctx.SourceFile.Text()[lastRange.Pos():lastRange.End()]
							}
							if calleeText == lastText {
								// The call is calling the expression we just checked - safe to continue
								isCallingCheckedExpr = true
							}
						}
					}
				}
				// Also check if previous operands had calls
				if !hasCallInChain {
					for _, chainOp := range currentChain {
						if chainOp.comparedExpr != nil {
							unwrappedChainOp := unwrapParentheses(chainOp.comparedExpr)
							if ast.IsCallExpression(unwrappedChainOp) {
								hasCallInChain = true
								break
							}
						}
					}
				}

				if hasExplicitCheck && hasCallInChain && !usedCallChainExtension && !chainHasSafeCallExtension && !isCallingCheckedExpr {
					// Finalize current chain and don't include this check
					// This is the conservative approach for call chains
					// BUT: if we used call chain extension (usedCallChainExtension=true) or
					// the chain has already been safely extended through a call (chainHasSafeCallExtension=true),
					// or we're calling the expression we just checked (isCallingCheckedExpr=true),
					// we're safely extending through a call that was already null-checked,
					// so we should continue building the chain instead of stopping.
					if len(currentChain) >= 2 {
						allChains = append(allChains, currentChain)
					}
					// Reset currentChain to prevent double-adding at loop end
					currentChain = nil
					chainComplete = true
					stopProcessing = true

					// Update processed range to include the ENTIRE node (top-level expression)
					// This prevents ANY sub-chains from being detected separately
					// We need to cover the full range of the top-level && expression
					topLevelRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, node)
					// Find the existing range entry and update it to cover the entire expression
					for idx := range processor.processedAndRanges {
						if processor.processedAndRanges[idx].start == nodeStart {
							processor.processedAndRanges[idx].end = topLevelRange.End()
							break
						}
					}

					// Mark ALL logical nodes in this expression tree to prevent any sub-chains
					// from being detected separately
					for _, logicalNode := range allLogicalNodes {
						if logicalNode != nil {
							processor.seenLogicals[logicalNode] = true
							unwrappedLogical := unwrapParentheses(logicalNode)
							processor.seenLogicals[unwrappedLogical] = true
						}
					}

					// Mark all remaining operand nodes as seen to prevent them from being
					// processed as a separate chain by the visitor
					var markAllNodes func(*ast.Node)
					markAllNodes = func(n *ast.Node) {
						if n == nil {
							return
						}
						processor.seenLogicals[n] = true
						unwrapped := unwrapParentheses(n)
						processor.seenLogicals[unwrapped] = true

						if ast.IsBinaryExpression(unwrapped) {
							binExpr := unwrapped.AsBinaryExpression()
							markAllNodes(binExpr.Left)
							markAllNodes(binExpr.Right)
						} else if ast.IsPropertyAccessExpression(unwrapped) {
							markAllNodes(unwrapped.AsPropertyAccessExpression().Expression)
						} else if ast.IsElementAccessExpression(unwrapped) {
							elemAccess := unwrapped.AsElementAccessExpression()
							markAllNodes(elemAccess.Expression)
							markAllNodes(elemAccess.ArgumentExpression)
						} else if ast.IsCallExpression(unwrapped) {
							call := unwrapped.AsCallExpression()
							markAllNodes(call.Expression)
						}
					}

					for j := i + 1; j < len(operandNodes); j++ {
						markAllNodes(operandNodes[j])
					}

					i++
					continue
				}
				// For property/element chains, continue adding to the chain
			}

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
							lastRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastChainOp.comparedExpr)
							sourceText := processor.ctx.SourceFile.Text()
							lastText := ""
							if lastRange.Pos() >= 0 && lastRange.End() <= len(sourceText) {
								lastText = sourceText[lastRange.Pos():lastRange.End()]
							}
							hasOptionalChaining := strings.Contains(lastText, "?.")

							// Also check if the call's base expression was checked earlier in our chain
							// If so, the base will be converted to optional chain
							callBaseWasChecked := false
							if !hasOptionalChaining {
								callExpr := unwrappedLast.AsCallExpression()
								if callExpr != nil && callExpr.Expression != nil {
									// Get the callee (e.g., foo.bar in foo.bar())
									calleeRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, callExpr.Expression)
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
						if chainOp.typ == OperandTypeNotStrictEqualNull || chainOp.typ == OperandTypeNotStrictEqualUndef {
							strictCheckCount++
						}
					}
					// Only apply Case 4 if:
					// - NOT using unsafe option (if unsafe, allow full conversion)
					// - We have at least one strict check already AND
					// - Current operand is also a strict check (resulting in 2+ strict checks)
					if !shouldStopChain && !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing &&
						strictCheckCount >= 1 && lastChainOp.comparedExpr != nil &&
						(op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeNotStrictEqualUndef) {
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
				(op.typ == OperandTypeNotStrictEqualNull || op.typ == OperandTypeNotStrictEqualUndef) {
				// Check if any PREVIOUS operand (not the one we just added) has an incomplete strict check
				for j := 0; j < len(currentChain)-1; j++ {
					prevOp := currentChain[j]
					if (prevOp.typ == OperandTypeNotStrictEqualNull || prevOp.typ == OperandTypeNotStrictEqualUndef) &&
						prevOp.comparedExpr != nil {
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

	// Check if we have multiple chains with different base identifiers
	// Without the unsafe option, we should skip ALL chains if they have different bases
	// This is to avoid partial conversions that might be confusing
	// Example: a.b && a.b.c && c.d && c.d.e (different bases: a and c)
	// With unsafe option: report both chains separately
	// Without unsafe option: skip all chains
	if !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing && len(chainsToReport) > 1 {
		// Check if all chains are plain AND chains (no explicit null checks)
		// If any chain has explicit null checks, allow conversion
		allPlain := true
		for _, chain := range chainsToReport {
			for _, op := range chain {
				if op.typ != OperandTypePlain {
					allPlain = false
					break
				}
			}
			if !allPlain {
				break
			}
		}

		if allPlain {
			// All chains are plain AND chains
			// Check if they have different base identifiers
			baseIdentifiers := make(map[string]bool)
			for _, chain := range chainsToReport {
				if len(chain) > 0 && chain[0].comparedExpr != nil {
					base := getBaseIdentifier(chain[0].comparedExpr)
					baseRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, base)
					sourceText := processor.ctx.SourceFile.Text()
					if baseRange.Pos() >= 0 && baseRange.End() <= len(sourceText) {
						baseText := sourceText[baseRange.Pos():baseRange.End()]
						baseIdentifiers[baseText] = true
					}
				}
			}

			// If we have multiple different bases, skip all chains
			if len(baseIdentifiers) > 1 {
				return // Skip all chains when multiple bases without unsafe option
			}
		}
	}

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
			isStrictCheck := firstOp.typ == OperandTypeNotStrictEqualNull ||
				firstOp.typ == OperandTypeNotStrictEqualUndef
			if isStrictCheck && firstOp.comparedExpr != nil && processor.containsOptionalChain(firstOp.comparedExpr) {
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
		hasPropertyAccess := false
		for _, op := range chain {
			if op.comparedExpr != nil {
				unwrapped := unwrapParentheses(op.comparedExpr)
				if ast.IsPropertyAccessExpression(unwrapped) ||
					ast.IsElementAccessExpression(unwrapped) ||
					ast.IsCallExpression(unwrapped) {
					hasPropertyAccess = true
					break
				}
			}
		}
		if !hasPropertyAccess {
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
				for j := 0; j < len(firstParts); j++ {
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

		// Skip chains where the subsequent operands already have optimal optional chaining
		// Example: x && x.y?.z -> already optimal, don't report
		// This happens when:
		// - First operand is a simple check (x)
		// - Second operand extends the first and already uses optional chaining (x.y?.z)
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
					for i := 0; i < len(firstParts); i++ {
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
						for i := 0; i < len(firstParts); i++ {
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
		if processor.opts.RequireNullish {
			hasNullishContext := false
			for i, op := range chain {
				// Check for explicit nullish check operators
				if op.typ != OperandTypePlain {
					hasNullishContext = true
					break
				}
				// For plain && checks, allow if the type explicitly includes null/undefined
				// (but only for intermediate operands, not the last one)
				if i < len(chain)-1 && op.comparedExpr != nil {
					if processor.includesExplicitNullish(op.comparedExpr) {
						hasNullishContext = true
						break
					}
				}
			}
			if !hasNullishContext {
				continue // Skip chains without nullish context when requireNullish is true
			}
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
				if lastOp.typ == OperandTypeNotStrictEqualNull || lastOp.typ == OperandTypeNotStrictEqualUndef ||
					lastOp.typ == OperandTypeNotEqualBoth || lastOp.typ == OperandTypeComparison {
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
			if !hasPlainTruthinessCheck && !hasBothCheck && !hasTypeofCheck {
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
		shouldSkip := false
		for i, op := range chain {
			if op.typ == OperandTypePlain {
				// Check if we should skip based on type
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
				if i == 0 {
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
		var hasComplementaryNullCheck bool // true if last two operands form a complementary null+undefined check

		// Check if the last two operands form a complementary pair (null + undefined checks on same expression)
		// If so, they should be simplified to `!= null`
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
						// Complementary pair! Simplify to `!= null`
						hasComplementaryNullCheck = true
						lastPropertyAccess = lastOp.comparedExpr
						hasTrailingComparison = true
						hasTrailingTypeofCheck = false
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
		for i := 0; i < len(chain); i++ {
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
				if strings.HasSuffix(parts[i].text, "!") {
					parts[i].text = parts[i].text[:len(parts[i].text)-1]
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
						for i := 0; i < len(opParts); i++ {
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
			// For optional chains (?.), use ANY operand that has it
			// For non-null assertions (!), use the SHORTEST operand that covers this index
			// This ensures we preserve ! from the earliest check, not from later extended checks
			for i := 0; i < len(parts); i++ {
				// Find the shortest operand that covers this index (for non-null assertions)
				var shortestCoveringOp *opPartsInfo
				for j := range allOpParts {
					op := &allOpParts[j]
					if i < op.len {
						if shortestCoveringOp == nil || op.len < shortestCoveringOp.len {
							shortestCoveringOp = op
						}
					}
				}

				// Merge from all operands for optional flag
				for _, op := range allOpParts {
					if i < op.len && op.parts[i].optional {
						parts[i].optional = true
					}
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
						trivia := processor.ctx.SourceFile.Text()[fullPos:trimmedPos]
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
			if hasComplementaryNullCheck {
				// Complementary null+undefined check - simplify to `!= null`
				newCode = newCode + " != null"
			} else {
				lastOperand := chain[len(chain)-1]
				if ast.IsBinaryExpression(lastOperand.node) {
					binExpr := lastOperand.node.AsBinaryExpression()

					// Special handling for typeof checks: typeof foo.bar !== 'undefined'
					// The binary expression is: (typeof foo.bar) !== 'undefined'
					// We need to wrap the optional chain with: typeof ... !== 'undefined'
					if hasTrailingTypeofCheck {
						// For typeof checks, we need to:
						// 1. Get the "typeof " prefix from the left side
						// 2. Get the " !== 'undefined'" suffix from after the comparedExpr
						leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)
						comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOperand.comparedExpr)

						// typeof prefix: from start of left side to start of comparedExpr
						typeofPrefix := processor.ctx.SourceFile.Text()[leftRange.Pos():comparedExprRange.Pos()]

						// comparison suffix: from end of comparedExpr to end of binary expression
						binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOperand.node).End()
						comparisonSuffix := processor.ctx.SourceFile.Text()[comparedExprRange.End():binExprEnd]

						newCode = typeofPrefix + newCode + comparisonSuffix
					} else {
						// Check if this is a yoda condition (literal/constant on left, property on right)
						// In yoda: '123' == foo.bar.baz
						// Not yoda: foo.bar.baz == '123'
						isYoda := false
						comparedExprRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOperand.comparedExpr)
						leftRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, binExpr.Left)

						// If comparedExpr is on the right side, it's yoda
						if comparedExprRange.Pos() > leftRange.Pos() {
							isYoda = true
						}

						if isYoda {
							// Yoda: prepend the left side + operator
							binExprStart := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOperand.node).Pos()
							comparedExprStart := comparedExprRange.Pos()
							yodaPrefix := processor.ctx.SourceFile.Text()[binExprStart:comparedExprStart]
							newCode = yodaPrefix + newCode
						} else {
							// Normal: append the operator + right side
							comparedExprEnd := comparedExprRange.End()
							binExprEnd := utils.TrimNodeTextRange(processor.ctx.SourceFile, lastOperand.node).End()
							comparisonSuffix := processor.ctx.SourceFile.Text()[comparedExprEnd:binExprEnd]
							newCode = newCode + comparisonSuffix
						}
					}
				}
			}
		}

		// Use trimmed ranges to preserve leading/trailing whitespace
		// If we're replacing the entire logical expression (all operands from this node),
		// use the node's range to include any wrapping parentheses
		// Otherwise, use the operand ranges
		var replaceStart, replaceEnd int

		// typeof checks should be removed when converting to optional chain
		// typeof foo !== 'undefined' && foo.bar -> foo?.bar
		// The optional chain handles the undefined check implicitly
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
		// When the unsafe option is enabled, always autofix
		// When it's not enabled, use suggestion UNLESS:
		//   1. The FINAL expression's type includes nullish (which makes it safe), OR
		//   2. We have a trailing comparison (which ensures return type is consistent)
		useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

		// Check if the FINAL operand (the one being accessed) includes nullish
		// If the final expression is already nullable, converting to optional chain is safe
		// Example: foo && foo.bar where foo.bar: string | null
		// Original: string | null | false (from foo being falsy)
		// With ?.: string | null | undefined - close enough, safe
		// Counter-example: foo && foo.bar where foo.bar: string
		// Original: string | false/null (from foo being falsy)
		// With ?.: string | undefined - changes return type, unsafe
		// NOTE: We use includesExplicitNullish here (not includesNullish) because
		// for 'any'/'unknown' types we should use suggestion since we can't know
		// if the conversion is safe or not
		if useSuggestion && len(chain) > 0 {
			lastOp := chain[len(chain)-1]
			if processor.includesExplicitNullish(lastOp.comparedExpr) {
				useSuggestion = false
			}
		}

		// For chains where ALL intermediate operands (except last) have EXPLICIT nullable types,
		// AND there are at least 2 intermediate operands (multiple nullable checks),
		// the conversion is safe because we're just adding optional chaining to already-nullable accesses.
		// Example: foo && foo.bar && foo.bar.toString() where foo: T | null | undefined and foo.bar: string | null | undefined
		// Even though toString() returns string (not nullable), the intermediate accesses are safe to chain.
		// Note: Single nullable check (foo && foo.bar) should remain suggestion only.
		// NOTE: We use includesExplicitNullish here (not includesNullish) because
		// for 'any'/'unknown' types we should use suggestion since we can't know
		// if the conversion is safe or not
		if useSuggestion && len(chain) > 2 {
			allIntermediateNullable := true
			for i := 0; i < len(chain)-1; i++ {
				op := chain[i]
				if op.comparedExpr != nil && !processor.includesExplicitNullish(op.comparedExpr) {
					allIntermediateNullable = false
					break
				}
			}
			if allIntermediateNullable {
				useSuggestion = false
			}
		}

		// For AND chains with trailing comparisons, always provide a fix
		// The comparison ensures the return type remains consistent
		// Example: foo && foo.bar === 0 -> foo?.bar === 0 (both return boolean)
		if useSuggestion && hasTrailingComparison {
			useSuggestion = false
		}

		// For chains guarded by typeof checks, always provide a fix
		// typeof x !== 'undefined' only checks for undefined, but x?.foo checks both null AND undefined
		// So converting is STRICTLY SAFER - it handles more cases
		// Example: typeof globalThis !== 'undefined' && globalThis.Array() -> globalThis?.Array()
		if useSuggestion && len(chain) > 0 {
			for _, op := range chain {
				if op.typ == OperandTypeTypeofCheck {
					useSuggestion = false
					break
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

	// Collect all || operands (keeping track of original nodes with parentheses)
	// Also collect all binary expression ranges to mark them as seen
	operandNodes := []*ast.Node{}
	var collectedBinaryRanges []textRange
	var collect func(*ast.Node)
	collect = func(n *ast.Node) {
		// Check the unwrapped node for the operator type
		unwrapped := unwrapParentheses(n)

		if ast.IsBinaryExpression(unwrapped) && unwrapped.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken {
			binExpr := unwrapped.AsBinaryExpression()
			collect(binExpr.Left)
			collect(binExpr.Right)
			// Mark nested binary expressions by range
			binRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, unwrapped)
			collectedBinaryRanges = append(collectedBinaryRanges, textRange{start: binRange.Pos(), end: binRange.End()})
		} else {
			// Store the original node (with parentheses) for range calculation
			operandNodes = append(operandNodes, n)
		}
	}
	collect(node)

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

	for i := 0; i < len(operands); i++ {
		op := operands[i]

		// Accept OperandTypeNot, OperandTypeComparison, OperandTypePlain, typeof checks, and null check types
		validOrOperand := op.typ == OperandTypeNot ||
			op.typ == OperandTypeComparison ||
			op.typ == OperandTypePlain ||
			op.typ == OperandTypeTypeofCheck ||
			op.typ == OperandTypeNotStrictEqualNull ||
			op.typ == OperandTypeNotStrictEqualUndef ||
			op.typ == OperandTypeNotEqualBoth

		// With unsafe option, also allow === null checks in OR chains
		// Example: foo === null || foo.bar === null -> foo?.bar === null
		// This changes semantics (when foo is null, original returns true, transformed returns false)
		// but is allowed with the unsafe option
		if !validOrOperand && processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
			validOrOperand = op.typ == OperandTypeStrictEqualNull ||
				op.typ == OperandTypeEqualNull ||
				op.typ == OperandTypeStrictEqualUndef
		}

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
			chain = append(chain, op)
			lastExpr = op.comparedExpr
			// Set hasTrailingComparison for both value comparisons AND null checks
			// Null checks like foo.bar == null should be preserved in the output
			isComparison := op.typ == OperandTypeComparison
			isNullCheck := op.typ == OperandTypeNotStrictEqualNull ||
				op.typ == OperandTypeNotStrictEqualUndef ||
				op.typ == OperandTypeNotEqualBoth ||
				op.typ == OperandTypeStrictEqualNull ||
				op.typ == OperandTypeEqualNull ||
				op.typ == OperandTypeStrictEqualUndef
			if isComparison || isNullCheck {
				hasTrailingComparison = true
			}
			continue
		}

		// Check if this continues the chain
		cmp := processor.compareNodes(lastExpr, op.comparedExpr)

		// Special case for OR chains with unsafe option enabled:
		// Allow extending call expressions even though they may have side effects
		// Example: foo.bar() || foo.bar().baz with unsafe option
		// This is different from: getFoo() && getFoo().bar (different calls, always unsafe)
		if cmp == NodeInvalid && processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
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
					sourceText := processor.ctx.SourceFile.Text()
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

		if cmp == NodeSubset || cmp == NodeEqual {
			// Special case: Don't add a value comparison (OperandTypeComparison) when it's
			// on the same expression as a previous negation/null check.
			// Pattern: !foo || !foo.bar || foo.bar > 5
			//   -> !foo?.bar || foo.bar > 5 (NOT foo?.bar > 5)
			// The comparison to a non-nullish value (> 5) is a different semantic check
			// and should NOT be part of the optional chain.
			if cmp == NodeEqual && op.typ == OperandTypeComparison && len(chain) > 0 {
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
			isComparison := op.typ == OperandTypeComparison
			isNullCheck := op.typ == OperandTypeNotStrictEqualNull ||
				op.typ == OperandTypeNotStrictEqualUndef ||
				op.typ == OperandTypeNotEqualBoth ||
				op.typ == OperandTypeStrictEqualNull ||
				op.typ == OperandTypeEqualNull ||
				op.typ == OperandTypeStrictEqualUndef
			if isComparison || isNullCheck {
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
			sourceText := processor.ctx.SourceFile.Text()
			if firstBaseRange.Pos() >= 0 && firstBaseRange.End() <= len(sourceText) &&
				baseRange.Pos() >= 0 && baseRange.End() <= len(sourceText) {
				firstBaseText := sourceText[firstBaseRange.Pos():firstBaseRange.End()]
				baseText := sourceText[baseRange.Pos():baseRange.End()]
				if firstBaseText != baseText {
					return // Different base identifiers in the same chain
				}
			}
		}
	}

	// Ensure at least one operand involves property/element/call access
	// Pattern to skip: foo === null || foo === undefined (just null checks, no access)
	// Pattern to allow: foo === null || foo.bar (has property access)
	hasPropertyAccess := false
	for _, op := range chain {
		if op.comparedExpr != nil {
			unwrapped := unwrapParentheses(op.comparedExpr)
			if ast.IsPropertyAccessExpression(unwrapped) ||
				ast.IsElementAccessExpression(unwrapped) ||
				ast.IsCallExpression(unwrapped) {
				hasPropertyAccess = true
				break
			}
		}
	}
	if !hasPropertyAccess {
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
				// Block patterns like:
				// - !foo || foo.bar (negation + plain property access)
				// - !foo || foo.bar === 'foo' (strict equality with non-undefined - NOT safe)
				// - !foo || foo.bar !== 'foo' (strict not-equal - NOT safe)
				// - !foo || foo.bar != null (loose not-equal with null/undefined - NOT safe)
				allNegatedOrSafeComparisonOrNullCheck := true
				for i := 1; i < len(chain); i++ {
					isNullCheck := chain[i].typ == OperandTypeNotStrictEqualNull ||
						chain[i].typ == OperandTypeNotStrictEqualUndef ||
						chain[i].typ == OperandTypeNotEqualBoth ||
						chain[i].typ == OperandTypeTypeofCheck
					isComparison := chain[i].typ == OperandTypeComparison
					isSafeComparison := isComparison && processor.isOrChainComparisonSafe(chain[i])

					if chain[i].typ != OperandTypeNot && !isSafeComparison && !isNullCheck {
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
			// Check if trailing comparisons are safe
			for i := 1; i < len(chain); i++ {
				if chain[i].typ == OperandTypeComparison && !processor.isOrChainComparisonSafe(chain[i]) {
					return // Unsafe comparison (e.g., with undeclared variable), skip this chain
				}
			}
		}
	}

	// When requireNullish is true, skip chains that start with negation (!foo || !foo.bar)
	// Only allow chains that start with explicit null checks (foo == null || foo.bar)
	if processor.opts.RequireNullish {
		firstOpIsNegation := chain[0].typ == OperandTypeNot
		if firstOpIsNegation {
			return
		}
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

		for _, op := range chain {
			if op.typ == OperandTypeNotStrictEqualNull {
				hasNullCheck = true
			} else if op.typ == OperandTypeNotStrictEqualUndef {
				hasUndefinedCheck = true
			} else if op.typ == OperandTypeNotEqualBoth {
				hasBothCheck = true
			} else if op.typ == OperandTypeTypeofCheck {
				// typeof checks are equivalent to undefined checks
				// (typeof x === 'undefined' in OR chains means undefined check)
				hasUndefinedCheck = true
			}
		}

		// If we have a strict null check or strict undefined check (but not both), skip
		// This is unsafe regardless of other checks in the chain
		// Note: typeof checks count as undefined checks
		if !hasBothCheck {
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
		for _, op := range chain {
			if op.typ == OperandTypeComparison && op.node != nil && op.comparedExpr != nil {
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

						if isStrictNullCheck || isStrictUndefCheck {
							// Get the type of the compared expression (property access)
							propType := processor.ctx.TypeChecker.GetTypeAtLocation(op.comparedExpr)
							typeParts := utils.UnionTypeParts(propType)

							// Check if type includes BOTH null and undefined
							typeHasNull := false
							typeHasUndefined := false
							for _, part := range typeParts {
								if utils.IsTypeFlagSet(part, checker.TypeFlagsNull) {
									typeHasNull = true
								}
								if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
									typeHasUndefined = true
								}
							}

							// If type has both null and undefined, but we only check one, reject
							if typeHasNull && typeHasUndefined {
								if isStrictNullCheck && !isStrictUndefCheck {
									return
								}
								if isStrictUndefCheck && !isStrictNullCheck {
									return
								}
							}
						}
					}
				}
			}
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
		for i := 0; i < len(chain)-1; i++ {
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
		if op.typ == OperandTypeComparison && ast.IsBinaryExpression(op.node) {
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

	// Special case: OR chain with trailing plain operand after MULTIPLE null checks
	// Pattern: !a || a.b == null || ... || a.b.c.d.e.f.g == null || a.b.c.d.e.f.g.h
	// Expected: a?.b?.c?.d?.e?.f?.g == null || a.b.c.d.e.f.g.h
	// The plain operand should NOT be converted to optional chain; it should remain separate
	//
	// BUT for simple 2-operand chains like: foo == null || foo.bar
	// Expected: foo?.bar (NOT keeping them separate)
	// So we only separate trailing plain when there are 3+ operands
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
				trailingPlainOperand = processor.ctx.SourceFile.Text()[lastOpRange.Pos():lastOpRange.End()]
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
			singleOpRange := utils.TrimNodeTextRange(processor.ctx.SourceFile, singleOp.comparedExpr)
			singleOpText := processor.ctx.SourceFile.Text()[singleOpRange.Pos():singleOpRange.End()]
			if strings.Contains(singleOpText, "?.") {
				return
			}
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
		isComparison := lastOpForFix.typ == OperandTypeComparison
		isNullCheck := lastOpForFix.typ == OperandTypeNotStrictEqualNull ||
			lastOpForFix.typ == OperandTypeNotStrictEqualUndef ||
			lastOpForFix.typ == OperandTypeNotEqualBoth ||
			lastOpForFix.typ == OperandTypeStrictEqualNull ||
			lastOpForFix.typ == OperandTypeEqualNull ||
			lastOpForFix.typ == OperandTypeStrictEqualUndef
		hasTrailingComparisonForFix = isComparison || isNullCheck
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
				opText := processor.ctx.SourceFile.Text()[opRange.Pos():opRange.End()]
				// Extract left side (the value being compared)
				valueText := strings.TrimSpace(processor.ctx.SourceFile.Text()[leftRange.Pos():leftRange.End()])
				newCode = optionalChainCode + " " + opText + " " + valueText
			} else {
				// Normal: append the operator + right side
				opStart := binExpr.OperatorToken.Pos()
				rightEnd := binExpr.Right.End()
				comparisonText := processor.ctx.SourceFile.Text()[opStart:rightEnd]
				newCode = optionalChainCode + comparisonText
			}
		} else {
			newCode = optionalChainCode
		}
	} else {
		// Check if first operand is negated (!foo || !foo.bar)
		// If ALL operands are negated, add negation: !foo || !foo.bar -> !foo?.bar
		// Otherwise no negation: foo || foo.bar -> foo?.bar
		//                        foo == null || foo.bar -> foo?.bar
		//                        typeof foo === 'undefined' || foo.bar -> foo?.bar
		firstOpIsNegated := chainForOptional[0].typ == OperandTypeNot

		if firstOpIsNegated {
			newCode = "!" + optionalChainCode
		} else {
			newCode = optionalChainCode
		}
	}

	// Append trailing plain operand if we kept one separate
	if trailingPlainOperand != "" {
		newCode = newCode + " ||\n            " + trailingPlainOperand
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

	useSuggestion := !processor.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing

	// Check if the FINAL operand (the one being accessed) includes nullish
	// If the final expression is already nullable, converting to optional chain is safe
	// NOTE: We use includesExplicitNullish here (not includesNullish) because
	// for 'any'/'unknown' types we should use suggestion since we can't know
	// if the conversion is safe or not
	if useSuggestion && len(chain) > 0 {
		lastOp := chain[len(chain)-1]
		if processor.includesExplicitNullish(lastOp.comparedExpr) {
			useSuggestion = false
		}
	}

	// For OR chains with trailing comparisons, always provide a fix
	// The comparison ensures the return type remains consistent
	// Example: !foo || foo.bar === undefined -> foo?.bar === undefined (both return boolean)
	if useSuggestion && hasTrailingComparison {
		useSuggestion = false
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

	// When requireNullish is true, skip empty object patterns entirely
	// These patterns are conceptually different from explicit nullish checks in && chains
	if processor.opts.RequireNullish {
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
	leftText := processor.ctx.SourceFile.Text()[leftRange.Pos():leftRange.End()]

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
		propertyText = "[" + processor.ctx.SourceFile.Text()[propRange.Pos():propRange.End()] + "]"
	} else {
		propertyText = processor.ctx.SourceFile.Text()[propRange.Pos():propRange.End()]
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

				// Handle && chains
				if operator == ast.KindAmpersandAmpersandToken {
					processor.processAndChain(node)
				}

				// Handle || chains
				if operator == ast.KindBarBarToken {
					processor.processOrChain(node)
					processor.handleEmptyObjectPattern(node)
				}

				// Handle ?? chains for empty object pattern
				if operator == ast.KindQuestionQuestionToken {
					processor.handleEmptyObjectPattern(node)
				}
			},
		}
	},
}
