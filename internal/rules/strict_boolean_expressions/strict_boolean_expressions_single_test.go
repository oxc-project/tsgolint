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
			Code: `if (('' && {}) || (0 && void 0)) { }`,
			Options: StrictBooleanExpressionsOptions{
				AllowNullableObject: utils.Ref(false),
				AllowNumber:         utils.Ref(false),
				AllowString:         utils.Ref(false),
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedString", Line: 1},
				{MessageId: "unexpectedObjectContext", Line: 1},
				{MessageId: "unexpectedNumber", Line: 1},
				{MessageId: "unexpectedNullish", Line: 1},
			},
		},
	})
}