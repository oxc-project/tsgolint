package require_array_sort_compare

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestRequireArraySortCompare(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RequireArraySortCompareRule, []rule_tester.ValidTestCase{
		{Code: `
      function f(a: any[]) {
        a.sort(undefined);
      }
    `},
		{Code: `
      function f(a: any[]) {
        a.sort((a, b) => a - b);
      }
    `},
		{Code: `
      function f(a: Array<string>) {
        a.sort(undefined);
      }
    `},
		{Code: `
      function f(a: Array<number>) {
        a.sort((a, b) => a - b);
      }
    `},
		{Code: `
      function f(a: { sort(): void }) {
        a.sort();
      }
    `},
		{Code: `
      class A {
        sort(): void {}
      }
      function f(a: A) {
        a.sort();
      }
    `},
		{Code: `
      interface A {
        sort(): void;
      }
      function f(a: A) {
        a.sort();
      }
    `},
		{Code: `
      interface A {
        sort(): void;
      }
      function f<T extends A>(a: T) {
        a.sort();
      }
    `},
		{Code: `
      function f(a: any) {
        a.sort();
      }
    `},
		{Code: `
      namespace UserDefined {
        interface Array {
          sort(): void;
        }
        function f(a: Array) {
          a.sort();
        }
      }
    `},
		{Code: `
      function f(a: any[]) {
        a?.sort((a, b) => a - b);
      }
    `},
		{Code: `
      namespace UserDefined {
        interface Array {
          sort(): void;
        }
        function f(a: Array) {
          a?.sort();
        }
      }
    `},
		{
			Code: `
        ['foo', 'bar', 'baz'].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
		},
		{
			Code: `
        function getString() {
          return 'foo';
        }
        [getString(), getString()].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
		},
		{
			Code: `
        const foo = 'foo';
        const bar = 'bar';
        const baz = 'baz';
        [foo, bar, baz].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
		},
		{
			Code: `
        declare const x: string[];
        x.sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
		},
		{
			Code: `
        function f(a: number[]) {
          a.toSorted((a, b) => a - b);
        }
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        function f(a: Array<any>) {
          a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number[]) {
          a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number[]) {
          a.sort();
        }
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number | number[]) {
          if (Array.isArray(a)) a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: string | string[]) {
          if (Array.isArray(a)) a.sort();
        }
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number[] | string[]) {
          a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f<T extends number[]>(a: T) {
          a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f<T, U extends T[]>(a: U) {
          a.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number[]) {
          a?.sort();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        [1, 2, 3].sort();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function getNumber() {
          return 1;
        }
        [getNumber(), getNumber()].sort();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        const foo = 1;
        const bar = 2;
        const baz = 3;
        [foo, bar, baz].sort();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        [2, 'bar', 'baz'].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function getNumber() {
          return 2;
        }
        [2, 3].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        const one = 1;
        const two = 2;
        const three = 3;
        [one, two, three].sort();
      `,
			Options: RequireArraySortCompareOptions{IgnoreStringArrays: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
		{
			Code: `
        function f(a: number[]) {
          a.toSorted();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireCompare",
				},
			},
		},
	})
}
