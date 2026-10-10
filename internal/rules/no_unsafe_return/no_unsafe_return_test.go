package no_unsafe_return

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeReturnRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noImplicitThis.json", t, &NoUnsafeReturnRule, []rule_tester.ValidTestCase{
		{Code: `
function foo() {
  return;
}
    `},
		{Code: `
function foo() {
  return 1;
}
    `},
		{Code: `
function foo() {
  return '';
}
    `},
		{Code: `
function foo() {
  return true;
}
    `},
		{Code: `
function foo() {
  return [];
}
    `},
		{Code: `
function foo(): any {
  return {} as any;
}
    `},
		{Code: `
declare function foo(arg: () => any): void;
foo((): any => 'foo' as any);
    `},
		{Code: `
declare function foo(arg: null | (() => any)): void;
foo((): any => 'foo' as any);
    `},
		{Code: `
function foo(): any[] {
  return [] as any[];
}
    `},
		{Code: `
function foo(): Set<any> {
  return new Set<any>();
}
    `},
		{Code: `
async function foo(): Promise<any> {
  return Promise.resolve({} as any);
}
    `},
		{Code: `
async function foo(): Promise<any> {
  return {} as any;
}
    `},
		{Code: `
function foo(): object {
  return Promise.resolve({} as any);
}
    `},
		{Code: `
function foo(): ReadonlySet<number> {
  return new Set<any>();
}
    `},
		{Code: `
function foo(): Set<number> {
  return new Set([1]);
}
    `},
		{Code: `
      type Foo<T = number> = { prop: T };
      function foo(): Foo {
        return { prop: 1 } as Foo<number>;
      }
    `},
		{Code: `
      type Foo = { prop: any };
      function foo(): Foo {
        return { prop: '' } as Foo;
      }
    `},
		{Code: `
      function fn<T extends any>(x: T) {
        return x;
      }
    `},
		{Code: `
      function fn<T extends any>(x: T): unknown {
        return x as any;
      }
    `},
		{Code: `
      function fn<T extends any>(x: T): unknown[] {
        return x as any[];
      }
    `},
		{Code: `
      function fn<T extends any>(x: T): Set<unknown> {
        return x as Set<any>;
      }
    `},
		{Code: `
      async function fn<T extends any>(x: T): Promise<unknown> {
        return x as any;
      }
    `},
		{Code: `
      function fn<T extends any>(x: T): Promise<unknown> {
        return Promise.resolve(x as any);
      }
    `},
		{Code: `
      function test(): Map<string, string> {
        return new Map();
      }
    `},
		{Code: `
      function foo(): any {
        return [] as any[];
      }
    `},
		{Code: `
      function foo(): unknown {
        return [] as any[];
      }
    `},
		{Code: `
      declare const value: Promise<any>;
      function foo() {
        return value;
      }
    `},
		{Code: "const foo: (() => void) | undefined = () => 1;"},
		{Code: `
      class Foo {
        public foo(): this {
          return this;
        }

        protected then(resolve: () => void): void {
          resolve();
        }
      }
    `},
		{Code: `
interface Schema<Output = any> {
  readonly _output: Output;
  parse(input: unknown): Output;
}
type Infer<S extends Schema> = S['_output'];
const parse = <S extends Schema>(schema: S, input: unknown): Infer<S> =>
  schema.parse(input) as Infer<S>;
    `},
		{Code: `
function f<T extends Set<any>>(x: T): Set<string> {
  return x;
}
    `},
		{
			FileName: "repro.js",
			TSConfig: "tsconfig.checkJs.json",
			Code: `
/** @callback Fn */

/** @type {Fn} */
const f = function () {
  return 1;
};`,
		},
		{
			FileName: "repro.js",
			TSConfig: "tsconfig.checkJs.json",
			Code: `
/** @callback Fn */

/** @type {Fn} */
const f = () => 1;`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function foo() {
  return 1 as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo() {
  return Object.create(null);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
const foo = () => {
  return 1 as any;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: "const foo = () => Object.create(null);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo() {
  return [] as any[];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo() {
  return [] as Array<any>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo() {
  return [] as readonly any[];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo() {
  return [] as Readonly<any[]>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
const foo = () => {
  return [] as any[];
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: "const foo = () => [] as any[];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
				},
			},
		},
		{
			Code: `
function foo(): Set<string> {
  return new Set<any>();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
				},
			},
		},
		{
			Code: `
function foo(): Map<string, string> {
  return new Map<string, any>();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
				},
			},
		},
		{
			Code: `
function foo(): Set<string[]> {
  return new Set<any[]>();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
				},
			},
		},
		{
			Code: `
function foo(): Set<Set<Set<string>>> {
  return new Set<Set<Set<any>>>();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
				},
			},
		},
		{
			Code: `
type Fn = () => Set<string>;
const foo1: Fn = () => new Set<any>();
const foo2: Fn = function test() {
  return new Set<any>();
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
					Line:      3,
				},
				{
					MessageId: "unsafeReturnAssignment",
					Line:      5,
				},
			},
		},
		{
			Code: `
type Fn = () => Set<string>;
function receiver(arg: Fn) {}
receiver(() => new Set<any>());
receiver(function test() {
  return new Set<any>();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
					Line:      4,
				},
				{
					MessageId: "unsafeReturnAssignment",
					Line:      6,
				},
			},
		},
		{
			Code: `
function foo() {
  return this;
}

function bar() {
  return () => this;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnThis",
					Line:      3,
					Column:    3,
					EndColumn: 15,
				},
				{
					MessageId: "unsafeReturnThis",
					Line:      7,
					Column:    16,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
declare function foo(arg: null | (() => any)): void;
foo(() => 'foo' as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    11,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
let value: NotKnown;

function example() {
  return value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      5,
					Column:    3,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
declare const value: any;
async function foo() {
  return value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare const value: Promise<any>;
async function foo(): Promise<number> {
  return value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo(arg: number) {
  return arg as Promise<any>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
function foo(): Promise<any> {
  return {} as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
function foo(): Promise<object> {
  return {} as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo(): Promise<object> {
  return Promise.resolve<any>({});
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo(): Promise<object> {
  return Promise.resolve<Promise<Promise<any>>>({} as Promise<any>);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo(): Promise<object> {
  return {} as Promise<Promise<Promise<Promise<any>>>>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo() {
  return {} as Promise<Promise<Promise<Promise<any>>>>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo() {
  return {} as Promise<any> | Promise<object>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo() {
  return {} as Promise<any | object>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
async function foo() {
  return {} as Promise<any> & { __brand: 'any' };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Alias<T> extends Promise<any> {
  foo: 'bar';
}

declare const value: Alias<number>;
async function foo() {
  return value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturn",
					Line:      8,
					Column:    3,
				},
			},
		},
		{
			Code: `
import type { Fn } from 'return-types';
const foo: Fn = () => new Set<any>();
      `,
			Files: map[string]string{
				"node_modules/return-types/package.json": `{
  "name": "return-types",
  "version": "1.0.0",
  "types": "index.d.ts"
}`,
				"node_modules/return-types/index.d.ts": `export type Fn = () => Set<string>;`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeReturnAssignment",
					Line:      3,
					Column:    23,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
interface Schema<Output = any> {
  readonly _output: Output;
}
type Infer<S extends Schema> = S['_output'];
const getOutput = <S extends Schema>(schema: S): Infer<S> =>
  schema._output;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn"},
			},
		},
		{
			FileName: "repro.js",
			TSConfig: "tsconfig.checkJs.json",
			Code: `
/**
 * @callback Fn
 * @param {*} value
 */

/** @type {Fn} */
const f = function (value) {
  return value;
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 9, Column: 3},
			},
		},
		{
			FileName: "repro.js",
			TSConfig: "tsconfig.checkJs.json",
			Code: `
/**
 * @callback Fn
 * @param {*} value
 */

/** @type {Fn} */
const f = value => value;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 8, Column: 20},
			},
		},
		{
			FileName: "repro.js",
			TSConfig: "tsconfig.checkJs.json",
			Code: `
/**
 * @callback Fn
 * @param {*} value
 * @returns {number}
 */

/** @type {Fn} */
const f = value => value;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeReturn", Line: 9, Column: 20},
			},
		},
	})
}

func TestNoUnsafeReturnRule_ReturnTypeHelp(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeReturnRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   "const foo = () => 1 as any;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
		{
			Code:   "const foo = function () { return 1 as any; };",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
		{
			Code:   "const foo = (): number => 1 as any;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
		{
			Code:   "const foo = function (): number { return 1 as any; };",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
		{
			Code:   "const foo: () => number = () => 1 as any;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
		{
			Code:   "const foo: () => number = function () { return 1 as any; };",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unsafeReturn"}},
		},
	})
}
