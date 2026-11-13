package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
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
			}, "/virtual")
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

	overlay := newOverlayFS(baseFS, overrides, "/tmp")

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

	overlay := newOverlayFS(baseFS, overrides, "/tmp")

	exists := overlay.FileExists("/nonexistent/file.ts")
	if exists {
		t.Error("Expected non-overridden non-existent file to not exist")
	}
}

func TestOverlayFSRelativeOverrideKey(t *testing.T) {
	// Override keys are resolved against the working directory, as the
	// payload's file paths are.
	overlay := newOverlayFS(osvfs.FS(), map[string]string{
		"nested/file.ts": "editor content",
	}, "/repo")

	content, ok := overlay.ReadFile("/repo/nested/file.ts")
	if !ok || content != "editor content" {
		t.Errorf("ReadFile(\"/repo/nested/file.ts\") = %q, %v", content, ok)
	}
}

func TestOverlayFSOverrideKeysNamingTheSameFile(t *testing.T) {
	// The keys are walked in byte order, so "file.ts" wins over "/repo/file.ts"
	// and "./file.ts".
	overlay := newOverlayFS(osvfs.FS(), map[string]string{
		"/repo/file.ts": "absolute",
		"./file.ts":     "dot slash",
		"file.ts":       "bare",
	}, "/repo")

	content, ok := overlay.ReadFile("/repo/file.ts")
	if !ok || content != "bare" {
		t.Errorf("ReadFile(\"/repo/file.ts\") = %q, %v, expected %q", content, ok, "bare")
	}
}

func TestOverlayFSOverrideKeyCollisionLogging(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		overrides       map[string]string
		expectedWarning string
	}{
		{
			name: "different contents are worth a warning",
			overrides: map[string]string{
				"file.ts":       "bare",
				"/repo/file.ts": "absolute",
			},
			expectedWarning: `WARNING: source overrides "/repo/file.ts" and "file.ts" name the same file with different contents: "file.ts" wins`,
		},
		{
			name: "identical contents are not",
			overrides: map[string]string{
				"file.ts":       "same",
				"/repo/file.ts": "same",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var logged bytes.Buffer
			log.SetOutput(&logged)
			t.Cleanup(func() { log.SetOutput(os.Stderr) })

			newOverlayFS(osvfs.FS(), testCase.overrides, "/repo")

			if testCase.expectedWarning == "" {
				// The debug line this may log depends on OXC_LOG.
				if strings.Contains(logged.String(), "WARNING") {
					t.Errorf("expected no warning, got %q", logged.String())
				}
				return
			}
			if !strings.Contains(logged.String(), testCase.expectedWarning) {
				t.Errorf("expected %q to contain %q", logged.String(), testCase.expectedWarning)
			}
		})
	}
}
