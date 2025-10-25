package utils

import (
	"errors"
	"fmt"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

func CreateCompilerHost(cwd string, fs vfs.FS) compiler.CompilerHost {
	defaultLibraryPath := bundled.LibPath()
	return compiler.NewCompilerHost(cwd, fs, defaultLibraryPath, nil, nil)
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	configParseResult, _ := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, host, nil)

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
		return nil, errors.New("couldn't create program")
	}

	programDiagnostics := program.GetProgramDiagnostics()
	if len(programDiagnostics) != 0 {
		return nil, fmt.Errorf("found %v configuration errors. Run `tsgo --noEmit` to see details", len(programDiagnostics))
	}

	// TODO: report syntactic diagnostics?

	program.BindSourceFiles()

	return program, nil
}

func CreateInferredProjectProgram(singleThreaded bool, fs vfs.FS, cwd string, host compiler.CompilerHost, fileNames []string) (*compiler.Program, error) {
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
	return program, nil
}
