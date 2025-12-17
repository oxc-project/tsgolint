package prefer_optional_chain

import (
	"strings"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferOptionalChainRule(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
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
		// https://github.com/typescript-eslint/typescript-eslint/issues/8380
		{Code: `
        const a = null;
        const b = 0;
        a === undefined || b === null || b === undefined;
      `},
		// https://github.com/typescript-eslint/typescript-eslint/issues/8380
		{Code: `
        const a = 0;
        const b = 0;
        a === undefined || b === undefined || b === null;
      `},
		// https://github.com/typescript-eslint/typescript-eslint/issues/8380
		{Code: `
        const a = 0;
        const b = 0;
        b === null || a === undefined || b === undefined;
      `},
		// https://github.com/typescript-eslint/typescript-eslint/issues/8380
		{Code: `
        const b = 0;
        b === null || b === undefined;
      `},
		// https://github.com/typescript-eslint/typescript-eslint/issues/8380
		{Code: `
        const a = 0;
        const b = 0;
        b != null && a !== null && a !== undefined;
      `},
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
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar == undefined;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== 0;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== '123';
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== {};
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== false;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== true;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== null;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar !== undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != 0;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != '123';
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != {};
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != false;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != true;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo && foo.bar != undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar == undefined;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== 0;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== '123';
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== {};
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== false;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== true;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== null;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar !== undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != 0;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != '123';
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != {};
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != false;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != true;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo != null && foo.bar != undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | 1;
        foo && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | 0;
        foo != null && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar == undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== 0;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== '123';
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== {};
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== false;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== true;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo && foo.bar !== undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar == undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== 0;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== '123';
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== {};
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== false;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== true;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo != null && foo.bar !== undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== null && foo !== undefined && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== null && foo !== undefined && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== null && foo !== undefined && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== null && foo !== undefined && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== null && foo !== undefined && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== null && foo !== undefined && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== null && foo !== undefined && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== null && foo !== undefined && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== undefined && foo !== undefined && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== undefined && foo !== undefined && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== undefined && foo !== undefined && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | null;
        foo !== undefined && foo !== undefined && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== undefined && foo !== undefined && foo.bar == null;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== undefined && foo !== undefined && foo.bar === undefined;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== undefined && foo !== undefined && foo.bar !== 1;
      `},
		{Code: `
        declare const foo: { bar: number } | undefined;
        foo !== undefined && foo !== undefined && foo.bar != 1;
      `},
		{Code: `
        declare const foo: { bar: number };
        !foo || foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        !foo || foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        !foo || foo.bar !== undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        !foo || foo.bar != null;
      `},
		{Code: `
        declare const foo: { bar: number };
        !foo || foo.bar != undeclaredVar;
      `},
		{Code: `!foo && foo.bar == 0;`},
		{Code: `!foo && foo.bar == 1;`},
		{Code: `!foo && foo.bar == '123';`},
		{Code: `!foo && foo.bar == {};`},
		{Code: `!foo && foo.bar == false;`},
		{Code: `!foo && foo.bar == true;`},
		{Code: `!foo && foo.bar === 0;`},
		{Code: `!foo && foo.bar === 1;`},
		{Code: `!foo && foo.bar === '123';`},
		{Code: `!foo && foo.bar === {};`},
		{Code: `!foo && foo.bar === false;`},
		{Code: `!foo && foo.bar === true;`},
		{Code: `!foo && foo.bar === null;`},
		{Code: `!foo && foo.bar !== undefined;`},
		{Code: `!foo && foo.bar != undefined;`},
		{Code: `!foo && foo.bar != null;`},
		{Code: `foo == null && foo.bar == 0;`},
		{Code: `foo == null && foo.bar == 1;`},
		{Code: `foo == null && foo.bar == '123';`},
		{Code: `foo == null && foo.bar == {};`},
		{Code: `foo == null && foo.bar == false;`},
		{Code: `foo == null && foo.bar == true;`},
		{Code: `foo == null && foo.bar === 0;`},
		{Code: `foo == null && foo.bar === 1;`},
		{Code: `foo == null && foo.bar === '123';`},
		{Code: `foo == null && foo.bar === {};`},
		{Code: `foo == null && foo.bar === false;`},
		{Code: `foo == null && foo.bar === true;`},
		{Code: `foo == null && foo.bar === null;`},
		{Code: `foo == null && foo.bar !== undefined;`},
		{Code: `foo == null && foo.bar != null;`},
		{Code: `foo == null && foo.bar != undefined;`},
		{Code: `
        declare const foo: false | { a: string };
        foo && foo.a == undeclaredVar;
      `},
		{Code: `
        declare const foo: '' | { a: string };
        foo && foo.a == undeclaredVar;
      `},
		{Code: `
        declare const foo: 0 | { a: string };
        foo && foo.a == undeclaredVar;
      `},
		{Code: `
        declare const foo: 0n | { a: string };
        foo && foo.a;
      `},
		{Code: `!foo || foo.bar != undeclaredVar;`},
		{Code: `!foo || foo.bar != null;`},
		{Code: `!foo || foo.bar != undefined;`},
		{Code: `!foo || foo.bar === 0;`},
		{Code: `!foo || foo.bar === 1;`},
		{Code: `!foo || foo.bar === '123';`},
		{Code: `!foo || foo.bar === {};`},
		{Code: `!foo || foo.bar === false;`},
		{Code: `!foo || foo.bar === true;`},
		{Code: `!foo || foo.bar === null;`},
		{Code: `!foo || foo.bar === undeclaredVar;`},
		{Code: `!foo || foo.bar == 0;`},
		{Code: `!foo || foo.bar == 1;`},
		{Code: `!foo || foo.bar == '123';`},
		{Code: `!foo || foo.bar == {};`},
		{Code: `!foo || foo.bar == false;`},
		{Code: `!foo || foo.bar == true;`},
		{Code: `!foo || foo.bar == undeclaredVar;`},
		{Code: `!foo || foo.bar !== undeclaredVar;`},
		{Code: `!foo || foo.bar !== undefined;`},
		{Code: `
        declare const foo: { bar: number };
        foo == null || foo.bar == undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo == null || foo.bar === undeclaredVar;
      `},
		{Code: `
        declare const foo: { bar: number };
        foo == null || foo.bar !== undeclaredVar;
      `},
		{Code: `foo == null || foo.bar != undeclaredVar;`},
		{Code: `foo == null || foo.bar != null;`},
		{Code: `foo == null || foo.bar != undefined;`},
		{Code: `foo == null || foo.bar === 0;`},
		{Code: `foo == null || foo.bar === 1;`},
		{Code: `foo == null || foo.bar === '123';`},
		{Code: `foo == null || foo.bar === {};`},
		{Code: `foo == null || foo.bar === false;`},
		{Code: `foo == null || foo.bar === true;`},
		{Code: `foo == null || foo.bar === null;`},
		{Code: `foo == null || foo.bar === undeclaredVar;`},
		{Code: `foo == null || foo.bar == 0;`},
		{Code: `foo == null || foo.bar == 1;`},
		{Code: `foo == null || foo.bar == '123';`},
		{Code: `foo == null || foo.bar == {};`},
		{Code: `foo == null || foo.bar == false;`},
		{Code: `foo == null || foo.bar == true;`},
		{Code: `foo == null || foo.bar == undeclaredVar;`},
		{Code: `foo == null || foo.bar !== undeclaredVar;`},
		{Code: `foo == null || foo.bar !== undefined;`},
		{Code: `foo || foo.bar != 0;`},
		{Code: `foo || foo.bar != 1;`},
		{Code: `foo || foo.bar != '123';`},
		{Code: `foo || foo.bar != {};`},
		{Code: `foo || foo.bar != false;`},
		{Code: `foo || foo.bar != true;`},
		{Code: `foo || foo.bar === undefined;`},
		{Code: `foo || foo.bar == undefined;`},
		{Code: `foo || foo.bar == null;`},
		{Code: `foo || foo.bar !== 0;`},
		{Code: `foo || foo.bar !== 1;`},
		{Code: `foo || foo.bar !== '123';`},
		{Code: `foo || foo.bar !== {};`},
		{Code: `foo || foo.bar !== false;`},
		{Code: `foo || foo.bar !== true;`},
		{Code: `foo || foo.bar !== null;`},
		{Code: `foo != null || foo.bar != 0;`},
		{Code: `foo != null || foo.bar != 1;`},
		{Code: `foo != null || foo.bar != '123';`},
		{Code: `foo != null || foo.bar != {};`},
		{Code: `foo != null || foo.bar != false;`},
		{Code: `foo != null || foo.bar != true;`},
		{Code: `foo != null || foo.bar === undefined;`},
		{Code: `foo != null || foo.bar == undefined;`},
		{Code: `foo != null || foo.bar == null;`},
		{Code: `foo != null || foo.bar !== 0;`},
		{Code: `foo != null || foo.bar !== 1;`},
		{Code: `foo != null || foo.bar !== '123';`},
		{Code: `foo != null || foo.bar !== {};`},
		{Code: `foo != null || foo.bar !== false;`},
		{Code: `foo != null || foo.bar !== true;`},
		{Code: `foo != null || foo.bar !== null;`},
		{Code: `
        declare const record: Record<string, { kind: string }>;
        record['key'] && record['key'].kind !== '1';
      `},
		{Code: `
        declare const array: { b?: string }[];
        !array[1] || array[1].b === 'foo';
      `},
		{Code: `!a || !b;`},
		{Code: `!a || a.b;`},
		{Code: `!a && a.b;`},
		{Code: `!a && !a.b;`},
		{Code: `!a.b || a.b?.();`},
		{Code: `!a.b || a.b();`},
		{Code: `foo ||= bar;`},
		{Code: `foo ||= bar?.baz;`},
		{Code: `foo ||= bar?.baz?.buzz;`},
		{Code: `foo && bar;`},
		{Code: `foo && foo;`},
		{Code: `foo || bar;`},
		{Code: `foo ?? bar;`},
		{Code: `foo || foo.bar;`},
		{Code: `foo ?? foo.bar;`},
		{Code: `file !== 'index.ts' && file.endsWith('.ts');`},
		{Code: `nextToken && sourceCode.isSpaceBetweenTokens(prevToken, nextToken);`},
		{Code: `result && this.options.shouldPreserveNodeMaps;`},
		{Code: `foo && fooBar.baz;`},
		{Code: `match && match$1 !== undefined;`},
		{Code: `typeof foo === 'number' && foo.toFixed();`},
		{Code: `foo === 'undefined' && foo.length;`},
		{Code: `foo == bar && foo.bar == null;`},
		{Code: `foo === 1 && foo.toFixed();`},
		{Code: `foo.bar(a) && foo.bar(a, b).baz;`},
		{Code: `foo.bar<a>() && foo.bar<a, b>().baz;`},
		{Code: `[1, 2].length && [1, 2, 3].length.toFixed();`},
		{Code: `[1,].length && [1, 2].length.toFixed();`},
		{Code: `(foo?.a).b && foo.a.b.c;`},
		{Code: `(foo?.a)() && foo.a().b;`},
		{Code: `(foo?.a)() && foo.a()();`},
		{Code: `foo !== null && foo !== undefined;`},
		{Code: `x['y'] !== undefined && x['y'] !== null;`},
		{Code: `this.#a && this.#b;`},
		{Code: `!this.#a || !this.#b;`},
		{Code: `a.#foo?.bar;`},
		{Code: `!a.#foo?.bar;`},
		{Code: `!foo().#a || a;`},
		{Code: `!a.b.#a || a;`},
		{Code: `!new A().#b || a;`},
		{Code: `!(await a).#b || a;`},
		{Code: `!(foo as any).bar || 'anything';`},
		{Code: `!foo[1 + 1] || !foo[1 + 2];`},
		{Code: `!foo[1 + 1] || !foo[1 + 2].foo;`},
		{Code: `this && this.foo;`},
		{Code: `!this || !this.foo;`},
		{Code: `!entity.__helper!.__initialized || options.refresh;`},
		{Code: `import.meta || true;`},
		{Code: `import.meta || import.meta.foo;`},
		{Code: `!import.meta && false;`},
		{Code: `!import.meta && !import.meta.foo;`},
		{Code: `new.target || new.target.length;`},
		{Code: `!new.target || true;`},
		// Do not handle direct optional chaining on private properties because this TS limitation (https://github.com/microsoft/TypeScript/issues/42734)
		{Code: `foo && foo.#bar;`},
		{Code: `!foo || !foo.#bar;`},
		{Code: `({}) && {}.toString();`},
		{Code: `[] && [].length;`},
		{Code: `(() => {}) && (() => {}).name;`},
		{Code: `(function () {}) && function () {}.name;`},
		{Code: `(class Foo {}) && class Foo {}.constructor;`},
		{Code: `new Map().get('a') && new Map().get('a').what;`},
		// https://github.com/typescript-eslint/typescript-eslint/issues/7654
		{Code: `data && data.value !== null;`},
		{Code: `<div /> && (<div />).wtf;`},
		{Code: `<></> && (<></>).wtf;`},
		{Code: `foo[x++] && foo[x++].bar;`},
		{Code: `foo[yield x] && foo[yield x].bar;`},
		{Code: `a = b && (a = b).wtf;`},
		{Code: `(x || y) != null && (x || y).foo;`},
		{Code: `(await foo) && (await foo).bar;`},
		{Code: `
        declare const foo: { bar: string } | null;
        foo !== null && foo.bar !== null;
      `},
		{Code: `
        declare const foo: { bar: string | null } | null;
        foo != null && foo.bar !== null;
      `},
		{Code: `
          declare const x: string;
          x && x.length;
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: string;
          foo && foo.toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const x: string | number | boolean | object;
          x && x.toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: { bar: string };
          foo && foo.bar && foo.bar.toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: string;
          foo && foo.toString() && foo.toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: { bar: string };
          foo && foo.bar && foo.bar.toString() && foo.bar.toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo1: { bar: string | null };
          foo1 && foo1.bar;
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: string;
          (foo || {}).toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const foo: string | null;
          (foo || 'a' || {}).toString();
        `, Options: PreferOptionalChainOptions{RequireNullish: true}},
		{Code: `
          declare const x: any;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckAny: false}},
		{Code: `
          declare const x: bigint;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckBigInt: false}},
		{Code: `
          declare const x: boolean;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckBoolean: false}},
		{Code: `
          declare const x: number;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckNumber: false}},
		{Code: `
          declare const x: string;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckString: false}},
		{Code: `
          declare const x: unknown;
          x && x.length;
        `, Options: PreferOptionalChainOptions{CheckUnknown: false}},
		{Code: `(x = {}) && (x.y = true) != null && x.y.toString();`},
		{Code: "('x' as `${'x'}`) && ('x' as `${'x'}`).length;"},
		{Code: "`x` && `x`.length;"},
		{Code: "`x${a}` && `x${a}`.length;"},
		{Code: `
        declare const x: false | { a: string };
        x && x.a;
      `},
		{Code: `
        declare const x: false | { a: string };
        !x || x.a;
      `},
		{Code: `
        declare const x: '' | { a: string };
        x && x.a;
      `},
		{Code: `
        declare const x: '' | { a: string };
        !x || x.a;
      `},
		{Code: `
        declare const x: 0 | { a: string };
        x && x.a;
      `},
		{Code: `
        declare const x: 0 | { a: string };
        !x || x.a;
      `},
		{Code: `
        declare const x: 0n | { a: string };
        x && x.a;
      `},
		{Code: `
        declare const x: 0n | { a: string };
        !x || x.a;
      `},
		{Code: `typeof globalThis !== 'undefined' && globalThis.Array();`},
		{Code: `
        declare const x: void | (() => void);
        x && x();
      `},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		{
			Code: `(foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo || ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(await foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(await foo)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo1?.foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo1?.foo2?.foo3;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo1?.foo2 || ({})).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo1?.foo2?.foo3;`,
						},
					},
				},
			},
		},
		{
			Code: `((() => foo())() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(() => foo())()?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `const foo = (bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `const foo = bar?.baz;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo.bar || {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo.bar?.[baz];`,
						},
					},
				},
			},
		},
		{
			Code: `((foo1 || {}).foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1 || {}).foo2?.foo3;`,
						},
					},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1?.foo2 || {}).foo3;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo || undefined || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo || undefined)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo() || bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo() || bar)?.baz;`,
						},
					},
				},
			},
		},
		{
			Code: `((foo1 ? foo2 : foo3) || {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1 ? foo2 : foo3)?.foo4;`,
						},
					},
				},
			},
		},
		{
			Code: `
          if (foo) {
            (foo || {}).bar;
          }
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          if (foo) {
            foo?.bar;
          }
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          if ((foo || {}).bar) {
            foo.bar;
          }
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          if (foo?.bar) {
            foo.bar;
          }
        `,
						},
					},
				},
			},
		},
		{
			Code: `(undefined && foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(undefined && foo)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo ?? ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(await foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(await foo)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo1?.foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo1?.foo2?.foo3;`,
						},
					},
				},
			},
		},
		{
			Code: `((() => foo())() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(() => foo())()?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `const foo = (bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `const foo = bar?.baz;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo.bar ?? {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `foo.bar?.[baz];`,
						},
					},
				},
			},
		},
		{
			Code: `((foo1 ?? {}).foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1 ?? {}).foo2?.foo3;`,
						},
					},
				},
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1?.foo2 ?? {}).foo3;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo ?? undefined ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo ?? undefined)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo() ?? bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo() ?? bar)?.baz;`,
						},
					},
				},
			},
		},
		{
			Code: `((foo1 ? foo2 : foo3) ?? {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo1 ? foo2 : foo3)?.foo4;`,
						},
					},
				},
			},
		},
		{
			Code: `if (foo) { (foo ?? {}).bar; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `if (foo) { foo?.bar; }`,
						},
					},
				},
			},
		},
		{
			Code: `if ((foo ?? {}).bar) { foo.bar; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `if (foo?.bar) { foo.bar; }`,
						},
					},
				},
			},
		},
		{
			Code: `(undefined && foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(undefined && foo)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(a > b || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(a > b)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(((typeof x) as string) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `((typeof x) as string)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(void foo() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(void foo())?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `((a ? b : c) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(a ? b : c)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `((a instanceof Error) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(a instanceof Error)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `((a << b) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(a << b)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `((foo ** 2) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo ** 2)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo ** 2 || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo ** 2)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(foo++ || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(foo++)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(+foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `(+foo)?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `(this || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    `this?.foo;`,
						},
					},
				},
			},
		},
		{
			Code:   `foo && foo.bar == 0;`,
			Output: []string{`foo?.bar == 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar == 1;`,
			Output: []string{`foo?.bar == 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar == '123';`,
			Output: []string{`foo?.bar == '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar == {};`,
			Output: []string{`foo?.bar == {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar == false;`,
			Output: []string{`foo?.bar == false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar == true;`,
			Output: []string{`foo?.bar == true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === 0;`,
			Output: []string{`foo?.bar === 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === 1;`,
			Output: []string{`foo?.bar === 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === '123';`,
			Output: []string{`foo?.bar === '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === {};`,
			Output: []string{`foo?.bar === {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === false;`,
			Output: []string{`foo?.bar === false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === true;`,
			Output: []string{`foo?.bar === true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar === null;`,
			Output: []string{`foo?.bar === null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar !== undefined;`,
			Output: []string{`foo?.bar !== undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar != undefined;`,
			Output: []string{`foo?.bar != undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar != null;`,
			Output: []string{`foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == 0;`,
			Output: []string{`foo?.bar == 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == 1;`,
			Output: []string{`foo?.bar == 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == '123';`,
			Output: []string{`foo?.bar == '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == {};`,
			Output: []string{`foo?.bar == {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == false;`,
			Output: []string{`foo?.bar == false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar == true;`,
			Output: []string{`foo?.bar == true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === 0;`,
			Output: []string{`foo?.bar === 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === 1;`,
			Output: []string{`foo?.bar === 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === '123';`,
			Output: []string{`foo?.bar === '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === {};`,
			Output: []string{`foo?.bar === {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === false;`,
			Output: []string{`foo?.bar === false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === true;`,
			Output: []string{`foo?.bar === true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar === null;`,
			Output: []string{`foo?.bar === null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar !== undefined;`,
			Output: []string{`foo?.bar !== undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar != undefined;`,
			Output: []string{`foo?.bar != undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar != null;`,
			Output: []string{`foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo && foo.bar != null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo != null && foo.bar != null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != 0;`,
			Output: []string{`foo?.bar != 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != 1;`,
			Output: []string{`foo?.bar != 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != '123';`,
			Output: []string{`foo?.bar != '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != {};`,
			Output: []string{`foo?.bar != {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != false;`,
			Output: []string{`foo?.bar != false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar != true;`,
			Output: []string{`foo?.bar != true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar === undefined;`,
			Output: []string{`foo?.bar === undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar == undefined;`,
			Output: []string{`foo?.bar == undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar == null;`,
			Output: []string{`foo?.bar == null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== 0;`,
			Output: []string{`foo?.bar !== 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== 1;`,
			Output: []string{`foo?.bar !== 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== '123';`,
			Output: []string{`foo?.bar !== '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== {};`,
			Output: []string{`foo?.bar !== {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== false;`,
			Output: []string{`foo?.bar !== false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== true;`,
			Output: []string{`foo?.bar !== true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || foo.bar !== null;`,
			Output: []string{`foo?.bar !== null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != 0;`,
			Output: []string{`foo?.bar != 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != 1;`,
			Output: []string{`foo?.bar != 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != '123';`,
			Output: []string{`foo?.bar != '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != {};`,
			Output: []string{`foo?.bar != {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != false;`,
			Output: []string{`foo?.bar != false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar != true;`,
			Output: []string{`foo?.bar != true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar === undefined;`,
			Output: []string{`foo?.bar === undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar == undefined;`,
			Output: []string{`foo?.bar == undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar == null;`,
			Output: []string{`foo?.bar == null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== 0;`,
			Output: []string{`foo?.bar !== 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== 1;`,
			Output: []string{`foo?.bar !== 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== '123';`,
			Output: []string{`foo?.bar !== '123';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== {};`,
			Output: []string{`foo?.bar !== {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== false;`,
			Output: []string{`foo?.bar !== false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== true;`,
			Output: []string{`foo?.bar !== true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo == null || foo.bar !== null;`,
			Output: []string{`foo?.bar !== null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar == null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar == undefined;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == undefined;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar === undefined;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar === undefined;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== 0;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 0;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== 1;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 1;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== '123';
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== '123';
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== {};
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== {};
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== false;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== false;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== true;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== true;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar !== null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != 0;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != 0;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != 1;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != 1;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != '123';
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != '123';
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != {};
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != {};
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != false;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != false;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          !foo || foo.bar != true;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar != true;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar == null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar == undefined;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar == undefined;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar === undefined;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar === undefined;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== 0;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 0;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== 1;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== 1;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== '123';
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== '123';
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== {};
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== {};
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== false;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== false;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== true;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== true;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number };
          foo == null || foo.bar !== null;
        `,
			Output: []string{`
          declare const foo: { bar: number };
          foo?.bar !== null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && null != foo.bar && '123' == foo.bar.baz;`,
			Output: []string{`'123' == foo?.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && null != foo.bar && '123' === foo.bar.baz;`,
			Output: []string{`'123' === foo?.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && null != foo.bar && undefined !== foo.bar.baz;`,
			Output: []string{`undefined !== foo?.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar && foo.bar.baz || baz && baz.bar && baz.bar.foo`,
			Output: []string{`foo?.bar?.baz || baz?.bar?.foo`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		// case with inconsistent checks should "break" the chain
		{
			Code:   `foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;`,
			Output: []string{`foo?.bar?.baz !== undefined && foo.bar.baz.buzz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          foo.bar &&
            foo.bar.baz != null &&
            foo.bar.baz.qux !== undefined &&
            foo.bar.baz.qux.buzz;
        `,
			Output: []string{`
          foo.bar?.baz?.qux !== undefined &&
            foo.bar.baz.qux.buzz;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar(baz => <This Requires Spaces />);`,
			Output: []string{`foo?.bar(baz => <This Requires Spaces />);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar(baz => typeof baz);`,
			Output: []string{`foo?.bar(baz => typeof baz);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo['some long string'] && foo['some long string'].baz;`,
			Output: []string{`foo?.['some long string']?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   "foo && foo[`some long string`] && foo[`some long string`].baz;",
			Output: []string{"foo?.[`some long string`]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   "foo && foo[`some ${long} string`] && foo[`some ${long} string`].baz;",
			Output: []string{"foo?.[`some ${long} string`]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo[bar as string] && foo[bar as string].baz;`,
			Output: []string{`foo?.[bar as string]?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo[1 + 2] && foo[1 + 2].baz;`,
			Output: []string{`foo?.[1 + 2]?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo[typeof bar] && foo[typeof bar].baz;`,
			Output: []string{`foo?.[typeof bar]?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar(a) && foo.bar(a, b).baz;`,
			Output: []string{`foo?.bar(a) && foo.bar(a, b).baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo() && foo()(bar);`,
			Output: []string{`foo()?.(bar);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo<string>() && foo<string>().bar;`,
			Output: []string{`foo?.<string>()?.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo<string>() && foo<string, number>().bar;`,
			Output: []string{`foo?.<string>() && foo<string, number>().bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          foo && foo.bar(/* comment */a,
            // comment2
            b, );
        `,
			Output: []string{`
          foo?.bar(/* comment */a,
            // comment2
            b, );
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar != null;`,
			Output: []string{`foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar != undefined;`,
			Output: []string{`foo?.bar != undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar != null && baz;`,
			Output: []string{`foo?.bar != null && baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `this.bar && this.bar.baz;`,
			Output: []string{`this.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo?.();`,
			Output: []string{`foo?.();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo.bar && foo.bar?.();`,
			Output: []string{`foo.bar?.();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.bar(baz => <This Requires Spaces />);`,
			Output: []string{`foo?.bar(baz => <This Requires Spaces />);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!this.bar || !this.bar.baz;`,
			Output: []string{`!this.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!a.b || !a.b();`,
			Output: []string{`!a.b?.();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo.bar || !foo.bar.baz;`,
			Output: []string{`!foo.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo[bar] || !foo[bar]?.[baz];`,
			Output: []string{`!foo[bar]?.[baz];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || !foo?.bar.baz;`,
			Output: []string{`!foo?.bar.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `(!foo || !foo.bar || !foo.bar.baz) && (!baz || !baz.bar || !baz.bar.foo);`,
			Output: []string{`(!foo?.bar?.baz) && (!baz?.bar?.foo);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          class Foo {
            constructor() {
              new.target && new.target.length;
            }
          }
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          class Foo {
            constructor() {
              new.target?.length;
            }
          }
        `,
						},
					},
				},
			},
		},
		{
			Code:   `import.meta && import.meta?.baz;`,
			Output: []string{`import.meta?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!import.meta || !import.meta?.baz;`,
			Output: []string{`!import.meta?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `import.meta && import.meta?.() && import.meta?.().baz;`,
			Output: []string{`import.meta?.()?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo() || !foo().bar;`,
			Output: []string{`!foo()?.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo!.bar || !foo!.bar.baz;`,
			Output: []string{`!foo!.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo!.bar!.baz || !foo!.bar!.baz!.paz;`,
			Output: []string{`!foo!.bar!.baz?.paz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo.bar!.baz || !foo.bar!.baz!.paz;`,
			Output: []string{`!foo.bar!.baz?.paz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo != null && foo.bar != null;`,
			Output: []string{`foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: string | null } | null;
          foo !== null && foo.bar != null;
        `,
			Output: []string{`
          declare const foo: { bar: string | null } | null;
          foo?.bar != null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		// https://github.com/typescript-eslint/typescript-eslint/issues/6332
		{
			Code:   `unrelated != null && foo != null && foo.bar != null;`,
			Output: []string{`unrelated != null && foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `unrelated1 != null && unrelated2 != null && foo != null && foo.bar != null;`,
			Output: []string{`unrelated1 != null && unrelated2 != null && foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		// https://github.com/typescript-eslint/typescript-eslint/issues/1461
		{
			Code:   `foo1 != null && foo1.bar != null && foo2 != null && foo2.bar != null;`,
			Output: []string{`foo1?.bar != null && foo2?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo && foo.a && bar && bar.a;`,
			Output: []string{`foo?.a && bar?.a;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo.bar.baz != null && foo?.bar?.baz.bam != null;`,
			Output: []string{`foo.bar.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo?.bar.baz != null && foo.bar?.baz.bam != null;`,
			Output: []string{`foo?.bar.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo?.bar?.baz != null && foo.bar.baz.bam != null;`,
			Output: []string{`foo?.bar?.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo.bar.baz != null && foo!.bar!.baz.bam != null;`,
			Output: []string{`foo.bar.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo!.bar.baz != null && foo.bar!.baz.bam != null;`,
			Output: []string{`foo!.bar.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo!.bar!.baz != null && foo.bar.baz.bam != null;`,
			Output: []string{`foo!.bar!.baz?.bam != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          a &&
            a.b != null &&
            a.b.c !== undefined &&
            a.b.c !== null &&
            a.b.c.d != null &&
            a.b.c.d.e !== null &&
            a.b.c.d.e !== undefined &&
            a.b.c.d.e.f != undefined &&
            typeof a.b.c.d.e.f.g !== 'undefined' &&
            a.b.c.d.e.f.g !== null &&
            a.b.c.d.e.f.g.h;
        `,
			Output: []string{`
          a?.b?.c?.d?.e?.f?.g?.h;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          !a ||
            a.b == null ||
            a.b.c === undefined ||
            a.b.c === null ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            typeof a.b.c.d.e.f.g === 'undefined' ||
            a.b.c.d.e.f.g === null ||
            !a.b.c.d.e.f.g.h;
        `,
			Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          !a ||
            a.b == null ||
            a.b.c === null ||
            a.b.c === undefined ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            typeof a.b.c.d.e.f.g === 'undefined' ||
            a.b.c.d.e.f.g === null ||
            !a.b.c.d.e.f.g.h;
        `,
			Output: []string{`
          !a?.b?.c?.d?.e?.f?.g?.h;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `undefined !== foo && null !== foo && null != foo.bar && foo.bar.baz;`,
			Output: []string{`foo?.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            foo.bar.baz;
        `,
			Output: []string{`
          foo?.bar?.baz;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            null != foo.bar.baz;
        `,
			Output: []string{`
          null != foo?.bar?.baz;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            null !== foo.bar.baz &&
            'undefined' !== typeof foo.bar.baz;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          null !== foo?.bar?.baz &&
            'undefined' !== typeof foo.bar.baz;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          foo != null &&
            typeof foo.bar !== 'undefined' &&
            foo.bar !== null &&
            foo.bar.baz !== null &&
            typeof foo.bar.baz !== 'undefined';
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          foo?.bar?.baz !== null &&
            typeof foo.bar.baz !== 'undefined';
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            null !== foo.bar.baz &&
            undefined !== foo.bar.baz;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          null !== foo?.bar?.baz &&
            undefined !== foo.bar.baz;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          foo != null &&
            typeof foo.bar !== 'undefined' &&
            foo.bar !== null &&
            foo.bar.baz !== null &&
            foo.bar.baz !== undefined;
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          foo?.bar?.baz !== null &&
            foo.bar.baz !== undefined;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          null != foo &&
            'undefined' !== typeof foo.bar &&
            null !== foo.bar &&
            undefined !== foo.bar.baz &&
            null !== foo.bar.baz;
        `,
			Output: []string{`
          undefined !== foo?.bar?.baz &&
            null !== foo.bar.baz;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          foo != null &&
            typeof foo.bar !== 'undefined' &&
            foo.bar !== null &&
            foo.bar.baz !== undefined &&
            foo.bar.baz !== null;
        `,
			Output: []string{`
          foo?.bar?.baz !== undefined &&
            foo.bar.baz !== null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `(await foo).bar && (await foo).bar.baz;`,
			Output: []string{`(await foo).bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          !a ||
            a.b == null ||
            a.b.c === undefined ||
            a.b.c === null ||
            a.b.c.d == null ||
            a.b.c.d.e === null ||
            a.b.c.d.e === undefined ||
            a.b.c.d.e.f == undefined ||
            a.b.c.d.e.f.g == null ||
            a.b.c.d.e.f.g.h;
        `,
			Output: []string{`
          a?.b?.c?.d?.e?.f?.g == null ||
            a.b.c.d.e.f.g.h;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number } | null | undefined;
          foo && foo.bar != null;
        `,
			Output: []string{`
          declare const foo: { bar: number } | null | undefined;
          foo?.bar != null;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number } | undefined;
          foo && typeof foo.bar !== 'undefined';
        `,
			Output: []string{`
          declare const foo: { bar: number } | undefined;
          typeof foo?.bar !== 'undefined';
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number } | undefined;
          foo && 'undefined' !== typeof foo.bar;
        `,
			Output: []string{`
          declare const foo: { bar: number } | undefined;
          'undefined' !== typeof foo?.bar;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const thing1: string | null;
          thing1 && thing1.toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const thing1: string | null;
          thing1?.toString();
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const thing1: string | null;
          thing1 && thing1.toString() && true;
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const thing1: string | null;
          thing1?.toString() && true;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: string | null;
          foo && foo.toString() && foo.toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const foo: string | null;
          foo?.toString() && foo.toString();
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo && foo.bar && foo.bar.toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Output: []string{`
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo?.bar?.toString();
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo && foo.bar && foo.bar.toString() && foo.bar.toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Output: []string{`
          declare const foo: { bar: string | null | undefined } | null | undefined;
          foo?.bar?.toString() && foo.bar.toString();
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: string | null;
          (foo || {}).toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const foo: string | null;
          foo?.toString();
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: string;
          (foo || undefined || {}).toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const foo: string;
          (foo || undefined)?.toString();
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: string | null;
          (foo || undefined || {}).toString();
        `,
			Options: PreferOptionalChainOptions{RequireNullish: true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const foo: string | null;
          (foo || undefined)?.toString();
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number } | null | undefined;
          foo != undefined && foo.bar;
        `,
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: true},
			Output: []string{`
          declare const foo: { bar: number } | null | undefined;
          foo?.bar;
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: number } | null | undefined;
          foo != undefined && foo.bar;
        `,
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          declare const foo: { bar: number } | null | undefined;
          foo?.bar;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
          declare const foo: { bar: boolean } | null | undefined;
          declare function acceptsBoolean(arg: boolean): void;
          acceptsBoolean(foo != null && foo.bar);
        `,
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: true},
			Output: []string{`
          declare const foo: { bar: boolean } | null | undefined;
          declare function acceptsBoolean(arg: boolean): void;
          acceptsBoolean(foo?.bar);
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          function foo(globalThis?: { Array: Function }) {
            typeof globalThis !== 'undefined' && globalThis.Array();
          }
        `,
			Output: []string{`
          function foo(globalThis?: { Array: Function }) {
            globalThis?.Array();
          }
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
          typeof globalThis !== 'undefined' && globalThis.Array && globalThis.Array();
        `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
          typeof globalThis !== 'undefined' && globalThis.Array?.();
        `,
						},
					},
				},
			},
		},
		{
			Code:   `a && (a.b && a.b.c)`,
			Output: []string{`a?.b?.c`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `(a && a.b) && a.b.c`,
			Output: []string{`a?.b?.c`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `((a && a.b)) && a.b.c`,
			Output: []string{`a?.b?.c`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo(a && (a.b && a.b.c))`,
			Output: []string{`foo(a?.b?.c)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `foo(a && a.b && a.b.c)`,
			Output: []string{`foo(a?.b?.c)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `!foo || !foo.bar || ((((!foo.bar.baz || !foo.bar.baz()))));`,
			Output: []string{`!foo?.bar?.baz?.();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code:   `a !== undefined && ((a !== null && a.prop));`,
			Output: []string{`a?.prop;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
declare const foo: {
  bar: undefined | (() => void);
};

foo.bar && foo.bar();
        `,
			Output: []string{`
declare const foo: {
  bar: undefined | (() => void);
};

foo.bar?.();
        `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
declare const foo: { bar: string };

const baz = foo && foo.bar;
        `,
			Options: PreferOptionalChainOptions{CheckString: false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output: `
declare const foo: { bar: string };

const baz = foo?.bar;
        `,
						},
					},
				},
			},
		},
		{
			Code: `
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar !== undefined &&
                  foo.bar() !== undefined &&
                  foo.bar().baz !== undefined &&
                  foo.bar().baz.buzz !== undefined &&
                  foo.bar().baz.buzz();
              `,
			Output: []string{`
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar?.() !== undefined &&
                  foo.bar().baz !== undefined &&
                  foo.bar().baz.buzz !== undefined &&
                  foo.bar().baz.buzz();
              `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar !== undefined && foo.bar?.() !== undefined && foo.bar?.().baz;
              `,
			Output: []string{`
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar?.() !== undefined && foo.bar?.().baz;
              `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar === undefined ||
                  foo.bar() === undefined ||
                  foo.bar().baz === undefined ||
                  foo.bar().baz.buzz === undefined ||
                  foo.bar().baz.buzz();
              `,
			Output: []string{`
                declare const foo: {
                  bar: () =>
                    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
                    | null
                    | undefined;
                };
                foo.bar?.() === undefined ||
                  foo.bar().baz === undefined ||
                  foo.bar().baz.buzz === undefined ||
                  foo.bar().baz.buzz();
              `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
		{
			Code: `
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar === undefined || foo.bar?.() === undefined || foo.bar?.().baz;
              `,
			Output: []string{`
                declare const foo: { bar: () => { baz: number } | null | undefined };
                foo.bar?.() === undefined || foo.bar?.().baz;
              `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
				},
			},
		},
	}

	// ========== BASE CASES ==========
	// These match the upstream typescript-eslint 'describe("base cases", ...)' tests

	// --- AND operator ---

	// AND boolean - basic
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{Operator: "&&"})...)

	// AND boolean - with trailing && bing
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: AddTrailingAnd,
	})...)

	// AND boolean - with trailing && bing.bong
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: AddTrailingAndBingBong,
	})...)

	// !== null - VALID with full | null | undefined type
	validCases = append(validCases, GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualNull,
	})...)

	// !== null - INVALID with | null only type (suggestion fixer)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:           "&&",
		MutateCode:         ReplaceOperatorWithStrictNotEqualNull,
		MutateDeclaration:  RemoveUndefinedFromType,
		MutateOutput:       Identity,
		UseSuggestionFixer: true,
	})...)

	// != null - INVALID (suggestion fixer)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:           "&&",
		MutateCode:         ReplaceOperatorWithNotEqualNull,
		MutateOutput:       Identity,
		UseSuggestionFixer: true,
	})...)

	// !== undefined - VALID with full type (skip IDs 20, 26)
	validCases = append(validCases, GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "&&",
		MutateCode: ReplaceOperatorWithStrictNotEqualUndefined,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})...)

	// !== undefined - INVALID with | undefined only type (suggestion fixer)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:           "&&",
		MutateCode:         ReplaceOperatorWithStrictNotEqualUndefined,
		MutateDeclaration:  RemoveNullFromType,
		MutateOutput:       Identity,
		UseSuggestionFixer: true,
	})...)

	// !== undefined - 2 hardcoded special cases
	invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
		Code: `
declare const foo: {
  bar: () =>
    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
    | null
    | undefined;
};
foo.bar !== undefined &&
  foo.bar() !== undefined &&
  foo.bar().baz !== undefined &&
  foo.bar().baz.buzz !== undefined &&
  foo.bar().baz.buzz();
`,
		Output: []string{`
declare const foo: {
  bar: () =>
    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
    | null
    | undefined;
};
foo.bar?.() !== undefined &&
  foo.bar().baz !== undefined &&
  foo.bar().baz.buzz !== undefined &&
  foo.bar().baz.buzz();
`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
	}, rule_tester.InvalidTestCase{
		Code: `
declare const foo: { bar: () => { baz: number } | null | undefined };
foo.bar !== undefined && foo.bar?.() !== undefined && foo.bar?.().baz;
`,
		Output: []string{`
declare const foo: { bar: () => { baz: number } | null | undefined };
foo.bar?.() !== undefined && foo.bar?.().baz;
`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
	})

	// != undefined - INVALID (suggestion fixer)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:           "&&",
		MutateCode:         ReplaceOperatorWithNotEqualUndefined,
		MutateOutput:       Identity,
		UseSuggestionFixer: true,
	})...)

	// --- OR operator ---

	// OR boolean - negated (!foo || !foo.bar pattern)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator: "||",
		MutateCode: func(s string) string {
			return "!" + strings.ReplaceAll(s, "||", "|| !")
		},
		MutateOutput: func(s string) string {
			return "!" + s
		},
	})...)

	// === null - VALID with full type
	validCases = append(validCases, GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "||",
		MutateCode: ReplaceOperatorWithStrictEqualNull,
	})...)

	// === null - INVALID with | null only type + trailing check (suggestion fixer)
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:           "||",
		MutateCode:         AddTrailingStrictEqualNull(ReplaceOperatorWithStrictEqualNull),
		MutateDeclaration:  RemoveUndefinedFromType,
		MutateOutput:       AddTrailingStrictEqualNull(Identity),
		UseSuggestionFixer: true,
	})...)

	// == null - INVALID with trailing check
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:     "||",
		MutateCode:   AddTrailingEqualNull(ReplaceOperatorWithEqualNull),
		MutateOutput: AddTrailingEqualNull(Identity),
	})...)

	// === undefined - VALID with full type (skip IDs 20, 26)
	validCases = append(validCases, GenerateValidBaseCases(BaseCaseOptions{
		Operator:   "||",
		MutateCode: ReplaceOperatorWithStrictEqualUndefined,
		SkipIDs:    map[int]bool{20: true, 26: true},
	})...)

	// === undefined - INVALID with | undefined only type + trailing check
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:          "||",
		MutateCode:        AddTrailingStrictEqualUndefined(ReplaceOperatorWithStrictEqualUndefined),
		MutateDeclaration: RemoveNullFromType,
		MutateOutput:      AddTrailingStrictEqualUndefined(Identity),
	})...)

	// === undefined - 2 hardcoded special cases
	invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
		Code: `
declare const foo: {
  bar: () =>
    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
    | null
    | undefined;
};
foo.bar === undefined ||
  foo.bar() === undefined ||
  foo.bar().baz === undefined ||
  foo.bar().baz.buzz === undefined ||
  foo.bar().baz.buzz();
`,
		Output: []string{`
declare const foo: {
  bar: () =>
    | { baz: { buzz: (() => number) | null | undefined } | null | undefined }
    | null
    | undefined;
};
foo.bar?.() === undefined ||
  foo.bar().baz === undefined ||
  foo.bar().baz.buzz === undefined ||
  foo.bar().baz.buzz();
`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
	}, rule_tester.InvalidTestCase{
		Code: `
declare const foo: { bar: () => { baz: number } | null | undefined };
foo.bar === undefined || foo.bar?.() === undefined || foo.bar?.().baz;
`,
		Output: []string{`
declare const foo: { bar: () => { baz: number } | null | undefined };
foo.bar?.() === undefined || foo.bar?.().baz;
`},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
	})

	// == undefined - INVALID with trailing check
	invalidCases = append(invalidCases, GenerateBaseCases(BaseCaseOptions{
		Operator:     "||",
		MutateCode:   AddTrailingEqualUndefined(ReplaceOperatorWithEqualUndefined),
		MutateOutput: AddTrailingEqualUndefined(Identity),
	})...)

	// --- Spacing sanity checks ---
	// These test that extra spacing in the code is handled correctly
	invalidCases = append(invalidCases, DedupeInvalidTestCases(
		GenerateBaseCases(BaseCaseOptions{
			Operator:     "&&",
			MutateCode:   AddSpacingAfterDots,
			MutateOutput: AddSpacingInsideBrackets,
		}),
		GenerateBaseCases(BaseCaseOptions{
			Operator:     "&&",
			MutateCode:   AddNewlineAfterDots,
			MutateOutput: AddNewlineInsideBrackets,
		}),
	)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, validCases, invalidCases)
}
