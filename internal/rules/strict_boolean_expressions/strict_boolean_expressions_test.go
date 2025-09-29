package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsRule_Generated(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&StrictBooleanExpressionsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "true ? 'a' : 'b';",
			},
			{
				Code: `
    if (false) {
    }
    `,
			},
			{
				Code: "while (true) {}",
			},
			{
				Code: "for (; false; ) {}",
			},
			{
				Code: "!true;",
			},
			{
				Code: "false || 123;",
			},
			{
				Code: "true && 'foo';",
			},
			{
				Code: "!(false || true);",
			},
			{
				Code: "true && false ? true : false;",
			},
			{
				Code: "(false && true) || false;",
			},
			{
				Code: "(false && true) || [];",
			},
			{
				Code: "(false && 1) || (true && 2);",
			},
			{
				Code: `
    declare const x: boolean;
    if (x) {
    }
    `,
			},
			{
				Code: "(x: boolean) => !x;",
			},
			{
				Code: "<T extends boolean>(x: T) => (x ? 1 : 0);",
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
				Code: "while ('x') {}"},
			{
				Code: "for (; ''; ) {}"},
			{
				Code: "('' && '1') || x;"},
			{
				Code: `
    declare const x: string;
    if (x) {
    }
    `,
			},
			{
				Code: "(x: string) => !x;"},
			{
				Code: "<T extends string>(x: T) => (x ? 1 : 0);"},
			{
				Code: `
    if (0) {
    }
    `,
			},
			{
				Code: "while (1n) {}"},
			{
				Code: "for (; Infinity; ) {}"},
			{
				Code: "(0 / 0 && 1 + 2) || x;"},
			{
				Code: `
    declare const x: number;
    if (x) {
    }
    `,
			},
			{
				Code: "(x: bigint) => !x;"},
			{
				Code: "<T extends number>(x: T) => (x ? 1 : 0);"},
			{
				Code: `
    declare const x: null | object;
    if (x) {
    }
    `,
			},
			{
				Code: "(x?: { a: any }) => !x;"},
			{
				Code: "<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);"},
			{
				Code: `
      declare const x: boolean | null;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{

					AllowNullableBoolean: utils.Ref(true),
				},
			},
			{
				Code: `
      (x?: boolean) => !x;
      `,
				Options: StrictBooleanExpressionsOptions{

					AllowNullableBoolean: utils.Ref(true),
				},
			},
			{
				Code: `
      <T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(true),
				},
			},
			{
				Code: `
      const a: (undefined | boolean | null)[] = [true, undefined, null];
    a.some(x => x);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const x: string | null;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableString: utils.Ref(true),
				},
			},
			{
				Code: `
      (x?: string) => !x;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableString: utils.Ref(true),
				},
			},
			{
				Code: `
      <T extends string | null | undefined>(x: T) => (x ? 1 : 0);
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableString: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const x: number | null;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableNumber: utils.Ref(true),
				},
			},
			{
				Code: `
      (x?: number) => !x;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableNumber: utils.Ref(true),
				},
			},
			{
				Code: `
      <T extends number | null | undefined>(x: T) => (x ? 1 : 0);
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableNumber: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const arrayOfArrays: (null | unknown[])[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array?.length);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableNumber: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const x: any;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowAny: utils.Ref(true),
				},
			},
			{
				Code: `
      x => !x;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowAny: utils.Ref(true),
				},
			},
			{
				Code: `
      <T extends any>(x: T) => (x ? 1 : 0);
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowAny: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const arrayOfArrays: any[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowAny: utils.Ref(true),
				},
			},
			{
				Code: `
      1 && true && 'x' && {};
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      let x = 0 || false || '' || null;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      if (1 && true && 'x') void 0;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      if (0 || false || '') void 0;
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      1 && true && 'x' ? {} : null;
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      0 || false || '' ? null : {};
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const arrayOfArrays: string[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const arrayOfArrays: number[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const arrayOfArrays: (null | object)[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
				},
			},
			{
				Code: `
      enum ExampleEnum {
      This = '',
      That = 0,
    }
    declare const arrayOfArrays: (ExampleEnum | null)[];
    const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(true),
				},
			},
			{
				Code: `
      declare const x: string[] | null;
    // eslint-disable-next-line
    if (x) {
    }
    `,
				TSConfig: "tsconfig.unstrict.json",
				Options: StrictBooleanExpressionsOptions{
					AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true),
				},
			},
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
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(true),
				},
			},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 2},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "while (false || 'a' + 'b') {}",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "(x: object) => (true || false || x ? true : false);",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "if (('' && {}) || (0 && void 0)) { }",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedObject", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedNullish", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: `
      declare const array: string[];
    array.some(x => x);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(true), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString"},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, explicitBooleanReturnType */},
			{
				Code: `
      declare const foo: true & { __BRAND: 'Foo' };
    if (('' && foo) || (0 && void 0)) { }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 3}, {MessageId: "unexpectedNumber", Line: 3}, {MessageId: "unexpectedNullish", Line: 3},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: `
      declare const foo: false & { __BRAND: 'Foo' };
    if (('' && {}) || (foo && void 0)) { }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false), AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 3}, {MessageId: "unexpectedObject", Line: 3}, {MessageId: "unexpectedNullish", Line: 3},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "'asd' && 123 && [] && null;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedObject", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "'asd' || 123 || [] || null;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedObject", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "let x = (1 && 'a' && null) || 0 || '' || {};",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedNullish", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "return (1 || 'a' || null) && 0 && '' && {};",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedNullish", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "console.log((1 && []) || ('a' && {}));",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedObject", Line: 1}, {MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "if ((1 && []) || ('a' && {})) void 0;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedObject", Line: 1}, {MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedObject", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "let x = null || 0 || 'a' || [] ? {} : undefined;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedObject", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "return !(null || 0 || 'a' || []);",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false), AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1}, {MessageId: "unexpectedNumber", Line: 1}, {MessageId: "unexpectedString", Line: 1}, {MessageId: "unexpectedObject", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "null || {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: "undefined && [];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: `
      declare const x: null;
      if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 3},
				}},
			{
				Code: "(x: undefined) => !x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: "<T extends null | undefined>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: "<T extends null>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: "<T extends undefined>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 1},
				}},
			{
				Code: "[] || 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "({}) && 'a';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: `
      declare const x: symbol;
      if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 3},
				}},
			{
				Code: "(x: () => void) => !x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "<T extends object>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "<T extends Object | Function>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "<T extends { a: number }>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "<T extends () => void>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 1},
				}},
			{
				Code: "while ('') {}",
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "for (; 'foo'; ) {}",
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare const x: string;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 3},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "(x: string) => !x;",
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "<T extends string>(x: T) => (x ? 1 : 0);",
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 1},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean */},
			{
				Code: "while (0n) {}",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "for (; 123; ) {}",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: `
      declare const x: number;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 3},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "(x: bigint) => !x;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "<T extends number>(x: T) => (x ? 1 : 0);",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: "![]['length']; // doesn't count as array.length when computed",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 1},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: `
      declare const a: any[] & { notLength: number };
    if (a.notLength) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 3},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean */},
			{
				Code: `
      if (![].length) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 2},
				} /* Suggestions: conditionFixCompareArrayLengthZero */},
			{
				Code: `
      (a: number[]) => a.length && '...';
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 2},
				} /* Suggestions: conditionFixCompareArrayLengthNonzero */},
			{
				Code: `
      <T extends unknown[]>(...a: T) => a.length || 'empty';
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 2},
				} /* Suggestions: conditionFixCompareArrayLengthNonzero */},
			{
				Code: `
      declare const x: string | number;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 3},
				}},
			{
				Code: "(x: bigint | string) => !x;",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 1},
				}},
			{
				Code: "<T extends number | bigint | string>(x: T) => (x ? 1 : 0);",
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(true), AllowString: utils.Ref(true),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 1},
				}},
			{
				Code: `
      declare const x: boolean | null;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 3},
				} /* Suggestions: conditionFixDefaultFalse, conditionFixCompareTrue */},
			{
				Code: "(x?: boolean) => !x;",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 1},
				} /* Suggestions: conditionFixDefaultFalse, conditionFixCompareFalse */},
			{
				Code: "<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 1},
				} /* Suggestions: conditionFixDefaultFalse, conditionFixCompareTrue */},
			{
				Code: `
      declare const x: object | null;
      if (x) {
    }
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableObject", Line: 3},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: "(x?: { a: number }) => !x;",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableObject", Line: 1},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: "<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);",
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableObject", Line: 1},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      declare const x: string | null;
      if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: "(x?: string) => !x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 1},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: "<T extends string | null | undefined>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 1},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      function foo(x: '' | 'bar' | null) {
      if (!x) {
    }
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare const x: number | null;
      if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableNumber", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean */},
			{
				Code: "(x?: number) => !x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableNumber", Line: 1},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean */},
			{
				Code: "<T extends number | null | undefined>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableNumber", Line: 1},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean */},
			{
				Code: `
      function foo(x: 0 | 1 | null) {
      if (!x) {
    }
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableNumber", Line: 3},
				}, /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean */
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				}, /* Suggestions: conditionFixCompareNullish */
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
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
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      enum ExampleEnum {
      This = 0,
      That = 'one',
    }
    (value?: ExampleEnum) => (value ? 1 : 0);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 6},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      enum ExampleEnum {
      This = '',
      That = 1,
    }
    (value?: ExampleEnum) => (!value ? 1 : 0);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 6},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      enum ExampleEnum {
      This = 'this',
      That = 1,
    }
    (value?: ExampleEnum) => (!value ? 1 : 0);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 6},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      enum ExampleEnum {
      This = '',
      That = 0,
    }
    (value?: ExampleEnum) => (!value ? 1 : 0);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableEnum: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableEnum", Line: 6},
				} /* Suggestions: conditionFixCompareNullish */},
			{
				Code: `
      if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 2},
				} /* Suggestions: conditionFixCastBoolean */},
			{
				Code: "x => !x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 1},
				} /* Suggestions: conditionFixCastBoolean */},
			{
				Code: "<T extends any>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 1},
				} /* Suggestions: conditionFixCastBoolean */},
			{
				Code: "<T,>(x: T) => (x ? 1 : 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 1},
				} /* Suggestions: conditionFixCastBoolean */},
			{
				Code: `
      declare const x: string[] | null;
    if (x) {
    }
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStrictNullCheck", Line: 0}, {MessageId: "unexpectedObject", Line: 3},
				}},
			{
				Code: `
      declare const obj: { x: number } | null;
      !obj ? 1 : 0
      !obj
      obj || 0
      obj && 1 || 0
      `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableObject", Line: 3}, {MessageId: "unexpectedNullableObject", Line: 4}, {MessageId: "unexpectedNullableObject", Line: 5}, {MessageId: "unexpectedNullableObject", Line: 6},
				} /* Suggestions: conditionFixCompareNullish, conditionFixCompareNullish, conditionFixCompareNullish, conditionFixCompareNullish */},
			{
				Code: `
      declare function assert(x: unknown): asserts x;
      declare const nullableString: string | null;
      assert(nullableString);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 4},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare function assert(a: number, b: unknown): asserts b;
      declare const nullableString: string | null;
      assert(foo, nullableString);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 4},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare function assert(a: number, b: unknown): asserts b;
      declare function assert(one: number, two: unknown): asserts two;
      declare const nullableString: string | null;
      assert(foo, nullableString);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 5},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare function assert(this: object, a: number, b: unknown): asserts b;
      declare const nullableString: string | null;
      assert(foo, nullableString);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 4},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
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
					{MessageId: "unexpectedNullableString"},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
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
					{MessageId: "unexpectedNullableString", Line: 18},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
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
					{MessageId: "unexpectedNullableString", Line: 19},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
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
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
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
					{MessageId: "unexpectedNullableString", Line: 8},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
    ['one', 'two', ''].find(x => {
      return x;
    });
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    ['one', 'two', ''].find(x => {
      return;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    ['one', 'two', ''].findLast(x => {
      return undefined;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    ['one', 'two', ''].find(x => {
      if (x) {
      return Math.random() > 0.5;
    }
  });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
      const predicate = (x: string) => {
      if (x) {
      return Math.random() > 0.5;
    }
  };

    ['one', 'two', ''].find(predicate);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableBoolean: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 8},
				}},
			{
				Code: `
    [1, null].every(async x => {
      return x != null;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "predicateCannotBeAsync", Line: 2},
				}},
			{
				Code: `
      const predicate = async x => {
      return x != null;
    };

    [1, null].every(predicate);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 6},
				}},
			{
				Code: `
    [1, null].every((x): boolean | number => {
      return x != null;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 2},
				}},
			{
				Code: `
    [1, null].every((x): boolean | undefined => {
      return x != null;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 2},
				}},
			{
				Code: `
    [1, null].every((x, i) => {});
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    [() => {}, null].every((x: () => void) => {});
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    [() => {}, null].every(function (x: () => void) {});
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
    [() => {}, null].every(() => {});
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullish", Line: 2},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
      declare function f(x: number): string;
      declare function f(x: string | null): boolean;

    [35].filter(f);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 5},
				}},
			{
				Code: `
      declare function f(x: number): string;
      declare function f(x: number | boolean): boolean;
      declare function f(x: string | null): boolean;

    [35].filter(f);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedMixedCondition", Line: 6},
				}},
			{
				Code: `
      declare function foo<T>(x: number): T;
    [1, null].every(foo);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 3},
				}},
			{
				Code: `
      function foo<T extends number>(x: number): T {}
    [1, null].every(foo);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 3},
				}},
			{
				Code: `
      declare const nullOrString: string | null;
    ['one', null].filter(x => nullOrString);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean, explicitBooleanReturnType */},
			{
				Code: `
      declare const nullOrString: string | null;
    ['one', null].filter(x => !nullOrString);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableString", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */},
			{
				Code: `
      declare const anyValue: any;
    ['one', null].filter(x => anyValue);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAny", Line: 3},
				} /* Suggestions: conditionFixCastBoolean, explicitBooleanReturnType */},
			{
				Code: `
      declare const nullOrBoolean: boolean | null;
    [true, null].filter(x => nullOrBoolean);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableBoolean", Line: 3},
				} /* Suggestions: conditionFixDefaultFalse, conditionFixCompareTrue, explicitBooleanReturnType */},
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
					{MessageId: "unexpectedNullableEnum", Line: 7},
				} /* Suggestions: conditionFixCompareNullish, explicitBooleanReturnType */},
			{
				Code: `
      declare const nullOrNumber: number | null;
    [0, null].filter(x => nullOrNumber);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableNumber", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean, explicitBooleanReturnType */},
			{
				Code: `
      const objectValue: object = {};
    [{ a: 0 }, {}].filter(x => objectValue);
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 3},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
      const objectValue: object = {};
    [{ a: 0 }, {}].filter(x => {
      return objectValue;
    });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObject", Line: 3},
				} /* Suggestions: explicitBooleanReturnType */},
			{
				Code: `
      declare const nullOrObject: object | null;
    [{ a: 0 }, null].filter(x => nullOrObject);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNullableObject: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNullableObject", Line: 3},
				} /* Suggestions: conditionFixCompareNullish, explicitBooleanReturnType */},
			{
				Code: `
      const numbers: number[] = [1];
    [1, 2].filter(x => numbers.length);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 3},
				} /* Suggestions: conditionFixCompareArrayLengthNonzero, explicitBooleanReturnType */},
			{
				Code: `
      const numberValue: number = 1;
    [1, 2].filter(x => numberValue);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowNumber: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNumber", Line: 3},
				} /* Suggestions: conditionFixCompareZero, conditionFixCompareNaN, conditionFixCastBoolean, explicitBooleanReturnType */},
			{
				Code: `
      const stringValue: string = 'hoge';
    ['hoge', 'foo'].filter(x => stringValue);
    `,
				Options: StrictBooleanExpressionsOptions{
					AllowString: utils.Ref(false),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedString", Line: 3},
				} /* Suggestions: conditionFixCompareStringLength, conditionFixCompareEmptyString, conditionFixCastBoolean, explicitBooleanReturnType */},
		})
}
