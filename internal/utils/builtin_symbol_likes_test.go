package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestIsSymbolFromDefaultLibrary(t *testing.T) {
	rootDir := fixtures.GetRootDir()

	getTypes := func(code string) (*compiler.Program, *ast.Symbol) {
		filePath := tspath.ResolvePath(rootDir, "file.ts")
		fs := NewOverlayVFSForFile(filePath, code)

		program, _, err := CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", CreateCompilerHost(rootDir, fs), nil)
		assert.NilError(t, err, "couldn't create program")
		sourceFile := program.GetSourceFile(filePath)
		c, done := program.GetTypeChecker(t.Context())
		defer done()
		t := c.GetTypeAtLocation(sourceFile.Statements.Nodes[0].AsTypeAliasDeclaration().Name())
		return program, checker.Type_symbol(t)
	}

	runTestForAliasDeclaration := func(code string, expected bool) {
		program, symbol := getTypes(code)
		result := IsSymbolFromDefaultLibrary(program, symbol)
		if result != expected {
			t.Errorf("Expected %v. Actual %v", expected, result)
		}
	}

	t.Run("is symbol from default library", func(t *testing.T) {
		cases := []string{
			"type Test = Array<number>;",
			"type Test = Map<string,number>;",
			"type Test = Promise<void>",
			"type Test = Error",
			"type Test = Object",
		}
		for _, code := range cases {
			t.Run(code, func(t *testing.T) {
				runTestForAliasDeclaration(code, true)
			})
		}
	})

	t.Run("is not symbol from default library", func(t *testing.T) {
		cases := []string{
			// "const test: Array<number> = [1,2,3];",
			"type Test = number;",
			"type Test = { bar: string; };",
		}
		for _, code := range cases {
			t.Run(code, func(t *testing.T) {
				runTestForAliasDeclaration(code, false)
			})
		}
	})
}