package restrict_plus_operands

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestRestrictPlusOperands(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RestrictPlusOperandsRule, []rule_tester.ValidTestCase{
		{Code: "let x = 5;"},
		{Code: "let y = '10';"},
		{Code: "let z = 8.2;"},
		{Code: "let w = '6.5';"},
		{Code: "let foo = 5 + 10;"},
		{Code: "let foo = '5.5' + '10';"},
		{Code: "let foo = parseInt('5.5', 10) + 10;"},
		{Code: "let foo = parseFloat('5.5', 10) + 10;"},
		{Code: "let foo = 1n + 1n;"},
		{Code: "let foo = BigInt(1) + 1n;"},
		{Code: `
      let foo = 1n;
      foo + 2n;
    `},
		{Code: `
function test(s: string, n: number): number {
  return 2;
}
let foo = test('5.5', 10) + 10;
    `},
		{Code: `
let x = 5;
let z = 8.2;
let foo = x + z;
    `},
		{Code: `
let w = '6.5';
let y = '10';
let foo = y + w;
    `},
		{Code: "let foo = 1 + 1;"},
		{Code: "let foo = '1' + '1';"},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = pair.first + 10;
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = pair.first + (10 as number);
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = '5.5' + pair.second;
    `},
		{Code: `
let pair: { first: number; second: string } = { first: 5, second: '10' };
let foo = ('5.5' as string) + pair.second;
    `},
		{Code: `
      const foo =
        'hello' +
        (someBoolean ? 'a' : 'b') +
        (() => (someBoolean ? 'c' : 'd'))() +
        'e';
    `},
		{Code: "const balls = true;"},
		{Code: "balls === true;"},
		{Code: `
function foo<T extends string>(a: T) {
  return a + '';
}
    `},
		{Code: `
function foo<T extends 'a' | 'b'>(a: T) {
  return a + '';
}
    `},
		{Code: `
function foo<T extends number>(a: T) {
  return a + 1;
}
    `},
		{Code: `
function foo<T extends 1>(a: T) {
  return a + 1;
}
    `},
		{Code: `
declare const a: {} & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: string & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: 'string literal' & string;
declare const b: string;
const x = a + b;
    `},
		{Code: `
declare const a: {} & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: number & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: 42 & number;
declare const b: number;
const x = a + b;
    `},
		{Code: `
declare const a: {} & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: unknown & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: bigint & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
declare const a: 42n & bigint;
declare const b: bigint;
const x = a + b;
    `},
		{Code: `
function A(s: string) {
  return ` + "`" + `a${s}b` + "`" + ` as const;
}
const b = A('') + '!';
    `},
		{Code: `
declare const a: ` + "`" + `template${string}` + "`" + `;
declare const b: '';
const x = a + b;
    `},
		{Code: `
const a: ` + "`" + `template${0}` + "`" + `;
declare const b: '';
const x = a + b;
    `},
		{
			Code: `
        declare const a: RegExp;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(true),
			},
		},
		{
			Code: `
        const a = /regexp/;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(true),
			},
		},
		{
			Code: `
const f = (a: RegExp, b: RegExp) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowRegExp: utils.Ref(true)},
		},
		{
			Code: `
let foo: string | undefined;
foo = foo + 'some data';
      `,
			Options: RestrictPlusOperandsOptions{AllowNullish: utils.Ref(true)},
		},
		{
			Code: `
let foo: string | null;
foo = foo + 'some data';
      `,
			Options: RestrictPlusOperandsOptions{AllowNullish: utils.Ref(true)},
		},
		{
			Code: `
let foo: string | null | undefined;
foo = foo + 'some data';
      `,
			Options: RestrictPlusOperandsOptions{AllowNullish: utils.Ref(true)},
		},
		{
			Code: `
let foo = '';
foo += 0;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:                utils.Ref(false),
				AllowBoolean:            utils.Ref(false),
				AllowNullish:            utils.Ref(false),
				AllowNumberAndString:    utils.Ref(false),
				AllowRegExp:             utils.Ref(false),
				SkipCompoundAssignments: utils.Ref(true),
			},
		},
		{
			Code: `
let foo = 0;
foo += '';
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:                utils.Ref(false),
				AllowBoolean:            utils.Ref(false),
				AllowNullish:            utils.Ref(false),
				AllowNumberAndString:    utils.Ref(false),
				AllowRegExp:             utils.Ref(false),
				SkipCompoundAssignments: utils.Ref(true),
			},
		},
		{
			Code: `
const f = (a: any, b: any) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true)},
		},
		{
			Code: `
const f = (a: any, b: string) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true)},
		},
		{
			Code: `
const f = (a: any, b: bigint) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true)},
		},
		{
			Code: `
const f = (a: any, b: number) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true)},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true), AllowBoolean: utils.Ref(true)},
		},
		{
			Code: `
const f = (a: string, b: string | number) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(true),
				AllowBoolean:         utils.Ref(true),
				AllowNullish:         utils.Ref(true),
				AllowNumberAndString: utils.Ref(true),
				AllowRegExp:          utils.Ref(true),
			},
		},
		{
			Code: `
const f = (a: string | number, b: number) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(true),
				AllowBoolean:         utils.Ref(true),
				AllowNullish:         utils.Ref(true),
				AllowNumberAndString: utils.Ref(true),
				AllowRegExp:          utils.Ref(true),
			},
		},
		{
			Code: `
const f = (a: string | number, b: string | number) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(true),
				AllowBoolean:         utils.Ref(true),
				AllowNullish:         utils.Ref(true),
				AllowNumberAndString: utils.Ref(true),
				AllowRegExp:          utils.Ref(true),
			},
		},
		{
			Code:    "let foo = '1' + 1n;",
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(true)},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "let foo = '1' + 1;",
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = '1' + 1;",
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = [] + {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
				{
					MessageId: "invalid",
					Line:      1,
					Column:    16,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = 5 + '10';",
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = [] + 5;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
			},
		},
		{
			Code: "let foo = [] + [];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    11,
					EndColumn: 13,
				},
				{
					MessageId: "invalid",
					Line:      1,
					Column:    16,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = 5 + [3];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    15,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "let foo = '5' + {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      1,
					Column:    17,
					EndColumn: 19,
				},
			},
		},
		{
			Code:    "let foo = 5.5 + '5';",
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code:    "let foo = '5.5' + 5;",
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let y = '10';
let foo = x + y;
      `,
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let y = '10';
let foo = y + x;
      `,
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
let x = 5;
let foo = x + {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
let y = '10';
let foo = [] + y;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
let pair = { first: 5, second: '10' };
let foo = pair + pair;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    11,
					EndColumn: 15,
				},
				{
					MessageId: "invalid",
					Line:      3,
					Column:    18,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
type Valued = { value: number };
let value: Valued = { value: 0 };
let combined = value + 0;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    16,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "let foo = 1n + 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let foo = 1 + 1n;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: `
        let foo = 1n;
        foo + 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
        let foo = 1;
        foo + 1n;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bigintAndNumber",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
function foo<T extends string>(a: T) {
  return a + 1;
}
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends 'a' | 'b'>(a: T) {
  return a + 1;
}
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends number>(a: T) {
  return a + '';
}
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo<T extends 1>(a: T) {
  return a + '';
}
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
        declare const a: ` + "`" + `template${number}` + "`" + `;
        declare const b: number;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: boolean & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: any & string;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        interface A {
          a: 1;
        }
        declare const a: A;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      7,
					Column:    19,
				},
			},
		},
		{
			Code: `
        interface A {
          a: 1;
        }
        interface A2 extends A {
          b: 2;
        }
        declare const a: A2;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      10,
					Column:    19,
				},
			},
		},
		{
			Code: `
        type A = { a: 1 } & { b: 2 };
        declare const a: A;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      5,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: number;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: never;
        declare const b: bigint;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: any;
        declare const b: bigint;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: { a: 1 } & { b: 2 };
        declare const b: bigint;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: RegExp;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        const a = /regexp/;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: Symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        declare const a: unique symbol;
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
        const a = Symbol('');
        declare const b: string;
        const x = a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      4,
					Column:    19,
				},
			},
		},
		{
			Code: `
let foo: string | undefined;
foo += 'some data';
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:                utils.Ref(false),
				AllowBoolean:            utils.Ref(false),
				AllowNullish:            utils.Ref(false),
				AllowNumberAndString:    utils.Ref(false),
				AllowRegExp:             utils.Ref(false),
				SkipCompoundAssignments: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo: string | null;
foo += 'some data';
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo: string = '';
foo += 1;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
let foo = 0;
foo += '';
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:             utils.Ref(false),
				AllowBoolean:         utils.Ref(false),
				AllowNullish:         utils.Ref(false),
				AllowNumberAndString: utils.Ref(false),
				AllowRegExp:          utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true), AllowBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: []) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false), AllowBoolean: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: any) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    31,
				},
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: string) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: bigint) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: number) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    34,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: boolean) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowAny: utils.Ref(false), AllowBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    35,
				},
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code: `
const f = (a: number, b: RegExp) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{AllowRegExp: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    41,
				},
			},
		},
		{
			Code: `
let foo: string | boolean;
foo = foo + 'some data';
      `,
			Options: RestrictPlusOperandsOptions{AllowBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
let foo: boolean;
foo = foo + 'some data';
      `,
			Options: RestrictPlusOperandsOptions{AllowBoolean: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
const f = (a: any, b: unknown) => a + b;
      `,
			Options: RestrictPlusOperandsOptions{
				AllowAny:     utils.Ref(true),
				AllowBoolean: utils.Ref(true),
				AllowNullish: utils.Ref(true),
				AllowRegExp:  utils.Ref(true),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalid",
					Line:      2,
					Column:    39,
				},
			},
		},
		{
			Code:    "let foo = '1' + 1n;",
			Options: RestrictPlusOperandsOptions{AllowNumberAndString: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatched",
					Line:      1,
					Column:    11,
				},
			},
		},
	})
}
