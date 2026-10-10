package no_unsafe_type_assertion

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeTypeAssertionRule_BasicAssertions(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const a: string;
a as string | number;
      `},
		{Code: `
declare const a: string;
<string | number>a;
      `},
		{Code: `
declare const a: string;
a as string | number as string | number | boolean;
      `},
		{Code: `
declare const a: string;
a as string;
      `},
		{Code: `
declare const a: { hello: 'world' };
a as { hello: string };
      `},
		{Code: `
'hello' as const;
      `},
		{Code: `
function foo<T extends boolean>(a: T) {
  return a as T | number;
}
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const a: string | number;
a as string;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
declare const a: string | number;
a satisfies string as string;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
declare const a: string | number;
<string>a;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
declare const a: string | undefined;
a as string | boolean;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare const a: string;
a as 'foo' as 'bar';
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
function foo<T extends boolean>(a: T) {
  return a as true;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
declare const a: string;
a as Omit<Required<Readonly<{ hello: 'world'; foo: 'bar' }>>, 'foo'>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 69,
				},
			},
		},
		{
			Code: `
declare const foo: readonly number[];
const bar = foo as number[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_AnyAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const _any_: any;
_any_ as any;
      `},
		{Code: `
declare const _any_: any;
_any_ as unknown;
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const _any_: any;
_any_ as string;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
declare const _unknown_: unknown;
_unknown_ as any;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
declare const _any_: any;
_any_ as Function;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
declare const _any_: any;
_any_ as never;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
'foo' as any;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
const bar = foo as number;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
const bar = 'foo' as errorType;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 31,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_NeverAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const _never_: never;
_never_ as never;
      `},
		{Code: `
declare const _never_: never;
_never_ as unknown;
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const _never_: never;
_never_ as any;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
declare const _string_: string;
_string_ as never;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 18,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_FunctionAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const _function_: Function;
_function_ as Function;
      `},
		{Code: `
declare const _function_: Function;
_function_ as unknown;
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const _function_: Function;
_function_ as () => void;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const _function_: Function;
_function_ as any;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
declare const _function_: Function;
_function_ as never;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_ObjectAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
// additional properties should be allowed
export const foo = { bar: 1, bazz: 1 } as {
  bar: number;
};
      `},
		{Code: `
declare const a: { hello: string } & { world: string };
a as { hello: string };
      `},
		{Code: `
declare const a: { hello: any };
a as { hello: unknown };
      `},
		{Code: `
declare const a: { hello: string };
a as { hello?: string };
      `},
		{Code: `
declare const a: { hello: string };
a satisfies Record<string, string> as { hello?: string };
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
var foo = {} as {
  bar: number;
  bas: string;
};
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      2,
					Column:    11,
					EndLine:   5,
					EndColumn: 2,
				},
			},
		},
		{
			Code: `
declare const a: { hello: string };
a satisfies Record<string, string> as { hello: string; world: string };
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 71,
				},
			},
		},
		{
			Code: `
declare const a: { hello?: string };
a as { hello: string };
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_ArrayAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const a: string[];
a as (string | number)[];
      `},
		{Code: `
declare const a: number[];
a as unknown[];
      `},
		{Code: `
declare const a: { hello: 'world'; foo: 'bar' }[];
a as { hello: 'world' }[];
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const a: (string | number)[];
a as string[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: any[];
a as number[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: number[];
a as any[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
declare const a: unknown[];
a as number[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: number[];
a as never[];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 13,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_TupleAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const a: [string];
a as [string | number];
      `},
		{Code: `
declare const a: [string, number];
a as [string, string | number];
      `},
		{Code: `
declare const a: [string];
a as [unknown];
      `},
		{Code: `
declare const a: [{ hello: 'world'; foo: 'bar' }];
a as [{ hello: 'world' }];
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const a: [string | number];
a as [string];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: [string, number];
a as [string, string];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare const a: [string];
a as [string, number];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare const a: [string, number];
a as [string];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: [any];
a as [number];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: [number, any];
a as [number, number];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare const a: [number];
a as [any];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
declare const a: [unknown];
a as [number];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const a: [number];
a as [never];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
declare const a: [Promise<string | number>];
a as [Promise<string>];
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_PromiseAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const a: Promise<string>;
a as Promise<string | number>;
      `},
		{Code: `
declare const a: Promise<number>;
a as Promise<unknown>;
      `},
		{Code: `
declare const a: Promise<{ hello: 'world'; foo: 'bar' }>;
a as Promise<{ hello: 'world' }>;
      `},
		{Code: `
declare const a: Promise<string>;
a as Promise<string> | string;
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const a: Promise<string | number>;
a as Promise<string>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
declare const a: Promise<any>;
a as Promise<number>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
declare const a: Promise<number>;
a as Promise<any>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
declare const a: Promise<number[]>;
a as Promise<any[]>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const a: Promise<unknown>;
a as Promise<number>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
declare const a: Promise<number>;
a as Promise<never>;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_ClassAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
class Foo {}
declare const a: Foo;
a as Foo | number;
      `},
		{Code: `
class Foo {}
class Bar {}
declare const a: Foo;
a as Bar;
      `},
		{Code: `
class Foo {
  hello() {}
}
class Bar {}
declare const a: Foo;
a as Bar;
      `},
		{Code: `
class Foo {
  hello() {}
}
class Bar extends Foo {}
declare const a: Bar;
a as Foo;
      `},
		{Code: `
class Foo {
  hello() {}
}
class Bar extends Foo {}
declare const a: Foo;
a as Bar;
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Foo {
  hello() {}
}
class Bar extends Foo {
  world() {}
}
declare const a: Foo;
a as Bar;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      9,
					Column:    1,
					EndLine:   9,
					EndColumn: 9,
				},
			},
		},
	})
}

func TestNoUnsafeTypeAssertionRule_GenericAssertions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
type Obj = { foo: string };
function func<T extends Obj>(a: T) {
  const b = a as T;
}
      `},
		{Code: `
function parameterExtendsOtherParameter<T extends string | number, V extends T>(
  x: T,
  y: V,
) {
  y as T;
}
      `},
		{Code: `
function parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {
  y as T;
}
      `},
		{Code: `
function unconstrainedToUnknown<T>(x: T) {
  x as unknown;
}
      `},
		{Code: `
function stringToWider<T extends string>(x: T) {
  x as number | string; // allowed
}
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
type Obj = { foo: string };
function func<T extends Obj>() {
  const myObj = { foo: 'hi' } as T;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
type Obj = { foo: string };
function func<T extends Obj>() {
  const o: Obj = { foo: 'hi' };
  const myObj = o as T;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      5,
					Column:    17,
					EndLine:   5,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
export function myfunc<CustomObjectT extends string>(
  input: number,
): CustomObjectT {
  const newCustomObject = input as CustomObjectT;
  return newCustomObject;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      5,
					Column:    27,
					EndLine:   5,
					EndColumn: 49,
				},
			},
		},
		{
			Code: `
function unknownConstraint<T extends unknown>(x: T, y: string) {
  y as T; // banned; generic arbitrary subtype
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function unconstrained<T>(x: T, y: string) {
  y as T;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToUnconstrainedTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
// constraint of any functions like constraint of ` + "`" + `unknown` + "`" + `
// (even the TS error message has this verbiage)
function anyConstraint<T extends any>(x: T, y: string) {
  y as T; // banned; generic arbitrary subtype
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function constraintWiderThanUncastType<T extends string | number>(
  x: T,
  y: string,
) {
  y as T; // banned; assignable to constraint
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function constraintEqualUncastType<T extends string>(x: T, y: string) {
  y as T; // banned; assignable to constraint
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function constraintNarrowerThanUncastType<T extends string>(
  x: T,
  y: string | number,
) {
  y as T; // banned; *not* assignable to constraint
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function assertFromAny<T extends string | number>(x: T, y: any) {
  y as T; // banned; just an ` + "`" + `any` + "`" + ` complaint. Not a generic subtype.
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeOfAnyTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function parameterExtendsOtherParameter<T extends string | number, V extends T>(
  x: T,
  y: V,
) {
  x as V; // banned; assignable to constraint
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertionAssignableToConstraint",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {
  x as V; // banned; unconstrained arbitrary type
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToUnconstrainedTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function twoUnconstrained<T, V>(x: T, y: V) {
  y as T;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToUnconstrainedTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
function toNarrower<T>(x: T, y: string) {
  x as string; // banned; ordinary 'string' narrower than 'T'.
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
function unconstrainedToAny<T>(x: T) {
  x as any;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
function stringToAny<T extends string>(x: T) {
  x as any;
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeToAnyTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
function stringToNarrower<T extends string>(x: T) {
  x as 'a' | 'b';
}
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTypeAssertion",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
	})
}
