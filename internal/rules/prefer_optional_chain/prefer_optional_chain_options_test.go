package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainOptions tests option-specific behavior
// This implements Category 9 from the test TODO: option-specific tests

// Category 9.1 & 9.2: requireNullish: true - should convert only with explicit nullish checks
// Note: Category 8.6 already tested some requireNullish cases, but these are more comprehensive
func TestPreferOptionalChainOptionRequireNullish(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Plain && should NOT convert when requireNullish is true
		{Code: `foo && foo.bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo.bar && foo.bar.baz;`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo && foo[bar];`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo[bar] && foo[bar].baz;`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo.bar && foo.bar();`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo && foo.bar && foo.bar.baz;`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo[bar] && foo[bar][baz];`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo && foo.bar && foo.bar.baz && foo.bar.baz.qux;`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo.bar && foo.bar.baz();`, Options: map[string]any{"requireNullish": true}},
		{Code: `foo && foo.bar.baz;`, Options: map[string]any{"requireNullish": true}},

		// With union types that are falsy but NOT nullish - should be valid
		{Code: `declare const foo: { bar: string } | false; foo && foo.bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: string | 0; foo && foo.length;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: string | ""; foo && foo.length;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: number | 0; foo && foo.toFixed();`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: bigint | 0n; foo && foo.toString();`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: any | false; foo && foo.bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: unknown | ""; foo && (foo as any).bar;`, Options: map[string]any{"requireNullish": true}},

		// With multiple union members but no null/undefined
		{Code: `declare const foo: string | number | false; foo && foo.toString();`, Options: map[string]any{"requireNullish": true}},

		// With array and object types containing falsy unions but no null/undefined
		{Code: `declare const foo: Array<string> | false; foo && foo.length;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const foo: { bar: string } | 0; foo && foo.bar;`, Options: map[string]any{"requireNullish": true}},
	}, []rule_tester.InvalidTestCase{
		// Union types with null/undefined should report with suggestion (no autofix)
		// These are INVALID because the type includes null/undefined
		{
			Code:    `declare const foo: string | null; foo && foo.length;`,
			Options: map[string]any{"requireNullish": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `declare const foo: string | null; foo?.length;`,
				}},
			}},
		},
		{
			Code:    `declare const foo: number | undefined; foo && foo.toFixed();`,
			Options: map[string]any{"requireNullish": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `declare const foo: number | undefined; foo?.toFixed();`,
				}},
			}},
		},
		{
			Code:    `declare const foo: boolean | null; foo && foo.valueOf();`,
			Options: map[string]any{"requireNullish": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `declare const foo: boolean | null; foo?.valueOf();`,
				}},
			}},
		},
		{
			Code:    `declare const foo: boolean | null | 0; foo && foo.valueOf();`,
			Options: map[string]any{"requireNullish": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `declare const foo: boolean | null | 0; foo?.valueOf();`,
				}},
			}},
		},
		{
			Code:    `declare const foo: string | false | null; foo && foo.length;`,
			Options: map[string]any{"requireNullish": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    `declare const foo: string | false | null; foo?.length;`,
				}},
			}},
		},
		// Explicit nullish checks SHOULD convert
		{
			Code:   `foo != null && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo !== null && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo !== undefined && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != undefined && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Typed variable with strict nullish check
		{
			Code:   `declare const foo: {bar: string} | null; foo !== null && foo.bar;`,
			Output: []string{`declare const foo: {bar: string} | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar && foo.bar.baz;`,
			Output: []string{`foo?.bar?.baz;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo !== null && foo[bar];`,
			Output: []string{`foo?.[bar];`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar();`,
			Output: []string{`foo?.bar();`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Explicit nullish checks with typed variables
		{
			Code:   `declare const foo: { bar: string } | null; foo !== null && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: string | undefined; foo !== undefined && foo.length;`,
			Output: []string{`declare const foo: string | undefined; foo?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.3: requireNullish with empty object pattern
func TestPreferOptionalChainOptionRequireNullishEmptyObject(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Empty object patterns should NOT convert with requireNullish: true
		{Code: `(foo || {}).bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `(foo ?? {}).bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `(foo.bar || {}).baz;`, Options: map[string]any{"requireNullish": true}},
		{Code: `(foo || {}).bar.baz;`, Options: map[string]any{"requireNullish": true}},
		{Code: `(foo ?? {}).bar();`, Options: map[string]any{"requireNullish": true}},
	}, []rule_tester.InvalidTestCase{})
}

// Category 9.5: checkAny: false - should not convert when type is any
func TestPreferOptionalChainOptionCheckAny(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkAny: false, should NOT convert any types
		{Code: `declare const foo: any; foo && foo.bar;`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo.bar && foo.bar.baz;`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo && foo[bar];`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo && foo.bar();`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo && foo.bar.baz;`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo.bar && foo.bar.baz && foo.bar.baz.qux;`, Options: map[string]any{"checkAny": false}},
		{Code: `declare const foo: any; foo != null && foo.bar;`, Options: map[string]any{"checkAny": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkAny: true (default), SHOULD convert any types
		{
			Code:   `declare const foo: any; foo && foo.bar;`,
			Output: []string{`declare const foo: any; foo?.bar;`},
			Options: map[string]any{
				"checkAny": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: any; foo.bar && foo.bar.baz;`,
			Output: []string{`declare const foo: any; foo.bar?.baz;`},
			Options: map[string]any{
				"checkAny": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: any; foo && foo[bar];`,
			Output: []string{`declare const foo: any; foo?.[bar];`},
			Options: map[string]any{
				"checkAny": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.6: checkBigInt: false - should not convert when type is bigint
func TestPreferOptionalChainOptionCheckBigInt(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkBigInt: false, should NOT convert bigint types
		{Code: `declare const foo: bigint; foo && foo.toString();`, Options: map[string]any{"checkBigInt": false}},
		{Code: `declare const foo: bigint | null; foo && foo.toString();`, Options: map[string]any{"checkBigInt": false}},
		{Code: `declare const foo: bigint; foo && foo.valueOf();`, Options: map[string]any{"checkBigInt": false}},
		{Code: `declare const foo: bigint | undefined; foo && foo.toString();`, Options: map[string]any{"checkBigInt": false}},
		{Code: `declare const foo: bigint | null | undefined; foo && foo.valueOf();`, Options: map[string]any{"checkBigInt": false}},
		{Code: `declare const foo: bigint; foo != null && foo.toString();`, Options: map[string]any{"checkBigInt": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkBigInt: true (default), SHOULD convert bigint types with nullish unions
		{
			Code:   `declare const foo: bigint | null; foo && foo.toString();`,
			Output: []string{`declare const foo: bigint | null; foo?.toString();`},
			Options: map[string]any{
				"checkBigInt": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: bigint | undefined; foo && foo.valueOf();`,
			Output: []string{`declare const foo: bigint | undefined; foo?.valueOf();`},
			Options: map[string]any{
				"checkBigInt": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: bigint | null | undefined; foo != null && foo.toString();`,
			Output: []string{`declare const foo: bigint | null | undefined; foo?.toString();`},
			Options: map[string]any{
				"checkBigInt": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.7: checkBoolean: false - should not convert when type is boolean
func TestPreferOptionalChainOptionCheckBoolean(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkBoolean: false, should NOT convert boolean types
		{Code: `declare const foo: boolean; foo && foo.valueOf();`, Options: map[string]any{"checkBoolean": false}},
		{Code: `declare const foo: boolean | undefined; foo && foo.valueOf();`, Options: map[string]any{"checkBoolean": false}},
		{Code: `declare const foo: boolean; foo && foo.toString();`, Options: map[string]any{"checkBoolean": false}},
		{Code: `declare const foo: boolean | null; foo && foo.toString();`, Options: map[string]any{"checkBoolean": false}},
		{Code: `declare const foo: boolean | null | undefined; foo && foo.valueOf();`, Options: map[string]any{"checkBoolean": false}},
		{Code: `declare const foo: boolean; foo != null && foo.toString();`, Options: map[string]any{"checkBoolean": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkBoolean: true (default), SHOULD convert boolean types with nullish unions
		{
			Code:   `declare const foo: boolean | null; foo && foo.valueOf();`,
			Output: []string{`declare const foo: boolean | null; foo?.valueOf();`},
			Options: map[string]any{
				"checkBoolean": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: boolean | undefined; foo && foo.toString();`,
			Output: []string{`declare const foo: boolean | undefined; foo?.toString();`},
			Options: map[string]any{
				"checkBoolean": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: boolean | null | undefined; foo != null && foo.valueOf();`,
			Output: []string{`declare const foo: boolean | null | undefined; foo?.valueOf();`},
			Options: map[string]any{
				"checkBoolean": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.8: checkNumber: false - should not convert when type is number
func TestPreferOptionalChainOptionCheckNumber(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkNumber: false, should NOT convert number types
		{Code: `declare const foo: number; foo && foo.toFixed();`, Options: map[string]any{"checkNumber": false}},
		{Code: `declare const foo: number | null; foo && foo.toFixed();`, Options: map[string]any{"checkNumber": false}},
		{Code: `declare const foo: number; foo && foo.toString();`, Options: map[string]any{"checkNumber": false}},
		{Code: `declare const foo: number | undefined; foo && foo.valueOf();`, Options: map[string]any{"checkNumber": false}},
		{Code: `declare const foo: number | null | undefined; foo && foo.toFixed();`, Options: map[string]any{"checkNumber": false}},
		{Code: `declare const foo: number; foo != null && foo.toFixed();`, Options: map[string]any{"checkNumber": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkNumber: true (default), SHOULD convert number types with nullish unions
		{
			Code:   `declare const foo: number | undefined; foo && foo.toFixed();`,
			Output: []string{`declare const foo: number | undefined; foo?.toFixed();`},
			Options: map[string]any{
				"checkNumber": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: number | null; foo && foo.toString();`,
			Output: []string{`declare const foo: number | null; foo?.toString();`},
			Options: map[string]any{
				"checkNumber": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: number | null | undefined; foo != null && foo.toFixed();`,
			Output: []string{`declare const foo: number | null | undefined; foo?.toFixed();`},
			Options: map[string]any{
				"checkNumber": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.9: checkString: false - should not convert when type is string
func TestPreferOptionalChainOptionCheckString(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkString: false, should NOT convert string types
		{Code: `declare const foo: string; foo && foo.length;`, Options: map[string]any{"checkString": false}},
		{Code: `declare const foo: string | undefined; foo && foo.length;`, Options: map[string]any{"checkString": false}},
		{Code: `declare const foo: string; foo && foo.toString();`, Options: map[string]any{"checkString": false}},
		{Code: `declare const foo: string | null; foo && foo.charAt(0);`, Options: map[string]any{"checkString": false}},
		{Code: `declare const foo: string | null | undefined; foo && foo.length;`, Options: map[string]any{"checkString": false}},
		{Code: `declare const foo: string; foo != null && foo.length;`, Options: map[string]any{"checkString": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkString: true (default), SHOULD convert string types with nullish unions
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
		{
			Code:   `declare const foo: string | null | undefined; foo != null && foo.length;`,
			Output: []string{`declare const foo: string | null | undefined; foo?.length;`},
			Options: map[string]any{
				"checkString": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.10: checkUnknown: false - should not convert when type is unknown
// NOTE: The invalid case (checkUnknown: true) is currently not working
// Issue: When checking `foo && (foo as any).bar` with foo: unknown,
// the rule should convert but doesn't. This might be related to how
// type assertions interact with the base type checking.
func TestPreferOptionalChainOptionCheckUnknown(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// With checkUnknown: false, should NOT convert unknown types
		{Code: `declare const foo: unknown; foo && (foo as any).bar;`, Options: map[string]any{"checkUnknown": false}},
		{Code: `declare const foo: unknown; foo && (foo as { bar: string }).bar;`, Options: map[string]any{"checkUnknown": false}},
		{Code: `declare const foo: unknown; foo && (foo as any).bar();`, Options: map[string]any{"checkUnknown": false}},
		{Code: `declare const foo: unknown; foo && (foo as any)[bar];`, Options: map[string]any{"checkUnknown": false}},
	}, []rule_tester.InvalidTestCase{
		// With checkUnknown: true (default), SHOULD convert unknown types
		{
			Code:   `declare const foo: unknown; foo && (foo as any).bar;`,
			Output: []string{`declare const foo: unknown; (foo as any)?.bar;`},
			Options: map[string]any{
				"checkUnknown": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: unknown; foo && (foo as { bar: string }).bar;`,
			Output: []string{`declare const foo: unknown; (foo as { bar: string })?.bar;`},
			Options: map[string]any{
				"checkUnknown": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.11: Option combinations - test multiple options together
func TestPreferOptionalChainMultipleOptions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Multiple check options disabled
		{
			Code: `declare const foo: string; foo && foo.length;`,
			Options: map[string]any{
				"checkString":  false,
				"checkNumber":  false,
				"checkBoolean": false,
			},
		},
		{
			Code: `declare const foo: number; foo && foo.toFixed();`,
			Options: map[string]any{
				"checkString":  false,
				"checkNumber":  false,
				"checkBoolean": false,
			},
		},

		// requireNullish with check options
		{
			Code: `declare const foo: string; foo && foo.length;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
			},
		},
		{
			Code: `declare const foo: string | null; foo && foo.length;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    false,
			},
		},

		// All check options disabled should NOT convert any primitive type
		{
			Code: `declare const foo: string | null; foo && foo.length;`,
			Options: map[string]any{
				"checkString":  false,
				"checkNumber":  false,
				"checkBoolean": false,
				"checkBigInt":  false,
			},
		},
		{
			Code: `declare const foo: number | undefined; foo && foo.toFixed();`,
			Options: map[string]any{
				"checkString":  false,
				"checkNumber":  false,
				"checkBoolean": false,
				"checkBigInt":  false,
			},
		},

		// requireNullish with all check options disabled
		{
			Code: `declare const foo: string | null; foo != null && foo.length;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    false,
				"checkNumber":    false,
			},
		},

		// checkAny false with other checks enabled
		{
			Code: `declare const foo: any; foo && foo.bar;`,
			Options: map[string]any{
				"checkAny":    false,
				"checkString": true,
				"checkNumber": true,
			},
		},

		// requireNullish true with checkAny false
		{
			Code: `declare const foo: any; foo && foo.bar;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkAny":       false,
			},
		},
		{
			Code: `declare const foo: any; foo != null && foo.bar;`,
			Options: map[string]any{
				"requireNullish": true,
				"checkAny":       false,
			},
		},
	}, []rule_tester.InvalidTestCase{
		// requireNullish with explicit check + check option enabled
		{
			Code:   `declare const foo: string | null; foo != null && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple check options enabled with object type
		{
			Code:   `declare const foo: { bar: string } | null; foo && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"checkString": true,
				"checkNumber": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// All check options enabled
		{
			Code:   `declare const foo: any; foo && foo.bar;`,
			Output: []string{`declare const foo: any; foo?.bar;`},
			Options: map[string]any{
				"checkAny":     true,
				"checkString":  true,
				"checkNumber":  true,
				"checkBoolean": true,
				"checkBigInt":  true,
				"checkUnknown": true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple types with selective check options enabled
		{
			Code:   `declare const foo: string | null; foo && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"checkString":  true,
				"checkNumber":  false,
				"checkBoolean": false,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// requireNullish false with explicit nullish check still converts
		{
			Code:   `declare const foo: string | null; foo != null && foo.length;`,
			Output: []string{`declare const foo: string | null; foo?.length;`},
			Options: map[string]any{
				"requireNullish": false,
				"checkString":    true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// checkAny true with requireNullish true and explicit check
		{
			Code:   `declare const foo: any; foo != null && foo.bar;`,
			Output: []string{`declare const foo: any; foo?.bar;`},
			Options: map[string]any{
				"requireNullish": true,
				"checkAny":       true,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Multiple different primitive types with their checks enabled
		{
			Code:   `declare const foo: number | null; foo && foo.toFixed();`,
			Output: []string{`declare const foo: number | null; foo?.toFixed();`},
			Options: map[string]any{
				"checkNumber": true,
				"checkString": false,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: boolean | undefined; foo && foo.valueOf();`,
			Output: []string{`declare const foo: boolean | undefined; foo?.valueOf();`},
			Options: map[string]any{
				"checkBoolean": true,
				"checkString":  false,
				"checkNumber":  false,
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}

// Category 9.4: allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: false
// When false, should provide suggestions instead of automatic fixes
func TestPreferOptionalChainOptionAllowUnsafeFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Without the option (or false), should report error and provide suggestion (not automatic fix)
		{
			Code:    `foo && foo.bar;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
					},
				},
			},
		},
		{
			Code:    `foo.bar && foo.bar.baz;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo.bar?.baz;`},
					},
				},
			},
		},
		{
			Code:    `foo && foo[bar];`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.[bar];`},
					},
				},
			},
		},
		{
			Code:    `foo && foo.bar();`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.bar();`},
					},
				},
			},
		},
		{
			Code:    `foo != null && foo.bar;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
					},
				},
			},
		},
		{
			Code:    `foo.bar && foo.bar.baz && foo.bar.baz.qux;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo.bar?.baz?.qux;`},
					},
				},
			},
		},

		// With the option set to true, should report error AND provide automatic fix
		{
			Code:   `foo && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo.bar && foo.bar.baz;`,
			Output: []string{`foo.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo[bar];`,
			Output: []string{`foo?.[bar];`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo && foo.bar();`,
			Output: []string{`foo?.bar();`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && foo.bar;`,
			Output: []string{`foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo.bar && foo.bar.baz && foo.bar.baz.qux;`,
			Output: []string{`foo.bar?.baz?.qux;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// With typed variables
		{
			Code:    `declare const foo: { bar: string } | null; foo && foo.bar;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: { bar: string } | null; foo?.bar;`},
					},
				},
			},
		},
		{
			Code:    `declare const foo: string | undefined; foo && foo.length;`,
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: string | undefined; foo?.length;`},
					},
				},
			},
		},
		{
			Code:   `declare const foo: { bar: string } | null; foo && foo.bar;`,
			Output: []string{`declare const foo: { bar: string } | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
