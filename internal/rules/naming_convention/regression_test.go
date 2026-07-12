package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Regression tests for divergences from upstream that the ported suite does
// not cover. Each case pins the upstream behavior the port was fixed to match.
//
// Each case is annotated with a judgement of whether the pinned upstream
// behavior is intentional or a bug:
//
//   - definitely intentional: upstream explicitly documents this exact behavior
//   - likely intentional: implied by upstream docs, or follows logically
//   - likely bug: unspecified behavior that would surprise users
//   - definite bug: contradicts upstream's own documentation
//
// Upstream docs: https://typescript-eslint.io/rules/naming-convention
// Upstream source (paths below are relative to packages/eslint-plugin/src/rules):
// https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/eslint-plugin/src/rules
//
// The rule is feature-frozen upstream but still accepts bug fixes, so the
// "definite bug" cases pin behavior upstream may legitimately change —
// revisit them if upstream ships a fix.
func TestNamingConventionUpstreamParity(t *testing.T) {
	t.Parallel()

	validCases := []rule_tester.ValidTestCase{
		// Names fully consumed by underscore trimming are valid (upstream
		// format checkers treat the empty string as matching).
		// Upstream: definitely intentional. The docs FAQ says "if the name
		// were to become empty via this trimming process, it is considered
		// to match all formats":
		// https://typescript-eslint.io/rules/naming-convention#how-does-the-rule-evaluate-a-names-format
		{Code: "const _ = 1;"},
		// `$` and caseless scripts satisfy camelCase/PascalCase/UPPER_CASE
		// upstream; only the first character's case-fold is checked.
		// Upstream: definitely intentional. naming-convention-utils/format.ts
		// explicitly rejects regexes in favor of case-fold comparisons so
		// that non-English (including caseless) scripts are accepted; `$`
		// passing every format follows from that implementation choice.
		{Code: "const data$ = 1;"},
		{Code: "const $data = 1;"},
		{Code: "class $Foo {}"},
		{Code: "const 名前 = 1;"},
		// Deliberate divergence from upstream: astral-plane cased letters are
		// case-checked as full runes (upstream sees a lone surrogate that
		// case-folds to itself and passes everything). Lowercase Adlam passes
		// camelCase...
		// Upstream: definite bug. Indexing UTF-16 code units makes name[0] a
		// lone surrogate, so cased astral letters bypass the case checks —
		// contradicting the documented format definitions and format.ts's own
		// stated goal of supporting non-English identifiers. Hence the port
		// diverges on purpose.
		{Code: "const \U0001E922abc = 1;"},
		{
			Code:    "const NAME$ = 1;",
			Options: []NamingConventionOption{{Selector: "variable", Format: &[]string{"UPPER_CASE"}}},
		},
		// Catch bindings are never visited upstream.
		// Upstream: likely bug. naming-convention.ts registers no CatchClause
		// listener and no doc mentions catch parameters, but the docs claim
		// the `default` selector "matches everything" and ESLint core's
		// camelcase rule does flag catch bindings. Feature-frozen, so a fix
		// upstream is unlikely.
		{Code: "try { void 0; } catch (Some_Error) { throw Some_Error; }"},
		// Parameters in type positions are never visited upstream.
		// Upstream: likely bug. The parameter handler only visits
		// Function*/TSDeclareFunction/TSEmptyBodyFunctionExpression nodes, so
		// bodiless `declare function` and overload-signature parameters ARE
		// linted while TSFunctionType/TSMethodSignature parameters are not —
		// an inconsistency the docs ("matches any function parameter") don't
		// sanction.
		{
			Code:    "type T = (foo_bar: number) => void;",
			Options: []NamingConventionOption{{Selector: "parameter", Format: &[]string{"camelCase"}}},
		},
		{
			Code:    "interface I { m(foo_bar: number): void }",
			Options: []NamingConventionOption{{Selector: "parameter", Format: &[]string{"camelCase"}}},
		},
		// Enum members carry no accessibility modifier upstream.
		// Upstream: definite bug (code contradicts docs). The docs FAQ
		// promises "members that cannot specify an accessibility will always
		// have the `public` modifier. This means that the following config
		// will always match any `enumMember`: {selector: 'memberLike',
		// modifiers: ['public']}" — but the TSEnumMember handler only ever
		// adds requiresQuotes, while object-literal/type properties and type
		// methods do get public.
		// https://typescript-eslint.io/rules/naming-convention#what-happens-if-i-provide-a-modifiers-to-a-group-selector
		{
			Code: "enum E { myValue }",
			Options: []NamingConventionOption{
				{Selector: "memberLike", Modifiers: []string{"public"}, Format: &[]string{"PascalCase"}},
			},
		},
		// Array-pattern elements are not `destructured` upstream.
		// Upstream: definitely intentional. The docs define the modifier as
		// matching "a variable declared via an object destructuring pattern",
		// and isDestructured() in naming-convention.ts checks only shorthand
		// object-Property parents.
		{
			Code: "const [fooBar] = [1]; export default fooBar;",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"destructured"}, Format: &[]string{"snake_case"}},
			},
		},
		// Same-selector precedence is decided by modifier bit value, not
		// modifier count: exported (1<<10) outranks const (1<<0), so the
		// PascalCase config wins and `MyConst` passes.
		// Upstream: likely intentional. The tier ordering (filter > types >
		// modifiers > bare) is documented; within the modifier tier the
		// bit-value tie-break is an undocumented implementation artifact
		// (parse-options.ts ORs modifier bits into modifierWeight and
		// validator.ts sorts it descending), but some deterministic
		// tie-break has to exist.
		{
			Code: "export const MyConst = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Modifiers: []string{"const"}, Format: &[]string{"UPPER_CASE"}},
				{Selector: "variable", Modifiers: []string{"exported"}, Format: &[]string{"PascalCase"}},
			},
		},
		// Mixed unions match no type modifier upstream (all union members
		// must match for array; typeToString equality for primitives).
		// Upstream: likely intentional. The docs define types: ['string'] as
		// "any type assignable to `string | null | undefined`", which
		// excludes `string | number`; isCorrectType in
		// naming-convention-utils/validator.ts implements this via
		// typeToString equality for primitives.
		{
			Code: "declare const mixed_union: string | number;",
			Options: []NamingConventionOption{
				{Selector: "variable", Types: []string{"string"}, Format: &[]string{"UPPER_CASE"}},
				{Selector: "variable", Format: &[]string{"snake_case"}},
			},
		},
		// An unknown selector string drops the config instead of turning it
		// into a catch-all `default` selector.
		// Upstream: definitely intentional that this is never a catch-all —
		// upstream's JSON schema (naming-convention-utils/schema.ts) rejects
		// the whole rule config loudly. The port diverges in failure mode by
		// silently dropping the config, matching upstream's linting outcome
		// but not its loud config error.
		{
			Code:    "const snake_name = 1;",
			Options: []NamingConventionOption{{Selector: "notARealSelector", Format: &[]string{"PascalCase"}}},
		},
		// A non-string selector value is likewise dropped, not treated as
		// `default`.
		// Upstream: definitely intentional (schema rejection, as above).
		{
			Code:    "const snake_name = 1;",
			Options: []NamingConventionOption{{Selector: 42, Format: &[]string{"PascalCase"}}},
		},
		// JS-only regex constructs (lookahead) are valid upstream and must not
		// crash the run; the negative lookahead excludes this name, so the
		// UPPER_CASE config never applies.
		// Upstream: definitely intentional. Filters are `new RegExp(...)`, so
		// JS regex features are inherent; the port uses regexp2 to match.
		{
			Code: "const mapStateToProps = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Filter: MatchRegex{Match: true, Regex: "^(?!mapStateToProps$).*"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// A #private member never carries the public modifier upstream, so a
		// public-requiring config must not match it.
		// Upstream: definitely intentional. getMemberModifiers in
		// naming-convention.ts has an explicit #private / accessibility /
		// public else-chain, and TypeScript itself forbids accessibility
		// modifiers on private identifiers.
		{
			Code: "class MyClass { #foo = 1; }",
			Options: []NamingConventionOption{
				{Selector: "classProperty", Modifiers: []string{"public"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// Enums never get the const modifier upstream, so a const-requiring
		// config never matches — even a `const enum`.
		// Upstream: definitely intentional. The docs/schema allow only
		// [exported, unused] modifiers on the enum selector. Note upstream's
		// schema would reject this config outright; the port instead keeps it
		// as a config that never matches.
		{
			Code: "const enum fooEnum {}",
			Options: []NamingConventionOption{
				{Selector: "enum", Modifiers: []string{"const"}, Format: &[]string{"UPPER_CASE"}},
			},
		},
		// A numeric key that only matches the requiresQuotes exemption config
		// is skipped by it (no format), mirroring upstream's exemption idiom.
		// Upstream: likely intentional. The runtime property name of
		// {123: 'a'} is "123", which is not a valid identifier (obj.123 is
		// invalid), so it gets requiresQuotes and the documented format:null
		// exemption idiom applies:
		// https://typescript-eslint.io/rules/naming-convention#ignore-properties-that-require-quotes
		{
			Code: "const x = { 123: 'a' };",
			Options: []NamingConventionOption{
				{Selector: "objectLiteralProperty", Modifiers: []string{"requiresQuotes"}},
				{Selector: "objectLiteralProperty", Format: &[]string{"snake_case"}},
			},
		},
		// Interface and type-literal accessors are TSMethodSignature nodes
		// upstream, so classicAccessor configs never visit them.
		// Upstream: definite bug (code contradicts docs). The docs say
		// classicAccessor "matches any accessor" and typeMethod "does not
		// match accessors", but the TSMethodSignature listener in
		// naming-convention.ts has no kind filter, so interface/type-literal
		// get/set route to typeMethod. The rule predates interface accessors
		// and is feature-frozen.
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
		// Upstream: likely intentional. validatePredefinedFormat in
		// naming-convention-utils/validator.ts consults the format checkers
		// only when the requiresQuotes modifier is absent, and the docs'
		// exemption guidance (requiresQuotes + format: null) implies such
		// names always fail any real format.
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
		// Upstream: definitely intentional. The docs say verbatim: "there is
		// no way to ignore any name that is quoted - only names that are
		// required to be quoted. This is intentional - adding quotes around a
		// name is not an escape hatch for proper naming."
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
		// Upstream: definitely intentional. `unused` is documented as
		// matching "anything that is not used", and the VariableDeclarator
		// handler applies isUnused to every binding it extracts.
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
		// Upstream: definitely intentional. `exported` is documented as
		// matching "anything that is exported from the module"; the nested
		// binding is a distinct variable that is not exported.
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
		// Upstream: definitely intentional (JS regex semantics; see the valid
		// lookahead case above).
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
		// Upstream: definite bug (code contradicts docs and schema). The docs
		// list allowed types for autoAccessor and schema.ts passes
		// allowType=true for it, but SelectorsAllowedToHaveTypes in
		// naming-convention-utils/validator.ts was not updated when
		// autoAccessor was added (typescript-eslint#8084), so the constraint
		// is accepted and then silently ignored.
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
		// Upstream: likely bug. The docs promise selectors are sorted
		// "most-specific to least specific" and accessor is a strict subset
		// of memberLike, yet the grouped-vs-grouped fallback in
		// naming-convention-utils/validator.ts is raw bit-value descending,
		// with only method/property special-cased "for backward
		// compatibility".
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
		// Upstream: definite bug. The ImportSpecifier guard in
		// naming-convention.ts is commented "Handle `import { default as
		// Foo }`" but only skips Identifier imports, so string-literal
		// imported names — a named-import form — fall through and are tagged
		// as default imports, contradicting the docs ("does not match named
		// imports").
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
		// Upstream: likely intentional. The runtime key of {123: 'a'} really
		// is "123" and is not a valid identifier, so requiresQuotes applies
		// and (per the requires-quotes rule above) every real format fails;
		// the documented requiresQuotes exemption is the escape hatch.
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
		// Upstream: likely intentional. `${node.value}` yields the actual
		// runtime property name — {0x10: 'a'} creates the key "16" — even
		// though reporting `16` for source text `0x10` may read oddly.
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
		// Upstream: likely intentional (same reasoning as the numeric
		// object-literal keys above).
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
		// Upstream: definite bug — same contradiction with the documented
		// classicAccessor/typeMethod split as the classicAccessor valid cases
		// above.
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
		// Upstream: definitely intentional — explicit MethodDefinition and
		// Property [kind=get/set] listeners route to the classicAccessor
		// validator.
		{
			Code: "const o = { get Foo() { return 1; } };",
			Options: []NamingConventionOption{
				{Selector: "classicAccessor", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Classic Accessor name `Foo` must match one of the following formats: camelCase"},
			},
		},
		// ...and capital Adlam fails it (upstream would pass both; see the
		// format checkers' divergence note).
		// Upstream: definite bug (lone-surrogate case-folding; see the
		// astral-plane divergence note in the valid cases).
		{
			Code: "const \U0001E900abc = 1;",
			Options: []NamingConventionOption{
				{Selector: "variable", Format: &[]string{"camelCase"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Message: "Variable name `\U0001E900abc` must match one of the following formats: camelCase"},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, validCases, invalidCases)
}
