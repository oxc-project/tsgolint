package jsx_no_leaked_render

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestJsxNoLeakedRender(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoLeakedRenderRule,
		[]rule_tester.ValidTestCase{
			// Boolean conditions - safe
			{Code: `declare const isVisible: boolean; <div>{isVisible && <span/>}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{count > 0 && <span/>}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{!!count && <span/>}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{Boolean(count) && <span/>}</div>`, Tsx: true},

			// String conditions - safe in React 18+ (empty strings treated as null)
			// See: https://github.com/facebook/react/pull/22807
			{Code: `declare const str: string; <div>{str && <span/>}</div>`, Tsx: true},
			{Code: `declare const str: 'hello'; <div>{str && <span/>}</div>`, Tsx: true},
			{Code: `declare const str: 'a' | 'b' | 'c'; <div>{str && <span/>}</div>`, Tsx: true},
			{Code: `declare const str: ''; <div>{str && <span/>}</div>`, Tsx: true},
			{Code: `declare const str: '' | 'hello'; <div>{str && <span/>}</div>`, Tsx: true},
			{Code: `declare const str: string | null; <div>{str && <span/>}</div>`, Tsx: true},

			// Non-falsy number literals - safe
			{Code: `declare const x: 1 | 2 | 3; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: -1; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: 42; <div>{x && <span/>}</div>`, Tsx: true},

			// Ternary - safe (always evaluates to one branch)
			{Code: `declare const count: number; <div>{count ? <span/> : null}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{count ? <span/> : <empty/>}</div>`, Tsx: true},

			// Outside JSX - not our concern
			{Code: `declare const count: number; const x = count && "hello";`},
			{Code: `declare const count: number; if (count && true) {}`},

			// Object/array - safe (even empty objects are truthy)
			{Code: `declare const obj: object; <div>{obj && <span/>}</div>`, Tsx: true},
			{Code: `declare const arr: string[]; <div>{arr && <span/>}</div>`, Tsx: true},
			{Code: `declare const obj: { a: number }; <div>{obj && <span/>}</div>`, Tsx: true},

			// null/undefined - safe (they don't render)
			{Code: `declare const x: null; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: undefined; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: null | undefined; <div>{x && <span/>}</div>`, Tsx: true},

			// OR operator - not our concern (returns the truthy value)
			{Code: `declare const count: number; <div>{count || <span/>}</div>`, Tsx: true},

			// Negation with number - safe (results in boolean)
			{Code: `declare const count: number; <div>{!count && <span/>}</div>`, Tsx: true},

			// Comparison operators - safe (result is boolean)
			{Code: `declare const count: number; <div>{count >= 0 && <span/>}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{count !== 0 && <span/>}</div>`, Tsx: true},
			{Code: `declare const count: number; <div>{count === 0 && <span/>}</div>`, Tsx: true},
			{Code: `declare const items: string[]; <div>{items.length > 0 && <span/>}</div>`, Tsx: true},

			// Nullish coalescing is safe (coerces to the type, not used as condition)
			{Code: `declare const count: number | null; <div>{(count ?? 0) > 0 && <span/>}</div>`, Tsx: true},

			// Non-zero bigint literals - safe
			{Code: `declare const x: 1n; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: 100n; <div>{x && <span/>}</div>`, Tsx: true},

			// Function returning boolean - safe
			{Code: `declare const hasItems: () => boolean; <div>{hasItems() && <span/>}</div>`, Tsx: true},

			// Boolean union - safe
			{Code: `declare const x: true | false; <div>{x && <span/>}</div>`, Tsx: true},

			// Nested JSX - the inner expression is not checked as it's not our target
			{Code: `declare const show: boolean; <div>{show && <span>{show}</span>}</div>`, Tsx: true},

			// Type narrowing with boolean check
			{Code: `declare const x: number | null; <div>{x !== null && x > 0 && <span/>}</div>`, Tsx: true},

			// never type - safe
			{Code: `declare const x: never; <div>{x && <span/>}</div>`, Tsx: true},

			// Symbol type - safe (always truthy)
			{Code: `declare const x: symbol; <div>{x && <span/>}</div>`, Tsx: true},

			// Function type - safe (always truthy)
			{Code: `declare const fn: () => void; <div>{fn && <span/>}</div>`, Tsx: true},

			// Generic with default (not constraint) - T can be any type, so unconstrained
			{Code: `function Foo<T = number>(x: T) { return <div>{x && <span/>}</div> }`, Tsx: true},

			// Inferred const with truthy value - literal type 1
			{Code: `const t = 1; <div>{t && <span/>}</div>`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Generic number type
			{
				Code: `declare const count: number; <div>{count && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 36, EndColumn: 41},
				},
			},
			// Array length
			{
				Code: `declare const items: string[]; <div>{items.length && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 38, EndColumn: 50},
				},
			},
			// any type
			{
				Code: `declare const x: any; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 29, EndColumn: 30},
				},
			},
			// Union with number (string part is safe in React 18+)
			{
				Code: `declare const x: number | string; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 41, EndColumn: 42},
				},
			},
			// Generic constrained to number
			{
				Code: `function Foo<T extends number>(x: T) { return <div>{x && <span/>}</div> }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 53, EndColumn: 54},
				},
			},
			// Literal 0
			{
				Code: `declare const x: 0; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 27, EndColumn: 28},
				},
			},
			// bigint (can also be 0n which renders "0")
			{
				Code: `declare const x: bigint; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 32, EndColumn: 33},
				},
			},
			// Literal 0n
			{
				Code: `declare const x: 0n; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 28, EndColumn: 29},
				},
			},
			// Nested property access
			{
				Code: `
declare const data: { items: string[] };
<div>{data.items.length && <span/>}</div>
      `,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 3, Column: 7, EndColumn: 24},
				},
			},
			// Union including 0 literal
			{
				Code: `declare const x: 0 | 1 | 2; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 35, EndColumn: 36},
				},
			},
			// Number with null union - number part is still problematic
			{
				Code: `declare const x: number | null; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 39, EndColumn: 40},
				},
			},
			// Number with undefined union
			{
				Code: `declare const x: number | undefined; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 44, EndColumn: 45},
				},
			},
			// Generic constrained to bigint
			{
				Code: `function Foo<T extends bigint>(x: T) { return <div>{x && <span/>}</div> }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 53, EndColumn: 54},
				},
			},
			// Type assertion to number
			{
				Code: `declare const x: unknown; <div>{(x as number) && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 33, EndColumn: 46},
				},
			},
			// Intersection with number
			{
				Code: `declare const x: number & { brand: 'count' }; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 53, EndColumn: 54},
				},
			},
			// Complex multiline case
			{
				Code: `
declare const items: {
  nested: {
    count: number;
  };
};
<div>
  {items.nested.count && <span/>}
</div>
      `,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 8, Column: 4, EndColumn: 22},
				},
			},
			// Index access returning number
			{
				Code: `declare const counts: number[]; <div>{counts[0] && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 39, EndColumn: 48},
				},
			},
			// Method call returning number
			{
				Code: `declare const getCount: () => number; <div>{getCount() && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 45, EndColumn: 55},
				},
			},
			// Optional number property - could be number (0), null, or undefined
			{
				Code: `declare const rating: { count?: number | null }; <div>{rating.count && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 56, EndColumn: 68},
				},
			},
			// NaN - has type number, so caught
			{
				Code: `const t = NaN; <div>{t && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 22, EndColumn: 23},
				},
			},
			// Inferred const 0 - literal type 0
			{
				Code: `const t = 0; <div>{t && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1, Column: 20, EndColumn: 21},
				},
			},
		})
}
