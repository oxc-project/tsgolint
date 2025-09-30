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
		{
			Code: `
      declare const x: string[] | null;
    // oxlint-disable-next-line
    if (x) {
    }
    `,
			TSConfig: "tsconfig.unstrict.json",
			Options: StrictBooleanExpressionsOptions{
				AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(true),
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
      declare const x: string[] | null;
    if (x) {
    }
    `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck", Line: 0},
				{MessageId: "unexpectedObject", Line: 3},
			},
		},
	})
}
