package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// GetAwaitTokenRemovalRange includes trailing whitespace but preserves comments.
func GetAwaitTokenRemovalRange(sourceFile *ast.SourceFile, pos int) core.TextRange {
	tokenRange := scanner.GetRangeOfTokenAtPosition(sourceFile, pos)
	end := scanner.SkipTriviaEx(sourceFile.Text(), tokenRange.End(), &scanner.SkipTriviaOptions{StopAtComments: true})
	return tokenRange.WithEnd(end)
}
