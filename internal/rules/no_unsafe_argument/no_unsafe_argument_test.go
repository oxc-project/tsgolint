package no_unsafe_argument

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeArgumentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeArgumentRule, []rule_tester.ValidTestCase{
		{Code: `
doesNotExist(1 as any);
    `},
		{Code: `
const foo = 1;
foo(1 as any);
    `},
		{Code: `
declare function foo(arg: number): void;
foo(1, 1 as any, 2 as any);
    `},
		{Code: `
declare function foo(arg: number, arg2: string): void;
foo(1, 'a');
    `},
		{Code: `
declare function foo(arg: any): void;
foo(1 as any);
    `},
		{Code: `
declare function foo(arg: unknown): void;
foo(1 as any);
    `},
		{Code: `
declare function foo(...arg: number[]): void;
foo(1, 2, 3);
    `},
		{Code: `
declare function foo(...arg: any[]): void;
foo(1, 2, 3, 4 as any);
    `},
		{Code: `
declare function foo(arg: number, arg2: number): void;
const x = [1, 2] as const;
foo(...x);
    `},
		{Code: `
declare function foo(arg: any, arg2: number): void;
const x = [1 as any, 2] as const;
foo(...x);
    `},
		{Code: `
declare function foo(arg1: string, arg2: string): void;
const x: string[] = [];
foo(...x);
    `},
		{Code: `
function foo(arg1: number, arg2: number) {}
foo(...([1, 1, 1] as [number, number, number]));
    `},
		{Code: `
declare function foo(arg1: Set<string>, arg2: Map<string, string>): void;

const x = [new Map<string, string>()] as const;
foo(new Set<string>(), ...x);
    `},
		{Code: `
declare function foo(arg1: unknown, arg2: Set<unknown>, arg3: unknown[]): void;
foo(1 as any, new Set<any>(), [] as any[]);
    `},
		{Code: `
declare function foo(...params: [number, string, any]): void;
foo(1, 'a', 1 as any);
    `},
		{Code: `
declare function foo<E extends string[]>(...params: E): void;

foo('a', 'b', 1 as any);
    `},
		{Code: `
declare function toHaveBeenCalledWith<E extends any[]>(...params: E): void;
toHaveBeenCalledWith(1 as any);
    `},
		{Code: `
declare function acceptsMap(arg: Map<string, string>): void;
acceptsMap(new Map());
    `},
		{Code: `
type T = [number, T[]];
declare function foo(t: T): void;
declare const t: T;

foo(t);
    `},
		{Code: `
type T = Array<T>;
declare function foo<T>(t: T): T;
const t: T = [];
foo(t);
    `},
		{Code: `
function foo(templates: TemplateStringsArray) {}
foo` + "`" + `` + "`" + `;
    `},
		{Code: `
function foo(templates: TemplateStringsArray, arg: any) {}
foo` + "`" + `${1 as any}` + "`" + `;
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare function foo(arg: number): void;
foo(1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    5,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
declare function foo(arg: number): void;
foo(error);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    5,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: number, arg2: string): void;
foo(1, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    8,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
declare function foo(...arg: number[]): void;
foo(1, 2, 3, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    14,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare function foo(arg: string, ...arg: number[]): void;
foo(1 as any, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    5,
					EndColumn: 13,
				},
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    15,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;

foo(...(x as any));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeSpread",
					Line:      4,
					Column:    5,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;

foo(...(x as any[]));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArraySpread",
					Line:      4,
					Column:    5,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;

declare const errors: error[];

foo(...errors);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArraySpread",
					Line:      6,
					Column:    5,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;

const x = ['a', 1 as any] as const;
foo(...x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTupleSpread",
					Line:      5,
					Column:    5,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;

const x = ['a', error] as const;
foo(...x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTupleSpread",
					Line:      5,
					Column:    5,
					EndColumn: 9,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number): void;
foo(...(['foo', 1, 2] as [string, any, number]));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTupleSpread",
					Line:      3,
					Column:    5,
					EndColumn: 48,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number, arg2: string): void;

const x = [1] as const;
foo('a', ...x, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    16,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: string, arg2: number, ...rest: string[]): void;

const x = [1, 2] as [number, ...number[]];
foo('a', ...x, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    16,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
declare function foo(arg1: Set<string>, arg2: Map<string, string>): void;

const x = [new Map<any, string>()] as const;
foo(new Set<any>(), ...x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    5,
					EndColumn: 19,
				},
				{
					MessageId: "unsafeTupleSpread",
					Line:      5,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare function foo(...params: [number, string, any]): void;
foo(1 as any, 'a' as any, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    5,
					EndColumn: 13,
				},
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    15,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare function foo(param1: string, ...params: [number, string, any]): void;
foo('a', 1 as any, 'a' as any, 1 as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    10,
					EndColumn: 18,
				},
				{
					MessageId: "unsafeArgument",
					Line:      3,
					Column:    20,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
type T = [number, T[]];
declare function foo(t: T): void;
declare const t: T;
foo(t as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    5,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
function foo(
  templates: TemplateStringsArray,
  arg1: number,
  arg2: any,
  arg3: string,
) {}
declare const arg: any;
foo<number>` + "`" + `${arg}${arg}${arg}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      9,
					Column:    15,
					EndColumn: 18,
				},
				{
					MessageId: "unsafeArgument",
					Line:      9,
					Column:    27,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
function foo(templates: TemplateStringsArray, arg: number) {}
declare const arg: any;
foo` + "`" + `${arg}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      4,
					Column:    7,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
type T = [number, T[]];
function foo(templates: TemplateStringsArray, arg: T) {}
declare const arg: any;
foo` + "`" + `${arg}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArgument",
					Line:      5,
					Column:    7,
					EndColumn: 10,
				},
			},
		},
	})
}
