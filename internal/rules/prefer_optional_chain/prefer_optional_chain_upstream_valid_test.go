package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamValidCases tests cases that should NOT trigger the prefer-optional-chain rule
// Source: https://github.com/typescript-eslint/typescript-eslint/.../prefer-optional-chain.test.ts
// These ensure the rule doesn't produce false positives
func TestUpstreamValidCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{
			// Different variables - not a chain
			{Code: `!a || !b;`},
			{Code: `!a || a.b;`},
			{Code: `!a && a.b;`},
			{Code: `!a && !a.b;`},

			// Already using optional chaining
			{Code: `!a.b || a.b?.();`},
			{Code: `!a.b || a.b();`},

			// Assignment operators
			{Code: `foo ||= bar;`},
			{Code: `foo ||= bar?.baz;`},
			{Code: `foo ||= bar?.baz?.buzz;`},

			// Different variables (not chaining)
			{Code: `foo && bar;`},
			{Code: `foo && foo;`},
			{Code: `foo || bar;`},
			{Code: `foo ?? bar;`},
			{Code: `foo || foo.bar;`},
			{Code: `foo ?? foo.bar;`},

			// Different property/method calls
			{Code: `file !== 'index.ts' && file.endsWith('.ts');`},
			{Code: `nextToken && sourceCode.isSpaceBetweenTokens(prevToken, nextToken);`},
			{Code: `result && this.options.shouldPreserveNodeMaps;`},
			{Code: `foo && fooBar.baz;`},              // fooBar is different from foo
			{Code: `match && match$1 !== undefined;`}, // match$1 is different from match

			// Type checks that don't form chains
			{Code: `typeof foo === 'number' && foo.toFixed();`},
			{Code: `foo === 'undefined' && foo.length;`},
			{Code: `foo == bar && foo.bar == null;`},
			{Code: `foo === 1 && foo.toFixed();`},

			// Call arguments differ
			{Code: `foo.bar(a) && foo.bar(a, b).baz;`},

			// Type parameters differ
			{Code: `foo.bar<a>() && foo.bar<a, b>().baz;`},

			// Array literals differ
			{Code: `[1, 2].length && [1, 2, 3].length.toFixed();`},
			{Code: `[1,].length && [1, 2].length.toFixed();`},

			// Short-circuiting chains (already has optional)
			{Code: `(foo?.a).b && foo.a.b.c;`},
			{Code: `(foo?.a)() && foo.a().b;`},
			{Code: `(foo?.a)() && foo.a()();`},

			// Strict nullish checks (not a chain, just redundant checks)
			{Code: `foo !== null && foo !== undefined;`},
			{Code: `x['y'] !== undefined && x['y'] !== null;`},

			// Private properties (different from each other)
			{Code: `this.#a && this.#b;`},
			{Code: `!this.#a || !this.#b;`},
			{Code: `a.#foo?.bar;`},
			{Code: `!a.#foo?.bar;`},
			{Code: `!foo().#a || a;`},
			{Code: `!a.b.#a || a;`},
			{Code: `!new A().#b || a;`},
			{Code: `!(await a).#b || a;`},
			{Code: `!(foo as any).bar || 'anything';`},

			// Computed properties differ
			{Code: `!foo[1 + 1] || !foo[1 + 2];`},
			{Code: `!foo[1 + 1] || !foo[1 + 2].foo;`},

			// Side effects or different expressions
			{Code: `foo() && foo().bar;`},         // foo() called twice may have side effects
			{Code: `foo.bar() && foo.bar().baz;`}, // foo.bar() called twice may have side effects

			// Complex non-chain patterns
			{Code: `foo && bar && baz;`}, // all different
			{Code: `foo || bar || baz;`}, // all different
			{Code: `foo ?? bar ?? baz;`}, // all different

			// Mixed operators that don't form chains
			{Code: `foo && bar || baz;`},
			{Code: `foo || bar && baz;`},

			// ternary operators
			{Code: `foo ? foo.bar : baz;`},
			{Code: `foo ? bar : foo.bar;`},

			// Function boundaries
			{Code: `foo && (() => foo.bar)();`},
			{Code: `foo || (() => foo.bar)();`},

			// Object/array literals
			{Code: `foo && { bar: foo.baz };`},
			{Code: `foo && [foo.bar];`},

			// typeof guards that are meaningful
			{Code: `typeof foo === 'object' && foo !== null && foo.bar;`},
			{Code: `typeof foo === 'function' && foo();`},

			// in operator
			{Code: `'bar' in foo && foo.bar;`},

			// instanceof
			{Code: `foo instanceof Bar && foo.baz;`},

			// Array.isArray
			{Code: `Array.isArray(foo) && foo.length;`},
			{Code: `Array.isArray(foo) && foo[0];`},

			// Boolean() checks
			{Code: `Boolean(foo) && foo.bar;`},

			// Parenthesized expressions that break chains
			{Code: `(foo && bar) || (baz && qux);`},

			// Logical expressions in different contexts
			{Code: `if (foo && bar) { baz; }`},     // different vars
			{Code: `while (foo && bar) { baz; }`},  // different vars
			{Code: `for (; foo && bar;) { baz; }`}, // different vars

			// Return statements
			{Code: `return foo && bar;`},   // different vars
			{Code: `return !foo || !bar;`}, // different vars

			// Variable declarations
			{Code: `const x = foo && bar;`}, // different vars
			{Code: `let y = foo || bar;`},   // different vars

			// Arrow function returns
			{Code: `() => foo && bar;`},   // different vars
			{Code: `() => !foo || !bar;`}, // different vars

			// this keyword - currently not handled
			{Code: `this && this.foo;`},
			{Code: `!this || !this.foo;`},

			// Non-null assertion operator prevents conversion
			{Code: `!entity.__helper!.__initialized || options.refresh;`},

			// import.meta - special cases
			{Code: `import.meta || true;`},
			{Code: `import.meta || import.meta.foo;`},
			{Code: `!import.meta && false;`},
			{Code: `!import.meta && !import.meta.foo;`},

			// new.target
			{Code: `new.target || new.target.length;`},
			{Code: `!new.target || true;`},

			// Direct optional chaining on private properties (TS limitation)
			{Code: `foo && foo.#bar;`},
			{Code: `!foo || !foo.#bar;`},

			// Non-constant expressions (weird cases)
			{Code: `({}) && {}.toString();`},
			{Code: `[] && [].length;`},
			{Code: `(() => {}) && (() => {}).name;`},
			{Code: `(function () {}) && function () {}.name;`},
			{Code: `(class Foo {}) && class Foo {}.constructor;`},
			{Code: `new Map().get('a') && new Map().get('a').what;`},

			// Property check that isn't a chain
			{Code: `data && data.value !== null;`},

			// JSX elements
			{Code: `<div /> && (<div />).wtf;`},
			{Code: `<></> && (<></>).wtf;`},

			// Side effects in computed properties
			{Code: `foo[x++] && foo[x++].bar;`},
			{Code: `foo[yield x] && foo[yield x].bar;`},

			// Assignment with side effects
			{Code: `a = b && (a = b).wtf;`},

			// Complex parenthesized expressions
			{Code: `(x || y) != null && (x || y).foo;`},
			{Code: `(await foo) && (await foo).bar;`},

			// Non-nullish property check
			{Code: `declare const foo: { bar: string } | null; foo !== null && foo.bar !== null;`},
			{Code: `declare const foo: { bar: string | null } | null; foo != null && foo.bar !== null;`},

			// requireNullish option - prevents conversion without explicit check
			{Code: `declare const x: string; x && x.length;`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: string; foo && foo.toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const x: string | number | boolean | object; x && x.toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: { bar: string }; foo && foo.bar && foo.bar.toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: string; foo && foo.toString() && foo.toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: { bar: string }; foo && foo.bar && foo.bar.toString() && foo.bar.toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo1: { bar: string | null }; foo1 && foo1.bar;`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: string; (foo || {}).toString();`, Options: map[string]any{"requireNullish": true}},
			{Code: `declare const foo: string | null; (foo || 'a' || {}).toString();`, Options: map[string]any{"requireNullish": true}},

			// checkAny option
			{Code: `declare const x: any; x && x.length;`, Options: map[string]any{"checkAny": false}},

			// checkBigInt option
			{Code: `declare const x: bigint; x && x.length;`, Options: map[string]any{"checkBigInt": false}},

			// checkBoolean option
			{Code: `declare const x: boolean; x && x.length;`, Options: map[string]any{"checkBoolean": false}},

			// checkNumber option
			{Code: `declare const x: number; x && x.length;`, Options: map[string]any{"checkNumber": false}},

			// checkString option
			{Code: `declare const x: string; x && x.length;`, Options: map[string]any{"checkString": false}},

			// checkUnknown option
			{Code: `declare const x: unknown; x && x.length;`, Options: map[string]any{"checkUnknown": false}},

			// Assignment in check
			{Code: `(x = {}) && (x.y = true) != null && x.y.toString();`},

			// Template literal types
			{Code: `('x' as ${'x'}) && ('x' as ${'x'}).length;`},
			{Code: "`x` && `x`.length;"},
			{Code: "`x${a}` && `x${a}`.length;"},

			// Falsy unions with requireNullish
			{Code: `declare const x: false | { a: string }; x && x.a;`},
			{Code: `declare const x: false | { a: string }; !x || x.a;`},
			{Code: `declare const x: '' | { a: string }; x && x.a;`},
			{Code: `declare const x: '' | { a: string }; !x || x.a;`},
			{Code: `declare const x: 0 | { a: string }; x && x.a;`},
			{Code: `declare const x: 0 | { a: string }; !x || x.a;`},
			{Code: `declare const x: 0n | { a: string }; x && x.a;`},
			{Code: `declare const x: 0n | { a: string }; !x || x.a;`},

			// globalThis check
			{Code: `typeof globalThis !== 'undefined' && globalThis.Array();`},

			// void union
			{Code: `declare const x: void | (() => void); x && x();`},

			// || {} pattern - valid cases (not convertible)
			{Code: `foo || {};`},
			{Code: `foo || ({} as any);`},
			{Code: `(foo || {})?.bar;`},
			{Code: `(foo || { bar: 1 }).bar;`},
			{Code: `(undefined && (foo || {})).bar;`},
			{Code: `foo ||= bar || {};`},
			{Code: `foo ||= bar?.baz || {};`},
			{Code: `(foo1 ? foo2 : foo3 || {}).foo4;`},
			{Code: `(foo = 2 || {}).bar;`},
			{Code: `func(foo || {}).bar;`},
			{Code: `foo ?? {};`},
			{Code: `(foo ?? {})?.bar;`},
			{Code: `foo ||= bar ?? {};`},

			// Issue #8380 - strict nullish checks that aren't chains
			{Code: `const a = null; const b = 0; a === undefined || b === null || b === undefined;`},
			{Code: `const a = 0; const b = 0; a === undefined || b === undefined || b === null;`},
			{Code: `const a = 0; const b = 0; b === null || a === undefined || b === undefined;`},
			{Code: `const b = 0; b === null || b === undefined;`},
			{Code: `const a = 0; const b = 0; b != null && a !== null && a !== undefined;`},

			// Ending with comparison - valid cases (comparison to non-nullish values)
			{Code: `foo && foo.bar == undeclaredVar;`},
			{Code: `foo && foo.bar == null;`},
			{Code: `foo && foo.bar == undefined;`},
			{Code: `foo && foo.bar === undeclaredVar;`},
			{Code: `foo && foo.bar === undefined;`},
			{Code: `foo && foo.bar !== 0;`},
			{Code: `foo && foo.bar !== 1;`},
			{Code: `foo && foo.bar !== '123';`},
			{Code: `foo && foo.bar !== {};`},
			{Code: `foo && foo.bar !== false;`},
			{Code: `foo && foo.bar !== true;`},
			{Code: `foo && foo.bar !== null;`},
			{Code: `foo && foo.bar !== undeclaredVar;`},
			{Code: `foo && foo.bar != 0;`},
			{Code: `foo && foo.bar != 1;`},
			{Code: `foo && foo.bar != '123';`},
			{Code: `foo && foo.bar != {};`},
			{Code: `foo && foo.bar != false;`},
			{Code: `foo && foo.bar != true;`},
			{Code: `foo && foo.bar != undeclaredVar;`},

			// With explicit nullish checks
			{Code: `foo != null && foo.bar == undeclaredVar;`},
			{Code: `foo != null && foo.bar == null;`},
			{Code: `foo != null && foo.bar == undefined;`},
			{Code: `foo != null && foo.bar === undeclaredVar;`},
			{Code: `foo != null && foo.bar === undefined;`},
			{Code: `foo != null && foo.bar !== 0;`},
			{Code: `foo != null && foo.bar !== 1;`},
			{Code: `foo != null && foo.bar !== '123';`},
			{Code: `foo != null && foo.bar !== {};`},
			{Code: `foo != null && foo.bar !== false;`},
			{Code: `foo != null && foo.bar !== true;`},
			{Code: `foo != null && foo.bar !== null;`},
			{Code: `foo != null && foo.bar !== undeclaredVar;`},
			{Code: `foo != null && foo.bar != 0;`},
			{Code: `foo != null && foo.bar != 1;`},
			{Code: `foo != null && foo.bar != '123';`},
			{Code: `foo != null && foo.bar != {};`},
			{Code: `foo != null && foo.bar != false;`},
			{Code: `foo != null && foo.bar != true;`},
			{Code: `foo != null && foo.bar != undeclaredVar;`},

			// With type declarations
			{Code: `declare const foo: { bar: number }; foo && foo.bar == undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar == null;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar == undefined;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar === undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== 0;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== '123';`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== {};`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== false;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== true;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== null;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar !== undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != 0;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != 1;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != '123';`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != {};`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != false;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != true;`},
			{Code: `declare const foo: { bar: number }; foo && foo.bar != undeclaredVar;`},

			// With explicit nullish checks and types
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar == undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar == null;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar == undefined;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar === undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar === undefined;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== 0;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== 1;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== '123';`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== {};`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== false;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== true;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== null;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar !== undeclaredVar;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != 0;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != 1;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != '123';`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != {};`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != false;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != true;`},
			{Code: `declare const foo: { bar: number }; foo != null && foo.bar != undeclaredVar;`},
		},
		[]rule_tester.InvalidTestCase{})
}
