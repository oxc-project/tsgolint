package naming_convention

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// namingScope stores file-local facts needed by the unused/exported/global
// naming-convention modifiers. It is built lazily by the rule when any of
// those modifiers are requested.
type namingScope struct {
	ctx      rule.RuleContext
	unused   map[*ast.Symbol]bool
	exported map[*ast.Symbol]bool
}

func newNamingScope(ctx rule.RuleContext) *namingScope {
	scope := &namingScope{
		ctx:      ctx,
		unused:   make(map[*ast.Symbol]bool),
		exported: make(map[*ast.Symbol]bool),
	}
	if ctx.SourceFile == nil || ctx.TypeChecker == nil {
		return scope
	}

	// This is one whole-file walk per rule invocation. Resolving identifiers
	// here lets all subsequent modifier checks use constant-time symbol maps.
	var identifiers []*ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if ast.IsIdentifier(node) {
			identifiers = append(identifiers, node)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(ctx.SourceFile.AsNode())

	// First establish the declaration symbols under this source file. Symbols
	// can have several merged declarations, so any read of the symbol marks
	// every declaration in the merge as used.
	for _, identifier := range identifiers {
		if !ast.IsDeclarationName(identifier) {
			continue
		}
		symbol := scope.symbolAtLocation(identifier)
		if symbol == nil {
			continue
		}
		if _, exists := scope.unused[symbol]; !exists {
			scope.unused[symbol] = true
		}
	}

	for _, identifier := range identifiers {
		symbol := scope.symbolAtLocation(identifier)
		if symbol == nil {
			continue
		}
		if isExportSpecifierName(identifier) || isExportDefaultReference(identifier) {
			scope.markExportedSymbol(symbol)
		}

		if ast.IsWriteOnlyAccess(identifier) && !isShorthandValueReference(identifier) || ast.IsDeclarationName(identifier) && !isShorthandValueReference(identifier) {
			continue
		}
		if isTypeQueryOrPredicateReference(identifier) || isUnusedSelfReference(symbol, identifier) || scope.isUnusedSelfAssignmentReference(symbol, identifier) || isUnusedSelfWrite(identifier) {
			continue
		}
		scope.unused[symbol] = false
	}

	// An export modifier applies to each declaration in a merged symbol.
	for _, identifier := range identifiers {
		if !ast.IsDeclarationName(identifier) {
			continue
		}
		symbol := scope.symbolAtLocation(identifier)
		if symbol == nil || scope.exported[symbol] {
			continue
		}
		if declarationIsDirectlyExported(identifier.Parent) {
			scope.exported[symbol] = true
		}
	}

	return scope
}

func isShorthandValueReference(identifier *ast.Node) bool {
	return identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment
}

func (s *namingScope) symbolAtLocation(identifier *ast.Node) *ast.Symbol {
	if isShorthandValueReference(identifier) {
		if symbol := checker.Checker_GetShorthandAssignmentValueSymbol(s.ctx.TypeChecker, identifier.Parent); symbol != nil {
			return symbol
		}
	}
	return s.ctx.TypeChecker.GetSymbolAtLocation(identifier)
}

func (s *namingScope) markExportedSymbol(symbol *ast.Symbol) {
	if symbol == nil {
		return
	}
	s.exported[symbol] = true
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if target, ok := s.ctx.TypeChecker.ResolveAlias(symbol); ok && target != nil {
			s.exported[target] = true
		}
	}
}

func (s *namingScope) isUnused(identifier *ast.Node) bool {
	if s == nil || s.ctx.TypeChecker == nil || identifier == nil {
		return false
	}
	symbol := s.symbolAtLocation(identifier)
	return symbol != nil && s.unused[symbol] && !s.exported[symbol] && !isImplicitlyUsed(identifier, symbol)
}

func (s *namingScope) isExported(identifier *ast.Node) bool {
	if s == nil || s.ctx.TypeChecker == nil || identifier == nil {
		return false
	}
	if declarationIsDirectlyExported(identifier.Parent) {
		return true
	}
	symbol := s.symbolAtLocation(identifier)
	return symbol != nil && s.exported[symbol]
}

func (s *namingScope) isGlobal(declaration *ast.Node) bool {
	if declaration == nil || s == nil || s.ctx.SourceFile == nil {
		return false
	}
	if declaration.Kind == ast.KindFunctionExpression || declaration.Kind == ast.KindClassExpression {
		return false
	}
	if !ast.IsDeclaration(declaration) {
		return false
	}

	// Naming convention follows the declaration's ESLint scope. Declarations
	// inside blocks are local even when JavaScript hoists a `var` binding.
	isVar := declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil &&
		declaration.Parent.Kind == ast.KindVariableDeclarationList &&
		declaration.Parent.Flags&ast.NodeFlagsBlockScoped == 0
	for node := declaration.Parent; node != nil && node.Kind != ast.KindSourceFile; node = node.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) || node.Kind == ast.KindModuleBlock || ast.IsClassLike(node) {
			return false
		}
		if !isVar && (node.Kind == ast.KindForStatement || node.Kind == ast.KindForInStatement || node.Kind == ast.KindForOfStatement) {
			return false
		}
		if node.Kind == ast.KindBlock || node.Kind == ast.KindCaseBlock {
			return false
		}
	}
	return true
}

func isExportSpecifierName(identifier *ast.Node) bool {
	if identifier.Parent == nil || identifier.Parent.Kind != ast.KindExportSpecifier {
		return false
	}
	return true
}

func isExportDefaultReference(identifier *ast.Node) bool {
	if identifier.Parent == nil {
		return false
	}
	return identifier.Parent.Kind == ast.KindExportAssignment || identifier.Parent.Kind == ast.KindExportDeclaration
}

func declarationIsDirectlyExported(declaration *ast.Node) bool {
	if declaration == nil || declaration.Parent == nil {
		return false
	}
	parent := declaration.Parent
	if parent.Kind == ast.KindVariableDeclarationList && parent.Parent != nil {
		parent = parent.Parent
	}
	if parent.Kind == ast.KindExportAssignment || parent.Kind == ast.KindExportDeclaration {
		return true
	}
	return ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport) || ast.HasSyntacticModifier(parent, ast.ModifierFlagsExport)
}

func isTypeQueryOrPredicateReference(identifier *ast.Node) bool {
	for node := identifier.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindTypeQuery, ast.KindTypePredicate:
			return true
		case ast.KindSourceFile:
			return false
		}
	}
	return false
}

func isUnusedSelfReference(symbol *ast.Symbol, identifier *ast.Node) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !isSelfReferenceDeclaration(declaration) {
			continue
		}
		for node := identifier; node != nil; node = node.Parent {
			if node == declaration {
				return true
			}
		}
	}
	return false
}

func isSelfReferenceDeclaration(declaration *ast.Node) bool {
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindModuleDeclaration, ast.KindEnumDeclaration, ast.KindClassDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return true
	case ast.KindVariableDeclaration:
		initializer := declaration.AsVariableDeclaration().Initializer
		return initializer != nil && ast.IsFunctionLike(ast.SkipParentheses(initializer))
	default:
		return false
	}
}

// A read on the right of a discarded self-assignment does not use the
// variable. References in a function on that right-hand side still count if
// the function can escape the assignment and be called later.
func (s *namingScope) isUnusedSelfAssignmentReference(symbol *ast.Symbol, identifier *ast.Node) bool {
	for node := identifier.Parent; node != nil; node = node.Parent {
		if node.Kind != ast.KindBinaryExpression {
			continue
		}
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken || !isUnusedExpression(node) {
			continue
		}
		left := ast.SkipParentheses(binary.Left)
		if !ast.IsIdentifier(left) || s.symbolAtLocation(left) != symbol {
			continue
		}
		if variableScope(node) != variableScopeOfSymbol(symbol) || isInLoop(node) {
			// An enclosing self-assignment may still own this entire RHS,
			// including assignments inside a function that cannot escape it.
			continue
		}
		return !isInsideOfStorableFunction(identifier, binary.Right)
	}
	return false
}

func variableScopeOfSymbol(symbol *ast.Symbol) *ast.Node {
	for _, declaration := range symbol.Declarations {
		if ast.IsDeclarationName(declaration) {
			declaration = declaration.Parent
		}
		if scope := variableScope(declaration); scope != nil {
			return scope
		}
	}
	return nil
}

func variableScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) || current.Kind == ast.KindSourceFile || current.Kind == ast.KindModuleBlock {
			return current
		}
		if parent := current.Parent; parent != nil && parent.Kind == ast.KindPropertyDeclaration && parent.Initializer() == current {
			// ESLint gives each class field initializer its own variable scope.
			return current
		}
	}
	return nil
}

// isInsideOfStorableFunction mirrors typescript-eslint's unused-variable
// analysis: a self-reference inside a function on the discarded assignment's
// RHS is only a read if that function can escape and be called later.
func isInsideOfStorableFunction(identifier, rhs *ast.Node) bool {
	if identifier == nil || rhs == nil || !isNodeWithin(identifier, rhs) {
		return false
	}
	var function *ast.Node
	for node := identifier.Parent; node != nil; node = node.Parent {
		if ast.IsFunctionLike(node) {
			function = node
			break
		}
	}
	if function == nil || !isNodeWithin(function, rhs) {
		return false
	}

	for node := function; node != nil && node != rhs; {
		parent := node.Parent
		if parent == nil || !isNodeWithin(parent, rhs) {
			break
		}
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindCommaToken {
				if binary.Right != node {
					return false
				}
			} else if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return true
			}
		case ast.KindCallExpression:
			return parent.AsCallExpression().Expression != node
		case ast.KindNewExpression:
			return parent.AsNewExpression().Expression != node
		case ast.KindTaggedTemplateExpression, ast.KindYieldExpression:
			return true
		default:
			// A function nested in a statement or declaration is conservatively
			// considered storable, matching typescript-eslint's complex-pattern
			// handling.
			// Native function bodies are Blocks too, but IsStatement excludes them.
			if ast.IsStatement(parent) || parent.Kind == ast.KindBlock {
				return true
			}
		}
		node = parent
	}
	return false
}

func isNodeWithin(node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func isInLoop(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return false
		}
		switch current.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement:
			return true
		}
	}
	return false
}

func isUnusedSelfWrite(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPrefixUnaryExpression {
		operator := parent.AsPrefixUnaryExpression().Operator
		return (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) && isUnusedExpression(parent)
	}
	if parent.Kind == ast.KindPostfixUnaryExpression {
		operator := parent.AsPostfixUnaryExpression().Operator
		return (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) && isUnusedExpression(parent)
	}
	if parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == identifier && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
		// Logical assignments read the existing value to choose whether to write.
		// Upstream excludes them from the discarded self-update heuristic.
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return false
		}
		return isUnusedExpression(parent)
	}
	return false
}

func isUnusedExpression(node *ast.Node) bool {
	for node != nil {
		parent := node.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindExpressionStatement:
			return true
		case ast.KindBinaryExpression:
			if parent.AsBinaryExpression().OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			if parent.AsBinaryExpression().Right != node {
				return true
			}
			node = parent
		default:
			return false
		}
	}
	return false
}

func isImplicitlyUsed(identifier *ast.Node, symbol *ast.Symbol) bool {
	if ast.IsIdentifier(identifier) && identifier.AsIdentifier().Text == "this" &&
		identifier.Parent != nil && ast.IsParameterDeclaration(identifier.Parent) {
		return true
	}
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindEnumMember, ast.KindFunctionExpression, ast.KindClassExpression:
			// Enum members are treated as used by collectVariables. Named
			// function/class expression names live in expression-local scopes;
			// class self-names are marked used by its visitor.
			return true
		}
	}
	var parameter *ast.Node
	for _, declaration := range symbol.Declarations {
		if parameter = parameterDeclarationForBinding(declaration); parameter != nil {
			break
		}
	}
	if parameter == nil {
		return false
	}
	for node := parameter.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindCallSignature, ast.KindConstructSignature, ast.KindConstructorType,
			ast.KindFunctionType, ast.KindMethodSignature:
			return true
		case ast.KindSetAccessor:
			return true
		case ast.KindSourceFile:
			return false
		}
		if ast.IsFunctionLike(node) {
			// typescript-eslint's collectVariables does not report parameters
			// from bodyless function-like declarations as unused. This includes
			// declare functions and abstract/ambient methods and constructors.
			return node.Body() == nil
		}
	}
	return false
}

// parameterDeclarationForBinding finds the parameter that owns a binding.
// Destructured parameters are represented by BindingElement declarations,
// rather than by the ParameterDeclaration itself.
func parameterDeclarationForBinding(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	if ast.IsParameterDeclaration(declaration) {
		return declaration
	}
	if declaration.Kind != ast.KindBindingElement {
		return nil
	}
	for node := declaration.Parent; node != nil; node = node.Parent {
		if ast.IsParameterDeclaration(node) {
			return node
		}
		if ast.IsFunctionLike(node) {
			return nil
		}
	}
	return nil
}
