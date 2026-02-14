package no_unnecessary_type_conversion

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildUnnecessaryTypeConversionDiagnostic(conversion, expression core.TextRange, expressionType string) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Message: rule.RuleMessage{
			Id:          "unnecessaryTypeConversion",
			Description: "This type conversion does not change the type or value of the expression.",
		},
		Range: conversion,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Range: expression,
				Label: fmt.Sprintf("This expression already has type '%s'.", expressionType),
			},
		},
	}
}

func buildSuggestRemoveMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemove",
		Description: "Remove the type conversion.",
	}
}

func buildSuggestSatisfiesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestSatisfies",
		Description: "Instead, assert that the value satisfies the primitive type.",
	}
}

var NoUnnecessaryTypeConversionRule = rule.Rule{
	Name: "no-unnecessary-type-conversion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()

		doesUnderlyingTypeMatchFlag := func(t *checker.Type, typeFlag checker.TypeFlags) bool {
			return utils.Every(utils.UnionTypeParts(t), func(part *checker.Type) bool {
				return utils.IsTypeFlagSet(part, typeFlag)
			})
		}

		isEmptyStringLiteral := func(node *ast.Node) bool {
			return ast.IsStringLiteral(node) && node.AsStringLiteral().Text == ""
		}

		isEnumType := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike)
		}

		isEnumMemberType := func(t *checker.Type) bool {
			symbol := checker.Type_symbol(t)
			return symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0
		}

		isAllNumberLiteralIntegers := func(t *checker.Type) bool {
			parts := utils.UnionTypeParts(t)
			if len(parts) == 0 {
				return false
			}

			for _, part := range parts {
				if !utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral) {
					return false
				}

				literal := part.AsLiteralType()
				if literal == nil {
					return false
				}

				value, err := strconv.ParseFloat(literal.String(), 64)
				if err != nil || math.Trunc(value) != value {
					return false
				}
			}

			return true
		}

		isNodeParenthesized := func(node *ast.Node) bool {
			return ast.IsParenthesizedExpression(node.Parent) &&
				node.Parent.AsParenthesizedExpression().Expression == node
		}

		isObjectExpressionInOneLineReturn := func(node *ast.Node, innerNode *ast.Node) bool {
			return ast.IsArrowFunction(node.Parent) &&
				node.Parent.AsArrowFunction().Body == node &&
				ast.IsObjectLiteralExpression(innerNode)
		}

		isWeakPrecedenceParent := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			switch parent.Kind {
			case ast.KindPostfixUnaryExpression,
				ast.KindPrefixUnaryExpression,
				ast.KindBinaryExpression,
				ast.KindConditionalExpression,
				ast.KindAwaitExpression:
				return true
			}

			if ast.IsPropertyAccessExpression(parent) {
				return parent.AsPropertyAccessExpression().Expression == node
			}
			if ast.IsElementAccessExpression(parent) {
				return parent.AsElementAccessExpression().Expression == node
			}
			if ast.IsCallExpression(parent) || ast.IsNewExpression(parent) {
				return parent.Expression() == node
			}
			if ast.IsTaggedTemplateExpression(parent) {
				return parent.AsTaggedTemplateExpression().Tag == node
			}

			return false
		}

		getNodeText := func(node *ast.Node) string {
			r := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return sourceText[r.Pos():r.End()]
		}

		buildWrappingFix := func(node *ast.Node, innerNodes []*ast.Node, wrap func(code ...string) string) rule.RuleFix {
			innerCodes := make([]string, len(innerNodes))
			for i, innerNode := range innerNodes {
				code := getNodeText(innerNode)
				if !utils.IsStrongPrecedenceNode(innerNode) || isObjectExpressionInOneLineReturn(node, innerNode) {
					code = "(" + code + ")"
				}
				innerCodes[i] = code
			}

			if wrap == nil {
				return rule.RuleFixReplace(ctx.SourceFile, node, strings.Join(innerCodes, ""))
			}

			code := wrap(innerCodes...)
			if isWeakPrecedenceParent(node) && !isNodeParenthesized(node) {
				code = "(" + code + ")"
			}

			return rule.RuleFixReplace(ctx.SourceFile, node, code)
		}

		buildSuggestions := func(node *ast.Node, primitiveType string, innerNodes ...*ast.Node) []rule.RuleSuggestion {
			return []rule.RuleSuggestion{
				{
					Message: buildSuggestRemoveMessage(),
					FixesArr: []rule.RuleFix{
						buildWrappingFix(node, innerNodes, nil),
					},
				},
				{
					Message: buildSuggestSatisfiesMessage(),
					FixesArr: []rule.RuleFix{
						buildWrappingFix(node, innerNodes, func(code ...string) string {
							return code[0] + " satisfies " + primitiveType
						}),
					},
				},
			}
		}

		reportBinaryPlus := func(node *ast.Node) {
			expr := node.AsBinaryExpression()

			if isEmptyStringLiteral(expr.Right) {
				leftType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Left)
				if doesUnderlyingTypeMatchFlag(leftType, checker.TypeFlagsStringLike) {
					ctx.ReportDiagnosticWithSuggestions(
						buildUnnecessaryTypeConversionDiagnostic(
							core.NewTextRange(expr.Left.End(), node.End()),
							utils.TrimNodeTextRange(ctx.SourceFile, expr.Left),
							ctx.TypeChecker.TypeToString(leftType),
						),
						func() []rule.RuleSuggestion {
							return buildSuggestions(node, "string", expr.Left)
						})
				}
				return
			}

			if isEmptyStringLiteral(expr.Left) {
				rightType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Right)
				if doesUnderlyingTypeMatchFlag(rightType, checker.TypeFlagsStringLike) {
					rightStart := utils.TrimNodeTextRange(ctx.SourceFile, expr.Right).Pos()
					ctx.ReportDiagnosticWithSuggestions(
						buildUnnecessaryTypeConversionDiagnostic(
							core.NewTextRange(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), rightStart),
							utils.TrimNodeTextRange(ctx.SourceFile, expr.Right),
							ctx.TypeChecker.TypeToString(rightType),
						),
						func() []rule.RuleSuggestion {
							return buildSuggestions(node, "string", expr.Right)
						},
					)
				}
			}
		}

		reportPlusEquals := func(node *ast.Node) {
			expr := node.AsBinaryExpression()
			if !isEmptyStringLiteral(expr.Right) {
				return
			}
			if !ast.IsIdentifier(expr.Left) {
				return
			}

			leftType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Left)
			if !doesUnderlyingTypeMatchFlag(leftType, checker.TypeFlagsStringLike) {
				return
			}

			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					utils.TrimNodeTextRange(ctx.SourceFile, node),
					utils.TrimNodeTextRange(ctx.SourceFile, expr.Left),
					ctx.TypeChecker.TypeToString(leftType),
				),
				func() []rule.RuleSuggestion {
					removeFix := buildWrappingFix(node, []*ast.Node{expr.Left}, nil)
					if ast.IsExpressionStatement(node.Parent) {
						removeFix = rule.RuleFixRemove(ctx.SourceFile, node.Parent)
					}

					return []rule.RuleSuggestion{
						{
							Message:  buildSuggestRemoveMessage(),
							FixesArr: []rule.RuleFix{removeFix},
						},
						{
							Message: buildSuggestSatisfiesMessage(),
							FixesArr: []rule.RuleFix{
								buildWrappingFix(node, []*ast.Node{expr.Left}, func(code ...string) string {
									return code[0] + " satisfies string"
								}),
							},
						},
					}
				},
			)
		}

		reportBuiltInConversionCall := func(node *ast.CallExpression) {
			if !ast.IsIdentifier(node.Expression) || len(node.Arguments.Nodes) != 1 {
				return
			}

			callee := node.Expression
			arg := node.Arguments.Nodes[0]
			if ast.IsSpreadElement(arg) {
				return
			}

			var (
				typeFlag checker.TypeFlags
				ok       bool
				builtins []string
			)
			switch callee.AsIdentifier().Text {
			case "BigInt":
				typeFlag = checker.TypeFlagsBigIntLike
				builtins = []string{"BigInt", "BigIntConstructor"}
				ok = true
			case "Boolean":
				typeFlag = checker.TypeFlagsBooleanLike
				builtins = []string{"Boolean", "BooleanConstructor"}
				ok = true
			case "Number":
				typeFlag = checker.TypeFlagsNumberLike
				builtins = []string{"Number", "NumberConstructor"}
				ok = true
			case "String":
				typeFlag = checker.TypeFlagsStringLike
				builtins = []string{"String", "StringConstructor"}
				ok = true
			}

			if !ok {
				return
			}

			calleeType := ctx.TypeChecker.GetTypeAtLocation(callee)
			if !utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, calleeType, builtins...) {
				return
			}

			argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg)
			if !doesUnderlyingTypeMatchFlag(argType, typeFlag) {
				return
			}

			primitiveType := strings.ToLower(callee.AsIdentifier().Text)
			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					utils.TrimNodeTextRange(ctx.SourceFile, callee),
					utils.TrimNodeTextRange(ctx.SourceFile, arg),
					ctx.TypeChecker.TypeToString(argType),
				),
				func() []rule.RuleSuggestion {
					return buildSuggestions(node.AsNode(), primitiveType, arg)
				},
			)
		}

		reportStringToStringCall := func(node *ast.CallExpression) {
			if !ast.IsPropertyAccessExpression(node.Expression) || len(node.Arguments.Nodes) != 0 {
				return
			}

			memberExpr := node.Expression.AsPropertyAccessExpression()
			if memberExpr.Name().Text() != "toString" {
				return
			}

			objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, memberExpr.Expression)
			if isEnumType(objectType) || isEnumMemberType(objectType) {
				return
			}

			if !doesUnderlyingTypeMatchFlag(objectType, checker.TypeFlagsStringLike) {
				return
			}

			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					core.NewTextRange(memberExpr.Name().Pos(), node.AsNode().End()),
					utils.TrimNodeTextRange(ctx.SourceFile, memberExpr.Expression),
					ctx.TypeChecker.TypeToString(objectType),
				),
				func() []rule.RuleSuggestion {
					return buildSuggestions(node.AsNode(), "string", memberExpr.Expression)
				},
			)
		}

		reportUnaryPlus := func(node *ast.Node) {
			expr := node.AsPrefixUnaryExpression()
			if expr.Operator != ast.KindPlusToken {
				return
			}

			argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Operand)
			if !doesUnderlyingTypeMatchFlag(argType, checker.TypeFlagsNumberLike) {
				return
			}

			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					core.NewTextRange(
						utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(),
						utils.TrimNodeTextRange(ctx.SourceFile, expr.Operand).Pos(),
					),
					utils.TrimNodeTextRange(ctx.SourceFile, expr.Operand),
					ctx.TypeChecker.TypeToString(argType),
				),
				func() []rule.RuleSuggestion {
					return buildSuggestions(node, "number", expr.Operand)
				},
			)
		}

		reportDoubleBang := func(node *ast.Node) {
			expr := node.AsPrefixUnaryExpression()
			if expr.Operator != ast.KindExclamationToken ||
				!ast.IsPrefixUnaryExpression(node.Parent) ||
				node.Parent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken ||
				node.Parent.AsPrefixUnaryExpression().Operand != node {
				return
			}

			outerNode := node.Parent
			argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Operand)
			if !doesUnderlyingTypeMatchFlag(argType, checker.TypeFlagsBooleanLike) {
				return
			}

			outerStart := utils.TrimNodeTextRange(ctx.SourceFile, outerNode).Pos()
			innerStart := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					core.NewTextRange(outerStart, innerStart+1),
					utils.TrimNodeTextRange(ctx.SourceFile, expr.Operand),
					ctx.TypeChecker.TypeToString(argType),
				),
				func() []rule.RuleSuggestion {
					return buildSuggestions(outerNode, "boolean", expr.Operand)
				},
			)
		}

		reportDoubleTilde := func(node *ast.Node) {
			expr := node.AsPrefixUnaryExpression()
			if expr.Operator != ast.KindTildeToken ||
				!ast.IsPrefixUnaryExpression(node.Parent) ||
				node.Parent.AsPrefixUnaryExpression().Operator != ast.KindTildeToken ||
				node.Parent.AsPrefixUnaryExpression().Operand != node {
				return
			}

			argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Operand)
			if !isAllNumberLiteralIntegers(argType) {
				return
			}

			outerNode := node.Parent
			outerStart := utils.TrimNodeTextRange(ctx.SourceFile, outerNode).Pos()
			innerStart := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
			ctx.ReportDiagnosticWithSuggestions(
				buildUnnecessaryTypeConversionDiagnostic(
					core.NewTextRange(outerStart, innerStart+1),
					utils.TrimNodeTextRange(ctx.SourceFile, expr.Operand),
					ctx.TypeChecker.TypeToString(argType),
				),
				func() []rule.RuleSuggestion {
					return buildSuggestions(outerNode, "number", expr.Operand)
				},
			)
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				switch expr.OperatorToken.Kind {
				case ast.KindPlusToken:
					reportBinaryPlus(node)
				case ast.KindPlusEqualsToken:
					reportPlusEquals(node)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				reportBuiltInConversionCall(callExpr)
				reportStringToStringCall(callExpr)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				reportUnaryPlus(node)
				reportDoubleBang(node)
				reportDoubleTilde(node)
			},
		}
	},
}
