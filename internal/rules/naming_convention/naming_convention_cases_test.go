package naming_convention

import (
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// This file is a port of typescript-eslint's `tests/rules/naming-convention/cases`.
//
// The upstream generator produces every combination of format × name × affix
// variant for every selector (well over ten thousand cases). Since every case
// needs its own program (and every invalid case a snapshot) here, the matrix is
// reduced: the full format × name matrix and all affix/underscore variants run
// for the `variable` selector, while the other selectors run one valid and one
// invalid name per format plus the valid affix/underscore variants (which are
// format independent).

type formatTestName struct {
	format  string
	valid   []string
	invalid []string
}

var formatTestNames = []formatTestName{
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

// filter to not match `[iI]gnored`
var ignoredFilter = map[string]any{"match": false, "regex": ".gnored"}

type generatedCaseGroup struct {
	code []string
	// the selector (and optional modifiers) the group tests
	selector  any
	modifiers []NamingConventionModifier
	// whether to run the full format × name matrix
	full bool
}

type caseVariant struct {
	// how the name is decorated, `%` is replaced with the name
	name string
	// mutates the selector options for the variant
	apply func(s *NamingConventionSelector)
	// invalid variants only
	messageId string
}

func withLeadingUnderscore(option NamingConventionUnderscoreOption) func(s *NamingConventionSelector) {
	return func(s *NamingConventionSelector) { s.LeadingUnderscore = &option }
}

func withTrailingUnderscore(option NamingConventionUnderscoreOption) func(s *NamingConventionSelector) {
	return func(s *NamingConventionSelector) { s.TrailingUnderscore = &option }
}

func withPrefix(prefix ...string) func(s *NamingConventionSelector) {
	return func(s *NamingConventionSelector) { s.Prefix = prefix }
}

func withSuffix(suffix ...string) func(s *NamingConventionSelector) {
	return func(s *NamingConventionSelector) { s.Suffix = suffix }
}

var validVariants = []caseVariant{
	// leadingUnderscore
	{name: "%", apply: withLeadingUnderscore("forbid")},
	{name: "_%", apply: withLeadingUnderscore("require")},
	{name: "__%", apply: withLeadingUnderscore("requireDouble")},
	{name: "_%", apply: withLeadingUnderscore("allow")},
	{name: "%", apply: withLeadingUnderscore("allow")},
	{name: "__%", apply: withLeadingUnderscore("allowDouble")},
	{name: "%", apply: withLeadingUnderscore("allowDouble")},
	{name: "_%", apply: withLeadingUnderscore("allowSingleOrDouble")},
	{name: "%", apply: withLeadingUnderscore("allowSingleOrDouble")},
	{name: "__%", apply: withLeadingUnderscore("allowSingleOrDouble")},

	// trailingUnderscore
	{name: "%", apply: withTrailingUnderscore("forbid")},
	{name: "%_", apply: withTrailingUnderscore("require")},
	{name: "%__", apply: withTrailingUnderscore("requireDouble")},
	{name: "%_", apply: withTrailingUnderscore("allow")},
	{name: "%", apply: withTrailingUnderscore("allow")},
	{name: "%__", apply: withTrailingUnderscore("allowDouble")},
	{name: "%", apply: withTrailingUnderscore("allowDouble")},
	{name: "%_", apply: withTrailingUnderscore("allowSingleOrDouble")},
	{name: "%", apply: withTrailingUnderscore("allowSingleOrDouble")},
	{name: "%__", apply: withTrailingUnderscore("allowSingleOrDouble")},

	// prefix
	{name: "MyPrefix%", apply: withPrefix("MyPrefix")},
	{name: "MyPrefix2%", apply: withPrefix("MyPrefix1", "MyPrefix2")},

	// suffix
	{name: "%MySuffix", apply: withSuffix("MySuffix")},
	{name: "%MySuffix2", apply: withSuffix("MySuffix1", "MySuffix2")},
}

var invalidVariants = []caseVariant{
	// leadingUnderscore
	{name: "_%", apply: withLeadingUnderscore("forbid"), messageId: "unexpectedUnderscore"},
	{name: "%", apply: withLeadingUnderscore("require"), messageId: "missingUnderscore"},
	{name: "%", apply: withLeadingUnderscore("requireDouble"), messageId: "missingUnderscore"},
	{name: "_%", apply: withLeadingUnderscore("requireDouble"), messageId: "missingUnderscore"},

	// trailingUnderscore
	{name: "%_", apply: withTrailingUnderscore("forbid"), messageId: "unexpectedUnderscore"},
	{name: "%", apply: withTrailingUnderscore("require"), messageId: "missingUnderscore"},
	{name: "%", apply: withTrailingUnderscore("requireDouble"), messageId: "missingUnderscore"},
	{name: "%_", apply: withTrailingUnderscore("requireDouble"), messageId: "missingUnderscore"},

	// prefix
	{name: "%", apply: withPrefix("MyPrefix"), messageId: "missingAffix"},
	{name: "%", apply: withPrefix("MyPrefix1", "MyPrefix2"), messageId: "missingAffix"},

	// suffix
	{name: "%", apply: withSuffix("MySuffix"), messageId: "missingAffix"},
	{name: "%", apply: withSuffix("MySuffix1", "MySuffix2"), messageId: "missingAffix"},
}

func (g generatedCaseGroup) makeCase(name string, format string, apply func(s *NamingConventionSelector)) (string, NamingConventionOptions) {
	selector := NamingConventionSelector{
		Selector:  g.selector,
		Modifiers: g.modifiers,
		Format:    []any{format},
	}
	if apply != nil {
		apply(&selector)
	}

	// the comment documents the options of the case (like upstream), the filter
	// is added afterwards so it doesn't clutter every case
	description, err := json.Marshal(selector)
	if err != nil {
		panic(err)
	}
	selector.Filter = ignoredFilter

	lines := make([]string, 0, len(g.code)+1)
	lines = append(lines, "// "+string(description))
	for _, line := range g.code {
		lines = append(lines, strings.ReplaceAll(line, "%", name))
	}
	return strings.Join(lines, "\n"), NamingConventionOptions{selector}
}

func (g generatedCaseGroup) selectorCount() int {
	if selectors, ok := g.selector.([]any); ok {
		return len(selectors)
	}
	return 1
}

func (g generatedCaseGroup) errors(messageId string) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(g.code)*g.selectorCount())
	for range g.code {
		for range g.selectorCount() {
			errors = append(errors, rule_tester.InvalidTestCaseError{MessageId: messageId})
		}
	}
	return errors
}

func createTestCases(t *testing.T, groups []generatedCaseGroup) {
	t.Helper()

	var validCases []rule_tester.ValidTestCase
	var invalidCases []rule_tester.InvalidTestCase

	for _, group := range groups {
		for i, names := range formatTestNames {
			validNames := names.valid
			invalidNames := names.invalid
			if !group.full {
				validNames = validNames[:1]
				invalidNames = invalidNames[:1]
			}

			for _, name := range validNames {
				code, options := group.makeCase(name, names.format, nil)
				validCases = append(validCases, rule_tester.ValidTestCase{Code: code, Options: options})
			}

			for _, name := range invalidNames {
				code, options := group.makeCase(name, names.format, nil)
				invalidCases = append(invalidCases, rule_tester.InvalidTestCase{Code: code, Options: options, Errors: group.errors("doesNotMatchFormat")})
			}

			// the underscore/affix variants are format independent, so they only
			// run for the first format
			if i != 0 {
				continue
			}

			for _, variant := range validVariants {
				code, options := group.makeCase(strings.ReplaceAll(variant.name, "%", names.valid[0]), names.format, variant.apply)
				validCases = append(validCases, rule_tester.ValidTestCase{Code: code, Options: options})
			}

			// the invalid variants produce large snapshots, so they only run for
			// the groups with the full matrix
			if !group.full {
				continue
			}
			for _, variant := range invalidVariants {
				code, options := group.makeCase(strings.ReplaceAll(variant.name, "%", names.invalid[0]), names.format, variant.apply)
				invalidCases = append(invalidCases, rule_tester.InvalidTestCase{Code: code, Options: options, Errors: group.errors(variant.messageId)})
			}
		}
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule, validCases, invalidCases)
}

func TestNamingConventionRuleGeneratedCases(t *testing.T) {
	t.Parallel()
	createTestCases(t, []generatedCaseGroup{
		// #region accessor
		{
			code: []string{
				"class Ignored { accessor % = 10; }",
				"class Ignored { accessor #% = 10; }",
				"class Ignored { static accessor % = 10; }",
				"class Ignored { static accessor #% = 10; }",
				"class Ignored { private accessor % = 10; }",
				"class Ignored { private static accessor % = 10; }",
				"class Ignored { override accessor % = 10; }",
				"class Ignored { accessor \"%\" = 10; }",
				"class Ignored { protected accessor % = 10; }",
				"class Ignored { public accessor % = 10; }",
				"class Ignored { abstract accessor %; }",
				"const ignored = { get %() {} };",
				"const ignored = { set \"%\"(ignored) {} };",
				"class Ignored { private get %() {} }",
				"class Ignored { private set \"%\"(ignored) {} }",
				"class Ignored { private static get %() {} }",
				"class Ignored { static get #%() {} }",
			},
			selector: "accessor",
		},
		// #endregion accessor

		// #region autoAccessor
		{
			code: []string{
				"class Ignored { accessor % = 10; }",
				"class Ignored { accessor #% = 10; }",
				"class Ignored { static accessor % = 10; }",
				"class Ignored { static accessor #% = 10; }",
				"class Ignored { private accessor % = 10; }",
				"class Ignored { private static accessor % = 10; }",
				"class Ignored { override accessor % = 10; }",
				"class Ignored { accessor \"%\" = 10; }",
				"class Ignored { protected accessor % = 10; }",
				"class Ignored { public accessor % = 10; }",
				"class Ignored { abstract accessor %; }",
			},
			selector: "autoAccessor",
		},
		// #endregion autoAccessor

		// #region class
		{
			code:     []string{"class % {}", "abstract class % {}", "const ignored = class % {}"},
			selector: "class",
		},
		// #endregion class

		// #region classicAccessor
		{
			code: []string{
				"const ignored = { get %() {} };",
				"const ignored = { set \"%\"(ignored) {} };",
				"class Ignored { private get %() {} }",
				"class Ignored { private set \"%\"(ignored) {} }",
				"class Ignored { private static get %() {} }",
				"class Ignored { static get #%() {} }",
				"abstract class Ignored { abstract get %(): number }",
				"abstract class Ignored { abstract set %(ignored: number) }",
			},
			selector: "classicAccessor",
		},
		// #endregion classicAccessor

		// #region default
		{
			code: []string{
				"const % = 1;",
				"function % () {}",
				"(function (%) {});",
				"class Ignored { constructor(private %) {} }",
				"const ignored = { % };",
				"interface Ignored { %: string }",
				"type Ignored = { %: string }",
				"class Ignored { private % = 1 }",
				"class Ignored { #% = 1 }",
				"class Ignored { constructor(private %) {} }",
				"class Ignored { #%() {} }",
				"class Ignored { private %() {} }",
				"const ignored = { %() {} };",
				"class Ignored { private get %() {} }",
				"enum Ignored { % }",
				"abstract class % {}",
				"interface % { }",
				"type % = { };",
				"enum % {}",
				"interface Ignored<%> extends Ignored<string> {}",
			},
			selector: "default",
		},
		// #endregion default

		// #region enum
		{
			code:     []string{"enum % {}"},
			selector: "enum",
		},
		// #endregion enum

		// #region enumMember
		{
			code:     []string{"enum Ignored { % }", "enum Ignored { \"%\" }"},
			selector: "enumMember",
		},
		// #endregion enumMember

		// #region function
		{
			code:     []string{"function % () {}", "(function % () {});", "declare function % ();"},
			selector: "function",
		},
		// #endregion function

		// #region interface
		{
			code:     []string{"interface % {}"},
			selector: "interface",
		},
		// #endregion interface

		// #region method
		{
			code: []string{
				"class Ignored { private %() {} }",
				"class Ignored { private \"%\"() {} }",
				"class Ignored { private async %() {} }",
				"class Ignored { private static %() {} }",
				"class Ignored { private static async %() {} }",
				"class Ignored { private % = () => {} }",
				"class Ignored { abstract %() }",
				"class Ignored { #%() }",
				"class Ignored { static #%() }",
			},
			selector: "classMethod",
		},
		{
			code: []string{
				"const ignored = { %() {} };",
				"const ignored = { \"%\"() {} };",
				"const ignored = { %: () => {} };",
			},
			selector: "objectLiteralMethod",
		},
		{
			code: []string{
				"interface Ignored { %(): string }",
				"interface Ignored { \"%\"(): string }",
				"interface Ignored { %: () => string }",
				"interface Ignored { \"%\": () => string }",
				"type Ignored = { %(): string }",
				"type Ignored = { \"%\"(): string }",
				"type Ignored = { %: () => string }",
				"type Ignored = { \"%\": () => string }",
			},
			selector: "typeMethod",
		},
		// #endregion method

		// #region parameter
		{
			code: []string{
				"function ignored(%) {}",
				"(function (%) {});",
				"declare function ignored(%);",
				"function ignored({%}) {}",
				"function ignored(...%) {}",
				"function ignored({% = 1}) {}",
				"function ignored({...%}) {}",
				"function ignored([%]) {}",
				"function ignored([% = 1]) {}",
				"function ignored([...%]) {}",
			},
			selector: "parameter",
		},
		// #endregion parameter

		// #region parameterProperty
		{
			code: []string{
				"class Ignored { constructor(private %) {} }",
				"class Ignored { constructor(readonly %) {} }",
				"class Ignored { constructor(private readonly %) {} }",
			},
			selector: "parameterProperty",
		},
		{
			code:      []string{"class Ignored { constructor(private readonly %) {} }"},
			selector:  "parameterProperty",
			modifiers: []NamingConventionModifier{"readonly"},
		},
		// #endregion parameterProperty

		// #region property
		{
			code: []string{
				"class Ignored { private % }",
				"class Ignored { private \"%\" = 1 }",
				"class Ignored { private readonly % = 1 }",
				"class Ignored { private static % }",
				"class Ignored { private static readonly % = 1 }",
				"class Ignored { abstract % }",
				"class Ignored { declare % }",
				"class Ignored { #% }",
				"class Ignored { static #% }",
			},
			selector: "classProperty",
		},
		{
			code:     []string{"const ignored = { % };", "const ignored = { \"%\": 1 };"},
			selector: "objectLiteralProperty",
		},
		{
			code: []string{
				"interface Ignored { % }",
				"interface Ignored { \"%\": string }",
				"type Ignored = { % }",
				"type Ignored = { \"%\": string }",
			},
			selector: "typeProperty",
		},
		// #endregion property

		// #region typeAlias
		{
			code:     []string{"type % = {};", "type % = 1;"},
			selector: "typeAlias",
		},
		// #endregion typeAlias

		// #region typeParameter
		{
			code: []string{
				"class Ignored<%> {}",
				"function ignored<%>() {}",
				"type Ignored<%> = { ignored: % };",
				"interface Ignored<%> extends Ignored<string> {}",
			},
			selector: "typeParameter",
		},
		// #endregion typeParameter

		// #region variable
		{
			code: []string{
				"const % = 1;",
				"let % = 1;",
				"var % = 1;",
				"const {%} = {ignored: 1};",
				"const {% = 2} = {ignored: 1};",
				"const {...%} = {ignored: 1};",
				"const [%] = [1];",
				"const [% = 1] = [1];",
				"const [...%] = [1];",
			},
			selector: "variable",
			full:     true,
		},
		// #endregion variable
	})
}
