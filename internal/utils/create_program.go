package utils

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
)

func CreateCompilerHost(cwd string, fs vfs.FS) compiler.CompilerHost {
	defaultLibraryPath := bundled.LibPath()
	return NewCompilerHost(cwd, fs, defaultLibraryPath, nil, nil)
}

func enhanceHelpDiagnosticMessage(msg string) string {
	if strings.Contains(msg, "Please remove it from your configuration.") {
		return msg + "\n" + "See https://github.com/oxc-project/tsgolint/issues/351 for more information."

	}
	return msg
}

// UseProjectReferenceOutputsEnvVar opts into reading project references' built declaration outputs instead of their sources.
const UseProjectReferenceOutputsEnvVar = "OXLINT_TSGOLINT_USE_PROJECT_REFERENCE_OUTPUTS"

// missingProjectReferenceOutputDiagnostics reports referenced projects, including transitive references,
// whose declaration outputs do not exist on disk. Without this check, imports of unbuilt projects
// would silently resolve to error types.
func missingProjectReferenceOutputDiagnostics(program *compiler.Program, configPath string) []diagnostic.Internal {
	fs := program.Host().FS()
	configDir := tspath.GetDirectoryPath(configPath)
	relative := func(fileName string) string {
		return tspath.GetRelativePathFromDirectory(configDir, fileName, tspath.ComparePathsOptions{
			CurrentDirectory:          configDir,
			UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
		})
	}
	var diags []diagnostic.Internal
	program.RangeResolvedProjectReference(func(_ tspath.Path, config *tsoptions.ParsedCommandLine, _ *tsoptions.ParsedCommandLine, _ int) bool {
		// Unresolved references are already reported as program diagnostics.
		if config == nil {
			return true
		}
		var missing []*tsoptions.SourceOutputAndProjectReference
		for path, ref := range config.SourceToProjectReference() {
			// A file in several referenced projects maps to only one of them; report it there.
			if ref.OutputDts == "" || program.GetProjectReferenceFromSource(path) != ref {
				continue
			}
			if !fs.FileExists(ref.OutputDts) {
				missing = append(missing, ref)
			}
		}
		if len(missing) == 0 {
			return true
		}
		slices.SortFunc(missing, func(a, b *tsoptions.SourceOutputAndProjectReference) int {
			return strings.Compare(a.Source, b.Source)
		})
		help := fmt.Sprintf(
			"Output file '%s' has not been built from source file '%s'",
			relative(missing[0].OutputDts), relative(missing[0].Source),
		)
		if len(missing) > 1 {
			help += fmt.Sprintf(" (and %d more)", len(missing)-1)
		}
		help += fmt.Sprintf(
			".\n%s is set, so referenced projects are read from their declaration outputs. Build referenced projects first (for example with `tsc --build`), or unset %s.",
			UseProjectReferenceOutputsEnvVar, UseProjectReferenceOutputsEnvVar,
		)
		diags = append(diags, diagnostic.Internal{
			Range:       core.UndefinedTextRange(),
			Id:          "tsconfig-error",
			Description: fmt.Sprintf("Referenced project '%s' has not been built", relative(config.ConfigName())),
			Help:        help,
			FilePath:    &configPath,
		})
		return true
	})
	return diags
}

// CreateProgram creates a program for the given tsconfig.
//
// By default, referenced projects contribute their source files, matching editor behavior.
// When useProjectReferenceOutputs is true, they contribute their built declaration outputs instead,
// matching `tsc --project`; missing outputs are reported unless program diagnostics are suppressed.
func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost, suppressProgramDiagnostics bool, useProjectReferenceOutputs bool) (*compiler.Program, []diagnostic.Internal, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	configParseResult, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, nil, host, nil)

	if len(diagnostics) > 0 {
		internalDiags := make([]diagnostic.Internal, len(diagnostics))
		for i, d := range diagnostics {
			loc := d.Loc()
			filePath := resolvedConfigPath
			if d.File() != nil && d.File().FileName() != "" {
				filePath = d.File().FileName()
			}
			internalDiags[i] = diagnostic.Internal{
				Range:       core.NewTextRange(loc.Pos(), loc.End()),
				Id:          "tsconfig-error",
				Description: "Invalid tsconfig",
				Help:        GetDiagnosticMessage(d),
				FilePath:    &filePath,
			}
		}
		return nil, internalDiags, nil
	}

	if len(configParseResult.Errors) > 0 && !suppressProgramDiagnostics {
		internalDiags := make([]diagnostic.Internal, len(configParseResult.Errors))
		for i, e := range configParseResult.Errors {
			loc := e.Loc()
			filePath := resolvedConfigPath
			if e.File() != nil && e.File().FileName() != "" {
				filePath = e.File().FileName()
			}
			internalDiags[i] = diagnostic.Internal{
				Range:       core.NewTextRange(loc.Pos(), loc.End()),
				Id:          "tsconfig-error",
				Description: "Invalid tsconfig",
				Help:        GetDiagnosticMessage(e),
				FilePath:    &filePath,
			}
		}
		return nil, internalDiags, nil
	}

	opts := compiler.ProgramOptions{
		Config:                      configParseResult,
		SingleThreaded:              core.TSTrue,
		Host:                        host,
		UseSourceOfProjectReference: !useProjectReferenceOutputs,
		// TODO: custom checker pool
		// CreateCheckerPool: func(p *compiler.Program) compiler.CheckerPool {},
	}
	if !singleThreaded {
		opts.SingleThreaded = core.TSFalse
	}
	program := compiler.NewProgram(opts)
	if program == nil {
		return nil, nil, errors.New("couldn't create program")
	}

	program_diagnostics := program.GetProgramDiagnostics()
	if len(program_diagnostics) > 0 && !suppressProgramDiagnostics {
		// Convert ast.Diagnostic to diagnostic.Internal
		internalDiags := make([]diagnostic.Internal, len(program_diagnostics))
		for i, d := range program_diagnostics {
			loc := d.Loc()
			filePath := resolvedConfigPath
			if d.File() != nil && d.File().FileName() != "" {
				filePath = d.File().FileName()
			}
			internalDiags[i] = diagnostic.Internal{
				Range:       core.NewTextRange(loc.Pos(), loc.End()),
				Id:          "tsconfig-error",
				Description: "Invalid tsconfig",
				Help:        enhanceHelpDiagnosticMessage(GetDiagnosticMessage(d)),
				FilePath:    &filePath,
			}
		}
		return nil, internalDiags, nil
	}

	if useProjectReferenceOutputs && !suppressProgramDiagnostics {
		if diags := missingProjectReferenceOutputDiagnostics(program, resolvedConfigPath); len(diags) > 0 {
			return nil, diags, nil
		}
	}

	// TODO: report syntactic diagnostics?

	program.BindSourceFiles()

	return program, nil, nil
}

func CreateInferredProjectProgram(singleThreaded bool, fs vfs.FS, cwd string, host compiler.CompilerHost, fileNames []string) (*compiler.Program, []diagnostic.Internal, error) {
	opts := compiler.ProgramOptions{
		Config: &tsoptions.ParsedCommandLine{
			ParsedConfig: &core.ParsedOptions{
				CompilerOptions: &core.CompilerOptions{
					AllowJs:                    core.TSTrue,
					Module:                     core.ModuleKindESNext,
					ModuleResolution:           core.ModuleResolutionKindBundler,
					Target:                     core.ScriptTargetES2022,
					Jsx:                        core.JsxEmitReactJSX,
					AllowImportingTsExtensions: core.TSTrue,
					StrictNullChecks:           core.TSTrue,
					StrictFunctionTypes:        core.TSTrue,
					SourceMap:                  core.TSTrue,
					ESModuleInterop:            core.TSTrue,
					AllowNonTsExtensions:       core.TSTrue,
					ResolveJsonModule:          core.TSTrue,
				},
				FileNames: fileNames,
			},
		},
		SingleThreaded: core.TSTrue,
		Host:           host,
	}
	if !singleThreaded {
		opts.SingleThreaded = core.TSFalse
	}
	program := compiler.NewProgram(opts)

	// TODO: report syntactic diagnostics?

	program.BindSourceFiles()
	return program, nil, nil
}
