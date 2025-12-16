package utils

import (
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

type TsConfigResolver struct {
	fs               vfs.FS
	currentDirectory string
}

func NewTsConfigResolver(fs vfs.FS, currentDirectory string) *TsConfigResolver {
	return &TsConfigResolver{
		fs:               fs,
		currentDirectory: currentDirectory,
	}
}

// Finds the tsconfig.json that governs the given file by walking up the directory tree
func (r *TsConfigResolver) FindTsconfigForFile(filePath string, skipSearchInDirectoryOfFile bool) (configPath string, found bool) {
	// Normalize the file path
	normalizedPath := tspath.NormalizeSlashes(filePath)
	if !filepath.IsAbs(normalizedPath) {
		normalizedPath = filepath.Join(r.currentDirectory, normalizedPath)
	}

	// Start from the file's directory (or parent if skipSearchInDirectoryOfFile is true)
	dir := filepath.Dir(normalizedPath)
	if skipSearchInDirectoryOfFile {
		dir = filepath.Dir(dir)
	}

	// Walk up the directory tree looking for tsconfig.json
	for {
		// Try tsconfig.json first
		configPath := filepath.Join(dir, "tsconfig.json")
		if r.fileExists(configPath) {
			return configPath, true
		}

		// Try jsconfig.json as fallback
		jsconfigPath := filepath.Join(dir, "jsconfig.json")
		if r.fileExists(jsconfigPath) {
			return jsconfigPath, true
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root
			break
		}

		// Stop if we've gone above the current directory (don't search outside the project)
		if !strings.HasPrefix(parent, r.currentDirectory) {
			break
		}

		dir = parent
	}

	return "", false
}

func (r *TsConfigResolver) fileExists(path string) bool {
	return r.fs.FileExists(path)
}

func (r *TsConfigResolver) FS() vfs.FS {
	return r.fs
}

func (r *TsConfigResolver) GetCurrentDirectory() string {
	return r.currentDirectory
}
