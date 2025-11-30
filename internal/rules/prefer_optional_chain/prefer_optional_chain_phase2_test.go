package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Phase 2: Comparison Ending Tests - Complete Set (Task 4.1-4.6)
func TestPreferOptionalChainPhase2Comparisons(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Task 4.6: Valid cases with undeclared variables
		{Code: `foo && foo.bar == undeclaredVar;`},
		{Code: `foo && foo.bar === undeclaredVar;`},
		{Code: `foo && foo.bar !== undeclaredVar;`},
		{Code: `foo && foo.bar != undeclaredVar;`},
		{Code: `foo != null && foo.bar == undeclaredVar;`},
		{Code: `foo != null && foo.bar === undeclaredVar;`},
		{Code: `foo != null && foo.bar !== undeclaredVar;`},
		{Code: `foo != null && foo.bar != undeclaredVar;`},
		{Code: `!foo || foo.bar != undeclaredVar;`},
		{Code: `!foo || foo.bar === undeclaredVar;`},
		{Code: `!foo || foo.bar !== undeclaredVar;`},
		{Code: `foo == null || foo.bar != undeclaredVar;`},
		{Code: `foo == null || foo.bar === undeclaredVar;`},
		{Code: `foo == null || foo.bar !== undeclaredVar;`},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Task 4.1: Negated OR with comparisons (20 cases)
		// ============================================================
		{
			Code:   `!foo || foo.bar != 1;`,
			Output: []string{`foo?.bar != 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar != '123';`,
			Output: []string{`foo?.bar != '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar != {};`,
			Output: []string{`foo?.bar != {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar != false;`,
			Output: []string{`foo?.bar != false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar != true;`,
			Output: []string{`foo?.bar != true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar === undefined;`,
			Output: []string{`foo?.bar === undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar == undefined;`,
			Output: []string{`foo?.bar == undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar == null;`,
			Output: []string{`foo?.bar == null;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== 0;`,
			Output: []string{`foo?.bar !== 0;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== 1;`,
			Output: []string{`foo?.bar !== 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== '123';`,
			Output: []string{`foo?.bar !== '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== {};`,
			Output: []string{`foo?.bar !== {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== false;`,
			Output: []string{`foo?.bar !== false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `!foo || foo.bar !== true;`,
			Output: []string{`foo?.bar !== true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// Task 4.2: Nullish OR with comparisons (20 cases)
		// ============================================================
		{
			Code:   `foo == null || foo.bar != 1;`,
			Output: []string{`foo?.bar != 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar != '123';`,
			Output: []string{`foo?.bar != '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar != {};`,
			Output: []string{`foo?.bar != {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar != false;`,
			Output: []string{`foo?.bar != false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar != true;`,
			Output: []string{`foo?.bar != true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar === undefined;`,
			Output: []string{`foo?.bar === undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar == undefined;`,
			Output: []string{`foo?.bar == undefined;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== 0;`,
			Output: []string{`foo?.bar !== 0;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== 1;`,
			Output: []string{`foo?.bar !== 1;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== '123';`,
			Output: []string{`foo?.bar !== '123';`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== {};`,
			Output: []string{`foo?.bar !== {};`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== false;`,
			Output: []string{`foo?.bar !== false;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo == null || foo.bar !== true;`,
			Output: []string{`foo?.bar !== true;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// Task 4.3: Typed declarations with all comparison operators (40 cases)
		// ============================================================
		{
			Code: `
				declare const foo: { bar: number };
				!foo || foo.bar == null;
			`,
			Output: []string{`
				declare const foo: { bar: number };
				foo?.bar == null;
			`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// Task 4.4: Yoda conditions
		// ============================================================
		{
			Code:   `foo != null && null != foo.bar && '123' == foo.bar.baz;`,
			Output: []string{`'123' == foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && null != foo.bar && '123' === foo.bar.baz;`,
			Output: []string{`'123' === foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `foo != null && null != foo.bar && undefined !== foo.bar.baz;`,
			Output: []string{`undefined !== foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ============================================================
		// Task 4.5: Typeof yoda conditions
		// ============================================================
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
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
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
