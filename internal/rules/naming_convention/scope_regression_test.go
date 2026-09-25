package naming_convention

import (
	"strings"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func modifierTestOptions(selector, modifier string) NamingConventionOptions {
	return NamingConventionOptions{
		{Selector: NamingSelector{selector}, Format: []string{"snake_case"}},
		{Selector: NamingSelector{selector}, Modifiers: []string{modifier}, Format: []string{"PascalCase"}},
	}
}

func modifierFailure(code string) rule_tester.InvalidTestCaseError {
	const name = "BadName"
	start := strings.Index(code, name)
	if start < 0 {
		panic("test name not found in code")
	}
	line := 1 + strings.Count(code[:start], "\n")
	lineStart := strings.LastIndex(code[:start], "\n") + 1
	column := start - lineStart + 1
	endColumn := column + len(name)
	if strings.HasPrefix(code[start+len(name):], ": unknown") {
		endColumn = column + len(name+": unknown")
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: "doesNotMatchFormat",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func TestNamingConventionScopeModifiers(t *testing.T) {
	unused := modifierTestOptions("default", "unused")
	exported := modifierTestOptions("variable", "exported")
	globalVariable := modifierTestOptions("variable", "global")
	globalFunction := modifierTestOptions("function", "global")

	valid := []rule_tester.ValidTestCase{
		{Code: `function BadName() { BadName(); }`, Options: modifierTestOptions("function", "unused")},
		{Code: `class BadName { self?: BadName }`, Options: unused},
		{Code: `interface BadName { self?: BadName }`, Options: unused},
		{Code: `type BadName = BadName;`, Options: unused},
		{Code: `const BadName = 1; type UsedOnlyForTypeof = typeof BadName;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `let BadName = 0; BadName++;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `let BadName = 0; ++BadName;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `let BadName = 0; BadName = 1;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `let BadName = 0; BadName = BadName + 1;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `const BadName = (((() => BadName())));`, Options: modifierTestOptions("variable", "unused")},
		{Code: `function f(this: object) {}`, Options: modifierTestOptions("parameter", "unused")},
		{Code: `declare namespace Ambient { function BadName(): void; }`, Options: modifierTestOptions("function", "unused")},
		{Code: `const BadName = 1; export { BadName as PublicName };`, Options: exported},
		{Code: `const BadName = 1; const obj = {}; obj.BadName;`, Options: modifierTestOptions("variable", "unused")},
		{Code: `var BadName = 1;`, Options: globalVariable},
		{Code: `for (var BadName = 0; false;) {}`, Options: globalVariable},
		{Code: `function BadName() {}`, Options: globalFunction},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: `const BadName = BadName;`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`const BadName = BadName;`)}},
		{Code: `const BadName = 1; const obj = { BadName }; void obj;`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`const BadName = 1; const obj = { BadName }; void obj;`)}},
		{Code: `if (true) { var BadName = 1; }`, Options: globalVariable, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`if (true) { var BadName = 1; }`)}},
		{Code: `for (let BadName = 0; false;) {}`, Options: globalVariable, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`for (let BadName = 0; false;) {}`)}},
		{Code: `const f = function BadName() {};`, Options: globalFunction, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`const f = function BadName() {};`)}},
		{Code: `let BadName = 0; !BadName;`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName = 0; !BadName;`)}},
		{Code: `let BadName = 0; +BadName;`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName = 0; +BadName;`)}},
		{Code: `let BadName = 0; -BadName;`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName = 0; -BadName;`)}},
		{Code: `let BadName = 0; BadName = BadName + 1; console.log(BadName);`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName = 0; BadName = BadName + 1; console.log(BadName);`)}},
		{Code: `const BadName = (((() => BadName()))); console.log(BadName);`, Options: modifierTestOptions("variable", "unused"), Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`const BadName = (((() => BadName()))); console.log(BadName);`)}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}

func TestNamingConventionUnusedClosures(t *testing.T) {
	unused := modifierTestOptions("variable", "unused")
	valid := []rule_tester.ValidTestCase{
		{Code: `let BadName: unknown = 0; BadName = () => BadName;`, Options: unused},
		{Code: `let BadName: unknown = 0; BadName = (() => BadName)();`, Options: unused},
		{Code: `let BadName: unknown = 0; BadName = { callback: () => BadName };`, Options: unused},
		{Code: `let BadName: unknown = 0; BadName = (0, () => BadName);`, Options: unused},
		{Code: `let BadName: unknown = 0; BadName = (() => BadName, 1);`, Options: unused},
		{Code: `let BadName: unknown; BadName = () => { BadName = BadName; };`, Options: unused},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: `let BadName: unknown = 0; BadName = consume(() => BadName);`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown = 0; BadName = consume(() => BadName);`)}},
		{Code: `let BadName: unknown = 0; BadName = (other = () => BadName);`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown = 0; BadName = (other = () => BadName);`)}},
		{Code: `let BadName: unknown; function f() { BadName = BadName + 1; }`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; function f() { BadName = BadName + 1; }`)}},
		{Code: `let BadName: unknown = 0; for (;;) { BadName = () => BadName; break; }`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown = 0; for (;;) { BadName = () => BadName; break; }`)}},
		{Code: `let BadName: unknown; BadName = () => { function inner() { return BadName; } };`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; BadName = () => { function inner() { return BadName; } };`)}},
		{Code: `let BadName: unknown; BadName = class { static { BadName = () => BadName; } };`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; BadName = class { static { BadName = () => BadName; } };`)}},
		{Code: `let BadName: unknown; namespace N { BadName = () => BadName; }`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; namespace N { BadName = () => BadName; }`)}},
		{Code: `let BadName: unknown; class C { value = (BadName = () => BadName, 0); }`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; class C { value = (BadName = () => BadName, 0); }`)}},
		{Code: `let BadName: unknown; BadName = consume(() => { BadName = BadName; });`, Options: unused, Errors: []rule_tester.InvalidTestCaseError{modifierFailure(`let BadName: unknown; BadName = consume(() => { BadName = BadName; });`)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}
