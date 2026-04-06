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
	assignment                              *destructuringEnabled
	variable                                *destructuringEnabled
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

func newDestructuringEnabled(array bool, object bool) *destructuringEnabled {
	return &destructuringEnabled{
		array:  array,
		object: object,
	}
}

func destructuringEnabledFromConfig(array *bool, object *bool) *destructuringEnabled {
	return newDestructuringEnabled(boolPtrValue(array), boolPtrValue(object))
}

func normalizeTupleEnabledTypes(result *normalizedOptions, enabledTypes tupleEnabledTypes) {
	hasFlatConfig := enabledTypes.Array != nil || enabledTypes.Object != nil
	hasPerNodeConfig := enabledTypes.AssignmentExpression != nil || enabledTypes.VariableDeclarator != nil

	if hasFlatConfig {
		result.assignment = destructuringEnabledFromConfig(enabledTypes.Array, enabledTypes.Object)
		result.variable = destructuringEnabledFromConfig(enabledTypes.Array, enabledTypes.Object)
	} else if hasPerNodeConfig {
		// Match ESLint core semantics: omitted node-type entries are treated as disabled.
		result.assignment = nil
		result.variable = nil
	}

	if enabledTypes.AssignmentExpression != nil {
		result.assignment = destructuringEnabledFromConfig(enabledTypes.AssignmentExpression.Array, enabledTypes.AssignmentExpression.Object)
	}
	if enabledTypes.VariableDeclarator != nil {
		result.variable = destructuringEnabledFromConfig(enabledTypes.VariableDeclarator.Array, enabledTypes.VariableDeclarator.Object)
	}
}

func parseOptions(options any) normalizedOptions {
	result := normalizedOptions{
		assignment: newDestructuringEnabled(true, true),
		variable:   newDestructuringEnabled(true, true),
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

func hasDeclarationTypeAnnotation(leftNode *ast.Node, reportNode *ast.Node) bool {
	if hasTypeAnnotation(leftNode) {
		return true
	}
	return reportNode != nil && reportNode.Kind == ast.KindVariableDeclaration && hasTypeAnnotation(reportNode)
}

func shouldSkipTypeAnnotatedLeft(leftNode *ast.Node, reportNode *ast.Node, enforceForDeclarationWithTypeAnnotation bool) bool {
	if leftNode == nil || enforceForDeclarationWithTypeAnnotation {
		return false
	}

	if ast.IsArrayBindingPattern(leftNode) || ast.IsIdentifier(leftNode) || ast.IsObjectBindingPattern(leftNode) {
		return hasDeclarationTypeAnnotation(leftNode, reportNode)
	}

	return false
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

func isArrayLiteralIntegerIndexAccess(node *ast.Node) bool {
	if !ast.IsElementAccessExpression(node) {
		return false
	}
	elementAccess := node.AsElementAccessExpression()
	return elementAccess.ArgumentExpression != nil && isIntegerLiteral(elementAccess.ArgumentExpression)
}

func getTrimmedText(sourceFile *ast.SourceFile, node *ast.Node) string {
	rangeToUse := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[rangeToUse.Pos():rangeToUse.End()]
}

func getObjectTextForFix(sourceFile *ast.SourceFile, objectNode *ast.Node) string {
	textNode := objectNode
	if ast.IsParenthesizedExpression(objectNode) {
		inner := ast.SkipParentheses(objectNode)
		if inner != nil {
			textNode = inner
		}
	}

	objectText := getTrimmedText(sourceFile, textNode)
	if ast.GetExpressionPrecedence(textNode) < ast.OperatorPrecedenceAssignment {
		return "(" + objectText + ")"
	}

	return objectText
}

func containsCommentText(text string) bool {
	return strings.Contains(text, "/*") || strings.Contains(text, "//")
}

func maybeFixVariableDeclaratorIntoObjectDestructuring(
	ctx rule.RuleContext,
	leftIdent *ast.Node,
	initializer *ast.Node,
	objectExpr *ast.Node,
	propertyName string,
) []rule.RuleFix {
	if leftIdent == nil || initializer == nil || objectExpr == nil {
		return nil
	}

	objectRange := utils.TrimNodeTextRange(ctx.SourceFile, objectExpr)
	leftRange := utils.TrimNodeTextRange(ctx.SourceFile, leftIdent)
	initializerRange := utils.TrimNodeTextRange(ctx.SourceFile, initializer)
	sourceText := ctx.SourceFile.Text()

	if containsCommentText(sourceText[leftRange.Pos():objectRange.Pos()]) ||
		containsCommentText(sourceText[objectRange.End():initializerRange.End()]) {
		return nil
	}

	replacement := fmt.Sprintf("{%s} = %s", propertyName, getObjectTextForFix(ctx.SourceFile, objectExpr))
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(leftRange.Pos(), initializerRange.End()), replacement),
	}
}

func getEnabledForNode(options normalizedOptions, node *ast.Node) *destructuringEnabled {
	if node != nil && node.Kind == ast.KindVariableDeclaration {
		return options.variable
	}
	return options.assignment
}

func shouldCheck(options normalizedOptions, node *ast.Node, destructuringType string) bool {
	enabled := getEnabledForNode(options, node)
	if enabled == nil {
		return false
	}
	switch destructuringType {
	case "array":
		return enabled.array
	case "object":
		return enabled.object
	default:
		return false
	}
}

func getStringLiteralLikeText(node *ast.Node) (string, bool) {
	switch {
	case ast.IsStringLiteral(node):
		return node.AsStringLiteral().Text, true
	case ast.IsNoSubstitutionTemplateLiteral(node):
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	default:
		return "", false
	}
}

func sameObjectPropertyName(leftName string, rightNode *ast.Node) bool {
	switch {
	case ast.IsPropertyAccessExpression(rightNode):
		propertyAccess := rightNode.AsPropertyAccessExpression()
		if propertyAccess.QuestionDotToken != nil || propertyAccess.Expression == nil || propertyAccess.Expression.Kind == ast.KindSuperKeyword {
			return false
		}
		propertyNameNode := propertyAccess.Name()
		return propertyNameNode != nil && ast.IsIdentifier(propertyNameNode) && propertyNameNode.AsIdentifier().Text == leftName
	case ast.IsElementAccessExpression(rightNode):
		elementAccess := rightNode.AsElementAccessExpression()
		if elementAccess.QuestionDotToken != nil || elementAccess.Expression == nil || elementAccess.Expression.Kind == ast.KindSuperKeyword || elementAccess.ArgumentExpression == nil {
			return false
		}
		propertyName, ok := getStringLiteralLikeText(elementAccess.ArgumentExpression)
		return ok && propertyName == leftName
	default:
		return false
	}
}

func report(ctx rule.RuleContext, node *ast.Node, kind string, fixes []rule.RuleFix) {
	msg := buildPreferDestructuringMessage(kind)
	if len(fixes) > 0 {
		ctx.ReportNodeWithFixes(node, msg, func() []rule.RuleFix {
			return fixes
		})
		return
	}
	ctx.ReportNode(node, msg)
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

			if shouldSkipTypeAnnotatedLeft(leftNode, reportNode, opts.enforceForDeclarationWithTypeAnnotation) {
				return
			}

			if isArrayLiteralIntegerIndexAccess(rightNode) && rightNode.Expression() != nil && rightNode.Expression().Kind != ast.KindSuperKeyword {
				objectType := ctx.TypeChecker.GetTypeAtLocation(rightNode.Expression())
				if !isTypeAnyOrIterableType(objectType, ctx.TypeChecker) {
					if !opts.enforceForRenamedProperties || !shouldCheck(opts, reportNode, "object") {
						return
					}
					report(ctx, reportNode, "object", nil)
					return
				}
			}

			if !ast.IsIdentifier(leftNode) {
				return
			}
			leftName := leftNode.AsIdentifier().Text

			switch {
			case !ast.IsPropertyAccessExpression(rightNode) && !ast.IsElementAccessExpression(rightNode):
				return
			case ast.IsPropertyAccessExpression(rightNode):
				propertyAccess := rightNode.AsPropertyAccessExpression()
				if propertyAccess.Expression == nil || propertyAccess.Expression.Kind == ast.KindSuperKeyword || propertyAccess.QuestionDotToken != nil {
					return
				}
				if nameNode := propertyAccess.Name(); nameNode != nil && ast.IsPrivateIdentifier(nameNode) {
					return
				}
			case ast.IsElementAccessExpression(rightNode):
				elementAccess := rightNode.AsElementAccessExpression()
				if elementAccess.Expression == nil || elementAccess.Expression.Kind == ast.KindSuperKeyword || elementAccess.QuestionDotToken != nil || elementAccess.ArgumentExpression == nil {
					return
				}
			}

			if isArrayLiteralIntegerIndexAccess(rightNode) {
				if shouldCheck(opts, reportNode, "array") {
					report(ctx, reportNode, "array", nil)
				}
				return
			}

			var fixes []rule.RuleFix
			if reportNode.Kind == ast.KindVariableDeclaration && !hasDeclarationTypeAnnotation(leftNode, reportNode) && ast.IsPropertyAccessExpression(rightNode) {
				propertyAccess := rightNode.AsPropertyAccessExpression()
				propertyNameNode := propertyAccess.Name()
				if propertyNameNode != nil && ast.IsIdentifier(propertyNameNode) && propertyNameNode.AsIdentifier().Text == leftName {
					fixes = maybeFixVariableDeclaratorIntoObjectDestructuring(
						ctx,
						leftNode,
						rightNode,
						propertyAccess.Expression,
						propertyNameNode.AsIdentifier().Text,
					)
				}
			}

			if shouldCheck(opts, reportNode, "object") && opts.enforceForRenamedProperties {
				report(ctx, reportNode, "object", fixes)
				return
			}

			if shouldCheck(opts, reportNode, "object") && sameObjectPropertyName(leftName, rightNode) {
				report(ctx, reportNode, "object", fixes)
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
				if ast.IsVarUsing(node) || ast.IsVarAwaitUsing(node) {
					return
				}
				performCheck(node.Name(), initializer, node)
			},
		}
	},
}
