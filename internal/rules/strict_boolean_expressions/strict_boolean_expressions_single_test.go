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
			Code: `if (true && 1 + 1) {}`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedNumber",
					Line:      1,
					Column:    13,
				},
			},
		},
	})
}