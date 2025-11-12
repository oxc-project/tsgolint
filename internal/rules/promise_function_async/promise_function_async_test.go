package promise_function_async

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPromiseFunctionAsyncRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PromiseFunctionAsyncRule, []rule_tester.ValidTestCase{
		{Code: `
const nonAsyncNonPromiseArrowFunction = (n: number) => n;
    `},
		{Code: `
function nonAsyncNonPromiseFunctionDeclaration(n: number) {
  return n;
}
    `},
		{Code: `
const asyncPromiseFunctionExpressionA = async function (p: Promise<void>) {
  return p;
};
    `},
		{Code: `
const asyncPromiseFunctionExpressionB = async function () {
  return new Promise<void>();
};
    `},
		{Code: `
class Test {
  public nonAsyncNonPromiseArrowFunction = (n: number) => n;
  public nonAsyncNonPromiseMethod() {
    return 0;
  }

  public async asyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public async asyncPromiseMethodB() {
    return new Promise<void>();
  }
}
    `},
		{Code: `
class InvalidAsyncModifiers {
  public constructor() {
    return new Promise<void>();
  }
  public get asyncGetter() {
    return new Promise<void>();
  }
  public set asyncGetter(p: Promise<void>) {
    return p;
  }
  public get asyncGetterFunc() {
    return async () => new Promise<void>();
  }
  public set asyncGetterFunc(p: () => Promise<void>) {
    return p;
  }
}
    `},
		{Code: `
const invalidAsyncModifiers = {
  get asyncGetter() {
    return new Promise<void>();
  },
  set asyncGetter(p: Promise<void>) {
    return p;
  },
  get asyncGetterFunc() {
    return async () => new Promise<void>();
  },
  set asyncGetterFunc(p: () => Promise<void>) {
    return p;
  },
};
    `},
		{Code: `
      export function valid(n: number) {
        return n;
      }
    `},
		{Code: `
      export default function invalid(n: number) {
        return n;
      }
    `},
		{Code: `
      class Foo {
        constructor() {}
      }
    `},
		{Code: `
class Foo {
  async catch<T>(arg: Promise<T>) {
    return arg;
  }
}
    `},
		{
			Code: `
function returnsAny(): any {
  return 0;
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": true}`),
		},
		{
			Code: `
function returnsUnknown(): unknown {
  return 0;
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": true}`),
		},
		{
			Code: `
interface ReadableStream {}
interface Options {
  stream: ReadableStream;
}

type Return = ReadableStream | Promise<void>;
const foo = (options: Options): Return => {
  return options.stream ? asStream(options) : asPromise(options);
};
      `,
		},
		{
			Code: `
function foo(): Promise<string> | boolean {
  return Math.random() > 0.5 ? Promise.resolve('value') : false;
}
      `,
		},
		{
			Code: `
abstract class Test {
  abstract test1(): Promise<number>;

  // abstract method with body is always an error but it still parses into valid AST
  abstract test2(): Promise<number> {
    return Promise.resolve(1);
  }
}
      `,
		},
		{Code: `
function promiseInUnionWithExplicitReturnType(
  p: boolean,
): Promise<number> | number {
  return p ? Promise.resolve(5) : 5;
}
    `},
		{Code: `
function explicitReturnWithPromiseInUnion(): Promise<number> | number {
  return 5;
}
    `},
		{Code: `
async function asyncFunctionReturningUnion(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
    `},
		{Code: `
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): number;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number> | number {
  return Promise.resolve(5);
}
    `},
		{Code: `
function overloadingThatCanReturnPromise(a: boolean): number;
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number> | number {
  return Promise.resolve(5);
}
    `},
		{Code: `
function a(): Promise<void>;
function a(x: boolean): void;
function a(x?: boolean) {
  if (x == null) return Promise.reject(new Error());
  throw new Error();
}
    `},
		{
			Code: `
function overloadingThatIncludeUnknown(): number;
function overloadingThatIncludeUnknown(a: boolean): unknown;
function overloadingThatIncludeUnknown(a?: boolean): unknown | number {
  return Promise.resolve(5);
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": true}`),
		},
		{
			Code: `
function overloadingThatIncludeAny(): number;
function overloadingThatIncludeAny(a: boolean): any;
function overloadingThatIncludeAny(a?: boolean): any | number {
  return Promise.resolve(5);
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": true}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function returnsAny(): any {
  return 0;
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function returnsUnknown(): unknown {
  return 0;
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpressionA = function (p: Promise<void>) {
  return p;
};
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpressionA =  async function (p: Promise<void>) {
  return p;
};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpressionB = function () {
  return new Promise<void>();
};
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpressionB =  async function () {
  return new Promise<void>();
};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function nonAsyncPromiseFunctionDeclarationA(p: Promise<void>) {
  return p;
}
      `,
			Output: []string{`
 async function nonAsyncPromiseFunctionDeclarationA(p: Promise<void>) {
  return p;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function nonAsyncPromiseFunctionDeclarationB() {
  return new Promise<void>();
}
      `,
			Output: []string{`
 async function nonAsyncPromiseFunctionDeclarationB() {
  return new Promise<void>();
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseArrowFunctionA = (p: Promise<void>) => p;
      `,
			Output: []string{`
const nonAsyncPromiseArrowFunctionA =  async (p: Promise<void>) => p;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseArrowFunctionB = () => new Promise<void>();
      `,
			Output: []string{`
const nonAsyncPromiseArrowFunctionB =  async () => new Promise<void>();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
const functions = {
  nonAsyncPromiseMethod() {
    return Promise.resolve(1);
  },
};
      `,
			Output: []string{`
const functions = {
   async nonAsyncPromiseMethod() {
    return Promise.resolve(1);
  },
};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
				},
			},
		},
		{
			Code: `
class Test {
  public nonAsyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public static nonAsyncPromiseMethodB() {
    return new Promise<void>();
  }
}
      `,
			Output: []string{`
class Test {
  public  async nonAsyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public static  async nonAsyncPromiseMethodB() {
    return new Promise<void>();
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
				},
				{
					MessageId: "missingAsync",
					Line:      7,
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpression =  async function (p: Promise<void>) {
  return p;
};

 async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public  async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"checkArrowFunctions": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      2,
				},
				{
					MessageId: "missingAsync",
					Line:      6,
				},
				{
					MessageId: "missingAsync",
					Line:      13,
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpression =  async function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction =  async (p: Promise<void>) => p;

class Test {
  public  async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"checkFunctionDeclarations": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      2,
				},
				{
					MessageId: "missingAsync",
					Line:      10,
				},
				{
					MessageId: "missingAsync",
					Line:      13,
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

 async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction =  async (p: Promise<void>) => p;

class Test {
  public  async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"checkFunctionExpressions": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      6,
				},
				{
					MessageId: "missingAsync",
					Line:      10,
				},
				{
					MessageId: "missingAsync",
					Line:      13,
				},
			},
		},
		{
			Code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			Output: []string{`
const nonAsyncPromiseFunctionExpression =  async function (p: Promise<void>) {
  return p;
};

 async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction =  async (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"checkMethodDeclarations": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      2,
				},
				{
					MessageId: "missingAsync",
					Line:      6,
				},
				{
					MessageId: "missingAsync",
					Line:      10,
				},
			},
		},
		{
			Code: `
class PromiseType {}

const returnAllowedType = () => new PromiseType();
      `,
			Output: []string{`
class PromiseType {}

const returnAllowedType =  async () => new PromiseType();
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowedPromiseNames": ["PromiseType"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      4,
				},
			},
		},
		{
			Code: `
interface SPromise<T> extends Promise<T> {}
function foo(): Promise<string> | SPromise<boolean> {
  return Math.random() > 0.5
    ? Promise.resolve('value')
    : Promise.resolve(false);
}
      `,
			Output: []string{`
interface SPromise<T> extends Promise<T> {}
 async function foo(): Promise<string> | SPromise<boolean> {
  return Math.random() > 0.5
    ? Promise.resolve('value')
    : Promise.resolve(false);
}
      `,
			},
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowedPromiseNames": ["SPromise"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
				},
			},
		},
		{
			Code: `
class Test {
  @decorator
  public test() {
    return Promise.resolve(123);
  }
}
      `,
			Output: []string{`
class Test {
  @decorator
  public  async test() {
    return Promise.resolve(123);
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Test {
  @decorator(async () => {})
  static protected[(1)]() {
    return Promise.resolve(1);
  }
  public'bar'() {
    return Promise.resolve(2);
  }
  private['baz']() {
    return Promise.resolve(3);
  }
}
      `,
			Output: []string{`
class Test {
  @decorator(async () => {})
  static protected async [(1)]() {
    return Promise.resolve(1);
  }
  public async 'bar'() {
    return Promise.resolve(2);
  }
  private async ['baz']() {
    return Promise.resolve(3);
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
					Column:    3,
				},
				{
					MessageId: "missingAsync",
					Line:      7,
					Column:    3,
				},
				{
					MessageId: "missingAsync",
					Line:      10,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  catch() {
    return Promise.resolve(1);
  }

  public default() {
    return Promise.resolve(2);
  }

  @decorator
  private case<T>() {
    return Promise.resolve(3);
  }
}
      `,
			Output: []string{`
class Foo {
   async catch() {
    return Promise.resolve(1);
  }

  public  async default() {
    return Promise.resolve(2);
  }

  @decorator
  private  async case<T>() {
    return Promise.resolve(3);
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
					Column:    3,
				},
				{
					MessageId: "missingAsync",
					Line:      7,
					Column:    3,
				},
				{
					MessageId: "missingAsync",
					Line:      11,
					Column:    3,
				},
			},
		},
		{
			Code: `
const foo = {
  catch() {
    return Promise.resolve(1);
  },
};
      `,
			Output: []string{`
const foo = {
   async catch() {
    return Promise.resolve(1);
  },
};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
function promiseInUnionWithoutExplicitReturnType(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
      `,
			Output: []string{`
 async function promiseInUnionWithoutExplicitReturnType(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): Promise<string>;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number | string> {
  return Promise.resolve(5);
}
      `,
			Output: []string{`
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): Promise<string>;
 async function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number | string> {
  return Promise.resolve(5);
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function overloadingThatIncludeAny(): number;
function overloadingThatIncludeAny(a: boolean): any;
function overloadingThatIncludeAny(a?: boolean): any | number {
  return Promise.resolve(5);
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
		{
			Code: `
function overloadingThatIncludeUnknown(): number;
function overloadingThatIncludeUnknown(a: boolean): unknown;
function overloadingThatIncludeUnknown(a?: boolean): unknown | number {
  return Promise.resolve(5);
}
      `,
			Options: rule_tester.OptionsFromJSON[PromiseFunctionAsyncOptions](`{"allowAny": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAsync",
				},
			},
		},
	})
}
