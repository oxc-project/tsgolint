package no_misused_promises

import (
	"fmt"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildConditionalMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditional",
		Description: "Expected non-Promise value in a boolean conditional.",
	}
}
func buildPredicateMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicate",
		Description: "Expected a non-Promise value to be returned.",
	}
}
func buildSpreadMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "spread",
		Description: "Expected a non-Promise value to be spread in an object.",
	}
}
func buildVoidReturnArgumentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnArgument",
		Description: "Promise returned in function argument where a void return was expected.",
	}
}
func buildVoidReturnAttributeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnAttribute",
		Description: "Promise-returning function provided to attribute where a void return was expected.",
	}
}
func buildVoidReturnInheritedMethodMessage(heritageTypeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnInheritedMethod",
		Description: fmt.Sprintf("Promise-returning method provided where a void return was expected by extended/implemented type '%v'.", heritageTypeName),
	}
}
func buildVoidReturnPropertyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnProperty",
		Description: "Promise-returning function provided to property where a void return was expected.",
	}
}
func buildVoidReturnReturnValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnReturnValue",
		Description: "Promise-returning function provided to return value where a void return was expected.",
	}
}
func buildVoidReturnVariableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnVariable",
		Description: "Promise-returning function provided to variable where a void return was expected.",
	}
}

var NoMisusedPromisesRule = rule.Rule{
	Name: "no-misused-promises",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoMisusedPromisesOptions](options, "no-misused-promises")

		// Helper function to create default ChecksVoidReturnOptions
		defaultChecksVoidReturnOpts := func() *ChecksVoidReturnOptions {
			return &ChecksVoidReturnOptions{
				Arguments:        true,
				Attributes:       true,
				InheritedMethods: true,
				Properties:       true,
				Returns:          true,
				Variables:        true,
			}
		}

		// Handle checksVoidReturn which can be either a boolean or an object
		checksVoidReturn := true
		var checksVoidReturnOpts *ChecksVoidReturnOptions

		if opts.ChecksVoidReturn != nil {
			switch v := opts.ChecksVoidReturn.(type) {
			case bool:
				checksVoidReturn = v
				if v {
					// If true, use default nested options
					checksVoidReturnOpts = defaultChecksVoidReturnOpts()
				}
			case map[string]interface{}:
				// It's an object, unmarshal it as ChecksVoidReturnOptions
				checksVoidReturn = true
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					// If marshaling fails, use default options
					checksVoidReturnOpts = defaultChecksVoidReturnOpts()
					break
				}
				var voidReturnOpts ChecksVoidReturnOptions
				if err := json.Unmarshal(jsonBytes, &voidReturnOpts); err == nil {
					checksVoidReturnOpts = &voidReturnOpts
				} else {
					// If unmarshaling fails, use default options
					checksVoidReturnOpts = defaultChecksVoidReturnOpts()
				}
			}
		} else {
			// Default behavior
			checksVoidReturnOpts = defaultChecksVoidReturnOpts()
		}

		anySignatureIsThenableType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			return utils.Some(utils.GetCallSignatures(ctx.TypeChecker, t), func(sig *checker.Signature) bool {
				return utils.IsThenableType(ctx.TypeChecker, node, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig))
			})
		}

		returnsThenable := func(node *ast.Node) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))
			return utils.Some(utils.UnionTypeParts(t), func(t *checker.Type) bool {
				return anySignatureIsThenableType(node, t)
			})
		}

		checkArrayPredicates := func(node *ast.Node) {
			parent := node.Parent

			if !ast.IsCallExpression(parent) {
				return
			}

			expr := parent.AsCallExpression()
			arguments := expr.Arguments.Nodes
			if len(arguments) == 0 {
				return
			}

			callback := arguments[0]

			if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, expr) && returnsThenable(callback) {
				ctx.ReportNode(callback, buildPredicateMessage())
			}
		}

		isFunctionParam := func(
			param *ast.Symbol,
			node *ast.Node,
		) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeOfSymbolAtLocation(param, node))
			if t == nil {
				return false
			}
			return utils.Some(utils.UnionTypeParts(t), func(t *checker.Type) bool {
				return len(utils.GetCallSignatures(ctx.TypeChecker, t)) != 0
			})
		}

		// Variation on the thenable check which requires all forms of the type (read:
		// alternates in a union) to be thenable. Otherwise, you might be trying to
		// check if something is defined or undefined and get caught because one of the
		// branches is thenable.
		isAlwaysThenable := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)

			for _, subType := range utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, t)) {
				thenProp := checker.Checker_getPropertyOfType(ctx.TypeChecker, subType, "then")

				// If one of the alternates has no then property, it is not thenable in all
				// cases.
				if thenProp == nil {
					return false
				}

				// We walk through each variation of the then property. Since we know it
				// exists at this point, we just need at least one of the alternates to
				// be of the right form to consider it thenable.
				thenType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(thenProp, node)
				hasThenableSignature := false
				for _, subType := range utils.UnionTypeParts(thenType) {
					for _, signature := range utils.GetCallSignatures(ctx.TypeChecker, subType) {
						params := checker.Signature_parameters(signature)
						if len(params) != 0 && isFunctionParam(params[0], node) {
							hasThenableSignature = true
						}
					}

					// We only need to find one variant of the then property that has a
					// function signature for it to be thenable.
					if hasThenableSignature {
						break
					}
				}

				// If no flavors of the then property are thenable, we don't consider the
				// overall type to be thenable
				if !hasThenableSignature {
					return false
				}
			}

			// If all variants are considered thenable (i.e. haven't returned false), we
			// consider the overall type thenable
			return true
		}

		checkedNodes := map[*ast.Node](struct{}){}

		var checkConditional func(
			node *ast.Expression,
			isTestExpr bool,
		)
		checkConditional = func(
			node *ast.Expression,
			isTestExpr bool,
		) {
			if node == nil || ast.IsAssignmentExpression(node, false) {
				return
			}
			// prevent checking the same node multiple times
			if _, ok := checkedNodes[node]; ok {
				return
			}
			checkedNodes[node] = struct{}{}

			node = ast.SkipParentheses(node)

			if ast.IsBinaryExpression(node) && ast.IsLogicalExpression(node) {
				expr := node.AsBinaryExpression()
				// ignore the left operand for nullish coalescing expressions not in a context of a test expression
				if expr.OperatorToken.Kind != ast.KindQuestionQuestionToken || isTestExpr {
					checkConditional(expr.Left, isTestExpr)
				}
				// we ignore the right operand when not in a context of a test expression
				if isTestExpr {
					checkConditional(expr.Right, isTestExpr)
				}
				return
			}

			if isAlwaysThenable(node) {
				ctx.ReportNode(node, buildConditionalMessage())
			}
		}

		getMemberIfExists := func(
			t *checker.Type,
			memberName string,
		) *ast.Symbol {
			// TODO(port)
			// const escapedMemberName = ts.escapeLeadingUnderscores(memberName);
			symbol := checker.Type_symbol(t)
			if symbol != nil {
				symbol = symbol.Members[memberName]
			}
			if symbol != nil {
				return symbol
			}

			return checker.Checker_getPropertyOfType(ctx.TypeChecker, t, memberName)
		}

		isVoidReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			hadVoidReturn := false
			for _, t := range utils.UnionTypeParts(t) {
				for _, sig := range utils.GetCallSignatures(ctx.TypeChecker, t) {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
					// If a certain positional argument accepts both thenable and void returns,
					// a promise-returning function is valid
					if utils.IsThenableType(ctx.TypeChecker, node, returnType) {
						return false
					}

					hadVoidReturn = hadVoidReturn || utils.IsTypeFlagSet(returnType, checker.TypeFlagsVoid)
				}
			}
			return hadVoidReturn
		}

		/**
		 * Checks `heritageType` for a member named `memberName` that returns void; reports the
		 * 'voidReturnInheritedMethod' message if found.
		 * @param nodeMember Node member that returns a Promise
		 * @param heritageType Heritage type to check against
		 * @param memberName Name of the member to check for
		 */
		checkHeritageTypeForMemberReturningVoid := func(
			nodeMember *ast.Node,
			heritageType *checker.Type,
			memberName string,
		) {
			heritageMember := getMemberIfExists(heritageType, memberName)
			if heritageMember == nil {
				return
			}
			memberType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
				heritageMember,
				nodeMember,
			)
			if !isVoidReturningFunctionType(nodeMember, memberType) {
				return
			}
			ctx.ReportNode(nodeMember, buildVoidReturnInheritedMethodMessage(ctx.TypeChecker.TypeToString(heritageType)))
		}

		checkJSXAttribute := func(node *ast.JsxAttribute) {
			if node.Initializer == nil || node.Initializer.Kind != ast.KindJsxExpression {
				return
			}
			expressionContainer := node.Initializer.AsJsxExpression()
			expression := expressionContainer.Expression
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Initializer, checker.ContextFlagsNone)
			if contextualType != nil && isVoidReturningFunctionType(node.Initializer, contextualType) && returnsThenable(expression) {
				ctx.ReportNode(node.Initializer, buildVoidReturnAttributeMessage())
			}
		}

		checkSpread := func(node *ast.Node) {
			if utils.IsThenableType(ctx.TypeChecker, node.Expression(), nil) {
				ctx.ReportNode(node.Expression(), buildSpreadMessage())
			}
		}

		isThenableReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			return utils.Some(utils.UnionTypeParts(t), func(t *checker.Type) bool {
				return anySignatureIsThenableType(node, t)
			})
		}

		var checkThenableOrVoidArgument func(
			node *ast.Expression,
			t *checker.Type,
			index int,
			thenableReturnIndices *[]int,
			voidReturnIndices *[]int,
		)
		checkThenableOrVoidArgument = func(
			node *ast.Expression,
			t *checker.Type,
			index int,
			thenableReturnIndices *[]int,
			voidReturnIndices *[]int,
		) {
			if isThenableReturningFunctionType(node.Expression(), t) {
				(*thenableReturnIndices) = append(*thenableReturnIndices, index)
			} else if isVoidReturningFunctionType(node.Expression(), t) &&
				// If a certain argument accepts both thenable and void returns,
				// a promise-returning function is valid
				!slices.Contains(*thenableReturnIndices, index) {

				(*voidReturnIndices) = append(*voidReturnIndices, index)
			}
			contextualType := checker.Checker_getContextualTypeForArgumentAtIndex(ctx.TypeChecker, node, index)

			if contextualType != t {
				checkThenableOrVoidArgument(
					node,
					contextualType,
					index,
					thenableReturnIndices,
					voidReturnIndices,
				)
			}
		}

		// Get the positions of arguments which are void functions (and not also
		// thenable functions). These are the candidates for the void-return check at
		// the current call site.
		// If the function parameters end with a 'rest' parameter, then we consider
		// the array type parameter (e.g. '...args:Array<SomeType>') when determining
		// if trailing arguments are candidates.
		voidFunctionArguments := func(
			node *ast.Expression,
		) []int {
			// 'new' can be used without any arguments, as in 'let b = new Object;'
			// In this case, there are no argument positions to check, so return early.
			if node.Arguments() == nil {
				return []int{}
			}
			thenableReturnIndices := []int{}
			voidReturnIndices := []int{}
			t := ctx.TypeChecker.GetTypeAtLocation(node.Expression())

			// We can't use checker.getResolvedSignature because it prefers an early '() => void' over a later '() => Promise<void>'
			// See https://github.com/microsoft/TypeScript/issues/48077

			for _, subType := range utils.UnionTypeParts(t) {
				// Standard function calls and `new` have two different types of signatures
				var signatures []*checker.Signature
				if ast.IsCallExpression(node) {
					signatures = utils.GetCallSignatures(ctx.TypeChecker, subType)
				} else {
					signatures = utils.GetConstructSignatures(ctx.TypeChecker, subType)
				}
				for _, signature := range signatures {
					for index, parameter := range checker.Signature_parameters(signature) {
						decl := parameter.ValueDeclaration
						t := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameter, node.Expression())

						// If this is a array 'rest' parameter, check all of the argument indices
						// from the current argument to the end.
						if decl != nil && utils.IsRestParameterDeclaration(decl) {
							if checker.Checker_isArrayType(ctx.TypeChecker, t) {
								// Unwrap 'Array<MaybeVoidFunction>' to 'MaybeVoidFunction',
								// so that we'll handle it in the same way as a non-rest
								// 'param: MaybeVoidFunction'
								t = checker.Checker_getTypeArguments(ctx.TypeChecker, t)[0]
								for i := index; i < len(node.Arguments()); i++ {
									checkThenableOrVoidArgument(
										node,
										t,
										i,
										&thenableReturnIndices,
										&voidReturnIndices,
									)
								}
							} else if checker.IsTupleType(t) {
								// Check each type in the tuple - for example, [boolean, () => void] would
								// add the index of the second tuple parameter to 'voidReturnIndices'
								typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
								for i := index; i < len(node.Arguments()) && i-index < len(typeArgs); i++ {
									checkThenableOrVoidArgument(
										node,
										typeArgs[i-index],
										i,
										&thenableReturnIndices,
										&voidReturnIndices,
									)
								}
							}
						} else {
							checkThenableOrVoidArgument(
								node,
								t,
								index,
								&thenableReturnIndices,
								&voidReturnIndices,
							)
						}
					}
				}
			}

			for _, index := range thenableReturnIndices {
				at := slices.Index(voidReturnIndices, index)
				if at >= 0 {
					voidReturnIndices = slices.Delete(voidReturnIndices, at, at+1)
				}
			}

			return voidReturnIndices
		}

		checkArguments := func(
			node *ast.Expression,
		) {
			voidArgs := voidFunctionArguments(node)
			if len(voidArgs) == 0 {
				return
			}

			for index, argument := range node.Arguments() {
				if !slices.Contains(voidArgs, index) {
					continue
				}
				if returnsThenable(argument) {
					ctx.ReportNode(argument, buildVoidReturnArgumentMessage())
				}
			}
		}

		checkClassLikeOrInterfaceNode := func(
			node *ast.Node,
		) {
			heritageClauses := utils.GetHeritageClauses(node)
			if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
				return
			}

			heritageTypes := utils.Flatten(utils.Map(heritageClauses.Nodes, func(h *ast.Node) []*checker.Type {
				return utils.Map(h.AsHeritageClause().Types.Nodes, func(n *ast.Node) *checker.Type {
					return ctx.TypeChecker.GetTypeAtLocation(n)
				})
			}))

			for _, nodeMember := range node.Members() {
				if nodeMember.Name() == nil {
					// Call/construct/index signatures don't have names. TS allows call signatures to mismatch,
					// and construct signatures can't be async.
					// TODO - Once we're able to use `checker.isTypeAssignableTo` (v8), we can check an index
					// signature here against its compatible index signatures in `heritageTypes`
					continue
				}
				if !(ast.IsIdentifier(nodeMember.Name()) || ast.IsPrivateIdentifier(nodeMember.Name()) || ast.IsStringLiteral(nodeMember.Name()) || ast.IsNumericLiteral(nodeMember.Name()) || ast.IsBigIntLiteral(nodeMember.Name())) {
					continue
				}
				memberName := nodeMember.Name().Text()
				if ast.IsStatic(nodeMember) {
					continue
				}
				if !returnsThenable(nodeMember) {
					continue
				}
				for _, heritageType := range heritageTypes {
					checkHeritageTypeForMemberReturningVoid(
						nodeMember,
						heritageType,
						memberName,
					)
				}
			}
		}

		checkProperty := func(node *ast.Node) {
			if ast.IsPropertyAssignment(node) {
				property := node.AsPropertyAssignment()
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, property.Initializer, checker.ContextFlagsNone)

				if contextualType != nil && isVoidReturningFunctionType(
					property.Initializer,
					contextualType,
				) && returnsThenable(property.Initializer) {
					if ast.IsFunctionLike(property.Initializer) {
						returnType := property.Initializer.Type()
						if returnType != nil {
							ctx.ReportNode(returnType, buildVoidReturnPropertyMessage())
						} else {
							ctx.ReportNode(
								// TODO(port): getFunctionHeadLoc(functionNode, context.sourceCode)
								property.Initializer,
								buildVoidReturnPropertyMessage(),
							)
						}
					} else {
						ctx.ReportNode(property.Initializer, buildVoidReturnPropertyMessage())
					}
				}
			} else if ast.IsShorthandPropertyAssignment(node) {
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Name(), checker.ContextFlagsNone)
				if contextualType != nil &&
					isVoidReturningFunctionType(node.Name(), contextualType) &&
					returnsThenable(node.Name()) {
					ctx.ReportNode(node.Name(), buildVoidReturnPropertyMessage())
				}
			} else if ast.IsMethodDeclaration(node) {
				if ast.IsComputedPropertyName(node.Name()) {
					return
				}
				obj := node.Parent

				// Below condition isn't satisfied unless something goes wrong,
				// but is needed for type checking.
				// 'node' does not include class method declaration so 'obj' is
				// always an object literal expression, but after converting 'node'
				// to TypeScript AST, its type includes MethodDeclaration which
				// does include the case of class method declaration.
				if !ast.IsObjectLiteralExpression(obj) {
					return
				}

				if !returnsThenable(node) {
					return
				}
				objType := checker.Checker_getContextualType(ctx.TypeChecker, obj, checker.ContextFlagsNone)
				if objType == nil {
					return
				}
				propertySymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objType, node.Name().Text())
				if propertySymbol == nil {
					return
				}

				contextualType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
					propertySymbol,
					node.Name(),
				)

				if isVoidReturningFunctionType(node.Name(), contextualType) {
					if ast.IsMethodDeclaration(node) {
					}

					if node.Type() != nil {
						ctx.ReportNode(node.Type(), buildVoidReturnPropertyMessage())
					} else {
						ctx.ReportNode(
							// TODO(port): getFunctionHeadLoc(functionNode, context.sourceCode)
							node,
							buildVoidReturnPropertyMessage(),
						)
					}
				}
			}
		}

		/**
		 * A syntactic check to see if an annotated type is maybe a function type.
		 * This is a perf optimization to help avoid requesting types where possible
		 */
		isPossiblyFunctionType := func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindConditionalType,
				ast.KindConstructorType,
				ast.KindFunctionType,
				ast.KindImportType,
				ast.KindIndexedAccessType,
				ast.KindInferType,
				ast.KindIntersectionType,
				ast.KindQualifiedName,
				ast.KindThisType,
				ast.KindTypeOperator,
				ast.KindTypeQuery,
				ast.KindTypeReference,
				ast.KindUnionType:
				return true

			case ast.KindTypeLiteral:
				return utils.Some(node.AsTypeLiteralNode().Members.Nodes, func(member *ast.Node) bool {
					return member.Kind == ast.KindCallSignature || member.Kind == ast.KindConstructSignature
				})

			case ast.KindAbstractKeyword,
				ast.KindAnyKeyword,
				ast.KindArrayType,
				ast.KindAsyncKeyword,
				ast.KindBigIntKeyword,
				ast.KindBooleanKeyword,
				ast.KindDeclareKeyword,
				ast.KindExportKeyword,
				ast.KindIntrinsicKeyword,
				ast.KindLiteralType,
				ast.KindMappedType,
				ast.KindNamedTupleMember,
				ast.KindNeverKeyword,
				ast.KindNullKeyword,
				ast.KindNumberKeyword,
				ast.KindObjectKeyword,
				ast.KindOptionalType,
				ast.KindPrivateKeyword,
				ast.KindProtectedKeyword,
				ast.KindPublicKeyword,
				ast.KindReadonlyKeyword,
				ast.KindRestType,
				ast.KindStaticKeyword,
				ast.KindStringKeyword,
				ast.KindSymbolKeyword,
				ast.KindTemplateLiteralType,
				ast.KindTupleType,
				ast.KindTypePredicate,
				ast.KindUndefinedKeyword,
				ast.KindUnknownKeyword,
				ast.KindVoidKeyword:
				return false
			}
			return false
		}

		checkReturnStatement := func(node *ast.ReturnStatement) {
			if node.Expression == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			functionNode := (func() *ast.Node {
				current := node.Parent
				for current != nil && !ast.IsFunctionLike(current) {
					current = current.Parent
				}
				return current
			})()

			if functionNode != nil && functionNode.Type() != nil && !isPossiblyFunctionType(functionNode.Type()) {
				return
			}

			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Expression, checker.ContextFlagsNone)
			if contextualType != nil &&
				isVoidReturningFunctionType(
					node.Expression,
					contextualType,
				) && returnsThenable(node.Expression) {
				ctx.ReportNode(node.Expression, buildVoidReturnReturnValueMessage())
			}
		}

		checkAssignment := func(node *ast.BinaryExpression) {
			varType := ctx.TypeChecker.GetTypeAtLocation(node.Left)
			if !isVoidReturningFunctionType(node.Left, varType) {
				return
			}

			if returnsThenable(node.Right) {
				ctx.ReportNode(node.Right, buildVoidReturnVariableMessage())
			}
		}

		checkVariableDeclaration := func(node *ast.VariableDeclaration) {
			if node.Initializer == nil ||
				node.Type == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			if !isPossiblyFunctionType(node.Type) {
				return
			}

			varType := ctx.TypeChecker.GetTypeAtLocation(node.Name())
			if !isVoidReturningFunctionType(node.Initializer, varType) {
				return
			}

			if returnsThenable(node.Initializer) {
				ctx.ReportNode(node.Initializer, buildVoidReturnVariableMessage())
			}
		}

		listeners := rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if opts.ChecksConditionals {
					checkConditional(node, false)
				}

				if checksVoidReturn && checksVoidReturnOpts != nil && checksVoidReturnOpts.Variables && ast.IsAssignmentExpression(node, false) {
					checkAssignment(node.AsBinaryExpression())
				}
			},
		}
		if opts.ChecksConditionals {
			listeners[ast.KindPropertyAccessExpression] = checkArrayPredicates
			listeners[ast.KindElementAccessExpression] = checkArrayPredicates

			listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()
				if expr.Operator == ast.KindExclamationToken {
					checkConditional(expr.Operand, true)
				}
			}

			listeners[ast.KindConditionalExpression] = func(node *ast.Node) { checkConditional(node.AsConditionalExpression().Condition, true) }
			listeners[ast.KindForStatement] = func(node *ast.Node) { checkConditional(node.AsForStatement().Condition, true) }
			listeners[ast.KindDoStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindWhileStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindIfStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
		}

		if checksVoidReturn && checksVoidReturnOpts != nil {
			if checksVoidReturnOpts.Arguments {
				listeners[ast.KindCallExpression] = checkArguments
				listeners[ast.KindNewExpression] = checkArguments
			}
			if checksVoidReturnOpts.Attributes {
				listeners[ast.KindJsxAttribute] = func(node *ast.Node) { checkJSXAttribute(node.AsJsxAttribute()) }
			}
			if checksVoidReturnOpts.InheritedMethods {
				listeners[ast.KindClassDeclaration] = checkClassLikeOrInterfaceNode
				listeners[ast.KindClassExpression] = checkClassLikeOrInterfaceNode
				listeners[ast.KindInterfaceDeclaration] = checkClassLikeOrInterfaceNode
			}
			if checksVoidReturnOpts.Properties {
				listeners[ast.KindPropertyAssignment] = checkProperty
				listeners[ast.KindMethodDeclaration] = checkProperty
				listeners[ast.KindShorthandPropertyAssignment] = checkProperty
			}
			if checksVoidReturnOpts.Returns {
				listeners[ast.KindReturnStatement] = func(node *ast.Node) { checkReturnStatement(node.AsReturnStatement()) }
			}
			if checksVoidReturnOpts.Variables {
				listeners[ast.KindVariableDeclaration] = func(node *ast.Node) { checkVariableDeclaration(node.AsVariableDeclaration()) }
			}

		}
		if opts.ChecksSpreads {
			listeners[ast.KindSpreadElement] = checkSpread
			listeners[ast.KindSpreadAssignment] = checkSpread
		}

		return listeners

	},
}
