package no_confusing_void_expression

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoConfusingVoidExpression(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingVoidExpressionRule, []rule_tester.ValidTestCase{
		{Code: "() => Math.random();"},
		{Code: "console.log('foo');"},
		{Code: "foo && console.log(foo);"},
		{Code: "foo || console.log(foo);"},
		{Code: "foo ? console.log(true) : console.log(false);"},
		{Code: "console?.log('foo');"},
		{
			Code: `
        () => console.log('foo');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreArrowShorthand": true}`),
		},
		{
			Code: `
        foo => foo && console.log(foo);
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreArrowShorthand": true}`),
		},
		{
			Code: `
        foo => foo || console.log(foo);
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreArrowShorthand": true}`),
		},
		{
			Code: `
        foo => (foo ? console.log(true) : console.log(false));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreArrowShorthand": true}`),
		},
		{
			Code: `
        !void console.log('foo');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        +void (foo && console.log(foo));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        -void (foo || console.log(foo));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        () => void ((foo && void console.log(true)) || console.log(false));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        const x = void (foo ? console.log(true) : console.log(false));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        !(foo && void console.log(foo));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        !!(foo || void console.log(foo));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        const x = (foo && void console.log(true)) || void console.log(false);
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        () => (foo ? void console.log(true) : void console.log(false));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{
			Code: `
        return void console.log('foo');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
		},
		{Code: `
function cool(input: string) {
  return console.log(input), input;
}
    `},
		{
			Code: `
function cool(input: string) {
  return input, console.log(input), input;
}
      `,
			// TODO(port): ts-eslint handles (input, console.log(input)), input differently, so this is a bug
			Skip: true,
		},
		{
			Code: `
function test(): void {
  return console.log('bar');
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const test = (): void => {
  return console.log('bar');
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const test = (): void => console.log('bar');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
function test(): void {
  {
    return console.log('foo');
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const obj = {
  test(): void {
    return console.log('foo');
  },
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
class Foo {
  test(): void {
    return console.log('foo');
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
function test() {
  function nestedTest(): void {
    return console.log('foo');
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = () => void;
const test = (() => console.log()) as Foo;
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = {
  foo: () => void;
};
const test: Foo = {
  foo: () => console.log(),
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const test = {
  foo: () => console.log(),
} as {
  foo: () => void;
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const test: {
  foo: () => void;
} = {
  foo: () => console.log(),
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = {
  foo: { bar: () => void };
};

const test = {
  foo: { bar: () => console.log() },
} as Foo;
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = {
  foo: { bar: () => void };
};

const test: Foo = {
  foo: { bar: () => console.log() },
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type MethodType = () => void;

class App {
  private method: MethodType = () => console.log();
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
interface Foo {
  foo: () => void;
}

function bar(): Foo {
  return {
    foo: () => console.log(),
  };
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = () => () => () => void;
const x: Foo = () => () => () => console.log();
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = {
  foo: () => void;
};

const test = {
  foo: () => console.log(),
} as Foo;
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = () => void;
const test: Foo = () => console.log('foo');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
interface Props {
  onEvent: () => void
}

declare function Component(props: Props): any;

<Component onEvent={() => console.log()} />;
			`,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Tsx:     true,
		},
		{
			Code: `
declare function foo(arg: () => void): void;
foo(() => console.log());
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
declare function foo(arg: (() => void) | (() => string)): void;
foo(() => console.log());
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
declare function foo(arg: (() => void) | (() => string) | string): void;
foo(() => console.log());
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
declare function foo(arg: () => void | string): void;
foo(() => console.log());
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
declare function foo(options: { cb: () => void }): void;
foo({ cb: () => console.log() });
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const obj = {
  foo: { bar: () => console.log() },
} as {
  foo: { bar: () => void };
};
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
function test(): void & void {
  return console.log('foo');
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = void;

declare function foo(): Foo;

function test(): Foo {
  return foo();
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
type Foo = void;
const test = (): Foo => console.log('err');
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
const test: () => any = (): void => console.log();
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
function test(): void | string {
  return console.log('bar');
}
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
		{
			Code: `
export function makeDate(): Date;
export function makeDate(m: number): void;
export function makeDate(m?: number): Date | void {
  if (m !== undefined) {
    return console.log('123');
  }
  return new Date();
}

declare const test: (cb: () => void) => void;

test((() => {
  return console.log('123');
}) as typeof makeDate | (() => string));
      `,
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        const x = console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    19,
				},
			},
		},
		{
			Code: `
        const x = console?.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    19,
				},
			},
		},
		{
			Code: `
        console.error(console.log('foo'));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    23,
				},
			},
		},
		{
			Code: `
        [console.log('foo')];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    10,
				},
			},
		},
		{
			Code: `
        ({ x: console.log('foo') });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    15,
				},
			},
		},
		{
			Code: `
        void console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    14,
				},
			},
		},
		{
			Code: `
        console.log('foo') ? true : false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    9,
				},
			},
		},
		{
			Code: `
        (console.log('foo') && true) || false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    10,
				},
			},
		},
		{
			Code: `
        (cond && console.log('ok')) || console.log('error');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    18,
				},
			},
		},
		{
			Code: `
        !console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Column:    10,
				},
			},
		},
		{
			Code: `
function notcool(input: string) {
  return input, console.log(input);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Line:      3,
					Column:    17,
				},
			},
		},
		{
			Code:   "() => console.log('foo');",
			Output: []string{"() =>{  console.log('foo'); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code: "foo => foo && console.log(foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    15,
				},
			},
		},
		{
			Code:   "(foo: undefined) => foo && console.log(foo);",
			Output: []string{"(foo: undefined) =>{  foo && console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    28,
				},
			},
		},
		{
			Code: "foo => foo || console.log(foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    15,
				},
			},
		},
		{
			Code:   "(foo: undefined) => foo || console.log(foo);",
			Output: []string{"(foo: undefined) =>{  foo || console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    28,
				},
			},
		},
		{
			Code:   "(foo: void) => foo || console.log(foo);",
			Output: []string{"(foo: void) =>{  foo || console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    23,
				},
			},
		},
		{
			Code:   "foo => (foo ? console.log(true) : console.log(false));",
			Output: []string{"foo =>{ foo ? console.log(true) : console.log(false); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    15,
				},
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    35,
				},
			},
		},
		{
			Code: `
        function f() {
          return console.log('foo');
          console.log('bar');
        }
      `,
			Output: []string{`
        function f() {
           console.log('foo');; return;
          console.log('bar');
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Line:      3,
					Column:    18,
				},
			},
		},
		{
			Code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
          console.log('quux')
        }
      `,
			Output: []string{`
        function f() {
          console.log('foo')
          ; ['bar', 'baz'].forEach(console.log); return;
          console.log('quux')
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Line:      4,
					Column:    18,
				},
			},
		},
		{
			Code: `
        function f() {
          console.log('foo');
          return console.log('bar');
        }
      `,
			Output: []string{`
        function f() {
          console.log('foo');
           console.log('bar');
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    18,
				},
			},
		},
		{
			Code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
        }
      `,
			Output: []string{`
        function f() {
          console.log('foo')
          ; ['bar', 'baz'].forEach(console.log)
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    18,
				},
			},
		},
		{
			Code: `
        const f = () => {
          if (cond) {
            return console.error('foo');
          }
          console.log('bar');
        };
      `,
			Output: []string{`
        const f = () => {
          if (cond) {
             console.error('foo');; return;
          }
          console.log('bar');
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Line:      4,
					Column:    20,
				},
			},
		},
		{
			Code: `
        const f = function () {
          if (cond) return console.error('foo');
          console.log('bar');
        };
      `,
			Output: []string{`
        const f = function () {
          if (cond) {  console.error('foo');; return; }
          console.log('bar');
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Line:      3,
					Column:    28,
				},
			},
		},
		{
			Code: `
        const f = function () {
          let num = 1;
          return num ? console.log('foo') : num;
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    24,
				},
			},
		},
		{
			Code: `
        const f = function () {
          let undef = undefined;
          return undef ? console.log('foo') : undef;
        };
      `,
			Output: []string{`
        const f = function () {
          let undef = undefined;
           undef ? console.log('foo') : undef;
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    26,
				},
			},
		},
		{
			Code: `
        const f = function () {
          let num = 1;
          return num || console.log('foo');
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    25,
				},
			},
		},
		{
			Code: `
        const f = function () {
          let bar = void 0;
          return bar || console.log('foo');
        };
      `,
			Output: []string{`
        const f = function () {
          let bar = void 0;
           bar || console.log('foo');
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    25,
				},
			},
		},
		{
			Code: `
        let num = 1;
        const foo = () => (num ? console.log('foo') : num);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    34,
				},
			},
		},
		{
			Code: `
        let bar = void 0;
        const foo = () => (bar ? console.log('foo') : bar);
      `,
			Output: []string{`
        let bar = void 0;
        const foo = () =>{ bar ? console.log('foo') : bar; };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    34,
				},
			},
		},
		{
			Code:    "return console.log('foo');",
			Output:  []string{"return void console.log('foo');"},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnWrapVoid",
					Line:      1,
					Column:    8,
				},
			},
		},
		{
			Code:    "console.error(console.log('foo'));",
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Line:      1,
					Column:    15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "voidExprWrapVoid",
							Output:    "console.error(void console.log('foo'));",
						},
					},
				},
			},
		},
		{
			Code:    "console.log('foo') ? true : false;",
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "voidExprWrapVoid",
							Output:    "void console.log('foo') ? true : false;",
						},
					},
				},
			},
		},
		{
			Code:    "const x = foo ?? console.log('foo');",
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Line:      1,
					Column:    18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "voidExprWrapVoid",
							Output:    "const x = foo ?? void console.log('foo');",
						},
					},
				},
			},
		},
		{
			Code:    "foo => foo || console.log(foo);",
			Output:  []string{"foo => foo || void console.log(foo);"},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrowWrapVoid",
					Line:      1,
					Column:    15,
				},
			},
		},
		{
			Code:    "!!console.log('foo');",
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidOperator": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Line:      1,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "voidExprWrapVoid",
							Output:    "!!void console.log('foo');",
						},
					},
				},
			},
		},
		{
			Code: `
function test() {
  return console.log('foo');
}
      `,
			Output: []string{`
function test() {
   console.log('foo');
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code:    "const test = () => console.log('foo');",
			Output:  []string{"const test = () =>{  console.log('foo'); };"},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      1,
					Column:    20,
				},
			},
		},
		{
			Code: `
const test = () => {
  return console.log('foo');
};
      `,
			Output: []string{`
const test = () => {
   console.log('foo');
};
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function foo(): void {
  const bar = () => {
    return console.log();
  };
}
      `,
			Output: []string{`
function foo(): void {
  const bar = () => {
     console.log();
  };
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    12,
				},
			},
		},
		{
			Code: `
        (): any => console.log('foo');
      `,
			Output: []string{`
        (): any =>{  console.log('foo'); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      2,
					Column:    20,
				},
			},
		},
		{
			Code: `
        (): unknown => console.log('foo');
      `,
			Output: []string{`
        (): unknown =>{  console.log('foo'); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      2,
					Column:    24,
				},
			},
		},
		{
			Code: `
function test(): void {
  () => () => console.log();
}
      `,
			Output: []string{`
function test(): void {
  () => () =>{  console.log(); };
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
type Foo = any;
(): Foo => console.log();
      `,
			Output: []string{`
type Foo = any;
(): Foo =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    12,
				},
			},
		},
		{
			Code: `
type Foo = unknown;
(): Foo => console.log();
      `,
			Output: []string{`
type Foo = unknown;
(): Foo =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    12,
				},
			},
		},
		{
			Code: `
function test(): any {
  () => () => console.log();
}
      `,
			Output: []string{`
function test(): any {
  () => () =>{  console.log(); };
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
function test(): unknown {
  return console.log();
}
      `,
			Output: []string{`
function test(): unknown {
   console.log();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function test(): any {
  return console.log();
}
      `,
			Output: []string{`
function test(): any {
   console.log();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
type Foo = () => any;
(): Foo => () => console.log();
      `,
			Output: []string{`
type Foo = () => any;
(): Foo => () =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    18,
				},
			},
		},
		{
			Code: `
type Foo = () => unknown;
(): Foo => () => console.log();
      `,
			Output: []string{`
type Foo = () => unknown;
(): Foo => () =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    18,
				},
			},
		},
		{
			Code: `
type Foo = () => any;
const test: Foo = () => console.log();
      `,
			Output: []string{`
type Foo = () => any;
const test: Foo = () =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    25,
				},
			},
		},
		{
			Code: `
type Foo = () => unknown;
const test: Foo = () => console.log();
      `,
			Output: []string{`
type Foo = () => unknown;
const test: Foo = () =>{  console.log(); };
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Line:      3,
					Column:    25,
				},
			},
		},
		{
			Code: `
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
			Output: []string{`
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
     console.log();
  }
};
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      6,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
			Output: []string{`
const foo = function () {
  function bar() {
     console.log();
  }
};
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      4,
					Column:    12,
				},
			},
		},
		{
			Code: `
return console.log('foo');
      `,
			Output: []string{`
{  console.log('foo');; return; }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Line:      2,
					Column:    8,
				},
			},
		},
		{
			Code: `
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
			Output: []string{`
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
   console.log();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      8,
					Column:    10,
				},
			},
		},
		{
			Code: `
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
			Output: []string{`
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
   console.log();
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoConfusingVoidExpressionOptions](`{"ignoreVoidReturningFunctions": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Line:      8,
					Column:    10,
				},
			},
		},
	})
}
