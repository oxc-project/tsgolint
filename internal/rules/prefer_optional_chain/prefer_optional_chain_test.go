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
			{Code: `(foo || {})?.bar;`},                // Already optional
			{Code: `(foo || { bar: 1 }).bar;`},         // Non-empty object
			{Code: `foo ||= bar || {};`},               // Assignment operator
			{Code: `(foo1 ? foo2 : foo3 || {}).foo4;`}, // Ternary in wrong position
			{Code: `(undefined && (foo || {})).bar;`},  // Complex condition
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
		})
}

// =============================================================================
// SECTION 1b: ?? {} Tests (Nullish Coalescing Empty Object Pattern)
// =============================================================================

func TestNullishCoalescingEmptyObjectPattern(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `foo ?? {};`},        // No property access
			{Code: `(foo ?? {})?.bar;`}, // Already optional
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
			// !foo && patterns with comparisons - inverted checks don't convert
			{Code: `!foo && foo.bar == 0;`},
			{Code: `!foo && foo.bar == 1;`},
			{Code: `!foo && foo.bar === 0;`},
			{Code: `!foo && foo.bar === null;`},
			{Code: `!foo && foo.bar !== undefined;`},
			{Code: `!foo && foo.bar != null;`},
			// foo == null && patterns - inverted nullish checks don't convert
			{Code: `foo == null && foo.bar == 0;`},
			{Code: `foo == null && foo.bar === 1;`},
			{Code: `foo == null && foo.bar === null;`},
			{Code: `foo == null && foo.bar != null;`},
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
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// With requireNullish, nullish checks should still convert
			{
				Code:   `foo != null && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Options: map[string]interface{}{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
					"requireNullish": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}
