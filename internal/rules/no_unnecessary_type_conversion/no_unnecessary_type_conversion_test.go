package no_unnecessary_type_conversion

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeConversion(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTypeConversionRule, []rule_tester.ValidTestCase{
		{Code: "String(1);"},
		{Code: "(1).toString();"},
		{Code: "`${1}`;"},
		{Code: "'' + 1;"},
		{Code: "1 + '';"},
		{Code: "\n      let str = 1;\n      str += '';\n    "},
		{Code: "const obj = { prop: 'asdf' }; obj.prop += '';"},
		{Code: "const arr = ['asdf']; arr[0] += '';"},
		{Code: "Number('2');"},
		{Code: "+'2';"},
		{Code: "~~'2';"},
		{Code: "~~1.1;"},
		{Code: "~~-1.1;"},
		{Code: "~~(1.5 + 2.3);"},
		{Code: "~~(1 / 3);"},
		{Code: "Boolean(0);"},
		{Code: "!!0;"},
		{Code: "BigInt(3);"},
		{Code: "new String('asdf');"},
		{Code: "new Number(2);"},
		{Code: "new Boolean(true);"},
		{Code: "!false;"},
		{Code: "~2;"},
		{Code: "\n      function String(value: unknown) {\n        return value;\n      }\n      String('asdf');\n      export {};\n    "},
		{Code: "\n      function Number(value: unknown) {\n        return value;\n      }\n      Number(2);\n      export {};\n    "},
		{Code: "\n      function Boolean(value: unknown) {\n        return value;\n      }\n      Boolean(true);\n      export {};\n    "},
		{Code: "\n      function BigInt(value: unknown) {\n        return value;\n      }\n      BigInt(3n);\n      export {};\n    "},
		{Code: "\n      function toString(value: unknown) {\n        return value;\n      }\n      toString('asdf');\n    "},
		{Code: "\n      export {};\n      declare const toString: string;\n      toString.toUpperCase();\n    "},
		{Code: "String(...(['asdf'] as [string]));"},
		{Code: "String('asdf', console.log('extra'));"},
		{Code: "'asdf'.toString(console.log('extra'));"},
		{Code: "String(new String());"},
		{Code: "new String().toString();"},
		{Code: "'' + new String();"},
		{Code: "new String() + '';"},
		{Code: "\n      let str = new String();\n      str += '';\n    "},
		{Code: "Number(new Number());"},
		{Code: "+new Number();"},
		{Code: "~~new Number();"},
		{Code: "Boolean(new Boolean());"},
		{Code: "!!new Boolean();"},
		{Code: "\n      enum CustomIds {\n        Id1 = 'id1',\n        Id2 = 'id2',\n      }\n      const customId = 'id1';\n      const compareWithToString = customId === CustomIds.Id1.toString();\n    "},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "String('asdf');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "'asdf';",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "'asdf' satisfies string;",
						},
					},
				},
			},
		},
		{
			Code: "'asdf'.toString();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    8,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "'asdf';",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "'asdf' satisfies string;",
						},
					},
				},
			},
		},
		{
			Code: "'' + 'asdf';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "'asdf';",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "'asdf' satisfies string;",
						},
					},
				},
			},
		},
		{
			Code: "'asdf' + '';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    7,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "'asdf';",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "'asdf' satisfies string;",
						},
					},
				},
			},
		},
		{
			Code: "\nlet str = 'asdf';\nstr += '';\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\nlet str = 'asdf';\n\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\nlet str = 'asdf';\nstr satisfies string;\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\nlet str = 'asdf';\n'asdf' + (str += '');\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\nlet str = 'asdf';\n'asdf' + (str);\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\nlet str = 'asdf';\n'asdf' + (str satisfies string);\n      ",
						},
					},
				},
			},
		},
		{
			Code: "Number(123);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "123;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "123 satisfies number;",
						},
					},
				},
			},
		},
		{
			Code: "+123;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "123;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "123 satisfies number;",
						},
					},
				},
			},
		},
		{
			Code: "~~123;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "123;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "123 satisfies number;",
						},
					},
				},
			},
		},
		{
			Code: "Boolean(true);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "true;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "true satisfies boolean;",
						},
					},
				},
			},
		},
		{
			Code: "!!true;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "true;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "true satisfies boolean;",
						},
					},
				},
			},
		},
		{
			Code: "BigInt(3n);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "3n;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "3n satisfies bigint;",
						},
					},
				},
			},
		},
		{
			Code: "\n        function f<T extends string>(x: T) {\n          return String(x);\n        }\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        function f<T extends string>(x: T) {\n          return x;\n        }\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        function f<T extends string>(x: T) {\n          return x satisfies string;\n        }\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\n        function f<T extends number>(x: T) {\n          return Number(x);\n        }\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        function f<T extends number>(x: T) {\n          return x;\n        }\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        function f<T extends number>(x: T) {\n          return x satisfies number;\n        }\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\n        function f<T extends boolean>(x: T) {\n          return Boolean(x);\n        }\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        function f<T extends boolean>(x: T) {\n          return x;\n        }\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        function f<T extends boolean>(x: T) {\n          return x satisfies boolean;\n        }\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\n        function f<T extends bigint>(x: T) {\n          return BigInt(x);\n        }\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        function f<T extends bigint>(x: T) {\n          return x;\n        }\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        function f<T extends bigint>(x: T) {\n          return x satisfies bigint;\n        }\n      ",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends string>(x: T) { return x + ''; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends string>(x: T) { return x; }",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends string>(x: T) { return x satisfies string; }",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends string>(x: T) { return '' + x; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends string>(x: T) { return x; }",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends string>(x: T) { return x satisfies string; }",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends string>(x: T) {\nx += '';\nreturn x;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends string>(x: T) {\n\nreturn x;\n}\n",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends string>(x: T) {\nx satisfies string;\nreturn x;\n}\n",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends number>(x: T) { return +x; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends number>(x: T) { return x; }",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends number>(x: T) { return x satisfies number; }",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends boolean>(x: T) { return !!x; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends boolean>(x: T) { return x; }",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends boolean>(x: T) { return x satisfies boolean; }",
						},
					},
				},
			},
		},
		{
			Code: "function f<T extends 3 | 4>(x: T) { return ~~x; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "function f<T extends 3 | 4>(x: T) { return x; }",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "function f<T extends 3 | 4>(x: T) { return x satisfies number; }",
						},
					},
				},
			},
		},
		{
			Code: "String('a' + 'b').length;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "('a' + 'b').length;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "(('a' + 'b') satisfies string).length;",
						},
					},
				},
			},
		},
		{
			Code: "('a' + 'b').toString().length;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    13,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "('a' + 'b').length;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "(('a' + 'b') satisfies string).length;",
						},
					},
				},
			},
		},
		{
			Code: "2 * +(2 + 2);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    5,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "2 * (2 + 2);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "2 * ((2 + 2) satisfies number);",
						},
					},
				},
			},
		},
		{
			Code: "2 * Number(2 + 2);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    5,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "2 * (2 + 2);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "2 * ((2 + 2) satisfies number);",
						},
					},
				},
			},
		},
		{
			Code: "false && !!(false || true);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    10,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "false && (false || true);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "false && ((false || true) satisfies boolean);",
						},
					},
				},
			},
		},
		{
			Code: "false && Boolean(false || true);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    10,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "false && (false || true);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "false && ((false || true) satisfies boolean);",
						},
					},
				},
			},
		},
		{
			Code: "2n * BigInt(2n + 2n);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    6,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "2n * (2n + 2n);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "2n * ((2n + 2n) satisfies bigint);",
						},
					},
				},
			},
		},
		{
			Code: "\n        let str = 'asdf';\n        String(str).length;\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    9,
					EndLine:   3,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        let str = 'asdf';\n        str.length;\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        let str = 'asdf';\n        (str satisfies string).length;\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\n        let str = 'asdf';\n        str.toString().length;\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    13,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        let str = 'asdf';\n        str.length;\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        let str = 'asdf';\n        (str satisfies string).length;\n      ",
						},
					},
				},
			},
		},
		{
			Code: "~~1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "1;",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "1 satisfies number;",
						},
					},
				},
			},
		},
		{
			Code: "~~-1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Column:    1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "(-1);",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "(-1) satisfies number;",
						},
					},
				},
			},
		},
		{
			Code: "\n        declare const threeOrFour: 3 | 4;\n        ~~threeOrFour;\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeConversion",
					Line:      3,
					Column:    9,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemove",
							Output:    "\n        declare const threeOrFour: 3 | 4;\n        threeOrFour;\n      ",
						},
						{
							MessageId: "suggestSatisfies",
							Output:    "\n        declare const threeOrFour: 3 | 4;\n        threeOrFour satisfies number;\n      ",
						},
					},
				},
			},
		},
	})
}
