package no_floating_promises

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestNoFloatingPromisesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoFloatingPromisesRule, []rule_tester.ValidTestCase{
		{Code: `
async function test() {
  await Promise.resolve('value');
  Promise.resolve('value').then(
    () => {},
    () => {},
  );
  Promise.resolve('value')
    .then(() => {})
    .catch(() => {});
  Promise.resolve('value')
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  Promise.resolve('value').catch(() => {});
  return Promise.resolve('value');
}
    `},
		{
			Code: `
async function test() {
  void Promise.resolve('value');
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(true)},
		},
		{Code: `
async function test() {
  await Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(
    () => {},
    () => {},
  );
  Promise.reject(new Error('message'))
    .then(() => {})
    .catch(() => {});
  Promise.reject(new Error('message'))
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  Promise.reject(new Error('message')).catch(() => {});
  return Promise.reject(new Error('message'));
}
    `},
		{Code: `
async function test() {
  await (async () => true)();
  (async () => true)().then(
    () => {},
    () => {},
  );
  (async () => true)()
    .then(() => {})
    .catch(() => {});
  (async () => true)()
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  (async () => true)().catch(() => {});
  return (async () => true)();
}
    `},
		{Code: `
async function test() {
  async function returnsPromise() {}
  await returnsPromise();
  returnsPromise().then(
    () => {},
    () => {},
  );
  returnsPromise()
    .then(() => {})
    .catch(() => {});
  returnsPromise()
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  returnsPromise().catch(() => {});
  return returnsPromise();
}
    `},
		{Code: `
async function test() {
  const x = Promise.resolve();
  const y = x.then(() => {});
  y.catch(() => {});
}
    `},
		{Code: `
async function test() {
  Math.random() > 0.5 ? Promise.resolve().catch(() => {}) : null;
}
    `},
		{Code: `
async function test() {
  Promise.resolve().catch(() => {}), 123;
  123,
    Promise.resolve().then(
      () => {},
      () => {},
    );
  123,
    Promise.resolve().then(
      () => {},
      () => {},
    ),
    123;
}
    `},
		{Code: `
async function test() {
  void Promise.resolve().catch(() => {});
}
    `},
		{Code: `
async function test() {
  Promise.resolve().catch(() => {}) ||
    Promise.resolve().then(
      () => {},
      () => {},
    );
}
    `},
		{Code: `
declare const promiseValue: Promise<number>;
async function test() {
  await promiseValue;
  promiseValue.then(
    () => {},
    () => {},
  );
  promiseValue.then(() => {}).catch(() => {});
  promiseValue
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  promiseValue.catch(() => {});
  return promiseValue;
}
    `},
		{Code: `
declare const promiseUnion: Promise<number> | number;
async function test() {
  await promiseUnion;
  promiseUnion.then(
    () => {},
    () => {},
  );
  promiseUnion.then(() => {}).catch(() => {});
  promiseUnion
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  promiseUnion.catch(() => {});
  promiseValue.finally(() => {});
  return promiseUnion;
}
    `},
		{Code: `
declare const promiseIntersection: Promise<number> & number;
async function test() {
  await promiseIntersection;
  promiseIntersection.then(
    () => {},
    () => {},
  );
  promiseIntersection.then(() => {}).catch(() => {});
  promiseIntersection.catch(() => {});
  return promiseIntersection;
}
    `},
		{Code: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  await canThen;
  canThen.then(
    () => {},
    () => {},
  );
  canThen.then(() => {}).catch(() => {});
  canThen
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  canThen.catch(() => {});
  return canThen;
}
    `},
		{Code: `
declare const intersectionPromise: Promise<number> & number;
async function test() {
  await (Math.random() > 0.5 ? numberPromise : 0);
  await (Math.random() > 0.5 ? foo : 0);
  await (Math.random() > 0.5 ? bar : 0);

  await intersectionPromise;
}
    `},
		{Code: `
async function test() {
  class Thenable {
    then(callback: () => void): Thenable {
      return new Thenable();
    }
  }
  const thenable = new Thenable();

  await thenable;
  thenable;
  thenable.then(() => {});
  return thenable;
}
    `},
		{Code: `
async function test() {
  class NonFunctionParamThenable {
    then(param: string, param2: number): NonFunctionParamThenable {
      return new NonFunctionParamThenable();
    }
  }
  const thenable = new NonFunctionParamThenable();

  await thenable;
  thenable;
  thenable.then('abc', 'def');
  return thenable;
}
    `},
		{Code: `
async function test() {
  class NonFunctionThenable {
    then: number;
  }
  const thenable = new NonFunctionThenable();

  thenable;
  thenable.then;
  return thenable;
}
    `},
		{Code: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  await thenable;
  return thenable;
}
    `},
		{Code: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  await promise;
  promise.then(
    () => {},
    () => {},
  );
  promise.then(() => {}).catch(() => {});
  promise
    .then(() => {})
    .catch(() => {})
    .finally(() => {});
  promise.catch(() => {});
  return promise;
}
    `},
		{Code: `
declare const returnsPromise: () => Promise<void> | null;
async function test() {
  await returnsPromise?.();
  returnsPromise()?.then(
    () => {},
    () => {},
  );
  returnsPromise()
    ?.then(() => {})
    ?.catch(() => {});
  returnsPromise()?.catch(() => {});
  return returnsPromise();
}
    `},
		{Code: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  await obj1.a?.b?.c?.();
  await obj2.a?.b?.c();
  await obj3.a?.b.c?.();
  await obj4.a.b.c?.();
  await obj5.a?.().b?.c?.();
  await obj6?.a.b.c?.();

  return callback?.();
};

void doSomething();
    `},
		{
			Code: `
        (async () => {
          await something();
        })();
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code: `
        (async () => {
          something();
        })();
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code:    "(async function foo() {})();",
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code: `
        function foo() {
          (async function bar() {})();
        }
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code: `
        const foo = () =>
          new Promise(res => {
            (async function () {
              await res(1);
            })();
          });
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code: `
        (async function () {
          await res(1);
        })();
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;
  void (condition && myPromise());
}
      `,
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;
  await (condition && myPromise());
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;
  condition && void myPromise();
}
      `,
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;
  condition && (await myPromise());
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  let condition = false;
  condition && myPromise();
  condition = true;
  condition || myPromise();
  condition ?? myPromise();
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
declare const definitelyCallable: () => void;
Promise.reject().catch(definitelyCallable);
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
Promise.reject()
  .catch(() => {})
  .finally(() => {});
      `,
		},
		{
			Code: `
Promise.reject()
  .catch(() => {})
  .finally(() => {})
  .finally(() => {});
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
Promise.reject()
  .catch(() => {})
  .finally(() => {})
  .finally(() => {})
  .finally(() => {});
      `,
		},
		{
			Code: `
await Promise.all([Promise.resolve(), Promise.resolve()]);
      `,
		},
		{
			Code: `
declare const promiseArray: Array<Promise<unknown>>;
void promiseArray;
      `,
		},
		{
			Code: `
[1, 2, void Promise.reject(), 3];
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
		},
		{
			Code: `
['I', 'am', 'just', 'an', 'array'];
      `,
		},
		{
			Code: `
interface SafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | SafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | SafeThenable<TResult2>)
      | undefined
      | null,
  ): SafeThenable<TResult1 | TResult2>;
}
let promise: SafeThenable<number> = Promise.resolve(5);
0, promise;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
			},
		},
		{
			Code: `
interface SafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | SafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | SafeThenable<TResult2>)
      | undefined
      | null,
  ): SafeThenable<TResult1 | TResult2>;
}
let promise: SafeThenable<number> = Promise.resolve(5);
0 ? promise : 3;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
			},
		},
		{
			Code: `
class SafePromise<T> extends Promise<T> {}
let promise: { a: SafePromise<number> } = { a: Promise.resolve(5) };
promise.a;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
		},
		{
			Code: `
class SafePromise<T> extends Promise<T> {}
let promise: SafePromise<number> = Promise.resolve(5);
promise;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
		},
		{
			Code: `
type Foo = Promise<number> & { hey?: string };
let promise: Foo = Promise.resolve(5);
0 || promise;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
type Foo = Promise<number> & { hey?: string };
let promise: Foo = Promise.resolve(5);
promise.finally();
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
interface SafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | SafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | SafeThenable<TResult2>)
      | undefined
      | null,
  ): SafeThenable<TResult1 | TResult2>;
}
let promise: () => SafeThenable<number> = () => Promise.resolve(5);
0, promise();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
			},
		},
		{
			Code: `
interface SafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | SafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | SafeThenable<TResult2>)
      | undefined
      | null,
  ): SafeThenable<TResult1 | TResult2>;
}
let promise: () => SafeThenable<number> = () => Promise.resolve(5);
0 ? promise() : 3;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
			},
		},
		{
			Code: `
type Foo = Promise<number> & { hey?: string };
let promise: () => Foo = () => Promise.resolve(5);
promise();
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
type Foo = Promise<number> & { hey?: string };
let promise: () => Foo = async () => 5;
promise().finally();
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
class SafePromise<T> extends Promise<T> {}
let promise: () => SafePromise<number> = async () => 5;
0 || promise();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
		},
		{
			Code: `
class SafePromise<T> extends Promise<T> {}
let promise: () => SafePromise<number> = async () => 5;
null ?? promise();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
		},
		{
			Code: `
let promise: () => PromiseLike<number> = () => Promise.resolve(5);
promise();
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromLib, Name: []string{"PromiseLike"}}}},
		},
		{
			Code: `
type Foo<T> = Promise<T> & { hey?: string };
declare const arrayOrPromiseTuple: Foo<unknown>[];
arrayOrPromiseTuple;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
type Foo<T> = Promise<T> & { hey?: string };
declare const arrayOrPromiseTuple: [Foo<unknown>, 5];
arrayOrPromiseTuple;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
		},
		{
			Code: `
type SafePromise = Promise<number> & { __linterBrands?: string };
declare const myTag: (strings: TemplateStringsArray) => SafePromise;
myTag` + "`" + `abc` + "`" + `;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
		},
		{
			Code: `
        declare function it(...args: unknown[]): Promise<void>;

        it('...', () => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "file.ts"}},
			},
		},
		{
			Code: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
myTag` + "`" + `abc` + "`" + `.catch(() => {});
      `,
		},
		{
			Code: `
declare const myTag: (strings: TemplateStringsArray) => string;
myTag` + "`" + `abc` + "`" + `;
      `,
		},
		{
			Code: `
declare let x: any;
declare const promiseArray: Array<Promise<unknown>>;
x = promiseArray;
      `,
		},
		{
			Code: `
declare let x: Promise<number>;
x = Promise.resolve(2);
      `,
		},
		{
			Code: `
declare const promiseArray: Array<Promise<unknown>>;
async function f() {
  return promiseArray;
}
      `,
		},
		{
			Code: `
declare const promiseArray: Array<Promise<unknown>>;
async function* generator() {
  yield* promiseArray;
}
      `,
		},
		{
			Code: `
async function* generator() {
  yield Promise.resolve();
}
      `,
		},
		{
			Code: `
interface SafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | SafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | SafeThenable<TResult2>)
      | undefined
      | null,
  ): SafeThenable<TResult1 | TResult2>;
}
let promise: () => SafeThenable<number> = () => Promise.resolve(5);
promise().then(() => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
			},
		},
		{
			Code: `
        declare module 'abc' {
          export function it(name: string, action: () => void): void;
        }
        it('...', () => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "abc"}},
			},
		},
		{
			Code: `
        declare module 'abc' {
          export function it(name: string, action: () => void): void;
        }

        it('...', () => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "abc"}},
			},
		},
		{
			Skip: true,
			Code: `
        import { it } from 'node:test';

        it('...', () => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "node:test"}},
			},
		},
		{
			Code: `
interface SafePromise<T> extends Promise<T> {
  brand: 'safe';
}

declare const createSafePromise: () => SafePromise<string>;
createSafePromise();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
				CheckThenables:            utils.Ref(true),
			},
		},
		{Code: `
declare const createPromiseLike: () => PromiseLike<number>;
createPromiseLike();
    `},
		{Code: `
interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

createMyThenable();
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
async function test() {
  Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.resolve('value');
  void Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.resolve('value');
  await Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  void Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  await Promise.resolve('value').catch();
  Promise.resolve('value').finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  void Promise.resolve('value').finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.resolve('value');
  Promise.resolve('value').then(() => {});
  Promise.resolve('value').catch();
  await Promise.resolve('value').finally();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  void obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  await obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  void obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  await obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  void obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  await obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  void obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  await obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  void obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  await obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  void obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  await obj6?.a.b.c?.();

  callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  void callback?.();
};

doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  await callback?.();
};

doSomething();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

void doSomething();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const doSomething = async (
  obj1: { a?: { b?: { c?: () => Promise<void> } } },
  obj2: { a?: { b?: { c: () => Promise<void> } } },
  obj3: { a?: { b: { c?: () => Promise<void> } } },
  obj4: { a: { b: { c?: () => Promise<void> } } },
  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },
  obj6?: { a: { b: { c?: () => Promise<void> } } },
  callback?: () => Promise<void>,
): Promise<void> => {
  obj1.a?.b?.c?.();
  obj2.a?.b?.c();
  obj3.a?.b.c?.();
  obj4.a.b.c?.();
  obj5.a?.().b?.c?.();
  obj6?.a.b.c?.();

  callback?.();
};

await doSomething();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
myTag` + "`" + `abc` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
void myTag` + "`" + `abc` + "`" + `;
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
await myTag` + "`" + `abc` + "`" + `;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
myTag` + "`" + `abc` + "`" + `.then(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
void myTag` + "`" + `abc` + "`" + `.then(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
await myTag` + "`" + `abc` + "`" + `.then(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
myTag` + "`" + `abc` + "`" + `.finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
void myTag` + "`" + `abc` + "`" + `.finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const myTag: (strings: TemplateStringsArray) => Promise<void>;
await myTag` + "`" + `abc` + "`" + `.finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  Promise.resolve('value');
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void Promise.resolve('value');
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await Promise.resolve('value');
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  void Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  await Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  void Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  await Promise.reject(new Error('message')).catch();
  Promise.reject(new Error('message')).finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  void Promise.reject(new Error('message')).finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.reject(new Error('message'));
  Promise.reject(new Error('message')).then(() => {});
  Promise.reject(new Error('message')).catch();
  await Promise.reject(new Error('message')).finally();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  (async () => true)();
  (async () => true)().then(() => {});
  (async () => true)().catch();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void (async () => true)();
  (async () => true)().then(() => {});
  (async () => true)().catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await (async () => true)();
  (async () => true)().then(() => {});
  (async () => true)().catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  (async () => true)();
  void (async () => true)().then(() => {});
  (async () => true)().catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  (async () => true)();
  await (async () => true)().then(() => {});
  (async () => true)().catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  (async () => true)();
  (async () => true)().then(() => {});
  void (async () => true)().catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  (async () => true)();
  (async () => true)().then(() => {});
  await (async () => true)().catch();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  returnsPromise().then(() => {});
  returnsPromise().catch();
  returnsPromise().finally();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  async function returnsPromise() {}

  void returnsPromise();
  returnsPromise().then(() => {});
  returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  async function returnsPromise() {}

  await returnsPromise();
  returnsPromise().then(() => {});
  returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  void returnsPromise().then(() => {});
  returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  await returnsPromise().then(() => {});
  returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  returnsPromise().then(() => {});
  void returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  returnsPromise().then(() => {});
  await returnsPromise().catch();
  returnsPromise().finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  returnsPromise().then(() => {});
  returnsPromise().catch();
  void returnsPromise().finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  async function returnsPromise() {}

  returnsPromise();
  returnsPromise().then(() => {});
  returnsPromise().catch();
  await returnsPromise().finally();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  Math.random() > 0.5 ? Promise.resolve() : null;
  Math.random() > 0.5 ? null : Promise.resolve();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void (Math.random() > 0.5 ? Promise.resolve() : null);
  Math.random() > 0.5 ? null : Promise.resolve();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await (Math.random() > 0.5 ? Promise.resolve() : null);
  Math.random() > 0.5 ? null : Promise.resolve();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Math.random() > 0.5 ? Promise.resolve() : null;
  void (Math.random() > 0.5 ? null : Promise.resolve());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Math.random() > 0.5 ? Promise.resolve() : null;
  await (Math.random() > 0.5 ? null : Promise.resolve());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  Promise.resolve(), 123;
  123, Promise.resolve();
  123, Promise.resolve(), 123;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void (Promise.resolve(), 123);
  123, Promise.resolve();
  123, Promise.resolve(), 123;
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await (Promise.resolve(), 123);
  123, Promise.resolve();
  123, Promise.resolve(), 123;
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.resolve(), 123;
  void (123, Promise.resolve());
  123, Promise.resolve(), 123;
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.resolve(), 123;
  await (123, Promise.resolve());
  123, Promise.resolve(), 123;
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  Promise.resolve(), 123;
  123, Promise.resolve();
  void (123, Promise.resolve(), 123);
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  Promise.resolve(), 123;
  123, Promise.resolve();
  await (123, Promise.resolve(), 123);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  void Promise.resolve();
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await Promise.resolve();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  const promise = new Promise((resolve, reject) => resolve('value'));
  promise;
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  const promise = new Promise((resolve, reject) => resolve('value'));
  await promise;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function returnsPromise() {
  return 'value';
}
void returnsPromise();
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function returnsPromise() {
  return 'value';
}
await returnsPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function returnsPromise() {
  return 'value';
}
void /* ... */ returnsPromise();
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function returnsPromise() {
  return 'value';
}
await /* ... */ returnsPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function returnsPromise() {
  return 'value';
}
1, returnsPromise();
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function returnsPromise() {
  return 'value';
}
await (1, returnsPromise());
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function returnsPromise() {
  return 'value';
}
bool ? returnsPromise() : null;
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function returnsPromise() {
  return 'value';
}
await (bool ? returnsPromise() : null);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  const obj = { foo: Promise.resolve() };
  obj.foo;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  const obj = { foo: Promise.resolve() };
  void obj.foo;
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  const obj = { foo: Promise.resolve() };
  await obj.foo;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  new Promise(resolve => resolve());
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  void new Promise(resolve => resolve());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  await new Promise(resolve => resolve());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  promiseValue.then(() => {});
  promiseValue.catch();
  promiseValue.finally();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  void promiseValue;
  promiseValue.then(() => {});
  promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  await promiseValue;
  promiseValue.then(() => {});
  promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  void promiseValue.then(() => {});
  promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  await promiseValue.then(() => {});
  promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  promiseValue.then(() => {});
  void promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  promiseValue.then(() => {});
  await promiseValue.catch();
  promiseValue.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  promiseValue.then(() => {});
  promiseValue.catch();
  void promiseValue.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseValue: Promise<number>;

async function test() {
  promiseValue;
  promiseValue.then(() => {});
  promiseValue.catch();
  await promiseValue.finally();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const promiseUnion: Promise<number> | number;

async function test() {
  promiseUnion;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseUnion: Promise<number> | number;

async function test() {
  void promiseUnion;
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseUnion: Promise<number> | number;

async function test() {
  await promiseUnion;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  void promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  await promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  promiseIntersection;
  void promiseIntersection.then(() => {});
  promiseIntersection.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  promiseIntersection;
  await promiseIntersection.then(() => {});
  promiseIntersection.catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  promiseIntersection;
  promiseIntersection.then(() => {});
  void promiseIntersection.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;

async function test() {
  promiseIntersection;
  promiseIntersection.then(() => {});
  await promiseIntersection.catch();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  canThen.then(() => {});
  canThen.catch();
  canThen.finally();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  void canThen;
  canThen.then(() => {});
  canThen.catch();
  canThen.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  await canThen;
  canThen.then(() => {});
  canThen.catch();
  canThen.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  void canThen.then(() => {});
  canThen.catch();
  canThen.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  await canThen.then(() => {});
  canThen.catch();
  canThen.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  canThen.then(() => {});
  void canThen.catch();
  canThen.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  canThen.then(() => {});
  await canThen.catch();
  canThen.finally();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  canThen.then(() => {});
  canThen.catch();
  void canThen.finally();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CanThen extends Promise<number> {}
  const canThen: CanThen = Foo.resolve(2);

  canThen;
  canThen.then(() => {});
  canThen.catch();
  await canThen.finally();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  thenable;
  thenable.then(() => {});
}
      `,
			Options: NoFloatingPromisesOptions{CheckThenables: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  void thenable;
  thenable.then(() => {});
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  await thenable;
  thenable.then(() => {});
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  thenable;
  void thenable.then(() => {});
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function test() {
  class CatchableThenable {
    then(callback: () => void, callback: () => void): CatchableThenable {
      return new CatchableThenable();
    }
  }
  const thenable = new CatchableThenable();

  thenable;
  await thenable.then(() => {});
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  promise;
  promise.then(() => {});
  promise.catch();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  void promise;
  promise.then(() => {});
  promise.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  await promise;
  promise.then(() => {});
  promise.catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  promise;
  void promise.then(() => {});
  promise.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  promise;
  await promise.then(() => {});
  promise.catch();
}
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  promise;
  promise.then(() => {});
  void promise.catch();
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts
// Type definitions for promise-polyfill 6.0
// Project: https://github.com/taylorhakes/promise-polyfill
// Definitions by: Steve Jenkins <https://github.com/skysteve>
//                 Daniel Cassidy <https://github.com/djcsdy>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

interface PromisePolyfillConstructor extends PromiseConstructor {
  _immediateFn?: (handler: (() => void) | string) => void;
}

declare const PromisePolyfill: PromisePolyfillConstructor;

async function test() {
  const promise = new PromisePolyfill(() => {});

  promise;
  promise.then(() => {});
  await promise.catch();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        (async () => {
          await something();
        })();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        void (async () => {
          await something();
        })();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        await (async () => {
          await something();
        })();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        (async () => {
          something();
        })();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        void (async () => {
          something();
        })();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        await (async () => {
          something();
        })();
      `,
						},
					},
				},
			},
		},
		{
			Code: "(async function foo() {})();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output:    "void (async function foo() {})();",
						},
						{
							MessageId: "floatingFixAwait",
							Output:    "await (async function foo() {})();",
						},
					},
				},
			},
		},
		{
			Code: `
        function foo() {
          (async function bar() {})();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        function foo() {
          void (async function bar() {})();
        }
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        function foo() {
          await (async function bar() {})();
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        const foo = () =>
          new Promise(res => {
            (async function () {
              await res(1);
            })();
          });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        const foo = () =>
          new Promise(res => {
            void (async function () {
              await res(1);
            })();
          });
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        const foo = () =>
          new Promise(res => {
            await (async function () {
              await res(1);
            })();
          });
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        (async function () {
          await res(1);
        })();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        void (async function () {
          await res(1);
        })();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        await (async function () {
          await res(1);
        })();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        (async function () {
          Promise.resolve();
        })();
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        (async function () {
          void Promise.resolve();
        })();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        (async function () {
          await Promise.resolve();
        })();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
			Options: NoFloatingPromisesOptions{IgnoreIIFE: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  void promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  await promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  void promiseIntersection.then(() => {});
  promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  await promiseIntersection.then(() => {});
  promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  promiseIntersection.then(() => {});
  void promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  promiseIntersection.then(() => {});
  await promiseIntersection.catch();
  promiseIntersection.finally();
})();
      `,
						},
					},
				},
				{
					MessageId: "floatingVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
  void promiseIntersection.finally();
})();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const promiseIntersection: Promise<number> & number;
(async function () {
  promiseIntersection;
  promiseIntersection.then(() => {});
  promiseIntersection.catch();
  await promiseIntersection.finally();
})();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  void condition || myPromise();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  void (void condition || myPromise());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  await (void condition || myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  (await condition) && myPromise();
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  await ((await condition) && myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  condition && myPromise();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  void (condition && myPromise());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = true;

  await (condition && myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  condition || myPromise();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  void (condition || myPromise());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  await (condition || myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = null;

  condition ?? myPromise();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = null;

  void (condition ?? myPromise());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = null;

  await (condition ?? myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = true;
  condition && myPromise;
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = true;
  await (condition && myPromise);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = false;
  condition || myPromise;
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = false;
  await (condition || myPromise);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = null;
  condition ?? myPromise;
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = Promise.resolve(true);
  let condition = null;
  await (condition ?? myPromise);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  condition || condition || myPromise();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  void (condition || condition || myPromise());
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
async function foo() {
  const myPromise = async () => void 0;
  const condition = false;

  await (condition || condition || myPromise());
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
void Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
await Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
void Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
await Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
void Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
await Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
void Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
await Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

void Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

await Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
void Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
await Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
void Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
await Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandlerVoid",
					Line:      13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
void Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
await Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject() || 3;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (Promise.reject() || 3);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (Promise.reject() || 3);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
void Promise.resolve().then(() => {}, undefined);
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.resolve().then(() => {}, undefined);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const maybeCallable: string | (() => void);
Promise.resolve().then(() => {}, maybeCallable);
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
await Promise.resolve().then(() => {}, maybeCallable);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
await Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
await Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
await Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
await Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

await Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
await Promise.resolve().catch(null);
Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
await Promise.resolve().catch(3);
Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
				{
					MessageId: "floatingUselessRejectionHandler",
					Line:      13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const maybeCallable: string | (() => void);
declare const definitelyCallable: () => void;
Promise.resolve().then(() => {}, undefined);
Promise.resolve().then(() => {}, null);
Promise.resolve().then(() => {}, 3);
Promise.resolve().then(() => {}, maybeCallable);
Promise.resolve().then(() => {}, definitelyCallable);

Promise.resolve().catch(undefined);
Promise.resolve().catch(null);
Promise.resolve().catch(3);
await Promise.resolve().catch(maybeCallable);
Promise.resolve().catch(definitelyCallable);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject() || 3;
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
await (Promise.reject() || 3);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject().finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void Promise.reject().finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.reject().finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject()
  .finally(() => {})
  .finally(() => {});
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floating",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.reject()
  .finally(() => {})
  .finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject()
  .finally(() => {})
  .finally(() => {})
  .finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void Promise.reject()
  .finally(() => {})
  .finally(() => {})
  .finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.reject()
  .finally(() => {})
  .finally(() => {})
  .finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject()
  .then(() => {})
  .finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void Promise.reject()
  .then(() => {})
  .finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.reject()
  .then(() => {})
  .finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const returnsPromise: () => Promise<void> | null;
returnsPromise()?.finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const returnsPromise: () => Promise<void> | null;
void returnsPromise()?.finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const returnsPromise: () => Promise<void> | null;
await returnsPromise()?.finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const promiseIntersection: Promise<number> & number;
promiseIntersection.finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const promiseIntersection: Promise<number> & number;
void promiseIntersection.finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const promiseIntersection: Promise<number> & number;
await promiseIntersection.finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.resolve().finally(() => {}), 123;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (Promise.resolve().finally(() => {}), 123);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (Promise.resolve().finally(() => {}), 123);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
(async () => true)().finally();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (async () => true)().finally();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (async () => true)().finally();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
Promise.reject(new Error('message')).finally(() => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void Promise.reject(new Error('message')).finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await Promise.reject(new Error('message')).finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
function _<T, S extends Array<T | Promise<T>>>(
  maybePromiseArray: S | undefined,
): void {
  maybePromiseArray?.[0];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
function _<T, S extends Array<T | Promise<T>>>(
  maybePromiseArray: S | undefined,
): void {
  void maybePromiseArray?.[0];
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
function _<T, S extends Array<T | Promise<T>>>(
  maybePromiseArray: S | undefined,
): void {
  await maybePromiseArray?.[0];
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
[1, 2, 3].map(() => Promise.reject());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      2,
				},
			},
		},
		{
			Code: `
declare const array: unknown[];
array.map(() => Promise.reject());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare const promiseArray: Array<Promise<unknown>>;
void promiseArray;
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArray",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare const promiseArray: Array<Promise<unknown>>;
async function f() {
  await promiseArray;
}
      `,
			Options: NoFloatingPromisesOptions{IgnoreVoid: utils.Ref(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArray",
					Line:      4,
				},
			},
		},
		{
			Code: `
[1, 2, Promise.reject(), 3];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      2,
				},
			},
		},
		{
			Code: `
[1, 2, Promise.reject().catch(() => {}), 3];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      2,
				},
			},
		},
		{
			Code: `
const data = ['test'];
data.map(async () => {
  await new Promise((_res, rej) => setTimeout(rej, 1000));
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
function _<T, S extends Array<T | Array<T | Promise<T>>>>(
  maybePromiseArrayArray: S | undefined,
): void {
  maybePromiseArrayArray?.[0];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      5,
				},
			},
		},
		{
			Code: `
function f<T extends Array<Promise<number>>>(a: T): void {
  a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare const a: Array<Promise<number>> | undefined;
a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
function f<T extends Array<Promise<number>>>(a: T | undefined): void {
  a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
[Promise.reject()] as const;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      2,
				},
			},
		},
		{
			Code: `
declare function cursed(): [Promise<number>, Promise<string>];
cursed();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
[
  'Type Argument number ',
  1,
  'is not',
  Promise.resolve(),
  'but it still is flagged',
] as const;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      2,
				},
			},
		},
		{
			Code: `
        declare const arrayOrPromiseTuple:
          | Array<number>
          | [number, number, Promise<unknown>, string];
        arrayOrPromiseTuple;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      5,
				},
			},
		},
		{
			Code: `
        declare const okArrayOrPromiseArray: Array<number> | Array<Promise<unknown>>;
        okArrayOrPromiseArray;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      3,
				},
			},
		},
		{
			Code: `
interface UnsafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)
      | undefined
      | null,
  ): UnsafeThenable<TResult1 | TResult2>;
}
let promise: UnsafeThenable<number> = Promise.resolve(5);
promise;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}},
				CheckThenables:            utils.Ref(true),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
interface UnsafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)
      | undefined
      | null,
  ): UnsafeThenable<TResult1 | TResult2>;
}
let promise: UnsafeThenable<number> = Promise.resolve(5);
void promise;
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
interface UnsafeThenable<T> {
  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)
      | undefined
      | null,
  ): UnsafeThenable<TResult1 | TResult2>;
}
let promise: UnsafeThenable<number> = Promise.resolve(5);
await promise;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class SafePromise<T> extends Promise<T> {}
let promise: SafePromise<number> = Promise.resolve(5);
promise.catch();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
class SafePromise<T> extends Promise<T> {}
let promise: SafePromise<number> = Promise.resolve(5);
void promise.catch();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
class SafePromise<T> extends Promise<T> {}
let promise: SafePromise<number> = Promise.resolve(5);
await promise.catch();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class UnsafePromise<T> extends Promise<T> {}
let promise: () => UnsafePromise<number> = async () => 5;
promise().finally();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
class UnsafePromise<T> extends Promise<T> {}
let promise: () => UnsafePromise<number> = async () => 5;
void promise().finally();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
class UnsafePromise<T> extends Promise<T> {}
let promise: () => UnsafePromise<number> = async () => 5;
await promise().finally();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: UnsafePromise = Promise.resolve(5);
0 ? promise.catch() : 2;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: UnsafePromise = Promise.resolve(5);
void (0 ? promise.catch() : 2);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: UnsafePromise = Promise.resolve(5);
await (0 ? promise.catch() : 2);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: () => UnsafePromise = async () => 5;
null ?? promise().catch();
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: () => UnsafePromise = async () => 5;
void (null ?? promise().catch());
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
type UnsafePromise = Promise<number> & { hey?: string };
let promise: () => UnsafePromise = async () => 5;
await (null ?? promise().catch());
      `,
						},
					},
				},
			},
		},
		{
			Code: `
type Foo<T> = Promise<T> & { hey?: string };
declare const arrayOrPromiseTuple: Foo<unknown>[];
arrayOrPromiseTuple;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Bar"}}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      4,
				},
			},
		},
		{
			Code: `
type SafePromise = Promise<number> & { hey?: string };
let foo: SafePromise = Promise.resolve(1);
let bar = [Promise.resolve(2), foo];
bar;
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      5,
				},
			},
		},
		{
			Code: `
type Foo<T> = Promise<T> & { hey?: string };
declare const arrayOrPromiseTuple: [Foo<unknown>, 5];
arrayOrPromiseTuple;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Bar"}}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingPromiseArrayVoid",
					Line:      4,
				},
			},
		},
		{
			Code: `
type SafePromise = Promise<number> & { __linterBrands?: string };
declare const myTag: (strings: TemplateStringsArray) => SafePromise;
myTag` + "`" + `abc` + "`" + `;
      `,
			Options: NoFloatingPromisesOptions{AllowForKnownSafePromises: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
type SafePromise = Promise<number> & { __linterBrands?: string };
declare const myTag: (strings: TemplateStringsArray) => SafePromise;
void myTag` + "`" + `abc` + "`" + `;
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
type SafePromise = Promise<number> & { __linterBrands?: string };
declare const myTag: (strings: TemplateStringsArray) => SafePromise;
await myTag` + "`" + `abc` + "`" + `;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function unsafe(...args: unknown[]): Promise<void>;

        unsafe('...', () => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        declare function unsafe(...args: unknown[]): Promise<void>;

        void unsafe('...', () => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        declare function unsafe(...args: unknown[]): Promise<void>;

        await unsafe('...', () => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function it(...args: unknown[]): Promise<void>;

        it('...', () => {}).then(() => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        declare function it(...args: unknown[]): Promise<void>;

        void it('...', () => {}).then(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        declare function it(...args: unknown[]): Promise<void>;

        await it('...', () => {}).then(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function it(...args: unknown[]): Promise<void>;

        it('...', () => {}).finally(() => {});
      `,
			Options: NoFloatingPromisesOptions{
				AllowForKnownSafeCalls: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
        declare function it(...args: unknown[]): Promise<void>;

        void it('...', () => {}).finally(() => {});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
        declare function it(...args: unknown[]): Promise<void>;

        await it('...', () => {}).finally(() => {});
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const createPromise: () => PromiseLike<number>;
createPromise();
      `,
			Options: NoFloatingPromisesOptions{CheckThenables: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const createPromise: () => PromiseLike<number>;
void createPromise();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const createPromise: () => PromiseLike<number>;
await createPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

createMyThenable();
      `,
			Options: NoFloatingPromisesOptions{CheckThenables: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

void createMyThenable();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

await createMyThenable();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const createPromise: () => Promise<number>;
createPromise();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const createPromise: () => Promise<number>;
void createPromise();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const createPromise: () => Promise<number>;
await createPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class MyPromise<T> extends Promise<T> {}
declare const createMyPromise: () => MyPromise<number>;
createMyPromise();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
class MyPromise<T> extends Promise<T> {}
declare const createMyPromise: () => MyPromise<number>;
void createMyPromise();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
class MyPromise<T> extends Promise<T> {}
declare const createMyPromise: () => MyPromise<number>;
await createMyPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class MyPromise<T> extends Promise<T> {
  additional: string;
}
declare const createMyPromise: () => MyPromise<number>;
createMyPromise();
      `,
			Options: NoFloatingPromisesOptions{CheckThenables: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
class MyPromise<T> extends Promise<T> {
  additional: string;
}
declare const createMyPromise: () => MyPromise<number>;
void createMyPromise();
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
class MyPromise<T> extends Promise<T> {
  additional: string;
}
declare const createMyPromise: () => MyPromise<number>;
await createMyPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const x: any;
function* generator(): Generator<number, void, Promise<number>> {
  yield x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const x: any;
function* generator(): Generator<number, void, Promise<number>> {
  void (yield x);
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const x: any;
function* generator(): Generator<number, void, Promise<number>> {
  await (yield x);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const x: Generator<number, Promise<number>, void>;
function* generator(): Generator<number, void, void> {
  yield* x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
declare const x: Generator<number, Promise<number>, void>;
function* generator(): Generator<number, void, void> {
  void (yield* x);
}
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
declare const x: Generator<number, Promise<number>, void>;
function* generator(): Generator<number, void, void> {
  await (yield* x);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const value = {};
value as Promise<number>;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
const value = {};
void (value as Promise<number>);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
const value = {};
await (value as Promise<number>);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
({}) as Promise<number> & number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (({}) as Promise<number> & number);
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (({}) as Promise<number> & number);
      `,
						},
					},
				},
			},
		},
		{
			Code: `
({}) as Promise<number> & { yolo?: string };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (({}) as Promise<number> & { yolo?: string });
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (({}) as Promise<number> & { yolo?: string });
      `,
						},
					},
				},
			},
		},
		{
			Code: `
<Promise<number>>{};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "floatingVoid",
					Line:      2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "floatingFixVoid",
							Output: `
void (<Promise<number>>{});
      `,
						},
						{
							MessageId: "floatingFixAwait",
							Output: `
await (<Promise<number>>{});
      `,
						},
					},
				},
			},
		},
	})
}
