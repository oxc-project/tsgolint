package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestCreateProgram_AllowsNonCompositeProjectReference(t *testing.T) {
	// Composite only constrains `tsc --build` emit; type-aware linting reads
	// the referenced project's sources and never emits.
	rootDir := "/repo"
	files := map[string]string{
		rootDir + "/tsconfig.json": `{
			"compilerOptions": {
				"strict": true,
				"target": "esnext",
				"module": "esnext",
				"moduleResolution": "bundler"
			},
			"references": [{ "path": "./sub" }],
			"files": ["a.ts"],
			"exclude": ["sub"]
		}`,
		rootDir + "/a.ts": `export const a: any = 1;`,
		rootDir + "/sub/tsconfig.json": `{
			"compilerOptions": {
				"strict": true,
				"target": "esnext",
				"module": "esnext",
				"moduleResolution": "bundler"
			},
			"files": ["b.ts"]
		}`,
		rootDir + "/sub/b.ts": `export const b = 1;`,
	}
	fs := NewOverlayVFS(cachedvfs.From(bundled.WrapFS(osvfs.FS())), files)
	host := CreateCompilerHost(rootDir, fs)

	program, diagnostics, err := CreateProgram(
		true,
		fs,
		rootDir,
		"tsconfig.json",
		host,
		false,
	)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(diagnostics))
	assert.Assert(t, program != nil)
	assert.Assert(t, program.GetSourceFile(rootDir+"/a.ts") != nil)
}
