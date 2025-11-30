package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamHandCraftedCases tests complex hand-crafted real-world cases
// Source: https://github.com/typescript-eslint/typescript-eslint/.../prefer-optional-chain.test.ts (lines 1885-3112)
// These tests cover edge cases, complex patterns, and real-world scenarios
func TestUpstreamHandCraftedCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Multiple chains in one expression (two errors)
			{
				Code:   `foo && foo.bar && foo.bar.baz || baz && baz.bar && baz.bar.foo`,
				Output: []string{`foo?.bar?.baz || baz?.bar?.foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Inconsistent checks should break the chain
			{
				Code:    `foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;`,
				Output:  []string{`foo?.bar?.baz !== undefined && foo.bar.baz.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          foo.bar &&
            foo.bar.baz != null &&
            foo.bar.baz.qux !== undefined &&
            foo.bar.baz.qux.buzz;
        `,
				Output: []string{`
          foo.bar?.baz?.qux !== undefined &&
            foo.bar.baz.qux.buzz;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// String literal element access
			{
				Code:    `foo && foo['some long string'] && foo['some long string'].baz;`,
				Output:  []string{`foo?.['some long string']?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "foo && foo[`some long string`] && foo[`some long string`].baz;",
				Output:  []string{"foo?.[`some long string`]?.baz;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "foo && foo[`some ${long} string`] && foo[`some ${long} string`].baz;",
				Output:  []string{"foo?.[`some ${long} string`]?.baz;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Complex computed properties
			{
				Code:    `foo && foo[bar as string] && foo[bar as string].baz;`,
				Output:  []string{`foo?.[bar as string]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[1 + 2] && foo[1 + 2].baz;`,
				Output:  []string{`foo?.[1 + 2]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[typeof bar] && foo[typeof bar].baz;`,
				Output:  []string{`foo?.[typeof bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Call expressions
			{
				Code:    `foo() && foo()(bar);`,
				Output:  []string{`foo()?.(bar);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Type parameters
			{
				Code:    `foo && foo<string>() && foo<string>().bar;`,
				Output:  []string{`foo?.<string>()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Binary expressions at end
			{
				Code:    `foo && foo.bar != null;`,
				Output:  []string{`foo?.bar != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar != undefined;`,
				Output:  []string{`foo?.bar != undefined;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar != null && baz;`,
				Output:  []string{`foo?.bar != null && baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// this keyword
			{
				Code:    `this.bar && this.bar.baz;`,
				Output:  []string{`this.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!this.bar || !this.bar.baz;`,
				Output:  []string{`!this.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Already has optional chain
			{
				Code:    `foo && foo?.();`,
				Output:  []string{`foo?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar?.();`,
				Output:  []string{`foo.bar?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Negation patterns
			{
				Code:    `!a.b || !a.b();`,
				Output:  []string{`!a.b?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo.bar || !foo.bar.baz;`,
				Output:  []string{`!foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo[bar] || !foo[bar]?.[baz];`,
				Output:  []string{`!foo[bar]?.[baz];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo || !foo?.bar.baz;`,
				Output:  []string{`!foo?.bar.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Multiple chains (two errors)
			{
				Code:   `(!foo || !foo.bar || !foo.bar.baz) && (!baz || !baz.bar || !baz.bar.foo);`,
				Output: []string{`(!foo?.bar?.baz) && (!baz?.bar?.foo);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// import.meta
			{
				Code:    `import.meta && import.meta?.baz;`,
				Output:  []string{`import.meta?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!import.meta || !import.meta?.baz;`,
				Output:  []string{`!import.meta?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `import.meta && import.meta?.() && import.meta?.().baz;`,
				Output:  []string{`import.meta?.()?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Non-null expressions
			{
				Code:    `!foo() || !foo().bar;`,
				Output:  []string{`!foo()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo!.bar || !foo!.bar.baz;`,
				Output:  []string{`!foo!.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo!.bar!.baz || !foo!.bar!.baz!.paz;`,
				Output:  []string{`!foo!.bar!.baz?.paz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo.bar!.baz || !foo.bar!.baz!.paz;`,
				Output:  []string{`!foo.bar!.baz?.paz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Null checks
			{
				Code:    `foo != null && foo.bar != null;`,
				Output:  []string{`foo?.bar != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          declare const foo: { bar: string | null } | null;
          foo !== null && foo.bar != null;
        `,
				Output: []string{`
          declare const foo: { bar: string | null } | null;
          foo?.bar != null;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Unrelated checks (issue #6332)
			{
				Code:    `unrelated != null && foo != null && foo.bar != null;`,
				Output:  []string{`unrelated != null && foo?.bar != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `unrelated1 != null && unrelated2 != null && foo != null && foo.bar != null;`,
				Output:  []string{`unrelated1 != null && unrelated2 != null && foo?.bar != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Multiple chains in expression (issue #1461 - two errors)
			{
				Code:   `foo1 != null && foo1.bar != null && foo2 != null && foo2.bar != null;`,
				Output: []string{`foo1?.bar != null && foo2?.bar != null;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo && foo.a && bar && bar.a;`,
				Output: []string{`foo?.a && bar?.a;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Parenthesis handling
			{
				Code:    `a && (a.b && a.b.c)`,
				Output:  []string{`a?.b?.c`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `(a && a.b) && a.b.c`,
				Output:  []string{`a?.b?.c`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `((a && a.b)) && a.b.c`,
				Output:  []string{`a?.b?.c`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo(a && (a.b && a.b.c))`,
				Output:  []string{`foo(a?.b?.c)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo(a && a.b && a.b.c)`,
				Output:  []string{`foo(a?.b?.c)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo || !foo.bar || ((((!foo.bar.baz || !foo.bar.baz()))));`,
				Output:  []string{`!foo?.bar?.baz?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a !== undefined && ((a !== null && a.prop));`,
				Output:  []string{`a?.prop;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Function optional call
			{
				Code: `
declare const foo: {
  bar: undefined | (() => void);
};

foo.bar && foo.bar();
        `,
				Output: []string{`
declare const foo: {
  bar: undefined | (() => void);
};

foo.bar?.();
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// typeof checks with globalThis
			{
				Code: `
          function foo(globalThis?: { Array: Function }) {
            typeof globalThis !== 'undefined' && globalThis.Array();
          }
        `,
				Output: []string{`
          function foo(globalThis?: { Array: Function }) {
            globalThis?.Array();
          }
        `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// JSX with typeof (ensure essential whitespace preserved)
			{
				Code:    `foo && foo.bar(baz => <This Requires Spaces />);`,
				Output:  []string{`foo?.bar(baz => <This Requires Spaces />);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar(baz => typeof baz);`,
				Output:  []string{`foo?.bar(baz => typeof baz);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Call arguments are different (should break chain)
			{
				Code:    `foo && foo.bar(a) && foo.bar(a, b).baz;`,
				Output:  []string{`foo?.bar(a) && foo.bar(a, b).baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Type parameters are different (should break chain)
			{
				Code:    `foo && foo<string>() && foo<string, number>().bar;`,
				Output:  []string{`foo?.<string>() && foo<string, number>().bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Preserve comments in call expressions
			{
				Code: `
          foo && foo.bar(/* comment */a,
            // comment2
            b, );
        `,
				Output: []string{`
          foo?.bar(/* comment */a,
            // comment2
            b, );
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// new.target support - produces suggestion, not autofix
			{
				Code: `
          class Foo {
            constructor() {
              new.target && new.target.length;
            }
          }
        `,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output: `
          class Foo {
            constructor() {
              new.target?.length;
            }
          }
        `,
					}},
				}},
			},

			// await expressions
			{
				Code:    `(await foo).bar && (await foo).bar.baz;`,
				Output:  []string{`(await foo).bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Long OR chain with multiple nullish checks
			// With unsafe option, produces single pass converting entire chain
			// Note: Starting with !a means the output keeps the negation
			{
				Code: `
          !a ||
            a.b == null ||
            a.b.c === undefined ||
            a.b.c === null ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            a.b.c.d.e.f.g == null ||
            a.b.c.d.e.f.g.h;
        `,
				Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Various typeof checks with different styles
			{
				Code: `
          declare const foo: { bar: number } | null | undefined;
          foo && foo.bar != null;
        `,
				Output: []string{`
          declare const foo: { bar: number } | null | undefined;
          foo?.bar != null;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          declare const foo: { bar: number } | undefined;
          foo && typeof foo.bar !== 'undefined';
        `,
				Output: []string{`
          declare const foo: { bar: number } | undefined;
          typeof foo?.bar !== 'undefined';
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          declare const foo: { bar: number } | undefined;
          foo && 'undefined' !== typeof foo.bar;
        `,
				Output: []string{`
          declare const foo: { bar: number } | undefined;
          'undefined' !== typeof foo?.bar;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            undefined !== foo.bar.baz &&
            null !== foo.bar.baz;
        `,
				Output: []string{`
          undefined !== foo?.bar?.baz &&
            null !== foo.bar.baz;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          foo != null &&
            typeof foo.bar !== 'undefined' &&
            foo.bar !== null &&
            foo.bar.baz !== undefined &&
            foo.bar.baz !== null;
        `,
				Output: []string{`
          foo?.bar?.baz !== undefined &&
            foo.bar.baz !== null;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Mixed checks that cause chain breaks
			{
				Code: `
          foo &&
            foo.bar !== null &&
            foo.bar.baz !== undefined &&
            foo.bar.baz.buzz;
        `,
				Output: []string{`
          foo?.bar?.baz !== undefined &&
            foo.bar.baz.buzz;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Randomly placed optional chain tokens (should still optimize)
			{
				Code:    `foo.bar.baz != null && foo?.bar?.baz.bam != null;`,
				Output:  []string{`foo.bar.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo?.bar.baz != null && foo.bar?.baz.bam != null;`,
				Output:  []string{`foo?.bar.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo?.bar?.baz != null && foo.bar.baz.bam != null;`,
				Output:  []string{`foo?.bar?.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Non-null assertions in earlier operands should be retained
			{
				Code:    `foo.bar.baz != null && foo!.bar!.baz.bam != null;`,
				Output:  []string{`foo.bar.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo!.bar.baz != null && foo.bar!.baz.bam != null;`,
				Output:  []string{`foo!.bar.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo!.bar!.baz != null && foo.bar.baz.bam != null;`,
				Output:  []string{`foo!.bar!.baz?.bam != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Long chain with many mixed binary checks
			// TODO: Rule currently stops at typeof check - should convert all the way to a?.b?.c?.d?.e?.f?.g?.h
			{
				Code: `
          a &&
            a.b != null &&
            a.b.c !== undefined &&
            a.b.c !== null &&
            a.b.c.d != null &&
            a.b.c.d.e !== null &&
            a.b.c.d.e !== undefined &&
            a.b.c.d.e.f != undefined &&
            typeof a.b.c.d.e.f.g !== 'undefined' &&
            a.b.c.d.e.f.g !== null &&
            a.b.c.d.e.f.g.h;
        `,
				Output: []string{`
          a?.b?.c?.d?.e?.f?.g !== null &&
            a.b.c.d.e.f.g.h;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR version of long mixed chain
			// With unsafe option, produces single pass
			{
				Code: `
          !a ||
            a.b == null ||
            a.b.c === undefined ||
            a.b.c === null ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            typeof a.b.c.d.e.f.g === 'undefined' ||
            a.b.c.d.e.f.g === null ||
            a.b.c.d.e.f.g.h;
        `,
				Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR chain with negation at end
			{
				Code: `
          !a ||
            a.b == null ||
            a.b.c === undefined ||
            a.b.c === null ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            typeof a.b.c.d.e.f.g === 'undefined' ||
            a.b.c.d.e.f.g === null ||
            !a.b.c.d.e.f.g.h;
        `,
				Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR chain with different order of undefined/null checks
			{
				Code: `
          !a ||
            a.b == null ||
            a.b.c === null ||
            a.b.c === undefined ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            typeof a.b.c.d.e.f.g === 'undefined' ||
            a.b.c.d.e.f.g === null ||
            !a.b.c.d.e.f.g.h;
        `,
				Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Yoda-style checks (value on left side)
			{
				Code:    `undefined !== foo && null !== foo && null != foo.bar && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            foo.bar.baz;
        `,
				Output: []string{`
          foo?.bar?.baz;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            null != foo.bar.baz;
        `,
				Output: []string{`
          null != foo?.bar?.baz;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// requireNullish with nested nullable types (has output)
			{
				Code: `
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo && foo.bar && foo.bar.toString();
        `,
				Output: []string{`
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo?.bar?.toString();
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"requireNullish": true},
			},
			{
				Code: `
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo && foo.bar && foo.bar.toString() && foo.bar.toString();
        `,
				Output: []string{`
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo?.bar?.toString() && foo.bar.toString();
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"requireNullish": true},
			},
			// allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: true (has output)
			{
				Code: `
          declare const foo: { bar: number } | null | undefined;
          foo != undefined && foo.bar;
        `,
				Output: []string{`
          declare const foo: { bar: number } | null | undefined;
          foo?.bar;
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing with function call
			{
				Code: `
          declare const foo: { bar: boolean } | null | undefined;
          declare function acceptsBoolean(arg: boolean): void;
          acceptsBoolean(foo != null && foo.bar);
        `,
				Output: []string{`
          declare const foo: { bar: boolean } | null | undefined;
          declare function acceptsBoolean(arg: boolean): void;
          acceptsBoolean(foo?.bar);
        `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
