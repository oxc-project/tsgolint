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
		// https://github.com/typescript-eslint/typescript-eslint/issues/12856
		{Code: `
const a = {};
a as Record<string, string>;
type Dict = Record<string, string>;
a as Dict;
a as { [key: string]: string };
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
		{Code: "const a = `a` as const;"},
		{Code: "const a = 'a' as const;"},
		{Code: "<const>'a';"},
		{Code: `
class T {
  readonly a = 'a' as const;
}
		`},
		{Code: `
enum T {
  Value1,
  Value2,
}
declare const a: T.Value1;
const b = a as const;
		`},
		{Code: `
function filterProps(props: PropertyKey[]): string[] {
  return props.filter((prop) =>
    !['foo', 'bar'].includes(prop as string)
  ) as string[];
}
		`},
		{Code: `
function filterProps(props: PropertyKey[]): string[] {
  return <string[]>props.filter((prop) =>
    !['foo', 'bar'].includes(<string>prop)
  );
}
		`},
		{Code: `
async function mergeWithDefaults(loadModule: () => Promise<unknown>) {
  const mod = (await loadModule()) as Record<string, unknown>;
  return { ...mod, extra: true };
}
		`},
		{Code: `
async function mergeWithDefaults(loadModule: () => Promise<unknown>) {
  const mod = <Record<string, unknown>>(await loadModule());
  return { ...mod, extra: true };
}
		`},
		{Code: `
type Wrapper<T> = { value: number; meta: T };

function unwrap<T>(input: number | string | Wrapper<T>): number {
  return typeof input === 'string' ? parseFloat(input) : (input as number);
}
		`},
		{Code: `
type Wrapper<T> = { value: number; meta: T };

function unwrap<T>(input: number | string | Wrapper<T>): number {
  return typeof input === 'string' ? parseFloat(input) : <number>input;
}
		`},
		{Code: `
const value = ((<T>(input: T): T | undefined => input)(1)) as number;
		`},
		{
			// https://github.com/oxc-project/oxc/issues/20656
			Code: `
interface Element {
  tagName: string;
}

interface HTMLCanvasElement extends Element {
  getContext(contextId: string): unknown;
}

interface HTMLElementTagNameMap {
  canvas: HTMLCanvasElement;
}

declare const document: {
  querySelector<K extends keyof HTMLElementTagNameMap>(selectors: K): HTMLElementTagNameMap[K] | null;
  querySelector<E extends Element = Element>(selectors: string): E | null;
};

export const a = document.querySelector('.foo') as HTMLCanvasElement | null;
		`},
		{
			Code: `
interface Element { tagName: string; }

interface HTMLCanvasElement extends Element { getContext(contextId: string): unknown; }

interface Factory { new <E extends Element = Element>(): E | null; }

declare const CanvasFactory: Factory;

export const a = new CanvasFactory() as HTMLCanvasElement | null;
		`},
		{
			Code: `
interface Element { tagName: string; }

interface HTMLCanvasElement extends Element { getContext(contextId: string): unknown; }

declare const query: { <E extends Element = Element>(strings: TemplateStringsArray): E | null; };

export const a = query` + "`" + `.foo` + "`" + ` as HTMLCanvasElement | null;
		`},
		{Code: `
declare function load<T = unknown>(): Promise<T>;

export async function main() {
  const actual = (await load()) as Record<string, unknown>;
  return { ...actual };
}
		`},
		{Code: `
declare function load<T = unknown>(): Promise<Promise<T>>;

export async function main() {
  const actual = (await await load()) as Record<string, unknown>;
  return { ...actual };
}
		`},
		{Code: `
declare function load<T = unknown>(): Promise<T>;
export async function main() {
  const actual = <Record<string, unknown>>(await load());
  return { ...actual };
}
		`},
		{Code: `
type NumberValueType = number | string;
type NumberValuePairType = [NumberValueType, NumberValueType];

type NumberCellValueType<T extends NumberValuePairType | NumberValueType> =
  T extends NumberValuePairType ? NumberValuePairType : NumberValueType;

function processValue<T extends NumberValuePairType | NumberValueType>(
  value: NumberCellValueType<T>
): number {
  if (Array.isArray(value)) {
    return 0;
  }

  const numberValue = typeof value === "string" ? parseFloat(value) : (value as number);
  //                                                                   ^^^^^^^^^^^^^^^^
  // tsgolint: "This assertion is unnecessary since it does not change the type of the expression."
  const negative = numberValue < 0;
  return negative ? -1 : 1;
}
		`},
		{Code: `const cb = async (importOriginal: unknown) => { const actual = (await importOriginal()) as Record<string, unknown>; return { ...actual, useLocation: vi.fn() }; });`},
		{
			Code: `
type Data<T> = { value?: T };
type ValueType<TData> = TData extends Data<infer T> ? T : never;

export const foo = <TData extends Data<any>>(data: TData) => {
  const getValue = () => data.value as ValueType<TData> | undefined;
  const value: ValueType<TData> = getValue()!;
  return value;
};
    `,
		},
		{
			Code: `
function bar<T extends any>(value: T | undefined): T {
  return value!;
}
    `,
		},
		{
			Code: `
const array: object[] = [{}];

let nullish: object | undefined;
nullish ??= array[1] as object | undefined;

let falsy: object | undefined;
falsy ||= array[1] as object | undefined;

let truthy: object | undefined = {};
truthy &&= array[1] as object | undefined;
    `,
		},
		{
			Code: "const a = <const>'a';",
		},
		{
			Code: `
class T {
  readonly a = 'a' as const;
}
      `,
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
		},
		{
			Code: `
(() => {})() as undefined;
      `,
		},
		{
			Code: `
const f = () => {};
f() as undefined;
      `,
		},
		{
			Code: `
(function () {})() as undefined;
      `,
		},
		{
			Code: `
interface Overloaded {
  (): undefined;
  (value: string): void;
}

((value => {}) as Overloaded)('') as undefined;
      `,
		},
		{
			Code: `
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((() => {}) as Overloaded)() as undefined;
      `,
		},
		{
			Code: `
interface GenericOverloaded {
  <T extends string>(value: T): void;
  (): undefined;
}
((value => {}) as GenericOverloaded)('') as undefined;
      `,
		},
		{
			Code: `
interface Unioned {
  (): undefined | void;
}

((() => {}) as Unioned)() as undefined;
      `,
		},
		{
			Code: `
function fn<T>(items: ReadonlyArray<T>) {}
fn([42] as const);
      `,
		},
		{
			Code: `
declare const a: any;
declare function foo(arg: string): void;
foo(a as string);
    `,
		},
		{
			Code: `
declare const a: object;
const b = a as { id?: number };
    `,
		},
		{
			Code: `
declare const array: any[];
function foo(strings: string[]): void {}
foo(array as string[]);
    `,
		},
		{
			Code: `
declare const record: Record<string, unknown>;
const obj = record as { id?: number };
    `,
		},
		{
			Code: `
declare const obj: { [key: string]: unknown };
const foo = obj as {};
    `,
		},
		{
			Code: `
interface Empty {}
declare function getAny(): any;
const result = getAny() as Empty;
    `,
		},
		{
			Code: `
interface Empty {}
declare function getObject(): object;
const result = getObject() as Empty;
    `,
		},
		{
			Code: `
interface Obj {
  id: number;
}
declare const obj: Readonly<Obj>;
const obj2 = obj as Obj;
    `,
		},
		{
			Code: `
declare const record: Record<string, unknown>;
const obj = record as { [additionalProperties: string]: unknown; id?: number };
    `,
		},
		{
			Code: `
interface PropsA {
  a?: number;
}
interface PropsB extends PropsA {
  b?: string;
}
declare const propsB: PropsB;
const propsA = propsB as PropsA;
    `,
		},
		{
			Code: `
interface PropsA {
  a?: number;
}
interface PropsB extends PropsA {
  b?: string;
}
declare const propsB: PropsB[];
const propsA = propsB as PropsA[];
    `,
		},
		{
			Code: `
class Box<T> {
  value: T;
}
class PairBox<T, U> {
  value: T;
}
declare const pairBox: PairBox<string, number>;
const box = pairBox as Box<string>;
    `,
		},
		{
			Code: `
type ObjectLike = Record<string, unknown>;
declare const result: ObjectLike;
declare const key: string;
result[key] = { ...(result[key] as ObjectLike) };
    `,
		},
		{
			Code: `
interface AST {
  comments: string[] | undefined;
}
const ast: AST = {
  comments: [],
};
const { comments } = ast as { comments: string[] };
    `,
		},
		{
			Code: `
type Tuple = [string | undefined, number];
const tuple: Tuple = ['hello', 42];
const [first, second] = tuple as [string, number];
    `,
		},
		{
			Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
const obj = { value: narrow as Wide } satisfies Record<string, Wide>;
    `,
		},
		{
			Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
const value = narrow as Wide satisfies Wide;
    `,
		},
		{
			Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
declare function identity<T>(x: T): T;
const result = identity({ value: narrow as Wide }) satisfies { value: Wide };
    `,
		},
		{
			Code: `
declare const x: string | number;
const result: { tag: string; value: string | number } | { value: number } = {
  value: x as number,
};
    `,
		},
		{
			Code: `
declare const x: string | number;
function fn(): { tag: string; value: string | number } | { value: number } {
  return {
    value: x as number,
  };
}
    `,
		},
		{
			Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare const a: A;
let result;
result = a as B;
result.b;
    `,
		},
		{
			Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
interface C extends B {
  c: string;
}
declare let a: A;
declare let b: B;
const c = (a = b as C);
c.c;
    `,
		},
		{
			Code: `
type NumberRecord = { readonly [P in number]: number };
function fn<T extends NumberRecord>(record: T) {
  for (const key of Object.keys(record)) {
    const index = +key as keyof T & number;
    record[index] = record[index] + 1;
  }
}
    `,
		},
		{
			Code: `
interface ReadonlyMap<K, V> {
  get(key: K): V | undefined;
}
type T = { get<K>(key: K): K };
declare const x: ReadonlyMap<string, string>;
declare let y: T;
y = x as T;
    `,
		},
		{
			Code: `
declare function find<T>(array: readonly T[] | undefined): T | undefined;
declare const array: string[] | number[];
find(array as (string | number)[]);
    `,
		},
		{
			Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function mapDefined<T>(fn: () => T): T;
declare const b: B;
declare const arrayA: A[];
const a = mapDefined(() => b as A);
[a].concat(arrayA);
    `,
		},
		{
			Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function mapDefined<T>(fn: () => T): T;
declare const b: B;
declare const arrayA: A[];
const a = mapDefined(() =>
  Math.random() > 0.5 ? (b as A) : (null as unknown as A),
);
[a].concat(arrayA);
    `,
		},
		{
			Code: `
interface Params {
  a?: string;
  b?: string;
}
declare const params: Omit<Params, 'a'> & { c?: string };
(params as Params).a = 'c';
    `,
		},
		{
			Code: `
const text: string | null = null as string | null;
if (text) {
  text.toLowerCase();
}
    `,
		},
		{
			Code: `
const text: string | undefined = undefined as string | undefined;
if (text) {
  text.toLowerCase();
}
    `,
		},
		{
			Code: `
type Infer<T> = T extends ObjectConstructor
  ? never
  : T extends () => infer V
    ? V
    : never;
declare function fn<T>(o: { p: T }): { [K in keyof T]: Infer<T[K]> };
const result = fn({ p: { a: Object as () => string } });
result.a.toLowerCase();
    `,
		},
		{
			Code: `
type Accessor<T> = () => T;
declare function inner<T>(): Accessor<T>;
function outer<T>(): Accessor<T> {
  return inner<string>() as Accessor<any>;
}
    `,
		},
		{
			Code: `
interface InjectionConstraint<T> {}
type InjectionKey<T> = symbol & InjectionConstraint<T>;
declare function inject<T>(key: InjectionKey<T>): T;
const context = Symbol('ctx') as InjectionKey<{ value: string }>;
inject(context).value;
    `,
		},
		{
			Code: `
declare function fn<U>(g: (memo: U) => U, initial: U): U;
declare function fn<T>(g: (memo: T) => T): T | undefined;
enum E {
  A = 1,
}
const x: E = fn(n => n | 0, 0 as E);
    `,
		},
		{
			Code: `
type BasePayload = { id: string };

abstract class AbstractHandler {
  constructor(_ctx: { token: number }) {}
}

abstract class AbstractPayloadHandler<
  TPayload extends BasePayload = BasePayload,
> extends AbstractHandler {}

type HandlerCtor = new (ctx: { token: number }) => AbstractHandler;

declare const registeredHandlers: HandlerCtor[];

function example<TItem extends BasePayload>(
  handlerClass: typeof AbstractPayloadHandler<TItem>,
) {
  registeredHandlers.includes(
    handlerClass as new (...args: any[]) => AbstractPayloadHandler<TItem>,
  );
}
    `,
		},
		{
			Code: `
function fn<T extends { type: string }, K extends string, V>(
  node: T,
): T & Record<K, V> {
  return node as T & Record<K, V>;
}
    `,
		},
		{
			Code: `
declare function fn<T extends boolean>(
  options: {
    a: T extends true ? never : unknown;
  } & {
    b: T;
  },
): void;

fn({
  a: true,
  b: true as any,
});
    `,
		},
		{
			Code: `
declare const a: number[] | number | undefined;
const b: number[] = a ?? ([0] as any);
    `,
		},
		{
			Code: `
const context: { meta: Record<string, unknown> | undefined } = { meta: {} };
const meta = context.meta as { schema?: object } | undefined;
    `,
		},
		{
			Code: `
type Test<T extends Record<string, unknown>> = {};

function inferred<T extends Test<never>[]>(_input: {
  addons?: T;
}): {
  options: T extends Test<infer C>[] ? C : never;
} {
  return {
    options: {} as T extends Test<infer C>[] ? C : never,
  };
}

const test = inferred({
  addons: [{} as Test<{ parameters: { potato: boolean } }>],
});

console.log(test.options.parameters.potato);
    `,
		},
		{
			Code: `
declare function fn(value: string | null): void;
fn((null) as string | null);
    `,
		},
		{
			Code: `
const value: undefined = (() => {})() as undefined;
    `,
		},
		{Code: `
type RecursiveConditional<T> =
  (value: T extends string ? any : number) => RecursiveConditional<string>;

declare const callable: RecursiveConditional<number>;
declare function consume(value: unknown): void;
consume(callable as unknown);
    `},
		// Assertions involving nested `any`, recursive callables, or asymmetric type variables are
		// intentionally retained.
		{Code: `
function f<T extends () => any>(value: T) {
  const result = value as NonNullable<T>;
}
    `},
		{Code: `
interface Empty<X> {}
type Alias<T> = T & Empty<any>;

function f<T extends object>(value: T) {
  const result = value as Alias<T>;
}
    `},
		{Code: `
interface Empty<X> {}
type Alias<T> = T & Empty<Promise<any>>;

function f<T extends object>(value: T) {
  const result = value as Alias<T>;
}
    `},
		{Code: `
interface Empty<X> {}
type Alias<T> = T & Empty<() => any>;

function f<T extends object>(value: T) {
  const result = value as Alias<T>;
}
    `},
		{Code: `
function f<T extends { (...args: any[]): unknown }>(value: T) {
  const result = value as NonNullable<T>;
}
    `},
		{Code: `
type RecursiveCallable<T> = (value: T) => RecursiveCallable<{ value: T }>;

function f<T extends RecursiveCallable<string>>(value: T) {
  const result = value as NonNullable<T>;
}
    `},
		{Code: `
interface Empty<X> {}
type Alias<T, U> = T & Empty<U>;

function f<T extends object, U>(value: T) {
  const result = value as Alias<T, U>;
}
    `},
		{Code: `
interface Source<T> {
  value: T;
}
interface Target<T> {
  value: T;
}

declare const source: Source<any>;
const result = source as Target<any>;
    `},
		// Matching type arguments and property names ensure this pair reaches mutual assignability.
		// The nested callable layers make recursive type inspection nontrivial.
		{Code: `
type Layer0<T> = T;
type Layer1<T> = (value: Layer0<T>) => Layer0<T>;
type Layer2<T> = (value: Layer1<T>) => Layer1<T>;
type Layer3<T> = (value: Layer2<T>) => Layer2<T>;
type Layer4<T> = (value: Layer3<T>) => Layer3<T>;

type Watcher<T> = ((handler: Layer4<T>) => void) & { readonly kind: 'watcher' };
type Subscription<T> = ((handler: T) => void) & { readonly kind: 'subscription' };

declare const watcher: Watcher<string>;
const subscription = watcher as Subscription<string>;
    `},
		{Code: `
type RecursiveCallable<Options = {}> =
  & (<NewOptions = {}>(options: NewOptions) => RecursiveCallable<Options & NewOptions>)
  & ((command: string) => void);

declare const callable: RecursiveCallable;
const value = callable as unknown;
const contextualValue: unknown = callable as unknown;
declare function consume(value: unknown): void;
consume(callable as unknown);
    `},
		{
			// https://github.com/oxc-project/tsgolint/issues/1044
			Code: `
type Item = { id: string | number };

declare function combine<T1, T2 extends T1>(a: readonly T1[], b: readonly T2[]): T1[];

declare const items: Item[];

const result = combine([{ id: 0 } as Item], items);
    `,
		},
		{
			Code: `
type Item = { id: string | number };

interface Combiner {
  new <T1, T2 extends T1>(a: readonly T1[], b: readonly T2[]): unknown;
}

declare const Combiner: Combiner;
declare const items: Item[];

const result = new Combiner([{ id: 0 } as Item], items);
    `,
		},
		{
			Code: `
type Item = { id: string | number };

declare function combine<T1, T2 extends T1>(
  ...args: [label: string, a: readonly T1[], b: readonly T2[]]
): T1[];

declare const items: Item[];

const result = combine('items', [{ id: 0 } as Item], items);
    `,
		},
		{
			Code: `
type Item = { id: string | number };

declare function combine<T>(
  enabled: boolean,
  ...args: [a: readonly T[], b: readonly Item[]]
): T[];

declare const items: Item[];

const result = combine(true, [{ id: 0 } as Item], items);
result[0].id = 'x';
    `,
		},
		{
			Code: `
type Item = { id: string | number };

declare function combine<T1, T2 extends T1>(
  ...args: [a: readonly T1[], b: readonly T2[]]
): T1[];

declare const items: Item[];

const result = combine(...[[{ id: 0 } as Item], items]);
result[0].id = 'x';
    `,
		},
		{
			// https://github.com/oxc-project/tsgolint/issues/1047
			Code: `
type IAnyObject = Record<any, any>;

const { isArray } = Array;

const isPlainObject = <InferredType extends IAnyObject = IAnyObject>(
  val: unknown,
): val is InferredType => Object.prototype.toString.call(val) === '[object Object]';

export const deepAssign = <
  TargetObject extends IAnyObject,
  MergedObject extends IAnyObject = TargetObject,
  FinalObject extends Partial<TargetObject> & Partial<MergedObject> = TargetObject & MergedObject,
>(
  originObject: TargetObject,
  ...overrideObjects: (MergedObject | null | undefined)[]
): FinalObject => {
  if (overrideObjects.length === 0) return originObject as unknown as FinalObject;

  const assignObject = overrideObjects.shift();

  if (assignObject) {
    Object.entries(assignObject).forEach(([property, value]) => {
      if (property === '__proto__' || property === 'constructor') return;
      if (isPlainObject(originObject[property]) && isPlainObject(value)) {
        deepAssign(originObject[property], value);
      } else if (isArray(value)) {
        (originObject as IAnyObject)[property] = [...(value as unknown[])];
      } else if (isPlainObject(value)) {
        (originObject as IAnyObject)[property] = {
          ...value,
        };
      } else {
        (originObject as IAnyObject)[property] = assignObject[property] as unknown;
      }
    });
  }

  return deepAssign(originObject, ...overrideObjects);
};
    `,
		},
		{
			// https://github.com/oxc-project/tsgolint/issues/1094
			Code: `
type AccountId = string & { __brand: 'AccountId' };

interface Req {
  params: Record<string, string>;
  query: Record<string, string>;
  method: string;
}

declare const accountId: AccountId;

export const req = {
  params: { id: accountId } as any,
  query: {},
} as Req;
    `,
		},
		{
			Code: `
enum Lang {
  EN = 'en-US',
  SV = 'sv-SE',
}

const LANGUAGE_NAMES: Record<Lang, string> = {
  [Lang.EN]: 'English',
  [Lang.SV]: 'Svenska',
};

export const options = Object.values(Lang)
  .map(lang => ({
    label: LANGUAGE_NAMES[lang],
    value: lang as string,
  }))
  .concat([{ label: 'български език', value: 'bg-BG' }]);
    `,
		},
		{
			Code: `
type AnyObject = Record<string, any>;

declare function isEditArray(item: unknown): item is { type: string }[];

export function getEdits(bucket: { data: unknown }): { type: string }[] {
  const { data } = bucket;
  if (!data) return [];

  const { edits } = data as AnyObject;

  if (!isEditArray(edits)) return [];
  return edits;
}
    `,
		},
		{
			Code: `
enum AppTracker {
  MOOD = 1,
  ENERGY = 2,
}

export function categoryFor(tracker: string): string {
  return AppTracker[Number(tracker) as AppTracker];
}
    `,
			TSConfig: "./tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
enum GuideId {
  MEASURE = 'measure',
}

interface Guide {
  id: GuideId;
  title: string;
  thumbnailUrl: string;
}

interface GlossaryItem {
  id: string;
  title: string;
}

export const item: { content: Guide | GlossaryItem } = {
  content: {
    id: 'how-to-measure' as any,
    title: 'How to measure',
    thumbnailUrl: 'https://example.com/thumb.jpg',
  },
};
    `,
		},
		// Original reproduction from #1122
		{Code: `
interface Node { id: string }
function transform<NodeOut extends Node | null>(
  nodeIn: Node,
  visit: (node: Node) => NodeOut,
): NodeOut {
  let transformedNode = visit(nodeIn);
  if (!transformedNode) return transformedNode;
  if (transformedNode === nodeIn) {
    transformedNode = { ...transformedNode };
  }
  transformedNode = transformedNode as NonNullable<NodeOut>;
  console.log(transformedNode.id);
  return transformedNode;
}
`},
		// Assignment narrowing must survive contextual typing
		{Code: `
function narrow(value: string | undefined) {
  value = value as string;
  return value.length;
}
`},
		// Parenthesized angle-bracket assignment
		{Code: `
function narrow(value: string | undefined) {
  value = ((<string>value));
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  value = value as unknown as string;
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined, condition: boolean) {
  value = condition ? (value as string) : "";
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  value = (value as string) && "";
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  value ||= (value as string);
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  value ??= (value as unknown as string);
  return value.length;
}
`},
		// Numeric enum widening and narrowing from #1122
		{Code: `
enum NumericEnum { a = 1, b = 2 }
const key: "a" | "b" = "a";
export const widened = NumericEnum[key] as number;
export const narrowed = (1 as number) as NumericEnum;
`},
		// Numeric enum union widening and narrowing
		{Code: `
enum NumericEnum { a = 1, b = 2 }
declare const value: NumericEnum | undefined;
declare const numberValue: number | undefined;
export const widened = value as number | undefined;
export const narrowed = numberValue as NumericEnum | undefined;
`},
		// Angle-bracket numeric enum assertions
		{Code: `
enum NumericEnum { a = 1, b = 2 }
declare const value: NumericEnum;
declare const numberValue: number;
export const widened = <number>value;
export const narrowed = <NumericEnum>numberValue;
`},
		{Code: `
function narrow(value: string | undefined) {
  value = (console.log(), value as string);
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  [value] = [value as string];
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined) {
  ({ value } = { value: value as string });
  return value.length;
}
`},
		{Code: `
enum E { A = 1, B = 2 }
declare const value: number & { brand: "x" };
const result = value as E & { brand: "x" };
const onlyMembers: 1 | 2 = result;
`},
		{Code: `
enum E { A = 1, B = 2 }
declare const value: E & { brand: "x" };
const result = value as number & { brand: "x" };
`},
		{Code: `
function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value as string } });
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined, condition: boolean) {
  let x: unknown, y: unknown;
  [x, y, value] = condition
    ? [...["a", "b"] as const, value as string]
    : [undefined, undefined, ""];
  return value.length;
}
`},
		{Code: `
function narrow(value: string | null | undefined) {
  ({ value = '' } = { value: value as string });
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined, fallback: { value?: string }) {
  ({ value } = { ...{ value: value as string }, ...fallback });
  return value.length;
}
`},
		{Code: `
class Fallback {
  get value() { return ''; }
}
function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value as string }, ...new Fallback() });
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined, other: string) {
  value = other && (value as string);
  return value.length;
}
`},
		{Code: `
function narrow(value: string | undefined, a: string | undefined, b: string | undefined) {
  ({ value: a = '', value: b } = { value: value as string });
  return b.length;
}
`},
		{Code: `
function narrow(value: string | undefined, a: string | undefined, b: string | undefined) {
  ({ nested: { value: a = '' }, nested: { value: b } } = { nested: { value: value as string } });
  return b.length;
}
`},
		{Code: `
function narrow(value: string | undefined, a: string | undefined, b: string | undefined) {
  ({ value: a = '', ['value']: b } = { value: value as string });
  return b.length;
}
`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "const foo = <3>3;",
			Output: []string{"const foo = 3;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 17,
				},
			},
		},
		{
			Code:   "const foo = 3 as 3;",
			Output: []string{"const foo = 3;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
const num = 42;
const alsoRedundant = num as 42;
      `,
			Output: []string{`
const num = 42;
const alsoRedundant = num;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    23,
					EndLine:   3,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
const str: string = 'hello';
const redundant =  str as string;
	      `,
			Output: []string{`
const str: string = 'hello';
const redundant =  str;
	      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    20,
					EndLine:   3,
					EndColumn: 33,
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
					EndLine:   3,
					EndColumn: 27,
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
        const foo = 3;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    21,
					EndLine:   3,
					EndColumn: 29,
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
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
const foo = (3 + 5) as number;
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
					EndLine:   2,
					EndColumn: 30,
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
					EndLine:   2,
					EndColumn: 28,
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
const foo = (3 + 5);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 27,
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
					EndLine:   3,
					EndColumn: 25,
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
					Column:    1,
					EndLine:   3,
					EndColumn: 5,
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
					Column:    1,
					EndLine:   3,
					EndColumn: 5,
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
					Column:    1,
					EndLine:   4,
					EndColumn: 5,
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
					Line:      3,
					Column:    21,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code:   "Proxy!;",
			Output: []string{"Proxy;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
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
					Column:    10,
					EndLine:   3,
					EndColumn: 14,
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
					Column:    13,
					EndLine:   3,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
declare const prop: string;
['foo', 'bar'].includes(prop as string);
      `,
			Output: []string{`
declare const prop: string;
['foo', 'bar'].includes(prop);
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    25,
					EndLine:   3,
					EndColumn: 39,
				},
			},
		},
		{
			Code: `
async function mergeWithDefaults(loadModule: () => Promise<Record<string, unknown>>) {
  const mod = (await loadModule()) as Record<string, unknown>;
  return { ...mod, extra: true };
}
      `,
			Output: []string{`
async function mergeWithDefaults(loadModule: () => Promise<Record<string, unknown>>) {
  const mod = (await loadModule());
  return { ...mod, extra: true };
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    15,
					EndLine:   3,
					EndColumn: 62,
				},
			},
		},
		{
			Code: `
function unwrap(input: string | number): number {
  return typeof input === 'string' ? parseFloat(input) : (input as number);
}
      `,
			Output: []string{`
function unwrap(input: string | number): number {
  return typeof input === 'string' ? parseFloat(input) : (input);
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    59,
					EndLine:   3,
					EndColumn: 74,
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
					Column:    9,
					EndLine:   4,
					EndColumn: 11,
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
					Column:    26,
					EndLine:   3,
					EndColumn: 28,
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
					Column:    25,
					EndLine:   4,
					EndColumn: 27,
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
					Column:    6,
					EndLine:   5,
					EndColumn: 8,
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
					Column:    20,
					EndLine:   9,
					EndColumn: 29,
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
					Column:    1,
					EndLine:   5,
					EndColumn: 3,
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
					EndLine:   3,
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
					EndLine:   3,
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
					Column:    11,
					EndLine:   2,
					EndColumn: 22,
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
					EndLine:   2,
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
const a = foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 26,
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
					Column:    11,
					EndLine:   3,
					EndColumn: 24,
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
(foo()).log;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
					Column:    2,
					EndLine:   4,
					EndColumn: 13,
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
					Line:      3,
					Column:    14,
					EndLine:   3,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
const foo = (  3 + 5  ) as number;
      `,
			Output: []string{`
const foo = (  3 + 5  );
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
const foo = (  3 + 5  ) /*as*/ as number;
      `,
			Output: []string{`
const foo = (  3 + 5  ) /*as*/;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 41,
				},
			},
		},
		{
			Code: `// Preserve the comment before as.
const foo = (  3 + 5
  ) /*as*/ as //as
  (
    number
  );
      `,
			Output: []string{`// Preserve the comment before as.
const foo = (  3 + 5
  ) /*as*/;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   6,
					EndColumn: 4,
				},
			},
		},
		{
			Code: `
const foo = (3 + (5 as number) ) as number;
      `,
			Output: []string{`
const foo = (3 + (5 as number) );
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 43,
				},
			},
		},
		{
			Code: `
const foo = 3 + 5/*as*/ as number;
      `,
			Output: []string{`
const foo = 3 + 5/*as*/;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
const foo = 3 + 5/*a*/ /*b*/ as number;
      `,
			Output: []string{`
const foo = 3 + 5/*a*/ /*b*/;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 39,
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
					EndLine:   2,
					EndColumn: 30,
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
					EndLine:   2,
					EndColumn: 36,
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
					EndLine:   2,
					EndColumn: 37,
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
					EndLine:   2,
					EndColumn: 36,
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
					EndLine:   5,
					EndColumn: 18,
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
const bar = foo.a;
      `,
			},
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    13,
					EndLine:   5,
					EndColumn: 40,
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
const bar = foo.a;
      `,
			},
			TSConfig: "./tsconfig.exactOptionalPropertyTypes.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    13,
					EndLine:   5,
					EndColumn: 40,
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
					Column:    1,
					EndLine:   3,
					EndColumn: 3,
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
					Column:    3,
					EndLine:   4,
					EndColumn: 5,
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
  readonly a = 3;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 22,
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
  readonly a = 10;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    16,
					EndLine:   5,
					EndColumn: 23,
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
  readonly a = (3 + 5);
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 33,
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
					Line:      3,
					Column:    31,
					EndLine:   3,
					EndColumn: 51,
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
const b = a;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      8,
					Column:    11,
					EndLine:   8,
					EndColumn: 24,
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
					Line:      3,
					Column:    22,
					EndLine:   3,
					EndColumn: 26,
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
					Line:      4,
					Column:    5,
					EndLine:   4,
					EndColumn: 9,
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
					Line:      7,
					Column:    3,
					EndLine:   7,
					EndColumn: 14,
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
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 18,
				},
			},
		},
		// Tests with checkLiteralConstAssertions: true
		{
			Code:    "const a = true as const;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = true;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 24,
				},
			},
		},
		{
			Code:    "const a = <const>true;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = true;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		{
			Code:    "const a = 1 as const;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code:    "const a = <const>1;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 1;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 19,
				},
			},
		},
		{
			Code:    "const a = 1n as const;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 1n;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		{
			Code:    "const a = <const>1n;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 1n;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 20,
				},
			},
		},
		{
			Code:    "const a = `a` as const;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = `a`;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 23,
				},
			},
		},
		{
			Code:    "const a = 'a' as const;",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 'a';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 23,
				},
			},
		},
		{
			Code:    "const a = <const>'a';",
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output:  []string{"const a = 'a';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
class T {
  readonly a = 'a' as const;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output: []string{`
class T {
  readonly a = 'a';
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 28,
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
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"checkLiteralConstAssertions": true}`),
			Output: []string{`
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      8,
					Column:    11,
					EndLine:   8,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `/** @type {string} */
const s = "foo";

const s2 = /** @type {string} */ (s);
`,
			FileName: "repro.js",
			TSConfig: "./tsconfig.checkJs.json",
			Output: []string{`/** @type {string} */
const s = "foo";

const s2 = (s);
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
					Column:    12,
				},
			},
		},
		// Additional upstream cases from typescript-eslint main.
		{
			Code: `
((): undefined => {})() as undefined;
      `,
			Output: []string{`
((): undefined => {})();
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
(() => 1)() as number;
      `,
			Output: []string{`
(() => 1)();
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((value => {}) as Overloaded)('') as undefined;
      `,
			Output: []string{`
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((value => {}) as Overloaded)('');
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      7,
					Column:    1,
					EndLine:   7,
					EndColumn: 47,
				},
			},
		},
		{
			Code: `
function doThing(a: number) {}
doThing(5 as any);
      `,
			Output: []string{`
function doThing(a: number) {}
doThing(5);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    9,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
interface A {
  required: string;
  alsoRequired: number;
}
function doThing(a: A) {}
doThing({ required: 'yes', alsoRequired: 1 } as any);
      `,
			Output: []string{`
interface A {
  required: string;
  alsoRequired: number;
}
function doThing(a: A) {}
doThing({ required: 'yes', alsoRequired: 1 });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      7,
					Column:    9,
					EndLine:   7,
					EndColumn: 52,
				},
			},
		},
		{
			Code:   "const x = 5 as any as 5;",
			Output: []string{"const x = 5;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
const v: number = 5;
const x = v as unknown as number;
      `,
			Output: []string{`
const v: number = 5;
const x = v;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 33,
				},
			},
		},
		{
			Code: `
const v: number = 5;
const x = v as any as number;
      `,
			Output: []string{`
const v: number = 5;
const x = v;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
const x = (1 + 1) as any as number;
      `,
			Output: []string{`
const x = 1 + 1;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
const x = 2 * ((1 + 1) as any as number);
      `,
			Output: []string{`
const x = 2 * (1 + 1);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      2,
					Column:    16,
					EndLine:   2,
					EndColumn: 40,
				},
			},
		},
		{
			Code: `
const v: number = 5;
const x = <number>(<any>v);
      `,
			Output: []string{`
const v: number = 5;
const x = v;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
const obj = { id: '' };
const obj2 = obj as { id: string };
      `,
			Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    14,
					EndLine:   3,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
const obj = { id: '' };
const obj2 = obj as any as { id: string };
      `,
			Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    14,
					EndLine:   3,
					EndColumn: 42,
				},
			},
		},
		{
			Code: `
const obj = { id: '' };
const obj2 = obj as unknown as { id: string };
      `,
			Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    14,
					EndLine:   3,
					EndColumn: 46,
				},
			},
		},
		{
			Code: `
const array = ['a', 'b'];
const array2 = array as any as string[];
      `,
			Output: []string{`
const array = ['a', 'b'];
const array2 = array;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 40,
				},
			},
		},
		{
			Code: `
const array = ['a', 'b'];
const array2 = array as unknown as string[];
      `,
			Output: []string{`
const array = ['a', 'b'];
const array2 = array;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 44,
				},
			},
		},
		{
			Code: `
type A = 'a';
type B = 'b';
type AorB = A | B;
function fn(aorb: AorB) {}
const a: A = 'a';
fn(a as AorB);
      `,
			Output: []string{`
type A = 'a';
type B = 'b';
type AorB = A | B;
function fn(aorb: AorB) {}
const a: A = 'a';
fn(a);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      7,
					Column:    4,
					EndLine:   7,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
interface Props {
  a: number;
}
const x = { a: 1 } as unknown as Props;
      `,
			Output: []string{`
interface Props {
  a: number;
}
const x = { a: 1 };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    11,
					EndLine:   5,
					EndColumn: 39,
				},
			},
		},
		{
			Code: `
interface Props {
  a: number;
}
const x = { a: 1 } as Props;
      `,
			Output: []string{`
interface Props {
  a: number;
}
const x = { a: 1 };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    11,
					EndLine:   5,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
interface Props {
  a: number;
}
const fn = (): Props => ({ a: 1 }) as unknown as Props;
      `,
			Output: []string{`
interface Props {
  a: number;
}
const fn = (): Props => ({ a: 1 });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    25,
					EndLine:   5,
					EndColumn: 55,
				},
			},
		},
		{
			Code: `
declare function fn(param: number): void;
fn(42 as unknown as number);
      `,
			Output: []string{`
declare function fn(param: number): void;
fn(42);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
declare function fn(param: number): void;
fn(42 as any as number);
      `,
			Output: []string{`
declare function fn(param: number): void;
fn(42);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
declare function fn(params: { param: number });
fn({ param: 42 as number });
      `,
			Output: []string{`
declare function fn(params: { param: number });
fn({ param: 42 });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
declare function fn(params: { param: number });
fn({ param: 42 as any });
      `,
			Output: []string{`
declare function fn(params: { param: number });
fn({ param: 42 });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
type StringOrNumber = string | number;
declare function fn(param: StringOrNumber);
fn(42 as any as StringOrNumber);
      `,
			Output: []string{`
type StringOrNumber = string | number;
declare function fn(param: StringOrNumber);
fn(42);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    4,
					EndLine:   4,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
const data = { a: 1 };
fn({ data: data as NumbersRecord });
      `,
			Output: []string{`
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
const data = { a: 1 };
fn({ data: data });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      5,
					Column:    12,
					EndLine:   5,
					EndColumn: 33,
				},
			},
		},
		{
			Code: `
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
fn({
  data: {
    a: 1,
  } as NumbersRecord,
});
      `,
			Output: []string{`
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
fn({
  data: {
    a: 1,
  },
});
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      5,
					Column:    9,
					EndLine:   7,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
type Json = string | number | boolean | null | { [key: string]: Json } | Json[];
type Tables<T extends 'my_table'> = { my_table: { my_column: Json } }[T];
declare const updatedColumn: Json;
const result = updatedColumn as unknown as Tables<'my_table'>['my_column'];
      `,
			Output: []string{`
type Json = string | number | boolean | null | { [key: string]: Json } | Json[];
type Tables<T extends 'my_table'> = { my_table: { my_column: Json } }[T];
declare const updatedColumn: Json;
const result = updatedColumn;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      5,
					Column:    16,
					EndLine:   5,
					EndColumn: 75,
				},
			},
		},
		{
			Code: `
interface T {
  a: string;
}
declare function fn<U extends T>(args: Pick<U, 'a'>): void;
fn<T>({ a: '' as string });
      `,
			Output: []string{`
interface T {
  a: string;
}
declare function fn<U extends T>(args: Pick<U, 'a'>): void;
fn<T>({ a: '' });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      6,
					Column:    12,
					EndLine:   6,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
declare function update<T extends string>(value: T): void;
update('hi' as unknown as string);
      `,
			Output: []string{`
declare function update<T extends string>(value: T): void;
update('hi');
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    8,
					EndLine:   3,
					EndColumn: 33,
				},
			},
		},
		{
			Code: `
declare function update<T extends string>(value: T): void;
update('hi' as string);
      `,
			Output: []string{`
declare function update<T extends string>(value: T): void;
update('hi');
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    8,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
declare function fn(x: string[]): void;
fn(['hello'] as any);
      `,
			Output: []string{`
declare function fn(x: string[]): void;
fn(['hello']);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update(values: { chat: Json[] }): void;
declare const chat: ChatMessage[];
update({ chat: chat as Json[] });
      `,
			Output: []string{`
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update(values: { chat: Json[] }): void;
declare const chat: ChatMessage[];
update({ chat: chat });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      6,
					Column:    16,
					EndLine:   6,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update<Row extends { chat: Json[] }>(values: Row): void;
declare const chat: ChatMessage[];
update({ chat: chat as Json[] });
      `,
			Output: []string{`
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update<Row extends { chat: Json[] }>(values: Row): void;
declare const chat: ChatMessage[];
update({ chat: chat });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      6,
					Column:    16,
					EndLine:   6,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
interface Node {
  parent: Node;
}
declare function fn<T extends Node>(node: T): T;
function fn2<T extends Node>(node: T): void {
  fn(node as NonNullable<T>);
}
      `,
			Output: []string{`
interface Node {
  parent: Node;
}
declare function fn<T extends Node>(node: T): T;
function fn2<T extends Node>(node: T): void {
  fn(node);
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      7,
					Column:    6,
					EndLine:   7,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function fn(a: A): A;
declare function fn(a: A): A | undefined;
declare const a: A;
fn(a as B);
      `,
			Output: []string{`
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function fn(a: A): A;
declare function fn(a: A): A | undefined;
declare const a: A;
fn(a);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      11,
					Column:    4,
					EndLine:   11,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
declare const a: string[];
declare const b: readonly string[];
const fileNames: string[] = a.concat(b as string[]);
      `,
			Output: []string{`
declare const a: string[];
declare const b: readonly string[];
const fileNames: string[] = a.concat(b);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    38,
					EndLine:   4,
					EndColumn: 51,
				},
			},
		},
		{
			Code: `
declare function fn(text: any): void;
declare const value: string | number;
fn(value as number);
      `,
			Output: []string{`
declare function fn(text: any): void;
declare const value: string | number;
fn(value);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    4,
					EndLine:   4,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
interface A {
  type: 'a';
  a: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a as string,
};
      `,
			Output: []string{`
interface A {
  type: 'a';
  a: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a,
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      12,
					Column:    6,
					EndLine:   12,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
interface A {
  type: 'a';
  a?: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a as string,
};
      `,
			Output: []string{`
interface A {
  type: 'a';
  a?: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a,
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      12,
					Column:    6,
					EndLine:   12,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
declare function fn1<T>(fn: () => void): void;
declare function fn2(text: string): void;
fn1(() => {
  fn2('hi' as any);
});
      `,
			Output: []string{`
declare function fn1<T>(fn: () => void): void;
declare function fn2(text: string): void;
fn1(() => {
  fn2('hi');
});
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      5,
					Column:    7,
					EndLine:   5,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
type Empty<T> = {};
declare function fn(value: Empty<string>): void;
fn({} as Empty<number>);
      `,
			Output: []string{`
type Empty<T> = {};
declare function fn(value: Empty<string>): void;
fn({});
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    4,
					EndLine:   4,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
type Empty<T> = {};
type Box<T> = { value: T };
declare const box: Box<Empty<number>>;
declare function identity<T>(value: T): T;
const result: { item: Box<Empty<string>> } = identity({
  item: box as Box<Empty<boolean>>,
});
      `,
			Output: []string{`
type Empty<T> = {};
type Box<T> = { value: T };
declare const box: Box<Empty<number>>;
declare function identity<T>(value: T): T;
const result: { item: Box<Empty<string>> } = identity({
  item: box,
});
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      7,
					Column:    9,
					EndLine:   7,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
declare const value: (text: string) => string;
const callback: <T extends string>(value: T) => void =
  value as (text: string) => void;`,
			Output: []string{`
declare const value: (text: string) => string;
const callback: <T extends string>(value: T) => void =
  value;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
function f<T extends string>(value: string) {
  const target: string = value as T;
}`,
			Output: []string{`
function f<T extends string>(value: string) {
  const target: string = value;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    26,
					EndLine:   3,
					EndColumn: 36,
				},
			},
		},
		{
			Code: `
declare function fn(value: {}): void;
fn({} as string extends string ? {} : never);
      `,
			Output: []string{`
declare function fn(value: {}): void;
fn({});
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 44,
				},
			},
		},
		{
			Code: `
declare function fn(value: void): void;
fn((() => {})() as undefined);
      `,
			Output: []string{`
declare function fn(value: void): void;
fn((() => {})());
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
declare function fn(value: unknown): void;
fn((() => {})() as undefined);`,
			Output: []string{`
declare function fn(value: unknown): void;
fn((() => {})());`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 29,
				},
			},
		},
		{
			Code:   "const x: string | null = 'a' as string | null;\nvoid x;",
			Output: []string{"const x: string | null = 'a';\nvoid x;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 46,
				},
			},
		},
		{
			Code: `
declare function consume(value: unknown): void;
consume(42 as unknown);
const value: unknown = 42 as unknown;`,
			Output: []string{`
declare function consume(value: unknown): void;
consume(42);
const value: unknown = 42;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    9,
					EndLine:   3,
					EndColumn: 22,
				},
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    24,
					EndLine:   4,
					EndColumn: 37,
				},
			},
		},
		{
			Code:   "declare function f<T>(xs: string[], tag: T): void;\nf(['x' as string], 1);",
			Output: []string{"declare function f<T>(xs: string[], tag: T): void;\nf(['x'], 1);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    4,
					EndLine:   2,
					EndColumn: 17,
				},
			},
		},
		{
			Code:   "type StringsWithMeta<T> = string[] & { meta?: T };\ndeclare function f<T>(xs: StringsWithMeta<T>, tag: T): void;\nf(['x' as string], 1);",
			Output: []string{"type StringsWithMeta<T> = string[] & { meta?: T };\ndeclare function f<T>(xs: StringsWithMeta<T>, tag: T): void;\nf(['x'], 1);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
declare function f(xs: string[]): void;
declare function f<T>(value: T): void;
f(['x' as string]);`,
			Output: []string{`
declare function f(xs: string[]): void;
declare function f<T>(value: T): void;
f(['x']);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      4,
					Column:    4,
					EndLine:   4,
					EndColumn: 17,
				},
			},
		},
		{
			Code: `
declare function f<T>(pair: readonly [T, string], tag: T): void;
f([1, 'x' as string], 1);`,
			Output: []string{`
declare function f<T>(pair: readonly [T, string], tag: T): void;
f([1, 'x'], 1);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    7,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
type AnyRecord = Record<any, any>;
declare const value: AnyRecord;
(value as AnyRecord)['property'] = 1;`,
			Output: []string{`
type AnyRecord = Record<any, any>;
declare const value: AnyRecord;
(value)['property'] = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
					Column:    2,
					EndLine:   4,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
function identity<T>(value: T): T {
  return value as unknown as T;
}`,
			Output: []string{`
function identity<T>(value: T): T {
  return value;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
function upcast<T, U extends T>(value: U): T {
  return value as unknown as T;
}`,
			Output: []string{`
function upcast<T, U extends T>(value: U): T {
  return value;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
function upcast<T, U extends T & { id: string }>(value: U): T {
  return value as unknown as T;
}`,
			Output: []string{`
function upcast<T, U extends T & { id: string }>(value: U): T {
  return value;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
const alreadyNumber: number = 1;
export const genuineNoOp = alreadyNumber as number;`,
			Output: []string{`
const alreadyNumber: number = 1;
export const genuineNoOp = alreadyNumber;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    28,
					EndLine:   3,
					EndColumn: 51,
				},
			},
		},
		{
			Code: `
function unchanged(value: string) {
  value = value as string;
  return value.length;
}`,
			Output: []string{`
function unchanged(value: string) {
  value = value;
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
enum NumericEnum { a = 1, b = 2 }
const value = NumericEnum.a;
export const sameMember = value as NumericEnum.a;`,
			Output: []string{`
enum NumericEnum { a = 1, b = 2 }
const value = NumericEnum.a;
export const sameMember = value;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
					Column:    27,
					EndLine:   4,
					EndColumn: 49,
				},
			},
		},
		{
			Code: `
enum NumericEnum { a = 1, b = 2 }
declare const value: NumericEnum | undefined;
export const sameEnum = value as NumericEnum | undefined;`,
			Output: []string{`
enum NumericEnum { a = 1, b = 2 }
declare const value: NumericEnum | undefined;
export const sameEnum = value;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryAssertion",
					Line:      4,
					Column:    25,
					EndLine:   4,
					EndColumn: 57,
				},
			},
		},
		{
			Code: `let value: number;
value = 1 as number;`,
			Output: []string{`let value: number;
value = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    9,
					EndLine:   2,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `let value: string | undefined;
value = 'x' as string;
console.log(value.length);`,
			Output: []string{`let value: string | undefined;
value = 'x';
console.log(value.length);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    9,
					EndLine:   2,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `let value: unknown;
value = 1 as number;`,
			Output: []string{`let value: unknown;
value = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    9,
					EndLine:   2,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `let value: string | undefined;
[value] = ['x' as string];`,
			Output: []string{`let value: string | undefined;
[value] = ['x'];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    12,
					EndLine:   2,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `let value: string | undefined;
({ value } = { value: 'x' as string });`,
			Output: []string{`let value: string | undefined;
({ value } = { value: 'x' });`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    23,
					EndLine:   2,
					EndColumn: 36,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined) {
  value = (value as string) || "";
  return value.length;
}`,
			Output: []string{`function narrow(value: string | undefined) {
  value = (value) || "";
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    12,
					EndLine:   2,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined) {
  value = (value as string) ?? "";
  return value.length;
}`,
			Output: []string{`function narrow(value: string | undefined) {
  value = (value) ?? "";
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    12,
					EndLine:   2,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined, condition: boolean) {
  value = condition ? value as string : undefined;
}`,
			Output: []string{`function narrow(value: string | undefined, condition: boolean) {
  value = condition ? value : undefined;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    23,
					EndLine:   2,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined, condition: boolean) {
  value = condition ? undefined : value as string;
}`,
			Output: []string{`function narrow(value: string | undefined, condition: boolean) {
  value = condition ? undefined : value;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    35,
					EndLine:   2,
					EndColumn: 50,
				},
			},
		},
		{
			Code: `let value: string | undefined;
({ value } = { ...{ value: 'x' as string } });`,
			Output: []string{`let value: string | undefined;
({ value } = { ...{ value: 'x' } });`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    28,
					EndLine:   2,
					EndColumn: 41,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value as string }, value: '' });
  return value.length;
}`,
			Output: []string{`function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value }, value: '' });
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    30,
					EndLine:   2,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined, condition: boolean) {
  ({ value } = condition ? { value: value as string } : { value: undefined });
}`,
			Output: []string{`function narrow(value: string | undefined, condition: boolean) {
  ({ value } = condition ? { value: value } : { value: undefined });
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    37,
					EndLine:   2,
					EndColumn: 52,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined, condition: boolean) {
  [value] = condition ? [value as string] : [undefined];
}`,
			Output: []string{`function narrow(value: string | undefined, condition: boolean) {
  [value] = condition ? [value] : [undefined];
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    26,
					EndLine:   2,
					EndColumn: 41,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined) {
  ({ value = '' } = { value: value as string });
  return value.length;
}`,
			Output: []string{`function narrow(value: string | undefined) {
  ({ value = '' } = { value: value });
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    30,
					EndLine:   2,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value as string }, ...{ value: '' } });
  return value.length;
}`,
			Output: []string{`function narrow(value: string | undefined) {
  ({ value } = { ...{ value: value }, ...{ value: '' } });
  return value.length;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    30,
					EndLine:   2,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `function narrow(value: string | undefined, other: string | undefined) {
  value = other && (value as string);
}`,
			Output: []string{`function narrow(value: string | undefined, other: string | undefined) {
  value = other && (value);
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "contextuallyUnnecessary",
					Line:      2,
					Column:    21,
					EndLine:   2,
					EndColumn: 36,
				},
			},
		},
	})
}

// Source: typescript-eslint 90241a75df6cffc5d1d4dc088854f5e756f09027.
func TestNoUnnecessaryTypeAssertionUpstreamReporting(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTypeAssertionRule, []rule_tester.ValidTestCase{
		{Code: "const value = (1 + 2) as (number);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = <(number)>(1 + 2);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = (1 + 2) as (/*keep*/ number);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = <(/*keep*/ number)>(1 + 2);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = (1 + 2) as /*before*/ (/*inside*/ number /*after*/);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = (1 + 2) as /*keep*/ number;", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},
		{Code: "const value = </*keep*/ number>(1 + 2);", Options: rule_tester.OptionsFromJSON[NoUnnecessaryTypeAssertionOptions](`{"typesToIgnore": ["number"]}`)},

		{Code: "let value: string; (value)!.length;"},
		{Code: `function narrow(value: string | undefined) {
  value = ((value!));
  return value.length;
 }`},
		{Code: "type ValuePath = 'values' | `values.${string}`;\ndeclare function apply(paths: ValuePath[]): void;\nexport function update(ids: string[]) {\n  apply(ids.map(id => `values.${id}` as ValuePath));\n}"},
		{Code: `declare const items: string[] | undefined;
const counts = items?.reduce((acc, item) => {
  acc[item] = (acc[item] ?? 0) + 1;
  return acc;
}, {} as Record<string, number>);`},
	}, []rule_tester.InvalidTestCase{
		{Code: "const value = (3 as 3);", Output: []string{"const value = (3);"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    16,
				EndLine:   1,
				EndColumn: 22,
			},
		}},
		{Code: "const foo = 3 as 3;", Output: []string{"const foo = 3;"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    13,
				EndLine:   1,
				EndColumn: 19,
			},
		}},
		{Code: "const foo = <3>3;", Output: []string{"const foo = 3;"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    13,
				EndLine:   1,
				EndColumn: 17,
			},
		}},
		{Code: "const s = 'x';\nconst t = s!;", Output: []string{"const s = 'x';\nconst t = s;"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      2,
				Column:    11,
				EndLine:   2,
				EndColumn: 13,
			},
		}},
		{Code: "const foo = (3 + 5) /*before*/ as /*after*/ number;", Output: []string{"const foo = (3 + 5) /*before*/;"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    13,
				EndLine:   1,
				EndColumn: 51,
			},
		}},
		{Code: "<number>{ lol: 32 as number }.lol;", Output: []string{"({ lol: 32 as number }.lol);"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    1,
				EndLine:   1,
				EndColumn: 34,
			},
		}},
		{Code: "<number>function Fun() {}.length;", Output: []string{"(function Fun() {}.length);"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    1,
				EndLine:   1,
				EndColumn: 33,
			},
		}},
		{Code: "<number>class Clazz {}.length;", Output: []string{"(class Clazz {}.length);"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    1,
				EndLine:   1,
				EndColumn: 30,
			},
		}},
		{Code: "const foo = () => <number>{ lol: 123 as number }.lol + 54321;", Output: []string{"const foo = () => ({ lol: 123 as number }.lol) + 54321;"}, Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unnecessaryAssertion",
				Line:      1,
				Column:    19,
				EndLine:   1,
				EndColumn: 53,
			},
		}},
	})
}

func TestNoUnnecessaryTypeAssertionAsyncFunctionFix(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTypeAssertionRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   "<number>async function() {}.length;",
			Output: []string{"(async function() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "<number>async function named() {}.length;",
			Output: []string{"(async function named() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "<number>async function*() {}.length;",
			Output: []string{"(async function*() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "<number>async /* comment */ function() {}.length;",
			Output: []string{"(async /* comment */ function() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "const length = () => <number>async function() {}.length;",
			Output: []string{"const length = () => async function() {}.length;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "<number>(async function() {}.length);",
			Output: []string{"(async function() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "const length = <number>async function() {}.length;",
			Output: []string{"const length = async function() {}.length;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "declare const async: { length: number };\n<number>async.length;",
			Output: []string{"declare const async: { length: number };\nasync.length;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "(<number>async function() {}.length);",
			Output: []string{"(async function() {}.length);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "declare const async: number;\n<number>async\nfunction named() {}",
			Output: []string{"declare const async: number;\nasync\nfunction named() {}"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
	})
}
