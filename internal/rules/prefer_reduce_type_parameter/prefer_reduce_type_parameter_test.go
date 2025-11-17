package prefer_reduce_type_parameter

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferReduceTypeParameterRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferReduceTypeParameterRule, []rule_tester.ValidTestCase{
		{Code: `
      new (class Mine {
        reduce() {}
      })().reduce(() => {}, 1 as any);
    `},
		{Code: `
      class Mine {
        reduce() {}
      }

      new Mine().reduce(() => {}, 1 as any);
    `},
		{Code: `
      import { Reducable } from './class';

      new Reducable().reduce(() => {}, 1 as any);
    `},
		{Code: "[1, 2, 3]['reduce']((sum, num) => sum + num, 0);"},
		{Code: "[1, 2, 3][null]((sum, num) => sum + num, 0);"},
		{Code: "[1, 2, 3]?.[null]((sum, num) => sum + num, 0);"},
		{Code: "[1, 2, 3].reduce((sum, num) => sum + num, 0);"},
		{Code: "[1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2), []);"},
		{Code: "[1, 2, 3]?.reduce<number[]>((a, s) => a.concat(s * 2), []);"},
		{Code: `
      declare const tuple: [number, number, number];
      tuple.reduce<number[]>((a, s) => a.concat(s * 2), []);
    `},
		{Code: `
      type Reducer = { reduce: (callback: (arg: any) => any, arg: any) => any };
      declare const tuple: [number, number, number] | Reducer;
      tuple.reduce(a => {
        return a.concat(1);
      }, [] as number[]);
    `},
		{Code: `
      type Reducer = { reduce: (callback: (arg: any) => any, arg: any) => any };
      declare const arrayOrReducer: number[] & Reducer;
      arrayOrReducer.reduce(a => {
        return a.concat(1);
      }, [] as number[]);
    `},
		{Code: `
      ['a', 'b'].reduce(
        (accum, name) => ({
          ...accum,
          [name]: true,
        }),
        {} as Record<'a' | 'b', boolean>,
      );
    `},
		{Code: `
      ['a', 'b'].reduce(
        (accum, name) => ({
          ...accum,
          [name]: true,
        }),
        { a: true, b: false, c: true } as Record<'a' | 'b', boolean>,
      );
    `},
		{Code: `
      function f<T extends Record<string, boolean>>() {
        ['a', 'b'].reduce(
          (accum, name) => ({
            ...accum,
            [name]: true,
          }),
          {} as T,
        );
      }
    `},
		{Code: `
      function f<T>() {
        ['a', 'b'].reduce(
          (accum, name) => ({
            ...accum,
            [name]: true,
          }),
          {} as T,
        );
      }
    `},
		{Code: `
      ['a', 'b'].reduce((accum, name) => ` + "`" + `${accum} | hello ${name}!` + "`" + `);
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const arr: string[];
arr.reduce<string | undefined>(acc => acc, arr.shift() as string | undefined);
      `,
			Output: []string{`
declare const arr: string[];
arr.reduce<string | undefined>(acc => acc, arr.shift());
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      3,
					Column:    44,
				},
			},
		},
		{
			Code:   "[1, 2, 3].reduce((a, s) => a.concat(s * 2), [] as number[]);",
			Output: []string{"[1, 2, 3].reduce< number[]>((a, s) => a.concat(s * 2), []);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      1,
					Column:    45,
				},
			},
		},
		{
			Code:   "[1, 2, 3].reduce((a, s) => a.concat(s * 2), <number[]>[]);",
			Output: []string{"[1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2),[]);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      1,
					Column:    45,
				},
			},
		},
		{
			Code:   "[1, 2, 3]?.reduce((a, s) => a.concat(s * 2), [] as number[]);",
			Output: []string{"[1, 2, 3]?.reduce< number[]>((a, s) => a.concat(s * 2), []);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      1,
					Column:    46,
				},
			},
		},
		{
			Code:   "[1, 2, 3]?.reduce((a, s) => a.concat(s * 2), <number[]>[]);",
			Output: []string{"[1, 2, 3]?.reduce<number[]>((a, s) => a.concat(s * 2),[]);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      1,
					Column:    46,
				},
			},
		},
		{
			Code: `
const names = ['a', 'b', 'c'];

names.reduce(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {} as Record<string, boolean>,
);
      `,
			Output: []string{`
const names = ['a', 'b', 'c'];

names.reduce< Record<string, boolean>>(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {},
);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      9,
					Column:    3,
				},
			},
		},
		{
			Code: `
['a', 'b'].reduce(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  <Record<string, boolean>>{},
);
      `,
			Output: []string{`
['a', 'b'].reduce<Record<string, boolean>>(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),{},
);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
['a', 'b']['reduce'](
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {} as Record<string, boolean>,
);
      `,
			Output: []string{`
['a', 'b']['reduce']< Record<string, boolean>>(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {},
);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
function f<T, U extends T[]>(a: U) {
  return a.reduce(() => {}, {} as Record<string, boolean>);
}
      `,
			Output: []string{`
function f<T, U extends T[]>(a: U) {
  return a.reduce< Record<string, boolean>>(() => {}, {});
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      3,
					Column:    29,
				},
			},
		},
		{
			Code: `
declare const tuple: [number, number, number];
tuple.reduce((a, s) => a.concat(s * 2), [] as number[]);
      `,
			Output: []string{`
declare const tuple: [number, number, number];
tuple.reduce< number[]>((a, s) => a.concat(s * 2), []);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      3,
					Column:    41,
				},
			},
		},
		{
			Code: `
declare const tupleOrArray: [number, number, number] | number[];
tupleOrArray.reduce((a, s) => a.concat(s * 2), [] as number[]);
      `,
			Output: []string{`
declare const tupleOrArray: [number, number, number] | number[];
tupleOrArray.reduce< number[]>((a, s) => a.concat(s * 2), []);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      3,
					Column:    48,
				},
			},
		},
		{
			Code: `
declare const tuple: [number, number, number] & number[];
tuple.reduce((a, s) => a.concat(s * 2), [] as number[]);
      `,
			Output: []string{`
declare const tuple: [number, number, number] & number[];
tuple.reduce< number[]>((a, s) => a.concat(s * 2), []);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      3,
					Column:    41,
				},
			},
		},
		{
			Code: `
['a', 'b'].reduce(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {} as Record<string, boolean>,
);
      `,
			Output: []string{`
['a', 'b'].reduce< Record<string, boolean>>(
  (accum, name) => ({
    ...accum,
    [name]: true,
  }),
  {},
);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
function f<T extends Record<string, boolean>>(t: T) {
  ['a', 'b'].reduce(
    (accum, name) => ({
      ...accum,
      [name]: true,
    }),
    t as Record<string, boolean | number>,
  );
}
      `,
			Output: []string{`
function f<T extends Record<string, boolean>>(t: T) {
  ['a', 'b'].reduce< Record<string, boolean | number>>(
    (accum, name) => ({
      ...accum,
      [name]: true,
    }),
    t,
  );
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferTypeParameter",
					Line:      8,
					Column:    5,
				},
			},
		},
	})
}
