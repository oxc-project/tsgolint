package prefer_find

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferFindRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferFindRule, []rule_tester.ValidTestCase{
		{Code: "\n      interface JerkCode<T> {\n        filter(predicate: (item: T) => boolean): JerkCode<T>;\n      }\n\n      declare const jerkCode: JerkCode<string>;\n\n      jerkCode.filter(item => item === 'aha')[0];\n    "},
		{Code: "\n      declare const arr: readonly string[];\n      arr.filter(item => item === 'aha')[1];\n    "},
		{Code: "\n      declare const arr: string[];\n      arr.filter(item => item === 'aha').at(1);\n    "},
		{Code: "\n      declare const notNecessarilyAnArray: unknown[] | undefined | null | string;\n      notNecessarilyAnArray?.filter(item => true)[0];\n    "},
		{Code: "[].filter(() => true)?.[0];"},
		{Code: "[].filter(() => true)?.at?.(0);"},
		{Code: "[].filter?.(() => true)[0];"},
		{Code: "[1, 2, 3].filter(x => x > 0).at(-Infinity);"},
		{Code: "\n      declare const arr: string[];\n      declare const cond: Parameters<Array<string>['filter']>[0];\n      const a = { arr };\n      a?.arr.filter(cond).at(1);\n    "},
		{Code: "['Just', 'a', 'filter'].filter(x => x.length > 4);"},
		{Code: "['Just', 'a', 'find'].find(x => x.length > 4);"},
		{Code: "undefined.filter(x => x)[0];"},
		{Code: "null?.filter(x => x)[0];"},
		{Code: "\n      declare function foo(param: any): any;\n      foo(Symbol.for('foo'));\n    "},
		{Code: "\n      declare const arr: string[];\n      const s = Symbol.for(\"Don't throw!\");\n      arr.filter(item => item === 'aha').at(s);\n    "},
		{Code: "[1, 2, 3].filter(x => x)[Symbol('0')];"},
		{Code: "[1, 2, 3].filter(x => x)[Symbol.for('0')];"},
		{Code: "(Math.random() < 0.5 ? [1, 2, 3].filter(x => true) : [1, 2, 3])[0];"},
		{Code: "\n      (Math.random() < 0.5\n        ? [1, 2, 3].find(x => true)\n        : [1, 2, 3].filter(x => true))[0];\n    "},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "\ndeclare const arr: string[];\narr.filter(item => item === 'aha')[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: string[];\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: Array<string>;\nconst zero = 0;\narr.filter(item => item === 'aha')[zero];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: Array<string>;\nconst zero = 0;\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: Array<string>;\nconst zero = 0n;\narr.filter(item => item === 'aha')[zero];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: Array<string>;\nconst zero = 0n;\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: Array<string>;\nconst zero = -0n;\narr.filter(item => item === 'aha')[zero];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: Array<string>;\nconst zero = -0n;\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: readonly string[];\narr.filter(item => item === 'aha').at(0);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: readonly string[];\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: ReadonlyArray<string>;\n(undefined, arr.filter(item => item === 'aha')).at(0);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: ReadonlyArray<string>;\n(undefined, arr.find(item => item === 'aha'));\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: string[];\nconst zero = 0;\narr.filter(item => item === 'aha').at(zero);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: string[];\nconst zero = 0;\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: string[];\narr.filter(item => item === 'aha')['0'];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: string[];\narr.find(item => item === 'aha');\n      "}}}},
		},
		{
			Code:   "const two = [1, 2, 3].filter(item => item === 2)[0];",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "const two = [1, 2, 3].find(item => item === 2);"}}}},
		},
		{
			Code:   "const fltr = \"filter\"; (([] as unknown[]))[fltr] ((item) => { return item === 2 }  ) [ 0  ] ;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "const fltr = \"filter\"; (([] as unknown[]))[\"find\"] ((item) => { return item === 2 }  )  ;"}}}},
		},
		{
			Code:   "(([] as unknown[]))?.[\"filter\"] ((item) => { return item === 2 }  ) [ 0  ] ;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "(([] as unknown[]))?.[\"find\"] ((item) => { return item === 2 }  )  ;"}}}},
		},
		{
			Code:   "\ndeclare const nullableArray: unknown[] | undefined | null;\nnullableArray?.filter(item => true)[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const nullableArray: unknown[] | undefined | null;\nnullableArray?.find(item => true);\n      "}}}},
		},
		{
			Code:   "([]?.filter(f))[0];",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "([]?.find(f));"}}}},
		},
		{
			Code:   "\ndeclare const objectWithArrayProperty: { arr: unknown[] };\ndeclare function cond(x: unknown): boolean;\nconsole.log((1, 2, objectWithArrayProperty?.arr['filter'](cond)).at(0));\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const objectWithArrayProperty: { arr: unknown[] };\ndeclare function cond(x: unknown): boolean;\nconsole.log((1, 2, objectWithArrayProperty?.arr[\"find\"](cond)));\n      "}}}},
		},
		{
			Code:   "\n[1, 2, 3].filter(x => x > 0).at(NaN);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\n[1, 2, 3].find(x => x > 0);\n      "}}}},
		},
		{
			Code:   "\nconst idxToLookUp = -0.12635678;\n[1, 2, 3].filter(x => x > 0).at(idxToLookUp);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\nconst idxToLookUp = -0.12635678;\n[1, 2, 3].find(x => x > 0);\n      "}}}},
		},
		{
			Code:   "\n[1, 2, 3].filter(x => x > 0)[`at`](0);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\n[1, 2, 3].find(x => x > 0);\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: string[];\ndeclare const cond: Parameters<Array<string>['filter']>[0];\nconst a = { arr };\na?.arr\n  .filter(cond) /* what a bad spot for a comment. Let's make sure\n  there's some yucky symbols too. [ . ?. <>   ' ' \\'] */\n  .at('0');\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: string[];\ndeclare const cond: Parameters<Array<string>['filter']>[0];\nconst a = { arr };\na?.arr\n  .find(cond) /* what a bad spot for a comment. Let's make sure\n  there's some yucky symbols too. [ . ?. <>   ' ' \\'] */\n  ;\n      "}}}},
		},
		{
			Code:   "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ 'filter' /* another comment */](\n    x => x[1] > 0,\n  ) /**/[`0`]!,\n);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ \"find\" /* another comment */](\n    x => x[1] > 0,\n  ) /**/!,\n);\n      "}}}},
		},
		{
			Code:   "\nfunction actingOnArray<T extends string[]>(values: T) {\n  return values.filter(filter => filter === 'filter')[\n    /* filter */ -0.0 /* filter */\n  ];\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\nfunction actingOnArray<T extends string[]>(values: T) {\n  return values.find(filter => filter === 'filter');\n}\n      "}}}},
		},
		{
			Code:   "\nconst nestedSequenceAbomination =\n  (1,\n  2,\n  (1,\n  2,\n  3,\n  (1, 2, 3, 4),\n  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].filter(x => x % 2 == 0)))['0']);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\nconst nestedSequenceAbomination =\n  (1,\n  2,\n  (1,\n  2,\n  3,\n  (1, 2, 3, 4),\n  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].find(x => x % 2 == 0))));\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: { a: 1 }[] & { b: 2 }[];\narr.filter(f, thisArg)[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: { a: 1 }[] & { b: 2 }[];\narr.find(f, thisArg);\n      "}}}},
		},
		{
			Code:   "\ndeclare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);\narr.filter(f, thisArg)[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);\narr.find(f, thisArg);\n      "}}}},
		},
		{
			Code:   "\n(Math.random() < 0.5\n  ? [1, 2, 3].filter(x => false)\n  : [1, 2, 3].filter(x => true))[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\n(Math.random() < 0.5\n  ? [1, 2, 3].find(x => false)\n  : [1, 2, 3].find(x => true));\n      "}}}},
		},
		{
			Code:   "\nMath.random() < 0.5\n  ? [1, 2, 3].find(x => true)\n  : [1, 2, 3].filter(x => true)[0];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\nMath.random() < 0.5\n  ? [1, 2, 3].find(x => true)\n  : [1, 2, 3].find(x => true);\n      "}}}},
		},
		{
			Code:   "\ndeclare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,\n  g: (arg0: unknown) => boolean;\nconst nestedTernaries = (\n  Math.random() < 0.5\n    ? Math.random() < 0.5\n      ? [1, 2, 3].filter(f)\n      : []?.filter(x => 'shrug')\n    : [2, 3, 4]['filter'](g)\n).at(0.2);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,\n  g: (arg0: unknown) => boolean;\nconst nestedTernaries = (\n  Math.random() < 0.5\n    ? Math.random() < 0.5\n      ? [1, 2, 3].find(f)\n      : []?.find(x => 'shrug')\n    : [2, 3, 4][\"find\"](g)\n);\n      "}}}},
		},
		{
			Code:   "\ndeclare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;\nconst nestedTernariesWithSequenceExpression = (\n  Math.random() < 0.5\n    ? ('sequence',\n      'expression',\n      Math.random() < 0.5 ? [1, 2, 3].filter(f) : []?.filter(x => 'shrug'))\n    : [2, 3, 4]['filter'](g)\n).at(0.2);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;\nconst nestedTernariesWithSequenceExpression = (\n  Math.random() < 0.5\n    ? ('sequence',\n      'expression',\n      Math.random() < 0.5 ? [1, 2, 3].find(f) : []?.find(x => 'shrug'))\n    : [2, 3, 4][\"find\"](g)\n);\n      "}}}},
		},
		{
			Code:   "\ndeclare const spreadArgs: [(x: unknown) => boolean];\n[1, 2, 3].filter(...spreadArgs).at(0);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferFind", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "preferFindSuggestion", Output: "\ndeclare const spreadArgs: [(x: unknown) => boolean];\n[1, 2, 3].find(...spreadArgs);\n      "}}}},
		},
	})
}
