package utils

import (
	"iter"
	"slices"
	"unicode"

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
