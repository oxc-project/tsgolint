package no_unnecessary_boolean_literal_compare

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryBooleanLiteralCompareRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryBooleanLiteralCompareRule, []rule_tester.ValidTestCase{
		{Code: `
      declare const varAny: any;
      varAny === true;
    `},
		{Code: `
      declare const varAny: any;
      varAny == false;
    `},
		{Code: `
      declare const varString: string;
      varString === false;
    `},
		{Code: `
      declare const varString: string;
      varString === true;
    `},
		{Code: `
      declare const varObject: {};
      varObject === true;
    `},
		{Code: `
      declare const varObject: {};
      varObject == false;
    `},
		{Code: `
      declare const varNullOrUndefined: null | undefined;
      varNullOrUndefined === false;
    `},
		{Code: `
      declare const varBooleanOrString: boolean | string;
      varBooleanOrString === false;
    `},
		{Code: `
      declare const varBooleanOrString: boolean | string;
      varBooleanOrString == true;
    `},
		{Code: `
      declare const varTrueOrStringOrUndefined: true | string | undefined;
      varTrueOrStringOrUndefined == true;
    `},
		{Code: `
      const test: <T>(someCondition: T) => void = someCondition => {
        if (someCondition === true) {
        }
      };
    `},
		{Code: `
      const test: <T>(someCondition: boolean | string) => void = someCondition => {
        if (someCondition === true) {
        }
      };
    `},
		{Code: `
      declare const varBooleanOrUndefined: boolean | undefined;
      varBooleanOrUndefined === true;
    `},
		{
			Code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        varBooleanOrUndefined === true;
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToFalse": false}`),
		},
		{
			Code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        varBooleanOrUndefined === false;
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToTrue": false}`),
		},
		{
			Code: `
        const test: <T extends boolean | undefined>(
          someCondition: T,
        ) => void = someCondition => {
          if (someCondition === true) {
          }
        };
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToFalse": false}`),
		},
		{
			Code: `
        const test: <T extends boolean | undefined>(
          someCondition: T,
        ) => void = someCondition => {
          if (someCondition === false) {
          }
        };
      `,
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToTrue": false}`),
		},
		{Code: "'false' === true;"},
		{Code: "'true' === false;"},
		{Code: `
const unconstrained: <T>(someCondition: T) => void = someCondition => {
  if (someCondition === true) {
  }
};
    `},
		{Code: `
const extendsUnknown: <T extends unknown>(
  someCondition: T,
) => void = someCondition => {
  if (someCondition === true) {
  }
};
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "true === true;",
			Output: []string{"true;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code:   "false !== true;",
			Output: []string{"!false;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (varBoolean !== false) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varTrue: true;
        if (varTrue !== true) {
        }
      `,
			Output: []string{`
        declare const varTrue: true;
        if (!varTrue) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varTrueOrUndefined: true | undefined;
        if (varTrueOrUndefined === true) {
        }
      `,
			Output: []string{`
        declare const varTrueOrUndefined: true | undefined;
        if (varTrueOrUndefined) {
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToTrue": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "comparingNullableToTrueDirect",
				},
			},
		},
		{
			Code: `
        declare const varFalseOrNull: false | null;
        if (varFalseOrNull !== true) {
        }
      `,
			Output: []string{`
        declare const varFalseOrNull: false | null;
        if (!varFalseOrNull) {
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToTrue": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "comparingNullableToTrueNegated",
				},
			},
		},
		{
			Code: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (varBooleanOrNull === false && otherBoolean) {
        }
      `,
			Output: []string{`
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (!(varBooleanOrNull ?? true) && otherBoolean) {
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToFalse": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "comparingNullableToFalse",
				},
			},
		},
		{
			Code: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (!(varBooleanOrNull === false) || otherBoolean) {
        }
      `,
			Output: []string{`
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if ((varBooleanOrNull ?? true) || otherBoolean) {
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToFalse": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "comparingNullableToFalse",
				},
			},
		},
		{
			Code: `
        declare const varTrueOrFalseOrUndefined: true | false | undefined;
        declare const otherBoolean: boolean;
        if (varTrueOrFalseOrUndefined !== false && !otherBoolean) {
        }
      `,
			Output: []string{`
        declare const varTrueOrFalseOrUndefined: true | false | undefined;
        declare const otherBoolean: boolean;
        if ((varTrueOrFalseOrUndefined ?? true) && !otherBoolean) {
        }
      `,
			},
			Options: rule_tester.OptionsFromJSON[NoUnnecessaryBooleanLiteralCompareOptions](`{"allowComparingNullableBooleansToFalse": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "comparingNullableToFalse",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (false !== varBoolean) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if ( varBoolean) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (true !== varBoolean) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (! varBoolean) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const x;
        if ((x instanceof Error) === false) {
        }
      `,
			Output: []string{`
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const x;
        if (false === (x instanceof Error)) {
        }
      `,
			Output: []string{`
        declare const x;
        if (! (x instanceof Error)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const x;
        if (x instanceof Error === false) {
        }
      `,
			Output: []string{`
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const x;
        if (typeof x === 'string' === false) {
        }
      `,
			Output: []string{`
        declare const x;
        if (!(typeof x === 'string')) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const x;
        if (x instanceof Error === (false)) {
        }
      `,
			Output: []string{`
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const x;
        if ((false) === x instanceof Error) {
        }
      `,
			Output: []string{`
        declare const x;
        if (!( x instanceof Error)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!(varBoolean !== false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!(varBoolean === false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!(varBoolean instanceof Event == false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (varBoolean instanceof Event) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (varBoolean instanceof Event == false) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (!(varBoolean instanceof Event)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? false) !== false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (!(varBoolean ?? false)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? false) === false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if ((varBoolean ?? false)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? true) !== false)) {
        }
      `,
			Output: []string{`
        declare const varBoolean: boolean;
        if (!(varBoolean ?? true)) {
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (someCondition === true) {
          }
        };
      `,
			Output: []string{`
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (someCondition) {
          }
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "direct",
				},
			},
		},
		{
			Code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!(someCondition !== false)) {
          }
        };
      `,
			Output: []string{`
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!someCondition) {
          }
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!((someCondition ?? true) !== false)) {
          }
        };
      `,
			Output: []string{`
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!(someCondition ?? true)) {
          }
        };
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "negated",
				},
			},
		},
		{
			Code: `
function foo(): boolean {}
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
				},
			},
		},
	})
}
