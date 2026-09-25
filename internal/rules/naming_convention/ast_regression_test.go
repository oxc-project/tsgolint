package naming_convention

import (
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"testing"
)

// These additional native-AST regressions were compared with typescript-eslint.
// They are separate from the frozen, generated upstream corpus.
func TestNamingConventionASTRegressions(t *testing.T) {
	t.Run("parenthesized async methods", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "class Example { Bad_Name = (async () => {}); } const obj = { Bad_Name: (async function () {}) };", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"default\",\"format\":null},{\"selector\":[\"classMethod\",\"objectLiteralMethod\"],\"modifiers\":[\"async\"],\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class Method name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 25}, {MessageId: "doesNotMatchFormat", Message: "Object Literal Method name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 62, EndLine: 1, EndColumn: 70}}}})
	})
	t.Run("parenthesized function type", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "interface Example { Bad_Name: (() => void); }", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"typeMethod\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Type Method name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 21, EndLine: 1, EndColumn: 29}}}})
	})
	t.Run("parenthesized async variable", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "const Bad_Name = (async () => {});", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"variable\",\"modifiers\":[\"async\"],\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Variable name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 15}}}})
	})
	t.Run("rest parameter annotation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "function example(...Bad_Name: number[]) {}", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"parameter\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Parameter name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 21, EndLine: 1, EndColumn: 29}}}})
	})
	t.Run("optional parameter annotation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "function example(Bad_Name?: number) {}", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"parameter\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Parameter name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 18, EndLine: 1, EndColumn: 35}}}})
	})
	t.Run("parameter property annotation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "class Example { constructor(public Bad_Name: number) {} }", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"parameterProperty\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Parameter Property name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 36, EndLine: 1, EndColumn: 52}}}})
	})
	t.Run("ignored binding contexts", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "let Bad_Name; ({Bad_Name} = {Good_Name:1}); try {} catch(Bad_Error) {} type Example<T> = T extends infer U ? { [Bad_Key in keyof U]: U[Bad_Key] } : never;", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"objectLiteralProperty\",\"format\":[\"camelCase\"]},{\"selector\":\"parameter\",\"format\":[\"camelCase\"]},{\"selector\":\"typeParameter\",\"format\":[\"PascalCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Object Literal Property name `Good_Name` must match one of the following formats: camelCase", Line: 1, Column: 30, EndLine: 1, EndColumn: 39}}}})
	})
	t.Run("computed names", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{{Code: "class Example { [\"Bad_Name\"] = 1; [\"Bad_Method\"]() {} } const obj = {[\"Bad_Name\"]:1};", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"memberLike\",\"format\":[\"camelCase\"]}]")}}, nil)
	})
	t.Run("numeric and private member names", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "class Example { #Bad_Name = 1; \"#Bad_Name\" = 1; 0x10 = 1; }", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"classProperty\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Class Property name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 26}, {MessageId: "doesNotMatchFormat", Message: "Class Property name `#Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 32, EndLine: 1, EndColumn: 43}, {MessageId: "doesNotMatchFormat", Message: "Class Property name `16` must match one of the following formats: camelCase", Line: 1, Column: 49, EndLine: 1, EndColumn: 53}}}})
	})
	t.Run("type signature parameters ignored", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{{Code: "interface Example { method(Bad_Name: number): void; } type Callable = (Bad_Name: number) => void;", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"parameter\",\"format\":[\"camelCase\"]}]")}}, nil)
	})
	t.Run("method parameters checked", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "class Example { method(Bad_Name: number) {} set value(Bad_Name: number) {} }", Options: rule_tester.OptionsFromJSON[any]("[{\"selector\":\"parameter\",\"format\":[\"camelCase\"]}]"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Parameter name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 24, EndLine: 1, EndColumn: 40}, {MessageId: "doesNotMatchFormat", Message: "Parameter name `Bad_Name` must match one of the following formats: camelCase", Line: 1, Column: 55, EndLine: 1, EndColumn: 71}}}})
	})
}

func TestNamingConventionForBindingPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			{Code: `let Bad_Name; for ({Bad_Name} of items) {}`, Options: rule_tester.OptionsFromJSON[any](`[{"selector":"objectLiteralProperty","format":["camelCase"]}]`)},
			{Code: `let Bad_Name; for ({Bad_Name} in items) {}`, Options: rule_tester.OptionsFromJSON[any](`[{"selector":"objectLiteralProperty","format":["camelCase"]}]`)},
		}, nil)
}

func TestNamingConventionComputedAutoAccessors(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, nil, []rule_tester.InvalidTestCase{{Code: "class Example { accessor [\"Bad_Name\"] = 1; accessor [Bad_Name] = 1; accessor [\"foo\" + \"bar\"] = 1; accessor [true] = 1; accessor [`foo`] = 1; }", Options: rule_tester.OptionsFromJSON[any](`[{"selector":"autoAccessor","format":["PascalCase"]}]`), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `Bad_Name` must match one of the following formats: PascalCase", Line: 1, Column: 27, EndLine: 1, EndColumn: 37}, {MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `Bad_Name` must match one of the following formats: PascalCase", Line: 1, Column: 54, EndLine: 1, EndColumn: 62}, {MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `undefined` must match one of the following formats: PascalCase", Line: 1, Column: 79, EndLine: 1, EndColumn: 92}, {MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `true` must match one of the following formats: PascalCase", Line: 1, Column: 109, EndLine: 1, EndColumn: 113}, {MessageId: "doesNotMatchFormat", Message: "Auto Accessor name `undefined` must match one of the following formats: PascalCase", Line: 1, Column: 130, EndLine: 1, EndColumn: 135}}}})
}
