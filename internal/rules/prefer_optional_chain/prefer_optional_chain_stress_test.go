package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainStressTests tests stress scenarios with very deep chains,
// very long expressions, and many chains in one statement
func TestPreferOptionalChainStressTests(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Very deep chains should not convert if different variables
		{Code: `declare const a: {b: any}; declare const c: {d: any}; a.b && a.b.c && c.d && c.d.e;`},
		// Chain in JSX - should NOT convert due to semantic differences
		// In JSX, foo && foo.bar returns false/null/undefined (for conditional rendering)
		// while foo?.bar always returns undefined (different rendering behavior)
		{Code: `declare const a: {b: {c: string} | null} | null; const element = <div>{a && a.b && a.b.c}</div>;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// 8-level deep chain
		{
			Code:   `declare const a: {b: {c: {d: {e: {f: {g: {h: number} | null} | null} | null} | null} | null} | null} | null; a && a.b && a.b.c && a.b.c.d && a.b.c.d.e && a.b.c.d.e.f && a.b.c.d.e.f.g && a.b.c.d.e.f.g.h;`,
			Output: []string{`declare const a: {b: {c: {d: {e: {f: {g: {h: number} | null} | null} | null} | null} | null} | null} | null; a?.b?.c?.d?.e?.f?.g?.h;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// 10-level deep chain
		{
			Code:   `declare const x: {a: {b: {c: {d: {e: {f: {g: {h: {i: {j: number} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null; x != null && x.a != null && x.a.b != null && x.a.b.c != null && x.a.b.c.d != null && x.a.b.c.d.e != null && x.a.b.c.d.e.f != null && x.a.b.c.d.e.f.g != null && x.a.b.c.d.e.f.g.h != null && x.a.b.c.d.e.f.g.h.i != null && x.a.b.c.d.e.f.g.h.i.j;`,
			Output: []string{`declare const x: {a: {b: {c: {d: {e: {f: {g: {h: {i: {j: number} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null; x?.a?.b?.c?.d?.e?.f?.g?.h?.i?.j;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple separate chains in one expression (should convert each independently)
		{
			Code:   `declare const a: {b: string} | null; declare const c: {d: number} | null; a && a.b && c && c.d;`,
			Output: []string{`declare const a: {b: string} | null; declare const c: {d: number} | null; a?.b && c?.d;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Three separate chains
		{
			Code:   `declare const a: {b: string} | null; declare const c: {d: number} | null; declare const e: {f: boolean} | null; a && a.b && c && c.d && e && e.f;`,
			Output: []string{`declare const a: {b: string} | null; declare const c: {d: number} | null; declare const e: {f: boolean} | null; a?.b && c?.d && e?.f;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Very long expression with mixed call/member/element access
		{
			Code:   `declare const foo: {bar: (() => {baz: {[key: string]: (() => {qux: {fizz: (() => number) | null} | null}) | null} | null}) | null} | null; foo && foo.bar && foo.bar() && foo.bar().baz && foo.bar().baz['key'] && foo.bar().baz['key']() && foo.bar().baz['key']().qux && foo.bar().baz['key']().qux.fizz && foo.bar().baz['key']().qux.fizz();`,
			Output: []string{`declare const foo: {bar: (() => {baz: {[key: string]: (() => {qux: {fizz: (() => number) | null} | null}) | null} | null}) | null} | null; foo?.bar?.()?.baz?.['key']?.()?.qux?.fizz?.();`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Deep chain with all different check types mixed
		{
			Code:   `declare const x: {a: {b: {c: {d: {e: number} | null} | undefined} | null} | undefined} | null; x != null && typeof x.a !== 'undefined' && null !== x.a.b && x.a.b.c !== undefined && x.a.b.c.d !== null && x.a.b.c.d.e;`,
			Output: []string{`declare const x: {a: {b: {c: {d: {e: number} | null} | undefined} | null} | undefined} | null; x?.a?.b?.c?.d?.e;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Very long OR chain with negations
		{
			Code:   `declare const a: {b: {c: {d: {e: {f: number} | null} | null} | null} | null} | null; !a || !a.b || !a.b.c || !a.b.c.d || !a.b.c.d.e || !a.b.c.d.e.f;`,
			Output: []string{`declare const a: {b: {c: {d: {e: {f: number} | null} | null} | null} | null} | null; !a?.b?.c?.d?.e?.f;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Many repeated access patterns (should optimize)
		{
			Code:   `declare const config: {api: {endpoint: string} | null} | null; const a = config && config.api && config.api.endpoint; const b = config && config.api && config.api.endpoint;`,
			Output: []string{`declare const config: {api: {endpoint: string} | null} | null; const a = config?.api?.endpoint; const b = config?.api?.endpoint;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Nested ternary with chains
		{
			Code:   `declare const a: {b: {c: string} | null} | null; const result = a && a.b ? a.b.c : null;`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; const result = a?.b ? a.b.c : null;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in array literal
		{
			Code:   `declare const a: {b: {c: string} | null} | null; const arr = [a && a.b && a.b.c, a && a.b];`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; const arr = [a?.b?.c, a?.b];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Chain in object literal
		{
			Code:   `declare const a: {b: {c: string} | null} | null; const obj = {x: a && a.b && a.b.c, y: a && a.b};`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; const obj = {x: a?.b?.c, y: a?.b};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Chain as function argument
		{
			Code:   `declare function fn(x: any): void; declare const a: {b: {c: string} | null} | null; fn(a && a.b && a.b.c);`,
			Output: []string{`declare function fn(x: any): void; declare const a: {b: {c: string} | null} | null; fn(a?.b?.c);`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain as return value
		{
			Code:   `declare const a: {b: {c: string} | null} | null; function foo() { return a && a.b && a.b.c; }`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; function foo() { return a?.b?.c; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in arrow function
		{
			Code:   `declare const a: {b: {c: string} | null} | null; const fn = () => a && a.b && a.b.c;`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; const fn = () => a?.b?.c;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in IIFE
		{
			Code:   `declare const a: {b: {c: string} | null} | null; (function() { return a && a.b && a.b.c; })();`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; (function() { return a?.b?.c; })();`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain with multiple conditions in if statement
		{
			Code:   `declare const a: {b: {c: string} | null} | null; declare const x: boolean; if (x && a && a.b && a.b.c) {}`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; declare const x: boolean; if (x && a?.b?.c) {}`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in while loop condition
		{
			Code:   `declare const a: {b: {c: string} | null} | null; while (a && a.b && a.b.c) {}`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; while (a?.b?.c) {}`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in for loop condition
		{
			Code:   `declare const a: {b: {c: number} | null} | null; for (let i = 0; a && a.b && a.b.c > i; i++) {}`,
			Output: []string{`declare const a: {b: {c: number} | null} | null; for (let i = 0; a?.b?.c > i; i++) {}`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in do-while loop
		{
			Code:   `declare const a: {b: {c: string} | null} | null; do {} while (a && a.b && a.b.c);`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; do {} while (a?.b?.c);`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in switch expression
		{
			Code:   `declare const a: {b: {c: string} | null} | null; switch (a && a.b && a.b.c) { case 'test': break; }`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; switch (a?.b?.c) { case 'test': break; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in case expression
		{
			Code:   `declare const a: {b: {c: string} | null} | null; declare const x: any; switch (x) { case a && a.b && a.b.c: break; }`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; declare const x: any; switch (x) { case a?.b?.c: break; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in throw statement
		{
			Code:   `declare const a: {b: {c: Error} | null} | null; throw a && a.b && a.b.c;`,
			Output: []string{`declare const a: {b: {c: Error} | null} | null; throw a?.b?.c;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain in try-catch
		{
			Code:   `declare const a: {b: {c: string} | null} | null; try { const x = a && a.b && a.b.c; } catch {}`,
			Output: []string{`declare const a: {b: {c: string} | null} | null; try { const x = a?.b?.c; } catch {}`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chain with template literal
		{
			Code:   "declare const a: {b: {c: string} | null} | null; const str = `value: ${a && a.b && a.b.c}`;",
			Output: []string{"declare const a: {b: {c: string} | null} | null; const str = `value: ${a?.b?.c}`;"},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple chains in template literal
		{
			Code:   "declare const a: {b: string} | null; declare const c: {d: number} | null; const str = `${a && a.b} and ${c && c.d}`;",
			Output: []string{"declare const a: {b: string} | null; declare const c: {d: number} | null; const str = `${a?.b} and ${c?.d}`;"},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Chain with spread operator
		{
			Code:   `declare const a: {b: {c: string[]} | null} | null; const arr = [...(a && a.b && a.b.c || [])];`,
			Output: []string{`declare const a: {b: {c: string[]} | null} | null; const arr = [...(a?.b?.c || [])];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Very complex real-world pattern: Redux-style selector
		{
			Code:   `declare const state: {user: {profile: {settings: {theme: string} | null} | null} | null} | null; const theme = state && state.user && state.user.profile && state.user.profile.settings && state.user.profile.settings.theme;`,
			Output: []string{`declare const state: {user: {profile: {settings: {theme: string} | null} | null} | null} | null; const theme = state?.user?.profile?.settings?.theme;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// API response pattern with nested data
		{
			Code:   `declare const response: {data: {user: {name: string} | null} | null} | null; const name = response && response.data && response.data.user && response.data.user.name;`,
			Output: []string{`declare const response: {data: {user: {name: string} | null} | null} | null; const name = response?.data?.user?.name;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// GraphQL-style nested query result
		{
			Code:   `declare const result: {data: {viewer: {repositories: {edges: {node: {name: string} | null}[] | null} | null} | null} | null} | null; const edges = result && result.data && result.data.viewer && result.data.viewer.repositories && result.data.viewer.repositories.edges;`,
			Output: []string{`declare const result: {data: {viewer: {repositories: {edges: {node: {name: string} | null}[] | null} | null} | null} | null} | null; const edges = result?.data?.viewer?.repositories?.edges;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Configuration object with many levels
		{
			Code:   `declare const config: {server: {database: {connection: {host: string} | null} | null} | null} | null; const host = config != null && config.server != null && config.server.database != null && config.server.database.connection != null && config.server.database.connection.host;`,
			Output: []string{`declare const config: {server: {database: {connection: {host: string} | null} | null} | null} | null; const host = config?.server?.database?.connection?.host;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Event handling with nested properties
		{
			Code:   `declare const event: {target: {dataset: {userId: string} | null} | null} | null; const userId = event && event.target && event.target.dataset && event.target.dataset.userId;`,
			Output: []string{`declare const event: {target: {dataset: {userId: string} | null} | null} | null; const userId = event?.target?.dataset?.userId;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple chains with mathematical operations
		{
			Code:   `declare const a: {b: number} | null; declare const c: {d: number} | null; const sum = (a && a.b || 0) + (c && c.d || 0);`,
			Output: []string{`declare const a: {b: number} | null; declare const c: {d: number} | null; const sum = (a?.b || 0) + (c?.d || 0);`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Chains with string concatenation
		{
			Code:   `declare const a: {b: string} | null; declare const c: {d: string} | null; const str = (a && a.b || '') + ' ' + (c && c.d || '');`,
			Output: []string{`declare const a: {b: string} | null; declare const c: {d: string} | null; const str = (a?.b || '') + ' ' + (c?.d || '');`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// Chain with bitwise operations
		{
			Code:   `declare const a: {b: {c: number} | null} | null; const result = a && a.b && a.b.c & 0xFF;`,
			Output: []string{`declare const a: {b: {c: number} | null} | null; const result = a?.b?.c & 0xFF;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// 12-level ultra-deep chain (extreme stress test)
		{
			Code:   `declare const x: {a: {b: {c: {d: {e: {f: {g: {h: {i: {j: {k: {l: number} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null; x && x.a && x.a.b && x.a.b.c && x.a.b.c.d && x.a.b.c.d.e && x.a.b.c.d.e.f && x.a.b.c.d.e.f.g && x.a.b.c.d.e.f.g.h && x.a.b.c.d.e.f.g.h.i && x.a.b.c.d.e.f.g.h.i.j && x.a.b.c.d.e.f.g.h.i.j.k && x.a.b.c.d.e.f.g.h.i.j.k.l;`,
			Output: []string{`declare const x: {a: {b: {c: {d: {e: {f: {g: {h: {i: {j: {k: {l: number} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null} | null; x?.a?.b?.c?.d?.e?.f?.g?.h?.i?.j?.k?.l;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
