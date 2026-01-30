package promise_function_async

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMissingAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAsync",
		Description: "Functions that return promises must be async.",
	}
}

func buildMissingAsyncHybridReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAsyncHybridReturn",
		Description: "Functions that return promises must be async. Consider adding an explicit return type annotation if the function is intended to return a union of promise and non-promise types.",
	}
}

var PromiseFunctionAsyncRule = rule.Rule{
	Name: "promise-function-async",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PromiseFunctionAsyncOptions](options, "promise-function-async")

		allAllowedPromiseNames := utils.NewSetWithSizeHint[string](len(opts.AllowedPromiseNames))
		allAllowedPromiseNames.Add("Promise")
		for _, name := range opts.AllowedPromiseNames {
			allAllowedPromiseNames.Add(name)
		}

		var containsAllTypesByName func(t *checker.Type, matchAnyInstead bool) bool
		containsAllTypesByName = func(t *checker.Type, matchAnyInstead bool) bool {
			if utils.IsTypeFlagSet(t, checker.TypeFlagsAnyOrUnknown) {
				return false
			}

			if utils.IsTypeFlagSet(t, checker.TypeFlagsObject) && checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
				t = t.Target()
			}

			symbol := checker.Type_symbol(t)
			if symbol != nil && allAllowedPromiseNames.Has(symbol.Name) {
				return true
			}

			predicate := func(t *checker.Type) bool {
				return containsAllTypesByName(t, matchAnyInstead)
			}

			if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
				if matchAnyInstead {
					return utils.Every(t.Types(), predicate)
				}
				return utils.Some(t.Types(), predicate)
			}

			if checker.Type_objectFlags(t)&checker.ObjectFlagsClassOrInterface == 0 {
				return false
			}

			bases := checker.Checker_getBaseTypes(ctx.TypeChecker, t)
			if matchAnyInstead {
				return utils.Some(bases, predicate)
			}
			return len(bases) > 0 && utils.Every(bases, predicate)
		}

		listeners := make(rule.RuleListeners, 3)

		validateNode := func(node *ast.Node) {
			if utils.IncludesModifier(node, ast.KindAsyncKeyword) || node.Body() == nil {
				return
			}

			t := ctx.TypeChecker.GetTypeAtLocation(node)
			signatures := utils.GetCallSignatures(ctx.TypeChecker, t)
			if len(signatures) == 0 {
				return
			}

			hasExplicitReturnType := node.Type() != nil

			everySignatureReturnsPromise := true
			for _, signature := range signatures {
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
				if !opts.AllowAny && utils.IsTypeFlagSet(returnType, checker.TypeFlagsAnyOrUnknown) {
					// Report without auto fixer because the return type is unknown
					// TODO(port): getFunctionHeadLoc
					ctx.ReportNode(node, buildMissingAsyncMessage())
					return
				}

				// require all potential return types to be promise/any/unknown
				everySignatureReturnsPromise = everySignatureReturnsPromise && containsAllTypesByName(
					returnType,
					// If no return type is explicitly set, we check if any parts of the return type match a Promise (instead of requiring all to match).
					hasExplicitReturnType,
				)
			}

			if !everySignatureReturnsPromise {
				return
			}

			// Check if any signature has a hybrid return type (union with both promise and non-promise parts)
			// This only applies when there's no explicit return type annotation
			isHybridReturnType := false
			if !hasExplicitReturnType {
				for _, signature := range signatures {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
					if utils.IsUnionType(returnType) {
						// Check if not every part of the union is a promise type
						allPartsArePromise := utils.Every(returnType.Types(), func(part *checker.Type) bool {
							return containsAllTypesByName(part, true)
						})
						if !allPartsArePromise {
							isHybridReturnType = true
							break
						}
					}
				}
			}

			insertAsyncBeforeNode := node
			if ast.IsMethodDeclaration(node) {
				insertAsyncBeforeNode = node.Name()
			}

			var message rule.RuleMessage
			if isHybridReturnType {
				message = buildMissingAsyncHybridReturnMessage()
			} else {
				message = buildMissingAsyncMessage()
			}

			// TODO(port): getFunctionHeadLoc
			ctx.ReportNodeWithFixes(node, message, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, insertAsyncBeforeNode, " async ")}
			})
		}

		if opts.CheckArrowFunctions {
			listeners[ast.KindArrowFunction] = validateNode
		}

		if opts.CheckFunctionDeclarations {
			listeners[ast.KindFunctionDeclaration] = validateNode
		}

		if opts.CheckFunctionExpressions {
			listeners[ast.KindFunctionExpression] = validateNode
		}

		if opts.CheckMethodDeclarations {
			listeners[ast.KindMethodDeclaration] = func(node *ast.Node) {
				if utils.IncludesModifier(node, ast.KindAbstractKeyword) {
					// Abstract method can't be async
					return
				}
				validateNode(node)
			}
		}

		return listeners
	},
}
