package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamComparisonChains tests chains ending with comparison operators
// Source: https://github.com/typescript-eslint/typescript-eslint/.../prefer-optional-chain.test.ts (lines 694-1884)
// These tests verify that expressions like `foo && foo.bar == 0` are converted to `foo?.bar == 0`
func TestUpstreamComparisonChains(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// These patterns should NOT be converted because the trailing comparison changes semantics
		// record['key'] && record['key'].kind !== '1' - element access with !== comparison
		{Code: `declare const record: Record<string, { kind: string }>; record['key'] && record['key'].kind !== '1';`},
		// !array[1] || array[1].b === 'foo' - negated element access with === comparison
		{Code: `declare const array: { b?: string }[]; !array[1] || array[1].b === 'foo';`},

		// !foo && patterns with comparisons - these are VALID (not converted)
		// The !foo check is inverted in && chains, so these don't match the optional chain pattern
		{Code: `!foo && foo.bar == 0;`},
		{Code: `!foo && foo.bar == 1;`},
		{Code: `!foo && foo.bar == '123';`},
		{Code: `!foo && foo.bar == {};`},
		{Code: `!foo && foo.bar == false;`},
		{Code: `!foo && foo.bar == true;`},
		{Code: `!foo && foo.bar === 0;`},
		{Code: `!foo && foo.bar === 1;`},
		{Code: `!foo && foo.bar === '123';`},
		{Code: `!foo && foo.bar === {};`},
		{Code: `!foo && foo.bar === false;`},
		{Code: `!foo && foo.bar === true;`},
		{Code: `!foo && foo.bar === null;`},
		{Code: `!foo && foo.bar !== undefined;`},
		{Code: `!foo && foo.bar != undefined;`},
		{Code: `!foo && foo.bar != null;`},

		// foo == null && patterns - these are VALID (not converted)
		// foo == null is inverted in && chains (would need foo != null for conversion)
		{Code: `foo == null && foo.bar == 0;`},
		{Code: `foo == null && foo.bar == 1;`},
		{Code: `foo == null && foo.bar == '123';`},
		{Code: `foo == null && foo.bar == {};`},
		{Code: `foo == null && foo.bar == false;`},
		{Code: `foo == null && foo.bar == true;`},
		{Code: `foo == null && foo.bar === 0;`},
		{Code: `foo == null && foo.bar === 1;`},
		{Code: `foo == null && foo.bar === '123';`},
		{Code: `foo == null && foo.bar === {};`},
		{Code: `foo == null && foo.bar === false;`},
		{Code: `foo == null && foo.bar === true;`},
		{Code: `foo == null && foo.bar === null;`},
		{Code: `foo == null && foo.bar !== undefined;`},
		{Code: `foo == null && foo.bar != null;`},
		{Code: `foo == null && foo.bar != undefined;`},
	},
		[]rule_tester.InvalidTestCase{
			// Basic && with == comparisons
			{Code: `foo && foo.bar == 0;`, Output: []string{`foo?.bar == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar == 1;`, Output: []string{`foo?.bar == 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar == '123';`, Output: []string{`foo?.bar == '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar == {};`, Output: []string{`foo?.bar == {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar == false;`, Output: []string{`foo?.bar == false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar == true;`, Output: []string{`foo?.bar == true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Basic && with === comparisons
			{Code: `foo && foo.bar === 0;`, Output: []string{`foo?.bar === 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === 1;`, Output: []string{`foo?.bar === 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === '123';`, Output: []string{`foo?.bar === '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === {};`, Output: []string{`foo?.bar === {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === false;`, Output: []string{`foo?.bar === false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === true;`, Output: []string{`foo?.bar === true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar === null;`, Output: []string{`foo?.bar === null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Basic && with !== and != comparisons
			{Code: `foo && foo.bar !== undefined;`, Output: []string{`foo?.bar !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar != undefined;`, Output: []string{`foo?.bar != undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo.bar != null;`, Output: []string{`foo?.bar != null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// != null && with comparisons
			{Code: `foo != null && foo.bar == 0;`, Output: []string{`foo?.bar == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar == 1;`, Output: []string{`foo?.bar == 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar == '123';`, Output: []string{`foo?.bar == '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar == {};`, Output: []string{`foo?.bar == {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar == false;`, Output: []string{`foo?.bar == false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar == true;`, Output: []string{`foo?.bar == true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === 0;`, Output: []string{`foo?.bar === 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === 1;`, Output: []string{`foo?.bar === 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === '123';`, Output: []string{`foo?.bar === '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === {};`, Output: []string{`foo?.bar === {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === false;`, Output: []string{`foo?.bar === false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === true;`, Output: []string{`foo?.bar === true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar === null;`, Output: []string{`foo?.bar === null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar !== undefined;`, Output: []string{`foo?.bar !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar != undefined;`, Output: []string{`foo?.bar != undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo.bar != null;`, Output: []string{`foo?.bar != null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// With type declarations
			{
				Code: `
          declare const foo: { bar: number };
          foo && foo.bar != null;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != null;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `
          declare const foo: { bar: number };
          foo != null && foo.bar != null;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != null;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// !foo || patterns (negation with OR)
			{Code: `!foo || foo.bar != 0;`, Output: []string{`foo?.bar != 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != 1;`, Output: []string{`foo?.bar != 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != '123';`, Output: []string{`foo?.bar != '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != {};`, Output: []string{`foo?.bar != {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != false;`, Output: []string{`foo?.bar != false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != true;`, Output: []string{`foo?.bar != true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar === undefined;`, Output: []string{`foo?.bar === undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar == undefined;`, Output: []string{`foo?.bar == undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar == null;`, Output: []string{`foo?.bar == null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== 0;`, Output: []string{`foo?.bar !== 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== 1;`, Output: []string{`foo?.bar !== 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== '123';`, Output: []string{`foo?.bar !== '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== {};`, Output: []string{`foo?.bar !== {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== false;`, Output: []string{`foo?.bar !== false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== true;`, Output: []string{`foo?.bar !== true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== null;`, Output: []string{`foo?.bar !== null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// foo == null || patterns
			{Code: `foo == null || foo.bar != 0;`, Output: []string{`foo?.bar != 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar != 1;`, Output: []string{`foo?.bar != 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar != '123';`, Output: []string{`foo?.bar != '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar != {};`, Output: []string{`foo?.bar != {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar != false;`, Output: []string{`foo?.bar != false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar != true;`, Output: []string{`foo?.bar != true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar === undefined;`, Output: []string{`foo?.bar === undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar == undefined;`, Output: []string{`foo?.bar == undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar == null;`, Output: []string{`foo?.bar == null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== 0;`, Output: []string{`foo?.bar !== 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== 1;`, Output: []string{`foo?.bar !== 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== '123';`, Output: []string{`foo?.bar !== '123';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== {};`, Output: []string{`foo?.bar !== {};`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== false;`, Output: []string{`foo?.bar !== false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== true;`, Output: []string{`foo?.bar !== true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo == null || foo.bar !== null;`, Output: []string{`foo?.bar !== null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// With type declarations and !foo ||
			{
				Code: `
          declare const foo: { bar: number };
          !foo || foo.bar == null;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == null;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `
          declare const foo: { bar: number };
          !foo || foo.bar == undefined;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == undefined;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `
          declare const foo: { bar: number };
          !foo || foo.bar === undefined;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar === undefined;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== 0;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 0;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== 1;
        `,
				Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 1;
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Note: !foo && and foo == null && patterns with comparisons are VALID (not converted)
			// See valid section below for these patterns
		})
}
