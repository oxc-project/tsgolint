package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

func TestClearSourceFileCache(t *testing.T) {
	dir := filepath.ToSlash(t.TempDir())
	file := dir + "/file.ts"
	assert.NilError(t, os.WriteFile(file, []byte("export const value = 1;\n"), 0o644))
	host := NewCachedFSCompilerHost(dir, bundled.WrapFS(osvfs.FS()), bundled.LibPath(), nil, nil)
	opts := ast.SourceFileParseOptions{FileName: file, Path: tspath.ToPath(file, dir, true)}

	first := host.GetSourceFile(opts)
	assert.Assert(t, first != nil)
	assert.Assert(t, host.GetSourceFile(opts) == first, "a second read returns the cached parse")

	ClearSourceFileCache()
	assert.Equal(t, sourceFileCache.Size(), 0)
	assert.Assert(t, host.GetSourceFile(opts) != first, "after clearing, the file is parsed again")
}
