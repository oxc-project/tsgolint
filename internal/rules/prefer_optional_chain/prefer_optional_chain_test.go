package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// =============================================================================
// SECTION 1: || {} Tests (Empty Object Pattern)
// Source: upstream prefer-optional-chain.test.ts lines 10-693
// =============================================================================

func TestOrEmptyObjectPattern(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - should NOT be converted
		[]rule_tester.ValidTestCase{
			{Code: `foo || {};`},                       // No property access
			{Code: `foo || ({} as any);`},              // Type cast empty object
			{Code: `(foo || {})?.bar;`},                // Already optional
			{Code: `(foo || { bar: 1 }).bar;`},         // Non-empty object
			{Code: `foo ||= bar || {};`},               // Assignment operator
			{Code: `foo ||= bar?.baz || {};`},          // Assignment with optional chain
			{Code: `(foo1 ? foo2 : foo3 || {}).foo4;`}, // Ternary in wrong position
			{Code: `(foo = 2 || {}).bar;`},             // Assignment expression
			{Code: `func(foo || {}).bar;`},             // Function call result
			{Code: `(undefined && (foo || {})).bar;`},  // Complex condition
			// https://github.com/typescript-eslint/typescript-eslint/issues/8380
			{Code: `
const a = null;
const b = 0;
a === undefined || b === null || b === undefined;
`},
			{Code: `
const a = 0;
const b = 0;
a === undefined || b === undefined || b === null;
`},
			{Code: `
const a = 0;
const b = 0;
b === null || a === undefined || b === undefined;
`},
			{Code: `
const b = 0;
b === null || b === undefined;
`},
			{Code: `
const a = 0;
const b = 0;
b != null && a !== null && a !== undefined;
`},
		},
		// Invalid cases - should be converted
		[]rule_tester.InvalidTestCase{
			// Basic || {} pattern
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
			// Parenthesized empty object
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
			// Await with empty object
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
			// Nested optional chain
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
			// Arrow function call
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
			// Const assignment
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
			// Computed property
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
			// Nested empty object patterns (multiple errors)
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
			// Multiple alternates
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
			// Chained calls
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
			// Ternary expression
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
			// Binary operators
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
			// Shift operators
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
			// Exponentiation
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
			// Unary operators
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
			// this keyword
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
			// Type cast
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
			// Void operator
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
			// New expression
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
			// Sequence expression
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
			// Class expression
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
			// Optional chaining on left
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
			// Multiple binary operators
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
			// Bitwise operators
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
			// Regex pattern
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
			// Tagged template
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
			// Deeply nested property chain
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
			// Mixed operators
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
			// Array literal access
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
			// Object literal with spread
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
			// Function expression
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
			// Multiple levels of nesting (3 errors)
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
			// Logical NOT with comparison
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
			// Tilde (bitwise NOT)
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
			// Decrement operators
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
			// Logical AND within empty object
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
			// Typeof operator
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
			// If-block context tests
			{
				Code: `if (foo) { (foo || {}).bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `if (foo) { foo?.bar; }`,
					}},
				}},
			},
			{
				Code: `if ((foo || {}).bar) { foo.bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `if (foo?.bar) { foo.bar; }`,
					}},
				}},
			},
			// Ternary expression
			{
				Code: `((a ? b : c) || {}).bar;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(a ? b : c)?.bar;`,
					}},
				}},
			},
		})
}

// =============================================================================
// SECTION 1b: ?? {} Tests (Nullish Coalescing Empty Object Pattern)
// =============================================================================

func TestNullishCoalescingEmptyObjectPattern(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `foo ?? {};`},         // No property access
			{Code: `(foo ?? {})?.bar;`},  // Already optional
			{Code: `foo ||= bar ?? {};`}, // Assignment operator with ??
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Basic ?? {} pattern
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
			// If-block context tests
			{
				Code: `if (foo) { (foo ?? {}).bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `if (foo) { foo?.bar; }`,
					}},
				}},
			},
			{
				Code: `if ((foo ?? {}).bar) { foo.bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `if (foo?.bar) { foo.bar; }`,
					}},
				}},
			},
			// undefined && foo ?? {} pattern
			{
				Code: `(undefined && foo ?? {}).bar;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `(undefined && foo)?.bar;`,
					}},
				}},
			},
		})
}

// =============================================================================
// SECTION 2: Chain Ending with Comparison Tests
// Source: upstream prefer-optional-chain.test.ts lines 694-1884
// =============================================================================

func TestChainEndingWithComparison(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - should NOT be converted
		[]rule_tester.ValidTestCase{
			// Element access with comparison - special handling required
			{Code: `declare const record: Record<string, { kind: string }>; record['key'] && record['key'].kind !== '1';`},
			{Code: `declare const array: { b?: string }[]; !array[1] || array[1].b === 'foo';`},

			// foo && foo.bar with undeclared/null/undefined comparisons (valid - shouldn't convert)
			{Code: `foo && foo.bar == undeclaredVar;`},
			{Code: `foo && foo.bar == null;`},
			{Code: `foo && foo.bar == undefined;`},
			{Code: `foo && foo.bar === undeclaredVar;`},
			{Code: `foo && foo.bar === undefined;`},
			{Code: `foo && foo.bar !== 0;`},
			{Code: `foo && foo.bar !== 1;`},
			{Code: `foo && foo.bar !== '123';`},
			{Code: `foo && foo.bar !== {};`},
			{Code: `foo && foo.bar !== false;`},
			{Code: `foo && foo.bar !== true;`},
			{Code: `foo && foo.bar !== null;`},
			{Code: `foo && foo.bar !== undeclaredVar;`},
			{Code: `foo && foo.bar != 0;`},
			{Code: `foo && foo.bar != 1;`},
			{Code: `foo && foo.bar != '123';`},
			{Code: `foo && foo.bar != {};`},
			{Code: `foo && foo.bar != false;`},
			{Code: `foo && foo.bar != true;`},
			{Code: `foo && foo.bar != undeclaredVar;`},

			// foo != null && foo.bar patterns
			{Code: `foo != null && foo.bar == undeclaredVar;`},
			{Code: `foo != null && foo.bar == null;`},
			{Code: `foo != null && foo.bar == undefined;`},
			{Code: `foo != null && foo.bar === undeclaredVar;`},
			{Code: `foo != null && foo.bar === undefined;`},
			{Code: `foo != null && foo.bar !== 0;`},
			{Code: `foo != null && foo.bar !== 1;`},
			{Code: `foo != null && foo.bar !== '123';`},
			{Code: `foo != null && foo.bar !== {};`},
			{Code: `foo != null && foo.bar !== false;`},
			{Code: `foo != null && foo.bar !== true;`},
			{Code: `foo != null && foo.bar !== null;`},
			{Code: `foo != null && foo.bar !== undeclaredVar;`},
			{Code: `foo != null && foo.bar != 0;`},
			{Code: `foo != null && foo.bar != 1;`},
			{Code: `foo != null && foo.bar != '123';`},
			{Code: `foo != null && foo.bar != {};`},
			{Code: `foo != null && foo.bar != false;`},
			{Code: `foo != null && foo.bar != true;`},
			{Code: `foo != null && foo.bar != undeclaredVar;`},

			// !foo && patterns with comparisons - inverted checks don't convert
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

			// foo == null && patterns - inverted nullish checks don't convert
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

			// Falsy union valid cases (false |, '' |, 0 |, 0n |)
			{Code: `declare const foo: false | { a: string }; foo && foo.a == undeclaredVar;`},
			{Code: `declare const foo: '' | { a: string }; foo && foo.a == undeclaredVar;`},
			{Code: `declare const foo: 0 | { a: string }; foo && foo.a == undeclaredVar;`},
			{Code: `declare const foo: 0n | { a: string }; foo && foo.a;`},

			// Type declaration with | null patterns
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar == undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar == null;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar == undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar === undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== 0;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== '123';`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== {};`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== false;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== true;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== null;`},
			{Code: `declare const foo: { bar: number } | null; foo && foo.bar !== undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar == undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar == null;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar == undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar === undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== 0;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== '123';`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== {};`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== false;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== true;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== null;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && foo.bar !== undeclaredVar;`},
			{Code: `declare const foo: { bar: number } | null; foo !== null && foo !== undefined && foo.bar == null;`},
			{Code: `declare const foo: { bar: number } | null; foo !== null && foo !== undefined && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo !== null && foo !== undefined && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number } | null; foo !== null && foo !== undefined && foo.bar != 1;`},

			// Type declaration with | undefined patterns
			{Code: `declare const foo: { bar: number } | undefined; foo !== null && foo !== undefined && foo.bar == null;`},
			{Code: `declare const foo: { bar: number } | undefined; foo !== null && foo !== undefined && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number } | undefined; foo !== null && foo !== undefined && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number } | undefined; foo !== null && foo !== undefined && foo.bar != 1;`},

			// Type declaration with foo !== undefined patterns
			{Code: `declare const foo: { bar: number } | null; foo !== undefined && foo !== undefined && foo.bar == null;`},
			{Code: `declare const foo: { bar: number } | null; foo !== undefined && foo !== undefined && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number } | null; foo !== undefined && foo !== undefined && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number } | null; foo !== undefined && foo !== undefined && foo.bar != 1;`},

			// =============================================================================
			// Yoda case tests (null != foo.bar, 0 === foo.bar, etc.)
			// =============================================================================

			// Yoda style: value on left side of comparison - foo && patterns
			{Code: `foo && undeclaredVar == foo.bar;`},
			{Code: `foo && null == foo.bar;`},
			{Code: `foo && undefined == foo.bar;`},
			{Code: `foo && undeclaredVar === foo.bar;`},
			{Code: `foo && undefined === foo.bar;`},
			{Code: `foo && 0 !== foo.bar;`},
			{Code: `foo && 1 !== foo.bar;`},
			{Code: `foo && '123' !== foo.bar;`},
			{Code: `foo && false !== foo.bar;`},
			{Code: `foo && true !== foo.bar;`},
			{Code: `foo && null !== foo.bar;`},
			{Code: `foo && undeclaredVar !== foo.bar;`},
			{Code: `foo && 0 != foo.bar;`},
			{Code: `foo && 1 != foo.bar;`},
			{Code: `foo && '123' != foo.bar;`},
			{Code: `foo && false != foo.bar;`},
			{Code: `foo && true != foo.bar;`},
			{Code: `foo && undeclaredVar != foo.bar;`},

			// Yoda style: foo != null && patterns
			{Code: `foo != null && undeclaredVar == foo.bar;`},
			{Code: `foo != null && null == foo.bar;`},
			{Code: `foo != null && undefined == foo.bar;`},
			{Code: `foo != null && undeclaredVar === foo.bar;`},
			{Code: `foo != null && undefined === foo.bar;`},
			{Code: `foo != null && 0 !== foo.bar;`},
			{Code: `foo != null && 1 !== foo.bar;`},
			{Code: `foo != null && '123' !== foo.bar;`},
			{Code: `foo != null && false !== foo.bar;`},
			{Code: `foo != null && true !== foo.bar;`},
			{Code: `foo != null && null !== foo.bar;`},
			{Code: `foo != null && undeclaredVar !== foo.bar;`},
			{Code: `foo != null && 0 != foo.bar;`},
			{Code: `foo != null && 1 != foo.bar;`},
			{Code: `foo != null && '123' != foo.bar;`},
			{Code: `foo != null && false != foo.bar;`},
			{Code: `foo != null && true != foo.bar;`},
			{Code: `foo != null && undeclaredVar != foo.bar;`},

			// Yoda style: null != foo && patterns (Yoda nullish check on left)
			{Code: `null != foo && undeclaredVar == foo.bar;`},
			{Code: `null != foo && null == foo.bar;`},
			{Code: `null != foo && undefined == foo.bar;`},
			{Code: `null != foo && undeclaredVar === foo.bar;`},
			{Code: `null != foo && undefined === foo.bar;`},
			{Code: `null != foo && 0 !== foo.bar;`},
			{Code: `null != foo && 1 !== foo.bar;`},
			{Code: `null != foo && '123' !== foo.bar;`},
			{Code: `null != foo && false !== foo.bar;`},
			{Code: `null != foo && true !== foo.bar;`},
			{Code: `null != foo && null !== foo.bar;`},
			{Code: `null != foo && undeclaredVar !== foo.bar;`},
			{Code: `null != foo && 0 != foo.bar;`},
			{Code: `null != foo && 1 != foo.bar;`},
			{Code: `null != foo && '123' != foo.bar;`},
			{Code: `null != foo && false != foo.bar;`},
			{Code: `null != foo && true != foo.bar;`},
			{Code: `null != foo && undeclaredVar != foo.bar;`},

			// Yoda style with type declarations
			{Code: `declare const foo: { bar: number } | null; foo && undeclaredVar == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo && null == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo && undefined == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo && 0 !== foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo && 1 !== foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo && null !== foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && undeclaredVar == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && null == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && undefined == foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && 0 !== foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && 1 !== foo.bar;`},
			{Code: `declare const foo: { bar: number } | null; foo != null && null !== foo.bar;`},

			// =============================================================================
			// Element access with comparison valid cases
			// =============================================================================

			// Element access with bracket notation - should NOT be converted when ending with comparison
			{Code: `foo && foo['bar'] == undeclaredVar;`},
			{Code: `foo && foo['bar'] == null;`},
			{Code: `foo && foo['bar'] == undefined;`},
			{Code: `foo && foo['bar'] === undeclaredVar;`},
			{Code: `foo && foo['bar'] === undefined;`},
			{Code: `foo && foo['bar'] !== 0;`},
			{Code: `foo && foo['bar'] !== null;`},
			{Code: `foo && foo['bar'] !== undeclaredVar;`},
			{Code: `foo && foo['bar'] != 0;`},
			{Code: `foo && foo['bar'] != undeclaredVar;`},

			// Element access with nullish check
			{Code: `foo != null && foo['bar'] == undeclaredVar;`},
			{Code: `foo != null && foo['bar'] == null;`},
			{Code: `foo != null && foo['bar'] == undefined;`},
			{Code: `foo != null && foo['bar'] === undeclaredVar;`},
			{Code: `foo != null && foo['bar'] === undefined;`},
			{Code: `foo != null && foo['bar'] !== 0;`},
			{Code: `foo != null && foo['bar'] !== null;`},
			{Code: `foo != null && foo['bar'] !== undeclaredVar;`},
			{Code: `foo != null && foo['bar'] != 0;`},
			{Code: `foo != null && foo['bar'] != undeclaredVar;`},

			// Computed property with variable key
			{Code: `declare const key: string; foo && foo[key] == null;`},
			{Code: `declare const key: string; foo && foo[key] == undefined;`},
			{Code: `declare const key: string; foo && foo[key] !== 0;`},
			{Code: `declare const key: string; foo != null && foo[key] == null;`},
			{Code: `declare const key: string; foo != null && foo[key] !== 0;`},

			// Array index access
			{Code: `foo && foo[0] == null;`},
			{Code: `foo && foo[0] == undefined;`},
			{Code: `foo && foo[0] !== 0;`},
			{Code: `foo && foo[0] !== null;`},
			{Code: `foo != null && foo[0] == null;`},
			{Code: `foo != null && foo[0] == undefined;`},
			{Code: `foo != null && foo[0] !== 0;`},
			{Code: `foo != null && foo[0] !== null;`},

			// Chained element access
			{Code: `foo && foo.bar && foo.bar['baz'] == null;`},
			{Code: `foo && foo.bar && foo.bar['baz'] !== 0;`},
			{Code: `foo != null && foo.bar != null && foo.bar['baz'] == null;`},
			{Code: `foo != null && foo.bar != null && foo.bar['baz'] !== 0;`},

			// Mixed property and element access
			{Code: `foo && foo['bar'].baz == null;`},
			{Code: `foo && foo['bar'].baz !== 0;`},
			{Code: `foo && foo.bar['baz'] == null;`},
			{Code: `foo && foo.bar['baz'] !== 0;`},
			{Code: `foo != null && foo['bar'].baz == null;`},
			{Code: `foo != null && foo['bar'].baz !== 0;`},
			{Code: `foo != null && foo.bar['baz'] == null;`},
			{Code: `foo != null && foo.bar['baz'] !== 0;`},

			// Yoda style element access
			{Code: `foo && undeclaredVar == foo['bar'];`},
			{Code: `foo && null == foo['bar'];`},
			{Code: `foo && 0 !== foo['bar'];`},
			{Code: `foo != null && undeclaredVar == foo['bar'];`},
			{Code: `foo != null && null == foo['bar'];`},
			{Code: `foo != null && 0 !== foo['bar'];`},

			// OR chain patterns that should NOT be converted (valid code)
			// These patterns have different semantics from optional chaining
			//
			// !foo || patterns with comparisons - unsafe because checking ALL falsy values
			{Code: `!foo || foo.bar != undeclaredVar;`},
			{Code: `!foo || foo.bar != null;`},
			{Code: `!foo || foo.bar != undefined;`},
			//
			// foo == null || patterns with comparisons involving undeclared vars
			{Code: `foo == null || foo.bar != undeclaredVar;`},
			//
			// foo || patterns (plain truthy check - should NOT be converted)
			{Code: `foo || foo.bar != 0;`},
			{Code: `foo || foo.bar != 1;`},
			{Code: `foo || foo.bar == 0;`},
		},
		// Invalid cases - should be converted
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
				Code:   "declare const foo: { bar: number };\nfoo && foo.bar != null;",
				Output: []string{"declare const foo: { bar: number };\nfoo?.bar != null;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:   "declare const foo: { bar: number };\nfoo != null && foo.bar != null;",
				Output: []string{"declare const foo: { bar: number };\nfoo?.bar != null;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// !foo || patterns (negated OR chains with comparisons)
			{Code: `!foo || foo.bar != 0;`, Output: []string{`foo?.bar != 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar != 1;`, Output: []string{`foo?.bar != 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `!foo || foo.bar !== null;`, Output: []string{`foo?.bar !== null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			// foo == null || patterns
			{Code: `foo == null || foo.bar != 0;`, Output: []string{`foo?.bar != 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// =============================================================================
			// Yoda case invalid tests (should convert, value on left side)
			// =============================================================================

			// Yoda style: foo && 0 == foo.bar (should convert to foo?.bar == 0, preserving Yoda style)
			{Code: `foo && 0 == foo.bar;`, Output: []string{`0 == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && 1 == foo.bar;`, Output: []string{`1 == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && '123' == foo.bar;`, Output: []string{`'123' == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && false == foo.bar;`, Output: []string{`false == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && true == foo.bar;`, Output: []string{`true == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && 0 === foo.bar;`, Output: []string{`0 === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && 1 === foo.bar;`, Output: []string{`1 === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && '123' === foo.bar;`, Output: []string{`'123' === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && false === foo.bar;`, Output: []string{`false === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && true === foo.bar;`, Output: []string{`true === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && null === foo.bar;`, Output: []string{`null === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && undefined !== foo.bar;`, Output: []string{`undefined !== foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && undefined != foo.bar;`, Output: []string{`undefined != foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && null != foo.bar;`, Output: []string{`null != foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Yoda style: foo != null && 0 == foo.bar
			{Code: `foo != null && 0 == foo.bar;`, Output: []string{`0 == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && 1 == foo.bar;`, Output: []string{`1 == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && '123' == foo.bar;`, Output: []string{`'123' == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && false == foo.bar;`, Output: []string{`false == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && true == foo.bar;`, Output: []string{`true == foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && 0 === foo.bar;`, Output: []string{`0 === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && 1 === foo.bar;`, Output: []string{`1 === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && '123' === foo.bar;`, Output: []string{`'123' === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && false === foo.bar;`, Output: []string{`false === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && true === foo.bar;`, Output: []string{`true === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && null === foo.bar;`, Output: []string{`null === foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && undefined !== foo.bar;`, Output: []string{`undefined !== foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && undefined != foo.bar;`, Output: []string{`undefined != foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && null != foo.bar;`, Output: []string{`null != foo?.bar;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// =============================================================================
			// Element access invalid tests (should convert)
			// =============================================================================

			// Element access with bracket notation - should convert
			{Code: `foo && foo['bar'] == 0;`, Output: []string{`foo?.['bar'] == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo['bar'] == 1;`, Output: []string{`foo?.['bar'] == 1;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo['bar'] === false;`, Output: []string{`foo?.['bar'] === false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo['bar'] === true;`, Output: []string{`foo?.['bar'] === true;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo['bar'] !== undefined;`, Output: []string{`foo?.['bar'] !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo['bar'] != null;`, Output: []string{`foo?.['bar'] != null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Element access with nullish check
			{Code: `foo != null && foo['bar'] == 0;`, Output: []string{`foo?.['bar'] == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo['bar'] === false;`, Output: []string{`foo?.['bar'] === false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo['bar'] !== undefined;`, Output: []string{`foo?.['bar'] !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo['bar'] != null;`, Output: []string{`foo?.['bar'] != null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Array index access - should convert
			{Code: `foo && foo[0] == 0;`, Output: []string{`foo?.[0] == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo[0] === false;`, Output: []string{`foo?.[0] === false;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo[0] !== undefined;`, Output: []string{`foo?.[0] !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && foo[0] != null;`, Output: []string{`foo?.[0] != null;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo[0] == 0;`, Output: []string{`foo?.[0] == 0;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && foo[0] !== undefined;`, Output: []string{`foo?.[0] !== undefined;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},

			// Yoda style element access - should convert
			{Code: `foo && 0 == foo['bar'];`, Output: []string{`0 == foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && false === foo['bar'];`, Output: []string{`false === foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && undefined !== foo['bar'];`, Output: []string{`undefined !== foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo && null != foo['bar'];`, Output: []string{`null != foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && 0 == foo['bar'];`, Output: []string{`0 == foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
			{Code: `foo != null && undefined !== foo['bar'];`, Output: []string{`undefined !== foo?.['bar'];`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}},
		})
}

// =============================================================================
// SECTION 3: Hand-Crafted Cases (Complex Real-World Patterns)
// Source: upstream prefer-optional-chain.test.ts lines 1885-3112
// =============================================================================

func TestHandCraftedCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Already using optional chain
			{Code: `foo?.bar;`},
			{Code: `foo?.bar?.baz;`},
			{Code: `foo?.bar?.();`},
			// Simple cases that don't benefit
			{Code: `foo && bar;`},
			{Code: `foo || bar;`},
			{Code: `foo && bar || baz;`},
			// Nullish checks without property access
			{Code: `foo != null && bar;`},
			{Code: `foo !== undefined && bar;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Basic && chain
			{
				Code:    `foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `foo && foo.bar && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz;`,
				Output:  []string{`foo?.bar?.baz?.buzz;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Computed property access
			{
				Code:    `foo && foo['bar'] && foo['bar'].baz;`,
				Output:  []string{`foo?.['bar']?.baz;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Method calls
			{
				Code:    `foo && foo.bar && foo.bar.baz && foo.bar.baz();`,
				Output:  []string{`foo?.bar?.baz?.();`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Nullish comparisons
			{
				Code:    `foo != null && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `foo !== null && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `foo !== undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Combined null and undefined checks
			{
				Code:    `foo !== null && foo !== undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Yoda conditions
			{
				Code:    `null !== foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `undefined !== foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Typeof check
			{
				Code:    `typeof foo !== 'undefined' && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Multi-level with nullish check
			{
				Code:    `foo != null && foo.bar && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Negated OR chain
			{
				Code:    `!foo || !foo.bar;`,
				Output:  []string{`!foo?.bar;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			{
				Code:    `!foo || !foo.bar || !foo.bar.baz;`,
				Output:  []string{`!foo?.bar?.baz;`},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Multiple chains in one expression (two errors)
			{
				Code:   `foo && foo.bar && foo.bar.baz || baz && baz.bar && baz.bar.foo`,
				Output: []string{`foo?.bar?.baz || baz?.bar?.foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Inconsistent checks break the chain
			{
				Code:    `foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;`,
				Output:  []string{`foo?.bar?.baz !== undefined && foo.bar.baz.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// String literal element access
			{
				Code:    `foo && foo['some long string'] && foo['some long string'].baz;`,
				Output:  []string{`foo?.['some long string']?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "foo && foo[`some long string`] && foo[`some long string`].baz;",
				Output:  []string{"foo?.[`some long string`]?.baz;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "foo && foo[`some ${long} string`] && foo[`some ${long} string`].baz;",
				Output:  []string{"foo?.[`some ${long} string`]?.baz;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Complex computed properties
			{
				Code:    `foo && foo[bar as string] && foo[bar as string].baz;`,
				Output:  []string{`foo?.[bar as string]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[1 + 2] && foo[1 + 2].baz;`,
				Output:  []string{`foo?.[1 + 2]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[typeof bar] && foo[typeof bar].baz;`,
				Output:  []string{`foo?.[typeof bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Call expressions
			{
				Code:    `foo() && foo()(bar);`,
				Output:  []string{`foo()?.(bar);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Type parameters
			{
				Code:    `foo && foo<string>() && foo<string>().bar;`,
				Output:  []string{`foo?.<string>()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Await expressions
			{
				Code:    `(await foo).bar && (await foo).bar.baz;`,
				Output:  []string{`(await foo).bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// import.meta
			{
				Code:    `import.meta && import.meta.baz;`,
				Output:  []string{`import.meta?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// new.target
			{
				Code:    "class Foo { constructor() { new.target && new.target.length; } }",
				Output:  []string{"class Foo { constructor() { new.target?.length; } }"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Complex chain with method call
			{
				Code:    `foo !== null && foo.bar !== undefined && foo.bar.baz();`,
				Output:  []string{`foo?.bar?.baz();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Nested property with computed access
			{
				Code:    `foo != null && foo.bar && foo.bar['baz'];`,
				Output:  []string{`foo?.bar?.['baz'];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

// =============================================================================
// SECTION 3b: This Keyword, Non-null Assertions, Mixed Optional Chain Tests
// Source: upstream prefer-optional-chain.test.ts
// =============================================================================

func TestThisKeywordAccess(t *testing.T) {
	// Tests for this.bar patterns
	// Note: `this` is always truthy in class methods, so the rule treats
	// `this && this.bar` as valid (no conversion suggested for simple truthiness check)
	// But `this != null && this.bar` IS converted since it's a nullish check pattern
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - patterns that should NOT be converted
		[]rule_tester.ValidTestCase{
			// Simple this expressions without chain patterns
			{Code: `class Foo { method() { return this; } }`},
			{Code: `class Foo { method() { return this.bar; } }`},
			// this in different contexts - different objects
			{Code: `class Foo { method() { foo && this.bar; } }`},
			// this && this.bar - truthiness check on `this` (always truthy)
			{Code: `class Foo { method() { this && this.bar; } }`},
			{Code: `class Foo { method() { this && this['bar']; } }`},
			{Code: `class Foo { method() { this && this.bar(); } }`},
		},
		// Invalid cases - patterns the rule DOES flag
		[]rule_tester.InvalidTestCase{
			// this with nullish check converts
			{
				Code:    `class Foo { method() { this != null && this.bar; } }`,
				Output:  []string{`class Foo { method() { this?.bar; } }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestNonNullAssertion(t *testing.T) {
	// Tests for non-null assertion (foo!.bar) patterns
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - should NOT be converted
		[]rule_tester.ValidTestCase{
			// Already using non-null assertion - don't suggest optional chain
			{Code: `foo!.bar;`},
			{Code: `foo!.bar!.baz;`},
			{Code: `foo!['bar'];`},
			{Code: `foo![0];`},
			{Code: `foo!.bar();`},
			// Non-null assertion on the base with chained access - no conversion suggested
			{Code: `foo! && foo!.bar;`},
			// Non-null on intermediate access only (no chain)
			{Code: `foo && foo.bar!;`},
		},
		// Invalid cases - patterns that convert even with non-null assertions
		[]rule_tester.InvalidTestCase{
			// Regular chain followed by non-null on result
			{
				Code:    `foo && foo.bar && foo.bar.baz!;`,
				Output:  []string{`foo?.bar?.baz!;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Non-null on base, but chain pattern still applies
			{
				Code:    `foo && foo!.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Non-null chain pattern
			{
				Code:    `foo!.bar && foo!.bar.baz;`,
				Output:  []string{`foo!.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Non-null on intermediate access with continuation
			{
				Code:    `foo && foo.bar!.baz;`,
				Output:  []string{`foo?.bar!?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestMixedOptionalChainTokens(t *testing.T) {
	// Tests for patterns mixing optional chain (?.) with regular access
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - already using optional chain correctly
		[]rule_tester.ValidTestCase{
			// Already fully optional
			{Code: `foo?.bar;`},
			{Code: `foo?.bar?.baz;`},
			{Code: `foo?.bar?.baz?.buzz;`},
			{Code: `foo?.[0];`},
			{Code: `foo?.['bar'];`},
			{Code: `foo?.bar?.();`},
			// Optional chain followed by regular access (valid, chain is already started)
			{Code: `foo?.bar.baz;`},
			{Code: `foo?.bar.baz.buzz;`},
			{Code: `foo?.bar.baz?.buzz;`},
			// Partially optional chain patterns - already has optional chain, no additional conversion
			{Code: `foo?.bar && foo.bar.baz;`},
			// When foo.bar is already optional, foo && foo.bar?.baz doesn't get flagged
			// because the rule sees the optional chain already present
			{Code: `foo && foo.bar?.baz;`},
		},
		// Invalid cases - patterns that should be converted
		[]rule_tester.InvalidTestCase{
			// Long chain with optional in middle - should combine
			{
				Code:    `foo && foo.bar && foo.bar?.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestLongMixedBinaryChains(t *testing.T) {
	// Tests for long chains (10+ parts) with mixed patterns
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Long chain but different base objects - shouldn't combine
			{Code: `a && a.b && b && b.c && c && c.d && d && d.e;`},
			// Long chain with breaks
			{Code: `a && a.b && c && c.d && a && a.e;`},
		},
		// Invalid cases - long chains that should be converted
		[]rule_tester.InvalidTestCase{
			// 5-part chain
			{
				Code:    `foo && foo.a && foo.a.b && foo.a.b.c && foo.a.b.c.d;`,
				Output:  []string{`foo?.a?.b?.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// 6-part chain
			{
				Code:    `foo && foo.a && foo.a.b && foo.a.b.c && foo.a.b.c.d && foo.a.b.c.d.e;`,
				Output:  []string{`foo?.a?.b?.c?.d?.e;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// 7-part chain with nullish checks
			{
				Code:    `foo != null && foo.a != null && foo.a.b && foo.a.b.c && foo.a.b.c.d && foo.a.b.c.d.e && foo.a.b.c.d.e.f;`,
				Output:  []string{`foo?.a?.b?.c?.d?.e?.f;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Mixed property and element access long chain
			{
				Code:    `foo && foo['a'] && foo['a'].b && foo['a'].b['c'] && foo['a'].b['c'].d;`,
				Output:  []string{`foo?.['a']?.b?.['c']?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Long negated OR chain
			{
				Code:    `!foo || !foo.a || !foo.a.b || !foo.a.b.c || !foo.a.b.c.d;`,
				Output:  []string{`!foo?.a?.b?.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Chain with method calls in between
			{
				Code:    `foo && foo.bar && foo.bar() && foo.bar().baz && foo.bar().baz.qux;`,
				Output:  []string{`foo?.bar?.()?.baz?.qux;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

// =============================================================================
// SECTION 3c: Private Properties, Parenthesis Grouping, Two-Error Cases
// Source: upstream prefer-optional-chain.test.ts
// =============================================================================

func TestPrivatePropertyAccess(t *testing.T) {
	// Tests for private property (#bar) patterns
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - should NOT be converted
		[]rule_tester.ValidTestCase{
			// Simple private property - already using private, not convertible in same way
			{Code: `class Foo { #bar: any; method() { this.#bar; } }`},
			{Code: `class Foo { #bar: any; #baz: any; method() { this.#bar && this.#baz; } }`},
			// Private property checks - these don't convert the same way
			{Code: `class Foo { #bar: any; method() { this && this.#bar; } }`},
			// Private property on external object
			{Code: `class Foo { method(obj: { #bar?: any }) { obj.#bar; } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Private property chain converts
			{
				Code:    `class Foo { #bar: { baz: any }; method() { this.#bar && this.#bar.baz; } }`,
				Output:  []string{`class Foo { #bar: { baz: any }; method() { this.#bar?.baz; } }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Private with nullish check
			{
				Code:    `class Foo { #bar: any; method() { this.#bar != null && this.#bar.baz; } }`,
				Output:  []string{`class Foo { #bar: any; method() { this.#bar?.baz; } }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Chain where private property is intermediate
			{
				Code:    `class Foo { #bar: { baz?: { qux: any } }; method() { this.#bar.baz && this.#bar.baz.qux; } }`,
				Output:  []string{`class Foo { #bar: { baz?: { qux: any } }; method() { this.#bar.baz?.qux; } }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestParenthesisGrouping(t *testing.T) {
	// Tests for parenthesis grouping patterns
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - parentheses change semantics or no chain pattern
		[]rule_tester.ValidTestCase{
			// Parentheses grouping with different operators - no chain
			{Code: `(a && b) || c;`},
			{Code: `a && (b || c);`},
			{Code: `(a || b) && c;`},
			{Code: `a || (b && c);`},
		},
		// Invalid cases - parentheses don't affect conversion
		[]rule_tester.InvalidTestCase{
			// Parentheses around the whole chain
			{
				Code:    `(foo && foo.bar);`,
				Output:  []string{`(foo?.bar);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Nested parentheses with inner chain
			{
				Code:    `a && (a.b && a.b.c);`,
				Output:  []string{`a?.b?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Multiple levels of grouping
			{
				Code:    `(a && (a.b && (a.b.c)));`,
				Output:  []string{`(a?.b?.c);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Parentheses around operand but same chain
			{
				Code:    `foo && (foo.bar) && (foo.bar.baz);`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Parentheses around individual parts
			{
				Code:    `a && (a).b;`,
				Output:  []string{`a?.b;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `(a) && (a).b;`,
				Output:  []string{`a?.b;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `((a)) && ((a)).b;`,
				Output:  []string{`a?.b;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Two chains in different parentheses groups
			{
				Code:   `(a && a.b) || (c && c.d);`,
				Output: []string{`(a?.b) || (c?.d);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestTwoErrorCases(t *testing.T) {
	// Tests for expressions with multiple independent chains (should report two errors)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Two completely separate expressions
			{Code: `foo?.bar; baz?.qux;`},
		},
		// Invalid cases - multiple chains in one expression
		[]rule_tester.InvalidTestCase{
			// Two chains separated by ||
			{
				Code:   `foo && foo.bar || baz && baz.qux;`,
				Output: []string{`foo?.bar || baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Three chains in one expression
			{
				Code:   `a && a.b || b && b.c || c && c.d;`,
				Output: []string{`a?.b || b?.c || c?.d;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Two chains with nullish checks
			{
				Code:   `foo != null && foo.bar || baz !== undefined && baz.qux;`,
				Output: []string{`foo?.bar || baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Long chains side by side
			{
				Code:   `foo && foo.bar && foo.bar.baz || baz && baz.bar && baz.bar.foo;`,
				Output: []string{`foo?.bar?.baz || baz?.bar?.foo;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Ternary with two chains
			{
				Code:   `cond ? foo && foo.bar : baz && baz.qux;`,
				Output: []string{`cond ? foo?.bar : baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

// =============================================================================
// SECTION 4: Base Cases Tests (Generated from 26 base patterns)
// Source: upstream prefer-optional-chain.test.ts lines 3113-3412
// These use the BaseCases generator with different operators and mutations
// =============================================================================

func TestBaseCasesAndBoolean(t *testing.T) {
	// Base cases with && operator - boolean truthiness check
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator: "&&",
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

func TestBaseCasesAndStrictNotEqualNull(t *testing.T) {
	// With `| null | undefined` type - `!== null` doesn't cover the `undefined` case
	// so optional chaining is NOT a valid conversion - these should be VALID (no error)
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualNull,
	})

	// But if the type is just `| null` (remove `| undefined`), then it covers the cases
	// and IS a valid conversion - these should be INVALID (convert to optional chain)
	// Note: upstream uses suggestions, but our rule auto-fixes so we provide expected output
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "&&",
		MutateCode:        ReplaceOperatorWithStrictNotEqualNull,
		MutateDeclaration: RemoveUndefinedFromType,
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases)
}

func TestBaseCasesAndNotEqualNull(t *testing.T) {
	// Base cases with && operator - != null checks
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithNotEqualNull,
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

func TestBaseCasesAndStrictNotEqualUndefined(t *testing.T) {
	// With `| null | undefined` type - `!== undefined` doesn't cover the `null` case
	// so optional chaining is NOT a valid conversion - these should be VALID (no error)
	// Note: upstream skips IDs 20, 26 for these valid cases
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualUndefined,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})

	// But if the type is just `| undefined` (remove `| null`), then it covers the cases
	// and IS a valid conversion - these should be INVALID (convert to optional chain)
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "&&",
		MutateCode:        ReplaceOperatorWithStrictNotEqualUndefined,
		MutateDeclaration: RemoveNullFromType,
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases)
}

func TestBaseCasesAndNotEqualUndefined(t *testing.T) {
	// Base cases with && operator - != undefined checks
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithNotEqualUndefined,
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

func TestBaseCasesOrBoolean(t *testing.T) {
	// Base cases with || operator - negated boolean truthiness check
	// For || chains, we negate the expressions: !foo || !foo.bar -> !foo?.bar
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator: "||",
		MutateCode: func(s string) string {
			// Add negation before each operand
			return NegateChainOperands(s, "||")
		},
		MutateOutput: func(s string) string {
			// The output should be negated optional chain
			return "!" + s
		},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

func TestBaseCasesOrStrictEqualNull(t *testing.T) {
	// With `| null | undefined` type - `=== null` doesn't cover the `undefined` case
	// so optional chaining is NOT a valid conversion - these should be VALID (no error)
	// Note: upstream skips IDs 20, 26 for these valid cases
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "||",
		MutateCode: ReplaceOperatorWithStrictEqualNull,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})

	// Invalid cases: if the type is just `| null` (remove `| undefined`), then it covers the cases
	// and IS a valid conversion - these should be INVALID (convert to optional chain)
	// Note: upstream adds trailing "=== null" to output for OR chains
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "||",
		MutateCode:        AddTrailingStrictEqualNull(ReplaceOperatorWithStrictEqualNull),
		MutateDeclaration: RemoveUndefinedFromType,
		MutateOutput:      AddTrailingStrictEqualNull(Identity),
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases)
}

func TestBaseCasesOrEqualNull(t *testing.T) {
	// Base cases with || operator - == null checks
	// Note: upstream adds trailing "== null" to the chain for OR patterns
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator:     "||",
		MutateCode:   AddTrailingEqualNull(ReplaceOperatorWithEqualNull),
		MutateOutput: AddTrailingEqualNull(Identity),
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

func TestBaseCasesOrStrictEqualUndefined(t *testing.T) {
	// With `| null | undefined` type - `=== undefined` doesn't cover the `null` case
	// so optional chaining is NOT a valid conversion - these should be VALID (no error)
	// Note: upstream skips IDs 20, 26 for these valid cases
	validCases := GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "||",
		MutateCode: ReplaceOperatorWithStrictEqualUndefined,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})

	// Invalid cases: if the type is just `| undefined` (remove `| null`), then it covers the cases
	// and IS a valid conversion - these should be INVALID (convert to optional chain)
	// Note: upstream adds trailing "=== undefined" to output for OR chains
	invalidCases := GenerateBaseCases(BaseCaseOptions{
		Operator:          "||",
		MutateCode:        AddTrailingStrictEqualUndefined(ReplaceOperatorWithStrictEqualUndefined),
		MutateDeclaration: RemoveNullFromType,
		MutateOutput:      AddTrailingStrictEqualUndefined(Identity),
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		validCases,
		invalidCases)
}

func TestBaseCasesOrEqualUndefined(t *testing.T) {
	// Base cases with || operator - == undefined checks
	// Note: upstream adds trailing "== undefined" to the chain for OR patterns
	cases := GenerateBaseCases(BaseCaseOptions{
		Operator:     "||",
		MutateCode:   AddTrailingEqualUndefined(ReplaceOperatorWithEqualUndefined),
		MutateOutput: AddTrailingEqualUndefined(Identity),
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		cases)
}

// =============================================================================
// SECTION 5: Options Tests
// =============================================================================

func TestOptionsCheckTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases with options disabled
		[]rule_tester.ValidTestCase{
			// checkString option disables string checks
			{
				Code: `declare const foo: string; foo && foo.length;`,
				Options: map[string]interface{}{
					"checkString": false,
				},
			},
			// checkNumber option disables number checks
			{
				Code: `declare const foo: number; foo && foo.toFixed();`,
				Options: map[string]interface{}{
					"checkNumber": false,
				},
			},
			// checkBoolean option disables boolean checks
			{
				Code: `declare const foo: boolean; foo && foo.valueOf();`,
				Options: map[string]interface{}{
					"checkBoolean": false,
				},
			},
			// checkBigInt option disables bigint checks
			{
				Code: `declare const foo: bigint; foo && foo.toString();`,
				Options: map[string]interface{}{
					"checkBigInt": false,
				},
			},
			// checkAny option disables any checks
			{
				Code: `declare const foo: any; foo && foo.bar;`,
				Options: map[string]interface{}{
					"checkAny": false,
				},
			},
			// checkUnknown option disables unknown checks
			{
				Code: `declare const foo: unknown; foo && (foo as any).bar;`,
				Options: map[string]interface{}{
					"checkUnknown": false,
				},
			},
		},
		[]rule_tester.InvalidTestCase{})
}

func TestOptionsRequireNullish(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - requireNullish option set to true
		[]rule_tester.ValidTestCase{
			// With requireNullish=true, truthiness checks on non-nullable types should NOT convert
			// Note: When the type explicitly includes null/undefined, the chain IS converted
			// because the implementation considers explicit nullish types as valid context
			{
				Code: `declare const foo: { bar: number }; foo && foo.bar;`,
				Options: map[string]interface{}{
					"requireNullish": true,
				},
			},
			// Negated truthiness checks on non-nullable types also shouldn't convert
			{
				Code: `declare const foo: { bar: number }; !foo || foo.bar;`,
				Options: map[string]interface{}{
					"requireNullish": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// With requireNullish, explicit nullish checks should still convert
			{
				Code:   `foo != null && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// !== null with requireNullish
			{
				Code:   `declare const foo: { bar: number } | null; foo !== null && foo.bar;`,
				Output: []string{`declare const foo: { bar: number } | null; foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// !== undefined with requireNullish
			{
				Code:   `declare const foo: { bar: number } | undefined; foo !== undefined && foo.bar;`,
				Output: []string{`declare const foo: { bar: number } | undefined; foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// == null || with requireNullish
			{
				Code:   `foo == null || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Multi-part nullish chain with requireNullish
			{
				Code:   `foo != null && foo.bar != null && foo.bar.baz;`,
				Output: []string{`foo?.bar?.baz;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// When the type explicitly includes null/undefined, truthiness checks convert
			// (current implementation behavior - type info is considered nullish context)
			{
				Code:   `declare const foo: { bar: number } | null | undefined; foo && foo.bar;`,
				Output: []string{`declare const foo: { bar: number } | null | undefined; foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Longer truthiness chain with nullable intermediate type
			{
				Code:   `declare const foo: { bar: { baz: number } | null }; foo && foo.bar && foo.bar.baz;`,
				Output: []string{`declare const foo: { bar: { baz: number } | null }; foo?.bar?.baz;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}

func TestOptionsAllowPotentiallyUnsafeFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - when option is false (default), some patterns shouldn't auto-fix
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Without allowPotentiallyUnsafeFixes, should get suggestion instead of auto-fix
			{
				Code: `foo && foo.bar;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `foo?.bar;`,
					}},
				}},
			},
			// With allowPotentiallyUnsafeFixes, should auto-fix
			{
				Code:   `foo && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// More complex chain without option - should suggest
			{
				Code: `foo && foo.bar && foo.bar.baz;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "optionalChainSuggest",
						Output:    `foo?.bar?.baz;`,
					}},
				}},
			},
			// More complex chain with option - should auto-fix
			{
				Code:   `foo && foo.bar && foo.bar.baz;`,
				Output: []string{`foo?.bar?.baz;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}

func TestOptionsCheckTypesInvalid(t *testing.T) {
	// Test invalid cases when check options are enabled (default)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// checkString enabled (default) - should flag string chains
			{
				Code:   `declare const foo: string; foo && foo.length;`,
				Output: []string{`declare const foo: string; foo?.length;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkString": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// checkNumber enabled (default)
			{
				Code:   `declare const foo: number; foo && foo.toFixed();`,
				Output: []string{`declare const foo: number; foo?.toFixed();`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkNumber": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// checkBoolean enabled (default)
			{
				Code:   `declare const foo: boolean; foo && foo.valueOf();`,
				Output: []string{`declare const foo: boolean; foo?.valueOf();`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkBoolean": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// checkBigInt enabled (default)
			{
				Code:   `declare const foo: bigint; foo && foo.toString();`,
				Output: []string{`declare const foo: bigint; foo?.toString();`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkBigInt": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// checkAny enabled (default)
			{
				Code:   `declare const foo: any; foo && foo.bar;`,
				Output: []string{`declare const foo: any; foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkAny": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// checkUnknown enabled (default)
			{
				Code:   `declare const foo: unknown; foo && (foo as any).bar;`,
				Output: []string{`declare const foo: unknown; (foo as any)?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"checkUnknown": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}

// =============================================================================
// SECTION 6: Base Case Variations and Edge Cases
// Source: upstream prefer-optional-chain.test.ts
// =============================================================================

func TestBaseCaseVariations(t *testing.T) {
	// Base case variations with trailing expressions
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - shouldn't convert (different base objects)
		[]rule_tester.ValidTestCase{
			// Different objects in chain - no conversion possible
			{Code: `foo && bar.baz;`},
			{Code: `foo.bar && baz.qux;`},
		},
		// Invalid cases - should convert (partial chains are still flagged)
		[]rule_tester.InvalidTestCase{
			// Chain followed by unrelated expression - still converts the foo chain
			{
				Code:    `foo && foo.bar && bing;`,
				Output:  []string{`foo?.bar && bing;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar && bing.bong;`,
				Output:  []string{`foo?.bar && bing.bong;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo.bar && bing;`,
				Output:  []string{`foo?.bar && bing;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo.bar && bing.bong;`,
				Output:  []string{`foo?.bar && bing.bong;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Assignment expressions are also flagged - rule converts the chain
			{
				Code:    `foo && (bar = foo.baz);`,
				Output:  []string{`foo?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Chain where bing is on same object - full conversion
			{
				Code:    `foo && foo.bar && foo.bar.bing;`,
				Output:  []string{`foo?.bar?.bing;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Trailing access on same chain
			{
				Code:    `foo != null && foo.bar && foo.bar.bing;`,
				Output:  []string{`foo?.bar?.bing;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestWhitespaceVariations(t *testing.T) {
	// Test various whitespace patterns to ensure parsing is robust
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Extra spaces
			{
				Code:    `foo  &&  foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Newlines
			{
				Code: `foo
					&& foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Tabs
			{
				Code:    "foo\t&&\tfoo.bar;",
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Multiple newlines in chain
			{
				Code: `foo
					&& foo.bar
					&& foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// No spaces
			{
				Code:    `foo&&foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Spacing sanity check: extra spaces within property access (`.      `)
			// Note: the rule normalizes whitespace in the fix output
			{
				Code:    `foo && foo.      bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Spacing sanity check: newline within property access (`.\n`)
			// Note: the rule normalizes whitespace in the fix output
			{
				Code: `foo && foo.
bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Spacing sanity check: deep chain with extra spaces
			// Note: the rule normalizes whitespace in the fix output
			{
				Code:    `foo && foo.      bar && foo.      bar.      baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestWeirdNonConstantCases(t *testing.T) {
	// Test valid weird non-constant cases that shouldn't be converted
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases - weird patterns that are valid but shouldn't convert
		[]rule_tester.ValidTestCase{
			// Empty object literal
			{Code: `({}) && ({}).foo;`},
			// Empty array literal
			{Code: `([]) && ([]).length;`},
			// Arrow function
			{Code: `(() => {}) && (() => {}).name;`},
			// Function expression
			{Code: `(function() {}) && (function() {}).name;`},
			// Class expression
			{Code: `(class {}) && (class {}).name;`},
			// Template literal - simple
			{Code: "(`` ) && (`` ).length;"},
			// New expression
			{Code: `new Foo() && new Foo().bar;`},
			// Different function calls - not the same reference
			{Code: `getFoo() && getFoo().bar;`},
			// Await expressions to different calls
			{Code: `(await getFoo()) && (await getFoo()).bar;`},
			// Different array indices
			{Code: `arr[0] && arr[1];`},
			// Computed with different keys
			{Code: `obj['a'] && obj['b'];`},
		},
		// Invalid cases - some "weird" patterns DO convert
		[]rule_tester.InvalidTestCase{
			// Tagged template - rule flags it
			{
				Code:    "(tag``) && (tag``).length;",
				Output:  []string{"tag``?.length;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Regex literal - rule flags it
			{
				Code:    `(/foo/) && (/foo/).source;`,
				Output:  []string{`/foo/?.source;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestSuperAndMetaProperties(t *testing.T) {
	// Tests for super and meta properties - these ARE flagged by the rule
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// super patterns are flagged
			{
				Code:    `class Sub extends Base { method() { super.foo && super.foo.bar; } }`,
				Output:  []string{`class Sub extends Base { method() { super.foo?.bar; } }`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// import.meta chain should convert
			{
				Code:    `import.meta && import.meta.url;`,
				Output:  []string{`import.meta?.url;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestTypeAssertionPatterns(t *testing.T) {
	// Tests for type assertions and casts - all flagged by the rule
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// as const pattern - flagged
			{
				Code:    `(foo as const) && (foo as const).bar;`,
				Output:  []string{`(foo as const)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Different type assertions - still flagged, converts to second assertion's type
			{
				Code:    `(foo as string) && (foo as number).bar;`,
				Output:  []string{`(foo as number)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Same type assertion should convert
			{
				Code:    `(foo as Bar) && (foo as Bar).baz;`,
				Output:  []string{`(foo as Bar)?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			// Angle bracket assertion
			{
				Code:    `(<Bar>foo) && (<Bar>foo).baz;`,
				Output:  []string{`(<Bar>foo)?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestSatisfiesExpression(t *testing.T) {
	// Tests for satisfies expressions (TS 4.9+)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// satisfies doesn't create a reference equality issue
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// satisfies expression should convert
			{
				Code:    `(foo satisfies Bar) && (foo satisfies Bar).baz;`,
				Output:  []string{`foo satisfies Bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}

func TestJSXPatterns(t *testing.T) {
	// Tests for JSX patterns - ensure essential whitespace isn't removed
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// JSX with arrow function containing self-closing element with spaces
			// Essential whitespace must be preserved: <This Requires Spaces />
			{
				Code:    `foo && foo.bar(baz => <This Requires Spaces />);`,
				Output:  []string{`foo?.bar(baz => <This Requires Spaces />);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Tsx:     true,
			},
			// JSX with arrow function containing element with children
			{
				Code:    `foo && foo.bar(baz => <div>{baz}</div>);`,
				Output:  []string{`foo?.bar(baz => <div>{baz}</div>);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Tsx:     true,
			},
			// JSX with fragment
			{
				Code:    `foo && foo.bar(baz => <><span>{baz}</span></>);`,
				Output:  []string{`foo?.bar(baz => <><span>{baz}</span></>);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
				Tsx:     true,
			},
			// Arrow function with typeof - ensure whitespace preserved
			{
				Code:    `foo && foo.bar(baz => typeof baz);`,
				Output:  []string{`foo?.bar(baz => typeof baz);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]interface{}{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
