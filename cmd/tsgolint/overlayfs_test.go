package main

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

type caseSensitivityFS struct {
	vfs.FS
	caseSensitive bool
}

func (fs caseSensitivityFS) UseCaseSensitiveFileNames() bool { return fs.caseSensitive }

func TestOverlayFSFileNameCasing(t *testing.T) {
	for _, caseSensitive := range []bool{false, true} {
		name := "case-insensitive"
		if caseSensitive {
			name = "case-sensitive"
		}
		t.Run(name, func(t *testing.T) {
			overlay := newOverlayFS(caseSensitivityFS{osvfs.FS(), caseSensitive}, map[string]string{
				"/virtual/Project/File.ts": "editor content",
			})
			path := "/virtual/project/file.ts"
			if got := overlay.FileExists(path); got != !caseSensitive {
				t.Errorf("FileExists(%q) = %v", path, got)
			}
			content, ok := overlay.ReadFile(path)
			if ok != !caseSensitive || (ok && content != "editor content") {
				t.Errorf("ReadFile(%q) = %q, %v", path, content, ok)
			}
		})
	}
}

func TestOverlayFS(t *testing.T) {
	baseFS := osvfs.FS()
	overrides := map[string]string{
		"/tmp/test.ts": "const x: number = 42;",
	}

	overlay := newOverlayFS(baseFS, overrides)

	content, ok := overlay.ReadFile("/tmp/test.ts")
	if !ok {
		t.Fatal("Expected to read overridden file")
	}

	if content != "const x: number = 42;" {
		t.Errorf("Expected 'const x: number = 42;', got %q", content)
	}

	if !overlay.FileExists("/tmp/test.ts") {
		t.Error("Expected file to exist")
	}

	if overlay.UseCaseSensitiveFileNames() != baseFS.UseCaseSensitiveFileNames() {
		t.Error("Expected UseCaseSensitiveFileNames to match base FS")
	}
}

func TestOverlayFSFallthrough(t *testing.T) {
	baseFS := osvfs.FS()
	overrides := map[string]string{
		"/tmp/override.ts": "overridden",
	}

	overlay := newOverlayFS(baseFS, overrides)

	exists := overlay.FileExists("/nonexistent/file.ts")
	if exists {
		t.Error("Expected non-overridden non-existent file to not exist")
	}
}
