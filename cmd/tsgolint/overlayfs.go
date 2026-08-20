package main

import (
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

type overlayFS struct {
	underlying vfs.FS
	overrides  map[string]string
}

func newOverlayFS(underlying vfs.FS, overrides map[string]string) vfs.FS {
	fs := &overlayFS{
		underlying: underlying,
		overrides:  make(map[string]string, len(overrides)),
	}
	for path, content := range overrides {
		fs.overrides[fs.canonicalPath(path)] = content
	}
	return fs
}

func (o *overlayFS) canonicalPath(path string) string {
	return tspath.GetCanonicalFileName(tspath.NormalizePath(path), o.UseCaseSensitiveFileNames())
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
