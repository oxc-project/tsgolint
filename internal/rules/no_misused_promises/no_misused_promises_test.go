package no_misused_promises

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoMisusedPromisesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoMisusedPromisesRule, []rule_tester.ValidTestCase{
		{Code: `
if (true) {
}
    `},
		{
			Code: `
if (Promise.resolve()) {
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: `
if (true) {
} else if (false) {
} else {
}
    `},
		{
			Code: `
if (Promise.resolve()) {
} else if (Promise.resolve()) {
} else {
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "for (;;) {}"},
		{Code: "for (let i; i < 10; i++) {}"},
		{
			Code:    "for (let i; Promise.resolve(); i++) {}",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "do {} while (true);"},
		{
			Code:    "do {} while (Promise.resolve());",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "while (true) {}"},
		{
			Code:    "while (Promise.resolve()) {}",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "true ? 123 : 456;"},
		{
			Code:    "Promise.resolve() ? 123 : 456;",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: `
if (!true) {
}
    `},
		{
			Code: `
if (!Promise.resolve()) {
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "(await Promise.resolve()) || false;"},
		{
			Code:    "Promise.resolve() || false;",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "(true && (await Promise.resolve())) || false;"},
		{
			Code:    "(true && Promise.resolve()) || false;",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": false}`),
		},
		{Code: "false || (true && Promise.resolve());"},
		{Code: "(true && Promise.resolve()) || false;"},
		{Code: `
async function test() {
  if (await Promise.resolve()) {
  }
}
    `},
		{Code: `
async function test() {
  const mixed: Promise | undefined = Promise.resolve();
  if (mixed) {
    await mixed;
  }
}
    `},
		{Code: `
if (~Promise.resolve()) {
}
    `},
		{Code: `
interface NotQuiteThenable {
  then(param: string): void;
  then(): void;
}
const value: NotQuiteThenable = { then() {} };
if (value) {
}
    `},
		{Code: "[1, 2, 3].forEach(val => {});"},
		{
			Code:    "[1, 2, 3].forEach(async val => {});",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": false}`),
		},
		{Code: "new Promise((resolve, reject) => resolve());"},
		{
			Code:    "new Promise(async (resolve, reject) => resolve());",
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": false}`),
		},
		{Code: `
Promise.all(
  ['abc', 'def'].map(async val => {
    await val;
  }),
);
    `},
		{Code: `
const fn: (arg: () => Promise<void> | void) => void = () => {};
fn(() => Promise.resolve());
    `},
		{Code: `
declare const returnsPromise: (() => Promise<void>) | null;
if (returnsPromise?.()) {
}
    `},
		{Code: `
declare const returnsPromise: { call: () => Promise<void> } | null;
if (returnsPromise?.call()) {
}
    `},
		{Code: "Promise.resolve() ?? false;"},
		{Code: `
function test(a: Promise<void> | undefined) {
  const foo = a ?? Promise.reject();
}
    `},
		{Code: `
function test(p: Promise<boolean> | undefined, bool: boolean) {
  if (p ?? bool) {
  }
}
    `},
		{Code: `
async function test(p: Promise<boolean | undefined>, bool: boolean) {
  if ((await p) ?? bool) {
  }
}
    `},
		{Code: `
async function test(p: Promise<boolean> | undefined) {
  if (await (p ?? Promise.reject())) {
  }
}
    `},
		// added in port
		{Code: `
declare let a: Promise<string> | undefined

declare let b: Promise<string>

a = a ?? b
		`},
		{Code: `
let f;
f = async () => 10;
    `},
		{Code: `
let f: () => Promise<void>;
f = async () => 10;
const g = async () => 0;
const h: () => Promise<void> = async () => 10;
    `},
		{Code: `
const obj = {
  f: async () => 10,
};
    `},
		{Code: `
const f = async () => 123;
const obj = {
  f,
};
    `},
		{Code: `
const obj = {
  async f() {
    return 0;
  },
};
    `},
		{Code: `
type O = { f: () => Promise<void>; g: () => Promise<void> };
const g = async () => 0;
const obj: O = {
  f: async () => 10,
  g,
};
    `},
		{Code: `
type O = { f: () => Promise<void> };
const name = 'f';
const obj: O = {
  async [name]() {
    return 10;
  },
};
    `},
		{Code: `
const obj: number = {
  g() {
    return 10;
  },
};
    `},
		{Code: `
const obj = {
  f: async () => 'foo',
  async g() {
    return 0;
  },
};
    `},
		{Code: `
function f() {
  return async () => 0;
}
function g() {
  return;
}
    `},
		{
			Code: `
type O = {
  bool: boolean;
  func: () => Promise<void>;
};
const Component = (obj: O) => null;
<Component bool func={async () => 10} />;
      `,
			Tsx: true,
		},
		{
			Code: `
const Component: any = () => null;
<Component func={async () => 10} />;
      `,
			Tsx: true,
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => void): void;
  (name: string, callback: () => Promise<void>): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => void): void;
}
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => Promise<void>): void;
}
interface ItLike {
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
		},
		{
			Code: `
interface Props {
  onEvent: (() => void) | (() => Promise<void>);
}

declare function Component(props: Props): any;

const _ = <Component onEvent={async () => {}} />;
      `,
			Tsx: true,
		},
		{Code: `
console.log({ ...(await Promise.resolve({ key: 42 })) });
    `},
		{Code: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...(await getData()),
});
    `},
		{Code: `
declare const condition: boolean;

console.log({ ...(condition && (await Promise.resolve({ key: 42 }))) });
console.log({ ...(condition || (await Promise.resolve({ key: 42 }))) });
console.log({ ...(condition ? {} : await Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? await Promise.resolve({ key: 42 }) : {}) });
    `},
		{Code: `
console.log([...(await Promise.resolve(42))]);
    `},
		{
			Code: `
console.log({ ...Promise.resolve({ key: 42 }) });
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksSpreads": false}`),
		},
		{
			Code: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...getData(),
});
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksSpreads": false}`),
		},
		{
			Code: `
declare const condition: boolean;

console.log({ ...(condition && Promise.resolve({ key: 42 })) });
console.log({ ...(condition || Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? {} : Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? Promise.resolve({ key: 42 }) : {}) });
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksSpreads": false}`),
		},
		{
			Code: `
// This is invalid Typescript, but it shouldn't trigger this linter specifically
console.log([...Promise.resolve(42)]);
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksSpreads": false}`),
		},
		{Code: `
function spreadAny(..._args: any): void {}

spreadAny(
  true,
  () => Promise.resolve(1),
  () => Promise.resolve(false),
);
    `},
		{Code: `
function spreadArrayAny(..._args: Array<any>): void {}

spreadArrayAny(
  true,
  () => Promise.resolve(1),
  () => Promise.resolve(false),
);
    `},
		{Code: `
function spreadArrayUnknown(..._args: Array<unknown>): void {}

spreadArrayUnknown(() => Promise.resolve(true), 1, 2);

function spreadArrayFuncPromise(
  ..._args: Array<() => Promise<undefined>>
): void {}

spreadArrayFuncPromise(
  () => Promise.resolve(undefined),
  () => Promise.resolve(undefined),
);
    `},
		{Code: `
class TakeCallbacks {
  constructor(...callbacks: Array<() => void>) {}
}

new TakeCallbacks;
new TakeCallbacks();
new TakeCallbacks(
  () => 1,
  () => true,
);
    `},
		{Code: `
class Foo {
  public static doThing(): void {}
}

class Bar extends Foo {
  public async doThing(): Promise<void> {}
}
    `},
		{Code: `
class Foo {
  public doThing(): void {}
}

class Bar extends Foo {
  public static async doThing(): Promise<void> {}
}
    `},
		{Code: `
class Foo {
  public doThing = (): void => {};
}

class Bar extends Foo {
  public static doThing = async (): Promise<void> => {};
}
    `},
		{Code: `
class Foo {
  public doThing = (): void => {};
}

class Bar extends Foo {
  public static accessor doThing = async (): Promise<void> => {};
}
    `},
		{Code: `
class Foo {
  public accessor doThing = (): void => {};
}

class Bar extends Foo {
  public static accessor doThing = (): void => {};
}
    `},
		{
			Code: `
class Foo {
  [key: string]: void;
}

class Bar extends Foo {
  [key: string]: Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{Code: `
function restTuple(...args: []): void;
function restTuple(...args: [string]): void;
function restTuple(..._args: string[]): void {}

restTuple();
restTuple('Hello');
    `},
		{Code: `
      let value: Record<string, () => void>;
      value.sync = () => {};
    `},
		{Code: `
      type ReturnsRecord = () => Record<string, () => void>;

      const test: ReturnsRecord = () => {
        return { sync: () => {} };
      };
    `},
		{Code: `
      type ReturnsRecord = () => Record<string, () => void>;

      function sync() {}

      const test: ReturnsRecord = () => {
        return { sync };
      };
    `},
		{Code: `
      function withTextRecurser<Text extends string>(
        recurser: (text: Text) => void,
      ): (text: Text) => void {
        return (text: Text): void => {
          if (text.length) {
            return;
          }

          return recurser(node);
        };
      }
    `},
		{Code: `
declare function foo(cb: undefined | (() => void));
declare const bar: undefined | (() => void);
foo(bar);
    `},
		{
			Code: `
        type OnSelectNodeFn = (node: string | null) => void;

        interface ASTViewerBaseProps {
          readonly onSelectNode?: OnSelectNodeFn;
        }

        declare function ASTViewer(props: ASTViewerBaseProps): null;
        declare const onSelectFn: OnSelectNodeFn;

        <ASTViewer onSelectNode={onSelectFn} />;
      `,
			Tsx:     true,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"attributes": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterfaceExtendsMyInterface extends MyInterface {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterfaceExtendsMyInterface extends MyInterface {
  setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

class MyClassImplementsMyInterface implements MyInterface {
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

class MyClassImplementsMyInterface implements MyInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
type MyTypeLiteralsIntersection = { setThing(): void } & { thing: number };

class MyClass implements MyTypeLiteralsIntersection {
  thing = 1;
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
type MyTypeLiteralsIntersection = { setThing(): void } & { thing: number };

class MyClass implements MyTypeLiteralsIntersection {
  thing = 1;
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType {
  setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType<false> {
  setThing(): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": false}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

interface MyThirdInterface extends MyInterface, MyOtherInterface {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MyOtherClass {
  setThing(): void {
    return;
  }
}

interface MyInterface extends MyClass, MyOtherClass {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclass extends MyClass implements MyInterface, MyOtherInterface {
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

const MyClassExpressionExtendsMyClass = class extends MyClass {
  setThing(): void {
    return;
  }
};
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};

class MyClassExtendsMyClassExpression extends MyClassExpression {
  setThing(): void {
    return;
  }
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};
type MyClassExpressionType = typeof MyClassExpression;

interface MyInterfaceExtendsMyClassExpression extends MyClassExpressionType {
  setThing(): void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MySyncCallSignatures {
  (): void;
  (arg: string): void;
}
interface MyAsyncInterface extends MySyncCallSignatures {
  (): Promise<void>;
  (arg: string): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MySyncConstructSignatures {
  new (): void;
  new (arg: string): void;
}
interface ThisIsADifferentIssue extends MySyncConstructSignatures {
  new (): Promise<void>;
  new (arg: string): Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MySyncIndexSignatures {
  [key: string]: void;
  [key: number]: void;
}
interface ThisIsADifferentIssue extends MySyncIndexSignatures {
  [key: string]: Promise<void>;
  [key: number]: Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MySyncInterfaceSignatures {
  (): void;
  (arg: string): void;
  new (): void;
  [key: string]: () => void;
  [key: number]: () => void;
}
interface MyAsyncInterface extends MySyncInterfaceSignatures {
  (): Promise<void>;
  (arg: string): Promise<void>;
  new (): Promise<void>;
  [key: string]: () => Promise<void>;
  [key: number]: () => Promise<void>;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{
			Code: `
interface MyCall {
  (): void;
  (arg: string): void;
}

interface MyIndex {
  [key: string]: () => void;
  [key: number]: () => void;
}

interface MyConstruct {
  new (): void;
  new (arg: string): void;
}

interface MyMethods {
  doSyncThing(): void;
  doOtherSyncThing(): void;
  syncMethodProperty: () => void;
}
interface MyInterface extends MyCall, MyIndex, MyConstruct, MyMethods {
  (): void;
  (arg: string): void;
  new (): void;
  new (arg: string): void;
  [key: string]: () => void;
  [key: number]: () => void;
  doSyncThing(): void;
  doAsyncThing(): Promise<void>;
  syncMethodProperty: () => void;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"inheritedMethods": true}}`),
		},
		{Code: "const notAFn1: string = '';"},
		{Code: "const notAFn2: number = 1;"},
		{Code: "const notAFn3: boolean = true;"},
		{Code: "const notAFn4: { prop: 1 } = { prop: 1 };"},
		{Code: "const notAFn5: {} = {};"},
		{Code: `
const array: number[] = [1, 2, 3];
array.filter(a => a > 1);
    `},
		{Code: `
type ReturnsPromiseVoid = () => Promise<void>;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsPromiseVoid>(async () => {});
    `},
		{Code: `
type ReturnsVoid = () => void;
type ReturnsPromiseVoid = () => Promise<void>;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsVoid | ReturnsPromiseVoid>(async () => {});
    `},
	{Code: `if (process.env.SKIP) { return x; }`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
if (Promise.resolve()) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      2,
				},
			},
		},
		{
			Code: `
if (Promise.resolve()) {
} else if (Promise.resolve()) {
} else {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      2,
				},
				{
					MessageId: "conditional",
					Line:      3,
				},
			},
		},
		{
			Code: "for (let i; Promise.resolve(); i++) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      1,
				},
			},
		},
		{
			Code: "do {} while (Promise.resolve());",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      1,
				},
			},
		},
		{
			Code: "while (Promise.resolve()) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      1,
				},
			},
		},
		{
			Code: "Promise.resolve() ? 123 : 456;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      1,
				},
			},
		},
		{
			Code: `
if (!Promise.resolve()) {
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      2,
				},
			},
		},
		{
			Code: "Promise.resolve() || false;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      1,
				},
			},
		},
		{
			Code: `
[Promise.resolve(), Promise.reject()].forEach(async val => {
  await val;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      2,
				},
			},
		},
		{
			Code: `
new Promise(async (resolve, reject) => {
  await Promise.resolve();
  resolve();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      2,
				},
			},
		},
		{
			Code: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', async (err, res) => {
  await res;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
			},
		},
		{
			Code: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => Promise.resolve(res));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
			},
		},
		{
			Code: `
const fnWithCallback = (arg: string, cb: (err: any, res: string) => void) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => {
  if (err) {
    return 'abc';
  } else {
    return Promise.resolve(res);
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
			},
		},
		{
			Code: `
const fnWithCallback:
  | ((arg: string, cb: (err: any, res: string) => void) => void)
  | null = (arg, cb) => {
  cb(null, arg);
};

fnWithCallback?.('val', (err, res) => Promise.resolve(res));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      8,
				},
			},
		},
		{
			Code: `
const fnWithCallback:
  | ((arg: string, cb: (err: any, res: string) => void) => void)
  | null = (arg, cb) => {
  cb(null, arg);
};

fnWithCallback('val', (err, res) => {
  if (err) {
    return 'abc';
  } else {
    return Promise.resolve(res);
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      8,
				},
			},
		},
		{
			Code: `
function test(bool: boolean, p: Promise<void>) {
  if (bool || p) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      3,
				},
			},
		},
		{
			Code: `
function test(bool: boolean, p: Promise<void>) {
  if (bool && p) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      3,
				},
			},
		},
		{
			Code: `
function test(a: any, p: Promise<void>) {
  if (a ?? p) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      3,
				},
			},
		},
		{
			Code: `
function test(p: Promise<void> | undefined) {
  if (p ?? Promise.reject()) {
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditional",
					Line:      3,
				},
			},
		},
		{
			Code: `
let f: () => void;
f = async () => {
  return 3;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnVariable",
					Line:      3,
				},
			},
		},
		{
			Code: `
let f: () => void;
f = async () => {
  return 3;
};
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"variables": true}}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnVariable",
					Line:      3,
				},
			},
		},
		{
			Code: `
const f: () => void = async () => {
  return 0;
};
const g = async () => 1,
  h: () => void = async () => {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnVariable",
					Line:      2,
				},
				{
					MessageId: "voidReturnVariable",
					Line:      6,
				},
			},
		},
		{
			Code: `
const obj: {
  f?: () => void;
} = {};
obj.f = async () => {
  return 0;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnVariable",
					Line:      5,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const obj: O = {
  f: async () => 'foo',
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 4,
					// Column: 3,
					// EndLine: 4,
					// EndColumn: 12,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const obj: O = {
  f: async () => 'foo',
};
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"properties": true}}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 4,
					// Column: 3,
					// EndLine: 4,
					// EndColumn: 12,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const f = async () => 0;
const obj: O = {
  f,
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      5,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const obj: O = {
  async f() {
    return 0;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 4,
					// Column: 3,
					// EndLine: 4,
					// EndColumn: 10,
				},
			},
		},
		{
			Code: `
type O = { f: () => void; g: () => void; h: () => void };
function f(): O {
  const h = async () => 0;
  return {
    async f() {
      return 123;
    },
    g: async () => 0,
    h,
  };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 6,
					// Column: 5,
					// EndLine: 6,
					// EndColumn: 12,
				},
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 9,
					// Column: 5,
					// EndLine: 9,
					// EndColumn: 14,
				},
				{
					MessageId: "voidReturnProperty",
					Line:      10,
					Column:    5,
					EndLine:   10,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function f(): () => void {
  return async () => 0;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnReturnValue",
					Line:      3,
				},
			},
		},
		{
			Code: `
function f(): () => void {
  return async () => 0;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"returns": true}}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnReturnValue",
					Line:      3,
				},
			},
		},
		{
			Code: `
type O = {
  func: () => void;
};
const Component = (obj: O) => null;
<Component func={async () => 0} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnAttribute",
					Line:      6,
				},
			},
		},
		{
			Code: `
type O = {
  func: () => void;
};
const Component = (obj: O) => null;
<Component func={async () => 0} />;
      `,
			Tsx: true,
			// TODO(port) getContextualTypeForJsxExpression is not yet implemented
			Skip:    true,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksVoidReturn": {"attributes": true}}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnAttribute",
					Line:      6,
				},
			},
		},
		{
			Code: `
type O = {
  func: () => void;
};
const g = async () => 'foo';
const Component = (obj: O) => null;
<Component func={g} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnAttribute",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => number): void;
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      9,
				},
			},
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => number): void;
}
interface ItLike {
  (name: string, callback: () => void): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      11,
				},
			},
		},
		{
			Code: `
interface ItLike {
  (name: string, callback: () => void): void;
}
interface ItLike {
  (name: string, callback: () => number): void;
}

declare const it: ItLike;

it('', async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      11,
				},
			},
		},
		{
			Code: `
console.log({ ...Promise.resolve({ key: 42 }) });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "spread",
					Line:      2,
				},
			},
		},
		{
			Code: `
const getData = () => Promise.resolve({ key: 42 });

console.log({
  someData: 42,
  ...getData(),
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "spread",
					Line:      6,
				},
			},
		},
		{
			Code: `
declare const condition: boolean;

console.log({ ...(condition && Promise.resolve({ key: 42 })) });
console.log({ ...(condition || Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? {} : Promise.resolve({ key: 42 })) });
console.log({ ...(condition ? Promise.resolve({ key: 42 }) : {}) });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "spread",
					Line:      4,
				},
				{
					MessageId: "spread",
					Line:      5,
				},
				{
					MessageId: "spread",
					Line:      6,
				},
				{
					MessageId: "spread",
					Line:      7,
				},
			},
		},
		{
			Code: `
function restPromises(first: Boolean, ...callbacks: Array<() => void>): void {}

restPromises(
  true,
  () => Promise.resolve(true),
  () => Promise.resolve(null),
  () => true,
  () => Promise.resolve('Hello'),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
				{
					MessageId: "voidReturnArgument",
					Line:      7,
				},
				{
					MessageId: "voidReturnArgument",
					Line:      9,
				},
			},
		},
		{
			Code: `
type MyUnion = (() => void) | boolean;

function restUnion(first: string, ...callbacks: Array<MyUnion>): void {}
restUnion('Testing', false, () => Promise.resolve(true));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      5,
				},
			},
		},
		{
			Code: `
function restTupleOne(first: string, ...callbacks: [() => void]): void {}
restTupleOne('My string', () => Promise.resolve(1));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      3,
				},
			},
		},
		{
			Code: `
function restTupleTwo(
  first: boolean,
  ...callbacks: [undefined, () => void, undefined]
): void {}

restTupleTwo(true, undefined, () => Promise.resolve(true), undefined);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      7,
				},
			},
		},
		{
			Code: `
function restTupleFour(
  first: number,
  ...callbacks: [() => void, boolean, () => void, () => void]
): void;

restTupleFour(
  1,
  () => Promise.resolve(true),
  false,
  () => {},
  () => Promise.resolve(1),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      9,
				},
				{
					MessageId: "voidReturnArgument",
					Line:      12,
				},
			},
		},
		{
			Code: `
class TakesVoidCb {
  constructor(first: string, ...args: Array<() => void>);
}

new TakesVoidCb;
new TakesVoidCb();
new TakesVoidCb(
  'Testing',
  () => {},
  () => Promise.resolve(true),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      11,
				},
			},
		},
		{
			Code: `
function restTuple(...args: []): void;
function restTuple(...args: [boolean, () => void]): void;
function restTuple(..._args: any[]): void {}

restTuple();
restTuple(true, () => Promise.resolve(1));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      7,
				},
			},
		},
		{
			Code: `
type ReturnsRecord = () => Record<string, () => void>;

const test: ReturnsRecord = () => {
  return { asynchronous: async () => {} };
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 5,
					// Column: 12,
					// EndLine: 5,
					// EndColumn: 32,
				},
			},
		},
		{
			Code: `
let value: Record<string, () => void>;
value.asynchronous = async () => {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnVariable",
					Line:      3,
				},
			},
		},
		{
			Code: `
type ReturnsRecord = () => Record<string, () => void>;

async function asynchronous() {}

const test: ReturnsRecord = () => {
  return { asynchronous };
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      7,
				},
			},
		},
		{
			Code: `
declare function foo(cb: undefined | (() => void));
declare const bar: undefined | (() => Promise<void>);
foo(bar);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      4,
				},
			},
		},
		{
			Code: `
declare function foo(cb: string & (() => void));
declare const bar: string & (() => Promise<void>);
foo(bar);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      4,
				},
			},
		},
		{
			Code: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs: Array<() => Promise<boolean>> = [
  () => Promise.resolve(true),
  () => Promise.resolve(true),
];
consume(...cbs);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      7,
				},
			},
		},
		{
			Code: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs = [() => Promise.resolve(true), () => Promise.resolve(true)] as const;
consume(...cbs);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      4,
				},
			},
		},
		{
			Code: `
function consume(..._callbacks: Array<() => void>): void {}
let cbs = [() => Promise.resolve(true), () => Promise.resolve(true)];
consume(...cbs);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      4,
				},
			},
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclassExtendsMyClass extends MyClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      9,
				},
			},
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

abstract class MyAbstractClassExtendsMyClass extends MyClass {
  abstract setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      9,
				},
			},
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

interface MyInterfaceExtendsMyClass extends MyClass {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      9,
				},
			},
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

class MySubclassExtendsMyAbstractClass extends MyAbstractClass {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

abstract class MyAbstractSubclassExtendsMyAbstractClass extends MyAbstractClass {
  abstract setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
abstract class MyAbstractClass {
  abstract setThing(): void;
}

interface MyInterfaceExtendsMyAbstractClass extends MyAbstractClass {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

class MyInterfaceSubclass implements MyInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

abstract class MyAbstractClassImplementsMyInterface implements MyInterface {
  abstract setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
class MyClass {
  accessor setThing = (): void => {
    return;
  };
}

class MySubclassExtendsMyClass extends MyClass {
  accessor setThing = async (): Promise<void> => {
    await Promise.resolve();
  };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      9,
				},
			},
		},
		{
			Code: `
abstract class MyClass {
  abstract accessor setThing: () => void;
}

abstract class MySubclassExtendsMyClass extends MyClass {
  abstract accessor setThing: () => Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MySubInterface extends MyInterface {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
type MyTypeIntersection = { setThing(): void } & { thing: number };

class MyClassImplementsMyTypeIntersection implements MyTypeIntersection {
  thing = 1;
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      6,
				},
			},
		},
		{
			Code: `
type MyGenericType<IsAsync extends boolean = true> = IsAsync extends true
  ? { setThing(): Promise<void> }
  : { setThing(): void };

interface MyAsyncInterface extends MyGenericType<false> {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

interface MyOtherInterface {
  setThing(): void;
}

interface MyThirdInterface extends MyInterface, MyOtherInterface {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      11,
				},
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      11,
				},
			},
		},
		{
			Code: `
class MyClass {
  setThing(): void {
    return;
  }
}

class MyOtherClass {
  setThing(): void {
    return;
  }
}

interface MyInterface extends MyClass, MyOtherClass {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      15,
				},
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      15,
				},
			},
		},
		{
			Code: `
interface MyAsyncInterface {
  setThing(): Promise<void>;
}

interface MySyncInterface {
  setThing(): void;
}

class MyClass {
  setThing(): void {
    return;
  }
}

class MySubclass extends MyClass implements MyAsyncInterface, MySyncInterface {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      17,
				},
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      17,
				},
			},
		},
		{
			Code: `
interface MyInterface {
  setThing(): void;
}

const MyClassExpressionExtendsMyClass = class implements MyInterface {
  setThing(): Promise<void> {
    await Promise.resolve();
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      7,
				},
			},
		},
		{
			Code: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};

class MyClassExtendsMyClassExpression extends MyClassExpression {
  async setThing(): Promise<void> {
    await Promise.resolve();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      9,
				},
			},
		},
		{
			Code: `
const MyClassExpression = class {
  setThing(): void {
    return;
  }
};
type MyClassExpressionType = typeof MyClassExpression;

interface MyInterfaceExtendsMyClassExpression extends MyClassExpressionType {
  setThing(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      10,
				},
			},
		},
		{
			Code: `
interface MySyncInterface {
  (): void;
  (arg: string): void;
  new (): void;
  [key: string]: () => void;
  [key: number]: () => void;
  myMethod(): void;
}
interface MyAsyncInterface extends MySyncInterface {
  (): Promise<void>;
  (arg: string): Promise<void>;
  new (): Promise<void>;
  [key: string]: () => Promise<void>;
  [key: number]: () => Promise<void>;
  myMethod(): Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      16,
				},
			},
		},
		{
			Code: `
interface MyCall {
  (): void;
  (arg: string): void;
}

interface MyIndex {
  [key: string]: () => void;
  [key: number]: () => void;
}

interface MyConstruct {
  new (): void;
  new (arg: string): void;
}

interface MyMethods {
  doSyncThing(): void;
  doOtherSyncThing(): void;
  syncMethodProperty: () => void;
}
interface MyInterface extends MyCall, MyIndex, MyConstruct, MyMethods {
  (): void;
  (arg: string): Promise<void>;
  new (): void;
  new (arg: string): void;
  [key: string]: () => Promise<void>;
  [key: number]: () => void;
  doSyncThing(): Promise<void>;
  doAsyncThing(): Promise<void>;
  syncMethodProperty: () => Promise<void>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      29,
				},
				{
					MessageId: "voidReturnInheritedMethod",
					Line:      31,
				},
			},
		},
		{
			Code: `
declare function isTruthy(value: unknown): Promise<boolean>;
[0, 1, 2].filter(isTruthy);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "predicate",
					Line:      3,
				},
			},
		},
		{
			Code: `
const array: number[] = [];
array.every(() => Promise.resolve(true));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "predicate",
					Line:      3,
				},
			},
		},
		{
			Code: `
const array: (string[] & { foo: 'bar' }) | (number[] & { bar: 'foo' }) = [];
array.every(() => Promise.resolve(true));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "predicate",
					Line:      3,
				},
			},
		},
		{
			Code: `
const tuple: [number, number, number] = [1, 2, 3];
tuple.find(() => Promise.resolve(false));
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedPromisesOptions](`{"checksConditionals": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "predicate",
					Line:      3,
				},
			},
		},
		{
			Code: `
type ReturnsVoid = () => void;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
declare const useCallbackReturningVoid: typeof useCallback<ReturnsVoid>;
useCallbackReturningVoid(async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      7,
				},
			},
		},
		{
			Code: `
type ReturnsVoid = () => void;
declare const useCallback: <T extends (...args: unknown[]) => unknown>(
  fn: T,
) => T;
useCallback<ReturnsVoid>(async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
			},
		},
		{
			Code: `
interface Foo<T> {
  (callback: () => T): void;
  (callback: () => number): void;
}
declare const foo: Foo<void>;

foo(async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      8,
				},
			},
		},
		{
			Code: `
declare function tupleFn<T extends (...args: unknown[]) => unknown>(
  ...fns: [T, string, T]
): void;
tupleFn<() => void>(
  async () => {},
  'foo',
  async () => {},
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
				{
					MessageId: "voidReturnArgument",
					Line:      8,
				},
			},
		},
		{
			Code: `
declare function arrayFn<T extends (...args: unknown[]) => unknown>(
  ...fns: (T | string)[]
): void;
arrayFn<() => void>(
  async () => {},
  'foo',
  async () => {},
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnArgument",
					Line:      6,
				},
				{
					MessageId: "voidReturnArgument",
					Line:      8,
				},
			},
		},
		{
			Code: `
type HasVoidMethod = {
  f(): void;
};

const o: HasVoidMethod = {
  async f() {
    return 3;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 7,
					// Column: 3,
					// EndLine: 7,
					// EndColumn: 10,
				},
			},
		},
		{
			Code: `
type HasVoidMethod = {
  f(): void;
};

const o: HasVoidMethod = {
  async f(): Promise<number> {
    return 3;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      7,
					Column:    14,
					EndLine:   7,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
type HasVoidMethod = {
  f(): void;
};
const obj: HasVoidMethod = {
  f() {
    return Promise.resolve('foo');
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					// TODO(port): implement getFunctionHeadLoc
					// Line: 6,
					// Column: 3,
					// EndLine: 6,
					// EndColumn: 4,
				},
			},
		},
		{
			Code: `
type HasVoidMethod = {
  f(): void;
};
const obj: HasVoidMethod = {
  f(): Promise<void> {
    throw new Error();
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      6,
					Column:    8,
					EndLine:   6,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const asyncFunction = async () => 'foo';
const obj: O = {
  f: asyncFunction,
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      5,
					Column:    6,
					EndLine:   5,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
type O = { f: () => void };
const obj: O = {
  f: async (): Promise<string> => 'foo',
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "voidReturnProperty",
					Line:      4,
					Column:    16,
					EndLine:   4,
					EndColumn: 31,
				},
			},
		},
	})
}
