package no_misused_spread

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildAddAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addAwait",
		Description: "Add await operator.",
	}
}
func buildNoArraySpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noArraySpreadInObject",
		Description: "Using the spread operator on an array in an object will result in a list of indices.",
	}
}
func buildNoClassDeclarationSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noClassDeclarationSpreadInObject",
		Description: "Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.",
	}
}
func buildNoClassInstanceSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noClassInstanceSpreadInObject",
		Description: "Using the spread operator on class instances will lose their class prototype.",
	}
}
func buildNoFunctionSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noFunctionSpreadInObject",
		Description: "Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?",
	}
}
func buildNoIterableSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noIterableSpreadInObject",
		Description: "Using the spread operator on an Iterable in an object can cause unexpected behavior.",
	}
}
func buildNoMapSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noMapSpreadInObject",
		Description: "Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?",
	}
}
func buildNoPromiseSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noPromiseSpreadInObject",
		Description: "Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?",
	}
}
func buildNoStringSpreadMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id: "noStringSpread",
		Description: "Using the spread operator on a string can mishandle special characters, as can `.split(\"\")`.\n'" +
			"- `...` produces Unicode code points, which will decompose complex emojis into individual emojis" +
			"- .split(\"\") produces UTF-16 code units, which breaks rich characters in many languages" +
			"Consider using `Intl.Segmenter` for locale-aware string decomposition." +
			"Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.",
	}
}
func buildReplaceMapSpreadInObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceMapSpreadInObject",
		Description: "Replace map spread in object with `Object.fromEntries()`",
	}
}

func isString(t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
	})
}

func isPromise(program *compiler.Program, typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return utils.IsPromiseLike(program, typeChecker, t)
	})
}

func isFunctionWithoutProps(typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return len(utils.GetCallSignatures(typeChecker, t)) != 0 && len(checker.Checker_getPropertiesOfType(typeChecker, t)) == 0
	})
}

func isMap(program *compiler.Program, typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return utils.IsBuiltinSymbolLike(program, typeChecker, t, "Map", "ReadonlyMap", "WeakMap")
	})
}

func isArray(typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return checker.Checker_isArrayOrTupleType(typeChecker, t)
	})
}

func isIterable(typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		e := utils.GetWellKnownSymbolPropertyOfType(t, "iterator", typeChecker)

		return e != nil
	})
}

func isClassInstance(typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		// If the type itself has a construct signature, it's a class(-like)
		if len(utils.GetConstructSignatures(typeChecker, t)) != 0 {
			return false
		}

		symbol := checker.Type_symbol(t)
		if symbol == nil || symbol.Declarations == nil {
			return false
		}

		// If the type's symbol has a construct signature, the type is an instance
		for _, decl := range symbol.Declarations {
			if len(utils.GetConstructSignatures(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(symbol, decl))) != 0 {
				return true
			}
		}
		return false
	})
}

func isClassDeclaration(t *checker.Type) bool {
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		if utils.IsObjectType(t) && checker.Type_objectFlags(t)&checker.ObjectFlagsInstantiationExpressionType != 0 {
			return true
		}

		symbol := checker.Type_symbol(t)
		return symbol != nil && symbol.ValueDeclaration != nil && (ast.IsClassExpression(symbol.ValueDeclaration) || ast.IsClassDeclaration(symbol.ValueDeclaration))
	})
}

var NoMisusedSpreadRule = rule.Rule{
	Name: "no-misused-spread",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoMisusedSpreadOptions](options, "no-misused-spread")

		checkArrayOrCallSpread := func(node *ast.Node) {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.AsSpreadElement().Expression)
			if !utils.TypeMatchesSomeSpecifier(t, opts.Allow, ctx.Program) && isString(t) {
				ctx.ReportNode(node, buildNoStringSpreadMessage())
			}
		}

		insertAwaitFix := func(node *ast.Node) []rule.RuleFix {
			if utils.IsHigherPrecedenceThanAwait(node) {
				return []rule.RuleFix{
					rule.RuleFixInsertBefore(ctx.SourceFile, node, "await "),
				}
			}
			return []rule.RuleFix{
				rule.RuleFixInsertBefore(ctx.SourceFile, node, "await ("),
				rule.RuleFixInsertAfter(node, ")"),
			}
		}

		getMapSpreadSuggestions := func(node *ast.Node, argument *ast.Node, t *checker.Type) []rule.RuleSuggestion {
			// TODO(port): do we need this loop?
			for _, t := range utils.UnionTypeParts(t) {
				if !isMap(ctx.Program, ctx.TypeChecker, t) {
					return []rule.RuleSuggestion{}
				}
			}

			if ast.IsObjectLiteralExpression(node.Parent) {
				properties := node.Parent.AsObjectLiteralExpression().Properties
				if len(properties.Nodes) == 1 {
					return []rule.RuleSuggestion{
						{
							Message: buildReplaceMapSpreadInObjectMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Parent.Pos())),                  // {
								rule.RuleFixReplaceRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos()), "Object.fromEntries("), // ...
								rule.RuleFixReplaceRange(properties.Loc.WithPos(argument.End()), ")"),
								rule.RuleFixRemoveRange(node.Parent.Loc.WithPos(node.Parent.End() - 1)), // }
							},
						},
					}
				}
			}

			return []rule.RuleSuggestion{
				{
					Message: buildReplaceMapSpreadInObjectMessage(),
					FixesArr: []rule.RuleFix{
						rule.RuleFixInsertBefore(ctx.SourceFile, argument, "Object.fromEntries("),
						rule.RuleFixInsertAfter(argument, ")"),
					},
				},
			}
		}

		checkObjectSpread := func(node *ast.Node, argument *ast.Node) {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)

			if utils.TypeMatchesSomeSpecifier(t, opts.Allow, ctx.Program) {
				return
			}

			if isPromise(ctx.Program, ctx.TypeChecker, t) {
				ctx.ReportNodeWithSuggestions(node, buildNoPromiseSpreadInObjectMessage(), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{{
						Message:  buildAddAwaitMessage(),
						FixesArr: insertAwaitFix(ast.SkipParentheses(argument)),
					}}
				})

				return
			}

			if isFunctionWithoutProps(ctx.TypeChecker, t) {
				ctx.ReportNode(node, buildNoFunctionSpreadInObjectMessage())

				return
			}

			if isMap(ctx.Program, ctx.TypeChecker, t) {
				ctx.ReportNodeWithSuggestions(node, buildNoMapSpreadInObjectMessage(), func() []rule.RuleSuggestion { return getMapSpreadSuggestions(node, argument, t) })

				return
			}

			if isArray(ctx.TypeChecker, t) {
				ctx.ReportNode(node, buildNoArraySpreadInObjectMessage())

				return
			}

			// Don't report when the type is string, since TS will flag it already
			if isIterable(ctx.TypeChecker, t) && !isString(t) {
				ctx.ReportNode(node, buildNoIterableSpreadInObjectMessage())

				return
			}

			if isClassInstance(ctx.TypeChecker, t) {
				ctx.ReportNode(node, buildNoClassInstanceSpreadInObjectMessage())

				return
			}

			if isClassDeclaration(t) {
				ctx.ReportNode(node, buildNoClassDeclarationSpreadInObjectMessage())

				return
			}
		}

		return rule.RuleListeners{
			rule.ListenerOnNotAllowPattern(ast.KindArrayLiteralExpression): func(node *ast.Node) {
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					if ast.IsSpreadElement(element) {
						checkArrayOrCallSpread(element)
					}
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				for _, element := range node.AsCallExpression().Arguments.Nodes {
					if ast.IsSpreadElement(element) {
						checkArrayOrCallSpread(element)
					}
				}
			},
			ast.KindJsxSpreadAttribute: func(node *ast.Node) {
				checkObjectSpread(node, node.AsJsxSpreadAttribute().Expression)
			},
			rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				for _, element := range node.AsObjectLiteralExpression().Properties.Nodes {
					if ast.IsSpreadAssignment(element) {
						checkObjectSpread(element, element.AsSpreadAssignment().Expression)
					}
				}
			},
		}
	},
}
