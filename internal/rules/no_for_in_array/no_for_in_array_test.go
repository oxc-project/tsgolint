package no_for_in_array

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoForInArray(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoForInArrayRule, []rule_tester.ValidTestCase{
		{Code: `
for (const x of [3, 4, 5]) {
  console.log(x);
}
    `},
		{Code: `
for (const x in { a: 1, b: 2, c: 3 }) {
  console.log(x);
}
    `},
		{Code: `
declare const nullish: null | undefined;
// @ts-expect-error
for (const k in nullish) {
}
    `},
		{Code: `
declare const obj: {
  [key: number]: number;
};

for (const key in obj) {
  console.log(key);
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
for (const x in [3, 4, 5]) {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
const z = [3, 4, 5];
for (const x in z) {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
const fn = (arr: number[]) => {
  for (const x in arr) {
    console.log(x);
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
const fn = (arr: number[] | string[]) => {
  for (const x in arr) {
    console.log(x);
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
const fn = <T extends any[]>(arr: T) => {
  for (const x in arr) {
    console.log(x);
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
for (const x
  in
    (
      (
        (
          [3, 4, 5]
        )
      )
    )
  )
  // weird
  /* spot for a */
  // comment
  /* ) */
  /* ( */
  {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      2,
					Column:    1,
					EndLine:   11,
					EndColumn: 4,
				},
			},
		},
		{
			Code: `
for (const x
  in
    (
      (
        (
          [3, 4, 5]
        )
      )
    )
  )
  // weird
  /* spot for a */
  // comment
  /* ) */
  /* ( */

  ((((console.log('body without braces ')))));

      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      2,
					Column:    1,
					EndLine:   11,
					EndColumn: 4,
				},
			},
		},
		{
			Code: `
declare const array: string[] | null;

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array: number[] | undefined;

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array: boolean[] | { a: 1; b: 2; c: 3 };

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array: [number, string];

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array: [number, string] | { a: 1; b: 2; c: 3 };

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array: string[] | Record<number, string>;

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
const arrayLike = /fe/.exec('foo');

for (const x in arrayLike) {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
declare const arrayLike: HTMLCollection;

for (const x in arrayLike) {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
declare const arrayLike: NodeList;

for (const x in arrayLike) {
  console.log(x);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
function foo() {
  for (const a in arguments) {
    console.log(a);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
declare const array:
  | (({ a: string } & string[]) | Record<string, boolean>)
  | Record<number, string>;

for (const key in array) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      6,
					Column:    1,
					EndLine:   6,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const array:
  | (({ a: string } & RegExpExecArray) | Record<string, boolean>)
  | Record<number, string>;

for (const key in array) {
  console.log(k);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      6,
					Column:    1,
					EndLine:   6,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare const obj: {
  [key: number]: number;
  length: 1;
};

for (const key in obj) {
  console.log(key);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forInViolation",
					Line:      7,
					Column:    1,
					EndLine:   7,
					EndColumn: 23,
				},
			},
		},
	})
}
