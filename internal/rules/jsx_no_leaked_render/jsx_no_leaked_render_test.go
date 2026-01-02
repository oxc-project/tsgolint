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

			// String conditions - safe (empty string doesn't render visibly)
			{Code: `declare const str: string; <div>{str && <span/>}</div>`, Tsx: true},

			// Non-falsy number literals - safe
			{Code: `declare const x: 1 | 2 | 3; <div>{x && <span/>}</div>`, Tsx: true},

			// Ternary - safe
			{Code: `declare const count: number; <div>{count ? <span/> : null}</div>`, Tsx: true},

			// Outside JSX - not our concern
			{Code: `declare const count: number; const x = count && "hello";`},

			// Object/array - safe (even empty objects are truthy)
			{Code: `declare const obj: object; <div>{obj && <span/>}</div>`, Tsx: true},

			// null/undefined - safe (they don't render)
			{Code: `declare const x: null; <div>{x && <span/>}</div>`, Tsx: true},
			{Code: `declare const x: undefined; <div>{x && <span/>}</div>`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Generic number type
			{
				Code: `declare const count: number; <div>{count && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// Array length
			{
				Code: `declare const items: string[]; <div>{items.length && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// any type
			{
				Code: `declare const x: any; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// Union with number
			{
				Code: `declare const x: number | string; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// Generic constrained to number
			{
				Code: `function Foo<T extends number>(x: T) { return <div>{x && <span/>}</div> }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// Literal 0
			{
				Code: `declare const x: 0; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
			// bigint (can also be 0n which renders "0")
			{
				Code: `declare const x: bigint; <div>{x && <span/>}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLeakedConditionalRendering", Line: 1},
				},
			},
		})
}
