package no_unnecessary_type_assertion

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

func TestNoUnnecessaryTypeAssertion_JSDocCastOnParenthesizedExpression(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	filePath := filepath.Join(rootDir, "repro.js")
	tsconfigPath := filepath.Join(rootDir, "tsconfig.json")
	code := `/** @type {string} */
const s = "foo";

const s2 = /** @type {string} */ (s);
`
	tsconfig := `{
  "compilerOptions": {
    "allowJs": true,
    "checkJs": true,
    "noEmit": true,
    "target": "esnext",
    "module": "commonjs",
    "strict": true,
    "esModuleInterop": true,
    "lib": ["esnext"],
    "skipLibCheck": true,
    "skipDefaultLibCheck": true,
    "types": []
  },
  "include": ["repro.js"]
}
`

	assert.NilError(t, os.WriteFile(filePath, []byte(code), 0o644))
	assert.NilError(t, os.WriteFile(tsconfigPath, []byte(tsconfig), 0o644))

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(rootDir, fs)

	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host, false)
	assert.NilError(t, err)

	sourceFile := program.GetSourceFile(filePath)
	assert.Assert(t, sourceFile != nil)

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic

	err = linter.RunLinterOnProgram(
		utils.LogLevelNormal,
		program,
		[]*ast.SourceFile{sourceFile},
		1,
		func(_ *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name: NoUnnecessaryTypeAssertionRule.Name,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnnecessaryTypeAssertionRule.Run(ctx, nil)
				},
			}}
		},
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		func(diagnostic.Internal) {},
		linter.Fixes{Fix: true, FixSuggestions: true},
		linter.TypeErrors{ReportSyntactic: false, ReportSemantic: false},
	)
	assert.NilError(t, err)
	assert.Equal(t, len(diagnostics), 1)

	fixedCode, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
	assert.Assert(t, fixed)
	assert.Equal(t, fixedCode, `/** @type {string} */
const s = "foo";

const s2 = (s);
`)
}
