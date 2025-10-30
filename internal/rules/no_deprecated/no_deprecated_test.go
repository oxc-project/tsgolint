package no_deprecated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
)

func TestNoDeprecated(t *testing.T) {
	rule_tester.Run(
		t,
		NoDeprecatedRule,
		map[string]rule_tester.ValidTestCase{
			// Declaring deprecated items should be allowed
			"deprecated var": {
				Code: `/** @deprecated */ var a;`,
			},
			"deprecated var with value": {
				Code: `/** @deprecated */ var a = 1;`,
			},
			"deprecated let": {
				Code: `/** @deprecated */ let a;`,
			},
			"deprecated let with value": {
				Code: `/** @deprecated */ let a = 1;`,
			},
			"deprecated const": {
				Code: `/** @deprecated */ const a = 1;`,
			},
			"deprecated declare var": {
				Code: `/** @deprecated */ declare var a: number;`,
			},
			"deprecated declare let": {
				Code: `/** @deprecated */ declare let a: number;`,
			},
			"deprecated declare const": {
				Code: `/** @deprecated */ declare const a: number;`,
			},
			"deprecated export var": {
				Code: `/** @deprecated */ export var a = 1;`,
			},
			"deprecated export let": {
				Code: `/** @deprecated */ export let a = 1;`,
			},
			"deprecated export const": {
				Code: `/** @deprecated */ export const a = 1;`,
			},
			"deprecated in array destructuring 1": {
				Code: `const [/** @deprecated */ a] = [b];`,
			},
			"deprecated in array destructuring 2": {
				Code: `const [/** @deprecated */ a] = b;`,
			},
			"access non-deprecated property": {
				Code: `
					const a = {
						b: 1,
						/** @deprecated */ c: 2,
					};
					a.b;
				`,
			},
			"access non-deprecated property with optional chaining": {
				Code: `
					const a = {
						b: 1,
						/** @deprecated */ c: 2,
					};
					a?.b;
				`,
			},
			"access non-deprecated property in type": {
				Code: `
					declare const a: {
						b: 1;
						/** @deprecated */ c: 2;
					};
					a.b;
				`,
			},
			"access non-deprecated class property": {
				Code: `
					class A {
						b: 1;
						/** @deprecated */ c: 2;
					}
					new A().b;
				`,
			},
			"access non-deprecated class accessor": {
				Code: `
					class A {
						accessor b: 1;
						/** @deprecated */ accessor c: 2;
					}
					new A().b;
				`,
			},
			"access non-deprecated static property": {
				Code: `
					declare class A {
						/** @deprecated */
						static b: string;
						static c: string;
					}
					A.c;
				`,
			},
			"access non-deprecated static accessor": {
				Code: `
					declare class A {
						/** @deprecated */
						static accessor b: string;
						static accessor c: string;
					}
					A.c;
				`,
			},
			"access non-deprecated namespace member": {
				Code: `
					namespace A {
						/** @deprecated */
						export const b = '';
						export const c = '';
					}
					A.c;
				`,
			},
			"access non-deprecated enum member": {
				Code: `
					enum A {
						/** @deprecated */
						b = 'b',
						c = 'c',
					}
					A.c;
				`,
			},
			"call non-deprecated overload": {
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
			"export default non-deprecated call": {
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
			"export default non-deprecated function": {
				Code: `
					function notDeprecated(): object {
						return {};
					}
					export default notDeprecated();
				`,
			},
			"call non-deprecated class method overload": {
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
			"call non-deprecated constructor overload": {
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
			"call non-deprecated function type overload": {
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
			"call non-deprecated constructor type overload": {
				Code: `
					declare const a: {
						new (value: 'b'): void;
						/** @deprecated */
						new (value: 'c'): void;
					};
					new a('b');
				`,
			},
			"call non-deprecated namespace function overload": {
				Code: `
					namespace assert {
						export function fail(message?: string | Error): never;
						/** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
						export function fail(actual: unknown, expected: unknown): never;
					}
					assert.fail('');
				`,
			},
			"import deprecated from module": {
				Code: `
					declare module 'deprecations' {
						/** @deprecated */
						export const value = true;
					}
					import { value } from 'deprecations';
				`,
			},
			"export deprecated with alias": {
				Code: `
					/** @deprecated Use ts directly. */
					export * as ts from 'typescript';
				`,
			},
			"export deprecated default with alias": {
				Code: `
					export {
						/** @deprecated Use ts directly. */
						default as ts,
					} from 'typescript';
				`,
			},
			"export deprecated type alias": {
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
			"object property shorthand with default": {
				Code: `
					interface Props {
						anchor: 'foo';
					}
					declare const x: Props;
					const { anchor = '' } = x;
				`,
			},
			"deprecated import export": {
				Code: `
					namespace Foo {}
					/**
					 * @deprecated
					 */
					export import Bar = Foo;
				`,
			},
			"deprecated require import": {
				Code: `
					/**
					 * @deprecated
					 */
					export import Bar = require('./deprecated');
				`,
			},
			"nested object destructuring with default": {
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
			"array destructuring with default": {
				Code: `
					interface Props {
						anchor: 'foo';
					}
					declare const x: [item: Props];
					const [{ anchor = 'bar' }] = x;
				`,
			},
			"function parameter with deprecated jsdoc": {
				Code: `function fn(/** @deprecated */ foo = 4) {}`,
			},
			"call without parentheses": {
				Code: `call();`,
			},
			"class implements itself": {
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
			"JSX with no intrinsic elements": {
				Code: `
					declare namespace JSX {}
					<foo bar={1} />;
				`,
			},
			"JSX with any intrinsic elements": {
				Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: any;
						}
					}
					<foo bar={1} />;
				`,
			},
			"JSX with unknown intrinsic elements": {
				Code: `
					declare namespace JSX {
						interface IntrinsicElements {
							foo: unknown;
						}
					}
					<foo bar={1} />;
				`,
			},
			"JSX with any property": {
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
			"JSX with unknown property": {
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
			"export deprecated identifier": {
				Code: `
					export {
						/** @deprecated */
						foo,
					};
				`,
			},
			"shorthand property with non-deprecated value": {
				Code: `
					declare const test: string;
					const bar = { test };
				`,
			},
			"computed property access with non-deprecated symbol": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const complex = Symbol() as any;
					const c = a[complex];
				`,
			},
			"computed property access with string": {
				Code: `
					const a = {
						b: 'string',
					};
					const c = a['b'];
				`,
			},
			"computed property access with non-literal": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a['nonExistentProperty'];
				`,
			},
			"computed property access with function call": {
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
			"computed property access with object key": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = {};
					const c = a[key];
				`,
			},
			"computed property access with String object": {
				Code: `
					const stringObj = new String('b');
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a[stringObj];
				`,
			},
			"computed property access with Symbol": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = Symbol('key');
					const c = a[key];
				`,
			},
			"computed property access with null": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = null;
					const c = a[key as any];
				`,
			},
			"computed property access with any key": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = {};
					const c = a[key as any];
				`,
			},
			"computed property access with Symbol any": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = Symbol();
					const c = a[key as any];
				`,
			},
			"computed property access with undefined": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = undefined;
					const c = a[key as any];
				`,
			},
		},
		map[string]rule_tester.InvalidTestCase{
			// JSX with deprecated attribute
			"JSX deprecated attribute": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      9,
						Column:    19,
					},
				},
			},

			// Using deprecated variable
			"deprecated var usage": {
				Code: `
					/** @deprecated */ var a = undefined;
					a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},
			"deprecated export var usage": {
				Code: `
					/** @deprecated */ export var a = undefined;
					a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},

			// Using deprecated let
			"deprecated let usage": {
				Code: `
					/** @deprecated */ let a = undefined;
					a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},
			"deprecated export let usage": {
				Code: `
					/** @deprecated */ export let a = undefined;
					a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},
			"deprecated let with long name": {
				Code: `
					/** @deprecated */ let aLongName = undefined;
					aLongName;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},

			// Using deprecated const
			"deprecated const usage": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const c = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    16,
					},
				},
			},
			"deprecated const usage with reason": {
				Code: `
					/** @deprecated Reason. */ const a = { b: 1 };
					const c = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      3,
						Column:    16,
					},
				},
			},
			"deprecated const in object destructuring default": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const { c = a } = {};
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    18,
					},
				},
			},
			"deprecated const in array destructuring default": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const [c = a] = [];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    17,
					},
				},
			},
			"deprecated const as function argument": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					console.log(a);
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    18,
					},
				},
			},
			"deprecated const in template literal": {
				Code: `
					/** @deprecated */ const a = 'foo';
					import(\`./path/\${a}.js\`);
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    23,
					},
				},
			},
			"deprecated const as spread argument": {
				Code: `
					declare function log(...args: unknown): void;
					/** @deprecated */ const a = { b: 1 };
					log(a);
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},
			"deprecated const in property access chain": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					console.log(a.b);
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    18,
					},
				},
			},
			"deprecated const in optional property access": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					console.log(a?.b);
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    18,
					},
				},
			},
			"deprecated const in nested property access 1": {
				Code: `
					/** @deprecated */ const a = { b: { c: 1 } };
					a.b.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},
			"deprecated const in nested property access 2": {
				Code: `
					/** @deprecated */ const a = { b: { c: 1 } };
					a.b?.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},
			"deprecated const in nested optional property access": {
				Code: `
					/** @deprecated */ const a = { b: { c: 1 } };
					a?.b?.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    6,
					},
				},
			},

			// Using deprecated property
			"deprecated property access": {
				Code: `
					const a = {
						/** @deprecated */ b: { c: 1 },
					};
					a.b.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      5,
						Column:    8,
					},
				},
			},
			"deprecated property in type": {
				Code: `
					declare const a: {
						/** @deprecated */ b: { c: 1 };
					};
					a.b.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      5,
						Column:    8,
					},
				},
			},
			"deprecated const property access": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const c = a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    16,
					},
				},
			},
			"deprecated const in nested destructuring": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const { c } = a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    20,
					},
				},
			},
			"deprecated in object property": {
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
				Errors: []rule_tester.ExpectedError{
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
			"deprecated in shorthand property": {
				Code: `
					/** @deprecated */
					declare const test: string;
					const bar = {
						test,
					};
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      5,
						Column:    8,
					},
				},
			},
			"deprecated const in destructuring with default": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const { c = 'd' } = a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    26,
					},
				},
			},
			"deprecated const in destructuring with rename": {
				Code: `
					/** @deprecated */ const a = { b: 1 };
					const { c: d } = a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    23,
					},
				},
			},
			"deprecated in array literal": {
				Code: `
					/** @deprecated */
					declare const a: string[];
					const [b] = [a];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    19,
					},
				},
			},

			// Using deprecated class
			"deprecated class instantiation": {
				Code: `
					/** @deprecated */
					class A {}
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},
			"deprecated export class instantiation": {
				Code: `
					/** @deprecated */
					export class A {}
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},
			"deprecated class expression": {
				Code: `
					/** @deprecated */
					const A = class {};
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},
			"deprecated declare class": {
				Code: `
					/** @deprecated */
					declare class A {}
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},
			"deprecated constructor": {
				Code: `
					const A = class {
						/** @deprecated */
						constructor() {}
					};
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    10,
					},
				},
			},
			"deprecated constructor overload": {
				Code: `
					const A = class {
						/** @deprecated */
						constructor();
						constructor(arg: string);
						constructor(arg?: string) {}
					};
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      8,
						Column:    10,
					},
				},
			},
			"deprecated constructor signature": {
				Code: `
					declare const A: {
						/** @deprecated */
						new (): string;
					};
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    10,
					},
				},
			},
			"deprecated class with non-deprecated constructor": {
				Code: `
					/** @deprecated */
					declare class A {
						constructor();
					}
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    10,
					},
				},
			},

			// Using deprecated class members
			"deprecated property in destructuring": {
				Code: `
					class A {
						/** @deprecated */
						b: string;
					}
					declare const a: A;
					const { b } = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated method access": {
				Code: `
					declare class A {
						/** @deprecated */
						b(): string;
					}
					declare const a: A;
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated method call": {
				Code: `
					declare class A {
						/** @deprecated */
						b(): string;
					}
					declare const a: A;
					a.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated property function access": {
				Code: `
					declare class A {
						/** @deprecated */
						b: () => string;
					}
					declare const a: A;
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated property function call": {
				Code: `
					declare class A {
						/** @deprecated */
						b: () => string;
					}
					declare const a: A;
					a.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated interface property function": {
				Code: `
					interface A {
						/** @deprecated */
						b: () => string;
					}
					declare const a: A;
					a.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated class method implementation": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      9,
						Column:    8,
					},
				},
			},
			"deprecated method overload with reason": {
				Code: `
					declare class A {
						/** @deprecated Use b(value). */
						b(): string;
						b(value: string): string;
					}
					declare const a: A;
					a.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      8,
						Column:    8,
					},
				},
			},
			"deprecated static property": {
				Code: `
					declare class A {
						/** @deprecated */
						static b: string;
					}
					A.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},
			"deprecated object type property": {
				Code: `
					declare const a: {
						/** @deprecated */
						b: string;
					};
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},
			"deprecated interface property": {
				Code: `
					interface A {
						/** @deprecated */
						b: string;
					}
					declare const a: A;
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated export interface property": {
				Code: `
					export interface A {
						/** @deprecated */
						b: string;
					}
					declare const a: A;
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated interface property in destructuring": {
				Code: `
					interface A {
						/** @deprecated */
						b: string;
					}
					declare const a: A;
					const { b } = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated type alias property in destructuring": {
				Code: `
					type A = {
						/** @deprecated */
						b: string;
					};
					declare const a: A;
					const { b } = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated export type property in destructuring": {
				Code: `
					export type A = {
						/** @deprecated */
						b: string;
					};
					declare const a: A;
					const { b } = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated return type property in destructuring": {
				Code: `
					type A = () => {
						/** @deprecated */
						b: string;
					};
					declare const a: A;
					const { b } = a();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated type in type annotation": {
				Code: `
					/** @deprecated */
					type A = string[];
					declare const a: A;
					const [b] = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    23,
					},
				},
			},

			// Using deprecated namespace members
			"deprecated namespace constant": {
				Code: `
					namespace A {
						/** @deprecated */
						export const b = '';
					}
					A.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},
			"deprecated export namespace constant": {
				Code: `
					export namespace A {
						/** @deprecated */
						export const b = '';
					}
					A.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},
			"deprecated namespace function": {
				Code: `
					namespace A {
						/** @deprecated */
						export function b() {}
					}
					A.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},
			"deprecated namespace function overload with reason": {
				Code: `
					namespace assert {
						export function fail(message?: string | Error): never;
						/** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */
						export function fail(actual: unknown, expected: unknown): never;
					}
					assert.fail({}, {});
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      7,
						Column:    13,
					},
				},
			},

			// Using deprecated enum
			"deprecated enum": {
				Code: `
					/** @deprecated */
					enum A {
						a,
					}
					A.a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    6,
					},
				},
			},
			"deprecated enum member": {
				Code: `
					enum A {
						/** @deprecated */
						a,
					}
					A.a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},

			// Using deprecated function
			"deprecated function call": {
				Code: `
					/** @deprecated */
					function a() {}
					a();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    6,
					},
				},
			},
			"deprecated function overload 1": {
				Code: `
					/** @deprecated */
					function a(): void;
					function a() {}
					a();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      5,
						Column:    6,
					},
				},
			},
			"deprecated function overload 2": {
				Code: `
					function a(): void;
					/** @deprecated */
					function a(value: string): void;
					function a(value?: string) {}
					a('');
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    6,
					},
				},
			},
			"deprecated function type": {
				Code: `
					type A = {
						(value: 'b'): void;
						/** @deprecated */
						(value: 'c'): void;
					};
					declare const foo: A;
					foo('c');
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      8,
						Column:    6,
					},
				},
			},
			"deprecated function parameter usage": {
				Code: `
					function a(
						/** @deprecated */
						b?: boolean,
					) {
						return b;
					}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    15,
					},
				},
			},
			"deprecated function parameter with reason": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      9,
						Column:    12,
					},
				},
			},
			"deprecated tagged template": {
				Code: `
					/** @deprecated */
					declare function a(...args: unknown[]): string;
					a\`\`;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Using deprecated JSX component
			"deprecated JSX function component": {
				Code: `
					/** @deprecated */
					const A = () => <div />;
					const a = <A />;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    17,
					},
				},
			},
			"deprecated JSX function component with closing tag": {
				Code: `
					/** @deprecated */
					const A = () => <div />;
					const a = <A></A>;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    17,
					},
				},
			},
			"deprecated JSX function declaration component": {
				Code: `
					/** @deprecated */
					function A() {
						return <div />;
					}
					const a = <A />;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    17,
					},
				},
			},
			"deprecated JSX function declaration component with closing tag": {
				Code: `
					/** @deprecated */
					function A() {
						return <div />;
					}
					const a = <A></A>;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    17,
					},
				},
			},

			// Using deprecated type
			"deprecated type in union": {
				Code: `
					/** @deprecated */
					export type A = string;
					export type B = string;
					export type C = string;
					export type D = A | B | C;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    22,
					},
				},
			},
			"deprecated namespace type in union": {
				Code: `
					namespace A {
						/** @deprecated */
						export type B = string;
						export type C = string;
						export type D = string;
					}
					export type D = A.B | A.C | A.D;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      8,
						Column:    24,
					},
				},
			},

			// Property destructuring with deprecated
			"deprecated property with default in destructuring": {
				Code: `
					interface Props {
						/** @deprecated */
						anchor: 'foo';
					}
					declare const x: Props;
					const { anchor = '' } = x;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},
			"deprecated property in nested destructuring": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      8,
						Column:    15,
					},
				},
			},
			"deprecated property in array destructuring": {
				Code: `
					interface Props {
						/** @deprecated */
						anchor: 'foo';
					}
					declare const x: [item: Props];
					const [{ anchor = 'bar' }] = x;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    15,
					},
				},
			},
			"deprecated property in nested destructuring with same name": {
				Code: `
					interface Props {
						/** @deprecated */
						foo: Props;
					}
					declare const x: Props;
					const { foo = x } = x;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    14,
					},
				},
			},

			// Using deprecated class accessor
			"deprecated accessor access": {
				Code: `
					declare class A {
						/** @deprecated */
						accessor b: () => string;
					}
					declare const a: A;
					a.b;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated accessor call": {
				Code: `
					declare class A {
						/** @deprecated */
						accessor b: () => string;
					}
					declare const a: A;
					a.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
			"deprecated class accessor implementation": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      9,
						Column:    8,
					},
				},
			},
			"deprecated static accessor": {
				Code: `
					declare class A {
						/** @deprecated */
						static accessor b: () => string;
					}
					A.b();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},

			// Using deprecated private identifier
			"deprecated private property": {
				Code: `
					class A {
						/** @deprecated */
						#b = () => {};
						c() {
							this.#b();
						}
					}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    15,
					},
				},
			},

			// Computed property access
			"deprecated computed property with string literal": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a['b'];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    18,
					},
				},
			},
			"deprecated computed property with const string": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const x = 'b';
					const c = a[x];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    18,
					},
				},
			},
			"deprecated computed property with number": {
				Code: `
					const a = {
						/** @deprecated */
						[2]: 'string',
					};
					const x = 'b';
					const c = a[2];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    18,
					},
				},
			},
			"deprecated computed property with const as": {
				Code: `
					const a = {
						/** @deprecated reason for deprecation */
						b: 'string',
					};
					const key = 'b';
					const stringKey = key as const;
					const c = a[stringKey];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      8,
						Column:    18,
					},
				},
			},
			"deprecated computed property with enum": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      10,
						Column:    18,
					},
				},
			},
			"deprecated computed property with template": {
				Code: `
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const key = \`b\`;
					const c = a[key];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    18,
					},
				},
			},
			"deprecated computed property with string variable": {
				Code: `
					const stringObj = 'b';
					const a = {
						/** @deprecated */
						b: 'string',
					};
					const c = a[stringObj];
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    18,
					},
				},
			},

			// Export with deprecated
			"export deprecated function": {
				Code: `
					import { deprecatedFunction } from './deprecated';
					export { deprecatedFunction };
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      3,
						Column:    15,
					},
				},
			},
			"export deprecated function from": {
				Code: `
					export { deprecatedFunction } from './deprecated';
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      2,
						Column:    15,
					},
				},
			},
			"export deprecated default as alias": {
				Code: `
					export { default as foo } from './deprecated';
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      2,
						Column:    26,
					},
				},
			},
			"export deprecated with alias": {
				Code: `
					export { deprecatedFunction as bar } from './deprecated';
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      2,
						Column:    37,
					},
				},
			},

			// Extends/implements with deprecated
			"implements deprecated interface": {
				Code: `
					/** @deprecated */
					interface Foo {}
					class Bar implements Foo {}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    27,
					},
				},
			},
			"export class implements deprecated": {
				Code: `
					/** @deprecated */
					interface Foo {}
					export class Bar implements Foo {}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    34,
					},
				},
			},
			"implements deprecated in list": {
				Code: `
					/** @deprecated */
					interface Foo {}
					interface Baz {}
					export class Bar implements Baz, Foo {}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      5,
						Column:    39,
					},
				},
			},
			"extends deprecated class": {
				Code: `
					/** @deprecated */
					class Foo {}
					export class Bar extends Foo {}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    31,
					},
				},
			},

			// Decorator with deprecated
			"deprecated decorator": {
				Code: `
					/** @deprecated */
					declare function decorator(constructor: Function);
					@decorator
					export class Foo {}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    7,
					},
				},
			},

			// Export default with deprecated
			"export default deprecated function call": {
				Code: `
					/** @deprecated */
					function a(): object {
						return {};
					}
					export default a();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    21,
					},
				},
			},

			// Super with deprecated
			"super call with deprecated constructor": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      9,
						Column:    8,
					},
				},
			},
			"super call with deprecated constructor and reason": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      9,
						Column:    8,
					},
				},
			},

			// JSX with deprecated aria attribute
			"JSX deprecated aria attribute": {
				Code: `const a = <div aria-grabbed></div>;`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      1,
						Column:    16,
					},
				},
			},

			// JSX with deprecated property in namespaced element
			"JSX deprecated property in namespaced element": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      13,
						Column:    56,
					},
				},
			},

			// JSX with deprecated property in component
			"JSX deprecated property in component": {
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
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      15,
						Column:    39,
					},
				},
			},
		},
	)
}
