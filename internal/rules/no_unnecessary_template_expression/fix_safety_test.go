package no_unnecessary_template_expression

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryTemplateExpressionPreservesWidening(t *testing.T) {
	t.Parallel()
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct{ code, suggestion string }{
		{
			"declare const mode: 'light' | 'dark';\nlet label = `${mode}`;\nlabel = 'auto';",
			"declare const mode: 'light' | 'dark';\nlet label = mode;\nlabel = 'auto';",
		},
		{
			"type UserId = string & { readonly __brand: 'UserId' };\ndeclare const id: UserId;\nlet label = `${id}`;\nlabel = 'guest';",
			"type UserId = string & { readonly __brand: 'UserId' };\ndeclare const id: UserId;\nlet label = id;\nlabel = 'guest';",
		},
		{
			"function f<T extends string>(value: T) { let label = `${value}`; label = 'other'; return label; }",
			"function f<T extends string>(value: T) { let label = value; label = 'other'; return label; }",
		},
		{
			"declare const value: `prefix-${string}`;\nlet label = `${value}`;\nlabel = 'other';",
			"declare const value: `prefix-${string}`;\nlet label = value;\nlabel = 'other';",
		},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: test.code,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noUnnecessaryTemplateExpression",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeTemplateExpression",
					Output:    test.suggestion,
				}},
			}},
		})
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   "declare const value: string;\nlet label = `${value}`;\nlabel = 'other';",
		Output: []string{"declare const value: string;\nlet label = value;\nlabel = 'other';"},
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUnnecessaryTemplateExpression"}},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTemplateExpressionRule, nil, invalid)
}
