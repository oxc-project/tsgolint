package no_unnecessary_type_parameters

import (
	"fmt"
	"math"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildSoleMessage(typeParameterRange core.TextRange, typeParameterReference core.TextRange, name string, uses string, descriptor string) rule.RuleDiagnostic {

	return rule.RuleDiagnostic{
		Message: rule.RuleMessage{
			Id:          "sole",
			Description: fmt.Sprintf("Type parameter %s is %s in the %s signature.", name, uses, descriptor),
		},
		Range: typeParameterRange,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("This is the only usage of type parameter %s in the signature.", name),
				Range: typeParameterReference,
			},
		},
	}
}

func buildReplaceUsagesWithConstraintMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceUsagesWithConstraint",
		Description: "Replace all usages of type parameter with its constraint.",
	}
}

func findTypeParameterIndex(typeParameters []*ast.Node, target *ast.Node) int {
	for i, typeParameter := range typeParameters {
		if typeParameter == target {
			return i
		}
	}
	return -1
}

func isComplexConstraint(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindUnionType, ast.KindIntersectionType, ast.KindConditionalType:
		return true
	default:
		return false
	}
}

func hasMatchingAncestorType(reference *ast.Node) bool {
	if reference == nil || reference.Parent == nil {
		return false
	}

	grandparent := reference.Parent.Parent
	if grandparent == nil {
		return false
	}

	switch grandparent.Kind {
	case ast.KindArrayType, ast.KindIndexedAccessType, ast.KindIntersectionType, ast.KindUnionType:
		return true
	default:
		return false
	}
}

func symbolFromTypeParameter(ctx rule.RuleContext, typeParameter *ast.Node) *ast.Symbol {
	if typeParameter == nil || !ast.IsTypeParameterDeclaration(typeParameter) {
		return nil
	}

	return ctx.TypeChecker.GetSymbolAtLocation(typeParameter.AsTypeParameter().Name())
}

func collectTypeParameterReferenceNodes(ctx rule.RuleContext, node *ast.Node, symbol *ast.Symbol, declarationName *ast.Node) []*ast.Node {
	if node == nil || symbol == nil {
		return nil
	}

	references := make([]*ast.Node, 0, 8)

	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil {
			return
		}

		if ast.IsIdentifier(current) {
			if current != declarationName {
				if current.Parent != nil && ast.IsTypeNode(current.Parent) {
					if currentSymbol := ctx.TypeChecker.GetSymbolAtLocation(current); currentSymbol == symbol {
						references = append(references, current)
					}
				}
			}
		}

		ast.ForEachChildAndJSDoc(current, ctx.SourceFile, func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(node)

	sort.Slice(references, func(i int, j int) bool {
		return references[i].Pos() < references[j].Pos()
	})

	return references
}

func getTypeParameterConstraintText(ctx rule.RuleContext, typeParameter *ast.Node) (constraintText string, constraintNode *ast.Node) {
	constraintNode = typeParameter.AsTypeParameter().Constraint
	if constraintNode == nil || constraintNode.Kind == ast.KindAnyKeyword {
		return "unknown", constraintNode
	}

	return scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, constraintNode, false /* includeTrivia */), constraintNode
}

func getTypeParameterListRemovalRange(ctx rule.RuleContext, typeParameters []*ast.Node, targetTypeParameter *ast.Node) core.TextRange {
	index := findTypeParameterIndex(typeParameters, targetTypeParameter)
	if index < 0 {
		return targetTypeParameter.Loc
	}

	if len(typeParameters) == 1 {
		firstTypeParameter := typeParameters[0]
		start := firstTypeParameter.Pos() - 1
		if start < 0 {
			start = firstTypeParameter.Pos()
		}

		text := ctx.SourceFile.Text()
		end := firstTypeParameter.End()
		for end < len(text) && text[end] != '>' {
			end++
		}
		if end < len(text) && text[end] == '>' {
			end++
		}
		if end <= start {
			end = firstTypeParameter.End()
		}

		return core.NewTextRange(start, end)
	}

	if index == 0 {
		end := typeParameters[index+1].Pos()
		tokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, end)
		if tokenRange.Pos() > targetTypeParameter.Pos() {
			end = tokenRange.Pos()
		}
		return core.NewTextRange(targetTypeParameter.Pos(), end)
	}

	text := ctx.SourceFile.Text()
	start := typeParameters[index-1].End()
	for i := typeParameters[index-1].End(); i < targetTypeParameter.Pos() && i < len(text); i++ {
		if text[i] == ',' {
			start = i
			break
		}
	}

	return core.NewTextRange(start, targetTypeParameter.End())
}

func classMembers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().Members.Nodes
	case ast.KindClassExpression:
		return node.AsClassExpression().Members.Nodes
	default:
		return nil
	}
}

func countTypeParameterUsage(ctx rule.RuleContext, node *ast.Node) map[*ast.Symbol]int {
	counts := make(map[*ast.Symbol]int)

	if ast.IsClassLike(node) {
		for _, typeParameter := range node.TypeParameters() {
			collectTypeParameterUsageCounts(ctx, typeParameter, counts, true)
		}
		if heritageClauses := utils.GetHeritageClauses(node); heritageClauses != nil {
			for _, heritageClause := range heritageClauses.Nodes {
				for _, heritageType := range heritageClause.AsHeritageClause().Types.Nodes {
					collectTypeParameterUsageCounts(ctx, heritageType, counts, true)
				}
			}
		}
		for _, member := range classMembers(node) {
			collectTypeParameterUsageCounts(ctx, member, counts, true)
		}
	} else {
		collectTypeParameterUsageCounts(ctx, node, counts, false)
	}

	return counts
}

func getStartOfBody(node *ast.Node) int {
	if node == nil {
		return math.MaxInt
	}

	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().Members.Pos()
	case ast.KindClassExpression:
		return node.AsClassExpression().Members.Pos()
	case ast.KindFunctionDeclaration:
		if node.AsFunctionDeclaration().Body != nil {
			return node.AsFunctionDeclaration().Body.Pos()
		}
		if node.AsFunctionDeclaration().Type != nil {
			return node.AsFunctionDeclaration().Type.End()
		}
	case ast.KindFunctionExpression:
		if node.AsFunctionExpression().Body != nil {
			return node.AsFunctionExpression().Body.Pos()
		}
		if node.AsFunctionExpression().Type != nil {
			return node.AsFunctionExpression().Type.End()
		}
	case ast.KindArrowFunction:
		if node.AsArrowFunction().Body != nil {
			if ast.IsFunctionLike(node.AsArrowFunction().Body) {
				// Returned function signatures contribute to the inferred return type.
				return math.MaxInt
			}
			return node.AsArrowFunction().Body.Pos()
		}
		if node.AsArrowFunction().Type != nil {
			return node.AsArrowFunction().Type.End()
		}
	case ast.KindMethodDeclaration:
		if node.AsMethodDeclaration().Body != nil {
			return node.AsMethodDeclaration().Body.Pos()
		}
		if node.AsMethodDeclaration().Type != nil {
			return node.AsMethodDeclaration().Type.End()
		}
	case ast.KindCallSignature:
		if node.AsCallSignatureDeclaration().Type != nil {
			return node.AsCallSignatureDeclaration().Type.End()
		}
	case ast.KindConstructSignature:
		if node.AsConstructSignatureDeclaration().Type != nil {
			return node.AsConstructSignatureDeclaration().Type.End()
		}
	case ast.KindConstructorType:
		if node.AsConstructorTypeNode().Type != nil {
			return node.AsConstructorTypeNode().Type.End()
		}
	case ast.KindFunctionType:
		if node.AsFunctionTypeNode().Type != nil {
			return node.AsFunctionTypeNode().Type.End()
		}
	case ast.KindMethodSignature:
		if node.AsMethodSignatureDeclaration().Type != nil {
			return node.AsMethodSignatureDeclaration().Type.End()
		}
	}

	return math.MaxInt
}

func isTypeParameterRepeatedInAST(typeParameter *ast.Node, references []*ast.Node, startOfBody int) bool {
	total := 0

	for _, reference := range references {
		// References inside the type parameter declaration itself don't count.
		if reference.Pos() < typeParameter.End() && reference.End() > typeParameter.Pos() {
			continue
		}

		// References in the body don't count for fast-path usage checks.
		if reference.Pos() > startOfBody {
			continue
		}

		total++
		if total >= 2 {
			return true
		}
	}

	return false
}

func collectTypeParameterUsageCounts(
	ctx rule.RuleContext,
	node *ast.Node,
	foundIdentifierUsages map[*ast.Symbol]int,
	fromClass bool,
) {
	typeUsages := make(map[*checker.Type]int)
	visitedConstraints := make(map[*ast.Node]bool)
	visitedDefault := false
	functionLikeType := false

	incrementIdentifierCount := func(symbol *ast.Symbol, assumeMultipleUses bool) {
		if symbol == nil {
			return
		}
		value := 1
		if assumeMultipleUses {
			value = 2
		}
		foundIdentifierUsages[symbol] = foundIdentifierUsages[symbol] + value
	}

	incrementTypeUsages := func(t *checker.Type) int {
		count := typeUsages[t] + 1
		typeUsages[t] = count
		return count
	}

	var visitType func(typeNode *checker.Type, assumeMultipleUses bool, isReturnType bool)
	var visitTypesList func(types []*checker.Type, assumeMultipleUses bool)
	var visitSignature func(signature *checker.Signature)
	var visitSymbolsList func(symbols []*ast.Symbol, assumeMultipleUses bool)

	visitTypesList = func(types []*checker.Type, assumeMultipleUses bool) {
		for _, typeNode := range types {
			visitType(typeNode, assumeMultipleUses, false)
		}
	}

	visitSymbolsList = func(symbols []*ast.Symbol, assumeMultipleUses bool) {
		for _, symbol := range symbols {
			visitType(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol), assumeMultipleUses, false)
		}
	}

	visitSignature = func(signature *checker.Signature) {
		if signature == nil {
			return
		}

		if thisParameter := signature.ThisParameter(); thisParameter != nil {
			visitType(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, thisParameter), false, false)
		}

		for _, parameter := range signature.Parameters() {
			visitType(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameter), false, false)
		}

		for _, typeParameter := range signature.TypeParameters() {
			visitType(typeParameter, false, false)
		}

		returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signature)
		if typePredicate := ctx.TypeChecker.GetTypePredicateOfSignature(signature); typePredicate != nil {
			if predicateType := checker.TypePredicate_t(typePredicate); predicateType != nil {
				returnType = predicateType
			}
		}

		visitType(returnType, false, true)
	}

	visitType = func(typeNode *checker.Type, assumeMultipleUses bool, isReturnType bool) {
		if typeNode == nil || incrementTypeUsages(typeNode) > 9 {
			return
		}

		if utils.IsTypeParameter(typeNode) {
			typeParameterSymbol := checker.Type_symbol(typeNode)
			if typeParameterSymbol != nil && len(typeParameterSymbol.Declarations) != 0 {
				declaration := typeParameterSymbol.Declarations[0]
				if ast.IsTypeParameterDeclaration(declaration) {
					incrementIdentifierCount(typeParameterSymbol, assumeMultipleUses)

					typeParameterDeclaration := declaration.AsTypeParameter()
					if typeParameterDeclaration.Constraint != nil && !visitedConstraints[typeParameterDeclaration.Constraint] {
						visitedConstraints[typeParameterDeclaration.Constraint] = true
						visitType(ctx.TypeChecker.GetTypeAtLocation(typeParameterDeclaration.Constraint), false, false)
					}

					if typeParameterDeclaration.DefaultType != nil && !visitedDefault {
						visitedDefault = true
						visitType(ctx.TypeChecker.GetTypeAtLocation(typeParameterDeclaration.DefaultType), false, false)
					}
				}
			}
			return
		}

		if alias := checker.Type_alias(typeNode); alias != nil {
			if aliasTypeArguments := alias.TypeArguments(); len(aliasTypeArguments) != 0 {
				visitTypesList(aliasTypeArguments, true)
				return
			}
		}

		if utils.IsUnionType(typeNode) || utils.IsIntersectionType(typeNode) {
			visitTypesList(typeNode.Types(), assumeMultipleUses)
			return
		}

		if checker.Type_flags(typeNode)&checker.TypeFlagsIndexedAccess != 0 {
			indexedAccessType := typeNode.AsIndexedAccessType()
			visitType(checker.IndexedAccessType_objectType(indexedAccessType), assumeMultipleUses, false)
			visitType(checker.IndexedAccessType_indexType(indexedAccessType), assumeMultipleUses, false)
			return
		}

		if checker.Type_flags(typeNode)&checker.TypeFlagsObject != 0 && checker.Type_objectFlags(typeNode)&checker.ObjectFlagsReference != 0 {
			typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, typeNode)
			if len(typeArguments) != 0 {
				target := typeNode.Target()
				for _, typeArgument := range typeArguments {
					thisAssumeMultipleUses := fromClass || assumeMultipleUses
					if checker.IsTupleType(target) {
						thisAssumeMultipleUses = thisAssumeMultipleUses || (isReturnType && !checker.TupleType_readonly(target.AsTupleType()))
					} else if checker.Checker_isArrayType(ctx.TypeChecker, target) {
						symbolName := ""
						if typeSymbol := checker.Type_symbol(typeNode); typeSymbol != nil {
							symbolName = typeSymbol.Name
						}
						thisAssumeMultipleUses = thisAssumeMultipleUses || (isReturnType && symbolName == "Array")
					} else {
						thisAssumeMultipleUses = true
					}
					visitType(typeArgument, thisAssumeMultipleUses, isReturnType)
				}
				return
			}
		}

		if checker.Type_flags(typeNode)&checker.TypeFlagsTemplateLiteral != 0 {
			visitTypesList(typeNode.Types(), assumeMultipleUses)
			return
		}

		if checker.Type_flags(typeNode)&checker.TypeFlagsConditional != 0 {
			conditionalType := typeNode.AsConditionalType()
			visitType(checker.ConditionalType_checkType(conditionalType), assumeMultipleUses, false)
			visitType(checker.ConditionalType_extendsType(conditionalType), assumeMultipleUses, false)
			return
		}

		if utils.IsObjectType(typeNode) {
			properties := checker.Checker_getPropertiesOfType(ctx.TypeChecker, typeNode)
			visitSymbolsList(properties, false)

			if checker.Type_objectFlags(typeNode)&checker.ObjectFlagsMapped != 0 {
				mappedType := typeNode.AsMappedType()
				visitType(checker.MappedType_typeParameter(mappedType), false, false)
				if len(properties) == 0 {
					templateType := checker.MappedType_templateType(mappedType)
					if templateType != nil {
						visitType(templateType, false, false)
					} else {
						visitType(checker.MappedType_constraintType(mappedType), false, false)
					}
				}
			}

			visitType(ctx.TypeChecker.GetNumberIndexType(typeNode), true, false)
			visitType(ctx.TypeChecker.GetStringIndexType(typeNode), true, false)

			for _, signature := range ctx.TypeChecker.GetCallSignatures(typeNode) {
				functionLikeType = true
				visitSignature(signature)
			}
			for _, signature := range ctx.TypeChecker.GetConstructSignatures(typeNode) {
				functionLikeType = true
				visitSignature(signature)
			}
			return
		}

		if checker.Type_flags(typeNode)&checker.TypeFlagsIndex != 0 ||
			checker.Type_flags(typeNode)&checker.TypeFlagsStringMapping != 0 {
			visitType(typeNode.Target(), assumeMultipleUses, false)
			return
		}

	}

	if ast.IsCallSignatureDeclaration(node) || ast.IsConstructorDeclaration(node) {
		functionLikeType = true
		visitSignature(ctx.TypeChecker.GetSignatureFromDeclaration(node))
	}

	if !functionLikeType {
		visitType(ctx.TypeChecker.GetTypeAtLocation(node), false, false)
	}
}

func checkNoUnnecessaryTypeParametersNode(ctx rule.RuleContext, node *ast.Node, descriptor string) {
	typeParameters := node.TypeParameters()
	if len(typeParameters) == 0 {
		return
	}

	identifierCounts := countTypeParameterUsage(ctx, node)
	startOfBody := getStartOfBody(node)

	for _, typeParameter := range typeParameters {
		typeParameterSymbol := symbolFromTypeParameter(ctx, typeParameter)
		if typeParameterSymbol == nil {
			continue
		}

		typeParameterNameNode := typeParameter.AsTypeParameter().Name()
		references := collectTypeParameterReferenceNodes(ctx, node, typeParameterSymbol, typeParameterNameNode)
		if isTypeParameterRepeatedInAST(typeParameter, references, startOfBody) {
			continue
		}

		identifierCount, ok := identifierCounts[typeParameterSymbol]
		if !ok || identifierCount > 2 {
			continue
		}

		uses := "used only once"
		if identifierCount == 1 {
			uses = "never used"
		}

		typeParameterName := typeParameterNameNode.Text()
		constraintText, constraintNode := getTypeParameterConstraintText(ctx, typeParameter)
		typeParameterReferenceRange := utils.TrimNodeTextRange(ctx.SourceFile, typeParameterNameNode)
		if len(references) > 0 {
			typeParameterReferenceRange = utils.TrimNodeTextRange(ctx.SourceFile, references[0])
		}

		ctx.ReportDiagnosticWithSuggestions(
			buildSoleMessage(
				utils.TrimNodeTextRange(ctx.SourceFile, typeParameter),
				typeParameterReferenceRange,
				typeParameterName,
				uses,
				descriptor,
			),
			func() []rule.RuleSuggestion {
				removalRange := getTypeParameterListRemovalRange(ctx, typeParameters, typeParameter)
				fixes := make([]rule.RuleFix, 0, len(references)+1)

				for _, reference := range references {
					referenceRange := utils.TrimNodeTextRange(ctx.SourceFile, reference)
					if referenceRange.Pos() < removalRange.End() && referenceRange.End() > removalRange.Pos() {
						continue
					}
					replacement := constraintText
					if isComplexConstraint(constraintNode) && hasMatchingAncestorType(reference) {
						replacement = "(" + replacement + ")"
					}
					fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, reference, replacement))
				}

				fixes = append(fixes, rule.RuleFixRemoveRange(removalRange))

				return []rule.RuleSuggestion{{
					Message:  buildReplaceUsagesWithConstraintMessage(),
					FixesArr: fixes,
				}}
			})
	}
}

var NoUnnecessaryTypeParametersRule = rule.Rule{
	Name: "no-unnecessary-type-parameters",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		_ = options

		isTypeParameterNode := func(node *ast.Node) bool {
			return node != nil && len(node.TypeParameters()) > 0
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindCallSignature: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindConstructSignature: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindConstructorType: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindFunctionType: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindMethodSignature: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "function")
				}
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "class")
				}
			},
			ast.KindClassExpression: func(node *ast.Node) {
				if isTypeParameterNode(node) {
					checkNoUnnecessaryTypeParametersNode(ctx, node, "class")
				}
			},
		}
	},
}
