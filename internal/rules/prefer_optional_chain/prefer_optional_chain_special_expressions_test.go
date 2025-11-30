package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainSpecialExpressions tests special expression types
// This implements Category 11 from the test plan: Special Expressions

// Category 11.1: Async/await expressions
// Per upstream: (await foo) && (await foo).bar is NOT converted (marked as TODO)
// But (await foo).bar && (await foo).bar.baz IS converted
func TestPreferOptionalChainAsyncAwait(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Property of awaited value - not detected (limitation)
		{Code: `async function test(foo: { bar: Promise<{ baz: string }> } | null) { foo && (await foo.bar).baz; }`},
		// Await at base level - not handled by upstream (marked as TODO in upstream tests)
		{Code: `async function test(bar: Promise<{ baz: string } | null>) { (await bar) && (await bar).baz; }`},
	}, []rule_tester.InvalidTestCase{
		// Await before property access - this works
		{
			Code:   `async function test() { const foo = await bar(); foo && foo.baz; }`,
			Output: []string{`async function test() { const foo = await bar(); foo?.baz; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.2: Typeof expressions
// NOTE: Our rule converts the entire chain, not preserving typeof checks
func TestPreferOptionalChainTypeOf(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Typeof check before property access - our rule is more aggressive
		{
			Code:   `declare const foo: { bar: string } | null; typeof foo !== 'undefined' && foo !== null && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.3: Void expressions
func TestPreferOptionalChainVoid(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Void 0 is undefined
		{
			Code:   `declare const foo: { bar: string } | undefined; foo !== void 0 && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | undefined; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.4: Template literal expressions
// INTENTIONAL LIMITATION: Property accesses inside template literals cannot be safely converted
// Reason: Semantic difference - short-circuit vs undefined in template
// foo && `value: ${foo.bar}` returns foo (falsy) if null, but `value: ${foo?.bar}` evaluates the template with undefined
func TestPreferOptionalChainTemplateLiterals(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Template literal property access - intentional limitation
		{Code: "declare const foo: { bar: string } | null; foo && `value: ${foo.bar}`;"},
		// Nested property in template - intentional limitation
		{Code: "declare const foo: { bar: { baz: string } | null }; foo.bar && `value: ${foo.bar.baz}`;"},
	}, []rule_tester.InvalidTestCase{})
}

// Category 11.5: Spread expressions
// INTENTIONAL LIMITATION: Property accesses inside spread expressions cannot be safely converted
// Reason: Semantic difference - returns falsy vs throws error
// foo && [...foo.items] returns foo (falsy) if null, but [...foo?.items] would try to spread undefined (error)
func TestPreferOptionalChainSpread(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Spread in array - intentional limitation
		{Code: `declare const foo: { items: string[] } | null; foo && [...foo.items];`},
		// Spread in object - intentional limitation
		{Code: `declare const foo: { bar: { baz: string } } | null; foo && { ...foo.bar };`},
	}, []rule_tester.InvalidTestCase{})
}

// Category 11.6: Destructuring with optional chaining
// NOTE: Our rule is more aggressive and creates more optional chains
func TestPreferOptionalChainDestructuring(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Destructure after check - our rule creates more optional chains
		{
			Code:   `declare const foo: { bar: { baz: string } } | null; const baz = foo && foo.bar.baz;`,
			Output: []string{`declare const foo: { bar: { baz: string } } | null; const baz = foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// In variable declaration
		{
			Code:   `declare const foo: { bar: string } | null; const x = foo && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null; const x = foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.7: Non-null assertion with optional chaining
// NOTE: Our rule is more aggressive with optional chains
func TestPreferOptionalChainNonNullAssertion(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Non-null assertion in chain - our rule creates more optional chains
		{
			Code:   `declare const foo: { bar: { baz: string | null } } | null; foo && foo.bar.baz!;`,
			Output: []string{`declare const foo: { bar: { baz: string | null } } | null; foo?.bar?.baz!;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Non-null on intermediate - our rule creates more optional chains
		{
			Code:   `declare const foo: { bar: { baz: string } | null } | null; foo && foo.bar!.baz;`,
			Output: []string{`declare const foo: { bar: { baz: string } | null } | null; foo?.bar!?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.8: Parenthesized expressions in chains
// FIXED: Chains with parentheses are now detected and stripped correctly!
func TestPreferOptionalChainParentheses(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Parentheses around entire check - FIXED!
		{
			Code:   `declare const foo: { bar: string } | null; (foo) && (foo).bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Parentheses around property access - FIXED!
		{
			Code:   `declare const foo: { bar: { baz: string } } | null; foo && (foo.bar).baz;`,
			Output: []string{`declare const foo: { bar: { baz: string } } | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Multiple levels of parentheses - FIXED!
		{
			Code:   `declare const foo: { bar: string } | null; ((foo)) && ((foo)).bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Parentheses in middle of chain
		{
			Code:   `declare const foo: { bar: { baz: { qux: string } } } | null; foo && (foo.bar) && (foo.bar).baz;`,
			Output: []string{`declare const foo: { bar: { baz: { qux: string } } } | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.9: JSX and TSX expressions
func TestPreferOptionalChainJSX(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// JSX attribute with chain (semantic difference, don't convert)
		{Code: `declare const foo: { bar: string } | null; <Component value={foo && foo.bar} />;`, Tsx: true},
		// JSX children with chain (semantic difference)
		{Code: `declare const foo: { bar: string } | null; <div>{foo && foo.bar}</div>;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{})
}

// Category 11.10: Computed property names with special characters
func TestPreferOptionalChainComputedPropertyEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Computed with string literal
		{
			Code:   `declare const foo: { ['bar-baz']: string } | null; foo && foo['bar-baz'];`,
			Output: []string{`declare const foo: { ['bar-baz']: string } | null; foo?.['bar-baz'];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Computed with number
		{
			Code:   `declare const foo: { [0]: string } | null; foo && foo[0];`,
			Output: []string{`declare const foo: { [0]: string } | null; foo?.[0];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Computed with expression
		{
			Code:   `declare const foo: Record<string, string> | null; const key = 'bar'; foo && foo[key];`,
			Output: []string{`declare const foo: Record<string, string> | null; const key = 'bar'; foo?.[key];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Computed with template literal
		{
			Code:   "declare const foo: Record<string, string> | null; foo && foo[`bar`];",
			Output: []string{"declare const foo: Record<string, string> | null; foo?.[`bar`];"},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.11: instanceof and constructor checks
func TestPreferOptionalChainInstanceOf(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// instanceof before property access
		{
			Code:   `class MyClass { bar: string; } declare const foo: MyClass | null; foo instanceof MyClass && foo && foo.bar;`,
			Output: []string{`class MyClass { bar: string; } declare const foo: MyClass | null; foo instanceof MyClass && foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.12: Falsy value distinctions
func TestPreferOptionalChainFalsyValues(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Empty string is falsy, semantic difference
		{Code: `declare const foo: { bar: string } | ''; foo && foo.bar;`},
		// Zero is falsy, semantic difference
		{Code: `declare const foo: { bar: string } | 0; foo && foo.bar;`},
		// False is falsy, semantic difference
		{Code: `declare const foo: { bar: string } | false; foo && foo.bar;`},
	}, []rule_tester.InvalidTestCase{})
}

// Category 11.13: Type assertions in chains
func TestPreferOptionalChainTypeAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// as assertion
		{
			Code:   `declare const foo: any; foo && (foo as { bar: string }).bar;`,
			Output: []string{`declare const foo: any; (foo as { bar: string })?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Angle bracket assertion (TSX compatible)
		{
			Code:   `declare const foo: unknown | null; foo && (<{ bar: string }>foo).bar;`,
			Output: []string{`declare const foo: unknown | null; (<{ bar: string }>foo)?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 11.14: Comments preservation
func TestPreferOptionalChainComments(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Comment before property access
		{
			Code: `declare const foo: { bar: string } | null;
foo && /* important */ foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null;
/* important */ foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
