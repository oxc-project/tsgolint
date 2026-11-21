// Package bundled provides compressed TypeScript standard library declarations.
package bundled

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/typescript-go/shim/vfs"
)

const scheme = "bundled:///"

//go:embed libs.zip
var archive string

type archivedLibrary struct {
	file     *zip.File
	contents func() string
}

var (
	libraries      map[string]*archivedLibrary
	libraryNames   []string
	libraryEntries []fs.DirEntry
)

func init() {
	zipReader, err := zip.NewReader(strings.NewReader(archive), int64(len(archive)))
	if err != nil {
		panic(fmt.Sprintf("bundled: read embedded archive: %v", err))
	}

	libraries = make(map[string]*archivedLibrary, len(zipReader.File))
	libraryNames = make([]string, 0, len(zipReader.File))
	for _, file := range zipReader.File {
		name, ok := strings.CutPrefix(file.Name, "libs/")
		if !ok || name == "" || strings.Contains(name, "/") || !strings.HasSuffix(name, ".d.ts") {
			panic(fmt.Sprintf("bundled: invalid archive entry %q", file.Name))
		}
		if _, exists := libraries[file.Name]; exists {
			panic(fmt.Sprintf("bundled: duplicate archive entry %q", file.Name))
		}

		zipFile := file
		library := &archivedLibrary{file: zipFile}
		library.contents = sync.OnceValue(func() string {
			entry, err := zipFile.Open()
			if err != nil {
				panic(fmt.Sprintf("bundled: open %q: %v", zipFile.Name, err))
			}
			defer entry.Close()

			var contents strings.Builder
			contents.Grow(int(zipFile.UncompressedSize64))
			if _, err := io.Copy(&contents, entry); err != nil {
				panic(fmt.Sprintf("bundled: decompress %q: %v", zipFile.Name, err))
			}
			return contents.String()
		})

		libraries[file.Name] = library
		libraryNames = append(libraryNames, name)
	}
	if len(libraries) == 0 {
		panic("bundled: embedded archive is empty")
	}

	slices.Sort(libraryNames)
	libraryEntries = make([]fs.DirEntry, 0, len(libraryNames))
	for _, name := range libraryNames {
		libraryEntries = append(libraryEntries, &fileInfo{
			name: name,
			size: int64(libraries["libs/"+name].file.UncompressedSize64),
		})
	}
}

func splitPath(path string) (rest string, ok bool) {
	return strings.CutPrefix(path, scheme)
}

// LibPath returns the virtual directory containing the TypeScript libraries.
func LibPath() string {
	return scheme + "libs"
}

// IsBundled reports whether path belongs to the bundled virtual filesystem.
func IsBundled(path string) bool {
	_, ok := splitPath(path)
	return ok
}

// WrapFS returns a filesystem that resolves bundled paths from the compressed archive.
func WrapFS(base vfs.FS) vfs.FS {
	return &wrappedFS{base: base}
}

type wrappedFS struct {
	base vfs.FS
}

var _ vfs.FS = (*wrappedFS)(nil)

func (w *wrappedFS) UseCaseSensitiveFileNames() bool {
	return w.base.UseCaseSensitiveFileNames()
}

func (w *wrappedFS) FileExists(path string) bool {
	if rest, ok := splitPath(path); ok {
		_, ok := libraries[rest]
		return ok
	}
	return w.base.FileExists(path)
}

func (w *wrappedFS) ReadFile(path string) (contents string, ok bool) {
	if rest, ok := splitPath(path); ok {
		library, ok := libraries[rest]
		if !ok {
			return "", false
		}
		return library.contents(), true
	}
	return w.base.ReadFile(path)
}

func (w *wrappedFS) DirectoryExists(path string) bool {
	if rest, ok := splitPath(path); ok {
		return rest == "libs"
	}
	return w.base.DirectoryExists(path)
}

func (w *wrappedFS) GetAccessibleEntries(path string) (result vfs.Entries) {
	if rest, ok := splitPath(path); ok {
		if rest == "" {
			result.Directories = []string{"libs"}
		} else if rest == "libs" {
			result.Files = libraryNames
		}
		return result
	}
	return w.base.GetAccessibleEntries(path)
}

var rootEntries = []fs.DirEntry{
	fs.FileInfoToDirEntry(&fileInfo{name: "libs", mode: fs.ModeDir}),
}

func (w *wrappedFS) Stat(path string) vfs.FileInfo {
	if rest, ok := splitPath(path); ok {
		if rest == "" || rest == "libs" {
			return &fileInfo{name: rest, mode: fs.ModeDir}
		}
		if library, ok := libraries[rest]; ok {
			name, _ := strings.CutPrefix(rest, "libs/")
			return &fileInfo{name: name, size: int64(library.file.UncompressedSize64)}
		}
		return nil
	}
	return w.base.Stat(path)
}

func (w *wrappedFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	if rest, ok := splitPath(root); ok {
		if err := w.walkDir(rest, walkFn); err != nil {
			if err == fs.SkipAll { //nolint:errorlint
				return nil
			}
			return err
		}
		return nil
	}
	return w.base.WalkDir(root, walkFn)
}

func (w *wrappedFS) walkDir(rest string, walkFn vfs.WalkDirFunc) error {
	var entries []fs.DirEntry
	switch rest {
	case "":
		entries = rootEntries
	case "libs":
		entries = libraryEntries
	default:
		return nil
	}

	for _, entry := range entries {
		name := rest + "/" + entry.Name()
		if err := walkFn(scheme+name, entry, nil); err != nil {
			if err == fs.SkipAll { //nolint:errorlint
				return fs.SkipAll
			}
			if err == fs.SkipDir { //nolint:errorlint
				continue
			}
			return err
		}
		if entry.IsDir() {
			if err := w.walkDir(strings.TrimPrefix(name, "/"), walkFn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *wrappedFS) Realpath(path string) string {
	if _, ok := splitPath(path); ok {
		return path
	}
	return w.base.Realpath(path)
}

func (w *wrappedFS) WriteFile(path string, data string) error {
	if _, ok := splitPath(path); ok {
		panic("cannot write to embedded file system")
	}
	return w.base.WriteFile(path, data)
}

func (w *wrappedFS) AppendFile(path string, data string) error {
	if _, ok := splitPath(path); ok {
		panic("cannot write to embedded file system")
	}
	return w.base.AppendFile(path, data)
}

func (w *wrappedFS) Remove(path string) error {
	if _, ok := splitPath(path); ok {
		panic("cannot remove from embedded file system")
	}
	return w.base.Remove(path)
}

func (w *wrappedFS) Chtimes(path string, accessTime time.Time, modificationTime time.Time) error {
	if _, ok := splitPath(path); ok {
		panic("cannot change times on embedded file system")
	}
	return w.base.Chtimes(path, accessTime, modificationTime)
}

type fileInfo struct {
	mode fs.FileMode
	name string
	size int64
}

var (
	_ fs.FileInfo = (*fileInfo)(nil)
	_ fs.DirEntry = (*fileInfo)(nil)
)

func (fi *fileInfo) IsDir() bool                { return fi.mode.IsDir() }
func (fi *fileInfo) ModTime() time.Time         { return time.Time{} }
func (fi *fileInfo) Mode() fs.FileMode          { return fi.mode }
func (fi *fileInfo) Name() string               { return fi.name }
func (fi *fileInfo) Size() int64                { return fi.size }
func (fi *fileInfo) Sys() any                   { return nil }
func (fi *fileInfo) Info() (fs.FileInfo, error) { return fi, nil }
func (fi *fileInfo) Type() fs.FileMode          { return fi.mode.Type() }
