package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferOptionalChainRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Valid - already using optional chain
		{Code: `
			foo?.bar;
		`},
		{Code: `
			foo?.bar?.baz;
		`},
		{Code: `
			foo?.bar?.();
		`},

		// Valid - simple cases that don't benefit from optional chain
		{Code: `
			foo && bar;
		`},
		{Code: `
			foo || bar;
		`},

		// Valid - mixed logical expressions
		{Code: `
			foo && bar || baz;
		`},

		// Valid - nullish checks without property access
		{Code: `
		foo != null && bar;
	`},
		{Code: `
		foo !== undefined && bar;
	`},

		// Valid - checkString option disables string checks
		{
			Code: `
		declare const foo: string;
		foo && foo.length;
	`,
			Options: map[string]interface{}{
				"checkString": false,
			},
		},

		// Valid - checkNumber option disables number checks
		{
			Code: `
		declare const foo: number;
		foo && foo.toFixed();
	`,
			Options: map[string]interface{}{
				"checkNumber": false,
			},
		},

		// Valid - checkBoolean option disables boolean checks
		{
			Code: `
		declare const foo: boolean;
		foo && foo.valueOf();
	`,
			Options: map[string]interface{}{
				"checkBoolean": false,
			},
		},

		// Valid - checkBigInt option disables bigint checks
		{
			Code: `
		declare const foo: bigint;
		foo && foo.toString();
	`,
			Options: map[string]interface{}{
				"checkBigInt": false,
			},
		},

		// Valid - checkAny option disables any checks
		{
			Code: `
		declare const foo: any;
		foo && foo.bar;
	`,
			Options: map[string]interface{}{
				"checkAny": false,
			},
		},

		// Valid - checkUnknown option disables unknown checks
		{
			Code: `
		declare const foo: unknown;
		foo && (foo as any).bar;
	`,
			Options: map[string]interface{}{
				"checkUnknown": false,
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic && chain
		{
			Code: `
				foo && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				foo && foo.bar && foo.bar.baz;
			`,
			Output: []string{`
				foo?.bar?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz;
			`,
			Output: []string{`
				foo?.bar?.baz?.buzz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Computed property access
		{
			Code: `
				foo && foo['bar'] && foo['bar'].baz;
			`,
			Output: []string{`
				foo?.['bar']?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Method calls
		{
			Code: `
				foo && foo.bar && foo.bar.baz && foo.bar.baz();
			`,
			Output: []string{`
				foo?.bar?.baz?.();
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Nullish comparison - != null
		{
			Code: `
				foo != null && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Nullish comparison - !== null
		{
			Code: `
				foo !== null && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Nullish comparison - !== undefined
		{
			Code: `
				foo !== undefined && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Combined null and undefined checks
		{
			Code: `
				foo !== null && foo !== undefined && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Yoda condition - null !== foo
		{
			Code: `
				null !== foo && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Yoda condition - undefined !== foo
		{
			Code: `
				undefined !== foo && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Typeof check
		{
			Code: `
				typeof foo !== 'undefined' && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multi-level with nullish check
		{
			Code: `
				foo != null && foo.bar && foo.bar.baz;
			`,
			Output: []string{`
				foo?.bar?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Negated OR chain - !foo || !foo.bar
		{
			Code: `
				!foo || !foo.bar;
			`,
			Output: []string{`
				!foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multi-level negated OR chain
		{
			Code: `
				!foo || !foo.bar || !foo.bar.baz;
			`,
			Output: []string{`
				!foo?.bar?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Test requireNullish option - should NOT convert plain && when requireNullish is true
		{
			Code: `
				foo != null && foo.bar;
			`,
			Output: []string{`
				foo?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				"requireNullish": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// TODO: Empty object pattern - (foo || {}).bar
		// This requires more complex AST analysis to detect the pattern
		// {
		// 	Code: `
		// 		(foo || {}).bar;
		// 	`,
		// 	Output: []string{`
		// 		foo?.bar;
		// 	`},
		// 	Options: map[string]interface{}{
		// 		"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
		// 	},
		// 	Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		// },

		// Complex chain with method call
		{
			Code: `
				foo !== null && foo.bar !== undefined && foo.bar.baz();
			`,
			Output: []string{`
				foo?.bar?.baz();
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Nested property with computed access
		{
			Code: `
				foo != null && foo.bar && foo.bar['baz'];
			`,
			Output: []string{`
				foo?.bar?.['baz'];
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Chains ending with comparison - these should be converted
		{
			Code: `
				foo && foo.bar == 0;
			`,
			Output: []string{`
				foo?.bar == 0;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				foo && foo.bar === null;
			`,
			Output: []string{`
				foo?.bar === null;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				foo != null && foo.bar != null;
			`,
			Output: []string{`
				foo?.bar != null;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Await expressions
		{
			Code: `
				(await foo).bar && (await foo).bar.baz;
			`,
			Output: []string{`
				(await foo).bar?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Await expressions
		{
			Code: `
			(await foo).bar && (await foo).bar.baz;
		`,
			Output: []string{`
			(await foo).bar?.baz;
		`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Type parameters
		{
			Code: `
				foo && foo<string>() && foo<string>().bar;
			`,
			Output: []string{`
				foo?.<string>()?.bar;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// import.meta
		{
			Code: `
				import.meta && import.meta.baz;
			`,
			Output: []string{`
				import.meta?.baz;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// new.target
		{
			Code: `
				class Foo {
					constructor() {
						new.target && new.target.length;
					}
				}
			`,
			Output: []string{`
				class Foo {
					constructor() {
						new.target?.length;
					}
				}
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Negated OR with comparison ending
		{
			Code: `
				!foo || foo.bar != 0;
			`,
			Output: []string{`
				foo?.bar != 0;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				!foo || foo.bar !== null;
			`,
			Output: []string{`
				foo?.bar !== null;
			`},
			Options: map[string]interface{}{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code: `
				foo == null || foo.bar != 0;
			`,
			Output: []string{`
				foo?.bar != 0;
			`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// PHASE 1: Empty Object Pattern with || Operator (Task 1.1-1.20)
		// ============================================================

		// Task 1.1: Basic empty object pattern
		{
			Code: `(foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.bar;`,
				}},
			}},
		},

		// Task 1.2: Parenthesized empty object
		{
			Code: `(foo || ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.bar;`,
				}},
			}},
		},

		// Task 1.3: Await with empty object
		{
			Code: `(await foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(await foo)?.bar;`,
				}},
			}},
		},

		// Task 1.4: Nested optional chain
		{
			Code: `(foo1?.foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo1?.foo2?.foo3;`,
				}},
			}},
		},

		// Task 1.5: Arrow function call
		{
			Code: `((() => foo())() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(() => foo())()?.bar;`,
				}},
			}},
		},

		// Task 1.6: Const assignment
		{
			Code: `const foo = (bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `const foo = bar?.baz;`,
				}},
			}},
		},

		// Task 1.7: Computed property
		{
			Code: `(foo.bar || {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo.bar?.[baz];`,
				}},
			}},
		},

		// Task 1.8: Nested empty object patterns
		// Both patterns are detected and can be fixed iteratively
		{
			Code: `((foo1 || {}).foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(foo1 || {}).foo2?.foo3;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(foo1?.foo2 || {}).foo3;`,
					}},
				},
			},
		},

		// Task 1.9: Multiple alternates
		{
			Code: `(foo || undefined || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo || undefined)?.bar;`,
				}},
			}},
		},

		// Task 1.10: Chained calls
		{
			Code: `(foo() || bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo() || bar)?.baz;`,
				}},
			}},
		},

		// Task 1.11: Ternary expression
		{
			Code: `((foo1 ? foo2 : foo3) || {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo1 ? foo2 : foo3)?.foo4;`,
				}},
			}},
		},

		// Task 1.12: Binary operators
		{
			Code: `(a > b || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a > b)?.bar;`,
				}},
			}},
		},
		{
			Code: `(a instanceof Error || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a instanceof Error)?.bar;`,
				}},
			}},
		},

		// Task 1.13: Shift operators
		{
			Code: `((a << b) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a << b)?.bar;`,
				}},
			}},
		},

		// Task 1.14: Exponentiation
		{
			Code: `((foo ** 2) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo ** 2)?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo ** 2 || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo ** 2)?.bar;`,
				}},
			}},
		},

		// Task 1.15: Unary operators
		{
			Code: `(foo++ || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo++)?.bar;`,
				}},
			}},
		},
		{
			Code: `(+foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(+foo)?.bar;`,
				}},
			}},
		},

		// Task 1.16: this keyword
		{
			Code: `(this || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `this?.foo;`,
				}},
			}},
		},

		// Task 1.17: Type cast
		{
			Code: `(((typeof x) as string) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `((typeof x) as string)?.bar;`,
				}},
			}},
		},

		// Task 1.18: Void operator
		{
			Code: `(void foo() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(void foo())?.bar;`,
				}},
			}},
		},

		// Task 1.19: New expression
		{
			Code: `(new Foo() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `new Foo()?.bar;`,
				}},
			}},
		},

		// Task 1.20: New expression with arguments
		{
			Code: `(new Foo(arg) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `new Foo(arg)?.bar;`,
				}},
			}},
		},

		// Task 1.21: Sequence expression
		{
			Code: `((foo, bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo, bar)?.baz;`,
				}},
			}},
		},

		// Task 1.22: Class expression
		{
			Code: `((class {}) || {}).name;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(class {})?.name;`,
				}},
			}},
		},

		// Task 1.23: Optional chaining on left
		{
			Code: `(foo?.() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.()?.bar;`,
				}},
			}},
		},

		// Task 1.24: Delete operator
		{
			Code: `(delete foo.bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(delete foo.bar)?.baz;`,
				}},
			}},
		},

		// Task 1.25: In operator
		{
			Code: `(('foo' in bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `('foo' in bar)?.baz;`,
				}},
			}},
		},

		// Task 1.26: Multiple binary operators (addition and subtraction)
		{
			Code: `(a + b - c || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a + b - c)?.foo;`,
				}},
			}},
		},

		// Task 1.27: Bitwise OR operator
		{
			Code: `(a | b || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a | b)?.foo;`,
				}},
			}},
		},

		// Task 1.28: Bitwise AND operator
		{
			Code: `(a & b || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a & b)?.foo;`,
				}},
			}},
		},

		// Task 1.29: Bitwise XOR operator
		{
			Code: `(a ^ b || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a ^ b)?.foo;`,
				}},
			}},
		},

		// Task 1.30: Regex pattern
		{
			Code: `(/test/ || {}).source;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `/test/?.source;`,
				}},
			}},
		},

		// Task 1.31: Tagged template
		{
			Code: "(tag`template` || {}).foo;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    "tag`template`?.foo;",
				}},
			}},
		},

		// Task 1.32: Deeply nested property chain
		{
			Code: `((foo1.foo2.foo3.foo4) || {}).foo5;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo1.foo2.foo3.foo4)?.foo5;`,
				}},
			}},
		},

		// Task 1.33: Mixed operators - logical and arithmetic
		{
			Code: `(foo && bar + baz || {}).qux;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo && bar + baz)?.qux;`,
				}},
			}},
		},

		// Task 1.34: Array literal access
		{
			Code: `([foo, bar] || {}).length;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `[foo, bar]?.length;`,
				}},
			}},
		},

		// Task 1.35: Object literal with spread
		{
			Code: `({...foo, bar: 1} || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `{...foo, bar: 1}?.baz;`,
				}},
			}},
		},

		// Task 1.36: Function expression
		{
			Code: `(function() { return foo; } || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `function() { return foo; }?.bar;`,
				}},
			}},
		},

		// Task 1.37: Multiple levels of nesting
		{
			Code: `(((foo || {}).bar || {}).baz || {}).qux;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo || {}).bar || {}).baz?.qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo || {}).bar?.baz || {}).qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo?.bar || {}).baz || {}).qux;`,
					}},
				},
			},
		},

		// Task 1.38: Logical NOT with comparison
		{
			Code: `(!(foo === bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(!(foo === bar))?.baz;`,
				}},
			}},
		},

		// ============================================================
		// PHASE 1: Empty Object Pattern with ?? Operator (Task 2.1-2.3)
		// ============================================================

		// Task 2.1: Basic nullish coalescing
		{
			Code: `(foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo ?? ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.bar;`,
				}},
			}},
		},

		// Task 2.2: All || patterns with ?? operator
		{
			Code: `(await foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(await foo)?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo1?.foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo1?.foo2?.foo3;`,
				}},
			}},
		},
		{
			Code: `((() => foo())() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(() => foo())()?.bar;`,
				}},
			}},
		},
		{
			Code: `const foo = (bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `const foo = bar?.baz;`,
				}},
			}},
		},
		{
			Code: `(foo.bar ?? {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo.bar?.[baz];`,
				}},
			}},
		},
		{
			Code: `((foo1 ?? {}).foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(foo1 ?? {}).foo2?.foo3;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(foo1?.foo2 ?? {}).foo3;`,
					}},
				},
			},
		},
		{
			Code: `(foo ?? undefined ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo ?? undefined)?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo() ?? bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo() ?? bar)?.baz;`,
				}},
			}},
		},
		{
			Code: `((foo1 ? foo2 : foo3) ?? {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo1 ? foo2 : foo3)?.foo4;`,
				}},
			}},
		},
		{
			Code: `(a > b ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a > b)?.bar;`,
				}},
			}},
		},
		{
			Code: `((a << b) ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a << b)?.bar;`,
				}},
			}},
		},
		{
			Code: `((foo ** 2) ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo ** 2)?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo++ ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo++)?.bar;`,
				}},
			}},
		},
		{
			Code: `(+foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(+foo)?.bar;`,
				}},
			}},
		},
		{
			Code: `(void foo() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(void foo())?.bar;`,
				}},
			}},
		},

		// Task 2.3: New patterns with ?? operator
		{
			Code: `(new Foo() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `new Foo()?.bar;`,
				}},
			}},
		},
		{
			Code: `((foo, bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo, bar)?.baz;`,
				}},
			}},
		},
		{
			Code: `((class {}) ?? {}).name;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(class {})?.name;`,
				}},
			}},
		},
		{
			Code: `(foo?.() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.()?.bar;`,
				}},
			}},
		},
		{
			Code: `(delete foo.bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(delete foo.bar)?.baz;`,
				}},
			}},
		},
		{
			Code: `(('foo' in bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `('foo' in bar)?.baz;`,
				}},
			}},
		},
		{
			Code: `(a + b - c ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a + b - c)?.foo;`,
				}},
			}},
		},
		{
			Code: `(a | b ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a | b)?.foo;`,
				}},
			}},
		},
		{
			Code: `(a & b ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a & b)?.foo;`,
				}},
			}},
		},
		{
			Code: `(/test/ ?? {}).source;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `/test/?.source;`,
				}},
			}},
		},
		{
			Code: "(tag`template` ?? {}).foo;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    "tag`template`?.foo;",
				}},
			}},
		},
		{
			Code: `([foo, bar] ?? {}).length;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `[foo, bar]?.length;`,
				}},
			}},
		},
		{
			Code: `({...foo, bar: 1} ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `{...foo, bar: 1}?.baz;`,
				}},
			}},
		},
		{
			Code: `(function() { return foo; } ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `function() { return foo; }?.bar;`,
				}},
			}},
		},
		{
			Code: `(!(foo === bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(!(foo === bar))?.baz;`,
				}},
			}},
		},

		// ============================================================
		// PHASE 1: Comparison Endings - Basic Set (Task 3.1-3.6)
		// ============================================================

		// Task 3.1: == comparisons
		{
			Code:   `foo && foo.bar == 1;`,
			Output: []string{`foo?.bar == 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar == '123';`,
			Output: []string{`foo?.bar == '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar == {};`,
			Output: []string{`foo?.bar == {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar == false;`,
			Output: []string{`foo?.bar == false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar == true;`,
			Output: []string{`foo?.bar == true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Task 3.2: === comparisons
		{
			Code:   `foo && foo.bar === 0;`,
			Output: []string{`foo?.bar === 0;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar === 1;`,
			Output: []string{`foo?.bar === 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar === '123';`,
			Output: []string{`foo?.bar === '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar === {};`,
			Output: []string{`foo?.bar === {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar === false;`,
			Output: []string{`foo?.bar === false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar === true;`,
			Output: []string{`foo?.bar === true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Task 3.3: != comparisons
		{
			Code:   `foo && foo.bar != undefined;`,
			Output: []string{`foo?.bar != undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Task 3.4: !== comparisons
		{
			Code:   `foo && foo.bar !== undefined;`,
			Output: []string{`foo?.bar !== undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar !== null;`,
			Output: []string{`foo?.bar !== null;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Task 3.5: With != null prefix
		{
			Code:   `foo != null && foo.bar == 1;`,
			Output: []string{`foo?.bar == 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar === 1;`,
			Output: []string{`foo?.bar === 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar === '123';`,
			Output: []string{`foo?.bar === '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar === {};`,
			Output: []string{`foo?.bar === {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar === false;`,
			Output: []string{`foo?.bar === false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar === true;`,
			Output: []string{`foo?.bar === true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar !== null;`,
			Output: []string{`foo?.bar !== null;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar !== undefined;`,
			Output: []string{`foo?.bar !== undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Task 3.6: With typed declarations
		{
			Code: `
				declare const foo: { bar: number };
				foo && foo.bar != null;
			`,
			Output: []string{`
				declare const foo: { bar: number };
				foo?.bar != null;
			`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// PHASE 3: Category 6.1 - Additional || Patterns (20 tests)
		// ============================================================

		// Optional chaining on left side
		{
			Code: `(foo?.() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.()?.bar;`,
				}},
			}},
		},

		// Negation operator
		{
			Code: `(!foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(!foo)?.bar;`,
				}},
			}},
		},

		// Delete operator
		{
			Code: `(delete foo.bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(delete foo.bar)?.baz;`,
				}},
			}},
		},

		// In operator
		{
			Code: `(('foo' in bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `('foo' in bar)?.baz;`,
				}},
			}},
		},

		// Bitwise OR operator
		{
			Code: `((a | b) || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a | b)?.foo;`,
				}},
			}},
		},

		// Bitwise AND operator
		{
			Code: `((a & b) || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a & b)?.foo;`,
				}},
			}},
		},

		// Bitwise XOR operator
		{
			Code: `((a ^ b) || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a ^ b)?.foo;`,
				}},
			}},
		},

		// Multiple binary operators
		{
			Code: `((a + b - c) || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a + b - c)?.foo;`,
				}},
			}},
		},

		// Tilde (bitwise NOT) operator
		{
			Code: `((~foo) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(~foo)?.bar;`,
				}},
			}},
		},

		// Decrement operator
		{
			Code: `(foo-- || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo--)?.bar;`,
				}},
			}},
		},

		// Pre-decrement operator
		{
			Code: `(--foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(--foo)?.bar;`,
				}},
			}},
		},

		// Multiple chained empty object patterns
		// All 3 patterns detected, each suggestion fixes one pattern in original code
		{
			Code: `(((foo || {}).bar || {}).baz || {}).qux;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo || {}).bar || {}).baz?.qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo || {}).bar?.baz || {}).qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo?.bar || {}).baz || {}).qux;`,
					}},
				},
			},
		},

		// Logical AND within empty object pattern
		{
			Code: `((foo && bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo && bar)?.baz;`,
				}},
			}},
		},

		// Comma operator
		{
			Code: `((foo, bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo, bar)?.baz;`,
				}},
			}},
		},

		// Nullish coalescing on left side
		{
			Code: `((foo ?? bar) || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo ?? bar)?.baz;`,
				}},
			}},
		},

		// Comparison operators
		{
			Code: `((foo > 0) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo > 0)?.bar;`,
				}},
			}},
		},
		{
			Code: `((foo < 10) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo < 10)?.bar;`,
				}},
			}},
		},
		{
			Code: `((foo >= 0) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo >= 0)?.bar;`,
				}},
			}},
		},
		{
			Code: `((foo <= 10) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo <= 10)?.bar;`,
				}},
			}},
		},

		// Typeof operator with empty object
		{
			Code: `((typeof foo) || {}).length;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(typeof foo)?.length;`,
				}},
			}},
		},

		// ============================================================
		// PHASE 3: Category 6.2 - Additional ?? Patterns (15 tests)
		// ============================================================

		// Optional chaining on left side with ??
		{
			Code: `(foo?.() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `foo?.()?.bar;`,
				}},
			}},
		},

		// Negation operator with ??
		{
			Code: `(!foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(!foo)?.bar;`,
				}},
			}},
		},

		// Delete operator with ??
		{
			Code: `(delete foo.bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(delete foo.bar)?.baz;`,
				}},
			}},
		},

		// Bitwise operators with ??
		{
			Code: `((a | b) ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a | b)?.foo;`,
				}},
			}},
		},
		{
			Code: `((a & b) ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a & b)?.foo;`,
				}},
			}},
		},
		{
			Code: `((a ^ b) ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a ^ b)?.foo;`,
				}},
			}},
		},

		// Multiple binary operators with ??
		{
			Code: `((a + b - c) ?? {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(a + b - c)?.foo;`,
				}},
			}},
		},

		// Unary operators with ??
		{
			Code: `((~foo) ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(~foo)?.bar;`,
				}},
			}},
		},
		{
			Code: `(foo-- ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo--)?.bar;`,
				}},
			}},
		},

		// Multiple chained ?? empty object patterns
		{
			Code: `(((foo ?? {}).bar ?? {}).baz ?? {}).qux;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo ?? {}).bar ?? {}).baz?.qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo ?? {}).bar?.baz ?? {}).qux;`,
					}},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `((foo?.bar ?? {}).baz ?? {}).qux;`,
					}},
				},
			},
		},

		// Logical AND within ?? empty object pattern
		{
			Code: `((foo && bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo && bar)?.baz;`,
				}},
			}},
		},

		// Comma operator with ??
		{
			Code: `((foo, bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo, bar)?.baz;`,
				}},
			}},
		},

		// Mixed ?? and || (|| within left side)
		{
			Code: `((foo || bar) ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo || bar)?.baz;`,
				}},
			}},
		},

		// Comparison operators with ??
		{
			Code: `((foo > 0) ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(foo > 0)?.bar;`,
				}},
			}},
		},

		// Typeof operator with ?? and empty object
		{
			Code: `((typeof foo) ?? {}).length;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `(typeof foo)?.length;`,
				}},
			}},
		},
	})
}

func TestPreferOptionalChainPhase1Valid(t *testing.T) {
	// Task 1.19-1.20 & 2.3: Valid cases that should NOT convert
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Task 1.19: Valid empty object cases
		{Code: `foo || {};`},                       // No property access
		{Code: `(foo || {})?.bar;`},                // Already optional
		{Code: `(foo || { bar: 1 }).bar;`},         // Non-empty object
		{Code: `foo ||= bar || {};`},               // Assignment operator
		{Code: `(foo1 ? foo2 : foo3 || {}).foo4;`}, // Ternary in wrong position
		{Code: `func(foo || {}).bar;`},             // Function call context (may be valid in some contexts)

		// Task 1.20: Edge valid cases
		{Code: `(undefined && (foo || {})).bar;`}, // Complex condition

		// Task 2.3: Valid nullish coalescing cases
		{Code: `foo ?? {};`},         // No property access
		{Code: `(foo ?? {})?.bar;`},  // Already optional
		{Code: `foo ||= bar ?? {};`}, // Assignment with ??
	}, []rule_tester.InvalidTestCase{})
}
