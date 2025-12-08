package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/typescript-eslint/tsgolint/internal/collections"
)

var _ compiler.CompilerHost = (*compilerHost)(nil)

type compilerHost struct {
	currentDirectory    string
	fs                  vfs.FS
	defaultLibraryPath  string
	extendedConfigCache tsoptions.ExtendedConfigCache
	trace               func(msg *ast.DiagnosticsMessage, args ...any)
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

func (h *compilerHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	commandLine, _ := tsoptions.GetParsedCommandLineOfConfigFilePath(fileName, path, nil, nil, h, h.extendedConfigCache)
	return commandLine
}
