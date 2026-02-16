package strict_boolean_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildConditionErrorNumberMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNumber",
		Description: "Unexpected number value in conditional. A number can be falsy (0, NaN) or truthy.",
	}
}

func buildConditionErrorStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorString",
		Description: "Unexpected string value in conditional. A string can be falsy (empty string) or truthy.",
	}
}

func buildConditionErrorObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorObject",
		Description: "Unexpected object value in conditional. An object is always truthy.",
	}
}

func buildConditionErrorNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullish",
		Description: "Unexpected nullish value in conditional. The expression is always falsy.",
	}
}

func buildConditionErrorOtherMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorOther",
		Description: "Unexpected value in conditional. A union of different types has inconsistent truthiness.",
	}
}

func buildConditionErrorNullableBooleanMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableBoolean",
		Description: "Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.",
	}
}

func buildConditionErrorNullableObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableObject",
		Description: "Unexpected nullable object value in conditional. Please handle the nullish case explicitly.",
	}
}

func buildConditionErrorNullableStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableString",
		Description: "Unexpected nullable string value in conditional. Please handle the nullish and empty string cases explicitly.",
	}
}

func buildConditionErrorNullableNumberMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableNumber",
		Description: "Unexpected nullable number value in conditional. Please handle the nullish and zero cases explicitly.",
	}
}

func buildConditionErrorNullableEnumMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableEnum",
		Description: "Unexpected nullable enum value in conditional. Please handle the nullish and falsy enum cases explicitly.",
	}
}

func buildConditionErrorAnyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorAny",
		Description: "Unexpected any value in conditional. Use a more specific type to ensure type safety.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPredicateCannotBeAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicateCannotBeAsync",
		Description: "Predicate function should not be 'async'; expected a boolean return type.",
	}
}

var StrictBooleanExpressionsRule = rule.Rule{
	Name: "strict-boolean-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[StrictBooleanExpressionsOptions](options, "strict-boolean-expressions")

		compilerOptions := ctx.Program.Options()
		if !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks) {
			ctx.ReportRange(
				core.NewTextRange(0, 0),
				buildNoStrictNullCheckMessage(),
			)
		}

		traversedNodes := utils.Set[*ast.Node]{}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				traverseNode(ctx, ifStmt.Expression, opts, &traversedNodes, true)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				traverseNode(ctx, whileStmt.Expression, opts, &traversedNodes, true)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				traverseNode(ctx, doStmt.Expression, opts, &traversedNodes, true)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt.Condition != nil {
					traverseNode(ctx, forStmt.Condition, opts, &traversedNodes, true)
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				condExpr := node.AsConditionalExpression()
				traverseNode(ctx, condExpr.Condition, opts, &traversedNodes, true)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if ast.IsLogicalExpression(node) && binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					traverseLogicalExpression(ctx, binExpr, opts, &traversedNodes, false)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()

				assertedArgument := findTruthinessAssertedArgument(ctx.TypeChecker, callExpr)
				if assertedArgument != nil {
					traverseNode(ctx, assertedArgument, opts, &traversedNodes, true)
				}

				if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) {
					if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
						arg := callExpr.Arguments.Nodes[0]
						if arg == nil {
							return
						}
						isFunction := arg.Kind&(ast.KindArrowFunction|ast.KindFunctionExpression|ast.KindFunctionDeclaration) != 0
						if isFunction && checker.GetFunctionFlags(arg)&checker.FunctionFlagsAsync != 0 {
							ctx.ReportNode(arg, buildPredicateCannotBeAsyncMessage())
							return
						}
						funcType := ctx.TypeChecker.GetTypeAtLocation(arg)
						signatures := ctx.TypeChecker.GetCallSignatures(funcType)
						var types []*checker.Type
						for _, signature := range signatures {
							returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signature)
							typeFlags := checker.Type_flags(returnType)
							if typeFlags&checker.TypeFlagsTypeParameter != 0 {
								constraint := ctx.TypeChecker.GetConstraintOfTypeParameter(returnType)
								if constraint != nil {
									returnType = constraint
								}
							}

							types = append(types, utils.UnionTypeParts(returnType)...)
						}
						checkCondition(ctx, node, types, opts)
					}
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				unaryExpr := node.AsPrefixUnaryExpression()
				if unaryExpr.Operator == ast.KindExclamationToken {
					traverseNode(ctx, unaryExpr.Operand, opts, &traversedNodes, true)
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

	calleeType := typeChecker.GetTypeAtLocation(callExpr.Expression)
	if calleeType == nil {
		return nil
	}

	unionTypes := utils.UnionTypeParts(calleeType)
	isUnionType := len(unionTypes) > 1

	node := callExpr.AsNode()
	signature := typeChecker.GetResolvedSignature(node)

	if signature == nil {
		if !isUnionType {
			return nil
		}

		for _, t := range unionTypes {
			signatures := typeChecker.GetCallSignatures(t)
			for _, sig := range signatures {
				typePredicate := typeChecker.GetTypePredicateOfSignature(sig)
				if typePredicate != nil &&
					checker.TypePredicate_kind(typePredicate) == checker.TypePredicateKindAssertsIdentifier &&
					checker.TypePredicate_t(typePredicate) == nil {
					parameterIndex := checker.TypePredicate_parameterIndex(typePredicate)
					if int(parameterIndex) < len(checkableArguments) {
						return checkableArguments[parameterIndex]
					}
				}
			}
		}
		return nil
	}

	firstTypePredicateResult := typeChecker.GetTypePredicateOfSignature(signature)
	if firstTypePredicateResult == nil {
		if !isUnionType {
			return nil
		}

		for _, t := range unionTypes {
			signatures := typeChecker.GetCallSignatures(t)
			for _, sig := range signatures {
				typePredicate := typeChecker.GetTypePredicateOfSignature(sig)
				if typePredicate != nil &&
					checker.TypePredicate_kind(typePredicate) == checker.TypePredicateKindAssertsIdentifier &&
					checker.TypePredicate_t(typePredicate) == nil {
					parameterIndex := checker.TypePredicate_parameterIndex(typePredicate)
					if int(parameterIndex) < len(checkableArguments) {
						return checkableArguments[parameterIndex]
					}
				}
			}
		}
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
	checkCondition(ctx, node, utils.UnionTypeParts(nodeType), opts)
}

func traverseLogicalExpression(ctx rule.RuleContext, binExpr *ast.BinaryExpression, opts StrictBooleanExpressionsOptions, traversedNodes *utils.Set[*ast.Node], isCondition bool) {
	traverseNode(ctx, binExpr.Left, opts, traversedNodes, true)
	traverseNode(ctx, binExpr.Right, opts, traversedNodes, isCondition)
}

func traverseNode(ctx rule.RuleContext, node *ast.Node, opts StrictBooleanExpressionsOptions, traversedNodes *utils.Set[*ast.Node], isCondition bool) {
	if traversedNodes.Has(node) {
		return
	}
	traversedNodes.Add(node)

	if node.Kind == ast.KindParenthesizedExpression {
		traverseNode(ctx, node.AsParenthesizedExpression().Expression, opts, traversedNodes, isCondition)
		return
	}

	if node.Kind == ast.KindBinaryExpression {
		binExpr := node.AsBinaryExpression()
		if ast.IsLogicalExpression(node) && binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
			traverseLogicalExpression(ctx, binExpr, opts, traversedNodes, isCondition)
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

func analyzeTypeParts(typeChecker *checker.Checker, types []*checker.Type) typeInfo {
	info := typeInfo{
		isUnion: len(types) > 1,
		types:   types,
	}
	variants := make(map[typeVariant]bool)

	metNotTruthy := false

	for _, part := range info.types {
		partInfo := analyzeTypePart(typeChecker, part)
		variants[partInfo.variant] = true
		if partInfo.variant == typeVariantNullish {
			info.isNullable = true
		}
		if partInfo.isEnum {
			info.isEnum = true
		}
		if partInfo.variant == typeVariantBoolean || partInfo.variant == typeVariantNumber || partInfo.variant == typeVariantString || partInfo.variant == typeVariantBigInt {
			if metNotTruthy {
				continue
			}
			info.isTruthy = partInfo.isTruthy
			metNotTruthy = !partInfo.isTruthy
		}
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

func analyzeTypePart(_typeChecker *checker.Checker, t *checker.Type) typeInfo {
	info := typeInfo{}
	flags := checker.Type_flags(t)

	if utils.IsIntersectionType(t) {
		info.isIntersection = true
		types := t.Types()
		isBoolean := false
		for _, t2 := range types {
			if analyzeTypePart(_typeChecker, t2).variant == typeVariantBoolean {
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
		if t.AsLiteralType().String() == "true" {
			info.isTruthy = true
		}
		info.variant = typeVariantBoolean
		return info
	}

	if flags&(checker.TypeFlagsEnum|checker.TypeFlagsEnumLiteral|checker.TypeFlagsEnumLike) != 0 {
		if flags&checker.TypeFlagsStringLiteral != 0 {
			info.variant = typeVariantString
		} else {
			info.variant = typeVariantNumber
		}
		info.isEnum = true
		return info
	}

	if flags&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral|checker.TypeFlagsStringLike) != 0 {
		info.variant = typeVariantString
		if t.IsStringLiteral() {
			literal := t.AsLiteralType()
			if literal != nil && literal.Value() != "" {
				info.isTruthy = true
			}
		}
		return info
	}

	if flags&(checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral|checker.TypeFlagsNumberLike) != 0 {
		info.variant = typeVariantNumber
		if t.IsNumberLiteral() {
			literal := t.AsLiteralType()
			if literal != nil && literal.String() != "0" {
				info.isTruthy = true
			}
		}
		return info
	}

	if flags&(checker.TypeFlagsBigInt|checker.TypeFlagsBigIntLiteral) != 0 {
		info.variant = typeVariantBigInt
		if t.IsBigIntLiteral() {
			literal := t.AsLiteralType()
			if literal != nil && literal.String() != "0" {
				info.isTruthy = true
			}
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

func checkCondition(ctx rule.RuleContext, node *ast.Node, types []*checker.Type, opts StrictBooleanExpressionsOptions) {
	info := analyzeTypeParts(ctx.TypeChecker, types)

	switch info.variant {
	case typeVariantAny, typeVariantUnknown, typeVariantGeneric:
		if !opts.AllowAny {
			ctx.ReportNode(node, buildConditionErrorAnyMessage())
		}
		return
	case typeVariantNever:
		return
	case typeVariantNullish:
		ctx.ReportNode(node, buildConditionErrorNullishMessage())
	case typeVariantString:
		// Known edge case: truthy primitives and nullish values are always valid boolean expressions
		if opts.AllowString && info.isTruthy {
			return
		}

		if info.isNullable {
			if info.isEnum {
				if !opts.AllowNullableEnum {
					ctx.ReportNode(node, buildConditionErrorNullableEnumMessage())
				}
			} else {
				if !opts.AllowNullableString {
					ctx.ReportNode(node, buildConditionErrorNullableStringMessage())
				}
			}
		} else if !opts.AllowString {
			ctx.ReportNode(node, buildConditionErrorStringMessage())
		}
	case typeVariantNumber:
		// Known edge case: truthy primitives and nullish values are always valid boolean expressions
		if opts.AllowNumber && info.isTruthy {
			return
		}

		if info.isNullable {
			if info.isEnum {
				if !opts.AllowNullableEnum {
					ctx.ReportNode(node, buildConditionErrorNullableEnumMessage())
				}
			} else {
				if !opts.AllowNullableNumber {
					ctx.ReportNode(node, buildConditionErrorNullableNumberMessage())
				}
			}
		} else if !opts.AllowNumber {
			ctx.ReportNode(node, buildConditionErrorNumberMessage())
		}
	case typeVariantBoolean:
		// Known edge case: truthy primitives and nullish values are always valid boolean expressions
		if info.isTruthy {
			return
		}

		if info.isNullable && !opts.AllowNullableBoolean {
			ctx.ReportNode(node, buildConditionErrorNullableBooleanMessage())
		}
	case typeVariantObject:
		if info.isNullable && !opts.AllowNullableObject {
			ctx.ReportNode(node, buildConditionErrorNullableObjectMessage())
		} else if !info.isNullable {
			ctx.ReportNode(node, buildConditionErrorObjectMessage())
		}
	case typeVariantMixed:
		if info.isEnum {
			if info.isNullable && !opts.AllowNullableEnum {
				ctx.ReportNode(node, buildConditionErrorNullableEnumMessage())
			}
			return
		}
		ctx.ReportNode(node, buildConditionErrorOtherMessage())
	case typeVariantBigInt:
		if info.isNullable && !opts.AllowNullableNumber {
			ctx.ReportNode(node, buildConditionErrorNullableNumberMessage())
		} else if !info.isNullable && !opts.AllowNumber {
			ctx.ReportNode(node, buildConditionErrorNumberMessage())
		}
	}
}
