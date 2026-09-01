package no_unnecessary_type_assertion

import (
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"

	"gotest.tools/v3/assert"
)

// Reporting TypeScript's own type errors alongside lint diagnostics checks every
// file before the rules run, which leaves the contextually inferred signature of
// every call cached on its node. Rules must not report differently because of it.
func runRuleWithTypeErrors(t *testing.T, code string, reportSemantic bool) []rule.RuleDiagnostic {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "type-errors.ts")

	fs := utils.NewOverlayVFS(cachedvfs.From(bundled.WrapFS(osvfs.FS())), map[string]string{filePath: code})
	host := utils.CreateCompilerHost(rootDir, fs)
	program, _, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.minimal.json", host, false)
	assert.NilError(t, err, "couldn't create program")

	var mu sync.Mutex
	diagnostics := make([]rule.RuleDiagnostic, 0, 1)

	err = linter.RunLinterOnProgram(linter.RunLinterOnProgramOptions{
		LogLevel: utils.LogLevelNormal,
		Program:  program,
		Files:    []*ast.SourceFile{program.GetSourceFile(filePath)},
		Workers:  1,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name: NoUnnecessaryTypeAssertionRule.Name,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnnecessaryTypeAssertionRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			diagnostics = append(diagnostics, d)
		},
		OnInternalDiagnostic: func(d diagnostic.Internal) {},
		Fixes:                linter.Fixes{Fix: true, FixSuggestions: true},
		TypeErrors:           linter.TypeErrors{ReportSyntactic: false, ReportSemantic: reportSemantic},
	})
	assert.NilError(t, err, "error running linter")

	return diagnostics
}

func TestNoUnnecessaryTypeAssertionWithTypeErrorsReported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			// https://github.com/typescript-eslint/typescript-eslint/issues/6951
			// `T` is only inferrable from the return position, so it is inferred from
			// the assertion. Removing the assertion falls back to `T = Base`, and
			// `v.extra` stops compiling.
			name: "type argument inferred from the assertion",
			code: `
interface Base {
  id: string;
}
interface Derived extends Base {
  extra: number;
}
declare function query<T extends Base = Base>(key: string): T;
export const run = (): number => {
  const v = query('k') as Derived;
  return v.extra;
};
`,
			expected: 0,
		},
		{
			name: "awaited type argument inferred from the assertion",
			code: `
declare const db: { list<T = unknown>(): Promise<Map<string, T>> };
export async function get(): Promise<Map<string, Uint8Array>> {
  return (await db.list()) as Map<string, Uint8Array>;
}
`,
			expected: 0,
		},
		{
			name: "tagged template type argument inferred from the assertion",
			code: `
interface Element { tagName: string; }
interface HTMLCanvasElement extends Element { getContext(contextId: string): unknown; }
declare const query: { <E extends Element = Element>(strings: TemplateStringsArray): E | null };
export const a = query` + "`" + `.foo` + "`" + ` as HTMLCanvasElement | null;
`,
			expected: 0,
		},
		{
			name: "doubly awaited type argument inferred from the assertion",
			code: `
declare function load<T = unknown>(): Promise<Promise<T>>;
export async function main() {
  const actual = (await await load()) as Record<string, unknown>;
  return { ...actual };
}
`,
			expected: 0,
		},
		{
			name: "overloaded call with a return-only type argument",
			code: `
interface Element { tagName: string; }
interface HTMLCanvasElement extends Element { getContext(contextId: string): unknown; }
interface HTMLElementTagNameMap { canvas: HTMLCanvasElement }
declare const document: {
  querySelector<K extends keyof HTMLElementTagNameMap>(selectors: K): HTMLElementTagNameMap[K] | null;
  querySelector<E extends Element = Element>(selectors: string): E | null;
};
export const a = document.querySelector('.foo') as HTMLCanvasElement | null;
`,
			expected: 0,
		},
		{
			// Control: the type argument comes from the argument, not the assertion.
			name: "type argument inferred from an argument",
			code: `
declare function identity<T>(value: T): T;
declare const s: string;
export const value = identity(s) as string;
`,
			expected: 1,
		},
		{
			// Control: nothing generic involved.
			name: "assertion to the same type",
			code: `
declare const s: string;
export const value = s as string;
`,
			expected: 1,
		},
	}

	for _, test := range tests {
		for _, reportSemantic := range []bool{false, true} {
			name := test.name
			if reportSemantic {
				name += " (with type errors)"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				diagnostics := runRuleWithTypeErrors(t, test.code, reportSemantic)
				assert.Equal(t, len(diagnostics), test.expected, "unexpected diagnostic count for:\n%v", test.code)
			})
		}
	}
}
