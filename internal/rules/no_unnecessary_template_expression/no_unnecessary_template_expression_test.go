package no_unnecessary_template_expression

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTemplateExpressionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTemplateExpressionRule, []rule_tester.ValidTestCase{
		{Code: "const string = 'a';"},
		{Code: "const string = `a`;"},
		{Code: "const string = `NaN: ${/* comment */ NaN}`;"},
		{Code: "const string = `undefined: ${/* comment */ undefined}`;"},
		{Code: "const string = `Infinity: ${Infinity /* comment */}`;"},
		{Code: `
      declare const string: 'a';
      ` + "`" + `${string}b` + "`" + `;
    `},
		{Code: `
      declare const number: 1;
      ` + "`" + `${number}b` + "`" + `;
    `},
		{Code: `
      declare const boolean: true;
      ` + "`" + `${boolean}b` + "`" + `;
    `},
		{Code: `
      declare const nullish: null;
      ` + "`" + `${nullish}-undefined` + "`" + `;
    `},
		{Code: `
      declare const undefinedish: undefined;
      ` + "`" + `${undefinedish}` + "`" + `;
    `},
		{Code: `
      declare const left: 'a';
      declare const right: 'b';
      ` + "`" + `${left}${right}` + "`" + `;
    `},
		{Code: `
      declare const left: 'a';
      declare const right: 'c';
      ` + "`" + `${left}b${right}` + "`" + `;
    `},
		{Code: `
      declare const left: 'a';
      declare const center: 'b';
      declare const right: 'c';
      ` + "`" + `${left}${center}${right}` + "`" + `;
    `},
		{Code: "`1 + 1 = ${1 + 1}`;"},
		{Code: "`true && false = ${true && false}`;"},
		{Code: "tag`${'a'}${'b'}`;"},
		{Code: "`${function () {}}`;"},
		{Code: "`${() => {}}`;"},
		{Code: "`${(...args: any[]) => args}`;"},
		{Code: `
      declare const number: 1;
      ` + "`" + `${number}` + "`" + `;
    `},
		{Code: `
      declare const boolean: true;
      ` + "`" + `${boolean}` + "`" + `;
    `},
		{Code: `
      declare const nullish: null;
      ` + "`" + `${nullish}` + "`" + `;
    `},
		{Code: `
      declare const union: string | number;
      ` + "`" + `${union}` + "`" + `;
    `},
		{Code: `
      declare const unknown: unknown;
      ` + "`" + `${unknown}` + "`" + `;
    `},
		{Code: `
      declare const never: never;
      ` + "`" + `${never}` + "`" + `;
    `},
		{Code: `
      declare const any: any;
      ` + "`" + `${any}` + "`" + `;
    `},
		{Code: `
      function func<T extends number>(arg: T) {
        ` + "`" + `${arg}` + "`" + `;
      }
    `},
		{Code: `
      ` + "`" + `with

      new line` + "`" + `;
    `},
		{Code: `
      declare const a: 'a';

      ` + "`" + `${a} with

      new line` + "`" + `;
    `},
		{Code: "`with windows \r new line`;"},
		{Code: `
` + "`" + `not a useless ${String.raw` + "`" + `nested interpolation ${a}` + "`" + `}` + "`" + `;
    `},
		{Code: `
` + "`" + `
this code has trailing whitespace: ${'    '}
    ` + "`" + `;
    `},
		{Code: `
` + "`" + `
this code has trailing whitespace: ${` + "`" + `    ` + "`" + `}
    ` + "`" + `;
    `},
		{Code: "`this code has trailing whitespace with a windows \\\r new line: ${' '}\r\n`;"},
		{Code: `
` + "`" + `trailing position interpolated empty string also makes whitespace clear    ${''}
` + "`" + `;
    `},
		{Code: `
` + "`" + `
${/* intentional comment before */ 'bar'}
...` + "`" + `;
    `},
		{Code: `
` + "`" + `
${'bar' /* intentional comment after */}
...` + "`" + `;
    `},
		{Code: `
` + "`" + `
${/* intentional comment before */ 'bar' /* intentional comment after */}
...` + "`" + `;
    `},
		{Code: `
` + "`" + `${/* intentional  before */ 'bar'}` + "`" + `;
    `},
		{Code: `
` + "`" + `${'bar' /* intentional comment after */}` + "`" + `;
    `},
		{Code: `
` + "`" + `${/* intentional comment before */ 'bar' /* intentional comment after */}` + "`" + `;
    `},
		{Code: `
` + "`" + `${
  // intentional comment before
  'bar'
}` + "`" + `;
    `},
		{Code: `
` + "`" + `${
  'bar'
  // intentional comment after
}` + "`" + `;
    `},
		{Code: `
      function getTpl<T>(input: T) {
        return ` + "`" + `${input}` + "`" + `;
      }
    `},
		{Code: `
type FooBarBaz = ` + "`" + `foo${/* comment */ 'bar'}"baz"` + "`" + `;
    `},
		{Code: `
enum Foo {
  A = 'A',
  B = 'B',
}
type Foos = ` + "`" + `${Foo}` + "`" + `;
    `},
		{Code: `
type Foo = 'A' | 'B';
type Bar = ` + "`" + `foo${Foo}foo` + "`" + `;
    `},
		{Code: `
type Foo =
  ` + "`" + `trailing position interpolated empty string also makes whitespace clear    ${''}
` + "`" + `;
    `},
		{Code: "type Foo = `this code has trailing whitespace with a windows \\\r new line: ${` `}\r\n`;"},
		{Code: "type Foo = `${'foo' | 'bar' | null}`;"},
		{Code: `
type StringOrNumber = string | number;
type Foo = ` + "`" + `${StringOrNumber}` + "`" + `;
    `},
		{Code: `
enum Foo {
  A = 1,
  B = 2,
}
type Bar = ` + "`" + `${Foo.A}` + "`" + `;
    `},
		{Code: `
enum Enum1 {
  A = 'A1',
  B = 'B1',
}

enum Enum2 {
  A = 'A2',
  B = 'B2',
}

type Union = ` + "`" + `${Enum1 | Enum2}` + "`" + `;
    `},
		{Code: `
enum Enum1 {
  A = 'A1',
  B = 'B1',
}

enum Enum2 {
  A = 'A2',
  B = 'B2',
}

type Union = ` + "`" + `${Enum1.A | Enum2.B}` + "`" + `;
    `},
		{Code: `
enum Enum1 {
  A = 'A1',
  B = 'B1',
}

enum Enum2 {
  A = 'A2',
  B = 'B2',
}
type Enums = Enum1 | Enum2;
type Union = ` + "`" + `${Enums}` + "`" + `;
    `},
		{Code: `
enum Enum {
  A = 'A',
  B = 'A',
}

type Intersection = ` + "`" + `${Enum1.A & string}` + "`" + `;
    `},
		{Code: `
enum Foo {
  A = 'A',
  B = 'B',
}
type Bar = ` + "`" + `${Foo.A}` + "`" + `;
    `},
		{Code: `
function foo<T extends string>() {
  const a: ` + "`" + `${T}` + "`" + ` = 'a';
}
    `},
		{Code: "type T<A extends string> = `${A}`;"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "`${1}`;",
			//       Output: []string{"`1`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 6,
				},
			},
		},
		{
			Code: "`${1n}`;",
			//       Output: []string{"`1`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 7,
				},
			},
		},
		{
			Code: "`${0o25}`;",
			//       Output: []string{"`21`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 9,
				},
			},
		},
		{
			Code: "`${0b1010} ${0b1111}`;",
			//       Output: []string{"`10 15`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    12,
					EndColumn: 21,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 11,
				},
			},
		},
		{
			Code: "`${0x25}`;",
			//       Output: []string{"`37`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 9,
				},
			},
		},
		{
			Code: "`${/a/}`;",
			//       Output: []string{"`/a/`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 8,
				},
			},
		},
		{
			Code: "`${/a/gim}`;",
			//       Output: []string{"`/a/gim`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 11,
				},
			},
		},
		{
			Code: "`${    1    }`;",
			//       Output: []string{"`1`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${    'a'    }`;",
			//       Output: []string{"'a';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${    \"a\"    }`;",
			//       Output: []string{"\"a\";",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${    'a' + 'b'    }`;",
			//       Output: []string{"'a' + 'b';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${true}`;",
			//       Output: []string{"`true`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 9,
				},
			},
		},
		{
			Code: "`${    true    }`;",
			//       Output: []string{"`true`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${null}`;",
			//       Output: []string{"`null`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 9,
				},
			},
		},
		{
			Code: "`${    null    }`;",
			//       Output: []string{"`null`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${undefined}`;",
			//       Output: []string{"`undefined`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 14,
				},
			},
		},
		{
			Code: "`${    undefined    }`;",
			//       Output: []string{"`undefined`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${Infinity}`;",
			//       Output: []string{"`Infinity`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 13,
				},
			},
		},
		{
			Code: "`${NaN}`;",
			//       Output: []string{"`NaN`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 8,
				},
			},
		},
		{
			Code: "`${'a'} ${'b'}`;",
			//       Output: []string{"`a b`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    9,
					EndColumn: 15,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 8,
				},
			},
		},
		{
			Code: "`${   'a'   } ${   'b'   }`;",
			//       Output: []string{"`a b`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`use${'less'}`;",
			//       Output: []string{"`useless`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
				},
			},
		},
		{
			Code: "`use${`less`}`;",
			//       Output: []string{"`useless`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
				},
			},
		},
		{
			Code: `
` + "`" + `u${
  // hopefully this comment is not needed.
  'se'

}${
  ` + "`" + `le${  ` + "`" + `ss` + "`" + `  }` + "`" + `
}` + "`" + `;
      `,
			//       Output: []string{`
			// ` + "`" + `u${
			//   // hopefully this comment is not needed.
			//   'se'
			//
			// }le${  ` + "`" + `ss` + "`" + `  }` + "`" + `;
			//       `,
			// `
			// ` + "`" + `u${
			//   // hopefully this comment is not needed.
			//   'se'
			//
			// }less` + "`" + `;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      6,
					Column:    2,
					EndLine:   8,
					EndColumn: 2,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      7,
					Column:    6,
					EndLine:   7,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
` + "`" + `use${
  ` + "`" + `less` + "`" + `
}` + "`" + `;
      `,
			//       Output: []string{`
			// ` + "`" + `useless` + "`" + `;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      2,
					Column:    5,
					EndLine:   4,
					EndColumn: 2,
				},
			},
		},
		{
			Code: "`${'1 + 1 ='} ${2}`;",
			//       Output: []string{"`1 + 1 = 2`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    15,
					EndColumn: 19,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 14,
				},
			},
		},
		{
			Code: "`${'a'} ${true}`;",
			//       Output: []string{"`a true`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    9,
					EndColumn: 16,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 8,
				},
			},
		},
		{
			Code: "`${String(Symbol.for('test'))}`;",
			//       Output: []string{"String(Symbol.for('test'));",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      1,
					Column:    2,
					EndColumn: 31,
				},
			},
		},
		{
			Code: "`${'`'}`;",
			//       Output: []string{"'`';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`back${'`'}tick`;",
			//       Output: []string{"`back\\`tick`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`dollar${'${`this is test`}'}sign`;",
			//       Output: []string{"`dollar\\${\\`this is test\\`}sign`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`complex${'`${\"`${test}`\"}`'}case`;",
			//       Output: []string{"`complex\\`\\${\"\\`\\${test}\\`\"}\\`case`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${'\\\\${test}'} string`;",
			//       Output: []string{"`some \\\\\\${test} string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${'\\\\`'} string`;",
			//       Output: []string{"`some \\\\\\` string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/`/} string`;",
			//       Output: []string{"`some /\\`/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\`/} string`;",
			//       Output: []string{"`some /\\\\\\`/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\\\`/} string`;",
			//       Output: []string{"`some /\\\\\\\\\\`/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\\\\\`/} string`;",
			//       Output: []string{"`some /\\\\\\\\\\\\\\`/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/${}/} string`;",
			//       Output: []string{"`some /\\${}/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/$ {}/} string`;",
			//       Output: []string{"`some /$ {}/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\\\/} string`;",
			//       Output: []string{"`some /\\\\\\\\/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\\\\\b/} string`;",
			//       Output: []string{"`some /\\\\\\\\\\\\b/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`some ${/\\\\\\\\/} string`;",
			//       Output: []string{"`some /\\\\\\\\\\\\\\\\/ string`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${''} `;",
			//       Output: []string{"`  `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${\"\"} `;",
			//       Output: []string{"`  `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${``} `;",
			//       Output: []string{"`  `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\`'} `;",
			//       Output: []string{"` \\` `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\\\`'} `;",
			//       Output: []string{"` \\\\\\` `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'}{} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\$'}{} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\\\$'}{} `;",
			//       Output: []string{"` \\\\\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\\\$ '}{} `;",
			//       Output: []string{"` \\\\$ {} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\\\\\$'}{} `;",
			//       Output: []string{"` \\\\\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` \\\\${'\\\\$'}{} `;",
			//       Output: []string{"` \\\\\\\\\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${'{$'}{} `;",
			//       Output: []string{"` \\${\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${'${$'}{} `;",
			//       Output: []string{"` $\\${\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'foo$'}{} `;",
			//       Output: []string{"` foo\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`$`} `;",
			//       Output: []string{"` $ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`$`}{} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`$`} {} `;",
			//       Output: []string{"` $ {} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`$`}${undefined}{} `;",
			//       Output: []string{"` $${undefined}{} `;",
			// "` $undefined{} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`foo$`}{} `;",
			//       Output: []string{"` foo\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'}${''}{} `;",
			//       Output: []string{"` \\$${''}{} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'}${``}{} `;",
			//       Output: []string{"` \\$${``}{} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'foo$'}${''}${``}{} `;",
			//       Output: []string{"` foo\\$${''}{} `;",
			// "` foo\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${'{}'} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${undefined}${'{}'} `;",
			//       Output: []string{"` $undefined${'{}'} `;",
			// "` $undefined{} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${''}${undefined}${'{}'} `;",
			//       Output: []string{"` $${undefined}{} `;",
			// "` $undefined{} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` \\$${'{}'} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${'foo'}${'{'} `;",
			//       Output: []string{"` $foo${'{'} `;",
			// "` $foo{ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${'{ foo'}${'{'} `;",
			//       Output: []string{"` \\${ foo${'{'} `;",
			// "` \\${ foo{ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` \\\\$${'{}'} `;",
			//       Output: []string{"` \\\\\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` \\\\\\$${'{}'} `;",
			//       Output: []string{"` \\\\\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` foo$${'{}'} `;",
			//       Output: []string{"` foo\\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${''}${'{}'} `;",
			//       Output: []string{"` \\$${'{}'} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${''} `;",
			//       Output: []string{"` $ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${`{}`} `;",
			//       Output: []string{"` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${``}${`{}`} `;",
			//       Output: []string{"` \\$${`{}`} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${``}${`foo{}`} `;",
			//       Output: []string{"` $${`foo{}`} `;",
			// "` $foo{} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${`${''}${`${``}`}`}${`{a}`} `;",
			//       Output: []string{"` \\$${''}${`${``}`}${`{a}`} `;",
			// "` \\$${``}{a} `;",
			// "` \\${a} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${''}${`{}`} `;",
			//       Output: []string{"` \\$${`{}`} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${``}${'{}'} `;",
			//       Output: []string{"` \\$${'{}'} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` $${''}${``}${'{}'} `;",
			//       Output: []string{"` \\$${``}{} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'} `;",
			//       Output: []string{"` $ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'}${'{}'} `;",
			//       Output: []string{"` \\$${'{}'} `;",
			// "` \\${} `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'$'}${''}${'{'} `;",
			//       Output: []string{"` \\$${''}{ `;",
			// "` \\${ `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `` + "`" + ` ${` + "`" + `
\$` + "`" + `}{} ` + "`" + `;`,
			//       Output: []string{`` + "`" + `
			// \${} ` + "`" + `;`,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `` + "`" + ` ${` + "`" + `
\\$` + "`" + `}{} ` + "`" + `;`,
			//       Output: []string{`` + "`" + `
			// \\\${} ` + "`" + `;`,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${'\\u00E5'}`;",
			//       Output: []string{"'\\u00E5';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "`${'\\n'}`;",
			//       Output: []string{"'\\n';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\u00E5'} `;",
			//       Output: []string{"` \\u00E5 `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\n'} `;",
			//       Output: []string{"` \\n `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${\"\\n\"} `;",
			//       Output: []string{"` \\n `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${`\\n`} `;",
			//       Output: []string{"` \\n `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${ 'A\\u0307\\u0323' } `;",
			//       Output: []string{"` A\\u0307\\u0323 `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'👨‍👩‍👧‍👦'} `;",
			//       Output: []string{"` 👨‍👩‍👧‍👦 `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "` ${'\\ud83d\\udc68'} `;",
			//       Output: []string{"` \\ud83d\\udc68 `;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
` + "`" + `
this code does not have trailing whitespace: ${' '}\n even though it might look it.` + "`" + `;
    `,
			//       Output: []string{`
			// ` + "`" + `
			// this code does not have trailing whitespace:  \n even though it might look it.` + "`" + `;
			//     `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
` + "`" + `
this code has trailing position template expression ${'but it isn\'t whitespace'}
    ` + "`" + `;
    `,
			//       Output: []string{`
			// ` + "`" + `
			// this code has trailing position template expression but it isn\'t whitespace
			//     ` + "`" + `;
			//     `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
` + "`" + `trailing whitespace followed by escaped windows newline: ${' '}\r\n` + "`" + `;
    `,
			//       Output: []string{`
			// ` + "`" + `trailing whitespace followed by escaped windows newline:  \r\n` + "`" + `;
			//     `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
` + "`" + `template literal with interpolations followed by newline: ${` + "`" + ` ${'interpolation'} ` + "`" + `}
` + "`" + `;
    `,
			//       Output: []string{`
			// ` + "`" + `template literal with interpolations followed by newline:  ${'interpolation'}
			// ` + "`" + `;
			//     `,
			// `
			// ` + "`" + `template literal with interpolations followed by newline:  interpolation
			// ` + "`" + `;
			//     `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
        function func<T extends string>(arg: T) {
          ` + "`" + `${arg}` + "`" + `;
        }
      `,
			//       Output: []string{`
			//         function func<T extends string>(arg: T) {
			//           arg;
			//         }
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    12,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
        declare const b: 'b';
        ` + "`" + `a${b}${'c'}` + "`" + `;
      `,
			//       Output: []string{`
			//         declare const b: 'b';
			//         ` + "`" + `a${b}c` + "`" + `;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    15,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
declare const nested: string, interpolation: string;
` + "`" + `use${` + "`" + `less${nested}${interpolation}` + "`" + `}` + "`" + `;
      `,
			//       Output: []string{`
			// declare const nested: string, interpolation: string;
			// ` + "`" + `useless${nested}${interpolation}` + "`" + `;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
        declare const string: 'a';
        ` + "`" + `${   string   }` + "`" + `;
      `,
			//       Output: []string{`
			//         declare const string: 'a';
			//         string;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
        declare const string: 'a';
        ` + "`" + `${string}` + "`" + `;
      `,
			//       Output: []string{`
			//         declare const string: 'a';
			//         string;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    10,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
        declare const intersection: string & { _brand: 'test-brand' };
        ` + "`" + `${intersection}` + "`" + `;
      `,
			//       Output: []string{`
			//         declare const intersection: string & { _brand: 'test-brand' };
			//         intersection;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    10,
					EndColumn: 25,
				},
			},
		},
		{
			Code: "true ? `${'test' || ''}`.trim() : undefined;",
			//       Output: []string{"true ? ('test' || '').trim() : undefined;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type Foo = `${1}`;",
			//       Output: []string{"type Foo = `1`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type Foo = `${null}`;",
			//       Output: []string{"type Foo = `null`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type Foo = `${undefined}`;",
			//       Output: []string{"type Foo = `undefined`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type Foo = `${'foo'}`;",
			//       Output: []string{"type Foo = 'foo';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: `
type Foo = 'A' | 'B';
type Bar = ` + "`" + `${Foo}` + "`" + `;
      `,
			//       Output: []string{`
			// type Foo = 'A' | 'B';
			// type Bar = Foo;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
type Foo = 'A' | 'B';
type Bar = ` + "`" + `${` + "`" + `${Foo}` + "`" + `}` + "`" + `;
      `,
			//       Output: []string{`
			// type Foo = 'A' | 'B';
			// type Bar = ` + "`" + `${Foo}` + "`" + `;
			//       `,
			// `
			// type Foo = 'A' | 'B';
			// type Bar = Foo;
			//       `,
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 24,
				},
				{
					MessageId: "noUnnecessaryTemplateExpression",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: "type FooBarBaz = `foo${'bar'}baz`;",
			//       Output: []string{"type FooBarBaz = `foobarbaz`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type FooBar = `foo${`bar`}`;",
			//       Output: []string{"type FooBar = `foobar`;",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
		{
			Code: "type FooBar = `${'foo' | 'bar'}`;",
			//       Output: []string{"type FooBar = 'foo' | 'bar';",
			// },
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUnnecessaryTemplateExpression",
				},
			},
		},
	})
}
