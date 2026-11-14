package no_unnecessary_type_assertion

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeAssertionDeclarationNarrowing(t *testing.T) {
	t.Parallel()
	var valid []rule_tester.ValidTestCase
	for _, declaration := range []string{"const", "let"} {
		for _, initializer := range []string{
			"maybe!",
			"(maybe!)",
			"maybe as string",
			"(maybe as string)",
			"<string>maybe",
			"flag ? maybe! : 'fallback'",
			"flag ? maybe as string : 'fallback'",
			"text && maybe!",
		} {
			valid = append(valid, rule_tester.ValidTestCase{
				Code: "declare const maybe: string | undefined;\ndeclare const flag: boolean;\ndeclare const text: string;\n" + declaration +
					" label: string | undefined = " + initializer + ";\nconsole.log(label.length);\nexport {};",
			})
		}
	}
	valid = append(valid, rule_tester.ValidTestCase{
		Code: "type Value = string | null;\ndeclare const maybe: Value;\nconst label: Value = maybe!;\nconsole.log(label.length);\nexport {};",
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryTypeAssertionRule, valid, []rule_tester.InvalidTestCase{
		{
			Code:   "declare const value: string;\nconst label: string | undefined = value!;",
			Output: []string{"declare const value: string;\nconst label: string | undefined = value;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
		{
			Code:   "declare const value: string;\nconst label: string | undefined = value as string;",
			Output: []string{"declare const value: string;\nconst label: string | undefined = value;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion"}},
		},
	})
}
