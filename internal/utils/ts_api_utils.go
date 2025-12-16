package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
)

func UnionTypeParts(t *checker.Type) []*checker.Type {
	if IsUnionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}
func IntersectionTypeParts(t *checker.Type) []*checker.Type {
	if IsIntersectionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

func IsTypeFlagSet(t *checker.Type, flags checker.TypeFlags) bool {
	return t != nil && checker.Type_flags(t)&flags != 0
}

func IsIntrinsicType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntrinsic)
}

func IsIntrinsicErrorType(t *checker.Type) bool {
	return IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "error"
}

func IsIntrinsicVoidType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsVoid)
}

func IsUnionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnion)
}
func IsIntersectionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntersection)
}
func IsTypeAnyType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsAny)
}
func IsTypeUnknownType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnknown)
}
func IsObjectType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsObject)
}
func IsTypeParameter(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsTypeParameter)
}
func IsTypeNullType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsNull)
}
func IsTypeUndefinedType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUndefined)
}
func IsTypeVoidType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsVoid)
}
func IsBooleanLiteralType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsBoolean)
}
func IsTrueLiteralType(t *checker.Type) bool {
	return IsBooleanLiteralType(t) && IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "true"
}
func IsFalseLiteralType(t *checker.Type) bool {
	return IsBooleanLiteralType(t) && IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "false"
}

func GetCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}
func GetConstructSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindConstruct)
}

// ex. getCallSignaturesOfType
func CollectAllCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	if IsUnionType(t) {
		signatures := []*checker.Signature{}
		for _, subtype := range t.Types() {
			signatures = append(signatures, GetCallSignatures(typeChecker, subtype)...)
		}
		return signatures
	}
	if IsIntersectionType(t) {
		var signatures []*checker.Signature
		for _, subtype := range t.Types() {
			sig := GetCallSignatures(typeChecker, subtype)
			if len(sig) != 0 {
				if signatures != nil {
					return []*checker.Signature{}
				}
				signatures = sig
			}
		}
		if signatures == nil {
			return []*checker.Signature{}
		}
		return signatures
	}
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}

func IsSymbolFlagSet(symbol *ast.Symbol, flag ast.SymbolFlags) bool {
	return symbol != nil && symbol.Flags&flag != 0
}

func IsCallback(
	typeChecker *checker.Checker,
	param *ast.Symbol,
	node *ast.Node,
) bool {
	t := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(param, node))

	if param.ValueDeclaration != nil && ast.IsParameter(param.ValueDeclaration) && param.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
		t = checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
		if t == nil {
			return false
		}
	}

	for _, subType := range UnionTypeParts(t) {
		if len(GetCallSignatures(typeChecker, subType)) != 0 {
			return true
		}
	}

	return false
}

// TODO(note): why there is no IntersectionTypeParts
func IsThenableType(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	if t == nil {
		t = typeChecker.GetTypeAtLocation(node)
	}
	for _, typePart := range UnionTypeParts(checker.Checker_getApparentType(typeChecker, t)) {
		then := checker.Checker_getPropertyOfType(typeChecker, typePart, "then")
		if then == nil {
			continue
		}

		thenType := typeChecker.GetTypeOfSymbolAtLocation(then, node)

		for _, subTypePart := range UnionTypeParts(thenType) {
			for _, signature := range checker.Checker_getSignaturesOfType(typeChecker, subTypePart, checker.SignatureKindCall) {
				if len(checker.Signature_parameters(signature)) != 0 && IsCallback(typeChecker, checker.Signature_parameters(signature)[0], node) {
					return true
				}
			}
		}
	}
	return false
}

func GetWellKnownSymbolPropertyOfType(t *checker.Type, name string, typeChecker *checker.Checker) *ast.Symbol {
	return checker.Checker_getPropertyOfType(typeChecker, t, checker.Checker_getPropertyNameForKnownSymbolName(typeChecker, name))
}

/**
 * Checks if a given compiler option is enabled, accounting for whether all flags
 * (except `strictPropertyInitialization`) have been enabled by `strict: true`.
 * @category Compiler Options
 * @example
 * ```ts
 * const optionsLenient = {
 * 	noImplicitAny: true,
 * };
 *
 * isStrictCompilerOptionEnabled(optionsLenient, "noImplicitAny"); // true
 * isStrictCompilerOptionEnabled(optionsLenient, "noImplicitThis"); // false
 * ```
 * @example
 * ```ts
 * const optionsStrict = {
 * 	noImplicitThis: false,
 * 	strict: true,
 * };
 *
 * isStrictCompilerOptionEnabled(optionsStrict, "noImplicitAny"); // true
 * isStrictCompilerOptionEnabled(optionsStrict, "noImplicitThis"); // false
 * ```
 */
func IsStrictCompilerOptionEnabled(
	options *core.CompilerOptions,
	option core.Tristate,
) bool {
	if options.Strict.IsTrue() {
		return option.IsTrueOrUnknown()
	}
	return option.IsTrue()
	// return (
	// 	(options.strict ? options[option] !== false : options[option] === true) &&
	// 	(option !== "strictPropertyInitialization" ||
	// 		isStrictCompilerOptionEnabled(options, "strictNullChecks"))
	// );
}
