package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamBaseCasesAndOperator tests all 26 base cases from typescript-eslint with && operator
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/prefer-optional-chain/base-cases.ts
func TestUpstreamBaseCasesAndOperator(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Base case 1: chained members
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
foo && foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 2: nested chained members
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar && foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 3: optional call
			{
				Code: `// 3
declare const foo: (() => number) | null | undefined;
foo && foo();`,
				Output: []string{`// 3
declare const foo: (() => number) | null | undefined;
foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 4: optional method call
			{
				Code: `// 4
declare const foo: {bar: (() => number) | null | undefined};
foo.bar && foo.bar();`,
				Output: []string{`// 4
declare const foo: {bar: (() => number) | null | undefined};
foo.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 5: long chain
			{
				Code: `// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz;`,
				Output: []string{`// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 6: partial chain
			{
				Code: `// 6
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};
foo.bar && foo.bar.baz && foo.bar.baz.buzz;`,
				Output: []string{`// 6
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};
foo.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 7: case with a jump (non-nullish prop)
			{
				Code: `// 7
declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz.buzz;`,
				Output: []string{`// 7
declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;
foo?.bar?.baz.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 8: jump with partial chain
			{
				Code: `// 8
declare const foo: {bar: {baz: {buzz: number}} | null | undefined};
foo.bar && foo.bar.baz.buzz;`,
				Output: []string{`// 8
declare const foo: {bar: {baz: {buzz: number}} | null | undefined};
foo.bar?.baz.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 9: doubled up expression
			{
				Code: `// 9
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz && foo.bar.baz.buzz;`,
				Output: []string{`// 9
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 10: doubled up expression with partial chain
			{
				Code: `// 10
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo.bar && foo.bar.baz && foo.bar.baz && foo.bar.baz.buzz;`,
				Output: []string{`// 10
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 11: chained members with element access
			{
				Code: `// 11
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo && foo[bar] && foo[bar].baz && foo[bar].baz.buzz;`,
				Output: []string{`// 11
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.[bar]?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 12: element access with jump
			{
				Code: `// 12
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo && foo[bar].baz && foo[bar].baz.buzz;`,
				Output: []string{`// 12
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
foo?.[bar].baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 13: property access in computed property
			{
				Code: `// 13
declare const bar: {baz: string};
declare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;
foo && foo[bar.baz] && foo[bar.baz].buzz;`,
				Output: []string{`// 13
declare const bar: {baz: string};
declare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;
foo?.[bar.baz]?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 14: chained calls
			{
				Code: `// 14
declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz();`,
				Output: []string{`// 14
declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 15: chained calls with optional call
			{
				Code: `// 15
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz && foo.bar.baz.buzz();`,
				Output: []string{`// 15
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 16: partial chain with optional call
			{
				Code: `// 16
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};
foo.bar && foo.bar.baz && foo.bar.baz.buzz && foo.bar.baz.buzz();`,
				Output: []string{`// 16
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};
foo.bar?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 17: call with jump
			{
				Code: `// 17
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz.buzz();`,
				Output: []string{`// 17
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;
foo?.bar?.baz.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 18: partial chain call with jump
			{
				Code: `// 18
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};
foo.bar && foo.bar.baz.buzz();`,
				Output: []string{`// 18
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};
foo.bar?.baz.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 19: jump with optional call
			{
				Code: `// 19
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz.buzz && foo.bar.baz.buzz();`,
				Output: []string{`// 19
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;
foo?.bar?.baz.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 20: call expr inside chain
			{
				Code: `// 20
declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};
foo.bar && foo.bar() && foo.bar().baz && foo.bar().baz.buzz && foo.bar().baz.buzz();`,
				Output: []string{`// 20
declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};
foo.bar?.()?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 21: chained calls with element access
			{
				Code: `// 21
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz[buzz]();`,
				Output: []string{`// 21
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.[buzz]();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 22: element access with optional call
			{
				Code: `// 22
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo && foo.bar && foo.bar.baz && foo.bar.baz[buzz] && foo.bar.baz[buzz]();`,
				Output: []string{`// 22
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.[buzz]?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 23: partially pre-optional chained
			{
				Code: `// 23
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo && foo?.bar && foo?.bar.baz && foo?.bar.baz[buzz] && foo?.bar.baz[buzz]();`,
				Output: []string{`// 23
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
foo?.bar?.baz?.[buzz]?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 24: partial optional chain with element access
			{
				Code: `// 24
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;
foo && foo?.bar.baz && foo?.bar.baz[buzz];`,
				Output: []string{`// 24
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;
foo?.bar.baz?.[buzz];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 25: optional call with property access
			{
				Code: `// 25
declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;
foo && foo?.() && foo?.().bar;`,
				Output: []string{`// 25
declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;
foo?.()?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 26: method optional call with property access
			{
				Code: `// 26
declare const foo: {bar: () => ({baz: number} | null | undefined)};
foo.bar && foo.bar?.() && foo.bar?.().baz;`,
				Output: []string{`// 26
declare const foo: {bar: () => ({baz: number} | null | undefined)};
foo.bar?.()?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}

// TestUpstreamBaseCasesOrOperator tests all 26 base cases from typescript-eslint with || operator
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/prefer-optional-chain/base-cases.ts
func TestUpstreamBaseCasesOrOperator(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Base case 1: chained members with ||
			{
				Code: `// 1
declare const foo: {bar: number} | null | undefined;
!foo || !foo.bar;`,
				Output: []string{`// 1
declare const foo: {bar: number} | null | undefined;
!foo?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 2: nested chained members with ||
			{
				Code: `// 2
declare const foo: {bar: {baz: number} | null | undefined};
!foo.bar || !foo.bar.baz;`,
				Output: []string{`// 2
declare const foo: {bar: {baz: number} | null | undefined};
!foo.bar?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 3: optional call with ||
			{
				Code: `// 3
declare const foo: (() => number) | null | undefined;
!foo || !foo();`,
				Output: []string{`// 3
declare const foo: (() => number) | null | undefined;
!foo?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 4: optional method call with ||
			{
				Code: `// 4
declare const foo: {bar: (() => number) | null | undefined};
!foo.bar || !foo.bar();`,
				Output: []string{`// 4
declare const foo: {bar: (() => number) | null | undefined};
!foo.bar?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 5: long chain with ||
			{
				Code: `// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz.buzz;`,
				Output: []string{`// 5
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 6: partial chain with ||
			{
				Code: `// 6
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};
!foo.bar || !foo.bar.baz || !foo.bar.baz.buzz;`,
				Output: []string{`// 6
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};
!foo.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 7: case with a jump (non-nullish prop) with ||
			{
				Code: `// 7
declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz.buzz;`,
				Output: []string{`// 7
declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;
!foo?.bar?.baz.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 8: jump with partial chain with ||
			{
				Code: `// 8
declare const foo: {bar: {baz: {buzz: number}} | null | undefined};
!foo.bar || !foo.bar.baz.buzz;`,
				Output: []string{`// 8
declare const foo: {bar: {baz: {buzz: number}} | null | undefined};
!foo.bar?.baz.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 9: doubled up expression with ||
			{
				Code: `// 9
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz || !foo.bar.baz.buzz;`,
				Output: []string{`// 9
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 10: doubled up expression with partial chain with ||
			{
				Code: `// 10
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo.bar || !foo.bar.baz || !foo.bar.baz || !foo.bar.baz.buzz;`,
				Output: []string{`// 10
declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo.bar?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 11: chained members with element access with ||
			{
				Code: `// 11
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo[bar] || !foo[bar].baz || !foo[bar].baz.buzz;`,
				Output: []string{`// 11
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo?.[bar]?.baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 12: element access with jump with ||
			{
				Code: `// 12
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo[bar].baz || !foo[bar].baz.buzz;`,
				Output: []string{`// 12
declare const bar: string;
declare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;
!foo?.[bar].baz?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 13: property access in computed property with ||
			{
				Code: `// 13
declare const bar: {baz: string};
declare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;
!foo || !foo[bar.baz] || !foo[bar.baz].buzz;`,
				Output: []string{`// 13
declare const bar: {baz: string};
declare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;
!foo?.[bar.baz]?.buzz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 14: chained calls with ||
			{
				Code: `// 14
declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz.buzz();`,
				Output: []string{`// 14
declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 15: chained calls with optional call with ||
			{
				Code: `// 15
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz.buzz || !foo.bar.baz.buzz();`,
				Output: []string{`// 15
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 16: partial chain with optional call with ||
			{
				Code: `// 16
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};
!foo.bar || !foo.bar.baz || !foo.bar.baz.buzz || !foo.bar.baz.buzz();`,
				Output: []string{`// 16
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};
!foo.bar?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 17: call with jump with ||
			{
				Code: `// 17
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz.buzz();`,
				Output: []string{`// 17
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;
!foo?.bar?.baz.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 18: partial chain call with jump with ||
			{
				Code: `// 18
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};
!foo.bar || !foo.bar.baz.buzz();`,
				Output: []string{`// 18
declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};
!foo.bar?.baz.buzz();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 19: jump with optional call with ||
			{
				Code: `// 19
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz.buzz || !foo.bar.baz.buzz();`,
				Output: []string{`// 19
declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;
!foo?.bar?.baz.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 20: call expr inside chain with ||
			{
				Code: `// 20
declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};
!foo.bar || !foo.bar() || !foo.bar().baz || !foo.bar().baz.buzz || !foo.bar().baz.buzz();`,
				Output: []string{`// 20
declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};
!foo.bar?.()?.baz?.buzz?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 21: chained calls with element access with ||
			{
				Code: `// 21
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz[buzz]();`,
				Output: []string{`// 21
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.[buzz]();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 22: element access with optional call with ||
			{
				Code: `// 22
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo.bar || !foo.bar.baz || !foo.bar.baz[buzz] || !foo.bar.baz[buzz]();`,
				Output: []string{`// 22
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.[buzz]?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 23: partially pre-optional chained with ||
			{
				Code: `// 23
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo || !foo?.bar || !foo?.bar.baz || !foo?.bar.baz[buzz] || !foo?.bar.baz[buzz]();`,
				Output: []string{`// 23
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;
!foo?.bar?.baz?.[buzz]?.();`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 24: partial optional chain with element access with ||
			{
				Code: `// 24
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;
!foo || !foo?.bar.baz || !foo?.bar.baz[buzz];`,
				Output: []string{`// 24
declare const buzz: string;
declare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;
!foo?.bar.baz?.[buzz];`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 25: optional call with property access with ||
			{
				Code: `// 25
declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;
!foo || !foo?.() || !foo?.().bar;`,
				Output: []string{`// 25
declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;
!foo?.()?.bar;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
			// Base case 26: method optional call with property access with ||
			{
				Code: `// 26
declare const foo: {bar: () => ({baz: number} | null | undefined)};
!foo.bar || !foo.bar?.() || !foo.bar?.().baz;`,
				Output: []string{`// 26
declare const foo: {bar: () => ({baz: number} | null | undefined)};
!foo.bar?.()?.baz;`},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			},
		})
}
