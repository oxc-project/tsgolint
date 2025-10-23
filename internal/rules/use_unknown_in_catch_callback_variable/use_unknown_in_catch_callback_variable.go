package use_unknown_in_catch_callback_variable

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUseUnknownMessageBase(method string) string {
	return fmt.Sprintf("Prefer the safe `: unknown` for a %v callback variable.", method)
}

func buildAddUnknownRestTypeAnnotationSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addUnknownRestTypeAnnotationSuggestion",
		Description: "Add an explicit `: [unknown]` type annotation to the rejection callback rest variable.",
	}
}
func buildAddUnknownTypeAnnotationSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addUnknownTypeAnnotationSuggestion",
		Description: "Add an explicit `: unknown` type annotation to the rejection callback variable.",
	}
}
func buildUseUnknownMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useUnknown",
		Description: buildUseUnknownMessageBase(method),
	}
}
func buildUseUnknownArrayDestructuringPatternMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useUnknownArrayDestructuringPattern",
		Description: buildUseUnknownMessageBase(method) + " The thrown error may not be iterable.",
	}
}
func buildUseUnknownObjectDestructuringPatternMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useUnknownObjectDestructuringPattern",
		Description: buildUseUnknownMessageBase(method) + " The thrown error may be nullable, or may not have the expected shape.",
	}
}
func buildWrongRestTypeAnnotationSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "wrongRestTypeAnnotationSuggestion",
		Description: "Change existing type annotation to `: [unknown]`.",
	}
}
func buildWrongTypeAnnotationSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "wrongTypeAnnotationSuggestion",
		Description: "Change existing type annotation to `: unknown`.",
	}
}

var UseUnknownInCatchCallbackVariableRule = rule.Rule{
	Name: "use-unknown-in-catch-callback-variable",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var collectFlaggedNodes func(node *ast.Node) []*ast.Node

		isFlaggableHandlerType := func(t *checker.Type) bool {
			for _, part := range utils.UnionTypeParts(t) {
				for _, callSignature := range utils.GetCallSignatures(ctx.TypeChecker, part) {
					params := checker.Signature_parameters(callSignature)
					if len(params) == 0 {
						continue
					}

					firstParam := params[0]

					firstParamType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, firstParam)
					decl := firstParam.ValueDeclaration

					if decl != nil && decl.AsParameterDeclaration().DotDotDotToken != nil {
						// a rest arg that's not an array or tuple should definitely be flagged.
						if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, firstParamType) {
							return true
						}
						firstParamType = checker.Checker_getTypeArguments(ctx.TypeChecker, firstParamType)[0]
					}

					if !utils.IsTypeFlagSet(firstParamType, checker.TypeFlagsUnknown) {
						return true
					}
				}
			}
			return false
		}

		collectFlaggedNodes = func(node *ast.Node) []*ast.Node {
			node = ast.SkipParentheses(node)
			switch node.Kind {
			case ast.KindBinaryExpression:
				n := node.AsBinaryExpression()
				if ast.IsLogicalExpression(node) {
					return append(collectFlaggedNodes(n.Left), collectFlaggedNodes(n.Right)...)
				}
				if n.OperatorToken.Kind == ast.KindCommaToken {
					return collectFlaggedNodes(n.Right)
				}
			case ast.KindConditionalExpression:
				n := node.AsConditionalExpression()
				return append(collectFlaggedNodes(n.WhenTrue), collectFlaggedNodes(n.WhenFalse)...)
			case ast.KindArrowFunction, ast.KindFunctionExpression:
				t := ctx.TypeChecker.GetTypeAtLocation(node)
				if isFlaggableHandlerType(t) {
					return []*ast.Node{node}
				}
			}
			return []*ast.Node{}
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				callee := expr.Expression

				if !ast.IsAccessExpression(callee) {
					return
				}

				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !found {
					return
				}

				var method string
				var argIndexToCheck int

				if propertyName == "catch" {
					method = "`catch`"
					argIndexToCheck = 0
				} else if propertyName == "then" {
					method = "`then` rejection"
					argIndexToCheck = 1
				} else {
					return
				}

				if len(expr.Arguments.Nodes) < argIndexToCheck+1 {
					return
				}

				// Argument to check, and all arguments before it, must be "ordinary" arguments (i.e. no spread arguments)
				// promise.catch(f), promise.catch(() => {}), promise.catch(<expression>, <<other-args>>)
				for i, arg := range expr.Arguments.Nodes {
					if ast.IsSpreadElement(arg) {
						return
					}
					if i == argIndexToCheck {
						break
					}
				}

				if !utils.IsThenableType(ctx.TypeChecker, callee, ctx.TypeChecker.GetTypeAtLocation(callee.Expression())) {
					return
				}

				for _, flagged := range collectFlaggedNodes(expr.Arguments.Nodes[argIndexToCheck]) {
					catchParamNode := flagged.Parameters()[0]
					catchParam := catchParamNode.AsParameterDeclaration()
					catchVariable := catchParam.Name()
					catchTypeAnnotation := catchParam.Type

					if catchParam.DotDotDotToken != nil {
						if catchTypeAnnotation == nil {
							ctx.ReportNodeWithSuggestions(catchParamNode, buildUseUnknownMessage(method), func() []rule.RuleSuggestion {
								return []rule.RuleSuggestion{{
									Message:  buildAddUnknownRestTypeAnnotationSuggestionMessage(),
									FixesArr: []rule.RuleFix{rule.RuleFixInsertAfter(catchVariable, ": [unknown]")},
								}}
							})
							continue
						}

						ctx.ReportNodeWithSuggestions(catchParamNode, buildUseUnknownMessage(method), func() []rule.RuleSuggestion {
							return []rule.RuleSuggestion{{
								Message: buildWrongRestTypeAnnotationSuggestionMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixReplace(ctx.SourceFile, catchTypeAnnotation, "[unknown]"),
								},
							}}
						})
						continue
					}

					switch catchVariable.Kind {
					case ast.KindIdentifier:
						if catchTypeAnnotation == nil {
							var fixes []rule.RuleFix
							if utils.IsParenlessArrowFunction(flagged) {
								fixes = []rule.RuleFix{
									rule.RuleFixInsertBefore(ctx.SourceFile, catchVariable, "("),
									rule.RuleFixInsertAfter(catchVariable, ": unknown)"),
								}
							} else {
								insertAfter := catchVariable
								if catchParam.QuestionToken != nil {
									insertAfter = catchParam.QuestionToken
								}
								fixes = []rule.RuleFix{
									rule.RuleFixInsertAfter(insertAfter, ": unknown"),
								}
							}

							ctx.ReportNodeWithSuggestions(catchParamNode, buildUseUnknownMessage(method), func() []rule.RuleSuggestion {
								return []rule.RuleSuggestion{{
									Message:  buildAddUnknownTypeAnnotationSuggestionMessage(),
									FixesArr: fixes,
								}}
							})
							break
						}
						ctx.ReportNodeWithSuggestions(catchParamNode, buildUseUnknownMessage(method), func() []rule.RuleSuggestion {
							return []rule.RuleSuggestion{{
								Message:  buildWrongTypeAnnotationSuggestionMessage(),
								FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, catchTypeAnnotation, "unknown")},
							}}
						})
					case ast.KindArrayBindingPattern:
						ctx.ReportNode(catchParamNode, buildUseUnknownArrayDestructuringPatternMessage(method))
					case ast.KindObjectBindingPattern:
						ctx.ReportNode(catchParamNode, buildUseUnknownObjectDestructuringPatternMessage(method))
					}
				}
			},
		}
	},
}
