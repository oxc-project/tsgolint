package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainOptionCombinations tests all combinations of the three options
func TestPreferOptionalChainOptionCombinations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// requireNullish: true prevents conversion without explicit nullish checks
		{
			Code:    `declare const foo: {bar: string} | false; foo && foo.bar;`,
			Options: map[string]any{"requireNullish": true},
		},
		{
			Code:    `declare const foo: string | 0; foo && foo.length;`,
			Options: map[string]any{"requireNullish": true},
		},

		// checkString: false prevents conversion on string types
		{
			Code:    `declare const foo: string | null; foo && foo.length;`,
			Options: map[string]any{"checkString": false},
		},
		{
			Code:    `declare const foo: string | undefined; foo && foo.charAt(0);`,
			Options: map[string]any{"checkString": false},
		},

		// Combination: requireNullish + checkString: false
		{
			Code: `declare const foo: string | false; foo && foo.length;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    false,
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Default options (all false)
		{
			Code:   `declare const foo: {bar: string} | null; foo && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": false,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// requireNullish: true with explicit nullish check
		{
			Code:   `declare const foo: {bar: string} | null; foo != null && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: {bar: string} | undefined; foo !== undefined && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | undefined; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// checkString: true with string types
		{
			Code:   `declare const foo: string | null; foo && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"checkString": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: string | undefined; foo && foo.charAt(0);`,
			Output: []string{`declare const foo: string | undefined; foo?.charAt(0);`},
			Options: map[string]any{
				"checkString": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: true
		{
			Code:   `declare const foo: {bar: string} | null; foo && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Combination: requireNullish + allowPotentiallyUnsafe
		{
			Code:   `declare const foo: {bar: string} | null; foo !== null && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Combination: checkString + allowPotentiallyUnsafe
		{
			Code:   `declare const foo: string | null; foo && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"checkString": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// All three options combined
		{
			Code:   `declare const foo: string | null; foo !== null && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: string | undefined; foo != undefined && foo.charAt(0);`,
			Output: []string{`declare const foo: string | undefined; foo?.charAt(0);`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Deep chains with options
		{
			Code:   `declare const foo: {bar: {baz: string} | null} | null; foo != null && foo.bar != null && foo.bar.baz;`,
			Output: []string{`declare const foo: {bar: {baz: string} | null} | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Call chains with options
		{
			Code:   `declare const foo: {bar: (() => string) | null} | null; foo !== null && foo.bar !== null && foo.bar();`,
			Output: []string{`declare const foo: {bar: (() => string) | null} | null; foo?.bar?.();`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Element access with options
		{
			Code:   `declare const foo: {[key: string]: string} | null; foo != null && foo['bar'];`,
			Output: []string{`declare const foo: {[key: string]: string} | null; foo?.['bar'];`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// String with deep chain and all options
		{
			Code:   `declare const foo: {bar: string | null} | null; foo !== null && foo.bar !== null && foo.bar.length;`,
			Output: []string{`declare const foo: {bar: string | null} | null; foo?.bar?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// OR operator with options
		{
			Code:   `declare const foo: {bar: string} | null; foo === null || foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Negation with options
		{
			Code:   `declare const foo: {bar: string} | null; !foo || !foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; !foo?.bar;`},
			Options: map[string]any{
				"requireNullish": false,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// typeof with all options
		{
			Code:   `declare const foo: string | null; typeof foo !== 'undefined' && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Yoda-style with options
		{
			Code:   `declare const foo: {bar: string} | null; null !== foo && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Mixed checks with options
		{
			Code:   `declare const foo: {bar: {baz: string} | undefined} | null; foo != null && foo.bar !== undefined && foo.bar.baz;`,
			Output: []string{`declare const foo: {bar: {baz: string} | undefined} | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Complex expression with options
		{
			Code:   `declare const foo: {bar: {baz: {qux: string} | null} | undefined} | null; foo !== null && typeof foo.bar !== 'undefined' && null !== foo.bar.baz && foo.bar.baz.qux;`,
			Output: []string{`declare const foo: {bar: {baz: {qux: string} | null} | undefined} | null; foo?.bar?.baz?.qux;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// String method chains with all options
		{
			Code:   `declare const foo: {bar: string | null} | null; foo != null && foo.bar != null && foo.bar.toUpperCase();`,
			Output: []string{`declare const foo: {bar: string | null} | null; foo?.bar?.toUpperCase();`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Array element access with options
		{
			Code:   `declare const foo: string[] | null; foo !== null && foo[0];`,
			Output: []string{`declare const foo: string[] | null; foo?.[0];`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// String property access with options
		{
			Code:   `declare const foo: {name: string | null} | null; foo !== null && foo.name !== null && foo.name.length;`,
			Output: []string{`declare const foo: {name: string | null} | null; foo?.name?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Function return with options
		{
			Code:   `declare const foo: {bar: () => {baz: string} | null} | null; function test() { return foo !== null && foo.bar !== null && foo.bar() !== null && foo.bar().baz; }`,
			Output: []string{`declare const foo: {bar: () => {baz: string} | null} | null; function test() { return foo?.bar?.()?.baz; }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Conditional with options
		{
			Code:   `declare const foo: {bar: string} | null; if (foo !== null && foo.bar) {}`,
			Output: []string{`declare const foo: {bar: string} | null; if (foo?.bar) {}`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// While loop with options
		{
			Code:   `declare const foo: {bar: {count: number} | null} | null; while (foo != null && foo.bar != null && foo.bar.count > 0) {}`,
			Output: []string{`declare const foo: {bar: {count: number} | null} | null; while (foo?.bar?.count > 0) {}`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Ternary with options
		{
			Code:   `declare const foo: {bar: string} | null; const x = foo !== null && foo.bar ? 'yes' : 'no';`,
			Output: []string{`declare const foo: {bar: string} | null; const x = foo?.bar ? 'yes' : 'no';`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Template literal with options
		{
			Code:   "declare const foo: {name: string | null} | null; const msg = `Hello ${foo !== null && foo.name !== null && foo.name}`;",
			Output: []string{"declare const foo: {name: string | null} | null; const msg = `Hello ${foo?.name}`;"},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Arrow function with options
		{
			Code:   `declare const foo: {bar: string} | null; const fn = () => foo !== null && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; const fn = () => foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Object property with options
		{
			Code:   `declare const foo: {bar: string} | null; const obj = { value: foo !== null && foo.bar };`,
			Output: []string{`declare const foo: {bar: string} | null; const obj = { value: foo?.bar };`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Array element with options
		{
			Code:   `declare const foo: {bar: string} | null; const arr = [foo !== null && foo.bar];`,
			Output: []string{`declare const foo: {bar: string} | null; const arr = [foo?.bar];`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Switch with options
		{
			Code:   `declare const foo: {bar: string} | null; switch (foo !== null && foo.bar) { case 'test': break; }`,
			Output: []string{`declare const foo: {bar: string} | null; switch (foo?.bar) { case 'test': break; }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Complex nested with all options
		{
			Code:   `declare const data: {user: {profile: {name: string | null} | null} | null} | null; const name = data !== null && data.user !== null && data.user.profile !== null && data.user.profile.name !== null && data.user.profile.name.toUpperCase();`,
			Output: []string{`declare const data: {user: {profile: {name: string | null} | null} | null} | null; const name = data?.user?.profile?.name?.toUpperCase();`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Generic function with options
		{
			Code:   `function get<T extends {bar: string} | null>(foo: T) { return foo !== null && foo.bar; }`,
			Output: []string{`function get<T extends {bar: string} | null>(foo: T) { return foo?.bar; }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Class method with options
		{
			Code:   `class Foo { bar: {baz: string} | null; method() { return this.bar !== null && this.bar.baz; } }`,
			Output: []string{`class Foo { bar: {baz: string} | null; method() { return this.bar?.baz; } }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Static method with options
		{
			Code:   `class Foo { static data: {value: string} | null; static get() { return Foo.data !== null && Foo.data.value; } }`,
			Output: []string{`class Foo { static data: {value: string} | null; static get() { return Foo.data?.value; } }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Getter with options
		{
			Code:   `class Foo { private _bar: {baz: string} | null; get bar() { return this._bar !== null && this._bar.baz; } }`,
			Output: []string{`class Foo { private _bar: {baz: string} | null; get bar() { return this._bar?.baz; } }`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
