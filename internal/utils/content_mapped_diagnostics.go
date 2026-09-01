package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// IsContentMapped reports whether file's text is virtual TypeScript produced by a content mapper, so
// that positions in it have to be mapped back before they are reported.
func IsContentMapped(file *ast.SourceFile) bool {
	return file != nil && file.SpanMap() != nil
}

// MapContentMappedRange maps a range in a content-mapped file's virtual text back to the original file.
// It reports false when the range lies entirely in synthesized code, which has no original counterpart.
// Ranges in files that are not content-mapped pass through unchanged.
//
// This is the mapping the compiler applies to its own diagnostics: only fully synthesized ranges are
// rejected, so a diagnostic anchored on generated code still points at the construct that produced it.
func MapContentMappedRange(file *ast.SourceFile, r core.TextRange) (core.TextRange, bool) {
	spanMap := file.SpanMap()
	if spanMap == nil {
		return r, true
	}
	mapped, fidelity := spanMap.VirtualToOriginalSpan(r)
	if fidelity.IsNone() {
		return r, false
	}
	return mapped, true
}

// MapContentMappedLintRange is MapContentMappedRange with the extra restrictions that apply to lint
// diagnostics.
//
// A rule fires on whatever the transform generated, including the mapper's own scaffolding, so two
// further ranges are rejected on top of a fully synthesized one:
//
//   - Ranges that map to an empty original range. A mapper that cannot express a nested original range
//     collapses the enclosing one to a zero-length anchor (ember-content-mapper does this for every
//     mapping whose original range strictly contains another). That anchor is a useful fallback for a
//     type error, but a lint finding pointing at a zero-width slice of the user's file is noise.
//   - Ranges covered by an Ignore diagnostic directive. The mapper uses these to disclaim generated
//     code it knows produces meaningless diagnostics; the compiler honours them for its own diagnostics
//     in applyContentMapperDiagnosticDirectives, and lint rules have the same reason to.
func MapContentMappedLintRange(file *ast.SourceFile, r core.TextRange) (core.TextRange, bool) {
	mapped, ok := MapContentMappedRange(file, r)
	if !ok || file.SpanMap() == nil {
		return mapped, ok
	}
	if mapped.Pos() >= mapped.End() {
		return mapped, false
	}
	for _, directive := range file.DiagnosticDirectives() {
		if directive.Policy != ast.MappedDiagnosticDirectivePolicyIgnore {
			continue
		}
		if r.Pos() >= directive.VirtualRange.Pos() && r.Pos() < directive.VirtualRange.End() {
			return mapped, false
		}
	}
	return mapped, true
}

// MapContentMappedDiagnosticRange maps a compiler diagnostic's range into the original file. A
// diagnostic the mapper produced itself already carries an original range — it is tagged with the
// mapper's own diagnostic source — so it passes through unmapped, as it does in tsc's diagnostic writer.
func MapContentMappedDiagnosticRange(file *ast.SourceFile, d *ast.Diagnostic) (core.TextRange, bool) {
	if d.Source() != "" {
		return d.Loc(), true
	}
	return MapContentMappedRange(file, d.Loc())
}

// MapContentMappedEditRange maps a range that an autofix would write back to. It reports false unless
// the range falls entirely within a single verbatim segment: only there are the virtual and original
// texts identical (the compiler validates this when it accepts the mapper's span map), so only there can
// an edit be transposed onto the original file.
func MapContentMappedEditRange(file *ast.SourceFile, r core.TextRange) (core.TextRange, bool) {
	spanMap := file.SpanMap()
	if spanMap == nil {
		return r, true
	}
	mapped, fidelity := spanMap.VirtualToOriginalSpan(r)
	if !fidelity.IsExact() {
		return r, false
	}
	return mapped, true
}
