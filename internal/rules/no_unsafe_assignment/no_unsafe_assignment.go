package no_unsafe_assignment

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func formatSenderType(senderType *checker.Type) string {
	if utils.IsIntrinsicErrorType(senderType) {
		return "error typed"
	}
	return "any"
}

func buildAnyAssignmentMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyAssignment",
		Description: fmt.Sprintf("Unsafe assignment of an %v value.", formatSenderType(sender)),
	}
}
func buildAnyAssignmentThisMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyAssignmentThis",
		Description: fmt.Sprintf("Unsafe assignment of an %v value. `this` is typed as `any`.\n", formatSenderType(sender)),
		Help:        "You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}
func buildUnsafeArrayPatternMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArrayPattern",
		Description: fmt.Sprintf("Unsafe array destructuring of an %v array value.", formatSenderType(sender)),
	}
}
func buildUnsafeArrayPatternFromTupleMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArrayPatternFromTuple",
		Description: fmt.Sprintf("Unsafe array destructuring of a tuple element with an %v value.", formatSenderType(sender)),
	}
}
func buildUnsafeArraySpreadMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArraySpread",
		Description: fmt.Sprintf("Unsafe spread of an %v value in an array.", formatSenderType(sender)),
	}
}
func buildUnsafeAssignmentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeAssignment",
		Description: "Unsafe assignment between incompatible types.",
	}
}

func buildAssignmentDiagnostic(
	primaryRange core.TextRange,
	senderRange core.TextRange,
	receiverRange core.TextRange,
	senderType string,
	receiverType string,
	message rule.RuleMessage,
) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range:   primaryRange,
		Message: message,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("Assigned value has type `%s`.", senderType),
				Range: senderRange,
			},
			{
				Label: fmt.Sprintf("Target expects type `%s`.", receiverType),
				Range: receiverRange,
			},
		},
	}
}

func buildThisAssignmentDiagnostic(
	primaryRange core.TextRange,
	thisRange core.TextRange,
	receiverRange core.TextRange,
	thisType string,
	receiverType string,
	message rule.RuleMessage,
) rule.RuleDiagnostic {
	diagnostic := buildAssignmentDiagnostic(
		primaryRange,
		thisRange,
		receiverRange,
		thisType,
		receiverType,
		message,
	)
	diagnostic.LabeledRanges[0].Label = fmt.Sprintf("`this` has type `%s`.", thisType)
	return diagnostic
}

func buildDestructureDiagnostic(
	receiverRange core.TextRange,
	senderRange core.TextRange,
	senderType string,
	unsafeType string,
	message rule.RuleMessage,
) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range:   receiverRange,
		Message: message,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("Destructured source provides type `%s`.", senderType),
				Range: senderRange,
			},
			{
				Label: fmt.Sprintf("This binding receives type `%s`.", unsafeType),
				Range: receiverRange,
			},
		},
	}
}

func buildArraySpreadDiagnostic(
	spreadRange core.TextRange,
	valueRange core.TextRange,
	valueType string,
	message rule.RuleMessage,
) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Range:   spreadRange,
		Message: message,
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("Spread value has type `%s`.", valueType),
				Range: valueRange,
			},
		},
	}
}

func diagnosticTypeText(typeChecker *checker.Checker, t *checker.Type) string {
	if utils.IsIntrinsicErrorType(t) {
		return "error"
	}
	return typeChecker.TypeToString(t)
}

func assignmentRelationRange(sourceFile *ast.SourceFile, receiverNode, senderNode *ast.Node) core.TextRange {
	s := scanner.GetScannerForSourceFile(sourceFile, receiverNode.End())
	var colonRange core.TextRange
	for s.Token() != ast.KindEndOfFile && s.TokenRange().Pos() < senderNode.End() {
		switch s.Token() {
		case ast.KindEqualsToken:
			return s.TokenRange()
		case ast.KindColonToken:
			colonRange = s.TokenRange()
		}
		if s.TokenRange().Pos() >= senderNode.Pos() {
			break
		}
		s.Scan()
	}
	if colonRange != (core.TextRange{}) {
		return colonRange
	}
	return utils.TrimNodeTextRange(sourceFile, receiverNode)
}

func localTargetRange(sourceFile *ast.SourceFile, receiverNode, typeAnnotationNode *ast.Node) core.TextRange {
	if typeAnnotationNode != nil && ast.GetSourceFileOfNode(typeAnnotationNode) == sourceFile {
		return utils.TrimNodeTextRange(sourceFile, typeAnnotationNode)
	}
	return utils.TrimNodeTextRange(sourceFile, receiverNode)
}

type comparisonType uint8

const (
	/** Do no assignment comparison */
	comparisonTypeNone comparisonType = iota
	/** Use the receiver's type for comparison */
	comparisonTypeBasic
	/** Use the sender's contextual type for comparison */
	comparisonTypeContextual
)

func canSkipSenderTypeCheck(node *ast.Node, compType comparisonType) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		expr := node.AsPrefixUnaryExpression()
		return (expr.Operator == ast.KindPlusToken || expr.Operator == ast.KindMinusToken) && ast.SkipParentheses(expr.Operand).Kind == ast.KindNumericLiteral
	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true
	case ast.KindObjectLiteralExpression:
		return !utils.Some(node.AsObjectLiteralExpression().Properties.Nodes, ast.IsSpreadAssignment)
	case ast.KindArrayLiteralExpression:
		return compType == comparisonTypeNone && !utils.Some(node.AsArrayLiteralExpression().Elements.Nodes, ast.IsSpreadElement)
	default:
		return false
	}
}

var NoUnsafeAssignmentRule = rule.Rule{
	Name: "no-unsafe-assignment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		isNoImplicitThis := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.NoImplicitThis,
		)

		reportDestructure := func(receiverNode, senderNode *ast.Node, senderType *checker.Type, unsafeType string, message rule.RuleMessage) {
			ctx.ReportDiagnostic(buildDestructureDiagnostic(
				utils.TrimNodeTextRange(ctx.SourceFile, receiverNode),
				utils.TrimNodeTextRange(ctx.SourceFile, senderNode),
				diagnosticTypeText(ctx.TypeChecker, senderType),
				unsafeType,
				message,
			))
		}

		var checkArrayDestructure func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool
		var checkObjectDestructure func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool

		// returns true if the assignment reported
		checkObjectDestructure = func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool {
			propertySymbols := checker.Checker_getPropertiesOfType(ctx.TypeChecker, senderType)
			if propertySymbols == nil {
				return false
			}
			properties := make(map[string]*checker.Type, len(propertySymbols))
			for _, property := range propertySymbols {
				properties[property.Name] = ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, senderNode)
			}

			checkObjectProperty := func(propertyKey *ast.Node, propertyValue *ast.Node) bool {
				var key string
				if !ast.IsComputedPropertyName(propertyKey) {
					key = propertyKey.Text()
				} else if ast.IsLiteralExpression(propertyKey.Expression()) {
					key = propertyKey.Expression().Text()
				} else {
					// can't figure out the name, so skip it
					return false
				}

				senderType, ok := properties[key]
				if !ok {
					return false
				}

				// check for the any type first so we can handle {x: {y: z}} = {x: any}
				if utils.IsTypeAnyType(senderType) {
					// TODO(port): why object reported with "array" message?
					reportDestructure(propertyValue, senderNode, senderType, diagnosticTypeText(ctx.TypeChecker, senderType), buildUnsafeArrayPatternFromTupleMessage(senderType))
					return true
				} else if ast.IsArrayBindingPattern(propertyValue) || ast.IsArrayLiteralExpression(propertyValue) {
					return checkArrayDestructure(
						propertyValue,
						senderType,
						senderNode,
					)
				} else if ast.IsObjectBindingPattern(propertyValue) || ast.IsObjectLiteralExpression(propertyValue) {
					return checkObjectDestructure(
						propertyValue,
						senderType,
						senderNode,
					)
				}
				return false
			}

			didReport := false
			if ast.IsObjectLiteralExpression(receiverNode) {
				for _, receiverProperty := range receiverNode.AsObjectLiteralExpression().Properties.Nodes {
					if ast.IsSpreadAssignment(receiverProperty) {
						// don't bother checking rest
						continue
					}

					if (ast.IsPropertyAssignment(receiverProperty) && checkObjectProperty(receiverProperty.Name(), receiverProperty.Initializer())) || (ast.IsShorthandPropertyAssignment(receiverProperty) && checkObjectProperty(receiverProperty.Name(), receiverProperty.Name())) {
						didReport = true
					}
				}
			} else if ast.IsObjectBindingPattern(receiverNode) {
				for _, receiverProperty := range receiverNode.AsBindingPattern().Elements.Nodes {
					property := receiverProperty.AsBindingElement()
					if property.DotDotDotToken != nil {
						// don't bother checking rest
						continue
					}

					propertyKey := property.PropertyName
					if propertyKey == nil {
						propertyKey = property.Name()
					}

					if checkObjectProperty(propertyKey, property.Name()) {
						didReport = true
					}
				}
			}

			return didReport
		}

		// returns true if the assignment reported
		checkObjectDestructureHelper := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
		) bool {
			if !ast.IsObjectBindingPattern(receiverNode) && !ast.IsObjectLiteralExpression(receiverNode) {
				return false
			}

			senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)

			return checkObjectDestructure(receiverNode, senderType, senderNode)
		}

		// returns true if the assignment reported
		checkArrayDestructure = func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool {
			// any array
			// const [x] = ([] as any[]);
			if utils.IsTypeAnyArrayType(senderType, ctx.TypeChecker) {
				reportDestructure(receiverNode, senderNode, senderType, "any", buildUnsafeArrayPatternMessage(senderType))
				return false
			}

			if !checker.IsTupleType(senderType) {
				return true
			}

			tupleElements := checker.Checker_getTypeArguments(ctx.TypeChecker, senderType)

			checkArrayElement := func(receiverElement *ast.Node, receiverIndex int) bool {
				if receiverElement == nil {
					return false
				}
				if receiverIndex >= len(tupleElements) {
					return false
				}
				senderType := tupleElements[receiverIndex]

				// check for the any type first so we can handle [[[x]]] = [any]
				if utils.IsTypeAnyType(senderType) {
					reportDestructure(receiverElement, senderNode, senderType, diagnosticTypeText(ctx.TypeChecker, senderType), buildUnsafeArrayPatternFromTupleMessage(senderType))
					return true
				} else if ast.IsArrayBindingPattern(receiverElement) || ast.IsArrayLiteralExpression(receiverElement) {
					return checkArrayDestructure(
						receiverElement,
						senderType,
						senderNode,
					)
				} else if ast.IsObjectBindingPattern(receiverElement) || ast.IsObjectLiteralExpression(receiverElement) {
					return checkObjectDestructure(
						receiverElement,
						senderType,
						senderNode,
					)
				}

				return false
			}

			// tuple with any
			// const [x] = [1 as any];
			didReport := false
			if ast.IsArrayLiteralExpression(receiverNode) {
				for receiverIndex, receiverElement := range receiverNode.AsArrayLiteralExpression().Elements.Nodes {
					if ast.IsSpreadElement(receiverElement) {
						// don't handle rests as they're not a 1:1 assignment
						continue
					}

					if checkArrayElement(receiverElement, receiverIndex) {
						didReport = true
					}
				}
			} else if ast.IsArrayBindingPattern(receiverNode) {
				for receiverIndex, receiverElement := range receiverNode.AsBindingPattern().Elements.Nodes {
					elem := receiverElement.AsBindingElement()
					if elem.DotDotDotToken != nil {
						// don't handle rests as they're not a 1:1 assignment
						continue
					}

					if checkArrayElement(receiverElement.Name(), receiverIndex) {
						// TODO(port): in original rule didReport was reassigned every time. isn't it a bug?
						didReport = true
					}
				}
			}

			return didReport
		}

		// returns true if the assignment reported
		checkArrayDestructureHelper := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
		) bool {
			if !ast.IsArrayBindingPattern(receiverNode) && !ast.IsArrayLiteralExpression(receiverNode) {
				return false
			}

			senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)

			return checkArrayDestructure(receiverNode, senderType, senderNode)
		}

		// returns true if the assignment reported
		checkAssignment := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
			typeAnnotationNode *ast.Node,
			primaryRange core.TextRange,
			compType comparisonType,
		) bool {
			// Fast path: return early when we know that the sender definitely cannot have an `any` type,
			// because it is syntactically impossible given the sender's node kind.
			if canSkipSenderTypeCheck(senderNode, compType) {
				return false
			}

			senderType := ctx.TypeChecker.GetTypeAtLocation(senderNode)

			getReceiverType := func() *checker.Type {
				if compType == comparisonTypeContextual {
					if receiverType := utils.GetContextualType(ctx.TypeChecker, senderNode); receiverType != nil {
						return receiverType
					}
					if receiverType := utils.GetContextualType(ctx.TypeChecker, receiverNode); receiverType != nil {
						return receiverType
					}
				}
				return ctx.TypeChecker.GetTypeAtLocation(receiverNode)
			}

			if utils.IsTypeAnyType(senderType) {
				receiverType := getReceiverType()
				receiverRange := localTargetRange(ctx.SourceFile, receiverNode, typeAnnotationNode)
				setInferredTargetLabel := func(diagnostic *rule.RuleDiagnostic) {
					if compType == comparisonTypeNone {
						diagnostic.LabeledRanges[1].Label = fmt.Sprintf(
							"Target is inferred as `%s`.",
							diagnosticTypeText(ctx.TypeChecker, receiverType),
						)
					}
				}

				// handle cases when we assign any ==> unknown.
				if utils.IsTypeUnknownType(receiverType) {
					return false
				}

				if !isNoImplicitThis {
					// `var foo = this`
					thisExpression := utils.GetThisExpression(senderNode)
					if thisExpression != nil {
						thisType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
						if utils.IsTypeAnyType(thisType) {
							diagnostic := buildThisAssignmentDiagnostic(
								primaryRange,
								utils.TrimNodeTextRange(ctx.SourceFile, thisExpression),
								receiverRange,
								diagnosticTypeText(ctx.TypeChecker, thisType),
								diagnosticTypeText(ctx.TypeChecker, receiverType),
								buildAnyAssignmentThisMessage(senderType),
							)
							setInferredTargetLabel(&diagnostic)
							ctx.ReportDiagnostic(diagnostic)
							return true
						}
					}
				}

				diagnostic := buildAssignmentDiagnostic(
					primaryRange,
					utils.TrimNodeTextRange(ctx.SourceFile, senderNode),
					receiverRange,
					diagnosticTypeText(ctx.TypeChecker, senderType),
					diagnosticTypeText(ctx.TypeChecker, receiverType),
					buildAnyAssignmentMessage(senderType),
				)
				setInferredTargetLabel(&diagnostic)
				ctx.ReportDiagnostic(diagnostic)
				return true
			}

			if compType == comparisonTypeNone {
				return false
			}
			if !checker.IsNonDeferredTypeReference(senderType) {
				return false
			}

			receiverType := getReceiverType()

			receiver, sender, unsafe := utils.IsUnsafeAssignment(
				senderType,
				receiverType,
				ctx.TypeChecker,
				senderNode,
			)
			if !unsafe {
				return false
			}

			ctx.ReportDiagnostic(buildAssignmentDiagnostic(
				primaryRange,
				utils.TrimNodeTextRange(ctx.SourceFile, senderNode),
				localTargetRange(ctx.SourceFile, receiverNode, typeAnnotationNode),
				diagnosticTypeText(ctx.TypeChecker, sender),
				diagnosticTypeText(ctx.TypeChecker, receiver),
				buildUnsafeAssignmentMessage(),
			))
			return true
		}

		getComparisonType := func(
			nodeWithTypeAnnotation *ast.Node,
		) comparisonType {
			if nodeWithTypeAnnotation.Type() != nil {
				// if there's a type annotation, we can do a comparison
				return comparisonTypeBasic
			}
			// no type annotation means the variable's type will just be inferred, thus equal
			return comparisonTypeNone
		}

		checkAssignmentFull := func(id *ast.Node, init *ast.Node, typeAnnotationNode *ast.Node, primaryRange core.TextRange) {
			if id == nil || init == nil {
				return
			}
			didReport := checkAssignment(
				id,
				init,
				typeAnnotationNode,
				primaryRange,
				// the variable already has some form of a type to compare against
				comparisonTypeBasic,
			)

			if !didReport {
				didReport = checkArrayDestructureHelper(id, init)
			}
			if !didReport {
				checkObjectDestructureHelper(id, init)
			}
		}

		return rule.RuleListeners{
			// ESTree PropertyDefinition, AccessorProperty
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				initializer := node.Initializer()
				if initializer == nil {
					return
				}
				checkAssignment(
					node.Name(),
					initializer,
					node.Type(),
					assignmentRelationRange(ctx.SourceFile, node.Name(), initializer),
					getComparisonType(node),
				)
			},

			// ESTree AssignmentExpression, AssignmentPattern
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, true) {
					return
				}

				expr := node.AsBinaryExpression()
				checkAssignmentFull(
					expr.Left,
					expr.Right,
					nil,
					utils.TrimNodeTextRange(ctx.SourceFile, expr.OperatorToken),
				)
			},

			// ESTree AssignmentPattern
			ast.KindBindingElement: func(node *ast.Node) {
				if initializer := node.Initializer(); initializer != nil {
					checkAssignmentFull(node.Name(), initializer, node.Type(), assignmentRelationRange(ctx.SourceFile, node.Name(), initializer))
				}
			},
			// ESTree AssignmentPattern
			ast.KindParameter: func(node *ast.Node) {
				if initializer := node.Initializer(); initializer != nil {
					checkAssignmentFull(node.Name(), initializer, node.Type(), assignmentRelationRange(ctx.SourceFile, node.Name(), initializer))
				}
			},
			// ESTree AssignmentPattern
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				assignment := node.AsShorthandPropertyAssignment()
				if initializer := assignment.ObjectAssignmentInitializer; initializer != nil {
					checkAssignmentFull(assignment.Name(), initializer, nil, assignmentRelationRange(ctx.SourceFile, assignment.Name(), initializer))
				}
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				init := node.Initializer()
				if init == nil {
					return
				}

				id := node.Name()
				didReport := checkAssignment(
					id,
					init,
					node.Type(),
					assignmentRelationRange(ctx.SourceFile, id, init),
					getComparisonType(node),
				)

				if !didReport {
					didReport = checkArrayDestructureHelper(id, init)
				}
				if !didReport {
					checkObjectDestructureHelper(id, init)
				}
			},

			// object pattern props are checked via assignments
			rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				for _, node := range node.AsObjectLiteralExpression().Properties.Nodes {
					var init *ast.Node
					if ast.IsPropertyAssignment(node) {
						init = node.Initializer()
					} else if ast.IsShorthandPropertyAssignment(node) {
						init = node.Name()
					} else {
						continue
					}

					if init == nil {
						return
					}
					init = ast.SkipParentheses(init)

					if ast.IsAssignmentExpression(init, false) {
						// node.value.type === AST_NODE_TYPES.TSEmptyBodyFunctionExpression
						// handled by other selector
						return
					}

					checkAssignment(
						node.Name(),
						init,
						nil,
						assignmentRelationRange(ctx.SourceFile, node.Name(), init),
						comparisonTypeContextual,
					)
				}
			},

			rule.ListenerOnNotAllowPattern(ast.KindArrayLiteralExpression): func(node *ast.Node) {
				for _, node := range node.AsArrayLiteralExpression().Elements.Nodes {
					if !ast.IsSpreadElement(node) {
						continue
					}

					restType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
					if utils.IsTypeAnyType(restType) || utils.IsTypeAnyArrayType(restType, ctx.TypeChecker) {
						nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
						ctx.ReportDiagnostic(buildArraySpreadDiagnostic(
							core.NewTextRange(nodeRange.Pos(), nodeRange.Pos()+3),
							utils.TrimNodeTextRange(ctx.SourceFile, node.Expression()),
							diagnosticTypeText(ctx.TypeChecker, restType),
							buildUnsafeArraySpreadMessage(restType),
						))
					}
				}
			},

			ast.KindJsxAttribute: func(node *ast.Node) {
				init := node.Initializer()
				if init == nil || init.Kind != ast.KindJsxExpression {
					return
				}

				expr := init.AsJsxExpression().Expression
				if expr == nil {
					return
				}

				checkAssignment(
					node.Name(),
					expr,
					nil,
					assignmentRelationRange(ctx.SourceFile, node.Name(), expr),
					comparisonTypeContextual,
				)
			},
		}
	},
}
