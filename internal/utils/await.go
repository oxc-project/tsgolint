package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// GetAwaitTokenRemovalRange includes trailing whitespace but preserves comments.
func GetAwaitTokenRemovalRange(sourceFile *ast.SourceFile, pos int) core.TextRange {
	tokenRange := scanner.GetRangeOfTokenAtPosition(sourceFile, pos)
	end := scanner.SkipTriviaEx(sourceFile.Text(), tokenRange.End(), &scanner.SkipTriviaOptions{StopAtComments: true})
	return tokenRange.WithEnd(end)
}
