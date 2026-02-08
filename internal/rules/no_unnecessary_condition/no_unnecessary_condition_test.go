package no_unnecessary_condition

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestNoUnnecessaryConditionRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Basic boolean conditions
		{Code: `
declare const b1: boolean;
declare const b2: boolean;
const t1 = b1 && b2;
const t2 = b1 || b2;
if (b1 && b2) {
}
while (b1 && b2) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
if (b1 && b2) {
}
while (b1 && b2) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
for (;;) {}
switch (b1) {
  case true:
  default:
}
    `},

		// Function returning void or number
		{Code: `
declare function foo(): number | void;
const result1 = foo() === undefined;
const result2 = foo() == null;
    `},

		// BigInt conditions
		{Code: `
declare const bigInt: 0n | 1n;
if (bigInt) {
}
    `},

		// Truthy/falsy literal combinations
		{Code: `
declare const b1: false | 5;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: boolean | "foo";
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: 0 | boolean;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: boolean | object;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: false | object;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: null | object;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: undefined | true;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: void | true;
declare const b2: boolean;
const t1 = b1 && b2;
`},

		// "Branded" types
		{Code: `
declare const b1: string & {};
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: number & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: boolean & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: bigint & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & {} & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & { __brandA: string } & { __brandB: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & { __brand: string } | number;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: (string | number) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & ({ __brand: string } | number);
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: ("" | "foo") & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: (string & { __brandA: string }) | (number & { __brandB: string });
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string }) & ("" | "foo"));
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: { __brandA: string} & (({ __brandB: string } & string) | ({ __brandC: string } & number));
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: (string | number) & ("foo" | 123 | { __brandA: string });
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: string & string;
declare const b2: boolean;
const t1 = b1 && b2;
`},

		// Any and unknown
		{Code: `
declare const b1: any;
declare const b2: boolean;
const t1 = b1 && b2;
`},
		{Code: `
declare const b1: unknown;
declare const b2: boolean;
const t1 = b1 && b2;
`},

		// Generic type params
		{Code: `
function test<T extends string>(t: T) {
  return t ? 'yes' : 'no';
}
    `},
		{Code: `
// Naked type param
function test<T>(t: T) {
  return t ? 'yes' : 'no';
}
    `},
		{Code: `
// Naked type param in union
function test<T>(t: T | []) {
  return t ? 'yes' : 'no';
}
    `},
		{Code: `
function test<T>(arg: T, key: keyof T) {
  if (arg[key]?.toString()) {
  }
}
    `},
		{Code: `
function test<T>(arg: T, key: keyof T) {
  if (arg?.toString()) {
  }
}
    `},
		{Code: `
function test<T>(arg: T | { value: string }) {
  if (arg?.value) {
  }
}
    `},

		// Boolean expressions
		{Code: `
function test(a: string) {
  const t1 = a === 'a';
  const t2 = 'a' === a;
}
    `},
		{Code: `
function test(a?: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t1 = a !== undefined;
  const t2 = undefined !== a;
}
    `},
		{Code: `
function test(a: null | string) {
  const t1 = a === null;
  const t2 = null === a;
  const t1 = a !== null;
  const t2 = null !== a;
}
    `},
		{Code: `
function test(a?: null | string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `},
		{Code: `
function test(a?: string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `},
		{Code: `
function test(a: null | string) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
}
    `},
		{Code: `
function test(a: any) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `},
		{Code: `
function test(a: unknown) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `},
		{Code: `
function test<T>(a: T) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
    `},
		{Code: `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] == null;
}
    `},

		// Predicate functions
		{Code: `
// with literal arrow function
[0, 1, 2].filter(x => x);

// filter with named function
function length(x: string) {
  return x.length;
}
['a', 'b', ''].filter(length);

// with non-literal array
function nonEmptyStrings(x: string[]) {
  return x.filter(length);
}

// filter-like predicate
function count(
  list: string[],
  predicate: (value: string, index: number, array: string[]) => unknown,
) {
  return list.filter(predicate).length;
}
    `},
		{Code: `
declare const test: <T>() => T;

[1, null].filter(test);
    `},
		{Code: `
declare const test: <T extends boolean>() => T;

[1, null].filter(test);
    `},
		{Code: `
[1, null].filter(1 as any);
    `},
		{Code: `
[1, null].filter(1 as never);
    `},

		// Ignores non-array methods of the same name
		{Code: `
const notArray = {
  filter: (func: () => boolean) => func(),
  find: (func: () => boolean) => func(),
};
notArray.filter(() => true);
notArray.find(() => true);
    `},

		// Nullish coalescing operator
		{Code: `
function test(a: string | null) {
  return a ?? 'default';
}
    `},
		{Code: `
function test(a: string | undefined) {
  return a ?? 'default';
}
    `},
		{Code: `
function test(a: string | null | undefined) {
  return a ?? 'default';
}
    `},
		{Code: `
function test(a: unknown) {
  return a ?? 'default';
}
    `},
		{Code: `
function test<T>(a: T) {
  return a ?? 'default';
}
    `},
		{Code: `
function test<T extends string | null>(a: T) {
  return a ?? 'default';
}
    `},
		{Code: `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] ?? 'default';
}
    `},

		// Indexing cases
		{Code: `
declare const arr: object[];
if (arr[42]) {
} // looks unnecessary from the types, but isn't

const tuple = [{}] as [object];
declare const n: number;
if (tuple[n]) {
}
    `},

		// Optional-chaining indexing
		{Code: `
declare const arr: Array<{ value: string } & (() => void)>;
if (arr[42]?.value) {
}
arr[41]?.();

// An array access can "infect" deeper into the chain
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x?.y?.z;

const tuple = ['foo'] as const;
declare const n: number;
tuple[n]?.toUpperCase();
    `},
		{
			Code: `
declare const arr: Array<{ value: string } & (() => void)>;
if (arr[42]?.value) {
}
arr[41]?.();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{Code: `
if (arr?.[42]) {
}
    `},
		{Code: `
type ItemA = { bar: string; baz: string };
type ItemB = { bar: string; qux: string };
declare const foo: ItemA[] | ItemB[];
foo[0]?.bar;
    `},
		{Code: `
type TupleA = [string, number];
type TupleB = [string, number];

declare const foo: TupleA | TupleB;
declare const index: number;
foo[index]?.toString();
    `},
		{
			Code: `
type TupleA = [string, number];
type TupleB = [string, number];

declare const foo: TupleA | TupleB;
declare const index: number;
foo[index]?.toString();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{Code: `
declare const returnsArr: undefined | (() => string[]);
if (returnsArr?.()[42]) {
}
returnsArr?.()[42]?.toUpperCase();
    `},

		// nullish + array index
		{Code: `
declare const arr: string[][];
arr[x] ?? [];
    `},

		// nullish + optional array index
		{Code: `
declare const arr: { foo: number }[];
const bar = arr[42]?.foo ?? 0;
    `},

		// Doesn't check the right-hand side of a logical expression in a non-conditional context
		{Code: `
declare const b1: boolean;
declare const b2: true;
const x = b1 && b2;
      `},

		// Allow constant loop conditions
		{
			Code: `
while (true) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(true)},
		},
		{
			Code: `
for (; true; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(true)},
		},
		{
			Code: `
do {} while (true);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(true)},
		},
		{
			Code: `
while (true) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "always"},
		},
		{
			Code: `
for (; true; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "always"},
		},
		{
			Code: `
do {} while (true);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "always"},
		},
		{
			Code: `
while (true) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
		},
		{
			Code: `
while (1) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
		},
		{
			Code: `
while (false) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
		},
		{
			Code: `
while (0) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
		},

		// Optional chaining on potentially undefined
		{Code: `
let variable = 'abc' as string | void;
variable?.[0];
    `},
		{Code: `
let foo: undefined | { bar: true };
foo?.bar;
    `},
		{Code: `
let foo: null | { bar: true };
foo?.bar;
    `},
		{Code: `
let foo: undefined;
foo?.bar;
    `},
		{Code: `
let foo: undefined;
foo?.bar.baz;
    `},
		{Code: `
let foo: null;
foo?.bar;
    `},
		{Code: `
let anyValue: any;
anyValue?.foo;
    `},
		{Code: `
let unknownValue: unknown;
unknownValue?.foo;
    `},
		{Code: `
let foo: undefined | (() => {});
foo?.();
    `},
		{Code: `
let foo: null | (() => {});
foo?.();
    `},
		{Code: `
let foo: undefined;
foo?.();
    `},
		{Code: `
let foo: undefined;
foo?.().bar;
    `},
		{Code: `
let foo: null;
foo?.();
    `},
		{Code: `
let anyValue: any;
anyValue?.();
    `},
		{Code: `
let unknownValue: unknown;
unknownValue?.();
    `},
		{Code: "const foo = [1, 2, 3][0];"},
		{Code: `
declare const foo: { bar?: { baz: { c: string } } } | null;
foo?.bar?.baz;
    `},
		{Code: `
foo?.bar?.baz?.qux;
    `},
		{Code: `
declare const foo: { bar: { baz: string } };
foo.bar.qux?.();
    `},
		{Code: `
type Foo = { baz: number } | null;
type Bar = { baz: null | string | { qux: string } };
declare const foo: { fooOrBar: Foo | Bar } | null;
foo?.fooOrBar?.baz?.qux;
    `},
		{Code: `
type Foo = { [key: string]: string } | null;
declare const foo: Foo;

const key = '1';
foo?.[key]?.trim();
    `},
		{Code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo';
declare const foo: Foo;
declare const key: Key;

foo?.[key].trim();
    `},
		{Code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
  };
}

function Foo(outer: Outer, key: string): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `},
		{Code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo';

function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `},
		{Code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo' | 'baz';
declare const foo: Foo;
declare const key: Key;

foo?.[key]?.trim();
    `},

		// Branded keys with index signatures
		{Code: `
type BrandedKey = string & { __brand: string };
type Foo = { [key: BrandedKey]: string } | null;
declare const foo: Foo;
const key = '1' as BrandedKey;
foo?.[key]?.trim();
    `},
		{Code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key].trim();
    `},
		{Code: `
type BrandedKey = string & { __brand: string };
interface Outer {
  inner?: {
    [key: BrandedKey]: string | undefined;
  };
}
function Foo(outer: Outer, key: BrandedKey): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `},
		{Code: `
interface Outer {
  inner?: {
    [key: string & { __brand: string }]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo' & { __brand: string };
function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
    `},
		{Code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key]?.trim();
    `},
		{
			Code: `
type BrandedKey = string & { __brand: string };
type Foo = { [key: BrandedKey]: string } | null;
declare const foo: Foo;
const key = '1' as BrandedKey;
foo?.[key]?.trim();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key].trim();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type BrandedKey = string & { __brand: string };
interface Outer {
  inner?: {
    [key: BrandedKey]: string | undefined;
  };
}
function Foo(outer: Outer, key: BrandedKey): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
interface Outer {
  inner?: {
    [key: string & { __brand: string }]: string | undefined;
    bar: 'bar';
  };
}
type Foo = 'foo' & { __brand: string };
function Foo(outer: Outer, key: Foo): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type BrandedKey<S extends string> = S & { __brand: string };
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;
declare const foo: Foo;
declare const key: Key;
foo?.[key]?.trim();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type Foo = {
  key?: Record<string, { key: string }>;
};
declare const foo: Foo;
foo.key?.someKey?.key;
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type Foo = {
  key?: {
    [key: string]: () => void;
  };
};
declare const foo: Foo;
foo.key?.value?.();
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: `
type A = {
  [name in Lowercase<string>]?: {
    [name in Lowercase<string>]: {
      a: 1;
    };
  };
};

declare const a: A;

a.a?.a?.a;
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},

		// Array indexing edge cases
		{Code: `
let latencies: number[][] = [];

function recordData(): void {
  if (!latencies[0]) latencies[0] = [];
  latencies[0].push(4);
}

recordData();
    `},
		{Code: `
let latencies: number[][] = [];

function recordData(): void {
  if (latencies[0]) latencies[0] = [];
  latencies[0].push(4);
}

recordData();
    `},
		{Code: `
function test(testVal?: boolean) {
  if (testVal ?? true) {
    console.log('test');
  }
}
    `},
		{Code: `
declare const x: string[];
if (!x[0]) {
}
    `},

		// Boolean functions
		{Code: `
const isEven = (val: number) => val % 2 === 0;
if (!isEven(1)) {
}
    `},
		{Code: `
declare const booleanTyped: boolean;
declare const unknownTyped: unknown;

if (!(booleanTyped || unknownTyped)) {
}
    `},

		// Unstrict mode with option
		{
			Code: `
declare const x: string[] | null;
// eslint-disable-next-line
if (x) {
}
      `,
			Options:  NoUnnecessaryConditionOptions{AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(true)},
			TSConfig: "tsconfig.unstrict.json",
		},

		// Index signatures with tuple types
		{Code: `
interface Foo {
  [key: string]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.test?.length;
    `},
		{Code: `
interface Foo {
  [key: number]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.[1]?.length;
    `},

		// Logical assignment operators
		{Code: `
declare let foo: number | null;
foo ??= 1;
    `},
		{Code: `
declare let foo: number;
foo ||= 1;
    `},
		{Code: `
declare const foo: { bar: { baz?: number; qux: number } };
type Key = 'baz' | 'qux';
declare const key: Key;
foo.bar[key] ??= 1;
    `},
		{Code: `
enum Keys {
  A = 'A',
  B = 'B',
}
type Foo = {
  [Keys.A]: number | null;
  [Keys.B]: number;
};
declare const foo: Foo;
declare const key: Keys;
foo[key] ??= 1;
    `},
		{
			Code: `
declare const foo: { bar?: number };
foo.bar ??= 1;
      `,
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: { bar: { baz?: number } };
foo['bar'].baz ??= 1;
      `,
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
		},
		{
			Code: `
declare const foo: { bar: { baz?: number; qux: number } };
type Key = 'baz' | 'qux';
declare const key: Key;
foo.bar[key] ??= 1;
      `,
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
		},
		{Code: `
declare let foo: number;
foo &&= 1;
    `},
		{Code: `
function foo<T extends object>(arg: T, key: keyof T): void {
  arg[key] ??= 'default';
}
    `},

		// Generic indexed access
		{Code: `
function get<Obj, Key extends keyof Obj>(obj: Obj, key: Key) {
  const value = obj[key];
  if (value) {
    return value;
  }
  throw new Error('BOOM!');
}

get({ foo: null }, 'foo');
    `},
		{
			Code: `
function getElem(dict: Record<string, { foo: string }>, key: string) {
  if (dict[key]) {
    return dict[key].foo;
  } else {
    return '';
  }
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},

		// Function return value optional chaining
		{Code: `
type Foo = { bar: () => number | undefined } | null;
declare const foo: Foo;
foo?.bar()?.toExponential();
    `},
		{Code: `
type Foo = (() => number | undefined) | null;
declare const foo: Foo;
foo?.()?.toExponential();
    `},
		{Code: `
type FooUndef = () => undefined;
type FooNum = () => number;
type Foo = FooUndef | FooNum | null;
declare const foo: Foo;
foo?.()?.toExponential();
    `},
		{Code: `
type Foo = { [key: string]: () => number | undefined } | null;
declare const foo: Foo;
foo?.['bar']()?.toExponential();
    `},
		{Code: `
declare function foo(): void | { key: string };
const bar = foo()?.key;
    `},
		{Code: `
type fn = () => void;
declare function foo(): void | fn;
const bar = foo()?.();
    `},

		// Private fields with exact optional property types
		{
			Code: `
class ConsistentRand {
  #rand?: number;

  getCachedRand() {
    this.#rand ??= Math.random();
    return this.#rand;
  }
}
      `,
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
		},

		// Type predicates
		{
			Code: `
declare function assert(x: unknown): asserts x;

assert(Math.random() > 0.5);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function assert(x: unknown, y: unknown): asserts x;

assert(Math.random() > 0.5, true);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert(true);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(false)},
		},
		{
			Code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(true);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
class ThisAsserter {
  assertThis(this: unknown, arg2: unknown): asserts this {}
}

const thisAsserter: ThisAsserter = new ThisAsserter();
thisAsserter.assertThis(Math.random());
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert(...[]);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert(...[], {});
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function assertString(x: unknown): asserts x is string;
declare const a: string;
assertString(a);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(false)},
		},
		{
			Code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString(a);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(false)},
		},
		{
			Code: `
declare function assertString(x: unknown): asserts x is string;
assertString('falafel');
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},
		{
			Code: `
declare function isString(x: unknown): x is string;
isString('falafel');
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
		},

		// Mapped types
		{Code: `
type A = { [name in Lowercase<string>]?: A };
declare const a: A;
a.a?.a?.a;
    `},
		{Code: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a.a.a.value;
t.A?.A?.A?.VALUE;
    `},
	}, []rule_tester.InvalidTestCase{
		// Basic always truthy/falsy cases
		{
			Code: `
const b1 = true;
declare const b2: boolean;
const t1 = b1 && b2;
const t2 = b1 || b2;
if (b1 && b2) {
}
if (b2 && b1) {
}
while (b1 && b2) {}
while (b2 && b1) {}
for (let i = 0; b1 && b2; i++) {
  break;
}
const t1 = b1 && b2 ? 'yes' : 'no';
const t1 = b2 && b1 ? 'yes' : 'no';
switch (b1) {
  case true:
  default:
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysTruthy"},
				{MessageId: "comparisonBetweenLiteralTypes"},
			},
		},

		// Always truthy/falsy type conditions
		{
			Code: `
declare const b1: object;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: object | true;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: "" | false;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: "always truthy";
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: undefined;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: null;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: void;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: never;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}},
		},
		{
			Code: `
declare const b1: string & number;
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}},
		},

		// BigInt literals
		{
			Code: `
declare const falseyBigInt: 0n;
if (falseyBigInt) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const posbigInt: 1n;
if (posbigInt) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const negBigInt: -2n;
if (negBigInt) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Complex logical expressions
		{
			Code: `
declare const b1: boolean;
declare const b2: boolean;
if (true && b1 && b2) {
}
if (b1 && false && b2) {
}
if (b1 || b2 || true) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysFalsy"},
				{MessageId: "alwaysTruthy"},
			},
		},

		// Generic type params
		{
			Code: `
function test<T extends object>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
function test<T extends false>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
function test<T extends 'a' | 'b'>(t: T) {
  return t ? 'yes' : 'no';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Boolean expressions
		{
			Code: `
function test(a: 'a') {
  return a === 'a';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
declare const a: '34';
declare const b: '56';
a > b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
const y = 1;
if (y === 0) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
// @ts-expect-error
if (1 == '1') {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
2.3 > 2.3;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
2.3 >= 2.3;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
2n < 2n;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
2n <= 2n;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
-2n !== 2n;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
// @ts-expect-error
if (1 == '2') {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
// @ts-expect-error
if (1 != '2') {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
enum Foo {
  a = 1,
  b = 2,
}

const x = Foo.a;
if (x === Foo.a) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
enum Foo {
  a = 1,
  b = 2,
}

const x = Foo.a;
if (x === 1) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a == undefined) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a === undefined) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a != undefined) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
function takesMaybeValue(a: null | object) {
  if (a) {
  } else if (a !== undefined) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
true === false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
true === true;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},
		{
			Code: `
true === undefined;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes"}},
		},

		// No overlap boolean expressions
		{
			Code: `
function test(a: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
			},
		},
		{
			Code: `
function test(a?: string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
			},
		},
		{
			Code: `
function test(a: null | string) {
  const t1 = a === undefined;
  const t2 = undefined === a;
  const t3 = a !== undefined;
  const t4 = undefined !== a;
  const t5 = a === null;
  const t6 = null === a;
  const t7 = a !== null;
  const t8 = null !== a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
			},
		},
		{
			Code: `
function test<T extends object>(a: T) {
  const t1 = a == null;
  const t2 = null == a;
  const t3 = a != null;
  const t4 = null != a;
  const t5 = a == undefined;
  const t6 = undefined == a;
  const t7 = a != undefined;
  const t8 = undefined != a;
  const t9 = a === null;
  const t10 = null === a;
  const t11 = a !== null;
  const t12 = null !== a;
  const t13 = a === undefined;
  const t14 = undefined === a;
  const t15 = a !== undefined;
  const t16 = undefined !== a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
				{MessageId: "noOverlapBooleanExpression"},
			},
		},

		// Nullish coalescing operator - never nullish
		{
			Code: `
function test(a: string) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
function test(a: string | false) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
function test<T extends string>(a: T) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
function test(a: { foo: string }[]) {
  return a[0].foo ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},

		// Nullish coalescing operator - always nullish
		{
			Code: `
function test(a: null) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish"}},
		},
		{
			Code: `
function test(a: null[]) {
  return a[0] ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish"}},
		},
		{
			Code: `
function test<T extends null>(a: T) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish"}},
		},
		{
			Code: `
function test(a: never) {
  return a ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}},
		},
		{
			Code: `
function test<T extends { foo: number }, K extends 'foo'>(num: T[K]) {
  num ?? 'default';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},

		// Predicate functions
		{
			Code: `
[1, 3, 5].filter(() => true);
[1, 2, 3].find(() => {
  return false;
});

// with non-literal array
function nothing(x: string[]) {
  return x.filter(() => false);
}
// with readonly array
function nothing2(x: readonly string[]) {
  return x.filter(() => false);
}
// with tuple
function nothing3(x: [string, string]) {
  return x.filter(() => false);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy"},
				{MessageId: "alwaysFalsy"},
				{MessageId: "alwaysFalsy"},
				{MessageId: "alwaysFalsy"},
				{MessageId: "alwaysFalsy"},
			},
		},
		{
			Code: `
declare const test: <T extends true>() => T;

[1, null].filter(test);
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthyFunc"}},
		},

		// Indexing cases
		{
			Code: `
declare const dict: Record<string, object>;
if (dict['mightNotExist']) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
const x = [{}] as [{ foo: string }];
if (x[0]) {
}
if (x[0]?.foo) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
declare const arr: object[];
if (arr.filter) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
function truthy() {
  return [];
}
function falsy() {}
[1, 3, 5].filter(truthy);
[1, 2, 3].find(falsy);
[1, 2, 3].findLastIndex(falsy);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthyFunc"},
				{MessageId: "alwaysFalsyFunc"},
				{MessageId: "alwaysFalsyFunc"},
			},
		},

		// Constant loop conditions
		{
			Code: `
declare const test: true;

while (test) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(false)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

for (; test; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(false)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

do {} while (test);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref(false)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

while (test) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

for (; test; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

do {} while (test);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

while (test) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: 1;

while (test) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

for (; test; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const test: true;

do {} while (test);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
for (; true; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
for (; 0; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
do {} while (0);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
let shouldRun = true;

while ((shouldRun = true)) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
while (2) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
while ('truthy') {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Unnecessary optional chains
		{
			Code: `
let foo = { bar: true };
foo?.bar;
foo ?. bar;
foo ?.
  bar;
foo
  ?. bar;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
let foo = () => {};
foo?.();
foo ?. ();
foo ?.
  ();
foo
  ?. ();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
let foo = () => {};
foo?.(bar);
foo ?. (bar);
foo ?.
  (bar);
foo
  ?. (bar);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code:   "const foo = [1, 2, 3]?.[0];",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const x: { a?: { b: string } };
x?.a?.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const x: { a: { b?: { c: string } } };
x.a?.b?.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
let x: { a?: string };
x?.a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const foo: { bar: { baz: { c: string } } } | null;
foo?.bar?.baz;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const foo: { bar?: { baz: { qux: string } } } | null;
foo?.bar?.baz?.qux;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const foo: { bar: { baz: { qux?: () => {} } } } | null;
foo?.bar?.baz?.qux?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
declare const foo: { bar: { baz: { qux: () => {} } } } | null;
foo?.bar?.baz?.qux?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
type baz = () => { qux: () => {} };
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
type baz = null | (() => { qux: () => {} });
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.().qux?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
type baz = null | (() => { qux: () => {} } | null);
declare const foo: { bar: { baz: baz } } | null;
foo?.bar?.baz?.()?.qux?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
type Foo = { baz: number };
type Bar = { baz: null | string | { qux: string } };
declare const foo: { fooOrBar: Foo | Bar } | null;
foo?.fooOrBar?.baz?.qux;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const x: { a: { b: number } }[];
x[0].a?.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
type Key = 'bar' | 'foo';
declare const foo: Foo;
declare const key: Key;

foo?.[key]?.trim();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
type Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;
declare const foo: Foo;
const key = 'bar';
foo?.[key]?.trim();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}

export function test(outer: Outer): number | undefined {
  const key = 'bar';
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
interface Outer {
  inner?: {
    [key: string]: string | undefined;
    bar: 'bar';
  };
}
type Bar = 'bar';

function Foo(outer: Outer, key: Bar): number | undefined {
  return outer.inner?.[key]?.charCodeAt(0);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},

		// Nullish coalescing with testVal
		{
			Code: `
function test(testVal?: true) {
  if (testVal ?? true) {
    console.log('test');
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Negation operators
		{
			Code: `
const a = null;
if (!a) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
const a = true;
if (!a) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
function sayHi(): void {
  console.log('Hi!');
}

let speech: never = sayHi();
if (!speech) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}},
		},

		// No strict null checks
		{
			Code: `
declare const x: string[] | null;
if (x) {
}
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck"},
				{MessageId: "alwaysTruthy"},
			},
		},

		// Index signature cases
		{
			Code: `
interface Foo {
  test: string;
  [key: string]: [string] | undefined;
}

type OptionalFoo = Foo | undefined;
declare const foo: OptionalFoo;
foo?.test?.length;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
function pick<Obj extends Record<string, 1 | 2 | 3>, Key extends keyof Obj>(
  obj: Obj,
  key: Key,
): Obj[Key] {
  const k = obj[key];
  if (obj[key]) {
    return obj[key];
  }
  throw new Error('Boom!');
}

pick({ foo: 1, bar: 2 }, 'bar');
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
function getElem(dict: Record<string, { foo: string }>, key: string) {
  if (dict[key]) {
    return dict[key].foo;
  } else {
    return '';
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Logical assignment operators - errors
		{
			Code: `
declare let foo: {};
foo ??= 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
declare let foo: number;
foo ??= 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
declare let foo: null;
foo ??= null;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish"}},
		},
		{
			Code: `
declare let foo: {};
foo ||= 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare let foo: null;
foo ||= null;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare let foo: {};
foo &&= 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare let foo: null;
foo &&= null;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const foo: { bar: number };
foo.bar ??= 1;
      `,
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},

		// Function return values with optional chaining
		{
			Code: `
type Foo = { bar: () => number } | null;
declare const foo: Foo;
foo?.bar()?.toExponential();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
type Foo = { bar: null | { baz: () => { qux: number } } } | null;
declare const foo: Foo;
foo?.bar?.baz()?.qux?.toExponential();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
type Foo = (() => number) | null;
declare const foo: Foo;
foo?.()?.toExponential();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']()?.toExponential();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
type Foo = { [key: string]: () => number } | null;
declare const foo: Foo;
foo?.['bar']?.()?.toExponential();
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
        const a = true;
        if (!!a) {
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Type predicates with checkTypePredicates option
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert(true);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert(false);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare function assert(x: unknown, y: unknown): asserts x;

assert(true, Math.random() > 0.5);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare function assert(x: unknown): asserts x;
assert({});
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare function assertsString(x: unknown): asserts x is string;
declare const a: string;
assertsString(a);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType"}},
		},
		{
			Code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString(a);
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType"}},
		},
		{
			Code: `
declare function isString(x: unknown): x is string;
declare const a: string;
isString('fa' + 'lafel');
      `,
			Options: NoUnnecessaryConditionOptions{CheckTypePredicates: utils.Ref(true)},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType"}},
		},

		// "Branded" types - always falsy/truthy
		{
			Code: `
declare const b1: "" & {};
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: "" & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: ("" | false) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & "";
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}},
		},
		{
			Code: `
declare const b1: ("foo" | "bar") & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: (123 | true) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: (string | number) & ("foo" | 123) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & "foo";
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
		{
			Code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & ("foo" | 123);
declare const b2: boolean;
const t1 = b1 && b2;
`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},

		// Mapped types with optional chains
		{
			Code: `
type A = {
  [name in Lowercase<string>]?: {
    [name in Lowercase<string>]: {
      a: 1;
    };
  };
};

declare const a: A;

a.a?.a?.a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
interface T {
  [name: Lowercase<string>]: {
    [name: Lowercase<string>]: {
      [name: Lowercase<string>]: {
        value: 'value';
      };
    };
  };
  [name: Uppercase<string>]: null | {
    [name: Uppercase<string>]: null | {
      [name: Uppercase<string>]: null | {
        VALUE: 'VALUE';
      };
    };
  };
}

declare const t: T;

t.a?.a?.a?.value;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},

		// noUncheckedIndexedAccess cases
		{
			Code: `
declare const test: Array<{ a?: string }>;

if (test[0]?.a) {
  test[0]?.a;
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain"}},
		},
		{
			Code: `
declare const arr2: Array<{ x: { y: { z: object } } }>;
arr2[42]?.x?.y?.z;
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain"},
				{MessageId: "neverOptionalChain"},
			},
		},
		{
			Code: `
declare const arr: string[];

if (arr[0]) {
  arr[0] ?? 'foo';
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
		},
		{
			Code: `
declare const arr: object[];

if (arr[42] && arr[42]) {
}
      `,
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}},
		},
	})
}
