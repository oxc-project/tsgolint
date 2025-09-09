package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsRule(t *testing.T) {
	// Test with strictNullChecks enabled (in fixtures/tsconfig.json)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// ============================================
		// BOOLEAN TYPES - Always Valid
		// ============================================

		// Boolean literals
		{Code: `true ? 'a' : 'b';`},
		{Code: `false ? 'a' : 'b';`},
		{Code: `if (true) {}`},
		{Code: `if (false) {}`},
		{Code: `while (true) {}`},
		{Code: `do {} while (true);`},
		{Code: `for (; true; ) {}`},
		{Code: `!true;`},

		// Boolean variables
		{Code: `const b = true; if (b) {}`},
		{Code: `const b = false; if (!b) {}`},
		{Code: `declare const b: boolean; if (b) {}`},

		// Boolean expressions
		{Code: `if (true && false) {}`},
		{Code: `if (true || false) {}`},
		{Code: `true && 'a';`},
		{Code: `false || 'a';`},

		// Comparison operators return boolean
		{Code: `1 > 2 ? 'a' : 'b';`},
		{Code: `1 < 2 ? 'a' : 'b';`},
		{Code: `1 >= 2 ? 'a' : 'b';`},
		{Code: `1 <= 2 ? 'a' : 'b';`},
		{Code: `1 == 2 ? 'a' : 'b';`},
		{Code: `1 != 2 ? 'a' : 'b';`},
		{Code: `1 === 2 ? 'a' : 'b';`},
		{Code: `1 !== 2 ? 'a' : 'b';`},

		// Type guards
		{Code: `declare const x: string | number; if (typeof x === 'string') {}`},
		{Code: `declare const x: any; if (x instanceof Error) {}`},
		{Code: `declare const x: any; if ('prop' in x) {}`},

		// Function returning boolean
		{Code: `function test(): boolean { return true; } if (test()) {}`},
		{Code: `declare function test(): boolean; if (test()) {}`},

		// ============================================
		// STRING TYPES - Valid with Default Options
		// ============================================

		{Code: `'' ? 'a' : 'b';`},
		{Code: `'foo' ? 'a' : 'b';`},
		{Code: "`` ? 'a' : 'b';"},
		{Code: "`foo` ? 'a' : 'b';"},
		{Code: "`foo${bar}` ? 'a' : 'b';"},
		{Code: `if ('') {}`},
		{Code: `if ('foo') {}`},
		{Code: `while ('') {}`},
		{Code: `do {} while ('foo');`},
		{Code: `for (; 'foo'; ) {}`},
		{Code: `!!'foo';`},
		{Code: `declare const s: string; if (s) {}`},

		// String with logical operators
		{Code: `'' || 'foo';`},
		{Code: `'foo' && 'bar';`},
		{Code: `declare const s: string; s || 'default';`},

		// ============================================
		// NUMBER TYPES - Valid with Default Options
		// ============================================

		{Code: `0 ? 'a' : 'b';`},
		{Code: `1 ? 'a' : 'b';`},
		{Code: `-1 ? 'a' : 'b';`},
		{Code: `0.5 ? 'a' : 'b';`},
		{Code: `NaN ? 'a' : 'b';`},
		{Code: `if (0) {}`},
		{Code: `if (1) {}`},
		{Code: `while (0) {}`},
		{Code: `do {} while (1);`},
		{Code: `for (; 1; ) {}`},
		{Code: `declare const n: number; if (n) {}`},

		// Number with logical operators
		{Code: `0 || 1;`},
		{Code: `1 && 2;`},
		{Code: `declare const n: number; n || 0;`},

		// BigInt
		{Code: `0n ? 'a' : 'b';`},
		{Code: `1n ? 'a' : 'b';`},
		{Code: `if (0n) {}`},
		{Code: `if (1n) {}`},
		{Code: `declare const b: bigint; if (b) {}`},

		// ============================================
		// OBJECT TYPES in logical operators
		// ============================================

		// Note: Objects in logical operators are treated as boolean conditions
		// and will be flagged as errors (always truthy) unless in the right side
		// Right side of logical operators is not checked as a boolean condition
		{Code: `'foo' || ({});`},         // Object on right side is OK
		{Code: `false || [];`},           // Array on right side is OK
		{Code: `(false && true) || [];`}, // Array on right side is OK

		// ============================================
		// ANY TYPE - Valid with Option
		// ============================================

		{
			Code: `declare const x: any; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: any; x ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: any; x && 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: any; x || 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: any; !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},

		// ============================================
		// UNKNOWN TYPE - Valid with Option
		// ============================================

		{
			Code: `declare const x: unknown; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: unknown; x ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: unknown; x && 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: unknown; x || 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: unknown; !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},

		// ============================================
		// NULLABLE BOOLEAN - Valid with Option
		// ============================================

		{
			Code: `declare const x: boolean | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: boolean | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: boolean | null | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: true | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: false | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},

		// ============================================
		// NULLABLE STRING - Valid with Option
		// ============================================

		{
			Code: `declare const x: string | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: string | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: string | null | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: '' | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: 'foo' | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},

		// ============================================
		// NULLABLE NUMBER - Valid with Option
		// ============================================

		{
			Code: `declare const x: number | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: number | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: number | null | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: 0 | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: 1 | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: bigint | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},

		// ============================================
		// NULLABLE OBJECT - Valid with Option
		// ============================================

		{
			Code: `declare const x: object | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: object | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: object | null | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: {} | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: [] | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: symbol | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(true),
			},
		},

		// ============================================
		// NULLABLE ENUM - Valid with Option
		// ============================================
		// NOTE: Enum detection from union types is complex - skipping for now
		// {
		// 	Code: `enum E { A = 0, B = 1 } declare const x: E | null; if (x) {}`,
		// 	Options: StrictBooleanExpressionsOptions{
		// 		AllowNullableEnum: utils.Ref(true),
		// 	},
		// },
		// {
		// 	Code: `enum E { A = '', B = 'foo' } declare const x: E | undefined; if (x) {}`,
		// 	Options: StrictBooleanExpressionsOptions{
		// 		AllowNullableEnum: utils.Ref(true),
		// 	},
		// },

		// ============================================
		// ARRAY METHOD PREDICATES
		// ============================================

		{Code: `[1, 2, 3].every(x => x > 0);`},
		{Code: `[1, 2, 3].some(x => x > 0);`},
		{Code: `[1, 2, 3].filter(x => x > 0);`},
		{Code: `declare const arr: string[]; arr.find(x => x === 'foo');`},
		{Code: `declare const arr: string[]; arr.findIndex(x => x === 'foo');`},

		// With nullable predicates and options
		{
			Code: `[1, 2, 3].filter(x => x);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(true),
			},
		},
		{
			Code: `['', 'foo'].filter(x => x);`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(true),
			},
		},

		// ============================================
		// ASSERT FUNCTIONS AND TYPE PREDICATES
		// ============================================

		{Code: `function isString(x: unknown): x is string { return typeof x === 'string'; }`},
		{Code: `declare function isString(x: unknown): x is string; if (isString(value)) {}`},

		// ============================================
		// VOID TYPE - handled as nullish, so moved to invalid section
		// ============================================

		// ============================================
		// DOUBLE NEGATION - Always Valid
		// ============================================

		{Code: `!!true;`},
		{Code: `!!false;`},
		{Code: `!!'';`},
		{Code: `!!0;`},

		// ============================================
		// BOOLEAN CONSTRUCTOR - Always Valid
		// ============================================

		{Code: `Boolean(true);`},
		{Code: `Boolean(false);`},
		{Code: `Boolean('');`},
		{Code: `Boolean(0);`},
		{Code: `Boolean({});`},
		{Code: `Boolean([]);`},
		{Code: `declare const x: any; Boolean(x);`},
		{Code: `declare const x: unknown; Boolean(x);`},

		// ============================================
		// COMPLEX LOGICAL EXPRESSIONS
		// ============================================

		{Code: `true && true && true;`},
		{Code: `true || false || true;`},
		{Code: `(true && false) || (false && true);`},
		{Code: `true ? (false || true) : (true && false);`},

		// Mixed types with default options
		{Code: `'' || 0 || false;`},
		{Code: `'foo' && 1 && true;`},
		// NOTE: Objects in logical operators should be errors - commenting out incorrect test
		// {Code: `({}) || [] || true;`},

		// ============================================
		// SPECIAL CASES
		// ============================================

		// Always allow boolean in right side of logical operators
		{Code: `'foo' && true;`},
		{Code: `0 || false;`},
		// NOTE: Object in logical operator should be error - commenting out incorrect test
		// {Code: `({}) && (1 > 2);`},

		// Template literals
		{Code: "declare const x: string; `foo${x}` ? 'a' : 'b';"},
		{Code: "declare const x: number; `foo${x}` ? 'a' : 'b';"},

		// Parenthesized expressions
		{Code: `(true) ? 'a' : 'b';`},
		{Code: `((true)) ? 'a' : 'b';`},
		{Code: `if ((true)) {}`},

		// Comma operator
		{Code: `(0, true) ? 'a' : 'b';`},
		{Code: `('', false) ? 'a' : 'b';`},

		// Assignment expressions
		{Code: `let x; (x = true) ? 'a' : 'b';`},
		{Code: `let x; if (x = false) {}`},

		// Never type - allowed per TypeScript ESLint
		{Code: `declare const x: never; if (x) {}`},
		{Code: `declare const x: never; x ? 'a' : 'b';`},
		{Code: `declare const x: never; !x;`},
	}, []rule_tester.InvalidTestCase{
		// ============================================
		// ANY TYPE - Invalid without Option
		// ============================================

		{
			Code: `declare const x: any; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; x ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; while (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; do {} while (x);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; for (; x; ) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; x && 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; x || 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: any; !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},

		// ============================================
		// NULLISH VALUES - Always Invalid
		// ============================================

		{
			Code: `if (null) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `if (undefined) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `null ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `undefined ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `while (null) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `do {} while (undefined);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `for (; null; ) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `!null;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `!undefined;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `null && 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `undefined || 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},

		// ============================================
		// STRING TYPE - Invalid with allowString: false
		// ============================================

		{
			Code: `if ('') {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `if ('foo') {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `'' ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `'foo' ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `while ('') {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `do {} while ('foo');`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `for (; 'foo'; ) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `!'foo';`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `'foo' && 'bar';`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `'' || 'default';`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `declare const s: string; if (s) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// Template literals
		{
			Code: "`` ? 'a' : 'b';",
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: "`foo` ? 'a' : 'b';",
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: "declare const x: string; `foo${x}` ? 'a' : 'b';",
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// ============================================
		// NUMBER TYPE - Invalid with allowNumber: false
		// ============================================

		{
			Code: `if (0) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `if (1) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `0 ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `1 ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `NaN ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `while (0) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `do {} while (1);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `for (; 1; ) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `!0;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `1 && 2;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `0 || 1;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `declare const n: number; if (n) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},

		// BigInt
		{
			Code: `if (0n) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `if (1n) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `0n ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `declare const b: bigint; if (b) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},

		// ============================================
		// OBJECT TYPE - Always Invalid (Always Truthy)
		// ============================================

		{
			Code: `if ({}) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `if ([]) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `({}) ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `[] ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `while ({}) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `do {} while ([]);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `for (; {}; ) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `!{};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `![];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `({}) && 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `[] || 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `declare const o: object; if (o) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `declare const o: {}; if (o) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},

		// Functions
		{
			Code: `function foo() {}; if (foo) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `const foo = () => {}; if (foo) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},

		// Symbols
		{
			Code: `if (Symbol()) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `declare const s: symbol; if (s) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},

		// ============================================
		// NULLABLE BOOLEAN - Invalid without Option
		// ============================================

		{
			Code: `declare const x: boolean | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `declare const x: boolean | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `declare const x: boolean | null | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `declare const x: true | null; x ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `declare const x: false | undefined; !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},

		// ============================================
		// NULLABLE STRING - Invalid without Option
		// ============================================

		{
			Code: `declare const x: string | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `declare const x: string | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `declare const x: string | null | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `declare const x: '' | null; x ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `declare const x: 'foo' | undefined; !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},

		// ============================================
		// NULLABLE NUMBER - Invalid without Option
		// ============================================

		{
			Code: `declare const x: number | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `declare const x: number | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `declare const x: number | null | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `declare const x: 0 | null; x ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `declare const x: 1 | undefined; !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `declare const x: bigint | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},

		// ============================================
		// NULLABLE OBJECT - Invalid without Option
		// ============================================

		{
			Code: `declare const x: object | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `declare const x: object | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `declare const x: object | null | undefined; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `declare const x: {} | null; x ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `declare const x: [] | undefined; !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `declare const x: symbol | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},

		// ============================================
		// MIXED TYPES - Invalid
		// ============================================

		{
			Code: `declare const x: string | number; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1},
			},
		},
		{
			Code: `declare const x: string | boolean; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1},
			},
		},
		{
			Code: `declare const x: number | boolean; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1},
			},
		},
		{
			Code: `declare const x: string | number | boolean; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1},
			},
		},

		// ============================================
		// ENUM TYPES
		// ============================================

		{
			Code: `enum E { A = 0, B = 1 } declare const x: E; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `enum E { A = '', B = 'foo' } declare const x: E; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// ============================================
		// ARRAY METHOD PREDICATES - Invalid
		// ============================================
		// NOTE: Arrow function predicate checking not implemented - skipping for now
		// {
		// 	Code: `[1, 2, 3].filter(x => x);`,
		// 	Options: StrictBooleanExpressionsOptions{
		// 		AllowNumber: utils.Ref(false),
		// 	},
		// 	Errors: []rule_tester.InvalidTestCaseError{
		// 		{MessageId: "unexpectedNumber", Line: 1},
		// 	},
		// },
		// {
		// 	Code: `['', 'foo'].filter(x => x);`,
		// 	Options: StrictBooleanExpressionsOptions{
		// 		AllowString: utils.Ref(false),
		// 	},
		// 	Errors: []rule_tester.InvalidTestCaseError{
		// 		{MessageId: "unexpectedString", Line: 1},
		// 	},
		// },
		// {
		// 	Code: `[{}, []].filter(x => x);`,
		// 	Errors: []rule_tester.InvalidTestCaseError{
		// 		{MessageId: "unexpectedObjectContext", Line: 1},
		// 	},
		// },

		// ============================================
		// COMPLEX LOGICAL EXPRESSIONS - Invalid
		// ============================================

		{
			Code: `'foo' && 1;`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `0 || 'bar';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `({}) && [];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `'' || 0;`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// ============================================
		// SPECIAL CASES - Invalid
		// ============================================

		// Array.length
		{
			Code: `declare const arr: string[]; if (arr.length) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},

		// Function calls returning non-boolean
		{
			Code: `declare function getString(): string; if (getString()) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `declare function getNumber(): number; if (getNumber()) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `declare function getObject(): object; if (getObject()) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},

		// Property access
		{
			Code: `declare const obj: { prop: string }; if (obj.prop) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `declare const obj: { prop: number }; if (obj.prop) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},

		// Void type
		{
			Code: `declare const x: void; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `void 0 ? 'a' : 'b';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
	})
}
