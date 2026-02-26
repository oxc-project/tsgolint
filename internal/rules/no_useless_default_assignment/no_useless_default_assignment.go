package no_useless_default_assignment

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const noUselessDefaultAssignmentRuleName = "no-useless-default-assignment"

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPreferOptionalSyntaxMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferOptionalSyntax",
		Description: "Using `= undefined` to make a parameter optional adds unnecessary runtime logic. Use the `?` optional syntax instead.",
	}
}

func buildUselessDefaultAssignmentMessage(assignmentType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessDefaultAssignment",
		Description: fmt.Sprintf("Default value is useless because the %s is not nullish. This default assignment will never be used.", assignmentType),
		Help:        "Remove the default assignment",
	}
}

func buildUselessDefaultAssignmentWithTypeMessage(assignmentType string, typeText string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessDefaultAssignment",
		Description: fmt.Sprintf("Default value is useless because the %s has type `%s` (not nullish). This default assignment will never be used.", assignmentType, typeText),
		Help:        "Remove the default assignment",
	}
}

func buildUselessUndefinedMessage(pluralAssignmentType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "uselessUndefined",
		Description: fmt.Sprintf("Default value is useless because it is undefined. Optional %s are already undefined by default.", pluralAssignmentType),
		Help:        "Remove the default assignment",
	}
}

func canBeUndefined(t *checker.Type) bool {
	if t == nil {
		return false
	}
	if utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) {
		return true
	}
	return slices.ContainsFunc(utils.UnionTypeParts(t), utils.IsTypeUndefinedType)
}

func getPropertyName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	if ast.IsComputedPropertyName(node) {
		node = node.AsComputedPropertyName().Expression
	}

	if ast.IsIdentifier(node) {
		return node.AsIdentifier().Text, true
	}
	if ast.IsLiteralExpression(node) || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return node.Text(), true
	}

	return "", false
}

func getArrayElementType(typeChecker *checker.Checker, arrayType *checker.Type, elementIndex int) *checker.Type {
	if checker.IsTupleType(arrayType) {
		tupleArgs := checker.Checker_getTypeArguments(typeChecker, arrayType)
		if elementIndex >= 0 && elementIndex < len(tupleArgs) {
			return tupleArgs[elementIndex]
		}
	}
	return utils.GetNumberIndexType(typeChecker, arrayType)
}

func findNodeIndex(nodes []*ast.Node, target *ast.Node) int {
	for i, node := range nodes {
		if node == target {
			return i
		}
	}
	return -1
}

func findParameterIndex(parameters []*ast.ParameterDeclarationNode, target *ast.Node) int {
	for i, parameter := range parameters {
		if parameter != nil && parameter.AsNode() == target {
			return i
		}
	}
	return -1
}

func getDefaultAssignmentStart(node *ast.Node) int {
	if ast.IsParameter(node) {
		parameter := node.AsParameterDeclaration()
		if parameter.Type != nil {
			return parameter.Type.End()
		}
		if parameter.Name() != nil {
			return parameter.Name().End()
		}
	}
	if ast.IsBindingElement(node) {
		if name := node.AsBindingElement().Name(); name != nil {
			return name.End()
		}
	}

	return node.Pos()
}

func isSimpleTypeForMessage(t *checker.Type) bool {
	if t == nil {
		return false
	}

	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		return false
	}

	if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsESSymbolLike) {
		return true
	}

	symbol := checker.Type_symbol(t)
	return symbol != nil && symbol.Name != ""
}

func getPluralAssignmentType(assignmentType string) string {
	switch assignmentType {
	case "property":
		return "properties"
	case "parameter":
		return "parameters"
	default:
		return assignmentType + "s"
	}
}

func getAssignmentTargetNode(node *ast.Node) *ast.Node {
	if ast.IsParameter(node) {
		parameter := node.AsParameterDeclaration()
		if parameter.Type != nil {
			return parameter.Type
		}
		return parameter.Name()
	}

	if ast.IsBindingElement(node) {
		return node.AsBindingElement().Name()
	}

	return nil
}

func hasPropertyInAllBranches(expression *ast.Node, propertyName string) bool {
	expression = ast.SkipParentheses(expression)
	if ast.IsObjectLiteralExpression(expression) {
		for _, property := range expression.AsObjectLiteralExpression().Properties.Nodes {
			if ast.IsSpreadAssignment(property) {
				continue
			}
			if name, ok := getPropertyName(property.Name()); ok && name == propertyName {
				return true
			}
		}
		return false
	}

	if ast.IsConditionalExpression(expression) {
		conditional := expression.AsConditionalExpression()
		return hasPropertyInAllBranches(conditional.WhenTrue, propertyName) && hasPropertyInAllBranches(conditional.WhenFalse, propertyName)
	}

	return false
}

var NoUselessDefaultAssignmentRule = rule.Rule{
	Name: noUselessDefaultAssignmentRuleName,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		var getSourceTypeForPattern func(pattern *ast.Node) *checker.Type
		var getTypeOfBindingElement func(bindingElement *ast.Node) *checker.Type

		getTypeOfBindingElement = func(bindingElement *ast.Node) *checker.Type {
			if bindingElement == nil || !ast.IsBindingElement(bindingElement) {
				return nil
			}

			parentPattern := bindingElement.Parent
			if parentPattern == nil {
				return nil
			}

			if ast.IsObjectBindingPattern(parentPattern) {
				sourceType := getSourceTypeForPattern(parentPattern)
				if sourceType == nil {
					return nil
				}

				bindingElementNode := bindingElement.AsBindingElement()
				propertyNameNode := bindingElementNode.PropertyName
				if propertyNameNode == nil {
					propertyNameNode = bindingElementNode.Name()
				}

				propertyName, ok := getPropertyName(propertyNameNode)
				if !ok {
					return nil
				}

				symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
				if symbol == nil {
					return nil
				}

				if utils.IsSymbolFlagSet(symbol, ast.SymbolFlagsOptional) {
					if parent := parentPattern.Parent; ast.IsVariableDeclaration(parent) && parent.Initializer() != nil {
						if !hasPropertyInAllBranches(parent.Initializer(), propertyName) {
							return nil
						}
					} else {
						return nil
					}
				}

				return ctx.TypeChecker.GetTypeOfSymbolAtLocation(symbol, bindingElement)
			}

			if ast.IsArrayBindingPattern(parentPattern) {
				sourceType := getSourceTypeForPattern(parentPattern)
				if sourceType == nil {
					return nil
				}
				elementIndex := findNodeIndex(parentPattern.AsBindingPattern().Elements.Nodes, bindingElement)
				if elementIndex == -1 {
					return nil
				}
				return getArrayElementType(ctx.TypeChecker, sourceType, elementIndex)
			}

			return nil
		}

		getSourceTypeForPattern = func(pattern *ast.Node) *checker.Type {
			if pattern == nil || pattern.Parent == nil {
				return nil
			}

			parent := pattern.Parent
			if ast.IsVariableDeclaration(parent) && parent.Initializer() != nil {
				return ctx.TypeChecker.GetTypeAtLocation(parent.Initializer())
			}

			if ast.IsParameter(parent) {
				functionNode := parent.Parent
				if functionNode == nil || !ast.IsFunctionLike(functionNode) {
					return nil
				}

				signature := ctx.TypeChecker.GetSignatureFromDeclaration(functionNode)
				if signature == nil {
					return nil
				}

				paramIndex := findParameterIndex(functionNode.Parameters(), parent)
				if paramIndex == -1 {
					return nil
				}

				if signature.ThisParameter() != nil && len(functionNode.Parameters()) > 0 {
					firstParameter := functionNode.Parameters()[0].AsNode()
					if firstParameter != nil && firstParameter.Name() != nil && ast.IsIdentifier(firstParameter.Name()) && firstParameter.Name().AsIdentifier().Text == "this" {
						paramIndex--
					}
				}

				parameters := checker.Signature_parameters(signature)
				if paramIndex < 0 || paramIndex >= len(parameters) {
					return nil
				}

				return checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[paramIndex])
			}

			if ast.IsBindingElement(parent) {
				return getTypeOfBindingElement(parent)
			}

			if ast.IsArrayBindingPattern(parent) {
				arrayType := getSourceTypeForPattern(parent)
				if arrayType == nil {
					return nil
				}
				elementIndex := findNodeIndex(parent.AsBindingPattern().Elements.Nodes, pattern)
				if elementIndex == -1 {
					return nil
				}
				return getArrayElementType(ctx.TypeChecker, arrayType, elementIndex)
			}

			return nil
		}

		buildRemoveDefaultFix := func(node *ast.Node) rule.RuleFix {
			return rule.RuleFixRemoveRange(core.NewTextRange(getDefaultAssignmentStart(node), node.End()))
		}

		reportUselessDefaultAssignment := func(node *ast.Node, assignmentType string, valueType *checker.Type) {
			initializer := node.Initializer()
			if initializer == nil {
				return
			}

			message := buildUselessDefaultAssignmentMessage(assignmentType)
			typeText := ""
			if valueType != nil && isSimpleTypeForMessage(valueType) {
				typeText = ctx.TypeChecker.TypeToString(valueType)
				message = buildUselessDefaultAssignmentWithTypeMessage(assignmentType, typeText)
			}

			diagnosticRange := utils.TrimNodeTextRange(ctx.SourceFile, initializer)
			fixes := []rule.RuleFix{buildRemoveDefaultFix(node)}

			labeledRanges := []rule.RuleLabeledRange{
				{
					Label: "Default value",
					Range: diagnosticRange,
				},
			}

			if typeText != "" {
				if targetNode := getAssignmentTargetNode(node); targetNode != nil {
					labeledRanges = append(labeledRanges, rule.RuleLabeledRange{
						Label: fmt.Sprintf("%s type `%s` is not nullish", assignmentType, typeText),
						Range: utils.TrimNodeTextRange(ctx.SourceFile, targetNode),
					})
				}
			}

			ctx.ReportDiagnostic(rule.RuleDiagnostic{
				Range:         diagnosticRange,
				Message:       message,
				FixesPtr:      &fixes,
				LabeledRanges: labeledRanges,
			})
		}

		reportUselessUndefined := func(node *ast.Node, assignmentType string) {
			initializer := node.Initializer()
			if initializer == nil {
				return
			}
			ctx.ReportNodeWithFixes(initializer, buildUselessUndefinedMessage(getPluralAssignmentType(assignmentType)), func() []rule.RuleFix {
				return []rule.RuleFix{buildRemoveDefaultFix(node)}
			})
		}

		reportPreferOptionalSyntax := func(node *ast.Node) {
			initializer := node.Initializer()
			if initializer == nil {
				return
			}

			ctx.ReportNodeWithFixes(initializer, buildPreferOptionalSyntaxMessage(), func() []rule.RuleFix {
				fixes := []rule.RuleFix{buildRemoveDefaultFix(node)}

				if ast.IsParameter(node) {
					parameter := node.AsParameterDeclaration()
					if name := parameter.Name(); ast.IsIdentifier(name) {
						identifierRange := utils.TrimNodeTextRange(ctx.SourceFile, name)
						fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(identifierRange.End(), identifierRange.End()), "?"))
					}
				}

				return fixes
			})
		}

		checkFunctionExpressionParameter := func(node *ast.Node) {
			parent := node.Parent
			if parent == nil || (!ast.IsArrowFunction(parent) && !ast.IsFunctionExpression(parent)) {
				return
			}

			paramIndex := findParameterIndex(parent.Parameters(), node)
			if paramIndex == -1 {
				return
			}

			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, parent, checker.ContextFlagsNone)
			if contextualType == nil {
				return
			}

			signatures := utils.GetCallSignatures(ctx.TypeChecker, contextualType)
			if len(signatures) == 0 || checker.Signature_declaration(signatures[0]) == parent {
				return
			}

			parameters := checker.Signature_parameters(signatures[0])
			if paramIndex >= len(parameters) {
				return
			}

			paramSymbol := parameters[paramIndex]
			if paramSymbol == nil {
				return
			}

			if paramSymbol.ValueDeclaration != nil && ast.IsParameter(paramSymbol.ValueDeclaration) && paramSymbol.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
				return
			}

			if !utils.IsSymbolFlagSet(paramSymbol, ast.SymbolFlagsOptional) {
				paramType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, paramSymbol)
				if !canBeUndefined(paramType) {
					reportUselessDefaultAssignment(node, "parameter", paramType)
				}
			}
		}

		checkAssignmentPattern := func(node *ast.Node) {
			initializer := node.Initializer()
			if initializer == nil {
				return
			}
			initializer = ast.SkipParentheses(initializer)

			if utils.IsUndefinedIdentifier(initializer) {
				if ast.IsParameter(node) {
					parameter := node.AsParameterDeclaration()
					if parameter.Type != nil && canBeUndefined(checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, parameter.Type)) {
						reportPreferOptionalSyntax(node)
						return
					}
					reportUselessUndefined(node, "parameter")
					return
				}

				reportUselessUndefined(node, "property")
				return
			}

			if ast.IsParameter(node) {
				checkFunctionExpressionParameter(node)
				return
			}

			if !ast.IsBindingElement(node) || node.Parent == nil {
				return
			}

			parent := node.Parent
			if ast.IsObjectBindingPattern(parent) {
				propertyType := getTypeOfBindingElement(node)
				if propertyType != nil && !canBeUndefined(propertyType) {
					reportUselessDefaultAssignment(node, "property", propertyType)
				}
				return
			}

			if ast.IsArrayBindingPattern(parent) {
				sourceType := getSourceTypeForPattern(parent)
				if sourceType == nil || !checker.IsTupleType(sourceType) {
					return
				}

				tupleArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, sourceType)
				elementIndex := findNodeIndex(parent.AsBindingPattern().Elements.Nodes, node)
				if elementIndex < 0 || elementIndex >= len(tupleArgs) {
					return
				}

				if !canBeUndefined(tupleArgs[elementIndex]) {
					reportUselessDefaultAssignment(node, "property", tupleArgs[elementIndex])
				}
			}
		}

		return rule.RuleListeners{
			ast.KindBindingElement: checkAssignmentPattern,
			ast.KindParameter:      checkAssignmentPattern,
		}
	},
}
