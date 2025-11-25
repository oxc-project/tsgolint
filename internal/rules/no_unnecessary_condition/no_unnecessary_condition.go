// Package no_unnecessary_condition implements the no-unnecessary-condition rule.
//
// This rule prevents unnecessary conditions in TypeScript code by detecting expressions
// that are always truthy, always falsy, or comparing values that have no overlap.
//
// The rule checks:
// - Conditional expressions (if, while, for, ternary operators)
// - Logical operators (&&, ||, !)
// - Nullish coalescing operators (??, ??=)
// - Optional chaining (?.)
// - Comparison operators (===, !==, ==, !=, <, >, <=, >=)
// - Type predicates and type guards
//
// This implementation is based on the @typescript-eslint/no-unnecessary-condition rule:
// https://typescript-eslint.io/rules/no-unnecessary-condition/
package no_unnecessary_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// NoUnnecessaryConditionOptions configures the no-unnecessary-condition rule.
type NoUnnecessaryConditionOptions struct {
	// AllowConstantLoopConditions controls whether constant loop conditions are allowed.
	// Values: "never" (default) | "always" | "only-allowed-literals" | boolean
	// - "never": Disallow all constant loop conditions
	// - "always": Allow all constant loop conditions
	// - "only-allowed-literals": Allow only literal true/false/0/1
	AllowConstantLoopConditions any

	// CheckTypePredicates enables checking of type predicate functions.
	// When true, reports when type guards are used on values that already match the predicate type.
	// Default: false
	CheckTypePredicates *bool

	// AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing allows the rule to run
	// without strictNullChecks enabled. Not recommended.
	// Default: false (DEPRECATED - will be removed in future versions)
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

func buildNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "never",
		Description: "Unnecessary conditional, value is `never`.",
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

func buildLiteralBinaryExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparisonBetweenLiteralTypes",
		Description: "Unnecessary comparison between literal values.",
	}
}

func buildNoOverlapMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noOverlapBooleanExpression",
		Description: "This condition will always return the same value since the types have no overlap.",
	}
}

func buildAlwaysNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysNullish",
		Description: "Unnecessary conditional, value is always nullish.",
	}
}

func buildTypeGuardAlreadyIsTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeGuardAlreadyIsType",
		Description: "Type predicate is unnecessary as the parameter type already satisfies the predicate.",
	}
}

// isIndeterminateType checks if a type cannot be determined at compile time.
//
// Indeterminate types include:
// - any: explicitly typed as any
// - unknown: could be anything
// - type parameters: generic types like T, K
// - indexed access types: types like T[K]
// - index types: types like keyof T
//
// For these types, we cannot determine their truthiness, nullishness, or overlap
// with other types at compile time, so we conservatively avoid reporting them.
func isIndeterminateType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess|checker.TypeFlagsIndex) != 0
}

// isAlwaysNullishType checks if a type is always null, undefined, or void.
//
// Returns true for types that can only be nullish values:
// - null
// - undefined
// - void (treated as undefined at runtime)
//
// Returns false for:
// - Non-nullish types (string, number, object, etc.)
// - Unions containing non-nullish types (string | null, number | undefined)
//
// Note: This is different from isNullishType which returns true if a type
// CAN BE nullish (including unions like string | null).
func isAlwaysNullishType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
}

// isSameExpression checks if two AST nodes represent the same expression.
//
// This is used for control flow narrowing to detect when the same expression
// is checked multiple times (e.g., `arr[42] && arr[42]`).
//
// Examples of same expressions:
// - `foo` and `foo` (same identifier)
// - `arr[42]` and `arr[42]` (same array access with same index)
// - `obj.prop` and `obj.prop` (same property access)
//
// Note: This is a shallow comparison and doesn't handle all cases.
// For complex expressions or expressions with side effects, it may return false negatives.
func isSameExpression(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}

	// Must be same kind
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case ast.KindIdentifier:
		// Compare identifier names
		aId := a.AsIdentifier()
		bId := b.AsIdentifier()
		return aId.Text == bId.Text

	case ast.KindPropertyAccessExpression:
		// Compare obj.prop
		aProp := a.AsPropertyAccessExpression()
		bProp := b.AsPropertyAccessExpression()
		// Check if base expressions are the same and property names match
		if !isSameExpression(aProp.Expression, bProp.Expression) {
			return false
		}
		aName := aProp.Name()
		bName := bProp.Name()
		if aName == nil || bName == nil {
			return false
		}
		return ast.GetTextOfPropertyName(aName) == ast.GetTextOfPropertyName(bName)

	case ast.KindElementAccessExpression:
		// Compare arr[index]
		aElem := a.AsElementAccessExpression()
		bElem := b.AsElementAccessExpression()
		// Check if base expressions and argument expressions are the same
		return isSameExpression(aElem.Expression, bElem.Expression) &&
			isSameExpression(aElem.ArgumentExpression, bElem.ArgumentExpression)

	case ast.KindNumericLiteral:
		// Compare numeric literals
		aLit := a.AsNumericLiteral()
		bLit := b.AsNumericLiteral()
		return aLit.Text == bLit.Text

	case ast.KindStringLiteral:
		// Compare string literals
		aLit := a.AsStringLiteral()
		bLit := b.AsStringLiteral()
		return aLit.Text == bLit.Text

	case ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		// Same keywords are always equal
		return true

	default:
		// For other expression types, we conservatively return false
		return false
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

		// Parse AllowConstantLoopConditions which can be string, *string, bool, or *bool
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
		case *bool:
			if v != nil && *v {
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

			// Skip array/tuple element access without noUncheckedIndexedAccess
			//
			// TypeScript's type system has a soundness hole with array element access:
			// Without noUncheckedIndexedAccess, arr[i] has type T instead of T|undefined,
			// even though accessing an out-of-bounds index returns undefined at runtime.
			//
			// Examples:
			//   const arr: number[] = [1, 2, 3]
			//   const x = arr[10]  // type: number, runtime value: undefined
			//
			//   With noUncheckedIndexedAccess:
			//   const x = arr[10]  // type: number | undefined (correct)
			//
			// We only skip this check for actual array/tuple types. Object types with index
			// signatures (like Record<string, T>) are still checked because they don't have
			// this soundness issue.
			if skipNode.Kind == ast.KindElementAccessExpression && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
				elemAccess := skipNode.AsElementAccessExpression()
				baseType := getResolvedType(elemAccess.Expression)
				if baseType != nil {
					// Check if it's a tuple type (e.g., [number, string])
					if checker.IsTupleType(baseType) {
						return
					}
					// Check if it's an array type (e.g., number[], Array<string>)
					if baseType.Symbol() != nil {
						symbolName := baseType.Symbol().Name
						// Array and ReadonlyArray are the built-in array type symbols
						if symbolName == "Array" || symbolName == "ReadonlyArray" {
							return
						}
					}
				}
			}

			isTruthy, isFalsy := checkTypeCondition(ctx.TypeChecker, nodeType)
			if isTruthy {
				ctx.ReportNode(node, buildAlwaysTruthyMessage())
			} else if isFalsy {
				// Check if it's specifically the never type
				flags := checker.Type_flags(nodeType)
				if flags&checker.TypeFlagsNever != 0 {
					ctx.ReportNode(node, buildNeverMessage())
				} else {
					ctx.ReportNode(node, buildAlwaysFalsyMessage())
				}
			}
		}

		// checkOptionalChain validates optional chaining (?.) to detect unnecessary usage.
		//
		// Optional chaining is unnecessary when the expression being accessed is never nullish.
		// This function handles the complexity of chained optional access like foo?.bar?.baz.
		//
		// Examples:
		//   const obj: { foo: string } = { foo: "hello" }
		//   obj?.foo  // unnecessary - obj is never nullish
		//
		//   const obj: { foo: { bar: string } } | null = getObj()
		//   obj?.foo?.bar  // first ?. is fine, but second ?. is unnecessary
		//                  // because when obj exists, obj.foo is never nullish
		//
		// Algorithm:
		// 1. Extract the expression being accessed (e.g., for foo?.bar, extract foo)
		// 2. For chained access (foo?.bar?.baz), we need to check the intermediate type:
		//    - Get the type of foo (excluding nullish parts)
		//    - Check if foo.bar can be nullish (not foo?.bar)
		// 3. For simple access (foo?.bar), check if foo can be nullish
		// 4. Allow indeterminate types (any, unknown, T, T[K]) since we can't determine nullishness
		checkOptionalChain := func(node *ast.Node) {
			var expression *ast.Node
			var hasQuestionDot bool

			// Extract the expression and check if this is optional chaining
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

			// Helper function to check if a type is an array or tuple type
			var isArrayOrTupleType func(*checker.Type) bool
			isArrayOrTupleType = func(t *checker.Type) bool {
				if t == nil {
					return false
				}

				// Check for tuple type
				if checker.IsTupleType(t) {
					return true
				}

				// Check for union type - if any constituent is an array/tuple, return true
				if utils.IsUnionType(t) {
					for _, part := range t.Types() {
						if isArrayOrTupleType(part) {
							return true
						}
					}
					return false
				}

				// Check for array type by symbol name
				if t.Symbol() != nil {
					symbolName := t.Symbol().Name
					if symbolName == "Array" || symbolName == "ReadonlyArray" {
						return true
					}
				}

				return false
			}

			// Helper function to check if an expression chain contains array element access
			// without noUncheckedIndexedAccess, which "infects" the entire chain
			var hasArrayAccessInChain func(*ast.Node) bool
			hasArrayAccessInChain = func(expr *ast.Node) bool {
				if expr == nil {
					return false
				}

				// Check if this expression is an array element access
				if expr.Kind == ast.KindElementAccessExpression && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
					elemAccess := expr.AsElementAccessExpression()
					baseType := getResolvedType(elemAccess.Expression)
					if baseType != nil && isArrayOrTupleType(baseType) {
						return true
					}
				}

				// Recursively check the base expression
				switch expr.Kind {
				case ast.KindPropertyAccessExpression:
					return hasArrayAccessInChain(expr.AsPropertyAccessExpression().Expression)
				case ast.KindElementAccessExpression:
					return hasArrayAccessInChain(expr.AsElementAccessExpression().Expression)
				case ast.KindCallExpression:
					return hasArrayAccessInChain(expr.AsCallExpression().Expression)
				}

				return false
			}

			// Check if expression or any part of the chain contains array element access
			// e.g., arr[42]?.value or arr[42]?.x?.y?.z
			// The array access "infects" the entire chain because arr[42] might be undefined
			if hasArrayAccessInChain(expression) {
				return
			}

			// Check if the expression is itself an optional chain (chained access)
			// For foo?.bar?.baz, when checking the second ?.:
			//   - node is foo?.bar?.baz
			//   - expression is foo?.bar
			//   - baseExpression is foo
			var baseExpression *ast.Node
			var isChainedAccess bool

			switch expression.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := expression.AsPropertyAccessExpression()
				if propAccess.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = propAccess.Expression
				}
			case ast.KindElementAccessExpression:
				elemAccess := expression.AsElementAccessExpression()
				if elemAccess.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = elemAccess.Expression
				}
			case ast.KindCallExpression:
				callExpr := expression.AsCallExpression()
				if callExpr.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = callExpr.Expression
				}
			}

			// Get the type that would result from the access (if it succeeds)
			// For optional chains, TypeScript gives us the type including | undefined
			// But we want to check if the property itself can be nullish, not including
			// the undefined from the optional chain short-circuit
			var exprType *checker.Type
			if isChainedAccess {
				// For chained access like foo?.bar?.baz, when checking the second ?.:
				// expression is foo?.bar
				// We want to check if bar can be nullish (from foo's perspective)

				// Get foo's type (non-nullish parts)
				baseType := getResolvedType(baseExpression)
				if baseType == nil {
					return
				}

				nonNullishBase := removeNullishFromType(ctx.TypeChecker, baseType)
				if nonNullishBase == nil {
					return
				}

				// Get the type of bar from foo's (non-nullish) type
				// For PropertyAccessExpression, we can get the property name
				if expression.Kind == ast.KindPropertyAccessExpression {
					propAccess := expression.AsPropertyAccessExpression()
					nameNode := propAccess.Name()
					if nameNode != nil {
						propName := ast.GetTextOfPropertyName(nameNode)
						if propName != "" {
							// Get the property type from the non-nullish base
							prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, propName)
							if prop != nil {
								exprType = ctx.TypeChecker.GetTypeOfSymbol(prop)
							} else {
								// Property doesn't exist, might be index signature
								exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
							}
						} else {
							exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
						}
					} else {
						exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
					}
				} else if expression.Kind == ast.KindElementAccessExpression {
					// For element access, check if we're accessing with a literal key
					// e.g., foo?.[key] where key is 'bar' | 'foo'
					elemAccess := expression.AsElementAccessExpression()
					argExpr := elemAccess.ArgumentExpression
					if argExpr != nil {
						// Get the type of the key
						keyType := ctx.TypeChecker.GetTypeAtLocation(argExpr)
						if keyType != nil {
							// Check if the key is a string literal type or union of string literals
							keyFlags := checker.Type_flags(keyType)
							isLiteralKey := false
							var literalKeys []string

							if keyFlags&checker.TypeFlagsStringLiteral != 0 {
								// Single string literal
								isLiteralKey = true
								if keyType.IsStringLiteral() {
									lit := keyType.AsLiteralType()
									if lit != nil {
										literalKeys = append(literalKeys, lit.Value().(string))
									}
								}
							} else if utils.IsUnionType(keyType) {
								// Union of string literals
								allLiterals := true
								for _, part := range keyType.Types() {
									partFlags := checker.Type_flags(part)
									if partFlags&checker.TypeFlagsStringLiteral != 0 {
										if part.IsStringLiteral() {
											lit := part.AsLiteralType()
											if lit != nil {
												literalKeys = append(literalKeys, lit.Value().(string))
											}
										}
									} else {
										allLiterals = false
										break
									}
								}
								isLiteralKey = allLiterals
							}

							// If we have literal keys, check if all of them have non-nullish property types
							if isLiteralKey && len(literalKeys) > 0 {
								allNonNullish := true
								for _, key := range literalKeys {
									prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, key)
									if prop == nil {
										// Property doesn't exist, might be index signature
										allNonNullish = false
										break
									}
									propType := ctx.TypeChecker.GetTypeOfSymbol(prop)
									if propType == nil || isNullishType(ctx.TypeChecker, propType) {
										allNonNullish = false
										break
									}
								}
								if allNonNullish {
									// All literal keys have non-nullish types
									// Get the actual property type for the first key (they're all non-nullish)
									prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, literalKeys[0])
									if prop != nil {
										exprType = ctx.TypeChecker.GetTypeOfSymbol(prop)
									} else {
										exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
									}
								} else {
									exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
								}
							} else {
								// Not a literal key, use default behavior
								exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
							}
						} else {
							exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
						}
					} else {
						exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
					}
				} else if expression.Kind == ast.KindCallExpression {
					// For call expressions in a chain, we need special handling
					// e.g., foo?.bar()?.baz where expression is foo?.bar()
					// or foo?.()?.baz where expression is foo?.()
					callExpr := expression.AsCallExpression()

					// For both optional and regular calls, get the function's return type
					// The full expression type includes | undefined from the optional chain/call
					funcType := getResolvedType(callExpr.Expression)
					if funcType != nil {
						// Remove nullish types to get the actual function type
						nonNullishFunc := removeNullishFromType(ctx.TypeChecker, funcType)
						if nonNullishFunc != nil {
							// Get call signatures from the function type
							signatures := ctx.TypeChecker.GetCallSignatures(nonNullishFunc)
							if len(signatures) > 0 {
								// Get the return type of the first signature
								returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
								if returnType != nil {
									exprType = returnType
								} else {
									// Can't determine return type, allow the optional chain
									return
								}
							} else {
								// No call signatures, allow the optional chain
								return
							}
						} else {
							// Function is always nullish, allow the optional chain
							return
						}
					} else {
						exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
					}
				} else {
					// For other expression types, use the full type
					exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
				}
			} else {
				// For simple access like foo?.bar, check foo's type
				// Also handle call expressions that aren't chained (e.g., foo?.bar()?.baz)
				if expression.Kind == ast.KindCallExpression {
					callExpr := expression.AsCallExpression()
					// For both optional calls (foo?.()) and regular calls (foo()),
					// get the function's return type, not the full expression type which includes undefined
					// e.g., for foo?.bar(), get the type of foo?.bar (which is () => number | undefined)
					// Then extract the return type (number)
					funcType := getResolvedType(callExpr.Expression)
					if funcType != nil {
						// Remove nullish types to get the actual function type
						nonNullishFunc := removeNullishFromType(ctx.TypeChecker, funcType)
						if nonNullishFunc != nil {
							// Get call signatures from the function type
							signatures := ctx.TypeChecker.GetCallSignatures(nonNullishFunc)
							if len(signatures) > 0 {
								// Get the return type of the first signature
								returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
								if returnType != nil {
									exprType = returnType
								} else {
									// Can't determine return type, allow the optional chain
									return
								}
							} else {
								// No call signatures, allow the optional chain
								return
							}
						} else {
							// Function is always nullish, allow the optional chain
							return
						}
					} else {
						exprType = getResolvedType(expression)
					}
				} else {
					exprType = getResolvedType(expression)
				}
			}

			if exprType == nil {
				return
			}

			// Allow optional chain on indeterminate types since we can't determine if they're nullish
			// This includes types like any, unknown, T, T[K], keyof T, etc.
			if isIndeterminateType(exprType) {
				return
			}

			// Also allow if it's a union that includes an indeterminate type
			if utils.IsUnionType(exprType) {
				for _, part := range exprType.Types() {
					if isIndeterminateType(part) {
						return
					}
				}
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
				// Note: Unlike while statements, do-while does NOT allow constant literals
				// even in "only-allowed-literals" mode
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
				// Note: "only-allowed-literals" does NOT apply to for loops
				// Only "always" mode skips checking for loops
				checkCondition(forStmt.Condition)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkCondition(node.AsConditionalExpression().Condition)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				opKind := binExpr.OperatorToken.Kind

				// Check nullish coalescing operator (??)
				if opKind == ast.KindQuestionQuestionToken {
					// Check if the left side is an array element access without noUncheckedIndexedAccess
					// In this case, the ?? is justified even if the type appears non-nullish
					// EXCEPT when the element type itself is always nullish (e.g., null[])
					leftSkip := ast.SkipParentheses(binExpr.Left)
					if leftSkip.Kind == ast.KindElementAccessExpression && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
						elemAccess := leftSkip.AsElementAccessExpression()
						baseType := getResolvedType(elemAccess.Expression)
						if baseType != nil {
							// Helper function to check if a type is an array or tuple type (reuse from checkOptionalChain)
							var isArrayOrTupleTypeLocal func(*checker.Type) bool
							isArrayOrTupleTypeLocal = func(t *checker.Type) bool {
								if t == nil {
									return false
								}
								if checker.IsTupleType(t) {
									return true
								}
								if utils.IsUnionType(t) {
									for _, part := range t.Types() {
										if isArrayOrTupleTypeLocal(part) {
											return true
										}
									}
									return false
								}
								if t.Symbol() != nil {
									symbolName := t.Symbol().Name
									if symbolName == "Array" || symbolName == "ReadonlyArray" {
										return true
									}
								}
								return false
							}
							if isArrayOrTupleTypeLocal(baseType) {
								// Check if the element type is always nullish (e.g., null[])
								// In that case, we should still report the error
								leftType := getResolvedType(binExpr.Left)
								if leftType != nil && !isAlwaysNullishType(leftType) {
									// Element type is not always nullish, so skip the check
									return
								}
								// Element type is always nullish, continue to check
							}
						}
					}

					leftType := getResolvedType(binExpr.Left)
					if leftType != nil {
						// Don't report on indeterminate types
						if isIndeterminateType(leftType) {
							return
						}

						// Check for never type first (never is a special case)
						flags := checker.Type_flags(leftType)
						if flags&checker.TypeFlagsNever != 0 {
							ctx.ReportNode(binExpr.Left, buildNeverMessage())
							return
						}

						// Check if the value is always nullish
						if isAlwaysNullishType(leftType) {
							ctx.ReportNode(binExpr.Left, buildAlwaysNullishMessage())
							return
						}

						// Check if the value is never nullish
						if !isNullishType(ctx.TypeChecker, leftType) {
							ctx.ReportNode(binExpr.Left, buildNeverNullishMessage())
						}
					}
					return
				}

				// Check nullish coalescing assignment operator (??=)
				if opKind == ast.KindQuestionQuestionEqualsToken {
					leftType := getResolvedType(binExpr.Left)
					if leftType != nil {
						// Don't report on indeterminate types
						if isIndeterminateType(leftType) {
							return
						}

						// Skip optional property access - with exactOptionalPropertyTypes,
						// the type doesn't include undefined but the property can still be absent
						if binExpr.Left.Kind == ast.KindPropertyAccessExpression || binExpr.Left.Kind == ast.KindElementAccessExpression {
							// Check if accessing an optional property or using indexed access with union keys
							var propSymbol *ast.Symbol
							switch binExpr.Left.Kind {
							case ast.KindPropertyAccessExpression:
								propAccess := binExpr.Left.AsPropertyAccessExpression()
								nameNode := propAccess.Name()
								if nameNode != nil {
									propName := ast.GetTextOfPropertyName(nameNode)
									if propName != "" {
										baseType := getResolvedType(propAccess.Expression)
										if baseType != nil {
											propSymbol = checker.Checker_getPropertyOfType(ctx.TypeChecker, baseType, propName)
										}
									}
								}
							case ast.KindElementAccessExpression:
								// For element access, we can't reliably determine if the property is optional
								// because the key might be a union type or computed at runtime
								// So we skip the check for element access
								return
							}
							// If property is optional, skip check (property may be undefined even if type doesn't show it)
							if propSymbol != nil && propSymbol.Flags&ast.SymbolFlagsOptional != 0 {
								return
							}
						}

						// Check if the value is always nullish
						flags := checker.Type_flags(leftType)
						if flags&checker.TypeFlagsNever != 0 {
							// Special case for never type
							ctx.ReportNode(binExpr.Left, buildNeverMessage())
							return
						}

						if isAlwaysNullishType(leftType) {
							ctx.ReportNode(binExpr.Left, buildAlwaysNullishMessage())
							return
						}

						// Check if the value is never nullish
						if !isNullishType(ctx.TypeChecker, leftType) {
							ctx.ReportNode(binExpr.Left, buildNeverNullishMessage())
						}
					}
					return
				}

				// Check logical operators (&&, ||) and logical assignment operators (&&=, ||=)
				if opKind == ast.KindAmpersandAmpersandToken ||
					opKind == ast.KindBarBarToken ||
					opKind == ast.KindAmpersandAmpersandEqualsToken ||
					opKind == ast.KindBarBarEqualsToken {

					isAndOperator := opKind == ast.KindAmpersandAmpersandToken || opKind == ast.KindAmpersandAmpersandEqualsToken
					isOrOperator := opKind == ast.KindBarBarToken || opKind == ast.KindBarBarEqualsToken

					// Check if left is a literal boolean (true/false keyword)
					leftSkipNode := ast.SkipParentheses(binExpr.Left)
					leftIsLiteralTrue := leftSkipNode.Kind == ast.KindTrueKeyword
					leftIsLiteralFalse := leftSkipNode.Kind == ast.KindFalseKeyword

					// Determine if we should skip the right side based on short-circuit behavior
					skipRight := false
					if isAndOperator && leftIsLiteralFalse {
						// Left is false, so right is never evaluated
						skipRight = true
					} else if isOrOperator && leftIsLiteralTrue {
						// Left is true, so right is never evaluated
						skipRight = true
					} else {
						// For non-literal cases, check the type
						leftType := getResolvedType(binExpr.Left)
						if leftType != nil {
							leftTruthy, leftFalsy := checkTypeCondition(ctx.TypeChecker, leftType)
							if isAndOperator && leftFalsy {
								skipRight = true
							} else if isOrOperator && leftTruthy {
								skipRight = true
							}
						}
					}

					// Check left side
					checkCondition(binExpr.Left)

					// Check right side only if it would be evaluated
					if !skipRight {
						// Control flow narrowing: if left and right are the same expression
						// and this is an &&, then right is always truthy (since we already checked left)
						if isAndOperator && isSameExpression(binExpr.Left, binExpr.Right) {
							// Report that the right side is always truthy
							ctx.ReportNode(binExpr.Right, buildAlwaysTruthyMessage())
						} else {
							checkCondition(binExpr.Right)
						}
					}
					return
				}

				// Check equality and comparison operators
				isLooseEqualityOp := opKind == ast.KindEqualsEqualsToken ||
					opKind == ast.KindExclamationEqualsToken
				isStrictEqualityOp := opKind == ast.KindEqualsEqualsEqualsToken ||
					opKind == ast.KindExclamationEqualsEqualsToken
				isEqualityOp := isLooseEqualityOp || isStrictEqualityOp

				isComparisonOp := opKind == ast.KindLessThanToken ||
					opKind == ast.KindGreaterThanToken ||
					opKind == ast.KindLessThanEqualsToken ||
					opKind == ast.KindGreaterThanEqualsToken

				if isEqualityOp || isComparisonOp {
					leftType := getResolvedType(binExpr.Left)
					rightType := getResolvedType(binExpr.Right)

					if leftType == nil || rightType == nil {
						return
					}

					// Skip if either side is any/unknown
					leftFlags := checker.Type_flags(leftType)
					rightFlags := checker.Type_flags(rightType)
					if leftFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 ||
						rightFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
						return
					}

					// Check for literal type comparisons
					leftIsLiteral := isLiteralValue(leftType)
					rightIsLiteral := isLiteralValue(rightType)

					if leftIsLiteral && rightIsLiteral {
						// Both sides are literal types
						ctx.ReportNode(node, buildLiteralBinaryExpressionMessage())
						return
					}

					// Check for type overlap in equality/inequality operations
					if isEqualityOp {
						// For equality operators, check if types can ever be equal
						// Only skip if BOTH sides are nullish OR one side is a union that includes nullish
						hasOverlap := typesHaveOverlap(leftType, rightType)

						if !hasOverlap {
							// Check if this is a valid nullish check (e.g., `a: string | null` with `a === null`)
							// We allow it if one side is exactly null/undefined and the other contains THE SAME nullish type
							leftIsNullish := leftFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
							rightIsNullish := rightFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0

							// If one side is nullish, check if the other side could contain a matching nullish type
							// For loose equality (==, !=), null and undefined are interchangeable
							if leftIsNullish {
								// Get the specific nullish flags from left
								leftNullishFlags := leftFlags & (checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid)
								rightParts := utils.UnionTypeParts(rightType)
								for _, part := range rightParts {
									partFlags := checker.Type_flags(part)
									// For loose equality, null matches undefined (and vice versa)
									if isLooseEqualityOp {
										// Check if this part has ANY nullish type
										if partFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
											return
										}
									} else {
										// For strict equality, check if this part has THE SAME nullish type
										if partFlags&leftNullishFlags != 0 {
											return
										}
									}
								}
							} else if rightIsNullish {
								// Get the specific nullish flags from right
								rightNullishFlags := rightFlags & (checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid)
								leftParts := utils.UnionTypeParts(leftType)
								for _, part := range leftParts {
									partFlags := checker.Type_flags(part)
									// For loose equality, null matches undefined (and vice versa)
									if isLooseEqualityOp {
										// Check if this part has ANY nullish type
										if partFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
											return
										}
									} else {
										// For strict equality, check if this part has THE SAME nullish type
										if partFlags&rightNullishFlags != 0 {
											return
										}
									}
								}
							}

							// Types don't overlap, report it
							ctx.ReportNode(node, buildNoOverlapMessage())
						}
					}
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

				callExpr := node.AsCallExpression()

				// Check array method predicates (filter, find, etc.)
				// This check is independent of CheckTypePredicates option
				if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) {
					if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
						if arg := callExpr.Arguments.Nodes[0]; arg != nil {
							checkPredicateFunction(ctx, arg, *opts.CheckTypePredicates)
						}
					}
				}

				// Check type guard or assertion function calls only if CheckTypePredicates is enabled
				if !*opts.CheckTypePredicates {
					return
				}

				// Check if this is a type guard or assertion function call
				callSignature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, 0)
				if callSignature != nil {
					typePredicate := ctx.TypeChecker.GetTypePredicateOfSignature(callSignature)
					if typePredicate != nil {
						// This is a type guard/assertion function call
						if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
							paramIndex := int(checker.TypePredicate_parameterIndex(typePredicate))

							if paramIndex >= 0 && paramIndex < len(callExpr.Arguments.Nodes) {
								arg := callExpr.Arguments.Nodes[paramIndex]
								if arg != nil {
									// Skip spread elements - their values are determined at runtime
									if arg.Kind == ast.KindSpreadElement {
										return
									}

									predicateKind := checker.TypePredicate_kind(typePredicate)
									argType := ctx.TypeChecker.GetTypeAtLocation(arg)

									if argType == nil {
										return
									}

									// Handle different predicate kinds
									switch predicateKind {
									case checker.TypePredicateKindAssertsIdentifier, checker.TypePredicateKindAssertsThis:
										// For "asserts x" (no type specified), check if argument is always truthy/falsy
										isTruthy, isFalsy := checkTypeCondition(ctx.TypeChecker, argType)
										if isTruthy {
											ctx.ReportNode(arg, buildAlwaysTruthyMessage())
										} else if isFalsy {
											ctx.ReportNode(arg, buildAlwaysFalsyMessage())
										}
									case checker.TypePredicateKindIdentifier, checker.TypePredicateKindThis:
										// For "x is Type" type guards, check if argument already satisfies the type
										predicateType := checker.TypePredicate_t(typePredicate)
										if predicateType != nil {
											// Check if argType is assignable to predicateType
											if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, argType, predicateType) {
												ctx.ReportNode(node, buildTypeGuardAlreadyIsTypeMessage())
											}
										}
									}
								}
							}
						}
					}
				}
			},
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				checkCondition(switchStmt.Expression)
			},
			ast.KindCaseClause: func(node *ast.Node) {
				if node.Expression() != nil {
					// Check if the case expression is a literal being compared
					// node.Parent is the CaseBlock, node.Parent.Parent is the SwitchStatement
					switchNode := node.Parent
					if switchNode != nil {
						switchNode = switchNode.Parent
					}
					if switchNode != nil && switchNode.Kind == ast.KindSwitchStatement {
						discriminant := switchNode.Expression()
						discriminantType := getResolvedType(discriminant)
						caseType := getResolvedType(node.Expression())

						if discriminantType != nil && caseType != nil {
							discriminantIsLiteral := isLiteralValue(discriminantType)
							caseIsLiteral := isLiteralValue(caseType)

							if discriminantIsLiteral && caseIsLiteral {
								ctx.ReportNode(node.Expression(), buildLiteralBinaryExpressionMessage())
							}
						}
					}
				}
			},
		}
	},
}

// checkTypeCondition determines if a type is always truthy or always falsy at runtime.
//
// Return values:
// - (true, false): type is always truthy (e.g., objects, "hello", 1, true)
// - (false, true): type is always falsy (e.g., null, undefined, false, 0, "", never)
// - (false, false): type could be either (e.g., string, number, boolean)
//
// Examples:
//   - { foo: string }: always truthy (objects are always truthy)
//   - "hello": always truthy (non-empty string literal)
//   - "": always falsy (empty string literal)
//   - 0: always falsy (zero is falsy)
//   - 1: always truthy (non-zero number)
//   - true: always truthy
//   - false: always falsy
//   - null: always falsy
//   - undefined: always falsy
//   - never: always falsy (type with no possible values)
//   - string: could be either (might be "" or "hello")
//   - number: could be either (might be 0 or 1)
//   - boolean: could be either (might be true or false)
//
// Type handling:
// - Union types: all parts must be truthy for (true, false), all must be falsy for (false, true)
// - Intersection types: if any part is always falsy, result is falsy; all must be truthy for truthy
// - Literal types: evaluates the actual literal value's truthiness
// - Object types: always truthy (even empty objects are truthy in JavaScript)
// - Symbols: always truthy (symbols are always truthy)
func checkTypeCondition(typeChecker *checker.Checker, t *checker.Type) (isTruthy bool, isFalsy bool) {
	flags := checker.Type_flags(t)

	// Never type is always falsy (empty type, no values exist)
	if flags&checker.TypeFlagsNever != 0 {
		return false, true
	}

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

	// Handle intersections - check all parts
	// For intersections, all parts must be truthy for the whole to be truthy
	if utils.IsIntersectionType(t) {
		allTruthy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(typeChecker, part)
			// If any part is always falsy, intersection is likely never/empty
			if partFalsy {
				return false, true
			}
			// If any part is not always truthy, we can't say the whole is always truthy
			if !partTruthy {
				allTruthy = false
			}
		}

		return allTruthy, false
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

// isNullishType checks if a type can be null, undefined, or void.
//
// For union types, returns true if any part of the union is nullish.
// This is used to determine if the nullish coalescing operator (??) or
// optional chaining (?.) might be necessary.
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

// removeNullishFromType removes null, undefined, and void from a union type.
// Returns the non-nullish part of the type, or nil if the type is entirely nullish.
func removeNullishFromType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	if !utils.IsUnionType(t) {
		// Not a union - check if it's nullish
		flags := checker.Type_flags(t)
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return nil
		}
		return t
	}

	// For union types, filter out nullish parts
	var nonNullishParts []*checker.Type
	for _, part := range t.Types() {
		if !isNullishType(typeChecker, part) {
			nonNullishParts = append(nonNullishParts, part)
		}
	}

	if len(nonNullishParts) == 0 {
		return nil
	}
	if len(nonNullishParts) == 1 {
		return nonNullishParts[0]
	}

	// Multiple non-nullish parts - return first one for now
	// (TypeScript would create a new union type, but we don't have that API)
	return nonNullishParts[0]
}

// isAllowedConstantLiteral checks if an expression is a literal that's allowed in loop conditions.
//
// When AllowConstantLoopConditions is set to "only-allowed-literals", only the
// following literals are allowed in loop conditions: true, false, 0, 1
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

// typesHaveOverlap determines if two types can have any overlapping values.
//
// This function is used to detect comparisons that will always return the same result,
// such as comparing a string with a number (always false) or comparing two different
// string literals (always false).
//
// Examples:
//   - string and number: no overlap (string === number is always false)
//   - string | null and null: overlap (null can match)
//   - string and "hello": overlap ("hello" is a string)
//   - "hello" and "world": no overlap (different literals)
//   - 1 and "1": no overlap (different primitive types)
//
// Special handling:
// - any/unknown: always overlap with everything (could be any value)
// - Type parameters (T, K): conservatively treated as overlapping (we don't know T at compile time)
// - Indexed access (T[K]): conservatively treated as overlapping
// - Nullish types: null/undefined/void overlap with each other (treated as interchangeable in some contexts)
// - Literals and base types: overlap (e.g., "hello" overlaps with string)
// - Union types: checked part by part (e.g., string | number overlaps with "hello")
func typesHaveOverlap(left, right *checker.Type) bool {
	// Handle any/unknown types - they overlap with everything
	leftFlags := checker.Type_flags(left)
	rightFlags := checker.Type_flags(right)

	if leftFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 ||
		rightFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}

	// Handle type parameters and indexed access types - we can't determine overlap at compile time
	// This includes T, T[K], keyof T, etc.
	genericFlags := checker.TypeFlagsTypeParameter | checker.TypeFlagsIndexedAccess | checker.TypeFlagsIndex
	if leftFlags&genericFlags != 0 || rightFlags&genericFlags != 0 {
		return true
	}

	// Get union parts
	leftParts := utils.UnionTypeParts(left)
	rightParts := utils.UnionTypeParts(right)

	// Check for overlap between any parts
	for _, leftPart := range leftParts {
		leftPartFlags := checker.Type_flags(leftPart)

		for _, rightPart := range rightParts {
			rightPartFlags := checker.Type_flags(rightPart)

			// Check if both are the same primitive type
			primitiveFlags := checker.TypeFlagsString | checker.TypeFlagsNumber |
				checker.TypeFlagsBoolean | checker.TypeFlagsBigInt |
				checker.TypeFlagsESSymbol | checker.TypeFlagsObject

			if leftPartFlags&primitiveFlags != 0 && rightPartFlags&primitiveFlags != 0 {
				if leftPartFlags&rightPartFlags&primitiveFlags != 0 {
					return true
				}
			}

			// Null and undefined/void overlap
			// void is treated as undefined at runtime
			nullishFlags := checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid
			if leftPartFlags&nullishFlags != 0 && rightPartFlags&nullishFlags != 0 {
				// Check if both have the same nullish type
				if leftPartFlags&rightPartFlags&nullishFlags != 0 {
					return true
				}
				// void overlaps with undefined
				if (leftPartFlags&checker.TypeFlagsVoid != 0 && rightPartFlags&checker.TypeFlagsUndefined != 0) ||
					(leftPartFlags&checker.TypeFlagsUndefined != 0 && rightPartFlags&checker.TypeFlagsVoid != 0) {
					return true
				}
			}

			// If one is nullish and the other is not, no overlap
			if (leftPartFlags&nullishFlags != 0) != (rightPartFlags&nullishFlags != 0) {
				continue
			}

			// Objects overlap with other objects
			if leftPartFlags&checker.TypeFlagsObject != 0 && rightPartFlags&checker.TypeFlagsObject != 0 {
				return true
			}

			// Literals overlap with their base types and other literals of the same type
			// String literal vs string (base type or literal)
			if (leftPartFlags&checker.TypeFlagsStringLiteral != 0 && rightPartFlags&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral) != 0) ||
				(leftPartFlags&checker.TypeFlagsString != 0 && rightPartFlags&checker.TypeFlagsStringLiteral != 0) {
				return true
			}
			// Number literal vs number (base type or literal)
			if (leftPartFlags&checker.TypeFlagsNumberLiteral != 0 && rightPartFlags&(checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral) != 0) ||
				(leftPartFlags&checker.TypeFlagsNumber != 0 && rightPartFlags&checker.TypeFlagsNumberLiteral != 0) {
				return true
			}
			// BigInt literal vs bigint (base type or literal)
			if (leftPartFlags&checker.TypeFlagsBigIntLiteral != 0 && rightPartFlags&(checker.TypeFlagsBigInt|checker.TypeFlagsBigIntLiteral) != 0) ||
				(leftPartFlags&checker.TypeFlagsBigInt != 0 && rightPartFlags&checker.TypeFlagsBigIntLiteral != 0) {
				return true
			}
			// Boolean literal vs boolean (base type or literal)
			if (leftPartFlags&checker.TypeFlagsBooleanLiteral != 0 && rightPartFlags&(checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0) ||
				(leftPartFlags&checker.TypeFlagsBoolean != 0 && rightPartFlags&checker.TypeFlagsBooleanLiteral != 0) {
				return true
			}
		}
	}

	return false
}

// checkPredicateFunction analyzes predicate functions used in array methods like filter/find.
//
// This function performs two checks:
//  1. If checkTypeGuards is true and the function is a type guard, it checks if the
//     parameter already satisfies the type predicate (making the guard unnecessary)
//  2. It checks if the function's return type is always truthy or always falsy,
//     which would make it a useless filter/find predicate
//
// Used for array methods like:
// - [1, 2, 3].filter(() => true) // always truthy, returns all elements
// - [1, 2, 3].find(() => false)  // always falsy, returns undefined
func checkPredicateFunction(ctx rule.RuleContext, funcNode *ast.Node, checkTypeGuards bool) {
	isFunction := funcNode.Kind&(ast.KindArrowFunction|ast.KindFunctionExpression|ast.KindFunctionDeclaration) != 0
	if !isFunction {
		return
	}

	funcType := ctx.TypeChecker.GetTypeAtLocation(funcNode)
	signatures := ctx.TypeChecker.GetCallSignatures(funcType)

	for _, signature := range signatures {
		// Check if this is a type predicate (type guard)
		typePredicate := ctx.TypeChecker.GetTypePredicateOfSignature(signature)
		if checkTypeGuards && typePredicate != nil {
			// Check if the argument already satisfies the type predicate
			params := checker.Signature_parameters(signature)
			if len(params) > 0 {
				// Get the parameter index being guarded
				paramIndex := int(checker.TypePredicate_parameterIndex(typePredicate))

				if paramIndex >= 0 && paramIndex < len(params) {
					param := params[paramIndex]
					if param != nil {
						paramType := ctx.TypeChecker.GetTypeOfSymbol(param)
						predicateKind := checker.TypePredicate_kind(typePredicate)

						if paramType != nil {
							// Only check "x is Type" predicates, not "asserts x" predicates
							// "asserts x" predicates in functions are checked via their return type below
							if predicateKind == checker.TypePredicateKindIdentifier ||
								predicateKind == checker.TypePredicateKindThis {
								predicateType := checker.TypePredicate_t(typePredicate)
								if predicateType != nil {
									// Check if paramType is assignable to predicateType
									// If so, the type guard is unnecessary
									if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, paramType, predicateType) {
										ctx.ReportNode(funcNode, buildTypeGuardAlreadyIsTypeMessage())
										return
									}
								}
							}
						}
					}
				}
			}
		}

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

		if isTruthy || isFalsy {
			// Use different message based on whether it's a literal function or function reference
			// Literal functions: () => true, () => false, function() { return true }
			// Function references: truthy, falsy (identifier)
			isLiteralFunction := funcNode.Kind == ast.KindArrowFunction || funcNode.Kind == ast.KindFunctionExpression

			if isTruthy {
				if isLiteralFunction {
					ctx.ReportNode(funcNode, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(funcNode, buildAlwaysTruthyFuncMessage())
				}
			} else if isFalsy {
				if isLiteralFunction {
					ctx.ReportNode(funcNode, buildAlwaysFalsyMessage())
				} else {
					ctx.ReportNode(funcNode, buildAlwaysFalsyFuncMessage())
				}
			}
		}
	}
}

// isLiteralValue checks if a type is a literal singleton type.
//
// Literal types include:
// - Nullish types: null, undefined, void
// - String literals: "hello", ""
// - Number literals: 1, 0, -5, NaN
// - BigInt literals: 1n, 0n
// - Boolean literals: true, false
//
// These types represent a single, specific value rather than a range of possible values.
func isLiteralValue(t *checker.Type) bool {
	flags := checker.Type_flags(t)

	// Nullish types are also literal singleton types
	if flags&checker.TypeFlagsNull != 0 {
		return true
	}
	if flags&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if flags&checker.TypeFlagsVoid != 0 {
		return true
	}

	if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if _, ok := literal.Value().(string); ok {
				return true
			}
		}
	}

	if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			return true
		}
	}

	if flags&checker.TypeFlagsBigIntLiteral != 0 && t.IsBigIntLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			return true
		}
	}

	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		if utils.IsIntrinsicType(t) {
			return true
		}
		if t.AsLiteralType() != nil {
			return true
		}
	}

	return false
}
