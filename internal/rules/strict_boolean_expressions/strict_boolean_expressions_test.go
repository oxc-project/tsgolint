package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&StrictBooleanExpressionsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `true ? 'a' : 'b';`,
			},
			{
				Code: `
	if (false) {
	}
    `,
			},
			{
				Code: `while (true) {}`,
			},
			{
				Code: `for (; false; ) {}`,
			},
			{
				Code: `!true;`,
			},
			{
				Code: `false || 123;`,
			},
			{
				Code: `true && 'foo';`,
			},
			{
				Code: `!(false || true);`,
			},
			{
				Code: `true && false ? true : false;`,
			},
			{
				Code: `(false && true) || false;`,
			},
			{
				Code: `(false && true) || [];`,
			},
			{
				Code: `(false && 1) || (true && 2);`,
			},
			{
				Code:
				 `
declare const x: boolean;
if (x) {
}
    `,
			},
		{
				Code: `(x: boolean) => !x;`,
			},
		{
				Code: `<T extends boolean>(x: T) => (x ? 1 : 0);`,
			},
		{
				Code: `
declare const x: never;
if (x) {
}
    `,
			},
		{
				Code: `
if ('') {
}
    `,
			},
		{
				Code: `while ('x') {}`,
			},
		{
				Code: `for (; ''; ) {}`,
			},
		{
				Code: `('' && '1') || x;`,
			},
		{
				Code: `
declare const x: string;
if (x) {
}
    `,
			},
		{
				Code: `(x: string) => !x;`,
			},
		{
				Code: `<T extends string>(x: T) => (x ? 1 : 0);`,
			},
		{
				Code: `
if (0) {
}
    `,
			},
		{
				Code: `while (1n) {}`,
			},
		{
				Code: `for (; Infinity; ) {}`,
			},
		{
				Code: `(0 / 0 && 1 + 2) || x;`,
			},
		{
				Code: `
declare const x: number;
if (x) {
}
    `,
			},
		{
				Code: `(x: bigint) => !x;`,
			},
		{
				Code: `<T extends number>(x: T) => (x ? 1 : 0);`,
			},
		{
				Code: `
declare const x: null | object;
if (x) {
}
    `,
			},
		{
				Code: `(x?: { a: any }) => !x;`,
			},
		{
				Code: `<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);`,
			},
		{
				Code: `
        declare const x: boolean | null;
        if (x) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(true)}},
		{
				Code: `
        (x?: boolean) => !x;
      `, Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(true)}},
		{
				Code: `
        <T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(true)}},
		{
				Code: `
        const a: (undefined | boolean | null)[] = [true, undefined, null];
        a.some(x => x);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(true)}},
		{
				Code: `
        declare const x: string | null;
        if (x) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableString: utils.Ref(true)}},
		{
				Code: `
        (x?: string) => !x;
      `, Options: StrictBooleanExpressionsOptions{AllowNullableString: utils.Ref(true)},
	},
		{
				Code: `
        <T extends string | null | undefined>(x: T) => (x ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableString: utils.Ref(true)}},
		{
				Code: `
        declare const x: number | null;
        if (x) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableNumber: utils.Ref(true)}},
		{
				Code: `
        (x?: number) => !x;
      `, Options: StrictBooleanExpressionsOptions{AllowNullableNumber: utils.Ref(true)}},
		{
				Code: `
        <T extends number | null | undefined>(x: T) => (x ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableNumber: utils.Ref(true)}},
		{
				Code: `
        declare const arrayOfArrays: (null | unknown[])[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array?.length);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableNumber: utils.Ref(true)}},
		{
				Code: `
        declare const x: any;
        if (x) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowAny: utils.Ref(true)}},
		{
				Code: `
        x => !x;
      `, Options: StrictBooleanExpressionsOptions{AllowAny: utils.Ref(true)}},
		{
				Code: `
        <T extends any>(x: T) => (x ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowAny: utils.Ref(true)}},
		{
				Code: `
        declare const arrayOfArrays: any[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `, Options: StrictBooleanExpressionsOptions{AllowAny: utils.Ref(true)}},
		{
				Code: `
        1 && true && 'x' && {};
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        let x = 0 || false || '' || null;
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        if (1 && true && 'x') void 0;
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        if (0 || false || '') void 0;
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        1 && true && 'x' ? {} : null;
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        0 || false || '' ? null : {};
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)}},
		{
				Code: `
        declare const arrayOfArrays: string[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `, Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(true)}},
		{
				Code: `
        declare const arrayOfArrays: number[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true)}},
		{
				Code: `
        declare const arrayOfArrays: (null | object)[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 0,
          That = 1,
        }
        const rand = Math.random();
        let theEnum: ExampleEnum | null = null;
        if (rand < 0.3) {
          theEnum = ExampleEnum.This;
        }
        if (theEnum) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 0,
          That = 1,
        }
        const rand = Math.random();
        let theEnum: ExampleEnum | null = null;
        if (rand < 0.3) {
          theEnum = ExampleEnum.This;
        }
        if (!theEnum) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 1,
          That = 2,
        }
        const rand = Math.random();
        let theEnum: ExampleEnum | null = null;
        if (rand < 0.3) {
          theEnum = ExampleEnum.This;
        }
        if (!theEnum) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 'one',
          That = 'two',
        }
        const rand = Math.random();
        let theEnum: ExampleEnum | null = null;
        if (rand < 0.3) {
          theEnum = ExampleEnum.This;
        }
        if (!theEnum) {
        }
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 0,
          That = 'one',
        }
        (value?: ExampleEnum) => (value ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = '',
          That = 1,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = 'this',
          That = 1,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = '',
          That = 0,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
        enum ExampleEnum {
          This = '',
          That = 0,
        }
        declare const arrayOfArrays: (ExampleEnum | null)[];
        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
      `, Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(true)}},
		{
				Code: `
declare const x: string[] | null;
// eslint-disable-next-line
if (x) {
}
      `, Options: StrictBooleanExpressionsOptions{AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(true)}},
		{
				Code: `
function f(arg: 'a' | null) {
  if (arg) console.log(arg);
}
    `,
			},
		{
				Code: `
function f(arg: 'a' | 'b' | null) {
  if (arg) console.log(arg);
}
    `,
			},
		{
				Code: `
declare const x: 1 | null;
declare const y: 1;
if (x) {
}
if (y) {
}
      `, Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true)}},
		{
				Code: `
function f(arg: 1 | null) {
  if (arg) console.log(arg);
}
    `,
			},
		{
				Code: `
function f(arg: 1 | 2 | null) {
  if (arg) console.log(arg);
}
    `,
			},
		{
				Code: `
interface Options {
  readonly enableSomething?: true;
}

function f(opts: Options): void {
  if (opts.enableSomething) console.log('Do something');
}
    `,
			},
		{
				Code: `
declare const x: true | null;
if (x) {
}
    `,
			},
		{
				Code: `
declare const x: 'a' | null;
declare const y: 'a';
if (x) {
}
if (y) {
}
      `, Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(true)}},
		{
				Code: `
declare const foo: boolean & { __BRAND: 'Foo' };
if (foo) {
}
    `,
			},
		{
				Code: `
declare const foo: true & { __BRAND: 'Foo' };
if (foo) {
}
    `,
			},
		{
				Code: `
declare const foo: false & { __BRAND: 'Foo' };
if (foo) {
}
    `,
			},
		{
				Code: `
declare function assert(a: number, b: unknown): asserts a;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
    `,
			},
		{
				Code: `
declare function assert(a: boolean, b: unknown): asserts b is string;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
    `,
			},
		{
				Code: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(nullableString, boo);
    `,
			},
		{
				Code: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(...nullableString, nullableString);
    `,
			},
		{
				Code: `
declare function assert(
  this: object,
  a: number,
  b?: unknown,
  c?: unknown,
): asserts c;
declare const nullableString: string | null;
declare const foo: number;
const o: { assert: typeof assert } = {
  assert,
};
o.assert(foo, nullableString);
    `,
			},
		{
				Code: `
declare function assert(x: unknown): x is string;
declare const nullableString: string | null;
assert(nullableString);
      `,
			},
		{
				Code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

declare const lol: string | number | unknown | null;

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(lol);
      `,
			},
		{
				Code: `
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
      `,
			},
		{
				Code: `
declare const assert: any;
declare const nullableString: string | null;
assert(nullableString);
    `,
			},
		{
				Code: `
      for (let x = 0; ; x++) {
        break;
      }
    `,
			},
		{
				Code: `
[true, false].some(function (x) {
  return x;
});
    `,
			},
		{
				Code: `
[true, false].some(function check(x) {
  return x;
});
    `,
			},
		{
				Code: `
[true, false].some(x => {
  return x;
});
    `,
			},
		{
				Code: `
[1, null].filter(function (x) {
  return x != null;
});
    `,
			},
		{
				Code: `
['one', 'two', ''].filter(function (x) {
  return !!x;
});
    `,
			},
		{
				Code: `
['one', 'two', ''].filter(function (x): boolean {
  return !!x;
});
    `,
			},
		{
				Code: `
['one', 'two', ''].filter(function (x): boolean {
  if (x) {
    return true;
  }
});
    `,
			},
		{
				Code: `
['one', 'two', ''].filter(function (x): boolean {
  if (x) {
    return true;
  }

  throw new Error('oops');
});
    `,
			},
		{
				Code: `
declare const predicate: (string) => boolean;
['one', 'two', ''].filter(predicate);
    `,
			},
		{
				Code: `
declare function notNullish<T>(x: T): x is NonNullable<T>;
['one', null].filter(notNullish);
    `,
			},
		{
				Code: `
declare function predicate(x: string | null): x is string;
['one', null].filter(predicate);
    `,
			},
		{
				Code: `
declare function predicate<T extends boolean>(x: string | null): T;
['one', null].filter(predicate);
    `,
			},
		{
				Code: `
declare function f(x: number): boolean;
declare function f(x: string | null): boolean;

[35].filter(f);
    `,
			},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
if (true && 1 + 1) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareZero",
// 							Output: `
// if (true && ((1 + 1) !== 0)) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareNaN",
// 							Output: `
// if (true && (!Number.isNaN((1 + 1)))) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// if (true && (Boolean((1 + 1)))) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `while (false || 'a' + 'b') {}`,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `while (false || (('a' + 'b').length > 0)) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `while (false || (('a' + 'b') !== "")) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `while (false || (Boolean(('a' + 'b')))) {}`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `(x: object) => (true || false || x ? true : false);`,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `if (('' && {}) || (0 && void 0)) { }`,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `if (((''.length > 0) && {}) || (0 && void 0)) { }`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `if ((('' !== "") && {}) || (0 && void 0)) { }`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `if (((Boolean('')) && {}) || (0 && void 0)) { }`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `if (('' && {}) || ((0 !== 0) && void 0)) { }`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `if (('' && {}) || ((!Number.isNaN(0)) && void 0)) { }`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `if (('' && {}) || ((Boolean(0)) && void 0)) { }`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `
        declare const array: string[];
        array.some(x => x);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(true), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareStringLength",
	// 						Output: `
    //     declare const array: string[];
    //     array.some(x => x.length > 0);
    //   `,
	// 					},
	// 					{
	// 						MessageId: "conditionFixCompareEmptyString",
	// 						Output: `
    //     declare const array: string[];
    //     array.some(x => x !== "");
    //   `,
	// 					},
	// 					{
	// 						MessageId: "conditionFixCastBoolean",
	// 						Output: `
    //     declare const array: string[];
    //     array.some(x => Boolean(x));
    //   `,
	// 					},
	// 					{
	// 						MessageId: "explicitBooleanReturnType",
	// 						Output: `
    //     declare const array: string[];
    //     array.some((x): boolean => x);
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
declare const foo: true & { __BRAND: 'Foo' };
if (('' && foo) || (0 && void 0)) { }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareStringLength",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if (((''.length > 0) && foo) || (0 && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareEmptyString",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if ((('' !== "") && foo) || (0 && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if (((Boolean('')) && foo) || (0 && void 0)) { }
//       `,
// 						},
// 					},
				},
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareZero",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if (('' && foo) || ((0 !== 0) && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareNaN",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if (('' && foo) || ((!Number.isNaN(0)) && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const foo: true & { __BRAND: 'Foo' };
// if (('' && foo) || ((Boolean(0)) && void 0)) { }
//       `,
// 						},
// 					},
				},
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `
declare const foo: false & { __BRAND: 'Foo' };
if (('' && {}) || (foo && void 0)) { }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareStringLength",
// 							Output: `
// declare const foo: false & { __BRAND: 'Foo' };
// if (((''.length > 0) && {}) || (foo && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareEmptyString",
// 							Output: `
// declare const foo: false & { __BRAND: 'Foo' };
// if ((('' !== "") && {}) || (foo && void 0)) { }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const foo: false & { __BRAND: 'Foo' };
// if (((Boolean('')) && {}) || (foo && void 0)) { }
//       `,
// 						},
// 					},
				},
				{
					MessageId: "conditionErrorObject",
				},
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `'asd' && 123 && [] && null;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `('asd'.length > 0) && 123 && [] && null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `('asd' !== "") && 123 && [] && null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(Boolean('asd')) && 123 && [] && null;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `'asd' && (123 !== 0) && [] && null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `'asd' && (!Number.isNaN(123)) && [] && null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `'asd' && (Boolean(123)) && [] && null;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `'asd' || 123 || [] || null;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `('asd'.length > 0) || 123 || [] || null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `('asd' !== "") || 123 || [] || null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(Boolean('asd')) || 123 || [] || null;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `'asd' || (123 !== 0) || [] || null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `'asd' || (!Number.isNaN(123)) || [] || null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `'asd' || (Boolean(123)) || [] || null;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `let x = (1 && 'a' && null) || 0 || '' || {};`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `let x = ((1 !== 0) && 'a' && null) || 0 || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `let x = ((!Number.isNaN(1)) && 'a' && null) || 0 || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = ((Boolean(1)) && 'a' && null) || 0 || '' || {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `let x = (1 && ('a'.length > 0) && null) || 0 || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `let x = (1 && ('a' !== "") && null) || 0 || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = (1 && (Boolean('a')) && null) || 0 || '' || {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorNullish",
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `let x = (1 && 'a' && null) || (0 !== 0) || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `let x = (1 && 'a' && null) || (!Number.isNaN(0)) || '' || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = (1 && 'a' && null) || (Boolean(0)) || '' || {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `let x = (1 && 'a' && null) || 0 || (''.length > 0) || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `let x = (1 && 'a' && null) || 0 || ('' !== "") || {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = (1 && 'a' && null) || 0 || (Boolean('')) || {};`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `return (1 || 'a' || null) && 0 && '' && {};`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `return ((1 !== 0) || 'a' || null) && 0 && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `return ((!Number.isNaN(1)) || 'a' || null) && 0 && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return ((Boolean(1)) || 'a' || null) && 0 && '' && {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `return (1 || ('a'.length > 0) || null) && 0 && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `return (1 || ('a' !== "") || null) && 0 && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return (1 || (Boolean('a')) || null) && 0 && '' && {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorNullish",
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `return (1 || 'a' || null) && (0 !== 0) && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `return (1 || 'a' || null) && (!Number.isNaN(0)) && '' && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return (1 || 'a' || null) && (Boolean(0)) && '' && {};`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `return (1 || 'a' || null) && 0 && (''.length > 0) && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `return (1 || 'a' || null) && 0 && ('' !== "") && {};`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return (1 || 'a' || null) && 0 && (Boolean('')) && {};`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `console.log((1 && []) || ('a' && {}));`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `console.log(((1 !== 0) && []) || ('a' && {}));`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `console.log(((!Number.isNaN(1)) && []) || ('a' && {}));`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `console.log(((Boolean(1)) && []) || ('a' && {}));`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `console.log((1 && []) || (('a'.length > 0) && {}));`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `console.log((1 && []) || (('a' !== "") && {}));`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `console.log((1 && []) || ((Boolean('a')) && {}));`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `if ((1 && []) || ('a' && {})) void 0;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					// MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `if (((1 !== 0) && []) || ('a' && {})) void 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `if (((!Number.isNaN(1)) && []) || ('a' && {})) void 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `if (((Boolean(1)) && []) || ('a' && {})) void 0;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `if ((1 && []) || (('a'.length > 0) && {})) void 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `if ((1 && []) || (('a' !== "") && {})) void 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `if ((1 && []) || ((Boolean('a')) && {})) void 0;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `let x = null || 0 || 'a' || [] ? {} : undefined;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `let x = null || (0 !== 0) || 'a' || [] ? {} : undefined;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `let x = null || (!Number.isNaN(0)) || 'a' || [] ? {} : undefined;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = null || (Boolean(0)) || 'a' || [] ? {} : undefined;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `let x = null || 0 || ('a'.length > 0) || [] ? {} : undefined;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `let x = null || 0 || ('a' !== "") || [] ? {} : undefined;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `let x = null || 0 || (Boolean('a')) || [] ? {} : undefined;`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `return !(null || 0 || 'a' || []);`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false), AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `return !(null || (0 !== 0) || 'a' || []);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `return !(null || (!Number.isNaN(0)) || 'a' || []);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return !(null || (Boolean(0)) || 'a' || []);`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `return !(null || 0 || ('a'.length > 0) || []);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `return !(null || 0 || ('a' !== "") || []);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `return !(null || 0 || (Boolean('a')) || []);`,
					// 	},
					// },
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `null || {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `undefined && [];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `
declare const x: null;
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `(x: undefined) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `<T extends null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `<T extends null>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `<T extends undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
				},
			},
		},
		{
			Code: `[] || 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `({}) && 'a';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `
declare const x: symbol;
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `(x: () => void) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `<T extends object>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `<T extends Object | Function>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `<T extends { a: number }>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `<T extends () => void>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `while ('') {}`,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `while (''.length > 0) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `while ('' !== "") {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `while (Boolean('')) {}`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `for (; 'foo'; ) {}`,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `for (; 'foo'.length > 0; ) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `for (; 'foo' !== ""; ) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `for (; Boolean('foo'); ) {}`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const x: string;
if (x) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareStringLength",
// 							Output: `
// declare const x: string;
// if (x.length > 0) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareEmptyString",
// 							Output: `
// declare const x: string;
// if (x !== "") {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const x: string;
// if (Boolean(x)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x: string) => !x;`,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `(x: string) => x.length === 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `(x: string) => x === "";`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(x: string) => !Boolean(x);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends string>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareStringLength",
					// 		Output: `<T extends string>(x: T) => ((x.length > 0) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareEmptyString",
					// 		Output: `<T extends string>(x: T) => ((x !== "") ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T extends string>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `while (0n) {}`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `while (0n !== 0) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `while (!Number.isNaN(0n)) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `while (Boolean(0n)) {}`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `for (; 123; ) {}`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `for (; 123 !== 0; ) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `for (; !Number.isNaN(123); ) {}`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `for (; Boolean(123); ) {}`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const x: number;
if (x) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareZero",
// 							Output: `
// declare const x: number;
// if (x !== 0) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareNaN",
// 							Output: `
// declare const x: number;
// if (!Number.isNaN(x)) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const x: number;
// if (Boolean(x)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x: bigint) => !x;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `(x: bigint) => x === 0;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `(x: bigint) => Number.isNaN(x);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(x: bigint) => !Boolean(x);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends number>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `<T extends number>(x: T) => ((x !== 0) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `<T extends number>(x: T) => ((!Number.isNaN(x)) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T extends number>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `![]['length']; // doesn't count as array.length when computed`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareZero",
					// 		Output: `[]['length'] === 0; // doesn't count as array.length when computed`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareNaN",
					// 		Output: `Number.isNaN([]['length']); // doesn't count as array.length when computed`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `!Boolean([]['length']); // doesn't count as array.length when computed`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const a: any[] & { notLength: number };
if (a.notLength) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareZero",
// 							Output: `
// declare const a: any[] & { notLength: number };
// if (a.notLength !== 0) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareNaN",
// 							Output: `
// declare const a: any[] & { notLength: number };
// if (!Number.isNaN(a.notLength)) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const a: any[] & { notLength: number };
// if (Boolean(a.notLength)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
if (![].length) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareArrayLengthZero",
// 							Output: `
// if ([].length === 0) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
(a: number[]) => a.length && '...';
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareArrayLengthNonzero",
// 							Output: `
// (a: number[]) => (a.length > 0) && '...';
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
<T extends unknown[]>(...a: T) => a.length || 'empty';
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareArrayLengthNonzero",
// 							Output: `
// <T extends unknown[]>(...a: T) => (a.length > 0) || 'empty';
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const x: string | number;
if (x) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `(x: bigint | string) => !x;`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `<T extends number | bigint | string>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(true), AllowString: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `
declare const x: boolean | null;
if (x) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixDefaultFalse",
// 							Output: `
// declare const x: boolean | null;
// if (x ?? false) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareTrue",
// 							Output: `
// declare const x: boolean | null;
// if (x === true) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x?: boolean) => !x;`,
			Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixDefaultFalse",
					// 		Output: `(x?: boolean) => !(x ?? false);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareFalse",
					// 		Output: `(x?: boolean) => x === false;`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixDefaultFalse",
					// 		Output: `<T extends boolean | null | undefined>(x: T) => ((x ?? false) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCompareTrue",
					// 		Output: `<T extends boolean | null | undefined>(x: T) => ((x === true) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const x: object | null;
if (x) {
}
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const x: object | null;
// if (x != null) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x?: { a: number }) => !x;`,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `(x?: { a: number }) => x == null;`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);`,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `<T extends {} | null | undefined>(x: T) => ((x != null) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const x: string | null;
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const x: string | null;
// if (x != null) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare const x: string | null;
// if (x ?? "") {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const x: string | null;
// if (Boolean(x)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x?: string) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `(x?: string) => x == null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixDefaultEmptyString",
					// 		Output: `(x?: string) => !(x ?? "");`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(x?: string) => !Boolean(x);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends string | null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `<T extends string | null | undefined>(x: T) => ((x != null) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixDefaultEmptyString",
					// 		Output: `<T extends string | null | undefined>(x: T) => ((x ?? "") ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T extends string | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
function foo(x: '' | 'bar' | null) {
  if (!x) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function foo(x: '' | 'bar' | null) {
//   if (x == null) {
//   }
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// function foo(x: '' | 'bar' | null) {
//   if (!(x ?? "")) {
//   }
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function foo(x: '' | 'bar' | null) {
//   if (!Boolean(x)) {
//   }
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const x: number | null;
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const x: number | null;
// if (x != null) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultZero",
// 							Output: `
// declare const x: number | null;
// if (x ?? 0) {
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const x: number | null;
// if (Boolean(x)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `(x?: number) => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `(x?: number) => x == null;`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixDefaultZero",
					// 		Output: `(x?: number) => !(x ?? 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `(x?: number) => !Boolean(x);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends number | null | undefined>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableNumber",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCompareNullish",
					// 		Output: `<T extends number | null | undefined>(x: T) => ((x != null) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixDefaultZero",
					// 		Output: `<T extends number | null | undefined>(x: T) => ((x ?? 0) ? 1 : 0);`,
					// 	},
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T extends number | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
function foo(x: 0 | 1 | null) {
  if (!x) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function foo(x: 0 | 1 | null) {
//   if (x == null) {
//   }
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultZero",
// 							Output: `
// function foo(x: 0 | 1 | null) {
//   if (!(x ?? 0)) {
//   }
// }
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function foo(x: 0 | 1 | null) {
//   if (!Boolean(x)) {
//   }
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 0,
          That = 1,
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 0,
    //       That = 1,
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum != null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 0,
          That = 1,
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 0,
    //       That = 1,
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This,
          That,
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This,
    //       That,
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = '',
          That = 'a',
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = '',
    //       That = 'a',
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = '',
          That = 0,
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = '',
    //       That = 0,
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 'one',
          That = 'two',
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 'one',
    //       That = 'two',
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 1,
          That = 2,
        }
        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
        if (!theEnum) {
        }
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 1,
    //       That = 2,
    //     }
    //     const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
    //     if (theEnum == null) {
    //     }
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 0,
          That = 'one',
        }
        (value?: ExampleEnum) => (value ? 1 : 0);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 0,
    //       That = 'one',
    //     }
    //     (value?: ExampleEnum) => ((value != null) ? 1 : 0);
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = '',
          That = 1,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = '',
    //       That = 1,
    //     }
    //     (value?: ExampleEnum) => ((value == null) ? 1 : 0);
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = 'this',
          That = 1,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = 'this',
    //       That = 1,
    //     }
    //     (value?: ExampleEnum) => ((value == null) ? 1 : 0);
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
        enum ExampleEnum {
          This = '',
          That = 0,
        }
        (value?: ExampleEnum) => (!value ? 1 : 0);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableEnum: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     enum ExampleEnum {
    //       This = '',
    //       That = 0,
    //     }
    //     (value?: ExampleEnum) => ((value == null) ? 1 : 0);
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// if (Boolean(x)) {
// }
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `x => !x;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `x => !(Boolean(x));`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T extends any>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T extends any>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `<T,>(x: T) => (x ? 1 : 0);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					// 	{
					// 		MessageId: "conditionFixCastBoolean",
					// 		Output: `<T,>(x: T) => ((Boolean(x)) ? 1 : 0);`,
					// 	},
					// },
				},
			},
		},
		{
			Code: `
declare const x: string[] | null;
if (x) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
				},
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `
        declare const obj: { x: number } | null;
        !obj ? 1 : 0
        !obj
        obj || 0
        obj && 1 || 0
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     declare const obj: { x: number } | null;
    //     (obj == null) ? 1 : 0
    //     !obj
    //     obj || 0
    //     obj && 1 || 0
    //   `,
	// 					},
	// 				},
				},
				{
					MessageId: "conditionErrorNullableObject",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     declare const obj: { x: number } | null;
    //     !obj ? 1 : 0
    //     obj == null
    //     obj || 0
    //     obj && 1 || 0
    //   `,
	// 					},
	// 				},
				},
				{
					MessageId: "conditionErrorNullableObject",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     declare const obj: { x: number } | null;
    //     !obj ? 1 : 0
    //     !obj
    //     ;(obj != null) || 0
    //     obj && 1 || 0
    //   `,
	// 					},
	// 				},
				},
				{
					MessageId: "conditionErrorNullableObject",
	// 				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
	// 					{
	// 						MessageId: "conditionFixCompareNullish",
	// 						Output: `
    //     declare const obj: { x: number } | null;
    //     !obj ? 1 : 0
    //     !obj
    //     obj || 0
    //     ;(obj != null) && 1 || 0
    //   `,
	// 					},
	// 				},
				},
			},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
declare const nullableString: string | null;
assert(nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare function assert(x: unknown): asserts x;
// declare const nullableString: string | null;
// assert(nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare function assert(x: unknown): asserts x;
// declare const nullableString: string | null;
// assert(nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare function assert(x: unknown): asserts x;
// declare const nullableString: string | null;
// assert(Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare function assert(a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare function assert(a: number, b: unknown): asserts b;
declare function assert(one: number, two: unknown): asserts two;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare function assert(one: number, two: unknown): asserts two;
// declare const nullableString: string | null;
// assert(foo, nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare function assert(one: number, two: unknown): asserts two;
// declare const nullableString: string | null;
// assert(foo, nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare function assert(a: number, b: unknown): asserts b;
// declare function assert(one: number, two: unknown): asserts two;
// declare const nullableString: string | null;
// assert(foo, Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare function assert(this: object, a: number, b: unknown): asserts b;
declare const nullableString: string | null;
assert(foo, nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare function assert(this: object, a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare function assert(this: object, a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare function assert(this: object, a: number, b: unknown): asserts b;
// declare const nullableString: string | null;
// assert(foo, Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
function asserts1(x: string | number | undefined): asserts x {}
function asserts2(x: string | number | undefined): asserts x {}

const maybeString = Math.random() ? 'string'.slice() : undefined;

const someAssert: typeof asserts1 | typeof asserts2 =
  Math.random() > 0.5 ? asserts1 : asserts2;

someAssert(maybeString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function asserts1(x: string | number | undefined): asserts x {}
// function asserts2(x: string | number | undefined): asserts x {}

// const maybeString = Math.random() ? 'string'.slice() : undefined;

// const someAssert: typeof asserts1 | typeof asserts2 =
//   Math.random() > 0.5 ? asserts1 : asserts2;

// someAssert(maybeString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// function asserts1(x: string | number | undefined): asserts x {}
// function asserts2(x: string | number | undefined): asserts x {}

// const maybeString = Math.random() ? 'string'.slice() : undefined;

// const someAssert: typeof asserts1 | typeof asserts2 =
//   Math.random() > 0.5 ? asserts1 : asserts2;

// someAssert(maybeString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function asserts1(x: string | number | undefined): asserts x {}
// function asserts2(x: string | number | undefined): asserts x {}

// const maybeString = Math.random() ? 'string'.slice() : undefined;

// const someAssert: typeof asserts1 | typeof asserts2 =
//   Math.random() > 0.5 ? asserts1 : asserts2;

// someAssert(Boolean(maybeString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
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

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
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
function assert(a: any, two: unknown, ...rest: any[]): asserts two;

function assert(...args: any[]) {
  throw new Error('lol');
}

declare const nullableString: string | null;
assert(3 as any, nullableString, 'more', 'args', 'afterwards');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;
// function assert(a: any, two: unknown, ...rest: any[]): asserts two;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, nullableString != null, 'more', 'args', 'afterwards');
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;
// function assert(a: any, two: unknown, ...rest: any[]): asserts two;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, nullableString ?? "", 'more', 'args', 'afterwards');
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function assert(this: object, a: number, b: unknown): asserts b;
// function assert(a: bigint, b: unknown): asserts b;
// function assert(this: object, a: string, two: string): asserts two;
// function assert(
//   this: object,
//   a: string,
//   assertee: string,
//   c: bigint,
//   d: object,
// ): asserts assertee;
// function assert(a: any, two: unknown, ...rest: any[]): asserts two;

// function assert(...args: any[]) {
//   throw new Error('lol');
// }

// declare const nullableString: string | null;
// assert(3 as any, Boolean(nullableString), 'more', 'args', 'afterwards');
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare function assert(a: boolean, b: unknown): asserts b;
declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
declare const nullableString: string | null;
declare const boo: boolean;
assert(boo, nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare function assert(a: boolean, b: unknown): asserts b;
// declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
// declare const nullableString: string | null;
// declare const boo: boolean;
// assert(boo, nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare function assert(a: boolean, b: unknown): asserts b;
// declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
// declare const nullableString: string | null;
// declare const boo: boolean;
// assert(boo, nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare function assert(a: boolean, b: unknown): asserts b;
// declare function assert({ a }: { a: boolean }, b: unknown): asserts b;
// declare const nullableString: string | null;
// declare const boo: boolean;
// assert(boo, Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
function assert(one: unknown): asserts one;
function assert(one: unknown, two: unknown): asserts two;
function assert(...args: unknown[]) {
  throw new Error('not implemented');
}
declare const nullableString: string | null;
assert(nullableString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// function assert(one: unknown): asserts one;
// function assert(one: unknown, two: unknown): asserts two;
// function assert(...args: unknown[]) {
//   throw new Error('not implemented');
// }
// declare const nullableString: string | null;
// assert(nullableString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// function assert(one: unknown): asserts one;
// function assert(one: unknown, two: unknown): asserts two;
// function assert(...args: unknown[]) {
//   throw new Error('not implemented');
// }
// declare const nullableString: string | null;
// assert(nullableString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// function assert(one: unknown): asserts one;
// function assert(one: unknown, two: unknown): asserts two;
// function assert(...args: unknown[]) {
//   throw new Error('not implemented');
// }
// declare const nullableString: string | null;
// assert(Boolean(nullableString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
['one', 'two', ''].find(x => {
  return x;
});
      `,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// ['one', 'two', ''].find((x): boolean => {
//   return x;
// });
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
['one', 'two', ''].find(x => {
  return;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// ['one', 'two', ''].find((x): boolean => {
//   return;
// });
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
['one', 'two', ''].findLast(x => {
  return undefined;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// ['one', 'two', ''].findLast((x): boolean => {
//   return undefined;
// });
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
['one', 'two', ''].find(x => {
  if (x) {
    return Math.random() > 0.5;
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// ['one', 'two', ''].find((x): boolean => {
//   if (x) {
//     return Math.random() > 0.5;
//   }
// });
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const predicate = (x: string) => {
  if (x) {
    return Math.random() > 0.5;
  }
};

['one', 'two', ''].find(predicate);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
				},
			},
		},
		{
			Code: `
[1, null].every(async x => {
  return x != null;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "predicateCannotBeAsync",
				},
			},
		},
		{
			Code: `
const predicate = async x => {
  return x != null;
};

[1, null].every(predicate);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
				},
			},
		},
		{
			Code: `
[1, null].every((x): boolean | number => {
  return x != null;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `
[1, null].every((x): boolean | undefined => {
  return x != null;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
				},
			},
		},
		{
			Code: `
[1, null].every((x, i) => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// [1, null].every((x, i): boolean => {});
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
[() => {}, null].every((x: () => void) => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// [() => {}, null].every((x: () => void): boolean => {});
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
[() => {}, null].every(function (x: () => void) {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// [() => {}, null].every(function (x: () => void): boolean {});
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
[() => {}, null].every(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullish",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// [() => {}, null].every((): boolean => {});
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare function f(x: number): string;
declare function f(x: string | null): boolean;

[35].filter(f);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `
declare function f(x: number): string;
declare function f(x: number | boolean): boolean;
declare function f(x: string | null): boolean;

[35].filter(f);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorOther",
				},
			},
		},
		{
			Code: `
declare function foo<T>(x: number): T;
[1, null].every(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
				},
			},
		},
		{
			Code: `
function foo<T extends number>(x: number): T {}
[1, null].every(foo);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
				},
			},
		},
		{
			Code: `
declare const nullOrString: string | null;
['one', null].filter(x => nullOrString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => nullOrString != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => nullOrString ?? "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => Boolean(nullOrString));
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter((x): boolean => nullOrString);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const nullOrString: string | null;
['one', null].filter(x => !nullOrString);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => nullOrString == null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultEmptyString",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => !(nullOrString ?? ""));
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const nullOrString: string | null;
// ['one', null].filter(x => !Boolean(nullOrString));
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const anyValue: any;
['one', null].filter(x => anyValue);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorAny",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const anyValue: any;
// ['one', null].filter(x => Boolean(anyValue));
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// declare const anyValue: any;
// ['one', null].filter((x): boolean => anyValue);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const nullOrBoolean: boolean | null;
[true, null].filter(x => nullOrBoolean);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableBoolean",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixDefaultFalse",
// 							Output: `
// declare const nullOrBoolean: boolean | null;
// [true, null].filter(x => nullOrBoolean ?? false);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareTrue",
// 							Output: `
// declare const nullOrBoolean: boolean | null;
// [true, null].filter(x => nullOrBoolean === true);
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// declare const nullOrBoolean: boolean | null;
// [true, null].filter((x): boolean => nullOrBoolean);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
enum ExampleEnum {
  This = 0,
  That = 1,
}
const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
[0, 1].filter(x => theEnum);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableEnum",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// enum ExampleEnum {
//   This = 0,
//   That = 1,
// }
// const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
// [0, 1].filter(x => theEnum != null);
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// enum ExampleEnum {
//   This = 0,
//   That = 1,
// }
// const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;
// [0, 1].filter((x): boolean => theEnum);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const nullOrNumber: number | null;
[0, null].filter(x => nullOrNumber);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const nullOrNumber: number | null;
// [0, null].filter(x => nullOrNumber != null);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixDefaultZero",
// 							Output: `
// declare const nullOrNumber: number | null;
// [0, null].filter(x => nullOrNumber ?? 0);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// declare const nullOrNumber: number | null;
// [0, null].filter(x => Boolean(nullOrNumber));
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// declare const nullOrNumber: number | null;
// [0, null].filter((x): boolean => nullOrNumber);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const objectValue: object = {};
[{ a: 0 }, {}].filter(x => objectValue);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// const objectValue: object = {};
// [{ a: 0 }, {}].filter((x): boolean => objectValue);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const objectValue: object = {};
[{ a: 0 }, {}].filter(x => {
  return objectValue;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorObject",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// const objectValue: object = {};
// [{ a: 0 }, {}].filter((x): boolean => {
//   return objectValue;
// });
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
declare const nullOrObject: object | null;
[{ a: 0 }, null].filter(x => nullOrObject);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNullableObject: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareNullish",
// 							Output: `
// declare const nullOrObject: object | null;
// [{ a: 0 }, null].filter(x => nullOrObject != null);
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// declare const nullOrObject: object | null;
// [{ a: 0 }, null].filter((x): boolean => nullOrObject);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const numbers: number[] = [1];
[1, 2].filter(x => numbers.length);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareArrayLengthNonzero",
// 							Output: `
// const numbers: number[] = [1];
// [1, 2].filter(x => numbers.length > 0);
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// const numbers: number[] = [1];
// [1, 2].filter((x): boolean => numbers.length);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const numberValue: number = 1;
[1, 2].filter(x => numberValue);
      `,
			Options: StrictBooleanExpressionsOptions{AllowNumber: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareZero",
// 							Output: `
// const numberValue: number = 1;
// [1, 2].filter(x => numberValue !== 0);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareNaN",
// 							Output: `
// const numberValue: number = 1;
// [1, 2].filter(x => !Number.isNaN(numberValue));
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// const numberValue: number = 1;
// [1, 2].filter(x => Boolean(numberValue));
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// const numberValue: number = 1;
// [1, 2].filter((x): boolean => numberValue);
//       `,
// 						},
// 					},
				},
			},
		},
		{
			Code: `
const stringValue: string = 'hoge';
['hoge', 'foo'].filter(x => stringValue);
      `,
			Options: StrictBooleanExpressionsOptions{AllowString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString",
// 					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
// 						{
// 							MessageId: "conditionFixCompareStringLength",
// 							Output: `
// const stringValue: string = 'hoge';
// ['hoge', 'foo'].filter(x => stringValue.length > 0);
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCompareEmptyString",
// 							Output: `
// const stringValue: string = 'hoge';
// ['hoge', 'foo'].filter(x => stringValue !== "");
//       `,
// 						},
// 						{
// 							MessageId: "conditionFixCastBoolean",
// 							Output: `
// const stringValue: string = 'hoge';
// ['hoge', 'foo'].filter(x => Boolean(stringValue));
//       `,
// 						},
// 						{
// 							MessageId: "explicitBooleanReturnType",
// 							Output: `
// const stringValue: string = 'hoge';
// ['hoge', 'foo'].filter((x): boolean => stringValue);
//       `,
// 						},
// 					},
				},
			},
		},
	})
}
