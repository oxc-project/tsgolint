package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamOrEmptyObject tests the || {} pattern from typescript-eslint
// Source: https://github.com/typescript-eslint/typescript-eslint/.../prefer-optional-chain.test.ts (lines 10-693)
// These tests verify that expressions like `(foo || {}).bar` are converted to `foo?.bar`
func TestUpstreamOrEmptyObject(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Basic || {} pattern
			{
				Code:    `(foo || {}).bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// With extra parens around {}
			{
				Code:    `(foo || ({})).bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// With await
			{
				Code:    `(await foo || {}).bar;`,
				Output:  []string{`(await foo)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Already has optional chain
			{
				Code:    `(foo1?.foo2 || {}).foo3;`,
				Output:  []string{`foo1?.foo2?.foo3;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Already has optional chain with extra parens
			{
				Code:    `(foo1?.foo2 || ({})).foo3;`,
				Output:  []string{`foo1?.foo2?.foo3;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Call expression
			{
				Code:    `((() => foo())() || {}).bar;`,
				Output:  []string{`(() => foo())()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// In variable declaration
			{
				Code:    `const foo = (bar || {}).baz;`,
				Output:  []string{`const foo = bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Element access
			{
				Code:    `(foo.bar || {})[baz];`,
				Output:  []string{`foo.bar?.[baz];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Nested || {} patterns - This creates multiple errors in upstream
			// Note: Upstream has 2 separate suggestions for this case, but our implementation
			// combines them into a single fix. TODO: Match upstream behavior
			{
				Code:   `((foo1 || {}).foo2 || {}).foo3;`,
				Output: []string{`(foo1 || {}).foo2?.foo3;`, `foo1?.foo2?.foo3;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// With undefined in OR chain
			{
				Code:    `(foo || undefined || {}).bar;`,
				Output:  []string{`(foo || undefined)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Multiple expressions in OR
			{
				Code:    `(foo() || bar || {}).baz;`,
				Output:  []string{`(foo() || bar)?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Ternary in OR
			{
				Code:    `((foo1 ? foo2 : foo3) || {}).foo4;`,
				Output:  []string{`(foo1 ? foo2 : foo3)?.foo4;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Inside if statement - without unsafe option, should use suggestions only
			{
				Code: `
          if (foo) {
            (foo || {}).bar;
          }
        `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output: `
          if (foo) {
            foo?.bar;
          }
        `,
					}},
				}},
			},
			// In if condition - without unsafe option, should use suggestions only
			{
				Code: `
          if ((foo || {}).bar) {
            foo.bar;
          }
        `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output: `
          if (foo?.bar) {
            foo.bar;
          }
        `,
					}},
				}},
			},
			// With && in OR chain
			{
				Code:    `(undefined && foo || {}).bar;`,
				Output:  []string{`(undefined && foo)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Using ?? instead of ||
			{
				Code:    `(foo ?? {}).bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with extra parens
			{
				Code:    `(foo ?? ({})).bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with await
			{
				Code:    `(await foo ?? {}).bar;`,
				Output:  []string{`(await foo)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with optional chain
			{
				Code:    `(foo1?.foo2 ?? {}).foo3;`,
				Output:  []string{`foo1?.foo2?.foo3;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with call expression
			{
				Code:    `((() => foo())() ?? {}).bar;`,
				Output:  []string{`(() => foo())()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? in variable declaration
			{
				Code:    `const foo = (bar ?? {}).baz;`,
				Output:  []string{`const foo = bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with element access
			{
				Code:    `(foo.bar ?? {})[baz];`,
				Output:  []string{`foo.bar?.[baz];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Nested ?? patterns
			// Note: Upstream has 2 separate suggestions, but our implementation combines them
			// TODO: Match upstream behavior
			{
				Code:   `((foo1 ?? {}).foo2 ?? {}).foo3;`,
				Output: []string{`(foo1 ?? {}).foo2?.foo3;`, `foo1?.foo2?.foo3;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with undefined in chain
			{
				Code:    `(foo ?? undefined ?? {}).bar;`,
				Output:  []string{`(foo ?? undefined)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with multiple expressions
			{
				Code:    `(foo() ?? bar ?? {}).baz;`,
				Output:  []string{`(foo() ?? bar)?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? with ternary
			{
				Code:    `((foo1 ? foo2 : foo3) ?? {}).foo4;`,
				Output:  []string{`(foo1 ? foo2 : foo3)?.foo4;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// ?? inside if statement - without unsafe option, should use suggestions only
			{
				Code: `
          if (foo) {
            (foo ?? {}).bar;
          }
        `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output: `
          if (foo) {
            foo?.bar;
          }
        `,
					}},
				}},
			},
			// ?? in if condition - without unsafe option, should use suggestions only
			{
				Code: `
          if ((foo ?? {}).bar) {
            foo.bar;
          }
        `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output: `
          if (foo?.bar) {
            foo.bar;
          }
        `,
					}},
				}},
			},
			// ?? with && in chain
			{
				Code:    `(undefined && foo ?? {}).bar;`,
				Output:  []string{`(undefined && foo)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Call expression chains
			{
				Code:    `(foo.bar() || {}).baz;`,
				Output:  []string{`foo.bar()?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Multiple property accesses
			{
				Code:    `(foo.bar.baz || {}).buzz;`,
				Output:  []string{`foo.bar.baz?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// With computed property access
			{
				Code:    `(foo[bar] || {}).baz;`,
				Output:  []string{`foo[bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// With method call and property
			{
				Code:    `(foo.bar().baz || {}).buzz;`,
				Output:  []string{`foo.bar().baz?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Deeply nested
			{
				Code:    `(foo.bar.baz.buzz || {}).fizz;`,
				Output:  []string{`foo.bar.baz.buzz?.fizz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// In return statement
			{
				Code:    `function test() { return (foo || {}).bar; }`,
				Output:  []string{`function test() { return foo?.bar; }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// In arrow function
			{
				Code:    `const test = () => (foo || {}).bar;`,
				Output:  []string{`const test = () => foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// As function argument
			{
				Code:    `console.log((foo || {}).bar);`,
				Output:  []string{`console.log(foo?.bar);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// In template literal
			{
				Code:    "const x = `value: ${(foo || {}).bar}`;",
				Output:  []string{"const x = `value: ${foo?.bar}`;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
