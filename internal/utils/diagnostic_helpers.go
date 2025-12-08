package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// GetDiagnosticMessage returns the localized message for a diagnostic using the default locale (English).
func GetDiagnosticMessage(d *ast.Diagnostic) string {
	// Use the Localize method with the default locale
	// locale.Default is accessed via the shim's linkname helper
	return ast.Diagnostic_Localize(d, ast.DefaultLocale())
}
