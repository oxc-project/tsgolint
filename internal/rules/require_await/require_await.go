package require_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMissingAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAwait",
		Description: "Function has no 'await' expression.",
	}
}

// func buildRemoveAsyncMessage() rule.RuleMessage {
// return rule.RuleMessage{
// Id:          "removeAsync",
// Description: "Remove 'async'.",
// }
// }

type scopeInfo struct {
	hasAwait      bool
	isAsyncYield  bool
	functionFlags checker.FunctionFlags
	upper         *scopeInfo
}

var RequireAwaitRule = rule.Rule{
	Name: "require-await",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var currentScope *scopeInfo

		enterFunction := func(node *ast.Node) {
			currentScope = &scopeInfo{
				hasAwait:      false,
				isAsyncYield:  false,
				functionFlags: checker.FunctionFlagsNormal,
				upper:         currentScope,
			}

			body := node.Body()
			if body != nil && (!ast.IsBlock(body) || len(body.AsBlock().Statements.Nodes) > 0) {
				currentScope.functionFlags = checker.GetFunctionFlags(node)
			}
		}

		exitFunction := func(node *ast.Node) {
			if currentScope.functionFlags&checker.FunctionFlagsAsync != 0 && !currentScope.hasAwait && !(currentScope.functionFlags&checker.FunctionFlagsGenerator != 0 && currentScope.isAsyncYield) {
				// TODO(port): implement suggestions
				// // If the function belongs to a method definition or
				// // property, then the function's range may not include the
				// // `async` keyword and we should look at the parent instead.
				// const nodeWithAsyncKeyword =
				//   (node.parent.type === AST_NODE_TYPES.MethodDefinition &&
				//     node.parent.value === node) ||
				//   (node.parent.type === AST_NODE_TYPES.Property &&
				//     node.parent.method &&
				//     node.parent.value === node)
				//     ? node.parent
				//     : node;
				//
				// const asyncToken = nullThrows(
				//   context.sourceCode.getFirstToken(
				//     nodeWithAsyncKeyword,
				//     token => token.value === 'async',
				//   ),
				//   'The node is an async function, so it must have an "async" token.',
				// );
				//
				// const asyncRange: Readonly<AST.Range> = [
				//   asyncToken.range[0],
				//   nullThrows(
				//     context.sourceCode.getTokenAfter(asyncToken, {
				//       includeComments: true,
				//     }),
				//     'There will always be a token after the "async" keyword.',
				//   ).range[0],
				// ] as const;
				//
				// // Removing the `async` keyword can cause parsing errors if the
				// // current statement is relying on automatic semicolon insertion.
				// // If ASI is currently being used, then we should replace the
				// // `async` keyword with a semicolon.
				// const nextToken = nullThrows(
				//   context.sourceCode.getTokenAfter(asyncToken),
				//   'There will always be a token after the "async" keyword.',
				// );
				// const addSemiColon =
				//   nextToken.type === AST_TOKEN_TYPES.Punctuator &&
				//   (nextToken.value === '[' || nextToken.value === '(') &&
				//   (nodeWithAsyncKeyword.type === AST_NODE_TYPES.MethodDefinition ||
				//     isStartOfExpressionStatement(nodeWithAsyncKeyword)) &&
				//   needsPrecedingSemicolon(context.sourceCode, nodeWithAsyncKeyword);
				//
				// const changes = [
				//   { range: asyncRange, replacement: addSemiColon ? ';' : undefined },
				// ];
				//
				// // If there's a return type annotation and it's a
				// // `Promise<T>`, we can also change the return type
				// // annotation to just `T` as part of the suggestion.
				// // Alternatively, if the function is a generator and
				// // the return type annotation is `AsyncGenerator<T>`,
				// // then we can change it to `Generator<T>`.
				// if (
				//   node.returnType?.typeAnnotation.type ===
				//   AST_NODE_TYPES.TSTypeReference
				// ) {
				//   if (scopeInfo.isGen) {
				//     if (hasTypeName(node.returnType.typeAnnotation, 'AsyncGenerator')) {
				//       changes.push({
				//         range: node.returnType.typeAnnotation.typeName.range,
				//         replacement: 'Generator',
				//       });
				//     }
				//   } else if (
				//     hasTypeName(node.returnType.typeAnnotation, 'Promise') &&
				//     node.returnType.typeAnnotation.typeArguments != null
				//   ) {
				//     const openAngle = nullThrows(
				//       context.sourceCode.getFirstToken(
				//         node.returnType.typeAnnotation,
				//         token =>
				//           token.type === AST_TOKEN_TYPES.Punctuator &&
				//           token.value === '<',
				//       ),
				//       'There are type arguments, so the angle bracket will exist.',
				//     );
				//     const closeAngle = nullThrows(
				//       context.sourceCode.getLastToken(
				//         node.returnType.typeAnnotation,
				//         token =>
				//           token.type === AST_TOKEN_TYPES.Punctuator &&
				//           token.value === '>',
				//       ),
				//       'There are type arguments, so the angle bracket will exist.',
				//     );
				//     changes.push(
				//       // Remove the closing angled bracket.
				//       { range: closeAngle.range, replacement: undefined },
				//       // Remove the "Promise" identifier
				//       // and the opening angled bracket.
				//       {
				//         range: [
				//           node.returnType.typeAnnotation.typeName.range[0],
				//           openAngle.range[1],
				//         ],
				//         replacement: undefined,
				//       },
				//     );
				//   }
				// }
				//
				// context.report({
				//   loc: getFunctionHeadLoc(node, context.sourceCode),
				//   node,
				//   messageId: 'missingAwait',
				//   data: {
				//     name: upperCaseFirst(getFunctionNameWithKind(node)),
				//   },
				//   suggest: [
				//     {
				//       messageId: 'removeAsync',
				//       fix: (fixer): RuleFix[] =>
				//         changes.map(change =>
				//           change.replacement != null
				//             ? fixer.replaceTextRange(change.range, change.replacement)
				//             : fixer.removeRange(change.range),
				//         ),
				//     },
				//   ],
				// });
				// TODO(port): getFunctionHeadLoc
				ctx.ReportNode(node, buildMissingAwaitMessage())
			}

			currentScope = currentScope.upper
		}

		markAsHasAwait := func() {
			if currentScope != nil {
				currentScope.hasAwait = true
			}
		}

		return rule.RuleListeners{
			// from isFunctionLikeDeclarationKind
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindConstructor:                              enterFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction: func(node *ast.Node) {
				enterFunction(node)
				// check body-less async arrow function.
				// ignore `async () => await foo` because it's obviously correct
				if currentScope.functionFlags&checker.FunctionFlagsAsync == 0 {
					return
				}

				body := ast.SkipParentheses(node.Body())
				if ast.IsBlock(body) || ast.IsAwaitExpression(body) {
					return
				}

				if utils.IsThenableType(ctx.TypeChecker, body, ctx.TypeChecker.GetTypeAtLocation(body)) {
					markAsHasAwait()
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): exitFunction,

			ast.KindAwaitExpression: func(_node *ast.Node) { markAsHasAwait() },
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier != nil {
					markAsHasAwait()
				}
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarAwaitUsing(node) {
					markAsHasAwait()
				}
			},
			/**
			 * Mark `scopeInfo.isAsyncYield` to `true` if it
			 *  1) delegates async generator function
			 *    or
			 *  2) yields thenable type
			 */
			ast.KindYieldExpression: func(node *ast.Node) {
				if currentScope == nil || currentScope.isAsyncYield {
					return
				}
				argument := node.Expression()
				if currentScope.functionFlags&checker.FunctionFlagsGenerator == 0 || argument == nil {
					return
				}

				if ast.IsLiteralExpression(argument) {
					// ignoring this as for literals we don't need to check the definition
					// eg : async function* run() { yield* 1 }
					return
				}

				if node.AsYieldExpression().AsteriskToken == nil {
					if utils.IsThenableType(ctx.TypeChecker, argument, ctx.TypeChecker.GetTypeAtLocation(argument)) {
						currentScope.isAsyncYield = true
					}
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(argument)
				hasAsyncYield := utils.TypeRecurser(t, func(t *checker.Type) bool {
					return utils.GetWellKnownSymbolPropertyOfType(t, "asyncIterator", ctx.TypeChecker) != nil
				})
				if hasAsyncYield {
					currentScope.isAsyncYield = true
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				if currentScope == nil || currentScope.hasAwait || currentScope.functionFlags&checker.FunctionFlagsAsync == 0 {
					return
				}

				expr := node.Expression()
				if expr != nil && utils.IsThenableType(ctx.TypeChecker, expr, ctx.TypeChecker.GetTypeAtLocation(expr)) {
					markAsHasAwait()
				}
			},
		}
	},
}
