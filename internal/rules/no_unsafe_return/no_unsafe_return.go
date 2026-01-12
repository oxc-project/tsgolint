package no_unsafe_return

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUnsafeReturnMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeReturn",
		Description: fmt.Sprintf("Unsafe return of a value of type %v.", t),
	}
}
func buildUnsafeReturnAssignmentMessage(sender, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeReturnAssignment",
		Description: fmt.Sprintf("Unsafe return of type `%v` from function with return type `%v`.", sender, receiver),
	}
}
func buildUnsafeReturnThisMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unsafeReturnThis",
		Description: fmt.Sprintf("Unsafe return of a value of type `%v`. `this` is typed as `any`.", t) +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

// containsExplicitAny checks if a node or its descendants contain an explicit `any` type annotation.
// This is used to distinguish between:
// - Explicit `any` usage (e.g., `x as any`, `[] as any[]`, `new Set<any>()`) - intentional
// - Inferred `any` due to type resolution issues - potentially false positive
func containsExplicitAny(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if this node is an `any` keyword
	if node.Kind == ast.KindAnyKeyword {
		return true
	}

	// Check type assertions: `x as any` or `x as any[]`
	if node.Kind == ast.KindAsExpression {
		asExpr := node.AsAsExpression()
		if containsExplicitAny(asExpr.Type) {
			return true
		}
		// Also check the expression in case of nested assertions
		return containsExplicitAny(asExpr.Expression)
	}

	// Check type assertions: `<any>x`
	if node.Kind == ast.KindTypeAssertionExpression {
		typeAssertion := node.AsTypeAssertion()
		if containsExplicitAny(typeAssertion.Type) {
			return true
		}
		return containsExplicitAny(typeAssertion.Expression)
	}

	// Check call expressions with type arguments: `foo<any>()`
	if ast.IsCallExpression(node) {
		callExpr := node.AsCallExpression()
		if callExpr.TypeArguments != nil {
			for _, typeArg := range callExpr.TypeArguments.Nodes {
				if containsExplicitAny(typeArg) {
					return true
				}
			}
		}
		// Check the callee and arguments
		if containsExplicitAny(callExpr.Expression) {
			return true
		}
		for _, arg := range callExpr.Arguments.Nodes {
			if containsExplicitAny(arg) {
				return true
			}
		}
		return false
	}

	// Check new expressions with type arguments: `new Set<any>()`
	if ast.IsNewExpression(node) {
		newExpr := node.AsNewExpression()
		if newExpr.TypeArguments != nil {
			for _, typeArg := range newExpr.TypeArguments.Nodes {
				if containsExplicitAny(typeArg) {
					return true
				}
			}
		}
		if containsExplicitAny(newExpr.Expression) {
			return true
		}
		if newExpr.Arguments != nil {
			for _, arg := range newExpr.Arguments.Nodes {
				if containsExplicitAny(arg) {
					return true
				}
			}
		}
		return false
	}

	// Check array type: `any[]`
	if node.Kind == ast.KindArrayType {
		arrayType := node.AsArrayTypeNode()
		return containsExplicitAny(arrayType.ElementType)
	}

	// Check type reference with type arguments: `Array<any>`, `Set<any>`, etc.
	if node.Kind == ast.KindTypeReference {
		typeRef := node.AsTypeReferenceNode()
		if typeRef.TypeArguments != nil {
			for _, typeArg := range typeRef.TypeArguments.Nodes {
				if containsExplicitAny(typeArg) {
					return true
				}
			}
		}
		return false
	}

	// Check property access: `obj.prop`
	if ast.IsPropertyAccessExpression(node) {
		propAccess := node.AsPropertyAccessExpression()
		return containsExplicitAny(propAccess.Expression)
	}

	// Check element access: `obj[key]`
	if ast.IsElementAccessExpression(node) {
		elemAccess := node.AsElementAccessExpression()
		return containsExplicitAny(elemAccess.Expression) || containsExplicitAny(elemAccess.ArgumentExpression)
	}

	// Check array literals
	if ast.IsArrayLiteralExpression(node) {
		arrayLit := node.AsArrayLiteralExpression()
		for _, elem := range arrayLit.Elements.Nodes {
			if containsExplicitAny(elem) {
				return true
			}
		}
		return false
	}

	// Check object literals
	if ast.IsObjectLiteralExpression(node) {
		objLit := node.AsObjectLiteralExpression()
		for _, prop := range objLit.Properties.Nodes {
			if containsExplicitAny(prop) {
				return true
			}
		}
		return false
	}

	return false
}

var NoUnsafeReturnRule = rule.Rule{
	Name: "no-unsafe-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		isNoImplicitThis := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.NoImplicitThis,
		)

		checkReturn := func(
			returnNode *ast.Node,
			reportingNode *ast.Node,
		) {
			t := ctx.TypeChecker.GetTypeAtLocation(returnNode)

			anyType := utils.DiscriminateAnyType(
				t,
				ctx.TypeChecker,
				ctx.Program,
				returnNode,
			)
			functionNode := utils.GetParentFunctionNode(returnNode)
			if functionNode == nil {
				return
			}

			// function has an explicit return type, so ensure it's a safe return
			returnNodeType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, returnNode)

			// function expressions will not have their return type modified based on receiver typing
			// so we have to use the contextual typing in these cases, i.e.
			// const foo1: () => Set<string> = () => new Set<any>();
			// the return type of the arrow function is Set<any> even though the variable is typed as Set<string>
			var functionType *checker.Type
			if ast.IsFunctionExpression(functionNode) || ast.IsArrowFunction(functionNode) {
				functionType = utils.GetContextualType(ctx.TypeChecker, functionNode)
			}
			if functionType == nil {
				functionType = ctx.TypeChecker.GetTypeAtLocation(functionNode)
			}
			callSignatures := utils.CollectAllCallSignatures(ctx.TypeChecker, functionType)
			// If there is an explicit type annotation *and* that type matches the actual
			// function return type, we shouldn't complain (it's intentional, even if unsafe)
			if functionNode.Type() != nil {
				for _, signature := range callSignatures {
					signatureReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

					if returnNodeType == signatureReturnType ||
						utils.IsTypeFlagSet(
							signatureReturnType,
							checker.TypeFlagsAny|checker.TypeFlagsUnknown,
						) {
						return
					}
					if ast.HasSyntacticModifier(functionNode, ast.ModifierFlagsAsync) {
						awaitedSignatureReturnType := checker.Checker_getAwaitedType(ctx.TypeChecker, signatureReturnType)
						awaitedReturnNodeType := checker.Checker_getAwaitedType(ctx.TypeChecker, returnNodeType)

						if awaitedSignatureReturnType == awaitedReturnNodeType || (awaitedSignatureReturnType != nil && utils.IsTypeFlagSet(awaitedSignatureReturnType, checker.TypeFlagsAny|checker.TypeFlagsUnknown)) {
							return
						}
					}
				}
			}

			if anyType != utils.DiscriminatedAnyTypeSafe {
				// Allow cases when the declared return type of the function is either unknown or unknown[]
				// and the function is returning any or any[].
				for _, signature := range callSignatures {
					functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
					if anyType == utils.DiscriminatedAnyTypeAny && utils.IsTypeUnknownType(functionReturnType) {
						return
					}

					if anyType == utils.DiscriminatedAnyTypeAnyArray && utils.IsTypeUnknownArrayType(functionReturnType, ctx.TypeChecker) {
						return
					}
					awaitedType := checker.Checker_getAwaitedType(ctx.TypeChecker, functionReturnType)
					if awaitedType != nil &&
						anyType == utils.DiscriminatedAnyTypePromiseAny &&
						utils.IsTypeUnknownType(awaitedType) {
						return
					}
				}

				if anyType == utils.DiscriminatedAnyTypePromiseAny && !ast.HasSyntacticModifier(functionNode, ast.ModifierFlagsAsync) {
					return
				}

				// For arrow functions/function expressions without explicit type annotations:
				// If the type is inferred as `any` but the source code doesn't contain explicit `any`,
				// and there's contextual typing with a concrete return type, trust TypeScript's
				// type inference. This prevents false positives when typescript-go's type resolution
				// differs from TypeScript's for complex types (e.g., conditional types, mapped types).
				if anyType == utils.DiscriminatedAnyTypeAny &&
					functionNode.Type() == nil &&
					(ast.IsFunctionExpression(functionNode) || ast.IsArrowFunction(functionNode)) &&
					!containsExplicitAny(returnNode) {
					for _, signature := range callSignatures {
						signatureReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
						// If the contextual return type is a concrete type (not any/unknown),
						// don't flag the return as unsafe
						if signatureReturnType != nil &&
							!utils.IsTypeFlagSet(signatureReturnType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
							return
						}
					}
				}

				var typeString string
				if utils.IsIntrinsicErrorType(returnNodeType) {
					typeString = "error"
				} else if anyType == utils.DiscriminatedAnyTypeAny {
					typeString = "`any`"
				} else if anyType == utils.DiscriminatedAnyTypePromiseAny {
					typeString = "`Promise<any>`"
				} else {
					typeString = "`any[]`"
				}

				if !isNoImplicitThis {
					// `return this`
					thisExpression := utils.GetThisExpression(returnNode)
					if thisExpression != nil && utils.IsTypeAnyType(utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)) {
						ctx.ReportNode(reportingNode, buildUnsafeReturnThisMessage(typeString))
						return
					}
				}
				// If the function return type was not unknown/unknown[], mark usage as unsafeReturn.
				ctx.ReportNode(reportingNode, buildUnsafeReturnMessage(typeString))
				return
			}

			if len(callSignatures) < 1 {
				return
			}

			signature := callSignatures[0]
			functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

			receiver, sender, unsafe := utils.IsUnsafeAssignment(
				returnNodeType,
				functionReturnType,
				ctx.TypeChecker,
				returnNode,
			)

			if !unsafe {
				return
			}

			ctx.ReportNode(reportingNode, buildUnsafeReturnAssignmentMessage(ctx.TypeChecker.TypeToString(sender), ctx.TypeChecker.TypeToString(receiver)))
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if !ast.IsBlock(body) {
					checkReturn(body, body)
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				argument := node.Expression()
				if argument == nil {
					return
				}

				checkReturn(argument, node)
			},
		}
	},
}
