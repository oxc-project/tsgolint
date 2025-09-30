// Code generated from strict-boolean-expressions.test.ts - DO NOT EDIT.

package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsSingleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{

	}, []rule_tester.InvalidTestCase{
		{
			Code: `
      enum ExampleEnum {
      This = 0,
      That = 'one',
    }
    (value?: ExampleEnum) => (value ? 1 : 0);
    `,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableEnum: utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableEnum", Line: 6},
			}, /* Suggestions: conditionFixCompareNullish */
		},
	})
}
