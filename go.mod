module github.com/typescript-eslint/tsgolint

go 1.26

replace (
	github.com/microsoft/TypeScript/tsc/shim/ast => ./shim/ast
	github.com/microsoft/TypeScript/tsc/shim/bundled => ./shim/bundled
	github.com/microsoft/TypeScript/tsc/shim/checker => ./shim/checker
	github.com/microsoft/TypeScript/tsc/shim/compiler => ./shim/compiler
	github.com/microsoft/TypeScript/tsc/shim/contentmapper => ./shim/contentmapper
	github.com/microsoft/TypeScript/tsc/shim/core => ./shim/core
	github.com/microsoft/TypeScript/tsc/shim/jsnum => ./shim/jsnum
	github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto => ./shim/lsp/lsproto
	github.com/microsoft/TypeScript/tsc/shim/parser => ./shim/parser
	github.com/microsoft/TypeScript/tsc/shim/project => ./shim/project
	github.com/microsoft/TypeScript/tsc/shim/scanner => ./shim/scanner
	github.com/microsoft/TypeScript/tsc/shim/tsoptions => ./shim/tsoptions
	github.com/microsoft/TypeScript/tsc/shim/tspath => ./shim/tspath
	github.com/microsoft/TypeScript/tsc/shim/vfs => ./shim/vfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs => ./shim/vfs/cachedvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs => ./shim/vfs/osvfs
)

require (
	github.com/microsoft/TypeScript/tsc/shim/ast v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/bundled v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/checker v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/compiler v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/contentmapper v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/core v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/jsnum v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/parser v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/project v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/scanner v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tsoptions v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tspath v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs v0.0.0
	golang.org/x/sys v0.47.0
	golang.org/x/tools v0.49.0
	gotest.tools/v3 v3.5.2
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/mackerelio/go-osstat v0.2.7 // indirect
	github.com/microsoft/TypeScript/tsc v0.0.0-20260820043310-6d44e0584a85 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/mod v0.39.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
)

require (
	github.com/dlclark/regexp2/v2 v2.7.1
	github.com/go-json-experiment/json v0.0.0-20260623181947-01eb4420fa68
	golang.org/x/text v0.41.0
)
