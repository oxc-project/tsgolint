package prefer_nullish_coalescing

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferNullishCoalescingRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferNullishCoalescingRule, []rule_tester.ValidTestCase{
		// ==========================================
		// Non-nullable types - typeValidTest pattern
		// types = ['string', 'number', 'boolean', 'object']
		// ==========================================
		// || operator
		{Code: `declare let x: string; (x || 'foo');`},
		{Code: `declare let x: number; (x || 'foo');`},
		{Code: `declare let x: boolean; (x || 'foo');`},
		{Code: `declare let x: object; (x || 'foo');`},
		// ||= operator
		{Code: `declare let x: string; (x ||= 'foo');`},
		{Code: `declare let x: number; (x ||= 'foo');`},
		{Code: `declare let x: boolean; (x ||= 'foo');`},
		{Code: `declare let x: object; (x ||= 'foo');`},

		// Additional non-nullable types
		{Code: `declare const x: symbol; x || Symbol('a');`},
		{Code: `declare const x: () => void; x || (() => {});`},
		{Code: `declare const x: string[]; x || [];`},

		// ==========================================
		// Already using nullish coalescing - nullishTypeTest pattern
		// nullishTypes = ['null', 'undefined', 'null | undefined']
		// types = ['string', 'number', 'boolean', 'object']
		// ==========================================
		// null variants
		{Code: `declare let x: string | null; x ?? 'foo';`},
		{Code: `declare let x: number | null; x ?? 'foo';`},
		{Code: `declare let x: boolean | null; x ?? 'foo';`},
		{Code: `declare let x: object | null; x ?? 'foo';`},
		{Code: `declare let x: string | null; x ??= 'foo';`},
		{Code: `declare let x: number | null; x ??= 'foo';`},
		{Code: `declare let x: boolean | null; x ??= 'foo';`},
		{Code: `declare let x: object | null; x ??= 'foo';`},
		// undefined variants
		{Code: `declare let x: string | undefined; x ?? 'foo';`},
		{Code: `declare let x: number | undefined; x ?? 'foo';`},
		{Code: `declare let x: boolean | undefined; x ?? 'foo';`},
		{Code: `declare let x: object | undefined; x ?? 'foo';`},
		{Code: `declare let x: string | undefined; x ??= 'foo';`},
		{Code: `declare let x: number | undefined; x ??= 'foo';`},
		{Code: `declare let x: boolean | undefined; x ??= 'foo';`},
		{Code: `declare let x: object | undefined; x ??= 'foo';`},
		// null | undefined variants
		{Code: `declare let x: string | null | undefined; x ?? 'foo';`},
		{Code: `declare let x: number | null | undefined; x ?? 'foo';`},
		{Code: `declare let x: boolean | null | undefined; x ?? 'foo';`},
		{Code: `declare let x: object | null | undefined; x ?? 'foo';`},
		{Code: `declare let x: string | null | undefined; x ??= 'foo';`},
		{Code: `declare let x: number | null | undefined; x ??= 'foo';`},
		{Code: `declare let x: boolean | null | undefined; x ??= 'foo';`},
		{Code: `declare let x: object | null | undefined; x ??= 'foo';`},

		// ==========================================
		// ignoreTernaryTests: true (default)
		// ==========================================
		{Code: `x !== undefined && x !== null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},

		// ==========================================
		// ignoreConditionalTests: true (default) - nullishTypeTest pattern
		// All type/nullish permutations for if, while, do-while, for, ternary condition
		// ==========================================
		// string | null
		{Code: `declare let x: string | null; (x || 'foo') ? null : null;`},
		{Code: `declare let x: string | null; (x ||= 'foo') ? null : null;`},
		{Code: `declare let x: string | null; if ((x || 'foo')) {}`},
		{Code: `declare let x: string | null; if ((x ||= 'foo')) {}`},
		{Code: `declare let x: string | null; do {} while ((x || 'foo'))`},
		{Code: `declare let x: string | null; do {} while ((x ||= 'foo'))`},
		{Code: `declare let x: string | null; for (;(x || 'foo');) {}`},
		{Code: `declare let x: string | null; for (;(x ||= 'foo');) {}`},
		{Code: `declare let x: string | null; while ((x || 'foo')) {}`},
		{Code: `declare let x: string | null; while ((x ||= 'foo')) {}`},
		// string | undefined
		{Code: `declare let x: string | undefined; (x || 'foo') ? null : null;`},
		{Code: `declare let x: string | undefined; (x ||= 'foo') ? null : null;`},
		{Code: `declare let x: string | undefined; if ((x || 'foo')) {}`},
		{Code: `declare let x: string | undefined; if ((x ||= 'foo')) {}`},
		{Code: `declare let x: string | undefined; do {} while ((x || 'foo'))`},
		{Code: `declare let x: string | undefined; do {} while ((x ||= 'foo'))`},
		{Code: `declare let x: string | undefined; for (;(x || 'foo');) {}`},
		{Code: `declare let x: string | undefined; for (;(x ||= 'foo');) {}`},
		{Code: `declare let x: string | undefined; while ((x || 'foo')) {}`},
		{Code: `declare let x: string | undefined; while ((x ||= 'foo')) {}`},
		// string | null | undefined
		{Code: `declare let x: string | null | undefined; (x || 'foo') ? null : null;`},
		{Code: `declare let x: string | null | undefined; (x ||= 'foo') ? null : null;`},
		{Code: `declare let x: string | null | undefined; if ((x || 'foo')) {}`},
		{Code: `declare let x: string | null | undefined; if ((x ||= 'foo')) {}`},
		{Code: `declare let x: string | null | undefined; do {} while ((x || 'foo'))`},
		{Code: `declare let x: string | null | undefined; do {} while ((x ||= 'foo'))`},
		{Code: `declare let x: string | null | undefined; for (;(x || 'foo');) {}`},
		{Code: `declare let x: string | null | undefined; for (;(x ||= 'foo');) {}`},
		{Code: `declare let x: string | null | undefined; while ((x || 'foo')) {}`},
		{Code: `declare let x: string | null | undefined; while ((x ||= 'foo')) {}`},
		// number | null
		{Code: `declare let x: number | null; (x || 'foo') ? null : null;`},
		{Code: `declare let x: number | null; if ((x || 'foo')) {}`},
		{Code: `declare let x: number | null; while ((x || 'foo')) {}`},
		// number | undefined
		{Code: `declare let x: number | undefined; (x || 'foo') ? null : null;`},
		{Code: `declare let x: number | undefined; if ((x || 'foo')) {}`},
		{Code: `declare let x: number | undefined; while ((x || 'foo')) {}`},
		// boolean | null
		{Code: `declare let x: boolean | null; (x || 'foo') ? null : null;`},
		{Code: `declare let x: boolean | null; if ((x || 'foo')) {}`},
		// boolean | undefined
		{Code: `declare let x: boolean | undefined; (x || 'foo') ? null : null;`},
		{Code: `declare let x: boolean | undefined; if ((x || 'foo')) {}`},
		// object | null
		{Code: `declare let x: object | null; (x || 'foo') ? null : null;`},
		{Code: `declare let x: object | null; if ((x || 'foo')) {}`},
		// object | undefined
		{Code: `declare let x: object | undefined; (x || 'foo') ? null : null;`},
		{Code: `declare let x: object | undefined; if ((x || 'foo')) {}`},

		// ==========================================
		// ignoreTernaryTests: true (must be explicitly set, default is false)
		// ==========================================
		{Code: `declare const x: string | undefined; x ? x : 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},
		{Code: `declare const x: string | undefined; !x ? 'a' : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},
		{Code: `declare const x: string | undefined; x !== undefined ? x : 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},
		{Code: `declare const x: string | null; x !== null ? x : 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},
		{Code: `declare const x: string | null | undefined; x != null ? x : 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},
		{Code: `declare const x: string | null | undefined; x == null ? 'a' : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": true}`)},

		// ==========================================
		// ignoreMixedLogicalExpressions: true - nullishTypeTest pattern
		// ==========================================
		// string | null
		{Code: `declare let a: string | null; declare let b: string | null; declare let c: string | null; a || b && c;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare let a: string | null; declare let b: string | null; declare let c: string | null; declare let d: string | null; a || b || c && d;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare let a: string | null; declare let b: string | null; declare let c: string | null; declare let d: string | null; a && b || c || d;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		// string | undefined
		{Code: `declare let a: string | undefined; declare let b: string | undefined; declare let c: string | undefined; a || b && c;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare let a: string | undefined; declare let b: string | undefined; declare let c: string | undefined; declare let d: string | undefined; a || b || c && d;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare let a: string | undefined; declare let b: string | undefined; declare let c: string | undefined; declare let d: string | undefined; a && b || c || d;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		// number | null
		{Code: `declare let a: number | null; declare let b: number | null; declare let c: number | null; a || b && c;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		// object | undefined
		{Code: `declare let a: object | undefined; declare let b: object | undefined; declare let c: object | undefined; a || b && c;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		// ==========================================
		// ignorePrimitives options - all ignorable primitives
		// ==========================================
		{Code: `declare let x: string | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: number | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: boolean | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: bigint | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		// ignorePrimitives: true (boolean form - ignores all primitives)
		{Code: `declare let x: string | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: number | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: boolean | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: bigint | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		// Branded types with ignorePrimitives: true (boolean form)
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		// Ternary with ignorePrimitives: true (boolean form)
		{Code: `declare let x: string | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: string | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: number | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: number | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: boolean | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: boolean | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: bigint | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: bigint | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		// Branded ternary with ignorePrimitives: true (boolean form)
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (string & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": true}`)},
		// ignorePrimitives: all set to true (object form - equivalent to ignorePrimitives: true)
		{Code: `declare let x: string | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: number | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: boolean | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: bigint | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		// Branded types with ignorePrimitives
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		// Branded types with ignorePrimitives (all set to true)
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		// Ternary with ignorePrimitives
		{Code: `declare let x: string | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: string | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: number | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: number | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: boolean | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: boolean | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: bigint | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		{Code: `declare let x: bigint | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		// Ternary with ignorePrimitives (all set to true)
		{Code: `declare let x: string | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: string | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: number | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: number | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: boolean | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: boolean | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: bigint | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: bigint | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		// Branded ternary with ignorePrimitives
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: (string & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		// Branded ternary with ignorePrimitives (all set to true)
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (string & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		// Multiple primitives
		{Code: `declare const x: string | number | undefined; x || 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true}}`)},
		// Literal types
		{Code: `declare const x: 'a' | 'b' | undefined; x || 'a';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare const x: 1 | 2 | undefined; x || 1;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},

		// ==========================================
		// ignoreMixedLogicalExpressions: true
		// ==========================================
		{Code: `declare const a: string | undefined; declare const b: string | undefined; a || (b && 'c');`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare const a: string | undefined; declare const b: string | undefined; (a && 'b') || c;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},
		{Code: `declare const a: string | undefined; declare const b: string | undefined; a || b && 'c';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": true}`)},

		// ==========================================
		// ignoreBooleanCoercion: true
		// ==========================================
		{Code: `declare const x: string | undefined; Boolean(x || 'a');`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": true}`)},
		{Code: `declare const x: string | undefined; Boolean(x || y || z);`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": true}`)},

		// ==========================================
		// ignoreIfStatements: true
		// ==========================================
		{Code: `declare let x: string | undefined; if (!x) { x = 'default'; }`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreIfStatements": true}`)},
		{Code: `declare let x: string | undefined; if (x === undefined) { x = 'default'; }`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreIfStatements": true}`)},
		{Code: `declare let x: string | null; if (x === null) { x = 'default'; }`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreIfStatements": true}`)},
		{Code: `declare let x: string | null | undefined; if (x == null) { x = 'default'; }`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreIfStatements": true}`)},

		// If statements that should not be flagged anyway (has else, multiple statements, etc.)
		{Code: `declare let x: string | undefined; if (!x) { x = 'a'; } else { x = 'b'; }`},
		{Code: `declare let x: string | undefined; if (!x) { x = 'a'; console.log(x); }`},
		{Code: `declare let x: string | undefined; let y: string; if (!x) { y = 'a'; }`}, // different variable

		// ==========================================
		// any/unknown types with ignorePrimitives - should not flag
		// ==========================================
		{Code: `declare const a: any; declare const b: any; a ? a : b;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare const a: any; declare const b: any; a ? a : b;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare const a: unknown; const b = a || 'bar';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": false, "number": false, "string": false}}`)},

		// ==========================================
		// Ternary edge cases that should NOT be flagged
		// ==========================================
		// Different identifiers
		{Code: `declare const x: string | undefined; declare const y: string; x ? y : 'a';`},
		// Non-member-access condition
		{Code: `declare const x: string | undefined; x.length ? x : 'a';`},
		// Complex conditions that don't match pattern
		{Code: `declare const x: string | undefined; x === 'foo' ? x : 'a';`},

		// ==========================================
		// Optional chaining - should not affect nullish detection
		// ==========================================
		{Code: `declare const x: {a?: string}; x?.a ?? 'a';`},

		// ==========================================
		// Enum types - should be flagged but we test valid cases
		// ==========================================
		{Code: `enum E { A, B } declare const x: E; x || E.A;`}, // non-nullable enum

		// ==========================================
		// Ternary tests that should NOT be flagged (from upstream)
		// ==========================================
		// Different branches
		{Code: `x !== undefined && x !== null ? "foo" : "bar";`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Extra conditions
		{Code: `x !== null && x !== undefined && x !== 5 ? x : y`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x === null || x === undefined || x === 5 ? x : y`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Mixed operators
		{Code: `x === undefined && x !== null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x === undefined && x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x !== undefined && x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x === undefined || x !== null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x === undefined || x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x !== undefined || x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x !== undefined || x === null ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Same null checks
		{Code: `x === null || x === null ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x === undefined || x === undefined ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Wrong operand order
		{Code: `x == null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x == undefined ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x != null ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `x != undefined ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Null literal comparisons
		{Code: `undefined == null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `undefined != z ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Non-nullish value comparisons
		{Code: `declare let x: number | undefined; x !== 15 && x !== undefined ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: number | undefined; x !== undefined && x !== 15 ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string | undefined; x !== 'foo' && x !== undefined ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Type mismatch for strict checks
		{Code: `declare let x: string; x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string | undefined; x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string | null; x === undefined ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Partial nullish check on full nullish type
		{Code: `declare let x: string | undefined | null; x !== undefined ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string | undefined | null; x !== null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// any/unknown type checks
		{Code: `declare let x: any; x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: unknown; x === null ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// ==========================================
		// Non-nullable types for ternary - should not flag
		// ==========================================
		{Code: `declare let x: string; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: number; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: number; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: bigint; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: bigint; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: boolean; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: boolean; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: object; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: object; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string[]; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: string[]; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// Function returning nullable - function itself is not nullable
		{Code: `declare let x: () => string | null; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string | null; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		// Function call results - separate invocations
		{Code: `declare let x: () => string | null; x() ? x() : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: () => string | null; !x() ? y : x();`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// Non-nullable member access
		{Code: `declare let x: { n: string }; x.n ? x.n : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: string }; !x.n ? y : x.n;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: number }; x.n ? x.n : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: number }; !x.n ? y : x.n;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: boolean }; x.n ? x.n : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: boolean }; !x.n ? y : x.n;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: bigint }; x.n ? x.n : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { n: bigint }; !x.n ? y : x.n;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// ==========================================
		// If statement edge cases that should NOT be flagged
		// ==========================================
		// Non-nullable type
		{Code: `declare let foo: string; declare function makeFoo(): string; function lazyInitialize() { if (!foo) { foo = makeFoo(); } }`},
		// Assignment in if with else
		{Code: `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo) { foo = makeFoo(); } }`},
		// Assignment with extra statement
		{Code: `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo == null) { foo = makeFoo(); return foo; } }`},
		// With else branch
		{Code: `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo == null) { foo = makeFoo(); } else { return 'bar'; } }`},
		// Shadowed variable
		{Code: `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function shadowed() { if (foo == null) { const foo = makeFoo(); } }`},

		// ==========================================
		// never type - should not be flagged
		// ==========================================
		{Code: `declare let x: never; declare let y: number; x || y;`},
		{Code: `declare let x: never; declare let y: number; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: never; declare let y: number; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// ==========================================
		// Optional chaining edge cases - different member access
		// ==========================================
		{Code: `const a = 'b'; declare let x: { a: string, b: string } | null; x?.a != null ? x[a] : 'foo'`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `const a = 'b'; declare let x: { a: string, b: string } | null; x?.[a] != null ? x.a : 'foo'`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},
		{Code: `declare let x: { a: string } | null; declare let y: { a: string } | null; x?.a ? y?.a : 'foo'`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`)},

		// ==========================================
		// ignorePrimitives: true (all primitives)
		// ==========================================
		{Code: `declare let x: string | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: number | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: boolean | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		{Code: `declare let x: bigint | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},
		// Branded types with ignorePrimitives
		{Code: `declare let x: (string & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: (number & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: (boolean & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: (bigint & { __brand?: any }) | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},
		// Ternary with ignorePrimitives
		{Code: `declare let x: string | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: string | undefined; !x ? y : x;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `declare let x: number | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `declare let x: boolean | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"boolean": true}}`)},
		{Code: `declare let x: bigint | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true}}`)},

		// ==========================================
		// Enum with ignorePrimitives
		// ==========================================
		{Code: `enum Enum { A = 0, B = 1, C = 2 } declare let x: Enum | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `enum Enum { A = 'a', B = 'b', C = 'c' } declare let x: Enum | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},
		{Code: `enum Enum { A = 0, B = 1, C = 2 } declare let x: Enum | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true}}`)},
		{Code: `enum Enum { A = 'a', B = 'b', C = 'c' } declare let x: Enum | undefined; x ? x : y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true}}`)},

		// ==========================================
		// Mixed unions with ignorePrimitives
		// ==========================================
		{Code: `declare let x: 0 | 1 | 0n | 1n | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`)},
		{Code: `declare let x: 0 | 1 | 0n | 1n | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`)},
		{Code: `declare let x: 0 | 'foo' | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true, "string": true}}`)},
		{Code: `declare let x: 0 | 'foo' | undefined; x || y;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"number": true, "string": false}}`)},

		// ==========================================
		// Boolean coercion with ignoreBooleanCoercion
		// ==========================================
		{Code: `let a: string | true | undefined; let b: string | boolean | undefined; const x = Boolean(a || b);`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": true}`)},
		{Code: `let a: string | boolean | undefined; let b: string | boolean | undefined; let c: string | boolean | undefined; const test = Boolean(a || b || c);`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": true}`)},
		{Code: `let a: string | boolean | undefined; let b: string | boolean | undefined; let c: string | boolean | undefined; const test = Boolean(a || (b && c));`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": true}`)},

		// ==========================================
		// any type with ignorePrimitives (all set to true)
		// ==========================================
		{Code: `declare const a: any; declare const b: any; a ? a : b;`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"string": true, "number": true, "boolean": true, "bigint": true}}`)},

		// ==========================================
		// unknown type
		// ==========================================
		{Code: `declare const a: unknown; const b = a || 'bar';`, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": false, "number": false, "string": false}}`)},
	}, []rule_tester.InvalidTestCase{
		// ==========================================
		// Basic || operator - string | undefined
		// ==========================================
		{
			Code:   `declare const x: string | undefined; x || 'a';`,
			Output: []string{`declare const x: string | undefined; x ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// string | null
		{
			Code:   `declare const x: string | null; x || 'a';`,
			Output: []string{`declare const x: string | null; x ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// string | null | undefined
		{
			Code:   `declare const x: string | null | undefined; x || 'a';`,
			Output: []string{`declare const x: string | null | undefined; x ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// number | undefined
		{
			Code:   `declare const x: number | undefined; x || 1;`,
			Output: []string{`declare const x: number | undefined; x ?? 1;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// number | null
		{
			Code:   `declare const x: number | null; x || 1;`,
			Output: []string{`declare const x: number | null; x ?? 1;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// boolean | undefined
		{
			Code:   `declare const x: boolean | undefined; x || true;`,
			Output: []string{`declare const x: boolean | undefined; x ?? true;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// bigint | undefined
		{
			Code:   `declare const x: bigint | undefined; x || 1n;`,
			Output: []string{`declare const x: bigint | undefined; x ?? 1n;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		// object | undefined
		{
			Code:   `declare const x: object | undefined; x || {};`,
			Output: []string{`declare const x: object | undefined; x ?? {};`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// ignorePrimitives: false (should still report errors)
		// ==========================================
		{
			Code:    `declare const x: string | undefined; x || 'a';`,
			Output:  []string{`declare const x: string | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: number | undefined; x || 1;`,
			Output:  []string{`declare const x: number | undefined; x ?? 1;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: boolean | undefined; x || true;`,
			Output:  []string{`declare const x: boolean | undefined; x ?? true;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: bigint | undefined; x || 1n;`,
			Output:  []string{`declare const x: bigint | undefined; x ?? 1n;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// ||= operator cases
		// ==========================================
		{
			Code:   `declare let x: string | undefined; x ||= 'a';`,
			Output: []string{`declare let x: string | undefined; x ??= 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare let x: number | null; x ||= 1;`,
			Output: []string{`declare let x: number | null; x ??= 1;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare let x: boolean | undefined; x ||= false;`,
			Output: []string{`declare let x: boolean | undefined; x ??= false;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Member access expressions
		// ==========================================
		{
			Code:   `declare const x: { n: string | undefined }; x.n || 'a';`,
			Output: []string{`declare const x: { n: string | undefined }; x.n ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare const x: { n?: string }; x.n || 'a';`,
			Output: []string{`declare const x: { n?: string }; x.n ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare const x: { a: { b: string | undefined } }; x.a.b || 'a';`,
			Output: []string{`declare const x: { a: { b: string | undefined } }; x.a.b ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Ternary expressions - ignoreTernaryTests: false
		// ==========================================
		// Simple truthiness check: x ? x : 'a'
		{
			Code:    `declare const x: string | undefined; x ? x : 'a';`,
			Output:  []string{`declare const x: string | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Negated truthiness check: !x ? 'a' : x
		{
			Code:    `declare const x: string | undefined; !x ? 'a' : x;`,
			Output:  []string{`declare const x: string | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Strict equality with null: x !== null ? x : 'a'
		{
			Code:    `declare const x: string | null; x !== null ? x : 'a';`,
			Output:  []string{`declare const x: string | null; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Strict equality with undefined: x !== undefined ? x : 'a'
		{
			Code:    `declare const x: string | undefined; x !== undefined ? x : 'a';`,
			Output:  []string{`declare const x: string | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Loose equality with null: x != null ? x : 'a'
		{
			Code:    `declare const x: string | null | undefined; x != null ? x : 'a';`,
			Output:  []string{`declare const x: string | null | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Loose equality with null (reversed): x == null ? 'a' : x
		{
			Code:    `declare const x: string | null | undefined; x == null ? 'a' : x;`,
			Output:  []string{`declare const x: string | null | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Compound null/undefined check: x === null || x === undefined ? 'a' : x
		{
			Code:    `declare const x: string | null | undefined; x === null || x === undefined ? 'a' : x;`,
			Output:  []string{`declare const x: string | null | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Compound null/undefined check reversed: x !== null && x !== undefined ? x : 'a'
		{
			Code:    `declare const x: string | null | undefined; x !== null && x !== undefined ? x : 'a';`,
			Output:  []string{`declare const x: string | null | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// With parentheses: (x) ? (x) : 'a'
		{
			Code:    `declare const x: string | undefined; (x) ? (x) : 'a';`,
			Output:  []string{`declare const x: string | undefined; x ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Member access in ternary
		{
			Code:    `declare const x: { n: string | undefined }; x.n ? x.n : 'a';`,
			Output:  []string{`declare const x: { n: string | undefined }; x.n ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// If statements - ignoreIfStatements: false (default)
		// ==========================================
		// Simple negation check
		{
			Code:   `declare let x: string | undefined; if (!x) { x = 'default'; }`,
			Output: []string{`declare let x: string | undefined; x ??= 'default';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		// Equality check with undefined
		{
			Code:   `declare let x: string | undefined; if (x === undefined) { x = 'default'; }`,
			Output: []string{`declare let x: string | undefined; x ??= 'default';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		// Equality check with null
		{
			Code:   `declare let x: string | null; if (x === null) { x = 'default'; }`,
			Output: []string{`declare let x: string | null; x ??= 'default';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		// Loose equality with null
		{
			Code:   `declare let x: string | null | undefined; if (x == null) { x = 'default'; }`,
			Output: []string{`declare let x: string | null | undefined; x ??= 'default';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		// Block with single statement
		{
			Code:   `declare let x: string | undefined; if (!x) { x = 'default'; }`,
			Output: []string{`declare let x: string | undefined; x ??= 'default';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},

		// ==========================================
		// ignoreConditionalTests: false
		// ==========================================
		{
			Code:    `declare const x: string | undefined; if (x || 'a') {}`,
			Output:  []string{`declare const x: string | undefined; if (x ?? 'a') {}`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: string | undefined; while (x || 'a') {}`,
			Output:  []string{`declare const x: string | undefined; while (x ?? 'a') {}`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: string | undefined; do {} while (x || 'a');`,
			Output:  []string{`declare const x: string | undefined; do {} while (x ?? 'a');`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: string | undefined; for (; x || 'a'; ) {}`,
			Output:  []string{`declare const x: string | undefined; for (; x ?? 'a'; ) {}`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: string | undefined; (x || 'a') ? 1 : 2;`,
			Output:  []string{`declare const x: string | undefined; (x ?? 'a') ? 1 : 2;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// ignoreBooleanCoercion: false (default)
		// ==========================================
		{
			Code:    `declare const x: string | undefined; Boolean(x || 'a');`,
			Output:  []string{`declare const x: string | undefined; Boolean(x ?? 'a');`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreBooleanCoercion": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Mixed logical expressions - ignoreMixedLogicalExpressions: false (default)
		// ==========================================
		{
			Code:    `declare const a: string | undefined; declare const b: string | undefined; a || (b && 'c');`,
			Output:  []string{`declare const a: string | undefined; declare const b: string | undefined; a ?? (b && 'c');`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreMixedLogicalExpressions": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Enum types
		// ==========================================
		{
			Code:   `enum E { A, B } declare const x: E | undefined; x || E.A;`,
			Output: []string{`enum E { A, B } declare const x: E | undefined; x ?? E.A;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Literal types
		// ==========================================
		{
			Code:   `declare const x: 'a' | 'b' | undefined; x || 'a';`,
			Output: []string{`declare const x: 'a' | 'b' | undefined; x ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare const x: 1 | 2 | undefined; x || 1;`,
			Output: []string{`declare const x: 1 | 2 | undefined; x ?? 1;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Function return types
		// ==========================================
		{
			Code:   `function f(): string | undefined { return undefined; } f() || 'a';`,
			Output: []string{`function f(): string | undefined { return undefined; } f() ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Array element access
		// ==========================================
		{
			Code:   `declare const arr: (string | undefined)[]; arr[0] || 'a';`,
			Output: []string{`declare const arr: (string | undefined)[]; arr[0] ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// strictNullChecks disabled
		// ==========================================
		{
			Code:     `function foo(): boolean { return true; }`,
			TSConfig: "tsconfig.unstrict.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck"}},
		},

		// ==========================================
		// Ternary with various null/undefined check patterns (from upstream)
		// ==========================================
		{
			Code:    `x !== undefined && x !== null ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x !== null && x !== undefined ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x === undefined || x === null ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x === null || x === undefined ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `undefined !== x && x !== null ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `null !== x && x !== undefined ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `undefined === x || x === null ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `null === x || x === undefined ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Loose equality patterns
		{
			Code:    `x != undefined && x != null ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x == undefined || x == null ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `undefined != x ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `null != x ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `undefined == x ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `null == x ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x != undefined ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x != null ? x : y;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x == undefined ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x == null ? y : x;`,
			Output:  []string{`x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// this keyword
		{
			Code:    `this != undefined ? this : y;`,
			Output:  []string{`this ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Typed ternary tests - various types with truthiness check
		// ==========================================
		{
			Code:    `declare let x: string | null; x ? x : y;`,
			Output:  []string{`declare let x: string | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string | null; !x ? y : x;`,
			Output:  []string{`declare let x: string | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string | null | undefined; x ? x : y;`,
			Output:  []string{`declare let x: string | null | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string | null | undefined; !x ? y : x;`,
			Output:  []string{`declare let x: string | null | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: number | null; x ? x : y;`,
			Output:  []string{`declare let x: number | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: number | undefined; x ? x : y;`,
			Output:  []string{`declare let x: number | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: bigint | null; x ? x : y;`,
			Output:  []string{`declare let x: bigint | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: bigint | undefined; x ? x : y;`,
			Output:  []string{`declare let x: bigint | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: boolean | null; x ? x : y;`,
			Output:  []string{`declare let x: boolean | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: boolean | undefined; x ? x : y;`,
			Output:  []string{`declare let x: boolean | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: object | null; x ? x : y;`,
			Output:  []string{`declare let x: object | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: object | undefined; x ? x : y;`,
			Output:  []string{`declare let x: object | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string[] | null; x ? x : y;`,
			Output:  []string{`declare let x: string[] | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string[] | undefined; x ? x : y;`,
			Output:  []string{`declare let x: string[] | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: (() => string) | null; x ? x : y;`,
			Output:  []string{`declare let x: (() => string) | null; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: (() => string) | undefined; x ? x : y;`,
			Output:  []string{`declare let x: (() => string) | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Member access ternary tests
		// ==========================================
		{
			Code:    `declare let x: { n: string | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: number | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: number | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: boolean | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: boolean | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: bigint | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: bigint | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Optional chaining ternary tests
		// ==========================================
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a !== undefined ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Parenthesized optional chain
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// If statement tests
		// ==========================================
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (!foo) { foo = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo == null) { foo = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo === null) { foo = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		{
			Code:   `declare let foo: { a: string } | undefined; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo === undefined) { foo = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | undefined; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		{
			Code:   `declare let foo: { a: string } | null | undefined; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo === undefined || foo === null) { foo = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null | undefined; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		// Member access if statement
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): string; function lazyInitialize() { if (foo.a == null) { foo.a = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): string; function lazyInitialize() { foo.a ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): string; function lazyInitialize() { if (foo?.a == null) { foo.a = makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): string; function lazyInitialize() { foo.a ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},

		// ==========================================
		// Enum tests
		// ==========================================
		{
			Code:   `enum Enum { A = 0, B = 1, C = 2 } declare let x: Enum | undefined; x || y;`,
			Output: []string{`enum Enum { A = 0, B = 1, C = 2 } declare let x: Enum | undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `enum Enum { A = 'a', B = 'b', C = 'c' } declare let x: Enum | undefined; x || y;`,
			Output: []string{`enum Enum { A = 'a', B = 'b', C = 'c' } declare let x: Enum | undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Pure null/undefined types
		// ==========================================
		{
			Code:   `declare let x: null; x || y;`,
			Output: []string{`declare let x: null; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `const x = undefined; x || y;`,
			Output: []string{`const x = undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `null || y;`,
			Output: []string{`null ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `undefined || y;`,
			Output: []string{`undefined ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Element access ternary tests
		// ==========================================
		{
			Code:    `declare let x: { a: string } | null; x?.['a'] != null ? x['a'] : 'foo';`,
			Output:  []string{`declare let x: { a: string } | null; x?.['a'] ?? 'foo';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `const a = 'b'; declare let x: { a: string; b: string } | null; x?.[a] != null ? x[a] : 'foo';`,
			Output:  []string{`const a = 'b'; declare let x: { a: string; b: string } | null; x?.[a] ?? 'foo';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Chained || expressions
		// ==========================================
		{
			Code:   `declare let a: string | null; declare let b: string; declare let c: string; a || b || c;`,
			Output: []string{`declare let a: string | null; declare let b: string; declare let c: string; (a ?? b) || c;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Nested ternary fix with parentheses
		// ==========================================
		{
			Code:    `let a: string | undefined; let b: { message: string } | undefined; const foo = a ? a : b ? 1 : 2;`,
			Output:  []string{`let a: string | undefined; let b: { message: string } | undefined; const foo = a ?? (b ? 1 : 2);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare const c: string | null; c !== null ? c : c ? 1 : 2;`,
			Output:  []string{`declare const c: string | null; c ?? (c ? 1 : 2);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Member access ternary tests with various types
		// ==========================================
		// string | null
		{
			Code:    `declare let x: { n: string | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string | undefined
		{
			Code:    `declare let x: { n: string | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string | null | undefined
		{
			Code:    `declare let x: { n: string | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string | object | null
		{
			Code:    `declare let x: { n: string | object | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | object | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | object | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | object | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string | object | undefined
		{
			Code:    `declare let x: { n: string | object | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | object | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | object | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | object | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string | object | null | undefined
		{
			Code:    `declare let x: { n: string | object | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string | object | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string | object | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string | object | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// number | null
		{
			Code:    `declare let x: { n: number | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: number | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: number | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: number | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// number | undefined
		{
			Code:    `declare let x: { n: number | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: number | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: number | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: number | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// number | null | undefined
		{
			Code:    `declare let x: { n: number | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: number | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: number | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: number | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// bigint | null
		{
			Code:    `declare let x: { n: bigint | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: bigint | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: bigint | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: bigint | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// bigint | undefined
		{
			Code:    `declare let x: { n: bigint | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: bigint | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: bigint | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: bigint | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// bigint | null | undefined
		{
			Code:    `declare let x: { n: bigint | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: bigint | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: bigint | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: bigint | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// boolean | null
		{
			Code:    `declare let x: { n: boolean | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: boolean | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: boolean | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: boolean | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// boolean | undefined
		{
			Code:    `declare let x: { n: boolean | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: boolean | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: boolean | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: boolean | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// boolean | null | undefined
		{
			Code:    `declare let x: { n: boolean | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: boolean | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: boolean | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: boolean | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string[] | null
		{
			Code:    `declare let x: { n: string[] | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string[] | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string[] | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string[] | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string[] | undefined
		{
			Code:    `declare let x: { n: string[] | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string[] | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string[] | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string[] | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// string[] | null | undefined
		{
			Code:    `declare let x: { n: string[] | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: string[] | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: string[] | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: string[] | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// object | null
		{
			Code:    `declare let x: { n: object | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: object | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: object | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: object | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// object | undefined
		{
			Code:    `declare let x: { n: object | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: object | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: object | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: object | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// object | null | undefined
		{
			Code:    `declare let x: { n: object | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: object | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: object | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: object | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Function | null
		{
			Code:    `declare let x: { n: Function | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: Function | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: Function | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: Function | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Function | undefined
		{
			Code:    `declare let x: { n: Function | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: Function | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: Function | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: Function | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// Function | null | undefined
		{
			Code:    `declare let x: { n: Function | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: Function | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: Function | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: Function | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// (() => string) | null
		{
			Code:    `declare let x: { n: (() => string) | null }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: (() => string) | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: (() => string) | null }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: (() => string) | null }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// (() => string) | undefined
		{
			Code:    `declare let x: { n: (() => string) | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: (() => string) | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: (() => string) | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: (() => string) | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		// (() => string) | null | undefined
		{
			Code:    `declare let x: { n: (() => string) | null | undefined }; x.n ? x.n : y;`,
			Output:  []string{`declare let x: { n: (() => string) | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n: (() => string) | null | undefined }; !x.n ? y : x.n;`,
			Output:  []string{`declare let x: { n: (() => string) | null | undefined }; x.n ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Optional chaining ternary: x.n?.a patterns
		// ==========================================
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a !== undefined ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a !== undefined ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a !== undefined ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != undefined ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != undefined ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != undefined ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != null ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != null ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x.n?.a != null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x.n?.a !== undefined && x.n.a !== null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Optional chaining ternary: x?.n?.a patterns
		// ==========================================
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a !== undefined ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a !== undefined ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a !== undefined ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a !== undefined ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != undefined ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != undefined ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != undefined ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != undefined ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != null ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != null ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != null ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string } }; x?.n?.a != null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? (x?.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? (x.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? (x?.n).a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; x?.n?.a !== undefined && x.n.a !== null ? (x.n).a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; x?.n?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Parenthesized optional chain: (x?.n)?.a patterns
		// ==========================================
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? (x?.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? (x.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x?.n)?.a ? (x?.n).a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x?.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Parenthesized optional chain: (x.n)?.a patterns
		// ==========================================
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? x?.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? x.n?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? x?.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? x.n.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? (x?.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? (x.n)?.a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: { n?: { a?: string | null } }; (x.n)?.a ? (x?.n).a : y;`,
			Output:  []string{`declare let x: { n?: { a?: string | null } }; (x.n)?.a ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Functions inside conditional tests - should still report
		// ==========================================
		{
			Code:   `declare let x: string | undefined; if (() => (x || 'foo')) {}`,
			Output: []string{`declare let x: string | undefined; if (() => (x ?? 'foo')) {}`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare let x: string | undefined; if (function weird() { return (x || 'foo') }) {}`,
			Output: []string{`declare let x: string | undefined; if (function weird() { return (x ?? 'foo') }) {}`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Template literal types
		// ==========================================
		{
			Code:    `declare let x: '' | undefined; x || y;`,
			Output:  []string{`declare let x: '' | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare let x: 'a' | undefined; x || y;`,
			Output:  []string{`declare let x: 'a' | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Falsy literal types - number
		// ==========================================
		{
			Code:    `declare let x: 0 | undefined; x || y;`,
			Output:  []string{`declare let x: 0 | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare let x: 1 | undefined; x || y;`,
			Output:  []string{`declare let x: 1 | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Falsy literal types - bigint
		// ==========================================
		{
			Code:    `declare let x: 0n | undefined; x || y;`,
			Output:  []string{`declare let x: 0n | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare let x: 1n | undefined; x || y;`,
			Output:  []string{`declare let x: 1n | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": false, "boolean": true, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Falsy literal types - boolean
		// ==========================================
		{
			Code:    `declare let x: false | undefined; x || y;`,
			Output:  []string{`declare let x: false | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare let x: true | undefined; x || y;`,
			Output:  []string{`declare let x: true | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Unions of same primitive - ternary
		// ==========================================
		{
			Code:    `declare let x: 'a' | 'b' | undefined; x ? x : y;`,
			Output:  []string{`declare let x: 'a' | 'b' | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: 'a' | 'b' | undefined; !x ? y : x;`,
			Output:  []string{`declare let x: 'a' | 'b' | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": true, "number": true, "string": false}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: 0 | 1 | undefined; x ? x : y;`,
			Output:  []string{`declare let x: 0 | 1 | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: 0 | 1 | undefined; !x ? y : x;`,
			Output:  []string{`declare let x: 0 | 1 | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": true, "number": false, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: true | false | undefined; x ? x : y;`,
			Output:  []string{`declare let x: true | false | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: true | false | undefined; !x ? y : x;`,
			Output:  []string{`declare let x: true | false | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false, "ignorePrimitives": {"bigint": true, "boolean": false, "number": true, "string": true}}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Box/interface type tests - object with optional nested properties
		// ==========================================
		{
			Code:   `interface Box { value: string; } declare function getFallbackBox(): Box; declare const defaultBox: Box | undefined; defaultBox || getFallbackBox();`,
			Output: []string{`interface Box { value: string; } declare function getFallbackBox(): Box; declare const defaultBox: Box | undefined; defaultBox ?? getFallbackBox();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `interface Box { value: string; } declare function getFallbackBox(): Box; declare const defaultBox: Box | undefined; defaultBox ? defaultBox : getFallbackBox();`,
			Output:  []string{`interface Box { value: string; } declare function getFallbackBox(): Box; declare const defaultBox: Box | undefined; defaultBox ?? getFallbackBox();`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// If statement with nullish-coalescing assignment already inside
		// ==========================================
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo == null) { foo ??= makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},

		// ==========================================
		// If statement without block (inline)
		// ==========================================
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { if (foo == null) foo = makeFoo(); const bar = 42; return bar; }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function lazyInitialize() { foo ??= makeFoo(); const bar = 42; return bar; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},

		// ==========================================
		// Complex member access with this keyword
		// ==========================================
		{
			Code:    `class T { a?: string; b() { return this.a ? this.a : 'foo'; } }`,
			Output:  []string{`class T { a?: string; b() { return this.a ?? 'foo'; } }`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `class T { a?: string; b() { return !this.a ? 'foo' : this.a; } }`,
			Output:  []string{`class T { a?: string; b() { return this.a ?? 'foo'; } }`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `class T { a?: string; b() { return this.a !== undefined ? this.a : 'foo'; } }`,
			Output:  []string{`class T { a?: string; b() { return this.a ?? 'foo'; } }`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:   `class T { a?: string; b() { return this.a || 'foo'; } }`,
			Output: []string{`class T { a?: string; b() { return this.a ?? 'foo'; } }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Complex computed property access
		// ==========================================
		{
			Code:   `declare let x: { a: { b: string | undefined } }; declare const key: 'b'; x.a[key] || 'foo';`,
			Output: []string{`declare let x: { a: { b: string | undefined } }; declare const key: 'b'; x.a[key] ?? 'foo';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare let x: { a: { b: string | undefined } }; declare const key: 'b'; x.a[key] ? x.a[key] : 'foo';`,
			Output:  []string{`declare let x: { a: { b: string | undefined } }; declare const key: 'b'; x.a[key] ?? 'foo';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Assignment expression as fallback in ternary
		// ==========================================
		{
			Code:    `declare let x: string | null; declare let z: string; declare let y: string; x !== null ? x : (z = y);`,
			Output:  []string{`declare let x: string | null; declare let z: string; declare let y: string; x ?? (z = y);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare let x: string | null; declare let z: string; declare let y: string; x === null ? (z = y) : x;`,
			Output:  []string{`declare let x: string | null; declare let z: string; declare let y: string; x ?? (z = y);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// NOTE: Weird parentheses patterns in if-statements are not yet supported
		// The implementation doesn't handle deeply nested parentheses like:
		// if (((((foo.a)) == null))) { ((((((((foo).a))))) = makeString())); }

		// ==========================================
		// Double report scenario - if statement with ||= inside
		// ==========================================
		{
			Code:   `declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function f() { if (foo == null) { foo ||= makeFoo(); } }`,
			Output: []string{`declare let foo: { a: string } | null; declare function makeFoo(): { a: string }; function f() { foo ??= makeFoo(); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferNullishOverAssignment"},
				{MessageId: "preferNullishOverOr"},
			},
		},

		// NOTE: Deeply nested member access with index signatures are not fully supported
		// The implementation may not correctly resolve types through complex computed property chains
		// like: x.z[1][o]['3'] where the type has [key: string] index signatures

		// ==========================================
		// Ternary with reversed null/undefined and equality operators
		// ==========================================
		{
			Code:    `declare const x: string | null | undefined; null !== x && undefined !== x ? x : y;`,
			Output:  []string{`declare const x: string | null | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare const x: string | null | undefined; null === x || undefined === x ? y : x;`,
			Output:  []string{`declare const x: string | null | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// any type tests - should be flagged because any can include null/undefined
		// When ignorePrimitives is set, any/unknown are NOT flagged (can't make assumptions)
		// When ignorePrimitives is NOT set, any/unknown ARE flagged
		// ==========================================
		{
			Code:   `declare const x: any; x || y;`,
			Output: []string{`declare const x: any; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare const x: any; x ||= y;`,
			Output: []string{`declare const x: any; x ??= y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: any; x ? x : y;`,
			Output:  []string{`declare const x: any; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare const x: any; !x ? y : x;`,
			Output:  []string{`declare const x: any; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// unknown type tests - should be flagged because unknown can include null/undefined
		// ==========================================
		{
			Code:   `declare const x: unknown; x || y;`,
			Output: []string{`declare const x: unknown; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `declare const x: unknown; x ||= y;`,
			Output: []string{`declare const x: unknown; x ??= y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: unknown; x ? x : y;`,
			Output:  []string{`declare const x: unknown; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare const x: unknown; !x ? y : x;`,
			Output:  []string{`declare const x: unknown; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Template literal types - should be flagged like string types
		// ==========================================
		{
			Code:   "declare const x: `hello${string}` | undefined; x || 'a';",
			Output: []string{"declare const x: `hello${string}` | undefined; x ?? 'a';"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    "declare const x: `hello${string}` | undefined; x ? x : 'a';",
			Output:  []string{"declare const x: `hello${string}` | undefined; x ?? 'a';"},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    "declare const x: `hello${string}` | undefined; !x ? 'a' : x;",
			Output:  []string{"declare const x: `hello${string}` | undefined; x ?? 'a';"},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Enum member access union tests (Enum.A | Enum.B | undefined)
		// ==========================================
		{
			Code:   `enum Enum { A = 0, B = 1, C = 2 } declare const x: Enum.A | Enum.B | undefined; x || y;`,
			Output: []string{`enum Enum { A = 0, B = 1, C = 2 } declare const x: Enum.A | Enum.B | undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `enum Enum { A = 'a', B = 'b', C = 'c' } declare const x: Enum.A | Enum.B | undefined; x || y;`,
			Output: []string{`enum Enum { A = 'a', B = 'b', C = 'c' } declare const x: Enum.A | Enum.B | undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `enum Enum { A = 0, B = 1, C = 2 } declare const x: Enum.A | Enum.B | undefined; x ? x : y;`,
			Output:  []string{`enum Enum { A = 0, B = 1, C = 2 } declare const x: Enum.A | Enum.B | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `enum Enum { A = 'a', B = 'b', C = 'c' } declare const x: Enum.A | Enum.B | undefined; x ? x : y;`,
			Output:  []string{`enum Enum { A = 'a', B = 'b', C = 'c' } declare const x: Enum.A | Enum.B | undefined; x ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Box interface with optional chaining tests
		// ==========================================
		{
			Code:   `interface Box { a?: { b?: string } } declare const defaultBoxOptional: Box | undefined; defaultBoxOptional?.a?.b || 'a';`,
			Output: []string{`interface Box { a?: { b?: string } } declare const defaultBoxOptional: Box | undefined; defaultBoxOptional?.a?.b ?? 'a';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `interface Box { a?: { b?: string } } declare const defaultBoxOptional: Box | undefined; defaultBoxOptional?.a?.b ? defaultBoxOptional.a.b : 'a';`,
			Output:  []string{`interface Box { a?: { b?: string } } declare const defaultBoxOptional: Box | undefined; defaultBoxOptional?.a?.b ?? 'a';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// NOTE: Double negation patterns like !!a || 'foo' are not flagged because
		// the !!a expression has type boolean, which is not nullable.

		// ==========================================
		// ignoreConditionalTests edge cases with nested logical expressions
		// ==========================================
		{
			Code:    `declare const x: string | null; (x || 'a') && 'b';`,
			Output:  []string{`declare const x: string | null; (x ?? 'a') && 'b';`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:    `declare const x: string | null; 'a' && (x || 'b');`,
			Output:  []string{`declare const x: string | null; 'a' && (x ?? 'b');`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreConditionalTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Complex member access with index signatures (from upstream)
		// These test deeply nested property access with various indexing patterns
		// ==========================================
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] !== undefined && x.z[1][this[this.o]]["3"][a.b.c] !== null ? x.z[1][this[this.o]]["3"][a.b.c] : y;`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] !== null && x.z[1][this[this.o]]["3"][a.b.c] !== undefined ? x.z[1][this[this.o]]["3"][a.b.c] : y;`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] === undefined || x.z[1][this[this.o]]["3"][a.b.c] === null ? y : x.z[1][this[this.o]]["3"][a.b.c];`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] != undefined ? x.z[1][this[this.o]]["3"][a.b.c] : y;`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] != null ? x.z[1][this[this.o]]["3"][a.b.c] : y;`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] == null ? y : x.z[1][this[this.o]]["3"][a.b.c];`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? y;`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Assignment expression in alternate with complex member access
		// ==========================================
		{
			Code:    `x.z[1][this[this.o]]["3"][a.b.c] !== undefined && x.z[1][this[this.o]]["3"][a.b.c] !== null ? x.z[1][this[this.o]]["3"][a.b.c] : (z = y);`,
			Output:  []string{`x.z[1][this[this.o]]["3"][a.b.c] ?? (z = y);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},

		// ==========================================
		// Weird parentheses in if-statements (noFormat from upstream)
		// ==========================================
		{
			Code:   `declare let foo: { a: string | null }; declare function makeString(): string; function weirdParens() { if (((((foo.a)) == null))) { ((((((((foo).a))))) = makeString()); } }`,
			Output: []string{`declare let foo: { a: string | null }; declare function makeString(): string; function weirdParens() { ((foo).a) ??= makeString(); }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverAssignment"}},
		},

		// ==========================================
		// Null and undefined literal tests (from upstream)
		// ==========================================
		{
			Code:   `declare let x: null; x || y;`,
			Output: []string{`declare let x: null; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `const x = undefined; x || y;`,
			Output: []string{`const x = undefined; x ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `null || y;`,
			Output: []string{`null ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},
		{
			Code:   `undefined || y;`,
			Output: []string{`undefined ?? y;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr"}},
		},

		// ==========================================
		// Nested ternary tests (from upstream)
		// ==========================================
		{
			Code:    `let a: string | undefined; let b: { message: string } | undefined; const foo = a ? a : (b ? 1 : 2);`,
			Output:  []string{`let a: string | undefined; let b: { message: string } | undefined; const foo = a ?? (b ? 1 : 2);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
		{
			Code:    `declare const c: string | null; c !== null ? c : c ? 1 : 2;`,
			Output:  []string{`declare const c: string | null; c ?? (c ? 1 : 2);`},
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignoreTernaryTests": false}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverTernary"}},
		},
	})
}
