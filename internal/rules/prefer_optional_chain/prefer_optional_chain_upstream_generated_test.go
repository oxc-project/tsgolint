package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamGeneratedVariations tests systematically generated variations
// These tests cover common patterns with different operators, types, and contexts
func TestUpstreamGeneratedVariations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// === Nullish check variations (!=, ==, !==, ===) with null ===

			// != null checks (loose equality)
			{
				Code:    `foo != null && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar != null && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo.bar != null && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar != null && foo.bar.baz != null && foo.bar.baz.buzz;`,
				Output:  []string{`foo.bar?.baz?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// == null checks (loose equality)
			{
				Code:    `foo == null || foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar == null || foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo == null || foo.bar == null || foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// !== null checks (strict inequality)
			{
				Code:    `foo !== null && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar !== null && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo !== null && foo.bar !== null && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === null checks (strict equality with OR)
			{
				Code:    `foo === null || foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar === null || foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Undefined check variations ===

			// != undefined checks (loose inequality)
			{
				Code:    `foo != undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar != undefined && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != undefined && foo.bar != undefined && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// == undefined checks (loose equality with OR)
			{
				Code:    `foo == undefined || foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar == undefined || foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// !== undefined checks (strict inequality)
			{
				Code:    `foo !== undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar !== undefined && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo !== undefined && foo.bar !== undefined && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === undefined checks (strict equality with OR)
			{
				Code:    `foo === undefined || foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar === undefined || foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === typeof checks ===

			// typeof !== 'undefined'
			{
				Code:    `typeof foo !== 'undefined' && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `typeof foo.bar !== 'undefined' && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `typeof foo !== 'undefined' && typeof foo.bar !== 'undefined' && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// typeof === 'undefined' with OR
			{
				Code:    `typeof foo === 'undefined' || foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `typeof foo.bar === 'undefined' || foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// typeof != 'undefined' (loose inequality)
			{
				Code:    `typeof foo != 'undefined' && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `typeof foo.bar != 'undefined' && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Yoda-style variations (value on left) ===

			// null !== foo (reversed)
			{
				Code:    `null !== foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `null !== foo.bar && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// undefined !== foo (reversed)
			{
				Code:    `undefined !== foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `undefined !== foo.bar && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// 'undefined' !== typeof (reversed)
			{
				Code:    `'undefined' !== typeof foo && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `'undefined' !== typeof foo.bar && foo.bar.baz;`,
				Output:  []string{`foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Element access variations ===

			// Simple element access
			{
				Code:    `foo && foo[bar];`,
				Output:  []string{`foo?.[bar];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar[baz];`,
				Output:  []string{`foo.bar?.[baz];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[bar] && foo[bar].baz;`,
				Output:  []string{`foo?.[bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Element access with nullish checks
			{
				Code:    `foo != null && foo[bar];`,
				Output:  []string{`foo?.[bar];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo[bar] != null && foo[bar].baz;`,
				Output:  []string{`foo[bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo[bar] != null && foo[bar].baz;`,
				Output:  []string{`foo?.[bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// String literal element access
			{
				Code:    `foo && foo['bar'];`,
				Output:  []string{`foo?.['bar'];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo['bar'] && foo['bar'].baz;`,
				Output:  []string{`foo?.['bar']?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Number literal element access
			{
				Code:    `foo && foo[0];`,
				Output:  []string{`foo?.[0];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo[0] && foo[0].bar;`,
				Output:  []string{`foo?.[0]?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Call expression variations ===

			// Simple call
			{
				Code:    `foo && foo();`,
				Output:  []string{`foo?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar();`,
				Output:  []string{`foo.bar?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo() && foo().bar;`,
				Output:  []string{`foo?.()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Call with nullish checks
			{
				Code:    `foo != null && foo();`,
				Output:  []string{`foo?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar != null && foo.bar();`,
				Output:  []string{`foo.bar?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Call with method chain
			{
				Code:    `foo && foo.bar() && foo.bar().baz;`,
				Output:  []string{`foo?.bar()?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar.baz() && foo.bar.baz().buzz;`,
				Output:  []string{`foo.bar?.baz()?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Mixed chains ===

			// Member -> Call -> Member
			{
				Code:    `foo && foo.bar && foo.bar() && foo.bar().baz;`,
				Output:  []string{`foo?.bar?.()?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Element -> Member
			{
				Code:    `foo && foo[bar] && foo[bar].baz && foo[bar].baz.buzz;`,
				Output:  []string{`foo?.[bar]?.baz?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Member -> Element -> Member
			{
				Code:    `foo && foo.bar && foo.bar[baz] && foo.bar[baz].buzz;`,
				Output:  []string{`foo?.bar?.[baz]?.buzz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Call -> Element
			{
				Code:    `foo && foo() && foo()[bar];`,
				Output:  []string{`foo?.()?.[bar];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === 4-level deep chains ===

			{
				Code:    `foo && foo.bar && foo.bar.baz && foo.bar.baz.buzz && foo.bar.baz.buzz.qux;`,
				Output:  []string{`foo?.bar?.baz?.buzz?.qux;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo.bar != null && foo.bar.baz != null && foo.bar.baz.buzz != null && foo.bar.baz.buzz.qux;`,
				Output:  []string{`foo?.bar?.baz?.buzz?.qux;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo !== undefined && foo.bar !== undefined && foo.bar.baz !== undefined && foo.bar.baz.buzz !== undefined && foo.bar.baz.buzz.qux;`,
				Output:  []string{`foo?.bar?.baz?.buzz?.qux;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === 5-level deep chains ===

			{
				Code:    `foo && foo.a && foo.a.b && foo.a.b.c && foo.a.b.c.d && foo.a.b.c.d.e;`,
				Output:  []string{`foo?.a?.b?.c?.d?.e;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo.a != null && foo.a.b != null && foo.a.b.c != null && foo.a.b.c.d != null && foo.a.b.c.d.e;`,
				Output:  []string{`foo?.a?.b?.c?.d?.e;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === OR operator variations ===

			// Simple OR with negation
			{
				Code:    `!foo || !foo.bar;`,
				Output:  []string{`!foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo.bar || !foo.bar.baz;`,
				Output:  []string{`!foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo || !foo.bar || !foo.bar.baz;`,
				Output:  []string{`!foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR with element access
			{
				Code:    `!foo || !foo[bar];`,
				Output:  []string{`!foo?.[bar];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo || !foo[bar] || !foo[bar].baz;`,
				Output:  []string{`!foo?.[bar]?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// OR with calls
			{
				Code:    `!foo || !foo();`,
				Output:  []string{`!foo?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `!foo.bar || !foo.bar();`,
				Output:  []string{`!foo.bar?.();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Combined null and undefined checks ===

			{
				Code:    `foo !== null && foo !== undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && foo !== undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo !== null && foo != undefined && foo.bar;`,
				Output:  []string{`foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Combined checks at multiple levels
			{
				Code:    `foo !== null && foo !== undefined && foo.bar !== null && foo.bar !== undefined && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo != null && typeof foo.bar !== 'undefined' && foo.bar != null && foo.bar.baz;`,
				Output:  []string{`foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Ending with comparison (should preserve comparison) ===

			{
				Code:    `foo && foo.bar !== null;`,
				Output:  []string{`foo?.bar !== null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar !== undefined;`,
				Output:  []string{`foo?.bar !== undefined;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar != null;`,
				Output:  []string{`foo?.bar != null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar.baz !== null;`,
				Output:  []string{`foo.bar?.baz !== null;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar && foo.bar.baz !== undefined;`,
				Output:  []string{`foo?.bar?.baz !== undefined;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Starting from deeper property ===

			{
				Code:    `obj.foo && obj.foo.bar;`,
				Output:  []string{`obj.foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.foo.bar && obj.foo.bar.baz;`,
				Output:  []string{`obj.foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.foo && obj.foo.bar && obj.foo.bar.baz;`,
				Output:  []string{`obj.foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
