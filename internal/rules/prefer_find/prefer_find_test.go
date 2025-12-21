package prefer_find

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferFindRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferFindRule, []rule_tester.ValidTestCase{
		{Code: `
      interface JerkCode<T> {
        filter(predicate: (item: T) => boolean): JerkCode<T>;
      }

      declare const jerkCode: JerkCode<string>;

      jerkCode.filter(item => item === 'aha')[0];
    `},
		{Code: `
      declare const arr: readonly string[];
      arr.filter(item => item === 'aha')[1];
    `},
		{Code: `
      declare const arr: string[];
      arr.filter(item => item === 'aha').at(1);
    `},
		{Code: `
      declare const notNecessarilyAnArray: unknown[] | undefined | null | string;
      notNecessarilyAnArray?.filter(item => true)[0];
    `},
		{Code: `[].filter(() => true)?.[0];`},
		{Code: `[].filter(() => true)?.at?.(0);`},
		{Code: `[].filter?.(() => true)[0];`},
		{Code: `[1, 2, 3].filter(x => x > 0).at(-Infinity);`},
		{Code: `
      declare const arr: string[];
      declare const cond: Parameters<Array<string>['filter']>[0];
      const a = { arr };
      a?.arr.filter(cond).at(1);
    `},
		{Code: `['Just', 'a', 'filter'].filter(x => x.length > 4);`},
		{Code: `['Just', 'a', 'find'].find(x => x.length > 4);`},
		{Code: `undefined.filter(x => x)[0];`},
		{Code: `null?.filter(x => x)[0];`},
		{Code: `
      declare function foo(param: any): any;
      foo(Symbol.for('foo'));
    `},
		{Code: `
      declare const arr: string[];
      const s = Symbol.for("Don't throw!");
      arr.filter(item => item === 'aha').at(s);
    `},
		{Code: `[1, 2, 3].filter(x => x)[Symbol('0')];`},
		{Code: `[1, 2, 3].filter(x => x)[Symbol.for('0')];`},
		{Code: `(Math.random() < 0.5 ? [1, 2, 3].filter(x => true) : [1, 2, 3])[0];`},
		{Code: `
      (Math.random() < 0.5
        ? [1, 2, 3].find(x => true)
        : [1, 2, 3].filter(x => true))[0];
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: Array<string>;
const zero = 0;
arr.filter(item => item === 'aha')[zero];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = 0;
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: Array<string>;
const zero = 0n;
arr.filter(item => item === 'aha')[zero];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = 0n;
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: Array<string>;
const zero = -0n;
arr.filter(item => item === 'aha')[zero];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = -0n;
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: readonly string[];
arr.filter(item => item === 'aha').at(0);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: readonly string[];
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: ReadonlyArray<string>;
(undefined, arr.filter(item => item === 'aha')).at(0);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: ReadonlyArray<string>;
(undefined, arr.find(item => item === 'aha'));
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: string[];
const zero = 0;
arr.filter(item => item === 'aha').at(zero);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
const zero = 0;
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')['0'];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
      `,
						},
					},
				},
			},
		},
		{
			Code: `const two = [1, 2, 3].filter(item => item === 2)[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `const two = [1, 2, 3].find(item => item === 2);`,
						},
					},
				},
			},
		},
		{
			Code: `const fltr = "filter"; (([] as unknown[]))[fltr] ((item) => { return item === 2 }  ) [ 0  ] ;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `const fltr = "filter"; (([] as unknown[]))["find"] ((item) => { return item === 2 }  )  ;`,
						},
					},
				},
			},
		},
		{
			Code: `(([] as unknown[]))?.["filter"] ((item) => { return item === 2 }  ) [ 0  ] ;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `(([] as unknown[]))?.["find"] ((item) => { return item === 2 }  )  ;`,
						},
					},
				},
			},
		},
		{
			Code: `
declare const nullableArray: unknown[] | undefined | null;
nullableArray?.filter(item => true)[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const nullableArray: unknown[] | undefined | null;
nullableArray?.find(item => true);
      `,
						},
					},
				},
			},
		},
		{
			Code: `([]?.filter(f))[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `([]?.find(f));`,
						},
					},
				},
			},
		},
		{
			Code: `
declare const objectWithArrayProperty: { arr: unknown[] };
declare function cond(x: unknown): boolean;
console.log((1, 2, objectWithArrayProperty?.arr['filter'](cond)).at(0));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const objectWithArrayProperty: { arr: unknown[] };
declare function cond(x: unknown): boolean;
console.log((1, 2, objectWithArrayProperty?.arr["find"](cond)));
      `,
						},
					},
				},
			},
		},
		{
			Code: `
[1, 2, 3].filter(x => x > 0).at(NaN);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
[1, 2, 3].find(x => x > 0);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const idxToLookUp = -0.12635678;
[1, 2, 3].filter(x => x > 0).at(idxToLookUp);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
const idxToLookUp = -0.12635678;
[1, 2, 3].find(x => x > 0);
      `,
						},
					},
				},
			},
		},
		{
			Code: "\n[1, 2, 3].filter(x => x > 0)[`at`](0);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
[1, 2, 3].find(x => x > 0);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: string[];
declare const cond: Parameters<Array<string>['filter']>[0];
const a = { arr };
a?.arr
  .filter(cond) /* what a bad spot for a comment. Let's make sure
  there's some yucky symbols too. [ . ?. <>   ' ' \'] */
  .at('0');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
declare const cond: Parameters<Array<string>['filter']>[0];
const a = { arr };
a?.arr
  .find(cond) /* what a bad spot for a comment. Let's make sure
  there's some yucky symbols too. [ . ?. <>   ' ' \'] */
  ;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ 'filter' /* another comment */](\n    x => x[1] > 0,\n  ) /**/[`0`]!,\n);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
const imNotActuallyAnArray = [
  [1, 2, 3],
  [2, 3, 4],
] as const;
const butIAm = [4, 5, 6];
butIAm.push(
  // line comment!
  ...imNotActuallyAnArray[/* comment */ "find" /* another comment */](
    x => x[1] > 0,
  ) /**/!,
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
function actingOnArray<T extends string[]>(values: T) {
  return values.filter(filter => filter === 'filter')[
    /* filter */ -0.0 /* filter */
  ];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
function actingOnArray<T extends string[]>(values: T) {
  return values.find(filter => filter === 'filter');
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const nestedSequenceAbomination =
  (1,
  2,
  (1,
  2,
  3,
  (1, 2, 3, 4),
  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].filter(x => x % 2 == 0)))['0']);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
const nestedSequenceAbomination =
  (1,
  2,
  (1,
  2,
  3,
  (1, 2, 3, 4),
  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].find(x => x % 2 == 0))));
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: { a: 1 }[] & { b: 2 }[];
arr.filter(f, thisArg)[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: { a: 1 }[] & { b: 2 }[];
arr.find(f, thisArg);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);
arr.filter(f, thisArg)[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);
arr.find(f, thisArg);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
(Math.random() < 0.5
  ? [1, 2, 3].filter(x => false)
  : [1, 2, 3].filter(x => true))[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
(Math.random() < 0.5
  ? [1, 2, 3].find(x => false)
  : [1, 2, 3].find(x => true));
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Math.random() < 0.5
  ? [1, 2, 3].find(x => true)
  : [1, 2, 3].filter(x => true)[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
Math.random() < 0.5
  ? [1, 2, 3].find(x => true)
  : [1, 2, 3].find(x => true);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,
  g: (arg0: unknown) => boolean;
const nestedTernaries = (
  Math.random() < 0.5
    ? Math.random() < 0.5
      ? [1, 2, 3].filter(f)
      : []?.filter(x => 'shrug')
    : [2, 3, 4]['filter'](g)
).at(0.2);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,
  g: (arg0: unknown) => boolean;
const nestedTernaries = (
  Math.random() < 0.5
    ? Math.random() < 0.5
      ? [1, 2, 3].find(f)
      : []?.find(x => 'shrug')
    : [2, 3, 4]["find"](g)
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;
const nestedTernariesWithSequenceExpression = (
  Math.random() < 0.5
    ? ('sequence',
      'expression',
      Math.random() < 0.5 ? [1, 2, 3].filter(f) : []?.filter(x => 'shrug'))
    : [2, 3, 4]['filter'](g)
).at(0.2);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;
const nestedTernariesWithSequenceExpression = (
  Math.random() < 0.5
    ? ('sequence',
      'expression',
      Math.random() < 0.5 ? [1, 2, 3].find(f) : []?.find(x => 'shrug'))
    : [2, 3, 4]["find"](g)
);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const spreadArgs: [(x: unknown) => boolean];
[1, 2, 3].filter(...spreadArgs).at(0);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const spreadArgs: [(x: unknown) => boolean];
[1, 2, 3].find(...spreadArgs);
      `,
						},
					},
				},
			},
		},
	})
}
