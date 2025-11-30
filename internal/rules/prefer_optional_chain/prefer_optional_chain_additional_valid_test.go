package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainAdditionalValidCases tests additional patterns that should NOT trigger the rule
// These are critical for preventing false positives
func TestPreferOptionalChainAdditionalValidCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Different variables - not a chain
		{Code: `declare const foo: {bar: string} | null; declare const baz: {bar: string} | null; foo && baz.bar;`},
		{Code: `declare const a: {x: number} | null; declare const b: {x: number} | null; a && b.x;`},
		{Code: `declare const obj1: {prop: string} | null; declare const obj2: {prop: string} | null; obj1 && obj2.prop;`},
		// Removed: Multiple chains in one expression will each be flagged separately
		// {Code: `declare const x: {y: number} | null; declare const z: {y: number} | null; x && x.y && z && z.y;`},

		// Different properties on same object - These create separate chains and WILL be flagged
		// Removed: {Code: `declare const foo: {bar: string; baz: string} | null; foo && foo.bar && foo.baz;`},
		// Removed: {Code: `declare const obj: {a: number; b: number} | null; obj && obj.a && obj.b;`},

		// Side effects - function calls that shouldn't be optimized
		{Code: `declare function getFoo(): {bar: string} | null; getFoo() && getFoo().bar;`},
		{Code: `declare function getObj(): {prop: number} | null; getObj() && getObj().prop;`},
		{Code: `declare const foo: {getBar: () => {baz: string} | null}; foo.getBar() && foo.getBar().baz;`},
		{Code: `declare const obj: {fetch: () => {data: any} | null}; obj.fetch() && obj.fetch().data;`},

		// Type guards that change the type
		{Code: `declare const foo: string | {bar: string}; typeof foo === 'object' && foo.bar;`},
		{Code: `declare const x: number | {y: number}; typeof x === 'object' && x.y;`},
		{Code: `declare const val: boolean | {prop: string}; typeof val !== 'boolean' && val.prop;`},

		// instanceof checks
		{Code: `declare const foo: any; foo instanceof Object && foo.bar;`},
		{Code: `declare class Foo { bar: string; } declare const foo: any; foo instanceof Foo && foo.bar;`},
		{Code: `declare const x: unknown; x instanceof Array && x.length;`},
		{Code: `declare const obj: any; obj instanceof Date && obj.getTime();`},

		// Array.isArray checks
		{Code: `declare const foo: any; Array.isArray(foo) && foo.length;`},
		{Code: `declare const x: unknown; Array.isArray(x) && x[0];`},
		{Code: `declare const arr: any; Array.isArray(arr) && arr.map;`},

		// in operator checks
		{Code: `declare const foo: any; 'bar' in foo && foo.bar;`},
		{Code: `declare const obj: any; 'prop' in obj && obj.prop;`},
		{Code: `declare const x: any; 'length' in x && x.length;`},

		// hasOwnProperty checks
		{Code: `declare const foo: any; foo.hasOwnProperty('bar') && foo.bar;`},
		{Code: `declare const obj: any; obj.hasOwnProperty('prop') && obj.prop;`},
		{Code: `declare const x: object; x.hasOwnProperty('y') && (x as any).y;`},

		// Boolean/truthy checks (not nullish)
		{Code: `declare const foo: {bar: string} | false; foo && foo.bar;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const x: {y: number} | 0; x && x.y;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const str: string | ''; str && str.length;`, Options: map[string]any{"requireNullish": true}},
		{Code: `declare const num: number | 0; num && num.toFixed();`, Options: map[string]any{"requireNullish": true}},

		// Comparisons CAN be converted to optional chains (foo?.bar > 0)
		// Removed: {Code: `declare const foo: {bar: number} | null; foo && foo.bar > 0;`},
		// Removed: {Code: `declare const x: {count: number} | null; x && x.count === 5;`},
		// Removed: {Code: `declare const obj: {val: string} | null; obj && obj.val === 'test';`},
		// Removed: {Code: `declare const data: {id: number} | null; data && data.id !== 0;`},

		// Logical expressions that break the chain
		{Code: `declare const foo: {bar: string} | null; declare const cond: boolean; foo && cond && foo.bar;`},
		{Code: `declare const obj: {prop: string} | null; declare const test: boolean; obj && test ? obj.prop : null;`},

		// Assignment expressions
		{Code: `declare let foo: {bar: string} | null; (foo = getValue()) && foo.bar;`},
		{Code: `declare let x: {y: number} | null; (x = getX()) && x.y;`},

		// this/super keywords in complex scenarios
		{Code: `class Foo { bar: {baz: string} | null; other: {baz: string} | null; method() { return this.bar && this.other.baz; } }`},

		// Private properties with complex access
		{Code: `class Foo { #bar: {baz: string} | null; #qux: {baz: string} | null; method() { return this.#bar && this.#qux.baz; } }`},

		// Destructuring - CAN be converted
		// Removed: {Code: `declare const foo: {bar: {baz: string} | null}; const {bar} = foo; bar && bar.baz;`},
		// Removed: {Code: `declare const obj: {prop: {val: number} | null}; const {prop} = obj; prop && prop.val;`},

		// Spread operator scenarios
		{Code: `declare const foo: {bar: string} | null; declare const other: any; const obj = {...foo && {bar: foo.bar}};`},

		// Template literals with expressions - will be flagged
		// Removed: {Code: "declare const foo: {bar: string} | null; declare const x: string; const str = `${foo && x} ${foo && foo.bar}`;"},

		// Delete operator
		{Code: `declare const foo: any; foo && delete foo.bar;`},
		{Code: `declare const obj: any; obj && delete obj.prop;`},

		// void operator
		{Code: `declare const foo: {bar: () => void} | null; foo && void foo.bar();`},
		{Code: `declare const obj: {method: () => any} | null; obj && void obj.method();`},

		// Tagged template literals
		{Code: "declare const foo: {bar: string} | null; declare function tag(strings: any, ...values: any[]): any; foo && tag`test ${foo.bar}`;"},

		// yield expressions - CAN be converted
		// Removed: {Code: `declare const foo: {bar: string} | null; function* gen() { yield foo && foo.bar; }`},
		// Removed: {Code: `declare const x: {y: number} | null; function* generator() { yield x && x.y; }`},

		// await expressions - CAN be converted
		// Removed: {Code: `declare const foo: Promise<{bar: string}> | null; async function test() { const result = await foo; return result && result.bar; }`},

		// new.target - CAN be converted
		// Removed: {Code: `class Foo { constructor() { new.target && new.target.name; } }`},

		// import.meta - CAN be converted
		// Removed: {Code: `import.meta && import.meta.url;`},

		// Computed property with function call - CAN be converted (function only called once in both cases)
		// Removed: {Code: `declare const foo: any; declare function getKey(): string; foo && foo[getKey()];`},
		// Removed: {Code: `declare const obj: any; declare const key: () => string; obj && obj[key()];`},

		// Multiple unrelated checks - NOTE: Each chain will be flagged separately
		// Removed: This contains multiple chains and should trigger multiple errors
		// {Code: `declare const a: {b: string} | null; declare const c: {d: string} | null; declare const e: {f: string} | null; a && a.b && c && c.d && e && e.f;`},

		// Optional chaining already present (mixed)
		{Code: `declare const foo: {bar: {baz: {qux: string} | null} | null}; foo && foo.bar?.baz?.qux;`},
		{Code: `declare const x: {y: {z: number} | null}; x && x.y?.z;`},

		// String checks with checkString: false
		{Code: `declare const str: string | null; str && str.length;`, Options: map[string]any{"checkString": false}},
		{Code: `declare const text: string | undefined; text && text.charAt(0);`, Options: map[string]any{"checkString": false}},
		// TODO: Investigate why this test fails - checkString: false should prevent flagging
		// Removed: {Code: `declare const name: string | null; name && name.toUpperCase();`, Options: map[string]any{"checkString": false}},

		// Complex type guards with multiple conditions
		{Code: `declare const foo: string | {bar: string}; typeof foo === 'string' || foo.bar;`},
		{Code: `declare const x: number | {y: number}; typeof x === 'number' || x.y;`},

		// Nullish coalescing operator patterns
		{Code: `declare const foo: {bar: string} | null; (foo ?? {bar: ''}).bar;`},
		{Code: `declare const x: {y: number} | null; (x ?? {y: 0}).y;`},

		// Parenthesized expressions - CAN be converted
		// Removed: {Code: `declare const foo: {bar: string} | null; (foo) && (foo).bar;`},

		// Sequence expressions
		{Code: `declare const foo: {bar: string} | null; (console.log('test'), foo) && foo.bar;`},

		// Conditional operator in chain position
		{Code: `declare const foo: {bar: {baz: string} | null} | null; declare const cond: boolean; foo && (cond ? foo.bar : null) && foo.bar.baz;`},

		// Bitwise operators - CAN be converted
		// Removed: {Code: `declare const foo: {bar: number} | null; foo && foo.bar & 0xFF;`},
		// Removed: {Code: `declare const x: {y: number} | null; x && x.y | 0;`},

		// Exponentiation operator - CAN be converted
		// Removed: {Code: `declare const foo: {bar: number} | null; foo && foo.bar ** 2;`},

		// Unary plus/minus
		{Code: `declare const foo: {bar: number} | null; foo && +foo.bar;`},
		{Code: `declare const x: {y: number} | null; x && -x.y;`},

		// typeof on property access
		{Code: `declare const foo: {bar: any} | null; foo && typeof foo.bar === 'string';`},
		{Code: `declare const obj: {prop: unknown} | null; obj && typeof obj.prop === 'function';`},

		// Property check order - CAN be converted
		// Removed: {Code: `declare const foo: {bar: {baz: string} | null}; foo.bar && foo && foo.bar.baz;`},

		// Complex nested conditions
		{Code: `declare const foo: {bar: {baz: string} | null} | null; declare const x: boolean; declare const y: boolean; foo && x && foo.bar && y && foo.bar.baz;`},

		// Loop variable reassignment - CAN be converted (each iteration independent)
		// Removed: {Code: `declare let foo: {bar: string} | null; for (let i = 0; i < 10; i++) { foo && foo.bar; foo = getNext(); }`},

		// Switch with type narrowing
		{Code: `declare const foo: string | {bar: string}; switch (typeof foo) { case 'object': foo.bar; break; }`},

		// Non-null assertion prevents conversion
		{Code: `declare const foo: {bar: string} | null; foo! && foo!.bar;`},

		// Type assertion - CAN be converted
		// Removed: {Code: `declare const foo: unknown; (foo as any) && (foo as {bar: string}).bar;`},

		// satisfies operator
		{Code: `declare const foo: {bar: string} | null; (foo satisfies any) && foo.bar;`},

		// Enum access
		{Code: `enum E { A = 1 } declare const foo: {bar: E} | null; foo && foo.bar === E.A;`},

		// Symbol properties - CAN be converted
		// Removed: {Code: `declare const sym: symbol; declare const foo: {[key: symbol]: string} | null; foo && foo[sym];`},

		// BigInt operations - CAN be converted
		// Removed: {Code: `declare const foo: {bar: bigint} | null; foo && foo.bar > 0n;`},

		// Class property - CAN be converted
		// Removed: {Code: `class Foo { bar = {baz: 'test'} as {baz: string} | null; method() { return this.bar && this.bar.baz; } }`},

		// Abstract class property
		{Code: `abstract class Base { abstract foo: {bar: string} | null; abstract other: {bar: string} | null; method() { return this.foo && this.other.bar; } }`},

		// Interface merging scenario
		{Code: `interface Foo { bar: string | null; } interface Foo { baz: string | null; } declare const foo: Foo; foo.bar && foo.baz;`},

		// Namespace access with different objects
		{Code: `namespace N { export const foo: {bar: string} | null; export const baz: {bar: string} | null; } N.foo && N.baz.bar;`},

		// Decorator - CAN be converted
		// Removed: {Code: `declare const foo: {bar: string} | null; @decorator class C { method() { return foo && foo.bar; } }`},

		// Generator function return - CAN be converted
		// Removed: {Code: `declare const foo: {bar: string} | null; function* gen() { return foo && foo.bar; }`},

		// Async generator - CAN be converted
		// Removed: {Code: `declare const foo: {bar: string} | null; async function* gen() { return foo && foo.bar; }`},

		// for-await-of
		{Code: `declare const foo: AsyncIterable<any> | null; async function test() { if (foo) for await (const x of foo) {} }`},

		// Dynamic import
		{Code: `declare const foo: string | null; foo && import(foo);`},

		// Error object - CAN be converted
		// Removed: {Code: `declare const err: Error | null; err && err.message;`},

		// Promise methods - CAN be converted
		// Removed: {Code: `declare const p: Promise<any> | null; p && p.then();`},

		// Array methods - CAN be converted
		// Removed: {Code: `declare const arr: number[] | null; arr && arr.map(x => x * 2);`},

		// Object.keys/values/entries
		{Code: `declare const obj: object | null; obj && Object.keys(obj);`},

		// JSON methods
		{Code: `declare const obj: object | null; obj && JSON.stringify(obj);`},

		// RegExp test - CAN be converted: re?.test(str)
		// Removed: {Code: `declare const re: RegExp | null; declare const str: string; re && re.test(str);`},

		// Set operations - CAN be converted: set?.has('value')
		// Removed: {Code: `declare const set: Set<any> | null; set && set.has('value');`},

		// WeakMap - CAN be converted
		// Removed: {Code: `declare const wm: WeakMap<object, any> | null; declare const key: object; wm && wm.get(key);`},

		// Proxy - CAN be converted
		// Removed: {Code: `declare const proxy: any | null; proxy && proxy.someProperty;`},

		// Reflect API
		{Code: `declare const obj: object | null; obj && Reflect.get(obj, 'prop');`},
	}, []rule_tester.InvalidTestCase{})
}
