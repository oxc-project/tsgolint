package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestValidCasesExtensive tests extensive cases that should NOT trigger the rule
// These are critical for preventing false positives
func TestValidCasesExtensive(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{
			// === Different variable names (not chaining) ===
			{Code: `a && b;`},
			{Code: `a && b && c;`},
			{Code: `a && b && c && d;`},
			{Code: `foo && bar && baz;`},
			{Code: `x || y;`},
			{Code: `x || y || z;`},
			{Code: `a ?? b;`},
			{Code: `a ?? b ?? c;`},

			// Mixed different variables
			{Code: `a && b || c;`},
			{Code: `a || b && c;`},
			{Code: `a && b || c && d;`},
			{Code: `(a && b) || (c && d);`},
			{Code: `(a || b) && (c || d);`},

			// === Different properties on different objects ===
			{Code: `foo.x && bar.y;`},
			{Code: `foo.a && bar.b && baz.c;`},
			{Code: `obj1.prop && obj2.prop;`}, // same prop name, different objects
			{Code: `a.b && c.d;`},
			{Code: `x.y.z && a.b.c;`},

			// === Different methods ===
			{Code: `foo.method1() && foo.method2();`}, // different methods
			{Code: `foo.bar() && foo.baz();`},
			{Code: `obj.get() && obj.set();`},
			{Code: `arr.push(x) && arr.pop();`},

			// === Type guards that change meaning ===
			{Code: `typeof x === 'string' && x.length;`},
			{Code: `typeof x === 'number' && x.toFixed();`},
			{Code: `typeof x === 'boolean' && x.valueOf();`},
			{Code: `typeof x === 'function' && x();`},
			{Code: `typeof x === 'symbol' && x.toString();`},
			{Code: `typeof x === 'bigint' && x > 0n;`},

			// === instanceof checks ===
			{Code: `x instanceof Array && x.length;`},
			{Code: `x instanceof Date && x.getTime();`},
			{Code: `x instanceof Error && x.message;`},
			{Code: `x instanceof RegExp && x.test('');`},
			{Code: `x instanceof Map && x.size;`},
			{Code: `x instanceof Set && x.has(1);`},
			{Code: `x instanceof Promise && x.then();`},

			// === Array.isArray checks ===
			{Code: `Array.isArray(x) && x.length;`},
			{Code: `Array.isArray(x) && x[0];`},
			{Code: `Array.isArray(x) && x.map(fn);`},
			{Code: `Array.isArray(x) && x.filter(fn);`},

			// === in operator ===
			{Code: `'prop' in obj && obj.prop;`},
			{Code: `'length' in x && x.length;`},
			{Code: `'then' in x && x.then();`},
			{Code: `0 in arr && arr[0];`},

			// === hasOwnProperty checks ===
			{Code: `obj.hasOwnProperty('prop') && obj.prop;`},
			{Code: `Object.prototype.hasOwnProperty.call(obj, 'key') && obj.key;`},

			// === Boolean() / !! checks ===
			{Code: `Boolean(x) && x.prop;`},
			{Code: `!!x && x.prop;`},
			{Code: `!x || x.prop;`}, // Valid: !x checks ALL falsy values, not just null/undefined

			// === Comparison checks that aren't nullish ===
			{Code: `x === 0 && x.toString();`},
			{Code: `x === '' && x.length;`},
			{Code: `x === false && x.valueOf();`},
			{Code: `x !== 0 && x.toFixed();`},
			{Code: `x !== '' && x.split('');`},
			{Code: `x !== false && x.toString();`},
			{Code: `x > 0 && x.toFixed();`},
			{Code: `x < 100 && x.toString();`},
			{Code: `x >= 0 && x.valueOf();`},
			{Code: `x <= 100 && x.toFixed();`},

			// === Equality checks between different things ===
			{Code: `x === y && x.prop;`}, // x === y is not a nullish check
			{Code: `a === b && a.method();`},
			{Code: `foo !== bar && foo.baz;`},

			// === Different element access ===
			{Code: `foo[a] && foo[b];`}, // different indices
			{Code: `arr[0] && arr[1];`},
			{Code: `obj['x'] && obj['y'];`},
			{Code: `map.get(a) && map.get(b);`},

			// === Different template literals ===
			{Code: "obj[`key1`] && obj[`key2`];"},
			{Code: "obj[`${a}`] && obj[`${b}`];"},

			// === Call expressions with different arguments ===
			{Code: `fn(1) && fn(2);`},
			{Code: `obj.method(a) && obj.method(b);`},
			{Code: `foo.bar(x, y) && foo.bar(x, y, z);`},

			// === Different type parameters ===
			{Code: `foo<string>() && foo<number>();`},
			{Code: `bar<A>() && bar<A, B>();`},

			// === Side effects (functions called multiple times) ===
			{Code: `getUser() && getUser().name;`}, // getUser() may return different values
			{Code: `fetch() && fetch().json();`},
			{Code: `Math.random() && Math.random().toFixed();`},
			{Code: `new Date() && new Date().getTime();`},

			// === Expressions that evaluate differently ===
			{Code: `x++ && x;`}, // x++ changes x
			{Code: `++x && x;`},
			{Code: `x-- && x;`},
			{Code: `--x && x;`},

			// === Assignment expressions ===
			{Code: `(x = foo) && x.bar;`}, // assignment is side effect
			{Code: `(x += 1) && x.toFixed();`},
			{Code: `(x ||= y) && x.prop;`},
			{Code: `(x &&= y) && x.prop;`},
			{Code: `(x ??= y) && x.prop;`},

			// === this keyword patterns that shouldn't convert ===
			{Code: `this && this.foo;`},      // Can't chain from 'this'
			{Code: `!this || !this.foo;`},    // Can't chain from 'this'
			{Code: `this.a && this.b;`},      // Different properties, both non-null
			{Code: `this.x || this.y;`},      // Different properties
			{Code: `this.foo && other.bar;`}, // Different objects
			{Code: `other.foo && this.bar;`}, // Different objects

			// === super keyword ===
			{Code: `super.foo && super.bar;`}, // Different properties
			{Code: `super.x && other.y;`},     // Different objects

			// === new.target ===
			// Removed - this SHOULD convert with && operator
			// {Code: `new.target && new.target.name;`},

			// === import.meta ===
			{Code: `import.meta || true;`},
			{Code: `import.meta || import.meta.url;`}, // Can't convert
			{Code: `!import.meta && false;`},
			{Code: `!import.meta && !import.meta.url;`}, // Can't convert

			// === Private properties ===
			{Code: `foo && foo.#bar;`}, // Direct optional chaining not supported on private
			{Code: `!foo || !foo.#bar;`},
			{Code: `this.#a && this.#b;`}, // Different private properties
			{Code: `obj.#x || obj.#y;`},   // Different private properties
			{Code: `foo().#bar && something;`},

			// === Tagged template literals ===
			// Removed - this SHOULD convert
			// {Code: "tag`template` && tag`template`.prop;"},
			{Code: "foo.tag`a` && foo.tag`b`;"}, // Different tags

			// === Sequence expressions ===
			{Code: `(a, b) && (a, b).prop;`},
			{Code: `(x, y, z) && result;`},

			// === Delete operator ===
			{Code: `delete obj.a && obj.b;`},

			// === void operator ===
			{Code: `void foo && bar;`},
			{Code: `void 0 === x && x.prop;`},

			// === Comma operator in unusual places ===
			{Code: `(foo, bar) && baz;`},

			// === Complex parenthesized expressions ===
			{Code: `((a)) && ((b));`},                   // Different vars
			{Code: `((foo && bar)) || ((baz && qux));`}, // Nested different vars
			{Code: `((a.b)) && ((c.d));`},               // Different objects

			// === Spread in different contexts ===
			{Code: `[...arr] && result;`},
			{Code: `{...obj} && result;`},

			// === Destructuring patterns ===
			// Removed - these SHOULD convert
			// {Code: `const {a} = foo && foo.bar;`},
			// {Code: `const [b] = foo && foo.bar;`},

			// === for...in and for...of ===
			// Removed - these SHOULD convert
			// {Code: `for (const key in obj && obj.props) {}`},
			// {Code: `for (const val of arr && arr.items) {}`},

			// === yield expressions ===
			{Code: `yield foo && bar;`}, // Different vars
			{Code: `yield* foo && bar;`},

			// === await expressions (already awaited) ===
			{Code: `await foo && bar;`},   // Different vars
			{Code: `await a && await b;`}, // Two different awaits

			// === class expressions ===
			{Code: `(class {}) && result;`},
			{Code: `(class Foo {}) && something;`},

			// === Dynamic import ===
			{Code: `import('module') && result;`},
			{Code: `import(foo) && import(foo).then();`}, // Side effect

			// === Non-null assertions (should not convert if it breaks the assertion) ===
			{Code: `foo! && bar;`}, // Different vars
			{Code: `entity.__helper!.__initialized || options.refresh;`}, // Complex pattern

			// === Computed property names in different contexts ===
			{Code: `obj[fn()] && obj[fn()];`}, // fn() may return different values

			// === Binary expressions that aren't chains ===
			{Code: `a + b && c;`},
			{Code: `a - b && c;`},
			{Code: `a * b && result;`},
			{Code: `a / b && result;`},
			{Code: `a % b && result;`},
			{Code: `a ** b && result;`},
			{Code: `a | b && result;`},
			{Code: `a & b && result;`},
			{Code: `a ^ b && result;`},
			{Code: `a << b && result;`},
			{Code: `a >> b && result;`},
			{Code: `a >>> b && result;`},

			// === Unary expressions ===
			{Code: `+x && result;`},
			{Code: `-x && result;`},
			{Code: `~x && result;`},
			{Code: `!x && y;`}, // x and y are different

			// === Ternary that doesn't form a chain ===
			{Code: `a ? b : c && d;`},
			{Code: `a && b ? c : d;`}, // Different from chain
			{Code: `condition ? foo : bar && baz;`},

			// === Literal values (always truthy/falsy, no need to chain) ===
			{Code: `true && foo.bar;`},
			{Code: `false || foo.bar;`},
			{Code: `1 && result;`},
			{Code: `0 || result;`},
			{Code: `'' || result;`},
			{Code: `'string' && result;`},
			{Code: `[] && result;`},
			{Code: `{} && result;`},
			{Code: `(() => {}) && result;`},
			{Code: `(function() {}) && result;`},

			// === Multiple chains but unrelated ===
			// Removed - each chain SHOULD be reported separately
			// {Code: `a && a.b || c && c.d;`},
			// {Code: `(p && p.q) || (r && r.s);`},
			// Removed - this has TWO independent chains and SHOULD report two errors
			// {Code: `x && x.y && z && z.w;`}

			// === Weird valid edge cases ===
			{Code: `data && data.value !== null;`}, // https://github.com/typescript-eslint/typescript-eslint/issues/7654
			// Removed - this SHOULD convert
			// {Code: `foo && typeof foo.bar !== 'undefined';`},

			// === Cases where the property name changes ===
			{Code: `obj.propA && obj.propB;`}, // Both are accessed, but they're different
			{Code: `foo.x && foo.y;`},         // Same object, different properties
			{Code: `bar.a || bar.b;`},

			// === Chains that start differently ===
			{Code: `foo.a.b && bar.a.b;`}, // Different roots
			{Code: `x.y.z || a.b.c;`},     // Different roots

			// === Tests with comments (should not interfere) ===
			{Code: `/* comment */ foo && bar;`},
			{Code: `foo /* comment */ && bar;`},
			{Code: `foo && /* comment */ bar;`},
			{Code: `foo && bar /* comment */;`},

			// === Whitespace variations (should not interfere) ===
			{Code: "foo&&bar;"},
			{Code: "foo  &&  bar;"},
			{Code: "foo\n&&\nbar;"},
			{Code: "foo\t&&\tbar;"},
		},
		[]rule_tester.InvalidTestCase{})
}
