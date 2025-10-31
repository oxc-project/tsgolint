package no_deprecated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestNoDeprecatedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, []rule_tester.ValidTestCase{
		{Code: `/** @deprecated */ var a;`},
		{Code: `/** @deprecated */ var a = 1;`},
		{Code: `/** @deprecated */ let a;`},
		{Code: `/** @deprecated */ let a = 1;`},
		{Code: `/** @deprecated */ const a = 1;`},
		{Code: `/** @deprecated */ declare var a: number;`},
		{Code: `/** @deprecated */ declare let a: number;`},
		{Code: `/** @deprecated */ declare const a: number;`},
		{Code: `/** @deprecated */ export var a = 1;`},
		{Code: `/** @deprecated */ export let a = 1;`},
		{Code: `/** @deprecated */ export const a = 1;`},
		{Code: `const [/** @deprecated */ a] = [b];`},
		{Code: `const [/** @deprecated */ a] = b;`},
		{Code: `
      const a = {
        b: 1,
        /** @deprecated */ c: 2,
      };

      a.b;
    `},
		{Code: `
      const a = {
        b: 1,
        /** @deprecated */ c: 2,
      };

      a?.b;
    `},
		{Code: `
      declare const a: {
        b: 1;
        /** @deprecated */ c: 2;
      };

      a.b;
    `},
		{Code: `
      class A {
        b: 1;
        /** @deprecated */ c: 2;
      }

      new A().b;
    `},
		{Code: `
      class A {
        accessor b: 1;
        /** @deprecated */ accessor c: 2;
      }

      new A().b;
    `},
		{Code: `
      declare class A {
        /** @deprecated */
        static b: string;
        static c: string;
      }

      A.c;
    `},
		{Code: `
      declare class A {
        /** @deprecated */
        static accessor b: string;
        static accessor c: string;
      }

      A.c;
    `},
		{Code: `
      namespace A {
        /** @deprecated */
        export const b = '';
        export const c = '';
      }

      A.c;
    `},
		{Code: `
      enum A {
        /** @deprecated */
        b = 'b',
        c = 'c',
      }

      A.c;
    `},
		{Code: `
      function a(value: 'b' | undefined): void;
      /** @deprecated */
      function a(value: 'c' | undefined): void;
      function a(value: string | undefined): void {
        // ...
      }

      a('b');
    `},
		{Code: `
      function a(value: 'b' | undefined): void;
      /** @deprecated */
      function a(value: 'c' | undefined): void;
      function a(value: string | undefined): void {
        // ...
      }

      export default a('b');
    `},
		{Code: `
      function notDeprecated(): object {
        return {};
      }

      export default notDeprecated();
    `},
		{Code: `
      import { deprecatedFunctionWithOverloads } from './deprecated';

      const foo = deprecatedFunctionWithOverloads();
    `},
		{Code: `
      import * as imported from './deprecated';

      const foo = imported.deprecatedFunctionWithOverloads();
    `},
		{Code: `
      import { ClassWithDeprecatedConstructor } from './deprecated';

      const foo = new ClassWithDeprecatedConstructor();
    `},
		{Code: `
      import * as imported from './deprecated';

      const foo = new imported.ClassWithDeprecatedConstructor();
    `},
		{Code: `
      class A {
        a(value: 'b'): void;
        /** @deprecated */
        a(value: 'c'): void;
      }
      declare const foo: A;
      foo.a('b');
    `},
		{Code: `
      const A = class {
        /** @deprecated */
        constructor();
        constructor(arg: string);
        constructor(arg?: string) {}
      };

      new A('a');
    `},
		{Code: `
      type A = {
        (value: 'b'): void;
        /** @deprecated */
        (value: 'c'): void;
      };
      declare const foo: A;
      foo('b');
    `},
		{Code: `
      declare const a: {
        new (value: 'b'): void;
        /** @deprecated */
        new (value: 'c'): void;
      };
      new a('b');
    `},
		{Code: `
      namespace assert {
        export function fail(message?: string | Error): never;
        /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
        export function fail(actual: unknown, expected: unknown): never;
      }

      assert.fail('');
    `},
		{Code: `
      import assert from 'node:assert';

      assert.fail('');
    `},
		{Code: `
      declare module 'deprecations' {
        /** @deprecated */
        export const value = true;
      }

      import { value } from 'deprecations';
    `},
		{Code: `
      /** @deprecated Use ts directly. */
      export * as ts from 'typescript';
    `},
		{Code: `
      export {
        /** @deprecated Use ts directly. */
        default as ts,
      } from 'typescript';
    `},
		{Code: `
      export { deprecatedFunction as 'bur' } from './deprecated';
    `},
		{Code: `
      export { 'deprecatedFunction' } from './deprecated';
    `},
		{Code: `
      namespace A {
        /** @deprecated */
        export type B = string;
        export type C = string;
        export type D = string;
      }

      export type D = A.C | A.D;
    `},
		{Code: `
      interface Props {
        anchor: 'foo';
      }
      declare const x: Props;
      const { anchor = '' } = x;
    `},
		{Code: `
      namespace Foo {}

      /**
       * @deprecated
       */
      export import Bar = Foo;
    `},
		{Code: `
      /**
       * @deprecated
       */
      export import Bar = require('./deprecated');
    `},
		{Code: `
      interface Props {
        anchor: 'foo';
      }
      declare const x: { bar: Props };
      const {
        bar: { anchor = '' },
      } = x;
    `},
		{Code: `
      interface Props {
        anchor: 'foo';
      }
      declare const x: [item: Props];
      const [{ anchor = 'bar' }] = x;
    `},
		{Code: `function fn(/** @deprecated */ foo = 4) {}`},
		{Code: `
        async function fn() {
          const d = await import('./deprecated.js');
          d.default;
        }
      `},
		{Code: `call();`},
		{Code: `
      class Foo implements Foo {
        get bar(): number {
          return 42;
        }

        baz(): number {
          return this.bar;
        }
      }
    `},
		{Code: `
      declare namespace JSX {}

      <foo bar={1} />;
    `},
		{Code: `
      declare namespace JSX {
        interface IntrinsicElements {
          foo: any;
        }
      }

      <foo bar={1} />;
    `},
		{Code: `
      declare namespace JSX {
        interface IntrinsicElements {
          foo: unknown;
        }
      }

      <foo bar={1} />;
    `},
		{Code: `
      declare namespace JSX {
        interface IntrinsicElements {
          foo: {
            bar: any;
          };
        }
      }
      <foo bar={1} />;
    `},
		{Code: `
      declare namespace JSX {
        interface IntrinsicElements {
          foo: {
            bar: unknown;
          };
        }
      }
      <foo bar={1} />;
    `},
		{Code: `
      export {
        /** @deprecated */
        foo,
      };
    `},
		{Code: `
/** @deprecated */
function A() {
  return <div />;
}

const a = <A></A>;
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"A"}}}}},
		{Code: `
/** @deprecated */
declare class A {}

new A();
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"A"}}}}},
		{Code: `
/** @deprecated */
const deprecatedValue = 45;
const bar = deprecatedValue;
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"deprecatedValue"}}}}},
		{Code: `
class MyClass {
  /** @deprecated */
  #privateProp = 42;
  value = this.#privateProp;
}
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"privateProp"}}}}},
		{Code: `
/** @deprecated */
const deprecatedValue = 45;
const bar = deprecatedValue;
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"deprecatedValue"}}}}},
		{Code: `
import { exists } from 'fs';
exists('/foo');
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"exists"}, Package: "fs"}}}},
		{Code: `
const { exists } = import('fs');
exists('/foo');
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"exists"}, Package: "fs"}}}},
		{Code: `
      declare const test: string;
      const bar = { test };
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const complex = Symbol() as any;
      const c = a[complex];
    `},
		{Code: `
      const a = {
        b: 'string',
      };

      const c = a['b'];
    `},
		{Code: `
        interface AllowedType {
          /** @deprecated */
          prop: string;
        }

        const obj: AllowedType = {
          prop: 'test',
        };

        const value = obj['prop'];
      `, Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromFile, Name: []string{"AllowedType"}}}}},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = {};
      const c = a[key as any];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = Symbol();
      const c = a[key as any];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = undefined;
      const c = a[key as any];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const c = a['nonExistentProperty'];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      function getKey() {
        return 'c';
      }

      const c = a[getKey()];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = {};
      const c = a[key];
    `},
		{Code: `
      const stringObj = new String('b');
      const a = {
        /** @deprecated */
        b: 'string',
      };
      const c = a[stringObj];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = Symbol('key');
      const c = a[key];
    `},
		{Code: `
      const a = {
        /** @deprecated */
        b: 'string',
      };

      const key = null;
      const c = a[key as any];
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        interface AProps {
          /** @deprecated */
          b: number | string;
        }

        function A(props: AProps) {
          return <div />;
        }

        const a = <A b="" />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ var a = undefined;
        a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ export var a = undefined;
        a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ let a = undefined;
        a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ export let a = undefined;
        a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ let aLongName = undefined;
        aLongName;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const c = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated Reason. */ const a = { b: 1 };
        const c = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const { c = a } = {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const [c = a] = [];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: "\n        /** @deprecated */ const a = 'foo';\n        import(\\`./path/\\${a}.js\\`);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare function log(...args: unknown): void;

        /** @deprecated */ const a = { b: 1 };

        log(a);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a.b);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a?.b);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: { c: 1 } };
        a.b.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: { c: 1 } };
        a.b?.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: { c: 1 } };
        a?.b?.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const a = {
          /** @deprecated */ b: { c: 1 },
        };
        a.b.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare const a: {
          /** @deprecated */ b: { c: 1 };
        };
        a.b.c;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const c = a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const { c } = a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare const test: string;
        const myObj = {
          prop: test,
          deep: {
            prop: test,
          },
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare const test: string;
        const bar = {
          test,
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const { c = 'd' } = a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */ const a = { b: 1 };
        const { c: d } = a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare const a: string[];
        const [b] = [a];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        class A {}

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        export class A {}

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        const A = class {};

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare class A {}

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const A = class {
          /** @deprecated */
          constructor() {}
        };

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const A = class {
          /** @deprecated */
          constructor();
          constructor(arg: string);
          constructor(arg?: string) {}
        };

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare const A: {
          /** @deprecated */
          new (): string;
        };

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare class A {
          constructor();
        }

        new A();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        class A {
          /** @deprecated */
          b: string;
        }

        declare const a: A;

        const { b } = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          b(): string;
        }

        declare const a: A;

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          b(): string;
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          b: () => string;
        }

        declare const a: A;

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          b: () => string;
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface A {
          /** @deprecated */
          b: () => string;
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        class A {
          /** @deprecated */
          b(): string {
            return '';
          }
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated Use b(value). */
          b(): string;
          b(value: string): string;
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          static b: string;
        }

        A.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare const a: {
          /** @deprecated */
          b: string;
        };

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface A {
          /** @deprecated */
          b: string;
        }

        declare const a: A;

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export interface A {
          /** @deprecated */
          b: string;
        }

        declare const a: A;

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface A {
          /** @deprecated */
          b: string;
        }

        declare const a: A;

        const { b } = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        type A = {
          /** @deprecated */
          b: string;
        };

        declare const a: A;

        const { b } = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export type A = {
          /** @deprecated */
          b: string;
        };

        declare const a: A;

        const { b } = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        type A = () => {
          /** @deprecated */
          b: string;
        };

        declare const a: A;

        const { b } = a();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        type A = string[];

        declare const a: A;

        const [b] = a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        namespace A {
          /** @deprecated */
          export const b = '';
        }

        A.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export namespace A {
          /** @deprecated */
          export const b = '';
        }

        A.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        namespace A {
          /** @deprecated */
          export function b() {}
        }

        A.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        namespace assert {
          export function fail(message?: string | Error): never;
          /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
          export function fail(actual: unknown, expected: unknown): never;
        }

        assert.fail({}, {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import assert from 'node:assert';

        assert.fail({}, {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        enum A {
          a,
        }

        A.a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        enum A {
          /** @deprecated */
          a,
        }

        A.a;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        function a() {}

        a();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        function a(): void;
        function a() {}

        a();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        function a(): void;
        /** @deprecated */
        function a(value: string): void;
        function a(value?: string) {}

        a('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        type A = {
          (value: 'b'): void;
          /** @deprecated */
          (value: 'c'): void;
        };
        declare const foo: A;
        foo('c');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        function a(
          /** @deprecated */
          b?: boolean,
        ) {
          return b;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export function isTypeFlagSet(
          type: ts.Type,
          flagsToCheck: ts.TypeFlags,
          /** @deprecated This param is not used and will be removed in the future. */
          isReceiver?: boolean,
        ): boolean {
          const flags = getTypeFlags(type);

          if (isReceiver && flags & ANY_OR_UNKNOWN) {
            return true;
          }

          return (flags & flagsToCheck) !== 0;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: "\n        /** @deprecated */\n        declare function a(...args: unknown[]): string;\n\n        a\\`\\`;\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        const A = () => <div />;

        const a = <A />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        const A = () => <div />;

        const a = <A></A>;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        function A() {
          return <div />;
        }

        const a = <A />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        function A() {
          return <div />;
        }

        const a = <A></A>;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        export type A = string;
        export type B = string;
        export type C = string;

        export type D = A | B | C;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        namespace A {
          /** @deprecated */
          export type B = string;
          export type C = string;
          export type D = string;
        }

        export type D = A.B | A.C | A.D;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface Props {
          /** @deprecated */
          anchor: 'foo';
        }
        declare const x: Props;
        const { anchor = '' } = x;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface Props {
          /** @deprecated */
          anchor: 'foo';
        }
        declare const x: { bar: Props };
        const {
          bar: { anchor = '' },
        } = x;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface Props {
          /** @deprecated */
          anchor: 'foo';
        }
        declare const x: [item: Props];
        const [{ anchor = 'bar' }] = x;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        interface Props {
          /** @deprecated */
          foo: Props;
        }
        declare const x: Props;
        const { foo = x } = x;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { DeprecatedClass } from './deprecated';

        const foo = new DeprecatedClass();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { DeprecatedClass } from './deprecated';

        declare function inject(something: new () => unknown): void;

        inject(DeprecatedClass);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { deprecatedVariable } from './deprecated';

        const foo = deprecatedVariable;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { DeprecatedClass } from './deprecated';

        declare const x: DeprecatedClass;

        const { foo } = x;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { deprecatedFunction } from './deprecated';

        deprecatedFunction();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = new imported.NormalClass();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { NormalClass } from './deprecated';

        const foo = new NormalClass();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.NormalClass;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { NormalClass } from './deprecated';

        const foo = NormalClass;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { normalVariable } from './deprecated';

        const foo = normalVariable;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.normalVariable;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const { normalVariable } = imported;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { deprecatedVariable } from './deprecated';

        const test = {
          someField: deprecatedVariable,
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { normalFunction } from './deprecated';

        const foo = normalFunction;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.normalFunction;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const { normalFunction } = imported;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { normalFunction } from './deprecated';

        const foo = normalFunction();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.normalFunction();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { deprecatedFunctionWithOverloads } from './deprecated';

        const foo = deprecatedFunctionWithOverloads('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.deprecatedFunctionWithOverloads('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

        const foo = reexportedDeprecatedFunctionWithOverloads;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.reexportedDeprecatedFunctionWithOverloads;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const { reexportedDeprecatedFunctionWithOverloads } = imported;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

        const foo = reexportedDeprecatedFunctionWithOverloads();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.reexportedDeprecatedFunctionWithOverloads();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';

        const foo = reexportedDeprecatedFunctionWithOverloads('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.reexportedDeprecatedFunctionWithOverloads('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import { ClassWithDeprecatedConstructor } from './deprecated';

        const foo = new ClassWithDeprecatedConstructor('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = new imported.ClassWithDeprecatedConstructor('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

        const foo = ReexportedClassWithDeprecatedConstructor;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.ReexportedClassWithDeprecatedConstructor;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const { ReexportedClassWithDeprecatedConstructor } = imported;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

        const foo = ReexportedClassWithDeprecatedConstructor();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.ReexportedClassWithDeprecatedConstructor();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import { ReexportedClassWithDeprecatedConstructor } from './deprecated';

        const foo = ReexportedClassWithDeprecatedConstructor('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import * as imported from './deprecated';

        const foo = imported.ReexportedClassWithDeprecatedConstructor('a');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        import imported from './deprecated';

        imported;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        async function fn() {
          const d = await import('./deprecated.js');
          d.default.default;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        interface Foo {}

        class Bar implements Foo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        interface Foo {}

        export class Bar implements Foo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        interface Foo {}

        interface Baz {}

        export class Bar implements Baz, Foo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        class Foo {}

        export class Bar extends Foo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        declare function decorator(constructor: Function);

        @decorator
        export class Foo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        /** @deprecated */
        function a(): object {
          return {};
        }

        export default a();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
class A {
  /** @deprecated */
  constructor() {}
}

class B extends A {
  constructor() {
    /** should report but does not */
    super();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
class A {
  /** @deprecated test reason*/
  constructor() {}
}

class B extends A {
  constructor() {
    /** should report but does not */
    super();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `const a = <div aria-grabbed></div>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        declare namespace JSX {
          interface IntrinsicElements {
            'foo-bar:baz-bam': {
              name: string;
              /**
               * @deprecated
               */
              deprecatedProp: string;
            };
          }
        }

        const componentDashed = <foo-bar:baz-bam name="e" deprecatedProp="oh no" />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import * as React from 'react';

        interface Props {
          /**
           * @deprecated
           */
          deprecatedProp: string;
        }

        interface Tab {
          List: React.FC<Props>;
        }

        const Tab: Tab = {
          List: () => <div>Hi</div>,
        };

        const anotherExample = <Tab.List deprecatedProp="oh no" />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
import { exists } from 'fs';
exists('/foo');
      `,
			Options: NoDeprecatedOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"exists"}, Package: "hoge"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          accessor b: () => string;
        }

        declare const a: A;

        a.b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          accessor b: () => string;
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        class A {
          /** @deprecated */
          accessor b = (): string => {
            return '';
          };
        }

        declare const a: A;

        a.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        declare class A {
          /** @deprecated */
          static accessor b: () => string;
        }

        A.b();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        class A {
          /** @deprecated */
          #b = () => {};

          c() {
            this.#b();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const a = {
          /** @deprecated */
          b: 'string',
        };

        const c = a['b'];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const a = {
          /** @deprecated */
          b: 'string',
        };
        const x = 'b';
        const c = a[x];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const a = {
          /** @deprecated */
          [2]: 'string',
        };
        const x = 'b';
        const c = a[2];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const a = {
          /** @deprecated reason for deprecation */
          b: 'string',
        };

        const key = 'b';
        const stringKey = key as const;
        const c = a[stringKey];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        enum Keys {
          B = 'b',
        }

        const a = {
          /** @deprecated reason for deprecation */
          b: 'string',
        };

        const key = Keys.B;
        const c = a[key];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: "\n        const a = {\n          /** @deprecated */\n          b: 'string',\n        };\n\n        const key = \\`b\\`;\n        const c = a[key];\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        const stringObj = 'b';
        const a = {
          /** @deprecated */
          b: 'string',
        };
        const c = a[stringObj];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        import { deprecatedFunction } from './deprecated';

        export { deprecatedFunction };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export { deprecatedFunction } from './deprecated';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export type { T, U } from './deprecated';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
				},
			},
		},
		{
			Code: `
        export { default as foo } from './deprecated';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
		{
			Code: `
        export { deprecatedFunction as bar } from './deprecated';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
				},
			},
		},
	})
}
