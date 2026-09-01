package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/contentmapper"
	"github.com/microsoft/TypeScript/tsc/shim/core"

	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/typescript-eslint/tsgolint/internal/collections"
)

var _ compiler.CompilerHost = (*compilerHost)(nil)

type compilerHost struct {
	currentDirectory          string
	fs                        vfs.FS
	defaultLibraryPath        string
	extendedConfigCache       tsoptions.ExtendedConfigCache
	trace                     func(msg *ast.DiagnosticsMessage, args ...any)
	resolvedProjectReferences collections.SyncMap[tspath.Path, *tsoptions.ParsedCommandLine]
	contentMapperProject      contentmapper.Project
}

func NewCachedFSCompilerHost(
	currentDirectory string,
	fs vfs.FS,
	defaultLibraryPath string,
	extendedConfigCache tsoptions.ExtendedConfigCache,
	trace func(msg *ast.DiagnosticsMessage, args ...any),
) compiler.CompilerHost {
	return NewCompilerHost(currentDirectory, cachedvfs.From(fs), defaultLibraryPath, extendedConfigCache, trace)
}

func NewCompilerHost(
	currentDirectory string,
	fs vfs.FS,
	defaultLibraryPath string,
	extendedConfigCache tsoptions.ExtendedConfigCache,
	trace func(msg *ast.DiagnosticsMessage, args ...any),
) compiler.CompilerHost {
	if trace == nil {
		trace = func(msg *ast.DiagnosticsMessage, args ...any) {}
	}
	return &compilerHost{
		currentDirectory:    currentDirectory,
		fs:                  fs,
		defaultLibraryPath:  defaultLibraryPath,
		extendedConfigCache: extendedConfigCache,
		trace:               trace,
	}
}

func (h *compilerHost) FS() vfs.FS {
	return h.fs
}

func (h *compilerHost) DefaultLibraryPath() string {
	return h.defaultLibraryPath
}

func (h *compilerHost) GetCurrentDirectory() string {
	return h.currentDirectory
}

func (h *compilerHost) Trace(msg *ast.DiagnosticsMessage, args ...any) {
	h.trace(msg, args...)
}

var sourceFileCache collections.SyncMap[SourceFileCacheKey, *ast.SourceFile]

type SourceFileCacheKey struct {
	opts       ast.SourceFileParseOptions
	text       string
	scriptKind core.ScriptKind
}

func GetSourceFileCacheKey(opts ast.SourceFileParseOptions, text string, scriptKind core.ScriptKind) SourceFileCacheKey {
	return SourceFileCacheKey{
		opts:       opts,
		text:       text,
		scriptKind: scriptKind,
	}
}

func (h *compilerHost) GetSourceFile(opts ast.SourceFileParseOptions) *ast.SourceFile {
	text, ok := h.FS().ReadFile(opts.FileName)
	if !ok {
		return nil
	}

	scriptKind := core.GetScriptKindFromFileName(opts.FileName)
	if scriptKind == core.ScriptKindUnknown {
		panic("Unknown script kind for file  " + opts.FileName)
	}

	key := GetSourceFileCacheKey(opts, text, scriptKind)

	if cached, ok := sourceFileCache.Load(key); ok {
		return cached
	}

	sourceFile := parser.ParseSourceFile(opts, text, scriptKind)
	result, _ := sourceFileCache.LoadOrStore(key, sourceFile)
	return result
}

func (h *compilerHost) GetContentMappedSourceFiles(parseOptions ast.SourceFileParseOptions, mapper *contentmapper.Mapper) (contentmapper.SourceFiles, error) {
	if h.contentMapperProject == nil {
		return contentmapper.SourceFiles{}, contentmapper.ErrProjectUnavailable
	}
	content, ok := h.FS().ReadFile(parseOptions.FileName)
	if !ok {
		return contentmapper.SourceFiles{}, nil
	}
	files, err := contentmapper.TransformAndParse(parseOptions, content, mapper, h.contentMapperProject)
	if err == nil {
		err = contentmapper.CheckSupplementalFileNameCollisions(files, h.FS().FileExists)
	}
	return files, err
}

func (h *compilerHost) ContentMapperProject() contentmapper.Project {
	return h.contentMapperProject
}

// SetContentMapperProject installs the project-scoped mapper view. The project can only be opened once
// the tsconfig has been parsed, which needs the host, so it is attached afterwards rather than passed to
// the constructor as it is in tsc. It must be set before the program loads any file.
func SetContentMapperProject(host compiler.CompilerHost, project contentmapper.Project) {
	if h, ok := host.(*compilerHost); ok {
		h.contentMapperProject = project
	}
}

func (h *compilerHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	if cached, ok := h.resolvedProjectReferences.Load(path); ok {
		return cached
	}
	commandLine, _ := tsoptions.GetParsedCommandLineOfConfigFilePath(fileName, path, nil, nil, h, h.extendedConfigCache)
	result, _ := h.resolvedProjectReferences.LoadOrStore(path, commandLine)
	return result
}
