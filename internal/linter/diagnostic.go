package linter

import "github.com/microsoft/typescript-go/shim/core"

type InternalDiagnostic struct {
	Range       core.TextRange
	Id          string
	Description string
}
