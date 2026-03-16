package utils

import (
	"github.com/microsoft/typescript-go/shim/checker"
)

func MatchesTypeOrBaseType(typeChecker *checker.Checker, t *checker.Type, predicate func(*checker.Type) bool) bool {
	if predicate(t) {
		return true
	}

	if !IsObjectType(t) {
		return false
	}

	target := t
	if checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
		target = t.Target()
	}

	if checker.Type_objectFlags(target)&checker.ObjectFlagsClassOrInterface != 0 {
		for _, baseType := range checker.Checker_getBaseTypes(typeChecker, target) {
			if MatchesTypeOrBaseType(typeChecker, baseType, predicate) {
				return true
			}
		}
	}

	return false
}
