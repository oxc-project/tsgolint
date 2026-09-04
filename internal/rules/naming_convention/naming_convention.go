package naming_convention

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var NamingConventionRule = rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		validators := createValidators(parseOptions(options))

		r := &namingConventionRunner{
			ctx:        ctx,
			validators: validators,
		}

		return rule.RuleListeners{
			// #region function
			ast.KindFunctionDeclaration: r.handleFunction,
			ast.KindFunctionExpression:  r.handleFunction,
			// #endregion function

			// #region import
			ast.KindImportClause:    r.handleImportClause,
			ast.KindNamespaceImport: r.handleNamespaceImport,
			ast.KindImportSpecifier: r.handleImportSpecifier,
			// #endregion import

			// #region variable
			ast.KindVariableDeclaration: r.handleVariableDeclaration,
			// #endregion variable

			// #region parameter, parameterProperty
			ast.KindParameter: r.handleParameter,
			// #endregion parameter, parameterProperty

			// #region class members
			ast.KindPropertyDeclaration: r.handlePropertyDeclaration,
			ast.KindMethodDeclaration:   r.handleMethodDeclaration,
			ast.KindGetAccessor:         r.handleAccessor,
			ast.KindSetAccessor:         r.handleAccessor,
			// #endregion class members

			// #region object literal members
			ast.KindPropertyAssignment:          r.handlePropertyAssignment,
			ast.KindShorthandPropertyAssignment: r.handleShorthandPropertyAssignment,
			// #endregion object literal members

			// #region type members
			ast.KindMethodSignature:   r.handleMethodSignature,
			ast.KindPropertySignature: r.handlePropertySignature,
			// #endregion type members

			// #region typeLike
			ast.KindClassDeclaration:     r.handleClass,
			ast.KindClassExpression:      r.handleClass,
			ast.KindEnumDeclaration:      r.handleEnum,
			ast.KindEnumMember:           r.handleEnumMember,
			ast.KindInterfaceDeclaration: r.handleInterface,
			ast.KindTypeAliasDeclaration: r.handleTypeAlias,
			ast.KindTypeParameter:        r.handleTypeParameter,
			// #endregion typeLike
		}
	},
}

type namingConventionRunner struct {
	ctx        rule.RuleContext
	validators map[selectorFlags]*validator

	// lazily computed reference information, see collectReferences
	referencesCollected bool
	// positions of every reference to the symbol within the file
	references map[*ast.Symbol][]int
	// symbols exported via `export { x }` or `export default x`
	exportedSymbols map[*ast.Symbol]struct{}
	// whether exportedSymbols has been computed
	exportedSymbolsCollected bool
}

// validate runs the validator for the given selector on the given name node.
func (r *namingConventionRunner) validate(selector selectorFlags, nameNode *ast.Node, modifiers modifierFlags) {
	r.validateWithRange(selector, nameNode, utils.TrimNodeTextRange(r.ctx.SourceFile, nameNode), modifiers)
}

func (r *namingConventionRunner) validateWithRange(selector selectorFlags, nameNode *ast.Node, reportRange core.TextRange, modifiers modifierFlags) {
	v := r.validators[selector]
	if v == nil || len(v.configs) == 0 {
		return
	}
	name, ok := getNameText(nameNode)
	if !ok {
		return
	}
	v.validate(r.ctx, nameNode, reportRange, name, modifiers)
}

// getDeclarationNameRange returns the range typescript-estree assigns to the
// identifier of a variable declaration or parameter: the identifier's
// `typeAnnotation` (and `?`) are part of the identifier node, so the range
// extends to the end of the type annotation.
func (r *namingConventionRunner) getDeclarationNameRange(declaration *ast.Node, id *ast.Node) core.TextRange {
	textRange := utils.TrimNodeTextRange(r.ctx.SourceFile, id)
	if id.Parent != declaration {
		// identifiers nested in binding patterns don't carry the type annotation
		return textRange
	}
	end := textRange.End()
	if typeNode := declaration.Type(); typeNode != nil {
		end = max(end, typeNode.End())
	}
	if declaration.Kind == ast.KindParameter {
		if questionToken := declaration.AsParameterDeclaration().QuestionToken; questionToken != nil {
			end = max(end, questionToken.End())
		}
	}
	return textRange.WithEnd(end)
}

// hasUnusedModifierConfigured reports whether any config for the given
// selector requires the `unused` modifier. It is used to avoid computing the
// (expensive) reference information when no config cares about it.
func (r *namingConventionRunner) hasUnusedModifierConfigured(selector selectorFlags) bool {
	v := r.validators[selector]
	if v == nil {
		return false
	}
	for _, config := range v.configs {
		for _, m := range config.modifiers {
			if m == modifierUnused {
				return true
			}
		}
	}
	return false
}

// getNameText returns the name of an identifier, private identifier or literal
// name node, the same way typescript-eslint does (`node.name` for identifiers,
// `${node.value}` for literals).
func getNameText(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text(), true
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(node.Text(), "#"), true
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	}
	return "", false
}

// isNonComputedName reports whether the member name can be validated. Computed
// names are never validated, except for computed string literal names which
// typescript-estree represents as plain literal keys.
func getMemberNameNode(name *ast.Node) *ast.Node {
	if name == nil {
		return nil
	}
	if name.Kind == ast.KindComputedPropertyName {
		return nil
	}
	if _, ok := getNameText(name); !ok {
		return nil
	}
	return name
}

func requiresQuoting(name string) bool {
	if name == "" {
		return true
	}
	return !scanner.IsIdentifierText(name, core.LanguageVariantStandard)
}

// #region modifiers

// getMemberModifiers mirrors typescript-eslint's `getMemberModifiers`.
func getMemberModifiers(node *ast.Node) modifierFlags {
	var modifiers modifierFlags
	flags := node.ModifierFlags()

	name := node.Name()
	switch {
	case name != nil && name.Kind == ast.KindPrivateIdentifier:
		modifiers |= modifierHashPrivate
	case flags&ast.ModifierFlagsPrivate != 0:
		modifiers |= modifierPrivate
	case flags&ast.ModifierFlagsProtected != 0:
		modifiers |= modifierProtected
	default:
		modifiers |= modifierPublic
	}
	if flags&ast.ModifierFlagsStatic != 0 {
		modifiers |= modifierStatic
	}
	if flags&ast.ModifierFlagsReadonly != 0 {
		modifiers |= modifierReadonly
	}
	if flags&ast.ModifierFlagsOverride != 0 {
		modifiers |= modifierOverride
	}
	if flags&ast.ModifierFlagsAbstract != 0 {
		modifiers |= modifierAbstract
	}

	return modifiers
}

// handleMember validates a member name, adding the `requiresQuotes` modifier
// when the name is not a valid identifier.
func (r *namingConventionRunner) handleMember(selector selectorFlags, nameNode *ast.Node, modifiers modifierFlags) {
	name, ok := getNameText(nameNode)
	if !ok {
		return
	}
	if requiresQuoting(name) {
		modifiers |= modifierRequiresQuotes
	}
	r.validate(selector, nameNode, modifiers)
}

func isAsyncFunctionNode(node *ast.Node) bool {
	return node != nil && node.ModifierFlags()&ast.ModifierFlagsAsync != 0
}

// isDestructured mirrors typescript-eslint's `isDestructured`: `const { x }`
// and `const { x = 2 }` match, `const { x: y }` does not.
func isDestructured(id *ast.Node) bool {
	parent := id.Parent
	if parent == nil || parent.Kind != ast.KindBindingElement {
		return false
	}
	element := parent.AsBindingElement()
	return element.PropertyName == nil &&
		element.DotDotDotToken == nil &&
		parent.Parent != nil &&
		parent.Parent.Kind == ast.KindObjectBindingPattern
}

// collectIdentifiersFromPattern collects all identifiers declared by a binding
// name (an identifier or a binding pattern).
func collectIdentifiersFromPattern(name *ast.Node, identifiers []*ast.Node) []*ast.Node {
	if name == nil {
		return identifiers
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return append(identifiers, name)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindBindingElement {
				identifiers = collectIdentifiersFromPattern(element.AsBindingElement().Name(), identifiers)
			}
		}
	}
	return identifiers
}

func isTopLevel(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindSourceFile
}

// isGlobalVariable reports whether a variable declaration is declared in the
// top-level scope (mirroring the `global` or `module` scope of eslint's scope
// manager).
func isGlobalVariable(declaration *ast.Node) bool {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	statement := list.Parent
	if statement == nil {
		return false
	}
	switch statement.Kind {
	case ast.KindVariableStatement:
		return isTopLevel(statement)
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		// `let`/`const` declarations in a `for` head create their own scope
		return isTopLevel(statement) && list.Flags&ast.NodeFlagsBlockScoped == 0
	}
	return false
}

// isExported mirrors typescript-eslint's `isExported`: the declaration has an
// `export` modifier, or the name is referenced by `export { name }` or
// `export default name`.
func (r *namingConventionRunner) isExported(declaration *ast.Node, nameNode *ast.Node) bool {
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return true
	}

	r.collectExportedSymbols()
	if len(r.exportedSymbols) == 0 {
		return false
	}
	symbol := r.ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return false
	}
	_, ok := r.exportedSymbols[symbol]
	return ok
}

func (r *namingConventionRunner) collectExportedSymbols() {
	if r.exportedSymbolsCollected {
		return
	}
	r.exportedSymbolsCollected = true

	for _, statement := range r.ctx.SourceFile.Statements.Nodes {
		switch statement.Kind {
		case ast.KindExportDeclaration:
			exportDeclaration := statement.AsExportDeclaration()
			if exportDeclaration.ModuleSpecifier != nil || exportDeclaration.ExportClause == nil || exportDeclaration.ExportClause.Kind != ast.KindNamedExports {
				continue
			}
			for _, specifier := range exportDeclaration.ExportClause.AsNamedExports().Elements.Nodes {
				if symbol := r.ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(specifier); symbol != nil {
					r.addExportedSymbol(symbol)
				}
			}
		case ast.KindExportAssignment:
			// `export default foo` (but not `export = foo`)
			exportAssignment := statement.AsExportAssignment()
			if exportAssignment.IsExportEquals || exportAssignment.Expression == nil || !ast.IsIdentifier(exportAssignment.Expression) {
				continue
			}
			if symbol := r.ctx.TypeChecker.GetSymbolAtLocation(exportAssignment.Expression); symbol != nil {
				r.addExportedSymbol(symbol)
			}
		}
	}
}

func (r *namingConventionRunner) addExportedSymbol(symbol *ast.Symbol) {
	if r.exportedSymbols == nil {
		r.exportedSymbols = make(map[*ast.Symbol]struct{})
	}
	r.exportedSymbols[symbol] = struct{}{}
}

// isUnused approximates typescript-eslint's `unused` modifier (which is based
// on the `no-unused-vars` logic): a declaration is unused when it is not
// exported, not ambient, and no read reference to its symbol exists outside of
// the declaration itself.
func (r *namingConventionRunner) isUnused(declaration *ast.Node, nameNode *ast.Node) bool {
	if r.isExported(declaration, nameNode) {
		return false
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsAmbient != 0 {
		return false
	}

	symbol := r.ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return false
	}

	r.collectReferences()

	references := r.references[symbol]
	if len(references) == 0 {
		return true
	}

	// references from within the declaration itself (e.g. recursion, or a type
	// referencing itself) don't count as usages
	selfRange := selfReferenceRange(declaration)
	if selfRange == nil {
		return false
	}
	for _, pos := range references {
		if !selfRange.ContainsInclusive(pos) {
			return false
		}
	}
	return true
}

// selfReferenceRange returns the range within which references to the
// declaration are considered self references, or nil if self references count
// as usages.
func selfReferenceRange(declaration *ast.Node) *core.TextRange {
	switch declaration.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration:
		return &declaration.Loc
	case ast.KindVariableDeclaration:
		initializer := declaration.Initializer()
		if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return &initializer.Loc
		}
	}
	return nil
}

// collectReferences walks the whole file once and records, for every symbol,
// the positions of the identifiers that read it.
func (r *namingConventionRunner) collectReferences() {
	if r.referencesCollected {
		return
	}
	r.referencesCollected = true
	r.references = make(map[*ast.Symbol][]int)

	typeChecker := r.ctx.TypeChecker

	var walk func(node *ast.Node)
	visitChild := func(child *ast.Node) bool {
		walk(child)
		return false
	}
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			var symbol *ast.Symbol
			parent := node.Parent
			switch {
			case parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node:
				// `{ x }` both declares a property and reads the variable `x`
				symbol = typeChecker.GetShorthandAssignmentValueSymbol(parent)
			case ast.IsDeclarationName(node), isWriteOnlyReference(node):
				// not a read
			default:
				symbol = typeChecker.GetSymbolAtLocation(node)
			}
			if symbol != nil {
				r.references[symbol] = append(r.references[symbol], node.Pos())
			}
			return
		}
		node.ForEachChild(visitChild)
	}
	r.ctx.SourceFile.Node.ForEachChild(visitChild)
}

// isWriteOnlyReference reports whether the identifier is only written to:
// `x = 1`, or `x += 1` / `x++` used as a statement.
func isWriteOnlyReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.Left != node {
			return false
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return true
		}
		return ast.IsAssignmentOperator(binary.OperatorToken.Kind) && parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement
	case ast.KindPrefixUnaryExpression:
		operator := parent.AsPrefixUnaryExpression().Operator
		return (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) && parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement
	case ast.KindPostfixUnaryExpression:
		operator := parent.AsPostfixUnaryExpression().Operator
		return (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) && parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement
	}
	return false
}

// #endregion modifiers

// #region handlers

func (r *namingConventionRunner) handleFunction(node *ast.Node) {
	name := node.Name()
	if name == nil {
		return
	}

	var modifiers modifierFlags

	// named function expressions create their own scope for their name, so
	// they are never global
	if node.Kind == ast.KindFunctionDeclaration && isTopLevel(node) {
		modifiers |= modifierGlobal
	}

	if r.isExported(node, name) {
		modifiers |= modifierExported
	}

	if r.hasUnusedModifierConfigured(selectorFunction) && r.isUnused(node, name) {
		modifiers |= modifierUnused
	}

	if isAsyncFunctionNode(node) {
		modifiers |= modifierAsync
	}

	r.validate(selectorFunction, name, modifiers)
}

func (r *namingConventionRunner) handleImportClause(node *ast.Node) {
	// `import Foo from 'foo'`
	name := node.Name()
	if name == nil {
		return
	}
	r.validate(selectorImport, name, modifierDefault)
}

func (r *namingConventionRunner) handleNamespaceImport(node *ast.Node) {
	// `import * as Foo from 'foo'`
	r.validate(selectorImport, node.Name(), modifierNamespace)
}

func (r *namingConventionRunner) handleImportSpecifier(node *ast.Node) {
	// Handle `import { default as Foo }`
	propertyName := node.PropertyName()
	if propertyName == nil {
		return
	}
	if propertyName.Kind == ast.KindIdentifier && propertyName.Text() != "default" {
		return
	}
	r.validate(selectorImport, node.Name(), modifierDefault)
}

func (r *namingConventionRunner) handleVariableDeclaration(node *ast.Node) {
	// catch clause variables are not covered by the `variable` selector
	if node.Parent == nil || node.Parent.Kind == ast.KindCatchClause {
		return
	}

	identifiers := collectIdentifiersFromPattern(node.Name(), nil)
	if len(identifiers) == 0 {
		return
	}

	var baseModifiers modifierFlags
	// `await using` sets both flags, and is not `const`
	if flags := node.Parent.Flags & ast.NodeFlagsAwaitUsing; flags == ast.NodeFlagsConst {
		baseModifiers |= modifierConst
	}
	if isGlobalVariable(node) {
		baseModifiers |= modifierGlobal
	}

	checkUnused := r.hasUnusedModifierConfigured(selectorVariable)
	initializer := node.Initializer()

	for _, id := range identifiers {
		modifiers := baseModifiers

		if isDestructured(id) {
			modifiers |= modifierDestructured
		}

		if r.isExported(node, id) {
			modifiers |= modifierExported
		}

		if checkUnused && r.isUnused(node, id) {
			modifiers |= modifierUnused
		}

		if id.Parent == node && initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) && isAsyncFunctionNode(initializer) {
			modifiers |= modifierAsync
		}

		r.validateWithRange(selectorVariable, id, r.getDeclarationNameRange(node, id), modifiers)
	}
}

func (r *namingConventionRunner) handleParameter(node *ast.Node) {
	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsParameterPropertyDeclaration(node, parent) {
		modifiers := getMemberModifiers(node)
		for _, id := range collectIdentifiersFromPattern(node.Name(), nil) {
			r.validateWithRange(selectorParameterProperty, id, r.getDeclarationNameRange(node, id), modifiers)
		}
		return
	}

	// typescript-eslint only checks the parameters of functions with a body or
	// function declarations, not the parameters of function types or signatures.
	switch parent.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
	default:
		return
	}

	checkUnused := r.hasUnusedModifierConfigured(selectorParameter)

	for _, id := range collectIdentifiersFromPattern(node.Name(), nil) {
		var modifiers modifierFlags

		if isDestructured(id) {
			modifiers |= modifierDestructured
		}

		if checkUnused && r.isUnused(node, id) {
			modifiers |= modifierUnused
		}

		r.validateWithRange(selectorParameter, id, r.getDeclarationNameRange(node, id), modifiers)
	}
}

func (r *namingConventionRunner) handlePropertyDeclaration(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}

	modifiers := getMemberModifiers(node)

	if ast.IsAutoAccessorPropertyDeclaration(node) {
		r.handleMember(selectorAutoAccessor, name, modifiers)
		return
	}

	// properties with direct function expression values are treated as methods
	initializer := node.Initializer()
	if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
		if isAsyncFunctionNode(initializer) {
			modifiers |= modifierAsync
		}
		r.handleMember(selectorClassMethod, name, modifiers)
		return
	}

	r.handleMember(selectorClassProperty, name, modifiers)
}

func (r *namingConventionRunner) handleMethodDeclaration(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}

	parent := node.Parent
	switch {
	case parent != nil && ast.IsClassLike(parent):
		modifiers := getMemberModifiers(node)
		if isAsyncFunctionNode(node) {
			modifiers |= modifierAsync
		}
		r.handleMember(selectorClassMethod, name, modifiers)
	case parent != nil && parent.Kind == ast.KindObjectLiteralExpression:
		if ast.IsAssignmentTarget(parent) {
			return
		}
		modifiers := modifierPublic
		if isAsyncFunctionNode(node) {
			modifiers |= modifierAsync
		}
		r.handleMember(selectorObjectLiteralMethod, name, modifiers)
	}
}

func (r *namingConventionRunner) handleAccessor(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}

	parent := node.Parent
	switch {
	case parent != nil && ast.IsClassLike(parent):
		r.handleMember(selectorClassicAccessor, name, getMemberModifiers(node))
	case parent != nil && parent.Kind == ast.KindObjectLiteralExpression:
		if ast.IsAssignmentTarget(parent) {
			return
		}
		r.handleMember(selectorClassicAccessor, name, modifierPublic)
	case parent != nil && (parent.Kind == ast.KindInterfaceDeclaration || parent.Kind == ast.KindTypeLiteral):
		// accessor signatures are method signatures in typescript-estree
		r.handleMember(selectorTypeMethod, name, modifierPublic)
	}
}

func (r *namingConventionRunner) handlePropertyAssignment(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}
	// object literals used as destructuring assignment targets are patterns, not
	// object literal properties
	if node.Parent == nil || ast.IsAssignmentTarget(node.Parent) {
		return
	}

	modifiers := modifierPublic

	// properties with direct function expression values are treated as methods
	initializer := node.Initializer()
	if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
		if isAsyncFunctionNode(initializer) {
			modifiers |= modifierAsync
		}
		r.handleMember(selectorObjectLiteralMethod, name, modifiers)
		return
	}

	r.handleMember(selectorObjectLiteralProperty, name, modifiers)
}

func (r *namingConventionRunner) handleShorthandPropertyAssignment(node *ast.Node) {
	name := node.Name()
	if name == nil || node.Parent == nil || ast.IsAssignmentTarget(node.Parent) {
		return
	}
	r.handleMember(selectorObjectLiteralProperty, name, modifierPublic)
}

func (r *namingConventionRunner) handleMethodSignature(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}
	r.handleMember(selectorTypeMethod, name, modifierPublic)
}

func (r *namingConventionRunner) handlePropertySignature(node *ast.Node) {
	name := getMemberNameNode(node.Name())
	if name == nil {
		return
	}

	// properties with a function type are treated as methods
	if typeNode := node.Type(); typeNode != nil && typeNode.Kind == ast.KindFunctionType {
		r.handleMember(selectorTypeMethod, name, modifierPublic)
		return
	}

	modifiers := modifierPublic
	if node.ModifierFlags()&ast.ModifierFlagsReadonly != 0 {
		modifiers |= modifierReadonly
	}
	r.handleMember(selectorTypeProperty, name, modifiers)
}

func (r *namingConventionRunner) handleClass(node *ast.Node) {
	name := node.Name()
	if name == nil {
		return
	}

	var modifiers modifierFlags

	if node.ModifierFlags()&ast.ModifierFlagsAbstract != 0 {
		modifiers |= modifierAbstract
	}

	if r.isExported(node, name) {
		modifiers |= modifierExported
	}

	if r.hasUnusedModifierConfigured(selectorClass) && r.isUnused(node, name) {
		modifiers |= modifierUnused
	}

	r.validate(selectorClass, name, modifiers)
}

func (r *namingConventionRunner) handleEnum(node *ast.Node) {
	r.handleTypeDeclaration(selectorEnum, node)
}

func (r *namingConventionRunner) handleInterface(node *ast.Node) {
	r.handleTypeDeclaration(selectorInterface, node)
}

func (r *namingConventionRunner) handleTypeAlias(node *ast.Node) {
	r.handleTypeDeclaration(selectorTypeAlias, node)
}

func (r *namingConventionRunner) handleTypeDeclaration(selector selectorFlags, node *ast.Node) {
	name := node.Name()
	if name == nil {
		return
	}

	var modifiers modifierFlags

	if r.isExported(node, name) {
		modifiers |= modifierExported
	}

	if r.hasUnusedModifierConfigured(selector) && r.isUnused(node, name) {
		modifiers |= modifierUnused
	}

	r.validate(selector, name, modifiers)
}

func (r *namingConventionRunner) handleEnumMember(node *ast.Node) {
	name := node.Name()
	if name == nil {
		return
	}
	// `enum Foo { ['a'] }` is represented with a plain literal key in typescript-estree
	if name.Kind == ast.KindComputedPropertyName {
		expression := name.Expression()
		if expression == nil || !ast.IsStringLiteralLike(expression) {
			return
		}
		name = expression
	}
	if _, ok := getNameText(name); !ok {
		return
	}

	r.handleMember(selectorEnumMember, name, 0)
}

func (r *namingConventionRunner) handleTypeParameter(node *ast.Node) {
	// only type parameters of type parameter declarations, i.e. not `infer U`
	// nor the key of a mapped type
	if node.Parent == nil || node.Parent.Kind == ast.KindInferType || node.Parent.Kind == ast.KindMappedType {
		return
	}

	name := node.Name()
	if name == nil {
		return
	}

	var modifiers modifierFlags

	if r.hasUnusedModifierConfigured(selectorTypeParameter) && r.isUnused(node, name) {
		modifiers |= modifierUnused
	}

	r.validate(selectorTypeParameter, name, modifiers)
}

// #endregion handlers
