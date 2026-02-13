package strict_void_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildAsyncFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "asyncFunc",
		Description: "Async function used in a context where a void function is expected.",
	}
}

func buildNonVoidFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonVoidFunc",
		Description: "Value-returning function used in a context where a void function is expected.",
	}
}

func buildNonVoidReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonVoidReturn",
		Description: "Value returned in a context where a void return is expected.",
	}
}

var StrictVoidReturnRule = rule.Rule{
	Name: "strict-void-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[StrictVoidReturnOptions](options, "strict-void-return")

		allowedReturnTypeFlags := checker.TypeFlagsVoid | checker.TypeFlagsNever | checker.TypeFlagsUndefined
		if opts.AllowReturnAny {
			allowedReturnTypeFlags |= checker.TypeFlagsAny
		}

		isAllowedType := func(t *checker.Type) bool {
			return utils.Every(utils.UnionTypeParts(t), func(typePart *checker.Type) bool {
				return utils.IsTypeFlagSet(typePart, allowedReturnTypeFlags)
			})
		}

		isVoidReturningFunctionType := func(t *checker.Type) bool {
			returnTypes := []*checker.Type{}
			for _, typePart := range utils.UnionTypeParts(t) {
				for _, signature := range utils.GetCallSignatures(ctx.TypeChecker, typePart) {
					returnTypes = append(returnTypes, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
				}
			}
			return len(returnTypes) > 0 && utils.Every(returnTypes, func(returnType *checker.Type) bool {
				return utils.Every(utils.UnionTypeParts(returnType), func(typePart *checker.Type) bool {
					return utils.IsTypeFlagSet(typePart, checker.TypeFlagsVoid)
				})
			})
		}

		var reportIfNonVoidFunction func(funcNode *ast.Node)
		reportIfNonVoidFunction = func(funcNode *ast.Node) {
			actualType := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(funcNode))
			if utils.Every(utils.GetCallSignatures(ctx.TypeChecker, actualType), func(signature *checker.Signature) bool {
				return isAllowedType(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
			}) {
				return
			}

			if !ast.IsArrowFunction(funcNode) && !ast.IsFunctionExpression(funcNode) && !ast.IsMethodDeclaration(funcNode) {
				ctx.ReportNode(funcNode, buildNonVoidFuncMessage())
				return
			}

			functionFlags := checker.GetFunctionFlags(funcNode)
			if functionFlags&checker.FunctionFlagsGenerator != 0 {
				ctx.ReportNode(funcNode, buildNonVoidFuncMessage())
				return
			}
			if functionFlags&checker.FunctionFlagsAsync != 0 {
				ctx.ReportNode(funcNode, buildAsyncFuncMessage())
				return
			}

			body := funcNode.Body()
			if body != nil && !ast.IsBlock(body) {
				ctx.ReportNode(body, buildNonVoidReturnMessage())
				return
			}

			if returnTypeNode := funcNode.Type(); returnTypeNode != nil && returnTypeNode.Kind != ast.KindVoidKeyword {
				ctx.ReportNode(returnTypeNode, buildNonVoidFuncMessage())
				return
			}

			var visit func(node *ast.Node)
			visit = func(node *ast.Node) {
				if node == nil {
					return
				}
				if node != funcNode && ast.IsFunctionLike(node) {
					return
				}

				if ast.IsReturnStatement(node) {
					returnStatement := node.AsReturnStatement()
					if returnStatement.Expression != nil {
						returnType := ctx.TypeChecker.GetTypeAtLocation(returnStatement.Expression)
						if !isAllowedType(returnType) {
							ctx.ReportNode(node, buildNonVoidReturnMessage())
						}
					}
				}

				node.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(funcNode)
		}

		checkExpressionNode := func(node *ast.Node) bool {
			expectedType := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
			if expectedType != nil && isVoidReturningFunctionType(expectedType) {
				reportIfNonVoidFunction(node)
				return true
			}
			return false
		}

		checkFunctionCallNode := func(callNode *ast.Expression) {
			funcType := ctx.TypeChecker.GetTypeAtLocation(callNode.Expression())
			signatures := utils.Flatten(utils.Map(utils.UnionTypeParts(funcType), func(typePart *checker.Type) []*checker.Signature {
				if ast.IsCallExpression(callNode) {
					return utils.GetCallSignatures(ctx.TypeChecker, typePart)
				}
				return utils.GetConstructSignatures(ctx.TypeChecker, typePart)
			}))

			for argIdx, argNode := range callNode.Arguments() {
				if argNode.Kind == ast.KindSpreadElement {
					continue
				}

				if checkExpressionNode(argNode) {
					continue
				}

				argExpectedReturnTypes := []*checker.Type{}
				for _, sig := range signatures {
					parameters := checker.Signature_parameters(sig)
					if argIdx >= len(parameters) {
						continue
					}
					paramType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameters[argIdx], callNode.Expression())
					for _, paramTypePart := range utils.UnionTypeParts(paramType) {
						for _, paramSignature := range utils.GetCallSignatures(ctx.TypeChecker, paramTypePart) {
							argExpectedReturnTypes = append(argExpectedReturnTypes, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, paramSignature))
						}
					}
				}

				if len(argExpectedReturnTypes) == 0 {
					continue
				}

				anyVoid := utils.Some(argExpectedReturnTypes, func(returnType *checker.Type) bool {
					return utils.IsTypeFlagSet(returnType, checker.TypeFlagsVoid)
				})
				allVoidLike := utils.Every(argExpectedReturnTypes, func(returnType *checker.Type) bool {
					return utils.IsTypeFlagSet(
						returnType,
						checker.TypeFlagsVoidLike|
							checker.TypeFlagsUndefined|
							checker.TypeFlagsNull|
							checker.TypeFlagsAny|
							checker.TypeFlagsNever,
					)
				})
				if anyVoid && allVoidLike {
					reportIfNonVoidFunction(argNode)
				}
			}
		}

		getMemberName := func(nameNode *ast.Node) string {
			if nameNode == nil {
				return ""
			}

			if symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode); symbol != nil && symbol.Name != "" {
				return symbol.Name
			}

			switch nameNode.Kind {
			case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral:
				return nameNode.Text()
			case ast.KindComputedPropertyName:
				expr := nameNode.AsComputedPropertyName().Expression
				if expr != nil {
					switch expr.Kind {
					case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral:
						return expr.Text()
					}
				}
			}

			return ""
		}

		getBaseMemberTypes := func(memberNode *ast.Node) []*checker.Type {
			classLikeNode := memberNode.Parent
			if classLikeNode == nil {
				return nil
			}
			heritageClauses := utils.GetHeritageClauses(classLikeNode)
			if heritageClauses == nil {
				return nil
			}

			memberNameNode := memberNode.Name()
			if memberNameNode == nil {
				return nil
			}

			memberSymbol := ctx.TypeChecker.GetSymbolAtLocation(memberNameNode)
			if memberSymbol == nil {
				return nil
			}

			baseMemberTypes := []*checker.Type{}
			for _, heritageClause := range heritageClauses.Nodes {
				for _, heritageTypeNode := range heritageClause.AsHeritageClause().Types.Nodes {
					heritageType := ctx.TypeChecker.GetTypeAtLocation(heritageTypeNode)
					heritageMember := checker.Checker_getPropertyOfType(ctx.TypeChecker, heritageType, memberSymbol.Name)
					if heritageMember == nil {
						continue
					}
					baseMemberTypes = append(baseMemberTypes, ctx.TypeChecker.GetTypeOfSymbolAtLocation(heritageMember, memberNode))
				}
			}

			return baseMemberTypes
		}

		checkObjectMethodNode := func(methodNode *ast.Node) {
			if methodNode.Name() != nil && ast.IsComputedPropertyName(methodNode.Name()) {
				return
			}

			objType := checker.Checker_getContextualType(ctx.TypeChecker, methodNode.Parent, checker.ContextFlagsNone)
			if objType == nil {
				return
			}
			memberName := getMemberName(methodNode.Name())
			if memberName == "" {
				return
			}
			propertySymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objType, memberName)
			if propertySymbol == nil {
				return
			}
			expectedType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(propertySymbol, methodNode)
			if isVoidReturningFunctionType(expectedType) {
				reportIfNonVoidFunction(methodNode)
			}
		}

		checkClassMethodNode := func(methodNode *ast.Node) {
			if methodNode.AsMethodDeclaration().Body == nil {
				return
			}
			for _, baseMemberType := range getBaseMemberTypes(methodNode) {
				if isVoidReturningFunctionType(baseMemberType) {
					reportIfNonVoidFunction(methodNode)
					return
				}
			}
		}

		checkClassPropertyNode := func(propertyNode *ast.Node) {
			for _, baseMemberType := range getBaseMemberTypes(propertyNode) {
				if isVoidReturningFunctionType(baseMemberType) && propertyNode.AsPropertyDeclaration().Initializer != nil {
					reportIfNonVoidFunction(propertyNode.AsPropertyDeclaration().Initializer)
					return
				}
			}
			if propertyNode.AsPropertyDeclaration().Initializer != nil {
				checkExpressionNode(propertyNode.AsPropertyDeclaration().Initializer)
			}
		}

		return rule.RuleListeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				for _, elem := range node.AsArrayLiteralExpression().Elements.Nodes {
					if elem != nil && elem.Kind != ast.KindSpreadElement {
						checkExpressionNode(elem)
					}
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if body != nil && !ast.IsBlock(body) {
					checkExpressionNode(body)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ast.IsAssignmentExpression(node, false) {
					checkExpressionNode(node.AsBinaryExpression().Right)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				checkFunctionCallNode(node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				checkFunctionCallNode(node)
			},
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				if attr.Initializer == nil || attr.Initializer.Kind != ast.KindJsxExpression {
					return
				}
				expression := attr.Initializer.AsJsxExpression().Expression
				if expression != nil && !ast.IsOmittedExpression(expression) {
					checkExpressionNode(expression)
				}
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				if ast.IsObjectLiteralExpression(node.Parent) {
					checkObjectMethodNode(node)
					return
				}
				checkClassMethodNode(node)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkClassPropertyNode(node)
			},
			ast.KindPropertyAssignment: func(node *ast.Node) {
				checkExpressionNode(node.Initializer())
			},
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				checkExpressionNode(node.Name())
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				if node.AsReturnStatement().Expression != nil {
					checkExpressionNode(node.AsReturnStatement().Expression)
				}
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if node.AsVariableDeclaration().Initializer != nil {
					checkExpressionNode(node.AsVariableDeclaration().Initializer)
				}
			},
		}
	},
}
