package naming_convention

import (
	"maps"
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// NamingConventionRule enforces the typescript-eslint naming-convention selectors.
var NamingConventionRule = rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		validator := newNamingValidator(ctx, options)
		var scope *namingScope
		getScope := func() *namingScope {
			if scope == nil {
				scope = newNamingScope(ctx)
			}
			return scope
		}
		scopeModifiers := func(node, id *ast.Node, global, exported, unused bool) map[string]bool {
			modifiers := map[string]bool{}
			if global && validator.needsModifier("global") {
				modifiers["global"] = getScope().isGlobal(node)
			}
			if exported && validator.needsModifier("exported") {
				modifiers["exported"] = getScope().isExported(id)
			}
			if unused && validator.needsModifier("unused") {
				modifiers["unused"] = getScope().isUnused(id)
			}
			return modifiers
		}
		checkNamed := func(node *ast.Node, selector string, global, exported, unused bool) {
			id := node.Name()
			if id == nil {
				return
			}
			modifiers := scopeModifiers(node, id, global, exported, unused)
			if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
				modifiers["async"] = true
			}
			if node.ModifierFlags()&ast.ModifierFlagsAbstract != 0 {
				modifiers["abstract"] = true
			}
			validator.check(id, selector, modifiers)
		}
		checkMember := func(node *ast.Node, selector string, classMember bool) {
			name := node.Name()
			if name == nil {
				return
			}
			if name.Kind == ast.KindComputedPropertyName {
				if selector != "autoAccessor" {
					return
				}
				name = ast.SkipParentheses(name.Expression())
			}
			modifiers := map[string]bool{"public": true}
			if classMember {
				modifiers = namingMemberModifiers(node)
			}
			if selector == "typeProperty" && node.ModifierFlags()&ast.ModifierFlagsReadonly != 0 {
				modifiers["readonly"] = true
			}
			if namingRequiresQuotes(name) {
				modifiers["requiresQuotes"] = true
			}
			if selector == "classMethod" || selector == "objectLiteralMethod" {
				function := node
				if node.Kind == ast.KindPropertyDeclaration || node.Kind == ast.KindPropertyAssignment {
					function = namingInitializer(node)
				}
				if function != nil && function.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
					modifiers["async"] = true
				}
			}
			validator.check(name, selector, modifiers)
		}
		checkParameters := func(node *ast.Node) {
			for _, parameter := range node.Parameters() {
				if ast.IsParameterPropertyDeclaration(parameter, node) {
					continue
				}
				namingBindings(parameter.Name(), func(id *ast.Node) {
					modifiers := scopeModifiers(parameter, id, false, false, true)
					if namingIsDestructured(id) {
						modifiers["destructured"] = true
					}
					validator.check(id, "parameter", modifiers)
				})
			}
		}
		functions := func(node *ast.Node) {
			checkNamed(node, "function", true, true, true)
			checkParameters(node)
		}
		methods := func(node *ast.Node) {
			selector := "classMethod"
			classMember := node.Parent.Kind != ast.KindObjectLiteralExpression
			if !classMember {
				selector = "objectLiteralMethod"
			}
			checkMember(node, selector, classMember)
			checkParameters(node)
		}
		accessors := func(node *ast.Node) {
			checkMember(node, "classicAccessor", node.Parent.Kind != ast.KindObjectLiteralExpression)
			checkParameters(node)
		}
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: functions,
			ast.KindFunctionExpression:  functions,
			ast.KindArrowFunction:       checkParameters,
			ast.KindConstructor:         checkParameters,
			ast.KindVariableDeclaration: func(node *ast.Node) {
				// A catch binding is not an ESTree VariableDeclarator.
				if node.Parent.Kind == ast.KindCatchClause {
					return
				}
				namingBindings(node.Name(), func(id *ast.Node) {
					modifiers := scopeModifiers(node, id, true, true, true)
					if node.Parent.Flags&ast.NodeFlagsConst != 0 {
						modifiers["const"] = true
					}
					if namingIsDestructured(id) {
						modifiers["destructured"] = true
					}
					if id == node.Name() {
						initializer := namingInitializer(node)
						if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) && initializer.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
							modifiers["async"] = true
						}
					}
					validator.check(id, "variable", modifiers)
				})
			},
			ast.KindParameter: func(node *ast.Node) {
				if !ast.IsParameterPropertyDeclaration(node, node.Parent) {
					return
				}
				modifiers := namingMemberModifiers(node)
				namingBindings(node.Name(), func(id *ast.Node) { validator.check(id, "parameterProperty", maps.Clone(modifiers)) })
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				selector := "classProperty"
				if ast.IsAutoAccessorPropertyDeclaration(node) {
					selector = "autoAccessor"
				} else if initializer := namingInitializer(node); initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
					selector = "classMethod"
				}
				checkMember(node, selector, true)
			},
			ast.KindPropertyAssignment: func(node *ast.Node) {
				if namingInAssignmentPattern(node) {
					return
				}
				selector := "objectLiteralProperty"
				if initializer := namingInitializer(node); initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
					selector = "objectLiteralMethod"
				}
				checkMember(node, selector, false)
			},
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				if !namingInAssignmentPattern(node) {
					checkMember(node, "objectLiteralProperty", false)
				}
			},
			ast.KindMethodDeclaration: methods,
			ast.KindGetAccessor:       accessors,
			ast.KindSetAccessor:       accessors,
			ast.KindMethodSignature:   func(node *ast.Node) { checkMember(node, "typeMethod", false) },
			ast.KindPropertySignature: func(node *ast.Node) {
				selector := "typeProperty"
				if node.Type() != nil && ast.SkipTypeParentheses(node.Type()).Kind == ast.KindFunctionType {
					selector = "typeMethod"
				}
				checkMember(node, selector, false)
			},
			ast.KindClassDeclaration:     func(node *ast.Node) { checkNamed(node, "class", false, true, true) },
			ast.KindClassExpression:      func(node *ast.Node) { checkNamed(node, "class", false, true, true) },
			ast.KindInterfaceDeclaration: func(node *ast.Node) { checkNamed(node, "interface", false, true, true) },
			ast.KindTypeAliasDeclaration: func(node *ast.Node) { checkNamed(node, "typeAlias", false, true, true) },
			ast.KindEnumDeclaration:      func(node *ast.Node) { checkNamed(node, "enum", false, true, true) },
			ast.KindEnumMember: func(node *ast.Node) {
				modifiers := map[string]bool{}
				if namingRequiresQuotes(node.Name()) {
					modifiers["requiresQuotes"] = true
				}
				validator.check(node.Name(), "enumMember", modifiers)
			},
			ast.KindTypeParameter: func(node *ast.Node) {
				if node.Parent.Kind != ast.KindMappedType && node.Parent.Kind != ast.KindInferType {
					checkNamed(node, "typeParameter", false, false, true)
				}
			},
			ast.KindImportClause: func(node *ast.Node) {
				if node.Name() != nil {
					validator.check(node.Name(), "import", map[string]bool{"default": true})
				}
			},
			ast.KindNamespaceImport: func(node *ast.Node) { validator.check(node.Name(), "import", map[string]bool{"namespace": true}) },
			ast.KindImportSpecifier: func(node *ast.Node) {
				imported := node.AsImportSpecifier().PropertyName
				if imported == nil {
					imported = node.Name()
				}
				if imported.Kind == ast.KindIdentifier && imported.Text() != "default" {
					return
				}
				validator.check(node.Name(), "import", map[string]bool{"default": true})
			},
		}
	},
}

func namingBindings(name *ast.Node, visit func(*ast.Node)) {
	if name == nil {
		return
	}
	if name.Kind == ast.KindIdentifier {
		visit(name)
		return
	}
	if ast.IsBindingPattern(name) {
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindBindingElement {
				namingBindings(element.Name(), visit)
			}
		}
	}
}

func namingIsDestructured(id *ast.Node) bool {
	parent := id.Parent
	return parent.Kind == ast.KindBindingElement && parent.Parent.Kind == ast.KindObjectBindingPattern && parent.AsBindingElement().PropertyName == nil && parent.AsBindingElement().DotDotDotToken == nil
}

func namingMemberModifiers(node *ast.Node) map[string]bool {
	modifiers := map[string]bool{}
	flags := node.ModifierFlags()
	switch {
	case node.Name() != nil && node.Name().Kind == ast.KindPrivateIdentifier:
		modifiers["#private"] = true
	case flags&ast.ModifierFlagsPrivate != 0:
		modifiers["private"] = true
	case flags&ast.ModifierFlagsProtected != 0:
		modifiers["protected"] = true
	default:
		modifiers["public"] = true
	}
	for _, entry := range []struct {
		flag ast.ModifierFlags
		name string
	}{
		{ast.ModifierFlagsStatic, "static"}, {ast.ModifierFlagsReadonly, "readonly"},
		{ast.ModifierFlagsOverride, "override"}, {ast.ModifierFlagsAbstract, "abstract"},
	} {
		if flags&entry.flag != 0 {
			modifiers[entry.name] = true
		}
	}
	return modifiers
}

// Upstream tests UTF-16 code units (charCodeAt), rather than Unicode code points.
func namingRequiresQuotes(name *ast.Node) bool {
	text := namingNodeName(name)
	units := utf16.Encode([]rune(text))
	if len(units) == 0 || !scanner.IsIdentifierStart(rune(units[0])) {
		return true
	}
	for _, ch := range units[1:] {
		if !scanner.IsIdentifierPart(rune(ch)) {
			return true
		}
	}
	return false
}

func namingInAssignmentPattern(node *ast.Node) bool {
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindForInStatement, ast.KindForOfStatement:
			return parent.Initializer() == current
		case ast.KindBinaryExpression:
			return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == current
		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindSpreadAssignment, ast.KindSpreadElement, ast.KindParenthesizedExpression:
		default:
			return false
		}
	}
	return false
}

func namingInitializer(node *ast.Node) *ast.Node {
	initializer := node.Initializer()
	if initializer == nil {
		return nil
	}
	return ast.SkipParentheses(initializer)
}
