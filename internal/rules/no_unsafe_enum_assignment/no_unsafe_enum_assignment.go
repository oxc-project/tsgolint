package no_unsafe_enum_assignment

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var NoUnsafeEnumAssignmentRule = rule.Rule{
	Name: "no-unsafe-enum-assignment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		c := &enumTypeChecker{checker: ctx.TypeChecker, genericTypes: make(map[*checker.Type]bool), mismatches: make(map[typePair]bool)}
		checked := make(map[*ast.Node]bool)
		report := func(node *ast.Node, messageID string, types []*checker.Type) {
			var description string
			names := c.describe(types)
			switch messageID {
			case "unsafeEnumAccess":
				description = "The computed key used here does not have a shared enum type with the expected enum %s."
			case "unsafeEnumArgument":
				description = "The argument passed here does not have a shared enum type with the expected enum %s."
			case "unsafeEnumAssertion":
				description = "The value asserted here does not have a shared enum type with the expected enum %s."
			case "unsafeEnumAssignment":
				description = "The value assigned here does not have a shared enum type with the expected enum %s."
			case "unsafeEnumMutation":
				description = "This mutation can produce a value outside of the expected enum %s."
			case "unsafeEnumReturn":
				description = "The value returned here does not have a shared enum type with the expected enum %s."
			}
			reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			if ast.IsParameterDeclaration(node) {
				reportRange = reportRange.WithPos(utils.TrimNodeTextRange(ctx.SourceFile, node.Name()).Pos())
			}
			ctx.ReportRange(reportRange, rule.RuleMessage{Id: messageID, Description: fmt.Sprintf(description, names)})
		}
		var markChecked func(*ast.Node)
		markChecked = func(node *ast.Node) {
			node = ast.SkipParentheses(node)
			checked[node] = true
			switch node.Kind {
			case ast.KindArrayLiteralExpression:
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					markChecked(element)
				}
			case ast.KindObjectLiteralExpression:
				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					markChecked(property)
				}
			case ast.KindConditionalExpression:
				expr := node.AsConditionalExpression()
				markChecked(expr.WhenTrue)
				markChecked(expr.WhenFalse)
			case ast.KindBinaryExpression:
				expr := node.AsBinaryExpression()
				switch expr.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
					markChecked(expr.Left)
					markChecked(expr.Right)
				}
			case ast.KindPropertyAssignment:
				markChecked(node.Initializer())
			case ast.KindShorthandPropertyAssignment:
				markChecked(node.Name())
			case ast.KindSpreadElement, ast.KindSpreadAssignment, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
				markChecked(node.Expression())
			}
		}
		var safeBitwise func(*ast.Node, *checker.Type) bool
		safeBitwise = func(node *ast.Node, receiver *checker.Type) bool {
			node = ast.SkipParentheses(node)
			safeOperand := func(operand *ast.Node) bool {
				return safeBitwise(operand, receiver) || !c.directMismatch(ctx.TypeChecker.GetTypeAtLocation(operand), receiver)
			}
			if ast.IsBinaryExpression(node) {
				expr := node.AsBinaryExpression()
				switch expr.OperatorToken.Kind {
				case ast.KindAmpersandToken, ast.KindCaretToken, ast.KindBarToken:
					return safeOperand(expr.Left) && safeOperand(expr.Right)
				}
			}
			if ast.IsPrefixUnaryExpression(node) {
				expr := node.AsPrefixUnaryExpression()
				return expr.Operator == ast.KindTildeToken && safeOperand(expr.Operand)
			}
			return false
		}
		unsafeAssignment := func(senderNode *ast.Node, receiver, sender *checker.Type) bool {
			if sender == nil {
				sender = ctx.TypeChecker.GetTypeAtLocation(senderNode)
			}
			return c.mismatch(sender, receiver) && !safeBitwise(senderNode, receiver)
		}
		checkAssignment := func(receiver *checker.Type, sender, reporting *ast.Node, messageID string, senderType *checker.Type) {
			sender = ast.SkipParentheses(sender)
			if ast.IsArrayLiteralExpression(sender) || ast.IsObjectLiteralExpression(sender) || !unsafeAssignment(sender, receiver, senderType) {
				return
			}
			report(reporting, messageID, []*checker.Type{receiver})
			markChecked(sender)
		}
		checkArguments := func(node *ast.Node, args []*ast.Node, tagged bool) {
			signature := newFunctionSignature(c, node)
			if signature == nil {
				return
			}
			if tagged {
				signature.next()
			}
			for _, argument := range args {
				if ast.IsSpreadElement(argument) {
					spreadType := ctx.TypeChecker.GetTypeAtLocation(argument.Expression())
					if checker.IsTupleType(spreadType) {
						var mismatches []*checker.Type
						for _, element := range c.typeArguments(spreadType) {
							if parameter := signature.next(); parameter != nil && c.mismatch(element, parameter) {
								mismatches = append(mismatches, parameter)
							}
						}
						if len(mismatches) > 0 {
							report(argument, "unsafeEnumArgument", mismatches)
						}
						if checker.TupleType_combinedFlags(spreadType.Target().AsTupleType())&checker.ElementFlagsVariable != 0 {
							signature.consumed = true
						}
						continue
					}
				}
				if parameter := signature.next(); parameter != nil {
					checkAssignment(parameter, argument, argument, "unsafeEnumArgument", nil)
				}
			}
		}
		checkReturn := func(sender, reporting *ast.Node) {
			function := utils.GetParentFunctionNode(sender)
			if function == nil {
				return
			}
			signature := ctx.TypeChecker.GetSignatureFromDeclaration(function)
			if signature == nil {
				return
			}
			receiver := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
			senderType := ctx.TypeChecker.GetTypeAtLocation(sender)
			if ast.IsAsyncFunction(function) {
				receiver = checker.Checker_getAwaitedType(ctx.TypeChecker, receiver)
				senderType = checker.Checker_getAwaitedType(ctx.TypeChecker, senderType)
			}
			if receiver != nil && senderType != nil {
				checkAssignment(receiver, sender, reporting, "unsafeEnumReturn", senderType)
			}
		}
		checkMutation := func(target, reporting *ast.Node) {
			t := ctx.TypeChecker.GetTypeAtLocation(target)
			if len(utils.GetEnumTypes(ctx.TypeChecker, c.constraint(t))) > 0 {
				report(reporting, "unsafeEnumMutation", []*checker.Type{t})
			}
		}
		checkInitializer := func(node *ast.Node) {
			if init := node.Initializer(); init != nil {
				checkAssignment(ctx.TypeChecker.GetTypeAtLocation(node.Name()), init, node, "unsafeEnumAssignment", nil)
			}
		}
		checkAssertion := func(node *ast.Node) {
			checkAssignment(ctx.TypeChecker.GetTypeAtLocation(node.Type()), node.Expression(), node, "unsafeEnumAssertion", nil)
		}
		return rule.RuleListeners{
			ast.KindVariableDeclaration: checkInitializer,
			ast.KindParameter:           checkInitializer,
			ast.KindBindingElement:      checkInitializer,
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				if init := node.AsShorthandPropertyAssignment().ObjectAssignmentInitializer; init != nil {
					checkAssignment(ctx.TypeChecker.GetTypeAtLocation(node.Name()), init, node, "unsafeEnumAssignment", nil)
				}
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				init := node.Initializer()
				if init == nil {
					return
				}
				var heritageTypes []*checker.Type
				name := node.Name()
				memberName := ""
				nameKnown := false
				if ast.IsComputedPropertyName(name) {
					t := ctx.TypeChecker.GetTypeAtLocation(name.Expression())
					if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) {
						memberName = fmt.Sprint(t.AsLiteralType().Value())
						nameKnown = true
					}
				} else if ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name) {
					memberName = name.Text()
					nameKnown = true
				}
				if nameKnown {
					if clauses := utils.GetHeritageClauses(node.Parent); clauses != nil {
						for _, clause := range clauses.Nodes {
							if ast.IsStatic(node) && clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
								continue
							}
							for _, heritage := range clause.AsHeritageClause().Types.Nodes {
								heritageNode := heritage
								if ast.IsStatic(node) {
									heritageNode = heritage.Expression()
								}
								if symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(heritageNode), memberName); symbol != nil {
									heritageTypes = append(heritageTypes, checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol))
								}
							}
						}
					}
				}
				if len(heritageTypes) > 0 && utils.Every(heritageTypes, func(t *checker.Type) bool { return unsafeAssignment(init, t, nil) }) {
					report(node, "unsafeEnumAssignment", heritageTypes)
					markChecked(init)
					return
				}
				checkAssignment(ctx.TypeChecker.GetTypeAtLocation(node), init, node, "unsafeEnumAssignment", nil)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, false) {
					return
				}
				expr := node.AsBinaryExpression()
				switch expr.OperatorToken.Kind {
				case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindCaretEqualsToken, ast.KindBarEqualsToken, ast.KindBarBarEqualsToken:
					checkAssignment(ctx.TypeChecker.GetTypeAtLocation(expr.Left), expr.Right, node, "unsafeEnumAssignment", nil)
				default:
					checkMutation(expr.Left, node)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()
				if expr.Operator == ast.KindPlusPlusToken || expr.Operator == ast.KindMinusMinusToken {
					checkMutation(expr.Operand, node)
				}
			},
			ast.KindPostfixUnaryExpression:  func(node *ast.Node) { checkMutation(node.AsPostfixUnaryExpression().Operand, node) },
			ast.KindAsExpression:            checkAssertion,
			ast.KindTypeAssertionExpression: checkAssertion,
			ast.KindSatisfiesExpression: func(node *ast.Node) {
				if !checked[node] {
					checkAssignment(ctx.TypeChecker.GetTypeAtLocation(node.Type()), node.Expression(), node, "unsafeEnumAssignment", nil)
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				if expr := node.Expression(); expr != nil {
					checkReturn(expr, node)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				if body := node.Body(); body != nil && !ast.IsBlock(body) {
					checkReturn(body, body)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) { checkArguments(node, node.AsCallExpression().Arguments.Nodes, false) },
			ast.KindNewExpression: func(node *ast.Node) {
				if args := node.AsNewExpression().Arguments; args != nil {
					checkArguments(node, args.Nodes, false)
				}
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				template := node.AsTaggedTemplateExpression().Template
				var args []*ast.Node
				if ast.IsTemplateExpression(template) {
					for _, span := range template.AsTemplateExpression().TemplateSpans.Nodes {
						args = append(args, span.Expression())
					}
				}
				checkArguments(node, args, true)
			},
			ast.KindJsxAttribute: func(node *ast.Node) {
				if init := node.Initializer(); init != nil && ast.IsJsxExpression(init) {
					if expr := init.Expression(); expr != nil {
						if receiver := checker.Checker_getContextualType(ctx.TypeChecker, expr, checker.ContextFlagsNone); receiver != nil {
							checkAssignment(receiver, expr, expr, "unsafeEnumAssignment", nil)
						}
					}
				}
			},
			rule.ListenerOnNotAllowPattern(ast.KindArrayLiteralExpression): func(node *ast.Node) {
				if checked[node] {
					return
				}
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					if element.Kind == ast.KindOmittedExpression {
						continue
					}
					if receiver := checker.Checker_getContextualType(ctx.TypeChecker, element, checker.ContextFlagsNone); receiver != nil {
						checkAssignment(receiver, element, element, "unsafeEnumAssignment", nil)
					}
				}
			},
			rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				if checked[node] {
					return
				}
				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					var sender, receiverNode *ast.Node
					switch property.Kind {
					case ast.KindSpreadAssignment:
						sender = property.Expression()
						receiverNode = node
					case ast.KindPropertyAssignment:
						sender = property.Initializer()
						receiverNode = sender
					case ast.KindShorthandPropertyAssignment:
						sender = property.Name()
						receiverNode = sender
					default:
						continue
					}
					if receiver := checker.Checker_getContextualType(ctx.TypeChecker, receiverNode, checker.ContextFlagsNone); receiver != nil {
						checkAssignment(receiver, sender, property, "unsafeEnumAssignment", nil)
					}
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				expr := node.AsElementAccessExpression()
				symbol := ctx.TypeChecker.GetSymbolAtLocation(expr.Expression)
				if symbol == nil {
					return
				}
				var receivers []*checker.Type
				for _, declaration := range symbol.Declarations {
					receivers = append(receivers, mappedKeyConstraintTypes(ctx.TypeChecker, declaration)...)
				}
				if len(receivers) == 0 {
					return
				}
				sender := ctx.TypeChecker.GetTypeAtLocation(expr.ArgumentExpression)
				if utils.Every(receivers, func(receiver *checker.Type) bool {
					return c.mismatch(sender, receiver) || !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, sender, receiver)
				}) {
					report(expr.ArgumentExpression, "unsafeEnumAccess", receivers)
				}
			},
		}
	},
}
