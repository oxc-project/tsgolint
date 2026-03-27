package linter_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/await_thenable"
	"github.com/typescript-eslint/tsgolint/internal/utils"
	"gotest.tools/v3/assert"
)

func runAwaitThenable(t *testing.T, dir string, workload linter.Workload, fs vfs.FS) ([]rule.RuleDiagnostic, []diagnostic.Internal, error) {
	t.Helper()
	var mu sync.Mutex
	var ruleDiags []rule.RuleDiagnostic
	var internalDiags []diagnostic.Internal
	err := linter.RunLinter(
		utils.LogLevelNormal, dir, workload, 1, fs,
		func(_ *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name: await_thenable.AwaitThenableRule.Name,
				Run:  func(ctx rule.RuleContext) rule.RuleListeners { return await_thenable.AwaitThenableRule.Run(ctx, nil) },
			}}
		},
		func(d rule.RuleDiagnostic) { mu.Lock(); defer mu.Unlock(); ruleDiags = append(ruleDiags, d) },
		func(d diagnostic.Internal) { mu.Lock(); defer mu.Unlock(); internalDiags = append(internalDiags, d) },
		linter.Fixes{}, linter.TypeErrors{}, false,
	)
	return ruleDiags, internalDiags, err
}

// No tsconfig at all — hardcoded defaults, rule must still fire on invalid code.
func TestInferredProgramNoNearestTsconfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "standalone.ts"), []byte("export async function test() {\n\tawait 0;\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	file := tspath.NormalizeSlashes(filepath.Join(dir, "standalone.ts"))
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	workload := linter.Workload{
		Programs:       make(map[string][]string),
		UnmatchedFiles: map[string][]string{"": {file}},
	}

	ruleDiags, internalDiags, err := runAwaitThenable(t, dir, workload, fs)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(ruleDiags), "await-thenable must fire on `await 0`")
	assert.Equal(t, 0, len(internalDiags))
}

// File outside tsconfig include inherits lib from nearest tsconfig.
// Without this fix, await-thenable false-positives on `await using` because
// the inferred program lacks lib.esnext.disposable.
func TestInferredProgramInheritsLibFromNearestTsconfig(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("tsconfig.json", `{
  "compilerOptions": { "lib": ["esnext"], "target": "esnext", "strict": true, "noEmit": true },
  "include": ["src"]
}
`)
	write("src/placeholder.ts", "export {};\n")
	write("tests/await-using.ts", `interface Resource extends AsyncDisposable { data: string }
declare const r: Resource;
export async function dispose() { await using _ = r; }
`)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	testFile := tspath.NormalizeSlashes(filepath.Join(dir, "tests", "await-using.ts"))
	resolver := utils.NewTsConfigResolver(fs, dir)
	result := resolver.FindTsConfigParallel([]string{testFile})

	res := result[testFile]
	assert.Equal(t, "", res.Config)
	tsconfig := tspath.NormalizeSlashes(filepath.Join(dir, "tsconfig.json"))
	assert.Equal(t, tsconfig, res.NearestConfig)

	workload := linter.Workload{
		Programs:       make(map[string][]string),
		UnmatchedFiles: map[string][]string{res.NearestConfig: {testFile}},
	}

	ruleDiags, internalDiags, err := runAwaitThenable(t, dir, workload, fs)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(ruleDiags), "await-thenable must not fire when lib includes esnext.disposable")
	assert.Equal(t, 0, len(internalDiags))
}
