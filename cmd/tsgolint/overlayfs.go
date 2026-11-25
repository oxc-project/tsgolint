package main

import (
	"log"
	"maps"
	"slices"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

// TODO: overrides are invisible to DirectoryExists and GetAccessibleEntries,
// unlike in utils.OverlayVFS, which this should be modeled on.
type overlayFS struct {
	underlying vfs.FS
	// cwd is the directory the override keys are resolved against, as the
	// payload's file paths are.
	cwd       string
	overrides map[string]string
}

func newOverlayFS(underlying vfs.FS, overrides map[string]string, cwd string) vfs.FS {
	fs := &overlayFS{
		underlying: underlying,
		cwd:        cwd,
		overrides:  make(map[string]string, len(overrides)),
	}

	// Several keys can name the same file, in which case a single override
	// must win. The keys are walked in byte order and the last one wins, so
	// that the same payload always resolves to the same contents.
	keptKeys := make(map[string]string, len(overrides))
	for _, key := range slices.Sorted(maps.Keys(overrides)) {
		path := fs.canonicalPath(key)

		if kept, collides := keptKeys[path]; collides {
			if overrides[kept] != overrides[key] {
				// One of the two overrides is discarded.
				log.Printf("WARNING: source overrides %q and %q name the same file with different contents: %q wins", kept, key, key)
			} else if utils.GetLogLevel() == utils.LogLevelDebug {
				log.Printf("Source overrides %q and %q name the same file: %q wins", kept, key, key)
			}
		}

		keptKeys[path] = key
		fs.overrides[path] = overrides[key]
	}

	return fs
}

func (o *overlayFS) canonicalPath(path string) string {
	return tspath.GetCanonicalFileName(tspath.GetNormalizedAbsolutePath(path, o.cwd), o.UseCaseSensitiveFileNames())
}

func (o *overlayFS) UseCaseSensitiveFileNames() bool {
	return o.underlying.UseCaseSensitiveFileNames()
}

func (o *overlayFS) FileExists(path string) bool {
	if _, ok := o.overrides[o.canonicalPath(path)]; ok {
		return true
	}
	return o.underlying.FileExists(path)
}

func (o *overlayFS) ReadFile(path string) (string, bool) {
	if content, ok := o.overrides[o.canonicalPath(path)]; ok {
		return content, true
	}
	return o.underlying.ReadFile(path)
}

func (o *overlayFS) WriteFile(path string, data string) error {
	return o.underlying.WriteFile(path, data)
}

func (o *overlayFS) AppendFile(path string, data string) error {
	return o.underlying.AppendFile(path, data)
}

func (o *overlayFS) Remove(path string) error {
	return o.underlying.Remove(path)
}

func (o *overlayFS) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	return o.underlying.Chtimes(path, aTime, mTime)
}

func (o *overlayFS) DirectoryExists(path string) bool {
	return o.underlying.DirectoryExists(path)
}

func (o *overlayFS) GetAccessibleEntries(path string) vfs.Entries {
	return o.underlying.GetAccessibleEntries(path)
}

func (o *overlayFS) Stat(path string) vfs.FileInfo {
	return o.underlying.Stat(path)
}

func (o *overlayFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	return o.underlying.WalkDir(root, walkFn)
}

func (o *overlayFS) Realpath(path string) string {
	return o.underlying.Realpath(path)
}
