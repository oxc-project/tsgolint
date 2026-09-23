package await_thenable

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// A leading parenthesis can continue the previous statement across a newline.
// Only insert a separator between sibling statements; inserting one inside an
// unbraced if/loop body would turn that body into an empty statement.
func needsPrecedingSemicolon(sourceFile *ast.SourceFile, node *ast.Node) bool {
	start := utils.TrimNodeTextRange(sourceFile, node).Pos()
	for statement := node; statement != nil && utils.TrimNodeTextRange(sourceFile, statement).Pos() == start; statement = statement.Parent {
		if !ast.IsExpressionStatement(statement) {
			continue
		}
		if statement.Parent == nil || !statement.Parent.CanHaveStatements() {
			return false
		}
		statements := statement.Parent.Statements()
		index, found := slices.BinarySearchFunc(statements, statement, func(a, b *ast.Node) int {
			return a.Pos() - b.Pos()
		})
		if !found || index == 0 {
			return false
		}
		previous := statements[index-1]
		return previous.End() > previous.Pos() && sourceFile.Text()[previous.End()-1] != ';'
	}
	return false
}
