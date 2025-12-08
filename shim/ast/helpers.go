// This file contains manual helpers for the ast shim package.
// It is NOT generated and should be preserved when regenerating shims.
//
// To regenerate shims: just shim

package ast

import (
	"github.com/microsoft/typescript-go/internal/diagnostics"
	"github.com/microsoft/typescript-go/internal/locale"
	_ "unsafe"
)

//go:linkname defaultLocale github.com/microsoft/typescript-go/internal/locale.Default
var defaultLocale locale.Locale

// DefaultLocale returns the default locale (English) for diagnostic localization.
// This is exported so that other packages can use it with Diagnostic_Localize.
func DefaultLocale() locale.Locale {
	return defaultLocale
}

// DiagnosticsMessage is an alias for the internal diagnostics.Message type
// This allows other packages to reference the type without importing internal packages
type DiagnosticsMessage = diagnostics.Message
