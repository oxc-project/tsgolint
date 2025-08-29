package main

import (
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/vfs"
)

func newProjectSession(fs vfs.FS, cwd string) *project.Session {
	sessionOptions := &project.SessionOptions{
		CurrentDirectory:   cwd,
		DefaultLibraryPath: bundled.LibPath(),
		TypingsLocation:    "",
		PositionEncoding:   lsproto.PositionEncodingKindUTF8,
		WatchEnabled:       false,
		LoggingEnabled:     false,
	}

	sessionInit := &project.SessionInit{
		Options: sessionOptions,
		FS:      fs,
		Client:  nil,
		Logger:  nil,
	}

	return project.NewSession(sessionInit)
}
