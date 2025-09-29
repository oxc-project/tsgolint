// Code generated from strict-boolean-expressions.test.ts - DO NOT EDIT.

package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestStrictBooleanExpressionsSingleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{

	}, []rule_tester.InvalidTestCase{
		{
			Code: `
      function foo(x: 0 | 1 | null) {
      if (!x) {
    }
    }
    `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableNumber", Line: 3},
			}, /* Suggestions: conditionFixCompareNullish, conditionFixDefaultZero, conditionFixCastBoolean */
		},
	})
}
