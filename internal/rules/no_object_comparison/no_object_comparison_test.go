package no_object_comparison

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoObjectComparisonRule(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoObjectComparisonRule, []rule_tester.ValidTestCase{
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}

				  plus(amount: number): ValueObject {
				    return new ValueObject(this.value + amount);
				  }

				  lte(other: ValueObject): boolean {
				    return this.value <= other.value;
				  }
				}

				const a = new ValueObject(10);
				const b = new ValueObject(11);

				if (a.lte(b)) {
				  b.plus(1);
				}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject"}]}`),
		},
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}

				  plus(amount: number): ValueObject {
				    return new ValueObject(this.value + amount);
				  }
				}

				const a = new ValueObject(10);
				if (a !== null) {
				  a.plus(1);
				}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject","forbidEqualityOperators":true}]}`),
		},
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}

				  plus(amount: number): ValueObject {
				    return new ValueObject(this.value + amount);
				  }
				}

				const a = new ValueObject(10);
				if (a !== undefined) {
				  a.plus(1);
				}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject","forbidEqualityOperators":true}]}`),
		},
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}
				}

				const a = new ValueObject(10);
				const b = a;

				if (a === b) {}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject"}]}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}
				}

				const a = new ValueObject(10);
				const b = new ValueObject(11);

				if (a <= b) {}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject"}]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}
				}

				class OtherValueObject {
				  constructor(private value: number) {}
				}

				if (new OtherValueObject(10) === new ValueObject(11)) {}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject"},{"name":"OtherValueObject","forbidEqualityOperators":true}]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
		{
			Code: `
				class ValueObject {
				  constructor(private value: number) {}
				}

				class OtherValueObject {
				  constructor(private value: number) {}
				}

				declare const a: ValueObject | number;
				declare const b: OtherValueObject | string;

				if (a === b) {}
			`,
			Options: rule_tester.OptionsFromJSON[NoObjectComparisonOptions](`{"classes":[{"name":"ValueObject","forbidEqualityOperators":true},{"name":"OtherValueObject","forbidEqualityOperators":true}]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "objectComparison"},
			},
		},
	})
}
