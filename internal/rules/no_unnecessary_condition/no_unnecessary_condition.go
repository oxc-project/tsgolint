// Package no_unnecessary_condition implements the no-unnecessary-condition rule.
//
// This rule prevents unnecessary conditions in TypeScript code by detecting expressions
// that are always truthy, always falsy, or comparing values that have no overlap.
//
// The rule checks:
// - Conditional expressions (if, while, for, ternary operators)
// - Logical operators (&&, ||, !)
// - Nullish coalescing operators (??, ??=)
// - Optional chaining (?.)
// - Comparison operators (===, !==, ==, !=, <, >, <=, >=)
// - Type predicates and type guards
//
// This implementation is based on the @typescript-eslint/no-unnecessary-condition rule:
// https://typescript-eslint.io/rules/no-unnecessary-condition/
package no_unnecessary_condition

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildAlwaysTruthyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthy",
		Description: "Unnecessary conditional, value is always truthy.",
	}
}

func buildAlwaysFalsyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsy",
		Description: "Unnecessary conditional, value is always falsy.",
	}
}

func buildNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "never",
		Description: "Unnecessary conditional, value is `never`.",
	}
}

func buildAlwaysTruthyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthyFunc",
		Description: "This callback should return a conditional, but return is always truthy.",
	}
}

func buildAlwaysFalsyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsyFunc",
		Description: "This callback should return a conditional, but return is always falsy.",
	}
}

func buildNeverNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverNullish",
		Description: "Unnecessary optional chain on a non-nullish value.",
	}
}

func buildNeverOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverOptionalChain",
		Description: "Unnecessary optional chain on a non-nullish value.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildLiteralBinaryExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparisonBetweenLiteralTypes",
		Description: "Unnecessary comparison between literal values.",
	}
}

func buildNoOverlapDiagnostic(leftType string, leftRange core.TextRange, rightType string, rightRange core.TextRange) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		Message: rule.RuleMessage{
			Id:          "noOverlapBooleanExpression",
			Description: "This condition will always return the same value since the types have no overlap.",
		},
		LabeledRanges: []rule.RuleLabeledRange{
			{
				Label: fmt.Sprintf("Type: %v", leftType),
				Range: leftRange,
			},
			{
				Label: fmt.Sprintf("Type: %v", rightType),
				Range: rightRange,
			},
		},
	}
}

func buildAlwaysNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysNullish",
		Description: "Unnecessary conditional, value is always nullish.",
	}
}

func buildTypeGuardAlreadyIsTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeGuardAlreadyIsType",
		Description: "Type predicate is unnecessary as the parameter type already satisfies the predicate.",
	}
}

// isIndeterminateType checks if a type cannot be determined at compile time.
//
// Indeterminate types include:
// - any: explicitly typed as any
// - unknown: could be anything
// - type parameters: generic types like T, K
// - indexed access types: types like T[K]
// - index types: types like keyof T
//
// For these types, we cannot determine their truthiness, nullishness, or overlap
// with other types at compile time, so we conservatively avoid reporting them.
func isIndeterminateType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	// Note: We don't include TypeFlagsIndexedAccess because TypeScript resolves
	// indexed access types like T[K] to concrete types (e.g., number, string)
	// TypeFlagsIndex is for the index type operator (keyof T), which is indeterminate
	return flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIndex) != 0
}

// isAlwaysNullishType checks if a type is always null, undefined, or void.
//
// Returns true for types that can only be nullish values:
// - null
// - undefined
// - void (treated as undefined at runtime)
//
// Returns false for:
// - Non-nullish types (string, number, object, etc.)
// - Unions containing non-nullish types (string | null, number | undefined)
//
// Note: This is different from isNullishType which returns true if a type
// CAN BE nullish (including unions like string | null).
func isAlwaysNullishType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
}

func toStaticValue(t *checker.Type) (any, bool) { //nolint:unparam // the value return is reserved for follow-up checks; current callers only need the static-ness signal
	if t == nil {
		return nil, false
	}

	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		return isTrueLiteralTypeValue(t)
	}
	if flags&checker.TypeFlagsUndefined != 0 {
		return "undefined", true
	}
	if flags&checker.TypeFlagsNull != 0 {
		return nil, true
	}

	if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
		if literal := t.AsLiteralType(); literal != nil {
			return literal.Value(), true
		}
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
		if literal := t.AsLiteralType(); literal != nil {
			return literal.String(), true
		}
	}
	if flags&checker.TypeFlagsBigIntLiteral != 0 && t.IsBigIntLiteral() {
		if literal := t.AsLiteralType(); literal != nil {
			return literal.String(), true
		}
	}

	return nil, false
}

func isTrueLiteralTypeValue(t *checker.Type) (bool, bool) {
	if t == nil || checker.Type_flags(t)&checker.TypeFlagsBooleanLiteral == 0 {
		return false, false
	}

	if utils.IsIntrinsicType(t) {
		switch t.AsIntrinsicType().IntrinsicName() {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}

	if literal := t.AsLiteralType(); literal != nil {
		switch literal.String() {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}

	return false, false
}

var NoUnnecessaryConditionRule = rule.Rule{
	Name: "no-unnecessary-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[NoUnnecessaryConditionOptions](options, "no-unnecessary-condition")

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		noUncheckedIndexedAccess := compilerOptions.NoUncheckedIndexedAccess.IsTrue()
		loopConditionMode := normalizeAllowConstantLoopConditions(opts.AllowConstantLoopConditions)

		resolveIndexedAccessType := func(t *checker.Type) *checker.Type {
			if t == nil || checker.Type_flags(t)&checker.TypeFlagsIndexedAccess == 0 {
				return t
			}

			indexedAccess := t.AsIndexedAccessType()
			objectType := checker.IndexedAccessType_objectType(indexedAccess)
			indexType := checker.IndexedAccessType_indexType(indexedAccess)
			if objectType == nil || indexType == nil {
				return t
			}

			if constrainedObject := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, objectType); constrainedObject != nil {
				objectType = constrainedObject
			}
			if constrainedIndex := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, indexType); constrainedIndex != nil {
				indexType = constrainedIndex
			}

			if checker.Type_flags(indexType)&checker.TypeFlagsStringLiteral != 0 && indexType.IsStringLiteral() {
				if literal := indexType.AsLiteralType(); literal != nil {
					if propertyName, ok := literal.Value().(string); ok {
						if propType := checker.Checker_getTypeOfPropertyOfType(ctx.TypeChecker, objectType, propertyName); propType != nil {
							return propType
						}
					}
				}
			}

			if checker.Type_flags(indexType)&checker.TypeFlagsNumberLiteral != 0 && indexType.IsNumberLiteral() {
				if literal := indexType.AsLiteralType(); literal != nil {
					if propType := checker.Checker_getTypeOfPropertyOfType(ctx.TypeChecker, objectType, literal.String()); propType != nil {
						return propType
					}
				}
			}

			indexParts := utils.UnionTypeParts(indexType)
			if len(indexParts) != 0 && slices.ContainsFunc(indexParts, func(part *checker.Type) bool {
				return checker.Type_flags(part)&checker.TypeFlagsStringLike != 0
			}) {
				if stringIndexType := ctx.TypeChecker.GetStringIndexType(objectType); stringIndexType != nil {
					return stringIndexType
				}
			}

			if len(indexParts) != 0 && slices.ContainsFunc(indexParts, func(part *checker.Type) bool {
				return checker.Type_flags(part)&checker.TypeFlagsNumberLike != 0
			}) {
				if numberIndexType := ctx.TypeChecker.GetNumberIndexType(objectType); numberIndexType != nil {
					return numberIndexType
				}
			}

			return t
		}

		getResolvedType := func(node *ast.Node) *checker.Type {
			nodeType := ctx.TypeChecker.GetTypeAtLocation(node)
			if nodeType == nil {
				return nil
			}

			constraintType, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, nodeType)
			if isTypeParameter && constraintType == nil {
				return nil
			}
			if isTypeParameter {
				return resolveIndexedAccessType(constraintType)
			}
			return resolveIndexedAccessType(nodeType)
		}

		nodeIsArrayType := func(node *ast.Node) bool {
			nodeType := getResolvedType(node)
			if nodeType == nil {
				return false
			}

			for _, part := range utils.UnionTypeParts(nodeType) {
				if checker.Checker_isArrayType(ctx.TypeChecker, part) {
					return true
				}
			}

			return false
		}

		nodeIsTupleType := func(node *ast.Node) bool {
			nodeType := getResolvedType(node)
			if nodeType == nil {
				return false
			}

			for _, part := range utils.UnionTypeParts(nodeType) {
				if checker.IsTupleType(part) {
					return true
				}
			}

			return false
		}

		isArrayIndexExpression := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			node = ast.SkipParentheses(node)
			if !ast.IsElementAccessExpression(node) {
				return false
			}

			elemAccess := node.AsElementAccessExpression()
			if elemAccess.ArgumentExpression == nil {
				return false
			}

			return nodeIsArrayType(elemAccess.Expression) ||
				(nodeIsTupleType(elemAccess.Expression) &&
					ast.SkipParentheses(elemAccess.ArgumentExpression).Kind != ast.KindNumericLiteral)
		}

		isConditionalAlwaysNecessary := func(t *checker.Type) bool {
			for _, part := range utils.UnionTypeParts(t) {
				flags := checker.Type_flags(part)
				if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeVariable) != 0 {
					return true
				}
			}

			return false
		}

		// Type check predicates for declarative, readable code

		isNeverType := func(flags checker.TypeFlags) bool {
			return flags&checker.TypeFlagsNever != 0
		}

		isTrueLiteralType := func(t *checker.Type) bool {
			value, ok := isTrueLiteralTypeValue(t)
			return ok && value
		}

		isIndexedAccessFlags := func(flags checker.TypeFlags) bool {
			return flags&checker.TypeFlagsIndexedAccess != 0
		}

		isPropertyAccess := func(node *ast.Node) bool {
			return node != nil && node.Kind == ast.KindPropertyAccessExpression
		}

		isElementAccess := func(node *ast.Node) bool {
			return node != nil && node.Kind == ast.KindElementAccessExpression
		}

		isCallExpr := func(node *ast.Node) bool {
			return node != nil && node.Kind == ast.KindCallExpression
		}

		// Helper to check if an expression has optional chaining
		var hasOptionalChain func(*ast.Node) bool
		hasOptionalChain = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			n = ast.SkipParentheses(n)
			// Check if this node has optional chaining
			if isPropertyAccess(n) && n.AsPropertyAccessExpression().QuestionDotToken != nil {
				return true
			}
			if isElementAccess(n) && n.AsElementAccessExpression().QuestionDotToken != nil {
				return true
			}
			if isCallExpr(n) && n.AsCallExpression().QuestionDotToken != nil {
				return true
			}
			// Check in the expression chain
			if isPropertyAccess(n) {
				return hasOptionalChain(n.AsPropertyAccessExpression().Expression)
			}
			if isElementAccess(n) {
				return hasOptionalChain(n.AsElementAccessExpression().Expression)
			}
			if isCallExpr(n) {
				return hasOptionalChain(n.AsCallExpression().Expression)
			}
			return false
		}

		// constrainIndexedAccessType resolves an indexed access type via its
		// base constraint. Returns (resolvedType, shouldSkip).
		//
		// When the indexed access can't be resolved to a concrete constraint,
		// callers should skip conservatively.
		constrainIndexedAccessType := func(t *checker.Type) (*checker.Type, bool) {
			if t == nil || !isIndexedAccessFlags(checker.Type_flags(t)) {
				return t, false
			}
			constraintType := resolveIndexedAccessType(t)
			if constraintType == nil || isIndexedAccessFlags(checker.Type_flags(constraintType)) {
				return nil, true
			}
			return constraintType, false
		}

		getPropertyNameFromLiteralType := func(t *checker.Type) (string, bool) {
			if t == nil {
				return "", false
			}

			flags := checker.Type_flags(t)
			if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
				literal := t.AsLiteralType()
				if literal != nil {
					if value, ok := literal.Value().(string); ok {
						return value, true
					}
				}
			}

			if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
				literal := t.AsLiteralType()
				if literal != nil {
					return literal.String(), true
				}
			}

			return "", false
		}

		var isNullablePropertyType func(objType *checker.Type, propertyType *checker.Type) bool
		isNullablePropertyType = func(objType *checker.Type, propertyType *checker.Type) bool {
			if objType == nil || propertyType == nil {
				return false
			}

			if utils.IsUnionType(propertyType) {
				for _, part := range propertyType.Types() {
					if isNullablePropertyType(objType, part) {
						return true
					}
				}
				return false
			}

			if propertyName, ok := getPropertyNameFromLiteralType(propertyType); ok {
				propType := checker.Checker_getTypeOfPropertyOfType(ctx.TypeChecker, objType, propertyName)
				if propType != nil {
					return isNullishType(propType)
				}
			}

			propertyTypeName := utils.GetTypeName(ctx.TypeChecker, propertyType)
			for _, info := range checker.Checker_getIndexInfosOfType(ctx.TypeChecker, objType) {
				if utils.GetTypeName(ctx.TypeChecker, info.KeyType()) == propertyTypeName {
					return true
				}
			}

			return false
		}

		isNullableElementAccessExpression := func(elemAccess *ast.ElementAccessExpression) bool {
			if elemAccess == nil || elemAccess.ArgumentExpression == nil {
				return false
			}

			objectType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.Expression)
			if objectType == nil {
				return false
			}

			propertyType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.ArgumentExpression)
			if propertyType == nil {
				return false
			}

			return isNullablePropertyType(objectType, propertyType)
		}

		// Helper: Get the effective type of a call expression.
		// For plain calls, prefer the resolved call-site type so overloaded APIs like
		// react-hook-form's getValues(path) use the selected overload.
		// For calls that are themselves part of an optional chain, fall back to the
		// callee's return type so we ignore undefined introduced only by short-circuiting
		// earlier chain segments.
		getCallReturnType := func(callExpr *ast.Node) *checker.Type {
			if callExpr == nil || callExpr.Kind != ast.KindCallExpression {
				return nil
			}

			call := callExpr.AsCallExpression()
			if call.QuestionDotToken == nil && !hasOptionalChain(call.Expression) {
				if callType := getResolvedType(callExpr); callType != nil {
					return callType
				}
			}

			if resolvedSignature := checker.Checker_getResolvedSignature(ctx.TypeChecker, callExpr, nil, checker.CheckModeNormal); resolvedSignature != nil {
				if returnType := ctx.TypeChecker.GetReturnTypeOfSignature(resolvedSignature); returnType != nil {
					return returnType
				}
			}

			funcType := getResolvedType(call.Expression)
			if funcType == nil {
				return nil
			}

			nonNullishFunc := removeNullishFromType(funcType)
			if nonNullishFunc == nil {
				return nil
			}

			// If it's a union type of functions, check each part's return type
			// e.g., (() => undefined) | (() => number) should check if any returns nullish
			if utils.IsUnionType(nonNullishFunc) {
				// For union of functions, check if any function returns a nullish type
				// If so, the optional chaining result can be nullish
				parts := nonNullishFunc.Types()
				for _, part := range parts {
					sigs := ctx.TypeChecker.GetCallSignatures(part)
					if len(sigs) > 0 {
						retType := ctx.TypeChecker.GetReturnTypeOfSignature(sigs[0])
						if retType != nil && isNullishType(retType) {
							// At least one function returns nullish, so use full expression type
							// which includes all possible return types
							return ctx.TypeChecker.GetTypeAtLocation(callExpr)
						}
					}
				}
				// No function returns nullish, get first signature's return type
				signatures := ctx.TypeChecker.GetCallSignatures(nonNullishFunc)
				if len(signatures) > 0 {
					return ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
				}
				return nil
			}

			signatures := ctx.TypeChecker.GetCallSignatures(nonNullishFunc)
			if len(signatures) == 0 {
				return nil
			}

			return ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
		}

		// Helper: Get property type from a base type given a property access expression
		getPropertyTypeFromBase := func(baseType *checker.Type, propAccess *ast.Node) *checker.Type {
			if baseType == nil || propAccess == nil {
				return nil
			}
			if propAccess.Kind != ast.KindPropertyAccessExpression {
				return nil
			}

			nonNullishBase := removeNullishFromType(baseType)
			if nonNullishBase == nil {
				return nil
			}

			pa := propAccess.AsPropertyAccessExpression()
			nameNode := pa.Name()
			if nameNode == nil {
				return ctx.TypeChecker.GetTypeAtLocation(propAccess)
			}

			propName := ast.GetTextOfPropertyName(nameNode)
			if propName == "" {
				return ctx.TypeChecker.GetTypeAtLocation(propAccess)
			}

			// Try to get the property directly first
			prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, propName)
			if prop != nil {
				return ctx.TypeChecker.GetTypeOfSymbol(prop)
			}

			// For mapped types, try the apparent type which may have the property
			apparentType := checker.Checker_getApparentType(ctx.TypeChecker, nonNullishBase)
			if apparentType != nil && apparentType != nonNullishBase {
				prop = checker.Checker_getPropertyOfType(ctx.TypeChecker, apparentType, propName)
				if prop != nil {
					return ctx.TypeChecker.GetTypeOfSymbol(prop)
				}
			}

			// Property doesn't exist as a declared property, check index signatures
			// For index signatures and mapped types, behavior depends on noUncheckedIndexedAccess:
			// - WITH noUncheckedIndexedAccess: index access returns T | undefined, be conservative
			// - WITHOUT noUncheckedIndexedAccess: index access returns T, use the actual type
			stringIndexType := ctx.TypeChecker.GetStringIndexType(nonNullishBase)
			if stringIndexType == nil && apparentType != nil {
				// Try the apparent type's index signature
				stringIndexType = ctx.TypeChecker.GetStringIndexType(apparentType)
			}

			// For mapped types with template literal keys (e.g., Lowercase<string>),
			// GetStringIndexType may return nil. Check if the type is a mapped type
			// and look at all its properties to find one that matches.
			if stringIndexType == nil {
				objectFlags := checker.Type_objectFlags(nonNullishBase)
				if objectFlags&checker.ObjectFlagsMapped != 0 {
					// This is a mapped type - get all its properties
					properties := checker.Checker_getPropertiesOfType(ctx.TypeChecker, nonNullishBase)
					for _, p := range properties {
						if p.Name == propName {
							// Found the property - check if it's optional
							if p.Flags&ast.SymbolFlagsOptional == 0 {
								// Non-optional property - use its type
								return ctx.TypeChecker.GetTypeOfSymbol(p)
							}
							// Optional property - let caller handle it
							return nil
						}
					}
					// For mapped types where we couldn't find the property,
					// return nil to signal we couldn't determine the property type.
					// The caller can then handle this case specially for mapped types.
					return nil
				}
			}

			if stringIndexType != nil {
				// With noUncheckedIndexedAccess, TypeScript adds undefined to index accesses
				// So we should let GetTypeAtLocation handle it to be conservative
				if ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
					return nil
				}
				// Without noUncheckedIndexedAccess, return the actual index type
				// This allows us to flag unnecessary optional chains on non-optional mapped types
				return stringIndexType
			}

			// For non-mapped types, fall back to GetTypeAtLocation
			return ctx.TypeChecker.GetTypeAtLocation(propAccess)
		}

		// Helper: Get type from property/element access on call expression result
		// e.g., foo?.().bar or foo().baz
		getTypeFromCallProperty := func(callExpr *ast.Node, accessExpr *ast.Node) *checker.Type {
			returnType := getCallReturnType(callExpr)
			if returnType == nil {
				// For union types, use GetTypeAtLocation
				returnType = ctx.TypeChecker.GetTypeAtLocation(callExpr)
				if returnType == nil {
					return nil
				}
			}

			nonNullishReturn := removeNullishFromType(returnType)
			if nonNullishReturn == nil {
				return nil
			}

			if isPropertyAccess(accessExpr) {
				return getPropertyTypeFromBase(nonNullishReturn, accessExpr)
			}

			// ElementAccessExpression
			return ctx.TypeChecker.GetTypeAtLocation(accessExpr)
		}

		isCallExpressionNullableOriginFromCallee := func(callExpr *ast.CallExpression) bool {
			if callExpr == nil {
				return false
			}

			prevType := getResolvedType(callExpr.Expression)
			if prevType == nil || !utils.IsUnionType(prevType) {
				return false
			}

			isOwnNullable := false
			for _, part := range prevType.Types() {
				signatures := ctx.TypeChecker.GetCallSignatures(part)
				for _, sig := range signatures {
					returnType := ctx.TypeChecker.GetReturnTypeOfSignature(sig)
					if returnType != nil && isNullishType(returnType) {
						isOwnNullable = true
						break
					}
				}
				if isOwnNullable {
					break
				}
			}

			return !isOwnNullable && isNullishType(prevType)
		}

		isMemberExpressionNullableOriginFromObject := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			node = ast.SkipParentheses(node)

			var objectExpr *ast.Node
			var propertyType *checker.Type
			var propertyName string
			var isComputed bool

			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				objectExpr = propAccess.Expression
				nameNode := propAccess.Name()
				if nameNode == nil {
					return false
				}
				propertyName = ast.GetTextOfPropertyName(nameNode)
			case ast.KindElementAccessExpression:
				elemAccess := node.AsElementAccessExpression()
				objectExpr = elemAccess.Expression
				isComputed = true
				if elemAccess.ArgumentExpression == nil {
					return false
				}
				propertyType = getResolvedType(elemAccess.ArgumentExpression)
			default:
				return false
			}

			prevType := getResolvedType(objectExpr)
			if prevType == nil || !utils.IsUnionType(prevType) {
				return false
			}

			isOwnNullable := false
			for _, part := range prevType.Types() {
				if isNullishType(part) {
					continue
				}

				if isComputed {
					if propertyType != nil && isNullablePropertyType(part, propertyType) {
						isOwnNullable = true
						break
					}
					continue
				}

				if propertyName == "" {
					continue
				}

				propType := checker.Checker_getTypeOfPropertyOfType(ctx.TypeChecker, part, propertyName)
				if propType != nil {
					if isNullishType(propType) {
						isOwnNullable = true
						break
					}
					continue
				}

				for _, info := range checker.Checker_getIndexInfosOfType(ctx.TypeChecker, part) {
					if utils.GetTypeName(ctx.TypeChecker, info.KeyType()) != "string" {
						continue
					}
					if noUncheckedIndexedAccess || isNullishType(info.ValueType()) {
						isOwnNullable = true
						break
					}
				}
				if isOwnNullable {
					break
				}
			}

			return !isOwnNullable && isNullishType(prevType)
		}

		isOptionableExpression := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			node = ast.SkipParentheses(node)

			nodeType := getResolvedType(node)
			if nodeType == nil {
				return false
			}

			if isConditionalAlwaysNecessary(nodeType) {
				return true
			}

			isOwnNullable := true
			switch node.Kind {
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				isOwnNullable = !isMemberExpressionNullableOriginFromObject(node)
			case ast.KindCallExpression:
				isOwnNullable = !isCallExpressionNullableOriginFromCallee(node.AsCallExpression())
			}

			return isOwnNullable && isNullishType(nodeType)
		}

		// checkOptionalChain validates optional chaining (?.) to detect unnecessary usage.
		//
		// Optional chaining is unnecessary when the expression being accessed is never nullish.
		// This function handles the complexity of chained optional access like foo?.bar?.baz.
		//
		// Examples:
		//   const obj: { foo: string } = { foo: "hello" }
		//   obj?.foo  // unnecessary - obj is never nullish
		//
		//   const obj: { foo: { bar: string } } | null = getObj()
		//   obj?.foo?.bar  // first ?. is fine, but second ?. is unnecessary
		//                  // because when obj exists, obj.foo is never nullish
		//
		// Algorithm:
		// 1. Extract the expression being accessed (e.g., for foo?.bar, extract foo)
		// 2. For chained access (foo?.bar?.baz), we need to check the intermediate type:
		//    - Get the type of foo (excluding nullish parts)
		//    - Check if foo.bar can be nullish (not foo?.bar)
		// 3. For simple access (foo?.bar), check if foo can be nullish
		// 4. Allow indeterminate types (any, unknown, T, T[K]) since we can't determine nullishness
		checkOptionalChain := func(node *ast.Node) {
			var expression *ast.Node
			var hasQuestionDot bool

			// Extract the expression and check if this is optional chaining
			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				expression = propAccess.Expression
				hasQuestionDot = propAccess.QuestionDotToken != nil
			case ast.KindElementAccessExpression:
				elemAccess := node.AsElementAccessExpression()
				expression = elemAccess.Expression
				hasQuestionDot = elemAccess.QuestionDotToken != nil
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				expression = callExpr.Expression
				hasQuestionDot = callExpr.QuestionDotToken != nil
			default:
				return
			}

			if !hasQuestionDot {
				return
			}

			// Check if we should skip the check due to unguarded element access
			// Rules:
			// 1. If expression itself is unguarded element access - skip (arr[0]?.foo)
			// 2. If expression uses optional chaining AND contains unguarded element access - skip (arr[0]?.foo?.bar)
			// 3. If expression doesn't use optional chaining - don't skip (arr[0].foo?.bar)

			// Helper: Check if expression or its chain contains unguarded element access
			var containsUnguardedElementAccess func(*ast.Node) bool
			containsUnguardedElementAccess = func(expr *ast.Node) bool {
				if expr == nil {
					return false
				}
				expr = ast.SkipParentheses(expr)

				// Check if this is an unguarded element access
				if isElementAccess(expr) {
					elemAccess := expr.AsElementAccessExpression()
					if elemAccess.QuestionDotToken == nil && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
						return true
					}
					// Check deeper
					return containsUnguardedElementAccess(elemAccess.Expression)
				}

				// Check deeper in property/call chains
				if isPropertyAccess(expr) {
					return containsUnguardedElementAccess(expr.AsPropertyAccessExpression().Expression)
				}
				if isCallExpr(expr) {
					return containsUnguardedElementAccess(expr.AsCallExpression().Expression)
				}

				return false
			}

			// Helper: Check if expression uses optional chaining
			var usesOptionalChaining func(*ast.Node) bool
			usesOptionalChaining = func(expr *ast.Node) bool {
				if expr == nil {
					return false
				}
				expr = ast.SkipParentheses(expr)

				if isPropertyAccess(expr) && expr.AsPropertyAccessExpression().QuestionDotToken != nil {
					return true
				}
				if isElementAccess(expr) && expr.AsElementAccessExpression().QuestionDotToken != nil {
					return true
				}
				if isCallExpr(expr) && expr.AsCallExpression().QuestionDotToken != nil {
					return true
				}
				return false
			}

			// Helper: Check if element access is a safe tuple access with literal index
			isSafeTupleAccess := func(elemAccess *ast.ElementAccessExpression) bool {
				if elemAccess == nil || elemAccess.ArgumentExpression == nil {
					return false
				}
				arg := ast.SkipParentheses(elemAccess.ArgumentExpression)
				// Check if argument is a numeric literal
				if arg.Kind == ast.KindNumericLiteral || arg.Kind == ast.KindFirstLiteralToken {
					// Check if base type is a tuple
					baseType := getResolvedType(elemAccess.Expression)
					if baseType != nil && checker.IsTupleType(baseType) {
						return true
					}
				}
				return false
			}

			expressionSkipped := ast.SkipParentheses(expression)

			// Rule 1: Expression itself is unguarded element access (but not safe tuple access)
			if isElementAccess(expressionSkipped) && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
				elemAccess := expressionSkipped.AsElementAccessExpression()
				if elemAccess.QuestionDotToken == nil {
					// Check if this is a safe tuple access with literal index
					if !isSafeTupleAccess(elemAccess) {
						return
					}
				}
			}

			// Rule 2: Expression uses optional chaining AND contains unguarded element access (but not safe tuple access)
			if usesOptionalChaining(expression) {
				// Check if contains unguarded element access that's not a safe tuple access
				var hasUnsafeElementAccess func(*ast.Node) bool
				hasUnsafeElementAccess = func(expr *ast.Node) bool {
					if expr == nil {
						return false
					}
					expr = ast.SkipParentheses(expr)

					if isElementAccess(expr) {
						elemAccess := expr.AsElementAccessExpression()
						if elemAccess.QuestionDotToken == nil && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
							// Check if this is safe tuple access
							if !isSafeTupleAccess(elemAccess) {
								return true
							}
						}
						return hasUnsafeElementAccess(elemAccess.Expression)
					}

					if isPropertyAccess(expr) {
						return hasUnsafeElementAccess(expr.AsPropertyAccessExpression().Expression)
					}
					if isCallExpr(expr) {
						return hasUnsafeElementAccess(expr.AsCallExpression().Expression)
					}

					return false
				}

				if hasUnsafeElementAccess(expression) {
					return
				}
			}

			if isOptionableExpression(expression) {
				return
			}

			// Check if the expression is itself an optional chain (chained access)
			// For foo?.bar?.baz, when checking the second ?.:
			//   - node is foo?.bar?.baz
			//   - expression is foo?.bar
			//   - baseExpression is foo
			var baseExpression *ast.Node
			var isChainedAccess bool

			switch expression.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := expression.AsPropertyAccessExpression()
				if propAccess.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = propAccess.Expression
				}
			case ast.KindElementAccessExpression:
				elemAccess := expression.AsElementAccessExpression()
				if elemAccess.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = elemAccess.Expression
				}
			case ast.KindCallExpression:
				callExpr := expression.AsCallExpression()
				if callExpr.QuestionDotToken != nil {
					isChainedAccess = true
					baseExpression = callExpr.Expression
				}
			}

			// Get the type that would result from the access (if it succeeds)
			// For optional chains, TypeScript gives us the type including | undefined
			// But we want to check if the property itself can be nullish, not including
			// the undefined from the optional chain short-circuit
			var exprType *checker.Type
			if isChainedAccess {
				// For chained access like foo?.bar?.baz, when checking the second ?.:
				// expression is foo?.bar
				// We want to check if bar can be nullish (from foo's perspective)

				// Get foo's type (non-nullish parts)
				baseType := getResolvedType(baseExpression)
				if baseType == nil {
					return
				}

				nonNullishBase := removeNullishFromType(baseType)
				if nonNullishBase == nil {
					return
				}

				// Get the type of bar from foo's (non-nullish) type
				// For PropertyAccessExpression, we can get the property name
				if isPropertyAccess(expression) {
					exprType = getPropertyTypeFromBase(nonNullishBase, expression)
					// If nil (couldn't resolve property directly), use GetTypeAtLocation
					if exprType == nil {
						exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
						// For chained access with optional chains, GetTypeAtLocation includes
						// short-circuit undefined. For non-optional mapped types, we should
						// remove this undefined. Check if the base type is a mapped type
						// and if so, remove nullish from the result.
						if exprType != nil && !ctx.Program.Options().NoUncheckedIndexedAccess.IsTrue() {
							objectFlags := checker.Type_objectFlags(nonNullishBase)
							if objectFlags&checker.ObjectFlagsMapped != 0 {
								// Check if the mapped type has the optional modifier (?)
								// If it does, the property genuinely can be undefined
								modifiers := checker.GetMappedTypeModifiers(nonNullishBase)
								if modifiers&checker.MappedTypeModifiersIncludeOptional == 0 {
									// Non-optional mapped type - remove the short-circuit undefined
									exprType = removeNullishFromType(exprType)
								}
							}
						}
					}
				} else if isElementAccess(expression) {
					// For element access, check if we're accessing with a literal key
					// e.g., foo?.[key] where key is 'bar' | 'foo'
					elemAccess := expression.AsElementAccessExpression()
					argExpr := elemAccess.ArgumentExpression
					if argExpr != nil {
						// Get the type of the key
						keyType := ctx.TypeChecker.GetTypeAtLocation(argExpr)
						if keyType != nil {
							// Check if the key is a string literal type or union of string literals
							keyFlags := checker.Type_flags(keyType)
							isLiteralKey := false
							var literalKeys []string

							if keyFlags&checker.TypeFlagsStringLiteral != 0 {
								// Single string literal
								isLiteralKey = true
								if keyType.IsStringLiteral() {
									lit := keyType.AsLiteralType()
									if lit != nil {
										literalKeys = append(literalKeys, lit.Value().(string))
									}
								}
							} else if utils.IsUnionType(keyType) {
								// Union of string literals
								allLiterals := true
								for _, part := range keyType.Types() {
									partFlags := checker.Type_flags(part)
									if partFlags&checker.TypeFlagsStringLiteral != 0 {
										if part.IsStringLiteral() {
											lit := part.AsLiteralType()
											if lit != nil {
												literalKeys = append(literalKeys, lit.Value().(string))
											}
										}
									} else {
										allLiterals = false
										break
									}
								}
								isLiteralKey = allLiterals
							}

							// If we have literal keys, check if all of them have non-nullish property types
							if isLiteralKey && len(literalKeys) > 0 {
								allNonNullish := true
								for _, key := range literalKeys {
									prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, key)
									if prop == nil {
										// Property doesn't exist, might be index signature
										allNonNullish = false
										break
									}
									propType := ctx.TypeChecker.GetTypeOfSymbol(prop)
									if propType == nil || isNullishType(propType) {
										allNonNullish = false
										break
									}
								}
								if allNonNullish {
									// All literal keys have non-nullish types
									// Get the actual property type for the first key (they're all non-nullish)
									prop := checker.Checker_getPropertyOfType(ctx.TypeChecker, nonNullishBase, literalKeys[0])
									if prop != nil {
										exprType = ctx.TypeChecker.GetTypeOfSymbol(prop)
									} else {
										exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
									}
								} else {
									exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
								}
							} else {
								// Not a literal key, use default behavior
								exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
							}
						} else {
							exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
						}
					} else {
						exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
					}
				} else if isCallExpr(expression) {
					// For call expressions in a chain, get the function's return type
					exprType = getCallReturnType(expression)
					if exprType == nil {
						// For union types, use GetTypeAtLocation
						exprType = getResolvedType(expression)
						if exprType == nil {
							return
						}
					}
				} else {
					// For other expression types, use the full type
					exprType = ctx.TypeChecker.GetTypeAtLocation(expression)
				}
			} else if isPropertyAccess(expression) || isElementAccess(expression) {
				// Handle property/element access on call expression result
				// e.g., foo?.().bar?.baz or foo?.bar?.().baz
				var innerExpr *ast.Node
				if isPropertyAccess(expression) {
					innerExpr = expression.AsPropertyAccessExpression().Expression
				} else {
					innerExpr = expression.AsElementAccessExpression().Expression
				}

				// Check if the inner expression is a call expression
				if innerExpr != nil && isCallExpr(innerExpr) {
					exprType = getTypeFromCallProperty(innerExpr, expression)
					if exprType == nil {
						return
					}
				} else {
					exprType = getResolvedType(expression)
				}
			} else {
				// For simple access like foo?.bar, check foo's type
				// Also handle call expressions that aren't chained (e.g., foo?.bar()?.baz)
				if isCallExpr(expression) {
					// For both optional calls (foo?.()) and regular calls (foo()),
					// get the function's return type, not the full expression type which includes undefined
					exprType = getCallReturnType(expression)
					if exprType == nil {
						// For union types or errors, use GetTypeAtLocation
						exprType = getResolvedType(expression)
						if exprType == nil {
							return
						}
					}
				} else if isPropertyAccess(expression) || isElementAccess(expression) {
					// Handle property/element access on call expression result
					// e.g., foo?.().bar?.baz or foo?.bar?.().baz
					var innerExpr *ast.Node
					if isPropertyAccess(expression) {
						innerExpr = expression.AsPropertyAccessExpression().Expression
					} else {
						innerExpr = expression.AsElementAccessExpression().Expression
					}

					// Check if the inner expression is a call expression
					if innerExpr != nil && isCallExpr(innerExpr) {
						exprType = getTypeFromCallProperty(innerExpr, expression)
						if exprType == nil {
							return
						}
					} else {
						exprType = getResolvedType(expression)
					}
				} else {
					exprType = getResolvedType(expression)
				}
			}

			if exprType == nil {
				return
			}

			if isCallExpr(expression) {
				callExpr := expression.AsCallExpression()
				if isCallExpressionNullableOriginFromCallee(callExpr) {
					exprType = removeNullishFromType(exprType)
					if exprType == nil {
						return
					}
				}
			}

			// Special case: if expression is a call to a union of functions
			// and any function returns nullish, allow the optional chain
			// e.g., type Foo = (() => undefined) | (() => number) | null
			//       foo?.()?.bar - second ?. is necessary because result could be undefined
			if isCallExpr(expression) {
				callExpr := expression.AsCallExpression()
				funcType := getResolvedType(callExpr.Expression)
				if funcType != nil {
					// Check the ORIGINAL type's parts, not after removeNullishFromType
					// because removeNullishFromType only returns the first non-nullish part
					parts := utils.UnionTypeParts(funcType)
					for _, part := range parts {
						// Skip nullish parts (null, undefined, void)
						partFlags := checker.Type_flags(part)
						if partFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
							continue
						}
						// Check if this function part returns nullish
						sigs := ctx.TypeChecker.GetCallSignatures(part)
						if len(sigs) > 0 {
							retType := ctx.TypeChecker.GetReturnTypeOfSignature(sigs[0])
							if retType != nil && isNullishType(retType) {
								// At least one function returns nullish, allow the optional chain
								return
							}
						}
					}
				}
			}

			// Allow optional chain on indeterminate types since we can't determine if they're nullish
			// This includes types like any, unknown, T, T[K], keyof T, etc.
			if isIndeterminateType(exprType) {
				return
			}

			// Also allow if it's a union that includes an indeterminate type
			if utils.IsUnionType(exprType) {
				if slices.ContainsFunc(exprType.Types(), isIndeterminateType) {
					return
				}
			}

			if !isNullishType(exprType) {
				ctx.ReportNode(node, buildNeverOptionalChainMessage())
			}
		}

		isNullableMemberExpression := func(node *ast.Node) bool {
			node = ast.SkipParentheses(node)

			switch node.Kind {
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				nameNode := propAccess.Name()
				if nameNode == nil {
					return false
				}

				baseType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
				if baseType == nil {
					return false
				}

				propName := ast.GetTextOfPropertyName(nameNode)
				if propName == "" {
					return false
				}

				propSymbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if propSymbol == nil {
					propSymbol = checker.Checker_getPropertyOfType(ctx.TypeChecker, baseType, propName)
				}
				if propSymbol == nil {
					for _, prop := range checker.Checker_getPropertiesOfType(ctx.TypeChecker, baseType) {
						if prop.Name == propName {
							propSymbol = prop
							break
						}
					}
				}
				return propSymbol != nil && propSymbol.Flags&ast.SymbolFlagsOptional != 0

			case ast.KindElementAccessExpression:
				return isNullableElementAccessExpression(node.AsElementAccessExpression())
			}

			return false
		}

		var optionChainContainsOptionArrayIndex func(*ast.Node) bool
		optionChainContainsOptionArrayIndex = func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			node = ast.SkipParentheses(node)

			var lhsNode *ast.Node
			var isOptional bool

			switch node.Kind {
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				lhsNode = callExpr.Expression
				isOptional = callExpr.QuestionDotToken != nil
			case ast.KindPropertyAccessExpression:
				propAccess := node.AsPropertyAccessExpression()
				lhsNode = propAccess.Expression
				isOptional = propAccess.QuestionDotToken != nil
			case ast.KindElementAccessExpression:
				elemAccess := node.AsElementAccessExpression()
				lhsNode = elemAccess.Expression
				isOptional = elemAccess.QuestionDotToken != nil
			default:
				return false
			}

			if isOptional && isArrayIndexExpression(lhsNode) {
				return true
			}

			switch lhsNode.Kind {
			case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				return optionChainContainsOptionArrayIndex(lhsNode)
			default:
				return false
			}
		}

		var checkNode func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node)
		checkNode = func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node) {
			if expression == nil {
				return
			}
			if reportNode == nil {
				reportNode = expression
			}

			expression = ast.SkipParentheses(expression)

			if expression.Kind == ast.KindPrefixUnaryExpression {
				unaryExpr := expression.AsPrefixUnaryExpression()
				if unaryExpr.Operator == ast.KindExclamationToken {
					checkNode(unaryExpr.Operand, !isUnaryNotArgument, reportNode)
					return
				}
			}

			if !noUncheckedIndexedAccess && isArrayIndexExpression(expression) {
				return
			}

			if expression.Kind == ast.KindBinaryExpression {
				binExpr := expression.AsBinaryExpression()
				if opKind := binExpr.OperatorToken.Kind; opKind != ast.KindQuestionQuestionToken &&
					(opKind == ast.KindAmpersandAmpersandToken || opKind == ast.KindBarBarToken) {
					checkNode(binExpr.Right, false, nil)
					return
				}
			}

			nodeType := getResolvedType(expression)
			if nodeType == nil || isConditionalAlwaysNecessary(nodeType) {
				return
			}

			if isElementAccess(expression) && isIndexedAccessFlags(checker.Type_flags(nodeType)) {
				elemAccess := expression.AsElementAccessExpression()
				baseType := getResolvedType(elemAccess.Expression)
				if baseType != nil {
					if stringIndexType := ctx.TypeChecker.GetStringIndexType(baseType); stringIndexType != nil {
						nodeType = stringIndexType
					}
				}
			}

			flags := checker.Type_flags(nodeType)
			if isNeverType(flags) {
				ctx.ReportNode(reportNode, buildNeverMessage())
				return
			}

			isTruthy, isFalsy := checkTypeCondition(nodeType)
			if isFalsy {
				if isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				}
				return
			}
			if isTruthy {
				if isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				}
			}
		}

		checkNodeForNullish := func(node *ast.Node) {
			if node == nil {
				return
			}

			nodeType := getResolvedType(node)
			if nodeType == nil || isConditionalAlwaysNecessary(nodeType) {
				return
			}

			if constrainedType, shouldSkip := constrainIndexedAccessType(nodeType); shouldSkip {
				return
			} else if constrainedType != nil {
				nodeType = constrainedType
			}

			flags := checker.Type_flags(nodeType)
			switch {
			case isNeverType(flags):
				ctx.ReportNode(node, buildNeverMessage())
				return
			case isAlwaysNullishType(nodeType):
				ctx.ReportNode(node, buildAlwaysNullishMessage())
				return
			}

			if !isNullishType(nodeType) && !isNullableMemberExpression(node) {
				node = ast.SkipParentheses(node)

				if noUncheckedIndexedAccess ||
					(!isArrayIndexExpression(node) &&
						!(hasOptionalChain(node) && optionChainContainsOptionArrayIndex(node))) {
					ctx.ReportNode(node, buildNeverNullishMessage())
				}
			}
		}

		checkIfBoolExpressionIsNecessaryConditional := func(node *ast.Node, left *ast.Node, right *ast.Node, opKind ast.Kind) {
			leftType := getResolvedType(left)
			rightType := getResolvedType(right)
			if leftType == nil || rightType == nil {
				return
			}

			if _, ok := toStaticValue(leftType); ok {
				if _, ok := toStaticValue(rightType); ok {
					ctx.ReportNode(node, buildLiteralBinaryExpressionMessage())
					return
				}
			}

			if !isStrictNullChecks {
				return
			}

			leftFlags := checker.Type_flags(leftType)
			rightFlags := checker.Type_flags(rightType)

			isComparable := func(t *checker.Type, flags checker.TypeFlags) bool {
				if t == nil {
					return false
				}

				flags |= checker.TypeFlagsAny | checker.TypeFlagsUnknown | checker.TypeFlagsTypeParameter | checker.TypeFlagsTypeVariable
				if opKind == ast.KindEqualsEqualsToken || opKind == ast.KindExclamationEqualsToken {
					flags |= checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid
				}

				if checker.Type_flags(t)&(checker.TypeFlagsConditional|checker.TypeFlagsTypeVariable|checker.TypeFlagsSubstitution|checker.TypeFlagsIncludesConstrainedTypeVariable) != 0 {
					return true
				}

				for _, part := range utils.UnionTypeParts(t) {
					partFlags := checker.Type_flags(part)
					if partFlags&flags != 0 {
						return true
					}
					if partFlags&(checker.TypeFlagsConditional|checker.TypeFlagsTypeVariable|checker.TypeFlagsSubstitution|checker.TypeFlagsIncludesConstrainedTypeVariable) != 0 {
						return true
					}
				}

				return false
			}

			if (leftFlags == checker.TypeFlagsUndefined && !isComparable(rightType, checker.TypeFlagsUndefined|checker.TypeFlagsVoid)) ||
				(rightFlags == checker.TypeFlagsUndefined && !isComparable(leftType, checker.TypeFlagsUndefined|checker.TypeFlagsVoid)) ||
				(leftFlags == checker.TypeFlagsNull && !isComparable(rightType, checker.TypeFlagsNull)) ||
				(rightFlags == checker.TypeFlagsNull && !isComparable(leftType, checker.TypeFlagsNull)) {
				ctx.ReportDiagnostic(buildNoOverlapDiagnostic(
					ctx.TypeChecker.TypeToString(leftType),
					left.Loc,
					ctx.TypeChecker.TypeToString(rightType),
					right.Loc,
				))
			}
		}

		checkLogicalExpressionForUnnecessaryConditionals := func(node *ast.BinaryExpression) {
			if node.OperatorToken.Kind == ast.KindQuestionQuestionToken {
				checkNodeForNullish(node.Left)
				return
			}

			checkNode(node.Left, false, nil)
		}

		checkIfLoopIsNecessaryConditional := func(test *ast.Node) {
			if test == nil {
				return
			}

			if loopConditionMode == "only-allowed-literals" && isAllowedConstantLiteral(test) {
				return
			}

			if loopConditionMode == "always" {
				if ast.SkipParentheses(test).Kind == ast.KindTrueKeyword {
					return
				}
				if testType := getResolvedType(test); isTrueLiteralType(testType) {
					return
				}
			}

			checkNode(test, false, nil)
		}

		checkAssignmentExpression := func(node *ast.BinaryExpression) {
			switch node.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken:
				checkNode(node.Left, false, nil)
			case ast.KindQuestionQuestionEqualsToken:
				checkNodeForNullish(node.Left)
			}
		}

		checkTypePredicateCallExpression := func(node *ast.Node, callExpr *ast.CallExpression) {
			if !opts.CheckTypePredicates {
				return
			}

			var checkableArguments []*ast.Node
			if callExpr.Arguments != nil {
				for _, arg := range callExpr.Arguments.Nodes {
					if arg == nil {
						continue
					}
					if arg.Kind == ast.KindSpreadElement {
						break
					}
					checkableArguments = append(checkableArguments, arg)
				}
			}
			if len(checkableArguments) == 0 {
				return
			}

			callSignature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
			if callSignature == nil {
				return
			}

			typePredicate := ctx.TypeChecker.GetTypePredicateOfSignature(callSignature)
			if typePredicate == nil {
				return
			}

			paramIndex := int(checker.TypePredicate_parameterIndex(typePredicate))
			if paramIndex < 0 || paramIndex >= len(checkableArguments) {
				return
			}

			arg := checkableArguments[paramIndex]
			predicateType := checker.TypePredicate_t(typePredicate)

			switch checker.TypePredicate_kind(typePredicate) {
			case checker.TypePredicateKindAssertsIdentifier:
				if predicateType == nil {
					checkNode(arg, false, nil)
					return
				}
				if argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg); argType != nil && argType == predicateType {
					ctx.ReportNode(arg, buildTypeGuardAlreadyIsTypeMessage())
				}
			case checker.TypePredicateKindIdentifier:
				if argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg); argType != nil && argType == predicateType {
					ctx.ReportNode(arg, buildTypeGuardAlreadyIsTypeMessage())
				}
			}
		}

		checkCallExpression := func(node *ast.Node) {
			checkOptionalChain(node)

			callExpr := node.AsCallExpression()
			if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) &&
				callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
				if arg := callExpr.Arguments.Nodes[0]; arg != nil {
					checkPredicateFunction(ctx, arg, opts.CheckTypePredicates)
				}
			}

			checkTypePredicateCallExpression(node, callExpr)
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				checkNode(node.AsIfStatement().Expression, false, nil)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node.AsWhileStatement().Expression)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node.AsDoStatement().Expression)
			},
			ast.KindForStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node.AsForStatement().Condition)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkNode(node.AsConditionalExpression().Condition, false, nil)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()

				switch binExpr.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
					checkLogicalExpressionForUnnecessaryConditionals(binExpr)
				case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
					checkAssignmentExpression(binExpr)
				case ast.KindLessThanToken,
					ast.KindGreaterThanToken,
					ast.KindLessThanEqualsToken,
					ast.KindGreaterThanEqualsToken,
					ast.KindEqualsEqualsToken,
					ast.KindEqualsEqualsEqualsToken,
					ast.KindExclamationEqualsToken,
					ast.KindExclamationEqualsEqualsToken:
					checkIfBoolExpressionIsNecessaryConditional(node, binExpr.Left, binExpr.Right, binExpr.OperatorToken.Kind)
				}
			},
			ast.KindPropertyAccessExpression: checkOptionalChain,
			ast.KindElementAccessExpression:  checkOptionalChain,
			ast.KindCallExpression:           checkCallExpression,
			ast.KindCaseClause: func(node *ast.Node) {
				if node.Expression() == nil || node.Parent == nil || node.Parent.Parent == nil || node.Parent.Parent.Kind != ast.KindSwitchStatement {
					return
				}

				switchStmt := node.Parent.Parent.AsSwitchStatement()
				checkIfBoolExpressionIsNecessaryConditional(node.Expression(), switchStmt.Expression, node.Expression(), ast.KindEqualsEqualsEqualsToken)
			},
		}
	},
}

// checkTypeCondition determines if a type is always truthy or always falsy at runtime.
//
// Return values:
// - (true, false): type is always truthy (e.g., objects, "hello", 1, true)
// - (false, true): type is always falsy (e.g., null, undefined, false, 0, "", never)
// - (false, false): type could be either (e.g., string, number, boolean)
//
// Examples:
//   - { foo: string }: always truthy (objects are always truthy)
//   - "hello": always truthy (non-empty string literal)
//   - "": always falsy (empty string literal)
//   - 0: always falsy (zero is falsy)
//   - 1: always truthy (non-zero number)
//   - true: always truthy
//   - false: always falsy
//   - null: always falsy
//   - undefined: always falsy
//   - never: always falsy (type with no possible values)
//   - string: could be either (might be "" or "hello")
//   - number: could be either (might be 0 or 1)
//   - boolean: could be either (might be true or false)
//
// Type handling:
// - Union types: all parts must be truthy for (true, false), all must be falsy for (false, true)
// - Intersection types: if any part is always falsy, result is falsy; all must be truthy for truthy
// - Literal types: evaluates the actual literal value's truthiness
// - Object types: always truthy (even empty objects are truthy in JavaScript)
// - Symbols: always truthy (symbols are always truthy)
func checkTypeCondition(t *checker.Type) (isTruthy bool, isFalsy bool) {
	flags := checker.Type_flags(t)

	// Never type is always falsy (empty type, no values exist)
	if flags&checker.TypeFlagsNever != 0 {
		return false, true
	}

	// Handle indexed access types (e.g., Obj[Key] where Obj: Record<string, 1|2|3>)
	// Try to resolve them by checking if the base type has an index signature
	if flags&checker.TypeFlagsIndexedAccess != 0 {
		// Indexed access types need special handling
		// For now, treat them as indeterminate (could be truthy or falsy)
		// unless we can determine the index signature type
		// TODO: Try to resolve the indexed access to get the actual element type
		return false, false
	}
	// Handle unions - check all parts
	if utils.IsUnionType(t) {
		allTruthy := true
		allFalsy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(part)
			if !partTruthy {
				allTruthy = false
			}
			if !partFalsy {
				allFalsy = false
			}
		}

		return allTruthy, allFalsy
	}

	// Handle intersections - check all parts
	// For intersections, all parts must be truthy for the whole to be truthy
	if utils.IsIntersectionType(t) {
		allTruthy := true

		for _, part := range t.Types() {
			partTruthy, partFalsy := checkTypeCondition(part)
			// If any part is always falsy, intersection is likely never/empty
			if partFalsy {
				return false, true
			}
			// If any part is not always truthy, we can't say the whole is always truthy
			if !partTruthy {
				allTruthy = false
			}
		}

		return allTruthy, false
	}

	// Nullish types are always falsy
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false, true
	}

	// Objects and non-primitive types are always truthy
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive) != 0 {
		return true, false
	}

	// ESSymbol is always truthy
	if flags&(checker.TypeFlagsESSymbol|checker.TypeFlagsUniqueESSymbol) != 0 {
		return true, false
	}

	// Boolean literals - check flags first
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		// Boolean literal types can be intrinsic or fresh literal types
		// Check if it's an intrinsic type first
		if utils.IsIntrinsicType(t) {
			intrinsicName := t.AsIntrinsicType().IntrinsicName()
			if intrinsicName == "true" {
				return true, false
			}
			if intrinsicName == "false" {
				return false, true
			}
		} else if t.AsLiteralType() != nil {
			// For fresh literal types, check via AsLiteralType
			litStr := t.AsLiteralType().String()
			if litStr == "true" {
				return true, false
			}
			if litStr == "false" {
				return false, true
			}
		}
	}

	// String literals
	if flags&checker.TypeFlagsStringLiteral != 0 && t.IsStringLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if literal.Value() == "" {
				return false, true
			}
			return true, false
		}
	}

	// Number literals
	if flags&checker.TypeFlagsNumberLiteral != 0 && t.IsNumberLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			value := literal.String()
			if value == "0" || value == "NaN" {
				return false, true
			}
			return true, false
		}
	}

	// BigInt literals
	if flags&checker.TypeFlagsBigIntLiteral != 0 && t.IsBigIntLiteral() {
		literal := t.AsLiteralType()
		if literal != nil {
			if literal.String() == "0" || literal.String() == "0n" {
				return false, true
			}
			return true, false
		}
	}

	// Generic types (boolean, string, number, etc.) are not always truthy or falsy
	return false, false
}

// isNullishType checks if a type can be null, undefined, or void.
//
// For union types, returns true if any part of the union is nullish.
// This is used to determine if the nullish coalescing operator (??) or
// optional chaining (?.) might be necessary.
func isNullishType(t *checker.Type) bool {
	if utils.IsUnionType(t) {
		return slices.ContainsFunc(t.Types(), isNullishType)
	}

	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
}

// removeNullishFromType removes null, undefined, and void from a union type.
// Returns the non-nullish part of the type, or nil if the type is entirely nullish.
func removeNullishFromType(t *checker.Type) *checker.Type {
	if !utils.IsUnionType(t) {
		// Not a union - check if it's nullish
		flags := checker.Type_flags(t)
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return nil
		}
		return t
	}

	// For union types, filter out nullish parts
	var nonNullishParts []*checker.Type
	for _, part := range t.Types() {
		if !isNullishType(part) {
			nonNullishParts = append(nonNullishParts, part)
		}
	}

	if len(nonNullishParts) == 0 {
		return nil
	}
	if len(nonNullishParts) == 1 {
		return nonNullishParts[0]
	}

	// Multiple non-nullish parts - return first one for now
	// (TypeScript would create a new union type, but we don't have that API)
	return nonNullishParts[0]
}

// isAllowedConstantLiteral checks if an expression is a literal that's allowed in loop conditions.
//
// When AllowConstantLoopConditions is set to "only-allowed-literals", only the
// following literals are allowed in loop conditions: true, false, 0, 1
func isAllowedConstantLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)

	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindNumericLiteral:
		literal := node.AsNumericLiteral()
		text := literal.Text
		return text == "0" || text == "1"
	}

	return false
}

func normalizeAllowConstantLoopConditions(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
	case bool:
		if v {
			return "always"
		}
	case *bool:
		if v != nil && *v {
			return "always"
		}
	}

	return "never"
}

// checkPredicateFunction analyzes predicate functions used in array methods like filter/find.
//
// This function performs two checks:
//  1. If checkTypeGuards is true and the function is a type guard, it checks if the
//     parameter already satisfies the type predicate (making the guard unnecessary)
//  2. It checks if the function's return type is always truthy or always falsy,
//     which would make it a useless filter/find predicate
//
// Used for array methods like:
// - [1, 2, 3].filter(() => true) // always truthy, returns all elements
// - [1, 2, 3].find(() => false)  // always falsy, returns undefined
func checkPredicateFunction(ctx rule.RuleContext, funcNode *ast.Node, checkTypeGuards bool) {
	isFunction := funcNode.Kind&(ast.KindArrowFunction|ast.KindFunctionExpression|ast.KindFunctionDeclaration) != 0
	if !isFunction {
		return
	}

	funcType := ctx.TypeChecker.GetTypeAtLocation(funcNode)
	signatures := ctx.TypeChecker.GetCallSignatures(funcType)

	for _, signature := range signatures {
		// Check if this is a type predicate (type guard)
		typePredicate := ctx.TypeChecker.GetTypePredicateOfSignature(signature)
		if checkTypeGuards && typePredicate != nil {
			// Check if the argument already satisfies the type predicate
			params := checker.Signature_parameters(signature)
			if len(params) > 0 {
				// Get the parameter index being guarded
				paramIndex := int(checker.TypePredicate_parameterIndex(typePredicate))

				if paramIndex >= 0 && paramIndex < len(params) {
					param := params[paramIndex]
					if param != nil {
						paramType := ctx.TypeChecker.GetTypeOfSymbol(param)
						predicateKind := checker.TypePredicate_kind(typePredicate)

						if paramType != nil {
							// Only check "x is Type" predicates, not "asserts x" predicates
							// "asserts x" predicates in functions are checked via their return type below
							if predicateKind == checker.TypePredicateKindIdentifier ||
								predicateKind == checker.TypePredicateKindThis {
								predicateType := checker.TypePredicate_t(typePredicate)
								if predicateType != nil {
									// Check if paramType is assignable to predicateType
									// If so, the type guard is unnecessary
									if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, paramType, predicateType) {
										ctx.ReportNode(funcNode, buildTypeGuardAlreadyIsTypeMessage())
										return
									}
								}
							}
						}
					}
				}
			}
		}

		returnType := ctx.TypeChecker.GetReturnTypeOfSignature(signature)

		// Handle type parameters
		typeFlags := checker.Type_flags(returnType)
		if typeFlags&checker.TypeFlagsTypeParameter != 0 {
			constraint := ctx.TypeChecker.GetConstraintOfTypeParameter(returnType)
			if constraint != nil {
				returnType = constraint
			}
		}

		isTruthy, isFalsy := checkTypeCondition(returnType)

		if isTruthy || isFalsy {
			// Use different message based on whether it's a literal function or function reference
			// Literal functions: () => true, () => false, function() { return true }
			// Function references: truthy, falsy (identifier)
			isLiteralFunction := funcNode.Kind == ast.KindArrowFunction || funcNode.Kind == ast.KindFunctionExpression

			if isTruthy {
				if isLiteralFunction {
					ctx.ReportNode(funcNode, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(funcNode, buildAlwaysTruthyFuncMessage())
				}
			} else if isFalsy {
				if isLiteralFunction {
					ctx.ReportNode(funcNode, buildAlwaysFalsyMessage())
				} else {
					ctx.ReportNode(funcNode, buildAlwaysFalsyFuncMessage())
				}
			}
		}
	}
}
