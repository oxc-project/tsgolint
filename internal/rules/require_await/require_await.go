package require_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildMissingAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAwait",
		Description: "Function has no 'await' expression.",
	}
}

func buildRemoveAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAsync",
		Description: "Remove 'async'.",
	}
}

type scopeInfo struct {
	hasAwait      bool
	isAsyncYield  bool
	functionFlags ast.FunctionFlags
	upper         *scopeInfo
}

func isMethodLike(node *ast.Node) bool {
	return ast.IsMethodOrAccessor(node) || ast.IsConstructorDeclaration(node)
}

// previousSiblingIsMethodLike reports whether the node immediately preceding `node`
// in `list` is a method, constructor, or accessor. Returns false if `node` is the
// first element or not present.
func previousSiblingIsMethodLike(list *ast.NodeList, node *ast.Node) bool {
	if list == nil {
		return false
	}
	for i, m := range list.Nodes {
		if m == node {
			return i > 0 && isMethodLike(list.Nodes[i-1])
		}
	}
	return false
}

// needsPrecedingSemicolon reports whether inserting a `[` or `(` at `node`'s head
// would require a preceding semicolon to avoid being parsed as part of the previous
// expression or statement (the ASI hazard).
//
// This is a pragmatic port of eslint's `needsPrecedingSemicolon` that inspects the
// last non-whitespace character before `node`. If a comment precedes `node` the
// scan stops at whatever character happens to be the comment's trailing content,
// which over-conservatively emits a `;` in a few edge cases — always safe.
func needsPrecedingSemicolon(sourceFile *ast.SourceFile, node *ast.Node) bool {
	pos := scanner.GetTokenPosOfNode(node, sourceFile, false)
	text := sourceFile.Text()

	lastPos := -1
	for i := pos - 1; i >= 0; i-- {
		c := text[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		lastPos = i
		break
	}
	if lastPos < 0 {
		return false
	}
	ch := text[lastPos]
	switch ch {
	case ';', ':', '{', ',', '(', '[':
		return false
	}
	if lastPos >= 1 {
		two := text[lastPos-1 : lastPos+1]
		if two == "=>" || two == "++" || two == "--" {
			return false
		}
	}
	// A closing `}` that ends a class/object-literal member is self-delimiting, so
	// no semicolon is required between it and the next member.
	if ch == '}' {
		if parent := node.Parent; parent != nil {
			switch parent.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration:
				if previousSiblingIsMethodLike(parent.MemberList(), node) {
					return false
				}
			case ast.KindObjectLiteralExpression:
				if previousSiblingIsMethodLike(parent.AsObjectLiteralExpression().Properties, node) {
					return false
				}
			}
		}
	}
	return true
}

// isTypeReferenceNamed reports whether `typeRef` names the built-in type `name`,
// either as a plain identifier (`Promise`) or as a direct globalThis qualifier
// (`globalThis.Promise`). Callers must pass a non-nil `KindTypeReference` node.
func isTypeReferenceNamed(typeRef *ast.TypeReferenceNode, name string) bool {
	tn := typeRef.TypeName
	if tn == nil {
		return false
	}
	if tn.Kind == ast.KindIdentifier {
		return tn.Text() == name
	}
	if !ast.IsQualifiedName(tn) {
		return false
	}
	qualifiedName := tn.AsQualifiedName()
	return qualifiedName.Right.Text() == name &&
		qualifiedName.Left.Kind == ast.KindIdentifier &&
		qualifiedName.Left.Text() == "globalThis"
}

// buildRemoveAsyncFixes computes the list of edits that remove the `async` keyword
// from a function-like declaration. It also adjusts the return type annotation (if
// present) to drop the `Promise<...>` wrapper or rename `AsyncGenerator` to
// `Generator` when the function is a generator.
func buildRemoveAsyncFixes(sourceFile *ast.SourceFile, node *ast.Node, asyncToken *ast.Node, isGenerator bool) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, 3)

	text := sourceFile.Text()
	asyncStart := scanner.GetTokenPosOfNode(asyncToken, sourceFile, false)
	// Remove the `async` keyword plus trailing whitespace, but preserve trailing
	// comments (mirrors `sourceCode.getTokenAfter(asyncToken, {includeComments: true})`).
	removeEnd := scanner.SkipTriviaEx(text, asyncToken.Loc.End(), &scanner.SkipTriviaOptions{StopAtComments: true})

	// ASI: when the next real token starts with `(`, `[`, or a template literal, and
	// the node begins an expression statement or class/object member, we may need to
	// insert a `;` to prevent the replacement from being absorbed into the previous
	// expression.
	addSemicolon := false
	nextToken := scanner.ScanTokenAtPosition(sourceFile, asyncToken.Loc.End())
	if nextToken == ast.KindOpenParenToken ||
		nextToken == ast.KindOpenBracketToken ||
		nextToken == ast.KindNoSubstitutionTemplateLiteral ||
		nextToken == ast.KindTemplateHead {
		if (isAtStartOfExpressionStatement(node) || isMethodLike(node)) && needsPrecedingSemicolon(sourceFile, node) {
			addSemicolon = true
		}
	}

	removeRange := core.NewTextRange(asyncStart, removeEnd)
	if addSemicolon {
		fixes = append(fixes, rule.RuleFixReplaceRange(removeRange, ";"))
	} else {
		fixes = append(fixes, rule.RuleFixRemoveRange(removeRange))
	}

	returnType := node.Type()
	if returnType == nil || returnType.Kind != ast.KindTypeReference {
		return fixes
	}
	typeRef := returnType.AsTypeReferenceNode()
	typeName := typeRef.TypeName
	typeNameStart := scanner.GetTokenPosOfNode(typeName, sourceFile, false)

	switch {
	case isGenerator && isTypeReferenceNamed(typeRef, "AsyncGenerator"):
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(typeNameStart, typeName.Loc.End()),
			"Generator",
		))
	case !isGenerator && isTypeReferenceNamed(typeRef, "Promise") && typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0:
		// Unwrap `Promise<T>` to `T` by deleting `Promise<` and the trailing `>`.
		// `TypeArguments.Loc` spans the content between the angle brackets, so
		// `Pos()-1` is the `<` and `End()` is the `>` (same invariant as parameter
		// lists; `parseBracketedList` sets these bounds).
		openAnglePos := typeRef.TypeArguments.Loc.Pos() - 1
		closeAnglePos := typeRef.TypeArguments.Loc.End()
		fixes = append(fixes,
			rule.RuleFixRemoveRange(core.NewTextRange(closeAnglePos, closeAnglePos+1)),
			rule.RuleFixRemoveRange(core.NewTextRange(typeNameStart, openAnglePos+1)),
		)
	}

	return fixes
}

// isAtStartOfExpressionStatement reports whether `node` is at the same source
// position as an ancestor `ExpressionStatement` — i.e. the node begins an
// expression statement. Mirrors eslint's helper of the same name.
func isAtStartOfExpressionStatement(node *ast.Node) bool {
	start := node.Loc.Pos()
	for a := node.Parent; a != nil && a.Loc.Pos() == start; a = a.Parent {
		if a.Kind == ast.KindExpressionStatement {
			return true
		}
	}
	return false
}

var RequireAwaitRule = rule.Rule{
	Name: "require-await",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var currentScope *scopeInfo

		enterFunction := func(node *ast.Node) {
			currentScope = &scopeInfo{
				hasAwait:      false,
				isAsyncYield:  false,
				functionFlags: ast.FunctionFlagsNormal,
				upper:         currentScope,
			}

			body := node.Body()
			if body != nil && (!ast.IsBlock(body) || len(body.AsBlock().Statements.Nodes) > 0) {
				currentScope.functionFlags = ast.GetFunctionFlags(node)
			}
		}

		exitFunction := func(node *ast.Node) {
			isAsync := currentScope.functionFlags&ast.FunctionFlagsAsync != 0
			isGen := currentScope.functionFlags&ast.FunctionFlagsGenerator != 0
			if isAsync && !currentScope.hasAwait && !(isGen && currentScope.isAsyncYield) {
				// `isAsync` guarantees the node has an `async` modifier.
				asyncToken := utils.FindModifier(node, ast.KindAsyncKeyword)
				ctx.ReportRangeWithSuggestions(
					utils.GetFunctionHeadLoc(ctx.SourceFile, node),
					buildMissingAwaitMessage(),
					func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildRemoveAsyncMessage(),
							FixesArr: buildRemoveAsyncFixes(ctx.SourceFile, node, asyncToken, isGen),
						}}
					},
				)
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
				if currentScope.functionFlags&ast.FunctionFlagsAsync == 0 {
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
				if currentScope.functionFlags&ast.FunctionFlagsGenerator == 0 || argument == nil {
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
				if currentScope == nil || currentScope.hasAwait || currentScope.functionFlags&ast.FunctionFlagsAsync == 0 {
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
