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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
declare const maybe: string | undefined;
const bar = maybe!;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const maybe: string | null;
const bar = maybe as string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
declare const maybe: string | null;
const bar = maybe!;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const maybe: string | null | undefined;
const bar = maybe as string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
declare const maybe: string | null | undefined;
const bar = maybe!;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe as Type;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
type Type = { value: string };
declare const maybe: Type | undefined;
const bar = maybe!;
      `,
						},
					},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      6,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
interface Interface {
  value: string;
}
declare const maybe: Interface | undefined;
const bar = maybe!;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type T = string | null;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
type T = string | null;
declare const x: T;

const y = x!;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type T = string | null | undefined;
declare const x: T;

const y = x as NonNullable<T>;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
type T = string | null | undefined;
declare const x: T;

const y = x!;
      `,
						},
					},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      5,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
declare function nullablePromise(): Promise<string | null>;

async function fn(): Promise<string> {
  return (await nullablePromise())!;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const a: string | null;

const b = (a || undefined) as string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      4,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
declare const a: string | null;

const b = (a || undefined)!;
      `,
						},
					},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNonNullAssertion",
					Line:      3,
					Column:    30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferNonNullAssertion",
							Output: `
function first<T extends string | number>(array: ArrayLike<T>): T | null {
  return array.length > 0 ? (array[0]!) : null;
}
        `,
						},
					},
				},
			},
		},
	})
}
