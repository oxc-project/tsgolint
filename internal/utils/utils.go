package utils

import (
	"iter"
	"slices"
	"unicode"

	"github.com/go-json-experiment/json"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

func TrimNodeTextRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	return scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos()).WithEnd(node.End())
}

func GetCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) iter.Seq[ast.CommentRange] {
	nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})

	return func(yield func(ast.CommentRange) bool) {
		for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, sourceFile.Text(), inRange.Pos()) {
			if commentRange.Pos() >= inRange.End() {
				break
			}
			if !yield(commentRange) {
				return
			}
		}

		for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, sourceFile.Text(), inRange.Pos()) {
			if commentRange.Pos() >= inRange.End() {
				break
			}
			if !yield(commentRange) {
				return
			}
		}
	}
}

func HasCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) bool {
	for range GetCommentsInRange(sourceFile, inRange) {
		return true
	}
	return false
}

func TypeRecurser(t *checker.Type, predicate func(t *checker.Type) /* should stop */ bool) bool {
	if IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, subtype := range t.Types() {
			if TypeRecurser(subtype, predicate) {
				return true
			}
		}
		return false
	} else {
		return predicate(t)
	}
}

// SUPER DIRTY HACK FOR OPTIONAL FIELDS :(
func Ref[T any](a T) *T {
	return &a
}

func GetNumberIndexType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	return checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
}
func GetHeritageClauses(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		return node.AsClassExpression().HeritageClauses
	case ast.KindInterfaceDeclaration:
		return node.AsInterfaceDeclaration().HeritageClauses
	}
	return nil
}

// Source: typescript-go/internal/core/core.go
func Filter[T any](slice []T, f func(T) bool) []T {
	for i, value := range slice {
		if !f(value) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func FilterIndex[T any](slice []T, f func(T, int, []T) bool) []T {
	for i, value := range slice {
		if !f(value, i, slice) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value, i, slice) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func Map[T, U any](slice []T, f func(T) U) []U {
	if len(slice) == 0 {
		return nil
	}
	result := make([]U, len(slice))
	for i, value := range slice {
		result[i] = f(value)
	}
	return result
}

// Source: typescript-go/internal/core/core.go
func Some[T any](slice []T, f func(T) bool) bool {
	return slices.ContainsFunc(slice, f)
}

// Source: typescript-go/internal/core/core.go
func Every[T any](slice []T, f func(T) bool) bool {
	for _, value := range slice {
		if !f(value) {
			return false
		}
	}
	return true
}

// Source: typescript-go/internal/core/core.go
func Flatten[T any](array [][]T) []T {
	var result []T
	for _, subArray := range array {
		result = append(result, subArray...)
	}
	return result
}

func IncludesModifier(node interface{ Modifiers() *ast.ModifierList }, modifier ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	return Some(modifiers.NodeList.Nodes, func(m *ast.Node) bool {
		return m.Kind == modifier
	})
}

// Source: https://github.com/microsoft/typescript-go/blob/5652e65d5ae944375676d3955f9755e554576d41/internal/jsnum/string.go#L99
func IsStrWhiteSpace(r rune) bool {
	// This is different than stringutil.IsWhiteSpaceLike.

	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-LineTerminator
	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-WhiteSpace

	switch r {
	// LineTerminator
	case '\n', '\r', 0x2028, 0x2029:
		return true
	// WhiteSpace
	case '\t', '\v', '\f', 0xFEFF:
		return true
	}

	// WhiteSpace
	return unicode.Is(unicode.Zs, r)
}

// UnmarshalOptions unmarshals rule options with proper JSON default handling.
// It accepts options as either the target type T or as any, and ensures that
// JSON unmarshalling occurs to apply default values defined in UnmarshalJSON.
func UnmarshalOptions[T any](options any, ruleName string) T {
	var result T

	// Always marshal and unmarshal to ensure defaults are applied via UnmarshalJSON
	optsBytes, err := json.Marshal(options)
	if err != nil {
		panic(ruleName + ": failed to marshal options: " + err.Error())
	}
	if err := json.Unmarshal(optsBytes, &result); err != nil {
		panic(ruleName + ": failed to unmarshal options: " + err.Error())
	}

	return result
}

// BoolOr represents a JSON value that can be either a boolean or an object of type T.
// This is useful for rule options that accept either `true`/`false` or a detailed config object.
//
// Usage:
//
//	type MyOptions struct {
//	    SomeOption BoolOr[SomeOptionsDetails] `json:"someOption"`
//	}
//
//	// Check the value:
//	if opts.SomeOption.IsTrue() {
//	    // boolean true was provided
//	} else if opts.SomeOption.IsFalse() {
//	    // boolean false was provided (or not provided at all)
//	} else if details := opts.SomeOption.Object(); details != nil {
//	    // object was provided, use details
//	}
type BoolOr[T any] struct {
	isSet     bool
	boolVal   bool
	objectVal *T
}

// IsTrue returns true if the value was set to boolean true.
func (b BoolOr[T]) IsTrue() bool {
	return b.isSet && b.boolVal && b.objectVal == nil
}

// IsFalse returns true if the value was set to boolean false or not set at all.
func (b BoolOr[T]) IsFalse() bool {
	return !b.isSet || (!b.boolVal && b.objectVal == nil)
}

// IsSet returns true if any value was explicitly provided.
func (b BoolOr[T]) IsSet() bool {
	return b.isSet
}

// Object returns the object value if one was provided, nil otherwise.
func (b BoolOr[T]) Object() *T {
	return b.objectVal
}

// Bool returns the boolean value. Returns true if explicitly set to true OR if an object was provided.
// Returns false if set to false or not set.
func (b BoolOr[T]) Bool() bool {
	return b.boolVal || b.objectVal != nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *BoolOr[T]) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		*b = BoolOr[T]{}
		return nil
	}

	// Try boolean first
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		*b = BoolOr[T]{isSet: true, boolVal: boolVal}
		return nil
	}

	// Try object
	var objectVal T
	if err := json.Unmarshal(data, &objectVal); err == nil {
		*b = BoolOr[T]{isSet: true, boolVal: true, objectVal: &objectVal}
		return nil
	}

	// If neither works, default to unset
	*b = BoolOr[T]{}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (b BoolOr[T]) MarshalJSON() ([]byte, error) {
	if !b.isSet {
		return []byte("null"), nil
	}
	if b.objectVal != nil {
		return json.Marshal(b.objectVal)
	}
	return json.Marshal(b.boolVal)
}

// BoolOrValue creates a BoolOr[T] with a boolean value.
// This is useful for setting default values in generated UnmarshalJSON methods.
func BoolOrValue[T any](val bool) BoolOr[T] {
	return BoolOr[T]{isSet: true, boolVal: val}
}
