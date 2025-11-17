package no_unnecessary_type_assertion

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeAssertion(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: `
import { TSESTree } from '@typescript-eslint/utils';
declare const member: TSESTree.TSEnumMember;
if (
  member.id.type === AST_NODE_TYPES.Literal &&
  typeof member.id.value === 'string'
) {
  const name = member.id as TSESTree.StringLiteral;
}
    `},
		{Code: `
      const c = 1;
      let z = c as number;
    `},
		{Code: `
      const c = 1;
      let z = c as const;
    `},
		{Code: `
      const c = 1;
      let z = c as 1;
    `},
		{Code: `
      type Bar = 'bar';
      const data = {
        x: 'foo' as 'foo',
        y: 'bar' as Bar,
      };
    `},
		{Code: "[1, 2, 3, 4, 5].map(x => [x, 'A' + x] as [number, string]);"},
		{Code: `
      let x: Array<[number, string]> = [1, 2, 3, 4, 5].map(
        x => [x, 'A' + x] as [number, string],
      );
    `},
		{Code: "let y = 1 as 1;"},
		{Code: "const foo = 3 as number;"},
		{Code: "const foo = <number>3;"},
		{Code: `
type Tuple = [3, 'hi', 'bye'];
const foo = [3, 'hi', 'bye'] as Tuple;
    `},
		{Code: `
type PossibleTuple = {};
const foo = {} as PossibleTuple;
    `},
		{Code: `
type PossibleTuple = { hello: 'hello' };
const foo = { hello: 'hello' } as PossibleTuple;
    `},
		{Code: `
type PossibleTuple = { 0: 'hello'; 5: 'hello' };
const foo = { 0: 'hello', 5: 'hello' } as PossibleTuple;
    `},
		{Code: `
let bar: number | undefined = x;
let foo: number = bar!;
    `},
		{Code: `
declare const a: { data?: unknown };

const x = a.data!;
    `},
		{Code: `
declare function foo(arg?: number): number | void;
const bar: number = foo()!;
    `},
		{
			Code: `
type Foo = number;
const foo = (3 + 5) as Foo;
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["Foo"]}`),
		},
		{
			Code:    "const foo = (3 + 5) as any;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["any"]}`),
		},
		{
			Code:    "(Syntax as any).ArrayExpression = 'foo';",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["any"]}`),
		},
		{
			Code:    "const foo = (3 + 5) as string;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["string"]}`),
		},
		{
			Code: `
type Foo = number;
const foo = <Foo>(3 + 5);
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["Foo"]}`),
		},
		{Code: `
let bar: number;
bar! + 1;
    `},
		{Code: `
let bar: undefined | number;
bar! + 1;
    `},
		{Code: `
let bar: number, baz: number;
bar! + 1;
    `},
		{Code: `
function foo<T extends string | undefined>(bar: T) {
  return bar!;
}
    `},
		{Code: `
declare function nonNull(s: string);
let s: string | null = null;
nonNull(s!);
    `},
		{Code: `
const x: number | null = null;
const y: number = x!;
    `},
		{Code: `
const x: number | null = null;
class Foo {
  prop: number = x!;
}
    `},
		{Code: `
class T {
  a = 'a' as const;
}
    `},
		{Code: `
class T {
  a = 3 as 3;
}
    `},
		{Code: `
const foo = 'foo';

class T {
  readonly test = ` + "`" + `${foo}` + "`" + ` as const;
}
    `},
		{Code: `
class T {
  readonly a = { foo: 'foo' } as const;
}
    `},
		{Code: `
      declare const y: number | null;
      console.log(y!);
    `},
		{Code: `
declare function foo(str?: string): void;
declare const str: string | null;

foo(str!);
    `},
		{Code: `
declare function a(a: string): any;
declare const b: string | null;
class Mx {
  @a(b!)
  private prop = 1;
}
    `},
		{Code: `
function testFunction(_param: string | undefined): void {
  /* noop */
}
const value = 'test' as string | null | undefined;
testFunction(value!);
    `},
		{Code: `
function testFunction(_param: string | null): void {
  /* noop */
}
const value = 'test' as string | null | undefined;
testFunction(value!);
    `},
		{
			Code: `
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: null | string | number }) {
  return <div key={props.id!} />;
}
      `,
			Tsx: true,
		},
		{
			Code: `
const a = [1, 2];
const b = [3, 4];
const c = [...a, ...b] as const;
      `,
		},
		{
			Code: "const a = [1, 2] as const;",
		},
		{
			Code: "const a = { foo: 'foo' } as const;",
		},
		{
			Code: `
const a = [1, 2];
const b = [3, 4];
const c = <const>[...a, ...b];
      `,
		},
		{
			Code: "const a = <const>[1, 2];",
		},
		{
			Code: "const a = <const>{ foo: 'foo' };",
		},
		{
			Code: `
let a: number | undefined;
let b: number | undefined;
let c: number;
a = b;
c = b!;
a! -= 1;
      `,
		},
		{
			Code: `
let a: { b?: string } | undefined;
a!.b = '';
      `,
		},
		{Code: `
let value: number | undefined;
let values: number[] = [];

value = values.pop()!;
    `},
		{Code: `
declare function foo(): number | undefined;
const a = foo()!;
    `},
		{Code: `
declare function foo(): number | undefined;
const a = foo() as number;
    `},
		{Code: `
declare function foo(): number | undefined;
const a = <number>foo();
    `},
		{Code: `
declare const arr: (object | undefined)[];
const item = arr[0]!;
    `},
		{Code: `
declare const arr: (object | undefined)[];
const item = arr[0] as object;
    `},
		{Code: `
declare const arr: (object | undefined)[];
const item = <object>arr[0];
    `},
		{
			Code: `
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]!);
  }
}
      `,
			TSConfig: "./tsconfig.noUncheckedIndexedAccess.json",
		},
		{Code: `
declare const myString: 'foo';
const templateLiteral = ` + "`" + `${myString}-somethingElse` + "`" + ` as const;
    `},
		{Code: `
declare const myString: 'foo';
const templateLiteral = <const>` + "`" + `${myString}-somethingElse` + "`" + `;
    `},
		{Code: `
const myString = 'foo';
const templateLiteral = ` + "`" + `${myString}-somethingElse` + "`" + ` as const;
    `},
		{Code: "let a = `a` as const;"},
		{
			Code: `
declare const foo: {
  a?: string;
};
const bar = foo.a as string;
      `,
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a as string;
      `,
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: {
  a: string;
};
const bar = foo.a as string | undefined;
      `,
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: {
  a?: string | null | number;
};
const bar = foo.a as string | undefined;
      `,
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: {
  a?: string | number;
};
const bar = foo.a as string | undefined | bigint;
      `,
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
if (Math.random()) {
  {
    var x = 1;
  }
}
x!;
      `,
		},
		{
			Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T.Value2;
      `,
		},
		{
			Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T;
      `,
		},
		{
			Code: `
enum T {
  Value1 = 0,
  Value2 = 1,
}

const b = 1 as T.Value2;
      `,
		},
		{Code: `
const foo: unknown = {};
const baz: {} = foo!;
    `},
		{Code: `
const foo: unknown = {};
const bar: object = foo!;
    `},
		{Code: `
declare function foo<T extends unknown>(bar: T): T;
const baz: unknown = {};
foo(baz!);
    `},
		{Code: `
declare const foo: any;
foo!;
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "const a = `a` as const;",
			Output: []string{"const a = `a`  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
				},
			},
		},
		{
			Code:   "const a = 'a' as const;",
			Output: []string{"const a = 'a'  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
				},
			},
		},
		{
			Code:   "const a = <const>'a';",
			Output: []string{"const a = 'a';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
				},
			},
		},
		{
			Code:   "const foo = <3>3;",
			Output: []string{"const foo = 3;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:   "const foo = 3 as 3;",
			Output: []string{"const foo = 3  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `
        type Foo = 3;
        const foo = <Foo>3;
      `,
			Output: []string{`
        type Foo = 3;
        const foo = 3;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    21,
				},
			},
		},
		{
			Code: `
        type Foo = 3;
        const foo = 3 as Foo;
      `,
			Output: []string{`
        type Foo = 3;
        const foo = 3  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    21,
				},
			},
		},
		{
			Code: `
const foo = 3;
const bar = foo!;
      `,
			Output: []string{`
const foo = 3;
const bar = foo;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = (3 + 5) as number;
      `,
			Output: []string{`
const foo = (3 + 5)  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = <number>(3 + 5);
      `,
			Output: []string{`
const foo = (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
type Foo = number;
const foo = (3 + 5) as Foo;
      `,
			Output: []string{`
type Foo = number;
const foo = (3 + 5)  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
type Foo = number;
const foo = <Foo>(3 + 5);
      `,
			Output: []string{`
type Foo = number;
const foo = (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
let bar: number = 1;
bar! + 1;
      `,
			Output: []string{`
let bar: number = 1;
bar + 1;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
let bar!: number;
bar! + 1;
      `,
			Output: []string{`
let bar!: number;
bar + 1;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
let bar: number | undefined;
bar = 1;
bar! + 1;
      `,
			Output: []string{`
let bar: number | undefined;
bar = 1;
bar + 1;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
				},
			},
		},
		{
			Code: `
        declare const y: number;
        console.log(y!);
      `,
			Output: []string{`
        declare const y: number;
        console.log(y);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code:   "Proxy!;",
			Output: []string{"Proxy;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
function foo<T extends string>(bar: T) {
  return bar!;
}
      `,
			Output: []string{`
function foo<T extends string>(bar: T) {
  return bar;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare const foo: Foo;
const bar = <Foo>foo;
      `,
			Output: []string{`
declare const foo: Foo;
const bar = foo;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare function nonNull(s: string | null);
let s: string | null = null;
nonNull(s!);
      `,
			Output: []string{`
declare function nonNull(s: string | null);
let s: string | null = null;
nonNull(s);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
				},
			},
		},
		{
			Code: `
const x: number | null = null;
const y: number | null = x!;
      `,
			Output: []string{`
const x: number | null = null;
const y: number | null = x;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
				},
			},
		},
		{
			Code: `
const x: number | null = null;
class Foo {
  prop: number | null = x!;
}
      `,
			Output: []string{`
const x: number | null = null;
class Foo {
  prop: number | null = x;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
				},
			},
		},
		{
			Code: `
declare function a(a: string): any;
const b = 'asdf';
class Mx {
  @a(b!)
  private prop = 1;
}
      `,
			Output: []string{`
declare function a(a: string): any;
const b = 'asdf';
class Mx {
  @a(b)
  private prop = 1;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
				},
			},
		},
		{
			Code: `
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: string | number }) {
  return <div key={props.id!} />;
}
      `,
			Output: []string{`
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: string | number }) {
  return <div key={props.id} />;
}
      `,
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      9,
				},
			},
		},
		{
			Code: `
let x: number | undefined;
let y: number | undefined;
y = x!;
y! = 0;
      `,
			Output: []string{`
let x: number | undefined;
let y: number | undefined;
y = x!;
y = 0;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      5,
				},
			},
		},
		{
			Code: `
declare function foo(arg?: number): number | void;
const bar: number | void = foo()!;
      `,
			Output: []string{`
declare function foo(arg?: number): number | void;
const bar: number | void = foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    28,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
declare function foo(): number;
const a = foo()!;
      `,
			Output: []string{`
declare function foo(): number;
const a = foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
const b = new Date()!;
      `,
			Output: []string{`
const b = new Date();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
				},
			},
		},
		{
			Code: `
const b = (1 + 1)!;
      `,
			Output: []string{`
const b = (1 + 1);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    11,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
declare function foo(): number;
const a = foo() as number;
      `,
			Output: []string{`
declare function foo(): number;
const a = foo()  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
declare function foo(): number;
const a = <number>foo();
      `,
			Output: []string{`
declare function foo(): number;
const a = foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
type RT = { log: () => void };
declare function foo(): RT;
(foo() as RT).log;
      `,
			Output: []string{`
type RT = { log: () => void };
declare function foo(): RT;
(foo()  ).log;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
declare const arr: object[];
const item = arr[0]!;
      `,
			Output: []string{`
declare const arr: object[];
const item = arr[0];
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
const foo = (  3 + 5  ) as number;
      `,
			Output: []string{`
const foo = (  3 + 5  )  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = (  3 + 5  ) /*as*/ as number;
      `,
			Output: []string{`
const foo = (  3 + 5  ) /*as*/  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = (  3 + 5
  ) /*as*/ as //as
  (
    number
  );
      `,
			Output: []string{`
const foo = (  3 + 5
  ) /*as*/  //as
  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = (3 + (5 as number) ) as number;
      `,
			Output: []string{`
const foo = (3 + (5 as number) )  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = 3 + 5/*as*/ as number;
      `,
			Output: []string{`
const foo = 3 + 5/*as*/  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = 3 + 5/*a*/ /*b*/ as number;
      `,
			Output: []string{`
const foo = 3 + 5/*a*/ /*b*/  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = <(number)>(3 + 5);
      `,
			Output: []string{`
const foo = (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = < ( number ) >( 3 + 5 );
      `,
			Output: []string{`
const foo = ( 3 + 5 );
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = <number> /* a */ (3 + 5);
      `,
			Output: []string{`
const foo =  /* a */ (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
const foo = <number /* a */>(3 + 5);
      `,
			Output: []string{`
const foo = (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
				},
			},
		},
		{
			Code: `
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]!);
  }
}
      `,
			Output: []string{`
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]);
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    9,
				},
			},
		},
		{
			Code: `
declare const foo: {
  a?: string;
};
const bar = foo.a as string | undefined;
      `,
			Output: []string{`
declare const foo: {
  a?: string;
};
const bar = foo.a  ;
      `,
			},
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    13,
				},
			},
		},
		{
			Code: `
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a as string | undefined;
      `,
			Output: []string{`
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a  ;
      `,
			},
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    13,
				},
			},
		},
		{
			Code: `
varDeclarationFromFixture!;
      `,
			Output: []string{`
varDeclarationFromFixture;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
				},
			},
		},
		{
			Code: `
var x = 1;
x!;
      `,
			Output: []string{`
var x = 1;
x;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
var x = 1;
{
  x!;
}
      `,
			Output: []string{`
var x = 1;
{
  x;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
				},
			},
		},
		{
			Code: `
class T {
  readonly a = 'a' as const;
}
      `,
			Output: []string{`
class T {
  readonly a = 'a'  ;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
class T {
  readonly a = 3 as 3;
}
      `,
			Output: []string{`
class T {
  readonly a = 3  ;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
type S = 10;

class T {
  readonly a = 10 as S;
}
      `,
			Output: []string{`
type S = 10;

class T {
  readonly a = 10  ;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
				},
			},
		},
		{
			Code: `
class T {
  readonly a = (3 + 5) as number;
}
      `,
			Output: []string{`
class T {
  readonly a = (3 + 5)  ;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
				},
			},
		},
		{
			Code: `
const a = '';
const b: string | undefined = (a ? undefined : a)!;
      `,
			Output: []string{`
const a = '';
const b: string | undefined = (a ? undefined : a);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
				},
			},
		},
		{
			Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T.Value1;
      `,
			Output: []string{`
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as const;
      `,
			Output: []string{`
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
const foo: unknown = {};
const bar: unknown = foo!;
      `,
			Output: []string{`
const foo: unknown = {};
const bar: unknown = foo;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
				},
			},
		},
		{
			Code: `
function foo(bar: unknown) {}
const baz: unknown = {};
foo(baz!);
      `,
			Output: []string{`
function foo(bar: unknown) {}
const baz: unknown = {};
foo(baz);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
				},
			},
		},
		{
			Code: `
declare const foo: string | RegExp;

declare function isString(v: unknown): v is string

if (isString(foo)) {
  <string>foo;
}
			`,
			Output: []string{`
declare const foo: string | RegExp;

declare function isString(v: unknown): v is string

if (isString(foo)) {
  foo;
}
			`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
		{
			Code: `
class Foo extends Promise {}
declare const bar: Promise<Foo>;
<Promise<Foo>>bar;
			`,
			Output: []string{`
class Foo extends Promise {}
declare const bar: Promise<Foo>;
bar;
			`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
				},
			},
		},
	})
}
