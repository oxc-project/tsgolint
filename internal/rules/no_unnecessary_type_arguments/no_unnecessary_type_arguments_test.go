package no_unnecessary_type_arguments

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeArguments(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeArgumentsRule, []rule_tester.ValidTestCase{
		{Code: "f();"},
		{Code: "f<string>();"},
		{Code: "class Foo extends Bar {}"},
		{Code: "class Foo extends Bar<string> {}"},
		{Code: "class Foo implements Bar {}"},
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
		{
			Code: `
type A<T = Element> = T;
type B = A<HTMLInputElement>;
      `,
			TSConfig: "tsconfig.lib-dom.json",
		},
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

		// OXC-specific cases not present in upstream.
		{Code: `
function f<T>(x: T) {}
f(10);
    `},
		{Code: `
function f<T>(x: T) {}
f<10>(10);
    `},
		{Code: `
function f<T>(x: T) {}
declare const x: any;
f<string>(x);
    `},
		{Code: `
function f<T>(x: T) {}
f<Record<string, boolean>>({});
    `},
		{Code: `
function f<T>(x: T) {}
declare const x: {};
f<Record<string, boolean>>(x);
    `},
		{Code: `
function f<T>(x: T) {}
declare const x: Record<string, never>;
f<Record<string, boolean>>(x);
    `},
		{Code: `
function f<T>(x: T) {}
declare const x: any;
f<{}>(x);
    `},
		{Code: `
function f<T>(x: T) {}
declare const x: {};
f<any>(x);
    `},
		{Code: `
function f<T>(x: T) {}
interface F {}
declare const x: {};
f<F>(x);
    `},
		{Code: `
function f<T>(x: T) {}
f<number[]>([]);
    `},
		{Code: `
function f<T = number>(x: T) {}
f(10);
    `},
		{Code: `
function f<T extends number>(x: T) {}
f(10);
    `},
		{Code: `
function f<T extends number | string>(x: T) {}
f(10);
    `},
		{Code: `
function f<T extends number | string>(x: T) {}
f<number | string>(10);
    `},
		{Code: `
const curried =
  <Outer,>(outer: Outer) =>
  <Inner,>(inner: Inner) => {};
curried(10)(10);
    `},
		{Code: `
const curried =
  <Outer,>(outer: Outer) =>
  <Inner,>(inner: Inner) => {};
curried<10>(10)<10>(10);
    `},
		{Code: `
declare function f<T>(x: T | (() => T)): [T, (x: T) => void];
declare function f<T>(): [T | undefined, (x: T | undefined) => void];
f(10);
f<number>();
    `},
		{Code: `
function f<T>(x: T) {}
f<boolean | null>(true);
    `},
		{Code: `
class C<T> {}
new C<string>();
    `},
		{Code: `
class Foo<T> {
  constructor<T>(x: T) {}
}
const foo = new Foo(10);
    `},
		// Ignore invalid type arguments.
		{Code: `
function f<T>() {}
f<number, number>();
    `},
		{Code: `
class Foo<T> {
  public constructor(a: any, b: any, c: any, d: any) {}
}
interface Bar {
  val: any;
}
let foo = new Foo<Bar>(0, 0, 0, { val: 0 });
    `},
		{Code: `
function f<T = string>() {}
f<any>();
    `},
		{Code: `
function f<T = any>() {}
f<string>();
    `},
		// Inference reporting was reverted upstream and remains disabled locally.
		{Code: `
declare function useState<T>(initialState: T | (() => T)): [T, (value: T) => void];
const [bookmarkedIds, setBookmarkedIds] = useState<Set<string>>(new Set());
    `},
		{Code: `
declare function useRef<T>(initialValue: T): { current: T };
const activeIndexesRef = useRef<Set<number>>(new Set());
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
		// https://github.com/oxc-project/tsgolint/issues/861
		{Code: `
type Data = Record<never, never>

type LocaleData<T extends Data = Data> = Record<string, T>

interface Data1 { a: string }

type Data2 = Partial<Data1>

type LocaleData2 = LocaleData<Data2>
    `},
		{Code: `
type Data = Record<never, never>

type LocaleData<T extends Data = Data> = Record<string, T>

interface Data1 { a: string }

type Data2 = Partial<Data1>

interface Wrapper {
  value: LocaleData<Data2>
}
    `},
		{Code: `

type SessionDataT = Record<string, any>
type SessionData<T extends SessionDataT = SessionDataT> = Partial<T>;
declare function useSession<T extends SessionData = SessionData>(): any;


export function scoped() {
  interface SessionData {
    userId?: string;
  }

  return useSession<SessionData>()
}
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
		// https://github.com/oxc-project/tsgolint/issues/875
		{Code: `
type SessionDataT = Record<string, any>;
type H3SessionData<T extends SessionDataT = SessionDataT> = T;
interface SessionManager<T extends SessionDataT = SessionDataT> {
  readonly data: H3SessionData<T>;
}
interface SessionConfig {
  password: string;
  name: string;
  cookie: {
    sameSite: 'strict';
    httpOnly: true;
    secure: true;
  };
}
interface EventHandlerRequest {}
declare class H3Event<_RequestT extends EventHandlerRequest = EventHandlerRequest> {}
declare function useSession<T extends SessionDataT = SessionDataT>(
  event: H3Event,
  config: SessionConfig,
): Promise<SessionManager<T>>;

interface SessionData {
  userId?: string;
}

function getSessionConfig(): SessionConfig {
  return {
    password: '',
    name: '',
    cookie: {
      sameSite: 'strict',
      httpOnly: true,
      secure: true,
    },
  };
}

async function useAppSession(event: H3Event) {
  const config = getSessionConfig();

  return useSession<SessionData>(event, config);
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
					Column:    3,
					MessageId: "unnecessaryTypeParameter",
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
					Column:    11,
					MessageId: "unnecessaryTypeParameter",
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
					Column:    3,
					MessageId: "unnecessaryTypeParameter",
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
			// TODO(port): upstream reports here; local checker still treats this as an intrinsic error type.
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    5,
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Column:    12,
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      3,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      5,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      7,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      6,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      6,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      6,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      5,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      5,
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},

		// OXC-specific cases not present in upstream.
		{
			Code: `
function f<T = number>(x: T) {}
f<number>(10);
      `,
			Output: []string{`
function f<T = number>(x: T) {}
f(10);
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
					Line:      3,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      4,
					MessageId: "unnecessaryTypeParameter",
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
					Line:      2,
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
		{
			Code: `
type Default = Record<string, any>;
type Alias = Default;
function f<T = Default>() {}
f<Alias>();
      `,
			Output: []string{`
type Default = Record<string, any>;
type Alias = Default;
function f<T = Default>() {}
f();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Line:      5,
					MessageId: "unnecessaryTypeParameter",
				},
			},
		},
	})
}
