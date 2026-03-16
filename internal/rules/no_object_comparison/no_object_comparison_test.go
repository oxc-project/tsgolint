package no_object_comparison

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoObjectComparisonRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoObjectComparisonRule, []rule_tester.ValidTestCase{
		// Allows for comparison using class's comparison function
		{
			Code: `
        declare class NormalizedNumber {
          lte(other: NormalizedNumber): boolean;
          plus(value: number): NormalizedNumber;
        }

        declare const a: NormalizedNumber;
        declare const b: NormalizedNumber;

        if (a.lte(b)) {
          b.plus(1);
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber"]}`),
		},
		// Allows for null comparison
		{
			Code: `
        declare class NormalizedNumber {}
        declare const a: NormalizedNumber | null;

        if (a !== null) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber"]}`),
		},
		// Allows for undefined comparison
		{
			Code: `
        declare class NormalizedNumber {}
        declare const a: NormalizedNumber | undefined;

        if (a !== undefined) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber"]}`),
		},
		// Added for the TSGolint migration: comparison of non-configured object types remains allowed.
		{
			Code: `
        declare class SafeNumber {}
        declare const a: SafeNumber;
        declare const b: SafeNumber;

        if (a === b) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber"]}`),
		},
	}, []rule_tester.InvalidTestCase{
		// Disallows comparison for the same banned class
		{
			Code: `
        declare class NormalizedNumber {}
        declare const a: NormalizedNumber;
        declare const b: NormalizedNumber;

        if (a <= b) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
		// Disallows comparison for different banned classes
		{
			Code: `
        declare class NormalizedNumber {}
        declare class BaseNumber {}

        if (({} as BaseNumber) === ({} as NormalizedNumber)) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber","BaseNumber"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
		// Disallows comparison for union types
		{
			Code: `
        declare class NormalizedNumber {}
        declare class BaseNumber {}

        declare const a: NormalizedNumber | number;
        declare const b: BaseNumber | string;

        if (a === b) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["NormalizedNumber","BaseNumber"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
		// Added for the TSGolint migration: interface names are matched alongside classes.
		{
			Code: `
        interface ComparableValue {
          value: number;
        }

        declare const a: ComparableValue;
        declare const b: ComparableValue;

        if (a === b) {
        }
      `,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classNames":["ComparableValue"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
	})
}
