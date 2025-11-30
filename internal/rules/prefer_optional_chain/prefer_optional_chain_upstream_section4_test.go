package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamBaseCasesWithTransformations tests base cases with various transformations
// Source: https://github.com/typescript-eslint/typescript-eslint/.../prefer-optional-chain.test.ts (lines 3113-end)
// These are variations of the 26 base cases with different operators and conditions
func TestUpstreamBaseCasesWithTransformations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Base cases with trailing expressions (should ignore non-chain parts)
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
foo && foo.bar && bing;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
foo?.bar && bing;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar && foo.bar.baz && bing.bong;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz && bing.bong;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// != null transformations (these are valid for both null and undefined)
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
foo != null && foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar != null && foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 3
declare const foo: (() => number) | null | undefined;
foo != null && foo();`,
				Output: []string{`// 3
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// != undefined transformations
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
foo != undefined && foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar != undefined && foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 3
declare const foo: (() => number) | null | undefined;
foo != undefined && foo();`,
				Output: []string{`// 3
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// !==  undefined with type only having undefined (not null)
			{
				Code: `// 1
declare const foo: {bar: number} | undefined;
foo !== undefined && foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | undefined};
foo.bar !== undefined && foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// !== null with type only having null (not undefined)
			{
				Code: `// 1
declare const foo: {bar: number} | null;
foo !== null && foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null};
foo.bar !== null && foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Complex !== undefined chains
			{
				Code: `
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar !== undefined &&
                  foo.bar() !== undefined &&
                  foo.bar().baz !== undefined &&
                  foo.bar().baz.buzz !== undefined &&
                  foo.bar().baz.buzz();
              `,
				Output: []string{`
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar?.() !== undefined &&
                  foo.bar().baz !== undefined &&
                  foo.bar().baz.buzz !== undefined &&
                  foo.bar().baz.buzz();
              `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar !== undefined && foo.bar?.() !== undefined && foo.bar?.().baz;
              `,
				Output: []string{`
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar?.() !== undefined && foo.bar?.().baz;
              `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR operator with negation (! prefix)
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
!foo || !foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
!foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
!foo.bar || !foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
!foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 3
declare const foo: (() => number) | null | undefined;
!foo || !foo();`,
				Output: []string{`// 3
declare const foo: (() => number) | null | undefined;
!foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 4
declare const foo: {bar: (() => number) | null | undefined};
!foo.bar || !foo.bar();`,
				Output: []string{`// 4
declare const foo: {bar: (() => number) | null | undefined};
!foo.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz.buzz;`,
				Output: []string{`// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// == null with OR (need to check end as well)
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
foo == null || foo.bar == null;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
foo?.bar == null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar == null || foo.bar.baz == null;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz == null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// === null with OR (only works when type has only null, not undefined)
			{
				Code: `// 1
declare const foo: {bar: number} | null;
foo === null || foo.bar === null;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null;
foo?.bar === null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null};
foo.bar === null || foo.bar.baz === null;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null};
foo.bar?.baz === null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// === undefined with OR (only works when type has only undefined, not null)
			{
				Code: `// 1
declare const foo: {bar: number} | undefined;
foo === undefined || foo.bar === undefined;`,
				Output: []string{`// 1
declare const foo: {bar: number} | undefined;
foo?.bar === undefined;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | undefined};
foo.bar === undefined || foo.bar.baz === undefined;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | undefined};
foo.bar?.baz === undefined;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Complex === undefined chains with OR
			// Note: Rule produces multiple passes for this complex chain
			{
				// With unsafe option, produces single pass converting entire chain
				Code: `
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar === undefined ||
                  foo.bar() === undefined ||
                  foo.bar().baz === undefined ||
                  foo.bar().baz.buzz === undefined ||
                  foo.bar().baz.buzz();
              `,
				Output: []string{`
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar?.()?.baz?.buzz?.();
              `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				// From upstream - produces single output as expected
				Code: `
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar === undefined || foo.bar?.() === undefined || foo.bar?.().baz;
              `,
				Output: []string{`
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar?.()?.baz;
              `},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// typeof variations with member access
			{
				Code: `// typeof !== 'undefined' with member access
declare const foo: {bar: number} | null | undefined;
typeof foo !== 'undefined' && foo.bar;`,
				Output: []string{`// typeof !== 'undefined' with member access
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// typeof != 'undefined' with nested member
declare const foo: {bar: {baz: number} | null | undefined};
typeof foo.bar != 'undefined' && foo.bar.baz;`,
				Output: []string{`// typeof != 'undefined' with nested member
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// typeof with call expression
declare const foo: (() => number) | null | undefined;
typeof foo !== 'undefined' && foo();`,
				Output: []string{`// typeof with call expression
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// typeof with method call
declare const foo: {bar: (() => number) | null | undefined};
typeof foo.bar !== 'undefined' && foo.bar();`,
				Output: []string{`// typeof with method call
declare const foo: {bar: (() => number) | null | undefined};
foo.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// typeof with deep chain
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
typeof foo !== 'undefined' && typeof foo.bar !== 'undefined' && typeof foo.bar.baz !== 'undefined' && foo.bar.baz.buzz;`,
				Output: []string{`// typeof with deep chain
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Yoda-style typeof comparisons
			{
				Code: `// Yoda: 'undefined' !== typeof
declare const foo: {bar: number} | null | undefined;
'undefined' !== typeof foo && foo.bar;`,
				Output: []string{`// Yoda: 'undefined' !== typeof
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda: 'undefined' != typeof with nested
declare const foo: {bar: {baz: number} | null | undefined};
'undefined' != typeof foo.bar && foo.bar.baz;`,
				Output: []string{`// Yoda: 'undefined' != typeof with nested
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Element access with various checks
			{
				Code: `// Element access with != null
declare const foo: {[key: string]: number} | null | undefined;
foo != null && foo['bar'];`,
				Output: []string{`// Element access with != null
declare const foo: {[key: string]: number} | null | undefined;
foo?.['bar'];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Element access with !== undefined
declare const foo: {[key: string]: {baz: number}} | null | undefined;
foo !== undefined && foo['bar'] !== undefined && foo['bar'].baz;`,
				Output: []string{`// Element access with !== undefined
declare const foo: {[key: string]: {baz: number}} | null | undefined;
foo?.['bar'] !== undefined && foo['bar'].baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Numeric element access
declare const foo: number[] | null | undefined;
foo != null && foo[0];`,
				Output: []string{`// Numeric element access
declare const foo: number[] | null | undefined;
foo?.[0];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Variable element access
declare const foo: {[key: string]: number} | null | undefined;
declare const key: string;
foo != null && foo[key];`,
				Output: []string{`// Variable element access
declare const foo: {[key: string]: number} | null | undefined;
declare const key: string;
foo?.[key];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Negation patterns with typeof (OR operator)
			{
				Code: `// typeof === 'undefined' with OR
declare const foo: {bar: number} | null | undefined;
typeof foo === 'undefined' || foo.bar;`,
				Output: []string{`// typeof === 'undefined' with OR
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda typeof === 'undefined' with OR
declare const foo: {bar: {baz: number} | null | undefined};
'undefined' === typeof foo.bar || foo.bar.baz;`,
				Output: []string{`// Yoda typeof === 'undefined' with OR
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// typeof == 'undefined' with OR and call
declare const foo: (() => number) | null | undefined;
typeof foo == 'undefined' || foo();`,
				Output: []string{`// typeof == 'undefined' with OR and call
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Mixed chains with various patterns
			{
				Code: `// Member then call then member
declare const foo: {bar: (() => {baz: number}) | null | undefined} | null | undefined;
foo != null && foo.bar != null && foo.bar() != null && foo.bar().baz;`,
				Output: []string{`// Member then call then member
declare const foo: {bar: (() => {baz: number}) | null | undefined} | null | undefined;
foo?.bar?.()?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Member then element then call
declare const foo: {bar: {[key: string]: (() => number) | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar['baz'] && foo.bar['baz']();`,
				Output: []string{`// Member then element then call
declare const foo: {bar: {[key: string]: (() => number) | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.['baz']?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Element then member then call
declare const foo: {[key: string]: {bar: (() => number) | null | undefined} | null | undefined} | null | undefined;
foo !== undefined && foo['key'] !== undefined && foo['key'].bar !== undefined && foo['key'].bar();`,
				Output: []string{`// Element then member then call
declare const foo: {[key: string]: {bar: (() => number) | null | undefined} | null | undefined} | null | undefined;
foo?.['key']?.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Call then element then member
declare const foo: (() => {[key: string]: {bar: number} | null | undefined} | null | undefined) | null | undefined;
foo && foo() && foo()['key'] && foo()['key'].bar;`,
				Output: []string{`// Call then element then member
declare const foo: (() => {[key: string]: {bar: number} | null | undefined} | null | undefined) | null | undefined;
foo?.()?.['key']?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Yoda-style null/undefined comparisons
			{
				Code: `// Yoda: null !== foo
declare const foo: {bar: number} | null | undefined;
null !== foo && foo.bar;`,
				Output: []string{`// Yoda: null !== foo
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda: undefined !== foo.bar
declare const foo: {bar: {baz: number} | null | undefined};
undefined !== foo.bar && foo.bar.baz;`,
				Output: []string{`// Yoda: undefined !== foo.bar
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda: null != foo with call
declare const foo: (() => number) | null | undefined;
null != foo && foo();`,
				Output: []string{`// Yoda: null != foo with call
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda: undefined != foo.bar with method
declare const foo: {bar: (() => number) | null | undefined};
undefined != foo.bar && foo.bar();`,
				Output: []string{`// Yoda: undefined != foo.bar with method
declare const foo: {bar: (() => number) | null | undefined};
foo.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Yoda-style with OR operator
			{
				Code: `// Yoda: null === foo with OR
declare const foo: {bar: number} | null;
null === foo || foo.bar === null;`,
				Output: []string{`// Yoda: null === foo with OR
declare const foo: {bar: number} | null;
foo?.bar === null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda: undefined === foo.bar with OR
declare const foo: {bar: {baz: number} | undefined};
undefined === foo.bar || undefined === foo.bar.baz;`,
				Output: []string{`// Yoda: undefined === foo.bar with OR
declare const foo: {bar: {baz: number} | undefined};
foo.bar?.baz === undefined;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Complex nested structures with different check patterns
			{
				Code: `// Nested with mixed != null and !== undefined
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo != null && foo.bar !== undefined && foo.bar.baz != null && foo.bar.baz.buzz;`,
				Output: []string{`// Nested with mixed != null and !== undefined
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Nested with mixed typeof and direct checks
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
typeof foo !== 'undefined' && foo.bar != null && typeof foo.bar.baz !== 'undefined' && foo.bar.baz.buzz;`,
				Output: []string{`// Nested with mixed typeof and direct checks
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Nested with Yoda and regular checks
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
null !== foo && undefined !== foo.bar && foo.bar.baz != null && foo.bar.baz.buzz;`,
				Output: []string{`// Nested with Yoda and regular checks
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Element access with Yoda-style
			{
				Code: `// Yoda with element access
declare const foo: {[key: string]: number} | null | undefined;
null != foo && foo['bar'];`,
				Output: []string{`// Yoda with element access
declare const foo: {[key: string]: number} | null | undefined;
foo?.['bar'];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Yoda with numeric element access
declare const foo: number[] | null | undefined;
undefined !== foo && foo[0];`,
				Output: []string{`// Yoda with numeric element access
declare const foo: number[] | null | undefined;
foo?.[0];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Call chains with various check patterns
			{
				Code: `// Call chain with != null checks
declare const foo: {bar: () => {baz: () => {buzz: () => number} | null} | null} | null;
foo != null && foo.bar != null && foo.bar() != null && foo.bar().baz != null && foo.bar().baz() != null && foo.bar().baz().buzz != null && foo.bar().baz().buzz();`,
				Output: []string{`// Call chain with != null checks
declare const foo: {bar: () => {baz: () => {buzz: () => number} | null} | null} | null;
foo?.bar?.()?.baz?.()?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Call chain with !== undefined checks
declare const foo: {bar: () => {baz: () => number} | undefined} | undefined;
foo !== undefined && foo.bar !== undefined && foo.bar() !== undefined && foo.bar().baz !== undefined && foo.bar().baz();`,
				Output: []string{`// Call chain with !== undefined checks
declare const foo: {bar: () => {baz: () => number} | undefined} | undefined;
foo?.bar?.()?.baz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// OR operator with element access
			{
				Code: `// OR with element access == null
declare const foo: {[key: string]: number} | null | undefined;
foo == null || foo['bar'] == null;`,
				Output: []string{`// OR with element access == null
declare const foo: {[key: string]: number} | null | undefined;
foo?.['bar'] == null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// OR with numeric element === null
declare const foo: (number | null)[] | null;
foo === null || foo[0] === null;`,
				Output: []string{`// OR with numeric element === null
declare const foo: (number | null)[] | null;
foo?.[0] === null;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Deep chains with alternating check styles
			{
				Code: `// 5-level deep with alternating checks
declare const foo: {bar: {baz: {buzz: {fizz: number} | null} | undefined} | null} | undefined;
foo !== undefined && foo.bar != null && typeof foo.bar.baz !== 'undefined' && foo.bar.baz.buzz !== null && foo.bar.baz.buzz.fizz;`,
				Output: []string{`// 5-level deep with alternating checks
declare const foo: {bar: {baz: {buzz: {fizz: number} | null} | undefined} | null} | undefined;
foo?.bar?.baz?.buzz?.fizz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 5-level deep with all Yoda checks
declare const foo: {bar: {baz: {buzz: {fizz: number} | null | undefined} | null | undefined} | null | undefined} | null | undefined;
null != foo && undefined !== foo.bar && null != foo.bar.baz && undefined !== foo.bar.baz.buzz && foo.bar.baz.buzz.fizz;`,
				Output: []string{`// 5-level deep with all Yoda checks
declare const foo: {bar: {baz: {buzz: {fizz: number} | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz?.fizz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Mixed access types in single chain
			{
				Code: `// Member -> Element -> Call -> Member
declare const foo: {bar: {[key: string]: (() => {baz: number}) | null} | null} | null;
foo && foo.bar && foo.bar['key'] && foo.bar['key']() && foo.bar['key']().baz;`,
				Output: []string{`// Member -> Element -> Call -> Member
declare const foo: {bar: {[key: string]: (() => {baz: number}) | null} | null} | null;
foo?.bar?.['key']?.()?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Element -> Call -> Member -> Element
declare const foo: {[key: string]: (() => {bar: {[key: string]: number} | null}) | null} | null;
foo != null && foo['key'] != null && foo['key']() != null && foo['key']().bar != null && foo['key']().bar['baz'];`,
				Output: []string{`// Element -> Call -> Member -> Element
declare const foo: {[key: string]: (() => {bar: {[key: string]: number} | null}) | null} | null;
foo?.['key']?.()?.bar?.['baz'];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				// invalid-65: Multi-pass conversion - first pass converts foo !== null && foo(), second pass continues
				Code: `// Call -> Member -> Call -> Element
declare const foo: (() => {bar: (() => {[key: string]: number} | null) | null}) | null;
foo !== null && foo() !== null && foo().bar !== null && foo().bar() !== null && foo().bar()['key'];`,
				Output: []string{`// Call -> Member -> Call -> Element
declare const foo: (() => {bar: (() => {[key: string]: number} | null) | null}) | null;
foo?.() !== null && foo().bar !== null && foo().bar() !== null && foo().bar()['key'];`, `// Call -> Member -> Call -> Element
declare const foo: (() => {bar: (() => {[key: string]: number} | null) | null}) | null;
foo?.() !== null && foo().bar?.()?.['key'];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Trailing comparisons with different operators
			{
				Code: `// Trailing && with > comparison
declare const foo: {bar: number} | null | undefined;
foo && foo.bar && foo.bar > 5;`,
				Output: []string{`// Trailing && with > comparison
declare const foo: {bar: number} | null | undefined;
foo?.bar && foo.bar > 5;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Trailing && with < comparison
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar && foo.bar.baz && foo.bar.baz < 10;`,
				Output: []string{`// Trailing && with < comparison
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz && foo.bar.baz < 10;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Trailing && with >= comparison
declare const foo: {bar: {baz: {buzz: number} | null} | null} | null;
foo != null && foo.bar != null && foo.bar.baz != null && foo.bar.baz.buzz >= 0;`,
				Output: []string{`// Trailing && with >= comparison
declare const foo: {bar: {baz: {buzz: number} | null} | null} | null;
foo?.bar?.baz?.buzz >= 0;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Trailing && with <= comparison
declare const foo: {bar: {baz: number} | null | undefined};
typeof foo !== 'undefined' && typeof foo.bar !== 'undefined' && typeof foo.bar.baz !== 'undefined' && foo.bar.baz <= 100;`,
				Output: []string{`// Trailing && with <= comparison
declare const foo: {bar: {baz: number} | null | undefined};
foo?.bar?.baz <= 100;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// OR operator with trailing comparisons
			{
				Code: `// OR with trailing || and > comparison
declare const foo: {bar: number} | null | undefined;
!foo || !foo.bar || foo.bar > 5;`,
				Output: []string{`// OR with trailing || and > comparison
declare const foo: {bar: number} | null | undefined;
!foo?.bar || foo.bar > 5;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// OR with == null and === comparison
declare const foo: {bar: {baz: number} | null} | null;
foo == null || foo.bar == null || foo.bar.baz === 42;`,
				Output: []string{`// OR with == null and === comparison
declare const foo: {bar: {baz: number} | null} | null;
foo?.bar?.baz === 42;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Complex real-world-like patterns
			{
				Code: `// Config object access pattern
declare const config: {api: {endpoint: {url: string} | null} | null} | null;
config && config.api && config.api.endpoint && config.api.endpoint.url;`,
				Output: []string{`// Config object access pattern
declare const config: {api: {endpoint: {url: string} | null} | null} | null;
config?.api?.endpoint?.url;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Event handler pattern
declare const event: {target: {value: string} | undefined} | undefined;
event !== undefined && event.target !== undefined && event.target.value;`,
				Output: []string{`// Event handler pattern
declare const event: {target: {value: string} | undefined} | undefined;
event?.target?.value;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// API response pattern with array access
declare const response: {data: {items: {name: string}[]} | null} | null;
response != null && response.data != null && response.data.items[0];`,
				Output: []string{`// API response pattern with array access
declare const response: {data: {items: {name: string}[]} | null} | null;
response?.data?.items[0];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// Method chaining pattern
declare const obj: {getUser: (() => {getName: (() => string) | null} | null) | null} | null;
obj && obj.getUser && obj.getUser() && obj.getUser().getName && obj.getUser().getName();`,
				Output: []string{`// Method chaining pattern
declare const obj: {getUser: (() => {getName: (() => string) | null} | null) | null} | null;
obj?.getUser?.()?.getName?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},

			// Very deep chains (stress test)
			{
				Code: `// 6-level deep chain
declare const foo: {a: {b: {c: {d: {e: {f: number} | null} | null} | null} | null} | null} | null;
foo && foo.a && foo.a.b && foo.a.b.c && foo.a.b.c.d && foo.a.b.c.d.e && foo.a.b.c.d.e.f;`,
				Output: []string{`// 6-level deep chain
declare const foo: {a: {b: {c: {d: {e: {f: number} | null} | null} | null} | null} | null} | null;
foo?.a?.b?.c?.d?.e?.f;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code: `// 7-level deep chain with mixed checks
declare const foo: {a: {b: {c: {d: {e: {f: {g: number} | undefined} | null} | undefined} | null} | undefined} | null} | undefined;
foo !== undefined && foo.a != null && foo.a.b !== undefined && foo.a.b.c != null && foo.a.b.c.d !== undefined && foo.a.b.c.d.e != null && foo.a.b.c.d.e.f !== undefined && foo.a.b.c.d.e.f.g;`,
				Output: []string{`// 7-level deep chain with mixed checks
declare const foo: {a: {b: {c: {d: {e: {f: {g: number} | undefined} | null} | undefined} | null} | undefined} | null} | undefined;
foo?.a?.b?.c?.d?.e?.f?.g;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}
