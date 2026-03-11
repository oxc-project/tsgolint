package prefer_destructuring

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type destructuringEnabled struct {
	array  bool
	object bool
}

type normalizedOptions struct {
	assignment                              destructuringEnabled
	variable                                destructuringEnabled
	enforceForDeclarationWithTypeAnnotation bool
	enforceForRenamedProperties             bool
}

type tupleEnabledTypes struct {
	AssignmentExpression *DestructuringTypeConfig `json:"AssignmentExpression,omitempty"`
	VariableDeclarator   *DestructuringTypeConfig `json:"VariableDeclarator,omitempty"`
	Array                *bool                    `json:"array,omitempty"`
	Object               *bool                    `json:"object,omitempty"`
}

type tupleAdditionalOptions struct {
	EnforceForDeclarationWithTypeAnnotation *bool `json:"enforceForDeclarationWithTypeAnnotation,omitempty"`
	EnforceForRenamedProperties             *bool `json:"enforceForRenamedProperties,omitempty"`
}

func buildPreferDestructuringMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: fmt.Sprintf("Use %s destructuring.", kind),
	}
}

func boolPtrValue(value *bool) bool {
	return value != nil && *value
}

func normalizeTupleEnabledTypes(result *normalizedOptions, enabledTypes tupleEnabledTypes) {
	if enabledTypes.Array != nil || enabledTypes.Object != nil {
		normalized := destructuringEnabled{
			array:  boolPtrValue(enabledTypes.Array),
			object: boolPtrValue(enabledTypes.Object),
		}
		result.assignment = normalized
		result.variable = normalized
	}

	if enabledTypes.AssignmentExpression != nil {
		result.assignment = destructuringEnabled{
			array:  boolPtrValue(enabledTypes.AssignmentExpression.Array),
			object: boolPtrValue(enabledTypes.AssignmentExpression.Object),
		}
	}
	if enabledTypes.VariableDeclarator != nil {
		result.variable = destructuringEnabled{
			array:  boolPtrValue(enabledTypes.VariableDeclarator.Array),
			object: boolPtrValue(enabledTypes.VariableDeclarator.Object),
		}
	}
}

func parseOptions(options any) normalizedOptions {
	result := normalizedOptions{
		assignment: destructuringEnabled{
			array:  true,
			object: true,
		},
		variable: destructuringEnabled{
			array:  true,
			object: true,
		},
	}

	optsBytes, err := json.Marshal(options)
	if err != nil {
		panic("prefer-destructuring: failed to marshal options: " + err.Error())
	}

	// prefer-destructuring options are a tuple: [enabledTypes?, additionalOptions?].
	var tuple []any
	if err := json.Unmarshal(optsBytes, &tuple); err != nil {
		panic("prefer-destructuring: expected tuple options [enabledTypes?, additionalOptions?]: " + err.Error())
	}

	if len(tuple) > 0 && tuple[0] != nil {
		enabledBytes, marshalErr := json.Marshal(tuple[0])
		if marshalErr != nil {
			panic("prefer-destructuring: failed to marshal tuple enabledTypes: " + marshalErr.Error())
		}

		var enabledTypes tupleEnabledTypes
		if unmarshalErr := json.Unmarshal(enabledBytes, &enabledTypes); unmarshalErr != nil {
			panic("prefer-destructuring: failed to unmarshal tuple enabledTypes: " + unmarshalErr.Error())
		}
		normalizeTupleEnabledTypes(&result, enabledTypes)
	}

	if len(tuple) > 1 && tuple[1] != nil {
		additionalBytes, marshalErr := json.Marshal(tuple[1])
		if marshalErr != nil {
			panic("prefer-destructuring: failed to marshal tuple additionalOptions: " + marshalErr.Error())
		}

		var additional tupleAdditionalOptions
		if unmarshalErr := json.Unmarshal(additionalBytes, &additional); unmarshalErr != nil {
			panic("prefer-destructuring: failed to unmarshal tuple additionalOptions: " + unmarshalErr.Error())
		}
		result.enforceForDeclarationWithTypeAnnotation = boolPtrValue(additional.EnforceForDeclarationWithTypeAnnotation)
		result.enforceForRenamedProperties = boolPtrValue(additional.EnforceForRenamedProperties)
	}

	return result
}

func hasTypeAnnotation(node *ast.Node) bool {
	return node != nil && node.Type() != nil
}

func isIntegerLiteral(node *ast.Node) bool {
	if node == nil || !ast.IsNumericLiteral(node) {
		return false
	}
	valueText := strings.ReplaceAll(node.AsNumericLiteral().Text, "_", "")
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil {
		return false
	}
	return value == float64(int64(value))
}

func isTypeAnyOrIterableType(t *checker.Type, typeChecker *checker.Checker) bool {
	if t == nil {
		return false
	}
	if utils.IsTypeAnyType(t) {
		return true
	}
	if utils.IsUnionType(t) {
		for _, part := range utils.UnionTypeParts(t) {
			if !isTypeAnyOrIterableType(part, typeChecker) {
				return false
			}
		}
		return true
	}
	return utils.GetWellKnownSymbolPropertyOfType(t, "iterator", typeChecker) != nil
}

func getTrimmedText(sourceFile *ast.SourceFile, node *ast.Node) string {
	rangeToUse := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[rangeToUse.Pos():rangeToUse.End()]
}

func getObjectTextForFix(sourceFile *ast.SourceFile, objectNode *ast.Node) string {
	if !ast.IsParenthesizedExpression(objectNode) {
		return getTrimmedText(sourceFile, objectNode)
	}

	inner := ast.SkipParentheses(objectNode)
	if inner == nil {
		return getTrimmedText(sourceFile, objectNode)
	}

	innerText := getTrimmedText(sourceFile, inner)
	if ast.IsBinaryExpression(inner) && inner.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		return "(" + innerText + ")"
	}

	return innerText
}

func reportWithoutFix(ctx rule.RuleContext, node *ast.Node, kind string) {
	ctx.ReportNode(node, buildPreferDestructuringMessage(kind))
}

func reportVariableDeclaratorWithFix(
	ctx rule.RuleContext,
	varDecl *ast.Node,
	leftIdent *ast.Node,
	initializer *ast.Node,
	objectExpr *ast.Node,
	propertyName string,
) {
	ctx.ReportNodeWithFixes(varDecl, buildPreferDestructuringMessage("object"), func() []rule.RuleFix {
		leftRange := utils.TrimNodeTextRange(ctx.SourceFile, leftIdent)
		initializerRange := utils.TrimNodeTextRange(ctx.SourceFile, initializer)
		replacement := fmt.Sprintf("{%s} = %s", propertyName, getObjectTextForFix(ctx.SourceFile, objectExpr))
		return []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(leftRange.Pos(), initializerRange.End()), replacement),
		}
	})
}

func getEnabledForNode(options normalizedOptions, node *ast.Node) destructuringEnabled {
	if node != nil && node.Kind == ast.KindVariableDeclaration {
		return options.variable
	}
	return options.assignment
}

var PreferDestructuringRule = rule.Rule{
	Name: "prefer-destructuring",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		performCheck := func(leftNode *ast.Node, rightNode *ast.Node, reportNode *ast.Node) {
			if leftNode == nil || rightNode == nil || reportNode == nil {
				return
			}

			if ast.IsArrayBindingPattern(leftNode) || ast.IsObjectBindingPattern(leftNode) || ast.IsArrayLiteralExpression(leftNode) || ast.IsObjectLiteralExpression(leftNode) {
				return
			}

			if (hasTypeAnnotation(leftNode) || hasTypeAnnotation(reportNode)) && !opts.enforceForDeclarationWithTypeAnnotation {
				return
			}

			if !ast.IsIdentifier(leftNode) {
				return
			}
			leftName := leftNode.AsIdentifier().Text
			enabledTypes := getEnabledForNode(opts, reportNode)

			if ast.IsPropertyAccessExpression(rightNode) {
				propertyAccess := rightNode.AsPropertyAccessExpression()
				if propertyAccess.QuestionDotToken != nil || propertyAccess.Expression == nil || propertyAccess.Expression.Kind == ast.KindSuperKeyword {
					return
				}

				propertyNameNode := propertyAccess.Name()
				if propertyNameNode == nil || !ast.IsIdentifier(propertyNameNode) {
					return
				}

				propertyName := propertyNameNode.AsIdentifier().Text
				if propertyName != leftName {
					if !opts.enforceForRenamedProperties || !enabledTypes.object {
						return
					}
					reportWithoutFix(ctx, reportNode, "object")
					return
				}

				if !enabledTypes.object {
					return
				}

				if reportNode.Kind == ast.KindVariableDeclaration && !hasTypeAnnotation(reportNode) && !hasTypeAnnotation(leftNode) {
					reportVariableDeclaratorWithFix(ctx, reportNode, leftNode, rightNode, propertyAccess.Expression, propertyName)
					return
				}

				reportWithoutFix(ctx, reportNode, "object")
				return
			}

			if !ast.IsElementAccessExpression(rightNode) {
				return
			}

			elementAccess := rightNode.AsElementAccessExpression()
			if elementAccess.QuestionDotToken != nil || elementAccess.Expression == nil || elementAccess.Expression.Kind == ast.KindSuperKeyword || elementAccess.ArgumentExpression == nil {
				return
			}

			if isIntegerLiteral(elementAccess.ArgumentExpression) {
				objectType := ctx.TypeChecker.GetTypeAtLocation(elementAccess.Expression)
				if isTypeAnyOrIterableType(objectType, ctx.TypeChecker) {
					if enabledTypes.array {
						reportWithoutFix(ctx, reportNode, "array")
					}
					return
				}

				if opts.enforceForRenamedProperties && enabledTypes.object {
					reportWithoutFix(ctx, reportNode, "object")
				}
				return
			}

			if opts.enforceForRenamedProperties && enabledTypes.object {
				reportWithoutFix(ctx, reportNode, "object")
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, true) {
					return
				}
				expr := node.AsBinaryExpression()
				if expr.OperatorToken == nil || expr.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				performCheck(expr.Left, expr.Right, node)
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				initializer := node.Initializer()
				if initializer == nil {
					return
				}
				performCheck(node.Name(), initializer, node)
			},
		}
	},
}
