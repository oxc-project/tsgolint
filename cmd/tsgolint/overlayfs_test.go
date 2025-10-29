package main

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

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
