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
		// A non-string selector value is likewise dropped, not treated as
		// `default`.
		{
			Code:    "const snake_name = 1;",
			Options: []NamingConventionOption{{Selector: 42, Format: &[]string{"PascalCase"}}},
		},
		// JS-only regex constructs (lookahead) are valid upstream and must not
		// crash the run; the negative lookahead excludes this name, so the
		// UPPER_CASE config never applies.
		{
			Code: "const mapStateToProps = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Filter: MatchRegex{Match: true, Regex: "^(?!mapStateToProps$).*"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// A #private member never carries the public modifier upstream, so a
		// public-requiring config must not match it.
		{
			Code: "class MyClass { #foo = 1; }",
			Options: []NamingConventionOption{
				{Selector: "classProperty", Modifiers: []string{"public"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// Enums never get the const modifier upstream, so a const-requiring
		// config never matches — even a `const enum`.
		{
			Code: "const enum fooEnum {}",
			Options: []NamingConventionOption{
				{Selector: "enum", Modifiers: []string{"const"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// A numeric key that only matches the requiresQuotes exemption config
		// is skipped by it (no format), mirroring upstream's exemption idiom.
		{
			Code: "const x = { 123: 'a' };",
			Options: []NamingConventionOption{
				{Selector: "objectLiteralProperty", Modifiers: []string{"requiresQuotes"}},
				{Selector: "objectLiteralProperty", Format: &[]string{"snake_case"}},
			},
		},
		// Interface and type-literal accessors are TSMethodSignature nodes
		// upstream, so classicAccessor configs never visit them.
		{
			Code: "interface I { get Foo(): string }",
			Options: []NamingConventionOption{
				{Selector: "classicAccessor", Format: &[]string{"camelCase"}},
			},
		},
		{
			Code: "type T = { set Foo(v: string) };",
			Options: []NamingConventionOption{
				{Selector: "classicAccessor", Format: &[]string{"camelCase"}},
			},
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
		// A lookahead filter that does apply still enforces the format
		// (exercises regexp2 matching, not just compilation).
		{
			Code: "const foo_bar = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Filter: MatchRegex{Match: true, Regex: "^(?!mapStateToProps$).*"}, Format: &[]string{"PascalCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Variable name `foo_bar` must match one of the following formats: PascalCase"},
			},
		},
		// Upstream never honors `types` on autoAccessor (it is not in
		// SelectorsAllowedToHaveTypes), so the config applies to a number-typed
		// accessor unconditionally.
		{
			Code: "class MyClass { accessor foo = 5; }",
			Options: []NamingConventionOption{
				{Selector: "autoAccessor", Types: []string{"string"}, Format: &[]string{"PascalCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `foo` must match one of the following formats: PascalCase"},
			},
		},
		// Between different meta selectors upstream orders by raw selector
		// value descending (accessor is NOT in the method/property tier), so
		// memberLike outranks accessor even when accessor is listed first.
		{
			Code: "class MyClass { get Foo() { return 1; } }",
			Options: []NamingConventionOption{
				{Selector: "accessor", Format: &[]string{"PascalCase"}},
				{Selector: "memberLike", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Classic Accessor name `Foo` must match one of the following formats: camelCase"},
			},
		},
		// A string-literal imported name gets the default modifier upstream
		// (only Identifier names other than `default` are skipped).
		{
			Code: "import { \"foo-bar\" as Bar } from 'foo_bar';",
			Options: []NamingConventionOption{
				{Selector: "import", Modifiers: []string{"default"}, Format: &[]string{"UPPER_CASE"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Import name `Bar` must match one of the following formats: UPPER_CASE"},
			},
		},
		// Numeric keys are validated upstream: the stringified value always
		// requires quotes, so a format config always reports it.
		{
			Code: "const x = { 123: 'a' };",
			Options: []NamingConventionOption{
				{Selector: "objectLiteralProperty", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `123` must match one of the following formats: camelCase"},
			},
		},
		// The numeric name is the JS value string (`${node.value}`), not the
		// source text: 0x10 is validated as `16`.
		{
			Code: "const x = { 0x10: 'a' };",
			Options: []NamingConventionOption{
				{Selector: "objectLiteralProperty", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `16` must match one of the following formats: camelCase"},
			},
		},
		// Numeric class member names get the same treatment via handleMember.
		{
			Code: "class MyClass { 123 = 'a'; }",
			Options: []NamingConventionOption{
				{Selector: "classProperty", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Class Property name `123` must match one of the following formats: camelCase"},
			},
		},
		// Interface and type-literal accessors are typeMethod upstream
		// (TSMethodSignature has no kind filter).
		{
			Code: "interface I { get Foo(): string }",
			Options: []NamingConventionOption{
				{Selector: "typeMethod", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Type Method name `Foo` must match one of the following formats: camelCase"},
			},
		},
		{
			Code: "type T = { set Foo(v: string) };",
			Options: []NamingConventionOption{
				{Selector: "typeMethod", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Type Method name `Foo` must match one of the following formats: camelCase"},
			},
		},
		// Class and object-literal accessors remain classicAccessor.
		{
			Code: "const o = { get Foo() { return 1; } };",
			Options: []NamingConventionOption{
				{Selector: "classicAccessor", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Classic Accessor name `Foo` must match one of the following formats: camelCase"},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, validCases, invalidCases)
}
