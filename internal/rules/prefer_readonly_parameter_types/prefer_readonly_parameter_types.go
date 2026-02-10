package prefer_readonly_parameter_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func buildShouldBeReadonlyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shouldBeReadonly",
		Description: "Parameter should be a readonly type.",
	}
}

type readonlyness uint8

const (
	readonlynessUnknown readonlyness = iota
	readonlynessMutable
	readonlynessReadonly
)

type readonlynessOptions struct {
	allow                  []utils.TypeOrValueSpecifier
	treatMethodsAsReadonly bool
}

func isTypeReadonlyArrayOrTuple(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts readonlynessOptions,
	seenTypes map[*checker.Type]struct{},
) readonlyness {
	checkTypeArguments := func(arrayType *checker.Type) readonlyness {
		typeArguments := checker.Checker_getTypeArguments(typeChecker, arrayType)
		if len(typeArguments) == 0 {
			return readonlynessReadonly
		}

		for _, typeArg := range typeArguments {
			if isTypeReadonlyRecurser(program, typeChecker, typeArg, opts, seenTypes) == readonlynessMutable {
				return readonlynessMutable
			}
		}

		return readonlynessReadonly
	}

	if checker.Checker_isArrayType(typeChecker, t) {
		symbol := checker.Type_symbol(t)
		if symbol != nil && symbol.Name == "Array" {
			return readonlynessMutable
		}

		return checkTypeArguments(t)
	}

	if checker.IsTupleType(t) {
		tupleTarget := t
		if tupleTarget.Target() != nil {
			tupleTarget = tupleTarget.Target()
		}

		if !checker.IsTupleType(tupleTarget) {
			return readonlynessUnknown
		}

		if !checker.TupleType_readonly(tupleTarget.AsTupleType()) {
			return readonlynessMutable
		}

		return checkTypeArguments(t)
	}

	return readonlynessUnknown
}

func propertyHasPrivateIdentifierName(property *ast.Symbol) bool {
	if property == nil {
		return false
	}

	if property.ValueDeclaration != nil {
		name := property.ValueDeclaration.Name()
		return name != nil && name.Kind == ast.KindPrivateIdentifier
	}

	for _, declaration := range property.Declarations {
		if declaration == nil {
			continue
		}
		name := declaration.Name()
		if name != nil && name.Kind == ast.KindPrivateIdentifier {
			return true
		}
	}

	return false
}

func propertyIsReadonly(typeChecker *checker.Checker, property *ast.Symbol) bool {
	if property == nil {
		return false
	}

	if checker.Checker_isReadonlySymbol(typeChecker, property) {
		return true
	}

	if property.CheckFlags&ast.CheckFlagsReadonly != 0 {
		return true
	}

	if property.CheckFlags&ast.CheckFlagsSyntheticMethod != 0 {
		return true
	}

	if property.Flags&ast.SymbolFlagsMethod != 0 && property.CheckFlags&ast.CheckFlagsMapped != 0 {
		return true
	}

	if property.ValueDeclaration == nil && property.CheckFlags&ast.CheckFlagsMapped != 0 {
		return true
	}

	return checker.GetDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsReadonly != 0
}

func isTypeReadonlyObject(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts readonlynessOptions,
	seenTypes map[*checker.Type]struct{},
) readonlyness {
	properties := checker.Checker_getPropertiesOfType(typeChecker, t)
	if len(properties) > 0 {
		for _, property := range properties {
			if opts.treatMethodsAsReadonly && property.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}

			if propertyIsReadonly(typeChecker, property) {
				continue
			}

			if propertyHasPrivateIdentifierName(property) {
				continue
			}
			return readonlynessMutable
		}

		for _, property := range properties {
			if property.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}

			propertyType := checker.Checker_getTypeOfPropertyOfType(typeChecker, t, property.Name)
			if propertyType == nil {
				propertyType = checker.Checker_getTypeOfSymbol(typeChecker, property)
			}

			if propertyType == nil {
				continue
			}

			if _, ok := seenTypes[propertyType]; ok {
				continue
			}

			if isTypeReadonlyRecurser(program, typeChecker, propertyType, opts, seenTypes) == readonlynessMutable {
				return readonlynessMutable
			}
		}
	}

	for _, info := range checker.Checker_getIndexInfosOfType(typeChecker, t) {
		if !info.IsReadonly() {
			return readonlynessMutable
		}

		valueType := info.ValueType()
		if valueType == nil || valueType == t {
			continue
		}
		if _, ok := seenTypes[valueType]; ok {
			continue
		}

		if isTypeReadonlyRecurser(program, typeChecker, valueType, opts, seenTypes) == readonlynessMutable {
			return readonlynessMutable
		}
	}

	return readonlynessReadonly
}

func isTypeReadonlyRecurser(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts readonlynessOptions,
	seenTypes map[*checker.Type]struct{},
) readonlyness {
	seenTypes[t] = struct{}{}

	if utils.TypeMatchesSomeSpecifier(t, opts.allow, program) {
		return readonlynessReadonly
	}

	if utils.IsUnionType(t) {
		for _, subType := range t.Types() {
			if _, ok := seenTypes[subType]; ok {
				continue
			}
			if isTypeReadonlyRecurser(program, typeChecker, subType, opts, seenTypes) != readonlynessReadonly {
				return readonlynessMutable
			}
		}
		return readonlynessReadonly
	}

	if utils.IsIntersectionType(t) {
		hasArrayOrTuple := false
		for _, subType := range t.Types() {
			if checker.Checker_isArrayType(typeChecker, subType) || checker.IsTupleType(subType) {
				hasArrayOrTuple = true
				break
			}
		}

		if hasArrayOrTuple {
			for _, subType := range t.Types() {
				if _, ok := seenTypes[subType]; ok {
					continue
				}
				if isTypeReadonlyRecurser(program, typeChecker, subType, opts, seenTypes) != readonlynessReadonly {
					return readonlynessMutable
				}
			}
			return readonlynessReadonly
		}

		readonlyObject := isTypeReadonlyObject(program, typeChecker, t, opts, seenTypes)
		if readonlyObject != readonlynessUnknown {
			return readonlyObject
		}
	}

	if !utils.IsObjectType(t) {
		return readonlynessReadonly
	}

	if len(utils.GetCallSignatures(typeChecker, t)) > 0 && len(checker.Checker_getPropertiesOfType(typeChecker, t)) == 0 {
		return readonlynessReadonly
	}

	readonlyArray := isTypeReadonlyArrayOrTuple(program, typeChecker, t, opts, seenTypes)
	if readonlyArray != readonlynessUnknown {
		return readonlyArray
	}

	readonlyObject := isTypeReadonlyObject(program, typeChecker, t, opts, seenTypes)
	if readonlyObject != readonlynessUnknown {
		return readonlyObject
	}

	return readonlynessReadonly
}

func isTypeReadonly(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts readonlynessOptions,
) bool {
	return isTypeReadonlyRecurser(program, typeChecker, t, opts, map[*checker.Type]struct{}{}) == readonlynessReadonly
}

func isLiteralOrTaggablePrimitiveLike(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsLiteral != 0 {
		return true
	}
	return flags&(checker.TypeFlagsBigInt|checker.TypeFlagsNumber|checker.TypeFlagsString|checker.TypeFlagsTemplateLiteral) != 0
}

func isObjectLiteralLike(typeChecker *checker.Checker, t *checker.Type) bool {
	return len(utils.GetCallSignatures(typeChecker, t)) == 0 &&
		len(utils.GetConstructSignatures(typeChecker, t)) == 0 &&
		utils.IsObjectType(t)
}

func isTypeBrandedLiteral(typeChecker *checker.Checker, t *checker.Type) bool {
	if !utils.IsIntersectionType(t) {
		return false
	}

	hadObjectLike := false
	hadPrimitiveLike := false

	for _, constituent := range t.Types() {
		if isObjectLiteralLike(typeChecker, constituent) {
			hadObjectLike = true
		} else if isLiteralOrTaggablePrimitiveLike(constituent) {
			hadPrimitiveLike = true
		} else {
			return false
		}
	}

	return hadObjectLike && hadPrimitiveLike
}

func isTypeBrandedLiteralLike(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, part := range t.Types() {
			if !isTypeBrandedLiteral(typeChecker, part) {
				return false
			}
		}
		return true
	}

	return isTypeBrandedLiteral(typeChecker, t)
}

func isParameterProperty(parameter *ast.Node) bool {
	if !ast.IsParameter(parameter) {
		return false
	}

	if parameter.Parent == nil || parameter.Parent.Kind != ast.KindConstructor {
		return false
	}

	return parameter.ModifierFlags()&ast.ModifierFlagsParameterPropertyModifier != 0
}

var PreferReadonlyParameterTypesRule = rule.Rule{
	Name: "prefer-readonly-parameter-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := utils.UnmarshalOptions[PreferReadonlyParameterTypesOptions](options, "prefer-readonly-parameter-types")

		listener := func(node *ast.Node) {
			for _, parameter := range node.Parameters() {
				parameterIsProperty := isParameterProperty(parameter)
				if !opts.CheckParameterProperties && parameterIsProperty {
					continue
				}

				actualParameter := parameter

				if opts.IgnoreInferredTypes && actualParameter.Type() == nil {
					continue
				}

				t := ctx.TypeChecker.GetTypeAtLocation(actualParameter)
				if t == nil {
					continue
				}

				isReadOnly := isTypeReadonly(ctx.Program, ctx.TypeChecker, t, readonlynessOptions{
					allow:                  opts.Allow,
					treatMethodsAsReadonly: opts.TreatMethodsAsReadonly,
				})

				if !isReadOnly && !isTypeBrandedLiteralLike(ctx.TypeChecker, t) {
					if parameterIsProperty {
						name := actualParameter.Name()
						if name != nil {
							nameStart := utils.TrimNodeTextRange(ctx.SourceFile, name).Pos()
							ctx.ReportRange(actualParameter.Loc.WithPos(nameStart), buildShouldBeReadonlyMessage())
							continue
						}
					}

					ctx.ReportNode(actualParameter, buildShouldBeReadonlyMessage())
				}
			}
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:       listener,
			ast.KindCallSignature:       listener,
			ast.KindConstructSignature:  listener,
			ast.KindConstructor:         listener,
			ast.KindFunctionDeclaration: listener,
			ast.KindFunctionExpression:  listener,
			ast.KindFunctionType:        listener,
			ast.KindMethodDeclaration:   listener,
			ast.KindMethodSignature:     listener,
		}
	},
}
