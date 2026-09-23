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
		// Assignment regressions from typescript-eslint/typescript-eslint#12873.
		{Code: `declare let x: number; void (x = 1);`},
		{Code: `declare let x: number; () => void (x = 1);`},
		{Code: `declare let x: number; void (x += 1);`},
		{Code: `declare let x: number; declare let y: number; void (x = y = 1);`},
		{Code: `declare let x: string; declare function getValue(): string; void (x = getValue());`},
		{Code: `declare const obj: { prop: number }; void (obj.prop = 1);`},
		{Code: `declare let x: number; declare let y: number; void ((x = 1), (y = 2));`},
		{Code: `declare let x: undefined; () => void (((x = undefined)));`},
		{Code: `declare let x: void; declare function fn(): void; void (x = fn());`},
		{Code: `declare let x: undefined; void (x ??= undefined);`},
		{Code: `declare let x: undefined; void ((x = undefined) as undefined);`},
		{Code: `declare let x: undefined; void (<undefined>(x = undefined));`},
		{Code: `declare let x: undefined; void ((x = undefined) satisfies undefined);`},
		{Code: `declare let x: undefined; void ((x = undefined), (x = undefined));`},
		{Code: `declare let x: undefined; void ((x = undefined)!, ((x = undefined) as undefined));`},
		{
			Code:    `declare let x: undefined; void ((x = undefined)!);`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
		},
		{
			Code:    `declare let x: never; declare function fail(): never; void (x = fail());`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
		},
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
		{
			Code: `
const foo = (() => {}) as (() => void) | undefined;
void foo?.();
      `,
			Output: []string{`
const foo = (() => {}) as (() => void) | undefined;
 foo?.();
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
			Code:   `declare let x: undefined; declare function fn(): void; void ((x = undefined), fn());`,
			Output: []string{`declare let x: undefined; declare function fn(): void;  ((x = undefined), fn());`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
	})
}
