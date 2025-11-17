package no_unnecessary_type_arguments

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeArguments(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeArgumentsRule, []rule_tester.ValidTestCase{
		{Code: "f<>();"},
		{Code: "f<string>();"},
		{Code: "expect().toBe<>();"},
		{Code: "class Foo extends Bar<> {}"},
		{Code: "class Foo extends Bar<string> {}"},
		{Code: "class Foo implements Bar<> {}"},
		{Code: "class Foo implements Bar<string> {}"},
		{Code: `
function f<T = number>() {}
f();
    `},
		{Code: `
function f<T = number>() {}
f<string>();
    `},
		{Code: `
declare const f: (<T = number>() => void) | null;
f?.();
    `},
		{Code: `
declare const f: (<T = number>() => void) | null;
f?.<string>();
    `},
		{Code: `
declare const f: any;
f();
    `},
		{Code: `
declare const f: any;
f<string>();
    `},
		{Code: `
declare const f: unknown;
f();
    `},
		{Code: `
declare const f: unknown;
f<string>();
    `},
		{Code: `
function g<T = number, U = string>() {}
g<number, number>();
    `},
		{Code: `
declare const g: any;
g<string, string>();
    `},
		{Code: `
declare const g: unknown;
g<string, string>();
    `},
		{Code: `
declare const f: unknown;
f<string>` + "`" + `` + "`" + `;
    `},
		{Code: `
function f<T = number>(template: TemplateStringsArray) {}
f<string>` + "`" + `` + "`" + `;
    `},
		{Code: `
class C<T = number> {}
new C<string>();
    `},
		{Code: `
declare const C: any;
new C<string>();
    `},
		{Code: `
declare const C: unknown;
new C<string>();
    `},
		{Code: `
class C<T = number> {}
class D extends C<string> {}
    `},
		{Code: `
declare const C: any;
class D extends C<string> {}
    `},
		{Code: `
declare const C: unknown;
class D extends C<string> {}
    `},
		{Code: `
interface I<T = number> {}
class Impl implements I<string> {}
    `},
		{Code: `
class C<TC = number> {}
class D<TD = number> extends C {}
    `},
		{Code: `
declare const C: any;
class D<TD = number> extends C {}
    `},
		{Code: `
declare const C: unknown;
class D<TD = number> extends C {}
    `},
		{Code: "let a: A<number>;"},
		{Code: `
class Foo<T> {}
const foo = new Foo<number>();
    `},
		{Code: "type Foo<T> = import('foo').Foo<T>;"},
		{Code: `
class Bar<T = number> {}
class Foo<T = number> extends Bar<T> {}
    `},
		{Code: `
interface Bar<T = number> {}
class Foo<T = number> implements Bar<T> {}
    `},
		{Code: `
class Bar<T = number> {}
class Foo<T = number> extends Bar<string> {}
    `},
		{Code: `
interface Bar<T = number> {}
class Foo<T = number> implements Bar<string> {}
    `},
		{Code: `
import { F } from './missing';
function bar<T = F>() {}
bar<F<number>>();
    `},
		{Code: `
type A<T = Element> = T;
type B = A<HTMLInputElement>;
    `},
		{Code: `
type A<T = Map<string, string>> = T;
type B = A<Map<string, number>>;
    `},
		{Code: `
type A = Map<string, string>;
type B<T = A> = T;
type C2 = B<Map<string, number>>;
    `},
		{Code: `
interface Foo<T = string> {}
declare var Foo: {
  new <T>(type: T): any;
};
class Bar extends Foo<string> {}
    `},
		{Code: `
interface Foo<T = string> {}
class Foo<T> {}
class Bar extends Foo<string> {}
    `},
		{Code: `
class Foo<T = string> {}
interface Foo<T> {}
class Bar implements Foo<string> {}
    `},
		{Code: `
class Foo<T> {}
namespace Foo {
  export class Bar {}
}
class Bar extends Foo<string> {}
    `},
		{
			Code: `
function Button<T>() {
  return <div></div>;
}
const button = <Button<string>></Button>;
      `,
			Tsx: true,
		},
		{
			Code: `
function Button<T>() {
  return <div></div>;
}
const button = <Button<string> />;
      `,
			Tsx: true,
		},
		{Code: `
function f<T = string>() {}
f<any>();
		`},
		{Code: `
function f<T = any>() {}
f<string>();
		`},
		{Code: `
interface Foo {
	foo?: string
}
interface Bar extends Foo {
	bar?: string
}

function f<T = Foo>() {}
f<Bar>();
		`},
		// https://github.com/oxc-project/oxc/issues/13164
		{Code: `
type OneParam<T = any> = T;
interface TestInterface {
  prop?: OneParam<string, number>;  // TypeScript error, but shouldn't panic
}
		`},
		{Code: `
type OneParam<T = any, U = any, V = any> = T;
interface TestInterface {
  prop?: OneParam<string, number>;  // TypeScript error, but shouldn't panic
}
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function f<T = number>() {}
f<number>();
      `,
			Output: []string{`
function f<T = number>() {}
f();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Column:    3,
				},
			},
		},
		{
			Code: `
function g<T = number, U = string>() {}
g<string, string>();
      `,
			Output: []string{`
function g<T = number, U = string>() {}
g<string>();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Column:    11,
				},
			},
		},
		{
			Code: `
function f<T = number>(templates: TemplateStringsArray, arg: T) {}
f<number>` + "`" + `${1}` + "`" + `;
      `,
			Output: []string{`
function f<T = number>(templates: TemplateStringsArray, arg: T) {}
f` + "`" + `${1}` + "`" + `;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Column:    3,
				},
			},
		},
		{
			Code: `
class C<T = number> {}
function h(c: C<number>) {}
      `,
			Output: []string{`
class C<T = number> {}
function h(c: C) {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
class C<T = number> {}
new C<number>();
      `,
			Output: []string{`
class C<T = number> {}
new C();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
class C<T = number> {}
class D extends C<number> {}
      `,
			Output: []string{`
class C<T = number> {}
class D extends C {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
interface I<T = number> {}
class Impl implements I<number> {}
      `,
			Output: []string{`
interface I<T = number> {}
class Impl implements I {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
class Foo<T = number> {}
const foo = new Foo<number>();
      `,
			Output: []string{`
class Foo<T = number> {}
const foo = new Foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
interface Bar<T = string> {}
class Foo<T = number> implements Bar<string> {}
      `,
			Output: []string{`
interface Bar<T = string> {}
class Foo<T = number> implements Bar {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
class Bar<T = string> {}
class Foo<T = number> extends Bar<string> {}
      `,
			Output: []string{`
class Bar<T = string> {}
class Foo<T = number> extends Bar {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
import { F } from './missing';
function bar<T = F<string>>() {}
bar<F<string>>();
      `,
			Output: []string{`
import { F } from './missing';
function bar<T = F<string>>() {}
bar();
      `,
			},
			// TODO(port): why do we need to report on `error` types?
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: `
type DefaultE = { foo: string };
type T<E = DefaultE> = { box: E };
type G = T<DefaultE>;
declare module 'bar' {
  type DefaultE = { somethingElse: true };
  type G = T<DefaultE>;
}
      `,
			Output: []string{`
type DefaultE = { foo: string };
type T<E = DefaultE> = { box: E };
type G = T;
declare module 'bar' {
  type DefaultE = { somethingElse: true };
  type G = T<DefaultE>;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
					Column:    12,
				},
			},
		},
		{
			Code: `
type A<T = Map<string, string>> = T;
type B = A<Map<string, string>>;
      `,
			Output: []string{`
type A<T = Map<string, string>> = T;
type B = A;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      3,
				},
			},
		},
		{
			Code: `
type A = Map<string, string>;
type B<T = A> = T;
type C = B<A>;
      `,
			Output: []string{`
type A = Map<string, string>;
type B<T = A> = T;
type C = B;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
				},
			},
		},
		{
			Code: `
type A = Map<string, string>;
type B<T = A> = T;
type C = B<Map<string, string>>;
      `,
			Output: []string{`
type A = Map<string, string>;
type B<T = A> = T;
type C = B;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
				},
			},
		},
		{
			Code: `
type A = Map<string, string>;
type B = Map<string, string>;
type C<T = A> = T;
type D = C<B>;
      `,
			Output: []string{`
type A = Map<string, string>;
type B = Map<string, string>;
type C<T = A> = T;
type D = C;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      5,
				},
			},
		},
		{
			Code: `
type A = Map<string, string>;
type B = A;
type C = Map<string, string>;
type D = C;
type E<T = B> = T;
type F = E<D>;
      `,
			Output: []string{`
type A = Map<string, string>;
type B = A;
type C = Map<string, string>;
type D = C;
type E<T = B> = T;
type F = E;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface Foo {}
declare var Foo: {
  new <T = string>(type: T): any;
};
class Bar extends Foo<string> {}
      `,
			Output: []string{`
interface Foo {}
declare var Foo: {
  new <T = string>(type: T): any;
};
class Bar extends Foo {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      6,
				},
			},
		},
		{
			Code: `
declare var Foo: {
  new <T = string>(type: T): any;
};
interface Foo {}
class Bar extends Foo<string> {}
      `,
			Output: []string{`
declare var Foo: {
  new <T = string>(type: T): any;
};
interface Foo {}
class Bar extends Foo {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      6,
				},
			},
		},
		{
			Code: `
class Foo<T> {}
interface Foo<T = string> {}
class Bar implements Foo<string> {}
      `,
			Output: []string{`
class Foo<T> {}
interface Foo<T = string> {}
class Bar implements Foo {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
				},
			},
		},
		{
			Code: `
class Foo<T = string> {}
namespace Foo {
  export class Bar {}
}
class Bar extends Foo<string> {}
      `,
			Output: []string{`
class Foo<T = string> {}
namespace Foo {
  export class Bar {}
}
class Bar extends Foo {}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      6,
				},
			},
		},
		{
			Code: `
function Button<T = string>() {
  return <div></div>;
}
const button = <Button<string>></Button>;
      `,
			Output: []string{`
function Button<T = string>() {
  return <div></div>;
}
const button = <Button></Button>;
      `,
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      5,
				},
			},
		},
		{
			Code: `
function Button<T = string>() {
  return <div></div>;
}
const button = <Button<string> />;
      `,
			Output: []string{`
function Button<T = string>() {
  return <div></div>;
}
const button = <Button />;
      `,
			},
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      5,
				},
			},
		},
		{
			Code: `
function foo<T = any>() {}
foo<any>();
      `,
			Output: []string{`
function foo<T = any>() {}
foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      3,
				},
			},
		},
		{
			Code: `
type Foo<T> = any & T
function foo<T = Foo<string>>() {}
foo<Foo<number>>();
      `,
			Output: []string{`
type Foo<T> = any & T
function foo<T = Foo<string>>() {}
foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      4,
				},
			},
		},
		{
			Code: `
declare type MessageEventHandler = ((ev: MessageEvent<any>) => any) | null;
      `,
			Output: []string{`
declare type MessageEventHandler = ((ev: MessageEvent) => any) | null;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryTypeParameter",
					Line:      2,
				},
			},
		},
	})
}
