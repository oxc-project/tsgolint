package no_unsafe_enum_assignment

import (
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"testing"
)

func TestNoUnsafeEnumAssignmentRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeEnumAssignmentRule, []rule_tester.ValidTestCase{
		{Code: `
enum Fruit {
  Apple,
}

const fruit: Fruit = Fruit.Apple;
    `},
		{Code: `
enum Fruit {
  Apple,
}

const getFruit = (): Fruit => Fruit.Apple;
    `},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  fruit: Fruit = Fruit.Apple;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

const box: { fruit: Fruit } = { fruit: Fruit.Apple };
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare function Basket(props: { fruit: Fruit }): JSX.Element;

<Basket fruit={Fruit.Apple} />;
      `, Tsx: true},
		{Code: `
({ fruit: 1 });
    `},
		{Code: `
enum Fruit {
  Apple,
}

const fruits: [Fruit] = [Fruit.Apple];
void fruits;
    `},
		{Code: `
enum Fruit {
  Apple,
}

const [fruit]: [Fruit] = [Fruit.Apple] as Fruit[];
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit;
[fruit] = [Fruit.Apple];
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takeFruit([fruit]: Fruit[] = [Fruit.Apple]): void {
  void fruit;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

const getFruit = (): Fruit => {
  return Fruit.Apple;
};

void getFruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruits(...fruits: Fruit[]): void {}

const fruits: Fruit[] = [Fruit.Apple];
takesFruits(...fruits);
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { plain: string };

foo[0];
    `},
		{Code: `
enum Fruit {
  Apple,
}

({})[0];
    `},
		{Code: `
enum Fruit {
  Apple,
}

const [fruit]: [Fruit] = [Fruit.Apple];
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

const { fruit }: { fruit: Fruit } = {} as {};
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruitKey: unique symbol;

type Basket = { [fruitKey]: Fruit };

const { [fruitKey]: fruit }: Basket = { [fruitKey]: Fruit.Apple };
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruitKey: unique symbol;

interface Basket {
  [fruitKey]: Fruit;
}

class TestBasket implements Basket {
  [fruitKey] = Fruit.Apple;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

const fruit = Fruit.Apple as Fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  fruit = Fruit.Apple;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class TestBasket implements Basket {
  fruit = Fruit.Apple;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { [key in Fruit]: string };

foo[Fruit.Apple];
foo?.[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { [key in Fruit | number]: string };

foo[0];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const something: any;

const fruit: Fruit = something;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const something: any;

function takesFruit(fruit: Fruit): void {}

takesFruit(something);
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const something: unknown;

const fruit: Fruit = something as Fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruitOrNull(fruit: Fruit | null): void {}

takesFruitOrNull(null);
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

declare const apple: Fruit;

const fruit: Fruit = apple;
    `},
		{Code: `
const enum Direction {
  Up,
  Down,
}

const dir: Direction = Direction.Up;
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = Flags.Read | Flags.Write;
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

let flags: Flags = Flags.Read;
flags |= Flags.Write;
flags &= Flags.Read;
flags ^= Flags.Write;
    `},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit | undefined;
fruit ??= Fruit.Apple;
fruit ||= Fruit.Apple;
fruit &&= Fruit.Apple;
    `},
		{Code: `
enum Fruit {
  Apple,
}

let count = 1;
count++;
count += 1;
void Fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruits: Record<Fruit, string>;

const hasApple = 0 in fruits;
    `},
		{Code: `
enum Fruit {
  Apple,
}

type Mapped<T> = { [K in keyof T]: T[K] };

declare const source: Mapped<{ fruit: Fruit }>;

const target: { fruit: Fruit } = source;
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Api<T> {
  extend<U>(value: U): Api<T & U>;
  fruit: T;
}

declare const api: Api<Fruit>;

const extended: Api<Fruit> = api;
    `},
		{Code: `
enum Fruit {
  Apple,
}

enum Vegetable {
  Asparagus = 'asparagus',
}

function takesEither(val: Fruit | Vegetable): void {}

takesEither(Fruit.Apple);
takesEither(Vegetable.Asparagus);
    `},
		{Code: `
enum Fruit {
  Apple,
}

async function getFruit(): Promise<Fruit> {
  return Fruit.Apple;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

function getFruitOrNull(): Fruit | null {
  return null;
}
    `},
		{Code: "\nenum Fruit {\n  Apple,\n}\n\ndeclare function tag(strings: TemplateStringsArray, fruit: Fruit): string;\n\ntag`${Fruit.Apple}`;\n    "},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  constructor(public fruit: Fruit) {}
}

new Basket(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

const fruit = Fruit.Apple;
const box: { fruit: Fruit } = { fruit };
    `},
		{Code: `
enum Mixed {
  A = 'a',
  B = 1,
}

const val: Mixed = Mixed.A;
const val2: Mixed = Mixed.B;
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface HasFruit {
  fruit: Fruit;
}

interface HasNumber {
  fruit: number;
}

class TestBasket implements HasFruit, HasNumber {
  fruit = 1;
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(fruit: Fruit): void {}

takesFruit(...[Fruit.Apple]);
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

const fruits: Fruit[] = [Fruit.Apple, Fruit.Banana];
    `},
		{Code: `
enum Fruit {
  Apple,
}

const fruits: Array<Fruit | number> = [1, Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const bar: { [Fruit.Apple]: string };

bar[Fruit.Apple];
bar?.[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesNothing(): void {}

takesNothing();
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesRest(first: Fruit, ...rest: Fruit[]): void {}

takesRest(Fruit.Apple, Fruit.Apple, Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruits(first: Fruit, second: Fruit): void {}

takesFruits(...[Fruit.Apple, ...[Fruit.Apple]]);
    `},
		{Code: `
enum Fruit {
  Apple,
}

const { ['fruit']: fruit }: { fruit: Fruit } = { fruit: Fruit.Apple };
void fruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

const [, ...rest]: Fruit[] = [Fruit.Apple, Fruit.Apple];
void rest;
    `},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit;
({ fruit } = { fruit: Fruit.Apple });
    `},
		{Code: `
enum Fruit {
  Apple,
}

function pick({ fruit }: { fruit: Fruit } = { fruit: Fruit.Apple }): Fruit {
  return fruit;
}

pick();
    `},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

declare function Basket(props: { vegetable: Vegetable }): JSX.Element;

<Basket vegetable="asparagus" />;
      `, Tsx: true},
		{Code: `
enum Fruit {
  Apple,
}

function noop(): void {
  return;
}

noop();
    `},
		{Code: `
enum Fruit {
  Apple,
}

const fruits: Fruit[] = [...[Fruit.Apple]];
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takeFruit(fruit: Fruit): void {}
takeFruit(...[Fruit.Apple, ...[1]]);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesGeneric<T extends readonly unknown[]>(...args: T): void {}

takesGeneric(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class Base {
  fruit: Fruit = Fruit.Apple;
}

class Child extends Base implements Basket {
  fruit = Fruit.Apple;
}

void Child;
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class Child implements Basket {
  other = Fruit.Apple;
  fruit: Fruit = Fruit.Apple;
}

void Child;
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = Flags.Read | (Flags.Write & Flags.Read);
void combined;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(fruit: Fruit): void {}
takesFruit(Fruit.Apple, Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const rec: Record<string, string>;
rec[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

doesNotExist(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruits(fruits: Fruit[]): void {}

takesFruits([Fruit.Apple]);
    `},
		{Code: `
enum Fruit {
  Apple,
}

[1];
    `},
		{Code: `
const numbers: number[] = [1];
void numbers;
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class TestBasket implements Basket {
  #other = 1;
  fruit = Fruit.Apple;
}

void TestBasket;
    `},
		{Code: `
function foo(): void {}

foo[0];
    `},
		{Code: `
const foo = {};

foo[0];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare function Basket(props: { fruit?: Fruit }): JSX.Element;

<Basket fruit={/* empty */} />;
      `, Tsx: true},
		{Code: `
enum Fruit {
  Apple,
}

function identity<T extends Fruit>(value: T): T {
  const next: T = value;
  return next;
}

identity(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function identity<T extends Fruit, V extends T>(value: V): V {
  const next: V = value;
  return next;
}

identity(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesBox<T extends { fruit: Fruit }>(box: T): T {
  const next: T = box;
  return next;
}

takesBox({ fruit: Fruit.Apple });
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesNestedBox<T extends { box: { fruit: Fruit } }>(box: T): T {
  return box;
}

takesNestedBox({ box: { fruit: Fruit.Apple } });
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesArray<T extends Fruit[]>(value: T): T {
  const next: T = value;
  return next;
}

takesArray([Fruit.Apple]);
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

function takesTuple<T extends [Fruit, Fruit]>(...value: T): T {
  return value;
}

takesTuple(Fruit.Apple, Fruit.Banana);
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

function takesReadonlyTuple<T extends readonly [Fruit, Fruit]>(value: T): T {
  return value;
}

takesReadonlyTuple([Fruit.Apple, Fruit.Banana] as const);
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

function takesRestTuple<T extends [Fruit, ...Fruit[]]>(...value: T): T {
  return value;
}

takesRestTuple(Fruit.Apple, Fruit.Banana);
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Box<T> {
  fruit: T;
}

function takesInterfaceBox<T extends Fruit>(box: Box<T>): Box<T> {
  return box;
}

takesInterfaceBox({ fruit: Fruit.Apple });
    `},
		{Code: `
enum Fruit {
  Apple,
}

class Box<T> {
  constructor(public fruit: T) {}
}

const box: Box<Fruit> = new Box(Fruit.Apple);
void box;
    `},
		{Code: `
enum Fruit {
  Apple,
}

async function getFruit<T extends Fruit>(value: T): Promise<T> {
  return value;
}

void getFruit(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket<T> {
  fruit: T;
}

class BasketImpl<T extends Fruit> implements Basket<T> {
  fruit: T;

  constructor(fruit: T) {
    this.fruit = fruit;
  }
}

void new BasketImpl(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

class Base<T> {
  fruit!: T;
}

class BasketImpl<T extends Fruit> extends Base<T> {
  fruit: T;

  constructor(fruit: T) {
    super();
    this.fruit = fruit;
  }
}

void new BasketImpl(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { [key in Fruit]: string };

function readValue<K extends Fruit>(key: K): string {
  return foo[key];
}

readValue(Fruit.Apple);
    `},
		{Code: `
enum Fruit {
  Apple,
}

function destructureArray<T extends [Fruit]>(value: T): T {
  const [fruit] = value;
  void fruit;
  return value;
}

destructureArray([Fruit.Apple]);
    `},
		{Code: "\nenum Fruit {\n  Apple,\n}\n\nconst notFn = 1;\nnotFn(Fruit.Apple);\n\nconst notTag = 1;\nnotTag`fruit`;\n    "},
		{Code: `
enum Fruit {
  Apple,
}

const box: { [key in Fruit]: number } = { [Fruit.Apple]: 1 };
void box;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare function takeBox(box: { fruit: Fruit }): void;

takeBox([Fruit.Apple]);

const value: {} = [Fruit.Apple];
void value;
    `},
		{Code: `
enum Fruit {
  Apple,
}

type BadThenable = {
  then(onfulfilled: string): unknown;
};

async function getFruit(): Promise<Fruit> {
  return {} as BadThenable;
}

void getFruit;
    `},
		{Code: `
enum Fruit {
  Apple,
}

type BadThenable = {
  then(onfulfilled: string): unknown;
};

async function getFruit(): BadThenable {
  return Promise.resolve(Fruit.Apple);
}

void getFruit;
    `},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

let foo: { [key in Fruit]: string } = {
  [Fruit.Apple]: 'apple',
  [Fruit.Banana]: 'banana',
};

foo[Math.random() > 0.5 ? Fruit.Apple : Fruit.Banana] = 'fruit';

const value = foo[Math.random() > 0.5 ? Fruit.Apple : Fruit.Banana];
void value;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function destructureObject<T extends { fruit: Fruit }>(value: T): T {
  const { fruit } = value;
  void fruit;
  return value;
}

destructureObject({ fruit: Fruit.Apple });
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare function Basket<T extends { fruit: Fruit }>(props: T): JSX.Element;

<Basket fruit={Fruit.Apple} />;
      `, Tsx: true},
		{Code: `
enum Fruit {
  Apple,
}

declare namespace JSX {
  interface IntrinsicElements {
    div: {};
  }
}

<div fruit={Fruit.Apple} />;
      `, Tsx: true},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

function getFlags(): Flags {
  return Flags.Read | Flags.Write;
}
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined = (Flags.Read | Flags.Write) as Flags;
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

declare function Basket(props: { flags: Flags }): JSX.Element;

<Basket flags={Flags.Read | Flags.Write} />;
      `, Tsx: true},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

function takesFlags(flags: Flags): void {}

takesFlags(Flags.Read | Flags.Write);
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

interface HasFlags {
  flags: Flags;
}

class Permissions implements HasFlags {
  flags = Flags.Read | Flags.Write;
}
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

let flags: Flags = Flags.Read;
flags |= Flags.Read | Flags.Write;
flags &= ~Flags.Write;
flags ^= Flags.Read ^ Flags.Write;
    `},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = Flags.Read ^ Flags.Write;
const masked: Flags = Flags.Read & ~Flags.Write;
    `},
		{Code: `
enum Fruit {
  Apple,
}

return 1;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruits: Fruit[];
declare const fruitSet: Set<Fruit>;

const more: Fruit[] = [...fruits, ...fruitSet, , Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const source: { fruit: Fruit };

const box: { fruit: Fruit } = { ...source };
({ ...source });
    `},
		{Code: `
enum Fruit {
  Apple,
}

const box: { getFruit(): Fruit } = {
  getFruit() {
    return Fruit.Apple;
  },
};
    `},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  fruits: { [key in Fruit]: string } = { [Fruit.Apple]: 'apple' };

  getApple(): string {
    return this.fruits[Fruit.Apple];
  }
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const basket: { fruits: { [key in Fruit]: string } };

basket.fruits[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const bar: { [Fruit.Apple]: string; other: string };

bar[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruits: Fruit[];
declare const fruitSet: Set<Fruit>;
const notFn = 1;

function takesFruits(...fruits: Fruit[]): void {}

takesFruits(...fruitSet);
notFn(...fruits);
    `},
		{Code: `
interface FruitNode {
  next: FruitNode;
  value: string;
}

interface OtherNode {
  next: OtherNode;
  value: string;
}

declare const other: OtherNode;

const node: FruitNode = other;
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takesStrings(...strings: string[]): void {}

takesStrings(...'abc');
    `},
		{Code: `
const label: string = 'label';
    `},
		{Code: `
enum Fruit {
  Apple,
}

enum Vegetable {
  Asparagus = 'asparagus',
}

declare const bar: {
  [Vegetable.Asparagus]: string;
  [Fruit.Apple]: string;
  other: string;
};

bar[Vegetable.Asparagus];
bar[Fruit.Apple];
    `},
		{Code: `
enum Fruit {
  Apple,
}

function takes(first: Fruit, second: number, ...rest: Fruit[]): void {}

takes(...([Fruit.Apple] as const), 1);
    `},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

declare const something: any;

const vegetable: Vegetable = something;
    `},
		{Code: `
enum Fruit {
  Apple,
}

declare function getFruits(): { [key in Fruit]: string };

getFruits[0];
    `},
		{Code: `
function first() {
  return arguments[0];
}
    `},
		{Code: `
type Unwrap<T> = T extends (infer U)[] ? U : T;

interface Schema<T> {
  in: Schema<Unwrap<T>>;
  out: Schema<Unwrap<T>[]>;
}

function widen<T>(schema: Schema<T>): Schema<T | undefined> {
  return schema;
}
    `},
		{Code: `
type Unwrap<T> = T extends (infer U)[] ? U : T;

interface Schema<T> {
  in: Schema<Unwrap<T>>;
  out: Schema<Unwrap<T>[]>;
}

declare function register(schema: Schema<unknown>): void;

function forward<T>(schema: Schema<T>) {
  register(schema);
}
    `},
		{Code: `
enum Fruit {
  Apple,
}

type Unwrap<T> = T extends (infer U)[] ? U : T;

interface Schema<T> {
  in: Schema<Unwrap<T>>;
  out: Schema<Unwrap<T>[]>;
}

declare function register(schema: Schema<{ fruit: Fruit }>): void;

function forward<T extends { fruit: Fruit }>(schema: Schema<T>) {
  register(schema);
}
    `},
		{Code: `
type Wrapper<T> = { next: Wrapper<[T]>; value: T };

declare const wrapper: Wrapper<0>;

const widened: Wrapper<0 | 1> = wrapper;
    `},
		{Code: `
enum Fruit { Apple }
const fruit = 'count';
interface Basket { fruit: Fruit; count: number }
class TestBasket implements Basket {
  fruit = Fruit.Apple;
  [fruit] = 1;
}
`},
		{Code: `
enum Fruit { Apple }
interface Basket { fruit: Fruit }
class TestBasket implements Basket {
  static fruit = 1;
  fruit = Fruit.Apple;
}
`},
		{Code: `
enum Fruit { Apple }
class Base { fruit: Fruit = Fruit.Apple }
class Child extends Base { static fruit = 1 }
`},
		{Code: `
enum Fruit { Apple }
interface Basket { fruit: Fruit }
class Base { static fruit = 0 }
class Child extends Base implements Basket {
  static fruit = 1;
  fruit = Fruit.Apple;
}
`},
		{Code: `
enum Fruit { Apple }
class Base { static fruit: Fruit = Fruit.Apple }
class Child extends Base { fruit = 1 }
`},
		{Code: `
enum Fruit { Apple }
const fruit = Fruit.Apple satisfies Fruit;
const value = 0 satisfies number;
`},
		{Code: `
enum Flags { Read = 1, Write = 2 }
const flags = (Flags.Read | Flags.Write) satisfies Flags;
`},
		{Code: `
enum Fruit { Apple }
type FruitMap = { [K in Fruit]: string };
declare const map: FruitMap;
map[Fruit.Apple];
map?.[Fruit.Apple];
`},
		{Code: `
enum Fruit { Apple }
type StringMap = { [K in string]: string };
declare const map: StringMap;
map[0];
`},
		{Code: `
enum Fruit { Apple }
type EnumMap<K extends PropertyKey> = { [P in K]: string };
declare const map: EnumMap<Fruit>;
map[Fruit.Apple];
`},
		{Code: `
enum Fruit { Apple }
type FruitMap = { [K in Fruit as 'key']: string };
declare const map: FruitMap;
map['key'];
`},
		{Code: `
declare const stringMap: Record<string, string>;
declare const numberMap: Record<number, string>;
stringMap[0];
numberMap[0];
`},
	}, []rule_tester.InvalidTestCase{
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit = Fruit.Apple;
fruit++;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 8, EndLine: 7, Line: 7, MessageId: "unsafeEnumMutation"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit = Fruit.Apple;
fruit += 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 11, EndLine: 7, Line: 7, MessageId: "unsafeEnumMutation"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const basket: { fruit: Fruit };
basket.fruit++;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 15, EndLine: 7, Line: 7, MessageId: "unsafeEnumMutation"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

declare const someNumber: number;

let flags: Flags = Flags.Read;
flags |= someNumber;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 20, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit | undefined;
fruit ??= 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare function takes(...args: [Fruit, Fruit]): void;
declare const spread: [Fruit, ...Fruit[]];
takes(...spread, 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 18, EndColumn: 19, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare function takes(...args: [Fruit, ...Fruit[]]): void;

takes(Fruit.Apple, Fruit.Apple, 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 33, EndColumn: 34, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruit: Fruit = 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 23, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit = Fruit.Apple;
fruit = 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 10, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(fruit: Fruit = 1) {}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 21, EndColumn: 37, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  fruit: Fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 20, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class TestBasket implements Basket {
  fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 13, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const box: { fruit: Fruit } = { fruit: 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 33, EndColumn: 41, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(fruit: Fruit): void {}

takesFruit(1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 13, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const maybeFruit: Fruit | number;

const fruit: Fruit = maybeFruit;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 32, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function getFruit(): Fruit {
  return 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

const getVegetable = (): Vegetable => 'asparagus';
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 39, EndColumn: 50, EndLine: 6, Line: 6, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruit = 1 as Fruit;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 25, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssertion"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruit = <Fruit>1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 23, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssertion"}}},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

const vegetable: Vegetable = 'asparagus';
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 41, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(...fruit: [Fruit]): void {}

takesFruit(...([1] as const));
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 29, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit(fruit: Fruit): void {}

takesFruit(...[1]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 18, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

function takesFruits(first: Fruit, second: Fruit): void {}

takesFruits(...[Fruit.Apple, 1]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 32, EndLine: 9, Line: 9, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

const fruits: Fruit[] = [1, 2];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 27, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}, {Column: 29, EndColumn: 30, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruits(fruits: Fruit[]): void {}

takesFruits([1]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 14, EndColumn: 15, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const basket: { fruit: Fruit } = { fruit: Fruit.Apple };

basket.fruit = 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 17, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

declare function Basket(props: { vegetable: Vegetable }): JSX.Element;

<Basket vegetable={'asparagus'} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{Column: 20, EndColumn: 31, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { [key in Fruit]: string };

foo[0];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 5, EndColumn: 6, EndLine: 8, Line: 8, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Vegetable {
  Asparagus = 'asparagus',
}

declare const foo: { [key in Vegetable]: string };

foo?.['asparagus'];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 18, EndLine: 8, Line: 8, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const bar: { [Fruit.Apple]: string };

bar[0];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 5, EndColumn: 6, EndLine: 8, Line: 8, MessageId: "unsafeEnumAccess"}}},
		{Code: `
const enum Direction {
  Up,
  Down,
}

const dir: Direction = 0;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 25, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

enum Vegetable {
  Asparagus,
}

const fruit: Fruit = Vegetable.Asparagus;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 41, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

async function getFruit(): Promise<Fruit> {
  return 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumReturn"}}},
		{Code: "\nenum Fruit {\n  Apple,\n}\n\ndeclare function tag(strings: TemplateStringsArray, fruit: Fruit): string;\n\ntag`${1}`;\n      ", Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 8, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  constructor(public fruit: Fruit) {}
}

new Basket(1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 13, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const num: number;

const fruit: Fruit = num;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 25, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  constructor(private fruit: Fruit = 1) {}
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 23, EndColumn: 39, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const fruit: number;

const box: { fruit: Fruit } = { fruit };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 33, EndColumn: 38, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const [fruit]: [Fruit] = [1];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 27, EndColumn: 28, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const { fruit = 1 }: { fruit?: Fruit } = {};
void fruit;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 18, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit([fruit]: [Fruit] = [1]): void {}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 41, EndColumn: 42, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const { fruit }: { fruit: Fruit } = { fruit: 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 39, EndColumn: 47, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesBox({ fruit }: { fruit: Fruit } = { fruit: 1 }): void {}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 51, EndColumn: 59, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruits: Set<Fruit> = new Set([1, 2]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 43, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const cond: boolean;

const fruit: Fruit = cond ? 1 : Fruit.Apple;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 44, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  accessor fruit: Fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 29, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruit<T extends Fruit>(fruit: T): void {}

takesFruit(1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 13, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

const x = [1, Fruit.Apple, 2] as const;
const y: readonly Fruit[] = x;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 30, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = 1 | 2;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 30, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesRest(first: Fruit, ...rest: Fruit[]): void {}

takesRest(Fruit.Apple, 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 24, EndColumn: 25, EndLine: 8, Line: 8, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = Flags.Read + Flags.Write;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 49, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesFruits(...fruits: Fruit[]): void {}
const numbers: number[] = [1, 2, 3];

takesFruits(...[1, ...numbers]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 31, EndLine: 9, Line: 9, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket {
  fruit: Fruit;
}

class TestBasket implements Basket {
  accessor fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 22, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignFruit<T extends Fruit>(): void {
  const fruit: T = 1;
  void fruit;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 21, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function getFruit<T extends Fruit>(): T {
  return 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

function castFruit<T extends Fruit>(): T {
  return 1 as T;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 16, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssertion"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Box<T extends Fruit> {
  constructor(public fruit: T) {}
}

new Box(1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 10, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare function Basket<T extends Fruit>(props: { fruit: T }): JSX.Element;

<Basket fruit={1} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{Column: 16, EndColumn: 17, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignFruit<T extends Fruit, V extends T>(): V {
  return 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesBox<T extends { fruit: Fruit }>(box: T): void {}

takesBox({ fruit: 1 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 20, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignBox<T extends { fruit: Fruit }>(): void {
  const box: T = { fruit: 1 };
  void box;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 20, EndColumn: 28, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function getBox<T extends { fruit: Fruit }>(): T {
  return { fruit: 1 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 20, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function castBox<T extends { fruit: Fruit }>(): T {
  return { fruit: 1 } as T;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 20, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Box<T extends { fruit: Fruit }> {
  constructor(public value: T) {}
}

new Box({ fruit: 1 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 19, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare function Basket<T extends { fruit: Fruit }>(props: T): JSX.Element;

<Basket fruit={1} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{Column: 16, EndColumn: 17, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesNestedBox<T extends { box: { fruit: Fruit } }>(box: T): void {}

takesNestedBox({ box: { fruit: 1 } });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 25, EndColumn: 33, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesArray<T extends Fruit[]>(value: T): void {}

takesArray([1]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 14, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignArray<T extends Fruit[]>(): void {
  const fruits: T = [1];
  void fruits;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 23, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function getArray<T extends Fruit[]>(): T {
  return [1];
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function castArray<T extends Fruit[]>(): T {
  return [1] as T;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesBoxes<T extends Array<{ fruit: Fruit }>>(boxes: T): void {}

takesBoxes([{ fruit: 1 }]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 23, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignBoxes<T extends readonly { fruit: Fruit }[]>(): void {
  const boxes: T = [{ fruit: 1 }] as const;
  void boxes;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 43, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
  Banana,
}

function assignReadonlyTuple<T extends readonly [Fruit, Fruit]>(): void {
  const value: T = [Fruit.Apple, 1] as const;
  void value;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 45, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

type Box<T> = { fruit: T };

function takesAliasBox<T extends Fruit>(box: Box<T>): void {}

takesAliasBox({ fruit: 1 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 17, EndColumn: 25, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface Box<T> {
  fruit: T;
}

const box: Box<Fruit> = { fruit: 1 };
void box;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 27, EndColumn: 35, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Box<T> {
  constructor(public fruit: T) {}
}

const box: Box<Fruit> = new Box(1);
void box;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 35, EndLine: 10, Line: 10, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

async function getBox<T extends { fruit: Fruit }>(): Promise<T> {
  return { fruit: 1 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 20, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket<T> {
  fruit: T;
}

class BasketImpl implements Basket<Fruit> {
  fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 13, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Base<T> {
  fruit!: T;
}

class BasketImpl extends Base<Fruit> {
  fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 13, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface Basket<T> {
  fruit: T;
}

class BasketImpl<T extends Fruit> implements Basket<T> {
  fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 13, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const foo: { [key in Fruit]: string };

function readValue<K extends number>(key: K): string {
  return foo[key];
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 14, EndColumn: 17, EndLine: 9, Line: 9, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

function readValue<K extends number>(
  foo: { [P in Fruit]: string },
  key: K,
): string {
  return foo[key];
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 14, EndColumn: 17, EndLine: 10, Line: 10, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const combined: Flags = Flags.Read | (Flags.Write | 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 55, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const negated: Flags = -Flags.Read;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 35, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

const shifted: Flags = Flags.Read << Flags.Write;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 49, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

let flags: Flags = Flags.Read;
flags |= Flags.Write | 4;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 25, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

let flags: Flags = Flags.Read;
flags &= 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 11, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Flags {
  Read = 1 << 0,
  Write = 1 << 1,
}

let flags: Flags = Flags.Read;
flags ^= 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 11, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit | undefined;
fruit ||= 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit | undefined;
fruit &&= 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 1, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numbers: number[];

const fruits: Fruit[] = [Fruit.Apple, ...numbers];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 39, EndColumn: 49, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const pair: [Fruit, number] = [1, 2];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 32, EndColumn: 33, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruitOrFruits: Fruit | Fruit[] = [1];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 41, EndColumn: 42, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const nested: Fruit[][] = [[1]];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 29, EndColumn: 30, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const box: { 0: Fruit; 'the-fruit': Fruit } = { 0: 1, 'the-fruit': 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 49, EndColumn: 53, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}, {Column: 55, EndColumn: 69, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const box: { fruits: Fruit[] } = { fruits: [1] };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 45, EndColumn: 46, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit;
[fruit] = [1];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 13, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

let fruit: Fruit;
({ fruit } = { fruit: 1 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 16, EndColumn: 24, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const pair: [number];

const [fruit]: [Fruit] = pair;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 30, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const source: { fruit: number };

const { fruit }: { fruit: Fruit } = source;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 43, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const source: { fruit: number };

const box: { fruit: Fruit } = { ...source };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 33, EndColumn: 42, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numbers: number[];

function takesFruits(...fruits: Fruit[]): void {}

takesFruits(...numbers);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 23, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numbers: number[];

function takesFruits(...fruits: Fruit[]): void {}

takesFruits(Fruit.Apple, ...numbers, 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 36, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}, {Column: 38, EndColumn: 39, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

async function getFruit(): Promise<Fruit> {
  return Promise.resolve(1);
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 29, EndLine: 7, Line: 7, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

function takesBoxes<T extends { fruit: Fruit }>(boxes: T[]): void {}

takesBoxes([{ fruit: 1 }]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 23, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  fruits: { [key in Fruit]: string } = { [Fruit.Apple]: 'apple' };

  getFirst(): string {
    return this.fruits[0];
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 24, EndColumn: 25, EndLine: 10, Line: 10, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const basket: { fruits: { [key in Fruit]: string } };

basket.fruits[0];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 16, EndLine: 8, Line: 8, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface HasFruit {
  fruit: Fruit;
}

interface AlsoHasFruit {
  fruit: Fruit;
}

class Basket implements HasFruit, AlsoHasFruit {
  fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 13, EndLine: 15, Line: 15, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface HasNumber {
  fruit: number;
}

class Basket implements HasNumber {
  fruit: Fruit = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 20, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function getFruits(): Fruit[] {
  return [1];
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 12, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const getFruits = (): Fruit[] => [1];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 35, EndColumn: 36, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const box = { fruit: 1 } satisfies { fruit: Fruit };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 23, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

const fruits: Map<string, Fruit> = new Map<string, number>();
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 61, EndLine: 6, Line: 6, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface FruitNode {
  fruit: Fruit;
  next: FruitNode;
}

interface NumberNode {
  fruit: number;
  next: NumberNode;
}

declare const numberNode: NumberNode;

const fruitNode: FruitNode = numberNode;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 40, EndLine: 18, Line: 18, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function increment<T extends Fruit>(fruit: T): void {
  fruit++;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 10, EndLine: 7, Line: 7, MessageId: "unsafeEnumMutation"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numberSet: Set<number>;

const fruits: Fruit[] = [...numberSet];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 38, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numberSet: Set<number>;

function takesFruits(...fruits: Fruit[]): void {}

takesFruits(...numberSet);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 25, EndLine: 10, Line: 10, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numbers: number[];

const fruits: { [index: number]: Fruit } = numbers;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 51, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const pair: [number];

const fruitPair: { [index: number]: Fruit } = pair;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 51, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

enum Vegetable {
  Asparagus = 'asparagus',
}

declare const bar: {
  [Vegetable.Asparagus]: string;
  [Fruit.Apple]: string;
  other: string;
};

bar[0];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 5, EndColumn: 6, EndLine: 16, Line: 16, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

enum Vegetable {
  Asparagus = 'asparagus',
}

declare const bar: {
  [Vegetable.Asparagus]: string;
  [Fruit.Apple]: string;
  other: string;
};

bar['asparagus'];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 5, EndColumn: 16, EndLine: 16, Line: 16, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const spread: [Fruit, ...Fruit[]];

function takes(
  first: Fruit,
  second: number,
  third: number,
  ...rest: Fruit[]
): void {}

takes(...spread, 1);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 18, EndColumn: 19, EndLine: 15, Line: 15, MessageId: "unsafeEnumArgument"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface FruitThenable {
  fruit: Fruit;
  then(onfulfilled: string): unknown;
}

interface NumberThenable {
  fruit: number;
  then(onfulfilled: string): unknown;
}

declare const numberThenable: NumberThenable;

function getThenable(): FruitThenable {
  return numberThenable;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 25, EndLine: 19, Line: 19, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

class Basket {
  get fruits(): { [key in Fruit]: string } {
    return { [Fruit.Apple]: 'apple' };
  }

  getFirst(): string {
    return this.fruits[0];
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 24, EndColumn: 25, EndLine: 12, Line: 12, MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignNestedBox<T extends { box: { fruit: Fruit } }>(): void {
  const value: T = { box: { fruit: 1 } } as const;
  void value;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 50, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

function assignSpreadBox<T extends { fruit: Fruit }>(): void {
  const value: T = { ...{ fruit: 1 } } as const;
  void value;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 48, EndLine: 7, Line: 7, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface HasBox {
  box: { fruit: Fruit };
}

class Basket implements HasBox {
  box: { fruit: Fruit } = { fruit: 1 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 40, EndLine: 11, Line: 11, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const cond: boolean;
declare const source: { fruit: Fruit };

const box: { fruit: Fruit } = cond ? { fruit: 1 } : source;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 59, EndLine: 9, Line: 9, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const source: { fruit: Fruit } | undefined;

const box: { fruit: Fruit } = source ?? { fruit: 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 53, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

interface NumberBox<T> {
  fruit: number;
  value: T;
}

interface FruitBox<T> {
  fruit: Fruit;
  value: T;
}

function getFruitBox<T>(box: NumberBox<T>): FruitBox<T> {
  return box;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 3, EndColumn: 14, EndLine: 17, Line: 17, MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const numbers: Promise<Map<string, Set<number>[]>>;

const fruits: Promise<Map<string, Set<Fruit>[]>> = numbers;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 59, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit {
  Apple,
}

declare const config: { a: { b: { c: { d: { fruit: number } } } } };

const fruitConfig: { a: { b: { c: { d: { fruit: Fruit } } } } } = config;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 7, EndColumn: 73, EndLine: 8, Line: 8, MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
const key = 'fruit';
interface Basket { fruit: Fruit }
class TestBasket implements Basket {
  [key] = 1;
}
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
interface Basket<T> {
  fruit: Fruit;
  next: Basket<T[]>;
}
declare const source: { fruit: number; next: never };
const target: Basket<number> = source;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
interface NumberBox { fruit: number }
interface FruitBox { fruit: Fruit }
interface Source {
  deep: { a: { b: { c: { d: NumberBox } } } };
  shallow: NumberBox;
}
interface Target {
  deep: { a: { b: { c: { d: FruitBox } } } };
  shallow: FruitBox;
}
declare const source: Source;
const target: Target = source;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
interface NumberBox { fruit: number }
interface FruitBox { fruit: Fruit }
interface Source {
  shallow: NumberBox;
  deep: { a: { b: { c: { d: NumberBox } } } };
}
interface Target {
  shallow: FruitBox;
  deep: { a: { b: { c: { d: FruitBox } } } };
}
declare const source: Source;
const target: Target = source;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
interface SourceBox<T> { fruit: number; value: T }
interface TargetBox<T> { fruit: Fruit; value: T }
interface SourceDeep<T> { box: SourceBox<T> }
interface TargetDeep<T> { box: TargetBox<T> }
function convert<T>(source: { deep: SourceDeep<T>; shallow: SourceBox<T> }) {
  const target: { deep: TargetDeep<T>; shallow: TargetBox<T> } = source;
}
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
class Base { static fruit: Fruit = Fruit.Apple }
class Child extends Base { static fruit = 1 }
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
const key = 'fruit';
class Base { static fruit: Fruit = Fruit.Apple }
class Child extends Base { static [key] = 1 }
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
class Basket { static fruit: Fruit = 1 }
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
const fruit = 0 satisfies Fruit;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment", Line: 3, Column: 15, EndLine: 3, EndColumn: 32}}},
		{Code: `
enum Vegetable { Asparagus = 'asparagus' }
const vegetable = 'asparagus' satisfies Vegetable;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
const fruit: Fruit = 0 satisfies Fruit;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAssignment"}}},
		{Code: `
enum Fruit { Apple }
function getFruit(): Fruit {
  return 0 satisfies Fruit;
}
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumReturn"}}},
		{Code: `
enum Fruit { Apple }
type FruitMap = { [K in Fruit]: string };
declare const map: FruitMap;
map[0];
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAccess", Line: 5, Column: 5, EndLine: 5, EndColumn: 6}}},
		{Code: `
enum Fruit { Apple }
type FruitMap = { [K in Fruit]: string };
type Alias = FruitMap;
declare const map: Alias;
map?.[0];
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit { Apple }
type EnumMap<K extends PropertyKey> = { [P in K]: string };
declare const map: EnumMap<Fruit>;
map[0];
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAccess"}}},
		{Code: `
enum Fruit { Apple }
declare const map: Record<Fruit, string>;
map[0];
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeEnumAccess"}}},
	})
}
