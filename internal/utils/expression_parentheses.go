package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Matching start positions limits traversal to the beginning of a statement
// and stops at existing parentheses.
func IsStartOfExpressionStatementNeedingParentheses(sourceFile *ast.SourceFile, node *ast.Node, firstToken ast.Kind) bool {
	if firstToken != ast.KindOpenBraceToken && firstToken != ast.KindClassKeyword && firstToken != ast.KindFunctionKeyword {
		return false
	}
	start := TrimNodeTextRange(sourceFile, node).Pos()
	for ancestor := node.Parent; ancestor != nil && TrimNodeTextRange(sourceFile, ancestor).Pos() == start; ancestor = ancestor.Parent {
		if ast.IsExpressionStatement(ancestor) {
			return true
		}
	}
	return false
}

func IsStartOfArrowFunctionBodyNeedingParentheses(sourceFile *ast.SourceFile, node *ast.Node, firstToken ast.Kind) bool {
	if firstToken != ast.KindOpenBraceToken {
		return false
	}
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			return false
		}
		if ast.IsArrowFunction(parent) && parent.Body() == current {
			return true
		}
		if TrimNodeTextRange(sourceFile, parent).Pos() != TrimNodeTextRange(sourceFile, current).Pos() {
			return false
		}
	}
	return false
}
