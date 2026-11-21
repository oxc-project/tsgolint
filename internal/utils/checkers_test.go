package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestCreateProgramCheckerCountOverride(t *testing.T) {
	rootDir := t.TempDir()
	configPath := filepath.Join(rootDir, "tsconfig.json")
	assert.NilError(t, os.WriteFile(filepath.Join(rootDir, "a.ts"), []byte("export const a = 1;\n"), 0o644))
	assert.NilError(t, os.WriteFile(configPath, []byte(`{"compilerOptions":{"checkers":3},"include":["*.ts"]}`), 0o644))

	for _, tc := range []struct {
		name     string
		env      string
		expected int
	}{
		{name: "unset keeps tsconfig", env: "", expected: 3},
		{name: "override", env: "1", expected: 1},
		{name: "zero is ignored", env: "0", expected: 3},
		{name: "non-numeric is ignored", env: "two", expected: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OXLINT_TSGOLINT_CHECKERS", tc.env)
			fs := bundled.WrapFS(osvfs.FS())
			program, diagnostics, err := CreateProgram(false, fs, rootDir, configPath, CreateCompilerHost(rootDir, fs), false)
			assert.NilError(t, err)
			assert.Equal(t, len(diagnostics), 0)
			assert.Equal(t, *program.Options().Checkers, tc.expected)
		})
	}
}

func TestCreateInferredProjectProgramCheckerCountOverride(t *testing.T) {
	rootDir := t.TempDir()
	fileName := filepath.Join(rootDir, "a.ts")
	assert.NilError(t, os.WriteFile(fileName, []byte("export const a = 1;\n"), 0o644))
	t.Setenv("OXLINT_TSGOLINT_CHECKERS", "2")

	fs := bundled.WrapFS(osvfs.FS())
	program, _, err := CreateInferredProjectProgram(false, fs, rootDir, CreateCompilerHost(rootDir, fs), []string{fileName})
	assert.NilError(t, err)
	assert.Equal(t, *program.Options().Checkers, 2)
}
