package prefer_nullish_coalescing

import (
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// Message builders
func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPreferNullishOverOrMessage(description, equals string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverOr",
		Description: "Prefer using nullish coalescing operator (`??" + equals + "`) instead of a logical " + description + " (`||" + equals + "`), as it is a safer operator.",
	}
}

func buildPreferNullishOverTernaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverTernary",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
	}
}

func buildPreferNullishOverAssignmentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverAssignment",
		Description: "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
	}
}

// NullishCheckOperator represents the operator used in nullish checks
type NullishCheckOperator string

const (
	OperatorEmpty       NullishCheckOperator = ""
	OperatorNot         NullishCheckOperator = "!"
	OperatorEqual       NullishCheckOperator = "=="
	OperatorStrictEqual NullishCheckOperator = "==="
	OperatorNotEqual    NullishCheckOperator = "!="
	OperatorNotStrictEq NullishCheckOperator = "!=="
)

// isLogicalOrOperator is a helper to check for || or ||= operators
func isLogicalOrOperator(node *ast.Node) bool {
	if !ast.IsBinaryExpression(node) {
		return false
	}
	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken == nil {
		return false
	}
	op := binExpr.OperatorToken.Kind
	return op == ast.KindBarBarToken || op == ast.KindBarBarEqualsToken
}

// isNullLiteral checks if node is null literal
func isNullLiteral(node *ast.Node) bool {
	return node.Kind == ast.KindNullKeyword
}

// isUndefinedIdentifier checks if node is undefined identifier
func isUndefinedIdentifier(node *ast.Node) bool {
	return ast.IsIdentifier(node) && node.Text() == "undefined"
}

// isNullLiteralOrUndefinedIdentifier checks if node is null or undefined
func isNullLiteralOrUndefinedIdentifier(node *ast.Node) bool {
	return isNullLiteral(node) || isUndefinedIdentifier(node)
}

// isMemberAccessLike checks if node is member access like
func isMemberAccessLike(node *ast.Node) bool {
	// Skip parentheses to handle deeply nested patterns like ((((foo.a))))
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	}
	if ast.IsOptionalChain(node) {
		return true
	}
	return false
}

// isNodeEqual checks if two nodes are semantically equal
func isNodeEqual(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case ast.KindIdentifier:
		return a.Text() == b.Text()
	case ast.KindPropertyAccessExpression:
		aProp := a.AsPropertyAccessExpression()
		bProp := b.AsPropertyAccessExpression()
		aName := aProp.Name()
		bName := bProp.Name()
		if aName == nil || bName == nil {
			return aName == bName
		}
		return isNodeEqual(aName.AsNode(), bName.AsNode()) && isNodeEqual(aProp.Expression, bProp.Expression)
	case ast.KindElementAccessExpression:
		aElem := a.AsElementAccessExpression()
		bElem := b.AsElementAccessExpression()
		if aElem.ArgumentExpression == nil || bElem.ArgumentExpression == nil {
			if aElem.ArgumentExpression != bElem.ArgumentExpression {
				return false
			}
		}
		if aElem.Expression == nil || bElem.Expression == nil {
			if aElem.Expression != bElem.Expression {
				return false
			}
		}
		return isNodeEqual(aElem.ArgumentExpression, bElem.ArgumentExpression) && isNodeEqual(aElem.Expression, bElem.Expression)
	case ast.KindNullKeyword:
		return true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return a.Text() == b.Text()
	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return a.Text() == b.Text()
	case ast.KindThisKeyword:
		return true
	}

	return false
}

// areNodesSimilarMemberAccess checks if two nodes have the same member access sequence
func areNodesSimilarMemberAccess(a, b *ast.Node) bool {
	// Unwrap parenthesized expressions
	for a.Kind == ast.KindParenthesizedExpression {
		a = a.AsParenthesizedExpression().Expression
	}
	for b.Kind == ast.KindParenthesizedExpression {
		b = b.AsParenthesizedExpression().Expression
	}

	// Unwrap non-null expressions (like x!)
	if a.Kind == ast.KindNonNullExpression {
		a = a.AsNonNullExpression().Expression
	}
	if b.Kind == ast.KindNonNullExpression {
		b = b.AsNonNullExpression().Expression
	}

	if ast.IsPropertyAccessExpression(a) && ast.IsPropertyAccessExpression(b) {
		aProp := a.AsPropertyAccessExpression()
		bProp := b.AsPropertyAccessExpression()
		if !areNodesSimilarMemberAccess(aProp.Expression, bProp.Expression) {
			return false
		}
		aName := aProp.Name()
		bName := bProp.Name()
		if aName == nil || bName == nil {
			return aName == bName
		}
		return isNodeEqual(aName.AsNode(), bName.AsNode())
	}
	if ast.IsElementAccessExpression(a) && ast.IsElementAccessExpression(b) {
		aElem := a.AsElementAccessExpression()
		bElem := b.AsElementAccessExpression()
		if !areNodesSimilarMemberAccess(aElem.Expression, bElem.Expression) {
			return false
		}
		return isNodeEqual(aElem.ArgumentExpression, bElem.ArgumentExpression)
	}
	return isNodeEqual(a, b)
}

// isConditionalTest checks if the node is part of a conditional test
func isConditionalTest(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	if ast.IsLogicalExpression(parent) {
		return isConditionalTest(parent)
	}

	// Handle parenthesized expressions - traverse up through them
	if parent.Kind == ast.KindParenthesizedExpression {
		return isConditionalTest(parent)
	}

	if ast.IsConditionalExpression(parent) {
		condExpr := parent.AsConditionalExpression()
		if condExpr.WhenTrue == node || condExpr.WhenFalse == node {
			return isConditionalTest(parent)
		}
	}

	if parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		binExpr := parent.AsBinaryExpression()
		if binExpr.Right == node {
			return isConditionalTest(parent)
		}
	}

	if ast.IsPrefixUnaryExpression(parent) && parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return isConditionalTest(parent)
	}

	switch parent.Kind {
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	}

	return false
}

// isBooleanConstructorContext checks if node is argument to Boolean constructor
func isBooleanConstructorContext(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	if ast.IsLogicalExpression(parent) {
		return isBooleanConstructorContext(ctx, parent)
	}

	if ast.IsConditionalExpression(parent) {
		condExpr := parent.AsConditionalExpression()
		if condExpr.WhenTrue == node || condExpr.WhenFalse == node {
			return isBooleanConstructorContext(ctx, parent)
		}
	}

	if parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		binExpr := parent.AsBinaryExpression()
		if binExpr.Right == node {
			return isBooleanConstructorContext(ctx, parent)
		}
	}

	// Check if it's a call to Boolean()
	if ast.IsCallExpression(parent) {
		callExpr := parent.AsCallExpression()
		if ast.IsIdentifier(callExpr.Expression) && callExpr.Expression.Text() == "Boolean" {
			// The Boolean being called - check if it's the global Boolean
			symbol := ctx.TypeChecker.GetSymbolAtLocation(callExpr.Expression)
			if symbol != nil {
				// Check if this is the global Boolean (no user-defined declarations in source files)
				for _, decl := range symbol.Declarations {
					sf := ast.GetSourceFileOfNode(decl)
					if sf != nil && !sf.IsDeclarationFile {
						return false // User-defined Boolean
					}
				}
			}
			return true
		}
	}

	return false
}

var PreferNullishCoalescingRule = rule.Rule{
	Name: "prefer-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferNullishCoalescingOptions](options, "prefer-nullish-coalescing")

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks)

		if !isStrictNullChecks && !opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		// isNullableType checks if a type includes null or undefined
		// Also returns true for any/unknown since they can include null/undefined
		isNullableType := func(t *checker.Type) bool {
			// any and unknown types could be null/undefined
			if utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return true
			}
			for _, part := range utils.UnionTypeParts(t) {
				flags := checker.Type_flags(part)
				if flags&checker.TypeFlagsNullable != 0 {
					return true
				}
			}
			return false
		}

		// isTypeEligibleForPreferNullish checks whether a type tested for truthiness
		// is eligible for conversion to a nullishness check, taking into account the rule's configuration.
		isTypeEligibleForPreferNullish := func(t *checker.Type) bool {
			if !isNullableType(t) {
				return false
			}

			// Build ignorable flags based on options
			var ignorableFlags checker.TypeFlags

			// Handle ignorePrimitives which can be either a boolean or an object
			if opts.IgnorePrimitives != nil {
				switch v := opts.IgnorePrimitives.(type) {
				case bool:
					if v {
						// If true, ignore all primitive types
						ignorableFlags = checker.TypeFlagsBigIntLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsNumberLike | checker.TypeFlagsStringLike
					}
				case map[string]interface{}:
					// It's an object, unmarshal it as IgnorePrimitivesOptions
					jsonBytes, err := json.Marshal(v)
					if err == nil {
						var primitivesOpts IgnorePrimitivesOptions
						if err := json.Unmarshal(jsonBytes, &primitivesOpts); err == nil {
							if primitivesOpts.Bigint {
								ignorableFlags |= checker.TypeFlagsBigIntLike
							}
							if primitivesOpts.Boolean {
								ignorableFlags |= checker.TypeFlagsBooleanLike
							}
							if primitivesOpts.Number {
								ignorableFlags |= checker.TypeFlagsNumberLike
							}
							if primitivesOpts.String {
								ignorableFlags |= checker.TypeFlagsStringLike
							}
						}
					}
				}
			}

			if ignorableFlags == 0 {
				// Any types are eligible for conversion
				return true
			}

			// If the type is `any` or `unknown` we can't make any assumptions
			if utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return false
			}

			// Check if any type constituents match the ignorable flags
			for _, part := range utils.UnionTypeParts(t) {
				for _, intersectionPart := range utils.IntersectionTypeParts(part) {
					if utils.IsTypeFlagSet(intersectionPart, ignorableFlags) {
						return false
					}
				}
			}

			return true
		}

		// isTruthinessCheckEligibleForPreferNullish determines whether a control flow construct
		// that uses the truthiness of a test expression is eligible for conversion
		isTruthinessCheckEligibleForPreferNullish := func(node, testNode *ast.Node) bool {
			testType := ctx.TypeChecker.GetTypeAtLocation(testNode)
			if !isTypeEligibleForPreferNullish(testType) {
				return false
			}

			if opts.IgnoreConditionalTests && isConditionalTest(node) {
				return false
			}

			if opts.IgnoreBooleanCoercion && isBooleanConstructorContext(ctx, node) {
				// For conditional expressions inside Boolean calls, still check
				if !(ast.IsConditionalExpression(node) && ast.IsCallExpression(node.Parent)) {
					return false
				}
			}

			return true
		}

		// shouldSkipDueToPartialNullishCheck determines whether a nullish check with only null or only undefined
		// should be skipped based on the type of the expression. Returns true if the fix should be skipped.
		// The allowNotEqualOperator parameter controls whether !== operator should be treated like == (for ternary vs if-statement handling).
		shouldSkipDueToPartialNullishCheck := func(hasNullCheck, hasUndefinedCheck bool, operator NullishCheckOperator, targetNode *ast.Node, allowNotEqualOperator bool) bool {
			// If we check for both null and undefined, it's always fixable
			if hasNullCheck && hasUndefinedCheck {
				return false
			}

			// If we check for neither null nor undefined, skip
			if !hasNullCheck && !hasUndefinedCheck {
				return true
			}

			// Only one check - need to verify the operator and type
			// For == and != operators, loose equality handles both null and undefined
			if operator == OperatorEqual {
				return false
			}
			if allowNotEqualOperator && operator == OperatorNotEqual {
				return false
			}

			t := ctx.TypeChecker.GetTypeAtLocation(targetNode)
			flags := checker.Type_flags(t)

			// Skip if the type is any or unknown
			if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
				return true
			}

			// Check for null/undefined in union parts
			hasNullType := false
			hasUndefinedType := false
			for _, part := range utils.UnionTypeParts(t) {
				partFlags := checker.Type_flags(part)
				if partFlags&checker.TypeFlagsNull != 0 {
					hasNullType = true
				}
				if partFlags&checker.TypeFlagsUndefined != 0 {
					hasUndefinedType = true
				}
			}

			// If checking for undefined but type includes null, skip (incomplete nullish check)
			if hasUndefinedCheck && hasNullType {
				return true
			}
			// If checking for null but type includes undefined, skip (incomplete nullish check)
			if hasNullCheck && hasUndefinedType {
				return true
			}

			return false
		}

		// isMixedLogicalExpression checks if node is part of a mixed logical expression (with &&)
		isMixedLogicalExpression := func(node *ast.Node) bool {
			seen := &utils.Set[*ast.Node]{}
			queue := []*ast.Node{node.Parent, node.AsBinaryExpression().Left, node.AsBinaryExpression().Right}

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]

				if current == nil || seen.Has(current) {
					continue
				}
				seen.Add(current)

				if ast.IsLogicalExpression(current) {
					// Skip parentheses to get to the actual binary expression
					unwrapped := ast.SkipParentheses(current)
					if !ast.IsBinaryExpression(unwrapped) {
						continue
					}
					binExpr := unwrapped.AsBinaryExpression()
					if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
						return true
					}
					if binExpr.OperatorToken.Kind == ast.KindBarBarToken ||
						(binExpr.OperatorToken.Kind == ast.KindEqualsToken && ast.IsLogicalExpression(binExpr.Left)) {
						queue = append(queue, current.Parent, binExpr.Left, binExpr.Right)
					}
				}
			}

			return false
		}

		// checkAndFixWithPreferNullishOverOr handles || and ||= operators
		checkAndFixWithPreferNullishOverOr := func(node *ast.Node, description, equals string) {
			binExpr := node.AsBinaryExpression()

			if !isTruthinessCheckEligibleForPreferNullish(node, binExpr.Left) {
				return
			}

			if opts.IgnoreMixedLogicalExpressions && isMixedLogicalExpression(node) {
				return
			}

			ctx.ReportNodeWithFixes(binExpr.OperatorToken, buildPreferNullishOverOrMessage(description, equals), func() []rule.RuleFix {
				fixes := []rule.RuleFix{}

				// If parent is a logical or expression (skipping parentheses), wrap with parentheses
				// But don't add parentheses if already wrapped in parentheses
				parentNode := node.Parent
				for parentNode != nil && ast.IsParenthesizedExpression(parentNode) {
					parentNode = parentNode.Parent
				}
				if parentNode != nil && ast.IsLogicalExpression(parentNode) {
					// Check if the node is already wrapped in parentheses
					if !ast.IsParenthesizedExpression(node.Parent) {
						leftExpr := binExpr.Left
						if ast.IsLogicalExpression(leftExpr) && !isLogicalOrOperator(leftExpr.AsBinaryExpression().Left) {
							fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, leftExpr.AsBinaryExpression().Right, "("))
						} else {
							fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, binExpr.Left, "("))
						}
						fixes = append(fixes, rule.RuleFixInsertAfter(binExpr.Right, ")"))
					}
				}

				// Replace || with ?? or ||= with ??=
				var newOperator string
				if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					newOperator = "??"
				} else {
					newOperator = "??="
				}
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, binExpr.OperatorToken, newOperator))

				return fixes
			})
		}

		// getOperatorAndNodesInsideTestExpression analyzes test expression
		getOperatorAndNodesInsideTestExpression := func(node *ast.Node) (NullishCheckOperator, []*ast.Node) {
			var test *ast.Node
			if ast.IsConditionalExpression(node) {
				test = node.AsConditionalExpression().Condition
			} else if ast.IsIfStatement(node) {
				test = node.AsIfStatement().Expression
			} else {
				return "", nil
			}

			test = ast.SkipParentheses(test)

			// Check for simple truthiness or negation check
			// Return empty slice (not nil) to indicate valid truthiness check
			if isMemberAccessLike(test) {
				return OperatorEmpty, []*ast.Node{}
			}
			if ast.IsPrefixUnaryExpression(test) && test.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
				arg := ast.SkipParentheses(test.AsPrefixUnaryExpression().Operand)
				if isMemberAccessLike(arg) {
					return OperatorNot, []*ast.Node{}
				}
				return "", nil
			}

			// Check for binary comparison
			if ast.IsBinaryExpression(test) {
				binExpr := test.AsBinaryExpression()
				left := ast.SkipParentheses(binExpr.Left)
				right := ast.SkipParentheses(binExpr.Right)

				switch binExpr.OperatorToken.Kind {
				case ast.KindEqualsEqualsToken:
					return OperatorEqual, []*ast.Node{left, right}
				case ast.KindExclamationEqualsToken:
					return OperatorNotEqual, []*ast.Node{left, right}
				case ast.KindEqualsEqualsEqualsToken:
					return OperatorStrictEqual, []*ast.Node{left, right}
				case ast.KindExclamationEqualsEqualsToken:
					return OperatorNotStrictEq, []*ast.Node{left, right}
				case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
					// Compound check like (a === null || a === undefined)
					if ast.IsBinaryExpression(binExpr.Left) && ast.IsBinaryExpression(binExpr.Right) {
						leftBin := binExpr.Left.AsBinaryExpression()
						rightBin := binExpr.Right.AsBinaryExpression()

						// Check if one side is a simple nullish comparison (null === null or undefined === undefined)
						leftIsNullishComparison := isNullLiteralOrUndefinedIdentifier(ast.SkipParentheses(leftBin.Left)) &&
							isNullLiteralOrUndefinedIdentifier(ast.SkipParentheses(leftBin.Right))
						rightIsNullishComparison := isNullLiteralOrUndefinedIdentifier(ast.SkipParentheses(rightBin.Left)) &&
							isNullLiteralOrUndefinedIdentifier(ast.SkipParentheses(rightBin.Right))

						if leftIsNullishComparison || rightIsNullishComparison {
							return "", nil
						}

						nodes := []*ast.Node{
							ast.SkipParentheses(leftBin.Left),
							ast.SkipParentheses(leftBin.Right),
							ast.SkipParentheses(rightBin.Left),
							ast.SkipParentheses(rightBin.Right),
						}

						if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
							if leftBin.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken &&
								rightBin.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken {
								return OperatorStrictEqual, nodes
							}
							if (leftBin.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
								leftBin.OperatorToken.Kind == ast.KindEqualsEqualsToken) &&
								(rightBin.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
									rightBin.OperatorToken.Kind == ast.KindEqualsEqualsToken) {
								return OperatorEqual, nodes
							}
						} else if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
							if leftBin.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken &&
								rightBin.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
								return OperatorNotStrictEq, nodes
							}
							if (leftBin.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken ||
								leftBin.OperatorToken.Kind == ast.KindExclamationEqualsToken) &&
								(rightBin.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken ||
									rightBin.OperatorToken.Kind == ast.KindExclamationEqualsToken) {
								return OperatorNotEqual, nodes
							}
						}
					}
				}
			}

			return "", nil
		}

		// getBranchNodes returns the branch nodes of a conditional expression
		getBranchNodes := func(node *ast.Node, operator NullishCheckOperator) (nonNullishBranch, nullishBranch *ast.Node) {
			condExpr := node.AsConditionalExpression()
			if operator == OperatorEmpty || operator == OperatorNotEqual || operator == OperatorNotStrictEq {
				return condExpr.WhenTrue, condExpr.WhenFalse
			}
			return condExpr.WhenFalse, condExpr.WhenTrue
		}

		return rule.RuleListeners{
			// Handle || operator (LogicalExpression)
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()

				// Handle ||
				if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					checkAndFixWithPreferNullishOverOr(node, "or", "")
					return
				}

				// Handle ||=
				if binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken {
					checkAndFixWithPreferNullishOverOr(node, "assignment", "=")
					return
				}
			},

			// Handle ternary expressions (ConditionalExpression)
			ast.KindConditionalExpression: func(node *ast.Node) {
				if opts.IgnoreTernaryTests {
					return
				}

				operator, nodesInsideTest := getOperatorAndNodesInsideTestExpression(node)
				if operator == "" && nodesInsideTest == nil {
					return
				}

				nonNullishBranch, nullishBranch := getBranchNodes(node, operator)

				// Skip parentheses for comparison
				nonNullishBranchUnwrapped := ast.SkipParentheses(nonNullishBranch)

				// Determine the nullish coalescing left node
				var nullishCoalescingLeftNode *ast.Node
				hasTruthinessCheck := false

				if len(nodesInsideTest) == 0 {
					// Simple truthiness check
					hasTruthinessCheck = true
					condExpr := node.AsConditionalExpression()
					testNode := ast.SkipParentheses(condExpr.Condition)
					if ast.IsPrefixUnaryExpression(testNode) && testNode.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
						nullishCoalescingLeftNode = ast.SkipParentheses(testNode.AsPrefixUnaryExpression().Operand)
					} else {
						nullishCoalescingLeftNode = testNode
					}

					if !areNodesSimilarMemberAccess(nullishCoalescingLeftNode, nonNullishBranchUnwrapped) {
						return
					}
				} else {
					// Check that the test only contains null, undefined and the identifier
					hasNullCheck := false
					hasUndefinedCheck := false

					for _, testNode := range nodesInsideTest {
						if isNullLiteral(testNode) {
							hasNullCheck = true
						} else if isUndefinedIdentifier(testNode) {
							hasUndefinedCheck = true
						} else if areNodesSimilarMemberAccess(testNode, nonNullishBranchUnwrapped) {
							if nullishCoalescingLeftNode == nil {
								nullishCoalescingLeftNode = testNode
							}
						} else {
							return
						}
					}

					if nullishCoalescingLeftNode == nil {
						return
					}

					// Check if fixable
					if !hasTruthinessCheck {
						if shouldSkipDueToPartialNullishCheck(hasNullCheck, hasUndefinedCheck, operator, nullishCoalescingLeftNode, true) {
							return
						}
					}
				}

				if hasTruthinessCheck && !isTruthinessCheckEligibleForPreferNullish(node, nullishCoalescingLeftNode) {
					return
				}

				// Report the issue
				ctx.ReportNodeWithFixes(node, buildPreferNullishOverTernaryMessage(), func() []rule.RuleFix {
					leftText := ctx.SourceFile.Text()[nullishCoalescingLeftNode.Pos():nullishCoalescingLeftNode.End()]
					rightText := ctx.SourceFile.Text()[nullishBranch.Pos():nullishBranch.End()]

					// Trim whitespace before processing
					rightText = strings.TrimSpace(rightText)

					// Wrap right side if needed
					if !utils.IsStrongPrecedenceNode(nullishBranch) && !ast.IsParenthesizedExpression(nullishBranch) {
						rightText = "(" + rightText + ")"
					}

					newText := strings.TrimSpace(leftText) + " ?? " + rightText
					return []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, node, newText),
					}
				})
			},

			// Handle if statements
			ast.KindIfStatement: func(node *ast.Node) {
				if opts.IgnoreIfStatements {
					return
				}

				ifStmt := node.AsIfStatement()
				if ifStmt.ElseStatement != nil {
					return
				}

				// Get the assignment expression from the consequent
				var assignmentExpr *ast.Node
				if ast.IsBlock(ifStmt.ThenStatement) {
					block := ifStmt.ThenStatement.AsBlock()
					if block.Statements != nil && len(block.Statements.Nodes) == 1 {
						stmt := block.Statements.Nodes[0]
						if ast.IsExpressionStatement(stmt) {
							assignmentExpr = stmt.AsExpressionStatement().Expression
						}
					}
				} else if ast.IsExpressionStatement(ifStmt.ThenStatement) {
					assignmentExpr = ifStmt.ThenStatement.AsExpressionStatement().Expression
				}

				if assignmentExpr == nil {
					return
				}

				// Skip parentheses around the assignment expression
				// to handle cases like ((((foo.a)))) = value;
				assignmentExprUnwrapped := ast.SkipParentheses(assignmentExpr)
				if !ast.IsAssignmentExpression(assignmentExprUnwrapped, false) {
					return
				}

				binExpr := assignmentExprUnwrapped.AsBinaryExpression()
				if !isMemberAccessLike(binExpr.Left) {
					return
				}

				nullishCoalescingLeftNode := binExpr.Left
				nullishCoalescingRightNode := binExpr.Right

				operator, nodesInsideTest := getOperatorAndNodesInsideTestExpression(node)
				if operator == "" {
					return
				}

				// Only handle negation or equality checks
				if operator != OperatorNot && operator != OperatorEqual && operator != OperatorStrictEqual {
					return
				}

				// Verify the test is checking the same variable being assigned
				if len(nodesInsideTest) == 0 {
					// Simple negation check
					testNode := ast.SkipParentheses(ifStmt.Expression)
					if ast.IsPrefixUnaryExpression(testNode) {
						testNode = ast.SkipParentheses(testNode.AsPrefixUnaryExpression().Operand)
					}
					if !areNodesSimilarMemberAccess(testNode, nullishCoalescingLeftNode) {
						return
					}
					if !isTruthinessCheckEligibleForPreferNullish(node, testNode) {
						return
					}
				} else {
					// Check null/undefined comparison
					hasNullCheck := false
					hasUndefinedCheck := false
					foundMatchingNode := false

					for _, testNode := range nodesInsideTest {
						if isNullLiteral(testNode) {
							hasNullCheck = true
						} else if isUndefinedIdentifier(testNode) {
							hasUndefinedCheck = true
						} else if areNodesSimilarMemberAccess(testNode, nullishCoalescingLeftNode) {
							foundMatchingNode = true
						} else {
							return
						}
					}

					if !foundMatchingNode {
						return
					}

					// Check if fixable
					if shouldSkipDueToPartialNullishCheck(hasNullCheck, hasUndefinedCheck, operator, nullishCoalescingLeftNode, false) {
						return
					}
				}

				// Report the issue
				ctx.ReportNodeWithFixes(node, buildPreferNullishOverAssignmentMessage(), func() []rule.RuleFix {
					// Strip all outer parentheses from the left node to get the inner expression
					leftNodeUnwrapped := ast.SkipParentheses(nullishCoalescingLeftNode)
					leftText := ctx.SourceFile.Text()[leftNodeUnwrapped.Pos():leftNodeUnwrapped.End()]
					rightText := ctx.SourceFile.Text()[nullishCoalescingRightNode.Pos():nullishCoalescingRightNode.End()]

					leftTextTrimmed := strings.TrimSpace(leftText)
					// Only wrap in parentheses if the original expression was parenthesized
					if ast.IsParenthesizedExpression(nullishCoalescingLeftNode) {
						leftTextTrimmed = "(" + leftTextTrimmed + ")"
					}
					newText := leftTextTrimmed + " ??= " + strings.TrimSpace(rightText) + ";"

					return []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, node, newText),
					}
				})
			},
		}
	},
}
