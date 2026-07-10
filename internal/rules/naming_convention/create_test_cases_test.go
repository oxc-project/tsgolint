package naming_convention

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Port of the upstream dynamic test case generator.
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/naming-convention/cases/createTestCases.ts
//
// Where upstream asserts on error `data` (name/type/formats/...), this port
// asserts the exact rendered message plus the reported line/column range
// (computed from the template, see placeholderRange). Snapshots are skipped
// for the generated invalid cases: they would add several megabytes of
// generated snapshot entries while asserting nothing the message and
// position assertions don't.

type formatNames struct {
	format  string
	valid   []string
	invalid []string
}

// formatTestNames mirrors upstream's formatTestNames, preserving entry order.
var formatTestNames = []formatNames{
	{
		format:  "camelCase",
		valid:   []string{"strictCamelCase", "lower", "camelCaseUNSTRICT"},
		invalid: []string{"snake_case", "UPPER_CASE", "UPPER", "StrictPascalCase"},
	},
	{
		format:  "PascalCase",
		valid:   []string{"StrictPascalCase", "Pascal", "I18n", "PascalCaseUNSTRICT", "UPPER"},
		invalid: []string{"snake_case", "UPPER_CASE", "strictCamelCase"},
	},
	{
		format:  "snake_case",
		valid:   []string{"snake_case", "lower"},
		invalid: []string{"UPPER_CASE", "SNAKE_case_UNSTRICT", "strictCamelCase", "StrictPascalCase"},
	},
	{
		format:  "strictCamelCase",
		valid:   []string{"strictCamelCase", "lower"},
		invalid: []string{"snake_case", "UPPER_CASE", "UPPER", "StrictPascalCase", "camelCaseUNSTRICT"},
	},
	{
		format:  "StrictPascalCase",
		valid:   []string{"StrictPascalCase", "Pascal", "I18n"},
		invalid: []string{"snake_case", "UPPER_CASE", "UPPER", "strictCamelCase", "PascalCaseUNSTRICT"},
	},
	{
		format:  "UPPER_CASE",
		valid:   []string{"UPPER_CASE", "UPPER"},
		invalid: []string{"lower", "snake_case", "SNAKE_case_UNSTRICT", "strictCamelCase", "StrictPascalCase"},
	},
}

// ignoredFilter mirrors upstream's IGNORED_FILTER: skip names matching `[iI]gnored`.
var ignoredFilter = MatchRegex{Match: false, Regex: ".gnored"}

// testCaseSpec mirrors upstream's `Cases` entries: a set of code templates
// (with `%` as the name placeholder) plus selector options without `format`.
type testCaseSpec struct {
	code    []string
	options NamingConventionOption
}

func selectorsOf(selector any) []string {
	switch s := selector.(type) {
	case string:
		return []string{s}
	case []string:
		return s
	case []any:
		out := make([]string, len(s))
		for i, v := range s {
			out[i] = v.(string)
		}
		return out
	}
	return nil
}

// buildCode renders the `// ${JSON.stringify(options)}` comment header
// followed by the code templates with `%` replaced by preparedName.
func buildCode(spec testCaseSpec, preparedName string, options NamingConventionOption) string {
	comment, err := json.Marshal(options)
	if err != nil {
		panic(err)
	}
	lines := make([]string, 0, len(spec.code)+1)
	lines = append(lines, "// "+string(comment))
	for _, code := range spec.code {
		lines = append(lines, strings.ReplaceAll(code, "%", preparedName))
	}
	return strings.Join(lines, "\n")
}

func createValidTestCases(cases []testCaseSpec) []rule_tester.ValidTestCase {
	newCases := make([]rule_tester.ValidTestCase, 0)

	for _, test := range cases {
		for _, names := range formatTestNames {
			format := []string{names.format}
			for _, name := range names.valid {
				createCase := func(preparedName string, mutate func(*NamingConventionOption)) rule_tester.ValidTestCase {
					options := test.options
					options.Format = &format
					mutate(&options)
					code := buildCode(test, preparedName, options)
					options.Filter = ignoredFilter
					return rule_tester.ValidTestCase{
						Code:    code,
						Options: []NamingConventionOption{options},
					}
				}
				noop := func(o *NamingConventionOption) {}
				leading := func(v string) func(*NamingConventionOption) {
					return func(o *NamingConventionOption) { o.LeadingUnderscore = &v }
				}
				trailing := func(v string) func(*NamingConventionOption) {
					return func(o *NamingConventionOption) { o.TrailingUnderscore = &v }
				}

				newCases = append(newCases,
					createCase(name, noop),

					// leadingUnderscore
					createCase(name, leading("forbid")),
					createCase("_"+name, leading("require")),
					createCase("__"+name, leading("requireDouble")),
					createCase("_"+name, leading("allow")),
					createCase(name, leading("allow")),
					createCase("__"+name, leading("allowDouble")),
					createCase(name, leading("allowDouble")),
					createCase("_"+name, leading("allowSingleOrDouble")),
					createCase(name, leading("allowSingleOrDouble")),
					createCase("__"+name, leading("allowSingleOrDouble")),

					// trailingUnderscore
					createCase(name, trailing("forbid")),
					createCase(name+"_", trailing("require")),
					createCase(name+"__", trailing("requireDouble")),
					createCase(name+"_", trailing("allow")),
					createCase(name, trailing("allow")),
					createCase(name+"__", trailing("allowDouble")),
					createCase(name, trailing("allowDouble")),
					createCase(name+"_", trailing("allowSingleOrDouble")),
					createCase(name, trailing("allowSingleOrDouble")),
					createCase(name+"__", trailing("allowSingleOrDouble")),

					// prefix
					createCase("MyPrefix"+name, func(o *NamingConventionOption) { o.Prefix = []string{"MyPrefix"} }),
					createCase("MyPrefix2"+name, func(o *NamingConventionOption) { o.Prefix = []string{"MyPrefix1", "MyPrefix2"} }),

					// suffix
					createCase(name+"MySuffix", func(o *NamingConventionOption) { o.Suffix = []string{"MySuffix"} }),
					createCase(name+"MySuffix2", func(o *NamingConventionOption) { o.Suffix = []string{"MySuffix1", "MySuffix2"} }),
				)
			}
		}
	}

	return newCases
}

// placeholderRange computes the 1-based UTF-16 column range the diagnostic
// must cover in a single-line template: the substituted name itself, plus the
// surrounding quotes for a string-literal name (`"%"`) or the leading hash
// for a private identifier (`#%`). Templates and test names are all ASCII,
// so byte offsets equal UTF-16 columns. The one template with two
// placeholders (a type parameter and a reference to it) reports on the
// first, so the first `%` is always the reported declaration name.
func placeholderRange(template, preparedName string) (column, endColumn int) {
	idx := strings.Index(template, "%")
	if idx < 0 {
		panic("template has no % placeholder: " + template)
	}
	start := idx
	span := len(preparedName)
	if idx > 0 {
		switch template[idx-1] {
		case '"':
			start--
			span += 2
		case '#':
			start--
			span++
		}
	}
	return start + 1, start + 1 + span
}

// metaSelectors mirrors upstream: for meta selectors the reported node type
// varies per matched node, so no message is asserted (snapshot-free either way).
var metaSelectors = map[string]bool{
	"default":      true,
	"variableLike": true,
	"memberLike":   true,
	"typeLike":     true,
	"property":     true,
	"method":       true,
	"accessor":     true,
}

func createInvalidTestCases(cases []testCaseSpec) []rule_tester.InvalidTestCase {
	newCases := make([]rule_tester.InvalidTestCase, 0)

	for _, test := range cases {
		selectors := selectorsOf(test.options.Selector)
		assertMessage := len(selectors) == 1 && !metaSelectors[selectors[0]]
		typeName := ""
		if assertMessage {
			typeName = selectorTypeToMessageString(selectors[0])
		}

		for _, names := range formatTestNames {
			format := []string{names.format}
			for _, name := range names.invalid {
				createCase := func(preparedName string, messageId string, message string, mutate func(*NamingConventionOption)) rule_tester.InvalidTestCase {
					options := test.options
					options.Format = &format
					mutate(&options)
					code := buildCode(test, preparedName, options)
					options.Filter = ignoredFilter

					if !assertMessage {
						message = ""
					}
					errors := make([]rule_tester.InvalidTestCaseError, 0, len(test.code)*len(selectors))
					for lineIdx, tmpl := range test.code {
						// The options comment is line 1, templates follow.
						line := lineIdx + 2
						column, endColumn := placeholderRange(tmpl, preparedName)
						for range selectors {
							errors = append(errors, rule_tester.InvalidTestCaseError{
								MessageId: messageId,
								Message:   message,
								Line:      line,
								Column:    column,
								EndLine:   line,
								EndColumn: endColumn,
							})
						}
					}

					return rule_tester.InvalidTestCase{
						Code:         code,
						Errors:       errors,
						Options:      []NamingConventionOption{options},
						SkipSnapshot: true,
					}
				}
				noop := func(o *NamingConventionOption) {}
				leading := func(v string) func(*NamingConventionOption) {
					return func(o *NamingConventionOption) { o.LeadingUnderscore = &v }
				}
				trailing := func(v string) func(*NamingConventionOption) {
					return func(o *NamingConventionOption) { o.TrailingUnderscore = &v }
				}

				newCases = append(newCases,
					createCase(name, "doesNotMatchFormat",
						fmt.Sprintf("%s name `%s` must match one of the following formats: %s", typeName, name, names.format),
						noop),

					// leadingUnderscore
					createCase("_"+name, "unexpectedUnderscore",
						fmt.Sprintf("%s name `_%s` must not have a leading underscore.", typeName, name),
						leading("forbid")),
					createCase(name, "missingUnderscore",
						fmt.Sprintf("%s name `%s` must have one leading underscore(s).", typeName, name),
						leading("require")),
					createCase(name, "missingUnderscore",
						fmt.Sprintf("%s name `%s` must have two leading underscore(s).", typeName, name),
						leading("requireDouble")),
					createCase("_"+name, "missingUnderscore",
						fmt.Sprintf("%s name `_%s` must have two leading underscore(s).", typeName, name),
						leading("requireDouble")),

					// trailingUnderscore
					createCase(name+"_", "unexpectedUnderscore",
						fmt.Sprintf("%s name `%s_` must not have a trailing underscore.", typeName, name),
						trailing("forbid")),
					createCase(name, "missingUnderscore",
						fmt.Sprintf("%s name `%s` must have one trailing underscore(s).", typeName, name),
						trailing("require")),
					createCase(name, "missingUnderscore",
						fmt.Sprintf("%s name `%s` must have two trailing underscore(s).", typeName, name),
						trailing("requireDouble")),
					createCase(name+"_", "missingUnderscore",
						fmt.Sprintf("%s name `%s_` must have two trailing underscore(s).", typeName, name),
						trailing("requireDouble")),

					// prefix
					createCase(name, "missingAffix",
						fmt.Sprintf("%s name `%s` must have one of the following prefixes: MyPrefix", typeName, name),
						func(o *NamingConventionOption) { o.Prefix = []string{"MyPrefix"} }),
					createCase(name, "missingAffix",
						fmt.Sprintf("%s name `%s` must have one of the following prefixes: MyPrefix1, MyPrefix2", typeName, name),
						func(o *NamingConventionOption) { o.Prefix = []string{"MyPrefix1", "MyPrefix2"} }),

					// suffix
					createCase(name, "missingAffix",
						fmt.Sprintf("%s name `%s` must have one of the following suffixes: MySuffix", typeName, name),
						func(o *NamingConventionOption) { o.Suffix = []string{"MySuffix"} }),
					createCase(name, "missingAffix",
						fmt.Sprintf("%s name `%s` must have one of the following suffixes: MySuffix1, MySuffix2", typeName, name),
						func(o *NamingConventionOption) { o.Suffix = []string{"MySuffix1", "MySuffix2"} }),
				)
			}
		}
	}

	return newCases
}

// createTestCases mirrors upstream's createTestCases: expand the given specs
// across every format/name/underscore/prefix/suffix combination and run them.
func createTestCases(t *testing.T, cases []testCaseSpec) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.minimal.json",
		t,
		&NamingConventionRule,
		createValidTestCases(cases),
		createInvalidTestCases(cases),
	)
}
