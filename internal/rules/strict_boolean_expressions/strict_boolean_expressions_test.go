// Code generated from strict-boolean-expressions.test.ts - DO NOT EDIT.

package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// ========================================
		// BOOLEAN IN BOOLEAN CONTEXT
		// ========================================
		{Code: `true ? 'a' : 'b';`},
		{Code: `if (false) {}`},
		{Code: `while (true) {}`},
		{Code: `for (; false; ) {}`},
		{Code: `!true;`},
		{Code: `false || 123;`},
		{Code: `true && 'foo';`},
		{Code: `!(false || true);`},
		{Code: `true && false ? true : false;`},
		{Code: `(false && true) || false;`},
		{Code: `(false && true) || [];`},
		{Code: `(false && 1) || (true && 2);`},
		{Code: `declare const x: boolean; if (x) {}`},
		{Code: `(x: boolean) => !x;`},
		{Code: `<T extends boolean>(x: T) => (x ? 1 : 0);`},
		{Code: `declare const x: never; if (x) {}`},

		// ========================================
		// STRING IN BOOLEAN CONTEXT
		// ========================================
		{Code: `if ('') {}`},
		{Code: `while ('x') {}`},
		{Code: `for (; ''; ) {}`},
		{Code: `('' && '1') || x;`},
		{Code: `declare const x: string; if (x) {}`},
		{Code: `(x: string) => !x;`},
		{Code: `<T extends string>(x: T) => (x ? 1 : 0);`},

		// ========================================
		// NUMBER IN BOOLEAN CONTEXT
		// ========================================
		{Code: `if (0) {}`},
		{Code: `while (1n) {}`},
		{Code: `for (; Infinity; ) {}`},
		{Code: `(0 / 0 && 1 + 2) || x;`},
		{Code: `declare const x: number; if (x) {}`},
		{Code: `(x: bigint) => !x;`},
		{Code: `<T extends number>(x: T) => (x ? 1 : 0);`},

		// ========================================
		// NULLABLE OBJECT IN BOOLEAN CONTEXT
		// ========================================
		{Code: `declare const x: null | object; if (x) {}`},
		{Code: `(x?: { a: any }) => !x;`},
		{Code: `<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);`},

		// ========================================
		// NULLABLE BOOLEAN IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: boolean | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `(x?: boolean) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `<T extends boolean | null>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},
		{
			Code: `declare const test?: boolean; if (test ?? false) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
			},
		},

		// ========================================
		// NULLABLE STRING IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: string | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `(x?: string) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `<T extends string | null>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `'string' != null ? 'asd' : '';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},

		// ========================================
		// NULLABLE NUMBER IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: number | undefined; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `(x?: number) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `<T extends number | null>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},
		{
			Code: `declare const x: bigint | undefined; if (x ?? 0n) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},

		// ========================================
		// ANY TYPE IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: any; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `(x?: any) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `<T extends any>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `const foo: undefined | any = 0; if (foo) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},

		// ========================================
		// UNKNOWN TYPE IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: unknown; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},
		{
			Code: `<T extends unknown>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(true),
			},
		},

		// ========================================
		// NULLABLE ENUM IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `
				enum E { A, B }
				declare const x: E | null;
				if (x) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(true),
			},
		},
		{
			Code: `
				enum E { A = 'a', B = 'b' }
				declare const x: E | undefined;
				if (x) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(true),
			},
		},

		// ========================================
		// ALLOW RULE TO RUN WITHOUT STRICT NULL CHECKS
		// ========================================
		{
			Code: `if (true) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(true),
			},
		},

		// ========================================
		// COMPLEX BOOLEAN EXPRESSIONS
		// ========================================
		{Code: `(x?: boolean) => x ?? false ? true : false;`},
		{Code: `<T extends boolean | null>(x: T) => x ?? false;`},
		{
			Code: `(x?: string) => x ?? false;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableString: utils.Ref(true),
			},
		},
		{
			Code: `(x?: number) => x ?? false;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableNumber: utils.Ref(true),
			},
		},

		// ========================================
		// ASSERT FUNCTIONS
		// ========================================
		{Code: `
			declare function assert(condition: unknown): asserts condition;
			declare const x: string | null;
			assert(x);
		`},
		{Code: `
			declare function assert(a: number, b: unknown): asserts a;
			declare const nullableString: string | null;
			declare const boo: boolean;
			assert(boo, nullableString);
		`},
		{Code: `
			declare function assert(a: boolean, b: unknown): asserts b is string;
			declare const nullableString: string | null;
			declare const boo: boolean;
			assert(boo, nullableString);
		`},
		{Code: `
			declare function assert(a: number, b: unknown): asserts b;
			declare const nullableString: string | null;
			declare const boo: boolean;
			assert(nullableString, boo);
		`},
		{Code: `
			declare function assert(a: number, b: unknown): asserts b;
			declare const nullableString: string | null;
			declare const boo: boolean;
			assert(...nullableString, nullableString);
		`},
		{Code: `
			declare function assert(
				this: object,
				a: number,
				b?: unknown,
				c?: unknown,
			): asserts c;
			declare const nullableString: string | null;
			declare const foo: number;
			const o: { assert: typeof assert } = { assert };
			o.assert(foo, nullableString);
		`},
		{Code: `
			declare function assert(x: unknown): x is string;
			declare const nullableString: string | null;
			assert(nullableString);
		`},
		{Code: `
			class ThisAsserter {
				assertThis(this: unknown, arg2: unknown): asserts this {}
			}
			declare const lol: string | number | unknown | null;
			const thisAsserter: ThisAsserter = new ThisAsserter();
			thisAsserter.assertThis(lol);
		`},
		{Code: `
			function assert(this: object, a: number, b: unknown): asserts b;
			function assert(a: bigint, b: unknown): asserts b;
			function assert(this: object, a: string, two: string): asserts two;
			function assert(
				this: object,
				a: string,
				assertee: string,
				c: bigint,
				d: object,
			): asserts assertee;
			function assert(...args: any[]): void;
			function assert(...args: any[]) {
				throw new Error('lol');
			}
			declare const nullableString: string | null;
			assert(3 as any, nullableString);
		`},
		{Code: `
			declare const assert: any;
			declare const nullableString: string | null;
			assert(nullableString);
		`},

		// ========================================
		// ARRAY PREDICATE FUNCTIONS
		// ========================================
		{Code: `['one', 'two', ''].some(x => x);`},
		{Code: `['one', 'two', ''].find(x => x);`},
		{Code: `['one', 'two', ''].every(x => x);`},
		{Code: `['one', 'two', ''].filter((x): boolean => x);`},
		{Code: `['one', 'two', ''].filter(x => Boolean(x));`},
		{Code: `['one', 'two', ''].filter(function (x): boolean { if (x) { return true; } });`},
		{Code: `['one', 'two', ''].filter(function (x): boolean { if (x) { return true; } throw new Error('oops'); });`},
		{Code: `declare const predicate: (string) => boolean; ['one', 'two', ''].filter(predicate);`},
		{Code: `declare function notNullish<T>(x: T): x is NonNullable<T>; ['one', null].filter(notNullish);`},
		{Code: `declare function predicate(x: string | null): x is string; ['one', null].filter(predicate);`},
		{Code: `declare function predicate<T extends boolean>(x: string | null): T; ['one', null].filter(predicate);`},
		{Code: `declare function f(x: number): boolean; declare function f(x: string | null): boolean; [35].filter(f);`},

		// ========================================
		// SPECIAL CASES
		// ========================================
		{Code: `for (let x = 0; ; x++) { break; }`},
	}, []rule_tester.InvalidTestCase{
		// ========================================
		// NON-BOOLEAN IN RHS OF TEST EXPRESSION
		// ========================================
		{
			Code: `if (true && 1 + 1) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedNumber",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `while (false || 'a' + 'b') {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedString",
					Line:      1,
					Column:    17,
				},
			},
		},
		{
			Code: `(x: object) => (true || false || x ? true : false);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedObjectContext",
					Line:      1,
					Column:    34,
				},
			},
		},

		// ========================================
		// CHECK OUTERMOST OPERANDS
		// ========================================
		{
			Code: `if (('' && {}) || (0 && void 0)) { }`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 6},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 12},
				{MessageId: "unexpectedNumber", Line: 1, Column: 20},
				{MessageId: "unexpectedNullish", Line: 1, Column: 25},
			},
		},

		// ========================================
		// ARRAY PREDICATE WITH NON-BOOLEAN
		// ========================================
		{
			Code: `declare const array: string[]; array.some(x => x);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(true),
				AllowString:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// ========================================
		// BRANDED TYPES
		// ========================================
		{
			Code: `
				declare const foo: true & { __BRAND: 'Foo' };
				if (('' && foo) || (0 && void 0)) { }
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 3, Column: 6},
				{MessageId: "unexpectedNumber", Line: 3, Column: 21},
				{MessageId: "unexpectedNullish", Line: 3, Column: 26},
			},
		},
		{
			Code: `
				declare const foo: false & { __BRAND: 'Foo' };
				if (('' && {}) || (foo && void 0)) { }
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 3, Column: 6},
				{MessageId: "unexpectedObjectContext", Line: 3, Column: 12},
				{MessageId: "unexpectedNullish", Line: 3, Column: 27},
			},
		},

		// ========================================
		// LOGICAL OPERANDS FOR CONTROL FLOW
		// ========================================
		{
			Code: `'asd' && 123 && [] && null;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 1},
				{MessageId: "unexpectedNumber", Line: 1, Column: 10},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 17},
			},
		},
		{
			Code: `'asd' || 123 || [] || null;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 1},
				{MessageId: "unexpectedNumber", Line: 1, Column: 10},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 17},
			},
		},
		{
			Code: `let x = (1 && 'a' && null) || 0 || '' || {};`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 10},
				{MessageId: "unexpectedString", Line: 1, Column: 15},
				{MessageId: "unexpectedNullish", Line: 1, Column: 22},
				{MessageId: "unexpectedNumber", Line: 1, Column: 31},
				{MessageId: "unexpectedString", Line: 1, Column: 36},
			},
		},
		{
			Code: `return (1 || 'a' || null) && 0 && '' && {};`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 9},
				{MessageId: "unexpectedString", Line: 1, Column: 14},
				{MessageId: "unexpectedNullish", Line: 1, Column: 21},
				{MessageId: "unexpectedNumber", Line: 1, Column: 30},
				{MessageId: "unexpectedString", Line: 1, Column: 35},
			},
		},
		{
			Code: `console.log((1 && []) || ('a' && {}));`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 14},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 19},
				{MessageId: "unexpectedString", Line: 1, Column: 27},
			},
		},

		// ========================================
		// CONDITIONALS WITH ALL OPERANDS CHECKED
		// ========================================
		{
			Code: `if ((1 && []) || ('a' && {})) void 0;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 6},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 11},
				{MessageId: "unexpectedString", Line: 1, Column: 19},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 26},
			},
		},
		{
			Code: `let x = null || 0 || 'a' || [] ? {} : undefined;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 9},
				{MessageId: "unexpectedNumber", Line: 1, Column: 17},
				{MessageId: "unexpectedString", Line: 1, Column: 22},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 29},
			},
		},
		{
			Code: `return !(null || 0 || 'a' || []);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 10},
				{MessageId: "unexpectedNumber", Line: 1, Column: 18},
				{MessageId: "unexpectedString", Line: 1, Column: 23},
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 30},
			},
		},

		// ========================================
		// NULLISH VALUES IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `null || {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 1},
			},
		},
		{
			Code: `undefined && [];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 1},
			},
		},
		{
			Code: `declare const x: null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
		{
			Code: `(x: undefined) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 20},
			},
		},
		{
			Code: `<T extends null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 40},
			},
		},
		{
			Code: `<T extends null>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 28},
			},
		},
		{
			Code: `<T extends undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullish", Line: 1, Column: 33},
			},
		},

		// ========================================
		// OBJECT IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `[] || 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 1},
			},
		},
		{
			Code: `({}) && 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 2},
			},
		},
		{
			Code: `declare const x: symbol; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `(x: () => void) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 21},
			},
		},
		{
			Code: `<T extends object>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 30},
			},
		},
		{
			Code: `<T extends Object | Function>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 41},
			},
		},
		{
			Code: `<T extends { a: number }>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 37},
			},
		},
		{
			Code: `<T extends () => void>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 34},
			},
		},

		// ========================================
		// STRING IN BOOLEAN CONTEXT WITH ALLOWSTRING: FALSE
		// ========================================
		{
			Code: `while ('') {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 8},
			},
		},
		{
			Code: `for (; 'foo'; ) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 8},
			},
		},
		{
			Code: `declare const x: string; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},
		{
			Code: `(x: string) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 17},
			},
		},
		{
			Code: `<T extends string>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1, Column: 30},
			},
		},

		// ========================================
		// NUMBER IN BOOLEAN CONTEXT WITH ALLOWNUMBER: FALSE
		// ========================================
		{
			Code: `while (0n) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 8},
			},
		},
		{
			Code: `for (; 123; ) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 8},
			},
		},
		{
			Code: `declare const x: number; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `(x: bigint) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 17},
			},
		},
		{
			Code: `<T extends number>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 30},
			},
		},
		{
			Code: `![]['length'];`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 2},
			},
		},
		{
			Code: `declare const a: any[] & { notLength: number }; if (a.notLength) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},

		// ========================================
		// ARRAY.LENGTH IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `if (![].length) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 6},
			},
		},
		{
			Code: `(a: number[]) => a.length && '...';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 18},
			},
		},
		{
			Code: `<T extends unknown[]>(...a: T) => a.length || 'empty';`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 35},
			},
		},

		// ========================================
		// MIXED STRING | NUMBER VALUE IN BOOLEAN CONTEXT
		// ========================================
		{
			Code: `declare const x: string | number; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(true),
				AllowString: utils.Ref(true),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1},
			},
		},
		{
			Code: `(x: bigint | string) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(true),
				AllowString: utils.Ref(true),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1, Column: 26},
			},
		},
		{
			Code: `<T extends number | bigint | string>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(true),
				AllowString: utils.Ref(true),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 1, Column: 48},
			},
		},

		// ========================================
		// NULLABLE BOOLEAN WITHOUT OPTION
		// ========================================
		{
			Code: `declare const x: boolean | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `(x?: boolean) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1, Column: 19},
			},
		},
		{
			Code: `<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1, Column: 50},
			},
		},

		// ========================================
		// NULLABLE OBJECT WITHOUT OPTION
		// ========================================
		{
			Code: `declare const x: object | null; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `(x?: { a: number }) => !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1, Column: 25},
			},
		},
		{
			Code: `<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1, Column: 45},
			},
		},

		// ========================================
		// NULLABLE STRING WITHOUT OPTION
		// ========================================
		{
			Code: `declare const x: string | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `(x?: string) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1, Column: 18},
			},
		},
		{
			Code: `<T extends string | null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1, Column: 49},
			},
		},
		{
			Code: `function foo(x: '' | 'bar' | null) { if (!x) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},

		// ========================================
		// NULLABLE NUMBER WITHOUT OPTION
		// ========================================
		{
			Code: `declare const x: number | null; if (x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `(x?: number) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1, Column: 18},
			},
		},
		{
			Code: `<T extends number | null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1, Column: 49},
			},
		},
		{
			Code: `function foo(x: 0 | 1 | null) { if (!x) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},

		// ========================================
		// NULLABLE ENUM WITHOUT OPTION
		// ========================================
		{
			Code: `
				enum ExampleEnum { This = 0, That = 1 }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = 0, That = 1 }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This, That }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = '', That = 'a' }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = '', That = 0 }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = 'one', That = 'two' }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 4},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = 1, That = 2 }
				const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
				if (!theEnum) {}
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 4},
			},
		},

		// ========================================
		// NULLABLE MIXED ENUM WITHOUT OPTION
		// ========================================
		{
			Code: `
				enum ExampleEnum { This = 0, That = 'one' }
				(value?: ExampleEnum) => (value ? 1 : 0);
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 3},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = '', That = 1 }
				(value?: ExampleEnum) => (!value ? 1 : 0);
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 3},
			},
		},
		{
			Code: `
				enum ExampleEnum { This = 'this', That = 1 }
				(value?: ExampleEnum) => (value ? 1 : 0);
			`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedMixedCondition", Line: 3},
			},
		},

		// ========================================
		// ANY WITHOUT OPTION
		// ========================================
		{
			Code: `if ((Boolean(x) || {}) || (typeof x === 'string' && x)) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1, Column: 20},
				{MessageId: "unexpectedString", Line: 1, Column: 53},
			},
		},
		{
			Code: `if (1) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1, Column: 5},
			},
		},

		// ========================================
		// ASSERT FUNCTIONS
		// ========================================
		{
			Code: `
				declare function assert(a: boolean, b: unknown): asserts b;
				declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
				declare const nullableString: string | null;
				declare const boo: boolean;
				assert(boo, nullableString);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 6},
			},
		},

		// ========================================
		// ARRAY FILTER PREDICATES
		// ========================================
		{
			Code: `declare const nullOrBool: boolean | null; [true, false, null].filter(x => nullOrBool);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableBoolean: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableBoolean", Line: 1},
			},
		},
		{
			Code: `declare const nullOrString: string | null; ['', 'foo', null].filter(x => nullOrString);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString", Line: 1},
			},
		},
		{
			Code: `declare const nullOrNumber: number | null; [0, null].filter(x => nullOrNumber);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 1},
			},
		},
		{
			Code: `const objectValue: object = {}; [{ a: 0 }, {}].filter(x => objectValue);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `const objectValue: object = {}; [{ a: 0 }, {}].filter(x => { return objectValue; });`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedObjectContext", Line: 1},
			},
		},
		{
			Code: `declare const nullOrObject: object | null; [{ a: 0 }, null].filter(x => nullOrObject);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableObject", Line: 1},
			},
		},
		{
			Code: `const numbers: number[] = [1]; [1, 2].filter(x => numbers.length);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `const numberValue: number = 1; [1, 2].filter(x => numberValue);`,
			Options: StrictBooleanExpressionsOptions{
				AllowNumber: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNumber", Line: 1},
			},
		},
		{
			Code: `const stringValue: string = 'hoge'; ['hoge', 'foo'].filter(x => stringValue);`,
			Options: StrictBooleanExpressionsOptions{
				AllowString: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
			},
		},

		// ========================================
		// UNKNOWN TYPE WITHOUT OPTION
		// ========================================
		{
			Code: `declare const x: unknown; if (x) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: unknown; x ? 'a' : 'b';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: unknown; x && 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: unknown; x || 'a';`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
		{
			Code: `declare const x: unknown; !x;`,
			Options: StrictBooleanExpressionsOptions{
				AllowAny: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAny", Line: 1},
			},
		},
	})
}