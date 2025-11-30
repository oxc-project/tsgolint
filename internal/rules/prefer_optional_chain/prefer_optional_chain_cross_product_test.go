package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestCrossProductVariations tests systematic cross-product of patterns × operators × contexts
// This provides comprehensive coverage through systematic combination
func TestCrossProductVariations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// === Pattern: foo.bar × All comparison operators at end ===

			{
				Code:   `foo.bar && foo.bar.baz === 0;`,
				Output: []string{`foo.bar?.baz === 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz !== 0;`,
				Output: []string{`foo.bar?.baz !== 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz == 0;`,
				Output: []string{`foo.bar?.baz == 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz != 0;`,
				Output: []string{`foo.bar?.baz != 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz > 0;`,
				Output: []string{`foo.bar?.baz > 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz < 100;`,
				Output: []string{`foo.bar?.baz < 100;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz >= 0;`,
				Output: []string{`foo.bar?.baz >= 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar && foo.bar.baz <= 100;`,
				Output: []string{`foo.bar?.baz <= 100;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Pattern: foo[bar] × All comparison operators ===

			{
				Code:   `foo[bar] && foo[bar].baz === 0;`,
				Output: []string{`foo[bar]?.baz === 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] && foo[bar].baz !== 0;`,
				Output: []string{`foo[bar]?.baz !== 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] && foo[bar].baz > 0;`,
				Output: []string{`foo[bar]?.baz > 0;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] && foo[bar].baz < 100;`,
				Output: []string{`foo[bar]?.baz < 100;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Pattern: foo() × All member access types ===


			// === All nullish operators × simple chain ===

			// != null
			{
				Code:   `foo != null && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar != null && foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] != null && foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// !== null
			{
				Code:   `foo !== null && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar !== null && foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] !== null && foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// != undefined
			{
				Code:   `foo != undefined && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar != undefined && foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] != undefined && foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// !== undefined
			{
				Code:   `foo !== undefined && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar !== undefined && foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] !== undefined && foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === OR operator variants × simple chain ===

			// == null (with OR)
			{
				Code:   `foo == null || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar == null || foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] == null || foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === null (with OR)
			{
				Code:   `foo === null || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar === null || foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] === null || foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// == undefined (with OR)
			{
				Code:   `foo == undefined || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar == undefined || foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] == undefined || foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === undefined (with OR)
			{
				Code:   `foo === undefined || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo.bar === undefined || foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo[bar] === undefined || foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === typeof checks × all access patterns ===

			{
				Code:   `typeof foo !== 'undefined' && foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `typeof foo.bar !== 'undefined' && foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `typeof foo[bar] !== 'undefined' && foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// typeof with OR
			{
				Code:   `typeof foo === 'undefined' || foo.bar;`,
				Output: []string{`foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `typeof foo.bar === 'undefined' || foo.bar.baz;`,
				Output: []string{`foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `typeof foo[bar] === 'undefined' || foo[bar].baz;`,
				Output: []string{`foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Negation (!) with OR × all access patterns ===

			{
				Code:   `!foo || !foo.bar;`,
				Output: []string{`!foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `!foo.bar || !foo.bar.baz;`,
				Output: []string{`!foo.bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `!foo[bar] || !foo[bar].baz;`,
				Output: []string{`!foo[bar]?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in if statements ===

			{
				Code:   `if (foo && foo.bar) {}`,
				Output: []string{`if (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `if (foo != null && foo.bar) {}`,
				Output: []string{`if (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `if (foo !== undefined && foo.bar) {}`,
				Output: []string{`if (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `if (typeof foo !== 'undefined' && foo.bar) {}`,
				Output: []string{`if (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `if (!foo || !foo.bar) {}`,
				Output: []string{`if (!foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in while statements ===

			{
				Code:   `while (foo && foo.bar) {}`,
				Output: []string{`while (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `while (foo != null && foo.bar) {}`,
				Output: []string{`while (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `while (foo !== undefined && foo.bar) {}`,
				Output: []string{`while (foo?.bar) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in return statements ===

			{
				Code:   `return foo && foo.bar;`,
				Output: []string{`return foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `return foo != null && foo.bar;`,
				Output: []string{`return foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `return foo !== undefined && foo.bar;`,
				Output: []string{`return foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `return typeof foo !== 'undefined' && foo.bar;`,
				Output: []string{`return foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `return !foo || !foo.bar;`,
				Output: []string{`return !foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in variable declarations ===

			{
				Code:   `const x = foo && foo.bar;`,
				Output: []string{`const x = foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `let x = foo != null && foo.bar;`,
				Output: []string{`let x = foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `var x = foo !== undefined && foo.bar;`,
				Output: []string{`var x = foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `const x = typeof foo !== 'undefined' && foo.bar;`,
				Output: []string{`const x = foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in function arguments ===

			{
				Code:   `fn(foo && foo.bar);`,
				Output: []string{`fn(foo?.bar);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `fn(foo != null && foo.bar);`,
				Output: []string{`fn(foo?.bar);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `fn(foo !== undefined && foo.bar);`,
				Output: []string{`fn(foo?.bar);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `fn(typeof foo !== 'undefined' && foo.bar);`,
				Output: []string{`fn(foo?.bar);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in array literals ===

			{
				Code:   `[foo && foo.bar];`,
				Output: []string{`[foo?.bar];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `[foo != null && foo.bar];`,
				Output: []string{`[foo?.bar];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `[foo !== undefined && foo.bar];`,
				Output: []string{`[foo?.bar];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === All patterns in object literals ===

			{
				Code:   `({prop: foo && foo.bar});`,
				Output: []string{`({prop: foo?.bar});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `({prop: foo != null && foo.bar});`,
				Output: []string{`({prop: foo?.bar});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `({prop: foo !== undefined && foo.bar});`,
				Output: []string{`({prop: foo?.bar});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === 3-level deep × all nullish operators ===

			{
				Code:   `foo && foo.bar && foo.bar.baz && foo.bar.baz.qux;`,
				Output: []string{`foo?.bar?.baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo != null && foo.bar != null && foo.bar.baz != null && foo.bar.baz.qux;`,
				Output: []string{`foo?.bar?.baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo !== null && foo.bar !== null && foo.bar.baz !== null && foo.bar.baz.qux;`,
				Output: []string{`foo?.bar?.baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo !== undefined && foo.bar !== undefined && foo.bar.baz !== undefined && foo.bar.baz.qux;`,
				Output: []string{`foo?.bar?.baz?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Call chains × all operators ===


			// === Element access chains × all operators ===

			{
				Code:   `foo && foo[bar] && foo[bar][baz];`,
				Output: []string{`foo?.[bar]?.[baz];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo != null && foo[bar] != null && foo[bar][baz];`,
				Output: []string{`foo?.[bar]?.[baz];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo !== undefined && foo[bar] !== undefined && foo[bar][baz];`,
				Output: []string{`foo?.[bar]?.[baz];`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Mixed access types × operators ===

			{
				Code:   `foo && foo.bar && foo.bar[baz] && foo.bar[baz].qux;`,
				Output: []string{`foo?.bar?.[baz]?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo != null && foo.bar != null && foo.bar[baz] != null && foo.bar[baz].qux;`,
				Output: []string{`foo?.bar?.[baz]?.qux;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo && foo[bar] && foo[bar].baz && foo[bar].baz();`,
				Output: []string{`foo?.[bar]?.baz?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `foo != null && foo[bar] != null && foo[bar].baz != null && foo[bar].baz();`,
				Output: []string{`foo?.[bar]?.baz?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
