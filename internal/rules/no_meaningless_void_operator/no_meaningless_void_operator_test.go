package no_meaningless_void_operator

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoMeaninglessVoidOperatorRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoMeaninglessVoidOperatorRule, []rule_tester.ValidTestCase{
		{Code: `
(() => {})();

function foo() {}
foo(); // nothing to discard

function bar(x: number) {
  void x;
  return 2;
}
void bar(); // discarding a number
    `},
		{Code: `
function bar(x: never) {
  void x;
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "void (() => {})();",
			Output: []string{" (() => {})();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo() {}
void foo();
      `,
			Output: []string{`
function foo() {}
 foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
function bar(x: never) {
  void x;
}
      `,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeVoid",
							Output: `
function bar(x: never) {
   x;
}
      `,
						},
					},
				},
			},
		},
	})
}
