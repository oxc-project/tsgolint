package no_unsafe_return

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
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
		Id:          "unsafeReturnThis",
		Description: fmt.Sprintf("Unsafe return of a value of type `%v`. `this` is typed as `any`.", t),
		Help:        "You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

func buildUnsafeReturnDiagnostic(
	message rule.RuleMessage,
	primaryRange core.TextRange,
	returnedRange core.TextRange,
	returnedType string,
	expectedRange *core.TextRange,
	expectedType string,
) rule.RuleDiagnostic {
	diagnostic := rule.RuleDiagnostic{
		Range:   primaryRange,
		Message: message,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("Returned expression has type `%s`.", returnedType),
				Range: returnedRange,
			},
		},
	}
	if expectedRange != nil {
		diagnostic.LabeledRanges = append(diagnostic.LabeledRanges, rule.RuleLabeledRange{
			Label: fmt.Sprintf("Function expects return type `%s`.", expectedType),
			Range: *expectedRange,
		})
	}
	return diagnostic
}

func renderReturnType(typeChecker *checker.Checker, t *checker.Type) string {
	if utils.IsIntrinsicErrorType(t) {
		return "error"
	}
	return typeChecker.TypeToString(t)
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
			primaryRange core.TextRange,
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
			usesContextualType := false
			if ast.IsFunctionExpression(functionNode) || ast.IsArrowFunction(functionNode) {
				functionType = utils.GetContextualType(ctx.TypeChecker, functionNode)
				usesContextualType = functionType != nil
			}
			if functionType == nil {
				functionType = ctx.TypeChecker.GetTypeAtLocation(functionNode)
			}
			callSignatures := utils.CollectAllCallSignatures(ctx.TypeChecker, functionType)
			var expectedRange *core.TextRange
			var expectedType string
			if returnTypeNode := functionNode.Type(); returnTypeNode != nil {
				r := utils.TrimNodeTextRange(ctx.SourceFile, returnTypeNode)
				expectedRange = &r
				expectedType = renderReturnType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(returnTypeNode))
			} else if usesContextualType {
				for _, signature := range callSignatures {
					declaration := checker.Signature_declaration(signature)
					if declaration == nil || declaration.Type() == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
						continue
					}
					r := utils.TrimNodeTextRange(ctx.SourceFile, declaration.Type())
					expectedRange = &r
					expectedType = renderReturnType(ctx.TypeChecker, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
					break
				}
			}
			report := func(message rule.RuleMessage) {
				ctx.ReportDiagnostic(buildUnsafeReturnDiagnostic(
					message,
					primaryRange,
					utils.TrimNodeTextRange(ctx.SourceFile, returnNode),
					renderReturnType(ctx.TypeChecker, returnNodeType),
					expectedRange,
					expectedType,
				))
			}
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
						report(buildUnsafeReturnThisMessage(typeString))
						return
					}
				}
				// If the function return type was not unknown/unknown[], mark usage as unsafeReturn.
				report(buildUnsafeReturnMessage(typeString))
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

			report(buildUnsafeReturnAssignmentMessage(ctx.TypeChecker.TypeToString(sender), ctx.TypeChecker.TypeToString(receiver)))
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if !ast.IsBlock(body) {
					checkReturn(body, utils.TrimNodeTextRange(ctx.SourceFile, node.AsArrowFunction().EqualsGreaterThanToken))
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				argument := node.Expression()
				if argument == nil {
					return
				}

				checkReturn(argument, scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos()))
			},
		}
	},
}
