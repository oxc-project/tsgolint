package prefer_readonly

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildPreferReadonlyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferReadonly",
		Description: fmt.Sprintf("Member '%s' is never reassigned; mark it as `readonly`.", name),
	}
}

const (
	outsideConstructor        = -1
	directlyInsideConstructor = 0
)

type typeToClassRelation uint8

const (
	typeToClassRelationClassAndInstance typeToClassRelation = iota
	typeToClassRelationClass
	typeToClassRelationInstance
	typeToClassRelationNone
)

type preparedReport struct {
	reportNode *ast.Node
	message    rule.RuleMessage
	fixes      []rule.RuleFix
}

type orderedDeclarationMap struct {
	order []string
	nodes map[string]*ast.Node
}

func newOrderedDeclarationMap() *orderedDeclarationMap {
	return &orderedDeclarationMap{
		order: make([]string, 0, 4),
		nodes: make(map[string]*ast.Node),
	}
}

func (m *orderedDeclarationMap) set(name string, node *ast.Node) {
	if _, ok := m.nodes[name]; !ok {
		m.order = append(m.order, name)
	}
	m.nodes[name] = node
}

func (m *orderedDeclarationMap) valuesWithout(skipped map[string]struct{}) []*ast.Node {
	values := make([]*ast.Node, 0, len(m.order))
	for _, name := range m.order {
		if _, ok := skipped[name]; ok {
			continue
		}
		node, ok := m.nodes[name]
		if !ok || node == nil {
			continue
		}
		values = append(values, node)
	}
	return values
}

type classScope struct {
	typeChecker *checker.Checker
	classType   *checker.Type

	onlyInlineLambdas bool

	constructorScopeDepth int

	memberVariableModifications                map[string]struct{}
	memberVariableWithConstructorModifications map[string]struct{}
	privateModifiableMembers                   *orderedDeclarationMap
	privateModifiableStatics                   *orderedDeclarationMap
	staticVariableModifications                map[string]struct{}
	deferredReports                            []preparedReport
}

func newClassScope(typeChecker *checker.Checker, classNode *ast.Node, onlyInlineLambdas bool) *classScope {
	classType := typeChecker.GetTypeAtLocation(classNode)
	if utils.IsIntersectionType(classType) {
		types := classType.Types()
		if len(types) > 0 {
			classType = types[0]
		}
	}

	scope := &classScope{
		typeChecker: typeChecker,
		classType:   classType,

		onlyInlineLambdas: onlyInlineLambdas,

		constructorScopeDepth: outsideConstructor,

		memberVariableModifications:                make(map[string]struct{}),
		memberVariableWithConstructorModifications: make(map[string]struct{}),
		privateModifiableMembers:                   newOrderedDeclarationMap(),
		privateModifiableStatics:                   newOrderedDeclarationMap(),
		staticVariableModifications:                make(map[string]struct{}),
		deferredReports:                            make([]preparedReport, 0, 2),
	}

	if ast.IsClassLike(classNode) {
		for _, member := range classNode.Members() {
			if ast.IsPropertyDeclaration(member) {
				scope.addDeclaredVariable(member)
			}
		}
	}

	return scope
}

func (scope *classScope) addDeclaredVariable(node *ast.Node) {
	if node == nil {
		return
	}

	nameNode := node.Name()
	if nameNode == nil {
		return
	}

	isPrivate := node.ModifierFlags()&ast.ModifierFlagsPrivate != 0 || nameNode.Kind == ast.KindPrivateIdentifier
	if !isPrivate {
		return
	}

	if node.ModifierFlags()&(ast.ModifierFlagsAccessor|ast.ModifierFlagsReadonly) != 0 {
		return
	}

	if ast.IsComputedPropertyName(nameNode) {
		return
	}

	initializer := node.Initializer()
	if scope.onlyInlineLambdas && (initializer == nil || !ast.IsArrowFunction(initializer)) {
		return
	}

	name := nameNode.Text()
	if name == "" {
		return
	}

	if node.ModifierFlags()&ast.ModifierFlagsStatic != 0 {
		scope.privateModifiableStatics.set(name, node)
		return
	}

	scope.privateModifiableMembers.set(name, node)
}

func (scope *classScope) addVariableModification(node *ast.Node) {
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return
	}

	modifierType := scope.typeChecker.GetTypeAtLocation(node.Expression())
	relation := scope.getTypeToClassRelation(modifierType)

	if relation == typeToClassRelationInstance && scope.constructorScopeDepth == directlyInsideConstructor {
		scope.memberVariableWithConstructorModifications[node.Name().Text()] = struct{}{}
		return
	}

	if relation == typeToClassRelationInstance || relation == typeToClassRelationClassAndInstance {
		scope.memberVariableModifications[node.Name().Text()] = struct{}{}
	}
	if relation == typeToClassRelationClass || relation == typeToClassRelationClassAndInstance {
		scope.staticVariableModifications[node.Name().Text()] = struct{}{}
	}
}

func (scope *classScope) enterConstructor(node *ast.Node) {
	scope.constructorScopeDepth = directlyInsideConstructor

	for _, parameter := range node.Parameters() {
		if parameter.ModifierFlags()&ast.ModifierFlagsPrivate != 0 {
			scope.addDeclaredVariable(parameter)
		}
	}
}

func (scope *classScope) enterNonConstructor() {
	if scope.constructorScopeDepth != outsideConstructor {
		scope.constructorScopeDepth += 1
	}
}

func (scope *classScope) exitConstructor() {
	scope.constructorScopeDepth = outsideConstructor
}

func (scope *classScope) exitNonConstructor() {
	if scope.constructorScopeDepth != outsideConstructor {
		scope.constructorScopeDepth -= 1
	}
}

func (scope *classScope) finalizeUnmodifiedPrivateNonReadonlys() []*ast.Node {
	result := scope.privateModifiableMembers.valuesWithout(scope.memberVariableModifications)
	result = append(result, scope.privateModifiableStatics.valuesWithout(scope.staticVariableModifications)...)
	return result
}

func (scope *classScope) memberHasConstructorModifications(name string) bool {
	_, ok := scope.memberVariableWithConstructorModifications[name]
	return ok
}

func (scope *classScope) getTypeToClassRelation(t *checker.Type) typeToClassRelation {
	if t == nil {
		return typeToClassRelationNone
	}

	if utils.IsIntersectionType(t) {
		result := typeToClassRelationNone
		for _, part := range t.Types() {
			subTypeResult := scope.getTypeToClassRelation(part)
			switch subTypeResult {
			case typeToClassRelationClass:
				if result == typeToClassRelationInstance {
					return typeToClassRelationClassAndInstance
				}
				result = typeToClassRelationClass
			case typeToClassRelationInstance:
				if result == typeToClassRelationClass {
					return typeToClassRelationClassAndInstance
				}
				result = typeToClassRelationInstance
			}
		}
		return result
	}

	if utils.IsUnionType(t) {
		types := t.Types()
		if len(types) == 0 {
			return typeToClassRelationNone
		}
		// Any union of class/instance and something else cannot access private members,
		// so we assume this union contains only classes or class instances because
		// otherwise TypeScript would report an error at the access site.
		return scope.getTypeToClassRelation(types[0])
	}

	if checker.Type_symbol(t) == nil || !typeIsOrHasBaseType(scope.typeChecker, t, scope.classType) {
		return typeToClassRelationNone
	}

	typeIsClass := utils.IsObjectType(t) && checker.Type_objectFlags(t)&checker.ObjectFlagsAnonymous != 0
	if typeIsClass {
		return typeToClassRelationClass
	}

	return typeToClassRelationInstance
}

func typeIsOrHasBaseType(typeChecker *checker.Checker, t *checker.Type, baseType *checker.Type) bool {
	if t == nil || baseType == nil {
		return false
	}

	baseSymbol := checker.Type_symbol(baseType)
	if baseSymbol == nil {
		return false
	}

	queue := []*checker.Type{t}
	seen := map[*checker.Type]struct{}{}

	for len(queue) > 0 {
		current := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if current == nil {
			continue
		}

		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}

		if checker.Type_symbol(current) == baseSymbol {
			return true
		}

		if !utils.IsObjectType(current) || checker.Type_objectFlags(current)&checker.ObjectFlagsClassOrInterface == 0 {
			continue
		}

		for _, base := range checker.Checker_getBaseTypes(typeChecker, current) {
			queue = append(queue, base)
		}
	}

	return false
}

func isDestructuringAssignment(node *ast.Node) bool {
	current := node.Parent

	for current != nil {
		parent := current.Parent
		if parent == nil {
			break
		}

		if ast.IsObjectLiteralExpression(parent) || ast.IsArrayLiteralExpression(parent) || ast.IsSpreadAssignment(parent) ||
			(ast.IsSpreadElement(parent) && parent.Parent != nil && ast.IsArrayLiteralExpression(parent.Parent)) {
			current = parent
			continue
		}

		if ast.IsBinaryExpression(parent) && !ast.IsPropertyAccessExpression(current) {
			binExpr := parent.AsBinaryExpression()
			return binExpr.Left == current && binExpr.OperatorToken.Kind == ast.KindEqualsToken
		}

		break
	}

	return false
}

func getTypeAnnotationForViolatingNode(
	ctx rule.RuleContext,
	node *ast.Node,
	violatingType *checker.Type,
	initializerType *checker.Type,
) string {
	annotation := ctx.TypeChecker.TypeToString(violatingType)
	if annotation == "" {
		return ""
	}

	if checker.Type_flags(initializerType)&checker.TypeFlagsEnumLiteral == 0 {
		return annotation
	}

	symbol := ctx.TypeChecker.ResolveName(annotation, node, ast.SymbolFlagsType, false)
	if symbol == nil {
		return ""
	}

	valueSymbol := ctx.TypeChecker.ResolveName(annotation, node, ast.SymbolFlagsValue, false)
	if valueSymbol != nil {
		valueSymbol = checker.SkipAlias(valueSymbol, ctx.TypeChecker)
	}

	symbol = checker.SkipAlias(symbol, ctx.TypeChecker)
	if valueSymbol != nil && valueSymbol != symbol {
		return ""
	}

	definitionType := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)
	if definitionType != violatingType {
		return ""
	}

	return annotation
}

func getCustomReportNode(sourceFile *ast.SourceFile, declarationNode *ast.Node, nameNode *ast.Node) *ast.Node {
	declarationRange := utils.TrimNodeTextRange(sourceFile, declarationNode)
	nameRange := utils.TrimNodeTextRange(sourceFile, nameNode)
	return &ast.Node{
		Kind: declarationNode.Kind,
		Loc:  core.NewTextRange(declarationRange.Pos(), nameRange.End()),
	}
}

var PreferReadonlyRule = rule.Rule{
	Name: "prefer-readonly",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferReadonlyOptions](options, "prefer-readonly")

		classScopeStack := make([]*classScope, 0, 1)

		popClassScope := func() *classScope {
			if len(classScopeStack) == 0 {
				return nil
			}

			finalizedScope := classScopeStack[len(classScopeStack)-1]
			classScopeStack = classScopeStack[:len(classScopeStack)-1]
			return finalizedScope
		}

		buildReports := func(finalizedScope *classScope) []preparedReport {
			if finalizedScope == nil {
				return nil
			}

			reports := make([]preparedReport, 0, 2)

			for _, violatingNode := range finalizedScope.finalizeUnmodifiedPrivateNonReadonlys() {
				if violatingNode == nil || violatingNode.Name() == nil {
					continue
				}

				nameNode := violatingNode.Name()
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
				reportNode := getCustomReportNode(ctx.SourceFile, violatingNode, nameNode)

				fixes := []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, nameNode, "readonly ")}

				if ast.IsPropertyDeclaration(violatingNode) {
					property := violatingNode.AsPropertyDeclaration()
					if property.Type == nil && property.Initializer != nil && ast.IsIdentifier(nameNode) && finalizedScope.memberHasConstructorModifications(nameNode.Text()) {
						violatingType := ctx.TypeChecker.GetTypeAtLocation(violatingNode)
						initializerType := ctx.TypeChecker.GetTypeAtLocation(property.Initializer)

						if violatingType != nil && initializerType != nil && violatingType != initializerType && checker.Type_flags(initializerType)&checker.TypeFlagsLiteral != 0 {
							typeAnnotation := getTypeAnnotationForViolatingNode(ctx, violatingNode, violatingType, initializerType)
							if typeAnnotation != "" {
								fixes = append(fixes, rule.RuleFixInsertAfter(nameNode, ": "+typeAnnotation))
							}
						}
					}
				}

				reports = append(reports, preparedReport{
					reportNode: reportNode,
					message:    buildPreferReadonlyMessage(nameText),
					fixes:      fixes,
				})
			}

			return reports
		}

		flushFinalizedScope := func(finalizedScope *classScope) {
			if finalizedScope == nil {
				return
			}

			reports := buildReports(finalizedScope)
			reports = append(reports, finalizedScope.deferredReports...)

			if len(classScopeStack) > 0 {
				parent := classScopeStack[len(classScopeStack)-1]
				parent.deferredReports = append(parent.deferredReports, reports...)
				return
			}

			for _, report := range reports {
				ctx.ReportNodeWithFixes(report.reportNode, report.message, func() []rule.RuleFix { return report.fixes })
			}
		}

		handleParentBinaryExpression := func(node *ast.Node, parent *ast.Node, scope *classScope) {
			if parent.AsBinaryExpression().Left == node && ast.IsAssignmentExpression(parent, false) {
				scope.addVariableModification(node)
			}
		}

		handleParentPostfixOrPrefixUnaryExpression := func(node *ast.Node, scope *classScope) {
			if node.Kind == ast.KindPostfixUnaryExpression {
				if node.AsPostfixUnaryExpression().Operator == ast.KindPlusPlusToken || node.AsPostfixUnaryExpression().Operator == ast.KindMinusMinusToken {
					operand := node.AsPostfixUnaryExpression().Operand
					if ast.IsPropertyAccessExpression(operand) {
						scope.addVariableModification(operand)
					}
				}
				return
			}

			if ast.IsPrefixUnaryExpression(node) {
				if node.AsPrefixUnaryExpression().Operator == ast.KindPlusPlusToken || node.AsPrefixUnaryExpression().Operator == ast.KindMinusMinusToken {
					operand := node.AsPrefixUnaryExpression().Operand
					if ast.IsPropertyAccessExpression(operand) {
						scope.addVariableModification(operand)
					}
				}
			}
		}

		handlePropertyAccessExpression := func(node *ast.Node, scope *classScope) {
			parent := node.Parent

			if ast.IsBinaryExpression(parent) {
				handleParentBinaryExpression(node, parent, scope)
				return
			}

			if parent.Kind == ast.KindDeleteExpression || isDestructuringAssignment(node) {
				scope.addVariableModification(node)
				return
			}

			if parent.Kind == ast.KindPostfixUnaryExpression || ast.IsPrefixUnaryExpression(parent) {
				handleParentPostfixOrPrefixUnaryExpression(parent, scope)
			}
		}

		enterNonConstructorScope := func() {
			if len(classScopeStack) == 0 {
				return
			}
			classScopeStack[len(classScopeStack)-1].enterNonConstructor()
		}

		exitNonConstructorScope := func() {
			if len(classScopeStack) == 0 {
				return
			}
			classScopeStack[len(classScopeStack)-1].exitNonConstructor()
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classScopeStack = append(classScopeStack, newClassScope(ctx.TypeChecker, node, opts.OnlyInlineLambdas))
			},
			ast.KindClassExpression: func(node *ast.Node) {
				classScopeStack = append(classScopeStack, newClassScope(ctx.TypeChecker, node, opts.OnlyInlineLambdas))
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(_ *ast.Node) {
				flushFinalizedScope(popClassScope())
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(_ *ast.Node) {
				flushFinalizedScope(popClassScope())
			},

			ast.KindConstructor: func(node *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				classScopeStack[len(classScopeStack)-1].enterConstructor(node)
			},
			rule.ListenerOnExit(ast.KindConstructor): func(_ *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				classScopeStack[len(classScopeStack)-1].exitConstructor()
			},

			ast.KindArrowFunction: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(_ *ast.Node) {
				exitNonConstructorScope()
			},
			ast.KindFunctionDeclaration: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(_ *ast.Node) {
				exitNonConstructorScope()
			},
			ast.KindFunctionExpression: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(_ *ast.Node) {
				exitNonConstructorScope()
			},
			ast.KindMethodDeclaration: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(_ *ast.Node) {
				exitNonConstructorScope()
			},
			ast.KindGetAccessor: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindGetAccessor): func(_ *ast.Node) {
				exitNonConstructorScope()
			},
			ast.KindSetAccessor: func(_ *ast.Node) {
				enterNonConstructorScope()
			},
			rule.ListenerOnExit(ast.KindSetAccessor): func(_ *ast.Node) {
				exitNonConstructorScope()
			},

			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				handlePropertyAccessExpression(node, classScopeStack[len(classScopeStack)-1])
			},
		}
	},
}
