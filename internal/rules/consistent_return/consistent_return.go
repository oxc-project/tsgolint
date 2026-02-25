package consistent_return

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMissingReturnValueMessage(functionNameWithKind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnValue",
		Description: functionNameWithKind + " expected a return value.",
	}
}

func buildUnexpectedReturnValueMessage(functionNameWithKind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedReturnValue",
		Description: functionNameWithKind + " expected no return value.",
	}
}

type functionState struct {
	node                 *ast.Node
	upper                *functionState
	hasReturn            bool
	hasReturnValue       bool
	hasMismatch          bool
	allowsVoidReturn     bool
	functionNameWithKind string
	messageId            string
}

func getFunctionNameWithKind(sourceFile *ast.SourceFile, node *ast.Node) string {
	kindPrefix := "Function"
	if checker.GetFunctionFlags(node)&checker.FunctionFlagsAsync != 0 {
		kindPrefix = "Async function"
	}

	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return kindPrefix
	}

	nameText := ""
	if ast.IsIdentifier(nameNode) {
		nameText = nameNode.AsIdentifier().Text
	} else {
		nameRange := utils.TrimNodeTextRange(sourceFile, nameNode)
		nameText = strings.TrimSpace(sourceFile.Text()[nameRange.Pos():nameRange.End()])
	}

	if nameText == "" {
		return kindPrefix
	}

	return fmt.Sprintf("%s '%s'", kindPrefix, nameText)
}

func hasExplicitReturnTypeAnnotation(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Type != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Type != nil
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Type != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Type != nil
	default:
		return false
	}
}

func hasUndefinedAndSpecifiedValueType(t *checker.Type) bool {
	parts := utils.UnionTypeParts(t)
	hasUndefined := false
	hasSpecifiedValue := false

	for _, part := range parts {
		if utils.IsTypeUndefinedType(part) {
			hasUndefined = true
			continue
		}

		if utils.IsIntrinsicVoidType(part) || utils.IsTypeFlagSet(part, checker.TypeFlagsNever) {
			continue
		}

		hasSpecifiedValue = true
	}

	return hasUndefined && hasSpecifiedValue
}

func isThenableTypeWithVoidValue(typeChecker *checker.Checker, node *ast.Node, t *checker.Type, visited map[*checker.Type]struct{}) bool {
	if t == nil {
		return false
	}
	if _, ok := visited[t]; ok {
		return false
	}
	visited[t] = struct{}{}

	if utils.IsIntrinsicVoidType(t) {
		return true
	}

	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		return utils.Some(t.Types(), func(part *checker.Type) bool {
			return isThenableTypeWithVoidValue(typeChecker, node, part, visited)
		})
	}

	if !utils.IsThenableType(typeChecker, node, t) {
		return false
	}

	awaitedType := checker.Checker_getAwaitedType(typeChecker, t)
	if awaitedType == nil || awaitedType == t {
		return false
	}

	return isThenableTypeWithVoidValue(typeChecker, node, awaitedType, visited)
}

func isThenableTypeWithUndefinedAndSpecifiedValue(typeChecker *checker.Checker, node *ast.Node, t *checker.Type, visited map[*checker.Type]struct{}) bool {
	if t == nil {
		return false
	}
	if _, ok := visited[t]; ok {
		return false
	}
	visited[t] = struct{}{}

	if hasUndefinedAndSpecifiedValueType(t) {
		return true
	}

	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		return utils.Some(t.Types(), func(part *checker.Type) bool {
			return isThenableTypeWithUndefinedAndSpecifiedValue(typeChecker, node, part, visited)
		})
	}

	if !utils.IsThenableType(typeChecker, node, t) {
		return false
	}

	awaitedType := checker.Checker_getAwaitedType(typeChecker, t)
	if awaitedType == nil || awaitedType == t {
		return false
	}

	return isThenableTypeWithUndefinedAndSpecifiedValue(typeChecker, node, awaitedType, visited)
}

func isReturnVoidOrThenableVoid(ctx rule.RuleContext, functionNode *ast.Node) bool {
	functionType := ctx.TypeChecker.GetTypeAtLocation(functionNode)
	callSignatures := utils.GetCallSignatures(ctx.TypeChecker, functionType)
	if len(callSignatures) == 0 {
		return false
	}

	functionFlags := checker.GetFunctionFlags(functionNode)
	isAsyncFunction := functionFlags&checker.FunctionFlagsAsync != 0
	hasExplicitReturnType := hasExplicitReturnTypeAnnotation(functionNode)

	return utils.Some(callSignatures, func(signature *checker.Signature) bool {
		returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
		if isAsyncFunction {
			if isThenableTypeWithVoidValue(ctx.TypeChecker, functionNode, returnType, map[*checker.Type]struct{}{}) {
				return true
			}

			return hasExplicitReturnType &&
				isThenableTypeWithUndefinedAndSpecifiedValue(ctx.TypeChecker, functionNode, returnType, map[*checker.Type]struct{}{})
		}

		if utils.Some(utils.UnionTypeParts(returnType), utils.IsIntrinsicVoidType) {
			return true
		}

		return hasExplicitReturnType && hasUndefinedAndSpecifiedValueType(returnType)
	})
}

func getHasReturnValue(ctx rule.RuleContext, node *ast.ReturnStatement, treatUndefinedAsUnspecified bool) bool {
	if node.Expression == nil {
		return false
	}

	if !treatUndefinedAsUnspecified {
		return true
	}

	returnValueType := ctx.TypeChecker.GetTypeAtLocation(node.Expression)
	if checker.Type_flags(returnValueType) == checker.TypeFlagsUndefined {
		return false
	}

	return !utils.IsUndefinedLiteral(node.Expression)
}

func reportMismatchedReturnStatement(ctx rule.RuleContext, node *ast.Node, currentFunction *functionState) {
	switch currentFunction.messageId {
	case "missingReturnValue":
		ctx.ReportNode(node, buildMissingReturnValueMessage(currentFunction.functionNameWithKind))
	case "unexpectedReturnValue":
		ctx.ReportNode(node, buildUnexpectedReturnValueMessage(currentFunction.functionNameWithKind))
	}
}

var ConsistentReturnRule = rule.Rule{
	Name: "consistent-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[ConsistentReturnOptions](options, "consistent-return")

		var currentFunction *functionState

		enterFunction := func(node *ast.Node) {
			currentFunction = &functionState{
				node:                 node,
				upper:                currentFunction,
				allowsVoidReturn:     isReturnVoidOrThenableVoid(ctx, node),
				functionNameWithKind: getFunctionNameWithKind(ctx.SourceFile, node),
			}
		}

		exitFunction := func(_node *ast.Node) {
			if currentFunction != nil {
				if !currentFunction.hasMismatch &&
					currentFunction.hasReturn &&
					currentFunction.hasReturnValue &&
					!currentFunction.allowsVoidReturn &&
					currentFunction.node.Flags&ast.NodeFlagsHasImplicitReturn != 0 {
					ctx.ReportNode(currentFunction.node, buildMissingReturnValueMessage(currentFunction.functionNameWithKind))
				}
				currentFunction = currentFunction.upper
			}
		}

		onReturnStatement := func(node *ast.Node) {
			if currentFunction == nil {
				return
			}

			returnStatement := node.AsReturnStatement()
			if returnStatement.Expression == nil && currentFunction.allowsVoidReturn {
				return
			}

			hasReturnValue := getHasReturnValue(ctx, returnStatement, opts.TreatUndefinedAsUnspecified)

			if !currentFunction.hasReturn {
				currentFunction.hasReturn = true
				currentFunction.hasReturnValue = hasReturnValue
				if hasReturnValue {
					currentFunction.messageId = "missingReturnValue"
				} else {
					currentFunction.messageId = "unexpectedReturnValue"
				}
				return
			}

			if currentFunction.hasReturnValue != hasReturnValue {
				currentFunction.hasMismatch = true
				reportMismatchedReturnStatement(ctx, node, currentFunction)
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindConstructor:                              enterFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
			ast.KindReturnStatement:                          onReturnStatement,
		}
	},
}
