package non_nullable_type_assertion_style

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNonNullableTypeAssertionStyleRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
		{Code: `
declare const original: number | string;
const cast = original as string;
    `},
		{Code: `
declare const original: number | undefined;
const cast = original as string | number | undefined;
    `},
		{Code: `
declare const original: number | any;
const cast = original as string | number | undefined;
    `},
		{Code: `
declare const original: number | undefined;
const cast = original as any;
    `},
		{Code: `
declare const original: number | null | undefined;
const cast = original as number | null;
    `},
		{Code: `
type Type = { value: string };
declare const original: Type | number;
const cast = original as Type;
    `},
		{Code: `
type T = string;
declare const x: T | number;

const y = x as NonNullable<T>;
    `},
		{Code: `
type T = string | null;
declare const x: T | number;

const y = x as NonNullable<T>;
    `},
		{Code: `
const foo = [] as const;
    `},
		{Code: `
const x = 1 as 1;
    `},
		{Code: `
declare function foo<T = any>(): T;
const bar = foo() as number;
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const maybe: string | undefined;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
declare const maybe: string | null;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | null;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
declare const maybe: string | null | undefined;
const bar = maybe as string;
      `,
			Output: []string{`
declare const maybe: string | null | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe as Type;
      `,
			Output: []string{`
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    13,
				},
			},
		},
		{
			Code: `
interface Interface {
  value: string;
}
declare const maybe: Interface | undefined;
const bar = maybe as Interface;
      `,
			Output: []string{`
interface Interface {
  value: string;
}
declare const maybe: Interface | undefined;
const bar = maybe!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      6,
					Column:    13,
				},
			},
		},
		{
			Code: `
type T = string | null;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Output: []string{`
type T = string | null;
declare const x: T;

const y = x!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
				},
			},
		},
		{
			Code: `
type T = string | null | undefined;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Output: []string{`
type T = string | null | undefined;
declare const x: T;

const y = x!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
				},
			},
		},
		{
			Code: `
declare function nullablePromise(): Promise<string | null>;

async function fn(): Promise<string> {
  return (await nullablePromise()) as string;
}
      `,
			Output: []string{`
declare function nullablePromise(): Promise<string | null>;

async function fn(): Promise<string> {
  return (await nullablePromise())!;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    10,
				},
			},
		},
		{
			Code: `
declare const a: string | null;

const b = (a || undefined) as string;
      `,
			Output: []string{`
declare const a: string | null;

const b = (a || undefined)!;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    11,
				},
			},
		},
	})
}

func TestNonNullableTypeAssertionStyleRule_noUncheckedIndexedAccess(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NonNullableTypeAssertionStyleRule, []rule_tester.ValidTestCase{
		{Code: `
function first<T>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | null>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | undefined>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
function first<T extends string | null | undefined>(
  array: ArrayLike<T>,
): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
		{Code: `
type A = 'a' | 'A';
type B = 'b' | 'B';
function first<T extends A | B | null>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function first<T extends string | number>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0] as T) : null;
}
        `,
			Output: []string{`
function first<T extends string | number>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0]!) : null;
}
        `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    30,
				},
			},
		},
	})
}
