package utils

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/contentmapper"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
)

// ContentMappersDisabledEnvVar turns content mappers off for a run. tsgolint enables
// `runExternalCode` by default (see RunExternalCode in create_program.go); this is the escape hatch
// for anyone who does not want a mapper's process spawned from their project's node_modules.
const ContentMappersDisabledEnvVar = "OXLINT_TSGOLINT_DISABLE_CONTENT_MAPPERS"

// ContentMappersEnabled reports whether configured content mappers may run.
var ContentMappersEnabled = sync.OnceValue(func() bool {
	return os.Getenv(ContentMappersDisabledEnvVar) != "true"
})

// spawnProcess launches a content mapper and adapts its stdio to an io.ReadWriteCloser (Read is its
// stdout, Write is its stdin). Mirrors the tsc CLI's spawner.
func spawnProcess(command []string, dir string, stderr io.Writer) (io.ReadWriteCloser, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &childProcess{cmd: cmd, stdin: stdin, stdout: stdout}, nil
}

// childProcess adapts a spawned process's stdout (read) and stdin (write) into one io.ReadWriteCloser.
// Close kills and reaps the process.
type childProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.Reader
}

func (p *childProcess) Read(b []byte) (int, error)  { return p.stdout.Read(b) }
func (p *childProcess) Write(b []byte) (int, error) { return p.stdin.Write(b) }

func (p *childProcess) ExitCode() (int, bool) {
	if p.cmd.ProcessState == nil {
		return 0, false
	}
	return p.cmd.ProcessState.ExitCode(), true
}

func (p *childProcess) Close() error {
	_ = p.stdin.Close()
	_ = p.cmd.Process.Kill()
	err := p.cmd.Wait()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) || errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}

// contentMapperHost is process-wide: mapper processes are consolidated by mapper identity, so several
// programs configuring the same mapper share one process. It is created on first use and torn down by
// ShutdownContentMappers.
var contentMapperHost = struct {
	mu       sync.Mutex
	host     contentmapper.Host
	cancel   context.CancelFunc
	projects []contentmapper.Project
}{}

func getContentMapperHost() contentmapper.Host {
	if contentMapperHost.host == nil {
		ctx, cancel := context.WithCancel(context.Background())
		contentMapperHost.cancel = cancel
		contentMapperHost.host = contentmapper.NewHostWithOptions(
			ctx,
			contentmapper.SpawnerFunc(spawnProcess),
			locale.Default,
			contentmapper.HostOptions{Logger: contentMapperLogger()},
		)
	}
	return contentMapperHost.host
}

func contentMapperLogger() contentmapper.Logger {
	if GetLogLevel() != LogLevelDebug {
		return nil
	}
	return func(message string) { log.Print(message) }
}

// OpenContentMapperProject returns the project-scoped mapper view for config, or nil when the config
// declares no (usable) content mappers. The returned project stays open until ShutdownContentMappers:
// tsgolint is a single-run process, and the program keeps loading mapped files lazily for as long as it
// is being linted.
func OpenContentMapperProject(config *tsoptions.ParsedCommandLine) contentmapper.Project {
	if config == nil || len(config.ContentMappers()) == 0 || !ContentMappersEnabled() {
		return nil
	}
	contentMapperHost.mu.Lock()
	defer contentMapperHost.mu.Unlock()
	project := getContentMapperHost().Project(contentmapper.ProjectSpec{
		ConfigFileName:  config.ConfigName(),
		Mappers:         config.ContentMappers(),
		CompilerOptions: config.CompilerOptions(),
	})
	contentMapperHost.projects = append(contentMapperHost.projects, project)
	return project
}

// ShutdownContentMappers stops every content mapper process spawned during this run. Callers that
// create programs should defer it; it is a no-op when no mapper was ever started.
func ShutdownContentMappers() {
	contentMapperHost.mu.Lock()
	defer contentMapperHost.mu.Unlock()
	for _, project := range contentMapperHost.projects {
		_ = project.Close()
	}
	contentMapperHost.projects = nil
	if contentMapperHost.host != nil {
		_ = contentMapperHost.host.Close()
		contentMapperHost.host = nil
	}
	if contentMapperHost.cancel != nil {
		contentMapperHost.cancel()
		contentMapperHost.cancel = nil
	}
}
