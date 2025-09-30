// Code generated from strict-boolean-expressions.test.ts - DO NOT EDIT.

package strict_boolean_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	//"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestStrictBooleanExpressionsSingleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{

	}, []rule_tester.InvalidTestCase{
		{
			Code: `
      function asserts1(x: string | number | undefined): asserts x {}
    function asserts2(x: string | number | undefined): asserts x {}

    const maybeString = Math.random() ? 'string'.slice() : undefined;

    const someAssert: typeof asserts1 | typeof asserts2 =
    Math.random() > 0.5 ? asserts1 : asserts2;

    someAssert(maybeString);
    `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNullableString"},
			}, /* Suggestions: conditionFixCompareNullish, conditionFixDefaultEmptyString, conditionFixCastBoolean */
		},
	})
}
