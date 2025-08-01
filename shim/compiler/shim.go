
// Code generated by tools/gen_shims. DO NOT EDIT.

package compiler

import "github.com/microsoft/typescript-go/internal/ast"
import "github.com/microsoft/typescript-go/internal/compiler"
import "github.com/microsoft/typescript-go/internal/vfs"
import _ "unsafe"

type CheckerPool = compiler.CheckerPool
type CompilerHost = compiler.CompilerHost
type EmitHost = compiler.EmitHost
type EmitOptions = compiler.EmitOptions
type EmitResult = compiler.EmitResult
type FileIncludeKind = compiler.FileIncludeKind
const FileIncludeKindAutomaticTypeDirectiveFile = compiler.FileIncludeKindAutomaticTypeDirectiveFile
const FileIncludeKindImport = compiler.FileIncludeKindImport
const FileIncludeKindLibFile = compiler.FileIncludeKindLibFile
const FileIncludeKindLibReferenceDirective = compiler.FileIncludeKindLibReferenceDirective
const FileIncludeKindOutputFromProjectReference = compiler.FileIncludeKindOutputFromProjectReference
const FileIncludeKindReferenceFile = compiler.FileIncludeKindReferenceFile
const FileIncludeKindRootFile = compiler.FileIncludeKindRootFile
const FileIncludeKindSourceFromProjectReference = compiler.FileIncludeKindSourceFromProjectReference
const FileIncludeKindTypeReferenceDirective = compiler.FileIncludeKindTypeReferenceDirective
type FileIncludeReason = compiler.FileIncludeReason
//go:linkname NewCachedFSCompilerHost github.com/microsoft/typescript-go/internal/compiler.NewCachedFSCompilerHost
func NewCachedFSCompilerHost(currentDirectory string, fs vfs.FS, defaultLibraryPath string) compiler.CompilerHost
//go:linkname NewCompilerHost github.com/microsoft/typescript-go/internal/compiler.NewCompilerHost
func NewCompilerHost(currentDirectory string, fs vfs.FS, defaultLibraryPath string) compiler.CompilerHost
//go:linkname NewProgram github.com/microsoft/typescript-go/internal/compiler.NewProgram
func NewProgram(opts compiler.ProgramOptions) *compiler.Program
type Program = compiler.Program
type ProgramOptions = compiler.ProgramOptions
//go:linkname SortAndDeduplicateDiagnostics github.com/microsoft/typescript-go/internal/compiler.SortAndDeduplicateDiagnostics
func SortAndDeduplicateDiagnostics(diagnostics []*ast.Diagnostic) []*ast.Diagnostic
type SourceFileMayBeEmittedHost = compiler.SourceFileMayBeEmittedHost
type SourceMapEmitResult = compiler.SourceMapEmitResult
type WriteFileData = compiler.WriteFileData
