package no_misused_spread

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoMisusedSpreadRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoMisusedSpreadRule, []rule_tester.ValidTestCase{
		{Code: "const a = [...[1, 2, 3]];"},
		{Code: "const a = [...([1, 2, 3] as const)];"},
		{Code: `
      declare const data: any;
      const a = [...data];
    `},
		{Code: `
      declare const data: unknown;
      const a = [...data];
    `},
		{Code: `
      const a = [1, 2, 3];
      const b = [...a];
    `},
		{Code: `
      const a = [1, 2, 3] as const;
      const b = [...a];
    `},
		{Code: `
      declare function getArray(): number[];
      const a = [...getArray()];
    `},
		{Code: `
      declare function getTuple(): readonly number[];
      const a = [...getTuple()];
    `},
		{Code: `
      const iterator = {
        *[Symbol.iterator]() {
          yield 1;
          yield 2;
          yield 3;
        },
      };

      const a = [...iterator];
    `},
		{Code: `
      declare const data: Iterable<number> | number[];

      const a = [...data];
    `},
		{Code: `
      declare const data: Iterable<number> & number[];

      const a = [...data];
    `},
		{Code: `
      declare function getIterable(): Iterable<number>;

      const a = [...getIterable()];
    `},
		{Code: `
      declare const data: Uint8Array;

      const a = [...data];
    `},
		{Code: `
      declare const data: TypedArray;

      const a = [...data];
    `},
		{Code: "const o = { ...{ a: 1, b: 2 } };"},
		{Code: "const o = { ...({ a: 1, b: 2 } as const) };"},
		{Code: `
      declare const obj: any;

      const o = { ...obj };
    `},
		{Code: `
      declare const obj: { a: number; b: number } | any;

      const o = { ...obj };
    `},
		{Code: `
      declare const obj: { a: number; b: number } & any;

      const o = { ...obj };
    `},
		{Code: `
      const obj = { a: 1, b: 2 };
      const o = { ...obj };
    `},
		{Code: `
      declare const obj: { a: number; b: number };
      const o = { ...obj };
    `},
		{Code: `
      declare function getObject(): { a: number; b: number };
      const o = { ...getObject() };
    `},
		{Code: `
      function f() {}

      f.prop = 1;

      const o = { ...f };
    `},
		{Code: `
      const f = () => {};

      f.prop = 1;

      const o = { ...f };
    `},
		{Code: `
      function* generator() {}

      generator.prop = 1;

      const o = { ...generator };
    `},
		{Code: `
      declare const promiseLike: PromiseLike<number>;

      const o = { ...promiseLike };
    `},
		{
			Code: `
        const obj = { a: 1, b: 2 };
        const o = <div {...x} />;
      `,
			Tsx: true,
		},
		{
			Code: `
        declare const obj: { a: number; b: number } | any;
        const o = <div {...x} />;
      `,
			Tsx: true,
		},
		{
			Code: `
        const promise = new Promise(() => {});
        const o = { ...promise };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "lib", "name": "Promise"}]}`),
		},
		{Code: `
      interface A {}

      declare const a: A;

      const o = { ...a };
    `},
		{Code: `
      const o = { ...'test' };
    `},
		{
			Code: `
        const str: string = 'test';
        const a = [...str];
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "lib", "name": "string"}]}`),
		},
		{
			Code: `
        function f() {}

        const a = { ...f };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "f"}]}`),
		},
		{
			Code: `
        declare const iterator: Iterable<string>;

        const a = { ...iterator };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "lib", "name": "Iterable"}]}`),
		},
		{
			Code: `
        type BrandedString = string & { __brand: 'safe' };

        declare const brandedString: BrandedString;

        const spreadBrandedString = [...brandedString];
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "BrandedString"}]}`),
		},
		{
			Code: `
        type CustomIterable = {
          [Symbol.iterator]: () => Generator<string>;
        };

        declare const iterator: CustomIterable;

        const a = { ...iterator };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "CustomIterable"}]}`),
		},
		{
			Code: `
        type CustomIterable = {
          [Symbol.iterator]: () => string;
        };

        declare const iterator: CustomIterable;

        const a = { ...iterator };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "CustomIterable"}]}`),
		},
		{
			Code: `
        declare module 'module' {
          export type CustomIterable = {
            [Symbol.iterator]: () => string;
          };
        }

        import { CustomIterable } from 'module';

        declare const iterator: CustomIterable;

        const a = { ...iterator };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "package", "name": "CustomIterable", "package": "module"}]}`),
		},
		{
			Code: `
        class A {
          a = 1;
        }

        const a = new A();

        const o = { ...a };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "A"}]}`),
		},
		{
			Code: `
        const a = {
          ...class A {
            static value = 1;
          },
        };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "A"}]}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "const a = [...'test'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      1,
					Column:    12,
					EndColumn: 21,
				},
			},
		},
		{
			Code: `
        function withText<Text extends string>(text: Text) {
          return [...text];
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    19,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
        const test = 'hello';
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        const test = ` + "`" + `he${'ll'}o` + "`" + `;
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const test: string;
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const test: string | number[];
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const test: string & { __brand: 'test' };
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const test: number | (boolean | (string & { __brand: true }));
        const a = [...test];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare function getString(): string;
        const a = [...getString()];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
        declare function textIdentity(...args: string[]);

        declare const text: string;

        textIdentity(...text);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      6,
					Column:    22,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
        declare function textIdentity(...args: string[]);

        declare const text: string;

        textIdentity(...text, 'and', ...text);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      6,
					Column:    22,
					EndColumn: 29,
				},
				{
					MessageId: "noStringSpread",
					Line:      6,
					Column:    38,
					EndColumn: 45,
				},
			},
		},
		{
			Code: `
        declare function textIdentity(...args: string[]);

        function withText<Text extends string>(text: Text) {
          textIdentity(...text);
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      5,
					Column:    24,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        declare function getString<T extends string>(): T;
        const a = [...getString()];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
        declare function getString(): string & { __brand: 'test' };
        const a = [...getString()];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStringSpread",
					Line:      3,
					Column:    20,
					EndColumn: 34,
				},
			},
		},
		{
			Code: "const o = { ...[1, 2, 3] };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      1,
					Column:    13,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        const arr = [1, 2, 3];
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        const arr = [1, 2, 3] as const;
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const arr: number[];
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const arr: readonly number[];
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const arr: number[] | string[];
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const arr: number[] & string[];
        const o = { ...arr };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare function getArray(): number[];
        const o = { ...getArray() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
        declare function getArray(): readonly number[];
        const o = { ...getArray() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 34,
				},
			},
		},
		{
			Code: "const o = { ...new Set([1, 2, 3]) };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      1,
					Column:    13,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
        const set = new Set([1, 2, 3]);
        const o = { ...set };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const set: Set<number>;
        const o = { ...set };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const set: WeakSet<object>;
        const o = { ...set };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const set: ReadonlySet<number>;
        const o = { ...set };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare const set: Set<number> | { a: number };
        const o = { ...set };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare function getSet(): Set<number>;
        const o = { ...getSet() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        const o = {
          ...new Map([
            ['test-1', 1],
            ['test-2', 2],
          ]),
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    11,
					EndLine:   6,
					EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        const o = 
          Object.fromEntries(new Map([
            ['test-1', 1],
            ['test-2', 2],
          ]))
        ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        const map = new Map([
          ['test-1', 1],
          ['test-2', 2],
        ]);

        const o = { ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      7,
					Column:    21,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        const map = new Map([
          ['test-1', 1],
          ['test-2', 2],
        ]);

        const o =  Object.fromEntries(map) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;
        const o = { ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;
        const o =  Object.fromEntries(map) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;
        const o = { ...(map) };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;
        const o =  Object.fromEntries((map)) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;
        const o = { ...(map, map) };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;
        const o =  Object.fromEntries((map, map)) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;
        const others = { a: 1 };
        const o = { ...map, ...others };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;
        const others = { a: 1 };
        const o = { ...Object.fromEntries(map), ...others };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;
        const o = { other: 1, ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    31,
					EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;
        const o = { other: 1, ...Object.fromEntries(map) };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: ReadonlyMap<string, number>;
        const o = { ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: ReadonlyMap<string, number>;
        const o =  Object.fromEntries(map) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: WeakMap<{ a: number }, string>;
        const o = { ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: WeakMap<{ a: number }, string>;
        const o =  Object.fromEntries(map) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number> | { a: number };
        const o = { ...map };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        declare function getMap(): Map<string, number>;
        const o = { ...getMap() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare function getMap(): Map<string, number>;
        const o =  Object.fromEntries(getMap()) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const a: Map<boolean, string> & Set<number>;
        const o = { ...a };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const a: Map<boolean, string> & Set<number>;
        const o =  Object.fromEntries(a) ;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        const ref = new WeakRef({ a: 1 });
        const o = { ...ref };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        const promise = new Promise(() => {});
        const o = { ...promise };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        const promise = new Promise(() => {});
        const o = { ...await promise };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const promise: Promise<{ a: 1 }>;
        async function foo() {
          return { ...(promise || {}) };
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      4,
					Column:    20,
					EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare const promise: Promise<{ a: 1 }>;
        async function foo() {
          return { ...(await (promise || {})) };
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const promise: Promise<any>;
        async function foo() {
          return { ...(Math.random() < 0.5 ? promise : {}) };
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      4,
					Column:    20,
					EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare const promise: Promise<any>;
        async function foo() {
          return { ...(await (Math.random() < 0.5 ? promise : {})) };
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        function withPromise<P extends Promise<void>>(promise: P) {
          return { ...promise };
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    20,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        function withPromise<P extends Promise<void>>(promise: P) {
          return { ...await promise };
        }
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const maybePromise: Promise<number> | { a: number };
        const o = { ...maybePromise };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare const maybePromise: Promise<number> | { a: number };
        const o = { ...await maybePromise };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare const promise: Promise<number> & { a: number };
        const o = { ...promise };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare const promise: Promise<number> & { a: number };
        const o = { ...await promise };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function getPromise(): Promise<number>;
        const o = { ...getPromise() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare function getPromise(): Promise<number>;
        const o = { ...await getPromise() };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        declare function getPromise<T extends Promise<number>>(arg: T): T;
        const o = { ...getPromise() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        declare function getPromise<T extends Promise<number>>(arg: T): T;
        const o = { ...await getPromise() };
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        function f() {}

        const o = { ...f };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        interface FunctionWithProps {
          (): string;
          prop: boolean;
        }

        type FunctionWithoutProps = () => string;

        declare const obj: FunctionWithProps | FunctionWithoutProps | object;

        const o = { ...obj };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      11,
					Column:    21,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
        const f = () => {};

        const o = { ...f };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        declare function f(): void;

        const o = { ...f };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        declare function getFunction(): () => void;

        const o = { ...getFunction() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
        declare const f: () => void;

        const o = { ...f };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        declare const f: () => void | { a: number };

        const o = { ...f };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        function* generator() {}

        const o = { ...generator };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 33,
				},
			},
		},
		{
			Code: `
        const iterator = {
          *[Symbol.iterator]() {
            yield 'test';
          },
        };

        const o = { ...iterator };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        type CustomIterable = {
          [Symbol.iterator]: () => Generator<string>;
        };

        const iterator: CustomIterable = {
          *[Symbol.iterator]() {
            yield 'test';
          },
        };

        const a = { ...iterator };
      `,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "file", "name": "AnotherIterable"}]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      12,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        declare module 'module' {
          export type CustomIterable = {
            [Symbol.iterator]: () => string;
          };
        }

        import { CustomIterable } from 'module';

        declare const iterator: CustomIterable;

        const a = { ...iterator };
      `,
			// TODO(port): for some reason tsgo returns `error` type for iterator
			Skip:    true,
			Options: rule_tester.OptionsFromJSON[NoMisusedSpreadOptions](`{"allow": [{"from": "package", "name": "Nothing", "package": "module"}]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      12,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        declare const iterator: Iterable<string>;

        const o = { ...iterator };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        declare const iterator: Iterable<string> | { a: number };

        const o = { ...iterator };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
        declare function getIterable(): Iterable<string>;

        const o = { ...getIterable() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
        class A {
          [Symbol.iterator]() {
            return {
              next() {
                return { done: true, value: undefined };
              },
            };
          }
        }

        const a = { ...new A() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      12,
					Column:    21,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        const o = { ...new Date() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      2,
					Column:    21,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
        declare class HTMLElementLike {}
        declare const element: HTMLElementLike;
        const o = { ...element };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        declare const regex: RegExp;
        const o = { ...regex };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      3,
					Column:    21,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
          public b = 2;
          private c = 3;
          protected d = 4;
          static e = 5;
        }

        const o = { ...new A() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      10,
					Column:    21,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        const a = new A();

        const o = { ...a };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        declare const a: A;

        const o = { ...a };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        declare function getA(): A;

        const o = { ...getA() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        declare function getA<T extends A>(arg: T): T;

        const o = { ...getA() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        class B extends A {}

        const o = { ...new B() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        declare const a: A | { b: string };

        const o = { ...a };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        class A {
          a = 1;
        }

        declare const a: A & { b: string };

        const o = { ...a };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      8,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        class A {}

        const o = { ...A };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        const A = class {};

        const o = { ...A };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        class Declaration {
          declaration?: boolean;
        }
        const Expression = class {
          expression?: boolean;
        };

        declare const either: typeof Declaration | typeof Expression;

        const o = { ...either };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      11,
					Column:    21,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        const A = Set<number>;

        const o = { ...A };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      4,
					Column:    21,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
        const a = {
          ...class A {
            static value = 1;
            nonStatic = 2;
          },
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      3,
					Column:    11,
					EndLine:   6,
					EndColumn: 12,
				},
			},
		},
		{
			Code: `
        const a = { ...(class A { static value = 1 }) }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      2,
					Column:    21,
					EndColumn: 54,
				},
			},
		},
		{
			Code: `
        const a = { ...new (class A { static value = 1; })() };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      2,
					Column:    21,
					EndColumn: 61,
				},
			},
		},
		{
			Code: `
        const o = <div {...[1, 2, 3]} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noArraySpreadInObject",
					Line:      2,
					Column:    24,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `
        class A {}

        const o = <div {...A} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassDeclarationSpreadInObject",
					Line:      4,
					Column:    24,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        const o = <div {...new Date()} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noClassInstanceSpreadInObject",
					Line:      2,
					Column:    24,
					EndColumn: 39,
				},
			},
		},
		{
			Code: `
        function f() {}

        const o = <div {...f} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionSpreadInObject",
					Line:      4,
					Column:    24,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
        const o = <div {...new Set([1, 2, 3])} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIterableSpreadInObject",
					Line:      2,
					Column:    24,
					EndColumn: 47,
				},
			},
		},
		{
			Code: `
        declare const map: Map<string, number>;

        const o = <div {...map} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMapSpreadInObject",
					Line:      4,
					Column:    24,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceMapSpreadInObject",
							Output: `
        declare const map: Map<string, number>;

        const o = <div {...Object.fromEntries(map)} />;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
        const promise = new Promise(() => {});

        const o = <div {...promise} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noPromiseSpreadInObject",
					Line:      4,
					Column:    24,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "addAwait",
							Output: `
        const promise = new Promise(() => {});

        const o = <div {...await promise} />;
      `,
						},
					},
				},
			},
		},
	})
}
