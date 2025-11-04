package diagnostic

import "github.com/microsoft/typescript-go/shim/core"

type Internal struct {
	Range       core.TextRange
	Id          string
	Description string
	Help        string
	FilePath    *string `json:"omitempty"`
}
