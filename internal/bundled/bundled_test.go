package bundled

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func TestEmbeddedLibrariesMatchTypescriptGo(t *testing.T) {
	base := osvfs.FS()
	wrapped := WrapFS(base)
	sourceDirectory := sourceLibraryDirectory(t)

	sourceEntries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := make([]string, 0, len(sourceEntries))
	for _, sourceEntry := range sourceEntries {
		if sourceEntry.IsDir() || filepath.Ext(sourceEntry.Name()) != ".ts" {
			continue
		}
		wantNames = append(wantNames, sourceEntry.Name())

		want, err := os.ReadFile(filepath.Join(sourceDirectory, sourceEntry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		path := LibPath() + "/" + sourceEntry.Name()
		got, ok := wrapped.ReadFile(path)
		if !ok {
			t.Fatalf("ReadFile(%q) did not find embedded library", path)
		}
		if got != string(want) {
			t.Fatalf("ReadFile(%q) differs from TypeScript-Go source", path)
		}

		info := wrapped.Stat(path)
		if info == nil {
			t.Fatalf("Stat(%q) returned nil", path)
		}
		if info.Name() != sourceEntry.Name() || info.Size() != int64(len(want)) || info.IsDir() {
			t.Fatalf("Stat(%q) = {Name: %q, Size: %d, IsDir: %v}", path, info.Name(), info.Size(), info.IsDir())
		}
	}
	slices.Sort(wantNames)

	entries := wrapped.GetAccessibleEntries(LibPath())
	if !reflect.DeepEqual(entries.Files, wantNames) {
		t.Fatalf("GetAccessibleEntries(%q) files differ", LibPath())
	}
	if len(entries.Directories) != 0 {
		t.Fatalf("GetAccessibleEntries(%q) returned unexpected directories: %v", LibPath(), entries.Directories)
	}

	var walkedNames []string
	if err := wrapped.WalkDir(LibPath(), func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			walkedNames = append(walkedNames, entry.Name())
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(walkedNames, wantNames) {
		t.Fatal("WalkDir library names differ")
	}
}

func TestBundledFilesystemBehavior(t *testing.T) {
	base := osvfs.FS()
	wrapped := WrapFS(base)

	if !IsBundled(LibPath()) || IsBundled(t.TempDir()) {
		t.Fatal("IsBundled returned an unexpected result")
	}
	if !wrapped.DirectoryExists(LibPath()) {
		t.Fatalf("DirectoryExists(%q) = false", LibPath())
	}
	if info := wrapped.Stat(LibPath()); info == nil || !info.IsDir() {
		t.Fatalf("Stat(%q) did not return a directory", LibPath())
	}
	rootEntries := wrapped.GetAccessibleEntries(scheme)
	if !reflect.DeepEqual(rootEntries.Directories, []string{"libs"}) {
		t.Fatalf("GetAccessibleEntries(%q) = %v", scheme, rootEntries.Directories)
	}

	missing := LibPath() + "/missing.d.ts"
	if wrapped.FileExists(missing) {
		t.Fatalf("FileExists(%q) = true", missing)
	}
	if contents, ok := wrapped.ReadFile(missing); ok || contents != "" {
		t.Fatalf("ReadFile(%q) = (%q, %v)", missing, contents, ok)
	}
	if info := wrapped.Stat(missing); info != nil {
		t.Fatalf("Stat(%q) returned %v", missing, info)
	}
	if got := wrapped.Realpath(missing); got != missing {
		t.Fatalf("Realpath(%q) = %q", missing, got)
	}

	delegatedPath := filepath.Join(t.TempDir(), "delegated.ts")
	if err := os.WriteFile(delegatedPath, []byte("export {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, ok := wrapped.ReadFile(delegatedPath); !ok || got != "export {}" {
		t.Fatalf("delegated ReadFile(%q) = (%q, %v)", delegatedPath, got, ok)
	}
}

func TestBundledLibrariesCanBeReadConcurrently(t *testing.T) {
	wrapped := WrapFS(osvfs.FS())
	path := LibPath() + "/lib.dom.d.ts"
	want, ok := wrapped.ReadFile(path)
	if !ok {
		t.Fatalf("ReadFile(%q) failed", path)
	}

	const readers = 32
	var waitGroup sync.WaitGroup
	waitGroup.Add(readers)
	for range readers {
		go func() {
			defer waitGroup.Done()
			got, ok := wrapped.ReadFile(path)
			if !ok || got != want {
				t.Errorf("concurrent ReadFile(%q) returned inconsistent contents", path)
			}
		}()
	}
	waitGroup.Wait()
}

func TestBundledFilesystemRejectsMutations(t *testing.T) {
	wrapped := WrapFS(osvfs.FS())
	path := LibPath() + "/lib.d.ts"
	tests := map[string]func() error{
		"write":  func() error { return wrapped.WriteFile(path, "") },
		"append": func() error { return wrapped.AppendFile(path, "") },
		"remove": func() error { return wrapped.Remove(path) },
		"chtimes": func() error {
			return wrapped.Chtimes(path, time.Time{}, time.Time{})
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("mutation did not panic")
				}
			}()
			_ = mutate()
		})
	}
}

func sourceLibraryDirectory(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test source path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "typescript-go", "internal", "bundled", "libs")
}
