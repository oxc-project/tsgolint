package strict_boolean_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

type StrictBooleanExpressionsOptions struct {
	AllowAny                                               *bool
	AllowNullableBoolean                                   *bool
	AllowNullableNumber                                    *bool
	AllowNullableString                                    *bool
	AllowNullableEnum                                      *bool
	AllowNullableObject                                    *bool
	AllowString                                            *bool
	AllowNumber                                            *bool
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool
}

func buildUnexpectedAny() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAny",
		Description: "Unexpected any value in conditional. An explicit comparison or type cast is required.",
	}
}

func buildUnexpectedNullableBoolean() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNullableBoolean",
		Description: "Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.",
	}
}

func buildUnexpectedNullableString() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNullableString",
		Description: "Unexpected nullable string value in conditional. Please handle the nullish case explicitly.",
	}
}

func buildUnexpectedNullableNumber() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNullableNumber",
		Description: "Unexpected nullable number value in conditional. Please handle the nullish case explicitly.",
	}
}

func buildUnexpectedNullableObject() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNullableObject",
		Description: "Unexpected nullable object value in conditional. An explicit null check is required.",
	}
}

func buildUnexpectedNullish() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNullish",
		Description: "Unexpected nullish value in conditional. An explicit null check is required.",
	}
}

func buildUnexpectedString() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedString",
		Description: "Unexpected string value in conditional. An explicit empty string check is required.",
	}
}

func buildUnexpectedNumber() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNumber",
		Description: "Unexpected number value in conditional. An explicit zero/NaN check is required.",
	}
}

func buildUnexpectedObjectContext() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedObjectContext",
		Description: "Unexpected object value in conditional. The condition is always true.",
	}
}

func buildUnexpectedMixedCondition() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedMixedCondition",
		Description: "Unexpected mixed type in conditional. The constituent types do not have a best common type.",
	}
}

func buildNoStrictNullCheck() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "msgNoStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPredicateCannotBeAsync() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicateCannotBeAsync",
		Description: "Predicate function should not be 'async'; expected a boolean return type.",
	}
}

var traversedNodes = utils.Set[*ast.Node]{}

var StrictBooleanExpressionsRule = rule.Rule{
	Name: "strict-boolean-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(StrictBooleanExpressionsOptions)
		if !ok {
			opts = StrictBooleanExpressionsOptions{}
		}

		if opts.AllowAny == nil {
			opts.AllowAny = utils.Ref(false)
		}
		if opts.AllowNullableBoolean == nil {
			opts.AllowNullableBoolean = utils.Ref(false)
		}
		if opts.AllowNullableNumber == nil {
			opts.AllowNullableNumber = utils.Ref(false)
		}
		if opts.AllowNullableString == nil {
			opts.AllowNullableString = utils.Ref(false)
		}
		if opts.AllowNullableEnum == nil {
			opts.AllowNullableEnum = utils.Ref(false)
		}
		if opts.AllowNullableObject == nil {
			opts.AllowNullableObject = utils.Ref(true)
		}
		if opts.AllowString == nil {
			opts.AllowString = utils.Ref(true)
		}
		if opts.AllowNumber == nil {
			opts.AllowNumber = utils.Ref(true)
		}
		if opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(false)
		}

		compilerOptions := ctx.Program.Options()
		if !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing &&
			!utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks) {
			ctx.ReportRange(
				core.NewTextRange(0, 0),
				buildNoStrictNullCheck(),
			)
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				traverseNode(ctx, ifStmt.Expression, opts, true)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				traverseNode(ctx, whileStmt.Expression, opts, true)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				traverseNode(ctx, doStmt.Expression, opts, true)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt.Condition != nil {
					traverseNode(ctx, forStmt.Condition, opts, true)
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				condExpr := node.AsConditionalExpression()
				traverseNode(ctx, condExpr.Condition, opts, true)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if ast.IsLogicalExpression(node) && binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					traverseLogicalExpression(ctx, binExpr, opts, false)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()

				assertedArgument := findTruthinessAssertedArgument(ctx.TypeChecker, callExpr)
				if assertedArgument != nil {
					traverseNode(ctx, assertedArgument, opts, true)
				}

				if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) {
					if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
						arg := callExpr.Arguments.Nodes[0]
						if arg != nil && (arg.Kind == ast.KindArrowFunction || arg.Kind == ast.KindFunctionExpression || arg.Kind == ast.KindFunctionDeclaration) {
							if checker.GetFunctionFlags(arg)&checker.FunctionFlagsAsync != 0 {
								ctx.ReportNode(arg, buildPredicateCannotBeAsync())
								return
							}
							funcType := ctx.TypeChecker.GetTypeAtLocation(arg)
							signatures := ctx.TypeChecker.GetCallSignatures(funcType)
							for _, signature := range signatures {
								returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signature)
								typeFlags := checker.Type_flags(returnType)
								if typeFlags&checker.TypeFlagsTypeParameter != 0 {
									constraint := ctx.TypeChecker.GetConstraintOfTypeParameter(returnType)
									if constraint != nil {
										returnType = constraint
									}
								}

								if returnType != nil && !isBooleanType(returnType) {
									checkCondition(ctx, node, returnType, opts)
								}
							}
						}
					}
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				unaryExpr := node.AsPrefixUnaryExpression()
				if unaryExpr.Operator == ast.KindExclamationToken {
					traverseNode(ctx, unaryExpr.Operand, opts, true)
				}
			},
		}
	},
}

func findTruthinessAssertedArgument(typeChecker *checker.Checker, callExpr *ast.CallExpression) *ast.Node {
	var checkableArguments []*ast.Node
	for _, argument := range callExpr.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			break
		}
		checkableArguments = append(checkableArguments, argument)
	}
	if len(checkableArguments) == 0 {
		return nil
	}
	node := callExpr.AsNode()
	signature := typeChecker.GetResolvedSignature(node)
	if signature == nil {
		return nil
	}
	firstTypePredicateResult := typeChecker.GetTypePredicateOfSignature(signature)
	if firstTypePredicateResult == nil {
		return nil
	}
	if !(checker.TypePredicate_kind(firstTypePredicateResult) == checker.TypePredicateKindAssertsIdentifier &&
		checker.TypePredicate_t(firstTypePredicateResult) == nil) {
		return nil
	}
	parameterIndex := checker.TypePredicate_parameterIndex(firstTypePredicateResult)
	if int(parameterIndex) >= len(checkableArguments) {
		return nil
	}
	return checkableArguments[parameterIndex]
}

func checkNode(ctx rule.RuleContext, node *ast.Node, opts StrictBooleanExpressionsOptions) {
	nodeType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
	checkCondition(ctx, node, nodeType, opts)
}

func traverseLogicalExpression(ctx rule.RuleContext, binExpr *ast.BinaryExpression, opts StrictBooleanExpressionsOptions, isCondition bool) {
	traverseNode(ctx, binExpr.Left, opts, true)
	traverseNode(ctx, binExpr.Right, opts, isCondition)
}

func traverseNode(ctx rule.RuleContext, node *ast.Node, opts StrictBooleanExpressionsOptions, isCondition bool) {
	if traversedNodes.Has(node) {
		return
	}
	traversedNodes.Add(node)

	if node.Kind == ast.KindBinaryExpression {
		binExpr := node.AsBinaryExpression()
		if ast.IsLogicalExpression(node) && binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
			traverseLogicalExpression(ctx, binExpr, opts, isCondition)
			return
		}
	}

	if !isCondition {
		return
	}

	checkNode(ctx, node, opts)
}

// Type analysis types
type typeVariant int

const (
	typeVariantNullish typeVariant = iota
	typeVariantBoolean
	typeVariantString
	typeVariantNumber
	typeVariantBigInt
	typeVariantObject
	typeVariantAny
	typeVariantUnknown
	typeVariantNever
	typeVariantMixed
	typeVariantGeneric
)

type typeInfo struct {
	variant        typeVariant
	isNullable     bool
	isTruthy       bool
	types          []*checker.Type
	isUnion        bool
	isIntersection bool
	isEnum         bool
}

func analyzeType(typeChecker *checker.Checker, t *checker.Type) typeInfo {
	info := typeInfo{
		types: []*checker.Type{t},
	}

	if utils.IsUnionType(t) {
		info.isUnion = true
		parts := utils.UnionTypeParts(t)
		variants := make(map[typeVariant]bool)

		for _, part := range parts {
			partInfo := analyzeTypePart(typeChecker, part)
			variants[partInfo.variant] = true
			if partInfo.variant == typeVariantNullish {
				info.isNullable = true
			}
			if partInfo.isEnum {
				info.isEnum = true
			}
			info.isTruthy = partInfo.isTruthy
		}

		if len(variants) == 1 {
			for v := range variants {
				info.variant = v
			}
		} else if len(variants) == 2 && info.isNullable {
			for v := range variants {
				if v != typeVariantNullish {
					info.variant = v
					break
				}
			}
		} else {
			info.variant = typeVariantMixed
		}

		return info
	}

	if utils.IsIntersectionType(t) {
		info.isIntersection = true
		types := t.Types()
		isBoolean := false
		for _, t2 := range types {
			if analyzeTypePart(typeChecker, t2).variant == typeVariantBoolean {
				isBoolean = true
				break
			}
		}
		if isBoolean {
			info.variant = typeVariantBoolean
		} else {
			info.variant = typeVariantObject
		}
		return info
	}

	return analyzeTypePart(typeChecker, t)
}

func analyzeTypePart(_ *checker.Checker, t *checker.Type) typeInfo {
	info := typeInfo{}
	flags := checker.Type_flags(t)

	if flags&checker.TypeFlagsTypeParameter != 0 {
		info.variant = typeVariantGeneric
		return info
	}

	if flags&checker.TypeFlagsAny != 0 {
		info.variant = typeVariantAny
		return info
	}

	if flags&checker.TypeFlagsUnknown != 0 {
		info.variant = typeVariantUnknown
		return info
	}

	if flags&checker.TypeFlagsNever != 0 {
		info.variant = typeVariantNever
		return info
	}

	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		info.variant = typeVariantNullish
		return info
	}

	if flags&(checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsBooleanLike) != 0 {
		if utils.IsTrueLiteralType(t) {
			info.isTruthy = true
		}
		info.variant = typeVariantBoolean
		return info
	}

	if flags&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral) != 0 {
		info.variant = typeVariantString
		if flags&checker.TypeFlagsStringLiteral != 0 && t.AsLiteralType().Value() != "" {
			info.isTruthy = true
		}
		return info
	}

	if flags&(checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral) != 0 {
		info.variant = typeVariantNumber
		if flags&checker.TypeFlagsNumberLiteral != 0 && t.AsLiteralType().Value() != 0 {
			info.isTruthy = true
		}
		return info
	}

	if flags&(checker.TypeFlagsEnum|checker.TypeFlagsEnumLiteral) != 0 {
		if flags&checker.TypeFlagsStringLiteral != 0 {
			info.variant = typeVariantString
		} else {
			info.variant = typeVariantNumber
		}
		info.isEnum = true
		return info
	}

	if flags&(checker.TypeFlagsBigInt|checker.TypeFlagsBigIntLiteral) != 0 {
		info.variant = typeVariantBigInt
		if flags&checker.TypeFlagsBigIntLiteral != 0 && t.AsLiteralType().Value() != 0 {
			info.isTruthy = true
		}
		return info
	}

	if flags&(checker.TypeFlagsESSymbol|checker.TypeFlagsUniqueESSymbol) != 0 {
		info.variant = typeVariantObject
		return info
	}

	if flags&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive) != 0 {
		info.variant = typeVariantObject
		return info
	}

	info.variant = typeVariantMixed
	return info
}

func checkCondition(ctx rule.RuleContext, node *ast.Node, t *checker.Type, opts StrictBooleanExpressionsOptions) {
	info := analyzeType(ctx.TypeChecker, t)

	switch info.variant {
	case typeVariantAny, typeVariantUnknown, typeVariantGeneric:
		if !*opts.AllowAny {
			ctx.ReportNode(node, buildUnexpectedAny())
		}
		return
	case typeVariantNever:
		return
	case typeVariantNullish:
		ctx.ReportNode(node, buildUnexpectedNullish())
	case typeVariantString:
		// Known edge case: truthy primitives and nullish values are always valid boolean expressions
		if *opts.AllowString && info.isTruthy {
			return
		}

		if info.isNullable {
			if info.isEnum {
				if !*opts.AllowNullableEnum {
					ctx.ReportNode(node, buildUnexpectedNullableString())
				}
			} else {
				if !*opts.AllowNullableString {
					ctx.ReportNode(node, buildUnexpectedNullableString())
				}
			}
		} else if !*opts.AllowString {
			ctx.ReportNode(node, buildUnexpectedString())
		}
	case typeVariantNumber:
		if *opts.AllowNumber && info.isTruthy {
			return
		}

		if info.isNullable {
			if info.isEnum {
				if !*opts.AllowNullableEnum {
					ctx.ReportNode(node, buildUnexpectedNullableNumber())
				}
			} else {
				if !*opts.AllowNullableNumber {
					ctx.ReportNode(node, buildUnexpectedNullableNumber())
				}
			}
		} else if !*opts.AllowNumber {
			ctx.ReportNode(node, buildUnexpectedNumber())
		}
	case typeVariantBoolean:
		if info.isNullable && !*opts.AllowNullableBoolean {
			ctx.ReportNode(node, buildUnexpectedNullableBoolean())
		}
	case typeVariantObject:
		if info.isNullable && !*opts.AllowNullableObject {
			ctx.ReportNode(node, buildUnexpectedNullableObject())
		} else if !info.isNullable {
			ctx.ReportNode(node, buildUnexpectedObjectContext())
		}
	case typeVariantMixed:
		ctx.ReportNode(node, buildUnexpectedMixedCondition())
	case typeVariantBigInt:
		if info.isNullable && !*opts.AllowNullableNumber {
			ctx.ReportNode(node, buildUnexpectedNullableNumber())
		} else if !info.isNullable && !*opts.AllowNumber {
			ctx.ReportNode(node, buildUnexpectedNumber())
		}
	}
}

func isBooleanType(t *checker.Type) bool {
	flags := checker.Type_flags(t)

	if flags&(checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0 {
		if utils.IsUnionType(t) {
			for _, part := range utils.UnionTypeParts(t) {
				partFlags := checker.Type_flags(part)
				if partFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
					return false
				}
			}
		}
		return true
	}

	return false
}
