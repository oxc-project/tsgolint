package no_deprecated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoDeprecated(t *testing.T) {
	validTests := []rule_tester.ValidTestCase{
		// Declaring deprecated items should be allowed
		{
			Code: `/** @deprecated */ var a;`,
		},
		{
			Code: `/** @deprecated */ var a = 1;`,
		},
		{
			Code: `/** @deprecated */ let a;`,
		},
		{
			Code: `/** @deprecated */ let a = 1;`,
		},
		{
			Code: `/** @deprecated */ const a = 1;`,
		},
		{
			Code: `/** @deprecated */ declare var a: number;`,
		},
		{
			Code: `/** @deprecated */ declare let a: number;`,
		},
		{
			Code: `/** @deprecated */ declare const a: number;`,
		},
		{
			Code: `/** @deprecated */ export var a = 1;`,
		},
		{
			Code: `/** @deprecated */ export let a = 1;`,
		},
		{
			Code: `/** @deprecated */ export const a = 1;`,
		},
		{
			Code: `const [/** @deprecated */ a] = [b];`,
		},
		{
			Code: `const [/** @deprecated */ a] = b;`,
		},
		{
			Code: `
					const a = {
						b: 1,
						/** @deprecated */ c: 2,
					};
					a.b;
				`,
		},
		{
			Code: `
					const a = {
						b: 1,
						/** @deprecated */ c: 2,
					};
					a?.b;
				`,
		},
		{
			Code: `
					declare const a: {
						b: 1;
						/** @deprecated */ c: 2;
					};
					a.b;
				`,
		},
		{
			Code: `
					class A {
						b: 1;
						/** @deprecated */ c: 2;
					}
					new A().b;
				`,
		},
		{
			Code: `
					class A {
						accessor b: 1;
						/** @deprecated */ accessor c: 2;
					}
					new A().b;
				`,
		},
		{
			Code: `
					declare class A {
						/** @deprecated */
						static b: string;
						static c: string;
					}
					A.c;
				`,
		},
		{
			Code: `
					declare class A {
						/** @deprecated */
						static accessor b: string;
						static accessor c: string;
					}
					A.c;
				`,
		},
		{
			Code: `
					namespace A {
						/** @deprecated */
						export const b = '';
						export const c = '';
					}
					A.c;
				`,
		},
		{
			Code: `
					enum A {
						/** @deprecated */
						b = 'b',
						c = 'c',
					}
					A.c;
				`,
		},
		{
			Code: `
					function a(value: 'b' | undefined): void;
					/** @deprecated */
					function a(value: 'c' | undefined): void;
					function a(value: string | undefined): void {
						// ...
					}
					a('b');
				`,
		},
		{
			Code: `
					function a(value: 'b' | undefined): void;
					/** @deprecated */
					function a(value: 'c' | undefined): void;
					function a(value: string | undefined): void {
						// ...
					}
					export default a('b');
				`,
		},
		{
			Code: `
					function notDeprecated(): object {
						return {};
					}
					export default notDeprecated();
				`,
		},
		{
			Code: `
					class A {
						a(value: 'b'): void;
						/** @deprecated */
						a(value: 'c'): void;
					}
					declare const foo: A;
					foo.a('b');
				`,
		},
		{
			Code: `
					const A = class {
						/** @deprecated */
						constructor();
						constructor(arg: string);
						constructor(arg?: string) {}
					};
					new A('a');
				`,
		},
		{
			Code: `
					type A = {
						(value: 'b'): void;
						/** @deprecated */
						(value: 'c'): void;
					};
					declare const foo: A;
					foo('b');
				`,
		},
		{
			Code: `
					declare const a: {
						new (value: 'b'): void;
						/** @deprecated */
						new (value: 'c'): void;
					};
					new a('b');
				`,
		},
		{
			Code: `
					namespace assert {
						export function fail(message?: string | Error): never;
						/** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
						export function fail(actual: unknown, expected: unknown): never;
					}
					assert.fail('');
				`,
		},
		{
			Code: `
					declare module 'deprecations' {
						/** @deprecated */
						export const value = true;
					}
					import { value } from 'deprecations';
				`,
		},
		{
			Code: `
					/** @deprecated Use ts directly. */
					export * as ts from 'typescript';
				`,
		},
		{
			Code: `
					export {
						/** @deprecated Use ts directly. */
						default as ts,
					} from 'typescript';
				`,
		},
		{
			Code: `
					namespace A {
						/** @deprecated */
						export type B = string;
						export type C = string;
						export type D = string;
					}
					export type D = A.C | A.D;
				`,
		},
		{
			Code: `
					interface Props {
						anchor: 'foo';
					}
					declare const x: Props;
					const { anchor = '' } = x;
				`,
		},
		{
			Code: `
					namespace Foo {}
					/**
					 * @deprecated
					 */
					export import Bar = Foo;
				`,
		},
		{
			Code: `
					/**
					 * @deprecated
					 */
					export import Bar = require('./deprecated');
				`,
		},
		{
			Code: `
					interface Props {
						anchor: 'foo';
					}
					declare const x: { bar: Props };
					const {
						bar: { anchor = '' },
					} = x;
				`,
		},
		{
			Code: `
					interface Props {
						anchor: 'foo';
					}
					declare const x: [item: Props];
					const [{ anchor = 'bar' }] = x;
				`,
		},
		{
			Code: `function fn(/** @deprecated */ foo = 4) {}`,
		},
		{
			Code: `call();`,
		},
		{
			Code: `
					class Foo implements Foo {
						get bar(): number {
							return 42;
						}
						baz(): number {
							return this.bar;
						}
					}
				`,
		},
		{
			Code: `
					declare namespace JSX {}
					<foo bar={1} />;
				`,
		},
		{
			Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: any;
						}
					}
					<foo bar={1} />;
				`,
		},
		{
			Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: unknown;
						}
					}
					<foo bar={1} />;
				`,
		},
		{
			Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: {
								bar: any;
							};
						}
					}
					<foo bar={1} />;
				`,
		},
		{
			Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: {
								bar: unknown;
							};
						}
					}
					<foo bar={1} />;
				`,
		},
		{
			Code: `
					export {
						/** @deprecated */
						foo,
					};
				`,
		},
		{
			Code: `
					declare const test: string;
					const bar = { test };
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const complex = Symbol() as any;
					const c = a[complex];
				`,
		},
		{
			Code: `
					const a = {
						b: 'string',
					};
					const c = a['b'];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a['nonExistentProperty'];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					function getKey() {
						return 'c';
					}
					const c = a[getKey()];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = {};
					const c = a[key];
				`,
		},
		{
			Code: `
					const stringObj = new String('b');
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a[stringObj];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = Symbol('key');
					const c = a[key];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = null;
					const c = a[key as any];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = {};
					const c = a[key as any];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = Symbol();
					const c = a[key as any];
				`,
		},
		{
			Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = undefined;
					const c = a[key as any];
				`,
		},
	}

	invalidTests := []rule_tester.InvalidTestCase{
		// JSX with deprecated attribute
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
					Line:      9,
					Column:    19,
				},
			},
		},

		// Using deprecated variable
		{
			Code: `
					/** @deprecated */ var a = undefined;
					a;
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      3,
					Column:    6,
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
					Line:      3,
					Column:    6,
				},
			},
		},

		// Using deprecated let
		{
			Code: `
					/** @deprecated */ let a = undefined;
					a;
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      3,
					Column:    6,
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
					Line:      3,
					Column:    6,
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
					Line:      3,
					Column:    6,
				},
			},
		},

		// Using deprecated const
		{
			Code: `
					/** @deprecated */ const a = { b: 1 };
					const c = a;
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      3,
					Column:    16,
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
					Line:      3,
					Column:    16,
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
					Line:      3,
					Column:    18,
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
					Line:      3,
					Column:    17,
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
					Line:      3,
					Column:    18,
				},
			},
		},
		{
			Code: "/** @deprecated */ const a = 'foo';\nimport(`./path/${a}.js`);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      3,
					Column:    23,
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
					Line:      4,
					Column:    10,
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
					Line:      3,
					Column:    18,
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
					Line:      3,
					Column:    18,
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
					Line:      3,
					Column:    6,
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
					Line:      3,
					Column:    6,
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
					Line:      3,
					Column:    6,
				},
			},
		},

		// Using deprecated property
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
					Line:      5,
					Column:    8,
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
					Line:      5,
					Column:    8,
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
					Line:      3,
					Column:    16,
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
					Line:      3,
					Column:    20,
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
					Line:      5,
					Column:    14,
				},
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    16,
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
					Line:      5,
					Column:    8,
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
					Line:      3,
					Column:    26,
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
					Line:      3,
					Column:    23,
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
					Line:      4,
					Column:    19,
				},
			},
		},

		// Using deprecated class
		{
			Code: `
					/** @deprecated */
					class A {}
					new A();
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    10,
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
					Line:      4,
					Column:    10,
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
					Line:      4,
					Column:    10,
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
					Line:      4,
					Column:    10,
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
					Line:      6,
					Column:    10,
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
					Line:      8,
					Column:    10,
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
					Line:      6,
					Column:    10,
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
					Line:      6,
					Column:    10,
				},
			},
		},

		// Using deprecated class members
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
					Line:      7,
					Column:    14,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      9,
					Column:    8,
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
					Line:      8,
					Column:    8,
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
					Line:      6,
					Column:    8,
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
					Line:      6,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    14,
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
					Line:      7,
					Column:    14,
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
					Line:      7,
					Column:    14,
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
					Line:      7,
					Column:    14,
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
					Line:      4,
					Column:    23,
				},
			},
		},

		// Using deprecated namespace members
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
					Line:      6,
					Column:    8,
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
					Line:      6,
					Column:    8,
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
					Line:      6,
					Column:    8,
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
					Line:      7,
					Column:    13,
				},
			},
		},

		// Using deprecated enum
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
					Line:      6,
					Column:    6,
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
					Line:      6,
					Column:    8,
				},
			},
		},

		// Using deprecated function
		{
			Code: `
					/** @deprecated */
					function a() {}
					a();
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    6,
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
					Line:      5,
					Column:    6,
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
					Line:      6,
					Column:    6,
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
					Line:      8,
					Column:    6,
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
					Line:      6,
					Column:    15,
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
					Line:      9,
					Column:    12,
				},
			},
		},
		{
			Code: "/** @deprecated */\ndeclare function a(...args: unknown[]): string;\na``;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    6,
				},
			},
		},

		// Using deprecated JSX component
		{
			Code: `
					/** @deprecated */
					const A = () => <div />;
					const a = <A />;
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    17,
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
					Line:      4,
					Column:    17,
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
					Line:      6,
					Column:    17,
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
					Line:      6,
					Column:    17,
				},
			},
		},

		// Using deprecated type
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
					Line:      6,
					Column:    22,
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
					Line:      8,
					Column:    24,
				},
			},
		},

		// Property destructuring with deprecated
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
					Line:      7,
					Column:    14,
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
					Line:      8,
					Column:    15,
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
					Line:      7,
					Column:    15,
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
					Line:      7,
					Column:    14,
				},
			},
		},

		// Using deprecated class accessor
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
					Line:      7,
					Column:    8,
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
					Line:      7,
					Column:    8,
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
					Line:      9,
					Column:    8,
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
					Line:      6,
					Column:    8,
				},
			},
		},

		// Using deprecated private identifier
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
					Line:      6,
					Column:    15,
				},
			},
		},

		// Computed property access
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
					Line:      6,
					Column:    18,
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
					Line:      7,
					Column:    18,
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
					Line:      7,
					Column:    18,
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
					Line:      8,
					Column:    18,
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
					Line:      10,
					Column:    18,
				},
			},
		},
		{
			Code: "const a = {\n\t/** @deprecated */\n\tb: 'string',\n};\nconst key = `b`;\nconst c = a[key];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    18,
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
					Line:      7,
					Column:    18,
				},
			},
		},

		// Export with deprecated
		{
			Code: `
					import { deprecatedFunction } from './deprecated';
					export { deprecatedFunction };
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      3,
					Column:    15,
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
					Line:      2,
					Column:    15,
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
					Line:      2,
					Column:    26,
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
					Line:      2,
					Column:    37,
				},
			},
		},

		// Extends/implements with deprecated
		{
			Code: `
					/** @deprecated */
					interface Foo {}
					class Bar implements Foo {}
				`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    27,
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
					Line:      4,
					Column:    34,
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
					Line:      5,
					Column:    39,
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
					Line:      4,
					Column:    31,
				},
			},
		},

		// Decorator with deprecated
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
					Line:      4,
					Column:    7,
				},
			},
		},

		// Export default with deprecated
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
					Line:      6,
					Column:    21,
				},
			},
		},

		// Super with deprecated
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
					Line:      9,
					Column:    8,
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
					Line:      9,
					Column:    8,
				},
			},
		},

		// JSX with deprecated aria attribute
		{
			Code: `const a = <div aria-grabbed></div>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecatedWithReason",
					Line:      1,
					Column:    16,
				},
			},
		},

		// JSX with deprecated property in namespaced element
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
					Line:      13,
					Column:    56,
				},
			},
		},

		// JSX with deprecated property in component
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
					Line:      15,
					Column:    39,
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, validTests, invalidTests)
}
