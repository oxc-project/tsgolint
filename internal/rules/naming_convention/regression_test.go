package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Regression tests for divergences from upstream that the ported suite does
// not cover. Each case pins the upstream behavior the port was fixed to match.
func TestNamingConventionUpstreamParity(t *testing.T) {
	t.Parallel()

	validCases := []rule_tester.ValidTestCase{
		// Names fully consumed by underscore trimming are valid (upstream
		// format checkers treat the empty string as matching).
		{Code: "const _ = 1;"},
		// `$` and caseless scripts satisfy camelCase/PascalCase/UPPER_CASE
		// upstream; only the first character's case-fold is checked.
		{Code: "const data$ = 1;"},
		{Code: "const $data = 1;"},
		{Code: "class $Foo {}"},
		{Code: "const 名前 = 1;"},
		{
			Code:    "const NAME$ = 1;",
			Options: []NamingConventionOption{{Selector: "variable", Format: &[]string{"UPPER_CASE"}}},
		},
		// Catch bindings are never visited upstream.
		{Code: "try { void 0; } catch (Some_Error) { throw Some_Error; }"},
		// Parameters in type positions are never visited upstream.
		{
			Code:    "type T = (foo_bar: number) => void;",
			Options: []NamingConventionOption{{Selector: "parameter", Format: &[]string{"camelCase"}}},
		},
		{
			Code:    "interface I { m(foo_bar: number): void }",
			Options: []NamingConventionOption{{Selector: "parameter", Format: &[]string{"camelCase"}}},
		},
		// Enum members carry no accessibility modifier upstream.
		{
			Code: "enum E { myValue }",
			Options: []NamingConventionOption{
				{Selector: "memberLike", Modifiers: []string{"public"}, Format: &[]string{"PascalCase"}},
			},
		},
		// Array-pattern elements are not `destructured` upstream.
		{
			Code: "const [fooBar] = [1]; export default fooBar;",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"destructured"}, Format: &[]string{"snake_case"}},
			},
		},
		// Same-selector precedence is decided by modifier bit value, not
		// modifier count: exported (1<<10) outranks const (1<<0), so the
		// PascalCase config wins and `MyConst` passes.
		{
			Code: "export const MyConst = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"const"}, Format: &[]string{"UPPER_CASE"}},
				{Selector: "variable", Modifiers: []string{"exported"}, Format: &[]string{"PascalCase"}},
			},
		},
		// Mixed unions match no type modifier upstream (all union members
		// must match for array; typeToString equality for primitives).
		{
			Code: "declare const mixed_union: string | number;",
			Options: []NamingConventionOption{
				{Selector: "variable", Types: []string{"string"}, Format: &[]string{"UPPER_CASE"}},
				{Selector: "variable", Format: &[]string{"snake_case"}},
			},
		},
		// An unknown selector string drops the config instead of turning it
		// into a catch-all `default` selector.
		{
			Code:    "const snake_name = 1;",
			Options: []NamingConventionOption{{Selector: "notARealSelector", Format: &[]string{"PascalCase"}}},
		},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		// Quoted names that require quoting always fail the format check
		// upstream, even when the raw text would pass a checker.
		{
			Code: "type T = { '123': number };",
			Options: []NamingConventionOption{
				{Selector: "typeProperty", Format: &[]string{"snake_case"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Type Property name `123` must match one of the following formats: snake_case"},
			},
		},
		// A quoted name that is a valid identifier does not get the
		// requiresQuotes modifier upstream, so the exemption config must not
		// apply and the name falls through to the snake_case config.
		{
			Code: "export const o = { 'validName': 1 };",
			Options: []NamingConventionOption{
				{Selector: "objectLiteralProperty", Modifiers: []string{"requiresQuotes"}},
				{Selector: "objectLiteralProperty", Format: &[]string{"snake_case"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `validName` must match one of the following formats: snake_case"},
			},
		},
		// Destructured bindings receive the `unused` modifier upstream.
		// (`export default obj` keeps obj itself referenced.)
		{
			Code: "const obj = { unused_var: 1 }; export default obj; const { unused_var } = obj;",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"unused"}, Format: &[]string{"PascalCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// `export { name }` only affects module-scope bindings: the shadowed
		// local in the nested function is NOT exported, so the exported-only
		// exemption must not apply to it.
		{
			Code: "const top_level = 1; export { top_level }; export function f() { const top_level = 2; return top_level; }",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"exported"}},
				{Selector: "variable", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Variable name `top_level` must match one of the following formats: camelCase"},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, validCases, invalidCases)
}
