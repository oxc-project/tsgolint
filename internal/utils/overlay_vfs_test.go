package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"gotest.tools/v3/assert"
)

type overlayEntriesFS struct {
	vfs.FS
	entries vfs.Entries
}

func (fs *overlayEntriesFS) GetAccessibleEntries(string) vfs.Entries {
	return fs.entries
}

func TestOverlayVFSGetAccessibleEntriesIsolation(t *testing.T) {
	t.Parallel()
	files := make([]string, 1, 4)
	files[0] = "physical.ts"
	directories := make([]string, 1, 4)
	directories[0] = "physical"
	base := &overlayEntriesFS{entries: vfs.Entries{Files: files, Directories: directories}}

	first := NewOverlayVFS(base, map[string]string{
		"/src/first.ts":       "",
		"/src/first/index.ts": "",
	})
	second := NewOverlayVFS(base, map[string]string{
		"/src/second.ts":       "",
		"/src/second/index.ts": "",
	})

	firstEntries := first.GetAccessibleEntries("/src")
	secondEntries := second.GetAccessibleEntries("/src")

	assert.DeepEqual(t, firstEntries.Files, []string{"physical.ts", "first.ts"})
	assert.DeepEqual(t, firstEntries.Directories, []string{"physical", "first"})
	assert.DeepEqual(t, secondEntries.Files, []string{"physical.ts", "second.ts"})
	assert.DeepEqual(t, secondEntries.Directories, []string{"physical", "second"})

	firstEntries.Files[0] = "changed.ts"
	firstEntries.Directories[0] = "changed"
	assert.DeepEqual(t, base.entries.Files, []string{"physical.ts"})
	assert.DeepEqual(t, base.entries.Directories, []string{"physical"})
}
