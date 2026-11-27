package utils

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

const referencedProjectOptions = `"composite": true, "emitDeclarationOnly": true, "strict": true, "module": "esnext", "moduleResolution": "bundler", "rootDir": "src", "outDir": "dist"`

// app -> lib -> core, with lib and core already built to dist/.
var projectReferenceFiles = map[string]string{
	"app/tsconfig.json": `{
	"compilerOptions": { "strict": true, "module": "esnext", "moduleResolution": "bundler", "noEmit": true },
	"include": ["src"],
	"references": [{ "path": "../lib" }]
}`,
	"app/src/index.ts": "import { getValue } from '../../lib/src/index';\nexport const value: number = getValue().id;\n",
	"lib/tsconfig.json": `{
	"compilerOptions": { ` + referencedProjectOptions + ` },
	"include": ["src"],
	"references": [{ "path": "../core" }]
}`,
	"lib/src/index.ts":    "import { createValue, type Value } from '../../core/src/index';\nexport function getValue(): Value { return createValue(); }\n",
	"lib/dist/index.d.ts": "import { type Value } from '../../core/src/index';\nexport declare function getValue(): Value;\n",
	"core/tsconfig.json": `{
	"compilerOptions": { ` + referencedProjectOptions + ` },
	"include": ["src"]
}`,
	"core/src/index.ts":    "export interface Value { id: number }\nexport function createValue(): Value { return { id: 1 }; }\n",
	"core/dist/index.d.ts": "export interface Value {\n    id: number;\n}\nexport declare function createValue(): Value;\n",
}

func writeProjectReferenceFixture(t *testing.T, skip ...string) string {
	t.Helper()
	dir := tspath.NormalizePath(t.TempDir())
	for name, content := range projectReferenceFiles {
		if slices.Contains(skip, name) {
			continue
		}
		path := filepath.Join(dir, name)
		assert.NilError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NilError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

func createProjectReferenceProgram(t *testing.T, dir string, suppressProgramDiagnostics bool, useProjectReferenceOutputs bool) (*compiler.Program, []string) {
	t.Helper()
	fs := bundled.WrapFS(osvfs.FS())
	cwd := tspath.CombinePaths(dir, "app")
	program, diags, err := CreateProgram(true, fs, cwd, "tsconfig.json", CreateCompilerHost(cwd, fs), suppressProgramDiagnostics, useProjectReferenceOutputs)
	assert.NilError(t, err)
	messages := make([]string, len(diags))
	for i, d := range diags {
		messages[i] = d.Description + ": " + d.Help
	}
	return program, messages
}

func assertProgramFiles(t *testing.T, program *compiler.Program, dir string, included []string, excluded []string) {
	t.Helper()
	for _, name := range included {
		assert.Assert(t, program.GetSourceFile(tspath.CombinePaths(dir, name)) != nil, "expected %s in program", name)
	}
	for _, name := range excluded {
		assert.Assert(t, program.GetSourceFile(tspath.CombinePaths(dir, name)) == nil, "expected %s not in program", name)
	}
}

func assertNoSemanticDiagnostics(t *testing.T, program *compiler.Program, dir string) {
	t.Helper()
	sourceFile := program.GetSourceFile(tspath.CombinePaths(dir, "app/src/index.ts"))
	diags := compiler.Program_GetSemanticDiagnostics(program, t.Context(), sourceFile)
	assert.Equal(t, len(diags), 0)
}

func TestCreateProgramProjectReferences(t *testing.T) {
	t.Parallel()

	sources := []string{"lib/src/index.ts", "core/src/index.ts"}
	outputs := []string{"lib/dist/index.d.ts", "core/dist/index.d.ts"}

	t.Run("uses sources of referenced projects by default", func(t *testing.T) {
		t.Parallel()
		dir := writeProjectReferenceFixture(t)
		program, diags := createProjectReferenceProgram(t, dir, false, false)
		assert.Equal(t, len(diags), 0, diags)
		assertProgramFiles(t, program, dir, sources, outputs)
		assertNoSemanticDiagnostics(t, program, dir)
	})

	t.Run("uses declaration outputs of referenced projects when enabled", func(t *testing.T) {
		t.Parallel()
		dir := writeProjectReferenceFixture(t)
		program, diags := createProjectReferenceProgram(t, dir, false, true)
		assert.Equal(t, len(diags), 0, diags)
		assertProgramFiles(t, program, dir, outputs, sources)
		assertNoSemanticDiagnostics(t, program, dir)
	})

	t.Run("reports transitive referenced projects with missing outputs", func(t *testing.T) {
		t.Parallel()
		dir := writeProjectReferenceFixture(t, "core/dist/index.d.ts")
		program, diags := createProjectReferenceProgram(t, dir, false, true)
		assert.Assert(t, program == nil)
		assert.Equal(t, len(diags), 1, diags)
		assert.Assert(t, strings.Contains(diags[0], "Referenced project '../core/tsconfig.json' has not been built"), diags[0])
		assert.Assert(t, strings.Contains(diags[0], "Output file '../core/dist/index.d.ts' has not been built from source file '../core/src/index.ts'"), diags[0])
		assert.Assert(t, strings.Contains(diags[0], UseProjectReferenceOutputsEnvVar), diags[0])
	})

	t.Run("does not require outputs by default", func(t *testing.T) {
		t.Parallel()
		dir := writeProjectReferenceFixture(t, "lib/dist/index.d.ts", "core/dist/index.d.ts")
		program, diags := createProjectReferenceProgram(t, dir, false, false)
		assert.Equal(t, len(diags), 0, diags)
		assertProgramFiles(t, program, dir, sources, nil)
	})

	t.Run("creates the program when missing outputs are suppressed", func(t *testing.T) {
		t.Parallel()
		dir := writeProjectReferenceFixture(t, "lib/dist/index.d.ts")
		program, diags := createProjectReferenceProgram(t, dir, true, true)
		assert.Equal(t, len(diags), 0, diags)
		assert.Assert(t, program != nil)
		assertProgramFiles(t, program, dir, nil, []string{"lib/src/index.ts", "lib/dist/index.d.ts"})
	})
}
