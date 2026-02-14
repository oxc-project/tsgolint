package unbound_method

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

//go:generate node ./generate-natively-bound-members.mjs

const baseMessage = "Avoid referencing unbound methods which may cause unintentional scoping of `this`."

func buildUnboundMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unbound",
		Description: baseMessage,
	}
}
func buildUnboundWithoutThisAnnotationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unboundWithoutThisAnnotation",
		Description: baseMessage,
		Help:        "If your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.",
	}
}

func isNodeInsideTypeDeclaration(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindClassDeclaration:
			if utils.IncludesModifier(parent, ast.KindDeclareKeyword) {
				return true
			}
		case ast.KindMethodDeclaration:
			if utils.IncludesModifier(parent, ast.KindAbstractKeyword) {
				return true
			}
		case ast.KindFunctionDeclaration:
			if parent.Body() == nil {
				return true
			}
		case ast.KindFunctionType, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			return true
		case ast.KindVariableStatement:
			if utils.IncludesModifier(parent, ast.KindDeclareKeyword) {
				return true
			}
		}
	}
	return false
}

func isSafeUse(node *ast.Node) bool {
	parent := node
	for {
		node = parent
		parent = parent.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			continue
		case ast.KindIfStatement,
			ast.KindForStatement,
			ast.KindPropertyAccessExpression,
			ast.KindElementAccessExpression,
			ast.KindSwitchStatement,
			ast.KindWhileStatement:
			return true
		case ast.KindPostfixUnaryExpression:
			operator := parent.AsPostfixUnaryExpression().Operator
			return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
		case ast.KindPrefixUnaryExpression:
			operator := parent.AsPrefixUnaryExpression().Operator
			return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken || operator == ast.KindExclamationToken
		case ast.KindCallExpression:
			return parent.Expression() == node
		case ast.KindConditionalExpression:
			return parent.AsConditionalExpression().Condition == node
		case ast.KindTaggedTemplateExpression:
			return parent.AsTaggedTemplateExpression().Tag == node
		case ast.KindDeleteExpression, ast.KindTypeOfExpression, ast.KindVoidExpression:
			return true
		case ast.KindBinaryExpression:
			expr := parent.AsBinaryExpression()
			operatorKind := expr.OperatorToken.Kind
			switch operatorKind {
			case ast.KindAmpersandAmpersandToken:
				if expr.Left == node {
					// this is safe, as && will return the left if and only if it"s falsy
					return true
				}
				// in all other cases, it's likely the logical expression will return the method ref
				// so make sure the parent is a safe usage
				continue
			case ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindInstanceOfKeyword:
				return true
			}
			if ast.IsLogicalBinaryOperator(operatorKind) {
				continue
			}
			if ast.IsAssignmentExpression(parent, true) {
				return node == expr.Left || (ast.IsAccessExpression(node) && node.Expression().Kind == ast.KindSuperKeyword && ast.IsAccessExpression(expr.Left) && expr.Left.Expression().Kind == ast.KindThisKeyword)
			}
			return false
		case ast.KindNonNullExpression,
			ast.KindAsExpression,
			ast.KindTypeAssertionExpression:
			continue
		}
		return false
	}
	return false
}

func isNotImported(symbol *ast.Symbol, currentSourceFile *ast.SourceFile) bool {
	decl := symbol.ValueDeclaration
	if decl == nil {
		// working around https://github.com/microsoft/TypeScript/issues/31294
		return false
	}

	return currentSourceFile != ast.GetSourceFileOfNode(decl)
}

var supportedGlobalTypes = []string{
	"NumberConstructor",
	"ObjectConstructor",
	"StringConstructor",
	"SymbolConstructor",
	"ArrayConstructor",
	"Array",
	"ProxyConstructor",
	"Console",
	"DateConstructor",
	"Atomics",
	"Math",
	"JSON",
}

func checkMethod(valueDeclaration *ast.Node, ignoreStatic bool) ( /* dangerous */ bool /* firstParamIsThis */, bool) {
	params := valueDeclaration.Parameters()

	firstParamIsThis := len(params) > 0 && ast.IsParameter(params[0]) && ast.IsIdentifier(params[0].Name()) && params[0].Name().Text() == "this"

	thisArgIsVoid := firstParamIsThis && params[0].Type().Kind == ast.KindVoidKeyword

	dangerous := !thisArgIsVoid && (!ignoreStatic || !utils.IncludesModifier(valueDeclaration, ast.KindStaticKeyword))

	return dangerous, firstParamIsThis
}

func checkIfMethod(symbol *ast.Symbol, ignoreStatic bool) ( /* dangerous */ bool /* firstParamIsThis */, bool) {
	valueDeclaration := symbol.ValueDeclaration
	if valueDeclaration == nil {
		// working around https://github.com/microsoft/TypeScript/issues/31294
		return false, false
	}

	switch valueDeclaration.Kind {
	case ast.KindPropertyDeclaration:
		init := valueDeclaration.Initializer()
		return init != nil && ast.IsFunctionExpression(valueDeclaration.Initializer()), true
	case ast.KindPropertyAssignment:
		assignee := valueDeclaration.Initializer()
		if !ast.IsFunctionExpression(assignee) {
			return false, false
		}

		return checkMethod(assignee, ignoreStatic)
	case ast.KindMethodDeclaration, ast.KindMethodSignature:
		return checkMethod(valueDeclaration, ignoreStatic)
	}

	return false, false
}

var UnboundMethodRule = rule.Rule{
	Name: "unbound-method",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[UnboundMethodOptions](options, "unbound-method")

		isNativelyBound := func(object *ast.Node, property *ast.Node) bool {
			// We can't rely entirely on the type-level checks made at the end of this
			// function, because sometimes type declarations don't come from the
			// default library, but come from, for example, "@types/node". And we can't
			// tell if a method is unbound just by looking at its signature declared in
			// the interface.
			//
			// See related discussion https://github.com/typescript-eslint/typescript-eslint/pull/8952#discussion_r1576543310
			if ast.IsIdentifier(object) && ast.IsIdentifier(property) {
				objectSymbol := ctx.TypeChecker.GetSymbolAtLocation(object)
				notImported := objectSymbol != nil && isNotImported(objectSymbol, ctx.SourceFile)

				if notImported {
					if members, ok := nativelyBoundMembers[object.Text()]; ok {
						if _, ok := members[property.Text()]; ok {
							return true
						}
					}
				}
			}

			// if `${object.name}.${property.name}` doesn't match any of
			// the nativelyBoundMembers, then we fallback to type-level checks
			return utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(object), supportedGlobalTypes...) && utils.IsAnyBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(property))
		}

		checkIfMethodAndReport := func(node *ast.Node, symbol *ast.Symbol) bool {
			if symbol == nil {
				return false
			}

			dangerous, firstParamIsThis := checkIfMethod(symbol, opts.IgnoreStatic)

			if !dangerous {
				return false
			}

			if firstParamIsThis {
				ctx.ReportNode(node, buildUnboundMessage())
			} else {
				ctx.ReportNode(node, buildUnboundWithoutThisAnnotationMessage())
			}
			return true
		}

		checkBindingProperty := func(patternNode *ast.Node, initNode *ast.Node, propertyName *ast.Node, parentIsAssignmentPatternLike bool) {
			// Skip computed property names as they cannot be statically analyzed
			if ast.IsComputedPropertyName(propertyName) {
				return
			}

			if initNode != nil {
				if !isNativelyBound(initNode, propertyName) {
					reported := checkIfMethodAndReport(propertyName, checker.Checker_getPropertyOfType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(initNode), propertyName.Text()))
					if reported {
						return
					}
					// In assignment patterns, we should also check the type of
					// Foo's nativelyBound method because initNode might be used as
					// default value:
					//   function ({ nativelyBound }: Foo = NativeObject) {}
				} else if !parentIsAssignmentPatternLike {
					return
				}
			}

			utils.TypeRecurser(ctx.TypeChecker.GetTypeAtLocation(patternNode), func(t *checker.Type) bool {
				return checkIfMethodAndReport(propertyName, checker.Checker_getPropertyOfType(ctx.TypeChecker, t, propertyName.Text()))
			})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if isSafeUse(node) || isNativelyBound(node.Expression(), node.Name()) {
					return
				}

				checkIfMethodAndReport(node, ctx.TypeChecker.GetSymbolAtLocation(node))
			},

			rule.ListenerOnAllowPattern(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node.Parent, true) {
					return
				}

				initNode := node.Parent.AsBinaryExpression().Right

				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					if !ast.IsPropertyAssignment(property) && !ast.IsShorthandPropertyAssignment(property) {
						continue
					}

					checkBindingProperty(node, initNode, property.Name(), true)
				}
			},
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if isNodeInsideTypeDeclaration(node) {
					return
				}

				var initNode *ast.Node

				parentIsAssignmentPatternLike := ast.IsBindingElement(node.Parent) || ast.IsParameter(node.Parent)
				if ast.IsVariableDeclaration(node.Parent) || parentIsAssignmentPatternLike {
					initNode = node.Parent.Initializer()
				}

				for _, property := range node.AsBindingPattern().Elements.Nodes {
					if !ast.IsBindingElement(property) {
						continue
					}

					bindingElem := property.AsBindingElement()
					propertyName := bindingElem.PropertyName
					if propertyName == nil {
						propertyName = bindingElem.Name()
					}
					if bindingElem.DotDotDotToken != nil || !ast.IsIdentifier(propertyName) {
						continue
					}

					checkBindingProperty(node, initNode, propertyName, parentIsAssignmentPatternLike)
				}
			},
		}
	},
}
