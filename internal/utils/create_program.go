package utils

import (
	"context"
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
		return nil, fmt.Errorf("couldn't create program")
	}

	diagnostics := program.GetSyntacticDiagnostics(context.Background(), nil)
	if len(diagnostics) != 0 {
		return nil, fmt.Errorf("found %v syntactic errors. Try running \"tsgo --noEmit\" first\n", len(diagnostics))
	}

	program.BindSourceFiles()

	return program, nil
}
