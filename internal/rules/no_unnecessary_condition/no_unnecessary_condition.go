package no_unnecessary_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type NoUnnecessaryConditionOptions struct {
	// Can be: "never" | "always" | "only-allowed-literals" | boolean (or pointers to these types)
	AllowConstantLoopConditions                            any
	CheckTypePredicates                                    *bool
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool
}

func buildAlwaysTruthyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthy",
		Description: "Unnecessary conditional, value is always truthy.",
	}
}

func buildAlwaysFalsyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsy",
		Description: "Unnecessary conditional, value is always falsy.",
	}
}

func buildAlwaysTruthyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthyFunc",
		Description: "This callback should return a conditional, but return is always truthy.",
	}
}

func buildAlwaysFalsyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsyFunc",
		Description: "This callback should return a conditional, but return is always falsy.",
	}
}

func buildNeverNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverNullish",
		Description: "Unnecessary optional chain on a non-nullish value.",
	}
}

func buildNeverOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverOptionalChain",
		Description: "Unnecessary optional chain on a non-nullish value.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

var NoUnnecessaryConditionRule = rule.Rule{
	Name: "no-unnecessary-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoUnnecessaryConditionOptions)
		if !ok {
			opts = NoUnnecessaryConditionOptions{}
		}
		if opts.AllowConstantLoopConditions == nil {
			opts.AllowConstantLoopConditions = "never"
		}
		if opts.CheckTypePredicates == nil {
			opts.CheckTypePredicates = utils.Ref(false)
		}
		
		// https://typescript-eslint.io/rules/no-unnecessary-condition/#:~:text=Default%3A%20false.-,DEPRECATED,-This%20option%20will
		// TLDR: This option will be removed in the next major version of typescript-eslint.
		if opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(false)
		}

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		// Parse AllowConstantLoopConditions which can be string, *string, or boolean
		var loopConditionMode string
		switch v := opts.AllowConstantLoopConditions.(type) {
		case string:
			loopConditionMode = v
		case *string:
			if v != nil {
				loopConditionMode = *v
			} else {
				loopConditionMode = "never"
			}
		case bool:
			if v {
				loopConditionMode = "always"
			} else {
				loopConditionMode = "never"
			}
		default:
			loopConditionMode = "never"
		}

		isAlwaysConstantLoopCondition := loopConditionMode == "always"
		isAllowedConstantLoopCondition := loopConditionMode == "only-allowed-literals"

		getResolvedType := func(node *ast.Node) *checker.Type {
			nodeType := ctx.TypeChecker.GetTypeAtLocation(node)
			if nodeType == nil {
				return nil
			}

			constraintType, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, nodeType)
			if isTypeParameter && constraintType == nil {
				return nil
			}
			if isTypeParameter {
				return constraintType
			}

			return nodeType
		}

		isLiteralBoolean := func(node *ast.Node) bool {
			skipNode := ast.SkipParentheses(node)
			return skipNode.Kind == ast.KindTrueKeyword || skipNode.Kind == ast.KindFalseKeyword
		}

		checkCondition := func(node *ast.Node) {
			// Skip negation expressions - they're handled by KindPrefixUnaryExpression listener
			skipNode := ast.SkipParentheses(node)
			if skipNode.Kind == ast.KindPrefixUnaryExpression {
				unaryExpr := skipNode.AsPrefixUnaryExpression()
				if unaryExpr.Operator == ast.KindExclamationToken {
					return
				}
			}

			// Check literal boolean keywords first
			if isLiteralBoolean(node) {
				if skipNode.Kind == ast.KindTrueKeyword {
					ctx.ReportNode(node, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(node, buildAlwaysFalsyMessage())
				}
				return
			}

			nodeType := getResolvedType(node)
			if nodeType == nil {
				return
			}

			isTruthy, isFalsy := checkTypeCondition(ctx.TypeChecker, nodeType)
			if isTruthy {
				ctx.ReportNode(node, buildAlwaysTruthyMessage())
			} else if isFalsy {
				ctx.ReportNode(node, buildAlwaysFalsyMessage())
			}
		}

		checkOptionalChain := func(node *ast.Node) {
			var expression *ast.Node
			var hasQuestionDot bool

			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				expression = propAccess.Expression
				hasQuestionDot = propAccess.QuestionDotToken != nil
			case ast.KindElementAccessExpression:
				elemAccess := node.AsElementAccessExpression()
				expression = elemAccess.Expression
				hasQuestionDot = elemAccess.QuestionDotToken != nil
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				expression = callExpr.Expression
				hasQuestionDot = callExpr.QuestionDotToken != nil
			default:
				return
			}

			if !hasQuestionDot {
				return
			}

			exprType := getResolvedType(expression)
			if exprType == nil {
				return
			}

			if !isNullishType(ctx.TypeChecker, exprType) {
				ctx.ReportNode(node, buildNeverOptionalChainMessage())
			}
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				checkCondition(node.AsIfStatement().Expression)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				if isAlwaysConstantLoopCondition {
					return
				}
				whileStmt := node.AsWhileStatement()
				if isAllowedConstantLoopCondition && isAllowedConstantLiteral(whileStmt.Expression) {
					return
				}
				checkCondition(whileStmt.Expression)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				if isAlwaysConstantLoopCondition {
					return
				}
				doStmt := node.AsDoStatement()
				if isAllowedConstantLoopCondition && isAllowedConstantLiteral(doStmt.Expression) {
					return
				}
				checkCondition(doStmt.Expression)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt.Condition == nil {
					return
				}
				if isAlwaysConstantLoopCondition {
					return
				}
				if isAllowedConstantLoopCondition && isAllowedConstantLiteral(forStmt.Condition) {
					return
				}
				checkCondition(forStmt.Condition)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkCondition(node.AsConditionalExpression().Condition)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
					binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					checkCondition(binExpr.Left)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				unaryExpr := node.AsPrefixUnaryExpression()
				if unaryExpr.Operator == ast.KindExclamationToken {
					checkCondition(unaryExpr.Operand)
				}
			},
			ast.KindPropertyAccessExpression: checkOptionalChain,
			ast.KindElementAccessExpression:  checkOptionalChain,
			ast.KindCallExpression: func(node *ast.Node) {
				checkOptionalChain(node)

				if !*opts.CheckTypePredicates {
					return
				}

				callExpr := node.AsCallExpression()
				if !utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) {
					return
				}

				if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
					if arg := callExpr.Arguments.Nodes[0]; arg != nil {
						checkPredicateFunction(ctx, arg)
					}
				}
			},
		}
	},
}

// Helper function to check if a type is always truthy or always falsy
func checkTypeCondition(typeChecker *checker.Checker, t *checker.Type) (isTruthy bool, isFalsy bool) {
	flags := checker.Type_flags(t)

	// Handle unions - check all parts
	if utils.IsUnionType(t) {
		allTruthy := true
		allFalsy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(typeChecker, part)
			if !partTruthy {
				allTruthy = false
			}
			if !partFalsy {
				allFalsy = false
			}
		}

		return allTruthy, allFalsy
	}

	// Nullish types are always falsy
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false, true
	}

	// Objects and non-primitive types are always truthy
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive) != 0 {
		return true, false
	}

	// ESSymbol is always truthy
	if flags&(checker.TypeFlagsESSymbol|checker.TypeFlagsUniqueESSymbol) != 0 {
		return true, false
	}

	// Boolean literals - check flags first
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		// Boolean literal types can be intrinsic or fresh literal types
		// Check if it's an intrinsic type first
		if utils.IsIntrinsicType(t) {
			intrinsicName := t.AsIntrinsicType().IntrinsicName()
			if intrinsicName == "true" {
				return true, false
			}
			if intrinsicName == "false" {
				return false, true
			}
		} else if t.AsLiteralType() != nil {
			// For fresh literal types, check via AsLiteralType
			litStr := t.AsLiteralType().String()
			if litStr == "true" {
				return true, false
			}
			if litStr == "false" {
				return false, true
			}
		}
	}

	// String literals
	if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if literal.Value() == "" {
				return false, true
			}
			return true, false
		}
	}

	// Number literals
	if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			value := literal.String()
			if value == "0" || value == "NaN" {
				return false, true
			}
			return true, false
		}
	}

	// BigInt literals
	if flags&checker.TypeFlagsBigIntLiteral != 0 && t.IsBigIntLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if literal.String() == "0" || literal.String() == "0n" {
				return false, true
			}
			return true, false
		}
	}

	// Generic types (boolean, string, number, etc.) are not always truthy or falsy
	return false, false
}

// Check if a type can be nullish
func isNullishType(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, part := range t.Types() {
			if isNullishType(typeChecker, part) {
				return true
			}
		}
		return false
	}

	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
}

// Check if expression is an allowed constant literal (true, false, 0, 1)
func isAllowedConstantLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindNumericLiteral:
		literal := node.AsNumericLiteral()
		text := literal.Text
		return text == "0" || text == "1"
	}

	return false
}

// Check type predicate functions for unnecessary conditions
func checkPredicateFunction(ctx rule.RuleContext, funcNode *ast.Node) {
	isFunction := funcNode.Kind&(ast.KindArrowFunction|ast.KindFunctionExpression|ast.KindFunctionDeclaration) != 0
	if !isFunction {
		return
	}

	funcType := ctx.TypeChecker.GetTypeAtLocation(funcNode)
	signatures := ctx.TypeChecker.GetCallSignatures(funcType)

	for _, signature := range signatures {
		returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signature)

		// Handle type parameters
		typeFlags := checker.Type_flags(returnType)
		if typeFlags&checker.TypeFlagsTypeParameter != 0 {
			constraint := ctx.TypeChecker.GetConstraintOfTypeParameter(returnType)
			if constraint != nil {
				returnType = constraint
			}
		}

		isTruthy, isFalsy := checkTypeCondition(ctx.TypeChecker, returnType)

		if isTruthy {
			ctx.ReportNode(funcNode, buildAlwaysTruthyFuncMessage())
		} else if isFalsy {
			ctx.ReportNode(funcNode, buildAlwaysFalsyFuncMessage())
		}
	}
}
