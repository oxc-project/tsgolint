package prefer_promise_reject_errors

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferPromiseRejectErrors(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferPromiseRejectErrorsRule, []rule_tester.ValidTestCase{
		{Code: "Promise.resolve(5);"},
		{
			Code:    "Promise.reject();",
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowEmptyReject": true}`),
		},
		{
			Code: `
        declare const someAnyValue: any;
        Promise.reject(someAnyValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": true, "allowThrowingUnknown": false}`),
		},
		{
			Code: `
        declare const someUnknownValue: unknown;
        Promise.reject(someUnknownValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": false, "allowThrowingUnknown": true}`),
		},
		{Code: "Promise.reject(new Error());"},
		{Code: "Promise.reject(new TypeError());"},
		{Code: "Promise.reject(new Error('foo'));"},
		{Code: `
      class CustomError extends Error {}
      Promise.reject(new CustomError());
    `},
		{Code: `
      declare const foo: () => { err: SyntaxError };
      Promise.reject(foo().err);
    `},
		{Code: `
      declare const foo: () => Promise<Error>;
      Promise.reject(await foo());
    `},
		{Code: "Promise.reject((foo = new Error()));"},
		{Code: `
      const foo = Promise;
      foo.reject(new Error());
    `},
		{Code: "Promise['reject'](new Error());"},
		{Code: "Promise.reject(true && new Error());"},
		{Code: `
      const foo = false;
      Promise.reject(false || new Error());
    `},
		{Code: `
      declare const foo: Readonly<Error>;
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Error> | Readonly<TypeError>;
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Error> & Readonly<TypeError>;
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Error> & { foo: 'bar' };
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Error & { bar: 'foo' }> & { foo: 'bar' };
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Readonly<Error>>;
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<Readonly<Readonly<Error>>>;
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Readonly<
        Readonly<Readonly<Error & { bar: 'foo' }> & { foo: 'bar' }> & {
          fooBar: 'barFoo';
        }
      > & { barFoo: 'fooBar' };
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo:
        | Readonly<Readonly<Error> | Readonly<TypeError & string>>
        | Readonly<Error>;
      Promise.reject(foo);
    `},
		{Code: `
      type Wrapper<T> = { foo: Readonly<T>[] };
      declare const foo: Wrapper<Error>['foo'][5];
      Promise.reject(foo);
    `},
		{Code: `
      declare const foo: Error[];
      Promise.reject(foo[5]);
    `},
		{Code: `
      declare const foo: ReadonlyArray<Error>;
      Promise.reject(foo[5]);
    `},
		{Code: `
      declare const foo: [Error];
      Promise.reject(foo[0]);
    `},
		{Code: `
      new Promise(function (resolve, reject) {
        resolve(5);
      });
    `},
		{Code: `
      new Promise(function (resolve, reject) {
        reject(new Error());
      });
    `},
		{Code: `
      new Promise((resolve, reject) => {
        reject(new Error());
      });
			`},
		{Code: "new Promise((resolve, reject) => reject(new Error()));"},
		{
			Code: `
        new Promise(function (resolve, reject) {
          reject();
        });
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowEmptyReject": true}`),
		},
		{Code: "new Promise((yes, no) => no(new Error()));"},
		{Code: "new Promise();"},
		{Code: "new Promise(5);"},
		{Code: "new Promise((resolve, { apply }) => {});"},
		{Code: "new Promise((resolve, reject) => {});"},
		{Code: "new Promise((resolve, reject) => reject);"},
		{Code: `
      class CustomError extends Error {}
      new Promise(function (resolve, reject) {
        reject(new CustomError());
      });
    `},
		{Code: `
      declare const foo: () => { err: SyntaxError };
      new Promise(function (resolve, reject) {
        reject(foo().err);
      });
    `},
		{Code: "new Promise((resolve, reject) => reject((foo = new Error())));"},
		{Code: `
      new Foo((resolve, reject) => reject(5));
    `},
		{Code: `
      class Foo {
        constructor(
          executor: (resolve: () => void, reject: (reason?: any) => void) => void,
        ): Promise<any> {}
      }
      new Foo((resolve, reject) => reject(5));
    `},
		{Code: `
      new Promise((resolve, reject) => {
        return function (reject) {
          reject(5);
        };
      });
    `},
		{Code: "new Promise((resolve, reject) => resolve(5, reject));"},
		{Code: `
      class C {
        #error: Error;
        foo() {
          Promise.reject(this.#error);
        }
      }
    `},
		{Code: `
      const foo = Promise;
      new foo((resolve, reject) => reject(new Error()));
    `},
		{Code: `
      declare const foo: Readonly<Error>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Error> | Readonly<TypeError>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Error> & Readonly<TypeError>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Error> & { foo: 'bar' };
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Error & { bar: 'foo' }> & { foo: 'bar' };
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Readonly<Error>>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<Readonly<Readonly<Error>>>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Readonly<
        Readonly<Readonly<Error & { bar: 'foo' }> & { foo: 'bar' }> & {
          fooBar: 'barFoo';
        }
      > & { barFoo: 'fooBar' };
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo:
        | Readonly<Readonly<Error> | Readonly<TypeError & string>>
        | Readonly<Error>;
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      type Wrapper<T> = { foo: Readonly<T>[] };
      declare const foo: Wrapper<Error>['foo'][5];
      new Promise((resolve, reject) => reject(foo));
    `},
		{Code: `
      declare const foo: Error[];
      new Promise((resolve, reject) => reject(foo[5]));
    `},
		{Code: `
      declare const foo: ReadonlyArray<Error>;
      new Promise((resolve, reject) => reject(foo[5]));
    `},
		{Code: `
      declare const foo: [Error];
      new Promise((resolve, reject) => reject(foo[0]));
    `},
		{Code: `
      class Foo extends Promise<number> {}
      Foo.reject(new Error());
    `},
		{Code: `
      class Foo extends Promise<number> {}
      new Foo((resolve, reject) => reject(new Error()));
    `},
		{Code: `
      declare const someRandomCall: {
        reject(arg: any): void;
      };
      someRandomCall.reject(5);
    `},
		{Code: `
      declare const foo: PromiseConstructor;
      foo.reject(new Error());
    `},
		{Code: "console[Symbol.iterator]();"},
		{Code: `
      class A {
        a = [];
        [Symbol.iterator]() {
          return this.a[Symbol.iterator]();
        }
      }
    `},
		{Code: `
      declare const foo: PromiseConstructor;
      function fun<T extends Error>(t: T): void {
        foo.reject(t);
      }
    `},
		{
			Code: `
        declare const someAnyValue: any;
        Promise.reject(someAnyValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": true, "allowThrowingUnknown": true}`),
		},
		{
			Code: `
        declare const someUnknownValue: unknown;
        Promise.reject(someUnknownValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": true, "allowThrowingUnknown": true}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "Promise.reject(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 18,
				},
			},
		},
		{
			Code: "Promise.reject('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		{
			Code: "Promise.reject(`foo`);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		{
			Code: "Promise.reject('foo', somethingElse);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				},
			},
		},
		{
			Code: "Promise.reject(false);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		{
			Code: "Promise.reject(void `foo`);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 27,
				},
			},
		},
		{
			Code: "Promise.reject();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 17,
				},
			},
		},
		{
			Code: "Promise.reject(undefined);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 26,
				},
			},
		},
		{
			Code:    "Promise.reject(undefined);",
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowEmptyReject": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "Promise.reject(null);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "Promise.reject({ foo: 1 });",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 27,
				},
			},
		},
		{
			Code: "Promise.reject([1, 2, 3]);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
declare const foo: Error | undefined;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: () => Promise<string>;
Promise.reject(await foo());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
declare const foo: boolean;
Promise.reject(foo && new Error());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
const foo = Promise;
foo.reject();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 13,
				},
			},
		},
		{
			Code: "Promise.reject?.(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 20,
				},
			},
		},
		{
			Code: "Promise?.reject(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
				},
			},
		},
		{
			Code: "Promise?.reject?.(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "(Promise?.reject)(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "(Promise?.reject)?.(5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 23,
				},
			},
		},
		{
			Code: "Promise['reject'](5);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				},
			},
		},
		{
			Code: "Promise.reject((foo += new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				},
			},
		},
		{
			Code: "Promise.reject((foo -= new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				},
			},
		},
		{
			Code: "Promise.reject((foo **= new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},
		{
			Code: "Promise.reject((foo <<= new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},
		{
			Code: "Promise.reject((foo |= new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				},
			},
		},
		{
			Code: "Promise.reject((foo &= new Error()));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
declare const foo: never;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: unknown;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
type FakeReadonly<T> = { 'fake readonly': T };
declare const foo: FakeReadonly<Error>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<'error'>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error | 'error'>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error> | 'error';
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error> | Readonly<TypeError> | Readonly<'error'>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<'error'>>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Readonly<Error> | 'error'>>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Readonly<Error> & TypeError>> | 'error';
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Error>> | Readonly<TypeError> | 'error';
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
type Wrapper<T> = { foo: Readonly<T>[] };
declare const foo: Wrapper<Error | 'error'>['foo'][5];
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: Error[];
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: ReadonlyArray<Error>;
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare const foo: [Error];
Promise.reject(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
new Promise(function (resolve, reject) {
  reject();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: `
new Promise(function (resolve, reject) {
  reject(5);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
new Promise((resolve, reject) => {
  reject();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 11,
				},
			},
		},
		{
			Code: "new Promise((resolve, reject) => reject(5));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    34,
					EndLine:   1,
					EndColumn: 43,
				},
			},
		},
		{
			Code: `
new Promise((resolve, reject) => {
  fs.readFile('foo.txt', (err, file) => {
    if (err) reject('File not found');
    else resolve(file);
  });
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    14,
					EndLine:   4,
					EndColumn: 38,
				},
			},
		},
		{
			Code: "new Promise((yes, no) => no(5));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 31,
				},
			},
		},
		{
			Code: "new Promise(({ foo, bar, baz }, reject) => reject(5));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 53,
				},
			},
		},
		{
			Code: `
new Promise(function (reject, reject) {
  reject(5);
});
      `,
			// TODO(port): this is invalid TypeScript code
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
new Promise(function (foo, arguments) {
  arguments(5);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 15,
				},
			},
		},
		{
			Code: "new Promise((foo, arguments) => arguments(5));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    33,
					EndLine:   1,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
new Promise(function ({}, reject) {
  reject(5);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 12,
				},
			},
		},
		{
			Code: "new Promise(({}, reject) => reject(5));",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},
		{
			Code: "new Promise((resolve, reject, somethingElse = reject(5)) => {});",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      1,
					Column:    47,
					EndLine:   1,
					EndColumn: 56,
				},
			},
		},
		{
			Code: `
declare const foo: {
  bar: PromiseConstructor;
};
new foo.bar((resolve, reject) => reject(5));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      5,
					Column:    34,
					EndLine:   5,
					EndColumn: 43,
				},
			},
		},
		{
			Code: `
declare const foo: {
  bar: PromiseConstructor;
};
new (foo?.bar)((resolve, reject) => reject(5));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      5,
					Column:    37,
					EndLine:   5,
					EndColumn: 46,
				},
			},
		},
		{
			Code: `
const foo = Promise;
new foo((resolve, reject) => reject(5));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    30,
					EndLine:   3,
					EndColumn: 39,
				},
			},
		},
		{
			Code: `
declare const foo: never;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: unknown;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
type FakeReadonly<T> = { 'fake readonly': T };
declare const foo: FakeReadonly<Error>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    34,
					EndLine:   4,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<'error'>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error | 'error'>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error> | 'error';
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Error> | Readonly<TypeError> | Readonly<'error'>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<'error'>>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Readonly<Error> | 'error'>>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Readonly<Error> & TypeError>> | 'error';
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Readonly<Readonly<Error>> | Readonly<TypeError> | 'error';
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
type Wrapper<T> = { foo: Readonly<T>[] };
declare const foo: Wrapper<Error | 'error'>['foo'][5];
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    34,
					EndLine:   4,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: Error[];
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: ReadonlyArray<Error>;
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
declare const foo: [Error];
new Promise((resolve, reject) => reject(foo));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    34,
					EndLine:   3,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
class Foo extends Promise<number> {}
Foo.reject(5);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const foo: PromiseConstructor & string;
foo.reject(5);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
class Foo extends Promise<number> {}
class Bar extends Foo {}
Bar.reject(5);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
declare const foo: PromiseConstructor;
function fun<T extends number>(t: T): void {
  foo.reject(t);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
        declare const someAnyValue: any;
        Promise.reject(someAnyValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": false, "allowThrowingUnknown": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
				},
			},
		},
		{
			Code: `
        declare const someUnknownValue: unknown;
        Promise.reject(someUnknownValue);
      `,
			Options: rule_tester.OptionsFromJSON[PreferPromiseRejectErrorsOptions](`{"allowThrowingAny": true, "allowThrowingUnknown": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
				},
			},
		},
		{
			Code: `
        declare const someUnknownValue: unknown;
        Promise.reject(someUnknownValue);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
				},
			},
		},
		{
			Code: `
        declare const someAnyValue: any;
        Promise.reject(someAnyValue);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "rejectAnError",
				},
			},
		},
	})
}
