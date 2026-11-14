package prefer_includes

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferIncludesPreservesComparisonSemantics(t *testing.T) {
	t.Parallel()
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct{ code, suggestion string }{
		{
			"const xs: number[] = [NaN];\nxs.indexOf(NaN) !== -1;",
			"const xs: number[] = [NaN];\nxs.includes(NaN);",
		},
		{
			"const xs: number[] = [NaN];\nxs.indexOf(NaN) === -1;",
			"const xs: number[] = [NaN];\n!xs.includes(NaN);",
		},
		{
			"const xs: (string | undefined)[] = [, 'a'];\nxs.indexOf(undefined) !== -1;",
			"const xs: (string | undefined)[] = [, 'a'];\nxs.includes(undefined);",
		},
		{
			"const xs = new Float64Array([NaN]);\nxs.indexOf(NaN) !== -1;",
			"const xs = new Float64Array([NaN]);\nxs.includes(NaN);",
		},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: test.code,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "preferIncludes",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "useIncludes",
					Output:    test.suggestion,
				}},
			}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferIncludesRule, nil, invalid)
}
