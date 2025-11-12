package no_array_delete

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoArrayDelete(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoArrayDeleteRule, []rule_tester.ValidTestCase{
		{Code: `
      declare const obj: { a: 1; b: 2 };
      delete obj.a;
    `},
		{Code: `
      declare const obj: { a: 1; b: 2 };
      delete obj['a'];
    `},
		{Code: `
      declare const arr: { a: 1; b: 2 }[][][][];
      delete arr[0][0][0][0].a;
    `},
		{Code: `
      declare const maybeArray: any;
      delete maybeArray[0];
    `},
		{Code: `
      declare const maybeArray: unknown;
      delete maybeArray[0];
    `},
		{Code: `
      declare function getObject<T extends { a: 1; b: 2 }>(): T;
      delete getObject().a;
    `},
		{Code: `
      declare function getObject<T extends number>(): { a: T; b: 2 };
      delete getObject().a;
    `},
		{Code: `
      declare const test: never;
      delete test[0];
    `},
		{Code: `
      delete console.log();
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        declare const arr: number[];
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[];
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: number[];
        declare const key: number;
        delete arr[key];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      4,
					Column:    9,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[];
        declare const key: number;
         arr.splice(key, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: number[];

        enum Keys {
          A,
          B,
        }

        delete arr[Keys.A];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      9,
					Column:    9,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[];

        enum Keys {
          A,
          B,
        }

         arr.splice(Keys.A, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: number[];
        declare function doWork(): void;
        delete arr[(doWork(), 1)];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      4,
					Column:    9,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[];
        declare function doWork(): void;
         arr.splice((doWork(), 1), 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: Array<number>;
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: Array<number>;
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: "delete [1, 2, 3][0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      1,
					Column:    1,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output:    " [1, 2, 3].splice(0, 1);",
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: unknown[];
        delete arr[Math.random() ? 0 : 1];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: unknown[];
         arr.splice(Math.random() ? 0 : 1, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: number[] | string[] | boolean[];
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[] | string[] | boolean[];
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: number[] & unknown;
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: number[] & unknown;
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: (number | string)[];
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: (number | string)[];
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const obj: { a: { b: { c: number[] } } };
        delete obj.a.b.c[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const obj: { a: { b: { c: number[] } } };
         obj.a.b.c.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function getArray<T extends number[]>(): T;
        delete getArray()[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare function getArray<T extends number[]>(): T;
         getArray().splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function getArray<T extends number>(): T[];
        delete getArray()[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare function getArray<T extends number>(): T[];
         getArray().splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        function deleteFromArray(a: number[]) {
          delete a[0];
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    11,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        function deleteFromArray(a: number[]) {
           a.splice(0, 1);
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        function deleteFromArray<T extends number>(a: T[]) {
          delete a[0];
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    11,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        function deleteFromArray<T extends number>(a: T[]) {
           a.splice(0, 1);
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        function deleteFromArray<T extends number[]>(a: T) {
          delete a[0];
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    11,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        function deleteFromArray<T extends number[]>(a: T) {
           a.splice(0, 1);
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const tuple: [number, string];
        delete tuple[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const tuple: [number, string];
         tuple.splice(0, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const a: number[];
        declare const b: number;

        delete [...a, ...a][b];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const a: number[];
        declare const b: number;

         [...a, ...a].splice(b, 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const a: number[];
        declare const b: number;

        // before expression
        delete /** multi
        line */ a[((
        // single-line
        b /* inline */ /* another-inline */ )
        ) /* another-one */ ] /* before semicolon */; /* after semicolon */
        // after expression
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const a: number[];
        declare const b: number;

        // before expression
         /** multi
        line */ a.splice(((
        // single-line
        b /* inline */ /* another-inline */ )
        ) /* another-one */ , 1) /* before semicolon */; /* after semicolon */
        // after expression
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const a: number[];
        declare const b: number;

        delete ((a[((b))]));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const a: number[];
        declare const b: number;

         ((a.splice(((b)), 1)));
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const a: number[];
        declare const b: number;

        delete a[(b + 1) * (b + 2)];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const a: number[];
        declare const b: number;

         a.splice((b + 1) * (b + 2), 1);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const arr: string & Array<number>;
        delete arr[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArrayDelete",
					Line:      3,
					Column:    9,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "useSplice",
							Output: `
        declare const arr: string & Array<number>;
         arr.splice(0, 1);
      `,
						},
					},
				},
			},
		},
	})
}
