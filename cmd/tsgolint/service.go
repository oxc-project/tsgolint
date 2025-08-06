package main

import (
	"io"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/vfs"
)

type singleRunHost struct {
	fs                 vfs.FS
	defaultLibraryPath string
	typingsLocation    string
	cwd                string
}

func (s *singleRunHost) FS() vfs.FS {
	return s.fs
}

func (s *singleRunHost) DefaultLibraryPath() string {
	return s.defaultLibraryPath
}

func (s *singleRunHost) TypingsLocation() string {
	return s.typingsLocation
}

func (s *singleRunHost) GetCurrentDirectory() string {
	return s.cwd
}

func (s *singleRunHost) Trace(msg string) {

}

func (s *singleRunHost) Client() project.Client {
	return nil
}

var _ project.ServiceHost = (*singleRunHost)(nil)

func newProjectService(fs vfs.FS, cwd string) *project.Service {
	host := singleRunHost{
		fs:                 fs,
		defaultLibraryPath: bundled.LibPath(),
		cwd:                cwd,
	}

	return project.NewService(&host, project.ServiceOptions{
		Logger:           project.NewLogger([]io.Writer{}, "", project.LogLevelTerse),
		PositionEncoding: lsproto.PositionEncodingKindUTF8,
		// TypingsInstallerOptions
	})
}
