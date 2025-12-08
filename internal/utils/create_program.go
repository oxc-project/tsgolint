package utils

import (
	"errors"
	"fmt"
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
	return compiler.NewCompilerHost(cwd, fs, defaultLibraryPath, nil, nil)
}

func enhanceHelpDiagnosticMessage(msg string) string {
	if strings.Contains(msg, "Please remove it from your configuration.") {
		return msg + "\n" + "See https://github.com/oxc-project/tsgolint/issues/351 for more information."

	}
	return msg
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, []diagnostic.Internal, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	configParseResult, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, nil, host, nil)

	if len(diagnostics) > 0 {
		internalDiags := make([]diagnostic.Internal, len(diagnostics))
		for i, d := range diagnostics {
			loc := d.Loc()
			var filePath string
			if d.File() != nil {
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

	if len(configParseResult.Errors) > 0 {
		internalDiags := make([]diagnostic.Internal, len(configParseResult.Errors))
		for i, e := range configParseResult.Errors {
			loc := e.Loc()
			var filePath string
			if e.File() != nil {
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
		Config:         configParseResult,
		SingleThreaded: core.TSTrue,
		Host:           host,
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
	if len(program_diagnostics) > 0 {
		// Convert ast.Diagnostic to diagnostic.Internal
		internalDiags := make([]diagnostic.Internal, len(program_diagnostics))
		for i, d := range program_diagnostics {
			loc := d.Loc()
			var filePath string
			if d.File() != nil {
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

	// TODO: report syntactic diagnostics?

	program.BindSourceFiles()

	return program, nil, nil
}

func CreateInferredProjectProgram(singleThreaded bool, fs vfs.FS, cwd string, host compiler.CompilerHost, fileNames []string) (*compiler.Program, []diagnostic.Internal, error) {
	program := compiler.NewProgram(compiler.ProgramOptions{
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
	})

	// TODO: report syntactic diagnostics?

	program.BindSourceFiles()
	return program, nil, nil
}
