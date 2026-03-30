package await_thenable

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestAwaitThenable(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &AwaitThenableRule, []rule_tester.ValidTestCase{
		{Code: `
async function test() {
  await Promise.resolve('value');
  await Promise.reject(new Error('message'));
}
    `},
		{Code: `
async function test() {
  await (async () => true)();
}
    `},
		{Code: `
async function test() {
  function returnsPromise() {
    return Promise.resolve('value');
  }
  await returnsPromise();
}
    `},
		{Code: `
async function test() {
  async function returnsPromiseAsync() {}
  await returnsPromiseAsync();
}
    `},
		{Code: `
async function test() {
  let anyValue: any;
  await anyValue;
}
    `},
		{Code: `
async function test() {
  let unknownValue: unknown;
  await unknownValue;
}
    `},
		{Code: `
async function test() {
  const numberPromise: Promise<number>;
  await numberPromise;
}
    `},
		{Code: `
async function test() {
  class Foo extends Promise<number> {}
  const foo: Foo = Foo.resolve(2);
  await foo;

  class Bar extends Foo {}
  const bar: Bar = Bar.resolve(2);
  await bar;
}
    `},
		{Code: `
async function test() {
  await (Math.random() > 0.5 ? numberPromise : 0);
  await (Math.random() > 0.5 ? foo : 0);
  await (Math.random() > 0.5 ? bar : 0);

  const intersectionPromise: Promise<number> & number;
  await intersectionPromise;
}
    `},
		{Code: `
async function test() {
  class Thenable {
    then(callback: () => {}) {}
  }
  const thenable = new Thenable();

  await thenable;
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
}
    `},
		{Code: `
// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/bluebird/index.d.ts
// Type definitions for bluebird 3.5
// Project: https://github.com/petkaantonov/bluebird
// Definitions by: Leonard Hecker <https://github.com/lhecker>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped
// TypeScript Version: 2.8

/*!
 * The code following this comment originates from:
 *   https://github.com/types/npm-bluebird
 *
 * Note for browser users: use bluebird-global typings instead of this one
 * if you want to use Bluebird via the global Promise symbol.
 *
 * Licensed under:
 *   The MIT License (MIT)
 *
 *   Copyright (c) 2016 unional
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in
 *   all copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *   THE SOFTWARE.
 */

type Constructor<E> = new (...args: any[]) => E;
type CatchFilter<E> = ((error: E) => boolean) | (object & E);
type IterableItem<R> = R extends Iterable<infer U> ? U : never;
type IterableOrNever<R> = Extract<R, Iterable<any>>;
type Resolvable<R> = R | PromiseLike<R>;
type IterateFunction<T, R> = (
  item: T,
  index: number,
  arrayLength: number,
) => Resolvable<R>;

declare class Bluebird<R> implements PromiseLike<R> {
  then<U>(
    onFulfill?: (value: R) => Resolvable<U>,
    onReject?: (error: any) => Resolvable<U>,
  ): Bluebird<U>; // For simpler signature help.
  then<TResult1 = R, TResult2 = never>(
    onfulfilled?: ((value: R) => Resolvable<TResult1>) | null,
    onrejected?: ((reason: any) => Resolvable<TResult2>) | null,
  ): Bluebird<TResult1 | TResult2>;
}

declare const bluebird: Bluebird;

async function test() {
  await bluebird;
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

  await callback?.();
};
    `},
		{Code: `
async function* asyncYieldNumbers() {
  yield 1;
  yield 2;
  yield 3;
}
for await (const value of asyncYieldNumbers()) {
  console.log(value);
}
      `},
		{Code: `
declare const anee: any;
async function forAwait() {
  for await (const value of anee) {
    console.log(value);
  }
}
      `},
		{Code: `
declare const asyncIter: AsyncIterable<string> | Iterable<string>;
for await (const s of asyncIter) {
}
      `},
		{Code: `
declare const d: AsyncDisposable;

await using foo = d;

export {};
      `},
		{Code: `
using foo = {
  [Symbol.dispose]() {},
};

export {};
      `},
		{Code: `
await using foo = 3 as any;

export {};
      `},
		{Code: `
using foo = {
  async [Symbol.dispose]() {},
};

export {};
      `},
		{Code: `
declare const maybeAsyncDisposable: Disposable | AsyncDisposable;
async function foo() {
  await using _ = maybeAsyncDisposable;
}
      `},
		{Code: `
async function iterateUsing(arr: Array<AsyncDisposable>) {
  for (await using foo of arr) {
  }
}
      `},
		{Code: `
async function wrapper<T>(value: T) {
  return await value;
}
      `},
		{Code: `
async function wrapper<T extends unknown>(value: T) {
  return await value;
}
      `},
		{Code: `
async function wrapper<T extends any>(value: T) {
  return await value;
}
      `},
		{Code: `
async function wrapper<T extends Promise<unknown>>(value: T) {
  return await value;
}
      `},
		{Code: `
async function wrapper<T extends number | Promise<unknown>>(value: T) {
  return await value;
}
      `},
		{Code: `
class C<T> {
  async wrapper<T>(value: T) {
    return await value;
  }
}
      `},
		{Code: `
class C<R> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `},
		{Code: `
class C<R extends unknown> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `},
		{Code: `
// @ts-expect-error
Promise.all();
      `},
		{Code: `
Promise.all([,]);
      `},
		{Code: `
declare const x: unknown;

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
declare const x: any;
Promise.all(x);
      `},
		{Code: `
declare const x: Array<Promise<unknown>>;
Promise.all(x);
      `},
		{Code: `
declare const x: Array<Promise<number>> | Array<Promise<string>>;
Promise.all(x);
      `},
		{Code: `
declare const x: Array<Promise<number> | Promise<string>>;
Promise.all(x);
      `},
		{Code: `
function f<T>(x: Array<Promise<T>>) {
  Promise.all(x);
}
      `},
		{Code: `
function f<T extends Promise<unknown>>(x: Array<T>) {
  Promise.all(x);
}
      `},
		{Code: `
declare const x: Array<unknown>;
Promise.all(x);
      `},
		{Code: `
declare const x: Array<any>;
Promise.all(x);
      `},
		{Code: `
declare const x: number | Array<Promise<number>>;

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
declare const x: [Promise<unknown>, Promise<void>];
Promise.all(x);
      `},
		{Code: `
declare const x: [Promise<number>] | [Promise<string>];
Promise.all(x);
      `},
		{Code: `
declare const x: [Promise<number> | Promise<string>];
Promise.all(x);
      `},
		{Code: `
function f<T>(x: [Promise<T>]) {
  Promise.all(x);
}
      `},
		{Code: `
function f<T extends Promise<unknown>>(x: [T]) {
  Promise.all(x);
}
      `},
		{Code: `
declare const x: [unknown, any];
Promise.all(x);
      `},
		{Code: `
declare const x: number | [Promise<number>];

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
declare const x: Iterable<Promise<unknown>>;
Promise.all(x);
      `},
		{Code: `
declare const x: Iterable<Promise<number>> | Iterable<Promise<string>>;

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
declare const x: Iterable<Promise<number | string>>;
Promise.all(x);
      `},
		{Code: `
function f<T>(x: Iterable<Promise<T>>) {
  Promise.all(x);
}
      `},
		{Code: `
function f<T extends Promise<unknown>>(x: Iterable<T>) {
  Promise.all(x);
}
      `},
		{Code: `
declare const x: Iterable<unknown>;
Promise.all(x);
      `},
		{Code: `
declare const x: Iterable<any>;
Promise.all(x);
      `},
		{Code: `
declare const x: number | Iterable<Promise<number>>;

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
declare const x: Iterable<Promise<unknown>, number>;
Promise.all(x);
      `},
		{Code: `
declare class Foo<A, B> implements Iterable<B> {
  [Symbol.iterator](): Iterator<B>;
}
declare const x: Foo<Promise<string>, number>;

Promise.all(x);
      `},
		{Code: `
declare class Foo<A, B> implements Iterable<B> {
  [Symbol.iterator](): Iterator<B>;
}
declare const x: Foo<number, Promise<string>>;

Promise.all([...x]);
      `},
		{Code: `
declare const x: number[] | Promise<number>[];

Promise.all([...x]);
      `},
		{Code: `
declare const x: Iterable<Promise<string>> | Array<Promise<string>>;
Promise.all(x);
      `},
		{Code: `
declare const x:
  | Iterable<Promise<string>>
  | [Promise<string>, Promise<unknown>];
Promise.all(x);
      `},
		{Code: `
declare const x: Array<Promise<string>> | [Promise<string>, Promise<unknown>];
Promise.all(x);
      `},
		{Code: `
// @ts-expect-error
Promise.all(1);
      `},
		{Code: `
declare const x: Promise<number>;

// @ts-expect-error
Promise.all(x);
      `},
		{Code: `
interface MyArray<Unused, T> extends Array<T> {}
declare const x: MyArray<null, Promise<void>>;

Promise.all(x);
      `},
		{Code: `
function* x() {
  yield Promise.resolve(1);
  yield Promise.resolve(2);
  yield Promise.resolve(3);
}

Promise.all(x());
      `},
		{Code: `
function* x() {
  yield 1 as unknown;
}

Promise.all(x());
      `},
		{Code: `
function* x() {
  yield 1 as any;
}

Promise.all(x());
      `},
		{Code: `
declare const x: Generator<Promise<number>>;
Promise.all(x);
      `},
		{Code: `
declare const x: Generator<unknown>;
Promise.all(x);
      `},
		{Code: `
declare const x: Generator<any>;
Promise.all(x);
      `},
		{Code: `
declare const x: ReadonlyArray<Promise<number>>;
Promise.all(x);
      `},
		{Code: `
declare const x: ReadonlyArray<unknown>;
Promise.all(x);
      `},
		{Code: `
declare const x: ReadonlyArray<any>;
Promise.all(x);
      `},
		{Code: `
Promise.all([Promise.resolve(1), Promise.resolve(2), Promise.resolve(3)]);
      `},
		{Code: `
declare const _unknown_: unknown;

Promise.all([
  _unknown_,
  Promise.resolve(1),
  Promise.resolve(2),
  Promise.resolve(3),
]);
      `},
		{Code: `
declare const _any_: any;

Promise.all([
  _any_,
  Promise.resolve(1),
  Promise.resolve(2),
  Promise.resolve(3),
]);
      `},
		{Code: `
declare const _promise_: Promise<number | string>;

Promise.all([
  _promise_,
  Promise.resolve(1),
  Promise.resolve(2),
  Promise.resolve(3),
]);
      `},
		{Code: `
Promise.all([
  Promise.resolve(1),
  Promise.resolve(2),
  Promise.resolve(3),
  ...[Promise.resolve(4), Promise.resolve(5), Promise.resolve(6)],
]);
      `},
		{Code: `
declare const maybePromise: Promise<number> | number;

Promise.all([
  maybePromise,
  Promise.resolve(1),
  Promise.resolve(2),
  Promise.resolve(3),
]);
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "await 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output:    " 0;",
						},
					},
				},
			},
		},
		{
			Code: "await 'value';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output:    " 'value';",
						},
					},
				},
			},
		},
		{
			Code: "async () => await (Math.random() > 0.5 ? '' : 0);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output:    "async () =>  (Math.random() > 0.5 ? '' : 0);",
						},
					},
				},
			},
		},
		{
			Code: "async () => await(Math.random() > 0.5 ? '' : 0);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output:    "async () => (Math.random() > 0.5 ? '' : 0);",
						},
					},
				},
			},
		},
		{
			Code: `
class NonPromise extends Array {}
await new NonPromise();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
class NonPromise extends Array {}
 new NonPromise();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
async function test() {
  class IncorrectThenable {
    then() {}
  }
  const thenable = new IncorrectThenable();

  await thenable;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
async function test() {
  class IncorrectThenable {
    then() {}
  }
  const thenable = new IncorrectThenable();

   thenable;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const callback: (() => void) | undefined;
await callback?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
declare const callback: (() => void) | undefined;
 callback?.();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const obj: { a?: { b?: () => void } };
await obj.a?.b?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
declare const obj: { a?: { b?: () => void } };
 obj.a?.b?.();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const obj: { a: { b: { c?: () => void } } } | undefined;
await obj?.a.b.c?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
declare const obj: { a: { b: { c?: () => void } } } | undefined;
 obj?.a.b.c?.();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
function* yieldNumbers() {
  yield 1;
  yield 2;
  yield 3;
}
for await (const value of yieldNumbers()) {
  console.log(value);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forAwaitOfNonAsyncIterable",
					Line:      7,
					Column:    1,
					EndLine:   7,
					EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "convertToOrdinaryFor",
							Output: `
function* yieldNumbers() {
  yield 1;
  yield 2;
  yield 3;
}
for  (const value of yieldNumbers()) {
  console.log(value);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
function* yieldNumberPromises() {
  yield Promise.resolve(1);
  yield Promise.resolve(2);
  yield Promise.resolve(3);
}
for await (const value of yieldNumberPromises()) {
  console.log(value);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forAwaitOfNonAsyncIterable",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "convertToOrdinaryFor",
							Output: `
function* yieldNumberPromises() {
  yield Promise.resolve(1);
  yield Promise.resolve(2);
  yield Promise.resolve(3);
}
for  (const value of yieldNumberPromises()) {
  console.log(value);
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const disposable: Disposable;
async function foo() {
  await using d = disposable;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      4,
					Column:    19,
					EndLine:   4,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
declare const disposable: Disposable;
async function foo() {
   using d = disposable;
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
  await using _ = {
    async [Symbol.dispose]() {},
  };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      3,
					Column:    19,
					EndLine:   5,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
async function foo() {
   using _ = {
    async [Symbol.dispose]() {},
  };
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const disposable: Disposable;
declare const asyncDisposable: AsyncDisposable;
async function foo() {
  await using a = disposable,
    b = asyncDisposable,
    c = disposable,
    d = asyncDisposable,
    e = disposable;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      5,
					Column:    19,
					EndLine:   5,
					EndColumn: 29,
				},
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      7,
					Column:    9,
					EndLine:   7,
					EndColumn: 19,
				},
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      9,
					Column:    9,
					EndLine:   9,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
declare const anee: any;
declare const disposable: Disposable;
async function foo() {
  await using a = anee,
    b = disposable;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "awaitUsingOfNonAsyncDisposable",
					Line:      6,
					Column:    9,
					EndLine:   6,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
async function wrapper<T extends number>(value: T) {
  return await value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
async function wrapper<T extends number>(value: T) {
  return  value;
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class C<T> {
  async wrapper<T extends string>(value: T) {
    return await value;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
class C<T> {
  async wrapper<T extends string>(value: T) {
    return  value;
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return await value;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "await",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeAwait",
							Output: `
class C<R extends number> {
  async wrapper<T extends R>(value: T) {
    return  value;
  }
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare const x: Array<number>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Array<number> | Array<Promise<number>>;
Promise.race(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Array<number> | Array<string>;
Promise.allSettled(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Array<number | Promise<number>>;
Promise.any(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: [number];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: [number] | [Promise<number>];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: [number | Promise<number>];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: [Promise<number>, number];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Iterable<number>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Iterable<number> | Iterable<Promise<number>>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Iterable<number | Promise<number>>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Iterable<string> | Array<Promise<unknown>>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Iterable<Promise<string>> | [string, Promise<unknown>];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Array<string> | [Promise<string>, Promise<unknown>];
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Array<Array<Promise<number>>>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
interface MyArray<Unused, T> extends Array<T> {}
declare const x: MyArray<Promise<void>, null>;

Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare class Foo<A, B> implements Iterable<B> {
  [Symbol.iterator](): Iterator<B>;
}
declare const x: Foo<number, Promise<string>>;

Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
function* x() {
  yield 1;
  yield 2;
  yield 3;
}

Promise.all(x());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
function* x() {
  yield 1 as number;
}

Promise.all(x());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
function* x() {
  yield 1 as number | Promise<number>;
}

Promise.all(x());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: Generator<number>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: ReadonlyArray<number>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
declare const x: ReadonlyArray<number | Promise<string>>;
Promise.all(x);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
				},
			},
		},
		{
			Code: `
Promise.all([Promise.resolve(1), 2, Promise.resolve(3)]);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
					Line:      2,
					Column:    34,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
Promise.all([1, 2, Promise.resolve(3)]);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
					Line:      2,
					Column:    14,
					EndColumn: 15,
				},
				{
					MessageId: "invalidPromiseAggregatorInput",
					Line:      2,
					Column:    17,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
Promise.all([...[1, 2, 3]]);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPromiseAggregatorInput",
					Line:      2,
					Column:    14,
					EndColumn: 26,
				},
			},
		},
	})
}
